package ps2

// state.go snapshots the whole machine, so a boot that takes minutes to reach an
// interesting place can be replayed in seconds.
//
// This is not a convenience. Every platform in this repository that gained savestates
// late paid for it in wasted hours, and the rule since is that they land in the first
// phase of a new machine, not the last. A cold boot to a menu is billions of
// instructions; the work that follows it is a hundred small experiments, and each of
// them has to start somewhere.
//
// What is serialised is architectural state: memory, the CPU, the threads, the kernel's
// bookkeeping. What is not is host wiring — the hooks, the mounted disc, the loaded
// executable. Those the caller sets up again on load, and they are re-attached rather
// than restored. The image hash is pinned, so a state cannot be loaded against a
// different disc and produce nonsense.

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/r5900"
	"retroreverse.com/tools/cpu/vu"
)

// ThreadState is one thread, in a form gob can carry.
type ThreadState struct {
	ID          uint32
	Entry       uint32
	Stack       uint32
	StackSz     uint32
	GP          uint32
	Priority    uint32
	State       int
	WakeupCount int32
	Ctx         r5900.State
}

// SemaState is one semaphore.
type SemaState struct {
	ID       uint32
	Count    int32
	MaxCount int32
	Waiting  []uint32
}

// HandlerState is one registered interrupt handler.
type HandlerState struct {
	Cause uint32
	Addr  uint32
	Arg   uint32
}

// EETimerState is one EE timer (eetimer.go).
type EETimerState struct {
	Base      uint32
	BaseSteps uint64
	Mode      uint32
	Comp      uint32
	Hold      uint32
}

// MachineState is a complete snapshot.
type MachineState struct {
	ImageHash string

	RAM    []byte
	SPRAM  []byte
	IOPRAM []byte

	CPU r5900.State
	IO  map[uint32]uint32

	Threads       []ThreadState
	NextThreadID  uint32
	CurrentThread uint32

	IntcHandlers []HandlerState
	DmacHandlers []HandlerState
	IntcMask     uint32
	IntcStat     uint32
	DmacMask     uint32
	DmacStat     uint32

	VSyncFlagPtr  uint32
	VSyncFlag2Ptr uint32

	GsInterlace, GsVideoMode, GsFieldMode, GsIMR uint32

	SifRegs map[uint32]uint32

	// The six registers both processors can see. They were not in the snapshot until they
	// held anything worth losing: SMCOM carries the address the IOP publishes its command
	// buffer at, and SMFLG the three bits that say how far the second processor has got.
	// A snapshot without them resumes into a machine whose two halves have never met.
	Sbus          [sbusRegs]uint32
	SifDmaID      uint32
	SifCmdBuf     uint32
	SifCmdHandler uint32
	Steps         uint64

	// The EE timers: the frame-ratio clock the GOAL engine paces its day by.
	EETimers [4]EETimerState

	// The syscalls the game replaced with its own routines. Losing these on load is
	// not a visible failure: the calls still "work", they just quietly do nothing,
	// because the kernel's stub answers instead of the game's code.
	UserSyscalls map[uint32]uint32

	Semas      []SemaState
	NextSemaID uint32

	HeapPtr, HeapEnd uint32
	VBlanks          uint32

	// The live pad: what a caller is holding down right now. PadScript is not here —
	// it is a pure function of VBlanks and replays itself — but this is not derivable
	// from anything, so a resume without it would drop a held button. The sticks are
	// deflections from centre, so a state written before the pad grew sticks restores
	// a resting stick rather than a hard up-left diagonal.
	PadLiveButtons       uint16
	PadLiveLX, PadLiveLY int8
	PadLiveRX, PadLiveRY int8
	TTY                  []byte
	SyscallCalls         map[string]int

	// The six registers both processors can see. They are shared silicon, and a snapshot
	// that dropped them would restore a machine in which one side of a handshake had
	// happened and the other had forgotten.
	SBus [sbusRegs]uint32

	// The second processor, entire (iopstate.go). Nil when the IOP was never started —
	// which is a state a PS2 really is in, right up until the game reboots it.
	IOP *IOPState

	// The render path, entire. None of this was in the snapshot until the machine could
	// draw; a resume without it gets a blank GS, empty VU memories and a fresh DMA
	// controller — fine for a CPU probe, and exactly wrong for the frame work, where a
	// state without its VRAM "has no displayable frame" and a VU without its microcode
	// dies on its first MSCAL.
	DmacChans                        [dmacChannels]DmacChanState
	DCtrl, DPcr, DSqwc, DRbsr, DRbor uint32
	GS                               *GSState
	VIF0, VIF1                       *VIFState
}

