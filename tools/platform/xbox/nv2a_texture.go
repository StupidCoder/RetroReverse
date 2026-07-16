package xbox

// nv2a_texture.go decodes and samples Kelvin textures. Each of the four texture units
// latches an offset (a physical address through the texture DMA contexts), a format
// word (dimensionality, color format, mip count, log2 sizes), address/wrap modes, and
// filters. Uncompressed "SZ" formats are stored swizzled (Morton/Z-order — the same
// family as the PSP and 3DS tilings the repo already decodes); the DXT formats are the
// published S3TC block compression, stored linearly; the "LU" formats are linear
// pitch images addressed in texels.
//
// Decoding happens once per (state, pusher run) into an RGBA8 image — the cache is
// dropped at the start of every pusher run, because between runs the CPU owns memory
// and may rewrite texture data; within one run the GPU sees a consistent snapshot.
// An unmodelled format halts loudly and names itself: sampling wrong bytes would
// paint a plausible-looking wrong frame, which is the failure mode this codebase
// treats as worse than a stop.

const (
	kelvinCtxDmaTexA = 0x0184 // SET_CONTEXT_DMA_A: texture source when fmt bit 0 set
	kelvinCtxDmaTexB = 0x0188 // SET_CONTEXT_DMA_B: texture source when fmt bit 1 set

	kelvinTexOffset  = 0x1B00 // +0x40*u
	kelvinTexFormat  = 0x1B04 // dma[1:0] cube[2] border[3] dim[7:4] color[15:8] mips[19:16] u[23:20] v[27:24] p[31:28]
	kelvinTexAddress = 0x1B08 // wrap U byte0, V byte1, P byte2
	kelvinTexControl = 0x1B0C // bit30 enable
	kelvinTexFilter  = 0x1B14 // min [23:16], mag [31:24]
	kelvinTexRect    = 0x1B1C // linear formats: width<<16 | height
	kelvinTexCtl1    = 0x1B10 // linear formats: pitch<<16
)

// Kelvin texture color formats (the format word's [15:8] field) this pipeline models.
const (
	texFmtSZ_A1R5G5B5 = 0x02
	texFmtSZ_A4R4G4B4 = 0x04
	texFmtSZ_R5G6B5   = 0x05
	texFmtSZ_A8R8G8B8 = 0x06
	texFmtSZ_X8R8G8B8 = 0x07
	texFmtDXT1        = 0x0C
	texFmtDXT3        = 0x0E
	texFmtDXT5        = 0x0F
	texFmtLU_R5G6B5   = 0x11
	texFmtLU_A8R8G8B8 = 0x12
	texFmtSZ_A8       = 0x19
	texFmtLU_X8R8G8B8 = 0x1E
)

// texKey identifies a decoded texture within one pusher run.
type texKey struct {
	offset, format, rect, ctl1 uint32
}

// texImage is a decoded RGBA8 texture.
type texImage struct {
	w, h int
	pix  []byte // 4 bytes per texel, R G B A
}

// texStateDecode resolves the enabled texture units' images for a draw.
func (g *pgraph) texStateDecode(st *rasterState) bool {
	shader := g.Regs[0x1E70>>2] // SET_SHADER_STAGE_PROGRAM: 5 bits per stage
	for u := 0; u < 4; u++ {
		r := uint32(u) * 0x40
		ctl := g.Regs[(kelvinTexControl+r)>>2]
		st.texEnable[u] = ctl>>30&1 != 0
		st.texStage[u] = shader >> (5 * uint(u)) & 0x1F
		if !st.texEnable[u] || st.texStage[u] == 0 {
			st.texEnable[u] = false
			st.texImg[u] = nil
			continue
		}
		addr := g.Regs[(kelvinTexAddress+r)>>2]
		st.texWrapU[u] = addr & 0xFF
		st.texWrapV[u] = addr >> 8 & 0xFF
		filt := g.Regs[(kelvinTexFilter+r)>>2]
		st.texBilinear[u] = filt>>24&0xFF >= 2 // mag filter linear (mips are not modelled yet)

		img, rect, ok := g.texDecode(u)
		if !ok {
			return false
		}
		st.texImg[u] = img
		st.texRect[u] = rect
	}
	return true
}

