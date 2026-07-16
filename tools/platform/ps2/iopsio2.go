package ps2

// iopsio2.go is the SIO2 controller — the serial front end the memory cards and the
// controllers hang off — and it is here for one reason: without it the title screen never
// arrives.
//
// The chain was long and every link was healthy but the last. Jak's title process
// (target-title) parks in a loop that calls mc-get-slot-info(slot 0) every frame and
// refuses to leave — refreshing the game's own set-blackout-frames(15) forever — until the
// memory card in slot 0 is *known*, present or not. mc-get-slot-info is MC_get_status, which
// reads a per-slot state word the game's MC_run state machine fills; MC_run drives it with
// sceMcGetInfo/sceMcSync, an async RPC to mcserv on the IOP; mcserv asks mcman to probe the
// slot; and mcman probes it by handing sio2man a command to run on this controller and
// waiting for the answer. On the board the controller runs the transfer, gets no ACK from an
// empty slot, and raises **interrupt 17**; sio2man's handler wakes the worker, which reads
// RECV1 and hands mcman a "no device" that it turns into "no card". Interrupt 17 is the whole
// mechanism, and it was never raised, because nothing modelled the silicon that raises it.
// So mcman's worker slept on an event flag forever, mcserv never replied, the EE's McThread
// blocked, MC_run never advanced, the slot stayed *unknown*, and the screen stayed black —
// a machine doing exactly what it should and getting nowhere, which is the shape of every
// bug on this processor.
//
// Everything below was read off sio2man's and mcman's own code (they are on the disc and run
// for real), not looked up:
//
//	the register map — from sio2man's one-instruction accessor functions, each a
//	  `lui 0xBF80 / ori 0x82xx / lw|sw`:
//	    0x8268 CTRL   write starts a transfer (bit 0); read polls; reset writes bit 2|3
//	    0x826C RECV1  the port's device-status word
//	    0x8270 RECV2  0x8274 RECV3
//	    0x8240+n*8 SEND1   0x8244+n*8 SEND2   0x8200+n*4 SEND3   (command parameters)
//	    0x8260 DATA_IN (byte FIFO)   0x8264 DATA_OUT (byte FIFO)
//	    0x8280 STAT   interrupt cause; the handler reads it and writes it back to ack
//	the completion contract — from sio2man's worker (sio2man+0x4FC..): it writes CTRL|1 to
//	  start, then WaitEventFlag(bit 0x2000); the IRQ-17 handler (sio2man+0x550) reads STAT,
//	  writes it back, and iSetEventFlag(0x2000). The handler never branches on STAT's value.
//	the "no card" answer — from mcman's detect (mcman_tool+0x2498): after the transfer it
//	  reads RECV1 and tests `(RECV1 & 0xF000) == 0x1000` for "device present"; anything else
//	  is retried five times and then reported as -1, "no card". 0x1D100 — the controller's
//	  own "no device" word — fails that test (0x1D100 & 0xF000 = 0xD000), so mcman concludes
//	  the slot is empty, which is the true state of a console with no card in it.
//
// So the model is small and it is honest: a register file, a single behaviour on the start
// bit, and one interrupt. It does not pretend a card is there and it does not invent an
// answer; it reports the empty slot the hardware would, and lets sio2man and mcman — real
// code, both — draw the conclusion themselves.

// The SIO2 register window. IOPRegionName already calls this range "SIO2".
const (
	iopSIO2Base = 0x1F808200
	iopSIO2End  = 0x1F808284

	sio2SEND3   = 0x1F808200 // 16 words: the transfer's command queue, one entry per command
	sio2DATAin  = 0x1F808260 // the command byte FIFO (PIO side; sio2man here uses DMA ch11)
	sio2CTRL    = 0x1F808268 // write bit 0 = start; hardware clears it on completion
	sio2RECV1   = 0x1F80826C // the probed port's device status
	sio2DATAout = 0x1F808264 // the response byte FIFO

	// The IOP interrupt line sio2man registers its completion handler on
	// (RegisterIntrHandler(17) / EnableIntr(17), at sio2man+0x740).
	iopSIO2IRQ = 17

	// RECV1 for an empty slot. mcman's detect keeps a slot only if (RECV1 & 0xF000) ==
	// 0x1000; 0x1D100 is the controller's "no device" word and fails that test, so mcman
	// reads it as "no card" — the honest state of a console with nothing plugged in.
	sio2NoDevice = 0x0001D100

	// RECV1 for a port with a device on it: the same test's passing value. The pad on
	// port 0 answers with this, and mcman's card ports keep getting sio2NoDevice.
	sio2Device = 0x00001100
)

