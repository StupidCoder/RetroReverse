// Package dsmachine is a Nintendo DS: two ARM cores (a 66 MHz ARM946E-S and a
// 33 MHz ARM7TDMI, both on tools/cpu/arm) sharing 4 MiB of main RAM, wired together
// by the DS's inter-processor hardware, and surrounded by the silicon a DS game
// actually talks to — the display controller and its timing, eight DMA channels,
// eight timers, the ARM9's hardware divider and square-root unit, the nine VRAM
// banks and the mapping register that decides what each of them currently *is*,
// the cartridge port, the ARM7's SPI bus (firmware, power, touchscreen), and both
// graphics engines: the two 2D engines and the 3D geometry/rasterising pipeline.
//
// It is a low-level emulation, like tools/platform/n64 and tools/platform/psx and
// unlike tools/platform/n3ds: there is no operating system to high-level-emulate on
// a DS, only hardware. The one thing lifted above the metal is the BIOS's software
// interrupts (bios.go) — the memory copies, the decompressors and the interrupt
// waits — which are HLE'd the way the PSX BIOS is, because they are a library, not
// a kernel.
//
// The model is honest about its gaps rather than quiet about them: an I/O register
// it does not implement is logged (Machine.Log) instead of reading back the last
// value written, because on a machine whose boot polls status bits constantly, a
// stub that happens to read "ready" is indistinguishable from working hardware
// right up until the frame it isn't.
package dsmachine

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/platform/nds"
)

// Address-space landmarks (byte addresses on the DS bus).
const (
	mainBase = 0x02000000
	mainSize = 0x00400000 // 4 MiB shared main RAM...
	mainEnd  = mainBase + mainSize
	// ...mirrored every 4 MiB across the whole region. See bus.slot: the DS's boot
	// parameter block is addressed through the top mirror (0x027FFxxx), so this is
	// load-bearing, not decoration.
	mainMirrorEnd = 0x03000000

	swramBase = 0x03000000 // shared WRAM block, mirrored up to 0x037FFFFF
	swramEnd  = 0x03800000
	swramSize = 0x8000 // 32 KiB, mirrored every 0x8000

	wram7Base = 0x03800000 // ARM7-private WRAM, 64 KiB
	wram7Size = 0x10000
	wram7End  = wram7Base + wram7Size

	itcmDefault = 0x01FF8000 // ARM9 ITCM window (fast code just below main RAM)
	itcmSize    = 0x8000

	palBase = 0x05000000 // palette RAM, 2 KiB, mirrored
	palSize = 0x800
	oamBase = 0x07000000 // object attribute memory, 2 KiB, mirrored
	oamSize = 0x800
)

// Machine holds the shared state both cores see and the two cores themselves.
type Machine struct {
	ram   []byte // shared main RAM
	swram []byte // shared WRAM (32 KiB, mirrored)
	pal   []byte // palette RAM (engine A BG/OBJ, then engine B BG/OBJ)
	oam   []byte // object attribute memory (engine A, then engine B)

	vram  *vram
	cd    *card
	spi   *spibus
	gpu2d *gpu2d
	gpu3d *gpu3d

	vid  video
	div  divider
	sqrt sqrter

	powcnt  uint32
	keys    uint32
	wramcnt uint8 // the shared-WRAM split (WRAMCNT); see bus.wramSlot

	ipc   ipc
	ARM9  *core
	ARM7  *core
	Steps uint64

	// Log records every unmodelled hardware access the model met, so a run's
	// assumptions are auditable rather than silent.
	Log     []string
	logSeen map[string]bool
	visited map[uint32]bool // ARM9 code pages entered (progress watchdog)

	// SyncTrace, if set, is called on every IPCSYNC nibble write.
	SyncTrace func(core string, nibble uint8, pc uint32)

	// OnStep, if set, is called before every instruction on either core — the seam a
	// tracer, a breakpoint or a frame debugger hangs off.
	OnStep func(arm9 bool, pc uint32)

	// OnIO observes every memory-mapped I/O access either core makes.
	OnIO func(arm9 bool, write bool, addr, val, pc uint32)

	// OnWrite observes every memory write either core makes.
	OnWrite func(arm9 bool, addr uint32, v byte, pc uint32)

	// OnIRQ observes every interrupt the model dispatches.
	OnIRQ func(arm9 bool, sources, handler, ret uint32)

	// OnFrame, if set, is called once per frame at the vertical blank — after the 3D
	// engine has swapped its buffers, and before the CPUs run the new frame. This is
	// the moment a screenshot or a frame-debugger capture is taken.
	OnFrame func()
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

	dma    [4]dmaChan
	timers [4]timer

	// interrupt controller (memory-mapped in the I/O block)
	ime bool
	ie  uint32
	if_ uint32

	// BIOS wait state: when set, the core is halted in an IntrWait/Halt SWI, waiting
	// for a matching interrupt; handlerBase locates the BIOS IRQ vector slots.
	waiting     bool
	waitMask    uint32
	waitAny     bool // Halt (SWI 6): wake on any interrupt
	handlerBase uint32
	io          map[uint32]uint32 // the register file (and the last value written)
	lastRecv    uint32            // last word popped from the recv FIFO

	sleep int // WaitByLoop budget: skip this core while > 0
}

