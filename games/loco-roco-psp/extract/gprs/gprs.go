// Package gprs decompresses Loco Roco's GPRS containers — the compression that
// wraps every packed archive in DATA.BIN (the GARC directories, the XUI screens).
//
// The format is a byte-aligned LZSS variant driven by an MSB-first bit stream.
// An 8-byte header (magic "GPRS", big-endian decompressed size) precedes the
// stream. The stream interleaves control bits with whole bytes:
//
//	bit 0             → literal: copy one byte from the stream
//	bits 1,0          → short match: one offset byte D (D=0 terminates the
//	                    stream), source = dst - (256-D), i.e. 1..255 back
//	bits 1,1          → long match: one offset byte D and four more bits N;
//	                    source = dst + ((D-256)<<4 | N) - 255, i.e. 256..4351 back
//
// A match's length is interleaved Elias-gamma: len = 1, then while the next
// bit is 1, len = len<<1 | following bit; len+1 bytes are copied byte-by-byte
// (overlap allowed). The game's decompressor is at 0x08846EDC in the loaded
// module (reached from the archive loader 0x08846A8), with the copy loop
// unrolled eight-wide through a jump table; this reimplementation preserves
// its byte-exact semantics.
package gprs

import (
	"encoding/binary"
	"fmt"
)

// HeaderSize is the size of the GPRS container header.
const HeaderSize = 8

// Size returns the decompressed size declared by a GPRS header.
func Size(data []byte) (int, error) {
	if len(data) < HeaderSize || string(data[:4]) != "GPRS" {
		return 0, fmt.Errorf("gprs: bad header")
	}
	return int(binary.BigEndian.Uint32(data[4:8])), nil
}

// Decompress inflates a whole GPRS container (header + stream).
func Decompress(data []byte) ([]byte, error) {
	n, err := Size(data)
	if err != nil {
		return nil, err
	}
	return decode(data[HeaderSize:], n)
}

// decode runs the LZSS bit stream into a buffer of the declared size.
func decode(src []byte, dstSize int) ([]byte, error) {
	dst := make([]byte, 0, dstSize)
	pos := 0
	var cur byte
	mask := byte(0)

	readByte := func() (byte, error) {
		if pos >= len(src) {
			return 0, fmt.Errorf("gprs: stream truncated at %d", pos)
		}
		b := src[pos]
		pos++
		return b, nil
	}
	readBit := func() (int, error) {
		if mask == 0 {
			b, err := readByte()
			if err != nil {
				return 0, err
			}
			cur, mask = b, 0x80
		}
		bit := 0
		if cur&mask != 0 {
			bit = 1
		}
		mask >>= 1
		return bit, nil
	}

	// The first stream byte primes the bit reader.
	b0, err := readByte()
	if err != nil {
		return nil, err
	}
	cur, mask = b0, 0x80

	for {
		bit, err := readBit()
		if err != nil {
			return nil, err
		}
		if bit == 0 { // literal
			b, err := readByte()
			if err != nil {
				return nil, err
			}
			dst = append(dst, b)
			continue
		}
		long, err := readBit()
		if err != nil {
			return nil, err
		}
		var off int
		if long == 1 {
			d, err := readByte()
			if err != nil {
				return nil, err
			}
			off = int(d) - 256
			for i := 0; i < 4; i++ {
				b, err := readBit()
				if err != nil {
					return nil, err
				}
				off = off<<1 | b
			}
			off -= 255
		} else {
			d, err := readByte()
			if err != nil {
				return nil, err
			}
			if d == 0 { // terminator
				break
			}
			off = int(d) - 256
		}
		length := 1
		for {
			b, err := readBit()
			if err != nil {
				return nil, err
			}
			if b == 0 {
				break
			}
			b2, err := readBit()
			if err != nil {
				return nil, err
			}
			length = length<<1 | b2
		}
		from := len(dst) + off
		if from < 0 {
			return nil, fmt.Errorf("gprs: match reaches before output start (off %d at %d)", off, len(dst))
		}
		for i := 0; i <= length; i++ { // length+1 bytes, overlap allowed
			dst = append(dst, dst[from+i])
		}
	}
	if len(dst) != dstSize {
		return nil, fmt.Errorf("gprs: decompressed %d bytes, header declares %d", len(dst), dstSize)
	}
	return dst, nil
}
