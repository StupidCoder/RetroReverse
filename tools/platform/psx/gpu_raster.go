package psx

// gpu_raster.go: polygon and rectangle rasterization for gpu.go. Triangles are
// filled with a barycentric edge walk; quads are two triangles. Flat and gouraud
// shading interpolate the vertex colours; textured primitives sample the texture
// page (4/8/15 bpp, with a CLUT for the indexed depths) from VRAM.

type vert struct {
	x, y    int
	r, g, b int
	u, v    int
}

func (g *gpu) polygon(op byte) {
	nv := 3
	if op&0x08 != 0 {
		nv = 4
	}
	gouraud := op&0x10 != 0
	textured := op&0x04 != 0

	vs := make([]vert, nv)
	idx := 1 // flat: opcode word holds the shared colour, vertices follow
	if gouraud {
		idx = 0 // gouraud: opcode word holds the first vertex colour
	}
	flat := g.fifo[0]
	var clut, tpage uint32
	for i := 0; i < nv; i++ {
		col := flat
		if gouraud {
			col = g.fifo[idx]
			idx++
		}
		vs[i].r = int(col & 0xFF)
		vs[i].g = int((col >> 8) & 0xFF)
		vs[i].b = int((col >> 16) & 0xFF)

		xy := g.fifo[idx]
		idx++
		vs[i].x = int(int16(xy&0xFFFF)) + g.offX
		vs[i].y = int(int16(xy>>16)) + g.offY

		if textured {
			uv := g.fifo[idx]
			idx++
			vs[i].u = int(uv & 0xFF)
			vs[i].v = int((uv >> 8) & 0xFF)
			switch i {
			case 0:
				clut = (uv >> 16) & 0xFFFF
			case 1:
				tpage = (uv >> 16) & 0xFFFF
			}
		}
	}
	if textured {
		g.texPageX = int(tpage&0x0F) * 64
		g.texPageY = int((tpage>>4)&1) * 256
		g.texDepth = int((tpage >> 7) & 3)
	}

	g.tri(vs[0], vs[1], vs[2], textured, clut)
	if nv == 4 {
		g.tri(vs[1], vs[2], vs[3], textured, clut)
	}
}

func edge(ax, ay, bx, by, px, py int) int {
	return (bx-ax)*(py-ay) - (by-ay)*(px-ax)
}

// modulate scales a 15-bit texel by an 8-bit-per-channel primitive colour, where
// 0x80 leaves the texel unchanged (the PSX texture-modulation rule). Each 5-bit
// texel lane is multiplied by the colour lane and shifted right 7, then clamped.
//
// An all-zero primitive colour is treated as unmodulated (raw texel) rather than
// black: the game only leaves a textured primitive's colour zero when that colour
// should come from a GTE lighting op we don't model yet (the demo track's INTPL
// fog), and modulating by zero would wrongly black those surfaces out. Lit
// primitives we do model (NCS/NCT — the car) carry non-zero colours and shade
// normally.
func modulate(t uint16, r, g, b int) uint16 {
	if r|g|b == 0 {
		return t
	}
	lane := func(c5 uint16, col int) uint16 {
		v := int(c5) * col >> 7
		if v > 0x1F {
			v = 0x1F
		}
		return uint16(v)
	}
	return lane(t&0x1F, r) | lane((t>>5)&0x1F, g)<<5 | lane((t>>10)&0x1F, b)<<10 | t&0x8000
}

func (g *gpu) tri(a, b, c vert, textured bool, clut uint32) {
	area := edge(a.x, a.y, b.x, b.y, c.x, c.y)
	if area == 0 {
		return
	}
	// Work with a positive area so the inside test and interpolation share a sign.
	if area < 0 {
		a, c = c, a
		area = -area
	}
	rx, by := mini(g.drawR-1, vramW-1), mini(g.drawB-1, vramH-1)
	minX := clampi(min3(a.x, b.x, c.x), g.drawL, rx)
	maxX := clampi(max3(a.x, b.x, c.x), g.drawL, rx)
	minY := clampi(min3(a.y, b.y, c.y), g.drawT, by)
	maxY := clampi(max3(a.y, b.y, c.y), g.drawT, by)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			w0 := edge(b.x, b.y, c.x, c.y, x, y)
			w1 := edge(c.x, c.y, a.x, a.y, x, y)
			w2 := edge(a.x, a.y, b.x, b.y, x, y)
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			var px uint16
			r := (w0*a.r + w1*b.r + w2*c.r) / area
			gg := (w0*a.g + w1*b.g + w2*c.g) / area
			bb := (w0*a.b + w1*b.b + w2*c.b) / area
			if textured {
				u := (w0*a.u + w1*b.u + w2*c.u) / area
				v := (w0*a.v + w1*b.v + w2*c.v) / area
				t, ok := g.texel(u, v, clut)
				if !ok {
					continue // fully transparent texel
				}
				// Modulate the texel by the (interpolated, lit) primitive colour:
				// out = texel * colour / 128, so 0x80 is neutral. Games rely on this
				// to shade lit textured models (GTE NCS/NCT feed the vertex colours).
				px = modulate(t, r, gg, bb)
			} else {
				px = uint16(r>>3) | uint16(gg>>3)<<5 | uint16(bb>>3)<<10
			}
			g.store(x, y, px)
		}
	}
}

