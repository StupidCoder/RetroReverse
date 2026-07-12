package psp

// ge_raster.go executes a captured GE display list into the framebuffer: it walks the
// commands maintaining render state (framebuffer target, viewport, matrices, vertex
// format, clear mode) and, on each PRIM, decodes the vertices and software-rasterizes
// the primitive. It covers the common path — framebuffer config, CLEAR fills,
// through-mode (2D screen-space) and transformed (3D) triangles/strips/fans/sprites
// with per-vertex or material colour. Texturing is sampled when a texture is bound.

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

// geDebugN: set PSP_GE_DEBUG=N to print the render state of the first N PRIMs.
var geDebugN, _ = strconv.Atoi(os.Getenv("PSP_GE_DEBUG"))

// geCullFlip / geNoCull: set PSP_GE_CULLFLIP=1 to invert the backface test and
// PSP_GE_NOCULL=1 to disable it, to settle the winding convention against
// rendered frames rather than assume it.
var geCullFlip = os.Getenv("PSP_GE_CULLFLIP") != ""
var geNoCull = os.Getenv("PSP_GE_NOCULL") != ""

// geNoZ: set PSP_GE_NOZ=1 to bypass the depth test (a bisection aid).
var geNoZ = os.Getenv("PSP_GE_NOZ") != ""

// geOnlyTex / geOnlyVA: set PSP_GE_ONLYTEX=<hex texaddr> or PSP_GE_ONLYVA=<hex
// prefix> to draw ONLY primitives bound to that texture / reading vertices from
// that 64 KiB region — isolates one object's geometry from everything drawn
// over it.
var geOnlyTex, _ = strconv.ParseUint(os.Getenv("PSP_GE_ONLYTEX"), 16, 32)
var geOnlyVA, _ = strconv.ParseUint(os.Getenv("PSP_GE_ONLYVA"), 16, 32)

// gePixelProbe: set PSP_GE_PIXEL=x,y to log every PRIM that writes that pixel
// (prim sequence number, render state, colour written). Correlate the seq
// numbers with the PSP_GE_DEBUG PRIM lines to find which draw owns a pixel.
var geProbeX, geProbeY = -1, -1

// gePrimDump: set PSP_GE_PRIM=N to dump, when PRIM sequence number N executes,
// the full render state, every vertex, and the raw command words issued since
// the previous PRIM (the state delta that conditioned this draw).
var gePrimDump = -1

func init() {
	if s := os.Getenv("PSP_GE_PIXEL"); s != "" {
		if n, err := fmt.Sscanf(s, "%d,%d", &geProbeX, &geProbeY); n != 2 || err != nil {
			geProbeX, geProbeY = -1, -1
		}
	}
	if s := os.Getenv("PSP_GE_PRIM"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			gePrimDump = v
		}
	}
}

// GE register command numbers (subset; pspsdk guInternal names).
const (
	cVADDR       = 0x01
	cIADDR       = 0x02
	cPRIM        = 0x04
	cBEZIER      = 0x05
	cSPLINE      = 0x06
	cBASE        = 0x10
	cLIGHTING    = 0x17
	cPATCHDIV    = 0x36 // bits 0-7 u divisions per span, 8-15 v divisions
	cPATCHPRIM   = 0x37 // 0 = triangles
	cMATDIFFUSE  = 0x56 // material diffuse RGB (the lit colour of colourless patches)
	cVTYPE       = 0x12
	cOFFADDR     = 0x13
	cCULLON      = 0x1D // backface cull enable
	cTEXENABLE   = 0x1E
	cBLENDON     = 0x21
	cCULLFACE    = 0x9B // which winding is the front face (0 = counter-clockwise)
	cTEXMTXN     = 0x40 // texture matrix write index
	cTEXMTXD     = 0x41 // texture matrix element
	cWORLDN      = 0x3A
	cWORLDD      = 0x3B
	cVIEWN       = 0x3C
	cVIEWD       = 0x3D
	cPROJN       = 0x3E
	cPROJD       = 0x3F
	cVPXSCALE    = 0x42
	cVPYSCALE    = 0x43
	cVPZSCALE    = 0x44
	cVPXCENTER   = 0x45
	cVPYCENTER   = 0x46
	cVPZCENTER   = 0x47
	cTEXSCALEU   = 0x48
	cTEXSCALEV   = 0x49
	cTEXOFFSETU  = 0x4A
	cTEXOFFSETV  = 0x4B
	cOFFSETX     = 0x4C
	cOFFSETY     = 0x4D
	cSHADEMODE   = 0x50 // 0 = flat, 1 = gouraud
	cMATAMBIENT  = 0x55 // material ambient RGB: the vertex colour when the format carries none
	cMATALPHA    = 0x58 // material ambient alpha
	cTEXADDR0    = 0xA0
	cTEXBW0      = 0xA8
	cCLUTADDR    = 0xB0
	cCLUTADDRH   = 0xB1
	cTEXSIZE0    = 0xB8
	cTEXMAPMODE  = 0xC0 // texcoord source: 0 = UV, 1 = texture matrix, 2 = env map
	cTEXMODE     = 0xC2 // bit 0 = swizzled, bits 16-18 = mip count
	cTEXFORMAT   = 0xC3
	cLOADCLUT    = 0xC4 // latches CLUT data from CLUTADDR (arg = 32-byte blocks)
	cCLUTFORMAT  = 0xC5
	cTEXFILTER   = 0xC6 // min/mag filter (bit 0 of each: 1 = linear)
	cTEXWRAP     = 0xC7 // u (bits 0-7) / v (bits 8-15) wrap: 0 = repeat, 1 = clamp
	cTEXFUNC     = 0xC9 // bits 0-2 func, bit 8 use texture alpha, bit 16 colour double
	cTEXENVCOL   = 0xCA
	cTRSRC       = 0xB2 // block transfer: source address (low 24 bits)
	cTRSRCW      = 0xB3 // source stride (bits 0-15) + high address byte (16-23)
	cTRDST       = 0xB4
	cTRDSTW      = 0xB5
	cTRSRCPOS    = 0xEB // source x (bits 0-9), y (bits 10-19)
	cTRDSTPOS    = 0xEC
	cTRSIZE      = 0xEE // width-1 (bits 0-9), height-1 (bits 10-19)
	cTRSTART     = 0xEA // bit 0: 0 = 16-bit pixels, 1 = 32-bit
	cFOGON       = 0x1F
	cFOG1        = 0xCD // fog end distance
	cFOG2        = 0xCE // fog scale = 1/(end-start)
	cFOGCOLOR    = 0xCF
	cLIGHT0ON    = 0x18 // lights 0-3 enable at 0x18..0x1B
	cAMBIENTCOL  = 0x5C // global ambient colour
	cMATEMISSIVE = 0x54
	cLIGHTTYPE0  = 0x5F // light types at 0x5F..0x62 (bits 8-9: 0 directional, 1 point, 2 spot)
	cLIGHTPOS0   = 0x63 // light positions: 3 words per light, 0x63..0x6E
	cLIGHTDIF0   = 0x90 // light diffuse colours: 3 words per light (amb/dif/spec) from 0x8F
	cSTENCILON   = 0x24 // stencil test enable
	cSTEST       = 0xDC // stencil test: bits 0-7 func, 8-15 ref, 16-23 mask
	cSOP         = 0xDD // stencil ops: bits 0-7 sfail, 8-15 zfail, 16-23 zpass
	cALPHATESTON = 0x22
	cATEST       = 0xDB // alpha test: bits 0-2 func, 8-15 ref, 16-23 mask
	cFBPTR       = 0x9C
	cFBWIDTH     = 0x9D
	cZBPTR       = 0x9E // depth buffer pointer (VRAM offset)
	cZBWIDTH     = 0x9F // depth buffer stride + high address bits
	cZTESTON     = 0x23 // depth test enable
	cFBPIXFMT    = 0xD2
	cCLEARMODE   = 0xD3
	cSCISSOR1    = 0xD4 // scissor top-left (x bits 0-9, y bits 10-19)
	cSCISSOR2    = 0xD5 // scissor bottom-right (inclusive)
	cZTEST       = 0xDE // depth test function (0 never .. 7 gequal)
	cBLENDFUNC   = 0xDF // blend: bits 0-3 src factor, 4-7 dst factor, 8-10 equation
	cBLENDFIXA   = 0xE0 // fixed source colour (factor 10)
	cBLENDFIXB   = 0xE1 // fixed destination colour (factor 10)
	cZMASK       = 0xE7 // depth write mask (1 = don't write z)
	cPMSKC       = 0xE8 // pixel colour write mask (1 bits = masked out; sceGuPixelMask)
	cPMSKA       = 0xE9 // pixel alpha write mask
)

