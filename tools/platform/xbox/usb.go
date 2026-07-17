package xbox

// usb.go models the MCPX southbridge's USB OHCI host controller at 0xFED00000 — the
// controller the console's four game-pad ports hang off. It is the machine's input
// front door, and the GameCube's serial interface is its closest sibling in this repo:
// a device the title programs once and then reads for ever.
//
// It began as a latch with two spec answers (apu.go's usbRead), which was honest while
// nothing was plugged in. What the latch could not survive was being BELIEVED. Two of
// its registers are not storage, and echoing a write back at the driver made the
// machine assert things no controller would:
//
//   - HcRhDescriptorA's NDP (bits 0-7) is the root hub's port count, read-only and set
//     by the hardware. XAPI writes NPS|NOCP into that register to configure power
//     switching; the latch echoed the write, so the hub answered "I have zero ports".
//   - HcInterruptEnable is write-1-to-SET, not storage. XAPI enables MIE|FNO|UE|WDH|SO
//     (0x80000033) and later RHSC (0x40); the latch REPLACED, so the enable mask ended
//     as 0x40 with the master interrupt enable clear — a controller that may never
//     signal anything, which is exactly the state every pre-Phase-E savestate holds.
//
// Neither fiction announced itself. Both read back as plausible numbers, and the second
// is why the snapshots taken before this file existed cannot reach a pad: a lost enable
// bit is a lost interrupt for ever, the same shape as the GameCube's dropped DSP
// interrupt. The fixtures were re-derived from a cold boot rather than patched.
//
// WHAT IS DERIVED, AND FROM WHERE. Generic OHCI register semantics are platform
// knowledge, on the footing apu.go already claims for HcRevision and the AC'97
// semaphore. But the facts that actually shape this model came from XAPI's own code,
// traced rather than remembered — which matters, because every one of them is a number
// it would have been easy to write down from the wrong kind of knowledge:
//
//	00240105  MOV BYTE [ESI+$460], $04   the port count is FOUR, hardcoded by XAPI.
//	                                     It never reads NDP at all.
//	00240133  MOV EDI, [EAX]             the port loop, over 0x54/0x58/0x5C/0x60
//	00240138  TEST BYTE [EBP-$8], $01    CCS — "connected" — is BIT 0
//	0024013E  OR  [EBP-$2], CX           a connected port sets its bit...
//	0024015B  SHL ECX, 1                 ...and the port bit is 1<<port
//	00240146  AND WORD [EBP-$8], $0000   it clears the LOW half and writes back:
//	0024014E  MOV [EAX], EDI             the CHANGE bits (16-31) are write-1-to-clear
//	00240161  MOV DWORD [EBX+$10], $40   then it re-enables RHSC and waits
//
// So the driver's whole contract is: report a connection in a port's CCS, raise
// RootHubStatusChange, and it will come and look. Interrupt vector 1 is likewise
// pinned rather than assumed — the title's own KeConnectInterrupt names it, and its
// ISR (0x245DC2) lives in the same XAPI code region as every access traced above.
//
// The frame counter is DERIVED from the machine tick, never counted. A counter that
// advances once per call drifts the instant a savestate skips a call, and savestate
// fidelity outranks a register nobody times against (the apuHandshake precedent).
//
// The register file's backing store is still m.usb.reg — the latch's sparse map — and
// deliberately so: that map IS the savestate's USBReg, so the controller XAPI
// programmed during the cold boot (HcControl=0xBE operational, HcHCCA=0x013EA310, the
// address-0 control ED at 0x013EB130) survives a restore without a state-version bump.
// Semantics live here; storage stays where it was.

