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
	"fmt"
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

	// Dispatch to the real engine (only the 3D class is modelled).
	if class == classKelvin {
		g.kelvinMethod(method, arg)
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
	n := int(g.Regs[kelvinCombinerControl>>2] & 0xFF)
	for i := 0; i < n && i < 8; i++ {
		fmt.Printf("RECV stage%d colorICW=%08X colorOCW=%08X alphaICW=%08X alphaOCW=%08X\n", i,
			g.Regs[kelvinCombinerColorICW>>2+uint32(i)], g.Regs[kelvinCombinerColorOCW>>2+uint32(i)],
			g.Regs[kelvinCombinerAlphaICW>>2+uint32(i)], g.Regs[kelvinCombinerAlphaOCW>>2+uint32(i)])
	}
	fmt.Printf("RECV final cw0=%08X cw1=%08X\n", g.Regs[kelvinSpecFogCW0>>2], g.Regs[kelvinSpecFogCW1>>2])
	// The oT3 texture matrix (transform constants c176..c179) and texgen mode.
	for c := 176; c <= 179; c++ {
		v := g.Const[c]
		fmt.Printf("RECV c%d = %08X %08X %08X %08X\n", c, v[0], v[1], v[2], v[3])
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
