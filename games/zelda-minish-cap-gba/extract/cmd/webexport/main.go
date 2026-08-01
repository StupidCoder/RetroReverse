// webexport publishes the decoded room to the Studio as a Retro-X game.
//
// The room is decoded from the ROM (the same path roomexport verifies against
// the running game), the two background layers are composited as the hardware
// stacks them, and the result is emitted as a tilemap: an atlas of DISTINCT
// 8x8 cells plus an index grid.
//
// Deduplicating by pixel CONTENT rather than by the game's tile id is
// deliberate. A GBA cell is a tile id plus a palette bank and two flip bits, so
// one id renders as up to 64 different pictures; keying the atlas on the id
// alone would collapse them into one and repaint half the map wrong. Content is
// what the viewer actually draws.
//
//	webexport [-out DIR]
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"retroreverse.com/tools/platform/gba"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "webexport: "+format+"\n", a...)
	os.Exit(2)
}

func main() {
	romPath := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	palPath := flag.String("palette", "../work/starting-area/palette.bin", "palette captured by mapexport")
	out := flag.String("out", "../../../site/public/zelda-minish-cap-gba", "Retro-X game directory")
	flag.Parse()

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		die("%v", err)
	}
	palRAM, err := os.ReadFile(*palPath)
	if err != nil {
		die("palette: %v (run mapexport first)", err)
	}
	pal := gba.ParsePalette(palRAM)

	// The starting area, by its asset lists (see roomexport).
	const (
		roomListAddr    = 0x081034D0
		metaListAddr    = 0x0810279C
		tilesetListAddr = 0x08100E88
	)
	roomList, err := ParseAssetList(rom, roomListAddr)
	if err != nil {
		die("room list: %v", err)
	}
	metaList, err := ParseAssetList(rom, metaListAddr)
	if err != nil {
		die("metatile list: %v", err)
	}
	tsList, err := ParseAssetList(rom, tilesetListAddr)
	if err != nil {
		die("tileset list: %v", err)
	}
	assets, err := Resolve(roomList, metaList, tsList)
	if err != nil {
		die("%v", err)
	}

	unpack := func(addr uint32) []byte {
		b, err := gba.LZ77Decompress(rom[gba.ROMOffset(addr):])
		if err != nil {
			die("%08X: %v", addr, err)
		}
		return b
	}
	vram := make([]byte, 0, 3*16384)
	for _, a := range assets.Tilesets {
		if a == 0 {
			vram = append(vram, make([]byte, 16384)...)
			continue
		}
		vram = append(vram, unpack(a)...)
	}

	const widthMeta = 63
	var layers []*image.RGBA
	var wt, ht int
	for _, l := range []struct {
		ids, tab uint32
		charBase int
	}{
		{assets.IDs[0], assets.Tables[0], 0},
		{assets.IDs[1], assets.Tables[1], 1},
	} {
		idb := unpack(l.ids)
		room := gba.Room{WidthMeta: widthMeta, HeightMeta: len(idb) / 2 / widthMeta, Table: unpack(l.tab)}
		room.IDs = make([]uint16, len(idb)/2)
		for j := range room.IDs {
			room.IDs[j] = binary.LittleEndian.Uint16(idb[j*2:])
		}
		tiles, w, h, _ := room.Expand()
		wt, ht = w, h
		layers = append(layers, gba.RenderTiles(tiles, w, h, vram[l.charBase*16384:], pal, 0, 4))
	}

	// Composite BG1 over BG2, as the hardware stacks them.
	comp := image.NewRGBA(layers[0].Bounds())
	draw.Draw(comp, comp.Bounds(), layers[0], image.Point{}, draw.Src)
	draw.Draw(comp, comp.Bounds(), layers[1], image.Point{}, draw.Over)

	// Slice into 8x8 cells, deduplicated by content.
	index := map[[32]byte]int{}
	var cellImgs []*image.RGBA
	cells := make([]int, wt*ht)
	for ty := 0; ty < ht; ty++ {
		for tx := 0; tx < wt; tx++ {
			sub := image.NewRGBA(image.Rect(0, 0, 8, 8))
			draw.Draw(sub, sub.Bounds(), comp, image.Pt(tx*8, ty*8), draw.Src)
			k := sha256.Sum256(sub.Pix)
			id, ok := index[k]
			if !ok {
				id = len(cellImgs)
				index[k] = id
				cellImgs = append(cellImgs, sub)
			}
			cells[ty*wt+tx] = id
		}
	}

	const cols = 16
	rows := (len(cellImgs) + cols - 1) / cols
	atlas := image.NewRGBA(image.Rect(0, 0, cols*8, rows*8))
	for i, c := range cellImgs {
		x, y := (i%cols)*8, (i/cols)*8
		draw.Draw(atlas, image.Rect(x, y, x+8, y+8), c, image.Point{}, draw.Src)
	}
	fmt.Printf("%dx%d tiles, %d distinct cells -> atlas %dx%d\n", wt, ht, len(cellImgs), cols*8, rows*8)

	levelsDir := filepath.Join(*out, "levels")
	if err := os.MkdirAll(levelsDir, 0o755); err != nil {
		die("%v", err)
	}
	writePNG(filepath.Join(levelsDir, "atlas_0.png"), atlas)

	level := map[string]any{
		"format": "retro-x", "version": 1, "type": "tilemap",
		"camera": map[string]any{"mode": "map2d", "map2d": map[string]any{"maxNativeFactor": 1}},
		"tilemap": map[string]any{
			"tileSize": 8, "width": wt, "height": ht,
			"atlas": map[string]any{"file": "atlas_0.png", "cols": cols},
			"cells": cells,
			// Where the game's own camera sat when the decode was verified.
			"view": map[string]any{"x": 56 * 8, "y": 35 * 8, "w": 240, "h": 160},
		},
	}
	writeJSON(filepath.Join(levelsDir, "hyrule-field.json"), level)

	man := map[string]any{
		"format": "retro-x", "version": 1,
		"id": "zelda-minish-cap-gba", "title": "The Legend of Zelda: The Minish Cap",
		"platform": "Game Boy Advance", "year": 2004,
		"description": "The 2004 Game Boy Advance Zelda, developed by Capcom/Flagship — internally, as its own build string admits, \"ZELDA 5\". This map is decoded straight from the cartridge. A room is not stored as a tilemap at all: it is a grid of metatile ids, each expanding through an 8-byte table into a 2x2 block of tiles, all of it behind the BIOS's LZ77 compression, which is why nothing tilemap-shaped is findable in the ROM. The decode is checked against the game itself — the decompressed tables are byte-identical to what the console's own BIOS call produces, and the expanded map reproduces every tile the running game uploaded to video memory.",
		"display": map[string]any{
			"native": map[string]any{"w": 240, "h": 160}, "tickHz": 60,
		},
		"assets": []any{
			map[string]any{
				"id": "hyrule-field", "category": "level", "name": "Hyrule Field (starting area)",
				"group":       "Overworld",
				"description": "The field outside Link's house, where the game begins: 63x43 metatiles — 126x86 tiles — with the river and its bridges to the west, Link's house at the centre, and the wall of Hyrule Castle across the north. Both background layers are composited as the hardware stacks them, so the overlay layer's tree canopies sit over the terrain layer's shadows.",
				"file":        "levels/hyrule-field.json",
			},
		},
	}
	writeJSON(filepath.Join(*out, "manifest.json"), man)
	fmt.Println("wrote", *out)
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		die("%v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		die("%v", err)
	}
}

func writeJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		die("%v", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		die("%v", err)
	}
}
