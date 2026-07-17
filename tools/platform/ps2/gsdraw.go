package ps2

// gsdraw.go is the Graphics Synthesizer's drawing half: the vertex queue, the primitive
// assembly, and the rasteriser that turns kicks into pixels in GS memory.
//
// The GS draws the way a register file draws. RGBAQ, ST and UV set the CURRENT vertex's
// attributes; writing one of the position registers pushes the position and the current
// attributes into the vertex queue, and the "2" registers (XYZ2, XYZF2) also KICK — when
// the queue holds enough vertices for the primitive PRIM names, the primitive is drawn
// and the queue advances by the primitive's stride (everything for a triangle, one vertex
// for a strip, the middle vertex for a fan). The "3" registers (XYZ3, XYZF3) push without
// kicking, which is how a strip is broken without drawing.
//
// What is rasterised here: flat and gouraud sprites and triangles (and their strips and
// fans) into a PSMCT32 frame buffer, honouring XYOFFSET, SCISSOR and FBMSK — now with
// texture sampling (gstex.go), the alpha and destination-alpha tests, and alpha blending.
// The depth test is still recorded per primitive and not applied (there is no Z buffer
// yet), and anything the sampler cannot serve draws as its vertex colour — visible and
// wrong in an honest way, where a skipped primitive is invisible in a misleading one.

import (
	"fmt"
	"math"
	"os"
)

// writeFile is the huge-primitive dump's file writer, in the current directory.
func writeFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}

// The GS drawing registers beyond those gs.go already names.
const (
	gsST        = 0x02
	gsUV        = 0x03
	gsXYZF2     = 0x04
	gsXYZF3     = 0x0C
	gsXYZ3      = 0x0D
	gsPRMODE    = 0x1B
	gsFRAME2    = 0x4D
	gsXYOFFSET2 = 0x19
	gsSCISSOR2  = 0x41
)

// The PRIM register's primitive types.
const (
	primPoint     = 0
	primLine      = 1
	primLineStrip = 2
	primTri       = 3
	primTriStrip  = 4
	primTriFan    = 5
	primSprite    = 6
)

var primNames = [8]string{"point", "line", "linestrip", "tri", "tristrip", "trifan", "sprite", "prim7"}

// gsVertex is one entry of the vertex queue: a position in window space (12.4 fixed,
// XYOFFSET already subtracted), a depth, and the attributes latched when it was pushed —
// the colour, and both kinds of texture coordinate (UV when PRIM.FST says texel space,
// STQ when it says perspective).
type gsVertex struct {
	x, y    int32 // 12.4 fixed point
	z       uint32
	rgba    uint32
	u, v    int32 // 10.4 fixed texels, from the UV register
	s, t, q float32
	f       uint32 // fog coefficient: 255 = all vertex colour, 0 = all FOGCOL
}

// pushVertex latches the current attributes with a position and, on a kicking write,
// assembles whatever primitive is now complete. f is the vertex's fog coefficient —
// XYZF2/XYZF3 carry it per vertex, XYZ2/XYZ3 vertices take the FOG register's latch.
func (gs *GS) pushVertex(x, y int32, z uint32, f uint32, kick bool) {
	xyoff := gs.reg[gsXYOFFSET1]
	if gs.ctxt() == 1 {
		xyoff = gs.reg[gsXYOFFSET2]
	}
	uv := gs.reg[gsUV]
	st := gs.reg[gsST]
	v := gsVertex{
		x:    x - int32(uint32(xyoff)&0xFFFF),
		y:    y - int32(uint32(xyoff>>32)&0xFFFF),
		z:    z,
		rgba: uint32(gs.reg[gsRGBAQ]),
		u:    int32(uint32(uv) & 0x3FFF),
		v:    int32(uint32(uv>>16) & 0x3FFF),
		s:    math.Float32frombits(uint32(st)),
		t:    math.Float32frombits(uint32(st >> 32)),
		// Q comes from the LATCH, not the RGBAQ register: a perspective stream is
		// RGBAQ once then ST,XYZF2 per vertex, and each PACKED ST carries the Q its
		// vertex divides by. Reading the register here left every vertex with a stale
		// Q of zero — and a zero Q maps every sample to texel (0,0), which is how a
		// whole title screen sampled a black border pixel.
		q: math.Float32frombits(gs.q),
		f: f,
	}
	if gs.vqN < len(gs.vq) {
		gs.vq[gs.vqN] = v
		gs.vqN++
	}
	if kick {
		gs.kick()
	}
}

