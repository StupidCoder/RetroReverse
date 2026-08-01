package merc

// glbprim.go turns a parsed effect into a glb.Prim. Topology comes from the
// microprogram itself (EmulateTopology): the kicked packet's vertex order
// and ADC bits, keyed back to input vertices by index-encoding the color
// stream. Positions/normals come from the direct lattice decode (verified
// against the rendered logo); colors from the fragment's record stream.

import (
	"encoding/binary"

	"retroreverse.com/tools/lib/glb"
)

// Decoder carries the emulation context for topology extraction.
type Decoder struct {
	Micro   []byte
	LowMem  []byte
	STMagic uint32
	CtrlRow []byte // the merc-ctrl's +28 quadword (low-mem 139, per model)
}

// BuildPrim assembles one effect as an indexed triangle primitive with
// microprogram-exact strips: the effect's fragments run sequentially in one
// double-buffered VU session, vertex identity globally index-encoded.
func (d *Decoder) BuildPrim(e *Effect, base [4]float32) (glb.Prim, error) {
	var p glb.Prim
	p.BaseColor = base
	p.DoubleSided = true
	sess, err := NewSession(d.Micro, d.LowMem, d.CtrlRow, d.STMagic)
	if err != nil {
		return p, err
	}
	var frBase []int
	for fi := range e.Fragments {
		fr := &e.Fragments[fi]
		frBase = append(frBase, len(p.Positions))
		vs := fr.Vertices()
		colBase := (int(fr.ByteData[12]) + 1) * 4
		for i, v := range vs {
			p.Positions = append(p.Positions, [3]float32{v.X, v.Y, v.Z})
			p.Normals = append(p.Normals, normalize(float32(v.NX), float32(v.NY), float32(v.NZ)))
			o := colBase + i*4
			c := [4]uint8{200, 200, 200, 255}
			if o+3 < len(fr.ByteData) {
				c = [4]uint8{fr.ByteData[o], fr.ByteData[o+1], fr.ByteData[o+2], 255}
			}
			p.Colors = append(p.Colors, c)
		}
	}
	for fi := range e.Fragments {
		fr := &e.Fragments[fi]
		topo, err := sess.RunFragment(fr, frBase[fi])
		if err != nil {
			return p, err
		}
		var w [3]int
		n := 0
		for ti, t := range topo {
			if t.Index < 0 || t.Index >= len(p.Positions) {
				n = 0
				continue
			}
			w[0], w[1], w[2] = w[1], w[2], t.Index
			if n < 2 {
				n++
				continue
			}
			if t.ADC {
				continue
			}
			a, b, c := uint32(w[0]), uint32(w[1]), uint32(w[2])
			if a == b || b == c || a == c {
				continue
			}
			// Strip-stitch carriers produce degenerate slivers the GS
			// draws off-screen; a fragment lattice spans 255 units, so a
			// long edge marks one.
			pa, pb, pc := p.Positions[a], p.Positions[b], p.Positions[c]
			if dist2(pa, pb) > 40*40 || dist2(pb, pc) > 40*40 || dist2(pa, pc) > 40*40 {
				continue
			}
			if ti&1 == 0 {
				p.Tris = append(p.Tris, [3]uint32{a, b, c})
			} else {
				p.Tris = append(p.Tris, [3]uint32{a, c, b})
			}
		}
	}
	return p, nil
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
	v := l
	for i := 0; i < 16; i++ {
		v = 0.5 * (v + l/v)
	}
	return [3]float32{x / v, y / v, z / v}
}

// CtrlSTMagic reads the merc-ctrl's +44 word (the STROW value).
func CtrlSTMagic(obj []byte, p uint32) uint32 {
	return binary.LittleEndian.Uint32(obj[p+44:])
}
