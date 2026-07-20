package xbox

// nv2a_vertex.go is the Kelvin vertex front end: it latches the 16 vertex-attribute
// array formats, accumulates inline vertex data and index/array draw commands between
// SET_BEGIN_END pairs, fetches and transforms each vertex through the LLE vertex
// program (nv2a_vsh.go), assembles primitives, and hands triangles to the rasteriser
// (nv2a_raster.go). It is the analogue of the PICA200 draw() vertex loop in
// n3ds/gpu_raster.go.
//
// The Xbox D3D runtime submits the loading screen entirely through the inline path:
// BEGIN(QUAD_STRIP), 28 INLINE_ARRAY words (4 vertices of float4 pos + D3DCOLOR +
// float2 uv), END — one quad per draw. Indexed (ARRAY_ELEMENT16/32) and array
// (DRAW_ARRAYS) submissions fetch from the vertex arrays the offset registers point
// at; they appear once real geometry replaces the loading UI.

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

// nvDrawTrace: RR_NV_DRAW=N dumps the state and first vertices of the first N draws.
var nvDrawTrace, _ = strconv.Atoi(os.Getenv("RR_NV_DRAW"))

const (
	kelvinCtxDmaVertexA = 0x019C // SET_CONTEXT_DMA_VERTEX_A
	kelvinCtxDmaVertexB = 0x01A0 // SET_CONTEXT_DMA_VERTEX_B

	kelvinVtxArrayOffset = 0x1720 // +4*i, i=0..15: bits 30:0 offset, bit 31 selects DMA B
	kelvinVtxArrayFormat = 0x1760 // +4*i: bits 3:0 type, 7:4 size, 15:8 stride

	kelvinBeginEnd     = 0x17FC // arg = primitive type; 0 = END
	kelvinElement16    = 0x1800 // two u16 indices per word
	kelvinElement32    = 0x1808 // one u32 index per word
	kelvinDrawArrays   = 0x1810 // bits 23:0 first index, 31:24 count-1
	kelvinInlineArray  = 0x1818 // raw vertex words per the attribute formats
	kelvinVertexData4C = 0x1940 // +4*i: persistent attribute i as 4 unsigned bytes
)

// Vertex attribute data types (SET_VERTEX_DATA_ARRAY_FORMAT).
const (
	vtxTypeUBD3D = 0 // 4 unsigned bytes, D3DCOLOR order (B,G,R,A in memory)
	vtxTypeS1    = 1 // normalized int16
	vtxTypeF     = 2 // float32
	vtxTypeUBOGL = 4 // 4 unsigned bytes, RGBA order
	vtxTypeS32K  = 5 // un-normalized int16
	vtxTypeCMP   = 6 // packed 11:11:10 signed normalized
)

// Primitive types (SET_BEGIN_END).
const (
	primPoints = iota + 1
	primLines
	primLineLoop
	primLineStrip
	primTriangles
	primTriStrip
	primTriFan
	primQuads
	primQuadStrip
	primPolygon
)

// kelvinVtx is one transformed vertex as the rasteriser consumes it: screen-space
// position (x, y in pixels, z in depth-buffer units, w = the clip-space w the
// D3D-appended epilogue preserves for perspective correction), colors, fog, and the
// four texture coordinate sets.
type kelvinVtx struct {
	pos    [4]float32
	d0, d1 [4]float32
	fog    float32
	uv     [4][4]float32
}

// vtxFmt is one attribute's decoded array format.
type vtxFmt struct {
	typ, size int
	stride    uint32
	words     int // dwords one vertex of this attribute occupies in the inline stream
}

func (g *pgraph) attrFormat(i int) vtxFmt {
	f := g.Regs[kelvinVtxArrayFormat>>2+uint32(i)]
	vf := vtxFmt{typ: int(f & 0xF), size: int(f >> 4 & 0xF), stride: f >> 8 & 0xFF}
	switch vf.typ {
	case vtxTypeUBD3D, vtxTypeUBOGL:
		if vf.size > 0 {
			vf.words = 1
		}
	case vtxTypeS1, vtxTypeS32K:
		vf.words = (vf.size + 1) / 2
	case vtxTypeF:
		vf.words = vf.size
	case vtxTypeCMP:
		if vf.size > 0 {
			vf.words = 1
		}
	}
	return vf
}

