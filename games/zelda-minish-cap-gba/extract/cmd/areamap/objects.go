// objdump lists the objects a room places: chests, pots, signs, doors, enemies
// and everything else the engine spawns when the player walks in.
//
// The lookup is three levels deep, which is why it resists a flat search:
//
//	$080D50FC[area] -> a per-room array
//	            [room] -> a per-KIND array
//	              [kind] -> the object list itself
//
// A list is 16-byte records terminated by a $FF type byte. The spawner
// ($0804AF68) reads three fields out of each record and hands them to the
// entity constructor:
//
//	+0x00  type, +0x01 subtype
//	+0x02  parameter (u16) — meaning depends on the type
//	+0x06  two bytes, a size or a pair of flags depending on the kind
//	+0x08  x, +0x0A y — RELATIVE TO THE ROOM, added to the room's origin
//	+0x0C  for spawned objects, the handler; for others, a second pair of u16s
//
// Positions are room-relative, so adding the room's own (x, y) from the area
// geometry table puts every object on the same canvas the maps are drawn on.
//
//	objdump -area 3 -room 1        # one room
//	objdump -area 3                # every room of an area
//	objdump -count                 # how many objects the cartridge holds
package main

import "fmt"

const objTable = 0x080D50FC

// Object is one 16-byte record.
type Object struct {
	Addr    string `json:"addr"`
	Kind    int    `json:"kind"`
	Type    int    `json:"type"`
	Subtype int    `json:"subtype"`
	Param   int    `json:"param"`
	X       int    `json:"x"` // room-relative
	Y       int    `json:"y"`
	Extra   string `json:"extra"` // the +0x0C word, unresolved
}

// ListFor resolves table[area][room][kind].
func ListFor(rom []byte, area, room, kind int) uint32 {
	arr := rd32(rom, uint32(objTable+area*4))
	if arr>>24 != 0x08 {
		return 0
	}
	per := rd32(rom, arr+uint32(room*4))
	if per>>24 != 0x08 {
		return 0
	}
	l := rd32(rom, per+uint32(kind*4))
	if l>>24 != 0x08 {
		return 0
	}
	return l
}

// Objects walks one list to its $FF terminator.
func Objects(rom []byte, list uint32, kind int) []Object {
	var out []Object
	for i := 0; i < 256; i++ {
		a := list + uint32(i*16)
		o := int(a & 0x01FFFFFF)
		if o+16 > len(rom) || rom[o] == 0xFF {
			break
		}
		out = append(out, Object{
			Addr: fmt.Sprintf("%08X", a), Kind: kind,
			Type: int(rom[o]), Subtype: int(rom[o+1]),
			Param: int(rd16(rom, a+2)),
			X:     int(rd16(rom, a+8)), Y: int(rd16(rom, a+10)),
			Extra: fmt.Sprintf("%08X", rd32(rom, a+12)),
		})
	}
	return out
}
