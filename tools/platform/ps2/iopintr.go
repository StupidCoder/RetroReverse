package ps2

import "fmt"

// iopintr.go is intrman, sysmem, heaplib and sysclib: the four libraries the IOP's
// modules cannot take a step without, and the four we have to write.
//
// Everything here was read out of the code that calls it. Sony's kernel modules are
// stripped and their functions are imported by number, so an identification is a small
// argument rather than a lookup, and each one is recorded where it was made. The
// arguments are short because the callers are honest: a routine that passes a pointer,
// a zero and its own .bss size and then relies on the memory being clear is memset, and
// it is not much else.
//
// What is *not* here matters as much. intrman #8, #9, #14 and #23, sysmem #5 to #8,
// heaplib #5, #7 and #8, and most of sysclib are still unmodelled — called, counted,
// and answered with zero. They are in the census, and the census is the work list.

func init() {
	// intrman. The interrupt controller.
	//
	// The two that matter most are the pair every module brackets its critical sections
	// with, and they are named by their own call convention: #17 is passed a *pointer*
	// to a stack slot and #18 is passed the value that was in it. That is
	// CpuSuspendIntr(int *old) / CpuResumeIntr(int old), and nothing else has that
	// shape.
	//
	// #23 is QueryIntrContext, and it is named by the only use its answer is ever put to.
	// LIBSD calls it at the top of its transfer routine and branches on the result to one
	// of *two versions of the same THREADMAN call* — thevent#8 on one side, thevent#9 on
	// the other, with identical arguments. A kernel exports a pair like that for exactly
	// one reason: one of them is safe to call from inside an interrupt and the other is
	// not, and the caller has to know which it is. So the question being asked is "am I in
	// an interrupt?", and the answer is how deep in a handler we are.
	//
	// #28 and #30 are THREADMAN's hooks into interrupt exit. It hands each of them one of
	// its own functions, and the one it hands to #30 is four instructions long: it compares
	// two of THREADMAN's globals and returns whether they differ. That is "does the
	// scheduler want to switch threads?" — so #30 registers the predicate the kernel
	// consults on the way out of an interrupt, and #28 the routine that acts on it. Both
	// are called now, on the path in intrDeliver.
	lib("intrman", map[uint16]iopFunc{
		4:  {"RegisterIntrHandler", (*IOP).intrRegister},
		5:  {"ReleaseIntrHandler", (*IOP).intrRelease},
		6:  {"EnableIntr", (*IOP).intrEnable},
		7:  {"DisableIntr", (*IOP).intrDisable},
		8:  {"CpuDisableIntr", (*IOP).intrCpuDisable},
		9:  {"CpuEnableIntr", (*IOP).intrCpuEnable},
		14: {"CpuInvokeInKmode", (*IOP).intrInvokeInKmode},
		17: {"CpuSuspendIntr", (*IOP).intrSuspend},
		18: {"CpuResumeIntr", (*IOP).intrResume},
		23: {"QueryIntrContext", (*IOP).intrQueryContext},
		28: {"<register scheduler hook>", (*IOP).intrSetSwitchHook},
		30: {"<register reschedule predicate>", (*IOP).intrSetReschedHook},
	})

	// sysmem. The allocator, and — because it is the lowest library in the machine and
	// the one every other module can already see — the kernel's printf.
	//
	// #14 takes a pointer into a module's .rodata and up to three more registers, and
	// THREADMAN calls it on a debug path with the current thread's id. It is a printf.
	// That there are two of them (this and stdio #4) is not a mistake: the IOP has a
	// kernel printf and a library printf, and modules import whichever they can reach.
	lib("sysmem", map[uint16]iopFunc{
		4:  {"AllocSysMemory", (*IOP).sysmemAlloc},
		5:  {"FreeSysMemory", (*IOP).sysmemFree},
		6:  unknown(),
		7:  unknown(),
		8:  unknown(),
		9:  {"QueryBlockTopAddress", (*IOP).sysmemBlockTop},
		10: {"QueryBlockSize", (*IOP).sysmemBlockSize},
		14: {"Kprintf", (*IOP).sysmemKprintf},
	})

	// heaplib. A heap on top of sysmem. THREADMAN creates one and then takes every
	// thread control block, semaphore and event flag out of it.
	//
	// #6 is named by what its caller does next: THREADMAN checks the returned pointer
	// against zero and then memsets that many bytes at it. An allocator, and one that
	// does not clear.
	lib("heaplib", map[uint16]iopFunc{
		4: {"CreateHeap", (*IOP).heapCreate},
		5: unknown(),
		6: {"AllocHeapMemory", (*IOP).heapAlloc},
		7: unknown(),
		8: unknown(),
	})

	// sysclib. The C library.
	//
	// Four of these are earned, and the three new ones were all earned from IOMAN, which is
	// the module that turns a path into a device. Between them they are the reason the second
	// processor could not open a file — any file, on any device — and the reason it said so
	// in a message that named the path and blamed the disc.
	//
	// #14 is memset: THREADMAN's first act is to call it with its own .bss, a zero, and
	// 1220 — which is exactly the size of its .bss — and then to rely on the result
	// being clear.
	//
	// #12 is memcpy: FILEIO calls it with the RPC buffer the EE has just DMA'd 32 bytes into,
	// a length of exactly 32, and then opens the path at dst+20 — which is exactly where the
	// path sits in the source.
	//
	// #17 is bzero: IOMAN's first act after registering its library is #17(0x1F1A0, 64), and
	// 0x1F1A0 is the address IOMAN later reads as the head of its driver list (IOMAN+0x1854).
	// A 64-byte table, cleared once, before anything can be added to it.
	//
	// #22 is strcmp: IOMAN walks that list of drivers (IOMAN+0x186C loads each one's name) and
	// calls #22(the device it is looking for, this driver's name), taking a *zero* answer as a
	// match and any other as "try the next one". Answered with zero it matched the first driver
	// in the list, whatever it was — which is a lie that would have been much harder to see
	// than the one in front of it.
	//
	// #25 is strchr: IOMAN calls #25(path, 58) — and 58 is ':'. If the answer is null it
	// prints "Unknown device '%s'" and gives up, which is precisely what the second processor
	// had been doing for rom0:, tty00: and, once the drive was working, for the game's own
	// cdrom0:\DRIVERS\SIO2MAN.IRX;1. Every path on this machine was a path with no device in
	// it, because the function that finds the colon always said there wasn't one.
	// And the two that finish the job, both from the same six instructions of IOMAN:
	//
	//	IOMAN+0x1694   #30(sp+16, path, colon - path)   copy the device name out of the path
	//	IOMAN+0x16A4   a0 = the LAST byte of that copy
	//	IOMAN+0x16A8   #8(a0)
	//	IOMAN+0x16B0   v0 & 4
	//
	// #30 is strncpy: the length is the distance from the start of the path to the colon that
	// #25 just found, and the instruction in the call's delay slot writes the NUL terminator
	// by hand at dst[n] — which is exactly the thing strncpy leaves for you to do.
	//
	// #8 is a character-class lookup. It is handed one character — the last of "cdrom0" — and
	// IOMAN tests bit 2 of the answer, which is how it decides that the trailing 0 is a unit
	// number and the driver it wants is called "cdrom". So bit 2 means "digit". That is the
	// only bit any caller on this disc tests, and the only one that has been earned; the rest
	// of the table is left clear rather than filled in from what a C library usually looks
	// like, because a bit invented here is a bit that will be believed later.
	// And two more from CDVDMAN's open, which is the routine that decides whether a filename
	// on this disc is a filename at all:
	//
	//	CDVDMAN+0x510   n = #27(path)          one argument, and the answer is a length
	//	CDVDMAN+0x520   if n < 3: goto append
	//	CDVDMAN+0x528   the character at path[n-2] — is it ';' ?
	//	CDVDMAN+0x538   the character at path[n-1] — is it '1' ?
	//	CDVDMAN+0x550   append: #20(path, ";1")
	//
	// #27 is strlen. #20 is strcat, and the string it is handed is at 0x29128 and reads ";1" —
	// ISO 9660's version suffix. CDVDMAN is checking that the name ends in ";1" and adding it
	// if it does not.
	//
	// Answered with zero, strlen made every path shorter than three characters, so CDVDMAN
	// skipped the check it could not do and appended ";1" to a name that already had one. What
	// it said afterwards was "open fail name \DRIVERS\SIO2MAN.IRX;1" — a message that names the
	// file, blames the disc, and is off by two characters it printed before it added them.
	lib("sysclib", map[uint16]iopFunc{
		8:  {"look_ctype_table", (*IOP).clibCtype},
		11: unknown(),
		12: {"memcpy", (*IOP).clibMemcpy},
		13: {"memmove", (*IOP).clibMemmove},
		14: {"memset", (*IOP).clibMemset},
		17: {"bzero", (*IOP).clibBzero},
		19: unknown(),
		20: {"strcat", (*IOP).clibStrcat},
		22: {"strcmp", (*IOP).clibStrcmp},
		23: {"strcpy", (*IOP).clibStrcpy},
		25: {"strchr", (*IOP).clibStrchr},
		27: {"strlen", (*IOP).clibStrlen},
		29: {"strncmp", (*IOP).clibStrncmp},
		30: {"strncpy", (*IOP).clibStrncpy},
		36: unknown(),
	})
}

