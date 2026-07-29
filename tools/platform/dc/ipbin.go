package dc

// ipbin.go parses the boot sector's metadata block — the first 256 bytes of
// IP.BIN, the "initial program" occupying the high-density area's first 16
// sectors. The layout is the platform's published submission format: fixed-
// width space-padded text fields, followed (from 0x100) by the bootstrap code
// the BIOS runs.

import (
	"fmt"
	"strings"
)

// IPBin is the metadata header of a disc's initial program.
type IPBin struct {
	HardwareID string // "SEGA SEGAKATANA"
	MakerID    string // "SEGA ENTERPRISES"
	DeviceInfo string // CRC + "GD-ROM1/1"
	AreaSyms   string // region letters, "JUE" positions
	Peripherals string
	ProductNo  string // "MK-51035"
	Version    string // "V1.004"
	ReleaseDate string
	BootFile   string // "1ST_READ.BIN"
	Company    string
	Title      string
}

// parseIPBin reads the metadata out of the first data sector of the
// high-density area. It refuses anything that does not open with the hardware
// identifier: that string is how a Dreamcast recognises a bootable disc, so
// its absence means the track mapping is wrong, not the disc.
func parseIPBin(sector []byte) (IPBin, error) {
	if len(sector) < 0x100 {
		return IPBin{}, fmt.Errorf("ipbin: sector too short (%d bytes)", len(sector))
	}
	field := func(off, n int) string { return strings.TrimRight(string(sector[off:off+n]), " ") }
	ip := IPBin{
		HardwareID:  field(0x00, 16),
		MakerID:     field(0x10, 16),
		DeviceInfo:  field(0x20, 16),
		AreaSyms:    field(0x30, 8),
		Peripherals: field(0x38, 8),
		ProductNo:   field(0x40, 10),
		Version:     field(0x4A, 6),
		ReleaseDate: field(0x50, 16),
		BootFile:    field(0x60, 16),
		Company:     field(0x70, 16),
		Title:       field(0x80, 128),
	}
	if ip.HardwareID != "SEGA SEGAKATANA" {
		return IPBin{}, fmt.Errorf("ipbin: hardware ID %q, want SEGA SEGAKATANA — not a Dreamcast boot sector", ip.HardwareID)
	}
	return ip, nil
}
