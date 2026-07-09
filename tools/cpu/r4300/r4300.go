// Package r4300 decodes and executes NEC VR4300 code — the CPU at the heart of
// the Nintendo 64. It is a 64-bit big-endian RISC implementing the MIPS III ISA,
// with a fixed 4-byte instruction, 32 general registers (R0 hardwired to zero),
// the HI/LO multiply-divide pair, a floating-point unit (COP1), and a system
// control coprocessor with a 32-entry TLB (COP0).
//
// It is a superset of the R3000A in tools/cpu/mips, but differs in four ways
// that pervade the decoder and the interpreter, so the two are separate packages
// exactly as tools/cpu/arm and tools/cpu/arm60 are:
//
//   - Endianness. Instructions and data are big-endian; the R3000A is little.
//   - Register width. The GPRs, HI/LO and the program counter are 64-bit. A
//     32-bit operation writes its result sign-extended into the full register,
//     so `addu` and `daddu` differ even when the operands fit in 32 bits.
//   - Coprocessors. There is an FPU (see fpu.go) and a TLB (see cop0.go); the
//     R3000A has neither, and its COP2 is the PlayStation's GTE.
//   - No load delay slot. The VR4300 interlocks loads, so the value is visible
//     to the next instruction. The R3000A exposes the hazard and tools/cpu/mips
//     models it with a two-register-file scheme; here a single file suffices.
//
// The branch delay slot remains, and is joined by the *branch-likely* family
// (beql, bnel, and the REGIMM variants), which annuls — skips — its delay slot
// when the branch is not taken. Both the recursive-descent tracer and the
// interpreter must account for it: see Inst.Annul.
//
// The package follows the shape of the other CPU packages: a table-driven Decode
// producing an Inst with a Flow classification for the disassembler and tracer
// commands, and a separate CPU with a small Bus interface for the machine model
// (see tools/platform/n64). Unimplemented encodings call Halt with the offending
// word, so gaps are explicit rather than silently wrong.
package r4300

import "fmt"

// Flow classifies how control leaves an instruction. It mirrors the enum used by
// tools/cpu/mips and tools/cpu/arm so the shared codetrace/dis command skeletons
// apply.
//
// MIPS mapping:
//
//	beq/bne/blez/bgtz/bltz/bgez/...  FlowBranch  (conditional, PC-relative)
//	beql/bnel/blezl/bgtzl/...        FlowBranch  (as above, but Annul is set)
//	j / b                            FlowJump    (unconditional, region target)
//	jal                              FlowCall    (call, region target, returns)
//	jalr                             FlowIndCall (call through a register)
//	jr $ra                           FlowReturn
//	jr <other>                       FlowIndJump (jump through a register/table)
//	syscall / break / eret / illegal FlowStop
type Flow int

const (
	FlowSeq     Flow = iota // falls through to the next instruction
	FlowBranch              // conditional branch: continues AND may take Target
	FlowJump                // unconditional jump to Target, no fall-through
	FlowCall                // jal: calls Target, normally returns after it
	FlowReturn              // jr $ra: path ends
	FlowIndJump             // jr <reg>: target not statically known
	FlowIndCall             // jalr: target unknown but returns, so continue
	FlowStop                // syscall/break/eret/illegal/truncated: treat as a stop
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

// Inst is one decoded VR4300 instruction. Len is always 4 for a real
// instruction; a decode of a short slice yields Len 0 and FlowStop.
type Inst struct {
	Addr      uint32 // address of this instruction
	Len       int    // 4 for a decoded instruction, 0 when out of range
	Mnem      string // bare mnemonic, e.g. "daddu", "ld", "beql"
	Text      string // formatted "mnem operands"
	Flow      Flow
	Target    uint32 // branch/jump/call destination, valid when HasTarget
	HasTarget bool
	HasDelay  bool // this instruction is followed by a delay slot (branch/jump)

	// Annul marks a branch-likely: the delay slot executes only when the branch
	// is taken. A tracer still covers the slot as code (it runs on the taken
	// path); the interpreter must skip it when the branch falls through.
	Annul bool
}

func (in Inst) String() string {
	return fmt.Sprintf("$%08X: %s", in.Addr, in.Text)
}
