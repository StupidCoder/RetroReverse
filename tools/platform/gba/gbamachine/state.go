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

	APU apuState

	EEPROMData  []byte
	EEPROMSized bool
	EEPROMIn    []byte
	EEPROMOut   []byte
	EEPROMReady bool

	ScreenPix []uint32
}

// apuState is the sound block. It is snapshotted for the same reason the rest
// is: a resume that silently restarts the music from bar one is not the machine
// the snapshot was taken of, and the tell (an envelope or a FIFO mid-note) is
// exactly the kind of state that looks unimportant until a sound bug is being
// chased across a savestate.
type apuState struct {
	Ch1, Ch2           squareState
	Ch3                waveState
	Ch4                noiseState
	DSAQueue, DSBQueue []int8
	DSACur, DSBCur     int8
	Powered            bool
	FrameAcc           float64
	FrameSeq           int
	SampleAcc          float64
	MixRegL, MixRegH   uint16
}

type squareState struct {
	Enabled, HasSweep, LengthEn, SwOn             bool
	Phase                                         float64
	DutySel, Freq, Length, Vol, EnvDir, EnvPeriod int
	EnvTimer, SwPeriod, SwShift, SwDir, SwTimer   int
	SwShadow                                      int
}

type waveState struct {
	Enabled, DacOn, LengthEn, Force75, TwoBanks bool
	Phase                                       float64
	Freq, Length, VolShift, Bank                int
	RAM                                         [2][16]byte
}

type noiseState struct {
	Enabled, LengthEn, Width7                bool
	LFSR                                     uint16
	Timer                                    float64
	Length, Vol, EnvDir, EnvPeriod, EnvTimer int
	Divisor, Shift                           int
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
		APU:         snapAPU(m.apu),
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
	restoreAPU(m.apu, &s.APU)
}

func snapSquare(c *square) squareState {
	return squareState{Enabled: c.enabled, HasSweep: c.hasSweep, LengthEn: c.lengthEn,
		SwOn: c.swOn, Phase: c.phase, DutySel: c.dutySel, Freq: c.freq, Length: c.length,
		Vol: c.vol, EnvDir: c.envDir, EnvPeriod: c.envPeriod, EnvTimer: c.envTimer,
		SwPeriod: c.swPeriod, SwShift: c.swShift, SwDir: c.swDir, SwTimer: c.swTimer,
		SwShadow: c.swShadow}
}

func loadSquare(c *square, s squareState) {
	c.enabled, c.hasSweep, c.lengthEn, c.swOn = s.Enabled, s.HasSweep, s.LengthEn, s.SwOn
	c.phase = s.Phase
	c.dutySel, c.freq, c.length, c.vol = s.DutySel, s.Freq, s.Length, s.Vol
	c.envDir, c.envPeriod, c.envTimer = s.EnvDir, s.EnvPeriod, s.EnvTimer
	c.swPeriod, c.swShift, c.swDir, c.swTimer, c.swShadow = s.SwPeriod, s.SwShift, s.SwDir, s.SwTimer, s.SwShadow
}

func snapAPU(a *apu) apuState {
	return apuState{
		Ch1: snapSquare(&a.ch1), Ch2: snapSquare(&a.ch2),
		Ch3: waveState{Enabled: a.ch3.enabled, DacOn: a.ch3.dacOn, LengthEn: a.ch3.lengthEn,
			Force75: a.ch3.force75, TwoBanks: a.ch3.twoBanks, Phase: a.ch3.phase,
			Freq: a.ch3.freq, Length: a.ch3.length, VolShift: a.ch3.volShift,
			Bank: a.ch3.bank, RAM: a.ch3.ram},
		Ch4: noiseState{Enabled: a.ch4.enabled, LengthEn: a.ch4.lengthEn, Width7: a.ch4.width7,
			LFSR: a.ch4.lfsr, Timer: a.ch4.timer, Length: a.ch4.length, Vol: a.ch4.vol,
			EnvDir: a.ch4.envDir, EnvPeriod: a.ch4.envPeriod, EnvTimer: a.ch4.envTimer,
			Divisor: a.ch4.divisor, Shift: a.ch4.shift},
		DSAQueue: append([]int8(nil), a.dsA.q...),
		DSBQueue: append([]int8(nil), a.dsB.q...),
		DSACur:   a.dsA.cur, DSBCur: a.dsB.cur,
		Powered: a.powered, FrameAcc: a.frameAcc, FrameSeq: a.frameSeq,
		SampleAcc: a.sampleAcc, MixRegL: a.mixRegL, MixRegH: a.mixRegH,
	}
}

func restoreAPU(a *apu, s *apuState) {
	loadSquare(&a.ch1, s.Ch1)
	loadSquare(&a.ch2, s.Ch2)
	a.ch3 = waveCh{enabled: s.Ch3.Enabled, dacOn: s.Ch3.DacOn, phase: s.Ch3.Phase,
		freq: s.Ch3.Freq, length: s.Ch3.Length, lengthEn: s.Ch3.LengthEn,
		volShift: s.Ch3.VolShift, force75: s.Ch3.Force75, ram: s.Ch3.RAM,
		bank: s.Ch3.Bank, twoBanks: s.Ch3.TwoBanks}
	a.ch4 = noiseCh{enabled: s.Ch4.Enabled, lfsr: s.Ch4.LFSR, timer: s.Ch4.Timer,
		length: s.Ch4.Length, lengthEn: s.Ch4.LengthEn, vol: s.Ch4.Vol,
		envDir: s.Ch4.EnvDir, envPeriod: s.Ch4.EnvPeriod, envTimer: s.Ch4.EnvTimer,
		divisor: s.Ch4.Divisor, shift: s.Ch4.Shift, width7: s.Ch4.Width7}
	a.dsA.q = append([]int8(nil), s.DSAQueue...)
	a.dsB.q = append([]int8(nil), s.DSBQueue...)
	a.dsA.cur, a.dsB.cur = s.DSACur, s.DSBCur
	a.powered, a.frameAcc, a.frameSeq = s.Powered, s.FrameAcc, s.FrameSeq
	a.sampleAcc, a.mixRegL, a.mixRegH = s.SampleAcc, s.MixRegL, s.MixRegH
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
