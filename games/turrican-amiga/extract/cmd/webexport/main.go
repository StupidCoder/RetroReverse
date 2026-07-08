// webexport renders Turrican's disk-streamed level maps for the companion
// website's tile viewer. For each world it writes a tile atlas PNG (the world's
// 32x32 tiles in its palette) and, for each scene in that world, a JSON file with
// the column-major map flattened to a row-major cell grid. A meta.json lists every
// scene. The viewer resolves the flip flag (cell >= ntiles -> tile cell-128,
// horizontally flipped) just like extract/cmd/map.
//
// It also emits the object layer: a per-world object atlas (objatlas<w>.png) holding
// the first animation frame of every sprite the scenes' objects use, and per scene
// an `objects` list (each object's pixel position and a sprite index into the
// scene's `objSprites` rects). Placement comes straight off the disk via the
// extract/scene package (the scroll-triggered spawner's grid + entry lists); the
// sprite each object uses is resolved type -> AI handler -> the frame table it
// installs (see Turrican.md Part V §6).
//
// Usage: webexport [-o dir] [Turrican.adf]
package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
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

type objSprite struct {
	X int `json:"x"` // sprite's rect inside the world object atlas
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

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

// hflipMask re-encodes the map's flip convention (raw byte >= ntiles = tile-128,
// h-flipped) as an explicit cell flag bit per the common format (site/FORMAT.md).
const hflipMask = 0x8000

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
	Sprite  string `json:"sprite"`
	Name    string `json:"name"`              // info-card key = the frame-table id (Turrican's stable object identity; the placement `type` is only scene-local)
	Type    int    `json:"type"`              // placement type byte (scene-local; resolves to the AI handler)
	Orient  int    `json:"orient"`            // placement low byte -> node+$1E: the orientation/direction the handler reads
	Frame   int    `json:"frame"`             // frame index the handler selected for this orientation (node+$C)
	Handler string `json:"handler,omitempty"` // AI handler routine address, hex (the steady-state behaviour after spawn-init)
}

type jsonLevel struct {
	Format    int           `json:"format"`
	Name      string        `json:"name"`
	Grid      jsonGrid      `json:"grid"`
	Collision jsonCollision `json:"collision"`
	Spawn     point         `json:"spawn"` // player spawn, world pixels (no sprite -> crosshair)
	View      rect          `json:"view"`  // the Amiga on-screen viewport at the spawn, world pixels
	Objects   []jsonObject  `json:"objects,omitempty"`
}

// amigaView is Turrican's visible playfield in pixels (one screen).
const amigaViewW, amigaViewH = 320, 256

type metaLevel struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"`
	Atlas   string `json:"atlas"`
}

// spriteKey identifies one displayed sprite frame within a world: a frame table plus the
// frame index the object's handler selects for its orientation (resident and scene tables
// share no addresses, but the flag keeps the mapping explicit).
type spriteKey struct {
	ft       int
	resident bool
	frame    int
}

// infoKey is the frame-independent object identity (the AI frame table) used as the
// info-card key; renderKey adds the frame so each orientation packs its own atlas rect.
func (k spriteKey) infoKey(w int) string {
	s := fmt.Sprintf("w%d/ft%X", w, k.ft)
	if k.resident {
		s += "r"
	}
	return s
}
func (k spriteKey) renderKey(w int) string { return fmt.Sprintf("%s.%d", k.infoKey(w), k.frame) }

func main() {
	out := flag.String("o", "site/public/turrican", "output directory")
	flag.Parse()
	adfPath := flag.Arg(0)
	if adfPath == "" {
		adfPath = "Turrican (Amiga)/Turrican.adf"
	}
	if err := run(adfPath, *out); err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}

