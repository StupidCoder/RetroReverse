package xbox

// nv2a_pgraph.go is the NV2A graphics engine (PGRAPH): it receives the (subchannel,
// method, argument) writes the PFIFO pusher decodes (nv2a_pfifo.go) and turns them into
// rendered pixels. This is the analogue of the PICA200 register file + pipeline in
// n3ds/gpu.go: most methods latch state into an object register file, and a few are
// triggers (BEGIN/END a primitive, the inline vertex-data FIFO, the framebuffer-clear
// surface) that run the vertex/combiner/raster pipeline.
//
// A subchannel is bound to an object class by method 0x0000 (NV_SET_OBJECT), whose
// argument is a RAMHT handle. On the Xbox the Direct3D runtime binds the 3D class
// (NV20/Kelvin, class 0x0097) to one subchannel and uses it for everything; the other
// classes it creates (2D, the surface/clip helpers) are set up but the frame's geometry
// all flows through the 3D object. The class numbers and the method map are the NV2A
// hardware's, not any game's — the sanctioned platform-spec exception, exactly as the
// PICA200 register ids are in n3ds.
//
// Bring-up discipline (house style, same as the kernel HLE): while the method map is
// being learned, `survey` mode records every (subchannel, method) it sees without acting,
// so one long run states the full command surface the title exercises. Once a method's
// semantics are pinned from the live stream it graduates into the real dispatch, and an
// unmodelled method in non-survey mode halts and names itself rather than drawing
// something plausible.

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"sort"
)

// The NV2A 3D object class (NV20_KELVIN_PRIMITIVE). D3D binds it to a subchannel via
// NV_SET_OBJECT; every 3D method arrives on that subchannel.
const (
	classKelvin = 0x0097 // NV20_KELVIN_PRIMITIVE (the Xbox 3D class)

	nvSetObject = 0x0000 // method 0 on any subchannel: bind a RAMHT handle to it
)

