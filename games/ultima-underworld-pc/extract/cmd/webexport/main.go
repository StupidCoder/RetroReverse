// webexport reconstructs Ultima Underworld's web assets from the installed game
// directory and writes the common format-2 asset tree (see FORMAT2.md /
// STANDARDS.md §4) under the output root:
//
//	manifest.json                      game index: native res, tick rate, level list
//	levels/level<N>.json               per level: the format-2 mesh3d envelope with an
//	                                   INLINE textured mesh (positions/uvs/groups/textures);
//	                                   the ceiling group is flagged singleSided so the
//	                                   viewer back-face-culls it (see-into-rooms-from-above)
//	levels/level<N>.objects.json       per level: the machine-readable object DB — first-class
//	                                   directional-billboard SPRITE objects (items, creatures,
//	                                   additive glow overlays) and click AABBs with their words
//	sprites/<id>.png                   one atlas per sprite object (billboard = single frame,
//	                                   creature = views×frames grid)
//
// The level-decode pipeline is shared with cmd/levexport (level.go + objects.go);
// this command reshapes the intermediate outMesh into the format-2 tree. -only
// gates which stages run (only "levels" today).
//
// Usage: webexport [-game/-in <game dir>] [-o <outdir>] [-only levels,all]
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/ultima-underworld-pc/extract/lev"
)

// Manifest is the format-2 game index (STANDARDS.md §4.2). Sections gated out by
// -only stay empty and omitempty drops them.
type Manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Levels   []LevelIndex   `json:"levels,omitempty"`
}

// LevelIndex is one manifest levels[] entry.
type LevelIndex struct {
	Name    string `json:"name"`
	Section string `json:"section,omitempty"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Objects string `json:"objects"`
}

// LevelFile is the format-2 mesh3d level envelope (FORMAT2.md).
type LevelFile struct {
	Format      int       `json:"format"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Mesh        MeshBody  `json:"mesh"`
	Spawn       SpawnBody `json:"spawn"`
	ObjectsFile string    `json:"objectsFile"`
}

// MeshBody is the inline mesh3d geometry (FORMAT2 mesh option B). Groups carry a
// per-material singleSided flag — the first-class inline-mesh analogue of a GLB
// material's doubleSided:false — set on the dungeon ceiling so the viewer back-
// face-culls it (the camera sees into rooms from above, ceiling still occludes
// from below).
type MeshBody struct {
	Positions []float32   `json:"positions"`
	UVs       []float32   `json:"uvs"`
	Groups    []MeshGroup `json:"groups"`
	Textures  []string    `json:"textures"` // data:image/png;base64,... indexed by group.material
}

type MeshGroup struct {
	Material    int  `json:"material"`
	Start       int  `json:"start"`
	Count       int  `json:"count"`
	SingleSided bool `json:"singleSided,omitempty"`
}

type SpawnBody struct {
	Pos [3]float32 `json:"pos"`
	Dir [3]float32 `json:"dir"`
}

// ObjectsDB is the format-2 machine-readable object DB (STANDARDS.md §4.4).
type ObjectsDB struct {
	Format  int       `json:"format"`
	Level   string    `json:"level"`
	Objects []FObject `json:"objects"`
}

// FObject is one object: a directional-billboard sprite (item/creature/glow
// overlay) or a bare click AABB carrying the object's words.
type FObject struct {
	ID     int        `json:"id"`
	Pos    [3]float32 `json:"pos"`
	Size   []float32  `json:"size,omitempty"` // [w,h,d] click extents (pick objects)
	Name   string     `json:"name,omitempty"`
	Sprite *FSprite   `json:"sprite,omitempty"`
	Props  *FProps    `json:"props,omitempty"`
}

// FSprite is the first-class directional-billboard sprite spec (FORMAT2.md
// §"3-D object rendering"). frames are atlas pixel rects [x,y,w,h], laid out as
// `views` blocks of `perView` frames.
type FSprite struct {
	Billboard string     `json:"billboard"`
	Views     int        `json:"views"`
	Heading   float64    `json:"heading,omitempty"`
	Size      [2]float32 `json:"size"`
	Anchor    string     `json:"anchor,omitempty"` // "bottom": UW places the feet at pos, not the centre
	Sheet     string     `json:"sheet"`
	Frames    [][4]int   `json:"frames"`
	PerView   int        `json:"perView"`
	Fps       float32    `json:"fps"`
	Blend     string     `json:"blend"`
}

