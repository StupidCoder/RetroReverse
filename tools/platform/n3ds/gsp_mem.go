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

// hidSharedSize is the HID shared-memory block size (one page). It holds the pad,
// touch-screen, accelerometer and gyroscope state rings the game polls each frame.
const hidSharedSize = 0x1000

// The console's dedicated video RAM. GPU render targets live here. Two address
// spaces name it: the GPU (and the physical addresses a command list carries,
// e.g. the colour-buffer register) uses the physical base 0x18000000, while the
// application reads and writes it through a fixed virtual mapping at 0x1F000000
// — the capture shows the game's GX MemoryFill/DisplayTransfer/DMA commands all
// carrying 0x1F...... addresses, while the same buffers appear as 0x18......>>3
// in the command lists' framebuffer registers.
const (
	vramPhysBase = 0x18000000
	vramVirtBase = 0x1F000000
	vramSize     = 0x00600000 // 6 MiB
)

// gpuAddrToVirt maps an address as the GPU sees it (physical: VRAM at
// 0x18000000, FCRAM at 0x20000000) to the process virtual address this machine
// has mapped (VRAM at 0x1F000000, the linear heap at 0x14000000 — the linear
// heap's virtual↔physical offset is the fixed 0x0C000000 the kernel gives an
// application). Addresses already in the virtual windows pass through: the
// game's GX commands carry virtual addresses.
func (m *Machine) gpuAddrToVirt(a uint32) uint32 {
	switch {
	case a >= vramPhysBase && a < vramPhysBase+vramSize:
		return a - vramPhysBase + vramVirtBase
	case a >= 0x20000000 && a < 0x28000000:
		return a - 0x0C000000
	}
	return a
}

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
	if obj.kind == "hid-shared" {
		m.hidSharedAddr = addr
	}
	if m.Verbose {
		fmt.Printf("  MapMemoryBlock handle=0x%08X (%s) -> 0x%08X size=0x%X\n", handle, obj.kind, addr, size)
	}
	c.R[0] = resultSuccess
}
