package dos

import "retroreverse.com/tools/cpu/x86"

// This file services the software interrupts a go32 image issues in protected
// mode. The dominant one is INT 31h — the DPMI host — through which the go32
// runtime allocates memory, manipulates descriptors, and (later) bridges to real
// mode for DOS file I/O. In the flat model most descriptor operations are
// cosmetic (every selector resolves to base 0), so they succeed as no-ops; the
// memory services hand out linear ranges from the machine's bump arenas. A DPMI
// function we have not modelled halts the CPU with its AX, so the next one to
// implement is an observed fact rather than a guess.

// handleInt is the CPU's IntHook in protected mode. Returning true means the
// interrupt was fully serviced; returning false lets the core halt (there is no
// real-mode IVT to fall back on in a flat go32 image).
func (p *PM) handleInt(c *x86.CPU, n byte) bool {
	switch n {
	case 0x31:
		return p.dpmi(c)
	case 0x21:
		return p.int21(c)
	case 0x2F: // multiplex — report "not installed"
		c.SetReg8(x86.AL, 0)
		c.CF = false
		p.countInt(n)
		return true
	case 0x10: // video BIOS — ignore (no display modelled in Phase A)
		c.CF = false
		p.countInt(n)
		return true
	default:
		return false // core halts: "unhandled INT $XX in protected mode"
	}
}

func (p *PM) countInt(n byte) {
	p.IntCounts[n]++
	if p.IntCounts[n] == 1 {
		p.logf("INT %02Xh (stubbed) AX=%04X at %08X", n, p.CPU.Reg16(x86.AX), p.CPU.IP)
	}
}