// prim returns the PRIM word that governs drawing: PRIM itself, or PRMODE's flags over
// PRIM's type when PRMODECONT says PRMODE is in charge.
func (gs *GS) prim() uint64 {
	p := gs.reg[gsPRIM]
	if gs.reg[gsPRMODECONT]&1 == 0 {
		p = p&7 | gs.reg[gsPRMODE]&^7
	}
	return p
}

// ctxt is the drawing context the current primitive uses: 0 or 1 (the GS's contexts 1
// and 2), from PRIM's CTXT bit.
func (gs *GS) ctxt() int { return int(gs.prim() >> 9 & 1) }

// kick assembles the primitive the queue now completes, if it does, and advances the
// queue by the primitive's stride.
func (gs *GS) kick() {
	p := gs.prim()
	typ := int(p & 7)

	switch typ {
	case primPoint:
		if gs.vqN >= 1 {
			gs.drawn(typ)
			gs.point(gs.vq[0])
			gs.vqN = 0
		}
	case primLine:
		if gs.vqN >= 2 {
			gs.drawn(typ)
			gs.count("line (not rasterised)")
			gs.vqN = 0
		}
	case primLineStrip:
		if gs.vqN >= 2 {
			gs.drawn(typ)
			gs.count("linestrip (not rasterised)")
			gs.vq[0] = gs.vq[1]
			gs.vqN = 1
		}
	case primTri:
		if gs.vqN >= 3 {
			gs.drawn(typ)
			gs.triangle(gs.vq[0], gs.vq[1], gs.vq[2], p)
			gs.vqN = 0
		}
	case primTriStrip:
		if gs.vqN >= 3 {
			gs.drawn(typ)
			gs.triangle(gs.vq[0], gs.vq[1], gs.vq[2], p)
			gs.vq[0], gs.vq[1] = gs.vq[1], gs.vq[2]
			gs.vqN = 2
		}
	case primTriFan:
		if gs.vqN >= 3 {
			gs.drawn(typ)
			gs.triangle(gs.vq[0], gs.vq[1], gs.vq[2], p)
			gs.vq[1] = gs.vq[2]
			gs.vqN = 2
		}
	case primSprite:
		if gs.vqN >= 2 {
			gs.drawn(typ)
			gs.sprite(gs.vq[0], gs.vq[1], p)
			gs.vqN = 0
		}
	default:
		gs.vqN = 0
	}
}