// The OHCI operational register map (offsets within the aperture).
const (
	hcRevision         = 0x00
	hcControl          = 0x04
	hcCommandStatus    = 0x08
	hcInterruptStatus  = 0x0C
	hcInterruptEnable  = 0x10
	hcInterruptDisable = 0x14
	hcHCCA             = 0x18
	hcPeriodCurrentED  = 0x1C
	hcControlHeadED    = 0x20
	hcControlCurrentED = 0x24
	hcBulkHeadED       = 0x28
	hcBulkCurrentED    = 0x2C
	hcDoneHead         = 0x30
	hcFmInterval       = 0x34
	hcFmRemaining      = 0x38
	hcFmNumber         = 0x3C
	hcPeriodicStart    = 0x40
	hcLSThreshold      = 0x44
	hcRhDescriptorA    = 0x48
	hcRhDescriptorB    = 0x4C
	hcRhStatus         = 0x50
	hcRhPortStatus1    = 0x54 // ports 1..4 at 0x54, 0x58, 0x5C, 0x60

	// usbPorts is the root hub's downstream port count — the console's four pad ports.
	// Not remembered: XAPI stores this exact literal into its own driver object at
	// 0x240105 and loops over it, so four is what the guest believes whatever we say.
	usbPorts = 4
)

// HcControl. The title's own value is 0xBE: operational, with every list enabled.
const (
	ctrlPLE = 1 << 2 // periodic list enable
	ctrlIE  = 1 << 3 // isochronous enable
	ctrlCLE = 1 << 4 // control list enable
	ctrlBLE = 1 << 5 // bulk list enable
	ctrlIR  = 1 << 8 // interrupt routing: SMM owns the controller

	hcfsMask        = 3 << 6 // host controller functional state
	hcfsReset       = 0 << 6
	hcfsResume      = 1 << 6
	hcfsOperational = 2 << 6
	hcfsSuspend     = 3 << 6
)

// HcInterruptStatus / Enable / Disable share one bit layout.
const (
	ohciIntSO   = 1 << 0  // scheduling overrun
	ohciIntWDH  = 1 << 1  // writeback done head
	ohciIntSF   = 1 << 2  // start of frame
	ohciIntRD   = 1 << 3  // resume detected
	ohciIntUE   = 1 << 4  // unrecoverable error
	ohciIntFNO  = 1 << 5  // frame number overflow
	ohciIntRHSC = 1 << 6  // root hub status change — the one the pad arrives on
	ohciIntOC   = 1 << 30 // ownership change
	ohciIntMIE  = 1 << 31 // master interrupt enable (enable register only)
)

// HcRhPortStatus, read side. The low half is state, the high half is change flags.
const (
	portCCS  = 1 << 0  // current connect status — XAPI's TEST BYTE [..],$01
	portPES  = 1 << 1  // port enable status
	portPSS  = 1 << 2  // port suspend status
	portPOCI = 1 << 3  // port over-current indicator
	portPRS  = 1 << 4  // port reset status
	portPPS  = 1 << 8  // port power status
	portLSDA = 1 << 9  // low speed device attached
	portCSC  = 1 << 16 // connect status change
	portPESC = 1 << 17 // port enable status change
	portPSSC = 1 << 18 // port suspend status change
	portOCIC = 1 << 19 // over-current indicator change
	portPRSC = 1 << 20 // port reset status change

	portChangeMask = portCSC | portPESC | portPSSC | portOCIC | portPRSC
)

// HcRhPortStatus, write side. A write is a command, not a store: each bit asks for one
// action, and writing zero asks for nothing. The change bits (16-20) are write-1-to-
// clear, which is the half XAPI acks in its port loop.
const (
	portWClearEnable  = 1 << 0
	portWSetEnable    = 1 << 1
	portWSetSuspend   = 1 << 2
	portWClearSuspend = 1 << 3
	portWSetReset     = 1 << 4
	portWSetPower     = 1 << 8
	portWClearPower   = 1 << 9
)

// usbFrame is the current USB frame number: one per millisecond, derived from the
// machine tick so it is monotonic and identical across a savestate round trip.
func (m *Machine) usbFrame() uint64 { return m.tick / instrsPerMs }

