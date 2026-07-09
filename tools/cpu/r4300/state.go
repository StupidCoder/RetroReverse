package r4300

// state.go exposes the whole architectural state as a serialisable value, so a
// machine model can snapshot a running core and restore it later.
//
// The delay-slot machinery is part of that state. A core restored mid-delay-slot
// without pendingDelay and branchAddr would resume having forgotten that a
// branch was in flight, and would diverge one instruction later — which is
// exactly the kind of bug a savestate is supposed to be immune to. Everything
// the interpreter reads is therefore carried here, private fields included.

// State is a complete VR4300 register-level snapshot.
type State struct {
	R      [32]uint64
	HI, LO uint64
	PC     uint64
	NextPC uint64
	COP0   [32]uint64
	TLB    [TLBSize]TLBEntry
	FGR    [32]uint64
	FCR31  uint32
	LLBit  bool

	Halted     bool
	HaltReason string
	Steps      uint64

	CurPC        uint64
	DelaySlot    bool
	PendingDelay bool
	BranchAddr   uint64
	CountFrac    uint64
}

// Snapshot captures the core's state.
func (c *CPU) Snapshot() State {
	return State{
		R: c.R, HI: c.HI, LO: c.LO, PC: c.PC, NextPC: c.nextPC,
		COP0: c.COP0, TLB: c.TLB, FGR: c.FGR, FCR31: c.FCR31, LLBit: c.LLBit,
		Halted: c.Halted, HaltReason: c.HaltReason, Steps: c.Steps,
		CurPC: c.curPC, DelaySlot: c.delaySlot, PendingDelay: c.pendingDelay,
		BranchAddr: c.branchAddr, CountFrac: c.countFrac,
	}
}

// Restore overwrites the core's state in place, leaving its Bus attached.
func (c *CPU) Restore(s State) {
	c.R, c.HI, c.LO, c.PC, c.nextPC = s.R, s.HI, s.LO, s.PC, s.NextPC
	c.COP0, c.TLB, c.FGR, c.FCR31, c.LLBit = s.COP0, s.TLB, s.FGR, s.FCR31, s.LLBit
	c.Halted, c.HaltReason, c.Steps = s.Halted, s.HaltReason, s.Steps
	c.curPC, c.delaySlot, c.pendingDelay = s.CurPC, s.DelaySlot, s.PendingDelay
	c.branchAddr, c.countFrac = s.BranchAddr, s.CountFrac
}
