// spawnprobe reports every level's player spawn: the FIRST type-1 entrance
// record (handler $020FE6C8, Part V §2), which is where the game stands the
// player when the level starts.
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

func main() {
	rom := flag.String("in", "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("ext", "extracted", "extracted binaries dir")
	flag.Parse()
	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		log.Fatal(err)
	}
	with, without, seen := 0, 0, map[string]bool{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil || lv.BMDPath == "" {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		if seen[stem] {
			continue
		}
		seen[stem] = true
		if len(lv.Entrances) == 0 {
			without++
			fmt.Printf("  %-22s NO ENTRANCE RECORD\n", stem)
			continue
		}
		with++
		e := lv.Entrances[0]
		fmt.Printf("  %-22s n=%d  spawn %7.0f %7.0f %7.0f  yaw %6.1f\n",
			stem, len(lv.Entrances), e.X, e.Y, e.Z, e.RotY)
	}
	fmt.Printf("%d stages with a spawn, %d without\n", with, without)
}
