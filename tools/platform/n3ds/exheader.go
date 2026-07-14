package n3ds

import (
	"encoding/binary"
	"fmt"
)

// CodeSegInfo describes where one segment of .code is mapped and how big it is.
// NumPages is the segment's *reserved* extent in 0x1000-byte pages; Size is the
// number of bytes actually occupied. The next segment's Address is the previous
// one's Address plus NumPages pages, so the reserved extents tile the address
// space contiguously while Size says how much of each tile holds content.
type CodeSegInfo struct {
	Address  uint32
	NumPages uint32
	Size     uint32
}

// PageSize is the ARM11 small-page size used by the code-set layout.
const PageSize = 0x1000

// Extent returns the segment's reserved byte extent (NumPages × PageSize).
func (s CodeSegInfo) Extent() uint32 { return s.NumPages * PageSize }

// ExHeader is the parsed system-control info of the NCCH extended header. This
// package parses the SCI (the part the loader needs); the access-control info
// that follows it is exposed raw.
type ExHeader struct {
	Title        string
	Flag         byte
	RemasterVer  uint16
	Text         CodeSegInfo
	StackSize    uint32
	ROData       CodeSegInfo
	Data         CodeSegInfo
	BSSSize      uint32
	Dependencies []uint64 // program IDs of the sysmodules this title links against
	SaveDataSize uint64
	JumpID       uint64
	ACI          []byte // access-control info, unparsed
}

// CompressedExeFSCode reports whether ExeFS/.code is BLZ-compressed (flag bit 0).
func (e *ExHeader) CompressedExeFSCode() bool { return e.Flag&1 != 0 }

// SDApplication reports the SD-application flag (bit 1).
func (e *ExHeader) SDApplication() bool { return e.Flag&2 != 0 }

// ParseExHeader parses the SCI at the start of the ExHeader.
func ParseExHeader(b []byte) (*ExHeader, error) {
	if len(b) < 0x200 {
		return nil, fmt.Errorf("n3ds: ExHeader too short (%d bytes, want >= 0x200)", len(b))
	}
	u32 := func(off int) uint32 { return binary.LittleEndian.Uint32(b[off:]) }
	seg := func(off int) CodeSegInfo {
		return CodeSegInfo{Address: u32(off), NumPages: u32(off + 4), Size: u32(off + 8)}
	}

	e := &ExHeader{
		Title:       string(trimNul(b[0x00:0x08])),
		Flag:        b[0x0D],
		RemasterVer: binary.LittleEndian.Uint16(b[0x0E:]),
		Text:        seg(0x10),
		StackSize:   u32(0x1C),
		ROData:      seg(0x20),
		Data:        seg(0x30),
		BSSSize:     u32(0x3C),
	}
	for i := 0; i < 48; i++ {
		if id := binary.LittleEndian.Uint64(b[0x40+i*8:]); id != 0 {
			e.Dependencies = append(e.Dependencies, id)
		}
	}
	e.SaveDataSize = binary.LittleEndian.Uint64(b[0x1C0:])
	e.JumpID = binary.LittleEndian.Uint64(b[0x1C8:])
	if len(b) >= 0x400 {
		e.ACI = b[0x200:0x400]
	}

	if err := e.validate(); err != nil {
		return nil, err
	}
	return e, nil
}

// validate checks the two structural invariants the loader depends on: each
// segment fits inside its reserved page extent, and the reserved extents tile
// the address space contiguously (rodata starts where text's extent ends, and
// data where rodata's does). A violation means the header was misread — most
// likely because the container is still encrypted.
func (e *ExHeader) validate() error {
	for _, s := range []struct {
		name string
		seg  CodeSegInfo
	}{{"text", e.Text}, {"rodata", e.ROData}, {"data", e.Data}} {
		if s.seg.Size > s.seg.Extent() {
			return fmt.Errorf("n3ds: ExHeader %s size 0x%x exceeds its %d-page extent 0x%x",
				s.name, s.seg.Size, s.seg.NumPages, s.seg.Extent())
		}
	}
	if got, want := e.ROData.Address, e.Text.Address+e.Text.Extent(); got != want {
		return fmt.Errorf("n3ds: ExHeader rodata address 0x%08x does not follow text extent (expected 0x%08x)", got, want)
	}
	if got, want := e.Data.Address, e.ROData.Address+e.ROData.Extent(); got != want {
		return fmt.Errorf("n3ds: ExHeader data address 0x%08x does not follow rodata extent (expected 0x%08x)", got, want)
	}
	return nil
}

// CodeSize is the total byte length the decompressed .code must have: the three
// segments' reserved page extents, laid end to end. Each segment is padded with
// zeroes from its Size out to its Extent, so .code is always a whole number of
// pages and every segment starts on one. Checking a decompressor's output
// against this is the round-trip proof that the BLZ decode is correct — the
// expected size is pinned by header arithmetic that the compressed stream knows
// nothing about.
func (e *ExHeader) CodeSize() uint32 {
	return e.Text.Extent() + e.ROData.Extent() + e.Data.Extent()
}

// BSSAddress is where the zero-initialised segment begins: immediately after the
// data segment's bytes.
func (e *ExHeader) BSSAddress() uint32 { return e.Data.Address + e.Data.Size }
