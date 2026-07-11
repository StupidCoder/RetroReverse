package psp

import "fmt"

// ge_patch.go tessellates the GE's curved-surface primitives — SPLINE (a uniform
// cubic B-spline surface with clamp-edge flags) and BEZIER (piecewise cubic Bézier
// patches) — into triangles for the software rasterizer. The game's soft-body
// characters are drawn as SPLINE surfaces whose control points the CPU physics
// writes to VRAM each frame (indexed via IADDR).

// drawPatch decodes the (ucount × vcount) control-point grid and rasterizes the
// evaluated surface. arg: bits 0-7 ucount, 8-15 vcount, 16-17 spline edge flags
// (bit 0 clamps the u starts/ends, bit 1 the v).
func (m *Machine) drawPatch(s *geState, arg uint32, spline bool) {
	nu := int(arg & 0xFF)
	nv := int((arg >> 8) & 0xFF)
	if nu < 4 || nv < 4 || nu*nv > 4096 {
		return
	}
	gePrimSeq++
	if gePrimSeq == gePrimDump {
		m.dumpPrim(s, 9, nu*nv)
	}
	if gePrimDump >= 0 {
		geWordTrail = geWordTrail[:0]
	}
	ctrl := m.decodeVerts(s, nu*nv)
	if len(ctrl) < nu*nv {
		return
	}
	if geDebugN > 0 {
		geDebugN--
		kind := "BEZIER"
		if spline {
			kind = "SPLINE"
		}
		logf := "%s#%d %dx%d vt=%06X va=%08X ia=%08X div=%d,%d tex=%v@%08X mat=%08X dif=%06X light=%v c0=(%.1f,%.1f,%.1f)\n"
		fmtPrintf(logf, kind, gePrimSeq, nu, nv, s.vtype, s.vaddr, s.iaddr,
			s.patchDivU, s.patchDivV, s.texEnable, s.texAddr, s.matColor, s.matDiffuse,
			s.lightOn, ctrl[0].x, ctrl[0].y, ctrl[0].z)
	}

	divU := int(s.patchDivU)
	divV := int(s.patchDivV)
	if divU < 1 {
		divU = 4
	}
	if divV < 1 {
		divV = 4
	}

	// sample counts across the whole surface
	var su, sv int
	if spline {
		su = (nu-3)*divU + 1
		sv = (nv-3)*divV + 1
	} else {
		su = (nu / 3 * divU) + 1
		sv = (nv / 3 * divV) + 1
	}
	if su < 2 || sv < 2 || su*sv > 1<<16 {
		return
	}

	clampU := !spline || arg&(1<<16) != 0
	clampV := !spline || arg&(1<<17) != 0

	// the colour of a colourless patch: material ambient, or the material diffuse
	// when lighting is on (the lit soft bodies read as their diffuse colour)
	cr, cg, cb := byte(s.matColor), byte(s.matColor>>8), byte(s.matColor>>16)
	ca := byte(s.matColor >> 24)
	if s.lightOn && (s.vtype>>2)&7 == 0 {
		cr, cg, cb = byte(s.matDiffuse), byte(s.matDiffuse>>8), byte(s.matDiffuse>>16)
	}
	hasUV := s.vtype&3 != 0

	grid := make([]vert, su*sv)
	for j := 0; j < sv; j++ {
		tv := float32(j) / float32(sv-1)
		for i := 0; i < su; i++ {
			tu := float32(i) / float32(su-1)
			var x, y, z, uu, vv float32
			if spline {
				x, y, z, uu, vv = splineSurf(ctrl, nu, nv, tu, tv, clampU, clampV, hasUV)
			} else {
				x, y, z, uu, vv = bezierSurf(ctrl, nu, nv, tu, tv, hasUV)
			}
			var out vert
			out.x, out.y, out.z = x, y, z
			if hasUV {
				out.u, out.v = uu, vv
			} else {
				// parametric texcoords through the texture scale/offset registers
				out.u = tu*s.texScaleU + s.texOffU
				out.v = tv*s.texScaleV + s.texOffV
			}
			out.r, out.g, out.b, out.a = cr, cg, cb, ca
			grid[j*su+i] = out
		}
	}
	// control points come out of decodeVerts already transformed to screen
	// space, so the evaluated surface needs no further transform
	for j := 0; j+1 < sv; j++ {
		for i := 0; i+1 < su; i++ {
			a := grid[j*su+i]
			b := grid[j*su+i+1]
			c := grid[(j+1)*su+i]
			d := grid[(j+1)*su+i+1]
			m.rasterTri(s, a, b, c)
			m.rasterTri(s, b, d, c)
		}
	}
}

