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

const snapshotVersion = 3

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

	// Graphics: the GSP bring-up state (v3) …
	NotifyWaiters   []uint32
	APTNotifyEv     uint32
	APTResumeEv     uint32
	APTWakePending  bool
	APTParams       []aptParam
	GSPShared       uint32
	GSPSharedAddr   uint32
	GSPEvent        uint32
	HIDShared       uint32
	HIDSharedAddr   uint32
	HIDEvents       []uint32
	NextFrameInstr  uint64
	VBlankCount     uint64
	FramesSubmitted int
	FramesSwapped   int
	DisplayXfers    int
	LastXferTop     xferState
	LastXferBottom  xferState

	// … and the PICA200 GPU (v3). The upload cursors and partial FIFO buffers
	// matter: a snapshot can land mid-upload.
	GPU gpuState

	// Open fs sessions (v3): stored by path and re-resolved from the RomFS on
	// load, so the snapshot stays small.
	FSFiles    []fsFileState
	FSDirs     []fsDirState
	FSArchives map[uint32]uint32
	SaveFiles      map[string][]byte
	SaveFormatted  bool
	SaveFormatInfo [4]uint32
}

type fsFileState struct {
	Handle uint32
	Path   string
	Save   string
}

type fsDirState struct {
	Handle uint32
	Path   string
	Cursor int
}

// xferState mirrors xferRecord for the snapshot.
type xferState struct {
	Dst, W, H, Format, BPP, Stride uint32
}

func toXferState(r xferRecord) xferState {
	return xferState{Dst: r.dst, W: r.w, H: r.h, Format: r.format, BPP: r.bpp, Stride: r.stride}
}

func (x xferState) into() xferRecord {
	return xferRecord{dst: x.Dst, w: x.W, h: x.H, format: x.Format, bpp: x.BPP, stride: x.Stride}
}

