package ps2

// iopcdvd.go is the CD/DVD controller — the drive itself, as CDVDMAN drives it.
//
// It is the last piece of silicon standing between the game and its own data. The EE
// prints "Initializing CD drive", binds CDVDFSV across the SIF, asks it for a file, and
// CDVDMAN then goes round a loop reading two registers, 91,697 times apiece, waiting for
// a drive that was not there. Everything after this — KERNEL.CGO, the GOAL linker, the
// first frame — comes off the disc, so nothing after this happens until the drive does.
//
// The register map below was not looked up. It was read out of CDVDMAN, which is on the
// disc, and every line of it can be pointed at the instruction that proves it. The module
// is stripped, so what follows names the code rather than the function:
//
//	0x1F402004  W  the N-command code. Writing it starts the command.
//	               CDVDMAN+0x2B60: `sb $s1, 0($v0)` with $v0 = 0xBF402004, $s1 the command,
//	               and it is the last store of the submit routine — everything else is set
//	               up first.
//	0x1F402005  W  a parameter byte, pushed into a FIFO. CDVDMAN+0x2B38: a loop storing
//	               $s5 bytes here, one per iteration, immediately before the command goes
//	               into 0x2004.
//	            R  the drive's status. The submit routine reads it and refuses to proceed
//	               unless `& 0xC0 == 0x40` (CDVDMAN+0x28AC), and the interrupt handler reads
//	               bit 0 as the command's error flag (CDVDMAN+0x15C8: bit set -> -1, clear
//	               -> 1, and the result is what the waiting thread collects).
//	0x1F402006  W  the transfer mode, written by the DMA-start routine (CDVDMAN+0x2814)
//	               out of the descriptor's byte 8 — 0x80 for a CD, 0x84 for a DVD, 0x8C for a
//	               DVD sector read.
//	            R  the ERROR the last command ended with, and it reads back as nothing like
//	               what is written to it. The interrupt handler's first act is to read it
//	               (CDVDMAN+0x1598) and file the byte at state+1; "what was the last error"
//	               hands back exactly that byte (CDVDMAN+0x2DBC); and a sector read fails
//	               unless it is zero (CDVDMAN+0x6D4C). Zero means the command worked.
//	0x1F402007  W  an abort. CDVDMAN writes 1 here on the path that also records error 8.
//	0x1F402008  R  the interrupt cause. W: write-1-to-clear, and the handler clears exactly
//	               the bit it served.
//	0x1F40200A  R  the drive's state. The submit routine will not issue an N-command until
//	               this reads 10 (CDVDMAN+0x2A14), waiting a millisecond at a time until it
//	               does. It is the only value the code ever compares against, so "10" means
//	               ready and every other value means not yet.
//	0x1F40200F  R  the disc type. CDVDMAN+0x3048 branches on it to choose the read command:
//	               16..19 -> 0x80, 20 -> 0x84, below 16 -> refuse. This disc is 1.74 GB, so
//	               it is not a CD by any reading of the geometry, and 20 is the only value
//	               that selects the other branch.
//	0x1F402016  W  the S-command code. Writing it runs the command (CDVDMAN+0x22C8).
//	0x1F402017  W  a parameter byte for the S-command, pushed into a FIFO (CDVDMAN+0x22A4).
//	            R  the S-command status: bit 7 = busy (the submit routine waits for it to
//	               clear, in two loops — a spin when it is in interrupt context and a
//	               DelayThread when it is not), bit 6 = the result FIFO is empty (the drain
//	               loop before the command and the collect loop after it are both driven by
//	               it, and by nothing else).
//	0x1F402018  R  the S-command result FIFO, one byte per read.
//
// The interrupt cause bits are named by the handler that serves them (CDVDMAN+0x158C):
//
//	bit 0   an N-command has finished. The handler reads the error flag out of 0x2005,
//	        files the result, acknowledges, and sets event-flag bit 1 — which is the bit
//	        the reading thread is asleep on (CDVDMAN+0x2890 waits for exactly bit 1).
//	bit 2   data is ready. The handler sets event-flag bit 4 and calls the module's
//	        registered callback. Nothing on this disc has registered one yet.
//
// THE REGISTER FILE IS BYTE-WIDE, AND THAT IS NOT A DETAIL.
//
// Every access CDVDMAN makes to it is an `lbu` or an `sb`. The IOP's bus is byte-wide too,
// so the machine's own register path composes a word, merges the byte into it, and writes
// the word back (see IOP.Write) — which is right for a device whose registers are words.
// This one's are not, and two of them share a word: the S-command code at 0x2016 and its
// parameter FIFO at 0x2017. A read-modify-write of the word holding both would push a
// phantom parameter into the FIFO on every command, and re-issue the command on every
// parameter. So the drive is wired to the bus as what it is — a byte device — and the word
// path never sees it.

