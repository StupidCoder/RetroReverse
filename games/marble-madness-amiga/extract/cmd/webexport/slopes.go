// slopes.go is the slopes stage: each course's 3-D height field becomes ONE
// model3d object GLB (Marble_Madness.md Part V §4-5) — the solid triangulated
// terrain coloured by height band, with the Track-layer markers baked in as
// coloured LINE geometry: a small pin cross per marker and a polyline per
// marble/slinky route (spawn pin at each route's start). The mesh geometry
// reproduces the old viewer's slopes.js _buildMesh exactly.
package main

import (
	"fmt"
	"strings"

	"retroreverse.com/games/marble-madness-amiga/extract/slope"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/amiga/adf"
	"retroreverse.com/tools/platform/amiga/hunk"
)

// heightScale is the viewer's slope-mesh vertical exaggeration (slopes.js
// HEIGHT_SCALE); reproduced here so the GLB frames identically.
const heightScale = 0.15

// numBands is the number of height quantisation bands (one TriGroup / material per
// non-empty band); slopes.js colours per-vertex continuously, we bake it per-tri.
const numBands = 16

// grid is the dense per-course height field the viewer meshes: a w×h row-major grid
// from (x0,y0) where each cell is the tile's height above lo, +1 (0 marks a pit / no
// surface). lo/hi give the real range.
type grid struct {
	x0, y0, w, h, lo, hi int
	heights              []int
}

// markerJSON is the course's Track-layer overlays in tile coords (single pins and
// patrol routes, each with an RGB colour int) — the intermediate the world-space
// markers sidecar is projected from.
type markerJSON struct {
	Points []ptJSON
	Paths  []pathJSON
}
type ptJSON struct{ X, Y, C int }
type pathJSON struct {
	C   int
	Pts [][2]int
}

// worldMarkers is the per-course markers sidecar: route markers with 3-D world
// positions (the viewer no longer carries the height field). Pins = single objects
// plus each path's spawn point; paths = the route polylines.
type worldMarkers struct {
	Pins  []pinJSON   `json:"pins"`
	Paths []pathWJSON `json:"paths"`
}
type pinJSON struct {
	Pos   [3]float32 `json:"pos"`
	Color string     `json:"color"`
}
type pathWJSON struct {
	Points [][3]float32 `json:"points"`
	Color  string       `json:"color"`
}

// marker colours, matching the offline *.wire.png overlays.
const (
	colPlacement = 0x46d4ff // cyan
	colOoze      = 0xff9430 // orange
	colDynRegion = 0xffe000 // yellow
	colMarble    = 0xff46c8 // magenta
	colSlinky    = 0x46e05a // green
)

// exportSlopes writes each course's slope as one model3d object: terrain
// triangles plus the markers as coloured line pins/routes, all in one GLB.
func exportSlopes(ctx *cli.Context, vol *adf.Volume, paths map[string]string) error {
	b := ctx.Builder
	for idx, c := range courses {
		tp, ok := paths[strings.ToLower(c.track)]
		if !ok {
			return fmt.Errorf("%s not found on disk", c.track)
		}
		td, err := vol.ReadFile(tp)
		if err != nil {
			return err
		}
		prog, err := hunk.Load(td, 0)
		if err != nil {
			return fmt.Errorf("%s: hunk load: %w", c.track, err)
		}

		g := buildGrid(slope.Build(prog.Image))
		mk := buildMarkers(slope.Markers(prog.Image))
		positions, groups := buildTerrain(g)
		wm := buildWorldMarkers(g, mk)

		// Bake the markers as LINE geometry sharing the positions array:
		// a small cross + upright per pin, a raised polyline per route,
		// one line material per marker colour.
		byColor := map[string][][2]uint32{}
		addSeg := func(color string, a, b [3]float32) {
			i := uint32(len(positions))
			positions = append(positions, a, b)
			byColor[color] = append(byColor[color], [2]uint32{i, i + 1})
		}
		for _, p := range wm.Pins {
			x, y, z := p.Pos[0], p.Pos[1]+0.05, p.Pos[2]
			addSeg(p.Color, [3]float32{x - 0.5, y, z}, [3]float32{x + 0.5, y, z})
			addSeg(p.Color, [3]float32{x, y, z - 0.5}, [3]float32{x, y, z + 0.5})
			addSeg(p.Color, [3]float32{x, y, z}, [3]float32{x, y + 1.2, z})
		}
		for _, p := range wm.Paths {
			for i := 1; i < len(p.Points); i++ {
				a, bb := p.Points[i-1], p.Points[i]
				a[1] += 0.08
				bb[1] += 0.08
				addSeg(p.Color, a, bb)
			}
		}
		var lines []glb.LineGroup
		for _, col := range []string{"#46d4ff", "#ffe000", "#ff9430", "#ff46c8", "#46e05a"} {
			if segs := byColor[col]; len(segs) > 0 {
				lines = append(lines, glb.LineGroup{Lines: segs, Color: hexRGB(col)})
			}
		}

		id := c.key + "-slope"
		p, err := b.Path("objects", id+".glb")
		if err != nil {
			return err
		}
		if err := glb.WriteMixed(p, positions, groups, lines); err != nil {
			return err
		}
		name := c.name + " slope"
		b.AddObject(schema.Asset{ID: id, Name: name, Group: "Slopes"}, &schema.Object{
			Type: schema.ObjectModel3D, Name: name, Model: id + ".glb",
			Props: map[string]any{
				"course": c.name,
				"legend": map[string]string{
					"#46d4ff cyan":    "scenery placements",
					"#ff9430 orange":  "ooze spawns",
					"#ffe000 yellow":  "dynamic regions",
					"#ff46c8 magenta": "marble routes (spawn pin at the start)",
					"#46e05a green":   "slinky routes (spawn pin at the start)",
				},
			},
		})

		nTri := 0
		for _, grp := range groups {
			nTri += len(grp.Tris)
		}
		ctx.Progress("objects", idx+1, len(courses),
			fmt.Sprintf("%-12s slope  %dx%d, %d tris, %d pins %d paths", c.name, g.w, g.h, nTri, len(wm.Pins), len(wm.Paths)))
	}
	return nil
}

