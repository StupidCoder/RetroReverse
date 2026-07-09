package n64

// state.go gives the oracle save-states: dump the entire machine to a file and
// restore it into a live Machine. Reaching a title screen costs tens of millions
// of instructions, and the RDP is built by iterating on the frame that follows —
// so an expensive boot is paid for once and resumed from thereafter, exactly as
// tools/platform/dos does for Ultima Underworld's character creation.
//
// The N64 snapshot is simpler than the DOS one: the cartridge is read-only and
// never written, so there are no host file handles to reopen and the ROM itself
// is not serialised. Its MD5 is, so a snapshot cannot be restored into a
// different game.
//
// Restore is exactly deterministic. The oracle paces interrupts by instruction
// count rather than wall-clock cycles, so resuming a snapshot and running N more
// steps lands on precisely the state an uninterrupted run reaches. That makes a
// savestate a regression instrument and not merely a speed-up — see
// TestSaveStateRoundTrip.
//
// Discipline: every device field added to Machine must be added to snapshot in
// the same commit. A savestate that silently omits a register produces a machine
// that resumes subtly wrong, which is worse than one that refuses to resume. The
// round-trip test is what enforces it.

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/r4300"
)

// snapshotVersion changes whenever the snapshot's shape does, so an old file is
// rejected rather than decoded into the wrong fields.
const snapshotVersion = 1

// snapshot is the fully-serialisable machine state (all fields exported for gob).
type snapshot struct {
	Version int
	ROMMD5  string

	// Memory. The cartridge is omitted: it is read-only, and ROMMD5 pins it.
	RDRAM []byte
	DMEM  []byte
	IMEM  []byte
	PIF   []byte

	CPU r4300.State

	// Devices.
	MI mi
	PI pi

	SPPC uint32
	SP   regFile
	RI   regFile
	RD   regFile
	VI   regFile
	AI   regFile
	SI   regFile
	DP   regFile
}

// SaveState writes the machine's full state to path (gzip-compressed gob).
func (m *Machine) SaveState(path string) error {
	s := &snapshot{
		Version: snapshotVersion,
		ROMMD5:  m.romMD5,
		RDRAM:   append([]byte(nil), m.RDRAM...),
		DMEM:    append([]byte(nil), m.DMEM...),
		IMEM:    append([]byte(nil), m.IMEM...),
		PIF:     append([]byte(nil), m.PIF...),
		CPU:     m.CPU.Snapshot(),
		MI:      m.mi,
		PI:      m.pi,
		SPPC:    m.spPC,
		SP:      m.sp.clone(),
		RI:      m.ri.clone(),
		RD:      m.rd.clone(),
		VI:      m.vi.clone(),
		AI:      m.ai.clone(),
		SI:      m.si.clone(),
		DP:      m.dp.clone(),
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

// LoadState restores a snapshot into this Machine in place, so the CPU's bus and
// the caller's hooks (OnStep, OnWrite, ...) stay wired to it.
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
	if s.Version != snapshotVersion {
		return fmt.Errorf("n64: snapshot version %d, want %d", s.Version, snapshotVersion)
	}
	if s.ROMMD5 != m.romMD5 {
		return fmt.Errorf("n64: snapshot was taken on a different cartridge (md5 %s, this one is %s)",
			s.ROMMD5, m.romMD5)
	}
	if len(s.RDRAM) != len(m.RDRAM) {
		return fmt.Errorf("n64: snapshot RDRAM is %d bytes, this machine has %d", len(s.RDRAM), len(m.RDRAM))
	}

	copy(m.RDRAM, s.RDRAM)
	copy(m.DMEM, s.DMEM)
	copy(m.IMEM, s.IMEM)
	copy(m.PIF, s.PIF)
	m.CPU.Restore(s.CPU)
	m.mi, m.pi = s.MI, s.PI
	m.spPC = s.SPPC
	m.sp, m.ri, m.rd = s.SP, s.RI, s.RD
	m.vi, m.ai, m.si, m.dp = s.VI, s.AI, s.SI, s.DP
	return nil
}

func (r regFile) clone() regFile {
	c := make(regFile, len(r))
	for k, v := range r {
		c[k] = v
	}
	return c
}
