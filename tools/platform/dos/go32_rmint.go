package dos

import "retroreverse.com/tools/cpu/x86"

// DPMI 0300h — Simulate Real Mode Interrupt — is how a go32 program reaches the
// DOS and BIOS services that only exist in real mode: it fills a "real-mode call
// structure" (RMCS) with the register values a real INT would see, and the DPMI
// host runs that interrupt on its behalf and writes the results back. We HLE it:
// read the RMCS, service the interrupt in Go against the host environment, and
// write the register/flags results back into the structure. This is the bridge
// the plan calls for — DOS file I/O and the startup's TSR probes land here.
//
// RMCS layout (DPMI 1.0, 0x32 bytes): EDI ESI EBP -reserved- EBX EDX ECX EAX at
// 0x00..0x1F, then flags(word) ES DS FS GS IP CS SP SS as words at 0x20..0x31.

type rmcs struct {
	addr                           uint32 // flat linear address of the structure
	edi, esi, ebp, ebx             uint32
	edx, ecx, eax                  uint32
	flags                          uint16
	es, ds, fs, gs, ip, cs, sp, ss uint16
}

func (p *PM) readRMCS(a uint32) rmcs {
	r := rmcs{addr: a}
	r.edi = p.r32(a + 0x00)
	r.esi = p.r32(a + 0x04)
	r.ebp = p.r32(a + 0x08)
	r.ebx = p.r32(a + 0x10)
	r.edx = p.r32(a + 0x14)
	r.ecx = p.r32(a + 0x18)
	r.eax = p.r32(a + 0x1C)
	r.flags = uint16(p.r32(a + 0x20))
	r.es = uint16(p.r32(a + 0x22))
	r.ds = uint16(p.r32(a + 0x24))
	r.fs = uint16(p.r32(a + 0x26))
	r.gs = uint16(p.r32(a + 0x28))
	r.ip = uint16(p.r32(a + 0x2A))
	r.cs = uint16(p.r32(a + 0x2C))
	r.sp = uint16(p.r32(a + 0x2E))
	r.ss = uint16(p.r32(a + 0x30))
	return r
}

// writeBack stores the (possibly modified) general registers, flags and segment
// registers back into the RMCS. Persisting the segment words matters for services
// that return a far pointer in a segment register — AH=52h (List of Lists) and
// AH=35h (interrupt vector) return ES:BX. Unmodified fields are written back with
// the values just read, so the store is a no-op for the common case.
func (p *PM) writeBack(r *rmcs) {
	p.w32(r.addr+0x00, r.edi)
	p.w32(r.addr+0x04, r.esi)
	p.w32(r.addr+0x08, r.ebp)
	p.w32(r.addr+0x10, r.ebx)
	p.w32(r.addr+0x14, r.edx)
	p.w32(r.addr+0x18, r.ecx)
	p.w32(r.addr+0x1C, r.eax)
	p.Write(r.addr+0x20, byte(r.flags))
	p.Write(r.addr+0x21, byte(r.flags>>8))
	p.w16(r.addr+0x22, r.es)
	p.w16(r.addr+0x24, r.ds)
	p.w16(r.addr+0x26, r.fs)
	p.w16(r.addr+0x28, r.gs)
}

// simulateRealInt services DPMI 0300h. The real-mode call structure is at ES:EDI
// (flat: SegBase[ES] + EDI). intno is BL.
func (p *PM) simulateRealInt(c *x86.CPU) bool {
	intno := c.Reg8(x86.BL)
	rmAddr := c.SegBase[x86.ES] + c.Regs[x86.DI]
	r := p.readRMCS(rmAddr)
	r.flags &^= 0x0001 // assume success (CF clear); a handler sets it on error

	switch intno {
	case 0x2F: // multiplex — report nothing installed
		r.eax &= 0xFFFFFF00 // AL = 0
	case 0x21:
		if !p.rmDOS(&r) {
			c.Halt("DPMI 0300h INT 21h AH=%02X not yet modelled at %08X", byte(r.eax>>8), c.IP)
			return true
		}
	case 0x10, 0x16, 0x33: // video / keyboard / mouse BIOS — benign no-ops
	default:
		c.Halt("DPMI 0300h simulate-real-mode-INT %02Xh (AX=%04X) not yet modelled at %08X",
			intno, uint16(r.eax), c.IP)
		return true
	}

	p.writeBack(&r)
	c.CF = false
	p.logf("DPMI 0300h: INT %02Xh serviced (AX=%04X)", intno, uint16(r.eax))
	return true
}

// rmDOS services the real-mode INT 21h issued through DPMI 0300h, operating on
// the RMCS registers. It returns false for a function not yet modelled, so the
// caller halts with the concrete AH — the next DOS service to implement.
func (p *PM) rmDOS(r *rmcs) bool {
	ah := byte(r.eax >> 8)
	p.DOSCounts[ah]++
	switch ah {
	case 0x30: // Get DOS version -> 7.0 in AL:AH
		r.eax = (r.eax & 0xFFFF0000) | 0x0007
	case 0x1A: // Set Disk Transfer Address — record it (FindFirst/Next would fill it)
		p.dtaSeg, p.dtaOff = r.ds, uint16(r.edx)
	case 0x25: // Set interrupt vector — ignore
	case 0x35: // Get interrupt vector -> ES:BX = 0
		r.es, r.ebx = 0, 0
	case 0x19: // Get current drive -> C:
		r.eax = (r.eax & 0xFFFFFF00) | 2
	case 0x52: // Get DOS "list of lists" (SysVars) -> ES:BX = fabricated LoL (fstat walks it)
		if p.writeDOSStructures() {
			r.es = p.lolSeg
			r.ebx = (r.ebx & 0xFFFF0000) | uint32(p.lolOff)
		} else {
			r.es, r.ebx = 0, 0 // no transfer buffer to build them in
		}
	case 0x33: // Get/set Ctrl-Break flag
		switch byte(r.eax) {
		case 0x00: // get -> DL = state (off)
			r.edx = r.edx & 0xFFFFFF00
		case 0x05: // get boot drive -> DL = C:
			r.edx = (r.edx & 0xFFFFFF00) | 2
		} // set (AL=01) and others: no-op success
	case 0x3C, 0x3D, 0x3E, 0x3F, 0x40, 0x41, 0x42, 0x43, 0x44, 0x47, 0x4E, 0x4F, 0x57, 0x60: // file I/O
		return p.dosFile(r)
	default:
		return false
	}
	return true
}
