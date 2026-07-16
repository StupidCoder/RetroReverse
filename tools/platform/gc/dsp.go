package gc

import (
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/gcdsp"
)

// dsp.go is the sound processor's front door: a pair of mailboxes, a control register, and
// the two DMA engines (one for the auxiliary RAM, one for audio) that share its address
// block.
//
// The DSP is a landmine, named in the package doc as the first of the three console-
// resident substitutions, and it is the one that bites soonest. The sound processor runs
// a program — but before it can run the game's program, it runs its own boot ROM, which is
// in the console and not on the disc, and it talks to that boot ROM over these mailboxes.
// OSInit does this handshake early, so a machine whose DSP never answers stalls the whole
// boot in the audio init, long before any sound is wanted.
//
// So this file does not emulate a DSP. It synthesizes the observable behaviour of the boot
// ROM's mailbox handshake — enough that OSInit's audio init completes and hands over to the
// game's own microcode (which IS on the disc, and which a later phase can run for real) —
// and it instruments every unexpected mail so the exact protocol the game expects becomes
// a thing the machine can report rather than a thing that has to be known in advance.
//
// A mailbox is a 32-bit value split into two 16-bit halves. The top bit of the high half is
// the "mail is present" flag: the sender sets it, the receiver reads the value and the act
// of reading (the low half) clears it. Two mailboxes give the two directions.

type dsp struct {
	ToDSP    uint32 // CPU -> DSP; bit 31 = the CPU has posted mail
	FromDSP  uint32 // DSP -> CPU; bit 31 = the DSP has posted mail
	CSR      uint32 // control/status: reset, halt, and the interrupt bits
	BootStep int    // where the synthesized boot-ROM handshake has got to

	// The synthesized boot ROM understands one more thing than its ready handshake: the
	// command sequence that loads a microcode and starts it. Once a ucode is "running" the
	// boot ROM's own re-post-on-halt behaviour is no longer what answers the mailbox.
	UcodeRunning  bool // the boot ROM has loaded a ucode image and jumped to it
	AwaitValue    bool   // the previous mail was a load command; this next mail is its parameter
	AwaitStartArg bool   // the "start" command was seen; its parameter completes the load
	LoadCmd       uint32 // the load command whose parameter mail is expected next

	// The parameters the boot-ROM load protocol carries, captured so the real DMA can be
	// performed when the "start" command arrives: the microcode's source in main memory, its
	// length in bytes, its destination in the DSP's instruction memory, and its entry point.
	UcodeSrc   uint32
	UcodeLen   uint32
	UcodeDst   uint32
	UcodeEntry uint32

	// The real DSP core, once the microcode is loaded and running. Until then the synthesized
	// boot-ROM handshake answers the mailbox; after, this core does, running the game's own
	// microcode. CoreHalt mirrors the CSR halt bit that pauses it.
	Core     *gcdsp.CPU
	CoreHalt bool

	CoreBlocked     bool // the core is waiting on an empty command mailbox; do not step it
	corePolledEmpty bool // the core just read the command mailbox and found no mail

	ARMMAddr uint32
	ARARAddr uint32
	ARCtrl   uint32
	ARSize   uint32

	// The audio-DMA (AID) engine at 0x030..0x03A. It streams 32-byte blocks of stereo PCM from
	// main memory to the DAC and, each time it drains its programmed block count, raises the AID
	// interrupt (DSP CSR bit 3) and reloads from the start/length shadow registers — a continuous
	// loop. That recurring interrupt is the audio-frame clock the AX sound driver schedules
	// against: its handler mixes and queues the next buffer, so without it the frame cadence
	// never advances. The counter advances on the instruction clock (paced like the video field),
	// so it is part of the savestate — every field here is exported for that reason.
	AIDStart     uint32 // AUDIO_DMA_START: the buffer's main-memory address (the shadow reloaded each loop)
	AIDControl   uint16 // AUDIO_DMA_CONTROL_LEN: bit 15 = enable, bits[14:0] = block count
	AIDRemaining uint16 // blocks left before the next interrupt
	AIDCur       uint32 // the address the DMA has advanced to within the current buffer
	AIDAccum     uint64 // instructions accumulated toward draining the next block

	// The DSP's own memory-DMA engine, which the running microcode drives through the
	// registers at the top of its data space (0xFFC9..0xFFCF): DSMAH/DSMAL are the main-memory
	// byte address, DSPA the DSP-memory word address, DSCR the control (direction and which DSP
	// memory), and DSBL the byte length whose write triggers the transfer. This is how the
	// mixing ucode pulls voice parameter blocks and sample data in from main RAM and writes its
	// results back out — distinct from the AR (auxiliary-RAM) DMA above.
	DSMAAddr uint32 // DSMAH:DSMAL, the main-memory byte address
	DSPAddr  uint16 // DSPA, the DSP-memory word address
	DSCtrl   uint16 // DSCR, the control word (bit 0 direction, bit 1 IMEM vs DRAM)

	// The sample accelerator's ADPCM predictor-coefficient table, sixteen words at
	// 0xFFA0..0xFFAF. The mixing ucode uploads it (DMA'd in from main memory, then block-copied
	// register by register) before it decodes any ADPCM voice; the accelerator reads it back
	// when it expands a sample. Latched here so a later accelerator model has the coefficients
	// the game actually programmed.
	Coef [16]uint16

	// The sample accelerator itself (0xFFD1..0xFFDE): the engine that serves ARAM samples to
	// the mixer, one per read of its data ports. The format word's low two bits are the sample
	// size (nibble/byte/word — and the address registers count in those units, which is why the
	// ucode shifts a byte address by (format&3)-1 before programming the current address); bits
	// 2..3 pick the decode (ADPCM or PCM, from ARAM or the input register); bits 4..5 scale the
	// PCM gain. Reading 0xFFD3 returns raw samples; reading 0xFFDD runs the ADPCM/PCM decode
	// with the coefficient table above. Writing 0xFFD3 stores a 16-bit word to ARAM, accepted
	// only when the current address carries its write flag in bit 31 — the flag the ucode sets
	// with `ori ac0.m, #0x8000` before its ARAM write loops. Semantics from Dolphin's
	// hardware-verified accelerator model, under the approved DSP exception (see gcdsp/doc.go);
	// reimplemented, not copied. All fields exported for the savestate.
	AccFormat  uint16 // 0xFFD1: sample format
	AccStart   uint32 // 0xFFD4/D5: loop start, in sample units (masked to 30 bits)
	AccEnd     uint32 // 0xFFD6/D7: end, in sample units (masked to 30 bits)
	AccCur     uint32 // 0xFFD8/D9: current position (bit 31 = raw-write enable, bit 30 masked)
	AccPred    uint16 // 0xFFDA: ADPCM predictor/scale byte (7 bits)
	AccYn1     uint16 // 0xFFDB: decode history y[n-1]
	AccYn2     uint16 // 0xFFDC: decode history y[n-2]; writing it re-arms a stopped accelerator
	AccGain    uint16 // 0xFFDE: PCM gain
	AccStopped bool   // a sample read hit the end address; reads return 0 until YN2 is rewritten
}

