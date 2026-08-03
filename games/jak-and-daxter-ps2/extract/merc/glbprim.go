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

// SkinTracker carries the running bone-palette state across a ctrl's
// fragments: MatXfer uploads persist in VU memory, so vertices may address
// matrices uploaded by earlier fragments; 0xFF reuses the previous
// vertex's matrix.
type SkinTracker struct {
	dests map[byte]byte
	last  uint8
}

func NewSkinTracker() *SkinTracker {
	return &SkinTracker{dests: map[byte]byte{}}
}

// TexturedPrims builds one effect's primitives, one per adgif block, with
// UVs (dest bytes / 128, REPEAT) and per-vertex colors (records, GS scale
// 0x80 = 1.0). resolve maps a shader to its texture; a nil Image falls
// back to flat gold.
func TexturedPrims(e *Effect, resolve func(ShaderRef) Material) []glb.Prim {
	return texturedPrims(e, resolve, nil, 1)
}

// TexturedPrimsSkinned additionally fills per-vertex joint indices
// (palette index - 1) and scales positions into joint space (the merc-ctrl
// +28 scale).
func TexturedPrimsSkinned(e *Effect, resolve func(ShaderRef) Material, tr *SkinTracker, posScale float32) []glb.Prim {
	return texturedPrims(e, resolve, tr, posScale)
}

func texturedPrims(e *Effect, resolve func(ShaderRef) Material, tr *SkinTracker, posScale float32) []glb.Prim {
	seq, shaders := EffectStream(e)

	// effect-wide vertex tables
	var pos [][3]float32
	var nrm [][3]float32
	var uv [][2]float32
	var col [][4]uint8
	var jnt []uint8
	for fi := range e.Fragments {
		fr := &e.Fragments[fi]
		cols := fr.Colors()
		if tr != nil {
			for _, m := range fr.Mats {
				tr.dests[m.Dest] = m.Index
			}
		}
		for i, v := range fr.Vertices() {
			pos = append(pos, [3]float32{v.X * posScale, v.Y * posScale, v.Z * posScale})
			nrm = append(nrm, normalize(float32(v.NX), float32(v.NY), float32(v.NZ)))
			uv = append(uv, [2]float32{v.U, v.V})
			c := [4]uint8{255, 255, 255, 255}
			if i < len(cols) {
				c = [4]uint8{gs(cols[i][0]), gs(cols[i][1]), gs(cols[i][2]), gs(cols[i][3])}
			}
			col = append(col, c)
			if tr != nil {
				j := tr.last
				if v.Mat != 0xFF {
					if idx, ok := tr.dests[v.Mat&0x7F]; ok && idx > 0 {
						j = idx - 1
					}
				}
				tr.last = j
				jnt = append(jnt, j)
			}
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
		prims[i].Unlit = true // GS look: texture x vertex color, no runtime lighting
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
		if jnt != nil {
			p.Joints = append(p.Joints, jnt[v])
		}
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
		// The GS never culls, so the strips carry no winding convention;
		// wind each triangle to agree with its own vertex normals.
		a, b, c := w[0], w[1], w[2]
		av, bv, cv := pos[a], pos[b], pos[c]
		ux, uy, uz := bv[0]-av[0], bv[1]-av[1], bv[2]-av[2]
		vx, vy, vz := cv[0]-av[0], cv[1]-av[1], cv[2]-av[2]
		gx, gy, gz := uy*vz-uz*vy, uz*vx-ux*vz, ux*vy-uy*vx
		ns := [3]float32{
			nrm[a][0] + nrm[b][0] + nrm[c][0],
			nrm[a][1] + nrm[b][1] + nrm[c][1],
			nrm[a][2] + nrm[b][2] + nrm[c][2],
		}
		if gx*ns[0]+gy*ns[1]+gz*ns[2] < 0 {
			a, b = b, a
		}
		p.Tris = append(p.Tris, [3]uint32{local(mi, a), local(mi, b), local(mi, c)})
	}
	out := prims[:0]
	for _, p := range prims {
		if len(p.Tris) > 0 {
			out = append(out, p)
		}
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
