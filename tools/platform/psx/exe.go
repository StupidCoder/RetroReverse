package psx

// exe.go loads the PS-X EXE executable format — the file a PSX disc boots (named
// by SYSTEM.CNF's BOOT= line). The layout is a fixed 0x800-byte header followed
// by the raw text image:
//
//	0x000  8  magic "PS-X EXE"
//	0x010  4  pc0     initial program counter
//	0x014  4  gp0     initial global pointer ($gp)
//	0x018  4  t_addr  RAM address the text is loaded to
//	0x01C  4  t_size  text size in bytes
//	0x020  4  d_addr / 0x024 d_size   (data section; usually 0)
//	0x028  4  b_addr / 0x02C b_size   (bss  section; cleared by the BIOS)
//	0x030  4  s_addr / 0x034 s_size   (initial stack base and size)
//	0x04C ..  ASCII region marker ("Sony Computer Entertainment Inc. for ...")
//	0x800 ..  text image (t_size bytes)
//
// The header mirrors the role of the MS-DOS MZ header in tools/dos/mz.go: it
// yields the entry point, the load address and the raw code to hand the CPU.

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// defaultSP is the stack top the BIOS uses when the header leaves s_addr zero.
const defaultSP = 0x801FFF00

// EXE is a parsed PS-X EXE.
type EXE struct {
	PC0, GP0     uint32
	TAddr, TSize uint32
	DAddr, DSize uint32
	BAddr, BSize uint32
	SAddr, SSize uint32
	Marker       string // region marker string at 0x4C
	Text         []byte // the text image (TSize bytes)
}

// ParseEXE validates the header and returns the executable.
func ParseEXE(data []byte) (*EXE, error) {
	if len(data) < 0x800 {
		return nil, fmt.Errorf("psx: EXE too small (%d bytes)", len(data))
	}
	if string(data[0:8]) != "PS-X EXE" {
		return nil, fmt.Errorf("psx: not a PS-X EXE (magic %q)", data[0:8])
	}
	u := func(off int) uint32 { return binary.LittleEndian.Uint32(data[off : off+4]) }
	e := &EXE{
		PC0:   u(0x10),
		GP0:   u(0x14),
		TAddr: u(0x18),
		TSize: u(0x1C),
		DAddr: u(0x20),
		DSize: u(0x24),
		BAddr: u(0x28),
		BSize: u(0x2C),
		SAddr: u(0x30),
		SSize: u(0x34),
	}
	e.Marker = strings.TrimRight(string(data[0x4C:0x800]), " \x00")
	end := 0x800 + int(e.TSize)
	if end > len(data) {
		return nil, fmt.Errorf("psx: text size %d overruns file (%d bytes)", e.TSize, len(data))
	}
	e.Text = data[0x800:end]
	return e, nil
}

// InitialSP is the stack pointer the BIOS seeds before entering the program.
func (e *EXE) InitialSP() uint32 {
	if e.SAddr != 0 {
		return e.SAddr + e.SSize
	}
	return defaultSP
}

// Describe renders the header fields for the psxinfo -exe command.
func (e *EXE) Describe() string {
	var b strings.Builder
	fmt.Fprintf(&b, "PS-X EXE\n")
	fmt.Fprintf(&b, "  entry PC  = 0x%08X\n", e.PC0)
	fmt.Fprintf(&b, "  gp        = 0x%08X\n", e.GP0)
	fmt.Fprintf(&b, "  text addr = 0x%08X  size %d (0x%X)\n", e.TAddr, e.TSize, e.TSize)
	fmt.Fprintf(&b, "  data addr = 0x%08X  size %d\n", e.DAddr, e.DSize)
	fmt.Fprintf(&b, "  bss  addr = 0x%08X  size %d\n", e.BAddr, e.BSize)
	fmt.Fprintf(&b, "  stack     = 0x%08X  size %d  (sp=0x%08X)\n", e.SAddr, e.SSize, e.InitialSP())
	if e.Marker != "" {
		fmt.Fprintf(&b, "  marker    = %q\n", e.Marker)
	}
	inText := e.PC0 >= e.TAddr && e.PC0 < e.TAddr+e.TSize
	fmt.Fprintf(&b, "  entry in text: %v\n", inText)
	return b.String()
}
