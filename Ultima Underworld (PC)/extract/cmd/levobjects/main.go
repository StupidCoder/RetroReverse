// Command levobjects lists a level's placed objects (item id, tile, fine
// position, height, heading), decoded from LEV.ARK via lev.Objects. Optional
// -lo/-hi filter by item id (e.g. doors are 320-335).
//
// Usage: levobjects [-game ../game] [-level 0] [-lo 0] [-hi 511]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"ultimaunderworld/extract/lev"
)

func main() {
	game := flag.String("game", "../game", "path to the game/ folder")
	level := flag.Int("level", 0, "level index (0-7)")
	lo := flag.Int("lo", 0, "min item id")
	hi := flag.Int("hi", 511, "max item id")
	flag.Parse()

	b, err := os.ReadFile(filepath.Join(*game, "DATA", "LEV.ARK"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ark, err := lev.ParseArk(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	block, _ := ark.Block(*level)
	grid, _ := lev.DecodeGrid(block)
	objs := lev.Objects(grid, block)

	hist := map[uint16]int{}
	shown := 0
	for _, o := range objs {
		hist[o.ItemID]++
		if int(o.ItemID) < *lo || int(o.ItemID) > *hi {
			continue
		}
		kind := "static"
		if o.Mobile {
			kind = "mobile"
		}
		fmt.Printf("tile(%2d,%2d) idx=%4d item=%3d(0x%03X) fine=(%d,%d) z=%3d hd=%d q=%d %s\n",
			o.TileX, o.TileY, o.Index, o.ItemID, o.ItemID, o.FineX, o.FineY, o.Z, o.Heading, o.Quality, kind)
		shown++
	}
	ids := make([]int, 0, len(hist))
	for id := range hist {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	fmt.Printf("--- level %d: %d objects, %d shown; item ids: ", *level, len(objs), shown)
	for _, id := range ids {
		fmt.Printf("%d:%d ", id, hist[uint16(id)])
	}
	fmt.Println()
}
