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

const snapshotVersion = 1

// cpuState is the serialisable ARM11 register/flag file.
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

// regionState is one mapped span.
type regionState struct {
	Name string
	Base uint32
	Data []byte
}

// kobjState is a serialised kernel object.
type kobjState struct {
	Handle uint32
	Kind   string
	Name   string
	Signal bool
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
	Tpidr      uint32

	Objects  []kobjState
	Ports     map[uint32]string
	SVCLog    []svcEvent
	DebugOut  []byte
}

// SaveState writes a gzip-compressed gob snapshot of the machine to path.
func (m *Machine) SaveState(path string) error {
	s := snapshot{
		Version:    snapshotVersion,
		ProgramID:  m.programID,
		HeapPtr:    m.heapPtr,
		LinearPtr:  m.linearPtr,
		NextHandle: m.nextHandle,
		Tick:       m.tick,
		Tpidr:      m.tpidr,
		Ports:      m.ports,
		SVCLog:     m.svcLog,
		DebugOut:   m.debugOut,
	}
	c := m.CPU
	s.CPU = cpuState{
		R: c.R, N: c.N, Z: c.Z, C: c.C, V: c.V, Q: c.Q, GE: c.GE,
		Thumb: c.Thumb, BigEndian: c.BigEndian, IRQDisable: c.IRQDisable,
		FIQDisable: c.FIQDisable, Mode: c.Mode, Arch: int(c.Arch), Instrs: c.Instrs,
		VFP: c.VFP.S, FPSCR: c.VFP.FPSCR, FPEXC: c.VFP.FPEXC,
	}
	for _, r := range m.regions {
		s.Regions = append(s.Regions, regionState{Name: r.name, Base: r.base, Data: r.data})
	}
	for h, o := range m.handles {
		s.Objects = append(s.Objects, kobjState{Handle: h, Kind: o.kind, Name: o.name, Signal: o.signal})
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

	c := m.CPU
	cs := s.CPU
	c.R, c.N, c.Z, c.C, c.V, c.Q, c.GE = cs.R, cs.N, cs.Z, cs.C, cs.V, cs.Q, cs.GE
	c.Thumb, c.BigEndian, c.IRQDisable, c.FIQDisable = cs.Thumb, cs.BigEndian, cs.IRQDisable, cs.FIQDisable
	c.Mode, c.Arch, c.Instrs = cs.Mode, arm.Variant(cs.Arch), cs.Instrs
	c.VFP.S, c.VFP.FPSCR, c.VFP.FPEXC = cs.VFP, cs.FPSCR, cs.FPEXC
	c.Halted, c.HaltReason = false, ""

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
	m.tick, m.tpidr = s.Tick, s.Tpidr
	m.ports, m.svcLog, m.debugOut = s.Ports, s.SVCLog, s.DebugOut
	m.handles = map[uint32]*kobject{}
	for _, o := range s.Objects {
		m.handles[o.Handle] = &kobject{kind: o.Kind, name: o.Name, signal: o.Signal}
	}
	return nil
}
