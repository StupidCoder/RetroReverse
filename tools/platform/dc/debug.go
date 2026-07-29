package dc

// debug.go is the oracle's window into a stopped machine: registers,
// disassembly and memory, addressed the way the CPU addresses them.

import (
	"fmt"
	"strings"

	"retroreverse.com/tools/cpu/sh4"
)

// RegString renders the register file the way the oracle prints it.
func (m *Machine) RegString() string {
	c := m.CPU
	var b strings.Builder
	for i := 0; i < 16; i++ {
		fmt.Fprintf(&b, "r%-2d %08X  ", i, c.R[i])
		if i%4 == 3 {
			b.WriteByte('\n')
		}
	}
	fmt.Fprintf(&b, "pc  %08X  pr  %08X  sr  %08X  gbr %08X\n", c.PC, c.PR, c.SR, c.GBR)
	fmt.Fprintf(&b, "vbr %08X  mach %08X macl %08X  fpscr %08X fpul %08X\n", c.VBR, c.MACH, c.MACL, c.FPSCR, c.FPUL)
	fmt.Fprintf(&b, "instrs %d  fields %d  ta-words %d\n", m.Instrs, m.Fields, m.TAWrites)
	fmt.Fprintf(&b, "istnrm %08X ext %08X  iml6 %08X/%08X iml4 %08X/%08X iml2 %08X/%08X\n",
		m.Holly.ISTNRM, m.Holly.ISTEXT, m.Holly.IML6NRM, m.Holly.IML6EXT,
		m.Holly.IML4NRM, m.Holly.IML4EXT, m.Holly.IML2NRM, m.Holly.IML2EXT)
	r := func(off uint32) uint32 { return m.PVRRegs[off/4] }
	fmt.Fprintf(&b, "pvr fb_r ctrl %08X sof %08X/%08X size %08X  fb_w ctrl %08X sof %08X  spg %08X vo %08X\n",
		r(0x44), r(0x50), r(0x54), r(0x5C), r(0x48), r(0x60), r(0xD0), r(0xE8))
	if m.ARMRunning {
		fmt.Fprintf(&b, "arm pc %08X  r0 %08X r1 %08X r2 %08X  instrs %d", m.ARM.R[15], m.ARM.R[0], m.ARM.R[1], m.ARM.R[2], m.ARM.Instrs)
		if m.ARM.Halted {
			fmt.Fprintf(&b, "  HALTED: %s", m.ARM.HaltReason)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// ReadBytes copies n bytes from a CPU address (any mirror).
func (m *Machine) ReadBytes(addr uint32, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = m.Read8((addr + uint32(i)) & 0x1FFFFFFF)
	}
	return out
}

// Disasm renders n instructions starting at a CPU address.
func (m *Machine) Disasm(addr uint32, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		a := addr + uint32(i)*2
		h := m.Read16(a & 0x1FFFFFFF)
		in := sh4.DecodeHalfword(h, a)
		out = append(out, fmt.Sprintf("%08X  %04X  %s", a, h, in.Text))
	}
	return out
}
