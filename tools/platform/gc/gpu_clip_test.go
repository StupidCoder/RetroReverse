package gc

import (
	"math"
	"testing"
)

// The projection this game programs for the intro's 3D scenes, read out of XF memory while the
// mansion shot was on screen. Its p4/p5 are the compact form's -n/(f-n) and -f*n/(f-n), which
// solve to a near plane of 1.0 and a far plane of 327680 — the numbers the clip-space
// convention in clipPos is derived from.
const (
	mansionP4 = -3.0517672e-06
	mansionP5 = -1.0000031
)

// TestProjectionSolvesToItsPlanes checks the reading of the projection registers that the
// clipper's near plane rests on: that treating p4/p5 as -n/(f-n) and -f*n/(f-n) reproduces the
// game's own constants. If this is the wrong reading of the compact form, the near plane is at
// the wrong place and every other test here is measuring the wrong thing.
func TestProjectionSolvesToItsPlanes(t *testing.T) {
	// Solve the pair for n and f: f = p5/p4, then n = -p4*(f-n) rearranged.
	f := float64(mansionP5) / float64(mansionP4)
	n := -float64(mansionP4) * f / (1 - float64(mansionP4))

	if math.Abs(n-1.0) > 1e-3 {
		t.Errorf("near plane solved to %g, want ~1.0", n)
	}
	// Both planes land on round numbers — near at 1.0, far at 327680 (2^15 * 10) — which is
	// the strongest evidence that this is the right reading of the compact form: an incorrect
	// interpretation of p4/p5 would not solve to values a person would plausibly have typed.
	if math.Abs(f-327680) > 1.0 {
		t.Errorf("far plane solved to %g, want ~327680", f)
	}

	// Round-trip: those planes must regenerate the constants the game actually wrote.
	p4 := -n / (f - n)
	p5 := -f * n / (f - n)
	if math.Abs(p4-float64(mansionP4)) > 1e-12 {
		t.Errorf("p4 round-trip %g, want %g", p4, mansionP4)
	}
	if math.Abs(p5-float64(mansionP5)) > 1e-6 {
		t.Errorf("p5 round-trip %g, want %g", p5, mansionP5)
	}
}

// gpuWithMansionProjection builds a gpu whose XF memory holds the game's perspective projection
// and a 640x480 viewport with the depth range the game uses.
func gpuWithMansionProjection() *gpu {
	g := &gpu{}
	set := func(a int, v float32) { g.XFMem[a] = math.Float32bits(v) }
	set(0x1020, 1.8684956) // p0
	set(0x1021, 0)         // p1
	set(0x1022, 2.2048247) // p2
	set(0x1023, 0)         // p3
	set(0x1024, mansionP4) // p4
	set(0x1025, mansionP5) // p5
	g.XFMem[0x1026] = 0    // perspective
	set(0x101A, 320)       // viewport scale x
	set(0x101B, -240)      // scale y
	set(0x101C, 16777215)  // scale z
	set(0x101D, 342+320)   // offset x (the hardware's 342 bias plus the centre)
	set(0x101E, 342+240)   // offset y
	set(0x101F, 16777215)  // offset z
	return g
}

// TestClipSpaceZConvention verifies against the live projection that normalised device z is -1
// at the near plane and 0 at the far plane, and that the viewport maps those to depth 0 and
// 0xFFFFFF. This is the convention planeNear (cz = -cw) is stated in.
func TestClipSpaceZConvention(t *testing.T) {
	g := gpuWithMansionProjection()
	const near, far = 1.0, 327680

	for _, c := range []struct {
		name    string
		ez      float32
		wantNDC float32
		wantSZ  float32
	}{
		{"near plane", -near, -1, 0},
		{"far plane", -far, 0, 16777215},
	} {
		cx, cy, cz, cw := g.clipPos(0, 0, c.ez)
		if cw <= 0 {
			t.Fatalf("%s: w = %g, want positive", c.name, cw)
		}
		if ndc := cz / cw; math.Abs(float64(ndc-c.wantNDC)) > 1e-3 {
			t.Errorf("%s: ndc z = %g, want %g", c.name, ndc, c.wantNDC)
		}
		_, _, sz := g.toScreen(cx, cy, cz, cw)
		if math.Abs(float64(sz-c.wantSZ)) > 64 {
			t.Errorf("%s: depth = %g, want ~%g", c.name, sz, c.wantSZ)
		}
	}
}

