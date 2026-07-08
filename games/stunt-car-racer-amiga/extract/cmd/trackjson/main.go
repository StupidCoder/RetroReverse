// trackjson exports the eight decoded circuits to JSON for the web level viewer.
// Geometry comes from package track (the verified pure-Go spine, Part IV §5) — the
// plan-view node X,Z per section, plus the raw section fields for reference. The
// output is written to site/public/stuntcar/tracks.json by default.
//
// Usage: trackjson game.dec.bin [-out path]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"retroreverse.com/games/stunt-car-racer-amiga/extract/track"
)

var names = []string{
	"Little Ramp", "Stepping Stones", "Hump Back", "Big Ramp",
	"Ski Jump", "Draw Bridge", "High Jump", "Roller Coaster",
}

type outTrack struct {
	Name      string  `json:"name"`
	Sections  int     `json:"sections"`
	FinishIdx int     `json:"finishIdx"`
	// [planX, planY, height, bank, type, p1, p2, attr] per section. planX/planY = the
	// 16x16 grid footprint; height = surface elevation; bank = camber.
	Nodes [][]int `json:"nodes"`
	// Per-section ABSOLUTE per-rung geometry: rungs[i][k] = [Lx,Lz,Rx,Rz,Lh,Rh,flags].
	// Plan coordinates are the engine's own baked model ($65BEC): the piece-shape
	// vertex pairs read by $5C6C4, rotated by the section quadrant ($1BBF2 = -$1BC4A,
	// engine order — reversed pieces run backwards and swap rails), placed at the
	// section's 16x16 grid cell exactly as the per-frame draw does (one cell = $800):
	// world = cell*$800 + local. Heights are the full-precision rail heights (profile
	// entry + $1C650/$1C718 base). Verified byte-exact against the engine running the
	// real $65BEC (cmd/modeloracle). flags: 1 = edge (a rung the game's decimated
	// model draws), 2 = hidden (gap-piece end rung, not drawn), 4 = crease, 8 = finish.
	Rungs [][][]int `json:"rungs"`
}

func main() {
	out := flag.String("out", "../../site/public/stuntcar/tracks.json", "output JSON path")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: trackjson game.dec.bin [-out path]")
		os.Exit(2)
	}
	img, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "trackjson:", err)
		os.Exit(1)
	}
	im := track.New(img)
	var tracks []outTrack
	for id, name := range names {
		t := im.Spine(id)
		geom := im.Geometry(&t)
		ns := make([][]int, len(t.Nodes))
		rungs := make([][][]int, len(t.Nodes))
		for i, n := range t.Nodes {
			ns[i] = []int{n.PlanX, n.PlanY, n.Height, n.Bank, n.Type, n.P1, n.P2, n.Attr}
			rs := make([][]int, len(geom[i]))
			for k, r := range geom[i] {
				fl := 0
				if r.Edge {
					fl |= 1
				}
				if r.Hidden {
					fl |= 2
				}
				if r.Crease {
					fl |= 4
				}
				if r.Finish {
					fl |= 8
				}
				rs[k] = []int{r.LX, r.LZ, r.RX, r.RZ, r.HL, r.HR, fl}
			}
			rungs[i] = rs
		}
		tracks = append(tracks, outTrack{name, t.Sections, t.FinishIdx, ns, rungs})
	}
	b, _ := json.Marshal(tracks)
	if err := os.WriteFile(*out, b, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "trackjson:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d tracks)\n", *out, len(tracks))
}
