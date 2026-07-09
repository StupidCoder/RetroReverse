package n64

// rom.go reads a Nintendo 64 cartridge image: it normalises the image's byte
// order, parses the 64-byte header, and identifies the boot chip (CIC) that the
// cartridge shipped with.
//
// Byte order. The cartridge bus is big-endian, but dumpers wrote images in three
// different orders and the file extension does not reliably say which. All three
// are identified by the first four bytes, whose big-endian value is always
// 0x80371240 (the PI domain-1 configuration word):
//
//	z64  80 37 12 40   big-endian, the native cartridge order
//	v64  37 80 40 12   16-bit halves swapped ("byteswapped")
//	n64  40 12 37 80   32-bit words byte-reversed (little-endian)
//
// Everything downstream of Load sees big-endian bytes regardless of the input.
//
// CIC. The boot chip lives in the cartridge, not in the ROM, so it is not on the
// image — but IPL3 is (ROM 0x40..0x1000), and each CIC ships with exactly one
// IPL3 revision. Hashing IPL3 therefore names the chip, which in turn supplies
// the seed value IPL2 leaves in $s6 for IPL3's integrity check. That seed is the
// one boot input this package cannot read off the medium (see boot.go).

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"strings"
)

// ByteOrder identifies how a cartridge image's bytes are arranged on disk.
type ByteOrder int

const (
	OrderZ64 ByteOrder = iota // big-endian, native
	OrderV64                  // 16-bit halves swapped
	OrderN64                  // 32-bit words byte-reversed
)

func (b ByteOrder) String() string {
	switch b {
	case OrderZ64:
		return "z64 (big-endian)"
	case OrderV64:
		return "v64 (byteswapped)"
	case OrderN64:
		return "n64 (little-endian)"
	}
	return "?"
}

// The header's first word, in each of the three orders.
const (
	magicZ64 = 0x80371240
	magicV64 = 0x37804012
	magicN64 = 0x40123780
)

// HeaderSize is the length of the cartridge header, which precedes IPL3.
const HeaderSize = 0x40

// IPL3Start and IPL3End bound the boot code the cartridge carries. IPL3 is
// copied into SP DMEM and executed from 0xA4000040.
const (
	IPL3Start = 0x40
	IPL3End   = 0x1000
)

// CIC names a boot chip revision and the seed it leaves in $s6 for IPL3.
type CIC struct {
	Name string
	Seed uint32 // the $s6 value IPL2 passes to IPL3
}

// cicByIPL3 maps the CRC-32 of the cartridge's IPL3 (ROM 0x40..0x1000) to the
// boot chip that revision shipped with.
//
// Only revisions confirmed against an image in hand appear here. A seed is not a
// value that can be guessed and left unchecked: IPL3 verifies itself against it,
// so a wrong seed makes the cartridge dead-loop at boot rather than fail loudly.
// Adding a revision means having the image, and confirming the seed by watching
// IPL3 run to its jump (see boot.go).
var cicByIPL3 = map[uint32]CIC{
	0x90BB6CB5: {"CIC-NUS-6102", 0x3F},
}

// Header is the 64-byte cartridge header, in native (big-endian) form.
type Header struct {
	PIConfig  uint32 // 0x00 PI BSD domain-1 timing + the 0x80371240 magic
	ClockRate uint32 // 0x04
	Entry     uint32 // 0x08 address IPL3 copies the boot segment to, and jumps to
	Release   uint32 // 0x0C libultra version
	CRC1      uint32 // 0x10 checksum over ROM 0x1000..0x101000
	CRC2      uint32 // 0x14
	Name      string // 0x20..0x33, space-padded ASCII
	CartType  byte   // 0x3B 'N' = cartridge
	CartID    string // 0x3C..0x3D
	Country   byte   // 0x3E 'E' = USA, 'J' = Japan, 'P' = Europe
	Version   byte   // 0x3F
}

// ROM is a loaded cartridge image, normalised to big-endian.
type ROM struct {
	Data   []byte    // the full image, big-endian
	Order  ByteOrder // the order the file was stored in
	Header Header
	CIC    CIC
	IPL3   uint32 // CRC-32 of ROM 0x40..0x1000, the CIC fingerprint
	MD5    string // MD5 of the file as stored on disk, for pinning in the writeup
}

