package main

// mercinfo parses the merc-ctrls of an art group and reports fragment
// tiling — the internal consistency check for the wire format.
import (
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

func main() {
	obj, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	for _, a := range os.Args[2:] {
		p, _ := strconv.ParseUint(a, 0, 32)
		c, err := merc.Parse(obj, uint32(p))
		if err != nil {
			fmt.Println("ERR:", err)
			continue
		}
		fmt.Printf("merc-ctrl @0x%X: %d effects\n", p, len(c.Effects))
		for i, e := range c.Effects {
			bytesTotal, mats := 0, 0
			for _, f := range e.Fragments {
				bytesTotal += len(f.ByteData) + len(f.LumpData) + len(f.FPData)
				mats += len(f.Mats)
			}
			fmt.Printf("  effect %d: %3d frags, %5d tris, %5d dverts, %6d geo bytes, %d mat xfers\n",
				i, e.FragCount, e.TriCount, e.DVertCount, bytesTotal, mats)
		}
	}
}
