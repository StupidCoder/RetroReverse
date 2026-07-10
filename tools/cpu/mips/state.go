package mips

// state.go — snapshot/restore for the CPU core and the GTE, so a platform
// oracle can save a machine state mid-run and branch experiments from it
// without re-executing the boot. The state structs are exported (and gob-
// friendly) because the platform package serializes them as part of its own
// machine snapshot.

// CPUState is the complete programmer-visible and pipeline state of the core.
type CPUState struct {
	R, Out       [32]uint32
	HI, LO       uint32
	PC, NextPC   uint32
	COP0         [32]uint32
	Halted       bool
	HaltReason   string
	Steps        uint64
	CurPC        uint32
	LdReg, LdVal uint32
	DelaySlot    bool
	PendingDelay bool
	BranchAddr   uint32
}

// SaveState captures the core's state (the bus and hooks are not part of it).
func (c *CPU) SaveState() CPUState {
	return CPUState{
		R: c.R, Out: c.out, HI: c.HI, LO: c.LO,
		PC: c.PC, NextPC: c.nextPC, COP0: c.COP0,
		Halted: c.Halted, HaltReason: c.HaltReason, Steps: c.Steps,
		CurPC: c.curPC, LdReg: c.ld.reg, LdVal: c.ld.val,
		DelaySlot: c.delaySlot, PendingDelay: c.pendingDelay, BranchAddr: c.branchAddr,
	}
}

// LoadState restores a state captured by SaveState.
func (c *CPU) LoadState(s CPUState) {
	c.R, c.out, c.HI, c.LO = s.R, s.Out, s.HI, s.LO
	c.PC, c.nextPC, c.COP0 = s.PC, s.NextPC, s.COP0
	c.Halted, c.HaltReason, c.Steps = s.Halted, s.HaltReason, s.Steps
	c.curPC, c.ld = s.CurPC, loadSlot{reg: s.LdReg, val: s.LdVal}
	c.delaySlot, c.pendingDelay, c.branchAddr = s.DelaySlot, s.PendingDelay, s.BranchAddr
}

// GTEState is the geometry engine's full register file.
type GTEState struct {
	Data, Ctrl [32]uint32
}

// SaveState captures the GTE registers.
func (g *GTE) SaveState() GTEState { return GTEState{Data: g.data, Ctrl: g.ctrl} }

// LoadState restores the GTE registers.
func (g *GTE) LoadState(s GTEState) { g.data, g.ctrl = s.Data, s.Ctrl }
