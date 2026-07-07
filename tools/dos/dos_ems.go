package dos

// A minimal LIM EMS (Expanded Memory) emulation. Ultima Underworld requires an
// EMS driver — without one it prints "Out of EMS Memory" and quits — so the
// oracle provides one: a 64 KiB page frame of four 16 KiB physical pages, a
// pool of expanded pages, and the INT 67h function set the game uses (status,
// page-frame/page-count queries, allocate, map, save/restore mapping, free).
//
// Mapping is emulated by copy-on-(re)map: a frame slot physically holds the
// bytes of whichever expanded page is mapped into it; remapping first flushes
// the slot back to its expanded page, then copies the newly-mapped page in. So
// long as the program only touches a frame slot while it is mapped (the EMS
// contract), this is faithful with a flat memory model.

import "retroreverse.com/tools/x86"

const (
	emsPageSize  = 16384 // 16 KiB per EMS page
	emsFrameSeg  = 0xE000
	emsSigSeg    = 0xC000 // where the fake EMM driver header + "EMMXXXX0" name live
	emsTotalPage = 512    // 8 MiB of expanded memory
)

type emsHandle struct{ base, count int }

type emsState struct {
	backing  []byte
	nextPage int // bump allocator over the expanded-page pool
	handles  map[uint16]emsHandle
	nextH    uint16
	slot     [4]int            // expanded page currently in each frame slot, -1 = none
	saved    map[uint16][4]int // per-handle saved mappings (INT 67h/47h,48h)
}

// setupEMS installs the fake EMM driver signature and initialises the pool. The
// game detects EMS by taking the segment of the INT 67h vector and checking for
// the device name "EMMXXXX0" at offset 0x0A of that segment.
func (m *Machine) setupEMS() {
	m.ems = &emsState{
		backing:  make([]byte, emsTotalPage*emsPageSize),
		nextPage: 0,
		handles:  map[uint16]emsHandle{},
		nextH:    1, // handle 0 is the system handle
		slot:     [4]int{-1, -1, -1, -1},
		saved:    map[uint16][4]int{},
	}
	// A character-device driver header at emsSigSeg:0000, name at +0x0A.
	base := uint32(emsSigSeg) << 4
	copy(m.Mem[base+0x0A:], []byte("EMMXXXX0"))
	// Point IVT[0x67] at the driver segment (the handler itself is emulated).
	m.w16(0x67*4, 0x0000)
	m.w16(0x67*4+2, emsSigSeg)
}

func (e *emsState) frameLin(slot int) uint32 { return emsFrameSeg<<4 + uint32(slot)*emsPageSize }

// EmsBacking exposes the expanded-memory backing store (all EMS pages) for
// diagnostics such as searching for a loaded resource that lives in EMS.
func (m *Machine) EmsBacking() []byte {
	if m.ems == nil {
		return nil
	}
	return m.ems.backing
}

// EmsHandleBase returns the first backing-pool page of EMS handle h (or -1). A
// handle's logical page N lives at backing[(base+N)*emsPageSize ...]; combined
// with EmsBacking this lets a caller read a resource straight from expanded
// memory without going through the frame mapping.
func (m *Machine) EmsHandleBase(h uint16) int {
	if m.ems == nil {
		return -1
	}
	if hd, ok := m.ems.handles[h]; ok {
		return hd.base
	}
	return -1
}

// EmsPageSize is the size of one expanded-memory page (16 KiB).
const EmsPageSize = emsPageSize

// flush copies a mapped frame slot back to its expanded page.
func (m *Machine) emsFlush(slot int) {
	p := m.ems.slot[slot]
	if p < 0 {
		return
	}
	copy(m.ems.backing[p*emsPageSize:], m.Mem[m.ems.frameLin(slot):m.ems.frameLin(slot)+emsPageSize])
}

