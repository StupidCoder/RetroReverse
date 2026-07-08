package nds

import (
	"encoding/binary"
	"fmt"
)

// LZ77 is Nintendo's standard forward LZSS, used throughout the DS filesystem
// (compressed archives, graphics, tables) and distinct from the BLZ backward variant
// the boot code uses (blz.go). A compressed stream begins with a 4-byte header:
//
//	byte 0  : compression type — 0x10 (LZ10) or 0x11 (LZ11)
//	bytes1-3: decompressed size (24-bit little-endian)
//
// then flag-driven blocks: a flag byte whose 8 bits (MSB first) each select a literal
// byte (0) or a back-reference (1) into the already-output data.
//
// Type 0x10 references are two bytes: length = (hi nibble)+3, displacement =
// (remaining 12 bits)+1. Type 0x11 extends the length encoding to reach longer runs.

// IsLZ77 reports whether data begins with an LZ10/LZ11 header whose size field is
// plausible.
func IsLZ77(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	if data[0] != 0x10 && data[0] != 0x11 {
		return false
	}
	size := uint32(data[1]) | uint32(data[2])<<8 | uint32(data[3])<<16
	return size > 0 && size < 0x10000000
}

// DecompressLZ77 decompresses an LZ10/LZ11 stream.
func DecompressLZ77(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("nds: LZ77 data too short")
	}
	typ := data[0]
	if typ != 0x10 && typ != 0x11 {
		return nil, fmt.Errorf("nds: not an LZ77 stream (type 0x%02X)", typ)
	}
	size := int(binary.LittleEndian.Uint32(data) >> 8) // 24-bit size in bytes 1..3
	out := make([]byte, 0, size)
	p := 4
	next := func() (byte, error) {
		if p >= len(data) {
			return 0, fmt.Errorf("nds: LZ77 input truncated")
		}
		b := data[p]
		p++
		return b, nil
	}
	for len(out) < size {
		flags, err := next()
		if err != nil {
			return nil, err
		}
		for bit := 0; bit < 8 && len(out) < size; bit++ {
			if flags&0x80 == 0 { // literal
				b, err := next()
				if err != nil {
					return nil, err
				}
				out = append(out, b)
			} else { // back-reference
				b0, err := next()
				if err != nil {
					return nil, err
				}
				b1, err := next()
				if err != nil {
					return nil, err
				}
				var length, disp int
				if typ == 0x10 {
					length = int(b0>>4) + 3
					disp = (int(b0&0xF)<<8 | int(b1)) + 1
				} else { // 0x11: variable-length encoding by the top nibble
					switch b0 >> 4 {
					case 0: // 8-bit length: (b0.lo, b1.hi) + 0x11
						b2, err := next()
						if err != nil {
							return nil, err
						}
						length = (int(b0&0xF)<<4 | int(b1>>4)) + 0x11
						disp = (int(b1&0xF)<<8 | int(b2)) + 1
					case 1: // 16-bit length
						b2, err := next()
						if err != nil {
							return nil, err
						}
						b3, err := next()
						if err != nil {
							return nil, err
						}
						length = (int(b0&0xF)<<12 | int(b1)<<4 | int(b2>>4)) + 0x111
						disp = (int(b2&0xF)<<8 | int(b3)) + 1
					default: // 4-bit length in the top nibble
						length = int(b0>>4) + 1
						disp = (int(b0&0xF)<<8 | int(b1)) + 1
					}
				}
				start := len(out) - disp
				if start < 0 {
					return nil, fmt.Errorf("nds: LZ77 back-reference before start")
				}
				for i := 0; i < length && len(out) < size; i++ {
					out = append(out, out[start+i])
				}
			}
			flags <<= 1
		}
	}
	return out, nil
}

// Decompress transparently handles the common single-layer wrappers a DS filesystem
// file may carry: an LZ10/LZ11 stream (e.g. inside a .carc) or nothing. It returns
// the data unchanged when it is not recognisably compressed.
func Decompress(data []byte) []byte {
	if IsLZ77(data) {
		if out, err := DecompressLZ77(data); err == nil {
			return out
		}
	}
	return data
}
