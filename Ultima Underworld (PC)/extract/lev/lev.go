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

// Tile is one decoded map cell. The two words split as: word0 = geometry +
// floor texture, word1 = wall texture + object-list head. The texture fields are
// per-level indices (0-63) into the level's texture list.
type Tile struct {
	Type     uint8  // word0 bits 0-3  — geometry (see Tile* constants)
	Height   uint8  // word0 bits 4-7  — floor height, 0-15
	FloorTex uint8  // word0 bits 10-15 — floor texture index (6 bits)
	WallTex  uint8  // word1 bits 0-5  — wall texture index (6 bits)
	Object   uint16 // word1 bits 6-15 — first object in this tile (0 = none)
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
			Type:     uint8(w0 & 0x0F),
			Height:   uint8((w0 >> 4) & 0x0F),
			FloorTex: uint8((w0 >> 10) & 0x3F),
			WallTex:  uint8(w1 & 0x3F),
			Object:   (w1 >> 6) & 0x3FF,
		}
	}
	return g, nil
}

// At returns the tile at (x, y).
func (g *Grid) At(x, y int) Tile { return g.Tiles[y*g.W+x] }

// A level block's siblings in the archive: after the nine tile maps (0-8) come
// nine 384-byte animation blocks (9-17) and then the 122-byte per-level texture
// lists (18+). A texture list maps a tile's small texture index to a global
// number in W64.TR (walls) / F32.TR (floors).
const (
	texMapBase  = 18 // block index of level 0's texture list
	numWallTex  = 48 // wall entries in a texture list (into W64.TR)
	numFloorTex = 10 // floor entries (into F32.TR)
)

// TexMap is a level's texture list: the global W64.TR/F32.TR texture numbers its
// tiles' 6-bit/4-bit texture indices select.
type TexMap struct {
	Wall  []uint16 // WallTex index -> W64.TR texture number
	Floor []uint16 // FloorTex index -> F32.TR texture number
}

// TexMapForLevel decodes the texture list for level (0-based).
func (a *Ark) TexMapForLevel(level int) (*TexMap, error) {
	b, err := a.Block(texMapBase + level)
	if err != nil {
		return nil, err
	}
	if len(b) < (numWallTex+numFloorTex)*2 {
		return nil, fmt.Errorf("lev: texture list too small (%d bytes)", len(b))
	}
	tm := &TexMap{Wall: make([]uint16, numWallTex), Floor: make([]uint16, numFloorTex)}
	for i := range tm.Wall {
		tm.Wall[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	for i := range tm.Floor {
		tm.Floor[i] = binary.LittleEndian.Uint16(b[(numWallTex+i)*2:])
	}
	return tm, nil
}

// WallTexture maps a tile's WallTex index to a W64.TR texture number.
func (m *TexMap) WallTexture(idx uint8) uint16 {
	if int(idx) < len(m.Wall) {
		return m.Wall[idx]
	}
	return 0
}

// FloorTexture maps a tile's FloorTex index to an F32.TR texture number. Only
// the low 4 bits index the floor list.
func (m *TexMap) FloorTexture(idx uint8) uint16 {
	if int(idx&0x0F) < len(m.Floor) {
		return m.Floor[idx&0x0F]
	}
	return 0
}

// Geometry describes a tile type's shape. Types 2-5 (diagonal) are open on one
// triangular half with a diagonal wall across; 6-9 (slope) ramp the floor up by
// one height unit toward the named side. The four diagonals and the four slopes
// are rotations of one shape — the game rotates them with a remap table at
// DGROUP 034E (row = camera facing) so its geometry code only handles a
// canonical orientation. The slope surface itself is a per-subregion height
// table (DGROUP 038F) the floor-height query (2CD3:0720) walks.
func Geometry(t uint8) string {
	switch t {
	case TileSolid:
		return "solid rock (no floor, full walls)"
	case TileOpen:
		return "open flat floor"
	case TileDiagSE:
		return "diagonal wall, open SE half"
	case TileDiagSW:
		return "diagonal wall, open SW half"
	case TileDiagNE:
		return "diagonal wall, open NE half"
	case TileDiagNW:
		return "diagonal wall, open NW half"
	case TileSlopeN:
		return "floor sloping up toward north"
	case TileSlopeS:
		return "floor sloping up toward south"
	case TileSlopeE:
		return "floor sloping up toward east"
	case TileSlopeW:
		return "floor sloping up toward west"
	default:
		return "unknown"
	}
}

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
