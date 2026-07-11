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
	TTY          []byte
	SyscallCalls map[string]int
}

type kobjectState struct {
	Kind  string
	Name  string
	Entry uint32
	Addr  uint32
}

// SaveState captures the machine.
func (m *Machine) SaveState() MachineState {
	names := make(map[uint32]string, len(m.syscalls))
	for code, sc := range m.syscalls {
		names[code] = sc.name
	}
	handles := make(map[uint32]kobjectState, len(m.handles))
	for h, o := range m.handles {
		handles[h] = kobjectState{o.kind, o.name, o.entry, o.addr}
	}
	return MachineState{
		ImageHash: m.imageHash,
		RAM:       append([]byte(nil), m.ram...),
		VRAM:      append([]byte(nil), m.vram...),
		Scratch:   append([]byte(nil), m.scratch...),
		CPU:       m.CPU.SaveState(),
		IO:        m.io, NextSyscall: m.nextSyscall, SyscallNames: names,
		Handles: handles, NextHandle: m.nextHandle,
		HeapPtr: m.heapPtr, HeapEnd: m.heapEnd, ThreadEntry: m.threadEntry,
		FBAddr: m.fbAddr, TTY: m.tty, SyscallCalls: m.SyscallCalls,
	}
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
	m.io = s.IO
	m.nextSyscall = s.NextSyscall
	m.syscalls = make(map[uint32]*syscall, len(s.SyscallNames))
	for code, name := range s.SyscallNames {
		m.syscalls[code] = &syscall{name: name, handler: handlerFor(name)}
	}
	m.handles = make(map[uint32]*kobject, len(s.Handles))
	for h, o := range s.Handles {
		m.handles[h] = &kobject{kind: o.Kind, name: o.Name, entry: o.Entry, addr: o.Addr}
	}
	m.nextHandle = s.NextHandle
	m.heapPtr, m.heapEnd, m.threadEntry = s.HeapPtr, s.HeapEnd, s.ThreadEntry
	m.fbAddr, m.tty, m.SyscallCalls = s.FBAddr, s.TTY, s.SyscallCalls
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
