// Package allegrex decodes and executes code for the Sony PSP's "Allegrex" CPU: a
// 32-bit little-endian MIPS32R2 core with a single-precision COP1 FPU and a 128-bit
// COP2 vector unit (the VFPU). It shares the R3000 skeleton of tools/cpu/mips — the
// same decode/exec split, the same branch and load delay-slot machinery, the same
// Bus interface — and diverges where the PSP does:
//
//   - MIPS32R2 integer additions absent from MIPS I: movz/movn/rotr in SPECIAL, the
//     SPECIAL2 group (op 0x1C: mul/madd/clz/clo) and the SPECIAL3 group (op 0x1F:
//     ext/ins/seb/seh/wsbh/bitrev).
//   - COP1 (op 0x11): a single-precision-only FPU (fpu.go). There is no double
//     precision. lwc1/swc1 move FPU registers to and from memory.
//   - COP2 (op 0x12) is the VFPU, not the PSX GTE (vfpu.go), together with the VFPU
//     load/store opcodes (lv/sv). The VFPU is decoded broadly for legibility; a
//     common subset (loads/stores, vector/matrix arithmetic, prefixes) executes and
//     the long tail Halts with its word, so gaps are explicit rather than wrong.
//
// MIPS32R2 removes the load delay slot (loads are interlocked), so — unlike the
// R3000 mips core — a load is visible to the very next instruction; the interpreter
// retires loads immediately. Branch delay slots remain. Unimplemented encodings call
// Halt with the offending word, exactly as the mips core does.
package allegrex

import "fmt"

// Flow classifies how control leaves an instruction. It mirrors the enum used by
// tools/arm and tools/sm83 so the shared codetrace/dis command skeletons apply.
//
// MIPS mapping:
//
//	beq/bne/blez/bgtz/bltz/bgez/...  FlowBranch  (conditional, PC-relative)
//	j / b                            FlowJump    (unconditional, region target)
//	jal                              FlowCall    (call, region target, returns)
//	jalr                             FlowIndCall (call through a register)
//	jr $ra                           FlowReturn
//	jr <other>                       FlowIndJump (jump through a register/table)
//	syscall / break / illegal        FlowStop
//
// Every branch and jump additionally executes its delay slot; the tracer and the
// interpreter account for that separately from the Flow classification.
type Flow int

const (
	FlowSeq     Flow = iota // falls through to the next instruction
	FlowBranch              // conditional branch: continues AND may take Target
	FlowJump                // unconditional jump to Target, no fall-through
	FlowCall                // jal: calls Target, normally returns after it
	FlowReturn              // jr $ra: path ends
	FlowIndJump             // jr <reg>: target not statically known
	FlowIndCall             // jalr: target unknown but returns, so continue
	FlowStop                // syscall/break/illegal/truncated: treat as a stop
)

func (f Flow) String() string {
	switch f {
	case FlowSeq:
		return "seq"
	case FlowBranch:
		return "branch"
	case FlowJump:
		return "jump"
	case FlowCall:
		return "call"
	case FlowReturn:
		return "return"
	case FlowIndJump:
		return "indjump"
	case FlowIndCall:
		return "indcall"
	case FlowStop:
		return "stop"
	}
	return "?"
}

// Inst is one decoded MIPS instruction. Len is always 4 for a real instruction;
// a decode of a short/empty slice yields Len 0 and FlowStop.
type Inst struct {
	Addr      uint32 // address of this instruction
	Len       int    // 4 for a decoded instruction, 0 when out of range
	Mnem      string // bare mnemonic, e.g. "addu", "lw", "beq"
	Text      string // formatted "mnem operands"
	Flow      Flow
	Target    uint32 // branch/jump/call destination, valid when HasTarget
	HasTarget bool
	HasDelay  bool // this instruction is followed by a delay slot (branch/jump)
}

func (in Inst) String() string {
	return fmt.Sprintf("$%08X: %s", in.Addr, in.Text)
}