// geState is the render state accumulated while walking a list.
type geState struct {
	fbLow, fbHigh uint32
	fbStride      uint32
	fbFmt         uint32

	base    uint32
	vaddr   uint32
	iaddr   uint32
	offAddr uint32
	vtype   uint32

	patchDivU, patchDivV uint32
	lightOn              bool
	matDiffuse           uint32 // material diffuse RGB (lit patches render with it)
	matEmissive          uint32
	ambientCol           uint32 // global ambient colour

	// Per-light state. Burnout lights its cars with one directional sun; the
	// unlit path (most of the world) bakes its light into vertex colours.
	lightEnable [4]bool
	lightType   [4]uint32
	lightPos    [4][3]float32
	lightDiff   [4]uint32

	vpXS, vpYS, vpZS float32 // viewport scale
	vpXC, vpYC, vpZC float32 // viewport center
	offX, offY       float32 // screen offset (OFFSETX/Y, 4-bit subpixel args)

	world, view, proj          [16]float32
	worldIdx, viewIdx, projIdx int

	matColor uint32 // material ambient RGBA (vertex colour when the format has none)
	clearOn  bool

	texEnable            bool
	texAddr              uint32
	texStride            uint32
	texW, texH           uint32
	texFmt               uint32
	texSwizzle           bool
	texScaleU, texScaleV float32
	texOffU, texOffV     float32

	clutAddr uint32
	clutFmt  uint32      // CLUTFORMAT: bits 0-1 entry format, 2-6 shift, 8-15 mask, 16-20 base
	clut     [256]uint32 // decoded RGBA entries

	blendOn   bool
	blendSrc  uint32 // BLENDFUNC src factor
	blendDst  uint32 // BLENDFUNC dst factor
	blendEq   uint32 // 0 add, 1 subtract, 2 reverse subtract, 3 min, 4 max, 5 abs
	blendFixA uint32 // fixed colours for factor 10
	blendFixB uint32

	alphaTestOn   bool
	alphaFunc     uint32 // 0 never, 1 always, 2 eq, 3 ne, 4 lt, 5 le, 6 gt, 7 ge
	alphaRef      uint32
	alphaTestMask uint32

	texMapMode uint32 // 0 = UV as submitted, 1 = texture matrix, 2 = env map
	texProjSrc uint32 // what the texture matrix transforms: 0 pos, 1 uv, 2 normalized normal, 3 normal
	texWrapU   uint32 // 0 = repeat, 1 = clamp
	texWrapV   uint32
	texLinear  bool   // TEXFILTER: bilinear magnification
	texFunc    uint32 // TEXFUNC: 0 modulate, 1 decal, 2 blend, 3 replace, 4 add
	texUseA    bool   // TEXFUNC bit 8: sample the texture's alpha
	texDouble  bool   // TEXFUNC bit 16: double the result (PSP's brightness trick)
	texEnvCol  uint32 // TEXENVCOLOR (the "blend" function's constant)
	texMtx     [16]float32
	texMtxIdx  int

	scX0, scY0, scX1, scY1 int // scissor rectangle (inclusive)

	maskRGB uint32 // PMSKC: colour bits masked out of framebuffer writes
	maskA   uint32 // PMSKA: alpha bits masked out

	// Depth buffer. A 3-D scene needs one: the front end got by on the painter's
	// algorithm, but the race draws the world in no particular order.
	zLow, zHigh uint32 // ZBP/ZBW: 16-bit depth buffer in VRAM
	zStride     uint32
	zTestOn     bool
	zFunc       uint32 // ZTST: 0 never, 1 always, 2 eq, 3 ne, 4 lt, 5 le, 6 gt, 7 ge
	zNoWrite    bool   // ZMSK
	clearDepth  bool   // CLEAR mode also clears the depth buffer (arg bit 10)

	cullOn   bool   // backface culling enabled
	cullFace uint32 // 0 = cull clockwise (screen-space) faces

	// The PSP's stencil buffer IS the framebuffer's alpha channel. Burnout uses
	// it as a MASK: most scene geometry replaces alpha with 0x00, a few objects
	// with 0xFF, and the post-process (the speed streaks, the bloom) reads that
	// mask back through the downsampled render target to know where it may draw.
	// Writing the fragment's own alpha instead let the streaks paint over the
	// whole car as white slashes.
	stencilOn                 bool
	stFunc, stRef, stMask     uint32
	stSFail, stZFail, stZPass uint32

	fogOn    bool
	fogEnd   float32
	fogScale float32
	fogColor uint32 // 0xBBGGRR

	// GE block transfer (sceGuCopyImage): Burnout downsamples its back buffer
	// into a 64x64 off-screen target and composites it back for the speed
	// streaks. Never running the copy left that target holding uninitialised
	// VRAM, which the streak quads then painted over the car as white slashes.
	trSrc, trDst   uint32
	trSrcStride    uint32
	trDstStride    uint32
	trSrcX, trSrcY uint32
	trDstX, trDstY uint32
	trW, trH       uint32
}

