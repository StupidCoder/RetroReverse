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

// --- the scheduler, the SIF and the RPC ---------------------------------------

func TestBlockedThreadStopsRunning(t *testing.T) {
	// The subtlest bug in the machine so far. A model that marks a thread blocked and
	// then keeps executing it has not blocked it: SleepThread returns immediately, and
	// every "ask the IOP, sleep, read the answer when you wake" sequence — which is all
	// of them — reads the answer before it has arrived. The symptom is a *blocking* call
	// that behaves like a polling one, and the game re-asking a question it should
	// already have the answer to.
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 16), MemSz: 16}},
	})

	m.sleepThread()

	if got := m.threads[1].state; got != thSleeping {
		t.Fatalf("SleepThread left the thread %v, not sleeping", got)
	}
	if !m.idle {
		t.Fatal("every thread is blocked, but the machine is not idle — it will keep executing one of them")
	}

	// And a wakeup brings it back.
	m.wakeupThread(1)
	if !m.resume() {
		t.Fatal("a woken thread did not become runnable")
	}
	if m.idle {
		t.Error("the machine is still idle after a thread woke")
	}
}

func TestWakeupBeforeSleepIsRemembered(t *testing.T) {
	// The kernel counts wakeups; it does not flag them. A WakeupThread that lands before
	// the matching SleepThread must make that sleep return at once, or the two race and
	// the thread sleeps forever.
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 16), MemSz: 16}},
	})

	m.wakeupThread(1) // arrives first
	m.sleepThread()

	if m.threads[1].state == thSleeping {
		t.Error("a sleep after an early wakeup blocked anyway — the wakeup was not remembered")
	}
	if m.idle {
		t.Error("the machine went idle on a sleep that should not have blocked")
	}
}

func TestSifCommandAndDataDescriptorsAreBothTransferred(t *testing.T) {
	// A remote call hands over two buffers: the arguments, then the command packet saying a
	// call has been made. Both cross. Only the second rings the doorbell.
	//
	// This test used to assert the opposite of half of that — that the arguments stayed in EE
	// memory and only the command was carried — and it was right about the machine it was
	// written for, where the thing serving the call was a Go function that could read EE
	// memory. The IOP cannot. An argument buffer that never crosses arrives on the second
	// processor as whatever was there before, and FILEIO opens a file called "".
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 16), MemSz: 16}},
	})

	const args = 0x00110000
	const pkt = 0x00120000
	const desc = 0x00130000

	copy(m.ram[args:], "cdrom0:\\DRIVERS\\SIO2MAN.IRX;1\x00")

	// The command packet: CHANGE_SADDR, whose payload is the EE's command buffer.
	m.Write32(pkt+0x00, 0x14)
	m.Write32(pkt+0x08, sifCmdChangeSaddr)
	m.Write32(pkt+0x10, 0x0013B9C0)

	m.Write32(desc+0x00, args) // descriptor 0: the arguments, with no interrupt
	m.Write32(desc+0x04, 0x00050000)
	m.Write32(desc+0x08, 32)
	m.Write32(desc+0x0C, 0)
	m.Write32(desc+0x10, pkt) // descriptor 1: the command, with one
	m.Write32(desc+0x14, 0x0001B9B0)
	m.Write32(desc+0x18, 0x14)
	m.Write32(desc+0x1C, 0x44)

	m.CPU.SetReg(4, desc)
	m.CPU.SetReg(5, 2)
	m.sifSetDma()

	if m.sifCmdBuf != 0x0013B9C0 {
		t.Errorf("the command packet was not read: cmdBuf = 0x%08X", m.sifCmdBuf)
	}
	if len(m.sifToIOPQueue) != 2 {
		t.Fatalf("%d transfers are queued for the IOP, want 2 — the arguments and the command",
			len(m.sifToIOPQueue))
	}
	if m.sifToIOPQueue[0].cmd {
		t.Error("the argument buffer will ring SIFCMD's doorbell; it will dispatch a command out of a file path")
	}
	if !m.sifToIOPQueue[1].cmd {
		t.Error("the command packet will not ring SIFCMD's doorbell, so the IOP will never look at it")
	}
	if got := string(m.sifToIOPQueue[0].data[:7]); got != "cdrom0:" {
		t.Errorf("the arguments queued for the IOP are %q, not the path the EE passed", got)
	}
	if uint32(m.CPU.Reg(2)) == 0 {
		t.Error("sceSifSetDma returned a zero handle, which the caller reads as failure")
	}
}