// pgraph is the graphics engine state owned by the Machine.
type pgraph struct {
	m *Machine

	// subObject[subchan] is the RAMHT handle last bound to that subchannel (method 0).
	subObject [8]uint32
	// subClass[subchan] is the resolved object class for that subchannel (via RAMHT).
	subClass [8]uint32

	// The Kelvin object's method register file. NV2A 3D methods are byte offsets below
	// 0x2000; storing them densely (offset>>2) keeps the pipeline's reads O(1). Only the
	// 3D object uses this; other classes' state is tracked ad hoc as needed.
	Regs [0x800]uint32

	// Transform-program machinery (nv2a_vsh.go): the 136-slot program store, the
	// 192-vec4 constant store, the load cursors, and the 4-dword upload staging
	// latches (an upload may split across two DMA kicks, so they are state).
	Prog      [vshProgSlots][4]uint32
	Const     [vshConstSlots][4]uint32
	ProgLoad  uint32
	ConstLoad uint32
	progBuf   [4]uint32
	progBufN  int
	constBuf  [4]uint32
	constBufN int

	// Vertex/primitive state (nv2a_vertex.go): the open BEGIN's primitive type, the
	// accumulated inline words / indices / array ranges, the persistent per-attribute
	// values (SET_VERTEX_DATA*), and the draw counter.
	prim    uint32
	inline  []uint32
	elems   []uint32
	ranges  [][2]uint32
	vtxAttr [16][4]float32
	Draws   int

	// Fragment tallies, maintained unconditionally (like the GameCube's) and read by the
	// frame profiler (profile.go) as per-frame deltas. Kept as running totals rather than
	// profiler-gated so the profiler never has to branch per fragment; they are summed in
	// from per-lane rstats after the parallel fill joins, never incremented across
	// goroutines (nv2a_raster.go mergeStats). Transient — outside the savestate.
	pixWritten, pixZRej, pixARej int

	lowWriteDraw int // RR_LOWWRITE: last draw reported, one line per draw
	// ffFragHalt, when non-empty, names why the current fixed-function draw cannot
	// shade a fragment (lighting/texgen unmodelled); the raster halts with it the
	// moment a fragment of that draw would actually land (nv2a_vertex.go runDraw).
	ffFragHalt string

	// presented is the colour surface as it stood at the last FLIP_STALL — the buffer the
	// title handed to the screen (nv2a_frame.go). It is how this machine knows what the TV
	// shows, because the CRTC scanout registers cannot say (see RenderPresented). Derived,
	// not state: a restored savestate simply has none until the next flip, one frame away.
	presented presentedSurface

	// Per-batch raster state (nv2a_raster.go) and the per-pusher-run texture decode
	// cache (nv2a_texture.go). Both are transient rebuilds from Regs/RAM — not state.
	rast      rasterState
	rastValid bool
	texCache  map[texKey]*texImage

	// The 2D blit engine's latched state (nv2a_blit.go). Re-programmed before every blit, so
	// like the raster state it is transient and outside the savestate.
	surf2D surfaces2D
	blit   imageBlit

	// --- survey instrumentation (bring-up) ---
	survey    bool
	seen      map[uint32]int    // (class<<16 | method) -> count, across the whole run
	firstArg  map[uint32]uint32 // first argument seen for each such key (a hint at semantics)
	Methods   int               // total methods dispatched
	SetObjs   int               // NV_SET_OBJECT calls
	unhandled map[uint32]int    // methods with no handler yet (non-survey: would halt)

	// --- RR_SHADOW census (nv2a_texture.go): zetaHist buckets every depth write in the
	// run by the zeta surface OFFSET it lands in, so a shadow caster (a pass writing
	// non-far depth into a non-framebuffer buffer) is visible whether or not it binds
	// color==zeta. Used to prove whether a sampled depth-texture (shadow map) is ever
	// populated. shadowDumped one-shots the receiver-state dump at the sample halt.
	zetaHist     map[uint32]*zetaBucket
	shadowDumped bool
	// RR_SHADOWFRAG per-draw comparand statistics (nv2a_texture.go traceShadowFrag).
	shadowFrag *shadowFragStats

	// triScratch is assemble()'s reusable triangle-index buffer (transient, not state).
	triScratch [][3]int

	// clipTris / clipVerts are the near-plane clipper's reusable output buffers
	// (nv2a_clip.go), used only when a draw actually crosses the near plane (transient).
	clipTris  [][3]int
	clipVerts []kelvinVtx
}

type zetaBucket struct {
	pix        int
	zmin, zmax uint32
	draws      int // distinct draws that wrote here
	lastDraw   int
}

func newPgraph(m *Machine) *pgraph {
	g := &pgraph{
		m:         m,
		seen:      map[uint32]int{},
		firstArg:  map[uint32]uint32{},
		unhandled: map[uint32]int{},
		texCache:  map[texKey]*texImage{},
		zetaHist:  map[uint32]*zetaBucket{},
	}
	for i := range g.vtxAttr {
		g.vtxAttr[i] = [4]float32{0, 0, 0, 1} // attribute registers reset to (0,0,0,1)
	}
	return g
}

// SetSurvey turns method-survey recording on (records instead of acting/halting).
func (g *pgraph) SetSurvey(v bool) { g.survey = v }

