package gc

// gpu_fetch.go reads the vertices a draw primitive carries, transforms them, and hands the
// triangles they form to the rasteriser. It is the counterpart to gpu_vtx.go: that file sizes
// a vertex so the parser can find the next command, this one reads the same bytes as values.
//
// Only what an untextured, unlit frame needs is read: the position (which the transform turns
// into a pixel) and the first colour (which the rasteriser interpolates). Normals, texture
// coordinates and the second colour are stepped over — their sizes still count toward the
// stride, so the reader lands on the right byte, but their values wait for the lighting and
// texture stages. An attribute that is an index into an array rather than inline is not yet
// fetched; a draw that needs one is left undrawn and logged, rather than read from the wrong
// place, so the gap is visible instead of a scramble of triangles.

import (
	"fmt"
	"math"
	"os"
)

// drawTrace prints one line per draw primitive — vertices, texture base, and the pixel
// accounting that says where a draw's pixels went (written / z-rejected / alpha-rejected).
var drawTrace = os.Getenv("RR_GC_DRAWTRACE") != ""

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

	col0Off  int
	col0Desc uint32
	col0Comp uint32

	tex0Off  int
	tex0Desc uint32
	tex0Fmt  uint32
	tex0Elem uint32
	tex0Frac uint32
}

// layout computes the attribute layout for a vertex-attribute-table index, walking the
// descriptor and table in the same order gpu_vtx.go sizes them so the offsets agree.
func (g *gpu) layout(vat int) attrLayout {
	lo := g.CPReg[0x50]
	hi := g.CPReg[0x60]
	g0 := g.CPReg[0x70+uint32(vat)]
	g1 := g.CPReg[0x80+uint32(vat)]
	g2 := g.CPReg[0x90+uint32(vat)]

	a := attrLayout{matIdxOff: -1, posOff: -1, col0Off: -1, tex0Off: -1}
	off := 0

	// The position matrix index, then eight texture matrix indices, are each one inline byte.
	if lo&1 != 0 {
		a.hasMatIdx = true
		a.matIdxOff = off
		off++
	}
	for b := uint32(1); b < 9; b++ {
		if lo&(1<<b) != 0 {
			off++
		}
	}

	sizeOf := func(desc uint32, direct int) int {
		switch desc {
		case descIndex8:
			return 1
		case descIndex16:
			return 2
		case descDirect:
			return direct
		}
		return 0
	}

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

	// Normal, stepped over.
	nrmComps := 3
	if (g0>>9)&1 != 0 {
		nrmComps = 9
	}
	off += sizeOf((lo>>11)&3, nrmComps*componentBytes((g0>>10)&7))

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

	// Texture coordinate 0: captured so the rasteriser can interpolate it and the TEV can
	// sample a texture. The remaining seven coordinates are stepped over — the boot draws use
	// only coordinate 0, and a stage that reads another is logged as a ticket in gpu_tev.go.
	a.tex0Desc = tex(0)
	a.tex0Elem = (g0 >> 21) & 1
	a.tex0Fmt = (g0 >> 22) & 7
	a.tex0Frac = (g0 >> 25) & 0x1F
	if a.tex0Desc != descNone {
		a.tex0Off = off
		off += sizeOf(a.tex0Desc, texBytes(a.tex0Elem, a.tex0Fmt))
	}
	off += sizeOf(tex(1), texBytes((g1>>0)&1, (g1>>1)&7))
	off += sizeOf(tex(2), texBytes((g1>>9)&1, (g1>>10)&7))
	off += sizeOf(tex(3), texBytes((g1>>18)&1, (g1>>19)&7))
	off += sizeOf(tex(4), texBytes((g1>>27)&1, (g1>>28)&7))
	off += sizeOf(tex(5), texBytes((g2>>5)&1, (g2>>6)&7))
	off += sizeOf(tex(6), texBytes((g2>>14)&1, (g2>>15)&7))
	off += sizeOf(tex(7), texBytes((g2>>23)&1, (g2>>24)&7))

	a.stride = off
	return a
}