// drawn tallies one completed primitive.
func (gs *GS) drawn(typ int) {
	gs.primCount[typ]++
	gs.prims++
	// The frame debugger's per-primitive hook: it records this primitive as one command
	// and hands back the index the pixels it draws should carry, so a pixel can name the
	// primitive that wrote it.
	gs.curCmd = -1
	if gs.m != nil && gs.m.OnGSPrim != nil {
		gs.curCmd = gs.m.OnGSPrim(typ, gs.src)
	}
	// The vertex-dump instrument: the first N completed primitives, with the exact
	// numbers the rasteriser is about to consume — position in pixels (12.4 fixed,
	// XYOFFSET already subtracted), Z, the RGBA the vertex latched, and its ST/Q.
	// This is the discriminator between "the transform is wrong" (positions huge or
	// offscreen), "the lighting is wrong" (RGBA black), and "the unpack is wrong"
	// (alpha zero, Q zero): each hypothesis is one column.
	dump := gs.m != nil && gs.m.GSVertDump > 0
	if !dump && gs.m != nil && gs.m.GSBigDump > 0 && gs.vqN >= 2 {
		// The huge-primitive hunter: only primitives whose bounding box exceeds 1024px
		// on an axis, with the producer named.
		minX, maxX, minY, maxY := gs.vq[0].x, gs.vq[0].x, gs.vq[0].y, gs.vq[0].y
		for i := 1; i < gs.vqN && i < len(gs.vq); i++ {
			v := gs.vq[i]
			if v.x < minX {
				minX = v.x
			}
			if v.x > maxX {
				maxX = v.x
			}
			if v.y < minY {
				minY = v.y
			}
			if v.y > maxY {
				maxY = v.y
			}
		}
		if maxX-minX > 1024*16 || maxY-minY > 1024*16 {
			gs.m.GSBigDump--
			fmt.Printf("  BIG (%dx%d px) from %s\n", (maxX-minX)/16, (maxY-minY)/16, gs.src)
			dump = true
			gs.m.GSVertDump++ // consumed by the shared printer below
			if !gs.srcDumped && gs.srcData != nil {
				// The whole packet, once: whether the period-2 garbage is in the
				// packet (the program built it) or not (our parse mangled it).
				gs.srcDumped = true
				n := len(gs.srcData)
				if n > 40*16 {
					n = 40 * 16
				}
				for o := 0; o+16 <= n; o += 16 {
					fmt.Printf("    pkt qw+%-3d %08X %08X %08X %08X\n", o/16,
						le32gs(gs.srcData[o:]), le32gs(gs.srcData[o+4:]),
						le32gs(gs.srcData[o+8:]), le32gs(gs.srcData[o+12:]))
				}
				n = len(gs.srcIn)
				if n > 24*16 {
					n = 24 * 16
				}
				for o := 0; o+16 <= n; o += 16 {
					fmt.Printf("    in  qw+%-3d %08X %08X %08X %08X\n", o/16,
						le32gs(gs.srcIn[o:]), le32gs(gs.srcIn[o+4:]),
						le32gs(gs.srcIn[o+8:]), le32gs(gs.srcIn[o+12:]))
				}
				// The live micro and data memories, to files: the program that built
				// this packet is only in memory NOW (uploads overwrite it constantly).
				if gs.srcMicro != nil {
					_ = writeFile("bigkick-micro.bin", gs.srcMicro)
					_ = writeFile("bigkick-data.bin", gs.srcVUData)
					fmt.Printf("    wrote bigkick-micro.bin / bigkick-data.bin\n")
				}
			}
		}
	}
	if dump {
		gs.m.GSVertDump--
		p := gs.prim()
		xyoff := gs.reg[gsXYOFFSET1]
		scis := gs.reg[gsSCISSOR1]
		if gs.ctxt() == 1 {
			xyoff = gs.reg[gsXYOFFSET2]
			scis = gs.reg[gsSCISSOR2]
		}
		tex0 := gs.reg[gsTEX0_1]
		frame := gs.reg[gsFRAME1]
		alpha := gs.reg[gsALPHA1]
		if gs.ctxt() == 1 {
			tex0 = gs.reg[gsTEX0_2]
			frame = gs.reg[gsFRAME2]
			alpha = gs.reg[gsALPHA2]
		}
		texS := ""
		if p&(1<<4) != 0 {
			texS = sprintf(" tex0 %016X", tex0)
		}
		if p&(1<<6) != 0 {
			texS += sprintf(" alpha %010X", alpha)
		}
		fmt.Printf("  prim %-9s PRIM=0x%03X ctx%d%s%s%s fb 0x%05X xyoff (%d,%d) scissor x %d..%d y %d..%d%s from %s:\n",
			primNames[typ&7], p, gs.ctxt()+1,
			map[bool]string{true: " TME", false: ""}[p&(1<<4) != 0],
			map[bool]string{true: " ABE", false: ""}[p&(1<<6) != 0],
			map[bool]string{true: " FGE", false: ""}[p&(1<<5) != 0],
			uint32(frame)&0x1FF*2048,
			uint32(xyoff)&0xFFFF>>4, uint32(xyoff>>32)&0xFFFF>>4,
			uint32(scis)&0x7FF, uint32(scis>>16)&0x7FF,
			uint32(scis>>32)&0x7FF, uint32(scis>>48)&0x7FF,
			texS, gs.src)
		for i := 0; i < gs.vqN && i < len(gs.vq); i++ {
			v := gs.vq[i]
			fmt.Printf("    v%d xy (%8.2f,%8.2f) z %08X rgba %08X stq (%g, %g, %g) uv (%.1f,%.1f)\n",
				i, float64(v.x)/16, float64(v.y)/16, v.z, v.rgba, v.s, v.t, v.q,
				float64(v.u)/16, float64(v.v)/16)
		}
	}
}

