package gc

// pi.go is the Processor Interface: the interrupt controller, and the pointers that tell
// the command processor where the graphics FIFO lives.
//
// The interrupt controller is the machine's nervous system. Fourteen sources — the disc,
// the controllers, the video retrace, the graphics pipe — each raise a cause bit, and the
// Gekko's single external-interrupt line is asserted when any raised cause is also
// unmasked. The line is level-sensitive: the machine holds it up until the handler clears
// the cause at the source, which is why a handler that forgets to acknowledge its device
// spins forever taking the same interrupt.

// The interrupt causes, by bit.
const (
	IntError    = 0  // a graphics-processor runtime error
	IntReset    = 1  // the reset button
	IntDI       = 2  // the disc drive
	IntSI       = 3  // the controllers
	IntEXI      = 4  // the memory card, the RTC
	IntAI       = 5  // the streaming audio clock
	IntDSP      = 6  // the sound processor's mailbox
	IntMEM      = 7  // a memory-protection violation
	IntVI       = 8  // the vertical retrace — the frame clock
	IntPEToken  = 9  // the pixel engine reached a token in the command stream
	IntPEFinish = 10 // the pixel engine finished a frame
	IntCP       = 11 // the command-processor FIFO
	IntDebug    = 12
	IntHSP      = 13 // the high-speed port
)

type pi struct {
	Cause     uint32 // INTSR: which sources are asserting
	Mask      uint32 // INTMR: which are allowed through to the CPU
	FIFOBase  uint32 // where the graphics FIFO is in memory
	FIFOEnd   uint32
	FIFOWrite uint32
	ResetCode uint32
}

func (p *pi) read(m *Machine, off uint32, size int) uint32 {
	switch off & 0xFFF {
	case 0x00:
		return p.Cause
	case 0x04:
		return p.Mask
	case 0x0C:
		return p.FIFOBase
	case 0x10:
		return p.FIFOEnd
	case 0x14:
		return p.FIFOWrite
	case 0x24:
		// The console-type register. Bits 28-31 are the revision; software reads it to
		// tell a retail unit from a development one. A retail GameCube reads back this.
		return 0x2020_0006
	case 0x2C:
		return p.ResetCode
	}
	m.logf("PI read unmodelled 0x%03X", off&0xFFF)
	return 0
}

func (p *pi) write(m *Machine, off uint32, v uint32, size int) {
	switch off & 0xFFF {
	case 0x00:
		// Writing INTSR acknowledges: the written 1-bits clear the corresponding causes.
		// This is how a handler dismisses the interrupt it is servicing.
		p.Cause &^= v
		m.updateIRQ()
	case 0x04:
		p.Mask = v
		m.updateIRQ()
	case 0x0C:
		p.FIFOBase = v & 0x03FFFFE0
	case 0x10:
		p.FIFOEnd = v & 0x03FFFFE0
	case 0x14:
		p.FIFOWrite = v & 0x03FFFFE0
	case 0x24:
		// The console type is read-only; a write is ignored.
	case 0x2C:
		p.ResetCode = v
	default:
		m.logf("PI write unmodelled 0x%03X = 0x%08X", off&0xFFF, v)
	}
}

// raiseInt asserts an interrupt cause and updates the CPU's line.
func (m *Machine) raiseInt(cause int) {
	m.pi.Cause |= 1 << cause
	m.updateIRQ()
}

// clearInt lowers a cause. A device calls it when a handler acknowledges it at the source.
func (m *Machine) clearInt(cause int) {
	m.pi.Cause &^= 1 << cause
	m.updateIRQ()
}

// updateIRQ recomputes the Gekko's external-interrupt line from the causes and the mask.
// It is level-sensitive: the line follows (cause & mask), so lowering the last unmasked
// cause lowers the line, and the CPU stops taking the exception.
func (m *Machine) updateIRQ() {
	m.CPU.Interrupt(m.pi.Cause&m.pi.Mask != 0)
}