// The control/status register bits (the 16-bit register at offset 0x00A). RES is the
// self-clearing reset the boot sequence waits on; the INT bits are write-one-to-clear
// acknowledgements; the MASK bits gate each interrupt to the CPU.
const (
	dspCSRReset   = 1 << 0 // RES: reset the DSP; the hardware clears it when the reset completes
	dspCSRPIInt   = 1 << 1 // the mailbox interrupt to the CPU
	dspCSRHalt    = 1 << 2 // halt the DSP core
	dspCSRAIInt   = 1 << 3
	dspCSRAIMask  = 1 << 4
	dspCSRARInt   = 1 << 5
	dspCSRARMask  = 1 << 6
	dspCSRDSPInt  = 1 << 7
	dspCSRDSPMask = 1 << 8

	// The write-one-to-clear interrupt acknowledgements, together — writing any of them
	// dismisses that interrupt at the source.
	dspCSRIntAck = dspCSRPIInt | dspCSRAIInt | dspCSRARInt | dspCSRDSPInt
)

func (d *dsp) init() {
	d.ARSize = ARAMSize
	// The boot ROM's first mail: on a real console the DSP posts 0x8071FEED when it comes
	// up, and the driver waits for exactly that before it does anything else. Seeding it
	// here is what lets the wait complete.
	d.FromDSP = 0x8071FEED
	d.BootStep = 1
}

func (d *dsp) read(m *Machine, off uint32, size int) uint32 {
	if dspTrace {
		defer func() {
			r := off & 0xFFF
			if r <= 0x00A {
				fmt.Fprintf(os.Stderr, "  DSP rd 0x%03X (pc 0x%08X)  toDSP=0x%08X fromDSP=0x%08X csr=0x%04X\n",
					r, m.CPU.PC, d.ToDSP, d.FromDSP, d.CSR)
			}
		}()
	}
	switch off & 0xFFF {
	case 0x000: // mailbox to DSP, high
		return d.ToDSP >> 16
	case 0x002: // to DSP, low
		return d.ToDSP & 0xFFFF
	case 0x004: // mailbox from DSP, high — reading it does not clear; reading low does
		return d.FromDSP >> 16
	case 0x006: // from DSP, low: reading it consumes the mail
		v := d.FromDSP & 0xFFFF
		d.FromDSP &^= 1 << 31
		d.advanceBoot(m)
		return v
	case 0x00A:
		return d.CSR
	case 0x012:
		return d.ARSize
	case 0x020:
		return d.ARMMAddr >> 16
	case 0x022:
		return d.ARMMAddr & 0xFFFF
	case 0x024:
		return d.ARARAddr >> 16
	case 0x026:
		return d.ARARAddr & 0xFFFF
	case 0x028:
		return d.ARCtrl >> 16
	case 0x02A:
		return d.ARCtrl & 0xFFFF
	case 0x016:
		// The ARAM controller's mode/status register. The init sequence configures the
		// ARAM and polls bit 0 for "ready"; since the DMA engine here is instantaneous, the
		// controller is always ready, so bit 0 reads back set and the probe completes.
		return 1
	case 0x01A:
		return d.ARCtrl // AR DMA control / "is it running" — 0 means idle, i.e. done
	case 0x030:
		return d.AIDStart >> 16
	case 0x032:
		return d.AIDStart & 0xFFFF
	case 0x036:
		return uint32(d.AIDControl)
	case 0x03A:
		// AUDIO_DMA_BYTES_LEFT reads back the blocks still to play in the current buffer; a
		// driver can poll it instead of taking the interrupt.
		return uint32(d.AIDRemaining)
	}
	m.logf("DSP read unmodelled 0x%03X", off&0xFFF)
	return 0
}

