package sh4

// decode.go turns a 16-bit halfword into an Inst. It is pure: no CPU, no
// memory, no state.
//
// The SH-4 encoding is nibble-structured: the top four bits select a group,
// and within a group the operand registers sit in fixed nibbles (n at bits
// 8-11, m at bits 4-7) while the remaining nibbles select the member. Groups
// 0x1, 0x5, 0x7, 0x9, 0xA, 0xB, 0xD and 0xE are a single format each; groups
// 0x2, 0x3 and 0x6 sub-decode on the low nibble; groups 0x0 and 0x4 need the
// low byte (with the register-bank forms keying on bit 7); groups 0x8 and 0xC
// sub-decode on bits 8-11 because their low byte is an immediate. Group 0xF is
// the FPU space and lives in decode_fpu.go, the same split the other packages
// give an extra instruction space of its own.
//
// The dispatch is a switch ladder in exactly the manual's table order, the
// organisation tools/cpu/arm uses for Thumb's formats, because SH-4's holes
// are irregular enough that a flat 65536-entry table would just move the
// ladder into a generator.

import (
	"encoding/binary"
	"fmt"
)

// Decode reads one instruction from the front of code, which is little-endian.
func Decode(code []byte, addr uint32) Inst {
	if len(code) < 2 {
		return Inst{Addr: addr, Len: 0, Mnem: ".hword", Text: ".hword ; truncated", Flow: FlowStop}
	}
	return DecodeHalfword(binary.LittleEndian.Uint16(code), addr)
}

// DecodeHalfword decodes one already-loaded instruction halfword.
func DecodeHalfword(h uint16, addr uint32) Inst {
	in := Inst{Addr: addr, Word: h, Len: 2, Flow: FlowSeq}
	decode(&in, h, addr)
	if in.Mnem == "" {
		in.Mnem = ".hword"
		in.Text = fmt.Sprintf(".hword 0x%04X", h)
		in.Flow = FlowStop // an unknown encoding ends a traced path rather than inventing one
	}
	return in
}

// set fills in the mnemonic and the rendered text together, so the two cannot drift.
func (in *Inst) set(mnem, format string, args ...interface{}) {
	in.Mnem = mnem
	if format == "" {
		in.Text = mnem
		return
	}
	in.Text = mnem + " " + fmt.Sprintf(format, args...)
}

func r(n uint32) string { return fmt.Sprintf("r%d", n) }

// bank renders a banked general register, the other bank's R0-R7 reachable
// from privileged code: r3_bank.
func bank(m uint32) string { return fmt.Sprintf("r%d_bank", m&7) }

// rn and rm extract the two register nibbles.
func rn(h uint16) uint32 { return uint32(h>>8) & 0xF }
func rm(h uint16) uint32 { return uint32(h>>4) & 0xF }

// s8 sign-extends the low byte, s12 the low twelve bits.
func s8(h uint16) int32  { return int32(int8(h)) }
func s12(h uint16) int32 { return int32(int16(h<<4)) >> 4 }

// branch fills the common fields of a PC-relative transfer: target PC+4+disp,
// where disp is already scaled to bytes.
func (in *Inst) branch(flow Flow, disp int32, delayed bool) {
	in.Flow = flow
	in.Target = uint32(int32(in.Addr+4) + disp)
	in.HasTarget = true
	in.HasDelay = delayed
}

func decode(in *Inst, h uint16, addr uint32) {
	switch h >> 12 {
	case 0x0:
		decode0(in, h)
	case 0x1: // mov.l Rm,@(disp,Rn)
		in.set("mov.l", "%s, @(0x%X,%s)", r(rm(h)), uint32(h&0xF)*4, r(rn(h)))
	case 0x2:
		decode2(in, h)
	case 0x3:
		decode3(in, h)
	case 0x4:
		decode4(in, h)
	case 0x5: // mov.l @(disp,Rm),Rn
		in.set("mov.l", "@(0x%X,%s), %s", uint32(h&0xF)*4, r(rm(h)), r(rn(h)))
	case 0x6:
		decode6(in, h)
	case 0x7: // add #imm,Rn
		in.set("add", "#%d, %s", s8(h), r(rn(h)))
	case 0x8:
		decode8(in, h)
	case 0x9: // mov.w @(disp,PC),Rn — a literal-pool load
		in.set("mov.w", "@(0x%X,pc), %s", uint32(h&0xFF)*2, r(rn(h)))
		in.LitAddr, in.LitSize = litW(addr, uint32(h&0xFF)), 2
	case 0xA: // bra
		in.set("bra", "")
		in.branch(FlowJump, s12(h)*2, true)
		in.Text = fmt.Sprintf("bra 0x%08X", in.Target)
	case 0xB: // bsr
		in.set("bsr", "")
		in.branch(FlowCall, s12(h)*2, true)
		in.Text = fmt.Sprintf("bsr 0x%08X", in.Target)
	case 0xC:
		decodeC(in, h, addr)
	case 0xD: // mov.l @(disp,PC),Rn — a literal-pool load
		in.set("mov.l", "@(0x%X,pc), %s", uint32(h&0xFF)*4, r(rn(h)))
		in.LitAddr, in.LitSize = litL(addr, uint32(h&0xFF)), 4
	case 0xE: // mov #imm,Rn
		in.set("mov", "#%d, %s", s8(h), r(rn(h)))
	case 0xF:
		decodeFPU(in, h) // the FPU space — see decode_fpu.go
	}
}

