package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"

	"retroreverse.com/tools/platform/gba"
)

// assembled is one area rendered onto its canvas.
type assembled struct {
	img              *image.RGBA
	originX, originY int
	rooms            int
}

// buildArea decodes every room of an area and composites it at its own
// position. Rooms whose data does not decode are skipped and counted rather
// than aborting the area — a partially assembled place with a stated gap is
// more useful than nothing.
func buildArea(rom []byte, ar *Area) (*assembled, error) {
	// The canvas is the bounding box of the rooms' own coordinates.
	minX, minY := 1<<30, 1<<30
	maxX, maxY := 0, 0
	for _, r := range ar.Rooms {
		if int(r.X) < minX {
			minX = int(r.X)
		}
		if int(r.Y) < minY {
			minY = int(r.Y)
		}
		if int(r.X)+int(r.W) > maxX {
			maxX = int(r.X) + int(r.W)
		}
		if int(r.Y)+int(r.H) > maxY {
			maxY = int(r.Y) + int(r.H)
		}
	}
	if maxX <= minX || maxY <= minY {
		return nil, fmt.Errorf("area %d has an empty extent", ar.ID)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, maxX-minX, maxY-minY))

	metaList, err := ParseAssetList(rom, ar.MetaList)
	if err != nil {
		return nil, fmt.Errorf("metatile list: %w", err)
	}
	unpack := func(addr uint32) ([]byte, error) {
		if addr == 0 {
			return nil, fmt.Errorf("no stream")
		}
		return gba.LZ77Decompress(rom[gba.ROMOffset(addr):])
	}

	placed := 0
	for i, rec := range ar.Rooms {
		roomList, err := ParseAssetList(rom, ar.RoomLists[i])
		if err != nil {
			continue
		}
		var tsetList []AssetEntry
		if int(rec.TilesetIdx) < len(ar.TsetLists) {
			tsetList, _ = ParseAssetList(rom, ar.TsetLists[rec.TilesetIdx])
		}
		assets, err := Resolve(roomList, metaList, tsetList)
		if err != nil {
			continue
		}

		// Character data: three 16 KiB banks, back to back.
		vram := make([]byte, 0, 3*16384)
		for _, a := range assets.Tilesets {
			b, err := unpack(a)
			if err != nil || len(b) < 16384 {
				b = make([]byte, 16384)
			}
			vram = append(vram, b[:16384]...)
		}

		pal := gba.Palette(BuildPalette(rom, PaletteSetOf(rom, tsetList)))
		w := rec.WidthMeta()
		if w <= 0 {
			continue
		}

		var layers []*image.RGBA
		for _, l := range []struct {
			ids, tab uint32
			bank     int
		}{
			{assets.IDs[0], assets.Tables[0], 0},
			{assets.IDs[1], assets.Tables[1], 1},
		} {
			idb, err1 := unpack(l.ids)
			tab, err2 := unpack(l.tab)
			if err1 != nil || err2 != nil {
				continue
			}
			h := len(idb) / 2 / w
			if h <= 0 {
				continue
			}
			// The record's height is authoritative; the stream can be longer.
			if hm := rec.HeightMeta(); hm > 0 && hm < h {
				h = hm
			}
			room := gba.Room{WidthMeta: w, HeightMeta: h, Table: tab}
			room.IDs = make([]uint16, len(idb)/2)
			for j := range room.IDs {
				room.IDs[j] = binary.LittleEndian.Uint16(idb[j*2:])
			}
			tiles, tw, th, _ := room.Expand()
			layers = append(layers, gba.RenderTiles(tiles, tw, th, vram[l.bank*16384:], pal, 0, 4))
		}
		if len(layers) == 0 {
			continue
		}
		at := image.Pt(int(rec.X)-minX, int(rec.Y)-minY)
		for _, ly := range layers {
			r := ly.Bounds().Add(at)
			draw.Draw(canvas, r, ly, image.Point{}, draw.Over)
		}
		placed++
	}
	if placed == 0 {
		return nil, fmt.Errorf("area %d: no room decoded", ar.ID)
	}
	return &assembled{canvas, minX, minY, placed}, nil
}