func (d *dsp) write(m *Machine, off uint32, v uint32, size int) {
	if dspTrace && (off&0xFFF) <= 0x00A {
		fmt.Fprintf(os.Stderr, "  DSP wr 0x%03X = 0x%08X (pc 0x%08X)\n", off&0xFFF, v, m.CPU.PC)
	}
	switch off & 0xFFF {
	case 0x000:
		d.ToDSP = (d.ToDSP & 0xFFFF) | (v << 16)
	case 0x002:
		d.ToDSP = (d.ToDSP & 0xFFFF0000) | (v & 0xFFFF) | (1 << 31)
		// The CPU has posted a full mail (it writes the high half then the low). A running
		// core was waiting on this mail — wake it. In boot-ROM mode the synthesized handshake
		// consumes it and, where the protocol calls for a reply, posts one.
		d.CoreBlocked = false
		d.consumeMail(m)
	case 0x00A:
		// The control/status register. Three kinds of bit: the write-one-to-clear
		// interrupt acknowledgements, the self-clearing reset, and the stored mask/halt
		// bits. Compose the new value from each.
		prevHalt := d.CSR & dspCSRHalt
		ack := v & dspCSRIntAck                   // the acks: written 1s clear these
		keep := v &^ (dspCSRIntAck | dspCSRReset) // the mask and halt bits: stored as written
		// The interrupt-status bits are write-one-to-clear: a pending bit written as 0 stays
		// pending. Stripping every pending bit on any CSR write (the old composition) lost a
		// DIRQ whenever the ucode raised it between another handler's CSR read and write-back
		// — one lost DIRQ ends the driver's mail<->interrupt frame loop permanently, because
		// each side only acts on the other's event.
		d.CSR = (d.CSR &^ (dspCSRAIMask | dspCSRDSPMask | dspCSRARMask | dspCSRHalt) &^ ack) | keep
		if v&dspCSRReset != 0 {
			// The reset was requested. It completes at once — the bit reads back clear,
			// which is exactly what the boot loop is waiting for — and the boot-ROM
			// handshake restarts from the top, with any loaded ucode discarded.
			d.BootStep = 1
			d.FromDSP = 0x8071FEED
			d.UcodeRunning = false
			d.AwaitValue = false
			d.AwaitStartArg = false
			// A reset discards any running core; the boot handshake starts over and a fresh
			// microcode is loaded before the core runs again.
			d.Core = nil
			d.CoreBlocked = false
			// RES self-clears, so it is simply never set in d.CSR.
		}
		// The boot-ROM handshake is a sequence of ready mails, one per run of the DSP. The
		// ROM posts 0x8071FEED, the CPU reads it and halts the core, then unhalts it, and
		// the ROM comes back to the top and posts its ready mail again. So the falling edge
		// of the halt bit — the CPU letting the DSP run — is where the next mail appears,
		// and modelling exactly that carries the handshake past its first exchange without a
		// full DSP behind it. Once a ucode is running this no longer applies: the halt then
		// belongs to the ucode, not the boot ROM's ready loop.
		if !d.UcodeRunning && prevHalt != 0 && d.CSR&dspCSRHalt == 0 && d.FromDSP&(1<<31) == 0 {
			d.FromDSP = 0x8071FEED
		}
		m.dspRefreshIRQ()
	case 0x012:
		d.ARSize = v // the AR size / mode configuration; a game programs it and reads it back
	// The AR DMA registers. A game addresses them either as 16-bit halves or, as Luigi's
	// Mansion does, with a single 32-bit stwu — so both widths are handled, and a full-word
	// write to the low pair sets the whole register.
	case 0x020:
		if size == 4 {
			d.ARMMAddr = v
		} else {
			d.ARMMAddr = (d.ARMMAddr & 0xFFFF) | (v << 16)
		}
	case 0x022:
		d.ARMMAddr = (d.ARMMAddr & 0xFFFF0000) | (v & 0xFFFF)
	case 0x024:
		if size == 4 {
			d.ARARAddr = v
		} else {
			d.ARARAddr = (d.ARARAddr & 0xFFFF) | (v << 16)
		}
	case 0x026:
		d.ARARAddr = (d.ARARAddr & 0xFFFF0000) | (v & 0xFFFF)
	case 0x028:
		// Writing the DMA count register triggers the transfer, whether it arrives as one
		// 32-bit word or as the low half of a 16-bit pair.
		if size == 4 {
			d.ARCtrl = v
		} else {
			d.ARCtrl = (d.ARCtrl & 0xFFFF) | (v << 16)
		}
		d.runARAMDMA(m)
	case 0x02A:
		d.ARCtrl = (d.ARCtrl & 0xFFFF0000) | (v & 0xFFFF)
		d.runARAMDMA(m)
	case 0x01A:
		d.ARCtrl = v // the ARAM DMA control mirror the init sequence writes and reads back
	case 0x030:
		d.AIDStart = (d.AIDStart & 0xFFFF) | (v << 16)
	case 0x032:
		d.AIDStart = (d.AIDStart & 0xFFFF0000) | (v & 0xFFFF)
	case 0x036:
		// The control/length register: bit 15 enables the DMA, the low 15 bits are the block
		// count. Writing it (re)loads the counter from the shadow start and length — which is
		// how the interrupt handler queues the next buffer, by writing START then this.
		d.AIDControl = uint16(v)
		d.AIDRemaining = d.AIDControl & 0x7FFF
		d.AIDCur = d.AIDStart
		d.AIDAccum = 0
	default:
		m.logf("DSP write unmodelled 0x%03X = 0x%08X", off&0xFFF, v)
	}
}

