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
	marks            []marker
}

// marker is one object or command record, placed in canvas coordinates and
// carrying the identifiers the Studio's info popup shows.
type marker struct {
	style string // the marker asset to draw
	name  string
	x, y  int
	props map[string]any
}

// markersFor turns one room's records into markers at the room's canvas
// offset. Managers are skipped: the spawner never places them ($0804ADF8
// skips the position step for class 9), so drawing their record bytes as a
// position would be an invention. The same honesty bounds everything else to
// the room: a few species reuse the position bytes for other data (enemy 49
// packs something else there), and a cmd-2 patch can name a tile in another
// room of the area — a record whose position falls outside the room it is
// listed in is not a position we can defend, so it gets no marker.
func markersFor(rom []byte, area, room int, at image.Point, w, h int) []marker {
	var out []marker
	inRoom := func(x, y int) bool { return x >= 0 && x <= w && y >= 0 && y <= h }
	for slot := 0; slot < 3; slot++ {
		l := ListFor(rom, area, room, slot)
		if l == 0 {
			continue
		}
		for _, o := range Objects(rom, l, slot) {
			if !o.Placed() || !inRoom(o.X, o.Y) {
				continue
			}
			style, name := entityStyle(o)
			props := map[string]any{
				"class": fmt.Sprintf("%d (%s)", o.Class, ClassNames[o.Class]),
				"id":    fmt.Sprintf("0x%02X", o.ID), "sub": fmt.Sprintf("0x%02X", o.Sub),
				"slot": o.Slot, "record": "0x" + o.Addr,
			}
			if o.Handler != 0 {
				props["handler"] = fmt.Sprintf("0x%08X", o.Handler)
			}
			if o.Mode == 4 {
				props["script"] = fmt.Sprintf("0x%08X", o.Extra)
			} else if o.Extra != 0 {
				props["extra"] = fmt.Sprintf("0x%08X", o.Extra)
			}
			if o.Mode != 0 {
				props["mode"] = o.Mode
			}
			if o.KillIdx >= 0 {
				props["respawn slot"] = o.KillIdx
			}
			if o.B4 != 0 || o.B5 != 0 {
				props["b4/b5"] = fmt.Sprintf("%02X/%02X", o.B4, o.B5)
			}
			if o.H6 != 0 {
				props["h6"] = fmt.Sprintf("0x%04X", o.H6)
			}
			out = append(out, marker{style, name, at.X + o.X, at.Y + o.Y, props})
		}
	}
	if l := ListFor(rom, area, room, 3); l != 0 {
		for _, c := range Commands(rom, l) {
			x, y, ok := c.Pos()
			if !ok || !inRoom(x, y) {
				continue
			}
			style, name := "tile-patch", ""
			props := map[string]any{"op": c.Op, "record": "0x" + c.Addr}
			switch c.Op {
			case 2:
				name = fmt.Sprintf("tile patch (flag 0x%02X)", c.B1)
				props["flag"] = fmt.Sprintf("0x%02X", c.B1)
				props["metatile"] = "0x74"
			case 10:
				name = fmt.Sprintf("tile patch (flag 0x%02X)", c.H2)
				props["flag"] = fmt.Sprintf("0x%02X", c.H2)
				props["metatile"] = fmt.Sprintf("0x%02X", c.H6)
			case 4:
				style, name = "spawn-marker", "manager spawn"
				props["manager"] = "0x24"
				props["params"] = fmt.Sprintf("%02X/%02X", c.B1, c.B2)
			}
			out = append(out, marker{style, name, at.X + x, at.Y + y, props})
		}
	}
	return out
}

// entityStyle picks the marker asset and display name for an entity record.
// Only identifications the decode actually supports get names: the doorway
// grids (object 00, six point records tiling the door of Link's house) and
// the postbox (object 2D, the marker lands on the drawn postbox).
func entityStyle(o Object) (style, name string) {
	switch o.Class {
	case 3:
		return "enemy", fmt.Sprintf("enemy 0x%02X", o.ID)
	case 7:
		return "npc", fmt.Sprintf("NPC 0x%02X", o.ID)
	case 6:
		switch o.ID {
		case 0x00:
			return "doorway", fmt.Sprintf("doorway (sub 0x%02X)", o.Sub)
		case 0x2D:
			return "object", "postbox"
		}
		return "object", fmt.Sprintf("object 0x%02X", o.ID)
	}
	return "object", fmt.Sprintf("%s 0x%02X", ClassNames[o.Class], o.ID)
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
	var marks []marker
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
		marks = append(marks, markersFor(rom, ar.ID, i, at, int(rec.W), int(rec.H))...)
		placed++
	}
	if placed == 0 {
		return nil, fmt.Errorf("area %d: no room decoded", ar.ID)
	}
	return &assembled{canvas, minX, minY, placed, marks}, nil
}
