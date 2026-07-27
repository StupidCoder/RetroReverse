// webexport serializes the decoded Fort Apocalypse levels as a Retro-X game
// tree (see RETROX.md). Everything is reconstructed from the pre-extracted
// game file ($7000 image) by the same decode path as the inspection tools and
// handed to the shared builder (tools/lib/retrox):
//
//	manifest.json                  game meta + asset index (written by the builder)
//	levels/<id>.json               per level: char grid, soft-char tile anims,
//	                               tank placements, randomized pools
//	levels/<id>/atlas.png          the 128 playfield chars in the level's colours
//	                               plus the soft-char animation frames, 16 wide
//	objects/<id>.json|.png         sprite2d objects: the tanks/prisoners/mines as
//	                               regular sprite sheets (opaque black background,
//	                               exactly as they appear stamped into the map) and
//	                               the two helicopter hardware sprites (white on
//	                               transparent; instances carry the tint)
//
// The playfield is a horizontal cylinder (tilemap wrap "x"): column 214 joins
// back to column 0. The game seeds prisoners/mines/the enemy helicopter at
// level load; those are pools the viewer re-rolls (seedable). The six tanks
// spawn at every fixed home, so they are ordinary placements. Pool instances
// are static in Retro-X v1 (the engine's patrol movement is not replayed);
// the walk/creep art still animates in place. No music (C64, no in-engine
// track here); no oracle.
//
// Usage (from games/fort-apocalypse-c64/): go run ./extract/cmd/webexport
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"

	"retroreverse.com/games/fort-apocalypse-c64/extract/fortgfx"
	"retroreverse.com/tools/lib/retrox/atlas"
	"retroreverse.com/tools/lib/retrox/build"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/c64/gfx"
)

var levels = []struct {
	id, name string
}{
	{"vaults", "Vaults of Draconis"},
	{"caves", "Crystalline Caves"},
}

// stamp is one char of a composed object: char code at a cell offset from the
// object's anchor (world position), in pixels.
type stamp struct{ dx, dy, ch int }

// Char-composed object art (Part V): the engines stamp these into the map
// buffer; here they become regular sprite frames on an opaque black
// background, which looks identical over the black playfield.
var (
	prisonerRight = [][]stamp{
		{{0, -8, 0x49}, {0, 0, 0x3B}},
		{{0, -8, 0x49}, {0, 0, 0x3C}},
	}
	prisonerLeft = [][]stamp{
		{{0, -8, 0x4A}, {0, 0, 0x3E}},
		{{0, -8, 0x4A}, {0, 0, 0x3D}},
	}
	tankAimLeft  = []stamp{{0, 0, 0x6C}, {8, 0, 0x6D}, {16, 0, 0x6E}, {8, -8, 0x6F}}
	tankAimRight = []stamp{{0, 0, 0x6C}, {8, 0, 0x6D}, {16, 0, 0x6E}, {8, -8, 0x70}}
	// The $963C char pairs are the SAME 26-pixel mine at four sub-cell
	// positions (blob centres 4.5/6.5/8.5/10.5 px within the 2-cell pair) —
	// the engine's slide phases become a ping-pong creep animation.
	minePhases = [][]stamp{
		{{0, 0, 0x40}},
		{{0, 0, 0x5B}, {8, 0, 0x5C}},
		{{0, 0, 0x5D}, {8, 0, 0x5E}},
		{{8, 0, 0x5F}},
	}
)

func main() {
	cli.Main("fort-apocalypse-c64", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		ctx.In = "extracted/FORT-fast-7000.prg" // the staged $7000 image (game dir)
	}
	game, err := fortgfx.LoadGame(ctx.In)
	if err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("Fort Apocalypse")
	b.SetPlatform("Commodore 64")
	b.SetYear(1982)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 320, H: 200},
		TickHz: 50,
		Filter: "crt",
	})

	var refs map[string]string // object key -> asset id
	if ctx.Stage("objects") {
		if refs, err = exportObjects(ctx, game); err != nil {
			return err
		}
	}
	if ctx.Stage("levels") {
		if err := exportLevels(ctx, game, refs); err != nil {
			return err
		}
	}
	return nil
}