// sio2xfer is the state of the transfer in flight: what the DMA fed in, what the devices
// answered, and the slice-DMA parameters sio2man armed before writing the start bit.
// It is savestated — a snapshot between the arm and the start bit must resume mid-shape.
type sio2xfer struct {
	in       []byte    // command bytes, fed by DMA ch11 (or PIO pushes to DATA_IN)
	out      []byte    // response bytes, drained by DMA ch12 (or PIO pops of DATA_OUT)
	dmaAddr  [2]uint32 // armed slice-DMA address for ch11 (in) and ch12 (out)
	dmaWords [2]uint32 // armed length in words
	outArmed bool      // ch12 is armed: flush the response to RAM when the transfer runs
	dumps    int       // the first-transfers instrument: dump SEND3 + bytes, a few times
}

// sio2PadState is the port-0 controller's own mode latches. A DualShock-class pad is a
// tiny state machine: command 0x43 moves it in and out of config (escape) mode, where it
// answers with id 0xF3 and serves the query commands; 0x44 switches digital/analog. The
// latches are device state, not transfer state, and they are savestated.
//
// The pad must be config-capable because padman refuses anything less: its identify
// sequence (padman+0x156C) queries the pad in config mode and returns success only if
// the responses carry the config id 0xF3 (the check at padman+0x16A8) — a pad answering
// like a plain digital controller fails the sequence ten times (padman+0x26D4 loop), the
// connect thread exits and restarts, and the port sits at state 5 (EXECCMD) forever,
// which the EE reads as "never stable" and no button ever arrives.
type sio2PadState struct {
	config bool    // in config mode: id 0xF3, query commands served
	analog bool    // analog mode (set by 0x44): id 0x73, stick bytes appended
	locked bool    // mode locked by 0x44's second parameter
	actMap [6]byte // actuator mapping set by 0x4D; the reply echoes the previous mapping
}

// sio2Contains reports whether an address is one of the SIO2 controller's registers.
func sio2Contains(a uint32) bool { return a >= iopSIO2Base && a < iopSIO2End }

// sio2Read serves a read of a SIO2 register.
//
// DATA_OUT pops the response FIFO; with nothing queued the line floats and a read returns
// 0xFF — open bus. mcman ignores those bytes on the no-card path (it has already decided
// from RECV1). Everything else — RECV1, the SEND parameters, STAT — is plain register
// storage, written by the start behaviour or by sio2man itself and read straight back.
func (p *IOP) sio2Read(a uint32) uint32 {
	if a == sio2DATAout {
		if len(p.sio2.out) > 0 {
			b := uint32(p.sio2.out[0])
			p.sio2.out = p.sio2.out[1:]
			return b
		}
		return 0xFF
	}
	return p.io[a]
}

// sio2Write serves a write of a SIO2 register.
//
// Two behaviours. DATA_IN pushes a command byte into the FIFO (the PIO road; Jak's sio2man
// moves its pad batches by DMA channel 11 instead, but the register is the same FIFO). The
// start bit: when sio2man's worker writes CTRL with bit 0 set, the controller runs the
// queued transfer — walks SEND3, feeds each command to the device on its port, queues the
// answers — and raises interrupt 17 for sio2man's handler. The start bit is cleared in the
// stored value, as the hardware clears it when the transfer finishes — sio2man's reset
// (CTRL|0xC) and config (CTRL=0x3BC) writes leave bit 0 clear and so do not trigger this.
func (p *IOP) sio2Write(a, v uint32) {
	if a == sio2DATAin {
		p.sio2.in = append(p.sio2.in, byte(v))
		return
	}
	if a == sio2CTRL {
		p.io[a] = v &^ 1
		if v&1 != 0 {
			p.sio2Start()
		}
		return
	}
	p.io[a] = v
}