// decode0 covers group 0000, dispatching on the low nibble; the middle nibble
// is a register for the @(R0,...) moves, mul.l and mac.l, and a selector for
// everything else.
func decode0(in *Inst, h uint16) {
	n, m := rn(h), rm(h)
	switch h & 0xF {
	case 0x2: // stc <ctrl>,Rn
		switch {
		case m == 0:
			in.set("stc", "sr, %s", r(n))
		case m == 1:
			in.set("stc", "gbr, %s", r(n))
		case m == 2:
			in.set("stc", "vbr, %s", r(n))
		case m == 3:
			in.set("stc", "ssr, %s", r(n))
		case m == 4:
			in.set("stc", "spc, %s", r(n))
		case m >= 8: // stc Rm_bank,Rn
			in.set("stc", "%s, %s", bank(m), r(n))
		}
	case 0x3:
		switch m {
		case 0x0: // bsrf Rn: PC+4+Rn, delayed call
			in.set("bsrf", "%s", r(n))
			in.Flow, in.HasDelay = FlowIndCall, true
		case 0x2: // braf Rn
			in.set("braf", "%s", r(n))
			in.Flow, in.HasDelay = FlowIndJump, true
		case 0x8:
			in.set("pref", "@%s", r(n))
		case 0x9:
			in.set("ocbi", "@%s", r(n))
		case 0xA:
			in.set("ocbp", "@%s", r(n))
		case 0xB:
			in.set("ocbwb", "@%s", r(n))
		case 0xC:
			in.set("movca.l", "r0, @%s", r(n))
		}
	case 0x4:
		in.set("mov.b", "%s, @(r0,%s)", r(m), r(n))
	case 0x5:
		in.set("mov.w", "%s, @(r0,%s)", r(m), r(n))
	case 0x6:
		in.set("mov.l", "%s, @(r0,%s)", r(m), r(n))
	case 0x7:
		in.set("mul.l", "%s, %s", r(m), r(n))
	case 0x8:
		if n != 0 {
			return
		}
		switch m {
		case 0x0:
			in.set("clrt", "")
		case 0x1:
			in.set("sett", "")
		case 0x2:
			in.set("clrmac", "")
		case 0x3:
			in.set("ldtlb", "")
		case 0x4:
			in.set("clrs", "")
		case 0x5:
			in.set("sets", "")
		}
	case 0x9:
		switch {
		case h == 0x0009:
			in.set("nop", "")
		case h == 0x0019:
			in.set("div0u", "")
		case m == 2: // movt Rn
			in.set("movt", "%s", r(n))
		}
	case 0xA: // sts <sys>,Rn
		switch m {
		case 0x0:
			in.set("sts", "mach, %s", r(n))
		case 0x1:
			in.set("sts", "macl, %s", r(n))
		case 0x2:
			in.set("sts", "pr, %s", r(n))
		case 0x3:
			in.set("stc", "sgr, %s", r(n))
		case 0x5:
			in.set("sts", "fpul, %s", r(n))
		case 0x6:
			in.set("sts", "fpscr, %s", r(n))
		case 0xF:
			in.set("stc", "dbr, %s", r(n))
		}
	case 0xB:
		switch h {
		case 0x000B: // rts: jump to PR after the slot
			in.set("rts", "")
			in.Flow, in.HasDelay = FlowReturn, true
		case 0x001B:
			in.set("sleep", "")
			in.Flow = FlowStop
		case 0x002B: // rte: SSR/SPC restore after the slot
			in.set("rte", "")
			in.Flow, in.HasDelay = FlowStop, true
		}
	case 0xC:
		in.set("mov.b", "@(r0,%s), %s", r(m), r(n))
	case 0xD:
		in.set("mov.w", "@(r0,%s), %s", r(m), r(n))
	case 0xE:
		in.set("mov.l", "@(r0,%s), %s", r(m), r(n))
	case 0xF:
		in.set("mac.l", "@%s+, @%s+", r(m), r(n))
	}
}