import (
	"fmt"
)

// The register block. CDVDMAN names its own base: the word 0xBF402000 sits in its constant
// pool (CDVDMAN+0x73A8), next to the DMA channel-3 registers it also drives.
const (
	cdvdBase = 0x1F402000
	cdvdEnd  = 0x1F402020
)

// The registers, as offsets from the base.
const (
	cdvdNCommand = 0x04
	cdvdNStatus  = 0x05 // W: a parameter. R: status.
	cdvdNMode    = 0x06 // W: the transfer mode. R: the last command's error.
	cdvdNAbort   = 0x07
	cdvdIntr     = 0x08
	cdvdDriveSt  = 0x0A
	cdvdTrayStat = 0x0B
	cdvdDiscType = 0x0F
	cdvdSCommand = 0x16
	cdvdSStatus  = 0x17 // W: a parameter. R: status.
	cdvdSResult  = 0x18
)

// The N-commands this disc issues.
const (
	// #8 is the DVD read, and CDVDMAN+0x4024 is the routine that builds it. It takes eleven
	// parameter bytes — the LBA as a little-endian word, the sector count as another, and then
	// three bytes of mode (a spindle speed chosen from a jump table, and two flags) — and it
	// carries a DMA descriptor, because this is the command that moves the disc into memory.
	cdvdNCmdReadDVD = 0x08
)

// A DVD sector, as the drive hands it over. Every number here is CDVDMAN's own: see
// readSectors for where each one is proved.
const (
	cdvdRawSectorBytes = 2064
	cdvdSectorHeader   = 12

	// Where a DVD's data area starts, in physical sectors. CDVDMAN adds 0xFFFD0000 to the ID
	// it reads out of a sector's header and expects the LBA it asked for, so the ID is the LBA
	// plus this.
	cdvdDVDDataStart = 0x30000
)

// The S-commands this disc issues, each named by the routine that issues it.
const (
	// #3 is the drive's firmware version. CDVDMAN+0x43E4 asks for it with one parameter (0)
	// and four result bytes, treats bit 7 of the first as "not ready yet" and polls until it
	// clears, and reads the other three as a 24-bit big-endian number. CDVDMAN+0x1100 then
	// compares that number against 0x16FF, 0x21FF and 0x27FF and sets three capability flags.
	cdvdSCmdVersion = 0x03

	// #5 is "have you finished?". No parameters, one result byte, and CDVDMAN+0x2C74 polls it
	// every four milliseconds until it reads zero — which is the drive saying it has settled.
	cdvdSCmdReady = 0x05
)

// The bits of 0x2005, the N-command status.
//
// The submit routine will not start a command unless `status & 0xC0 == 0x40`, so of the two
// top bits exactly one must be set and it is bit 6. Bit 7 is therefore the drive's "busy"
// and bit 6 its "ready", and a drive that is neither — or both — is one the module reports
// and gives up on. Bit 0 is the error flag the interrupt handler reads.
const (
	cdvdNStatusError = 1 << 0
	cdvdNStatusReady = 1 << 6
	cdvdNStatusBusy  = 1 << 7
)

