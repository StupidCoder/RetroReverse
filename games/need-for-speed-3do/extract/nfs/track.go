// Package nfs decodes The Need for Speed's own data formats — the .trk track
// files, the RoadObjects placement block, the static slice-geometry tables in
// LaunchMe, and the car .wrapFam texture binding. Everything here was derived
// by tracing the game's own consumers under the oracle (see the game doc,
// "Part X — the track geometry"): the segment array pointer at [0x4CCEC], the
// world-vertex builder at 0x15F70, the slice-face scan at 0x1A76C, the
// RoadObjects draw at 0x1A2E8 and the car texset builder at 0x24800.
package nfs

import (
	"encoding/binary"
	"fmt"
)

func be32(b []byte) uint32  { return binary.BigEndian.Uint32(b) }
func sbe32(b []byte) int32  { return int32(binary.BigEndian.Uint32(b)) }
func be16(b []byte) uint16  { return binary.BigEndian.Uint16(b) }
func sbe16(b []byte) int16  { return int16(binary.BigEndian.Uint16(b)) }
func s24(w uint32) int      { return int(int32(w<<8) >> 8) } // signed low 24 bits

// Fixed 16.16 world coordinate.
type Fx16 = int32

// Vec3 is a world-space point in 16.16 fixed point, exactly as stored on disc.
type Vec3 struct{ X, Y, Z Fx16 }

// Float converts a 16.16 coordinate to float64 world units.
func Float(v Fx16) float64 { return float64(v) / 65536.0 }

// Track file layout (all offsets confirmed against the loader at 0x158D8):
//
//	head block: 0x16C44 bytes, read whole into the "RoadSection" allocation.
//	  +0x00  u32 ??? (14)
//	  +0x04  u32 group count (number of streamed TRKD chunks)
//	  +0x2C  u32[600]  per-group destination offsets in the slice ring arena
//	  +0x98C u32[...]  per-group file offsets of the TRKD chunks
//	  +0x13B4 segment records, 0x24 bytes × 2400 ([0x4CCEC] points here +0x13B4)
//	  +0x16534 per-segment table (2 × 2400 bytes region, purpose uncharted)
//	  +0x16C3C u32 object-definition count
//	  +0x16C40 u32 object-placement count
//	RoadObjects block: at file 0x16C44 {u32 ?, u32 size, u32 ?} then payload:
//	  defs (16 B × defCount), placements (16 B × placementCount) at +defCount*16.
//	TRKD chunks: at the per-group file offsets. Each chunk is a pre-baked
//	  sliding window: groups g..g+12 (13 × 0x258), then 5 × 0x9C near-window
//	  records at +0x1E78. Only the FIRST group of each chunk is canonical for
//	  extraction; the rest repeat neighbouring chunks' data.
//	Group block (0x258): 0x18 header (material bytes at [2..14]) then 4 rows.
//	Row (0x90): 12-byte row header, then 11 × 12-byte world-space (x,y,z)
//	  16.16 vertices — the slice cross-section, stored in absolute world space.
const (
	headSize       = 0x16C44
	segArrayOff    = 0x13B4
	segCount       = 2400
	segStride      = 0x24
	groupCountOff  = 0x04
	fileTableOff   = 0x98C
	defCountOff    = 0x16C3C
	placeCountOff  = 0x16C40
	groupSize      = 0x258
	rowSize        = 0x90
	rowHeaderSize  = 0xC
	rowPoints      = 11
	groupHdrSize   = 0x18
	segsPerGroup   = 4
)

// Segment is one 0x24-byte road-spline record. Pos is the segment's point on
// the track centreline; Heading is the yaw angle in the game's units (a full
// circle is 0x1000000 after the <<10 the consumers apply, so heading<<10 =
// angle, 0x400000 = 90°). SliceFamily (byte +7) selects the slice-type row of
// the static selector table; the width bytes are at +2/+3.
type Segment struct {
	Pos         Vec3
	Heading     int16 // bytes +0x18/+0x19, big-endian signed
	SliceFamily byte  // byte +0x7
	WidthL      byte  // byte +0x2 (read by the lane code at 0x28B70)
	WidthR      byte  // byte +0x3
	Raw         [segStride]byte
}

// Group is one 4-segment streaming group: the material remap bytes from its
// header plus the 4 slice cross-section rows (11 world-space points each).
type Group struct {
	Materials [12]byte // header bytes [2..14): face material -> texture group
	Rows      [segsPerGroup][rowPoints]Vec3
}

// Track is a fully decoded .trk file.
type Track struct {
	Segments []Segment
	Groups   []Group // one per streamed chunk; Groups[g] covers segments 4g..4g+3
	Objects  RoadObjects
}

// UsedSegments is the number of segments covered by streamed geometry
// (4 × len(Groups)); the segment array is padded beyond it.
func (t *Track) UsedSegments() int { return len(t.Groups) * segsPerGroup }

// ParseTrack decodes a whole .trk file.
func ParseTrack(trk []byte) (*Track, error) {
	if len(trk) < headSize {
		return nil, fmt.Errorf("nfs: trk file too short (%d bytes)", len(trk))
	}
	t := &Track{}

	for i := 0; i < segCount; i++ {
		o := segArrayOff + i*segStride
		var s Segment
		copy(s.Raw[:], trk[o:o+segStride])
		s.Pos = Vec3{sbe32(trk[o+0x8:]), sbe32(trk[o+0xC:]), sbe32(trk[o+0x10:])}
		s.Heading = sbe16(trk[o+0x18:])
		s.SliceFamily = trk[o+0x7]
		s.WidthL, s.WidthR = trk[o+0x2], trk[o+0x3]
		t.Segments = append(t.Segments, s)
	}

	nGroups := int(be32(trk[groupCountOff:]))
	if nGroups <= 0 || nGroups > 600 {
		return nil, fmt.Errorf("nfs: implausible group count %d", nGroups)
	}
	for g := 0; g < nGroups; g++ {
		foff := int(be32(trk[fileTableOff+4*g:]))
		if foff == 0 || foff+12+groupSize > len(trk) {
			return nil, fmt.Errorf("nfs: group %d file offset 0x%X out of range", g, foff)
		}
		if string(trk[foff:foff+4]) != "TRKD" {
			return nil, fmt.Errorf("nfs: group %d at 0x%X: no TRKD magic", g, foff)
		}
		// chunk header {magic, u32 size, u32 group index}; payload holds the
		// 13-group window — group g itself is the first block.
		if idx := int(be32(trk[foff+8:])); idx != g {
			return nil, fmt.Errorf("nfs: chunk at 0x%X says group %d, want %d", foff, idx, g)
		}
		p := foff + 12
		var grp Group
		copy(grp.Materials[:], trk[p+2:p+14])
		for r := 0; r < segsPerGroup; r++ {
			ro := p + groupHdrSize + r*rowSize + rowHeaderSize
			for k := 0; k < rowPoints; k++ {
				po := ro + k*12
				grp.Rows[r][k] = Vec3{sbe32(trk[po:]), sbe32(trk[po+4:]), sbe32(trk[po+8:])}
			}
		}
		t.Groups = append(t.Groups, grp)
	}

	obj, err := parseRoadObjects(trk)
	if err != nil {
		return nil, err
	}
	t.Objects = *obj
	return t, nil
}

// RowPoints is the number of cross-section points per slice row.
const RowPoints = rowPoints