// consumeMail is the synthesized boot ROM receiving a mail from the CPU. The real protocol
// is a short back-and-forth; here it is a state machine that posts the next expected reply
// and, for anything it does not recognise, logs the value so the true protocol can be read
// off a run rather than guessed at.
func (d *dsp) consumeMail(m *Machine) {
	// Once the real core runs, it takes the mail from its own side (reading the mailbox low
	// half clears the present bit). Leave the mail present and let the core see it.
	if d.Core != nil {
		return
	}

	mail := d.ToDSP &^ (1 << 31) // the write path always tags the mail present; the value is the low 31 bits
	d.ToDSP &^= 1 << 31          // the DSP has taken it

	// Once a ucode is running the driver talks to it in a command/reply rhythm: it sends a
	// command mail and polls for the ucode's answer, rejecting only a "busy" sentinel. We do
	// not run the ucode, so we synthesize the answer — a benign present mail that reads as
	// "accepted". This is a substitution in the same class as the boot-ROM handshake, not a
	// working audio DSP: it carries the audio system's init far enough to stop blocking the
	// boot, and the values are fictional. A real DSP core is the honest fix.
	if d.UcodeRunning {
		d.FromDSP = 0x80000000 | (1 << 31)
		return
	}

	// A command mail is followed by its parameter mail. When the previous mail was a load
	// command, this one is its value and carries no command of its own — except that the
	// value after the "start" command is the point at which the boot ROM would have finished
	// the DMA and jumped to the ucode. We cannot run the ucode, so we synthesize what the
	// driver would then observe: the ucode comes up and posts its first mail.
	if d.AwaitValue {
		d.AwaitValue = false
		if dspTrace {
			fmt.Fprintf(os.Stderr, "  DSP ucode-load param for cmd 0x%08X = 0x%08X\n", d.LoadCmd, mail)
		}
		// Record the parameter against the load command it belongs to. Which command carries
		// which value is read straight off the boot protocol: the source address, the byte
		// length, the DSP-memory destination and the entry point.
		switch d.LoadCmd {
		case 0x00F3A001:
			d.UcodeSrc = mail
		case 0x00F3A002:
			d.UcodeLen = mail
		case 0x00F3C002:
			d.UcodeDst = mail
		}
		if d.AwaitStartArg {
			d.AwaitStartArg = false
			d.UcodeEntry = mail
			d.startCore(m)
		}
		return
	}

	// The boot ROM's microcode-load protocol: a short run of commands (high half 0x80F3),
	// each with a parameter mail, ending in the "start" command. The parameters — the source
	// address in main memory, the length, the destination in DSP memory — are what a real
	// boot ROM would DMA; here they are recognised so the sequence is followed rather than
	// logged as a mystery, and the "start" command is what arms the ucode-came-up reply.
	switch mail {
	case 0x00F3A001, 0x00F3A002, 0x00F3C002, 0x00F3B002:
		d.AwaitValue = true // its parameter follows
		d.LoadCmd = mail
	case 0x00F3D001:
		d.AwaitValue = true // its parameter follows, and completes the load
		d.AwaitStartArg = true
		d.LoadCmd = mail
	default:
		m.logf("DSP mail from CPU: 0x%08X (boot step %d) — acknowledged; the exact protocol is a work item", mail, d.BootStep)
	}
}

// advanceBoot moves the handshake on after the CPU has read a mail.
func (d *dsp) advanceBoot(m *Machine) {
	// After the CPU reads the initial 0x8071FEED, the boot ROM waits for the driver's
	// reply, handled in consumeMail. Nothing further to volunteer here yet.
}

// post puts a mail in the DSP->CPU box and raises the mailbox interrupt if it is unmasked.
func (d *dsp) post(m *Machine, v uint32) {
	d.FromDSP = v | (1 << 31)
	if d.CSR&dspCSRDSPMask != 0 {
		if dspTrace {
			fmt.Fprintf(os.Stderr, "  DSP IRQ raise (post 0x%08X) csr=0x%04X\n", d.FromDSP, d.CSR)
		}
		d.CSR |= dspCSRDSPInt
		m.dspRefreshIRQ()
	}
}

// runARAMDMA moves a block between main memory and the auxiliary RAM. Direction is in the
// control word's top bit: one way stages samples into ARAM, the other reads them back.
func (d *dsp) runARAMDMA(m *Machine) {
	length := d.ARCtrl & 0x03FFFFE0
	if length == 0 {
		return
	}
	toARAM := d.ARCtrl&0x80000000 == 0 // 0 = main memory -> ARAM
	mm := d.ARMMAddr & 0x03FFFFE0
	ar := d.ARARAddr & 0x03FFFFFF
	for i := uint32(0); i < length; i++ {
		if int(mm+i) >= len(m.RAM) || int(ar+i) >= len(m.ARAM) {
			break
		}
		if toARAM {
			m.ARAM[ar+i] = m.RAM[mm+i]
		} else {
			m.RAM[mm+i] = m.ARAM[ar+i]
		}
	}
	d.ARCtrl = 0 // the transfer is instantaneous here; clearing the count marks it done

	// The AR-complete status bit is set whether or not the interrupt is unmasked: the
	// status is what the boot loop polls directly, and the mask only decides whether the
	// completion also reaches the CPU as an interrupt. Setting the status only when masked
	// — the earlier mistake — leaves a polling loop waiting forever.
	d.CSR |= dspCSRARInt
	if d.CSR&dspCSRARMask != 0 {
		m.dspRefreshIRQ()
	}
}

