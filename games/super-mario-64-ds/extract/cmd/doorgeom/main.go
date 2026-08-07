// doorgeom measures each door leaf model's X extent and the constant
// translation its clip carries, to settle where the record position sits
// relative to the doorway.
package main

import (
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

func bbox(m *sm64ds.Model) (lo, hi [3]float64) {
	lo = [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	hi = [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, tris := range m.ByMat {
		for _, t := range tris {
			for _, v := range t.V {
				for i, c := range [3]float64{v.X, v.Y, v.Z} {
					lo[i] = math.Min(lo[i], c)
					hi[i] = math.Max(hi[i], c)
				}
			}
		}
	}
	return
}

func main() {
	ls, err := sm64ds.OpenLevels("Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "extracted")
	if err != nil {
		log.Fatal(err)
	}
	fs := []string{"obj_door1", "obj_door2_boro", "obj_door4_yami", "obj_door5_horror",
		"obj_door0_star", "obj_door0_star1", "obj_door0_star3", "obj_door0_star10", "obj_stargate"}
	for _, s := range fs {
		p, _ := filepath.Glob("extracted/files/data/*/*/" + s + ".bmd")
		if len(p) == 0 {
			p, _ = filepath.Glob("extracted/files/data/*/" + s + ".bmd")
		}
		if len(p) == 0 {
			fmt.Println(s, "not found")
			continue
		}
		m, err := sm64ds.LoadBMD(p[0])
		if err != nil {
			fmt.Println(s, err)
			continue
		}
		lo, hi := bbox(m)
		fmt.Printf("%-18s bones=%d  x %8.3f..%8.3f (w %7.3f)  y %8.3f..%8.3f  z %8.3f..%8.3f\n",
			s, m.NumBones, lo[0], hi[0], hi[0]-lo[0], lo[1], hi[1], lo[2], hi[2])
	}
	for _, ref := range []string{"ar1_8", "ar1_9", "ar1_10", "ar1_11"} {
		var arc string
		var num int
		fmt.Sscanf(strings.Replace(ref, "_", " ", 1), "%s %d", &arc, &num)
		d, err := ls.ArchiveMember(sm64ds.ArchiveRef{Archive: arc, Member: num})
		if err != nil {
			fmt.Println(ref, err)
			continue
		}
		if m, err := sm64ds.Decode(d, ref); err == nil {
			lo, hi := bbox(m)
			fmt.Printf("%-18s bones=%d  x %8.3f..%8.3f (w %7.3f)  y %8.3f..%8.3f  z %8.3f..%8.3f\n",
				ref, m.NumBones, lo[0], hi[0], hi[0]-lo[0], lo[1], hi[1], lo[2], hi[2])
		} else if a, err2 := sm64ds.DecodeBCA(d); err2 == nil {
			fmt.Printf("%-18s CLIP bones=%d frames=%d\n", ref, a.NumBones, a.NumFrames)
		} else {
			fmt.Printf("%-18s neither bmd nor bca\n", ref)
		}
	}
}
