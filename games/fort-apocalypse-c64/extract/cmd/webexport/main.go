// webexport serializes the decoded Fort Apocalypse levels for the static web
// viewer as the common format-2 asset tree (see STANDARDS.md / site/FORMAT2.md).
// Everything is reconstructed from the pre-extracted game file ($7000 image) by
// the same decode path as the inspection tools, then written under the output root:
//
//	manifest.json               game index: native res, tick rate, wrap, level list
//	levels/level<N>.json        per level (kind "tilemap2d"): char grid, soft-char
//	                            tile animations, the randomized object POOLS
//	levels/level<N>.objects.json machine-readable object DB (fixed objects + the
//	                            flattened pool candidates, each tagged with its pool)
//	levels/atlas-L<N>.png       the 128 playfield chars in the level's colours plus
//	                            the extra soft-char animation frames, 16 wide
//	sprites/index.json          helicopter sprite metadata + the sprite PNGs
//
// The playfield is a horizontal cylinder (manifest/level wrap "x"): column 214
// joins back to column 0. There is no music stage (C64, no in-engine track here).
// Both producers (levels, sprites) are folded into this one command; -only gates
// which stages run.
//
// Usage: webexport [-prg FORT-fast-7000.prg | -in FILE] [-o dir] [-only levels,sprites,all]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/fort-apocalypse-c64/extract/fortgfx"
	"retroreverse.com/tools/platform/c64/gfx"
)

type jsonAnim struct {
	Tiles        []int   `json:"tiles"`
	Frames       [][]int `json:"frames"` // atlas tile per step (one slot: the char itself)
	PeriodFrames int     `json:"periodFrames"`
}

type jsonStamp struct {
	DX   int `json:"dx"`
	DY   int `json:"dy"`
	Tile int `json:"tile"`
}

// dirStamps: direction-dependent art for patrolling placements — one stamps
// array per animation phase, per facing; the viewer swaps facing when the
// patrol reverses and advances the phase each update.
type jsonDirStamps struct {
	Right [][]jsonStamp `json:"right"`
	Left  [][]jsonStamp `json:"left"`
}

type jsonVariant struct {
	Stamps    []jsonStamp    `json:"stamps,omitempty"`
	DirStamps *jsonDirStamps `json:"dirStamps,omitempty"`
	Sprite    string         `json:"sprite,omitempty"`
	Tint      string         `json:"tint,omitempty"`
}

// patrol: back-and-forth movement, played by the viewer (site/FORMAT.md):
// every stepFrames engine frames one update fires (phase +1); every
// updatesPerStep-th update the placement moves stepPx in its current
// direction, reversing at the candidate's [minDx,maxDx] span. In slide mode
// the phase is the object's sub-cell position instead: it walks up/down with
// the direction, the stepPx move commits when it wraps, and a span edge
// turns the walk around without a jump (Fort's mines).
type jsonPatrol struct {
	StepPx         int     `json:"stepPx"`
	StepFrames     float64 `json:"stepFrames"`
	UpdatesPerStep int     `json:"updatesPerStep,omitempty"` // default 1
	Slide          bool    `json:"slide,omitempty"`
	StartPhase     int     `json:"startPhase,omitempty"` // rest/reset phase
	Start          string  `json:"start"`                // "random" | "right" | "left"
}

// objectPools: the game seeds prisoners/tanks/mines at level load — the viewer
// re-rolls `count` picks from the exported candidate list each time the objects
// layer is shown, choosing a random variant per placement. Candidates are
// [x, y] world px, or [x, y, minDx, maxDx] when the pool has a patrol block
// (the span precomputed from the engine's own turn-around rule).
type jsonPool struct {
	Count int `json:"count"`
	// Name identifies the pool's object kind for the viewer's info cards: the
	// shared viewer keys site/src/fort/config.js objectInfo[name] off it when the
	// player clicks a placement (site/FORMAT.md).
	Name       string        `json:"name,omitempty"`
	Candidates [][]int       `json:"candidates"`
	Patrol     *jsonPatrol   `json:"patrol,omitempty"`
	Variants   []jsonVariant `json:"variants"`
}

