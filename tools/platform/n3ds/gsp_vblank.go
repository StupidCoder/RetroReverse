package n3ds

import "fmt"

// gsp_vblank.go delivers the graphics heartbeat. On real hardware the GPU raises
// a VBlank interrupt ~60 times a second; the GSP module relays it to the
// application by pushing an entry into the interrupt queue in GSP shared memory
// and signalling the per-process GSP event. The application's GSP event thread
// wakes, drains the queue, and signals the specific per-interrupt event (e.g.
// VBlank0) the game's render loop is blocked on. This whole path rides on the
// Phase-1 event/wait machinery — Horizon delivers GPU interrupts as event
// signals through shared memory, not as ARM IRQ vectoring.
//
// Pacing is by the machine's monotonic tick (not wall clock) so a resumed
// savestate stays deterministic, mirroring the N64 VI's stepsPerField
// accumulator. The tick — not CPU.Instrs — is the machine-global clock:
// each thread context carries its own retired-instruction counter, so once
// work migrates to younger threads (the frame pacer runs on the APT thread)
// an Instrs comparison would freeze the heartbeat.

// stepsPerFrame is one display frame in system ticks (~268 MHz / ~60 Hz).
const stepsPerFrame = sysclockHz / 60

// GSP interrupt ids (GSPGPU_Event), as delivered in the shared-memory queue.
const (
	gspIntPSC0    = 0
	gspIntPSC1    = 1
	gspIntVBlank0 = 2
	gspIntVBlank1 = 3
	gspIntPPF     = 4
	gspIntP3D     = 5
	gspIntDMA     = 6
)

// vblankDue reports whether it is time to deliver the next VBlank.
func (m *Machine) vblankDue() bool {
	return m.gspEvent != 0 && m.instrs >= m.nextFrameInstr
}

// FBPresent is one consumed framebuffer-info entry — the framebuffer the GSP
// module pointed the LCD at on behalf of the application.
type FBPresent struct {
	Active     uint32 // which of the entry's two framebuffers is front
	AddrLeft   uint32 // framebuffer virtual address (left eye)
	AddrRight  uint32 // right-eye address (2D: same as left)
	Stride     uint32 // bytes per row
	Format     uint32 // GSP framebuffer format word
	DispSelect uint32
	Valid      bool
}

// Scanout is what the LCD is pointed at for a screen, as the GSP last applied it.
//
// This is the only authority on "what is on the screen". A DisplayTransfer writes a
// framebuffer; it does not make it visible. The game double-buffers, and it renders the
// top screen TWICE — once per eye — so the last transfer to a screen is not the picture
// being shown, it is the right eye of it, which with the 3D slider down is a cleared
// buffer nobody looks at. AddrLeft is the one the panel scans.
//
// screen is 0 for the top, 1 for the bottom.
func (m *Machine) Scanout(screen int) FBPresent {
	if screen < 0 || screen > 1 {
		return FBPresent{}
	}
	return m.screenFB[screen]
}

// deliverVBlank pushes the VBlank interrupts into the GSP shared-memory queue
// and signals the GSP event, waking the game's GSP event thread.
func (m *Machine) deliverVBlank() {
	m.vblankCount++
	m.nextFrameInstr = m.instrs + stepsPerFrame

	m.consumeFBInfo()
	m.pushGSPInterrupt(gspIntVBlank0)
	m.pushGSPInterrupt(gspIntVBlank1)
	m.signalGSPEvent()

	// Publish a fresh HID input sample (the input driver's per-frame job), so the
	// game's pad polling sees live button state and any injected -keys press.
	m.updateHIDShared()

	// A pending APT wake (NotifyToWait) is delivered here, asynchronously to
	// the request that armed it — by now the requester has released the APT
	// session and parked (see ipcAPT 0x0043).
	if m.aptWakePending {
		m.aptWakePending = false
		m.signalAPTEvents()
	}

	// Close the frame's timing buckets (profile.go) before the hook below, so a
	// debugger that stops on the frame boundary reads the frame it just watched.
	m.profFrame()

	// The debugger's frame boundary (debug.go). Last, so a hook that stops the run
	// here sees a fully delivered VBlank: the swap consumed, the interrupts queued,
	// the pad sampled.
	if m.OnFrame != nil {
		m.OnFrame(m)
	}
}

