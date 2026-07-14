package dsmachine

import "strings"

// The DS's buttons. Ten of them live in KEYINPUT, which both CPUs can read; X and
// Y arrived with the DS and did not fit, so they were put in EXTKEYIN alongside the
// pen-down and lid sensors — and EXTKEYIN is wired to the ARM7 only. A game's ARM9
// therefore never reads X or Y directly: it asks the ARM7 for them over the IPC
// FIFO, which is why an oracle that injects keys has to inject them into the
// hardware and let the game's own two halves talk, rather than pretend at either
// end.
//
// Both registers are active LOW: a bit reads 0 while its button is held.

const (
	regKEYINPUT = 0x04000130
	regKEYCNT   = 0x04000132
	regRCNT     = 0x04000134 // the word EXTKEYIN lives in, as its high half
	regEXTKEYIN = 0x04000136
)

// Key bits, in KEYINPUT order, plus the two that live in EXTKEYIN.
const (
	KeyA = 1 << iota
	KeyB
	KeySelect
	KeyStart
	KeyRight
	KeyLeft
	KeyUp
	KeyDown
	KeyR
	KeyL
	KeyX // EXTKEYIN bit 0
	KeyY // EXTKEYIN bit 1
)

var keyNames = map[string]uint32{
	"a": KeyA, "b": KeyB, "x": KeyX, "y": KeyY,
	"l": KeyL, "r": KeyR,
	"start": KeyStart, "select": KeySelect,
	"up": KeyUp, "down": KeyDown, "left": KeyLeft, "right": KeyRight,
}

// ParseKeys turns a comma-separated list ("a,start,up") into a key mask.
func ParseKeys(s string) (uint32, bool) {
	var mask uint32
	for _, f := range strings.Split(s, ",") {
		f = strings.ToLower(strings.TrimSpace(f))
		if f == "" {
			continue
		}
		k, ok := keyNames[f]
		if !ok {
			return 0, false
		}
		mask |= k
	}
	return mask, true
}

// SetKeys sets which buttons are held (a mask of the Key* constants).
func (m *Machine) SetKeys(mask uint32) { m.keys = mask }

// keyinput assembles KEYINPUT: the ten main buttons, active low.
func (m *Machine) keyinput() uint32 {
	return ^m.keys & 0x3FF
}

// extkeyin assembles EXTKEYIN (ARM7 only): X and Y, the pen-down sensor and the lid.
//
// Every bit here is inverted, and one of them is inverted the other way. X, Y and the
// pen read 0 when pressed/down, like the main buttons. The LID DOES NOT: bit 7 reads
// 0 when the lid is OPEN and 1 when it is CLOSED. Set it the "obvious" way round —
// 1 for open, matching every other bit's "1 means not pressed" — and the game reads a
// console that has been folded shut, and its ARM7 does exactly what it should: it
// puts the machine into deep sleep, mid-boot, and the oracle stops with the whole
// thing working perfectly and nothing on the screen.
func (m *Machine) extkeyin() uint32 {
	v := uint32(0x007F) // nothing pressed; bit 7 clear = the lid is open
	if m.keys&KeyX != 0 {
		v &^= 1 << 0
	}
	if m.keys&KeyY != 0 {
		v &^= 1 << 1
	}
	if m.spi.touchDown {
		v &^= 1 << 6
	}
	return v
}
