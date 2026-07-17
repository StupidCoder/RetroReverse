package xbox

// nv2a_raster.go rasterises Kelvin triangles into the current render surface with
// perspective-correct interpolation, the depth/alpha/blend/mask pipeline, and the
// register combiners (nv2a_combiner.go) for the fragment color. It is the repo's
// barycentric fill (psx/n3ds shape) against the surface state the title itself
// programmed (SET_SURFACE_* — nv2a_frame.go's constants).
//
// Coordinate contract: the vertex program's position output is already in screen
// space — the Xbox D3D runtime appends the viewport transform to every program (the
// c[58]/c[59] scale/offset epilogue), so oPos.xy arrive in pixels, oPos.z in depth-
// buffer units, and oPos.w preserves the clip-space w for perspective correction.
// The surface's anti-aliasing mode scales logical pixels to stored samples: AA=1
// (CENTER_CORNER_2) doubles x, AA=2 (SQUARE_OFFSET_4) doubles x and y — OutRun's
// loading phase renders a 320x240 logical clip into a 640x480-sample pitch-2560
// surface, which is exactly what the survey's clear rect + pitch said.

// PixelEvent is one fragment the rasteriser produced, kept or killed — the debugger's
// per-pixel provenance record (the n3ds/gc shape). A depth-killed fragment carries no
// color; an alpha-tested one carries the color it would have had; a drawn one carries
// what actually reached memory.
type PixelEvent struct {
	Drawn                bool
	ZReject, AlphaReject bool
	R, G, B, A           uint8
}

// Kelvin raster-state registers (byte offsets into the Regs file).
const (
	kelvinAlphaTestEnable = 0x0300
	kelvinBlendEnable     = 0x0304
	kelvinCullFaceEnable  = 0x0308
	kelvinDepthTestEnable = 0x030C
	kelvinAlphaFunc       = 0x033C
	kelvinAlphaRef        = 0x0340
	kelvinBlendSrcFactor  = 0x0344
	kelvinBlendDstFactor  = 0x0348
	kelvinBlendColor      = 0x034C
	kelvinBlendEquation   = 0x0350
	kelvinDepthFunc       = 0x0354
	kelvinColorMask       = 0x0358
	kelvinDepthWriteMask  = 0x035C
	kelvinShadeMode       = 0x037C
	kelvinCullFace        = 0x039C
	kelvinFrontFace       = 0x03A0

	kelvinWindowClipType = 0x02B4
	kelvinWindowClipH    = 0x02C0 // x1 | x2<<16, inclusive
	kelvinWindowClipV    = 0x02E0
)

// rasterState is one draw's decoded pipeline state (registers read once per draw).
type rasterState struct {
	colorPhys   uint32 // physical base of the color surface
	zetaPhys    uint32
	colorPitch  uint32
	zetaPitch   uint32
	aaX, aaY    int // logical->sample scale from the surface's anti-aliasing mode
	x0, y0      int // scissor, in samples (inclusive lo, exclusive hi)
	x1, y1      int
	surfW       int // sample-space surface extent (from the clip rect)
	surfH       int
	swizzle     bool // color+zeta stored Morton-interleaved (SURFACE_FORMAT type 2), not pitch
	swW, swH    int  // swizzle extent in samples (the allocated power-of-two square)
	hasZeta     bool
	depthTest   bool
	depthWrite  bool
	depthFunc   uint32
	alphaTest   bool
	alphaFunc   uint32
	alphaRef    float32
	blend       bool
	blendSrc    uint32
	blendDst    uint32
	blendEq     uint32
	blendConst  [4]float32
	colorMask   [4]bool // A, R, G, B
	cull        bool
	cullFace    uint32
	frontCW     bool
	flatShade   bool
	comb        combState
	texEnable   [4]bool
	texImg      [4]*texImage
	texStage    [4]uint32 // shader stage program per unit
	texWrapU    [4]uint32
	texWrapV    [4]uint32
	texBilinear [4]bool
	texRect     [4]bool // linear formats address in texels, not normalized coords
}

