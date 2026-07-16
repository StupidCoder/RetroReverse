package ps2

// iopvblank.go — the vblank library: the IOP kernel's per-frame heartbeat.
//
// The vblank library lives in the IOP's ROM alongside intrman and sysmem, so it is
// ours to model. Its identity was earned the way dmacman's was — from the caller's
// own code, not a header:
//
//   - The census showed `vblank#8` called three times at module start
//     (`vblank#8(0x0, 0x20, 0xA35B0, 0x0) from start+0x74`), and OVERLORD's shipped
//     symbols name 0xA35B0 `VBlank_Handler`: a routine that reads a module global,
//     bails if it is zero, does its per-frame work, and returns 1. The signature is
//     therefore (edge, priority, handler, arg): a *registration*, #8 =
//     RegisterVblankHandler, #9 its inverse.
//
//   - padman's poll threads (padman+0x864, +0x8A8, +0x23B4) sit WAITing on event-flag
//     bits (0x1001/0x1004) that nothing else on the machine ever sets, while the EE
//     calls scePadGetState 766 times per three frames and reads a status byte parked
//     at 5, one short of the 6 it needs. The thing that sets those bits — the thing
//     that turns "a pad exists on the SIO2" into "the EE sees buttons" — is a routine
//     run once per vertical blank in interrupt context. Without it padman never
//     issues a single port-0 transfer, the pad never goes stable, and the title's
//     memory-card dialog waits forever for an OK nobody can press.
//
// Delivery: once per displayed frame the machine's frame clock (deliverVBlank in
// intr.go) pokes vblankTick, and the next serviceable moment runs every registered
// handler on the interrupt stack, masked, exactly as intrDeliver runs an intrman
// handler — including the reschedule question afterwards, because a vblank handler's
// whole purpose is to make a sleeping thread ready. On the board the start and end of
// the blank are two interrupt lines (0 and 11); this model runs both edges' handlers
// at the one tick, start-edge first, which keeps the per-frame count honest even
// though the sub-frame phase is fiction. No caller so far registers an end-edge
// handler; the registration log will name the first one that does.
//
// The handler's return value is left alone. The one handler read so far
// (VBlank_Handler) returns 1 on its worked path and its early-out path alike, so no
// removal-on-zero contract can be derived from it; handlers leave the list only
// through vblank#9.

// iopVblankHandler is one registration made through vblank#8.
type iopVblankHandler struct {
	edge uint32 // 0 = vblank start; anything else = the other edge
	prio uint32
	fn   uint32
	arg  uint32
}

// iopVblankLine is the interrupt line the blank arrives on, used here only to file
// the raised/delivered counts where the census reports them.
const iopVblankLine = 0

// vblankRegister is vblank#8, RegisterVblankHandler(edge, priority, handler, arg).
func (p *IOP) vblankRegister() {
	edge, prio, fn, arg := p.arg(0), p.arg(1), p.arg(2), p.arg(3)

	// Re-registering the same routine on the same edge replaces it.
	for i, h := range p.vblankHandlers {
		if h.fn == fn && h.edge == edge {
			p.vblankHandlers[i] = iopVblankHandler{edge: edge, prio: prio, fn: fn, arg: arg}
			p.setRet(0)
			return
		}
	}
	p.vblankHandlers = append(p.vblankHandlers, iopVblankHandler{edge: edge, prio: prio, fn: fn, arg: arg})

	// Start-edge handlers first, then by ascending priority number, insertion order
	// breaking ties. The list is tiny and registration is rare; a stable insertion
	// sort keeps the order readable in a debugger.
	for i := len(p.vblankHandlers) - 1; i > 0; i-- {
		a, b := p.vblankHandlers[i-1], p.vblankHandlers[i]
		if (a.edge != 0 && b.edge == 0) || (a.edge == b.edge && a.prio > b.prio) {
			p.vblankHandlers[i-1], p.vblankHandlers[i] = b, a
		}
	}

	p.ps2.note("IOP: vblank handler %s registered (edge %d, priority 0x%X, arg 0x%X)",
		p.Sym(fn), edge, prio, arg)
	p.setRet(0)
}

// vblankRelease is vblank#9, ReleaseVblankHandler(edge, handler).
func (p *IOP) vblankRelease() {
	edge, fn := p.arg(0), p.arg(1)
	for i, h := range p.vblankHandlers {
		if h.fn == fn && h.edge == edge {
			p.vblankHandlers = append(p.vblankHandlers[:i], p.vblankHandlers[i+1:]...)
			p.ps2.note("IOP: vblank handler %s released (edge %d)", p.Sym(fn), edge)
			p.setRet(0)
			return
		}
	}
	p.setRet(0)
}

// vblankTick marks that a vertical blank has begun. The machine's frame clock calls
// it once per frame; the handlers run at the next moment an interrupt could be taken
// (serviceIntr), which is the same deferral every real line gets.
func (p *IOP) vblankTick() {
	if len(p.vblankHandlers) == 0 {
		return
	}
	p.vblankPending = true
	p.raised[iopVblankLine]++
}

// vblankDeliver runs every registered vblank handler in interrupt context. It is
// intrDeliver for a line whose dispatcher is ours: same frame on the interrupted
// thread's stack, same masked execution on the interrupt stack, same reschedule
// question on the way out — a vblank handler exists to wake a thread, and a delivery
// that never asked would leave that thread ready and unrun.
func (p *IOP) vblankDeliver() {
	frame := (p.CPU.Reg(29) - iopFrameSize) &^ 7
	p.saveFrame(frame)

	// Same rule as intrDeliver: while a module's entry point runs on the machine's
	// own stack there is no thread context to switch away from, so the handlers run
	// but the switch is declined.
	preemptible := p.callDepth == 0

	p.inIntr++
	p.intrEnabled = false
	p.ieEvent("deliver", frame, iopVblankLine, p.CPU.Reg(31))
	p.delivered[iopVblankLine]++

	var failed uint32
	var failErr error
	for _, h := range p.vblankHandlers {
		p.CPU.SetReg(4, h.arg)
		if _, err := p.callGuestOn(h.fn, iopIntrStack); err != nil {
			failed, failErr = h.fn, err
			break
		}
	}

	resume := frame
	if failErr == nil && preemptible {
		if next, ok := p.intrReschedule(frame); ok {
			resume = next
		}
	}

	p.inIntr--
	p.loadFrame(resume)

	if failErr != nil {
		p.halt("the vblank handler %s did not return: %v", p.Sym(failed), failErr)
	}
}