// DmacChanState is one EE DMA channel's register file, exported for gob.
type DmacChanState struct {
	CHCR, MADR, QWC, TADR, ASR0, ASR1, SADR uint32
}

// GSXferState mirrors gsXfer: the image upload in progress.
type GSXferState struct {
	Active                               bool
	DBP, DBW, DPSM, DSAX, DSAY, RRW, RRH uint32
	X, Y                                 uint32
	Partial                              []byte
}

// GSVertexState mirrors gsVertex: one entry of the vertex queue.
type GSVertexState struct {
	PX, PY  int32
	Z       uint32
	RGBA    uint32
	U, V    int32
	S, T, Q float32
}

// GSState is the Graphics Synthesizer, entire: its 4 MiB, its register file, the upload
// cursor, the vertex queue and the on-chip CLUT. The census counters ride along so a
// resumed run's report still says what the whole boot did, not what happened since the
// load. (The per-pixel outcome tallies are diagnostic prints, not machine state, and are
// deliberately left behind.)
type GSState struct {
	VRAM       []byte
	Reg        [0x80]uint64
	CSR        uint64
	Xfer       GSXferState
	VQ         [3]GSVertexState
	VQN        int
	Q          uint32
	CLUT       [512]uint32
	CBP0, CBP1 uint32
	Uploads    int
	Prims      int
	PrimCount  [8]int
	DrawCensus map[string]int
	// PATH2's cross-DIRECT stream state (gif.go's gifStream).
	Path2ImageRemain int
	Path2SkipRemain  int
	Path2Carry       []byte
}

// VIFState is one VPU interface and its vector unit: the sticky decode registers, the
// command mid-parse, and the program and data memories the VIF filled — the memories are
// here rather than in the VU's own snapshot because the VIF owns the slices and hands
// them to vu.New. VU0's state doubles as the EE's COP2 state, which is where the GOAL
// engine's matrix stack lives between VCALLMS programs.
type VIFState struct {
	CL, WL, Mode, Mask           uint32
	Row, Col                     [4]uint32
	Base, Ofst, ITop, Mark, Tops uint32
	Cmd                          uint32
	Pending                      int
	Buf                          []byte
	Micro, Data                  []byte
	VU                           *vu.State
	VUSteps                      uint64
	Census                       map[string]int
	MSCALs                       map[uint32]int
}

