package dsmachine

// The 3D engine's back end: the rasteriser. gpu3d_geom.go produced a list of clipped
// polygons in clip space; this file turns them into the 256x192 image the 2D engine
// composites as engine A's BG0.
//
// The DS is a *scanline* rasteriser, not a half-space one. It walks a polygon's left
// and right edges down the screen and fills the span between them, which is why the
// hardware can get away with 2048 polygons and no depth prepass, and why its filling
// rules are edge rules rather than coverage rules. Modelling it as a scanline filler
// is not nostalgia: the span is where the perspective divide happens, and a span-based
// divide is what stops textures swimming.
//
// Three things in here are the classic bugs, and each is called out again where it
// happens:
//
//   - The Y FLIP. The DS's viewport has its origin at the bottom-left; the screen's is
//     at the top-left. Get this wrong and the whole world renders upside down, and it
//     will look plausible enough in a title screen that you will not notice for a day.
//   - The 4-COLOUR PALETTE SHIFT. PLTT_BASE is scaled by 16 for every texture format
//     except format 2, where it is scaled by 8. Miss it and only the 2-bit textures are
//     wrong, in colour but not in shape — which reads as a palette bug somewhere else.
//   - W VERSUS Z BUFFERING. SWAP_BUFFERS bit 1 chooses which coordinate is depth-
//     tested. They are not interchangeable: a scene authored for W and tested in Z
//     z-fights at distance, and the DS's own libraries switch between them per frame.
//
// Not implemented, and noted through m.note rather than approximated: fog, edge
// marking, anti-aliasing, shadow polygons (POLYGON_ATTR mode 3), and the clear-image
// rear-plane bitmap. Drawing a shadow polygon as an ordinary one would paint a black
// quad across the scene, which is worse than not drawing it.

import "time"

const (
	rastW = 256
	rastH = 192
)

// rfrag is one pixel of the 3D colour buffer, kept in the hardware's own depth: six
// bits per channel and five bits of alpha. Converting to RGBA8888 only at the end
// keeps repeated translucent blends from quantising twice.
//
// An alpha of zero means *nothing was drawn here*: the 2D layers below show through.
// That is also what the clear colour's alpha can say, and it is the reason the frame
// handed to the 2D engine carries a meaningful alpha byte at all.
type rfrag struct {
	r, g, b uint8 // 0..63
	a       uint8 // 0..31
}

// raster owns the 3D engine's colour and depth buffers and the RGBA frame it hands
// to the 2D engine. The buffers are the rasteriser's alone: the DS's 3D engine
// composites nothing itself, it just produces a layer.
type raster struct {
	col   [rastW * rastH]rfrag
	depth [rastW * rastH]uint32

	frame []uint32 // RGBA8888, 256x192, alpha 0 where nothing was drawn
}

func (r *raster) reset() {
	for i := range r.col {
		r.col[i] = rfrag{}
		r.depth[i] = 0xFFFFFF
	}
	if r.frame == nil {
		r.frame = make([]uint32, rastW*rastH)
	}
	for i := range r.frame {
		r.frame[i] = 0
	}
}

// rvert is a vertex ready to interpolate: screen position, a depth value, and every
// varying already divided by w so that a linear interpolation in screen space is the
// perspective-correct one. The reciprocal iw comes along for the ride so a pixel can
// undo the division.
type rvert struct {
	x, y  float64 // screen space, pixels, y already flipped
	depth float64 // 0..0xFFFFFF; linear in screen space (Z mode) or unused (W mode)
	iw    float64 // 1/w
	w     float64 // clip w, in 20.12 units; the depth value in W mode

	// Varyings, premultiplied by iw.
	r, g, b float64 // vertex colour, 0..63 before the division
	s, t    float64 // texture coordinates, 12.4 before the division
}

func lerpRV(a, b rvert, u float64) rvert {
	l := func(x, y float64) float64 { return x + (y-x)*u }
	return rvert{
		x: l(a.x, b.x), y: l(a.y, b.y), depth: l(a.depth, b.depth),
		iw: l(a.iw, b.iw), w: l(a.w, b.w),
		r: l(a.r, b.r), g: l(a.g, b.g), b: l(a.b, b.b),
		s: l(a.s, b.s), t: l(a.t, b.t),
	}
}