// hexRGB parses "#rrggbb" into 0..1 floats.
func hexRGB(s string) [3]float32 {
	var r, g, bl int
	fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &bl)
	return [3]float32{float32(r) / 255, float32(g) / 255, float32(bl) / 255}
}

// buildGrid flattens a slope field into the dense grid the viewer meshes: each cell
// is the tile's height above lo, +1 (so 0 marks a pit / no surface).
func buildGrid(f slope.Field) grid {
	w, h := f.MaxX-f.MinX+1, f.MaxY-f.MinY+1
	heights := make([]int, w*h)
	for ty := f.MinY; ty <= f.MaxY; ty++ {
		for tx := f.MinX; tx <= f.MaxX; tx++ {
			if hv, ok := f.H[[2]int{tx, ty}]; ok && hv > 8000 {
				heights[(ty-f.MinY)*w+(tx-f.MinX)] = hv - f.Lo + 1
			}
		}
	}
	return grid{x0: f.MinX, y0: f.MinY, w: w, h: h, lo: f.Lo, hi: f.Hi, heights: heights}
}

// geom captures the viewer's coordinate mapping (slopes.js _buildMesh): the axis
// swap (tile-Y -> world-X, tile-X -> world-Z) and the surface height.
type geom struct {
	g      grid
	cx, cz float32
}

func newGeom(g grid) geom {
	return geom{g: g, cx: float32(g.w-1) / 2, cz: float32(g.h-1) / 2}
}
func (m geom) present(gx, gy int) bool {
	return gx >= 0 && gy >= 0 && gx < m.g.w && gy < m.g.h && m.g.heights[gy*m.g.w+gx] > 0
}
func (m geom) wX(gx, gy int) float32 { return float32(gy) - m.cz }
func (m geom) wZ(gx, gy int) float32 { return float32(gx) - m.cx }
func (m geom) surfY(gx, gy int) float32 {
	return float32(m.g.heights[gy*m.g.w+gx]-1) * heightScale
}