// exportObjects writes the sprite2d objects. The char-composed ones
// (prisoner/tank/mine) are rendered per level because the multicolor slot 01
// is the per-level $D022 colour; the helicopter hardware sprites are shared
// (white art, tinted per instance like the VIC colours them).
func exportObjects(ctx *cli.Context, game *fortgfx.Game) (map[string]string, error) {
	b := ctx.Builder
	refs := map[string]string{}
	cs := game.PlayfieldCharset()

	for li, lv := range levels {
		pal := palette(game.MulticolorValue(li))

		frame := func(stamps []stamp) atlas.Frame {
			minX, minY, maxX, maxY := 0, 0, 0, 0
			for _, s := range stamps {
				minX, minY = min(minX, s.dx), min(minY, s.dy)
				maxX, maxY = max(maxX, s.dx+8), max(maxY, s.dy+8)
			}
			img := image.NewRGBA(image.Rect(0, 0, maxX-minX, maxY-minY))
			for y := range img.Rect.Dy() { // opaque black background = the playfield
				for x := range img.Rect.Dx() {
					img.SetRGBA(x, y, color.RGBA{0, 0, 0, 255})
				}
			}
			for _, s := range stamps {
				gfx.DrawChar(img, cs[s.ch*8:s.ch*8+8], s.dx-minX, s.dy-minY, 1, pal)
			}
			return atlas.Frame{Image: img, Anchor: image.Pt(-minX, -minY)}
		}
		frames := func(phases [][]stamp) []atlas.Frame {
			out := make([]atlas.Frame, len(phases))
			for i, p := range phases {
				out[i] = frame(p)
			}
			return out
		}

		add := func(key, name string, anims []atlas.Animation, meta []schema.Animation, desc string) error {
			id := key + "-" + lv.id
			packed, err := atlas.Pack(anims)
			if err != nil {
				return err
			}
			// Opaque black everywhere, including cell padding: stamped into the
			// map these chars sit on the black playfield, so a fully black cell
			// reproduces the in-game look exactly.
			blackout(packed.Image)
			f, err := b.CreateFile("objects", id+".png")
			if err != nil {
				return err
			}
			if err := packed.EncodePNG(f); err != nil {
				f.Close()
				return err
			}
			f.Close()
			doc := &schema.Object{
				Type: schema.ObjectSprite2D,
				Name: name,
				Atlas: &schema.SpriteAtlas{
					File: id + ".png", CellW: packed.CellW, CellH: packed.CellH,
					Anchor: []int{packed.Anchor.X, packed.Anchor.Y},
				},
				Animations: meta,
				Props:      map[string]any{"level": lv.name},
			}
			b.AddObject(schema.Asset{ID: id, Name: name, Group: lv.name, Description: desc}, doc)
			refs[fmt.Sprintf("%d/%s", li, key)] = id
			return nil
		}

		if err := add("prisoner", "Prisoner",
			[]atlas.Animation{
				{ID: "walk", Frames: frames(prisonerRight)},
				{ID: "walk-left", Frames: frames(prisonerLeft)},
			},
			[]schema.Animation{
				{ID: "walk", Name: "Walk (right)", Row: 0, Frames: 2, Loop: "loop",
					Durations: []int{20, 20}, Mirror: "walk-left",
					Description: "The walkway run: torso $49 over alternating legs $3B/$3C, one step every ~20 frames."},
				{ID: "walk-left", Name: "Walk (left)", Row: 1, Frames: 2, Loop: "loop",
					Durations: []int{20, 20},
					Description: "The left-facing art ($4A + $3E/$3D) the engine draws after a turn."},
			}, ""); err != nil {
			return nil, err
		}
		if err := add("tank", "Tank",
			[]atlas.Animation{
				{ID: "main", Frames: []atlas.Frame{frame(tankAimLeft)}},
				{ID: "aim-right", Frames: []atlas.Frame{frame(tankAimRight)}},
			},
			[]schema.Animation{
				{ID: "main", Name: "Turret left", Row: 0, Frames: 1, Loop: "hold",
					Description: "Body $6C $6D $6E with the $6F turret; in play the turret always aims at the player."},
				{ID: "aim-right", Name: "Turret right", Row: 1, Frames: 1, Loop: "hold",
					Description: "The mirrored $70 turret pose."},
			}, ""); err != nil {
			return nil, err
		}
		if err := add("mine", "Self-Propelled Mine",
			[]atlas.Animation{{ID: "creep", Frames: frames(minePhases)}},
			[]schema.Animation{
				{ID: "creep", Name: "Creep", Row: 0, Frames: 4, Loop: "pingpong",
					Durations: []int{7, 6, 7, 6},
					Description: "The four $963C sub-cell phases — the mine slides 2 px per phase " +
						"through its 2-cell window (~6.5 frames per phase), reversing at the ends."},
			}, ""); err != nil {
			return nil, err
		}
		ctx.Progress("objects", li+1, len(levels), lv.name+" char objects")
	}

	// The helicopter hardware sprites, shared by both levels: white on
	// transparent (the VIC colours the sprite; instances carry the tint).
	poses := game.HelicopterPoses()
	shapes := game.SpriteShapes()
	heli := func(id, name, desc string, block []byte) error {
		img := chopperImg(block)
		f, err := b.CreateFile("objects", id+".png")
		if err != nil {
			return err
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			return err
		}
		f.Close()
		b.AddObject(schema.Asset{ID: id, Name: name, Group: "Helicopters", Description: desc}, &schema.Object{
			Type:  schema.ObjectSprite2D,
			Name:  name,
			Atlas: &schema.SpriteAtlas{File: id + ".png", CellW: 32, CellH: 18},
			Animations: []schema.Animation{
				{ID: "main", Frames: 1, Loop: "hold"},
			},
			Props: map[string]any{"art": "hardware sprite, white; instances are tinted like the VIC colours them"},
		})
		refs[id] = id
		return nil
	}
	// poses are in tilt order full-left .. level .. full-right; [0] is the
	// sprite block of each pose's first rotor frame.
	if err := heli("chopper", "Rocket Copter", "", shapes[poses[len(poses)/2][0]-1]); err != nil {
		return nil, err
	}
	if err := heli("enemy-helicopter", "Enemy Helicopter", "", shapes[poses[0][0]-1]); err != nil {
		return nil, err
	}
	return refs, nil
}

