package main

import (
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

func main() {
	obj, _ := os.ReadFile(os.Args[1])
	p, _ := strconv.ParseUint(os.Args[2], 0, 32)
	c, _ := merc.Parse(obj, uint32(p))
	maxIdx, minIdx := 0, 255
	used := map[byte]bool{}
	unmatched := 0
	dests := map[byte]byte{} // running across the whole ctrl
	bad := map[byte]int{}
	for ei := range c.Effects {
		for fi := range c.Effects[ei].Fragments {
			fr := &c.Effects[ei].Fragments[fi]
			for _, m := range fr.Mats {
				if int(m.Index) > maxIdx {
					maxIdx = int(m.Index)
				}
				if int(m.Index) < minIdx {
					minIdx = int(m.Index)
				}
				dests[m.Dest] = m.Index
				used[m.Index] = true
			}
			for _, v := range fr.Vertices() {
				a := v.Mat & 0x7F
				if _, ok := dests[a]; !ok {
					unmatched++
					if bad[a] == 0 {
						fmt.Printf("  first unmatched addr %d (raw %02x) in e%d f%d\n", a, v.Mat, ei, fi)
					}
					bad[a]++
				}
			}
		}
	}
	fmt.Printf("palette idx %d..%d, distinct %d, unmatched vert-mats %d, bad addrs %v\n",
		minIdx, maxIdx, len(used), unmatched, bad)
}