// count tallies a drawing feature the rasteriser saw, so the census names what is being
// asked for before it is built.
func (gs *GS) count(what string) {
	if gs.drawCensus == nil {
		gs.drawCensus = map[string]int{}
	}
	gs.drawCensus[what]++
}

// target describes the frame buffer the current context draws into, and the per-pixel
// state the current primitive draws with: the tests, the blend, and the write masks.
type gsTarget struct {
	fbp, fbw uint32 // base (64-word blocks) and width (units of 64px)
	psm      uint32
	fbmsk    uint32
	sx0, sy0 int32 // the scissor, inclusive, in pixels
	sx1, sy1 int32

	abe      bool   // PRIM: blend this primitive
	alpha    uint64 // the context's ALPHA register
	pabe     bool   // blend only pixels whose source alpha has its MSB set
	colclamp bool
	fba      uint32 // force the written alpha's MSB
	test     uint64 // the context's TEST register
	fge      bool   // PRIM: fog this primitive
	fogcol   uint32 // FOGCOL's RGB, blended in by 255-f before the alpha blend

	// The Z buffer, from the context's ZBUF register and TEST's ZTE/ZTST.
	zbp  uint32 // base, in 64-word blocks (the register speaks 2048-word pages)
	zpsm uint32
	zmsk bool // don't write Z
	ztst uint32

	primType int // for the per-type pixel tallies
}

// target reads the current context's FRAME, SCISSOR and per-pixel-pipeline registers.
func (gs *GS) target(p uint64) gsTarget {
	frame := gs.reg[gsFRAME1]
	scis := gs.reg[gsSCISSOR1]
	alpha := gs.reg[gsALPHA1]
	test := gs.reg[gsTEST1]
	fba := gs.reg[gsFBA1]
	zbuf := gs.reg[gsZBUF1]
	if p>>9&1 == 1 {
		frame = gs.reg[gsFRAME2]
		scis = gs.reg[gsSCISSOR2]
		alpha = gs.reg[gsALPHA2]
		test = gs.reg[gsTEST2]
		fba = gs.reg[gsFBA2]
		zbuf = gs.reg[gsZBUF2]
	}
	ztst := uint32(1) // ZTE off is "always pass" too
	if test>>16&1 != 0 {
		ztst = uint32(test>>17) & 3
	}
	return gsTarget{
		zbp:  uint32(zbuf) & 0x1FF * 32,
		zpsm: uint32(zbuf>>24)&0xF | 0x30, // the register stores the low nibble of a Z PSM
		zmsk: zbuf>>32&1 != 0,
		ztst: ztst,

		fbp:   uint32(frame) & 0x1FF * 32, // FBP is in 2048-word pages; addrPSMCT32 wants 64-word blocks
		fbw:   uint32(frame>>16) & 0x3F,
		psm:   uint32(frame>>24) & 0x3F,
		fbmsk: uint32(frame >> 32),
		sx0:   int32(uint32(scis) & 0x7FF),
		sx1:   int32(uint32(scis>>16) & 0x7FF),
		sy0:   int32(uint32(scis>>32) & 0x7FF),
		sy1:   int32(uint32(scis>>48) & 0x7FF),

		abe:      p&(1<<6) != 0,
		alpha:    alpha,
		pabe:     gs.reg[gsPABE]&1 != 0,
		colclamp: gs.reg[gsCOLCLAMP]&1 != 0,
		fba:      uint32(fba) & 1,
		test:     test,
		fge:      p&(1<<5) != 0,
		fogcol:   uint32(gs.reg[gsFOGCOL]) & 0xFFFFFF,

		primType: int(p & 7),
	}
}