// texel samples the current texture page at (u,v); ok is false for the fully
// transparent index-0 / black texel that the GPU treats as see-through.
func (g *gpu) texel(u, v int, clut uint32) (uint16, bool) {
	u &= 0xFF
	v &= 0xFF
	// Texture window (GP0 E2): mask/offset the coords into the active window so
	// texture repeat works. U = (U & ~(mask*8)) | ((offset & mask)*8).
	u = (u &^ (g.texWinMX * 8)) | ((g.texWinOX & g.texWinMX) * 8)
	v = (v &^ (g.texWinMY * 8)) | ((g.texWinOY & g.texWinMY) * 8)
	clutBase := int((clut>>6)&0x1FF)*vramW + int(clut&0x3F)*16
	var raw uint16
	switch g.texDepth {
	case 0: // 4bpp: 4 texels per VRAM word column
		word := g.vram[((g.texPageY+v)&(vramH-1))*vramW+((g.texPageX+u/4)&(vramW-1))]
		shift := uint(u&3) * 4
		raw = g.vram[(clutBase+int((word>>shift)&0xF))&(len(g.vram)-1)]
	case 1: // 8bpp
		word := g.vram[((g.texPageY+v)&(vramH-1))*vramW+((g.texPageX+u/2)&(vramW-1))]
		shift := uint(u&1) * 8
		raw = g.vram[(clutBase+int((word>>shift)&0xFF))&(len(g.vram)-1)]
	default: // 15bpp direct
		raw = g.vram[((g.texPageY+v)&(vramH-1))*vramW+((g.texPageX+u)&(vramW-1))]
	}
	if raw == 0 {
		return 0, false
	}
	return raw, true
}

// rect draws a (possibly textured) rectangle: colour word, top-left, optional
// texcoord+clut, optional size.
func (g *gpu) rect(op byte) {
	col := g.fifo[0]
	xy := g.fifo[1]
	x0 := int(int16(xy&0xFFFF)) + g.offX
	y0 := int(int16(xy>>16)) + g.offY
	idx := 2
	textured := op&0x04 != 0
	var u0, v0 int
	var clut uint32
	if textured {
		uv := g.fifo[idx]
		idx++
		u0 = int(uv & 0xFF)
		v0 = int((uv >> 8) & 0xFF)
		clut = (uv >> 16) & 0xFFFF
	}
	w, h := 1, 1
	switch op & 0x18 {
	case 0x00: // variable
		sz := g.fifo[idx]
		w = int(sz & 0xFFFF)
		h = int((sz >> 16) & 0xFFFF)
	case 0x08:
		w, h = 1, 1
	case 0x10:
		w, h = 8, 8
	case 0x18:
		w, h = 16, 16
	}
	r := uint16(col&0xFF) >> 3
	gg := uint16((col>>8)&0xFF) >> 3
	bb := uint16((col>>16)&0xFF) >> 3
	flat := r | gg<<5 | bb<<10
	for y := 0; y < h; y++ {
		py := y0 + y
		if py < g.drawT || py >= g.drawB || py >= vramH {
			continue
		}
		for x := 0; x < w; x++ {
			px := x0 + x
			if px < g.drawL || px >= g.drawR || px >= vramW {
				continue
			}
			out := flat
			if textured {
				t, ok := g.texel(u0+x, v0+y, clut)
				if !ok {
					continue
				}
				out = t
				// Modulate by the primitive colour unless the raw-texture bit is set.
				if op&0x01 == 0 {
					out = modulate(t, int(col&0xFF), int((col>>8)&0xFF), int((col>>16)&0xFF))
				}
			}
			g.store(px, py, out)
		}
	}
}

func min3(a, b, c int) int { return mini(mini(a, b), c) }
func max3(a, b, c int) int { return maxi(maxi(a, b), c) }
func mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func clampi(v, lo, hi int) int { return maxi(lo, mini(v, hi)) }
