package n3ds

import (
	"encoding/binary"
	"fmt"
)

// The ExeFS "banner" file is a CBMD ("Common Banner Model Data") container: the
// little animated 3-D scene the HOME Menu shows for a title. It holds one common
// CGFX model and up to 13 language-specific overrides, plus an optional CWAV
// sound. Each CGFX is LZ11-compressed (lz11.go).
//
// Layout:
//
//	0x00 "CBMD"
//	0x04 u32   version (0)
//	0x08 u32   offset to the common CGFX (LZ11-compressed)
//	0x0C u32[13] offsets to language-specific CGFX (0 if absent)
//	0x40 u32   offset to the CWAV audio (0 if absent)
//	0x44 …     padding to 0x88, where the first CGFX begins
const (
	cbmdMagic       = "CBMD"
	cbmdNumLangs    = 13
	cbmdCommonOff   = 0x08
	cbmdLangOff     = 0x0C
	cbmdCWAVOff     = 0x40
)

// Banner is a parsed CBMD container.
type Banner struct {
	raw       []byte
	CommonOff uint32
	LangOff   [cbmdNumLangs]uint32
	CWAVOff   uint32
}

// ParseBanner parses the CBMD header of an ExeFS banner file.
func ParseBanner(b []byte) (*Banner, error) {
	if len(b) < 0x88 {
		return nil, fmt.Errorf("banner: too short for a CBMD header (%d bytes)", len(b))
	}
	if string(b[0:4]) != cbmdMagic {
		return nil, fmt.Errorf("banner: bad magic %q, want %q", b[0:4], cbmdMagic)
	}
	bn := &Banner{raw: b}
	bn.CommonOff = binary.LittleEndian.Uint32(b[cbmdCommonOff:])
	for i := 0; i < cbmdNumLangs; i++ {
		bn.LangOff[i] = binary.LittleEndian.Uint32(b[cbmdLangOff+i*4:])
	}
	bn.CWAVOff = binary.LittleEndian.Uint32(b[cbmdCWAVOff:])
	if bn.CommonOff == 0 || bn.CommonOff >= uint32(len(b)) {
		return nil, fmt.Errorf("banner: common CGFX offset 0x%x out of range", bn.CommonOff)
	}
	return bn, nil
}

// modelBound returns the end offset of the CGFX blob starting at off: the next
// non-zero section offset above it, or the file end (its CWAV/EOF).
func (bn *Banner) modelBound(off uint32) uint32 {
	end := uint32(len(bn.raw))
	bounds := append([]uint32{bn.CWAVOff}, bn.LangOff[:]...)
	for _, o := range bounds {
		if o > off && o < end {
			end = o
		}
	}
	return end
}

// CommonModel returns the decompressed common-banner CGFX bytes.
func (bn *Banner) CommonModel() ([]byte, error) {
	end := bn.modelBound(bn.CommonOff)
	cgfx, err := DecompressLZ11(bn.raw[bn.CommonOff:end])
	if err != nil {
		return nil, fmt.Errorf("banner: decompressing common CGFX: %w", err)
	}
	if len(cgfx) < 4 || string(cgfx[0:4]) != "CGFX" {
		return nil, fmt.Errorf("banner: decompressed common model is not a CGFX (starts %q)", cgfx[:min(4, len(cgfx))])
	}
	return cgfx, nil
}