// --- intrman ----------------------------------------------------------------------

// iopIRQs is how big intrman's vector table is, and it is bigger than the interrupt
// controller.
//
// The chip has 32 lines. intrman offers more, and the extra ones are the DMA channels:
// the controller raises a single line for "some channel finished" and the kernel demuxes
// it, so from a driver's point of view each channel has an interrupt of its own. The
// numbers were read off the registrations, and four modules agree:
//
//	36  LIBSD    channel 4    the sound chip's first core
//	40  LIBSD    channel 7    its second
//	42  SIFMAN   channel 9    SIF0
//	43  SIFCMD   channel 10   SIF1
//
// which is 32 + n for the first block of channels and 40 + (n-7) for the second — eight
// numbers to a block, with the second block's numbering starting at a round 40 rather than
// carrying on from the first. That is not a guess between two readings; it is the only
// arithmetic that fits all four, and it independently confirms the channel identities
// derived from DPCR (iopdma.go), which were established from an entirely different
// register by an entirely different argument.
//
// A table of 32 does not merely lose these. It loses them *silently*: RegisterIntrHandler
// is handed 36, finds it out of range, returns an error the caller does not check, and the
// sound chip's completion interrupt is never registered at all. The handler that would have
// run is the one that calls the callback that sets the flag OVERLORD is spinning on.
const iopIRQs = 64

// iopHandler is one registered interrupt handler: an entry in the vector table, and
// nothing more.
//
// Whether the line is *unmasked* is deliberately not in here, and the reason is worth
// recording, because putting it here is a bug that hides perfectly. On the hardware the
// vector table and the mask register are two different pieces of state, and a driver may
// write them in either order — LIBSD writes them in the order nobody expects, calling
// EnableIntr(9) and only then RegisterIntrHandler(9). With the two conflated in one
// struct, registering the handler stores a fresh record over the top of the enable that
// had already happened, and the sound chip's line is masked from that moment on. The
// symptom is not a masked interrupt. The symptom is OVERLORD waiting forever for a DMA
// that completed, on a machine where the DMA controller, the SPU and the interrupt path
// are all working.
type iopHandler struct {
	fn  uint32 // the module's own routine
	arg uint32
}

func (p *IOP) intrRegister() {
	irq, _, fn, arg := p.arg(0), p.arg(1), p.arg(2), p.arg(3)
	if irq >= iopIRQs {
		p.ps2.note("IOP: a handler was registered on interrupt %d, which is past the end of the table", irq)
		p.setRet(0xFFFFFF9A) // out of range; the caller checks
		return
	}
	p.handlers[irq] = iopHandler{fn: fn, arg: arg}
	p.ps2.note("IOP: interrupt %d handled by %s", irq, p.Sym(fn))
	p.setRet(0)
}

func (p *IOP) intrRelease() {
	if irq := p.arg(0); irq < iopIRQs {
		p.handlers[irq] = iopHandler{}
	}
	p.setRet(0)
}

// intrLine reads the interrupt line out of an argument to EnableIntr or DisableIntr.
//
// The line is in the low bits, and the bits above it are not part of it. SIFCMD is the only
// module on the disc that sets any: it registers its handler on 43, and its teardown disables
// 43 and releases 43 — but the call that *enables* it passes 555, which is 43 with bit 9 set,
// as a literal in the delay slot. Three calls name the line plainly and one decorates it, so
// the decoration cannot itself be a line; it is a flag the real intrman masks away. SIFMAN,
// doing the same job one DMA channel along, passes a plain 42 — which is what says the flag is
// not required, and that nothing is lost by ignoring it.
//
// Range-checking the raw argument instead — which is what this did — silently drops the call.
// And of everything that can be dropped silently on this machine, this is the worst, because
// interrupt 43 is the EE's doorbell into the IOP. Every module that talks to the other
// processor ends its initialisation in SIFCMD, waiting on an event flag that SIFCMD's handler
// sets; that handler is on 43; and a masked 43 means it never runs. What that looks like from
// outside is not a masked interrupt. It is five threads and a module entry point all blocked,
// and a scheduler quite correctly running its idle thread — a machine that is doing exactly
// what it should, and getting nowhere.
func (p *IOP) intrLine(arg uint32) uint32 { return arg & (iopIRQs - 1) }

func (p *IOP) intrEnable() {
	p.imask |= 1 << uint64(p.intrLine(p.arg(0)))
	p.setRet(0)
}

