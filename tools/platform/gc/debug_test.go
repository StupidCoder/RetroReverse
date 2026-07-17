package gc

// debug_test.go pins the stack walker against a stack built the way the game's own compiler
// builds one.
//
// This test exists because the walker was wrong for the whole life of the file and nothing
// noticed. It read each frame's return address out of the frame itself, which on PowerPC is
// the one place it is not — and the failure is quiet, because the address it finds instead is
// a real return address into the very function being reported, so every trace looked sane.
// The test therefore does not ask "is the trace plausible"; it builds a stack whose right
// answer is known by construction and demands exactly that answer.

import (
	"testing"

	"retroreverse.com/tools/cpu/gekko"
)

// btMachine is a machine with the console's own block map installed, because the walker reads
// the stack through the CPU's translation and a stack pointer is an 0x8-prefixed effective
// address. Without setupState's BATs the machine starts with translation off, every 0x802FFExx
// read runs off the end of RAM, and the test would measure the absence of a memory map rather
// than the walker.
func btMachine() *Machine {
	m := &Machine{RAM: make([]byte, RAMSize), logSeen: map[string]bool{}}
	m.CPU = gekko.NewCPU(m)
	m.setupState()
	return m
}

// pushFrame writes one stack frame the way a CodeWarrior prologue leaves it:
//
//	stw   r0,4(r1)      the RETURN ADDRESS goes into the CALLER's frame, at caller+4
//	stwu  r1,-N(r1)     the new frame's first word points back at the caller's
//
// so a frame carries its caller's pointer and its CALLEE's return address. It returns the new
// frame's stack pointer.
func pushFrame(m *Machine, callerSP, sp, retAddr uint32) uint32 {
	write32(m, sp, callerSP)        // back chain: [sp] = caller's frame
	write32(m, callerSP+4, retAddr) // the return address, in the CALLER's frame
	return sp
}

func write32(m *Machine, addr, v uint32) {
	p := addr &^ 0x80000000
	m.RAM[p] = byte(v >> 24)
	m.RAM[p+1] = byte(v >> 16)
	m.RAM[p+2] = byte(v >> 8)
	m.RAM[p+3] = byte(v)
}

// TestBacktraceReportsCallersNotItself builds A -> B -> C and demands the callers.
//
// The stack grows down, so C's frame is at the lowest address and the back chain climbs. The
// answer is fixed by construction: C's return address is an address in B, B's is in A, and A's
// own is never reported because A is the root — there is no frame above it to have saved one.
func TestBacktraceReportsCallersNotItself(t *testing.T) {
	m := btMachine()

	const (
		spA = 0x80300000 // the outermost frame
		spB = 0x802FFF00
		spC = 0x802FFE00 // the innermost — where r1 points now

		retIntoA = 0x80001111 // B's return address: an address inside A
		retIntoB = 0x80002222 // C's return address: an address inside B
		retIntoC = 0x80003333 // a callee of C's return address: an address inside C
	)

	write32(m, spA, 0) // the root's back chain terminates the walk
	pushFrame(m, spA, spB, retIntoA)
	pushFrame(m, spB, spC, retIntoB)

	// THE TRAP THIS TEST IS FOR: C has already called something and returned, so C's own
	// frame's +4 slot holds a return address pointing back into C. A walker reading [sp+4]
	// reports this as C's caller, and it is not a caller at all.
	write32(m, spC+4, retIntoC)

	m.CPU.GPR[1] = spC

	got := m.Backtrace()
	want := []uint32{retIntoB, retIntoA}
	if len(got) != len(want) {
		t.Fatalf("Backtrace() returned %d frames, want %d: got %#x, want %#x", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("frame #%d = 0x%08X, want 0x%08X (full trace %#x)", i, got[i], want[i], got)
		}
	}
	for _, a := range got {
		if a == retIntoC {
			t.Errorf("the trace reports 0x%08X — that is C's own callee's return address, "+
				"not a caller; the walker is reading [sp+4] of the current frame", retIntoC)
		}
	}
}

// TestBacktraceStopsOnBrokenChain: a corrupt or absent chain must yield a short trace rather
// than a loop, a panic, or an address read out of unrelated memory.
func TestBacktraceStopsOnBrokenChain(t *testing.T) {
	for _, tc := range []struct {
		name  string
		chain uint32
	}{
		{"null back chain", 0},
		{"unaligned", 0x802FFF01},
		{"below RAM", 0x7FFFFFF0},
		{"past RAM", 0x81800000},
		{"does not climb", 0x80100000}, // a chain pointing DOWN would walk forever
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &Machine{RAM: make([]byte, RAMSize), logSeen: map[string]bool{}}
			m.CPU = gekko.NewCPU(m)
			const sp = 0x80200000
			write32(m, sp, tc.chain)
			m.CPU.GPR[1] = sp
			if got := m.Backtrace(); len(got) != 0 {
				t.Errorf("Backtrace() on a %s returned %#x, want no frames", tc.name, got)
			}
		})
	}
}
