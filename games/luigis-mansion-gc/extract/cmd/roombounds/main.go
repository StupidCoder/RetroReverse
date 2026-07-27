package main

import (
	"math"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/platform/gc"
)

func main() {
	d, err := gc.Open(os.Args[1])
	if err != nil { panic(err) }
	defer d.Close()
	var arcs []string
	for _, f := range d.FST.Entries {
		if !f.Dir && strings.HasPrefix(f.Path, "/Iwamoto/map2/room_") && strings.HasSuffix(f.Path, ".arc") {
			arcs = append(arcs, f.Path)
		}
	}
	sort.Strings(arcs)
	for _, arc := range arcs {
		var b []byte
		for _, f := range d.FST.Entries {
			if !f.Dir && f.Path == arc { b, _ = d.Read(f.Offset, int(f.Size)) }
		}
		if len(b) >= 4 && string(b[:4]) == "Yay0" { b, _ = lm.Yay0(b) }
		members, err := lm.RARC(b)
		if err != nil { continue }
		for _, mem := range members {
			if mem.Name != "room.bin" { continue }
			m, err := lm.ParseBin(mem.Data)
			if err != nil { fmt.Printf("%s: %v\n", arc, err); continue }
			// world bounds via graph compose (like binGLB)
			world := make([]lm.Mtx34, len(m.Nodes))
			var walk func(i int, parent lm.Mtx34)
			walk = func(i int, parent lm.Mtx34) {
				for i >= 0 && i < len(m.Nodes) {
					n := &m.Nodes[i]
					world[i] = parent // rooms are identity mostly; compose anyway
					_ = n
					world[i] = parent.Mul(nodeMtx(n))
					if c := n.Child; c >= 0 { walk(c, world[i]) }
					i = n.Next
				}
			}
			if len(m.Nodes) > 0 { walk(0, lm.Identity34) }
			min := [3]float32{1e9, 1e9, 1e9}
			max := [3]float32{-1e9, -1e9, -1e9}
			ident := true
			for ni := range m.Nodes {
				w := world[ni]
				if w != lm.Identity34 { ident = false }
				for _, pair := range m.Nodes[ni].Pairs {
					if pair.Mesh >= len(m.Meshes) { continue }
					dl, err := m.ParseBinDL(&m.Meshes[pair.Mesh])
					if err != nil { continue }
					for _, p := range dl {
						for _, v := range p.Verts {
							pt := w.Apply(m.Positions[v.Pos])
							for c := 0; c < 3; c++ {
								if pt[c] < min[c] { min[c] = pt[c] }
								if pt[c] > max[c] { max[c] = pt[c] }
							}
						}
					}
				}
			}
			name := strings.TrimSuffix(strings.TrimPrefix(arc, "/Iwamoto/map2/"), ".arc")
			fmt.Printf("%-9s ident=%v  x %7.0f..%7.0f  y %6.0f..%6.0f  z %7.0f..%7.0f\n",
				name, ident, min[0], max[0], min[1], max[1], min[2], max[2])
		}
	}
}

func nodeMtx(n *lm.BinNode) lm.Mtx34 {
	// same as binNodeMatrix
	rad := func(d float32) float64 { return float64(d) * 3.14159265358979 / 180 }
	sx, cx := sincos(rad(n.Rot[0]))
	sy, cy := sincos(rad(n.Rot[1]))
	sz, cz := sincos(rad(n.Rot[2]))
	r := lm.Mtx34{
		cy * cz, cz*sy*sx - sz*cx, cz*sy*cx + sz*sx, n.Trans[0],
		sz * cy, sz*sy*sx + cz*cx, sz*sy*cx - cz*sx, n.Trans[1],
		-sy, cy * sx, cy * cx, n.Trans[2],
	}
	for row := 0; row < 3; row++ {
		r[row*4] *= n.Scale[0]
		r[row*4+1] *= n.Scale[1]
		r[row*4+2] *= n.Scale[2]
	}
	return r
}

func sincos(a float64) (float32, float32) {
	s, c := math.Sin(a), math.Cos(a)
	return float32(s), float32(c)
}
