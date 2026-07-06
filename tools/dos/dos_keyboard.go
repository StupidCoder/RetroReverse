package dos

// Scripted input injection — keyboard and mouse. Ultima Underworld installs its
// own INT 9 (IRQ1) handler at 214A:062E: it does `IN AL,60h`, stores the raw
// scancode into a 64-byte ring buffer, acknowledges the keyboard and sends the
// PIC EOI, then IRETs; its input code drains that ring. Its pointer-driven UI
// polls the INT 33h mouse (see dos_mouse.go). So driving the game from a script
// needs nothing exotic:
//
//   - a key press = put a scancode in the 8042 output latch and raise IRQ1, and
//     the game's own ISR consumes it as a real key press (make code, then break
//     code = make|80h). IRQ1 delivery waits for interrupts to be enabled,
//     mirroring real hardware.
//   - a mouse move/click = update the driver's position/button state, which the
//     game sees on its next INT 33h poll.
//
// Events are paced against the injected timer tick (see onStep) so the game gets
// frames to react between them, mirroring a human at the controls.

import (
	"fmt"
	"strconv"
	"strings"
)

// injKind distinguishes the scheduled input actions.
type injKind byte

const (
	injWait   injKind = iota // pure pause
	injKey                   // deliver scancode via IRQ1
	injHome                  // slam the cursor to (0,0) via a large negative delta
	injMove                  // move the mouse to (x, y)
	injButton                // set the mouse button mask
	injDelta                 // feed a raw mickey delta (dx, dy) into the motion counters
)

// injEvent is one scheduled input action, followed by a `delay`-tick pause.
type injEvent struct {
	kind  injKind
	code  byte // scancode (injKey) or button mask (injButton)
	x, y  int  // target position (injMove)
	delay int
}

// makeScancodes maps a key token to its US-layout Set-1 (XT) make scancode. The
// break code is the make code | 0x80. Arrow/navigation keys use their numeric-
// keypad scancodes, which UW's raw-scancode ISR reads directly.
var makeScancodes = map[string]byte{
	"esc": 0x01, "1": 0x02, "2": 0x03, "3": 0x04, "4": 0x05, "5": 0x06,
	"6": 0x07, "7": 0x08, "8": 0x09, "9": 0x0A, "0": 0x0B, "-": 0x0C, "=": 0x0D,
	"bs": 0x0E, "backspace": 0x0E, "tab": 0x0F,
	"q": 0x10, "w": 0x11, "e": 0x12, "r": 0x13, "t": 0x14, "y": 0x15, "u": 0x16,
	"i": 0x17, "o": 0x18, "p": 0x19, "enter": 0x1C, "return": 0x1C,
	"a": 0x1E, "s": 0x1F, "d": 0x20, "f": 0x21, "g": 0x22, "h": 0x23, "j": 0x24,
	"k": 0x25, "l": 0x26, ";": 0x27, "z": 0x2C, "x": 0x2D, "c": 0x2E, "v": 0x2F,
	"b": 0x30, "n": 0x31, "m": 0x32, ",": 0x33, ".": 0x34, "/": 0x35,
	"space": 0x39, "sp": 0x39,
	"lshift": 0x2A, "shift": 0x2A, "rshift": 0x36, "ctrl": 0x1D, "alt": 0x38,
	"up": 0x48, "left": 0x4B, "right": 0x4D, "down": 0x50,
	"home": 0x47, "end": 0x4F, "pgup": 0x49, "pgdn": 0x51, "ins": 0x52, "del": 0x53,
	"f1": 0x3B, "f2": 0x3C, "f3": 0x3D, "f4": 0x3E, "f5": 0x3F, "f6": 0x40,
	"f7": 0x41, "f8": 0x42, "f9": 0x43, "f10": 0x44,
}