func (s *geState) zAddress() uint32 { return vramBase | (s.zHigh << 24) | s.zLow }

// blockTransfer runs the GE's 2-D block copy (GE_CMD_TRANSFERSTART): a
// rectangle of trW x trH pixels from (trSrcX, trSrcY) in the source buffer to
// (trDstX, trDstY) in the destination, at 16 or 32 bits per pixel.
func (m *Machine) blockTransfer(s *geState, is32 bool) {
	if s.trSrcStride == 0 || s.trDstStride == 0 || s.trW == 0 || s.trH == 0 {
		return
	}
	bpp := uint32(2)
	if is32 {
		bpp = 4
	}
	for y := uint32(0); y < s.trH; y++ {
		src := s.trSrc + ((s.trSrcY+y)*s.trSrcStride+s.trSrcX)*bpp
		dst := s.trDst + ((s.trDstY+y)*s.trDstStride+s.trDstX)*bpp
		for x := uint32(0); x < s.trW*bpp; x++ {
			m.Write(dst+x, m.Read(src+x))
		}
	}
}

// zPass runs the depth test for one pixel and, when it passes, writes z back
// (unless ZMSK masks the write). Depth is 16-bit, one entry per zStride pixels.
// In CLEAR mode the test is bypassed: the clear either stamps z (bit 10 set) or
// leaves the buffer alone.
func (m *Machine) zPass(s *geState, x, y int, z float32) bool {
	if s.zLow == 0 || s.zStride == 0 {
		return true
	}
	zi := uint32(clampF(z, 0, 65535))
	addr := s.zAddress() + (uint32(y)*s.zStride+uint32(x))*2
	if s.clearOn {
		if s.clearDepth {
			m.write16(addr, uint16(zi))
		}
		return true
	}
	if !s.zTestOn || geNoZ {
		return true
	}
	old := uint32(u16(m, addr))
	pass := false
	switch s.zFunc {
	case 0: // never
	case 1: // always
		pass = true
	case 2:
		pass = zi == old
	case 3:
		pass = zi != old
	case 4:
		pass = zi < old
	case 5:
		pass = zi <= old
	case 6:
		pass = zi > old
	default: // 7: gequal — Burnout's convention (viewport z is reversed: near = 65535)
		pass = zi >= old
	}
	if pass && !s.zNoWrite {
		m.write16(addr, uint16(zi))
	}
	return pass
}

