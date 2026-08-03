// objdump lists what a room places: the entities (enemies, objects, NPCs,
// managers) and the slot-3 command records (song choice, flag-gated metatile
// patches, positioned manager spawns).
//
// The record model — the eight per-room slots, the 16-byte entity record, the
// 8-byte command record, and the per-class behaviour dispatch tables — is
// documented in objects.go, all of it read out of the game's own loader chain
// ($0804ADDC/$0804ADF8/$0804B058/$0804B1AC).
//
//	objdump -area 3 -room 1        # one room
//	objdump -area 3                # every room of an area
//	objdump -all -json FILE        # every record in the cartridge
//	objdump -count                 # totals
//	objdump -handlers              # objects grouped by behaviour handler
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "objdump: "+format+"\n", a...)
	os.Exit(2)
}

// walk visits every room of every area, bounded by each area's own geometry
// table — the object table has no length of its own.
func walk(rom []byte, fn func(area, room int)) {
	for a := 0; a < 160; a++ {
		ar, err := LoadArea(rom, a)
		if err != nil {
			continue
		}
		for r := 0; r < len(ar.Rooms); r++ {
			fn(a, r)
		}
	}
}

func roomRecords(rom []byte, a, r int) ([]Object, []Command) {
	var objs []Object
	for slot := 0; slot < 3; slot++ {
		l := ListFor(rom, a, r, slot)
		if l == 0 {
			continue
		}
		for _, o := range Objects(rom, l, slot) {
			o.Area, o.Room = a, r
			objs = append(objs, o)
		}
	}
	var cmds []Command
	if l := ListFor(rom, a, r, 3); l != 0 {
		for _, c := range Commands(rom, l) {
			c.Area, c.Room = a, r
			cmds = append(cmds, c)
		}
	}
	return objs, cmds
}

func main() {
	romPath := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	area := flag.Int("area", 3, "area id")
	room := flag.Int("room", -1, "room index (-1 = every room of the area)")
	count := flag.Bool("count", false, "count the records in the whole cartridge and exit")
	all := flag.Bool("all", false, "dump every record in the cartridge")
	handlers := flag.Bool("handlers", false, "group entities by behaviour handler")
	jsonOut := flag.String("json", "", "write the records as JSON")
	flag.Parse()

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		die("%v", err)
	}

	if *count {
		objs, cmds, rooms := 0, 0, 0
		walk(rom, func(a, r int) {
			o, c := roomRecords(rom, a, r)
			objs += len(o)
			cmds += len(c)
			if len(o)+len(c) > 0 {
				rooms++
			}
		})
		fmt.Printf("%d entities and %d commands across %d rooms\n", objs, cmds, rooms)
		return
	}

	if *handlers {
		type key struct{ class, id int }
		byHandler := map[uint32][]key{}
		counts := map[key]int{}
		walk(rom, func(a, r int) {
			o, _ := roomRecords(rom, a, r)
			for _, ob := range o {
				k := key{ob.Class, ob.ID}
				if counts[k] == 0 {
					byHandler[ob.Handler] = append(byHandler[ob.Handler], k)
				}
				counts[k]++
			}
		})
		type row struct {
			h    uint32
			keys []key
			n    int
		}
		var rows []row
		for h, ks := range byHandler {
			n := 0
			for _, k := range ks {
				n += counts[k]
			}
			rows = append(rows, row{h, ks, n})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].n > rows[j].n })
		for _, r := range rows {
			fmt.Printf("handler %08X: %5d records ", r.h, r.n)
			for _, k := range r.keys {
				fmt.Printf(" %s/%02X(%d)", ClassNames[k.class], k.id, counts[k])
			}
			fmt.Println()
		}
		return
	}

	if *all {
		var objs []Object
		var cmds []Command
		walk(rom, func(a, r int) {
			o, c := roomRecords(rom, a, r)
			objs = append(objs, o...)
			cmds = append(cmds, c...)
		})
		fmt.Printf("%d entities, %d commands\n", len(objs), len(cmds))
		if *jsonOut != "" {
			b, _ := json.Marshal(map[string]any{"objects": objs, "commands": cmds})
			if err := os.WriteFile(*jsonOut, b, 0o644); err != nil {
				die("%v", err)
			}
			fmt.Println("wrote", *jsonOut)
		}
		return
	}

	rooms := []int{*room}
	if *room < 0 {
		ar, err := LoadArea(rom, *area)
		if err != nil {
			die("%v", err)
		}
		rooms = nil
		for r := 0; r < len(ar.Rooms); r++ {
			rooms = append(rooms, r)
		}
	}
	var allObjs []Object
	for _, r := range rooms {
		objs, cmds := roomRecords(rom, *area, r)
		if len(objs)+len(cmds) == 0 {
			continue
		}
		fmt.Printf("area %d room %d: %d entities, %d commands\n", *area, r, len(objs), len(cmds))
		for _, o := range objs {
			pos := fmt.Sprintf("(%4d,%4d)", o.X, o.Y)
			if !o.Placed() {
				pos = "  unplaced "
			}
			kill := ""
			if o.KillIdx >= 0 {
				kill = fmt.Sprintf(" kill#%d", o.KillIdx)
			}
			fmt.Printf("   %s slot %d  %-7s id %02X/%02X mode %X %s handler %08X extra %08X%s\n",
				o.Addr, o.Slot, ClassNames[o.Class], o.ID, o.Sub, o.Mode, pos, o.Handler, o.Extra, kill)
		}
		for _, c := range cmds {
			x, y, ok := c.Pos()
			pos := "           "
			if ok {
				pos = fmt.Sprintf("(%4d,%4d)", x, y)
			}
			fmt.Printf("   %s cmd %2d  b %02X %02X %02X  h %04X %04X %s\n",
				c.Addr, c.Op, c.B1, c.B2, c.B3, c.H4, c.H6, pos)
		}
		allObjs = append(allObjs, objs...)
	}
	if *jsonOut != "" {
		b, _ := json.MarshalIndent(allObjs, "", " ")
		if err := os.WriteFile(*jsonOut, b, 0o644); err != nil {
			die("%v", err)
		}
		fmt.Println("wrote", *jsonOut)
	}
}
