package xbox

import (
	"testing"

	"retroreverse.com/tools/cpu/x86"
)

// timer_test.go pins the timer queue. Two of these are the reason the pad works, and
// both would have been easy to get subtly wrong in a way that still ran.

// timerMachine builds the smallest machine the timer queue needs: RAM to hold KTIMERs
// and KDPCs, an object table, and a clock.
func timerMachine(t *testing.T) *Machine {
	t.Helper()
	m := mmioMachine(t)
	m.objects = map[uint32]*kobject{}
	return m
}

// TestRelativeDueTimeIsAnIntervalNotADeadline is the one that matters most, because
// getting it wrong fires the timer rather than failing to.
//
// The kernel's convention is that a NEGATIVE due time is an interval from now and a
// positive one is an absolute time. XAPI's port debounce passes -1,000,000 (-100 ms).
// Read as an absolute deadline that is a moment three millennia in the past, so the
// timer would expire on the very next tick — the debounce would be skipped, the port
// reset immediately, and it would all look like it worked.
func TestRelativeDueTimeIsAnIntervalNotADeadline(t *testing.T) {
	m := timerMachine(t)
	m.tick = 5000 * instrsPerMs // some time has already passed
	const tm, dpc = 0x1000, 0x2000

	// The title's own value, from 0x241C01: DueTime = 0xFFFFFFFF_FFF0BDC0.
	m.armTimer(tm, dpc, 0xFFF0BDC0, 0xFFFFFFFF, 0)

	if len(m.timers) != 1 {
		t.Fatalf("armed %d timers, want 1", len(m.timers))
	}
	want := m.systemTime100ns() + 1_000_000 // 100 ms in 100-ns units
	if got := m.timers[0].Due; got != want {
		t.Errorf("deadline = %d, want %d (a negative due time is an interval from NOW)", got, want)
	}
	// And it must NOT be due yet.
	m.timerTick()
	if len(m.timers) != 1 {
		t.Error("the timer expired immediately: a relative due time was read as absolute")
	}
}

// TestAbsoluteDueTimeIsHonoured — the other half of the same convention.
func TestAbsoluteDueTimeIsHonoured(t *testing.T) {
	m := timerMachine(t)
	m.tick = 100 * instrsPerMs
	const tm = 0x1000
	deadline := m.systemTime100ns() + 50_000_000
	m.armTimer(tm, 0, uint32(deadline), uint32(deadline>>32), 0)
	if got := m.timers[0].Due; got != deadline {
		t.Errorf("deadline = %d, want the absolute %d", got, deadline)
	}
}

// TestTimerExpiryQueuesItsDPC is the debounce completing.
func TestTimerExpiryQueuesItsDPC(t *testing.T) {
	m := timerMachine(t)
	const tm, dpc, routine = 0x1000, 0x2000, 0x3000
	m.write32(dpc+0x0C, routine) // DeferredRoutine, where KeInitializeDpc puts it

	m.armTimer(tm, dpc, 0xFFF0BDC0, 0xFFFFFFFF, 0) // -100 ms
	m.timerTick()
	if len(m.dpcQueue) != 0 {
		t.Fatal("the DPC ran before its deadline")
	}

	m.tick += 200 * instrsPerMs // 200 ms later: well past
	m.CPU.IF = false            // gates shut, so the DPC queues but does not run yet
	m.timerTick()

	if len(m.timers) != 0 {
		t.Error("a one-shot timer stayed armed after expiring")
	}
	if len(m.dpcQueue) != 1 || m.dpcQueue[0].Dpc != dpc {
		t.Fatalf("dpcQueue = %+v, want the armed DPC %08X", m.dpcQueue, uint32(dpc))
	}
	if m.read32(tm+dhSignalState) == 0 {
		t.Error("the expired timer was not signalled")
	}
}

// TestExpiredTimerRunsItsDPCFrame: a timer's DPC has no ISR in front of it, so without
// its own entry point it would sit in the queue until an unrelated device's interrupt
// happened to drain it.
func TestExpiredTimerRunsItsDPCFrame(t *testing.T) {
	m := timerMachine(t)
	const tm, dpc, routine, ctx = 0x1000, 0x2000, 0x3000, 0x4444
	m.write32(dpc+0x0C, routine)
	m.write32(dpc+0x10, ctx)
	m.CPU.IF = true
	m.CPU.IP = 0x9999
	m.CPU.Regs[x86.SP] = 0x8000 // a stack for the frame to be built on

	m.armTimer(tm, dpc, 0xFFF0BDC0, 0xFFFFFFFF, 0)
	m.tick += 200 * instrsPerMs
	m.timerTick()

	if !m.isrActive {
		t.Fatal("no DPC frame was entered")
	}
	if m.CPU.IP != routine {
		t.Errorf("PC = %08X, want the DeferredRoutine %08X", m.CPU.IP, uint32(routine))
	}
	// DeferredRoutine(Dpc, DeferredContext, SystemArgument1, SystemArgument2), with the
	// sentinel on top so isrReturn restores the interrupted context.
	sp := m.CPU.Regs[x86.SP]
	if got := m.read32(sp); got != isrExitAddr {
		t.Errorf("return address = %08X, want the ISR sentinel %08X", got, uint32(isrExitAddr))
	}
	if got := m.read32(sp + 4); got != dpc {
		t.Errorf("arg0 = %08X, want the KDPC %08X", got, uint32(dpc))
	}
	if got := m.read32(sp + 8); got != ctx {
		t.Errorf("arg1 = %08X, want the DeferredContext %08X", got, uint32(ctx))
	}
}

