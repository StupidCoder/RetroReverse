// pipes lists a Super Mario Land level's warp pipes (the bonus-room entrances), decoded
// from the $651C table in ROM (Super_Mario_Land.md Part V §4).
//
//	go run ./cmd/pipes [-rom PATH] [-id NN]   # NN hex, e.g. 11; omit to list all 12 levels
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/games/super-mario-land-gb/extract/level"
)

func main() {
	rom := flag.String("rom", "../Super Mario Land (World).gb", "ROM path")
	id := flag.String("id", "", "level id (hex); omit to list every level")
	flag.Parse()
	data, err := os.ReadFile(*rom)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pipes:", err)
		os.Exit(1)
	}
	var ids []byte
	if *id != "" {
		v, _ := strconv.ParseUint(*id, 16, 8)
		ids = []byte{byte(v)}
	} else {
		for w := 1; w <= 4; w++ {
			for l := 1; l <= 3; l++ {
				ids = append(ids, byte(w<<4|l))
			}
		}
	}
	for _, lid := range ids {
		ps := level.DecodePipes(data, lid)
		fmt.Printf("%d-%d: %d pipe(s)\n", lid>>4, lid&0xF, len(ps))
		for _, p := range ps {
			fmt.Printf("    on screen $%02X col $%02X  ->  bonus screen %d, return to screen $%02X at (%d,%d)\n",
				p.Screen, p.Col, p.Dest, p.RetScreen, p.RetX, p.RetY)
		}
	}
}
