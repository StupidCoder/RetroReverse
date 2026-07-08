// slopes.go is the slopes stage: each course's 3-D height field + Track-layer
// markers (Marble_Madness.md Part V §4-5) is written as a standalone
// slopes/<key>.slope.json and indexed in manifest.views[] as a bespoke
// "marble-slope" three.js view. The slopes are DECOUPLED from the levels — no
// level references them.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/marble-madness-amiga/extract/slope"
	"retroreverse.com/tools/platform/amiga/adf"
	"retroreverse.com/tools/platform/amiga/hunk"
)

// slopeJSON is the per-course height field for the 3-D view: a dense grid (w×h,
// row-major from x0,y0) where each cell is the tile's height above lo plus one, or
// 0 where there is no rolling surface (a pit). lo/hi give the real range.
type slopeJSON struct {
	X0      int        `json:"x0"`
	Y0      int        `json:"y0"`
	W       int        `json:"w"`
	H       int        `json:"h"`
	Lo      int        `json:"lo"`
	Hi      int        `json:"hi"`
	Heights []int      `json:"heights"`
	Markers markerJSON `json:"markers"`
}

// markerJSON is the course's Track-layer overlays for the 3-D view: single pins
// (Points) and patrol routes (Paths), each carrying an RGB colour, in tile coords.
type markerJSON struct {
	Points []ptJSON   `json:"points"`
	Paths  []pathJSON `json:"paths"`
}
type ptJSON struct {
	X int `json:"x"`
	Y int `json:"y"`
	C int `json:"c"`
}
type pathJSON struct {
	C   int      `json:"c"`
	Pts [][2]int `json:"pts"`
}

// marker colours, matching the offline *.wire.png overlays.
const (
	colPlacement = 0x46d4ff // cyan
	colOoze      = 0xff9430 // orange
	colDynRegion = 0xffe000 // yellow
	colMarble    = 0xff46c8 // magenta
	colSlinky    = 0x46e05a // green
)

// exportSlopes writes slopes/<key>.slope.json for every course and returns the manifest
// views index (the bespoke "marble-slope" three.js views).
func exportSlopes(vol *adf.Volume, paths map[string]string, outDir string) []ViewIndex {
	slopesDir := filepath.Join(outDir, "slopes")
	chk(os.MkdirAll(slopesDir, 0o755))

	var views []ViewIndex
	for idx, c := range courses {
		tp, ok := paths[strings.ToLower(c.track)]
		if !ok {
			fail(fmt.Errorf("%s not found on disk", c.track))
		}
		td, err := vol.ReadFile(tp)
		chk(err)
		prog, err := hunk.Load(td, 0)
		if err != nil {
			fail(fmt.Errorf("%s: hunk load: %w", c.track, err))
		}

		sj := buildSlope(slope.Build(prog.Image))
		sj.Markers = buildMarkers(slope.Markers(prog.Image))
		file := c.key + ".slope.json"
		chk(writeJSON(filepath.Join(slopesDir, file), sj))

		views = append(views, ViewIndex{
			Name: c.name, Section: "Slopes", File: "slopes/" + file, Kind: "marble-slope",
		})
		fmt.Fprintf(os.Stderr, "[slopes] %d/%d  %-12s %s  %dx%d, h %d..%d, %d points %d paths\n",
			idx+1, len(courses), c.name, file, sj.W, sj.H, sj.Lo, sj.Hi, len(sj.Markers.Points), len(sj.Markers.Paths))
	}
	fmt.Fprintf(os.Stderr, "[slopes] done: %d views\n", len(views))
	return views
}

// buildSlope flattens a slope field into the dense grid the viewer meshes: each cell is
// the tile's height above lo, +1 (so 0 marks a pit / no surface).
func buildSlope(f slope.Field) slopeJSON {
	w, h := f.MaxX-f.MinX+1, f.MaxY-f.MinY+1
	heights := make([]int, w*h)
	for ty := f.MinY; ty <= f.MaxY; ty++ {
		for tx := f.MinX; tx <= f.MaxX; tx++ {
			if hv, ok := f.H[[2]int{tx, ty}]; ok && hv > 8000 {
				heights[(ty-f.MinY)*w+(tx-f.MinX)] = hv - f.Lo + 1
			}
		}
	}
	return slopeJSON{X0: f.MinX, Y0: f.MinY, W: w, H: h, Lo: f.Lo, Hi: f.Hi, Heights: heights}
}

func buildMarkers(m slope.MarkerSet) markerJSON {
	mj := markerJSON{Points: []ptJSON{}, Paths: []pathJSON{}}
	add := func(pts [][2]int, c int) {
		for _, p := range pts {
			mj.Points = append(mj.Points, ptJSON{p[0], p[1], c})
		}
	}
	add(m.Placement, colPlacement)
	add(m.Ooze, colOoze)
	add(m.DynRegion, colDynRegion)
	for _, p := range m.Marbles {
		mj.Paths = append(mj.Paths, pathJSON{colMarble, p})
	}
	for _, p := range m.Slinkies {
		mj.Paths = append(mj.Paths, pathJSON{colSlinky, p})
	}
	return mj
}
