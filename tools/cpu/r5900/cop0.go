package r5900

// cop0.go is the system control coprocessor: virtual-to-physical address
// translation (the segment map and the TLB) and the COP0 register file.
//
// Unlike the N64, a PS2 program does not run unmapped. The boot ELF is linked at
// 0x00100000 — inside KUSEG, the *mapped* segment — so every instruction it
// executes is translated through the TLB. On real hardware the kernel in the BIOS
// ROM installs the entries that map main memory before it hands over; there is no
// BIOS here, so tools/platform/ps2 installs the same mapping as part of its kernel
// HLE (see Machine.mapMemory). The CPU stays honest: it translates, and faults if
// nothing maps the address.
//
// The scratchpad is the one exception. It answers at 0x70000000 without a TLB
// entry — a small fast RAM wired directly into the address decoder — so it is
// handled in the segment map alongside KSEG0 and KSEG1.

import "fmt"

// TLBSize is the number of TLB entries. Each maps a *pair* of pages.
const TLBSize = 48

// TLBEntry is one translation. A single entry covers two consecutive pages of
// PageMask-selected size: the even page through EntryLo0 and the odd through
// EntryLo1.
type TLBEntry struct {
	PageMask uint32 // bits 13..24 select the page size (4 KiB .. 16 MiB)
	EntryHi  uint64 // VPN2 and the ASID
	EntryLo0 uint64 // even page: PFN, and the cache/dirty/valid/global bits
	EntryLo1 uint64 // odd page
}

// EntryLo bits.
const (
	entryLoG = 1 << 0 // global: match regardless of ASID (must be set in both)
	entryLoV = 1 << 1 // valid
	entryLoD = 1 << 2 // dirty: writable
)

// pairSize is the span an entry covers: both of its pages together.
func (e *TLBEntry) pairSize() uint64 { return uint64(e.PageMask|0x1FFF) + 1 }

// pageSize is the size of each of the entry's two pages — half the pair.
func (e *TLBEntry) pageSize() uint64 { return e.pairSize() / 2 }

// Segment boundaries in the 32-bit address space.
const (
	kuseg = 0x00000000 // 0x00000000..0x7FFFFFFF  mapped
	kseg0 = 0x80000000 // 0x80000000..0x9FFFFFFF  unmapped, cached
	kseg1 = 0xA0000000 // 0xA0000000..0xBFFFFFFF  unmapped, uncached
	ksseg = 0xC0000000 // 0xC0000000..0xDFFFFFFF  mapped
	kseg3 = 0xE0000000 // 0xE0000000..0xFFFFFFFF  mapped

	// The scratchpad: 16 KiB of fast RAM answering directly at this virtual
	// address, with no TLB entry behind it.
	spramBase = 0x70000000
	spramSize = 0x4000
)

// Translate maps a virtual address to a physical one. store selects whether a
// write permission check applies (a store to a valid but clean page raises the
// TLB-modification exception).
//
// On a miss or a protection fault it raises the appropriate exception and returns
// ok == false; the caller must abandon the access.
func (c *CPU) Translate(vaddr uint64, store bool) (uint32, bool) {
	v := uint32(vaddr)

	switch {
	case v >= spramBase && v < spramBase+spramSize:
		return v, true // the scratchpad answers at its virtual address
	case v >= kseg0 && v < kseg1:
		return v - kseg0, true // unmapped, cached
	case v >= kseg1 && v < ksseg:
		return v - kseg1, true // unmapped, uncached
	}
	// KUSEG, KSSEG and KSEG3 go through the TLB.
	return c.tlbTranslate(vaddr, store)
}

// tlbTranslate searches the TLB for an entry covering vaddr.
func (c *CPU) tlbTranslate(vaddr uint64, store bool) (uint32, bool) {
	asid := c.COP0[cop0EntryHi] & 0xFF

	for i := range c.TLB {
		e := &c.TLB[i]
		ps := e.pageSize()
		// The entry covers an aligned pair of pages; compare the tags above them.
		pairMask := ^(e.pairSize() - 1)
		if (vaddr & pairMask & 0xFFFFFFFF) != (e.EntryHi & pairMask & 0xFFFFFFFF) {
			continue
		}
		global := e.EntryLo0&e.EntryLo1&entryLoG != 0
		if !global && e.EntryHi&0xFF != asid {
			continue
		}

		// The odd page of the pair is selected by the address bit just above it.
		lo := e.EntryLo0
		if vaddr&ps != 0 {
			lo = e.EntryLo1
		}
		if lo&entryLoV == 0 {
			// An entry matched but the page is invalid: not a refill, so this takes
			// the general vector rather than the fast refill handler.
			c.tlbException(vaddr, store, false)
			return 0, false
		}
		if store && lo&entryLoD == 0 {
			c.setFaultAddress(vaddr)
			c.setEntryHiVPN(vaddr)
			c.Exception(excMod)
			return 0, false
		}
		// PFN counts 4 KiB frames regardless of the page size, so the frame is
		// aligned down to the entry's page size before the offset is added.
		pfn := (lo >> 6) & 0xFFFFF
		return uint32((pfn<<12)&^uint64(ps-1) | (vaddr & (ps - 1))), true
	}

	c.tlbException(vaddr, store, true)
	return 0, false
}

