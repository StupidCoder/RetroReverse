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

import "math"

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
}

// pushVertex latches the current attributes with a position and, on a kicking write,
// assembles whatever primitive is now complete.
func (gs *GS) pushVertex(x, y int32, z uint32, kick bool) {
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
		q:    math.Float32frombits(uint32(gs.reg[gsRGBAQ] >> 32)),
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

	// The Z buffer, from the context's ZBUF register and TEST's ZTE/ZTST.
	zbp  uint32 // base, in 64-word blocks (the register speaks 2048-word pages)
	zpsm uint32
	zmsk bool // don't write Z
	ztst uint32
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
	}
}

// plot writes one pixel: the alpha, destination-alpha and depth tests, the blend, then
// the frame's format, mask and scissor. rgba is the source colour after texturing; z is
// the pixel's depth.
func (gs *GS) plot(t *gsTarget, x, y int32, z uint32, rgba uint32) {
	if x < t.sx0 || x > t.sx1 || y < t.sy0 || y > t.sy1 || x < 0 || y < 0 {
		return
	}
	if t.psm != psmCT32 && t.psm != psmCT24 {
		return // counted at the primitive level; a wrong-format write is worse than none
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
			return
		case 2: // GEQUAL
			if z < zold {
				return
			}
		case 3: // GREATER
			if z <= zold {
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
			return
		}
	}

	if t.abe && (!t.pabe || srcA&0x80 != 0) {
		dstA := int64(old >> 24)
		if t.psm == psmCT24 {
			dstA = 0x80 // a frame with no alpha reads as 1.0
		}
		rgba = blendPixel(t.alpha, rgba, old, dstA, t.colclamp)
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
	gs.vram[addr+0] = byte(px)
	gs.vram[addr+1] = byte(px >> 8)
	gs.vram[addr+2] = byte(px >> 16)
	gs.vram[addr+3] = byte(px >> 24)
}

// point draws the one-vertex primitive.
func (gs *GS) point(v gsVertex) {
	p := gs.prim()
	t := gs.target(p)
	gs.plot(&t, v.x>>4, v.y>>4, v.z, v.rgba)
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
		gs.count(sprintf("  tex base 0x%05X %dx%d psm 0x%02X", t.tbp*64, 1<<t.tw, 1<<t.th, t.psm))
	}
	if p&(1<<6) != 0 {
		gs.count("alpha-blended")
	}
	if p&(1<<5) != 0 {
		gs.count("fogged (fog not applied)")
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