// render rasterises the frame's polygons and publishes the result to the 2D engine.
func (g *gpu3d) render(m *Machine) {
	if m.prof.on {
		t0 := time.Now()
		defer func() { m.prof.raster += time.Since(t0) }()
	}
	m.prof.polys += len(g.geom.polys)

	disp3d := m.ARM9.io[regDISP3DCNT]

	g.clear(m, disp3d)

	for i := range g.geom.polys {
		if m.OnPoly != nil {
			m.OnPoly(g.geom.polys[i].cmd)
		}
		g.drawPoly(m, &g.geom.polys[i], disp3d)
	}

	// The effects we do not do. They are noted, not faked: a run's Log is the honest
	// list of what the image is missing.
	if disp3d&(1<<7) != 0 {
		m.note("3D: fog enabled (DISP3DCNT bit 7) — not implemented, the image is unfogged")
	}
	if disp3d&(1<<5) != 0 {
		m.note("3D: edge marking enabled (DISP3DCNT bit 5) — not implemented")
	}
	if disp3d&(1<<4) != 0 {
		m.note("3D: anti-aliasing enabled (DISP3DCNT bit 4) — not implemented, edges are hard")
	}

	g.rast.publish()
	if m.gpu2d != nil {
		if m.gpu2d.threeD == nil {
			m.gpu2d.threeD = make([]uint32, rastW*rastH)
		}
		copy(m.gpu2d.threeD, g.rast.frame)
	}
}

// publish converts the 6/6/6/5 colour buffer into the RGBA8888 the 2D engine blends.
// Alpha zero survives as alpha zero, and that is load-bearing: it is how the 2D engine
// knows a pixel of BG0 is absent rather than black.
func (r *raster) publish() {
	if r.frame == nil {
		r.frame = make([]uint32, rastW*rastH)
	}
	for i, c := range r.col {
		if c.a == 0 {
			r.frame[i] = 0
			continue
		}
		r.frame[i] = uint32(c6to8(c.r)) | uint32(c6to8(c.g))<<8 | uint32(c6to8(c.b))<<16 |
			uint32(a5to8(c.a))<<24
	}
}

func c6to8(v uint8) uint8 { return uint8((int(v)*255 + 31) / 63) }
func a5to8(v uint8) uint8 { return uint8((int(v)*255 + 15) / 31) }

// clear fills the colour and depth buffers from CLEAR_COLOR and CLEAR_DEPTH.
//
// CLEAR_DEPTH is a 15-bit value and the depth buffer is 24-bit; the hardware expands
// it as v*0x200 + 0x1FF, which puts the usual clear value (0x7FFF) exactly at the far
// plane, 0xFFFFFF. Expanding it by a plain shift instead leaves the far plane one
// short and the backmost polygons of a skybox fail their depth test.
func (g *gpu3d) clear(m *Machine, disp3d uint32) {
	if disp3d&(1<<14) != 0 {
		m.note("3D: clear-image / rear-plane bitmap enabled (DISP3DCNT bit 14) — not implemented, clearing to CLEAR_COLOR instead")
	}

	cc := g.regs[regCLEARCOLOR]
	r5, g5, b5 := int32(cc&0x1F), int32((cc>>5)&0x1F), int32((cc>>10)&0x1F)
	ca := uint8((cc >> 16) & 0x1F)
	c := rfrag{r: uint8(r5 * 2), g: uint8(g5 * 2), b: uint8(b5 * 2), a: ca}

	d := (g.regs[regCLEARDEPTH] & 0x7FFF) * 0x200
	d += 0x1FF

	for i := range g.rast.col {
		g.rast.col[i] = c
		g.rast.depth[i] = d
	}
}

