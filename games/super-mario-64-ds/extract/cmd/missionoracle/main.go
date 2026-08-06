// missionoracle answers "does this object exist in this mission?" by RUNNING
// the object's own create+init under a seeded progress context, instead of
// decoding fifty bespoke conditions. The sweep itself is sm64ds.MissionGates —
// read its doc comment for the method; this is the standalone driver and the
// instrument's own mutation test.
//
//	missionoracle [-rom img] [-extracted dir] [-o gates.json]
//	missionoracle -probe     # the Whomp's Fortress pair must swap over, and the
//	                         # tower's refusal must depend on its own height
//
// webexport runs the same pass in-process, so this exists for inspection.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

// binding mirrors the actor oracle's table entry — the record of which overlay
// each actor's profile is meaningful under.
type binding struct {
	Params [3]int `json:"params"`
	Config int    `json:"config"`
}

var bindings map[string][]binding

// configFor picks the config for one placement: the binding table's entry for
// these exact params, else its entry for the same par1, else its first.
func configFor(actor int, par [3]int) (int, bool) {
	bs := bindings[fmt.Sprint(actor)]
	if len(bs) == 0 {
		return -1, false
	}
	for _, b := range bs {
		if b.Params == par {
			return b.Config, true
		}
	}
	for _, b := range bs {
		if b.Params[0] == par[0] {
			return b.Config, true
		}
	}
	return bs[0].Config, true
}

func main() {
	rom := flag.String("rom", "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "extracted", "extracted binaries dir")
	bind := flag.String("bindings", "extracted/actorbind.json", "actor oracle binding table")
	out := flag.String("o", "extracted/missiongates.json", "output table")
	probe := flag.Bool("probe", false, "mutation-test the instrument on the Whomp's Fortress pair")
	verbose := flag.Bool("v", false, "trace")
	flag.Parse()

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		sm64ds.Die(err)
	}
	o, err := sm64ds.NewOracle(ls)
	if err != nil {
		sm64ds.Die(err)
	}
	if *verbose {
		o.Trace = func(s string) { fmt.Fprintln(os.Stderr, "  |", s) }
	}
	if err := o.InitEngine(); err != nil {
		sm64ds.Die(err)
	}
	buf, err := os.ReadFile(*bind)
	if err != nil {
		sm64ds.Die(err)
	}
	if err := json.Unmarshal(buf, &bindings); err != nil {
		sm64ds.Die(err)
	}

	if *probe {
		runProbe(o)
		return
	}
	gates := sm64ds.MissionGates(ls, o, configFor, func(stem string, n int) {
		fmt.Printf("[gates] %-24s %3d placements gated\n", stem, n)
	})
	js, _ := json.MarshalIndent(gates, "", " ")
	if err := os.WriteFile(*out, js, 0o644); err != nil {
		sm64ds.Die(err)
	}
	fmt.Printf("wrote %s (%d stages)\n", *out, len(gates))
}

// runProbe is the instrument's mutation test. The Whomp's Fortress summit pair
// must come out as exact complements, and moving the tower off the summit must
// remove its gate — the height is part of its condition, so if the position we
// publish never reached the actor, the third row would not differ from the
// second.
func runProbe(o *sm64ds.Oracle) {
	const level, course = 7, 1
	cases := []struct {
		name    string
		actor   int
		par     [3]int
		x, y, z int
	}{
		{"Whomp King", 165, [3]int{0xFF01, 0, 0}, 704, 3500, 705},
		{"tower (summit, y=3500)", 49, [3]int{0xFFFF, 0, 0}, 0, 3500, 0},
		{"tower (moved down, y=0)", 49, [3]int{0xFFFF, 0, 0}, 0, 0, 0},
	}
	fmt.Printf("Whomp's Fortress (level %d, course %d)\n", level, course)
	fmt.Printf("%-26s %s\n", "", "star 1 2 3 4 5 6 7")
	for _, c := range cases {
		fmt.Printf("%-26s      ", c.name)
		for star := 1; star <= sm64ds.StarsPerCourse; star++ {
			cfg, ok := configFor(c.actor, c.par)
			if !ok {
				sm64ds.Die(fmt.Errorf("actor %d not in the binding table", c.actor))
			}
			o.SetMission(level, course, star, sm64ds.FirstPlaySave(star))
			o.SetSpawnPos(int32(c.x)<<12, int32(c.y)<<12, int32(c.z)<<12)
			if err := o.LoadConfig(cfg); err != nil {
				sm64ds.Die(err)
			}
			run := o.RunActor(c.actor, cfg, c.par)
			mark := "."
			if !run.Refused && !run.Destroyed {
				mark = "#"
			}
			fmt.Printf("%s ", mark)
		}
		fmt.Println()
	}
	fmt.Println("  # = the actor's own code let it exist,  . = it refused or destroyed itself")
}
