package dc

// pvr.go is the beginning of the rasteriser: enough of the PVR to turn the
// frame's recorded TA parameter stream into pixels in the write framebuffer
// when STARTRENDER fires. It is deliberately partial and loudly so — every
// parameter shape it skips is censused — but what it draws, it draws from
// the real stream with real textures:
//
//   - sprite parameters (type 5): textured or flat quads, 16-bit UVs, drawn
//     as two affine triangles;
//   - polygon vertices (type 4/7): packed-colour triangle strips, textured
//     or flat.
//
// The region array and tile binning are bypassed: the recorded stream is
// rasterised directly at full frame, which ignores per-tile depth and sort
// order. Translucent parameters blend source-over in submission order —
// correct for the UI/logo scenes this milestone targets, wrong for depth-
// sorted world geometry, and the census will say so when the time comes.
//
// Textures live in the linear VRAM model (the 64-bit-path interleave is
// still the machine's declared simplification) and decode 1555/565/4444,
// twiddled or raster order. Palettes, VQ, mipmaps and YUV log once and draw
// magenta, so an unimplemented format can never pass for black.

import (
	"encoding/binary"
	"math"
)

// taRecord appends a submitted polygon-path stream to the frame recording.
func (m *Machine) taRecord(src, byteLen uint32) {
	for i := uint32(0); i < byteLen; i += 4 {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], m.ram32(src+i))
		m.TAFrame = append(m.TAFrame, b[:]...)
	}
}

// renderFrame rasterises the recorded stream into the write framebuffer.
// Called when the deferred STARTRENDER completion lands.
func (m *Machine) renderFrame() {
	fbW := m.PVRRegs[0x48/4]
	base := m.PVRRegs[0x60/4] & 0x00FFFFFF
	if fbW&7 != 1 {
		m.logf("render: FB_W_CTRL packmode %d unimplemented (only 565)", fbW&7)
		return
	}
	size := m.PVRRegs[fbRSize]
	w := (int(size&0x3FF) + 1) * 2 // 32-bit units per line -> 565 pixels
	h := int(size>>10&0x3FF) + 1

	// The background: ISP_BACKGND_T names a tag in the parameter space; a
	// real implementation reads the background plane's colour from it. Until
	// then the frame clears to black, which is also what the intro wants.
	m.clearFB(base, w, h)

	st := renderState{m: m, base: base, w: w, h: h}
	stream := m.TAClosed
	vtx := uint32(32)
	for off := 0; off+32 <= len(stream); {
		pcw := binary.LittleEndian.Uint32(stream[off:])
		typ := pcw >> 29
		switch typ {
		case 0, 1, 2: // EOL / user clip / OL set
			off += 32
		case 4: // polygon header
			st.loadHeader(pcw, stream[off:])
			st.sprite = false
			st.strip = 0
			textured := pcw&8 != 0
			colType := pcw >> 4 & 3
			volume := pcw&0x40 != 0
			vtx = 32
			if volume || (textured && colType == 1) || pcw>>24&7 == 1 || pcw>>24&7 == 3 {
				vtx = 64
			}
			off += 32
			if volume && textured {
				off += 32
			}
			if colType != 0 {
				st.skip("polygon colour type", colType)
			}
		case 5: // sprite header
			st.loadHeader(pcw, stream[off:])
			st.sprite = true
			vtx = 64
			off += 32
		case 7: // vertex
			if st.sprite {
				if off+64 <= len(stream) {
					st.drawSprite(stream[off:])
				}
				off += 64
			} else {
				st.pushStripVertex(stream[off:], pcw)
				off += int(vtx)
			}
		default:
			off += 32
		}
	}
	m.logf("render: stream %dB, %d tris, %d px", len(stream), st.tris, st.px)
}

func (m *Machine) clearFB(base uint32, w, h int) {
	n := uint32(w * h * 2)
	if base+n > VRAMSize {
		n = VRAMSize - base
	}
	for i := uint32(0); i < n; i++ {
		m.VRAM[base+i] = 0
	}
}

// renderState is the live header context during a frame walk.
type renderState struct {
	m       *Machine
	base    uint32
	w, h    int
	sprite  bool
	strip   int
	tris    int
	px      int
	sv      [3]pvrVert
	texture bool
	blend   bool
	texAddr uint32
	texFmt  uint32
	texTwid bool
	texVQ   bool
	texBank uint32
	texUW   int
	texVH   int
	baseCol uint32
}

type pvrVert struct {
	x, y  float32
	u, v  float32
	color uint32
}

func (st *renderState) skip(what string, v uint32) {
	st.m.logf("render: %s %d unimplemented", what, v)
}

