package sh4

// exec_fpu.go executes group 1111 under the live FPSCR: PR selects
// single/double arithmetic on the same halfwords, SZ selects 32/64-bit fmov,
// FR selects which bank the names FR0-15 and XF0-15 reach.
//
// Precision policy, stated once: fdiv and fsqrt are correctly rounded here
// (the hardware is too). fmac computes the product-sum in float64 and rounds
// once, which matches the hardware's fused behaviour on every case a game has
// produced so far. fsca and fsrra are *approximate on the hardware* —
// documented to 2^-21 error — and exact here (math.Sincos, 1/math.Sqrt,
// rounded to float32); the divergence is a documented fiction of this core,
// visible only below the hardware's own error bound. fipr and ftrv accumulate
// in float64 where the hardware accumulates in an extended internal format.

import "math"

func (c *CPU) execFPU(h uint16, n, m uint32) {
	if c.SR&SRFD != 0 {
		c.Halt("FPU instruction %04X with SR.FD set at %08X", h, c.curPC)
		return
	}
	fr := &c.fpr[(c.FPSCR>>21)&1]
	pr := c.FPSCR&FPSCRPR != 0
	switch h & 0xF {
	case 0x0:
		c.fpuArith(fr, pr, n, m, func(a, b float64) float64 { return a + b })
	case 0x1:
		c.fpuArith(fr, pr, n, m, func(a, b float64) float64 { return a - b })
	case 0x2:
		c.fpuArith(fr, pr, n, m, func(a, b float64) float64 { return a * b })
	case 0x3:
		c.fpuArith(fr, pr, n, m, func(a, b float64) float64 { return a / b })
	case 0x4:
		if pr {
			c.setT(getDR(fr, n) == getDR(fr, m))
		} else {
			c.setT(getFR(fr, n) == getFR(fr, m))
		}
	case 0x5:
		if pr {
			c.setT(getDR(fr, n) > getDR(fr, m))
		} else {
			c.setT(getFR(fr, n) > getFR(fr, m))
		}
	case 0x6:
		c.fmovLoad(n, m, c.R[0]+c.R[m])
	case 0x7:
		c.fmovStore(n, m, c.R[0]+c.R[n])
	case 0x8:
		c.fmovLoad(n, m, c.R[m])
	case 0x9: // fmov @Rm+
		c.fmovLoad(n, m, c.R[m])
		if c.FPSCR&FPSCRSZ != 0 {
			c.R[m] += 8
		} else {
			c.R[m] += 4
		}
	case 0xA:
		c.fmovStore(n, m, c.R[n])
	case 0xB: // fmov @-Rn
		sz := uint32(4)
		if c.FPSCR&FPSCRSZ != 0 {
			sz = 8
		}
		c.R[n] -= sz
		c.fmovStore(n, m, c.R[n])
	case 0xC: // fmov register-to-register
		if c.FPSCR&FPSCRSZ != 0 {
			src := c.pair(m)
			dst := c.pair(n)
			dst[0], dst[1] = src[0], src[1]
		} else {
			fr[n] = fr[m]
		}
	case 0xD:
		c.execFPUxD(h, n, fr, pr)
	case 0xE: // fmac: single-precision only
		if pr {
			c.Halt("fmac with FPSCR.PR set at %08X", c.curPC)
			return
		}
		fr[n] = math.Float32bits(float32(float64(getFR(fr, 0))*float64(getFR(fr, m)) + float64(getFR(fr, n))))
	default:
		c.unknown(h)
	}
}

