package merc

// glbprim.go turns a parsed effect into a glb.Prim: concatenated fragment
// vertices, strips reassembled from the output-slot scatter, mis-stitched
// crossings dropped by edge length (a fragment lattice spans 255 units, so
// a legitimate triangle edge is short).

import (
	"retroreverse.com/tools/lib/glb"
)

// BuildPrim assembles one effect as an indexed triangle primitive.
func BuildPrim(e *Effect, base [4]float32) glb.Prim {
	var p glb.Prim
	p.BaseColor = base
	p.DoubleSided = true
	for fi := range e.Fragments {
		fr := &e.Fragments[fi]
		vs := fr.Vertices()
		bi := uint32(len(p.Positions))
		for _, v := range vs {
			p.Positions = append(p.Positions, [3]float32{v.X, v.Y, v.Z})
			p.Normals = append(p.Normals, normalize(float32(v.NX), float32(v.NY), float32(v.NZ)))
		}
		slots := map[int]uint32{}
		for i, v := range vs {
			if v.Slot1 >= 0 {
				slots[v.Slot1] = bi + uint32(i)
			}
			if v.Slot2 >= 0 {
				slots[v.Slot2] = bi + uint32(i)
			}
		}
		mx := -1
		for s := range slots {
			if s > mx {
				mx = s
			}
		}
		for i := 2; i <= mx; i++ {
			a, aok := slots[i-2]
			b, bok := slots[i-1]
			c, cok := slots[i]
			if !aok || !bok || !cok || a == b || b == c || a == c {
				continue
			}
			pa, pb, pc := p.Positions[a], p.Positions[b], p.Positions[c]
			if dist2(pa, pb) > 40*40 || dist2(pb, pc) > 40*40 || dist2(pa, pc) > 40*40 {
				continue
			}
			if i&1 == 0 {
				p.Tris = append(p.Tris, [3]uint32{a, b, c})
			} else {
				p.Tris = append(p.Tris, [3]uint32{a, c, b})
			}
		}
	}
	return p
}

func dist2(a, b [3]float32) float32 {
	dx, dy, dz := a[0]-b[0], a[1]-b[1], a[2]-b[2]
	return dx*dx + dy*dy + dz*dz
}

func normalize(x, y, z float32) [3]float32 {
	x -= 128
	y -= 128
	z -= 128
	l := x*x + y*y + z*z
	if l == 0 {
		return [3]float32{0, 1, 0}
	}
	// Newton iterations keep the package free of math imports elsewhere;
	// precision here is cosmetic.
	inv := float32(1)
	v := l
	for i := 0; i < 16; i++ {
		v = 0.5 * (v + l/v)
	}
	inv = 1 / v
	return [3]float32{x * inv, y * inv, z * inv}
}