// loadHeader captures the TSP/texture words of a global parameter.
func (st *renderState) loadHeader(pcw uint32, p []byte) {
	tsp := binary.LittleEndian.Uint32(p[8:])
	tcw := binary.LittleEndian.Uint32(p[12:])
	st.texture = pcw&8 != 0
	list := pcw >> 24 & 7
	st.blend = list == 2 // the translucent list blends source-over
	st.texUW = 8 << (tsp >> 3 & 7)
	st.texVH = 8 << (tsp & 7)
	st.texAddr = tcw & 0x1FFFFF << 3
	st.texFmt = tcw >> 27 & 7
	st.texTwid = tcw>>26&1 == 0
	st.texVQ = tcw>>30&1 != 0
	if st.texFmt == 6 {
		st.texBank = tcw >> 25 & 3
	} else {
		st.texBank = tcw >> 21 & 0x3F
	}
	if tcw>>31&1 != 0 {
		st.skip("mipmapped texture", 1) // sampled at full size; base level offset ignored
	}
	// Sprites carry their base colour in the header's word 4.
	if len(p) >= 20 {
		st.baseCol = binary.LittleEndian.Uint32(p[16:])
	}
}

func f32(p []byte) float32 {
	return float32frombits(binary.LittleEndian.Uint32(p))
}

// drawSprite draws a type-5 quad: corners A,B,C given fully, D's position in
// x/y only and its UV derived (the quad is a parallelogram in texture space).
func (st *renderState) drawSprite(p []byte) {
	ax, ay := f32(p[4:]), f32(p[8:])
	bx, by := f32(p[16:]), f32(p[20:])
	cx, cy := f32(p[28:]), f32(p[32:])
	dx, dy := f32(p[40:]), f32(p[44:])
	au, av := unpackUV16(binary.LittleEndian.Uint32(p[52:]))
	bu, bv := unpackUV16(binary.LittleEndian.Uint32(p[56:]))
	cu, cv := unpackUV16(binary.LittleEndian.Uint32(p[60:]))
	du, dv := au+cu-bu, av+cv-bv
	a := pvrVert{ax, ay, au, av, st.baseCol}
	b := pvrVert{bx, by, bu, bv, st.baseCol}
	c := pvrVert{cx, cy, cu, cv, st.baseCol}
	d := pvrVert{dx, dy, du, dv, st.baseCol}
	st.tri(a, b, c)
	st.tri(a, c, d)
}

// pushStripVertex feeds a polygon-list vertex (packed colour) into the
// running triangle strip.
func (st *renderState) pushStripVertex(p []byte, pcw uint32) {
	v := pvrVert{
		x:     f32(p[4:]),
		y:     f32(p[8:]),
		u:     f32(p[16:]),
		v:     f32(p[20:]),
		color: binary.LittleEndian.Uint32(p[24:]),
	}
	if !st.texture {
		v.color = binary.LittleEndian.Uint32(p[16:])
	}
	if st.strip < 3 {
		st.sv[st.strip] = v
		st.strip++
	} else {
		st.sv[0], st.sv[1] = st.sv[1], st.sv[2]
		st.sv[2] = v
	}
	if st.strip == 3 {
		st.tri(st.sv[0], st.sv[1], st.sv[2])
	}
	if pcw>>28&1 != 0 { // end of strip
		st.strip = 0
	}
}

// tri rasterises one affine triangle.
func (st *renderState) tri(a, b, c pvrVert) {
	minX := imax(0, int(min3(a.x, b.x, c.x)))
	maxX := imin(st.w-1, int(max3(a.x, b.x, c.x)))
	minY := imax(0, int(min3(a.y, b.y, c.y)))
	maxY := imin(st.h-1, int(max3(a.y, b.y, c.y)))
	if minX > maxX || minY > maxY {
		return
	}
	area := (b.x-a.x)*(c.y-a.y) - (c.x-a.x)*(b.y-a.y)
	if area == 0 {
		return
	}
	st.tris++
	inv := 1 / area
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			px, py := float32(x)+0.5, float32(y)+0.5
			w0 := ((b.x-px)*(c.y-py) - (c.x-px)*(b.y-py)) * inv
			w1 := ((c.x-px)*(a.y-py) - (a.x-px)*(c.y-py)) * inv
			w2 := 1 - w0 - w1
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			var cr, cg, cb, ca uint32
			if st.texture {
				u := w0*a.u + w1*b.u + w2*c.u
				v := w0*a.v + w1*b.v + w2*c.v
				cr, cg, cb, ca = st.sample(u, v)
			} else {
				col := a.color
				cr, cg, cb, ca = col>>16&0xFF, col>>8&0xFF, col&0xFF, col>>24&0xFF
			}
			st.plot(x, y, cr, cg, cb, ca)
		}
	}
}

