package uw

// True VGA memory: four 64 KiB planes behind the A000 window, with the
// sequencer / graphics-controller / CRTC registers that steer it. This is what
// settles the "four repeated images" question definitively: if the game
// unchains the VGA (Mode X), byte writes reach only the planes selected by the
// sequencer map mask, and the flat-RAM model loses three of every four pixels
// (the display would interleave the planes 4-pixels-per-address). Modelling the
// planes and honouring chain-4 / map-mask / read-map / write-mode-1 latches
// reproduces exactly what a real VGA stores, and Screenshot() can then
// reconstruct the single true image from the plane data and the CRTC start
// address + pitch.
//
// Implemented: chain-4 (mode 13h) addressing; unchained write modes 0 (with
// set/reset, enable set/reset and bit mask), 1 (latch copy) and 2; read mode 0
// (with latch load) and read mode 1 (colour compare). Not modelled: the data
// rotator and ALU function select (gc[3]) — rarely used by games.

type vgaState struct {
	planes  [4][0x10000]byte
	latch   [4]byte
	seqIdx  byte
	seq     [8]byte
	gcIdx   byte
	gc      [16]byte
	crtcIdx byte
	crtc    [32]byte
}

func (m *Machine) vgaInit13h(clear bool) {
	v := m.vga
	v.seq[2] = 0x0F // map mask: all planes
	v.seq[4] = 0x0E // chain-4 on, extended memory
	v.gc[4] = 0     // read map select
	v.gc[5] = 0x40  // 256-colour shift, write mode 0, read mode 0
	v.gc[8] = 0xFF  // bit mask
	v.crtc[0x0C], v.crtc[0x0D] = 0, 0
	v.crtc[0x13] = 0x28 // pitch: 40 words = 80 bytes per scanline per plane
	if clear {
		for p := range v.planes {
			for i := range v.planes[p] {
				v.planes[p][i] = 0
			}
		}
	}
}

func (v *vgaState) chained() bool { return v.seq[4]&0x08 != 0 }

// vgaWrite handles a CPU byte write into the A000 window.
func (m *Machine) vgaWrite(a uint32, val byte) {
	v := m.vga
	off := a - 0xA0000
	if v.chained() {
		// Chain-4: the low two address bits select the plane.
		v.planes[off&3][(off>>2)&0xFFFF] = val
		return
	}
	off &= 0xFFFF
	mask := v.seq[2] & 0x0F
	switch v.gc[5] & 3 {
	case 1: // write mode 1: copy the latches (VRAM-to-VRAM blits)
		for p := 0; p < 4; p++ {
			if mask&(1<<p) != 0 {
				v.planes[p][off] = v.latch[p]
			}
		}
	case 2: // write mode 2: CPU data bits 0-3 expand to plane colours
		bm := v.gc[8]
		for p := 0; p < 4; p++ {
			if mask&(1<<p) == 0 {
				continue
			}
			var d byte
			if val&(1<<p) != 0 {
				d = 0xFF
			}
			v.planes[p][off] = (d & bm) | (v.latch[p] &^ bm)
		}
	default: // write mode 0 (and 3, approximated): set/reset + bit mask
		bm := v.gc[8]
		sr, esr := v.gc[0], v.gc[1]
		for p := 0; p < 4; p++ {
			if mask&(1<<p) == 0 {
				continue
			}
			d := val
			if esr&(1<<p) != 0 {
				if sr&(1<<p) != 0 {
					d = 0xFF
				} else {
					d = 0
				}
			}
			v.planes[p][off] = (d & bm) | (v.latch[p] &^ bm)
		}
	}
}

// vgaRead handles a CPU byte read from the A000 window (loads the latches).
func (m *Machine) vgaRead(a uint32) byte {
	v := m.vga
	off := a - 0xA0000
	if v.chained() {
		return v.planes[off&3][(off>>2)&0xFFFF]
	}
	off &= 0xFFFF
	for p := 0; p < 4; p++ {
		v.latch[p] = v.planes[p][off]
	}
	if v.gc[5]&0x08 != 0 { // read mode 1: colour compare
		cmp, dc := v.gc[2], v.gc[7]
		var out byte = 0xFF
		for p := 0; p < 4; p++ {
			if dc&(1<<p) == 0 {
				continue // "don't care" planes match anything
			}
			var want byte
			if cmp&(1<<p) != 0 {
				want = 0xFF
			}
			out &= ^(v.latch[p] ^ want)
		}
		return out
	}
	return v.latch[v.gc[4]&3] // read mode 0
}

// vgaRegOut handles a byte OUT to the VGA index/data register ports.
func (m *Machine) vgaRegOut(port uint16, b byte) {
	v := m.vga
	switch port {
	case 0x3C4:
		v.seqIdx = b
	case 0x3C5:
		if v.seqIdx&7 == 4 { // memory-mode register: log chain-4 transitions
			was := v.chained()
			v.seq[4] = b
			if was != v.chained() {
				m.logf("VGA: chain-4 %v -> %v (seq[4]=%02X) at %04X:%04X",
					was, v.chained(), b, m.CPU.Seg[1], m.CPU.IP)
			}
			return
		}
		v.seq[v.seqIdx&7] = b
	case 0x3CE:
		v.gcIdx = b
	case 0x3CF:
		v.gc[v.gcIdx&15] = b
	case 0x3D4:
		v.crtcIdx = b
	case 0x3D5:
		v.crtc[v.crtcIdx&31] = b
	}
}