// plot writes one pixel: the alpha, destination-alpha and depth tests, the blend, then
// the frame's format, mask and scissor. rgba is the source colour after texturing; z is
// the pixel's depth.
func (gs *GS) plot(t *gsTarget, x, y int32, z uint32, rgba uint32) {
	if x < t.sx0 || x > t.sx1 || y < t.sy0 || y > t.sy1 || x < 0 || y < 0 {
		gs.rejScissor++
		return
	}
	if t.psm != psmCT32 && t.psm != psmCT24 {
		// A wrong-format write is worse than none — but a DROPPED write must be loud in
		// the census: a 16-bit render target that stays garbage explains every sampler
		// that reads it back.
		gs.count(sprintf("DROPPED pixel writes: frame psm 0x%02X at fb 0x%05X", t.psm, t.fbp*64))
		return
	}

	fbmsk := t.fbmsk
	writeZ := !t.zmsk
	srcA := rgba >> 24

	// The alpha test compares the source alpha against AREF; a failing pixel is dropped
	// or writes partially, by AFAIL.
	if t.test&1 != 0 {
		atst := t.test >> 1 & 7
		aref := uint32(t.test>>4) & 0xFF
		pass := false
		switch atst {
		case 0: // NEVER
		case 1:
			pass = true
		case 2:
			pass = srcA < aref
		case 3:
			pass = srcA <= aref
		case 4:
			pass = srcA == aref
		case 5:
			pass = srcA >= aref
		case 6:
			pass = srcA > aref
		case 7:
			pass = srcA != aref
		}
		if !pass {
			switch t.test >> 12 & 3 {
			case 0: // KEEP: nothing is written
				gs.rejAlpha++
				return
			case 1: // FB_ONLY: the frame is written, the Z buffer is not
				writeZ = false
			case 2: // ZB_ONLY: only the Z buffer is written
				fbmsk = 0xFFFFFFFF
			case 3: // RGB_ONLY: the colour is written, the alpha and Z are not
				fbmsk |= 0xFF000000
				writeZ = false
			}
		}
	}

	// The depth test, against the context's Z buffer.
	zaddr := uint32(0xFFFFFFFF)
	if t.ztst != 1 || writeZ {
		var zold uint32
		switch t.zpsm {
		case psmZ32, psmZ24:
			zaddr = addrPSMZ32(t.zbp, t.fbw, uint32(x), uint32(y))
			if zaddr+4 > uint32(len(gs.vram)) {
				return
			}
			zold = le32gs(gs.vram[zaddr:])
			if t.zpsm == psmZ24 {
				zold &= 0xFFFFFF
				z &= 0xFFFFFF
			}
		case psmZ16, psmZ16S:
			zaddr = addrPSMZ16(t.zbp, t.fbw, uint32(x), uint32(y), t.zpsm == psmZ16S)
			if zaddr+2 > uint32(len(gs.vram)) {
				return
			}
			zold = uint32(gs.vram[zaddr]) | uint32(gs.vram[zaddr+1])<<8
			if z > 0xFFFF {
				z = 0xFFFF
			}
		}
		switch t.ztst {
		case 0: // NEVER
			gs.rejZ++
			return
		case 2: // GEQUAL
			if z < zold {
				gs.rejZ++
				return
			}
		case 3: // GREATER
			if z <= zold {
				gs.rejZ++
				return
			}
		}
	}

	addr := addrPSMCT32(t.fbp, t.fbw, uint32(x), uint32(y))
	if addr+4 > uint32(len(gs.vram)) {
		return
	}
	old := le32gs(gs.vram[addr:])

	// The destination-alpha test keys on the frame's own alpha MSB.
	if t.test>>14&1 != 0 {
		datm := uint32(t.test>>15) & 1
		if old>>31 != datm {
			gs.rejDate++
			return
		}
	}

	gs.plotted++
	preBlend := rgba
	blended := false
	if t.abe && (!t.pabe || srcA&0x80 != 0) {
		dstA := int64(old >> 24)
		if t.psm == psmCT24 {
			dstA = 0x80 // a frame with no alpha reads as 1.0
		}
		rgba = blendPixel(t.alpha, rgba, old, dstA, t.colclamp)
		blended = true
	}
	if t.fba != 0 {
		rgba |= 0x80000000
	}

	if writeZ && zaddr != 0xFFFFFFFF {
		switch t.zpsm {
		case psmZ32, psmZ24:
			gs.vram[zaddr+0] = byte(z)
			gs.vram[zaddr+1] = byte(z >> 8)
			gs.vram[zaddr+2] = byte(z >> 16)
			if t.zpsm == psmZ32 {
				gs.vram[zaddr+3] = byte(z >> 24)
			}
		case psmZ16, psmZ16S:
			gs.vram[zaddr+0] = byte(z)
			gs.vram[zaddr+1] = byte(z >> 8)
		}
	}
	if fbmsk == 0xFFFFFFFF {
		return
	}

	px := rgba&^fbmsk | old&fbmsk
	if px&0xFFFFFF != 0 {
		gs.plotNonBlack[t.primType]++
	}
	if gs.m != nil && gs.m.GSPixelN > 0 && x == gs.m.GSPixelX && y == gs.m.GSPixelY {
		gs.m.GSPixelN--
		blendS := ""
		if blended {
			blendS = sprintf(" blend(%08X %X->%08X)", preBlend, uint32(t.alpha), rgba)
		}
		fmt.Printf("  pixel (%d,%d) fb 0x%05X <- %08X (src %08X over %08X, %s%s%s) from %s\n",
			x, y, t.fbp*64, px, rgba, old,
			primNames[t.primType],
			map[bool]string{true: " ABE", false: ""}[t.abe],
			blendS,
			gs.src)
	}
	gs.vram[addr+0] = byte(px)
	gs.vram[addr+1] = byte(px >> 8)
	gs.vram[addr+2] = byte(px >> 16)
	gs.vram[addr+3] = byte(px >> 24)
	// The frame debugger's per-pixel hook: this colour write lands at VRAM (x, y) in the
	// buffer at t.fbp, and is attributed to the primitive drawn() named. The debugger
	// maps it onto whichever buffer it is showing as the frame.
	if gs.m != nil && gs.m.OnGSPixel != nil {
		gs.m.OnGSPixel(gs.curCmd, t.fbp*64, x, y)
	}
}

