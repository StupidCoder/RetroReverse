package xbox

import (
	"path/filepath"
	"testing"

	"retroreverse.com/tools/cpu/x86"
)

// usb_test.go pins the OHCI register semantics that the latch got wrong.
//
// These are unit tests against a bare machine rather than a booted title, and that is
// the point: the guest that exercises this controller only does so once, ~1.2 billion
// instructions into a cold boot, which is far too slow a loop to develop against and far
// too coarse a signal to debug with. Every rule asserted here was read out of XAPI's own
// code (see usb.go's header for the disassembly), so these tests are that reading,
// written down in a form that fails.

// mmioMachine builds the smallest machine an MMIO test needs: RAM, a CPU (the latch's
// trace and log-once guard both ask it for a PC), and the device latches.
func mmioMachine(t *testing.T) *Machine {
	t.Helper()
	m := &Machine{RAM: make([]byte, ramSize)}
	m.apu = newMMIOLatch("APU")
	m.ac97 = newMMIOLatch("AC97")
	m.usb = newMMIOLatch("USB")
	m.nic = newMMIOLatch("NIC")
	m.interrupts = map[uint32]uint32{}
	m.CPU = x86.NewCPU(m) // the latch's trace and log-once guard both ask it for a PC
	return m
}

// rd32 / wr32 drive the aperture the way the guest does: a dword at a time, which the
// bus splits into the four byte calls the model actually sees.
func rd32(m *Machine, off uint32) uint32 {
	var v uint32
	for i := uint32(0); i < 4; i++ {
		v |= uint32(m.usbRead(off+i)) << (8 * i)
	}
	return v
}

func wr32(m *Machine, off, v uint32) {
	for i := uint32(0); i < 4; i++ {
		m.usbWrite(off+i, byte(v>>(8*i)))
	}
}

// TestInterruptEnableAccumulates is the regression for the bug that cost every
// pre-Phase-E savestate its pad. HcInterruptEnable is write-1-to-set, so XAPI's two
// separate enables must accumulate; the latch replaced, leaving the second write's 0x40
// with the master interrupt enable clear — a controller that can never signal.
func TestInterruptEnableAccumulates(t *testing.T) {
	m := mmioMachine(t)

	// The title's own sequence, from the trace: 0x240056 then 0x240161.
	wr32(m, hcInterruptEnable, ohciIntMIE|ohciIntFNO|ohciIntUE|ohciIntWDH|ohciIntSO)
	wr32(m, hcInterruptEnable, ohciIntRHSC)

	want := uint32(ohciIntMIE | ohciIntFNO | ohciIntUE | ohciIntWDH | ohciIntSO | ohciIntRHSC)
	if got := rd32(m, hcInterruptEnable); got != want {
		t.Errorf("enable mask = %08X, want %08X (a write-1-to-set register must not replace)", got, want)
	}
	// The disable window is a view of the same mask, not a register of its own.
	if got := rd32(m, hcInterruptDisable); got != want {
		t.Errorf("disable window reads %08X, want the enable mask %08X", got, want)
	}
}

// TestInterruptDisableClearsEnable pins the other half: a write to the disable window
// clears those bits in the one shared mask.
func TestInterruptDisableClearsEnable(t *testing.T) {
	m := mmioMachine(t)
	wr32(m, hcInterruptEnable, ohciIntMIE|ohciIntRHSC|ohciIntWDH)
	wr32(m, hcInterruptDisable, ohciIntWDH)

	want := uint32(ohciIntMIE | ohciIntRHSC)
	if got := rd32(m, hcInterruptEnable); got != want {
		t.Errorf("enable mask = %08X, want %08X", got, want)
	}
}

// TestInterruptStatusIsWriteOneToClear pins the acknowledgement. A latch here would
// have made an ISR's ack re-raise the very interrupt it was ending.
func TestInterruptStatusIsWriteOneToClear(t *testing.T) {
	m := mmioMachine(t)
	m.usb.reg[hcInterruptStatus>>2] = ohciIntRHSC | ohciIntWDH

	wr32(m, hcInterruptStatus, ohciIntRHSC) // the ISR acks the one it handled

	if got := rd32(m, hcInterruptStatus); got != ohciIntWDH {
		t.Errorf("status = %08X, want %08X (only the acked bit clears)", got, uint32(ohciIntWDH))
	}
}

// TestRootHubReportsItsPortCount is the NDP regression. XAPI writes NPS|NOCP into
// HcRhDescriptorA to configure power switching; a register that echoed that write back
// made the root hub answer "I have zero ports".
func TestRootHubReportsItsPortCount(t *testing.T) {
	m := mmioMachine(t)
	wr32(m, hcRhDescriptorA, 0x1200) // exactly what the trace shows XAPI writing

	got := rd32(m, hcRhDescriptorA)
	if got&0xFF != usbPorts {
		t.Errorf("HcRhDescriptorA NDP = %d, want %d (the port count is read-only hardware state)", got&0xFF, usbPorts)
	}
	if got&0x1200 != 0x1200 {
		t.Errorf("HcRhDescriptorA = %08X, lost the writable NPS|NOCP bits", got)
	}
}

// TestIRQNeedsMasterEnable pins why a restored pre-Phase-E snapshot is honestly silent:
// its enable mask is RHSC without MIE, and such a controller genuinely never signals.
func TestIRQNeedsMasterEnable(t *testing.T) {
	m := mmioMachine(t)
	m.usb.reg[hcInterruptEnable>>2] = ohciIntRHSC // the old latch's surviving value
	m.usb.reg[hcInterruptStatus>>2] = ohciIntRHSC

	if m.usbIRQ() {
		t.Error("asserted the interrupt line with MIE clear")
	}
	m.usb.reg[hcInterruptEnable>>2] |= ohciIntMIE
	if !m.usbIRQ() {
		t.Error("did not assert with MIE set and an enabled source pending")
	}
	// A source that is pending but not enabled is not an interrupt.
	m.usb.reg[hcInterruptEnable>>2] = ohciIntMIE | ohciIntWDH
	if m.usbIRQ() {
		t.Error("asserted for a pending source that is not enabled")
	}
}