// load maps expanded page p into frame slot (after the caller has flushed it).
func (m *Machine) emsLoad(slot, p int) {
	copy(m.Mem[m.ems.frameLin(slot):m.ems.frameLin(slot)+emsPageSize], m.ems.backing[p*emsPageSize:(p+1)*emsPageSize])
	m.ems.slot[slot] = p
}

// int67 services an EMS call. Returns AH=0 on success or an EMS error code.
func (m *Machine) int67(c *x86.CPU) bool {
	e := m.ems
	ah := c.Reg8(x86.AH)
	ok := func() { c.SetReg8(x86.AH, 0) }
	fail := func(code byte) { c.SetReg8(x86.AH, code) }

	switch ah {
	case 0x40: // get manager status
		ok()
	case 0x41: // get page-frame segment
		c.SetReg16(x86.BX, emsFrameSeg)
		ok()
	case 0x42: // get number of pages
		c.SetReg16(x86.BX, uint16(emsTotalPage-e.nextPage)) // free
		c.SetReg16(x86.DX, uint16(emsTotalPage))            // total
		ok()
	case 0x43: // allocate handle + pages
		want := int(c.Reg16(x86.BX))
		if want == 0 {
			want = 0 // some callers allocate 0 then reallocate; allow it
		}
		if e.nextPage+want > emsTotalPage {
			fail(0x88) // not enough pages available
			return true
		}
		h := e.nextH
		e.nextH++
		e.handles[h] = emsHandle{base: e.nextPage, count: want}
		e.nextPage += want
		c.SetReg16(x86.DX, h)
		m.logf("EMS alloc handle %d: %d pages (%d KiB)", h, want, want*16)
		ok()
	case 0x44: // map handle page: AL=phys slot, BX=logical page, DX=handle
		phys := int(c.Reg8(x86.AL))
		logical := c.Reg16(x86.BX)
		h, exists := e.handles[c.Reg16(x86.DX)]
		if !exists {
			fail(0x83)
			return true
		}
		if phys > 3 {
			fail(0x8B)
			return true
		}
		if logical == 0xFFFF { // unmap
			m.emsFlush(phys)
			e.slot[phys] = -1
			ok()
			return true
		}
		if int(logical) >= h.count {
			fail(0x8A)
			return true
		}
		m.emsFlush(phys)
		m.emsLoad(phys, h.base+int(logical))
		ok()
	case 0x45: // free handle
		delete(e.handles, c.Reg16(x86.DX))
		ok()
	case 0x46: // get EMM version
		c.SetReg8(x86.AL, 0x40) // 4.0
		ok()
	case 0x47: // save page map for handle
		e.saved[c.Reg16(x86.DX)] = e.slot
		ok()
	case 0x48: // restore page map for handle
		if s, exsts := e.saved[c.Reg16(x86.DX)]; exsts {
			for slot := 0; slot < 4; slot++ {
				m.emsFlush(slot)
			}
			for slot := 0; slot < 4; slot++ {
				if s[slot] >= 0 {
					m.emsLoad(slot, s[slot])
				} else {
					e.slot[slot] = -1
				}
			}
		}
		ok()
	case 0x4B: // get number of EMM handles
		c.SetReg16(x86.BX, uint16(len(e.handles)))
		ok()
	case 0x4C: // get pages owned by handle
		if h, exsts := e.handles[c.Reg16(x86.DX)]; exsts {
			c.SetReg16(x86.BX, uint16(h.count))
			ok()
		} else {
			fail(0x83)
		}
	case 0x4D: // get pages for all handles
		di := c.Reg16(x86.DI)
		dst := lin(c.Seg[x86.ES], di)
		n := 0
		for h, hd := range e.handles {
			m.w16(dst+uint32(n)*4, h)
			m.w16(dst+uint32(n)*4+2, uint16(hd.count))
			n++
		}
		c.SetReg16(x86.BX, uint16(n))
		ok()
	case 0x4E: // get/set page map (context switch)
		m.emsPageMap(c)
	case 0x50: // EMS 4.0: map/unmap multiple handle pages
		// AL=0: DS:SI -> CX pairs of {logical page, physical page} words.
		// AL=1: physical given as segment addresses instead of page numbers.
		h, exists := e.handles[c.Reg16(x86.DX)]
		if !exists {
			fail(0x83)
			return true
		}
		src := lin(c.Seg[x86.DS], c.Reg16(x86.SI))
		n := int(c.Reg16(x86.CX))
		for i := 0; i < n; i++ {
			logical := m.r16(src + uint32(i)*4)
			physRaw := m.r16(src + uint32(i)*4 + 2)
			phys := int(physRaw)
			if c.Reg8(x86.AL) == 1 { // segment form -> frame slot index
				phys = int(physRaw-emsFrameSeg) * 16 / emsPageSize
			}
			if phys < 0 || phys > 3 {
				fail(0x8B)
				return true
			}
			if logical == 0xFFFF { // unmap
				m.emsFlush(phys)
				e.slot[phys] = -1
				continue
			}
			if int(logical) >= h.count {
				fail(0x8A)
				return true
			}
			m.emsFlush(phys)
			m.emsLoad(phys, h.base+int(logical))
		}
		ok()
	case 0x51: // reallocate pages for handle (EMS 4.0)
		h := c.Reg16(x86.DX)
		want := int(c.Reg16(x86.BX))
		if hd, exsts := e.handles[h]; exsts {
			// Simple model: only grow the topmost handle; otherwise keep count.
			hd.count = want
			e.handles[h] = hd
			c.SetReg16(x86.BX, uint16(want))
			ok()
		} else {
			fail(0x83)
		}
	case 0x53: // get/set handle name — accept
		ok()
	case 0x58: // get mappable physical address array / count
		if c.Reg8(x86.AL) == 1 { // get count
			c.SetReg16(x86.CX, 4)
		}
		ok()
	case 0x59: // get hardware config / raw page count
		c.SetReg16(x86.BX, uint16(emsTotalPage-e.nextPage))
		c.SetReg16(x86.DX, uint16(emsTotalPage))
		ok()
	default:
		m.logf("EMS: unhandled INT 67h AH=%02X AL=%02X BX=%04X DX=%04X", ah, c.Reg8(x86.AL), c.Reg16(x86.BX), c.Reg16(x86.DX))
		ok() // assume success so the boot proceeds; revisit if it misbehaves
	}
	return true
}

