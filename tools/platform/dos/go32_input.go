package dos

// Scripted keyboard injection for the flat protected-mode (go32) machine — the
// PM counterpart to the real-mode Machine's dos_io.go/dos_keyboard.go path.
//
// Quake blocks in its sound init on "PRESS A KEY." The wait does NOT poll INT 16h
// or the 8042 directly; it drains Quake's OWN keyboard buffer, which its installed
// IRQ1 handler fills. That handler is a *protected-mode* interrupt handler: the
// DJGPP runtime hooked INT 9 through DPMI (INT 31h 0205h), and while waiting the
// game reflects only INT 31h. So to advance the game we must raise a real IRQ1 —
// vector through the game's recorded PM INT 9 handler with a 32-bit interrupt
// frame — and answer the scancode when that handler does `IN AL, 60h`.
//
// The mechanism mirrors the real-mode injector exactly, only in PM:
//
//   - a key press = latch a scancode at port 0x60 and raise IRQ1 through the
//     recorded PM vector (CPU.InterruptPM). The game's own ISR reads 0x60, stores
//     the scancode, EOIs the PIC, and IRETDs; its input code drains the buffer.
//   - delivery waits for IF (InterruptPM honours it), so a key is never dropped
//     while interrupts are masked — it retries every step until interrupts open,
//     exactly as a real IRQ1 would wait for the next STI.
//
// Events are paced against retired instructions (the same clock the PIT uses),
// not a per-step OnStep the PM machine owns — go32run owns c.OnStep, so it calls
// PumpInput from there, the way the real-mode CPU calls Machine.onStep.

import "retroreverse.com/tools/cpu/x86"

// pmVector is a protected-mode interrupt handler entry recorded from DPMI 0205
// (Set PM Interrupt Vector): the selector and 32-bit offset the client installed.
type pmVector struct {
	sel uint16
	off uint32
	set bool
}

// setPMVector records the handler a DPMI 0205h call installs for interrupt bl.
// The load-bearing one is INT 9 (IRQ1): the game's keyboard ISR, which scripted
// input vectors through.
func (p *PM) setPMVector(bl byte, sel uint16, off uint32) {
	if !p.pmVectors[bl].set {
		p.logf("DPMI 0205h: PM INT %02Xh handler -> %04X:%08X", bl, sel, off)
	}
	p.pmVectors[bl] = pmVector{sel: sel, off: off, set: true}
}

// go32InjectPeriod is the number of retired instructions between scripted-input
// events. It only needs to leave the game frames to react between keystrokes; the
// exact value is not physical (like pitInstrsPerTick). One event's `delay` counts
// in these periods.
const go32InjectPeriod = 40000

// SetKeys installs a scripted input schedule (see ParseKeys) to be delivered as
// the PM machine runs. Only injKey/injWait events act here; mouse events (Quake
// needs none in this phase) are skipped.
func (p *PM) SetKeys(events []injEvent) { p.keyEvents = events }

// KeysPending reports whether scripted input events remain.
func (p *PM) KeysPending() bool { return len(p.keyEvents) > 0 }

// PumpInput is driven once per instruction from the run loop's step hook. It
// paces the scripted events against retired instructions and delivers each key as
// an IRQ1 through the game's recorded PM INT 9 handler, retrying every step until
// interrupts are open so a keystroke is never starved by a masked instant.
func (p *PM) PumpInput(c *x86.CPU) {
	if len(p.keyEvents) == 0 {
		return
	}
	// A key that couldn't land (IF masked, or the handler not yet installed)
	// retries every step until it does — InterruptPM no-ops while IF is clear.
	if p.keyRetry {
		if p.deliverKey(c) {
			p.keyRetry = false
			p.popKeyEvent()
		}
		return
	}
	if p.injTick++; p.injTick < go32InjectPeriod {
		return
	}
	p.injTick = 0
	if p.keyWait > 0 {
		p.keyWait--
		return
	}
	if p.keyEvents[0].kind == injKey {
		if !p.deliverKey(c) {
			p.keyRetry = true
			return
		}
	}
	p.popKeyEvent()
}

// deliverKey latches the current event's scancode at the 8042 output buffer and
// raises IRQ1 through the recorded PM INT 9 handler. It returns false (to retry)
// until the game has installed its real keyboard IRQ handler and interrupts are
// open.
//
// Timing matters: at startup INT 9 briefly points at the CRT's in-image Ctrl-C
// filter, which stores nothing to the game's own key ring and only chains onward.
// The game's actual keyboard ISR is a go32-generated wrapper allocated on the heap
// (it reads port 0x60 and fills the ring the input code drains). Holding the key
// until INT 9 points into the heap delivers it to that real handler, exactly when
// the game is ready to receive it — and never fires a stray IRQ into the CRT filter
// during the long init.
func (p *PM) deliverKey(c *x86.CPU) bool {
	v := p.pmVectors[9]
	if !v.set || v.off < p.heapBase {
		return false
	}
	p.kbdData, p.kbdFull = p.keyEvents[0].code, true
	if !c.InterruptPM(v.sel, v.off) {
		return false
	}
	p.keyHits++
	if p.keyHits <= 40 {
		p.logf("INJECT KEY scancode %02X via INT9 %04X:%08X", p.kbdData, v.sel, v.off)
	}
	return true
}

// popKeyEvent removes the head event and arms its trailing pause.
func (p *PM) popKeyEvent() {
	p.keyWait = p.keyEvents[0].delay
	p.keyEvents = p.keyEvents[1:]
}
