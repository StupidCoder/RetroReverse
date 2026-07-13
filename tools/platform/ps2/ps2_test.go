package ps2

// ps2_test.go checks the parts of the machine that a boot exercises so early, and so
// silently, that a mistake in them looks like a fault somewhere else entirely.
//
// Every case here corresponds to a bug that actually happened while bringing the
// machine up. That is deliberate: a test that never could have failed is not worth
// the lines.

import (
	"os"
	"path/filepath"
	"testing"

	"retroreverse.com/tools/cpu/r5900"
)

func TestKUSEGIsMappedForBothMemoryAndHardware(t *testing.T) {
	// A PS2 program runs in KUSEG, which is *mapped*, and reaches its peripherals at
	// their bare physical addresses — also KUSEG. The kernel maps both before handing
	// over. Without the second, the first library call Jak makes (a timer read at
	// 0x10001810) faults on its own instruction.
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: []byte{0, 0, 0, 0}, MemSz: 4}},
	})

	for _, c := range []struct {
		vaddr uint32
		what  string
	}{
		{0x00100000, "the executable's load address"},
		{0x001FFFFC, "main memory"},
		{0x01FFFFFC, "the top of main memory"},
		{0x10001810, "the timer Jak reads first"},
		{0x1000E000, "the DMA controller"},
		{0x12000000, "the GS's privileged registers"},
	} {
		if _, ok := m.CPU.Translate(uint64(c.vaddr), false); !ok {
			t.Errorf("0x%08X (%s) does not translate — the kernel's mapping is missing it", c.vaddr, c.what)
		}
	}
}

func TestNewThreadInheritsTheTLB(t *testing.T) {
	// The TLB belongs to the machine, not to a thread. A thread context built from a
	// zero value restores an empty TLB the moment it is switched to, and every address
	// it touches then faults — including addresses that are plainly mapped, which is
	// what makes the symptom so misleading.
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 16), MemSz: 16}},
	})

	// A thread param block in memory: status, func, stack, stackSize, gp, priority.
	const p = 0x00110000
	m.Write32(p+0x04, 0x00100000) // entry
	m.Write32(p+0x08, 0x01F00000) // stack
	m.Write32(p+0x0C, 0x1000)     // stack size
	m.Write32(p+0x10, 0)          // gp
	m.Write32(p+0x14, 10)         // priority

	m.CPU.SetReg(4, p)
	m.createThread()
	id := uint32(m.CPU.Reg(2))
	if id == 0 {
		t.Fatal("CreateThread returned no thread")
	}
	th := m.threads[id]
	if th.entry != 0x00100000 {
		t.Fatalf("CreateThread read entry 0x%08X — the param struct opens with `status`, not `func`", th.entry)
	}

	m.CPU.SetReg(4, uint64(id))
	m.CPU.SetReg(5, 0)
	m.startThread()

	// The new thread outranks the main one, so it is running now. Its TLB must be the
	// machine's.
	if m.currentThread != id {
		t.Fatalf("a higher-priority thread did not preempt: current is %d", m.currentThread)
	}
	if _, ok := m.CPU.Translate(0x00100000, false); !ok {
		t.Error("the new thread cannot translate its own entry point — its context did not inherit the TLB")
	}
}

func TestSetSyscallReplacesAKernelCall(t *testing.T) {
	// SetSyscall lets the game claim a kernel call for itself, and GOAL uses it: `Copy`
	// is really the game's own `kCopy`. Dispatch must be a jump, so the handler's
	// `jr $ra` returns to whoever called the SDK wrapper.
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 64), MemSz: 64}},
	})

	const handler = 0x00123FC8
	m.CPU.SetReg(4, 90)
	m.CPU.SetReg(5, handler)
	m.setSyscall()

	// A syscall 90 now vectors to the game's routine rather than the kernel's.
	m.CPU.SetReg(3, 90) // $v1
	if !m.handleSyscall(m.CPU) {
		t.Fatal("handleSyscall declined")
	}
	if got := uint32(m.CPU.PC); got != handler {
		t.Errorf("syscall 90 went to 0x%08X, want the installed handler 0x%08X", got, handler)
	}

	// Clearing it puts the kernel's back.
	m.CPU.SetReg(4, 90)
	m.CPU.SetReg(5, 0)
	m.setSyscall()
	if _, ok := m.userSyscalls[90]; ok {
		t.Error("SetSyscall with a null handler did not restore the kernel's call")
	}
}

