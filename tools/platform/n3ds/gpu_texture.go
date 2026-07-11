package n3ds

// gpu_texture.go samples the PICA200 texture units. Textures live in VRAM (or
// FCRAM) in the same 8×8-Morton-tiled formats the CGFX banner textures used,
// so the per-format texel decoders mirror cgfx_texture.go's; here they decode
// a whole texture once into an RGBA slice and cache it keyed by address (a
// draw samples thousands of texels — decoding per texel would crawl, and the
// game's texture memory is stable between uploads).
//
// Formats are implemented as the game's draws demand them; an unknown format
// halts.

import "fmt"

// texKey identifies one decoded texture in the cache.
type texKey struct {
	addr uint32
	fmt  uint32
	w, h uint32
}

type texImage struct {
	w, h uint32
	pix  []byte // RGBA order, w*h*4
}

// texUnitRegs returns the register base for a unit's config block. Unit 0 is
// 0x081-0x08E; units 1 and 2 are compacted blocks at 0x091 / 0x099.
func texUnitRegs(u int) (dim, param, addr, typ uint32) {
	switch u {
	case 0:
		return 0x082, 0x083, 0x085, 0x08E
	case 1:
		return 0x092, 0x093, 0x095, 0x096
	default:
		return 0x09A, 0x09B, 0x09D, 0x09E
	}
}

// sampleTexture samples unit u at (s, t) in texture coordinates. Nearest-
// neighbour: filtering quality is not what the bring-up verifies.
func (g *GPU) sampleTexture(u int, s, t float32) (rgba, bool) {
	dimR, paramR, addrR, typR := texUnitRegs(u)
	w := g.Regs[dimR] >> 16 & 0x7FF
	h := g.Regs[dimR] & 0x7FF
	if w == 0 || h == 0 {
		return rgba{255, 255, 255, 255}, true
	}
	format := g.Regs[typR] & 0xF
	addr := g.m.gpuAddrToVirt(g.Regs[addrR] << 3)

	img, ok := g.texture(addr, format, w, h)
	if !ok {
		return rgba{}, false
	}

	// Wrap per 0x083 bits 12-14 (S) and 8-10 (T): clamp-to-edge, border,
	// repeat, mirrored repeat.
	param := g.Regs[paramR]
	x := wrapCoord(s, int32(w), param>>12&7)
	// t=0 is the bottom row in PICA texture space; the decoded image is
	// top-down.
	y := int32(h) - 1 - wrapCoord(t, int32(h), param>>8&7)
	if x < 0 || y < 0 { // border mode outside: transparent black
		return rgba{}, true
	}
	p := (uint32(y)*w + uint32(x)) * 4
	return rgba{int32(img.pix[p]), int32(img.pix[p+1]), int32(img.pix[p+2]), int32(img.pix[p+3])}, true
}

func wrapCoord(f float32, n int32, mode uint32) int32 {
	v := int32(floor32(f * float32(n)))
	// The scaled coordinate indexes texels directly (nearest neighbour).
	switch mode {
	case 0: // clamp to edge
		if v < 0 {
			v = 0
		}
		if v >= n {
			v = n - 1
		}
	case 1: // clamp to border: caller treats negative as outside
		if v < 0 || v >= n {
			return -1
		}
	case 2: // repeat
		v %= n
		if v < 0 {
			v += n
		}
	default: // mirrored repeat
		period := 2 * n
		v %= period
		if v < 0 {
			v += period
		}
		if v >= n {
			v = period - 1 - v
		}
	}
	return v
}