// tlbException raises a TLB miss. A miss on an entry that is absent entirely
// (refill) vectors at offset 0x000 so the handler can be a fast path; a miss on an
// entry present but invalid vectors at the general 0x180 handler.
func (c *CPU) tlbException(vaddr uint64, store, refill bool) {
	code := uint32(excTLBL)
	if store {
		code = excTLBS
	}
	c.setFaultAddress(vaddr)
	c.setEntryHiVPN(vaddr)
	c.exceptionAt(code, refill)
}

// setEntryHiVPN writes the faulting address's VPN2 into EntryHi, preserving the
// ASID, so the handler can fill the entry without recomputing it.
func (c *CPU) setEntryHiVPN(vaddr uint64) {
	c.COP0[cop0EntryHi] = (c.COP0[cop0EntryHi] & 0xFF) | (vaddr &^ 0x1FFF)
}

// setFaultAddress records a faulting address in the two registers a handler reads:
// BadVAddr and Context.
//
// Context is a page-table pointer, not an address: it holds a base the operating
// system wrote, with the faulting address's virtual page number spliced into the
// middle of it, so a handler can load the page-table entry with a single
// instruction. The base must survive, which is why this is a masked write rather
// than an assignment.
//
// It is updated by *every* fault that has an address, not just TLB misses: an
// unaligned load writes it too. A model that only touched it on a TLB miss leaves
// a handler reading the last miss's page.
func (c *CPU) setFaultAddress(vaddr uint64) {
	c.COP0[cop0BadVAddr] = vaddr

	// Context: a 19-bit page number at bits 4..22. The low four bits are hardwired
	// to zero, so they are cleared rather than preserved.
	badVPN2 := (vaddr >> 13) & 0x7FFFF
	c.COP0[cop0Context] = (c.COP0[cop0Context] &^ 0x7FFFFF) | (badVPN2 << 4)
}

// --- the TLB instructions ---------------------------------------------------

// tlbr reads the entry selected by Index into the EntryHi/Lo/PageMask registers.
func (c *CPU) tlbr() {
	i := c.COP0[cop0Index] & 0x3F
	if i >= TLBSize {
		c.Halt("tlbr: Index %d out of range at 0x%08X", i, uint32(c.curPC))
		return
	}
	e := &c.TLB[i]
	c.COP0[cop0PageMask] = uint64(e.PageMask)
	c.COP0[cop0EntryHi] = e.EntryHi
	c.COP0[cop0EntryLo0] = e.EntryLo0
	c.COP0[cop0EntryLo1] = e.EntryLo1
}

// tlbw writes the EntryHi/Lo/PageMask registers into entry i.
func (c *CPU) tlbw(i uint64) {
	if i >= TLBSize {
		c.Halt("tlbw: index %d out of range at 0x%08X", i, uint32(c.curPC))
		return
	}
	pm := uint32(c.COP0[cop0PageMask]) & 0x01FFE000
	c.TLB[i] = TLBEntry{
		PageMask: pm,
		// The stored tag is masked to the entry's page size, so a later compare
		// against a differently-sized entry cannot alias.
		EntryHi:  c.COP0[cop0EntryHi] &^ uint64(pm),
		EntryLo0: c.COP0[cop0EntryLo0],
		EntryLo1: c.COP0[cop0EntryLo1],
	}
}

// SetTLB installs an entry directly. The machine's kernel HLE uses it to lay down
// the mapping the BIOS would have created before the boot ELF was entered.
func (c *CPU) SetTLB(i int, e TLBEntry) {
	if i >= 0 && i < TLBSize {
		c.TLB[i] = e
	}
}

