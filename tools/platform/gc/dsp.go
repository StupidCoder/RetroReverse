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
	AIAddr   uint32
	AILen    uint32
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
		return d.AIAddr >> 16
	case 0x032:
		return d.AIAddr & 0xFFFF
	case 0x036:
		return d.AILen
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
		d.CSR = (d.CSR &^ dspCSRIntAck &^ (dspCSRAIMask | dspCSRDSPMask | dspCSRARMask | dspCSRHalt)) | keep
		d.CSR &^= ack
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
		d.AIAddr = (d.AIAddr & 0xFFFF) | (v << 16)
	case 0x032:
		d.AIAddr = (d.AIAddr & 0xFFFF0000) | (v & 0xFFFF)
	case 0x036:
		d.AILen = v
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
		return
	}
	d.Core.Halt("DSP write of unmodelled hardware register 0x%04X = 0x%04X at ucode 0x%04X", a, v, d.Core.PC)
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

var dspTrace = os.Getenv("RR_GC_DSPTRACE") != ""