func (p *IOP) intrDisable() {
	p.imask &^= 1 << uint64(p.intrLine(p.arg(0)))
	p.setRet(0)
}

// intrSuspend is CpuSuspendIntr(int *old): it saves the current interrupt-enable state
// through the pointer it is given, disables interrupts, and returns success.
//
// The pointer is the whole of the identification. Every module brackets its critical
// sections with this and its partner, and the partner is passed exactly the word this
// one wrote.
func (p *IOP) intrSuspend() {
	old := b2u(p.intrEnabled)
	if ptr := p.arg(0); ptr != 0 {
		p.Write32(ptr, old)
	}
	p.intrEnabled = false
	p.ieEvent("suspend", p.arg(0), old, p.CPU.Reg(31))
	p.setRet(0)
}

// intrResume is CpuResumeIntr(int old).
func (p *IOP) intrResume() {
	p.intrEnabled = p.arg(0) != 0
	p.ieEvent("resume", 0, p.arg(0), p.CPU.Reg(31))
	p.setRet(0)
}

// intrCpuDisable and intrCpuEnable are CpuDisableIntr() and CpuEnableIntr(): the
// processor's master interrupt switch, as against the per-line mask that #6 and #7 drive.
//
// Neither takes an argument — both are called with a `nop` in the delay slot and no
// register set up before them — and they bracket a module's *initialisation*, which is
// what distinguishes them from the suspend/resume pair. 989SND calls #8 as the first
// thing it does; THREADMAN calls #9 as the last instruction before its entry point
// returns.
//
// THREADMAN is the argument. Its entry point opens by calling CpuSuspendIntr, and it
// never calls CpuResumeIntr — not once, anywhere. If #9 is not the thing that turns the
// processor's interrupts back on, then nothing is, and the IOP runs the rest of its life
// with them off: every module's Suspend faithfully saves "disabled", every matching Resume
// faithfully restores it, and the machine looks entirely healthy while the sound chip's
// completed transfer waits at a masked door forever. Which is exactly what it did.
func (p *IOP) intrCpuDisable() {
	p.intrEnabled = false
	p.ieEvent("cpu-off", 0, 0, p.CPU.Reg(31))
	p.setRet(0)
}

func (p *IOP) intrCpuEnable() {
	p.intrEnabled = true
	p.ieEvent("cpu-on", 0, 0, p.CPU.Reg(31))
	p.setRet(0)
}

// intrSetSwitchHook and intrSetReschedHook record THREADMAN's two hooks into the
// interrupt-exit path. The contract they describe is kept in intrDeliver: run the
// handler, ask the predicate, and if it says yes, call the switcher.
func (p *IOP) intrSetSwitchHook() {
	p.schedSwitch = p.arg(0)
	p.ps2.note("IOP: the scheduler's switch routine is %s", p.Sym(p.schedSwitch))
	p.setRet(0)
}

func (p *IOP) intrSetReschedHook() {
	p.schedResched = p.arg(0)
	p.ps2.note("IOP: the scheduler's reschedule predicate is %s", p.Sym(p.schedResched))

	p.deriveSchedIsRun()
	p.setRet(0)
}

// deriveSchedIsRun works out the address of the scheduler's "is running" pointer from the
// reschedule predicate's own code, so the is-running write log (and the thread inspector)
// need no hardcoded address. The predicate's first two instructions form the "ought to run"
// pointer — lui HI; addiu LO — and the predicate reads 0($v1) for ought-to-run and -4($v1)
// for is-running, so the is-running pointer is that address minus four. Called both when
// the hook registers on a fresh boot and after a resume, where the hook does not re-run.
func (p *IOP) deriveSchedIsRun() {
	if p.schedResched == 0 {
		return
	}
	lui, addiu := p.Read32(p.schedResched), p.Read32(p.schedResched+4)
	if lui>>26 == 0x0F && addiu>>26 == 0x09 {
		g := (lui&0xFFFF)<<16 + uint32(int32(int16(uint16(addiu&0xFFFF))))
		p.schedIsRun = g - 4
	}
}

// intrQueryContext is QueryIntrContext(): non-zero while an interrupt handler is running.
func (p *IOP) intrQueryContext() { p.setRet(b2u(p.inIntr > 0)) }

// intrInvokeInKmode is #14: run this function, and give me back what it returns.
//
// The first argument is a function pointer and the rest are its arguments, and that is the
// whole shape of it — THREADMAN wraps it in a three-line exported routine that passes one
// of its own functions and the caller's value, and the function it passes goes straight on
// to call a timer routine through timrman. So the name is about *where* the call runs, and
// the answer here is that it does not matter: this machine has no privilege levels to cross
// and no kernel mode to enter, so an invocation in kernel mode is an invocation.
//
// What matters is that it happens. Answering zero and calling nothing — which is what an
// unmodelled function does, and what this one did — quietly deletes whatever the caller
// asked to have run. Here it deleted the routine that arms THREADMAN's timer, and the
// second processor lost its heartbeat: the scheduler ran out of ready threads, fell into
// its idle loop, and waited for a clock that nobody had started.
func (p *IOP) intrInvokeInKmode() {
	fn := p.arg(0)
	if fn == 0 {
		p.setRet(0)
		return
	}
	a1, a2, a3 := p.arg(1), p.arg(2), p.arg(3)

	p.CPU.SetReg(4, a1)
	p.CPU.SetReg(5, a2)
	p.CPU.SetReg(6, a3)
	p.CPU.SetReg(7, 0)

	// On the caller's own stack, below its frame: this is a call the caller made, and it
	// returns to the caller. It is not an interrupt and it does not get a stack of its own.
	res, err := p.callGuestOn(fn, (p.CPU.Reg(29)-64)&^7)
	if err != nil {
		p.halt("the routine %s, invoked through intrman #14, did not return: %v", p.Sym(fn), err)
		return
	}
	p.setRet(res)
}

// --- delivering an interrupt ---------------------------------------------------------

// iopIntrStack is where an interrupt handler runs.
//
// It cannot be the stack callGuest uses, and that is not a detail. An interrupt arrives
// in the middle of whatever was running, so a handler starting at the top of the shared
// call stack would lay its own frame straight over the frames of the routine it
// interrupted — and the routine it interrupts is, almost always, the one waiting for it.
const iopIntrStack = iopRAMSizeBytes - 0x8400

// raiseIRQ marks an interrupt line. It is the hardware's I_STAT, and it is a Go field
// rather than a register because intrman is ours: no module on this disc reads the
// interrupt controller's registers, because none of them has to.
func (p *IOP) raiseIRQ(irq uint32) {
	p.pending |= 1 << uint64(irq)
	p.raised[irq]++
}

