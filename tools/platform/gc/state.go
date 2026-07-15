package gc

// state.go snapshots the whole machine, so a boot that takes tens of seconds becomes a
// one-second load and a run can be resumed just before the thing worth watching.
//
// The rule this repository keeps is that a savestate must be resumable into a run that is
// bit-identical to the one that never stopped. That holds here because everything that
// advances without an instruction — the video field counter, the decrementer — is paced by
// the instruction count and is part of the state. The disc is not serialized (it is a large
// read-only file, re-mounted by the loader) but its MD5 is pinned, so a state cannot be
// resumed onto a different disc and quietly produce nonsense.
//
// The enforcement is TestSaveStateRoundTrip in state_test.go: every device field added to
// the machine has to be added here in the same change, or the round trip stops being
// identical and the test says so.

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/gekko"
)

const snapshotVersion = 1

// MachineState is the whole machine, with every field exported so gob can carry it.
type MachineState struct {
	Version int
	DiscMD5 string

	RAM  []byte
	ARAM []byte
	CPU  gekko.State

	PI  pi
	MI  mi
	VI  vi
	DI  di
	SI  si
	EXI exi
	AI  ai
	DSP dsp
	CP  cp
	PE  pe
	GPU gpu
	WG  wgPipe
}

// gob needs the unexported device fields, so each device provides an exported mirror. To
// avoid a second set of structs, the devices are stored whole — which works because gob can
// encode unexported fields only if... it cannot. So the devices expose their state through
// GobEncode/GobDecode below, keeping the register blocks' fields unexported while still
// making them serializable.

// SaveState captures the machine.
func (m *Machine) SaveState() MachineState {
	s := MachineState{
		Version: snapshotVersion,
		DiscMD5: m.discMD5,
		RAM:     append([]byte(nil), m.RAM...),
		ARAM:    append([]byte(nil), m.ARAM...),
		CPU:     m.CPU.Snapshot(),
		PI:      m.pi, MI: m.mi, VI: m.vi, DI: m.di, SI: m.si,
		EXI: m.exi, AI: m.ai, DSP: m.dsp, CP: m.cp, PE: m.pe, GPU: m.gpu, WG: m.wgFIFO,
	}
	// The DSP device is copied by value, but it holds a pointer to the running DSP core; deep-
	// copy that so the snapshot does not share mutable core state with the live machine.
	s.DSP.Core = m.dsp.Core.Clone()
	return s
}

// LoadState restores one, checking it belongs to this disc.
func (m *Machine) LoadState(s MachineState) error {
	if s.Version != snapshotVersion {
		return fmt.Errorf("savestate version %d, want %d", s.Version, snapshotVersion)
	}
	if m.discMD5 != "" && s.DiscMD5 != "" && s.DiscMD5 != m.discMD5 {
		return fmt.Errorf("savestate was taken on a different disc (%s, not %s)", s.DiscMD5, m.discMD5)
	}
	copy(m.RAM, s.RAM)
	copy(m.ARAM, s.ARAM)
	m.CPU.Restore(s.CPU)
	m.pi, m.mi, m.vi, m.di, m.si = s.PI, s.MI, s.VI, s.DI, s.SI
	m.exi, m.ai, m.dsp, m.cp, m.pe, m.gpu, m.wgFIFO = s.EXI, s.AI, s.DSP, s.CP, s.PE, s.GPU, s.WG
	// Deep-copy the DSP core out of the snapshot (the device was copied by value, sharing the
	// core pointer) and reattach its hardware bus, the back-reference left out of the state.
	if s.DSP.Core != nil {
		m.dsp.Core = s.DSP.Core.Clone()
		m.dsp.Core.SetBus(dspBus{m})
	}
	return nil
}

// SaveStateFile writes a snapshot to disk: gzip over gob, the format every machine in this
// repository uses.
func (m *Machine) SaveStateFile(path string) error {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if err := gob.NewEncoder(zw).Encode(m.SaveState()); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// LoadStateFile reads one back.
func (m *Machine) LoadStateFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	var s MachineState
	if err := gob.NewDecoder(zr).Decode(&s); err != nil {
		return err
	}
	return m.LoadState(s)
}
