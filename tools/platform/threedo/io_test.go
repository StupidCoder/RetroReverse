package threedo

import "testing"

// putBytes writes a byte slice into DRAM at addr for test setup.
func putBytes(m *Machine, addr uint32, b []byte) {
	for i, c := range b {
		m.Write(addr+uint32(i), c)
	}
}

// TestKprintf checks the kernel printf HLE formats the conversions the boot code
// uses and pulls arguments from r1..r3 then the stack.
func TestKprintf(t *testing.T) {
	m := NewMachine()
	const fmtAddr, strAddr = 0x1000, 0x1100
	putBytes(m, fmtAddr, []byte("n=%d s=%s x=%X%%\n\x00"))
	putBytes(m, strAddr, []byte("hi\x00"))
	m.CPU.SetReg(0, fmtAddr)
	m.CPU.SetReg(1, 0xFFFFFFFF) // %d -> -1
	m.CPU.SetReg(2, strAddr)    // %s -> "hi"
	m.CPU.SetReg(3, 0xBEEF)     // %X -> BEEF
	m.kprintf()
	got := m.TTY()
	want := "n=-1 s=hi x=BEEF%\n"
	if got != want {
		t.Fatalf("kprintf = %q, want %q", got, want)
	}
}

// TestTagArg checks TagArg-list walking for both a present and an absent tag.
func TestTagArg(t *testing.T) {
	m := NewMachine()
	const p = 0x2000
	// {tag 0xB: 0x1234, tag 0xA: 0x5678, TAG_END}
	for i, w := range []uint32{0xB, 0x1234, 0xA, 0x5678, 0, 0} {
		m.write32(p+uint32(i)*4, w)
	}
	if v := m.tagArg(p, 0xB); v != 0x1234 {
		t.Errorf("tagArg(0xB) = 0x%X, want 0x1234", v)
	}
	if v := m.tagArg(p, 0xA); v != 0x5678 {
		t.Errorf("tagArg(0xA) = 0x%X, want 0x5678", v)
	}
	if v := m.tagArg(p, 0xC); v != 0 {
		t.Errorf("tagArg(missing) = 0x%X, want 0", v)
	}
}

// TestDiskStreamReadSeek checks the File-folio stream read/seek/close logic
// against an injected in-memory file (no disc image needed).
func TestDiskStreamReadSeek(t *testing.T) {
	m := NewMachine()
	data := []byte("ABCDEFGHIJ")
	h := m.dheap.alloc(0x20)
	m.streams[h] = &diskStream{name: "test", data: data}

	const buf = 0x3000
	// Read 4 bytes from the start.
	if n := m.readDiskStream(h, buf, 4); n != 4 {
		t.Fatalf("read = %d, want 4", n)
	}
	if got := string([]byte{m.Read(buf), m.Read(buf + 1), m.Read(buf + 2), m.Read(buf + 3)}); got != "ABCD" {
		t.Fatalf("read data = %q, want ABCD", got)
	}
	// Seek to offset 8 (SEEK_SET) and read to EOF (2 bytes available).
	if p := m.seekDiskStream(h, 8, seekSet); p != 8 {
		t.Fatalf("seek = %d, want 8", p)
	}
	if n := m.readDiskStream(h, buf, 100); n != 2 {
		t.Fatalf("read at EOF-2 = %d, want 2", n)
	}
	if got := string([]byte{m.Read(buf), m.Read(buf + 1)}); got != "IJ" {
		t.Fatalf("tail read = %q, want IJ", got)
	}
	// A further read returns 0 (EOF).
	if n := m.readDiskStream(h, buf, 10); n != 0 {
		t.Fatalf("read past EOF = %d, want 0", n)
	}
	// A bad handle returns -1.
	if n := m.readDiskStream(0xDEAD, buf, 10); n != -1 {
		t.Fatalf("read bad handle = %d, want -1", n)
	}
	m.closeDiskStream(h)
	if _, ok := m.streams[h]; ok {
		t.Fatal("close did not remove the stream")
	}
}