// serviceIntr delivers one pending interrupt, if the processor will take it.
//
// A line whose handler is not enabled stays raised. That is what the hardware does, and
// it is also the honest thing: the interrupt has happened whether or not anyone is
// listening yet, and a module that enables its handler afterwards should see it.
func (p *IOP) serviceIntr() {
	if p.inIntr > 0 || !p.intrEnabled || (p.pending == 0 && !p.vblankPending) {
		return
	}

	// Not in the shadow of a branch, and not with a load still in the air.
	//
	// The frame a context is saved into holds one program counter, and one program counter
	// cannot describe a processor that is about to execute a delay slot — the branch it
	// belongs to has already gone, and the address it is going to is nowhere in the frame.
	// The same is true of the R3000A's load delay: the value is in flight, not in the
	// register file, and the frame has no room for it either.
	//
	// Both resolve in a single instruction, so declining here costs the interrupt one step
	// of latency and buys the guarantee that every context this machine ever saves is one
	// it can restore exactly. It is the same guarantee the hardware gives itself with the
	// BD bit, arrived at from the other side.
	if st := p.CPU.SaveState(); st.PendingDelay || st.LdReg != 0 {
		return
	}

	// The vertical blank goes first: it is line 0, the highest-priority line on the
	// controller, and its dispatcher is ours (iopvblank.go) rather than a registered
	// guest routine.
	if p.vblankPending {
		p.vblankPending = false
		p.vblankDeliver()
		return
	}

	for irq := uint32(0); irq < iopIRQs; irq++ {
		bit := uint64(1) << uint64(irq)
		if p.pending&bit == 0 || p.imask&bit == 0 || p.handlers[irq].fn == 0 {
			continue
		}
		p.pending &^= bit
		p.delivered[irq]++
		p.intrDeliver(irq)
		return
	}
}

// intrDeliver is the exception path: save the interrupted thread into a frame on its own
// stack, run the handler, ask THREADMAN whether the interrupt has made a different thread
// the one that ought to be running, and resume whichever thread it names.
//
// The frame is not a Go struct, and that is the whole point. It lives in IOP memory, in
// the layout THREADMAN builds for a thread it starts (iopframe.go), because THREADMAN is
// going to be handed a pointer to it and is going to file it in the outgoing thread's
// control block. A context saved anywhere else would be a context Sony's scheduler could
// not put away — which is why the hooks sat recorded and uncalled for so long.
func (p *IOP) intrDeliver(irq uint32) {
	h := p.handlers[irq]

	// The frame goes below the interrupted thread's stack pointer, which is where the
	// kernel's exception prologue puts it and where THREADMAN's stack-overflow check
	// expects to find it.
	frame := (p.CPU.Reg(29) - iopFrameSize) &^ 7
	p.saveFrame(frame)

	// Whether the interrupted context is a thread at all.
	//
	// A module's entry point is not one. loadcore calls it, and here "loadcore" is Go, so
	// the entry runs as a nested interpreter loop on a stack of the machine's own —
	// whereas on the board it would be running on whichever thread called LoadModule, and
	// THREADMAN would know all about it. That difference is invisible until the scheduler
	// tries to *put the context away*: the switch routine files the frame in the control
	// block of the thread it believes is current, and the thread it believes is current is
	// not the one that is actually running. The frame goes to the wrong owner, the entry
	// point is never resumed, and the machine wanders into the low 64 KiB and executes
	// zeroes.
	//
	// So while a module is starting, the interrupt is delivered — the handler runs, the
	// DMA completion is seen, the flag the entry is spinning on gets set — but the thread
	// switch is declined. Nothing is lost by declining it: the interrupt has already made
	// its thread ready, and THREADMAN will pick it up at the next scheduling point, which
	// is the moment the entry point returns and the machine goes back to stepping threads.
	//
	// The honest fix is to give the loader a thread of its own, so there is always a real
	// context to file. That is a change to how modules are started, not to how interrupts
	// are delivered, and it belongs with the work that turns the IOP on in the main boot.
	preemptible := p.callDepth == 0

	p.inIntr++
	p.intrEnabled = false // a handler runs masked, as on the hardware
	p.ieEvent("deliver", frame, irq, p.CPU.Reg(31))

	p.CPU.SetReg(4, h.arg) // the handler's one argument, given at registration
	_, err := p.callGuestOn(h.fn, iopIntrStack)

	resume := frame
	if err == nil && preemptible {
		if next, ok := p.intrReschedule(frame); ok {
			resume = next
		}
	}

	p.inIntr--

	// The exception return. It restores the registers, the program counter *and* the
	// interrupt-enable, all three from the frame — which after a switch is a different thread's
	// frame, and that is the point.
	//
	// The enable used to be restored here instead, from a variable saved on the way in, and the
	// comment explained that the handler and the scheduler hooks are kernel code entitled to
	// leave the processor masked, so the restore had to be written down because there was no
	// status register to take it from. There is one: it is in the frame, at offset 136, and
	// THREADMAN has been putting the right value in it all along. Restoring from the variable
	// gives the *incoming* thread the *outgoing* one's interrupt state, which is right only
	// because it is usually the same thread — and wrong, silently and permanently, on exactly
	// the switches that are not preemptions.
	p.loadFrame(resume)

	if err != nil {
		p.halt("the handler for interrupt %d (%s) did not return: %v", irq, p.Sym(h.fn), err)
	}
}

// intrReschedule asks THREADMAN whether to switch threads, and switches if it says so.
// It returns the frame the processor should resume from.
//
// The predicate and the switch routine are THREADMAN's own, registered through intrman
// #30 and #28. The kernel does not need to understand the scheduler — it only has to ask
// the question the scheduler asked to be asked, and to believe the answer.
func (p *IOP) intrReschedule(frame uint32) (uint32, bool) {
	if p.schedResched == 0 || p.schedSwitch == 0 {
		return 0, false
	}
	want, err := p.callGuestOn(p.schedResched, iopIntrStack)
	if err != nil || want == 0 {
		return 0, false
	}

	// Hand the switch routine the frame we saved. It files it in the outgoing thread's
	// control block, makes the incoming thread current, and returns *its* frame.
	p.CPU.SetReg(4, frame)
	next, err := p.callGuestOn(p.schedSwitch, iopIntrStack)
	if err != nil {
		p.halt("THREADMAN's thread switch did not return: %v", err)
		return 0, false
	}
	if next == 0 {
		return 0, false
	}
	p.switches++
	return next, true
}

