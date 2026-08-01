package gbamachine

// The scheduler. A GBA is paced by its display: 228 scanlines of 1232 cycles at
// 16.78 MHz, 160 visible + 68 of vertical blank, ~59.73 frames/s. Each pass of
// the loop is one scanline: the PPU emits it, the CPU gets the line's worth of
// instructions, the horizontal blank lands after the visible dots, and the
// timers tick the line's cycles. Interrupts and DMA hang off those boundaries,
// which is what makes a VBlank arrive at the right rate without simulating a
// clock.
//
// Interrupt delivery goes through a Go model of the BIOS's dispatch shim — the
// same scheme dsmachine uses, and the GBA is where that scheme comes from: the
// BIOS saves {r0-r3, r12, lr}, points r0 at the I/O block, calls through the
// game's pointer at 0x03007FFC, and returns via `subs pc, lr, #4`, restoring
// the interrupted mode and Thumb bit from SPSR. The handler is entered with LR
// aimed at a sentinel PC the scheduler recognises as the BIOS epilogue.

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm"
)

const (
	linesPerFrame = 228
	visibleLines  = 160
	cyclesPerLine = 1232
	// The ARM7TDMI averages a few cycles per instruction once ROM waitstates and
	// Thumb's 16-bit fetches are accounted for; nominal, like all timing here.
	instrsPerLine = cyclesPerLine / 3
)

// biosIRQReturn is a sentinel PC standing in for the BIOS's interrupt epilogue.
// It is never fetched: the scheduler notices a core that has returned to it and
// performs the epilogue in Go.
const biosIRQReturn = 0xFFFF1000

// Result reports how a run ended.
type Result struct {
	Steps     uint64
	Frames    uint64
	Reason    string
	Milestone map[uint32]uint64 // milestone PC -> instruction count when first reached
}

// Run executes until the ADDITIONAL step budget runs out, the CPU halts on
// something unimplemented, a breakpoint fires, or the machine settles into a
// spin that is making no progress. milestones are PCs to report on first visit.
func (m *Machine) Run(budget uint64, milestones map[uint32]string) Result {
	return m.run(budget, milestones, 0)
}

// RunFrames runs until the display completes n more frames (budget as ceiling).
func (m *Machine) RunFrames(n, budget uint64) Result {
	return m.run(budget, nil, m.vid.frames+n)
}

func (m *Machine) run(budget uint64, milestones map[uint32]string, untilFrame uint64) Result {
	res := Result{Milestone: map[uint32]uint64{}}
	if m.visited == nil {
		m.visited = map[uint32]bool{}
	}
	// Progress watchdog state — RELATIVE to the current step count (a machine
	// restored from a savestate is already old; see dsmachine's scar tissue).
	lastProgress := m.Steps
	prevSig := m.progressSig()
	prevPages := len(m.visited)
	m.stop, m.stopped, m.stoppedPC = false, false, 0
	end := m.Steps + budget

	for m.Steps < end {
		// The frame check comes BEFORE startLine, so a run always stops on a
		// COMPLETED scanline. Checking it after startLine (as this did) breaks
		// out of the line that just incremented the frame counter, skipping that
		// line's timers and audio. That is invisible in a screenshot — the
		// display for that line had already been composed — and a real loss
		// across a savestate: a resumed run dropped one scanline of timer ticks,
		// which desynchronised the Direct Sound FIFO against the straight run for
		// 27 output frames. Only comparing the AUDIO across a save/resume
		// boundary could find it.
		if untilFrame != 0 && m.vid.frames >= untilFrame {
			res.Reason = fmt.Sprintf("reached frame %d", m.vid.frames)
			break
		}
		m.startLine()
		if m.stop {
			res.Reason = "stopped"
			break
		}

		// The line's CPU time, with the horizontal blank landing where it does on
		// hardware: after 240 of the 308 dot-times.
		const quantum = 32
		hb := false
		for spent := 0; spent < instrsPerLine; spent += quantum {
			if !hb && spent*308/instrsPerLine >= 240 {
				m.hblankNow()
				hb = true
			}
			m.deliver()
			m.runQuantum(quantum, milestones, res.Milestone)
			m.Steps += quantum
			if m.cpu.Halted || m.stop {
				break
			}
		}
		if !hb {
			m.hblankNow()
		}
		m.tickTimers(cyclesPerLine)
		m.apu.mixCycles(cyclesPerLine)

		if m.stop {
			res.Reason = "stopped"
			if m.stopped {
				res.Reason = fmt.Sprintf("breakpoint at 0x%08X", m.stoppedPC)
			}
			break
		}
		if m.cpu.Halted {
			res.Reason = "halted: " + m.cpu.HaltReason
			break
		}

		sig := m.progressSig()
		if sig != prevSig || len(m.visited) != prevPages {
			prevSig, prevPages = sig, len(m.visited)
			lastProgress = m.Steps
		} else if m.Steps-lastProgress > 24_000_000 {
			res.Reason = fmt.Sprintf("settled — CPU at 0x%08X (%s); no new code or interrupt traffic",
				m.cpu.R[15], m.Parked())
			break
		}
	}
	if res.Reason == "" {
		res.Reason = fmt.Sprintf("step budget (%d) reached", budget)
	}
	res.Steps, res.Frames = m.Steps, m.vid.frames
	return res
}