// TestIOCompletionSignal checks that serviceIO delivers a completion signal to
// the owning task (SIGF_IODONE when the IOReq has no reply port).
func TestIOCompletionSignal(t *testing.T) {
	m := NewMachine()
	it := &item{num: 0x2000, typ: 0x10E, owner: bootTaskNum, addr: m.dheap.alloc(0x100)}
	m.items[it.num] = it

	const info = 0x4000
	m.Write(info+ioiCommand, cmdStatus) // a command performIO acknowledges with no data
	m.CPU.SetReg(0, uint32(it.num))
	m.CPU.SetReg(1, info)
	m.serviceIO(m.CPU, true)

	if got := m.read32(it.addr + ioFlagsOff); got&ioDone == 0 {
		t.Errorf("io_Flags = 0x%X, IO_DONE not set", got)
	}
	if m.taskByNum(bootTaskNum).sig&sigfIODONE == 0 {
		t.Errorf("owner task did not receive SIGF_IODONE (sig=0x%X)", m.taskByNum(bootTaskNum).sig)
	}
}

// TestFieldWaitPacing checks the PaceFields path: a WaitVBL (timer command 3 via
// SendIO) parks instead of completing in the submit call, then completes only when
// the field clock reaches the target — and it signals the task that SUBMITTED the
// request, not the IOReq's creator. That submitter-not-owner rule is what lets the
// race's asset loader WaitVBL on a global timer IOReq owned by another task without
// hanging.
func TestFieldWaitPacing(t *testing.T) {
	m := NewMachine()
	m.PaceFields = true

	// A "timer" device and an IOReq bound to it, OWNED by the boot task #1.
	dev := &item{num: 0x1F00, typ: 0x10F, name: "timer"}
	m.items[dev.num] = dev
	ioReq := &item{num: 0x2000, typ: 0x10E, owner: bootTaskNum, device: dev.num, addr: m.dheap.alloc(0x100)}
	m.items[ioReq.num] = ioReq

	// A second task submits the WaitVBL on that IOReq (as the race loader does on
	// the front-end's global timer IOReq).
	submitter := &task{num: 0x9000, state: stRunning}
	m.tasks = append(m.tasks, submitter)
	m.cur = len(m.tasks) - 1

	const info = 0x4000
	m.Write(info+ioiCommand, timerCmdWaitField)
	m.write32(info+ioiOffset, 3) // wait 3 fields
	m.CPU.SetReg(0, uint32(ioReq.num))
	m.CPU.SetReg(1, info)
	startVBL := m.vblank
	m.serviceIO(m.CPU, true) // SendIO

	// It must have parked, not completed.
	if got := m.read32(ioReq.addr + ioFlagsOff); got&ioDone != 0 {
		t.Fatalf("field-wait completed in the submit call (io_Flags=0x%X); it must block", got)
	}
	if len(m.fieldWaits) != 1 {
		t.Fatalf("field-wait not parked (fieldWaits=%d)", len(m.fieldWaits))
	}

	// Two fields is not enough (target is +3); three is.
	m.fieldTick()
	m.fieldTick()
	if m.read32(ioReq.addr+ioFlagsOff)&ioDone != 0 {
		t.Fatalf("field-wait completed after 2 of 3 fields")
	}
	m.fieldTick()
	if got := m.read32(ioReq.addr + ioFlagsOff); got&ioDone == 0 {
		t.Fatalf("field-wait not done after 3 fields (io_Flags=0x%X)", got)
	}
	if m.vblank != startVBL+3 {
		t.Errorf("field clock advanced to %d, want %d", m.vblank, startVBL+3)
	}
	// The SUBMITTER got SIGF_IODONE; the IOReq's owner (boot task) did NOT.
	if submitter.sig&sigfIODONE == 0 {
		t.Errorf("submitter did not receive SIGF_IODONE (sig=0x%X)", submitter.sig)
	}
	if m.taskByNum(bootTaskNum).sig&sigfIODONE != 0 {
		t.Errorf("IOReq owner wrongly received SIGF_IODONE — it must go to the submitter")
	}
	if len(m.fieldWaits) != 0 {
		t.Errorf("completed field-wait not removed (fieldWaits=%d)", len(m.fieldWaits))
	}
}
