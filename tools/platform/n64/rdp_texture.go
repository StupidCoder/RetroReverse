package n64

// rdp_texture.go samples texels out of TMEM.
//
// TMEM is 4 KiB of on-chip memory that the Load_* commands fill from RDRAM. A
// tile descriptor says how to read it: the pixel format and size, how many
// 64-bit words a row occupies, where the tile starts, and how coordinates
// outside its declared extent behave — wrapped by a mask, clamped, or mirrored.
//
// The colour-indexed formats look their palette up in the upper half of TMEM,
// where Load_TLUT put it. That is why a texture and its palette can coexist:
// they live in different halves of the same memory.
//
// Filtering is where the RDP is unlike every other rasteriser of its era. It
// does not bilinearly average four texels: it takes *three*, the corners of
// whichever half of the texel quad the sample point falls in, and interpolates
// across that triangle. The result is close to bilinear and cheaper, and it is
// why N64 textures have their particular soft, slightly faceted look.

// tlutBase is where Load_TLUT writes, halfway up TMEM.
const tlutBase = 0x800

// texCoord applies a tile's wrap, mirror and clamp rules to one axis.
//
// The mask selects how many low bits survive, which is the wrap. Mirroring
// reflects every other repeat, so the bit above the mask flips the coordinate.
// Clamping, which the hardware only honours when no mask is set, pins the
// coordinate inside the tile's declared extent.
func texCoord(v int32, mask, cm, lo, hi uint32) uint32 {
	if mask == 0 {
		// No mask: clamp into the tile's extent.
		if v < int32(lo) {
			v = int32(lo)
		}
		if v > int32(hi) {
			v = int32(hi)
		}
		return uint32(v - int32(lo))
	}
	v -= int32(lo)
	m := int32(1)<<mask - 1
	if cm&1 != 0 { // clamp
		if v < 0 {
			v = 0
		}
		if v > int32(hi-lo) {
			v = int32(hi - lo)
		}
		return uint32(v)
	}
	if cm&2 != 0 && v&(m+1) != 0 { // mirror: every other repeat runs backwards
		return uint32(m - (v & m))
	}
	return uint32(v & m)
}

// tmem reads a byte of texture memory, wrapping.
func (r *rdp) tmem(off uint32) byte { return r.TMem[off&0xFFF] }

func (r *rdp) tmem16(off uint32) uint16 {
	return uint16(r.tmem(off))<<8 | uint16(r.tmem(off+1))
}

// tlut reads palette entry i, which Load_TLUT stored as a 16-bit value every
// eight bytes.
func (r *rdp) tlut(i uint32) uint16 { return r.tmem16(tlutBase + i*8) }

func fromRGBA16(v uint16) rgba {
	return rgba{
		R: uint32(v>>11&31) << 3,
		G: uint32(v>>6&31) << 3,
		B: uint32(v>>1&31) << 3,
		A: uint32(v&1) * 255,
	}
}

// omSampleType selects filtered sampling over point sampling.
const omSampleType = 1 << 45

// sample reads a texel through a tile. The coordinates are 10.5 fixed point —
// whole texels in the high bits, a fraction of a texel in the low five.
//
// With filtering off the fraction is discarded. With it on, the three-tap filter
// runs: the fraction chooses which half of the texel quad the sample lies in,
// and the three corners of that half are weighted by it. When the fractions sum
// to more than a whole texel the sample is in the far half, and the interpolation
// runs backwards from the opposite corner.
func (m *Machine) sample(t *tile, sFix, tFix int32) (rgba, bool) {
	s, tt := sFix>>5, tFix>>5
	if m.rdp.OtherModes&omSampleType == 0 {
		return m.texelAt(t, s, tt)
	}

	sf, tf := uint32(sFix&31), uint32(tFix&31)
	c00, ok := m.texelAt(t, s, tt)
	if !ok {
		return rgba{}, false
	}
	c10, ok := m.texelAt(t, s+1, tt)
	if !ok {
		return rgba{}, false
	}
	c01, ok := m.texelAt(t, s, tt+1)
	if !ok {
		return rgba{}, false
	}

	if sf+tf < 32 {
		// The near half: interpolate away from (0,0).
		return tap3(c00, c10, c01, sf, tf), true
	}
	// The far half: the opposite corner replaces (0,0), and the weights invert.
	c11, ok := m.texelAt(t, s+1, tt+1)
	if !ok {
		return rgba{}, false
	}
	return tap3(c11, c01, c10, 32-sf, 32-tf), true
}