// startLine advances the raster one scanline and runs everything that hangs off
// the line boundary: the PPU render, the V-blank and V-count events, their DMA.
func (m *Machine) startLine() {
	m.vid.hblank = false
	m.vid.line++
	if m.vid.line >= linesPerFrame {
		m.vid.line = 0
	}
	line := m.vid.line
	dispstat := m.io[0x004]

	switch {
	case line == 0:
		m.ppu.startFrame(m)
	case line == visibleLines: // entering the vertical blank
		m.vid.frames++
		if dispstat&(1<<3) != 0 {
			m.raise(irqVBlank)
		}
		m.dmaTrigger(1)
		if m.OnFrame != nil {
			m.OnFrame()
		}
	}
	if line < visibleLines {
		m.ppu.renderLine(m, line)
		m.ppu.stepAffine(m)
	}
	if line == int(dispstat>>8) && dispstat&(1<<5) != 0 {
		m.raise(irqVCount)
	}
}

// hblankNow marks the horizontal blank and runs its events (visible lines only,
// as on hardware).
func (m *Machine) hblankNow() {
	m.vid.hblank = true
	if m.vid.line < visibleLines {
		if m.io[0x004]&(1<<4) != 0 {
			m.raise(irqHBlank)
		}
		m.dmaTrigger(2)
	}
}

// raise latches interrupt sources into IF.
func (m *Machine) raise(mask uint16) { m.if_ |= mask }

// deliver wakes a parked CPU and dispatches a pending, unmasked interrupt
// through the BIOS shim.
func (m *Machine) deliver() {
	pending := m.ie & m.if_
	if pending == 0 {
		return
	}
	if m.waiting {
		// Halt/IntrWait wake on any enabled+pending source; IntrWait's mask is
		// checked against the CHECK FLAGS after the handler runs, not against IF.
		m.waiting = false
		m.waitAny = false
	} else if m.cpu.IRQDisable {
		return // a critical section; let it finish
	}
	if !m.ime {
		return // woken but not vectored (Halt with IME off is legal GBA idiom)
	}

	b := &bus{m: m}
	handler := b.r32(irqHandlerSlot)
	if handler == 0 {
		return // nothing installed: leave the flags latched
	}
	ret := m.cpu.R[15]
	if m.OnIRQ != nil {
		m.OnIRQ(pending, handler, ret)
	}

	// The hardware's part: bank into IRQ mode, SPSR = CPSR, LR = ret + 4.
	m.cpu.Exception(arm.ModeIRQ, handler, ret+4)

	// The BIOS's part: push {r0-r3, r12, lr}, point r0 at the I/O block, and hand
	// the handler an LR that lands back in the epilogue.
	sp := m.cpu.R[13] - 24
	m.cpu.R[13] = sp
	b.w32(sp+0, m.cpu.R[0])
	b.w32(sp+4, m.cpu.R[1])
	b.w32(sp+8, m.cpu.R[2])
	b.w32(sp+12, m.cpu.R[3])
	b.w32(sp+16, m.cpu.R[12])
	b.w32(sp+20, m.cpu.R[14])
	m.cpu.R[0] = ioBase
	m.cpu.R[14] = biosIRQReturn
}

// biosIRQExit is the BIOS epilogue: restore what the shim pushed, return via
// SPSR (mode, IRQ mask and Thumb bit), then settle any IntrWait the CPU was in.
func (m *Machine) biosIRQExit() {
	b := &bus{m: m}
	sp := m.cpu.R[13]
	m.cpu.R[0] = b.r32(sp + 0)
	m.cpu.R[1] = b.r32(sp + 4)
	m.cpu.R[2] = b.r32(sp + 8)
	m.cpu.R[3] = b.r32(sp + 12)
	m.cpu.R[12] = b.r32(sp + 16)
	lr := b.r32(sp + 20)
	m.cpu.R[13] = sp + 24

	spsr := m.cpu.SPSR()
	m.cpu.SetCPSR(spsr)
	m.cpu.R[15] = lr - 4

	if m.waitMask != 0 {
		// The BIOS's IntrWait loop: the game's handler reports serviced sources in
		// the check flags; wanted ones end the wait (and are consumed), anything
		// else re-parks.
		flags := uint16(b.r32(irqCheckFlags))
		if flags&m.waitMask != 0 {
			b.w32(irqCheckFlags, uint32(flags&^m.waitMask))
			m.waitMask = 0
		} else {
			m.waiting = true
		}
	}
}

// runQuantum steps the CPU up to n instructions, unless it is parked.
func (m *Machine) runQuantum(n int, milestones map[uint32]string, hit map[uint32]uint64) {
	if m.waiting {
		return
	}
	for i := 0; i < n; i++ {
		if m.waiting || m.cpu.Halted || m.stop {
			return
		}
		if m.cpu.R[15] == biosIRQReturn {
			m.biosIRQExit()
			continue
		}
		pc := m.cpu.R[15]
		if m.OnStep != nil {
			m.OnStep(pc)
		}
		if m.bps[pc] {
			m.stop, m.stopped, m.stoppedPC = true, true, pc
			return
		}
		m.visited[pc>>8] = true
		if milestones != nil {
			if _, ok := milestones[pc]; ok {
				if _, seen := hit[pc]; !seen {
					hit[pc] = m.Steps + uint64(i)
				}
			}
		}
		m.cpu.Step()
	}
}

// progressSig fingerprints interrupt-level progress. Deliberately omits the PC
// (a spin oscillates it) and the frame count (the display never stops, so it
// would report progress for ever — dsmachine's lesson).
func (m *Machine) progressSig() uint64 {
	return uint64(m.if_) ^ uint64(m.ie)<<16 ^ uint64(m.keys)<<32 ^
		uint64(len(m.eeprom.inBits))<<48
}