type FProps struct {
	Text string `json:"text,omitempty"`
}

func main() {
	game := flag.String("game", "../game", "path to the installed game/ folder")
	in := flag.String("in", "", "alias for -game (input game dir); overrides -game when set")
	outdir := flag.String("o", "../../site/public/ultima-underworld-pc", "output asset root")
	only := flag.String("only", "all", "comma-separated subset of stages: levels,all")
	palN := flag.Int("pal", 0, "PALS.DAT palette index")
	ceil := flag.Bool("ceilings", true, "include ceiling faces (enclosed dungeon)")
	flag.Parse()

	gameDir := *game
	if *in != "" {
		gameDir = *in
	}
	sel := parseOnly(*only)
	chk(os.MkdirAll(*outdir, 0o755))

	man := Manifest{
		Format: 2, Game: "ultima-underworld-pc", Platform: "MS-DOS",
		Native: map[string]int{"w": 320, "h": 200}, TickHz: 60,
	}
	if sel["levels"] {
		man.Levels = exportLevels(gameDir, *outdir, *palN, *ceil)
	}
	writeJSON(filepath.Join(*outdir, "manifest.json"), man)
	fmt.Fprintf(os.Stderr, "[manifest] %d levels -> %s\n", len(man.Levels), *outdir)
}

// countLevels reports how many LEV.ARK blocks are full dungeon-level blocks
// (mirrors cmd/levinfo): blocks 0..N-1 of exactly LevelBlockSize bytes.
func countLevels(gameDir string) int {
	b, err := os.ReadFile(filepath.Join(gameDir, "DATA", "LEV.ARK"))
	chk(err)
	ark, err := lev.ParseArk(b)
	chk(err)
	n := 0
	for i := range ark.Offsets {
		if blk, err := ark.Block(i); err == nil && len(blk) == lev.LevelBlockSize {
			n++
		}
	}
	return n
}

// exportLevels emits every dungeon level's format-2 mesh3d tree and returns the
// manifest levels[] index. Object ids run globally across levels so each sprite's
// sprites/<id>.png atlas file name is unique.
func exportLevels(gameDir, outdir string, palN int, ceil bool) []LevelIndex {
	nLevels := countLevels(gameDir)
	levelsDir := filepath.Join(outdir, "levels")
	spritesDir := filepath.Join(outdir, "sprites")
	chk(os.MkdirAll(levelsDir, 0o755))
	chk(os.MkdirAll(spritesDir, 0o755))

	nextID := 0
	var idx []LevelIndex
	for n := 0; n < nLevels; n++ {
		name := fmt.Sprintf("Level %d", n+1)
		o := buildLevel(gameDir, n, palN, ceil)

		// mesh3d envelope with inline geometry; carry each group's ceiling flag
		// through as singleSided (the inline-mesh single-sided-material path).
		mesh := MeshBody{Positions: o.Positions, UVs: o.UVs}
		for _, t := range o.Textures {
			mesh.Textures = append(mesh.Textures, t.PNG)
		}
		for _, g := range o.Groups {
			single := g.Material >= 0 && g.Material < len(o.Textures) && o.Textures[g.Material].Ceiling
			mesh.Groups = append(mesh.Groups, MeshGroup{
				Material: g.Material, Start: g.Start, Count: g.Count, SingleSided: single,
			})
		}
		lf := LevelFile{
			Format: 2, Name: name, Kind: "mesh3d", Mesh: mesh,
			Spawn:       SpawnBody{Pos: vec3(o.Spawn), Dir: vec3(o.SpawnDir)},
			ObjectsFile: fmt.Sprintf("level%d.objects.json", n+1),
		}
		levelFile := filepath.Join("levels", fmt.Sprintf("level%d.json", n+1))
		writeJSON(filepath.Join(outdir, levelFile), lf)

		// Object DB: first-class sprite/pick objects.
		db := ObjectsDB{Format: 2, Level: name}
		nSprite, nCreature, nOverlay, nPick := buildObjects(&db, o, spritesDir, &nextID)
		objFile := filepath.Join("levels", lf.ObjectsFile)
		writeJSON(filepath.Join(outdir, objFile), db)

		idx = append(idx, LevelIndex{
			Name: name, File: filepath.ToSlash(levelFile), Kind: "mesh3d",
			Objects: filepath.ToSlash(objFile),
		})
		fmt.Fprintf(os.Stderr, "[levels] %2d/%2d  %-8s %6d tris, %2d groups, %d sprites, %d creatures, %d glow, %d picks\n",
			n+1, nLevels, name, len(o.Positions)/9, len(mesh.Groups),
			nSprite, nCreature, nOverlay, nPick)
	}
	fmt.Fprintf(os.Stderr, "[levels] done: %d levels\n", len(idx))
	return idx
}

