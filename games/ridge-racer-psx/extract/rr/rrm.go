package rr

// rrm.go decodes MAP.RRM, the track geometry. The file is a section count, a
// directory of three per-section record counts, and then every section's
// records back to back — each record a 40-byte textured quad in the section's
// local model space (the directory walker at 0x80011D64 computes the running
// offsets exactly this way; the renderer transforms and draws the records in
// place).
//
//	u16 count, u16 pad
//	count × { u16 nA, nB, nC, pad }   ; three draw classes per section
//	Σ(nA+nB+nC) × TrackQuad           ; 40 bytes each, grouped A then B then C
//
// TrackQuad (the in-file form of a GPU textured quad, traced at 0x800341EC
// and 0x800461E8):
//
//	+0  int16 x,y,z ×4      ; the four corners, local to the section cell
//	+24 u8 u0,v0  u16 clut
//	+28 u8 u1,v1  u16 tpage
//	+32 u8 u2,v2  int16 bias    ; ordering-table depth bias
//	+36 u8 u3,v3  u8 shade ×2   ; depth-cue shade pair
type TrackQuad struct {
	V     [4][3]int16
	UV    [4]UV
	CLUT  uint16
	TPage uint16
	Bias  int16
	Shade [2]uint8
}

// Section is one grid cell's geometry: three consecutively stored draw
// classes (the renderer picks per class how to shade and subdivide).
type Section struct {
	A, B, C []TrackQuad
}

// Track is the decoded MAP.RRM.
type Track struct {
	Sections []Section
}

// ParseRRM decodes a MAP.RRM image.
func ParseRRM(d []byte) (*Track, error) {
	if len(d) < 4 {
		return nil, errTruncated("rrm", 4, len(d))
	}
	count := int(u16(d, 0))
	dir := 4
	geo := dir + count*8
	if geo > len(d) {
		return nil, errTruncated("rrm", geo, len(d))
	}
	t := &Track{Sections: make([]Section, count)}
	off := geo
	for i := 0; i < count; i++ {
		nA := int(u16(d, dir+i*8))
		nB := int(u16(d, dir+i*8+2))
		nC := int(u16(d, dir+i*8+4))
		need := off + (nA+nB+nC)*40
		if need > len(d) {
			return nil, errTruncated("rrm", need, len(d))
		}
		s := &t.Sections[i]
		s.A, off = trackQuads(d, off, nA)
		s.B, off = trackQuads(d, off, nB)
		s.C, off = trackQuads(d, off, nC)
	}
	return t, nil
}

func trackQuads(d []byte, off, n int) ([]TrackQuad, int) {
	qs := make([]TrackQuad, n)
	for i := range qs {
		qs[i] = trackQuadAt(d, off)
		off += 40
	}
	return qs, off
}

func trackQuadAt(d []byte, o int) TrackQuad {
	var q TrackQuad
	for v := 0; v < 4; v++ {
		q.V[v] = [3]int16{s16(d, o+v*6), s16(d, o+v*6+2), s16(d, o+v*6+4)}
	}
	q.UV[0] = uvOf(u16(d, o+24))
	q.CLUT = u16(d, o+26)
	q.UV[1] = uvOf(u16(d, o+28))
	q.TPage = u16(d, o+30)
	q.UV[2] = uvOf(u16(d, o+32))
	q.Bias = s16(d, o+34)
	q.UV[3] = uvOf(u16(d, o+36))
	q.Shade = [2]uint8{d[o+38], d[o+39]}
	return q
}