// surfaceAAScale decodes the surface format's anti-aliasing field into the
// logical-pixel -> stored-sample scale.
func surfaceAAScale(format uint32) (ax, ay int) {
	switch format >> 12 & 0xF {
	case 1: // CENTER_CORNER_2
		return 2, 1
	case 2: // SQUARE_OFFSET_4
		return 2, 2
	}
	return 1, 1
}

// decodeSwizzleExtent reads a swizzled (type-2) surface's power-of-two extent from
// SURFACE_FORMAT and cross-checks it against SURFACE_CLIP.
//
// The extent is where SET_SURFACE_PITCH would be for a pitch surface: a swizzled surface
// has no pitch, so the two format nibbles that pitch surfaces leave at zero carry log2 of
// the dimensions instead — bits 16-19 and 24-27 (a pitch surface's format is 0x00000128;
// OutRun's swizzled one is 0x07070228, the two 7s giving 1<<7 = 128 on each axis). Three
// facts agree on 128x128 for OutRun's render-to-texture and make this a derivation, not a
// guess: (1) these nibbles read 7/7; (2) SURFACE_CLIP is 128x128; (3) the title packs the
// three targets 0x10000 = 128*128*4 bytes apart, which is the byte extent of one.
//
// What is NOT pinned: which nibble is width and which is height. Every swizzled surface
// this title programs is square, so the assignment is invisible here (swizzleOffset treats
// the axes symmetrically when w==h) and pinning it would need a non-square case, which the
// image never presents. So this halts on any non-square swizzled surface rather than guess
// an orientation it cannot verify (the discipline of [[stored-registers-are-not-read-registers]]).
// swizzleGeom decodes a type-2 (swizzled) surface's square power-of-two extent from the
// format word's size nibbles (bits 16-19, 24-27). reason is non-empty on anything the
// derivation cannot pin — a caller must halt-and-name it rather than fill blind.
func swizzleGeom(format uint32) (w, h int, reason string) {
	if format>>20&0xF != 0 || format>>28&0xF != 0 {
		return 0, 0, "unexpected format high nibbles"
	}
	w = 1 << (format >> 16 & 0xF)
	h = 1 << (format >> 24 & 0xF)
	if w != h {
		return w, h, "non-square (width/height orientation unverified)"
	}
	return w, h, ""
}

func (g *pgraph) decodeSwizzleExtent(st *rasterState, format uint32) bool {
	m := g.m
	w, h, reason := swizzleGeom(format)
	if reason != "" {
		m.CPU.Halt("nv2a: draw %d swizzled surface %dx%d (0x208=%08X) — %s", g.Draws, w, h, format, reason)
		return false
	}
	st.swizzle = true
	st.swW, st.swH = w, h
	// The render region (the clip) must fit inside the allocated swizzle extent; if it
	// does not, the nibble reading is wrong and the fill would address out of bounds.
	clipW := int(g.Regs[kelvinSurfaceClipH>>2] >> 16)
	clipH := int(g.Regs[kelvinSurfaceClipV>>2] >> 16)
	if clipW > st.swW || clipH > st.swH {
		m.CPU.Halt("nv2a: draw %d swizzled extent %dx%d smaller than clip %dx%d (0x208=%08X)",
			g.Draws, st.swW, st.swH, clipW, clipH, format)
		return false
	}
	return true
}

