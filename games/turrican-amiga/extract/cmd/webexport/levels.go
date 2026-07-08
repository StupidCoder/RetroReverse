// levels.go is the levels stage: for each world it writes a tile atlas and object atlas into
// levels/, and for each scene a format-2 tilemap2d level file (the format-1 grid/collision/
// objects body inside the common envelope) plus a machine-readable object DB. Placement comes
// straight off the disk via the scene package; each object's displayed sprite/frame/position
// is resolved by running its AI handler's spawn-init in the 68000 interpreter (scene.Sim).
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/turrican-amiga/extract/scene"
)

const (
	blockBase  = 0x1B980
	tileSide   = 32
	tileBytes  = tileSide * 4 * (tileSide / 8) // 512
	atlasCols  = 16
	tileGutter = 1                       // extruded border per tile (kills atlas bleed at fractional zoom)
	atlasCell  = tileSide + 2*tileGutter // 34
	objAtlasW  = 256                     // object-atlas shelf width in pixels
)

// amigaView is Turrican's visible playfield in pixels (one screen).
const amigaViewW, amigaViewH = 320, 256

// hflipMask re-encodes the map's flip convention (raw byte >= ntiles = tile-128, h-flipped)
// as an explicit cell flag bit per the common format (FORMAT2.md).
const hflipMask = 0x8000

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type objSprite struct {
	X int `json:"x"` // sprite's rect inside the world object atlas
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type jsonGrid struct {
	TileSize    int    `json:"tileSize"`
	Atlas       string `json:"atlas"`
	AtlasCols   int    `json:"atlasCols"`
	AtlasGutter int    `json:"atlasGutter"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Cells       []int  `json:"cells"`
	HflipMask   int    `json:"hflipMask"`
}

type jsonCollision struct {
	Kind   string            `json:"kind"`
	Sub    int               `json:"sub"`
	Solid  []int             `json:"solid"`
	Legend map[string]string `json:"legend"`
}

type jsonObject struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Sprite  string `json:"sprite"`            // key into the level's objSprites (renderKey)
	Name    string `json:"name"`              // info-card key = the frame-table id (stable object identity)
	Type    int    `json:"type"`              // placement type byte (scene-local; resolves to the AI handler)
	Orient  int    `json:"orient"`            // placement low byte -> node+$1E: the orientation the handler reads
	Frame   int    `json:"frame"`             // frame index the handler selected for this orientation
	Handler string `json:"handler,omitempty"` // AI handler routine address, hex
}

// spriteRef is one object-layer sprite entry: its rect(s) inside the world object atlas.
type spriteRef struct {
	Src    string   `json:"src"`
	Frames [][4]int `json:"frames"`
}

// extents is the format-2 tilemap2d extent (size in cells).
type extents struct {
	TileSize int `json:"tileSize"`
	Width    int `json:"width"`
	Height   int `json:"height"`
}

// jsonLevel is the format-2 envelope carrying the format-1 tilemap body inline plus the
// object-layer sprite atlas refs (objSprites) so the level is self-contained.
type jsonLevel struct {
	Format      int                  `json:"format"`
	Name        string               `json:"name"`
	Kind        string               `json:"kind"`
	Extents     extents              `json:"extents"`
	Wrap        string               `json:"wrap"`
	Grid        jsonGrid             `json:"grid"`
	Collision   jsonCollision        `json:"collision"`
	Spawn       point                `json:"spawn"`
	View        rect                 `json:"view"`
	Objects     []jsonObject         `json:"objects,omitempty"`
	ObjSprites  map[string]spriteRef `json:"objSprites,omitempty"`
	ObjectsFile string               `json:"objectsFile,omitempty"`
}

// objectsDB is the machine-readable object database written alongside each level.
type objectsDB struct {
	Format  int        `json:"format"`
	Level   string     `json:"level"`
	Objects []dbObject `json:"objects"`
}

type dbObject struct {
	ID    int            `json:"id"`
	Type  int            `json:"type,omitempty"`
	Name  string         `json:"name,omitempty"`
	Pos   []int          `json:"pos"`
	Props map[string]any `json:"props,omitempty"`
}

// spriteKey identifies one displayed sprite frame within a world: a frame table plus the
// frame index the object's handler selects for its orientation.
type spriteKey struct {
	ft       int
	resident bool
	frame    int
}

func (k spriteKey) infoKey(w int) string {
	s := fmt.Sprintf("w%d/ft%X", w, k.ft)
	if k.resident {
		s += "r"
	}
	return s
}
func (k spriteKey) renderKey(w int) string { return fmt.Sprintf("%s.%d", k.infoKey(w), k.frame) }

// exportLevels writes the levels/ tree (per-world atlases, per-scene format-2 tilemaps + object
// DBs) and returns the manifest level index.
func exportLevels(game *scene.Game, outDir string) ([]levelIndex, error) {
	levelsDir := filepath.Join(outDir, "levels")
	if err := os.MkdirAll(levelsDir, 0o755); err != nil {
		return nil, err
	}
	var levels []levelIndex
	for w := 0; w < scene.NumWorlds; w++ {
		block := game.Block(w).Data
		be := func(o int) int { return be32(block, o) }
		at := func(addr int) int { return addr - blockBase }

		pal := readPalette(block, at(be(0x08)))
		tableOff := at(be(0x00))
		nTiles := be32(block, tableOff) / 4

		atlasName := fmt.Sprintf("atlas%d.png", w)
		atlasVer, err := writeAtlas(filepath.Join(levelsDir, atlasName), block, tableOff, nTiles, pal)
		if err != nil {
			return nil, err
		}
		atlasURL := "levels/" + atlasName + "?v=" + atlasVer // root-relative + content-hash cache-buster

		// Per-tile collision (16 bytes/tile = 4x4 of 8x8-block solidity).
		collBytes, _ := game.TileCollision(w)
		collision := make([]int, len(collBytes))
		for i, b := range collBytes {
			collision[i] = int(b)
		}

		scenes := game.Scenes(w)

		// Resolve each placed object's displayed sprite/frame/position by running its AI
		// handler's spawn-init in the 68000 interpreter (scene.Sim).
		sim := game.NewSim(w)
		for si := range scenes {
			for oi := range scenes[si].Objects {
				game.Resolve(sim, &scenes[si].Objects[oi])
			}
		}

		// Collect every displayed sprite frame the world's objects use, then pack each into the
		// world object atlas.
		used := map[spriteKey]bool{}
		for _, sc := range scenes {
			for _, o := range sc.Objects {
				if o.FT != 0 {
					used[spriteKey{o.FT, o.Resident, o.Frame}] = true
				}
			}
		}
		objAtlasName := fmt.Sprintf("objatlas%d.png", w)
		rectOf, objVer, err := writeObjAtlas(filepath.Join(levelsDir, objAtlasName), game, w, used)
		if err != nil {
			return nil, err
		}
		objAtlasURL := "levels/" + objAtlasName + "?v=" + objVer
		// object-layer sprite refs (renderKey -> rect into the world object atlas), embedded in
		// each of the world's scene files so a level is self-contained.
		objSprites := map[string]spriteRef{}
		for k, r := range rectOf {
			objSprites[k.renderKey(w)] = spriteRef{
				Src: objAtlasURL, Frames: [][4]int{{r.X, r.Y, r.W, r.H}},
			}
		}

		for _, sc := range scenes {
			descOff := at(be(0x16 + sc.Index*4))
			mapOff := at(be(descOff + 0x00))
			if sc.Width <= 0 || sc.Height <= 0 || mapOff+sc.Width*sc.Height > len(block) {
				return nil, fmt.Errorf("world %d scene %d: bad map %dx%d", w, sc.Index, sc.Width, sc.Height)
			}
			cells := make([]int, sc.Width*sc.Height)
			for col := 0; col < sc.Width; col++ {
				for row := 0; row < sc.Height; row++ {
					v := int(block[mapOff+col*sc.Height+row]) // col-major -> row-major
					if v >= nTiles {                          // raw flip convention -> explicit flag bit
						v = (v - 128) | hflipMask
					}
					cells[row*sc.Width+col] = v
				}
			}

			var objects []jsonObject
			for _, o := range sc.Objects {
				k := spriteKey{o.FT, o.Resident, o.Frame}
				if _, ok := rectOf[k]; !ok {
					continue // no sprite resolved (unhandled type) — skip
				}
				obj := jsonObject{
					X: o.X, Y: o.Y,
					Sprite: k.renderKey(w), Name: k.infoKey(w),
					Type: o.Type, Orient: o.Orient, Frame: o.Frame,
				}
				if o.Handler != 0 {
					obj.Handler = fmt.Sprintf("%X", o.Handler)
				}
				objects = append(objects, obj)
			}

			stem := fmt.Sprintf("world%d_scene%d", w, sc.Index)
			objectsFile := stem + ".objects.json"
			lvl := jsonLevel{
				Format:  2,
				Name:    fmt.Sprintf("World %d · Scene %d", w+1, sc.Index+1),
				Kind:    "tilemap2d",
				Extents: extents{TileSize: tileSide, Width: sc.Width, Height: sc.Height},
				Wrap:    "none",
				Grid: jsonGrid{
					TileSize: tileSide, Atlas: atlasURL, AtlasCols: atlasCols, AtlasGutter: tileGutter,
					Width: sc.Width, Height: sc.Height, Cells: cells, HflipMask: hflipMask,
				},
				Collision: jsonCollision{
					Kind: "grid", Sub: 4, Solid: collision,
					Legend: map[string]string{"1": "#ff3030", "127": "#33ddff", "128": "#ffe020", "211": "#ff33cc"},
				},
				Spawn:       point{sc.SpawnX, sc.SpawnY},
				View:        rect{sc.CamX, sc.CamY, amigaViewW, amigaViewH},
				Objects:     objects,
				ObjSprites:  objSprites,
				ObjectsFile: objectsFile,
			}
			if err := writeJSON(filepath.Join(levelsDir, stem+".json"), lvl); err != nil {
				return nil, err
			}

			// machine-readable object DB (sibling of the level file)
			db := objectsDB{Format: 2, Level: lvl.Name}
			for i, o := range objects {
				props := map[string]any{"sprite": o.Sprite, "orient": o.Orient, "frame": o.Frame}
				if o.Handler != "" {
					props["handler"] = o.Handler
				}
				db.Objects = append(db.Objects, dbObject{
					ID: i, Type: o.Type, Name: o.Name, Pos: []int{o.X, o.Y}, Props: props,
				})
			}
			if err := writeJSON(filepath.Join(levelsDir, objectsFile), db); err != nil {
				return nil, err
			}

			levels = append(levels, levelIndex{
				Name:    fmt.Sprintf("Scene %d", sc.Index+1),
				Section: fmt.Sprintf("World %d", w+1),
				File:    "levels/" + stem + ".json",
				Kind:    "tilemap2d",
				Atlas:   atlasURL,
				Objects: "levels/" + objectsFile,
			})
			fmt.Fprintf(os.Stderr, "[levels] world %d scene %d: %dx%d, %d tiles, %d objects -> %s.json\n",
				w, sc.Index, sc.Width, sc.Height, nTiles, len(objects), stem)
		}
	}
	fmt.Fprintf(os.Stderr, "[levels] done: %d scenes across %d worlds\n", len(levels), scene.NumWorlds)
	return levels, nil
}

// writeObjAtlas shelf-packs the first frame of each used sprite into one paletted PNG (index 0
// transparent) and returns each sprite's rect in it.
func writeObjAtlas(path string, game *scene.Game, w int, used map[spriteKey]bool) (map[spriteKey]objSprite, string, error) {
	// Deterministic order: residents first, then by address.
	keys := make([]spriteKey, 0, len(used))
	for k := range used {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].resident != keys[j].resident {
			return keys[i].resident
		}
		if keys[i].ft != keys[j].ft {
			return keys[i].ft < keys[j].ft
		}
		return keys[i].frame < keys[j].frame
	})

	rectOf := map[spriteKey]objSprite{}
	type placed struct {
		sp   scene.Space
		f    scene.Frame
		x, y int
	}
	var ps []placed
	cx, cy, shelfH, atlasH := 0, 0, 0, 0
	for _, k := range keys {
		sp := game.Space(w, k.resident)
		f, ok := game.FrameAt(w, k.ft, k.frame, k.resident)
		if !ok {
			continue
		}
		pw := f.DW * 16
		if cx+pw > objAtlasW {
			cy += shelfH
			cx, shelfH = 0, 0
		}
		ps = append(ps, placed{sp: sp, f: f, x: cx, y: cy})
		rectOf[k] = objSprite{X: cx, Y: cy, W: pw, H: f.H}
		cx += pw
		if f.H > shelfH {
			shelfH = f.H
		}
		if cy+shelfH > atlasH {
			atlasH = cy + shelfH
		}
	}
	if atlasH == 0 {
		return rectOf, "", nil
	}
	img := image.NewPaletted(image.Rect(0, 0, objAtlasW, atlasH), game.WorldPalette(w, true))
	for _, p := range ps {
		scene.DrawFirstFrame(img, p.sp, p.f, p.x, p.y)
	}
	ver, err := writePNG(path, img)
	return rectOf, ver, err
}

// writeAtlas packs the world's 32x32 tiles into a grid PNG, each tile in a
// (tileSide+2*tileGutter)-pixel cell whose 1-pixel border duplicates the tile's edge pixels
// (extrusion guard against atlas bleed at fractional zoom).
func writeAtlas(path string, block []byte, tableOff, nTiles int, pal color.Palette) (string, error) {
	rows := (nTiles + atlasCols - 1) / atlasCols
	img := image.NewPaletted(image.Rect(0, 0, atlasCols*atlasCell, rows*atlasCell), pal)
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v >= tileSide {
			return tileSide - 1
		}
		return v
	}
	for n := 0; n < nTiles; n++ {
		off := tableOff + int(binary.BigEndian.Uint32(block[tableOff+n*4:]))
		if off+tileBytes > len(block) {
			break
		}
		var t [tileSide][tileSide]uint8
		for y := 0; y < tileSide; y++ {
			var planes [4]uint32
			for p := 0; p < 4; p++ {
				planes[p] = binary.BigEndian.Uint32(block[off+(y*4+p)*4:])
			}
			for x := 0; x < tileSide; x++ {
				var v uint8
				for p := 0; p < 4; p++ {
					v |= uint8((planes[p]>>(31-uint(x)))&1) << uint(p)
				}
				t[y][x] = v
			}
		}
		ox, oy := (n%atlasCols)*atlasCell, (n/atlasCols)*atlasCell
		for y := -tileGutter; y < tileSide+tileGutter; y++ {
			for x := -tileGutter; x < tileSide+tileGutter; x++ {
				img.SetColorIndex(ox+tileGutter+x, oy+tileGutter+y, t[clamp(y)][clamp(x)])
			}
		}
	}
	return writePNG(path, img)
}

// readPalette reads a 16-colour Amiga playfield palette (fully opaque, for the tile atlas).
func readPalette(block []byte, off int) color.Palette {
	pal := make(color.Palette, 16)
	for i := range pal {
		c := binary.BigEndian.Uint16(block[off+i*2:])
		pal[i] = color.RGBA{R: uint8((c>>8)&0xF) * 17, G: uint8((c>>4)&0xF) * 17, B: uint8(c&0xF) * 17, A: 255}
	}
	return pal
}