type jsonGrid struct {
	TileSize    int    `json:"tileSize"`
	Atlas       string `json:"atlas"`
	AtlasCols   int    `json:"atlasCols"`
	AtlasGutter int    `json:"atlasGutter"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Cells       []int  `json:"cells"`
}

type jsonRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type jsonSpawn struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Sprite string `json:"sprite"`
	Tint   string `json:"tint"`
}

// jsonExtents is the format-2 tilemap2d extent (size in cells).
type jsonExtents struct {
	TileSize int `json:"tileSize"`
	Width    int `json:"width"`
	Height   int `json:"height"`
}

// jsonLevel is the format-2 level envelope: the common header (format/name/kind/
// extents/wrap) followed by the unchanged format-1 tilemap body (grid, view,
// spawn, tileAnims, objectPools) and the sibling machine-readable object DB ref.
type jsonLevel struct {
	Format      int         `json:"format"`
	Name        string      `json:"name"`
	Kind        string      `json:"kind"` // "tilemap2d"
	Extents     jsonExtents `json:"extents"`
	Wrap        string      `json:"wrap"` // "x" — horizontal cylinder
	Grid        jsonGrid    `json:"grid"`
	View        jsonRect    `json:"view"`
	Spawn       jsonSpawn   `json:"spawn"`
	TileAnims   []jsonAnim  `json:"tileAnims"`
	ObjectPools []jsonPool  `json:"objectPools"`
	ObjectsFile string      `json:"objectsFile,omitempty"`
}

// Manifest is the format-2 per-game index (site/FORMAT2.md). Fort has no music.
type Manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Wrap     string         `json:"wrap,omitempty"`
	Levels   []LevelIndex   `json:"levels,omitempty"`
	Sprites  string         `json:"sprites,omitempty"`
}

type LevelIndex struct {
	Name    string `json:"name"`
	Section string `json:"section,omitempty"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Atlas   string `json:"atlas,omitempty"`
	Objects string `json:"objects,omitempty"`
}

// ObjectsDB is the machine-readable object database written alongside each level
// (levels/<stem>.objects.json). Fort has no fixed objects; every entry is a
// flattened object-pool candidate tagged (in props) with its pool.
type ObjectsDB struct {
	Format  int     `json:"format"`
	Level   string  `json:"level"`
	Objects []DBObj `json:"objects"`
}

type DBObj struct {
	ID    int            `json:"id"`
	Type  int            `json:"type,omitempty"`
	Name  string         `json:"name,omitempty"`
	Pos   []int          `json:"pos"`
	Props map[string]any `json:"props,omitempty"`
}

var levelNames = []string{"Vaults of Draconis", "Crystalline Caves"}

// parseOnly turns the -only selection into a stage set. "all" (or empty) selects
// every stage; otherwise only the named stages (levels,sprites) run.
func parseOnly(sel string) map[string]bool {
	out := map[string]bool{}
	for _, s := range strings.Split(sel, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if s == "all" {
			out["levels"], out["sprites"] = true, true
			continue
		}
		if s != "levels" && s != "sprites" {
			fmt.Fprintf(os.Stderr, "webexport: unknown -only stage %q (want levels,sprites,all)\n", s)
			os.Exit(2)
		}
		out[s] = true
	}
	return out
}

func main() {
	prg := flag.String("prg", "../extracted/FORT-fast-7000.prg", "pre-extracted game file ($7000)")
	in := flag.String("in", "", "alias for -prg (input game file)")
	outDir := flag.String("o", "../../site/public/fort-apocalypse-c64", "output asset root")
	only := flag.String("only", "all", "comma-separated subset of stages to run: levels,sprites,all")
	flag.Parse()
	path := *prg
	if *in != "" {
		path = *in
	}
	if err := run(path, *outDir, parseOnly(*only)); err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}

func run(prgPath, outDir string, sel map[string]bool) error {
	game, err := fortgfx.LoadGame(prgPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// The manifest covers whatever stages ran: a stage gated out via -only leaves
	// its section empty, and omitempty drops it, so a partial run stays consistent.
	man := Manifest{
		Format: 2, Game: "fort-apocalypse-c64", Platform: "Commodore 64",
		Native: map[string]int{"w": 320, "h": 200}, TickHz: 50, Wrap: "x",
	}
	if sel["levels"] {
		man.Levels, err = exportLevels(game, outDir)
		if err != nil {
			return err
		}
	}
	if sel["sprites"] {
		if err := exportSprites(game, outDir); err != nil {
			return err
		}
		man.Sprites = "sprites/index.json"
	}
	if err := writeJSON(filepath.Join(outDir, "manifest.json"), man); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[manifest] %d levels, sprites=%q -> %s\n",
		len(man.Levels), man.Sprites, outDir)
	return nil
}

