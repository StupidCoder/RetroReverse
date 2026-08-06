package main

import (
	"fmt"
	"math"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

func archiveRefByStem(ls *sm64ds.LevelSet, stem string) (sm64ds.ArchiveRef, bool) {
	i := strings.LastIndexByte(stem, '_')
	if i < 0 {
		return sm64ds.ArchiveRef{}, false
	}
	var member int
	if _, err := fmt.Sscanf(stem[i+1:], "%d", &member); err != nil {
		return sm64ds.ArchiveRef{}, false
	}
	name := stem[:i]
	for _, arc := range []string{"arc0", "ar1", "c2d", "en1"} {
		if name == arc {
			return sm64ds.ArchiveRef{Archive: name, Member: member}, true
		}
	}
	return sm64ds.ArchiveRef{}, false
}

// dumpClip prints a .bca's per-frame bone rotation, so a clip can be read
// rather than assumed: which bone swings, through what angle, and whether the
// LAST frame is the open one.
func dumpClip(ls *sm64ds.LevelSet, model, clip string) {
	mref, ok := archiveRefByStem(ls, model)
	if !ok {
		fmt.Println("no such model member", model)
		return
	}
	md, _ := ls.ArchiveMember(mref)
	m, err := sm64ds.Decode(md, model)
	if err != nil {
		fmt.Println("model:", err)
		return
	}
	cref, ok := archiveRefByStem(ls, clip)
	if !ok {
		fmt.Println("no such clip member", clip)
		return
	}
	cd, _ := ls.ArchiveMember(cref)
	a, err := sm64ds.DecodeBCA(cd)
	if err != nil {
		fmt.Println("clip:", err)
		return
	}
	fmt.Printf("%s: %d bones, %d frames (model %s has %d bones)\n",
		clip, a.NumBones, a.NumFrames, model, m.NumBones)
	for b := 0; b < a.NumBones && b < 6; b++ {
		fmt.Printf("  bone %d:", b)
		for f := 0; f < a.NumFrames; f += max(1, a.NumFrames/8) {
			t := a.BoneTRS(b, f)
			fmt.Printf("\n    f%-3d rot(%.1f,%.1f,%.1f) trans(%.2f,%.2f,%.2f) scale(%.2f,%.2f,%.2f)",
				f, deg(t[3]), deg(t[4]), deg(t[5]), t[6], t[7], t[8], t[0], t[1], t[2])
		}
		peak, at := 0.0, 0
		for f := 0; f < a.NumFrames; f++ {
			if v := deg(a.BoneTRS(b, f)[4]); math.Abs(v) > math.Abs(peak) {
				peak, at = v, f
			}
		}
		t := a.BoneTRS(b, a.NumFrames-1)
		fmt.Printf("\n    last f%d rot(%.1f,%.1f,%.1f);  PEAK yaw %.3f deg at frame %d\n",
			a.NumFrames-1, deg(t[3]), deg(t[4]), deg(t[5]), peak, at)
	}
}

func deg(v float64) float64 { return v * 180 / 3.141592653589793 }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
