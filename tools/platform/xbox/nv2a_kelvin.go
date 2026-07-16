package xbox

// nv2a_kelvin.go implements the NV20_KELVIN_PRIMITIVE object (class 0x0097) — the Xbox's
// 3D graphics engine. It is the destination for the method stream the PFIFO pusher
// decodes on the subchannel the Direct3D runtime binds Kelvin to. Most methods latch
// state into the object register file (pgraph.Regs, indexed by method>>2); a few are
// triggers that run the pipeline (nv2a_pgraph.go dispatches those). The method numbers
// are the NV2A hardware's.
//
// This starts as the survey's graduation target: methods move here from the survey's
// "seen" set one at a time, each pinned from the live stream, so an unmodelled method
// halts and names itself instead of the engine drawing something plausible.

import (
	"fmt"
	"os"
)

var nvSemTrace = os.Getenv("RR_NV_SEM") != ""

// Kelvin methods with modelled side effects (NV2A method numbers).
const (
	kelvinCtxDmaSemaphore  = 0x01A4 // SET_CONTEXT_DMA_SEMAPHORE: DMA handle for the semaphore surface
	kelvinSemaphoreOffset  = 0x1D6C // SET_SEMAPHORE_OFFSET: byte offset within that surface
	kelvinSemaphoreRelease = 0x1D70 // BACK_END_WRITE_SEMAPHORE_RELEASE: write the value there
)

// kelvinMethod handles one method write to the bound Kelvin (3D) object. Every method
// latches into the register file; the cases below are the ones with modelled side
// effects (triggers and FIFO-shaped data ports). A method with no case is a plain
// state latch the pipeline reads back at draw time — or an unmodelled one, which the
// unhandled map records so a run's reach through the 3D command surface stays a
// concrete statement.
func (g *pgraph) kelvinMethod(method, arg uint32) {
	if method < uint32(len(g.Regs))*4 {
		g.Regs[method>>2] = arg
	}
	if nvSemTrace {
		switch method {
		case 0x0100, 0x0110, 0x0130, 0x1D6C, 0x1D70, 0x17D0:
			fmt.Printf("KELVIN sync: method %04X arg %08X (steps=%d)\n", method, arg, g.m.CPU.Steps)
		}
	}
	switch {
	case method == kelvinClearSurface:
		// CLEAR_SURFACE: fill the clear rect of the color/zeta surfaces with the
		// latched clear values (nv2a_frame.go) — the first Kelvin method that
		// produced pixels.
		g.clearSurface(arg)
		return
	case method == kelvinSemaphoreRelease:
		// BACK_END_WRITE_SEMAPHORE_RELEASE: the back end writes the release value into
		// the bound semaphore surface at the latched offset — this is how the Direct3D
		// runtime observes GPU progress (its per-frame sync values arrive here, odd and
		// ascending by 2). The D3D busy-wait at 0x1AE550 additionally polls PGRAPH
		// register 0x400B10, comparing its bits 2..6 against the semaphore value <<2 —
		// the register mirrors the back end's semaphore progress. Our pipeline is
		// synchronous, so both update the instant the method executes: the release has
		// retired by the time the CPU can look.
		base, limit := g.m.dmaObjectTarget(g.Regs[kelvinCtxDmaSemaphore>>2])
		off := g.Regs[kelvinSemaphoreOffset>>2]
		if base != 0 && off <= limit {
			g.m.write32(base+off, arg)
		}
		g.m.nv.reg[nvPGRAPH_SEMAPHORE>>2] = arg << 2
		return

	// --- the vertex front end (nv2a_vertex.go) ---
	case method == kelvinBeginEnd:
		g.rastValid = false // the batch's raster state is decoded fresh per BEGIN/END
		g.beginEnd(arg)
		return
	case method == kelvinInlineArray:
		g.inline = append(g.inline, arg)
		return
	case method == kelvinElement16:
		g.elems = append(g.elems, arg&0xFFFF, arg>>16)
		return
	case method == kelvinElement32:
		g.elems = append(g.elems, arg)
		return
	case method == kelvinDrawArrays:
		g.ranges = append(g.ranges, [2]uint32{arg & 0xFFFFFF, (arg >> 24) + 1})
		return
	case method >= kelvinVertexData4C && method < kelvinVertexData4C+0x40:
		// SET_VERTEX_DATA4UB: a persistent attribute value as 4 unsigned bytes
		// (RGBA byte order), read by every vertex whose arrays do not supply it.
		i := (method - kelvinVertexData4C) >> 2
		g.vtxAttr[i] = [4]float32{
			float32(arg&0xFF) / 255, float32(arg>>8&0xFF) / 255,
			float32(arg>>16&0xFF) / 255, float32(arg>>24&0xFF) / 255,
		}
		return

	// --- the transform program (nv2a_vsh.go) ---
	case method >= kelvinProgData && method < kelvinProgData+0x80:
		g.progData(arg)
		return
	case method >= kelvinConstData && method < kelvinConstData+0x80:
		g.constData(arg)
		return
	case method == kelvinProgLoad:
		g.ProgLoad = arg
		g.progBufN = 0
		return
	case method == kelvinConstLoad:
		g.ConstLoad = arg
		g.constBufN = 0
		return
	}
	// Everything else is a plain state latch (the raster/texture/combiner registers
	// the pipeline reads at draw time) or an unmodelled method; the survey/unhandled
	// map is what states the frontier.
	g.unhandled[classKelvin<<16|(method&0xFFFF)]++
}
