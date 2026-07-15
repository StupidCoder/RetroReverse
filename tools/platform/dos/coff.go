// go32/COFF container parsing. A DJGPP "go32" DOS executable is a small real-
// mode MZ *stub* (the STUB.EXE loader that finds/loads a DPMI host such as
// CWSDPMI.EXE) with an i386 COFF image appended right after it. Under DOS the
// stub switches to 32-bit protected mode via DPMI and runs the COFF payload
// flat; our oracle instead HLE's the DPMI host and starts the COFF directly in
// flat protected mode. This file parses the appended COFF so a loader can place
// its sections and find the entry point, and so the disassembler can be pointed
// at the .text bytes in 32-bit mode.
//
// The image was first proven on Quake shareware (quake.exe, go32stub v2.00T).
// Nothing here is Quake-specific.
package dos

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// COFF magic numbers.
const (
	coffI386Magic = 0x014C // filehdr f_magic: Intel 386
	coffZMagic    = 0x010B // aouthdr magic: demand-paged (ZMAGIC), what go32 emits
)

// COFF section-header flags (s_flags) we care about.
const (
	coffStypText = 0x0020 // executable code
	coffStypData = 0x0040 // initialised data
	coffStypBss  = 0x0080 // uninitialised data (occupies no file bytes)
)

// COFFSection is one parsed COFF section.
type COFFSection struct {
	Name    string
	VAddr   uint32 // runtime virtual address (flat, zero-based in the go32 model)
	Size    uint32 // size in bytes
	FileOff uint32 // offset of the section's raw bytes within the whole go32 file (0 for .bss)
	Flags   uint32
	Data    []byte // raw section bytes (nil for .bss)
}

// IsText reports whether the section holds executable code.
func (s COFFSection) IsText() bool { return s.Flags&coffStypText != 0 }

// IsBSS reports whether the section is uninitialised (no file bytes).
func (s COFFSection) IsBSS() bool { return s.Flags&coffStypBss != 0 }

// COFF is a parsed go32 COFF image.
type COFF struct {
	StubEnd   int    // file offset where the MZ stub ends and the COFF begins
	NSections int    // number of sections
	Flags     uint16 // filehdr f_flags
	Entry     uint32 // program entry point (virtual address)
	TextStart uint32 // aouthdr text_start
	DataStart uint32 // aouthdr data_start
	TextSize  uint32 // aouthdr tsize
	DataSize  uint32 // aouthdr dsize
	BSSSize   uint32 // aouthdr bsize
	Sections  []COFFSection
}

// ParseGo32COFF locates the appended COFF image inside a whole go32 executable
// file and decodes its headers and sections. The COFF begins immediately after
// the MZ stub's load image (the stub's own e_cp/e_cblp describe its size).
func ParseGo32COFF(data []byte) (*COFF, error) {
	mz, err := ParseMZ(data)
	if err != nil {
		return nil, fmt.Errorf("go32: parsing the MZ stub: %w", err)
	}
	off := mz.LoadImageEnd
	if off <= 0 || off+20 > len(data) {
		return nil, fmt.Errorf("go32: stub image end %#x leaves no room for a COFF header in a %d-byte file", off, len(data))
	}

	c := &COFF{StubEnd: off}
	u16 := func(o int) uint16 { return binary.LittleEndian.Uint16(data[o:]) }
	u32 := func(o int) uint32 { return binary.LittleEndian.Uint32(data[o:]) }

	// COFF file header (filehdr, 20 bytes).
	if magic := u16(off); magic != coffI386Magic {
		return nil, fmt.Errorf("go32: at file offset %#x: COFF magic %#04x, want %#04x (i386)", off, magic, coffI386Magic)
	}
	c.NSections = int(u16(off + 2))
	optSize := int(u16(off + 16))
	c.Flags = u16(off + 18)

	// Optional (a.out) header, if present: magic, vstamp, tsize, dsize, bsize,
	// entry, text_start, data_start.
	optOff := off + 20
	if optSize >= 28 {
		if m := u16(optOff); m != coffZMagic {
			return nil, fmt.Errorf("go32: aouthdr magic %#04x, want %#04x (ZMAGIC)", m, coffZMagic)
		}
		c.TextSize = u32(optOff + 4)
		c.DataSize = u32(optOff + 8)
		c.BSSSize = u32(optOff + 12)
		c.Entry = u32(optOff + 16)
		c.TextStart = u32(optOff + 20)
		c.DataStart = u32(optOff + 24)
	}

	// Section headers (scnhdr, 40 bytes each) follow the optional header.
	shOff := optOff + optSize
	for i := 0; i < c.NSections; i++ {
		base := shOff + i*40
		if base+40 > len(data) {
			return nil, fmt.Errorf("go32: section header %d at %#x runs past end of file", i, base)
		}
		name := strings.TrimRight(string(data[base:base+8]), "\x00")
		s := COFFSection{
			Name:    name,
			VAddr:   u32(base + 12),
			Size:    u32(base + 16),
			FileOff: u32(base + 20),
			Flags:   u32(base + 36),
		}
		if !s.IsBSS() && s.FileOff != 0 && s.Size != 0 {
			// The section's raw bytes are at StubEnd+FileOff? No: COFF scnptr is a
			// file offset relative to the start of the COFF image, so add StubEnd.
			start := c.StubEnd + int(s.FileOff)
			end := start + int(s.Size)
			if start < 0 || end > len(data) {
				return nil, fmt.Errorf("go32: section %q bytes %#x..%#x outside file (%d bytes)", name, start, end, len(data))
			}
			s.Data = data[start:end]
		}
		c.Sections = append(c.Sections, s)
	}
	if c.Entry == 0 && len(c.Sections) == 0 {
		return nil, errors.New("go32: no optional header and no sections — not a go32 COFF image")
	}
	return c, nil
}

// TextSection returns the first executable section (typically ".text"), or nil.
func (c *COFF) TextSection() *COFFSection {
	for i := range c.Sections {
		if c.Sections[i].IsText() {
			return &c.Sections[i]
		}
	}
	return nil
}

// SectionAt returns the section whose virtual-address range contains vaddr.
func (c *COFF) SectionAt(vaddr uint32) *COFFSection {
	for i := range c.Sections {
		s := &c.Sections[i]
		if vaddr >= s.VAddr && vaddr < s.VAddr+s.Size {
			return s
		}
	}
	return nil
}
