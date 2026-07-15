package dos

// Save-states for the protected-mode (go32) machine — the PM counterpart to
// dos_state.go, which covers only the real-mode Machine. A boot to a rendered
// Quake frame is tens of millions of instructions; a savestate lets that be
// reached once and then restored instantly, as many times as needed, which is
// what the debugger's resume and the oracle-capability-parity rule both want.
//
// The state is a deep copy so it survives the machine that made it and can be
// restored repeatedly: the whole 65 MiB linear space, the CPU (registers, flags,
// mode, the x87 stack), and the PM bookkeeping (the bump arenas, the modelled LDT,
// the recorded PM interrupt vectors, the PIT, the DAC palette, and the input
// queues). Three things are NOT in it and are rebound rather than serialised:
//
//   - the x86.CPU function fields (SegResolve/IntHook/OnStep/PortIn/PortOut). They
//     close over this *PM, so an in-memory restore keeps the live ones; a from-disk
//     load rebuilds a fresh machine first and restores into it, so they are already
//     wired to that machine.
//   - the open host files (map[uint16]*os.File). In memory they are kept as-is; on
//     disk each is recorded by path + offset + writability and reopened under
//     gameDir (Quake's data files live there and persist on disk, so there is no
//     scratch overlay to recreate as the real-mode Machine has).
//   - defIntVec, the synthesized default IRQ reflector. Its bytes live in Mem (which
//     is restored) and it points into the info block (whose base is restored), so it
//     is simply reconstructed with setupDefaultIntVec.

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"retroreverse.com/tools/cpu/x86"
)

// PMState is a fully-serialisable snapshot of a PM machine. Every field is
// exported, and every nested struct with unexported fields (the PIT, the PM
// vectors, the scripted-input events) is mirrored by an exported one here, because
// gob encodes only exported fields.
type PMState struct {
	Mem []byte

	// CPU
	Regs                               [8]uint32
	Seg                                [8]uint16
	SegBase                            [8]uint32
	IP                                 uint32
	CF, PF, AF, ZF, SF, TF, IF, DF, OF bool
	Mode                               int
	Steps                              uint64
	Ext386                             uint64
	Halted                             bool
	HaltReason                         string
	FPU                                x86.FPUState // all fields exported

	// PM bookkeeping / arenas
	ConvBase, ConvNext, ConvTop uint32
	HeapBase, HeapNext          uint32
	StackFloor                  uint32
	InfoBase                    uint32
	ImgEnd                      uint32

	// modelled LDT + selector allocation
	Sels         map[uint16]uint32
	NextSel      uint16
	NextCallback uint16
	VirtIF       bool

	// fabricated DOS far pointers
	LolSeg, LolOff uint16
	DtaSeg, DtaOff uint16

	// PM interrupt vectors (mirror of the unexported pmVector)
	PMVectors [256]pmVecSnap

	// devices
	Pit      pitSnap
	Retrace  bool
	DacIndex int
	Pal      [768]byte

	// input
	KeyEvents []injSnap // scripted schedule (usually empty under the debugger)
	KeyWait   int
	KeyRetry  bool
	InjTick   int
	KeyHits   int
	KbdData   byte
	KbdFull   bool
	InjKeys   []byte // interactive make/break scancodes queued for ASAP delivery

	// program status
	Terminated bool
	ExitCode   byte
	Console    []byte

	// open files
	Files []pmFileSnap
}

// pmVecSnap mirrors pmVector (whose fields are unexported) for gob.
type pmVecSnap struct {
	Sel uint16
	Off uint32
	Set bool
}

// pitSnap mirrors pitState for gob.
type pitSnap struct {
	Reload    uint16
	WriteHi   bool
	Latched   uint16
	HaveLatch bool
	ReadHi    bool
}

// injSnap mirrors injEvent for gob.
type injSnap struct {
	Kind  byte
	Code  byte
	X, Y  int
	Delay int
}

