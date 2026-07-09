package n64

// pi.go is the Peripheral Interface: the DMA engine between the cartridge and
// RDRAM, and the timing configuration for the cartridge bus.
//
// Every asset a game loads arrives through here. libultra's osPiStartDma writes
// the source, destination and length, then blocks on the PI interrupt that
// signals completion — so a model that performs the copy but never raises the
// interrupt stalls the boot just as surely as one that never copies.
//
// The transfer is instantaneous here. Nothing in a game can observe the transfer
// rate except by timing it, and the oracle's clock is instruction count.

const (
	piDramAddr = 0x00
	piCartAddr = 0x04
	piRdLen    = 0x08
	piWrLen    = 0x0C
	piStatus   = 0x10
)

// PI_STATUS bits. Reads report progress; writes are commands.
const (
	piStatusDMABusy = 1 << 0
	piStatusIOBusy  = 1 << 1

	piStatusResetCmd     = 1 << 0 // write: abort the current transfer
	piStatusClearIntrCmd = 1 << 1 // write: acknowledge the PI interrupt
)

// Fields are exported so encoding/gob carries them into a save-state; see
// state.go.
type pi struct {
	DramAddr uint32
	CartAddr uint32
	Status   uint32
	Regs     regFile // the four domain-timing registers, which have no effect here
}

func (p *pi) init() { p.Regs = regFile{} }

func (m *Machine) piRead(addr uint32) uint32 {
	switch addr & 0xFF {
	case piDramAddr:
		return m.pi.DramAddr
	case piCartAddr:
		return m.pi.CartAddr
	case piRdLen, piWrLen:
		// The length registers read back as the last value written, which no
		// real code relies on; the transfer has already completed.
		return m.pi.Regs[addr&0xFF]
	case piStatus:
		// DMA completes within the write that starts it, so the busy bits are
		// never observed set. Code that polls instead of waiting on the
		// interrupt therefore falls straight through, which is what we want.
		return m.pi.Status
	}
	return m.pi.Regs[addr&0xFF]
}

func (m *Machine) piWrite(addr uint32, v uint32) {
	switch addr & 0xFF {
	case piDramAddr:
		m.pi.DramAddr = v & 0x00FFFFFF & ^uint32(1)
	case piCartAddr:
		m.pi.CartAddr = v &^ 1
	case piRdLen: // RDRAM -> cartridge: only save hardware has a writable domain
		m.pi.Regs[piRdLen] = v
		m.note("PI: write-to-cartridge DMA of %d bytes (ignored: the ROM is read-only)", v+1)
		m.raiseIRQ(intrPI)
	case piWrLen: // cartridge -> RDRAM: the transfer every asset load uses
		m.pi.Regs[piWrLen] = v
		m.piDMA(v + 1)
	case piStatus:
		if v&piStatusResetCmd != 0 {
			m.pi.Status = 0
		}
		if v&piStatusClearIntrCmd != 0 {
			m.clearIRQ(intrPI)
		}
	default:
		m.pi.Regs[addr&0xFF] = v
	}
}

// piDMA copies length bytes from the cartridge into RDRAM and raises the PI
// interrupt. The cartridge address is physical (0x10000000-based); the RDRAM
// address is an offset into RDRAM.
func (m *Machine) piDMA(length uint32) {
	dram, cart := m.pi.DramAddr, m.pi.CartAddr
	if m.OnDMA != nil {
		m.OnDMA("pi", dram, cart, length)
	}
	for i := uint32(0); i < length; i++ {
		d := dram + i
		if int(d) >= len(m.RDRAM) {
			m.note("PI: DMA of %d bytes to 0x%08X runs past RDRAM (truncated)", length, dram)
			break
		}
		src := cart + i
		var b byte
		if s, off := m.backing(src); s != nil {
			b = s[off]
		} else {
			m.note("PI: DMA source 0x%08X is not the cartridge (reading 0)", src)
		}
		m.RDRAM[d] = b
	}
	// The transfer is complete before the write that started it returns, so the
	// busy bits never latch; the interrupt is what libultra waits on.
	m.pi.DramAddr = dram + length
	m.pi.CartAddr = cart + length
	m.raiseIRQ(intrPI)
}
