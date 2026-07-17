package gc

// gpu_fetch.go reads the vertices a draw primitive carries, transforms them, and hands the
// triangles they form to the rasteriser. It is the counterpart to gpu_vtx.go: that file sizes
// a vertex so the parser can find the next command, this one reads the same bytes as values.
//
// What is read is the position, the normal (which the lighting channel needs), the first
// colour, and the first texture coordinate. The second colour and the other seven texture
// coordinates are stepped over — their sizes still count toward the stride, so the reader lands
// on the right byte, but their values wait for a stage that needs them, and a TEV stage that
// samples one is logged as a ticket rather than fed a coordinate it did not ask for. Each of
// those attributes may be inline or an index into an array elsewhere in memory (see attrData);
// an index that points past the end of RAM leaves the draw undrawn and logged, rather than read
// from the wrong place, so the gap is visible instead of a scramble of triangles.
//
// A vertex leaves here in clip space, not screen space: the triangles are cut against the near
// plane (gpu_clip.go) before the perspective divide, because the divide is only meaningful in
// front of the eye.

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"strings"
)

// drawTrace prints one line per draw primitive — vertices, texture base, and the pixel
// accounting that says where a draw's pixels went (written / z-rejected / alpha-rejected).
var drawTrace = os.Getenv("RR_GC_DRAWTRACE") != ""

// texDump0, when set to a hex physical base address, decodes texture map 0 through the TX
// unit the first time a traced draw binds that base, and writes it to texdump0.png — the
// direct look at what the decoder produces for one suspect texture.
var texDump0 = os.Getenv("RR_GC_TEXDUMP0")

// dumpTex0Once writes the decoded texture map 0 once when its base matches texDump0.
func (g *gpu) dumpTex0Once(m *Machine, base uint32) {
	if texDump0 == "" || fmt.Sprintf("%X", base) != strings.ToUpper(strings.TrimPrefix(texDump0, "0x")) {
		return
	}
	texDump0 = ""
	img, err := m.DumpTexture(0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "TEXDUMP0:", err)
		return
	}
	f, err := os.Create("texdump0.png")
	if err != nil {
		fmt.Fprintln(os.Stderr, "TEXDUMP0:", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Fprintln(os.Stderr, "TEXDUMP0:", err)
		return
	}
	fmt.Fprintf(os.Stderr, "TEXDUMP0: wrote texdump0.png (base 0x%06X)\n", base)
}

// attrLayout is where each attribute a draw needs sits within one vertex, and how to read it.
type attrLayout struct {
	stride int

	hasMatIdx bool
	matIdxOff int

	posOff   int
	posDesc  uint32
	posFmt   uint32
	posComps int
	posFrac  uint32

	nrmOff   int
	nrmDesc  uint32
	nrmFmt   uint32
	nrmComps int
	// nrmIdxComps is how many components the ARRAY element an indexed normal names holds, which
	// is not always nrmComps: a three-index NBT normal points each index at its own
	// three-component vector, so the normal proper is three components, not nine.
	nrmIdxComps int

	col0Off  int
	col0Desc uint32
	col0Comp uint32

	texMtxIdxOff [maxTexCoord]int // per-vertex texture matrix index bytes, -1 when absent

	// All eight texture coordinates, not just the first. A coordinate the vertex carries is a
	// SOURCE row the transform unit may generate any other coordinate from (see gpu_texgen.go's
	// srcRow), so capturing only coordinate 0 does not merely lose coordinates 1..7 — it loses
	// every generated coordinate that reads one of them.
	tex [maxTexCoord]texAttr
}

// texAttr is where one texture coordinate sits in a vertex and how to read it.
type texAttr struct {
	off  int // -1 when the vertex does not carry this coordinate
	desc uint32
	fmt  uint32
	elem uint32
	frac uint32
}