func clampF(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (m *Machine) rasterList(list GeList) {
	// GE registers persist across list submissions (a frame's FBP/matrix setup
	// list conditions the draw lists that follow), so the state lives on the
	// machine and is created once.
	if m.geSt == nil {
		s := &geState{}
		ident(&s.world)
		ident(&s.view)
		ident(&s.proj)
		s.vpXS, s.vpYS = dispW/2, -dispH/2
		s.vpXC, s.vpYC = 2048+dispW/2, 2048+dispH/2
		s.offX, s.offY = 2048, 2048
		s.texScaleU, s.texScaleV = 1, 1
		s.matColor = 0xFFFFFFFF
		// Stencil, like ZTST, is enabled once at engine init and never re-sent in
		// a mid-run list. The game programs STENCILTEST/STENCILOP thousands of
		// times per frame, which only means anything with the test enabled.
		s.stencilOn = true
		s.stFunc = 1 // always
		s.stMask = 0xFF
		// ZTST is programmed once at engine init, so a list captured mid-run
		// never re-sends it. GEQUAL is what the viewport implies: Burnout maps
		// z with scale -32767.5 / center +32767.5, i.e. near = 65535, so the
		// nearer fragment is the one with the LARGER depth value.
		s.zFunc = 7
		// Scissor defaults to the whole screen; the game programs it per target.
		s.scX1, s.scY1 = dispW-1, dispH-1
		m.geSt = s
	}
	s := m.geSt

	for _, w := range list.Words {
		cmd := w >> 24
		arg := w & 0x00FFFFFF
		if gePrimDump >= 0 {
			geWordTrail = append(geWordTrail, w)
			if len(geWordTrail) > 400 {
				geWordTrail = geWordTrail[len(geWordTrail)-400:]
			}
		}
		switch cmd {
		case cBASE:
			s.base = geBaseAddr(arg)
		case cVADDR:
			s.vaddr = s.base | arg
		case cIADDR:
			s.iaddr = s.base | arg
		case cOFFADDR:
			s.offAddr = arg << 8
		case cLIGHTING:
			s.lightOn = arg&1 != 0
		case cAMBIENTCOL:
			s.ambientCol = arg
		case cMATEMISSIVE:
			s.matEmissive = arg
		case cLIGHT0ON, cLIGHT0ON + 1, cLIGHT0ON + 2, cLIGHT0ON + 3:
			s.lightEnable[cmd-cLIGHT0ON] = arg&1 != 0
		case cLIGHTTYPE0, cLIGHTTYPE0 + 1, cLIGHTTYPE0 + 2, cLIGHTTYPE0 + 3:
			s.lightType[cmd-cLIGHTTYPE0] = (arg >> 8) & 3
		case cLIGHTPOS0, cLIGHTPOS0 + 1, cLIGHTPOS0 + 2, cLIGHTPOS0 + 3, cLIGHTPOS0 + 4,
			cLIGHTPOS0 + 5, cLIGHTPOS0 + 6, cLIGHTPOS0 + 7, cLIGHTPOS0 + 8,
			cLIGHTPOS0 + 9, cLIGHTPOS0 + 10, cLIGHTPOS0 + 11:
			i := (cmd - cLIGHTPOS0) / 3
			s.lightPos[i][(cmd-cLIGHTPOS0)%3] = math.Float32frombits(arg << 8)
		case cLIGHTDIF0, cLIGHTDIF0 + 3, cLIGHTDIF0 + 6, cLIGHTDIF0 + 9:
			s.lightDiff[(cmd-cLIGHTDIF0)/3] = arg
		case cPATCHDIV:
			s.patchDivU = arg & 0xFF
			s.patchDivV = (arg >> 8) & 0xFF
		case cMATDIFFUSE:
			s.matDiffuse = arg
		case cVTYPE:
			s.vtype = arg
		case cFBPTR:
			s.fbLow = arg
		case cFBWIDTH:
			s.fbStride = arg & 0xFFFF
			s.fbHigh = (arg >> 16) & 0xFF
		case cCULLON:
			s.cullOn = arg&1 != 0
		case cCULLFACE:
			s.cullFace = arg & 1
		case cZBPTR:
			s.zLow = arg
		case cZBWIDTH:
			s.zStride = arg & 0xFFFF
			s.zHigh = (arg >> 16) & 0xFF
		case cZTESTON:
			s.zTestOn = arg&1 != 0
		case cZTEST:
			s.zFunc = arg & 7
		case cZMASK:
			s.zNoWrite = arg&1 != 0
		case cTRSRC:
			s.trSrc = (s.trSrc & 0xFF000000) | arg
		case cTRSRCW:
			s.trSrcStride = arg & 0xFFFF
			s.trSrc = (s.trSrc & 0x00FFFFFF) | ((arg >> 16) & 0xFF << 24)
		case cTRDST:
			s.trDst = (s.trDst & 0xFF000000) | arg
		case cTRDSTW:
			s.trDstStride = arg & 0xFFFF
			s.trDst = (s.trDst & 0x00FFFFFF) | ((arg >> 16) & 0xFF << 24)
		case cTRSRCPOS:
			s.trSrcX, s.trSrcY = arg&0x3FF, (arg>>10)&0x3FF
		case cTRDSTPOS:
			s.trDstX, s.trDstY = arg&0x3FF, (arg>>10)&0x3FF
		case cTRSIZE:
			s.trW, s.trH = (arg&0x3FF)+1, ((arg>>10)&0x3FF)+1
		case cTRSTART:
			m.blockTransfer(s, arg&1 != 0)
		case cFOGON:
			s.fogOn = arg&1 != 0
		case cFOG1:
			s.fogEnd = math.Float32frombits(arg << 8)
		case cFOG2:
			s.fogScale = math.Float32frombits(arg << 8)
		case cFOGCOLOR:
			s.fogColor = arg
		case cSTENCILON:
			s.stencilOn = arg&1 != 0
		case cSTEST:
			s.stFunc = arg & 7
			s.stRef = (arg >> 8) & 0xFF
			s.stMask = (arg >> 16) & 0xFF
		case cSOP:
			s.stSFail = arg & 0xFF
			s.stZFail = (arg >> 8) & 0xFF
			s.stZPass = (arg >> 16) & 0xFF
		case cALPHATESTON:
			s.alphaTestOn = arg&1 != 0
		case cATEST:
			s.alphaFunc = arg & 7
			s.alphaRef = (arg >> 8) & 0xFF
			s.alphaTestMask = (arg >> 16) & 0xFF
		case cBLENDFUNC:
			s.blendSrc = arg & 0xF
			s.blendDst = (arg >> 4) & 0xF
			s.blendEq = (arg >> 8) & 7
		case cBLENDFIXA:
			s.blendFixA = arg
		case cBLENDFIXB:
			s.blendFixB = arg
		case cSCISSOR1:
			s.scX0, s.scY0 = int(arg&0x3FF), int((arg>>10)&0x3FF)
		case cSCISSOR2:
			s.scX1, s.scY1 = int(arg&0x3FF), int((arg>>10)&0x3FF)
		case cTEXMAPMODE:
			s.texMapMode = arg & 3
			s.texProjSrc = (arg >> 8) & 3
		case cTEXWRAP:
			s.texWrapU, s.texWrapV = arg&0xFF, (arg>>8)&0xFF
		case cTEXFILTER:
			s.texLinear = (arg>>8)&1 != 0 // magnification filter
		case cTEXFUNC:
			s.texFunc = arg & 7
			s.texUseA = (arg>>8)&1 != 0
			s.texDouble = (arg>>16)&1 != 0
		case cTEXENVCOL:
			s.texEnvCol = arg
		case cTEXMTXN:
			s.texMtxIdx = int(arg & 0xF)
		case cTEXMTXD:
			matPush(&s.texMtx, &s.texMtxIdx, arg)
		case cFBPIXFMT:
			s.fbFmt = arg & 3
		case cCLEARMODE:
			s.clearOn = arg&1 != 0
			// CLEAR mode: bits 8-10 select the buffers the clear writes
			// (colour / alpha / depth). Depth is written straight from the
			// vertex z with the test bypassed.
			s.clearDepth = arg&0x400 != 0
		case cMATAMBIENT:
			s.matColor = (s.matColor & 0xFF000000) | arg
		case cMATALPHA:
			s.matColor = (s.matColor & 0x00FFFFFF) | (arg&0xFF)<<24
		case cBLENDON:
			s.blendOn = arg&1 != 0
		case cVPXSCALE:
			s.vpXS = f24(arg)
		case cVPYSCALE:
			s.vpYS = f24(arg)
		case cVPZSCALE:
			s.vpZS = f24(arg)
		case cVPXCENTER:
			s.vpXC = f24(arg)
		case cVPYCENTER:
			s.vpYC = f24(arg)
		case cVPZCENTER:
			s.vpZC = f24(arg)
		case cOFFSETX:
			s.offX = float32(arg) / 16 // 4-bit subpixel fixed point
		case cOFFSETY:
			s.offY = float32(arg) / 16
		case cWORLDN:
			s.worldIdx = int(arg & 0xF) // write index 0-11 into the 4x3 matrix
		case cWORLDD:
			matPush(&s.world, &s.worldIdx, arg)
		case cVIEWN:
			s.viewIdx = int(arg & 0xF)
		case cVIEWD:
			matPush(&s.view, &s.viewIdx, arg)
		case cPROJN:
			s.projIdx = int(arg & 0x1F) // write index 0-15 into the 4x4 matrix
		case cPROJD:
			matPush16(&s.proj, &s.projIdx, arg)
		case cTEXENABLE:
			s.texEnable = arg&1 != 0
		case cTEXADDR0:
			s.texAddr = (s.texAddr & 0xFF000000) | (arg & 0xFFFFFF)
		case cTEXBW0:
			s.texStride = arg & 0xFFFF
			s.texAddr = (s.texAddr & 0x00FFFFFF) | ((arg & 0xFF0000) << 8)
		case cTEXSIZE0:
			s.texW = 1 << (arg & 0xFF)
			s.texH = 1 << ((arg >> 8) & 0xFF)
		case cTEXFORMAT:
			s.texFmt = arg & 0xF
		case cTEXMODE:
			s.texSwizzle = arg&1 != 0
		case cTEXSCALEU:
			s.texScaleU = f24(arg)
		case cTEXSCALEV:
			s.texScaleV = f24(arg)
		case cTEXOFFSETU:
			s.texOffU = f24(arg)
		case cTEXOFFSETV:
			s.texOffV = f24(arg)
		case cCLUTADDR:
			s.clutAddr = (s.clutAddr & 0xFF000000) | (arg & 0xFFFFFF)
		case cCLUTADDRH:
			s.clutAddr = (s.clutAddr & 0x00FFFFFF) | ((arg & 0xFF0000) << 8)
		case cCLUTFORMAT:
			s.clutFmt = arg
		case cLOADCLUT:
			m.loadClut(s, arg&0x3F)
		case cPMSKC:
			s.maskRGB = arg
		case cPMSKA:
			s.maskA = arg & 0xFF
		case cPRIM:
			m.drawPrim(s, arg)
		case cBEZIER, cSPLINE:
			m.drawPatch(s, arg, cmd == cSPLINE)
		}
	}
	if s.fbStride != 0 {
		m.fbAddr = s.fbAddress()
		m.fbWidth = s.fbStride
		m.fbFormat = s.fbFmt
	}
}

// fbAddress is the CPU address of the current framebuffer.
func (s *geState) fbAddress() uint32 { return vramBase | (s.fbHigh << 24) | s.fbLow }

// loadClut latches CLUT entries from CLUTADDR into the state, decoded to RGBA.
// blocks is the LOADCLUT argument: 32-byte units (8 entries of 32-bit, or 16 of
// 16-bit, per the CLUTFORMAT entry format).
func (m *Machine) loadClut(s *geState, blocks uint32) {
	if s.clutAddr == 0 {
		return
	}
	entryFmt := s.clutFmt & 3
	n := blocks * 8
	if entryFmt != 3 {
		n = blocks * 16
	}
	if n > 256 {
		n = 256
	}
	for i := uint32(0); i < n; i++ {
		if entryFmt == 3 { // 8888 (RGBA byte order)
			s.clut[i] = m.read32(s.clutAddr + i*4)
		} else {
			r, g, b, a := decode16a(u16(m, s.clutAddr+i*2), entryFmt)
			s.clut[i] = uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
		}
	}
}

// gePrimSeq numbers every PRIM drawn since process start, so the pixel probe
// and the PSP_GE_DEBUG lines can be correlated.
var gePrimSeq int

// geWordTrail keeps the raw command words seen since the previous PRIM, for the
// PSP_GE_PRIM deep dump.
var geWordTrail []uint32

// dumpPrim prints the full render state, the command trail since the previous
// PRIM, and every decoded vertex of the target PRIM.
func (m *Machine) dumpPrim(s *geState, ptype uint32, count int) {
	fmt.Printf("=== PRIM#%d t%d n%d ===\n", gePrimSeq, ptype, count)
	fmt.Printf("state: vt=%06X va=%08X clear=%v blend=%v texEn=%v tex@%08X f%d sw=%v %dx%d stride=%d clut@%08X clutFmt=%06X mat=%08X fb=%08X/%d fmt=%d\n",
		s.vtype, s.vaddr, s.clearOn, s.blendOn, s.texEnable, s.texAddr, s.texFmt,
		s.texSwizzle, s.texW, s.texH, s.texStride, s.clutAddr, s.clutFmt, s.matColor,
		s.fbAddress(), s.fbStride, s.fbFmt)
	fmt.Printf("viewport: S(%.1f,%.1f,%.1f) C(%.1f,%.1f,%.1f) off(%.1f,%.1f)\n",
		s.vpXS, s.vpYS, s.vpZS, s.vpXC, s.vpYC, s.vpZC, s.offX, s.offY)
	fmt.Printf("trail (%d words since previous PRIM):\n", len(geWordTrail))
	for _, w := range geWordTrail {
		fmt.Printf("  %08X  %s %06X\n", w, GeCmdName(w>>24), w&0xFFFFFF)
	}
	for i, v := range m.decodeVerts(s, count) {
		fmt.Printf("  v%-3d (%8.2f,%8.2f,%8.2f) uv(%.3f,%.3f) rgba %02X%02X%02X%02X\n",
			i, v.x, v.y, v.z, v.u, v.v, v.r, v.g, v.b, v.a)
	}
}

// drawPrim decodes and rasterizes one PRIM command (arg: bits 0-15 count, 16-18 type).
func (m *Machine) drawPrim(s *geState, arg uint32) {
	count := int(arg & 0xFFFF)
	ptype := (arg >> 16) & 7
	if count == 0 || s.vaddr == 0 {
		return
	}
	gePrimSeq++
	if gePrimSeq == gePrimDump {
		m.dumpPrim(s, ptype, count)
	}
	if gePrimDump >= 0 {
		geWordTrail = geWordTrail[:0]
	}
	if geDebugN > 0 {
		geDebugN--
		vs := m.decodeVerts(s, min(count, 2))
		var v0 vert
		if len(vs) > 0 {
			v0 = vs[0]
		}
		fmt.Printf("PRIM#%d t%d n%d vt=%06X va=%08X tex=%v@%08X f%d sw%v %dx%d clut@%08X mat=%08X wT=(%.1f,%.1f,%.1f) v0=(%.1f,%.1f,%.1f uv %.2f,%.2f)\n",
			gePrimSeq, ptype, count, s.vtype, s.vaddr, s.texEnable, s.texAddr, s.texFmt, s.texSwizzle,
			s.texW, s.texH, s.clutAddr, s.matColor,
			s.world[12], s.world[13], s.world[14], v0.x, v0.y, v0.z, v0.u, v0.v)
	}
	if geOnlyTex != 0 && uint64(s.texAddr) != geOnlyTex {
		return
	}
	if geOnlyVA != 0 && uint64(s.vaddr)>>16 != geOnlyVA>>16 {
		return
	}
	verts := m.decodeVerts(s, count)

	// THE GE CONSUMES VERTICES: its vertex pointer advances by the vertices a
	// primitive reads, so a mesh submitted as a run of PRIMs sets VADDR once and
	// every following PRIM continues where the last left off. Burnout draws its
	// cars that way — 10,416 primitives a frame behind a single VADDR. Re-reading
	// from the same address each time drew the mesh's first strip over and over
	// (the diagonal slashes across the car) and never drew the rest of it (the
	// missing half). Indexed primitives advance the INDEX pointer instead; the
	// vertex base stays put.
	if l := layoutOf(s.vtype); l.stride != 0 {
		if l.idxFmt != 0 {
			s.iaddr += uint32(count) * l.idxFmt // idxFmt 1 = u8, 2 = u16
		} else {
			s.vaddr += uint32(count) * l.stride
		}
	}

	if len(verts) < 1 {
		return
	}
	switch ptype {
	case 3: // triangles
		for i := 0; i+2 < len(verts); i += 3 {
			m.rasterTriClipped(s, verts[i], verts[i+1], verts[i+2])
		}
	case 4: // triangle strip
		for i := 0; i+2 < len(verts); i++ {
			a, b, c := verts[i], verts[i+1], verts[i+2]
			if i&1 == 1 {
				b, c = c, b
			}
			m.rasterTriClipped(s, a, b, c)
		}
	case 5: // triangle fan
		for i := 1; i+1 < len(verts); i++ {
			m.rasterTriClipped(s, verts[0], verts[i], verts[i+1])
		}
	case 6: // sprites (pairs of corners)
		for i := 0; i+1 < len(verts); i += 2 {
			m.rasterSprite(s, verts[i], verts[i+1])
		}
	}
}

// vert is a decoded, screen-space vertex. For transformed primitives it also
// carries its clip-space coordinates, which the near-plane clipper needs: a
// vertex behind the eye has w <= 0 and its projection is meaningless (it wraps
// to a huge screen coordinate), so such triangles must be clipped, not drawn.
type vert struct {
	x, y, z    float32
	u, v       float32
	r, g, b, a byte

	cx, cy, cz, cw float32 // clip space (before the perspective divide)
	invW           float32 // 1/w — the weight perspective-correct interpolation needs
	fog            float32 // fog coefficient: 1 = unfogged, 0 = fully fogged
	clip           bool    // clip coords are valid (non-through vertex)
}

// through reports whether the vertex bypassed the transform (screen coords as
// submitted): those are never culled.
func (v vert) through() bool { return !v.clip }

// decodeVerts reads count vertices from s.vaddr per the vertex type, applying the
// transform (world*view*proj*viewport) for non-through primitives.
// vertexLayout describes how one vertex of the current VTYPE is packed.
type vertexLayout struct {
	tfmt, cfmt, nfmt, pfmt, idxFmt uint32
	offTex, offCol, offNrm, offPos uint32
	stride                         uint32
	through                        bool
}

// layoutOf computes the packing of the current vertex type. A PSP vertex lays
// its components out in a fixed order — weights, texcoord, colour, normal,
// position — each aligned to its own element size, and the whole vertex padded
// to the largest of those.
func layoutOf(vtype uint32) vertexLayout {
	var l vertexLayout
	l.tfmt = vtype & 3
	l.cfmt = (vtype >> 2) & 7
	l.nfmt = (vtype >> 5) & 3
	l.pfmt = (vtype >> 7) & 3
	wfmt := (vtype >> 9) & 3
	l.idxFmt = (vtype >> 11) & 3
	wcount := ((vtype >> 14) & 7) + 1
	l.through = vtype&(1<<23) != 0

	elem := []uint32{0, 1, 2, 4} // component element size by format (u8/u16/float)
	texSz := elem[l.tfmt] * 2
	colSz := map[uint32]uint32{0: 0, 4: 2, 5: 2, 6: 2, 7: 4}[l.cfmt]
	nrmSz := elem[l.nfmt] * 3
	posSz := elem[l.pfmt] * 3
	wgtSz := elem[wfmt] * wcount
	if wfmt == 0 {
		wgtSz = 0
	}

	align := uint32(1)
	for _, e := range []uint32{elem[l.tfmt], elem[l.nfmt], elem[l.pfmt], elem[wfmt], colSz} {
		if e > align {
			align = e
		}
	}
	pad := func(off, a uint32) uint32 {
		if a == 0 {
			return off
		}
		return (off + a - 1) &^ (a - 1)
	}
	off := pad(0, elem[wfmt]) + wgtSz
	l.offTex = pad(off, elem[l.tfmt])
	off = l.offTex + texSz
	l.offCol = pad(off, colSz)
	off = l.offCol + colSz
	l.offNrm = pad(off, elem[l.nfmt])
	off = l.offNrm + nrmSz
	l.offPos = pad(off, elem[l.pfmt])
	off = l.offPos + posSz
	l.stride = pad(off, align)
	return l
}

func (m *Machine) decodeVerts(s *geState, count int) []vert {
	l := layoutOf(s.vtype)
	tfmt, cfmt, nfmt, pfmt, idxFmt := l.tfmt, l.cfmt, l.nfmt, l.pfmt, l.idxFmt
	offTex, offCol, offNrm, offPos := l.offTex, l.offCol, l.offNrm, l.offPos
	stride, through := l.stride, l.through
	if stride == 0 {
		return nil
	}

	mvp := mul4(mul4(s.proj, s.view), s.world)
	out := make([]vert, 0, count)
	addr := s.vaddr
	for i := 0; i < count; i++ {
		p := addr
		// indexed vertices: the i-th index (u8/u16 at IADDR) selects the vertex
		switch idxFmt {
		case 1:
			p = s.vaddr + uint32(m.Read(s.iaddr+uint32(i)))*stride
		case 2:
			p = s.vaddr + uint32(u16(m, s.iaddr+uint32(i)*2))*stride
		}
		var vv vert
		vv.a = 0xFF
		if tfmt != 0 {
			vv.u, vv.v = readUV(m, p+offTex, tfmt)
			if through {
				// through-mode texcoords are absolute texels; normalize for sampling
				if s.texW != 0 && s.texH != 0 {
					vv.u /= float32(s.texW)
					vv.v /= float32(s.texH)
				}
			} else {
				// transformed-mode fixed-point texcoords are fractional (u8/128,
				// s16/32768), then scaled and offset by the texture matrix registers
				switch tfmt {
				case 1:
					vv.u /= 128
					vv.v /= 128
				case 2:
					vv.u /= 32768
					vv.v /= 32768
				}
				vv.u = vv.u*s.texScaleU + s.texOffU
				vv.v = vv.v*s.texScaleV + s.texOffV
			}
		}
		// TEXMAPMODE 1: the texcoords are NOT the submitted UVs — they are a
		// vertex attribute (chosen by the projection-source field) run through
		// the texture matrix, then divided by the projected q. Burnout maps its
		// car's reflections this way, with the NORMAL as the source: those
		// vertices all carry the same dummy UV, so reading it gave one flat
		// bright texel and painted white ribbons across the car.
		if !through && s.texMapMode == 1 {
			var sx, sy, sz float32
			switch s.texProjSrc {
			case 1: // the submitted texcoords
				sx, sy, sz = vv.u, vv.v, 0
			case 2, 3: // the normal (normalized for source 2)
				sx, sy, sz = readNormal(m, p+offNrm, nfmt)
				if s.texProjSrc == 2 {
					if l := float32(math.Sqrt(float64(sx*sx + sy*sy + sz*sz))); l > 1e-8 {
						sx, sy, sz = sx/l, sy/l, sz/l
					}
				}
			default: // the model position
				sx, sy, sz = readPos(m, p+offPos, pfmt)
				switch pfmt {
				case 1:
					sx, sy, sz = sx/128, sy/128, sz/128
				case 2:
					sx, sy, sz = sx/32768, sy/32768, sz/32768
				}
			}
			tu := s.texMtx[0]*sx + s.texMtx[4]*sy + s.texMtx[8]*sz + s.texMtx[12]
			tv := s.texMtx[1]*sx + s.texMtx[5]*sy + s.texMtx[9]*sz + s.texMtx[13]
			tq := s.texMtx[2]*sx + s.texMtx[6]*sy + s.texMtx[10]*sz + s.texMtx[14]
			if tq > 1e-6 || tq < -1e-6 {
				tu, tv = tu/tq, tv/tq
			}
			vv.u, vv.v = tu, tv
		}
		if cfmt != 0 {
			vv.r, vv.g, vv.b, vv.a = readColor(m, p+offCol, cfmt)
		} else {
			vv.r, vv.g, vv.b = byte(s.matColor), byte(s.matColor>>8), byte(s.matColor>>16)
			vv.a = byte(s.matColor >> 24)
		}
		// Per-vertex lighting. Most of Burnout's world is drawn UNLIT (its light
		// is baked into the vertex colours), but the cars are lit by a single
		// directional sun; unlit, they render as flat dark shapes.
		if s.lightOn && !through && nfmt != 0 {
			nx, ny, nz := readNormal(m, p+offNrm, nfmt)
			vv.r, vv.g, vv.b = s.lightVertex(nx, ny, nz)
		}
		px, py, pz := readPos(m, p+offPos, pfmt)
		if through {
			// through-mode positions are absolute screen coordinates
			vv.x, vv.y, vv.z = px, py, pz
		} else {
			// transformed-mode fixed-point positions are FRACTIONAL, like the
			// texcoords: s8/128, s16/32768. The model matrix carries the scale
			// back. Burnout's track and cars are s16 (vtype 0x116).
			switch pfmt {
			case 1:
				px, py, pz = px/128, py/128, pz/128
			case 2:
				px, py, pz = px/32768, py/32768, pz/32768
			}
			vv.cx, vv.cy, vv.cz, vv.cw = clipCoords(mvp, px, py, pz)
			vv.clip = true
			vv.x, vv.y, vv.z = project(s, vv.cx, vv.cy, vv.cz, vv.cw)
			vv.invW = invW(vv.cw)
			vv.fog = fogCoef(s, vv.cw)
		}
		out = append(out, vv)
		addr += stride
	}
	return out
}

func readUV(m *Machine, p, fmt uint32) (float32, float32) {
	switch fmt {
	case 1:
		return float32(m.Read(p)), float32(m.Read(p + 1))
	case 2:
		return float32(u16(m, p)), float32(u16(m, p+2))
	default:
		return f32(m, p), f32(m, p+4)
	}
}

func readColor(m *Machine, p, fmt uint32) (byte, byte, byte, byte) {
	switch fmt {
	case 7: // 8888
		c := m.read32(p)
		return byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)
	case 6: // 4444
		c := u16(m, p)
		e := func(v uint16) byte { return byte((v & 0xF) * 17) }
		return e(c), e(c >> 4), e(c >> 8), e(c >> 12)
	case 5: // 5551
		c := u16(m, p)
		e := func(v uint16, b uint16) byte { return byte(uint32(v&((1<<b)-1)) * 255 / ((1 << b) - 1)) }
		a := byte(0)
		if c&0x8000 != 0 {
			a = 0xFF
		}
		return e(c, 5), e(c>>5, 5), e(c>>10, 5), a
	default: // 565
		c := u16(m, p)
		e := func(v uint16, b uint16) byte { return byte(uint32(v&((1<<b)-1)) * 255 / ((1 << b) - 1)) }
		return e(c, 5), e(c>>5, 6), e(c>>11, 5), 0xFF
	}
}