// buildTerrain reproduces slopes.js _buildMesh as a SOLID triangulated terrain:
// one vertex per present cell (axis-swapped), two tris per cell whose 4 corners are
// all present, grouped by a height band into one TriGroup (one unlit colour) each.
func buildTerrain(g grid) ([][3]float32, []glb.TriGroup) {
	m := newGeom(g)
	w, h := g.w, g.h

	// Compact vertex list over present cells (matches the viewer's fill mesh), plus
	// the raw grid height per vertex for band classification.
	vidx := make([]int, w*h)
	for i := range vidx {
		vidx[i] = -1
	}
	var positions [][3]float32
	var vHeight []int
	for gy := 0; gy < h; gy++ {
		for gx := 0; gx < w; gx++ {
			if !m.present(gx, gy) {
				continue
			}
			vidx[gy*w+gx] = len(positions)
			positions = append(positions, [3]float32{m.wX(gx, gy), m.surfY(gx, gy), m.wZ(gx, gy)})
			vHeight = append(vHeight, g.heights[gy*w+gx])
		}
	}

	rng := g.hi - g.lo
	if rng < 1 {
		rng = 1
	}
	// One triangle bucket per band; a triangle's band comes from its average
	// normalised height t = (avgHeight-1)/range.
	buckets := make([][][3]uint32, numBands)
	classify := func(a, b, c uint32) {
		avg := float64(vHeight[a]+vHeight[b]+vHeight[c]) / 3
		t := (avg - 1) / float64(rng)
		band := int(t * numBands)
		if band < 0 {
			band = 0
		}
		if band >= numBands {
			band = numBands - 1
		}
		buckets[band] = append(buckets[band], [3]uint32{a, b, c})
	}
	for gy := 0; gy < h-1; gy++ {
		for gx := 0; gx < w-1; gx++ {
			a, b := vidx[gy*w+gx], vidx[gy*w+gx+1]
			c, d := vidx[(gy+1)*w+gx], vidx[(gy+1)*w+gx+1]
			if a < 0 || b < 0 || c < 0 || d < 0 {
				continue
			}
			classify(uint32(a), uint32(c), uint32(b)) // tri (a,c,b)
			classify(uint32(b), uint32(c), uint32(d)) // tri (b,c,d)
		}
	}

	var groups []glb.TriGroup
	for band := 0; band < numBands; band++ { // sorted band iteration -> deterministic
		if len(buckets[band]) == 0 {
			continue
		}
		t := (float64(band) + 0.5) / numBands // band-centre through the ramp
		groups = append(groups, glb.TriGroup{Tris: buckets[band], Color: heightRamp(t)})
	}
	return positions, groups
}

// rampAnchor is one control point of the height ramp (slopes.js heightRamp): a
// normalised height and its RGB in 0..255.
type rampAnchor struct {
	t       float64
	r, g, b float64
}

var ramp = []rampAnchor{
	{0.0, 30, 40, 120}, {0.3, 40, 140, 150}, {0.55, 60, 170, 80}, {0.78, 220, 205, 70}, {1.0, 250, 250, 250},
}

// heightRamp maps t in [0,1] to a blue(low)..white(high) colour (0..1 floats),
// linearly between anchors — matching slopes.js heightRamp / the offline renderer.
func heightRamp(t float64) [3]float32 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	for i := 0; i < len(ramp)-1; i++ {
		if t <= ramp[i+1].t {
			a, b := ramp[i], ramp[i+1]
			f := (t - a.t) / (b.t - a.t + 1e-9)
			return [3]float32{
				float32((a.r + (b.r-a.r)*f) / 255),
				float32((a.g + (b.g-a.g)*f) / 255),
				float32((a.b + (b.b-a.b)*f) / 255),
			}
		}
	}
	return [3]float32{250.0 / 255, 250.0 / 255, 250.0 / 255}
}

// buildMarkers flattens the parsed Track-layer markers into the tile-space
// intermediate (points + routes, each carrying a colour int).
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

// buildWorldMarkers projects the tile-space markers into the GLB's world space
// (same wX/wZ/surfY as buildTerrain): pins = single objects plus each path's first
// point (the spawn), paths = the route polylines. A marker off the surface sits at
// y=0 (matching slopes.js _buildMarkers).
func buildWorldMarkers(g grid, mk markerJSON) worldMarkers {
	m := newGeom(g)
	world := func(x, y int) [3]float32 {
		gx, gy := x-g.x0, y-g.y0
		surf := float32(0)
		if m.present(gx, gy) {
			surf = m.surfY(gx, gy)
		}
		return [3]float32{m.wX(gx, gy), surf, m.wZ(gx, gy)}
	}
	hex := func(c int) string { return fmt.Sprintf("#%06x", c&0xffffff) }

	wm := worldMarkers{Pins: []pinJSON{}, Paths: []pathWJSON{}}
	for _, p := range mk.Points {
		wm.Pins = append(wm.Pins, pinJSON{Pos: world(p.X, p.Y), Color: hex(p.C)})
	}
	for _, p := range mk.Paths {
		pts := make([][3]float32, 0, len(p.Pts))
		for _, q := range p.Pts {
			pts = append(pts, world(q[0], q[1]))
		}
		wm.Paths = append(wm.Paths, pathWJSON{Points: pts, Color: hex(p.C)})
		if len(p.Pts) > 0 { // spawn pin at the route's first point
			wm.Pins = append(wm.Pins, pinJSON{Pos: world(p.Pts[0][0], p.Pts[0][1]), Color: hex(p.C)})
		}
	}
	return wm
}
