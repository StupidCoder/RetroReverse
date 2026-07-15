package gcdsp

// disasm.go decodes one GameCube-DSP instruction to text and, as importantly, tells the
// caller how many 16-bit words it spans — most instructions are one word, a handful carry a
// second word of immediate or address. The opcode encodings are the documented instruction
// set; where a field's meaning is well established it is named, and where a bit pattern is not
// yet identified it is shown raw as `.word` so a decode never silently mis-sizes the stream.
//
// The instruction set has two shapes. The control, load/store, branch and immediate ops in the
// low opcode space (below 0x4000) use the whole 16 bits. The arithmetic ops (0x4000 and up)
// keep their operation in the high bits and leave the low byte free for a "parallel extension"
// — a second, simultaneous load/store/move — so those are always one word, with the extension
// decoded and appended after a colon.

import (
	"fmt"
	"strings"
)

// regName gives the assembler name of a register-file index.
func regName(r uint16) string {
	switch {
	case r < 4:
		return fmt.Sprintf("ar%d", r)
	case r < 8:
		return fmt.Sprintf("ix%d", r-4)
	case r < 12:
		return fmt.Sprintf("wr%d", r-8)
	case r < 16:
		return fmt.Sprintf("st%d", r-12)
	}
	switch r {
	case regAC0H:
		return "ac0.h"
	case regAC1H:
		return "ac1.h"
	case regCONFIG:
		return "config"
	case regSR:
		return "sr"
	case regPRODL:
		return "prod.l"
	case regPRODM1:
		return "prod.m1"
	case regPRODH:
		return "prod.h"
	case regPRODM2:
		return "prod.m2"
	case regAX0L:
		return "ax0.l"
	case regAX1L:
		return "ax1.l"
	case regAX0H:
		return "ax0.h"
	case regAX1H:
		return "ax1.h"
	case regAC0L:
		return "ac0.l"
	case regAC1L:
		return "ac1.l"
	case regAC0M:
		return "ac0.m"
	case regAC1M:
		return "ac1.m"
	}
	return fmt.Sprintf("r%d", r)
}

// condName names a 4-bit branch condition. 0xF is unconditional (empty suffix); the rest carry
// their documented mnemonic, with an unidentified code shown as its number.
func condName(cc uint16) string {
	switch cc {
	case 0x0:
		return "ge"
	case 0x1:
		return "l"
	case 0x2:
		return "g"
	case 0x3:
		return "le"
	case 0x4:
		return "nz"
	case 0x5:
		return "z"
	case 0x6:
		return "nc"
	case 0x7:
		return "c"
	case 0xC:
		return "lnz"
	case 0xD:
		return "lz"
	case 0xE:
		return "o"
	case 0xF:
		return "" // always
	}
	return fmt.Sprintf("?%X", cc)
}