// pmFileSnap records an open host file so it can be reopened on a from-disk load.
type pmFileSnap struct {
	Handle uint16
	Path   string
	Offset int64
	Writable bool // opened for writing (reopen O_RDWR), else read-only
}

// SaveState captures the whole machine as an independent deep copy.
func (p *PM) SaveState() *PMState {
	c := p.CPU
	s := &PMState{
		Mem: append([]byte(nil), p.Mem...),

		Regs: c.Regs, Seg: c.Seg, SegBase: c.SegBase, IP: c.IP,
		CF: c.CF, PF: c.PF, AF: c.AF, ZF: c.ZF, SF: c.SF,
		TF: c.TF, IF: c.IF, DF: c.DF, OF: c.OF,
		Mode: c.Mode, Steps: c.Steps, Ext386: c.Ext386,
		Halted: c.Halted, HaltReason: c.HaltReason, FPU: c.FPU,

		ConvBase: p.convBase, ConvNext: p.convNext, ConvTop: p.convTop,
		HeapBase: p.heapBase, HeapNext: p.heapNext,
		StackFloor: p.stackFloor, InfoBase: p.infoBase, ImgEnd: p.imgEnd,

		NextSel: p.nextSel, NextCallback: p.nextCallback, VirtIF: p.virtIF,
		LolSeg: p.lolSeg, LolOff: p.lolOff, DtaSeg: p.dtaSeg, DtaOff: p.dtaOff,

		Pit:      pitSnap{Reload: p.pit.reload, WriteHi: p.pit.writeHi, Latched: p.pit.latched, HaveLatch: p.pit.haveLatch, ReadHi: p.pit.readHi},
		Retrace:  p.retrace,
		DacIndex: p.dacIndex,
		Pal:      p.Pal,

		KeyWait: p.keyWait, KeyRetry: p.keyRetry, InjTick: p.injTick, KeyHits: p.keyHits,
		KbdData: p.kbdData, KbdFull: p.kbdFull,
		InjKeys: append([]byte(nil), p.injKeys...),

		Terminated: p.Terminated, ExitCode: p.ExitCode,
		Console: append([]byte(nil), p.Console...),
	}
	s.Sels = map[uint16]uint32{}
	for k, v := range p.sels {
		s.Sels[k] = v
	}
	for i := range p.pmVectors {
		s.PMVectors[i] = pmVecSnap{Sel: p.pmVectors[i].sel, Off: p.pmVectors[i].off, Set: p.pmVectors[i].set}
	}
	for _, e := range p.keyEvents {
		s.KeyEvents = append(s.KeyEvents, injSnap{Kind: byte(e.kind), Code: e.code, X: e.x, Y: e.y, Delay: e.delay})
	}
	for h, f := range p.files {
		if f == nil {
			continue
		}
		off, _ := f.Seek(0, io.SeekCurrent)
		s.Files = append(s.Files, pmFileSnap{Handle: h, Path: f.Name(), Offset: off, Writable: fileWritable(f)})
	}
	return s
}

