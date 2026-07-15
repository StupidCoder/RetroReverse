package gc

// gpu_efb.go is Flipper's embedded framebuffer and the pixel-engine copy that ends a frame —
// the step that finally puts a picture on screen.
//
// The EFB is the on-chip colour buffer the graphics pipe draws into: 640x528 at most, private
// to the graphics chip, never addressed by the CPU. It is not what the video interface scans
// out. When a frame is finished the game issues a pixel-engine copy — GXCopyDisp — which reads
// a rectangle of the EFB, converts it from RGB to the YUV 4:2:2 the display wants, and writes
// it to the external framebuffer (the XFB) in main memory, where the video interface finds it.
// That copy is the whole of this file's work, and the copy with its clear flag set is also how
// the EFB is wiped to the background colour between frames.
//
// Until the rasteriser exists, nothing draws into the EFB, so a copy carries only the clear
// colour the game programmed — which is exactly the picture this stage is meant to produce: a
// solid frame of the colour the game chose for its background, proving the whole EFB -> XFB ->
// YUV -> video path with no geometry in the way. A copy the game aims at a texture instead of
// the display is a later stage, and it halts loudly here naming the register that asked for it.

const (
	efbWidth  = 640
	efbHeight = 528
)

// ensureEFB allocates the embedded framebuffer on first use and fills it with the current
// clear colour, which models an EFB that starts at the background the game programmed rather
// than at uninitialised garbage — so even the first copied frame shows that colour.
func (g *gpu) ensureEFB() {
	if g.EFB == nil {
		g.EFB = make([]uint32, efbWidth*efbHeight)
		g.clearEFB()
	}
}

// clearColor reads the copy-clear colour the game set through the pixel-engine registers. The
// AR register holds alpha in its high byte and red in its low byte; the GB register holds
// green then blue.
func (g *gpu) clearColor() (r, gg, b, a uint8) {
	ar := g.BP[0x4F]
	gb := g.BP[0x50]
	return uint8(ar & 0xFF), uint8(gb >> 8), uint8(gb & 0xFF), uint8(ar >> 8)
}

// clearEFB fills the whole embedded framebuffer with the clear colour.
func (g *gpu) clearEFB() {
	r, gg, b, a := g.clearColor()
	px := packRGBA(r, gg, b, a)
	for i := range g.EFB {
		g.EFB[i] = px
	}
}

// efbAt reads one EFB pixel, clamped to the buffer so a copy rectangle that runs past the
// programmed edge reads the edge rather than panicking.
func (g *gpu) efbAt(x, y int) uint32 {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= efbWidth {
		x = efbWidth - 1
	}
	if y >= efbHeight {
		y = efbHeight - 1
	}
	return g.EFB[y*efbWidth+x]
}

// copyDisplay performs the pixel-engine copy the trigger register (BP 0x52) asked for. It
// reads the source rectangle from the EFB registers, converts each pair of pixels to one YUY2
// quad, and writes them to the destination the copy address register names. With the copy's
// clear flag set it then wipes the EFB to the clear colour for the next frame.
func (g *gpu) copyDisplay(m *Machine, params uint32) {
	g.ensureEFB()

	toXFB := params&(1<<14) != 0
	clear := params&(1<<11) != 0
	if !toXFB {
		g.copyTexture(m, params)
		if clear && !m.CPU.Halted {
			g.clearEFB()
		}
		return
	}

	tl := g.BP[0x49]
	wh := g.BP[0x4A]
	sx := int(tl & 0x3FF)
	sy := int((tl >> 10) & 0x3FF)
	w := int(wh&0x3FF) + 1
	h := int((wh>>10)&0x3FF) + 1
	dst := phys((g.BP[0x4B] & 0x00FFFFFF) << 5)
	stride := uint32(w * 2) // YUY2: two bytes per pixel, one row after another

	r, gg, b, a := g.clearColor()
	m.logf("PE copy-to-XFB: EFB (%d,%d) %dx%d -> 0x%08X clear=%v (clear colour R%d G%d B%d A%d)",
		sx, sy, w, h, dst, clear, r, gg, b, a)

	for y := 0; y < h; y++ {
		base := dst + uint32(y)*stride
		for x := 0; x < w; x += 2 {
			r0, g0, b0 := unpackRGB(g.efbAt(sx+x, sy+y))
			r1, g1, b1 := unpackRGB(g.efbAt(sx+x+1, sy+y))
			y0 := luma(r0, g0, b0)
			y1 := luma(r1, g1, b1)
			// The two pixels share one chroma pair; the encoder box-filters them, which for a
			// pair is their average.
			cb := chromaB((r0+r1)/2, (g0+g1)/2, (b0+b1)/2)
			cr := chromaR((r0+r1)/2, (g0+g1)/2, (b0+b1)/2)
			o := base + uint32(x)*2
			if int(o)+3 >= len(m.RAM) {
				break
			}
			m.RAM[o], m.RAM[o+1], m.RAM[o+2], m.RAM[o+3] = y0, cb, y1, cr
		}
	}

	if clear {
		g.clearEFB()
	}
}