// pgraphMethod is the PFIFO pusher's entry point: dispatch one decoded method write.
func (m *Machine) pgraphMethod(subchan, method, arg uint32) {
	g := m.pgraph
	g.Methods++

	// The debugger's command hook fires BEFORE the engine acts, so a hook can number
	// the command whose fragments are about to arrive at OnPixel.
	if m.OnNVMethod != nil {
		m.OnNVMethod(m, subchan, method, arg)
	}
	// The command scrubber's countdown trips AFTER this method has run (deferred), so
	// position k shows the frame WITH command k's pixels in it.
	if m.stopAfterArmed {
		defer func() {
			if m.stopAfterMethod--; m.stopAfterMethod <= 0 {
				m.StopRequested = true
			}
		}()
	}

	if method == nvSetObject {
		g.SetObjs++
		g.subObject[subchan&7] = arg
		g.subClass[subchan&7] = m.ramhtClass(arg)
		return
	}

	class := g.subClass[subchan&7]

	// Survey recording is additive: it observes the stream but never diverts it, so a
	// -survey run is behaviourally identical to a plain one (an early-out here once made
	// survey runs skip the modelled side effects, e.g. the semaphore release write).
	if g.survey {
		key := class<<16 | (method & 0xFFFF)
		if g.seen[key] == 0 {
			g.firstArg[key] = arg
		}
		g.seen[key]++
	}

	// Dispatch to the real engine. D3D drives geometry through the 3D Kelvin object, but the
	// bloom post-process feeds its composite through the 2D BLIT engine (nv2a_blit.go) — so
	// that has to run too, or the composite samples a stale buffer (an3-drive's grey wash).
	switch class {
	case classKelvin:
		g.kelvinMethod(method, arg)
		return
	case class2DSurfaces:
		g.surf2DMethod(method, arg)
		return
	case classImageBlit:
		g.blitMethod(method, arg)
		return
	}
	// A method on an unmodelled class. Latch nothing; record it as the frontier.
	g.unhandled[class<<16|(method&0xFFFF)]++
}

// dumpReceiverState prints the full pipeline state of the shadow-receiver draw: every
// texture unit's registers, the shader stage program, the combiner, the alpha test, and
// the texture matrix that produces oT3. RR_SHADOW census — documents what consumes the
// depth sample so the compare semantics can be reasoned about from the register state.
func (g *pgraph) dumpReceiverState() {
	fmt.Printf("RECV shaderStages=%08X combCtl=%08X alphaTest=%d alphaFunc=%03X alphaRef=%02X\n",
		g.Regs[0x1E70>>2], g.Regs[kelvinCombinerControl>>2],
		g.Regs[kelvinAlphaTestEnable>>2], g.Regs[kelvinAlphaFunc>>2], g.Regs[kelvinAlphaRef>>2]&0xFF)
	for u := uint32(0); u < 4; u++ {
		r := u * 0x40
		fmt.Printf("RECV unit%d off=%08X fmt=%08X addr=%08X ctl0=%08X ctl1=%08X filt=%08X rect=%08X enable=%d\n",
			u, g.Regs[(kelvinTexOffset+r)>>2], g.Regs[(kelvinTexFormat+r)>>2], g.Regs[(kelvinTexAddress+r)>>2],
			g.Regs[(kelvinTexControl+r)>>2], g.Regs[(kelvinTexCtl1+r)>>2], g.Regs[(kelvinTexFilter+r)>>2],
			g.Regs[(kelvinTexRect+r)>>2], g.Regs[(kelvinTexControl+r)>>2]>>30&1)
	}
	fmt.Printf("RECV blend=%d src=%04X dst=%04X eq=%04X const=%08X\n",
		g.Regs[kelvinBlendEnable>>2]&1, g.Regs[kelvinBlendSrcFactor>>2]&0xFFFF,
		g.Regs[kelvinBlendDstFactor>>2]&0xFFFF, g.Regs[kelvinBlendEquation>>2]&0xFFFF,
		g.Regs[kelvinBlendColor>>2])
	n := int(g.Regs[kelvinCombinerControl>>2] & 0xFF)
	for i := 0; i < n && i < 8; i++ {
		fmt.Printf("RECV stage%d colorICW=%08X colorOCW=%08X alphaICW=%08X alphaOCW=%08X factor0=%08X factor1=%08X\n", i,
			g.Regs[kelvinCombinerColorICW>>2+uint32(i)], g.Regs[kelvinCombinerColorOCW>>2+uint32(i)],
			g.Regs[kelvinCombinerAlphaICW>>2+uint32(i)], g.Regs[kelvinCombinerAlphaOCW>>2+uint32(i)],
			g.Regs[kelvinCombinerFactor0>>2+uint32(i)], g.Regs[kelvinCombinerFactor1>>2+uint32(i)])
	}
	fmt.Printf("RECV final cw0=%08X cw1=%08X\n", g.Regs[kelvinSpecFogCW0>>2], g.Regs[kelvinSpecFogCW1>>2])
	// The oT3 texture matrix (transform constants c176..c179) and texgen mode.
	for c := 176; c <= 179; c++ {
		v := g.Const[c]
		fmt.Printf("RECV c%d = %08X %08X %08X %08X\n", c, v[0], v[1], v[2], v[3])
	}
	// Every register currently holding a GL compare enum (0x200..0x207): if the
	// silicon has a shadow-compare-function method the game latched, it is one of
	// these — an offset outside the known depth/alpha/stencil funcs is the candidate.
	for i, v := range g.Regs {
		if v >= 0x200 && v <= 0x207 {
			fmt.Printf("RECV cmp-enum reg %04X = %03X\n", i*4, v)
		}
	}
	// The bound shadow map itself, as a PNG (depth high byte grayscale; far = white):
	// the caster footprint the derivation projects back onto the scene.
	if img, _, ok := g.texDecode(3); ok && img != nil && img.depth != nil {
		out := image.NewRGBA(image.Rect(0, 0, img.w, img.h))
		copy(out.Pix, img.pix)
		var buf bytes.Buffer
		if err := png.Encode(&buf, out); err == nil {
			name := fmt.Sprintf("shadow-map-draw%d.png", g.Draws)
			if err := os.WriteFile(name, buf.Bytes(), 0644); err == nil {
				fmt.Printf("RECV shadow map written to %s\n", name)
			}
		}
	}
}