// exportSprites writes the sprites/ tree (index.json + the helicopter PNGs) from
// the game image. The two poses are shared by both levels: the level-flight pose
// for the player and a banked pose for the enemy, white on transparent so the
// viewer can tint them (yellow / blue). No oracle.
func exportSprites(game *fortgfx.Game, outDir string) error {
	sprDir := filepath.Join(outDir, "sprites")
	if err := os.MkdirAll(sprDir, 0o755); err != nil {
		return err
	}
	poses := game.HelicopterPoses()
	shapes := game.SpriteShapes()
	// poses are in tilt order full-left .. level .. full-right; [0] is the sprite
	// block of each pose's first rotor frame.
	if err := gfx.WritePNG(filepath.Join(sprDir, "chopper-fwd.png"),
		chopperImg(shapes[poses[len(poses)/2][0]-1])); err != nil {
		return err
	}
	if err := gfx.WritePNG(filepath.Join(sprDir, "chopper-side.png"),
		chopperImg(shapes[poses[0][0]-1])); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(sprDir, "index.json"), map[string]any{
		"chopper-fwd":  map[string]any{"src": "sprites/chopper-fwd.png", "frames": [][4]int{{0, 0, 32, 18}}},
		"chopper-side": map[string]any{"src": "sprites/chopper-side.png", "frames": [][4]int{{0, 0, 32, 18}}},
	}); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[sprites] done: 2 sprites + index.json\n")
	return nil
}

