package x86

// mmxint.go executes the MMX / SSE-integer opcode group — the packed-integer SIMD a
// video codec is made of. OutRun 2006's XMV movie decoder (the title-side library at
// 0x180xxx) is the first code on any host to reach these: it zeroes frame planes with
// PXOR mm0,mm0 + MOVQ runs, and its IDCT/motion-compensation loops use the classic
// MMX arithmetic set. Every op here exists in two widths sharing one implementation:
// no prefix = the 64-bit mm registers (8 lanes of bytes), 0x66 = the 128-bit xmm
// registers (SSE2 integer). Values are raw little-endian bytes; lanes are interpreted
// on use, exactly like sse.go.

// intOperands decodes a packed-integer op's operands: destination register index,
// source value, width in bytes, and whether the 0x66 (xmm) form is active.
func (c *CPU) intOperands(k sseKind) (reg byte, src [16]byte, o ea, wide bool) {
	reg, o = c.modrmE()
	wide = k == ssePD // the 0x66 prefix in this group selects the xmm form
	if wide {
		src = c.sseRM(o, 16)
	} else {
		s8 := c.mmxRM(o)
		copy(src[:8], s8[:])
	}
	return
}

func (c *CPU) intReg(reg byte, wide bool) [16]byte {
	if wide {
		return c.XMM[reg]
	}
	var v [16]byte
	copy(v[:8], c.MMX[reg&7][:])
	return v
}

func (c *CPU) setIntReg(reg byte, wide bool, v [16]byte) {
	if wide {
		c.XMM[reg] = v
		return
	}
	var m [8]byte
	copy(m[:], v[:8])
	c.MMX[reg&7] = m
}

// intWidth is the operand width in bytes for the active form.
func intWidth(wide bool) int {
	if wide {
		return 16
	}
	return 8
}

// --- saturation helpers ---

func satI8(v int32) byte {
	if v < -128 {
		v = -128
	}
	if v > 127 {
		v = 127
	}
	return byte(v)
}

func satU8(v int32) byte {
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return byte(v)
}

func satI16(v int32) uint16 {
	if v < -32768 {
		v = -32768
	}
	if v > 32767 {
		v = 32767
	}
	return uint16(v)
}

func satU16(v int32) uint16 {
	if v < 0 {
		v = 0
	}
	if v > 65535 {
		v = 65535
	}
	return uint16(v)
}

// lane accessors (little-endian words/dwords/qwords inside the byte vectors).
func w16(b []byte) uint16       { return uint16(b[0]) | uint16(b[1])<<8 }
func putw16(b []byte, v uint16) { b[0] = byte(v); b[1] = byte(v >> 8) }

