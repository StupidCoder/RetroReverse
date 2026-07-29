package sh4

// decode_fpu.go names group 1111 — the whole FPU space — following the same
// split the other packages give an extra instruction space of its own
// (tools/cpu/r5900's mmi_decode.go).
//
// The listing always shows the single-precision reading: `fadd fr2, fr4` is
// what the halfword says, and whether it executes as fadd dr2, dr4 depends on
// FPSCR.PR at run time, which a static decoder cannot know (see the package
// doc). The only registers rendered as doubles are the ones that are doubles
// in every mode: the fcnvsd/fcnvds conversions and fsca's sine/cosine pair.

import "fmt"

func fr(n uint32) string { return fmt.Sprintf("fr%d", n) }
func dr(n uint32) string { return fmt.Sprintf("dr%d", n) }
func fv(n uint32) string { return fmt.Sprintf("fv%d", n*4) }

func decodeFPU(in *Inst, h uint16) {
	n, m := rn(h), rm(h)
	switch h & 0xF {
	case 0x0:
		in.set("fadd", "%s, %s", fr(m), fr(n))
	case 0x1:
		in.set("fsub", "%s, %s", fr(m), fr(n))
	case 0x2:
		in.set("fmul", "%s, %s", fr(m), fr(n))
	case 0x3:
		in.set("fdiv", "%s, %s", fr(m), fr(n))
	case 0x4:
		in.set("fcmp/eq", "%s, %s", fr(m), fr(n))
	case 0x5:
		in.set("fcmp/gt", "%s, %s", fr(m), fr(n))
	case 0x6:
		in.set("fmov.s", "@(r0,%s), %s", r(m), fr(n))
	case 0x7:
		in.set("fmov.s", "%s, @(r0,%s)", fr(m), r(n))
	case 0x8:
		in.set("fmov.s", "@%s, %s", r(m), fr(n))
	case 0x9:
		in.set("fmov.s", "@%s+, %s", r(m), fr(n))
	case 0xA:
		in.set("fmov.s", "%s, @%s", fr(m), r(n))
	case 0xB:
		in.set("fmov.s", "%s, @-%s", fr(m), r(n))
	case 0xC:
		in.set("fmov", "%s, %s", fr(m), fr(n))
	case 0xD:
		decodeFPUxD(in, h, n)
	case 0xE:
		in.set("fmac", "fr0, %s, %s", fr(m), fr(n))
	}
}

// decodeFPUxD covers the 1111nnnnxxxx1101 column: the one-register transforms
// and, behind selector 0xF, the fsca/ftrv/frchg/fschg corner where the
// register field itself carries more selector bits.
func decodeFPUxD(in *Inst, h uint16, n uint32) {
	switch rm(h) {
	case 0x0:
		in.set("fsts", "fpul, %s", fr(n))
	case 0x1:
		in.set("flds", "%s, fpul", fr(n))
	case 0x2:
		in.set("float", "fpul, %s", fr(n))
	case 0x3:
		in.set("ftrc", "%s, fpul", fr(n))
	case 0x4:
		in.set("fneg", "%s", fr(n))
	case 0x5:
		in.set("fabs", "%s", fr(n))
	case 0x6:
		in.set("fsqrt", "%s", fr(n))
	case 0x7:
		in.set("fsrra", "%s", fr(n))
	case 0x8:
		in.set("fldi0", "%s", fr(n))
	case 0x9:
		in.set("fldi1", "%s", fr(n))
	case 0xA: // double in every mode
		in.set("fcnvsd", "fpul, %s", dr(n))
	case 0xB:
		in.set("fcnvds", "%s, fpul", dr(n))
	case 0xE: // fipr FVm,FVn: 1111nnmm1110 1101
		in.set("fipr", "%s, %s", fv(n&3), fv(n>>2))
	case 0xF:
		switch {
		case h == 0xFBFD:
			in.set("frchg", "")
		case h == 0xF3FD:
			in.set("fschg", "")
		case n&3 == 1: // ftrv XMTRX,FVn: 1111nn01 1111 1101
			in.set("ftrv", "xmtrx, %s", fv(n>>2))
		case n&1 == 0: // fsca FPUL,DRn: 1111nnn0 1111 1101
			in.set("fsca", "fpul, %s", dr(n))
		}
	}
}
