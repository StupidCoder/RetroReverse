// tracks.go is the tracks stage: it decodes the eight circuits and writes each as its
// own tracks/<slug>.json — the SAME per-circuit object cmd/trackjson emits (as an array
// element). Geometry comes from package track (the verified pure-Go spine, Part IV §5) —
// the plan-view node X,Z per section, plus the absolute per-rung rail geometry baked by
// the engine's own $65BEC model. The decode below is copied from cmd/trackjson so that
// command stays standalone; keep the two in sync if the track format changes.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/stunt-car-racer-amiga/extract/track"
)

// trackNames is the eight circuits in engine order (trackid 0..7). Source: the game's
// own circuit list (Part IV); names are unique so their slugs are stable file stems.
var trackNames = []string{
	"Little Ramp", "Stepping Stones", "Hump Back", "Big Ramp",
	"Ski Jump", "Draw Bridge", "High Jump", "Roller Coaster",
}

// outTrack is one circuit's exported JSON — identical in shape to the array element
// cmd/trackjson writes, now emitted one-per-file.
type outTrack struct {
	Name      string `json:"name"`
	Sections  int    `json:"sections"`
	FinishIdx int    `json:"finishIdx"`
	// [planX, planY, height, bank, type, p1, p2, attr] per section. planX/planY = the
	// 16x16 grid footprint; height = surface elevation; bank = camber.
	Nodes [][]int `json:"nodes"`
	// Per-section ABSOLUTE per-rung geometry: rungs[i][k] = [Lx,Lz,Rx,Rz,Lh,Rh,flags].
	// Plan coordinates are the engine's own baked model ($65BEC), placed at the section's
	// 16x16 grid cell exactly as the per-frame draw does (one cell = $800): world =
	// cell*$800 + local. Heights are the full-precision rail heights. Verified byte-exact
	// against the engine running the real $65BEC (cmd/modeloracle). flags: 1 = edge, 2 =
	// hidden, 4 = crease, 8 = finish.
	Rungs [][][]int `json:"rungs"`
}

// decodeTrack decodes one circuit (trackid id) into its exported JSON object.
func decodeTrack(im *track.Image, id int, name string) outTrack {
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
	return outTrack{name, t.Sections, t.FinishIdx, ns, rungs}
}

func exportTracks(inPath, outDir string) []ViewIndex {
	img, err := os.ReadFile(inPath)
	chk(err)
	im := track.New(img)

	tracksDir := filepath.Join(outDir, "tracks")
	chk(os.MkdirAll(tracksDir, 0o755))

	var idx []ViewIndex
	for id, name := range trackNames {
		tim := im
		if id == track.DrawBridgeTrack {
			// the bridge profiles on disk are placeholders; the game patches
			// them before anything draws ($5A794) — export the same static
			// pose the game's first preview shows (phase 1)
			tim = im.Drawbridge(1)
		}
		ot := decodeTrack(tim, id, name)
		stem := slug(name)
		file := filepath.Join("tracks", stem+".json")
		writeJSON(filepath.Join(outDir, file), ot)

		idx = append(idx, ViewIndex{
			Name: name, Section: "Circuits", Kind: "stunt-track",
			File: filepath.ToSlash(file),
		})
		fmt.Fprintf(os.Stderr, "[tracks] %d/%d  %-20s %2d sections finish@%d\n",
			id+1, len(trackNames), stem, ot.Sections, ot.FinishIdx)
	}
	fmt.Fprintf(os.Stderr, "[tracks] done: %d circuits\n", len(idx))
	return idx
}