// execMMXInt executes one packed-integer opcode (both mm and xmm widths); it reports
// whether the opcode was handled. Called after execSSE declines an 0F opcode.
func (c *CPU) execMMXInt(op, rep byte) bool {
	k := c.sseKindOf(rep)

	switch op {

	// --- unpack / pack (0x60-0x6B) ---
	case 0x60, 0x61, 0x62, 0x68, 0x69, 0x6A, 0x6C, 0x6D:
		if (op == 0x6C || op == 0x6D) && k != ssePD {
			return false // PUNPCKLQDQ/HQDQ are xmm-only
		}
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var elem int // element size in bytes
		switch op & 0x0F {
		case 0x0, 0x8:
			elem = 1
		case 0x1, 0x9:
			elem = 2
		case 0x2, 0xA:
			elem = 4
		default:
			elem = 8
		}
		high := op == 0x68 || op == 0x69 || op == 0x6A || op == 0x6D
		half := n / 2
		off := 0
		if high {
			off = half
		}
		var r [16]byte
		pairs := half / elem
		for i := 0; i < pairs; i++ {
			copy(r[i*2*elem:], dst[off+i*elem:off+i*elem+elem])
			copy(r[i*2*elem+elem:], src[off+i*elem:off+i*elem+elem])
		}
		c.setIntReg(reg, wide, r)
		return true

	case 0x63: // PACKSSWB: words -> signed-saturated bytes (dst low, src high)
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n/2; i++ {
			r[i] = satI8(int32(int16(w16(dst[i*2:]))))
		}
		for i := 0; i < n/2; i++ {
			r[n/2+i] = satI8(int32(int16(w16(src[i*2:]))))
		}
		c.setIntReg(reg, wide, r)
		return true

	case 0x67: // PACKUSWB: words -> unsigned-saturated bytes
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n/2; i++ {
			r[i] = satU8(int32(int16(w16(dst[i*2:]))))
		}
		for i := 0; i < n/2; i++ {
			r[n/2+i] = satU8(int32(int16(w16(src[i*2:]))))
		}
		c.setIntReg(reg, wide, r)
		return true

	case 0x6B: // PACKSSDW: dwords -> signed-saturated words
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n/4; i++ {
			v := int32(le32b(dst[i*4:]))
			putw16(r[i*2:], satI16Clamp(v))
		}
		for i := 0; i < n/4; i++ {
			v := int32(le32b(src[i*4:]))
			putw16(r[n/2+i*2:], satI16Clamp(v))
		}
		c.setIntReg(reg, wide, r)
		return true

	// --- compares (0x64-0x66 PCMPGT, 0x74-0x76 PCMPEQ) ---
	case 0x64, 0x65, 0x66, 0x74, 0x75, 0x76:
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		eq := op >= 0x74
		var r [16]byte
		switch op & 3 {
		case 0: // bytes
			for i := 0; i < n; i++ {
				hit := (eq && dst[i] == src[i]) || (!eq && int8(dst[i]) > int8(src[i]))
				if hit {
					r[i] = 0xFF
				}
			}
		case 1: // words
			for i := 0; i < n/2; i++ {
				a, b := int16(w16(dst[i*2:])), int16(w16(src[i*2:]))
				if (eq && a == b) || (!eq && a > b) {
					putw16(r[i*2:], 0xFFFF)
				}
			}
		case 2: // dwords
			for i := 0; i < n/4; i++ {
				a, b := int32(le32b(dst[i*4:])), int32(le32b(src[i*4:]))
				if (eq && a == b) || (!eq && a > b) {
					putle32(r[i*4:], 0xFFFFFFFF)
				}
			}
		}
		c.setIntReg(reg, wide, r)
		return true

	// --- shuffles (0x70) ---
	case 0x70:
		reg, o := c.modrmE()
		imm := byte(c.fetch8())
		switch k {
		case ssePS: // PSHUFW mm, mm/m64, imm8
			src := c.mmxRM(o)
			var r [8]byte
			for i := 0; i < 4; i++ {
				s := (imm >> (i * 2)) & 3
				copy(r[i*2:], src[s*2:s*2+2])
			}
			c.MMX[reg&7] = r
		case ssePD: // PSHUFD xmm, xmm/m128, imm8
			src := c.sseRM(o, 16)
			var r [16]byte
			for i := 0; i < 4; i++ {
				s := (imm >> (i * 2)) & 3
				copy(r[i*4:], src[s*4:s*4+4])
			}
			c.XMM[reg] = r
		case sseSS: // PSHUFHW (F3)
			src := c.sseRM(o, 16)
			r := src
			for i := 0; i < 4; i++ {
				s := (imm >> (i * 2)) & 3
				copy(r[8+i*2:], src[8+s*2:8+s*2+2])
			}
			c.XMM[reg] = r
		case sseSD: // PSHUFLW (F2)
			src := c.sseRM(o, 16)
			r := src
			for i := 0; i < 4; i++ {
				s := (imm >> (i * 2)) & 3
				copy(r[i*2:], src[s*2:s*2+2])
			}
			c.XMM[reg] = r
		}
		return true

	// --- shift-by-immediate groups (0x71-0x73) ---
	case 0x71, 0x72, 0x73:
		sub, o := c.modrmE() // reg field = sub-op, rm = the target register
		imm := uint32(c.fetch8())
		if !o.isReg {
			return false // these forms only take a register operand
		}
		wide := k == ssePD
		tgt := o.reg
		v := c.intReg(tgt, wide)
		n := intWidth(wide)
		var elem int
		switch op {
		case 0x71:
			elem = 2
		case 0x72:
			elem = 4
		default:
			elem = 8
		}
		switch sub {
		case 2: // PSRLW/D/Q
			shiftLanes(&v, n, elem, imm, shiftRightLogical)
		case 3: // PSRLDQ (0x73, xmm only): byte shift right
			if op != 0x73 || !wide {
				return false
			}
			byteShift(&v, int(imm), false)
		case 4: // PSRAW/D
			if op == 0x73 {
				return false
			}
			shiftLanes(&v, n, elem, imm, shiftRightArith)
		case 6: // PSLLW/D/Q
			shiftLanes(&v, n, elem, imm, shiftLeft)
		case 7: // PSLLDQ (0x73, xmm only): byte shift left
			if op != 0x73 || !wide {
				return false
			}
			byteShift(&v, int(imm), true)
		default:
			return false
		}
		c.setIntReg(tgt, wide, v)
		return true

	// --- shift-by-register (0xD1-D3, 0xE1-E2, 0xF1-F3) ---
	case 0xD1, 0xD2, 0xD3, 0xE1, 0xE2, 0xF1, 0xF2, 0xF3:
		reg, src, _, wide := c.intOperands(k)
		v := c.intReg(reg, wide)
		n := intWidth(wide)
		cnt := le32b(src[0:4])
		if le32b(src[4:8]) != 0 {
			cnt = 0xFFFFFFFF // count taken from the low 64 bits; huge counts saturate
		}
		var elem int
		switch op {
		case 0xD1, 0xE1, 0xF1:
			elem = 2
		case 0xD2, 0xE2, 0xF2:
			elem = 4
		default:
			elem = 8
		}
		switch {
		case op >= 0xF1: // PSLL
			shiftLanes(&v, n, elem, cnt, shiftLeft)
		case op >= 0xE1: // PSRA
			shiftLanes(&v, n, elem, cnt, shiftRightArith)
		default: // PSRL
			shiftLanes(&v, n, elem, cnt, shiftRightLogical)
		}
		c.setIntReg(reg, wide, v)
		return true

	// --- logic ---
	case 0xDB, 0xDF, 0xEB, 0xEF: // PAND / PANDN / POR / PXOR
		if op == 0xEF && k == ssePD {
			return false // the xmm PXOR already lives in sse.go
		}
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n; i++ {
			switch op {
			case 0xDB:
				r[i] = dst[i] & src[i]
			case 0xDF:
				r[i] = ^dst[i] & src[i]
			case 0xEB:
				r[i] = dst[i] | src[i]
			case 0xEF:
				r[i] = dst[i] ^ src[i]
			}
		}
		c.setIntReg(reg, wide, r)
		return true

	// --- byte/word/dword/qword add & subtract, wrapping ---
	case 0xFC, 0xFD, 0xFE, 0xD4, 0xF8, 0xF9, 0xFA, 0xFB:
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		sub := op >= 0xF8 && op <= 0xFB
		var elem int
		switch op {
		case 0xFC, 0xF8:
			elem = 1
		case 0xFD, 0xF9:
			elem = 2
		case 0xFE, 0xFA:
			elem = 4
		default:
			elem = 8 // PADDQ / PSUBQ
		}
		var r [16]byte
		for i := 0; i < n; i += elem {
			var a, b uint64
			for j := 0; j < elem; j++ {
				a |= uint64(dst[i+j]) << (8 * j)
				b |= uint64(src[i+j]) << (8 * j)
			}
			var v uint64
			if sub {
				v = a - b
			} else {
				v = a + b
			}
			for j := 0; j < elem; j++ {
				r[i+j] = byte(v >> (8 * j))
			}
		}
		c.setIntReg(reg, wide, r)
		return true

	// --- saturating add/subtract ---
	case 0xEC, 0xED, 0xDC, 0xDD, 0xE8, 0xE9, 0xD8, 0xD9:
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		signed := op == 0xEC || op == 0xED || op == 0xE8 || op == 0xE9
		sub := op == 0xE8 || op == 0xE9 || op == 0xD8 || op == 0xD9
		words := op == 0xED || op == 0xDD || op == 0xE9 || op == 0xD9
		if words {
			for i := 0; i < n/2; i++ {
				var a, b int32
				if signed {
					a, b = int32(int16(w16(dst[i*2:]))), int32(int16(w16(src[i*2:])))
				} else {
					a, b = int32(w16(dst[i*2:])), int32(w16(src[i*2:]))
				}
				v := a + b
				if sub {
					v = a - b
				}
				if signed {
					putw16(r[i*2:], satI16(v))
				} else {
					putw16(r[i*2:], satU16(v))
				}
			}
		} else {
			for i := 0; i < n; i++ {
				var a, b int32
				if signed {
					a, b = int32(int8(dst[i])), int32(int8(src[i]))
				} else {
					a, b = int32(dst[i]), int32(src[i])
				}
				v := a + b
				if sub {
					v = a - b
				}
				if signed {
					r[i] = satI8(v)
				} else {
					r[i] = satU8(v)
				}
			}
		}
		c.setIntReg(reg, wide, r)
		return true

	// --- multiplies ---
	case 0xD5, 0xE5, 0xE4: // PMULLW / PMULHW / PMULHUW
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n/2; i++ {
			switch op {
			case 0xD5:
				v := int32(int16(w16(dst[i*2:]))) * int32(int16(w16(src[i*2:])))
				putw16(r[i*2:], uint16(v))
			case 0xE5:
				v := int32(int16(w16(dst[i*2:]))) * int32(int16(w16(src[i*2:])))
				putw16(r[i*2:], uint16(uint32(v)>>16))
			case 0xE4:
				v := uint32(w16(dst[i*2:])) * uint32(w16(src[i*2:]))
				putw16(r[i*2:], uint16(v>>16))
			}
		}
		c.setIntReg(reg, wide, r)
		return true

	case 0xF5: // PMADDWD: signed word pairs multiplied and summed into dwords
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n/4; i++ {
			a0 := int32(int16(w16(dst[i*4:])))
			b0 := int32(int16(w16(src[i*4:])))
			a1 := int32(int16(w16(dst[i*4+2:])))
			b1 := int32(int16(w16(src[i*4+2:])))
			putle32(r[i*4:], uint32(a0*b0+a1*b1))
		}
		c.setIntReg(reg, wide, r)
		return true

	case 0xF4: // PMULUDQ: even dword lanes multiplied into qwords
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n/8; i++ {
			v := uint64(le32b(dst[i*8:])) * uint64(le32b(src[i*8:]))
			for j := 0; j < 8; j++ {
				r[i*8+j] = byte(v >> (8 * j))
			}
		}
		c.setIntReg(reg, wide, r)
		return true

	case 0xF6: // PSADBW: sum of absolute byte differences per 8-byte group
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for g := 0; g < n/8; g++ {
			var sum uint32
			for i := 0; i < 8; i++ {
				d := int32(dst[g*8+i]) - int32(src[g*8+i])
				if d < 0 {
					d = -d
				}
				sum += uint32(d)
			}
			putw16(r[g*8:], uint16(sum))
		}
		c.setIntReg(reg, wide, r)
		return true

	// --- min/max/average ---
	case 0xDA, 0xDE: // PMINUB / PMAXUB
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n; i++ {
			if (op == 0xDA) == (src[i] < dst[i]) {
				r[i] = src[i]
			} else {
				r[i] = dst[i]
			}
		}
		c.setIntReg(reg, wide, r)
		return true

	case 0xEA, 0xEE: // PMINSW / PMAXSW
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		for i := 0; i < n/2; i++ {
			a, b := int16(w16(dst[i*2:])), int16(w16(src[i*2:]))
			v := a
			if (op == 0xEA) == (b < a) {
				v = b
			}
			putw16(r[i*2:], uint16(v))
		}
		c.setIntReg(reg, wide, r)
		return true

	case 0xE0, 0xE3: // PAVGB / PAVGW: unsigned rounded average
		reg, src, _, wide := c.intOperands(k)
		dst := c.intReg(reg, wide)
		n := intWidth(wide)
		var r [16]byte
		if op == 0xE0 {
			for i := 0; i < n; i++ {
				r[i] = byte((uint32(dst[i]) + uint32(src[i]) + 1) >> 1)
			}
		} else {
			for i := 0; i < n/2; i++ {
				putw16(r[i*2:], uint16((uint32(w16(dst[i*2:]))+uint32(w16(src[i*2:]))+1)>>1))
			}
		}
		c.setIntReg(reg, wide, r)
		return true

	// --- insert/extract/mask ---
	case 0xC4: // PINSRW reg, r32/m16, imm8
		reg, o := c.modrmE()
		imm := int(c.fetch8())
		var w uint16
		if o.isReg {
			w = uint16(c.Regs[o.reg])
		} else {
			w = uint16(c.memRead(o.base, o.off, 2))
		}
		if k == ssePD {
			v := c.XMM[reg]
			putw16(v[(imm&7)*2:], w)
			c.XMM[reg] = v
		} else {
			v := c.MMX[reg&7]
			putw16(v[(imm&3)*2:], w)
			c.MMX[reg&7] = v
		}
		return true

	case 0xC5: // PEXTRW r32, reg, imm8
		reg, o := c.modrmE()
		imm := int(c.fetch8())
		if !o.isReg {
			return false
		}
		if k == ssePD {
			c.Regs[reg] = uint32(w16(c.XMM[o.reg][(imm&7)*2:]))
		} else {
			c.Regs[reg] = uint32(w16(c.MMX[o.reg&7][(imm&3)*2:]))
		}
		return true

	case 0xD7: // PMOVMSKB r32, reg
		reg, o := c.modrmE()
		if !o.isReg {
			return false
		}
		var mask, n uint32
		if k == ssePD {
			n = 16
			for i := uint32(0); i < n; i++ {
				if c.XMM[o.reg][i]&0x80 != 0 {
					mask |= 1 << i
				}
			}
		} else {
			n = 8
			for i := uint32(0); i < n; i++ {
				if c.MMX[o.reg&7][i]&0x80 != 0 {
					mask |= 1 << i
				}
			}
		}
		c.Regs[reg] = mask
		return true

	case 0xE7: // MOVNTQ m64 <- mm / MOVNTDQ m128 <- xmm: non-temporal = plain store
		reg, o := c.modrmE()
		if o.isReg {
			return false
		}
		if k == ssePD {
			c.sseStoreRM(o, c.XMM[reg], 16)
		} else {
			c.mmxStoreRM(o, c.MMX[reg&7])
		}
		return true
	}
	return false
}