// Disasm decodes the instruction at word address pc. read returns the instruction word at a
// given address (the caller supplies IRAM/IROM access). It returns the text and the number of
// words the instruction occupies.
func Disasm(read func(uint16) uint16, pc uint16) (text string, words uint16) {
	op := read(pc)
	next := func() uint16 { return read(pc + 1) }

	// A cond-branch mnemonic: name plus condition suffix, e.g. "jmp" / "jmpnz".
	cond := func(base string, cc uint16) string {
		return base + condName(cc)
	}

	switch {
	// --- no-operation and halt -----------------------------------------------------------
	case op == 0x0000:
		return "nop", 1
	case op == 0x0021:
		return "halt", 1

	// --- address-register arithmetic -----------------------------------------------------
	case op&0xFFFC == 0x0004:
		return fmt.Sprintf("dar    %s", regName(op&3)), 1
	case op&0xFFFC == 0x0008:
		return fmt.Sprintf("iar    %s", regName(op&3)), 1
	case op&0xFFFC == 0x000C:
		return fmt.Sprintf("subarn %s", regName(op&3)), 1
	case op&0xFFF0 == 0x0010:
		return fmt.Sprintf("addarn %s, %s", regName(op&3), regName(4+((op>>2)&3))), 1

	// --- hardware loops ------------------------------------------------------------------
	case op&0xFFE0 == 0x0040:
		return fmt.Sprintf("loop   %s", regName(op&0x1F)), 1
	case op&0xFFE0 == 0x0060:
		return fmt.Sprintf("bloop  %s, 0x%04X", regName(op&0x1F), next()), 2
	case op&0xFF00 == 0x1000:
		return fmt.Sprintf("loopi  #0x%02X", op&0xFF), 1
	case op&0xFF00 == 0x1100:
		return fmt.Sprintf("bloopi #0x%02X, 0x%04X", op&0xFF, next()), 2

	// --- status-bit set/clear ------------------------------------------------------------
	case op&0xFF00 == 0x1200:
		return fmt.Sprintf("sbclr  #%d", op&0xFF), 1
	case op&0xFF00 == 0x1300:
		return fmt.Sprintf("sbset  #%d", op&0xFF), 1

	// --- register immediate / load / store (low space) -----------------------------------
	case op&0xFFE0 == 0x0080:
		return fmt.Sprintf("lri    %s, #0x%04X", regName(op&0x1F), next()), 2
	case op&0xFFE0 == 0x00A0:
		return fmt.Sprintf("lrr?   %s <- (0x%04X)", regName(op&0x1F), next()), 2
	case op&0xFFE0 == 0x00C0:
		return fmt.Sprintf("lr     %s, @0x%04X", regName(op&0x1F), next()), 2
	case op&0xFFE0 == 0x00E0:
		return fmt.Sprintf("sr     @0x%04X, %s", next(), regName(op&0x1F)), 2
	case op&0xFF00 == 0x1600:
		return fmt.Sprintf("si     @0x%02X, #0x%04X", op&0xFF, next()), 2
	case op&0xFF00 == 0x1B00:
		return fmt.Sprintf("srr    @%s, %s", regName((op>>5)&3), regName(op&0x1F)), 1
	case op&0xFF00 == 0x1900:
		return fmt.Sprintf("lrr    %s, @%s", regName(op&0x1F), regName((op>>5)&3)), 1
	case op&0xFF00 == 0x1A00:
		return fmt.Sprintf("lrrd/lrri %s, @%s", regName(op&0x1F), regName((op>>5)&3)), 1

	// --- move register to register -------------------------------------------------------
	case op&0xFC00 == 0x1C00:
		return fmt.Sprintf("mrr    %s, %s", regName((op>>5)&0x1F), regName(op&0x1F)), 1

	// --- load/store a short register to a hardware/data address (0xFF00|M) ----------------
	case op&0xF800 == 0x2000:
		return fmt.Sprintf("lrs    %s, @0xFF%02X", regName(0x18+((op>>8)&7)), op&0xFF), 1
	case op&0xF800 == 0x2800:
		return fmt.Sprintf("srs    @0xFF%02X, %s", op&0xFF, regName(0x18+((op>>8)&7))), 1

	// --- immediate arithmetic to accumulator (two-word) ----------------------------------
	// These sit below the branch group in the 0x02xx/0x03xx space; each takes a data word.
	case op&0xFEFF == 0x0200:
		return fmt.Sprintf("addi   ac%d, #0x%04X", (op>>8)&1, next()), 2
	case op&0xFEFF == 0x0220:
		return fmt.Sprintf("xori   ac%d.m, #0x%04X", (op>>8)&1, next()), 2
	case op&0xFEFF == 0x0240:
		return fmt.Sprintf("andi   ac%d.m, #0x%04X", (op>>8)&1, next()), 2
	case op&0xFEFF == 0x0260:
		return fmt.Sprintf("ori    ac%d.m, #0x%04X", (op>>8)&1, next()), 2
	case op&0xFEFF == 0x0280:
		return fmt.Sprintf("cmpi   ac%d, #0x%04X", (op>>8)&1, next()), 2
	case op&0xFEFF == 0x02A0:
		return fmt.Sprintf("andf   ac%d.m, #0x%04X", (op>>8)&1, next()), 2
	case op&0xFEFF == 0x02C0:
		return fmt.Sprintf("orf    ac%d.m, #0x%04X", (op>>8)&1, next()), 2

	// --- branches, calls, returns --------------------------------------------------------
	case op&0xFFF0 == 0x0290:
		return fmt.Sprintf("%-6s 0x%04X", cond("jmp", op&0xF), next()), 2
	case op&0xFFF0 == 0x02B0:
		return fmt.Sprintf("%-6s 0x%04X", cond("call", op&0xF), next()), 2
	case op&0xFFF0 == 0x02D0:
		return cond("ret", op&0xF), 1
	case op&0xFFF0 == 0x02F0:
		return cond("rti", op&0xF), 1
	case op&0xFFF0 == 0x1700:
		return fmt.Sprintf("%-6s %s", cond("jmpr", op&0xF), regName((op>>5)&7)), 1
	case op&0xFFF0 == 0x1710:
		return fmt.Sprintf("%-6s %s", cond("callr", op&0xF), regName((op>>5)&7)), 1

	// --- interrupt-enable / misc single-word control -------------------------------------
	case op == 0x1201: // (some ucodes) — fall through to raw if unmatched
	}

	// --- arithmetic ops (0x4000 and up): operation in the high bits, low byte a parallel
	// extension. All single-word. ---------------------------------------------------------
	if op >= 0x4000 {
		main := arithMnemonic(op)
		ext := op & 0xFF
		if ext != 0 {
			return fmt.Sprintf("%-18s : %s", main, extMnemonic(ext)), 1
		}
		return main, 1
	}

	// Unidentified low-space word: show it raw so the stream stays word-aligned.
	return fmt.Sprintf(".word  0x%04X", op), 1
}