// tap3 interpolates across the triangle (base, alongS, alongT).
func tap3(base, alongS, alongT rgba, sf, tf uint32) rgba {
	ch := func(b, s, t uint32) uint32 {
		v := int32(b) + (int32(s)-int32(b))*int32(sf)/32 + (int32(t)-int32(b))*int32(tf)/32
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint32(v)
	}
	return rgba{
		ch(base.R, alongS.R, alongT.R),
		ch(base.G, alongS.G, alongT.G),
		ch(base.B, alongS.B, alongT.B),
		ch(base.A, alongS.A, alongT.A),
	}
}

// swizzle undoes TMEM's odd-row word swap.
//
// The hardware stores every odd row of a texture with the two 32-bit halves of
// each 64-bit word exchanged; Load_Block's dxt field drives the swap as the
// block is copied in. The sampler must flip bit 2 of the byte offset on odd rows
// to read a texel back from where it actually landed.
//
// Left unapplied, a texture reads back with its odd rows broken into alternating
// four-byte dashes, which is exactly how this game's 32x32 grass tile appears
// when this returns off unchanged.
func swizzle(off, row uint32) uint32 {
	if row&1 != 0 {
		return off ^ 4
	}
	return off
}

// texelAt reads one texel, in whole texels, through a tile's wrap rules.
func (m *Machine) texelAt(t *tile, s, tt int32) (rgba, bool) {
	r := &m.rdp
	sx := texCoord(s, t.MaskS, t.CMS, t.SL>>2, t.SH>>2)
	ty := texCoord(tt, t.MaskT, t.CMT, t.TL>>2, t.TH>>2)
	row := t.TMem*8 + ty*t.Line*8

	switch {
	case t.Format == fmtRGBA && t.Size == size16:
		return fromRGBA16(r.tmem16(swizzle(row+sx*2, ty))), true
	case t.Format == fmtRGBA && t.Size == size32:
		// A 32-bit texture is split: red and green in the low half of TMEM,
		// blue and alpha in the high half.
		lo := r.tmem16(swizzle(row+sx*2, ty))
		hi := r.tmem16(swizzle(row+sx*2, ty) + 0x800)
		return rgba{uint32(lo >> 8), uint32(lo & 255), uint32(hi >> 8), uint32(hi & 255)}, true

	case t.Format == fmtI && t.Size == size8:
		i := uint32(r.tmem(swizzle(row+sx, ty)))
		return rgba{i, i, i, i}, true
	case t.Format == fmtI && t.Size == size4:
		v := r.tmem(swizzle(row+sx/2, ty))
		n := uint32(v >> 4)
		if sx&1 == 1 {
			n = uint32(v & 15)
		}
		i := n * 17
		return rgba{i, i, i, i}, true

	case t.Format == fmtIA && t.Size == size16:
		v := r.tmem16(swizzle(row+sx*2, ty))
		i := uint32(v >> 8)
		return rgba{i, i, i, uint32(v & 255)}, true
	case t.Format == fmtIA && t.Size == size8:
		v := r.tmem(swizzle(row+sx, ty))
		i := uint32(v>>4) * 17
		return rgba{i, i, i, uint32(v&15) * 17}, true
	case t.Format == fmtIA && t.Size == size4:
		v := r.tmem(swizzle(row+sx/2, ty))
		n := uint32(v >> 4)
		if sx&1 == 1 {
			n = uint32(v & 15)
		}
		i := (n >> 1) * 36 // three bits of intensity, scaled to 0..252
		a := uint32(0)
		if n&1 != 0 {
			a = 255
		}
		return rgba{i, i, i, a}, true

	case t.Format == fmtCI && t.Size == size8:
		return fromRGBA16(r.tlut(uint32(r.tmem(swizzle(row+sx, ty))))), true
	case t.Format == fmtCI && t.Size == size4:
		v := r.tmem(swizzle(row+sx/2, ty))
		n := uint32(v >> 4)
		if sx&1 == 1 {
			n = uint32(v & 15)
		}
		return fromRGBA16(r.tlut(t.Palette<<4 | n)), true
	}

	m.CPU.Halt("unmodelled texture format %d size %d (%d bits)", t.Format, t.Size, 4<<t.Size)
	return rgba{}, false
}

// --- the pixel pipeline ------------------------------------------------------

