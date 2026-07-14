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