// The bits of 0x2017, the S-command status.
const (
	cdvdSStatusNoData = 1 << 6 // the result FIFO is empty
	cdvdSStatusBusy   = 1 << 7
)

// The bits of 0x2008, the interrupt cause. Write 1 to clear.
const (
	cdvdIntrDone    = 1 << 0 // an N-command has finished
	cdvdIntrDataRdy = 1 << 2 // data is ready; the handler calls the module's callback
	cdvdIntrLine    = 2      // the interrupt line, from `interrupt 2 handled by CDVDMAN+0x158C`
	cdvdDriveReady  = 0x0A   // what 0x200A must read before an N-command may be issued
	cdvdDriveBusy   = 0x06   // and what it reads while one is in flight. Only "not 10" is
	// load-bearing: no code on this disc compares it against
	// anything else, so this value is a placeholder for a state
	// we have no evidence about, not a claim about the silicon.
	cdvdDiscTypeDVD = 20 // CDVDMAN+0x3048's DVD branch
	cdvdSectorBytes = 2048
)

// cdvdCmdLatency is how long a command takes, in IOP instructions.
//
// It is not zero, for the same reason the DMA's is not: the module writes the command byte
// and then goes on doing its own bookkeeping — releasing the drive's lock, recording what
// it asked for — before the thread that is waiting for the answer gets to run. An interrupt
// delivered on the instruction after the store would land inside the submit routine, and a
// real drive cannot possibly answer that fast.
const cdvdCmdLatency = 2000

// cdvd is the drive.
type cdvd struct {
	ps2 *Machine

	// The N-command side: the command in flight, its parameters, and the result the
	// interrupt handler will collect out of the status register.
	nCommand   byte
	nParams    []byte
	nStatus    byte
	nError     byte // what 0x2006 reads back: zero unless the command failed
	lastParams []byte
	nMode      byte
	nBusy      bool
	nDoneAt    uint64 // the IOP step count at which the interrupt arrives

	// The S-command side, which is synchronous: no interrupt, and the module polls.
	sCommand byte
	sParams  []byte
	sResult  []byte

	// The interrupt cause register, and the disc.
	intr byte

	// data is what the drive has read and not yet handed to the DMA controller, and the
	// channel-3 transfer that is waiting for it.
	data     []byte
	dmaMadr  uint32
	dmaLen   uint32
	dmaArmed bool

	// The commands nobody has claimed, and how often each was asked for. Same instrument as
	// the kernel-call census, and it is the work list for this device: a command answered
	// with nothing is a command whose caller is about to be told a lie.
	unknownN map[byte]int
	unknownS map[byte]int
}

func newCDVD(m *Machine) *cdvd {
	return &cdvd{
		ps2:      m,
		nStatus:  cdvdNStatusReady,
		unknownN: map[byte]int{},
		unknownS: map[byte]int{},
	}
}

// discType is what 0x200F reads.
//
// The disc decides, and it decides on its own evidence: a 1.74 GB image is not a CD, and
// CDVDMAN's read routine has exactly two branches — one for a type in 16..19 and one for
// the type 20. Nothing here is chosen for our convenience; if the image were a CD this
// would have to answer the other way, and the module would issue the other command.
func (c *cdvd) discType() byte {
	return cdvdDiscTypeDVD
}

// --- the bus ---------------------------------------------------------------------

func (c *cdvd) contains(a uint32) bool { return a >= cdvdBase && a < cdvdEnd }