// fogPixel folds FOGCOL into a source colour by the vertex fog coefficient: f=255 keeps
// the colour, f=0 replaces its RGB with FOGCOL. Fog happens after texturing and before
// the alpha blend, and touches only RGB — the source alpha rides through.
func fogPixel(rgba, f, fogcol uint32) uint32 {
	g := 255 - f
	r := (f*(rgba&0xFF) + g*(fogcol&0xFF)) >> 8
	gr := (f*(rgba>>8&0xFF) + g*(fogcol>>8&0xFF)) >> 8
	b := (f*(rgba>>16&0xFF) + g*(fogcol>>16&0xFF)) >> 8
	return r | gr<<8 | b<<16 | rgba&0xFF000000
}

// point draws the one-vertex primitive.
func (gs *GS) point(v gsVertex) {
	p := gs.prim()
	t := gs.target(p)
	rgba := v.rgba
	if t.fge {
		rgba = fogPixel(rgba, v.f&0xFF, t.fogcol)
	}
	gs.plot(&t, v.x>>4, v.y>>4, v.z, rgba)
}

// texAxis interpolates one texture-coordinate axis of a sprite: the texel coordinate
// (10.4 fixed) at pixel-centre pc, on the line from (p0, uv0) to (p1, uv1) in 12.4
// window space.
func texAxis(pc, p0, p1, uv0, uv1 int32) int32 {
	if p1 == p0 {
		return uv0
	}
	return uv0 + int32(int64(pc-p0)*int64(uv1-uv0)/int64(p1-p0))
}