// decodeAttr decodes one attribute's raw dwords into a vec4 (missing components take
// the 0,0,0,1 defaults).
func decodeAttr(vf *vtxFmt, w []uint32) [4]float32 {
	val := [4]float32{0, 0, 0, 1}
	switch vf.typ {
	case vtxTypeUBD3D:
		// D3DCOLOR: the dword is AARRGGBB, so x=R, y=G, z=B, w=A.
		val[0] = float32(w[0]>>16&0xFF) / 255
		val[1] = float32(w[0]>>8&0xFF) / 255
		val[2] = float32(w[0]&0xFF) / 255
		val[3] = float32(w[0]>>24&0xFF) / 255
	case vtxTypeUBOGL:
		for i := 0; i < vf.size && i < 4; i++ {
			val[i] = float32(w[0]>>(8*uint(i))&0xFF) / 255
		}
	case vtxTypeS1, vtxTypeS32K:
		for i := 0; i < vf.size && i < 4; i++ {
			raw := int16(w[i/2] >> (16 * uint(i&1)))
			if vf.typ == vtxTypeS1 {
				v := float32(raw) / 32767
				if v < -1 {
					v = -1
				}
				val[i] = v
			} else {
				val[i] = float32(raw)
			}
		}
	case vtxTypeF:
		for i := 0; i < vf.size && i < 4; i++ {
			val[i] = math.Float32frombits(w[i])
		}
	case vtxTypeCMP:
		sext := func(v uint32, bits uint) float32 {
			s := int32(v<<(32-bits)) >> (32 - bits)
			f := float32(s) / float32(int32(1)<<(bits-1)-1)
			if f < -1 {
				f = -1
			}
			return f
		}
		val[0] = sext(w[0]&0x7FF, 11)
		val[1] = sext(w[0]>>11&0x7FF, 11)
		val[2] = sext(w[0]>>22&0x3FF, 10)
	}
	return val
}

// beginEnd runs SET_BEGIN_END: a nonzero arg opens a primitive batch, zero closes it
// and executes the accumulated draw.
func (g *pgraph) beginEnd(arg uint32) {
	if arg != 0 {
		g.prim = arg
		g.inline = g.inline[:0]
		g.elems = g.elems[:0]
		g.ranges = g.ranges[:0]
		return
	}
	if g.prim == 0 {
		return // END with no BEGIN: the stream can carry one at init; nothing to draw
	}
	g.runDraw()
	g.prim = 0
}