// SaveState captures the machine.
func (m *Machine) SaveState() MachineState {
	m.drainVIF1() // never snapshot a kick mid-flight
	s := MachineState{
		ImageHash:     m.imageHash,
		RAM:           append([]byte(nil), m.ram...),
		SPRAM:         append([]byte(nil), m.spram...),
		IOPRAM:        append([]byte(nil), m.iopRAM...),
		CPU:           m.CPU.Snapshot(),
		IO:            map[uint32]uint32{},
		NextThreadID:  m.nextThreadID,
		CurrentThread: m.currentThread,
		IntcMask:      m.intcMask,
		IntcStat:      m.intcStat,
		DmacMask:      m.dmacMask,
		DmacStat:      m.dmacStat,
		VSyncFlagPtr:  m.vsyncFlagPtr,
		VSyncFlag2Ptr: m.vsyncFlag2Ptr,
		GsInterlace:   m.gsInterlace,
		GsVideoMode:   m.gsVideoMode,
		GsFieldMode:   m.gsFieldMode,
		GsIMR:         m.gsIMR,
		SifRegs:       m.sifRegs,
		Sbus:          m.sbus,
		SifDmaID:      m.sifDmaID,
		SifCmdBuf:     m.sifCmdBuf,
		SifCmdHandler: m.sifCmdHandler,
		Steps:         m.steps,
		EETimers: [4]EETimerState{
			{m.eeTimers[0].base, m.eeTimers[0].baseSteps, m.eeTimers[0].mode, m.eeTimers[0].comp, m.eeTimers[0].hold},
			{m.eeTimers[1].base, m.eeTimers[1].baseSteps, m.eeTimers[1].mode, m.eeTimers[1].comp, m.eeTimers[1].hold},
			{m.eeTimers[2].base, m.eeTimers[2].baseSteps, m.eeTimers[2].mode, m.eeTimers[2].comp, m.eeTimers[2].hold},
			{m.eeTimers[3].base, m.eeTimers[3].baseSteps, m.eeTimers[3].mode, m.eeTimers[3].comp, m.eeTimers[3].hold},
		},
		UserSyscalls:   map[uint32]uint32{},
		NextSemaID:     m.nextSemaID,
		HeapPtr:        m.heapPtr,
		HeapEnd:        m.heapEnd,
		VBlanks:        m.vblanks,
		PadLiveButtons: m.padLive.buttons,
		PadLiveLX:      m.padLive.lx,
		PadLiveLY:      m.padLive.ly,
		PadLiveRX:      m.padLive.rx,
		PadLiveRY:      m.padLive.ry,
		TTY:            append([]byte(nil), m.tty...),
		SyscallCalls:   map[string]int{},
		SBus:           m.sbus,
	}
	if m.IOP != nil {
		iop := m.IOP.SaveState()
		s.IOP = &iop
	}
	for i := range m.dmac {
		c := &m.dmac[i]
		s.DmacChans[i] = DmacChanState{
			CHCR: c.chcr, MADR: c.madr, QWC: c.qwc, TADR: c.tadr,
			ASR0: c.asr0, ASR1: c.asr1, SADR: c.sadr,
		}
	}
	s.DCtrl, s.DPcr, s.DSqwc, s.DRbsr, s.DRbor = m.dCtrl, m.dPcr, m.dSqwc, m.dRbsr, m.dRbor
	if gs := m.gs; gs != nil {
		g := &GSState{
			VRAM: append([]byte(nil), gs.vram...),
			Reg:  gs.reg, CSR: gs.csr,
			Xfer: GSXferState{
				Active: gs.xfer.active, DBP: gs.xfer.dbp, DBW: gs.xfer.dbw, DPSM: gs.xfer.dpsm,
				DSAX: gs.xfer.dsax, DSAY: gs.xfer.dsay, RRW: gs.xfer.rrw, RRH: gs.xfer.rrh,
				X: gs.xfer.x, Y: gs.xfer.y, Partial: append([]byte(nil), gs.xfer.partial...),
			},
			VQN: gs.vqN, Q: gs.q, CLUT: gs.clut, CBP0: gs.cbp0, CBP1: gs.cbp1,
			Uploads: gs.uploads, Prims: gs.prims, PrimCount: gs.primCount,
			Path2ImageRemain: gs.path2ImageRemain, Path2SkipRemain: gs.path2SkipRemain,
			Path2Carry: append([]byte(nil), gs.path2Carry...),
			DrawCensus: map[string]int{},
		}
		for i, v := range gs.vq {
			g.VQ[i] = GSVertexState{PX: v.x, PY: v.y, Z: v.z, RGBA: v.rgba, U: v.u, V: v.v, S: v.s, T: v.t, Q: v.q}
		}
		for k, v := range gs.drawCensus {
			g.DrawCensus[k] = v
		}
		s.GS = g
	}
	for i, v := range m.vifs {
		if v == nil {
			continue
		}
		vs := &VIFState{
			CL: v.cl, WL: v.wl, Mode: v.mode, Mask: v.mask,
			Row: v.row, Col: v.col,
			Base: v.base, Ofst: v.ofst, ITop: v.itop, Mark: v.mark, Tops: v.tops,
			Cmd: v.cmd, Pending: v.pending,
			Buf:     append([]byte(nil), v.buf...),
			Micro:   append([]byte(nil), v.micro...),
			Data:    append([]byte(nil), v.data...),
			VUSteps: v.vuSteps,
			Census:  map[string]int{}, MSCALs: map[uint32]int{},
		}
		if v.vu != nil {
			st := v.vu.Snapshot()
			vs.VU = &st
		}
		for k, n := range v.census {
			vs.Census[k] = n
		}
		for k, n := range v.mscal {
			vs.MSCALs[k] = n
		}
		// Two named fields rather than an array: gob refuses a nil element inside an
		// array, and a machine that never touched VIF0 is a state worth saving.
		if i == 0 {
			s.VIF0 = vs
		} else {
			s.VIF1 = vs
		}
	}
	for k, v := range m.io {
		s.IO[k] = v
	}
	for k, v := range m.SyscallCalls {
		s.SyscallCalls[k] = v
	}
	for k, v := range m.userSyscalls {
		s.UserSyscalls[k] = v
	}
	for _, sm := range m.semas {
		s.Semas = append(s.Semas, SemaState{
			ID: sm.id, Count: sm.count, MaxCount: sm.maxCount,
			Waiting: append([]uint32(nil), sm.waiting...),
		})
	}
	// The running thread's context lives in the CPU, not in its saved slot, so it is
	// taken from there. A state that saved the stale slot would resume the thread at
	// wherever it last yielded — which is a bug that only shows up after the first
	// context switch, long after the savestate looked fine.
	for _, t := range m.threads {
		ts := ThreadState{
			ID: t.id, Entry: t.entry, Stack: t.stack, StackSz: t.stackSz,
			GP: t.gp, Priority: t.priority, State: int(t.state),
			WakeupCount: t.wakeupCount, Ctx: t.ctx,
		}
		if t.id == m.currentThread {
			ts.Ctx = s.CPU
		}
		s.Threads = append(s.Threads, ts)
	}
	for _, h := range m.intcHandlers {
		s.IntcHandlers = append(s.IntcHandlers, HandlerState{h.cause, h.addr, h.arg})
	}
	for _, h := range m.dmacHandlers {
		s.DmacHandlers = append(s.DmacHandlers, HandlerState{h.cause, h.addr, h.arg})
	}
	return s
}