func (c *cdvd) read(a uint32) byte {
	switch a - cdvdBase {
	case cdvdNCommand:
		return c.nCommand
	case cdvdNStatus:
		return c.nStatus
	case cdvdNMode:
		// Reading 0x2006 gives the ERROR the last command ended with, and this cost a day.
		//
		// It reads back as something quite different from what is written to it — the DMA-start
		// routine writes the transfer mode here (0x8C for a DVD), which is what led to the
		// guess, recorded in this file for a while, that a read gave the command in flight.
		// It does not, and CDVDMAN says so in three steps: its interrupt handler reads this
		// register and files the byte at state+1 (CDVDMAN+0x15B4); "what was the last error"
		// returns exactly that byte (CDVDMAN+0x2DBC, at 0x404F1 = state+1); and the routine
		// that reads a sector fails the whole read unless it is zero (CDVDMAN+0x6D4C).
		//
		// Answering with the command number meant every read on this disc came back "failed
		// with error 8" — 8 being the read command itself — and what the game printed was
		// "open fail name \DRIVERS\SIO2MAN.IRX;1", four layers up and about a file that was
		// exactly where it should have been, on a drive that had just fetched the sector
		// correctly and DMA'd it into the right place.
		return c.nError
	case cdvdIntr:
		return c.intr
	case cdvdDriveSt:
		if c.nBusy {
			return cdvdDriveBusy
		}
		return cdvdDriveReady
	case cdvdTrayStat:
		// The tray. CDVDMAN reads bit 0 of this and compares it against its own cached copy
		// (CDVDMAN+0x2CB8), and what it is looking for is a *change* — the disc having been
		// swapped while it was not watching. It never tests the bit against a constant, so what
		// this answers does not matter; that it always answers the same thing does. On this
		// machine the disc is a file and the tray does not open.
		return 0
	case cdvdDiscType:
		return c.discType()

	case cdvdSCommand:
		// The read-back. CDVDMAN reads this register immediately after writing the command to
		// it (CDVDMAN+0x22D8) and throws the answer away — a store to a register on the far
		// side of a bus, followed by a load to make sure it has got there.
		return c.sCommand

	case cdvdSStatus:
		// Never busy on a read: the S-commands here are answered the moment they are
		// issued, so the module's two wait loops fall straight through. What matters is the
		// other bit — the empty flag drives both the drain loop before the command and the
		// collect loop after it, and a drive that never says "empty" is a drive the module
		// waits on for ever. Which is exactly what it did.
		var st byte
		if len(c.sResult) == 0 {
			st |= cdvdSStatusNoData
		}
		return st

	case cdvdSResult:
		if len(c.sResult) == 0 {
			return 0
		}
		v := c.sResult[0]
		c.sResult = c.sResult[1:]
		return v
	}
	c.ps2.iopUnknownCDVD(a, false)
	return 0
}

// peek reads a register without disturbing it — no FIFO is popped. It is for the machine's
// own instruments, which have no business changing what they are watching.
func (c *cdvd) peek(a uint32) byte {
	if a-cdvdBase == cdvdSResult {
		if len(c.sResult) == 0 {
			return 0
		}
		return c.sResult[0]
	}
	return c.read(a)
}

func (c *cdvd) write(a uint32, v byte) {
	switch a - cdvdBase {
	case cdvdNCommand:
		c.startN(v)
	case cdvdNStatus:
		c.nParams = append(c.nParams, v)
	case cdvdNMode:
		c.nMode = v
	case cdvdNAbort:
		// The module writes 1 here on its error-recovery path. There is nothing in flight to
		// abort on a drive that answers immediately, and nothing reads back a result, so this
		// is recorded rather than acted on.
	case cdvdIntr:
		c.intr &^= v // write 1 to clear, exactly the bit the handler served
	case cdvdSCommand:
		c.startS(v)
	case cdvdSStatus:
		c.sParams = append(c.sParams, v)
	default:
		c.ps2.iopUnknownCDVD(a, true)
	}
}

// --- the commands ------------------------------------------------------------------

// startN runs an N-command. These are the ones that move data: they answer with an
// interrupt, and their payload arrives over DMA channel 3.
func (c *cdvd) startN(cmd byte) {
	c.nCommand = cmd
	c.nError = 0
	params := c.nParams
	c.nParams = nil

	if !c.exec(cmd, params) {
		c.unknownN[cmd]++
		if c.unknownN[cmd] == 1 {
			c.ps2.note("CDVD: N-command 0x%02X (%s) — nothing models it", cmd, hexBytes(params))
		}
	}

	// The sectors, if there are any, into the channel that has been waiting for them since
	// before the command was written.
	c.pump(c.ps2.IOP)

	// The command is accepted now and answered later, and the interrupt is the answer.
	c.nBusy = true
	c.nDoneAt = c.ps2.IOP.steps + cdvdCmdLatency
}