// usbRead answers a byte read from the OHCI aperture. Registers with behaviour are
// computed here; everything else reads back what was written, through the latch (which
// keeps the log-once cold-register guard and the RR_APU_TRACE trace).
func (m *Machine) usbRead(off uint32) byte {
	dw, written := m.usb.reg[off>>2]
	switch off &^ 3 {
	case hcRevision:
		dw, written = 0x10, true // OHCI 1.0

	case hcCommandStatus:
		// HCR, CLF, BLF and OCR are self-clearing command bits, and this model does
		// their work synchronously inside the write — so by the time anything can read
		// them back, they have all completed. SOC (bits 16-17) is 0: no overrun.
		dw, written = 0, true

	case hcInterruptDisable:
		// Not a register of its own: the disable port reads the same enable mask it
		// clears bits in. Both windows are views of one value.
		dw, written = m.usb.reg[hcInterruptEnable>>2], true

	case hcFmNumber:
		dw, written = uint32(m.usbFrame()&0xFFFF), true

	case hcFmRemaining:
		// Ticks left in the current frame, counting down from FrameInterval. Derived
		// from the same clock as the frame number, so the two cannot disagree.
		fi := m.usb.reg[hcFmInterval>>2] & 0x3FFF
		if fi != 0 {
			pos := m.tick % instrsPerMs
			rem := fi - uint32(uint64(fi)*pos/instrsPerMs)
			dw = rem | (m.usb.reg[hcFmNumber>>2]&1)<<31 // FRT toggles with the frame
		}
		written = true

	case hcRhDescriptorA:
		// NDP is the hardware's own port count and is read-only. Overlaying it on the
		// written value is the whole point: the driver writes NPS|NOCP into this
		// register, and a latch that echoed the write made the hub claim no ports.
		dw = (dw &^ 0xFF) | usbPorts
		written = true
	}
	return m.latchRead(&m.usb, off, dw, written)
}

// usbWrite applies a byte write to the OHCI aperture.
//
// The guest writes dwords, which arrive here as four byte calls, and every set/clear
// rule below is bitwise — so applying each byte to its own slice of the register is
// exactly equivalent to applying the dword once, with no need to buffer.
func (m *Machine) usbWrite(off uint32, v byte) {
	reg := off &^ 3
	bits := uint32(v) << (8 * (off & 3)) // this byte's bits, in register position

	// Reassemble the dword the guest is writing, purely so the trace can report the
	// command rather than its aftermath. It cannot straddle a savestate: a store is one
	// instruction, and snapshots are taken between them.
	if off&3 == 0 {
		m.usbWrDword = 0
	}
	m.usbWrDword |= bits
	wrote := m.usbWrDword

	switch reg {
	case hcInterruptEnable:
		// Write-1-to-set. Writing 0 to a bit leaves it alone; this is how XAPI's two
		// separate enables (0x80000033 then 0x40) accumulate rather than replace.
		m.usb.reg[hcInterruptEnable>>2] |= bits
		m.latchTrace(&m.usb, off, wrote)
		m.usbUpdateIRQ()
		return

	case hcInterruptDisable:
		// Write-1-to-clear, of the enable mask the read side shows.
		m.usb.reg[hcInterruptEnable>>2] &^= bits
		m.latchTrace(&m.usb, off, wrote)
		m.usbUpdateIRQ()
		return

	case hcInterruptStatus:
		// Write-1-to-clear: this is how an ISR acks what it has handled. A latch here
		// would have made the acknowledgement raise the interrupt instead of ending it.
		m.usb.reg[hcInterruptStatus>>2] &^= bits
		m.latchTrace(&m.usb, off, wrote)
		m.usbUpdateIRQ()
		return

	case hcCommandStatus:
		// HostControllerReset returns the controller to its cold state. Nothing here
		// keeps state across it yet beyond the registers themselves, and XAPI programs
		// everything after the reset anyway (it is the third thing the trace shows) —
		// but the bit must not latch, or a later read would report a reset in progress
		// for ever.
		m.latchTrace(&m.usb, off, wrote)
		return

	case hcDoneHead:
		// Read-only: the controller publishes retired TDs here. XAPI writes 0 during
		// init (the trace's 0x24037A); accepting the store is harmless and honest,
		// since 0 is also what an idle controller reports.
		m.latchWrite(&m.usb, off, v)
		return
	}

	if reg >= hcRhPortStatus1 && reg < hcRhPortStatus1+4*usbPorts {
		m.usbPortWrite((reg-hcRhPortStatus1)/4, bits)
		m.latchTrace(&m.usb, off, wrote)
		return
	}

	m.latchWrite(&m.usb, off, v)
}

