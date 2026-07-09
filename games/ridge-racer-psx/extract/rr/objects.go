package rr

// objects.go decodes the roadside object placements: where each OBJ.RRO
// scenery object (buildings, signs, the grandstand, roadside structures) sits
// in the world. The placements live in six static tables in the executable,
// drawn per frame by four near-identical iterators (0x800157D8, 0x800158E8,
// 0x800159F8, 0x80036778) and culled against the visible-cell mask; every
// table shares one 24-byte record:
//
//	+0   s16 id        ; OBJ.RRO object index; a negative id ends the table
//	+2   s16 —         ; 0
//	+4   s32 X         ; world position, quarter model units (×4 → model units)
//	+8   s32 Y
//	+12  s32 Z
//	+16  s32 —         ; 0 in the model lists; a per-list draw-mode flag otherwise
//	+20  s16 yaw       ; Y-axis rotation, 4096 = one turn (RotMatrix at 0x80017EAC)
//	+22  s16 —
//
// The iterator at 0x80036778 builds a Y-rotation from the yaw and translates
// the object to (X, Y, Z)×4; the object's own geometry (obj.go) is drawn under
// that transform. Two of the six tables hold the same 59 placements (a second
// draw pass with a mode flag), so placements are deduplicated by id+position.

import "math"

// Placement lists in the executable, drawn in the daytime race, all sharing
// the 24-byte record. Two dispatchers feed the four iterators (0x800157D8
// plain, 0x800158E8 translucent, 0x800159F8, 0x80036778):
//
//   - The roadside buildings (0x800129E0, mirrored at 0x80014F80): E85C (the
//     first-tunnel warning placards 191-194 — draw-only scenery; nothing else
//     ever reads or writes the table) and E904 (glow/flag panels 247-249)
//     always, then the halfword at 0x8017693C selects the day pair
//     (EAFC via 0x80036778, E9AC via 0x800158E8) or the night pair (F09C via
//     0x800159F8, EA54 via 0x800157D8) — the same buildings as different
//     object variants (e.g. 177 by day, 178 by night), so exporting both
//     z-fights.
//
//   - The structures head of the scenery dispatcher (0x80015B98), selected by
//     a scene mode: 0x70360 (start barriers) then 0x703C0 (grandstand 60, 61)
//     by day — at night 0x70468 replaces the barriers and the grandstand list
//     comes from a pointer variable (0x80079F98/9C -> 0x70408/0x70438, single
//     61-variant records); then the course-length word at 0x80176CBC picks
//     0x70510 (short courses) or 0x705E8 (long) for the course-split barrier
//     walls, and at night 0x706C0 adds translucent extras (101, 142, 165).
//
// The set below is exactly what the daytime race draws (verified against the
// drawers); the dynamic, code-placed objects are decoded by Dynamics()
// (dynamics.go).
var placementLists = []uint32{
	0x8006E85C, 0x8006E904, 0x8006EAFC, 0x8006E9AC, // buildings + placards (daytime)
	0x80070360, 0x800703C0, 0x80070510, // grandstand + barriers
}

const exeTextBase = 0x80010000

// Placement is one roadside object: an OBJ.RRO index, a world position in
// model units, a Y-axis rotation in turns/4096, and the source record's RAM
// address and draw-mode flag (the +16 word) for diagnostics.
type Placement struct {
	Obj     int
	X, Y, Z int32  // model units
	Yaw     int16  // 4096 = 360°
	Addr    uint32 // RAM address of the source record
	Flag    uint32 // the +16 word (a per-list draw-mode flag)
}

// Placements decodes every roadside object placement from the executable's
// text segment, deduplicated (the same object at the same position appears in
// more than one draw list). Dedup keys on id+position+yaw, keeping the first
// occurrence's address and flag.
func Placements(text []byte) []Placement {
	var out []Placement
	type key struct {
		id, yaw    int
		x, y, z    int32
	}
	seen := map[key]bool{}
	for _, head := range placementLists {
		off := int(head - exeTextBase)
		for off+24 <= len(text) {
			id := s16(text, off)
			if id < 0 {
				break
			}
			p := Placement{
				Obj:  int(id),
				X:    int32(u32(text, off+4)) * 4,
				Y:    int32(u32(text, off+8)) * 4,
				Z:    int32(u32(text, off+12)) * 4,
				Yaw:  s16(text, off+20),
				Addr: exeTextBase + uint32(off),
				Flag: u32(text, off+16),
			}
			k := key{p.Obj, int(p.Yaw), p.X, p.Y, p.Z}
			if !seen[k] {
				seen[k] = true
				out = append(out, p)
			}
			off += 24
		}
	}
	return out
}

// YawMatrix returns the Y-axis rotation the game builds from a yaw (the matrix
// at 0x80017EAC), as a 3×3 fixed-point matrix (4096 = 1.0):
//
//	[ cos  0  -sin ]
//	[  0  1.0   0  ]
//	[ sin  0   cos ]
func YawMatrix(yaw int16) [3][3]int32 {
	s := rsin(yaw)
	c := rcos(yaw)
	return [3][3]int32{
		{c, 0, -s},
		{0, 4096, 0},
		{s, 0, c},
	}
}

// PitchMatrix returns the X-axis rotation the game builds through 0x80017F18
// (libgte RotMatrixX layout), 4096 = 1.0:
//
//	[ 1.0   0     0  ]
//	[  0   cos  -sin ]
//	[  0   sin   cos ]
func PitchMatrix(pitch int16) [3][3]int32 {
	s := rsin(pitch)
	c := rcos(pitch)
	return [3][3]int32{
		{4096, 0, 0},
		{0, c, -s},
		{0, s, c},
	}
}

// rsin/rcos evaluate the PSX fixed-point sine the game's RotMatrix uses (4096 =
// one turn, output 4096 = 1.0). The exact table lives in the BIOS; a float
// bridge matches it to within a unit, immaterial to a placed object's facing.
func rsin(a int16) int32 { return int32(math.Round(math.Sin(float64(a)/4096*2*math.Pi) * 4096)) }
func rcos(a int16) int32 { return int32(math.Round(math.Cos(float64(a)/4096*2*math.Pi) * 4096)) }
