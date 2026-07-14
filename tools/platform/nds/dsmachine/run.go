package dsmachine

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
)

// The scheduler. A DS is paced by its display, so this loop is written around the
// scanline rather than around an instruction count: each pass emits one line, and
// the two CPUs are given the instruction budget that line's worth of clock buys
// them (video.go). Interrupts, DMA and the timers all hang off that cadence, which
// is what makes a VBlank arrive at the right rate without anything having to
// simulate a clock.
//
// The two cores are interleaved in small quanta rather than run a line at a time
// each, because they talk: the IPC handshake is a ping-pong of single words, and a
// core that runs four thousand instructions before its partner gets a turn spends
// nearly all of them waiting for a reply that is already sitting in the FIFO.

// Result reports how a run ended.
type Result struct {
	Steps      uint64
	Frames     uint64
	Reason     string
	ARM9Milest map[uint32]uint64 // milestone PC -> instruction count when first reached
}

// Run co-executes both cores until the step budget runs out, a core halts, or the
// machine settles into a mutual spin. milestones are ARM9 PCs to report the first
// time each is reached; quantum is how many ARM9 instructions run before the ARM7
// gets a turn.
func (m *Machine) Run(budget uint64, quantum int, milestones map[uint32]string) Result {
	return m.run(budget, quantum, milestones, 0)
}

// RunFrames runs until the display has completed n more frames, with budget as a
// ceiling. A graphics workload is measured in frames; an instruction budget only
// contains one by guesswork, and the guess moves whenever anything about the
// machine's idling changes.
func (m *Machine) RunFrames(n, budget uint64, quantum int) Result {
	return m.run(budget, quantum, nil, m.vid.frames+n)
}

func (m *Machine) run(budget uint64, quantum int, milestones map[uint32]string, untilFrame uint64) Result {
	res := Result{ARM9Milest: map[uint32]uint64{}}
	if quantum <= 0 {
		quantum = 64
	}
	if m.visited == nil {
		m.visited = map[uint32]bool{}
	}

	// Forward-progress watchdog: track how many distinct 256-byte ARM9 code pages
	// have been entered. When that stops growing, and no IPC traffic moves either,
	// both cores are in a mutual spin and the run has settled.
	//
	// lastProgress starts at the CURRENT step count, not at zero. Starting it at zero
	// looks harmless in a single run from reset and is fatal to every run after it: on
	// a machine that has already executed, say, 80 million steps, the very first line
	// of the next call sees `Steps - lastProgress` far past the threshold and declares
	// the machine settled before it has run an instruction. Every RunFrames after the
	// first then returns instantly, and the caller measures a machine that never moved.
	lastProgress := m.Steps
	prevSig := m.progressSig()
	prevPages := len(m.visited)

	for m.Steps < budget {
		m.startLine()
		if untilFrame != 0 && m.vid.frames >= untilFrame {
			res.Reason = fmt.Sprintf("reached frame %d", m.vid.frames)
			break
		}

		// One scanline of CPU time, interleaved, with the horizontal blank landing
		// where it does on hardware: once the visible dots have been emitted.
		hb := false
		for spent := 0; spent < cyclesPerLine9; spent += quantum {
			if !hb && spent*dotsPerLine/cyclesPerLine9 >= visibleDots {
				m.hblankNow()
				hb = true
			}
			m.deliver(m.ARM9)
			m.deliver(m.ARM7)
			m.runQuantum(m.ARM9, quantum, milestones, res.ARM9Milest)
			m.runQuantum(m.ARM7, quantum/2, nil, nil) // the ARM7 clocks at half the ARM9
			m.Steps += uint64(quantum)
			if m.ARM9.cpu.Halted || m.ARM7.cpu.Halted {
				break
			}
		}
		if !hb {
			m.hblankNow()
		}
		// The timers clock off the system clock, not off either CPU — a timer measures
		// the machine's time, not how fast we happen to interpret instructions.
		for _, c := range m.cores() {
			c.tickTimers(cyclesPerLine7)
		}

		if m.ARM9.cpu.Halted {
			res.Reason = "ARM9 halted: " + m.ARM9.cpu.HaltReason
			break
		}
		if m.ARM7.cpu.Halted {
			res.Reason = "ARM7 halted: " + m.ARM7.cpu.HaltReason
			break
		}

		sig := m.progressSig()
		if sig != prevSig || len(m.visited) != prevPages {
			prevSig, prevPages = sig, len(m.visited)
			lastProgress = m.Steps
		} else if m.Steps-lastProgress > 24_000_000 {
			res.Reason = fmt.Sprintf("settled — ARM9 spinning at 0x%08X (%s), ARM7 at 0x%08X (%s); no new code or IPC traffic",
				m.ARM9.cpu.R[15], parkState(m.ARM9), m.ARM7.cpu.R[15], parkState(m.ARM7))
			break
		}
	}
	if res.Reason == "" {
		res.Reason = fmt.Sprintf("step budget (%d) reached", budget)
	}
	res.Steps, res.Frames = m.Steps, m.vid.frames
	return res
}