// TestConnectRaisesRootHubStatusChange is the contract XAPI's port loop waits on:
// report a connection in CCS and raise RHSC, and the driver comes and looks.
func TestConnectRaisesRootHubStatusChange(t *testing.T) {
	m := mmioMachine(t)
	wr32(m, hcInterruptEnable, ohciIntMIE|ohciIntRHSC)

	m.usbSetPortConnected(0, true)

	st := rd32(m, hcRhPortStatus1)
	if st&portCCS == 0 {
		t.Error("port reports no device after a connect")
	}
	if st&portCSC == 0 {
		t.Error("port raised no connect-status change")
	}
	if !m.usbIRQ() {
		t.Error("a connect did not assert the interrupt line")
	}
}

// TestPortChangeBitsAreWriteOneToClear reproduces XAPI's own ack, derived from its port
// loop: it clears the LOW half of the status it read and writes the value back, which
// leaves the change bits set and is therefore only an acknowledgement if bits 16-31 are
// write-1-to-clear. If they latched instead, the change would survive its own ack and
// RHSC would re-fire for ever.
func TestPortChangeBitsAreWriteOneToClear(t *testing.T) {
	m := mmioMachine(t)
	m.usbSetPortConnected(0, true)

	st := rd32(m, hcRhPortStatus1)
	wr32(m, hcRhPortStatus1, st&0xFFFF0000) // AND WORD [..],$0000 then write back

	if got := rd32(m, hcRhPortStatus1); got&portCSC != 0 {
		t.Errorf("port status = %08X, the connect-status change survived its ack", got)
	}
	// The ack must not have unplugged the device: CCS is state, not a change flag.
	if got := rd32(m, hcRhPortStatus1); got&portCCS == 0 {
		t.Error("acking the change bits dropped the connection")
	}
}

// TestPortResetEnablesAConnectedPort pins the enumeration step between "a device is
// there" and "it can be talked to".
func TestPortResetEnablesAConnectedPort(t *testing.T) {
	m := mmioMachine(t)
	wr32(m, hcInterruptEnable, ohciIntMIE|ohciIntRHSC)
	m.usbSetPortConnected(0, true)
	// Ack the connect, so what follows can only come from the reset.
	wr32(m, hcRhPortStatus1, portCSC)
	wr32(m, hcInterruptStatus, ohciIntRHSC)

	wr32(m, hcRhPortStatus1, portWSetReset)

	st := rd32(m, hcRhPortStatus1)
	if st&portPES == 0 {
		t.Error("a reset port did not come out enabled")
	}
	if st&portPRS != 0 {
		t.Error("the port is still reporting a reset in progress")
	}
	if st&portPRSC == 0 {
		t.Error("the reset raised no reset-status change for the driver to ack")
	}
	// And the change must KNOCK. XAPI resets a port and then waits for the completion
	// interrupt, with a 100 ms timer as its retry — so a PRSC that sets silently makes
	// the driver reset the port, time out, and reset it again, for ever. Every step of
	// that loop is individually correct, which is why it needs a test of its own.
	if !m.usbIRQ() {
		t.Error("a completed port reset raised no RootHubStatusChange: the driver would retry for ever")
	}
}

// TestEmptyPortCannotBeEnabled: a port with nothing in it must refuse, or the driver
// would enumerate a device that is not there.
func TestEmptyPortCannotBeEnabled(t *testing.T) {
	m := mmioMachine(t)
	wr32(m, hcRhPortStatus1, portWSetEnable)
	if rd32(m, hcRhPortStatus1)&portPES != 0 {
		t.Error("enabled a port with no device connected")
	}
	wr32(m, hcRhPortStatus1, portWSetReset)
	if rd32(m, hcRhPortStatus1)&portPES != 0 {
		t.Error("a reset enabled a port with no device connected")
	}
}

// TestFrameNumberIsDerivedFromTheClock pins the savestate-stability rule: the frame
// number is a function of the machine tick, so it cannot drift when a snapshot skips
// the calls a counter would have needed.
func TestFrameNumberIsDerivedFromTheClock(t *testing.T) {
	m := mmioMachine(t)
	m.tick = 5000 * instrsPerMs
	first := rd32(m, hcFmNumber)
	if first != 5000&0xFFFF {
		t.Errorf("HcFmNumber = %d at tick %d, want %d", first, m.tick, 5000&0xFFFF)
	}
	// Reading it again — with no tick between — must not advance it.
	if again := rd32(m, hcFmNumber); again != first {
		t.Errorf("HcFmNumber advanced on a read (%d -> %d): it is a clock, not a counter", first, again)
	}
	// A restore that rewinds the clock rewinds the frame, rather than desynchronising.
	m.tick = 100 * instrsPerMs
	if got := rd32(m, hcFmNumber); got != 100 {
		t.Errorf("HcFmNumber = %d after the clock moved back, want 100", got)
	}
}

// TestUsbTickCatchesUpBounded pins the catch-up loop's bound. usbTick runs from
// schedTick's 1024-instruction block against a 2000-instruction frame, so it must
// service the frames it missed — but a savestate restored beside a clock that has moved
// on must not replay the entire gap.
func TestUsbTickCatchesUpBounded(t *testing.T) {
	m := mmioMachine(t)
	m.usbFrameServed = 1
	m.tick = 10_000 * instrsPerMs // a ten-second jump: the shape of a restore

	m.usbTick()

	if m.usbFrameServed != m.usbFrame() {
		t.Errorf("frame cursor = %d, want %d (usbTick must reach the present)", m.usbFrameServed, m.usbFrame())
	}
}

// --- the transfer engine (usb_ohci.go) ---
//
// The walker and the device landed together, because neither can be proven by the guest
// without the other: XAPI only queues a TD once a device answers, and a device can only
// answer down a walked list. These tests are the way out of that — a hand-built ED/TD
// chain in RAM is a guest that needs no booting, so the walker is falsifiable before
// there is anything to walk for.

// buildED writes an endpoint descriptor and its TD chain into RAM, returning the ED.
func buildED(m *Machine, ed, fa, endpoint uint32, tds []uint32) {
	m.write32(ed+0x00, fa|endpoint<<7|8<<16) // FA, EN, MPS 8 — the control pipe's shape
	m.write32(ed+0x0C, 0)                    // NextED: end of list
	if len(tds) == 0 {
		m.write32(ed+0x04, 0)
		m.write32(ed+0x08, 0)
		return
	}
	m.write32(ed+0x08, tds[0]) // HeadP
	// The queue ends BEFORE TailP, so the driver always leaves one spare TD there.
	m.write32(ed+0x04, tds[len(tds)-1])
}

// buildTD writes one transfer descriptor. dp is the direction field, already shifted.
func buildTD(m *Machine, td, dp, cbp, be, next uint32) {
	m.write32(td+0x00, dp)
	m.write32(td+0x04, cbp)
	m.write32(td+0x08, next)
	m.write32(td+0x0C, be)
}

