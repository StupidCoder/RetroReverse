package gekko

// mmu.go translates a virtual address to a physical one.
//
// A GameCube does this almost entirely with the block-address-translation registers —
// eight of them, four for instructions and four for data, each mapping a block of up to
// 256 MiB. That is enough to cover the whole machine in a handful of entries, and so the
// page table, which PowerPC also has, is never installed: a GameCube program sets its
// BATs at boot and never thinks about translation again.
//
// The consequence is a memory map that looks like this to every GameCube program, and
// it is worth writing down because every address in this project is one of these:
//
//	0x80000000..0x81800000   main memory, cached      -> physical 0x00000000
//	0xC0000000..0xC1800000   main memory, uncached    -> physical 0x00000000
//	0xCC000000..             the hardware registers   -> physical 0x0C000000
//	0xE0000000..0xE0004000   the locked cache          (not memory at all; see cache.go)
//
// The same physical memory appears twice, and which alias a program uses is a statement
// about coherency: a write through 0x80000000 may sit in the data cache, and a write
// through 0xC0000000 goes straight out. It matters enormously to hardware that reads
// memory behind the CPU's back — the graphics FIFO, the DMA engines — and not at all to
// this core, which has no write-back cache to be incoherent with.
//
// The page-table walk is implemented so that a program which does install one is not
// silently mistranslated, but the BAT miss is the interesting event: on this machine it
// means something has gone wrong, so it halts loudly and names the address rather than
// quietly walking a page table that was never set up.

import "fmt"

// Translate maps an effective address to a physical one. store says the access writes;
// insn says it is an instruction fetch (which consults the IBATs rather than the DBATs).
//
// When translation is off — MSR[IR]/[DR] clear, which is how the machine starts and how
// an exception handler runs — the effective address is the physical address.
func (c *CPU) Translate(ea uint32, store, insn bool) (uint32, bool) {
	on := c.MSR&MSRDR != 0
	if insn {
		on = c.MSR&MSRIR != 0
	}
	if !on {
		return ea, true
	}

	// The locked cache is mapped by address, not by a BAT, and it is not memory: it
	// answers before translation does. (Software still points a DBAT at it, but the
	// cache wins.)
	if c.LC.Enabled() && c.LC.Contains(ea) {
		return ea, true
	}

	bats := &c.DBAT
	if insn {
		bats = &c.IBAT
	}
	for i := 0; i < 4; i++ {
		upper, lower := bats[i][0], bats[i][1]
		if pa, ok := batMatch(upper, lower, ea, c.MSR&MSRPR != 0); ok {
			return pa, true
		}
	}

	// No block covers it. On a GameCube this is not a page fault to be serviced — it is
	// a program that has gone somewhere it never mapped, and pretending otherwise buries
	// the bug. If a title ever does install a page table, this halt is the ticket to
	// implementing the walk.
	kind := "data"
	if insn {
		kind = "instruction"
	}
	c.Halt("gekko: no BAT maps the %s address 0x%08X (PC 0x%08X); this machine does not use a page table", kind, ea, c.PC)
	return 0, false
}

// batMatch tests one block-translation entry.
//
// The upper register holds the effective block start, a mask saying how big the block is,
// and two valid bits — one for supervisor state and one for user state. The lower holds
// the physical block start and the protection bits. A block's size is encoded as a mask
// of the address bits *below* the block granule, which is why the comparison and the
// address composition both use it.
func batMatch(upper, lower, ea uint32, user bool) (uint32, bool) {
	vs := upper&2 != 0 // valid in supervisor state
	vp := upper&1 != 0 // valid in user state
	if user && !vp {
		return 0, false
	}
	if !user && !vs {
		return 0, false
	}

	bepi := upper & 0xFFFE0000 // the block's effective address (the top 15 bits)
	bl := (upper >> 2) & 0x7FF // the block length, in units of 128 KiB, minus one

	// The block spans (BL+1) × 128 KiB, so the bits of an address that fall *within* the
	// block are ((BL+1) << 17) - 1 — equivalently (BL << 17) | 0x1FFFF. The rest select
	// the block. Getting this mask wrong does not fail loudly: it produces a translation
	// that drops the low bits of every address, so the machine reads the right block and
	// the wrong word in it.
	offsetMask := (bl << 17) | 0x1FFFF
	blockMask := ^offsetMask

	if ea&blockMask != bepi&blockMask {
		return 0, false
	}

	brpn := lower & 0xFFFE0000 // the block's physical address
	// The block comes from the physical base; the offset within it comes from the
	// effective address.
	return (brpn & blockMask) | (ea & offsetMask), true
}

// BATString renders the block translations, for a machine that wants to print its own
// memory map — which is the fastest way to see that a boot has set itself up correctly.
func (c *CPU) BATString() string {
	var s string
	for i := 0; i < 4; i++ {
		s += batLine("IBAT", i, c.IBAT[i][0], c.IBAT[i][1])
	}
	for i := 0; i < 4; i++ {
		s += batLine("DBAT", i, c.DBAT[i][0], c.DBAT[i][1])
	}
	return s
}

func batLine(kind string, i int, upper, lower uint32) string {
	if upper&3 == 0 {
		return ""
	}
	bl := (upper >> 2) & 0x7FF
	size := (uint64(bl) + 1) * 128 * 1024
	return fmt.Sprintf("%s%d  0x%08X..0x%08X -> 0x%08X  (%d KiB)\n",
		kind, i, upper&0xFFFE0000, uint64(upper&0xFFFE0000)+size, lower&0xFFFE0000, size/1024)
}