// satI16Clamp clamps a full int32 to the signed-16 range (PACKSSDW).
func satI16Clamp(v int32) uint16 {
	if v < -32768 {
		return 0x8000
	}
	if v > 32767 {
		return 0x7FFF
	}
	return uint16(v)
}

// shift kinds for shiftLanes.
func shiftLeft(v uint64, bits, cnt uint32) uint64 {
	if cnt >= bits {
		return 0
	}
	mask := uint64(1)<<bits - 1
	if bits == 64 {
		mask = ^uint64(0)
	}
	return (v << cnt) & mask
}

func shiftRightLogical(v uint64, bits, cnt uint32) uint64 {
	if cnt >= bits {
		return 0
	}
	return v >> cnt
}

func shiftRightArith(v uint64, bits, cnt uint32) uint64 {
	if cnt >= bits {
		cnt = bits - 1
	}
	// sign-extend the lane to 64 bits, arithmetic shift, re-mask
	sh := 64 - bits
	s := int64(v<<sh) >> sh
	r := uint64(s >> cnt)
	mask := uint64(1)<<bits - 1
	if bits == 64 {
		mask = ^uint64(0)
	}
	return r & mask
}

// shiftLanes applies a per-lane shift across the first n bytes of v.
func shiftLanes(v *[16]byte, n, elem int, cnt uint32, fn func(uint64, uint32, uint32) uint64) {
	bits := uint32(elem * 8)
	for i := 0; i < n; i += elem {
		var lane uint64
		for j := 0; j < elem; j++ {
			lane |= uint64(v[i+j]) << (8 * j)
		}
		lane = fn(lane, bits, cnt)
		for j := 0; j < elem; j++ {
			v[i+j] = byte(lane >> (8 * j))
		}
	}
}

// byteShift shifts a 16-byte vector left/right by whole bytes (PSLLDQ/PSRLDQ).
func byteShift(v *[16]byte, n int, left bool) {
	if n > 16 {
		n = 16
	}
	var r [16]byte
	if left {
		for i := 15; i >= n; i-- {
			r[i] = v[i-n]
		}
	} else {
		for i := 0; i < 16-n; i++ {
			r[i] = v[i+n]
		}
	}
	*v = r
}
