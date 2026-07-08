// Package mips decodes and executes MIPS R3000A code — the CPU at the heart of
// the Sony PlayStation (PSX). It is a 32-bit little-endian RISC (MIPS I ISA)
// with a fixed 4-byte instruction, 32 general registers (R0 hardwired to zero),
// the HI/LO multiply-divide pair, and coprocessors COP0 (system control /
// exceptions) and COP2 (the GTE geometry engine — see gte.go). There is no FPU.
//
// Two features distinguish MIPS from every other core in this repo and drive the
// design of both the decoder and the interpreter:
//
//   - Branch/jump delay slots. The instruction *after* a branch or jump always
//     executes before control transfers. A recursive-descent tracer must decode
//     and cover the delay-slot instruction before honouring the branch, and the
//     interpreter runs it before committing the new PC.
//   - Load delay slots. On real hardware the target of a load is not visible to
//     the immediately following instruction. Well-behaved compiler output does
//     not depend on this, and (like most PSX emulators) the interpreter here
//     retires loads at the end of the instruction, which the SingleStepTests
//     vectors validate.
//
// The package follows the shape of the other CPU packages (arm, sm83, x86):
// a table-driven Decode producing an Inst with a Flow classification for the
// disassembler/tracer commands, and a separate CPU with a small Bus interface
// for the machine models (see tools/psx). COP0/COP2 are reached through optional
// caller hooks, exactly as tools/arm routes CP15 through CPU.Coproc.
package mips

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