// sample reads the bound texture at normalised u,v with point sampling.
func (st *renderState) sample(u, v float32) (r, g, b, a uint32) {
	tx := int(u*float32(st.texUW)) & (st.texUW - 1)
	ty := int(v*float32(st.texVH)) & (st.texVH - 1)
	// Palettized formats index the palette RAM at 5F9000 through the bank
	// bits of the texture-control word; PAL_RAM_CTRL says how an entry
	// decodes.
	if st.texFmt == 5 || st.texFmt == 6 {
		side := st.texUW
		if st.texVH < side {
			side = st.texVH
		}
		block := (ty/side)*(st.texUW/side) + tx/side
		idx := uint32(block*side*side) + twiddle(uint32(tx%side), uint32(ty%side))
		var entry uint32
		if st.texFmt == 6 { // PAL8: a byte per texel, bank from tcw bits 25-26
			off := st.texAddr + idx
			if off >= VRAMSize {
				return 0, 0, 0, 0
			}
			entry = st.texBank<<8 | uint32(st.m.VRAM[off])
		} else { // PAL4: a nibble per texel
			off := st.texAddr + idx/2
			if off >= VRAMSize {
				return 0, 0, 0, 0
			}
			entry = st.texBank<<4 | uint32(st.m.VRAM[off]>>(4*(idx&1)))&0xF
		}
		pal := st.m.PVRRegs[0x1000/4+entry&0x3FF]
		switch st.m.PVRRegs[0x108/4] & 3 {
		case 0: // ARGB1555
			return pal >> 10 & 31 << 3, pal >> 5 & 31 << 3, pal & 31 << 3, pal >> 15 & 1 * 255
		case 1: // RGB565
			return pal >> 11 & 31 << 3, pal >> 5 & 63 << 2, pal & 31 << 3, 255
		case 2: // ARGB4444
			return pal >> 8 & 15 * 17, pal >> 4 & 15 * 17, pal & 15 * 17, pal >> 12 & 15 * 17
		default: // ARGB8888
			return pal >> 16 & 0xFF, pal >> 8 & 0xFF, pal & 0xFF, pal >> 24 & 0xFF
		}
	}

	var px uint32
	if st.texVQ {
		// VQ: a 2048-byte codebook of 256 2x2-texel entries, then one index
		// byte per block, blocks in twiddled order.
		bw := st.texUW / 2
		bh := st.texVH / 2
		side := bw
		if bh < side {
			side = bh
		}
		bx, by := tx/2, ty/2
		block := uint32((by/side)*(bw/side)+bx/side)*uint32(side*side) + twiddle(uint32(bx%side), uint32(by%side))
		iOff := st.texAddr + 2048 + block
		if iOff >= VRAMSize {
			return 0, 0, 0, 0
		}
		entry := st.texAddr + uint32(st.m.VRAM[iOff])*8 + uint32((tx&1)*2+(ty&1))*2
		if entry+2 > VRAMSize {
			return 0, 0, 0, 0
		}
		px = uint32(st.m.VRAM[entry]) | uint32(st.m.VRAM[entry+1])<<8
	} else {
		var idx uint32
		if st.texTwid {
			side := st.texUW
			if st.texVH < side {
				side = st.texVH
			}
			block := (ty/side)*(st.texUW/side) + tx/side
			idx = uint32(block*side*side) + twiddle(uint32(tx%side), uint32(ty%side))
		} else {
			idx = uint32(ty*st.texUW + tx)
		}
		off := st.texAddr + idx*2
		if off+2 > VRAMSize {
			return 0, 0, 0, 0
		}
		px = uint32(st.m.VRAM[off]) | uint32(st.m.VRAM[off+1])<<8
	}
	switch st.texFmt {
	case 0: // ARGB1555
		a = px >> 15 & 1 * 255
		r = px >> 10 & 31 << 3
		g = px >> 5 & 31 << 3
		b = px & 31 << 3
	case 1: // RGB565
		a = 255
		r = px >> 11 & 31 << 3
		g = px >> 5 & 63 << 2
		b = px & 31 << 3
	case 2: // ARGB4444
		a = px >> 12 & 15 * 17
		r = px >> 8 & 15 * 17
		g = px >> 4 & 15 * 17
		b = px & 15 * 17
	default:
		st.skip("texture format", st.texFmt)
		return 255, 0, 255, 255 // magenta: unimplemented must not pass for black
	}
	return
}

// plot writes one 565 pixel, blending when the list asks for it.
func (st *renderState) plot(x, y int, r, g, b, a uint32) {
	off := st.base + uint32(y*st.w+x)*2
	if off+2 > VRAMSize {
		return
	}
	if st.blend {
		if a == 0 {
			return
		}
		old := uint32(st.m.VRAM[off]) | uint32(st.m.VRAM[off+1])<<8
		or := old >> 11 & 31 << 3
		og := old >> 5 & 63 << 2
		ob := old & 31 << 3
		r = (r*a + or*(255-a)) / 255
		g = (g*a + og*(255-a)) / 255
		b = (b*a + ob*(255-a)) / 255
	} else if a == 0 {
		return
	}
	st.px++
	px := uint16(r>>3<<11 | g>>2<<5 | b>>3)
	st.m.VRAM[off] = uint8(px)
	st.m.VRAM[off+1] = uint8(px >> 8)
}

// twiddle interleaves x into odd bits and y into even bits — the PVR's
// Morton texture order within a square.
func twiddle(x, y uint32) uint32 {
	var out uint32
	for i := uint32(0); i < 16; i++ {
		out |= (y >> i & 1) << (2 * i)
		out |= (x >> i & 1) << (2*i + 1)
	}
	return out
}

func unpackUV16(w uint32) (u, v float32) {
	return float32frombits(w & 0xFFFF0000), float32frombits(w << 16)
}

func float32frombits(b uint32) float32 { return math.Float32frombits(b) }

func min3(a, b, c float32) float32 {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

func max3(a, b, c float32) float32 {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

func imin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func imax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
