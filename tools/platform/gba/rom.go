// Package gba reads Game Boy Advance cartridge images (.gba files).
//
// A GBA cartridge image is a verbatim dump of the cartridge mask ROM: byte N of
// the file is the byte the CPU reads at bus address 0x08000000+N (the ROM is
// also mirrored at 0x0A000000 and 0x0C000000 with different waitstates). There
// is no container and no filesystem — the only imposed structure is the 192-byte
// cartridge header at the front:
//
//	0x00  4    ARM branch to the entry point (usually B past the header)
//	0x04  156  Nintendo logo bitmap (checked by the BIOS at boot)
//	0xA0  12   game title, ASCII, zero-padded
//	0xAC  4    game code (type letter, 2-letter short name, region)
//	0xB0  2    maker code ("01" = Nintendo)
//	0xB2  1    fixed 0x96
//	0xB3  1    main unit code (0x00 = AGB)
//	0xB4  1    device type
//	0xB5  7    reserved, zero
//	0xBC  1    software version
//	0xBD  1    header checksum over 0xA0..0xBC
//	0xBE  2    reserved, zero
//
// The save memory type is not declared in the header: Nintendo's library
// drivers each embed an ASCII ID string ("EEPROM_V124", "FLASH1M_V103", ...)
// and detection scans the ROM for them, which is what real flash carts and
// emulators do too.
package gba

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strings"
)

// HeaderLen is the size of the cartridge header in bytes.
const HeaderLen = 0xC0

// ROMBase is the CPU bus address of ROM byte 0 (waitstate-0 mirror).
const ROMBase = 0x08000000

// Header is the decoded 192-byte cartridge header.
type Header struct {
	EntryInstr uint32 // raw ARM word at 0x00
	EntryAddr  uint32 // branch target as a CPU address, 0 if 0x00 is not a B
	LogoMD5    string // MD5 of the 156-byte compressed Nintendo logo bitmap
	Title      string // 0xA0, trailing NULs stripped
	GameCode   string // 0xAC
	MakerCode  string // 0xB0
	Fixed96    byte   // 0xB2, must be 0x96
	UnitCode   byte   // 0xB3
	DeviceType byte   // 0xB4
	Version    byte   // 0xBC
	Checksum   byte   // 0xBD as stored
	ChecksumOK bool   // stored checksum matches ComplementCheck
}

// ROM is a loaded cartridge image.
type ROM struct {
	Data   []byte
	Header Header
}

// Parse decodes a cartridge image. It fails only on images too short to carry
// a header; a bad checksum or missing branch is reported in the Header fields
// (the real BIOS refuses such a cart, but we still want to look inside one).
func Parse(data []byte) (*ROM, error) {
	if len(data) < HeaderLen {
		return nil, fmt.Errorf("gba: image is %d bytes, shorter than the %d-byte header", len(data), HeaderLen)
	}
	h := Header{
		EntryInstr: binary.LittleEndian.Uint32(data[0x00:]),
		LogoMD5:    fmt.Sprintf("%x", md5.Sum(data[0x04:0xA0])),
		Title:      strings.TrimRight(string(data[0xA0:0xAC]), "\x00"),
		GameCode:   string(data[0xAC:0xB0]),
		MakerCode:  string(data[0xB0:0xB2]),
		Fixed96:    data[0xB2],
		UnitCode:   data[0xB3],
		DeviceType: data[0xB4],
		Version:    data[0xBC],
		Checksum:   data[0xBD],
	}
	// The entry word is virtually always an unconditional ARM branch
	// (cond=AL, opcode B): 0xEA | signed 24-bit word offset from PC+8.
	if h.EntryInstr>>24 == 0xEA {
		off := int32(h.EntryInstr<<8) >> 8 // sign-extend the 24-bit field
		h.EntryAddr = uint32(int64(ROMBase) + 8 + int64(off)*4)
	}
	h.ChecksumOK = h.Checksum == ComplementCheck(data)
	return &ROM{Data: data, Header: h}, nil
}

// ComplementCheck computes the header checksum the BIOS verifies:
// the two's complement of (sum of bytes 0xA0..0xBC) + 0x19.
func ComplementCheck(data []byte) byte {
	var sum byte
	for _, b := range data[0xA0:0xBD] {
		sum += b
	}
	return -(sum + 0x19)
}

// saveIDs are the ASCII ID strings Nintendo's save drivers embed in ROM,
// longest-match first so FLASH512/FLASH1M win over their FLASH prefix.
var saveIDs = []struct{ id, desc string }{
	{"EEPROM_V", "EEPROM (serial, 512 B or 8 KiB)"},
	{"FLASH1M_V", "Flash 128 KiB"},
	{"FLASH512_V", "Flash 64 KiB"},
	{"FLASH_V", "Flash 64 KiB"},
	{"SRAM_V", "SRAM 32 KiB"},
	{"SRAM_F_V", "SRAM 32 KiB"},
}

// SaveType scans the image for a save-driver ID string and returns the full
// versioned string ("EEPROM_V124") and a human description. Empty strings mean
// no driver ID was found (password-save games have none).
func (r *ROM) SaveType() (id, desc string) {
	s := string(r.Data)
	for _, c := range saveIDs {
		i := strings.Index(s, c.id)
		if i < 0 {
			continue
		}
		// Take the ID plus its trailing version digits.
		j := i + len(c.id)
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		return s[i:j], c.desc
	}
	return "", ""
}
