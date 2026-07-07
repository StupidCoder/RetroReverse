package mips

// gte.go implements the Geometry Transformation Engine (COP2), the PlayStation's
// fixed-point 3-D coprocessor. It transforms and perspective-projects vertices,
// does lighting, and maintains the screen-XY / screen-Z ordering-table FIFOs.
// Everything is integer/fixed-point — no floats — and the saturation and overflow
// FLAG semantics matter, so this reproduces them bit-for-bit, including the
// hardware unsigned Newton-Raphson reciprocal used by the perspective divide.
//
// State is held as the raw 32-bit data[]/ctrl[] register files (the layout the
// CPU sees through mfc2/mtc2/cfc2/ctc2) and decoded on demand inside each
// command. It implements the mips.Coprocessor2 interface; wire it to a CPU with
//
//	cpu.GTE = mips.NewGTE()
//
// Implemented commands: RTPS, RTPT, NCLIP, AVSZ3, AVSZ4, MVMVA — the transform +
// ordering-table + matrix-vector subset a racer's pipeline leans on. Lighting
// commands (NCDS, NCCS, CDP, ...) are not yet modelled; they are accepted and
// leave the accumulators untouched rather than faulting, so a boot trace does
// not stall (see the note in Command).

// GTE FLAG register bits (data/ctrl reg 31 for ctrl; here the calc-error flags).
const (
	flagIR0     = 1 << 12
	flagSY2     = 1 << 13
	flagSX2     = 1 << 14
	flagMac0Neg = 1 << 15
	flagMac0Pos = 1 << 16
	flagDivOvf  = 1 << 17
	flagSZ3OTZ  = 1 << 18
	flagIR3     = 1 << 22
	flagIR2     = 1 << 23
	flagIR1     = 1 << 24

	flagErrMask = 0x7F87E000 // bits 30..23 and 18..13 → sets bit 31
)

// GTE is the COP2 state.
type GTE struct {
	data [32]uint32
	ctrl [32]uint32
}

// NewGTE returns a zeroed geometry engine.
func NewGTE() *GTE { return &GTE{} }

// s16 sign-extends the low 16 bits.
func s16(v uint32) int32 { return int32(int16(v)) }

// --- register interface (mips.Coprocessor2) --------------------------------

// Read returns data register reg (mfc2 / swc2 source).
func (g *GTE) Read(reg uint32) uint32 {
	switch reg & 31 {
	case 15: // SXYP mirrors SXY2
		return g.data[14]
	case 28, 29: // IRGB / ORGB: pack IR1..IR3 into 5-bit lanes
		r := clamp5(s16(g.data[9]) >> 7)
		gg := clamp5(s16(g.data[10]) >> 7)
		b := clamp5(s16(g.data[11]) >> 7)
		return uint32(r | gg<<5 | b<<10)
	default:
		return g.data[reg&31]
	}
}

// Write sets data register reg (mtc2 / lwc2 dest), honouring the FIFO and
// derived registers.
func (g *GTE) Write(reg, v uint32) {
	switch reg & 31 {
	case 15: // SXYP: manual push of the SXY FIFO
		g.data[12] = g.data[13]
		g.data[13] = g.data[14]
		g.data[14] = v
	case 28: // IRGB: unpack 5-bit lanes into IR1..IR3 (×0x80)
		g.data[9] = uint32(int32(v&0x1F) << 7)
		g.data[10] = uint32(int32((v>>5)&0x1F) << 7)
		g.data[11] = uint32(int32((v>>10)&0x1F) << 7)
		g.data[28] = v & 0x7FFF
	case 30: // LZCS: store and recompute LZCR
		g.data[30] = v
		g.data[31] = lzc(v)
	case 7, 23, 29, 31: // OTZ hi bits / RES / ORGB / LZCR are read-mostly; store raw low
		g.data[reg&31] = v
	default:
		g.data[reg&31] = v
	}
}

// ReadCtrl returns control register reg (cfc2).
func (g *GTE) ReadCtrl(reg uint32) uint32 { return g.ctrl[reg&31] }

// WriteCtrl sets control register reg (ctc2).
func (g *GTE) WriteCtrl(reg, v uint32) { g.ctrl[reg&31] = v }

// clamp5 saturates a value to the 0..0x1F colour lane.
func clamp5(v int32) int32 {
	if v < 0 {
		return 0
	}
	if v > 0x1F {
		return 0x1F
	}
	return v
}