// texDecode fetches unit u's texture through its DMA context and decodes it to RGBA8,
// via the per-run cache. The bool result distinguishes "halted" from "cached".
func (g *pgraph) texDecode(u int) (*texImage, bool, bool) {
	r := uint32(u) * 0x40
	offset := g.Regs[(kelvinTexOffset+r)>>2]
	format := g.Regs[(kelvinTexFormat+r)>>2]
	rect := g.Regs[(kelvinTexRect+r)>>2]
	ctl1 := g.Regs[(kelvinTexCtl1+r)>>2]

	key := texKey{offset, format, rect, ctl1}
	if img, ok := g.texCache[key]; ok {
		return img, isLinearTexFmt(format >> 8 & 0xFF), true
	}

	if dim := format >> 4 & 0xF; dim != 2 {
		g.m.CPU.Halt("nv2a: texture unit %d dimensionality %d unmodelled (fmt=%08X)", u, dim, format)
		return nil, false, false
	}
	if format>>2&1 != 0 {
		g.m.CPU.Halt("nv2a: texture unit %d is a cube map — unmodelled (fmt=%08X)", u, format)
		return nil, false, false
	}

	// Resolve the base address through the bound DMA context (fmt bits 1:0 select it).
	dmaReg := uint32(kelvinCtxDmaTexA)
	if format&3 == 2 {
		dmaReg = kelvinCtxDmaTexB
	}
	base, _ := g.m.dmaObjectTarget(g.Regs[dmaReg>>2])
	phys, mmio, ok := g.m.translate(base + offset)
	if !ok || mmio {
		g.m.CPU.Halt("nv2a: texture unit %d at %08X is not RAM", u, base+offset)
		return nil, false, false
	}

	colorFmt := format >> 8 & 0xFF
	linear := isLinearTexFmt(colorFmt)
	var w, h int
	if linear {
		w, h = int(rect>>16), int(rect&0xFFFF)
	} else {
		w, h = 1<<(format>>20&0xF), 1<<(format>>24&0xF)
	}
	if w <= 0 || h <= 0 || w > 4096 || h > 4096 {
		g.m.CPU.Halt("nv2a: texture unit %d has impossible size %dx%d (fmt=%08X rect=%08X)", u, w, h, format, rect)
		return nil, false, false
	}

	img := &texImage{w: w, h: h, pix: make([]byte, w*h*4)}
	ram := g.m.RAM
	switch colorFmt {
	case texFmtDXT1:
		decodeDXT(img, ram, phys, 1)
	case texFmtDXT3:
		decodeDXT(img, ram, phys, 3)
	case texFmtDXT5:
		decodeDXT(img, ram, phys, 5)
	case texFmtSZ_A8R8G8B8, texFmtSZ_X8R8G8B8:
		decodeSwizzled(img, ram, phys, 4, func(pix []byte, o int, b []byte) {
			pix[o], pix[o+1], pix[o+2], pix[o+3] = b[2], b[1], b[0], b[3]
			if colorFmt == texFmtSZ_X8R8G8B8 {
				pix[o+3] = 0xFF
			}
		})
	case texFmtSZ_R5G6B5:
		decodeSwizzled(img, ram, phys, 2, decode565)
	case texFmtSZ_A1R5G5B5:
		decodeSwizzled(img, ram, phys, 2, decode1555)
	case texFmtSZ_A4R4G4B4:
		decodeSwizzled(img, ram, phys, 2, decode4444)
	case texFmtSZ_A8:
		decodeSwizzled(img, ram, phys, 1, func(pix []byte, o int, b []byte) {
			pix[o], pix[o+1], pix[o+2], pix[o+3] = 0xFF, 0xFF, 0xFF, b[0]
		})
	case texFmtLU_A8R8G8B8, texFmtLU_X8R8G8B8:
		decodeLinear(img, ram, phys, ctl1>>16, 4, func(pix []byte, o int, b []byte) {
			pix[o], pix[o+1], pix[o+2], pix[o+3] = b[2], b[1], b[0], b[3]
			if colorFmt == texFmtLU_X8R8G8B8 {
				pix[o+3] = 0xFF
			}
		})
	case texFmtLU_R5G6B5:
		decodeLinear(img, ram, phys, ctl1>>16, 2, decode565)
	default:
		g.m.CPU.Halt("nv2a: texture unit %d color format 0x%02X unmodelled (fmt=%08X)", u, colorFmt, format)
		return nil, false, false
	}

	g.texCache[key] = img
	return img, linear, true
}

// DebugDecodeTexture decodes texture unit u's current binding to RGBA8 — the
// debugger/probe surface (the n3ds RenderTexture analogue). It shares texDecode, so
// what it shows is exactly what a draw would sample.
func (g *pgraph) DebugDecodeTexture(u int) (w, h int, pix []byte, ok bool) {
	img, _, ok := g.texDecode(u)
	if !ok || img == nil {
		return 0, 0, nil, false
	}
	return img.w, img.h, img.pix, true
}

func isLinearTexFmt(colorFmt uint32) bool {
	switch colorFmt {
	case texFmtLU_R5G6B5, texFmtLU_A8R8G8B8, texFmtLU_X8R8G8B8:
		return true
	}
	return false
}

// --- texel decoders ---