// TestWalkerRetiresSetupAndInDataStage drives one whole control transfer the way XAPI
// does — SETUP then IN — through the walker, and checks the device's answer landed in
// the driver's own buffer with its TDs retired and chained onto the done queue.
func TestWalkerRetiresSetupAndInDataStage(t *testing.T) {
	m := mmioMachine(t)
	const (
		ed   = 0x2000
		td0  = 0x2100 // SETUP
		td1  = 0x2110 // IN (data)
		td2  = 0x2120 // the spare TailP TD
		pkt  = 0x2200 // the setup packet's buffer
		dest = 0x2300 // where the IN data must land
	)
	// A device that answers one request with known bytes. This is deliberately not the
	// XID: the walker must not know what is plugged into it.
	want := []byte{0x12, 0x34, 0x56, 0x78}
	m.usbDev[0] = &fakeDevice{addr: 0, in: want}

	buildTD(m, td0, 0, pkt, pkt+7, td1)        // DP=0 SETUP, 8 bytes
	buildTD(m, td1, tdDPIn, dest, dest+3, td2) // DP=IN, 4 bytes
	buildED(m, ed, 0, 0, []uint32{td0, td1, td2})

	m.usbWalkList(ed)

	if got := m.read32(td0) >> 28; got != ccNoError {
		t.Errorf("SETUP TD condition code = %X, want NoError", got)
	}
	if got := m.read32(td1) >> 28; got != ccNoError {
		t.Errorf("IN TD condition code = %X, want NoError", got)
	}
	for i, w := range want {
		if got := m.Read(dest + uint32(i)); got != w {
			t.Errorf("data byte %d = %02X, want %02X", i, got, w)
		}
	}
	// The endpoint's head must have advanced past both, to the spare.
	if head := m.read32(ed+0x08) & edPtrMask; head != td2 {
		t.Errorf("ED HeadP = %08X, want %08X (retired TDs must be unlinked)", head, td2)
	}
	if len(m.usbDone) != 2 {
		t.Fatalf("done queue has %d TDs, want 2", len(m.usbDone))
	}
}

// TestRetiredTDConditionCodeIsWritten is the trap this file's header warns about, and
// it is the one a passing pad would never reveal. The driver reads TDs back out of
// memory it zeroed itself, and NoError is 0 — so a walker that "succeeds" by leaving
// the field alone looks identical to one that never ran. The test pre-poisons the
// field, which is the only way to tell those apart.
func TestRetiredTDConditionCodeIsWritten(t *testing.T) {
	m := mmioMachine(t)
	const td = 0x3000
	m.write32(td+0x00, tdDPOut|(ccNotAccessed<<28)) // poisoned: NotAccessed
	m.write32(td+0x04, 0)
	m.write32(td+0x08, 0)
	m.write32(td+0x0C, 0)

	m.usbRetire(td, m.read32(td), 0, 0, 0, ccNoError)

	if got := m.read32(td) >> 28; got != ccNoError {
		t.Errorf("condition code = %X, want NoError written explicitly over the poison", got)
	}
}

// TestWalkerNAKsWhenTheDeviceHasNothing pins that a NAK is not a failure: the TD stays
// queued for a later frame rather than retiring with a made-up success.
func TestWalkerNAKsWhenTheDeviceHasNothing(t *testing.T) {
	m := mmioMachine(t)
	const (
		ed   = 0x4000
		td0  = 0x4100
		td1  = 0x4110
		dest = 0x4200
	)
	m.usbDev[0] = &fakeDevice{addr: 1, in: nil} // nothing to report
	buildTD(m, td0, tdDPIn, dest, dest+3, td1)
	buildED(m, ed, 1, 2, []uint32{td0, td1}) // endpoint 2: the interrupt pipe

	m.usbWalkList(ed)

	if head := m.read32(ed+0x08) & edPtrMask; head != td0 {
		t.Errorf("ED HeadP = %08X, want %08X (a NAK must leave the TD queued)", head, td0)
	}
	if len(m.usbDone) != 0 {
		t.Errorf("a NAK put %d TDs on the done queue, want 0", len(m.usbDone))
	}
}

// TestWalkerSkipsUnaddressedEndpoints: an ED whose function address nothing answers on
// must go unserviced, not retire.
func TestWalkerSkipsUnaddressedEndpoints(t *testing.T) {
	m := mmioMachine(t)
	const (
		ed  = 0x5000
		td0 = 0x5100
		td1 = 0x5110
	)
	m.usbDev[0] = &fakeDevice{addr: 1}
	buildTD(m, td0, tdDPIn, 0x5200, 0x5203, td1)
	buildED(m, ed, 7, 0, []uint32{td0, td1}) // address 7: nobody

	m.usbWalkList(ed)

	if len(m.usbDone) != 0 {
		t.Errorf("retired %d TDs for an address no device answers on", len(m.usbDone))
	}
}

// TestWalkerHonoursSkipAndHalt: both are the driver's way of parking an endpoint, and
// a controller that ran them anyway would be running transfers behind its driver's back.
func TestWalkerHonoursSkipAndHalt(t *testing.T) {
	m := mmioMachine(t)
	const (
		ed  = 0x6000
		td0 = 0x6100
		td1 = 0x6110
	)
	m.usbDev[0] = &fakeDevice{addr: 0, in: []byte{1, 2, 3, 4}}
	buildTD(m, td0, tdDPIn, 0x6200, 0x6203, td1)

	buildED(m, ed, 0, 0, []uint32{td0, td1})
	m.write32(ed+0x00, m.read32(ed)|edSkip)
	m.usbWalkList(ed)
	if len(m.usbDone) != 0 {
		t.Errorf("serviced a skipped endpoint (%d TDs retired)", len(m.usbDone))
	}

	buildED(m, ed, 0, 0, []uint32{td0, td1})
	m.write32(ed+0x08, m.read32(ed+0x08)|edHeadHalted)
	m.usbWalkList(ed)
	if len(m.usbDone) != 0 {
		t.Errorf("serviced a halted endpoint (%d TDs retired)", len(m.usbDone))
	}
}