// drawPoly rasterises one polygon.
func (g *gpu3d) drawPoly(m *Machine, p *gxPolygon, disp3d uint32) {
	if len(p.verts) < 3 {
		return
	}

	mode := (p.attr >> 4) & 3
	if mode == 3 {
		// Shadow polygons need the stencil-like shadow bit of the depth buffer and a
		// two-pass protocol against the shadow *volume*'s polygon ID. Drawing them as
		// ordinary polygons would smear an opaque quad over the scene, so we draw
		// nothing and say so.
		m.note("3D: shadow polygon (POLYGON_ATTR mode 3) skipped — shadow volumes not implemented")
		return
	}

	// The viewport. VIEWPORT gives an inclusive rectangle in the DS's own coordinates,
	// whose Y origin is at the BOTTOM of the screen.
	vx1, vy1 := int(g.geom.viewX1), int(g.geom.viewY1)
	vx2, vy2 := int(g.geom.viewX2), int(g.geom.viewY2)
	vw := vx2 - vx1 + 1
	vh := vy2 - vy1 + 1
	if vw <= 0 || vh <= 0 {
		return
	}

	verts := make([]rvert, 0, len(p.verts))
	for _, v := range p.verts {
		w := float64(v.w)
		if v.w == 0 {
			// A vertex exactly on the eye plane has no projection. The clipper should have
			// removed it (w >= |x|,|y|,|z| >= 0 and the near plane is w+z >= 0), but a
			// degenerate matrix can still produce one, and a divide by zero here would
			// poison the whole span.
			return
		}

		// The perspective divide, in the DS's form: x is mapped from [-w, +w] onto the
		// viewport. Written as (x + w) * width / (2w) so that it stays honest for negative
		// w — which cannot reach here, but the shape of the expression is the hardware's.
		sx := (float64(v.x)+w)*float64(vw)/(2*w) + float64(vx1)
		sy := (float64(v.y)+w)*float64(vh)/(2*w) + float64(vy1)

		// THE Y FLIP. sy is measured up from the bottom of the screen; the framebuffer is
		// measured down from the top. This single subtraction is the difference between a
		// world and its reflection, and nothing else in the pipeline will complain.
		sy = float64(rastH-1) - sy

		iw := 1.0 / w

		// Z-buffer depth: z/w maps [-1,+1] onto the 24-bit range. Written the hardware's
		// way — ((z << 14) / w + 0x3FFF) << 9 — so the rounding is the rounding a game's
		// depth-biased decals were tuned against.
		var depth float64
		if !p.wbuffer {
			d := (float64(v.z)*16384.0/w + 16383.0) * 512.0
			if d < 0 {
				d = 0
			}
			if d > 0xFFFFFF {
				d = 0xFFFFFF
			}
			depth = d
		}

		verts = append(verts, rvert{
			x: sx, y: sy, depth: depth, iw: iw, w: w,
			r: float64(v.r) * iw, g: float64(v.g) * iw, b: float64(v.b) * iw,
			s: float64(v.s) * iw, t: float64(v.t) * iw,
		})
	}

	// CULLING, from POLYGON_ATTR bits 6 and 7: bit 6 draws back faces, bit 7 draws
	// front faces. Both clear means the polygon is submitted only to advance a strip,
	// and must not appear.
	drawBack := p.attr&(1<<6) != 0
	drawFront := p.attr&(1<<7) != 0
	if !drawBack && !drawFront {
		return
	}
	// The winding is computed in SCREEN space, after the flip, by the shoelace sum over
	// the whole polygon (not just its first three vertices — a clipped polygon's first
	// three can be collinear). With y pointing down, a polygon that was counter-
	// clockwise in the DS's y-up view comes out with a negative area, and that is the
	// front face. TRAP: if models render inside out — you see the far wall of a room and
	// not the near one — this sign is the first thing to flip.
	var area float64
	for i := range verts {
		j := (i + 1) % len(verts)
		area += verts[i].x*verts[j].y - verts[j].x*verts[i].y
	}
	front := area < 0
	if area == 0 {
		return // degenerate: zero screen area, nothing to fill
	}
	if front && !drawFront {
		return
	}
	if !front && !drawBack {
		return
	}

	// The scissor is the viewport, in screen (flipped) coordinates.
	clipX0, clipX1 := vx1, vx2
	clipY0 := rastH - 1 - vy2
	clipY1 := rastH - 1 - vy1
	if clipX0 < 0 {
		clipX0 = 0
	}
	if clipX1 > rastW-1 {
		clipX1 = rastW - 1
	}
	if clipY0 < 0 {
		clipY0 = 0
	}
	if clipY1 > rastH-1 {
		clipY1 = rastH - 1
	}

	minY, maxY := verts[0].y, verts[0].y
	for _, v := range verts[1:] {
		if v.y < minY {
			minY = v.y
		}
		if v.y > maxY {
			maxY = v.y
		}
	}
	y0 := int(rastCeil(minY - 0.5))
	y1 := int(rastCeil(maxY-0.5)) - 1
	if y0 < clipY0 {
		y0 = clipY0
	}
	if y1 > clipY1 {
		y1 = clipY1
	}

	st := g.polyState(m, p, disp3d)

	for y := y0; y <= y1; y++ {
		yc := float64(y) + 0.5

		// The span. The polygon is convex (the clipper guarantees it), so a scanline
		// crosses its boundary exactly twice; collecting every crossing and taking the
		// leftmost and rightmost is the same walk as tracking two edge chains, and it does
		// not have to special-case the turn at the top and bottom vertices.
		var have bool
		var left, right rvert
		for i := range verts {
			a, b := verts[i], verts[(i+1)%len(verts)]
			if a.y == b.y {
				continue // a horizontal edge contributes no crossing; its endpoints do
			}
			lo, hi := a, b
			if lo.y > hi.y {
				lo, hi = hi, lo
			}
			// Half-open in y: the scanline belongs to the edge that starts on it, never to
			// the one that ends on it. Two polygons sharing an edge then fill it once.
			if yc < lo.y || yc >= hi.y {
				continue
			}
			u := (yc - lo.y) / (hi.y - lo.y)
			s := lerpRV(lo, hi, u)
			if !have {
				left, right, have = s, s, true
			} else if s.x < left.x {
				left = s
			} else if s.x > right.x {
				right = s
			}
		}
		if !have {
			continue
		}

		x0 := int(rastCeil(left.x - 0.5))
		x1 := int(rastCeil(right.x-0.5)) - 1
		if x0 == x1+1 {
			// A span thinner than a pixel centre still exists on hardware: the DS always
			// draws its left and right edge pixels. Keep one pixel rather than dropping the
			// sliver, or thin geometry (a railing, a distant pole) flickers out entirely.
			x1 = x0
		}
		if x0 < clipX0 {
			x0 = clipX0
		}
		if x1 > clipX1 {
			x1 = clipX1
		}
		if x0 > x1 {
			continue
		}

		dx := right.x - left.x
		for x := x0; x <= x1; x++ {
			if st.wireframe && x != x0 && x != x1 && y != y0 && y != y1 {
				continue // alpha 0 is wireframe: only the outline is drawn (see polyState)
			}
			var u float64
			if dx > 0 {
				u = (float64(x) + 0.5 - left.x) / dx
				if u < 0 {
					u = 0
				}
				if u > 1 {
					u = 1
				}
			}
			f := lerpRV(left, right, u)
			g.shade(m, &st, p, f, y*rastW+x)
		}
	}
}