// cv builds a clip-space vertex at an eye-space z through the mansion projection.
func cv(g *gpu, ex, ey, ez float32) clipVertex {
	cx, cy, cz, cw := g.clipPos(ex, ey, ez)
	return clipVertex{cx: cx, cy: cy, cz: cz, cw: cw, r: 255, a: 255}
}

// TestClipNearCases covers the three outcomes of cutting a triangle by one plane.
func TestClipNearCases(t *testing.T) {
	g := gpuWithMansionProjection()

	for _, c := range []struct {
		name    string
		ez      [3]float32 // eye-space z of the three vertices (negative is in front)
		wantLen int
	}{
		{"all in front", [3]float32{-10, -20, -30}, 3},
		{"all behind the eye", [3]float32{10, 20, 30}, 0},
		{"all nearer than the near plane", [3]float32{-0.5, -0.2, -0.1}, 0},
		{"one behind: a corner is cut off, making a quad", [3]float32{-10, -20, 30}, 4},
		{"two behind: a corner survives as a triangle", [3]float32{-10, 20, 30}, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			tri := [3]clipVertex{
				cv(g, 1, 1, c.ez[0]), cv(g, -1, 1, c.ez[1]), cv(g, 0, -1, c.ez[2]),
			}
			got := clipByPlane(nil, tri[:], planeNear)
			if len(got) != c.wantLen {
				t.Fatalf("got %d vertices, want %d", len(got), c.wantLen)
			}
			// Whatever survives must be safe to divide and must be in front of the plane.
			for i, v := range got {
				if v.cw <= 0 {
					t.Errorf("vertex %d survived with w = %g; the divide would wrap it", i, v.cw)
				}
				if v.cz+v.cw < -1e-3 {
					t.Errorf("vertex %d survived behind the near plane (cz+cw = %g)", i, v.cz+v.cw)
				}
			}
		})
	}
}

// TestClipNearLeavesFrontTriangleExactlyAlone guards the common case: a triangle wholly in
// front must pass through bit-for-bit, or clipping would perturb every frame that was already
// correct.
func TestClipNearLeavesFrontTriangleExactlyAlone(t *testing.T) {
	g := gpuWithMansionProjection()
	tri := [3]clipVertex{cv(g, 1, 1, -10), cv(g, -1, 1, -20), cv(g, 0, -1, -30)}
	tri[0].u, tri[0].v = 0.25, 0.75
	tri[1].g, tri[1].b = 128, 64

	got := clipByPlane(nil, tri[:], planeNear)
	if len(got) != 3 {
		t.Fatalf("got %d vertices, want 3", len(got))
	}
	for i := range got {
		if got[i] != tri[i] {
			t.Errorf("vertex %d changed: got %+v, want %+v", i, got[i], tri[i])
		}
	}
}

// TestClipNearInterpolatesAttributes checks that a vertex the clipper invents carries the
// attributes the rasteriser will read. An edge crossing the plane exactly halfway must land a
// vertex with the midpoint's colour and texture coordinate; a clipper that emitted the right
// position but a zero colour would still draw a plausible — and wrong — frame.
func TestClipNearInterpolatesAttributes(t *testing.T) {
	// A hand-built clip-space triangle, so the crossing is exactly at t = 0.5 and the expected
	// attribute values are exact rather than approximate.
	//
	// Vertex 0 sits one unit inside the plane (cz+cw = +1), vertex 1 one unit outside (-1), so
	// the 0->1 edge crosses at the midpoint.
	v0 := clipVertex{cx: 0, cy: 0, cz: 0, cw: 1, r: 0, g: 0, b: 0, a: 0, u: 0, v: 0}
	v1 := clipVertex{cx: 0, cy: 0, cz: -2, cw: 1, r: 100, g: 200, b: 40, a: 80, u: 1, v: 3}
	v2 := clipVertex{cx: 4, cy: 0, cz: 0, cw: 1, r: 0, g: 0, b: 0, a: 0, u: 0, v: 0}

	if d0, d1 := v0.cz+v0.cw, v1.cz+v1.cw; d0 != 1 || d1 != -1 {
		t.Fatalf("test setup wrong: distances %g and %g, want +1 and -1", d0, d1)
	}

	got := clipByPlane(nil, []clipVertex{v0, v1, v2}, planeNear)
	if len(got) != 4 {
		t.Fatalf("got %d vertices, want 4", len(got))
	}

	// got[1] is the crossing on the 0->1 edge: the midpoint of v0 and v1.
	mid := got[1]
	for _, c := range []struct {
		name     string
		got, exp float32
	}{
		{"r", float32(mid.r), 50}, {"g", float32(mid.g), 100},
		{"b", float32(mid.b), 20}, {"a", float32(mid.a), 40},
		{"u", mid.u, 0.5}, {"v", mid.v, 1.5},
		{"cz", mid.cz, -1}, {"cw", mid.cw, 1},
	} {
		if math.Abs(float64(c.got-c.exp)) > 0.51 {
			t.Errorf("crossing vertex %s = %v, want %v", c.name, c.got, c.exp)
		}
	}
	// The crossing must land exactly on the plane.
	if d := mid.cz + mid.cw; math.Abs(float64(d)) > 1e-6 {
		t.Errorf("crossing vertex is %g off the near plane, want on it", d)
	}
}