// rasterStateDecode reads the registers a draw depends on. It returns false (after
// halting) on a surface layout the fill would silently corrupt.
func (g *pgraph) rasterStateDecode(st *rasterState) bool {
	m := g.m
	format := g.Regs[kelvinSurfaceFormat>>2]
	if cf := format & 0xF; cf != 0x8 && cf != 0x3 {
		// 0x8 = LE_A8R8G8B8; 0x3 = LE_R5G6B5 would need a 16-bit write path.
		m.CPU.Halt("nv2a: draw %d into unmodelled color surface format %X (0x208=%08X)", g.Draws, cf, format)
		return false
	}
	switch ty := format >> 8 & 0xF; ty {
	case 1: // PITCH — the main framebuffer and every scratch buffer this port has rendered
		st.swizzle = false
	case 2: // SWIZZLE — Morton-interleaved render target (OutRun's 128x128 render-to-texture)
		if !g.decodeSwizzleExtent(st, format) {
			return false
		}
	default:
		m.CPU.Halt("nv2a: draw %d into surface type %d (0x208=%08X) — unmodelled layout", g.Draws, ty, format)
		return false
	}
	st.aaX, st.aaY = surfaceAAScale(format)

	var ok bool
	st.colorPhys, ok = g.surfPhys(g.Regs[kelvinSurfaceColorOffset>>2])
	if !ok {
		m.CPU.Halt("nv2a: draw %d color surface %08X is not RAM", g.Draws, g.Regs[kelvinSurfaceColorOffset>>2])
		return false
	}
	st.colorPitch = g.Regs[kelvinSurfacePitch>>2] & 0xFFFF
	st.zetaPitch = g.Regs[kelvinSurfacePitch>>2] >> 16
	st.hasZeta = format>>4&0xF != 0
	if st.hasZeta {
		st.zetaPhys, ok = g.surfPhys(g.Regs[kelvinSurfaceZetaOffset>>2])
		if !ok {
			st.hasZeta = false
		}
		if zf := format >> 4 & 0xF; zf != 2 && st.hasZeta {
			// Only Z24S8 is modelled; Z16 would need a 16-bit depth path.
			m.CPU.Halt("nv2a: draw %d zeta format %d unmodelled (0x208=%08X)", g.Draws, zf, format)
			return false
		}
	}

	clipH := g.Regs[kelvinSurfaceClipH>>2]
	clipV := g.Regs[kelvinSurfaceClipV>>2]
	st.surfW = int(clipH>>16) * st.aaX
	st.surfH = int(clipV>>16) * st.aaY
	st.x0 = int(clipH&0xFFFF) * st.aaX
	st.y0 = int(clipV&0xFFFF) * st.aaY
	st.x1 = st.x0 + st.surfW
	st.y1 = st.y0 + st.surfH

	// The window clip (scissor). Type 0 is inclusive rectangles; only rect 0 is
	// modelled (the XDK sets exactly one).
	if g.Regs[kelvinWindowClipType>>2] == 0 {
		wh := g.Regs[kelvinWindowClipH>>2]
		wv := g.Regs[kelvinWindowClipV>>2]
		if wh != 0 || wv != 0 {
			x0, x1 := int(wh&0xFFFF)*st.aaX, (int(wh>>16)+1)*st.aaX
			y0, y1 := int(wv&0xFFFF)*st.aaY, (int(wv>>16)+1)*st.aaY
			if x0 > st.x0 {
				st.x0 = x0
			}
			if y0 > st.y0 {
				st.y0 = y0
			}
			if x1 < st.x1 {
				st.x1 = x1
			}
			if y1 < st.y1 {
				st.y1 = y1
			}
		}
	}

	st.depthTest = g.Regs[kelvinDepthTestEnable>>2]&1 != 0 && st.hasZeta
	st.depthWrite = g.Regs[kelvinDepthWriteMask>>2]&1 != 0 && st.hasZeta
	st.depthFunc = g.Regs[kelvinDepthFunc>>2]
	st.alphaTest = g.Regs[kelvinAlphaTestEnable>>2]&1 != 0
	st.alphaFunc = g.Regs[kelvinAlphaFunc>>2]
	st.alphaRef = float32(g.Regs[kelvinAlphaRef>>2]&0xFF) / 255
	st.blend = g.Regs[kelvinBlendEnable>>2]&1 != 0
	st.blendSrc = g.Regs[kelvinBlendSrcFactor>>2]
	st.blendDst = g.Regs[kelvinBlendDstFactor>>2]
	st.blendEq = g.Regs[kelvinBlendEquation>>2]
	bc := g.Regs[kelvinBlendColor>>2]
	st.blendConst = [4]float32{
		float32(bc>>24&0xFF) / 255, float32(bc>>16&0xFF) / 255,
		float32(bc>>8&0xFF) / 255, float32(bc&0xFF) / 255,
	}
	cm := g.Regs[kelvinColorMask>>2]
	st.colorMask = [4]bool{cm>>24&0xFF != 0, cm>>16&0xFF != 0, cm>>8&0xFF != 0, cm&0xFF != 0}
	st.cull = g.Regs[kelvinCullFaceEnable>>2]&1 != 0
	st.cullFace = g.Regs[kelvinCullFace>>2]
	st.frontCW = g.Regs[kelvinFrontFace>>2] == 0x900 // GL_CW; 0x901 GL_CCW
	st.flatShade = g.Regs[kelvinShadeMode>>2] == 0x1D00

	g.combDecode(&st.comb)
	return g.texStateDecode(st)
}

