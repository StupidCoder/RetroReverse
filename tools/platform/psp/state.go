package psp

// state.go snapshots the machine so an oracle can save a mid-run state and branch
// experiments from it. The format is gzip+gob (the pattern the other platform oracles
// use); the source image's hash is pinned so a state cannot resume on a different
// game. Host-side config (hooks, the module) is not serialized; the syscall handler
// table is rebuilt from the persisted (code -> name) map on load.

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/allegrex"
)

// MachineState is everything the running game can observe.
type MachineState struct {
	ImageHash string
	RAM       []byte
	VRAM      []byte
	Scratch   []byte
	CPU       allegrex.CPUState

	IO           map[uint32]uint32
	NextSyscall  uint32
	SyscallNames map[uint32]string // code -> name (handlers rebuilt from these)
	Handles      map[uint32]kobjectState
	NextHandle   uint32
	HeapPtr      uint32
	HeapEnd      uint32
	ThreadEntry  uint32
	FBAddr       uint32
	FBWidth      uint32
	FBFormat     uint32
	GE           *geWire // the GE register file; nil before the first display list
	TTY          []byte
	SyscallCalls map[string]int
	SubIntrs     []subIntrState
	VBlanks      uint32
	Current      uint32 // handle of the running thread (0 = the anonymous context)
	Files        map[uint32]ioFileState
	NextFd       uint32
	Pad          uint32
	Savedata     uint32 // savedata-utility dialog status (0 none .. 4 shutdown)
	Mpeg         mpegState
	Atrac        map[uint32]atracState
	NextAtrac    uint32
	VolatileLock bool // the 4 MiB volatile block is held (sceKernelVolatileMemLock)
}

type ioFileState struct {
	Path string
	Pos  int64
}

type subIntrState struct {
	Intno, Subno, Handler, Arg uint32
	Enabled                    bool
}

type kobjectState struct {
	Kind  string
	Name  string
	Entry uint32
	Addr  uint32
	Size  uint32
	Used  uint32
	Bits  uint32

	// Thread fields (sched.go).
	Priority uint32
	StackTop uint32
	Tstate   int
	Ctx      allegrex.CPUState

	WaitEv, WaitBits, WaitMode, WaitOutPtr uint32
	Count                                  int32
	WaitSema                               uint32
	WaitNeed                               int32
	WakeVblank                             uint32
}

// SaveState captures the machine.
func (m *Machine) SaveState() MachineState {
	names := make(map[uint32]string, len(m.syscalls))
	for code, sc := range m.syscalls {
		names[code] = sc.name
	}
	handles := make(map[uint32]kobjectState, len(m.handles))
	var current uint32
	for h, o := range m.handles {
		handles[h] = kobjectState{
			Kind: o.kind, Name: o.name, Entry: o.entry, Addr: o.addr,
			Size: o.size, Used: o.used, Bits: o.bits,
			Priority: o.priority, StackTop: o.stackTop, Tstate: int(o.tstate), Ctx: o.ctx,
			WaitEv: o.waitEv, WaitBits: o.waitBits, WaitMode: o.waitMode, WaitOutPtr: o.waitOutPtr,
			Count: o.count, WaitSema: o.waitSema, WaitNeed: o.waitNeed,
			WakeVblank: o.wakeVblank,
		}
		if o == m.current {
			current = h
		}
	}
	var intrs []subIntrState
	for _, si := range m.subIntrs {
		intrs = append(intrs, subIntrState{si.intno, si.subno, si.handler, si.arg, si.enabled})
	}
	atracs := make(map[uint32]atracState, len(m.atrac))
	for id, a := range m.atrac {
		atracs[id] = *a
	}
	files := make(map[uint32]ioFileState, len(m.files))
	for fd, f := range m.files {
		files[fd] = ioFileState{Path: f.path, Pos: f.pos}
	}
	// The GE register file. It survives across display lists, so it is machine state
	// and belongs here; a restore without it silently falls back to the executor's
	// defaults (see ge_wire.go).
	var ge *geWire
	if m.geSt != nil {
		ge = geWireOf(m.geSt)
	}
	// The maps and slices below are copied, not shared. A snapshot has to be
	// independent of the machine that made it — framedbg restores one repeatedly while
	// the live machine runs on, and an aliased map would let the present rewrite the
	// past. (The file format never noticed, because gob copies on the way out.)
	return MachineState{
		ImageHash: m.imageHash,
		RAM:       append([]byte(nil), m.ram...),
		VRAM:      append([]byte(nil), m.vram...),
		Scratch:   append([]byte(nil), m.scratch...),
		CPU:       m.CPU.SaveState(),
		IO:        copyMap(m.io), NextSyscall: m.nextSyscall, SyscallNames: names,
		Handles: handles, NextHandle: m.nextHandle,
		HeapPtr: m.heapPtr, HeapEnd: m.heapEnd, ThreadEntry: m.threadEntry,
		FBAddr: m.fbAddr, FBWidth: m.fbWidth, FBFormat: m.fbFormat, GE: ge,
		TTY:          append([]byte(nil), m.tty...),
		SyscallCalls: copyMap(m.SyscallCalls),
		SubIntrs:     intrs, VBlanks: m.vblanks, Current: current,
		Files: files, NextFd: m.nextFd,
		Pad: m.pad, Savedata: m.savedataStatus, Mpeg: m.mpeg,
		Atrac: atracs, NextAtrac: m.nextAtrac,
		VolatileLock: m.volatileLocked,
	}
}