// layout computes the attribute layout for a vertex-attribute-table index, walking the
// descriptor and table in the same order gpu_vtx.go sizes them so the offsets agree.
func (g *gpu) layout(vat int) attrLayout {
	lo := g.CPReg[0x50]
	hi := g.CPReg[0x60]
	g0 := g.CPReg[0x70+uint32(vat)]
	g1 := g.CPReg[0x80+uint32(vat)]
	g2 := g.CPReg[0x90+uint32(vat)]

	a := attrLayout{matIdxOff: -1, posOff: -1, nrmOff: -1, col0Off: -1}
	for i := range a.texMtxIdxOff {
		a.texMtxIdxOff[i] = -1
		a.tex[i].off = -1
	}
	off := 0

	// The position matrix index, then eight texture matrix indices, are each one inline byte.
	// A texture matrix index here overrides the one the matrix-index register selects, for
	// that coordinate on this vertex alone (see genTexCoord's caller).
	if lo&1 != 0 {
		a.hasMatIdx = true
		a.matIdxOff = off
		off++
	}
	for b := uint32(1); b < 9; b++ {
		if lo&(1<<b) != 0 {
			a.texMtxIdxOff[b-1] = off
			off++
		}
	}

	sizeOfN := func(desc uint32, direct, nIdx int) int {
		switch desc {
		case descIndex8:
			return nIdx
		case descIndex16:
			return 2 * nIdx
		case descDirect:
			return direct
		}
		return 0
	}
	sizeOf := func(desc uint32, direct int) int { return sizeOfN(desc, direct, 1) }

	// Position.
	a.posDesc = (lo >> 9) & 3
	a.posFmt = (g0 >> 1) & 7
	a.posComps = 2
	if g0&1 != 0 {
		a.posComps = 3
	}
	a.posFrac = (g0 >> 4) & 0x1F
	if a.posDesc != descNone {
		a.posOff = off
		off += sizeOf(a.posDesc, a.posComps*componentBytes(a.posFmt))
	}

	// Normal — read for the lighting channel. With the normal/binormal/tangent set enabled
	// the element is nine components; the normal proper is the first three. Such a normal may
	// also be indexed three times over, one index per vector (see normalIndexCount): then the
	// index at nrmOff names a THREE-component element — the normal proper — and the binormal
	// and tangent follow as indices of their own, which the lighting channel does not read but
	// every attribute after them is offset by.
	a.nrmDesc = (lo >> 11) & 3
	a.nrmFmt = (g0 >> 10) & 7
	a.nrmComps = 3
	if (g0>>9)&1 != 0 {
		a.nrmComps = 9
	}
	a.nrmIdxComps = a.nrmComps
	if normalIndexCount(g0) == 3 {
		a.nrmIdxComps = 3
	}
	if a.nrmDesc != descNone {
		a.nrmOff = off
		off += sizeOfN(a.nrmDesc, a.nrmComps*componentBytes(a.nrmFmt), normalIndexCount(g0))
	}

	// Colour 0.
	a.col0Desc = (lo >> 13) & 3
	a.col0Comp = (g0 >> 14) & 7
	if a.col0Desc != descNone {
		a.col0Off = off
		off += sizeOf(a.col0Desc, colorBytes(a.col0Comp))
	}

	// Colour 1, stepped over.
	off += sizeOf((lo>>15)&3, colorBytes((g0>>18)&7))
	texBytes := func(elem, format uint32) int {
		comps := 1
		if elem != 0 {
			comps = 2
		}
		return comps * componentBytes(format)
	}
	tex := func(k int) uint32 { return (hi >> (2 * uint32(k))) & 3 }

	// The eight texture coordinates. Coordinate 0's element/format/fraction live in the first
	// general register above the colours; 1 through 7 are packed nine bits apiece across the
	// second and third, which is why their fields are spelled out rather than indexed.
	texFields := [maxTexCoord][3]uint32{
		{(g0 >> 21) & 1, (g0 >> 22) & 7, (g0 >> 25) & 0x1F},
		{(g1 >> 0) & 1, (g1 >> 1) & 7, (g1 >> 4) & 0x1F},
		{(g1 >> 9) & 1, (g1 >> 10) & 7, (g1 >> 13) & 0x1F},
		{(g1 >> 18) & 1, (g1 >> 19) & 7, (g1 >> 22) & 0x1F},
		{(g1 >> 27) & 1, (g1 >> 28) & 7, (g2 >> 0) & 0x1F},
		{(g2 >> 5) & 1, (g2 >> 6) & 7, (g2 >> 9) & 0x1F},
		{(g2 >> 14) & 1, (g2 >> 15) & 7, (g2 >> 18) & 0x1F},
		{(g2 >> 23) & 1, (g2 >> 24) & 7, (g2 >> 27) & 0x1F},
	}
	for k := 0; k < maxTexCoord; k++ {
		t := &a.tex[k]
		t.desc = tex(k)
		t.elem, t.fmt, t.frac = texFields[k][0], texFields[k][1], texFields[k][2]
		if t.desc != descNone {
			t.off = off
			off += sizeOf(t.desc, texBytes(t.elem, t.fmt))
		}
	}

	a.stride = off
	return a
}

