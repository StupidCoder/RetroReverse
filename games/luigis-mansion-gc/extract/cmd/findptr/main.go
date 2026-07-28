// findptr scans a bootoracle savestate's main RAM for a 32-bit big-endian
// value and hex-dumps the context around every hit — the pointer hunter that
// walks from a known heap array to the object that owns it (the demo actor
// objects were found this way: search for the world-array address, land in
// the object's pointer block).
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/gc"
)

func main() {
	image := flag.String("image", "", "")
	state := flag.String("state", "", "")
	val := flag.Uint64("val", 0, "32-bit big-endian value to find")
	ctx := flag.Int("ctx", 0x20, "bytes of context to dump before/after")
	flag.Parse()
	disc, err := gc.Open(*image)
	if err != nil {
		panic(err)
	}
	m, err := gc.NewMachine(disc)
	if err != nil {
		panic(err)
	}
	if err := m.LoadStateFile(*state); err != nil {
		panic(err)
	}
	want := uint32(*val)
	for i := 0; i+4 <= len(m.RAM); i += 4 {
		if binary.BigEndian.Uint32(m.RAM[i:]) == want {
			lo := i - *ctx
			if lo < 0 {
				lo = 0
			}
			fmt.Printf("hit at 0x%08X\n", 0x80000000+i)
			for o := lo; o < i+*ctx && o+16 <= len(m.RAM); o += 16 {
				fmt.Printf("  %08X ", 0x80000000+o)
				for k := 0; k < 16; k += 4 {
					fmt.Printf(" %08X", binary.BigEndian.Uint32(m.RAM[o+k:]))
				}
				fmt.Println()
			}
		}
	}
	os.Exit(0)
}
