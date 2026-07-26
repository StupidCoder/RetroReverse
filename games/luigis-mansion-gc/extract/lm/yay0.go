// Package lm decodes Luigi's Mansion's own file formats, as reverse-engineered
// from the game: the .szp archives the disc carries are Yay0 streams, and the
// decompressor reimplemented here is the game's own, read out of the DOL at
// 0x800071D0 (found by read-watching the buffer the drive had just filled).
package lm

import (
	"encoding/binary"
	"fmt"
)

// Yay0 decompresses a Yay0 stream the way the game's decompressor at 0x800071D0
// does.
//
// The header is four big-endian words: the magic "Yay0", the decompressed size,
// the offset of the link table, and the offset of the literal ("chunk") bytes.
// The mask stream starts right after the header at +0x10: each mask word gives
// 32 flags, MSB first. A set bit copies one literal byte from the chunk stream;
// a clear bit reads one 16-bit link from the link table — count in the top
// nibble, distance in the low 12 bits — and copies count bytes from
// dst[pos-dist-1]. A zero count nibble means the real count is the next chunk
// byte plus 18; otherwise the count is the nibble plus 2.
func Yay0(src []byte) ([]byte, error) {
	if len(src) < 16 || string(src[:4]) != "Yay0" {
		return nil, fmt.Errorf("lm: not a Yay0 stream")
	}
	size := binary.BigEndian.Uint32(src[4:])
	link := binary.BigEndian.Uint32(src[8:])
	chunk := binary.BigEndian.Uint32(src[12:])

	dst := make([]byte, 0, size)
	var mask uint32
	var bits int
	maskPos := uint32(16)
	for uint32(len(dst)) < size {
		if bits == 0 {
			if maskPos+4 > uint32(len(src)) {
				return nil, fmt.Errorf("lm: Yay0 mask stream truncated at %d/%d", len(dst), size)
			}
			mask = binary.BigEndian.Uint32(src[maskPos:])
			maskPos += 4
			bits = 32
		}
		if mask&0x80000000 != 0 {
			if chunk >= uint32(len(src)) {
				return nil, fmt.Errorf("lm: Yay0 chunk stream truncated at %d/%d", len(dst), size)
			}
			dst = append(dst, src[chunk])
			chunk++
		} else {
			if link+2 > uint32(len(src)) {
				return nil, fmt.Errorf("lm: Yay0 link table truncated at %d/%d", len(dst), size)
			}
			entry := binary.BigEndian.Uint16(src[link:])
			link += 2
			count := uint32(entry >> 12)
			if count == 0 {
				if chunk >= uint32(len(src)) {
					return nil, fmt.Errorf("lm: Yay0 chunk stream truncated at %d/%d", len(dst), size)
				}
				count = uint32(src[chunk]) + 18
				chunk++
			} else {
				count += 2
			}
			dist := uint32(entry&0xFFF) + 1
			if dist > uint32(len(dst)) {
				return nil, fmt.Errorf("lm: Yay0 back-reference past the start at %d/%d", len(dst), size)
			}
			for i := uint32(0); i < count; i++ {
				dst = append(dst, dst[uint32(len(dst))-dist])
			}
		}
		mask <<= 1
		bits--
	}
	return dst[:size], nil
}