// exportLevels writes the two level documents + per-level atlases.
func exportLevels(ctx *cli.Context, game *fortgfx.Game, refs map[string]string) error {
	b := ctx.Builder
	cs := game.PlayfieldCharset()
	anim := game.SoftCharAnim()

	for li, lvMeta := range levels {
		lm, err := game.LevelMap(li)
		if err != nil {
			return err
		}

		// Atlas tiles: the 128 base chars at fixed indices 0..127, then any
		// extra frame bitmaps the animations need, appended and de-duplicated.
		tiles := make([][8]byte, 128)
		for ch := 0; ch < 128; ch++ {
			copy(tiles[ch][:], cs[ch*8:])
		}
		idxOf := map[[8]byte]int{}
		for i := 127; i >= 0; i-- {
			idxOf[tiles[i]] = i
		}
		addTile := func(fr [8]byte) int {
			if i, ok := idxOf[fr]; ok {
				return i
			}
			i := len(tiles)
			tiles = append(tiles, fr)
			idxOf[fr] = i
			return i
		}
		var tileAnims []schema.TileAnim
		for _, a := range anim {
			ta := schema.TileAnim{Tiles: []int{int(a.Char)}, PeriodFrames: a.Period}
			for _, fr := range a.Frames {
				ta.Frames = append(ta.Frames, []int{addTile(fr)})
			}
			tileAnims = append(tileAnims, ta)
		}
		pal := palette(game.MulticolorValue(li))
		if err := writeAtlas(b, tiles, pal, "levels", lvMeta.id, "atlas.png"); err != nil {
			return err
		}

		// Cells: the 215 content columns (Part IV §4). The playfield is a
		// cylinder — column 214 joins back to column 0 — so the stored
		// wrap-seam column (a duplicate of column 0) is dropped; the viewer
		// wraps instead.
		w, h := fortgfx.ContentWidth, fortgfx.MapHeight
		cells := make([]int, w*h)
		for r := 0; r < h; r++ {
			for c := 0; c < w; c++ {
				cells[r*w+c] = int(lm.Cells[r][c])
			}
		}
		cellAt := func(c, r int) int { return cells[r*w+c] }

		sx, sy := lm.PlayerSpawn.Col*8, lm.PlayerSpawn.Row*8
		doc := &schema.Level{
			Type: schema.LevelTilemap,
			Camera: &schema.Camera{Mode: "map2d",
				// never zoom out past one cylinder period (objects would repeat)
				Map2D: &schema.Map2D{MinFitFactor: 1, MaxNativeFactor: 3}},
			Tilemap: &schema.Tilemap{
				TileSize: 8, Width: w, Height: h,
				Atlas:     schema.TileAtlas{File: lvMeta.id + "/atlas.png", Cols: 16},
				Cells:     cells,
				Wrap:      "x",
				View:      &schema.Rect{X: sx + 16 - 160, Y: 0, W: 320, H: 200},
				Spawn:     &schema.Spawn{X: sx, Y: sy, Tint: "#b8c76f"},
				TileAnims: tileAnims,
			},
		}
		if refs != nil {
			doc.Tilemap.Spawn.Object = refs["chopper"]
			doc.Tilemap.Spawn.Anim = "main"

			// The six tanks spawn at every fixed home — ordinary placements.
			obst := game.ObstacleChars()
			for i, p := range lm.TankHomes {
				l, r := lm.TankRange(obst, p.Col, p.Row)
				doc.Placements = append(doc.Placements, schema.Placement{
					ID:     i,
					Object: refs[fmt.Sprintf("%d/tank", li)],
					Anim:   "main",
					Pos:    []float64{float64(p.Col * 8), float64(p.Row * 8)},
					Name:   "Tank",
					Props: map[string]any{
						"home":      []int{p.Col, p.Row},
						"patrolPx":  []int{l * 8, r * 8},
						"behaviour": "patrols its span in lockstep with the other tanks (not replayed)",
					},
				})
			}

			// Pools: the game's load-time seeding rules, as candidates the
			// viewer re-rolls (seedable). Spans and cadences live in the
			// object/animation descriptions; instances are static in v1.
			var prisoners [][]float64
			for _, p := range lm.PrisonerSpawns {
				prisoners = append(prisoners, []float64{float64(p.Col * 8), float64(p.Row * 8)})
			}
			doc.Pools = append(doc.Pools, schema.Pool{
				ID: "prisoners", Count: 8,
				Object:     refs[fmt.Sprintf("%d/prisoner", li)],
				Anim:       "walk",
				Candidates: prisoners,
				Seedable:   true,
			})

			// SPMs: empty 2-cell spots in the column band $2D..$C8; 13 at
			// base difficulty (26/39 by Pilot Skill).
			var spmSpots [][]float64
			for r := 0; r < h; r++ {
				for c := fortgfx.SPMBandMin; c <= fortgfx.SPMBandMax && c+1 < w; c++ {
					if cellAt(c, r) == 0 && cellAt(c+1, r) == 0 {
						spmSpots = append(spmSpots, []float64{float64(c * 8), float64(r * 8)})
					}
				}
			}
			doc.Pools = append(doc.Pools, schema.Pool{
				ID: "mines", Count: 13,
				Object:     refs[fmt.Sprintf("%d/mine", li)],
				Anim:       "creep",
				Candidates: spmSpots,
				Seedable:   true,
			})

			// One enemy helicopter at a random 4x2-clear spot in the band.
			var heliSpots [][]float64
			for r := 0; r+1 < h; r++ {
				for c := fortgfx.SPMBandMin; c <= fortgfx.SPMBandMax && c+3 < w; c++ {
					clear := true
					for dy := 0; dy < 2 && clear; dy++ {
						for dx := 0; dx < 4; dx++ {
							if cellAt(c+dx, r+dy) != 0 {
								clear = false
								break
							}
						}
					}
					if clear {
						heliSpots = append(heliSpots, []float64{float64(c * 8), float64(r * 8)})
					}
				}
			}
			doc.Pools = append(doc.Pools, schema.Pool{
				ID: "enemy-helicopter", Count: 1,
				Object:     refs["enemy-helicopter"],
				Anim:       "main",
				Tint:       "#352879",
				Candidates: heliSpots,
				Seedable:   true,
			})
		}

		b.AddLevel(schema.Asset{ID: lvMeta.id, Name: lvMeta.name}, doc)
		ctx.Progress("levels", li+1, len(levels),
			fmt.Sprintf("%-18s %3dx%-2d cells  %d atlas tiles  %d placements  %d pools",
				lvMeta.name, w, h, len(tiles), len(doc.Placements), len(doc.Pools)))
	}
	return nil
}

