package xbox

// state.go is the Xbox machine's savestate: a versioned, gzip'd gob snapshot, the same
// discipline as tools/platform/n3ds/state.go and tools/platform/dos/go32_state.go.
// Savestate is mandatory in a platform's first phase (the oracle-capability-parity
// rule): a boot to the first NV2A push is tens of millions of instructions, and a
// snapshot lets that be reached once and restored instantly, as many times as the
// debugger's resume and later render passes need.
//
// The snapshot is a deep copy that survives the machine that made it: the whole 64 MB
// of RAM, the CPU (registers, flags, mode, the x87 stack), the allocators and clock,
// the sparse NV2A register map, and the threads/objects. Two kinds of thing are
// rebound rather than serialised:
//
//   - the x86.CPU function fields (SegResolve/OnStep/PortIn/PortOut). They close over
//     this *Machine, so a from-disk load builds a fresh machine and restores into it,
//     leaving them wired to that machine.
//   - the mounted disc handle. It is reopened by the caller (the bootoracle passes the
//     same -image), not held in the state.
//
// Thread and object graphs are flattened to index-addressed snapshots and rebuilt on
// load, so the shared *thread pointers (Machine.current, kobject.thread) are restored
// without gob having to chase cycles.

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"retroreverse.com/tools/cpu/x86"
)

const stateVersion = 1

// XboxState is a fully-serialisable snapshot. Every field is exported; nested structs
// with unexported fields are mirrored by exported snapshot types below.
type XboxState struct {
	Version int

	RAM []byte

	// CPU
	Regs                               [8]uint32
	Seg                                [8]uint16
	SegBase                            [8]uint32
	IP                                 uint32
	CF, PF, AF, ZF, SF, TF, IF, DF, OF bool
	Mode                               int
	Steps                              uint64
	Halted                             bool
	HaltReason                         string
	FPU                                x86.FPUState
	XMM                                [8][16]byte

	// Allocators / clock
	PoolNext, HeapNext, HeapTop uint32
	NextObjAddr, KbandNext      uint32
	Tick                        uint64

	// NV2A
	NVReg     map[uint32]uint32
	NVPut     uint32
	NVGet     uint32
	NVKicked  bool
	FirstPush bool

	// MCPX latch apertures (apu.go). The log-once bookkeeping is not state.
	// NICReg decodes as nil from pre-NIC snapshots, which restores an honest
	// cold latch (copyU32Map(nil) is an empty map).
	APUReg  map[uint32]uint32
	AC97Reg map[uint32]uint32
	USBReg  map[uint32]uint32
	NICReg  map[uint32]uint32

	// NV2A graphics engine (nv2a_pgraph.go) + the PFIFO pusher's decode position
	// (nv2a_pfifo.go). The survey/unhandled instrumentation maps are NOT state.
	PgSubObject [8]uint32
	PgSubClass  [8]uint32
	PgRegs      [0x800]uint32
	PgMethods   int
	PgSetObjs   int
	Push        pusherSnap

	PCIAddr  uint32
	PCISpace map[uint32]byte

	PoolSizes map[uint32]uint32 // ExAllocatePoolWithTag block -> size

	ShaCtx map[uint32][]byte // XcSHA* streaming contexts (marshalled crypto/sha1 state)
	Rc4Ctx map[uint32][]byte // XcRC4* key schedules (258-byte S/i/j state)

	// Kernel HLE bookkeeping
	OrdinalHits map[uint16]int
	NextTID     uint32
	RRCursor    int
	QuantumLeft int

	// Threads and objects (index-addressed)
	Threads   []threadSnap
	CurThread int // index into Threads, or -1
	Objects   []objSnap
}

// pusherSnap mirrors pusherState (all unexported), so a savestate resumes mid-method
// if a batch was split across two DMA_PUT writes.
type pusherSnap struct {
	Method, Subchan, Count uint32
	NonInc                 bool
	SubReturn              uint32
	SubActive              bool
}

// threadSnap mirrors thread with its saved CPU context.
type threadSnap struct {
	ID       uint32
	KThread  uint32
	Ctx      cpuSnap
	Priority int32
	State    int
	WakeTick uint64
	WaitAll  bool
	WaitObjs []uint32
	WaitReg  int
}

// cpuSnap mirrors the parts of x86.CPU a saved (non-running) thread carries.
type cpuSnap struct {
	Regs                               [8]uint32
	Seg                                [8]uint16
	SegBase                            [8]uint32
	IP                                 uint32
	CF, PF, AF, ZF, SF, TF, IF, DF, OF bool
	Mode                               int
	Steps                              uint64
	Halted                             bool
	FPU                                x86.FPUState
	XMM                                [8][16]byte
}

// objSnap mirrors kobject; the thread it may reference is restored by KThread lookup.
type objSnap struct {
	Kind     string
	Addr     uint32
	Signaled bool
	Count    int32
	Limit    int32
	IsThread bool
}

