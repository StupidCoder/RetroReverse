package n3ds

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
)

// gsp_mem.go backs the shared-memory blocks a title maps into its address space
// — chiefly the GSP shared memory, through which the graphics service and the
// application exchange the GX command queue, the framebuffer descriptors, and
// the interrupt/VBlank relay. A memory-block handle (from svcCreateMemoryBlock,
// or handed back by a service such as gsp: RegisterInterruptRelayQueue) is bound
// to real bytes here so the game reads and writes an actual region rather than
// faulting.

// gspSharedSize is the GSP shared-memory block size (one page). It holds, per
// GSP thread: the interrupt relay queue, the GX command FIFO, and the
// framebuffer-info structures.
const gspSharedSize = 0x1000

// vramBase/vramSize: the console's dedicated video RAM. GPU render targets and
// the display framebuffers live here; DisplayTransfer/MemoryFill address it.
const (
	vramBase = 0x18000000
	vramSize = 0x00600000 // 6 MiB
)

// svcMapMemoryBlock (0x1F) maps a memory-block handle into this process at the
// requested address. ABI (from the wrapper at 0x0010B298): r0=handle, r1=addr,
// r2=my permissions, r3=other permissions. The block is backed by a fresh zeroed
// region of its size; the mapped address is recorded so the kernel side (VBlank
// delivery into the GSP relay queue) can reach it.
func (m *Machine) svcMapMemoryBlock(c *arm.CPU) {
	handle, addr := c.R[0], c.R[1]
	obj := m.handles[handle]
	if obj == nil {
		c.Halt("MapMemoryBlock: unknown handle 0x%08X at 0x%08X after %d instructions", handle, c.PC(), c.Instrs)
		return
	}
	size := obj.blockSize
	if size == 0 {
		size = gspSharedSize
	}
	// Back the block with real bytes at addr. If it already has a mapped region
	// (a second mapping of an app-created block), alias those bytes; otherwise
	// allocate a fresh zeroed page-multiple region.
	if obj.blockAddr != 0 {
		if r := m.regionOf(obj.blockAddr); r != nil {
			m.mapRegion(obj.kind, addr, r.data)
		}
	} else {
		m.mapRegion(obj.kind, addr, make([]byte, size))
	}
	obj.blockAddr = addr
	obj.blockSize = size
	if obj.kind == "gsp-shared" {
		m.gspSharedAddr = addr
	}
	if m.Verbose {
		fmt.Printf("  MapMemoryBlock handle=0x%08X (%s) -> 0x%08X size=0x%X\n", handle, obj.kind, addr, size)
	}
	c.R[0] = resultSuccess
}