// tick advances the second processor's own clocks by one instruction: it finishes any
// DMA whose latency has run out, and delivers any interrupt that has become deliverable.
// Every loop that steps the IOP goes through it.
func (p *IOP) tick() {
	p.ioTraceFlush()
	if p.timerAck != 0 {
		p.timerAckFlush()
	}
	p.steps++

	// The dead zone. Below the first module is the kernel's own memory, and on this machine
	// it is empty — which is worse than it sounds, because a word of zero is a `nop` on
	// MIPS. A processor that jumps to null does not fault here; it nops its way up through
	// sixty-four kilobytes of nothing and falls into the first module's code sideways, and
	// the boot goes on being wrong for a very long time. Catching the jump is the only way
	// to hear about it, and the address it came *from* is the only thing worth knowing.
	if pc := iopPhys(p.CPU.PC); pc < iopModuleBase && pc != iopIdleLoop && pc != iopIdleLoop+4 {
		p.halt("jumped into the empty kernel area at 0x%08X, from %s\n%s", p.CPU.PC, p.Sym(p.lastPC), p.IOPTrail())
		return
	}
	p.lastPC = p.CPU.PC
	p.trail[p.trailN%iopTrailLen] = p.CPU.PC
	p.trailN++

	if p.Trap != 0 && p.CPU.PC == p.Trap {
		p.halt("reached the trap at %s\n%s", p.Sym(p.Trap), p.IOPTrail())
		return
	}
	// Every call through an import stub, as it is made. A stub is the one place the caller
	// and the callee are both named — the address is in the middle of somebody's .text, but
	// the linker knows which (library, function) it patched there.
	if p.OnCall != nil {
		if name, ok := p.stubName[iopPhys(p.CPU.PC)]; ok {
			p.OnCall(name,
				[4]uint32{p.CPU.Reg(4), p.CPU.Reg(5), p.CPU.Reg(6), p.CPU.Reg(7)},
				p.CPU.Reg(31))
		}
	}
	if p.steps%iopProfileEvery == 0 {
		if p.prof == nil {
			p.prof = map[string]int{}
		}
		p.prof[p.symFunc(p.CPU.PC)]++
	}
	p.timerTick()
	if p.cdvd.nBusy {
		p.cdvd.tick(p)
	}
	if len(p.dmaPending) > 0 {
		p.dmaTick()
	}
	if p.pending != 0 {
		p.serviceIntr()
	}
}