func snapCPU(c *x86.CPU) cpuSnap {
	return cpuSnap{
		Regs: c.Regs, Seg: c.Seg, SegBase: c.SegBase, IP: c.IP,
		CF: c.CF, PF: c.PF, AF: c.AF, ZF: c.ZF, SF: c.SF, TF: c.TF, IF: c.IF, DF: c.DF, OF: c.OF,
		Mode: c.Mode, Steps: c.Steps, Halted: c.Halted, FPU: c.FPU, XMM: c.XMM,
	}
}

func (s cpuSnap) into(c *x86.CPU) {
	c.Regs, c.Seg, c.SegBase, c.IP = s.Regs, s.Seg, s.SegBase, s.IP
	c.CF, c.PF, c.AF, c.ZF, c.SF, c.TF, c.IF, c.DF, c.OF =
		s.CF, s.PF, s.AF, s.ZF, s.SF, s.TF, s.IF, s.DF, s.OF
	c.Mode, c.Steps, c.Halted, c.FPU, c.XMM = s.Mode, s.Steps, s.Halted, s.FPU, s.XMM
}

// SaveState captures the machine into an in-memory snapshot.
func (m *Machine) SaveState() *XboxState {
	c := m.CPU
	// Flush the running thread's live CPU into its ctx so every thread snapshot is
	// uniform (the current thread's authoritative state is the live CPU).
	curIdx := -1
	for i, t := range m.threads {
		if t == m.current {
			curIdx = i
		}
	}
	st := &XboxState{
		Version: stateVersion,
		RAM:     append([]byte(nil), m.RAM...),
		Regs:    c.Regs, Seg: c.Seg, SegBase: c.SegBase, IP: c.IP,
		CF: c.CF, PF: c.PF, AF: c.AF, ZF: c.ZF, SF: c.SF, TF: c.TF, IF: c.IF, DF: c.DF, OF: c.OF,
		Mode: c.Mode, Steps: c.Steps, Halted: c.Halted, HaltReason: c.HaltReason, FPU: c.FPU, XMM: c.XMM,
		PoolNext: m.poolNext, HeapNext: m.heapNext, HeapTop: m.heapTop,
		NextObjAddr: m.nextObjAddr, KbandNext: m.kbandNext, Tick: m.tick,
		NVReg: copyU32Map(m.nv.reg), NVPut: m.nv.dmaPut, NVGet: m.nv.dmaGet, NVKicked: m.nv.kicked,
		APUReg:      copyU32Map(m.apu.reg),
		AC97Reg:     copyU32Map(m.ac97.reg),
		USBReg:      copyU32Map(m.usb.reg),
		NICReg:      copyU32Map(m.nic.reg),
		PgSubObject: m.pgraph.subObject, PgSubClass: m.pgraph.subClass, PgRegs: m.pgraph.Regs,
		PgMethods: m.pgraph.Methods, PgSetObjs: m.pgraph.SetObjs,
		Push: pusherSnap{
			Method: m.push.method, Subchan: m.push.subchan, Count: m.push.count,
			NonInc: m.push.nonInc, SubReturn: m.push.subReturn, SubActive: m.push.subActive,
		},
		FirstPush: m.firstPush, PCIAddr: m.pciAddr, PCISpace: copyByteMap(m.pciSpace),
		PoolSizes:   copyU32Map(m.poolSizes),
		ShaCtx:      copyByteSliceMap(m.shaCtx),
		Rc4Ctx:      copyByteSliceMap(m.rc4Ctx),
		OrdinalHits: copyOrdMap(m.OrdinalHits),
		NextTID:     m.nextTID, RRCursor: m.rrCursor, QuantumLeft: m.quantumLeft,
		CurThread: curIdx,
	}
	for _, t := range m.threads {
		ctx := t.ctx
		if t == m.current {
			ctx = *c // the live CPU is the current thread's true context
		}
		st.Threads = append(st.Threads, threadSnap{
			ID: t.id, KThread: t.kthread, Ctx: snapCPU(&ctx), Priority: t.priority,
			State: int(t.state), WakeTick: t.wakeTick, WaitAll: t.waitAll,
			WaitObjs: append([]uint32(nil), t.waitObjs...), WaitReg: t.waitReg,
		})
	}
	// Objects in a stable order (by address) for deterministic snapshots.
	addrs := make([]uint32, 0, len(m.objects))
	for a := range m.objects {
		addrs = append(addrs, a)
	}
	sortU32(addrs)
	for _, a := range addrs {
		o := m.objects[a]
		st.Objects = append(st.Objects, objSnap{
			Kind: o.kind, Addr: o.addr, Signaled: o.signaled,
			Count: o.count, Limit: o.limit, IsThread: o.thread != nil,
		})
	}
	return st
}

