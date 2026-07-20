package threedo

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm60"
)

// audiofolio.go high-level-emulates the Portfolio audio folio's timer services —
// the audio clock and Cue items games pace their simulation with. The interface
// is from the SDK's audio.h (AUDIONODE=4, AUDIO_CUE_NODE=5, AUDIOSWI=0x40000,
// SignalAtTime = AUDIOSWI+13) and the folio's user-function vector table
// (audio_folio.c AudioUserFuncs: byte offset = 4×index, so GetAudioTime is
// index 42 → -0xA8, GetCueSignal 18 → -0x48); the behavior is reimplemented
// here, not copied.
//
// The audio clock ticks at 240 Hz (audio_timer.c TICKSPERSECOND) — 4 ticks per
// 60 Hz video field — driven from the HLE's virtual field clock (advanceVBlank).
// A game frame loop does: SignalAtTime(cue, GetAudioTime()+N) then
// WaitSignal(GetCueSignal(cue)) — the folio raises the cue's signal when the
// clock reaches the requested time, releasing one loop iteration per N ticks.
// With this stubbed the loop's time delta reads zero and the simulation
// freezes, which is exactly how NFS's in-race world stood still.
const (
	// Audio folio SWIs (AUDIOSWI = 0x40000, +n per audio.h). The movie's audio
	// subscriber drives a DSP sample-player graph through these.
	swiStartInstrument  = 0x40001 // StartInstrument(instrument, tags)
	swiStopInstrument   = 0x40003 // StopInstrument(instrument, tags)
	swiTestHack         = 0x40007 // TestHack(tags) — knob-name probe glue
	swiConnectInstr     = 0x40008 // ConnectInstruments(src, srcName, dst, dstName)
	swiSignalAtTime     = 0x4000D // SignalAtTime(cue, time)
	swiSetAudioRate     = 0x4000F // SetAudioRate(owner, frac16 rate)
	swiSetAudioDuration = 0x40010 // SetAudioDuration(owner, frames)
	swiTweakRawKnob     = 0x40011 // TweakRawKnob(knob, value)
	swiStartAttachment  = 0x40012 // StartAttachment(attachment, tags)
	swiStopAttachment   = 0x40014 // StopAttachment(attachment, tags)
	swiLinkAttachments  = 0x40015 // LinkAttachments(at1, at2)
	swiMonitorAttach    = 0x40016 // MonitorAttachment(attachment, cue, cueAt)
	swiSetAudioItemInfo = 0x4001B // SetAudioItemInfo(item, tags)
	swiAbortTimerCue    = 0x40021 // AbortTimerCue(cue)

	// Audio node item types = MKNODEID(AUDIONODE=4, subtype) = 0x0400|subtype.
	typeInsTemplate = 0x401 // AUDIO_TEMPLATE_NODE
	typeInstrument  = 0x402 // AUDIO_INSTRUMENT_NODE
	typeKnob        = 0x403 // AUDIO_KNOB_NODE
	typeSample      = 0x404 // AUDIO_SAMPLE_NODE
	typeAudioCue    = 0x405 // AUDIO_CUE_NODE
	typeAttachment  = 0x407 // AUDIO_ATTACHMENT_NODE

	audioTicksPerField = 4     // 240 Hz clock / 60 fields per second
	audioSampleRate    = 44100 // DSP sample rate the clock duration is quoted in
)

// audioEvent is a pending SignalAtTime request: raise the cue's signal when the
// audio clock reaches time.
type audioEvent struct {
	cue  int32
	time uint32
}