// exportLevels writes the levels/ tree (per-level format-2 tilemaps, object DBs,
// atlases) and returns the manifest level index.
func exportLevels(game *fortgfx.Game, outDir string) ([]LevelIndex, error) {
	levelsDir := filepath.Join(outDir, "levels")
	if err := os.MkdirAll(levelsDir, 0o755); err != nil {
		return nil, err
	}
	cs := game.PlayfieldCharset()
	anim := game.SoftCharAnim()
	obst := game.ObstacleChars()

	var levels []LevelIndex
	fmt.Fprintf(os.Stderr, "[levels] %d levels\n", len(levelNames))
	for level := 0; level <= 1; level++ {
		lm, err := game.LevelMap(level)
		if err != nil {
			return nil, err
		}
		stem := fmt.Sprintf("level%d", level)
		atlasName := fmt.Sprintf("atlas-L%d.png", level)
		atlasRef := "levels/" + atlasName // path relative to the asset root

		jl := jsonLevel{Format: 2, Name: levelNames[level], Kind: "tilemap2d", Wrap: "x"}
		jl.Grid = jsonGrid{TileSize: 8, Atlas: atlasRef, AtlasCols: 16}

		// Atlas tiles: the 128 base chars at fixed indices 0..127, then any extra
		// frame bitmaps the animations need, appended and de-duplicated.
		tiles := make([][8]byte, 128)
		for ch := 0; ch < 128; ch++ {
			copy(tiles[ch][:], cs[ch*8:])
		}
		idxOf := map[[8]byte]int{}
		for i := 127; i >= 0; i-- {
			idxOf[tiles[i]] = i
		}
		addTile := func(b [8]byte) int {
			if i, ok := idxOf[b]; ok {
				return i
			}
			i := len(tiles)
			tiles = append(tiles, b)
			idxOf[b] = i
			return i
		}
		for _, a := range anim {
			ja := jsonAnim{Tiles: []int{int(a.Char)}, PeriodFrames: a.Period}
			for _, fr := range a.Frames {
				ja.Frames = append(ja.Frames, []int{addTile(fr)})
			}
			jl.TileAnims = append(jl.TileAnims, ja)
		}

		pal := palette(game.MulticolorValue(level))
		if err := writeAtlas(filepath.Join(levelsDir, atlasName), tiles, pal); err != nil {
			return nil, err
		}

		// Cells: the 215 content columns (Part IV §4). The playfield is a
		// cylinder — column 214 joins back to column 0 — so the stored wrap-seam
		// column (a duplicate of column 0) is dropped; the viewer wraps instead.
		w := fortgfx.ContentWidth
		jl.Extents = jsonExtents{TileSize: 8, Width: w, Height: fortgfx.MapHeight}
		jl.Grid.Width, jl.Grid.Height = w, fortgfx.MapHeight
		jl.Grid.Cells = make([]int, w*fortgfx.MapHeight)
		for r := 0; r < fortgfx.MapHeight; r++ {
			for c := 0; c < fortgfx.ContentWidth; c++ {
				jl.Grid.Cells[r*w+c] = int(lm.Cells[r][c])
			}
		}
		cellAt := func(c, r int) int { return jl.Grid.Cells[r*w+c] }

		// Player spawn (the yellow helicopter, tinted white sprite) + initial view:
		// one C64 screen (320x200) with the spawn centred horizontally, top row up.
		sx, sy := lm.PlayerSpawn.Col*8, lm.PlayerSpawn.Row*8
		jl.Spawn = jsonSpawn{X: sx, Y: sy, Sprite: "chopper-fwd", Tint: "#b8c76f"}
		jl.View = jsonRect{X: sx + 16 - 160, Y: 0, W: 320, H: 200}

		// Object pools — the game's load-time seeding rules, as candidates the
		// viewer re-rolls (Part V; previously hardcoded in the viewer). Movers
		// carry a patrol block with per-candidate spans from the engines' own
		// turn-around rules. CADENCE: the engines count main-loop PASSES, and
		// in gameplay the loop free-runs (no frame wait, $8BB1) at a measured
		// ~2.3 frames per pass under cmd/paceprobe — exported as 2.5 frames
		// per pass to cover the VIC's badline/sprite DMA the oracle doesn't
		// model. At base difficulty:
		//  - prisoners: 8 of the $90A4 floor-with-rock candidates. Run back and
		//    forth along the $48 walkway ($AB9A), one of 8 slots serviced per
		//    pass ($EA cursor) = 8px every 8 passes = 20 frames, legs
		//    alternating per step; torso $49 + legs $3B/$3C facing right,
		//    torso $4A + legs $3E/$3D facing left (draw tables $AC12).
		//  - tanks: every fixed home; body $6C $6D $6E, turret $6F/$70 random
		//    (the game aims it at the player). One global countdown ($F0,
		//    reload 4) steps all six in lockstep, 8px every 4 passes =
		//    10 frames, reversing on the $A45D obstacle chars ($9A8F).
		//  - SPMs: spmCount mines at empty 2-cell spots in the column band
		//    $2D..$C8. 15 of 39 slots serviced per pass = one update per
		//    2.6 passes = 6.5 frames; 4 draw phases per cell of travel ($963C
		//    pairs), so 8px every 4 updates, reversing unless both destination
		//    cells stay empty inside the band ($9617).
		//  - one enemy helicopter at a random 4x2-clear spot in the same band
		//    (a player-pursuit AI, not a patroller — left static).
		var prisoners [][]int
		for _, p := range lm.PrisonerSpawns {
			l, r := lm.PrisonerRange(p.Col, p.Row)
			prisoners = append(prisoners, []int{p.Col * 8, p.Row * 8, l * 8, r * 8})
		}
		jl.ObjectPools = append(jl.ObjectPools, jsonPool{
			Count: 8, Name: "prisoner", Candidates: prisoners,
			Patrol: &jsonPatrol{StepPx: 8, StepFrames: 20, Start: "random"},
			Variants: []jsonVariant{
				{DirStamps: &jsonDirStamps{
					Right: [][]jsonStamp{
						{{0, -8, 0x49}, {0, 0, 0x3B}},
						{{0, -8, 0x49}, {0, 0, 0x3C}},
					},
					Left: [][]jsonStamp{
						{{0, -8, 0x4A}, {0, 0, 0x3E}},
						{{0, -8, 0x4A}, {0, 0, 0x3D}},
					},
				}},
			},
		})
		var tanks [][]int
		for _, p := range lm.TankHomes {
			l, r := lm.TankRange(obst, p.Col, p.Row)
			tanks = append(tanks, []int{p.Col * 8, p.Row * 8, l * 8, r * 8})
		}
		jl.ObjectPools = append(jl.ObjectPools, jsonPool{
			Count: len(tanks), Name: "tank", Candidates: tanks,
			Patrol: &jsonPatrol{StepPx: 8, StepFrames: 10, Start: "random"},
			Variants: []jsonVariant{
				{Stamps: []jsonStamp{{0, 0, 0x6C}, {8, 0, 0x6D}, {16, 0, 0x6E}, {8, -8, 0x6F}}},
				{Stamps: []jsonStamp{{0, 0, 0x6C}, {8, 0, 0x6D}, {16, 0, 0x6E}, {8, -8, 0x70}}},
			},
		})
		// SPM slide phases: the $963C char pairs are the SAME 26-pixel mine at
		// four sub-cell positions — blob centres 4.5 / 6.5 / 8.5 / 10.5 px
		// within the 2-cell pair, i.e. a 2px step per phase. The phase counter
		// IS the mine's sub-cell position (it walks down again moving left, and
		// the column commits on wrap), so the viewer plays these in slide mode:
		// phase-coupled position, identical art both directions, resting on
		// phase 1 (the classic straddling mine).
		spmPhases := [][]jsonStamp{
			{{0, 0, 0x40}},
			{{0, 0, 0x5B}, {8, 0, 0x5C}},
			{{0, 0, 0x5D}, {8, 0, 0x5E}},
			{{8, 0, 0x5F}},
		}
		const bandMin, bandMax = fortgfx.SPMBandMin, fortgfx.SPMBandMax
		var spmSpots [][]int
		for r := 0; r < fortgfx.MapHeight; r++ {
			for c := bandMin; c <= bandMax && c+1 < w; c++ {
				if cellAt(c, r) == 0 && cellAt(c+1, r) == 0 {
					l, rr := lm.SPMRange(c, r)
					spmSpots = append(spmSpots, []int{c * 8, r * 8, l * 8, rr * 8})
				}
			}
		}
		jl.ObjectPools = append(jl.ObjectPools, jsonPool{
			Count: 13, Name: "mine", Candidates: spmSpots, // base difficulty (13 / 26 / 39 by variant)
			Patrol: &jsonPatrol{StepPx: 8, StepFrames: 6.5, Slide: true, StartPhase: 1, Start: "right"},
			Variants: []jsonVariant{
				{DirStamps: &jsonDirStamps{Right: spmPhases, Left: spmPhases}},
			},
		})
		var heliSpots [][]int
		for r := 0; r+1 < fortgfx.MapHeight; r++ {
			for c := bandMin; c <= bandMax && c+3 < w; c++ {
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
					heliSpots = append(heliSpots, []int{c * 8, r * 8})
				}
			}
		}
		jl.ObjectPools = append(jl.ObjectPools, jsonPool{
			Count: 1, Name: "enemy-helicopter", Candidates: heliSpots,
			Variants: []jsonVariant{
				{Sprite: "chopper-side", Tint: "#352879"},
			},
		})

		jl.ObjectsFile = stem + ".objects.json"
		if err := writeJSON(filepath.Join(levelsDir, stem+".json"), jl); err != nil {
			return nil, err
		}

		// Machine-readable object DB (sibling of the level file): the fixed objects
		// (Fort has none) plus every pool candidate flattened into one placement,
		// tagged with its pool name and, for movers, its precomputed patrol span.
		db := ObjectsDB{Format: 2, Level: levelNames[level]}
		id := 0
		for _, pool := range jl.ObjectPools {
			for _, cand := range pool.Candidates {
				o := DBObj{ID: id, Name: pool.Name, Pos: []int{cand[0], cand[1]},
					Props: map[string]any{"pool": pool.Name}}
				if len(cand) >= 4 {
					o.Props["patrol"] = []int{cand[2], cand[3]} // world-px span [minX, maxX]
				}
				db.Objects = append(db.Objects, o)
				id++
			}
		}
		if err := writeJSON(filepath.Join(levelsDir, stem+".objects.json"), db); err != nil {
			return nil, err
		}

		levels = append(levels, LevelIndex{
			Name: levelNames[level], File: "levels/" + stem + ".json",
			Kind: "tilemap2d", Atlas: atlasRef, Objects: "levels/" + stem + ".objects.json",
		})
		fmt.Fprintf(os.Stderr, "[levels] %d/%d  %-18s %3dx%-2d cells  %d atlas tiles  %s  %d objects\n",
			level+1, len(levelNames), levelNames[level], w, fortgfx.MapHeight, len(tiles), atlasName, len(db.Objects))
	}
	fmt.Fprintf(os.Stderr, "[levels] done: %d levels\n", len(levels))
	return levels, nil
}