// LoadState restores an in-memory snapshot into this machine in place, keeping the
// live CPU's function-field wiring (they close over this *Machine).
func (m *Machine) LoadState(st *XboxState) error {
	if st.Version != stateVersion {
		return fmt.Errorf("xbox: savestate version %d, want %d", st.Version, stateVersion)
	}
	copy(m.RAM, st.RAM)
	c := m.CPU
	c.Regs, c.Seg, c.SegBase, c.IP = st.Regs, st.Seg, st.SegBase, st.IP
	c.CF, c.PF, c.AF, c.ZF, c.SF, c.TF, c.IF, c.DF, c.OF =
		st.CF, st.PF, st.AF, st.ZF, st.SF, st.TF, st.IF, st.DF, st.OF
	c.Mode, c.Steps, c.Halted, c.HaltReason, c.FPU, c.XMM = st.Mode, st.Steps, st.Halted, st.HaltReason, st.FPU, st.XMM
	m.poolNext, m.heapNext, m.heapTop = st.PoolNext, st.HeapNext, st.HeapTop
	m.nextObjAddr, m.kbandNext, m.tick = st.NextObjAddr, st.KbandNext, st.Tick
	m.nv.reg = copyU32Map(st.NVReg)
	m.nv.dmaPut, m.nv.dmaGet, m.nv.kicked = st.NVPut, st.NVGet, st.NVKicked
	m.apu.reg = copyU32Map(st.APUReg)
	m.ac97.reg = copyU32Map(st.AC97Reg)
	m.usb.reg = copyU32Map(st.USBReg)
	m.nic.reg = copyU32Map(st.NICReg)
	m.pgraph.subObject, m.pgraph.subClass, m.pgraph.Regs = st.PgSubObject, st.PgSubClass, st.PgRegs
	m.pgraph.Methods, m.pgraph.SetObjs = st.PgMethods, st.PgSetObjs
	m.push = pusherState{
		method: st.Push.Method, subchan: st.Push.Subchan, count: st.Push.Count,
		nonInc: st.Push.NonInc, subReturn: st.Push.SubReturn, subActive: st.Push.SubActive,
	}
	m.firstPush, m.pciAddr = st.FirstPush, st.PCIAddr
	m.pciSpace = copyByteMap(st.PCISpace)
	m.poolSizes = copyU32Map(st.PoolSizes)
	m.shaCtx = copyByteSliceMap(st.ShaCtx)
	m.rc4Ctx = copyByteSliceMap(st.Rc4Ctx)
	m.Halted, m.HaltReason = st.Halted, st.HaltReason
	m.OrdinalHits = copyOrdMap(st.OrdinalHits)
	m.nextTID, m.rrCursor, m.quantumLeft = st.NextTID, st.RRCursor, st.QuantumLeft

	// Rebuild the thread list and objects, then re-link shared pointers by address.
	m.threads = m.threads[:0]
	m.current = nil
	for i, ts := range st.Threads {
		t := &thread{
			id: ts.ID, kthread: ts.KThread, priority: ts.Priority,
			state: threadState(ts.State), wakeTick: ts.WakeTick, waitAll: ts.WaitAll,
			waitObjs: append([]uint32(nil), ts.WaitObjs...), waitReg: ts.WaitReg,
		}
		// Seed the saved context from the live CPU first, so its unexported bus and
		// hook pointers (which close over this machine and are NOT serialised) are the
		// live ones; then overlay the snapshot's architectural register state. Without
		// this, switchTo's `*m.CPU = t.ctx` would clobber the bus with a nil pointer.
		t.ctx = *c
		ts.Ctx.into(&t.ctx)
		m.threads = append(m.threads, t)
		if i == st.CurThread {
			m.current = t
			// The current thread's context is the live CPU (already restored above).
		}
	}
	m.objects = map[uint32]*kobject{}
	for _, os := range st.Objects {
		o := &kobject{kind: os.Kind, addr: os.Addr, signaled: os.Signaled, count: os.Count, limit: os.Limit}
		if os.IsThread {
			for _, t := range m.threads {
				if t.kthread == os.Addr {
					o.thread = t
					break
				}
			}
		}
		m.objects[os.Addr] = o
	}
	return nil
}

// SaveStateFile / LoadStateFile persist a snapshot to disk (gzip'd gob).
func (m *Machine) SaveStateFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := gzip.NewWriter(f)
	defer zw.Close()
	return gob.NewEncoder(zw).Encode(m.SaveState())
}

func (m *Machine) LoadStateFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer zr.Close()
	var st XboxState
	if err := gob.NewDecoder(zr).Decode(&st); err != nil && err != io.EOF {
		return err
	}
	return m.LoadState(&st)
}

// --- small helpers ---

func copyU32Map(src map[uint32]uint32) map[uint32]uint32 {
	dst := make(map[uint32]uint32, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
func copyByteMap(src map[uint32]byte) map[uint32]byte {
	dst := make(map[uint32]byte, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
func copyByteSliceMap(src map[uint32][]byte) map[uint32][]byte {
	dst := make(map[uint32][]byte, len(src))
	for k, v := range src {
		dst[k] = append([]byte(nil), v...)
	}
	return dst
}
func copyOrdMap(src map[uint16]int) map[uint16]int {
	dst := make(map[uint16]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
func sortU32(a []uint32) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}
