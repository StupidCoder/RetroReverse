package n3ds

// dsp.go models the 3DS DSP — the CEVA TeakLite-class audio coprocessor — at the
// level of the dsp::DSP service and the protocol Nintendo's audio firmware
// speaks over it. The firmware itself is NOT run: LoadComponent hands us a
// signed Teak binary (0x100-byte signature + "DSP1" container), and executing
// it would mean a whole second CPU core. Instead this HLEs what that firmware
// *does*, the same posture as the Horizon kernel HLE (svc.go):
//
//   - a 512 KiB DSP RAM window at 0x1FF00000, whose second half holds two
//     0x8000-byte shared-memory regions (0x1FF50000 / 0x1FF70000) the app and
//     the DSP exchange once per audio frame, double-buffered by a trailing
//     frame counter the APP increments; the DSP reads whichever region has the
//     higher counter and writes the other,
//   - byte pipes between CPU and DSP; pipe 2 (audio) carries the control
//     protocol: the app writes a state change (Initialize/Shutdown/Wakeup/
//     Sleep), the DSP answers Initialize/Wakeup by writing the DSP-word
//     addresses of the 15 shared-memory structures and raising the pipe
//     interrupt; the app converts each address to a process pointer via
//     ConvertProcessAddressFromDspDram,
//   - an audio frame tick every 160 samples ≈ 1,310,720 ARM11 cycles: the DSP
//     consumes the per-source configurations (clearing their dirty flags),
//     advances each source's buffer queue, publishes per-source statuses, and
//     raises the pipe-2 interrupt and the frame semaphore the app's sound
//     thread blocks on. THIS is what was load-bearing: without the DSP the
//     sound thread has nothing to wait on and free-runs at its (higher)
//     priority, starving the whole game (Captain Toad, writeup Part IV).
//
// What the sources model is the CONTROL protocol, not the audio: buffer
// positions advance by sample count and buffer ids complete in order, so the
// app's streaming logic sees coherent progress — but no PCM is decoded or
// mixed. Audio fidelity is a later phase.
//
// Clean-room note: the DSP is platform hardware, not game data, so this file
// is built from platform-level documentation under the user-approved exception
// (like the PSP KIRK keys): the protocol, structure layout and per-source
// semantics follow 3dbrew ("DSP Services", "DSP Memory") and the Citra
// project's HLE (src/audio_core/hle/{hle.cpp,shared_memory.h,source.cpp},
// src/core/hle/service/dsp/dsp_dsp.cpp), reimplemented in Go. Everything about
// what the TITLES do with it still comes from tracing their code with our own
// tools. Two deltas from Citra, both forced by traced game behaviour: the
// frame clock arms at LoadComponent (not at pipe Initialize) because Super
// Mario 3D Land waits on the frame semaphore without ever writing a pipe, and
// the semaphore event fires every frame from boot for the same reason.

import (
	"fmt"
	"math"
)

// The DSP RAM window an application sees, and the two shared-memory regions
// inside it. ConvertProcessAddressFromDspDram maps a DSP word address to
// (addr<<1) + dspRAMBase + 0x40000; region 0 begins at DSP word 0x8000.
const (
	dspRAMBase = 0x1FF00000
	dspRAMSize = 0x80000

	dspRegion0    = 0x1FF50000
	dspRegion1    = 0x1FF70000
	dspSharedSize = 0x8000
)

// dspFrameTicks is one audio frame in ARM11 cycles: 160 samples per frame ×
// 4096 TeakLite cycles per sample × 2 ARM11 cycles per TeakLite cycle
// (Citra's hardware-verified ratio). Same nominal unit as stepsPerFrame, so
// ~3.4 audio frames elapse per VBlank — paced on Machine.instrs (monotonic),
// never CPU.Instrs (per-thread), and independent of the VBlank: the sound
// thread starts waiting before the game brings up GSP at all.
const dspFrameTicks = 160 * 4096 * 2

const dspNumSources = 24

// Byte offsets of the 15 structures within one shared-memory region, in the
// order they sit in memory. The audio pipe announces them as DSP words
// (0x8000 + offset/2). The trailing frame counter is the app's.
const (
	dspOffDSPStatus     = 0x0800 // u16 unknown, u16 dropped_frames, padding
	dspOffDebug         = 0x0820
	dspOffFinalSamples  = 0x0A80 // s16 pcm16[160][2]
	dspOffSourceStatus  = 0x0D00 // 24 × 12-byte SourceStatus
	dspOffCompressor    = 0x0E20
	dspOffDSPConfig     = 0x2860
	dspOffIntermediate  = 0x2924
	dspOffSourceConfigs = 0x3D24 // 24 × 192-byte SourceConfiguration
	dspOffAdpcmCoeffs   = 0x4F24
	dspOffUnknown10     = 0x5224
	dspOffUnknown11     = 0x5424
	dspOffUnknown12     = 0x55A4
	dspOffUnknown13     = 0x58A4
	dspOffUnknown14     = 0x58B8
	dspOffFrameCounter  = 0x7FFE // u16, incremented by the APP each frame
)

