package gc

// boot.go holds the direct-load boot path, the bisection tool that sits beside the real
// apploader one in ipl.go.
//
// LoadDOL does what the apploader would do — copy the executable's segments into memory,
// zero its BSS, set the entry point — but without running the loader and without touching
// the disc drive. It exists because when a boot goes wrong, the first question is whether
// the fault is in the CPU or in the machinery around it, and the way to answer it is to
// remove the machinery: if the game misbehaves the same way loaded directly as it does
// loaded through the apploader, the apploader and the disc interface are not the cause.

import "fmt"

// LoadDOL puts the game's executable into memory directly and leaves the machine at its
// entry point, in the post-IPL state — the apploader's result, without the apploader.
func (m *Machine) LoadDOL() (entry uint32, err error) {
	dol, err := m.disc.DOL()
	if err != nil {
		return 0, err
	}
	m.setupLowMem()
	m.setupState()
	dol.Load(func(addr uint32, b []byte) {
		m.dmaToRAM(addr&0x03FFFFFF, b)
	})
	m.CPU.PC = dol.Entry
	return dol.Entry, nil
}

// Poke writes a value into memory after loading and before running — the -poke flag's
// mechanism, for forcing a value the machine does not yet produce so the run can be carried
// past the gap it leaves.
func (m *Machine) Poke(addr, val uint32) {
	m.setRAM32(addr&0x03FFFFFF, val)
}

// PoisonLowMem fills the low-memory globals with a recognisable pattern before the IPL
// writes them, so a read of one the IPL did not set stands out. It is the mechanism behind
// the oracle's -lowmem derivation: run with the region poisoned and an rwatch over it, and
// the game names, in order, every global it depends on — which is how the low-memory map is
// derived from the game rather than transcribed from anywhere.
func (m *Machine) PoisonLowMem() {
	for a := uint32(0); a < 0x3000; a += 4 {
		m.setRAM32(a, 0xF00D0000|a)
	}
}

// String describes the machine's current point, for a run's final report.
func (m *Machine) String() string {
	return fmt.Sprintf("Gekko PC 0x%08X, %d steps", m.CPU.PC, m.CPU.Steps)
}
