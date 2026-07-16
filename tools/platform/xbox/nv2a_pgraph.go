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

	// --- survey instrumentation (bring-up) ---
	survey    bool
	seen      map[uint32]int    // (class<<16 | method) -> count, across the whole run
	firstArg  map[uint32]uint32 // first argument seen for each such key (a hint at semantics)
	Methods   int               // total methods dispatched
	SetObjs   int               // NV_SET_OBJECT calls
	unhandled map[uint32]int    // methods with no handler yet (non-survey: would halt)
}

func newPgraph(m *Machine) *pgraph {
	return &pgraph{
		m:         m,
		seen:      map[uint32]int{},
		firstArg:  map[uint32]uint32{},
		unhandled: map[uint32]int{},
	}
}

// SetSurvey turns method-survey recording on (records instead of acting/halting).
func (g *pgraph) SetSurvey(v bool) { g.survey = v }

// pgraphMethod is the PFIFO pusher's entry point: dispatch one decoded method write.
func (m *Machine) pgraphMethod(subchan, method, arg uint32) {
	g := m.pgraph
	g.Methods++

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