// lzc counts the leading bits equal to bit 31 (leading zeros of a positive
// value, leading ones of a negative one), 1..32.
func lzc(v uint32) uint32 {
	top := v & 0x80000000
	n := uint32(0)
	for i := 0; i < 32; i++ {
		if v&0x80000000 != top {
			break
		}
		n++
		v <<= 1
	}
	return n
}

// --- field accessors -------------------------------------------------------

func (g *GTE) vec(n int) (int32, int32, int32) {
	switch n {
	case 0:
		return s16(g.data[0]), s16(g.data[0] >> 16), s16(g.data[1])
	case 1:
		return s16(g.data[2]), s16(g.data[2] >> 16), s16(g.data[3])
	default:
		return s16(g.data[4]), s16(g.data[4] >> 16), s16(g.data[5])
	}
}

// rt reads rotation-matrix entry [r][c] from ctrl[0..4].
func (g *GTE) rt(r, c int) int32 { return matEntry(g.ctrl[0:5], r, c) }

// light reads the light matrix from ctrl[8..12], colour matrix from ctrl[16..20].
func (g *GTE) light(r, c int) int32 { return matEntry(g.ctrl[8:13], r, c) }
func (g *GTE) color(r, c int) int32 { return matEntry(g.ctrl[16:21], r, c) }

// matEntry decodes a 3×3 signed-16 matrix packed into five 32-bit words
// (m00,m01 | m02,m10 | m11,m12 | m20,m21 | m22).
func matEntry(w []uint32, r, c int) int32 {
	idx := r*3 + c // 0..8
	switch idx {
	case 0:
		return s16(w[0])
	case 1:
		return s16(w[0] >> 16)
	case 2:
		return s16(w[1])
	case 3:
		return s16(w[1] >> 16)
	case 4:
		return s16(w[2])
	case 5:
		return s16(w[2] >> 16)
	case 6:
		return s16(w[3])
	case 7:
		return s16(w[3] >> 16)
	default:
		return s16(w[4])
	}
}

func (g *GTE) tr(i int) int32  { return int32(g.ctrl[5+i]) }  // translation
func (g *GTE) bk(i int) int32  { return int32(g.ctrl[13+i]) } // background colour
func (g *GTE) fc(i int) int32  { return int32(g.ctrl[21+i]) } // far colour
func (g *GTE) ofx() int32      { return int32(g.ctrl[24]) }
func (g *GTE) ofy() int32      { return int32(g.ctrl[25]) }
func (g *GTE) h() uint32       { return g.ctrl[26] & 0xFFFF }
func (g *GTE) dqa() int32      { return s16(g.ctrl[27]) }
func (g *GTE) dqb() int32      { return int32(g.ctrl[28]) }
func (g *GTE) zsf3() int32     { return s16(g.ctrl[29]) }
func (g *GTE) zsf4() int32     { return s16(g.ctrl[30]) }
func (g *GTE) irVal(i int) int32 { return s16(g.data[8+i]) }

// --- flag / saturation helpers ---------------------------------------------

func (g *GTE) setFlag(bits uint32) { g.ctrl[31] |= bits }
func (g *GTE) clearFlags()        { g.ctrl[31] = 0 }
func (g *GTE) finishFlags() {
	if g.ctrl[31]&flagErrMask != 0 {
		g.ctrl[31] |= 1 << 31
	}
}

// macCheck flags a 44-bit overflow of MAC1..3 (i in 1..3).
func (g *GTE) macCheck(i int, val int64) {
	if val > (1<<43)-1 {
		g.setFlag(1 << (31 - uint(i))) // MAC1 pos = bit30 ...
	} else if val < -(1 << 43) {
		g.setFlag(1 << (28 - uint(i))) // MAC1 neg = bit27 ...
	}
}

func (g *GTE) setMac(i int, v int32) { g.data[24+i] = uint32(v) }

// setIR saturates value into IRi (i in 1..3) and stores it sign-extended.
func (g *GTE) setIR(i int, v int32, lm bool) {
	min := int32(-0x8000)
	if lm {
		min = 0
	}
	if v > 0x7FFF {
		v = 0x7FFF
		g.setFlag(flagIR1 >> uint(i-1))
	} else if v < min {
		v = min
		g.setFlag(flagIR1 >> uint(i-1))
	}
	g.data[8+i] = uint32(v)
}

