package ps2

// iophw_test.go checks the IOP's hardware — the DMA controller, the sound chip, the
// timers and the interrupt path — on the points where being wrong is quiet.
//
// Every test here is a bug that was actually made, and every one of them presented as the
// same symptom: a machine that ran, and ran, and never arrived. That is what makes them
// worth pinning. A processor that crashes tells you where it crashed; a processor that has
// silently masked an interrupt tells you nothing at all, for two hundred million
// instructions at a time.

import "testing"

func TestFreeingAHighBufferReturnsItsSpace(t *testing.T) {
	// The game loads a dozen modules by reading each into a high-end buffer and freeing it the
	// moment the module is placed low. If a freed buffer's space is not genuinely returned, the
	// buffers pile up on the free list in the wrong sizes to reuse, the high pointer never
	// recovers, and OVERLORD's 811 KiB ramdisk cannot find room in a two-megabyte machine with
	// half of it still free. The retraction in freeInsert is what makes the two ends work.
	m := NewMachine()
	p := m.StartIOP()

	high := func(size uint32) uint32 { // AllocSysMemory(mode=1)
		p.CPU.SetReg(4, iopAllocHigh)
		p.CPU.SetReg(5, size)
		p.CPU.SetReg(6, 0)
		p.sysmemAlloc()
		return p.CPU.Reg(2)
	}
	free := func(ptr uint32) {
		p.CPU.SetReg(4, ptr)
		p.sysmemFree()
	}

	before := p.allocHighPtr
	if before == 0 {
		before = iopStackArea
	}
	// A run of differently-sized buffers, each freed before the next — the module-load pattern.
	for _, n := range []uint32{6161, 43861, 26245, 139052, 124132} {
		b := high(n)
		if b == 0 {
			t.Fatalf("the high allocator could not place %d bytes", n)
		}
		free(b)
	}
	if p.allocHighPtr != before {
		t.Errorf("after freeing every high buffer, the high pointer is at 0x%X, not back at 0x%X: %d bytes stranded",
			p.allocHighPtr, before, before-p.allocHighPtr)
	}
	if len(p.freeBlocks) != 0 {
		t.Errorf("%d free block(s) stranded on the list after everything was freed", len(p.freeBlocks))
	}
}

func TestEachDMAChannelHasAnInterruptOfItsOwn(t *testing.T) {
	// intrman does not hand a driver the controller's single DMA line. It demuxes it, and
	// gives each channel an interrupt number of its own — 32 + n for the first block of
	// channels, 40 + (n-7) for the second. Four modules on this disc agree on it: LIBSD
	// registers on 36 and 40 for the two sound cores, SIFMAN on 42 and SIFCMD on 43 for the
	// two SIF channels.
	//
	// A vector table of 32 entries does not merely fail to deliver these. It refuses the
	// *registration*, returns an error the caller does not check, and leaves the sound
	// chip's completion interrupt with no handler at all — so the transfer completes, and
	// the callback that would have reported it never runs, and OVERLORD waits for a flag
	// that nothing will ever set.
	for _, c := range []struct {
		ch   int
		want uint32
		who  string
	}{
		{iopDMAChSPU0, 36, "LIBSD, sound core 0"},
		{iopDMAChSPU1, 40, "LIBSD, sound core 1"},
		{iopDMAChSIF0, 42, "SIFMAN, SIF0"},
		{iopDMAChSIF1, 43, "SIFCMD, SIF1"},
	} {
		if got := iopDMAIRQ(c.ch); got != c.want {
			t.Errorf("DMA channel %d raises interrupt %d, but %s registers its handler on %d",
				c.ch, got, c.who, c.want)
		}
		if c.want >= iopIRQs {
			t.Fatalf("interrupt %d is past the end of a %d-entry vector table: the registration "+
				"would be refused and the handler silently lost", c.want, iopIRQs)
		}
	}
}

