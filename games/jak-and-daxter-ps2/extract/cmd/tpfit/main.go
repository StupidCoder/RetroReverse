package main

// tpfit: solve tier-texture atlas positions by mip consistency. The deepest
// mip of every texture lives in tier 0 (the verified far tier); the game's
// mip chain is an exact 2x2 box filter (observed: correct decodes match
// MAD 0.0). So for each texture, search the 512-byte-wide raster atlas for
// the (X, Y) whose decode box-downscales exactly onto the next-deeper mip.

import (
	"fmt"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/tpage"
	"retroreverse.com/tools/lib/iso9660"
)

func main() {
	f, _ := os.Open(os.Args[1])
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	data, err := vol.ReadFile(os.Args[2] + ";1")
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	tab, err := goalobj.LoadSymTab(os.Args[3])
	check(err)
	var raw []byte
	for _, e := range d.Entries {
		if e.Name == os.Args[4] {
			raw = e.Data
			break
		}
	}
	obj, _, err := goalobj.Link(raw, 0, tab)
	check(err)
	pg, err := tpage.Load(obj)
	check(err)
	solved, total := 0, 0
	for i := range pg.Textures {
		t := &pg.Textures[i]
		if t.Mips < 2 {
			continue
		}
		res := pg.FitTiers(t)
		for m, r := range res {
			if r.Searched {
				total++
				if r.OK {
					solved++
					fmt.Printf("%-26s mip%d at atlas+0x%06X (X=%d,Y=%d) err %.2f\n",
						t.Name, m, r.Offset, r.X, r.Y, r.Err)
				} else {
					fmt.Printf("%-26s mip%d NOT FOUND (best err %.2f at 0x%X)\n", t.Name, m, r.Err, r.Offset)
				}
			}
		}
	}
	fmt.Printf("solved %d/%d searched mips\n", solved, total)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "tpfit:", err)
		os.Exit(1)
	}
}