// arithMnemonic names the arithmetic op in the high bits of a 0x4000+ instruction. The common
// mixing-ucode ops are named; anything else is shown by its high byte so a run is still
// readable and correctly sized.
func arithMnemonic(op uint16) string {
	d := (op >> 8) & 1 // most of these select an accumulator in bit 8
	switch {
	// The standalone mode ops occupy fixed high bytes 0x8A..0x8F and must be matched exactly,
	// before the accumulator-paired ops below, whose masks would otherwise swallow them.
	case op&0xFF00 == 0x8A00:
		return "m2"
	case op&0xFF00 == 0x8B00:
		return "m0"
	case op&0xFF00 == 0x8C00:
		return "clr15"
	case op&0xFF00 == 0x8D00:
		return "set15"
	case op&0xFF00 == 0x8E00:
		return "set40"
	case op&0xFF00 == 0x8F00:
		return "set16"
	case op&0xFF00 == 0x8000:
		return "nx"
	case op&0xF700 == 0x8100: // 0x8100 / 0x8900, accumulator in bit 11
		return fmt.Sprintf("clr    ac%d", (op>>11)&1)
	case op&0xFF00 == 0x8200:
		return "cmp"
	case op&0xFF00 == 0x8600:
		return "tst    ac0"
	case op&0xFF00 == 0x8700:
		return "tst    ac1"
	case op&0xF800 == 0x9000:
		return fmt.Sprintf("mul    ax%d", (op>>11)&1)
	case op&0xF800 == 0x9800:
		return fmt.Sprintf("mulmv? ax%d", (op>>11)&1)
	case op&0xFC00 == 0xA000:
		return "mulx"
	case op&0xFC00 == 0xA400:
		return "mulxmv?"
	case op&0xF000 == 0xB000:
		return "mulc?"
	case op&0xF800 == 0xC000:
		return fmt.Sprintf("mulcac? ac%d", (op>>11)&1)
	case op&0xFC00 == 0x4000:
		return fmt.Sprintf("addr   ac%d", d)
	case op&0xFC00 == 0x4400:
		return fmt.Sprintf("addax  ac%d", d)
	case op&0xFE00 == 0x4800:
		return fmt.Sprintf("add    ac%d", d)
	case op&0xFE00 == 0x4A00:
		return fmt.Sprintf("addp   ac%d", d)
	case op&0xFC00 == 0x5000:
		return fmt.Sprintf("subr   ac%d", d)
	case op&0xFC00 == 0x5400:
		return fmt.Sprintf("subax  ac%d", d)
	case op&0xFE00 == 0x5800:
		return fmt.Sprintf("sub    ac%d", d)
	case op&0xFC00 == 0x6000:
		return fmt.Sprintf("movr   ac%d", d)
	case op&0xFC00 == 0x6400:
		return fmt.Sprintf("movax  ac%d", d)
	case op&0xFE00 == 0x6800:
		return fmt.Sprintf("addaxl ac%d", d)
	case op&0xFE00 == 0x6C00:
		return fmt.Sprintf("mov    ac%d", d)
	case op&0xFC00 == 0x7000:
		return fmt.Sprintf("addpaxz ac%d", d)
	case op&0xFE00 == 0x7800:
		return fmt.Sprintf("decm/inc ac%d", d)
	case op&0xF000 == 0xE000:
		return "maddx?"
	case op&0xF000 == 0xF000:
		return "madd?"
	}
	return fmt.Sprintf("op8_%02X", op>>8)
}