// DSP state-change values the app writes to the audio pipe (byte 0).
const (
	dspStateOff      = 0 // also the Initialize command value
	dspStateOn       = 1
	dspStateSleeping = 2
)

// Pipe channels and interrupt types (RegisterInterruptEvents arguments).
const (
	dspPipeDebug  = 0
	dspPipeDMA    = 1
	dspPipeAudio  = 2
	dspPipeBinary = 3

	dspIntZero = 0
	dspIntOne  = 1
	dspIntPipe = 2
)

// dspHLE is the whole DSP model's state. Fields are exported for the savestate
// gob encoder (state.go serialises the struct wholesale).
type dspHLE struct {
	ComponentLoaded bool   // LoadComponent seen; the Teak core is "booted" and the frame clock runs
	ComponentSize   uint32 // for reporting
	State           uint32 // dspStateOff/On/Sleeping — the audio pipe's state machine
	Pipes           [8][]byte
	IntEvents       map[uint32]uint32 // interrupt<<8|channel → the game's event handle
	SemEvent        uint32            // the event GetSemaphoreEventHandle minted (0 = not yet)
	Semaphore       uint32            // last SetSemaphore value (app → DSP; recorded, unused)
	SemMask         uint32            // SetSemaphoreMask value (recorded, unused)
	NextFrame       uint64            // m.instrs deadline of the next audio frame
	Ticks           uint64            // audio frames delivered (instrumentation)
	Sources         [dspNumSources]dspSource

	// The mixers (dsp_voice.go): the volume each of the three intermediate
	// mixes carries into the final mix, the two aux busses that route a mix out
	// through the application, and the output format.
	MixVolume    [3]float32
	AuxBusEnable [2]bool
	OutputFormat uint16
	ClippingMode uint16
	Headphones   bool
}

// --- service commands --------------------------------------------------------