// surfPhys resolves a surface offset (a physical address on the base-0 color/zeta DMA
// contexts the XDK binds) to a RAM offset.
func (g *pgraph) surfPhys(addr uint32) (uint32, bool) {
	phys, mmio, ok := g.m.translate(addr)
	if !ok || mmio {
		return 0, false
	}
	return phys, true
}

// colorAddr / zetaAddr map a sample (px, py) to the byte address of its texel in the
// bound surface. A pitch surface is row-major (py*pitch + px*bpp); a swizzled surface is
// Morton-interleaved through swizzleOffset — the SAME interleave the texture unit reads
// (nv2a_texture.go), which is what makes a render-to-texture round-trip: the raster writes
// the target with this function and a later draw samples it back with decodeSwizzled, and
// the picture only survives if the two permutations match. The zeta buffer is private to
// the raster (written and read only by the depth test, never sampled), so its swizzle is
// self-consistent by construction whatever the absolute layout; it shares the color
// surface's type field and 128x128 extent because they are one surface.
func (st *rasterState) colorAddr(px, py int) uint32 {
	if st.swizzle {
		return st.colorPhys + uint32(swizzleOffset(px, py, st.swW, st.swH))*4
	}
	return st.colorPhys + uint32(py)*st.colorPitch + uint32(px)*4
}

func (st *rasterState) zetaAddr(px, py int) uint32 {
	if st.swizzle {
		return st.zetaPhys + uint32(swizzleOffset(px, py, st.swW, st.swH))*4
	}
	return st.zetaPhys + uint32(py)*st.zetaPitch + uint32(px)*4
}

