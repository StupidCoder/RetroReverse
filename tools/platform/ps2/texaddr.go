package ps2

// Exported GS texel addressing for offline asset decoders (game extract
// modules decode texture data laid out in GS memory formats without running
// the machine). These are the sampler's own address swizzles (gs.go /
// gstex.go); the wrappers exist so the arithmetic lives in exactly one place.

// TexAddrPSMCT32 returns the byte offset of texel (x, y) in a 32-bit buffer
// at block bp with buffer width bw (units of 64 texels).
func TexAddrPSMCT32(bp, bw, x, y uint32) uint32 { return addrPSMCT32(bp, bw, x, y) }

// TexAddrPSMCT16 is the 16-bit swizzle; s selects the CT16S block table.
func TexAddrPSMCT16(bp, bw, x, y uint32, s bool) uint32 { return addrPSMCT16(bp, bw, x, y, s) }

// TexAddrPSMT8 is the 8-bit index swizzle.
func TexAddrPSMT8(bp, bw, x, y uint32) uint32 { return addrPSMT8(bp, bw, x, y) }

// TexAddrPSMT4 is the 4-bit index swizzle: byte offset plus nibble (0 low).
func TexAddrPSMT4(bp, bw, x, y uint32) (uint32, uint32) { return addrPSMT4(bp, bw, x, y) }

// CSM1ClutXY returns the (x, y) position of CLUT entry i in the CSM1
// arrangement of an n-entry table (256 entries fill a swizzled 16x16 patch,
// 16 entries an 8x2 one).
func CSM1ClutXY(i, n uint32) (x, y uint32) {
	if n == 256 {
		return i&7 | i&0x10>>1, i >> 3 & 1 | i >> 4 & 0xE
	}
	return i & 7, i >> 3 & 1
}