// consumeFBInfo is the GSP module's VBlank-side of the buffer-swap protocol.
// The application presents a frame by writing a framebuffer-info entry into GSP
// shared memory (top screen at +0x200, bottom at +0x240 for interrupt-relay
// thread 0) and flagging it: header byte 0 = the index of the entry just
// written, byte 1 = the new-data flag. Each VBlank the GSP module applies the
// flagged entry to the LCD framebuffer registers and clears the flag — and the
// game's present fence waits for exactly that clear before starting its next
// frame (Captain Toad: writer 0x00126F28, fence loop 0x00271840 polling the
// flag via 0x003428C8; leaving the flag set blocks the engine before its first
// draw). Entry layout (traced from the writer): {active_framebuf, fb0_vaddr,
// fb1_vaddr, stride, format, dispselect, attr}, 0x1C bytes, two entries after
// the 4-byte header.
func (m *Machine) consumeFBInfo() {
	if m.gspSharedAddr == 0 {
		return
	}
	for screen := uint32(0); screen < 2; screen++ {
		base := m.gspSharedAddr + 0x200 + screen*0x40
		if m.Read(base+1) == 0 {
			continue
		}
		idx := uint32(m.Read(base)) & 1
		e := base + 4 + idx*0x1C
		m.screenFB[screen] = FBPresent{
			Active:     m.ReadWord(e),
			AddrLeft:   m.ReadWord(e + 4),
			AddrRight:  m.ReadWord(e + 8),
			Stride:     m.ReadWord(e + 0xC),
			Format:     m.ReadWord(e + 0x10),
			DispSelect: m.ReadWord(e + 0x14),
			Valid:      true,
		}
		m.Write(base+1, 0)
		m.framesSwapped++
	}
}

// signalGSPEvent signals the per-process GSP event, waking the game's GSP event
// thread to drain the shared-memory interrupt queue.
func (m *Machine) signalGSPEvent() {
	if obj := m.handles[m.gspEvent]; obj != nil {
		obj.signal = true
		if m.signalObject(obj) {
			m.reschedule = true
		}
	}
}

// pushGSPInterrupt appends one interrupt id to GSP thread 0's relay queue in
// shared memory. Layout (derived by tracing the game's GSP event thread): a
// per-thread 0x40-byte header block starting at the shared-memory base — byte 0
// is the read index, byte 1 the pending count, and the interrupt-id list begins
// at byte 0xC, wrapping at 0x34 entries.
func (m *Machine) pushGSPInterrupt(id byte) {
	if m.gspSharedAddr == 0 {
		return
	}
	base := m.gspSharedAddr // GSP thread index 0
	idx := m.Read(base + 0)
	cnt := m.Read(base + 1)
	const listLen = 0x34
	if uint32(cnt) >= listLen {
		// The queue is full: the app's GSP event thread has stopped draining it. The
		// hardware queue cannot grow, so neither may this one — the old code carried
		// on writing, running the count past 0x34 (cnt=98 was seen) and scribbling
		// interrupt ids through whatever follows the ring in GSP shared memory. Drop
		// the interrupt, as a full queue must, and say so once: an overflow is never
		// the disease, it is the sign that something upstream has parked the thread
		// that should be consuming (for a whole session, that was the title
		// suspending itself — see signalAPTEvents).
		if !m.gspOverflowed {
			m.gspOverflowed = true
			fmt.Printf("GSP interrupt queue overflow (id=%d idx=%d cnt=%d at instr %d): the app has stopped "+
				"draining it — interrupts are being dropped, and something has blocked its GSP thread\n",
				id, idx, cnt, m.instrs)
		}
		return
	}
	pos := (uint32(idx) + uint32(cnt)) % listLen
	m.Write(base+0xC+pos, id)
	m.Write(base+1, cnt+1)
}

// VBlanks reports how many VBlank interrupts have been delivered to the game.
func (m *Machine) VBlanks() uint64 { return m.vblankCount }
