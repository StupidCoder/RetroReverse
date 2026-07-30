package dc

// pvr.go is the beginning of the rasteriser: enough of the PVR to turn the
// frame's recorded TA parameter stream into pixels in the write framebuffer
// when STARTRENDER fires. It is deliberately partial and loudly so — every
// parameter shape it skips is censused — but what it draws, it draws from
// the real stream with real textures:
//
//   - sprite parameters (type 5): textured or flat quads, 16-bit UVs, drawn
//     as two triangles;
//   - polygon vertices (type 4/7): triangle strips in packed, floating or
//     intensity colour (face colour from the header, mode 2 reusing the
//     previous face colour), textured or flat, 32-bit or 16-bit UVs.
//
// Depth is the PVR's own convention — vertices carry 1/w, larger is closer —
// kept in a full-frame z-buffer and tested per pixel with the ISP compare
// mode; UVs interpolate perspective-correctly (u/w, v/w, 1/w are the
// screen-linear quantities and the hardware hands us 1/w as z). The region
// array and tile binning are bypassed: the recorded stream is rasterised
// directly at full frame. Translucent parameters blend source-over in
// submission order — the auto-sort's depth ordering is not reproduced, and
// scenes that rely on it will say so on screen.
//
// The VRAM array is the 64-bit path's own layout, so texture control words
// address it directly; the framebuffer pointers are 32-bit-path addresses
// and go through the bank interleave (machine.go's vram32to64). Textures
// decode 1555/565/4444, twiddled or raster order, VQ and PAL4/PAL8,
// mipmapped (sampled at the top level, chain offsets from the format's own
// layout). YUV and offset colour log once; unimplemented formats draw
// magenta, never black.

import (
	"encoding/binary"
	"math"
)

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
	// The clear is COMMAND ONE of the capture — it writes every pixel of the
	// frame, and it is usually the answer to "why is this pixel black" — so
	// the hook fires before its pixels do.
	if m.OnPVRClear != nil {
		m.OnPVRClear(w, h)
	}
	m.clearFB(base, w, h)
	cmds := 1

	st := renderState{m: m, base: base, w: w, h: h,
		zbuf: make([]float32, w*h), faceCol: 0xFFFFFFFF}
	stream := m.TAClosed
	vtx := uint32(32)
	for off := 0; off+32 <= len(stream); {
		if m.RenderStopAfter > 0 && cmds >= m.RenderStopAfter {
			break // the scrubber's halt: a plain Go loop needs no sentinel
		}
		pcw := binary.LittleEndian.Uint32(stream[off:])
		typ := pcw >> 29
		// The parameter's size, decided exactly as the walk below advances,
		// so the hook carries the parameter's full bytes — a 64-byte header's
		// face colours included.
		sz := 32
		switch typ {
		case 4:
			textured := pcw&8 != 0
			if (pcw&0x40 != 0 && textured) || (pcw>>4&3 == 2 && pcw&4 != 0 && textured) {
				sz = 64
			}
		case 7:
			if st.sprite {
				sz = 64
			} else {
				sz = int(vtx)
			}
		}
		if m.OnPVRCmd != nil {
			end := off + sz
			if end > len(stream) {
				end = len(stream)
			}
			m.OnPVRCmd(stream[off:end])
		}
		cmds++
		switch typ {
		case 0, 1, 2: // EOL / user clip / OL set
		case 4: // polygon header
			st.loadHeader(pcw, stream[off:])
			st.sprite = false
			st.strip = 0
			textured := pcw&8 != 0
			colType := pcw >> 4 & 3
			offsCol := pcw&4 != 0
			volume := pcw&0x40 != 0
			st.colType = colType
			st.uv16 = textured && pcw&1 != 0
			vtx = 32
			if volume || (textured && colType == 1) || pcw>>24&7 == 1 || pcw>>24&7 == 3 {
				vtx = 64
			}
			switch {
			case volume && textured:
			case colType == 2 && offsCol && textured:
				// intensity with offset: face colours in words 8-15
				if off+64 <= len(stream) {
					st.faceCol = packFloatCol(stream[off+32:])
					st.faceOffs = packFloatCol(stream[off+48:])
				}
			case colType == 2:
				// intensity, no offset: face colour in words 4-7
				st.faceCol = packFloatCol(stream[off+16:])
			}
			// colour type 3 reuses the previous face colour as-is
		case 5: // sprite header
			st.loadHeader(pcw, stream[off:])
			st.sprite = true
			vtx = 64
		case 7: // vertex
			if st.sprite {
				if off+64 <= len(stream) {
					st.drawSprite(stream[off:])
				}
			} else {
				st.pushStripVertex(stream[off:], pcw)
			}
		}
		off += sz
	}
	m.logf("render: stream %dB, %d tris, %d px -> %06X %dx%d", len(stream), st.tris, st.px, base, w, h)
}

