package lev

import "encoding/binary"

// Object placement, reverse-engineered from the object accessor 28B3:08E8 and
// the tile-object encoding. A level block stores objects in two arrays right
// after the 64x64x4 tile map:
//
//   - MOBILE objects: indices 0-255, 27 bytes each, at block offset 0x4000
//     (28B3: index<0x100 -> IMUL 0x1B, base [255C]).
//   - STATIC objects: indices 256-1023, 8 bytes each, right after the mobile
//     array (28B3: (index-0x100) SHL 3, base [2584]).
//
// Both begin with the same 8-byte "object info" header (4 little-endian words):
//
//	word0: bits 0-8   item id (which kind of thing)
//	       bits 9-15  flags / enchant / doordir (per-type)
//	word1: bits 0-6   Z (height within the level, 0-127)
//	       bits 7-9   heading (0-7, one of 8 facings)
//	       bits 10-12 fine Y within the tile (0-7)
//	       bits 13-15 fine X within the tile (0-7)
//	word2: bits 0-5   quality
//	       bits 6-15  link to the NEXT object in the same tile (0 = end)
//
// A tile's Object field (word1 bits 6-15) is the head of that tile's chain; the
// same bits in word2 of each object give the next, so a tile's objects are a
// singly linked list. The tile (X,Y) is implied by which chain the object is in;
// only the fine offset lives in the object.
const (
	objMobileBase = 0x4000          // mobile array start in a level block
	objStaticBase = 0x4000 + 256*27 // static array follows the 256 mobiles
	objMobileSize = 27
	objStaticSize = 8
)

// Object is one placed object with its tile resolved.
type Object struct {
	Index        int    // object-array index (its id in the chain)
	TileX, TileY int    // tile it belongs to
	ItemID       uint16 // word0 bits 0-8 — the kind of object
	FineX, FineY uint8  // 0-7 position within the tile
	Z            uint8  // 0-127 height
	Heading      uint8  // 0-7 facing
	Quality      uint8
	Flags        uint16 // word0 bits 9-15
	Next         int    // next object index in the tile (0 = end)
	Special      int    // word3 bits 6-15 — quantity / link / (>=512) string reference
	Mobile       bool   // mobile (27-byte) vs static (8-byte)
}

// objectHeader returns the 8-byte info offset for an object index, or ok=false.
func objectHeader(idx int) (off int, ok bool) {
	if idx <= 0 || idx >= 1024 {
		return 0, false
	}
	if idx < 256 {
		return objMobileBase + idx*objMobileSize, true
	}
	return objStaticBase + (idx-256)*objStaticSize, true
}

// DecodeObject reads one object's header from a level block.
func DecodeObject(block []byte, idx int) (Object, bool) {
	off, ok := objectHeader(idx)
	if !ok || off+8 > len(block) {
		return Object{}, false
	}
	w0 := binary.LittleEndian.Uint16(block[off:])
	w1 := binary.LittleEndian.Uint16(block[off+2:])
	w2 := binary.LittleEndian.Uint16(block[off+4:])
	w3 := binary.LittleEndian.Uint16(block[off+6:])
	return Object{
		Index:   idx,
		ItemID:  w0 & 0x1FF,
		Flags:   (w0 >> 9) & 0x7F,
		Z:       uint8(w1 & 0x7F),
		Heading: uint8((w1 >> 7) & 0x7),
		FineY:   uint8((w1 >> 10) & 0x7),
		FineX:   uint8((w1 >> 13) & 0x7),
		Quality: uint8(w2 & 0x3F),
		Next:    int((w2 >> 6) & 0x3FF),
		Special: int(w3 >> 6),
		Mobile:  idx < 256,
	}, true
}

// Objects walks every tile's object chain and returns all placed objects with
// their tile resolved. Chains are followed with a guard against cycles.
func Objects(g *Grid, block []byte) []Object {
	var out []Object
	for ty := 0; ty < g.H; ty++ {
		for tx := 0; tx < g.W; tx++ {
			idx := int(g.At(tx, ty).Object)
			for n := 0; idx != 0 && n < 1024; n++ {
				o, ok := DecodeObject(block, idx)
				if !ok {
					break
				}
				o.TileX, o.TileY = tx, ty
				out = append(out, o)
				idx = o.Next
			}
		}
	}
	return out
}