// buildObjects appends the level's sprite and pick objects to db, writing each
// sprite's atlas PNG under spritesDir. It returns per-kind counts. *nextID is the
// running global object id (also the atlas file stem), advanced as objects emit.
func buildObjects(db *ObjectsDB, o *outMesh, spritesDir string, nextID *int) (nSprite, nCreature, nOverlay, nPick int) {
	// Which SpriteTex indices draw additively (index-252 glow bodies).
	additive := map[int]bool{}
	for _, ti := range o.AddTex {
		additive[ti] = true
	}
	decode := func(i int) *image.RGBA { return decodeSpriteTex(o.SpriteTex[i]) }

	// Camera-facing billboards (items): one-frame sprite atlas.
	for _, sp := range o.Sprites {
		id := *nextID
		*nextID++
		img := decode(sp.Tex)
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		writeSpritePNG(filepath.Join(spritesDir, fmt.Sprintf("%d.png", id)), o.SpriteTex[sp.Tex])
		blend := "alpha"
		if additive[sp.Tex] {
			blend = "additive"
		}
		db.Objects = append(db.Objects, FObject{
			ID: id, Pos: sp.Pos,
			Sprite: &FSprite{
				Billboard: "camera", Views: 1,
				Size:   [2]float32{sp.W, sp.H},
				Anchor: "bottom",
				Sheet:  fmt.Sprintf("sprites/%d.png", id),
				Frames: [][4]int{{0, 0, w, h}}, PerView: 1, Fps: 0, Blend: blend,
			},
		})
		nSprite++
	}

	// Directional creatures (yaw billboards, up to 8 views): views×frames atlas.
	for _, cr := range o.Creatures {
		views := len(cr.Views)
		if views == 0 {
			continue
		}
		perView := 0
		for _, cyc := range cr.Views {
			if len(cyc) > perView {
				perView = len(cyc)
			}
		}
		if perView == 0 {
			continue
		}
		// A single world quad size per the sprite spec; the first view's first
		// frame sizes it (matching the level's own click box).
		hb := cr.Views[0][0]
		size := [2]float32{hb.W, hb.H}

		// Gather base (Tex) frame images and any additive (Add) overlay images,
		// laid out row-major as `views` blocks of `perView` (short cycles repeat
		// their last frame to keep the grid rectangular).
		baseImgs := make([]*image.RGBA, views*perView)
		overImgs := make([]*image.RGBA, views*perView)
		hasOverlay := false
		for v := 0; v < views; v++ {
			cyc := cr.Views[v]
			for f := 0; f < perView; f++ {
				src := cyc[min(f, len(cyc)-1)]
				baseImgs[v*perView+f] = decode(src.Tex)
				if src.Add >= 0 {
					overImgs[v*perView+f] = decode(src.Add)
					hasOverlay = true
				}
			}
		}

		heading := float64(cr.Heading) * (math.Pi / 4)
		blend := "alpha"
		if cr.Translucent {
			blend = "additive"
		}

		// Base creature sprite.
		id := *nextID
		*nextID++
		frames := packAtlas(filepath.Join(spritesDir, fmt.Sprintf("%d.png", id)), baseImgs, perView)
		db.Objects = append(db.Objects, FObject{
			ID: id, Pos: cr.Pos,
			Sprite: &FSprite{
				Billboard: "yaw", Views: views, Heading: heading, Size: size, Anchor: "bottom",
				Sheet:  fmt.Sprintf("sprites/%d.png", id),
				Frames: frames, PerView: perView, Fps: cr.Fps, Blend: blend,
			},
		})
		nCreature++

		// Per-frame additive glow overlay → a SECOND additive sprite at the same
		// position with its own atlas (frames without a glow layer stay empty).
		if hasOverlay {
			oid := *nextID
			*nextID++
			oframes := packAtlas(filepath.Join(spritesDir, fmt.Sprintf("%d.png", oid)), overImgs, perView)
			db.Objects = append(db.Objects, FObject{
				ID: oid, Pos: cr.Pos,
				Sprite: &FSprite{
					Billboard: "yaw", Views: views, Heading: heading, Size: size, Anchor: "bottom",
					Sheet:  fmt.Sprintf("sprites/%d.png", oid),
					Frames: oframes, PerView: perView, Fps: cr.Fps, Blend: "additive",
				},
			})
			nOverlay++
		}
	}

	// Click AABBs: pos + full extents, carrying the object's name and words.
	for _, pk := range o.Picks {
		id := *nextID
		*nextID++
		obj := FObject{
			ID: id, Pos: pk.Pos,
			Size: []float32{pk.Size[0], pk.Size[1], pk.Size[2]},
			Name: o.Names[pk.ID],
		}
		if pk.Text != "" {
			obj.Props = &FProps{Text: pk.Text}
		}
		db.Objects = append(db.Objects, obj)
		nPick++
	}
	return nSprite, nCreature, nOverlay, nPick
}