// sprite fills the axis-aligned rectangle between its two vertices. The GS takes the
// SECOND vertex's colour for a flat sprite, and the rectangle excludes its right and
// bottom edges (the same half-open rule the triangles use). Texture coordinates map
// linearly from the first vertex's to the second's — a sprite is never perspective, so
// STQ divides once per vertex and interpolates like UV.
func (gs *GS) sprite(a, b gsVertex, p uint64) {
	gs.noteFeatures(p)
	t := gs.target(p)
	smp := gs.sampler(p)
	fst := p&(1<<8) != 0

	au, av, bu, bv := a.u, a.v, b.u, b.v
	if smp != nil && !fst {
		// STQ: perspective-divide at the two corners, into 10.4 texels.
		if a.q != 0 {
			au = int32(a.s / a.q * float32(int32(1)<<smp.tex.tw) * 16)
			av = int32(a.t / a.q * float32(int32(1)<<smp.tex.th) * 16)
		}
		if b.q != 0 {
			bu = int32(b.s / b.q * float32(int32(1)<<smp.tex.tw) * 16)
			bv = int32(b.t / b.q * float32(int32(1)<<smp.tex.th) * 16)
		}
	}

	x0, x1 := a.x>>4, b.x>>4
	y0, y1 := a.y>>4, b.y>>4
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y < y1; y++ {
		var v int32
		if smp != nil {
			v = texAxis(y<<4+8, a.y, b.y, av, bv)
		}
		for x := x0; x < x1; x++ {
			rgba := b.rgba
			if smp != nil {
				u := texAxis(x<<4+8, a.x, b.x, au, bu)
				rgba = smp.combine(smp.at(u>>4, v>>4), rgba)
			}
			if t.fge {
				rgba = fogPixel(rgba, b.f&0xFF, t.fogcol) // flat in fog too: the second vertex's
			}
			gs.plot(&t, x, y, b.z, rgba) // a sprite is flat in Z: the second vertex's
		}
	}
}

// triangle rasterises with the half-open edge rule, flat or gouraud by PRIM's IIP bit.
// Positions are 12.4; the walk is per-pixel at pixel centres. Texture coordinates
// interpolate barycentrically — UV directly, STQ as three linear terms divided per pixel,
// which is what makes the perspective correct.
func (gs *GS) triangle(v0, v1, v2 gsVertex, p uint64) {
	gs.noteFeatures(p)
	t := gs.target(p)
	smp := gs.sampler(p)
	fst := p&(1<<8) != 0
	gouraud := p&(1<<3) != 0

	// Bounding box in whole pixels, clipped by the scissor.
	minX := min3(v0.x, v1.x, v2.x) >> 4
	maxX := (max3(v0.x, v1.x, v2.x) + 15) >> 4
	minY := min3(v0.y, v1.y, v2.y) >> 4
	maxY := (max3(v0.y, v1.y, v2.y) + 15) >> 4
	if minX < t.sx0 {
		minX = t.sx0
	}
	if maxX > t.sx1+1 {
		maxX = t.sx1 + 1
	}
	if minY < t.sy0 {
		minY = t.sy0
	}
	if maxY > t.sy1+1 {
		maxY = t.sy1 + 1
	}
	if minX >= maxX || minY >= maxY {
		return
	}

	// Edge functions in 12.4 space; the area is in 8.8.
	area := int64(v1.x-v0.x)*int64(v2.y-v0.y) - int64(v1.y-v0.y)*int64(v2.x-v0.x)
	if area == 0 {
		return
	}
	if area < 0 {
		v1, v2 = v2, v1
		area = -area
	}

	c0r, c0g, c0b, c0a := unpackRGBA(v0.rgba)
	c1r, c1g, c1b, c1a := unpackRGBA(v1.rgba)
	c2r, c2g, c2b, c2a := unpackRGBA(v2.rgba)

	for y := minY; y < maxY; y++ {
		py := y<<4 + 8 // the pixel centre in 12.4
		for x := minX; x < maxX; x++ {
			px := x<<4 + 8
			w0 := int64(v1.x-px)*int64(v2.y-py) - int64(v1.y-py)*int64(v2.x-px)
			w1 := int64(v2.x-px)*int64(v0.y-py) - int64(v2.y-py)*int64(v0.x-px)
			w2 := int64(v0.x-px)*int64(v1.y-py) - int64(v0.y-py)*int64(v1.x-px)
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			rgba := v2.rgba // flat shading takes the LAST vertex's colour
			if gouraud {
				r := (w0*c0r + w1*c1r + w2*c2r) / area
				g := (w0*c0g + w1*c1g + w2*c2g) / area
				b := (w0*c0b + w1*c1b + w2*c2b) / area
				a := (w0*c0a + w1*c1a + w2*c2a) / area
				rgba = uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
			}
			if smp != nil {
				var tu, tv int32
				if fst {
					tu = int32((w0*int64(v0.u) + w1*int64(v1.u) + w2*int64(v2.u)) / area >> 4)
					tv = int32((w0*int64(v0.v) + w1*int64(v1.v) + w2*int64(v2.v)) / area >> 4)
				} else {
					fw0, fw1, fw2 := float32(w0), float32(w1), float32(w2)
					fa := float32(area)
					s := (fw0*v0.s + fw1*v1.s + fw2*v2.s) / fa
					tt := (fw0*v0.t + fw1*v1.t + fw2*v2.t) / fa
					q := (fw0*v0.q + fw1*v1.q + fw2*v2.q) / fa
					if q != 0 {
						tu = int32(s / q * float32(int32(1)<<smp.tex.tw))
						tv = int32(tt / q * float32(int32(1)<<smp.tex.th))
					}
				}
				rgba = smp.combine(smp.at(tu, tv), rgba)
			}
			if t.fge {
				f := uint32((w0*int64(v0.f&0xFF) + w1*int64(v1.f&0xFF) + w2*int64(v2.f&0xFF)) / area)
				rgba = fogPixel(rgba, f, t.fogcol)
			}
			z := uint32((w0*int64(v0.z) + w1*int64(v1.z) + w2*int64(v2.z)) / area)
			gs.plot(&t, x, y, z, rgba)
		}
	}
}

