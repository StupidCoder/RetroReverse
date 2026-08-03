package merc

// glbprim.go turns a parsed effect into glb.Prims from disc data alone:
// topology via EffectStream (strips.go), positions/normals/colors/UVs from
// the vertex and record streams, one primitive per adgif block, textures
// resolved by the caller (tpage + level remap table).

import (
	"encoding/binary"
	"image"

	"retroreverse.com/tools/lib/glb"
)

// Material is a resolved adgif block: what the GLB needs to render it.
type Material struct {
	Image image.Image
	Blend bool
}

// TexturedPrims builds one effect's primitives, one per adgif block, with
// UVs (dest bytes / 128, REPEAT) and per-vertex colors (records, GS scale
// 0x80 = 1.0). resolve maps a shader to its texture; a nil Image falls
// back to flat gold.
func TexturedPrims(e *Effect, resolve func(ShaderRef) Material) []glb.Prim {
	seq, shaders := EffectStream(e)

	// effect-wide vertex tables
	var pos [][3]float32
	var nrm [][3]float32
	var uv [][2]float32
	var col [][4]uint8
	for fi := range e.Fragments {
		fr := &e.Fragments[fi]
		cols := fr.Colors()
		for i, v := range fr.Vertices() {
			pos = append(pos, [3]float32{v.X, v.Y, v.Z})
			nrm = append(nrm, normalize(float32(v.NX), float32(v.NY), float32(v.NZ)))
			uv = append(uv, [2]float32{v.U, v.V})
			c := [4]uint8{255, 255, 255, 255}
			if i < len(cols) {
				c = [4]uint8{gs(cols[i][0]), gs(cols[i][1]), gs(cols[i][2]), gs(cols[i][3])}
			}
			col = append(col, c)
		}
	}

	prims := make([]glb.Prim, len(shaders))
	remap := make([]map[int]uint32, len(shaders))
	for i := range prims {
		m := Material{}
		if resolve != nil {
			m = resolve(shaders[i])
		}
		prims[i].DoubleSided = true
		prims[i].Image = m.Image
		prims[i].Blend = m.Blend
		if m.Image == nil {
			prims[i].BaseColor = [4]float32{0.83, 0.68, 0.28, 1}
		} else {
			prims[i].BaseColor = [4]float32{1, 1, 1, 1}
		}
		remap[i] = map[int]uint32{}
	}
	local := func(mi, v int) uint32 {
		if li, ok := remap[mi][v]; ok {
			return li
		}
		p := &prims[mi]
		li := uint32(len(p.Positions))
		p.Positions = append(p.Positions, pos[v])
		p.Normals = append(p.Normals, nrm[v])
		p.UVs = append(p.UVs, uv[v])
		p.Colors = append(p.Colors, col[v])
		remap[mi][v] = li
		return li
	}

	var w [3]int
	n := 0
	for _, sv := range seq {
		w[0], w[1], w[2] = w[1], w[2], sv.Index
		if n < 2 {
			n++
			continue
		}
		if sv.ADC || w[0] == w[1] || w[1] == w[2] || w[0] == w[2] {
			continue
		}
		mi := sv.Mat
		if mi < 0 {
			mi = 0
		}
		p := &prims[mi]
		p.Tris = append(p.Tris, [3]uint32{local(mi, w[0]), local(mi, w[1]), local(mi, w[2])})
	}
	out := prims[:0]
	for _, p := range prims {
		if len(p.Tris) == 0 {
			continue
		}
		for _, c := range p.Colors {
			if c[3] < 250 {
				p.Blend = true
				break
			}
		}
		out = append(out, p)
	}
	return out
}

// gs rescales a GS color byte (0x80 = 1.0) to linear 0-255.
func gs(v uint8) uint8 {
	x := int(v) * 255 / 128
	if x > 255 {
		x = 255
	}
	return uint8(x)
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