func run(adfPath, outDir string) error {
	adf, err := os.ReadFile(adfPath)
	if err != nil {
		return err
	}
	game, err := scene.Load(adf)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(outDir, "sprites"), 0o755); err != nil {
		return err
	}
	var meta []metaLevel
	spriteIndex := map[string]any{}
	for w := 0; w < scene.NumWorlds; w++ {
		block := game.Block(w).Data
		be32 := func(o int) int { return int(binary.BigEndian.Uint32(block[o:])) }
		at := func(addr int) int { return addr - blockBase }

		pal := readPalette(block, at(be32(0x08)))
		tableOff := at(be32(0x00))
		nTiles := be32(tableOff) / 4

		atlasName := fmt.Sprintf("atlas%d.png", w)
		atlasVer, err := writeAtlas(filepath.Join(outDir, atlasName), block, tableOff, nTiles, pal)
		if err != nil {
			return err
		}
		atlasURL := atlasName + "?v=" + atlasVer // content-hash cache-buster (see writePNG)

		// Per-tile collision (16 bytes/tile = 4x4 of 8x8-block solidity).
		collBytes, _ := game.TileCollision(w)
		collision := make([]int, len(collBytes))
		for i, b := range collBytes {
			collision[i] = int(b)
		}

		scenes := game.Scenes(w)

		// Resolve each placed object's displayed sprite/frame/position by running its
		// AI handler's spawn-init in the 68000 interpreter (scene.Sim) — reproducing the
		// engine's orientation/direction handling instead of always showing frame 0.
		sim := game.NewSim(w)
		for si := range scenes {
			for oi := range scenes[si].Objects {
				game.Resolve(sim, &scenes[si].Objects[oi])
			}
		}

		// Collect every displayed sprite frame the world's objects use, then pack each
		// one into the world object atlas.
		used := map[spriteKey]bool{}
		for _, sc := range scenes {
			for _, o := range sc.Objects {
				if o.FT != 0 {
					used[spriteKey{o.FT, o.Resident, o.Frame}] = true
				}
			}
		}
		objAtlasName := fmt.Sprintf("sprites/objatlas%d.png", w)
		rectOf, objVer, err := writeObjAtlas(filepath.Join(outDir, objAtlasName), game, w, used)
		if err != nil {
			return err
		}
		objAtlasURL := objAtlasName + "?v=" + objVer // content-hash cache-buster (see writePNG)
		// world-scoped render keys (frame table + frame) -> single-frame entries.
		for k, r := range rectOf {
			spriteIndex[k.renderKey(w)] = map[string]any{
				"src":    objAtlasURL,
				"frames": [][4]int{{r.X, r.Y, r.W, r.H}},
			}
		}

		for _, sc := range scenes {
			descOff := at(be32(0x16 + sc.Index*4))
			mapOff := at(be32(descOff + 0x00))
			if sc.Width <= 0 || sc.Height <= 0 || mapOff+sc.Width*sc.Height > len(block) {
				return fmt.Errorf("world %d scene %d: bad map %dx%d", w, sc.Index, sc.Width, sc.Height)
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

			file := fmt.Sprintf("world%d_scene%d.json", w, sc.Index)
			lvl := jsonLevel{
				Format: 1,
				Name:   fmt.Sprintf("World %d · Scene %d", w+1, sc.Index+1),
				Grid: jsonGrid{
					TileSize: tileSide, Atlas: atlasURL, AtlasCols: atlasCols, AtlasGutter: tileGutter,
					Width: sc.Width, Height: sc.Height, Cells: cells, HflipMask: hflipMask,
				},
				Collision: jsonCollision{
					Kind: "grid", Sub: 4, Solid: collision,
					Legend: map[string]string{"1": "#ff3030", "127": "#33ddff", "128": "#ffe020", "211": "#ff33cc"},
				},
				Spawn:   point{sc.SpawnX, sc.SpawnY},
				View:    rect{sc.CamX, sc.CamY, amigaViewW, amigaViewH},
				Objects: objects,
			}
			if err := writeJSON(filepath.Join(outDir, file), lvl); err != nil {
				return err
			}
			meta = append(meta, metaLevel{
				Name:    fmt.Sprintf("Scene %d", sc.Index+1),
				Section: fmt.Sprintf("World %d", w+1),
				File:    file, Atlas: atlasURL,
			})
			fmt.Printf("world %d scene %d: %dx%d, %d tiles, %d objects -> %s\n", w, sc.Index, sc.Width, sc.Height, nTiles, len(objects), file)
		}
	}
	if err := writeJSON(filepath.Join(outDir, "sprites", "index.json"), spriteIndex); err != nil {
		return err
	}
	return writeJSON(filepath.Join(outDir, "meta.json"), map[string]any{
		"format": 1, "game": "turrican",
		"native": map[string]int{"w": amigaViewW, "h": amigaViewH},
		"tickHz": 50,
		"levels": meta,
	})
}

// writeObjAtlas shelf-packs the first frame of each used sprite into one paletted PNG
// (index 0 transparent) and returns each sprite's rect in it.
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
		f, ok := game.FrameAt(w, k.ft, k.frame, k.resident) // the orientation's frame
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

// writePNG encodes img, hashes the bytes and writes the file, returning a short
// content-hash cache-buster (`?v=<hash>`). The atlas layout changes whenever the object
// set or a resolved frame changes, so a stale browser-cached atlas would sample the new
// index.json rects at the wrong pixels (sprites replaced with fragments of others); the
// versioned URL forces a fresh fetch. Empty img (no atlas) returns "".
func writePNG(path string, img *image.Paletted) (string, error) {
	if img.Bounds().Empty() {
		return "", nil
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(buf.Bytes()))[:8], nil
}

// writeAtlas packs the world's 32x32 tiles into a grid PNG. Each tile sits in a
// (tileSide+2*tileGutter)-pixel cell whose 1-pixel border duplicates the tile's edge
// pixels (extrusion): when the viewer samples a sub-pixel past a tile's edge at
// fractional zoom, it hits the tile's own colour instead of bleeding the neighbour.
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

func readPalette(block []byte, off int) color.Palette {
	pal := make(color.Palette, 16)
	for i := range pal {
		c := binary.BigEndian.Uint16(block[off+i*2:])
		pal[i] = color.RGBA{R: uint8((c>>8)&0xF) * 17, G: uint8((c>>4)&0xF) * 17, B: uint8(c&0xF) * 17, A: 255}
	}
	return pal
}

func writeJSON(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
