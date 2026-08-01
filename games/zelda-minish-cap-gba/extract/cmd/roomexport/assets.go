package main

// The game's asset-list format — game-specific, so it lives here rather than in
// tools/platform/gba (which stays platform-generic).
//
// The loader at $080197D4 walks a list of 12-byte records:
//
//	+0  source, as an offset RELATIVE to $08324AE4, with bit 31 = "another
//	    record follows". This is why searching the ROM for an absolute pointer
//	    to a compressed stream finds nothing at all.
//	+4  destination address
//	+8  size, with bit 31 = "the stream is compressed" (BIOS LZ77)
//
// It picks SWI $12 over SWI $11 when the destination is in VRAM, because a
// byte store to VRAM does not store a byte (see gbamachine/bus.go).
//
// A room needs TWO lists: one carries its tilesets and its metatile-id grids,
// the other the metatile tables — which are shared between areas, so several
// rooms point at the same one.

import (
	"encoding/binary"
	"fmt"
)

// assetBase is the value the loader adds to every record's source field,
// read from its literal pool at $0801980C.
const assetBase = 0x08324AE4

// AssetEntry is one record of a list.
type AssetEntry struct {
	Src        uint32
	Dst        uint32
	Size       int
	Compressed bool
	more       bool
}

// romAt returns the ROM bytes at a cartridge address.
func romAt(rom []byte, addr uint32) []byte {
	off := int(addr & 0x01FFFFFF)
	if off < 0 || off >= len(rom) {
		return nil
	}
	return rom[off:]
}

// parseAssetEntry decodes one record, rejecting anything that cannot be one.
// The validation is what makes a structural scan of the ROM possible.
func parseAssetEntry(rom []byte, addr uint32) (AssetEntry, bool) {
	b := romAt(rom, addr)
	if len(b) < 12 {
		return AssetEntry{}, false
	}
	w0 := binary.LittleEndian.Uint32(b[0:])
	w1 := binary.LittleEndian.Uint32(b[4:])
	w2 := binary.LittleEndian.Uint32(b[8:])
	e := AssetEntry{
		Src:        (w0 & 0x7FFFFFFF) + assetBase,
		Dst:        w1,
		Size:       int(w2 & 0x7FFFFFFF),
		Compressed: w2&0x80000000 != 0,
		more:       w0&0x80000000 != 0,
	}
	if int(e.Src&0x01FFFFFF) >= len(rom) {
		return AssetEntry{}, false
	}
	// A destination of 0 is not a load at all: the loader branches at $08019824
	// and treats the record as a command. Rejecting it as malformed breaks the
	// walk of any list that contains one — which our own starting area does.
	if e.Dst == 0 {
		return e, true
	}
	switch e.Dst >> 24 {
	case 2, 3, 6: // EWRAM, IWRAM, VRAM — the only places assets land
	default:
		return AssetEntry{}, false
	}
	if e.Size == 0 || e.Size > 0x20000 {
		return AssetEntry{}, false
	}
	return e, true
}

// ParseAssetList walks a list from its first record.
func ParseAssetList(rom []byte, addr uint32) ([]AssetEntry, error) {
	var out []AssetEntry
	for i := 0; i < 32; i++ {
		e, ok := parseAssetEntry(rom, addr)
		if !ok {
			return nil, fmt.Errorf("record %d at %08X is not an asset entry", i, addr)
		}
		out = append(out, e)
		if !e.more {
			return out, nil
		}
		addr += 12
	}
	return nil, fmt.Errorf("asset list at %08X does not terminate", addr)
}

// RoomAssets is what a room's two lists resolve to.
type RoomAssets struct {
	Tilesets [3]uint32 // compressed streams for VRAM 0x00000 / 0x04000 / 0x08000
	IDs      [2]uint32 // metatile-id grids: BG2 then BG1
	Tables   [2]uint32 // metatile tables: BG2 then BG1
	IDSize   int
}

// Resolve classifies the entries of a room list and a metatile-set list.
// Classification is by DESTINATION, which is the only thing that says what a
// stream is for: the id grids go to two fixed RAM buffers, the tables to two
// others, and the tilesets to their VRAM character bases.
// tilesetList supplies any character bank the room's own list does not carry:
// a room list is often an INCREMENTAL load that replaces only the banks that
// changed since the last area, so taking it as the whole story leaves a bank
// zeroed and renders that half of the tiles as blanks.
func Resolve(roomList, metaList, tilesetList []AssetEntry) (RoomAssets, error) {
	var r RoomAssets
	for _, e := range append(append([]AssetEntry{}, tilesetList...), roomList...) {
		switch {
		case e.Dst == 0x06000000:
			r.Tilesets[0] = e.Src
		case e.Dst == 0x06004000:
			r.Tilesets[1] = e.Src
		case e.Dst == 0x06008000:
			r.Tilesets[2] = e.Src
		case e.Dst == 0x02025EB4: // BG2 metatile-id grid
			r.IDs[0], r.IDSize = e.Src, e.Size
		case e.Dst == 0x0200B654: // BG1 metatile-id grid
			r.IDs[1] = e.Src
		}
	}
	for _, e := range metaList {
		switch e.Dst {
		case 0x0202CEB4:
			r.Tables[0] = e.Src
		case 0x02012654:
			r.Tables[1] = e.Src
		}
	}
	if r.IDs[0] == 0 || r.Tables[0] == 0 {
		return r, fmt.Errorf("the lists do not resolve to a room (ids %08X, table %08X)", r.IDs[0], r.Tables[0])
	}
	return r, nil
}

// ScanRoomLists finds every asset list in the ROM that loads a metatile-id grid
// — that is, every list that describes a room's map. A list is recognised by
// its own record format rather than by finding an index that points at it, so
// this enumerates rooms without having to reverse the whole table hierarchy.
func ScanRoomLists(rom []byte) []uint32 {
	var out []uint32
	seen := map[uint32]bool{}
	for off := 0; off+12 <= len(rom); off += 4 {
		addr := uint32(0x08000000 + off)
		if _, ok := parseAssetEntry(rom, addr); !ok {
			continue
		}
		// Only consider a record that starts a list: the previous 12 bytes must
		// not be a record that continues into it.
		if off >= 12 {
			if p, ok := parseAssetEntry(rom, addr-12); ok && p.more {
				continue
			}
		}
		list, err := ParseAssetList(rom, addr)
		if err != nil || len(list) < 2 {
			continue
		}
		for _, x := range list {
			if x.Dst == 0x02025EB4 && !seen[addr] {
				seen[addr] = true
				out = append(out, addr)
				break
			}
		}
	}
	return out
}
