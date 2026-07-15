package ps2

// iopthreads.go is the thread-state inspector: it walks THREADMAN's control blocks and
// says, for every thread on the second processor, what state it is in and where it is
// parked. It is the instrument the memory-card probe needed and did not have — a way to
// look at a *blocked* thread and read the PC it went to sleep at, without which a
// deadlock between three threads on two servers looks from every other angle like a busy
// machine going nowhere.
//
// Nothing here is guessed. The layout was read off THREADMAN's own scheduler code and its
// live control blocks on a frozen title state:
//
//   - The reschedule predicate (registered through intrman #30, so we hold its address in
//     schedResched) is four instructions, and its first two form the address of the pair
//     of globals the scheduler compares:
//         lui   $v1, HI
//         addiu $v1, $v1, LO      -> G = (HI<<16) + LO
//         lw    $v0, 0($v1)       -> the thread that OUGHT to run   [G]
//         lw    $v1, -4($v1)      -> the thread that IS running     [G-4]
//     So the running thread's control block is *(G-4), read straight from the module's own
//     code rather than from a constant that moves the day a module ahead of it grows.
//
//   - Every control block carries a link at +0x24 to the thread created before it, newest
//     to oldest — the master chain. The running ISOThread points at Thread_Loader points
//     at Thread_Player points at Thread_Server..., none of them in priority order, so the
//     link is a registry of every thread and not a run queue. Walking it from the head
//     visits them all. (The head is the newest thread; threads made after the one running
//     — the card and controller workers — sit above it, so a walk that started at the
//     running thread would miss exactly the threads a card probe cares about. The head is
//     found by scan, below, and cross-checked against the chain.)
//
//   - The fields that matter, from comparing a running block against a blocked one:
//         +0x0C  (priority<<16) | state   state: 1 RUN, 2 READY, 4 WAIT, 8 SUSPEND,
//                                                 0xC WAIT|SUSPEND, 0x10 DORMANT
//         +0x10  the saved register frame — the same 184-byte frame the interrupt path
//                builds (iopframe.go); frame+140 is the EPC, the PC the thread resumes at.
//                Only meaningful for a thread that is not the one currently running, whose
//                live PC is in the CPU.
//         +0x24  the master-chain link (above)
//         +0x1C  the wait type     — 4 when the thread is blocked on an event flag
//         +0x20  the wait object   — the id of that event flag (or semaphore, ...)
//         +0x28  the wait bits     — the mask an event-flag waiter is blocked on: the
//                                    memory-card worker's 0x155 and mcserv's 0x8 name the
//                                    two ends of the handshake that never completes
//         +0x38  the entry point   — what the thread runs, and so its name
//         +0x3C  the stack base    +0x40 its size; the frame lives inside this range
//
// The saved PC of a blocked thread is the whole point: an event-flag wait parks the thread
// at the return of the reschedule syscall inside THREADMAN's WaitEventFlag, and the frame's
// registers still hold the flag id and the bits it is waiting for. Reading those turns
// "the worker never completes the handshake" into "the worker is asleep at THREADMAN+0xNNNN
// waiting for bits 0x155 on flag 0xIIIIIIII" — which is either the answer or the next question.

import (
	"fmt"
	"sort"
	"strings"
)

// The control-block field offsets, named.
const (
	iopTCBStatus  = 0x0C // (priority<<16) | state
	iopTCBFrame   = 0x10 // the saved register frame; frame+iopFrameEPC = resume PC
	iopTCBWaitTyp = 0x1C // what kind of thing a WAIT thread is blocked on (4 = event flag)
	iopTCBWaitObj = 0x20 // the id of that thing (the event flag / semaphore)
	iopTCBWaitMsk = 0x28 // the event-flag bits an event-flag waiter is blocked on
	iopTCBChain   = 0x24 // link to the thread created before this one (newest -> oldest)
	iopTCBEntry   = 0x38 // the thread's entry point
	iopTCBStack   = 0x3C // its stack base
	iopTCBStackSz = 0x40 // its stack size
)

// iopThreadStates names the state byte at TCB+0x0C. They are the IOP's own THS_ values,
// arrived at by watching which byte the running thread carries (1) versus a blocked one (4).
var iopThreadStates = map[uint32]string{
	0x01: "RUN",
	0x02: "READY",
	0x04: "WAIT",
	0x08: "SUSPEND",
	0x0C: "WAIT|SUSPEND",
	0x10: "DORMANT",
}

// currentTCB returns the control block of the thread the processor is running, read out of
// the global the reschedule predicate compares — which is the module's own answer to the
// question, not ours.
func (p *IOP) currentTCB() uint32 {
	if p.schedIsRun == 0 {
		return 0
	}
	return p.Read32(p.schedIsRun) // the "is running" pointer, derived when schedResched registered
}

