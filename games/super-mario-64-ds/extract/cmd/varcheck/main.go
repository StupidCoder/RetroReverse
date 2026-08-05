// varcheck verifies the SHIPPED Retro-X level documents against a fresh decode
// of the cartridge: every placement's mission (variant) membership must be
// exactly the star-layer set the level's own object table lists that object
// under, and the declared variants must cover every layer in use.
//
// It opens the emitted JSON rather than re-deriving it from the exporter's
// structs, so an exporter that never wrote the field fails here.
//
// Usage (from games/super-mario-64-ds/):
//
//	go run ./extract/cmd/varcheck -in <rom.nds> [-ext extracted] [-levels DIR] [-v]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

type placement struct {
	ID       int            `json:"id"`
	Object   string         `json:"object"`
	Name     string         `json:"name"`
	Pos      []float64      `json:"pos"`
	Variants []string       `json:"variants"`
	Behavior *behavior      `json:"behavior"`
	Props    map[string]any `json:"props"`
}

type behavior struct {
	Kind string    `json:"kind"`
	Axis []float64 `json:"axis"`
	Rate float64   `json:"rate"`
}

type variant struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type levelDoc struct {
	Variants   []variant   `json:"variants"`
	Placements []placement `json:"placements"`
}

func main() {
	in := flag.String("in", "", "cartridge image")
	ext := flag.String("ext", "extracted", "extracted binaries directory")
	dir := flag.String("levels", filepath.Join("..", "..", "site", "public", "super-mario-64-ds", "levels"), "shipped level documents")
	verbose := flag.Bool("v", false, "print the per-mission placement census")
	flag.Parse()
	if *in == "" {
		log.Fatal("usage: varcheck -in <rom.nds> [-ext extracted] [-levels DIR] [-v]")
	}

	ls, err := sm64ds.OpenLevels(*in, *ext)
	if err != nil {
		log.Fatal(err)
	}

	fails, checked, files := 0, 0, 0
	fail := func(f, format string, a ...any) {
		fails++
		fmt.Printf("FAIL %s: %s\n", f, fmt.Sprintf(format, a...))
	}

	seenStem := map[string]bool{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil || lv.BMDPath == "" {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		if seenStem[stem] {
			continue // the exporter keeps the first level that uses a stage
		}
		seenStem[stem] = true
		name := slugify(strings.TrimSuffix(stem, "_all")) + ".json"
		raw, err := os.ReadFile(filepath.Join(*dir, name))
		if err != nil {
			continue // not a shipped stage
		}
		var doc levelDoc
		if err := json.Unmarshal(raw, &doc); err != nil {
			fail(name, "unreadable: %v", err)
			continue
		}
		files++

		// The cartridge's own answer: (actor, position) -> set of star layers.
		// The ACTOR is part of the key because two placements can share a
		// spot and belong to different missions — Bob-omb Battlefield's race
		// flagpole (star 2) and its second King Bob-omb (star 4) stand on the
		// same point.
		want := map[string]map[int]bool{}
		used := map[int]bool{}
		for _, o := range lv.Objects {
			k := poskey(o.Actor, o.X/1000, o.Y/1000, o.Z/1000)
			if want[k] == nil {
				want[k] = map[int]bool{}
			}
			want[k][o.Layer] = true
			if o.Layer != 0 {
				used[o.Layer] = true
			}
		}

		declared := map[string]bool{}
		for _, v := range doc.Variants {
			declared[v.ID] = true
		}
		defaults := 0
		for _, v := range doc.Variants {
			if v.Default {
				defaults++
			}
		}
		if len(doc.Variants) > 0 && defaults != 1 {
			fail(name, "%d default variants", defaults)
		}
		if len(used) > 0 && len(doc.Variants) == 0 {
			fail(name, "level uses star layers %v but declares no variants", sorted(used))
		}
		for l := range used {
			if len(doc.Variants) > 0 && !declared[fmt.Sprintf("star%d", l)] {
				fail(name, "star layer %d is in use but star%d is not declared", l, l)
			}
		}

		census := map[string]int{}
		unmatched := 0
		for _, pl := range doc.Placements {
			if len(pl.Pos) != 3 || pl.ID == 0 {
				continue
			}
			for _, v := range pl.Variants {
				if !declared[v] {
					fail(name, "placement %d claims undeclared variant %q", pl.ID, v)
				}
				census[v]++
			}
			if len(pl.Variants) == 0 {
				census["(all)"]++
			}
			actor, hasActor := pl.Props["actor"].(float64)
			var layers map[int]bool
			if hasActor {
				layers = want[poskey(int(actor), pl.Pos[0], pl.Pos[1], pl.Pos[2])]
			}
			if layers == nil {
				// synthesised: the chain chomp's spawned stake and drawn
				// links, and the bodies the exporter lifts off their record
				unmatched++
				continue
			}
			checked++
			got := strings.Join(pl.Variants, ",")
			exp := strings.Join(expect(layers, declared), ",")
			if got != exp {
				fail(name, "placement %d (%v) variants %q, cartridge says %q",
					pl.ID, pl.Pos, got, exp)
			}
			// a coin is a coin in every mission it appears in
			if a := int(actor); a == 288 || a == 289 || a == 290 {
				if pl.Behavior == nil || pl.Behavior.Kind != "spin" || pl.Behavior.Rate < 8.8 || pl.Behavior.Rate > 8.9 {
					fail(name, "coin placement %d does not spin: %+v", pl.ID, pl.Behavior)
				}
			}
		}
		if *verbose && len(doc.Variants) > 0 {
			fmt.Printf("%-24s", name)
			for _, v := range doc.Variants {
				fmt.Printf("  %s=%d", v.ID, census[v.ID])
			}
			fmt.Printf("  all=%d  synthesised=%d\n", census["(all)"], unmatched)
		}
	}
	fmt.Printf("%d level documents, %d placements matched to a cartridge record, %d failures\n",
		files, checked, fails)
	if fails > 0 {
		os.Exit(1)
	}
}

// expect is the exporter's rule, restated: layer 0 means every mission, which
// Retro-X spells as no variants field at all, and so does membership of every
// declared mission.
func expect(layers map[int]bool, declared map[string]bool) []string {
	if len(declared) == 0 || layers[0] {
		return nil
	}
	var out []string
	for _, l := range sorted(layers) {
		if id := fmt.Sprintf("star%d", l); declared[id] {
			out = append(out, id)
		}
	}
	if len(out) == len(declared) {
		return nil
	}
	return out
}

func sorted(m map[int]bool) []int {
	var out []int
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// poskey rounds to the exporter's three decimals so a shipped position and a
// cartridge short land on the same key.
func poskey(actor int, x, y, z float64) string {
	return fmt.Sprintf("%d/%.3f/%.3f/%.3f", actor, x, y, z)
}

func slugify(s string) string {
	var out []rune
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
		default:
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	return strings.Trim(string(out), "-")
}