// setMacIR runs the MAC-then-IR pipeline for lane i.
func (g *GTE) setMacIR(i int, val int64, sf uint, lm bool) {
	g.macCheck(i, val)
	m := int32(val >> sf)
	g.setMac(i, m)
	g.setIR(i, m, lm)
}

func (g *GTE) setIR0(v int32) {
	if v < 0 {
		v = 0
		g.setFlag(flagIR0)
	} else if v > 0x1000 {
		v = 0x1000
		g.setFlag(flagIR0)
	}
	g.data[8] = uint32(v)
}

func (g *GTE) setMac0(val int64) {
	if val > (1<<31)-1 {
		g.setFlag(flagMac0Pos)
	} else if val < -(1 << 31) {
		g.setFlag(flagMac0Neg)
	}
	g.data[24] = uint32(int32(val))
}

// pushSZ3 clamps a Z value to 0..FFFF and pushes the SZ FIFO (SZ0..SZ3).
func (g *GTE) pushSZ3(v int32) {
	if v < 0 {
		v = 0
		g.setFlag(flagSZ3OTZ)
	} else if v > 0xFFFF {
		v = 0xFFFF
		g.setFlag(flagSZ3OTZ)
	}
	g.data[16] = g.data[17]
	g.data[17] = g.data[18]
	g.data[18] = g.data[19]
	g.data[19] = uint32(v)
}

func (g *GTE) setOTZ(v int32) {
	if v < 0 {
		v = 0
		g.setFlag(flagSZ3OTZ)
	} else if v > 0xFFFF {
		v = 0xFFFF
		g.setFlag(flagSZ3OTZ)
	}
	g.data[7] = uint32(v)
}

// pushSXY clamps SX/SY to -0x400..0x3FF and pushes the SXY FIFO.
func (g *GTE) pushSXY(x, y int32) {
	sx := clampSXY(x, flagSX2, g)
	sy := clampSXY(y, flagSY2, g)
	g.data[12] = g.data[13]
	g.data[13] = g.data[14]
	g.data[14] = uint32(uint16(int16(sx))) | uint32(uint16(int16(sy)))<<16
}

func clampSXY(v int32, bit uint32, g *GTE) int32 {
	if v < -0x400 {
		g.setFlag(bit)
		return -0x400
	}
	if v > 0x3FF {
		g.setFlag(bit)
		return 0x3FF
	}
	return v
}

// --- Command dispatch ------------------------------------------------------

// Command executes a GTE command (the 25-bit COP2 command field).
func (g *GTE) Command(cmd uint32) {
	g.clearFlags()
	sf := uint(0)
	if cmd&(1<<19) != 0 {
		sf = 12
	}
	lm := cmd&(1<<10) != 0

	switch cmd & 0x3F {
	case 0x01: // RTPS
		g.rtp(0, sf, lm, true)
	case 0x30: // RTPT
		g.rtp(0, sf, lm, false)
		g.rtp(1, sf, lm, false)
		g.rtp(2, sf, lm, true)
	case 0x06: // NCLIP
		g.nclip()
	case 0x12: // MVMVA
		g.mvmva(cmd, sf, lm)
	case 0x2D: // AVSZ3
		g.avsz(false)
	case 0x2E: // AVSZ4
		g.avsz(true)
	default:
		// Lighting/colour ops not yet modelled; accept without faulting so a boot
		// trace continues (geometry will be incomplete until they are added).
	}
	g.finishFlags()
}