// TestWritebackPublishesDoneHead: the driver reads the done chain out of the HCCA, so
// a writeback that raises WDH without publishing the chain reports successes that
// point nowhere.
func TestWritebackPublishesDoneHead(t *testing.T) {
	m := mmioMachine(t)
	const hcca = 0x7000
	m.usb.reg[hcHCCA>>2] = hcca
	wr32(m, hcInterruptEnable, ohciIntMIE|ohciIntWDH)
	m.usbDone = []uint32{0x7100, 0x7110}

	m.usbWriteback()

	// Newest first: the chain head is the last TD retired.
	if got := m.read32(hcca + 0x84); got != 0x7110 {
		t.Errorf("HCCA DoneHead = %08X, want %08X", got, 0x7110)
	}
	if got := m.read32(0x7110 + 0x08); got != 0x7100 {
		t.Errorf("done chain NextTD = %08X, want %08X", got, 0x7100)
	}
	if !m.usbIRQ() {
		t.Error("a writeback raised no WritebackDoneHead interrupt")
	}
}

// TestWalkerToleratesACyclicList: an ED list is guest data, and a cycle in it must cost
// a frame, not the machine.
func TestWalkerToleratesACyclicList(t *testing.T) {
	m := mmioMachine(t)
	const ed = 0x8000
	buildED(m, ed, 0, 0, nil)
	m.write32(ed+0x0C, ed) // NextED points at itself
	m.usbWalkList(ed)      // must return
}

// fakeDevice is a device that is not a pad: it exists so the walker's tests cannot
// accidentally be testing the XID's behaviour instead of the controller's.
type fakeDevice struct {
	addr     uint32
	in       []byte
	statuses int // how many control transfers have run their status stage
}

func (f *fakeDevice) address() uint32                                { return f.addr }
func (f *fakeDevice) setup(m *Machine, pkt []byte) ([]byte, error)   { return f.in, nil }
func (f *fakeDevice) interruptIn(m *Machine, endpoint uint32) []byte { return f.in }
func (f *fakeDevice) controlStatusDone(m *Machine)                   { f.statuses++ }

// TestIdleControllerRaisesNothing is E3's whole claim: a modelled controller with an
// empty root hub is more alive than a latch and behaves identically to one. If this
// fails, the title's boot changed.
func TestIdleControllerRaisesNothing(t *testing.T) {
	m := mmioMachine(t)
	wr32(m, hcInterruptEnable, ohciIntMIE|ohciIntRHSC)
	wr32(m, hcControl, 0xBE) // operational, all lists enabled — the title's own value

	for i := 0; i < 100; i++ {
		m.tick += instrsPerMs
		m.usbTick()
	}
	if m.usbIRQ() {
		t.Error("an operational controller with an empty root hub asserted an interrupt")
	}
	if got := rd32(m, hcRhPortStatus1); got != 0 {
		t.Errorf("an untouched port reads %08X, want 0", got)
	}
}

// TestStartOfFrameIsRaisedEveryFrame pins the interrupt that a frame boundary announces
// itself with. The frame number was always published; only the announcement was missing,
// and XAPI waits on it to retire an endpoint descriptor safely — so without this the pad
// enumerated, took its address, and stopped for ever.
//
// The status bit must be set whether or not anyone has enabled it. That is not pedantry:
// XAPI clears SF and only then enables it, which is a write that only makes sense against
// hardware that had been setting it all along.
func TestStartOfFrameIsRaisedEveryFrame(t *testing.T) {
	m := mmioMachine(t)
	wr32(m, hcControl, 0xBE) // operational — the title's own value
	// Deliberately do NOT enable SF: the status bit is a record of what happened.
	wr32(m, hcInterruptEnable, ohciIntMIE)

	// Prime the catch-up cursor from a non-zero frame: usbFrameServed uses 0 to mean
	// "never run", so a machine still on frame 0 cannot advance past it.
	m.tick = instrsPerMs
	m.usbTick()
	m.tick += instrsPerMs
	m.usbTick()

	if got := m.usb.reg[hcInterruptStatus>>2]; got&ohciIntSF == 0 {
		t.Fatalf("HcInterruptStatus = %08X, want StartofFrame set on a frame boundary", got)
	}
	if m.usbIRQ() {
		t.Error("SF pulled the interrupt line while it was not enabled")
	}
	// ...and once enabled, the standing SF is what XAPI would see the instant it unmasks.
	wr32(m, hcInterruptEnable, ohciIntSF)
	if !m.usbIRQ() {
		t.Error("an enabled, pending StartofFrame did not assert the interrupt line")
	}
}

// TestRetiredTDReportsZeroBufferWhenComplete pins the encoding a TD uses to say "all of
// it moved". Zero is not "nothing here" — it is the ONLY way to say "all", and XAPI reads
// it that way (0x24617E) to compute every transfer's byte count. A pointer one past the
// end is a legal address and arithmetically correct, and it means the opposite.
func TestRetiredTDReportsZeroBufferWhenComplete(t *testing.T) {
	const td, buf = 0x3000, 0x4000
	for _, c := range []struct {
		name          string
		moved, length uint32
		want          uint32
	}{
		{"whole buffer consumed", 8, 8, 0},
		{"short packet", 1, 8, buf + 1},
		{"nothing moved", 0, 8, buf},
		{"zero-length packet", 0, 0, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := mmioMachine(t)
			m.write32(td+0x04, 0xDEADBEEF) // poison: the field must be written, not left
			m.usbRetire(td, 0, buf, c.moved, c.length, ccNoError)
			if got := m.read32(td + 0x04); got != c.want {
				t.Errorf("CurrentBufferPointer = %08X, want %08X", got, c.want)
			}
		})
	}
}

