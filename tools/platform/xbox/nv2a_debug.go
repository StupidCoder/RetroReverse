package xbox

// nv2a_debug.go is what the NV2A offers a debugger: names for the methods in its command
// stream, the bound textures as images, and a free view of RAM decoded as a texture. It
// is the read-only counterpart to the pipeline — nothing here may change what the machine
// would have done, which is why the texture view shares the pipeline's own decoder rather
// than reimplementing it (what the panel shows is then exactly what a draw would sample)
// and why the region view never touches pgraph state at all.
//
// The method names are the NV2A hardware's (envytools/nouveau), not any game's — the same
// sanctioned platform-spec exception the register map itself is under. Only the methods
// this pipeline models are named; anything else reports as its raw offset rather than as a
// guess, because a confidently-wrong label in a command list is worse than a number.

import (
	"fmt"
	"image"
)

// NVMethodName gives a short label for one (class, method) pair. An unmodelled method
// gets its offset back rather than an invented name.
func NVMethodName(class, method uint32) string {
	if method == nvSetObject {
		return "SET_OBJECT"
	}
	if class != classKelvin {
		return fmt.Sprintf("CLASS_%04X_M%04X", class, method)
	}
	// The banked ranges first: a texture unit's registers repeat every 0x40 bytes, a
	// vertex array's every 4, and the program/constant windows are FIFOs.
	switch {
	case method >= kelvinTexOffset && method < kelvinTexOffset+4*0x40:
		u := (method - kelvinTexOffset) / 0x40
		switch kelvinTexOffset + (method-kelvinTexOffset)%0x40 {
		case kelvinTexOffset:
			return fmt.Sprintf("SET_TEXTURE_OFFSET[%d]", u)
		case kelvinTexFormat:
			return fmt.Sprintf("SET_TEXTURE_FORMAT[%d]", u)
		case kelvinTexAddress:
			return fmt.Sprintf("SET_TEXTURE_ADDRESS[%d]", u)
		case kelvinTexControl:
			return fmt.Sprintf("SET_TEXTURE_CONTROL0[%d]", u)
		case kelvinTexCtl1:
			return fmt.Sprintf("SET_TEXTURE_CONTROL1[%d]", u)
		case kelvinTexFilter:
			return fmt.Sprintf("SET_TEXTURE_FILTER[%d]", u)
		case kelvinTexRect:
			return fmt.Sprintf("SET_TEXTURE_IMAGE_RECT[%d]", u)
		}
	case method >= kelvinProgData && method < kelvinProgData+0x80:
		return "TRANSFORM_PROGRAM"
	case method >= kelvinConstData && method < kelvinConstData+0x80:
		return "TRANSFORM_CONSTANT"
	case method >= kelvinVertexData4C && method < kelvinVertexData4C+0x40:
		return fmt.Sprintf("SET_VERTEX_DATA4UB[%d]", (method-kelvinVertexData4C)>>2)
	case method >= kelvinVtxArrayOffset && method < kelvinVtxArrayOffset+0x40:
		return fmt.Sprintf("SET_VERTEX_DATA_ARRAY_OFFSET[%d]", (method-kelvinVtxArrayOffset)>>2)
	case method >= kelvinVtxArrayFormat && method < kelvinVtxArrayFormat+0x40:
		return fmt.Sprintf("SET_VERTEX_DATA_ARRAY_FORMAT[%d]", (method-kelvinVtxArrayFormat)>>2)
	}
	if n, ok := kelvinMethodNames[method]; ok {
		return n
	}
	return fmt.Sprintf("KELVIN_M%04X", method)
}

