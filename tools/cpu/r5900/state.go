package r5900

// state.go exposes the whole architectural state as a serialisable value, so a
// machine model can snapshot a running core and restore it later.
//
// The delay-slot machinery is part of that state. A core restored mid-delay-slot
// without pendingDelay and branchAddr would resume having forgotten that a branch
// was in flight, and would diverge one instruction later — which is exactly the
// kind of bug a savestate is supposed to be immune to. Everything the interpreter
// reads is therefore carried here, private fields included.
//
// What is *not* carried: the Bus, the Syscall hook and the COP2 unit. Those are
// host wiring, rebuilt by the machine model on load, not architectural state.

// State is a complete R5900 register-level snapshot.
type State struct {
	R        [32]Quad
	HI, LO   uint64
	HI1, LO1 uint64
	SA       uint32
	PC       uint64
	NextPC   uint64
	COP0     [32]uint64
	TLB      [TLBSize]TLBEntry
	FPR      [32]uint32
	ACC      uint32
	FCR31    uint32
	LLBit    bool

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
		R: c.R, HI: c.HI, LO: c.LO, HI1: c.HI1, LO1: c.LO1, SA: c.SA,
		PC: c.PC, NextPC: c.nextPC,
		COP0: c.COP0, TLB: c.TLB,
		FPR: c.FPR, ACC: c.ACC, FCR31: c.FCR31, LLBit: c.LLBit,
		Halted: c.Halted, HaltReason: c.HaltReason, Steps: c.Steps,
		CurPC: c.curPC, DelaySlot: c.delaySlot, PendingDelay: c.pendingDelay,
		BranchAddr: c.branchAddr, CountFrac: c.countFrac,
	}
}

// Restore overwrites the core's state in place, leaving its Bus attached.
func (c *CPU) Restore(s State) {
	c.R, c.HI, c.LO, c.HI1, c.LO1, c.SA = s.R, s.HI, s.LO, s.HI1, s.LO1, s.SA
	c.PC, c.nextPC = s.PC, s.NextPC
	c.COP0, c.TLB = s.COP0, s.TLB
	c.FPR, c.ACC, c.FCR31, c.LLBit = s.FPR, s.ACC, s.FCR31, s.LLBit
	c.Halted, c.HaltReason, c.Steps = s.Halted, s.HaltReason, s.Steps
	c.curPC, c.delaySlot, c.pendingDelay = s.CurPC, s.DelaySlot, s.PendingDelay
	c.branchAddr, c.countFrac = s.BranchAddr, s.CountFrac
}
