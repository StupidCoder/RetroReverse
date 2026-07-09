// Package rsp decodes and executes Reality Signal Processor code — the vector
// coprocessor inside the Nintendo 64's RCP.
//
// The RSP is a CPU, and it belongs here rather than in tools/platform/n64
// because the code it runs is on the cartridge. A game's display lists are not
// drawn by the RDP directly: they are interpreted by *microcode*, a program the
// game ships and DMAs into the RSP's instruction memory. Reversing the render
// pipeline therefore means disassembling that program (see tools/cmd/disrsp),
// which is why the microcode is never high-level-emulated here.
//
// The scalar half is a cut-down MIPS I: 32 general registers, the familiar
// arithmetic, logic, shifts, loads, stores, branches and their delay slots. What
// it does *not* have shapes the core as much as what it does:
//
//   - No multiply or divide, and so no HI/LO pair. Multiplication is the vector
//     unit's job.
//   - No 64-bit anything, no unaligned load/store family, no branch-likely.
//   - No TLB, no exceptions, no interrupts. BREAK stops the processor and tells
//     the CPU about it; nothing else diverts control.
//   - A 12-bit program counter. It wraps inside the 4 KiB of instruction memory,
//     so a jump past the end lands back at the start.
//
// Its coprocessors are unusual. COP0 is not a system-control unit at all: it is
// the RSP's window onto its own memory-mapped registers — the DMA engine, the
// status register, and the RDP's command queue. COP2 is the vector unit (vu.go):
// 32 registers of eight 16-bit lanes, a 48-bit-per-lane accumulator, and the
// element-addressed load/store family that feeds them.
package rsp

import "fmt"

// Flow classifies how control leaves an instruction, mirroring tools/cpu/r4300
// and tools/cpu/mips so the shared dis/codetrace command skeletons apply.
type Flow int

const (
	FlowSeq     Flow = iota // falls through to the next instruction
	FlowBranch              // conditional branch: continues AND may take Target
	FlowJump                // unconditional jump to Target, no fall-through
	FlowCall                // jal
	FlowReturn              // jr $ra
	FlowIndJump             // jr <reg>
	FlowIndCall             // jalr
	FlowStop                // break, or an unmodelled encoding
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

// Inst is one decoded RSP instruction. Addresses are IMEM offsets, 12 bits wide.
type Inst struct {
	Addr      uint32
	Len       int // 4 for a decoded instruction, 0 when out of range
	Mnem      string
	Text      string
	Flow      Flow
	Target    uint32
	HasTarget bool
	HasDelay  bool
}

func (in Inst) String() string { return fmt.Sprintf("$%03X: %s", in.Addr, in.Text) }

// IMEMSize and DMEMSize are the RSP's two 4 KiB memories.
const (
	IMEMSize = 0x1000
	DMEMSize = 0x1000
)