// TestShortPacketHaltsTheEndpointAndAdvancesHead pins both halves of how a transfer ends
// early. XAPI queues one TD per packet and sets bufferRounding on the last one only, so a
// short packet on any earlier TD is DataUnderrun — condition code NINE, the value XAPI
// singles out (0x24604D) as "that is all the device had, and it is fine".
//
// The endpoint must halt, or the TDs behind it run against a device with nothing left to
// say. And the queue must still ADVANCE: XAPI has already accounted the failed TD and
// walks the queue from HeadP to dequeue what never ran.
func TestShortPacketHaltsTheEndpointAndAdvancesHead(t *testing.T) {
	m := mmioMachine(t)
	const (
		ed   = 0x2000
		td0  = 0x2100 // IN, expects 8, R=0
		td1  = 0x2110 // must NOT run
		td2  = 0x2120 // TailP
		dest = 0x2300
	)
	m.usbDev[0] = &fakeDevice{addr: 0, in: []byte{0xAA}} // one byte for an 8-byte TD

	buildTD(m, td0, tdDPIn, dest, dest+7, td1) // 8 bytes, bufferRounding clear
	buildTD(m, td1, tdDPIn, dest+8, dest+15, td2)
	// Poison the TD behind the halt: NotAccessed is what an untouched TD reads as, and
	// leaving it at zero would make "never ran" and "succeeded" the same bytes.
	m.write32(td1+0x00, m.read32(td1)|ccNotAccessed<<28)
	buildED(m, ed, 0, 0, []uint32{td0, td1, td2})

	m.usbWalkList(ed)

	// The LITERAL nine, not the constant. Asserting against ccDataUnderrun would compare
	// the constant with itself and pass for any value it was given — including the 0xD
	// this started as, which is BufferUnderrun and sent XAPI down its generic failure
	// path. The number is the fact here (XAPI's own CMP EAX,$9 at 0x24604D), so the test
	// has to say the number.
	if got := m.read32(td0) >> 28; got != 9 {
		t.Errorf("short IN TD condition code = %X, want 9 = DataUnderrun — the value XAPI "+
			"singles out at 0x24604D as a short packet rather than a failure", got)
	}
	head := m.read32(ed + 0x08)
	if head&edHeadHalted == 0 {
		t.Errorf("ED HeadP = %08X, want the Halted bit set after an errored TD", head)
	}
	if got := head & edPtrMask; got != td1 {
		t.Errorf("ED HeadP = %08X, want it advanced past the failed TD to %08X", got, td1)
	}
	if got := m.read32(td1) >> 28; got != ccNotAccessed {
		t.Errorf("TD behind the halt has code %X, want NotAccessed — it must not have run", got)
	}
}

// TestSetAddressTakesEffectOnlyAfterTheStatusStage pins the rule that the SETUP is not
// the end of the transfer. The status stage rides the same endpoint descriptor, addressed
// the OLD way, so a device that moves when the SETUP is parsed vanishes between two TDs
// of one transfer — and the driver waits for ever on a transfer that can no longer
// complete.
func TestSetAddressTakesEffectOnlyAfterTheStatusStage(t *testing.T) {
	m := mmioMachine(t)
	d := &xidDevice{}
	m.usbDev[0] = d

	if _, err := d.setup(m, []byte{0x00, usbReqSetAddress, 0x01, 0, 0, 0, 0, 0}); err != nil {
		t.Fatalf("SET_ADDRESS: %v", err)
	}
	if d.address() != 0 {
		t.Fatalf("address = %d immediately after the SETUP, want 0 — the transfer is not over", d.address())
	}
	if m.usbDeviceFor(0) == nil {
		t.Error("the device stopped answering on address 0 while its own status stage was still to run")
	}

	d.controlStatusDone(m)

	if d.address() != 1 {
		t.Errorf("address = %d after the status stage, want 1", d.address())
	}
	if m.usbDeviceFor(1) == nil {
		t.Error("the device does not answer on its new address")
	}
}

// TestDeviceDescriptorAnswersOnlyWhatXAPIChecks guards the descriptor against the one
// failure this phase is most able to hide: a field that is right because it was
// remembered. Each assertion below cites the comparison that forces it, and the synthetic
// fields are asserted to be ABSURD — if someone later "tidies" bDeviceSubClass to a
// real-looking value, that is a claim about a product, and this test calls it.
func TestDeviceDescriptorAnswersOnlyWhatXAPIChecks(t *testing.T) {
	d := deviceDescriptor
	if d[0] != 8 && d[0] != 18 {
		t.Errorf("bLength = %d, want 8 or 18 (0x2423B9/0x2423BE)", d[0])
	}
	if d[1] != 1 {
		t.Errorf("bDescriptorType = %d, want 1 (0x2423AA)", d[1])
	}
	if d[4] != 0 {
		t.Errorf("bDeviceClass = %02X, want 0: the only tag-0x81 driver in XAPI's table "+
			"(0x0023F3F4) is the hub at class 9, so any other non-zero class loses its port", d[4])
	}
	if d[7] > 0x40 {
		t.Errorf("bMaxPacketSize0 = %02X, want <= 0x40 (0x2423A2)", d[7])
	}
	// This assertion used to be the opposite: "want >= 0x20, because XAPI bounds the XID
	// input report size against this field (0x24407E)". That was wrong, and it is worth
	// leaving the correction in the test rather than only in the comment — the old
	// assertion PASSED, bit on every honest attempt to fix the value, and was defending a
	// misreading. 0x24407E's [ESI+6] belongs to the XID driver's own extension record, not
	// to the device object 0x2423CB writes; the byte it really bounds the report against is
	// the IN ENDPOINT's wMaxPacketSize, put there by 0x2443AF. Two objects, one register
	// name.
	//
	// What the field actually decides is the CONTROL pipe's packet size, and 8 is what XAPI
	// itself programmed into its own address-0 control ED before any device existed
	// (0x00080000, E2's trace). See usb_xid.go for the 0x40 failure this pins down.
	if d[7] != 8 {
		t.Errorf("bMaxPacketSize0 = %02X, want 8: it becomes the control ED's MPS, and 8 is "+
			"the packet size XAPI chose for its own default control pipe (0x00080000)", d[7])
	}
	if got, want := xidDescriptor[6], byte(xidReportSize); got != want {
		t.Errorf("XID input report size = %d, want %d: the report is bounded against the "+
			"ENDPOINT's wMaxPacketSize (0x24407E via 0x2443AF), not against bMaxPacketSize0", got, want)
	}
	if int(d[0]) != len(d) {
		t.Errorf("bLength = %d but the descriptor is %d bytes", d[0], len(d))
	}
}