// clearFB zeroes the write framebuffer; FB_W_SOF is a 32-bit-path address,
// so each word goes through the bank interleave. With the pixel hook armed
// it reports every pixel it writes — a clear that writes memory without
// reporting fragments leaves provenance claiming nothing wrote the whole
// background, which is false in the one place a user is sure to click.
func (m *Machine) clearFB(base uint32, w, h int) {
	n := uint32(w * h * 2)
	if base+n > VRAMSize {
		n = VRAMSize - base
	}
	for i := uint32(0); i+4 <= n; i += 4 {
		off := vram32to64(base + i)
		m.VRAM[off] = 0
		m.VRAM[off+1] = 0
		m.VRAM[off+2] = 0
		m.VRAM[off+3] = 0
		if m.OnPVRPixel != nil && w > 0 {
			p := int(i / 2) // the word's two 565 pixels
			m.OnPVRPixel(p%w, p/w, 0, 0, 0, 255, true)
			m.OnPVRPixel((p+1)%w, (p+1)/w, 0, 0, 0, 255, true)
		}
	}
}

// renderState is the live header context during a frame walk.
type renderState struct {
	m        *Machine
	base     uint32
	w, h     int
	zbuf     []float32
	sprite   bool
	strip    int
	tris     int
	px       int
	sv       [3]pvrVert
	texture  bool
	blend    bool
	gouraud  bool
	uv16     bool
	colType  uint32
	offsEn   bool   // the header's offset-colour bit (textured only)
	shade    uint32 // TSP shading instruction: decal/modulate/decal-A/modulate-A
	faceCol  uint32 // intensity modes scale this; mode 3 reuses the last one
	faceOffs uint32 // the 64-byte intensity header's offset face colour
	depthCmp uint32
	zWrite   bool
	texAddr  uint32
	texMipIx uint32 // VQ mip: index-area offset to the top level
	texFmt   uint32
	texTwid  bool
	texVQ    bool
	texBank  uint32
	texUW    int
	texVH    int
	baseCol  uint32
	baseOffs uint32
}

type pvrVert struct {
	x, y, z float32 // z is the PVR's 1/w: larger is closer
	u, v    float32
	color   uint32
	offs    uint32 // offset colour, added after the shading instruction
}

func (st *renderState) skip(what string, v uint32) {
	st.m.logf("render: %s %d unimplemented", what, v)
}

// loadHeader captures the ISP/TSP/texture words of a global parameter.
func (st *renderState) loadHeader(pcw uint32, p []byte) {
	isp := binary.LittleEndian.Uint32(p[4:])
	tsp := binary.LittleEndian.Uint32(p[8:])
	tcw := binary.LittleEndian.Uint32(p[12:])
	st.texture = pcw&8 != 0
	st.gouraud = pcw&2 != 0
	list := pcw >> 24 & 7
	st.blend = list == 2 // the translucent list blends source-over
	st.offsEn = pcw&4 != 0 && st.texture
	st.shade = tsp >> 6 & 3
	st.depthCmp = isp >> 29 & 7
	st.zWrite = isp>>26&1 == 0
	st.texUW = 8 << (tsp >> 3 & 7)
	st.texVH = 8 << (tsp & 7)
	st.texAddr = tcw & 0x1FFFFF << 3
	st.texMipIx = 0
	st.texFmt = tcw >> 27 & 7
	st.texTwid = tcw>>26&1 == 0
	st.texVQ = tcw>>30&1 != 0
	if st.texFmt == 6 {
		st.texBank = tcw >> 25 & 3
	} else {
		st.texBank = tcw >> 21 & 0x3F
	}
	if tcw>>31&1 != 0 {
		// Mipmapped: the address names the chain (smallest level first, a few
		// dummy texels of padding at the front); sample the top level.
		if st.texVQ {
			st.texMipIx = vqMipIndexOffset(st.texUW)
		} else {
			st.texAddr += mipTopOffset(st.texUW, st.texFmt)
		}
		st.texVH = st.texUW // mipmapped textures are square
	}
	// Sprites carry their base colour in the header's word 4 and their
	// offset colour in word 5.
	if len(p) >= 24 {
		st.baseCol = binary.LittleEndian.Uint32(p[16:])
		st.baseOffs = binary.LittleEndian.Uint32(p[20:])
	}
}