// exp5/exp6 widen 5/6-bit channels to 8 bits by replicating the high bits, so full
// intensity maps to 255 (a plain shift tops out at 248/252).
func exp5(v uint16) byte { return byte(v<<3 | v>>2) }
func exp6(v uint16) byte { return byte(v<<2 | v>>4) }

func decode565(pix []byte, o int, b []byte) {
	v := uint16(b[0]) | uint16(b[1])<<8
	pix[o] = exp5(v >> 11 & 0x1F)
	pix[o+1] = exp6(v >> 5 & 0x3F)
	pix[o+2] = exp5(v & 0x1F)
	pix[o+3] = 0xFF
}

func decode1555(pix []byte, o int, b []byte) {
	v := uint16(b[0]) | uint16(b[1])<<8
	pix[o] = exp5(v >> 10 & 0x1F)
	pix[o+1] = exp5(v >> 5 & 0x1F)
	pix[o+2] = exp5(v & 0x1F)
	if v>>15 != 0 {
		pix[o+3] = 0xFF
	}
}

func decode4444(pix []byte, o int, b []byte) {
	v := uint16(b[0]) | uint16(b[1])<<8
	pix[o] = byte(v >> 8 & 0xF * 17)
	pix[o+1] = byte(v >> 4 & 0xF * 17)
	pix[o+2] = byte(v & 0xF * 17)
	pix[o+3] = byte(v >> 12 & 0xF * 17)
}

// decodeSwizzled walks the Morton (Z-order) layout the SZ formats store: the texel at
// (x, y) lives at the offset formed by interleaving the bits of x and y (the wider
// dimension's leftover high bits run linearly beyond the interleaved square).
func decodeSwizzled(img *texImage, ram []byte, phys uint32, bpp int, put func([]byte, int, []byte)) {
	for y := 0; y < img.h; y++ {
		for x := 0; x < img.w; x++ {
			a := phys + uint32(swizzleOffset(x, y, img.w, img.h))*uint32(bpp)
			if int(a)+bpp > len(ram) {
				continue
			}
			put(img.pix, (y*img.w+x)*4, ram[a:a+uint32(bpp)])
		}
	}
}

// swizzleOffset interleaves x/y bits, low bit first; when one dimension runs out of
// bits the other's remaining bits are appended contiguously.
func swizzleOffset(x, y, w, h int) int {
	off, shift := 0, 0
	for w > 1 || h > 1 {
		if w > 1 {
			off |= (x & 1) << shift
			x >>= 1
			w >>= 1
			shift++
		}
		if h > 1 {
			off |= (y & 1) << shift
			y >>= 1
			h >>= 1
			shift++
		}
	}
	return off
}

func decodeLinear(img *texImage, ram []byte, phys, pitch uint32, bpp int, put func([]byte, int, []byte)) {
	if pitch == 0 {
		pitch = uint32(img.w * bpp)
	}
	for y := 0; y < img.h; y++ {
		row := phys + uint32(y)*pitch
		for x := 0; x < img.w; x++ {
			a := row + uint32(x*bpp)
			if int(a)+bpp > len(ram) {
				continue
			}
			put(img.pix, (y*img.w+x)*4, ram[a:a+uint32(bpp)])
		}
	}
}

// decodeDXT decodes the S3TC block formats (published spec): 4x4 texel blocks; DXT1
// packs two RGB565 endpoints + 2-bit selectors in 8 bytes, DXT3 prepends 64 bits of
// explicit 4-bit alpha, DXT5 prepends interpolated alpha.
func decodeDXT(img *texImage, ram []byte, phys uint32, variant int) {
	blockBytes := uint32(8)
	if variant != 1 {
		blockBytes = 16
	}
	bw, bh := (img.w+3)/4, (img.h+3)/4
	for by := 0; by < bh; by++ {
		for bx := 0; bx < bw; bx++ {
			a := phys + uint32(by*bw+bx)*blockBytes
			if int(a)+int(blockBytes) > len(ram) {
				continue
			}
			blk := ram[a : a+blockBytes]
			color := blk
			if variant != 1 {
				color = blk[8:]
			}
			c0 := uint16(color[0]) | uint16(color[1])<<8
			c1 := uint16(color[2]) | uint16(color[3])<<8
			var pal [4][4]byte
			expand := func(c uint16) (r, g, b byte) {
				return exp5(c >> 11 & 0x1F), exp6(c >> 5 & 0x3F), exp5(c & 0x1F)
			}
			r0, g0, b0 := expand(c0)
			r1, g1, b1 := expand(c1)
			pal[0] = [4]byte{r0, g0, b0, 0xFF}
			pal[1] = [4]byte{r1, g1, b1, 0xFF}
			if variant != 1 || c0 > c1 {
				pal[2] = [4]byte{byte((2*int(r0) + int(r1)) / 3), byte((2*int(g0) + int(g1)) / 3), byte((2*int(b0) + int(b1)) / 3), 0xFF}
				pal[3] = [4]byte{byte((int(r0) + 2*int(r1)) / 3), byte((int(g0) + 2*int(g1)) / 3), byte((int(b0) + 2*int(b1)) / 3), 0xFF}
			} else {
				pal[2] = [4]byte{byte((int(r0) + int(r1)) / 2), byte((int(g0) + int(g1)) / 2), byte((int(b0) + int(b1)) / 2), 0xFF}
				pal[3] = [4]byte{0, 0, 0, 0} // DXT1's transparent black
			}
			sel := uint32(color[4]) | uint32(color[5])<<8 | uint32(color[6])<<16 | uint32(color[7])<<24
			for ty := 0; ty < 4; ty++ {
				y := by*4 + ty
				if y >= img.h {
					break
				}
				for tx := 0; tx < 4; tx++ {
					x := bx*4 + tx
					if x >= img.w {
						continue
					}
					p := pal[sel>>(2*uint(ty*4+tx))&3]
					switch variant {
					case 3:
						a4 := blk[ty*2+tx/2] >> (4 * uint(tx&1)) & 0xF
						p[3] = a4 * 17
					case 5:
						p[3] = dxt5Alpha(blk, ty*4+tx)
					}
					o := (y*img.w + x) * 4
					img.pix[o], img.pix[o+1], img.pix[o+2], img.pix[o+3] = p[0], p[1], p[2], p[3]
				}
			}
		}
	}
}