// attrData resolves where one attribute's bytes live: inline in the vertex for a direct
// attribute, or in the attribute's array — CP base register 0xA0+attr, stride 0xB0+attr,
// pinned from the game's own GXSetArray (0x801F46DC: reg = (attr-9)|0xA0 for the base and
// |0xB0 for the stride, the base a physical address, the stride a byte) — for an indexed one.
// A vertex whose index points past the end of memory returns nil, and the caller skips the
// draw with a log rather than reading a wrong place.
func (g *gpu) attrData(m *Machine, v []byte, off int, desc uint32, attr, size int) []byte {
	switch desc {
	case descDirect:
		return v[off:]
	case descIndex8, descIndex16:
		idx := int(v[off])
		if desc == descIndex16 {
			idx = int(be16(v[off:]))
		}
		base := g.CPReg[0xA0+uint32(attr)]
		stride := g.CPReg[0xB0+uint32(attr)] & 0xFF
		a := phys(base + uint32(idx)*stride)
		if int(a)+size > len(m.RAM) {
			return nil
		}
		return m.RAM[a : int(a)+size]
	}
	return nil
}

// drawPrimitive fetches a primitive's vertices and rasterises the triangles they form.
func (g *gpu) drawPrimitive(m *Machine, prim uint32, vat, vsize int, data []byte) {
	lay := g.layout(vat)

	if lay.posDesc == descNone {
		m.logf("CP: draw with no position attribute (primitive 0x%02X)", prim)
		return
	}

	// The two halves of a draw are timed separately, because they are the two things an
	// optimisation has to choose between: the per-vertex work below (fetch, transform,
	// light) and the per-fragment work in rasterPrimitive.
	tv := m.profStart()
	g.profDraws++
	count := len(data) / vsize
	verts := make([]clipVertex, count)
	nTexGen := g.texGenCount()
	for i := 0; i < count; i++ {
		v := data[i*vsize:]
		mtxIdx := int(g.CPReg[0x30] & 0x3F)
		if lay.hasMatIdx {
			mtxIdx = int(v[lay.matIdxOff])
		}
		pos := g.attrData(m, v, lay.posOff, lay.posDesc, 0, lay.posComps*componentBytes(lay.posFmt))
		if pos == nil {
			m.logf("CP: draw's position index reads past RAM (primitive 0x%02X)", prim)
			return
		}
		mx, my, mz := g.readPos(pos, lay)
		ex, ey, ez := g.eyePos(mtxIdx, mx, my, mz)
		cx, cy, cz, cw := g.clipPos(ex, ey, ez)
		r, gg, b, a := uint8(255), uint8(255), uint8(255), uint8(255)
		if lay.col0Off >= 0 {
			if col := g.attrData(m, v, lay.col0Off, lay.col0Desc, 2, colorBytes(lay.col0Comp)); col != nil {
				r, gg, b, a = readColor(col, lay.col0Comp)
			}
		}
		var nex, ney, nez float32
		var mnx, mny, mnz float32
		if lay.nrmOff >= 0 {
			if nb := g.attrData(m, v, lay.nrmOff, lay.nrmDesc, 1, lay.nrmIdxComps*componentBytes(lay.nrmFmt)); nb != nil {
				mnx, mny, mnz = readNormal(nb, lay.nrmFmt)
				nex, ney, nez = g.normalToEye(mtxIdx, mnx, mny, mnz)
			}
		}
		// The colour channel: what the hardware rasterises is not the vertex colour but the
		// lighting channel's output, which may pass the vertex colour through, replace it
		// with a register, or scale it by ambient+lights (see gpu_light.go).
		col := g.rasChannel(m, r, gg, b, a, ex, ey, ez, nex, ney, nez)
		// Every coordinate the vertex carries, because any of them can be a generator's source.
		var vtc [maxTexCoord]texCoord
		for k := 0; k < maxTexCoord; k++ {
			ta := &lay.tex[k]
			if ta.off < 0 {
				continue
			}
			comps := 1
			if ta.elem != 0 {
				comps = 2
			}
			if tc := g.attrData(m, v, ta.off, ta.desc, 4+k, comps*componentBytes(ta.fmt)); tc != nil {
				u, vv := readTexCoord(tc, ta)
				vtc[k] = texCoord{s: u, t: vv, q: 1}
			}
		}

		// The transform unit generates this vertex's texture coordinates. What the vertex
		// carried (tu,tv) is only one possible SOURCE for them — a coordinate generated from
		// the position never touches it (see gpu_texgen.go).
		cv := clipVertex{cx: cx, cy: cy, cz: cz, cw: cw, col: col, ntc: nTexGen}
		for k := 0; k < nTexGen; k++ {
			row := g.texMtxRow(k)
			if lay.texMtxIdxOff[k] >= 0 {
				row = int(v[lay.texMtxIdxOff[k]])
			}
			cv.tc[k] = g.genTexCoord(m, k, row, mx, my, mz, mnx, mny, mnz, &vtc, &col)
		}
		verts[i] = cv
	}
	m.profEnd(bucketVertex, tv)

	if drawTrace && count > 0 {
		// The pixel tallies are lifetime counters (the profiler takes deltas of them), so
		// this draw's own numbers are a before/after difference rather than a reset.
		w0, z0, a0 := g.pixWritten, g.pixZRej, g.pixARej
		tr := m.profStart()
		g.rasterPrimitive(m, prim, verts)
		m.profEnd(bucketRaster, tr)
		// The trace reports vertex 0 where the rasteriser would put it, so the numbers here are
		// the same pixels the frame shows. A vertex behind the near plane has no screen position
		// — the clipper replaces it — so its w is reported instead of a fictitious pixel.
		c0 := verts[0]
		v0x, v0y, v0z := g.toScreen(c0.cx, c0.cy, c0.cz, c0.cw)
		v0 := screenVertex{x: v0x, y: v0y, z: v0z, col: c0.col, tc: c0.tc, ntc: c0.ntc}
		clipped := ""
		if c0.cz+c0.cw < 0 {
			clipped = fmt.Sprintf(" CLIPPED(w=%.2f)", c0.cw)
		}
		_, _, _, c0a := tevColorReg(g.TevColorReg[1])
		_, _, _, k0a := tevColorReg(g.TevKonstReg[0])
		t0 := g.texSetup(0)
		g.dumpTex0Once(m, t0.base)
		fmt.Fprintf(os.Stderr, "DRAW prim 0x%02X vat %d n %d  v0 (%.1f,%.1f,z%.0f)%s rgba %d,%d,%d,%d ntc=%d uv (%.2f,%.2f)  tex0 0x%06X fmt%X %dx%d  px w=%d zrej=%d arej=%d  stages=%d bp41=%06X af=%06X zm=%02X a0=%.0f ka0=%.0f vcd=%03X mat0=%08X amb0=%08X cc0=%08X ca0=%08X\n",
			prim, vat, count, v0.x, v0.y, v0.z, clipped, v0.col[0][0], v0.col[0][1], v0.col[0][2], v0.col[0][3], v0.ntc, v0.tc[0].s, v0.tc[0].t,
			t0.base, t0.format, t0.width, t0.height,
			g.pixWritten-w0, g.pixZRej-z0, g.pixARej-a0,
			int((g.BP[0x00]>>10)&0xF)+1, g.BP[0x41], g.BP[0xF3], g.BP[0x40], c0a, k0a,
			g.CPReg[0x50]&0xFFFF, g.XFMem[0x100C], g.XFMem[0x100A],
			g.XFMem[0x100E], g.XFMem[0x1010])
		// The scissor, decoded. Both corners carry hardware's fixed 342 bias, and the offset
		// register (0x59) shifts the box in 2-pixel units. A draw whose pixels land outside
		// this box does not reach the EFB on hardware, however plausible it looks here.
		scisY0 := int(g.BP[0x20]&0x7FF) - 342
		scisX0 := int((g.BP[0x20]>>12)&0x7FF) - 342
		scisY1 := int(g.BP[0x21]&0x7FF) - 342
		scisX1 := int((g.BP[0x21]>>12)&0x7FF) - 342
		fmt.Fprintf(os.Stderr, "  SCIS box (%d,%d)-(%d,%d) off=(%d,%d) raw20=%06X raw21=%06X raw59=%06X\n",
			scisX0, scisY0, scisX1, scisY1,
			(int(g.BP[0x59]&0x3FF))*2-342, (int((g.BP[0x59]>>10)&0x3FF))*2-342,
			g.BP[0x20], g.BP[0x21], g.BP[0x59])
		// The second half of the draw's transform state: the channel COUNT and channel 1's
		// registers (a draw lighting two channels says so here), every texture-coordinate
		// generator's info word, and the TEV ordering registers whose ras_sel field says which
		// channel each stage reads. This is the view that says what a scene ASKS FOR, which is
		// not always what the pipe implements.
		fmt.Fprintf(os.Stderr, "  XFDBG nchan=%d cc1=%08X ca1=%08X mat1=%08X amb1=%08X ntg=%d tg=[%08X %08X %08X %08X %08X %08X %08X %08X] dual=%X ord=[%06X %06X %06X %06X]\n",
			g.XFMem[0x1009], g.XFMem[0x100F], g.XFMem[0x1011], g.XFMem[0x100D], g.XFMem[0x100B],
			g.texGenCount(),
			g.XFMem[0x1040], g.XFMem[0x1041], g.XFMem[0x1042], g.XFMem[0x1043],
			g.XFMem[0x1044], g.XFMem[0x1045], g.XFMem[0x1046], g.XFMem[0x1047],
			g.XFMem[0x1012], g.BP[0x28], g.BP[0x29], g.BP[0x2A], g.BP[0x2B])
		// Every light either channel enables, as the XF holds it. The cos-attenuation triple is
		// the one to read: a spot light's a0+a1+a2 comes to exactly 1, because GXInitLightSpot
		// normalises the cone to full brightness on its own axis — which is what pins the sign
		// of the cosine in gpu_light.go.
		for li := 0; li < 8; li++ {
			if lightMask(g.XFMem[0x100E])&(1<<li) == 0 && lightMask(g.XFMem[0x100F])&(1<<li) == 0 {
				continue
			}
			b := 0x600 + li*16
			fmt.Fprintf(os.Stderr, "  LIGHT%d col=%08X a=(%g,%g,%g) k=(%g,%g,%g) pos=(%.2f,%.2f,%.2f) dir=(%.4f,%.4f,%.4f)\n",
				li, g.XFMem[b+3],
				g.xfFloat(b+4), g.xfFloat(b+5), g.xfFloat(b+6),
				g.xfFloat(b+7), g.xfFloat(b+8), g.xfFloat(b+9),
				g.xfFloat(b+10), g.xfFloat(b+11), g.xfFloat(b+12),
				g.xfFloat(b+13), g.xfFloat(b+14), g.xfFloat(b+15))
		}
		return
	}
	tr := m.profStart()
	g.rasterPrimitive(m, prim, verts)
	m.profEnd(bucketRaster, tr)
}

