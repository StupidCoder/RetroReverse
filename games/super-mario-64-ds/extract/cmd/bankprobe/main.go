// bankprobe: which companion overlay does an actor need banked alongside its
// own before its init runs to completion?
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
	"retroreverse.com/tools/platform/nds"
)

func main() {
	rom := flag.String("rom", "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "../extracted", "extracted binaries dir")
	actor := flag.Int("actor", 70, "actor id")
	cfg := flag.Int("ovl", 22, "the actor's own overlay")
	par := flag.Int("par", 65535, "par1")
	flag.Parse()

	img, _ := os.ReadFile(*rom)
	r, err := nds.Open(img)
	if err != nil {
		log.Fatal(err)
	}
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
	_ = r
	// One run with only the actor's own overlay: where does it leave code?
	if err := o.LoadConfig(*cfg); err != nil {
		log.Fatal(err)
	}
	run := o.RunActor(*actor, *cfg, [3]int{*par, 0, 0})
	from, to := o.Escape()
	fmt.Printf("actor %d, overlay %d alone: models %v", *actor, *cfg, o.Models(run))
	if to == 0 {
		fmt.Println("  (never left loaded code)")
		return
	}
	fmt.Printf("\n  execution left code at %08X -> %08X\n", from, to)
	cands := o.OverlaysCovering(to)
	fmt.Printf("  overlays covering %08X: %v\n", to, cands)
	for _, id := range cands {
		if id == *cfg {
			continue
		}
		if err := o.LoadConfigMulti([]int{*cfg, id}); err != nil {
			continue
		}
		run := o.RunActor(*actor, *cfg, [3]int{*par, 0, 0})
		_, to2 := o.Escape()
		fmt.Printf("  + ovl%-3d  escaped=%v  models %v  kcl %v\n",
			id, to2 != 0, o.Models(run), o.KCLs(run))
	}
}
