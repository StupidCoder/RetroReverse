package merc

// strips.go builds a fragment's triangles from the DISC DATA alone.
//
// A fragment's vertices scatter into output slots (the two dest bytes in
// each lump vertex; slot = (byte-7)/3), and the GS consumes that output
// buffer as one triangle strip. Each vertex also owns a 4-byte record in
// the byte region at header[12]*4: {u, v, ?, flag}. The flag is 0x80 for
// an ordinary strip vertex; 0x01/0x00 mark the vertices the microprogram
// tags ADC — the strip restarts there. Reconstructing with those restarts
// needs no emulation: the file states the topology.

import "retroreverse.com/tools/lib/glb"

// StripPrim builds one effect's primitive from file data only.
func StripPrim(e *Effect, base [4]float32) glb.Prim {
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
		if len(fr.ByteData) < 16 {
			continue
		}
		recBase := int(fr.ByteData[12]) * 4
		flagOf := func(v int) byte {
			o := recBase + v*4 + 3
			if o < len(fr.ByteData) {
				return fr.ByteData[o]
			}
			return 0x80
		}
		// slot -> vertex
		slots := map[int]int{}
		for i, v := range vs {
			if v.Slot1 >= 0 {
				slots[v.Slot1] = i
			}
			if v.Slot2 >= 0 {
				slots[v.Slot2] = i
			}
		}
		mx := -1
		for s := range slots {
			if s > mx {
				mx = s
			}
		}
		var w [3]int
		n := 0
		for s := 0; s <= mx; s++ {
			v, ok := slots[s]
			if !ok {
				n = 0
				continue
			}
			if flagOf(v) != 0x80 {
				// ADC-tagged: enters the window but closes the run before it
				w[0], w[1], w[2] = w[1], w[2], v
				n = 1
				continue
			}
			w[0], w[1], w[2] = w[1], w[2], v
			if n < 2 {
				n++
				continue
			}
			a, b, c := bi+uint32(w[0]), bi+uint32(w[1]), bi+uint32(w[2])
			if a == b || b == c || a == c {
				continue
			}
			if s&1 == 0 {
				p.Tris = append(p.Tris, [3]uint32{a, b, c})
			} else {
				p.Tris = append(p.Tris, [3]uint32{a, c, b})
			}
		}
	}
	return p
}
