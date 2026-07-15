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

// ramhtClass resolves a RAMHT handle to its object class (0 if not found).
func (m *Machine) ramhtClass(handle uint32) uint32 {
	if handle == 0 {
		return 0
	}
	base := uint32(0xFD000000 + pramin)
	for off := uint32(0); off < ramhtScan; off += 8 {
		if m.read32(base+off) != handle {
			continue
		}
		ctx := m.read32(base + off + 4)
		inst := (ctx & 0xFFFF) << 4 // RAMIN offset of the object, in 16-byte units
		obj0 := m.read32(base + inst)
		return obj0 & 0xFF // the object's class is in the low byte of its first context word
	}
	return 0
}
