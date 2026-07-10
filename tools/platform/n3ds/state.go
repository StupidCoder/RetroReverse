package n3ds

// state.go gives the 3DS oracle save-states: dump the whole machine to a file and
// restore it into a live Machine. Per the repo's oracle-capability-parity rule,
// state comes up with the first platform phase — reaching an interesting point in
// a title costs a long run, and a snapshot pays that once and resumes from it.
//
// Determinism: the oracle paces everything by instruction count (Run, the tick
// counter), so restoring a snapshot and running N more instructions lands on
// exactly the state an uninterrupted run reaches. That makes a savestate a
// regression instrument, not merely a speed-up (TestSaveStateRoundTrip).
//
// Discipline: every field of Machine and its CPU that affects future execution
// must be serialised here in the same commit that adds it. A snapshot that
// silently omits a register resumes subtly wrong, which is worse than one that
// refuses to load. The round-trip test enforces it.

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/arm"
)

const snapshotVersion = 2

// cpuState is the serialisable ARM11 register/flag file, used for the running CPU
// and for each thread's saved context.
type cpuState struct {
	R          [16]uint32
	N, Z, C, V bool
	Q          bool
	GE         uint32
	Thumb      bool
	BigEndian  bool
	IRQDisable bool
	FIQDisable bool
	Mode       uint32
	Arch       int
	Instrs     uint64

	// VFPv2 coprocessor.
	VFP   [32]uint32
	FPSCR uint32
	FPEXC uint32
}

func toCPUState(c *arm.CPU) cpuState {
	return cpuState{
		R: c.R, N: c.N, Z: c.Z, C: c.C, V: c.V, Q: c.Q, GE: c.GE,
		Thumb: c.Thumb, BigEndian: c.BigEndian, IRQDisable: c.IRQDisable,
		FIQDisable: c.FIQDisable, Mode: c.Mode, Arch: int(c.Arch), Instrs: c.Instrs,
		VFP: c.VFP.S, FPSCR: c.VFP.FPSCR, FPEXC: c.VFP.FPEXC,
	}
}

func (cs cpuState) into(c *arm.CPU) {
	c.R, c.N, c.Z, c.C, c.V, c.Q, c.GE = cs.R, cs.N, cs.Z, cs.C, cs.V, cs.Q, cs.GE
	c.Thumb, c.BigEndian, c.IRQDisable, c.FIQDisable = cs.Thumb, cs.BigEndian, cs.IRQDisable, cs.FIQDisable
	c.Mode, c.Arch, c.Instrs = cs.Mode, arm.Variant(cs.Arch), cs.Instrs
	c.VFP.S, c.VFP.FPSCR, c.VFP.FPEXC = cs.VFP, cs.FPSCR, cs.FPEXC
}

// regionState is one mapped span.
type regionState struct {
	Name string
	Base uint32
	Data []byte
}

// kobjState is a serialised kernel object.
type kobjState struct {
	Handle      uint32
	Kind        string
	Name        string
	Signal      bool
	ManualReset bool
	SemCount    int32
	MutexOwner  uint32
	MutexDepth  int
	Waiters     []uint32
	BlockAddr   uint32
	BlockSize   uint32
	ThreadID    uint32 // for kind=="thread": the thread's id (0 = none), relinked on load
}

// threadSnap is a serialised thread.
type threadSnap struct {
	ID, Handle, TLSBase, Tpidr uint32
	Priority                   int32
	State                      int
	WakeTick                   uint64
	WaitAll                    bool
	WaitOn                     []uint32
	ArbAddr                    uint32
	Ctx                        cpuState
}

type snapshot struct {
	Version   int
	ProgramID uint64

	CPU     cpuState
	Regions []regionState

	HeapPtr    uint32
	LinearPtr  uint32
	NextHandle uint32
	Tick       uint64

	Threads     []threadSnap
	CurThreadID uint32
	NextThread  uint32
	NextTLS     uint32
	RRCursor    int

	Objects  []kobjState
	Ports    map[uint32]string
	Services map[uint32]string
	SVCLog   []svcEvent
	DebugOut []byte
}