// runQuantum steps one core up to n instructions, unless it is parked in an
// interrupt wait (then it does nothing until the scheduler delivers an IRQ).
func (m *Machine) runQuantum(c *core, n int, milestones map[uint32]string, hit map[uint32]uint64) {
	if c.waiting {
		return
	}
	if c.sleep > 0 { // in a WaitByLoop delay: let wall-time pass for the other core
		c.sleep -= n
		return
	}
	for i := 0; i < n; i++ {
		if c.waiting || c.sleep > 0 || c.cpu.Halted {
			return
		}
		if c.cpu.R[15] == biosIRQReturn {
			c.biosIRQExit() // the handler has returned into the BIOS epilogue
			continue
		}
		pc := c.cpu.R[15]
		if m.OnStep != nil {
			m.OnStep(c.arm9, pc)
		}
		if c.arm9 {
			m.visited[pc>>8] = true
			if milestones != nil {
				if _, ok := milestones[pc]; ok {
					if _, seen := hit[pc]; !seen {
						hit[pc] = m.Steps + uint64(i)
					}
				}
			}
		}
		c.cpu.Step()
	}
}

// biosIRQReturn is a sentinel PC standing in for the DS BIOS's interrupt epilogue.
// It is not a real address and is never fetched: the scheduler notices a core whose
// PC has reached it and performs the epilogue in Go (see biosIRQExit).
const biosIRQReturn = 0xFFFF1000

