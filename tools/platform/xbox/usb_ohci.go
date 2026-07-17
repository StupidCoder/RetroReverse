package xbox

import "fmt"

// usb_ohci.go is the host controller's transfer engine: the part that walks the
// endpoint and transfer descriptors the driver builds in guest RAM and actually moves
// bytes. usb.go is the register face; this is what the registers command.
//
// Nothing here is game-specific, and almost nothing here is ours. An OHCI controller's
// job is to read structures the driver wrote, in memory the driver allocated, and write
// results back where the driver will look — so this walks XAPI's own descriptors and
// hands XAPI's own memory to a device model. That is as low-level as this repo's
// emulation gets, and it is why the descriptor formats below are the OHCI spec's rather
// than anything derived: they are the format of data the guest wrote, not a claim about
// what the guest wants.
//
// THE ONE RULE THAT IS NOT SPEC. A retired TD must have its condition code written
// explicitly, every time, including on success. The driver reads its TDs back out of
// memory it zeroed itself, and NoError is zero — so a controller that "returns success"
// by leaving the field alone is indistinguishable from a controller that never ran. The
// same trap sits under the done queue: publishing a done head without having written
// the codes it points at reports a batch of successes that never happened.

// OHCI endpoint descriptor (16 bytes, dword-aligned), as the driver lays it out.
//
//	+0  control: FA[0:6] EN[7:10] D[11:12] S[13] K[14] F[15] MPS[16:26]
//	+4  TailP:   the TD the queue ends BEFORE (an empty queue is Head == Tail)
//	+8  HeadP:   H[0] halted, C[1] toggle carry, the head TD in [4:31]
//	+C  NextED
const (
	edFA   = 0x7F
	edEN   = 0xF << 7
	edD    = 3 << 11
	edSkip = 1 << 14
	edISO  = 1 << 15
	edMPS  = 0x7FF << 16

	edHeadHalted = 1 << 0
	edHeadToggle = 1 << 1
	edPtrMask    = ^uint32(0xF)
)

// OHCI general transfer descriptor (16 bytes).
//
//	+0  control: R[18] DP[19:20] DI[21:23] T[24:25] EC[26:27] CC[28:31]
//	+4  CBP: current buffer pointer (0 = a zero-length packet)
//	+8  NextTD
//	+C  BE: the LAST byte's address, not one past it
const (
	tdDP    = 3 << 19
	tdDPOut = 1 << 19
	tdDPIn  = 2 << 19

	// tdRounding is R: whether a packet shorter than the TD's buffer is acceptable. It
	// is how a driver says where a transfer is allowed to end. XAPI packetises a control
	// transfer by hand — ten 8-byte TDs for an 80-byte descriptor — and sets R on the
	// LAST one only, which is a statement about the device: every packet before the end
	// must be full, and a short one means the device has run out of data to give.
	tdRounding = 1 << 18
	// tdDPSetup is 0: a SETUP packet is DP == 0, which is why the direction must be
	// read as a field and never tested as a flag.
	tdCC = 0xF << 28
)

// TD condition codes (the CC field, written on retirement).
//
// ccDataUnderrun is NINE, and the guest is what says so rather than a memory of the
// table. It was 0xD here — BufferUnderrun, a host-side fault — which is a plausible name
// for the wrong thing: a short packet is the DEVICE giving less than was asked for, not
// the host running dry. XAPI reads the code out of the TD and singles that one value out
// by number:
//
//	00246043  MOV EAX, [EDX]      the retired TD's control word
//	0024604A  SHR EAX, $1C        the condition code
//	0024604D  CMP EAX, $00000009  is it a short packet...
//	0024605E  JNZ $002460AC       ...no -> the generic failure path, byte count discarded
//	00246066  MOV EAX, [EDX+$4]   ...yes -> count what DID arrive, from CBP and BE
//
// So nine is the only code that means "that is all the device had, and it is fine". At
// 0xD the pad's descriptor arrived, byte for byte, into a driver that then threw the
// count away and rejected the configuration for a length it had itself refused to add up.
const (
	ccNoError      = 0x0
	ccStall        = 0x4
	ccDataUnderrun = 0x9
	ccNotAccessed  = 0xE
)

// usbEndpointDir is a transfer's direction, from the device's point of view.
type usbEndpointDir int

const (
	dirSetup usbEndpointDir = iota
	dirIn                   // device -> host
	dirOut                  // host -> device
)