// packAtlas composes imgs (row-major, `cols` per row) into one atlas PNG at path
// and returns each cell's pixel rect [x,y,w,h]. Cells are max-frame sized; a nil
// image (a missing glow layer) leaves its cell transparent with a 1×1 rect.
func packAtlas(path string, imgs []*image.RGBA, cols int) [][4]int {
	rows := (len(imgs) + cols - 1) / cols
	cellW, cellH := 1, 1 // max frame dims across the grid
	for _, im := range imgs {
		if im == nil {
			continue
		}
		b := im.Bounds()
		cellW = max(cellW, b.Dx())
		cellH = max(cellH, b.Dy())
	}
	atlas := image.NewRGBA(image.Rect(0, 0, cellW*cols, cellH*rows))
	rects := make([][4]int, len(imgs))
	for i, im := range imgs {
		cx := (i % cols) * cellW
		cy := (i / cols) * cellH
		if im == nil {
			rects[i] = [4]int{cx, cy, 1, 1} // transparent placeholder
			continue
		}
		w, h := im.Bounds().Dx(), im.Bounds().Dy()
		draw.Draw(atlas, image.Rect(cx, cy, cx+w, cy+h), im, im.Bounds().Min, draw.Src)
		rects[i] = [4]int{cx, cy, w, h}
	}
	writePNG(path, atlas)
	return rects
}

// decodeSpriteTex decodes a SpriteTex data URI into an RGBA image.
func decodeSpriteTex(uri string) *image.RGBA {
	i := strings.IndexByte(uri, ',')
	if i < 0 {
		fmt.Fprintln(os.Stderr, "webexport: sprite tex is not a data URI")
		os.Exit(1)
	}
	raw, err := base64.StdEncoding.DecodeString(uri[i+1:])
	chk(err)
	im, err := png.Decode(bytes.NewReader(raw))
	chk(err)
	if r, ok := im.(*image.RGBA); ok {
		return r
	}
	r := image.NewRGBA(im.Bounds())
	draw.Draw(r, r.Bounds(), im, im.Bounds().Min, draw.Src)
	return r
}

// writeSpritePNG writes a SpriteTex data URI straight to a PNG file (byte-exact
// decode, no re-encode).
func writeSpritePNG(path, uri string) {
	i := strings.IndexByte(uri, ',')
	if i < 0 {
		fmt.Fprintln(os.Stderr, "webexport: sprite tex is not a data URI")
		os.Exit(1)
	}
	raw, err := base64.StdEncoding.DecodeString(uri[i+1:])
	chk(err)
	chk(os.WriteFile(path, raw, 0o644))
}

func writePNG(path string, im image.Image) {
	var buf bytes.Buffer
	chk(png.Encode(&buf, im))
	chk(os.WriteFile(path, buf.Bytes(), 0o644))
}

func vec3(s []float32) [3]float32 {
	var v [3]float32
	copy(v[:], s)
	return v
}

// parseOnly turns the -only selection into a stage set. "all" (or empty) selects
// every stage; otherwise only the named stages run.
func parseOnly(sel string) map[string]bool {
	out := map[string]bool{}
	for _, s := range strings.Split(sel, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if s == "all" {
			out["levels"] = true
			continue
		}
		if s != "levels" {
			fmt.Fprintf(os.Stderr, "webexport: unknown -only stage %q (want levels,all)\n", s)
			os.Exit(2)
		}
		out[s] = true
	}
	return out
}

func chk(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}

func writeJSON(path string, v any) {
	b, err := json.Marshal(v)
	chk(err)
	chk(os.WriteFile(path, b, 0o644))
}
