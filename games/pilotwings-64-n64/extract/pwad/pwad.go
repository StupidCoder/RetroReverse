// Package pwad reads Pilotwings 64's cartridge archive.
//
// From cart offset 0x0DE720 to 0x0618B6C the ROM holds a contiguous IFF-85
// archive: 1,273 back-to-back FORM chunks, each a big-endian fourCC, a 32-bit
// payload size, a FORM type, and a list of chunks. Nothing outside this range
// is part of it — the program image lies below, the audio banks above.
//
//	FORM <u32 size> <type>            type: UVTX, UVMD, UVCT, UPWT, ... (21 seen)
//	  PAD  <u32 4>  ...               padding, always first
//	  <tag> <u32 size> <bytes>        a plain chunk, or:
//	  GZIP <u32 size>                 "the chunk that follows is compressed"
//	    <tag> <u32 uncompressedSize>
//	    MIO0 ...                      the codec in extract/mio0
//
// GZIP is a wrapper tag, not RFC 1952: its payload is a chunk header whose size
// field is the *decompressed* length, followed by an MIO0 stream. A FORM carries
// between zero and thirteen of them.
//
// The first FORM (type UVRM) holds the directory, chunk TABL: 1,272 entries of
// (fourCC type, u32 total byte length), one per following FORM, in ROM order.
// The length counts the FORM's own 8-byte header, so the entries are a walkable
// index — resource i sits at 0x0DF5B0 plus the sum of the preceding lengths, and
// the running sum lands exactly on the archive's end. That arithmetic, checked
// against an independent walk of the FORM chain, is what proves the container
// (Check).
//
// Chunk sizes in this archive are all even and every FORM's chunk list ends
// exactly on its boundary, so the IFF word-padding rule never fires; Open
// asserts both rather than silently tolerating a violation.
package pwad

import (
	"encoding/binary"
	"fmt"

	"retroreverse.com/games/pilotwings-64-n64/extract/mio0"
)

// ArchiveStart is the cart offset of the directory FORM. Nothing locates it for
// us: the game's loader reads the directory from this literal address.
const ArchiveStart = 0x000DE720

const (
	dirFormType  = "UVRM"
	dirChunkTag  = "TABL"
	compressTag  = "GZIP"
	maxFormSize  = 0x200000
	maxGapBytes  = 16 // zero padding tolerated between FORMs (one 12-byte gap exists)
	dirEntrySize = 8
)

// Chunk is one chunk inside a FORM.
type Chunk struct {
	Tag  string
	Size uint32 // stored length of the chunk body
	Off  uint32 // cart offset of the chunk body (just past the 8-byte header)

	// Set only when Tag == "GZIP": the compressed payload's own header.
	InnerTag         string
	UncompressedSize uint32
}

// Compressed reports whether the chunk body is an MIO0 stream behind a GZIP
// wrapper.
func (c Chunk) Compressed() bool { return c.Tag == compressTag }

// Form is one FORM chunk of the archive.
type Form struct {
	Index  int    // resource index (directory order); -1 for the directory FORM itself
	Off    uint32 // cart offset of the FORM header
	Type   string // the FORM type: UVTX, UVMD, ...
	Size   uint32 // the FORM's payload size, as stored
	Chunks []Chunk
}

// Total is the FORM's byte length including its 8-byte header — the value the
// directory stores.
func (f Form) Total() uint32 { return f.Size + 8 }

// DirEntry is one TABL record.
type DirEntry struct {
	Type  string
	Total uint32 // total byte length of the FORM, header included
}

// Archive is the parsed container.
type Archive struct {
	rom []byte

	// Forms holds every FORM in ROM order. Forms[0] is the UVRM directory; the
	// resources are Forms[1:], numbered from 0 by Index.
	Forms []Form

	// Dir is the TABL directory: Dir[i] describes Forms[i+1].
	Dir []DirEntry

	// End is the cart offset one past the last FORM.
	End uint32
}

func be32(b []byte, o uint32) uint32 { return binary.BigEndian.Uint32(b[o:]) }
func tag(b []byte, o uint32) string  { return string(b[o : o+4]) }

// Open parses the archive out of a byte-order-normalised cartridge image (the
// bytes n64.Load returns, not the raw file).
func Open(rom []byte) (*Archive, error) {
	a := &Archive{rom: rom}
	if len(rom) < ArchiveStart+12 || tag(rom, ArchiveStart) != "FORM" {
		return nil, fmt.Errorf("pwad: no FORM at the archive start 0x%07X", ArchiveStart)
	}

	off := uint32(ArchiveStart)
	for off+12 <= uint32(len(rom)) && tag(rom, off) == "FORM" {
		f, err := readForm(rom, off)
		if err != nil {
			return nil, err
		}
		f.Index = len(a.Forms) - 1 // Forms[0] is the directory, so it gets -1
		a.Forms = append(a.Forms, f)

		next := off + f.Total()
		a.End = next
		// The archive is contiguous but for one 12-byte run of zeros after the
		// directory FORM. Skip only zeros, and only a few of them.
		gap := uint32(0)
		for next+gap < uint32(len(rom)) && rom[next+gap] == 0 && gap < maxGapBytes {
			gap++
		}
		if next+gap+4 > uint32(len(rom)) || tag(rom, next+gap) != "FORM" {
			break // end of the archive
		}
		if gap != 0 {
			// Zeros only, already checked. Recorded so a future change to the
			// image shows up rather than being absorbed.
			if gap%4 != 0 {
				return nil, fmt.Errorf("pwad: %d-byte gap after FORM at 0x%07X is not word-sized", gap, off)
			}
		}
		off = next + gap
	}

	if len(a.Forms) == 0 || a.Forms[0].Type != dirFormType {
		return nil, fmt.Errorf("pwad: first FORM is not %s", dirFormType)
	}
	dir, err := a.readDir()
	if err != nil {
		return nil, err
	}
	a.Dir = dir
	return a, nil
}

