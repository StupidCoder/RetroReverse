package ps2

// iop_test.go checks the IOP's kernel HLE, on the points where being wrong is quiet.

import "testing"

func TestHeapGrowsPastItsFirstChunk(t *testing.T) {
	// THREADMAN creates its heap with a 2 KiB *chunk* size and then takes every thread
	// control block, semaphore and event flag in the machine out of it. A heap that stops
	// at its first chunk does not report an error anybody sees: AllocHeapMemory returns
	// zero, THREADMAN's CreateSema turns that into an error code, and OVERLORD — which
	// checks — jumps to itself forever, two instructions, no message, immediately after
	// printing that it was initialising sound. It took a disassembly to find, and it is
	// the reason this test exists.
	m := NewMachine()
	p := m.StartIOP()

	const chunk = 2048
	p.CPU.SetReg(4, chunk) // $a0 = the chunk size
	p.CPU.SetReg(5, 1)     // $a1 = flags
	p.heapCreate()

	heap := p.CPU.Reg(2)
	if heap == 0 {
		t.Fatal("CreateHeap returned null")
	}

	// Take four times the chunk out of it, in pieces, as a boot does.
	const (
		piece = 192
		want  = 4 * chunk / piece
	)
	seen := map[uint32]bool{}
	for i := 0; i < want; i++ {
		p.CPU.SetReg(4, heap)
		p.CPU.SetReg(5, piece)
		p.heapAlloc()

		a := p.CPU.Reg(2)
		if a == 0 {
			t.Fatalf("allocation %d of %d returned null: the heap did not grow past its first %d-byte chunk",
				i+1, want, chunk)
		}
		if seen[a] {
			t.Fatalf("allocation %d handed back 0x%08X, which was already given out", i+1, a)
		}
		if a+piece > iopRAMSizeBytes {
			t.Fatalf("allocation %d is at 0x%08X, past the end of IOP memory", i+1, a)
		}
		seen[a] = true
	}
}

func TestTheTwoProcessorsSeeOneSetOfSIFRegisters(t *testing.T) {
	// The SBUS registers appear at 0x1000F200 to the EE and at 0x1D000000 to the IOP, and
	// they are the same six words. If they are not — if each processor gets its own copy —
	// then every handshake between the two succeeds locally and means nothing, which is a
	// failure that looks like success right up until nothing arrives.
	m := NewMachine()
	p := m.StartIOP()

	// The EE raises the bit that says its half of the SIF is up.
	m.sbusFlagSet(sbusMSFLG, sifEESIFReady)

	// The IOP reads it, through its own address for the register, in KSEG1 — which is how
	// SIFMAN reads it.
	if got := p.Read32(0xBD000000 + sbusMSFLG); got&sifEESIFReady == 0 {
		t.Errorf("the IOP reads MSFLG as 0x%08X: it cannot see the bit the EE raised", got)
	}

	// And the other way: the IOP writes the flag the EE reads.
	p.Write32(0xBD000000+sbusSMFLG, 0x1234)
	if got := m.Read32(sbusEEBase + sbusSMFLG); got != 0x1234 {
		t.Errorf("the EE reads SMFLG as 0x%08X, want 0x1234", got)
	}
}

func TestRescheduleRestoresTheEnableDelegatedInA2(t *testing.T) {
	// THREADMAN's blocking primitives sleep INSIDE a CpuSuspendIntr critical section, and
	// they never call CpuResumeIntr on the way out — the paths that keep running restore
	// the saved enable themselves, and the paths that block hand that same saved value to
	// the reschedule syscall in $a2, delegating the restore to the kernel (sixteen call
	// sites load $a2 straight from the CpuSuspendIntr slot; THREADMAN+0x5D0 passes one
	// value to CpuResumeIntr on its no-switch exit and to $a2 on its switch exit). A yield
	// that saves the live enable instead parks every blocked thread with interrupts off —
	// which mostly heals, because the thread soon blocks again, and deadlocks the day a
	// woken thread busy-waits on an interrupt-fed flag. 989snd's SPU-DMA completion spin
	// was that day.
	m := NewMachine()
	p := m.StartIOP()

	const sp = 0x00100000
	frame := (uint32(sp) - iopFrameSize) &^ 7

	// The thread suspends interrupts, saving "they were on" — and then blocks, passing
	// what it saved in $a2, exactly as THREADMAN does.
	p.intrEnabled = true
	p.CPU.SetReg(4, 0x00100100) // CpuSuspendIntr's out-pointer
	p.intrSuspend()
	if p.intrEnabled {
		t.Fatal("CpuSuspendIntr left interrupts on")
	}

	p.CPU.SetReg(29, sp)
	p.CPU.SetReg(6, p.Read32(0x00100100)) // $a2 = the saved enable
	p.CPU.SetPC(0x00200000)
	p.yield()

	// No scheduler hooks are registered, so no switch happens — which on the hardware is
	// still an exception return, and the rfe brings $a2 into the enable.
	if !p.intrEnabled {
		t.Error("the reschedule returned with interrupts off: the enable delegated in $a2 was dropped")
	}
	// And the frame carries it too, so a real switch resumes this thread the same way.
	if sr := p.Read32(frame + iopFrameSR); sr>>2&1 != 1 {
		t.Errorf("the yielded frame's SR is 0x%08X: bit 2 should carry the $a2 enable", sr)
	}

	// The one caller that passes 0 — THREADMAN's stack-overflow trap, parking a dead
	// thread — must stay masked.
	p.intrEnabled = false
	p.CPU.SetReg(29, sp)
	p.CPU.SetReg(6, 0)
	p.CPU.SetPC(0x00200000)
	p.yield()
	if p.intrEnabled {
		t.Error("a reschedule with $a2=0 turned interrupts on")
	}
}
