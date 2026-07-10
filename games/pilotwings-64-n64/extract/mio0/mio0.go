// Package mio0 decompresses Pilotwings 64's compressed asset blobs.
//
// The format was derived by disassembling the game's own decompressor at
// 0x80231A20 (watch a texture region with bootoracle -watch, take the writing
// PC, disassemble back to the routine entry). The stream magic spells "MIO0"
// on the cartridge. Layout, all offsets relative to the MIO0 header:
//
//	+0  'MIO0'
//	+4  decompressed length
//	+8  offset of the back-reference stream (16-bit big-endian records)
//	+12 offset of the literal stream (raw bytes)
//	+16 flag bits, MSB-first in 32-bit words: 1 = copy one literal byte,
//	    0 = back-reference: len = (rec>>12)+3, dist = (rec&0xFFF)+1, copying
//	    len bytes from dist bytes behind the write cursor (the game's loop is
//	    `sub $t1,$a1,$t2` then `lb -1($t1)`, so the stored field is one less
//	    than the distance; runs may overlap themselves).
//
// On the cartridge each MIO0 image sits 8 bytes into an outer container:
// a fourCC ('COMM', 'TABL', ...) and a 32-bit payload size.
package mio0

import (
	"encoding/binary"
	"fmt"
)

// Decompress expands the MIO0 image at the start of src.
func Decompress(src []byte) ([]byte, error) {
	if len(src) < 16 || string(src[0:4]) != "MIO0" {
		return nil, fmt.Errorf("mio0: bad magic %q", src[0:4])
	}
	outLen := binary.BigEndian.Uint32(src[4:])
	refOff := binary.BigEndian.Uint32(src[8:])
	litOff := binary.BigEndian.Uint32(src[12:])
	out := make([]byte, 0, outLen)

	flagPos := uint32(16)
	var flags uint32
	bits := 0
	for uint32(len(out)) < outLen {
		if bits == 0 {
			if int(flagPos)+4 > len(src) {
				return nil, fmt.Errorf("mio0: flag word out of range at %d", flagPos)
			}
			flags = binary.BigEndian.Uint32(src[flagPos:])
			flagPos += 4
			bits = 32
		}
		if flags&0x80000000 != 0 {
			if int(litOff) >= len(src) {
				return nil, fmt.Errorf("mio0: literal out of range at %d", litOff)
			}
			out = append(out, src[litOff])
			litOff++
		} else {
			if int(refOff)+2 > len(src) {
				return nil, fmt.Errorf("mio0: back-reference out of range at %d", refOff)
			}
			rec := binary.BigEndian.Uint16(src[refOff:])
			refOff += 2
			length := int(rec>>12) + 3
			dist := int(rec&0xFFF) + 1
			if dist > len(out) {
				return nil, fmt.Errorf("mio0: back-reference distance %d exceeds output %d", dist, len(out))
			}
			for i := 0; i < length; i++ {
				out = append(out, out[len(out)-dist])
			}
		}
		flags <<= 1
		bits--
	}
	return out, nil
}
