package n3ds

import (
	"encoding/binary"
	"fmt"
)

const (
	ncchHeaderOff = 0x100 // after the RSA-2048 signature
	ncchMagic     = "NCCH"
)

// NCCH flag-byte indices, and the bits of flags[7] this package acts on.
const (
	flagCryptoMethod  = 3
	flagContentPlatfm = 4
	flagContentType   = 5
	flagContentUnit   = 6
	flagOther         = 7

	otherFixedKey = 1 << 0
	otherNoRomFS  = 1 << 1
	otherNoCrypto = 1 << 2
	otherSeedCryp = 1 << 5
)

// region is an offset/size pair from the NCCH header, already resolved from
// media units into byte offsets relative to the start of the partition.
type region struct {
	Offset int64
	Size   int64
}

func (r region) empty() bool { return r.Size == 0 }

// NCCH is one parsed NCCH container (a cartridge partition).
type NCCH struct {
	raw []byte // the partition's bytes

	ContentSize   int64
	PartitionID   uint64
	ProgramID     uint64
	MakerCode     string
	Version       uint16
	ProductCode   string
	MediaUnitSize int64
	Flags         [8]byte

	ExHeaderSize int64 // the *signed* ExHeader size field (0x400); the on-media
	// ExHeader is twice that — SCI+ACI followed by the access descriptor.

	PlainRegion region
	LogoRegion  region
	ExeFSRegion region
	RomFSRegion region
}

// ParseNCCH parses an NCCH partition header.
func ParseNCCH(part []byte) (*NCCH, error) {
	if len(part) < ncchHeaderOff+0x200 {
		return nil, fmt.Errorf("n3ds: partition too short for an NCCH header (%d bytes)", len(part))
	}
	h := part[ncchHeaderOff:]
	if string(h[0:4]) != ncchMagic {
		return nil, fmt.Errorf("n3ds: not an NCCH partition: magic %q, want %q", h[0:4], ncchMagic)
	}

	c := &NCCH{raw: part}
	copy(c.Flags[:], h[0x88:0x90])
	c.MediaUnitSize = 512 << c.Flags[flagContentUnit]

	c.ContentSize = int64(binary.LittleEndian.Uint32(h[0x04:])) * c.MediaUnitSize
	c.PartitionID = binary.LittleEndian.Uint64(h[0x08:])
	c.MakerCode = string(trimNul(h[0x10:0x12]))
	c.Version = binary.LittleEndian.Uint16(h[0x12:])
	c.ProgramID = binary.LittleEndian.Uint64(h[0x18:])
	c.ProductCode = string(trimNul(h[0x50:0x60]))
	c.ExHeaderSize = int64(binary.LittleEndian.Uint32(h[0x80:]))

	rd := func(off int) region {
		return region{
			Offset: int64(binary.LittleEndian.Uint32(h[off:])) * c.MediaUnitSize,
			Size:   int64(binary.LittleEndian.Uint32(h[off+4:])) * c.MediaUnitSize,
		}
	}
	c.PlainRegion = rd(0x90)
	c.LogoRegion = rd(0x98)
	c.ExeFSRegion = rd(0xA0)
	c.RomFSRegion = rd(0xB0)

	for _, r := range []struct {
		name string
		reg  region
	}{{"plain", c.PlainRegion}, {"logo", c.LogoRegion}, {"exefs", c.ExeFSRegion}, {"romfs", c.RomFSRegion}} {
		if !r.reg.empty() && r.reg.Offset+r.reg.Size > int64(len(part)) {
			return nil, fmt.Errorf("n3ds: %s region [0x%x+0x%x] runs past the partition end (0x%x)", r.name, r.reg.Offset, r.reg.Size, len(part))
		}
	}
	return c, nil
}

// Encrypted reports whether the container's contents are AES-CTR encrypted. A
// decrypted dump sets the NoCrypto bit in flags[7]; nothing in this package can
// decrypt an encrypted image, because the keys are console state, not cartridge
// data.
func (c *NCCH) Encrypted() bool { return c.Flags[flagOther]&otherNoCrypto == 0 }