func b2u(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// --- sysmem -----------------------------------------------------------------------

// iopBlock is one allocation, remembered so that the two functions that ask about a
// block — its base and its size — can answer, and so that freeing one returns its space.
type iopBlock struct {
	base, size uint32
}

// sysmemAlloc is AllocSysMemory(mode, size, addr).
//
// It has a free list now, and it must. For the boot it did not: the twelve kernel modules
// are loaded once and kept, so a bump allocator that never reclaims is a set of addresses
// that never move, which during bring-up is worth more than the memory. But the game loads
// its own modules at runtime, and it does it by reading each file into a buffer, handing the
// buffer to MODLOAD, and *freeing it* — so a machine that never frees leaks one buffer per
// module, and runs out of its two megabytes somewhere around OVERLORD with every resident
// module still comfortably fitting. The leaked buffers, not the modules, are what overflow.
//
// So a freed block goes on a list and the next allocation that fits is served from it. New
// space is still bumped when nothing on the list will do. The mode chooses a free-list
// strategy on the real allocator; here best-fit stands in for all of it, and an explicit
// address is still honoured because a caller that named one will use it whatever we say.
// The allocation modes AllocSysMemory is called with on this disc, read off the callers:
//
//	0  from the LOW end. The modules (MODLOAD+0xFA4), and the big persistent buffers a
//	   module allocates at startup — OVERLORD's 811 KiB ramdisk (InitRamdisk).
//	1  from the HIGH end. The file MODLOAD reads each module out of (MODLOAD+0xDF0), and
//	   every thread's stack (THREADMAN+0xD10).
//	2  at an explicit address (the `addr` argument).
//
// The two ends are the whole point, and this disc proves why. The file buffers are allocated,
// used and freed once per module; the modules are allocated and kept. Grow both from the same
// end and a freed file buffer leaves a hole no later module is the right size to fill, and the
// two megabytes are exhausted with everything that is meant to stay resident still fitting
// twice over. Grow the transient allocations down from the top and the permanent ones up from
// the bottom, and the transient end reclaims itself completely — which is exactly what a real
// IOP's sysmem does, and why the mode argument exists.
const (
	iopAllocLow  = 0
	iopAllocHigh = 1
	iopAllocAddr = 2
)

func (p *IOP) sysmemAlloc() {
	mode, size, addr := p.arg(0), p.arg(1), p.arg(2)
	size = (size + 63) &^ 63 // the allocator's granularity, as p.alloc's is

	if mode == iopAllocAddr && addr != 0 {
		p.blocks = append(p.blocks, iopBlock{base: addr, size: size})
		p.setRet(addr)
		return
	}

	base := p.allocReuse(size, mode == iopAllocHigh)
	if base == 0 {
		if mode == iopAllocHigh {
			base = p.allocHigh(size)
		} else {
			base = p.alloc(size)
		}
	}
	if base == 0 {
		var free uint32
		for _, f := range p.freeBlocks {
			free += f.size
		}
		p.ps2.note("IOP: AllocSysMemory(mode=%d) could not find %d bytes; low=0x%X high=0x%X free-list=%d bytes in %d blocks",
			mode, size, p.allocPtr, p.allocHighPtr, free, len(p.freeBlocks))
		p.setRet(0)
		return
	}
	p.blocks = append(p.blocks, iopBlock{base: base, size: size})
	p.setRet(base)
}

// allocHigh bumps a block off the top of the arena, growing down. It is the mirror of
// p.alloc, and the two must not cross: the low allocations grow up and the high ones grow
// down, and when they meet the machine is out of memory.
func (p *IOP) allocHigh(size uint32) uint32 {
	top := p.allocHighPtr
	if top == 0 {
		top = iopStackArea
	}
	base := (top - size) &^ 63
	if base < p.allocPtr {
		p.halt("out of IOP memory: %d bytes wanted from the high end, and it has reached the low allocations at 0x%08X",
			size, p.allocPtr)
		return 0
	}
	p.allocHighPtr = base
	return base
}

// allocReuse serves size from the smallest free block that will hold it, splitting off the
// remainder. Best-fit, and from the end the caller asked for: a high request takes the top of
// its chosen hole and a low request the bottom, so a reused block stays on the side of memory
// its neighbours are on and the two ends do not interleave through the free list.
func (p *IOP) allocReuse(size uint32, high bool) uint32 {
	best := -1
	for i, f := range p.freeBlocks {
		if f.size >= size && (best < 0 || f.size < p.freeBlocks[best].size) {
			best = i
		}
	}
	if best < 0 {
		return 0
	}
	f := p.freeBlocks[best]
	if f.size == size {
		p.freeBlocks = append(p.freeBlocks[:best], p.freeBlocks[best+1:]...)
		return f.base
	}
	if high {
		p.freeBlocks[best] = iopBlock{base: f.base, size: f.size - size}
		return f.base + f.size - size
	}
	p.freeBlocks[best] = iopBlock{base: f.base + size, size: f.size - size}
	return f.base
}

// sysmemFree is FreeSysMemory(ptr): return a block's space to the free list, coalescing it
// with any neighbour so that a run of freed load buffers becomes one hole big enough for the
// next module rather than a scatter of small ones.
func (p *IOP) sysmemFree() {
	ptr := p.arg(0)
	for i, b := range p.blocks {
		if b.base == ptr {
			p.blocks = append(p.blocks[:i], p.blocks[i+1:]...)
			p.freeInsert(b)
			p.setRet(0)
			return
		}
	}
	// A free of something we never handed out. Harmless, and worth a note rather than a
	// crash, because it means either a double free or a pointer we lost track of.
	p.setRet(0)
}

// freeInsert adds a block to the free list, merges it with any adjacent free block, and — if
// the result sits at either bump frontier — gives the space back to the bump pointer.
//
// The giving-back is what makes the two ends work under churn. The game loads a dozen modules,
// each read into a high-end buffer that is freed the moment the module is placed; without
// retraction those freed buffers pile up on the free list in the wrong sizes to be reused,
// and the high pointer never recovers, so it meets the rising low pointer with hundreds of
// kilobytes of dead holes between them. Retract the high pointer past a freed block on its
// frontier and the buffer space is genuinely returned, which is the difference between
// OVERLORD's 811 KiB ramdisk fitting and the machine reporting itself full with room to spare.
func (p *IOP) freeInsert(b iopBlock) {
	for {
		merged := false
		for i, f := range p.freeBlocks {
			if f.base+f.size == b.base {
				b.base, b.size = f.base, f.size+b.size
			} else if b.base+b.size == f.base {
				b.size += f.size
			} else {
				continue
			}
			p.freeBlocks = append(p.freeBlocks[:i], p.freeBlocks[i+1:]...)
			merged = true
			break
		}
		if !merged {
			break
		}
	}

	// At the low frontier: the block ends where the low bump pointer is, so retract it.
	if b.base+b.size == p.allocPtr {
		p.allocPtr = b.base
		return
	}
	// At the high frontier: the block begins where the high bump pointer is, so retract it.
	if p.allocHighPtr != 0 && b.base == p.allocHighPtr {
		p.allocHighPtr = b.base + b.size
		return
	}
	p.freeBlocks = append(p.freeBlocks, b)
}

// sysmemBlockTop and sysmemBlockSize answer for the block containing a pointer.
//
// THREADMAN calls them as a pair and stores the results as a thread's stack base and
// stack size — the same two slots it fills from AllocSysMemory's return and its size
// argument on the path where it allocates the stack itself. That is what says which of
// the two is which.
func (p *IOP) sysmemBlockTop() {
	if b, ok := p.blockOf(p.arg(0)); ok {
		p.setRet(b.base)
		return
	}
	p.setRet(0)
}

func (p *IOP) sysmemBlockSize() {
	if b, ok := p.blockOf(p.arg(0)); ok {
		p.setRet(b.size)
		return
	}
	p.setRet(0)
}

func (p *IOP) blockOf(ptr uint32) (iopBlock, bool) {
	for _, b := range p.blocks {
		if ptr >= b.base && ptr < b.base+b.size {
			return b, true
		}
	}
	return iopBlock{}, false
}

// sysmemKprintf is the kernel's printf, and the IOP's other voice.
func (p *IOP) sysmemKprintf() {
	s := p.formatArgs(p.CString(p.arg(0)), 1)
	p.ps2.iopPrint(s)
	p.setRet(uint32(len(s)))
}

// --- heaplib ----------------------------------------------------------------------

// iopHeap is one heap: a chunk taken from the allocator and handed out of, and grown
// with another chunk when it runs out.
//
// The growth is the whole point, and it was learned the hard way. THREADMAN creates its
// heap with a *chunk* size, not a total — 2 KiB — and then takes every thread control
// block, every semaphore and every event flag in the machine out of it. A heap that
// stops at its first chunk runs dry about the time OVERLORD starts up, and what you see
// then is not "out of memory": it is CreateSema quietly returning an error code, and
// OVERLORD, which checks, jumping to itself forever. Two instructions, no message, in a
// module whose sound initialisation is the last thing it printed.
type iopHeap struct {
	chunk uint32 // how much to take from sysmem each time it grows
	base  uint32 // the current chunk
	size  uint32
	ptr   uint32 // the bump pointer within it
	total uint32 // handed out, over the heap's life
}

// heapCreate is CreateHeap(size, flags). It returns the heap's address, which is also
// the handle every later call passes back.
func (p *IOP) heapCreate() {
	chunk := p.arg(0)
	base := p.alloc(chunk)
	if base == 0 {
		p.setRet(0)
		return
	}
	p.heaps[base] = &iopHeap{chunk: chunk, base: base, size: chunk, ptr: base}
	p.setRet(base)
}

// heapAlloc is AllocHeapMemory(heap, size). The caller clears what it gets, so this
// does not have to.
func (p *IOP) heapAlloc() {
	h, size := p.heaps[p.arg(0)], p.arg(1)
	if h == nil {
		// A heap we never made. THREADMAN takes every thread block out of one, so a null
		// here would be a thread that silently does not exist — say so instead.
		p.halt("AllocHeapMemory from heap 0x%08X, which was never created", p.arg(0))
		p.setRet(0)
		return
	}

	a := (h.ptr + 7) &^ 7
	if a+size > h.base+h.size {
		// Grow it. A fresh chunk, at least as big as the request.
		n := h.chunk
		if size > n {
			n = size
		}
		base := p.alloc(n)
		if base == 0 {
			p.setRet(0)
			return
		}
		h.base, h.size, h.ptr = base, n, base
		a = base
	}
	h.ptr = a + size
	h.total += size
	p.setRet(a)
}

// --- sysclib ----------------------------------------------------------------------

// clibMemcpy is memcpy(dst, src, n), and it is the function that was opening a file called "".
//
// The argument is FILEIO's, and it is the cleanest one on this processor. FILEIO's remote-call
// handler is passed the buffer the EE has just DMA'd its request into, and the first thing it
// does with it is
//
//	sysclib#12(0x53340, 0x4CFB8, 0x20)
//
// where 0x4CFB8 is that buffer, 0x20 is exactly the number of bytes the EE transferred into it,
// and 0x53340 is a block FILEIO allocated for itself. It then opens the path at 0x53354 —
// twenty bytes into the destination, which is exactly where the path sits in the source. Three
// arguments, a destination, a source, and a length that is the length of the thing at the
// source: there is no other function this can be, and memset is already #14, which is where a
// C library keeps its neighbour.
//
// Answered with zero, it copied nothing, and FILEIO opened the empty string it found in a
// buffer nobody had written. The failure surfaced four layers away and looked nothing like a
// missing memcpy: the EE's threads sat on a semaphore waiting for a file operation that had
// been dispatched correctly, executed faithfully, and asked for a file whose name was "".
func (p *IOP) clibMemcpy() {
	dst, src, n := p.arg(0), p.arg(1), p.arg(2)
	for i := uint32(0); i < n; i++ {
		p.Write(dst+i, p.Read(src+i))
	}
	p.setRet(dst)
}

// clibMemmove is memmove(dst, src, n) — sysclib's copy that lives one index above memcpy (#12),
// which is where a C library keeps memmove. Both of its call sites are in 989snd's sound-bank
// block reader: the header pull `#13(sp+24, ring, 0x20)` (32 bytes from the streaming ring
// buffer into a stack header) and the body copy that walks the ring in <=2048-byte chunks —
// (dst, src, len) each time. Left unmodelled it copied nothing, so the block header the parser
// read back was uninitialised stack; its type word was neither of the 1/3 it accepts, and 989snd
// printed "cause 84" and retried, forever — a busy-wait that pinned the IOP in CheckDiscID and
// starved the RPC-server threads the EE was waiting on. Copied overlap-safe because a ring the
// reader also advances through can hand back a source below the destination.
func (p *IOP) clibMemmove() {
	dst, src, n := p.arg(0), p.arg(1), p.arg(2)
	if dst > src && dst < src+n {
		for i := n; i > 0; i-- {
			p.Write(dst+i-1, p.Read(src+i-1))
		}
	} else {
		for i := uint32(0); i < n; i++ {
			p.Write(dst+i, p.Read(src+i))
		}
	}
	p.setRet(dst)
}

// clibBzero is bzero(dst, n) — memset's two-argument cousin, and the one IOMAN clears its
// driver table with.
func (p *IOP) clibBzero() {
	dst, n := p.arg(0), p.arg(1)
	for i := uint32(0); i < n; i++ {
		p.Write(dst+i, 0)
	}
	p.setRet(dst)
}

// clibStrcmp is strcmp(a, b): zero when they are the same, and IOMAN takes zero as the match
// that ends its walk of the driver list.
func (p *IOP) clibStrcmp() {
	a, b := p.arg(0), p.arg(1)
	for i := uint32(0); ; i++ {
		ca, cb := p.Read(a+i), p.Read(b+i)
		if ca != cb {
			p.setRet(uint32(int32(ca) - int32(cb)))
			return
		}
		if ca == 0 {
			p.setRet(0)
			return
		}
	}
}

// clibStrncpy is strncpy(dst, src, n). It does not terminate the copy when it fills it, and
// IOMAN relies on that: it writes the NUL itself, at dst[n], in the call's delay slot.
func (p *IOP) clibStrncpy() {
	dst, src, n := p.arg(0), p.arg(1), p.arg(2)
	for i := uint32(0); i < n; i++ {
		c := p.Read(src + i)
		p.Write(dst+i, c)
		if c == 0 {
			for ; i < n; i++ { // strncpy pads the rest with NULs
				p.Write(dst+i, 0)
			}
			break
		}
	}
	p.setRet(dst)
}

// iopCtypeDigit is the one bit of the character-class table that has been earned: IOMAN tests
// it on the last character of a device name to find the unit number. See the note above the
// sysclib table — nothing else on this disc asks this function anything.
const iopCtypeDigit = 0x04

// clibCtype is the character-class lookup.
func (p *IOP) clibCtype() {
	c := byte(p.arg(0))
	var class uint32
	if c >= '0' && c <= '9' {
		class |= iopCtypeDigit
	}
	p.setRet(class)
}

// clibStrncmp is strncmp(a, b, n), and it is CDVDMAN's check that the disc has a filesystem
// on it: #29(sector + 1, "CD001", 5), against the volume descriptor it has just read out of
// LBA 16 (CDVDMAN+0x6520).
//
// Answered with zero it did not fail that check — it PASSED it, unconditionally, because zero
// is what "these are the same" means. CDVDMAN went on to read the path table's address out of
// a buffer nobody had checked, got 48, read sector 48, and reported that it could not find the
// file. The reads it made were the only sign, and they were three plausible-looking numbers.
//
// Whether it is strncmp or memcmp cannot be told from this call — five bytes of text with no
// NUL in them, and the two agree. It stops at a NUL, which is the reading that also puts it
// next to strcmp at #22 and strncpy at #30 in a table that is otherwise in C-library order.
func (p *IOP) clibStrncmp() {
	a, b, n := p.arg(0), p.arg(1), p.arg(2)
	for i := uint32(0); i < n; i++ {
		ca, cb := p.Read(a+i), p.Read(b+i)
		if ca != cb {
			p.setRet(uint32(int32(ca) - int32(cb)))
			return
		}
		if ca == 0 {
			break
		}
	}
	p.setRet(0)
}

// clibStrlen is strlen(s).
func (p *IOP) clibStrlen() {
	s := p.arg(0)
	n := uint32(0)
	for ; n < iopCStringMax && p.Read(s+n) != 0; n++ {
	}
	p.setRet(n)
}

// clibStrcat is strcat(dst, src).
func (p *IOP) clibStrcat() {
	dst, src := p.arg(0), p.arg(1)
	end := dst
	for ; end-dst < iopCStringMax && p.Read(end) != 0; end++ {
	}
	for i := uint32(0); i < iopCStringMax; i++ {
		c := p.Read(src + i)
		p.Write(end+i, c)
		if c == 0 {
			break
		}
	}
	p.setRet(dst)
}

// clibStrcpy is strcpy(dst, src). It sits at #23, one below strchr (#25) and above the block
// of comparison/copy routines, in a table that is otherwise in C-library order.
//
// This is what makes the sound-bank NAME arrive. OVERLORD's LoadSoundBank (0xA72F8) copies the
// bank name out of the EE's RPC request (record+16 = "common") into the ISO command's inline
// name field (CMD+40) with #23; FS_LoadSoundBank then reads CMD+40, appends ".sbk" and FS_Finds
// it. Left unmodelled, CMD+40 stayed empty, FS_Find failed, and the game fell back to loading
// "empty1" into bank slot 0 — a slot whose header-skip is 10, which walks four sectors into VAG
// data and makes 989snd's block parser SndError(84) for ever. A no-op copy here, four calls and
// two modules away, presented as an interrupt storm that never resolved. The earlier reading —
// that strcpy "copies from the already-empty a0" and so cannot matter — looked only at
// FS_LoadSoundBank's own downstream strcpy and missed this upstream one. See
// [[hle-fictions-become-evidence]]. (Modelling it clears the SndError-84 hog; the real "common"
// bank then loads and the boot advances to a deeper, pre-existing scheduler bug — a thread
// busy-waits on the SPU-DMA completion interrupt with global interrupts disabled.)
func (p *IOP) clibStrcpy() {
	dst, src := p.arg(0), p.arg(1)
	for i := uint32(0); i < iopCStringMax; i++ {
		c := p.Read(src + i)
		p.Write(dst+i, c)
		if c == 0 {
			break
		}
	}
	p.setRet(dst)
}

// iopCStringMax bounds the C library's walks. A string with no terminator is a bug in the
// guest or in us, and either way an unbounded loop over 2 MiB of IOP memory is a hang rather
// than a diagnosis.
const iopCStringMax = 1024

// clibStrchr is strchr(s, c): a pointer to the first c in s, or null. IOMAN asks it for the
// colon, and a null answer is what "Unknown device" means.
func (p *IOP) clibStrchr() {
	s, c := p.arg(0), byte(p.arg(1))
	for i := uint32(0); i < 1024; i++ {
		ch := p.Read(s + i)
		if ch == c {
			p.setRet(s + i)
			return
		}
		if ch == 0 {
			break
		}
	}
	p.setRet(0)
}

// clibMemset is memset(dst, c, n).
func (p *IOP) clibMemset() {
	dst, c, n := p.arg(0), p.arg(1), p.arg(2)
	for i := uint32(0); i < n; i++ {
		p.Write(dst+i, byte(c))
	}
	p.setRet(dst)
}

// --- the kernel's own syscall --------------------------------------------------------

// iopSyscallReschedule is the service number THREADMAN asks the kernel for, and the only
// one anything on this disc asks for at all.
//
// It is the missing half of the scheduler, and its absence was invisible for a long time
// because of an accident of the instruction set. THREADMAN has exactly one `syscall` in
// twenty-eight kilobytes of code:
//
//	THREADMAN+0x46C   addiu $v0, $zero, 32
//	THREADMAN+0x470   syscall
//	THREADMAN+0x474   jr $ra
//
// — a three-instruction exported function that takes no arguments, returns nothing, and
// asks the kernel for service 32. The interrupt-exit path already had the machinery this
// needs: a predicate that says whether a switch is wanted and a routine that performs one,
// both of them THREADMAN's own, registered through intrman (iopframe.go). What was missing
// was the *other* door into it — the one a thread uses when it gives up the processor of
// its own accord rather than being taken off it. A thread that blocks on a semaphore does
// not wait for a timer; it asks to be rescheduled now.
//
// With no exception handler at 0x80000080 — and there is none, because intrman is Go — the
// core took the exception into empty memory. And a word of zero is a `nop` on MIPS, so the
// processor did not fault. It nopped its way up through sixty-four kilobytes of nothing and
// fell sideways into the first module's code, and the boot went on, wrongly, for a very
// long time. It took putting an idle loop in that empty memory to hear about it at all.
const iopSyscallReschedule = 32

// kernelSyscall serves a `syscall` whose service number is in $v0.
func (p *IOP) kernelSyscall() bool {
	switch svc := p.CPU.Reg(2); svc {
	case iopSyscallReschedule:
		p.yield()

	default:
		// Counted, not guessed at, and above all not taken as an exception into memory that
		// does not fault. The census is the work list.
		name := fmt.Sprintf("kernel syscall %d", svc)
		p.unmodelledCalls[name]++
		if p.unmodelledCalls[name] == 1 {
			p.ps2.note("IOP: %s from %s — unmodelled", name, p.Sym(p.CPU.CurPC()))
		}
	}
	return true
}

// yield is a thread giving up the processor: the synchronous twin of intrDeliver's exit
// path, and the same steps without the handler in front of them.
//
// The program counter saved is the instruction *after* the syscall, which is where the core
// has already advanced to — so a thread resumed from this frame comes back to the `jr $ra`
// and returns to whoever asked to be rescheduled, with no idea it was ever away.
//
// The service takes an argument, and it is the one this machine deadlocked without: $a2 is
// the interrupt-enable state the calling thread is to RESUME with. The contract is written
// all over THREADMAN. Every one of its blocking primitives opens with CpuSuspendIntr(&old)
// and then leaves one of two ways — the paths that keep the thread running end in
// CpuResumeIntr(old), and the paths that give the processor up end in the reschedule leaf
// (THREADMAN+0x46C: $v0=32, syscall) with *that same old* loaded into $a2 (sixteen sites
// load it straight from the CpuSuspendIntr slot; the wake-a-better-thread path at
// THREADMAN+0x5D0 is the proof in one screenful, passing one value to CpuResumeIntr on its
// no-switch exit and to $a2 on its switch exit). The restore is DELEGATED to the kernel:
// the thread sleeps inside its critical section, and the kernel dissolves the section on
// its behalf, filing the frame with SR's IEp = $a2 so the eventual exception return brings
// interrupts back to what they were before the Suspend. (The one caller that passes 0 is
// THREADMAN's stack-overflow trap — CpuDisableIntr, Kprintf, park forever — a thread that
// is meant to stay masked.)
//
// Saving the live intrEnabled instead — which is what this did — parked every blocked
// thread with the enable OFF, since a blocking thread is always inside a Suspend when it
// yields. Mostly that healed itself: the woken thread soon blocked again and the next
// thread's frame carried a 1. It became a deadlock the day a woken thread BUSY-WAITED on
// an interrupt-fed flag — 989snd spinning on its SPU-DMA completion with the completion
// pending, unmasked, and undeliverable behind the very bit the kernel had been told to
// restore.
func (p *IOP) yield() {
	frame := (p.CPU.Reg(29) - iopFrameSize) &^ 7
	p.saveFrame(frame)
	p.Write32(frame+iopFrameSR, b2u(p.CPU.Reg(6) != 0)<<2)

	// The wait-result pointer, and it is the other half of the reschedule contract.
	//
	// A blocking primitive that owes the sleeper a value when it wakes — WaitEventFlag,
	// WaitSema — passes the address to deposit it at in $a0 and then blocks through the
	// reschedule leaf. The waker looks for that address in the sleeping thread's frame at
	// +8: SetEventFlag does `lw v1, 8(frame); if v1: sw flagbits, 0(v1); sw zero, 8(frame)`.
	// So the syscall frame's +8 is not a saved register — it is the wait-result pointer,
	// and it differs from the interrupt frame there. Our saveFrame writes the *interrupt*
	// layout, which puts the caller's $v0 (= 32, the reschedule service number) in that slot,
	// and the waker was dutifully writing the answer to address 0x20 and leaving the sleeper
	// to read a stale word off its own stack. That is what killed the memory-card worker:
	// woken with a result of 0x1800 left over from a previous transfer wait, it matched none
	// of its request bits and ran off the end of its loop into ExitThread. $a0 is that
	// pointer (0 when the caller wants nothing back, which the waker treats as "skip").
	p.Write32(frame+iopFrameWaitResult, p.CPU.Reg(4))

	p.ieEvent("yield", frame, p.CPU.Reg(6), p.CPU.Reg(31))

	resume := frame
	if next, ok := p.intrReschedule(frame); ok {
		resume = next
	}
	// The exception return, from whichever frame won. When no switch was wanted this is the
	// frame just saved, and restoring it is not a no-op: it is the rfe that carries $a2 into
	// the caller's interrupt-enable, exactly as it would have been on the way out of the
	// syscall exception.
	p.loadFrame(resume)
}
