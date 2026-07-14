package ps2

// iopframe.go is the register frame an interrupt saves a thread into, and it is the
// contract that lets Sony's own scheduler preempt on this machine.
//
// intrman registers two hooks with THREADMAN — #30 a predicate, #28 a switch routine —
// and until now they were recorded and never called, because calling them without
// understanding what they expect is worse than not calling them at all. (It was tried.
// The scheduler duly switched threads, the machine duly threw the switch away by
// restoring the context it had saved in Go, and THREADMAN spent the rest of the boot
// believing a thread was running that was not.)
//
// The contract was read out of THREADMAN, which is the only authority on it. Three
// routines say the whole thing.
//
// The predicate (THREADMAN+0xBC4) is four instructions:
//
//	lw   $v0, 0($v1)        the thread that ought to be running
//	lw   $v1, -4($v1)       the thread that is
//	xor  $v0, $v0, $v1
//	sltu $v0, $zero, $v0    "do they differ?"
//
// The switch routine (THREADMAN+0x940) takes a pointer in $a0 and does this with it:
//
//	lw $v0, -28880($v0)     the current thread's control block
//	sw $s1, 16($v0)         *** the pointer goes into TCB+16 ***
//	lw $v0, 60($v0)         the thread's stack base
//	sltu $v0, $s1, $v0      and the pointer is checked against it
//
// — so $a0 is a pointer to something living on the interrupted thread's own stack. And
// its epilogue:
//
//	lw $v0, -28876($v0)     the thread that is now current
//	lw $v0, 16($v0)         *** its TCB+16 ***
//	jr $ra                  *** and that is the return value ***
//
// So: the kernel saves the interrupted context somewhere on its stack, hands the switch
// routine the pointer, and the switch routine gives back the pointer for the thread that
// should run next. The kernel restores from that. That is the whole of preemption on the
// IOP, and none of it needed to be guessed.
//
// The *layout* of the frame comes from StartThread, which has to build one from nothing so
// that a brand-new thread can be resumed by the same code that resumes an interrupted one:
//
//	addiu $a2, $zero, 184      it memsets 184 bytes
//	... at (stackBase + stackSize - 184)
//	sw $a0, 16($s0)            and stores that address in TCB+16
//	sw $s2, 16($v0)            frame+16  = the thread's argument   -> $a0  is register 4
//	sw $v0, 112($v1)           frame+112 = the thread's $gp        -> $gp  is register 28
//	sw $v0, 116($v1)           frame+116 = frame + 152             -> $sp  is register 29
//	sw $v0, 120($v1)           frame+120 = the same                -> $fp  is register 30
//	sw $v0, 124($v1)           frame+124 = a THREADMAN address     -> $ra  is register 31
//	sw $v0, 136($v1)           frame+136 = a status word
//	sw $v0, 140($v1)           frame+140 = the thread's entry point
//
// Every one of those lands on 4*n for the register n it holds. The frame is the register
// file at four bytes apiece, with the slot that would belong to $zero used as a tag, and
// the machine's own state above it.

// The frame: 184 bytes, and where the parts of a context sit in it.
const (
	iopFrameSize = 184

	iopFrameTag = 0   // the $zero slot, which no register needs. THREADMAN writes -2.
	iopFrameHI  = 128 // above the registers: the multiplier's two halves,
	iopFrameLO  = 132
	iopFrameSR  = 136 // the status the thread runs with,
	iopFrameEPC = 140 // and where it resumes.
)

// iopFrameFresh is what THREADMAN writes into the tag slot of a frame it has just built.
// Nothing on this disc reads it back, and its meaning has not been established — so the
// machine writes what THREADMAN writes rather than inventing a value of its own.
const iopFrameFresh = 0xFFFFFFFE

// iopThreadExit is the routine a thread returns to when its entry point returns, and it is not
// a number anybody had to guess.
//
// StartThread builds a frame for a brand-new thread, and one of the words it puts in is the
// thread's $ra — slot 31, at offset 124. It is the same address for every thread the machine
// has ever started (the frames say so: `ra=THREADMAN+0x123C`, on all of them), and it is by
// construction the answer to "what does a thread do when it is finished", because it is where
// control goes when the thread's own entry function does the one thing it must eventually do.
//
// It is found rather than hardcoded, and the search is the derivation. A thread's stack comes
// out of AllocSysMemory, which is ours, so the machine knows every block it ever handed
// THREADMAN; and StartThread puts the frame at the *top* of the stack it was given, tagged with
// a value of its own (iopFrameFresh) that nothing else writes. So: look at the top of every
// block, keep the ones that carry the tag, and read slot 31 out of each. They agree, and what
// they agree on is the answer — which is a far better authority than an offset copied out of a
// listing, because it is still right the day a module moves and the listing is not.
func (p *IOP) findThreadExit() uint32 {
	counts := map[uint32]int{}
	for _, b := range p.blocks {
		if b.size < iopFrameSize {
			continue
		}
		frame := (b.base + b.size - iopFrameSize) &^ 7
		if p.Read32(frame+iopFrameTag) != iopFrameFresh {
			continue
		}
		counts[p.Read32(frame+4*31)]++
	}
	best, n := uint32(0), 0
	for addr, c := range counts {
		if c > n {
			best, n = addr, c
		}
	}
	return best
}

