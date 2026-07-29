// Package sh4 decodes and executes Hitachi SH-4 code — the SH7091 at the heart
// of the Sega Dreamcast. It is a 32-bit little-endian RISC with a fixed 16-bit
// instruction word, sixteen general registers of which R0-R7 are banked (SR.RB
// selects the bank in privileged mode), the MACH/MACL multiply-accumulate pair,
// a procedure register PR that holds return addresses, and a floating-point
// unit with two register banks of sixteen 32-bit registers (FPSCR.FR swaps
// which bank is FR0-FR15 and which is XF0-XF15).
//
// Three properties pervade the decoder and the interpreter:
//
//   - Delay slots. Every unconditional transfer (bra, braf, bsr, bsrf, jmp,
//     jsr, rts, rte) executes the following instruction before the transfer,
//     and the conditional branches come in both undelayed (bt, bf) and delayed
//     (bt/s, bf/s) forms. Inst.HasDelay marks the delayed ones; an interrupt is
//     never accepted between a delayed transfer and its slot.
//
//   - PC-relative literal pools. The only way to build a 32-bit constant in one
//     instruction is mov.w/mov.l @(disp,PC) — a load from a pool of data words
//     the compiler drops between functions. Decode resolves the pool address
//     into Inst.LitAddr/LitSize so the tracer can mark pools as data instead of
//     misdecoding them as code. The PC base is the instruction address plus 4,
//     and the mov.l/mova form clears the base's low two bits.
//
//   - Mode-dependent meaning. The FPU halfwords name single-precision registers
//     but execute on doubles when FPSCR.PR is set, and fmov moves 64 bits when
//     FPSCR.SZ is set. A static decoder cannot know either bit, so the listing
//     always shows the single-precision reading (the convention SH toolchains
//     use); the interpreter applies the live FPSCR.
//
// The instruction map is transcribed from the Hitachi/Renesas SH-4 software
// manual's encoding tables (platform documentation, not game-derived). The
// package follows the shape of the other CPU packages: a pure Decode producing
// an Inst with a Flow classification for the disassembler and tracer commands,
// and a separate CPU with a small Bus interface for the machine model (see
// tools/platform/dc). Unimplemented encodings call Halt with the offending
// halfword, so gaps are explicit rather than silently wrong.
package sh4

import "fmt"

// Flow classifies how control leaves an instruction. It mirrors the enum used
// by tools/cpu/mips and tools/cpu/arm so the shared codetrace/dis command
// skeletons apply.
//
// SH-4 mapping:
//
//	bt / bf / bt/s / bf/s   FlowBranch  (conditional, PC-relative)
//	bra                     FlowJump    (unconditional, PC-relative)
//	bsr                     FlowCall    (call, PC-relative, returns)
//	braf Rn                 FlowIndJump (PC + register, not statically known)
//	bsrf Rn                 FlowIndCall (call through PC + register)
//	jmp @Rn                 FlowIndJump
//	jsr @Rn                 FlowIndCall
//	rts                     FlowReturn  (jump to PR: path ends)
//	rte / trapa / sleep     FlowStop
type Flow int

const (
	FlowSeq     Flow = iota // falls through to the next instruction
	FlowBranch              // conditional branch: continues AND may take Target
	FlowJump                // unconditional jump to Target, no fall-through
	FlowCall                // bsr: calls Target, normally returns after it
	FlowReturn              // rts: path ends
	FlowIndJump             // jmp @Rn / braf: target not statically known
	FlowIndCall             // jsr @Rn / bsrf: target unknown but returns
	FlowStop                // rte/trapa/sleep/illegal/truncated: treat as a stop
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

// Inst is one decoded SH-4 instruction. Len is always 2 for a real
// instruction; a decode of a short slice yields Len 0 and FlowStop.
type Inst struct {
	Addr      uint32 // address of this instruction
	Word      uint16 // the raw halfword, kept so a halt can name what it couldn't do
	Len       int    // 2 for a decoded instruction, 0 when out of range
	Mnem      string // bare mnemonic, e.g. "mov.l", "bf/s", "fmac"
	Text      string // formatted "mnem operands"
	Flow      Flow
	Target    uint32 // branch/jump/call destination, valid when HasTarget
	HasTarget bool
	HasDelay  bool // this instruction is followed by a delay slot

	// LitAddr/LitSize name the literal-pool word a PC-relative load reaches:
	// mov.w @(disp,PC) (LitSize 2), mov.l @(disp,PC) and mova (LitSize 4).
	// The tracer treats the pool word as data and a barrier to fall-through
	// decoding. LitSize is 0 for every other instruction.
	LitAddr uint32
	LitSize int
}

func (in Inst) String() string {
	return fmt.Sprintf("$%08X: %s", in.Addr, in.Text)
}

// litW resolves the pool address of mov.w @(disp,PC),Rn: base PC+4, halfword
// scaled. litL is the mov.l/mova form: the base additionally clears its low
// two bits, so the reach does not depend on which half of a longword the
// instruction sits in. Decode and the interpreter share these two helpers so
// the alignment quirk lives in exactly one place.
func litW(addr uint32, disp uint32) uint32 { return addr + 4 + disp*2 }
func litL(addr uint32, disp uint32) uint32 { return (addr+4)&^3 + disp*4 }