// dpmi services INT 31h. The function is the full 16-bit AX; results follow the
// DPMI 1.0 register conventions (CF clear on success).
func (p *PM) dpmi(c *x86.CPU) bool {
	fn := c.Reg16(x86.AX)
	p.DPMICounts[fn]++
	c.CF = false

	// 32-bit register shorthands.
	get32 := func(i int) uint32 { return c.Regs[i] }
	set16 := func(i int, v uint16) { c.SetReg16(i, v) }
	setPair := func(hi, lo int, v uint32) { c.SetReg16(hi, uint16(v>>16)); c.SetReg16(lo, uint16(v)) }
	fail := func(errCode uint16) { c.CF = true; c.SetReg16(x86.AX, errCode) }

	switch fn {
	// --- descriptor management (cosmetic in a flat model) ---
	case 0x0000: // Allocate LDT Descriptors (CX = count)
		set16(x86.AX, p.allocSel(c.Reg16(x86.CX)))
	case 0x0001: // Free LDT Descriptor
	case 0x0003: // Get Selector Increment Value
		set16(x86.AX, 8)
	case 0x0006: // Get Segment Base Address (BX = sel) -> CX:DX = base
		setPair(x86.CX, x86.DX, p.resolveSel(c.Reg16(x86.BX)))
	case 0x0007: // Set Segment Base Address (BX = sel, CX:DX = base)
		p.mapSel(c.Reg16(x86.BX), uint32(c.Reg16(x86.CX))<<16|uint32(c.Reg16(x86.DX)))
	case 0x0008: // Set Segment Limit — flat (4 GiB), ignore
	case 0x0009: // Set Descriptor Access Rights — ignore
	case 0x000A: // Create Alias Descriptor (BX = sel) -> AX = new selector (same base)
		alias := p.allocSel(1)
		p.mapSel(alias, p.resolveSel(c.Reg16(x86.BX)))
		set16(x86.AX, alias)
	case 0x000B: // Get Descriptor (BX = sel) -> write its 8-byte descriptor at ES:EDI.
		// go32 reads a block's descriptor here, flips the access byte to make it a
		// code segment, and installs the copy on a fresh selector (000Ch) — an
		// executable alias of a DOS-memory block. Encoding the base is what lets
		// the later CALLF through that alias land on the real trampoline bytes.
		p.writeDescriptor(get32(x86.DI), p.resolveSel(c.Reg16(x86.BX)))
	case 0x000C: // Set Descriptor (BX = sel) -> adopt the base from the buffer at ES:EDI
		p.mapSel(c.Reg16(x86.BX), p.readDescriptorBase(get32(x86.DI)))

	// --- real-mode / PM interrupt vectors (stubbed) ---
	case 0x0200: // Get Real Mode Interrupt Vector -> CX:DX = 0:0
		setPair(x86.CX, x86.DX, 0)
	case 0x0201: // Set Real Mode Interrupt Vector — ignore
	case 0x0204: // Get PM Interrupt Vector -> CX = selector, EDX = 0
		set16(x86.CX, 0x08)
		c.Regs[x86.DX] = 0
	case 0x0205: // Set PM Interrupt Vector — ignore
	case 0x0202, 0x0203: // Get/Set PM Exception handler — ignore

	// --- DOS (conventional) memory ---
	case 0x0100: // Allocate DOS Memory Block (BX = paragraphs)
		seg, sel, ok := p.allocConv(c.Reg16(x86.BX))
		if !ok {
			fail(0x0008) // insufficient memory
			set16(x86.BX, uint16((p.convTop-p.convNext)>>4))
			break
		}
		set16(x86.AX, seg)
		set16(x86.DX, sel)
	case 0x0101: // Free DOS Memory Block — ignore (bump arena, no reclaim)

	// --- extended memory ---
	case 0x0400: // Get DPMI Version
		set16(x86.AX, 0x0100)  // DPMI 1.0
		set16(x86.BX, 0x0001)  // bit0: 32-bit host
		c.SetReg8(x86.CL, 0x06) // processor: Pentium Pro / P6 class
		c.SetReg8(x86.DH, 0x08) // master PIC base
		c.SetReg8(x86.DL, 0x70) // slave PIC base
	case 0x0500: // Get Free Memory Information -> fill 0x30-byte block at ES:EDI
		p.freeMemInfo(get32(x86.DI))
	case 0x0501: // Allocate Memory Block (BX:CX = size) -> BX:CX = addr, SI:DI = handle
		size := uint32(c.Reg16(x86.BX))<<16 | uint32(c.Reg16(x86.CX))
		base, ok := p.allocHeap(size)
		if !ok {
			fail(0x0008)
			break
		}
		setPair(x86.BX, x86.CX, base)
		setPair(x86.SI, x86.DI, base) // handle == base (bump arena, no reclaim)
	case 0x0502: // Free Memory Block — ignore
	case 0x0503: // Resize Memory Block (BX:CX size, SI:DI handle) -> grow in place
		size := uint32(c.Reg16(x86.BX))<<16 | uint32(c.Reg16(x86.CX))
		base, ok := p.allocHeap(size)
		if !ok {
			fail(0x0008)
			break
		}
		setPair(x86.BX, x86.CX, base)
		setPair(x86.SI, x86.DI, base)

	// --- page / lock management (no paging modelled: succeed) ---
	// --- coprocessor (we model a real x87) ---
	case 0x0E00: // Get Coprocessor Status -> AX (bit0 present, bit1 enabled, bits8-11 type)
		set16(x86.AX, 0x0045) // present + enabled, 80387-class
	case 0x0E01: // Set Coprocessor Emulation — ignore (a real FPU is always present)

	case 0x0507: // (go32/CWSDPMI page attributes) — no-op
	case 0x0600, 0x0601: // Lock / Unlock Linear Region — no-op
	case 0x0604: // Get Page Size -> BX:CX
		setPair(x86.BX, x86.CX, go32PageSize)

	// --- virtual interrupt state (go32 critical sections) ---
	case 0x0900: // Get and Disable Virtual Interrupt State -> AL = previous
		c.SetReg8(x86.AL, b2u8(p.virtIF))
		p.virtIF = false
	case 0x0901: // Get and Enable Virtual Interrupt State -> AL = previous
		c.SetReg8(x86.AL, b2u8(p.virtIF))
		p.virtIF = true
	case 0x0902: // Get Virtual Interrupt State -> AL = current
		c.SetReg8(x86.AL, b2u8(p.virtIF))

	// --- real-mode bridge (file I/O) — not yet modelled ---
	case 0x0300: // Simulate Real Mode Interrupt
		return p.simulateRealInt(c)
	case 0x0303: // Allocate Real Mode Callback Address -> CX:DX = real-mode far addr
		// go32 installs callbacks for signal delivery / interrupt reflection. We
		// hand out a unique fabricated real-mode address; nothing in real mode calls
		// it in the HLE model, so it need only be distinct and remembered.
		seg, off := p.allocCallback()
		set16(x86.CX, seg)
		set16(x86.DX, off)
	case 0x0304: // Free Real Mode Callback Address — ignore

	default:
		c.Halt("unimplemented DPMI function AX=%04X (BX=%04X CX=%04X) at %08X",
			fn, c.Reg16(x86.BX), c.Reg16(x86.CX), c.IP)
		return true
	}
	return true
}

