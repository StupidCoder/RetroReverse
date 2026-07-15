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

	// The syscalls the game replaced with its own routines. Losing these on load is
	// not a visible failure: the calls still "work", they just quietly do nothing,
	// because the kernel's stub answers instead of the game's code.
	UserSyscalls map[uint32]uint32

	Semas      []SemaState
	NextSemaID uint32

	HeapPtr, HeapEnd uint32
	VBlanks          uint32
	TTY              []byte
	SyscallCalls     map[string]int

	// The six registers both processors can see. They are shared silicon, and a snapshot
	// that dropped them would restore a machine in which one side of a handshake had
	// happened and the other had forgotten.
	SBus [sbusRegs]uint32

	// The second processor, entire (iopstate.go). Nil when the IOP was never started —
	// which is a state a PS2 really is in, right up until the game reboots it.
	IOP *IOPState
}

// SaveState captures the machine.
func (m *Machine) SaveState() MachineState {
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
		UserSyscalls:  map[uint32]uint32{},
		NextSemaID:    m.nextSemaID,
		HeapPtr:       m.heapPtr,
		HeapEnd:       m.heapEnd,
		VBlanks:       m.vblanks,
		TTY:           append([]byte(nil), m.tty...),
		SyscallCalls:  map[string]int{},
		SBus:          m.sbus,
	}
	if m.IOP != nil {
		iop := m.IOP.SaveState()
		s.IOP = &iop
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