// usbPortWrite applies a root hub port command. A port write is a request — one action
// per bit — not a value to store, so bits that are not set do nothing at all.
func (m *Machine) usbPortWrite(port, bits uint32) {
	off := hcRhPortStatus1 + port*4
	st := m.usb.reg[off>>2]
	was := st

	switch {
	case bits&portWClearEnable != 0:
		st &^= portPES
	case bits&portWSetEnable != 0:
		if st&portCCS != 0 { // a port with nothing in it cannot be enabled
			st |= portPES
		}
	}
	if bits&portWSetSuspend != 0 && st&portCCS != 0 {
		st |= portPSS
	}
	if bits&portWClearSuspend != 0 {
		st &^= portPSS
	}
	if bits&portWSetReset != 0 && st&portCCS != 0 {
		// Reset completes synchronously here: the model has no way to be mid-reset
		// between two guest instructions, and a driver that polls PRS wants to see it
		// finish. The port comes out enabled, with the reset-change flag raised.
		st &^= portPRS
		st |= portPES | portPRSC
	}
	if bits&portWSetPower != 0 {
		st |= portPPS
	}
	if bits&portWClearPower != 0 {
		st &^= portPPS | portPES
	}

	// The change bits are write-1-to-clear — the half XAPI's port loop acks.
	st &^= bits & portChangeMask

	m.usb.reg[off>>2] = st

	// A port that has just raised a CHANGE has to tell the driver so. EVERY change bit is
	// a RootHubStatusChange, not only the connect — and the omission is worth recording,
	// because it did not look like a bug from in here. The reset above completed and set
	// its PRSC exactly as the spec says; what it did not do was knock on the door.
	//
	// XAPI resets a port and then WAITS for the completion interrupt, with a 100 ms timer
	// as its retry (0x245824 arms a second KTIMER right after the reset write). So a
	// silent PRSC produced a machine that reset the port, timed out, reset it again, for
	// ever — a loop in which every individual step was correct. The symptom was not "no
	// reset". It was too many.
	if st&^was&portChangeMask != 0 {
		m.usbRaise(ohciIntRHSC)
		return
	}
	m.usbUpdateIRQ()
}

// usbSetPortConnected attaches or detaches a device on a root hub port, raising the
// connect-status change and the RootHubStatusChange interrupt that makes the driver
// come and look.
//
// It is the only way a device ever appears, and it is deliberately the last thing the
// port model gained: a port that reports a connection the rest of the controller cannot
// then service is worse than an empty one, because XAPI would build an enumeration
// chain, get no answer, and settle into a "device attached, not enumerated" state whose
// pad reads as a pad with no buttons held. That is byte-identical to having no pad at
// all — except that we would believe we had one.
func (m *Machine) usbSetPortConnected(port uint32, connected bool) {
	if port >= usbPorts {
		return
	}
	off := hcRhPortStatus1 + port*4
	st := m.usb.reg[off>>2]
	was := st&portCCS != 0
	if was == connected {
		return
	}
	if connected {
		st |= portCCS
	} else {
		st &^= portCCS | portPES
	}
	st |= portCSC // the change flag the driver acks after it has looked
	m.usb.reg[off>>2] = st
	m.usbRaise(ohciIntRHSC)
}

// usbRaise sets an interrupt status bit and re-evaluates the controller's interrupt
// line. The bit stays up until the ISR write-1-clears it, which is what lets delivery
// be retried while the CPU's gates are shut.
func (m *Machine) usbRaise(bit uint32) {
	m.usb.reg[hcInterruptStatus>>2] |= bit
	m.usbUpdateIRQ()
}

// usbIRQ reports whether the controller is asserting its interrupt line: some enabled
// source is pending, and the master interrupt enable is set.
//
// MIE is not decoration. A restored pre-Phase-E savestate holds an enable mask of 0x40
// — RHSC without MIE — because the old latch dropped the first of XAPI's two enable
// writes, and this predicate is what makes that state's silence honest rather than a
// mystery: such a controller genuinely never signals.
func (m *Machine) usbIRQ() bool {
	en := m.usb.reg[hcInterruptEnable>>2]
	if en&ohciIntMIE == 0 {
		return false
	}
	return m.usb.reg[hcInterruptStatus>>2]&en&^ohciIntMIE != 0
}