// usbDevice is what hangs off a root hub port.
//
// It is deliberately narrow: the controller does not know what an XID pad is, only that
// something on an address answers control setups and produces interrupt-endpoint data.
// The split is the hardware's own, and it keeps the pad's shape (usb_xid.go) out of the
// transfer engine entirely.
type usbDevice interface {
	// setup begins a control transfer. The 8 bytes are the guest's own SETUP packet.
	// It returns the data the device will return in the transfer's IN phase (nil for
	// none), or an error to STALL the endpoint.
	setup(m *Machine, pkt []byte) ([]byte, error)

	// interruptIn is polled for the device's interrupt endpoint. It returns the report
	// to deliver, or nil to NAK (nothing new).
	interruptIn(m *Machine, endpoint uint32) []byte

	// controlStatusDone reports that a control transfer's status stage has retired —
	// the transfer is over, and the device may now act on what it was asked to do.
	//
	// A request whose effect is visible to the BUS cannot take that effect when the
	// SETUP is parsed, because the transfer is not finished: its status stage still has
	// to be answered, and it rides the same endpoint descriptor, addressed the old way.
	// SET_ADDRESS is the request that proves it (see usb_xid.go).
	controlStatusDone(m *Machine)

	// address is the USB address the device currently answers on: 0 until the driver
	// has assigned one with SET_ADDRESS.
	address() uint32
}

// usbDeviceFor finds the device answering on a USB function address, across the ports.
func (m *Machine) usbDeviceFor(fa uint32) usbDevice {
	for i := range m.usbDev {
		d := m.usbDev[i]
		if d != nil && d.address() == fa {
			return d
		}
	}
	return nil
}

// usbWalkList walks one ED list, running the transfers each endpoint has queued.
//
// The walk is bounded rather than trusting the driver's NextED chain to terminate: a
// list is guest data, and a cycle in it must stop this frame rather than the machine.
func (m *Machine) usbWalkList(head uint32) {
	for ed, n := head&edPtrMask, 0; ed != 0 && n < 64; n++ {
		next := m.read32(ed+0x0C) & edPtrMask
		m.usbRunEndpoint(ed)
		ed = next
	}
}

// usbRunEndpoint drains one endpoint's TD queue.
func (m *Machine) usbRunEndpoint(ed uint32) {
	ctrl := m.read32(ed + 0x00)
	if ctrl&edSkip != 0 || ctrl&edISO != 0 {
		return // skipped, or isochronous (nothing here streams audio over USB)
	}
	head := m.read32(ed + 0x08)
	if head&edHeadHalted != 0 {
		return // the driver must clear the halt itself, after it has looked
	}
	tail := m.read32(ed+0x04) & edPtrMask

	for n := 0; n < 64; n++ {
		td := head & edPtrMask
		if td == 0 || td == tail {
			return // an empty queue: Head == Tail is the idle endpoint, not an error
		}
		nextTD, done := m.usbRunTD(ed, ctrl, td)
		if !done {
			return // the device had nothing to say: leave the TD queued for a later frame
		}
		// Retire: the endpoint's head advances past the TD, keeping the toggle carry.
		head = nextTD | (head & edHeadToggle)
		m.write32(ed+0x08, head)
	}
}

// usbRunTD attempts one transfer descriptor. It reports the next TD and whether this
// one retired; a device with nothing to send leaves the TD in place and returns false,
// which is a NAK — the transfer is still pending, not failed.
func (m *Machine) usbRunTD(ed, edCtrl, td uint32) (next uint32, retired bool) {
	ctrl := m.read32(td + 0x00)
	cbp := m.read32(td + 0x04)
	nextTD := m.read32(td+0x08) & edPtrMask
	be := m.read32(td + 0x0C)

	fa := edCtrl & edFA
	endpoint := (edCtrl & edEN) >> 7
	dev := m.usbDeviceFor(fa)
	if dev == nil {
		// No device answers on this address. That is not an error to invent a code for:
		// on real hardware the packets simply go unanswered until the driver times out.
		return nextTD, false
	}

	// The buffer is [CBP, BE], inclusive of BE — a zero CBP is a zero-length packet.
	var length uint32
	if cbp != 0 {
		length = be - cbp + 1
	}

	switch ctrl & tdDP {
	case 0: // SETUP
		pkt := make([]byte, 8)
		for i := range pkt {
			pkt[i] = m.Read(cbp + uint32(i))
		}
		data, err := dev.setup(m, pkt)
		if err != nil {
			m.usbStall(ed, td, ctrl, cbp, length, nextTD)
			return nextTD, false
		}
		m.usbCtrlData, m.usbCtrlOff = data, 0
		m.usbRetire(td, ctrl, cbp, length, length, ccNoError)

	case tdDPIn: // device -> host
		var src []byte
		if endpoint == 0 {
			src = m.usbCtrlData[min(m.usbCtrlOff, len(m.usbCtrlData)):]
		} else if src = dev.interruptIn(m, endpoint); src == nil {
			return nextTD, false // NAK: nothing new to report this frame
		}
		n := min(uint32(len(src)), length)
		for i := uint32(0); i < n; i++ {
			m.Write(cbp+i, src[i])
		}
		if endpoint == 0 {
			m.usbCtrlOff += int(n)
		}
		cc := uint32(ccNoError)
		if n < length && ctrl&tdRounding == 0 {
			// Short packet with buffer-rounding off: the driver asked for exactly this
			// many bytes and did not get them. Saying so is the point — an underrun
			// reported as success is a descriptor the driver will read past.
			cc = ccDataUnderrun
		}
		m.usbRetire(td, ctrl, cbp, n, length, cc)
		if cc != ccNoError {
			m.usbHaltEndpoint(ed, nextTD)
			return nextTD, false
		}
		m.usbControlStatus(dev, endpoint, length)

	case tdDPOut: // host -> device
		// The status stage of an IN control transfer, and the data stage of an OUT one.
		// Nothing this device has needs the bytes, but the TD must still retire, and it
		// must retire with a code that was written rather than assumed.
		m.usbRetire(td, ctrl, cbp, length, length, ccNoError)
		m.usbControlStatus(dev, endpoint, length)

	default:
		m.CPU.Halt("USB: TD %08X has reserved direction %d (ED %08X)", td, (ctrl&tdDP)>>19, ed)
		m.Halted, m.HaltReason = true, m.CPU.HaltReason
		return nextTD, false
	}
	return nextTD, true
}