// polyStateHolds everything about a polygon that does not vary per pixel: its texture,
// its blending mode, its alpha, and the depth rules it plays by.
type polyState struct {
	tex     texState
	hasTex  bool
	mode    uint32 // POLYGON_ATTR bits 4-5
	alpha   uint8  // 0..31
	trans   bool   // 1..30: blended against the colour buffer
	depthEq bool   // POLYGON_ATTR bit 14: test "equal" rather than "less"
	depthWr bool   // write depth (always for opaque; bit 11 for translucent)
	wbuf    bool

	blend   bool // DISP3DCNT bit 3
	toonHi  bool // DISP3DCNT bit 1: highlight rather than toon
	alphaTe bool // DISP3DCNT bit 2
	alphaRf uint8

	toon [32][3]uint8 // the toon table, unpacked to 6-bit channels

	wireframe bool
}

func (g *gpu3d) polyState(m *Machine, p *gxPolygon, disp3d uint32) polyState {
	var st polyState
	st.mode = (p.attr >> 4) & 3
	st.alpha = uint8((p.attr >> 16) & 0x1F)
	st.depthEq = p.attr&(1<<14) != 0
	st.wbuf = p.wbuffer
	st.blend = disp3d&(1<<3) != 0
	st.toonHi = disp3d&(1<<1) != 0
	st.alphaTe = disp3d&(1<<2) != 0
	st.alphaRf = uint8(g.regs[0x04000340] & 0x1F)

	// Alpha 0 does not mean invisible: it selects the wireframe mode, in which only the
	// polygon's outline is drawn, at full opacity. We approximate the outline with the
	// first and last pixel of every span and the polygon's first and last scanline,
	// which is the right shape but not the hardware's exact edge set.
	if st.alpha == 0 {
		st.wireframe = true
		st.alpha = 31
		m.note("3D: wireframe polygons (POLYGON_ATTR alpha 0) approximated by span-edge pixels")
	}
	st.trans = st.alpha < 31
	// Translucent polygons do not write depth unless POLYGON_ATTR bit 11 asks them to.
	// This is what lets a game draw a window and still see the wall behind it through a
	// second translucent pane; forcing the write turns overlapping glass opaque-black.
	st.depthWr = !st.trans || p.attr&(1<<11) != 0

	if disp3d&1 != 0 { // DISP3DCNT bit 0: texture mapping enabled at all
		st.tex, st.hasTex = texStateOf(p)
	}

	if st.mode == 2 {
		// The toon/highlight table: 32 BGR555 entries at 0x04000380, two per word.
		for i := 0; i < 32; i++ {
			w := g.regs[0x04000380+uint32(i/2)*4]
			c := uint16(w >> (16 * uint(i&1)))
			st.toon[i] = [3]uint8{
				uint8((c & 0x1F) * 2), uint8(((c >> 5) & 0x1F) * 2), uint8(((c >> 10) & 0x1F) * 2),
			}
		}
	}
	return st
}

