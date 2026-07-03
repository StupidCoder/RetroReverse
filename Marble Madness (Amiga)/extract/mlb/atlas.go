package mlb

import (
	"image"
	"image/color"
)

// Course is a decoded .mlb ready to emit as an atlas + tilemap (the form the web
// viewer composes): the row-major tilemap of 16-bit tile indices (CourseW wide), the
// height in tiles, the number of distinct tiles, and the 16-colour palette.
//
// H counts the tilemap DATA rows; PlayableH is the unpacked buffer's leading word —
// the row count the engine actually plays (the scroll clamp is (PlayableH-25)*8, so
// rows beyond it can never scroll on screen). Rows past PlayableH are off-screen
// storage: Ultimate carries 90 (three hidden 30-row variants of its final screen for
// the path-swap animation), Silly one cropped wall row.
type Course struct {
	Cells     []int // row-major, CourseW wide; len = CourseW*H
	H         int
	PlayableH int
	NTiles    int
	Palette   color.Palette
	PalWords  [16]uint16 // the stored $0RGB palette words, before any fill-in
	buf       []byte
	planes    [4]int
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
	playable := 0
	if len(buf) >= 2 {
		playable = int(buf[0])<<8 | int(buf[1])
	}
	if playable <= 0 || playable > h {
		playable = h
	}
	co := &Course{Cells: cells, H: h, PlayableH: playable, NTiles: maxIdx + 1, Palette: palette(buf), buf: buf, planes: planes}
	for i := 0; i < 16; i++ {
		if o := 0x16 + 2*i; o+1 < len(buf) {
			co.PalWords[i] = uint16(buf[o])<<8 | uint16(buf[o+1])
		}
	}
	return co
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
	return c.AtlasVariants(cols, c.Palette, nil)
}

// TileMask returns a bitmask of the palette indices (bit i = index i) that tile
// n's pixels use — how the exporter finds tiles touched by the display list's
// per-band COLOR10-15 overrides.
func (c *Course) TileMask(n int) uint16 {
	t := c.tile(n)
	var m uint16
	for _, row := range t {
		for _, v := range row {
			m |= 1 << v
		}
	}
	return m
}

// Variant is an extra atlas tile: source tile Src re-rendered with its palette
// indices remapped (band recolours / shimmer phases). Remap targets index into
// the extended palette passed to AtlasVariants.
type Variant struct {
	Src   int
	Remap map[uint8]uint8
}

// AtlasVariants renders the base tiles followed by the variant tiles into a
// cols-wide gutter-extruded sheet. pal[0:16] must equal the course palette
// (base tiles keep their raw indices); variants may reference appended colours.
func (c *Course) AtlasVariants(cols int, pal color.Palette, variants []Variant) *image.Paletted {
	const tile, gut = 8, 1
	const cell = tile + 2*gut
	n := c.NTiles + len(variants)
	rows := (n + cols - 1) / cols
	img := image.NewPaletted(image.Rect(0, 0, cols*cell, rows*cell), pal)
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v >= tile {
			return tile - 1
		}
		return v
	}
	put := func(slot int, t [8][8]uint8, remap map[uint8]uint8) {
		ox, oy := (slot%cols)*cell+gut, (slot/cols)*cell+gut
		for y := -gut; y < tile+gut; y++ {
			for x := -gut; x < tile+gut; x++ {
				v := t[clamp(y)][clamp(x)]
				if r, ok := remap[v]; ok {
					v = r
				}
				img.SetColorIndex(ox+x, oy+y, v)
			}
		}
	}
	for i := 0; i < c.NTiles; i++ {
		put(i, c.tile(i), nil)
	}
	for i, v := range variants {
		put(c.NTiles+i, c.tile(v.Src), v.Remap)
	}
	return img
}
