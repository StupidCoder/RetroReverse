package nfs

import "fmt"

// The slice GEOMETRY TOPOLOGY is static data inside LaunchMe, shared by all
// nine courses — only the cross-section vertex positions stream from the .trk.
// Three tables (addresses are LaunchMe file offsets = load addresses):
//
//	0x3F2F0  slice-type selector: 128-byte rows indexed by Segment.SliceFamily;
//	         each row is 32 u32 slice-type ids indexed by distance-from-camera
//	         in groups (a draw-LOD ramp — entry 0 is the full-detail type).
//	0x41164  slice-type records, 76 bytes:
//	         +0x00 u32 total vertex count over the group's 4 rows
//	         +0x04 u32 face count
//	         +0x08 4×u32 packed per-LOD face ranges (unused here)
//	         +0x18 4 × 0xFF-terminated byte lists: per-row vertex subsets
//	         +0x48 u32 face-list start index
//	0x402BC  face records, 8 bytes: v0,v1,v2,v3 (0-10 = row s, 11-21 = row
//	         s+1), material (indexes the group header's material bytes),
//	         b5 (backface-test arg), texture sub-index, b7 (effect count).
//
// All three were traced from the world-vertex builder 0x15F70 (type selection
// + vertex lists) and the per-segment face scan 0x1A7DC (faces + materials).
const (
	lmSelectorOff  = 0x3F2F0
	lmSliceTypeOff = 0x41164
	lmFaceListOff  = 0x402BC
	sliceTypeSize  = 76
	selectorStride = 128
)

// SliceFace is one 8-byte static face record.
type SliceFace struct {
	V        [4]int // 0-10 = this row's points, 11-21 = next row's
	Material byte
	B5       byte
	TexSub   byte
	B7       byte
}

// SliceType is one 76-byte slice-type record.
type SliceType struct {
	ID        int
	VertTotal int
	FaceStart int
	FaceCount int
	RowVerts  [4][]int // per-row vertex-index subsets (from the byte lists)
}

// SliceTables is the static topology extracted from LaunchMe.
type SliceTables struct {
	lm    []byte
	types map[int]*SliceType
}

// LoadSliceTables wraps a LaunchMe image for slice-topology lookups.
func LoadSliceTables(launchMe []byte) (*SliceTables, error) {
	if len(launchMe) < lmSliceTypeOff+64*sliceTypeSize {
		return nil, fmt.Errorf("nfs: LaunchMe image too short")
	}
	return &SliceTables{lm: launchMe, types: map[int]*SliceType{}}, nil
}

// TypeFor returns the full-detail slice type for a segment's family byte.
func (st *SliceTables) TypeFor(family byte) (*SliceType, error) {
	id := int(be32(st.lm[lmSelectorOff+int(family)*selectorStride:]))
	return st.Type(id)
}

// Type decodes slice-type record id.
func (st *SliceTables) Type(id int) (*SliceType, error) {
	if t, ok := st.types[id]; ok {
		return t, nil
	}
	if id < 0 || id > 255 {
		return nil, fmt.Errorf("nfs: implausible slice type %d", id)
	}
	o := lmSliceTypeOff + id*sliceTypeSize
	t := &SliceType{
		ID:        id,
		VertTotal: int(be32(st.lm[o:])),
		FaceCount: int(be32(st.lm[o+4:])),
		FaceStart: int(be32(st.lm[o+0x48:])),
	}
	p := o + 0x18
	for row := 0; row < 4; row++ {
		var list []int
		for p < o+0x48 && st.lm[p] != 0xFF {
			list = append(list, int(st.lm[p]))
			p++
		}
		p++ // consume the 0xFF
		t.RowVerts[row] = list
	}
	if t.VertTotal > 4*2*rowPoints || t.FaceCount > 256 {
		return nil, fmt.Errorf("nfs: slice type %d looks corrupt (%d verts, %d faces)", id, t.VertTotal, t.FaceCount)
	}
	st.types[id] = t
	return t, nil
}

// Faces returns the face records [start, start+count).
func (st *SliceTables) Faces(t *SliceType) []SliceFace {
	out := make([]SliceFace, 0, t.FaceCount)
	for i := 0; i < t.FaceCount; i++ {
		o := lmFaceListOff + 8*(t.FaceStart+i)
		f := SliceFace{
			Material: st.lm[o+4],
			B5:       st.lm[o+5],
			TexSub:   st.lm[o+6],
			B7:       st.lm[o+7],
		}
		for k := 0; k < 4; k++ {
			f.V[k] = int(st.lm[o+k])
		}
		out = append(out, f)
	}
	return out
}
