package threedo

import "fmt"

// folio.go reimplements the Portfolio OS folio/kernel functions the game calls,
// so the boot oracle can run past the OS-dependent startup instead of merely
// logging stubbed calls. Each function is identified from LaunchMe's own code
// (clean-room: derived from the disc, not external sources) by the negative
// offset it is called at from the kernel base (`LDR pc, [r9, #-N]`).
//
// Identified so far, from the game's use of the return values:
//
//	-0x1C  AllocMem(memlist, size, flags) -> ptr   (0 on failure)
//	-0x20  FreeMem(memlist, ptr, size)
//
// The rest are still stubbed (return 0) and logged; each is named as its use is
// pinned down. AllocMem/FreeMem are backed by a real first-fit heap over a
// reserved DRAM region, so the game's binary-search memory probe converges on a
// real free size and startup proceeds.

// span is one free region of the HLE heap.
type span struct{ addr, size uint32 }

// heap is a minimal first-fit allocator with coalescing, over [base,base+total).
type heap struct {
	base, total uint32
	free        []span
	live        map[uint32]uint32 // addr -> size of live allocations
}

func newHeap(base, total uint32) *heap {
	return &heap{base: base, total: total, free: []span{{base, total}}, live: map[uint32]uint32{}}
}

// alloc returns an aligned block of at least size bytes, or 0 if none fits.
func (h *heap) alloc(size uint32) uint32 {
	if size == 0 {
		return 0
	}
	size = (size + 15) &^ 15 // 16-byte align
	for i, s := range h.free {
		if s.size >= size {
			addr := s.addr
			if s.size == size {
				h.free = append(h.free[:i], h.free[i+1:]...)
			} else {
				h.free[i] = span{s.addr + size, s.size - size}
			}
			h.live[addr] = size
			return addr
		}
	}
	return 0
}

// freeBlock returns a block to the free list, coalescing neighbours.
func (h *heap) freeBlock(addr uint32) {
	size, ok := h.live[addr]
	if !ok {
		return // double free or foreign pointer: ignore
	}
	delete(h.live, addr)
	h.free = append(h.free, span{addr, size})
	// Coalesce: sort by address and merge adjacent spans.
	for i := 0; i < len(h.free); i++ {
		for j := i + 1; j < len(h.free); j++ {
			if h.free[j].addr < h.free[i].addr {
				h.free[i], h.free[j] = h.free[j], h.free[i]
			}
		}
	}
	merged := h.free[:1]
	for _, s := range h.free[1:] {
		last := &merged[len(merged)-1]
		if last.addr+last.size == s.addr {
			last.size += s.size
		} else {
			merged = append(merged, s)
		}
	}
	h.free = merged
}

// serviceFolio dispatches an intercepted kernel-folio call (a PC that landed in
// the HLE window) by its negative offset. It returns true if the call was
// reimplemented; false means it was left as a logged stub.
func (m *Machine) serviceFolio(off uint32) bool {
	switch off {
	case 0x1C: // AllocMem(memlist, size, flags)
		size := m.CPU.Reg(1)
		flags := m.CPU.Reg(2)
		ptr := m.poolFor(flags).alloc(size)
		m.note(fmt.Sprintf("AllocMem(size=0x%X, flags=0x%X) -> 0x%08X", size, flags, ptr))
		m.SetResultAndReturn(ptr)
		return true
	case 0x20: // FreeMem(memlist, ptr, size)
		ptr := m.CPU.Reg(1)
		m.poolOf(ptr).freeBlock(ptr)
		m.SetResultAndReturn(0)
		return true
	case 0x34: // SampleSystemTimeTT(timer, TimeVal*) — fill an advancing time.
		// Each call advances virtual time so the game's timing/calibration loops
		// (which decrement counters by the elapsed delta) converge instead of
		// spinning forever. The TimeVal is {seconds, microseconds}.
		m.simTime += simTick
		buf := m.CPU.Reg(0)
		m.writeWord(buf, uint32(m.simTime/1_000_000))
		m.writeWord(buf+4, uint32(m.simTime%1_000_000))
		m.SetResultAndReturn(0)
		return true
	}
	return false
}

// simTick is how much virtual time each time sample advances (a generous slice so
// timing loops finish quickly).
const simTick = 100_000 // 100 ms

// 3DO AllocMem memory-type flags (from the Portfolio SDK mem.h).
const (
	memtypeVRAM = 0x00010000
	memtypeDRAM = 0x00080000
)

// poolFor selects the allocation pool by AllocMem flags. Per mem.h,
// MEMTYPE_VRAM is 0x10000 and MEMTYPE_DRAM is 0x80000; anything else
// (MEMTYPE_ANY = 0, DMA/CEL/etc.) draws from DRAM.
func (m *Machine) poolFor(flags uint32) *heap {
	if flags&memtypeVRAM != 0 {
		return m.vheap
	}
	return m.dheap
}

// poolOf selects the pool a live pointer belongs to, by its address range.
func (m *Machine) poolOf(ptr uint32) *heap {
	if ptr >= vheapBase && ptr < vheapTop {
		return m.vheap
	}
	return m.dheap
}
