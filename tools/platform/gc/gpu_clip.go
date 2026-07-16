package gc

// gpu_clip.go cuts triangles against the view frustum before the perspective divide.
//
// Two different things go wrong without it, and they look nothing alike.
//
// The NEAR plane is about direction. The divide is only meaningful for a vertex in front of the
// eye; a vertex behind it has a negative w, and dividing by a negative number reflects the
// vertex through the origin. The triangle does not vanish, it wraps across the screen, and
// because the rasteriser clamps a triangle's bounding box to the framebuffer, each wrapped
// triangle costs a full-screen fill. A scene where a tenth of the vertices sit behind the eye —
// the mansion shot of the intro, the camera looking out from inside the geometry — is then a
// soup of enormous triangles that is both wrong and slow.
//
// The SIDE planes are about precision, and this is the part that is easy to talk yourself out
// of. Coverage does not need them: the rasteriser already clamps the bounding box, so a
// triangle hanging far off-screen paints the same pixels clipped or not. But a vertex that
// passes the near plane can still project enormously far off-screen — sitting just in front of
// the near plane and far to one side is enough — and the mansion shot draws one spanning
// x[-8335705..837], y[363..1357221]. The edge functions and barycentric weights are float32
// products of coordinate differences; over eight million pixels they carry no significant bits
// left, the interpolated texture coordinate comes out of the wrong part of the texture, and the
// draw lands as a solid black band across the frame. Bounding the coordinates is the entire
// reason hardware carries a guard band, and it is why "the scissor makes side clipping
// unnecessary" is true of the pixels and false of the arithmetic that fills them.
//
// The planes are read off the game's own projection registers (see clipPos in gpu_xf.go): near
// is cz = -cw, and normalised device x and y run over [-1,1], so the sides are cx = ±cw.
//
// The far plane is not cut. A vertex beyond it has a positive, well-conditioned w, so it
// neither wraps nor loses precision; it divides to a depth past the far value and the depth
// test is left to reject it.

// guardBand is how far outside the viewport, in normalised device units, a vertex may sit
// before the side planes cut it. Any value >= 1 renders the same pixels — everything it lets
// through beyond the viewport edge is clamped away by the rasteriser's bounding box either way
// — so this is a precision and work knob, not a fidelity one, and nothing here depends on
// matching whatever width the real Flipper's band is. It is kept above 1 so that the very many
// triangles poking a little off-screen are not cut for nothing, and well below the point where
// float32 runs out of bits: at 2, screen coordinates stay inside about ±1300 pixels, where the
// edge-function products are exact to a fraction of a pixel.
const guardBand = 2

// clipPlane is a half-space in homogeneous clip space: the vertices it keeps are those where
// a*cx + b*cy + c*cz + d*cw >= 0.
type clipPlane struct{ a, b, c, d float32 }

func (p clipPlane) dist(v clipVertex) float32 {
	return p.a*v.cx + p.b*v.cy + p.c*v.cz + p.d*v.cw
}

// planeNear is cz + cw >= 0. It is also what makes the divide safe: for any perspective
// projection the hardware stores, every point at or behind the eye fails it, so no vertex with
// w <= 0 survives to be divided.
var planeNear = clipPlane{0, 0, 1, 1}

// clipPlanes is the set cut, near first — it is the one that rejects whole triangles most
// often, and it is the only one whose result the others can rely on (a positive w).
var clipPlanes = [5]clipPlane{
	planeNear,
	{1, 0, 0, guardBand},  // left:   cx >= -guardBand*cw
	{-1, 0, 0, guardBand}, // right:  cx <=  guardBand*cw
	{0, 1, 0, guardBand},  // bottom: cy >= -guardBand*cw
	{0, -1, 0, guardBand}, // top:    cy <=  guardBand*cw
}

// outcode is the set of planes a vertex falls outside, one bit per plane. A triangle whose
// three outcodes are all zero is entirely inside and needs no work; a triangle whose three
// outcodes share a bit is entirely outside that one plane and can be dropped whole. Only what
// survives both of those tests is actually cut, which is the rare case.
func outcode(v clipVertex) uint8 {
	var c uint8
	for i, p := range clipPlanes {
		if p.dist(v) < 0 {
			c |= 1 << uint(i)
		}
	}
	return c
}