// decode2 covers group 0010: register-to-memory moves and the two-register
// logic/compare family. Low nibble 3 is a hole.
func decode2(in *Inst, h uint16) {
	n, m := rn(h), rm(h)
	switch h & 0xF {
	case 0x0:
		in.set("mov.b", "%s, @%s", r(m), r(n))
	case 0x1:
		in.set("mov.w", "%s, @%s", r(m), r(n))
	case 0x2:
		in.set("mov.l", "%s, @%s", r(m), r(n))
	case 0x4:
		in.set("mov.b", "%s, @-%s", r(m), r(n))
	case 0x5:
		in.set("mov.w", "%s, @-%s", r(m), r(n))
	case 0x6:
		in.set("mov.l", "%s, @-%s", r(m), r(n))
	case 0x7:
		in.set("div0s", "%s, %s", r(m), r(n))
	case 0x8:
		in.set("tst", "%s, %s", r(m), r(n))
	case 0x9:
		in.set("and", "%s, %s", r(m), r(n))
	case 0xA:
		in.set("xor", "%s, %s", r(m), r(n))
	case 0xB:
		in.set("or", "%s, %s", r(m), r(n))
	case 0xC:
		in.set("cmp/str", "%s, %s", r(m), r(n))
	case 0xD:
		in.set("xtrct", "%s, %s", r(m), r(n))
	case 0xE:
		in.set("mulu.w", "%s, %s", r(m), r(n))
	case 0xF:
		in.set("muls.w", "%s, %s", r(m), r(n))
	}
}

// decode3 covers group 0011: compares and the carry/overflow arithmetic. Low
// nibbles 1 and 9 are holes.
func decode3(in *Inst, h uint16) {
	n, m := rn(h), rm(h)
	switch h & 0xF {
	case 0x0:
		in.set("cmp/eq", "%s, %s", r(m), r(n))
	case 0x2:
		in.set("cmp/hs", "%s, %s", r(m), r(n))
	case 0x3:
		in.set("cmp/ge", "%s, %s", r(m), r(n))
	case 0x4:
		in.set("div1", "%s, %s", r(m), r(n))
	case 0x5:
		in.set("dmulu.l", "%s, %s", r(m), r(n))
	case 0x6:
		in.set("cmp/hi", "%s, %s", r(m), r(n))
	case 0x7:
		in.set("cmp/gt", "%s, %s", r(m), r(n))
	case 0x8:
		in.set("sub", "%s, %s", r(m), r(n))
	case 0xA:
		in.set("subc", "%s, %s", r(m), r(n))
	case 0xB:
		in.set("subv", "%s, %s", r(m), r(n))
	case 0xC:
		in.set("add", "%s, %s", r(m), r(n))
	case 0xD:
		in.set("dmuls.l", "%s, %s", r(m), r(n))
	case 0xE:
		in.set("addc", "%s, %s", r(m), r(n))
	case 0xF:
		in.set("addv", "%s, %s", r(m), r(n))
	}
}