// rasterTri is the per-triangle entry: decode state on the first triangle of a draw
// batch is deliberately NOT cached across draws (registers may change between
// BEGIN/END pairs), but within one assemble() the state is constant — the caller
// arranges that by keeping the decoded state on the pgraph for the batch.
func (g *pgraph) rasterTri(v0, v1, v2 *kelvinVtx) {
	if g.m.CPU.Halted {
		return
	}
	st := &g.rast
	if !g.rastValid {
		if !g.rasterStateDecode(st) {
			return
		}
		g.rastValid = true
	}

	// Positions arrive in SAMPLE space already: the screen-space epilogue the D3D
	// runtime appends to every program bakes the surface's anti-aliasing scale into
	// its viewport constants (confirmed live — the loading quads span 640x480 while
	// the logical clip is 320x240 with AA=2). Only the clip rects scale by AA.
	x0, y0 := float64(v0.pos[0]), float64(v0.pos[1])
	x1, y1 := float64(v1.pos[0]), float64(v1.pos[1])
	x2, y2 := float64(v2.pos[0]), float64(v2.pos[1])

	area := (x1-x0)*(y2-y0) - (y1-y0)*(x2-x0)
	if area == 0 {
		return
	}
	// With y down, positive signed area = clockwise winding.
	if st.cull {
		front := (area > 0) == st.frontCW
		switch st.cullFace {
		case 0x404: // GL_FRONT
			if front {
				return
			}
		case 0x408: // GL_FRONT_AND_BACK
			return
		default: // GL_BACK
			if !front {
				return
			}
		}
	}
	// Orient the edge functions so inside is positive regardless of winding.
	sign := 1.0
	if area < 0 {
		sign, area = -1.0, -area
	}

	minX := int(min3(x0, x1, x2))
	maxX := int(max3(x0, x1, x2)) + 1
	minY := int(min3(y0, y1, y2))
	maxY := int(max3(y0, y1, y2)) + 1
	if minX < st.x0 {
		minX = st.x0
	}
	if minY < st.y0 {
		minY = st.y0
	}
	if maxX > st.x1 {
		maxX = st.x1
	}
	if maxY > st.y1 {
		maxY = st.y1
	}
	if minX >= maxX || minY >= maxY {
		return
	}

	// Per-vertex 1/w for perspective-correct attribute interpolation (the program
	// preserves clip w in oPos.w; w==0 degrades to affine, which only a degenerate
	// vertex produces).
	iw0, iw1, iw2 := invW(v0.pos[3]), invW(v1.pos[3]), invW(v2.pos[3])

	edge := func(px, py, ex0, ey0, ex1, ey1 float64) float64 {
		return sign * ((ex1-ex0)*(py-ey0) - (ey1-ey0)*(px-ex0))
	}
	// Top-left fill rule: a fragment exactly on an edge belongs to the triangle only
	// if that edge is a top or left edge, so shared edges never double-blend.
	topLeft := func(ex0, ey0, ex1, ey1 float64) bool {
		dx, dy := sign*(ex1-ex0), sign*(ey1-ey0)
		return dy < 0 || (dy == 0 && dx > 0)
	}
	tl0 := topLeft(x1, y1, x2, y2)
	tl1 := topLeft(x2, y2, x0, y0)
	tl2 := topLeft(x0, y0, x1, y1)

	for py := minY; py < maxY; py++ {
		fy := float64(py) + 0.5
		for px := minX; px < maxX; px++ {
			fx := float64(px) + 0.5
			w0 := edge(fx, fy, x1, y1, x2, y2)
			w1 := edge(fx, fy, x2, y2, x0, y0)
			w2 := edge(fx, fy, x0, y0, x1, y1)
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			if (w0 == 0 && !tl0) || (w1 == 0 && !tl1) || (w2 == 0 && !tl2) {
				continue
			}
			b0, b1, b2 := float32(w0/area), float32(w1/area), float32(w2/area)
			g.shadePixel(st, px, py, b0, b1, b2, iw0, iw1, iw2, v0, v1, v2)
			if g.m.CPU.Halted {
				return
			}
		}
	}
}