// ParseKeys compiles a comma-separated input script into an injection schedule.
// Tokens:
//
//	<key>          a key name/character — "down", "enter", "a" (make + break)
//	kdown:<key>    press a key WITHOUT releasing it (send only the make code) —
//	               the game sees the key held down until a matching kup
//	kup:<key>      release a held key (send the break code)
//	wait:N         pause N timer-ticks
//	at:X,Y         move the mouse to screen pixel (X, Y)
//	lclick/rclick  press then release the left/right mouse button
//	ldown/lup      hold/release the left button (rdown/rup for the right)
//
// A held key (kdown/…/kup) is how continuous actions like walking are expressed:
// UW tracks make/break from its scancode ring, so a key stays "down" between a
// kdown and its kup. Shift is a real key too, so "kdown:shift,kdown:w,wait:20,
// kup:w,kup:shift" walks with shift held. Example:
// "at:120,140,lclick,wait:60,d,a,e,enter" clicks a button then types.
func ParseKeys(spec string) ([]injEvent, error) {
	const (
		makeBreakGap = 2  // ticks between a key's make and break codes
		interKeyGap  = 10 // ticks after a key/click, before the next event
		clickGap     = 4  // ticks a mouse button is held down before release —
		// UW detects a click as a short press→release pulse polled within one UI
		// frame; a long hold reads as a hold/drag gesture and does not activate.
	)
	toks := splitTokens(spec)
	var out []injEvent
	for i := 0; i < len(toks); i++ {
		tok := strings.ToLower(toks[i])
		switch {
		case tok == "":
			continue
		case strings.HasPrefix(tok, "wait:"):
			n, err := strconv.Atoi(tok[len("wait:"):])
			if err != nil {
				return nil, fmt.Errorf("keys: bad wait %q", tok)
			}
			out = append(out, injEvent{kind: injWait, delay: n})
		case strings.HasPrefix(tok, "at:"):
			// "at:X" consumes the following token as Y (the comma split them).
			if i+1 >= len(toks) {
				return nil, fmt.Errorf("keys: %q needs X,Y", tok)
			}
			x, err1 := strconv.Atoi(tok[len("at:"):])
			y, err2 := strconv.Atoi(strings.TrimSpace(toks[i+1]))
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("keys: bad coords %q,%q", tok, toks[i+1])
			}
			i++
			// Home the cursor into the top-left corner first (a big negative delta
			// clamps UW's own cursor at 0,0), then move by a known delta — this
			// makes absolute targeting exact despite UW tracking only relative
			// motion. A short wait between lets the game poll the home delta.
			out = append(out,
				injEvent{kind: injHome, delay: 3},
				injEvent{kind: injMove, x: x, y: y, delay: makeBreakGap})
		case strings.HasPrefix(tok, "dxy:"):
			// "dxy:DX" consumes the following token as DY (the comma split them).
			if i+1 >= len(toks) {
				return nil, fmt.Errorf("keys: %q needs DX,DY", tok)
			}
			dx, err1 := strconv.Atoi(tok[len("dxy:"):])
			dy, err2 := strconv.Atoi(strings.TrimSpace(toks[i+1]))
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("keys: bad delta %q,%q", tok, toks[i+1])
			}
			i++
			out = append(out, injEvent{kind: injDelta, x: dx, y: dy, delay: makeBreakGap})
		case tok == "lclick", tok == "rclick":
			mask := byte(1)
			if tok == "rclick" {
				mask = 2
			}
			out = append(out,
				injEvent{kind: injButton, code: mask, delay: clickGap},
				injEvent{kind: injButton, code: 0, delay: interKeyGap})
		case tok == "ldown":
			out = append(out, injEvent{kind: injButton, code: 1, delay: makeBreakGap})
		case tok == "rdown":
			out = append(out, injEvent{kind: injButton, code: 2, delay: makeBreakGap})
		case tok == "lup", tok == "rup":
			out = append(out, injEvent{kind: injButton, code: 0, delay: interKeyGap})
		case strings.HasPrefix(tok, "kdown:"), strings.HasPrefix(tok, "kup:"):
			name := tok[strings.IndexByte(tok, ':')+1:]
			sc, ok := makeScancodes[name]
			if !ok {
				return nil, fmt.Errorf("keys: unknown key %q in %q", name, tok)
			}
			if strings.HasPrefix(tok, "kup:") {
				sc |= 0x80 // break code
			}
			out = append(out, injEvent{kind: injKey, code: sc, delay: makeBreakGap})
		default:
			sc, ok := makeScancodes[tok]
			if !ok {
				return nil, fmt.Errorf("keys: unknown token %q", tok)
			}
			out = append(out,
				injEvent{kind: injKey, code: sc, delay: makeBreakGap},
				injEvent{kind: injKey, code: sc | 0x80, delay: interKeyGap})
		}
	}
	return out, nil
}

// splitTokens splits on commas and trims — kept separate so "at:X,Y" can be
// rejoined by the parser (the comma inside a coordinate pair is significant).
func splitTokens(spec string) []string {
	parts := strings.Split(spec, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// SetKeys installs an input-injection schedule to be delivered as the machine runs.
func (m *Machine) SetKeys(events []injEvent) { m.keyEvents = events }

// KeysPending reports whether unscheduled input events remain.
func (m *Machine) KeysPending() bool { return len(m.keyEvents) > 0 }

// pumpKeys is called from onStep once per injected timer-tick, on a phase offset
// from the timer IRQ. It processes the next scheduled event; a keystroke is only
// delivered when the game will accept an interrupt (IF set, not shadowed), so a
// key is never dropped while interrupts are masked. Mouse events touch driver
// state the game polls, so they always apply immediately.
func (m *Machine) pumpKeys() {
	if len(m.keyEvents) == 0 {
		return
	}
	if m.keyWait > 0 {
		m.keyWait--
		return
	}
	ev := m.keyEvents[0]
	switch ev.kind {
	case injKey:
		m.io.kbdOut, m.io.kbdOutFull = ev.code, true
		if !m.CPU.Interrupt(9) {
			m.io.kbdOutFull = false // interrupts masked now — keep retrying every
			m.keyRetry = true       // step (not just next tick) until IF is set, so
			return                  // key delivery can't be starved by a bad phase
		}
		m.keyRetry = false
		m.logInject("KEY scancode %02X (IVT9=%04X:%04X)", ev.code, m.r16(9*4+2), m.r16(9*4))
	case injHome:
		m.HomeMouse()
		m.logInject("MOUSE home -> 0,0")
	case injMove:
		m.MoveMouseTo(ev.x, ev.y)
		m.logInject("MOUSE move -> %d,%d", ev.x, ev.y)
	case injButton:
		m.SetMouseButtons(ev.code)
		m.logInject("MOUSE buttons=%02X at %d,%d", ev.code, m.ms.x, m.ms.y)
	case injDelta:
		m.FeedMickeys(ev.x, ev.y)
		m.logInject("MOUSE mickeys += %d,%d", ev.x, ev.y)
	}
	m.keyEvents = m.keyEvents[1:]
	m.keyWait = ev.delay
}

func (m *Machine) logInject(format string, args ...interface{}) {
	m.keyHits++
	if m.keyHits <= 60 {
		m.logf("INJECT: "+format, args...)
	}
}