// ipc is the shared IPCSYNC mailbox and the two directional FIFOs.
type ipc struct {
	sync9 uint8    // ARM9's outgoing 4-bit nibble (the ARM7 reads it as its input)
	sync7 uint8    // ARM7's outgoing nibble
	to7   []uint32 // ARM9→ARM7 FIFO
	to9   []uint32 // ARM7→ARM9 FIFO
}

// New builds the machine from a parsed ROM: it loads the compressed ARM9 at its RAM
// address (the crt0 self-decompresses) and the ARM7 image at its own, each into the
// memories its core sees. dtcm9Base is the ARM9 DTCM base the game programs via
// CP15; pass 0 to route DTCM to shared RAM.
func New(rom *nds.ROM, dtcm9Base uint32) *Machine {
	m := &Machine{
		ram:     make([]byte, mainSize),
		swram:   make([]byte, swramSize),
		pal:     make([]byte, palSize),
		oam:     make([]byte, oamSize),
		vram:    newVRAM(),
		spi:     newSPI(),
		cd:      &card{rom: rom.Data},
		logSeen: map[string]bool{},
	}
	m.gpu2d = newGPU2D()
	m.gpu3d = newGPU3D()
	m.vid.line = linesPerFrame - 1 // the first startLine() wraps it to line 0

	// WRAMCNT = 3: the whole 32 KiB shared block belongs to the ARM7. This is the
	// state the cartridge boot leaves behind, and it is not an arbitrary default — it
	// is what makes the ARM7's memory map make sense at all. The DS boot loads most
	// ARM7 binaries at 0x037F8000, and SM64DS's ARM7 relocates itself into one
	// CONTIGUOUS run from 0x037F8000 to 0x03809903 (~72 KiB). That only holds if the
	// shared block's last mirror (0x037F8000..0x037FFFFF) sits directly below the
	// ARM7's private 64 KiB at 0x03800000 — 96 KiB of contiguous WRAM, which is
	// exactly the layout NitroSDK links its ARM7 for.
	//
	// Boot with the block unassigned instead and that copy silently wraps around a
	// 64 KiB mirror, laying the back half of the ARM7's own code over its front half.
	// It then runs off into the data it just corrupted, and nothing about the crash
	// points here.
	m.wramcnt = 3

	// ARM9: private ITCM + DTCM, high BIOS vectors.
	a9 := &core{m: m, name: "ARM9", arm9: true, io: map[uint32]uint32{},
		itcm: make([]byte, itcmSize), itcmBase: itcmDefault}
	if dtcm9Base != 0 {
		a9.dtcm = make([]byte, 0x4000)
		a9.dtcmBase = dtcm9Base
		a9.handlerBase = dtcm9Base + 0x4000 // the BIOS IRQ vectors sit at the DTCM top
	}
	a9.cpu = arm.NewCPU(&bus{c: a9})
	a9.cpu.Mode = arm.ModeSVC
	a9.cpu.R[15] = rom.Header.ARM9Entry
	a9.cpu.SWI = biosSWI(a9)
	a9.cpu.Coproc = cp15(a9)
	copyInto(m, a9, rom.Header.ARM9RAMAddr, rom.ARM9())
	m.ARM9 = a9

	// ARM7: private low memory + WRAM, low BIOS vectors, handler pointer at WRAM top.
	a7 := &core{m: m, name: "ARM7", arm9: false, io: map[uint32]uint32{},
		low: make([]byte, 0x4000), wram7: make([]byte, wram7Size), handlerBase: wram7End}
	a7.cpu = arm.NewCPU(&bus{c: a7})
	a7.cpu.Mode = arm.ModeSVC
	a7.cpu.R[15] = rom.Header.ARM7Entry
	a7.cpu.SWI = biosSWI(a7)
	copyInto(m, a7, rom.Header.ARM7RAMAddr, rom.ARM7())
	m.ARM7 = a7

	m.directBoot(rom) // leave behind what the BIOS cart-boot would have (boot.go)
	return m
}

// copyInto writes a loaded binary through a core's bus, so it lands in whichever
// memory the address maps to.
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

// Frame reports how many frames the display has completed, and Line the scanline
// it is on.
func (m *Machine) Frame() uint64 { return m.vid.frames }
func (m *Machine) Line() int     { return m.vid.line }