// shadePixel runs one fragment: interpolate, sample, combine, test, blend, write.
func (g *pgraph) shadePixel(st *rasterState, px, py int, b0, b1, b2, iw0, iw1, iw2 float32, v0, v1, v2 *kelvinVtx) {
	m := g.m

	// Depth first (the hardware kills depth-failed fragments before the combiners).
	var zAddr uint32
	var zOld uint32
	if st.depthTest || st.depthWrite {
		zAddr = st.zetaAddr(px, py)
		if zAddr+4 > uint32(len(m.RAM)) {
			return
		}
	}
	z := b0*v0.pos[2] + b1*v1.pos[2] + b2*v2.pos[2]
	zi := uint32(0)
	if z > 0 {
		if z >= 0xFFFFFF {
			zi = 0xFFFFFF
		} else {
			zi = uint32(z)
		}
	}
	if st.depthTest {
		stored := uint32(m.RAM[zAddr]) | uint32(m.RAM[zAddr+1])<<8 | uint32(m.RAM[zAddr+2])<<16 | uint32(m.RAM[zAddr+3])<<24
		zOld = stored >> 8
		if !depthPass(st.depthFunc, zi, zOld) {
			if m.OnPixel != nil {
				m.OnPixel(uint32(px), uint32(py), PixelEvent{ZReject: true})
			}
			return
		}
	}

	// Perspective-correct attribute interpolation.
	iw := b0*iw0 + b1*iw1 + b2*iw2
	if iw == 0 {
		iw = 1
	}
	pc4 := func(a0, a1, a2 *[4]float32) [4]float32 {
		var out [4]float32
		for i := 0; i < 4; i++ {
			out[i] = (b0*a0[i]*iw0 + b1*a1[i]*iw1 + b2*a2[i]*iw2) / iw
		}
		return out
	}

	var in combInput
	if st.flatShade {
		in.col0, in.col1 = v2.d0, v2.d1 // flat shading takes the provoking (last) vertex
	} else {
		in.col0 = pc4(&v0.d0, &v1.d0, &v2.d0)
		in.col1 = pc4(&v0.d1, &v1.d1, &v2.d1)
	}
	for u := 0; u < 4; u++ {
		if !st.texEnable[u] || st.texImg[u] == nil {
			in.tex[u] = [4]float32{1, 1, 1, 1}
			continue
		}
		uvw := pc4(&v0.uv[u], &v1.uv[u], &v2.uv[u])
		s, t := uvw[0], uvw[1]
		if st.texStage[u] == 2 || (st.texStage[u] == 1 && uvw[3] != 0 && uvw[3] != 1) {
			// projective: divide by q
			s, t = s/uvw[3], t/uvw[3]
		}
		in.tex[u] = g.texSample(st, u, s, t)
	}

	col := g.combine(&st.comb, &in)

	if st.alphaTest && !alphaPass(st.alphaFunc, col[3], st.alphaRef) {
		if m.OnPixel != nil {
			m.OnPixel(uint32(px), uint32(py), PixelEvent{
				AlphaReject: true,
				R:           u8(col[0]), G: u8(col[1]), B: u8(col[2]), A: u8(col[3]),
			})
		}
		return
	}

	cAddr := st.colorAddr(px, py)
	if cAddr+4 > uint32(len(m.RAM)) {
		return
	}
	if st.blend {
		db := float32(m.RAM[cAddr]) / 255
		dg := float32(m.RAM[cAddr+1]) / 255
		dr := float32(m.RAM[cAddr+2]) / 255
		da := float32(m.RAM[cAddr+3]) / 255
		dst := [4]float32{dr, dg, db, da}
		col = blendPixel(st, col, dst)
	}

	// A8R8G8B8 little-endian: B, G, R, A.
	if st.colorMask[3] {
		m.RAM[cAddr] = u8(col[2])
	}
	if st.colorMask[2] {
		m.RAM[cAddr+1] = u8(col[1])
	}
	if st.colorMask[1] {
		m.RAM[cAddr+2] = u8(col[0])
	}
	if st.colorMask[0] {
		m.RAM[cAddr+3] = u8(col[3])
	}
	if st.depthWrite {
		stored := uint32(m.RAM[zAddr]) | uint32(m.RAM[zAddr+1])<<8 | uint32(m.RAM[zAddr+2])<<16 | uint32(m.RAM[zAddr+3])<<24
		stored = zi<<8 | stored&0xFF // keep stencil (unmodelled) intact
		m.RAM[zAddr] = byte(stored)
		m.RAM[zAddr+1] = byte(stored >> 8)
		m.RAM[zAddr+2] = byte(stored >> 16)
		m.RAM[zAddr+3] = byte(stored >> 24)
	}
	if m.OnPixel != nil {
		m.OnPixel(uint32(px), uint32(py), PixelEvent{
			Drawn: true,
			R:     u8(col[0]), G: u8(col[1]), B: u8(col[2]), A: u8(col[3]),
		})
	}
}

