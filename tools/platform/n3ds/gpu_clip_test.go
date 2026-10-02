package n3ds

import (
	"math"
	"testing"
)

func TestClipCameraCrossingGround(t *testing.T) {
	g := testGPU()
	fb := fbState{width: 8, height: 8, vpHalfW: 4, vpHalfH: 4}
	a := vsOut{pos: [4]float32{-.5, -.5, -.5, 1}}
	b := vsOut{pos: [4]float32{.5, -.5, -.5, 1}}
	c := vsOut{pos: [4]float32{0, 2, .5, -1}}
	if _, ok := g.setupTri(&a, &b, &c, &fb); ok {
		t.Fatal("fixture should cross camera plane")
	}
	tris := g.clipTri(nil, a, b, c, &fb)
	if len(tris) == 0 {
		t.Fatal("visible ground was discarded")
	}
	covered := false
	for _, tri := range tris {
		for _, v := range []scrVert{tri.v0, tri.v1, tri.v2} {
			if math.IsNaN(float64(v.x)) || v.x < -.001 || v.x > 8.001 || v.y < -.001 || v.y > 8.001 || v.z < -1.001 || v.z > .001 || v.iw <= 0 {
				t.Fatalf("unclipped vertex: %+v", v)
			}
		}
		if edgeFn(tri.v0.x, tri.v0.y, tri.v1.x, tri.v1.y, 4, 3) >= 0 && edgeFn(tri.v1.x, tri.v1.y, tri.v2.x, tri.v2.y, 4, 3) >= 0 && edgeFn(tri.v2.x, tri.v2.y, tri.v0.x, tri.v0.y, 4, 3) >= 0 {
			covered = true
		}
	}
	if !covered {
		t.Fatal("visible ground sample is still a hole")
	}
}

func TestClipInsideOutsideAndInvalid(t *testing.T) {
	g := testGPU()
	fb := fbState{width: 8, height: 8, vpHalfW: 4, vpHalfH: 4}
	a := vsOut{pos: [4]float32{-.5, -.5, -.5, 1}}
	b := vsOut{pos: [4]float32{.5, -.5, -.5, 1}}
	c := vsOut{pos: [4]float32{0, .5, -.5, 1}}
	want, ok := g.setupTri(&a, &b, &c, &fb)
	if !ok {
		t.Fatal("fixture")
	}
	got := g.clipTri(nil, a, b, c, &fb)
	if len(got) != 1 || got[0] != want {
		t.Fatal("inside triangle changed")
	}
	for plane := 0; plane < 6; plane++ {
		verts := [3]vsOut{a, b, c}
		axis := plane / 2
		value := float32(-2)
		if plane%2 == 1 {
			value = 2
		}
		for i := range verts {
			verts[i].pos[axis] = value
		}
		if len(g.clipTri(nil, verts[0], verts[1], verts[2], &fb)) != 0 {
			t.Fatalf("outside plane %d survived", plane)
		}
	}
	for plane := 0; plane < 6; plane++ {
		v := c
		value := float32(-2)
		if plane%2 == 1 {
			value = 2
		}
		v.pos[plane/2] = value
		clipped := g.clipTri(nil, a, b, v, &fb)
		if len(clipped) != 2 {
			t.Fatalf("partial plane %d: got %d triangles", plane, len(clipped))
		}
		for _, tri := range clipped {
			for _, p := range []scrVert{tri.v0, tri.v1, tri.v2} {
				if p.x < -.001 || p.x > 8.001 || p.y < -.001 || p.y > 8.001 || p.z < -1.001 || p.z > .001 {
					t.Fatalf("partial plane %d: %+v", plane, p)
				}
			}
		}
	}
	for _, bad := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		for axis := 0; axis < 4; axis++ {
			v := c
			v.pos[axis] = bad
			if len(g.clipTri(nil, a, b, v, &fb)) != 0 {
				t.Fatalf("invalid component %d survived", axis)
			}
		}
	}
}

func TestClipInterpolatesAllVaryings(t *testing.T) {
	var a, b vsOut
	for i := 0; i < 4; i++ {
		b.pos[i] = 4
		b.color[i] = 8
		b.quat[i] = 12
	}
	for i := 0; i < 3; i++ {
		b.view[i] = 16
		for j := 0; j < 2; j++ {
			b.uv[i][j] = 20
		}
	}
	b.uv0w = 24
	v := clipVertex(a, b, .25)
	for i := 0; i < 4; i++ {
		if v.pos[i] != 1 || v.color[i] != 2 || v.quat[i] != 3 {
			t.Fatal("position, colour or quaternion interpolation")
		}
	}
	for i := 0; i < 3; i++ {
		if v.view[i] != 4 || v.uv[i][0] != 5 || v.uv[i][1] != 5 {
			t.Fatal("view or UV interpolation")
		}
	}
	if v.uv0w != 6 {
		t.Fatal("projective texture coordinate interpolation")
	}
}
