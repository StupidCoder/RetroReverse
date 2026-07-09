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
	swiSignalAtTime     = 0x4000D // SignalAtTime(cue, time)
	swiSetAudioRate     = 0x4000F // SetAudioRate(owner, frac16 rate)
	swiSetAudioDuration = 0x40010 // SetAudioDuration(owner, frames)
	swiAbortTimerCue    = 0x40021 // AbortTimerCue(cue)

	typeAudioCue = 0x405 // MKNODEID(AUDIONODE, AUDIO_CUE_NODE)

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
	default:
		return false
	}
	return true
}

// serviceAudioFolio dispatches a call into the audio folio's user-function
// vector table (LDR pc, [base, #-off]; off = 4×table index per audio_folio.c).
func (m *Machine) serviceAudioFolio(foff uint32) {
	c := m.CPU
	switch foff {
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