// LoadState restores a snapshot into this machine in place. The CPU's hooks and bus
// stay wired to this PM, so it must be a machine LoadGo32 built (its function fields
// close over it). Open files are closed and reopened from the snapshot.
func (p *PM) LoadState(s *PMState) error {
	if len(s.Mem) != len(p.Mem) {
		return fmt.Errorf("go32: snapshot mem size %d != %d", len(s.Mem), len(p.Mem))
	}
	copy(p.Mem, s.Mem)

	c := p.CPU
	c.Regs, c.Seg, c.SegBase, c.IP = s.Regs, s.Seg, s.SegBase, s.IP
	c.CF, c.PF, c.AF, c.ZF, c.SF = s.CF, s.PF, s.AF, s.ZF, s.SF
	c.TF, c.IF, c.DF, c.OF = s.TF, s.IF, s.DF, s.OF
	c.Mode, c.Steps, c.Ext386 = s.Mode, s.Steps, s.Ext386
	c.Halted, c.HaltReason, c.FPU = s.Halted, s.HaltReason, s.FPU

	p.convBase, p.convNext, p.convTop = s.ConvBase, s.ConvNext, s.ConvTop
	p.heapBase, p.heapNext = s.HeapBase, s.HeapNext
	p.stackFloor, p.infoBase, p.imgEnd = s.StackFloor, s.InfoBase, s.ImgEnd

	p.nextSel, p.nextCallback, p.virtIF = s.NextSel, s.NextCallback, s.VirtIF
	p.lolSeg, p.lolOff, p.dtaSeg, p.dtaOff = s.LolSeg, s.LolOff, s.DtaSeg, s.DtaOff

	p.sels = map[uint16]uint32{}
	for k, v := range s.Sels {
		p.sels[k] = v
	}
	for i := range s.PMVectors {
		p.pmVectors[i] = pmVector{sel: s.PMVectors[i].Sel, off: s.PMVectors[i].Off, set: s.PMVectors[i].Set}
	}

	p.pit = pitState{reload: s.Pit.Reload, writeHi: s.Pit.WriteHi, latched: s.Pit.Latched, haveLatch: s.Pit.HaveLatch, readHi: s.Pit.ReadHi}
	p.retrace = s.Retrace
	p.dacIndex = s.DacIndex
	p.Pal = s.Pal

	p.keyEvents = nil
	for _, e := range s.KeyEvents {
		p.keyEvents = append(p.keyEvents, injEvent{kind: injKind(e.Kind), code: e.Code, x: e.X, y: e.Y, delay: e.Delay})
	}
	p.keyWait, p.keyRetry, p.injTick, p.keyHits = s.KeyWait, s.KeyRetry, s.InjTick, s.KeyHits
	p.kbdData, p.kbdFull = s.KbdData, s.KbdFull
	p.injKeys = append([]byte(nil), s.InjKeys...)

	p.Terminated, p.ExitCode = s.Terminated, s.ExitCode
	p.Console = append([]byte(nil), s.Console...)

	// The default IRQ reflector is derived, not stored: its bytes are in the restored
	// Mem and it points into the (restored) info block.
	p.setupDefaultIntVec()

	// Reopen the open files. On an in-memory restore this path is not taken (Restore
	// keeps the live handles); it runs for a from-disk load, where the files array was
	// filled from a fresh SaveState and the handles are gone.
	for _, f := range p.files {
		if f != nil {
			f.Close()
		}
	}
	p.files = map[uint16]*os.File{}
	for _, fs := range s.Files {
		var f *os.File
		var err error
		if fs.Writable {
			f, err = os.OpenFile(fs.Path, os.O_RDWR, 0644)
		}
		if f == nil { // not writable, or the RDWR open failed — fall back to read-only
			f, err = os.Open(fs.Path)
		}
		if err != nil {
			return fmt.Errorf("go32: reopen %q: %w", fs.Path, err)
		}
		f.Seek(fs.Offset, io.SeekStart)
		p.files[fs.Handle] = f
	}
	return nil
}

// SaveStateFile writes the machine's state to path (gzip-compressed gob).
func (p *PM) SaveStateFile(path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := gzip.NewWriter(out)
	defer zw.Close()
	return gob.NewEncoder(zw).Encode(p.SaveState())
}

// LoadStateFile restores a state written by SaveStateFile.
func (p *PM) LoadStateFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	zr, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer zr.Close()
	var s PMState
	if err := gob.NewDecoder(zr).Decode(&s); err != nil {
		return err
	}
	return p.LoadState(&s)
}

// fileWritable reports whether a host file was opened with write access, so a
// from-disk load reopens it the same way. It probes rather than tracks the open
// mode (os.File does not expose it): a zero-length write at the current offset
// succeeds on a writable handle and fails on a read-only one, and it moves nothing.
func fileWritable(f *os.File) bool {
	off, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	_, werr := f.WriteAt(nil, off)
	return werr == nil
}