// blackout replaces every fully transparent pixel with opaque black.
func blackout(img *image.NRGBA) {
	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i+3] == 0 {
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = 0, 0, 0, 255
		}
	}
}

// palette is the playfield's multicolor map: 00=black, 01=$D022 (per level),
// 10=white, 11=colour-RAM green (Part IV §2).
func palette(d022 byte) [4]color.RGBA {
	return [4]color.RGBA{gfx.Palette[0], gfx.Palette[d022&0x0F], gfx.Palette[1], gfx.Palette[5]}
}

// chopperImg renders one helicopter sprite block (the [left][right][$00] rows
// the game expands, 16x18 used pixels) into a 32x18 RGBA image — white where
// set, transparent elsewhere, X-expanded (drawn double-wide in game).
func chopperImg(block []byte) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 32, 18))
	white := color.RGBA{255, 255, 255, 255}
	for r := 0; r < 18; r++ {
		for bi := 0; bi < 2; bi++ {
			bb := block[r*3+bi]
			for bit := 0; bit < 8; bit++ {
				if bb&(0x80>>bit) != 0 {
					x := (bi*8 + bit) * 2
					img.SetRGBA(x, r, white)
					img.SetRGBA(x+1, r, white)
				}
			}
		}
	}
	return img
}

// writeAtlas renders the tiles as a 16-wide grid of 8x8 multicolor chars.
func writeAtlas(b *build.Builder, tiles [][8]byte, pal [4]color.RGBA, rel ...string) error {
	const cols = 16
	rows := (len(tiles) + cols - 1) / cols
	img := image.NewRGBA(image.Rect(0, 0, cols*8, rows*8))
	for i, t := range tiles {
		gfx.DrawChar(img, t[:], (i%cols)*8, (i/cols)*8, 1, pal)
	}
	f, err := b.CreateFile(rel...)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