// ipcDSP services dsp::DSP. Every command a title has been seen to issue is
// modelled; anything else halts loudly so the frontier stays explicit.
func (m *Machine) ipcDSP(hdr ipcHeader) bool {
	switch hdr.Command {
	case 0x0001: // RecvData(register) → u16: the DSP's reply register 0 reads 0
		// while the component runs and 1 once it is shut down / sleeping —
		// SM3DL's applet-resume path polls it until it reads 1 (retry loop
		// 0x001F9488; wrappers 0x001F2FF4/0x001F3244 read cmdbuf[2]).
		v := uint32(1)
		if m.dsp.State == dspStateOn {
			v = 0
		}
		m.ipcReply(hdr.Command, v)
		return true
	case 0x0002: // RecvDataIsReady(register) → u8: always ready
		m.ipcReply(hdr.Command, 1)
		return true
	case 0x0007: // SetSemaphore(u16) — the app raises the CPU→DSP semaphore
		// ("finished writing shared memory"). The HLE has no Teak core to
		// notice; record it.
		m.dsp.Semaphore = m.ipcArg(1) & 0xFFFF
		m.ipcReply(hdr.Command)
		return true
	case 0x000C: // ConvertProcessAddressFromDspDram(word addr) → process vaddr.
		// Captain Toad's wrapper 0x00126A14 (header 0x000C0040) converts each
		// structure address the audio pipe announced.
		m.ipcReply(hdr.Command, (m.ipcArg(1)<<1)+dspRAMBase+0x40000)
		return true
	case 0x000D: // WriteProcessPipe(channel, size, static buffer)
		return m.dspWritePipe(hdr)
	case 0x000E: // ReadPipe(channel, peer, size) — hardware HANGS if the pipe
		// holds less than asked; that would be a bug worth seeing, so halt.
		ch, size := m.ipcArg(1)&7, m.ipcArg(3)&0xFFFF
		if uint32(len(m.dsp.Pipes[ch])) < size {
			m.CPU.Halt("dsp ReadPipe channel %d wants 0x%X bytes, pipe holds 0x%X (hardware would hang) at 0x%08X",
				ch, size, len(m.dsp.Pipes[ch]), m.CPU.PC())
			return true
		}
		m.dspReplyPipeData(hdr, ch, size, false)
		return true
	case 0x000F: // GetPipeReadableSize(channel, peer) → u16
		m.ipcReply(hdr.Command, uint32(len(m.dsp.Pipes[m.ipcArg(1)&7])))
		return true
	case 0x0010: // ReadPipeIfPossible(channel, peer, size) → u16 read + data
		m.dspReplyPipeData(hdr, m.ipcArg(1)&7, m.ipcArg(3)&0xFFFF, true)
		return true
	case 0x0011: // LoadComponent(size, prog_mask, data_mask, mapped buffer) →
		// is_loaded + the buffer echoed back. The buffer is Nintendo's signed
		// Teak firmware ("DSP1" container after a 0x100-byte signature); the
		// HLE does not run it, but from here the DSP is booted: its frame
		// clock runs and the frame semaphore fires each audio frame. (SM3DL
		// depends on that: it waits on the semaphore event without ever
		// writing a pipe command.)
		m.dsp.ComponentLoaded = true
		m.dsp.ComponentSize = m.ipcArg(1)
		m.dsp.NextFrame = m.instrs + dspFrameTicks
		if m.Verbose {
			fmt.Printf("    dsp LoadComponent size=0x%X progMask=0x%X dataMask=0x%X\n",
				m.ipcArg(1), m.ipcArg(2), m.ipcArg(3))
		}
		desc, ptr := m.ipcArg(4), m.ipcArg(5)
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|2<<6|2)
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, 1) // is_loaded
		m.WriteWord(m.cmdBuf()+12, desc)
		m.WriteWord(m.cmdBuf()+16, ptr)
		return true
	case 0x0012: // UnloadComponent
		m.dsp.ComponentLoaded = false
		m.dsp.State = dspStateOff
		m.ipcReply(hdr.Command)
		return true
	case 0x0013, 0x0014: // FlushDataCache / InvalidateDataCache(addr, size, process)
		m.ipcReply(hdr.Command)
		return true
	case 0x0015: // RegisterInterruptEvents(interrupt, channel, event handle) —
		// the game hands over ITS OWN event (traced: the handle is cmdbuf[4]);
		// handle 0 unregisters. The audio-frame tick signals the pipe-2 event.
		intr, ch, ev := m.ipcArg(1), m.ipcArg(2), m.ipcArg(4)
		key := intr<<8 | ch
		if ev == 0 {
			delete(m.dsp.IntEvents, key)
		} else {
			m.dsp.IntEvents[key] = ev
		}
		if m.Verbose {
			fmt.Printf("    dsp RegisterInterruptEvents interrupt=%d channel=%d event=0x%08X\n", intr, ch, ev)
		}
		m.ipcReply(hdr.Command)
		return true
	case 0x0016: // GetSemaphoreEventHandle → the DSP→CPU frame semaphore's
		// event. Minted ONCE and real: the audio-frame tick signals it, which
		// is what finally makes a real handle here safe (see the writeup's
		// two failed shortcuts — a real handle nothing signals blocks the
		// game forever; a pulsed handle without a coherent DSP behind it
		// crash-restarts it).
		if m.dsp.SemEvent == 0 {
			m.dsp.SemEvent = m.newHandle("event", false) // auto-reset, like svcCreateEvent(OneShot)
			m.handles[m.dsp.SemEvent].name = "dsp-semaphore"
		}
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6|2)
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, 0) // translate descriptor: 1 handle
		m.WriteWord(m.cmdBuf()+12, m.dsp.SemEvent)
		return true
	case 0x0017: // SetSemaphoreMask(u16) — which DSP semaphore bits raise the
		// interrupt. The firmware raises its frame semaphore each tick; the
		// HLE signals the event unconditionally and records the mask.
		m.dsp.SemMask = m.ipcArg(1) & 0xFFFF
		m.ipcReply(hdr.Command)
		return true
	case 0x001F: // GetHeadphoneStatus → u8: not inserted
		m.ipcReply(hdr.Command, 0)
		return true
	case 0x0020: // ForceHeadphoneOut(u8)
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("dsp command 0x%04X unimplemented at 0x%08X after %d instructions",
		hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

// dspWritePipe handles WriteProcessPipe. Only the audio pipe's control
// protocol is modelled; a write to any other pipe halts loudly.
func (m *Machine) dspWritePipe(hdr ipcHeader) bool {
	ch, size, ptr := m.ipcArg(1), m.ipcArg(2), m.ipcArg(4)
	if ch != dspPipeAudio {
		m.CPU.Halt("dsp WriteProcessPipe to unmodelled pipe %d (size 0x%X) at 0x%08X after %d instructions",
			ch, size, m.CPU.PC(), m.CPU.Instrs)
		return true
	}
	if size < 4 {
		m.CPU.Halt("dsp WriteProcessPipe audio: %d-byte write, want 4 at 0x%08X", size, m.CPU.PC())
		return true
	}
	// Audio-pipe messages are 4 bytes; only byte 0 (the state change) is
	// meaningful — bytes 2..3 are stack garbage the real service zeroes.
	state := m.Read(ptr)
	if m.Verbose {
		fmt.Printf("    dsp WriteProcessPipe audio state-change=%d\n", state)
	}
	switch state {
	case 0, 2: // Initialize / Wakeup: reset the pipes, announce the shared-
		// memory structure addresses on the audio pipe, start running. (The
		// difference on hardware is whether input state survives; neither
		// title has needed it modelled.)
		m.dsp.Pipes = [8][]byte{}
		m.dspAnnounceStructs()
		m.dsp.State = dspStateOn
	case 1: // Shutdown
		m.dsp.State = dspStateOff
	case 3: // Sleep
		m.dspAnnounceStructs()
		m.dsp.State = dspStateSleeping
	default:
		m.CPU.Halt("dsp audio pipe: unknown state change %d at 0x%08X", state, m.CPU.PC())
		return true
	}
	m.ipcReply(hdr.Command)
	return true
}

// dspAnnounceStructs writes the audio pipe's answer to Initialize/Wakeup: a
// u16 count then the 15 shared-memory structure addresses as DSP words
// (0x8000 + region-relative byte offset / 2), in the pipe's fixed order —
// NOT their order in memory. Raises the pipe-2 interrupt: data is available.
func (m *Machine) dspAnnounceStructs() {
	offs := [15]uint32{
		dspOffFrameCounter, dspOffSourceConfigs, dspOffSourceStatus,
		dspOffAdpcmCoeffs, dspOffDSPConfig, dspOffDSPStatus,
		dspOffFinalSamples, dspOffIntermediate, dspOffCompressor,
		dspOffDebug, dspOffUnknown10, dspOffUnknown11, dspOffUnknown12,
		dspOffUnknown13, dspOffUnknown14,
	}
	p := make([]byte, 0, 2+2*len(offs))
	p = append(p, byte(len(offs)), 0)
	for _, off := range offs {
		w := 0x8000 + off/2
		p = append(p, byte(w), byte(w>>8))
	}
	m.dsp.Pipes[dspPipeAudio] = append(m.dsp.Pipes[dspPipeAudio], p...)
	m.dspSignalInterrupt(dspIntPipe, dspPipeAudio)
}

// dspReplyPipeData answers ReadPipe/ReadPipeIfPossible: consume up to size
// bytes from the pipe into the caller's static buffer 0 (declared in its TLS
// at +0x180, like the APT parameter payloads) and reply with the byte count.
// ifPossible determines the response shape (2 result words vs 1).
func (m *Machine) dspReplyPipeData(hdr ipcHeader, ch, size uint32, ifPossible bool) {
	var out []byte
	if data := m.dsp.Pipes[ch]; uint32(len(data)) >= size {
		out = data[:size]
		m.dsp.Pipes[ch] = data[size:]
	}
	ptr := m.ReadWord(m.curThread.tlsBase + 0x184)
	if max := m.ReadWord(m.curThread.tlsBase+0x180) >> 14; uint32(len(out)) > max {
		out = out[:max]
	}
	for i, b := range out {
		m.Write(ptr+uint32(i), b)
	}
	n := uint32(len(out))
	if ifPossible {
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|2<<6|2)
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, n)
		m.WriteWord(m.cmdBuf()+12, n<<14|2)
		m.WriteWord(m.cmdBuf()+16, ptr)
		return
	}
	m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6|2)
	m.WriteWord(m.cmdBuf()+4, resultSuccess)
	m.WriteWord(m.cmdBuf()+8, n<<14|2)
	m.WriteWord(m.cmdBuf()+12, ptr)
}

