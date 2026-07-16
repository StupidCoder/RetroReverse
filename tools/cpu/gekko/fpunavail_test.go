package gekko

import "testing"

// TestFPUnavailable covers the lazy floating-point context switch, which is the reason
// the FP-unavailable exception exists on this machine and not an academic corner.
//
// An OS with threads does not save the FPRs on a context switch. It runs a thread with
// MSR[FP] clear until the thread's first floating-point instruction traps here, and only
// then swaps the register file. A core that executes floating point regardless never
// raises the trap, the swap never happens, and the FPRs quietly become shared by every
// thread on the machine — which corrupts whichever thread was preempted mid-computation.
func TestFPUnavailable(t *testing.T) {
	// One encoding per class that requires the FPU, so that the whole gate is covered
	// rather than the one opcode the bug was found on.
	fp := []struct {
		name string
		op   uint32
	}{
		{"lfs f2,0(r6)", 0xC0460000},
		{"lfsu f2,4(r6)", 0xC4460004},
		{"lfd f2,0(r6)", 0xC8460000},
		{"lfdu f2,8(r6)", 0xCC460008},
		{"stfs f2,0(r6)", 0xD0460000},
		{"stfsu f2,4(r6)", 0xD4460004},
		{"stfd f2,0(r6)", 0xD8460000},
		{"stfdu f2,8(r6)", 0xDC460008},
		{"fadds f0,f2,f0", 0xEC02002A},
		{"fadd f0,f2,f0", 0xFC02002A},
		{"fmr f0,f2", 0xFC001090},
		{"ps_add f0,f2,f0", 0x1002002A},
		{"psq_l f2,0(r6)", 0xE0460000},
		{"psq_st f2,0(r6)", 0xF0460000},
		{"lfsx f2,r0,r6", 0x7C40342E},
		{"lfsux f2,r6,r0", 0x7C46046E},
		{"lfdx f2,r0,r6", 0x7C4034AE},
		{"lfdux f2,r6,r0", 0x7C4604EE},
		{"stfsx f2,r0,r6", 0x7C40352E},
		{"stfsux f2,r6,r0", 0x7C46056E},
		{"stfdx f2,r0,r6", 0x7C4035AE},
		{"stfdux f2,r6,r0", 0x7C4605EE},
		{"stfiwx f2,r0,r6", 0x7C4037AE},
	}

	for _, c := range fp {
		t.Run(c.name, func(t *testing.T) {
			cpu, m := fpuTestCPU(c.op)
			cpu.MSR &^= MSRFP // the thread does not own the FPU
			before := cpu.FPR[2]
			msr := cpu.MSR

			cpu.Step()

			if cpu.Halted {
				t.Fatalf("the core halted: %s", cpu.HaltReason)
			}
			if cpu.PC != VecFPUnavail {
				t.Errorf("PC = 0x%08X, want the FP-unavailable vector 0x%08X", cpu.PC, VecFPUnavail)
			}
			// SRR0 must name the offending instruction, not the one after it: the
			// handler returns to it and it runs again with the right registers.
			if cpu.SRR0 != 0x1000 {
				t.Errorf("SRR0 = 0x%08X, want 0x1000 (the instruction itself)", cpu.SRR0)
			}
			if cpu.SRR1 != msr&0x87C0FF73 {
				t.Errorf("SRR1 = 0x%08X, want 0x%08X", cpu.SRR1, msr&0x87C0FF73)
			}
			// The instruction must not have taken effect.
			if cpu.FPR[2] != before {
				t.Errorf("f2 = %v, want it untouched at %v — the instruction ran anyway",
					cpu.FPR[2], before)
			}
			if m.Read32(0x2000) != 0xDEADBEEF {
				t.Errorf("memory at 0x2000 = 0x%08X, want it untouched — the store ran anyway",
					m.Read32(0x2000))
			}
		})
	}

	// With MSR[FP] set the same instructions must run, or the gate is simply blocking
	// floating point rather than deferring it.
	for _, c := range fp {
		t.Run(c.name+"/enabled", func(t *testing.T) {
			cpu, _ := fpuTestCPU(c.op)
			cpu.MSR |= MSRFP

			cpu.Step()

			if cpu.Halted {
				t.Fatalf("the core halted: %s", cpu.HaltReason)
			}
			if cpu.PC != 0x1004 {
				t.Errorf("PC = 0x%08X, want 0x1004 — it trapped with MSR[FP] set", cpu.PC)
			}
		})
	}

	// Instructions that do not name a floating-point register must never trap, however
	// MSR[FP] stands. A gate that is too wide would send integer code to the handler.
	notFP := []struct {
		name string
		op   uint32
	}{
		{"addi r3,r6,1", 0x38660001},
		{"lwz r3,0(r6)", 0x80660000},
		{"stw r3,0(r6)", 0x90660000},
		{"lwzx r3,r0,r6", 0x7C60302E},
		{"stwx r3,r0,r6", 0x7C60312E},
		{"dcbz r0,r6", 0x7C0037EC},
		{"or r3,r6,r6", 0x7CC33378},
	}
	for _, c := range notFP {
		t.Run(c.name+"/must-not-trap", func(t *testing.T) {
			cpu, _ := fpuTestCPU(c.op)
			cpu.MSR &^= MSRFP

			cpu.Step()

			if cpu.Halted {
				t.Fatalf("the core halted: %s", cpu.HaltReason)
			}
			if cpu.PC == VecFPUnavail {
				t.Errorf("it trapped to the FP-unavailable vector, but it is not an FP instruction")
			}
		})
	}
}

// fpuTestCPU builds a core with the one instruction at 0x1000, translation off, and a
// known word at the address the loads and stores use.
func fpuTestCPU(op uint32) (*CPU, *testRAM) {
	m := &testRAM{}
	cpu := NewCPU(m)
	cpu.MSR = 0 // translation off: effective addresses are physical
	cpu.PC = 0x1000
	m.Write32(0x1000, op)
	m.Write32(0x2000, 0xDEADBEEF)
	m.Write32(0x2004, 0xDEADBEEF)
	cpu.GPR[6] = 0x2000
	cpu.GPR[0] = 0
	cpu.FPR[2] = FPR{PS0: 1.5, PS1: 1.5}
	cpu.setHID2(HID2PSE) // the paired-single cases need the unit enabled
	return cpu, m
}
