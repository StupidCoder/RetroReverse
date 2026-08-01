// areamap assembles an AREA — every room of it, placed where the game's own
// tables say it sits — into one image, straight from the ROM.
//
// This is what turns 600 room maps into a few dozen places: a room record
// carries its position on the area's shared pixel canvas, so the rooms of an
// area tile together into the map a player actually walks through.
//
//	areamap -list                 # every area, with its rooms
//	areamap -area 3 -out DIR      # assemble one area
//	areamap -all -out DIR         # assemble all of them
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"retroreverse.com/tools/platform/gba"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "areamap: "+format+"\n", a...)
	os.Exit(2)
}

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
		// The room's objects, at their own coordinates plus the room origin.
		if markObjects {
			for k := 0; k < 4; k++ {
				l := ListFor(rom, ar.ID, i, k)
				if l == 0 {
					continue
				}
				for _, ob := range Objects(rom, l, k) {
					drawMarker(canvas, at.X+ob.X, at.Y+ob.Y, kindColour(k))
				}
			}
		}
		placed++
	}
	if placed == 0 {
		return nil, fmt.Errorf("area %d: no room decoded", ar.ID)
	}
	return &assembled{canvas, minX, minY, placed}, nil
}

// markObjects draws the room's object records over the assembled map — the
// visual check on the three-level object lookup: a correct decode puts markers
// on doors, chests and signs, and a wrong one scatters them at random.
var markObjects bool

func kindColour(k int) color.RGBA {
	switch k {
	case 0:
		return color.RGBA{255, 80, 80, 255}
	case 1:
		return color.RGBA{80, 255, 80, 255}
	case 2:
		return color.RGBA{80, 160, 255, 255}
	default:
		return color.RGBA{255, 230, 60, 255}
	}
}

func drawMarker(img *image.RGBA, x, y int, c color.RGBA) {
	for dy := -3; dy <= 3; dy++ {
		for dx := -3; dx <= 3; dx++ {
			if dx*dx+dy*dy > 9 {
				continue
			}
			px, py := x+dx, y+dy
			if px < 0 || py < 0 || px >= img.Bounds().Dx() || py >= img.Bounds().Dy() {
				continue
			}
			img.SetRGBA(px, py, c)
		}
	}
}

func main() {
	romPath := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	list := flag.Bool("list", false, "list every area and its rooms")
	area := flag.Int("area", 3, "area to assemble")
	all := flag.Bool("all", false, "assemble every area")
	maxArea := flag.Int("max", 128, "highest area id to consider")
	out := flag.String("out", "areas", "output directory")
	marks := flag.Bool("objects", false, "draw the rooms' object records as markers")
	flag.Parse()

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		die("%v", err)
	}
	markObjects = *marks

	if *list {
		total := 0
		for a := 0; a < *maxArea; a++ {
			ar, err := LoadArea(rom, a)
			if err != nil {
				continue
			}
			total += len(ar.Rooms)
			w, h := 0, 0
			for _, r := range ar.Rooms {
				if int(r.X)+int(r.W) > w {
					w = int(r.X) + int(r.W)
				}
				if int(r.Y)+int(r.H) > h {
					h = int(r.Y) + int(r.H)
				}
			}
			fmt.Printf("area %3d: %2d rooms, extent %5dx%-5d meta list %08X\n", a, len(ar.Rooms), w, h, ar.MetaList)
		}
		fmt.Printf("%d rooms across the areas listed\n", total)
		return
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		die("%v", err)
	}
	areas := []int{*area}
	if *all {
		areas = nil
		for a := 0; a < *maxArea; a++ {
			areas = append(areas, a)
		}
	}
	made := 0
	for _, a := range areas {
		ar, err := LoadArea(rom, a)
		if err != nil {
			continue
		}
		as, err := buildArea(rom, ar)
		if err != nil {
			fmt.Printf("area %3d: %v\n", a, err)
			continue
		}
		name := filepath.Join(*out, fmt.Sprintf("area-%02d.png", a))
		writePNG(name, as.img)
		fmt.Printf("area %3d: %2d/%2d rooms placed, %dx%d px at origin (%d,%d) -> %s\n",
			a, as.rooms, len(ar.Rooms), as.img.Bounds().Dx(), as.img.Bounds().Dy(),
			as.originX, as.originY, name)
		made++
	}
	fmt.Printf("assembled %d area(s)\n", made)
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
