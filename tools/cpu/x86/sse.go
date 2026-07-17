package x86

// sse.go executes the SSE/SSE2 subset the original-Xbox titles need. A Pentium III has
// eight 128-bit XMM registers and the packed/scalar single- and double-precision float
// ops; the XDK C runtime uses them pervasively (zeroing float members with XORPS, moving
// scalars with MOVSS/MOVSD, and all the vector maths a 3-D game does). This is the one
// Xbox-only CPU piece — real-mode DOS games never reach it, so it is gated behind the
// two-byte page's default case and guarded by its own tests.
//
// Mandatory-prefix decoding. On the 0x0F page a 0xF3/0xF2/0x66 prefix does not mean
// REP/operand-size for these opcodes — it selects the SSE variant:
//
//	(none)  packed single   (PS)   — four f32 lanes
//	0x66    packed double   (PD)   — two  f64 lanes   (seen here as dOpsize==16)
//	0xF3    scalar single   (SS)   — one  f32 lane, high lanes preserved
//	0xF2    scalar double   (SD)   — one  f64 lane, high 8 bytes preserved
//
// Registers are held as raw little-endian bytes (CPU.XMM); lanes are interpreted on use.

import "math"

// sseKind classifies the operation width from the mandatory prefix.
type sseKind int

const (
	ssePS sseKind = iota // packed single (4x f32)
	ssePD                // packed double (2x f64)
	sseSS                // scalar single (1x f32)
	sseSD                // scalar double (1x f64)
)

func (c *CPU) sseKindOf(rep byte) sseKind {
	switch {
	case rep == 0xF3:
		return sseSS
	case rep == 0xF2:
		return sseSD
	case c.dOpsize == 16: // a 0x66 prefix in protected mode
		return ssePD
	default:
		return ssePS
	}
}

// --- operand access ---

// sseRM reads up to 16 bytes of an SSE r/m operand: the whole XMM register when it is a
// register operand, or n bytes from memory otherwise.
func (c *CPU) sseRM(o ea, n int) [16]byte {
	var b [16]byte
	if o.isReg {
		return c.XMM[o.reg]
	}
	for i := 0; i < n && i < 16; i++ {
		b[i] = byte(c.memRead(o.base, o.off+uint32(i), 1))
	}
	return b
}

// mmxRM reads an MMX r/m operand: the 64-bit MMX register, or 8 bytes from memory.
func (c *CPU) mmxRM(o ea) [8]byte {
	var b [8]byte
	if o.isReg {
		return c.MMX[o.reg&7]
	}
	for i := 0; i < 8; i++ {
		b[i] = byte(c.memRead(o.base, o.off+uint32(i), 1))
	}
	return b
}

// mmxStoreRM writes a 64-bit MMX value to an r/m operand (register or 8 bytes of memory).
func (c *CPU) mmxStoreRM(o ea, v [8]byte) {
	if o.isReg {
		c.MMX[o.reg&7] = v
		return
	}
	for i := 0; i < 8; i++ {
		c.memWrite(o.base, o.off+uint32(i), 1, uint32(v[i]))
	}
}

// sseStoreRM writes n bytes of v to an SSE r/m operand (all 16 to a register).
func (c *CPU) sseStoreRM(o ea, v [16]byte, n int) {
	if o.isReg {
		c.XMM[o.reg] = v
		return
	}
	for i := 0; i < n; i++ {
		c.memWrite(o.base, o.off+uint32(i), 1, uint32(v[i]))
	}
}

// lane accessors over a 16-byte value.
func f32Lane(b [16]byte, i int) float32        { return math.Float32frombits(le32b(b[i*4:])) }
func f64Lane(b [16]byte, i int) float64        { return math.Float64frombits(le64b(b[i*8:])) }
func setF32Lane(b *[16]byte, i int, v float32) { putle32(b[i*4:], math.Float32bits(v)) }
func setF64Lane(b *[16]byte, i int, v float64) { putle64(b[i*8:], math.Float64bits(v)) }