// DumpZetaHist prints the RR_SHADOW census: every zeta surface offset that received a
// depth write in the run, with pixel count, depth range, and how many draws wrote there.
// A caster pass is a bucket with non-far depth (zmax < 0xFFFFFF) at a non-framebuffer
// address; an all-far bucket means only the clear's far value is present.
func (g *pgraph) DumpZetaHist() {
	if !shadowTrace {
		return
	}
	if g.shadowFrag != nil {
		g.shadowFrag.print()
		g.shadowFrag = nil
	}
	offs := make([]uint32, 0, len(g.zetaHist))
	for o := range g.zetaHist {
		offs = append(offs, o)
	}
	sort.Slice(offs, func(i, j int) bool { return offs[i] < offs[j] })
	fmt.Printf("SHADOW zeta-write census: %d distinct zeta offsets\n", len(offs))
	for _, o := range offs {
		b := g.zetaHist[o]
		fmt.Printf("  zeta=%08X pix=%-9d draws=%-6d zmin=%06X zmax=%06X\n",
			o, b.pix, b.draws, b.zmin&0xFFFFFF, b.zmax&0xFFFFFF)
	}
}

// SurveyReport returns the recorded method surface, most-frequent first — the concrete
// statement of what the title's command stream contains.
func (g *pgraph) SurveyReport() []string {
	type row struct {
		key   uint32
		count int
	}
	rows := make([]row, 0, len(g.seen))
	for k, c := range g.seen {
		rows = append(rows, row{k, c})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].key>>16 != rows[j].key>>16 {
			return rows[i].key>>16 < rows[j].key>>16
		}
		return rows[i].key&0xFFFF < rows[j].key&0xFFFF
	})
	out := make([]string, 0, len(rows)+2)
	out = append(out, fmt.Sprintf("PGRAPH survey: %d methods, %d SET_OBJECT, %d distinct (class,method)",
		g.Methods, g.SetObjs, len(rows)))
	for _, r := range rows {
		class, mthd := r.key>>16, r.key&0xFFFF
		out = append(out, fmt.Sprintf("  class %04X method %04X  x%-6d  firstArg=%08X",
			class, mthd, r.count, g.firstArg[r.key]))
	}
	return out
}