// rtp performs Rotate-Translate-Perspective on vector n. depth requests the
// depth-cue IR0 step (done once, on the last vertex of an RTPT).
func (g *GTE) rtp(n int, sf uint, lm bool, depth bool) {
	vx, vy, vz := g.vec(n)
	g.setMacIR(1, int64(g.tr(0))<<12+int64(g.rt(0, 0))*int64(vx)+int64(g.rt(0, 1))*int64(vy)+int64(g.rt(0, 2))*int64(vz), sf, lm)
	g.setMacIR(2, int64(g.tr(1))<<12+int64(g.rt(1, 0))*int64(vx)+int64(g.rt(1, 1))*int64(vy)+int64(g.rt(1, 2))*int64(vz), sf, lm)

	mac3full := int64(g.tr(2))<<12 + int64(g.rt(2, 0))*int64(vx) + int64(g.rt(2, 1))*int64(vy) + int64(g.rt(2, 2))*int64(vz)
	g.macCheck(3, mac3full)
	g.setMac(3, int32(mac3full>>sf))
	g.setIR(3, int32(mac3full>>sf), lm)
	g.pushSZ3(int32(mac3full >> 12)) // SZ always uses the sf=1 (>>12) value

	div := g.divide(g.h(), g.data[19]&0xFFFF)

	macx := int64(div)*int64(g.irVal(1)) + int64(g.ofx())
	g.setMac0(macx)
	macy := int64(div)*int64(g.irVal(2)) + int64(g.ofy())
	sx := int32(macx >> 16)
	g.setMac0(macy)
	sy := int32(macy >> 16)
	g.pushSXY(sx, sy)

	if depth {
		macz := int64(div)*int64(g.dqa()) + int64(g.dqb())
		g.setMac0(macz)
		g.setIR0(int32(macz >> 12))
	}
}

// nclip computes the signed area (cross product) of the SXY0/1/2 triangle into
// MAC0 — used for back-face culling.
func (g *GTE) nclip() {
	sx0, sy0 := s16(g.data[12]), s16(g.data[12]>>16)
	sx1, sy1 := s16(g.data[13]), s16(g.data[13]>>16)
	sx2, sy2 := s16(g.data[14]), s16(g.data[14]>>16)
	g.setMac0(int64(sx0)*int64(sy1) + int64(sx1)*int64(sy2) + int64(sx2)*int64(sy0) -
		int64(sx0)*int64(sy2) - int64(sx1)*int64(sy0) - int64(sx2)*int64(sy1))
}

// avsz averages 3 or 4 screen-Z values into OTZ via the ZSF scale factor.
func (g *GTE) avsz(four bool) {
	var sum int64
	var zsf int32
	if four {
		zsf = g.zsf4()
		sum = int64(g.data[16]&0xFFFF) + int64(g.data[17]&0xFFFF) + int64(g.data[18]&0xFFFF) + int64(g.data[19]&0xFFFF)
	} else {
		zsf = g.zsf3()
		sum = int64(g.data[17]&0xFFFF) + int64(g.data[18]&0xFFFF) + int64(g.data[19]&0xFFFF)
	}
	mac0 := int64(zsf) * sum
	g.setMac0(mac0)
	g.setOTZ(int32(mac0 >> 12))
}

// mvmva multiplies a selected matrix by a selected vector plus a selected
// translation, per the command's mx/v/cv fields.
func (g *GTE) mvmva(cmd uint32, sf uint, lm bool) {
	mx := (cmd >> 17) & 3
	vsel := (cmd >> 15) & 3
	cv := (cmd >> 13) & 3

	mat := func(r, c int) int32 {
		switch mx {
		case 0:
			return g.rt(r, c)
		case 1:
			return g.light(r, c)
		default:
			return g.color(r, c)
		}
	}
	var vx, vy, vz int32
	switch vsel {
	case 0, 1, 2:
		vx, vy, vz = g.vec(int(vsel))
	default:
		vx, vy, vz = g.irVal(1), g.irVal(2), g.irVal(3)
	}
	tr := func(i int) int64 {
		switch cv {
		case 0:
			return int64(g.tr(i))
		case 1:
			return int64(g.bk(i))
		case 2:
			return int64(g.fc(i))
		default:
			return 0
		}
	}
	for i := 0; i < 3; i++ {
		val := tr(i)<<12 + int64(mat(i, 0))*int64(vx) + int64(mat(i, 1))*int64(vy) + int64(mat(i, 2))*int64(vz)
		g.setMacIR(i+1, val, sf, lm)
	}
}

// --- perspective divide (hardware unsigned Newton-Raphson) -----------------

// divide computes (h*0x20000/sz3 + 1)/2 the way the GTE hardware does, via the
// 257-entry reciprocal seed table, clamped to 0x1FFFF (setting the divide-
// overflow flag when sz3*2 <= h).
func (g *GTE) divide(h, sz3 uint32) uint32 {
	if h >= sz3*2 { // includes sz3 == 0
		g.setFlag(flagDivOvf)
		return 0x1FFFF
	}
	z := lzc16(uint16(sz3))
	n := h << z
	d := sz3 << z
	u := uint32(unrTable[(d-0x7FC0)>>7]) + 0x101
	d = (0x2000080 - d*u) >> 8
	d = (0x0000080 + d*u) >> 8
	res := (uint64(n)*uint64(d) + 0x8000) >> 16
	if res > 0x1FFFF {
		return 0x1FFFF
	}
	return uint32(res)
}