func TestUnmaskingALineBeforeItsHandlerIsRegisteredSticks(t *testing.T) {
	// LIBSD calls EnableIntr(9) and only *then* RegisterIntrHandler(9). Nothing says it may
	// not: on the hardware the mask register and the vector table are two different pieces
	// of state, and a driver may write them in either order.
	//
	// Keep the mask bit inside the handler record and registering the handler stores a fresh
	// record over the top of the enable that already happened. The line is masked from that
	// moment on, and it is masked in a way no instrument shows, because a handler *is*
	// registered and the interrupt *is* raised.
	m := NewMachine()
	p := m.StartIOP()

	p.CPU.SetReg(4, 9) // EnableIntr(9)
	p.intrEnable()

	p.CPU.SetReg(4, 9)      // RegisterIntrHandler(9, mode, handler, arg)
	p.CPU.SetReg(5, 0)      //
	p.CPU.SetReg(6, 0x1234) //
	p.CPU.SetReg(7, 0)      //
	p.intrRegister()

	if p.imask&(1<<9) == 0 {
		t.Fatal("registering a handler cleared the enable that came before it: the line is masked")
	}
}

func TestASoundTransferMovesItsBytesAndReportsItself(t *testing.T) {
	// The whole chain OVERLORD waits on: the DMA controller takes a block out of IOP memory
	// and hands it to the sound chip, the chip says it has taken it, and the channel's
	// interrupt is raised so that LIBSD's handler can call the callback that sets the flag.
	//
	// Any one link missing looks identical from outside — DMA_SendToSPUAndSync spinning on
	// seven addresses — so the test walks all of it.
	m := NewMachine()
	p := m.StartIOP()
	p.running = true

	const (
		src  = 0x00040000 // where the data is in IOP memory
		tsa  = 0x8820     // where it goes in the sound chip, in the chip's own units
		size = 64
	)
	for i := uint32(0); i < size; i++ {
		p.Write(src+i, byte(i)+1)
	}

	// LIBSD's own sequence: enable the channel in DPCR, set the chip's transfer address,
	// then program the channel and set the start bit.
	p.ioWrite(iopDPCR, 0x000B8800)
	p.spu.setHalf(iopSPU2TSAHi, tsa>>16)
	p.spu.setHalf(iopSPU2TSALo, tsa&0xFFFF)

	base := uint32(iopDMA1Base + iopDMAChSPU0*0x10)
	p.ioWrite(base+iopDMAMadr, src)
	p.ioWrite(base+iopDMABcr, 0x0001_0010) // 0x10 words to a block, one block: 64 bytes
	p.ioWrite(base+iopDMAChcr, 0x0100_0201)

	// The bytes are in the sound memory the moment the transfer runs.
	for i := uint32(0); i < size; i++ {
		if got, want := p.spu.ram[tsa*2+i], byte(i)+1; got != want {
			t.Fatalf("sound memory at 0x%X holds 0x%02X, want 0x%02X", tsa*2+i, got, want)
		}
	}

	// But the channel is still busy, and the chip has not said it is done. A driver that
	// polls either of these must see the transfer in flight.
	if v, _ := p.dmaRead(base + iopDMAChcr); v&iopChcrStart == 0 {
		t.Error("the channel reports itself idle before its completion has been reported")
	}
	if p.raised[iopDMAIRQ(iopDMAChSPU0)] != 0 {
		t.Error("the completion interrupt was raised before the transfer's latency had elapsed")
	}

	// Let the latency run out.
	for i := 0; i < iopDMALatency+2; i++ {
		p.steps++
		p.dmaTick()
	}

	if v, _ := p.dmaRead(base + iopDMAChcr); v&iopChcrStart != 0 {
		t.Error("the channel is still busy after it completed")
	}
	if p.raised[iopDMAIRQ(iopDMAChSPU0)] != 1 {
		t.Fatalf("channel %d raised interrupt %d %d times, want once",
			iopDMAChSPU0, iopDMAIRQ(iopDMAChSPU0), p.raised[iopDMAIRQ(iopDMAChSPU0)])
	}

	// And the chip's own "I have taken it" bit, which LIBSD's handler sits in a timed loop
	// waiting for. Leave it clear and the driver spins out its timeout on every transfer.
	if p.spu.half(iopSPU2CoreStat)&iopSPU2CoreDone == 0 {
		t.Error("the sound chip never reported the transfer complete in its own status register")
	}
	if p.spu.half(iopSPU2Stat)&(1<<2) == 0 {
		t.Error("the chip-wide status does not say which core finished")
	}
}

