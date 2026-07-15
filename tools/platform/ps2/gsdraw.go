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
// What is rasterised here is the part the game has shown it needs first: flat and gouraud
// sprites and triangles (and their strips and fans) into a PSMCT32 frame buffer, honouring
// XYOFFSET, SCISSOR and FBMSK. Texturing, blending and the depth test are recorded per
// primitive in the census and drawn WITHOUT their effect for now — a textured HUD quad
// appears as its vertex colour, which is visible and wrong in an honest way, where a
// skipped primitive is invisible in a misleading one.

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
// XYOFFSET already subtracted), a depth, and the attributes latched when it was pushed.
type gsVertex struct {
	x, y int32 // 12.4 fixed point
	z    uint32
	rgba uint32
}

// pushVertex latches the current attributes with a position and, on a kicking write,
// assembles whatever primitive is now complete.
func (gs *GS) pushVertex(x, y int32, z uint32, kick bool) {
	xyoff := gs.reg[gsXYOFFSET1]
	if gs.ctxt() == 1 {
		xyoff = gs.reg[gsXYOFFSET2]
	}
	v := gsVertex{
		x:    x - int32(uint32(xyoff)&0xFFFF),
		y:    y - int32(uint32(xyoff>>32)&0xFFFF),
		z:    z,
		rgba: uint32(gs.reg[gsRGBAQ]),
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

// target describes the frame buffer the current context draws into.
type gsTarget struct {
	fbp, fbw uint32 // base (64-word blocks) and width (units of 64px)
	psm      uint32
	fbmsk    uint32
	sx0, sy0 int32 // the scissor, inclusive, in pixels
	sx1, sy1 int32
}

// target reads the current context's FRAME and SCISSOR.
func (gs *GS) target() gsTarget {
	frame := gs.reg[gsFRAME1]
	scis := gs.reg[gsSCISSOR1]
	if gs.ctxt() == 1 {
		frame = gs.reg[gsFRAME2]
		scis = gs.reg[gsSCISSOR2]
	}
	return gsTarget{
		fbp:   uint32(frame) & 0x1FF * 32, // FBP is in 2048-word pages; addrPSMCT32 wants 64-word blocks
		fbw:   uint32(frame>>16) & 0x3F,
		psm:   uint32(frame>>24) & 0x3F,
		fbmsk: uint32(frame >> 32),
		sx0:   int32(uint32(scis) & 0x7FF),
		sx1:   int32(uint32(scis>>16) & 0x7FF),
		sy0:   int32(uint32(scis>>32) & 0x7FF),
		sy1:   int32(uint32(scis>>48) & 0x7FF),
	}
}

// plot writes one pixel through the frame's format, mask and scissor.
func (gs *GS) plot(t *gsTarget, x, y int32, rgba uint32) {
	if x < t.sx0 || x > t.sx1 || y < t.sy0 || y > t.sy1 || x < 0 || y < 0 {
		return
	}
	if t.psm != psmCT32 && t.psm != psmCT24 {
		return // counted at the primitive level; a wrong-format write is worse than none
	}
	addr := addrPSMCT32(t.fbp, t.fbw, uint32(x), uint32(y))
	if addr+4 > uint32(len(gs.vram)) {
		return
	}
	old := le32gs(gs.vram[addr:])
	px := rgba&^t.fbmsk | old&t.fbmsk
	gs.vram[addr+0] = byte(px)
	gs.vram[addr+1] = byte(px >> 8)
	gs.vram[addr+2] = byte(px >> 16)
	gs.vram[addr+3] = byte(px >> 24)
}

// point draws the one-vertex primitive.
func (gs *GS) point(v gsVertex) {
	t := gs.target()
	gs.plot(&t, v.x>>4, v.y>>4, v.rgba)
}

// sprite fills the axis-aligned rectangle between its two vertices. The GS takes the
// SECOND vertex's colour for a flat sprite, and the rectangle excludes its right and
// bottom edges (the same half-open rule the triangles use).
func (gs *GS) sprite(a, b gsVertex, p uint64) {
	gs.noteFeatures(p)
	t := gs.target()

	x0, x1 := a.x>>4, b.x>>4
	y0, y1 := a.y>>4, b.y>>4
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			gs.plot(&t, x, y, b.rgba)
		}
	}
}

// triangle rasterises with the half-open edge rule, flat or gouraud by PRIM's IIP bit.
// Positions are 12.4; the walk is per-pixel at pixel centres.
func (gs *GS) triangle(v0, v1, v2 gsVertex, p uint64) {
	gs.noteFeatures(p)
	t := gs.target()
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
			gs.plot(&t, x, y, rgba)
		}
	}
}

// noteFeatures records the PRIM features a drawn primitive asked for that the rasteriser
// does not apply yet — the work list for the next pass.
func (gs *GS) noteFeatures(p uint64) {
	if p&(1<<4) != 0 {
		gs.count("textured (drawn as vertex colour)")
	}
	if p&(1<<6) != 0 {
		gs.count("alpha-blended (drawn opaque)")
	}
	if p&(1<<9) != 0 {
		gs.count("context 2")
	}
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
