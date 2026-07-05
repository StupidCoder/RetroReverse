// MZ container parsing. An "MZ" executable is a plain 16-bit real-mode DOS
// program (no protected-mode extender). This file parses the header so a caller
// can report the load layout, entry point and relocation table, and so the x86
// oracle can place the load module at a chosen segment and apply the fixups.
package dos

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// MZ is the parsed header of a real-mode DOS executable plus a view of where
// its load module and any appended overlay data sit in the file.
type MZ struct {
	// Raw header fields (all little-endian, as stored at file offset 0).
	LastPageBytes uint16 // e_cblp: bytes used on the final 512-byte page (0 => full page)
	Pages         uint16 // e_cp:   number of 512-byte pages, rounded up
	Relocations   uint16 // e_crlc: number of relocation entries
	HeaderParas   uint16 // e_cparhdr: header size in 16-byte paragraphs
	MinAlloc      uint16 // e_minalloc: minimum extra paragraphs needed
	MaxAlloc      uint16 // e_maxalloc: maximum extra paragraphs requested
	InitSS        uint16 // e_ss: initial stack segment, relative to load base
	InitSP        uint16 // e_sp: initial stack pointer
	Checksum      uint16 // e_csum: header checksum (usually 0, unused by DOS)
	InitIP        uint16 // e_ip: initial instruction pointer
	InitCS        uint16 // e_cs: initial code segment, relative to load base
	RelocOffset   uint16 // e_lfarlc: file offset of the relocation table
	OverlayNumber uint16 // e_ovno: overlay number (0 = main program)

	// Derived layout, in bytes from the start of the file.
	FileSize         int // total size of the image on disk
	LoadImageEnd     int // last byte of the load image as the header describes it
	LoadModuleOffset int // first byte of the load module (past the MZ header)
	LoadModuleSize   int // bytes of code+data DOS copies into memory
	AppendedSize     int // bytes past LoadImageEnd (UW's swapped overlay segments live here)

	// Relocs are the fixups DOS adds the load segment to at load time.
	Relocs []Reloc
}

// Reloc is one relocation entry: a far pointer (segment:offset) into the load
// module whose word DOS patches by adding the runtime load segment.
type Reloc struct {
	Segment uint16
	Offset  uint16
}

// FarAddr returns the byte offset of the reloc target within the load module.
func (r Reloc) FarAddr() int { return int(r.Segment)*16 + int(r.Offset) }

// ParseMZ decodes the MZ header and relocation table from a whole-file image.
func ParseMZ(data []byte) (*MZ, error) {
	if len(data) < 0x1C {
		return nil, errors.New("mz: file too small for a DOS header")
	}
	if !(data[0] == 'M' && data[1] == 'Z') && !(data[0] == 'Z' && data[1] == 'M') {
		return nil, fmt.Errorf("mz: bad signature %q, want \"MZ\"", string(data[0:2]))
	}
	u16 := func(off int) uint16 { return binary.LittleEndian.Uint16(data[off:]) }

	m := &MZ{
		LastPageBytes: u16(0x02),
		Pages:         u16(0x04),
		Relocations:   u16(0x06),
		HeaderParas:   u16(0x08),
		MinAlloc:      u16(0x0A),
		MaxAlloc:      u16(0x0C),
		InitSS:        u16(0x0E),
		InitSP:        u16(0x10),
		Checksum:      u16(0x12),
		InitIP:        u16(0x14),
		InitCS:        u16(0x16),
		RelocOffset:   u16(0x18),
		OverlayNumber: u16(0x1A),
		FileSize:      len(data),
	}

	// The header's own record of the load image size. When LastPageBytes is 0
	// the final page is full; otherwise only that many bytes of it are used.
	if m.Pages > 0 {
		if m.LastPageBytes == 0 {
			m.LoadImageEnd = int(m.Pages) * 512
		} else {
			m.LoadImageEnd = (int(m.Pages)-1)*512 + int(m.LastPageBytes)
		}
	}
	m.LoadModuleOffset = int(m.HeaderParas) * 16
	m.LoadModuleSize = m.LoadImageEnd - m.LoadModuleOffset
	m.AppendedSize = m.FileSize - m.LoadImageEnd

	// Relocation table: e_crlc entries of {offset u16, segment u16}.
	base := int(m.RelocOffset)
	for i := 0; i < int(m.Relocations); i++ {
		off := base + i*4
		if off+4 > len(data) {
			return nil, fmt.Errorf("mz: relocation %d at %#x runs past end of file", i, off)
		}
		m.Relocs = append(m.Relocs, Reloc{
			Offset:  binary.LittleEndian.Uint16(data[off:]),
			Segment: binary.LittleEndian.Uint16(data[off+2:]),
		})
	}
	return m, nil
}

// EntryLinear returns the entry point CS:IP as a byte offset within the load
// module (CS is relative to the load base, so this is a module-local address).
func (m *MZ) EntryLinear() int { return int(m.InitCS)*16 + int(m.InitIP) }

// StackLinear returns the initial SS:SP as a byte offset within the load module.
func (m *MZ) StackLinear() int { return int(m.InitSS)*16 + int(m.InitSP) }
