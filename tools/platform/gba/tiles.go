package gba

// Offline decoding of the AGB's tile formats: the 4bpp/8bpp character data, the
// 15-bit palettes, and the text-mode tilemaps that arrange them.
//
// This is the *export* path, and it is deliberately a separate implementation
// from the PPU's scanline renderer in gbamachine/ppu.go. The two answer
// different questions — the PPU asks "what does the wire carry on line Y, with
// this scroll, these windows and this blend", while an exporter asks "what does
// the whole map look like" — and forcing one to serve both makes each worse.
//
// The duplication is guarded rather than tolerated: gbamachine's tests render a
// synthetic scene through BOTH paths and require the pixels to agree, so the
// two cannot silently drift apart.
//
// Everything here works on a plain VRAM/palette byte slice, so it decodes a
// live machine's memory, a savestate's, or a region lifted out of a ROM.

import (
	"image"
	"image/color"
	"sort"
)

// Palette is the 15-bit BGR colour table: 256 background entries followed by
// 256 object entries, as palette RAM stores them.
type Palette []uint16

// ParsePalette decodes palette RAM (1 KiB) into colour entries.
func ParsePalette(palRAM []byte) Palette {
	p := make(Palette, len(palRAM)/2)
	for i := range p {
		p[i] = uint16(palRAM[i*2]) | uint16(palRAM[i*2+1])<<8
	}
	return p
}

// RGBA expands entry i to a colour. Index 0 of every 16-colour bank is the
// transparent entry; callers that care pass it through Transparent instead.
func (p Palette) RGBA(i int) color.RGBA {
	if i < 0 || i >= len(p) {
		return color.RGBA{}
	}
	c := p[i]
	exp := func(v uint16) uint8 { return uint8(v<<3 | v>>2) }
	return color.RGBA{R: exp(c & 31), G: exp(c >> 5 & 31), B: exp(c >> 10 & 31), A: 255}
}

// TileRow returns the 8 palette indices of row ty of tile n. bpp is 4 or 8; for
// 4bpp the indices are within the tile's 16-colour bank and 0 means transparent.
func TileRow(vram []byte, charBase uint32, n, ty, bpp int) [8]byte {
	var out [8]byte
	if bpp == 8 {
		off := charBase + uint32(n)*64 + uint32(ty)*8
		if int(off)+8 > len(vram) {
			return out
		}
		copy(out[:], vram[off:off+8])
		return out
	}
	off := charBase + uint32(n)*32 + uint32(ty)*4
	if int(off)+4 > len(vram) {
		return out
	}
	for i := 0; i < 4; i++ {
		b := vram[off+uint32(i)]
		out[i*2] = b & 0xF
		out[i*2+1] = b >> 4
	}
	return out
}

// TileSheet renders count tiles from charBase as a 16-tile-wide sheet — the
// character data as the artist drew it, independent of any map that uses it.
// For 4bpp, palBank selects which of the 16 palette banks to colour it with.
func TileSheet(vram []byte, pal Palette, charBase uint32, count, bpp, palBank int) *image.RGBA {
	const perRow = 16
	rows := (count + perRow - 1) / perRow
	img := image.NewRGBA(image.Rect(0, 0, perRow*8, rows*8))
	for n := 0; n < count; n++ {
		ox, oy := (n%perRow)*8, (n/perRow)*8
		for ty := 0; ty < 8; ty++ {
			row := TileRow(vram, charBase, n, ty, bpp)
			for tx := 0; tx < 8; tx++ {
				idx := int(row[tx])
				if bpp == 4 {
					if idx == 0 {
						continue // transparent
					}
					idx += palBank * 16
				} else if idx == 0 {
					continue
				}
				img.SetRGBA(ox+tx, oy+ty, pal.RGBA(idx))
			}
		}
	}
	return img
}

// BGLayer describes one text-mode background, decoded from its BGxCNT register.
type BGLayer struct {
	Priority                int
	CharBase                uint32 // byte offset into VRAM of the character data
	ScrBase                 uint32 // byte offset into VRAM of the screen (tilemap) data
	EightBpp                bool
	Mosaic                  bool
	WidthTiles, HeightTiles int
}

// DecodeBGCNT decodes a BGxCNT register value.
func DecodeBGCNT(cnt uint16) BGLayer {
	b := BGLayer{
		Priority:   int(cnt & 3),
		CharBase:   uint32(cnt>>2&3) * 0x4000,
		Mosaic:     cnt&(1<<6) != 0,
		EightBpp:   cnt&(1<<7) != 0,
		ScrBase:    uint32(cnt>>8&31) * 0x800,
		WidthTiles: 32, HeightTiles: 32,
	}
	switch cnt >> 14 & 3 {
	case 1:
		b.WidthTiles = 64
	case 2:
		b.HeightTiles = 64
	case 3:
		b.WidthTiles, b.HeightTiles = 64, 64
	}
	return b
}

// MapEntry is one text-mode tilemap cell.
type MapEntry struct {
	Tile    int
	HFlip   bool
	VFlip   bool
	PalBank int
}