// emsPageMap implements INT 67h/4Eh (save/restore the whole page-map, used
// around context switches). The map is the four slot page-indices as words.
func (m *Machine) emsPageMap(c *x86.CPU) {
	e := m.ems
	switch c.Reg8(x86.AL) {
	case 0x00, 0x02: // get (0) or get-and-set (2): write current map to ES:DI
		dst := lin(c.Seg[x86.ES], c.Reg16(x86.DI))
		for i := 0; i < 4; i++ {
			m.emsFlush(i)
			m.w16(dst+uint32(i)*2, uint16(int16(e.slot[i])))
		}
		if c.Reg8(x86.AL) == 0x00 {
			c.SetReg8(x86.AH, 0)
			return
		}
		fallthrough
	case 0x01: // set: restore map from DS:SI
		src := lin(c.Seg[x86.DS], c.Reg16(x86.SI))
		for i := 0; i < 4; i++ {
			m.emsFlush(i)
		}
		for i := 0; i < 4; i++ {
			p := int(int16(m.r16(src + uint32(i)*2)))
			if p >= 0 {
				m.emsLoad(i, p)
			} else {
				e.slot[i] = -1
			}
		}
		c.SetReg8(x86.AH, 0)
	case 0x03: // get size of save array
		c.SetReg8(x86.AL, 8) // four words
		c.SetReg8(x86.AH, 0)
	}
}