func TestSifCommandWaitsForTheIOPToBeListening(t *testing.T) {
	// SIFCMD has one receive slot. Its handler empties the slot and only then re-arms the
	// receive channel, so nothing can land until it has. The EE runs eight times faster than
	// the IOP here and sends INIT_CMD(opt=0) and INIT_CMD(opt=1) back to back; deliver both
	// the instant they are asked for and the second overwrites the first in the slot they
	// share. The first is the packet that releases every thread on the second processor.
	m := NewMachine()
	m.LoadExecutable(&Executable{
		Entry:    0x00100000,
		Segments: []Segment{{VAddr: 0x00100000, Data: make([]byte, 16), MemSz: 16}},
	})
	m.StartIOP()

	const dest = 0x0001B9B0
	m.Write32(0x00110000, 0x18)
	m.Write32(0x00110008, sifCmdInitCmd)

	m.sifToIOP(0x00110000, dest, 0x18)
	if m.IOP.Read32(dest) != 0 {
		t.Fatal("the packet landed in a receive slot the IOP had not armed")
	}
	if len(m.sifToIOPQueue) != 1 {
		t.Fatal("the packet was dropped rather than held for a receiver that is not listening")
	}

	// SIFMAN arms the channel: a start bit, a block size, and no MADR at all.
	m.IOP.dma[iopDMAChSIF1].chcr = iopChcrStart
	m.sifPump()

	if m.IOP.Read32(dest+8) != sifCmdInitCmd {
		t.Error("the channel was armed and the waiting packet was not delivered")
	}
	if m.IOP.dma[iopDMAChSIF1].chcr&iopChcrStart != 0 {
		t.Error("the channel is still armed after a transfer landed in it")
	}
}

func TestTheIOPsCommandBufferIsNotCachedUntilTheHandshake(t *testing.T) {
	// The whole SIF handshake turns on this register, and it is not the register it looks
	// like. sceSifGetReg(0x80000000) is the EE kernel's *cache* of the IOP's command buffer,
	// and what sceSifInitCmd is really asking with it is "have I met this IOP before?".
	//
	//   not zero: yes — just send CHANGE_SADDR and carry on.
	//   zero:     no  — wait for the IOP, then send INIT_CMD with opt = 0.
	//
	// Only the second releases the second processor: INIT_CMD(opt=0) is what sets the event
	// flag every module on the IOP blocks on. Answer this register with anything non-zero and
	// the EE takes the short branch, the IOP is never introduced to it, and both processors
	// wait for each other forever while looking, from every other angle, perfectly busy.
	//
	// It used to be answered with a constant 1, because the argument was masked to five bits
	// and collapsed onto register 0. That constant reached the IOP as the destination address
	// of every packet the EE sent.
	m := NewMachine()

	m.CPU.SetReg(4, sifRegIOPCmdBuf)
	m.sifGetReg()
	if got := uint32(m.CPU.Reg(2)); got != 0 {
		t.Errorf("the EE thinks it already knows the IOP's command buffer (0x%08X); it will never send INIT_CMD", got)
	}

	// The flags come from the IOP, across the SBUS, and each bit has a module that raises it.
	m.CPU.SetReg(4, sifRegIOPFlags)
	m.sifGetReg()
	if uint32(m.CPU.Reg(2))&sifIOPRebootDone != 0 {
		t.Error("the IOP says it has finished rebooting before it has started")
	}

	m.sbusFlagSet(sbusSMFLG, sifIOPRebootDone) // EESYNC's boot callback
	m.CPU.SetReg(4, sifRegIOPFlags)
	m.sifGetReg()
	if uint32(m.CPU.Reg(2))&sifIOPRebootDone == 0 {
		t.Error("the IOP has rebooted and does not say so; the game syncs forever")
	}

	// And the EE clears the bit it has acted on, which is what sceSifSyncIop does with it.
	m.CPU.SetReg(4, sifRegIOPFlags)
	m.CPU.SetReg(5, sifIOPRebootDone)
	m.sifSetReg()
	if m.sbusRead(sbusSMFLG)&sifIOPRebootDone != 0 {
		t.Error("the EE acknowledged the reboot flag and it is still raised")
	}
}

func TestIOPRebootImageIsTheOneTheGameNames(t *testing.T) {
	// The reboot request is a sentence, and the image is in it. This is the exact string the
	// game's sceSifResetIop sends, read off the packet.
	got, err := iopRebootImage(`rom0:UDNL cdrom0:\DRIVERS\IOPRP221.IMG;1`)
	if err != nil {
		t.Fatal(err)
	}
	if want := "/DRIVERS/IOPRP221.IMG"; got != want {
		t.Errorf("the EE asked for %s and the disc was asked for %s", want, got)
	}

	// And a request naming something that is not on the disc is an error, not a shrug that
	// boots the file we were expecting anyway.
	if _, err := iopRebootImage("rom0:UDNL rom0:SOMETHING"); err == nil {
		t.Error("a reboot from somewhere other than the disc was accepted")
	}
}