func le32b(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}
func le64b(b []byte) uint64 { return uint64(le32b(b)) | uint64(le32b(b[4:]))<<32 }
func putle32(b []byte, v uint32) {
	b[0], b[1], b[2], b[3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
}
func putle64(b []byte, v uint64) {
	putle32(b, uint32(v))
	putle32(b[4:], uint32(v>>32))
}

// opWidth is the byte count a scalar/packed variant reads from memory.
func sseWidth(k sseKind) int {
	switch k {
	case sseSS:
		return 4
	case sseSD:
		return 8
	default:
		return 16
	}
}

// execSSE executes one SSE opcode; it reports whether the opcode was handled (false
// falls through to the caller's unimplemented-opcode halt).
func (c *CPU) execSSE(op, rep byte) bool {
	k := c.sseKindOf(rep)
	switch op {

	// --- moves ---
	case 0x10: // MOVUPS/MOVSS/MOVUPD/MOVSD  xmm <- r/m
		reg, o := c.modrmE()
		src := c.sseRM(o, sseWidth(k))
		dst := c.XMM[reg]
		switch k {
		case sseSS:
			copy(dst[0:4], src[0:4]) // scalar load from mem zeroes upper? only when mem
			if !o.isReg {
				for i := 4; i < 16; i++ {
					dst[i] = 0
				}
			}
		case sseSD:
			copy(dst[0:8], src[0:8])
			if !o.isReg {
				for i := 8; i < 16; i++ {
					dst[i] = 0
				}
			}
		default:
			dst = src
		}
		c.XMM[reg] = dst
		return true
	case 0x11: // MOVUPS/MOVSS/MOVUPD/MOVSD  r/m <- xmm
		reg, o := c.modrmE()
		src := c.XMM[reg]
		c.sseStoreRM(o, src, sseWidth(k))
		return true
	case 0x28, 0x29: // MOVAPS/MOVAPD (aligned, semantics identical here)
		reg, o := c.modrmE()
		if op == 0x28 {
			c.XMM[reg] = c.sseRM(o, 16)
		} else {
			c.sseStoreRM(o, c.XMM[reg], 16)
		}
		return true
	case 0x2B: // MOVNTPS/MOVNTPD: non-temporal aligned store — no cache modelled, so a
		// plain aligned store. The XDK vertex/matrix code streams results out this way.
		reg, o := c.modrmE()
		c.sseStoreRM(o, c.XMM[reg], 16)
		return true
	case 0x12, 0x16: // MOVLPS (12) / MOVHPS (16): load 8 bytes into low/high half
		reg, o := c.modrmE()
		src := c.sseRM(o, 8)
		dst := c.XMM[reg]
		half := 0
		if op == 0x16 {
			half = 8
		}
		copy(dst[half:half+8], src[0:8])
		c.XMM[reg] = dst
		return true
	case 0x13, 0x17: // MOVLPS / MOVHPS store: store low/high 8 bytes to memory
		reg, o := c.modrmE()
		src := c.XMM[reg]
		half := 0
		if op == 0x17 {
			half = 8
		}
		var v [16]byte
		copy(v[0:8], src[half:half+8])
		c.sseStoreRM(o, v, 8)
		return true
	case 0x6E: // MOVD xmm <- r/m32 (66 prefix) ; no prefix: MOVD mm <- r/m32
		reg, o := c.modrmE()
		v := c.rEA(o, 4)
		if k == ssePD {
			var b [16]byte
			putle32(b[0:], v)
			c.XMM[reg] = b
		} else {
			var b [8]byte
			putle32(b[0:], v)
			c.MMX[reg&7] = b
		}
		return true
	case 0x7E: // 66: MOVD r/m32 <- xmm ; F3: MOVQ xmm <- xmm/m64
		reg, o := c.modrmE()
		if rep == 0xF3 { // MOVQ xmm <- xmm/m64
			src := c.sseRM(o, 8)
			var b [16]byte
			copy(b[0:8], src[0:8])
			c.XMM[reg] = b
			return true
		}
		if k == ssePD { // MOVD r/m32 <- xmm
			c.wEA(o, 4, le32b(c.XMM[reg][0:]))
			return true
		}
		// no prefix: MOVD r/m32 <- mm (the XMV decoder's scalar extract)
		c.wEA(o, 4, le32b(c.MMX[reg&7][0:]))
		return true
	case 0x77: // EMMS: leave MMX state — a no-op here (separate MMX/x87 files)
		return true
	case 0x6F: // MOVDQA (66) / MOVDQU (F3) xmm <- m128 ; no-prefix: MMX MOVQ mm <- mm/m64
		reg, o := c.modrmE()
		if k == ssePD || rep == 0xF3 {
			c.XMM[reg] = c.sseRM(o, 16)
		} else {
			c.MMX[reg&7] = c.mmxRM(o)
		}
		return true
	case 0x7F: // MOVDQA (66) / MOVDQU (F3) m128 <- xmm ; no-prefix: MMX MOVQ mm/m64 <- mm
		reg, o := c.modrmE()
		if k == ssePD || rep == 0xF3 {
			c.sseStoreRM(o, c.XMM[reg], 16)
		} else {
			c.mmxStoreRM(o, c.MMX[reg&7])
		}
		return true
	case 0xD6: // MOVQ r/m64 <- xmm (66 prefix)
		if k != ssePD {
			return false
		}
		reg, o := c.modrmE()
		var v [16]byte
		copy(v[0:8], c.XMM[reg][0:8])
		c.sseStoreRM(o, v, 8)
		return true
	case 0xC6: // SHUFPS / SHUFPD: select lanes from dest and src per imm8
		reg, o := c.modrmE()
		a, b := c.XMM[reg], c.sseRM(o, 16)
		imm := byte(c.fetch8())
		var r [16]byte
		if k == ssePD { // SHUFPD: 2 doubleword-pair lanes (8 bytes each)
			copy(r[0:8], a[(imm&1)*8:(imm&1)*8+8])
			copy(r[8:16], b[((imm>>1)&1)*8:((imm>>1)&1)*8+8])
		} else { // SHUFPS: 4 single lanes (4 bytes each); low two from dest, high two from src
			l0 := (imm >> 0) & 3
			l1 := (imm >> 2) & 3
			l2 := (imm >> 4) & 3
			l3 := (imm >> 6) & 3
			copy(r[0:4], a[l0*4:l0*4+4])
			copy(r[4:8], a[l1*4:l1*4+4])
			copy(r[8:12], b[l2*4:l2*4+4])
			copy(r[12:16], b[l3*4:l3*4+4])
		}
		c.XMM[reg] = r
		return true
	case 0x14, 0x15: // UNPCKLPS/UNPCKLPD (14), UNPCKHPS/UNPCKHPD (15): interleave the
		// operands' low (14) or high (15) lanes. The high form is the WMA decoder's
		// (site 0x1EAD4D, the WMADEC section) — the menu's music, which the title only
		// reaches now that it is past the save-game dialogue.
		reg, o := c.modrmE()
		a, b := c.XMM[reg], c.sseRM(o, 16)
		h := 0 // byte offset of the half being interleaved: low (0) or high (8)
		if op == 0x15 {
			h = 8
		}
		var r [16]byte
		if k == ssePD {
			copy(r[0:8], a[h:h+8])
			copy(r[8:16], b[h:h+8])
		} else {
			copy(r[0:4], a[h:h+4])
			copy(r[4:8], b[h:h+4])
			copy(r[8:12], a[h+4:h+8])
			copy(r[12:16], b[h+4:h+8])
		}
		c.XMM[reg] = r
		return true

	// --- bitwise logic (operate on all 128 bits) ---
	case 0x54, 0x55, 0x56, 0x57: // ANDPS, ANDNPS, ORPS, XORPS (and PD forms)
		reg, o := c.modrmE()
		a, b := c.XMM[reg], c.sseRM(o, 16)
		var r [16]byte
		for i := 0; i < 16; i++ {
			switch op {
			case 0x54:
				r[i] = a[i] & b[i]
			case 0x55:
				r[i] = (^a[i]) & b[i]
			case 0x56:
				r[i] = a[i] | b[i]
			case 0x57:
				r[i] = a[i] ^ b[i]
			}
		}
		c.XMM[reg] = r
		return true

	// --- arithmetic ---
	case 0x58, 0x59, 0x5C, 0x5D, 0x5E, 0x5F: // ADD MUL SUB MIN DIV MAX
		reg, o := c.modrmE()
		c.sseArith(op, k, reg, o)
		return true
	case 0x51: // SQRT
		reg, o := c.modrmE()
		c.sseUnary(k, reg, o, func(x float64) float64 { return math.Sqrt(x) })
		return true
	case 0x53: // RCP (approximate reciprocal)
		reg, o := c.modrmE()
		c.sseUnary(k, reg, o, func(x float64) float64 { return 1.0 / x })
		return true
	case 0x52: // RSQRT (approximate reciprocal sqrt)
		reg, o := c.modrmE()
		c.sseUnary(k, reg, o, func(x float64) float64 { return 1.0 / math.Sqrt(x) })
		return true

	// --- conversions ---
	case 0x2A: // F3/F2: CVTSI2SS/SD xmm <- r/m32 ; no prefix / 66: CVTPI2PS/PD xmm <- mm/m64
		reg, o := c.modrmE()
		dst := c.XMM[reg]
		switch k {
		case sseSS, sseSD:
			iv := int32(c.rEA(o, 4))
			if k == sseSD {
				setF64Lane(&dst, 0, float64(iv))
			} else {
				setF32Lane(&dst, 0, float32(iv))
			}
		default: // packed: source is an MMX register or m64 holding two int32s.
			// The scalar handler this replaced read a GPR here — for the packed
			// forms that is the WRONG FILE, and the XMV movie decoder is full of
			// them (its int<->float stages run through mm registers).
			src := c.mmxRM(o)
			a, b := int32(le32b(src[0:])), int32(le32b(src[4:]))
			if k == ssePD {
				setF64Lane(&dst, 0, float64(a))
				setF64Lane(&dst, 1, float64(b))
			} else {
				setF32Lane(&dst, 0, float32(a))
				setF32Lane(&dst, 1, float32(b))
			}
		}
		c.XMM[reg] = dst
		return true
	case 0x2C, 0x2D: // F3/F2: CVT(T)SS/SD2SI r32 <- xmm ; no prefix / 66: CVT(T)PS/PD2PI mm <- xmm/m64
		reg, o := c.modrmE()
		trunc := op == 0x2C
		cvt := func(f float64) uint32 {
			if trunc {
				return uint32(int32(f)) // Go float->int conversion truncates
			}
			return uint32(int32(math.RoundToEven(f)))
		}
		switch k {
		case sseSS, sseSD:
			src := c.sseRM(o, sseWidth(k))
			var f float64
			if k == sseSD {
				f = f64Lane(src, 0)
			} else {
				f = float64(f32Lane(src, 0))
			}
			c.setReg(reg, 4, cvt(f))
		default: // packed: destination is an MMX register (two int32 lanes) — the
			// scalar handler wrote a GPR here, silently clobbering a live pointer.
			n := 8 // CVT(T)PS2PI reads xmm/m64
			if k == ssePD {
				n = 16 // CVT(T)PD2PI reads xmm/m128
			}
			src := c.sseRM(o, n)
			var m [8]byte
			if k == ssePD {
				putle32(m[0:], cvt(f64Lane(src, 0)))
				putle32(m[4:], cvt(f64Lane(src, 1)))
			} else {
				putle32(m[0:], cvt(float64(f32Lane(src, 0))))
				putle32(m[4:], cvt(float64(f32Lane(src, 1))))
			}
			c.MMX[reg&7] = m
		}
		return true
	case 0x5A: // CVTSS2SD / CVTSD2SS / CVTPS2PD / CVTPD2PS
		reg, o := c.modrmE()
		src := c.sseRM(o, sseWidth(k))
		dst := c.XMM[reg]
		switch k {
		case sseSS: // SS->SD
			setF64Lane(&dst, 0, float64(f32Lane(src, 0)))
		case sseSD: // SD->SS
			setF32Lane(&dst, 0, float32(f64Lane(src, 0)))
		case ssePS: // PS->PD (two lanes)
			setF64Lane(&dst, 0, float64(f32Lane(src, 0)))
			setF64Lane(&dst, 1, float64(f32Lane(src, 1)))
		default: // PD->PS
			setF32Lane(&dst, 0, float32(f64Lane(src, 0)))
			setF32Lane(&dst, 1, float32(f64Lane(src, 1)))
		}
		c.XMM[reg] = dst
		return true
	case 0x5B: // CVTDQ2PS / CVTPS2DQ / CVTTPS2DQ
		reg, o := c.modrmE()
		src := c.sseRM(o, 16)
		dst := c.XMM[reg]
		if rep == 0xF3 || c.dOpsize == 16 { // (T)PS2DQ: float -> int32 lanes
			for i := 0; i < 4; i++ {
				putle32(dst[i*4:], uint32(int32(f32Lane(src, i))))
			}
		} else { // DQ2PS: int32 lanes -> float
			for i := 0; i < 4; i++ {
				setF32Lane(&dst, i, float32(int32(le32b(src[i*4:]))))
			}
		}
		c.XMM[reg] = dst
		return true

	// --- compares (set EFLAGS) ---
	case 0x2E, 0x2F: // UCOMISS/UCOMISD (2E) and COMISS/COMISD (2F)
		reg, o := c.modrmE()
		src := c.sseRM(o, sseWidth(k))
		var a, b float64
		if k == sseSD || (k == ssePD) {
			a, b = f64Lane(c.XMM[reg], 0), f64Lane(src, 0)
		} else {
			a, b = float64(f32Lane(c.XMM[reg], 0)), float64(f32Lane(src, 0))
		}
		c.sseCompareFlags(a, b)
		return true

	// --- integer SSE2 used for zeroing/moving ---
	case 0xEF: // PXOR (66) — commonly `pxor xmm,xmm` to zero
		if k != ssePD {
			return false
		}
		reg, o := c.modrmE()
		a, b := c.XMM[reg], c.sseRM(o, 16)
		var r [16]byte
		for i := 0; i < 16; i++ {
			r[i] = a[i] ^ b[i]
		}
		c.XMM[reg] = r
		return true
	}
	return false
}

// sseArith applies a binary float op across the lanes selected by k.
func (c *CPU) sseArith(op byte, k sseKind, reg byte, o ea) {
	src := c.sseRM(o, sseWidth(k))
	dst := c.XMM[reg]
	f32 := func(a, b float32) float32 {
		switch op {
		case 0x58:
			return a + b
		case 0x59:
			return a * b
		case 0x5C:
			return a - b
		case 0x5D:
			return float32(math.Min(float64(a), float64(b)))
		case 0x5E:
			return a / b
		default: // 0x5F max
			return float32(math.Max(float64(a), float64(b)))
		}
	}
	f64 := func(a, b float64) float64 {
		switch op {
		case 0x58:
			return a + b
		case 0x59:
			return a * b
		case 0x5C:
			return a - b
		case 0x5D:
			return math.Min(a, b)
		case 0x5E:
			return a / b
		default:
			return math.Max(a, b)
		}
	}
	switch k {
	case sseSS:
		setF32Lane(&dst, 0, f32(f32Lane(dst, 0), f32Lane(src, 0)))
	case sseSD:
		setF64Lane(&dst, 0, f64(f64Lane(dst, 0), f64Lane(src, 0)))
	case ssePS:
		for i := 0; i < 4; i++ {
			setF32Lane(&dst, i, f32(f32Lane(dst, i), f32Lane(src, i)))
		}
	default: // PD
		for i := 0; i < 2; i++ {
			setF64Lane(&dst, i, f64(f64Lane(dst, i), f64Lane(src, i)))
		}
	}
	c.XMM[reg] = dst
}

// sseUnary applies a unary float op across the lanes selected by k.
func (c *CPU) sseUnary(k sseKind, reg byte, o ea, fn func(float64) float64) {
	src := c.sseRM(o, sseWidth(k))
	dst := c.XMM[reg]
	switch k {
	case sseSS:
		setF32Lane(&dst, 0, float32(fn(float64(f32Lane(src, 0)))))
	case sseSD:
		setF64Lane(&dst, 0, fn(f64Lane(src, 0)))
	case ssePS:
		for i := 0; i < 4; i++ {
			setF32Lane(&dst, i, float32(fn(float64(f32Lane(src, i)))))
		}
	default:
		for i := 0; i < 2; i++ {
			setF64Lane(&dst, i, fn(f64Lane(src, i)))
		}
	}
	c.XMM[reg] = dst
}

// sseCompareFlags sets ZF/PF/CF from an ordered scalar compare (COMISS/UCOMISS), the
// way the hardware maps float ordering onto the integer flags. Unordered (a NaN) sets
// ZF=PF=CF=1; otherwise PF=0 and ZF/CF encode ==,<,>.
func (c *CPU) sseCompareFlags(a, b float64) {
	c.OF, c.SF, c.AF = false, false, false
	switch {
	case math.IsNaN(a) || math.IsNaN(b):
		c.ZF, c.PF, c.CF = true, true, true
	case a > b:
		c.ZF, c.PF, c.CF = false, false, false
	case a < b:
		c.ZF, c.PF, c.CF = false, false, true
	default: // equal
		c.ZF, c.PF, c.CF = true, false, false
	}
}