// exec is where a command's meaning lives. Each one has to be earned from the code that
// issues it, and the argument recorded next to it — a command answered with nothing is not
// a gap that stays where you left it, it is a lie that surfaces layers away.
func (c *cdvd) exec(cmd byte, params []byte) bool {
	switch cmd {
	case cdvdNCmdReadDVD:
		if len(params) < 8 {
			return false
		}
		lba := le32(params[0:])
		count := le32(params[4:])
		c.readSectors(lba, count)
		return true
	}
	return false
}

// readSectors stages the sectors the drive has been asked for, in the form the drive hands
// them over: whole DVD physical sectors, 2064 bytes apiece.
//
// The size and the shape are both CDVDMAN's own arithmetic, not a specification:
//
//   - CDVDMAN+0x40D4 sets the DMA's block size to 12 words — 48 bytes — and CDVDMAN+0x40D8
//     computes the block count as 43 times the number of sectors. 43 x 48 = 2064, per sector.
//   - CDVDMAN+0x3F4C points the transfer at its own staging buffer, 0x2CE20, and the loop
//     that copies the sectors out to the caller reads from 0x2CE2C (CDVDMAN+0x3CA0), which is
//     twelve bytes further on. The stride between sectors is 2064 (CDVDMAN+0x39C0).
//
// So: twelve bytes of header, the 2048 bytes the filesystem wants, and four bytes after. The
// twelve and the four are the disc's own framing and nothing on this disc reads them; they
// are handed over as zeroes rather than invented, and if something ever turns out to want
// them it will want them here.
func (c *cdvd) readSectors(lba, count uint32) {
	vol := c.ps2.vol
	if vol == nil {
		c.ps2.note("CDVD: asked to read %d sectors at LBA %d, and no disc is mounted", count, lba)
		return
	}
	for i := uint32(0); i < count; i++ {
		sec := make([]byte, cdvdRawSectorBytes)

		// The sector's own ID, and CDVDMAN checks it. This is not framing we can leave as
		// zeroes: CDVDMAN+0x3960 reads bytes 1..3 of the header as a big-endian number, adds
		// 0xFFFD0000 — which is -0x30000 — and compares what comes out against the LBA it
		// asked for. If they differ it retries, drifts, and gives up with "Read error in
		// disc_read(PVD)".
		//
		// So the number in the header is not the LBA; it is the LBA plus 0x30000, which is
		// where a DVD's data area begins. The module's own arithmetic is the whole derivation:
		// it subtracts exactly 196608 and expects the sector it asked for. A header of zeroes
		// makes that come out as -196608, and that is precisely the number CDVDMAN printed
		// when it was asked to narrate itself.
		//
		// Byte 0 is a flags byte; CDVDMAN takes bit 0 of it as the layer (CDVDMAN+0x39A4).
		// This disc is single-layer, so it is zero. Bytes 4..11 are the sector's other framing
		// and nothing on this disc reads them.
		phys := cdvdDVDDataStart + lba + i
		sec[1] = byte(phys >> 16)
		sec[2] = byte(phys >> 8)
		sec[3] = byte(phys)

		blk, err := vol.ReadBlock(int(lba + i))
		if err != nil {
			c.ps2.note("CDVD: reading LBA %d: %v", lba+i, err)
		} else {
			copy(sec[cdvdSectorHeader:], blk)
		}
		c.data = append(c.data, sec...)
	}
}

func le32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// startS runs an S-command. These are the small ones — the drive's own bookkeeping — and
// they are synchronous: the module writes the command, spins on the busy bit, and reads the
// answer out of the result FIFO.
func (c *cdvd) startS(cmd byte) {
	c.sCommand = cmd
	params := c.sParams
	c.sParams = nil
	c.sResult = nil

	if !c.execS(cmd, params) {
		c.unknownS[cmd]++
		if c.unknownS[cmd] == 1 {
			c.ps2.note("CDVD: S-command 0x%02X (%s) — nothing models it", cmd, hexBytes(params))
		}
	}
}

