package track

// horizon.go decodes the race view's horizon scenery — the mountain range —
// drawn by the yaw-placed 2-D object renderer $6953E. The race-init config
// step $696FC reads the placement list via the pointer table at $69A80
// (the track index is overridden to 0, so every track shows the same range):
// a count byte then (yaw, model) pairs into $69736/$69766. Each frame $6953E
// walks the list and draws every object whose compass heading falls in the
// view window: screen-left x = (yaw*256 - cameraYaw16) >> 3 (one yaw unit =
// 32 pixels, the full 256-unit compass = 8192 pixels), base line pinned to
// the horizon (y_screen = -($1BC42>>3) - y).
//
// A model (pointer table $699B8, 8 bytes per entry: shape ptr + variable
// stream ptr) is: word vertex count; per vertex two words (x, y) where a
// NEGATIVE word is a placeholder filled from the variable stream (this is how
// 14 mountain silhouettes share one 4-vertex shape with per-entry peak
// widths/heights); byte edge count + (a,b) vertex byte-offset pairs; byte
// face count + faces {colour byte, vertex count, edge display-list offsets}.
// The mountains are palette 5 (mid grey); a second two-triangle shape
// (entries 14/15, palettes 4/5) exists in the table but config 0 never
// places it. Verified against the engine by cmd/horizonoracle.

const (
	horizonCfgTable   = 0x69A80
	horizonModelTable = 0x699B8
)

// HorizonFace is one filled face: its palette colour and the indexes of the
// edges bounding it (3 = triangle fill $6950C, 4 = quad fill $66618).
type HorizonFace struct {
	Pal   byte
	Edges []int
}

// HorizonModel is one decoded object shape with its variable stream applied:
// vertex (x, y) in object space (x right, y up from the horizon line, in
// screen pixels), the edge list (vertex index pairs), and the faces.
type HorizonModel struct {
	Verts [][2]int
	Edges [][2]int
	Faces []HorizonFace
}

// HorizonPlacement is one entry of the placement list: the object's compass
// heading (256 units = full circle; screen-left edge at yaw*32 px) and its
// model number.
type HorizonPlacement struct {
	Yaw   byte
	Model int
}

func (im *Image) s16(a int) int {
	v := im.u16(a)
	if v >= 0x8000 {
		return v - 0x10000
	}
	return v
}

// horizonModel decodes model n from the $699B8 table, substituting the
// negative coordinate placeholders from the entry's variable stream.
func (im *Image) horizonModel(n int) HorizonModel {
	shape := int(uint32(im.u16(horizonModelTable+8*n))<<16 | uint32(im.u16(horizonModelTable+8*n+2)))
	vs := int(uint32(im.u16(horizonModelTable+8*n+4))<<16 | uint32(im.u16(horizonModelTable+8*n+6)))

	var m HorizonModel
	a := shape
	cnt := im.s16(a)
	a += 2
	for i := 0; i < cnt; i++ {
		var xy [2]int
		for c := 0; c < 2; c++ {
			v := im.s16(a)
			a += 2
			if v < 0 { // placeholder: next word from the variable stream
				v = im.s16(vs)
				vs += 2
			}
			xy[c] = v
		}
		m.Verts = append(m.Verts, xy)
	}
	ec := im.u8(a)
	a++
	for i := 0; i < ec; i++ {
		// edge operands are byte offsets into the screen-coord arrays
		// (vertex index * 2)
		m.Edges = append(m.Edges, [2]int{im.u8(a) / 2, im.u8(a+1) / 2})
		a += 2
	}
	fc := im.u8(a)
	a++
	for i := 0; i < fc; i++ {
		f := HorizonFace{Pal: byte(im.u8(a))}
		nv := im.u8(a + 1)
		a += 2
		for j := 0; j < nv; j++ {
			// face operands are display-list offsets of the edges (index * 4)
			f.Edges = append(f.Edges, im.u8(a)/4)
			a++
		}
		m.Faces = append(m.Faces, f)
	}
	return m
}

// Horizon decodes the placement list (config 0 — $696FC forces it for every
// track) and the distinct models it uses, each with its variable stream
// applied.
func (im *Image) Horizon() ([]HorizonPlacement, map[int]HorizonModel) {
	cfg := int(uint32(im.u16(horizonCfgTable))<<16 | uint32(im.u16(horizonCfgTable+2)))
	cnt := im.u8(cfg)
	var out []HorizonPlacement
	models := map[int]HorizonModel{}
	for i := 0; i < cnt; i++ {
		p := HorizonPlacement{
			Yaw:   byte(im.u8(cfg + 1 + 2*i)),
			Model: im.u8(cfg + 2 + 2*i),
		}
		out = append(out, p)
		if _, ok := models[p.Model]; !ok {
			models[p.Model] = im.horizonModel(p.Model)
		}
	}
	return out, models
}
