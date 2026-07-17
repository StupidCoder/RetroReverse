package xbox

import "sort"

// usb_xid.go is the game pad: an XID-class USB device on a root hub port.
//
// This is the file with the most opportunity in the phase to write down something that
// is true and unearned. I know the shape of a Microsoft XID gamepad. Writing that shape
// down and watching the title accept it would prove nothing at all — the game would
// work, the screenshot would be right, and the model would rest on a memory instead of
// on evidence. So the rule here is the same one the rest of the port runs on: every
// field below is either something the guest's own code compares against and we read, or
// it is a value the guest never inspects and we say so.
//
// WHAT THE GUEST TOLD US. The consumer chain was traced end to end before this device
// existed, from the record the title polls back to the buffer nobody filled:
//
//	0x14630  MOV CX, [ESI]            a WORD of digital buttons at +0
//	         TEST CL,$10 -> OR EAX,$01   ...each bit remapped to a game bit
//	0x14680  MOVSX ECX, [ESI+$A]      a SIGNED WORD at +0xA — a stick axis
//	         CMP [ESI+$7], CL         BYTES at +6..+9, compared against a threshold:
//	                                  pressure-sensitive analog buttons
//
// That is the layout of XINPUT_GAMEPAD, the struct XAPI publishes — NOT of the wire
// report this device sends. XAPI translates one into the other, and the translation is
// its business, not ours. What this file owes the machine is the wire format; where that
// is not yet pinned from XAPI's own copy, the request halts by name rather than
// guessing, and the fix-and-resume loop brings the next question back.

// padButtons are the pad's digital buttons, by the names every tool that drives this
// machine uses for them — the oracle's -keys scripts and the debugger's keyboard alike.
// It lives here, next to the device that delivers them, so the two vocabularies cannot
// drift apart (the tools/platform/gc/si.go precedent, and the reason the PSP's pad names
// living in its oracle instead was a mistake worth not repeating).
//
// The BIT VALUES are the guest's, not ours: they are the wButtons bits XAPI publishes,
// each read off the title's own remapping at 0x14630 —
//
//	TEST CL,$10 -> OR EAX,$01   so 0x0010 is what the title treats as START
//
// — which is also why this map is digital-only for now. The pad's analog buttons are
// pressure bytes the title thresholds (CMP [ESI+$7], CL), and its sticks are signed
// words; naming those needs a vocabulary for axes that no platform in this repo has yet.
var padButtons = map[string]uint16{
	"up": 0x0001, "down": 0x0002, "left": 0x0004, "right": 0x0008,
	"start": 0x0010, "back": 0x0020, "lstick": 0x0040, "rstick": 0x0080,
}

// PadButton looks up a pad button's bit by name.
func PadButton(name string) (uint16, bool) {
	b, ok := padButtons[name]
	return b, ok
}

