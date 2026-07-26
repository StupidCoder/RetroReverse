package lm

// sls.go reads the .sls/.slk pair — the characters' vertex (blend-shape)
// animation, reverse-engineered from the evaluator at 0x8005C464 (found by
// read-watching the file in RAM while the cutscene drew).
//
// The .sls carries groups of vertices — Luigi's facial regions, all bound to
// the head joints — whose entries in the model's POSITION array are rebuilt
// every frame: first zeroed, then accumulated as weight * shape for each
// active shape ("influence"), then divided by the summed weight. A shape's
// geometry is a stream of u16s, one per vertex: the low 13 bits index a
// shared dictionary of f32[3] vectors and the top three bits negate x/y/z
// (mirror compression for a symmetric face). The weights come from the .slk —
// the same {count, offset} channel tables over a float pool, hermite on the
// 30 fps frame timeline, as everywhere else in this engine — except that a
// group with a single shape is hardwired to weight 1.0 and the .slk is never
// consulted (the opening Luigis are all single-shape: the scared face).
//
// .sls layout:
//
//	+0x08 u16 groupCount   +0x0a u16 shapeEntryCount   +0x0c u16 posDictCount
//	+0x0e u16 nrmDictCount
//	+0x10 → group records, 20 B: {u16 group, u16 -, u16 chanBase, u16 -[5],
//	         u16 shapeCount, u16 vertexCount}
//	+0x14 → shape entries, 8 B: {u16 posCount, u16 posStreamOff,
//	         u16 nrmCount, u16 nrmOff}   (shapeCount+1 per group)
//	+0x18 → position dictionary, f32[3] each
//	+0x1c → normal dictionary, f32[3] each
//	+0x20 → u16 shape streams        +0x24 → (unused here)
//	+0x28 → active-shape entry table: per influence, its shape-entry index
//	+0x2c → target vertex-index list (groups' slices, in order)
//
// The .slk: {u32 -, u16 frameCount, u16 -, u16 channelCount}, then offsets to
// a float pool, a per-channel u32 offset table, a per-channel u16 count table
// (1 = constant) and a u16 channel map.

import (
	"encoding/binary"
	"fmt"
	"math"
)

// SLSGroup is one rebuilt vertex region.
type SLSGroup struct {
	ChanBase   int
	ShapeCount int
	Verts      []uint16 // indices into the model's position array
}

// SLS is a parsed vertex-animation file.
type SLS struct {
	Groups []SLSGroup
	// ShapePos[g] is group g's first active shape decoded to vectors, one per
	// vertex — the pose the opening cutscenes show (their groups are all
	// single-shape, weight 1).
	ShapePos [][][3]float32
}

// ParseSLS reads a .sls file and decodes each group's first active shape.
func ParseSLS(b []byte) (*SLS, error) {
	if len(b) < 0x30 {
		return nil, fmt.Errorf("lm: sls too short")
	}
	u16 := func(o uint32) uint32 { return uint32(binary.BigEndian.Uint16(b[o:])) }
	u32 := func(o uint32) uint32 { return binary.BigEndian.Uint32(b[o:]) }
	f32 := func(o uint32) float32 { return math.Float32frombits(binary.BigEndian.Uint32(b[o:])) }
	groups := u16(0x08)
	rec, entries, dictA := u32(0x10), u32(0x14), u32(0x18)
	streams, active, idxl := u32(0x20), u32(0x28), u32(0x2C)
	if rec+groups*20 > uint32(len(b)) {
		return nil, fmt.Errorf("lm: sls group records out of range")
	}
	s := &SLS{}
	cursor := uint32(0)
	for g := uint32(0); g < groups; g++ {
		o := rec + g*20
		gr := SLSGroup{
			ChanBase:   int(u16(o + 4)),
			ShapeCount: int(u16(o + 16)),
		}
		nv := u16(o + 18)
		for k := uint32(0); k < nv; k++ {
			gr.Verts = append(gr.Verts, binary.BigEndian.Uint16(b[idxl+2*(cursor+k):]))
		}
		cursor += nv

		// The active-shape table names the entry of each of the group's
		// shapes; decode the first (the opening shots' only) one.
		entry := u16(active + 2*g)
		eo := entries + entry*8
		cnt, soff := u16(eo), u16(eo+2)
		if cnt != nv {
			return nil, fmt.Errorf("lm: sls group %d shape count %d != vertex count %d", g, cnt, nv)
		}
		shape := make([][3]float32, cnt)
		for k := uint32(0); k < cnt; k++ {
			w := binary.BigEndian.Uint16(b[streams+2*(soff+k):])
			di := uint32(w & 0x1FFF)
			v := [3]float32{f32(dictA + 12*di), f32(dictA + 12*di + 4), f32(dictA + 12*di + 8)}
			if w&0x8000 != 0 {
				v[0] = -v[0]
			}
			if w&0x4000 != 0 {
				v[1] = -v[1]
			}
			if w&0x2000 != 0 {
				v[2] = -v[2]
			}
			shape[k] = v
		}
		s.Groups = append(s.Groups, gr)
		s.ShapePos = append(s.ShapePos, shape)
	}
	return s, nil
}

// Apply overwrites the model's positions with each group's decoded first
// shape — what the game's evaluator produces for single-shape groups.
func (s *SLS) Apply(positions [][3]float32) {
	for g, gr := range s.Groups {
		for k, vi := range gr.Verts {
			if int(vi) < len(positions) {
				positions[vi] = s.ShapePos[g][k]
			}
		}
	}
}
