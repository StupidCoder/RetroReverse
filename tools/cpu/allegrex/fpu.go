package allegrex

// fpu.go is the Allegrex COP1 (op 0x11): a single-precision-only FPU. The 32 FP
// registers are stored as float32 bit patterns in CPU.F; the one condition bit the
// branches test is CPU.FCC. There is no double precision (no .d / .l formats), so the
// only numeric formats are single (fmt 0x10) and word (fmt 0x14).

import (
	"fmt"
	"math"
)

// fpr names an FPU register for disassembly.
func fpr(i uint32) string { return fmt.Sprintf("$f%d", i&31) }

// decodeCop1 disassembles a COP1 instruction.
func decodeCop1(in Inst, w, rs, rt, rd, shamt uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	fs, fd := rd, shamt // for the fmt-dispatched forms: fs=rd, ft=rt, fd=shamt
	switch rs {
	case 0x00:
		return set("mfc1", fmt.Sprintf("mfc1 %s, %s", reg(rt), fpr(fs)))
	case 0x02:
		return set("cfc1", fmt.Sprintf("cfc1 %s, $%d", reg(rt), fs))
	case 0x04:
		return set("mtc1", fmt.Sprintf("mtc1 %s, %s", reg(rt), fpr(fs)))
	case 0x06:
		return set("ctc1", fmt.Sprintf("ctc1 %s, $%d", reg(rt), fs))
	case 0x08: // BC1
		in.Flow, in.Target, in.HasTarget, in.HasDelay = FlowBranch, in.Addr+4+uint32(int32(int16(w)))*4, true, true
		if rt&1 != 0 {
			return set("bc1t", fmt.Sprintf("bc1t $%08X", in.Target))
		}
		return set("bc1f", fmt.Sprintf("bc1f $%08X", in.Target))
	case 0x10, 0x14: // fmt = S / W
		return decodeCop1Fmt(in, w, rs, rt, fs, fd)
	}
	return word(in, w)
}

func decodeCop1Fmt(in Inst, w, rs, ft, fs, fd uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	funct := w & 0x3F
	sfx := "s"
	if rs == 0x14 {
		sfx = "w"
	}
	tri := func(name string) Inst {
		return set(name+"."+sfx, fmt.Sprintf("%s.%s %s, %s, %s", name, sfx, fpr(fd), fpr(fs), fpr(ft)))
	}
	bin := func(name string) Inst {
		return set(name+"."+sfx, fmt.Sprintf("%s.%s %s, %s", name, sfx, fpr(fd), fpr(fs)))
	}
	switch funct {
	case 0x00:
		return tri("add")
	case 0x01:
		return tri("sub")
	case 0x02:
		return tri("mul")
	case 0x03:
		return tri("div")
	case 0x04:
		return bin("sqrt")
	case 0x05:
		return bin("abs")
	case 0x06:
		return bin("mov")
	case 0x07:
		return bin("neg")
	case 0x0C:
		return bin("round.w")
	case 0x0D:
		return bin("trunc.w")
	case 0x0E:
		return bin("ceil.w")
	case 0x0F:
		return bin("floor.w")
	case 0x20:
		return set("cvt.s.w", fmt.Sprintf("cvt.s.w %s, %s", fpr(fd), fpr(fs)))
	case 0x24:
		return set("cvt.w.s", fmt.Sprintf("cvt.w.s %s, %s", fpr(fd), fpr(fs)))
	}
	if funct >= 0x30 { // c.cond.s
		return set("c.cond."+sfx, fmt.Sprintf("c.cond.%s %s, %s", sfx, fpr(fs), fpr(ft)))
	}
	return word(in, w)
}

// --- execution -------------------------------------------------------------

func (c *CPU) ff(i uint32) float32      { return math.Float32frombits(c.F[i&31]) }
func (c *CPU) setf(i uint32, v float32) { c.F[i&31] = math.Float32bits(v) }