// audioFolioSWI services the audio folio's timer SWIs (folio 4 << 16). Returns
// false for audio SWIs not yet modelled so they surface as logged stubs.
func (m *Machine) audioFolioSWI(c *arm60.CPU, swi uint32) bool {
	switch swi {
	case swiSignalAtTime:
		// SignalAtTime(cue, time): raise the cue's signal on its owner when the
		// audio clock reaches time (AudioTime compares are wraparound-signed).
		cue := m.items[int32(c.Reg(0))]
		if cue == nil || cue.typ != typeAudioCue {
			c.SetReg(0, ^uint32(0)) // BADITEM
			return true
		}
		t := c.Reg(1)
		if int32(t-m.audioTime) <= 0 {
			m.sendSignal(cue.owner, cue.signal)
		} else {
			m.audioEvents = append(m.audioEvents, audioEvent{cue: cue.num, time: t})
		}
		c.SetReg(0, 0)
	case swiAbortTimerCue:
		// AbortTimerCue(cue): cancel this cue's pending SignalAtTime requests.
		cueNum := int32(c.Reg(0))
		kept := m.audioEvents[:0]
		for _, e := range m.audioEvents {
			if e.cue != cueNum {
				kept = append(kept, e)
			}
		}
		m.audioEvents = kept
		c.SetReg(0, 0)
	case swiSetAudioRate, swiSetAudioDuration:
		// Clock rate stays at the 240 Hz default; log if a game tries to change it.
		m.note(fmt.Sprintf("audio SetAudioRate/Duration(0x%X, 0x%X) ignored (clock stays 240 Hz)", c.Reg(0), c.Reg(1)))
		c.SetReg(0, 0)
	case swiMonitorAttach:
		// MonitorAttachment(attachment, cue, cueAt): arm a cue to fire when the
		// attachment's sample reaches cueAt (-2 = end of sample). Record which cue
		// belongs to which attachment so StartAttachment can raise it. The DSP that
		// would fire it on real playback is not modelled; we schedule it instead.
		if att := int32(c.Reg(0)); att != 0 {
			m.attachCue[att] = int32(c.Reg(1))
		}
		c.SetReg(0, 0)
	case swiStartAttachment:
		// StartAttachment(attachment): begin playing the attached sample. We produce
		// no audio, so schedule the cue MonitorAttachment armed on it to fire on the
		// next audio tick — modelling the sample reaching its (end) monitor point.
		// The subscriber wakes on that cue and replies the sample chunk's held stream
		// message, letting its stream buffer recycle to the data acquirer.
		m.completeAttachment(int32(c.Reg(0)))
		c.SetReg(0, 0)
	case swiLinkAttachments:
		// LinkAttachments(at1, at2): chain at2 to play after at1. On the DSP, one
		// StartAttachment on the chain head plays the whole linked list, firing each
		// segment's monitor cue in turn. We do not run the DSP, so treat the freshly
		// linked segment (at2) the same as a started one and schedule its cue — that
		// is how every sample chunk after the first in a slot gets completed.
		m.completeAttachment(int32(c.Reg(1)))
		c.SetReg(0, 0)
	case swiStopAttachment:
		// StopAttachment(attachment): drop any pending cue for it.
		if cueNum, ok := m.attachCue[int32(c.Reg(0))]; ok {
			kept := m.audioEvents[:0]
			for _, e := range m.audioEvents {
				if e.cue != cueNum {
					kept = append(kept, e)
				}
			}
			m.audioEvents = kept
		}
		c.SetReg(0, 0)
	case swiStartInstrument, swiStopInstrument, swiTestHack, swiConnectInstr,
		swiTweakRawKnob, swiSetAudioItemInfo:
		// The DSP-graph control calls: with no real DSP there is nothing to do, but
		// they must report success (0) — the subscriber aborts a slot on any negative
		// return, and the generic stub would leak a non-zero argument into r0.
		c.SetReg(0, 0)
	default:
		return false
	}
	return true
}

// completeAttachment schedules the cue a MonitorAttachment armed on this
// attachment to fire on the next audio tick, standing in for the DSP playing the
// attached sample to its monitored end. Fires nothing if the attachment has no
// monitored cue. Scheduling (rather than raising the signal here) defers the
// subscriber's completion handler until it is back in its WaitSignal loop.
func (m *Machine) completeAttachment(att int32) {
	cueNum, ok := m.attachCue[att]
	if !ok {
		return
	}
	cue := m.items[cueNum]
	if cue == nil || cue.typ != typeAudioCue {
		return
	}
	m.audioEvents = append(m.audioEvents, audioEvent{cue: cue.num, time: m.audioTime + 1})
}