func (c *CPU) execFPUxD(h uint16, n uint32, fr *[16]uint32, pr bool) {
	switch rm(h) {
	case 0x0: // fsts fpul,FRn
		fr[n] = c.FPUL
	case 0x1: // flds FRn,fpul
		c.FPUL = fr[n]
	case 0x2: // float fpul,FRn
		if pr {
			setDR(fr, n, float64(int32(c.FPUL)))
		} else {
			fr[n] = math.Float32bits(float32(int32(c.FPUL)))
		}
	case 0x3: // ftrc FRn,fpul — truncate toward zero, saturating
		var v float64
		if pr {
			v = getDR(fr, n)
		} else {
			v = float64(getFR(fr, n))
		}
		switch {
		case math.IsNaN(v):
			c.FPUL = 0x80000000
		case v >= 2147483647:
			c.FPUL = 0x7FFFFFFF
		case v <= -2147483648:
			c.FPUL = 0x80000000
		default:
			c.FPUL = uint32(int32(v))
		}
	case 0x4: // fneg: a sign-bit flip in any mode
		fr[n] ^= 0x80000000
	case 0x5: // fabs
		fr[n] &^= 0x80000000
	case 0x6: // fsqrt
		if pr {
			setDR(fr, n, math.Sqrt(getDR(fr, n)))
		} else {
			fr[n] = math.Float32bits(float32(math.Sqrt(float64(getFR(fr, n)))))
		}
	case 0x7: // fsrra: approximate on hardware, exact here (see file doc)
		fr[n] = math.Float32bits(float32(1 / math.Sqrt(float64(getFR(fr, n)))))
	case 0x8:
		fr[n] = 0
	case 0x9:
		fr[n] = math.Float32bits(1)
	case 0xA: // fcnvsd fpul,DRn
		setDR(fr, n&^1, float64(math.Float32frombits(c.FPUL)))
	case 0xB: // fcnvds DRn,fpul
		c.FPUL = math.Float32bits(float32(getDR(fr, n&^1)))
	case 0xE: // fipr FVm,FVn: 4-element dot product into FR[n+3]
		vm, vn := (n&3)*4, (n>>2)*4
		var sum float64
		for i := uint32(0); i < 4; i++ {
			sum += float64(getFR(fr, vm+i)) * float64(getFR(fr, vn+i))
		}
		fr[vn+3] = math.Float32bits(float32(sum))
	case 0xF:
		switch {
		case h == 0xFBFD: // frchg
			c.FPSCR ^= FPSCRFR
		case h == 0xF3FD: // fschg
			c.FPSCR ^= FPSCRSZ
		case n&3 == 1: // ftrv xmtrx,FVn
			c.ftrv(fr, (n>>2)*4)
		case n&1 == 0: // fsca fpul,DRn: sin/cos of a 2^16-per-turn angle
			s, co := math.Sincos(2 * math.Pi * float64(c.FPUL&0xFFFF) / 65536)
			fr[n] = math.Float32bits(float32(s))
			fr[n+1] = math.Float32bits(float32(co))
		default:
			c.unknown(h)
		}
	default:
		c.unknown(h)
	}
}

// fpuArith is the fadd/fsub/fmul/fdiv shape: FRm op FRn in single precision,
// DRm op DRn when FPSCR.PR is set.
func (c *CPU) fpuArith(fr *[16]uint32, pr bool, n, m uint32, op func(a, b float64) float64) {
	if pr {
		setDR(fr, n, op(getDR(fr, n), getDR(fr, m)))
		return
	}
	fr[n] = math.Float32bits(float32(op(float64(getFR(fr, n)), float64(getFR(fr, m)))))
}

// pair resolves an fmov operand in SZ mode: an even register names DRn in the
// FR bank, an odd one names XDn-1 in the other bank. Returns the two words,
// high first.
func (c *CPU) pair(i uint32) []uint32 {
	b := (c.FPSCR >> 21) & 1
	if i&1 != 0 {
		b = 1 - b
	}
	base := i &^ 1
	return c.fpr[b][base : base+2]
}

// fmovLoad is every fmov memory-to-register form; fmovStore the reverse. In
// SZ mode 64 bits move as two words, the pair's high word (FRn) at the lower
// address.
func (c *CPU) fmovLoad(n, m, addr uint32) {
	if c.FPSCR&FPSCRSZ != 0 {
		dst := c.pair(n)
		dst[0] = c.read32(addr)
		dst[1] = c.read32(addr + 4)
		return
	}
	c.fpr[(c.FPSCR>>21)&1][n] = c.read32(addr)
}

func (c *CPU) fmovStore(n, m, addr uint32) {
	if c.FPSCR&FPSCRSZ != 0 {
		src := c.pair(m)
		c.write32(addr, src[0])
		c.write32(addr+4, src[1])
		return
	}
	c.write32(addr, c.fpr[(c.FPSCR>>21)&1][m])
}

// ftrv transforms FVn by XMTRX, the matrix living in the other bank:
// res[i] = Σ_j XF[i+4j]·v[j].
func (c *CPU) ftrv(fr *[16]uint32, v uint32) {
	xf := &c.fpr[1-(c.FPSCR>>21)&1]
	var in [4]float64
	for j := uint32(0); j < 4; j++ {
		in[j] = float64(getFR(fr, v+j))
	}
	for i := uint32(0); i < 4; i++ {
		var sum float64
		for j := uint32(0); j < 4; j++ {
			sum += float64(math.Float32frombits(xf[i+4*j])) * in[j]
		}
		fr[v+i] = math.Float32bits(float32(sum))
	}
}

func getFR(fr *[16]uint32, i uint32) float32 { return math.Float32frombits(fr[i&15]) }

// getDR/setDR view an even-odd pair as a double: FRn is the high word.
func getDR(fr *[16]uint32, i uint32) float64 {
	i &= 0xE
	return math.Float64frombits(uint64(fr[i])<<32 | uint64(fr[i+1]))
}

func setDR(fr *[16]uint32, i uint32, v float64) {
	i &= 0xE
	b := math.Float64bits(v)
	fr[i], fr[i+1] = uint32(b>>32), uint32(b)
}
