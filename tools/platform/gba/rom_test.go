package gba

import (
	"encoding/binary"
	"testing"
)

// synthHeader builds a minimal valid image: B +0xC0 entry, a title/code block,
// and a correct complement check.
func synthHeader() []byte {
	d := make([]byte, 0x100)
	// B 0xC0: from PC=0x08 the word offset is (0xC0-0x08)/4 = 0x2E.
	binary.LittleEndian.PutUint32(d[0x00:], 0xEA00002E)
	copy(d[0xA0:], "ZELDA MC")
	copy(d[0xAC:], "BZME")
	copy(d[0xB0:], "01")
	d[0xB2] = 0x96
	d[0xBD] = ComplementCheck(d)
	return d
}

func TestParseHeader(t *testing.T) {
	r, err := Parse(synthHeader())
	if err != nil {
		t.Fatal(err)
	}
	h := r.Header
	if h.Title != "ZELDA MC" || h.GameCode != "BZME" || h.MakerCode != "01" {
		t.Errorf("title/code/maker = %q %q %q", h.Title, h.GameCode, h.MakerCode)
	}
	if h.EntryAddr != 0x080000C0 {
		t.Errorf("entry = %#x, want 0x080000C0", h.EntryAddr)
	}
	if !h.ChecksumOK {
		t.Errorf("checksum %#02x rejected, ComplementCheck = %#02x", h.Checksum, ComplementCheck(r.Data))
	}
}

func TestChecksumCatchesEdit(t *testing.T) {
	d := synthHeader()
	d[0xA0] ^= 0xFF
	if r, _ := Parse(d); r.Header.ChecksumOK {
		t.Error("checksum accepted a corrupted title")
	}
}

func TestSaveTypeScan(t *testing.T) {
	d := append(synthHeader(), []byte("...EEPROM_V124...")...)
	r, _ := Parse(d)
	id, _ := r.SaveType()
	if id != "EEPROM_V124" {
		t.Errorf("save id = %q, want EEPROM_V124", id)
	}
}
