package xbox

// nv2a_ramht.go resolves the RAMHT hash-table handles the PFIFO pusher's NV_SET_OBJECT
// method carries into an object class. The Direct3D runtime creates a few graphics
// objects (the 3D/Kelvin class, plus 2D/surface helpers), inserts each into RAMHT keyed
// by a handle, and thereafter binds a subchannel to one by writing its handle to method
// 0. To know which class a subchannel drives, we look the handle up.
//
// RAMHT lives in PRAMIN (instance memory, the 0x700000 window of the MMIO aperture). Each
// entry is 8 bytes: the handle, then a context word whose low bits hold the object's
// instance (a PRAMIN offset in 16-byte units). The graphics object at that instance
// begins with a word whose low bits are its class. Rather than track the RAMHT base
// register, we scan the small RAMHT region for the handle — it holds only a handful of
// entries and the scan runs once per SET_OBJECT.

const (
	pramin = 0x700000 // PRAMIN window base within the NV2A aperture (0xFD700000)
	// The RAMHT region the Xbox kernel programs is at the start of PRAMIN. 4 KB covers the
	// handful of entries a title inserts with room to spare.
	ramhtScan = 0x1000
)

// ramhtInstance resolves a RAMHT handle to its object's PRAMIN address (0 if not found).
func (m *Machine) ramhtInstance(handle uint32) uint32 {
	if handle == 0 {
		return 0
	}
	base := uint32(0xFD000000 + pramin)
	for off := uint32(0); off < ramhtScan; off += 8 {
		if m.read32(base+off) != handle {
			continue
		}
		ctx := m.read32(base + off + 4)
		return base + (ctx&0xFFFF)<<4 // RAMIN offset of the object, in 16-byte units
	}
	return 0
}

// ramhtClass resolves a RAMHT handle to its object class (0 if not found).
func (m *Machine) ramhtClass(handle uint32) uint32 {
	inst := m.ramhtInstance(handle)
	if inst == 0 {
		return 0
	}
	return m.read32(inst) & 0xFF // the object's class is in the low byte of its first context word
}

// dmaObjectTarget decodes an NV "DMA in memory" context object (classes 0x02/0x03/0x3D)
// into the physical base address and byte limit it grants access to. Layout read off
// this title's own PRAMIN (instances 0x1180-0x11C0): word0 = class | flags with the
// page-offset "adjust" in bits 20-31, word1 = limit (bytes-1), word2 = page-aligned
// address with target flags in the low bits. The pusher's channel DMA (base 0, limit
// 0x07FFAFFF) and the three 32-byte semaphore surfaces at 0x14C8000/20/40 all decode
// consistently under this reading.
func (m *Machine) dmaObjectTarget(handle uint32) (base, limit uint32) {
	inst := m.ramhtInstance(handle)
	if inst == 0 {
		return 0, 0
	}
	w0 := m.read32(inst)
	w1 := m.read32(inst + 4)
	w2 := m.read32(inst + 8)
	return (w2 &^ 0xFFF) + (w0 >> 20 & 0xFFF), w1
}
