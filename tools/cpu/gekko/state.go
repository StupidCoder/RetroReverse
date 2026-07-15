package gekko

// state.go snapshots the processor.
//
// Everything architectural is here, and one thing that would not be architectural on
// another machine: the locked cache. It is 16 KiB of the L1 data cache that software has
// taken out of the coherency protocol and is using as a scratchpad, and its contents are
// not a copy of anything in main memory. Dropping it from a snapshot would lose the
// program's working set, and the restored machine would run on stale data — a bug that
// would appear as a game that resumes and then quietly does the wrong thing.
//
// The pacing state — the clock fraction, the decrementer's armed edge — is here too, and
// for the same reason: a savestate must be resumable into a run that is bit-identical to
// the one that never stopped, and a timer that restarted mid-tick would not be.

// State is the whole processor.
type State struct {
	GPR [32]uint32
	FPR [32]FPR

	PC, LR, CTR uint32
	CR, XER     uint32
	MSR, FPSCR  uint32

	GQR                    [8]uint32
	HID0, HID1, HID2, HID4 uint32
	WPAR, DMAU, DMAL, L2CR uint32
	SRR0, SRR1             uint32
	SPRG                   [4]uint32
	DSISR, DAR, SDR1, PVR  uint32
	SR                     [16]uint32
	IBAT, DBAT             [4][2]uint32

	TB        uint64
	DEC       uint32
	ClockFrac uint32
	DecArmed  bool

	Reserved    bool
	ReserveAddr uint32

	LCData    []byte
	LCBase    uint32
	LCEnabled bool

	ExtInt bool
	Steps  uint64

	Halted     bool
	HaltReason string
}

// Snapshot copies the processor out.
func (c *CPU) Snapshot() State {
	s := State{
		GPR: c.GPR, FPR: c.FPR,
		PC: c.PC, LR: c.LR, CTR: c.CTR,
		CR: c.CR, XER: c.XER, MSR: c.MSR, FPSCR: c.FPSCR,
		GQR: c.GQR, HID0: c.HID0, HID1: c.HID1, HID2: c.HID2, HID4: c.HID4,
		WPAR: c.WPAR, DMAU: c.DMAU, DMAL: c.DMAL, L2CR: c.L2CR,
		SRR0: c.SRR0, SRR1: c.SRR1, SPRG: c.SPRG,
		DSISR: c.DSISR, DAR: c.DAR, SDR1: c.SDR1, PVR: c.PVR,
		SR: c.SR, IBAT: c.IBAT, DBAT: c.DBAT,
		TB: c.TB, DEC: c.DEC, ClockFrac: c.clockFrac, DecArmed: c.decArmed,
		Reserved: c.Reserved, ReserveAddr: c.ReserveAddr,
		LCBase: c.LC.Base, LCEnabled: c.LC.Enabled_,
		ExtInt: c.ExtInt, Steps: c.Steps,
		Halted: c.Halted, HaltReason: c.HaltReason,
	}
	s.LCData = make([]byte, LockedCacheSize)
	copy(s.LCData, c.LC.Data[:])
	return s
}

// Restore puts one back.
//
// The halt is deliberately not restored as a halt: a state captured at a stop must not
// re-stop the moment it is loaded, or the snapshot would be useless for the one thing it
// is most needed for — resuming just before the thing that went wrong and looking at it.
func (c *CPU) Restore(s State) {
	c.GPR, c.FPR = s.GPR, s.FPR
	c.PC, c.LR, c.CTR = s.PC, s.LR, s.CTR
	c.CR, c.XER, c.MSR, c.FPSCR = s.CR, s.XER, s.MSR, s.FPSCR
	c.GQR, c.HID0, c.HID1, c.HID2, c.HID4 = s.GQR, s.HID0, s.HID1, s.HID2, s.HID4
	c.WPAR, c.DMAU, c.DMAL, c.L2CR = s.WPAR, s.DMAU, s.DMAL, s.L2CR
	c.SRR0, c.SRR1, c.SPRG = s.SRR0, s.SRR1, s.SPRG
	c.DSISR, c.DAR, c.SDR1, c.PVR = s.DSISR, s.DAR, s.SDR1, s.PVR
	c.SR, c.IBAT, c.DBAT = s.SR, s.IBAT, s.DBAT
	c.TB, c.DEC, c.clockFrac, c.decArmed = s.TB, s.DEC, s.ClockFrac, s.DecArmed
	c.Reserved, c.ReserveAddr = s.Reserved, s.ReserveAddr
	c.LC.Base, c.LC.Enabled_ = s.LCBase, s.LCEnabled
	copy(c.LC.Data[:], s.LCData)
	c.ExtInt, c.Steps = s.ExtInt, s.Steps
	// Restore the halted flag faithfully: a snapshot taken after the core stopped must resume
	// stopped, or a run continued from it diverges from the one that never stopped (which is
	// exactly what the machine's savestate round-trip test enforces). Snapshot saves these;
	// clearing them here — the old behaviour — silently let a restored machine run past a halt.
	c.Halted, c.HaltReason = s.Halted, s.HaltReason
}
