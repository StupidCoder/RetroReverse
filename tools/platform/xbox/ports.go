package xbox

// ports.go models the x86 I/O-port space. The Xbox's southbridge exposes the PCI
// configuration mechanism (0xCF8/0xCFC), the ACPI/PM timer, and the SMBus at fixed
// ports; XDK boot code touches a few of them. Most of the platform's device state is
// MMIO (nv2a.go), so this is a thin layer: answer the reads the boot path makes with
// plausible values and record the writes. Anything that turns out to matter gets
// modelled when a run shows the title depending on it.

// portIn services an IN instruction (size 1/2/4 bytes).
func (m *Machine) portIn(port uint16, size int) uint32 {
	switch port {
	case 0xCFC, 0xCFD, 0xCFE, 0xCFF: // PCI config data
		return m.pciConfigRead(size)
	case 0x8008: // ACPI PM timer (24-bit, ~3.375 MHz) — derived from the tick
		return uint32(m.tick) & 0x00FFFFFF
	}
	// Unmodelled port: reads as all-ones, the bus-idle value.
	return 0xFFFFFFFF >> uint(32-size*8)
}

// portOut services an OUT instruction.
func (m *Machine) portOut(port uint16, size int, v uint32) {
	switch port {
	case 0xCF8: // PCI config address
		m.pciAddr = v
	}
}

// pciConfigRead answers a read from the PCI config data port using the last address
// written to 0xCF8. Only the fields the boot path reads are meaningful; the rest are
// zero. The NV2A (bus 1, dev 0) and the nForce host bridge answer their vendor/device
// IDs so a bus scan finds the expected parts.
func (m *Machine) pciConfigRead(size int) uint32 {
	// Config address: bit31 enable, bus[23:16], dev[15:11], fn[10:8], reg[7:2].
	reg := m.pciAddr & 0xFC
	bus := (m.pciAddr >> 16) & 0xFF
	dev := (m.pciAddr >> 11) & 0x1F
	var v uint32
	switch {
	case bus == 0 && dev == 0 && reg == 0: // host bridge: nVidia nForce
		v = 0x02A010DE
	case bus == 1 && dev == 0 && reg == 0: // NV2A GPU
		v = 0x02A010DE
	default:
		v = 0xFFFFFFFF // no device
	}
	return v >> uint((m.pciAddr&3)*8) & (0xFFFFFFFF >> uint(32-size*8))
}
