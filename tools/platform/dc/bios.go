package dc

// bios.go services the syscall vectors boot.go plants — the psx/bios.go
// shape: the run loop catches a trap PC before fetching it, the call is
// serviced in Go with the game's own registers as arguments, control returns
// through PR, and every distinct call is counted by name. An unknown
// (vector, selector) pair is logged once with its registers and returns 0 —
// the census names the next syscall worth implementing.
//
// The GD-ROM interface is the documented command-queue protocol: SEND_COMMAND
// files a request and returns its id, EXEC_SERVER performs the work, and
// CHECK_COMMAND reports 1 (processing) until the server has run — the
// deliberate non-instantness the instant-I/O lesson asks for: the game must
// pump the server loop, exactly as it does on hardware.

import (
	"fmt"
	"sort"
)

// GD-ROM command numbers (the SEND_COMMAND r4 values a game uses).
const (
	gdCmdPIORead = 16
	gdCmdDMARead = 17
	gdCmdGetTOC2 = 19
	gdCmdInit    = 24
)

// gdRequest is one queued GD command. Exported fields: it savestates.
type gdRequest struct {
	Cmd    uint32
	Params [4]uint32
	Done   bool
	Result [4]uint32
}

// biosState is everything the HLE remembers between calls; it rides in the
// machine savestate.
type biosState struct {
	NextID   uint32
	Requests map[uint32]*gdRequest
}

type biosHLE struct {
	m     *Machine
	state biosState
	calls map[string]int
}

func newBIOS(m *Machine) *biosHLE {
	return &biosHLE{
		m:     m,
		state: biosState{NextID: 1, Requests: map[uint32]*gdRequest{}},
		calls: map[string]int{},
	}
}

// trapPC dispatches a caught trap address; run.go calls it before every
// fetch. Returning true means the call was serviced and PC has moved on.
func (m *Machine) trapPC(pc uint32) bool {
	if m.bios == nil || pc < trapSysinfo || pc > trapMenu {
		return false
	}
	c := m.CPU
	if c.NextIsDelaySlot() {
		return false // a trap executing as a delay slot would corrupt the pipeline
	}
	b := m.bios
	switch pc {
	case trapSysinfo:
		b.sysinfo()
	case trapRomfont:
		b.count("romfont")
		m.logf("romfont syscall (r1=%d) unimplemented; returning 0", c.R[1])
		c.R[0] = 0
	case trapFlashrom:
		b.flashrom()
	case trapGdrom:
		b.gdrom()
	case trapMenu:
		c.Halt("game called the BIOS menu vector (8C0000E0) — exit to BIOS")
		return true
	}
	c.SetPC(c.PR)
	return true
}

func (b *biosHLE) count(name string) { b.calls[name]++ }

// census lists the serviced calls and their counts.
func (b *biosHLE) census() []string {
	var names []string
	for k := range b.calls {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]string, 0, len(names))
	for _, n := range names {
		out = append(out, fmt.Sprintf("syscall %s x%d", n, b.calls[n]))
	}
	return out
}

// sysinfo is vector 8C0000B0: r7 selects.
func (b *biosHLE) sysinfo() {
	c := b.m.CPU
	switch c.R[7] {
	case 0: // INIT
		b.count("sysinfo.init")
		c.R[0] = 0
	case 3: // ID: the address of the console's 8-byte identity
		b.count("sysinfo.id")
		c.R[0] = 0x8C000068
	default:
		b.count("sysinfo.?")
		b.m.logf("sysinfo syscall r7=%d unimplemented (r4=%08X r5=%08X)", c.R[7], c.R[4], c.R[5])
		c.R[0] = 0
	}
}

// flashrom is vector 8C0000B8: r7 selects. The flash content is the erased
// state (FF), which is real for a console that has never saved settings.
func (b *biosHLE) flashrom() {
	c := b.m.CPU
	switch c.R[7] {
	case 0: // INFO(r4=partition, r5=dest): the partition's {offset, size}
		b.count("flashrom.info")
		parts := [5][2]uint32{
			{0x1A000, 0x2000}, // factory
			{0x18000, 0x2000}, // reserved
			{0x1C000, 0x4000}, // block 1: user settings
			{0x10000, 0x8000}, // game settings
			{0x00000, 0x10000},
		}
		if p := c.R[4]; p < 5 {
			b.m.Write32(c.R[5]&0x1FFFFFFF, parts[p][0])
			b.m.Write32((c.R[5]+4)&0x1FFFFFFF, parts[p][1])
			c.R[0] = 0
		} else {
			c.R[0] = ^uint32(0)
		}
	case 1: // READ(r4=offset, r5=buffer, r6=count) -> bytes read
		b.count("flashrom.read")
		off, buf, n := c.R[4], c.R[5], c.R[6]
		if off >= FlashSize {
			c.R[0] = ^uint32(0)
			return
		}
		if off+n > FlashSize {
			n = FlashSize - off
		}
		for i := uint32(0); i < n; i++ {
			b.m.Write8((buf+i)&0x1FFFFFFF, b.m.Flash[off+i])
		}
		c.R[0] = n
	default:
		b.count("flashrom.?")
		b.m.logf("flashrom syscall r7=%d unimplemented (r4=%08X r5=%08X r6=%08X)", c.R[7], c.R[4], c.R[5], c.R[6])
		c.R[0] = ^uint32(0)
	}
}