// --- the audio-frame tick ----------------------------------------------------

// dspDue reports whether the next audio frame's deadline has passed. The clock
// runs from LoadComponent (the Teak core is booted even before the audio
// pipeline is initialised).
func (m *Machine) dspDue() bool {
	return m.dsp.ComponentLoaded && m.instrs >= m.dsp.NextFrame
}

// dspDeadline reports the pending audio-frame deadline for the run loop's
// idle fast-forward.
func (m *Machine) dspDeadline() (uint64, bool) {
	if !m.dsp.ComponentLoaded {
		return 0, false
	}
	return m.dsp.NextFrame, true
}

// dspTick is one audio frame. While the pipeline runs (state On) it performs
// the DSP's half of the shared-memory exchange — consume configurations from
// the region the app last completed, advance the sources, publish statuses to
// the other region — and raises the audio-pipe interrupt. The frame semaphore
// fires every frame from component boot, pipeline running or not.
func (m *Machine) dspTick() {
	defer m.profEnd(bucketDSP, m.profStart()) // profile.go
	m.dsp.NextFrame = m.instrs + dspFrameTicks
	m.dsp.Ticks++
	if m.dsp.State == dspStateOn {
		read, write := m.dspReadRegion(), m.dspWriteRegion()
		var mixes [3]dspQuadFrame
		for i := 0; i < dspNumSources; i++ {
			m.dspParseConfig(i, read)
			if m.dsp.Sources[i].Enabled {
				m.dspSourceFrame(i)
			}
			m.dspWriteStatus(i, write)
			for mix := 0; mix < 3; mix++ {
				m.dsp.Sources[i].mixInto(&mixes[mix], mix)
			}
		}
		m.dspMixerConfig(read)
		final := m.dspMix(read, write, &mixes)
		m.dspWriteFinal(write, &final)
		m.WriteWord(write+dspOffDSPStatus, 0) // DspStatus: unknown, dropped_frames

		// The audio frame IS the pipe-2 interrupt: that is the event the app's
		// sound thread blocks on, and it is the one it hands to
		// RegisterInterruptEvents. Captain Toad makes that unmissable — it
		// registers an event, unregisters it, then registers a *different* one
		// (0x54) and waits on exactly that handle forever. Do not "tidy" this
		// signal away: without it the sound thread never wakes, and because that
		// same thread is the engine's resource-loader producer, the whole scene
		// build deadlocks behind it (its stage-init worker waits on a job whose
		// worker waits on a load that never arrives).
		m.dspSignalInterrupt(dspIntPipe, dspPipeAudio)
	}
	// The semaphore event fires every frame from component boot: Super Mario 3D
	// Land waits on it without ever writing a pipe command. (Scoping it to the
	// pipeline-off phase was tried and changes nothing — the sound thread blocks
	// on the pipe interrupt, not on this — so it stays as the simpler rule.)
	if m.dsp.SemEvent != 0 {
		m.dspSignalHandle(m.dsp.SemEvent)
	}
}

