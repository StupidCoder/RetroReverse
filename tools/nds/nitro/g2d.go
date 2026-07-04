package nitro

import (
	"fmt"
	"image"
	"image/color"
)

// The NITRO 2D graphics pipeline — the DS's tile-based UI art — is three files that
// mirror the 2D hardware's three memories:
//
//	NCLR ("RLCN", block TTLP) — a palette: BGR555 colours for palette RAM;
//	NCGR ("RGCN", block RAHC) — character graphics: 8x8 tiles at 4bpp or 8bpp,
//	                            exactly as they load into tile VRAM;
//	NSCR ("RCSN", block NRCS) — a screen: the 32x32-style tilemap of 16-bit entries
//	                            (tile number, H/V flip, palette row) for map VRAM.
//
// Composing screen→tiles→palette reproduces the image the 2D engine would display.
// The same blocks appear NARC-embedded inside .carc bundles.

// NCLR is a decoded palette file.
type NCLR struct {
	Colors []color.NRGBA // all colours, in palette order
	Is8bpp bool
}

// NCGR is a decoded character-graphics (tile) file.
type NCGR struct {
	Tiles  [][]byte // each tile: 64 palette indices, row-major 8x8
	Is8bpp bool
}

// NSCR is a decoded screen (tilemap) file.
type NSCR struct {
	W, H    int      // pixels
	Entries []uint16 // W/8 × H/8 map entries
}

// nitroBlock finds a top-level block by its (byte-reversed on disk) magic inside a
// generic NITRO container file and returns the offset of the block header.
func nitroBlock(data []byte, magic string) (int, error) {
	if len(data) < 0x10 {
		return 0, fmt.Errorf("nitro: file too short")
	}
	hdrSize := int(le.Uint16(data[0x0C:]))
	n := int(le.Uint16(data[0x0E:]))
	off := hdrSize
	for i := 0; i < n && off+8 <= len(data); i++ {
		if string(data[off:off+4]) == magic {
			return off, nil
		}
		size := int(le.Uint32(data[off+4:]))
		if size <= 0 {
			break
		}
		off += size
	}
	return 0, fmt.Errorf("nitro: no %s block", magic)
}

// ParseNCLR decodes a palette file (NCLR / embedded RLCN; the SDK also emits the
// RPCN variant magic, seen on the debug font's palette).
func ParseNCLR(data []byte) (*NCLR, error) {
	if len(data) < 4 || (string(data[0:4]) != "RLCN" && string(data[0:4]) != "RPCN") {
		return nil, fmt.Errorf("nitro: not an NCLR")
	}
	b, err := nitroBlock(data, "TTLP")
	if err != nil {
		return nil, err
	}
	depth := le.Uint32(data[b+8:]) // 3 = 4bpp, 4 = 8bpp
	size := int(le.Uint32(data[b+0x10:]))
	if b+0x18+size > len(data) {
		size = len(data) - b - 0x18
	}
	cols := make([]color.NRGBA, size/2)
	for i := range cols {
		cols[i] = bgr555(le.Uint16(data[b+0x18+i*2:]))
	}
	return &NCLR{Colors: cols, Is8bpp: depth == 4}, nil
}

// ParseNCGR decodes a character-graphics file (NCGR / embedded RGCN).
func ParseNCGR(data []byte) (*NCGR, error) {
	if len(data) < 4 || string(data[0:4]) != "RGCN" {
		return nil, fmt.Errorf("nitro: not an NCGR")
	}
	b, err := nitroBlock(data, "RAHC")
	if err != nil {
		return nil, err
	}
	depth := le.Uint32(data[b+0x0C:]) // 3 = 4bpp, 4 = 8bpp
	is8 := depth == 4
	size := int(le.Uint32(data[b+0x18:]))
	tileBytes := 32
	if is8 {
		tileBytes = 64
	}
	dataOff := b + 0x20
	if dataOff+size > len(data) {
		size = len(data) - dataOff
	}
	n := size / tileBytes
	tiles := make([][]byte, n)
	for t := 0; t < n; t++ {
		px := make([]byte, 64)
		src := data[dataOff+t*tileBytes:]
		if is8 {
			copy(px, src[:64])
		} else {
			for i := 0; i < 32; i++ {
				px[i*2] = src[i] & 0xF
				px[i*2+1] = src[i] >> 4
			}
		}
		tiles[t] = px
	}
	return &NCGR{Tiles: tiles, Is8bpp: is8}, nil
}

// ParseNSCR decodes a screen (tilemap) file (NSCR / embedded RCSN).
func ParseNSCR(data []byte) (*NSCR, error) {
	if len(data) < 4 || string(data[0:4]) != "RCSN" {
		return nil, fmt.Errorf("nitro: not an NSCR")
	}
	b, err := nitroBlock(data, "NRCS")
	if err != nil {
		return nil, err
	}
	w := int(le.Uint16(data[b+8:]))
	h := int(le.Uint16(data[b+10:]))
	size := int(le.Uint32(data[b+0x10:]))
	dataOff := b + 0x14
	if dataOff+size > len(data) {
		size = len(data) - dataOff
	}
	entries := make([]uint16, size/2)
	for i := range entries {
		entries[i] = le.Uint16(data[dataOff+i*2:])
	}
	return &NSCR{W: w, H: h, Entries: entries}, nil
}

// ComposeScreen renders a screen through its tiles and palette: each 16-bit map
// entry is (bits 0-9) a tile number, (10) H-flip, (11) V-flip, (12-15) the palette
// row used for 4bpp tiles. Palette index 0 is transparent (the backdrop shows
// through on hardware).
func ComposeScreen(scr *NSCR, chr *NCGR, pal *NCLR) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, scr.W, scr.H))
	tw := scr.W / 8
	for i, e := range scr.Entries {
		tx, ty := (i%tw)*8, (i/tw)*8
		if ty >= scr.H {
			break
		}
		tile := int(e & 0x3FF)
		hf, vf := e&(1<<10) != 0, e&(1<<11) != 0
		palRow := int(e>>12) & 0xF
		if tile >= len(chr.Tiles) {
			continue
		}
		px := chr.Tiles[tile]
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				sx, sy := x, y
				if hf {
					sx = 7 - x
				}
				if vf {
					sy = 7 - y
				}
				idx := int(px[sy*8+sx])
				var c color.NRGBA
				if idx != 0 { // index 0 = transparent
					ci := idx
					if !chr.Is8bpp {
						ci += palRow * 16
					}
					if ci < len(pal.Colors) {
						c = pal.Colors[ci]
					}
				}
				img.SetNRGBA(tx+x, ty+y, c)
			}
		}
	}
	return img
}

// TileSheet renders an NCGR's tiles as a sheet (16 tiles per row) with one palette
// row — for inspecting tile data that has no screen file.
func TileSheet(chr *NCGR, pal *NCLR, palRow int) *image.NRGBA {
	cols := 16
	rows := (len(chr.Tiles) + cols - 1) / cols
	img := image.NewNRGBA(image.Rect(0, 0, cols*8, rows*8))
	for t, px := range chr.Tiles {
		tx, ty := (t%cols)*8, (t/cols)*8
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				idx := int(px[y*8+x])
				var c color.NRGBA
				if idx != 0 {
					ci := idx
					if !chr.Is8bpp {
						ci += palRow * 16
					}
					if ci < len(pal.Colors) {
						c = pal.Colors[ci]
					}
				}
				img.SetNRGBA(tx+x, ty+y, c)
			}
		}
	}
	return img
}