// lightVertex evaluates the lighting model for one vertex normal: the emissive
// term, the global ambient against the material ambient, and each enabled
// light's diffuse against the material diffuse by Lambert's cosine. Specular
// and the attenuation/spot terms are not modelled.
func (s *geState) lightVertex(nx, ny, nz float32) (byte, byte, byte) {
	// The normal is a model-space direction: rotate it by the world matrix
	// (its 3x3 part) to reach the space the lights live in.
	wx := s.world[0]*nx + s.world[4]*ny + s.world[8]*nz
	wy := s.world[1]*nx + s.world[5]*ny + s.world[9]*nz
	wz := s.world[2]*nx + s.world[6]*ny + s.world[10]*nz
	if l := float32(math.Sqrt(float64(wx*wx + wy*wy + wz*wz))); l > 1e-8 {
		wx, wy, wz = wx/l, wy/l, wz/l
	}
	chOf := func(c uint32, i int) float32 { return float32(byte(c>>(8*i))) / 255 }

	var out [3]float32
	for i := 0; i < 3; i++ {
		out[i] = chOf(s.matEmissive, i) + chOf(s.ambientCol, i)*chOf(s.matColor, i)
	}
	for li := 0; li < 4; li++ {
		if !s.lightEnable[li] {
			continue
		}
		lx, ly, lz := s.lightPos[li][0], s.lightPos[li][1], s.lightPos[li][2]
		if l := float32(math.Sqrt(float64(lx*lx + ly*ly + lz*lz))); l > 1e-8 {
			lx, ly, lz = lx/l, ly/l, lz/l
		}
		nd := wx*lx + wy*ly + wz*lz
		if nd <= 0 {
			continue
		}
		for i := 0; i < 3; i++ {
			out[i] += chOf(s.lightDiff[li], i) * chOf(s.matDiffuse, i) * nd
		}
	}
	return clamp255(int32(out[0] * 255)), clamp255(int32(out[1] * 255)), clamp255(int32(out[2] * 255))
}