// dspReadRegion / dspWriteRegion implement the double-buffer handoff: the
// region with the higher frame counter (with 0xFFFF wraparound handling) is
// the one the app finished writing — the DSP reads it and answers into the
// other.
func (m *Machine) dspReadRegion() uint32 {
	c0 := m.dspRead16(dspRegion0 + dspOffFrameCounter)
	c1 := m.dspRead16(dspRegion1 + dspOffFrameCounter)
	switch {
	case c0 == 0xFFFF && c1 != 0xFFFE:
		return dspRegion1
	case c1 == 0xFFFF && c0 != 0xFFFE:
		return dspRegion0
	case c0 > c1:
		return dspRegion0
	}
	return dspRegion1
}

func (m *Machine) dspWriteRegion() uint32 {
	if m.dspReadRegion() == dspRegion0 {
		return dspRegion1
	}
	return dspRegion0
}

// SourceConfiguration byte offsets within one 192-byte entry, and the dirty
// bits (the DSP clears the whole dirty word each frame after acting on it).
const (
	srcCfgDirty        = 0x00 // u32 dirty flags
	srcCfgGain         = 0x04 // f32[3][4]: per intermediate mixer, per channel
	srcCfgRate         = 0x34 // f32 rate multiplier
	srcCfgInterp       = 0x38 // u8 interpolation mode
	srcCfgFiltersOn    = 0x3A // u16: bit0 simple, bit1 biquad
	srcCfgSimpleFilter = 0x3C // s16 b0, s16 a1 (s1.15)
	srcCfgBiquadFilter = 0x40 // s16 a2, a1, b2, b1, b0 (s2.14)
	srcCfgBuffersDirty = 0x4A // u16 bitmap over the 4 queued buffers
	srcCfgBuffers      = 0x4C // 4 × 20-byte Buffer
	srcCfgEnable       = 0xA0 // u8
	srcCfgSyncCount    = 0xA2 // u16
	srcCfgPlayPosition = 0xA4 // u32-dsp (embedded buffer start sample)
	srcCfgEmbPhysAddr  = 0xAC // u32-dsp
	srcCfgEmbLength    = 0xB0 // u32-dsp, IN SAMPLES
	srcCfgFlags1       = 0xB4 // u16: bits0-1 channels, bits2-3 codec, bit5 fade-in
	srcCfgEmbAdpcmPS   = 0xB6 // u16: predictor (bits 4-7) and scale (bits 0-3)
	srcCfgEmbAdpcmYn   = 0xB8 // s16[2]: y[n-1], y[n-2]
	srcCfgFlags2       = 0xBC // u16: bit0 adpcm_dirty, bit1 is_looping
	srcCfgEmbBufferID  = 0xBE // u16

	srcBufPhysAddr   = 0x00 // u32-dsp
	srcBufLength     = 0x04 // u32-dsp, IN SAMPLES
	srcBufAdpcmPS    = 0x08 // u16
	srcBufAdpcmYn    = 0x0A // s16[2]
	srcBufAdpcmDirty = 0x0E // u8
	srcBufIsLooping  = 0x0F // u8
	srcBufBufferID   = 0x10 // u16

	dirtyFormat          = 1 << 0
	dirtyMonoStereo      = 1 << 1
	dirtyAdpcmCoeffs     = 1 << 2
	dirtyPartialEmbedded = 1 << 3
	dirtyPartialReset    = 1 << 4
	dirtyEnable          = 1 << 16
	dirtyInterp          = 1 << 17
	dirtyRate            = 1 << 18
	dirtyBufferQueue     = 1 << 19
	dirtyPlayPosition    = 1 << 21
	dirtyFiltersEnabled  = 1 << 22
	dirtySimpleFilter    = 1 << 23
	dirtyBiquadFilter    = 1 << 24
	dirtyGain0           = 1 << 25
	dirtyGain1           = 1 << 26
	dirtyGain2           = 1 << 27
	dirtySyncCount       = 1 << 28
	dirtyReset           = 1 << 29
	dirtyEmbeddedBuffer  = 1 << 30
)

