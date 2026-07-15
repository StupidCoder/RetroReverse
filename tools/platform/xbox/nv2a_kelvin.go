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

// kelvinMethod handles one method write to the bound Kelvin (3D) object. During bring-up
// it latches the register and records anything without an explicit effect as unhandled,
// so a run's reach through the 3D command surface stays a concrete statement.
func (g *pgraph) kelvinMethod(method, arg uint32) {
	if method < uint32(len(g.Regs))*4 {
		g.Regs[method>>2] = arg
	}
	// Triggers and side-effecting methods will graduate here from the survey. Until a
	// method is modelled, latching its register is harmless; the survey/unhandled map
	// is what states the frontier.
	g.unhandled[classKelvin<<16|(method&0xFFFF)]++
}