// kelvinMethodNames names the singleton methods the pipeline models.
var kelvinMethodNames = map[uint32]string{
	kelvinCtxDmaTexA:         "SET_CONTEXT_DMA_A",
	kelvinCtxDmaTexB:         "SET_CONTEXT_DMA_B",
	kelvinCtxDmaVertexA:      "SET_CONTEXT_DMA_VERTEX_A",
	kelvinCtxDmaVertexB:      "SET_CONTEXT_DMA_VERTEX_B",
	kelvinCtxDmaSemaphore:    "SET_CONTEXT_DMA_SEMAPHORE",
	kelvinSurfaceClipH:       "SET_SURFACE_CLIP_HORIZONTAL",
	kelvinSurfaceClipV:       "SET_SURFACE_CLIP_VERTICAL",
	kelvinSurfaceFormat:      "SET_SURFACE_FORMAT",
	kelvinSurfacePitch:       "SET_SURFACE_PITCH",
	kelvinSurfaceColorOffset: "SET_SURFACE_COLOR_OFFSET",
	kelvinSurfaceZetaOffset:  "SET_SURFACE_ZETA_OFFSET",
	kelvinWindowClipType:     "SET_WINDOW_CLIP_TYPE",
	kelvinAlphaTestEnable:    "SET_ALPHA_TEST_ENABLE",
	kelvinBlendEnable:        "SET_BLEND_ENABLE",
	kelvinCullFaceEnable:     "SET_CULL_FACE_ENABLE",
	kelvinDepthTestEnable:    "SET_DEPTH_TEST_ENABLE",
	kelvinAlphaFunc:          "SET_ALPHA_FUNC",
	kelvinAlphaRef:           "SET_ALPHA_REF",
	kelvinBlendSrcFactor:     "SET_BLEND_FUNC_SFACTOR",
	kelvinBlendDstFactor:     "SET_BLEND_FUNC_DFACTOR",
	kelvinBlendColor:         "SET_BLEND_COLOR",
	kelvinBlendEquation:      "SET_BLEND_EQUATION",
	kelvinDepthFunc:          "SET_DEPTH_FUNC",
	kelvinColorMask:          "SET_COLOR_MASK",
	kelvinDepthWriteMask:     "SET_DEPTH_MASK",
	kelvinShadeMode:          "SET_SHADE_MODE",
	kelvinCullFace:           "SET_CULL_FACE",
	kelvinFrontFace:          "SET_FRONT_FACE",
	kelvinBeginEnd:           "SET_BEGIN_END",
	kelvinElement16:          "ARRAY_ELEMENT16",
	kelvinElement32:          "ARRAY_ELEMENT32",
	kelvinDrawArrays:         "DRAW_ARRAYS",
	kelvinInlineArray:        "INLINE_ARRAY",
	kelvinSemaphoreOffset:    "SET_SEMAPHORE_OFFSET",
	kelvinSemaphoreRelease:   "BACK_END_WRITE_SEMAPHORE_RELEASE",
	kelvinZStencilClearValue: "SET_ZSTENCIL_CLEAR_VALUE",
	kelvinColorClearValue:    "SET_COLOR_CLEAR_VALUE",
	kelvinClearSurface:       "CLEAR_SURFACE",
	kelvinClearRectH:         "SET_CLEAR_RECT_HORIZONTAL",
	kelvinClearRectV:         "SET_CLEAR_RECT_VERTICAL",
	kelvinCombinerControl:    "SET_COMBINER_CONTROL",
	kelvinTransformExecMode:  "SET_TRANSFORM_EXECUTION_MODE",
	kelvinCxtWriteEnable:     "SET_TRANSFORM_CONSTANT_LOAD_CXT",
	kelvinProgLoad:           "SET_TRANSFORM_PROGRAM_LOAD",
	kelvinProgStart:          "SET_TRANSFORM_PROGRAM_START",
	kelvinConstLoad:          "SET_TRANSFORM_CONSTANT_LOAD",
	0x1E70:                   "SET_SHADER_STAGE_PROGRAM",
}

// NVMethodDecode is the one-line human decode of a method's argument — what the
// debugger shows beside the raw word. Only the arguments whose structure this
// pipeline actually reads are decoded; everything else gets "".
func NVMethodDecode(class, method, arg uint32) string {
	if class != classKelvin {
		return ""
	}
	switch method {
	case kelvinBeginEnd:
		if arg == 0 {
			return "END — the batch draws here"
		}
		return fmt.Sprintf("BEGIN %s", nvPrimName(arg))
	case kelvinClearSurface:
		var s string
		if arg&0xF0 != 0 {
			s += "color "
		}
		if arg&1 != 0 {
			s += "depth "
		}
		if arg&2 != 0 {
			s += "stencil "
		}
		return "clear " + s
	case kelvinSurfaceClipH, kelvinSurfaceClipV, kelvinClearRectH, kelvinClearRectV:
		return fmt.Sprintf("%d .. %d", arg&0xFFFF, arg>>16)
	case kelvinSurfacePitch:
		return fmt.Sprintf("color pitch %d, zeta pitch %d", arg&0xFFFF, arg>>16)
	case kelvinSurfaceColorOffset, kelvinSurfaceZetaOffset:
		return fmt.Sprintf("@%08X", arg)
	case kelvinColorClearValue:
		return fmt.Sprintf("A%02X R%02X G%02X B%02X", arg>>24, arg>>16&0xFF, arg>>8&0xFF, arg&0xFF)
	case kelvinDrawArrays:
		return fmt.Sprintf("start %d, %d vertices", arg&0xFFFFFF, (arg>>24)+1)
	case kelvinSemaphoreRelease:
		return fmt.Sprintf("release %d", arg)
	}
	if method >= kelvinTexOffset && method < kelvinTexOffset+4*0x40 {
		u := (method - kelvinTexOffset) / 0x40
		switch kelvinTexOffset + (method-kelvinTexOffset)%0x40 {
		case kelvinTexOffset:
			return fmt.Sprintf("unit %d @%08X", u, arg)
		case kelvinTexFormat:
			return fmt.Sprintf("unit %d %s, %dx%d", u, nvTexFormatName(arg>>8&0xFF),
				1<<(arg>>20&0xF), 1<<(arg>>24&0xF))
		case kelvinTexControl:
			return fmt.Sprintf("unit %d enable=%t", u, arg>>30&1 != 0)
		}
	}
	return ""
}

