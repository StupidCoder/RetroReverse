package dos

// Save-states for the oracle: dump the entire machine (the 1 MiB address space,
// the CPU, the VGA/EMS/mouse/timer device state, and the open files) to a file,
// and restore it into a freshly-loaded Machine. This lets an expensive boot —
// e.g. character creation all the way into the dungeon — be reached once and
// then resumed instantly, as many times as needed, from the snapshot.
//
// Two things can't be serialised directly and are handled specially:
//   - Open files are host os.File handles. We record each handle's path, current
//     offset, and (for files the game created in the writable scratch overlay)
//     its contents, then reopen/recreate them on load — preserving the DOS
//     handle numbers, which the MS-C runtime indexes by.
//   - The scratch overlay is a fresh temp dir each run, so scratch files are
//     recreated under the new run's scratch dir from the saved contents.
//
// LoadState overwrites a live Machine in place, so the CPU's hooks (IntHook,
// OnStep, ports, bus) stay wired to that Machine.

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// snapshot is the fully-serialisable machine state (all fields exported for gob).
type snapshot struct {
	Mem []byte

	// CPU
	Regs                               [8]uint32
	Seg                                [8]uint16
	IP                                 uint32
	CF, PF, AF, ZF, SF, TF, IF, DF, OF bool
	Steps                              uint64
	Ext386                             uint64
	Halted                             bool
	HaltReason                         string

	// Machine layout / DOS state
	PspSeg, EnvSeg, LoadSeg, MemTop, FirstMCB uint16
	DtaSeg, DtaOff                            uint16
	KeyWait, KeyHits                          int
	EnableIRQ                                 bool
	Terminated                                bool
	ExitCode                                  byte

	Files []fileSnap

	// EMS
	EmsBacking  []byte
	EmsNextPage int
	EmsHandles  map[uint16][2]int // base,count
	EmsNextH    uint16
	EmsSlot     [4]int
	EmsSaved    map[uint16][4]int

	// I/O ports (keyboard/PIT/SB/DAC)
	KbdOut     byte
	KbdOutFull bool
	ExpectData bool
	Retrace    bool
	Pit        uint16
	OplReg     byte
	OplTimer   bool
	Tick       uint32
	MixReg     byte
	MixRegs    [256]byte
	DspQueue   []byte
	DspReset   byte
	DacIndex   int
	Pal        [768]byte

	// VGA
	Planes  [4][0x10000]byte
	Latch   [4]byte
	SeqIdx  byte
	Seq     [8]byte
	GcIdx   byte
	Gc      [16]byte
	CrtcIdx byte
	Crtc    [32]byte

	// Mouse
	MsX, MsY, MsAccX, MsAccY int
	MsButtons                byte
	MsXMin, MsXMax           int
	MsYMin, MsYMax           int
	MsHidden                 bool
	MsPressL, MsPressR       int
}

// fileSnap records an open DOS file so it can be reopened on load.
type fileSnap struct {
	Handle  uint16
	Path    string // host path (as opened)
	Offset  int64
	Scratch bool   // under the writable scratch overlay
	Rel     string // path relative to the scratch dir (for recreation)
	Data    []byte // scratch-file contents to recreate
}

