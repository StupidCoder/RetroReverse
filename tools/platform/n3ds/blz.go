package n3ds

import (
	"encoding/binary"
	"fmt"
)

// BLZ — the "backward LZ77" Nintendo applies to ExeFS/.code (and, on the DS, to
// the ARM9 binary). It is an ordinary LZSS with 3-byte minimum matches, run
// *backwards*: the stream is decoded from the last byte of the file toward the
// first, writing the output from its end toward its beginning.
//
// The reason for that inversion is in-place decompression. The compressed file
// is copied to the front of a buffer sized for the *decompressed* output, and
// the decoder then reads backwards from the end of the compressed bytes while
// writing backwards from the end of the buffer. The write pointer always stays
// ahead of the read pointer, so the two never collide and the loader needs no
// second buffer. It also means a prefix of the file — the part that compressed
// badly — is simply stored verbatim and never touched by the decoder.
//
// Layout, reading from the end of the file:
//
//	[ ... verbatim head ... | ... compressed stream ... | footer ]
//	                                                      ^ 8 bytes, +0..3 padding
//
//	footer[0:4] = bufferTopAndBottom
//	              bits 0..23  compressedSize — the length of the trailing region
//	                          made up of the compressed stream plus the footer
//	              bits 24..31 footerSize — 8, plus up to 3 bytes of alignment
//	footer[4:8] = originalBottom — how many bytes to add to the file's length to
//	              get the decompressed length
//
// A flag byte precedes (in backward order: follows, in memory) each group of
// eight tokens, MSB first. A clear bit is a literal; a set bit is a 16-bit
// match token whose top nibble is (length-3) and whose low 12 bits are
// (distance-3). Distances count *forward* from the write pointer, since output
// grows downward.
const (
	blzFooterSize   = 8
	blzMinFooterPad = 0 // footerSize is 8..11: eight bytes plus 0..3 alignment
	blzMaxFooterPad = 3
	blzThreshold    = 2 // minimum encodable match is blzThreshold+1 = 3 bytes
)

// DecompressBLZ decodes a BLZ-compressed buffer. A file whose originalBottom is
// zero was stored rather than compressed; its bytes are returned unchanged.
func DecompressBLZ(pak []byte) ([]byte, error) {
	if len(pak) < blzFooterSize {
		return nil, fmt.Errorf("blz: input too short for a footer (%d bytes)", len(pak))
	}
	n := len(pak)
	bufferTopAndBottom := binary.LittleEndian.Uint32(pak[n-8:])
	originalBottom := binary.LittleEndian.Uint32(pak[n-4:])

	if originalBottom == 0 {
		out := make([]byte, n)
		copy(out, pak)
		return out, nil
	}

	compressedSize := int(bufferTopAndBottom & 0xFFFFFF)
	footerSize := int(bufferTopAndBottom >> 24)

	if footerSize < blzFooterSize+blzMinFooterPad || footerSize > blzFooterSize+blzMaxFooterPad {
		return nil, fmt.Errorf("blz: footer size %d out of range 8..11 (bufferTopAndBottom=0x%08x)", footerSize, bufferTopAndBottom)
	}
	if compressedSize < footerSize || compressedSize > n {
		return nil, fmt.Errorf("blz: compressed size 0x%x out of range [0x%x, 0x%x]", compressedSize, footerSize, n)
	}

	headSize := n - compressedSize    // bytes stored verbatim at the front
	rawLen := n + int(originalBottom) // total decompressed length

	out := make([]byte, rawLen)
	copy(out, pak)

	// Read backwards from just before the footer; write backwards from the end.
	// Both pointers stop at headSize, where the verbatim prefix begins.
	src := n - footerSize
	dst := rawLen
	end := headSize

	var flags byte
	mask := byte(0)

	for dst > end {
		if mask >>= 1; mask == 0 {
			if src <= end {
				break
			}
			src--
			flags = pak[src]
			mask = 0x80
		}

		if flags&mask == 0 { // literal
			if src <= end {
				return nil, fmt.Errorf("blz: literal token underruns the compressed stream at src=0x%x", src)
			}
			src--
			dst--
			out[dst] = pak[src]
			continue
		}

		// Match token: two bytes, high byte first in backward order.
		if src-2 < end {
			return nil, fmt.Errorf("blz: match token underruns the compressed stream at src=0x%x", src)
		}
		src--
		tok := uint32(pak[src]) << 8
		src--
		tok |= uint32(pak[src])

		length := int(tok>>12) + blzThreshold + 1
		dist := int(tok&0xFFF) + 3

		// The final match may overrun the head boundary; the encoder relies on
		// the decoder clamping it rather than emitting a shorter token.
		if dst-end < length {
			length = dst - end
		}
		if dst+dist > rawLen {
			return nil, fmt.Errorf("blz: match distance %d at dst=0x%x reads past the buffer end", dist, dst)
		}
		for ; length > 0; length-- {
			dst--
			out[dst] = out[dst+dist]
		}
	}

	if dst != end {
		return nil, fmt.Errorf("blz: stream ended with 0x%x bytes still to write (dst=0x%x, head ends at 0x%x)", dst-end, dst, end)
	}
	return out, nil
}
