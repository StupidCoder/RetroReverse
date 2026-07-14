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

// saveFrame writes the interrupted context into a frame at `at`.
func (p *IOP) saveFrame(at uint32) {
	st := p.CPU.SaveState()

	p.Write32(at+iopFrameTag, iopFrameFresh)
	for i := uint32(1); i < 32; i++ {
		p.Write32(at+4*i, st.R[i])
	}
	p.Write32(at+iopFrameHI, st.HI)
	p.Write32(at+iopFrameLO, st.LO)

	// The status word. This machine's intrman is Go and has no COP0 status register to
	// save, so what goes here is the one bit of processor state an interrupt actually has
	// to remember — whether interrupts were on — written where THREADMAN puts it. Nothing
	// reads it: the restore takes the enable back from the machine's own record. It is
	// filled in because a frame with a hole in it is a frame someone will one day read.
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
}