// runDraw executes the accumulated batch: build the vertex list, transform, assemble,
// rasterise.
func (g *pgraph) runDraw() {
	g.Draws++
	// Vertex bucket: fetch + transform + shader, up to the point assemble takes over the
	// rasterise timing (profile.go). The early returns below are halts or empty brackets —
	// paths that do no fill and whose vertex time is immaterial — so leaving tv unbooked on
	// them is correct, not a leak.
	tv := g.m.profStart()
	ff := g.Regs[kelvinTransformExecMode>>2]&3 != 2
	g.ffFragHalt = ""
	if ff {
		// Fixed-function T&L. The TRANSFORM is modelled: clip = pos · composite
		// (SET_COMPOSITE_MATRIX 0x0680, row-vector — NDC-verified in Part XII), then
		// the viewport in the aliased c58/c59 slots, exactly what the program-mode
		// epilogue computes. Lighting (0x0314) and FF texgen are NOT modelled — no
		// reachable on-screen FF draw exists to verify them against — so a draw that
		// needs either arms a fragment-level halt: an off-screen or degenerate FF
		// draw passes through honestly, while one that would actually paint still
		// halts and names the gap instead of shading something plausible.
		if g.Regs[kelvinLightingEnable>>2]&1 != 0 {
			g.ffFragHalt = "lighting is enabled and the FF lighting equation/state is unmodelled"
		} else {
			shader := g.Regs[0x1E70>>2]
			for u := uint32(0); u < 4; u++ {
				if g.Regs[(kelvinTexControl+u*0x40)>>2]>>30&1 != 0 && shader>>(5*u)&0x1F != 0 {
					g.ffFragHalt = "textures are enabled and FF texgen is unmodelled"
				}
			}
		}
	}

	// The enabled attribute set and the inline stride, from the latched formats.
	var fmts [16]vtxFmt
	stride := 0
	for i := range fmts {
		fmts[i] = g.attrFormat(i)
		stride += fmts[i].words
	}

	trace := g.Draws <= nvDrawTrace
	if trace {
		g.traceDraw(&fmts, stride)
	}

	var verts []kelvinVtx
	switch {
	case len(g.inline) > 0:
		if len(g.elems) > 0 || len(g.ranges) > 0 {
			g.m.CPU.Halt("nv2a: draw %d mixes inline and indexed vertex submission", g.Draws)
			return
		}
		if stride == 0 {
			g.m.CPU.Halt("nv2a: draw %d has inline data but no enabled attributes", g.Draws)
			return
		}
		if len(g.inline)%stride != 0 {
			g.m.CPU.Halt("nv2a: draw %d inline data %d words is not a multiple of the vertex stride %d",
				g.Draws, len(g.inline), stride)
			return
		}
		n := len(g.inline) / stride
		verts = make([]kelvinVtx, 0, n)
		for vi := 0; vi < n; vi++ {
			w := g.inline[vi*stride : (vi+1)*stride]
			var in [16][4]float32
			copy(in[:], g.vtxAttr[:])
			for a := 0; a < 16; a++ {
				if fmts[a].words == 0 {
					continue
				}
				in[a] = decodeAttr(&fmts[a], w[:fmts[a].words])
				w = w[fmts[a].words:]
			}
			var v kelvinVtx
			var ok bool
			if ff {
				v, ok = g.transformFF(&in), true
			} else {
				v, ok = g.transform(&in, trace && vi < 8)
			}
			if !ok {
				return
			}
			verts = append(verts, v)
		}
	case len(g.elems) > 0 || len(g.ranges) > 0:
		idx := g.elems
		for _, r := range g.ranges {
			for k := uint32(0); k < r[1]; k++ {
				idx = append(idx, r[0]+k)
			}
		}
		verts = make([]kelvinVtx, 0, len(idx))
		for vi, ix := range idx {
			in, ok := g.fetchArrayVertex(&fmts, ix)
			if !ok {
				return
			}
			var v kelvinVtx
			if ff {
				v = g.transformFF(&in)
			} else {
				v, ok = g.transform(&in, trace && vi < 8)
				if !ok {
					return
				}
			}
			verts = append(verts, v)
		}
	default:
		return // an empty BEGIN/END pair (state-only bracket)
	}

	g.m.profEnd(bucketVertex, tv)
	g.assemble(verts)
}

// fetchArrayVertex reads one indexed vertex from the attribute arrays.
func (g *pgraph) fetchArrayVertex(fmts *[16]vtxFmt, ix uint32) ([16][4]float32, bool) {
	var in [16][4]float32
	copy(in[:], g.vtxAttr[:])
	for a := 0; a < 16; a++ {
		vf := &fmts[a]
		if vf.words == 0 {
			continue
		}
		off := g.Regs[kelvinVtxArrayOffset>>2+uint32(a)]
		dmaReg := uint32(kelvinCtxDmaVertexA)
		if off>>31 != 0 {
			dmaReg = kelvinCtxDmaVertexB
		}
		base, limit := g.m.dmaObjectTarget(g.Regs[dmaReg>>2])
		addr := base + (off & 0x7FFFFFFF) + ix*vf.stride
		if base == 0 && limit == 0 {
			g.m.CPU.Halt("nv2a: draw %d attr %d fetches through an unbound vertex DMA context", g.Draws, a)
			return in, false
		}
		var w [4]uint32
		for k := 0; k < vf.words; k++ {
			w[k] = g.m.read32(addr + uint32(k)*4)
		}
		in[a] = decodeAttr(vf, w[:vf.words])
	}
	return in, true
}