// readNormal reads a vertex normal; like positions, the fixed-point forms are
// fractional (s8/128, s16/32768).
func readNormal(m *Machine, p, fmt uint32) (float32, float32, float32) {
	switch fmt {
	case 1: // s8
		return float32(int8(m.Read(p))) / 128, float32(int8(m.Read(p+1))) / 128, float32(int8(m.Read(p+2))) / 128
	case 2: // s16
		return float32(int16(u16(m, p))) / 32768, float32(int16(u16(m, p+2))) / 32768, float32(int16(u16(m, p+4))) / 32768
	case 3:
		return f32(m, p), f32(m, p+4), f32(m, p+8)
	}
	return 0, 0, 0
}

func readPos(m *Machine, p, fmt uint32) (float32, float32, float32) {
	switch fmt {
	case 1: // s8
		return float32(int8(m.Read(p))), float32(int8(m.Read(p + 1))), float32(int8(m.Read(p + 2)))
	case 2: // s16
		return float32(int16(u16(m, p))), float32(int16(u16(m, p+2))), float32(int16(u16(m, p+4)))
	default: // float
		return f32(m, p), f32(m, p+4), f32(m, p+8)
	}
}

// clipCoords applies the model-view-projection, leaving the vertex in clip space.
func clipCoords(mvp [16]float32, x, y, z float32) (float32, float32, float32, float32) {
	return mvp[0]*x + mvp[4]*y + mvp[8]*z + mvp[12],
		mvp[1]*x + mvp[5]*y + mvp[9]*z + mvp[13],
		mvp[2]*x + mvp[6]*y + mvp[10]*z + mvp[14],
		mvp[3]*x + mvp[7]*y + mvp[11]*z + mvp[15]
}