// sio2Start runs the queued transfer: walk the SEND3 command queue, hand each command's
// bytes to the device on its port, and queue what it answers.
//
// The port map is the console's: 0 and 1 are the controllers, 2 and 3 the memory cards.
// This machine has a pad in port 0 (sio2Pad) and nothing anywhere else, so a port-0
// command gets a controller's answer and RECV1 reports a device, while a card probe keeps
// getting the empty-slot word that the whole 20th-pass chain was built to deliver.
//
// The SEND3 entry's layout is read off the transfers padman itself queues (dumped by the
// first-transfers instrument below): bits 0-1 the port, bits 8-16 the command's length in
// bytes; one entry per command, a zero entry ends the queue. The same length serves both
// directions — every pad command is answered byte for byte.
//
// The interrupt is raised now rather than after a latency, and that is safe because an
// event flag is level, not edge: whether interrupt 17 is delivered before the worker
// reaches its WaitEventFlag or after it has gone to sleep, the handler's
// iSetEventFlag(0x2000) leaves the bit set and the wait returns. RECV1 is settled before
// the interrupt so the worker reads the right word when it wakes.
func (p *IOP) sio2Start() {
	in := p.sio2.in
	p.sio2.in = nil
	p.sio2.out = p.sio2.out[:0]
	recv1 := uint32(sio2NoDevice)

	dump := p.sio2.dumps < 4 && len(in) > 0
	if dump {
		p.sio2.dumps++
		p.ps2.note("SIO2: transfer, %d command bytes", len(in))
	}

	off := 0
	for i := 0; i < 16; i++ {
		s3 := p.io[sio2SEND3+uint32(i)*4]
		if s3 == 0 {
			break
		}
		port := s3 & 3
		n := int(s3>>8) & 0x1FF
		if dump {
			end := off + n
			if end > len(in) {
				end = len(in)
			}
			p.ps2.note("SIO2:   send3[%d] = %08X (port %d, len %d): % x", i, s3, port, n, in[off:end])
		}
		if n == 0 {
			continue
		}
		cmd := []byte{}
		if off < len(in) {
			end := off + n
			if end > len(in) {
				end = len(in)
			}
			cmd = in[off:end]
		}
		off += n

		if port == 0 {
			p.sio2.out = append(p.sio2.out, p.sio2Pad(cmd, n)...)
			recv1 = sio2Device
		} else {
			// No device: the response line floats and the FIFO reads back 0xFF.
			for j := 0; j < n; j++ {
				p.sio2.out = append(p.sio2.out, 0xFF)
			}
		}
	}

	p.io[sio2RECV1] = recv1

	// The response's DMA half: sio2man armed channel 12 before it wrote the start bit,
	// so the answer is flushed to its buffer now, as the hardware's slice DMA would have
	// drained the FIFO while the transfer ran.
	if p.sio2.outArmed {
		p.sio2.outArmed = false
		a := iopPhys(p.sio2.dmaAddr[1])
		n := int(p.sio2.dmaWords[1] * 4)
		if n > len(p.sio2.out) {
			n = len(p.sio2.out)
		}
		for j := 0; j < n && int(a)+j < len(p.ram); j++ {
			p.ram[int(a)+j] = p.sio2.out[j]
		}
		if dump {
			d := n
			if d > 24 {
				d = 24
			}
			p.ps2.note("SIO2:   response, %d bytes to 0x%X: % x", n, p.sio2.dmaAddr[1], p.sio2.out[:d])
		}
		p.sio2.out = p.sio2.out[n:]
	}

	p.raiseIRQ(iopSIO2IRQ)
}

