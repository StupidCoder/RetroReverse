package nds

import "encoding/binary"

// BLZ (backward LZSS) is the compression Nintendo's SDK applies to the ARM9 static
// module and to code overlays: an LZSS variant encoded so that it decompresses
// *backward*, in place, growing the image upward from its stored (compressed) size to
// its runtime size. The compressed image carries an 8–12 byte footer:
//
//	[end-8] u32: (bits 0..23) enc_len — length of the compressed region incl. footer;
//	             (bits 24..31) hdr_len — footer size (8..12)
//	[end-4] u32: inc_len — how much larger the decompressed data is
//
// The bytes below the compressed region ([0, len-enc_len)) are stored verbatim (the
// ARM9's uncompressed secure-area/startup stub); the compressed stream sits above it.
//
// Implementation follows the standard trick (CUE's blz): reverse the compressed
// bytes, run an ordinary forward LZSS, then reverse the produced bytes back — which
// realises the backward decode without fragile end-to-start pointer arithmetic.

// IsBLZ reports whether data looks BLZ-compressed (a non-zero inc_len and a plausible
// footer size). It is a heuristic on the footer, not a guarantee.
func IsBLZ(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	inc := binary.LittleEndian.Uint32(data[len(data)-4:])
	if inc == 0 {
		return false
	}
	hdr := int(data[len(data)-5])
	enc := int(binary.LittleEndian.Uint32(data[len(data)-8:]) & 0x00FFFFFF)
	return hdr >= 8 && hdr <= 12 && enc >= hdr && enc <= len(data)
}

// DecompressBLZ decompresses a BLZ-compressed image. If the footer indicates the data
// is not compressed (inc_len == 0) it returns a copy unchanged.
func DecompressBLZ(data []byte) []byte {
	n := len(data)
	le := binary.LittleEndian
	incLen := le.Uint32(data[n-4:])
	if incLen == 0 {
		return append([]byte(nil), data...)
	}
	hdrLen := int(data[n-5])
	encLen := int(le.Uint32(data[n-8:]) & 0x00FFFFFF)
	decLen := n - encLen      // verbatim bottom (uncompressed stub)
	pakLen := encLen - hdrLen // pure compressed stream, footer excluded
	rawLen := n + int(incLen) // final decompressed length

	out := make([]byte, rawLen)
	copy(out[:decLen], data[:decLen])

	// Reverse the compressed stream so it can be decoded forward.
	comp := make([]byte, pakLen)
	copy(comp, data[decLen:decLen+pakLen])
	reverse(comp)

	ip, op := 0, decLen
	var flags, mask byte
	for op < rawLen && ip < len(comp) {
		if mask == 0 {
			flags = comp[ip]
			ip++
			mask = 0x80
		}
		if flags&mask == 0 { // literal
			out[op] = comp[ip]
			ip++
			op++
		} else { // back-reference
			b1 := comp[ip]
			b2 := comp[ip+1]
			ip += 2
			length := int(b1>>4) + 3
			disp := (int(b1&0xF)<<8 | int(b2)) + 3
			for k := 0; k < length && op < rawLen; k++ {
				out[op] = out[op-disp]
				op++
			}
		}
		mask >>= 1
	}
	// Reverse the produced region back into forward order.
	reverseRange(out, decLen, rawLen)
	return out
}

func reverse(b []byte) { reverseRange(b, 0, len(b)) }

func reverseRange(b []byte, lo, hi int) {
	for i, j := lo, hi-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}