// LoadState restores a snapshot in place, leaving the machine's host wiring — its
// hooks, its disc, its executable — as the caller set it up.
func (m *Machine) LoadState(s MachineState) error {
	if m.imageHash != "" && s.ImageHash != "" && m.imageHash != s.ImageHash {
		return fmt.Errorf("ps2: savestate was taken from a different disc (%s, not %s)",
			s.ImageHash, m.imageHash)
	}
	copy(m.ram, s.RAM)
	copy(m.spram, s.SPRAM)
	copy(m.iopRAM, s.IOPRAM)
	m.CPU.Restore(s.CPU)

	m.io = map[uint32]uint32{}
	for k, v := range s.IO {
		m.io[k] = v
	}

	m.threads = map[uint32]*thread{}
	for _, ts := range s.Threads {
		m.threads[ts.ID] = &thread{
			id: ts.ID, entry: ts.Entry, stack: ts.Stack, stackSz: ts.StackSz,
			gp: ts.GP, priority: ts.Priority, state: threadState(ts.State),
			wakeupCount: ts.WakeupCount, ctx: ts.Ctx,
		}
	}
	m.nextThreadID = s.NextThreadID
	m.currentThread = s.CurrentThread

	m.intcHandlers = nil
	for _, h := range s.IntcHandlers {
		m.intcHandlers = append(m.intcHandlers, handler{cause: h.Cause, addr: h.Addr, arg: h.Arg})
	}
	m.dmacHandlers = nil
	for _, h := range s.DmacHandlers {
		m.dmacHandlers = append(m.dmacHandlers, handler{cause: h.Cause, addr: h.Addr, arg: h.Arg})
	}

	m.intcMask, m.intcStat = s.IntcMask, s.IntcStat
	m.dmacMask, m.dmacStat = s.DmacMask, s.DmacStat
	m.vsyncFlagPtr, m.vsyncFlag2Ptr = s.VSyncFlagPtr, s.VSyncFlag2Ptr
	m.gsInterlace, m.gsVideoMode, m.gsFieldMode = s.GsInterlace, s.GsVideoMode, s.GsFieldMode
	m.gsIMR = s.GsIMR
	m.sifDmaID = s.SifDmaID
	m.sbus = s.Sbus
	m.sifRegs = map[uint32]uint32{}
	for k, v := range s.SifRegs {
		m.sifRegs[k] = v
	}
	m.sifCmdBuf, m.sifCmdHandler = s.SifCmdBuf, s.SifCmdHandler
	m.steps = s.Steps
	// The idle detector (idle.go) is bookkeeping over consecutive instructions, not
	// machine state; a resume starts its search afresh rather than inheriting a snapshot
	// taken against a step count and a store count from before the load.
	m.idleDet = idleState{}
	m.stores = 0
	for i, ts := range s.EETimers {
		m.eeTimers[i] = eeTimer{base: ts.Base, baseSteps: ts.BaseSteps, mode: ts.Mode, comp: ts.Comp, hold: ts.Hold}
	}

	m.userSyscalls = map[uint32]uint32{}
	for k, v := range s.UserSyscalls {
		m.userSyscalls[k] = v
	}
	m.semas = map[uint32]*sema{}
	for _, sm := range s.Semas {
		m.semas[sm.ID] = &sema{
			id: sm.ID, count: sm.Count, maxCount: sm.MaxCount,
			waiting: append([]uint32(nil), sm.Waiting...),
		}
	}
	m.nextSemaID = s.NextSemaID
	if m.nextSemaID == 0 {
		m.nextSemaID = 1
	}

	m.heapPtr, m.heapEnd = s.HeapPtr, s.HeapEnd
	m.vblanks = s.VBlanks
	m.padLive = padLiveState{
		buttons: s.PadLiveButtons,
		lx:      s.PadLiveLX,
		ly:      s.PadLiveLY,
		rx:      s.PadLiveRX,
		ry:      s.PadLiveRY,
	}
	m.tty = append([]byte(nil), s.TTY...)
	m.sbus = s.SBus

	// The second processor. It is rebuilt over the same 2 MiB the EE shares with it, which
	// LoadState has already restored above — so the modules' code, their patched stubs and
	// every thread's stack are in place before the CPU that runs them exists.
	m.IOP = nil
	if s.IOP != nil {
		if err := m.LoadIOPState(*s.IOP); err != nil {
			return err
		}
		// The second processor is restored, not started — so it never passed through
		// StartIOP, where the harness's instruments (the trap, the write-watch, the
		// register and stub-call traces) attach to it. Without this they silently do
		// nothing on a `-loadstate` resume: the very state you snapshot in order to
		// examine the frontier is the one state in which you cannot watch the IOP. So
		// the same hook StartIOP fires runs here too, and a resumed IOP is instrumented
		// exactly as a booted one is.
		if m.OnIOPStart != nil {
			m.OnIOPStart(m.IOP)
		}
		// A symbol-named trap is resolved to an address when its module loads; a resumed
		// IOP loads no modules, so it is resolved here instead, now the hook above has
		// named it.
		m.IOP.resolveTrap()
		// Pokes apply on a resume too, for the same reason the instruments do: the state
		// you snapshot to examine the frontier is exactly the one where you want to turn a
		// module's tracing on or nudge a global and watch what unblocks. On a boot they are
		// written in RebootIOPFrom; a resume never passes through it.
		for addr, val := range m.IOPPokes {
			m.IOP.Write32(addr, val)
			m.note("IOP: poked 0x%08X = 0x%08X (%s) on resume", addr, val, m.IOP.Sym(addr))
		}
	}

	m.SyscallCalls = map[string]int{}
	for k, v := range s.SyscallCalls {
		m.SyscallCalls[k] = v
	}

	for i := range m.dmac {
		cs := s.DmacChans[i]
		m.dmac[i] = dmacChan{
			chcr: cs.CHCR, madr: cs.MADR, qwc: cs.QWC, tadr: cs.TADR,
			asr0: cs.ASR0, asr1: cs.ASR1, sadr: cs.SADR,
		}
	}
	m.dCtrl, m.dPcr, m.dSqwc, m.dRbsr, m.dRbor = s.DCtrl, s.DPcr, s.DSqwc, s.DRbsr, s.DRbor
	m.gs = nil
	if g := s.GS; g != nil {
		gs := m.ensureGS()
		copy(gs.vram, g.VRAM)
		gs.reg, gs.csr = g.Reg, g.CSR
		gs.xfer = gsXfer{
			active: g.Xfer.Active, dbp: g.Xfer.DBP, dbw: g.Xfer.DBW, dpsm: g.Xfer.DPSM,
			dsax: g.Xfer.DSAX, dsay: g.Xfer.DSAY, rrw: g.Xfer.RRW, rrh: g.Xfer.RRH,
			x: g.Xfer.X, y: g.Xfer.Y, partial: append([]byte(nil), g.Xfer.Partial...),
		}
		gs.vqN, gs.q, gs.clut, gs.cbp0, gs.cbp1 = g.VQN, g.Q, g.CLUT, g.CBP0, g.CBP1
		gs.uploads, gs.prims, gs.primCount = g.Uploads, g.Prims, g.PrimCount
		gs.primsRunBase = g.Prims
		gs.path2ImageRemain, gs.path2SkipRemain = g.Path2ImageRemain, g.Path2SkipRemain
		gs.path2Carry = append([]byte(nil), g.Path2Carry...)
		for i, v := range g.VQ {
			gs.vq[i] = gsVertex{x: v.PX, y: v.PY, z: v.Z, rgba: v.RGBA, u: v.U, v: v.V, s: v.S, t: v.T, q: v.Q}
		}
		gs.drawCensus = map[string]int{}
		for k, v := range g.DrawCensus {
			gs.drawCensus[k] = v
		}
	}
	// The VIFs are rebuilt through ensureVIF so the host wiring — the XGKICK callback,
	// VU0 as the EE's COP2 — is re-established, then their memories and registers are
	// overlaid. A machine restored without this ran every resume with empty VU memories,
	// which a CPU probe never notices and the first MSCAL after a resume dies on.
	m.vifs = [2]*vif{}
	m.CPU.COP2 = nil
	for i, vs := range []*VIFState{s.VIF0, s.VIF1} {
		if vs == nil {
			continue
		}
		v := m.ensureVIF(i)
		v.cl, v.wl, v.mode, v.mask = vs.CL, vs.WL, vs.Mode, vs.Mask
		v.row, v.col = vs.Row, vs.Col
		v.base, v.ofst, v.itop, v.mark, v.tops = vs.Base, vs.Ofst, vs.ITop, vs.Mark, vs.Tops
		v.cmd, v.pending = vs.Cmd, vs.Pending
		v.buf = append([]byte(nil), vs.Buf...)
		copy(v.micro, vs.Micro)
		copy(v.data, vs.Data)
		v.vuSteps = vs.VUSteps
		if vs.VU != nil {
			v.vu.Restore(*vs.VU)
		}
		for k, n := range vs.Census {
			v.census[k] = n
		}
		for k, n := range vs.MSCALs {
			v.mscal[k] = n
		}
	}

	// A state saved at a halt must not re-halt on the same instruction the moment it
	// is loaded — the whole reason to save one there is to look at what happened next.
	m.Halted, m.HaltReason = false, ""
	m.CPU.Halted, m.CPU.HaltReason = false, ""
	return nil
}

// SaveStateFile writes a snapshot, gzipped.
//
// It used to refuse while the IOP was running, because the snapshot carried the IOP's
// memory and not the IOP. It carries the IOP now (iopstate.go).
//
// What it still refuses is a snapshot taken from *inside* a routine the machine itself
// called — a module's entry point, an interrupt handler, a scheduler hook. That is not a
// state the second processor is ever in on its own account; it is a state the host has put
// it in, half way down a Go call stack that no file can hold. A snapshot taken there would
// restore a processor mid-call with nothing to return to.
func (m *Machine) SaveStateFile(path string) error {
	if m.IOP != nil && m.IOP.callDepth > 0 {
		return fmt.Errorf("ps2: the IOP is inside a routine the machine called (%s), and there is "+
			"no way to carry a Go call stack in a savestate — snapshot it between calls",
			m.IOP.Sym(m.IOP.CPU.PC))
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z := gzip.NewWriter(f)
	defer z.Close()
	return gob.NewEncoder(z).Encode(m.SaveState())
}

// LoadStateFile reads a snapshot back.
func (m *Machine) LoadStateFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer z.Close()
	var s MachineState
	if err := gob.NewDecoder(z).Decode(&s); err != nil {
		return err
	}
	return m.LoadState(s)
}