// usbUpdateIRQ re-evaluates the line after anything that could change it. Delivery
// itself belongs to interrupt.go, which owns the CPU's gates.
func (m *Machine) usbUpdateIRQ() {
	if m.usbIRQ() {
		m.deliverPending()
	}
}

// usbTick services the controller's frame boundaries. It runs from schedTick's coarse
// block (every 1024 instructions) while a USB frame is 2000, so it must catch up rather
// than assume it was called once per frame: it services every frame that has elapsed
// since the last, which costs at worst one frame of jitter and never loses one.
//
// The bound is not decoration either. usbFrameServed rides in the savestate, and a
// state restored beside a machine tick that has moved on would otherwise spin here for
// as many frames as the gap.
func (m *Machine) usbTick() {
	now := m.usbFrame()
	if m.usbFrameServed == 0 || now < m.usbFrameServed {
		m.usbFrameServed = now // first run, or a restore that moved the clock backwards
	}
	if n := now - m.usbFrameServed; n > 64 {
		m.usbFrameServed = now - 64
	}
	for m.usbFrameServed < now {
		m.usbFrameServed++
		m.usbFrameTick()
	}
	// A pending interrupt may have been raised while the CPU's gates were shut (an
	// active frame, IF clear, a raised IRQL); retry while it stands, the same
	// discipline vblankTick follows.
	if m.usbIRQ() {
		m.deliverPending()
	}
}

// usbFrameTick is one USB frame's worth of controller work: publish the frame number,
// walk the lists the driver has enabled, and hand back whatever retired.
func (m *Machine) usbFrameTick() {
	ctrl := m.usb.reg[hcControl>>2]
	if ctrl&hcfsMask != hcfsOperational {
		return // the controller is not running; a frame is not its business
	}
	m.usbFrameNumberToHCCA()

	// StartofFrame, every frame, whether or not anyone is listening. The status bit is
	// what the hardware DID; the enable mask only decides whether it pulls the line, and
	// usbIRQ already gates on that — so raising it unconditionally is not noise, it is
	// the difference between a frame boundary happening and a frame boundary being
	// announced.
	//
	// XAPI waits on this to retire an endpoint descriptor safely. Having enumerated the
	// pad at address 0, it must move the default control pipe to address 1, and an ED
	// cannot be edited under a running controller — so it unlinks the ED from the control
	// list and waits one frame for the hardware to let go of it before touching FA:
	//
	//	002459DB  HcControlHeadED  <- 013EB130   unlink the address-0 ED
	//	0024479D  HcInterruptStatus <- 04        clear any SF already standing...
	//	002447A2  HcInterruptEnable <- 04        ...and only then listen for one
	//
	// That order is the proof, and it is worth more than the spec sentence it agrees
	// with: XAPI clears SF before it enables it. A controller that only set SF when
	// someone had asked to hear it would have nothing to clear, and the write would be
	// pointless — the driver writes it because on real hardware SF has been setting every
	// millisecond all along, and a stale one would fire the instant it unmasked. Without
	// this the pad enumerated, took its address, and stopped: XAPI unlinked the pipe and
	// waited forever for a frame that was silently already happening.
	m.usbRaise(ohciIntSF)

	// Each list is walked only when the driver has enabled it. Walking a disabled list
	// would run transfers the driver believes are parked — the lists are its scheduling,
	// not just its storage.
	if ctrl&ctrlCLE != 0 {
		m.usbWalkList(m.usb.reg[hcControlHeadED>>2])
	}
	if ctrl&ctrlBLE != 0 {
		m.usbWalkList(m.usb.reg[hcBulkHeadED>>2])
	}
	if ctrl&ctrlPLE != 0 {
		m.usbWalkPeriodic()
	}
	m.usbWriteback()
}

// usbWalkPeriodic walks the interrupt endpoints for this frame. The HCCA holds a
// 32-entry table the driver populates; the hardware picks the entry by the low 5 bits
// of the frame number, which is how a driver spreads polling across frames.
func (m *Machine) usbWalkPeriodic() {
	hcca := m.usb.reg[hcHCCA>>2]
	if hcca == 0 {
		return
	}
	m.usbWalkList(m.read32(hcca + (uint32(m.usbFrame())&31)*4))
}