// TestDPCFrameWaitsForTheGates: a DPC must not be forced into a machine that is already
// running an interrupt frame or has interrupts off.
func TestDPCFrameWaitsForTheGates(t *testing.T) {
	m := timerMachine(t)
	const dpc, routine = 0x2000, 0x3000
	m.write32(dpc+0x0C, routine)
	m.queueDPC(dpc, 0, 0)

	m.isrActive = true
	m.CPU.IF = true
	m.deliverDPC()
	if m.CPU.IP == routine {
		t.Error("entered a DPC frame while an interrupt frame was already running")
	}

	m.isrActive = false
	m.CPU.IF = false
	m.deliverDPC()
	if m.CPU.IP == routine {
		t.Error("entered a DPC frame with interrupts disabled")
	}

	m.CPU.IF = true
	m.deliverDPC()
	if m.CPU.IP != routine {
		t.Error("did not enter the DPC frame once the gates opened")
	}
}

// TestPeriodicTimerRearms — KeSetTimerEx's period.
func TestPeriodicTimerRearms(t *testing.T) {
	m := timerMachine(t)
	const tm, dpc, routine = 0x1000, 0x2000, 0x3000
	m.write32(dpc+0x0C, routine)
	m.CPU.IF = false

	m.armTimer(tm, dpc, 0xFFE17B80, 0xFFFFFFFF, 200) // the NVNET poll: -200 ms, period 200 ms
	m.tick += 300 * instrsPerMs
	m.timerTick()

	if len(m.timers) != 1 {
		t.Fatal("a periodic timer did not re-arm after expiring")
	}
	if want := m.systemTime100ns() + 200*10000; m.timers[0].Due != want {
		t.Errorf("re-armed deadline = %d, want %d (one period from now)", m.timers[0].Due, want)
	}
}

// TestReArmingReplaces: arming a timer that is already armed must move it, not stack a
// second copy that would fire twice.
func TestReArmingReplaces(t *testing.T) {
	m := timerMachine(t)
	const tm = 0x1000
	if was := m.armTimer(tm, 0, 0xFFF0BDC0, 0xFFFFFFFF, 0); was {
		t.Error("a fresh timer reported it was already queued")
	}
	if was := m.armTimer(tm, 0, 0xFFF0BDC0, 0xFFFFFFFF, 0); !was {
		t.Error("re-arming an armed timer did not report it was already queued")
	}
	if len(m.timers) != 1 {
		t.Errorf("%d timers armed for one KTIMER; re-arming must replace", len(m.timers))
	}
}

// TestCancelTimerAnswersForTheRealQueue is the regression for a stub that was correct
// only because its neighbour did nothing: KeCancelTimer used to answer "not pending"
// always, which was true of a machine that never queued a timer.
func TestCancelTimerAnswersForTheRealQueue(t *testing.T) {
	m := timerMachine(t)
	const tm = 0x1000
	if m.cancelTimer(tm) {
		t.Error("cancelling an unarmed timer reported it was pending")
	}
	m.armTimer(tm, 0, 0xFFF0BDC0, 0xFFFFFFFF, 0)
	if !m.cancelTimer(tm) {
		t.Error("cancelling an armed timer reported it was not pending")
	}
	if len(m.timers) != 0 {
		t.Error("the cancelled timer is still queued")
	}
}

// TestDeadlineSurvivesAClockJump: Due is an absolute deadline, so a savestate restored
// past it expires immediately rather than restarting the wait.
func TestDeadlineSurvivesAClockJump(t *testing.T) {
	m := timerMachine(t)
	const tm, dpc, routine = 0x1000, 0x2000, 0x3000
	m.write32(dpc+0x0C, routine)
	m.CPU.IF = false
	m.armTimer(tm, dpc, 0xFFF0BDC0, 0xFFFFFFFF, 0)
	due := m.timers[0].Due

	// The shape of a restore: the clock is somewhere else entirely.
	m.tick = 900_000 * instrsPerMs
	if m.timers[0].Due != due {
		t.Error("the deadline moved when the clock did: it is not absolute")
	}
	m.timerTick()
	if len(m.dpcQueue) != 1 {
		t.Error("a deadline long past did not expire on the next tick")
	}
}

// TestQueueDPCDoesNotDuplicate mirrors KeInsertQueueDpc's own rule.
func TestQueueDPCDoesNotDuplicate(t *testing.T) {
	m := timerMachine(t)
	if !m.queueDPC(0x2000, 0, 0) {
		t.Error("queueing a fresh DPC reported it was already queued")
	}
	if m.queueDPC(0x2000, 0, 0) {
		t.Error("queueing an already-queued DPC reported success")
	}
	if len(m.dpcQueue) != 1 {
		t.Errorf("dpcQueue has %d entries, want 1", len(m.dpcQueue))
	}
}
