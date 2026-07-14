// Package dsmachine is a two-core Nintendo DS machine model: an ARM9 and an ARM7
// executing on tools/arm cores over one shared 4 MiB main RAM, wired together by
// the DS's inter-processor communication hardware (IPCSYNC + the two IPC FIFOs)
// and a per-core interrupt controller. It is the "full machine" the bare CPU core
// leaves to its caller — enough of the memory map, TCM overlays, WRAM and BIOS
// SWIs to run both boot chains together and drive the ARM9 past the ARM9↔ARM7
// rendezvous that a single core cannot clear.
//
// It models the programmer-visible machine, not the hardware: no cache/MPU timing,
// no video/sound/SPI silicon. Registers those subsystems expose are stubbed to
// keep the boot progressing (status bits read "ready/idle"), and the choices are
// logged so the model stays honest. It is game-neutral — usable by any DS title.
package dsmachine

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/platform/nds"
)

// Address-space landmarks (byte addresses on the DS bus).
const (
	mainBase = 0x02000000
	mainSize = 0x00400000 // 4 MiB shared main RAM
	mainEnd  = mainBase + mainSize

	swramBase = 0x03000000 // shared WRAM block, mirrored up to 0x037FFFFF
	swramEnd  = 0x03800000
	swramSize = 0x8000 // 32 KiB, mirrored every 0x8000

	wram7Base = 0x03800000 // ARM7-private WRAM, 64 KiB
	wram7Size = 0x10000
	wram7End  = wram7Base + wram7Size

	itcmDefault = 0x01FF8000 // ARM9 ITCM window (fast code just below main RAM)
	itcmSize    = 0x8000

	ioBase = 0x04000000
)

// Machine holds the shared state both cores see and the two cores themselves.
type Machine struct {
	ram   []byte // shared main RAM (mainBase..mainEnd)
	swram []byte // shared WRAM (32 KiB, mirrored)

	ipc   ipc
	ARM9  *core
	ARM7  *core
	Steps uint64

	// Log records every stubbed hardware access the model chose to satisfy, so a
	// run's assumptions are auditable rather than silent.
	Log      []string
	logSeen  map[string]bool
	vcount   uint16
	vblankAt uint64 // next step count at which to raise a synthetic VBlank

	// wramcnt is the shared-WRAM split (WRAMCNT, 0x04000247 — ARM9-only). 0 at reset,
	// which gives the ARM9 all 32 KiB and leaves the ARM7's window mirroring its own
	// WRAM. See bus.wramSlot.
	wramcnt uint8

	// SyncTrace, if set, is called on every IPCSYNC nibble write (for diagnostics).
	SyncTrace func(core string, nibble uint8, pc uint32)

	// OnStep, if set, is called before every instruction on either core — the seam a
	// tracer, a breakpoint or a frame debugger hangs off.
	OnStep func(arm9 bool, pc uint32)

	// OnIO, if set, observes every memory-mapped I/O access either core makes.
	OnIO func(arm9 bool, write bool, addr, val uint32, pc uint32)

	// OnWrite, if set, observes every memory write either core makes.
	OnWrite func(arm9 bool, addr uint32, v byte, pc uint32)

	// OnIRQ, if set, observes every interrupt the model dispatches.
	OnIRQ   func(arm9 bool, sources, handler, ret uint32)
	visited map[uint32]bool // ARM9 code pages entered (progress watchdog)
}

// core is one processor: its CPU, its private memory, and its interrupt state.
type core struct {
	m    *Machine
	cpu  *arm.CPU
	name string
	arm9 bool

	// private memories
	itcm     []byte // ARM9 only
	itcmBase uint32
	dtcm     []byte // ARM9 only
	dtcmBase uint32
	low      []byte // ARM7 BIOS/low vectors (0..0x4000)
	wram7    []byte // ARM7-private WRAM (64 KiB at 0x03800000)

	// interrupt controller (memory-mapped in the I/O block)
	ime bool
	ie  uint32
	if_ uint32

	// BIOS wait state: when >0 the core is halted in an IntrWait/Halt SWI, waiting
	// for (checkFlag & waitMask); handlerBase locates the BIOS IRQ vector table.
	waiting     bool
	waitMask    uint32
	waitAny     bool // Halt (SWI 6): wake on any interrupt
	handlerBase uint32
	io          map[uint32]uint32 // last value written to misc I/O (for reads back)
	ioseq       []uint32
	lastRecv    uint32 // last word popped from the recv FIFO
	resumePC    uint32 // PC to resume at after an IntrWait/Halt SWI
	sleep       int    // WaitByLoop budget: skip this core while >0 (lets the other run)
}

// ipc is the shared IPCSYNC mailbox and the two directional FIFOs.
type ipc struct {
	sync9 uint8    // ARM9's outgoing 4-bit nibble (ARM7 reads it as its input)
	sync7 uint8    // ARM7's outgoing nibble
	cnt9  uint16   // ARM9 IPCFIFOCNT
	cnt7  uint16   // ARM7 IPCFIFOCNT
	to7   []uint32 // ARM9→ARM7 FIFO
	to9   []uint32 // ARM7→ARM9 FIFO
}