// SaveState writes the machine's full state to path (gzip-compressed gob).
func (m *Machine) SaveState(path string) error {
	s := &snapshot{
		Mem: append([]byte(nil), m.Mem...),

		Regs: m.CPU.Regs, Seg: m.CPU.Seg, IP: m.CPU.IP,
		CF: m.CPU.CF, PF: m.CPU.PF, AF: m.CPU.AF, ZF: m.CPU.ZF, SF: m.CPU.SF,
		TF: m.CPU.TF, IF: m.CPU.IF, DF: m.CPU.DF, OF: m.CPU.OF,
		Steps: m.CPU.Steps, Ext386: m.CPU.Ext386,
		Halted: m.CPU.Halted, HaltReason: m.CPU.HaltReason,

		PspSeg: m.pspSeg, EnvSeg: m.envSeg, LoadSeg: m.loadSeg, MemTop: m.memTop,
		FirstMCB: m.firstMCB, DtaSeg: m.dtaSeg, DtaOff: m.dtaOff,
		KeyWait: m.keyWait, KeyHits: m.keyHits, EnableIRQ: m.EnableIRQ,
		Terminated: m.Terminated, ExitCode: m.ExitCode,
	}

	// EMS
	if m.ems != nil {
		s.EmsBacking = append([]byte(nil), m.ems.backing...)
		s.EmsNextPage = m.ems.nextPage
		s.EmsNextH = m.ems.nextH
		s.EmsSlot = m.ems.slot
		s.EmsHandles = map[uint16][2]int{}
		for h, v := range m.ems.handles {
			s.EmsHandles[h] = [2]int{v.base, v.count}
		}
		s.EmsSaved = map[uint16][4]int{}
		for h, v := range m.ems.saved {
			s.EmsSaved[h] = v
		}
	}
	// I/O
	if m.io != nil {
		s.KbdOut, s.KbdOutFull, s.ExpectData, s.Retrace = m.io.kbdOut, m.io.kbdOutFull, m.io.expectData, m.io.retrace
		s.Pit, s.OplReg, s.OplTimer, s.Tick = m.io.pit, m.io.oplReg, m.io.oplTimer, m.io.tick
		s.MixReg, s.MixRegs, s.DspReset = m.io.mixReg, m.io.mixRegs, m.io.dspReset
		s.DspQueue = append([]byte(nil), m.io.dspQueue...)
		s.DacIndex, s.Pal = m.io.dacIndex, m.io.Pal
	}
	// VGA
	if m.vga != nil {
		s.Planes = m.vga.planes
		s.Latch, s.SeqIdx, s.Seq = m.vga.latch, m.vga.seqIdx, m.vga.seq
		s.GcIdx, s.Gc, s.CrtcIdx, s.Crtc = m.vga.gcIdx, m.vga.gc, m.vga.crtcIdx, m.vga.crtc
	}
	// Mouse
	if m.ms != nil {
		s.MsX, s.MsY, s.MsAccX, s.MsAccY = m.ms.x, m.ms.y, m.ms.accX, m.ms.accY
		s.MsButtons = m.ms.buttons
		s.MsXMin, s.MsXMax, s.MsYMin, s.MsYMax = m.ms.xMin, m.ms.xMax, m.ms.yMin, m.ms.yMax
		s.MsHidden, s.MsPressL, s.MsPressR = m.ms.hidden, m.ms.pressL, m.ms.pressR
	}
	// Open files
	for h, f := range m.files {
		if f == nil {
			continue
		}
		off, _ := f.Seek(0, io.SeekCurrent)
		fs := fileSnap{Handle: h, Path: f.Name(), Offset: off}
		if m.scratchDir != "" && strings.HasPrefix(f.Name(), m.scratchDir) {
			fs.Scratch = true
			fs.Rel = strings.TrimPrefix(f.Name(), m.scratchDir)
			fs.Data, _ = os.ReadFile(f.Name()) // separate handle: doesn't move off
		}
		s.Files = append(s.Files, fs)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := gzip.NewWriter(out)
	defer zw.Close()
	return gob.NewEncoder(zw).Encode(s)
}

// LoadState restores a snapshot into this (freshly LoadEXE'd) Machine in place.
func (m *Machine) LoadState(path string) error {
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
	var s snapshot
	if err := gob.NewDecoder(zr).Decode(&s); err != nil {
		return err
	}
	if len(s.Mem) != len(m.Mem) {
		return fmt.Errorf("dos: snapshot mem size %d != %d", len(s.Mem), len(m.Mem))
	}
	copy(m.Mem, s.Mem)

	m.CPU.Regs, m.CPU.Seg, m.CPU.IP = s.Regs, s.Seg, s.IP
	m.CPU.CF, m.CPU.PF, m.CPU.AF, m.CPU.ZF, m.CPU.SF = s.CF, s.PF, s.AF, s.ZF, s.SF
	m.CPU.TF, m.CPU.IF, m.CPU.DF, m.CPU.OF = s.TF, s.IF, s.DF, s.OF
	m.CPU.Steps, m.CPU.Ext386 = s.Steps, s.Ext386
	m.CPU.Halted, m.CPU.HaltReason = s.Halted, s.HaltReason

	m.pspSeg, m.envSeg, m.loadSeg, m.memTop = s.PspSeg, s.EnvSeg, s.LoadSeg, s.MemTop
	m.firstMCB, m.dtaSeg, m.dtaOff = s.FirstMCB, s.DtaSeg, s.DtaOff
	m.keyWait, m.keyHits, m.EnableIRQ = s.KeyWait, s.KeyHits, s.EnableIRQ
	m.Terminated, m.ExitCode = s.Terminated, s.ExitCode

	if m.ems != nil && s.EmsBacking != nil {
		copy(m.ems.backing, s.EmsBacking)
		m.ems.nextPage, m.ems.nextH, m.ems.slot = s.EmsNextPage, s.EmsNextH, s.EmsSlot
		m.ems.handles = map[uint16]emsHandle{}
		for h, v := range s.EmsHandles {
			m.ems.handles[h] = emsHandle{base: v[0], count: v[1]}
		}
		m.ems.saved = map[uint16][4]int{}
		for h, v := range s.EmsSaved {
			m.ems.saved[h] = v
		}
	}
	if m.io != nil {
		m.io.kbdOut, m.io.kbdOutFull, m.io.expectData, m.io.retrace = s.KbdOut, s.KbdOutFull, s.ExpectData, s.Retrace
		m.io.pit, m.io.oplReg, m.io.oplTimer, m.io.tick = s.Pit, s.OplReg, s.OplTimer, s.Tick
		m.io.mixReg, m.io.mixRegs, m.io.dspReset = s.MixReg, s.MixRegs, s.DspReset
		m.io.dspQueue = append([]byte(nil), s.DspQueue...)
		m.io.dacIndex, m.io.Pal = s.DacIndex, s.Pal
	}
	if m.vga != nil {
		m.vga.planes = s.Planes
		m.vga.latch, m.vga.seqIdx, m.vga.seq = s.Latch, s.SeqIdx, s.Seq
		m.vga.gcIdx, m.vga.gc, m.vga.crtcIdx, m.vga.crtc = s.GcIdx, s.Gc, s.CrtcIdx, s.Crtc
	}
	if m.ms != nil {
		m.ms.x, m.ms.y, m.ms.accX, m.ms.accY = s.MsX, s.MsY, s.MsAccX, s.MsAccY
		m.ms.buttons = s.MsButtons
		m.ms.xMin, m.ms.xMax, m.ms.yMin, m.ms.yMax = s.MsXMin, s.MsXMax, s.MsYMin, s.MsYMax
		m.ms.hidden, m.ms.pressL, m.ms.pressR = s.MsHidden, s.MsPressL, s.MsPressR
	}

	// Close any files the fresh boot left open, then reopen the snapshot's.
	for _, f := range m.files {
		if f != nil {
			f.Close()
		}
	}
	m.files = map[uint16]*os.File{}
	m.finds = map[uint32]*findState{} // no FindNext in progress across a snapshot
	for _, fs := range s.Files {
		var f *os.File
		if fs.Scratch {
			np := filepath.Join(m.scratchDir, fs.Rel)
			if err := os.MkdirAll(filepath.Dir(np), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(np, fs.Data, 0o644); err != nil {
				return err
			}
			f, err = os.OpenFile(np, os.O_RDWR, 0o644)
		} else {
			f, err = os.Open(fs.Path)
		}
		if err != nil {
			return fmt.Errorf("dos: reopen %q: %w", fs.Path, err)
		}
		f.Seek(fs.Offset, io.SeekStart)
		m.files[fs.Handle] = f
	}
	return nil
}
