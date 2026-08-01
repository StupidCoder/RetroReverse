// Package gbamachine is a Game Boy Advance: one 16.78 MHz ARM7TDMI (tools/cpu/arm)
// over the flat AGB bus — 256 KiB EWRAM, 32 KiB IWRAM, the PPU with its palette,
// VRAM and OAM, four DMA channels, four timers, the keypad, the interrupt
// controller, and the cartridge with its serial EEPROM. It is the single-core
// sibling of tools/platform/nds/dsmachine, and deliberately shares its shape: the
// scanline scheduler, the BIOS-shim interrupt delivery, the HLE'd BIOS SWIs, the
// honest unimplemented-hardware log, and savestates from day one.
//
// There is no BIOS image. Like the DS's, the GBA BIOS is a library, not a kernel —
// SWIs (bios.go) and the interrupt dispatch shim (run.go) are reimplemented in Go.
//
// The model is honest about its gaps: an I/O register it does not implement is
// logged (Machine.Log) instead of reading back the last value written, because a
// stub that happens to read "ready" is indistinguishable from working hardware
// right up until the frame it isn't.
package gbamachine

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/platform/gba"
)

// Address-space landmarks (byte addresses on the AGB bus).
const (
	biosSize  = 0x4000
	ewramBase = 0x02000000
	ewramSize = 0x40000 // 256 KiB, mirrored across 0x02XXXXXX
	iwramBase = 0x03000000
	iwramSize = 0x8000 // 32 KiB, mirrored across 0x03XXXXXX
	ioBase    = 0x04000000
	palBase   = 0x05000000
	palSize   = 0x400 // 1 KiB, mirrored
	vramBase  = 0x06000000
	vramSize  = 0x18000 // 96 KiB, mirrored in 128 KiB steps (64K + 32K + 32K-again)
	oamBase   = 0x07000000
	oamSize   = 0x400 // 1 KiB, mirrored
	romBase   = 0x08000000

	// The BIOS leaves these behind in top-of-IWRAM: the game's IRQ handler pointer
	// and, two words below it, the "IRQ check" flags the game's handler reports
	// serviced interrupts in (IntrWait polls them). Same layout the DS inherited.
	irqHandlerSlot = 0x03007FFC
	irqCheckFlags  = 0x03007FF8

	screenW = 240
	screenH = 160
)

// Interrupt sources (bits in IE/IF).
const (
	irqVBlank  = 1 << 0
	irqHBlank  = 1 << 1
	irqVCount  = 1 << 2
	irqTimer0  = 1 << 3
	irqSerial  = 1 << 7
	irqDMA0    = 1 << 8
	irqKeypad  = 1 << 12
	irqGamePak = 1 << 13
)

// Machine is the whole console.
type Machine struct {
	cpu *arm.CPU
	rom []byte

	ewram []byte
	iwram []byte
	pal   []byte
	vram  []byte
	oam   []byte

	io map[uint32]uint16 // 16-bit I/O register file: the last value written

	// Interrupt controller.
	ime bool
	ie  uint16
	if_ uint16

	// BIOS wait state: parked in Halt/IntrWait until a wanted interrupt arrives.
	waiting  bool
	waitAny  bool   // Halt: wake on any enabled interrupt
	waitMask uint16 // IntrWait: resume only when the handler reports one of these

	vid    video
	ppu    ppu
	apu    *apu
	dma    [4]dmaChan
	timers [4]timer
	eeprom eeprom

	keys uint16 // pressed-button mask (1 = held), inverted into KEYINPUT

	Steps uint64

	// screen is the composed frame, one uint32 0xAARRGGBB per pixel, updated a
	// scanline at a time as the PPU emits them.
	screen [screenW * screenH]uint32

	// Log records every unmodelled hardware access, so a run's assumptions are
	// auditable rather than silent.
	Log     []string
	logSeen map[string]bool
	visited map[uint32]bool // 256-byte code pages entered (progress watchdog)

	// OnStep, if set, is called before every instruction — the seam a tracer, a
	// breakpoint or a frame debugger hangs off.
	OnStep func(pc uint32)
	// OnIO observes every memory-mapped I/O register access.
	OnIO func(write bool, addr uint32, val uint16, pc uint32)
	// OnIRQ observes every interrupt the model dispatches.
	OnIRQ func(sources uint16, handler, ret uint32)
	// OnFrame is called once per frame at the vertical blank — the moment a
	// screenshot is taken.
	OnFrame func()

	// Debugger control.
	bps       map[uint32]bool
	stop      bool
	stopped   bool
	stoppedPC uint32
}

// video is the display's position in its raster.
type video struct {
	line   int // 0..227; 160..227 is the vertical blank
	hblank bool
	frames uint64
}

