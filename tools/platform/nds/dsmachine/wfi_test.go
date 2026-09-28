package dsmachine

import (
	"retroreverse.com/tools/cpu/arm"
	"testing"
)

func TestCP15WaitForInterrupt(t *testing.T) {
	m := testMachine()
	c := m.ARM9
	b := &bus{c: c}
	c.cpu = arm.NewCPU(b)
	c.cpu.Coproc = cp15(c)
	c.cpu.R[15] = 0x02000000
	m.visited = map[uint32]bool{}
	b.w32(0x02000000, 0xee070f90) // MCR p15,0,r0,c7,c0,4
	b.w32(0x02000004, 0xe2800001) // ADD r0,r0,#1
	m.runQuantum(c, 64, nil, nil)
	if !c.wfi || c.cpu.Instrs != 1 || c.cpu.R[15] != 0x02000004 {
		t.Fatal("WFI did not stop immediately after the instruction")
	}
	saved := coreSnapshot(c)
	c.wfi = false
	coreRestore(c, &saved)
	if !c.wfi {
		t.Fatal("state lost WFI")
	}
	c.ie, c.if_ = 1, 1
	c.ime = false
	m.deliver(c)
	if !c.wfi {
		t.Fatal("woke without an asserted IRQ line")
	}
	c.ime, c.cpu.IRQDisable = true, true
	m.deliver(c)
	if c.wfi || c.cpu.R[15] != 0x02000004 {
		t.Fatal("masked IRQ must wake without exception entry")
	}
	m.runQuantum(c, 1, nil, nil)
	if c.cpu.R[0] != 1 {
		t.Fatal("instruction after WFI did not resume")
	}
	c.cpu.IRQDisable = false
	c.wfi = true
	c.handlerBase = 0x02004000
	b.w32(c.handlerBase-4, 0x02001000)
	m.deliver(c)
	if c.wfi || c.cpu.R[15] != 0x02001000 || c.cpu.Mode != arm.ModeIRQ {
		t.Fatal("unmasked IRQ did not enter its handler")
	}
}