func readForm(rom []byte, off uint32) (Form, error) {
	size := be32(rom, off+4)
	if size < 4 || size > maxFormSize || uint64(off)+8+uint64(size) > uint64(len(rom)) {
		return Form{}, fmt.Errorf("pwad: FORM at 0x%07X has implausible size %#x", off, size)
	}
	f := Form{Off: off, Type: tag(rom, off+8), Size: size}
	end := off + 8 + size
	for q := off + 12; q+8 <= end; {
		c := Chunk{Tag: tag(rom, q), Size: be32(rom, q+4), Off: q + 8}
		if c.Size%2 != 0 {
			return Form{}, fmt.Errorf("pwad: odd-sized chunk %q at 0x%07X", c.Tag, q)
		}
		if c.Off+c.Size > end {
			return Form{}, fmt.Errorf("pwad: chunk %q at 0x%07X overruns its FORM", c.Tag, q)
		}
		if c.Compressed() {
			if c.Size < 8 {
				return Form{}, fmt.Errorf("pwad: GZIP chunk at 0x%07X is too small", q)
			}
			c.InnerTag = tag(rom, c.Off)
			c.UncompressedSize = be32(rom, c.Off+4)
		}
		f.Chunks = append(f.Chunks, c)
		q = c.Off + c.Size
		if q == end {
			return f, nil
		}
	}
	return Form{}, fmt.Errorf("pwad: FORM %s at 0x%07X: chunk list does not end on the FORM boundary", f.Type, off)
}

// readDir parses the TABL chunk of the directory FORM.
func (a *Archive) readDir() ([]DirEntry, error) {
	var raw []byte
	for _, c := range a.Forms[0].Chunks {
		if c.Compressed() && c.InnerTag == dirChunkTag {
			b, err := a.Data(c)
			if err != nil {
				return nil, fmt.Errorf("pwad: directory: %w", err)
			}
			raw = b
		}
	}
	if raw == nil {
		return nil, fmt.Errorf("pwad: directory FORM has no %s chunk", dirChunkTag)
	}
	if len(raw)%dirEntrySize != 0 {
		return nil, fmt.Errorf("pwad: directory length %d is not a multiple of %d", len(raw), dirEntrySize)
	}
	out := make([]DirEntry, len(raw)/dirEntrySize)
	for i := range out {
		o := uint32(i * dirEntrySize)
		out[i] = DirEntry{Type: tag(raw, o), Total: be32(raw, o+4)}
	}
	return out, nil
}

// Data returns a chunk's body, decompressing it if it is GZIP-wrapped. The
// declared uncompressed size is asserted, so a codec regression fails here
// rather than downstream.
func (a *Archive) Data(c Chunk) ([]byte, error) {
	if !c.Compressed() {
		return a.rom[c.Off : c.Off+c.Size], nil
	}
	out, err := mio0.Decompress(a.rom[c.Off+8 : c.Off+c.Size])
	if err != nil {
		return nil, fmt.Errorf("pwad: %s chunk at 0x%07X: %w", c.InnerTag, c.Off, err)
	}
	if uint32(len(out)) != c.UncompressedSize {
		return nil, fmt.Errorf("pwad: %s chunk at 0x%07X: inflated to %d bytes, header says %d",
			c.InnerTag, c.Off, len(out), c.UncompressedSize)
	}
	return out, nil
}

// Resource returns the FORM for directory index i.
func (a *Archive) Resource(i int) (Form, error) {
	if i < 0 || i >= len(a.Dir) {
		return Form{}, fmt.Errorf("pwad: resource %d out of range (%d resources)", i, len(a.Dir))
	}
	return a.Forms[i+1], nil
}

// ByType returns the resource indices of every FORM of the given type, in
// archive order.
func (a *Archive) ByType(t string) []int {
	var out []int
	for _, f := range a.Forms[1:] {
		if f.Type == t {
			out = append(out, f.Index)
		}
	}
	return out
}

// Check verifies the container against itself: the independently-walked FORM
// chain must agree with the directory on every resource's type and byte length,
// and the directory's running sum must reproduce every FORM's offset and land
// exactly on the archive's end. Either alone could be a coincidence of a wrong
// reading; together they pin the format.
func (a *Archive) Check() error {
	if len(a.Dir) != len(a.Forms)-1 {
		return fmt.Errorf("pwad: directory has %d entries but the walk found %d resources",
			len(a.Dir), len(a.Forms)-1)
	}
	cur := a.Forms[1].Off
	for i, d := range a.Dir {
		f := a.Forms[i+1]
		if f.Type != d.Type {
			return fmt.Errorf("pwad: resource %d: walk says %s, directory says %s", i, f.Type, d.Type)
		}
		if f.Total() != d.Total {
			return fmt.Errorf("pwad: resource %d (%s): walk says %d bytes, directory says %d",
				i, f.Type, f.Total(), d.Total)
		}
		if f.Off != cur {
			return fmt.Errorf("pwad: resource %d (%s): at 0x%07X, directory sum says 0x%07X",
				i, f.Type, f.Off, cur)
		}
		cur += d.Total
	}
	if cur != a.End {
		return fmt.Errorf("pwad: directory sum ends at 0x%07X, the archive ends at 0x%07X", cur, a.End)
	}
	return nil
}