// readPixel fetches the framebuffer's current colour, which the blender needs.
func (m *Machine) readPixel(x, y uint32) rgba {
	r := &m.rdp
	a := r.pixelAddr(x, y)
	switch r.Color.Size {
	case size16:
		if int(a)+1 >= len(m.RDRAM) {
			return rgba{}
		}
		return fromRGBA16(uint16(m.RDRAM[a])<<8 | uint16(m.RDRAM[a+1]))
	case size32:
		if int(a)+3 >= len(m.RDRAM) {
			return rgba{}
		}
		return rgba{uint32(m.RDRAM[a]), uint32(m.RDRAM[a+1]), uint32(m.RDRAM[a+2]), uint32(m.RDRAM[a+3])}
	}
	return rgba{}
}

// depthAt reads the depth buffer.
func (m *Machine) depthAt(x, y uint32) uint32 {
	r := &m.rdp
	a := r.Mask + (y*r.Color.Width+x)*2
	if int(a)+1 >= len(m.RDRAM) {
		return 0xFFFF
	}
	return uint32(m.RDRAM[a])<<8 | uint32(m.RDRAM[a+1])
}

func (m *Machine) setDepth(x, y, z uint32) {
	r := &m.rdp
	a := r.Mask + (y*r.Color.Width+x)*2
	m.storeRDRAM16(a, uint16(z))
}

// zRanges is the depth encoding's piecewise table, one row per exponent: the
// first z the exponent covers, and how far the mantissa is shifted down.
//
// Each row spans exactly 2048 mantissa steps, so the eight of them tile the whole
// 18-bit range, and the encoding is monotonic across the joins.
var zRanges = [8]struct{ base, shift uint32 }{
	{0x00000, 6}, {0x20000, 5}, {0x30000, 4}, {0x38000, 3},
	{0x3C000, 2}, {0x3E000, 1}, {0x3F000, 0}, {0x3F800, 0},
}

// depthOf converts an interpolated 15.16 depth into the 16-bit value the buffer
// holds.
//
// The buffer does not store depth linearly. It stores a 14-bit floating-point
// value — a 3-bit exponent and an 11-bit mantissa — in the word's top bits, with
// the low two bits reserved for the delta-z the anti-aliaser uses. That gives the
// far plane, where perspective crowds every surface together, far more precision
// than a linear encoding would.
//
// This matters more than "the depth buffer is only compared against itself"
// suggests. A projected scene spends almost all of its z range near the far
// plane: this game's terrain arrives with the integer part of z between 32626 and
// 32748, out of a 32767 maximum. A linear encoding that saturates anywhere in
// that band collapses the entire world onto one depth value, the comparison stops
// discriminating, and whatever is drawn last wins.
func depthOf(z int64) uint32 {
	// The encoder works on 18 bits: the 15-bit integer part and three bits of
	// fraction. The rest of the fraction is below the buffer's resolution.
	if z < 0 {
		z = 0
	}
	v := uint32(z >> 13)
	if v > 0x3FFFF {
		v = 0x3FFFF
	}

	exp := 7
	for i := 1; i < 8; i++ {
		if v < zRanges[i].base {
			exp = i - 1
			break
		}
	}
	r := zRanges[exp]
	mantissa := (v - r.base) >> r.shift

	// The encoded value sits above the two delta-z bits, which is why a cleared
	// buffer reads 0xFFFC rather than 0xFFFF.
	return (uint32(exp)<<11 | mantissa&0x7FF) << 2
}

// drawPixel runs one pixel through the combiner, the depth test and the blender,
// and stores it. It is the single place a shaded pixel reaches memory.
func (m *Machine) drawPixel(x, y uint32, in *combineInputs, z int64, useZ bool) {
	r := &m.rdp

	if useZ && r.OtherModes&omZCompare != 0 && r.Mask != 0 {
		zv := depthOf(z)
		if zv >= m.depthAt(x, y) {
			return // hidden by something already drawn
		}
	}

	col := r.combine(in)

	// The alpha test discards a pixel outright, which is how a cut-out texture
	// gets its holes.
	if r.OtherModes&omAlphaCompare != 0 && col.A < r.BlendColor&255 {
		return
	}

	out := col
	if r.OtherModes&omForceBlend != 0 || r.blenderReadsMemory() {
		out = r.blend(col, m.readPixel(x, y), in.Shade.A)
	}
	m.writePixel(x, y, out.R, out.G, out.B, out.A)

	if useZ && r.OtherModes&omZUpdate != 0 && r.Mask != 0 {
		m.setDepth(x, y, depthOf(z))
	}
}

// blenderReadsMemory reports whether either blender source is the framebuffer,
// which is the only case where the pixel underneath matters.
func (r *rdp) blenderReadsMemory() bool {
	cycle := 0
	if r.cycleType() == cycle2 {
		cycle = 1
	}
	p, _, mm, b := r.blenderSelects(cycle)
	return p == 1 || mm == 1 || b == 1
}