// decode4 covers group 0100: the shift/rotate singles, the control/system
// register traffic, and the three two-register members on low nibbles C/D/F.
// The register-bank forms key on bit 7 with the bank number at bits 4-6.
func decode4(in *Inst, h uint16) {
	n, m := rn(h), rm(h)
	switch h & 0xF {
	case 0xC:
		in.set("shad", "%s, %s", r(m), r(n))
		return
	case 0xD:
		in.set("shld", "%s, %s", r(m), r(n))
		return
	case 0xF:
		in.set("mac.w", "@%s+, @%s+", r(m), r(n))
		return
	case 0x3:
		if h&0x80 != 0 { // stc.l Rm_bank,@-Rn
			in.set("stc.l", "%s, @-%s", bank(m), r(n))
			return
		}
	case 0x7:
		if h&0x80 != 0 { // ldc.l @Rm+,Rn_bank
			in.set("ldc.l", "@%s+, %s", r(n), bank(m))
			return
		}
	case 0xE:
		if h&0x80 != 0 { // ldc Rm,Rn_bank
			in.set("ldc", "%s, %s", r(n), bank(m))
			return
		}
	}
	switch h & 0xFF {
	case 0x00:
		in.set("shll", "%s", r(n))
	case 0x01:
		in.set("shlr", "%s", r(n))
	case 0x02:
		in.set("sts.l", "mach, @-%s", r(n))
	case 0x03:
		in.set("stc.l", "sr, @-%s", r(n))
	case 0x04:
		in.set("rotl", "%s", r(n))
	case 0x05:
		in.set("rotr", "%s", r(n))
	case 0x06:
		in.set("lds.l", "@%s+, mach", r(n))
	case 0x07:
		in.set("ldc.l", "@%s+, sr", r(n))
	case 0x08:
		in.set("shll2", "%s", r(n))
	case 0x09:
		in.set("shlr2", "%s", r(n))
	case 0x0A:
		in.set("lds", "%s, mach", r(n))
	case 0x0B: // jsr @Rn: delayed call through a register
		in.set("jsr", "@%s", r(n))
		in.Flow, in.HasDelay = FlowIndCall, true
	case 0x0E:
		in.set("ldc", "%s, sr", r(n))
	case 0x10:
		in.set("dt", "%s", r(n))
	case 0x11:
		in.set("cmp/pz", "%s", r(n))
	case 0x12:
		in.set("sts.l", "macl, @-%s", r(n))
	case 0x13:
		in.set("stc.l", "gbr, @-%s", r(n))
	case 0x15:
		in.set("cmp/pl", "%s", r(n))
	case 0x16:
		in.set("lds.l", "@%s+, macl", r(n))
	case 0x17:
		in.set("ldc.l", "@%s+, gbr", r(n))
	case 0x18:
		in.set("shll8", "%s", r(n))
	case 0x19:
		in.set("shlr8", "%s", r(n))
	case 0x1A:
		in.set("lds", "%s, macl", r(n))
	case 0x1B:
		in.set("tas.b", "@%s", r(n))
	case 0x1E:
		in.set("ldc", "%s, gbr", r(n))
	case 0x20:
		in.set("shal", "%s", r(n))
	case 0x21:
		in.set("shar", "%s", r(n))
	case 0x22:
		in.set("sts.l", "pr, @-%s", r(n))
	case 0x23:
		in.set("stc.l", "vbr, @-%s", r(n))
	case 0x24:
		in.set("rotcl", "%s", r(n))
	case 0x25:
		in.set("rotcr", "%s", r(n))
	case 0x26:
		in.set("lds.l", "@%s+, pr", r(n))
	case 0x27:
		in.set("ldc.l", "@%s+, vbr", r(n))
	case 0x28:
		in.set("shll16", "%s", r(n))
	case 0x29:
		in.set("shlr16", "%s", r(n))
	case 0x2A:
		in.set("lds", "%s, pr", r(n))
	case 0x2B: // jmp @Rn: delayed jump through a register
		in.set("jmp", "@%s", r(n))
		in.Flow, in.HasDelay = FlowIndJump, true
	case 0x2E:
		in.set("ldc", "%s, vbr", r(n))
	case 0x32:
		in.set("stc.l", "sgr, @-%s", r(n))
	case 0x33:
		in.set("stc.l", "ssr, @-%s", r(n))
	case 0x37:
		in.set("ldc.l", "@%s+, ssr", r(n))
	case 0x3E:
		in.set("ldc", "%s, ssr", r(n))
	case 0x43:
		in.set("stc.l", "spc, @-%s", r(n))
	case 0x47:
		in.set("ldc.l", "@%s+, spc", r(n))
	case 0x4E:
		in.set("ldc", "%s, spc", r(n))
	case 0x52:
		in.set("sts.l", "fpul, @-%s", r(n))
	case 0x56:
		in.set("lds.l", "@%s+, fpul", r(n))
	case 0x5A:
		in.set("lds", "%s, fpul", r(n))
	case 0x62:
		in.set("sts.l", "fpscr, @-%s", r(n))
	case 0x66:
		in.set("lds.l", "@%s+, fpscr", r(n))
	case 0x6A:
		in.set("lds", "%s, fpscr", r(n))
	case 0xF2:
		in.set("stc.l", "dbr, @-%s", r(n))
	case 0xF6:
		in.set("ldc.l", "@%s+, dbr", r(n))
	case 0xFA:
		in.set("ldc", "%s, dbr", r(n))
	}
}

