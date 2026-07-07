package threedo

// aif.go loads the ARM Image Format (AIF) executable the 3DO boots — the file
// named by AppStartup ("$boot/Launchme") and the Portfolio programs under
// System/. It is the 3DO counterpart of tools/psx/exe.go and tools/dos/mz.go.
//
// An AIF file is a 128-byte header followed by the image, big-endian on the 3DO:
//
//	0x00  BL DecompressCode   (MOV r0,r0 / NOP when not compressed)
//	0x04  BL SelfRelocCode    (NOP when not self-relocating)
//	0x08  BL ZeroInitCode     (clears the zero-init region)
//	0x0C  BL EntryPoint
//	0x10  Program-exit SWI
//	0x14  u32 ReadOnly size   (code + RO data; includes the 128-byte header)
//	0x18  u32 ReadWrite size  (initialised data)
//	0x1C  u32 Debug size
//	0x20  u32 ZeroInit size   (BSS)
//	0x28  u32 Image base      (the link address; 0 for the relocatable game)
//	0x30  u32 Address mode    (26 or 32; the 3DO uses 32)
//
// For an executable AIF the header itself is entered at its first word, so the
// four leading branch words run the decompress / self-relocate / zero-init /
// entry sequence in order. This loader keeps the whole image and exposes those
// fields; the machine (machine.go) copies the image into RAM and enters at the
// header.

import "fmt"

// AIF is a parsed ARM Image Format executable.
type AIF struct {
	Decompress  uint32 // first header word (0xE1A00000 == not compressed)
	SelfReloc   bool   // the self-relocate slot is a real BL, not a NOP
	EntryTarget uint32 // absolute entry (image base + resolved BL entry)
	ExitSWI     uint32
	ROSize      uint32 // includes the 128-byte header
	RWSize      uint32
	ZeroSize    uint32 // BSS
	ImageBase   uint32
	AddrMode    uint32 // 26 or 32
	Image       []byte // the whole file (header + code + data), loaded at ImageBase
}

const aifNOP = 0xE1A00000 // MOV r0,r0

func be32at(b []byte, o int) uint32 {
	return uint32(b[o])<<24 | uint32(b[o+1])<<16 | uint32(b[o+2])<<8 | uint32(b[o+3])
}

// ParseAIF validates and decodes an AIF header. It recognises the format by the
// header's four leading branch/NOP words plus a plausible 32-bit address mode.
func ParseAIF(data []byte) (*AIF, error) {
	if len(data) < 0x80 {
		return nil, fmt.Errorf("threedo: AIF too small (%d bytes)", len(data))
	}
	a := &AIF{
		Decompress: be32at(data, 0x00),
		ExitSWI:    be32at(data, 0x10),
		ROSize:     be32at(data, 0x14),
		RWSize:     be32at(data, 0x18),
		ZeroSize:   be32at(data, 0x20),
		ImageBase:  be32at(data, 0x28),
		AddrMode:   be32at(data, 0x30),
		Image:      data,
	}
	// The decompress/self-reloc slots are either a NOP or a BL (0xEBxxxxxx).
	isBranchOrNOP := func(w uint32) bool { return w == aifNOP || w>>24 == 0xEB }
	if !isBranchOrNOP(a.Decompress) || !isBranchOrNOP(be32at(data, 0x04)) {
		return nil, fmt.Errorf("threedo: not an AIF (leading words 0x%08X 0x%08X)", a.Decompress, be32at(data, 0x04))
	}
	if a.AddrMode != 26 && a.AddrMode != 32 {
		return nil, fmt.Errorf("threedo: AIF address mode %d (want 26 or 32)", a.AddrMode)
	}
	a.SelfReloc = be32at(data, 0x04) != aifNOP
	// Resolve BL entry at 0x0C: target = 0x0C + 8 + (signExtend24(imm) << 2).
	imm := be32at(data, 0x0C) & 0xFFFFFF
	off := imm << 2
	if imm&0x800000 != 0 { // sign-extend the 24-bit branch offset
		off |= 0xFC000000
	}
	a.EntryTarget = a.ImageBase + 0x0C + 8 + off
	return a, nil
}

// Describe renders the header fields for the operainfo/bootoracle listings.
func (a *AIF) Describe() string {
	reloc := "no"
	if a.SelfReloc {
		reloc = "yes"
	}
	return fmt.Sprintf("AIF executable\n"+
		"  image base = 0x%08X   address mode = %d-bit\n"+
		"  RO size    = 0x%X (%d, incl. 128-byte header)\n"+
		"  RW size    = 0x%X (%d)\n"+
		"  zero-init  = 0x%X (%d)\n"+
		"  self-reloc = %s   entry (BL) = 0x%08X\n",
		a.ImageBase, a.AddrMode, a.ROSize, a.ROSize, a.RWSize, a.RWSize,
		a.ZeroSize, a.ZeroSize, reloc, a.EntryTarget)
}