// The audio DAC plays a fixed 48 kHz; one DMA block is 32 bytes = 8 stereo 16-bit samples.
// Pacing the block drain against the same instruction clock the video field uses (one field of
// fieldInstructions is ~1/60 s) keeps the audio-frame interrupt proportional to the display, so
// a game that counts audio frames to time an intro — play the "Nintendo" voice, wait for it, then
// leave the logo — advances at the right rate relative to what is on screen.
const (
	aidSamplesPerBlock = 8
	aidInstrPerSample  = fieldInstructions * 60 / 48000 // instructions per DAC sample
	aidInstrPerBlock   = aidInstrPerSample * aidSamplesPerBlock
)

// tickAID advances the audio DMA one instruction's worth. When enabled it drains a block every
// aidInstrPerBlock instructions; when the whole buffer has drained it raises the AID interrupt
// (DSP CSR bit 3) and reloads from the shadow start/length, so the interrupt recurs on the audio
// frame period — the heartbeat the AX driver's completion callback rides to queue the next frame.
func (m *Machine) tickAID() {
	d := &m.dsp
	if d.AIDControl&0x8000 == 0 || d.AIDControl&0x7FFF == 0 {
		return // the DMA is disabled or has no blocks to play
	}
	d.AIDAccum++
	if d.AIDAccum < aidInstrPerBlock {
		return
	}
	d.AIDAccum = 0
	if d.AIDRemaining > 0 {
		d.AIDRemaining--
		if m.AIDTap != nil {
			a := phys(d.AIDCur)
			if int(a)+32 <= len(m.RAM) {
				m.AIDTap(m.RAM[a : a+32])
			}
		}
		d.AIDCur += 32
	}
	if d.AIDRemaining == 0 {
		// The buffer finished. Reload it and raise the AID interrupt, which the sound driver's
		// callback services — mixing and queueing the next buffer. The status bit is set whether
		// or not the interrupt is unmasked (a driver may poll AUDIO_DMA_BYTES_LEFT instead); the
		// mask only decides whether it also reaches the CPU, gated in dspRefreshIRQ.
		d.AIDCur = d.AIDStart
		d.AIDRemaining = d.AIDControl & 0x7FFF
		d.CSR |= dspCSRAIInt
		m.dspRefreshIRQ()
	}
}

func (m *Machine) dspRefreshIRQ() {
	c := m.dsp.CSR
	pending := (c&dspCSRDSPInt != 0 && c&dspCSRDSPMask != 0) ||
		(c&dspCSRARInt != 0 && c&dspCSRARMask != 0) ||
		(c&dspCSRAIInt != 0 && c&dspCSRAIMask != 0)
	if pending {
		m.raiseInt(IntDSP)
	} else {
		m.clearInt(IntDSP)
	}
}

// startCore loads the microcode the boot protocol described into a fresh DSP core and starts it
// at its entry point. From here the real core, not the synthesized handshake, answers the
// mailbox — running the game's own microcode, which is what the audio system is waiting for.
func (d *dsp) startCore(m *Machine) {
	src := d.UcodeSrc & 0x03FFFFFF
	core := gcdsp.New(dspBus{m})
	// DMA the microcode from main memory into instruction RAM as big-endian 16-bit words.
	dst := uint16(d.UcodeDst)
	for i := uint32(0); i+1 < d.UcodeLen; i += 2 {
		if int(src+i)+1 >= len(m.RAM) {
			break
		}
		w := uint16(m.RAM[src+i])<<8 | uint16(m.RAM[src+i+1])
		if wi := dst + uint16(i/2); int(wi) < len(core.IRAM) {
			core.IRAM[wi] = w
		}
	}
	core.PC = uint16(d.UcodeEntry)
	d.Core = core
	d.UcodeRunning = true
	m.logf("DSP: real core started — %d-word ucode from 0x%08X, entry 0x%04X", d.UcodeLen/2, src, d.UcodeEntry)
}

// dspBus lets the DSP core reach the console hardware it addresses through the top of its data
// space: the mailboxes to and from the CPU, and — as they are implemented — the DMA engine and
// the sample accelerator.
type dspBus struct{ m *Machine }

func (b dspBus) HWRead(a uint16) uint16     { return b.m.dsp.hwRead(b.m, a) }
func (b dspBus) HWWrite(a uint16, v uint16) { b.m.dsp.hwWrite(b.m, a, v) }

