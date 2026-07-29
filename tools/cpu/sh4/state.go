package sh4

// state.go exposes the whole architectural state as a serialisable value, so
// a machine model can snapshot a running core and restore it later.
//
// The delay-slot machinery is part of that state. A core restored mid-delay-
// slot without PendingDelay would resume having forgotten that a branch was
// in flight, and would diverge one instruction later — which is exactly the
// kind of bug a savestate is supposed to be immune to. Everything the
// interpreter reads is therefore carried here, private fields included: both
// floating-point banks, both R0-R7 banks, the store queues, the timers'
// fractional accumulators, and the on-chip raw/gap maps (deep-copied, because
// a shared map is not a snapshot).

// State is a complete SH7091 register-level snapshot.
type State struct {
	R     [16]uint32
	Rbank [8]uint32
	SR    uint32
	GBR, VBR, SSR, SPC, SGR, DBR uint32
	MACH, MACL, PR               uint32
	PC, NextPC                   uint32

	FPR   [2][16]uint32
	FPSCR uint32
	FPUL  uint32

	MMUCR, CCR, TRA, EXPEVT, INTEVT uint32
	PTEH, PTEL, TTB, TEA            uint32
	QACR0, QACR1                    uint32
	ICR, IPRA, IPRB, IPRC           uint32
	TMU                             TMUState
	SQ                              [2][8]uint32

	IRLLevel, IRLCode uint32
	SerialTX          []byte

	Halted     bool
	HaltReason string
	Steps      uint64

	CurPC        uint32
	DelaySlot    bool
	PendingDelay bool

	OnchipRaw  map[uint32]uint32
	OnchipGaps map[uint32]int
}

// Snapshot captures the core's state.
func (c *CPU) Snapshot() State {
	s := State{
		R: c.R, Rbank: c.Rbank, SR: c.SR,
		GBR: c.GBR, VBR: c.VBR, SSR: c.SSR, SPC: c.SPC, SGR: c.SGR, DBR: c.DBR,
		MACH: c.MACH, MACL: c.MACL, PR: c.PR, PC: c.PC, NextPC: c.nextPC,
		FPR: c.fpr, FPSCR: c.FPSCR, FPUL: c.FPUL,
		MMUCR: c.MMUCR, CCR: c.CCR, TRA: c.TRA, EXPEVT: c.EXPEVT, INTEVT: c.INTEVT,
		PTEH: c.PTEH, PTEL: c.PTEL, TTB: c.TTB, TEA: c.TEA,
		QACR0: c.QACR0, QACR1: c.QACR1,
		ICR: c.ICR, IPRA: c.IPRA, IPRB: c.IPRB, IPRC: c.IPRC,
		TMU: c.TMU, SQ: c.SQ,
		IRLLevel: c.irlLevel, IRLCode: c.irlCode,
		SerialTX: append([]byte(nil), c.SerialTX...),
		Halted: c.Halted, HaltReason: c.HaltReason, Steps: c.Steps,
		CurPC: c.curPC, DelaySlot: c.delaySlot, PendingDelay: c.pendingDelay,
		OnchipRaw:  make(map[uint32]uint32, len(c.onchipRaw)),
		OnchipGaps: make(map[uint32]int, len(c.onchipGaps)),
	}
	for k, v := range c.onchipRaw {
		s.OnchipRaw[k] = v
	}
	for k, v := range c.onchipGaps {
		s.OnchipGaps[k] = v
	}
	return s
}

// Restore overwrites the core's state in place, leaving its Bus attached.
func (c *CPU) Restore(s State) {
	c.R, c.Rbank, c.SR = s.R, s.Rbank, s.SR
	c.GBR, c.VBR, c.SSR, c.SPC, c.SGR, c.DBR = s.GBR, s.VBR, s.SSR, s.SPC, s.SGR, s.DBR
	c.MACH, c.MACL, c.PR, c.PC, c.nextPC = s.MACH, s.MACL, s.PR, s.PC, s.NextPC
	c.fpr, c.FPSCR, c.FPUL = s.FPR, s.FPSCR, s.FPUL
	c.MMUCR, c.CCR, c.TRA, c.EXPEVT, c.INTEVT = s.MMUCR, s.CCR, s.TRA, s.EXPEVT, s.INTEVT
	c.PTEH, c.PTEL, c.TTB, c.TEA = s.PTEH, s.PTEL, s.TTB, s.TEA
	c.QACR0, c.QACR1 = s.QACR0, s.QACR1
	c.ICR, c.IPRA, c.IPRB, c.IPRC = s.ICR, s.IPRA, s.IPRB, s.IPRC
	c.TMU, c.SQ = s.TMU, s.SQ
	c.irlLevel, c.irlCode = s.IRLLevel, s.IRLCode
	c.SerialTX = append([]byte(nil), s.SerialTX...)
	c.Halted, c.HaltReason, c.Steps = s.Halted, s.HaltReason, s.Steps
	c.curPC, c.delaySlot, c.pendingDelay = s.CurPC, s.DelaySlot, s.PendingDelay
	c.onchipRaw = make(map[uint32]uint32, len(s.OnchipRaw))
	for k, v := range s.OnchipRaw {
		c.onchipRaw[k] = v
	}
	c.onchipGaps = make(map[uint32]int, len(s.OnchipGaps))
	for k, v := range s.OnchipGaps {
		c.onchipGaps[k] = v
	}
}