// splineSurf evaluates the uniform cubic B-spline surface at parameter (tu, tv)
// in [0,1]², with clamped or open ends per axis, returning position and (when the
// control points carry texcoords) interpolated UVs.
func splineSurf(ctrl []vert, nu, nv int, tu, tv float32, clampU, clampV, hasUV bool) (x, y, z, u, v float32) {
	ku := splineKnots(nu, clampU)
	kv := splineKnots(nv, clampV)
	// map t in [0,1] to the valid domain [knot[3], knot[n]]
	pu := ku[3] + tu*(ku[nu]-ku[3])
	pv := kv[3] + tv*(kv[nv]-kv[3])
	wu := splineBasis(ku, nu, pu)
	wv := splineBasis(kv, nv, pv)
	for j := 0; j < nv; j++ {
		if wv[j] == 0 {
			continue
		}
		for i := 0; i < nu; i++ {
			w := wu[i] * wv[j]
			if w == 0 {
				continue
			}
			p := ctrl[j*nu+i]
			x += w * p.x
			y += w * p.y
			z += w * p.z
			if hasUV {
				u += w * p.u
				v += w * p.v
			}
		}
	}
	return
}

// splineKnots builds the cubic B-spline knot vector for n control points:
// uniform integers, with the three outer knots collapsed onto the domain edge
// when that edge is clamped (the GE's spline edge flags).
func splineKnots(n int, clamped bool) []float32 {
	k := make([]float32, n+4)
	for i := range k {
		k[i] = float32(i)
	}
	if clamped {
		k[0], k[1], k[2] = k[3], k[3], k[3]
		k[n+1], k[n+2], k[n+3] = k[n], k[n], k[n]
	}
	return k
}

// splineBasis computes the n cubic B-spline basis weights at parameter t by the
// Cox-de Boor recursion (0/0 taken as 0).
func splineBasis(knot []float32, n int, t float32) []float32 {
	// clamp inside the domain so the end point lands in the last span
	if t >= knot[n] {
		t = knot[n] - 1e-4
	}
	if t < knot[3] {
		t = knot[3]
	}
	w := make([]float32, n+3)
	// degree 0
	for i := 0; i < n+3; i++ {
		if knot[i] <= t && t < knot[i+1] {
			w[i] = 1
		}
	}
	// raise degree
	for d := 1; d <= 3; d++ {
		for i := 0; i+d+1 < n+4; i++ {
			var a, b float32
			if den := knot[i+d] - knot[i]; den != 0 {
				a = (t - knot[i]) / den * w[i]
			}
			if den := knot[i+d+1] - knot[i+1]; den != 0 {
				b = (knot[i+d+1] - t) / den * w[i+1]
			}
			w[i] = a + b
		}
	}
	return w[:n]
}

// bezierSurf evaluates piecewise cubic Bézier patches ((nu,nv) = 3k+1 control
// points) at (tu, tv) in [0,1]².
func bezierSurf(ctrl []vert, nu, nv int, tu, tv float32, hasUV bool) (x, y, z, u, v float32) {
	wu := bezierBasis(nu, tu)
	wv := bezierBasis(nv, tv)
	for j := 0; j < nv; j++ {
		if wv[j] == 0 {
			continue
		}
		for i := 0; i < nu; i++ {
			w := wu[i] * wv[j]
			if w == 0 {
				continue
			}
			p := ctrl[j*nu+i]
			x += w * p.x
			y += w * p.y
			z += w * p.z
			if hasUV {
				u += w * p.u
				v += w * p.v
			}
		}
	}
	return
}

// bezierBasis computes per-control-point Bernstein weights for a chain of cubic
// patches: parameter t in [0,1] spans the chain; each patch takes 4 points and
// shares its last with the next patch's first.
func bezierBasis(n int, t float32) []float32 {
	w := make([]float32, n)
	patches := (n - 1) / 3
	pt := t * float32(patches)
	pi := int(pt)
	if pi >= patches {
		pi = patches - 1
	}
	lt := pt - float32(pi)
	it := 1 - lt
	b0 := it * it * it
	b1 := 3 * lt * it * it
	b2 := 3 * lt * lt * it
	b3 := lt * lt * lt
	base := pi * 3
	w[base] += b0
	w[base+1] += b1
	w[base+2] += b2
	w[base+3] += b3
	return w
}

// fmtPrintf is a tiny indirection so the debug print in drawPatch matches the
// PRIM debug output stream.
func fmtPrintf(format string, args ...any) { fmt.Printf(format, args...) }
