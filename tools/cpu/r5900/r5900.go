// Package r5900 decodes and executes Emotion Engine code — the CPU at the heart
// of the PlayStation 2. It is a 64-bit little-endian RISC built on the MIPS
// III/IV ISA, with 128-bit general registers, a SIMD integer unit (the
// "multimedia instructions", MMI), a single-precision FPU (COP1) and a vector
// unit reachable from the instruction stream (COP2 = VU0 in macro mode).
//
// It is a relative of the VR4300 in tools/cpu/r4300, and this package is shaped
// like it, but four differences pervade the decoder and the interpreter, so the
// two are separate packages exactly as tools/cpu/arm and tools/cpu/arm60 are:
//
//   - Endianness. Instructions and data are little-endian; the VR4300 is big.
//   - Register width. The GPRs are 128 bits. Ordinary MIPS operations use the low
//     64 and leave the high 64 alone; only lq/sq and the MMI family touch the
//     whole register. See Quad.
//   - MMI. A whole second instruction space (primary opcode 0x1C, four sub-tables)
//     of packed 8/16/32-bit SIMD arithmetic, plus a second multiply accumulator
//     (HI1/LO1) and a shift-amount register (SA). See mmi.go.
//   - Floating point. COP1 is single-precision only and deliberately not IEEE 754:
//     there are no infinities, no NaNs and no denormals. See fpu.go.
//
// The branch delay slot and the branch-likely family behave as on the VR4300, and
// as there the load delay slot does not exist — the core interlocks, so a single
// register file suffices.
//
// The package follows the shape of the other CPU packages: a table-driven Decode
// producing an Inst with a Flow classification for the disassembler and tracer
// commands, and a separate CPU with a small Bus interface for the machine model
// (see tools/platform/ps2).
package r5900

import "fmt"

// Quad is one 128-bit general register. Lo is the low 64 bits — the half every
// ordinary MIPS instruction reads and writes. Hi is the upper 64, which only
// lq/sq and the MMI instructions touch; a 64-bit operation leaves it unchanged.
//
// Keeping the halves as two uint64s rather than a byte array is what lets the
// integer interpreter stay a straight port of the VR4300's: it reads .Lo and is
// otherwise identical.
type Quad struct {
	Lo, Hi uint64
}

func (q Quad) String() string { return fmt.Sprintf("%016X_%016X", q.Hi, q.Lo) }

// Flow classifies how control leaves an instruction. It mirrors the enum used by
// tools/cpu/r4300 and tools/cpu/mips so the shared codetrace/dis command
// skeletons apply.
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

// Inst is one decoded Emotion Engine instruction. Len is always 4 for a real
// instruction; a decode of a short slice yields Len 0 and FlowStop.
type Inst struct {
	Addr      uint32 // address of this instruction
	Len       int    // 4 for a decoded instruction, 0 when out of range
	Mnem      string // bare mnemonic, e.g. "daddu", "lq", "paddw"
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