// nvPrimName names a BEGIN's primitive type.
func nvPrimName(p uint32) string {
	switch p {
	case 1:
		return "POINTS"
	case 2:
		return "LINES"
	case 3:
		return "LINE_LOOP"
	case 4:
		return "LINE_STRIP"
	case 5:
		return "TRIANGLES"
	case 6:
		return "TRIANGLE_STRIP"
	case 7:
		return "TRIANGLE_FAN"
	case 8:
		return "QUADS"
	case 9:
		return "QUAD_STRIP"
	case 10:
		return "POLYGON"
	}
	return fmt.Sprintf("PRIM_%d", p)
}

// nvTexFormatName names a texture color format code.
func nvTexFormatName(code uint32) string {
	switch code {
	case texFmtSZ_A1R5G5B5:
		return "SZ_A1R5G5B5"
	case texFmtSZ_A4R4G4B4:
		return "SZ_A4R4G4B4"
	case texFmtSZ_R5G6B5:
		return "SZ_R5G6B5"
	case texFmtSZ_A8R8G8B8:
		return "SZ_A8R8G8B8"
	case texFmtSZ_X8R8G8B8:
		return "SZ_X8R8G8B8"
	case texFmtDXT1:
		return "DXT1"
	case texFmtDXT3:
		return "DXT3"
	case texFmtDXT5:
		return "DXT5"
	case texFmtLU_R5G6B5:
		return "LU_R5G6B5"
	case texFmtLU_A8R8G8B8:
		return "LU_A8R8G8B8"
	case texFmtSZ_A8:
		return "SZ_A8"
	case texFmtLU_X8R8G8B8:
		return "LU_X8R8G8B8"
	}
	return fmt.Sprintf("FMT_%02X", code)
}

// SurfaceAAScale is how many stored samples the current colour surface holds per logical
// pixel, horizontally and vertically.
//
// It matters to a debugger because the rasteriser and the picture do not share a
// coordinate space: fragments arrive at OnPixel in SAMPLE coordinates (the transform
// program's screen-space epilogue bakes the AA scale into its viewport, so positions
// arrive already scaled), while RenderDrawTarget resolves samples down to logical pixels.
// On a surface with anti-aliasing off the two coincide exactly — which is precisely why a
// caller that conflates them looks perfectly correct until the first anti-aliased scene,
// and then quietly attributes a quarter of the frame.
func (m *Machine) SurfaceAAScale() (ax, ay int) {
	return surfaceAAScale(m.pgraph.Regs[kelvinSurfaceFormat>>2])
}

// --- the bound textures, as pictures ---

// NVTextureUnits is how many texture units the Kelvin object has.
const NVTextureUnits = 4

// TextureBound reports whether unit u currently has a texture the pipeline would
// sample. A unit has to be ASKED: an untouched unit's registers read back zero, and a
// zero format word decodes as a 1x1 texture at address 0 rather than as nothing — so
// offering every unit unconditionally would put three phantom surfaces in the list.
// The test is the pipeline's own (nv2a_texture.go texStateDecode): the control
// register's enable bit AND a non-zero shader stage, because a unit enabled but not
// referenced by the shader program is not sampled either.
func (m *Machine) TextureBound(u int) bool {
	if u < 0 || u >= NVTextureUnits {
		return false
	}
	g := m.pgraph
	ctl := g.Regs[(kelvinTexControl+uint32(u)*0x40)>>2]
	stage := g.Regs[0x1E70>>2] >> (5 * uint(u)) & 0x1F
	return ctl>>30&1 != 0 && stage != 0
}

// DumpTexture decodes unit u's current binding to an image, through the pipeline's own
// decoder — so what the panel shows is exactly what a draw would sample.
func (m *Machine) DumpTexture(u int) (*image.RGBA, error) {
	if !m.TextureBound(u) {
		return nil, fmt.Errorf("xbox: texture unit %d is not bound", u)
	}
	w, h, pix, ok := m.pgraph.DebugDecodeTexture(u)
	if !ok {
		return nil, fmt.Errorf("xbox: texture unit %d did not decode", u)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix)
	return img, nil
}

