package rsp

// state.go exposes the RSP's architectural state so a machine model can snapshot
// it. The vector registers and the accumulator are the bulk of it; the reciprocal
// unit's staging registers matter too, because a 32-bit divide spans three
// instructions and a snapshot taken between them would otherwise lose its input.

// State is a complete RSP register-level snapshot. The two memories are not
// carried here: they belong to the machine, which snapshots them itself.
type State struct {
	R   [32]uint32
	V   [32][8]uint16
	Acc [8]uint64
	VCO uint16
	VCC uint16
	VCE uint8

	DivIn       uint16
	DivOut      uint16
	DivInLoaded bool

	PC     uint32
	NextPC uint32
	CurPC  uint32

	Halted     bool
	Broke      bool
	HaltReason string
	Steps      uint64
}

// Snapshot captures the core's state.
func (c *CPU) Snapshot() State {
	return State{
		R: c.R, V: c.V, Acc: c.Acc, VCO: c.VCO, VCC: c.VCC, VCE: c.VCE,
		DivIn: c.divIn, DivOut: c.divOut, DivInLoaded: c.divInLoaded,
		PC: c.PC, NextPC: c.nextPC, CurPC: c.curPC,
		Halted: c.Halted, Broke: c.Broke, HaltReason: c.HaltReason, Steps: c.Steps,
	}
}

// Restore overwrites the core's state in place, leaving its memories attached.
func (c *CPU) Restore(s State) {
	c.R, c.V, c.Acc = s.R, s.V, s.Acc
	c.VCO, c.VCC, c.VCE = s.VCO, s.VCC, s.VCE
	c.divIn, c.divOut, c.divInLoaded = s.DivIn, s.DivOut, s.DivInLoaded
	c.PC, c.nextPC, c.curPC = s.PC, s.NextPC, s.CurPC
	c.Halted, c.Broke, c.HaltReason, c.Steps = s.Halted, s.Broke, s.HaltReason, s.Steps
}
