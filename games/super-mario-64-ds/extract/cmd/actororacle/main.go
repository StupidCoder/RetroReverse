// actororacle reports actor -> model/clip bindings by RUNNING each actor's
// create+init code on the tools/arm CPU (see sm64ds/oracle.go). It boots the
// ARM9 to main(), loads overlay 0 + overlay 2 + one bank/level overlay at a
// time, runs the overlay static initializers natively, and traps the file
// loader ($0201818C) so every model/animation the actor's own code requests is
// recorded — the heuristic-free replacement for the static binding scan.
//
//	actororacle [-rom img] [-extracted dir] [-v] [-boot] [-actor N -ovl X -par P]
//	            [-o bindings.json]
//
// Default: runs every distinct (actor, params) combination that the levels
// actually place, under the level's overlay (engine/level actors) or under
// each enemy bank that carries the actor's profile (bank actors), and writes
// the merged binding table exportlevelobjs consumes.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

func main() {
	rom := flag.String("rom", "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "../extracted", "extracted binaries dir")
	verbose := flag.Bool("v", false, "trace progress")
	bootOnly := flag.Bool("boot", false, "boot + engine init only (bring-up)")
	oneActor := flag.Int("actor", -1, "run a single actor")
	oneOvl := flag.Int("ovl", -1, "overlay config for -actor")
	onePar := flag.Int("par", 0, "par1 for -actor")
	outPath := flag.String("o", "../extracted/actorbind.json", "binding table output")
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
	fmt.Printf("booted + engine overlays initialized (%d init-phase file requests)\n", len(o.InitRequests()))
	for _, r := range o.InitRequests() {
		fmt.Printf("  init load: id=%-6d %-40s (%s)\n", r.ID, r.Name, r.Phase)
	}
	if *bootOnly {
		return
	}

	if *oneActor >= 0 {
		if err := o.LoadConfig(*oneOvl); err != nil {
			sm64ds.Die(err)
		}
		run := o.RunActor(*oneActor, *oneOvl, [3]int{*onePar, 0, 0})
		dump(o, run)
		return
	}

	table := sweep(ls, o, *verbose)
	buf, _ := json.MarshalIndent(table, "", " ")
	if err := os.WriteFile(*outPath, buf, 0o644); err != nil {
		sm64ds.Die(err)
	}
	fmt.Printf("wrote %s (%d actors bound)\n", *outPath, len(table))
}

func dump(o *sm64ds.Oracle, run *sm64ds.ActorRun) {
	fmt.Printf("actor %d cfg %d par %v: create=%08X obj=%08X\n", run.Actor, run.Config, run.Params, run.Create, run.Obj)
	for _, f := range run.Files {
		tag := f.Kind
		if tag == "" {
			tag = "load"
		}
		fmt.Printf("  %-7s %-7s id=%-6d %s\n", f.Phase, tag, f.ID, f.Name)
	}
	for _, n := range run.Notes {
		fmt.Printf("  note: %s\n", n)
	}
	fmt.Printf("  models: %v  clips: %v  kcl: %v\n", o.Models(run), o.Clips(run), o.KCLs(run))
	for _, c := range run.Colliders {
		fmt.Printf("  collider: %s (%s) mtx=%v scaleY=%d\n", c.KCL, c.Class, c.Mtx, c.ScaleY)
	}
}

// Binding is the oracle's answer for one actor under one parameter set.
type Binding struct {
	Params    [3]int            `json:"params"`
	Config    int               `json:"config"`
	Models    []string          `json:"models,omitempty"`
	Clips     []string          `json:"clips,omitempty"`
	KCL       []string          `json:"kcl,omitempty"`
	Colliders []sm64ds.Collider `json:"colliders,omitempty"`
	Notes     []string          `json:"notes,omitempty"`
}

// sweep runs every distinct (actor, params) the levels place.
func sweep(ls *sm64ds.LevelSet, o *sm64ds.Oracle, verbose bool) map[int][]Binding {
	type combo struct {
		actor, p1, p2, p3 int
	}
	// distinct combos per level overlay
	perLevel := map[int]map[combo]bool{}
	all := map[combo][]int{} // combo -> level overlays placing it
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			continue
		}
		if perLevel[lv.Overlay] == nil {
			perLevel[lv.Overlay] = map[combo]bool{}
		}
		for _, ob := range lv.Objects {
			c := combo{ob.Actor, ob.Params[0], ob.Params[1], ob.Params[2]}
			if !perLevel[lv.Overlay][c] {
				perLevel[lv.Overlay][c] = true
				all[c] = append(all[c], lv.Overlay)
			}
		}
	}

	table := map[int][]Binding{}
	addRun := func(run *sm64ds.ActorRun) {
		if len(run.Files) == 0 && run.Obj == 0 {
			return
		}
		b := Binding{Params: run.Params, Config: run.Config, Models: o.Models(run), Clips: o.Clips(run), KCL: o.KCLs(run), Colliders: run.Colliders, Notes: run.Notes}
		// identical result under another config/params: keep one per distinct model list
		for _, e := range table[run.Actor] {
			if fmt.Sprint(e.Models, e.Params) == fmt.Sprint(b.Models, b.Params) {
				return
			}
		}
		table[run.Actor] = append(table[run.Actor], b)
	}

	// pass 1: per level overlay, actors that resolve there (engine + level actors)
	unresolved := map[combo]bool{}
	var ovls []int
	for ov := range perLevel {
		ovls = append(ovls, ov)
	}
	sort.Ints(ovls)
	for _, ov := range ovls {
		if err := o.LoadConfig(ov); err != nil {
			fmt.Fprintf(os.Stderr, "config ovl%d: %v\n", ov, err)
			continue
		}
		n := 0
		for c := range perLevel[ov] {
			if _, ok := o.Profile(c.actor, ov); !ok {
				unresolved[c] = true
				continue
			}
			addRun(o.RunActor(c.actor, ov, [3]int{c.p1, c.p2, c.p3}))
			n++
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "ovl %d: %d combos run\n", ov, n)
		}
	}

	// pass 2: bank overlays for actors unresolved under their levels
	bankRuns := map[string]bool{}
	for bank := 60; bank <= 102; bank++ {
		var todo []combo
		for c := range unresolved {
			todo = append(todo, c)
		}
		if len(todo) == 0 {
			break
		}
		if err := o.LoadConfig(bank); err != nil {
			continue
		}
		n := 0
		for _, c := range todo {
			if _, ok := o.Profile(c.actor, bank); !ok {
				continue
			}
			key := fmt.Sprint(bank, c)
			if bankRuns[key] {
				continue
			}
			bankRuns[key] = true
			addRun(o.RunActor(c.actor, bank, [3]int{c.p1, c.p2, c.p3}))
			n++
		}
		if verbose && n > 0 {
			fmt.Fprintf(os.Stderr, "bank %d: %d combos run\n", bank, n)
		}
	}
	return table
}