// palette is the playfield's multicolor map: 00=black, 01=$D022 (per level),
// 10=white, 11=colour-RAM green (Part IV §2).
func palette(d022 byte) [4]color.RGBA {
	return [4]color.RGBA{gfx.Palette[0], gfx.Palette[d022&0x0F], gfx.Palette[1], gfx.Palette[5]}
}

// chopperImg renders one helicopter sprite block (the [left][right][$00] rows
// the game expands, 16x18 used pixels) into a 32x18 RGBA image — white where set,
// transparent elsewhere, X-expanded (the sprites are drawn double-wide in game).
func chopperImg(block []byte) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 32, 18))
	white := color.RGBA{255, 255, 255, 255}
	for r := 0; r < 18; r++ {
		for bi := 0; bi < 2; bi++ {
			b := block[r*3+bi]
			for bit := 0; bit < 8; bit++ {
				if b&(0x80>>bit) != 0 {
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
func writeAtlas(path string, tiles [][8]byte, pal [4]color.RGBA) error {
	const cols = 16
	rows := (len(tiles) + cols - 1) / cols
	img := image.NewRGBA(image.Rect(0, 0, cols*8, rows*8))
	for i, t := range tiles {
		gfx.DrawChar(img, t[:], (i%cols)*8, (i/cols)*8, 1, pal)
	}
	return gfx.WritePNG(path, img)
}

func writeJSON(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