// hwRead answers a DSP-side hardware read. The mailboxes are the shared state the CPU-facing
// registers also touch; anything else halts the core, naming the register, so the protocol is
// discovered from a run rather than guessed.
func (d *dsp) hwRead(m *Machine, a uint16) uint16 {
	switch a {
	case 0xFFFC: // DMBH: the mail the DSP queued to the CPU, read back — bit 15 is "present",
		// which the CPU clears by reading; the ucode polls this waiting for the CPU to consume.
		return uint16(d.FromDSP >> 16)
	case 0xFFFD: // DMBL
		return uint16(d.FromDSP)
	case 0xFFFE: // CMBH: mail from the CPU, high half — bit 15 is "mail present"
		if d.ToDSP&(1<<31) == 0 {
			d.corePolledEmpty = true // an empty poll: the core is waiting for a command
		}
		return uint16(d.ToDSP >> 16)
	case 0xFFFF: // CMBL: reading the low half consumes the mail
		v := uint16(d.ToDSP)
		d.ToDSP &^= 1 << 31
		return v
	case 0xFFC9: // DSCR: the memory-DMA control. Bit 2 is "DMA in progress"; the ucode polls it
		// waiting for a transfer to finish. This engine transfers instantaneously, so the busy
		// bit is never set and the poll exits at once — the stored control word reads back with
		// bit 2 clear.
		return d.DSCtrl
	}
	if a >= 0xFFA0 && a <= 0xFFAF { // the ADPCM coefficient table, read back
		return d.Coef[a-0xFFA0]
	}
	switch a { // the sample accelerator
	case 0xFFD1:
		return d.AccFormat
	case 0xFFD3: // the raw data port: the next sample, undecoded
		return d.accReadRaw(m)
	case 0xFFD4:
		return uint16(d.AccStart >> 16)
	case 0xFFD5:
		return uint16(d.AccStart)
	case 0xFFD6:
		return uint16(d.AccEnd >> 16)
	case 0xFFD7:
		return uint16(d.AccEnd)
	case 0xFFD8:
		return uint16(d.AccCur >> 16)
	case 0xFFD9:
		return uint16(d.AccCur)
	case 0xFFDA:
		return d.AccPred
	case 0xFFDB:
		return d.AccYn1
	case 0xFFDC:
		return d.AccYn2
	case 0xFFDD: // the decoding data port: the next sample, through the ADPCM/PCM decoder
		return d.accReadSample(m)
	case 0xFFDE:
		return d.AccGain
	}
	d.Core.Halt("DSP read of unmodelled hardware register 0x%04X at ucode 0x%04X", a, d.Core.PC)
	return 0
}

// hwWrite answers a DSP-side hardware write. Writing the outgoing mailbox (DMBH then DMBL)
// posts a mail to the CPU, whose present bit rides in DMBH's high bit and which the CPU reads
// or polls. Posting a mailbox does NOT interrupt the CPU on this hardware: the interrupt is a
// separate signal the microcode raises explicitly by writing DIRQ. Conflating the two — raising
// the interrupt on the mailbox write — storms the CPU with an interrupt it has no handler for
// during the boot handshake, when the driver means to consume the ready mail by polling.
func (d *dsp) hwWrite(m *Machine, a uint16, v uint16) {
	switch a {
	case 0xFFFB: // DIRQ: raise the DSP -> CPU interrupt. Bit 0 asserts it; the CPU takes it only
		// if it has unmasked the DSP interrupt (dspRefreshIRQ gates on the mask). The microcode
		// writes this after it has serviced a command — never during the ready-mail handshake,
		// which is why the driver's synchronous poll of that mail must not be preempted here.
		if v&1 != 0 {
			if dspTrace {
				fmt.Fprintf(os.Stderr, "  DSP IRQ raise (DIRQ, ucode pc 0x%04X) csr=0x%04X\n", d.Core.PC, d.CSR)
			}
			d.CSR |= dspCSRDSPInt
			m.dspRefreshIRQ()
		}
		return
	case 0xFFFC: // DMBH: queue the high half of a mail to the CPU (its bit 15 is "present")
		d.FromDSP = (d.FromDSP & 0xFFFF) | (uint32(v) << 16)
		return
	case 0xFFFD: // DMBL: complete the mail. This makes it present for the CPU to read or poll;
		// it does not interrupt the CPU (see DIRQ above).
		d.FromDSP = (d.FromDSP &^ 0xFFFF) | uint32(v)
		if dspTrace {
			fmt.Fprintf(os.Stderr, "  DSP mail out 0x%08X (ucode pc 0x%04X)\n", d.FromDSP, d.Core.PC)
		}
		return

	// The memory-DMA registers. The address and control registers are latched; writing the
	// block-length register DSBL is what starts the transfer.
	case 0xFFC9: // DSCR: control (direction in bit 0, IMEM vs DRAM in bit 1)
		d.DSCtrl = v
		return
	case 0xFFCD: // DSPA: DSP-memory word address
		d.DSPAddr = v
		return
	case 0xFFCE: // DSMAH: main-memory address, high half
		d.DSMAAddr = (d.DSMAAddr & 0x0000FFFF) | (uint32(v) << 16)
		return
	case 0xFFCF: // DSMAL: main-memory address, low half
		d.DSMAAddr = (d.DSMAAddr & 0xFFFF0000) | uint32(v)
		return
	case 0xFFCB: // DSBL: block length in bytes — writing it triggers the transfer
		d.runMemDMA(m, v)
		return
	}
	if a >= 0xFFA0 && a <= 0xFFAF { // the ADPCM coefficient table, uploaded before decoding
		d.Coef[a-0xFFA0] = v
		return
	}
	switch a { // the sample accelerator. Start and end mask to 30 bits; the current address
	// keeps bit 31, the raw-write enable.
	case 0xFFD1:
		d.AccFormat = v
		return
	case 0xFFD3:
		d.accWriteRaw(m, v)
		return
	case 0xFFD4:
		d.AccStart = (d.AccStart&0xFFFF | uint32(v)<<16) & 0x3FFFFFFF
		return
	case 0xFFD5:
		d.AccStart = (d.AccStart&0xFFFF0000 | uint32(v)) & 0x3FFFFFFF
		return
	case 0xFFD6:
		d.AccEnd = (d.AccEnd&0xFFFF | uint32(v)<<16) & 0x3FFFFFFF
		return
	case 0xFFD7:
		d.AccEnd = (d.AccEnd&0xFFFF0000 | uint32(v)) & 0x3FFFFFFF
		return
	case 0xFFD8:
		d.AccCur = (d.AccCur&0xFFFF | uint32(v)<<16) & 0xBFFFFFFF
		return
	case 0xFFD9:
		d.AccCur = (d.AccCur&0xFFFF0000 | uint32(v)) & 0xBFFFFFFF
		return
	case 0xFFDA:
		d.AccPred = v & 0x7F
		return
	case 0xFFDB:
		d.AccYn1 = v
		return
	case 0xFFDC:
		d.AccYn2 = v
		d.AccStopped = false // rewriting the history re-arms a stopped accelerator
		return
	case 0xFFDE:
		d.AccGain = v
		return
	}
	d.Core.Halt("DSP write of unmodelled hardware register 0x%04X = 0x%04X at ucode 0x%04X", a, v, d.Core.PC)
}