// transformFF runs one vertex through the fixed-function transform: clip = pos · M
// with M the SET_COMPOSITE_MATRIX (0x0680) applied row-vector (the convention Part XII
// verified produces valid NDC), then the viewport scale/offset in the aliased c58/c59
// constant slots — the same mapping every program-mode epilogue computes, and the same
// slots the hardware's viewport methods write. Diffuse/specular pass through from
// attributes 3/4 and texcoords from 9-12; a draw that needs lighting or texgen never
// gets here with those visible (runDraw arms the fragment halt).
func (g *pgraph) transformFF(in *[16][4]float32) kelvinVtx {
	pos := in[0]
	var clip [4]float32
	base := uint32(kelvinCompositeMatrix) >> 2
	for j := 0; j < 4; j++ {
		var s float32
		for i := 0; i < 4; i++ {
			s += pos[i] * math.Float32frombits(g.Regs[base+uint32(i*4+j)])
		}
		clip[j] = s
	}
	if clip[3] == 0 {
		// A degenerate vertex (the reachable FF draws park 800 of them at the origin):
		// an all-zero position gives every triangle zero area and the raster culls it.
		return kelvinVtx{}
	}
	f := math.Float32frombits
	sc, off := &g.Const[vshSlotViewportScale], &g.Const[vshSlotViewportOffset]
	v := kelvinVtx{d0: in[3], d1: in[4], uv: [4][4]float32{in[9], in[10], in[11], in[12]}}
	v.pos[0] = clip[0]/clip[3]*f(sc[0]) + f(off[0])
	v.pos[1] = clip[1]/clip[3]*f(sc[1]) + f(off[1])
	v.pos[2] = clip[2]/clip[3]*f(sc[2]) + f(off[2])
	v.pos[3] = clip[3]
	return v
}

// transform runs one vertex through the vertex program and maps the output registers
// onto the rasteriser's vertex.
func (g *pgraph) transform(in *[16][4]float32, trace bool) (kelvinVtx, bool) {
	var out [13][4]float32
	if !g.vshRun(in, &out) {
		return kelvinVtx{}, false
	}
	v := kelvinVtx{
		pos: out[0],
		d0:  out[3],
		d1:  out[4],
		fog: out[5][0],
		uv:  [4][4]float32{out[9], out[10], out[11], out[12]},
	}
	if trace {
		fmt.Printf("  vtx in v0=%v v3=%v v9=%v -> pos=%v d0=%v uv0=%v\n",
			in[0], in[3], in[9], v.pos, v.d0, v.uv[0])
	}
	return v, true
}

// assemble cuts the vertex list into triangles per the BEGIN primitive type and
// rasterises them. Point and line primitives halt loudly: silently dropping them
// would corrupt the frame without naming the gap.
//
// A pixel-heavy draw rasterises its triangles on a scanline-interleaved worker pool
// (rasterParallel): every worker walks the SAME triangle list in the same order but
// owns only rows py %% workers == lane, so each pixel is computed exactly once, by
// exactly the arithmetic the serial path would use, with triangles retiring in
// submission order per pixel — the output is byte-identical by construction, and
// the frame hashes verify it (see Part XV). Draws that carry per-fragment
// instruments/hooks, an armed fragment halt, or a stage-mode combination the
// per-draw prevalidation can't clear stay on the serial path.
func (g *pgraph) assemble(verts []kelvinVtx) {
	tris := g.triScratch[:0]
	tri := func(a, b, c int) {
		tris = append(tris, [3]int{a, b, c})
	}
	n := len(verts)
	switch g.prim {
	case primTriangles:
		for i := 0; i+2 < n; i += 3 {
			tri(i, i+1, i+2)
		}
	case primTriStrip:
		for i := 0; i+2 < n; i++ {
			if i&1 == 0 {
				tri(i, i+1, i+2)
			} else {
				tri(i+1, i, i+2)
			}
		}
	case primTriFan, primPolygon:
		for i := 1; i+1 < n; i++ {
			tri(0, i, i+1)
		}
	case primQuads:
		for i := 0; i+3 < n; i += 4 {
			tri(i, i+1, i+2)
			tri(i, i+2, i+3)
		}
	case primQuadStrip:
		// Quad i is vertices (2i, 2i+1, 2i+3, 2i+2), like GL.
		for i := 0; i+3 < n; i += 2 {
			tri(i, i+1, i+3)
			tri(i, i+3, i+2)
		}
	default:
		g.m.CPU.Halt("nv2a: draw %d primitive type %d not implemented", g.Draws, g.prim)
		return
	}
	g.triScratch = tris
	if len(tris) == 0 {
		return
	}
	// Clip against the near plane before rasterising: a triangle with a vertex at or behind
	// the eye (clip w <= 0) otherwise projects to a framebuffer-spanning garbage triangle
	// (nv2a_clip.go). A draw that does not cross returns here unchanged, byte-for-byte.
	verts, tris = g.clipNearPlane(verts, tris)
	if len(tris) == 0 {
		return
	}
	if !g.rastValid {
		if !g.rasterStateDecode(&g.rast) {
			return
		}
		g.rastValid = true
	}
	// Rasterise bucket: the triangle fill, whether it runs on the worker pool or serially.
	// Timing it here, on the machine's own goroutine and after rasterParallel's join,
	// captures the parallel region's wall-clock — the number that matters — without the
	// accumulator ever being touched by a worker (profile.go).
	tr := g.m.profStart()
	if g.rasterParallelOK(&g.rast, verts, tris) {
		g.rasterParallel(verts, tris)
		g.m.profEnd(bucketRaster, tr)
		return
	}
	var rs rstats
	for _, t := range tris {
		g.rasterTri(&verts[t[0]], &verts[t[1]], &verts[t[2]], &rs)
		if g.m.CPU.Halted {
			break
		}
	}
	g.mergeStats(&rs)
	g.m.profEnd(bucketRaster, tr)
}

