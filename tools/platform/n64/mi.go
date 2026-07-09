package n64

// mi.go is the MIPS Interface: the RCP's interrupt aggregator.
//
// Six devices raise interrupts — SP, SI, AI, VI, PI and DP. MI collects them in
// MI_INTR and gates them with MI_INTR_MASK; if any unmasked interrupt is
// pending, MI asserts the CPU's Int0 line, which the VR4300 sees as Cause IP2.
// Every device is therefore invisible to the game unless this indirection is
// modelled exactly.
//
// Each device acknowledges its own interrupt through its own register, not
// through MI, so clearing is spread across the other device files.

// MI registers, as offsets within its range.
const (
	miInitMode = 0x00
	miVersion  = 0x04
	miIntr     = 0x08
	miIntrMask = 0x0C
)

// The six interrupt sources, in MI_INTR bit order.
const (
	intrSP = 1 << 0
	intrSI = 1 << 1
	intrAI = 1 << 2
	intrVI = 1 << 3
	intrPI = 1 << 4
	intrDP = 1 << 5
)

// miVersionValue is what the RCP reports: the revisions of its sub-blocks.
const miVersionValue = 0x02020102

// Fields are exported because the whole struct is serialised into a save-state
// by encoding/gob, which ignores unexported fields — a silently-lost register
// here would resume a machine that had forgotten its pending interrupts.
type mi struct {
	InitMode uint32
	Intr     uint32 // pending interrupts
	Mask     uint32 // which of them reach the CPU
}

// raiseIRQ latches a device interrupt. It becomes visible to the CPU on the next
// interrupt check, if unmasked.
func (m *Machine) raiseIRQ(bit uint32) { m.mi.Intr |= bit }

// clearIRQ acknowledges a device interrupt.
func (m *Machine) clearIRQ(bit uint32) { m.mi.Intr &^= bit }

// irqPending reports whether an unmasked interrupt is asserting Int0.
func (m *Machine) irqPending() bool { return m.mi.Intr&m.mi.Mask != 0 }

func (m *Machine) miRead(addr uint32) uint32 {
	switch addr & 0xFF {
	case miInitMode:
		return m.mi.InitMode
	case miVersion:
		return miVersionValue
	case miIntr:
		return m.mi.Intr
	case miIntrMask:
		return m.mi.Mask
	}
	m.note("MI: read from undecoded register 0x%08X", addr)
	return 0
}

// The mode and mask registers are written with paired set/clear bits rather than
// a plain value, so that a device can enable its own interrupt without a
// read-modify-write racing another.
func (m *Machine) miWrite(addr uint32, v uint32) {
	switch addr & 0xFF {
	case miInitMode:
		m.mi.InitMode = (m.mi.InitMode &^ 0x7F) | (v & 0x7F)
		if v&(1<<11) != 0 { // clear DP interrupt
			m.clearIRQ(intrDP)
		}
	case miIntrMask:
		// Bits come in pairs: clear at 2n, set at 2n+1, for the six sources.
		for i := uint32(0); i < 6; i++ {
			if v&(1<<(2*i)) != 0 {
				m.mi.Mask &^= 1 << i
			}
			if v&(1<<(2*i+1)) != 0 {
				m.mi.Mask |= 1 << i
			}
		}
	case miIntr, miVersion:
		// read-only
	default:
		m.note("MI: write 0x%08X to undecoded register 0x%08X", v, addr)
	}
}