// TestConfigBundleIsWalkableByXAPI walks the configuration bundle exactly the way XAPI
// does and asserts it finds what XAPI requires. The point is not that the bytes are the
// ones written down: it is that the SHAPE holds up under the guest's own walk, which is
// the only thing about a descriptor bundle the guest can actually check.
//
// XAPI's walk (0x24247C) steps by each descriptor's own bLength, so a single wrong length
// desynchronises everything after it while every individual field still looks right.
func TestConfigBundleIsWalkableByXAPI(t *testing.T) {
	c := configDescriptor[:]

	// 0x2426AE: wTotalLength must be <= 0x50 (XAPI's buffer) and must EQUAL the bytes
	// actually transferred (0x2426D8) — the one length in the enumeration a lying
	// transfer engine could not fool.
	total := int(c[2]) | int(c[3])<<8
	if total != len(c) {
		t.Errorf("wTotalLength = %d but the bundle is %d bytes (0x2426D8 compares it "+
			"against the transferred count)", total, len(c))
	}
	if total > 0x50 {
		t.Errorf("wTotalLength = %d, want <= 0x50, the size of XAPI's own buffer (0x2426B5)", total)
	}
	if c[1] != usbDescConfiguration {
		t.Errorf("bDescriptorType = %d, want %d", c[1], usbDescConfiguration)
	}
	if c[4] != 1 {
		t.Errorf("bNumInterfaces = %d, want exactly 1 (0x24249B)", c[4])
	}

	// The walk itself: step by bLength looking for bDescriptorType 4 (0x242495). A zero
	// bLength stops it dead (0x24248C), which is also how a desynced walk ends up.
	iface := -1
	for off := 0; off < total; {
		if c[off] == 0 {
			t.Fatalf("zero bLength at offset %d stops XAPI's walk (0x24248C)", off)
		}
		if off+1 < total && c[off+1] == usbDescInterface {
			iface = off
			break
		}
		off += int(c[off])
	}
	if iface < 0 {
		t.Fatal("XAPI's walk finds no INTERFACE descriptor in the bundle (0x242495)")
	}
	if c[iface+5] != xidInterfaceClass {
		t.Errorf("bInterfaceClass = %02X, want %02X", c[iface+5], xidInterfaceClass)
	}
	if xidInterfaceClass != 0x03 && xidInterfaceClass != 0x58 {
		t.Errorf("bInterfaceClass = %02X: XAPI's driver table claims only 0x03 and 0x58 at "+
			"tag 0x82 (and gives both the SAME driver); anything else finds no driver",
			xidInterfaceClass)
	}

	// The endpoint search (0x242A02) starts AFTER the interface and steps by bLength,
	// stopping at the next INTERFACE descriptor (0x242A67). It must find an endpoint that
	// is INTERRUPT (bmAttributes&3 == 3, 0x242A33) and IN (bEndpointAddress bit 7 set,
	// 0x242A46).
	ep := -1
	for off := iface + int(c[iface]); off < total; off += int(c[off]) {
		if c[off] == 0 || c[off+1] == usbDescInterface {
			break
		}
		if c[off+1] == usbDescEndpoint && c[off+3]&3 == 3 && c[off+2]&0x80 != 0 {
			ep = off
			break
		}
	}
	if ep < 0 {
		t.Fatal("XAPI's search (0x242A02) finds no INTERRUPT IN endpoint: the pad would " +
			"enumerate and then have no pipe to report on")
	}

	// 0x24407E bounds the XID input report size against this endpoint's wMaxPacketSize
	// lo — the byte 0x2443AF copies onto the extension. If the report does not fit its
	// own endpoint the pad is rejected at the last gate before it works.
	if int(c[ep+4]) < xidReportSize {
		t.Errorf("wMaxPacketSize lo = %d but the input report is %d bytes: 0x24407E rejects "+
			"a report larger than its endpoint's packet", c[ep+4], xidReportSize)
	}
}

// TestXIDDescriptorMeetsItsValidator asserts every bound the XID validator at 0x244011
// applies, by the numbers. "When the number is the fact, the test must say the number" —
// this phase has already shipped one test that compared a constant with itself and passed
// with the exact bug it was written to catch.
func TestXIDDescriptorMeetsItsValidator(t *testing.T) {
	x := xidDescriptor[:]
	if len(x) < 8 || x[0] < 8 {
		t.Errorf("bLength = %d (len %d), want >= 8 (0x244011)", x[0], len(x))
	}
	if int(x[0]) != len(x) {
		t.Errorf("bLength = %d but the descriptor is %d bytes", x[0], len(x))
	}
	if x[1] != 0x42 {
		t.Errorf("bDescriptorType = %02X, want 42 (0x24401E)", x[1])
	}
	if xidDescType != 0x42 {
		t.Errorf("xidDescType = %02X, want 42: it is also the wValue high byte XAPI asks "+
			"with, so the two cannot drift apart", xidDescType)
	}
	if x[2] == 0 && x[3] == 0 {
		t.Error("offsets 2..3 are a word XAPI requires to be NONZERO (0x24402B)")
	}
	if x[4] != 0x01 {
		t.Errorf("bType = %02X, want 01: the type object whose byte 0 matches this "+
			"(0x240AA4) must be 0x0023F4D4, whose +4 is 0x0023F49C — the struct the game "+
			"itself hands to XGetDeviceChanges at 0x39670", x[4])
	}
	if x[6] < 2 {
		t.Errorf("input report size = %d, want >= 2 (0x244073, and 0x2438E2 has nothing "+
			"to cook below 2)", x[6])
	}
	if x[6] > 0x20 {
		t.Errorf("input report size = %d, want <= 32 (0x244077)", x[6])
	}
	// The pad declares no interrupt OUT endpoint, so it must declare no output report:
	// 0x243E0A turns a zero here into a clean "not supported" (status 0x32), while a
	// nonzero would send XAPI down 0x243E1A to drive an endpoint that does not exist.
	// The validator itself cannot catch this — its bound at 0x24408A is guarded by the
	// OUT endpoint's address, which is zero here, so the check is skipped entirely.
	if x[7] != 0 {
		t.Errorf("output report size = %d, want 0: the bundle declares no interrupt OUT "+
			"endpoint, and 0x243E0A is what makes that coherent", x[7])
	}
}