// extMnemonic names the parallel extension carried in the low byte of an arithmetic op — a
// simultaneous load, store, or register move. The common forms are named; the rest shown raw.
func extMnemonic(ext uint16) string {
	switch {
	case ext == 0x00:
		return ""
	case ext&0xC0 == 0x40 && ext&0x30 != 0x30:
		// 01ssdddd style loads/stores through address registers — shown generically.
		return fmt.Sprintf("mv? r%d", ext&0x0F)
	case ext&0xF0 == 0x80:
		return fmt.Sprintf("ls @ar%d", ext&3)
	}
	return fmt.Sprintf("ext 0x%02X", ext)
}

// branchTarget returns the absolute word address a two-word branch/loop at pc points to, and
// whether the instruction is such a branch. It is the check DisasmValidate uses to confirm a
// decode is self-consistent: every jump must land on a decoded instruction boundary.
func branchTarget(read func(uint16) uint16, pc uint16) (target uint16, isBranch bool) {
	op := read(pc)
	switch {
	case op&0xFFF0 == 0x0290, // jmp cc
		op&0xFFF0 == 0x02B0, // call cc
		op&0xFFE0 == 0x0060, // bloop
		op&0xFF00 == 0x1100: // bloopi
		return read(pc + 1), true
	}
	return 0, false
}

// DisasmValidate decodes the whole image and reports every static branch whose target does not
// land on an instruction boundary, plus whether the stream decodes without overrunning its
// end. A coherent image produces no misaligned targets — which is what confirms the decoder's
// instruction lengths are right without a reference disassembly to compare against.
func DisasmValidate(read func(uint16) uint16, nWords uint16) (misaligned []uint16, boundaries int, overrun bool) {
	starts := make(map[uint16]bool)
	for w := uint16(0); w < nWords; {
		starts[w] = true
		_, span := Disasm(read, w)
		if w+span > nWords {
			overrun = true
		}
		w += span
	}
	for w := uint16(0); w < nWords; {
		if t, ok := branchTarget(read, w); ok && t < nWords && !starts[t] {
			misaligned = append(misaligned, w)
		}
		_, span := Disasm(read, w)
		w += span
	}
	return misaligned, len(starts), overrun
}

// DisasmRange disassembles n words starting at pc, returning a listing. It is the harness the
// tests and the disgcdsp tool share.
func DisasmRange(read func(uint16) uint16, pc uint16, nWords uint16) string {
	var b strings.Builder
	for w := uint16(0); w < nWords; {
		addr := pc + w
		text, span := Disasm(read, addr)
		raw := read(addr)
		if span == 2 {
			fmt.Fprintf(&b, "%04X  %04X %04X  %s\n", addr, raw, read(addr+1), text)
		} else {
			fmt.Fprintf(&b, "%04X  %04X       %s\n", addr, raw, text)
		}
		w += span
	}
	return b.String()
}
