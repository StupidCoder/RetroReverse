// Package gekko decodes and executes code for the IBM "Gekko" — the processor at the
// heart of the Nintendo GameCube. It is a 32-bit big-endian PowerPC 750 (a G3, of the
// 603e line) with three additions Nintendo asked IBM for, and those additions are why
// this is a package of its own rather than a variant flag on a generic PowerPC core:
//
//   - Paired singles. The 32 floating-point registers are 64 bits wide, and in paired-
//     single mode each holds *two* 32-bit floats that arithmetic operates on at once.
//     A whole second instruction space (primary opcode 4) does SIMD on them. See ps.go.
//   - Quantised load/store. psq_l and psq_st move a pair of floats to and from memory
//     through a converter that can pack them as 8- or 16-bit integers, scaled by a power
//     of two — the format and the scale selected by one of eight graphics quantisation
//     registers (the GQRs). It is a compression unit in the load path, and it is how a
//     GameCube game keeps its vertices small. See ps.go.
//   - The locked cache. Half the L1 data cache can be dropped out of the coherency
//     protocol and mapped at a fixed address as a fast 16 KiB scratchpad, with a DMA
//     engine (driven through two SPRs) to move lines between it and main memory. It is
//     architectural state, not a transparent cache, so it lives in the CPU. See cache.go.
//
// Three things about PowerPC shape the decoder and the tracer, and they are exactly the
// places where an intuition carried over from the MIPS cores in this repository is wrong:
//
//   - There are no delay slots. Nothing executes after a taken branch. So Inst carries
//     no HasDelay and no Annul, and the tracer's descent is simpler than the VR4300's.
//   - Indirect branches can be conditional. `bclr`/`bcctr` branch to the link register
//     or the count register, and they take the same condition field as any other branch —
//     so "return" and "fall through" are not exclusive. An unconditional bclr is a
//     return; a *conditional* one continues as well, and the tracer must walk both. The
//     eight-value Flow enum absorbs this without extension.
//   - The condition register is eight independent 4-bit fields, any of which a compare
//     may target and any of which a branch may test — so a branch's meaning is not in its
//     opcode but in a field five bits away from it.
//
// The package follows the shape of the other CPU packages here: a table-driven Decode
// producing an Inst with a Flow classification for the disassembler and the code tracer,
// and a separate CPU with a small Bus interface for the machine model (tools/platform/gc).
package gekko

import "fmt"

// Flow classifies how control leaves an instruction. It mirrors the enum used by
// tools/cpu/r4300, tools/cpu/r5900 and tools/cpu/mips so the shared codetrace/dis
// command skeletons apply.
type Flow int