// Snapshot copies n bytes from addr as the given core sees them (for disassembling
// the relocated/decompressed code a run produced).
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

// Poke writes bytes at addr as the given core sees them — an experiment hook, and
// what a debugger's memory editor needs.
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

// Regs returns a core's 16 general-purpose registers (R15 is the PC).
func (m *Machine) Regs(arm9 bool) [16]uint32 {
	if arm9 {
		return m.ARM9.cpu.R
	}
	return m.ARM7.cpu.R
}

// Thumb reports whether a core is executing Thumb code — which a disassembler must
// know before it can decode anything at the PC.
func (m *Machine) Thumb(arm9 bool) bool {
	if arm9 {
		return m.ARM9.cpu.Thumb
	}
	return m.ARM7.cpu.Thumb
}

// IRQState reports a core's interrupt controller: the enable mask, the pending
// flags, and the master enable. The first thing to look at when a boot stops making
// progress is whether the interrupt it is waiting for can even be delivered.
func (m *Machine) IRQState(arm9 bool) (ie, if_ uint32, ime bool) {
	c := m.ARM7
	if arm9 {
		c = m.ARM9
	}
	return c.ie, c.if_, c.ime
}

// Parked reports whether a core is idle in an interrupt wait (for diagnostics).
func (m *Machine) Parked(arm9 bool) bool {
	if arm9 {
		return m.ARM9.waiting
	}
	return m.ARM7.waiting
}

func (m *Machine) onFrame() {
	if m.OnFrame != nil {
		m.OnFrame()
	}
}

func (m *Machine) note(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	if !m.logSeen[s] {
		m.logSeen[s] = true
		m.Log = append(m.Log, s)
	}
}

// IRQDisabled reports a core's CPSR I bit — whether it is running with interrupts
// masked at the processor, which is a different thing from IME and from IE, and is
// the third place an interrupt can be silently stopped.
func (m *Machine) IRQDisabled(arm9 bool) bool {
	if arm9 {
		return m.ARM9.cpu.IRQDisable
	}
	return m.ARM7.cpu.IRQDisable
}

// Reg reads an I/O register's last-written value from the ARM9's register file — a
// diagnostic, not a bus read: it does not run the register's side effects.
func (m *Machine) Reg(a uint32) uint32 { return m.ARM9.io[a&^3] }

// Sleep reports how much of a WaitByLoop delay a core still owes — a diagnostic for
// the case where a core seems alive but is barely executing.
func (m *Machine) Sleep(arm9 bool) int {
	if arm9 {
		return m.ARM9.sleep
	}
	return m.ARM7.sleep
}

// OnCardXfer installs a hook reporting every cartridge transfer the game performs:
// the command, the ROM address, and the size. The map of what the game loads, when —
// drawn by its own loader rather than guessed.
func (m *Machine) OnCardXfer(f func(cmd [8]byte, addr uint32, size int)) { m.cd.OnXfer = f }

// GX reports what the 3D engine did on its last completed frame: how many polygons
// the geometry engine handed the rasteriser at the buffer swap, and how many swaps
// there have been. This is the first number to look at when the screen is black —
// a black frame with polygons is a shading bug, a black frame with none is a geometry
// bug, and they are not investigated the same way.
func (m *Machine) GX() (polys, swaps int) {
	return m.gpu3d.lastPolys, m.gpu3d.swaps
}

// GXHist returns how many times each 3D command has been executed — the DS's
// -gxdump. A frame that draws nothing has a recognisable shape in this table.
func (m *Machine) GXHist() [256]int { return m.gpu3d.cmdHist }

// GXClip reports how many primitives the vertex assembler produced and how many the
// clipper rejected — the two numbers that separate "no geometry arrived" from
// "the geometry arrived and my transform threw it off the screen".
func (m *Machine) GXClip() (emitted, clipped int) {
	return m.gpu3d.geom.nEmit, m.gpu3d.geom.nClipped
}

// Reg7 reads an I/O register's last-written value from the ARM7's register file.
func (m *Machine) Reg7(a uint32) uint32 { return m.ARM7.io[a&^3] }

// GXRegs exposes the 3D engine's register file (clear colour, fog, the toon table).
func (m *Machine) GXRegs() map[uint32]uint32 { return m.gpu3d.regs }

// OnGXCmd installs a hook reporting every 3D command the FIFO decodes, with its
// parameters — the DS's command-list dump.
func (m *Machine) OnGXCmd(f func(cmd uint8, p []uint32)) { m.gpu3d.OnCmd = f }

// VRAMTexPal reads a halfword of the 3D texture-palette space — for inspecting the
// colours a texture actually resolves to.
func (m *Machine) VRAMTexPal(off uint32) uint16 { return m.vram.read16(spTexPal, off) }