// shade computes and writes one pixel: depth test, texture fetch, the polygon's
// blending mode, then the alpha blend into the colour buffer.
func (g *gpu3d) shade(m *Machine, st *polyState, p *gxPolygon, f rvert, idx int) {
	// The frame debugger's per-pixel provenance. A fragment is reported whether it was
	// kept or killed, because "the rasteriser produced this pixel and then threw it
	// away" is usually the answer to why a pixel is not the colour it should be — and it
	// is an answer no colour buffer can give, having never held it.
	reject := func(z, a bool) {
		if m.OnPixel != nil {
			m.OnPixel(idx%rastW, idx/rastW, PixelEvent{ZReject: z, AlphaReject: a})
		}
	}
	// Undo the perspective division. Everything interpolated across the span was
	// premultiplied by 1/w, so the varyings are linear in screen space and this one
	// divide recovers them — this, and not the interpolation, is what keeps a texture
	// pinned to a surface instead of swimming across it as the camera turns.
	if f.iw == 0 {
		return
	}
	w := 1.0 / f.iw

	var depth uint32
	if st.wbuf {
		// W-buffer mode. The depth *is* w, and w is a 20.12 clip-space distance; the
		// hardware normalises it into 24 bits with a per-frame shift derived from the
		// projection. We take w's integer part, which is the same ordering and the same
		// resolution for the near-field magnitudes a DS scene actually uses.
		// UNCERTAIN: the hardware's exact normalisation shift is not modelled, so absolute
		// depth values in W mode are ours and not the DS's. The *comparisons* are correct,
		// which is all the depth test consumes.
		d := w
		if d < 0 {
			d = 0
		}
		if d > 0xFFFFFF {
			d = 0xFFFFFF
		}
		depth = uint32(d)
	} else {
		d := f.depth
		if d < 0 {
			d = 0
		}
		if d > 0xFFFFFF {
			d = 0xFFFFFF
		}
		depth = uint32(d)
	}

	old := g.rast.depth[idx]
	if st.depthEq {
		// "Equal" mode, used for decals: the polygon is drawn only where it lands on the
		// surface it decorates. Exact equality would fail on almost every pixel once the
		// two polygons were interpolated by different edge slopes, so the hardware allows
		// a tolerance; 0x200 is the documented margin.
		diff := int64(depth) - int64(old)
		if diff < 0 {
			diff = -diff
		}
		if diff > 0x200 {
			reject(true, false)
			return
		}
	} else if depth >= old {
		reject(true, false)
		return
	}

	// The fragment's own colour, before the texture: the interpolated vertex colour, in
	// the DS's six bits per channel.
	vr := rastClamp63(f.r * w)
	vg := rastClamp63(f.g * w)
	vb := rastClamp63(f.b * w)

	cr, cg, cb := vr, vg, vb
	ca := st.alpha

	if st.hasTex {
		s := int(f.s * w) // 12.4 fixed point: sixteenths of a texel
		t := int(f.t * w)
		tr, tg, tb, ta, ok := g.sampleTex(m, &st.tex, s>>4, t>>4)
		if !ok {
			reject(false, true) // a transparent texel is not drawn at all, and writes no depth
			return
		}
		switch st.mode {
		case 0: // modulation: texture * vertex colour, and the alphas multiply too
			cr = uint8((int(tr)*int(vr) + 31) / 63)
			cg = uint8((int(tg)*int(vg) + 31) / 63)
			cb = uint8((int(tb)*int(vb) + 31) / 63)
			ca = uint8((int(ta)*int(st.alpha) + 15) / 31)
		case 1: // decal: the texture's alpha chooses between texture and vertex colour,
			// rather than scaling it. The polygon's own alpha is kept.
			cr = uint8((int(tr)*int(ta) + int(vr)*(31-int(ta))) / 31)
			cg = uint8((int(tg)*int(ta) + int(vg)*(31-int(ta))) / 31)
			cb = uint8((int(tb)*int(ta) + int(vb)*(31-int(ta))) / 31)
			ca = st.alpha
		case 2: // toon / highlight
			cr, cg, cb, ca = st.toonShade(tr, tg, tb, ta, vr)
		}
	} else if st.mode == 2 {
		cr, cg, cb, ca = st.toonShade(63, 63, 63, 31, vr)
	}

	// The alpha test. A zero-alpha fragment is always discarded; with DISP3DCNT bit 2
	// set the game supplies its own threshold in ALPHA_TEST_REF.
	if ca == 0 {
		reject(false, true)
		return
	}
	if st.alphaTe && ca <= st.alphaRf {
		reject(false, true)
		return
	}

	dst := g.rast.col[idx]
	out := rfrag{r: cr, g: cg, b: cb, a: ca}

	if st.trans && st.blend && dst.a != 0 {
		// Translucent: blend against what the 3D engine has already drawn. Note the
		// destination here is the 3D colour buffer, not the 2D layers — the DS composites
		// the finished 3D layer with the backgrounds afterwards, and a translucent polygon
		// over empty 3D space stays translucent all the way to the 2D blender.
		a := int(ca)
		mix := func(s, d uint8) uint8 {
			return uint8((int(s)*a + int(d)*(31-a)) / 31)
		}
		out.r = mix(cr, dst.r)
		out.g = mix(cg, dst.g)
		out.b = mix(cb, dst.b)
		if int(dst.a) > a {
			out.a = dst.a // the blend keeps the greater alpha, as the hardware does
		}
	}
	// With DISP3DCNT bit 3 clear the blender is off and a translucent polygon is simply
	// written opaque — that is the hardware's behaviour, not a shortcut.

	m.prof.frags++
	g.rast.col[idx] = out
	if st.depthWr {
		g.rast.depth[idx] = depth
	}
	if m.OnPixel != nil {
		m.OnPixel(idx%rastW, idx/rastW, PixelEvent{
			Drawn: true, R: out.r, G: out.g, B: out.b, A: out.a,
		})
	}
}

