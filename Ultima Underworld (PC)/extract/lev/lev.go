// Package lev decodes Ultima Underworld's level archive, LEV.ARK, and the tile
// map inside each level block.
//
// The format was derived, not looked up: the archive layout from the file
// itself (the block count + offset table), and the tile-word fields from the
// game's OWN decode code, read out of the oracle's live memory while it played
// the dungeon —
//
//	tile TYPE   = word0 bits 0-3   (2CD3:0740: MOV AX,[tile]; AND AX,0x000F)
//	FLOOR HGT   = word0 bits 4-7   (2CD3:0835: SHR AX,4; AND AX,0x000F)
//	OBJECT head = word1 bits 6-15  (28B3:08F9: MOV AX,[tile+2]; SHR 6; AND 0x3FF)
//
// then confirmed against the data: decoding block 0 as a 64x64 grid of 4-byte
// tiles and printing the types yields a coherent dungeon of rooms and corridors.
// See disasm/ and Ultima_Underworld.md Part V.
package lev

import (
	"encoding/binary"
	"fmt"
)

// Ark is a parsed LEV.ARK: a count-prefixed table of variable-length blocks.
// Blocks 0..8 are the nine dungeon levels (31752 bytes each); the rest are the
// per-level texture-mapping, automap and object-animation blocks.
type Ark struct {
	data    []byte
	Offsets []uint32 // block start offsets into data; a block runs to the next offset
}

// LevelBlockSize is the fixed size of a level block (tile map + object lists).
const LevelBlockSize = 31752

// ParseArk reads the archive header: a uint16 block count, then that many uint32
// block offsets. Block i occupies [Offsets[i], Offsets[i+1]).
func ParseArk(data []byte) (*Ark, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("lev: short archive (%d bytes)", len(data))
	}
	n := int(binary.LittleEndian.Uint16(data))
	if 2+n*4 > len(data) {
		return nil, fmt.Errorf("lev: %d-entry offset table overruns %d bytes", n, len(data))
	}
	offs := make([]uint32, n)
	for i := range offs {
		offs[i] = binary.LittleEndian.Uint32(data[2+i*4:])
	}
	return &Ark{data: data, Offsets: offs}, nil
}

// Block returns the bytes of block i.
func (a *Ark) Block(i int) ([]byte, error) {
	if i < 0 || i >= len(a.Offsets) {
		return nil, fmt.Errorf("lev: block %d out of range (%d blocks)", i, len(a.Offsets))
	}
	start := a.Offsets[i]
	end := uint32(len(a.data))
	if i+1 < len(a.Offsets) && a.Offsets[i+1] > start {
		end = a.Offsets[i+1]
	}
	if start > end || end > uint32(len(a.data)) {
		return nil, fmt.Errorf("lev: block %d span %#x..%#x outside file", i, start, end)
	}
	return a.data[start:end], nil
}

// Tile-type values (word0 bits 0-3). Types 2-5 are the four diagonal
// (half-solid) corners; 6-9 are the four sloping floors.
const (
	TileSolid  = 0 // fully solid rock — no floor
	TileOpen   = 1 // open, flat floor
	TileDiagSE = 2 // diagonal walls (solid on one corner)
	TileDiagSW = 3
	TileDiagNE = 4
	TileDiagNW = 5
	TileSlopeN = 6 // sloping floors rising toward the named side
	TileSlopeS = 7
	TileSlopeE = 8
	TileSlopeW = 9
)

// Tile is one decoded map cell.
type Tile struct {
	Type    uint8  // bits 0-3 of word0 (see Tile* constants)
	Height  uint8  // bits 4-7 of word0 — floor height, 0-15
	Word0Hi uint16 // bits 8-15 of word0 (floor/wall texture indices — split TBD)
	Flags   uint8  // bits 0-5 of word1
	Object  uint16 // bits 6-15 of word1 — index of the first object in this tile (0 = none)
}

// Grid is a level's WidthxHeight tile map.
type Grid struct {
	W, H  int
	Tiles []Tile // row-major, len W*H
}

const (
	MapW = 64
	MapH = 64
)

// DecodeGrid reads the 64x64 tile map at the start of a level block. Each tile
// is two little-endian words: word0 = geometry (type, height, texture), word1 =
// flags + object-list head.
func DecodeGrid(block []byte) (*Grid, error) {
	need := MapW * MapH * 4
	if len(block) < need {
		return nil, fmt.Errorf("lev: block too small for tile map (%d < %d)", len(block), need)
	}
	g := &Grid{W: MapW, H: MapH, Tiles: make([]Tile, MapW*MapH)}
	for i := range g.Tiles {
		w0 := binary.LittleEndian.Uint16(block[i*4:])
		w1 := binary.LittleEndian.Uint16(block[i*4+2:])
		g.Tiles[i] = Tile{
			Type:    uint8(w0 & 0x0F),
			Height:  uint8((w0 >> 4) & 0x0F),
			Word0Hi: w0 >> 8,
			Flags:   uint8(w1 & 0x3F),
			Object:  (w1 >> 6) & 0x3FF,
		}
	}
	return g, nil
}

// At returns the tile at (x, y).
func (g *Grid) At(x, y int) Tile { return g.Tiles[y*g.W+x] }

// TypeGlyph maps a tile type to a single character for an ASCII dump.
func TypeGlyph(t uint8) byte {
	switch {
	case t == TileSolid:
		return '#'
	case t == TileOpen:
		return '.'
	case t >= TileDiagSE && t <= TileDiagNW:
		return '/'
	case t >= TileSlopeN && t <= TileSlopeW:
		return 's'
	default:
		return '?'
	}
}
