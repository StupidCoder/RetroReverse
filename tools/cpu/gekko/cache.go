package gekko

// cache.go is the cache instructions, and the one cache on this machine that is not an
// invisible optimisation but a piece of addressable hardware: the locked cache.
//
// Setting HID2[LCE] takes half of the 32 KiB L1 data cache out of the coherency protocol
// and maps it at a fixed address, by convention 0xE0000000. From then on it is 16 KiB of
// single-cycle memory that no bus transaction ever touches — a scratchpad. A program
// fills a line with dcbz_l (which allocates it, zeroed, *without* reading main memory),
// or it moves whole lines in and out with a DMA engine driven through two SPRs.
//
// This is the reason the locked cache lives in the CPU rather than the machine, and the
// reason it is in the savestate: it holds data, that data is not a copy of anything in
// main memory, and losing it loses the program's working set. A model that treated it as
// a cache — a transparent copy of memory that may be dropped — would be silently wrong
// in exactly the cases the program uses it for.
//
// The rest of the cache instructions are, on this core, either no-ops or memory writes.
// There is no write-back cache to flush, so dcbf and dcbst have nothing to do; icbi has
// nothing to invalidate. dcbz is not a hint at all, though: it *writes zeroes*, a whole
// 32-byte line of them, and a GameCube program uses it to clear memory quickly. Treating
// it as a no-op leaves stale data where the program expects zeroes.

// The cache line is 32 bytes on this processor, and every cache instruction operates on
// the line containing its address rather than the address itself.
const CacheLine = 32

// LockedCacheBase is where the locked cache is mapped. The address is not fixed by the
// hardware — HID2 does not name it — but it is where every GameCube program puts it, and
// the DMA registers address it as though it were.
const (
	LockedCacheBase = 0xE0000000
	LockedCacheSize = 16 * 1024
)

// LockedCache is the scratchpad.
type LockedCache struct {
	Data     [LockedCacheSize]byte
	Base     uint32
	Enabled_ bool
}

func (l *LockedCache) Reset() {
	l.Data = [LockedCacheSize]byte{}
	l.Base = LockedCacheBase
	l.Enabled_ = false
}

func (l *LockedCache) Enabled() bool { return l.Enabled_ }

// Contains reports whether an address falls in the scratchpad. It is only true once
// software has enabled it, so a program that has not turned the cache on reaches memory
// at that address like anything else — which is what the hardware does.
func (l *LockedCache) Contains(a uint32) bool {
	return l.Enabled_ && a >= l.Base && a < l.Base+LockedCacheSize
}

func (l *LockedCache) off(a uint32) uint32 { return (a - l.Base) & (LockedCacheSize - 1) }

func (l *LockedCache) Read8(a uint32) uint8 { return l.Data[l.off(a)] }

func (l *LockedCache) Read16(a uint32) uint16 {
	o := l.off(a)
	return uint16(l.Data[o])<<8 | uint16(l.Data[o+1])
}

func (l *LockedCache) Read32(a uint32) uint32 {
	o := l.off(a)
	return uint32(l.Data[o])<<24 | uint32(l.Data[o+1])<<16 | uint32(l.Data[o+2])<<8 | uint32(l.Data[o+3])
}

func (l *LockedCache) Write8(a uint32, v uint8) { l.Data[l.off(a)] = v }

func (l *LockedCache) Write16(a uint32, v uint16) {
	o := l.off(a)
	l.Data[o], l.Data[o+1] = uint8(v>>8), uint8(v)
}

func (l *LockedCache) Write32(a uint32, v uint32) {
	o := l.off(a)
	l.Data[o], l.Data[o+1] = uint8(v>>24), uint8(v>>16)
	l.Data[o+2], l.Data[o+3] = uint8(v>>8), uint8(v)
}

// setHID2 acts on a write to HID2, which is where the locked cache is switched on.
func (c *CPU) setHID2(v uint32) {
	c.HID2 = v
	c.LC.Enabled_ = v&HID2LCE != 0
}

// dcbz zeroes the cache line containing ea — a real write of 32 zero bytes to memory,
// not a hint.
func (c *CPU) dcbz(ea uint32) {
	base := ea &^ (CacheLine - 1)
	for i := uint32(0); i < CacheLine; i += 4 {
		c.write32(base+i, 0)
	}
}

// dcbzL allocates a line in the locked cache, zeroed, without touching main memory.
// This is how a program starts using the scratchpad: there is nothing to load from, so
// the line is simply created.
func (c *CPU) dcbzL(ea uint32) {
	if !c.LC.Enabled() {
		c.Halt("gekko: dcbz_l at PC 0x%08X but the locked cache is not enabled (HID2 = 0x%08X)", c.PC, c.HID2)
		return
	}
	if !c.LC.Contains(ea) {
		c.Halt("gekko: dcbz_l on 0x%08X, which is outside the locked cache at 0x%08X", ea, c.LC.Base)
		return
	}
	base := ea &^ (CacheLine - 1)
	o := c.LC.off(base)
	for i := uint32(0); i < CacheLine; i++ {
		c.LC.Data[o+i] = 0
	}
}

// runDMA services a write to DMAL, which is what triggers the locked cache's DMA engine.
//
// The pair of registers is packed tightly: DMAU holds the memory address and the top of
// the length, DMAL holds the cache address, the bottom of the length, the direction and
// the trigger. A transfer moves whole 32-byte lines, and it completes instantly here —
// there is no queue to model, because a program that starts one waits for it.
func (c *CPU) runDMA() {
	if c.DMAL&2 == 0 { // the trigger bit
		return
	}
	memAddr := c.DMAU & 0xFFFFFFE0
	lcAddr := c.DMAL & 0xFFFFFFE0
	// The length is 5 bits in DMAU's low bits and 2 in DMAL's, counted in 32-byte lines.
	length := ((c.DMAU & 0x1F) << 2) | ((c.DMAL >> 2) & 3)
	if length == 0 {
		length = 128
	}
	load := c.DMAL&0x10 != 0 // 1 = memory -> cache, 0 = cache -> memory

	if !c.LC.Enabled() {
		c.Halt("gekko: a locked-cache DMA was triggered but the cache is not enabled (HID2 = 0x%08X)", c.HID2)
		return
	}
	for i := uint32(0); i < length*CacheLine; i++ {
		if load {
			c.LC.Data[c.LC.off(lcAddr+i)] = c.bus.Read8(c.memPhys(memAddr + i))
		} else {
			c.bus.Write8(c.memPhys(memAddr+i), c.LC.Data[c.LC.off(lcAddr+i)])
		}
	}
	c.DMAL &^= 2 // the transfer is done
}

// memPhys translates a DMA's memory address. The engine sees the same effective addresses
// the program does, so it goes through the block translations like anything else.
func (c *CPU) memPhys(ea uint32) uint32 {
	pa, ok := c.Translate(ea, false, false)
	if !ok {
		return 0
	}
	return pa
}