// toonShade applies the toon/highlight table. The table is indexed by the vertex
// colour's RED channel alone — the DS expects a game to have written a grey vertex
// colour whose intensity is the lighting result, and the table turns that intensity
// into a colour ramp. In toon mode the table's colour replaces the vertex colour; in
// highlight mode (DISP3DCNT bit 1) it is added on top of the modulated result, which
// is how Mario's cel-shaded rim appears.
func (st *polyState) toonShade(tr, tg, tb, ta, vr uint8) (uint8, uint8, uint8, uint8) {
	tc := st.toon[vr>>1] // 6-bit red down to the table's 32 entries
	a := uint8((int(ta)*int(st.alpha) + 15) / 31)
	if !st.toonHi {
		return uint8((int(tr)*int(tc[0]) + 31) / 63),
			uint8((int(tg)*int(tc[1]) + 31) / 63),
			uint8((int(tb)*int(tc[2]) + 31) / 63), a
	}
	add := func(t, v, h uint8) uint8 {
		c := (int(t)*int(v)+31)/63 + int(h)
		if c > 63 {
			c = 63
		}
		return uint8(c)
	}
	return add(tr, vr, tc[0]), add(tg, vr, tc[1]), add(tb, vr, tc[2]), a
}

func rastClamp63(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 63 {
		return 63
	}
	return uint8(v + 0.5)
}

func rastCeil(v float64) float64 {
	i := float64(int64(v))
	if v > i {
		i++
	}
	return i
}

// --- textures ---------------------------------------------------------------
//
// TEXIMAGE_PARAM packs everything about a texture into one word: where it lives in the
// texture space, how big it is, how it repeats, and which of the seven formats it is
// in. PLTT_BASE says where its palette lives — and the scale factor there is a format-
// dependent trap (see texStateOf).

// texState is a decoded TEXIMAGE_PARAM plus its palette base.
type texState struct {
	base    uint32 // byte offset into spTex
	pal     uint32 // byte offset into spTexPal
	sizeS   int
	sizeT   int
	format  uint32
	repeatS bool
	repeatT bool
	flipS   bool
	flipT   bool
	color0  bool // colour index 0 is transparent
}