// TestClipNearPreservesWinding is the guard that ties clipping to back-face culling: the
// clipped polygon must wind the same way the source triangle did, or a cut triangle would flip
// its facing and the culler would drop exactly the triangles that should be drawn. (The repo
// has already been bitten once by a culling sign convention that looked fine either way.)
func TestClipNearPreservesWinding(t *testing.T) {
	g := gpuWithMansionProjection()

	// area2 is twice the signed screen-space area — its sign is the winding.
	area2 := func(a, b, c screenVertex) float32 {
		return (b.x-a.x)*(c.y-a.y) - (c.x-a.x)*(b.y-a.y)
	}
	toScreen := func(v clipVertex) screenVertex {
		sx, sy, sz := g.toScreen(v.cx, v.cy, v.cz, v.cw)
		return screenVertex{x: sx, y: sy, z: sz}
	}

	// The cut vertices sit at ez = -0.5: in front of the eye, so their w is positive and their
	// unclipped screen position is finite and meaningful, but nearer than the near plane at
	// 1.0, so the clipper still cuts them. That is what makes the source winding a valid
	// reference — taking it from a vertex behind the eye would measure the wrapped position
	// this whole change exists to prevent.
	for _, c := range []struct {
		name string
		tri  [3]clipVertex
	}{
		{"uncut", [3]clipVertex{cv(g, 1, 1, -10), cv(g, -1, 1, -10), cv(g, 0, -1, -10)}},
		{"one vertex cut", [3]clipVertex{cv(g, 1, 1, -10), cv(g, -1, 1, -10), cv(g, 0, -1, -0.5)}},
		{"two vertices cut", [3]clipVertex{cv(g, 1, 1, -10), cv(g, -1, 1, -0.5), cv(g, 0, -1, -0.5)}},
		{"reversed winding, one vertex cut", [3]clipVertex{cv(g, 0, -1, -0.5), cv(g, -1, 1, -10), cv(g, 1, 1, -10)}},
	} {
		t.Run(c.name, func(t *testing.T) {
			want := area2(toScreen(c.tri[0]), toScreen(c.tri[1]), toScreen(c.tri[2]))
			poly := clipByPlane(nil, c.tri[:], planeNear)
			if len(poly) < 3 {
				t.Fatal("nothing survived; this case should produce a polygon")
			}
			// Every triangle of the fan the drawer builds must wind like the source.
			a := toScreen(poly[0])
			for i := 1; i+1 < len(poly); i++ {
				got := area2(a, toScreen(poly[i]), toScreen(poly[i+1]))
				if (got > 0) != (want > 0) {
					t.Errorf("fan triangle %d winds %g, source winds %g: facing flipped", i, got, want)
				}
			}
		})
	}
}

// TestClipNearRejectsEveryNonPositiveW is the property the divide depends on, checked across a
// sweep of eye-space depths through the game's own projection: no vertex at or behind the eye
// may ever reach toScreen. This is what makes the "if cw == 0" guard there unreachable rather
// than load-bearing.
func TestClipNearRejectsEveryNonPositiveW(t *testing.T) {
	g := gpuWithMansionProjection()
	for a := -50; a <= 50; a += 3 {
		for b := -50; b <= 50; b += 7 {
			for c := -50; c <= 50; c += 11 {
				tri := [3]clipVertex{
					cv(g, 1, 1, float32(a)), cv(g, -1, 1, float32(b)), cv(g, 0, -1, float32(c)),
				}
				for i, v := range clipByPlane(nil, tri[:], planeNear) {
					if v.cw <= 0 {
						t.Fatalf("ez=(%d,%d,%d): vertex %d survived with w=%g", a, b, c, i, v.cw)
					}
				}
			}
		}
	}
}

