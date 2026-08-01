package main

// The game's area/room model, read out of its own tables.
//
// A room is not a standalone thing: it belongs to an AREA, and the area's
// tables say where each of its rooms sits on one shared pixel canvas. That is
// what turns 600 room maps into a few dozen assembled places.
//
// Five index tables, all indexed by area id (the descriptor builder at
// $08053020 reads them with r6 = area*4):
//
//	$0811E214  geometry  -> a table of 10-byte room records, 0xFFFF-terminated
//	$08107988  roomlists -> an array of per-room asset lists (the id grids)
//	$0810246C  tsetlists -> an array of tileset lists, chosen per room
//	$0810309C  metalist  -> the area's metatile-table list (shared by its rooms)
//	$080B755C  script    -> the area's code entry point (not decoded here)
//
// A room record is 10 bytes: {x, y, w, h, tilesetIndex}, x/y/w/h in PIXELS.
// The width the expander needs is w/16, because a metatile is 2x2 tiles of 8
// pixels. The starting area is area 3, room 1: (1488,2480) 1008x688 — the
// same 63x43 metatiles derived independently from the compressed data.
//
// Palettes hang off the tileset list's COMMAND record (destination 0, which the
// loader routes to $0801D714 instead of copying): its word carries a palette
// set id. A set indexes $080FF850 to a list of 4-byte records
// {srcIndex, destSlot, countAndFlag}, each copying count 16-colour palettes
// from $085A2E80 + srcIndex*32 into palette slot destSlot.

import (
	"encoding/binary"
	"fmt"
)

const (
	idxGeometry  = 0x0811E214
	idxRoomLists = 0x08107988
	idxTsetLists = 0x0810246C
	idxMetaList  = 0x0810309C

	palGroupTable = 0x080FF850
	palDataBase   = 0x085A2E80
)

func rd32(rom []byte, addr uint32) uint32 {
	o := int(addr & 0x01FFFFFF)
	if o+4 > len(rom) {
		return 0
	}
	return binary.LittleEndian.Uint32(rom[o:])
}

func rd16(rom []byte, addr uint32) uint16 {
	o := int(addr & 0x01FFFFFF)
	if o+2 > len(rom) {
		return 0
	}
	return binary.LittleEndian.Uint16(rom[o:])
}

func rd8(rom []byte, addr uint32) byte {
	o := int(addr & 0x01FFFFFF)
	if o >= len(rom) {
		return 0
	}
	return rom[o]
}

// RoomRec is one 10-byte room record.
type RoomRec struct {
	X, Y, W, H uint16 // pixels
	TilesetIdx uint16
}

// WidthMeta and HeightMeta convert the pixel size to the metatile grid the
// compressed id array is packed at.
func (r RoomRec) WidthMeta() int  { return int(r.W) / 16 }
func (r RoomRec) HeightMeta() int { return int(r.H) / 16 }

// Area is one area's tables, resolved.
type Area struct {
	ID        int
	Rooms     []RoomRec
	RoomLists []uint32 // asset list per room
	MetaList  uint32
	TsetLists []uint32 // tileset list per tileset index
}

// LoadArea resolves area a from the index tables. Rooms are read until the
// geometry table's 0xFFFF terminator.
func LoadArea(rom []byte, a int) (*Area, error) {
	ar := &Area{ID: a}
	geo := rd32(rom, uint32(idxGeometry+a*4))
	rl := rd32(rom, uint32(idxRoomLists+a*4))
	tl := rd32(rom, uint32(idxTsetLists+a*4))
	ar.MetaList = rd32(rom, uint32(idxMetaList+a*4))
	if geo == 0 || rl == 0 || geo>>24 != 0x08 || rl>>24 != 0x08 {
		return nil, fmt.Errorf("area %d has no tables", a)
	}
	for i := 0; i < 64; i++ {
		p := geo + uint32(i*10)
		r := RoomRec{rd16(rom, p), rd16(rom, p+2), rd16(rom, p+4), rd16(rom, p+6), rd16(rom, p+8)}
		if r.X == 0xFFFF || r.Y == 0xFFFF || r.W == 0 || r.H == 0 || r.W > 4096 || r.H > 4096 {
			break
		}
		ar.Rooms = append(ar.Rooms, r)
		ar.RoomLists = append(ar.RoomLists, rd32(rom, rl+uint32(i*4)))
	}
	if len(ar.Rooms) == 0 {
		return nil, fmt.Errorf("area %d has no rooms", a)
	}
	maxT := 0
	for _, r := range ar.Rooms {
		if int(r.TilesetIdx) > maxT {
			maxT = int(r.TilesetIdx)
		}
	}
	if tl != 0 && tl>>24 == 0x08 {
		for i := 0; i <= maxT; i++ {
			ar.TsetLists = append(ar.TsetLists, rd32(rom, tl+uint32(i*4)))
		}
	}
	return ar, nil
}

// PaletteSetOf returns the palette set id a tileset list carries in its command
// record (the entry whose destination is 0), or -1 if it has none.
func PaletteSetOf(rom []byte, list []AssetEntry) int {
	for _, e := range list {
		if e.Dst == 0 {
			return int(e.Src - assetBase)
		}
	}
	return -1
}

// baseSets are the palette sets the engine loads BEFORE an area's own: the
// common banks (0, 1 and 15) that every overworld room shares. An area set
// covers only the slots it actually replaces — set 0x10 fills 2..14 — so
// rebuilding from it alone leaves those common banks black, which is exactly
// how the missing colours showed up. The pair was read off the running game
// rather than guessed: with them applied first, the rebuilt table matches the
// palette the console holds in all but two entries, and both of those are the
// index-0 slot of a bank, which is the transparent/backdrop entry and is never
// drawn.
var baseSets = []int{0x0B, 0x0C}

// BuildPalette assembles the 512-entry palette a palette set produces, as
// palette RAM would hold it, over the common base sets.
func BuildPalette(rom []byte, set int) []uint16 {
	pal := make([]uint16, 512)
	for _, b := range baseSets {
		applyPaletteSet(rom, pal, b)
	}
	applyPaletteSet(rom, pal, set)
	return pal
}

// applyPaletteSet runs one set's copy records over pal.
func applyPaletteSet(rom []byte, pal []uint16, set int) {
	if set < 0 {
		return
	}
	rec := rd32(rom, uint32(palGroupTable+set*4))
	if rec == 0 || rec>>24 != 0x08 {
		return
	}
	for i := 0; i < 64; i++ {
		p := rec + uint32(i*4)
		srcIdx := rd16(rom, p)
		slot := rd8(rom, p+2)
		ctl := rd8(rom, p+3)
		n := int(ctl & 0xF)
		if n == 0 {
			n = 16
		}
		src := uint32(palDataBase) + uint32(srcIdx)*32
		for c := 0; c < n*16; c++ {
			d := int(slot)*16 + c
			if d >= len(pal) {
				break
			}
			pal[d] = rd16(rom, src+uint32(c)*2)
		}
		if ctl&0x80 == 0 {
			return
		}
	}
}
