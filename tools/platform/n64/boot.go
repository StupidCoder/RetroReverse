package n64

// boot.go supplies the PIF boot handoff — the only machine state this oracle
// invents rather than reads off the medium.
//
// The N64's reset vector points into the PIF ROM, which holds IPL1 and IPL2.
// That chip is in the console, not in the cartridge, so its code is not on the
// image and cannot be executed here. What it leaves behind, however, is small
// and fully specified:
//
//   - The first 0x1000 bytes of the cartridge — the 64-byte header followed by
//     IPL3 — are copied into the RSP's data memory.
//   - The CPU is started at 0xA4000040, which is IPL3's first instruction, with
//     DMEM otherwise addressable so IPL3 can read the header behind it.
//   - Six registers carry the console's configuration into IPL3.
//
// Everything after that instruction is the cartridge's own code: IPL3 initialises
// RDRAM, sizes it, copies the game's boot segment out of the cartridge, and jumps
// to the entry point in the header. So the escalation stops here, at the smallest
// irreducible handoff, exactly as the analysis playbook requires.
//
// The load-bearing value is $s6, the CIC seed. IPL3 checksums itself against a
// value derived from it, and a wrong seed makes the cartridge dead-loop rather
// than fail loudly — which is why rom.go refuses to guess a boot chip it has not
// seen.

import (
	"fmt"

	"retroreverse.com/tools/cpu/r4300"
)

// Where IPL3 runs, and where IPL2 leaves the stack and return address.
const (
	ipl3Entry  = 0xA4000040 // DMEM + 0x40: IPL3's first instruction
	ipl2Stack  = 0xA4001FF0 // the top of IMEM, which IPL2 uses as IPL3's stack
	ipl2Return = 0xA4001550 // $ra: back into IPL2, which IPL3 never takes

	// osMemSize is the RDRAM size in bytes, which libultra's allocator reads.
	osMemSizeAddr = 0x80000318
	osMemSizePhys = 0x00000318

	// riSelect is the RI register IPL3 tests to decide whether RDRAM has already
	// been brought up; riSelectInitialised is the value IPL3 itself writes there
	// once it has (see A40000B4 in Pilotwings' IPL3).
	riSelect            = 0x0C
	riSelectInitialised = 0x14
)

// TV standards, as $s4 reports them.
const (
	TVPAL  = 0
	TVNTSC = 1
	TVMPAL = 2
)

// Reset types, as $s5 reports them.
const (
	ResetCold = 0
	ResetNMI  = 1
)

// BootConfig is the console configuration IPL2 passes to IPL3.
type BootConfig struct {
	TVType    uint32 // $s4
	ResetType uint32 // $s5
	OSVersion uint32 // $s7
}

// DefaultBoot is a cold power-on of an NTSC console.
func DefaultBoot() BootConfig {
	return BootConfig{TVType: TVNTSC, ResetType: ResetCold, OSVersion: 0}
}

// Boot places the machine in the state IPL2 leaves behind and points the CPU at
// IPL3. It reads the CIC seed from the cartridge's identified boot chip.
//
// One further substitution is unavoidable, and it is worth being precise about.
// On a cold power-on IPL3 initialises RDRAM itself: it walks the memory devices,
// tunes their access timing by writing patterns and counting stable read-backs,
// then sums the devices it found into osMemSize. That routine measures the
// electrical behaviour of real memory chips. Here RDRAM is a Go slice — there is
// nothing to tune, and no honest answer to give the probe, which consequently
// reports no memory at all.
//
// So the oracle presents the machine as IPL3 itself leaves it once RDRAM is up:
// RI_SELECT holds 0x14 (the value IPL3 writes at the end of a successful init),
// which makes it take its own already-initialised path, and osMemSize holds the
// installed size. How much memory is fitted is a property of the console, not of
// the cartridge — the same class of fact as the CIC seed. Everything downstream
// remains the cartridge's own code: the cache invalidation, the relocation of
// the loader stub into RDRAM, the megabyte DMA, the seeded checksum over it, and
// the jump to the entry point.
func (m *Machine) Boot(rom *ROM, cfg BootConfig) error {
	if len(rom.Data) < IPL3End {
		return fmt.Errorf("n64: cartridge is too short to hold IPL3")
	}
	// IPL2 copies the header and IPL3 into DMEM. The header stays visible at
	// 0xA4000000 so IPL3 can read CRC1 from it.
	copy(m.DMEM, rom.Data[:spMemSize])

	m.ri[riSelect] = riSelectInitialised
	m.writePhys32(osMemSizePhys, uint32(len(m.RDRAM)))

	c := m.CPU
	c.Reset()
	// IPL3 runs from DMEM with the caches and the TLB untouched, so clear the
	// reset-forced ERL/BEV: the vectors it would use are never taken, and leaving
	// ERL set would make KSEG0 behave as unmapped-uncached for the game too.
	c.COP0[r4300.Cop0Status] = r4300.StatusCU1

	c.SetReg(3, 0)                      // $v1
	c.SetReg(11, ipl3Entry)             // $t3: IPL2 leaves IPL3's own address here
	c.SetReg(19, 0)                     // $s3: osRomType, 0 = cartridge
	c.SetReg(20, uint64(cfg.TVType))    // $s4: osTvType
	c.SetReg(21, uint64(cfg.ResetType)) // $s5: osResetType
	c.SetReg(22, uint64(rom.CIC.Seed))  // $s6: the CIC seed IPL3 checks itself against
	c.SetReg(23, uint64(cfg.OSVersion)) // $s7: osVersion
	c.SetReg(29, ipl2Stack)             // $sp
	c.SetReg(31, ipl2Return)            // $ra

	c.SetPC(ipl3Entry)
	return nil
}

// OSMemSize reads the RDRAM size IPL3 discovered and recorded for libultra.
func (m *Machine) OSMemSize() uint32 {
	v, _ := m.ReadVirt(osMemSizeAddr)
	return v
}