// copyTexture performs the pixel-engine copy aimed at a texture: it reads the source rectangle
// from the EFB and encodes it into the destination's tiled texture format in main memory — the
// exact inverse of gpu_texture.go's per-format decoders, so a copied texture read back through
// the TX unit returns what the EFB held (to the format's precision). Render-to-texture is how
// this game builds the logo fade, the flashlight beam and the ghost effects, so this copy is on
// the critical path of every scene past the boot logo.
//
// The control word's field layout is pinned from the game's own GXSetTexCopyDst/GXCopyTex code
// (0x801F543C/0x801F5B2C in the DOL), not from a wiki: the function inserts the destination
// format's bit 3 at bit 3 (rlwimi r0,r31,0,28,28) and its low three bits at bits 4..6
// (slwi r0,r31,4 after r31 &= 7), the half-scale flag at bit 9 (rlwinm r3,r30,9,15,22), sets
// bit 15 for the intensity formats (I4/I8/IA4/IA8, and YUVA8) and bit 16 unconditionally, and
// GXCopyTex rotates its clear argument to bit 11 and clears bit 14 — the same bit GXCopyDisp
// ORs in as 0x4000 — before triggering. The destination base arrives shifted right five into
// BP 0x4B, and the stride between rows of tiles, in 32-byte units, into BP 0x4D.
func (g *gpu) copyTexture(m *Machine, params uint32) {
	tl := g.BP[0x49]
	wh := g.BP[0x4A]
	sx := int(tl & 0x3FF)
	sy := int((tl >> 10) & 0x3FF)
	w := int(wh&0x3FF) + 1
	h := int((wh>>10)&0x3FF) + 1
	dst := (g.BP[0x4B] & 0x00FFFFFF) << 5
	stride := int(g.BP[0x4D]&0x3FF) << 5
	format := int((params>>4)&7 | ((params>>3)&1)<<3)
	intensity := params&(1<<15) != 0
	halfScale := params&(1<<9) != 0

	// Gaps that have not been exercised halt loudly rather than write a plausible wrong
	// texture: the intensity conversion's luma weights, the half-scale box filter, and a copy
	// whose source is the Z buffer (the game's GXCopyTex forces the PE pixel format to Z24
	// around one, so that register names it) each wait for the draw that first asks.
	if intensity {
		m.CPU.Halt("PE copy-to-texture: intensity format not yet implemented (params 0x%06X)", params)
		return
	}
	if halfScale {
		m.CPU.Halt("PE copy-to-texture: half-scale box filter not yet implemented (params 0x%06X)", params)
		return
	}
	if g.BP[0x43]&7 == 3 {
		m.CPU.Halt("PE copy-to-texture: Z-buffer source (PE_CONTROL pixel format Z24) not yet implemented")
		return
	}

	m.logf("PE copy-to-texture: EFB (%d,%d) %dx%d -> 0x%08X format 0x%X stride %d",
		sx, sy, w, h, dst, format, stride)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := g.efbAt(sx+x, sy+y)
			r, gg, b := unpackRGB(px)
			a := uint8(px)
			if !g.encodeCopyTexel(m, dst, stride, format, x, y, r, gg, b, a) {
				m.CPU.Halt("PE copy-to-texture: destination format 0x%X not yet implemented (params 0x%06X)",
					format, params)
				return
			}
		}
	}
}

// copyTexelOffset locates texel (x,y) in the tiled copy destination: the byte offset of its
// tile (rows of tiles are stride bytes apart — the register-programmed pitch, not a value
// derived from the width) and its row-major index within the tile.
func copyTexelOffset(stride, bw, bh, tileBytes, x, y int) (tileBase, inTile int) {
	return (y/bh)*stride + (x/bw)*tileBytes, (y%bh)*bw + x%bw
}