// writeDescriptor lays out a standard 8-byte x86 segment descriptor at flat
// linear address a: the given base, a 4 GiB granular limit, and present/data
// access. Only the base round-trips through go32's get/modify/set dance, but the
// rest keeps the buffer a well-formed descriptor.
func (p *PM) writeDescriptor(a, base uint32) {
	var limit uint32 = 0xFFFFF // with 4K granularity -> 4 GiB
	p.Write(a+0, byte(limit))
	p.Write(a+1, byte(limit>>8))
	p.Write(a+2, byte(base))
	p.Write(a+3, byte(base>>8))
	p.Write(a+4, byte(base>>16))
	p.Write(a+5, 0xF3)                         // present, DPL 3, data, writable
	p.Write(a+6, byte(limit>>16)&0x0F|0xC0) // limit[19:16] | G=1,B=1
	p.Write(a+7, byte(base>>24))
}

// readDescriptorBase extracts the 32-bit segment base from an 8-byte descriptor
// at flat linear address a (bytes 2,3,4,7).
func (p *PM) readDescriptorBase(a uint32) uint32 {
	return uint32(p.Read(a+2)) | uint32(p.Read(a+3))<<8 |
		uint32(p.Read(a+4))<<16 | uint32(p.Read(a+7))<<24
}

// allocCallback hands out a unique fabricated real-mode callback far address. The
// segment is a reserved marker (0xC000) that no real memory block uses; the
// offset counts up so each callback is distinct.
func (p *PM) allocCallback() (seg, off uint16) {
	off = p.nextCallback
	p.nextCallback += 4
	return 0xC000, off
}

func b2u8(v bool) byte {
	if v {
		return 1
	}
	return 0
}

// allocSel hands out n cosmetic LDT selectors and returns the first. Values are
// kept in the LDT range (bit 2 set) and 8-aligned; nothing in the flat model
// depends on them beyond being distinct and non-zero.
func (p *PM) allocSel(n uint16) uint16 {
	if n == 0 {
		n = 1
	}
	sel := p.nextSel
	p.nextSel += n * 8
	return sel
}

// allocConv bump-allocates paras paragraphs (16 bytes) of conventional memory
// (below 1 MiB). The returned real-mode segment is base>>4, so seg<<4 recovers
// the linear base — the identity go32 relies on when it reaches the DOS transfer
// buffer through a flat near pointer.
func (p *PM) allocConv(paras uint16) (seg, sel uint16, ok bool) {
	bytes := uint32(paras) << 4
	if p.convTop == 0 || p.convNext+bytes > p.convTop {
		return 0, 0, false
	}
	base := p.convNext
	p.convNext = align(p.convNext+bytes, 16)
	sel = uint16(base >> 4)
	p.mapSel(sel, base) // the returned selector's descriptor base is the block's linear address
	p.logf("DPMI 0100h: %d paras conv mem -> seg %04X sel %04X (linear %05X)", paras, uint16(base>>4), sel, base)
	return uint16(base >> 4), sel, true
}

