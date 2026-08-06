// bssprobe asks where an address lives relative to an overlay's image, and what
// the oracle actually has there after loading it.
//
//	bssprobe -ovl 22 -at 02114690
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
	rom := flag.String("rom", "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "extracted", "extracted binaries dir")
	ovl := flag.Int("ovl", 22, "overlay to load")
	at := flag.String("at", "", "hex address to dump after loading")
	n := flag.Int("n", 8, "words to dump")
	flag.Parse()

	img, err := os.ReadFile(*rom)
	if err != nil {
		log.Fatal(err)
	}
	r, err := nds.Open(img)
	if err != nil {
		log.Fatal(err)
	}
	for _, o := range r.ARM9Overlays() {
		if int(o.ID) != *ovl {
			continue
		}
		data, err := os.ReadFile(fmt.Sprintf("%s/ovl9_%03d_dec.bin", *ext, o.ID))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("overlay %d: RAM %08X  RAMSize %X  file %X  BSS %X  -> image %08X..%08X, bss ..%08X\n",
			o.ID, o.RAMAddr, o.RAMSize, len(data), o.BSSSize,
			o.RAMAddr, o.RAMAddr+uint32(len(data)), o.RAMAddr+uint32(len(data))+o.BSSSize)
		fmt.Printf("  static init list %08X..%08X (%d entries)\n",
			o.StaticInitStart, o.StaticInitEnd, (o.StaticInitEnd-o.StaticInitStart)/4)
	}

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		log.Fatal(err)
	}
	o, err := sm64ds.NewOracle(ls)
	if err != nil {
		log.Fatal(err)
	}
	o.Trace = func(s string) { fmt.Fprintln(os.Stderr, "  |", s) }
	if err := o.InitEngine(); err != nil {
		log.Fatal(err)
	}
	if err := o.LoadConfig(*ovl); err != nil {
		log.Fatal(err)
	}
	if *at != "" {
		var a uint32
		fmt.Sscanf(*at, "%x", &a)
		fmt.Printf("after LoadConfig(%d), memory at %08X:\n", *ovl, a)
		for i := 0; i < *n; i++ {
			fmt.Printf("  +%02X  %08X\n", i*4, o.R32(a+uint32(i*4)))
		}
	}
}
