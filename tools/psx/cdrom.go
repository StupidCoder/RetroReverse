package psx

import "fmt"

// cdrom.go emulates the PlayStation CD-ROM controller at 0x1F801800-0x1F801803.
// Ridge Racer drives it directly (its libcd keeps the four register addresses in
// a pointer table shipped in the EXE at 0x8007945C, so they never appear as
// lui-immediate 0x1F80xxxx accesses). The four ports are byte-wide and banked by
// a 2-bit index written to port 0:
//
//	port 0 : write = index select;         read = status register
//	port 1 : write = command (index 0);    read = response FIFO
//	port 2 : write = parameter FIFO (i0);  read = data FIFO
//	         write = IRQ enable (index 1)
//	port 3 : write = request register (i0) / IRQ flag ack (index 1)
//	         read = IRQ enable (i0/2) / IRQ flag (i1/3)
//
// A command produces one or two responses, each delivered as a CD interrupt
// (I_STAT bit 2) carrying an INT cause in the low 3 bits of the IRQ-flag
// register: INT3 acknowledge, INT2 second-response complete, INT1 data-ready,
// INT5 error. The game reads the flag, drains the response FIFO, and acks by
// writing the flag back; only then is the next queued response delivered. Sector
// data is served from the mounted disc image (cd.go).
//
// Timing is approximate: responses are queued with a coarse step delay so the
// game's poll/critical-section sequencing sees an interrupt arrive "later", not
// in the same instruction as the command write.

// CD status-register bits (port 0 read).
const (
	cdStatPRMEMPT = 1 << 3 // parameter FIFO empty
	cdStatPRMWRDY = 1 << 4 // parameter FIFO not full (ready for writes)
	cdStatRSLRRDY = 1 << 5 // response FIFO not empty
	cdStatDRQSTS  = 1 << 6 // data FIFO not empty
	cdStatBUSYSTS = 1 << 7 // busy transmitting a command
)

// CD drive status byte (returned inside responses).
const (
	cdMotorOn = 1 << 1 // spun up / ready
	cdReading = 1 << 5 // reading data sectors
	cdSeeking = 1 << 6 // seeking
)

// pending is one queued CD response (an interrupt with its cause and data).
type pending struct {
	delay int    // steps until delivery
	cause byte   // INT cause (1,2,3,5)
	resp  []byte // response FIFO contents
	read  bool   // this INT1 also loads the next data sector
}

type cdrom struct {
	m *Machine

	index    byte   // port-0 low bits: register bank
	stat     byte   // drive status byte
	mode     byte   // Setmode value
	params   []byte // parameter FIFO (host -> controller)
	response []byte // response FIFO (controller -> host)
	data     []byte // data FIFO (sector bytes)
	dataPos  int
	irqEnable byte // IRQ mask (port 3, index 0)
	irqFlags  byte // pending INT cause (port 3, index 1); nonzero = unacked

	loc     int  // Setloc target LBA
	readLBA int  // next sector to deliver while reading
	reading bool // a ReadN/ReadS is streaming sectors

	queue []pending // responses awaiting delivery

	cmdLog []string // ordered log of issued commands (for tracing)
	trace  []string // capped log of every port access (for tracing)
}

func (c *cdrom) logf(dir string, port uint32, v byte) {
	if len(c.trace) < 400 {
		c.trace = append(c.trace, fmt.Sprintf("%s port%d idx%d val=0x%02X flags=0x%02X qlen=%d",
			dir, port, c.index, v, c.irqFlags, len(c.queue)))
	}
}

func newCDROM(m *Machine) *cdrom {
	return &cdrom{m: m, stat: cdMotorOn}
}

// CDCommands returns the ordered list of CD commands the game has issued, for a
// run summary / trace.
func (m *Machine) CDCommands() []string { return m.cd.cmdLog }

// CDTrace returns the capped per-access CD port trace.
func (m *Machine) CDTrace() []string { return m.cd.trace }

// --- register interface ----------------------------------------------------

func (c *cdrom) read(port uint32) byte {
	v := c.read0(port)
	c.logf("RD", port, v)
	return v
}

func (c *cdrom) read0(port uint32) byte {
	switch port {
	case 0: // status register
		s := c.index
		if len(c.params) == 0 {
			s |= cdStatPRMEMPT
		}
		s |= cdStatPRMWRDY
		if len(c.response) > 0 {
			s |= cdStatRSLRRDY
		}
		if c.dataPos < len(c.data) {
			s |= cdStatDRQSTS
		}
		return s
	case 1: // response FIFO
		if len(c.response) == 0 {
			return 0
		}
		b := c.response[0]
		c.response = c.response[1:]
		return b
	case 2: // data FIFO
		if c.dataPos < len(c.data) {
			b := c.data[c.dataPos]
			c.dataPos++
			return b
		}
		return 0
	case 3:
		if c.index&1 == 0 {
			return c.irqEnable | 0xE0
		}
		return c.irqFlags | 0xE0
	}
	return 0
}