// gpuState is the serialised PICA200.
type gpuState struct {
	Regs   [0x300]uint32
	Code   [4096]uint32
	Opdesc [128]uint32
	Float  [96][4]float32
	Bool   uint32
	Int    [4][4]uint8

	CodeIdx, OpdIdx int
	FltIdx          int
	FltF32          bool
	FltBuf          []uint32
	FixedIdx        int
	FixedBuf        []uint32
	FixedVal        [16][4]float32

	GshCodeIdx, GshOpdIdx, GshFltIdx int
	GshFltF32                        bool
	GshFltBuf                        []uint32

	Draws int
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

		NotifyWaiters:   m.notifyWaiters,
		APTNotifyEv:     m.aptNotifyEv,
		APTResumeEv:     m.aptResumeEv,
		APTWakePending:  m.aptWakePending,
		APTParams:       m.aptParams,
		GSPShared:       m.gspShared,
		GSPSharedAddr:   m.gspSharedAddr,
		GSPEvent:        m.gspEvent,
		HIDShared:       m.hidShared,
		HIDSharedAddr:   m.hidSharedAddr,
		HIDEvents:       m.hidEvents,
		NextFrameInstr:  m.nextFrameInstr,
		VBlankCount:     m.vblankCount,
		FramesSubmitted: m.framesSubmitted,
		FramesSwapped:   m.framesSwapped,
		DisplayXfers:    m.displayTransfers,
		LastXferTop:     toXferState(m.lastXferTop),
		LastXferBottom:  toXferState(m.lastXferBottom),
	}
	g := m.gpu
	s.GPU = gpuState{
		Regs: g.Regs, Code: g.Code, Opdesc: g.Opdesc, Float: g.Float,
		Bool: g.Bool, Int: g.Int,
		CodeIdx: g.codeIdx, OpdIdx: g.opdIdx,
		FltIdx: g.fltIdx, FltF32: g.fltF32, FltBuf: g.fltBuf,
		FixedIdx: g.fixedIdx, FixedBuf: g.fixedBuf, FixedVal: g.fixedVal,
		GshCodeIdx: g.gshCodeIdx, GshOpdIdx: g.gshOpdIdx,
		GshFltIdx: g.gshFltIdx, GshFltF32: g.gshFltF32, GshFltBuf: g.gshFltBuf,
		Draws: g.Draws,
	}
	for h, f := range m.fsFiles {
		s.FSFiles = append(s.FSFiles, fsFileState{Handle: h, Path: f.path, Save: f.save})
	}
	for h, d := range m.fsDirs {
		s.FSDirs = append(s.FSDirs, fsDirState{Handle: h, Path: d.path, Cursor: d.cursor})
	}
	s.FSArchives = m.fsArchives
	s.SaveFiles = m.saveFiles
	s.SaveFormatted, s.SaveFormatInfo = m.saveFormatted, m.saveFormatInfo
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

	m.notifyWaiters = s.NotifyWaiters
	m.aptNotifyEv, m.aptResumeEv = s.APTNotifyEv, s.APTResumeEv
	m.aptWakePending = s.APTWakePending
	m.aptParams = s.APTParams
	m.gspShared, m.gspSharedAddr, m.gspEvent = s.GSPShared, s.GSPSharedAddr, s.GSPEvent
	m.hidShared, m.hidSharedAddr, m.hidEvents = s.HIDShared, s.HIDSharedAddr, s.HIDEvents
	// Older snapshots predate HID serialisation; recover the mapped address from the
	// restored region so -hidtrace/-keys work on states saved before this field existed.
	if m.hidSharedAddr == 0 {
		for _, r := range m.regions {
			if r.name == "hid-shared" {
				m.hidSharedAddr = r.base
				break
			}
		}
	}
	m.nextFrameInstr, m.vblankCount = s.NextFrameInstr, s.VBlankCount
	m.framesSubmitted, m.framesSwapped = s.FramesSubmitted, s.FramesSwapped
	m.displayTransfers = s.DisplayXfers
	m.lastXferTop, m.lastXferBottom = s.LastXferTop.into(), s.LastXferBottom.into()

	g := m.gpu
	g.Regs, g.Code, g.Opdesc, g.Float = s.GPU.Regs, s.GPU.Code, s.GPU.Opdesc, s.GPU.Float
	g.Bool, g.Int = s.GPU.Bool, s.GPU.Int
	g.codeIdx, g.opdIdx = s.GPU.CodeIdx, s.GPU.OpdIdx
	g.fltIdx, g.fltF32, g.fltBuf = s.GPU.FltIdx, s.GPU.FltF32, s.GPU.FltBuf
	g.fixedIdx, g.fixedBuf, g.fixedVal = s.GPU.FixedIdx, s.GPU.FixedBuf, s.GPU.FixedVal
	g.gshCodeIdx, g.gshOpdIdx = s.GPU.GshCodeIdx, s.GPU.GshOpdIdx
	g.gshFltIdx, g.gshFltF32, g.gshFltBuf = s.GPU.GshFltIdx, s.GPU.GshFltF32, s.GPU.GshFltBuf
	g.Draws = s.GPU.Draws
	g.texCache = nil // rebuilt on demand from restored memory

	m.threads = nil
	for _, ts := range s.Threads {
		t := &thread{
			id: ts.ID, handle: ts.Handle, tlsBase: ts.TLSBase, tpidr: ts.Tpidr,
			priority: ts.Priority, state: threadState(ts.State), wakeTick: ts.WakeTick,
			waitAll: ts.WaitAll, waitOn: ts.WaitOn, arbAddr: ts.ArbAddr,
		}
		// Seed the context from the live CPU so the bus and the SWI/coproc
		// hooks are populated — a context switch copies the whole CPU value,
		// and a zero-value context would hand the core a nil bus.
		t.ctx = *m.CPU
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
		// Snapshots from before these fields existed recover them by kind.
		if m.aptNotifyEv == 0 && o.Kind == "apt-notify" {
			m.aptNotifyEv = o.Handle
		}
		if m.aptResumeEv == 0 && o.Kind == "apt-resume" {
			m.aptResumeEv = o.Handle
		}
	}

	m.fsArchives = s.FSArchives
	if m.fsArchives == nil {
		m.fsArchives = map[uint32]uint32{}
	}
	m.saveFiles = s.SaveFiles
	m.saveFormatted, m.saveFormatInfo = s.SaveFormatted, s.SaveFormatInfo
	if m.saveFiles == nil {
		m.saveFiles = map[string][]byte{}
	}

	// Re-open the fs sessions from their paths.
	m.fsFiles = map[uint32]*fsFile{}
	for _, fsSt := range s.FSFiles {
		f := &fsFile{path: fsSt.Path, save: fsSt.Save}
		if fsSt.Save != "" {
			f.data = m.saveFiles[fsSt.Save]
			m.fsFiles[fsSt.Handle] = f
			continue
		}
		if fsSt.Path == "<romfs-l3>" && m.romfsRaw != nil {
			l3 := int64(0)
			if m.romfs != nil {
				l3 = m.romfs.Levels[2].Offset
			}
			f.data = m.romfsRaw[l3:]
		} else if m.romfs != nil {
			if d, err := m.romfs.File(fsSt.Path); err == nil {
				f.data = d
			}
		}
		m.fsFiles[fsSt.Handle] = f
	}
	m.fsDirs = map[uint32]*fsDir{}
	for _, ds := range s.FSDirs {
		if d, ok := m.romfsChildren(ds.Path); ok {
			d.cursor = ds.Cursor
			m.fsDirs[ds.Handle] = d
		}
	}
	return nil
}
