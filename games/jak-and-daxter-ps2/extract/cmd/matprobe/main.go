package main

import (
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

func main() {
	obj, _ := os.ReadFile(os.Args[1])
	for _, a := range os.Args[2:] {
		p, _ := strconv.ParseUint(a, 0, 32)
		c, _ := merc.Parse(obj, uint32(p))
		fmt.Printf("=== ctrl %s\n", a)
		for ei := range c.Effects {
			e := &c.Effects[ei]
			seq, shaders := merc.EffectStream(e)
			cnt := map[int]int{}
			n := 0
			var w [3]int
			for _, sv := range seq {
				w[0], w[1], w[2] = w[1], w[2], sv.Index
				if n < 2 {
					n++
					continue
				}
				if !sv.ADC && w[0] != w[1] && w[1] != w[2] && w[0] != w[2] {
					cnt[sv.Mat]++
				}
			}
			fmt.Printf("e%d: shaders:", ei)
			for si, s := range shaders {
				fmt.Printf(" [%d]=%08x(%d tris)", si, s.RawID, cnt[si])
			}
			fmt.Println()
			// also list per-fragment shader count and hole sizes
			for fi := range e.Fragments {
				fr := &e.Fragments[fi]
				sh := fr.Shaders()
				if len(sh) > 0 {
					fmt.Printf("   f%d: %d blocks:", fi, len(sh))
					for _, s := range sh {
						fmt.Printf(" %08x", s.RawID)
					}
					fmt.Println()
				}
			}
		}
	}
}