// MapEntryAt reads the tilemap cell at tile coordinates (tx, ty). Text-mode
// maps larger than 32x32 are stored as up to four 32x32 SCREENBLOCKS laid out
// left-to-right, top-to-bottom — not as one wide array, which is the mistake
// that makes the right half of a 512-wide map come out as garbage.
func (b BGLayer) MapEntryAt(vram []byte, tx, ty int) MapEntry {
	quad := 0
	if b.WidthTiles > 32 && tx >= 32 {
		quad++
	}
	if b.HeightTiles > 32 && ty >= 32 {
		quad += b.WidthTiles / 32
	}
	off := b.ScrBase + uint32(quad)*0x800 + uint32(ty%32)*64 + uint32(tx%32)*2
	if int(off)+2 > len(vram) {
		return MapEntry{}
	}
	e := uint16(vram[off]) | uint16(vram[off+1])<<8
	return MapEntry{
		Tile:    int(e & 0x3FF),
		HFlip:   e&(1<<10) != 0,
		VFlip:   e&(1<<11) != 0,
		PalBank: int(e >> 12),
	}
}

// Render draws the whole background at its native size (up to 512x512), with
// transparent pixels left transparent so layers can be inspected separately.
func (b BGLayer) Render(vram []byte, pal Palette) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, b.WidthTiles*8, b.HeightTiles*8))
	bpp := 4
	if b.EightBpp {
		bpp = 8
	}
	for ty := 0; ty < b.HeightTiles; ty++ {
		for tx := 0; tx < b.WidthTiles; tx++ {
			e := b.MapEntryAt(vram, tx, ty)
			for py := 0; py < 8; py++ {
				sy := py
				if e.VFlip {
					sy = 7 - py
				}
				row := TileRow(vram, b.CharBase, e.Tile, sy, bpp)
				for px := 0; px < 8; px++ {
					sx := px
					if e.HFlip {
						sx = 7 - px
					}
					idx := int(row[sx])
					if idx == 0 {
						continue // transparent
					}
					if bpp == 4 {
						idx += e.PalBank * 16
					}
					img.SetRGBA(tx*8+px, ty*8+py, pal.RGBA(idx))
				}
			}
		}
	}
	return img
}

// TilesUsed reports the distinct tile indices the map references, and the
// highest one — how much of the character data this map actually uses, which
// is what bounds a useful tile-sheet export.
func (b BGLayer) TilesUsed(vram []byte) (set map[int]bool, max int) {
	set = map[int]bool{}
	for ty := 0; ty < b.HeightTiles; ty++ {
		for tx := 0; tx < b.WidthTiles; tx++ {
			n := b.MapEntryAt(vram, tx, ty).Tile
			set[n] = true
			if n > max {
				max = n
			}
		}
	}
	return set, max
}

// TilePalBanks reports, for each tile the map references, the 16-colour palette
// bank it is drawn with. A 4bpp tile carries no colour of its own — the MAP
// entry chooses the bank — so a tile sheet rendered with bank 0 for everything
// is not the game's art, it is one arbitrary recolouring of it. Where a tile is
// used with more than one bank the most frequent wins, and the count of such
// tiles is returned so a caller can say how much of the sheet is ambiguous.
func (b BGLayer) TilePalBanks(vram []byte) (banks map[int]int, multi int) {
	counts := map[int]map[int]int{}
	for ty := 0; ty < b.HeightTiles; ty++ {
		for tx := 0; tx < b.WidthTiles; tx++ {
			e := b.MapEntryAt(vram, tx, ty)
			if counts[e.Tile] == nil {
				counts[e.Tile] = map[int]int{}
			}
			counts[e.Tile][e.PalBank]++
		}
	}
	banks = map[int]int{}
	for tile, cs := range counts {
		if len(cs) > 1 {
			multi++
		}
		best, bestN := 0, -1
		for bank, n := range cs {
			if n > bestN || (n == bestN && bank < best) {
				best, bestN = bank, n
			}
		}
		banks[tile] = best
	}
	return banks, multi
}

// TileSheetBanked renders a tile sheet colouring each tile with its own palette
// bank (see TilePalBanks). Tiles absent from banks fall back to bank 0.
func TileSheetBanked(vram []byte, pal Palette, charBase uint32, count, bpp int, banks map[int]int) *image.RGBA {
	const perRow = 16
	rows := (count + perRow - 1) / perRow
	img := image.NewRGBA(image.Rect(0, 0, perRow*8, rows*8))
	for n := 0; n < count; n++ {
		ox, oy := (n%perRow)*8, (n/perRow)*8
		bank := banks[n]
		for ty := 0; ty < 8; ty++ {
			row := TileRow(vram, charBase, n, ty, bpp)
			for tx := 0; tx < 8; tx++ {
				idx := int(row[tx])
				if idx == 0 {
					continue
				}
				if bpp == 4 {
					idx += bank * 16
				}
				img.SetRGBA(ox+tx, oy+ty, pal.RGBA(idx))
			}
		}
	}
	return img
}