// TestReportIsTheGamepadBehindATwoByteHeader pins the wire layout to XAPI's own copy.
//
// The cook at 0x2438DA moves size-2 bytes from the report buffer +2 (0x2438EF: LEA
// ESI,[EAX+$34], where the buffer is at +0x32) into XINPUT_GAMEPAD at +0x14. So the
// report IS the gamepad with a two-byte header in front of it, and the title's own
// reader (0x14630) says where each field lands.
func TestReportIsTheGamepadBehindATwoByteHeader(t *testing.T) {
	d := &xidDevice{}
	d.Buttons, d.Fresh = 0x0010, true // START: 0x14630's TEST CL,$10 -> OR EAX,$01
	r := d.report()

	if len(r) != xidReportSize || len(r) != 20 {
		t.Fatalf("report is %d bytes, want 20 (2 header + the 18-byte gamepad XAPI copies)", len(r))
	}
	// The gamepad is what survives the cook. Everything below indexes IT, not the report,
	// so the two-byte skew is stated once and then cannot be fudged.
	gp := r[2:]
	if len(gp) != 18 {
		t.Fatalf("the cook copies %d bytes into the gamepad, want 18 (the title reads "+
			"through +0x11 at 0x14788)", len(gp))
	}
	if got := uint16(gp[0]) | uint16(gp[1])<<8; got != 0x0010 {
		t.Errorf("gamepad wButtons = %04X, want 0010 — START, as the title reads it at "+
			"0x14630 (MOV CX,[ESI] / TEST CL,$10)", got)
	}
	// The eight analog buttons (gamepad+2..+9) and the four stick axes (+0xA..+0x11) are
	// LEVELS this pad is setting, and with only START held they are at rest. Asserting them
	// keeps the digital mask from leaking into a byte the title thresholds (0x147DE) or
	// reads as a signed axis (0x14680).
	for i := 2; i < 18; i++ {
		if gp[i] != 0 {
			t.Errorf("gamepad+%#x = %02X, want 0: only START is held, so every analog "+
				"button and every stick axis must be at rest", i, gp[i])
		}
	}

	// A poll with nothing new NAKs; XAPI must never see a report it did not earn.
	if got := d.interruptIn(nil, 1); got == nil {
		t.Error("the first poll after a level change must report, not NAK")
	}
	if got := d.interruptIn(nil, 1); got != nil {
		t.Errorf("a poll with no change must NAK (nil), got %d bytes", len(got))
	}
}

// TestReportCarriesTheAnalogButtonsAndSticks pins the two thirds of the report the pad could
// not say until this phase, at the offsets the title's own readers name.
//
// It is the test the old model could not have failed: a centred stick and an unmodelled stick
// produce identical bytes (usb_xid.go's own warning), and a released analog button and an
// unmodelled one do too. The whole pad was at rest, so every assertion about it passed while
// nothing was wired at all. What breaks that symmetry is DRIVING the fields — which is why
// this test sets each to a value it could not reach by accident.
func TestReportCarriesTheAnalogButtonsAndSticks(t *testing.T) {
	d := &xidDevice{}
	// A distinct pressure per byte, so a report that transposed two of them fails rather
	// than looking right. Each is >= 0x20, which is what both readers call pressed
	// (0x24390A zeroes below it; 0x147E5 asks > 0x1E).
	for i := range d.Analog {
		d.Analog[i] = byte(0x20 + i)
	}
	// Signed, and deliberately including a NEGATIVE and the extremes: the title reads these
	// with MOVSX (0x14680) and compares against negative thresholds, so a model that stored
	// them unsigned would pass every centred test and fail every real one.
	d.Axes = [4]int16{PadStickFull, -PadStickFull, -1, 0x1234}
	d.Fresh = true

	r := d.report()
	gp := r[2:]

	for i := range d.Analog {
		if gp[2+i] != byte(0x20+i) {
			t.Errorf("gamepad+%d = %02X, want %02X — the eight pressure bytes XAPI walks at "+
				"0x243906 land at gamepad+2..+9", 2+i, gp[2+i], 0x20+i)
		}
	}
	for i, want := range d.Axes {
		off := 0xA + 2*i
		got := int16(uint16(gp[off]) | uint16(gp[off+1])<<8)
		if got != want {
			t.Errorf("gamepad+%#x = %d, want %d — a stick axis is a little-endian SIGNED "+
				"word (0x14680: MOVSX ECX,[ESI+$A])", off, got, want)
		}
	}
}

// TestAnalogChangeIsNotNAKedAway is the bug the old Fresh/Sent pair would have shipped.
//
// They tracked Buttons alone, so a level change that moved only a pressure byte or a stick
// left Fresh clear and was NAKed away — the pad could be driven and the wire would never say
// so. The failure is invisible from every direction that matters: the device enumerates, the
// digital buttons work, and the one part of the pad the title reads as an ANALOG value is the
// one part that cannot change.
func TestAnalogChangeIsNotNAKedAway(t *testing.T) {
	m := gpuMachine(t)
	m.AttachPad(0)
	d := m.usbDev[0].(*xidDevice)

	for _, c := range []struct {
		name string
		set  func()
		want byte
		at   int
	}{
		{"an analog button", func() { m.SetPadAnalog(0, 3, PadPressed) }, PadPressed, 4 + 3},
		{"a stick axis", func() { m.SetPadAxis(0, 0, PadStickFull) }, byte(PadStickFull & 0xFF), 0xC},
	} {
		d.Fresh = false
		c.set()
		r := d.interruptIn(m, 1)
		if r == nil {
			t.Fatalf("moving %s NAKed: the report changed and the wire did not say so", c.name)
		}
		if r[c.at] != c.want {
			t.Errorf("%s: report[%d] = %02X, want %02X", c.name, c.at, r[c.at], c.want)
		}
		// ...and the same level twice is still not news.
		d.Fresh = true
		if r := d.interruptIn(m, 1); r != nil {
			t.Errorf("%s: a report identical to the last one must NAK, got %X", c.name, r)
		}
	}
}

// TestPadStateOfIsTheOneVocabulary covers the table both drivers of this machine share.
func TestPadStateOfIsTheOneVocabulary(t *testing.T) {
	// Every name resolves to something the report can actually carry. A control pointing at
	// an offset off the end of the pad would panic in SetPad, and only for whoever pressed
	// that one key.
	for _, n := range PadControlNames() {
		c, _ := PadControlByName(n)
		switch c.Kind {
		case PadAnalogButton:
			if c.Index < 0 || c.Index >= 8 {
				t.Errorf("%q: analog index %d is not one of the eight pressure bytes", n, c.Index)
			}
		case PadAxisDirection:
			if c.Index < 0 || c.Index >= 4 {
				t.Errorf("%q: axis index %d is not one of the four stick words", n, c.Index)
			}
			if c.Sign != +1 && c.Sign != -1 {
				t.Errorf("%q: sign %d is neither +1 nor -1", n, c.Sign)
			}
		case PadDigitalButton:
			if c.Bit == 0 {
				t.Errorf("%q: a digital button with no bit", n)
			}
		}
	}

	// The four derived pairings, as the on-screen keyboard showed them: each d-pad bit and
	// its stick direction moved the cursor the same way. That is not a claim the code can
	// check, but the DIRECTION each resolves to is.
	for _, c := range []struct {
		stick string
		axis  int
		sign  int16
	}{
		{"stickup", 1, +1}, {"stickdown", 1, -1},
		{"stickleft", 0, -1}, {"stickright", 0, +1},
	} {
		s := PadStateOf(map[string]bool{c.stick: true})
		if s.Axes[c.axis] != c.sign*PadStickFull {
			t.Errorf("%s: axes = %v, want axis %d at %d", c.stick, s.Axes, c.axis, c.sign*PadStickFull)
		}
		if s.Buttons != 0 || s.Analog != ([8]byte{}) {
			t.Errorf("%s: a stick direction pressed a button (%04X / %v)", c.stick, s.Buttons, s.Analog)
		}
	}

	// A held=false entry is not a press. The debugger's map deletes released keys, but the
	// oracle's schedule builds a map fresh each frame, and a caller that passed false would
	// otherwise be pressing.
	if s := PadStateOf(map[string]bool{"a": false, "start": false}); s != (PadState{}) {
		t.Errorf("PadStateOf with everything false = %+v, want a pad at rest", s)
	}
}

