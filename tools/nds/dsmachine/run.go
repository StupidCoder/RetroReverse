package dsmachine

import (
	"fmt"

	"retroreverse.com/tools/arm"
)

// Result reports how a run ended.
type Result struct {
	Steps      uint64
	Reason     string
	ARM9Milest map[uint32]uint64 // milestone PC -> step count when first reached
}

// vblankPeriod is the synthetic VBlank cadence (in scheduler rounds). A round runs
// both cores one quantum, so this is a coarse frame tick that keeps VBlank waits
// progressing without modelling the video timing.
const vblankPeriod = 20000

// Run co-executes both cores until one stops, the ARM9 reaches a milestone-and-idle
// condition, or the step budget runs out. milestones are ARM9 PCs to report the
// first time each is reached. quantum is how many instructions each core runs per
// round before switching (a small number keeps the IPC handshake tight).
func (m *Machine) Run(budget uint64, quantum int, milestones map[uint32]string) Result {
	res := Result{ARM9Milest: map[uint32]uint64{}}
	m.vblankAt = vblankPeriod

	// Forward-progress watchdog: track how many distinct 256-byte ARM9 code pages
	// have been entered. When that stops growing (and no FIFO/sync activity moves)
	// for a window, both cores are in a mutual spin and the run has settled.
	visited := map[uint32]bool{}
	m.visited = visited
	var lastProgress uint64
	prevSig := m.progressSig()
	prevPages := 0

	for m.Steps < budget {
		// deliver interrupts to parked or running cores
		m.deliver(m.ARM9)
		m.deliver(m.ARM7)

		// synthetic VBlank
		if m.Steps >= m.vblankAt {
			m.ARM9.if_ |= irqVBlank
			m.ARM7.if_ |= irqVBlank
			m.vblankAt += vblankPeriod
		}

		before := len(res.ARM9Milest)
		m.runQuantum(m.ARM9, quantum, milestones, res.ARM9Milest)
		m.runQuantum(m.ARM7, quantum, nil, nil)
		m.Steps += uint64(quantum)

		if m.ARM9.cpu.Halted {
			res.Reason = "ARM9 halted: " + m.ARM9.cpu.HaltReason
			break
		}
		if m.ARM7.cpu.Halted {
			res.Reason = "ARM7 halted: " + m.ARM7.cpu.HaltReason
			break
		}

		sig := m.progressSig()
		if sig != prevSig || len(visited) != prevPages || len(res.ARM9Milest) != before {
			prevSig, prevPages = sig, len(visited)
			lastProgress = m.Steps
		} else if m.Steps-lastProgress > 24_000_000 { // > any single WaitByLoop sleep
			res.Reason = fmt.Sprintf("settled — ARM9 spinning at 0x%08X (%s), ARM7 at 0x%08X (%s); no new code or IPC traffic",
				m.ARM9.cpu.R[15], parkState(m.ARM9), m.ARM7.cpu.R[15], parkState(m.ARM7))
			break
		}
	}
	if res.Reason == "" {
		res.Reason = fmt.Sprintf("step budget (%d) reached", budget)
	}
	res.Steps = m.Steps
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
		pc := c.cpu.R[15]
		if c.arm9 && m.visited != nil {
			m.visited[pc>>8] = true
		}
		if milestones != nil {
			if _, ok := milestones[pc]; ok {
				if _, seen := hit[pc]; !seen {
					hit[pc] = m.Steps + uint64(i)
					fmt.Printf("  ARM9 reached %-20s 0x%08X  (round %d)\n", milestones[pc], pc, m.Steps+uint64(i))
				}
			}
		}
		// halt-on-return: main() falling through to the BIOS exit vector
		if pc >= 0xFFFF0000 || (!c.arm9 && pc < 0x20 && m.Steps > 1000) {
			c.cpu.Halt("%s: returned to BIOS vector 0x%08X", c.name, pc)
			return
		}
		c.cpu.Step()
	}
}

// deliver dispatches a pending, unmasked interrupt to a core: it mimics the DS BIOS
// IRQ path — set the OS "IRQ check flag", read the user handler pointer from the
// fixed slot, and enter the handler in IRQ mode. A parked core resumes just past
// its wait once the handler returns.
func (m *Machine) deliver(c *core) {
	pending := c.ie & c.if_
	if pending == 0 || !c.ime {
		return
	}
	if c.waiting {
		// wake only for the awaited sources (Halt waits for any)
		if !c.waitAny && c.waitMask != 0 && pending&c.waitMask == 0 {
			return
		}
	} else if c.cpu.IRQDisable {
		return // running with IRQs masked: let the code finish its critical section
	}
	b := &bus{c: c}
	// OS IRQ check flag (BIOS ORs the acknowledged sources here; IntrWait polls it)
	flag := b.r32(c.handlerBase - 8)
	b.w32(c.handlerBase-8, flag|pending)
	handler := b.r32(c.handlerBase - 4)
	if handler < mainBase { // no handler installed yet: just latch the flag
		return
	}
	ret := c.cpu.R[15]
	if c.waiting {
		ret = c.resumePC
		c.waiting = false
	}
	// the handler returns via `subs pc, lr, #4`, so LR = returnAddr + 4
	c.cpu.Exception(arm.ModeIRQ, handler, ret+4)
}

// progressSig fingerprints IPC-level progress: the sync nibbles, FIFO depths and
// pending-interrupt flags. It deliberately omits the PCs (a tight spin oscillates
// them without making progress); code progress is tracked separately by the set of
// ARM9 pages entered. When neither changes for a window, the run has settled.
func (m *Machine) progressSig() uint64 {
	return uint64(m.ipc.sync9)<<1 ^ uint64(m.ipc.sync7)<<5 ^
		uint64(len(m.ipc.to7))<<8 ^ uint64(len(m.ipc.to9))<<12 ^
		uint64(m.ARM9.if_)<<16 ^ uint64(m.ARM7.if_)<<32 ^
		uint64(len(m.ARM9.ioseq))<<40 ^ uint64(len(m.ARM7.ioseq))<<48
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
