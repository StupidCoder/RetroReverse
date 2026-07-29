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

	"retroreverse.com/tools/cpu/sh4"
)

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

	PVRRegs  [0x2000 / 4]uint32
	TAWrites uint64

	Instrs       uint64
	Fields       uint64
	InstrInField uint64
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
		PVRRegs: m.PVRRegs, TAWrites: m.TAWrites,
		Instrs: m.Instrs, Fields: m.Fields, InstrInField: m.instrInField,
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
	m.PVRRegs, m.TAWrites = s.PVRRegs, s.TAWrites
	m.Instrs, m.Fields, m.instrInField = s.Instrs, s.Fields, s.InstrInField
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
