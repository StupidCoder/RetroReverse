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
		m.CPU.Halt("PE copy-to-texture not yet implemented (BP 0x52 params 0x%06X)", params)
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