// gdrom is vector 8C0000BC: r6=0 is the GD-ROM interface, r6=-1 misc.
func (b *biosHLE) gdrom() {
	c := b.m.CPU
	if c.R[6] == ^uint32(0) { // misc
		b.count("gdrom.misc")
		c.R[0] = 0
		return
	}
	if c.R[6] != 0 {
		b.count("gdrom.?")
		b.m.logf("vector BC with r6=%08X unimplemented", c.R[6])
		c.R[0] = 0
		return
	}
	switch c.R[7] {
	case 0: // SEND_COMMAND(r4=cmd, r5=params) -> request id
		b.count(fmt.Sprintf("gdrom.send.%d", c.R[4]))
		req := &gdRequest{Cmd: c.R[4]}
		for i := uint32(0); i < 4; i++ {
			req.Params[i] = b.m.ram32(c.R[5] + 4*i)
		}
		id := b.state.NextID
		b.state.NextID++
		b.state.Requests[id] = req
		c.R[0] = id
	case 1: // CHECK_COMMAND(r4=id, r5=status buffer) -> 0 none / 1 processing / 2 done
		b.count("gdrom.check")
		req, ok := b.state.Requests[c.R[4]]
		switch {
		case !ok:
			c.R[0] = 0
		case !req.Done:
			c.R[0] = 1
		default:
			for i := uint32(0); i < 4; i++ {
				b.m.Write32((c.R[5]+4*i)&0x1FFFFFFF, req.Result[i])
			}
			delete(b.state.Requests, c.R[4])
			c.R[0] = 2
		}
	case 2: // EXEC_SERVER: perform the queued work
		b.count("gdrom.exec")
		for _, req := range b.state.Requests {
			if !req.Done {
				b.execGD(req)
			}
		}
		c.R[0] = 0
	case 3: // INIT
		b.count("gdrom.init")
		c.R[0] = 0
	case 4: // CHECK_DRIVE(r4=buffer): drive status + disc type. A drive that
		// has served reads sits in PAUSE (1), not STANDBY — the disc-check
		// state machine distinguishes them.
		b.count("gdrom.checkdrive")
		b.m.Write32(c.R[4]&0x1FFFFFFF, 0x01)     // pause
		b.m.Write32((c.R[4]+4)&0x1FFFFFFF, 0x80) // GD-ROM
		c.R[0] = 0
	default:
		b.count(fmt.Sprintf("gdrom.?%d", c.R[7]))
		b.m.logf("gdrom syscall r7=%d unimplemented (r4=%08X r5=%08X)", c.R[7], c.R[4], c.R[5])
		c.R[0] = 0
	}
}

// execGD performs one queued command against the disc.
func (b *biosHLE) execGD(req *gdRequest) {
	switch req.Cmd {
	case gdCmdPIORead, gdCmdDMARead:
		// params: {FAD, sector count, buffer, x}; the game passes frame
		// addresses (LBA+150), the convention the whole disc chain uses.
		fad, count, buf := req.Params[0], req.Params[1], req.Params[2]
		if b.m.OnGDRead != nil {
			b.m.OnGDRead(fad, count)
		}
		for i := uint32(0); i < count; i++ {
			sec, err := b.m.Disc.ReadSector(int(fad-150) + int(i))
			if err != nil {
				b.m.logf("gdrom read FAD %d failed: %v", fad+i, err)
				req.Result[0] = ^uint32(0)
				break
			}
			for j, by := range sec {
				b.m.Write8((buf+i*2048+uint32(j))&0x1FFFFFFF, by)
			}
		}
	case gdCmdGetTOC2:
		// params: {session, buffer}. 102 words: 99 track entries
		// (ctrl<<28|adr<<24|FAD), then first, last, leadout.
		buf := req.Params[1]
		for i := uint32(0); i < 99; i++ {
			b.m.Write32((buf+4*i)&0x1FFFFFFF, 0xFFFFFFFF)
		}
		d := b.m.Disc
		var first, last int
		for _, t := range d.Tracks {
			if !t.IsData() || t.StartLBA < 0 {
				continue
			}
			if first == 0 {
				first = t.Number
			}
			last = t.Number
			entry := uint32(0x4)<<28 | uint32(1)<<24 | uint32(t.StartLBA+150)&0xFFFFFF
			b.m.Write32((buf+4*uint32(t.Number-1))&0x1FFFFFFF, entry)
		}
		end := d.data.StartLBA + int(d.data.Length/rawSector) + 150
		b.m.Write32((buf+4*99)&0x1FFFFFFF, uint32(0x4)<<28|uint32(first)<<16)
		b.m.Write32((buf+4*100)&0x1FFFFFFF, uint32(0x4)<<28|uint32(last)<<16)
		b.m.Write32((buf+4*101)&0x1FFFFFFF, uint32(end)&0xFFFFFF)
	case gdCmdInit:
		// nothing to spin up
	default:
		b.m.logf("gdrom command %d unimplemented (params %08X %08X %08X %08X)",
			req.Cmd, req.Params[0], req.Params[1], req.Params[2], req.Params[3])
	}
	req.Done = true
}