// traceDraw prints the draw's decoded fetch/raster state (RR_NV_DRAW).
func (g *pgraph) traceDraw(fmts *[16]vtxFmt, stride int) {
	fmt.Printf("nv2a draw %d: prim=%d inline=%d words elems=%d ranges=%d stride=%d\n",
		g.Draws, g.prim, len(g.inline), len(g.elems), len(g.ranges), stride)
	for a := 0; a < 16; a++ {
		if fmts[a].words != 0 {
			fmt.Printf("  attr%-2d type=%d size=%d stride=%d words=%d\n",
				a, fmts[a].typ, fmts[a].size, fmts[a].stride, fmts[a].words)
		}
	}
	fmt.Printf("  progStart=%d constLoadAt=%d execMode=%08X surf fmt=%08X pitch=%08X color=%08X zeta=%08X clipH=%08X clipV=%08X\n",
		g.Regs[kelvinProgStart>>2], g.ConstLoad, g.Regs[kelvinTransformExecMode>>2],
		g.Regs[kelvinSurfaceFormat>>2], g.Regs[kelvinSurfacePitch>>2],
		g.Regs[kelvinSurfaceColorOffset>>2], g.Regs[kelvinSurfaceZetaOffset>>2],
		g.Regs[kelvinSurfaceClipH>>2], g.Regs[kelvinSurfaceClipV>>2])
	fmt.Printf("  blend en=%d sf=%03X df=%03X eq=%04X  alpha en=%d func=%03X ref=%d  depth en=%d func=%03X wr=%d  cull en=%d face=%03X front=%03X  combiners=%d shader=%08X\n",
		g.Regs[0x304>>2], g.Regs[0x344>>2], g.Regs[0x348>>2], g.Regs[0x350>>2],
		g.Regs[0x300>>2], g.Regs[0x33C>>2], g.Regs[0x340>>2],
		g.Regs[0x30C>>2], g.Regs[0x354>>2], g.Regs[0x35C>>2],
		g.Regs[0x308>>2], g.Regs[0x39C>>2], g.Regs[0x3A0>>2],
		g.Regs[0x1E60>>2]&0xFF, g.Regs[0x1E70>>2])
	for u := 0; u < 4; u++ {
		ctl := g.Regs[(0x1B0C+uint32(u)*0x40)>>2]
		if ctl>>30&1 == 0 {
			continue
		}
		fmt.Printf("  tex%d off=%08X fmt=%08X addr=%08X ctl0=%08X filt=%08X rect=%08X\n",
			u, g.Regs[(0x1B00+uint32(u)*0x40)>>2], g.Regs[(0x1B04+uint32(u)*0x40)>>2],
			g.Regs[(0x1B08+uint32(u)*0x40)>>2], ctl,
			g.Regs[(0x1B14+uint32(u)*0x40)>>2], g.Regs[(0x1B1C+uint32(u)*0x40)>>2])
	}
	if nvVSTrace {
		pc := int(g.Regs[kelvinProgStart>>2]) % vshProgSlots
		for k := 0; k < vshProgSlots; k++ {
			i := (pc + k) % vshProgSlots
			fmt.Printf("  vsh[%3d] %s\n", i, g.vshDisasm(i))
			if g.vshDecode(i).final {
				break
			}
		}
	}
}
