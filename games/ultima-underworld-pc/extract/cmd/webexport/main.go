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
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/ultima-underworld-pc/extract/lev"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
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
	cli.Main("ultima-underworld-pc", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in <game dir> [-o DIR]")
	}
	b := ctx.Builder
	b.SetTitle("Ultima Underworld: The Stygian Abyss")
	b.SetPlatform("MS-DOS")
	b.SetYear(1992)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 320, H: 200},
		TickHz: 60,
		Filter: "crt",
	})
	ctx.Stage("objects")
	ctx.Stage("levels")
	return exportAll(ctx, ctx.In, 0, true)
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
// exportAll bakes each dungeon level's textured mesh to a GLB layer, dedups
// every placed sprite's art into billboard3d objects, and writes the level
// documents with camera-facing placements at the objects' engine positions.
func exportAll(ctx *cli.Context, gameDir string, palN int, ceil bool) error {
	b := ctx.Builder
	nLevels := countLevels(gameDir)

	scratch, err := os.MkdirTemp("", "uw-sprites-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)

	type levelData struct {
		o  *outMesh
		db ObjectsDB
	}
	var lvls []levelData
	nextID := 0
	for n := 0; n < nLevels; n++ {
		o := buildLevel(gameDir, n, palN, ceil)
		db := ObjectsDB{Format: 2, Level: fmt.Sprintf("Level %d", n+1)}
		buildObjects(&db, o, scratch, &nextID)
		lvls = append(lvls, levelData{o, db})
		ctx.Progress("objects", n+1, nLevels, fmt.Sprintf("level %d: %d objects decoded", n+1, len(db.Objects)))
	}

	// ---- dedup sprite art into billboard3d assets --------------------------
	type variant struct {
		sp    *FSprite
		name  string
		asset string
	}
	byKey := map[string]*variant{}
	usedIDs := map[string]bool{}
	refs := map[string]string{} // sheet path -> asset id
	for _, ld := range lvls {
		for _, fo := range ld.db.Objects {
			if fo.Sprite == nil {
				continue
			}
			sp := fo.Sprite
			stem := strings.TrimSuffix(filepath.Base(sp.Sheet), ".png")
			png0, err := os.ReadFile(filepath.Join(scratch, stem+".png"))
			if err != nil {
				continue
			}
			key := fmt.Sprintf("%x|%v|%d|%d|%v|%s|%s|%v|%s",
				md5sum(png0), sp.Frames, sp.Views, sp.PerView, sp.Size, sp.Anchor, sp.Blend, sp.Fps, fo.Name)
			if v, ok := byKey[key]; ok {
				refs[sp.Sheet] = v.asset
				continue
			}
			name := fo.Name
			if name == "" {
				name = "Sprite " + stem
			}
			base := slug(name)
			id := base
			for n2 := 2; usedIDs[id]; n2++ {
				id = fmt.Sprintf("%s-%d", base, n2)
			}
			usedIDs[id] = true

			// Repack the (possibly ragged) frames onto a uniform cell grid:
			// rows = views, cols = perView; frames centre-x, bottom-aligned so
			// the feet stay on the anchor.
			src, err := loadPNGFile(filepath.Join(scratch, stem+".png"))
			if err != nil {
				continue
			}
			cw, ch := 1, 1
			for _, f := range sp.Frames {
				if f[2] > cw {
					cw = f[2]
				}
				if f[3] > ch {
					ch = f[3]
				}
			}
			rows := sp.Views
			cols := sp.PerView
			if rows < 1 {
				rows = 1
			}
			if cols < 1 {
				cols = 1
			}
			atlasImg := image.NewNRGBA(image.Rect(0, 0, cw*cols, ch*rows))
			for fi, f := range sp.Frames {
				r, c := fi/cols, fi%cols
				if r >= rows {
					break
				}
				ox := c*cw + (cw-f[2])/2
				oy := r*ch + (ch - f[3])
				draw.Draw(atlasImg, image.Rect(ox, oy, ox+f[2], oy+f[3]), src, image.Pt(f[0], f[1]), draw.Src)
			}
			f, err := b.CreateFile("objects", id+".png")
			if err != nil {
				return err
			}
			err = png.Encode(f, atlasImg)
			f.Close()
			if err != nil {
				return err
			}

			group := "Items"
			if sp.Views > 1 {
				group = "Creatures"
			} else if sp.Blend == "additive" {
				group = "Glow overlays"
			}
			doc := &schema.Object{
				Type: schema.ObjectBillboard3D, Name: name,
				Atlas:      &schema.SpriteAtlas{File: id + ".png", CellW: cw, CellH: ch},
				Views:      rows,
				Mode:       "camera",
				Size:       []float64{float64(sp.Size[0]), float64(sp.Size[1])},
				AnchorMode: sp.Anchor,
				Blend:      sp.Blend,
				Animations: []schema.Animation{{ID: "main", Loop: "loop", Col: 0, FramesPerView: cols, FPS: float64(sp.Fps)}},
			}
			if sp.Billboard == "yaw" {
				doc.Mode = "yaw"
			}
			if sp.Heading != 0 {
				doc.Heading = sp.Heading
			}
			b.AddObject(schema.Asset{ID: id, Name: name, Group: group}, doc)
			byKey[key] = &variant{sp: sp, name: name, asset: id}
			refs[sp.Sheet] = id
		}
	}
	ctx.Logf("%d distinct sprite assets from %d placed objects", len(byKey), nextID)

	// ---- level documents ----------------------------------------------------
	for n, ld := range lvls {
		o := ld.o
		// bake the inline textured mesh to a GLB layer
		positions := make([][3]float32, len(o.Positions)/3)
		for i := range positions {
			positions[i] = [3]float32{o.Positions[i*3], o.Positions[i*3+1], o.Positions[i*3+2]}
		}
		uvs := make([][2]float32, len(o.UVs)/2)
		for i := range uvs {
			uvs[i] = [2]float32{o.UVs[i*2], o.UVs[i*2+1]}
		}
		var groups []glb.TexturedGroup
		for _, g := range o.Groups {
			if g.Material < 0 || g.Material >= len(o.Textures) {
				continue
			}
			// Group Start/Count are in vertex units (the three.js addGroup
			// convention level.go and objects.go emit).
			var tris [][3]uint32
			for v := g.Start; v+2 < g.Start+g.Count; v += 3 {
				tris = append(tris, [3]uint32{uint32(v), uint32(v + 1), uint32(v + 2)})
			}
			groups = append(groups, glb.TexturedGroup{
				Tris:        tris,
				Image:       decodeSpriteTex(o.Textures[g.Material].PNG),
				SingleSided: o.Textures[g.Material].Ceiling,
				WrapS:       10497, WrapT: 10497,
			})
		}
		glbFile := fmt.Sprintf("level%d.glb", n+1)
		gp, err := b.Path("levels", glbFile)
		if err != nil {
			return err
		}
		if err := glb.WriteTextured(gp, positions, uvs, groups, nil); err != nil {
			return err
		}

		spawn, dir := vec3(o.Spawn), vec3(o.SpawnDir)
		doc := &schema.Level{
			Type: schema.LevelScene3D,
			Scene: &schema.Scene{Layers: []schema.Layer{{
				ID: "dungeon", File: glbFile,
			}}},
			Camera: &schema.Camera{
				Mode: "fly", FOV: 60, Near: 0.02, Far: 200, Fly: &schema.Fly{Speed: 3},
				Pos: []float64{float64(spawn[0]), float64(spawn[1]), float64(spawn[2])},
				Target: []float64{float64(spawn[0]) + float64(dir[0])*4,
					float64(spawn[1]) + float64(dir[1])*4, float64(spawn[2]) + float64(dir[2])*4},
			},
		}
		for _, fo := range ld.db.Objects {
			if fo.Sprite == nil {
				continue
			}
			asset, ok := refs[fo.Sprite.Sheet]
			if !ok {
				continue
			}
			pl := schema.Placement{
				ID:     fo.ID,
				Object: asset,
				Pos:    []float64{float64(fo.Pos[0]), float64(fo.Pos[1]), float64(fo.Pos[2])},
			}
			if fo.Name != "" {
				pl.Name = fo.Name
			}
			if fo.Props != nil && fo.Props.Text != "" {
				pl.OnClick = &schema.OnClick{Action: schema.ActionText, Title: fo.Name, Body: fo.Props.Text}
			}
			doc.Placements = append(doc.Placements, pl)
		}
		b.AddLevel(schema.Asset{
			ID: fmt.Sprintf("level-%d", n+1), Name: fmt.Sprintf("Level %d", n+1), Group: "The Abyss",
		}, doc)
		ctx.Progress("levels", n+1, len(lvls), fmt.Sprintf("level %d: %d placements", n+1, len(doc.Placements)))
	}
	return nil
}

func md5sum(b []byte) [16]byte { return md5.Sum(b) }

func slug(s string) string {
	var out []rune
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
		default:
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	return strings.Trim(string(out), "-")
}

func loadPNGFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
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