// decode6 covers group 0110: memory-to-register moves and the one-source
// register transforms. No holes.
func decode6(in *Inst, h uint16) {
	n, m := rn(h), rm(h)
	switch h & 0xF {
	case 0x0:
		in.set("mov.b", "@%s, %s", r(m), r(n))
	case 0x1:
		in.set("mov.w", "@%s, %s", r(m), r(n))
	case 0x2:
		in.set("mov.l", "@%s, %s", r(m), r(n))
	case 0x3:
		in.set("mov", "%s, %s", r(m), r(n))
	case 0x4:
		in.set("mov.b", "@%s+, %s", r(m), r(n))
	case 0x5:
		in.set("mov.w", "@%s+, %s", r(m), r(n))
	case 0x6:
		in.set("mov.l", "@%s+, %s", r(m), r(n))
	case 0x7:
		in.set("not", "%s, %s", r(m), r(n))
	case 0x8:
		in.set("swap.b", "%s, %s", r(m), r(n))
	case 0x9:
		in.set("swap.w", "%s, %s", r(m), r(n))
	case 0xA:
		in.set("negc", "%s, %s", r(m), r(n))
	case 0xB:
		in.set("neg", "%s, %s", r(m), r(n))
	case 0xC:
		in.set("extu.b", "%s, %s", r(m), r(n))
	case 0xD:
		in.set("extu.w", "%s, %s", r(m), r(n))
	case 0xE:
		in.set("exts.b", "%s, %s", r(m), r(n))
	case 0xF:
		in.set("exts.w", "%s, %s", r(m), r(n))
	}
}

// decode8 covers group 1000, whose selector is bits 8-11 because the low byte
// is a displacement or immediate: the R0-relative short moves, cmp/eq
// immediate, and all four conditional branches.
func decode8(in *Inst, h uint16) {
	m := rm(h)
	switch (h >> 8) & 0xF {
	case 0x0: // mov.b R0,@(disp,Rn) — the register nibble here is the base
		in.set("mov.b", "r0, @(0x%X,%s)", uint32(h&0xF), r(m))
	case 0x1:
		in.set("mov.w", "r0, @(0x%X,%s)", uint32(h&0xF)*2, r(m))
	case 0x4:
		in.set("mov.b", "@(0x%X,%s), r0", uint32(h&0xF), r(m))
	case 0x5:
		in.set("mov.w", "@(0x%X,%s), r0", uint32(h&0xF)*2, r(m))
	case 0x8:
		in.set("cmp/eq", "#%d, r0", s8(h))
	case 0x9: // bt: undelayed conditional branch
		in.branch(FlowBranch, s8(h)*2, false)
		in.set("bt", "0x%08X", in.Target)
	case 0xB: // bf
		in.branch(FlowBranch, s8(h)*2, false)
		in.set("bf", "0x%08X", in.Target)
	case 0xD: // bt/s: delayed conditional branch
		in.branch(FlowBranch, s8(h)*2, true)
		in.set("bt/s", "0x%08X", in.Target)
	case 0xF: // bf/s
		in.branch(FlowBranch, s8(h)*2, true)
		in.set("bf/s", "0x%08X", in.Target)
	}
}

// decodeC covers group 1100: the GBR-relative moves, trapa, mova, and the
// R0-immediate logic family (with the @(R0,GBR) byte forms in the top half).
func decodeC(in *Inst, h uint16, addr uint32) {
	d := uint32(h & 0xFF)
	switch (h >> 8) & 0xF {
	case 0x0:
		in.set("mov.b", "r0, @(0x%X,gbr)", d)
	case 0x1:
		in.set("mov.w", "r0, @(0x%X,gbr)", d*2)
	case 0x2:
		in.set("mov.l", "r0, @(0x%X,gbr)", d*4)
	case 0x3:
		in.set("trapa", "#%d", d)
		in.Flow = FlowStop
	case 0x4:
		in.set("mov.b", "@(0x%X,gbr), r0", d)
	case 0x5:
		in.set("mov.w", "@(0x%X,gbr), r0", d*2)
	case 0x6:
		in.set("mov.l", "@(0x%X,gbr), r0", d*4)
	case 0x7: // mova: the address of a pool word, PC-relative like mov.l
		in.set("mova", "@(0x%X,pc), r0", d*4)
		in.LitAddr, in.LitSize = litL(addr, d), 4
	case 0x8:
		in.set("tst", "#%d, r0", d)
	case 0x9:
		in.set("and", "#%d, r0", d)
	case 0xA:
		in.set("xor", "#%d, r0", d)
	case 0xB:
		in.set("or", "#%d, r0", d)
	case 0xC:
		in.set("tst.b", "#%d, @(r0,gbr)", d)
	case 0xD:
		in.set("and.b", "#%d, @(r0,gbr)", d)
	case 0xE:
		in.set("xor.b", "#%d, @(r0,gbr)", d)
	case 0xF:
		in.set("or.b", "#%d, @(r0,gbr)", d)
	}
}