// allocHeap bump-allocates size bytes of extended memory (>= 1 MiB) and returns
// its linear base.
func (p *PM) allocHeap(size uint32) (base uint32, ok bool) {
	base = p.heapNext
	next := align(base+size, 16)
	if next > p.stackFloor { // keep the heap below the high PM stack region
		return 0, false
	}
	p.heapNext = next
	p.logf("DPMI 0501h: %d bytes heap -> linear %08X", size, base)
	return base, true
}

// go32MaxFreeReport caps the free memory DPMI 0500h advertises. Quake sizes its
// main heap to the free memory it sees (COM_Init/Memory_Init), and reporting the
// whole ~62 MiB backing makes it reserve a heap that nearly fills the flat space —
// so its startup CRC over that heap reads off the top of memory. Real Quake
// shareware ran happily on 8–16 MiB; capping the report to 16 MiB keeps the heap
// well inside the arena and cuts the multi-second Sys_PageIn on every boot. This is
// a behavioral change, but a legitimate one — it is exactly what a smaller DOS box
// would report.
const go32MaxFreeReport = 16 << 20

// freeMemInfo fills the DPMI 0500h information block (0x30 bytes) at flat linear
// address a. Only the fields go32's allocator reads matter: the largest free
// block and the total free/physical counts, all set to the heap remaining.
func (p *PM) freeMemInfo(a uint32) {
	free := p.stackFloor - p.heapNext
	if free > go32MaxFreeReport {
		free = go32MaxFreeReport
	}
	for i := uint32(0); i < 0x30; i += 4 {
		p.w32(a+i, 0xFFFFFFFF) // "unknown" for the fields we don't track
	}
	p.w32(a+0x00, free)               // largest available free block (bytes)
	p.w32(a+0x04, free/go32PageSize)  // maximum unlocked page allocation
	p.w32(a+0x08, free/go32PageSize)  // maximum locked page allocation
	p.w32(a+0x10, free/go32PageSize)  // free pages
	p.w32(a+0x14, free/go32PageSize)  // total physical pages
}

// int21 services the few INT 21h functions a go32 image issues directly in
// protected mode (mostly the DPMI-failure exit path). Anything else halts.
func (p *PM) int21(c *x86.CPU) bool {
	ah := c.Reg8(x86.AH)
	c.CF = false
	switch ah {
	case 0x00, 0x4C: // terminate
		p.Terminated = true
		if ah == 0x4C {
			p.ExitCode = c.Reg8(x86.AL)
		}
		c.Halt("INT 21h/%02X — program exit (code %d) at %08X", ah, p.ExitCode, c.IP)
	case 0x09: // print $-terminated string at DS:EDX
		p.logf("DOS print: %q", p.dollarStr(c.Regs[x86.DX]))
	case 0x25: // Set Interrupt Vector — ignore
	case 0x30: // Get DOS Version -> 7.0
		c.SetReg8(x86.AL, 7)
		c.SetReg8(x86.AH, 0)
	case 0x35: // Get Interrupt Vector -> ES:EBX = 0
		c.Seg[x86.ES] = 0
		c.Regs[x86.BX] = 0
	default:
		c.Halt("unimplemented INT 21h AH=%02X in protected mode at %08X", ah, c.IP)
	}
	return true
}

// dollarStr reads a '$'-terminated DOS string at flat linear address a.
func (p *PM) dollarStr(a uint32) string {
	var b []byte
	for i := uint32(0); i < 4096; i++ {
		ch := p.Read(a + i)
		if ch == '$' {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}