// project applies the perspective divide and the viewport transform.
func project(s *geState, cx, cy, cz, cw float32) (float32, float32, float32) {
	if cw == 0 {
		cw = 1e-6
	}
	nx, ny, nz := cx/cw, cy/cw, cz/cw
	sx := nx*s.vpXS + s.vpXC - s.offX
	sy := ny*s.vpYS + s.vpYC - s.offY
	sz := nz*s.vpZS + s.vpZC
	return sx, sy, sz
}

// fogCoef computes the per-vertex fog coefficient from the view-space depth
// (the clip w): 1 where the vertex is nearer than the fog start, falling to 0
// at the fog end. Burnout fogs its city to a pale blue-grey; without it the
// distance stays flat black.
func fogCoef(s *geState, w float32) float32 {
	if !s.fogOn {
		return 1
	}
	f := (s.fogEnd - w) * s.fogScale
	return clampF(f, 0, 1)
}

// invW returns 1/w, the weight that makes texture interpolation
// perspective-correct. Through-mode vertices (w unset) get 1.
func invW(w float32) float32 {
	if w == 0 {
		return 1
	}
	return 1 / w
}

// wNear is the near-plane epsilon in clip space: a vertex at or behind the eye
// (w <= wNear) cannot be projected.
const wNear = 1e-4