func texStateOf(p *gxPolygon) (texState, bool) {
	var st texState
	st.format = (p.texParam >> 26) & 7
	if st.format == 0 {
		return st, false // no texture: the vertex colour stands alone
	}
	// The VRAM offset field counts 8-byte units, which is why a 16-bit field can address
	// the whole 512 KiB texture space.
	st.base = (p.texParam & 0xFFFF) << 3
	st.repeatS = p.texParam&(1<<16) != 0
	st.repeatT = p.texParam&(1<<17) != 0
	st.flipS = p.texParam&(1<<18) != 0
	st.flipT = p.texParam&(1<<19) != 0
	st.sizeS = 8 << ((p.texParam >> 20) & 7)
	st.sizeT = 8 << ((p.texParam >> 23) & 7)
	st.color0 = p.texParam&(1<<29) != 0

	// THE PALETTE SHIFT. PLTT_BASE's 13-bit field is scaled by 16 for every format
	// except the 4-colour one, where it is scaled by 8 — the 4-colour palettes are only
	// eight bytes long, so the hardware gives them a finer base granularity. Use 16
	// everywhere and every 2-bit texture picks up the wrong palette (right shapes, wrong
	// colours), which is a maddening bug to chase from the image alone.
	base := (p.pltt & 0x1FFF)
	if st.format == 2 {
		st.pal = base << 3
	} else {
		st.pal = base << 4
	}
	return st, true
}

// wrapTex applies the repeat/flip/clamp rules to one texture axis. Sizes are always
// powers of two, so the repeat is a mask and the flip is the bit just above it.
func wrapTex(v, size int, repeat, flip bool) int {
	if !repeat {
		if v < 0 {
			return 0
		}
		if v >= size {
			return size - 1
		}
		return v
	}
	mask := size - 1
	if flip && v&size != 0 {
		return mask - (v & mask) // the odd tile is mirrored
	}
	return v & mask
}

// rastBGR555 turns a palette halfword into the engine's six-bit channels. The five-bit
// value is doubled, exactly as unpackRGB5 does for vertex colours — the two must agree
// or a modulated texture will not match a flat-coloured one.
func rastBGR555(c uint16) (uint8, uint8, uint8) {
	return uint8((c & 0x1F) * 2), uint8(((c >> 5) & 0x1F) * 2), uint8(((c >> 10) & 0x1F) * 2)
}

// sampleTex fetches one texel. It returns six-bit colour, five-bit alpha, and whether
// the texel exists at all: a transparent texel is not a black one, it is a hole, and
// the caller must not write depth for it.
func (g *gpu3d) sampleTex(m *Machine, st *texState, s, t int) (uint8, uint8, uint8, uint8, bool) {
	s = wrapTex(s, st.sizeS, st.repeatS, st.flipS)
	t = wrapTex(t, st.sizeT, st.repeatT, st.flipT)
	i := uint32(t*st.sizeS + s)

	switch st.format {
	case 1: // A3I5: five bits of palette index, three of alpha, in one byte
		b := m.vram.read8(spTex, st.base+i)
		idx := uint32(b & 0x1F)
		a3 := uint32(b >> 5)
		// Three bits of alpha stretched to five: (a<<2)|(a>>1), so 7 reaches 31 and 0 is 0.
		a := uint8((a3 << 2) | (a3 >> 1))
		r, gg, bb := rastBGR555(m.vram.read16(spTexPal, st.pal+idx*2))
		return r, gg, bb, a, a != 0

	case 2: // 4 colours, two bits per texel
		b := m.vram.read8(spTex, st.base+i/4)
		idx := uint32(b>>(2*(i&3))) & 3
		if idx == 0 && st.color0 {
			return 0, 0, 0, 0, false
		}
		r, gg, bb := rastBGR555(m.vram.read16(spTexPal, st.pal+idx*2))
		return r, gg, bb, 31, true

	case 3: // 16 colours, four bits per texel
		b := m.vram.read8(spTex, st.base+i/2)
		idx := uint32(b>>(4*(i&1))) & 0xF
		if idx == 0 && st.color0 {
			return 0, 0, 0, 0, false
		}
		r, gg, bb := rastBGR555(m.vram.read16(spTexPal, st.pal+idx*2))
		return r, gg, bb, 31, true

	case 4: // 256 colours, a byte per texel
		idx := uint32(m.vram.read8(spTex, st.base+i))
		if idx == 0 && st.color0 {
			return 0, 0, 0, 0, false
		}
		r, gg, bb := rastBGR555(m.vram.read16(spTexPal, st.pal+idx*2))
		return r, gg, bb, 31, true

	case 5:
		return g.sampleCompressed(m, st, s, t)

	case 6: // A5I3: three bits of palette index, five of alpha
		b := m.vram.read8(spTex, st.base+i)
		idx := uint32(b & 7)
		a := uint8(b >> 3) // already five bits; no stretching needed
		r, gg, bb := rastBGR555(m.vram.read16(spTexPal, st.pal+idx*2))
		return r, gg, bb, a, a != 0

	case 7: // direct colour: a BGR555 halfword per texel, bit 15 is "this texel exists"
		c := m.vram.read16(spTex, st.base+i*2)
		if c&0x8000 == 0 {
			return 0, 0, 0, 0, false
		}
		r, gg, bb := rastBGR555(c)
		return r, gg, bb, 31, true
	}
	return 0, 0, 0, 0, false
}