// noteFeatures tallies what a drawn primitive asked for — the applied features by name,
// and the ones the rasteriser records but does not apply yet, marked so.
func (gs *GS) noteFeatures(p uint64) {
	if p&(1<<4) != 0 {
		tex0 := gs.reg[gsTEX0_1]
		if p>>9&1 == 1 {
			tex0 = gs.reg[gsTEX0_2]
		}
		t := decodeTEX0(tex0)
		gs.count(sprintf("textured PSM 0x%02X", t.psm))
		// Each distinct texture, so the bases can be laid against the uploads': a
		// texture whose page nothing ever filled samples black, and this line is how
		// that is seen rather than suspected.
		gs.count(sprintf("  tex base 0x%05X %dx%d psm 0x%02X tfx %d cbp 0x%05X csa %d cld %d tex0 %016X",
			t.tbp*64, 1<<t.tw, 1<<t.th, t.psm, t.tfx, t.cbp*64, t.csa, t.cld, tex0))
	}
	if p&(1<<6) != 0 {
		gs.count("alpha-blended")
	}
	if p&(1<<5) != 0 {
		gs.count("fogged")
	}
	if p&(1<<9) != 0 {
		gs.count("context 2")
	}
	test := gs.reg[gsTEST1]
	frame := gs.reg[gsFRAME1]
	zbuf := gs.reg[gsZBUF1]
	if p>>9&1 == 1 {
		test = gs.reg[gsTEST2]
		frame = gs.reg[gsFRAME2]
		zbuf = gs.reg[gsZBUF2]
	}
	if test>>16&1 != 0 && test>>17&3 >= 2 {
		gs.count("depth-tested")
	}
	// Each distinct render target, laid against the texture bases above: a page that
	// is drawn into and then sampled is a dynamic texture, and this pair of lines is
	// how that is seen.
	gs.count(sprintf("  draw target fb 0x%05X z 0x%05X", uint32(frame)&0x1FF*2048, uint32(zbuf)&0x1FF*2048))
}

func unpackRGBA(c uint32) (r, g, b, a int64) {
	return int64(c & 0xFF), int64(c >> 8 & 0xFF), int64(c >> 16 & 0xFF), int64(c >> 24 & 0xFF)
}

func min3(a, b, c int32) int32 {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

func max3(a, b, c int32) int32 {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}