// CryptoMethod names the key generation an encrypted container would need.
func (c *NCCH) CryptoMethod() string {
	switch c.Flags[flagCryptoMethod] {
	case 0x00:
		return "standard"
	case 0x01:
		return "7.x"
	case 0x0A:
		return "Secure3"
	case 0x0B:
		return "Secure4"
	}
	return fmt.Sprintf("unknown(0x%02x)", c.Flags[flagCryptoMethod])
}

// ContentType decodes the flags[5] bitmask into a human-readable list.
func (c *NCCH) ContentType() string {
	f := c.Flags[flagContentType]
	names := []struct {
		bit  byte
		name string
	}{
		{0x01, "Data"}, {0x02, "Executable"}, {0x04, "SystemUpdate"},
		{0x08, "Manual"}, {0x10, "Child"}, {0x20, "Trial"},
	}
	out := ""
	for _, n := range names {
		if f&n.bit != 0 {
			if out != "" {
				out += "|"
			}
			out += n.name
		}
	}
	if out == "" {
		return fmt.Sprintf("none(0x%02x)", f)
	}
	return out
}

// checkPlain guards every content accessor: reading an encrypted region would
// return ciphertext that parses as noise, which is worse than an error.
func (c *NCCH) checkPlain(what string) error {
	if c.Encrypted() {
		return fmt.Errorf("n3ds: %s is AES-CTR encrypted (crypto method %s); supply a decrypted dump", what, c.CryptoMethod())
	}
	return nil
}

// ExHeaderBytes returns the raw ExHeader (SCI+ACI, ExHeaderSize bytes), which
// immediately follows the 0x200-byte NCCH header.
func (c *NCCH) ExHeaderBytes() ([]byte, error) {
	if err := c.checkPlain("the ExHeader"); err != nil {
		return nil, err
	}
	if c.ExHeaderSize == 0 {
		return nil, fmt.Errorf("n3ds: partition has no ExHeader")
	}
	start := int64(ncchHeaderOff + 0x100)
	if start+c.ExHeaderSize > int64(len(c.raw)) {
		return nil, fmt.Errorf("n3ds: ExHeader runs past the partition end")
	}
	return c.raw[start : start+c.ExHeaderSize], nil
}

// ExHeader parses the ExHeader's system-control info.
func (c *NCCH) ExHeader() (*ExHeader, error) {
	b, err := c.ExHeaderBytes()
	if err != nil {
		return nil, err
	}
	return ParseExHeader(b)
}

// ExeFSBytes returns the raw ExeFS region.
func (c *NCCH) ExeFSBytes() ([]byte, error) {
	if err := c.checkPlain("the ExeFS"); err != nil {
		return nil, err
	}
	if c.ExeFSRegion.empty() {
		return nil, fmt.Errorf("n3ds: partition has no ExeFS")
	}
	r := c.ExeFSRegion
	return c.raw[r.Offset : r.Offset+r.Size], nil
}

// ExeFS parses the ExeFS region's file table.
func (c *NCCH) ExeFS() (*ExeFS, error) {
	b, err := c.ExeFSBytes()
	if err != nil {
		return nil, err
	}
	return ParseExeFS(b)
}

// RomFSBytes returns the raw RomFS region, IVFC hash tree and all.
func (c *NCCH) RomFSBytes() ([]byte, error) {
	if err := c.checkPlain("the RomFS"); err != nil {
		return nil, err
	}
	if c.RomFSRegion.empty() || c.Flags[flagOther]&otherNoRomFS != 0 {
		return nil, fmt.Errorf("n3ds: partition has no RomFS")
	}
	r := c.RomFSRegion
	return c.raw[r.Offset : r.Offset+r.Size], nil
}

// RomFS parses the RomFS region into a navigable filesystem.
func (c *NCCH) RomFS() (*RomFS, error) {
	b, err := c.RomFSBytes()
	if err != nil {
		return nil, err
	}
	return ParseRomFS(b)
}

func trimNul(b []byte) []byte {
	for i, v := range b {
		if v == 0 {
			return b[:i]
		}
	}
	return b
}
