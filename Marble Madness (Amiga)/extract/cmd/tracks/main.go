// tracks decodes the per-course *Track files — the container for ALL of a course's
// gameplay data (not music; that is *Snd). Each Track is a plain AmigaDOS hunk module
// the engine LoadSegs at course init (load_track_data $003176); its first segment is a
// header of ten relocated pointers, each fanned out to a different actor-system global
// (see Marble_Madness.md Part IV §5 for the full map).
//
// This tool prints, per course, the counts of the four Track structures whose record
// format is pinned — objects (placement table, header +4: 3-byte [X][Y][type] records,
// $FF-terminated), slope regions (the static height field, +0/$9A6), dynamic regions
// (the scripted seesaws/holes/triggers, +$14/$FD2C) and coarse zones (+8/$9D4) — plus
// the full placement table with its per-type histogram.
//
// Usage: tracks <disk.adf>
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"stupidcoder.com/tools/amiga/adf"
	"stupidcoder.com/tools/amiga/hunk"
)

// courses maps the course key to its Track filename (case-insensitive lookup).
var courses = []struct{ key, track string }{
	{"practy", "PrcTrack"}, {"beginr", "BegTrack"}, {"interm", "IntTrack"},
	{"aerial", "AerTrack"}, {"silly", "SilTrack"}, {"ultima", "UltTrack"},
}

func u32(b []byte, o uint32) uint32 {
	if int(o)+4 > len(b) {
		return 0
	}
	return uint32(b[o])<<24 | uint32(b[o+1])<<16 | uint32(b[o+2])<<8 | uint32(b[o+3])
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: tracks <disk.adf>")
		os.Exit(2)
	}
	img, err := os.ReadFile(os.Args[1])
	chk(err)
	vol, err := adf.Open(img)
	chk(err)

	// case-insensitive filename -> path
	paths := map[string]string{}
	chk(vol.Walk(func(e adf.Entry) error {
		if !e.IsDir {
			paths[strings.ToLower(e.Name)] = e.Path
		}
		return nil
	}))

	for _, c := range courses {
		p, ok := paths[strings.ToLower(c.track)]
		if !ok {
			fmt.Printf("\n== %s (%s): not found\n", c.key, c.track)
			continue
		}
		d, err := vol.ReadFile(p)
		if err != nil {
			fmt.Printf("\n== %s (%s): %v\n", c.key, c.track, err)
			continue
		}
		prog, err := hunk.Load(d, 0)
		if err != nil {
			fmt.Printf("\n== %s (%s): hunk load: %v\n", c.key, c.track, err)
			continue
		}
		im := prog.Image
		// The Track header is 10 relocated pointers (+0..+$24); each fans out to a
		// per-course gameplay structure. Count the ones we've identified:
		//   +0  -> $9A6  static slope field: +$1A = region count (Marble_Madness.md V§4)
		//   +4  -> $129FC placement table: +0 -> [X][Y][type] list, $FF-terminated
		//   +8  -> $9D4   coarse-zone partition: 5-byte records, $FFFF-terminated
		//   +$14-> $FD2C  anim block: [0] -> [X][Y][scriptPtr] dynamic-region list ($FF)
		listPtr := u32(im, u32(im, 4))
		recs := parsePlacement(im, listPtr)
		nSlope := u16(im, u32(im, 0)+0x1A)
		nDyn := countRecs(im, u32(im, u32(im, 0x14)), 6, func(b []byte) bool { return b[0] == 0xFF })
		nZone := countRecs(im, u32(im, 8), 5, func(b []byte) bool { return b[0] == 0xFF && b[1] == 0xFF })
		byType := map[int]int{}
		for _, r := range recs {
			byType[r[2]]++
		}
		fmt.Printf("\n== %s (%s)  %d objects, %d slope regions, %d dynamic regions, %d coarse zones\n",
			c.key, c.track, len(recs), nSlope, nDyn, nZone)
		fmt.Printf("   per-type counts: %s\n", typeHist(byType))
		fmt.Printf("   %-4s %-4s %-4s   %-9s\n", "idx", "X", "Y", "type")
		for i, r := range recs {
			fmt.Printf("   %-4d %-4d %-4d   %d   (px %d,%d)\n", i, r[0], r[1], r[2], r[0]*8+4, r[1]*8+4)
		}
	}
}

func u16(b []byte, o uint32) int {
	if int(o)+2 > len(b) {
		return 0
	}
	return int(b[o])<<8 | int(b[o+1])
}

// countRecs counts fixed-size records starting at off, stopping when end() is
// true for the next record (the array's terminator).
func countRecs(im []byte, off uint32, size int, end func([]byte) bool) int {
	n := 0
	for int(off)+size <= len(im) && n < 1000 {
		if end(im[off:]) {
			break
		}
		n++
		off += uint32(size)
	}
	return n
}

// parsePlacement reads 3-byte [X][Y][type] records until a leading $FF.
func parsePlacement(im []byte, off uint32) [][3]int {
	var out [][3]int
	for int(off)+3 <= len(im) {
		if im[off] == 0xFF {
			break
		}
		out = append(out, [3]int{int(im[off]), int(im[off+1]), int(im[off+2])})
		off += 3
	}
	return out
}

func typeHist(m map[int]int) string {
	ks := make([]int, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	parts := make([]string, len(ks))
	for i, k := range ks {
		parts[i] = fmt.Sprintf("type%d=%d", k, m[k])
	}
	return strings.Join(parts, "  ")
}

func chk(e error) {
	if e != nil {
		panic(e)
	}
}
