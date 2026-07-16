package gc

// gpu_clip.go cuts triangles against the near plane before the perspective divide.
//
// The divide is what turns a projected vertex into a pixel, and it is only meaningful for a
// vertex in front of the eye. A vertex behind the eye has a negative w, and dividing by a
// negative number reflects the vertex through the origin: the triangle does not vanish, it
// wraps across the screen as a huge one, and because the rasteriser clamps a triangle's
// bounding box to the framebuffer, each wrapped triangle costs a full-screen fill. A scene
// where a tenth of the vertices sit behind the eye — which is what the mansion shot of the
// intro is, the camera looking out from inside the geometry — is then a soup of enormous
// triangles that is both wrong and slow. Clipping is not an optimisation here; it is the
// difference between drawing the scene and drawing its reflection through the eye point.
//
// The plane is cz = -cw, read off the game's own projection registers (see clipPos in
// gpu_xf.go). Clipping against it is also what makes the divide safe: for any perspective
// projection the hardware stores, every point at or behind the eye fails cz + cw >= 0, so no
// vertex with w <= 0 survives to be divided.
//
// Only the near plane is cut. The left/right/top/bottom planes need no cut because the
// rasteriser already clamps to the framebuffer and a vertex that survives the near plane has a
// positive w, so its screen position is finite and the clamp is exact — this is the same
// reasoning the hardware's guard band rests on. The far plane needs no cut because a vertex
// beyond it divides to a depth past the far value and the depth test rejects it.

// clipNear cuts a triangle against the near plane and returns the polygon that survives, as a
// vertex fan. The result is empty when the triangle is entirely behind the plane, the original
// three vertices when it is entirely in front, and a four-vertex polygon when the plane cuts a
// corner off — the classic Sutherland-Hodgman result of clipping a triangle by one plane.
//
// The buffer is the caller's, so a draw of many triangles allocates nothing per triangle.
func clipNear(out []clipVertex, tri [3]clipVertex) []clipVertex {
	out = out[:0]

	// The signed distance to the near plane: positive inside, negative behind. Because the
	// plane is cz = -cw, the distance is simply cz + cw.
	dist := func(v clipVertex) float32 { return v.cz + v.cw }

	for i := 0; i < 3; i++ {
		cur := tri[i]
		nxt := tri[(i+1)%3]
		dc, dn := dist(cur), dist(nxt)

		if dc >= 0 {
			out = append(out, cur)
		}
		// The edge crosses the plane when its ends sit on opposite sides. The crossing point is
		// where the distance reaches zero, which on a straight line in clip space is at
		// dc/(dc-dn) — a denominator that cannot be zero here, because the two distances have
		// strictly opposite signs.
		if (dc >= 0) != (dn >= 0) {
			out = append(out, lerpClip(cur, nxt, dc/(dc-dn)))
		}
	}
	return out
}

// clipAndDraw takes a triangle in clip space, cuts it against the near plane, and rasterises
// whatever survives — dividing and viewport-mapping each surviving vertex on the way, so that
// the divide only ever sees a vertex the clipper has vouched for.
func (g *gpu) clipAndDraw(m *Machine, buf []clipVertex, v0, v1, v2 clipVertex) []clipVertex {
	poly := clipNear(buf, [3]clipVertex{v0, v1, v2})
	if len(poly) < 3 {
		return poly
	}
	toScreen := func(c clipVertex) screenVertex {
		sx, sy, sz := g.toScreen(c.cx, c.cy, c.cz, c.cw)
		// The clipper guarantees a positive w, so this reciprocal is always finite.
		return screenVertex{x: sx, y: sy, z: sz, r: c.r, g: c.g, b: c.b, a: c.a,
			u: c.u, v: c.v, invW: 1 / c.cw}
	}
	// The surviving polygon is convex and its vertices are in the original winding order, so a
	// fan from its first vertex retriangulates it without changing which way any triangle faces
	// — which matters, because back-face culling reads that winding.
	a := toScreen(poly[0])
	for i := 1; i+1 < len(poly); i++ {
		g.drawTriangle(m, a, toScreen(poly[i]), toScreen(poly[i+1]))
	}
	return poly
}