// sio2Pad is the controller in port 0: a DualShock 2. It speaks the pad protocol's shape —
// a command frame in, a same-length frame out, header 0xFF, then the pad's identity byte
// and 0x5A ("here it comes"), then the payload, buttons active-low, little end first.
//
// It began life as a plain digital pad (id 0x41 to everything), and padman rejected it:
// the identify sequence padman runs after the first successful poll (padman+0x156C,
// called from the connect thread at padman+0x26D4) puts the pad in config mode with 0x43
// and queries it — model 0x45, actuators 0x46/0x47, mode table 0x4C — and returns success
// ONLY if those replies carry the config-mode id 0xF3 (padman+0x16A8). A digital pad's
// replies fail that check ten times, the connect thread exits and restarts, the port
// state ping-pongs between 5 (EXECCMD) and re-probe, and the EE — which polls
// scePadGetState wanting 6 (STABLE) — never sees a button. So the pad Jak actually
// shipped against is what sits here: config mode, the query set, and the analog switch,
// with the latches in sio2PadState. The query replies are the standard DualShock 2
// answers from the pad-protocol documentation; padman's parsers are the verification
// (its acceptance check at padman+0x1704 requires the model reply's mode count < 4,
// which the honest 0x02 satisfies).
//
// The buttons come from the machine's injection schedule (Machine.padButtons), which is
// how the oracle presses X on a dialog: the answer is derived from the vblank count, so it
// is reproducible run to run and needs no state in the snapshot.
func (p *IOP) sio2Pad(cmd []byte, n int) []byte {
	resp := make([]byte, n)
	for i := range resp {
		resp[i] = 0xFF
	}
	if n < 2 || len(cmd) < 2 {
		return resp
	}

	// The identity byte: high nibble = type, low nibble = payload halfwords. 0x41 is
	// the digital pad (1 halfword of buttons), 0x73 the DualShock in analog mode
	// (buttons + 4 stick bytes), 0xF3 config mode (3 halfwords of command payload).
	id := byte(0x41)
	if p.pad.analog {
		id = 0x73
	}
	if p.pad.config {
		id = 0xF3
	}
	resp[1] = id
	if n > 2 {
		resp[2] = 0x5A
	}

	// put writes one payload byte, if the transfer is long enough to carry it.
	put := func(i int, b byte) {
		if 3+i < n {
			resp[3+i] = b
		}
	}
	zero := func(k int) {
		for i := 0; i < k; i++ {
			put(i, 0)
		}
	}
	arg := func(i int) byte {
		if 3+i < len(cmd) {
			return cmd[3+i]
		}
		return 0
	}
	pollData := func() {
		buttons := p.ps2.padButtons()
		put(0, ^byte(buttons))
		put(1, ^byte(buttons>>8))
		if p.pad.analog {
			// Right stick, then left, both centred: the oracle's pad has no sticks.
			put(2, 0x80)
			put(3, 0x80)
			put(4, 0x80)
			put(5, 0x80)
		}
	}

	switch cmd[1] {
	case 0x42: // poll
		pollData()

	case 0x43: // config (escape) mode enter/exit; the data half is a normal poll
		if p.pad.config {
			zero(6) // inside config mode the payload is zeros
		} else {
			pollData()
		}
		p.pad.config = arg(0) == 1

	case 0x44: // set mode: arg0 = analog, arg1 = 3 locks it (config mode only)
		if p.pad.config {
			zero(6)
			p.pad.analog = arg(0) == 1
			p.pad.locked = arg(1) == 3
		}

	case 0x45: // query model. padman files bytes 1..5 of this at record+194..198 and
		// requires byte 2 (the mode count) < 4 (padman+0x1704) to accept the pad.
		if p.pad.config {
			mode := byte(0)
			if p.pad.analog {
				mode = 1
			}
			put(0, 0x03) // DualShock 2
			put(1, 0x02) // two modes
			put(2, mode)
			put(3, 0x02)
			put(4, 0x01)
			put(5, 0x00)
		}

	case 0x46: // query actuator, two halves selected by the argument
		if p.pad.config {
			if arg(0) == 0 {
				put(0, 0x00)
				put(1, 0x00)
				put(2, 0x01)
				put(3, 0x02)
				put(4, 0x00)
				put(5, 0x0A)
			} else {
				put(0, 0x00)
				put(1, 0x00)
				put(2, 0x01)
				put(3, 0x01)
				put(4, 0x01)
				put(5, 0x14)
			}
		}

	case 0x47: // query actuator combinations
		if p.pad.config {
			zero(6)
			put(2, 0x02)
			put(4, 0x01)
		}

	case 0x4C: // query mode table entry, selected by the argument
		if p.pad.config {
			zero(6)
			if arg(0) == 0 {
				put(3, 0x04)
			} else {
				put(3, 0x07)
			}
		}

	case 0x4D: // map actuators: the reply is the PREVIOUS mapping, then the new one latches
		if p.pad.config {
			for i, b := range p.pad.actMap {
				put(i, b)
			}
			for i := range p.pad.actMap {
				p.pad.actMap[i] = arg(i)
			}
		}

	case 0x4F: // set poll response format (DualShock 2); acknowledged with a trailing 0x5A
		if p.pad.config {
			zero(5)
			put(5, 0x5A)
		}
	}
	return resp
}