// encodeCopyTexel writes one EFB pixel into the copy destination in the given texture format.
// Each arm is the inverse of the corresponding decoder in gpu_texture.go — same tile geometry,
// same channel packing — quantised to the format's precision. It reports false for a format it
// does not carry so the caller can halt naming it.
func (g *gpu) encodeCopyTexel(m *Machine, dst uint32, stride, format, x, y int, r, gg, b, a uint8) bool {
	switch format {
	case 0x0: // R4: 8x8 tile of nibbles, the red channel, high nibble first
		tb, in := copyTexelOffset(stride, 8, 8, 32, x, y)
		off := dst + uint32(tb+in/2)
		old := m.texByte(off)
		if in&1 == 0 {
			m.texWriteByte(off, old&0x0F|(r&0xF0))
		} else {
			m.texWriteByte(off, old&0xF0|(r>>4))
		}
	case 0x2: // RA4: 8x4 tile of bytes, red in the low nibble, alpha in the high (see texIA4)
		tb, in := copyTexelOffset(stride, 8, 4, 32, x, y)
		m.texWriteByte(dst+uint32(tb+in), a&0xF0|(r>>4))
	case 0x3: // RA8: 4x4 tile of halfwords, alpha in the high byte (see texIA8)
		tb, in := copyTexelOffset(stride, 4, 4, 32, x, y)
		off := dst + uint32(tb+in*2)
		m.texWriteByte(off, a)
		m.texWriteByte(off+1, r)
	case 0x4: // RGB565: 4x4 tile of big-endian 5-6-5 halfwords
		tb, in := copyTexelOffset(stride, 4, 4, 32, x, y)
		v := uint16(r>>3)<<11 | uint16(gg>>2)<<5 | uint16(b>>3)
		off := dst + uint32(tb+in*2)
		m.texWriteByte(off, uint8(v>>8))
		m.texWriteByte(off+1, uint8(v))
	case 0x5: // RGB5A3: opaque pixels take the 15-bit RGB form, translucent the 3-bit-alpha form
		tb, in := copyTexelOffset(stride, 4, 4, 32, x, y)
		var v uint16
		if a >= 0xE0 {
			v = 0x8000 | uint16(r>>3)<<10 | uint16(gg>>3)<<5 | uint16(b>>3)
		} else {
			v = uint16(a>>5)<<12 | uint16(r>>4)<<8 | uint16(gg>>4)<<4 | uint16(b>>4)
		}
		off := dst + uint32(tb+in*2)
		m.texWriteByte(off, uint8(v>>8))
		m.texWriteByte(off+1, uint8(v))
	case 0x6: // RGBA8: 4x4 tile in two 32-byte halves, alpha/red pairs then green/blue pairs
		tb, in := copyTexelOffset(stride, 4, 4, 64, x, y)
		m.texWriteByte(dst+uint32(tb+in*2), a)
		m.texWriteByte(dst+uint32(tb+in*2)+1, r)
		m.texWriteByte(dst+uint32(tb+32+in*2), gg)
		m.texWriteByte(dst+uint32(tb+32+in*2)+1, b)
	case 0x7, 0x8, 0x9, 0xA: // A8/R8/G8/B8: 8x4 tile of bytes, one channel
		tb, in := copyTexelOffset(stride, 8, 4, 32, x, y)
		var v uint8
		switch format {
		case 0x7:
			v = a
		case 0x8:
			v = r
		case 0x9:
			v = gg
		default:
			v = b
		}
		m.texWriteByte(dst+uint32(tb+in), v)
	default: // RG8/GB8 and anything unnamed wait for a draw that asks for them
		return false
	}
	return true
}

// texWriteByte writes one byte of texture data to main memory, masking to the RAM size the same
// way texByte reads — a copy programmed past the end of memory is a bug to log, not a crash.
func (m *Machine) texWriteByte(addr uint32, v uint8) {
	a := phys(addr)
	if int(a) >= len(m.RAM) {
		return
	}
	m.RAM[a] = v
}

// packRGBA and unpackRGB move between an EFB pixel and its components. The EFB carries alpha,
// but the display copy discards it, so the readers take only the colour.
func packRGBA(r, g, b, a uint8) uint32 {
	return uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8 | uint32(a)
}

func unpackRGB(px uint32) (r, g, b uint8) {
	return uint8(px >> 24), uint8(px >> 16), uint8(px >> 8)
}

// luma, chromaB and chromaR are the RGB-to-YUV encode, written as the exact inverse of the
// YUV-to-RGB decode the video path uses (see yuv2rgb in debug.go), so a colour copied out and
// scanned back in survives the round trip. The luma weights are the standard BT.601 set; the
// chroma scales are the video encoder's own, matched to the decode's coefficients.
func luma(r, g, b uint8) uint8 {
	return clamp8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
}

func chromaB(r, g, b uint8) uint8 {
	y := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return clamp8(128 + (float64(b)-y)/1.732)
}

func chromaR(r, g, b uint8) uint8 {
	y := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return clamp8(128 + (float64(r)-y)/1.371)
}