// aramByte reads one byte of auxiliary RAM for the accelerator, bounds-checked the same way the
// AR DMA is — a run past the end reads zero rather than crashing.
func (d *dsp) aramByte(m *Machine, addr uint32) uint8 {
	if int(addr) >= len(m.ARAM) {
		return 0
	}
	return m.ARAM[addr]
}

// accCurrentSample fetches the sample the accelerator's current address points at, in the width
// the format's low two bits select — and those bits also fix the unit the address registers
// count in: nibbles, bytes or 16-bit words. A format with both bits set is not a real width;
// it halts rather than serve garbage.
func (d *dsp) accCurrentSample(m *Machine) uint16 {
	switch d.AccFormat & 3 {
	case 0: // 4-bit: the address counts nibbles, high nibble of each byte first
		v := d.aramByte(m, d.AccCur>>1)
		if d.AccCur&1 != 0 {
			return uint16(v & 0xF)
		}
		return uint16(v >> 4)
	case 1: // 8-bit: the address counts bytes
		return uint16(d.aramByte(m, d.AccCur))
	case 2: // 16-bit: the address counts big-endian words
		return uint16(d.aramByte(m, d.AccCur*2))<<8 | uint16(d.aramByte(m, d.AccCur*2+1))
	default:
		d.Core.Halt("DSP accelerator: invalid sample width in format 0x%04X", d.AccFormat)
		return 0
	}
}

// accReadRaw serves one undecoded sample from the raw data port (0xFFD3) and advances the
// current address. Reading the sample at the end address wraps the accelerator back to the
// start — the microcode's ARAM block reads size their loops to stay inside, so the wrap is the
// contract, not a corner. This is the port the mixing ucode moves whole buffers through.
func (d *dsp) accReadRaw(m *Machine) uint16 {
	v := d.accCurrentSample(m)
	if d.Core.Halted {
		return 0
	}
	d.AccCur++
	if d.AccCur-1 == d.AccEnd {
		// Wrap to the start. The overflow exception this raises on hardware lands on a bare
		// rti in this game's ucode, so it is not modelled.
		d.AccCur = d.AccStart
	}
	d.AccCur &= 0xBFFFFFFF
	return v
}

// accWriteRaw stores one 16-bit word to ARAM through the raw data port. The hardware accepts
// the write only when the current address carries its write flag in bit 31 — the ucode sets it
// with `ori ac0.m, #0x8000` before a write loop — and the address is then treated as 16-bit
// units regardless of the format width. Multiplying the flagged address by two drops the flag
// out of the top bit and lands on the byte address, exactly as the hardware's adder does.
func (d *dsp) accWriteRaw(m *Machine, v uint16) {
	if d.AccCur&0x80000000 == 0 {
		m.logf("DSP accelerator: raw write without the address write flag (cur=0x%08X)", d.AccCur)
		return
	}
	byteAddr := d.AccCur * 2
	if int(byteAddr)+1 < len(m.ARAM) {
		m.ARAM[byteAddr] = byte(v >> 8)
		m.ARAM[byteAddr+1] = byte(v)
	}
	d.AccCur++
}