// TestClipByPlaneReusesBuffer checks the allocation contract rasterPrimitive relies on: the
// clipper writes into the caller's buffer, so a long strip clips without allocating per
// triangle.
func TestClipByPlaneReusesBuffer(t *testing.T) {
	g := gpuWithMansionProjection()
	buf := make([]clipVertex, 0, 8)
	tri := [3]clipVertex{cv(g, 1, 1, -10), cv(g, -1, 1, -20), cv(g, 0, -1, 30)}

	got := clipByPlane(buf, tri[:], planeNear)
	if len(got) != 4 {
		t.Fatalf("got %d vertices, want 4", len(got))
	}
	if cap(got) != cap(buf) || &got[:1][0] != &buf[:1][0] {
		t.Error("clipByPlane allocated instead of using the caller's buffer")
	}
}

// TestClipTriangleBoundsScreenCoordinates is the guard band's whole reason for existing, stated
// as the thing that went wrong: the mansion shot draws a triangle spanning x[-8335705..837],
// y[363..1357221], whose vertices all pass the near plane and whose w is perfectly positive.
// Nothing about it wraps — it is simply enormous, and float32 edge functions over eight million
// pixels have no significant bits left, so its texture coordinate came out of the wrong part of
// the texture and it painted a black band. After clipping, every surviving vertex must sit
// within the guard band, which bounds the arithmetic.
func TestClipTriangleBoundsScreenCoordinates(t *testing.T) {
	g := gpuWithMansionProjection()

	// A vertex just in front of the near plane and far to one side: in front of the eye, w
	// positive, and projecting millions of pixels away.
	tri := [3]clipVertex{
		cv(g, -14000, 2000, -1.01),
		cv(g, 10, 8, -30),
		cv(g, -5, -8, -30),
	}
	// The setup must really be the pathological case, or the test proves nothing.
	sx, _, _ := g.toScreen(tri[0].cx, tri[0].cy, tri[0].cz, tri[0].cw)
	if tri[0].cw <= 0 {
		t.Fatalf("test setup wrong: w = %g, want positive (this is not a wrap)", tri[0].cw)
	}
	if math.Abs(float64(sx)) < 1e6 {
		t.Fatalf("test setup wrong: unclipped x = %g, want millions of pixels off-screen", sx)
	}

	poly, _ := clipTriangle(nil, nil, tri)
	if len(poly) < 3 {
		t.Fatal("the whole triangle was dropped; it crosses the viewport and must survive")
	}
	for i, v := range poly {
		x, y, _ := g.toScreen(v.cx, v.cy, v.cz, v.cw)
		// The guard band is in NDC; at scale 320/240 plus the viewport offset, staying inside
		// it keeps screen coordinates within a few thousand pixels rather than millions.
		if math.Abs(float64(x)) > 4000 || math.Abs(float64(y)) > 4000 {
			t.Errorf("vertex %d survived at (%g,%g); the guard band did not bound it", i, x, y)
		}
		if v.cw <= 0 {
			t.Errorf("vertex %d survived with w = %g", i, v.cw)
		}
	}
}

// TestClipTriangleLeavesInsideTriangleExactlyAlone pins the fast path: a triangle wholly within
// the guard band must come through bit-for-bit. Every ordinary draw in every scene takes this
// path, so anything else would perturb frames that are already correct.
func TestClipTriangleLeavesInsideTriangleExactlyAlone(t *testing.T) {
	g := gpuWithMansionProjection()
	tri := [3]clipVertex{cv(g, 1, 1, -10), cv(g, -1, 1, -20), cv(g, 0, -1, -30)}
	tri[0].u, tri[0].v = 0.25, 0.75

	poly, _ := clipTriangle(nil, nil, tri)
	if len(poly) != 3 {
		t.Fatalf("got %d vertices, want the original 3", len(poly))
	}
	for i := range poly {
		if poly[i] != tri[i] {
			t.Errorf("vertex %d changed: got %+v, want %+v", i, poly[i], tri[i])
		}
	}
}

