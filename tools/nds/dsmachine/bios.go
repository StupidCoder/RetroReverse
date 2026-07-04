package dsmachine

import "retroreverse.com/tools/arm"

// biosSWI models the DS BIOS software interrupts a boot needs: the memory
// copy/fill helpers, the busy-wait, and the interrupt waits (IntrWait / Halt),
// which park the core until its interrupt controller has a matching pending
// request (delivered by the scheduler). Decompression SWIs are added on demand.
func biosSWI(c *core) func(*arm.CPU, uint32) bool {
	b := &bus{c: c}
	return func(cpu *arm.CPU, comment uint32) bool {
		n := comment & 0xFF
		if n == 0 { // Thumb SWI encodes the number in the low byte of the halfword
			n = (comment >> 16) & 0xFF
		}
		switch n {
		case 0x03: // WaitByLoop(count): a busy-wait. Rather than skip it (which
			// collapses the delays a polled handshake relies on), sleep the caller
			// for that many instruction-times so the *other* core runs meanwhile.
			d := int(cpu.R[0])
			if d > 1<<20 {
				d = 1 << 20 // cap pathological waits
			}
			c.sleep += d
			return true
		case 0x0B: // CpuSet
			cpuSet(b, cpu, false)
			return true
		case 0x0C: // CpuFastSet
			cpuSet(b, cpu, true)
			return true
		case 0x04: // IntrWait(discardOld, mask)
			c.park(cpu, cpu.R[1], false)
			return true
		case 0x05: // VBlankIntrWait
			c.park(cpu, irqVBlank, false)
			return true
		case 0x06: // Halt — wake on any interrupt
			c.park(cpu, 0, true)
			return true
		case 0x07: // Stop (deep sleep) — end of the line for a boot trace
			cpu.Halt("%s: SWI Stop (deep sleep)", c.name)
			return true
		default:
			c.m.note("%s: SWI 0x%02X (ignored)", c.name, n)
			return true
		}
	}
}

// park puts the core into an interrupt wait: it stops stepping until the
// scheduler delivers a matching IRQ, which resumes it just past the SWI.
func (c *core) park(cpu *arm.CPU, mask uint32, any bool) {
	c.waiting = true
	c.waitMask = mask
	c.waitAny = any
	c.resumePC = cpu.R[15] // instruction after the SWI wrapper's call
}

// cpuSet implements CpuSet / CpuFastSet (memory copy or fill).
func cpuSet(b *bus, c *arm.CPU, fast bool) {
	src, dst, ctrl := c.R[0], c.R[1], c.R[2]
	fill := ctrl&(1<<24) != 0
	n := ctrl & 0x1FFFFF
	word := fast || ctrl&(1<<26) != 0
	if word {
		var v uint32
		if fill {
			v = b.r32(src)
		}
		for i := uint32(0); i < n; i++ {
			if !fill {
				v = b.r32(src + i*4)
			}
			b.w32(dst+i*4, v)
		}
		return
	}
	var v uint32 // 16-bit units
	if fill {
		v = uint32(b.Read(src)) | uint32(b.Read(src+1))<<8
	}
	for i := uint32(0); i < n; i++ {
		if !fill {
			v = uint32(b.Read(src+i*2)) | uint32(b.Read(src+i*2+1))<<8
		}
		b.Write(dst+i*2, byte(v))
		b.Write(dst+i*2+1, byte(v>>8))
	}
}

// cp15 handles the ARM9's CP15 coprocessor: it captures the ITCM/DTCM base+size
// the boot programs (registers c9,c1) so the machine maps the TCMs where the game
// expects them; everything else (caches, MPU) is accepted and ignored.
func cp15(c *core) func(*arm.CPU, bool, uint32, uint32, uint32, uint32, uint32, *uint32) {
	return func(cpu *arm.CPU, load bool, cp, op1, crn, crm, op2 uint32, rd *uint32) {
		if load || crn != 9 || crm != 1 {
			return
		}
		v := *rd
		base := v &^ 0xFFF
		if op2 == 0 { // DTCM base/size
			if c.dtcm == nil {
				c.dtcm = make([]byte, 0x4000)
			}
			c.dtcmBase = base
			c.handlerBase = base + uint32(len(c.dtcm))
			c.m.note("ARM9: CP15 DTCM base 0x%08X", base)
		} else { // op2==1 ITCM base/size (base is fixed at 0 on the DS; size varies)
			c.m.note("ARM9: CP15 ITCM control 0x%08X", v)
		}
	}
}
