package xbox

// nv2a_texture.go decodes and samples Kelvin textures. Each of the four texture units
// latches an offset (a physical address through the texture DMA contexts), a format
// word (dimensionality, color format, mip count, log2 sizes), address/wrap modes, and
// filters. Uncompressed "SZ" formats are stored swizzled (Morton/Z-order — the same
// family as the PSP and 3DS tilings the repo already decodes); the DXT formats are the
// published S3TC block compression, stored linearly; the "LU" formats are linear
// pitch images addressed in texels.
//
// Decoding happens into an RGBA8 image held in a content-addressed cache (texEntry): an
// entry persists across pusher runs and is trusted for the current run only after its
// source bytes are re-confirmed by an FNV-1a hash (hashRAM over the span texSource
// computes). This keeps a static texture decoded once for the whole trajectory while a
// reflection RTT — or the shadow map the caster rewrites mid-run — is re-decoded the moment
// its bytes change, which is why even the depth texture is cacheable now. Within one run the
// GPU still sees a consistent snapshot (the first decode is reused, as the old per-run cache
// did). An unmodelled format halts loudly and names itself: sampling wrong bytes would paint
// a plausible-looking wrong frame, which is the failure mode this codebase treats as worse
// than a stop.

import (
	"encoding/binary"
	"fmt"
)

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
	// Linear depth+stencil (X8_Y24: depth [31:8], stencil [7:0] — the raster's own
	// zeta layout). A shadow-map sample: modelled only for an ALL-FAR map (the sample
	// is the unoccluded result, 1.0, under every candidate compare); a populated map
	// halts, because the compare itself is underivable here. See texDecode.
	texFmtLU_DepthX8Y24 = 0x2E
)

// texKey identifies a decoded texture within one pusher run.
type texKey struct {
	offset, format, rect, ctl1 uint32
}

// texEntry is a cached decode plus what it takes to trust it across pusher runs. The cache used
// to be dropped whole at every run start (the CPU, and a mid-run render-to-texture pass, own
// texture memory between runs); it now persists, and an entry is trusted for the current run only
// after its source bytes are re-confirmed. srcHash is an FNV-1a of exactly the bytes the decode
// read (texSource computes the span); validated is the g.texRun at which that check last passed,
// so a texture sampled by many draws in one run is hashed at most once per run and a static
// texture is decoded once for the whole trajectory. A 64-bit content hash can in principle
// collide (~2^-64) — the standard bet a texture cache makes; the frame gate is the backstop.
type texEntry struct {
	img       *texImage
	srcHash   uint64
	srcLen    uint32
	validated uint64
}

// hashRAM is FNV-1a over [phys, phys+n), eight bytes at a time. It is the cross-run staleness
// check: cheaper than a decode (no swizzle/DXT arithmetic, no allocation) and writer-agnostic —
// a CPU store and a GPU RTT pass into the source both change the hash and force a re-decode.
func hashRAM(ram []byte, phys, n uint32) uint64 {
	end := uint64(phys) + uint64(n)
	if end > uint64(len(ram)) {
		end = uint64(len(ram))
	}
	h := uint64(1469598103934665603)
	i := uint64(phys)
	for ; i+8 <= end; i += 8 {
		h = (h ^ binary.LittleEndian.Uint64(ram[i:])) * 1099511628211
	}
	for ; i < end; i++ {
		h = (h ^ uint64(ram[i])) * 1099511628211
	}
	return h
}

