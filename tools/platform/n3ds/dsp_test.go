package n3ds

import "testing"

// newDSPMachine is newSyncMachine plus the DSP RAM window and event map —
// enough to exercise the DSP model without booting an image.
func newDSPMachine() *Machine {
	m := newSyncMachine()
	m.mapRegion("dspram", dspRAMBase, make([]byte, dspRAMSize))
	m.dsp.IntEvents = map[uint32]uint32{}
	return m
}

// The audio pipe's Initialize answer: a count of 15, then the shared-memory
// structure addresses as DSP words, and each must convert back to the byte
// offset it names via the ConvertProcessAddressFromDspDram formula.
func TestDSPAnnounceStructs(t *testing.T) {
	m := newDSPMachine()
	m.dspAnnounceStructs()

	p := m.dsp.Pipes[dspPipeAudio]
	if len(p) != 32 {
		t.Fatalf("audio pipe holds %d bytes, want 32", len(p))
	}
	if n := uint16(p[0]) | uint16(p[1])<<8; n != 15 {
		t.Fatalf("struct count = %d, want 15", n)
	}
	// First entry is the frame counter; converting its DSP word address must
	// land on region 0's trailing u16.
	w := uint32(p[2]) | uint32(p[3])<<8
	if got := (w << 1) + dspRAMBase + 0x40000; got != dspRegion0+dspOffFrameCounter {
		t.Fatalf("frame-counter pointer = 0x%08X, want 0x%08X", got, dspRegion0+dspOffFrameCounter)
	}
	// Second entry is the source-configuration array.
	w = uint32(p[4]) | uint32(p[5])<<8
	if got := (w << 1) + dspRAMBase + 0x40000; got != dspRegion0+dspOffSourceConfigs {
		t.Fatalf("source-config pointer = 0x%08X, want 0x%08X", got, dspRegion0+dspOffSourceConfigs)
	}
}

// The double-buffer handoff: the region with the higher frame counter is read,
// the other written, with wraparound at 0xFFFF.
func TestDSPRegionSelection(t *testing.T) {
	m := newDSPMachine()
	set := func(c0, c1 uint16) {
		m.dspWrite16(dspRegion0+dspOffFrameCounter, c0)
		m.dspWrite16(dspRegion1+dspOffFrameCounter, c1)
	}
	set(4, 2)
	if m.dspReadRegion() != dspRegion0 || m.dspWriteRegion() != dspRegion1 {
		t.Fatal("higher counter in region 0 must make it the read region")
	}
	set(2, 4)
	if m.dspReadRegion() != dspRegion1 {
		t.Fatal("higher counter in region 1 must make it the read region")
	}
	set(0xFFFF, 0x0001) // region 1 wrapped past region 0
	if m.dspReadRegion() != dspRegion1 {
		t.Fatal("wraparound: counter 0x0001 must beat 0xFFFF")
	}
}

// A source with a finite buffer plays it out over the right number of frames,
// then reports completion the way streaming code watches for it: the buffer id
// moves to last_buffer_id, current goes to 0, the source disables itself, and
// the status dirty byte fires exactly once.
func TestDSPSourcePlaysOutBuffer(t *testing.T) {
	m := newDSPMachine()
	m.dsp.ComponentLoaded = true
	m.dsp.State = dspStateOn

	// The app writes region 0 (counter 2 > 0) with source 0: enabled, one
	// embedded buffer of 400 samples at rate 1 (= 2.5 frames of audio).
	m.dspWrite16(dspRegion0+dspOffFrameCounter, 2)
	cfg := uint32(dspRegion0 + dspOffSourceConfigs)
	m.WriteWord(cfg+srcCfgDirty, dirtyEnable|dirtyEmbeddedBuffer|dirtyRate|dirtySyncCount)
	m.Write(cfg+srcCfgEnable, 1)
	m.dspWrite16(cfg+srcCfgSyncCount, 7)
	m.WriteWord(cfg+srcCfgRate, 0x3F800000) // 1.0f
	m.dspWrite32(cfg+srcCfgEmbPhysAddr, 0x1000000)
	m.dspWrite32(cfg+srcCfgEmbLength, 400)
	m.dspWrite16(cfg+srcCfgEmbBufferID, 5)

	st := uint32(dspRegion1 + dspOffSourceStatus) // write region is region 1
	m.dspTick()
	if m.ReadWord(cfg+srcCfgDirty) != 0 {
		t.Fatal("the DSP must clear the config dirty flags after consuming them")
	}
	if m.Read(st) != 1 {
		t.Fatal("status must report the source enabled")
	}
	if got := m.dspRead16(st + 2); got != 7 {
		t.Fatalf("status sync_count = %d, want 7", got)
	}
	if got := m.dspRead16(st + 8); got != 5 {
		t.Fatalf("current_buffer_id = %d, want 5", got)
	}

	// Frame 1 dequeues (no consumption), frames 2-4 consume 160+160+80; the
	// frame after that finds the queue dry.
	for i := 0; i < 4; i++ {
		m.dspTick()
	}
	if m.Read(st) != 0 {
		t.Fatal("source must disable itself once the queue runs dry")
	}
	if got := m.dspRead16(st + 10); got != 5 {
		t.Fatalf("last_buffer_id = %d, want 5 (the finished buffer)", got)
	}
	if got := m.dspRead16(st + 8); got != 0 {
		t.Fatalf("current_buffer_id = %d, want 0 after playout", got)
	}
	if m.Read(st+1) != 1 {
		t.Fatal("the buffer-update dirty byte must fire on completion")
	}
	m.dspTick()
	if m.Read(st+1) != 0 {
		t.Fatal("the buffer-update dirty byte must fire exactly once")
	}
}

