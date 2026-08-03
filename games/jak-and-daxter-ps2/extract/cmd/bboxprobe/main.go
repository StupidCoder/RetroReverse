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
		c, err := merc.Parse(obj, uint32(p))
		if err != nil {
			panic(err)
		}
		var gminx, gminy, gminz = float32(1e9), float32(1e9), float32(1e9)
		var gmaxx, gmaxy, gmaxz = float32(-1e9), float32(-1e9), float32(-1e9)
		wide := 0
		for ei := range c.Effects {
			for fi := range c.Effects[ei].Fragments {
				fr := &c.Effects[ei].Fragments[fi]
				var mnx, mny, mnz = float32(1e9), float32(1e9), float32(1e9)
				var mxx, mxy, mxz = float32(-1e9), float32(-1e9), float32(-1e9)
				for _, v := range fr.Vertices() {
					if v.X < mnx {
						mnx = v.X
					}
					if v.X > mxx {
						mxx = v.X
					}
					if v.Y < mny {
						mny = v.Y
					}
					if v.Y > mxy {
						mxy = v.Y
					}
					if v.Z < mnz {
						mnz = v.Z
					}
					if v.Z > mxz {
						mxz = v.Z
					}
				}
				if mxx-mnx > 255 || mxy-mny > 255 || mxz-mnz > 255 {
					wide++
					fmt.Printf("  WIDE ctrl %s e%d f%d: dx=%.0f dy=%.0f dz=%.0f\n", a, ei, fi, mxx-mnx, mxy-mny, mxz-mnz)
				}
				if mnx < gminx {
					gminx = mnx
				}
				if (mxx) > gmaxx {
					gmaxx = mxx
				}
				if mny < gminy {
					gminy = mny
				}
				if mxy > gmaxy {
					gmaxy = mxy
				}
				if mnz < gminz {
					gminz = mnz
				}
				if mxz > gmaxz {
					gmaxz = mxz
				}
			}
		}
		fmt.Printf("ctrl %s: model bbox [%.0f,%.0f,%.0f]..[%.0f,%.0f,%.0f], wide frags: %d\n",
			a, gminx, gminy, gminz, gmaxx, gmaxy, gmaxz, wide)
	}
}
