package dc

// state.go is the savestate — mandatory in a platform's first phase (the
// oracle-capability-parity rule). The format is the repo's: gzip over gob
// with a version constant, the disc identity pinned by MD5 rather than
// serialized, and every slice deep-copied on the way in AND the way out —
// a snapshot that shares a backing array with the live machine is not a
// snapshot, and the round-trip test is structurally blind to it (see
// state_test.go's independence test).

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/cpu/sh4"
)

// ARMState is the sound CPU's snapshot: active registers, packed CPSR, and
// the banked files a mode switch would swap in.
type ARMState struct {
	R          [16]uint32
	CPSR       uint32
	Banks      arm.Banks
	Halted     bool
	HaltReason string
	Instrs     uint64
}

func saveARM(c *arm.CPU) ARMState {
	return ARMState{R: c.R, CPSR: c.CPSR(), Banks: c.SaveBanks(),
		Halted: c.Halted, HaltReason: c.HaltReason, Instrs: c.Instrs}
}

func restoreARM(c *arm.CPU, s ARMState) {
	c.R = s.R
	c.SetCPSR(s.CPSR)
	c.RestoreBanks(s.Banks)
	c.Halted, c.HaltReason, c.Instrs = s.Halted, s.HaltReason, s.Instrs
}

const snapshotVersion = 1

// MachineState is a complete Dreamcast snapshot.
type MachineState struct {
	Version int
	DiscMD5 string

	RAM     []byte
	VRAM    []byte
	AICARAM []byte
	Flash   []byte

	CPU   sh4.State
	Holly Holly
	Maple mapleState
	Pad   PadState

	// The sound side: the ARM7's whole programmer's model (banked registers
	// included — the DS lesson), its run state, and the register block.
	ARM        ARMState
	ARMRunning bool
	ArmAcc     int
	AICARegs   map[uint32]uint32
	Timers     aicaTimers

	C2DStat, C2DLen                                uint32
	RenderCountdown, C2DCountdown, C2DPendingBits uint32
	TAList                                        int32
	TAVtx, TANeed, TAFifoCountdown                uint32
	TAFrame                                       []byte
	TAClosed                                      []byte
	TAFifo                                        []byte

	// BIOS is nil for a machine that never booted; the request map is
	// deep-copied both directions.
	BIOS *biosState

	PVRRegs  [0x2000 / 4]uint32
	TAWrites uint64

	Instrs       uint64
	Fields       uint64
	CurLine      uint32
	FieldNum     uint32
	InstrInField uint64
}

// cloneBIOS deep-copies the HLE state: the map and every request behind it.
func cloneBIOS(s *biosState) *biosState {
	if s == nil {
		return nil
	}
	out := &biosState{NextID: s.NextID, Requests: make(map[uint32]*gdRequest, len(s.Requests))}
	for id, req := range s.Requests {
		r := *req
		out.Requests[id] = &r
	}
	return out
}

func cloneBytes(b []byte) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

// SaveState captures the machine.
func (m *Machine) SaveState() MachineState {
	s := MachineState{
		Version: snapshotVersion,
		RAM:     cloneBytes(m.RAM),
		VRAM:    cloneBytes(m.VRAM),
		AICARAM: cloneBytes(m.AICARAM),
		Flash:   cloneBytes(m.Flash),
		CPU:     m.CPU.Snapshot(),
		Holly:   m.Holly,
		Maple:   m.Maple,
		Pad:     m.Pad,
		C2DStat: m.C2DStat, C2DLen: m.C2DLen,
		RenderCountdown: m.RenderCountdown, C2DCountdown: m.C2DCountdown, C2DPendingBits: m.C2DPendingBits,
		TAList: m.TAList, TAVtx: m.TAVtx, TANeed: m.TANeed, TAFifoCountdown: m.TAFifoCountdown,
		TAFrame: cloneBytes(m.TAFrame), TAClosed: cloneBytes(m.TAClosed), TAFifo: cloneBytes(m.TAFifo),
		ARM: saveARM(m.ARM), ARMRunning: m.ARMRunning, ArmAcc: m.armAcc, Timers: m.Timers,
		AICARegs: make(map[uint32]uint32, len(m.AICARegs)),
		PVRRegs:  m.PVRRegs, TAWrites: m.TAWrites,
		Instrs: m.Instrs, Fields: m.Fields, CurLine: m.CurLine, FieldNum: m.FieldNum, InstrInField: m.instrInField,
	}
	for k, v := range m.AICARegs {
		s.AICARegs[k] = v
	}
	if m.bios != nil {
		s.BIOS = cloneBIOS(&m.bios.state)
	}
	if m.Disc != nil {
		s.DiscMD5, _ = m.Disc.MD5()
	}
	return s
}

// LoadState restores a snapshot into the machine, refusing one taken on a
// different disc. The disc itself is re-mounted by the caller; the state
// carries only its identity.
func (m *Machine) LoadState(s MachineState) error {
	if s.Version != snapshotVersion {
		return fmt.Errorf("savestate version %d, this build reads %d", s.Version, snapshotVersion)
	}
	if m.Disc != nil && s.DiscMD5 != "" {
		if sum, _ := m.Disc.MD5(); sum != s.DiscMD5 {
			return fmt.Errorf("savestate is for a different disc (state %s, mounted %s)", s.DiscMD5, sum)
		}
	}
	if len(s.RAM) != RAMSize || len(s.VRAM) != VRAMSize {
		return fmt.Errorf("savestate memory sizes do not match this machine")
	}
	copy(m.RAM, s.RAM)
	copy(m.VRAM, s.VRAM)
	copy(m.AICARAM, s.AICARAM)
	copy(m.Flash, s.Flash)
	m.CPU.Restore(s.CPU)
	m.Holly = s.Holly
	m.Maple = s.Maple
	m.Pad = s.Pad
	m.C2DStat, m.C2DLen = s.C2DStat, s.C2DLen
	m.RenderCountdown, m.C2DCountdown, m.C2DPendingBits = s.RenderCountdown, s.C2DCountdown, s.C2DPendingBits
	m.TAList, m.TAVtx = s.TAList, s.TAVtx
	m.TANeed, m.TAFifoCountdown = s.TANeed, s.TAFifoCountdown
	m.TAFrame = cloneBytes(s.TAFrame)
	m.TAClosed = cloneBytes(s.TAClosed)
	m.TAFifo = cloneBytes(s.TAFifo)
	restoreARM(m.ARM, s.ARM)
	m.ARMRunning, m.armAcc, m.Timers = s.ARMRunning, s.ArmAcc, s.Timers
	m.AICARegs = make(map[uint32]uint32, len(s.AICARegs))
	for k, v := range s.AICARegs {
		m.AICARegs[k] = v
	}
	if s.BIOS != nil {
		m.bios = newBIOS(m)
		m.bios.state = *cloneBIOS(s.BIOS)
	} else {
		m.bios = nil
	}
	m.PVRRegs, m.TAWrites = s.PVRRegs, s.TAWrites
	m.Instrs, m.Fields, m.instrInField = s.Instrs, s.Fields, s.InstrInField
	m.CurLine, m.FieldNum = s.CurLine, s.FieldNum
	return nil
}

// SaveStateFile writes gzip(gob(state)).
func (m *Machine) SaveStateFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := gzip.NewWriter(f)
	if err := gob.NewEncoder(zw).Encode(m.SaveState()); err != nil {
		return err
	}
	return zw.Close()
}

// LoadStateFile restores from a file written by SaveStateFile.
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
	var s MachineState
	if err := gob.NewDecoder(zr).Decode(&s); err != nil {
		return err
	}
	return m.LoadState(s)
}