// New builds the machine from a cartridge image in the state the BIOS's cart boot
// leaves behind: System mode, the three stacks the BIOS sets, IRQs enabled at the
// CPU, PC at the cartridge entry point.
func New(rom *gba.ROM) *Machine {
	m := &Machine{
		rom:     rom.Data,
		ewram:   make([]byte, ewramSize),
		iwram:   make([]byte, iwramSize),
		pal:     make([]byte, palSize),
		vram:    make([]byte, vramSize),
		oam:     make([]byte, oamSize),
		io:      map[uint32]uint16{},
		logSeen: map[string]bool{},
	}
	m.eeprom.init()
	m.apu = newAPU()
	m.cpu = arm.NewCPU(&bus{m: m})
	// The ARM7TDMI is ARMv4T, a strict SUBSET of the DS-era default: no BLX,
	// CLZ, LDRD/STRD, PLD, the saturating group or the SMLAxy multiplies. Left
	// at the default a v5-only word would quietly execute as something this
	// console cannot do (arm/variant.go).
	m.cpu.Arch = arm.V4T
	m.cpu.SWI = biosSWI(m)

	// The BIOS hands over in System mode with IRQs enabled and stacks planted:
	// sp_svc = 0x03007FE0, sp_irq = 0x03007FA0, sp_sys = 0x03007F00. The crt0
	// re-plants them (Minish Cap does), but a game is entitled to inherit these.
	c := m.cpu
	c.SetCPSR((c.CPSR() &^ 0x1F) | arm.ModeIRQ)
	c.R[13] = 0x03007FA0
	c.SetCPSR((c.CPSR() &^ 0x1F) | arm.ModeSVC)
	c.R[13] = 0x03007FE0
	c.SetCPSR((c.CPSR() &^ 0x1F) | arm.ModeSYS)
	c.R[13] = 0x03007F00
	c.IRQDisable, c.FIQDisable = false, true
	c.R[15] = romBase

	m.io[0x88] = 0x200 // SOUNDBIAS resets to midpoint
	return m
}

// note records an unmodelled access once.
func (m *Machine) note(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	if !m.logSeen[s] {
		m.logSeen[s] = true
		m.Log = append(m.Log, s)
	}
}

// --- introspection (the oracle's diagnostics) --------------------------------

// PC reports the CPU's current program counter.
func (m *Machine) PC() uint32 { return m.cpu.R[15] }

// Regs returns the 16 general-purpose registers (R15 is the PC).
func (m *Machine) Regs() [16]uint32 { return m.cpu.R }

// ThumbState reports whether the CPU is executing Thumb code.
func (m *Machine) ThumbState() bool { return m.cpu.Thumb }

// IRQState reports the interrupt controller: enable mask, pending flags, master
// enable. The first thing to look at when a boot stops making progress.
func (m *Machine) IRQState() (ie, if_ uint16, ime bool) { return m.ie, m.if_, m.ime }

// IRQDisabled reports the CPSR I bit — the third place an interrupt can be stopped.
func (m *Machine) IRQDisabled() bool { return m.cpu.IRQDisable }

// Parked reports whether the CPU is idle in a BIOS interrupt wait.
func (m *Machine) Parked() string {
	switch {
	case !m.waiting:
		return "running"
	case m.waitAny:
		return "halted for IRQ"
	default:
		return fmt.Sprintf("IntrWait 0x%X", m.waitMask)
	}
}

// Frame reports how many frames the display has completed, and Line its scanline.
func (m *Machine) Frame() uint64 { return m.vid.frames }
func (m *Machine) Line() int     { return m.vid.line }

// Halted reports whether (and why) the CPU hit something unimplemented.
func (m *Machine) Halted() (bool, string) { return m.cpu.Halted, m.cpu.HaltReason }

// Instrs reports how many instructions the CPU has executed.
func (m *Machine) Instrs() uint64 { return m.cpu.Instrs }

// Reg reads an I/O register's last-written value — a diagnostic, not a bus read:
// it does not run the register's side effects.
func (m *Machine) Reg(a uint32) uint16 { return m.io[a&0xFFFF&^1] }

// Snapshot copies n bytes from addr as the CPU sees them.
func (m *Machine) Snapshot(addr, n uint32) []byte {
	b := &bus{m: m}
	out := make([]byte, n)
	for i := range out {
		out[i] = b.Read(addr + uint32(i))
	}
	return out
}

// Poke writes bytes at addr as the CPU sees them — an experiment hook.
func (m *Machine) Poke(addr uint32, data []byte) {
	b := &bus{m: m}
	for i, v := range data {
		b.Write(addr+uint32(i), v)
	}
}

// SetKeys sets the pressed-button mask: bit 0 A, 1 B, 2 Select, 3 Start, 4 Right,
// 5 Left, 6 Up, 7 Down, 8 R, 9 L (1 = held; KEYINPUT inverts).
func (m *Machine) SetKeys(mask uint16) { m.keys = mask & 0x3FF }

// AddBreakpoint arms a halting breakpoint.
func (m *Machine) AddBreakpoint(pc uint32) {
	if m.bps == nil {
		m.bps = map[uint32]bool{}
	}
	m.bps[pc] = true
}

// GameCode reports the four-character cartridge code.
func (m *Machine) GameCode() string {
	if len(m.rom) < 0xB0 {
		return ""
	}
	return string(m.rom[0xAC:0xB0])
}
