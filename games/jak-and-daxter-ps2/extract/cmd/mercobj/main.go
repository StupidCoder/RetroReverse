package main

// mercobj writes a merc-ctrl's decoded geometry as a Wavefront OBJ
// (positions + naive strip triangles) for visual verification.
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
	p, _ := strconv.ParseUint(os.Args[2], 0, 32)
	c, err := merc.Parse(obj, uint32(p))
	if err != nil {
		panic(err)
	}
	out, err := os.Create(os.Args[3])
	if err != nil {
		panic(err)
	}
	defer out.Close()
	base := 1
	nv, nt := 0, 0
	for _, e := range c.Effects {
		for _, fr := range e.Fragments {
			vs := fr.Vertices()
			slots := map[int]int{}
			for i, v := range vs {
				fmt.Fprintf(out, "v %g %g %g\n", v.X, v.Y, v.Z)
				if v.Slot1 >= 0 {
					slots[v.Slot1] = base + i
				}
				if v.Slot2 >= 0 {
					slots[v.Slot2] = base + i
				}
			}
			mx := -1
			for s := range slots {
				if s > mx {
					mx = s
				}
			}
			for i := 2; i <= mx; i++ {
				a, b, cc := slots[i-2], slots[i-1], slots[i]
				if a != 0 && b != 0 && cc != 0 && a != b && b != cc && a != cc {
					fmt.Fprintf(out, "f %d %d %d\n", a, b, cc)
					nt++
				}
			}
			base += len(vs)
			nv += len(vs)
		}
	}
	fmt.Printf("wrote %d verts, %d tris\n", nv, nt)
}