// TileSheetUsed renders ONLY the tiles a map references, in index order, each
// with the palette bank that map pairs it with. This is usually the sheet worth
// looking at: a background's character base commonly holds several screens'
// worth of tiles, so a full-range sheet is mostly art the current scene never
// mentions, rendered in an arbitrary palette. It returns the sheet and the tile
// indices in the order drawn, so a caller can map a cell back to its index.
func TileSheetUsed(vram []byte, pal Palette, charBase uint32, bpp int, used map[int]bool, banks map[int]int) (*image.RGBA, []int) {
	idx := make([]int, 0, len(used))
	for n := range used {
		idx = append(idx, n)
	}
	sort.Ints(idx)

	const perRow = 16
	rows := (len(idx) + perRow - 1) / perRow
	if rows == 0 {
		rows = 1
	}
	img := image.NewRGBA(image.Rect(0, 0, perRow*8, rows*8))
	for i, n := range idx {
		ox, oy := (i%perRow)*8, (i/perRow)*8
		bank := banks[n]
		for ty := 0; ty < 8; ty++ {
			row := TileRow(vram, charBase, n, ty, bpp)
			for tx := 0; tx < 8; tx++ {
				v := int(row[tx])
				if v == 0 {
					continue
				}
				if bpp == 4 {
					v += bank * 16
				}
				img.SetRGBA(ox+tx, oy+ty, pal.RGBA(v))
			}
		}
	}
	return img, idx
}

// --- rooms: the metatile layer -----------------------------------------------

// A Minish Cap room is not stored as a tilemap. It is stored as a grid of
// METATILE ids, each of which indexes an 8-byte table entry holding four tile
// entries — the 2x2 block that id expands to. The game's expander (traced to
// the routine at $0801AB40) walks the id array writing two tile rows at a time,
// which is why the expanded map's row stride is twice the metatile grid's.
//
// This is a common 2-D console idiom rather than a Zelda invention: it buys a
// 4x smaller map at the cost of one indirection, and it is why searching a ROM
// for something that looks like a tilemap finds nothing.
type Room struct {
	// IDs is the metatile grid, row-major, WidthMeta entries per row.
	IDs        []uint16
	WidthMeta  int
	HeightMeta int
	// Table is the metatile definition block: 8 bytes (4 tile entries) each,
	// ordered top-left, top-right, bottom-left, bottom-right.
	Table []byte
}

// MetatileFlag marks an id that does not index the table directly. The game
// compares each id against 0x3FFF and sends anything above it down a separate
// path (an animated or scripted block); those are emitted as tile 0 here and
// counted, rather than silently indexing past the table.
const MetatileFlag = 0x3FFF

// Expand builds the tile map: WidthMeta*2 tiles per row, HeightMeta*2 rows.
// Returns the map as raw 16-bit entries in the layout the game's expander
// produces, and how many ids were above MetatileFlag.
func (r Room) Expand() (tiles []uint16, widthTiles, heightTiles, flagged int) {
	widthTiles = r.WidthMeta * 2
	heightTiles = r.HeightMeta * 2
	tiles = make([]uint16, widthTiles*heightTiles)
	for my := 0; my < r.HeightMeta; my++ {
		for mx := 0; mx < r.WidthMeta; mx++ {
			i := my*r.WidthMeta + mx
			if i >= len(r.IDs) {
				continue
			}
			id := int(r.IDs[i])
			if id > MetatileFlag {
				flagged++
				continue
			}
			off := id * 8
			if off+8 > len(r.Table) {
				continue
			}
			q := func(n int) uint16 {
				return uint16(r.Table[off+n*2]) | uint16(r.Table[off+n*2+1])<<8
			}
			top := my*2*widthTiles + mx*2
			bot := top + widthTiles
			tiles[top] = q(0)
			tiles[top+1] = q(1)
			tiles[bot] = q(2)
			tiles[bot+1] = q(3)
		}
	}
	return tiles, widthTiles, heightTiles, flagged
}

// RenderTiles draws a raw tile-entry map (as Expand produces) with the given
// character data and palette.
func RenderTiles(tiles []uint16, widthTiles, heightTiles int, vram []byte, pal Palette, charBase uint32, bpp int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, widthTiles*8, heightTiles*8))
	for ty := 0; ty < heightTiles; ty++ {
		for tx := 0; tx < widthTiles; tx++ {
			e := tiles[ty*widthTiles+tx]
			tile := int(e & 0x3FF)
			hflip, vflip := e&(1<<10) != 0, e&(1<<11) != 0
			bank := int(e >> 12)
			for py := 0; py < 8; py++ {
				sy := py
				if vflip {
					sy = 7 - py
				}
				row := TileRow(vram, charBase, tile, sy, bpp)
				for px := 0; px < 8; px++ {
					sx := px
					if hflip {
						sx = 7 - px
					}
					v := int(row[sx])
					if v == 0 {
						continue
					}
					if bpp == 4 {
						v += bank * 16
					}
					img.SetRGBA(tx*8+px, ty*8+py, pal.RGBA(v))
				}
			}
		}
	}
	return img
}