// texSpan is the number of source bytes a decode of the given format reads — the exact extent
// hashRAM must cover, so any byte the decoder would read changing forces a re-decode. It mirrors
// the per-format cases in texDecode; linear formats return an upper bound (h*pitch, ≥ the actual
// last-row extent). ok is false for depth and unmodelled formats, which are never cross-run
// cached (depth already bypasses the cache; an unmodelled format halts in texDecode).
func texSpan(colorFmt uint32, cube bool, w, h int, pitch uint32) (uint32, bool) {
	var faceBytes uint32
	switch colorFmt {
	case texFmtDXT1:
		faceBytes = uint32((w+3)/4*((h+3)/4)) * 8
	case texFmtDXT3, texFmtDXT5:
		faceBytes = uint32((w+3)/4*((h+3)/4)) * 16
	case texFmtSZ_A8R8G8B8, texFmtSZ_X8R8G8B8:
		faceBytes = uint32(w * h * 4)
	case texFmtSZ_R5G6B5, texFmtSZ_A1R5G5B5, texFmtSZ_A4R4G4B4:
		faceBytes = uint32(w * h * 2)
	case texFmtSZ_A8:
		faceBytes = uint32(w * h)
	case texFmtLU_A8R8G8B8, texFmtLU_X8R8G8B8:
		if pitch == 0 {
			pitch = uint32(w * 4)
		}
		faceBytes = uint32(h) * pitch
	case texFmtLU_R5G6B5:
		if pitch == 0 {
			pitch = uint32(w * 2)
		}
		faceBytes = uint32(h) * pitch
	case texFmtLU_DepthX8Y24:
		// The shadow map: h rows of pitch bytes of X8_Y24 depth. Hashing exactly these bytes
		// is what lets the depth texture be cached at all — the caster pass writes them mid-run
		// and the hash sees the change, which is the safety the old cache bypass gave up on.
		if pitch == 0 {
			pitch = uint32(w * 4)
		}
		faceBytes = uint32(h) * pitch
	default:
		return 0, false
	}
	if cube {
		return faceBytes * 6, true
	}
	return faceBytes, true
}

// texSource resolves texture unit u's source address and byte span, the way texDecode does but
// without decoding or halting — the cross-run cache check and the store both use it so the bytes
// hashed are exactly the bytes decoded. ok is false when the binding is not cacheable this way
// (not RAM, an unmodelled span, or out of range), and the caller then decodes fresh.
func (g *pgraph) texSource(u int) (phys, span uint32, ok bool) {
	r := uint32(u) * 0x40
	offset := g.Regs[(kelvinTexOffset+r)>>2]
	format := g.Regs[(kelvinTexFormat+r)>>2]
	rect := g.Regs[(kelvinTexRect+r)>>2]
	ctl1 := g.Regs[(kelvinTexCtl1+r)>>2]
	if format>>4&0xF != 2 {
		return 0, 0, false
	}
	colorFmt := format >> 8 & 0xFF
	dmaReg := uint32(kelvinCtxDmaTexA)
	if format&3 == 2 {
		dmaReg = kelvinCtxDmaTexB
	}
	base, _ := g.m.dmaObjectTarget(g.Regs[dmaReg>>2])
	p, mmio, okT := g.m.translate(base + offset)
	if !okT || mmio {
		return 0, 0, false
	}
	var w, h int
	if isLinearTexFmt(colorFmt) {
		w, h = int(rect>>16), int(rect&0xFFFF)
	} else {
		w, h = 1<<(format>>20&0xF), 1<<(format>>24&0xF)
	}
	if w <= 0 || h <= 0 || w > 4096 || h > 4096 {
		return 0, 0, false
	}
	span, ok = texSpan(colorFmt, format>>2&1 != 0, w, h, ctl1>>16)
	if !ok || uint64(p)+uint64(span) > uint64(len(g.m.RAM)) {
		return 0, 0, false
	}
	return p, span, true
}

// cacheTex stores a fresh decode keyed by its source content, so the next run can trust it
// without re-decoding as long as the bytes are unchanged.
func (g *pgraph) cacheTex(key texKey, img *texImage, u int) {
	var hsh uint64
	phys, span, ok := g.texSource(u)
	if ok {
		hsh = hashRAM(g.m.RAM, phys, span)
	}
	g.texCache[key] = &texEntry{img: img, srcHash: hsh, srcLen: span, validated: g.texRun}
}