// readTexCoord reads one texture coordinate and scales it by its fixed-point fraction. A
// single-component coordinate leaves the second axis at zero.
func readTexCoord(b []byte, ta *texAttr) (u, v float32) {
	scale := float32(1) / float32(uint32(1)<<ta.frac)
	sz := componentBytes(ta.fmt)
	u = readComponent(b, ta.fmt) * scale
	if ta.elem != 0 {
		v = readComponent(b[sz:], ta.fmt) * scale
	}
	return u, v
}

// readNormal reads a normal's three components. Normals ignore the VAT's fraction field:
// their fixed-point formats carry an implicit scale — s8 is 1.1.6 (divide by 64), s16 is
// 1.1.14 (divide by 16384) — so a unit normal fits the type's range.
func readNormal(b []byte, format uint32) (x, y, z float32) {
	scale := float32(1)
	switch format {
	case 1: // s8
		scale = 1.0 / 64
	case 3: // s16
		scale = 1.0 / 16384
	}
	sz := componentBytes(format)
	x = readComponent(b, format) * scale
	y = readComponent(b[sz:], format) * scale
	z = readComponent(b[2*sz:], format) * scale
	return
}

// readPos reads a direct position attribute and scales it by its fixed-point fraction.
func (g *gpu) readPos(b []byte, lay attrLayout) (x, y, z float32) {
	var c [3]float32
	scale := float32(1) / float32(uint32(1)<<lay.posFrac)
	sz := componentBytes(lay.posFmt)
	for k := 0; k < lay.posComps; k++ {
		c[k] = readComponent(b[k*sz:], lay.posFmt) * scale
	}
	return c[0], c[1], c[2]
}

