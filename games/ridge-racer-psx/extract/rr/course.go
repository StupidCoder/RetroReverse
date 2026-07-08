package rr

// course.go decodes the executable's course tables: the 256-entry checkpoint
// table that defines course progress, and the scenery-set triggers.
//
// The car's course progress is segment*256 + fraction, where the segment is
// the index of the checkpoint gate the car is inside (the per-frame search at
// 0x8001B68C over the table at 0x80059164, 20 bytes per entry). A checkpoint's
// position decodes as X = 0xF000 - (word0 >> 14), Z = word4 >> 14 in position
// units (quarter model units); halfword +10 is the gate heading, +14/+16 the
// half-widths.
//
// The race scenery quadrant (vram.go) holds one of three texture sets,
// switched by course progress (selector at 0x800374F0, initialised at
// 0x80012D60): the city set (0) while progress is between trigger A = 0xD00
// and trigger B = 0x6800, the seaside set (1) on the rest of the lap, with a
// ±512-unit row-by-row crossfade around each trigger. The third set is
// requested explicitly by attract-demo scenes and course variants, never by
// lap position on the standard course.

import "fmt"

const (
	checkpointTableOff = 0x80059164 - 0x80010000 // in the text segment
	checkpointCount    = 256
	courseOriginX      = 0xF000 // ori at 0x8001547C

	// Scenery-set triggers in progress units (immediates at 0x80012D78 and
	// 0x80012DA8; the alternate course variant uses 0x6400 for B).
	TriggerCityStart = 0x0D00
	TriggerCityEnd   = 0x6800
)

// Checkpoint is one course-progress gate, in position units (model units /4).
type Checkpoint struct {
	X, Z    int32
	Heading int16
}

// Checkpoints decodes the course checkpoint table from the text segment.
func Checkpoints(text []byte) ([]Checkpoint, error) {
	if len(text) < checkpointTableOff+checkpointCount*20 {
		return nil, fmt.Errorf("text segment too short for the checkpoint table")
	}
	cps := make([]Checkpoint, checkpointCount)
	for i := range cps {
		e := checkpointTableOff + i * 20
		x := int32(u32(text, e)) >> 14
		z := int32(u32(text, e+4)) >> 14
		cps[i] = Checkpoint{
			X:       courseOriginX - x,
			Z:       z,
			Heading: int16(u16(text, e+10)),
		}
	}
	return cps, nil
}

// Scenery set indices, in the order the boot streams bank them.
const (
	SetCity    = 0 // quadrant as of the end of TEX1's stream
	SetSeaside = 1 // ... of TEX2's stream
	SetThird   = 2 // ... of TEX3's stream (demo scenes / variants)
)

// SetForProgress returns the scenery set active at a course progress value
// (segment*256), per the traced selector: city between the triggers, seaside
// on the rest of the lap. Positions inside a crossfade window resolve to the
// nearer side.
func SetForProgress(progress int32) int {
	p := ((progress % 0x10000) + 0x10000) % 0x10000
	mid := func(a, b int32) bool { return p >= a && p < b }
	if mid(TriggerCityStart, TriggerCityEnd) {
		return SetCity
	}
	return SetSeaside
}

// NearestSegment returns the checkpoint index closest to a point given in
// model units.
func NearestSegment(cps []Checkpoint, x, z int32) int {
	px, pz := x/4, z/4 // model units → position units
	best, bestD := 0, int64(1)<<62
	for i, c := range cps {
		dx, dz := int64(px-c.X), int64(pz-c.Z)
		if d := dx*dx + dz*dz; d < bestD {
			best, bestD = i, d
		}
	}
	return best
}