// clipTriNear clips a triangle against the near plane (w > wNear) and returns
// the resulting screen-space triangles — none, one, or two. Vertices are
// interpolated in clip space (colour and texcoords linearly with them) and only
// then projected, which is what stops off-screen wraparound from painting the
// whole frame.
func clipTriNear(s *geState, a, b, c vert) [][3]vert {
	if !a.clip || !b.clip || !c.clip {
		return [][3]vert{{a, b, c}}
	}
	if a.cw > wNear && b.cw > wNear && c.cw > wNear {
		return [][3]vert{{a, b, c}}
	}
	in := []vert{a, b, c}
	var out []vert
	for i := 0; i < len(in); i++ {
		cur, nxt := in[i], in[(i+1)%len(in)]
		curIn, nxtIn := cur.cw > wNear, nxt.cw > wNear
		if curIn {
			out = append(out, cur)
		}
		if curIn != nxtIn {
			t := (wNear - cur.cw) / (nxt.cw - cur.cw)
			out = append(out, lerpVert(s, cur, nxt, t))
		}
	}
	if len(out) < 3 {
		return nil
	}
	tris := make([][3]vert, 0, 2)
	for i := 1; i+1 < len(out); i++ {
		tris = append(tris, [3]vert{out[0], out[i], out[i+1]})
	}
	return tris
}

// lerpVert interpolates two clip-space vertices and projects the result.
func lerpVert(s *geState, p, q vert, t float32) vert {
	li := func(a, b float32) float32 { return a + (b-a)*t }
	lb := func(a, b byte) byte { return byte(float32(a) + (float32(b)-float32(a))*t) }
	var v vert
	v.clip = true
	v.cx, v.cy = li(p.cx, q.cx), li(p.cy, q.cy)
	v.cz, v.cw = li(p.cz, q.cz), li(p.cw, q.cw)
	v.u, v.v = li(p.u, q.u), li(p.v, q.v)
	v.r, v.g, v.b, v.a = lb(p.r, q.r), lb(p.g, q.g), lb(p.b, q.b), lb(p.a, q.a)
	v.x, v.y, v.z = project(s, v.cx, v.cy, v.cz, v.cw)
	v.invW = invW(v.cw)
	v.fog = fogCoef(s, v.cw)
	return v
}

// --- matrix helpers --------------------------------------------------------

func ident(m *[16]float32) {
	*m = [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

// matPush appends a 4x3 world/view matrix element (column-major, 3 rows per column).
func matPush(m *[16]float32, idx *int, arg uint32) {
	if *idx >= 12 {
		return
	}
	col := *idx / 3
	row := *idx % 3
	m[col*4+row] = math.Float32frombits(arg << 8)
	if row == 2 {
		m[col*4+3] = 0
	}
	if *idx == 11 {
		m[15] = 1
	}
	*idx++
}

// matPush16 appends a 4x4 projection matrix element.
func matPush16(m *[16]float32, idx *int, arg uint32) {
	if *idx >= 16 {
		return
	}
	m[*idx] = math.Float32frombits(arg << 8)
	*idx++
}

func mul4(a, b [16]float32) [16]float32 {
	var r [16]float32
	for c := 0; c < 4; c++ {
		for rr := 0; rr < 4; rr++ {
			var s float32
			for k := 0; k < 4; k++ {
				s += a[k*4+rr] * b[c*4+k]
			}
			r[c*4+rr] = s
		}
	}
	return r
}

// f24 expands a GE float argument (top 24 bits of an IEEE float) to float32.
func f24(arg uint32) float32 { return math.Float32frombits(arg << 8) }

func u16(m *Machine, a uint32) uint16  { return uint16(m.Read(a)) | uint16(m.Read(a+1))<<8 }
func f32(m *Machine, a uint32) float32 { return math.Float32frombits(m.read32(a)) }