// Load reads a cartridge image from path and normalises it to big-endian.
func Load(path string) (*ROM, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Decode(data)
}

// Normalize converts a cartridge image to big-endian in place, reporting the
// order it was stored in. It recognises an image only by its first word, so it
// works on any cartridge; ok is false when that word matches no known order.
//
// Tools that read a raw image (the disassembler, the tracer) call this rather
// than Decode, which additionally insists on knowing the boot chip.
func Normalize(data []byte) (ByteOrder, bool) {
	if len(data) < 4 {
		return 0, false
	}
	switch binary.BigEndian.Uint32(data) {
	case magicZ64:
		return OrderZ64, true
	case magicV64:
		swapPairs(data)
		return OrderV64, true
	case magicN64:
		reverseWords(data)
		return OrderN64, true
	}
	return 0, false
}

// Decode normalises a cartridge image in place and parses its header. The MD5 is
// taken before normalisation, so it identifies the file as distributed.
func Decode(data []byte) (*ROM, error) {
	if len(data) < IPL3End {
		return nil, fmt.Errorf("n64: image is %d bytes, shorter than the header + IPL3 (%d)", len(data), IPL3End)
	}
	sum := md5.Sum(data)
	r := &ROM{Data: data, MD5: fmt.Sprintf("%x", sum)}

	order, ok := Normalize(data)
	if !ok {
		return nil, fmt.Errorf("n64: not a cartridge image: first word 0x%08X matches no known byte order",
			binary.BigEndian.Uint32(data))
	}
	r.Order = order

	r.Header = parseHeader(data)
	r.IPL3 = crc32.ChecksumIEEE(data[IPL3Start:IPL3End])
	if c, ok := cicByIPL3[r.IPL3]; ok {
		r.CIC = c
	} else {
		return nil, fmt.Errorf("n64: unrecognised IPL3 (crc32 0x%08X): the boot chip, and so the $s6 "+
			"seed IPL3 checks itself against, cannot be determined", r.IPL3)
	}
	return r, nil
}

// swapPairs converts v64 order to big-endian by exchanging each byte pair.
func swapPairs(d []byte) {
	for i := 0; i+1 < len(d); i += 2 {
		d[i], d[i+1] = d[i+1], d[i]
	}
}

// reverseWords converts n64 order to big-endian by reversing each 32-bit word.
func reverseWords(d []byte) {
	for i := 0; i+3 < len(d); i += 4 {
		d[i], d[i+1], d[i+2], d[i+3] = d[i+3], d[i+2], d[i+1], d[i]
	}
}

func parseHeader(d []byte) Header {
	be := binary.BigEndian
	return Header{
		PIConfig:  be.Uint32(d[0x00:]),
		ClockRate: be.Uint32(d[0x04:]),
		Entry:     be.Uint32(d[0x08:]),
		Release:   be.Uint32(d[0x0C:]),
		CRC1:      be.Uint32(d[0x10:]),
		CRC2:      be.Uint32(d[0x14:]),
		Name:      strings.TrimRight(string(d[0x20:0x34]), " \x00"),
		CartType:  d[0x3B],
		CartID:    string(d[0x3C:0x3E]),
		Country:   d[0x3E],
		Version:   d[0x3F],
	}
}

// Region names the country code in human terms.
func (h Header) Region() string {
	switch h.Country {
	case 'E':
		return "USA"
	case 'J':
		return "Japan"
	case 'P':
		return "Europe"
	case 'D':
		return "Germany"
	case 'F':
		return "France"
	case 'U':
		return "Australia"
	}
	return fmt.Sprintf("unknown (0x%02X)", h.Country)
}

func (r *ROM) String() string {
	return fmt.Sprintf("%q [%s%s] %s, %d bytes, entry 0x%08X, %s, stored as %s",
		r.Header.Name, r.Header.CartID, string(r.Header.Country), r.CIC.Name,
		len(r.Data), r.Header.Entry, r.Header.Region(), r.Order)
}
