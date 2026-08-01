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

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

const objTable = 0x080D50FC

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "objdump: "+format+"\n", a...)
	os.Exit(2)
}

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

func main() {
	romPath := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	area := flag.Int("area", 3, "area id")
	room := flag.Int("room", -1, "room index (-1 = every room of the area)")
	kinds := flag.Int("kinds", 4, "how many kind slots to read per room")
	count := flag.Bool("count", false, "count the objects in the whole cartridge and exit")
	jsonOut := flag.String("json", "", "write the objects as JSON")
	flag.Parse()

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		die("%v", err)
	}

	if *count {
		// Bound the room index by the area's OWN geometry table. The object
		// table has no length of its own, so reading a fixed 64 rooms per area
		// walks off the end into whatever follows and counts garbage as objects
		// — 158,658 of them, against 419 rooms that actually exist.
		total, rooms, areas := 0, 0, 0
		for a := 0; a < 160; a++ {
			ar, err := LoadArea(rom, a)
			if err != nil {
				continue
			}
			any := false
			for r := 0; r < len(ar.Rooms); r++ {
				had := false
				for k := 0; k < *kinds; k++ {
					if l := ListFor(rom, a, r, k); l != 0 {
						n := len(Objects(rom, l, k))
						total += n
						had = had || n > 0
					}
				}
				if had {
					rooms++
					any = true
				}
			}
			if any {
				areas++
			}
		}
		fmt.Printf("%d objects across %d rooms in %d areas\n", total, rooms, areas)
		return
	}

	var all []Object
	rooms := []int{*room}
	if *room < 0 {
		rooms = nil
		for r := 0; r < 64; r++ {
			rooms = append(rooms, r)
		}
	}
	for _, r := range rooms {
		var got []Object
		for k := 0; k < *kinds; k++ {
			l := ListFor(rom, *area, r, k)
			if l == 0 {
				continue
			}
			got = append(got, Objects(rom, l, k)...)
		}
		if len(got) == 0 {
			continue
		}
		fmt.Printf("area %d room %d: %d objects\n", *area, r, len(got))
		for _, o := range got {
			fmt.Printf("   %s kind %d  type %02X/%02X param %5d  pos (%4d,%4d)  extra %s\n",
				o.Addr, o.Kind, o.Type, o.Subtype, o.Param, o.X, o.Y, o.Extra)
		}
		all = append(all, got...)
	}
	if *jsonOut != "" {
		b, _ := json.MarshalIndent(all, "", " ")
		if err := os.WriteFile(*jsonOut, b, 0o644); err != nil {
			die("%v", err)
		}
		fmt.Println("wrote", *jsonOut)
	}
}