func TestSetupThreadReturnsAUsableStack(t *testing.T) {
	// crt0 moves SetupThread's return value straight into $sp. A stack of -1 means "the
	// top of memory". Return a thread id here — as a table that read the syscall
	// numbers as hex rather than decimal would — and the program runs on a garbage
	// stack and faults on its first store.
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 16), MemSz: 16}},
	})

	m.CPU.SetReg(4, 0x0013EC70) // gp
	m.CPU.SetReg(5, 0xFFFFFFFF) // stack: the top of memory
	m.CPU.SetReg(6, 0x4000)     // stack size
	m.CPU.SetReg(7, 0x00137100) // args
	m.CPU.SetReg(8, 0x001000C0) // root
	m.setupThread()

	sp := uint32(m.CPU.Reg(2))
	if sp&0xF != 0 {
		t.Errorf("SetupThread returned 0x%08X, which is not quadword-aligned", sp)
	}
	if sp >= ramSize || sp < ramSize-0x4000 {
		t.Errorf("SetupThread returned 0x%08X, which is not inside the stack it was given", sp)
	}
	if _, ok := m.CPU.Translate(uint64(sp), true); !ok {
		t.Errorf("SetupThread returned 0x%08X, which does not translate — a store there would fault", sp)
	}
}

func TestSemaphoreBlocksAndReleases(t *testing.T) {
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 16), MemSz: 16}},
	})

	const p = 0x00110000
	m.Write32(p+0x00, 0) // no count: a waiter must block
	m.Write32(p+0x04, 1) // max count
	m.CPU.SetReg(4, p)
	m.createSema()
	id := uint32(m.CPU.Reg(2))

	m.waitSema(id)
	if got := m.threads[1].state; got != thWaitSema {
		t.Fatalf("a wait on an empty semaphore left the thread %v, not blocked", got)
	}

	m.signalSema(id)
	if got := m.threads[1].state; got != thReady {
		t.Errorf("a signal did not release the waiter: it is %v", got)
	}
	// The count must not also have been incremented — a waiter takes it directly, or a
	// second thread could take it in between.
	if s := m.semas[id]; s.count != 0 {
		t.Errorf("the signal both released a waiter and raised the count to %d", s.count)
	}
}

func TestSavestateRoundTrip(t *testing.T) {
	// Savestates land in the first phase of a platform, not the last. This checks the
	// thing that is easy to get wrong: the *running* thread's context lives in the CPU,
	// not in its saved slot, so a state that serialised the stale slot would resume it
	// wherever it last yielded.
	m := NewMachine()
	m.SetImageHash("abc")
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 64), MemSz: 64}},
	})
	m.CPU.SetPC(0x00100020)
	m.CPU.SetReg(16, 0xDEADBEEF)
	m.CPU.SetQuad(17, r5900.Quad{Lo: 1, Hi: 2})
	m.userSyscalls[90] = 0x123456
	m.heapPtr, m.heapEnd = 0x200000, 0x300000
	m.Write32(0x00100010, 0xCAFEBABE)

	dir := t.TempDir()
	path := filepath.Join(dir, "s.gz")
	if err := m.SaveStateFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	d := NewMachine()
	d.SetImageHash("abc")
	if err := d.LoadStateFile(path); err != nil {
		t.Fatal(err)
	}

	if d.CPU.PC != 0x00100020 {
		t.Errorf("PC = 0x%X, want 0x00100020", d.CPU.PC)
	}
	if d.CPU.Reg(16) != 0xDEADBEEF {
		t.Errorf("$s0 did not round-trip")
	}
	if got := d.CPU.Quad(17); got.Lo != 1 || got.Hi != 2 {
		t.Errorf("the 128-bit half of $s1 did not round-trip: %v", got)
	}
	if d.Read32(0x00100010) != 0xCAFEBABE {
		t.Errorf("memory did not round-trip")
	}
	if d.userSyscalls[90] != 0x123456 {
		t.Errorf("the game's installed syscalls did not round-trip")
	}
	// The running thread's saved context must be the live CPU, not a stale slot.
	if ts := d.threads[d.currentThread]; ts.ctx.PC != 0x00100020 {
		t.Errorf("the running thread's saved context is stale: PC = 0x%X", ts.ctx.PC)
	}

	// A state from a different disc must be refused rather than produce nonsense.
	e := NewMachine()
	e.SetImageHash("different")
	if err := e.LoadStateFile(path); err == nil {
		t.Error("a savestate from another disc was accepted")
	}
}