func TestAByteStoreToAChannelMergesAgainstTheLiveRegister(t *testing.T) {
	// The R3000A's bus is byte-wide: a word store to a DMA register arrives four times, and
	// each byte has to be merged into the register's *current* value. Merge against a stale
	// copy kept somewhere else and three quarters of every command is lost — which for CHCR
	// means a transfer with the right start bit and the wrong direction.
	m := NewMachine()
	p := m.StartIOP()

	base := uint32(iopDMA1Base + iopDMAChSPU0*0x10)
	p.ioWrite(iopDPCR, 0x000B8800)
	p.ioWrite(base+iopDMABcr, 0x0001_0004)

	// Write MADR one byte at a time, the way the core does.
	const madr uint32 = 0x00123450
	for i := uint32(0); i < 4; i++ {
		p.Write(base+iopDMAMadr+i, byte(madr>>(8*i)))
	}
	if got := p.dma[iopDMAChSPU0].madr; got != madr {
		t.Fatalf("MADR came out as 0x%08X after four byte stores, want 0x%08X — the read half of "+
			"the merge is not reading the register", got, madr)
	}
}

func TestATimerReachingItsTargetRaisesItsLine(t *testing.T) {
	// The IOP's only sense of time passing. Without it THREADMAN's DelayThread computes a
	// deadline that never arrives, the thread sleeps forever, the scheduler finds nothing to
	// run, and the machine spends its whole life in a 64-bit division routine converting
	// microseconds to ticks that never tick.
	m := NewMachine()
	p := m.StartIOP()

	// Counter 5, programmed as THREADMAN programs it: interrupt on target, no auto-reset.
	const n = 5
	base := uint32(iopTimerHiBase + (n-3)*0x10)
	p.ioWrite(base+iopTimerTarget, 100)
	p.ioWrite(base+iopTimerMode, iopTimerIRQOnTarget|iopTimerIRQOnOverflow|iopTimerRepeat)

	for i := 0; i < 99; i++ {
		p.timerTick()
	}
	if p.raised[iopTimerIRQ(n)] != 0 {
		t.Fatal("the counter raised its line before it reached its target")
	}
	p.timerTick()
	if p.raised[iopTimerIRQ(n)] != 1 {
		t.Fatalf("counter %d raised interrupt %d %d times on reaching its target, want once",
			n, iopTimerIRQ(n), p.raised[iopTimerIRQ(n)])
	}

	// It is an alarm, not a metronome: it does not go off again until it is re-armed.
	for i := 0; i < 500; i++ {
		p.timerTick()
	}
	if p.raised[iopTimerIRQ(n)] != 1 {
		t.Errorf("the counter went off %d times without being re-armed", p.raised[iopTimerIRQ(n)])
	}

	// Re-arming it is a write to any of its three registers, which is what TIMEMANI does.
	p.ioWrite(base+iopTimerCount, 0)
	p.ioWrite(base+iopTimerTarget, 10)
	for i := 0; i < 11; i++ {
		p.timerTick()
	}
	if p.raised[iopTimerIRQ(n)] != 2 {
		t.Error("the counter did not go off again after being re-armed")
	}
}

func TestReadingATimersModeClearsItsFlagsButPeekingDoesNot(t *testing.T) {
	// TIMEMANI reads the mode register to find out *why* it was interrupted, and the two
	// "it happened" bits clear on that read. But the machine reads these registers too —
	// on the read half of every byte store's merge, and whenever an instrument prints one —
	// and a flag cleared by a read the guest never made is a reason the guest never learns.
	m := NewMachine()
	p := m.StartIOP()

	const n = 4
	base := uint32(iopTimerHiBase + (n-3)*0x10)
	p.ioWrite(base+iopTimerTarget, 4)
	p.ioWrite(base+iopTimerMode, iopTimerResetOnTarget|iopTimerIRQOnTarget|iopTimerRepeat)
	for i := 0; i < 4; i++ {
		p.timerTick()
	}

	if v, _ := p.timerPeek(base + iopTimerMode); v&iopTimerHitTarget == 0 {
		t.Fatal("the counter did not record that it reached its target")
	}
	// Peeking twice must not consume it.
	if v, _ := p.timerPeek(base + iopTimerMode); v&iopTimerHitTarget == 0 {
		t.Fatal("peeking the mode register cleared the flag: the machine's own reads are stealing " +
			"the guest's interrupts")
	}

	// The guest reads the mode with an `lw`, and on this bus an `lw` is four reads of the
	// same word. Every one of them has to see the flag, because the flag is bit 11 — byte 1
	// — and a register that cleared itself on the first byte's read would hand the driver a
	// word in which nothing had ever happened.
	//
	// That is not a hypothetical. THREADMAN's alarm handler dispatches on exactly this bit,
	// and clearing it three bytes early meant the handler was told the alarm it had just been
	// interrupted for had not gone off. Every thread that ever called DelayThread slept for
	// ever, on a machine where the interrupt was raised, delivered, and handled.
	var w uint32
	for i := uint32(0); i < 4; i++ {
		w |= uint32(p.Read(base+iopTimerMode+i)) << (8 * i)
	}
	if w&iopTimerHitTarget == 0 {
		t.Fatal("the word the guest loaded did not carry the flag: the read cleared it by the byte")
	}
	if v, _ := p.timerPeek(base + iopTimerMode); v&iopTimerHitTarget == 0 {
		t.Error("the flag went out in the middle of the load that was reading it")
	}

	// And it goes out when the instruction is over — which is what IOP.tick does before the
	// next one begins.
	p.timerAckFlush()
	if v, _ := p.timerPeek(base + iopTimerMode); v&iopTimerHitTarget != 0 {
		t.Error("the flag survived the read that should have cleared it")
	}
}

