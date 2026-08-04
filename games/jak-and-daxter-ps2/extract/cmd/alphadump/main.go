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
	c, err := merc.Parse(obj, uint32(p))
	if err != nil { panic(err) }
	for ei := range c.Effects {
		e := &c.Effects[ei]
		seen := map[uint64]int{}
		for fi := range e.Fragments {
			for _, s := range e.Fragments[fi].Shaders() {
				seen[s.Alpha]++
			}
		}
		fmt.Printf("e%d flags %#x tris %d:", ei, e.Flags, e.TriCount)
		for a, n := range seen {
			fmt.Printf("  ALPHA{A=%d B=%d C=%d D=%d FIX=0x%02X}x%d", a&3, a>>2&3, a>>4&3, a>>6&3, a>>32&0xFF, n)
		}
		fmt.Println()
	}
}