// deliver dispatches a pending, unmasked interrupt to a core, THE WAY THE BIOS DOES.
//
// The game's interrupt handler is not the interrupt vector. It is a routine the BIOS
// *calls*, and the difference is not a detail — it decides whether the machine comes
// back from an interrupt at all. The DS BIOS's IRQ path is:
//
//	stmfd sp!, {r0-r3, r12, lr}   ; save the caller's scratch registers and the return
//	add   lr, pc, #0              ; lr = the epilogue below
//	ldr   pc, [r0, #-4]           ; jump to the user handler at the fixed slot
//	ldmfd sp!, {r0-r3, r12, lr}   ; ...and it comes back HERE
//	subs  pc, lr, #4              ; return, restoring CPSR from SPSR
//
// So the user handler is entered with LR pointing at the BIOS's own epilogue, and it
// returns there — SM64DS's ARM7 handler does it with `ldr pc, [sp], #4`, a plain load
// that restores nothing. Only that final `subs pc, lr, #4` restores the CPSR, and
// with it the mode and the THUMB BIT.
//
// Jump straight to the user handler instead, as the obvious model does, and the
// handler returns to an address with the T bit still clear. Interrupt an ARM7 that
// was executing Thumb — which it is, most of the time, because the BIOS call thunks
// are Thumb — and it resumes in ARM state on a Thumb address, decodes the thunk table
// as ARM, and runs off into nonsense. The boot then hangs somewhere with no visible
// relationship to interrupts at all.
func (m *Machine) deliver(c *core) {
	pending := c.ie & c.if_
	if pending == 0 || !c.ime {
		return
	}
	if c.waiting {
		// Wake only for the sources actually waited on (Halt waits for any).
		if !c.waitAny && c.waitMask != 0 && pending&c.waitMask == 0 {
			return
		}
	} else if c.cpu.IRQDisable {
		return // running with interrupts masked: let the critical section finish
	}
	b := &bus{c: c}
	// The OS IRQ check flag: the BIOS ORs the acknowledged sources into it, and
	// IntrWait polls it. It lives two words below the handler pointer.
	flag := b.r32(c.handlerBase - 8)
	b.w32(c.handlerBase-8, flag|pending)
	handler := b.r32(c.handlerBase - 4)
	if handler == 0 {
		return // no handler installed yet: latch the flag and leave it pending
	}
	// The test is `== 0`, and not "does this look like a sensible code address?".
	// The sensible-looking version — reject anything below main RAM — silently threw
	// every ARM9 interrupt away: SM64DS puts its handler in ITCM, at 0x01FFD97C, which
	// is BELOW main RAM and is exactly where a DS game wants its interrupt handler,
	// because ITCM is the only memory the ARM9 reaches in a single cycle. The machine
	// then latched interrupt flags for ever and dispatched none of them, and the boot
	// deadlocked with the ARM7's reply sitting unread in the FIFO.

	// R15 is already the instruction that would have run next — for a running core
	// because we deliver before stepping it, and for a parked one because the SWI that
	// parked it completed and advanced the PC before we stopped stepping.
	ret := c.cpu.R[15]
	c.waiting = false
	if m.OnIRQ != nil {
		m.OnIRQ(c.arm9, pending, handler, ret)
	}

	// The hardware's part: bank in IRQ mode, save CPSR into SPSR, LR = ret + 4.
	c.cpu.Exception(arm.ModeIRQ, handler, ret+4)

	// The BIOS's part: push the scratch registers and the real return, and hand the
	// user handler an LR that lands back in the epilogue.
	sp := c.cpu.R[13] - 24
	c.cpu.R[13] = sp
	b.w32(sp+0, c.cpu.R[0])
	b.w32(sp+4, c.cpu.R[1])
	b.w32(sp+8, c.cpu.R[2])
	b.w32(sp+12, c.cpu.R[3])
	b.w32(sp+16, c.cpu.R[12])
	b.w32(sp+20, c.cpu.R[14]) // the hardware LR: ret + 4
	c.cpu.R[14] = biosIRQReturn
}

// biosIRQExit is the BIOS's interrupt epilogue: restore what it saved, then return
// with `subs pc, lr, #4`, which puts SPSR back into CPSR — restoring the mode, the
// interrupt mask, and the Thumb bit the interrupted code was running under.
func (c *core) biosIRQExit() {
	b := &bus{c: c}
	sp := c.cpu.R[13]
	c.cpu.R[0] = b.r32(sp + 0)
	c.cpu.R[1] = b.r32(sp + 4)
	c.cpu.R[2] = b.r32(sp + 8)
	c.cpu.R[3] = b.r32(sp + 12)
	c.cpu.R[12] = b.r32(sp + 16)
	lr := b.r32(sp + 20)
	c.cpu.R[13] = sp + 24

	spsr := c.cpu.SPSR() // read it while still banked into IRQ mode
	c.cpu.SetCPSR(spsr)
	c.cpu.R[15] = lr - 4
}

// progressSig fingerprints IPC-level progress: the sync nibbles, the FIFO depths and
// the pending-interrupt flags. It deliberately omits the PCs (a tight spin oscillates
// them without making progress); code progress is tracked separately, by the set of
// ARM9 pages entered. When neither moves for a window, the run has settled.
//
// It also, deliberately, omits the frame count. Putting it in seemed reasonable — a
// machine that is still drawing frames is surely alive — but it is exactly wrong: the
// display keeps running whatever the CPUs do, so a frame counter in this signature
// reports "progress" for ever and the watchdog can never fire. A pair of cores
// deadlocked against each other would then simply run to the step budget and report
// that it had run out of steps, which is true and useless.
func (m *Machine) progressSig() uint64 {
	return uint64(m.ipc.sync9)<<1 ^ uint64(m.ipc.sync7)<<5 ^
		uint64(len(m.ipc.to7))<<8 ^ uint64(len(m.ipc.to9))<<12 ^
		uint64(m.ARM9.if_)<<16 ^ uint64(m.ARM7.if_)<<32
}

func parkState(c *core) string {
	if c.waiting {
		if c.waitAny {
			return "halted for IRQ"
		}
		return fmt.Sprintf("IntrWait 0x%X", c.waitMask)
	}
	return "running"
}