// dspParseConfig consumes source i's configuration from the read region:
// apply what the control model tracks, clear ALL the dirty flags (the app's
// dirty-flag protocol depends on the DSP doing so — it ORs new flags in and
// waits for them to vanish).
func (m *Machine) dspParseConfig(i int, region uint32) {
	cfg := region + dspOffSourceConfigs + uint32(i)*192
	dirty := m.ReadWord(cfg + srcCfgDirty)
	if dirty == 0 {
		return
	}
	s := &m.dsp.Sources[i]
	if m.DSPTrace {
		fmt.Printf("    dsp[%d] cfg r%d(c=%d/%d) dirty=%08X enable=%d sync=%d flags1=%04X len=%d id=%d buffers=%04X (frame %d)\n",
			i, (region-dspRegion0)/(dspRegion1-dspRegion0),
			m.dspRead16(dspRegion0+dspOffFrameCounter), m.dspRead16(dspRegion1+dspOffFrameCounter),
			dirty, m.Read(cfg+srcCfgEnable), m.dspRead16(cfg+srcCfgSyncCount),
			m.dspRead16(cfg+srcCfgFlags1), m.dspRead32(cfg+srcCfgEmbLength),
			m.dspRead16(cfg+srcCfgEmbBufferID), m.dspRead16(cfg+srcCfgBuffersDirty), m.dsp.Ticks)
	}
	if dirty&dirtyReset != 0 {
		*s = dspSource{Rate: 1}
		s.Filters.reset()
	}
	if dirty&dirtyPartialReset != 0 {
		s.Queue = nil
	}
	if dirty&dirtyEnable != 0 {
		s.Enabled = m.Read(cfg+srcCfgEnable) != 0
	}
	// sync_count latches on its dirty bit, like every other field: the app writes
	// each region with only the fields it has changed, so a region's sync word can
	// belong to an older generation than its frame counter suggests.
	if dirty&dirtySyncCount != 0 {
		s.SyncCount = m.dspRead16(cfg + srcCfgSyncCount)
	}

	if dirty&dirtyRate != 0 {
		s.Rate = math.Float32frombits(m.ReadWord(cfg + srcCfgRate))
		if !(s.Rate > 0) { // zero, negative or NaN: the firmware degrades; keep 1:1
			s.Rate = 1
		}
	}
	if dirty&dirtyInterp != 0 {
		s.Interp = m.Read(cfg + srcCfgInterp)
	}
	// Format and channel count latch on their own dirty bits AND with a new
	// embedded buffer: the app declares them alongside it.
	if dirty&(dirtyFormat|dirtyEmbeddedBuffer) != 0 {
		s.Format = uint8(m.dspRead16(cfg+srcCfgFlags1) >> 2 & 3)
	}
	if dirty&(dirtyMonoStereo|dirtyEmbeddedBuffer) != 0 {
		s.Stereo = m.dspRead16(cfg+srcCfgFlags1)&3 == 2
	}
	if dirty&dirtyAdpcmCoeffs != 0 {
		co := region + dspOffAdpcmCoeffs + uint32(i)*32
		for k := 0; k < 16; k++ {
			s.AdpcmCoeffs[k] = int16(m.dspRead16(co + uint32(k)*2))
		}
	}
	for g := 0; g < 3; g++ {
		if dirty&(dirtyGain0<<g) == 0 {
			continue
		}
		for c := 0; c < 4; c++ {
			s.Gain[g][c] = dspFloat(m.ReadWord(cfg + srcCfgGain + uint32(g*4+c)*4))
		}
	}
	if dirty&dirtyFiltersEnabled != 0 {
		en := m.dspRead16(cfg + srcCfgFiltersOn)
		s.Filters.enable(en&1 != 0, en&2 != 0)
	}
	if dirty&dirtySimpleFilter != 0 {
		s.Filters.SB0 = int32(int16(m.dspRead16(cfg + srcCfgSimpleFilter)))
		s.Filters.SA1 = int32(int16(m.dspRead16(cfg + srcCfgSimpleFilter + 2)))
	}
	if dirty&dirtyBiquadFilter != 0 {
		s.Filters.BA2 = int16(m.dspRead16(cfg + srcCfgBiquadFilter))
		s.Filters.BA1 = int16(m.dspRead16(cfg + srcCfgBiquadFilter + 2))
		s.Filters.BB2 = int16(m.dspRead16(cfg + srcCfgBiquadFilter + 4))
		s.Filters.BB1 = int16(m.dspRead16(cfg + srcCfgBiquadFilter + 6))
		s.Filters.BB0 = int16(m.dspRead16(cfg + srcCfgBiquadFilter + 8))
	}
	// The play position applies to the embedded buffer's first playthrough
	// and defaults to 0 without its dirty bit.
	playPos := uint32(0)
	if dirty&dirtyPlayPosition != 0 {
		playPos = m.dspRead32(cfg + srcCfgPlayPosition)
	}
	if dirty&dirtyPartialEmbedded != 0 && len(s.CurBuf) > 0 {
		// The app re-declared the length of the buffer currently playing — how a
		// looped stream is extended in place. Re-decode it at the new length from
		// the LATCHED address (the configuration's may already point elsewhere)
		// and re-consume what has already been played.
		buf := dspBuffer{
			PhysAddr: s.CurPhysAddr,
			Length:   m.dspRead32(cfg + srcCfgEmbLength),
			Stereo:   s.Stereo,
			Format:   s.Format,
		}
		if pcm := m.dspDecodeBuffer(s, buf); pcm != nil {
			if int(s.CurSample) < len(pcm) {
				s.CurBuf = pcm[s.CurSample:]
			} else {
				s.CurSample, s.CurBuf = 0, pcm
			}
		}
	}
	if dirty&dirtyEmbeddedBuffer != 0 {
		s.Queue = append(s.Queue, dspBuffer{
			PhysAddr:     m.dspRead32(cfg + srcCfgEmbPhysAddr),
			Length:       m.dspRead32(cfg + srcCfgEmbLength),
			AdpcmPS:      uint8(m.dspRead16(cfg + srcCfgEmbAdpcmPS)),
			AdpcmYn:      [2]int16{int16(m.dspRead16(cfg + srcCfgEmbAdpcmYn)), int16(m.dspRead16(cfg + srcCfgEmbAdpcmYn + 2))},
			AdpcmDirty:   m.dspRead16(cfg+srcCfgFlags2)&1 != 0,
			IsLooping:    m.dspRead16(cfg+srcCfgFlags2)&2 != 0,
			BufferID:     m.dspRead16(cfg + srcCfgEmbBufferID),
			Stereo:       s.Stereo,
			Format:       s.Format,
			PlayPosition: playPos,
		})
	}
	if dirty&dirtyBufferQueue != 0 {
		mask := m.dspRead16(cfg + srcCfgBuffersDirty)
		if m.DSPTrace {
			fmt.Printf("    dsp[%d] ENQUEUE r%d mask=%04X ids=[%d %d %d %d] (frame %d)\n", i,
				(region-dspRegion0)/(dspRegion1-dspRegion0), mask,
				m.dspRead16(cfg+srcCfgBuffers+0*20+srcBufBufferID), m.dspRead16(cfg+srcCfgBuffers+1*20+srcBufBufferID),
				m.dspRead16(cfg+srcCfgBuffers+2*20+srcBufBufferID), m.dspRead16(cfg+srcCfgBuffers+3*20+srcBufBufferID),
				m.dsp.Ticks)
		}
		for b := uint32(0); b < 4; b++ {
			if mask&(1<<b) == 0 {
				continue
			}
			bb := cfg + srcCfgBuffers + b*20
			s.Queue = append(s.Queue, dspBuffer{
				PhysAddr:   m.dspRead32(bb + srcBufPhysAddr),
				Length:     m.dspRead32(bb + srcBufLength),
				AdpcmPS:    uint8(m.dspRead16(bb + srcBufAdpcmPS)),
				AdpcmYn:    [2]int16{int16(m.dspRead16(bb + srcBufAdpcmYn)), int16(m.dspRead16(bb + srcBufAdpcmYn + 2))},
				AdpcmDirty: m.Read(bb+srcBufAdpcmDirty) != 0,
				IsLooping:  m.Read(bb+srcBufIsLooping) != 0,
				BufferID:   m.dspRead16(bb + srcBufBufferID),
				Stereo:     s.Stereo,
				Format:     s.Format,
				FromQueue:  true,
			})
		}
		m.dspWrite16(cfg+srcCfgBuffersDirty, 0)
	}
	m.WriteWord(cfg+srcCfgDirty, 0)
}