// tlbp searches for an entry matching EntryHi and writes its index into Index, or
// sets Index's top bit to report no match.
func (c *CPU) tlbp() {
	hi := c.COP0[cop0EntryHi]
	asid := hi & 0xFF
	for i := range c.TLB {
		e := &c.TLB[i]
		pairMask := ^(e.pairSize() - 1)
		if (hi & pairMask & 0xFFFFFFFF) != (e.EntryHi & pairMask & 0xFFFFFFFF) {
			continue
		}
		global := e.EntryLo0&e.EntryLo1&entryLoG != 0
		if !global && e.EntryHi&0xFF != asid {
			continue
		}
		c.COP0[cop0Index] = uint64(i)
		return
	}
	c.COP0[cop0Index] = 1 << 31 // probe failure
}

// random returns the Random register, which free-runs downward over the entries
// above Wired. tlbwr uses it to pick a victim.
func (c *CPU) random() uint64 {
	wired := c.COP0[cop0Wired] & 0x3F
	if wired >= TLBSize {
		return TLBSize - 1
	}
	r := c.COP0[cop0Random] & 0x3F
	if r <= wired || r >= TLBSize {
		r = TLBSize - 1
	} else {
		r--
	}
	c.COP0[cop0Random] = r
	return r
}

// --- the COP0 register file -------------------------------------------------

// readCop0 returns a COP0 register. Widths differ: some registers are 32-bit and
// read sign-extended, others (the addresses) are 64-bit.
func (c *CPU) readCop0(i uint32) uint64 {
	switch i {
	case cop0Random:
		return c.COP0[cop0Random]
	case cop0Count, cop0Compare, cop0Status, cop0Cause, cop0PRId, cop0Config,
		cop0Index, cop0Wired, cop0PageMask:
		return sext32(uint32(c.COP0[i]))
	}
	return c.COP0[i]
}

// writeCop0 writes a COP0 register, applying the side effects the hardware has:
// writing Compare acknowledges the timer interrupt.
func (c *CPU) writeCop0(i uint32, v uint64) {
	switch i {
	case cop0Compare:
		c.COP0[cop0Compare] = uint64(uint32(v))
		c.COP0[cop0Cause] &^= causeIP7 // writing Compare clears the timer interrupt
		return
	case cop0Cause:
		// Only the two software interrupt bits are writable.
		c.COP0[cop0Cause] = (c.COP0[cop0Cause] &^ 0x300) | (v & 0x300)
		return
	case cop0Count:
		c.COP0[cop0Count] = uint64(uint32(v))
		c.countFrac = 0
		return
	case cop0PRId, cop0Random:
		return // read-only
	}
	c.COP0[i] = v
}

// cop0 executes an op == 0x10 instruction.
func (c *CPU) cop0(w, rs, rt, rd uint32) {
	if w&(1<<25) != 0 { // COP0 command
		switch w & 0x3F {
		case 0x01:
			c.tlbr()
		case 0x02:
			c.tlbw(c.COP0[cop0Index] & 0x3F)
		case 0x06:
			c.tlbw(c.random())
		case 0x08:
			c.tlbp()
		case 0x18:
			c.eret()
		case 0x38: // ei — set the EE's second interrupt enable
			c.COP0[cop0Status] |= statusEIE
		case 0x39: // di
			c.COP0[cop0Status] &^= statusEIE
		default:
			// The coprocessor-0 command space has a handful of defined functions and
			// everything else does nothing at all. It does not fault, which is why a
			// program can probe the space safely — so this must not halt.
		}
		return
	}
	switch rs {
	case 0x00: // mfc0
		c.set(rt, sext32(uint32(c.readCop0(rd))))
	case 0x04: // mtc0
		c.writeCop0(rd, sext32(uint32(c.R[rt].Lo)))
	default:
		// The EE has no dmfc0/dmtc0: COP0 is a 32-bit file on this core.
		c.Exception(excRI)
	}
}

// DumpTLB renders the live TLB, for the oracle's diagnostics.
func (c *CPU) DumpTLB() string {
	s := ""
	for i := range c.TLB {
		e := &c.TLB[i]
		if e.EntryLo0&entryLoV == 0 && e.EntryLo1&entryLoV == 0 {
			continue
		}
		s += fmt.Sprintf("%2d: hi=%016X lo0=%016X lo1=%016X mask=%08X size=%d\n",
			i, e.EntryHi, e.EntryLo0, e.EntryLo1, e.PageMask, e.pageSize())
	}
	if s == "" {
		return "TLB: no valid entries\n"
	}
	return s
}
