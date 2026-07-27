// levels.go is the levels stage: each course becomes a Retro-X tilemap level
// (Marble_Madness.md Part IV §3/§6). The colour bands are baked into the atlas
// as recoloured variant tiles and the gold shimmer as tileAnims, so the viewer
// needs no palette machinery. The scenery-overlay pieces the placements
// reference are the objects stage's sprite2d assets.
package main

import (
	"fmt"

	"retroreverse.com/games/marble-madness-amiga/extract/mlb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/amiga/adf"
	"retroreverse.com/tools/platform/c64/gfx"
)

// shimmerPeriod is the vblank frames each gold-rotation phase holds: the stepper
// $84DC runs per WORLD frame (2 vblanks each), stepping the phase past 6 — every
// 4th world frame = 8 vblanks.
const shimmerPeriod = 8

// exportLevels writes the level documents + per-course atlases. refs may be
// nil (objects stage disabled): then the levels ship without placements.
func exportLevels(ctx *cli.Context, vol *adf.Volume, paths map[string]string, refs map[string]objRef) error {
	b := ctx.Builder

	for idx, c := range courses {
		cr, err := loadCourse(vol, paths, c.key, c.track)
		if err != nil {
			return err
		}
		co, bake := cr.co, cr.bake

		// The scenery overlays give us the level's object placements + the screen-swap
		// tile animations. render=false: the sprite PNGs come from the objects stage.
		objects, cellAnims, err := exportOverlays(vol, paths, c.key, cr.prog.Image, co, bake.paletteAt, "", nil, false)
		if err != nil {
			return err
		}

		cells := make([]int, mlb.CourseW*co.PlayableH)
		tilesInBand := map[int]map[int]bool{}
		note := func(band, t int) {
			if tilesInBand[band] == nil {
				tilesInBand[band] = map[int]bool{}
			}
			tilesInBand[band][t] = true
		}
		for r := 0; r < co.PlayableH; r++ {
			band := bake.bandAt(r)
			for x := 0; x < mlb.CourseW; x++ {
				t := co.Cells[r*mlb.CourseW+x]
				note(band, t)
				cells[r*mlb.CourseW+x] = bake.tileFor(t, band)
			}
		}
		// swap-variant rows repaint into the finale band: substitute per destination
		// row, and count their tiles toward that band's shimmer.
		for _, ca := range cellAnims {
			ty, tw := ca["ty"].(int), ca["tw"].(int)
			for _, ph := range ca["phases"].([]map[string]any) {
				tiles := ph["tiles"].([]int)
				for i, t := range tiles {
					band := bake.bandAt(ty + i/tw)
					note(band, t)
					tiles[i] = bake.tileFor(t, band)
				}
			}
		}
		tileAnims := bake.shimmerAnims(tilesInBand, shimmerPeriod)

		atlasFile := c.key + ".atlas.png"
		p, err := b.Path("levels", atlasFile)
		if err != nil {
			return err
		}
		if err := gfx.WritePNG(p, co.AtlasVariants(16, bake.ext, bake.varList)); err != nil {
			return err
		}

		// The map shows the PLAYABLE rows (the .mlb header count = the engine's
		// scroll clamp). Data rows beyond it are off-screen variant storage; the
		// swap animation replays them from the full cell array.
		doc := &schema.Level{
			Type:  schema.LevelTilemap,
			Music: c.key + "-theme",
			Tilemap: &schema.Tilemap{
				TileSize: 8, Width: mlb.CourseW, Height: co.PlayableH,
				Atlas: schema.TileAtlas{File: atlasFile, Cols: 16, Gutter: 1},
				Cells: cells,
				// Frame the Amiga's on-screen view (288x200 playfield) at the course top.
				View: &schema.Rect{X: (mlb.CourseW*8 - 288) / 2, Y: 0, W: 288, H: 200},
			},
		}
		for _, ca := range cellAnims {
			sca := schema.CellAnim{
				TX: ca["tx"].(int), TY: ca["ty"].(int),
				TW: ca["tw"].(int), TH: ca["th"].(int),
			}
			for _, ph := range ca["phases"].([]map[string]any) {
				sca.Phases = append(sca.Phases, schema.CellPhase{
					Tiles: ph["tiles"].([]int), Frames: ph["frames"].(int),
				})
			}
			doc.Tilemap.CellAnims = append(doc.Tilemap.CellAnims, sca)
		}
		for _, ta := range tileAnims {
			doc.Tilemap.TileAnims = append(doc.Tilemap.TileAnims, schema.TileAnim{
				Tiles:        ta["tiles"].([]int),
				Frames:       ta["frames"].([][]int),
				PeriodFrames: ta["periodFrames"].(int),
			})
		}
		for i, o := range objects {
			key, _ := o["sprite"].(string)
			ref, ok := refs[key]
			if !ok {
				continue // objects stage disabled, or a piece without art
			}
			pl := schema.Placement{
				ID:     i,
				Object: ref.asset,
				Pos:    []float64{float64(o["x"].(int)), float64(o["y"].(int))},
			}
			if n, ok := o["name"].(string); ok {
				pl.Name = n
			}
			if t, ok := o["type"].(int); ok {
				pl.Props = map[string]any{"region": t}
			}
			doc.Placements = append(doc.Placements, pl)
		}

		b.AddLevel(schema.Asset{ID: c.key, Name: c.name, Group: "Courses"}, doc)
		ctx.Progress("levels", idx+1, len(courses),
			fmt.Sprintf("%-12s %d×%d tiles, %d objects", c.name, mlb.CourseW, co.PlayableH, len(doc.Placements)))
	}
	return nil
}