// mipTopOffset is the byte offset of the top (largest) mip level: the levels
// below it hold (side²-1)/3 texels, plus three dummy texels of padding at the
// start of the chain.
func mipTopOffset(side int, fmt uint32) uint32 {
	texels := uint32((side*side-1)/3 + 3)
	switch fmt {
	case 5: // PAL4: a nibble per texel
		return texels / 2
	case 6: // PAL8
		return texels
	default: // 16-bit formats
		return texels * 2
	}
}

// vqMipIndexOffset is the byte offset of the top level within a VQ texture's
// index area: one index byte per 2x2 block, smaller levels first (a level
// below 2x2 still spends a byte).
func vqMipIndexOffset(side int) uint32 {
	var off uint32
	for s := 1; s < side; s *= 2 {
		n := s * s / 4
		if n < 1 {
			n = 1
		}
		off += uint32(n)
	}
	return off
}

// packFloatCol packs a header's A,R,G,B floats into ARGB8888.
func packFloatCol(p []byte) uint32 {
	cl := func(f float32) uint32 {
		if f <= 0 {
			return 0
		}
		if f >= 1 {
			return 255
		}
		return uint32(f * 255)
	}
	return cl(f32(p))<<24 | cl(f32(p[4:]))<<16 | cl(f32(p[8:]))<<8 | cl(f32(p[12:]))
}

func f32(p []byte) float32 {
	return float32frombits(binary.LittleEndian.Uint32(p))
}

// drawSprite draws a type-5 quad: corners A,B,C given fully, D's position in
// x/y only, its z and UV derived (the quad is planar and a parallelogram in
// texture space).
func (st *renderState) drawSprite(p []byte) {
	ax, ay, az := f32(p[4:]), f32(p[8:]), f32(p[12:])
	bx, by, bz := f32(p[16:]), f32(p[20:]), f32(p[24:])
	cx, cy, cz := f32(p[28:]), f32(p[32:]), f32(p[36:])
	dx, dy := f32(p[40:]), f32(p[44:])
	au, av := unpackUV16(binary.LittleEndian.Uint32(p[52:]))
	bu, bv := unpackUV16(binary.LittleEndian.Uint32(p[56:]))
	cu, cv := unpackUV16(binary.LittleEndian.Uint32(p[60:]))
	du, dv := au+cu-bu, av+cv-bv
	dz := az + cz - bz
	offs := uint32(0)
	if st.offsEn {
		offs = st.baseOffs
	}
	a := pvrVert{ax, ay, az, au, av, st.baseCol, offs}
	b := pvrVert{bx, by, bz, bu, bv, st.baseCol, offs}
	c := pvrVert{cx, cy, cz, cu, cv, st.baseCol, offs}
	d := pvrVert{dx, dy, dz, du, dv, st.baseCol, offs}
	st.tri(a, b, c)
	st.tri(a, c, d)
}

