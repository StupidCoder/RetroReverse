package xbox

// nv2a_blit.go: the NV2A 2D engine — NV_CONTEXT_SURFACES_2D (class 0x0062) and NV_IMAGE_BLIT
// (class 0x009F).
//
// Direct3D drives all GEOMETRY through the 3D Kelvin object (class 0x0097), which is why the
// rest of this machine only ever needed that one class. But a frame is not only geometry: the
// 2D blit engine copies rectangles of memory, and OutRun's bloom post-process leans on it.
//
// The bloom pipeline renders the scene, downsamples and blurs it through a chain of small
// off-screen buffers, and finally BLITS the blurred result into the buffer its full-screen
// composite pass samples. With the 2D engine unmodelled that blit was dropped: the composite
// sampled whatever stale bytes were already at the destination, decoded them as a flat grey,
// and alpha-blended that over the whole picture — the grey wash that covered the an3-drive
// scene (and, being a per-frame step, never cleared). Running the blit fills the buffer with
// the real blurred glow, and the composite becomes the subtle bloom the game intends.
//
// The engine is deliberately minimal: the two surface descriptors (source and destination
// offset/pitch/format) and a straight rectangular copy — a SRCCOPY blit, which is all the
// bloom (and the other in-frame blits) use. Both objects' state is re-programmed before every
// blit, so like the raster state it is transient and stays out of the savestate.

const (
	class2DSurfaces = 0x0062 // NV_CONTEXT_SURFACES_2D
	classImageBlit  = 0x009F // NV_IMAGE_BLIT

	// NV_CONTEXT_SURFACES_2D methods.
	surf2DFormat = 0x0300 // SET_COLOR_FORMAT
	surf2DPitch  = 0x0304 // SET_PITCH: source in the low half-word, destination in the high
	surf2DSrcOff = 0x0308 // SET_OFFSET_SOURCE
	surf2DDstOff = 0x030C // SET_OFFSET_DESTIN

	// NV_IMAGE_BLIT methods.
	blitPointIn  = 0x0300 // SET_POINT_IN:  source top-left, packed y<<16 | x
	blitPointOut = 0x0304 // SET_POINT_OUT: destination top-left
	blitSize     = 0x0308 // SET_SIZE: h<<16 | w — writing it performs the blit
)

// surfaces2D is the source/destination pair a blit copies between.
type surfaces2D struct {
	format             uint32
	srcPitch, dstPitch uint32 // bytes per row
	srcOffset          uint32 // physical byte address of the source image
	dstOffset          uint32 // physical byte address of the destination image
}

// imageBlit is the blit's source and destination top-left corners; the size method carries
// the extent and triggers the copy.
type imageBlit struct {
	inX, inY   uint32
	outX, outY uint32
}

func (g *pgraph) surf2DMethod(method, arg uint32) {
	switch method {
	case surf2DFormat:
		g.surf2D.format = arg
	case surf2DPitch:
		g.surf2D.srcPitch = arg & 0xFFFF
		g.surf2D.dstPitch = arg >> 16
	case surf2DSrcOff:
		g.surf2D.srcOffset = arg
	case surf2DDstOff:
		g.surf2D.dstOffset = arg
	}
}

func (g *pgraph) blitMethod(method, arg uint32) {
	switch method {
	case blitPointIn:
		g.blit.inX, g.blit.inY = arg&0xFFFF, arg>>16
	case blitPointOut:
		g.blit.outX, g.blit.outY = arg&0xFFFF, arg>>16
	case blitSize:
		g.doBlit(arg&0xFFFF, arg>>16)
	}
}

// doBlit copies a w x h rectangle from the source surface to the destination surface, one row
// at a time, honouring each surface's own pitch. The offsets are physical addresses (the
// image DMA contexts on the Xbox target main RAM at base 0); an out-of-range row aborts the
// blit rather than corrupting memory.
func (g *pgraph) doBlit(w, h uint32) {
	bpp := surf2DBpp(g.surf2D.format)
	if bpp == 0 || w == 0 || h == 0 {
		return
	}
	m := g.m
	s, b := &g.surf2D, &g.blit
	n := int(w * bpp)
	for y := uint32(0); y < h; y++ {
		src := int(s.srcOffset + (b.inY+y)*s.srcPitch + b.inX*bpp)
		dst := int(s.dstOffset + (b.outY+y)*s.dstPitch + b.outX*bpp)
		if src < 0 || dst < 0 || src+n > len(m.RAM) || dst+n > len(m.RAM) {
			return
		}
		copy(m.RAM[dst:dst+n], m.RAM[src:src+n])
	}
}

// surf2DBpp is the byte size of one pixel in a NV_CONTEXT_SURFACES_2D colour format. Only the
// depth matters for a SRCCOPY blit — the exact channel layout is the consumer's problem.
func surf2DBpp(format uint32) uint32 {
	switch format & 0xFF {
	case 0x01: // Y8
		return 1
	case 0x03, 0x04, 0x05, 0x06: // X1R5G5B5 / R5G6B5 / Y16 family
		return 2
	case 0x07, 0x08, 0x09, 0x0A, 0x0B: // X8R8G8B8 / A8R8G8B8 / Y32 family
		return 4
	default:
		return 4 // the Xbox's 2D surfaces are 32bpp; default to that rather than dropping the blit
	}
}