// --- the DMA half: dmacman #28/#32/#33/#34, identified from sio2man's own calls -------
//
// sio2man moves a transfer's bytes with the IOP DMA controller's channels 11 (to the
// SIO2) and 12 (from it), through four dmacman functions whose contract is legible in
// one screenful of sio2man:
//
//	init (+0x76C..+0x790):   #33(11, 3); #33(12, 3); #34(11); #34(12)
//	per transfer (+0x2AC):   #28(11, in_dma.addr,  in_dma.size,  in_dma.count) [sp+16]=1
//	                         #32(11, ...)
//	         (+0x2DC):       #28(12, out_dma.addr, out_dma.size, out_dma.count) [sp+16]=0
//	                         #32(12, ...)
//
// The addr/size/count triples are the descriptor fields padman filled (in_dma at +124,
// out_dma at +136); the stack argument is the direction, 1 toward the device. So #28
// parks a slice transfer on a channel, #32 starts it, and #33/#34 are the one-time
// channel setup and enable. The names Sony gave them are not known and not guessed;
// what they must *do* is, and it is done here.
func (p *IOP) dmacmanSetSlice() {
	ch, addr, size, count := p.arg(0), p.arg(1), p.arg(2), p.arg(3)
	if ch == 11 || ch == 12 {
		i := ch - 11
		p.sio2.dmaAddr[i] = addr
		p.sio2.dmaWords[i] = size * count
	}
	p.setRet(1)
}

func (p *IOP) dmacmanStart() {
	ch := p.arg(0)
	switch ch {
	case 11: // IOP RAM -> the SIO2's command FIFO: the bytes are available now
		a := iopPhys(p.sio2.dmaAddr[0])
		n := int(p.sio2.dmaWords[0] * 4)
		for j := 0; j < n && int(a)+j < len(p.ram); j++ {
			p.sio2.in = append(p.sio2.in, p.ram[int(a)+j])
		}
	case 12: // the SIO2's response FIFO -> IOP RAM: armed; drained when the transfer runs
		p.sio2.outArmed = true
	}
	p.setRet(1)
}

// dmacmanChanSetup and dmacmanChanEnable are #33 and #34: called once per channel at
// sio2man's init, (channel, 3) and (channel). Nothing here needs a priority or an enable
// bit — a parked slice moves when #32 says so — so they succeed and are otherwise the
// record that they were identified.
func (p *IOP) dmacmanChanSetup() { p.setRet(0) }

func (p *IOP) dmacmanChanEnable() { p.setRet(0) }