// New builds the machine from a parsed ROM: it loads the compressed ARM9 at its
// RAM address (the crt0 self-decompresses) and the ARM7 image at its own address,
// each into the memories its core sees. dtcm9Base is the ARM9 DTCM base the game
// programs via CP15 (e.g. 0x023C0000); pass 0 to route DTCM to shared RAM.
func New(rom *nds.ROM, dtcm9Base uint32) *Machine {
	m := &Machine{
		ram:     make([]byte, mainSize),
		swram:   make([]byte, swramSize),
		logSeen: map[string]bool{},
	}
	m.ipc.cnt9, m.ipc.cnt7 = 0x0101, 0x0101 // both send FIFOs empty at reset

	// WRAMCNT = 3: the whole 32 KiB shared block belongs to the ARM7. This is the state
	// the cartridge boot leaves behind, and it is not an arbitrary default — it is what
	// makes the ARM7's memory map make sense at all. The DS boot loads most ARM7
	// binaries at 0x037F8000, and SM64DS's ARM7 relocates itself into one CONTIGUOUS
	// run from 0x037F8000 to 0x03809903 (~72 KiB). That only holds if the shared block's
	// last mirror (0x037F8000..0x037FFFFF) sits directly below the ARM7's private 64 KiB
	// at 0x03800000 — 96 KiB of contiguous WRAM, which is exactly the layout NitroSDK
	// links its ARM7 for.
	//
	// Boot with the block unassigned instead and that copy silently wraps around a 64 KiB
	// mirror, laying the back half of the ARM7's own code over its front half. It then
	// runs off into the data it just corrupted, and nothing about the crash points here.
	m.wramcnt = 3

	// ARM9: private ITCM + DTCM, high BIOS vectors.
	a9 := &core{m: m, name: "ARM9", arm9: true, io: map[uint32]uint32{},
		itcm: make([]byte, itcmSize), itcmBase: itcmDefault,
		handlerBase: 0}
	if dtcm9Base != 0 {
		a9.dtcm = make([]byte, 0x4000)
		a9.dtcmBase = dtcm9Base
		a9.handlerBase = dtcm9Base + 0x4000 // BIOS IRQ vectors at DTCM top
	}
	a9.cpu = arm.NewCPU(&bus{c: a9})
	a9.cpu.Mode = arm.ModeSVC
	a9.cpu.R[15] = rom.Header.ARM9Entry
	a9.cpu.SWI = biosSWI(a9)
	a9.cpu.Coproc = cp15(a9)
	// load the (compressed) ARM9 image; the crt0 decompresses it in place
	copyInto(m, a9, rom.Header.ARM9RAMAddr, rom.ARM9())
	m.ARM9 = a9

	// ARM7: private low + WRAM, low BIOS vectors, handler pointer at WRAM top.
	a7 := &core{m: m, name: "ARM7", arm9: false, io: map[uint32]uint32{},
		low: make([]byte, 0x4000), wram7: make([]byte, wram7Size), handlerBase: wram7End}
	a7.cpu = arm.NewCPU(&bus{c: a7})
	a7.cpu.Mode = arm.ModeSVC
	a7.cpu.R[15] = rom.Header.ARM7Entry
	a7.cpu.SWI = biosSWI(a7)
	copyInto(m, a7, rom.Header.ARM7RAMAddr, rom.ARM7())
	m.ARM7 = a7

	return m
}

// copyInto writes a loaded binary through a core's bus (so it lands in whichever
// memory the address maps to).
func copyInto(m *Machine, c *core, addr uint32, data []byte) {
	b := &bus{c: c}
	for i, v := range data {
		b.Write(addr+uint32(i), v)
	}
}

// ARM9PC and ARM7PC report each core's current program counter.
func (m *Machine) ARM9PC() uint32 { return m.ARM9.cpu.R[15] }
func (m *Machine) ARM7PC() uint32 { return m.ARM7.cpu.R[15] }

// SyncNibbles returns the two cores' outgoing IPCSYNC nibbles (the handshake state).
func (m *Machine) SyncNibbles() (arm9, arm7 uint8) { return m.ipc.sync9, m.ipc.sync7 }

// FifoLens returns the depths of the two IPC FIFOs (ARM9→ARM7, ARM7→ARM9).
func (m *Machine) FifoLens() (to7, to9 int) { return len(m.ipc.to7), len(m.ipc.to9) }

// StepsNow reports the scheduler round count so far.
func (m *Machine) StepsNow() uint64 { return m.Steps }

// Snapshot copies n bytes from addr as the given core sees them (for disassembling
// relocated/decompressed code the run produced).
func (m *Machine) Snapshot(arm9 bool, addr, n uint32) []byte {
	c := m.ARM7
	if arm9 {
		c = m.ARM9
	}
	b := &bus{c: c}
	out := make([]byte, n)
	for i := range out {
		out[i] = b.Read(addr + uint32(i))
	}
	return out
}

// Poke writes n bytes at addr as the given core sees them — an experiment hook (and
// what a debugger's memory editor needs).
func (m *Machine) Poke(arm9 bool, addr uint32, data []byte) {
	c := m.ARM7
	if arm9 {
		c = m.ARM9
	}
	b := &bus{c: c}
	for i, v := range data {
		b.Write(addr+uint32(i), v)
	}
}

// Parked reports whether a core is idle in an interrupt wait (for diagnostics).
func (m *Machine) Parked(arm9 bool) bool {
	if arm9 {
		return m.ARM9.waiting
	}
	return m.ARM7.waiting
}

func (m *Machine) note(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	if !m.logSeen[s] {
		m.logSeen[s] = true
		m.Log = append(m.Log, s)
	}
}