// dspWriteStatus publishes source i's status into the write region. The
// current_buffer_id / buffer-update handshake is what applications use to
// synchronise streaming (and audio with video), so its semantics follow the
// firmware's: the dirty byte reports a buffer change exactly once.
func (m *Machine) dspWriteStatus(i int, region uint32) {
	st := region + dspOffSourceStatus + uint32(i)*12
	s := &m.dsp.Sources[i]
	if m.DSPTrace && (s.Enabled || s.BufferUpdate || len(s.Queue) > 0) {
		fmt.Printf("    dsp[%d] status enabled=%v update=%v sync=%d pos=%d cur=%d last=%d queued=%d (frame %d)\n",
			i, s.Enabled, s.BufferUpdate, s.SyncCount, s.CurSample, s.CurBufferID, s.LastBufferID, len(s.Queue), m.dsp.Ticks)
	}
	enabled, update := byte(0), byte(0)
	if s.Enabled {
		enabled = 1
	}
	if s.BufferUpdate {
		update = 1
		s.BufferUpdate = false
	}
	m.Write(st+0, enabled)
	m.Write(st+1, update)
	m.dspWrite16(st+2, s.SyncCount)
	m.dspWrite32(st+4, s.CurSample)
	m.dspWrite16(st+8, s.CurBufferID)
	m.dspWrite16(st+10, s.LastBufferID)
}

