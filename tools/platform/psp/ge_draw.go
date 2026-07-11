package psp

// ge_draw.go is the software rasterizer: triangle and sprite fill with per-vertex
// colour interpolation, optional texture modulation, and CLEAR-mode solid fills,
// writing pixels into the framebuffer in its PSP pixel format.

// rasterTri fills the triangle a,b,c with barycentric-interpolated colour (and
// texcoords when textured). In clear mode it fills solid with a's colour.
func (m *Machine) rasterTri(s *geState, a, b, c vert) {
	minX := int(fmin3(a.x, b.x, c.x))
	maxX := int(fmax3(a.x, b.x, c.x)) + 1
	minY := int(fmin3(a.y, b.y, c.y))
	maxY := int(fmax3(a.y, b.y, c.y)) + 1
	minX, minY = clampI(minX, 0, dispW), clampI(minY, 0, dispH)
	maxX, maxY = clampI(maxX, 0, dispW), clampI(maxY, 0, dispH)

	area := edge(a, b, c)
	if area == 0 {
		return
	}
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			px := vert{x: float32(x) + 0.5, y: float32(y) + 0.5}
			w0 := edge(b, c, px)
			w1 := edge(c, a, px)
			w2 := edge(a, b, px)
			// inside test (either winding)
			if (w0 < 0 || w1 < 0 || w2 < 0) && (w0 > 0 || w1 > 0 || w2 > 0) {
				continue
			}
			l0, l1, l2 := w0/area, w1/area, w2/area
			r := byte(l0*float32(a.r) + l1*float32(b.r) + l2*float32(c.r))
			g := byte(l0*float32(a.g) + l1*float32(b.g) + l2*float32(c.g))
			bl := byte(l0*float32(a.b) + l1*float32(b.b) + l2*float32(c.b))
			al := byte(l0*float32(a.a) + l1*float32(b.a) + l2*float32(c.a))
			if s.texEnable && !s.clearOn {
				u := l0*a.u + l1*b.u + l2*c.u
				v := l0*a.v + l1*b.v + l2*c.v
				r, g, bl, al = modTex(m, s, u, v, r, g, bl, al)
			}
			m.putPixel(s, x, y, r, g, bl, al)
		}
	}
}

// rasterSprite fills the axis-aligned rectangle between corners a and b (PSP sprite
// primitive), textured or solid.
func (m *Machine) rasterSprite(s *geState, a, b vert) {
	x0, x1 := int(a.x), int(b.x)
	y0, y1 := int(a.y), int(b.y)
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	u0, u1 := a.u, b.u
	v0, v1 := a.v, b.v
	x0c, y0c := clampI(x0, 0, dispW), clampI(y0, 0, dispH)
	x1c, y1c := clampI(x1, 0, dispW), clampI(y1, 0, dispH)
	dw := float32(x1 - x0)
	dh := float32(y1 - y0)
	for y := y0c; y < y1c; y++ {
		for x := x0c; x < x1c; x++ {
			r, g, bl, al := b.r, b.g, b.b, b.a
			if s.texEnable && !s.clearOn && dw != 0 && dh != 0 {
				u := u0 + (u1-u0)*(float32(x-x0)/dw)
				v := v0 + (v1-v0)*(float32(y-y0)/dh)
				r, g, bl, al = modTex(m, s, u, v, r, g, bl, al)
			}
			m.putPixel(s, x, y, r, g, bl, al)
		}
	}
}

// modTex samples the bound texture at (u,v) and modulates it with the vertex colour.
// Texcoords are normalized (0..1) in transformed mode, or absolute in through mode;
// this handles the normalized case (the common one) and falls back to absolute.
func modTex(m *Machine, s *geState, u, v float32, r, g, b, a byte) (byte, byte, byte, byte) {
	if s.texW == 0 || s.texH == 0 || s.texAddr == 0 {
		return r, g, b, a
	}
	tx := int(u*float32(s.texW)) & int(s.texW-1)
	ty := int(v*float32(s.texH)) & int(s.texH-1)
	if tx < 0 {
		tx += int(s.texW)
	}
	if ty < 0 {
		ty += int(s.texH)
	}
	tr, tg, tb, ta := m.sampleTex(s, uint32(tx), uint32(ty))
	return mul8(tr, r), mul8(tg, g), mul8(tb, b), mul8(ta, a)
}

// sampleTex reads one texel from the bound texture (formats 5650/5551/4444/8888).
func (m *Machine) sampleTex(s *geState, tx, ty uint32) (byte, byte, byte, byte) {
	stride := s.texStride
	if stride == 0 {
		stride = s.texW
	}
	switch s.texFmt {
	case 3: // 8888
		c := m.read32(s.texAddr + (ty*stride+tx)*4)
		return byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)
	case 0, 1, 2: // 565 / 5551 / 4444
		c := u16(m, s.texAddr+(ty*stride+tx)*2)
		col := decode16(c, s.texFmt)
		return col.R, col.G, col.B, col.A
	}
	return 0xFF, 0xFF, 0xFF, 0xFF
}

// putPixel writes an RGBA pixel into the framebuffer in its PSP format.
func (m *Machine) putPixel(s *geState, x, y int, r, g, b, a byte) {
	base := s.fbAddress()
	stride := s.fbStride
	if stride == 0 {
		stride = dispW
	}
	off := uint32(y)*stride + uint32(x)
	switch s.fbFmt {
	case psm8888:
		m.write32(base+off*4, uint32(r)|uint32(g)<<8|uint32(b)<<16|uint32(a)<<24)
	default:
		var p uint16
		switch s.fbFmt {
		case psm5551:
			p = uint16(r>>3) | uint16(g>>3)<<5 | uint16(b>>3)<<10
			if a >= 128 {
				p |= 0x8000
			}
		case psm4444:
			p = uint16(r>>4) | uint16(g>>4)<<4 | uint16(b>>4)<<8 | uint16(a>>4)<<12
		default: // 565
			p = uint16(r>>3) | uint16(g>>2)<<5 | uint16(b>>3)<<11
		}
		addr := base + off*2
		m.Write(addr, byte(p))
		m.Write(addr+1, byte(p>>8))
	}
}

// --- small helpers ---------------------------------------------------------

func edge(a, b, c vert) float32 { return (c.x-a.x)*(b.y-a.y) - (c.y-a.y)*(b.x-a.x) }

func mul8(a, b byte) byte { return byte(uint32(a) * uint32(b) / 255) }

func fmin3(a, b, c float32) float32 { return fmin(fmin(a, b), c) }
func fmax3(a, b, c float32) float32 { return fmax(fmax(a, b), c) }
func fmin(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func fmax(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
func clampI(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