// clipByPlane cuts a polygon by one plane and returns what survives, appended to out.
//
// This is Sutherland-Hodgman: walk the edges, keep every vertex on the inside, and wherever an
// edge crosses the plane, emit the crossing. The result of clipping a convex polygon by a plane
// is convex and keeps the original winding, which is what lets the caller fan-triangulate it
// and lets back-face culling still read the facing correctly.
func clipByPlane(out, in []clipVertex, p clipPlane) []clipVertex {
	out = out[:0]
	if len(in) == 0 {
		return out
	}
	prev := in[len(in)-1]
	dp := p.dist(prev)
	for _, cur := range in {
		dc := p.dist(cur)
		// The edge crosses when its ends sit on opposite sides. The crossing is where the
		// distance reaches zero, which along a straight line in clip space is at dp/(dp-dc) —
		// a denominator that cannot be zero here, because the signs are strictly opposite.
		if (dp >= 0) != (dc >= 0) {
			out = append(out, lerpClip(prev, cur, dp/(dp-dc)))
		}
		if dc >= 0 {
			out = append(out, cur)
		}
		prev, dp = cur, dc
	}
	return out
}

// clipTriangle cuts a triangle against the frustum and returns the surviving polygon as a
// fan, in the caller's buffers — a draw of many triangles allocates nothing per triangle. The
// result is empty when nothing survives, and the three original vertices, unmodified, when the
// triangle is entirely inside.
func clipTriangle(dst, scratch []clipVertex, tri [3]clipVertex) ([]clipVertex, []clipVertex) {
	c0, c1, c2 := outcode(tri[0]), outcode(tri[1]), outcode(tri[2])
	if c0|c1|c2 == 0 { // wholly inside: the common case, and it must not be perturbed
		return append(dst[:0], tri[0], tri[1], tri[2]), scratch
	}
	if c0&c1&c2 != 0 { // wholly outside one plane
		return dst[:0], scratch
	}

	poly := append(dst[:0], tri[0], tri[1], tri[2])
	other := scratch
	for i, p := range clipPlanes {
		// Only cut by a plane some vertex is actually outside of.
		if (c0|c1|c2)&(1<<uint(i)) == 0 {
			continue
		}
		other = clipByPlane(other, poly, p)
		poly, other = other, poly
		if len(poly) < 3 {
			return poly[:0], other
		}
	}
	return poly, other
}

// clipAndSetup cuts a triangle against the frustum and SETS UP whatever survives, dividing and
// viewport-mapping each surviving vertex on the way — so the divide only ever sees a vertex the
// clipper has vouched for. It returns the two buffers for the next triangle to reuse.
//
// It sets up rather than draws: the triangles land in g.tris and the whole primitive's worth is
// filled together, which is what lets the fill partition the screen between workers rather than
// hand each triangle its own. See gpu.fill.
func (g *gpu) clipAndSetup(m *Machine, dst, scratch []clipVertex, v0, v1, v2 clipVertex) ([]clipVertex, []clipVertex) {
	poly, scratch := clipTriangle(dst, scratch, [3]clipVertex{v0, v1, v2})
	if len(poly) < 3 {
		return poly, scratch
	}
	toScreen := func(c clipVertex) screenVertex {
		sx, sy, sz := g.toScreen(c.cx, c.cy, c.cz, c.cw)
		// The near clip guarantees a positive w, so this reciprocal is always finite.
		return screenVertex{x: sx, y: sy, z: sz, r: c.r, g: c.g, b: c.b, a: c.a,
			tc: c.tc, ntc: c.ntc, invW: 1 / c.cw}
	}
	// The surviving polygon is convex and still wound as the source was, so a fan from its
	// first vertex retriangulates it without changing which way any triangle faces — which
	// matters, because back-face culling reads that winding.
	a := toScreen(poly[0])
	for i := 1; i+1 < len(poly); i++ {
		if t, ok := g.setupTri(a, toScreen(poly[i]), toScreen(poly[i+1])); ok {
			g.tris = append(g.tris, t)
		}
	}
	return poly, scratch
}
