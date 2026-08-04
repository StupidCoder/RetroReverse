package main

// contcheck: concatenated-spool sanity — quat norms everywhere, and are
// chunk-boundary frames duplicates?
import (
	"fmt"
	"math"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/tools/lib/iso9660"
)

func u32(b []byte, o int) uint32 {
	return uint32(b[o]) | uint32(b[o+1])<<8 | uint32(b[o+2])<<16 | uint32(b[o+3])<<24
}

func main() {
	f, _ := os.Open("image/Jak and Daxter - The Precursor Legacy.iso")
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	if err != nil { panic(err) }
	tab, err := goalobj.LoadSymTab("work/goal.txt")
	if err != nil { panic(err) }
	data, err := vol.ReadFile("STR/NDINTRO.STR;1")
	if err != nil { panic(err) }
	var starts []int
	for o := 0; o < 2048; o += 4 {
		v := int(u32(data, o))
		if v == 0 { break }
		starts = append(starts, v*2048)
	}
	animType := tab.Syms["art-joint-anim"].Value
	byName := map[string][]*merc.JointAnim{}
	for ci, s := range starts {
		end := len(data)
		if ci+1 < len(starts) { end = starts[ci+1] }
		obj, _, err := goalobj.Link(data[s:end], 0, tab)
		if err != nil { panic(err) }
		for _, ap := range merc.FindAnims(obj, animType) {
			a, err := merc.DecodeJointAnim(obj, ap)
			if err != nil { panic(err) }
			byName[a.Name] = append(byName[a.Name], a)
		}
	}
	for name, parts := range byName {
		bad := 0
		for _, a := range parts {
			for _, fr := range a.Frames {
				for j := 2; j < len(fr); j++ {
					q := fr[j].Quat
					n := math.Sqrt(float64(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]))
					if math.Abs(n-1) > 0.01 { bad++ }
				}
			}
		}
		// boundary duplicate test: last frame of part i vs first of part i+1
		dup := 0
		for i := 0; i+1 < len(parts); i++ {
			la := parts[i].Frames[len(parts[i].Frames)-1]
			fb := parts[i+1].Frames[0]
			same := true
			for j := range la {
				if la[j] != fb[j] { same = false; break }
			}
			if same { dup++ }
		}
		total := 0
		for _, a := range parts { total += len(a.Frames) }
		fmt.Printf("%-24s parts %d, frames %d, bad-quats %d, boundary-dups %d/%d\n",
			name, len(parts), total, bad, dup, len(parts)-1)
	}
}
