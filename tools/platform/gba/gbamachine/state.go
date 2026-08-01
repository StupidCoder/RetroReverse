package gbamachine

// Savestates — in the platform's FIRST phase, per the repository's
// oracle-capability-parity rule: every platform that did this late wished it
// had done it first. A cold boot is cheap on a GBA compared to a DS, but every
// question asked of a title screen still costs the boot before it can be asked.
//
// What is captured is everything the machine's behaviour depends on and nothing
// it does not. The ROM is NOT in the snapshot (it is the immutable input,
// supplied again at load, checked by game code). Every slice is deep-copied in
// BOTH directions — a snapshot that shares the machine's slices passes every
// resume test and is still not a snapshot (the repo's snapshot-aliasing scar).

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/arm"
)

const stateVersion = 1

type snapshot struct {
	Version  int
	GameCode string

	EWRAM, IWRAM, Pal, VRAM, OAM []byte
	IO                           map[uint32]uint16

	R          [16]uint32
	CPSR       uint32
	Banks      arm.Banks
	CPUHalted  bool
	HaltReason string
	Instrs     uint64

	IME      bool
	IE, IF   uint16
	Waiting  bool
	WaitAny  bool
	WaitMask uint16

	Line   int
	HBlank bool
	Frames uint64
	Steps  uint64
	Keys   uint16

	BG2X, BG2Y, BG3X, BG3Y int32

	DMA    [4]dmaChanState
	Timers [4]timerState

	EEPROMData  []byte
	EEPROMSized bool
	EEPROMIn    []byte
	EEPROMOut   []byte
	EEPROMReady bool

	ScreenPix []uint32
}

type dmaChanState struct {
	Src, Dst, LatchSrc, LatchDst uint32
	Count, Ctrl                  uint16
}

type timerState struct {
	Counter, Reload, Ctrl uint16
	Frac                  int
}

func cloneB(b []byte) []byte { return append([]byte(nil), b...) }

func (m *Machine) snapshot() *snapshot {
	s := &snapshot{
		Version: stateVersion, GameCode: m.GameCode(),
		EWRAM: cloneB(m.ewram), IWRAM: cloneB(m.iwram),
		Pal: cloneB(m.pal), VRAM: cloneB(m.vram), OAM: cloneB(m.oam),
		IO: map[uint32]uint16{},
		R:  m.cpu.R, CPSR: m.cpu.CPSR(), Banks: m.cpu.SaveBanks(),
		CPUHalted: m.cpu.Halted, HaltReason: m.cpu.HaltReason, Instrs: m.cpu.Instrs,
		IME: m.ime, IE: m.ie, IF: m.if_,
		Waiting: m.waiting, WaitAny: m.waitAny, WaitMask: m.waitMask,
		Line: m.vid.line, HBlank: m.vid.hblank, Frames: m.vid.frames,
		Steps: m.Steps, Keys: m.keys,
		BG2X: m.ppu.bg2x, BG2Y: m.ppu.bg2y, BG3X: m.ppu.bg3x, BG3Y: m.ppu.bg3y,
		EEPROMData: cloneB(m.eeprom.data), EEPROMSized: m.eeprom.sized,
		EEPROMIn: cloneB(m.eeprom.inBits), EEPROMOut: cloneB(m.eeprom.outBits),
		EEPROMReady: m.eeprom.ready,
		ScreenPix:   append([]uint32(nil), m.screen[:]...),
	}
	for k, v := range m.io {
		s.IO[k] = v
	}
	for i, d := range m.dma {
		s.DMA[i] = dmaChanState{d.src, d.dst, d.latchSrc, d.latchDst, d.count, d.ctrl}
	}
	for i, t := range m.timers {
		s.Timers[i] = timerState{t.counter, t.reload, t.ctrl, t.frac}
	}
	return s
}

func (m *Machine) restore(s *snapshot) {
	copy(m.ewram, s.EWRAM)
	copy(m.iwram, s.IWRAM)
	copy(m.pal, s.Pal)
	copy(m.vram, s.VRAM)
	copy(m.oam, s.OAM)
	m.io = map[uint32]uint16{}
	for k, v := range s.IO {
		m.io[k] = v
	}

	// CPSR first (it selects the bank), banks second, visible registers last —
	// the restore order dsmachine documents.
	m.cpu.SetCPSR(s.CPSR)
	m.cpu.RestoreBanks(s.Banks)
	m.cpu.R = s.R
	m.cpu.Halted, m.cpu.HaltReason, m.cpu.Instrs = s.CPUHalted, s.HaltReason, s.Instrs

	m.ime, m.ie, m.if_ = s.IME, s.IE, s.IF
	m.waiting, m.waitAny, m.waitMask = s.Waiting, s.WaitAny, s.WaitMask
	m.vid.line, m.vid.hblank, m.vid.frames = s.Line, s.HBlank, s.Frames
	m.Steps, m.keys = s.Steps, s.Keys
	m.ppu.bg2x, m.ppu.bg2y, m.ppu.bg3x, m.ppu.bg3y = s.BG2X, s.BG2Y, s.BG3X, s.BG3Y

	for i, d := range s.DMA {
		m.dma[i] = dmaChan{src: d.Src, dst: d.Dst, latchSrc: d.LatchSrc,
			latchDst: d.LatchDst, count: d.Count, ctrl: d.Ctrl}
	}
	for i, t := range s.Timers {
		m.timers[i] = timer{counter: t.Counter, reload: t.Reload, ctrl: t.Ctrl, frac: t.Frac}
	}

	m.eeprom.data = cloneB(s.EEPROMData)
	m.eeprom.sized = s.EEPROMSized
	m.eeprom.inBits = cloneB(s.EEPROMIn)
	m.eeprom.outBits = cloneB(s.EEPROMOut)
	m.eeprom.ready = s.EEPROMReady

	copy(m.screen[:], s.ScreenPix)
}

// SaveState writes the machine to a gzipped gob file.
func (m *Machine) SaveState(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z := gzip.NewWriter(f)
	if err := gob.NewEncoder(z).Encode(m.snapshot()); err != nil {
		return err
	}
	if err := z.Close(); err != nil {
		return err
	}
	return f.Close()
}

// LoadState restores a machine saved by SaveState. The machine must have been
// built from the same cartridge — the ROM is the one thing not carried.
func (m *Machine) LoadState(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	var s snapshot
	if err := gob.NewDecoder(z).Decode(&s); err != nil {
		return err
	}
	if s.Version != stateVersion {
		return fmt.Errorf("gbamachine: savestate version %d, want %d", s.Version, stateVersion)
	}
	if got := m.GameCode(); s.GameCode != got {
		return fmt.Errorf("gbamachine: savestate is for game %q, this machine runs %q", s.GameCode, got)
	}
	m.restore(&s)
	return nil
}
