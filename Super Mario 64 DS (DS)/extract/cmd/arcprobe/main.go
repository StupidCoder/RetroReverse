// arcprobe: decode every member of arc0/ar1 (the game's flagged-ID archives) as a
// .bmd and print its material/texture names.
package main

import (
	"fmt"
	"os"

	"retroreverse.com/tools/nds"
	"supermario64ds/extract/sm64ds"
)

// plausibleBMD sanity-checks the fixed header before a full decode.
func plausibleBMD(d []byte) bool {
	if len(d) < 0x40 {
		return false
	}
	u := func(o int) uint32 { return uint32(d[o]) | uint32(d[o+1])<<8 | uint32(d[o+2])<<16 | uint32(d[o+3])<<24 }
	if u(0) > 16 {
		return false
	} // scale shift
	for _, o := range []int{4, 0xC, 0x14, 0x1C, 0x24} {
		n, p := u(o), u(o+4)
		if n > 512 || p > uint32(len(d)) {
			return false
		}
	}
	return true
}

func main() {
	for _, arc := range []string{"arc0", "ar1"} {
		d, err := os.ReadFile("../extracted/files/ARCHIVE/" + arc + ".narc")
		if err != nil {
			panic(err)
		}
		files, err := nds.ParseNARCFiles(d)
		if err != nil {
			panic(err)
		}
		for i, f := range files {
			data := f.Data
			if len(data) > 4 && string(data[:4]) == "LZ77" {
				data = nds.Decompress(data[4:])
			}
			if !plausibleBMD(data) {
				continue
			}
			m, err := sm64ds.Decode(data, fmt.Sprintf("%s_%d", arc, i))
			if err != nil {
				continue
			}
			names := []string{}
			for _, mt := range m.Mats {
				names = append(names, mt.Name)
			}
			tris := 0
			for _, ts := range m.ByMat {
				tris += len(ts)
			}
			fmt.Printf("%s[%3d] (id %#06x) %4d tris mats=%v\n", arc, i, map[string]int{"arc0": 0x8000, "ar1": 0x9C00}[arc]+i, tris, names)
		}
	}
}
