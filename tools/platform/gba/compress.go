package gba

// The BIOS compression formats, decoded from a plain byte slice.
//
// These are the same algorithms gbamachine's BIOS HLE implements for the SWIs
// (bios.go), and they are deliberately a second implementation for the same
// reason tiles.go is: that one decodes THROUGH A BUS, into a live machine's
// memory, with the GBA's byte-store quirks in the way; this one is a pure
// function over bytes, which is what an offline asset extractor needs. A test
// runs both over the same streams and requires identical output.
//
// Every stream starts with a 4-byte header: a type nibble, a 4-bit reserved
// nibble, and a 24-bit decompressed size.
//
//	byte 0: bits 7:4 = type (1 = LZ77, 2 = Huffman, 3 = RLE), bits 3:0 = reserved
//	bytes 1-3: decompressed size, little-endian

import "fmt"

// CompHeader reports a compression stream's type nibble and decompressed size.
// (Named for the stream, not just "Header", because the package already has a
// cartridge Header type.)
func CompHeader(src []byte) (typ int, size int, err error) {
	if len(src) < 4 {
		return 0, 0, fmt.Errorf("gba: compression header needs 4 bytes, got %d", len(src))
	}
	return int(src[0] >> 4), int(src[1]) | int(src[2])<<8 | int(src[3])<<16, nil
}

// LZ77Decompress decodes a BIOS LZ77 stream (type 1). src may extend past the
// end of the stream — the header's size decides when to stop.
//
// The format is a flag byte followed by eight tokens: a clear flag bit means a
// literal byte, a set bit means a two-byte back-reference of (length 3..18,
// displacement 1..4096) copied from the ALREADY DECOMPRESSED output. Copying
// byte by byte matters: a reference may be longer than its displacement, which
// is how the format encodes a run, and a block copy would read bytes that have
// not been written yet.
func LZ77Decompress(src []byte) ([]byte, error) {
	typ, size, err := CompHeader(src)
	if err != nil {
		return nil, err
	}
	if typ != 1 {
		return nil, fmt.Errorf("gba: stream is type %d, not LZ77 (1)", typ)
	}
	out := make([]byte, 0, size)
	p := 4
	for len(out) < size {
		if p >= len(src) {
			return nil, fmt.Errorf("gba: LZ77 stream truncated at %d/%d bytes out", len(out), size)
		}
		flags := src[p]
		p++
		for bit := 7; bit >= 0 && len(out) < size; bit-- {
			if flags>>uint(bit)&1 == 0 {
				if p >= len(src) {
					return nil, fmt.Errorf("gba: LZ77 literal past end of input")
				}
				out = append(out, src[p])
				p++
				continue
			}
			if p+1 >= len(src) {
				return nil, fmt.Errorf("gba: LZ77 reference past end of input")
			}
			b1, b2 := src[p], src[p+1]
			p += 2
			length := int(b1>>4) + 3
			disp := int(b1&0xF)<<8 | int(b2) + 1
			if disp > len(out) {
				return nil, fmt.Errorf("gba: LZ77 back-reference of %d with only %d bytes out", disp, len(out))
			}
			for i := 0; i < length && len(out) < size; i++ {
				out = append(out, out[len(out)-disp])
			}
		}
	}
	return out, nil
}

// RLEDecompress decodes a BIOS run-length stream (type 3).
func RLEDecompress(src []byte) ([]byte, error) {
	typ, size, err := CompHeader(src)
	if err != nil {
		return nil, err
	}
	if typ != 3 {
		return nil, fmt.Errorf("gba: stream is type %d, not RLE (3)", typ)
	}
	out := make([]byte, 0, size)
	p := 4
	for len(out) < size {
		if p >= len(src) {
			return nil, fmt.Errorf("gba: RLE stream truncated at %d/%d bytes out", len(out), size)
		}
		f := src[p]
		p++
		if f&0x80 != 0 { // a run
			if p >= len(src) {
				return nil, fmt.Errorf("gba: RLE run past end of input")
			}
			v := src[p]
			p++
			for i := 0; i < int(f&0x7F)+3 && len(out) < size; i++ {
				out = append(out, v)
			}
			continue
		}
		for i := 0; i < int(f&0x7F)+1 && len(out) < size; i++ {
			if p >= len(src) {
				return nil, fmt.Errorf("gba: RLE literals past end of input")
			}
			out = append(out, src[p])
			p++
		}
	}
	return out, nil
}

// ROMOffset converts a cartridge bus address (0x08000000-based, or one of the
// waitstate mirrors) to a file offset.
func ROMOffset(addr uint32) int { return int(addr & 0x01FFFFFF) }