// looksLikeTCB reports whether the aligned word at `a` is a thread control block. It is a
// structural test — a valid state, a sane priority, an entry point inside a loaded module,
// and a stack the frame sits within — used only to find the *set* of blocks so the chain
// walk can be checked against it. The chain (a real pointer THREADMAN maintains) is the
// authority; this is the net that proves the walk missed nothing.
func (p *IOP) looksLikeTCB(a uint32) bool {
	status := p.Read32(a + iopTCBStatus)
	state := status & 0xFF
	prio := status >> 16
	if _, ok := iopThreadStates[state]; !ok {
		return false
	}
	if status&0x0000FF00 != 0 || prio > 0x7F {
		return false
	}
	if !p.inModule(p.Read32(a + iopTCBEntry)) {
		return false
	}
	stack := p.Read32(a + iopTCBStack)
	size := p.Read32(a + iopTCBStackSz)
	if stack < iopModuleBase || stack >= iopRAMSize || size == 0 || size > 0x40000 {
		return false
	}
	frame := p.Read32(a + iopTCBFrame)
	// A running thread's frame slot is stale, so only require containment when it is set.
	if frame != 0 && (frame < stack || frame >= stack+size) {
		return false
	}
	return true
}

// inModule reports whether an address falls inside a loaded module's image.
func (p *IOP) inModule(addr uint32) bool {
	for _, m := range p.modules {
		if addr >= m.Base && addr < m.Base+m.Size {
			return true
		}
	}
	return false
}

// iopThread is one thread, as the inspector reports it.
type iopThread struct {
	tcb, prio, state, entry, stack, frame, savedPC, waitTyp, waitObj, waitMsk uint32
	running                                                                   bool
}

// IOPThreads walks every thread on the second processor and reports its state and the PC it
// is parked at. It is the answer to "what is every thread waiting for", asked of a machine
// that is doing exactly what it should and getting nowhere.
func (p *IOP) IOPThreads() string {
	cur := p.currentTCB()

	// Find every control block by structural scan, then confirm the master chain reaches
	// all of them. The scan gives completeness (it cannot miss a block in a heap chunk the
	// chain would have to be walked from its head to reach); the chain gives confidence
	// (a block the scan turned up that the chain does not link is not a thread).
	var tcbs []uint32
	seen := map[uint32]bool{}
	for a := uint32(iopModuleBase); a < iopRAMSize; a += 4 {
		if p.looksLikeTCB(a) {
			tcbs = append(tcbs, a)
			seen[a] = true
		}
	}
	sort.Slice(tcbs, func(i, j int) bool { return tcbs[i] < tcbs[j] })

	// The chain, walked from the highest block found (the newest thread). Its length is the
	// check: a chain that visits every scanned block, and only those, is the whole list.
	chainLen, chainOK := 0, true
	if len(tcbs) > 0 {
		visited := map[uint32]bool{}
		for a := tcbs[len(tcbs)-1]; a != 0 && seen[a] && !visited[a]; a = p.Read32(a + iopTCBChain) {
			visited[a] = true
			chainLen++
		}
		chainOK = chainLen == len(tcbs)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "the IOP's threads (%d; running = %s):\n", len(tcbs), p.Sym(p.CPU.PC))
	for _, a := range tcbs {
		t := p.readThread(a, cur)
		mark := " "
		if t.running {
			mark = "*"
		}
		fmt.Fprintf(&b, "  %s %08X  p%-3d %-12s  %-28s  @ %s\n",
			mark, a, t.prio, iopThreadStates[t.state], p.Sym(t.entry), p.Sym(t.savedPC))
		if t.state == 0x04 || t.state == 0x0C { // WAIT: say what on, and where it went to sleep
			if t.waitTyp == 4 {
				fmt.Fprintf(&b, "        waiting on event flag %08X, bits %X\n", t.waitObj, t.waitMsk)
			} else {
				fmt.Fprintf(&b, "        waiting (type %d) on %08X\n", t.waitTyp, t.waitObj)
			}
		}
	}
	if !chainOK {
		fmt.Fprintf(&b, "  (note: the +0x24 chain from the newest block visits %d of %d — the scan and the chain disagree)\n", chainLen, len(tcbs))
	}
	return b.String()
}

// readThread reads one control block into an iopThread, resolving the saved PC from the
// frame — except for the running thread, whose live PC is in the processor, not the frame.
func (p *IOP) readThread(a, cur uint32) iopThread {
	status := p.Read32(a + iopTCBStatus)
	t := iopThread{
		tcb:     a,
		prio:    status >> 16,
		state:   status & 0xFF,
		entry:   p.Read32(a + iopTCBEntry),
		stack:   p.Read32(a + iopTCBStack),
		frame:   p.Read32(a + iopTCBFrame),
		waitTyp: p.Read32(a + iopTCBWaitTyp),
		waitObj: p.Read32(a + iopTCBWaitObj),
		waitMsk: p.Read32(a + iopTCBWaitMsk),
		running: a == cur,
	}
	if t.running {
		t.savedPC = p.CPU.PC
	} else if t.frame != 0 {
		t.savedPC = p.Read32(t.frame + iopFrameEPC)
	}
	return t
}
