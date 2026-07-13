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

// audioBuf maps a page of application memory at the address a DSP physical
// address resolves to (FCRAM 0x2xxxxxxx → the linear heap) and fills it with
// PCM16 samples.
func audioBuf(m *Machine, phys uint32, samples []int16) {
	virt := m.gpuAddrToVirt(phys)
	if m.regionOf(virt) == nil {
		m.mapRegion("audio", virt&^0xFFF, make([]byte, 0x10000))
	}
	for i, v := range samples {
		m.dspWrite16(virt+uint32(i)*2, uint16(v))
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

	// 400 mono PCM16 samples = 2.5 frames of audio at rate 1.
	const phys = 0x20100000
	pcm := make([]int16, 400)
	for i := range pcm {
		pcm[i] = int16(i)
	}
	audioBuf(m, phys, pcm)

	// The app writes region 0 (counter 2 > 0) with source 0: enabled, one
	// embedded buffer, mono PCM16, rate 1.
	m.dspWrite16(dspRegion0+dspOffFrameCounter, 2)
	cfg := uint32(dspRegion0 + dspOffSourceConfigs)
	m.WriteWord(cfg+srcCfgDirty, dirtyEnable|dirtyEmbeddedBuffer|dirtyRate|dirtySyncCount)
	m.Write(cfg+srcCfgEnable, 1)
	m.dspWrite16(cfg+srcCfgSyncCount, 7)
	m.WriteWord(cfg+srcCfgRate, 0x3F800000) // 1.0f
	m.dspWrite32(cfg+srcCfgEmbPhysAddr, phys)
	m.dspWrite32(cfg+srcCfgEmbLength, 400)
	m.dspWrite16(cfg+srcCfgFlags1, 1|dspFmtPCM16<<2) // mono, PCM16
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

	// Frame 1 dequeued the buffer; frames 2-4 consume 160+160+80 samples, and the
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

// The samples come out the other end: a mono PCM16 buffer at rate 1, unity gain
// on mix 0 and master volume 1, must appear verbatim in the final mix — which
// is what makes the mixer usable as a verification oracle.
func TestDSPMixesSamplesThrough(t *testing.T) {
	m := newDSPMachine()
	m.dsp.ComponentLoaded = true
	m.dsp.State = dspStateOn

	const phys = 0x20200000
	pcm := make([]int16, 320)
	for i := range pcm {
		pcm[i] = int16(100 + i)
	}
	audioBuf(m, phys, pcm)

	m.dspWrite16(dspRegion0+dspOffFrameCounter, 2)
	cfg := uint32(dspRegion0 + dspOffSourceConfigs)
	m.WriteWord(cfg+srcCfgDirty, dirtyEnable|dirtyEmbeddedBuffer|dirtyRate|dirtyGain0)
	m.Write(cfg+srcCfgEnable, 1)
	m.WriteWord(cfg+srcCfgRate, 0x3F800000)   // 1.0f
	m.WriteWord(cfg+srcCfgGain+0, 0x3F800000) // front left  = 1.0
	m.WriteWord(cfg+srcCfgGain+4, 0x3F800000) // front right = 1.0
	m.dspWrite32(cfg+srcCfgEmbPhysAddr, phys)
	m.dspWrite32(cfg+srcCfgEmbLength, 320)
	m.dspWrite16(cfg+srcCfgFlags1, 1|dspFmtPCM16<<2)
	m.dspWrite16(cfg+srcCfgEmbBufferID, 1)

	// The DSP configuration: master volume 1.0, stereo out.
	dcfg := uint32(dspRegion0 + dspOffDSPConfig)
	m.WriteWord(dcfg+cfgDirty, cfgDirtyMasterVol|cfgDirtyOutFormat)
	m.WriteWord(dcfg+cfgMasterVolume, 0x3F800000)
	m.dspWrite16(dcfg+cfgOutputFormat, dspOutStereo)

	m.dspTick() // dequeues the buffer (a frame of silence)
	m.dspTick() // plays the buffer's first frame

	// The resampler carries two input samples of history across the frame
	// boundary, so a freshly started voice leads in with two zeros — the same
	// two-sample latency the firmware has.
	final := uint32(dspRegion1 + dspOffFinalSamples)
	for i := 2; i < 160; i++ {
		l := int16(m.dspRead16(final + uint32(i)*4))
		r := int16(m.dspRead16(final + uint32(i)*4 + 2))
		if want := pcm[i-2]; l != want || r != want {
			t.Fatalf("final mix sample %d = (%d,%d), want (%d,%d)", i, l, r, want, want)
		}
	}
}

// PCM8 is unsigned-biased into the top byte; a stereo buffer interleaves its
// channels. Byte↔sample arithmetic is the whole point of the format field.
func TestDSPDecodePCM8Stereo(t *testing.T) {
	m := newDSPMachine()
	const phys = 0x20300000
	virt := m.gpuAddrToVirt(phys)
	m.mapRegion("audio8", virt&^0xFFF, make([]byte, 0x1000))
	for i := 0; i < 8; i++ {
		m.Write(virt+uint32(i)*2, byte(i))       // left
		m.Write(virt+uint32(i)*2+1, byte(128+i)) // right
	}
	s := &m.dsp.Sources[0]
	got := m.dspDecodeBuffer(s, dspBuffer{PhysAddr: phys, Length: 8, Format: dspFmtPCM8, Stereo: true})
	if len(got) != 8 {
		t.Fatalf("decoded %d samples, want 8", len(got))
	}
	for i := range got {
		wantL, wantR := int16(uint16(i)<<8), int16(uint16(128+i)<<8)
		if got[i][0] != wantL || got[i][1] != wantR {
			t.Fatalf("sample %d = (%d,%d), want (%d,%d)", i, got[i][0], got[i][1], wantL, wantR)
		}
	}
}

// ADPCM: 8-byte blocks carry a header (scale + coefficient index) and 14 4-bit
// samples. With the predictor taps zeroed, each nibble decodes to its signed
// value times the scale — which pins the nibble sign handling and the block
// arithmetic independently of the filter.
func TestDSPDecodeADPCM(t *testing.T) {
	m := newDSPMachine()
	const phys = 0x20400000
	virt := m.gpuAddrToVirt(phys)
	m.mapRegion("audioA", virt&^0xFFF, make([]byte, 0x1000))
	// Header: scale 2^2, coefficient index 0 (whose taps we leave at zero).
	m.Write(virt+0, 0x02)
	m.Write(virt+1, 0x1F) // nibbles 1 and 15 (= -1)
	m.Write(virt+2, 0x80) // nibbles 8 (= -8) and 0

	s := &m.dsp.Sources[0]
	got := m.dspDecodeBuffer(s, dspBuffer{PhysAddr: phys, Length: 4, Format: dspFmtADPCM})
	want := []int16{4, -4, -32, 0} // nibble × scale(4), rounding 0.5 toward +inf
	if len(got) != len(want) {
		t.Fatalf("decoded %d samples, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i][0] != w || got[i][1] != w {
			t.Fatalf("adpcm sample %d = %d, want %d (mono fills both channels)", i, got[i][0], w)
		}
	}
}

// A looping buffer never exhausts: it re-queues itself and restarts at 0.
func TestDSPSourceLoops(t *testing.T) {
	m := newDSPMachine()
	const phys = 0x20500000
	audioBuf(m, phys, make([]int16, 100))
	s := &m.dsp.Sources[0]
	*s = dspSource{Enabled: true, Rate: 1}
	s.Queue = []dspBuffer{{PhysAddr: phys, Length: 100, BufferID: 3, IsLooping: true, Format: dspFmtPCM16}}
	for i := 0; i < 50; i++ {
		m.dspSourceFrame(0)
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
	m := newDSPMachine()
	const phys = 0x20600000
	audioBuf(m, phys, make([]int16, 160))
	s := &m.dsp.Sources[0]
	*s = dspSource{Enabled: true, Rate: 1}
	s.Queue = []dspBuffer{
		{PhysAddr: phys, Length: 160, BufferID: 9, FromQueue: true, Format: dspFmtPCM16},
		{PhysAddr: phys, Length: 160, BufferID: 8, FromQueue: true, Format: dspFmtPCM16},
	}
	m.dspSourceFrame(0) // dequeues id 8
	if s.CurBufferID != 8 || !s.BufferUpdate {
		t.Fatalf("first dequeue: id=%d update=%v, want 8/true", s.CurBufferID, s.BufferUpdate)
	}
	s.BufferUpdate = false
	m.dspSourceFrame(0) // consumes id 8 (one frame's worth)
	m.dspSourceFrame(0) // dequeues id 9
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

	// Each audio frame raises BOTH the pipe-2 interrupt and the frame semaphore.
	// The pipe interrupt is the one that matters: it is the event the app's sound
	// thread blocks on (Captain Toad registers it, unregisters it, registers a
	// second one, and waits on that handle forever).
	m.instrs += dspFrameTicks
	m.dspTick()
	if !m.handles[sem].signal {
		t.Fatal("pipeline On: each audio frame must raise the frame semaphore")
	}
	if !m.handles[ev].signal {
		t.Fatal("pipeline On: each audio frame must raise the pipe-2 interrupt — the event the sound thread waits on")
	}
}