// copyMap returns an independent copy of a map, so a snapshot does not alias the live
// machine's.
func copyMap[K comparable, V any](in map[K]V) map[K]V {
	out := make(map[K]V, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// LoadState restores a state, rebuilding the syscall handler table from its names.
func (m *Machine) LoadState(s MachineState) error {
	if m.imageHash != "" && s.ImageHash != "" && m.imageHash != s.ImageHash {
		return fmt.Errorf("psp: savestate image hash %s != current %s", s.ImageHash, m.imageHash)
	}
	copy(m.ram, s.RAM)
	copy(m.vram, s.VRAM)
	copy(m.scratch, s.Scratch)
	m.CPU.LoadState(s.CPU)
	// A halt stops the run BEFORE the faulting instruction retires, so a state
	// saved at a halt resumes at that instruction. Clear the flag: with the
	// cause fixed the run continues; without it, it re-halts on the same word.
	m.CPU.Halted, m.CPU.HaltReason = false, ""
	m.Halted, m.HaltReason = false, ""
	// Copy, do not alias: the same snapshot is restored again and again (every command
	// the frame debugger's scrubber lands on is a fresh replay from one capture), so a
	// map taken by reference would carry the last replay's writes into the next one.
	m.io = copyMap(s.IO)
	m.nextSyscall = s.NextSyscall
	m.syscalls = make(map[uint32]*syscall, len(s.SyscallNames))
	for code, name := range s.SyscallNames {
		// A state may predate a NID becoming known: names were saved in the
		// "library:0xNID" fallback form. Re-resolve so new handlers bind.
		if i := strings.LastIndex(name, ":0x"); i >= 0 {
			if nid, err := strconv.ParseUint(name[i+3:], 16, 32); err == nil {
				if known, ok := nidLookup[uint32(nid)]; ok {
					name = known
				}
			}
		}
		m.syscalls[code] = &syscall{name: name, handler: handlerFor(name)}
	}
	m.handles = make(map[uint32]*kobject, len(s.Handles))
	m.current = nil
	for h, o := range s.Handles {
		ko := &kobject{
			kind: o.Kind, name: o.Name, entry: o.Entry, addr: o.Addr,
			size: o.Size, used: o.Used, bits: o.Bits,
			priority: o.Priority, stackTop: o.StackTop, tstate: threadState(o.Tstate), ctx: o.Ctx,
			waitEv: o.WaitEv, waitBits: o.WaitBits, waitMode: o.WaitMode, waitOutPtr: o.WaitOutPtr,
			count: o.Count, waitSema: o.WaitSema, waitNeed: o.WaitNeed,
			wakeVblank: o.WakeVblank,
		}
		m.handles[h] = ko
		if h == s.Current && h != 0 {
			m.current = ko
		}
	}
	m.nextHandle = s.NextHandle
	m.heapPtr, m.heapEnd, m.threadEntry = s.HeapPtr, s.HeapEnd, s.ThreadEntry
	m.fbAddr, m.fbWidth, m.fbFormat = s.FBAddr, s.FBWidth, s.FBFormat
	m.tty, m.SyscallCalls = append([]byte(nil), s.TTY...), copyMap(s.SyscallCalls)
	// The GE register file, restored as it stood — not rebuilt from defaults.
	if s.GE != nil {
		m.geSt = s.GE.state()
	} else {
		m.geSt = nil // a state taken before the first display list
	}
	m.subIntrs = map[uint32]*subIntr{}
	for _, si := range s.SubIntrs {
		m.subIntrs[si.Intno<<16|si.Subno] = &subIntr{
			intno: si.Intno, subno: si.Subno, handler: si.Handler, arg: si.Arg, enabled: si.Enabled,
		}
	}
	m.vblanks = s.VBlanks
	m.pad, m.savedataStatus, m.mpeg = s.Pad, s.Savedata, s.Mpeg
	m.volatileLocked = s.VolatileLock
	m.atrac = map[uint32]*atracState{}
	for id, a := range s.Atrac {
		a := a
		m.atrac[id] = &a
	}
	m.nextAtrac = s.NextAtrac
	m.files = map[uint32]*ioFile{}
	for fd, f := range s.Files {
		if m.vol == nil {
			m.note("savestate open file %q dropped: no volume mounted", f.Path)
			continue
		}
		e, err := m.vol.resolve(f.Path)
		if err != nil {
			m.note("savestate open file %q dropped: %v", f.Path, err)
			continue
		}
		m.files[fd] = &ioFile{path: f.Path, ent: e, pos: f.Pos}
	}
	if s.NextFd >= fdFirstFile {
		m.nextFd = s.NextFd
	}
	return nil
}

// SaveStateFile writes a gzip+gob snapshot.
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

// LoadStateFile reads a snapshot written by SaveStateFile.
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
	var s MachineState
	if err := gob.NewDecoder(zr).Decode(&s); err != nil {
		return err
	}
	return m.LoadState(s)
}
