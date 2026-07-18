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

// TestCompleteIOReplyPort checks that an IOReq carrying a reply port is delivered
// AS A MESSAGE on completion — queued on the port (retrievable with GetMsg) with
// its own item number stamped in msg_DataPtr — not merely signalled. The
// DataStreamer's DataAcq reaps completed reads this way and keys each one to its
// buffer by that item number; delivering only the signal left its GetMsg empty so
// it rejected every completion as an "unknown i/o reply message" and never
// recycled a buffer, stalling the movie after the first ring-fill.
func TestCompleteIOReplyPort(t *testing.T) {
	m := NewMachine()

	const portSig = 0x200
	port := &item{num: 0x1067, typ: 0x10A, owner: bootTaskNum, signal: portSig}
	m.items[port.num] = port
	io := &item{num: 0x104E, typ: 0x10E, owner: bootTaskNum,
		addr: m.iheap.alloc(0x80), replyPort: port.num}
	m.items[io.num] = io

	// Blocked owner waiting on the port signal, as the DataAcq's main loop is.
	m.taskByNum(bootTaskNum).state = stWaiting
	m.taskByNum(bootTaskNum).wait = portSig

	m.completeIO(io)

	if len(port.msgs) != 1 || port.msgs[0] != io.num {
		t.Fatalf("reply port queue = %v, want [%d]", port.msgs, io.num)
	}
	if got := m.read32(io.addr + msgDataPtr); got != uint32(io.num) {
		t.Errorf("msg_DataPtr = 0x%X, want the IOReq item number 0x%X", got, io.num)
	}
	if m.taskByNum(bootTaskNum).sig&portSig == 0 &&
		m.taskByNum(bootTaskNum).state != stReady {
		t.Errorf("port owner not woken by the completion")
	}
}

// TestMessagePortPreempt checks that posting a message to a port whose owner is
// blocked waiting on that port's signal yields to the owner: the Portfolio kernel
// switches to the runnable server task the instant a request is posted, so a
// client that builds its request on its own stack keeps it live until the server
// reads it. Without the yield the client runs on and overwrites the request (this
// is why the DataStreamer rejected every movie request as a bad opcode).
func TestMessagePortPreempt(t *testing.T) {
	m := NewMachine() // boot task #1 is the current task (the "client")
	client := m.curTask()

	// A server task blocked waiting on the port's signal.
	const portSig = 0x100
	server := &task{num: 0x5000, state: stWaiting, wait: portSig}
	m.tasks = append(m.tasks, server)

	port := &item{num: 0x1053, typ: 0x10A, name: "ds", owner: server.num, signal: portSig}
	m.items[port.num] = port
	msg := &item{num: 0x1043, typ: typeMsg, owner: client.num, addr: m.dheap.alloc(0x40)}
	m.items[msg.num] = msg

	m.CPU.SetReg(0, uint32(port.num))
	m.CPU.SetReg(1, uint32(msg.num))
	m.CPU.SetReg(2, 0x6000) // data ptr
	m.CPU.SetReg(3, 0x1C)   // data size
	m.serviceMsg(m.CPU, swiPutMsg)

	if server.state != stReady {
		t.Errorf("server not woken (state=%d)", server.state)
	}
	if !m.needSchedule {
		t.Errorf("posting to a waiting server did not request a reschedule")
	}
	if client.state != stReady {
		t.Errorf("client not left Ready for resume (state=%d)", client.state)
	}
	if got := server.ctx.R[0]; got != portSig {
		t.Errorf("server WaitSignal result = 0x%X, want 0x%X", got, portSig)
	}

	// Posting again when the server is already runnable must NOT force a yield.
	m.needSchedule = false
	m.serviceMsg(m.CPU, swiPutMsg)
	if m.needSchedule {
		t.Errorf("posting to an already-ready server should not reschedule")
	}
}

// TestWaitPort covers the blocking WaitPort folio call (kernel -0x60): with a
// matching message already queued it dequeues and returns it in one shot; with an
// empty port it blocks the caller (retry-on-resume) and, crucially, a later reply
// wakes it WITHOUT clobbering its still-live argument registers (folioWait), so the
// re-run reads the same port/msg and returns the message.
func TestWaitPort(t *testing.T) {
	m := NewMachine()
	client := m.curTask()

	const portSig = 0x200
	port := &item{num: 0x1200, typ: 0x10A, name: "reply", owner: client.num, signal: portSig}
	m.items[port.num] = port
	msg := &item{num: 0x1043, typ: typeMsg, owner: client.num, addr: m.dheap.alloc(0x40)}
	m.items[msg.num] = msg

	// Case 1: message already queued -> return it immediately, advance PC to LR.
	port.msgs = []int32{msg.num}
	m.CPU.SetReg(0, uint32(port.num))
	m.CPU.SetReg(1, uint32(msg.num))
	m.CPU.SetReg(14, 0x40000) // LR
	if !m.serviceFolio(0x60) {
		t.Fatal("serviceFolio(0x60) not handled")
	}
	if got := m.CPU.Reg(0); got != uint32(msg.num) {
		t.Errorf("WaitPort result = 0x%X, want 0x%X", got, msg.num)
	}
	if got := m.CPU.Reg(15); got != 0x40000 {
		t.Errorf("WaitPort did not return to LR (pc=0x%X)", got)
	}
	if len(port.msgs) != 0 {
		t.Errorf("WaitPort did not dequeue the message (queue=%v)", port.msgs)
	}
	if client.folioWait {
		t.Errorf("folioWait should be clear after an immediate return")
	}

	// Case 2: empty port -> block. PC must stay at the trampoline (not LR) so the
	// run loop re-dispatches on resume; the task waits on the port signal.
	m.needSchedule = false
	m.CPU.SetReg(0, uint32(port.num))
	m.CPU.SetReg(1, uint32(msg.num))
	m.CPU.SetReg(15, hleBase+0x60) // the WaitPort trampoline PC
	m.CPU.SetReg(14, 0x40000)
	m.serviceFolio(0x60)
	if !m.needSchedule {
		t.Errorf("WaitPort on an empty port did not request a reschedule")
	}
	if client.state != stWaiting || client.wait != portSig {
		t.Errorf("client not blocked on the port signal (state=%d wait=0x%X)", client.state, client.wait)
	}
	if !client.folioWait {
		t.Errorf("blocked WaitPort should set folioWait")
	}
	if got := m.CPU.Reg(15); got != hleBase+0x60 {
		t.Errorf("blocked WaitPort advanced PC (0x%X); it must retry", got)
	}

	// A reply arrives: it must wake the client and NOT overwrite r0/r1 (its args).
	// The message names the client's port as its reply port, as the sender set.
	m.write32(msg.addr+msgReplyPort, uint32(port.num))
	client.ctx = m.CPU.SaveContext()
	m.replyMsg(msg, 0)
	if client.state != stReady {
		t.Errorf("reply did not wake the blocked WaitPort (state=%d)", client.state)
	}
	if got := client.ctx.R[0]; got != uint32(port.num) {
		t.Errorf("sendSignal clobbered WaitPort's port arg r0 = 0x%X, want 0x%X", got, port.num)
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
