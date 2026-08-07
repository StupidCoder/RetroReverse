// rotcensus counts what a standard object's record actually carries at +$8 and
// +$C — the two shorts that were read as extra parameter words until the type-0
// spawn handler turned out to hand `record + 8` to the base ctor as the object's
// three-short ROTATION (see sm64ds.LevelObject).
//
// It exists because "it is the rotation" is a claim about 2,575 placements, not
// about the four that made it visible, and the distribution is the check: 268
// records are non-zero, and the values cluster on exact angles — $4000, $8000,
// $C000, $2AAA and $5555 (90, 180, -90, 60 and 120 degrees).
//
// It also shows the other half of the truth. An actor is free to read its own
// rotation shorts as data, and several do: 106 signposts carry $FFFF at +$8 and
// the cave's lift carries 1465/2200/2665/4265. So the storage is the rotation
// and the actor decides what it means — which is why the export applies it per
// actor, on the evidence of that actor's own draw, rather than to everything.
//
//	rotcensus [-rom img] [-extracted dir]
package main

import (
	"flag"
	"fmt"
	"sort"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

func main() {
	rom := flag.String("rom", "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "extracted", "extracted binaries dir")
	flag.Parse()
	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		sm64ds.Die(err)
	}
	cx := map[int]int{}
	cz := map[int]int{}
	n, nz := 0, 0
	byActor := map[int]map[[2]int]int{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil {
			continue
		}
		for _, ob := range lv.Objects {
			if ob.Simple || ob.Door {
				continue
			}
			n++
			cx[ob.Params[1]]++
			cz[ob.Params[2]]++
			if ob.Params[1] != 0 || ob.Params[2] != 0 {
				nz++
				if byActor[ob.Actor] == nil {
					byActor[ob.Actor] = map[[2]int]int{}
				}
				byActor[ob.Actor][[2]int{ob.Params[1], ob.Params[2]}]++
			}
		}
	}
	fmt.Printf("%d standard placements, %d with a non-zero +$8 or +$C\n", n, nz)
	dump := func(name string, m map[int]int) {
		keys := []int{}
		for k := range m {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		fmt.Printf("%s values: ", name)
		for _, k := range keys {
			fmt.Printf("%04X(n=%d, %.1f deg) ", k, m[k], float64(int16(k))*360/0x10000)
		}
		fmt.Println()
	}
	dump("+$8", cx)
	dump("+$C", cz)
	actors := []int{}
	for a := range byActor {
		actors = append(actors, a)
	}
	sort.Ints(actors)
	for _, a := range actors {
		fmt.Printf("  actor %3d: %v\n", a, byActor[a])
	}
}