func depthPass(fn, z, old uint32) bool {
	switch fn {
	case 0x200: // NEVER
		return false
	case 0x201: // LESS
		return z < old
	case 0x202: // EQUAL
		return z == old
	case 0x203: // LEQUAL
		return z <= old
	case 0x204: // GREATER
		return z > old
	case 0x205: // NOTEQUAL
		return z != old
	case 0x206: // GEQUAL
		return z >= old
	default: // ALWAYS (0x207)
		return true
	}
}

func alphaPass(fn uint32, a, ref float32) bool {
	switch fn {
	case 0x200:
		return false
	case 0x201:
		return a < ref
	case 0x202:
		return a == ref
	case 0x203:
		return a <= ref
	case 0x204:
		return a > ref
	case 0x205:
		return a != ref
	case 0x206:
		return a >= ref
	default:
		return true
	}
}

// blendPixel applies the latched blend equation and factors (GL enums, as the
// hardware encodes them).
func blendPixel(st *rasterState, src, dst [4]float32) [4]float32 {
	factor := func(f uint32, other bool) [4]float32 {
		switch f {
		case 0: // ZERO
			return [4]float32{}
		case 1: // ONE
			return [4]float32{1, 1, 1, 1}
		case 0x300: // SRC_COLOR
			return src
		case 0x301: // ONE_MINUS_SRC_COLOR
			return [4]float32{1 - src[0], 1 - src[1], 1 - src[2], 1 - src[3]}
		case 0x302: // SRC_ALPHA
			return [4]float32{src[3], src[3], src[3], src[3]}
		case 0x303: // ONE_MINUS_SRC_ALPHA
			return [4]float32{1 - src[3], 1 - src[3], 1 - src[3], 1 - src[3]}
		case 0x304: // DST_ALPHA
			return [4]float32{dst[3], dst[3], dst[3], dst[3]}
		case 0x305: // ONE_MINUS_DST_ALPHA
			return [4]float32{1 - dst[3], 1 - dst[3], 1 - dst[3], 1 - dst[3]}
		case 0x306: // DST_COLOR
			return dst
		case 0x307: // ONE_MINUS_DST_COLOR
			return [4]float32{1 - dst[0], 1 - dst[1], 1 - dst[2], 1 - dst[3]}
		case 0x308: // SRC_ALPHA_SATURATE
			f := minf32(src[3], 1-dst[3])
			return [4]float32{f, f, f, 1}
		case 0x8001: // CONSTANT_COLOR
			return st.blendConst
		case 0x8002:
			return [4]float32{1 - st.blendConst[0], 1 - st.blendConst[1], 1 - st.blendConst[2], 1 - st.blendConst[3]}
		case 0x8003: // CONSTANT_ALPHA
			a := st.blendConst[3]
			return [4]float32{a, a, a, a}
		case 0x8004:
			a := 1 - st.blendConst[3]
			return [4]float32{a, a, a, a}
		}
		return [4]float32{1, 1, 1, 1}
	}
	sf, df := factor(st.blendSrc, false), factor(st.blendDst, true)
	var out [4]float32
	switch st.blendEq {
	case 0x800A: // SUBTRACT
		for i := range out {
			out[i] = src[i]*sf[i] - dst[i]*df[i]
		}
	case 0x800B: // REVERSE_SUBTRACT
		for i := range out {
			out[i] = dst[i]*df[i] - src[i]*sf[i]
		}
	case 0x8007: // MIN
		for i := range out {
			out[i] = minf32(src[i], dst[i])
		}
	case 0x8008: // MAX
		for i := range out {
			out[i] = maxf32(src[i], dst[i])
		}
	default: // FUNC_ADD (0x8006)
		for i := range out {
			out[i] = src[i]*sf[i] + dst[i]*df[i]
		}
	}
	return out
}

func u8(f float32) uint8 {
	if f <= 0 {
		return 0
	}
	if f >= 1 {
		return 255
	}
	return uint8(f*255 + 0.5)
}

func invW(w float32) float32 {
	if w == 0 {
		return 1
	}
	return 1 / w
}

func min3(a, b, c float64) float64 {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}
func max3(a, b, c float64) float64 {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}
