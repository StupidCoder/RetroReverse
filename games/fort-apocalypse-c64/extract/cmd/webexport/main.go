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
	"bytes"
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

// charObject is one char-composed object rendered at a level's palette,
// held in memory so identical renders can be deduplicated across levels.
type charObject struct {
	png  []byte
	doc  schema.Object
	name string
}

// exportObjects writes the sprite2d objects. The char-composed ones
// (prisoner/tank/mine) are rendered per level — the multicolor slot 01 is the
// per-level $D022 colour — then DEDUPLICATED: a kind whose pixels come out
// identical in both levels (the mine uses no $D022 pixel) ships once, without
// a level suffix. The helicopter is ONE object: the game has a single craft
// sprite set (7 bank poses x 2 rotor frames); player and enemy instances
// differ only by tint, exactly like the VIC colours the hardware sprite.
func exportObjects(ctx *cli.Context, game *fortgfx.Game) (map[string]string, error) {
	b := ctx.Builder
	refs := map[string]string{}
	cs := game.PlayfieldCharset()

	render := func(li int) (map[string]charObject, error) {
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
		pack := func(name string, anims []atlas.Animation, meta []schema.Animation) (charObject, error) {
			packed, err := atlas.Pack(anims)
			if err != nil {
				return charObject{}, err
			}
			// Opaque black everywhere, including cell padding: stamped into
			// the map these chars sit on the black playfield, so a fully
			// black cell reproduces the in-game look exactly.
			blackout(packed.Image)
			var buf bytes.Buffer
			if err := packed.EncodePNG(&buf); err != nil {
				return charObject{}, err
			}
			return charObject{
				png:  buf.Bytes(),
				name: name,
				doc: schema.Object{
					Type: schema.ObjectSprite2D,
					Name: name,
					Atlas: &schema.SpriteAtlas{
						CellW: packed.CellW, CellH: packed.CellH,
						Anchor: []int{packed.Anchor.X, packed.Anchor.Y},
					},
					Animations: meta,
				},
			}, nil
		}

		out := map[string]charObject{}
		var err error
		if out["prisoner"], err = pack("Prisoner",
			[]atlas.Animation{
				{ID: "walk", Frames: frames(prisonerRight)},
				{ID: "walk-left", Frames: frames(prisonerLeft)},
			},
			[]schema.Animation{
				{ID: "walk", Name: "Walk (right)", Row: 0, Frames: 2, Loop: "loop",
					Durations: []int{20, 20}, Mirror: "walk-left",
					Description: "The walkway run: torso $49 over alternating legs $3B/$3C, one step every ~20 frames. " +
						"The bottom two rows are the walkway char $48's own floor line — the engine stamps the prisoner " +
						"INTO the walkway cell, and the baked floor keeps it continuous."},
				{ID: "walk-left", Name: "Walk (left)", Row: 1, Frames: 2, Loop: "loop",
					Durations:   []int{20, 20},
					Description: "The left-facing art ($4A + $3E/$3D) the engine draws after a turn."},
			}); err != nil {
			return nil, err
		}
		if out["tank"], err = pack("Tank",
			[]atlas.Animation{
				{ID: "main", Frames: []atlas.Frame{frame(tankAimLeft)}},
				{ID: "aim-right", Frames: []atlas.Frame{frame(tankAimRight)}},
			},
			[]schema.Animation{
				{ID: "main", Name: "Turret left", Row: 0, Frames: 1, Loop: "hold",
					Description: "Body $6C $6D $6E with the $6F turret; in play the turret always aims at the player."},
				{ID: "aim-right", Name: "Turret right", Row: 1, Frames: 1, Loop: "hold",
					Description: "The mirrored $70 turret pose."},
			}); err != nil {
			return nil, err
		}
		if out["mine"], err = pack("Self-Propelled Mine",
			[]atlas.Animation{{ID: "creep", Frames: frames(minePhases)}},
			[]schema.Animation{
				{ID: "creep", Name: "Creep", Row: 0, Frames: 4, Loop: "pingpong",
					Durations: []int{7, 6, 7, 6},
					Description: "The four $963C sub-cell phases — the mine slides 2 px per phase " +
						"through its 2-cell window (~6.5 frames per phase), reversing at the ends."},
			}); err != nil {
			return nil, err
		}
		return out, nil
	}

	l0, err := render(0)
	if err != nil {
		return nil, err
	}
	l1, err := render(1)
	if err != nil {
		return nil, err
	}
	emit := func(id string, co charObject, group string) error {
		f, err := b.CreateFile("objects", id+".png")
		if err != nil {
			return err
		}
		if _, err := f.Write(co.png); err != nil {
			f.Close()
			return err
		}
		f.Close()
		doc := co.doc
		doc.Atlas.File = id + ".png"
		b.AddObject(schema.Asset{ID: id, Name: co.name, Group: group}, &doc)
		return nil
	}
	for _, kind := range []string{"prisoner", "tank", "mine"} {
		if bytes.Equal(l0[kind].png, l1[kind].png) {
			if err := emit(kind, l0[kind], "Objects"); err != nil {
				return nil, err
			}
			refs["0/"+kind], refs["1/"+kind] = kind, kind
			ctx.Logf("objects: %s is level-independent (no $D022 pixel) — shared", kind)
			continue
		}
		for li, co := range []charObject{l0[kind], l1[kind]} {
			id := kind + "-" + levels[li].id
			if err := emit(id, co, levels[li].name); err != nil {
				return nil, err
			}
			refs[fmt.Sprintf("%d/%s", li, kind)] = id
		}
	}
	ctx.Progress("objects", 1, 2, "char objects (deduped across levels)")

	// The helicopter: ONE craft, the game's full sprite set — 7 bank poses
	// (full-left .. level .. full-right), each with 2 rotor frames. White on
	// transparent; instances are tinted like the VIC colours the sprite.
	// The rotor flips once per engine main-loop pass (~2.5 frames, the
	// measured gameplay cadence) — exported as alternating 2/3-frame holds.
	poses := game.HelicopterPoses()
	shapes := game.SpriteShapes()
	poseIDs := []string{"left-full", "left-2", "left-1", "level", "right-1", "right-2", "right-full"}
	poseNames := []string{"Full left bank", "Left bank", "Slight left bank", "Level flight",
		"Slight right bank", "Right bank", "Full right bank"}
	var anims []atlas.Animation
	var meta []schema.Animation
	for i, p := range poses {
		var fr []atlas.Frame
		for _, block := range p {
			fr = append(fr, atlas.Frame{Image: chopperImg(shapes[block-1]), Anchor: image.Pt(0, 0)})
		}
		anims = append(anims, atlas.Animation{ID: poseIDs[i], Frames: fr})
		m := schema.Animation{ID: poseIDs[i], Name: poseNames[i], Row: i, Frames: len(p), Loop: "loop"}
		if len(p) == 1 {
			m.Loop = "hold"
		} else {
			m.Durations = make([]int, len(p))
			for j := range m.Durations {
				m.Durations[j] = []int{2, 3}[j%2] // the ~2.5-frame main-loop pass
			}
		}
		meta = append(meta, m)
	}
	packed, err := atlas.Pack(anims)
	if err != nil {
		return nil, err
	}
	f, err := b.CreateFile("objects", "helicopter.png")
	if err != nil {
		return nil, err
	}
	if err := packed.EncodePNG(f); err != nil {
		f.Close()
		return nil, err
	}
	f.Close()
	b.AddObject(schema.Asset{ID: "helicopter", Name: "Helicopter", Group: "Objects"}, &schema.Object{
		Type: schema.ObjectSprite2D,
		Name: "Helicopter",
		Atlas: &schema.SpriteAtlas{
			File: "helicopter.png", CellW: packed.CellW, CellH: packed.CellH,
			Anchor: []int{packed.Anchor.X, packed.Anchor.Y},
		},
		Animations: meta,
		Props: map[string]any{
			"art": "hardware sprite, white; the player (yellow) and the enemy (blue) are tints of this one craft",
		},
	})
	refs["helicopter"] = "helicopter"
	ctx.Progress("objects", 2, 2, fmt.Sprintf("helicopter: %d poses x 2 rotor frames", len(poses)))
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
			doc.Tilemap.Spawn.Object = refs["helicopter"]
			doc.Tilemap.Spawn.Anim = "level"

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
				ID: "prisoners", Count: 8, Name: "Prisoner",
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
				ID: "mines", Count: 13, Name: "Self-Propelled Mine",
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
			enemy := schema.Pool{
				ID: "enemy-helicopter", Count: 1, Name: "Enemy Helicopter",
				Object:     refs["helicopter"],
				Anim:       "left-full", // it banks into its pursuit; the classic on-screen pose
				Tint:       "#352879",
				Candidates: heliSpots,
				Seedable:   true,
			}
			// The enemy prose describes these INSTANCES (the object is the
			// shared craft), so it rides the pool's info card.
			if info, ok := ctx.Curation.Objects["enemy-helicopter"]; ok {
				enemy.Info = &schema.Info{Title: info.Title, Body: info.Text}
				delete(ctx.Curation.Objects, "enemy-helicopter")
			}
			doc.Pools = append(doc.Pools, enemy)
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