const (
	FlowSeq     Flow = iota // falls through to the next instruction
	FlowBranch              // conditional branch: continues AND may take Target
	FlowJump                // unconditional branch to Target, no fall-through
	FlowCall                // bl: calls Target, normally returns after it
	FlowReturn              // blr: path ends
	FlowIndJump             // bctr: target not statically known
	FlowIndCall             // bctrl/blrl: target unknown but returns, so continue
	FlowStop                // sc/rfi/illegal/truncated: treat as a stop
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

// Inst is one decoded Gekko instruction. Len is 4 for a real instruction; a decode of a
// short slice yields Len 0 and FlowStop.
//
// There is deliberately no HasDelay and no Annul: PowerPC has no delay slot, and the
// instruction after a taken branch does not execute.
type Inst struct {
	Addr      uint32 // address of this instruction
	Word      uint32 // the raw encoding, kept so a halt can name what it could not do
	Len       int    // 4 for a decoded instruction, 0 when out of range
	Mnem      string // bare mnemonic, e.g. "addi", "psq_l", "bdnzf"
	Text      string // formatted "mnem operands"
	Flow      Flow
	Target    uint32 // branch destination, valid when HasTarget
	HasTarget bool
}

func (in Inst) String() string {
	return fmt.Sprintf("$%08X: %s", in.Addr, in.Text)
}

// Field accessors. PowerPC numbers its instruction bits from the most significant end,
// so bit 0 is the top bit of the word; these convert once, here, so that nothing else
// in the package has to think about it.
func opcd(w uint32) uint32 { return w >> 26 }          // bits 0-5
func rs(w uint32) uint32   { return (w >> 21) & 31 }   // bits 6-10 (rD or rS or frD/frS)
func ra(w uint32) uint32   { return (w >> 16) & 31 }   // bits 11-15
func rb(w uint32) uint32   { return (w >> 11) & 31 }   // bits 16-20
func rc(w uint32) uint32   { return (w >> 6) & 31 }    // bits 21-25 (frC in the A-forms)
func xo10(w uint32) uint32 { return (w >> 1) & 0x3FF } // bits 21-30
func xo5(w uint32) uint32  { return (w >> 1) & 0x1F }  // bits 26-30
func xo6(w uint32) uint32  { return (w >> 1) & 0x3F }  // bits 25-30
func rcbit(w uint32) bool  { return w&1 != 0 }         // the record bit: update CR0 (or CR1)
func oe(w uint32) bool     { return w&(1<<10) != 0 }   // overflow enable
func lk(w uint32) bool     { return w&1 != 0 }         // link: write the return address to LR
func aa(w uint32) bool     { return w&2 != 0 }         // absolute: the target is not PC-relative
func simm(w uint32) int32  { return int32(int16(w)) }
func uimm(w uint32) uint32 { return w & 0xFFFF }
func crfD(w uint32) uint32 { return (w >> 23) & 7 }
func crfS(w uint32) uint32 { return (w >> 18) & 7 }
func shOf(w uint32) uint32 { return (w >> 11) & 31 }
func mbOf(w uint32) uint32 { return (w >> 6) & 31 }
func meOf(w uint32) uint32 { return (w >> 1) & 31 }

// sprOf reads the special-purpose-register number, whose two five-bit halves are stored
// swapped in the instruction word — a quirk of the encoding, and a reliable source of
// off-by-a-nibble bugs if it is done anywhere but here.
func sprOf(w uint32) uint32 {
	raw := (w >> 11) & 0x3FF
	return ((raw & 0x1F) << 5) | (raw >> 5)
}

// The special-purpose registers this core implements. The Gekko's own — the GQRs, HID2,
// the locked-cache DMA pair, the write-gather pipe — sit above the architected PowerPC
// ones, and are the reason a generic PowerPC SPR file would not do.
const (
	SPRXER               = 1
	SPRLR                = 8
	SPRCTR               = 9
	SPRDSISR             = 18
	SPRDAR               = 19
	SPRDEC               = 22
	SPRSDR1              = 25
	SPRSRR0              = 26
	SPRSRR1              = 27
	SPRSPRG0             = 272
	SPRSPRG1             = 273
	SPRSPRG2             = 274
	SPRSPRG3             = 275
	SPREAR               = 282
	SPRTBL               = 284 // written through mtspr; read through mftb
	SPRTBU               = 285
	SPRPVR               = 287
	SPRIBAT0U, SPRIBAT0L = 528, 529
	SPRIBAT1U, SPRIBAT1L = 530, 531
	SPRIBAT2U, SPRIBAT2L = 532, 533
	SPRIBAT3U, SPRIBAT3L = 534, 535
	SPRDBAT0U, SPRDBAT0L = 536, 537
	SPRDBAT1U, SPRDBAT1L = 538, 539
	SPRDBAT2U, SPRDBAT2L = 540, 541
	SPRDBAT3U, SPRDBAT3L = 542, 543

	// The Gekko's additions.
	SPRGQR0                      = 912 // ...through SPRGQR7 = 919
	SPRGQR7                      = 919
	SPRHID2                      = 920 // paired-single enable, locked-cache enable, write-gather enable
	SPRWPAR                      = 921 // write-gather pipe address
	SPRDMAU                      = 922 // locked-cache DMA: upper (address and length)
	SPRDMAL                      = 923 // locked-cache DMA: lower (address, trigger, direction)
	SPRHID0                      = 1008
	SPRHID1                      = 1009
	SPRIABR                      = 1010
	SPRHID4                      = 1011
	SPRDABR                      = 1013
	SPRL2CR                      = 1017
	SPRICTC                      = 1019
	SPRTHRM1, SPRTHRM2, SPRTHRM3 = 1020, 1021, 1022
)

// sprName gives the SPRs their architectural names, so a disassembly reads
// "mtspr LR, r0" rather than "mtspr 8, r0". An SPR with no name prints as a number,
// which is honest: the Gekko has undocumented ones, and inventing a name for one would
// be inventing a fact.
var sprName = map[uint32]string{
	SPRXER: "XER", SPRLR: "LR", SPRCTR: "CTR", SPRDSISR: "DSISR", SPRDAR: "DAR",
	SPRDEC: "DEC", SPRSDR1: "SDR1", SPRSRR0: "SRR0", SPRSRR1: "SRR1",
	SPRSPRG0: "SPRG0", SPRSPRG1: "SPRG1", SPRSPRG2: "SPRG2", SPRSPRG3: "SPRG3",
	SPREAR: "EAR", SPRTBL: "TBL", SPRTBU: "TBU", SPRPVR: "PVR",
	SPRIBAT0U: "IBAT0U", SPRIBAT0L: "IBAT0L", SPRIBAT1U: "IBAT1U", SPRIBAT1L: "IBAT1L",
	SPRIBAT2U: "IBAT2U", SPRIBAT2L: "IBAT2L", SPRIBAT3U: "IBAT3U", SPRIBAT3L: "IBAT3L",
	SPRDBAT0U: "DBAT0U", SPRDBAT0L: "DBAT0L", SPRDBAT1U: "DBAT1U", SPRDBAT1L: "DBAT1L",
	SPRDBAT2U: "DBAT2U", SPRDBAT2L: "DBAT2L", SPRDBAT3U: "DBAT3U", SPRDBAT3L: "DBAT3L",
	912: "GQR0", 913: "GQR1", 914: "GQR2", 915: "GQR3",
	916: "GQR4", 917: "GQR5", 918: "GQR6", 919: "GQR7",
	SPRHID2: "HID2", SPRWPAR: "WPAR", SPRDMAU: "DMAU", SPRDMAL: "DMAL",
	SPRHID0: "HID0", SPRHID1: "HID1", SPRIABR: "IABR", SPRHID4: "HID4",
	SPRDABR: "DABR", SPRL2CR: "L2CR", SPRICTC: "ICTC",
	SPRTHRM1: "THRM1", SPRTHRM2: "THRM2", SPRTHRM3: "THRM3",
}

func sprStr(n uint32) string {
	if s, ok := sprName[n]; ok {
		return s
	}
	return fmt.Sprintf("%d", n)
}

// Bits of the machine state register that the interpreter acts on.
const (
	MSRLE  uint32 = 1 << 0  // little-endian mode (a GameCube never sets it)
	MSRRI  uint32 = 1 << 1  // recoverable interrupt
	MSRDR  uint32 = 1 << 4  // data address translation on
	MSRIR  uint32 = 1 << 5  // instruction address translation on
	MSRIP  uint32 = 1 << 6  // exception vectors at 0xFFF00000 rather than 0x00000000
	MSRFE1 uint32 = 1 << 8  // floating-point exception mode 1
	MSRBE  uint32 = 1 << 9  // branch trace
	MSRSE  uint32 = 1 << 10 // single-step trace
	MSRFE0 uint32 = 1 << 11
	MSRME  uint32 = 1 << 12 // machine check enable
	MSRFP  uint32 = 1 << 13 // floating point available
	MSRPR  uint32 = 1 << 14 // problem state (user mode)
	MSREE  uint32 = 1 << 15 // external interrupts enabled
	MSRPOW uint32 = 1 << 18
)

// Bits of XER.
const (
	XERSO uint32 = 1 << 31 // summary overflow: sticky, cleared only by mtspr/mcrxr
	XEROV uint32 = 1 << 30
	XERCA uint32 = 1 << 29 // carry
)

// Bits of HID2 — the register that turns the Gekko's own hardware on.
const (
	HID2WPE uint32 = 1 << 30 // write-gather pipe enable
	HID2PSE uint32 = 1 << 29 // paired-single enable
	HID2LCE uint32 = 1 << 28 // locked-cache enable
)