// --- event plumbing and DSP-endian helpers -----------------------------------

// dspSignalInterrupt raises the event the game registered for (interrupt,
// channel), if any.
func (m *Machine) dspSignalInterrupt(interrupt, channel uint32) {
	if h, ok := m.dsp.IntEvents[interrupt<<8|channel]; ok {
		m.dspSignalHandle(h)
	}
}

func (m *Machine) dspSignalHandle(h uint32) {
	if obj := m.handles[h]; obj != nil {
		obj.signal = true
		if m.signalObject(obj) {
			m.reschedule = true
		}
	}
}

func (m *Machine) dspRead16(a uint32) uint16 {
	return uint16(m.Read(a)) | uint16(m.Read(a+1))<<8
}

func (m *Machine) dspWrite16(a uint32, v uint16) {
	m.Write(a, byte(v))
	m.Write(a+1, byte(v>>8))
}

// dspRead32/dspWrite32 handle the DSP's "middle-endian" 32-bit quantities: the
// 16-bit-native DSP stores u32s with the halfwords swapped relative to the
// ARM11's little-endian view (buffer addresses, lengths, play positions).
func (m *Machine) dspRead32(a uint32) uint32 {
	v := m.ReadWord(a)
	return v<<16 | v>>16
}

func (m *Machine) dspWrite32(a, v uint32) {
	m.WriteWord(a, v<<16|v>>16)
}