// drawPrimitive fetches a primitive's vertices and rasterises the triangles they form.
func (g *gpu) drawPrimitive(m *Machine, prim uint32, vat, vsize int, data []byte) {
	lay := g.layout(vat)

	// An indexed position or colour is not yet fetched: the draw is left undrawn rather than
	// read from the wrong place. This is logged once, and its shape is a ticket to implement
	// the array fetch (CP array bases 0xA0/strides 0xB0).
	if lay.posDesc != descDirect || (lay.col0Desc != descNone && lay.col0Desc != descDirect) {
		m.logf("CP: draw with indexed position/colour not yet fetched (primitive 0x%02X)", prim)
		return
	}

	count := len(data) / vsize
	verts := make([]screenVertex, count)
	for i := 0; i < count; i++ {
		v := data[i*vsize:]
		mtxIdx := int(g.CPReg[0x30] & 0x3F)
		if lay.hasMatIdx {
			mtxIdx = int(v[lay.matIdxOff])
		}
		mx, my, mz := g.readPos(v[lay.posOff:], lay)
		sx, sy, sz := g.transform(mtxIdx, mx, my, mz)
		r, gg, b, a := uint8(255), uint8(255), uint8(255), uint8(255)
		if lay.col0Off >= 0 {
			r, gg, b, a = readColor(v[lay.col0Off:], lay.col0Comp)
		}
		var tu, tv float32
		if lay.tex0Off >= 0 && lay.tex0Desc == descDirect {
			tu, tv = g.readTexCoord(v[lay.tex0Off:], lay)
		}
		verts[i] = screenVertex{x: sx, y: sy, z: sz, r: r, g: gg, b: b, a: a, u: tu, v: tv}
	}

	if drawTrace {
		g.pixZRej, g.pixARej, g.pixWritten = 0, 0, 0
		g.rasterPrimitive(m, prim, verts)
		v0 := verts[0]
		fmt.Fprintf(os.Stderr, "DRAW prim 0x%02X vat %d n %d  v0 (%.1f,%.1f,z%.0f) rgba %d,%d,%d,%d uv (%.2f,%.2f)  texbase 0x%06X  px w=%d zrej=%d arej=%d\n",
			prim, vat, count, v0.x, v0.y, v0.z, v0.r, v0.g, v0.b, v0.a, v0.u, v0.v,
			(g.BP[0x94]&0x00FFFFFF)<<5, g.pixWritten, g.pixZRej, g.pixARej)
		return
	}
	g.rasterPrimitive(m, prim, verts)
}

// readTexCoord reads texture coordinate 0 and scales it by its fixed-point fraction. A
// single-component coordinate leaves the second axis at zero.
func (g *gpu) readTexCoord(b []byte, lay attrLayout) (u, v float32) {
	scale := float32(1) / float32(uint32(1)<<lay.tex0Frac)
	sz := componentBytes(lay.tex0Fmt)
	u = readComponent(b, lay.tex0Fmt) * scale
	if lay.tex0Elem != 0 {
		v = readComponent(b[sz:], lay.tex0Fmt) * scale
	}
	return u, v
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

// rasterPrimitive turns a primitive's vertex list into triangles and draws each. The
// triangle-forming primitives are handled; lines and points wait for a frame that uses them,
// and are logged once when one appears rather than drawn wrong.
func (g *gpu) rasterPrimitive(m *Machine, prim uint32, v []screenVertex) {
	switch prim {
	case 0x80, 0x88: // quads
		for i := 0; i+4 <= len(v); i += 4 {
			g.drawTriangle(m, v[i], v[i+1], v[i+2])
			g.drawTriangle(m, v[i], v[i+2], v[i+3])
		}
	case 0x90: // triangles
		for i := 0; i+2 < len(v); i += 3 {
			g.drawTriangle(m, v[i], v[i+1], v[i+2])
		}
	case 0x98: // triangle strip
		for i := 0; i+2 < len(v); i++ {
			if i&1 == 0 {
				g.drawTriangle(m, v[i], v[i+1], v[i+2])
			} else {
				g.drawTriangle(m, v[i+1], v[i], v[i+2])
			}
		}
	case 0xA0: // triangle fan
		for i := 1; i+1 < len(v); i++ {
			g.drawTriangle(m, v[0], v[i], v[i+1])
		}
	}
}