// pushStripVertex feeds a polygon-list vertex into the running triangle
// strip, resolving the header's colour type and UV width.
func (st *renderState) pushStripVertex(p []byte, pcw uint32) {
	v := pvrVert{
		x: f32(p[4:]),
		y: f32(p[8:]),
		z: f32(p[12:]),
	}
	if st.texture {
		if st.uv16 {
			v.u, v.v = unpackUV16(binary.LittleEndian.Uint32(p[16:]))
		} else {
			v.u, v.v = f32(p[16:]), f32(p[20:])
		}
	}
	switch st.colType {
	case 1: // floating colour: A,R,G,B floats, after the UVs when textured
		if st.texture {
			v.color = packFloatCol(p[32:])
			if st.offsEn {
				v.offs = packFloatCol(p[48:])
			}
		} else {
			v.color = packFloatCol(p[16:])
		}
	case 2, 3: // intensity: the face colours scaled by the vertex's intensities
		v.color = scaleCol(st.faceCol, f32(p[24:]))
		if st.offsEn {
			v.offs = scaleCol(st.faceOffs, f32(p[28:]))
		}
	default: // packed colour in word 6, packed offset colour in word 7
		v.color = binary.LittleEndian.Uint32(p[24:])
		if st.offsEn {
			v.offs = binary.LittleEndian.Uint32(p[28:])
		}
		if !st.texture {
			v.color = binary.LittleEndian.Uint32(p[16:])
		}
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

// scaleCol multiplies a packed colour's RGB by an intensity, keeping alpha.
func scaleCol(col uint32, i float32) uint32 {
	if i <= 0 {
		return col & 0xFF000000
	}
	if i >= 1 {
		return col
	}
	r := uint32(float32(col>>16&0xFF) * i)
	g := uint32(float32(col>>8&0xFF) * i)
	b := uint32(float32(col&0xFF) * i)
	return col&0xFF000000 | r<<16 | g<<8 | b
}

// lerp3 interpolates one byte-wide channel of three packed colours.
func lerp3(w0, w1, w2 float32, a, b, c uint32, shift uint) uint32 {
	return uint32(w0*float32(a>>shift&0xFF) + w1*float32(b>>shift&0xFF) + w2*float32(c>>shift&0xFF))
}

// shadePixel applies the TSP shading instruction to a sampled texel: decal,
// modulate, decal-alpha or modulate-alpha, plus the offset colour when the
// header enables it. The base and offset colours come from the vertices,
// gouraud-interpolated when the header asks, flat from the first otherwise.
func (st *renderState) shadePixel(w0, w1, w2 float32, a, b, c pvrVert, tr, tg, tb, ta uint32) (cr, cg, cb, ca uint32) {
	var br, bg, bb, ba uint32
	var or, og, ob uint32
	if st.gouraud {
		br = lerp3(w0, w1, w2, a.color, b.color, c.color, 16)
		bg = lerp3(w0, w1, w2, a.color, b.color, c.color, 8)
		bb = lerp3(w0, w1, w2, a.color, b.color, c.color, 0)
		ba = lerp3(w0, w1, w2, a.color, b.color, c.color, 24)
		if st.offsEn {
			or = lerp3(w0, w1, w2, a.offs, b.offs, c.offs, 16)
			og = lerp3(w0, w1, w2, a.offs, b.offs, c.offs, 8)
			ob = lerp3(w0, w1, w2, a.offs, b.offs, c.offs, 0)
		}
	} else {
		br, bg, bb, ba = a.color>>16&0xFF, a.color>>8&0xFF, a.color&0xFF, a.color>>24&0xFF
		if st.offsEn {
			or, og, ob = a.offs>>16&0xFF, a.offs>>8&0xFF, a.offs&0xFF
		}
	}
	switch st.shade {
	case 0: // decal: the texture as-is
		cr, cg, cb, ca = tr, tg, tb, ta
	case 1: // modulate
		cr, cg, cb, ca = tr*br/255, tg*bg/255, tb*bb/255, ta
	case 2: // decal alpha: texture over base by the texture's own alpha
		cr = (tr*ta + br*(255-ta)) / 255
		cg = (tg*ta + bg*(255-ta)) / 255
		cb = (tb*ta + bb*(255-ta)) / 255
		ca = ba
	default: // modulate alpha
		cr, cg, cb, ca = tr*br/255, tg*bg/255, tb*bb/255, ta*ba/255
	}
	cr, cg, cb = clamp255(cr+or), clamp255(cg+og), clamp255(cb+ob)
	return
}

func clamp255(v uint32) uint32 {
	if v > 255 {
		return 255
	}
	return v
}

// depthPass applies the header's ISP compare mode; z is 1/w, larger closer.
func depthPass(mode uint32, z, old float32) bool {
	switch mode {
	case 0:
		return false
	case 1:
		return z < old
	case 2:
		return z == old
	case 3:
		return z <= old
	case 4:
		return z > old
	case 5:
		return z != old
	case 6:
		return z >= old
	default:
		return true
	}
}

// tri rasterises one triangle: depth-tested, UVs perspective-correct (the
// screen-linear quantities are u·z, v·z and z itself), colours gouraud when
// the header asks.
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
	if debugPixelX >= minX && debugPixelX <= maxX && debugPixelY >= minY && debugPixelY <= maxY {
		st.m.logf("debug-pixel tri tex=%v addr=%06X fmt=%d shade=%d offsEn=%v colType=%d gouraud=%v %dx%d blend=%v a.color=%08X a.offs=%08X", st.texture, st.texAddr, st.texFmt, st.shade, st.offsEn, st.colType, st.gouraud, st.texUW, st.texVH, st.blend, a.color, a.offs)
	}
	st.tris++
	inv := 1 / area
	au, av := a.u*a.z, a.v*a.z
	bu, bv := b.u*b.z, b.v*b.z
	cu, cv := c.u*c.z, c.v*c.z
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			px, py := float32(x)+0.5, float32(y)+0.5
			w0 := ((b.x-px)*(c.y-py) - (c.x-px)*(b.y-py)) * inv
			w1 := ((c.x-px)*(a.y-py) - (a.x-px)*(c.y-py)) * inv
			w2 := 1 - w0 - w1
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			z := w0*a.z + w1*b.z + w2*c.z
			zi := y*st.w + x
			// Depth-rejected fragments are deliberately NOT reported to the
			// debugger's pixel hook: the only place that knows about them is
			// this loop, and any textual edit here shifts the compiler's
			// float codegen (FMA contraction) and moved the pinned drive
			// picture by two one-quantum pixels. Alpha-rejects still report
			// through plot; a depth-reject is invisible to overdraw. An
			// honest absence, like the 3DO's.
			if !depthPass(st.depthCmp, z, st.zbuf[zi]) {
				continue
			}
			var cr, cg, cb, ca uint32
			if st.texture {
				u := w0*au + w1*bu + w2*cu
				v := w0*av + w1*bv + w2*cv
				if z != 0 {
					u, v = u/z, v/z
				}
				tr, tg, tb, ta := st.sample(u, v)
				cr, cg, cb, ca = st.shadePixel(w0, w1, w2, a, b, c, tr, tg, tb, ta)
			} else if st.gouraud {
				cr = uint32(w0*float32(a.color>>16&0xFF) + w1*float32(b.color>>16&0xFF) + w2*float32(c.color>>16&0xFF))
				cg = uint32(w0*float32(a.color>>8&0xFF) + w1*float32(b.color>>8&0xFF) + w2*float32(c.color>>8&0xFF))
				cb = uint32(w0*float32(a.color&0xFF) + w1*float32(b.color&0xFF) + w2*float32(c.color&0xFF))
				ca = uint32(w0*float32(a.color>>24&0xFF) + w1*float32(b.color>>24&0xFF) + w2*float32(c.color>>24&0xFF))
			} else {
				col := a.color
				cr, cg, cb, ca = col>>16&0xFF, col>>8&0xFF, col&0xFF, col>>24&0xFF
			}
			if st.plot(x, y, cr, cg, cb, ca) {
				if x == debugPixelX && y == debugPixelY {
					st.m.logf("debug-pixel plot tex=%v addr=%06X rgba=%d,%d,%d,%d z=%v", st.texture, st.texAddr, cr, cg, cb, ca, z)
				}
				if st.zWrite {
					st.zbuf[zi] = z
				}
			}
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
		iOff := st.texAddr + 2048 + st.texMipIx + block
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

// plot writes one 565 pixel, blending when the list asks for it; it reports
// whether the pixel landed so the caller knows to update the z-buffer.
//
// The debugger's fragment hook fires HERE rather than in tri()'s loop, and
// the placement is load-bearing: any textual edit inside tri's pixel loop
// shifts the compiler's float codegen (FMA contraction — the Gekko lesson)
// and moved the drive gate's picture by two one-quantum pixels. plot is
// all-integer, so hooking it cannot move a float. The colour reported is
// the fragment's own source colour, pre-blend.
func (st *renderState) plot(x, y int, r, g, b, a uint32) bool {
	drawn := st.plotStore(x, y, r, g, b, a)
	if st.m.OnPVRPixel != nil {
		st.m.OnPVRPixel(x, y, uint8(r), uint8(g), uint8(b), uint8(a), drawn)
	}
	return drawn
}

func (st *renderState) plotStore(x, y int, r, g, b, a uint32) bool {
	off := st.base + uint32(y*st.w+x)*2
	if off+2 > VRAMSize {
		return false
	}
	// The write pointer is a 32-bit-path address; a 565 pixel's two bytes
	// stay adjacent through the interleave.
	off = vram32to64(off)
	if a == 0 {
		return false
	}
	if st.blend {
		old := uint32(st.m.VRAM[off]) | uint32(st.m.VRAM[off+1])<<8
		or := old >> 11 & 31 << 3
		og := old >> 5 & 63 << 2
		ob := old & 31 << 3
		r = (r*a + or*(255-a)) / 255
		g = (g*a + og*(255-a)) / 255
		b = (b*a + ob*(255-a)) / 255
	}
	st.px++
	px := uint16(r>>3<<11 | g>>2<<5 | b>>3)
	st.m.VRAM[off] = uint8(px)
	st.m.VRAM[off+1] = uint8(px >> 8)
	return true
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

// debugPixel: when >= 0, every triangle whose bounding box covers the pixel
// logs its texture binding — a temporary lens for "what drew this garble".
var debugPixelX, debugPixelY = -1, -1
