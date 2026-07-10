// Package audio decodes the Nintendo 64 libultra audio formats: the ALBankFile
// instrument bank (".CTL"), its VADPCM sample table (".TBL"), and the sequence
// bank ("S1") of Type-0 "compressed MIDI" songs. It is platform code, reused by
// any N64 title; the game-specific wiring (which cartridge offsets hold which
// resource) lives in the game's extract tree.
//
// The structures follow libultra's audio library (ALBankFile / ALBank /
// ALInstrument / ALSound / ALKeyMap / ALEnvelope / ALWaveTable / ALADPCMBook).
// Every pointer field in a ".CTL" is a byte offset from the start of the file,
// fixed up to a real pointer at load time by the running game; here they stay
// offsets and we chase them by hand.
package audio

import (
	"encoding/binary"
	"fmt"
)

// AL_BANK_VERSION, the "B1" magic every ALBankFile opens with.
const bankVersion = 0x4231

// Wave types in an ALWaveTable.
const (
	waveADPCM = 0 // AL_ADPCM_WAVE — VADPCM, decoded by vadpcm.go
	waveRaw16 = 1 // AL_RAW16_WAVE — big-endian s16 PCM
)

// Envelope holds an ALEnvelope: an attack/decay/release amplitude contour. Times
// are in microseconds; volumes are 0..127 amplitudes the synth ramps between.
type Envelope struct {
	AttackTime, DecayTime, ReleaseTime int32
	AttackVolume, DecayVolume          uint8
}

// KeyMap is an ALKeyMap: the note/velocity window a sound answers to, plus the
// key it was sampled at (keyBase) and a fine detune in cents.
type KeyMap struct {
	VelocityMin, VelocityMax uint8
	KeyMin, KeyMax, KeyBase  uint8
	Detune                   int8
}

// ADPCMBook is an ALADPCMBook: the per-sample predictor codebook used to expand
// VADPCM frames. Predictors holds Order*NPredictors*8 s16 coefficients.
type ADPCMBook struct {
	Order       int32
	NPredictors int32
	Predictors  []int16
}

// ADPCMLoop is an ALADPCMLoop: a sample loop with the decoder state to resume at
// the loop point.
type ADPCMLoop struct {
	Start, End, Count uint32
	State             [16]int16
}

// WaveTable is an ALWaveTable: one sample in the ".TBL", plus (for ADPCM) its
// predictor book and optional loop.
type WaveTable struct {
	Base int32 // byte offset into the ".TBL"
	Len  int32 // byte length in the ".TBL"
	Type uint8 // waveADPCM or waveRaw16
	Book *ADPCMBook
	Loop *ADPCMLoop
}

// Sound is an ALSound: an envelope + keymap + wavetable, with per-sound pan and
// volume trims.
type Sound struct {
	Env      *Envelope
	KeyMap   *KeyMap
	Wave     *WaveTable
	Pan, Vol uint8
	Flags    uint8
}

// Instrument is an ALInstrument: a set of key-mapped sounds sharing volume, pan
// and pitch-bend range. Program-change in a sequence selects one of these.
type Instrument struct {
	Volume, Pan, Priority, Flags uint8
	BendRange                    int16
	Sounds                       []*Sound
}

// SoundFor returns the sound whose keymap covers the given note and velocity, or
// nil if the instrument is silent there.
func (in *Instrument) SoundFor(key, vel uint8) *Sound {
	for _, s := range in.Sounds {
		k := s.KeyMap
		if k == nil {
			return s
		}
		if key >= k.KeyMin && key <= k.KeyMax && vel >= k.VelocityMin && vel <= k.VelocityMax {
			return s
		}
	}
	return nil
}

// Bank is an ALBank: the instrument array (plus an optional percussion program)
// and the sample rate its samples were recorded at.
type Bank struct {
	SampleRate  int32
	Percussion  *Instrument
	Instruments []*Instrument
}

// BankFile is a parsed ".CTL": one or more banks, all pointing into a shared
// ".TBL" sample table the caller supplies to the synth.
type BankFile struct {
	Banks []*Bank
	ctl   []byte // retained for validation/debug
}

type reader struct {
	b   []byte
	err error
}

func (r *reader) u8(o int32) uint8 {
	if r.err != nil || int(o) < 0 || int(o) >= len(r.b) {
		r.fail(o)
		return 0
	}
	return r.b[o]
}
func (r *reader) u16(o int32) uint16 {
	if r.err != nil || int(o) < 0 || int(o)+2 > len(r.b) {
		r.fail(o)
		return 0
	}
	return binary.BigEndian.Uint16(r.b[o:])
}
func (r *reader) u32(o int32) uint32 {
	if r.err != nil || int(o) < 0 || int(o)+4 > len(r.b) {
		r.fail(o)
		return 0
	}
	return binary.BigEndian.Uint32(r.b[o:])
}
func (r *reader) fail(o int32) {
	if r.err == nil {
		r.err = fmt.Errorf("audio: offset 0x%x out of range (len 0x%x)", o, len(r.b))
	}
}

