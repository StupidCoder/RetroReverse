package mlb

import (
	"image"
	"image/color"
)

// Course is a decoded .mlb ready to emit as an atlas + tilemap (the form the web
// viewer composes): the row-major tilemap of 16-bit tile indices (CourseW wide), the
// height in tiles, the number of distinct tiles, and the 16-colour palette.
type Course struct {
	Cells   []int // row-major, CourseW wide; len = CourseW*H
	H       int
	NTiles  int
	Palette color.Palette
	buf     []byte
	planes  [4]int
}

// Decode unpacks a .mlb into its tilemap, tile count and palette (without compositing
// the course; use Atlas for the tile graphics).
func Decode(file []byte) *Course {
	buf := unpackByteRun1(file)
	lw := func(o int) int {
		if o+4 > len(buf) {
			return 0
		}
		return int(buf[o])<<24 | int(buf[o+1])<<16 | int(buf[o+2])<<8 | int(buf[o+3])
	}
	planes := [4]int{lw(2), lw(6), lw(10), lw(14)}
	tilemap := buf[lw(0x12):]
	h := (len(tilemap) / 2) / CourseW
	cells := make([]int, CourseW*h)
	maxIdx := 0
	for i := 0; i < len(cells); i++ {
		idx := int(tilemap[i*2])<<8 | int(tilemap[i*2+1])
		cells[i] = idx
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	return &Course{Cells: cells, H: h, NTiles: maxIdx + 1, Palette: palette(buf), buf: buf, planes: planes}
}

// tile decodes one 8x8 tile (index n) from the bitplane data.
func (c *Course) tile(n int) [8][8]uint8 {
	var t [8][8]uint8
	b0 := (n>>1)*16 + (n & 1)
	for r := 0; r < 8; r++ {
		var bp [4]byte
		for p := 0; p < 4; p++ {
			if o := c.planes[p] + b0 + 2*r; o >= 0 && o < len(c.buf) {
				bp[p] = c.buf[o]
			}
		}
		for x := 0; x < 8; x++ {
			bit := uint(7 - x)
			t[r][x] = bp[0]>>bit&1 | (bp[1]>>bit&1)<<1 | (bp[2]>>bit&1)<<2 | (bp[3]>>bit&1)<<3
		}
	}
	return t
}

// Atlas renders the course's NTiles tiles into a cols-wide sheet with a 1px extruded
// gutter, matching the format the web composer slices (8x8 tiles, gutter 1).
func (c *Course) Atlas(cols int) *image.Paletted {
	const tile, gut = 8, 1
	const cell = tile + 2*gut
	rows := (c.NTiles + cols - 1) / cols
	img := image.NewPaletted(image.Rect(0, 0, cols*cell, rows*cell), c.Palette)
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v >= tile {
			return tile - 1
		}
		return v
	}
	for n := 0; n < c.NTiles; n++ {
		t := c.tile(n)
		ox, oy := (n%cols)*cell+gut, (n/cols)*cell+gut
		for y := -gut; y < tile+gut; y++ {
			for x := -gut; x < tile+gut; x++ {
				img.SetColorIndex(ox+x, oy+y, t[clamp(y)][clamp(x)])
			}
		}
	}
	return img
}