// texture returns the decoded RGBA image for a texture, from cache or by
// decoding the tiled data now.
func (g *GPU) texture(addr, format, w, h uint32) (*texImage, bool) {
	k := texKey{addr, format, w, h}
	if img, hit := g.texCache[k]; hit {
		return img, true
	}
	img := &texImage{w: w, h: h, pix: make([]byte, w*h*4)}

	put := func(x, y uint32, r, gr, b, a byte) {
		p := (y*w + x) * 4
		img.pix[p], img.pix[p+1], img.pix[p+2], img.pix[p+3] = r, gr, b, a
	}

	switch format {
	case 0x0: // RGBA8, bytes [A,B,G,R]
		g.eachTexel(addr, w, h, 4, func(x, y, p uint32) {
			put(x, y, g.m.Read(p+3), g.m.Read(p+2), g.m.Read(p+1), g.m.Read(p))
		})
	case 0x1: // RGB8, bytes [B,G,R]
		g.eachTexel(addr, w, h, 3, func(x, y, p uint32) {
			put(x, y, g.m.Read(p+2), g.m.Read(p+1), g.m.Read(p), 255)
		})
	case 0x2: // RGBA5551
		g.eachTexel(addr, w, h, 2, func(x, y, p uint32) {
			v := uint16(g.m.Read(p)) | uint16(g.m.Read(p+1))<<8
			r := byte(v>>11) & 31
			gr := byte(v>>6) & 31
			b := byte(v>>1) & 31
			put(x, y, r<<3|r>>2, gr<<3|gr>>2, b<<3|b>>2, byte(v&1)*255)
		})
	case 0x3: // RGB565
		g.eachTexel(addr, w, h, 2, func(x, y, p uint32) {
			v := uint16(g.m.Read(p)) | uint16(g.m.Read(p+1))<<8
			r := byte(v>>11) & 31
			gr := byte(v>>5) & 63
			b := byte(v) & 31
			put(x, y, r<<3|r>>2, gr<<2|gr>>4, b<<3|b>>2, 255)
		})
	case 0x4: // RGBA4
		g.eachTexel(addr, w, h, 2, func(x, y, p uint32) {
			v := uint16(g.m.Read(p)) | uint16(g.m.Read(p+1))<<8
			put(x, y, byte(v>>12)*17, byte(v>>8&15)*17, byte(v>>4&15)*17, byte(v&15)*17)
		})
	case 0x5: // LA8: luminance + alpha
		g.eachTexel(addr, w, h, 2, func(x, y, p uint32) {
			l, a := g.m.Read(p+1), g.m.Read(p)
			put(x, y, l, l, l, a)
		})
	case 0x7: // L8
		g.eachTexel(addr, w, h, 1, func(x, y, p uint32) {
			l := g.m.Read(p)
			put(x, y, l, l, l, 255)
		})
	case 0x8: // A8
		g.eachTexel(addr, w, h, 1, func(x, y, p uint32) {
			put(x, y, 255, 255, 255, g.m.Read(p))
		})
	case 0xA: // L4: 4-bit luminance
		g.decode4(addr, w, h, img, func(n byte) [4]byte { return [4]byte{n * 17, n * 17, n * 17, 255} })
	case 0xB: // A4: 4-bit alpha
		g.decode4(addr, w, h, img, func(n byte) [4]byte { return [4]byte{255, 255, 255, n * 17} })
	case 0xC: // ETC1 compressed blocks (the CGFX banner decoder, from memory)
		size := w * h / 2
		data := make([]byte, size)
		for i := uint32(0); i < size; i++ {
			data[i] = g.m.Read(addr + i)
		}
		copy(img.pix, decodeETC1(data, int(w), int(h)).Pix)
	default:
		g.m.CPU.Halt("gpu: texture format 0x%X unimplemented (%dx%d at 0x%08X)", format, w, h, addr)
		return nil, false
	}

	if g.texCache == nil {
		g.texCache = map[texKey]*texImage{}
	}
	g.texCache[k] = img
	return img, true
}

// eachTexel walks a Morton-tiled texture's texels in storage order, calling
// visit with the pixel coordinate and the texel's address.
func (g *GPU) eachTexel(addr, w, h, bpp uint32, visit func(x, y, p uint32)) {
	p := addr
	for ty := uint32(0); ty < h; ty += 8 {
		for tx := uint32(0); tx < w; tx += 8 {
			for i := uint32(0); i < 64; i++ {
				x := i&1 | i>>1&2 | i>>2&4
				y := i>>1&1 | i>>2&2 | i>>3&4
				visit(tx+x, ty+y, p)
				p += bpp
			}
		}
	}
}

// decode4 expands a 4-bit-per-texel texture (L4 luminance or A4 alpha): two
// texels per byte, low nibble first, in the usual Morton tile walk.
func (g *GPU) decode4(addr, w, h uint32, img *texImage, px func(byte) [4]byte) {
	p := addr
	half := false
	var hold byte
	for ty := uint32(0); ty < h; ty += 8 {
		for tx := uint32(0); tx < w; tx += 8 {
			for i := uint32(0); i < 64; i++ {
				x := i&1 | i>>1&2 | i>>2&4
				y := i>>1&1 | i>>2&2 | i>>3&4
				var n byte
				if !half {
					hold = g.m.Read(p)
					p++
					n = hold & 0xF
				} else {
					n = hold >> 4
				}
				half = !half
				c := px(n)
				q := ((ty+y)*w + tx + x) * 4
				copy(img.pix[q:q+4], c[:])
			}
		}
	}
}

var _ = fmt.Sprintf
