package xbox

// nv2a_clip.go: homogeneous near-plane clipping.
//
// The push buffer submits geometry that can cross the eye — a vertex with clip w <= 0 is at
// or behind the camera. Its perspective divide (screen = clip/w) produces a screen position
// of the wrong sign and unbounded magnitude, so a triangle joining an on-screen vertex to a
// behind-eye one spans the whole framebuffer. Left unclipped that is not cosmetic: on
// an3-drive the tyre-smoke particle strips put 218 such triangles on the screen and they alone
// produced 7.6M of the frame's 10.3M fragments (74%) — a translucent grey wash over the
// picture and the reason the rasteriser's cost exploded. (The GameCube has the same gap; see
// the reference-match-hides-bugs lesson.)
//
// The hardware clips primitives against the frustum before the divide. We do the essential
// part of that here: clip every triangle against the near plane w >= clipWEps so no vertex
// reaching the rasteriser has a non-positive w. A triangle wholly in front is passed through
// UNCHANGED (byte-identical — the point of the fast path), one wholly behind is dropped, and a
// straddling one is cut, generating new vertices ON the plane whose attributes are the
// clip-space-linear interpolation of the edge's endpoints.
//
// The wrinkle is that the program-mode vertex path hands us positions the guest's shader
// epilogue has ALREADY divided (screen x,y,z, with clip w preserved in pos[3]). To clip in
// clip space we invert the viewport map (the c58/c59 scale/offset the epilogue and the
// fixed-function path both apply), clip, then re-project the new vertices. The inversion is
// exact (an affine map composed with the divide), so it changes nothing for a triangle that
// does not cross — only crossing triangles, which are garbage as it stands, are rebuilt.

import "math"

// clipWEps is the near-plane threshold in clip-space w. It is a small positive value: the
// pathology is w <= 0 (at/behind the eye), and any positive w projects to a finite position,
// so clipping just above zero removes the divide blow-up while leaving genuine geometry
// (whose w is the view-space depth, far larger) untouched.
const clipWEps = 1.0 / 65536

// clipNearPlane clips a draw's triangle list against the near plane. It returns the vertices
// and triangles to rasterise — the originals unchanged when nothing crosses (the common case,
// no allocation), or the reusable scratch buffers carrying the clipped geometry.
func (g *pgraph) clipNearPlane(verts []kelvinVtx, tris [][3]int) ([]kelvinVtx, [][3]int) {
	crosses := false
	for i := range verts {
		if verts[i].pos[3] <= clipWEps {
			crosses = true
			break
		}
	}
	if !crosses {
		return verts, tris // byte-exact fast path: nothing to clip
	}

	sc, off := &g.Const[vshSlotViewportScale], &g.Const[vshSlotViewportOffset]
	sx, sy, sz := math.Float32frombits(sc[0]), math.Float32frombits(sc[1]), math.Float32frombits(sc[2])
	ox, oy, oz := math.Float32frombits(off[0]), math.Float32frombits(off[1]), math.Float32frombits(off[2])
	if sx == 0 || sy == 0 || sz == 0 {
		return verts, tris // no viewport to invert: leave the draw as-is rather than divide by zero
	}
	clipOf := func(v *kelvinVtx) [4]float32 {
		w := v.pos[3]
		return [4]float32{(v.pos[0] - ox) / sx * w, (v.pos[1] - oy) / sy * w, (v.pos[2] - oz) / sz * w, w}
	}
	project := func(c [4]float32) [4]float32 {
		w := c[3]
		return [4]float32{c[0]/w*sx + ox, c[1]/w*sy + oy, c[2]/w*sz + oz, w}
	}

	// out starts as a copy of the draw's vertices, so a kept triangle keeps its original
	// indices; clipped triangles append their new vertices past the copied prefix.
	g.clipVerts = append(g.clipVerts[:0], verts...)
	out := g.clipVerts
	otris := g.clipTris[:0]

	var poly [4]kelvinVtx // one triangle clipped against one plane yields at most 4 vertices
	for _, t := range tris {
		va, vb, vc := &verts[t[0]], &verts[t[1]], &verts[t[2]]
		ia, ib, ic := va.pos[3] > clipWEps, vb.pos[3] > clipWEps, vc.pos[3] > clipWEps
		nin := 0
		if ia {
			nin++
		}
		if ib {
			nin++
		}
		if ic {
			nin++
		}
		if nin == 3 {
			otris = append(otris, t) // wholly in front: original vertices, unchanged
			continue
		}
		if nin == 0 {
			continue // wholly behind the eye: dropped
		}

		// Sutherland-Hodgman over the closed edge loop a->b->c->a: emit each inside vertex as
		// the start of its outgoing edge, plus the crossing point of every in/out edge.
		n := 0
		edge := func(p, q *kelvinVtx, pin, qin bool) {
			if pin {
				poly[n] = *p
				n++
			}
			if pin != qin {
				cp, cq := clipOf(p), clipOf(q)
				tt := (clipWEps - cp[3]) / (cq[3] - cp[3])
				poly[n] = clipVertex(p, q, cp, cq, tt, project)
				n++
			}
		}
		edge(va, vb, ia, ib)
		edge(vb, vc, ib, ic)
		edge(vc, va, ic, ia)

		base := len(out)
		out = append(out, poly[:n]...)
		for k := 1; k+1 < n; k++ { // fan-triangulate the clipped polygon
			otris = append(otris, [3]int{base, base + k, base + k + 1})
		}
	}
	g.clipVerts, g.clipTris = out, otris
	return out, otris
}

// clipVertex builds the vertex where edge p->q crosses the near plane at parameter t (linear
// in clip space, where w reaches clipWEps). Its position is the reconstructed clip point
// re-projected to screen; every shaded attribute is the clip-space-linear interpolation, which
// is what the rasteriser's own perspective-correct step expects to be handed.
func clipVertex(p, q *kelvinVtx, cp, cq [4]float32, t float32, project func([4]float32) [4]float32) kelvinVtx {
	var cn [4]float32
	for i := 0; i < 4; i++ {
		cn[i] = cp[i] + t*(cq[i]-cp[i])
	}
	cn[3] = clipWEps // exact, so no vertex reaching the rasteriser divides by <= 0
	var v kelvinVtx
	v.pos = project(cn)
	for i := 0; i < 4; i++ {
		v.d0[i] = p.d0[i] + t*(q.d0[i]-p.d0[i])
		v.d1[i] = p.d1[i] + t*(q.d1[i]-p.d1[i])
	}
	v.fog = p.fog + t*(q.fog-p.fog)
	for u := 0; u < 4; u++ {
		for i := 0; i < 4; i++ {
			v.uv[u][i] = p.uv[u][i] + t*(q.uv[u][i]-p.uv[u][i])
		}
	}
	return v
}