// accReadSample serves one decoded sample from the decoding data port (0xFFDD): ADPCM expansion
// or PCM scaling, both filtered through the predictor history. Reaching the end address stops
// the accelerator — reads return zero until the driver rewrites YN2 — which is how a one-shot
// voice goes quiet instead of looping.
func (d *dsp) accReadSample(m *Machine) uint16 {
	if d.AccStopped {
		return 0
	}
	decode := (d.AccFormat >> 2) & 3
	if decode == 1 || decode == 3 {
		// The MMIO-input modes decode a sample the CPU wrote to the input register rather than
		// one fetched from ARAM; nothing has written that register yet, so the path waits for
		// the first user rather than inventing one.
		d.Core.Halt("DSP accelerator: MMIO-input PCM decode (format 0x%04X) not yet implemented", d.AccFormat)
		return 0
	}
	raw := int32(d.accCurrentSample(m))
	if d.Core.Halted {
		return 0
	}
	coefIdx := (d.AccPred >> 4) & 7
	c1 := int32(int16(d.Coef[coefIdx*2]))
	c2 := int32(int16(d.Coef[coefIdx*2+1]))
	yn1 := int32(int16(d.AccYn1))
	yn2 := int32(int16(d.AccYn2))

	var val uint16
	step := uint32(2)
	if decode == 0 { // ADPCM: 4-bit deltas against the predictor pair, scaled by the frame header
		raw &= 0xF
		if raw >= 8 {
			raw -= 16
		}
		scale := int32(1) << (d.AccPred & 0xF)
		v32 := scale*raw + ((0x400 + c1*yn1 + c2*yn2) >> 11)
		if v32 > 0x7FFF {
			v32 = 0x7FFF
		} else if v32 < -0x7FFF {
			v32 = -0x7FFF
		}
		val = uint16(int16(v32))
		d.AccYn2, d.AccYn1 = d.AccYn1, val
		d.AccCur++
		switch {
		// A frame-aligned end is handled apart from the plain wrap: the predictor byte is not
		// reloaded and the restart lands just past the header nibble.
		case d.AccEnd&0xF == 0x0 && d.AccCur == d.AccEnd:
			d.AccCur = d.AccStart + 1
		case d.AccEnd&0xF == 0x1 && d.AccCur == d.AccEnd-1:
			d.AccCur = d.AccStart
		case d.AccCur&15 == 0:
			// Every 16 nibbles the next frame's predictor/scale byte is consumed in-line.
			d.AccPred = uint16(d.aramByte(m, (d.AccCur&^15)>>1)) & 0x7F
			d.AccCur += 2
			step += 2
		}
	} else { // PCM: the sample scaled by the gain, plus the filtered history
		var gainShift uint
		switch (d.AccFormat >> 4) & 3 {
		case 0:
			gainShift = 11 // gain counts in 1/2048ths
		case 1:
			gainShift = 0
		case 2:
			gainShift = 16
		default:
			d.Core.Halt("DSP accelerator: invalid gain scale in format 0x%04X", d.AccFormat)
			return 0
		}
		v32 := (int32(int16(d.AccGain))*int32(int16(raw)))>>gainShift +
			(c1*yn1)>>gainShift + (c2*yn2)>>gainShift
		val = uint16(int16(v32))
		d.AccYn2, d.AccYn1 = d.AccYn1, val
		d.AccCur++
	}

	if d.AccCur == d.AccEnd+step-1 {
		// The voice ran off its end: back to the start and stop serving samples until the
		// driver rewrites the history registers. The overflow exception lands on a bare rti in
		// this ucode, so only the stop is modelled.
		d.AccCur = d.AccStart
		d.AccStopped = true
	}
	d.AccCur &= 0xBFFFFFFF
	return val
}

// runMemDMA performs the DSP's memory DMA: it moves lenBytes between main memory (a big-endian
// byte address in DSMAH:DSMAL) and the DSP's own memory (a word address in DSPA), the direction
// and target chosen by the control word DSCR. Bit 0 of DSCR is the direction — 0 pulls data from
// main memory into the DSP, 1 pushes it back out — and bit 1 selects instruction memory over
// data memory, which the mixing ucode does not use, so that case halts loudly rather than moving
// data into the wrong space. The transfer is instantaneous, so no busy bit is ever raised.
func (d *dsp) runMemDMA(m *Machine, lenBytes uint16) {
	if lenBytes == 0 {
		return
	}
	if d.DSCtrl&0x2 != 0 {
		d.Core.Halt("DSP memory DMA into instruction memory (DSCR=0x%04X) not modelled", d.DSCtrl)
		return
	}
	toDSP := d.DSCtrl&0x1 == 0 // bit 0 clear = main memory -> DSP; set = DSP -> main memory
	main := d.DSMAAddr & 0x03FFFFFF
	words := uint32(lenBytes) / 2
	for i := uint32(0); i < words; i++ {
		mb := main + i*2
		dw := d.DSPAddr + uint16(i)
		if int(mb)+1 >= len(m.RAM) || int(dw) >= len(d.Core.DRAM) {
			break
		}
		if toDSP {
			d.Core.DRAM[dw] = uint16(m.RAM[mb])<<8 | uint16(m.RAM[mb+1])
		} else {
			w := d.Core.DRAM[dw]
			m.RAM[mb] = byte(w >> 8)
			m.RAM[mb+1] = byte(w)
		}
	}
}

// tickDSP steps the running DSP core. It is called from the machine's main loop alongside the
// video tick. The core runs in short bursts and stops the moment it polls an empty command
// mailbox — it is then waiting on the CPU, and stays parked until the CPU posts the next mail,
// which keeps an idle DSP from spinning through the whole run. A core halt (an unmodelled op or
// register) stops the whole machine, surfacing the reason.
func (m *Machine) tickDSP() {
	d := &m.dsp
	if d.Core == nil || d.CoreHalt || d.CoreBlocked || d.Core.Halted {
		return
	}
	for i := 0; i < 64; i++ {
		d.corePolledEmpty = false
		if dspPCTrace {
			fmt.Fprintf(os.Stderr, "  ucode pc 0x%04X\n", d.Core.PC)
		}
		if !d.Core.Step() {
			m.CPU.Halt("DSP core halted: %s", d.Core.Reason)
			return
		}
		if d.corePolledEmpty {
			d.CoreBlocked = true
			return
		}
	}
}

var dspPCTrace = os.Getenv("RR_GC_DSPPC") != ""

var dspTrace = os.Getenv("RR_GC_DSPTRACE") != ""
