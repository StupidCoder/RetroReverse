package xbox

import (
	"math"
	"testing"
)

// clipTestGraph builds a bare pgraph with a standard 640x480 viewport in the c58/c59 slots,
// enough for clipNearPlane to reconstruct clip space and re-project.
func clipTestGraph() *pgraph {
	g := &pgraph{}
	set := func(slot int, x, y, z float32) {
		g.Const[slot][0] = math.Float32bits(x)
		g.Const[slot][1] = math.Float32bits(y)
		g.Const[slot][2] = math.Float32bits(z)
	}
	set(vshSlotViewportScale, 320, -240, 16777215)
	set(vshSlotViewportOffset, 320, 240, 0)
	return g
}

// vtx makes a vertex at a given screen position and clip w, with attributes seeded from w so
// interpolation can be checked.
func vtx(sx, sy, sz, w float32) kelvinVtx {
	return kelvinVtx{pos: [4]float32{sx, sy, sz, w}, d0: [4]float32{w, 0, 0, 1}}
}

func TestClipAllInFrontIsIdentity(t *testing.T) {
	g := clipTestGraph()
	verts := []kelvinVtx{vtx(100, 100, 100, 5), vtx(300, 100, 100, 5), vtx(200, 300, 100, 5)}
	tris := [][3]int{{0, 1, 2}}
	ov, ot := g.clipNearPlane(verts, tris)
	if len(ot) != 1 || ot[0] != tris[0] {
		t.Fatalf("a wholly-in-front triangle must pass through unchanged, got %v", ot)
	}
	if &ov[0] != &verts[0] {
		t.Errorf("the fast path must return the original vertex slice, not a copy")
	}
}

func TestClipAllBehindIsCulled(t *testing.T) {
	g := clipTestGraph()
	verts := []kelvinVtx{vtx(100, 100, 100, -5), vtx(300, 100, 100, -3), vtx(200, 300, 100, -1)}
	_, ot := g.clipNearPlane(verts, [][3]int{{0, 1, 2}})
	if len(ot) != 0 {
		t.Fatalf("a wholly-behind-eye triangle must be culled, got %d triangles", len(ot))
	}
}

func TestClipStraddleProducesPositiveW(t *testing.T) {
	g := clipTestGraph()
	// Two vertices in front, one behind: clipping yields a 4-vertex polygon -> 2 triangles.
	for _, tc := range []struct {
		name  string
		w     [3]float32
		wantN int
	}{
		{"two-in-one-out", [3]float32{5, 5, -5}, 2},
		{"one-in-two-out", [3]float32{5, -5, -5}, 1},
	} {
		verts := []kelvinVtx{vtx(100, 100, 100, tc.w[0]), vtx(300, 100, 100, tc.w[1]), vtx(200, 300, 100, tc.w[2])}
		ov, ot := g.clipNearPlane(verts, [][3]int{{0, 1, 2}})
		if len(ot) != tc.wantN {
			t.Errorf("%s: got %d triangles, want %d", tc.name, len(ot), tc.wantN)
		}
		for _, tri := range ot {
			for _, idx := range tri {
				if w := ov[idx].pos[3]; w <= 0 {
					t.Errorf("%s: a clipped vertex reached the rasteriser with w=%g (<= 0)", tc.name, w)
				}
			}
		}
	}
}