func (c *cdrom) write(port uint32, v byte) {
	c.logf("WR", port, v)
	switch port {
	case 0:
		c.index = v & 3
	case 1:
		if c.index == 0 {
			c.command(v)
		}
	case 2:
		switch c.index {
		case 0:
			c.params = append(c.params, v)
		case 1:
			c.irqEnable = v & 0x1F
		}
	case 3:
		switch c.index {
		case 0:
			// Request register: bit 7 (BFRD) loads the sector into the data FIFO.
			if v&0x80 != 0 {
				c.loadData()
			} else {
				c.data, c.dataPos = nil, 0
			}
		case 1:
			// IRQ-flag ack: clear the flagged bits; bit 6 resets the parameter FIFO.
			c.irqFlags &^= v & 0x1F
			if v&0x40 != 0 {
				c.params = nil
			}
			if c.irqFlags == 0 {
				c.deliverNext()
			}
		}
	}
}

// --- command dispatch ------------------------------------------------------

func (c *cdrom) command(cmd byte) {
	name := cdCmdName(cmd)
	c.cmdLog = append(c.cmdLog, name)
	c.m.note("CD cmd " + name)

	switch cmd {
	case 0x01: // CdlNop / GetStat
		c.ack()
	case 0x02: // CdlSetloc: params = min, sec, frame (BCD)
		if len(c.params) >= 3 {
			c.loc = bcdToLBA(c.params[0], c.params[1], c.params[2])
			c.cmdLog[len(c.cmdLog)-1] = fmt.Sprintf("Setloc(%02X:%02X:%02X=LBA%d)",
				c.params[0], c.params[1], c.params[2], c.loc)
		}
		c.ack()
	case 0x06, 0x1B: // CdlReadN / CdlReadS
		c.stat |= cdReading
		c.reading = true
		c.readLBA = c.loc
		c.ack()
		c.queueRead()
	case 0x09: // CdlPause
		c.stat &^= cdReading
		c.reading = false
		c.ack()
		c.queueSecond()
	case 0x0A: // CdlInit
		c.stat = cdMotorOn
		c.mode = 0
		c.reading = false
		c.ack()
		c.queueSecond()
	case 0x0B, 0x0C: // CdlMute / CdlDemute
		c.ack()
	case 0x0E: // CdlSetmode
		if len(c.params) >= 1 {
			c.mode = c.params[0]
		}
		c.ack()
	case 0x15, 0x16: // CdlSeekL / CdlSeekP
		c.stat |= cdSeeking
		c.ack()
		c.stat &^= cdSeeking
		c.queueSecond()
	case 0x13: // CdlGetTN: first & last track numbers (BCD). Data-only disc: 1..1.
		c.enqueue(pending{delay: cdAckDelay, cause: 3, resp: []byte{c.stat, 0x01, 0x01}})
	case 0x14: // CdlGetTD: start MSF of a track (param = track, BCD; 0 = lead-out)
		track := 0
		if len(c.params) >= 1 {
			track = int(fromBCD(c.params[0]))
		}
		lba := 0 // track 1 (the data track) starts at LBA 0
		if track == 0 && c.m.disc != nil {
			lba = c.m.disc.nsect // lead-out = end of disc
		}
		mm, ss := lbaToMSF(lba)
		c.enqueue(pending{delay: cdAckDelay, cause: 3, resp: []byte{c.stat, toBCD(mm), toBCD(ss)}})
	case 0x19: // CdlTest
		c.testCommand()
	default:
		// Unknown: acknowledge with status so the game can proceed / log it.
		c.ack()
	}
	c.params = nil
}

// ack queues the immediate INT3 acknowledge carrying the status byte.
func (c *cdrom) ack() { c.enqueue(pending{delay: cdAckDelay, cause: 3, resp: []byte{c.stat}}) }

// queueSecond queues the INT2 completion for two-phase commands (init/seek/pause).
func (c *cdrom) queueSecond() {
	c.enqueue(pending{delay: cdSecondDelay, cause: 2, resp: []byte{c.stat}})
}

// queueRead queues the INT1 data-ready response for the next streamed sector.
func (c *cdrom) queueRead() {
	c.enqueue(pending{delay: cdReadDelay, cause: 1, resp: []byte{c.stat}, read: true})
}

// testCommand handles CdlTest sub-functions (only the BIOS-version query matters).
func (c *cdrom) testCommand() {
	if len(c.params) >= 1 && c.params[0] == 0x20 {
		// Return a plausible controller date/version (yy,mm,dd,ver).
		c.enqueue(pending{delay: cdAckDelay, cause: 3, resp: []byte{0x94, 0x09, 0x19, 0xC0}})
		return
	}
	c.ack()
}