func TestAnInterruptFrameRoundTrips(t *testing.T) {
	// The frame is what THREADMAN is handed and what it gives back, so it has to carry a
	// context exactly. A register dropped here is a register that comes back wrong in a
	// thread that has been running perfectly well for a million instructions, which is a bug
	// that looks like anything but an interrupt.
	m := NewMachine()
	p := m.StartIOP()

	for i := uint32(1); i < 32; i++ {
		p.CPU.SetReg(i, 0x1000+i)
	}
	p.CPU.SetPC(0x00030000)
	before := p.CPU.SaveState()

	const frame = 0x00060000
	p.saveFrame(frame)

	// Scribble over everything, as a handler would.
	for i := uint32(1); i < 32; i++ {
		p.CPU.SetReg(i, 0xDEAD0000+i)
	}
	p.CPU.SetPC(0x00099999)

	p.loadFrame(frame)

	after := p.CPU.SaveState()
	for i := uint32(1); i < 32; i++ {
		if after.R[i] != before.R[i] {
			t.Errorf("register %d came back as 0x%08X, want 0x%08X", i, after.R[i], before.R[i])
		}
	}
	if after.PC != before.PC {
		t.Errorf("the frame resumed at 0x%08X, want 0x%08X", after.PC, before.PC)
	}

	// And the frame's shape is the one THREADMAN builds: register n at 4n, the entry point
	// at 140. StartThread writes a thread's argument to +16 and expects it in $a0, its $gp
	// to +112, its stack pointer to +116 and its entry to +140 — so if these move, a thread
	// THREADMAN starts comes up with its arguments in the wrong registers.
	if got := p.Read32(frame + 4*4); got != before.R[4] {
		t.Errorf("$a0 is not at frame+16: found 0x%08X, want 0x%08X", got, before.R[4])
	}
	if got := p.Read32(frame + 4*29); got != before.R[29] {
		t.Errorf("$sp is not at frame+116: found 0x%08X, want 0x%08X", got, before.R[29])
	}
	if got := p.Read32(frame + iopFrameEPC); got != before.PC {
		t.Errorf("the resume address is not at frame+140: found 0x%08X, want 0x%08X", got, before.PC)
	}
}

func TestInterruptsAreOnBeforeTheFirstModuleRuns(t *testing.T) {
	// A module is called, not booted: on the board the kernel that loads it has already
	// armed the processor's interrupts. Start this at false and it can never become true,
	// because every module brackets its critical sections with CpuSuspendIntr and
	// CpuResumeIntr — and Resume is passed *the value Suspend saved*. Suspend saves
	// "disabled", Resume faithfully restores "disabled", the round trip is perfect, and the
	// processor never takes another interrupt as long as it lives.
	m := NewMachine()
	p := m.StartIOP()

	if !p.intrEnabled {
		t.Fatal("the IOP starts with interrupts off, and nothing on the disc will ever turn them on")
	}

	// THREADMAN's own entry point suspends interrupts and never resumes them. What turns
	// them back on is CpuEnableIntr, the last thing it calls before it returns.
	p.CPU.SetReg(4, 0)
	p.intrSuspend()
	if p.intrEnabled {
		t.Fatal("CpuSuspendIntr did not disable interrupts")
	}
	p.intrCpuEnable()
	if !p.intrEnabled {
		t.Fatal("CpuEnableIntr did not re-enable them")
	}
}

