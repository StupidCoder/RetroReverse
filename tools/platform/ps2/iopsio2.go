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
)

// sio2Contains reports whether an address is one of the SIO2 controller's registers.
func sio2Contains(a uint32) bool { return a >= iopSIO2Base && a < iopSIO2End }

// sio2Read serves a read of a SIO2 register.
//
// Only DATA_OUT is special: with no device on the port the response line floats, so a read
// of the FIFO returns 0xFF — open bus. mcman ignores those bytes on the no-card path (it has
// already decided from RECV1), but returning a settled value rather than whatever was last
// merged into the word keeps the FIFO from reading back a fragment of a command. Everything
// else — RECV1's "no device" word, the SEND parameters, STAT — is plain register storage,
// written by the start behaviour or by sio2man itself and read straight back.
func (p *IOP) sio2Read(a uint32) uint32 {
	if a == sio2DATAout {
		return 0xFF
	}
	return p.io[a]
}

// sio2Write serves a write of a SIO2 register.
//
// The one behaviour is the start bit. When sio2man's worker writes CTRL with bit 0 set, the
// controller runs the queued transfer; with no card in the slot there is no device to answer,
// so the transfer completes immediately with RECV1 reporting "no device" and interrupt 17
// raised for sio2man's handler to see. The start bit is cleared in the stored value, as the
// hardware clears it when the transfer finishes — sio2man's reset (CTRL|0xC) and config
// (CTRL=0x3BC) writes leave bit 0 clear and so do not trigger this.
func (p *IOP) sio2Write(a, v uint32) {
	if a == sio2CTRL {
		p.io[a] = v &^ 1
		if v&1 != 0 {
			p.sio2Start()
		}
		return
	}
	p.io[a] = v
}

// sio2Start completes a transfer to an empty slot: report "no device" and raise the
// completion interrupt.
//
// The interrupt is raised now rather than after a latency, and that is safe because an event
// flag is level, not edge: whether interrupt 17 is delivered before the worker reaches its
// WaitEventFlag or after it has gone to sleep, the handler's iSetEventFlag(0x2000) leaves the
// bit set and the wait returns. RECV1 is settled before the interrupt so the worker reads the
// right word when it wakes.
func (p *IOP) sio2Start() {
	p.io[sio2RECV1] = sio2NoDevice
	p.raiseIRQ(iopSIO2IRQ)
}