// TestGuardBandDoesNotMoveTheSurface is the property that makes the guard-band width a free
// choice rather than a fidelity claim: clipping cuts the polygon but must not move the surface
// it represents. For a point inside the viewport, the perspective-correct texture coordinate
// must be the same whether or not the triangle was cut — otherwise picking guardBand would be
// picking what the frame looks like.
//
// The check reconstructs the coordinate at a screen point from the clipped fan and compares it
// to the unclipped triangle's own plane equation, evaluated independently.
func TestGuardBandDoesNotMoveTheSurface(t *testing.T) {
	g := gpuWithMansionProjection()

	// A big ground-plane-ish triangle reaching well outside the guard band, with texture
	// coordinates that vary across it.
	tri := [3]clipVertex{
		cv(g, -3000, -20, -8),
		cv(g, 3000, -20, -8),
		cv(g, 0, -20, -4000),
	}
	tri[0].u, tri[0].v = 0, 0
	tri[1].u, tri[1].v = 8, 0
	tri[2].u, tri[2].v = 0, 8

	poly, _ := clipTriangle(nil, nil, tri)
	if len(poly) < 3 {
		t.Fatal("the triangle was dropped entirely")
	}

	toScreen := func(c clipVertex) screenVertex {
		sx, sy, sz := g.toScreen(c.cx, c.cy, c.cz, c.cw)
		return screenVertex{x: sx, y: sy, z: sz, u: c.u, v: c.v, invW: 1 / c.cw}
	}
	// The unclipped triangle, as ground truth.
	t0, t1, t2 := toScreen(tri[0]), toScreen(tri[1]), toScreen(tri[2])
	area := edge(t0.x, t0.y, t1.x, t1.y, t2.x, t2.y)

	checked := 0
	for _, p := range [][2]float32{{320, 300}, {200, 400}, {450, 350}, {320, 420}, {100, 460}} {
		// Where does the unclipped triangle say the surface is at this pixel?
		w0 := edge(t1.x, t1.y, t2.x, t2.y, p[0], p[1]) / area
		w1 := edge(t2.x, t2.y, t0.x, t0.y, p[0], p[1]) / area
		w2 := edge(t0.x, t0.y, t1.x, t1.y, p[0], p[1]) / area
		if w0 < 0 || w1 < 0 || w2 < 0 {
			continue // not covered by the source triangle; nothing to compare
		}
		wantU, wantV := perspUV(w0, w1, w2, t0, t1, t2)

		// And what does the clipped fan say at the same pixel?
		gotU, gotV, found := float32(0), float32(0), false
		a := toScreen(poly[0])
		for i := 1; i+1 < len(poly); i++ {
			b, c := toScreen(poly[i]), toScreen(poly[i+1])
			ar := edge(a.x, a.y, b.x, b.y, c.x, c.y)
			if ar == 0 {
				continue
			}
			e0 := edge(b.x, b.y, c.x, c.y, p[0], p[1]) / ar
			e1 := edge(c.x, c.y, a.x, a.y, p[0], p[1]) / ar
			e2 := edge(a.x, a.y, b.x, b.y, p[0], p[1]) / ar
			if e0 < -1e-4 || e1 < -1e-4 || e2 < -1e-4 {
				continue
			}
			gotU, gotV = perspUV(e0, e1, e2, a, b, c)
			found = true
			break
		}
		if !found {
			t.Errorf("pixel (%g,%g) is inside the source triangle but no clipped fan triangle covers it", p[0], p[1])
			continue
		}
		if math.Abs(float64(gotU-wantU)) > 1e-2 || math.Abs(float64(gotV-wantV)) > 1e-2 {
			t.Errorf("pixel (%g,%g): clipped fan says uv (%v,%v), source surface says (%v,%v)",
				p[0], p[1], gotU, gotV, wantU, wantV)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no pixel was actually compared; the test proved nothing")
	}
}

// TestClipTriangleRejectsWhollyOutside checks the cheap reject: a triangle entirely off one
// side must be dropped, not cut.
func TestClipTriangleRejectsWhollyOutside(t *testing.T) {
	g := gpuWithMansionProjection()
	for _, c := range []struct {
		name string
		tri  [3]clipVertex
	}{
		{"far off the left", [3]clipVertex{cv(g, -900, 0, -10), cv(g, -800, 5, -10), cv(g, -850, -5, -10)}},
		{"far off the top", [3]clipVertex{cv(g, 0, 900, -10), cv(g, 5, 800, -10), cv(g, -5, 850, -10)}},
		{"behind the eye", [3]clipVertex{cv(g, 1, 1, 10), cv(g, -1, 1, 20), cv(g, 0, -1, 30)}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if poly, _ := clipTriangle(nil, nil, c.tri); len(poly) != 0 {
				t.Errorf("got %d vertices, want the triangle dropped", len(poly))
			}
		})
	}
}
