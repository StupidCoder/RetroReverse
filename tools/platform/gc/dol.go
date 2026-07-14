package gc

// dol.go reads the game's executable.
//
// A DOL is the plainest executable format in this repository. There is no symbol
// table, no relocation, no dynamic linking, no compression, and no metadata beyond a
// 0x100-byte header of three parallel arrays — file offset, load address, and size —
// for a fixed eighteen segments: seven of code and eleven of data. The loader copies
// each non-empty segment to the address the header names, zeroes the BSS, and jumps to
// the entry point. That is the whole of the format.
//
// The absence of a symbol table is worth stating plainly, because it shapes every tool
// downstream. The PS2's boot ELF ships .symtab intact and so its oracle can set a
// breakpoint on a name; here there are no names at all, only addresses, and the tools
// must say so rather than pretend otherwise.

import (
	"errors"
	"fmt"
)

const (
	dolHeaderSize = 0x100
	dolTextSegs   = 7
	dolDataSegs   = 11
	dolSegs       = dolTextSegs + dolDataSegs
)

// Segment is one loadable piece of the executable.
type Segment struct {
	Index  int    // 0..6 are text, 7..17 are data
	Text   bool
	Offset uint32 // byte offset within the DOL
	Addr   uint32 // where it loads
	Size   uint32
	Data   []byte
}

func (s Segment) Name() string {
	if s.Text {
		return fmt.Sprintf("text%d", s.Index)
	}
	return fmt.Sprintf("data%d", s.Index-dolTextSegs)
}

// DOL is a parsed executable.
type DOL struct {
	Entry    uint32
	BSSAddr  uint32
	BSSSize  uint32
	Segments []Segment // the non-empty ones only, in header order
}

// dolLength is how many bytes of the file the segment table actually reaches. The disc
// header says where the executable starts but not how long it is, so its own table has
// to answer that.
func dolLength(hdr []byte) int {
	n := uint32(dolHeaderSize)
	for i := 0; i < dolSegs; i++ {
		off := be32(hdr[i*4:])
		size := be32(hdr[0x90+i*4:])
		if size == 0 {
			continue
		}
		if end := off + size; end > n {
			n = end
		}
	}
	return int(n)
}

// ParseDOL reads an executable. b must be the whole file.
func ParseDOL(b []byte) (*DOL, error) {
	if len(b) < dolHeaderSize {
		return nil, errors.New("gc: the executable is shorter than its own header")
	}
	d := &DOL{
		BSSAddr: be32(b[0xD8:]),
		BSSSize: be32(b[0xDC:]),
		Entry:   be32(b[0xE0:]),
	}
	for i := 0; i < dolSegs; i++ {
		off := be32(b[i*4:])
		addr := be32(b[0x48+i*4:])
		size := be32(b[0x90+i*4:])
		if size == 0 {
			continue // an unused slot; the header keeps all eighteen whether or not they are filled
		}
		if int(off)+int(size) > len(b) {
			return nil, fmt.Errorf("gc: segment %d spans %#x+%#x, past the %d-byte executable", i, off, size, len(b))
		}
		d.Segments = append(d.Segments, Segment{
			Index:  i,
			Text:   i < dolTextSegs,
			Offset: off,
			Addr:   addr,
			Size:   size,
			Data:   b[off : off+size],
		})
	}
	if len(d.Segments) == 0 {
		return nil, errors.New("gc: the executable has no segments")
	}
	return d, nil
}

// Load hands every segment to a writer, then the BSS as zeroes — the whole of what the
// apploader does with the file, and what -loaddol does in the oracle when it stands in
// for the apploader.
func (d *DOL) Load(write func(addr uint32, b []byte)) {
	for _, s := range d.Segments {
		write(s.Addr, s.Data)
	}
	if d.BSSSize > 0 {
		write(d.BSSAddr, make([]byte, d.BSSSize))
	}
}

// Flat lays the executable out as one contiguous image spanning its segments, and
// returns the address it starts at. This is what the disassembler and the code tracer
// want: a byte slice indexed by address, in which a branch to anywhere in the program
// lands on the instruction that is really there.
//
// The gaps between segments read as zero, which decodes to an illegal instruction — so
// a tracer that wanders into one stops, which is the correct thing for it to do.
func (d *DOL) Flat() (base uint32, mem []byte) {
	lo, hi := d.Segments[0].Addr, uint32(0)
	for _, s := range d.Segments {
		if s.Addr < lo {
			lo = s.Addr
		}
		if end := s.Addr + s.Size; end > hi {
			hi = end
		}
	}
	mem = make([]byte, hi-lo)
	for _, s := range d.Segments {
		copy(mem[s.Addr-lo:], s.Data)
	}
	return lo, mem
}

// Text reports whether an address falls in a code segment — what the tracer needs in
// order to refuse to walk into data.
func (d *DOL) Text(addr uint32) bool {
	for _, s := range d.Segments {
		if s.Text && addr >= s.Addr && addr < s.Addr+s.Size {
			return true
		}
	}
	return false
}