// lzc16 counts leading zeros of a 16-bit value (0..15; 16 stays out of range).
func lzc16(v uint16) uint32 {
	n := uint32(0)
	for i := 0; i < 16; i++ {
		if v&0x8000 != 0 {
			break
		}
		n++
		v <<= 1
	}
	return n
}

// unrTable is the GTE reciprocal seed table (Nocash psx-spx).
var unrTable = [257]uint8{
	0xFF, 0xFD, 0xFB, 0xF9, 0xF7, 0xF5, 0xF3, 0xF1, 0xEF, 0xEE, 0xEC, 0xEA, 0xE8, 0xE6, 0xE4, 0xE3,
	0xE1, 0xDF, 0xDD, 0xDC, 0xDA, 0xD8, 0xD6, 0xD5, 0xD3, 0xD1, 0xD0, 0xCE, 0xCD, 0xCB, 0xC9, 0xC8,
	0xC6, 0xC5, 0xC3, 0xC1, 0xC0, 0xBE, 0xBD, 0xBB, 0xBA, 0xB8, 0xB7, 0xB5, 0xB4, 0xB2, 0xB1, 0xB0,
	0xAE, 0xAD, 0xAB, 0xAA, 0xA9, 0xA7, 0xA6, 0xA4, 0xA3, 0xA2, 0xA0, 0x9F, 0x9E, 0x9C, 0x9B, 0x9A,
	0x99, 0x97, 0x96, 0x95, 0x94, 0x92, 0x91, 0x90, 0x8F, 0x8D, 0x8C, 0x8B, 0x8A, 0x89, 0x87, 0x86,
	0x85, 0x84, 0x83, 0x82, 0x81, 0x7F, 0x7E, 0x7D, 0x7C, 0x7B, 0x7A, 0x79, 0x78, 0x77, 0x75, 0x74,
	0x73, 0x72, 0x71, 0x70, 0x6F, 0x6E, 0x6D, 0x6C, 0x6B, 0x6A, 0x69, 0x68, 0x67, 0x66, 0x65, 0x64,
	0x63, 0x62, 0x61, 0x60, 0x5F, 0x5E, 0x5D, 0x5D, 0x5C, 0x5B, 0x5A, 0x59, 0x58, 0x57, 0x56, 0x55,
	0x54, 0x53, 0x53, 0x52, 0x51, 0x50, 0x4F, 0x4E, 0x4D, 0x4D, 0x4C, 0x4B, 0x4A, 0x49, 0x48, 0x48,
	0x47, 0x46, 0x45, 0x44, 0x43, 0x43, 0x42, 0x41, 0x40, 0x3F, 0x3F, 0x3E, 0x3D, 0x3C, 0x3C, 0x3B,
	0x3A, 0x39, 0x39, 0x38, 0x37, 0x36, 0x36, 0x35, 0x34, 0x33, 0x33, 0x32, 0x31, 0x31, 0x30, 0x2F,
	0x2E, 0x2E, 0x2D, 0x2C, 0x2C, 0x2B, 0x2A, 0x2A, 0x29, 0x28, 0x28, 0x27, 0x26, 0x26, 0x25, 0x24,
	0x24, 0x23, 0x22, 0x22, 0x21, 0x20, 0x20, 0x1F, 0x1E, 0x1E, 0x1D, 0x1D, 0x1C, 0x1B, 0x1B, 0x1A,
	0x19, 0x19, 0x18, 0x18, 0x17, 0x16, 0x16, 0x15, 0x15, 0x14, 0x14, 0x13, 0x12, 0x12, 0x11, 0x11,
	0x10, 0x0F, 0x0F, 0x0E, 0x0E, 0x0D, 0x0D, 0x0C, 0x0C, 0x0B, 0x0A, 0x0A, 0x09, 0x09, 0x08, 0x08,
	0x07, 0x07, 0x06, 0x06, 0x05, 0x05, 0x04, 0x04, 0x03, 0x03, 0x02, 0x02, 0x01, 0x01, 0x00, 0x00,
	0x00,
}