// --- RAM as a picture ---

// RegionFormats are the pixel formats the free surface can decode. They are the
// uncompressed formats a title's surfaces and textures actually use, addressed
// linearly: a free view aims at an arbitrary address, where a swizzle or a DXT block
// layout would need a size the caller cannot generally know is right.
func RegionFormats() []string {
	return []string{"a8r8g8b8", "x8r8g8b8", "r5g6b5", "a1r5g5b5", "a4r4g4b4", "a8", "z24s8"}
}

// RegionSpec aims the free surface at memory. Addr is a guest address — any of the
// machine's RAM windows, or a plain physical address — because that is what the
// addresses in a title's own registers and pointers look like.
type RegionSpec struct {
	Addr   uint32
	W, H   int
	Stride int // bytes per row; 0 = tightly packed
	Format string
}

// RenderRegion decodes a span of RAM as an image. It reads memory directly rather than
// through the pipeline, so it works on any address at any time — including a buffer
// nothing has bound, which is usually the one you are looking for.
func (m *Machine) RenderRegion(s RegionSpec) (*image.RGBA, error) {
	bpp, ok := regionBPP(s.Format)
	if !ok {
		return nil, fmt.Errorf("xbox: no region format %q (have %v)", s.Format, RegionFormats())
	}
	if s.W <= 0 || s.H <= 0 {
		return nil, fmt.Errorf("xbox: a %dx%d region has nothing to show", s.W, s.H)
	}
	if s.W*s.H > 4<<20 {
		return nil, fmt.Errorf("xbox: a %dx%d region is too big to draw", s.W, s.H)
	}
	phys, mmio, ok := m.translate(s.Addr)
	if !ok || mmio {
		return nil, fmt.Errorf("xbox: %08X is not RAM", s.Addr)
	}
	stride := s.Stride
	if stride <= 0 {
		stride = s.W * bpp
	}
	img := image.NewRGBA(image.Rect(0, 0, s.W, s.H))
	for y := 0; y < s.H; y++ {
		row := int(phys) + y*stride
		for x := 0; x < s.W; x++ {
			o := row + x*bpp
			if o+bpp > len(m.RAM) {
				continue // past the end of RAM: leave it transparent rather than wrap
			}
			r, g, b, a := decodeRegionTexel(s.Format, m.RAM[o:o+bpp])
			i := img.PixOffset(x, y)
			img.Pix[i+0], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r, g, b, a
		}
	}
	return img, nil
}

func regionBPP(format string) (int, bool) {
	switch format {
	case "a8r8g8b8", "x8r8g8b8", "z24s8":
		return 4, true
	case "r5g6b5", "a1r5g5b5", "a4r4g4b4":
		return 2, true
	case "a8":
		return 1, true
	}
	return 0, false
}

// decodeRegionTexel decodes one texel. The 16- and 32-bit formats are little-endian in
// RAM, so an A8R8G8B8 texel's bytes arrive B,G,R,A.
func decodeRegionTexel(format string, b []byte) (r, g, bl, a byte) {
	switch format {
	case "a8r8g8b8":
		return b[2], b[1], b[0], b[3]
	case "x8r8g8b8":
		return b[2], b[1], b[0], 0xFF
	case "z24s8":
		// Depth is not a colour, so it is shown as one honestly: the 24-bit depth as
		// grey, the stencil byte in alpha. A depth buffer read as RGB is a picture of
		// nothing in particular; read as depth it is the frame's shape.
		d := uint32(b[0])<<8 | uint32(b[1])<<16 | uint32(b[2])<<24
		v := byte(d >> 24)
		return v, v, v, 0xFF
	case "r5g6b5":
		v := uint16(b[0]) | uint16(b[1])<<8
		return exp5(v >> 11), exp6(v >> 5 & 0x3F), exp5(v & 0x1F), 0xFF
	case "a1r5g5b5":
		v := uint16(b[0]) | uint16(b[1])<<8
		a = 0
		if v&0x8000 != 0 {
			a = 0xFF
		}
		return exp5(v >> 10 & 0x1F), exp5(v >> 5 & 0x1F), exp5(v & 0x1F), a
	case "a4r4g4b4":
		v := uint16(b[0]) | uint16(b[1])<<8
		ex := func(c uint16) byte { return byte(c<<4 | c) }
		return ex(v >> 8 & 0xF), ex(v >> 4 & 0xF), ex(v & 0xF), ex(v >> 12 & 0xF)
	case "a8":
		return b[0], b[0], b[0], 0xFF
	}
	return 0, 0, 0, 0xFF
}
