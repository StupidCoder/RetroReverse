package n3ds

// gsp_vblank.go delivers the graphics heartbeat. On real hardware the GPU raises
// a VBlank interrupt ~60 times a second; the GSP module relays it to the
// application by pushing an entry into the interrupt queue in GSP shared memory
// and signalling the per-process GSP event. The application's GSP event thread
// wakes, drains the queue, and signals the specific per-interrupt event (e.g.
// VBlank0) the game's render loop is blocked on. This whole path rides on the
// Phase-1 event/wait machinery — Horizon delivers GPU interrupts as event
// signals through shared memory, not as ARM IRQ vectoring.
//
// Pacing is by retired-instruction count (not wall clock) so a resumed savestate
// stays deterministic, mirroring the N64 VI's stepsPerField accumulator.

// stepsPerFrame is how many retired ARM instructions correspond to one display
// frame. The ARM11 runs at ~268 MHz and the LCDs refresh at ~60 Hz.
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
	return m.gspEvent != 0 && m.CPU.Instrs >= m.nextFrameInstr
}

// deliverVBlank pushes the VBlank interrupts into the GSP shared-memory queue
// and signals the GSP event, waking the game's GSP event thread.
func (m *Machine) deliverVBlank() {
	m.vblankCount++
	m.nextFrameInstr = m.CPU.Instrs + stepsPerFrame

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
	pos := (uint32(idx) + uint32(cnt)) % listLen
	m.Write(base+0xC+pos, id)
	m.Write(base+1, cnt+1)
}

// VBlanks reports how many VBlank interrupts have been delivered to the game.
func (m *Machine) VBlanks() uint64 { return m.vblankCount }