// SaveState writes a gzip-compressed gob snapshot of the machine to path.
func (m *Machine) SaveState(path string) error {
	// The current thread's live state is in CPU; sync it back to its ctx so the
	// snapshot is self-consistent.
	m.curThread.ctx = *m.CPU

	s := snapshot{
		Version:     snapshotVersion,
		ProgramID:   m.programID,
		HeapPtr:     m.heapPtr,
		LinearPtr:   m.linearPtr,
		NextHandle:  m.nextHandle,
		Tick:        m.tick,
		CurThreadID: m.curThread.id,
		NextThread:  m.nextThread,
		NextTLS:     m.nextTLS,
		RRCursor:    m.rrCursor,
		Ports:       m.ports,
		Services:    m.services,
		SVCLog:      m.svcLog,
		DebugOut:    m.debugOut,
	}
	s.CPU = toCPUState(m.CPU)
	for _, r := range m.regions {
		s.Regions = append(s.Regions, regionState{Name: r.name, Base: r.base, Data: r.data})
	}
	for _, t := range m.threads {
		s.Threads = append(s.Threads, threadSnap{
			ID: t.id, Handle: t.handle, TLSBase: t.tlsBase, Tpidr: t.tpidr,
			Priority: t.priority, State: int(t.state), WakeTick: t.wakeTick,
			WaitAll: t.waitAll, WaitOn: t.waitOn, ArbAddr: t.arbAddr, Ctx: toCPUState(&t.ctx),
		})
	}
	for h, o := range m.handles {
		ks := kobjState{
			Handle: h, Kind: o.kind, Name: o.name, Signal: o.signal, ManualReset: o.manualReset,
			SemCount: o.semCount, MutexOwner: o.mutexOwner, MutexDepth: o.mutexDepth,
			Waiters: o.waiters, BlockAddr: o.blockAddr, BlockSize: o.blockSize,
		}
		if o.thread != nil {
			ks.ThreadID = o.thread.id
		}
		s.Objects = append(s.Objects, ks)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	return gob.NewEncoder(gz).Encode(&s)
}

// LoadState restores a snapshot into the machine. It refuses a snapshot from a
// different title (program ID) or an incompatible version — a mismatched restore
// is a silent corruption, so it is made an error.
func (m *Machine) LoadState(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	var s snapshot
	if err := gob.NewDecoder(gz).Decode(&s); err != nil {
		return err
	}
	if s.Version != snapshotVersion {
		return fmt.Errorf("n3ds: snapshot version %d, want %d", s.Version, snapshotVersion)
	}
	if s.ProgramID != m.programID {
		return fmt.Errorf("n3ds: snapshot is for program %016x, this machine is %016x", s.ProgramID, m.programID)
	}

	s.CPU.into(m.CPU)
	m.CPU.Halted, m.CPU.HaltReason = false, ""

	m.regions = nil
	m.codeReg, m.stackReg, m.tlsReg, m.heapReg, m.linearReg = nil, nil, nil, nil, nil
	for _, rs := range s.Regions {
		r := m.mapRegion(rs.Name, rs.Base, rs.Data)
		switch rs.Name {
		case "code":
			m.codeReg = r
		case "stack":
			m.stackReg = r
		case "tls":
			m.tlsReg = r
		case "heap":
			m.heapReg = r
		case "linear":
			m.linearReg = r
		}
	}
	m.heapPtr, m.linearPtr, m.nextHandle = s.HeapPtr, s.LinearPtr, s.NextHandle
	m.tick = s.Tick
	m.nextThread, m.nextTLS, m.rrCursor = s.NextThread, s.NextTLS, s.RRCursor
	m.ports, m.services, m.svcLog, m.debugOut = s.Ports, s.Services, s.SVCLog, s.DebugOut

	m.threads = nil
	for _, ts := range s.Threads {
		t := &thread{
			id: ts.ID, handle: ts.Handle, tlsBase: ts.TLSBase, tpidr: ts.Tpidr,
			priority: ts.Priority, state: threadState(ts.State), wakeTick: ts.WakeTick,
			waitAll: ts.WaitAll, waitOn: ts.WaitOn, arbAddr: ts.ArbAddr,
		}
		ts.Ctx.into(&t.ctx)
		m.threads = append(m.threads, t)
		if t.id == s.CurThreadID {
			m.curThread = t
		}
	}

	m.handles = map[uint32]*kobject{}
	for _, o := range s.Objects {
		ko := &kobject{
			kind: o.Kind, name: o.Name, signal: o.Signal, manualReset: o.ManualReset,
			semCount: o.SemCount, mutexOwner: o.MutexOwner, mutexDepth: o.MutexDepth,
			waiters: o.Waiters, blockAddr: o.BlockAddr, blockSize: o.BlockSize,
		}
		if o.ThreadID != 0 {
			ko.thread = m.threadByID(o.ThreadID)
		}
		m.handles[o.Handle] = ko
	}
	return nil
}