// cop1 executes a COP1 instruction.
func (c *CPU) cop1(w, rs, rt, rd, shamt uint32) {
	fs, ft, fd := rd, rt, shamt
	switch rs {
	case 0x00: // mfc1 (delayed like a load)
		c.load(rt, c.F[fs&31])
	case 0x02: // cfc1
		var v uint32
		if fs == 31 && c.FCC {
			v = 1 << 23
		}
		c.load(rt, v)
	case 0x04: // mtc1
		c.F[fs&31] = c.reg(rt)
	case 0x06: // ctc1
		if fs == 31 {
			c.FCC = c.reg(rt)&(1<<23) != 0
		}
	case 0x08: // BC1
		simm := uint32(int32(int16(w)))
		target := c.curPC + 4 + simm<<2
		taken := (rt&1 != 0) == c.FCC // bc1t when FCC set, bc1f when clear
		c.doBranch(taken, target)
	case 0x10: // fmt = S
		c.cop1S(w, ft, fs, fd)
	case 0x14: // fmt = W
		c.cop1W(w, ft, fs, fd)
	default:
		c.Halt("unimplemented cop1 rs=0x%02X (word 0x%08X) at 0x%08X", rs, w, c.curPC)
	}
}

func (c *CPU) cop1S(w, ft, fs, fd uint32) {
	funct := w & 0x3F
	switch funct {
	case 0x00:
		c.setf(fd, c.ff(fs)+c.ff(ft))
	case 0x01:
		c.setf(fd, c.ff(fs)-c.ff(ft))
	case 0x02:
		c.setf(fd, c.ff(fs)*c.ff(ft))
	case 0x03:
		c.setf(fd, c.ff(fs)/c.ff(ft))
	case 0x04:
		c.setf(fd, float32(math.Sqrt(float64(c.ff(fs)))))
	case 0x05:
		c.setf(fd, float32(math.Abs(float64(c.ff(fs)))))
	case 0x06:
		c.F[fd&31] = c.F[fs&31]
	case 0x07:
		c.setf(fd, -c.ff(fs))
	case 0x0C: // round.w.s
		c.F[fd&31] = uint32(int32(math.RoundToEven(float64(c.ff(fs)))))
	case 0x0D: // trunc.w.s
		c.F[fd&31] = uint32(int32(math.Trunc(float64(c.ff(fs)))))
	case 0x0E: // ceil.w.s
		c.F[fd&31] = uint32(int32(math.Ceil(float64(c.ff(fs)))))
	case 0x0F: // floor.w.s
		c.F[fd&31] = uint32(int32(math.Floor(float64(c.ff(fs)))))
	case 0x24: // cvt.w.s (default round-to-nearest)
		c.F[fd&31] = uint32(int32(math.RoundToEven(float64(c.ff(fs)))))
	default:
		if funct >= 0x30 { // c.cond.s — compare, set FCC (ordered less/equal family)
			c.FCC = fcompare(funct, c.ff(fs), c.ff(ft))
			return
		}
		c.Halt("unimplemented cop1.s funct 0x%02X at 0x%08X", funct, c.curPC)
	}
}

func (c *CPU) cop1W(w, ft, fs, fd uint32) {
	funct := w & 0x3F
	switch funct {
	case 0x20: // cvt.s.w — integer word to single
		c.setf(fd, float32(int32(c.F[fs&31])))
	default:
		c.Halt("unimplemented cop1.w funct 0x%02X at 0x%08X", funct, c.curPC)
	}
}

// fcompare implements the MIPS c.cond.s predicate for the common conditions. The
// low 4 bits of funct select the condition; bit 1 = "less", bit 0 = "equal".
func fcompare(funct uint32, a, b float32) bool {
	cond := funct & 0xF
	if a != a || b != b { // unordered (NaN)
		return cond&1 != 0 && cond&0x8 != 0 // only the unordered-signalling forms
	}
	lt := a < b
	eq := a == b
	res := false
	if cond&2 != 0 {
		res = res || lt
	}
	if cond&1 != 0 {
		res = res || eq
	}
	return res
}