// serviceAudioFolio dispatches a call into the audio folio's user-function
// vector table (LDR pc, [base, #-off]; off = 4×table index per audio_folio.c).
func (m *Machine) serviceAudioFolio(foff uint32) {
	c := m.CPU
	switch foff {
	// Item-creation user functions. The byte offset is 4×the folio call number in
	// AudioUserFuncs[] (audio_folio.c): LoadInsTemplate -1, AllocInstrument -2,
	// GrabKnob -4, MakeSample -14, AttachSample -36. Each returns a real Item; the
	// subscriber treats any non-negative result as success and, crucially, only
	// pumps a sample slot once its instrument (AllocInstrument) is a positive Item,
	// so these MUST create tracked items rather than the old zero stubs.
	case 0x04: // LoadInsTemplate(name, tags) -> instrument template
		m.SetResultAndReturn(uint32(m.createItem(typeInsTemplate, 0, 0).num))
	case 0x08: // AllocInstrument(insTemplate, priority) -> instrument
		m.SetResultAndReturn(uint32(m.createItem(typeInstrument, 0, 0).num))
	case 0x10: // GrabKnob(instrument, knobName) -> knob
		m.SetResultAndReturn(uint32(m.createItem(typeKnob, 0, 0).num))
	case 0x38: // MakeSample(numBytes, tags) -> sample
		m.SetResultAndReturn(uint32(m.createItem(typeSample, 0, 0).num))
	case 0x90: // AttachSample(instrument, sample, fifoName) -> attachment
		m.SetResultAndReturn(uint32(m.createItem(typeAttachment, 0, 0).num))
	case 0x14: // SleepAudioTicks(ticks) -> a one-shot startup delay; do not block
		m.SetResultAndReturn(0)
	case 0xA8: // GetAudioTime() -> the 240 Hz audio clock
		m.SetResultAndReturn(m.audioTime)
	case 0x48: // GetCueSignal(cue) -> the signal SignalAtTime raises for it
		if it := m.items[int32(c.Reg(0))]; it != nil {
			m.SetResultAndReturn(it.signal)
		} else {
			m.SetResultAndReturn(0)
		}
	case 0x4C: // OwnAudioClock() -> ownership token Item
		if m.audioClockOwner == 0 {
			m.audioClockOwner = m.createItem(0x400, 0, 0).num
		}
		m.SetResultAndReturn(uint32(m.audioClockOwner))
	case 0x50: // DisownAudioClock(owner)
		m.SetResultAndReturn(0)
	case 0x3C: // GetAudioRate() -> frac16 ticks per second
		m.SetResultAndReturn(240 << 16)
	case 0x40: // GetAudioDuration() -> DSP frames per tick
		m.SetResultAndReturn(audioSampleRate / 240)
	case 0x44: // SleepUntilTime(cue, time): block the caller until the clock arrives
		cue := m.items[int32(c.Reg(0))]
		t := c.Reg(1)
		m.SetResultAndReturn(0)
		if cue == nil || cue.typ != typeAudioCue || int32(t-m.audioTime) <= 0 {
			return
		}
		m.audioEvents = append(m.audioEvents, audioEvent{cue: cue.num, time: t})
		task := m.curTask()
		task.wait = cue.signal
		task.state = stWaiting
		m.needSchedule = true
	default:
		m.note(fmt.Sprintf("AudioFolio[-0x%X] stub (r0=0x%08X r1=0x%08X r2=0x%08X)", foff, c.Reg(0), c.Reg(1), c.Reg(2)))
		m.SetResultAndReturn(0)
	}
}

// advanceAudioClock moves the 240 Hz audio clock forward by n video fields and
// fires the SignalAtTime requests that came due, raising each cue's signal on
// its owner. Called from advanceVBlank so audio time tracks field time.
func (m *Machine) advanceAudioClock(fields uint32) {
	m.audioTime += fields * audioTicksPerField
	if len(m.audioEvents) == 0 {
		return
	}
	kept := m.audioEvents[:0]
	for _, e := range m.audioEvents {
		if int32(m.audioTime-e.time) >= 0 {
			if cue := m.items[e.cue]; cue != nil {
				m.sendSignal(cue.owner, cue.signal)
			}
		} else {
			kept = append(kept, e)
		}
	}
	m.audioEvents = kept
}
