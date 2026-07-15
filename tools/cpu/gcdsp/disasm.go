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
	case op&0xFC00 == 0x1800:
		// Register-indirect load/store with an optional post-modify of the address register:
		// bit 9 = store, bits 8..7 = modify (00 none, 01 decrement, 10 increment, 11 +index).
		ar := regName((op >> 5) & 3)
		reg := regName(op & 0x1F)
		store := op&0x0200 != 0
		var suffix string
		switch op & 0x0180 {
		case 0x0080:
			suffix = "d" // post-decrement
		case 0x0100:
			suffix = "i" // post-increment
		case 0x0180:
			suffix = "n" // post-add index register
		}
		if store {
			return fmt.Sprintf("%-6s @%s, %s", "srr"+suffix, ar, reg), 1
		}
		return fmt.Sprintf("%-6s %s, @%s", "lrr"+suffix, reg, ar), 1

	// --- move register to register -------------------------------------------------------
	case op&0xFC00 == 0x1C00:
		return fmt.Sprintf("mrr    %s, %s", regName((op>>5)&0x1F), regName(op&0x1F)), 1

	// --- accumulator shift by immediate (signed 7-bit amount, +left/-right) ---------------
	case op&0xFE00 == 0x1400:
		r := (op >> 8) & 1
		arith := op&0x0080 != 0
		amt := int(op & 0x7F)
		if amt&0x40 != 0 {
			amt -= 0x80
		}
		var mn string
		switch {
		case arith && amt >= 0:
			mn = "asl"
		case arith:
			mn = "asr"
			amt = -amt
		case amt >= 0:
			mn = "lsl"
		default:
			mn = "lsr"
			amt = -amt
		}
		return fmt.Sprintf("%-6s ac%d, #%d", mn, r, amt), 1

	// --- load short immediate (8-bit signed into register 0x18..0x1F) ---------------------
	case op&0xF800 == 0x0800:
		return fmt.Sprintf("lris   %s, #0x%02X", regName(0x18+((op>>8)&7)), op&0xFF), 1

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
		return fmt.Sprintf("andcf  ac%d.m, #0x%04X", (op>>8)&1, next()), 2

	// --- branches, calls, returns --------------------------------------------------------
	case op&0xFFF0 == 0x0290:
		return fmt.Sprintf("%-6s 0x%04X", cond("jmp", op&0xF), next()), 2
	case op&0xFFF0 == 0x02B0:
		return fmt.Sprintf("%-6s 0x%04X", cond("call", op&0xF), next()), 2
	case op&0xFFF0 == 0x02D0:
		return cond("ret", op&0xF), 1
	case op&0xFFF0 == 0x02F0:
		return cond("rti", op&0xF), 1
	case op&0xFF00 == 0x1700:
		mn := "jmpr"
		if op&0x10 != 0 { // bit 4 = call flag
			mn = "callr"
		}
		return fmt.Sprintf("%-6s %s", cond(mn, op&0xF), regName((op>>5)&7)), 1

	// --- interrupt-enable / misc single-word control -------------------------------------
	case op == 0x1201: // (some ucodes) — fall through to raw if unmatched

	// --- short-immediate arithmetic (single word, an 8-bit immediate into the middle) -----
	case op&0xFE00 == 0x0400:
		return fmt.Sprintf("addis  ac%d, #0x%02X", (op>>8)&1, op&0xFF), 1
	case op&0xFE00 == 0x0600:
		return fmt.Sprintf("cmpis  ac%d, #0x%02X", (op>>8)&1, op&0xFF), 1
	}

	// --- arithmetic / logic / multiply / move ops (0x3000 and up): operation in the high bits,
	// a parallel extension in the low byte (for the ops that carry one). Single-word. --------
	if op >= 0x3000 {
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

// arithMnemonic names the arithmetic/logic/multiply/move op in the high bits of a 0x3000+
// instruction. The opcodes and masks are the documented set (gamecube-tools' opcode table);
// anything unmatched is shown by its high byte so a run is still readable and correctly sized.
func arithMnemonic(op uint16) string {
	d := (op >> 8) & 1 // most accumulator-writing ops select the accumulator in bit 8
	switch {
	// The standalone mode ops occupy fixed high bytes and must be matched first.
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
	case op&0xFF00 == 0x8400:
		return "clrp"

	// Logic ops on an accumulator middle word.
	case op&0xFE80 == 0x3280:
		return fmt.Sprintf("not    ac%d.m", d)
	case op&0xFC80 == 0x3000:
		return fmt.Sprintf("xorr   ac%d.m, ax%d.h", d, (op>>9)&1)
	case op&0xFC80 == 0x3400:
		return fmt.Sprintf("andr   ac%d.m, ax%d.h", d, (op>>9)&1)
	case op&0xFC80 == 0x3800:
		return fmt.Sprintf("orr    ac%d.m, ax%d.h", d, (op>>9)&1)

	// clr / cmp / tst.
	case op&0xF700 == 0x8100:
		return fmt.Sprintf("clr    ac%d", (op>>11)&1)
	case op&0xFF00 == 0x8200:
		return "cmp"
	case op&0xFE00 == 0x8600:
		return fmt.Sprintf("tst    ac%d", d)

	// The multiply / multiply-accumulate family.
	case op&0xF700 == 0x9000:
		return fmt.Sprintf("mul    ax%d", (op>>11)&1)
	case op&0xF600 == 0x9200:
		return fmt.Sprintf("mulmvz ac%d", d)
	case op&0xF600 == 0x9400:
		return fmt.Sprintf("mulac  ac%d", d)
	case op&0xF600 == 0x9600:
		return fmt.Sprintf("mulmv  ac%d", d)
	case op&0xE700 == 0xA000:
		return "mulx"
	case op&0xE600 == 0xA200:
		return fmt.Sprintf("mulxmvz ac%d", d)
	case op&0xE600 == 0xA400:
		return fmt.Sprintf("mulxac ac%d", d)
	case op&0xE600 == 0xA600:
		return fmt.Sprintf("mulxmv ac%d", d)
	case op&0xE700 == 0xC000:
		return "mulc"
	case op&0xE600 == 0xC400:
		return fmt.Sprintf("mulcac ac%d", d)
	case op&0xFC00 == 0xE000:
		return "maddx"
	case op&0xFC00 == 0xE400:
		return "msubx"
	case op&0xFC00 == 0xE800:
		return "maddc"
	case op&0xFC00 == 0xEC00:
		return "msubc"
	case op&0xFE00 == 0xF200:
		return fmt.Sprintf("madd   ax%d", d)
	case op&0xFE00 == 0xF600:
		return fmt.Sprintf("msub   ax%d", d)

	// The accumulator arithmetic and moves.
	case op&0xF800 == 0x4000:
		return fmt.Sprintf("addr   ac%d", d)
	case op&0xFC00 == 0x4800:
		return fmt.Sprintf("addax  ac%d, ax%d", d, (op>>9)&1)
	case op&0xFE00 == 0x4C00:
		return fmt.Sprintf("add    ac%d", d)
	case op&0xFE00 == 0x4E00:
		return fmt.Sprintf("addp   ac%d", d)
	case op&0xF800 == 0x5000:
		return fmt.Sprintf("subr   ac%d", d)
	case op&0xFC00 == 0x5800:
		return fmt.Sprintf("subax  ac%d, ax%d", d, (op>>9)&1)
	case op&0xFE00 == 0x5C00:
		return fmt.Sprintf("sub    ac%d", d)
	case op&0xF800 == 0x6000:
		return fmt.Sprintf("movr   ac%d, r%d", d, 0x18+((op>>9)&3))
	case op&0xFC00 == 0x6800:
		return fmt.Sprintf("movax  ac%d, ax%d", d, (op>>9)&1)
	case op&0xFE00 == 0x6C00:
		return fmt.Sprintf("mov    ac%d", d)
	case op&0xFE00 == 0x6E00:
		return fmt.Sprintf("movp   ac%d", d)
	case op&0xFC00 == 0x7000:
		return fmt.Sprintf("addaxl ac%d, ax%d", d, (op>>9)&1)
	case op&0xFC00 == 0xF800:
		return fmt.Sprintf("addpaxz ac%d, ax%d", (op>>9)&1, d)
	case op&0xFE00 == 0xF000:
		return fmt.Sprintf("lsl16  ac%d", d)
	case op&0xFE00 == 0xF400:
		return fmt.Sprintf("lsr16  ac%d", d)
	}
	return fmt.Sprintf("op8_%02X", op>>8)
}

// extMnemonic names the parallel extension carried in the low byte — a simultaneous load, store,
// register move, or address-register step. The encodings are the documented opcodes_ext table.
func extMnemonic(ext uint16) string {
	switch {
	case ext == 0x00:
		return ""
	case ext&0xC0 == 0x80: // ls/sl and their n/m variants: load AR0, store AR3
		name := "ls"
		if ext&0x02 != 0 {
			name = "sl"
		}
		if ext&0x04 != 0 {
			name += "n"
		}
		if ext&0x08 != 0 {
			name += "m"
		}
		return fmt.Sprintf("%s r%d, ac%d.m", name, 0x18+((ext>>4)&3), ext&1)
	case ext&0xC0 == 0xC0: // ld/ldx: dual load through address registers
		return fmt.Sprintf("ld r?, r?, @ar%d", ext&3)
	case ext&0xC4 == 0x40: // l/ln: load a register
		name := "l"
		if ext&0x04 != 0 {
			name = "ln"
		}
		return fmt.Sprintf("%s r%d, @ar%d", name, 0x18+((ext>>3)&7), ext&3)
	case ext&0xE4 == 0x20: // s/sn: store a register
		name := "s"
		if ext&0x04 != 0 {
			name = "sn"
		}
		return fmt.Sprintf("%s @ar%d, r%d", name, ext&3, 0x1C+((ext>>3)&3))
	case ext&0xF0 == 0x10: // mv: register move
		return fmt.Sprintf("mv r%d, r%d", 0x18+((ext>>2)&3), 0x1C+(ext&3))
	case ext&0xFC == 0x04:
		return fmt.Sprintf("dr ar%d", ext&3)
	case ext&0xFC == 0x08:
		return fmt.Sprintf("ir ar%d", ext&3)
	case ext&0xFC == 0x0C:
		return fmt.Sprintf("nr ar%d", ext&3)
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
