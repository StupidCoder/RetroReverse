package n3ds

import (
	"encoding/binary"
	"fmt"
)

// LZ11 is Nintendo's forward LZSS variant (compression type 0x11), used across
// the 3DS for CGFX/asset payloads — here, the banner's model. It is a *forward*
// dictionary coder (unlike ExeFS/.code's backward BLZ): decode from the first
// byte to the last, emitting literals and copies from earlier output.
//
// Header: one type byte (0x11) then a 24-bit little-endian decompressed size; if
// that size is zero, a 32-bit size follows instead. Then the body: a flag byte
// whose eight bits (MSB first) each say whether the next token is a literal byte
// (0) or a back-reference (1). LZ11's back-references have three length ranges,
// selected by the top nibble of the first token byte, so a single copy can span
// up to 0x10110 bytes — the extension over plain LZ10 that makes it "11".
func DecompressLZ11(src []byte) ([]byte, error) {
	if len(src) < 4 {
		return nil, fmt.Errorf("lz11: input too short (%d bytes)", len(src))
	}
	if src[0] != 0x11 {
		return nil, fmt.Errorf("lz11: bad type byte 0x%02x, want 0x11", src[0])
	}
	size := uint32(src[1]) | uint32(src[2])<<8 | uint32(src[3])<<16
	pos := 4
	if size == 0 { // extended 32-bit size
		if len(src) < 8 {
			return nil, fmt.Errorf("lz11: truncated extended size header")
		}
		size = binary.LittleEndian.Uint32(src[4:])
		pos = 8
	}

	out := make([]byte, 0, size)
	readByte := func() (byte, error) {
		if pos >= len(src) {
			return 0, fmt.Errorf("lz11: input exhausted at output %d/%d", len(out), size)
		}
		b := src[pos]
		pos++
		return b, nil
	}

	for uint32(len(out)) < size {
		flags, err := readByte()
		if err != nil {
			return nil, err
		}
		for bit := 7; bit >= 0 && uint32(len(out)) < size; bit-- {
			if flags&(1<<uint(bit)) == 0 { // literal
				b, err := readByte()
				if err != nil {
					return nil, err
				}
				out = append(out, b)
				continue
			}

			// Back-reference. The first byte's top nibble selects the length
			// encoding; the remaining bits, with following bytes, give the copy
			// length and the displacement.
			b0, err := readByte()
			if err != nil {
				return nil, err
			}
			indicator := b0 >> 4

			var count int
			var dispHi byte
			switch indicator {
			case 0: // 8-bit length: count in [0x11, 0x110]
				b1, err := readByte()
				if err != nil {
					return nil, err
				}
				count = int(b0&0xF)<<4 | int(b1>>4)
				count += 0x11
				dispHi = b1
			case 1: // 16-bit length: count in [0x111, 0x10110]
				b1, err := readByte()
				if err != nil {
					return nil, err
				}
				b2, err := readByte()
				if err != nil {
					return nil, err
				}
				count = int(b0&0xF)<<12 | int(b1)<<4 | int(b2>>4)
				count += 0x111
				dispHi = b2
			default: // 4-bit length: count in [1, 16]
				count = int(indicator) + 1
				dispHi = b0
			}

			b, err := readByte()
			if err != nil {
				return nil, err
			}
			disp := int(dispHi&0xF)<<8 | int(b)
			disp++ // stored displacement is one less than the true distance

			if disp > len(out) {
				return nil, fmt.Errorf("lz11: displacement %d exceeds output %d", disp, len(out))
			}
			// Copy byte by byte: overlapping copies (disp < count) are legal and
			// produce run-length expansion.
			start := len(out) - disp
			for i := 0; i < count && uint32(len(out)) < size; i++ {
				out = append(out, out[start+i])
			}
		}
	}
	return out, nil
}
