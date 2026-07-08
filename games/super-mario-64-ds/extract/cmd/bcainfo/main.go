// bcainfo decodes a .bca skeletal animation (sm64ds/bca.go) and prints its
// header and each bone's local TRS at a few frames — a decoder smoke-test and
// inspection tool.
package main

import (
	"fmt"
	"os"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: bcainfo file.bca")
		os.Exit(2)
	}
	a, err := sm64ds.LoadBCA(os.Args[1])
	if err != nil {
		sm64ds.Die(err)
	}
	fmt.Printf("%s: %d bones, %d frames, loop=%v\n", os.Args[1], a.NumBones, a.NumFrames, a.Loop)
	step := a.NumFrames / 5
	if step < 1 {
		step = 1
	}
	for b := 0; b < a.NumBones; b++ {
		fmt.Printf("bone %d\n", b)
		for f := 0; f < a.NumFrames; f += step {
			t := a.BoneTRS(b, f)
			fmt.Printf("  f%03d  s(%.2f %.2f %.2f)  r(%+.2f %+.2f %+.2f)  t(%+.2f %+.2f %+.2f)\n",
				f, t[0], t[1], t[2], t[3], t[4], t[5], t[6], t[7], t[8])
		}
	}
}
