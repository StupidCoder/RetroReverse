package main

// nprobe: normals audit. Measures the signed agreement between triangle
// geometric normals (as wound in the export) and the decoded vertex
// normals — the winding-phase/orientation check.
import (
	"fmt"
	"math"
	"os"
	"strconv"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

func main() {
	obj, _ := os.ReadFile(os.Args[1])
	for _, a := range os.Args[2:] {
		p, _ := strconv.ParseUint(a, 0, 32)
		c, _ := merc.Parse(obj, uint32(p))
		var sum float64
		var pos, neg, cnt int
		for ei := range c.Effects {
			for _, pr := range merc.TexturedPrims(&c.Effects[ei], nil) {
				for _, t := range pr.Tris {
					av, bv, cv := pr.Positions[t[0]], pr.Positions[t[1]], pr.Positions[t[2]]
					ux, uy, uz := bv[0]-av[0], bv[1]-av[1], bv[2]-av[2]
					vx, vy, vz := cv[0]-av[0], cv[1]-av[1], cv[2]-av[2]
					gx := float64(uy*vz - uz*vy)
					gy := float64(uz*vx - ux*vz)
					gz := float64(ux*vy - uy*vx)
					gl := math.Sqrt(gx*gx + gy*gy + gz*gz)
					if gl < 1e-9 {
						continue
					}
					for _, vi := range t {
						n := pr.Normals[vi]
						d := (gx*float64(n[0]) + gy*float64(n[1]) + gz*float64(n[2])) / gl
						sum += d
						if d > 0 {
							pos++
						} else {
							neg++
						}
						cnt++
					}
				}
			}
		}
		fmt.Printf("ctrl %s: mean signed dot %.3f  agree %d / disagree %d\n", a, sum/float64(cnt), pos, neg)
	}
}
