// doorprobe runs the door actor for every distinct parameter pair the levels
// place, and reports what each asks the loader for.
package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

func main() {
	rom := flag.String("rom", "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "extracted", "extracted binaries dir")
	actor := flag.Int("actor", 353, "actor id")
	cfg := flag.Int("ovl", 100, "overlay carrying the profile")
	clip := flag.String("clip", "", "dump this clip against -model instead of sweeping")
	model := flag.String("model", "ar1_9", "model for -clip")
	models := flag.String("models", "", "comma-separated stems: print bone counts")
	flag.Parse()

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		log.Fatal(err)
	}
	o, err := sm64ds.NewOracle(ls)
	if err != nil {
		log.Fatal(err)
	}
	if err := o.InitEngine(); err != nil {
		log.Fatal(err)
	}

	if *models != "" {
		dumpModel(ls, strings.Split(*models, ","))
		return
	}
	if *clip != "" {
		dumpClip(ls, *model, *clip)
		return
	}
	seen := map[[3]int]int{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil {
			continue
		}
		for _, ob := range lv.Objects {
			if ob.Actor == *actor {
				seen[ob.Params]++
			}
		}
	}
	var pars [][3]int
	for p := range seen {
		pars = append(pars, p)
	}
	sort.Slice(pars, func(i, j int) bool {
		if pars[i][1] != pars[j][1] {
			return pars[i][1] < pars[j][1]
		}
		return pars[i][0] < pars[j][0]
	})
	fmt.Printf("%d distinct parameter pairs for actor %d\n", len(pars), *actor)
	for _, p := range pars {
		if err := o.LoadConfig(*cfg); err != nil {
			log.Fatal(err)
		}
		run := o.RunActorBanked(*actor, *cfg, p, func(extra int) error {
			if extra < 0 {
				return o.LoadConfig(*cfg)
			}
			return o.LoadConfigMulti([]int{*cfg, extra})
		})
		_, esc := o.Escape()
		fmt.Printf("  par1 %04X par2 %04X  x%-3d models %-28v clips %-14v esc=%v obj=%v\n",
			p[0], p[1], seen[p], o.Models(run), o.Clips(run), esc != 0, run.Obj != 0)
	}
}