// --- response delivery -----------------------------------------------------

const (
	cdAckDelay    = 1000  // steps to the INT3 acknowledge
	cdSecondDelay = 20000 // steps to the INT2 completion
	cdReadDelay   = 15000 // steps between streamed INT1 sectors
)

func (c *cdrom) enqueue(p pending) { c.queue = append(c.queue, p) }

// tick advances queued responses; called once per CPU step from the run loop.
func (c *cdrom) tick() {
	if len(c.queue) == 0 || c.irqFlags != 0 {
		return // nothing pending, or the last interrupt is still unacked
	}
	if c.queue[0].delay > 0 {
		c.queue[0].delay--
		return
	}
	c.deliverNext()
}

// deliverNext hands the head-of-queue response to the game and raises the CD IRQ.
func (c *cdrom) deliverNext() {
	if len(c.queue) == 0 || c.irqFlags != 0 {
		return
	}
	p := c.queue[0]
	if p.delay > 0 {
		return
	}
	c.queue = c.queue[1:]
	c.response = append([]byte(nil), p.resp...)
	c.irqFlags = p.cause
	if p.read && c.reading {
		// INT1 only signals "a sector is ready"; the sector is pulled on demand by
		// the DMA (or a BFRD request) so readLBA advances once per sector actually
		// consumed. Keep streaming while the read is active.
		c.queueRead()
	}
	c.m.raiseIRQ(2) // CD interrupt line; the game gates delivery via I_STAT/I_MASK
}

// dmaTo services a CDROM DMA (channel 3): it streams `bcr` words from the data
// FIFO into RAM at `madr`, pulling further sectors off the disc as the FIFO
// drains while a read is active. bcr is blocksize(words) | blockcount<<16.
func (c *cdrom) dmaTo(madr, bcr uint32) {
	words := bcr & 0xFFFF
	if bc := bcr >> 16; bc > 1 {
		words *= bc
	}
	for i := uint32(0); i < words; i++ {
		if c.dataPos+4 > len(c.data) {
			if !c.reading {
				break
			}
			c.loadData() // refill from the next sector
		}
		if c.dataPos+4 > len(c.data) {
			break
		}
		w := uint32(c.data[c.dataPos]) | uint32(c.data[c.dataPos+1])<<8 |
			uint32(c.data[c.dataPos+2])<<16 | uint32(c.data[c.dataPos+3])<<24
		c.dataPos += 4
		c.m.write32(madr, w)
		madr += 4
	}
}

// loadData copies the current sector (2048 user bytes) into the data FIFO and
// advances the read pointer.
func (c *cdrom) loadData() {
	if c.m.disc == nil {
		return
	}
	blk, err := c.m.disc.block(c.readLBA)
	if err != nil {
		return
	}
	c.data = append([]byte(nil), blk...)
	c.dataPos = 0
	c.readLBA++
}

// --- helpers ---------------------------------------------------------------

// bcdToLBA converts a BCD min:sec:frame address to a logical block number,
// removing the 2-second (150-sector) pregap.
func bcdToLBA(m, s, f byte) int {
	mm := int(m>>4)*10 + int(m&0xF)
	ss := int(s>>4)*10 + int(s&0xF)
	ff := int(f>>4)*10 + int(f&0xF)
	lba := (mm*60+ss)*75 + ff - 150
	if lba < 0 {
		lba = 0
	}
	return lba
}

func fromBCD(b byte) int  { return int(b>>4)*10 + int(b&0x0F) }
func toBCD(n int) byte     { return byte((n/10)<<4 | n%10) }

// lbaToMSF converts a logical block number to minutes/seconds on the disc,
// including the 2-second (150-sector) pregap.
func lbaToMSF(lba int) (mm, ss int) {
	t := lba + 150
	return t / (75 * 60), (t / 75) % 60
}

func cdCmdName(cmd byte) string {
	switch cmd {
	case 0x01:
		return "GetStat"
	case 0x02:
		return "Setloc"
	case 0x06:
		return "ReadN"
	case 0x09:
		return "Pause"
	case 0x0A:
		return "Init"
	case 0x0B:
		return "Mute"
	case 0x0C:
		return "Demute"
	case 0x0E:
		return "Setmode"
	case 0x15:
		return "SeekL"
	case 0x16:
		return "SeekP"
	case 0x19:
		return "Test"
	case 0x1B:
		return "ReadS"
	case 0x0D:
		return "Setfilter"
	case 0x0F:
		return "Getparam"
	case 0x10:
		return "GetlocL"
	case 0x11:
		return "GetlocP"
	case 0x13:
		return "GetTN"
	case 0x14:
		return "GetTD"
	case 0x1A:
		return "GetID"
	case 0x03:
		return "Play"
	case 0x08:
		return "Stop"
	}
	return fmt.Sprintf("cmd?0x%02X", cmd)
}
