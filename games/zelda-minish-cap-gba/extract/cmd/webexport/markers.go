package main

// The marker assets. No object sprites are decoded yet, so each marker style
// is a small generated shape — a disc per entity class, a diamond for the
// slot-3 command records — published as a one-frame sprite2d asset the level
// placements point at.

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
)

const markerCell = 16

type markerStyle struct {
	id, name, desc string
	fill           color.RGBA
	diamond        bool
}

var markerStyles = []markerStyle{
	{"enemy", "Enemy",
		"A class-3 entity record. The class is identified structurally: slot-2 records of this class carry a respawn index (0..31) that the room loader checks against a per-room kill flag before spawning — the mechanism that keeps a defeated enemy dead. The behaviour handler comes from the game's own dispatch table at $080D3BF8, indexed by the record's id byte.",
		color.RGBA{225, 60, 60, 255}, false},
	{"object", "Object",
		"A class-6 entity record — the placed props of the world: pots, bushes, chests and their kin. The behaviour handler comes from the dispatch table at $080B2D4C, indexed by the record's id byte; equal ids mean equal behaviour everywhere.",
		color.RGBA{70, 200, 90, 255}, false},
	{"doorway", "Doorway",
		"Object id 00: point records that tile an entry region — six of them cover the door of Link's house in a 16-pixel grid. The sub byte selects where the door leads; the transition itself reads an 8-byte {area, room, x, y} record from the table at $080FCA20.",
		color.RGBA{80, 140, 255, 255}, false},
	{"npc", "NPC",
		"A class-7 entity record. 88 of the cartridge's 111 NPC records attach a ROM pointer at spawn (mode 4 via $0807DAD0) — a script for the 139-opcode interpreter at $0807DFBC, which is where an NPC's dialogue and behaviour live. The handler table is at $080B313C, 12 bytes per id.",
		color.RGBA{70, 200, 230, 255}, false},
	{"tile-patch", "Tile patch",
		"A slot-3 command record (op 10, or op 2 for the inverse): when the named save flag is set, the room loader rewrites the metatile at this position ($0807B314). This is how a revealed secret — a burned bush, an opened passage — persists on the map.",
		color.RGBA{235, 200, 60, 255}, true},
	{"spawn-marker", "Manager spawn",
		"A slot-3 command record (op 4): spawns manager 0x24 at this pixel position with two byte parameters ($0804B300). Managers are room-logic entities; this is the only command that carries a pixel position of its own.",
		color.RGBA{200, 110, 255, 255}, true},
}

// writeMarkers draws each style's PNG and sprite2d doc and returns the
// manifest asset entries.
func writeMarkers(outDir string) []any {
	dir := filepath.Join(outDir, "objects")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		die("%v", err)
	}
	var assets []any
	for _, st := range markerStyles {
		writePNG(filepath.Join(dir, st.id+".png"), drawMarkerCell(st))
		writeJSON(filepath.Join(dir, st.id+".json"), map[string]any{
			"format": "retro-x", "version": 1, "type": "sprite2d", "name": st.name,
			"atlas": map[string]any{
				"file": st.id + ".png", "cellW": markerCell, "cellH": markerCell,
				"anchor": []int{markerCell / 2, markerCell / 2},
			},
			"animations": []any{
				map[string]any{"id": "main", "loop": "loop", "frames": 1, "durations": []int{30}},
			},
		})
		assets = append(assets, map[string]any{
			"id": st.id, "category": "object", "name": st.name, "group": "Markers",
			"description": st.desc,
			"file":        "objects/" + st.id + ".json",
		})
	}
	return assets
}

// drawMarkerCell renders one 16x16 marker: a filled disc or diamond with a
// dark rim and a light core so it reads on any map underneath.
func drawMarkerCell(st markerStyle) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, markerCell, markerCell))
	c := markerCell / 2
	rim := color.RGBA{20, 20, 30, 255}
	core := color.RGBA{
		uint8(min(int(st.fill.R)+90, 255)),
		uint8(min(int(st.fill.G)+90, 255)),
		uint8(min(int(st.fill.B)+90, 255)), 255}
	for y := 0; y < markerCell; y++ {
		for x := 0; x < markerCell; x++ {
			dx, dy := x-c, y-c
			var d int
			if st.diamond {
				d = (abs(dx) + abs(dy)) * 10
			} else {
				d = isqrt((dx*dx + dy*dy) * 100)
			}
			switch {
			case d <= 25:
				img.SetRGBA(x, y, core)
			case d <= 55:
				img.SetRGBA(x, y, st.fill)
			case d <= 70:
				img.SetRGBA(x, y, rim)
			}
		}
	}
	return img
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isqrt(v int) int {
	r := 0
	for (r+1)*(r+1) <= v {
		r++
	}
	return r
}