func dxt5Alpha(blk []byte, i int) byte {
	a0, a1 := int(blk[0]), int(blk[1])
	bits := uint64(blk[2]) | uint64(blk[3])<<8 | uint64(blk[4])<<16 |
		uint64(blk[5])<<24 | uint64(blk[6])<<32 | uint64(blk[7])<<40
	code := int(bits >> (3 * uint(i)) & 7)
	switch {
	case code == 0:
		return byte(a0)
	case code == 1:
		return byte(a1)
	case a0 > a1:
		return byte(((8-code)*a0 + (code-1)*a1) / 7)
	case code == 6:
		return 0
	case code == 7:
		return 255
	default:
		return byte(((6-code)*a0 + (code-1)*a1) / 5)
	}
}

// --- sampling ---

// texWrapCoord applies a wrap mode to a texel-space coordinate.
func texWrapCoord(v float32, size int, mode uint32) float32 {
	n := float32(size)
	switch mode {
	case 1: // WRAP
		v -= n * floorf32(v/n)
	case 2: // MIRROR
		p := n * 2
		v -= p * floorf32(v/p)
		if v >= n {
			v = p - v - 1.0/256
		}
	default: // CLAMP variants (3 edge, 4 border→edge, 5 OGL clamp)
		if v < 0 {
			v = 0
		}
		if v > n-1 {
			v = n - 1
		}
	}
	return v
}

// texSample samples unit u at (s, t): normalized coordinates for swizzled/compressed
// textures, texel coordinates for the linear rect formats.
func (g *pgraph) texSample(st *rasterState, u int, s, t float32) [4]float32 {
	img := st.texImg[u]
	fx, fy := s, t
	if !st.texRect[u] {
		fx *= float32(img.w)
		fy *= float32(img.h)
	}
	fetch := func(x, y int) [4]float32 {
		o := (y*img.w + x) * 4
		return [4]float32{
			float32(img.pix[o]) / 255, float32(img.pix[o+1]) / 255,
			float32(img.pix[o+2]) / 255, float32(img.pix[o+3]) / 255,
		}
	}
	if !st.texBilinear[u] {
		x := int(texWrapCoord(fx, img.w, st.texWrapU[u]))
		y := int(texWrapCoord(fy, img.h, st.texWrapV[u]))
		return fetch(x, y)
	}
	// Bilinear: sample at texel centers around (fx-0.5, fy-0.5).
	gx, gy := fx-0.5, fy-0.5
	x0f, y0f := floorf32(gx), floorf32(gy)
	wx, wy := gx-x0f, gy-y0f
	x0 := int(texWrapCoord(x0f, img.w, st.texWrapU[u]))
	x1 := int(texWrapCoord(x0f+1, img.w, st.texWrapU[u]))
	y0 := int(texWrapCoord(y0f, img.h, st.texWrapV[u]))
	y1 := int(texWrapCoord(y0f+1, img.h, st.texWrapV[u]))
	c00, c10 := fetch(x0, y0), fetch(x1, y0)
	c01, c11 := fetch(x0, y1), fetch(x1, y1)
	var out [4]float32
	for i := 0; i < 4; i++ {
		top := c00[i] + (c10[i]-c00[i])*wx
		bot := c01[i] + (c11[i]-c01[i])*wx
		out[i] = top + (bot-top)*wy
	}
	return out
}
