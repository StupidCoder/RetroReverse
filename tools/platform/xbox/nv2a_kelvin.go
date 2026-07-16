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

// kelvinMethod handles one method write to the bound Kelvin (3D) object. During bring-up
// it latches the register and records anything without an explicit effect as unhandled,
// so a run's reach through the 3D command surface stays a concrete statement.
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
	switch method {
	case kelvinClearSurface:
		// CLEAR_SURFACE: fill the clear rect of the color surface with the latched
		// clear color (nv2a_frame.go) — the first Kelvin method that produces pixels.
		g.clearSurface(arg)
	case kelvinSemaphoreRelease:
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
	}
	// Other triggers and side-effecting methods graduate here from the survey. Until a
	// method is modelled, latching its register is harmless; the survey/unhandled map
	// is what states the frontier.
	g.unhandled[classKelvin<<16|(method&0xFFFF)]++
}