// exitLoaderThread is the last thing the loader does, and without it the IOP never runs a
// thread of its own.
//
// The problem it solves is the one intrDeliver's long comment describes, arriving at the end
// rather than the middle. Every module's entry point ran as a bare Go call, borrowing whichever
// thread THREADMAN believed was current; each of them created its threads, started them, and
// returned. The threads are all ready. But THREADMAN still believes the borrowed thread is
// running — because from where it stands, it is — and that thread outranks everything the
// modules made, so its reschedule predicate says no switch is wanted and every interrupt exits
// without scheduling anybody. The profile shows it plainly: ninety-five per cent of the second
// processor's life spent in the kernel HLE's own idle loop, and *zero* thread switches in a
// boot with twelve modules and a dozen threads in it.
//
// On the board there is no such gap, because the code that loads the modules is itself a
// thread, and when it has finished it stops being runnable. So that is what this does. It sends
// the borrowed thread to the address THREADMAN itself nominates for a thread that has finished
// — the $ra it writes into every frame it builds — and lets THREADMAN take it off the run queue
// by its own rules. The next thing the processor runs is the first real thread of its life.
func (p *IOP) exitLoaderThread() {
	exit := p.findThreadExit()
	if exit == 0 {
		p.ps2.note("IOP: no thread was ever started, so the loader has nothing to stand down through")
		return
	}
	p.CPU.SetPC(exit)
	p.CPU.SetReg(31, iopIdleLoop) // if it ever does return, the idle loop is where it belongs
	p.ps2.note("IOP: the loader stands down through %s, and the scheduler takes over", p.Sym(exit))
}

// saveFrame writes the interrupted context into a frame at `at`.
func (p *IOP) saveFrame(at uint32) {
	st := p.CPU.SaveState()

	p.Write32(at+iopFrameTag, iopFrameFresh)
	for i := uint32(1); i < 32; i++ {
		p.Write32(at+4*i, st.R[i])
	}
	p.Write32(at+iopFrameHI, st.HI)
	p.Write32(at+iopFrameLO, st.LO)

	// The status word, and it is not decoration. It carries the one piece of processor state an
	// interrupt has to remember — whether interrupts were on — in bit 2, and THREADMAN agrees:
	// every frame it builds for a new thread carries sr = 0x00000404, and bit 2 is the bit that
	// is set. (On an R3000A that is IEp, the "previous" half of the interrupt-enable stack, and
	// bit 2 is where an exception return takes the enable back from. So the convention is the
	// hardware's, arrived at from both ends independently — Sony's scheduler writes it because
	// that is what rfe reads, and this writes it because that is what Sony's scheduler writes.)
	p.Write32(at+iopFrameSR, b2u(p.intrEnabled)<<2)

	// Where the thread resumes. The core's PC is the instruction that has not run yet,
	// which is exactly the right answer — provided it is not a branch delay slot, and
	// serviceIntr guarantees that by declining to interrupt one.
	p.Write32(at+iopFrameEPC, st.PC)
}

// loadFrame restores a context from a frame — which, after a switch, is a different
// thread's than the one that was saved.
func (p *IOP) loadFrame(at uint32) {
	st := p.CPU.SaveState()

	for i := uint32(1); i < 32; i++ {
		st.R[i] = p.Read32(at + 4*i)
	}
	st.R[0] = 0
	st.Out = st.R // no load is in flight: see serviceIntr
	st.HI = p.Read32(at + iopFrameHI)
	st.LO = p.Read32(at + iopFrameLO)

	pc := p.Read32(at + iopFrameEPC)
	st.PC, st.NextPC, st.CurPC = pc, pc+4, pc

	// A resumed context is never mid-branch and never mid-load, because a context is only
	// ever saved when it is neither. Clearing these says so, rather than carrying the
	// interrupted thread's pipeline into the thread being resumed — which is the kind of
	// bug that shows up once in ten thousand switches and looks like a compiler error.
	st.DelaySlot, st.PendingDelay, st.BranchAddr = false, false, 0
	st.LdReg, st.LdVal = 0, 0

	p.CPU.LoadState(st)

	// And the interrupt-enable, from the frame — which is to say, from the thread being
	// resumed, and not from whatever the thread being left behind happened to be doing.
	//
	// This is the exception return, and it is the half of it that was missing. A thread's
	// interrupt state belongs to the thread: it is saved in its frame when it goes away and it
	// comes back with it. Taking the enable from the machine's own record instead works for as
	// long as every switch is a *preemption* — the interrupted thread is coming straight back,
	// so its state and the machine's are the same thing — and it fails the moment a thread
	// gives the processor up on purpose, because the code that does that is inside the kernel
	// with interrupts masked, and it never comes back to unmask them.
	//
	// Which is exactly what the loader standing down does. ExitThread suspends interrupts,
	// takes the thread off the run queue and switches away; the thread that arrives inherits
	// the mask, runs forever with the door shut, and the EE's doorbell is raised twice and
	// delivered never. Every thread THREADMAN starts is meant to run with interrupts on — it
	// says so, in bit 2 of every frame it builds — and now it does.
	p.intrEnabled = p.Read32(at+iopFrameSR)>>2&1 != 0
}