// ParseBankFile decodes a ".CTL" instrument bank. It walks every pointer so a
// malformed bank fails here rather than at synth time.
func ParseBankFile(ctl []byte) (*BankFile, error) {
	r := &reader{b: ctl}
	if rev := r.u16(0); rev != bankVersion {
		return nil, fmt.Errorf("audio: bad ALBankFile revision 0x%04x (want 0x%04x)", rev, bankVersion)
	}
	bankCount := int32(r.u16(2))
	bf := &BankFile{ctl: ctl}
	for i := int32(0); i < bankCount; i++ {
		off := int32(r.u32(4 + i*4))
		if off == 0 {
			continue
		}
		bank, err := r.bank(off)
		if err != nil {
			return nil, fmt.Errorf("audio: bank %d: %w", i, err)
		}
		bf.Banks = append(bf.Banks, bank)
	}
	if r.err != nil {
		return nil, r.err
	}
	return bf, nil
}

// bank parses one ALBank at off. Layout: instCount u16, flags u8, pad u8,
// sampleRate s32, percussion offset, then instCount instrument offsets.
func (r *reader) bank(off int32) (*Bank, error) {
	instCount := int32(r.u16(off))
	b := &Bank{SampleRate: int32(r.u32(off + 4))}
	if perc := int32(r.u32(off + 8)); perc != 0 {
		b.Percussion = r.instrument(perc)
	}
	for i := int32(0); i < instCount; i++ {
		io := int32(r.u32(off + 12 + i*4))
		if io == 0 {
			b.Instruments = append(b.Instruments, nil)
			continue
		}
		b.Instruments = append(b.Instruments, r.instrument(io))
	}
	return b, r.err
}

// instrument parses an ALInstrument. Layout: volume, pan, priority, flags,
// tremType/Rate/Depth/Delay, vibType/Rate/Depth/Delay (8 bytes of LFO config we
// keep only structurally), bendRange s16, soundCount s16, sound offsets.
func (r *reader) instrument(off int32) *Instrument {
	in := &Instrument{
		Volume:    r.u8(off),
		Pan:       r.u8(off + 1),
		Priority:  r.u8(off + 2),
		Flags:     r.u8(off + 3),
		BendRange: int16(r.u16(off + 12)),
	}
	soundCount := int32(int16(r.u16(off + 14)))
	for i := int32(0); i < soundCount; i++ {
		so := int32(r.u32(off + 16 + i*4))
		if so == 0 {
			continue
		}
		in.Sounds = append(in.Sounds, r.sound(so))
	}
	return in
}

// sound parses an ALSound: envelope offset, keymap offset, wavetable offset,
// then samplePan, sampleVolume, flags.
func (r *reader) sound(off int32) *Sound {
	s := &Sound{
		Pan:   r.u8(off + 12),
		Vol:   r.u8(off + 13),
		Flags: r.u8(off + 14),
	}
	if e := int32(r.u32(off)); e != 0 {
		s.Env = r.envelope(e)
	}
	if k := int32(r.u32(off + 4)); k != 0 {
		s.KeyMap = r.keymap(k)
	}
	if w := int32(r.u32(off + 8)); w != 0 {
		s.Wave = r.wavetable(w)
	}
	return s
}

func (r *reader) envelope(off int32) *Envelope {
	return &Envelope{
		AttackTime:   int32(r.u32(off)),
		DecayTime:    int32(r.u32(off + 4)),
		ReleaseTime:  int32(r.u32(off + 8)),
		AttackVolume: r.u8(off + 12),
		DecayVolume:  r.u8(off + 13),
	}
}

func (r *reader) keymap(off int32) *KeyMap {
	return &KeyMap{
		VelocityMin: r.u8(off),
		VelocityMax: r.u8(off + 1),
		KeyMin:      r.u8(off + 2),
		KeyMax:      r.u8(off + 3),
		KeyBase:     r.u8(off + 4),
		Detune:      int8(r.u8(off + 5)),
	}
}

// wavetable parses an ALWaveTable. Layout: base s32, len s32, type u8, flags u8,
// 2 pad, then the ALADPCMWaveInfo {loop offset, book offset} for ADPCM waves.
func (r *reader) wavetable(off int32) *WaveTable {
	w := &WaveTable{
		Base: int32(r.u32(off)),
		Len:  int32(r.u32(off + 4)),
		Type: r.u8(off + 8),
	}
	if w.Type == waveADPCM {
		if loop := int32(r.u32(off + 12)); loop != 0 {
			w.Loop = r.loop(loop)
		}
		if book := int32(r.u32(off + 16)); book != 0 {
			w.Book = r.book(book)
		}
	}
	return w
}

func (r *reader) book(off int32) *ADPCMBook {
	bk := &ADPCMBook{Order: int32(r.u32(off)), NPredictors: int32(r.u32(off + 4))}
	n := bk.Order * bk.NPredictors * 8
	if n < 0 || n > 4096 {
		r.err = fmt.Errorf("audio: implausible ADPCM book order=%d npred=%d", bk.Order, bk.NPredictors)
		return bk
	}
	bk.Predictors = make([]int16, n)
	for i := int32(0); i < n; i++ {
		bk.Predictors[i] = int16(r.u16(off + 8 + i*2))
	}
	return bk
}

func (r *reader) loop(off int32) *ADPCMLoop {
	lp := &ADPCMLoop{
		Start: r.u32(off),
		End:   r.u32(off + 4),
		Count: r.u32(off + 8),
	}
	for i := int32(0); i < 16; i++ {
		lp.State[i] = int16(r.u16(off + 12 + i*2))
	}
	return lp
}
