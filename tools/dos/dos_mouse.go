package dos

// Microsoft-compatible mouse driver (INT 33h). Ultima Underworld's whole UI —
// the menus, character creation, and in-world interaction — is mouse-driven. It
// polls the driver rather than installing a callback: function 0Bh reads the
// relative motion counters (mickeys) and function 03h reads the absolute
// position and button state. The game integrates the 0Bh deltas to move its own
// crosshair cursor, so injection drives both consistently: a scripted move sets
// the absolute position AND feeds the matching delta through the motion
// counters, and a scripted click sets the button mask 03h reports.
//
// Coordinates are mode-13h screen pixels (0..319 × 0..199). The classic driver
// keeps an internal 640×200 virtual space; UW sets its own range via 07h/08h,
// which we honour for clamping.

import "retroreverse.com/tools/x86"

// mickeysPerPixel scales an injected pixel move into the motion-counter units
// the game integrates. Calibrated against where UW's own crosshair lands: it
// advances its cursor ~1 screen pixel per 2 mickeys, so 2 mickeys/pixel gives a
// 1:1 mapping from requested screen pixels to cursor pixels.
const mickeysPerPixel = 2

type mouseState struct {
	x, y       int  // absolute position, mode-13h pixels
	accX, accY int  // motion counters (mickeys) accumulated since the last 0Bh
	buttons    byte // bit0 = left, bit1 = right, bit2 = middle
	xMin, xMax int
	yMin, yMax int
	hidden     bool
	pressL     int // left-button press counter (function 05h)
	pressR     int
}

func (m *Machine) mouse() *mouseState {
	if m.ms == nil {
		m.ms = &mouseState{x: 160, y: 100, xMin: 0, xMax: 319, yMin: 0, yMax: 199}
	}
	return m.ms
}

func (m *Machine) int33(c *x86.CPU) bool {
	ms := m.mouse()
	switch c.Reg16(x86.AX) {
	case 0x0000: // reset driver and read status
		ms.x, ms.y = 160, 100
		ms.accX, ms.accY, ms.buttons = 0, 0, 0
		ms.hidden = true
		c.SetReg16(x86.AX, 0xFFFF) // driver installed
		c.SetReg16(x86.BX, 2)      // two buttons
	case 0x0001: // show cursor
		ms.hidden = false
	case 0x0002: // hide cursor
		ms.hidden = true
	case 0x0003: // get position and button status
		c.SetReg16(x86.CX, uint16(ms.x))
		c.SetReg16(x86.DX, uint16(ms.y))
		c.SetReg16(x86.BX, uint16(ms.buttons))
	case 0x0004: // set cursor position
		ms.x, ms.y = int(int16(c.Reg16(x86.CX))), int(int16(c.Reg16(x86.DX)))
	case 0x0005: // get button press info: BX=button, returns count + last-press pos
		cnt := ms.pressL
		if c.Reg16(x86.BX) == 1 {
			cnt = ms.pressR
			ms.pressR = 0
		} else {
			ms.pressL = 0
		}
		c.SetReg16(x86.AX, uint16(ms.buttons))
		c.SetReg16(x86.BX, uint16(cnt))
		c.SetReg16(x86.CX, uint16(ms.x))
		c.SetReg16(x86.DX, uint16(ms.y))
	case 0x0006: // get button release info
		c.SetReg16(x86.AX, uint16(ms.buttons))
		c.SetReg16(x86.BX, 0)
		c.SetReg16(x86.CX, uint16(ms.x))
		c.SetReg16(x86.DX, uint16(ms.y))
	case 0x0007: // set horizontal range
		ms.xMin, ms.xMax = int(c.Reg16(x86.CX)), int(c.Reg16(x86.DX))
	case 0x0008: // set vertical range
		ms.yMin, ms.yMax = int(c.Reg16(x86.CX)), int(c.Reg16(x86.DX))
	case 0x000B: // read motion counters (relative), then clear them
		c.SetReg16(x86.CX, uint16(int16(ms.accX)))
		c.SetReg16(x86.DX, uint16(int16(ms.accY)))
		ms.accX, ms.accY = 0, 0
	case 0x000C, 0x0014: // set event handler — UW polls instead; accept and ignore
	case 0x000F, 0x001A, 0x001C: // set mickey/pixel ratio, sensitivity, interrupt rate
	default:
		// Unhandled sub-function: report the driver present and continue.
	}
	return true
}

// clampMouse keeps the absolute position inside the game-set range.
func (ms *mouseState) clamp() {
	if ms.x < ms.xMin {
		ms.x = ms.xMin
	}
	if ms.x > ms.xMax {
		ms.x = ms.xMax
	}
	if ms.y < ms.yMin {
		ms.y = ms.yMin
	}
	if ms.y > ms.yMax {
		ms.y = ms.yMax
	}
}

// HomeMouse slams the cursor to the top-left corner by feeding a motion delta
// large enough that the game clamps its own cursor at (0,0), giving injection a
// known origin from which the next MoveMouseTo delta is exact.
func (m *Machine) HomeMouse() {
	ms := m.mouse()
	ms.accX -= 4000
	ms.accY -= 4000
	ms.x, ms.y = 0, 0
}

// MoveMouseTo sets the absolute cursor position and feeds the equivalent motion
// delta through the counters. UW integrates the delta from its current cursor,
// so this is exact only from a known origin — call HomeMouse first (the "at:X,Y"
// script token does this automatically).
func (m *Machine) MoveMouseTo(x, y int) {
	ms := m.mouse()
	dx, dy := x-ms.x, y-ms.y
	ms.x, ms.y = x, y
	ms.clamp()
	ms.accX += dx * mickeysPerPixel
	ms.accY += dy * mickeysPerPixel
}

// SetMouseButtons sets the button mask (bit0 left, bit1 right) reported by 03h,
// counting a fresh left/right press for function 05h.
func (m *Machine) SetMouseButtons(mask byte) {
	ms := m.mouse()
	if mask&1 != 0 && ms.buttons&1 == 0 {
		ms.pressL++
	}
	if mask&2 != 0 && ms.buttons&2 == 0 {
		ms.pressR++
	}
	ms.buttons = mask
}
