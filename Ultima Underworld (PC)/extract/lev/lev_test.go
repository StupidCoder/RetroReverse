package lev

import (
	"encoding/binary"
	"testing"
)

// synthArk builds a tiny archive: 2 blocks, the first a level block whose first
// tile is open/height-3/object-5 and second tile is solid.
func synthArk() []byte {
	block := make([]byte, LevelBlockSize)
	// tile 0: type=1 (open), height=3, floorTex=0x2A, wallTex=0x15, object=5
	w0 := uint16(TileOpen) | uint16(3)<<4 | uint16(0x2A)<<10
	w1 := uint16(0x15) | uint16(5)<<6
	binary.LittleEndian.PutUint16(block[0:], w0)
	binary.LittleEndian.PutUint16(block[2:], w1)
	// tile 1: solid (all zero) — leave as is.

	var buf []byte
	buf = binary.LittleEndian.AppendUint16(buf, 2) // 2 blocks
	// offsets table (2 entries) then the two blocks
	tableEnd := 2 + 2*4
	buf = binary.LittleEndian.AppendUint32(buf, uint32(tableEnd))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(tableEnd+len(block)))
	buf = append(buf, block...)
	buf = append(buf, 0xAB) // a one-byte trailing block
	return buf
}

func TestParseArkAndTile(t *testing.T) {
	ark, err := ParseArk(synthArk())
	if err != nil {
		t.Fatal(err)
	}
	if len(ark.Offsets) != 2 {
		t.Fatalf("blocks = %d, want 2", len(ark.Offsets))
	}
	b0, err := ark.Block(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(b0) != LevelBlockSize {
		t.Fatalf("block0 = %d bytes, want %d", len(b0), LevelBlockSize)
	}
	g, err := DecodeGrid(b0)
	if err != nil {
		t.Fatal(err)
	}
	t0 := g.At(0, 0)
	if t0.Type != TileOpen || t0.Height != 3 || t0.Object != 5 ||
		t0.FloorTex != 0x2A || t0.WallTex != 0x15 {
		t.Errorf("tile0 = %+v, want open/h3/obj5/ftex2A/wtex15", t0)
	}
	if t1 := g.At(1, 0); t1.Type != TileSolid || t1.Object != 0 {
		t.Errorf("tile1 = %+v, want solid/no-object", t1)
	}
}