// usbControlStatus tells a device that a control transfer has finished, if the TD just
// retired was the transfer's status stage.
//
// The discriminator is the packet itself, not bookkeeping: on endpoint 0, a zero-length
// packet that is not the SETUP is a status stage and nothing else. A data stage of zero
// length does not exist — wLength == 0 means the transfer HAS no data stage — so the
// two cannot be confused.
func (m *Machine) usbControlStatus(dev usbDevice, endpoint, length uint32) {
	if endpoint == 0 && length == 0 {
		dev.controlStatusDone(m)
	}
}

// usbRetire completes a TD: it writes the condition code, updates the buffer pointer,
// and puts the TD on the done queue the next writeback publishes. moved is how many of
// the TD's length bytes actually crossed the wire.
//
// The condition code is written unconditionally, success included. The driver reads
// these back out of memory it zeroed, and ccNoError is 0 — so "leave it alone on
// success" and "never ran" are the same bytes, and the difference between them is the
// difference between a working pad and a silent one.
func (m *Machine) usbRetire(td, ctrl, cbp, moved, length, cc uint32) {
	ctrl = (ctrl &^ tdCC) | (cc << 28)
	m.write32(td+0x00, ctrl)
	m.write32(td+0x04, usbCurrentBuffer(cbp, moved, length))
	m.usbDone = append(m.usbDone, td)
}

// usbCurrentBuffer is the CurrentBufferPointer a retiring TD reports: ZERO when the whole
// buffer was consumed, and otherwise the address of the next byte that would have moved.
//
// Zero is not "nothing here" — it is the success encoding, and the ONLY way a TD can say
// "all of it". There is no other room to say it: a pointer one past the end is a legal
// address, indistinguishable from a transfer that stopped there. XAPI reads it exactly
// that way, and the byte count of every transfer in the machine comes off this branch:
//
//	00246178  MOV EAX, [ESI+$4]      the retired TD's CBP
//	0024617E  JZ  $002461A6          zero -> all of it...
//	002461A6  MOVZX EAX, [ESI+$1D]   ...so add the length that was ASKED for
//	002461AA  ADD [EDI+$14], EAX     URB.transferred += EAX
//	                                 (non-zero takes the partial path above, which
//	                                  derives the count from CBP and BE)
//
// Writing cbp+moved unconditionally — a number that is right in every arithmetic sense
// and never zero — sent XAPI down the partial path on transfers that had completed in
// full, and it summed nonsense: 0xFFFFF000, 0xFFFFE008. The device descriptor SURVIVED
// that, because its only test is "did at least 8 bytes arrive" (0x242393) and garbage
// that large passes. The configuration descriptor is checked for EQUALITY against
// wTotalLength (0x2426D8), and that is where a wrong number finally had to be wrong.
func usbCurrentBuffer(cbp, moved, length uint32) uint32 {
	if moved >= length {
		return 0
	}
	return cbp + moved
}

// usbStall halts an endpoint: the device refused the request. The driver notices the
// halt, and that is the honest answer to a request this model does not implement — far
// better than a silent success against a zeroed buffer.
//
// Nothing moved, so the buffer pointer stays where the transfer would have begun. It
// must NOT be zeroed: zero is this field's way of claiming the whole buffer was
// transferred, which is the opposite of what a stall means.
func (m *Machine) usbStall(ed, td, ctrl, cbp, length, nextTD uint32) {
	m.usbRetire(td, ctrl, cbp, 0, length, ccStall)
	m.usbHaltEndpoint(ed, nextTD)
}