func (c *cdvd) execS(cmd byte, params []byte) bool {
	switch cmd {
	case cdvdSCmdVersion:
		// Four bytes: a status whose bit 7 means "ask me again", and a 24-bit firmware version.
		//
		// THE VERSION IS OURS AND NOT THE DRIVE'S, and it is written down here so that it is
		// never mistaken for an observation. Nothing on this disc tells us what a real one
		// reads; all we have is the three thresholds CDVDMAN compares it against, and a zero
		// puts it below all of them — the oldest drive there is. That is a coherent answer
		// rather than a nonsense one, and it is a safe one: the three flags it clears only gate
		// *extra* checks in the read path (CDVDMAN+0x3108 skips one of them outright when the
		// first flag is clear), and the choice between the CD and the DVD read command is made
		// from the disc type at 0x200F, not from this. If something ever behaves oddly around
		// the drive's capabilities, look here first.
		c.sResult = []byte{0, 0, 0, 0}
		return true

	case cdvdSCmdReady:
		// Zero: the drive has settled. It is the answer CDVDMAN polls for, and on a machine
		// whose disc is a file it is true the moment it is asked.
		c.sResult = []byte{0}
		return true
	}
	_ = params
	return false
}

// --- the interrupt --------------------------------------------------------------------

// tick delivers a command's completion once its latency has run out.
func (c *cdvd) tick(p *IOP) {
	if !c.nBusy || p.steps < c.nDoneAt {
		return
	}
	c.nBusy = false
	c.intr |= cdvdIntrDone
	p.raiseIRQ(cdvdIntrLine)
}

// --- the DMA ---------------------------------------------------------------------------

// arm records that channel 3 is waiting for the drive: where the sectors are to land, and
// how many bytes are wanted.
func (c *cdvd) arm(madr, n uint32) {
	c.dmaMadr, c.dmaLen, c.dmaArmed = madr, n, true
}

// pump moves the drive's data into IOP memory once both halves of the transfer exist — the
// channel armed by CDVDMAN and the sectors produced by the command. It is called from both
// ends, because either can happen first, and on this disc it is always the channel.
//
// The transfer is all-or-nothing. A disc read hands over whole sectors, and the module's own
// wait is for the start bit to go out — so a partial transfer would release it early, on a
// buffer half of which is still last time's.
func (c *cdvd) pump(p *IOP) {
	if !c.dmaArmed || uint32(len(c.data)) < c.dmaLen {
		return
	}
	for i := uint32(0); i < c.dmaLen; i++ {
		p.ram[(c.dmaMadr+i)&(iopRAMSizeBytes-1)] = c.data[i]
	}
	c.data = c.data[c.dmaLen:]
	c.dmaArmed = false

	// And now the channel may report. The start bit goes out in dmaTick, which is what
	// CDVDMAN is spinning on.
	p.dmaPending = append(p.dmaPending, iopDMADone{at: p.steps + iopDMALatency, ch: iopDMAChCDVD})
}

// --- the census -------------------------------------------------------------------------

// CDVDCensus reports what the drive was asked for that it could not answer.
func (c *cdvd) census() string {
	if len(c.unknownN) == 0 && len(c.unknownS) == 0 {
		return ""
	}
	s := "the CD/DVD drive's unanswered commands (the work list):\n"
	for _, e := range []struct {
		what string
		m    map[byte]int
	}{{"N", c.unknownN}, {"S", c.unknownS}} {
		for cmd, n := range e.m {
			s += fmt.Sprintf("      %s-command 0x%02X   %d time%s\n", e.what, cmd, n, plural(n))
		}
	}
	return s
}

func hexBytes(b []byte) string {
	if len(b) == 0 {
		return "no parameters"
	}
	s := ""
	for i, v := range b {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%02X", v)
	}
	return s
}