func TestASnapshotCarriesTheWholeSecondProcessor(t *testing.T) {
	// The snapshot always carried the IOP's *memory*, because the EE can see it. Carrying
	// the memory and not the processor is the worst of both: the restored machine has all
	// the right code in all the right places and no registers, no syscall bindings, no heaps
	// and no interrupt handlers to run it with. It does not fail. It resumes, and goes wrong
	// somewhere else.
	m := NewMachine()
	p := m.StartIOP()
	p.running = true

	// A machine with something in every part of it that has to survive.
	p.CPU.SetReg(16, 0xCAFEBABE)
	p.CPU.SetPC(0x00012345 &^ 3)
	p.allocPtr = 0x00050000
	p.imask = 1<<9 | 1<<36
	p.handlers[36] = iopHandler{fn: 0x00051B7C, arg: 0x1234}
	p.schedSwitch, p.schedResched = 0x13040, 0x132C4
	p.timers[5] = iopTimer{count: 42, mode: 0x70, target: 0x1CCC}
	p.spu.ram[0x11040] = 0x5A
	p.steps = 999

	// A heap that has grown, so its handle and its base are different numbers.
	p.CPU.SetReg(4, 2048)
	p.CPU.SetReg(5, 1)
	p.heapCreate()
	handle := p.CPU.Reg(2)
	for i := 0; i < 40; i++ {
		p.CPU.SetReg(4, handle)
		p.CPU.SetReg(5, 192)
		p.heapAlloc()
	}
	if p.heaps[handle].base == handle {
		t.Fatal("the heap never grew, so this test is not testing what it says it is")
	}

	// And a syscall binding, which is the one thing that cannot be serialised as it stands:
	// its handler is a Go function.
	code := p.bindCall("intrman", 17)

	wantAlloc := p.allocPtr // the heap moved it; the point is that it comes back where it was

	before := m.SaveState()
	if before.IOP == nil {
		t.Fatal("the machine's snapshot does not carry the IOP at all")
	}

	// Restore into a fresh machine, as loading a file does.
	m2 := NewMachine()
	if err := m2.LoadState(before); err != nil {
		t.Fatal(err)
	}
	q := m2.IOP
	if q == nil {
		t.Fatal("the restored machine has no IOP")
	}

	if got := q.CPU.Reg(16); got != 0xCAFEBABE {
		t.Errorf("$s0 came back as 0x%08X", got)
	}
	if q.allocPtr != wantAlloc {
		t.Errorf("the allocator came back at 0x%08X, want 0x%08X — a restored machine that hands "+
			"out memory it has already given away is one that corrupts a module at random",
			q.allocPtr, wantAlloc)
	}
	if q.imask != 1<<9|1<<36 {
		t.Errorf("the interrupt mask came back as 0x%X — and an interrupt number past 32 is a "+
			"DMA channel, which is where the sound chip reports", q.imask)
	}
	if q.handlers[36].fn != 0x00051B7C || q.handlers[36].arg != 0x1234 {
		t.Error("the sound chip's DMA handler did not survive")
	}
	if q.schedSwitch != 0x13040 || q.schedResched != 0x132C4 {
		t.Error("THREADMAN's scheduler hooks did not survive: the restored machine cannot preempt")
	}
	if q.timers[5].target != 0x1CCC || q.timers[5].count != 42 {
		t.Error("the clock did not survive")
	}
	if q.spu.ram[0x11040] != 0x5A {
		t.Error("the sound memory did not survive")
	}
	if q.steps != 999 {
		t.Errorf("the instruction count came back as %d", q.steps)
	}

	// The heap, by the handle the guest is holding — not by its base, which has moved.
	h := q.heaps[handle]
	if h == nil {
		t.Fatal("the heap is gone from under its handle: every AllocHeapMemory from it now " +
			"comes back null, and THREADMAN runs out of thread control blocks")
	}
	if h.base == handle {
		t.Error("the heap came back at its handle rather than at the chunk it had grown into")
	}

	// And the syscall table, rebuilt: the code is baked into the `syscall` instruction in a
	// patched stub sitting in the RAM we just restored, so it has to come back as the same
	// number, bound to the same function.
	if got := q.bound[iopBinding{"intrman", 17}]; got != code {
		t.Errorf("intrman #17 was syscall %d and came back as %d: the stubs in memory still say %d",
			code, got, code)
	}
	if q.calls[code].fn == nil {
		t.Error("the rebuilt syscall has no handler: CpuSuspendIntr would return zero and do nothing")
	}
}