// readComponent reads one position/texcoord component by its format code.
func readComponent(b []byte, format uint32) float32 {
	switch format {
	case 0: // u8
		return float32(b[0])
	case 1: // s8
		return float32(int8(b[0]))
	case 2: // u16
		return float32(uint16(b[0])<<8 | uint16(b[1]))
	case 3: // s16
		return float32(int16(uint16(b[0])<<8 | uint16(b[1])))
	case 4: // f32
		return math.Float32frombits(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
	}
	return 0
}

// readColor reads a direct colour attribute by its component format.
func readColor(b []byte, comp uint32) (r, g, bl, a uint8) {
	switch comp {
	case 0: // RGB565
		v := uint16(b[0])<<8 | uint16(b[1])
		return expand5(v >> 11), expand6((v >> 5) & 0x3F), expand5(v & 0x1F), 255
	case 1: // RGB888
		return b[0], b[1], b[2], 255
	case 2: // RGB888x
		return b[0], b[1], b[2], 255
	case 3: // RGBA4444
		v := uint16(b[0])<<8 | uint16(b[1])
		return expand4(v >> 12), expand4((v >> 8) & 0xF), expand4((v >> 4) & 0xF), expand4(v & 0xF)
	case 5: // RGBA8888
		return b[0], b[1], b[2], b[3]
	}
	return 255, 255, 255, 255
}

func expand4(v uint16) uint8 { return uint8(v<<4 | v) }
func expand5(v uint16) uint8 { return uint8(v<<3 | v>>2) }
func expand6(v uint16) uint8 { return uint8(v<<2 | v>>4) }

// rasterPrimitive turns a primitive's vertex list into triangles and draws each, cutting every
// one against the near plane on the way (gpu_clip.go) — this is where a triangle first exists
// as a triangle, so it is where clipping belongs. The triangle-forming primitives are handled;
// lines and points wait for a frame that uses them, and are logged once when one appears rather
// than drawn wrong.
func (g *gpu) rasterPrimitive(m *Machine, prim uint32, v []clipVertex) {
	// Two scratch polygons for the whole primitive: the clipper ping-pongs between them as it
	// cuts by each plane, so a strip of hundreds of triangles still allocates once. Cutting a
	// triangle by the five planes can leave at most eight vertices.
	buf := make([]clipVertex, 0, 8)
	scratch := make([]clipVertex, 0, 8)

	// The TEV's configuration, decoded once for the whole primitive. It cannot change while
	// this runs: only the command processor writes a BP register, and it is not running.
	tev := g.tevstate()

	// The setup stage's output, reused draw to draw so a strip of hundreds of triangles does
	// not allocate. The fill takes the whole primitive at once — that is what lets it split
	// the screen between workers instead of splitting the triangles.
	g.tris = g.tris[:0]
	draw := func(a, b, c clipVertex) { buf, scratch = g.clipAndSetup(m, buf, scratch, a, b, c) }

	switch prim {
	case 0x80, 0x88: // quads
		for i := 0; i+4 <= len(v); i += 4 {
			draw(v[i], v[i+1], v[i+2])
			draw(v[i], v[i+2], v[i+3])
		}
	case 0x90: // triangles
		for i := 0; i+2 < len(v); i += 3 {
			draw(v[i], v[i+1], v[i+2])
		}
	case 0x98: // triangle strip
		for i := 0; i+2 < len(v); i++ {
			if i&1 == 0 {
				draw(v[i], v[i+1], v[i+2])
			} else {
				draw(v[i+1], v[i], v[i+2])
			}
		}
	case 0xA0: // triangle fan
		for i := 1; i+1 < len(v); i++ {
			draw(v[0], v[i], v[i+1])
		}
	}

	g.fill(m, &tev, g.tris)
}
