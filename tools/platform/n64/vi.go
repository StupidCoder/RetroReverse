package n64

// vi.go is the Video Interface: it scans a framebuffer out of RDRAM to the
// display, and raises an interrupt once per field when the beam reaches a
// chosen halfline.
//
// That interrupt is the machine's heartbeat. libultra's main thread blocks on
// osRecvMesg from a retrace message queue, so with no VI interrupt every thread
// eventually sleeps and only the idle thread runs — a `b .` loop that a boot
// trace mistakes for a crash. Nothing else in the console will restart it.
//
// The beam is paced by instruction count rather than cycles, as the PSX oracle
// paces its vertical blank. That keeps a restored save-state exactly
// deterministic, which is worth more here than a faithful clock: nothing in a
// game can observe the difference except by timing a loop against the beam, and
// the ratio below is chosen so a frame's worth of code runs per field.

const (
	viStatus  = 0x00 // control: bit depth, AA mode, interlace
	viOrigin  = 0x04 // physical address of the framebuffer being scanned out
	viWidth   = 0x08 // framebuffer width in pixels
	viIntr    = 0x0C // the halfline at which to raise the interrupt
	viCurrent = 0x10 // the halfline being scanned; a write acknowledges the interrupt
	viBurst   = 0x14
	viVSync   = 0x18 // halflines per field
	viHSync   = 0x1C
	viLeap    = 0x20
	viHStart  = 0x24
	viVStart  = 0x28
	viVBurst  = 0x2C
	viXScale  = 0x30
	viYScale  = 0x34
)

// viStatusType selects the framebuffer pixel format.
const (
	viTypeBlank    = 0 // no data, no sync
	viTypeReserved = 1
	viTypeRGBA16   = 2
	viTypeRGBA32   = 3
)

// stepsPerField paces the synthetic retrace. The VR4300 runs at 93.75 MHz and an
// NTSC field lasts ~1.56 M cycles, but the oracle counts executed instructions,
// not cycles, and the CPU averages well over one cycle per instruction (uncached
// RDRAM access dominates). A field's worth of code is therefore a good deal
// fewer instructions than cycles.
const stepsPerField = 750000

// halflinesPerField is the NTSC default; libultra reprograms VI_V_SYNC, and
// V_CURRENT is reported against whatever it wrote.
const halflinesPerField = 525

// Fields are exported so encoding/gob carries them into a save-state.
type vi struct {
	Regs    regFile
	Acc     uint64 // instructions executed since the last retrace
	Current uint32 // the halfline last reported through VI_CURRENT
}

func (v *vi) init() { v.Regs = regFile{} }

func (m *Machine) viRead(addr uint32) uint32 {
	switch addr & 0xFF {
	case viCurrent:
		// The beam position, interpolated across the field from the step
		// accumulator, so code that polls V_CURRENT (rather than waiting on the
		// interrupt) sees it advance. Bit 0 is the field in an interlaced mode.
		lines := m.vi.Regs[viVSync]
		if lines == 0 {
			lines = halflinesPerField
		}
		return uint32(m.vi.Acc*uint64(lines)/stepsPerField) % lines
	}
	return m.vi.Regs[addr&0xFF]
}

func (m *Machine) viWrite(addr uint32, v uint32) {
	switch addr & 0xFF {
	case viCurrent:
		// Any write acknowledges the retrace interrupt; the value is ignored.
		m.clearIRQ(intrVI)
		return
	case viOrigin:
		m.vi.Regs[viOrigin] = v & 0x00FFFFFF
		return
	}
	m.vi.Regs[addr&0xFF] = v
}

// tickVI advances the beam and raises the retrace interrupt once per field.
func (m *Machine) tickVI() {
	m.vi.Acc++
	if m.vi.Acc < stepsPerField {
		return
	}
	m.vi.Acc = 0
	m.vi.Current = m.vi.Regs[viIntr]
	m.raiseIRQ(intrVI)
	if m.OnDisplay != nil {
		m.OnDisplay(m)
	}
}

// Origin is the physical address of the framebuffer the VI is scanning out, and
// Width its pixel width — what a screenshot reads.
func (m *Machine) Origin() uint32 { return m.vi.Regs[viOrigin] }
func (m *Machine) Width() uint32  { return m.vi.Regs[viWidth] }

// PixelType reports the framebuffer format the VI is configured for.
func (m *Machine) PixelType() uint32 { return m.vi.Regs[viStatus] & 3 }