// usbHaltEndpoint stops an endpoint's queue after a TD retired with a condition code
// other than NoError: nothing else on this endpoint moves until the driver has looked.
//
// EVERY error halts, not just a refusal — and the one that matters is the mundane one. A
// short packet is how a device says "that is all I have", and for a driver that
// packetises a transfer by hand it is the ONLY way the transfer can end early: XAPI
// queues ten 8-byte TDs for a descriptor that might be nine bytes long, and the halt on
// TD 2 is what stops the other eight from running. Retiring the underrun without halting
// left them all to run against a device with nothing left to say, and the driver summed
// their empty answers into a byte count that matched nothing it had asked for.
//
// THE QUEUE STILL ADVANCES. The failed TD is finished — it has been retired onto the done
// queue and the driver has already read it there — so HeadP moves PAST it, and the Halted
// bit rides alongside as the reason the endpoint stopped. Leaving HeadP pointing AT the
// TD that failed looks like the more conservative choice and is a different claim
// entirely: it says the transfer has not happened yet. XAPI settles it by what it does
// next, having just counted that TD's bytes and marked the transfer a success:
//
//	002460A2  AND DWORD [EBX+$4], $00000000   the URB SUCCEEDED (a short read is not a fault)
//	002460C5  MOV ESI, [EDI+$8]               then re-read HeadP...
//	002460C8  AND ESI, $FFFFFFF0              ...mask the flags off...
//	002460E6  LEA EDX, [EAX+ESI*1]            ...and walk the queue from there to dequeue
//	002460ED  MOV ESI, [EDX+$8]               the TDs that never ran.
//
// A HeadP still naming the retired TD hands that cleanup the one TD it has already
// accounted for, and the walk goes round the queue it is trying to drain.
func (m *Machine) usbHaltEndpoint(ed, nextTD uint32) {
	head := m.read32(ed + 0x08)
	m.write32(ed+0x08, nextTD|(head&edHeadToggle)|edHeadHalted)
}

// usbWriteback publishes the frame's retired TDs to the driver: the done queue becomes
// a chain in HcDoneHead / the HCCA, and WritebackDoneHead is raised.
//
// The chain is built in reverse, newest first, which is the order the hardware reports
// and therefore the order the driver's own list-walk expects.
func (m *Machine) usbWriteback() {
	if len(m.usbDone) == 0 {
		return
	}
	var head uint32
	for _, td := range m.usbDone {
		m.write32(td+0x08, head) // NextTD -> the previously-done TD
		head = td
	}
	m.usbDone = m.usbDone[:0]

	// The HCCA's DoneHead is where the driver actually reads it from; the register is
	// the same value. Bit 0 set would mean "an unrelated interrupt is also pending",
	// which this model has no way to mean, so it stays clear.
	if hcca := m.usb.reg[hcHCCA>>2]; hcca != 0 {
		m.write32(hcca+0x84, head)
	}
	m.usb.reg[hcDoneHead>>2] = head
	m.usbRaise(ohciIntWDH)
}

// usbFrameNumberToHCCA publishes the frame number where the driver reads it.
func (m *Machine) usbFrameNumberToHCCA() {
	if hcca := m.usb.reg[hcHCCA>>2]; hcca != 0 {
		m.write16(hcca+0x80, uint16(m.usbFrame()))
	}
}

// usbUnsupported halts the machine naming a request the device model does not
// implement, the way dispatchKernel halts on an unmodelled ordinal — and for a sharper
// reason than symmetry.
//
// Zero is a legal answer everywhere in USB. A zero-length descriptor, a NAK, a report
// of no buttons held: each is something a real device can say, so a control pipe that
// returned zeros for "I do not know this request" would produce an enumeration that
// SUCCEEDS against garbage, and a pad that is present and permanently idle. The latch
// doctrine — log once, return zero, stay visible — is right where zero is implausible.
// Here it is the most plausible wrong answer there is.
//
// Halting instead buys the workflow this port already runs on: the halt names the
// request, EIP is untouched, and -loadstate clears it, so the next run retries the very
// packet that stopped the last one.
func (m *Machine) usbUnsupported(what string, args ...any) error {
	reason := fmt.Sprintf("USB: "+what, args...)
	m.CPU.Halt("%s", reason)
	m.Halted, m.HaltReason = true, m.CPU.HaltReason
	return errUSBUnsupported
}

var errUSBUnsupported = fmt.Errorf("usb: unmodelled request")