// TestUSBSaveState covers the root hub across a savestate. A pad is the only way into
// this title's menus, so every fixture from here on is taken with one plugged in — and
// until this state existed, such a restore came back with an empty hub.
//
// It mutation-tests BOTH directions. "Does it resume identically" cannot see aliasing:
// a snapshot that shares the machine's device pointer resumes perfectly and is still not
// a snapshot. So the test writes to each side after the copy and demands the other side
// not move.
func TestUSBSaveState(t *testing.T) {
	m := gpuMachine(t)
	m.AttachPad(0)
	m.SetPadButtons(0, 0x0010) // START
	// The analog half of the level rides the same snapshot, and it is only free because the
	// fields are EXPORTED — state.go copies [usbPorts]xidDevice by value, so an unexported
	// pressure byte would gob to zero and a restored pad would come back with its analog
	// buttons released and its sticks centred. Which is exactly the failure this port's own
	// warning says is invisible: a centred stick and a lost stick produce identical bytes.
	m.SetPadAnalog(0, 3, PadPressed)
	m.SetPadAxis(0, 0, PadStickFull)
	m.SetPadAxis(0, 1, -PadStickFull)
	d := m.usbDev[0].(*xidDevice)
	d.Addr, d.Config = 3, 1
	m.usbDone = []uint32{0x7100, 0x7110}
	m.usbCtrlData, m.usbCtrlOff = []byte{0xAA, 0xBB, 0xCC}, 1

	st := m.SaveState()
	// Through the DISK path, not just the in-memory struct: an empty port is a nil
	// element that gob refuses to encode, so a hub that survives LoadState(SaveState())
	// can still fail to survive a file. Written before the mutation below, so the file
	// holds the values the assertions expect.
	path := filepath.Join(t.TempDir(), "usb.state")
	if err := m.SaveStateFile(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Mutating the LIVE machine must not reach the snapshot.
	m.SetPadButtons(0, 0x0020)
	m.SetPadAnalog(0, 3, 0)
	m.SetPadAxis(0, 0, 0)
	d.Addr = 9
	m.usbDone[0] = 0xDEAD
	m.usbCtrlData[0] = 0xFF
	if st.UsbDev[0].Buttons != 0x0010 || st.UsbDev[0].Addr != 3 {
		t.Errorf("snapshot aliases the live pad: buttons=%04X addr=%d",
			st.UsbDev[0].Buttons, st.UsbDev[0].Addr)
	}
	if st.UsbDev[0].Analog[3] != PadPressed || st.UsbDev[0].Axes[0] != PadStickFull {
		t.Errorf("snapshot aliases the live pad's analog level: analog[3]=%02X axes[0]=%d",
			st.UsbDev[0].Analog[3], st.UsbDev[0].Axes[0])
	}
	if st.UsbDone[0] != 0x7100 || st.UsbCtrlData[0] != 0xAA {
		t.Errorf("snapshot aliases the live transfer state: done0=%08X ctrl0=%02X",
			st.UsbDone[0], st.UsbCtrlData[0])
	}

	m2 := gpuMachine(t)
	if err := m2.LoadStateFile(path); err != nil {
		t.Fatalf("load: %v", err)
	}
	d2, ok := m2.usbDev[0].(*xidDevice)
	if !ok || d2 == nil {
		t.Fatal("restored machine has no pad on port 0")
	}
	if d2.Buttons != 0x0010 || d2.Addr != 3 || d2.Config != 1 {
		t.Errorf("restored pad = buttons %04X addr %d config %d, want 0010/3/1",
			d2.Buttons, d2.Addr, d2.Config)
	}
	if d2.Analog[3] != PadPressed || d2.Axes[0] != PadStickFull || d2.Axes[1] != -PadStickFull {
		t.Errorf("restored pad's analog level = analog[3] %02X axes %v, want %02X and "+
			"[%d %d 0 0] — through the FILE, so an unexported field would show here",
			d2.Analog[3], d2.Axes, PadPressed, PadStickFull, -PadStickFull)
	}
	if len(m2.usbDone) != 2 || m2.usbDone[0] != 0x7100 {
		t.Errorf("restored usbDone = %v", m2.usbDone)
	}
	if string(m2.usbCtrlData) != "\xAA\xBB\xCC" || m2.usbCtrlOff != 1 {
		t.Errorf("restored control transfer = %X off %d", m2.usbCtrlData, m2.usbCtrlOff)
	}
	for i := 1; i < usbPorts; i++ {
		if m2.usbDev[i] != nil {
			t.Errorf("port %d should be empty after restore", i)
		}
	}

	// Mutating the RESTORED machine must not reach the snapshot either.
	m3 := gpuMachine(t)
	if err := m3.LoadState(st); err != nil {
		t.Fatal(err)
	}
	m3.usbDev[0].(*xidDevice).Addr = 7
	m3.SetPadAxis(0, 0, 0x0123)
	m3.SetPadAnalog(0, 3, 0x7F)
	m3.usbDone[0] = 0xBEEF
	m3.usbCtrlData[0] = 0x11
	if st.UsbDev[0].Addr != 3 || st.UsbDone[0] != 0x7100 || st.UsbCtrlData[0] != 0xAA {
		t.Error("restore aliases the snapshot: a second LoadState would not see the saved values")
	}
	if st.UsbDev[0].Axes[0] != PadStickFull || st.UsbDev[0].Analog[3] != PadPressed {
		t.Errorf("restore aliases the snapshot's analog level: axes[0]=%d analog[3]=%02X",
			st.UsbDev[0].Axes[0], st.UsbDev[0].Analog[3])
	}
}