// sampleCompressed reads the 4x4-texel compressed format — the one Mario is made of,
// and the only DS format whose texel data and palette *selection* live in two different
// places at once.
//
// The image is a grid of 4x4 blocks, each four bytes: one byte per row, four two-bit
// indices per byte. Alongside it, in a parallel array, sits one 16-bit word per block:
// its low 14 bits are that block's palette offset (in units of four bytes, on top of
// PLTT_BASE) and its top two bits choose how the block's four indices map to colours.
//
// The address of that parallel array is the trap. The texel data must live in texture
// slot 0 or 2 (each slot is 128 KiB), and the palette-index array lives in slot 1: the
// first half of slot 1 for data in slot 0, the second half for data in slot 2. Hence
//
//	palIdx = 0x20000 + (addr & 0x1FFFF)/2 + (slot == 2 ? 0x10000 : 0)
//
// — half the address, because there is one halfword per four bytes of texel data.
//
// UNCERTAIN: the interpolated modes' rounding. We take mode 1's midpoint as (c0+c1)/2
// and mode 3's as (5*c0+3*c1)/8 and (3*c0+5*c1)/8, computed in the five-bit palette
// space before the doubling to six bits, which is where the hardware's dividers sit.
// The colours are right to within a bit; if a compressed texture ever looks banded
// rather than wrong, this is the arithmetic to check.
func (g *gpu3d) sampleCompressed(m *Machine, st *texState, s, t int) (uint8, uint8, uint8, uint8, bool) {
	blocksW := st.sizeS / 4
	blk := uint32((t/4)*blocksW + (s / 4))
	addr := st.base + blk*4

	row := m.vram.read8(spTex, addr+uint32(t&3))
	idx := uint32(row>>(2*uint(s&3))) & 3

	slot := st.base >> 17 // 128 KiB texture slots
	palIdxAddr := 0x20000 + (addr&0x1FFFF)/2
	if slot == 2 {
		palIdxAddr += 0x10000
	}
	info := m.vram.read16(spTex, palIdxAddr)
	palAddr := st.pal + uint32(info&0x3FFF)*4
	mode := info >> 14

	// Read the two halfwords every mode needs, and the third where it is a real colour.
	raw := func(n uint32) uint16 { return m.vram.read16(spTexPal, palAddr+n*2) }

	switch mode {
	case 0: // four palette colours; index 3 is transparent
		if idx == 3 {
			return 0, 0, 0, 0, false
		}
		r, gg, bb := rastBGR555(raw(idx))
		return r, gg, bb, 31, true

	case 2: // four palette colours, all opaque
		r, gg, bb := rastBGR555(raw(idx))
		return r, gg, bb, 31, true

	case 1: // c2 is the midpoint of c0 and c1; index 3 is transparent
		switch idx {
		case 3:
			return 0, 0, 0, 0, false
		case 2:
			r, gg, bb := rastMix555(raw(0), raw(1), 1, 1, 1) // (c0 + c1) / 2
			return r, gg, bb, 31, true
		default:
			r, gg, bb := rastBGR555(raw(idx))
			return r, gg, bb, 31, true
		}

	default: // mode 3: two interpolated colours at 5/8 and 3/8, all four opaque
		switch idx {
		case 2:
			r, gg, bb := rastMix555(raw(0), raw(1), 5, 3, 3) // (5*c0 + 3*c1) / 8
			return r, gg, bb, 31, true
		case 3:
			r, gg, bb := rastMix555(raw(0), raw(1), 3, 5, 3) // (3*c0 + 5*c1) / 8
			return r, gg, bb, 31, true
		default:
			r, gg, bb := rastBGR555(raw(idx))
			return r, gg, bb, 31, true
		}
	}
}

// rastMix555 blends two BGR555 colours as (wa*a + wb*b) >> shift, per channel, in the
// five-bit palette space, and returns the engine's six-bit channels.
func rastMix555(a, b uint16, wa, wb, shift int) (uint8, uint8, uint8) {
	ch := func(n uint) uint8 {
		ca := int((a >> (5 * n)) & 0x1F)
		cb := int((b >> (5 * n)) & 0x1F)
		v := (ca*wa + cb*wb) >> uint(shift)
		if v > 31 {
			v = 31
		}
		return uint8(v * 2)
	}
	return ch(0), ch(1), ch(2)
}