// A looping buffer never exhausts: it re-queues itself and restarts at 0.
func TestDSPSourceLoops(t *testing.T) {
	s := dspSource{Enabled: true, Rate: 1}
	s.Queue = []dspBuffer{{Length: 100, BufferID: 3, IsLooping: true}}
	for i := 0; i < 50; i++ {
		s.advanceFrame()
	}
	if !s.Enabled {
		t.Fatal("a looping source must stay enabled")
	}
	if s.CurBufferID != 3 {
		t.Fatalf("current_buffer_id = %d, want 3", s.CurBufferID)
	}
}

// Queued buffers play in buffer-id order regardless of arrival order, and each
// queue buffer's first play raises the update flag with the new id.
func TestDSPSourceQueueOrder(t *testing.T) {
	s := dspSource{Enabled: true, Rate: 1}
	s.Queue = []dspBuffer{
		{Length: 160, BufferID: 9, FromQueue: true},
		{Length: 160, BufferID: 8, FromQueue: true},
	}
	s.advanceFrame() // dequeues id 8
	if s.CurBufferID != 8 || !s.BufferUpdate {
		t.Fatalf("first dequeue: id=%d update=%v, want 8/true", s.CurBufferID, s.BufferUpdate)
	}
	s.BufferUpdate = false
	s.advanceFrame() // consumes id 8 fully (exactly one frame's worth)
	s.advanceFrame() // dequeues id 9
	if s.CurBufferID != 9 {
		t.Fatalf("second buffer: id=%d, want 9", s.CurBufferID)
	}
}

// The DSP state machine over the audio pipe, as the service sees it: the frame
// clock arms at LoadComponent, Initialize turns the pipeline on (RecvData(0)
// reads 0), Shutdown turns it off (reads 1) and stops the pipe interrupts.
func TestDSPStateMachine(t *testing.T) {
	m := newDSPMachine()
	if m.dspDue() {
		t.Fatal("no component loaded: the frame clock must not run")
	}
	m.dsp.ComponentLoaded = true
	m.dsp.NextFrame = m.instrs + dspFrameTicks
	if m.dsp.State != dspStateOff {
		t.Fatal("fresh component: state must be Off")
	}

	// The game's interrupt event for pipe 2, and the semaphore event.
	ev := m.newHandle("event", false)
	m.dsp.IntEvents[dspIntPipe<<8|dspPipeAudio] = ev
	sem := m.newHandle("event", false)
	m.dsp.SemEvent = sem

	m.instrs += dspFrameTicks
	m.dspTick()
	if m.handles[ev].signal {
		t.Fatal("pipeline Off: the audio-pipe interrupt must not fire")
	}
	if !m.handles[sem].signal {
		t.Fatal("the frame semaphore must fire from component boot, pipeline on or off")
	}
	m.handles[sem].signal = false

	m.dsp.Pipes = [8][]byte{}
	m.dspAnnounceStructs()
	m.dsp.State = dspStateOn // what an audio-pipe Initialize write does
	if !m.handles[ev].signal {
		t.Fatal("announcing the struct addresses must raise the pipe interrupt")
	}
	m.handles[ev].signal = false

	// The per-frame heartbeat is the SEMAPHORE, and only the semaphore. A pipe
	// interrupt means "there is a pipe message to read" — raising one every
	// frame makes an app that counts signals run its audio frame twice per DSP
	// frame (Captain Toad: its sync-count handshake then never re-matches and
	// its voice-command list corrupts).
	m.instrs += dspFrameTicks
	m.dspTick()
	if !m.handles[sem].signal {
		t.Fatal("pipeline On: each audio frame must raise the frame semaphore")
	}
	if m.handles[ev].signal {
		t.Fatal("an audio frame carries no pipe message: it must NOT raise the pipe interrupt")
	}
}
