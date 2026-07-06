// Command levinfo decodes LEV.ARK: it lists the archive blocks and, for a chosen
// level, prints the tile-type map and a per-field summary — the reimplementation
// that verifies the format derived in Part V. It renders the tile map to a PNG
// too (walls dark, floors by height).
//
// Usage: levinfo [-game ../game] [-level 0] [-png out.png]
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"ultimaunderworld/extract/lev"
)

func main() {
	game := flag.String("game", "../game", "path to the game/ folder")
	level := flag.Int("level", 0, "level block index (0-8 are the nine dungeon levels)")
	pngOut := flag.String("png", "", "render the tile map to this PNG file")
	flag.Parse()

	data, err := os.ReadFile(filepath.Join(*game, "DATA", "LEV.ARK"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "levinfo:", err)
		os.Exit(1)
	}
	ark, err := lev.ParseArk(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "levinfo:", err)
		os.Exit(1)
	}
	fmt.Printf("LEV.ARK: %d blocks, %d bytes\n", len(ark.Offsets), len(data))
	nLevels := 0
	for i := range ark.Offsets {
		if b, err := ark.Block(i); err == nil && len(b) == lev.LevelBlockSize {
			nLevels++
		}
	}
	fmt.Printf("%d blocks are %d-byte level blocks\n\n", nLevels, lev.LevelBlockSize)

	block, err := ark.Block(*level)
	if err != nil {
		fmt.Fprintln(os.Stderr, "levinfo:", err)
		os.Exit(1)
	}
	g, err := lev.DecodeGrid(block)
	if err != nil {
		fmt.Fprintln(os.Stderr, "levinfo:", err)
		os.Exit(1)
	}

	// Field summaries.
	types := map[uint8]int{}
	heights := map[uint8]int{}
	objTiles := 0
	for _, t := range g.Tiles {
		types[t.Type]++
		heights[t.Height]++
		if t.Object != 0 {
			objTiles++
		}
	}
	fmt.Printf("level %d: %dx%d tiles\n", *level, g.W, g.H)
	fmt.Printf("  type histogram: %s\n", histo(types))
	fmt.Printf("  height histogram: %s\n", histo(heights))
	fmt.Printf("  tiles with objects: %d\n\n", objTiles)

	fmt.Printf("tile-type map (# solid, . open, / diagonal, s slope):\n")
	for y := 0; y < g.H; y++ {
		row := make([]byte, g.W)
		for x := 0; x < g.W; x++ {
			row[x] = lev.TypeGlyph(g.At(x, y).Type)
		}
		fmt.Println(string(row))
	}

	if *pngOut != "" {
		if err := renderPNG(g, *pngOut); err != nil {
			fmt.Fprintln(os.Stderr, "levinfo: png:", err)
		} else {
			fmt.Printf("\nwrote %s\n", *pngOut)
		}
	}
}

func histo(m map[uint8]int) string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	s := ""
	for _, k := range keys {
		s += fmt.Sprintf("%d:%d ", k, m[uint8(k)])
	}
	return s
}

// renderPNG draws the map at 8px/tile: solid rock dark, open floor shaded by
// height, diagonals/slopes in accent colours. Y is flipped so north is up.
func renderPNG(g *lev.Grid, path string) error {
	const s = 8
	img := image.NewRGBA(image.Rect(0, 0, g.W*s, g.H*s))
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			t := g.At(x, y)
			var c color.RGBA
			switch {
			case t.Type == lev.TileSolid:
				c = color.RGBA{20, 18, 16, 255}
			case t.Type == lev.TileOpen:
				v := uint8(60 + int(t.Height)*10)
				c = color.RGBA{v, v - 20, v - 40, 255}
			case t.Type <= lev.TileDiagNW:
				c = color.RGBA{80, 60, 90, 255}
			default:
				c = color.RGBA{60, 90, 80, 255}
			}
			dy := (g.H - 1 - y) * s
			for py := 0; py < s; py++ {
				for px := 0; px < s; px++ {
					img.Set(x*s+px, dy+py, c)
				}
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
