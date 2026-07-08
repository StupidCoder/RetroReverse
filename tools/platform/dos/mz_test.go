package dos

import (
	"encoding/binary"
	"testing"
)

// synthMZ builds a minimal but well-formed MZ image: an 8-paragraph (128-byte)
// header, one relocation entry, a small load module, and some appended bytes.
func synthMZ() []byte {
	const (
		headerParas = 8 // 128-byte header
		headerBytes = headerParas * 16
		moduleBytes = 64 // load module payload
		appended    = 16 // trailing overlay-style data
	)
	total := headerBytes + moduleBytes + appended
	imageEnd := headerBytes + moduleBytes // header claims the load image ends here
	buf := make([]byte, total)

	put := func(off int, v uint16) { binary.LittleEndian.PutUint16(buf[off:], v) }
	buf[0], buf[1] = 'M', 'Z'
	// imageEnd = (pages-1)*512 + lastPage.  With imageEnd=192: pages=1, last=192.
	put(0x02, uint16(imageEnd)) // last page bytes (fits in one 512 page)
	put(0x04, 1)                // pages
	put(0x06, 1)                // relocations
	put(0x08, headerParas)      // header paragraphs
	put(0x0E, 0x0002)           // SS
	put(0x10, 0x0040)           // SP
	put(0x14, 0x0010)           // IP
	put(0x16, 0x0001)           // CS
	put(0x18, 0x001C)           // reloc table offset
	// one relocation entry {offset, segment} at 0x1C
	put(0x1C, 0x0005) // offset
	put(0x1E, 0x0003) // segment
	return buf
}

func TestParseMZLayout(t *testing.T) {
	m, err := ParseMZ(synthMZ())
	if err != nil {
		t.Fatal(err)
	}
	if m.HeaderParas != 8 {
		t.Errorf("HeaderParas = %d, want 8", m.HeaderParas)
	}
	if m.LoadModuleOffset != 128 {
		t.Errorf("LoadModuleOffset = %d, want 128", m.LoadModuleOffset)
	}
	if m.LoadImageEnd != 192 {
		t.Errorf("LoadImageEnd = %d, want 192", m.LoadImageEnd)
	}
	if m.LoadModuleSize != 64 {
		t.Errorf("LoadModuleSize = %d, want 64", m.LoadModuleSize)
	}
	if m.AppendedSize != 16 {
		t.Errorf("AppendedSize = %d, want 16", m.AppendedSize)
	}
	if got := m.EntryLinear(); got != 0x0001*16+0x0010 {
		t.Errorf("EntryLinear = %#x, want %#x", got, 0x0001*16+0x0010)
	}
	if got := m.StackLinear(); got != 0x0002*16+0x0040 {
		t.Errorf("StackLinear = %#x, want %#x", got, 0x0002*16+0x0040)
	}
	if len(m.Relocs) != 1 {
		t.Fatalf("Relocs = %d, want 1", len(m.Relocs))
	}
	if r := m.Relocs[0]; r.Segment != 3 || r.Offset != 5 || r.FarAddr() != 53 {
		t.Errorf("Relocs[0] = %+v (far %d), want seg 3 off 5 far 53", r, r.FarAddr())
	}
}

func TestParseMZBadSignature(t *testing.T) {
	if _, err := ParseMZ([]byte("XX\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")); err == nil {
		t.Error("expected error on bad signature")
	}
}