// texImage is a decoded RGBA8 texture. A cube map stores its six faces
// consecutively in pix (face f at rows [f*h, (f+1)*h)), in the published
// +X,-X,+Y,-Y,+Z,-Z order.
type texImage struct {
	w, h int // one face's dimensions
	cube bool
	pix  []byte // 4 bytes per texel, R G B A
	// depth is set for a depth-format texture (a shadow map): the raw 24-bit depth
	// per texel. A unit bound to one samples through texSampleShadow (the hardware
	// depth compare), never texSample; pix holds a grayscale visualization for the
	// debugger only.
	depth []uint32
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
	linear := isLinearTexFmt(format >> 8 & 0xFF)
	// The cache is content-addressed: an entry is trusted for the current run only after its
	// source bytes are re-confirmed by hash. That is what makes a depth texture (shadow map)
	// cacheable at all — the old code bypassed it because the caster pass writes the map
	// mid-run, but a mid-run write changes the hash and forces exactly one re-decode after the
	// caster, so the receiver draws that follow share it instead of each re-decoding 512x512
	// depths. (The one assumption: at most one caster pass per run, which the frame gate guards.)
	if e, ok := g.texCache[key]; ok {
		// Already confirmed this run (the common case: many draws sample one texture) — trust
		// it, exactly as the old per-run cache did within a run.
		if e.validated == g.texRun {
			return e.img, linear, true
		}
		// First look this run: re-confirm the source bytes are what we decoded from. If
		// unchanged, trust the decode for the rest of the run; if changed (a CPU rewrite, an
		// RTT pass, or the shadow caster), fall through and re-decode below.
		if phys, span, ok2 := g.texSource(u); ok2 && hashRAM(g.m.RAM, phys, span) == e.srcHash {
			e.validated = g.texRun
			return e.img, linear, true
		}
	}

	if dim := format >> 4 & 0xF; dim != 2 {
		g.m.CPU.Halt("nv2a: texture unit %d dimensionality %d unmodelled (fmt=%08X)", u, dim, format)
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

	colorFmt := format >> 8 & 0xFF // linear was resolved from the same bits at the top
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

	ram := g.m.RAM
	if format>>2&1 != 0 {
		// A cube map: six w x h faces stored consecutively, +X,-X,+Y,-Y,+Z,-Z (the
		// published D3D/GL face order for this hardware family). Modelled for the DXT
		// block formats only (OutRun's race environment maps are single-mip DXT3);
		// anything else halts and names itself — a swizzled cube's per-face padding
		// and a mip chain's face stride are unverified here, and guessing either
		// would sample plausible-looking wrong texels.
		if w != h {
			g.m.CPU.Halt("nv2a: texture unit %d cube map %dx%d is not square (fmt=%08X)", u, w, h, format)
			return nil, false, false
		}
		if mips := format >> 16 & 0xF; mips > 1 {
			g.m.CPU.Halt("nv2a: texture unit %d cube map with %d mip levels — face stride unmodelled (fmt=%08X)", u, mips, format)
			return nil, false, false
		}
		var faceBytes uint32
		var decodeFace func(*texImage, uint32)
		switch colorFmt {
		case texFmtDXT1, texFmtDXT3, texFmtDXT5:
			variant := map[uint32]int{texFmtDXT1: 1, texFmtDXT3: 3, texFmtDXT5: 5}[colorFmt]
			blockBytes := uint32(8)
			if variant != 1 {
				blockBytes = 16
			}
			faceBytes = uint32((w+3)/4*((h+3)/4)) * blockBytes
			decodeFace = func(dst *texImage, addr uint32) { decodeDXT(dst, ram, addr, variant) }
		case texFmtSZ_A8R8G8B8, texFmtSZ_X8R8G8B8:
			// The face stride w*h*4 with no padding is the game's own layout: its
			// environment cube is the three (of six) 128x128 reflection RTTs, which it
			// packs exactly 0x10000 = 128*128*4 bytes apart (Part XII's census).
			faceBytes = uint32(w * h * 4)
			decodeFace = func(dst *texImage, addr uint32) {
				decodeSwizzled(dst, ram, addr, 4, func(pix []byte, o int, b []byte) {
					pix[o], pix[o+1], pix[o+2], pix[o+3] = b[2], b[1], b[0], b[3]
					if colorFmt == texFmtSZ_X8R8G8B8 {
						pix[o+3] = 0xFF
					}
				})
			}
		default:
			g.m.CPU.Halt("nv2a: texture unit %d cube map color format 0x%02X unmodelled (fmt=%08X)", u, colorFmt, format)
			return nil, false, false
		}
		img := &texImage{w: w, h: h, cube: true, pix: make([]byte, w*h*4*6)}
		face := &texImage{w: w, h: h, pix: make([]byte, w*h*4)}
		for f := 0; f < 6; f++ {
			decodeFace(face, phys+uint32(f)*faceBytes)
			copy(img.pix[f*w*h*4:], face.pix)
		}
		g.cacheTex(key, img, u)
		return img, linear, true
	}

	img := &texImage{w: w, h: h, pix: make([]byte, w*h*4)}
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
	case texFmtLU_DepthX8Y24:
		// A shadow-map sample. The game binds this 512x512 buffer as BOTH colour and zeta
		// (a depth-render pass), clears it to far (FFFFFF00), the caster pass writes real
		// occluder depth into it, and a receiver draw samples it here as a linear X8_Y24
		// depth image (dwords of depth<<8|stencil, the raster's own zeta layout). The
		// decode keeps the raw 24-bit depths; the SAMPLE is the hardware depth compare of
		// the fragment's projected oT3.r/q against the stored texel — a per-fragment op
		// that lives in texSampleShadow (nv2a_raster.go routes depth textures there),
		// never a decodable image. pix gets a grayscale view (depth high byte) for the
		// debugger. See outrun-2006-xbox.md Parts XII-XV.
		if shadowTrace {
			g.DumpZetaHist()
			if !g.shadowDumped {
				g.shadowDumped = true
				g.dumpReceiverState()
			}
		}
		pitch := ctl1 >> 16
		if pitch == 0 {
			pitch = uint32(w * 4)
		}
		img.depth = make([]uint32, w*h)
		for y := 0; y < h; y++ {
			row := phys + uint32(y)*pitch
			for x := 0; x < w; x++ {
				a := row + uint32(x*4)
				if int(a)+4 > len(ram) {
					continue
				}
				v := uint32(ram[a]) | uint32(ram[a+1])<<8 | uint32(ram[a+2])<<16 | uint32(ram[a+3])<<24
				d := v >> 8
				img.depth[y*w+x] = d
				o := (y*w + x) * 4
				img.pix[o], img.pix[o+1], img.pix[o+2], img.pix[o+3] = byte(d>>16), byte(d>>16), byte(d>>16), 0xFF
			}
		}
	default:
		g.m.CPU.Halt("nv2a: texture unit %d color format 0x%02X unmodelled (fmt=%08X)", u, colorFmt, format)
		return nil, false, false
	}

	g.cacheTex(key, img, u)
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
	case texFmtLU_R5G6B5, texFmtLU_A8R8G8B8, texFmtLU_X8R8G8B8, texFmtLU_DepthX8Y24:
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

// texSampleCube samples a cube-map unit with the interpolated 3-component direction:
// major-axis face selection and the per-face (s, t) mapping follow the published
// D3D/GL cube conventions for this hardware family. Filtering clamps within the
// selected face (no cross-face seam filtering).
func (g *pgraph) texSampleCube(st *rasterState, u int, x, y, z float32) [4]float32 {
	img := st.texImg[u]
	ax, ay, az := absf32(x), absf32(y), absf32(z)
	var face int
	var sc, tc, ma float32
	switch {
	case ax >= ay && ax >= az:
		if x >= 0 {
			face, sc, tc = 0, -z, -y
		} else {
			face, sc, tc = 1, z, -y
		}
		ma = ax
	case ay >= az:
		if y >= 0 {
			face, sc, tc = 2, x, z
		} else {
			face, sc, tc = 3, x, -z
		}
		ma = ay
	default:
		if z >= 0 {
			face, sc, tc = 4, x, -y
		} else {
			face, sc, tc = 5, -x, -y
		}
		ma = az
	}
	if ma == 0 {
		return [4]float32{0, 0, 0, 1}
	}
	fx := (sc/ma + 1) * 0.5 * float32(img.w)
	fy := (tc/ma + 1) * 0.5 * float32(img.h)
	clampi := func(v, hi int) int {
		if v < 0 {
			return 0
		}
		if v > hi {
			return hi
		}
		return v
	}
	fetch := func(xi, yi int) [4]float32 {
		o := ((face*img.h+yi)*img.w + xi) * 4
		return [4]float32{
			float32(img.pix[o]) / 255, float32(img.pix[o+1]) / 255,
			float32(img.pix[o+2]) / 255, float32(img.pix[o+3]) / 255,
		}
	}
	if !st.texBilinear[u] {
		return fetch(clampi(int(fx), img.w-1), clampi(int(fy), img.h-1))
	}
	gx, gy := fx-0.5, fy-0.5
	x0f, y0f := floorf32(gx), floorf32(gy)
	wx, wy := gx-x0f, gy-y0f
	x0, x1 := clampi(int(x0f), img.w-1), clampi(int(x0f)+1, img.w-1)
	y0, y1 := clampi(int(y0f), img.h-1), clampi(int(y0f)+1, img.h-1)
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

func absf32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// traceShadowFrag is the RR_SHADOWFRAG instrument: per-draw statistics of the
// comparand evidence at the receiver — r (post q-divide) against the stored 24-bit
// texel, split into fragments that hit populated (non-far) texels vs far ones, with
// the screen box and the r-D range. Printed when the draw number advances and at
// DumpZetaHist time.
func (g *pgraph) traceShadowFrag(st *rasterState, u, px, py int, s, t, r, q float32) {
	img := st.texImg[u]
	x := int(texWrapCoord(s, img.w, st.texWrapU[u]))
	y := int(texWrapCoord(t, img.h, st.texWrapV[u]))
	d := img.depth[y*img.w+x]
	if g.shadowFrag != nil && g.shadowFrag.draw != g.Draws {
		g.shadowFrag.print()
		g.shadowFrag = nil
	}
	if g.shadowFrag == nil {
		g.shadowFrag = &shadowFragStats{draw: g.Draws, pxMin: px, pxMax: px, pyMin: py, pyMax: py,
			rMin: r, rMax: r, dMin: 0xFFFFFF, dMax: 0, diffMin: 0, diffMax: 0}
	}
	f := g.shadowFrag
	if px < f.pxMin {
		f.pxMin = px
	}
	if px > f.pxMax {
		f.pxMax = px
	}
	if py < f.pyMin {
		f.pyMin = py
	}
	if py > f.pyMax {
		f.pyMax = py
	}
	if r < f.rMin {
		f.rMin = r
	}
	if r > f.rMax {
		f.rMax = r
	}
	if d != 0xFFFFFF {
		diff := r - float32(d)
		if f.near == 0 {
			f.diffMin, f.diffMax = diff, diff
		}
		if diff < f.diffMin {
			f.diffMin = diff
		}
		if diff > f.diffMax {
			f.diffMax = diff
		}
		if d < f.dMin {
			f.dMin = d
		}
		if d > f.dMax {
			f.dMax = d
		}
		f.near++
		if diff > 0 {
			f.behind++
		}
	} else {
		f.far++
	}
}

type shadowFragStats struct {
	draw                       int
	near, far, behind          int
	pxMin, pxMax, pyMin, pyMax int
	rMin, rMax                 float32
	dMin, dMax                 uint32
	diffMin, diffMax           float32
}

func (f *shadowFragStats) print() {
	fmt.Printf("SHADOWFRAG draw=%d frags=%d near=%d behind=%d px=[%d,%d]x[%d,%d] r=[%.0f,%.0f] D=[%06X,%06X] r-D=[%.0f,%.0f]\n",
		f.draw, f.near+f.far, f.near, f.behind, f.pxMin, f.pxMax, f.pyMin, f.pyMax,
		f.rMin, f.rMax, f.dMin, f.dMax, f.diffMin, f.diffMax)
}

// shadowComparePass is the shadow map's depth-compare function: does the fragment's
// comparand r (24-bit depth units) pass against the stored texel d — pass meaning
// unoccluded, sampling as 1.0. Derived at the Part XV start-line caster (the first
// populated map any reachable frame provides), the direction pinned three
// independent ways:
//
//   - an all-far map must sample as 1.0 everywhere (Part XIII's combiner argument),
//     which only the r<=D family satisfies (r is always below the far value);
//   - the r-D SIGN partitions the receiver's geometry exactly along "is the caster
//     between this surface and the light": road fragments under the grid opponent
//     are behind its stored depth (+1.7k..+8.4k, 100% of them), the overhead start
//     gantry crossing the same texels is far in front (~-160k), and the caster's
//     own lit surfaces sit in a small negative bias band (the game's acne guard);
//   - the artefact: under r<=D exactly 73 pixels darken, all inside x[262,282]
//     y[260,267] — the road contact directly beneath the mapped caster; under
//     r>=D the whole frame darkens EXCEPT that footprint (207,432 pixels) — the
//     geometric inversion, refuted.
//
// Strictness is unobservable: the comparand is a float against integer depths and
// no reachable fragment lands exactly on a stored value — the lt and le frames are
// byte-identical (as are ge/gt). LEQUAL is the declared choice, echoing the depth
// func the game runs its own zeta through (0x0354 = 0x203). No latched register
// carries a shadow compare func: at the receiver the only GL compare enums in the
// register file are the known alpha/depth/stencil funcs (dumpReceiverState scans).
// RR_SHADOWCMP overrides for probes (lt/le/gt/ge/one/zero).
func shadowComparePass(r float32, d uint32) bool {
	switch shadowCmpEnv {
	case "lt":
		return r < float32(d)
	case "gt":
		return r > float32(d)
	case "ge":
		return r >= float32(d)
	case "one": // probe: force unoccluded — the no-shadow control frame
		return true
	case "zero": // probe: force occluded — the all-shadow control frame
		return false
	}
	return r <= float32(d)
}

// texSampleShadow samples a depth texture: the texture unit's hardware depth compare.
// s and t are texel coordinates (linear rect format, already divided by q); r is the
// fragment's comparand — oT3.r/q, in the map's own 24-bit depth units (the game's
// texture matrix bakes the scale in; measured at the Part XV halt, r/q lands within
// quantization of the stored depth at caster contact points). The compare returns
// 0/1 replicated into all four channels. With a linear mag filter the hardware
// compares at the four neighbouring texels and bilinearly weights the four 0/1
// results (percentage-closer filtering) — a depth map's values must never themselves
// be interpolated.
func (g *pgraph) texSampleShadow(st *rasterState, u int, s, t, r float32) [4]float32 {
	img := st.texImg[u]
	cmp := func(x, y int) float32 {
		if shadowComparePass(r, img.depth[y*img.w+x]) {
			return 1
		}
		return 0
	}
	if !st.texBilinear[u] {
		x := int(texWrapCoord(s, img.w, st.texWrapU[u]))
		y := int(texWrapCoord(t, img.h, st.texWrapV[u]))
		v := cmp(x, y)
		return [4]float32{v, v, v, v}
	}
	gx, gy := s-0.5, t-0.5
	x0f, y0f := floorf32(gx), floorf32(gy)
	wx, wy := gx-x0f, gy-y0f
	x0 := int(texWrapCoord(x0f, img.w, st.texWrapU[u]))
	x1 := int(texWrapCoord(x0f+1, img.w, st.texWrapU[u]))
	y0 := int(texWrapCoord(y0f, img.h, st.texWrapV[u]))
	y1 := int(texWrapCoord(y0f+1, img.h, st.texWrapV[u]))
	top := cmp(x0, y0) + (cmp(x1, y0)-cmp(x0, y0))*wx
	bot := cmp(x0, y1) + (cmp(x1, y1)-cmp(x0, y1))*wx
	v := top + (bot-top)*wy
	return [4]float32{v, v, v, v}
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
