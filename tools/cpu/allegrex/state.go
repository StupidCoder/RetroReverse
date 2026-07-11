package allegrex

// state.go — snapshot/restore for the CPU core and the GTE, so a platform
// oracle can save a machine state mid-run and branch experiments from it
// without re-executing the boot. The state structs are exported (and gob-
// friendly) because the platform package serializes them as part of its own
// machine snapshot.

// CPUState is the complete programmer-visible and pipeline state of the core,
// including the COP1 FPU and COP2 VFPU register files.
type CPUState struct {
	R, Out                       [32]uint32
	HI, LO                       uint32
	PC, NextPC                   uint32
	COP0                         [32]uint32
	F                            [32]uint32
	FCC                          bool
	V                            [128]uint32
	VpfxS, VpfxT, VpfxD          uint32
	VpfxSset, VpfxTset, VpfxDset bool
	Halted                       bool
	HaltReason                   string
	Steps                        uint64
	CurPC                        uint32
	LdReg, LdVal                 uint32
	DelaySlot                    bool
	PendingDelay                 bool
	BranchAddr                   uint32
	NullifyNext                  bool
}

// SaveState captures the core's state (the bus and hooks are not part of it).
func (c *CPU) SaveState() CPUState {
	return CPUState{
		R: c.R, Out: c.out, HI: c.HI, LO: c.LO,
		PC: c.PC, NextPC: c.nextPC, COP0: c.COP0,
		F: c.F, FCC: c.FCC, V: c.V,
		VpfxS: c.vpfxS, VpfxT: c.vpfxT, VpfxD: c.vpfxD,
		VpfxSset: c.vpfxSset, VpfxTset: c.vpfxTset, VpfxDset: c.vpfxDset,
		Halted: c.Halted, HaltReason: c.HaltReason, Steps: c.Steps,
		CurPC: c.curPC, LdReg: c.ld.reg, LdVal: c.ld.val,
		DelaySlot: c.delaySlot, PendingDelay: c.pendingDelay, BranchAddr: c.branchAddr,
		NullifyNext: c.nullifyNext,
	}
}

// LoadState restores a state captured by SaveState.
func (c *CPU) LoadState(s CPUState) {
	c.R, c.out, c.HI, c.LO = s.R, s.Out, s.HI, s.LO
	c.PC, c.nextPC, c.COP0 = s.PC, s.NextPC, s.COP0
	c.F, c.FCC, c.V = s.F, s.FCC, s.V
	c.vpfxS, c.vpfxT, c.vpfxD = s.VpfxS, s.VpfxT, s.VpfxD
	c.vpfxSset, c.vpfxTset, c.vpfxDset = s.VpfxSset, s.VpfxTset, s.VpfxDset
	c.Halted, c.HaltReason, c.Steps = s.Halted, s.HaltReason, s.Steps
	c.curPC, c.ld = s.CurPC, loadSlot{reg: s.LdReg, val: s.LdVal}
	c.delaySlot, c.pendingDelay, c.branchAddr = s.DelaySlot, s.PendingDelay, s.BranchAddr
	c.nullifyNext = s.NullifyNext
}
