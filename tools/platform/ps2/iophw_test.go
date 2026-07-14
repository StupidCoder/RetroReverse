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
	// Reading it does.
	if v, _ := p.timerRead(base + iopTimerMode); v&iopTimerHitTarget == 0 {
		t.Fatal("the guest's read did not see the flag")
	}
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
