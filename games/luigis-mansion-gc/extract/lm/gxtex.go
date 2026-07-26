package lm

// gxtex.go decodes the GX texture formats the .mdl files use. The file's
// one-byte format index maps through the game's own table (at 0x80338C30 in
// the DOL) to a GX format:
//
//	0 CI4   1 CI8   2 CI14X2   3 I4   4 I8   5 IA4   6 IA8
//	7 RGB565   8 RGB5A3   9 RGBA8   10 CMPR
//
// All GX formats are stored in tiles (8x8, 8x4 or 4x4 depending on texel
// size); decoding walks tiles in row-major order. The colour math (RGB565,
// RGB5A3, CMPR's 2-bit blends) is the same as the platform's LLE texture unit.

import (
	"fmt"
	"math"
)

func float32frombits(u uint32) float32 { return math.Float32frombits(u) }

// decodeGXTexture decodes an image to RGBA bytes.
func decodeGXTexture(fmtIdx uint8, w, h int, data []byte) ([]byte, error) {
	out := make([]byte, w*h*4)
	set := func(x, y int, r, g, b, a uint8) {
		if x >= w || y >= h {
			return
		}
		o := (y*w + x) * 4
		out[o], out[o+1], out[o+2], out[o+3] = r, g, b, a
	}
	switch fmtIdx {
	case 3: // I4: 8x8 tiles, 4 bits per texel
		tw, th := (w+7)/8, (h+7)/8
		for ty := 0; ty < th; ty++ {
			for tx := 0; tx < tw; tx++ {
				base := (ty*tw + tx) * 32
				for py := 0; py < 8; py++ {
					for px := 0; px < 8; px++ {
						v := data[base+py*4+px/2]
						if px%2 == 0 {
							v >>= 4
						}
						v &= 0xF
						i := v<<4 | v
						set(tx*8+px, ty*8+py, i, i, i, i)
					}
				}
			}
		}
	case 4: // I8: 8x4 tiles
		tw, th := (w+7)/8, (h+3)/4
		for ty := 0; ty < th; ty++ {
			for tx := 0; tx < tw; tx++ {
				base := (ty*tw + tx) * 32
				for py := 0; py < 4; py++ {
					for px := 0; px < 8; px++ {
						i := data[base+py*8+px]
						set(tx*8+px, ty*4+py, i, i, i, i)
					}
				}
			}
		}
	case 5: // IA4: 8x4 tiles
		tw, th := (w+7)/8, (h+3)/4
		for ty := 0; ty < th; ty++ {
			for tx := 0; tx < tw; tx++ {
				base := (ty*tw + tx) * 32
				for py := 0; py < 4; py++ {
					for px := 0; px < 8; px++ {
						v := data[base+py*8+px]
						i := v & 0xF
						i = i<<4 | i
						a := v >> 4
						a = a<<4 | a
						set(tx*8+px, ty*4+py, i, i, i, a)
					}
				}
			}
		}
	case 6: // IA8: 4x4 tiles, 2 bytes per texel
		tw, th := (w+3)/4, (h+3)/4
		for ty := 0; ty < th; ty++ {
			for tx := 0; tx < tw; tx++ {
				base := (ty*tw + tx) * 32
				for py := 0; py < 4; py++ {
					for px := 0; px < 4; px++ {
						o := base + (py*4+px)*2
						a, i := data[o], data[o+1]
						set(tx*4+px, ty*4+py, i, i, i, a)
					}
				}
			}
		}
	case 7: // RGB565: 4x4 tiles
		tw, th := (w+3)/4, (h+3)/4
		for ty := 0; ty < th; ty++ {
			for tx := 0; tx < tw; tx++ {
				base := (ty*tw + tx) * 32
				for py := 0; py < 4; py++ {
					for px := 0; px < 4; px++ {
						o := base + (py*4+px)*2
						v := uint16(data[o])<<8 | uint16(data[o+1])
						r, g, b := rgb565(v)
						set(tx*4+px, ty*4+py, r, g, b, 255)
					}
				}
			}
		}
	case 8: // RGB5A3: 4x4 tiles
		tw, th := (w+3)/4, (h+3)/4
		for ty := 0; ty < th; ty++ {
			for tx := 0; tx < tw; tx++ {
				base := (ty*tw + tx) * 32
				for py := 0; py < 4; py++ {
					for px := 0; px < 4; px++ {
						o := base + (py*4+px)*2
						v := uint16(data[o])<<8 | uint16(data[o+1])
						r, g, b, a := rgb5a3(v)
						set(tx*4+px, ty*4+py, r, g, b, a)
					}
				}
			}
		}
	case 9: // RGBA8: 4x4 tiles, two 32-byte halves (AR then GB)
		tw, th := (w+3)/4, (h+3)/4
		for ty := 0; ty < th; ty++ {
			for tx := 0; tx < tw; tx++ {
				base := (ty*tw + tx) * 64
				for py := 0; py < 4; py++ {
					for px := 0; px < 4; px++ {
						o := base + (py*4+px)*2
						a, r := data[o], data[o+1]
						g, b := data[o+32], data[o+33]
						set(tx*4+px, ty*4+py, r, g, b, a)
					}
				}
			}
		}
	case 10: // CMPR: 8x8 tiles of four 4x4 DXT1 blocks
		tw, th := (w+7)/8, (h+7)/8
		for ty := 0; ty < th; ty++ {
			for tx := 0; tx < tw; tx++ {
				base := (ty*tw + tx) * 32
				for sub := 0; sub < 4; sub++ {
					bo := base + sub*8
					ox, oy := tx*8+(sub%2)*4, ty*8+(sub/2)*4
					c0 := uint16(data[bo])<<8 | uint16(data[bo+1])
					c1 := uint16(data[bo+2])<<8 | uint16(data[bo+3])
					var pal [4][4]uint8
					r0, g0, b0 := rgb565(c0)
					r1, g1, b1 := rgb565(c1)
					pal[0] = [4]uint8{r0, g0, b0, 255}
					pal[1] = [4]uint8{r1, g1, b1, 255}
					if c0 > c1 {
						pal[2] = [4]uint8{lerp2(r0, r1), lerp2(g0, g1), lerp2(b0, b1), 255}
						pal[3] = [4]uint8{lerp2(r1, r0), lerp2(g1, g0), lerp2(b1, b0), 255}
					} else {
						pal[2] = [4]uint8{avg2(r0, r1), avg2(g0, g1), avg2(b0, b1), 255}
						pal[3] = [4]uint8{0, 0, 0, 0}
					}
					for py := 0; py < 4; py++ {
						row := data[bo+4+py]
						for px := 0; px < 4; px++ {
							c := pal[(row>>uint((3-px)*2))&3]
							set(ox+px, oy+py, c[0], c[1], c[2], c[3])
						}
					}
				}
			}
		}
	default:
		return nil, fmt.Errorf("texture format index %d not yet needed/implemented", fmtIdx)
	}
	return out, nil
}

func rgb565(v uint16) (r, g, b uint8) {
	r = uint8(v >> 11 & 0x1F)
	g = uint8(v >> 5 & 0x3F)
	b = uint8(v & 0x1F)
	return r<<3 | r>>2, g<<2 | g>>4, b<<3 | b>>2
}

func rgb5a3(v uint16) (r, g, b, a uint8) {
	if v&0x8000 != 0 {
		r = uint8(v >> 10 & 0x1F)
		g = uint8(v >> 5 & 0x1F)
		b = uint8(v & 0x1F)
		return r<<3 | r>>2, g<<3 | g>>2, b<<3 | b>>2, 255
	}
	a = uint8(v >> 12 & 0x7)
	r = uint8(v >> 8 & 0xF)
	g = uint8(v >> 4 & 0xF)
	b = uint8(v & 0xF)
	return r<<4 | r, g<<4 | g, b<<4 | b, a<<5 | a<<2 | a>>1
}

func lerp2(a, b uint8) uint8 { return uint8((uint16(a)*2 + uint16(b)) / 3) }
func avg2(a, b uint8) uint8  { return uint8((uint16(a) + uint16(b)) / 2) }
