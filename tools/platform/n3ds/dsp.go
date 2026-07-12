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
	Sources         [dspNumSources]dspSource
}

// dspSource is one of the 24 voices' control state (Citra's Source::state,
// minus everything that touches PCM).
type dspSource struct {
	Enabled      bool
	SyncCount    uint16
	Rate         float32 // rate multiplier (input samples per output sample)
	Queue        []dspBuffer
	HasCurrent   bool
	Pos          float64 // play position in the current buffer, in samples
	CurLength    uint32
	CurPhysAddr  uint32
	CurBufferID  uint16
	LastBufferID uint16
	BufferUpdate bool // current_buffer_id changed; reported once then cleared
}

// dspBuffer is one queued audio buffer's metadata (contents are never read).
type dspBuffer struct {
	PhysAddr     uint32
	Length       uint32 // in samples
	BufferID     uint16
	IsLooping    bool
	FromQueue    bool // came from the buffer queue (vs the embedded buffer)
	PlayPosition uint32
	HasPlayed    bool
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
	m.dsp.NextFrame = m.instrs + dspFrameTicks
	if m.dsp.State == dspStateOn {
		read, write := m.dspReadRegion(), m.dspWriteRegion()
		for i := 0; i < dspNumSources; i++ {
			m.dspParseConfig(i, read)
			s := &m.dsp.Sources[i]
			if s.Enabled {
				s.advanceFrame()
			}
			m.dspWriteStatus(i, write)
		}
		m.WriteWord(write+dspOffDSPStatus, 0) // DspStatus: unknown, dropped_frames
		m.dspSignalInterrupt(dspIntPipe, dspPipeAudio)
	}
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
	srcCfgRate         = 0x34 // f32 rate multiplier
	srcCfgBuffersDirty = 0x4A // u16 bitmap over the 4 queued buffers
	srcCfgBuffers      = 0x4C // 4 × 20-byte Buffer
	srcCfgEnable       = 0xA0 // u8
	srcCfgSyncCount    = 0xA2 // u16
	srcCfgPlayPosition = 0xA4 // u32-dsp (embedded buffer start sample)
	srcCfgEmbPhysAddr  = 0xAC // u32-dsp
	srcCfgEmbLength    = 0xB0 // u32-dsp, in samples
	srcCfgFlags2       = 0xBC // u16: bit0 adpcm_dirty, bit1 is_looping
	srcCfgEmbBufferID  = 0xBE // u16

	srcBufPhysAddr  = 0x00 // u32-dsp
	srcBufLength    = 0x04 // u32-dsp
	srcBufIsLooping = 0x0F // u8
	srcBufBufferID  = 0x10 // u16

	dirtyPartialEmbedded = 1 << 3
	dirtyPartialReset    = 1 << 4
	dirtyEnable          = 1 << 16
	dirtyRate            = 1 << 18
	dirtyBufferQueue     = 1 << 19
	dirtyPlayPosition    = 1 << 21
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
	if dirty&dirtyReset != 0 {
		*s = dspSource{Rate: 1}
	}
	if dirty&dirtyPartialReset != 0 {
		s.Queue = nil
	}
	if dirty&dirtyEnable != 0 {
		s.Enabled = m.Read(cfg+srcCfgEnable) != 0
	}
	if dirty&dirtySyncCount != 0 {
		s.SyncCount = m.dspRead16(cfg + srcCfgSyncCount)
	}
	if dirty&dirtyRate != 0 {
		s.Rate = math.Float32frombits(m.ReadWord(cfg + srcCfgRate))
		if !(s.Rate > 0) { // zero, negative or NaN: the firmware degrades; keep 1:1
			s.Rate = 1
		}
	}
	// The play position applies to the embedded buffer's first playthrough
	// and defaults to 0 without its dirty bit.
	playPos := uint32(0)
	if dirty&dirtyPlayPosition != 0 {
		playPos = m.dspRead32(cfg + srcCfgPlayPosition)
	}
	if dirty&dirtyPartialEmbedded != 0 && s.HasCurrent {
		// The app re-declared the length of the buffer currently playing
		// (how looped streams are extended in place).
		newLen := m.dspRead32(cfg + srcCfgEmbLength)
		if float64(newLen) < s.Pos {
			s.Pos = 0
		} else {
			s.CurLength = newLen
		}
	}
	if dirty&dirtyEmbeddedBuffer != 0 {
		s.Queue = append(s.Queue, dspBuffer{
			PhysAddr:     m.dspRead32(cfg + srcCfgEmbPhysAddr),
			Length:       m.dspRead32(cfg + srcCfgEmbLength),
			BufferID:     m.dspRead16(cfg + srcCfgEmbBufferID),
			IsLooping:    m.dspRead16(cfg+srcCfgFlags2)&2 != 0,
			PlayPosition: playPos,
		})
	}
	if dirty&dirtyBufferQueue != 0 {
		mask := m.dspRead16(cfg + srcCfgBuffersDirty)
		for b := uint32(0); b < 4; b++ {
			if mask&(1<<b) == 0 {
				continue
			}
			bb := cfg + srcCfgBuffers + b*20
			s.Queue = append(s.Queue, dspBuffer{
				PhysAddr:  m.dspRead32(bb + srcBufPhysAddr),
				Length:    m.dspRead32(bb + srcBufLength),
				BufferID:  m.dspRead16(bb + srcBufBufferID),
				IsLooping: m.Read(bb+srcBufIsLooping) != 0,
				FromQueue: true,
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
	m.dspWrite32(st+4, uint32(s.Pos))
	m.dspWrite16(st+8, s.CurBufferID)
	m.dspWrite16(st+10, s.LastBufferID)
}

// advanceFrame moves an enabled source one audio frame forward: 160 output
// samples' worth of input is consumed from the current buffer, dequeuing the
// next (lowest buffer id first) as buffers run out. A source whose queue runs
// dry disables itself and reports the final buffer id in last_buffer_id —
// the "everything finished" signal streaming code watches for.
func (s *dspSource) advanceFrame() {
	if !s.HasCurrent {
		if !s.dequeue() {
			s.Enabled = false
			s.BufferUpdate = true
			s.LastBufferID = s.CurBufferID
			s.CurBufferID = 0
		}
		return
	}
	rate := float64(s.Rate)
	if !(rate > 0) {
		rate = 1
	}
	need := 160.0 // output samples per frame
	for need > 0 {
		if !s.HasCurrent && !s.dequeue() {
			break
		}
		remain := float64(s.CurLength) - s.Pos
		if remain <= 0 {
			s.HasCurrent = false
			continue
		}
		take := need
		if avail := remain / rate; avail < take {
			take = avail
		}
		s.Pos += take * rate
		need -= take
		if s.Pos >= float64(s.CurLength) {
			s.HasCurrent = false
		}
	}
}

// dequeue pops the lowest-buffer-id entry into the current-buffer slot. A
// looping buffer re-queues itself (marked played, so it restarts at 0).
func (s *dspSource) dequeue() bool {
	if len(s.Queue) == 0 {
		return false
	}
	best := 0
	for i := range s.Queue {
		if s.Queue[i].BufferID < s.Queue[best].BufferID {
			best = i
		}
	}
	buf := s.Queue[best]
	s.Queue = append(s.Queue[:best], s.Queue[best+1:]...)

	start := uint32(0)
	if !buf.HasPlayed {
		start = buf.PlayPosition
	}
	s.Pos = float64(start)
	s.CurLength = buf.Length
	s.CurPhysAddr = buf.PhysAddr
	s.CurBufferID = buf.BufferID
	s.LastBufferID = 0
	s.BufferUpdate = buf.FromQueue && !buf.HasPlayed
	s.HasCurrent = s.Pos < float64(s.CurLength)

	if buf.IsLooping {
		buf.HasPlayed = true
		s.Queue = append(s.Queue, buf)
	}
	return true
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