// PadButtonNames lists the button names PadButton accepts, sorted, for a caller that
// needs to show or validate the vocabulary.
func PadButtonNames() []string {
	names := make([]string, 0, len(padButtons))
	for n := range padButtons {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// AttachPad plugs a game pad into a root hub port, if one is not already there. It is
// idempotent, so a caller that cannot easily tell whether it has a pad yet may simply
// ask for one.
//
// A pad exists only because something asked for one — the oracle's -keys, the debugger's
// keyboard. It is NOT plugged in at construction, and that is the design, not laziness:
// a machine nobody intends to drive has an empty root hub, which is both the truth and
// the state every existing boot and savestate was taken in. Attaching flips the port's
// CCS and raises RootHubStatusChange, which is the whole of what the driver is waiting
// for (see usb.go) — and it is the last thing in the phase to be wired, because a port
// that reports a device the rest of the controller cannot then enumerate is worse than
// an empty one.
func (m *Machine) AttachPad(port int) {
	if port < 0 || port >= usbPorts || m.usbDev[port] != nil {
		return
	}
	m.usbDev[port] = &xidDevice{}
	m.usbSetPortConnected(uint32(port), true)
}

// SetPadButtons sets the digital level a port's pad reports.
//
// A level, not an event: the pad reports what is held, and the title edge-detects
// presses from consecutive reports (its own ~prev&cur at 0xA51C8). So a press that goes
// down and up between two polls is a press the title never sees, which is the hazard
// every caller of this has to pace for.
func (m *Machine) SetPadButtons(port int, buttons uint16) {
	if port < 0 || port >= usbPorts {
		return
	}
	d, _ := m.usbDev[port].(*xidDevice)
	if d == nil {
		return
	}
	if d.Buttons != buttons {
		d.Buttons, d.Fresh = buttons, true
	}
}

// PadButtons reads back the digital level a port's pad currently reports.
func (m *Machine) PadButtons(port int) uint16 {
	if port < 0 || port >= usbPorts {
		return 0
	}
	if d, _ := m.usbDev[port].(*xidDevice); d != nil {
		return d.Buttons
	}
	return 0
}

// xidDevice is one Xbox game pad.
type xidDevice struct {
	Addr    uint32 // 0 until the driver assigns one with SET_ADDRESS
	Config  uint32 // the configuration SET_CONFIGURATION selected
	Buttons uint16 // the digital level SetPadButtons holds
	Sent    uint16 // the last level actually reported, so a poll can NAK when nothing changed
	Fresh   bool   // a level is waiting to be reported

	// AddrNext is the address SET_ADDRESS asked for, held until that transfer's own
	// status stage has been answered. See setup().
	AddrNext  uint32
	AddrArmed bool
}

func (d *xidDevice) address() uint32 { return d.Addr }

// controlStatusDone applies whatever the finished control transfer asked for that could
// not be done while it was still running.
func (d *xidDevice) controlStatusDone(m *Machine) {
	if d.AddrArmed {
		d.Addr, d.AddrArmed = d.AddrNext, false
	}
}

// USB standard requests (bRequest), and the bmRequestType bits that classify them.
const (
	usbReqGetStatus        = 0x00
	usbReqClearFeature     = 0x01
	usbReqSetFeature       = 0x03
	usbReqSetAddress       = 0x05
	usbReqGetDescriptor    = 0x06
	usbReqSetDescriptor    = 0x07
	usbReqGetConfiguration = 0x08
	usbReqSetConfiguration = 0x09
	usbReqGetInterface     = 0x0A
	usbReqSetInterface     = 0x0B

	// xidReqGetReport is bRequest 1 under a CLASS-typed request, which is how XAPI asks
	// the pad for its opening input report. See setup().
	xidReqGetReport = 0x01

	usbTypeMask     = 0x60 // bmRequestType[6:5]: 0 standard, 1 class, 2 vendor
	usbTypeStandard = 0x00
	usbTypeClass    = 0x20
	usbTypeVendor   = 0x40
)

// setup answers a control transfer's SETUP packet.
//
//	pkt[0] bmRequestType  pkt[1] bRequest  pkt[2:4] wValue
//	pkt[4:6] wIndex       pkt[6:8] wLength
func (d *xidDevice) setup(m *Machine, pkt []byte) ([]byte, error) {
	bmType := pkt[0]
	bReq := pkt[1]
	wValue := uint16(pkt[2]) | uint16(pkt[3])<<8

	// The XID descriptor request, and it is a VENDOR request addressed to the interface —
	// which is worth pinning, because it is easy to read as a class one and this port's
	// earlier notes did. The class driver issues it the moment it owns the interface:
	//
	//	SETUP bmRequestType=C1 bRequest=06 wValue=4200 wIndex=<bInterfaceNumber> wLength=16
	//
	// 0xC1 is 1_10_00001: device->host, type 0b10 = VENDOR (not 0b01 = class, which would
	// have been 0xA1), recipient 1 = interface. bRequest 6 is GET_DESCRIPTOR's number
	// reused under vendor type, and wValue's high byte is the XID descriptor's type.
	if bmType == 0xC1 && bReq == usbReqGetDescriptor && wValue>>8 == xidDescType {
		return xidDescriptor[:], nil
	}

	// The pad's opening input report, fetched over the CONTROL pipe before the interrupt
	// poll is ever armed:
	//
	//	SETUP bmRequestType=A1 bRequest=01 wValue=0100 wLength=<the input report size>
	//
	// 0xA1 is 1_01_00001: device->host, type 0b01 = CLASS this time, recipient interface.
	// That this is an INPUT REPORT is not read off the request — it is read off what XAPI
	// does with the answer, which is the same thing it does with every interrupt report:
	//
	//	00244202  ...                      the request is built with
	//	0024420F  MOVZX EAX, [ESI+$C]        the INPUT REPORT SIZE as its length...
	//	0024421D  LEA ECX, [EBX+$32]         ...into the INPUT REPORT BUFFER...
	//	0024422A  MOV DWORD [EBX+$5A], $002428DF
	//	002442AC  MOV ECX, EBX / CALL [EAX+$24]   ...and then cooked by the type's own
	//	002442B1  ...                        routine, before the interrupt IN poll that
	//	                                     will deliver every later report is armed.
	//
	// Same buffer, same length, same cook. It is a poll, asked once, down a different
	// pipe — so it is answered with the same bytes the pad would report right now. XAPI
	// skips it entirely for a type whose object has bit 0x40 at +0x28 (0x2441F5); the
	// gamepad's does not, so the pad is asked.
	if bmType == 0xA1 && bReq == xidReqGetReport && wValue == 0x0100 {
		return d.report(), nil
	}

	// ---------------------------------------------------------------------------------
	// THE CAPABILITY REQUESTS ARE A TRACER, NOT A MODEL. THIS IS THE PHASE'S OPEN EDGE.
	// ---------------------------------------------------------------------------------
	//
	// The GAME asks for these, not the driver: XInputGetCapabilities (0x240B4A) is called
	// straight from the title's own pad enumeration (0x3963C and 0x39701, beside the
	// XGetDeviceChanges poll E1 traced). It issues two vendor requests and the SIZES of
	// both are derived — they come out of XAPI's own type object for the gamepad
	// (0x0023F4D4), each field a {byte length, pointer} pair whose length it reads and
	// then asks for two more than:
	//
	//	00240BA0  MOV EAX,[EAX+$C] / MOVZX EAX,[EAX]   type+0xC -> 0x23F4C8: length 4
	//	00240BBD  LEA ECX,[EAX+$2]                     -> wValue 0200 asks for 4+2 = 6
	//	00240C55  MOV EAX,[EAX+$8] / MOVZX EAX,[EAX]   type+8   -> 0x23F4BC: length 0x12
	//	                                               -> wValue 0100 asks for 18+2 = 20
	//
	// That 18+2 is worth pausing on: it is xidReportSize, arrived at from a completely
	// different direction — XAPI's own registry saying the gamepad's payload is 18 bytes
	// behind a 2-byte header, agreeing with the memcpy that gave us 20 in the first place.
	//
	// The CONTENT is not derived, and these bytes are absurd on purpose. Two things are
	// known about them and the second is why this is unfinished rather than free:
	//
	//  1. Nothing that RUNS reads the payload. A read watch over the game's caps buffer
	//     (0x54616E) across the whole milestone sees exactly one consumer — the title
	//     reading caps[0], the subtype, at 0x14C95 — and nothing whatever touching
	//     caps+1.. The counterfactual agrees: STALLING both requests outright leaves the
	//     ★ milestone bit-for-bit intact, gate and all. So no reached code depends on
	//     these bytes.
	//  2. THAT CENSUS IS BLIND, AND IN THE EXACT WAY THIS PHASE HAS ALREADY BEEN BURNED.
	//     XInputGetCapabilities FAILS. It returns 0x48F and zeroes caps+1..+2 on its way
	//     out (0x240D00: MOV DWORD [EBP-$8],$048F / 0x240D0D: AND WORD [EAX+$1],$0000),
	//     which means the code that would consume the payload — everything past 0x240C52 —
	//     NEVER RAN. "Nothing reads it" is a statement about a machine that died before
	//     the reader, which is precisely what nearly shipped a synthetic bDeviceClass in
	//     E3. The census is not evidence until the failure is understood.
	//
	// And the failure looks like it may not be about these bytes at all. It is not the URB
	// status check (0x240C49 JL, which lands elsewhere) — it is 0x240C39 or 0x240C43:
	//
	//	00240C32  MOV EAX, [EBP+$8]      the handle XInputOpen returned...
	//	00240C37  CMP [EAX], ECX
	//	00240C39  JZ  $00240D00          ...whose device object is NULL -> "not connected"
	//	00240C3F  TEST BYTE [ESI+$4], $02
	//	00240C43  JNZ $00240D00          ...or the device is flagged REMOVED
	//
	// A pad that enumerates, answers, drives the title's own START handler, and is
	// simultaneously reported to the game as disconnected is the shape of a real bug in
	// this model, not a missing descriptor. That is the thread to pull first.
	if bmType == 0xC1 && bReq == 0x01 && (wValue == 0x0100 || wValue == 0x0200) {
		n := 20 // wValue 0100: the input capabilities, 18 + the 2-byte header
		if wValue == 0x0200 {
			n = 6 // wValue 0200: the output capabilities, 4 + the same header
		}
		r := make([]byte, n)
		for i := range r {
			r[i] = byte(0xD0 + i) // each byte names its own offset under a read watch
		}
		return r, nil
	}
	if bmType&usbTypeMask != usbTypeStandard {
		return nil, m.usbUnsupported(
			"unmodelled control request bmRequestType=%02X bRequest=%02X wValue=%04X",
			bmType, bReq, wValue)
	}

	switch bReq {
	case usbReqSetAddress:
		// The device keeps answering on its OLD address until this transfer's status
		// stage has been answered, and only then moves.
		//
		// That is the spec's rule, and the reason for it is not ceremony: the status
		// stage is part of THIS transfer, and it rides the very endpoint descriptor the
		// SETUP came in on — the one XAPI built for address 0. Taking the address here
		// used to look safe, guarded by a comment reasoning that the control pipe drains
		// within one frame so nothing could address the device in between. Nothing else
		// had to: the transfer addressed it itself. The instant Addr became 1 the
		// address-0 lookup found no device, the zero-length status TD NAKed, and XAPI
		// waited forever on a transfer that could no longer complete — a device that
		// enumerated perfectly and then vanished between two TDs of one transfer.
		d.AddrNext, d.AddrArmed = uint32(wValue)&0x7F, true
		return nil, nil

	case usbReqSetConfiguration:
		d.Config = uint32(wValue) & 0xFF
		return nil, nil

	case usbReqGetConfiguration:
		return []byte{byte(d.Config)}, nil

	case usbReqGetDescriptor:
		return d.descriptor(m, wValue>>8, byte(wValue))
	}

	return nil, m.usbUnsupported(
		"unmodelled standard control request bmRequestType=%02X bRequest=%02X wValue=%04X",
		bmType, bReq, wValue)
}

// descriptor answers GET_DESCRIPTOR. Each type stays refused until XAPI's own matching
// code has been read for it: a descriptor is exactly the kind of thing that would be
// accepted whether or not it was earned.
func (d *xidDevice) descriptor(m *Machine, dtype uint16, index byte) ([]byte, error) {
	switch dtype {
	case usbDescDevice:
		return deviceDescriptor[:], nil
	case usbDescConfiguration:
		return configDescriptor[:], nil
	}
	return nil, m.usbUnsupported(
		"unmodelled GET_DESCRIPTOR type=%02X index=%02X — read XAPI's comparison before answering",
		dtype, index)
}

// The descriptor types, as they appear in GET_DESCRIPTOR's wValue high byte and in each
// descriptor's own second byte. Each of these is a number the guest compares against:
// DEVICE at 0x2423AA, CONFIGURATION as the type XAPI asks for after the class gate,
// INTERFACE at 0x242495, ENDPOINT at 0x242A2F.
const (
	usbDescDevice        = 1
	usbDescConfiguration = 2
	usbDescInterface     = 4
	usbDescEndpoint      = 5
)

// deviceDescriptor is the pad's DEVICE descriptor, and every byte of it is either a value
// XAPI compares or a value XAPI never looks at. Nothing here is remembered.
//
// XAPI asks for the first 8 bytes only (wLength=8, on the address-0 pipe) and validates
// them at 0x242365, which is where the first four of these came from:
//
//	00242393  CMP DWORD [EDX+$14], $00000008   at least 8 bytes must have arrived
//	0024239D  MOV AL, [$0057CEEB]              [7] bMaxPacketSize0...
//	002423A2  CMP AL, $40 / JA                 ...<= 0x40
//	002423AA  CMP BYTE [$0057CEE5], $01 / JNZ  [1] bDescriptorType == 1
//	002423B3  MOV CL, [$0057CEE4]              [0] bLength...
//	002423B9  CMP CL, $08 / JZ                 ...== 8...
//	002423BE  CMP CL, $12 / JNZ                ...or == 18
//
// and [4] came from the gate 2 ms later, at 0x242740, which is the one that decides what
// the pad IS:
//
//	00242765  MOV AL, [$0057CEE8]   [4] bDeviceClass
//	0024276C  JZ  $002427A2           == 0 -> the generic path: read the CONFIGURATION
//	                                  descriptor and match a driver at the INTERFACE
//	0024276E  CMP AL, $09             == 9 -> a hub
//	0024278F  MOV BYTE [EBP-$4], $81  else -> look for a DEVICE-level driver, tag 0x81
//	00242798  CALL $00242331
//
// and there is exactly one tag-0x81 driver in XAPI's table (0x0023F3F4): the hub, at
// class 9. So a pad that declares any other non-zero class matches nothing, and XAPI
// disables its port — which is not a guess, it is what this model did while it was
// answering 0xE4 there. Zero is the only class that leads anywhere.
var deviceDescriptor = [18]byte{
	0x12, // [0] bLength: XAPI takes 8 or 18; 18 is this descriptor's size
	0x01, // [1] bDescriptorType: DEVICE, XAPI requires it
	0x00, // [2] bcdUSB lo  ) SYNTHETIC. XAPI never reads either byte — proven by a read
	0x00, // [3] bcdUSB hi  ) watch over the whole buffer, not by their looking unused.
	0x00, // [4] bDeviceClass: forced to 0 by the driver table, see above
	0xE5, // [5] bDeviceSubClass ) SYNTHETIC, and deliberately absurd: the class==0 branch
	0xE6, // [6] bDeviceProtocol ) jumps OVER the code that reads these (0x24277F/0x242787)
	0x08, // [7] bMaxPacketSize0: see below
	// [8..17] are never requested: XAPI asks for 8 bytes here and never comes back for
	// the other ten, so idVendor, idProduct, bcdDevice, the string indices and
	// bNumConfigurations have no evidence behind them at all. They are left zero rather
	// than filled with plausible numbers, because a plausible number here would be a
	// claim about a real product that nothing in this port has ever observed.
}

// bMaxPacketSize0 is 8. It has now been wrong in BOTH directions, and the two mistakes
// are worth keeping because they are opposite shapes of the same error.
//
// XAPI bounds it at 0x40 (0x2423A2) and reads nothing else out of it here, which made 8
// look like a free and even principled answer: 8 is the MPS XAPI had itself programmed
// into the address-0 control ED. Then this comment claimed the field was a trap — that it
// is stashed on a device object and read back much later to cap the pad's input report:
//
//	002423CB  MOV [ESI+$6], AL     bMaxPacketSize0, kept at the state machine's ESI+6...
//	0024407E  CMP AL, [ESI+$6]     ...and the XID input report size, bounded against ESI+6
//
// THOSE ARE TWO DIFFERENT OBJECTS, and the two lines were welded together by nothing but
// the register they happen to share. 0x2423CB's ESI is the enumeration state machine's
// USB device object (0x2423C4: MOV ESI,[ESP+$10]). 0x24407E's ESI is the XID driver's own
// extension record — element n of a 0x16-byte array at [0x57CF4C], carved out by its
// AddDevice (0x24436B: IMUL ESI,ESI,$16 / ADD ESI,EAX) and handed to the validator as its
// context (0x2443F3 stores it; 0x244038 reads it back). The extension is not the device
// object: it holds a POINTER to one, at its own offset 0 (0x24438C: MOV [ESI], EDI).
//
// And the byte the report is really bounded against is written in between, by AddDevice,
// out of the ENDPOINT descriptor (0x2443AF: MOV AL,[EAX+$4] / MOV [ESI+$6], AL — the
// interrupt IN endpoint's wMaxPacketSize lo). Which is the sensible thing for a USB stack
// to check, and it is checked one field over from its sibling: [ESI+7] is the OUT
// endpoint's packet size, bounding the OUTPUT report at 0x24408A, guarded by [ESI+9] — the
// OUT endpoint's address, zero when the pad has no OUT endpoint at all. Three fields
// written by one function, read by the next: the extension's layout closes on itself, and
// the device object's does not appear in it anywhere.
//
// So the field does not cap the report, and the argument that raised it to 0x40 — "0x40
// constrains the least, so it decides the least" — was answering a question that was not
// being asked. It is 8 now, and NOT because 0x40 was merely unmotivated. The field is
// load-bearing after all, just somewhere else again:
//
//	XAPI copies bMaxPacketSize0 into the CONTROL ED's own MPS field, and then packetises
//	every later control transfer by it. With 0x40 the pad's control ED reads ctrl=00400001
//	— MPS 64 — and XAPI splits its 0x50-byte configuration read into a 64-byte data TD and
//	a remainder. With 8 it queues 8-byte TDs, which is what E2 watched it do to its own
//	default address-0 pipe (0x00080000) before it had ever met a device.
//
// And at 0x40 the guest DIES, ~7.6M instructions past the title, inside its own OHCI
// cleanup at 0x2460ED, reading FFFFFFF8. The cause is visible in its own queue. When a
// short answer underruns a data TD the endpoint halts, and XAPI walks the remaining TDs to
// dequeue them — a walk with NO TailP check (0x2460CB..0x2460FA), terminated only by a TD
// whose software byte at +0x1C has bit 1 set, meaning "last of this URB":
//
//	td=013EC290  next=013EC2D0  [+1C]=00     keep walking
//	td=013EC2D0  next=013EC250  [+1C]=02     <- the marker: stop here
//	td=013EC250  next=FFFFFFFF  [+1C]=00     <- the ED's DUMMY TAIL. Its NextTD is
//	                                            uninitialised, and is meant to be: nothing
//	                                            is ever supposed to read it.
//
// A walk that reaches the dummy tail reads FFFFFFFF as a TD pointer and is gone. At MPS 8
// the marker lands where XAPI expects and the walk stops; at MPS 64 it does not. This
// model is NOT papering over that with a value that happens to work — but the honest
// statement of what is known has two halves, and only the first is a derivation:
//
//  1. DERIVED: 8 is the packet size XAPI itself chose for a control pipe, observed before
//     any device existed to influence it. A pad that declares 8 is telling the driver the
//     thing the driver had already assumed.
//  2. OBSERVED, NOT EXPLAINED: at 0x40 XAPI's own cleanup walks off its own queue. Whether
//     that is a latent bug in a first-party stack that never met a 64-byte control pipe, or
//     whether this transfer engine hands it something subtly wrong that only a multi-TD
//     control transfer exposes, is NOT settled here. It is written down rather than
//     smoothed over, because "the value that works" is precisely the kind of evidence this
//     phase cannot take at face value, and if it is our bug it will come back.
//
// The through-line of all three readings — free, then load-bearing at 0x24407E, then
// load-bearing at the ED — is one lesson: ask what a value is READ BY, and answer it by
// following the OBJECT, not the register that happens to hold it.

// xidInterfaceClass is the class the pad declares at its interface, and it is the one
// field in this file the evidence does NOT pin down.
//
// XAPI matches an interface to a driver on {tag, class} and nothing else (0x240ECF
// compares [EAX] and [EAX+1]; subclass and protocol are packed into the key and never
// looked at). Its table holds five interface-level (tag 0x82) classes — 0x03, 0x08, 0x58,
// 0x78, 0x79 — and the pad was run once per candidate. Only 0x03 and 0x58 are ever
// claimed, and they are claimed by THE SAME DRIVER OBJECT:
//
//	0023F564  tag 0x82  class 0x03  -> 0023FC97 / 0024432E / 00243D0F
//	0023F54C  tag 0x82  class 0x58  -> 0023FC97 / 0024432E / 00243D0F
//
// The two runs are identical instruction for instruction. So the evidence constrains the
// class to that pair and no further, and 0x58 below is a PICK, not a derivation: there is
// no comparison anywhere in XAPI that can tell the two apart, because XAPI registered one
// XID driver under both. Anything that claimed to have chosen between them would be
// reasoning from outside the image. If a later frontier ever distinguishes them, this is
// the constant to revisit, and the run that distinguishes it is the evidence to cite.
const xidInterfaceClass = 0x58

// configDescriptor is the pad's CONFIGURATION descriptor and the bundle that follows it:
// the interface, and the interrupt IN endpoint the reports ride on. XAPI asks for the
// whole thing in one 0x50-byte request.
//
// The SHAPE is forced by the guest, not chosen. XAPI walks the bundle by each descriptor's
// own bLength (0x24247C: MOVZX EDX,[EAX] / ADD EAX,EDX), so every bLength must land
// exactly on the next descriptor or the walk desynchronises and finds nothing. The lengths
// below are therefore self-referential — each is the size of the descriptor we emit — and
// the bytes those lengths make room for that XAPI never reads carry deliberately absurd
// tracer values, so that a reader we have not yet found announces itself as garbage rather
// than passing silently on a plausible number. (Emitting the unread bytes rather than
// eliding them is the safer half of that: a shorter descriptor would put the NEXT
// descriptor under an offset the census only ever proved unread on the code the run had
// REACHED — which is exactly the trap that nearly shipped a synthetic bDeviceClass.)
//
// The endpoint's fields came last, and they came from a generic search helper at 0x242A02
// which the XID driver calls twice from its own AddDevice:
//
//	00244396  PUSH EBX (0) / PUSH $1 / PUSH $3 / CALL $00242A02   -> the INTERRUPT IN ep
//	002443A8  MOV CL,[EAX+$2] / MOV [ESI+$8], CL    bEndpointAddress -> the extension
//	002443AF  MOV AL,[EAX+$4] / MOV [ESI+$6], AL    wMaxPacketSize lo -> the extension
//	002443AB  PUSH EBX (0) / PUSH EBX (0) / PUSH $3 / CALL $00242A02  -> the INTERRUPT OUT
//	002443BF  CMP EAX, EBX / JZ  -> absent is FINE: [ESI+9] and [ESI+7] are zeroed
//
// and the helper itself is what the three arguments mean:
//
//	00242A2F  CMP AL, $05 / JNZ        bDescriptorType MUST BE 5 (ENDPOINT)
//	00242A33  MOV DL,[ECX+$3] / AND DL,$03 / CMP DL,[EBP+$8]   bmAttributes&3 == the
//	                                                            wanted transfer type (3)
//	00242A3E  CMP BYTE [EBP+$8], $00 / JZ    ...and type 0 is rejected outright
//	00242A46  MOV DL,[ECX+$2] / SHR EDX,$07 / NOT EDX / AND EDX,$01
//	00242A53  CMP [EBP+$C], BL / SETZ BL / CMP EDX, EBX / JNZ
//	          i.e. match when (!bEndpointAddress[7]) == (dirArg == 0): a nonzero dirArg
//	          demands bit 7 SET. The IN call passes 1, so the endpoint must be an IN.
//	00242A67  CMP AL, $04 / JNZ loop         the walk STOPS at the next INTERFACE
//
// The search starts at the interface descriptor and steps before it looks, so the endpoint
// must FOLLOW the interface, and be reached before any other interface descriptor would
// be. There is exactly one of each here, so both hold trivially.
var configDescriptor = [25]byte{
	// --- CONFIGURATION, 9 bytes
	0x09, // [0] bLength: the walker's step; 9 is the size of this descriptor
	0x02, // [1] bDescriptorType: CONFIGURATION
	0x19, // [2] wTotalLength lo ) XAPI: MUST be <= 0x50, the size of its own buffer
	0x00, // [3] wTotalLength hi ) (0x2426B5), and MUST EQUAL the bytes actually
	//       transferred (0x2426D8). 0x19 = 25 = the whole bundle below. This is the
	//       only length in the enumeration that an under-reporting transfer engine
	//       could not fool — see the work log's E4.
	0x01, // [4] bNumInterfaces: XAPI requires exactly 1 (0x24249B)
	0xC5, // [5] bConfigurationValue: free, and PROVEN to be plumbed — 0x2426E6 reads it
	//       and 0x242711 makes it SET_CONFIGURATION's wValue, where this tracer came
	//       back to us as wValue=00C5. A device's own name for its configuration.
	0xE8, // [6] bmAttributes    ) SYNTHETIC. Read by nothing in XAPI's walk; absurd on
	0xE9, // [7] bMaxPower       ) purpose. A real hub would care about the power bits;
	0xEA, // [8] iConfiguration  ) this pad is asked for none of it.

	// --- INTERFACE, 9 bytes
	0x09, // [0] bLength: the walker's step (0x24248C reads it; zero stops the walk)
	0x04, // [1] bDescriptorType: XAPI hunts for exactly this (0x242495)
	0xD2, // [2] bInterfaceNumber: free, and PROVEN plumbed — 0x24253A reads it and the
	//       class driver sent it straight back as the XID request's wIndex=00D2.
	0xEB, // [3] bAlternateSetting ) SYNTHETIC — never read.
	0xEC, // [4] bNumEndpoints     ) NOT read either, which is worth naming: XAPI does not
	//       trust this count, it walks until it runs out of buffer or hits the next
	//       interface. So the pad's endpoint count is told by the bundle's shape alone.
	xidInterfaceClass, // [5] bInterfaceClass: the tag-0x82 driver key (0x242545). See above.
	0xED,              // [6] bInterfaceSubClass ) Packed into the driver key at 0x24254B/0x24254E
	0xEE,              // [7] bInterfaceProtocol ) and then NEVER MATCHED — 0x240ECF compares the
	//       tag and the class only. Absurd values ride into the key and change nothing,
	//       which is the proof.
	0xEF, // [8] iInterface: SYNTHETIC — never read.

	// --- ENDPOINT, 7 bytes: the interrupt IN pipe the reports arrive on
	0x07, // [0] bLength: the search helper's step (0x242A1D)
	0x05, // [1] bDescriptorType: ENDPOINT, the helper requires it (0x242A2F)
	0x81, // [2] bEndpointAddress: bit 7 MUST be set — the helper's IN call demands it
	//       (0x242A46). The endpoint NUMBER in the low bits is free; 1 is ours, and it
	//       is the number interruptIn() below answers on, which is the only thing that
	//       has to agree with it.
	0xE3, // [3] bmAttributes: the low two bits MUST be 3 (0x242A33/0x242A36 mask with
	//       $03 and compare against the driver's request of 3 = INTERRUPT). The upper
	//       six bits are masked away unread, so they are absurd on purpose.
	0x20, // [4] wMaxPacketSize lo: the report size is bounded against this
	//       (0x24407E: CMP AL,[ESI+$6] / JA reject, where [ESI+6] is THIS byte, stored
	//       at 0x2443B7). 0x20 = 32 is the value that decides the least: XAPI already
	//       caps any XID report at 32 (0x244077 CMP AL,$20 / JA), so a 32-byte packet
	//       admits every report XAPI would accept and this bound never binds. The same
	//       reasoning that put 0x40 in bMaxPacketSize0 — for the same reason: the
	//       report size is not yet known and must not be decided from over here.
	0xF0, // [5] wMaxPacketSize hi ) SYNTHETIC — only the LO byte is ever read (0x2443AF
	//       takes a single byte), so the high half of this word is unexamined.
	0xF1, // [6] bInterval: SYNTHETIC — never read. The poll rate is XAPI's own business;
	//       it builds the interrupt ED's schedule without asking the device for it.
}

// xidDescType is the XID class descriptor's type. The class driver asks for it the moment
// it owns the interface — SETUP bmRequestType=C1 bRequest=06 wValue=4200 wIndex=<our
// bInterfaceNumber> wLength=16 — and its validator requires the type back in the
// descriptor's own second byte (0x24401E: CMP BYTE [$0057CF55], $42).
const xidDescType = 0x42

// xidReportSize is the pad's input report size, in bytes on the wire, and it is DERIVED —
// it is the one number in this file that the rest of the layout hangs off, so it is worth
// saying exactly what forced it.
//
// XAPI reads it out of the XID descriptor below into its extension record (0x244060: MOV
// [ESI+$C], CL), bounds it, and then uses it as the interrupt transfer's length:
//
//	00244073  CMP AL, $02 / JB   reject   the report must be at least 2 bytes...
//	00244077  CMP AL, $20 / JA   reject   ...and at most 32
//	0024407E  CMP AL, [ESI+$6] / JA       ...and fit one endpoint packet (0x2443AF put
//	                                       the IN endpoint's wMaxPacketSize there)
//	002438B0  MOVZX EAX, [EBX+$C]         and THEN it is simply how many bytes the URB
//	002438B4  MOV [ESI+$66], EAX          asks the pad for, into the buffer at +0x32
//	00243898  LEA EAX, [ESI+$32]          (0x2438AD: MOV [ESI+$6A], EAX)
//
// Those bounds leave 2..32 open. What CLOSES it is the other end — what XAPI does with the
// bytes once they land. Its cook routine (0x2438DA, reached as the completion callback's
// CALL [EAX+$24], where EAX is the gamepad type object at 0x23F4D4) is a plain memcpy:
//
//	002438DC  MOV ECX, [EAX+$66]     ECX = the report size we declared
//	002438E2  CMP ECX, $2 / JB       under 2 bytes, there is nothing to cook
//	002438E5  LEA EDX, [EAX+$14]     the DESTINATION: XAPI's XINPUT_GAMEPAD
//	002438EB  ADD ECX, $FFFFFFFE     ...and it moves the report MINUS TWO BYTES...
//	002438EF  LEA ESI, [EAX+$34]     ...from the wire buffer PLUS TWO (0x32 + 2)
//	002438F9  REP MOVSD / REP MOVSB
//
// So the wire report is a two-byte header XAPI throws away, followed by XINPUT_GAMEPAD
// verbatim. The report size is therefore 2 + however much of that struct anything reads,
// and the readers are known, both of them:
//
//	00243906  LEA EAX,[EDX+$2] / PUSH $8 / CMP BYTE [EAX],$20   XAPI itself walks EIGHT
//	                                       analog bytes at gamepad+2..+9, unconditionally
//	00014630  MOV CX, [ESI]                the title: a WORD of digital buttons at +0
//	000147DE  CMP [ESI+$2..$9], CL         the title: the same eight analog bytes
//	00014680  MOVSX ECX, [ESI+$A]          the title: four SIGNED WORDS — the sticks —
//	000146D8/00014730/00014788             at +0xA, +0xC, +0xE and +0x10
//
// The furthest byte any of them touches is gamepad+0x11, so the struct XAPI fills is 18
// bytes, and 18 + 2 = 20. That is the smallest report that leaves no field a reader reads
// unwritten, and every byte past it would be a byte nothing in this image ever looks at.
// It also satisfies every bound above, and it fits: the gamepad lives at +0x14 and the
// wire buffer at +0x32, which is 30 bytes of room for 18.
const xidReportSize = 20

// xidDescriptor is the XID class descriptor. XAPI asks for 22 bytes; this answers with 8,
// which is a SHORT PACKET and is fine — the driver requires at least 8 (0x244007: CMP
// DWORD [EAX+$14], $8) and reads no further than offset 7. Eight bytes is everything there
// is evidence for, and a descriptor claiming to be longer would have to invent the rest.
//
// Its validator is 0x244011, and the whole of it is quoted against the fields below.
var xidDescriptor = [8]byte{
	0x08,        // [0] bLength: XAPI requires >= 8 (0x244011); 8 is what we emit
	xidDescType, // [1] bDescriptorType: 0x42, required (0x24401E)
	0xF2,        // [2] ) A WORD that XAPI requires to be NONZERO and inspects no further
	0xF3,        // [3] ) (0x24402B: CMP [$0057CF56], BX / JZ reject, with BX = 0). Absurd on
	//       purpose: the guest's only interest in it is that it exists.
	0x01, // [4] bType: 0x01, and this one is DERIVED THROUGH XAPI'S OWN REGISTRY to the
	//       struct the title polls. 0x240A86 matches this byte against byte 0 of each
	//       type object at 0x0023F3E4..0x0023F3EC (0x240AA4: CMP [EDI], CL). There are
	//       two. The one whose byte 0 is 0x01 is 0x0023F4D4, and its +4 field points at
	//       0x0023F49C — which is the struct the GAME hands to XGetDeviceChanges at
	//       0x39670, pinned as the pad's in E1 long before this device existed. That
	//       same type object's +0x24 is the cook routine quoted above. So 0x01 is not a
	//       remembered constant: it is whatever byte XAPI filed beside a pointer to the
	//       pad's own device-type struct.
	0xF4, // [5] the SUBTYPE, and the one field here whose constraint is a single bit of
	//       information. This was briefly written down as "synthetic, read by nothing in
	//       the driver" — true, and useless, because its reader is not in the driver. It
	//       is kept at extension+0xB (0x244057: MOV [ESI+$B], CL), and XInputGetCapabilities
	//       publishes it to the GAME as the first byte of its capabilities struct:
	//
	//	00240B86  MOV AL, [ESI+$B]      the subtype off the extension...
	//	00240B89  MOV [EDX], AL         ...becomes caps[0], the caller's buffer
	//	00014C95  MOV BL, [ECX]         and the title reads it back, per pad...
	//	00014CA7  CMP BL, $10           ...and asks ONE question of it:
	//	00014CAA  SETZ DL               is it 0x10?
	//	00014CC3  MOV [ESI-$1D4], EDX   the answer is filed in the pad's own record
	//
	//       (Watched live: this tracer came back to the title as 0xF4 at 0x14C95.) So the
	//       image constrains this byte to exactly two classes — 0x10, and everything else —
	//       and the game keeps a per-pad flag for which. Nothing observed distinguishes
	//       further, and nothing tells us which one this pad is. 0xF4 answers "not 0x10"
	//       while staying obviously synthetic, so the flag is 0 and the value cannot be
	//       mistaken for a derived one. If a later frontier shows the title treating the
	//       0x10 pads differently in a way that matters, THAT is the evidence to change it
	//       on — not a memory of what 0x10 means.
	xidReportSize, // [6] the INPUT report size. See above.
	0x00,          // [7] the OUTPUT report size, and it is FORCED to zero by our own endpoint
	//       bundle. The config descriptor gives this pad no interrupt OUT endpoint —
	//       nothing in the image has yet demanded one — and XAPI, asked to send an
	//       output report to a pad that declares none, refuses in its own words:
	//
	//	00243E0A  CMP [EDI+$D], BL     the output report size...
	//	00243E0D  JNZ $00243E1A
	//	00243E0F  MOV DWORD [ESI], $32  ...zero -> the request fails, no pipe is built
	//
	//       A nonzero value here would send XAPI down 0x243E1A to drive an endpoint that
	//       does not exist. Zero is the answer that matches the pad we actually built.
	//       (The validator would not have caught it: its output-size bound at 0x24408A is
	//       guarded by extension+9, the OUT endpoint's address, which is zero here — so
	//       the check is skipped entirely. The coherence has to come from us.)
}

// interruptIn is the pad's report endpoint, polled by the driver every frame or so.
//
// It NAKs — returns nil — when the level has not changed since the last report. That is
// not an optimisation: a report is only interesting because the level moved, and a
// device that answers every poll with the same bytes is indistinguishable from one that
// is stuck. It also means a press and its release cannot collapse into one poll.
//
// THE LAYOUT IS XAPI'S, read off the memcpy quoted at xidReportSize: two header bytes the
// driver strips, and then XINPUT_GAMEPAD verbatim. Nothing here was laid out from
// knowledge of what an XID pad looks like and then confirmed by the game working — which
// would have proved nothing, because the game working is precisely the evidence that
// cannot tell a derived layout from a remembered one.
func (d *xidDevice) interruptIn(m *Machine, endpoint uint32) []byte {
	if !d.Fresh {
		return nil
	}
	d.Fresh = false
	d.Sent = d.Buttons
	return d.report()
}

// report builds the pad's input report — the bytes that go on the wire, whether they are
// asked for down the interrupt pipe or fetched once over the control pipe.
func (d *xidDevice) report() []byte {
	var r [xidReportSize]byte
	// [0] and [1] are the header the cook skips over: it copies from the wire buffer +2
	// (0x2438EF) for size-2 bytes, so these two bytes are read by nothing at all. They
	// are absurd on purpose — the only two bytes of the report that provably cannot
	// matter, and if that ever stops being true this is what will say so.
	r[0], r[1] = 0xF5, 0xF6

	// [2..3] -> gamepad+0: wButtons. The BITS are the title's own, off its remapping at
	// 0x14630 (TEST CL,$10 -> OR EAX,$01 is START), which is where padButtons above got
	// them. The title reads the full word but tests only the low byte, and XAPI's own
	// idle check reads only the low six bits (0x2438DA's TEST BYTE [EDX],$3F), so the
	// high byte carries nothing this pad knows how to say.
	r[2], r[3] = byte(d.Buttons), byte(d.Buttons>>8)

	// [4..0xB] -> gamepad+2..+9: the eight analog buttons, as PRESSURE bytes. Zero is
	// "released", and it is a LEVEL WE ARE SETTING, not a hardware default and not a
	// gap in the model: this pad's vocabulary is digital-only for now (see padButtons),
	// so no name reaches these bytes and every one of them is honestly at rest. XAPI
	// zeroes anything under 0x20 here anyway (0x24390A) and the title thresholds them
	// again itself (0x147DE), so 0 reads as released on both sides.
	//
	// [0xC..0x13] -> gamepad+0xA..+0x11: four signed words, the sticks, centred. Same
	// point, and it is worth making twice: a centred stick and an unmodelled stick
	// produce identical bytes. These are centred. The distinction is invisible in the
	// report and must stay visible here.
	return r[:]
}
