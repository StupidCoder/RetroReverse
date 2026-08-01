package gbamachine

// The AGB sound hardware: the Game Boy's four PSG channels carried forward
// unchanged in spirit, plus the GBA's own addition — two **Direct Sound** PCM
// channels fed by DMA and clocked by the timers.
//
// The two halves are modelled differently because the hardware works
// differently, and the difference is the interesting part of this file:
//
//   - the PSG channels are SYNTHESISED. They are a description of a waveform
//     (frequency, duty, envelope, sweep), so the model keeps a phase
//     accumulator per channel and evaluates it per output sample, exactly as
//     tools/platform/gameboy's APU does for the same four channels.
//   - the Direct Sound channels are TRANSPORTED. They carry no description at
//     all: the game's sound driver mixes its own PCM into a buffer in RAM, a
//     DMA channel in "special" timing refills a 32-byte FIFO, and a timer
//     overflow pops one signed 8-bit sample out of it. There is nothing to
//     synthesise — the model's whole job is to move the game's own bytes at
//     the rate the game's own timer asks for.
//
// That second path is why a GBA game's music can be a streamed mixdown rather
// than a chiptune, and it is also why sound here is inseparable from the DMA
// and timer models: get the timer rate wrong and the game's music plays at the
// wrong pitch while every register in the sound block reads back correctly.
//
// TIMING HONESTY: the scheduler ticks timers once per scanline (run.go), so a
// FIFO pop is placed with scanline granularity (~16 µs) rather than at its
// exact cycle. The RATE is right — the number of pops per second is exactly
// the timer's overflow rate — so pitch and duration are correct and the error
// is a sub-scanline jitter in when each sample lands. The mixer resamples from
// the popped values to the output rate.

import (
	"encoding/binary"
	"fmt"
	"os"
)

// The output rate. 32768 Hz is the rate GBA sound drivers themselves most
// commonly ask their timers for, so the common case needs no resampling.
const audioRate = 32768

// cpuHz is the ARM7TDMI's clock — what the timers and the FIFO rate derive from.
const cpuHz = 16777216.0

// dutyTable holds the four square-wave duty patterns (8 steps each).
var dutyTable = [4][8]byte{
	{0, 0, 0, 0, 0, 0, 0, 1}, // 12.5%
	{1, 0, 0, 0, 0, 0, 0, 1}, // 25%
	{1, 0, 0, 0, 0, 1, 1, 1}, // 50%
	{0, 1, 1, 1, 1, 1, 1, 0}, // 75%
}

var noiseDivisors = [8]int{8, 16, 32, 48, 64, 80, 96, 112}

// square is PSG channel 1 or 2 (channel 1 also has the frequency sweep).
type square struct {
	enabled   bool
	hasSweep  bool
	phase     float64
	dutySel   int
	freq      int // 11-bit
	length    int
	lengthEn  bool
	vol       int
	envDir    int
	envPeriod int
	envTimer  int
	swPeriod  int
	swShift   int
	swDir     int
	swTimer   int
	swOn      bool
	swShadow  int
}

// waveCh is PSG channel 3. The GBA's version has TWO banks of 32 4-bit samples
// (the Game Boy had one), selectable and optionally played back to back.
type waveCh struct {
	enabled  bool
	dacOn    bool
	phase    float64
	freq     int
	length   int
	lengthEn bool
	volShift int  // 0=mute, 1=full, 2=half, 3=quarter, as a right-shift
	force75  bool // SOUND3CNT_H bit 15: 75% volume, overriding volShift
	ram      [2][16]byte
	bank     int  // which bank playback reads
	twoBanks bool // play both banks as one 64-sample wave
}

// noiseCh is PSG channel 4.
type noiseCh struct {
	enabled   bool
	lfsr      uint16
	timer     float64
	length    int
	lengthEn  bool
	vol       int
	envDir    int
	envPeriod int
	envTimer  int
	divisor   int
	shift     int
	width7    bool
}

// fifo is one Direct Sound channel: a 32-byte queue the DMA refills and a timer
// drains, plus the sample currently on its output.
type fifo struct {
	q   []int8
	cur int8
}

func (f *fifo) push(b byte) {
	if len(f.q) < 32 {
		f.q = append(f.q, int8(b))
	}
}

// pop advances the channel by one sample. An empty FIFO holds its last value —
// on hardware the DAC simply keeps its level, which is what an underrun sounds
// like (a click or a stall, not silence).
func (f *fifo) pop() {
	if len(f.q) > 0 {
		f.cur = f.q[0]
		f.q = f.q[1:]
	}
}

func (f *fifo) reset() { f.q, f.cur = f.q[:0], 0 }

// apu is the whole sound block.
type apu struct {
	ch1, ch2 square
	ch3      waveCh
	ch4      noiseCh
	dsA, dsB fifo

	powered  bool
	frameAcc float64
	frameSeq int

	// The two mixing registers, mirrored here so the mixer does not reach back
	// into the machine's I/O file for a value it reads once per output sample.
	mixRegL uint16 // SOUNDCNT_L: PSG master volume and per-channel panning
	mixRegH uint16 // SOUNDCNT_H: PSG/Direct Sound levels, DS panning and timer select

	// Output accumulation: fractional sample position carried across scanlines.
	sampleAcc float64

	// PCM is the captured stereo mix, interleaved L,R. Only filled while
	// Capture is set — a run nobody asked for audio from should not pay for it.
	PCM     []int16
	Capture bool
}

func newAPU() *apu {
	a := &apu{}
	a.ch1.hasSweep = true
	a.ch4.lfsr = 0x7FFF
	return a
}

// --- register writes ---------------------------------------------------------

// soundWrite services a store to the sound register block (0x060-0x0A7) and the
// wave RAM (0x090-0x09F). Returns false if the address is not a sound register.
func (m *Machine) soundWrite(reg uint32, v uint16) bool {
	a := m.apu
	switch {
	case reg == 0x060: // SOUND1CNT_L — sweep
		a.ch1.swPeriod = int(v >> 4 & 7)
		a.ch1.swShift = int(v & 7)
		a.ch1.swDir = 1
		if v&(1<<3) != 0 {
			a.ch1.swDir = -1
		}
	case reg == 0x062: // SOUND1CNT_H — duty/length/envelope
		a.writeDutyEnv(&a.ch1, v)
	case reg == 0x064: // SOUND1CNT_X — frequency/control
		a.writeFreqCtl(&a.ch1, v)
	case reg == 0x068: // SOUND2CNT_L
		a.writeDutyEnv(&a.ch2, v)
	case reg == 0x06C: // SOUND2CNT_H
		a.writeFreqCtl(&a.ch2, v)

	case reg == 0x070: // SOUND3CNT_L — wave bank/DAC
		a.ch3.twoBanks = v&(1<<5) != 0
		a.ch3.bank = int(v >> 6 & 1)
		a.ch3.dacOn = v&(1<<7) != 0
		if !a.ch3.dacOn {
			a.ch3.enabled = false
		}
	case reg == 0x072: // SOUND3CNT_H — length/volume
		a.ch3.length = 256 - int(v&0xFF)
		a.ch3.volShift = int(v >> 13 & 3)
		a.ch3.force75 = v&(1<<15) != 0
	case reg == 0x074: // SOUND3CNT_X — frequency/control
		a.ch3.freq = int(v & 0x7FF)
		a.ch3.lengthEn = v&(1<<14) != 0
		if v&(1<<15) != 0 && a.ch3.dacOn {
			a.ch3.enabled = true
			a.ch3.phase = 0
			if a.ch3.length == 0 {
				a.ch3.length = 256
			}
		}

	case reg == 0x078: // SOUND4CNT_L — length/envelope
		a.ch4.length = 64 - int(v&0x3F)
		a.ch4.envPeriod = int(v >> 8 & 7)
		a.ch4.vol = int(v >> 12 & 0xF)
		a.ch4.envDir = -1
		if v&(1<<11) != 0 {
			a.ch4.envDir = 1
		}
		if v&0xF800 == 0 {
			a.ch4.enabled = false // DAC off
		}
	case reg == 0x07C: // SOUND4CNT_H — frequency/control
		a.ch4.divisor = noiseDivisors[v&7]
		a.ch4.width7 = v&(1<<3) != 0
		a.ch4.shift = int(v >> 4 & 0xF)
		a.ch4.lengthEn = v&(1<<14) != 0
		if v&(1<<15) != 0 {
			a.ch4.enabled = true
			a.ch4.lfsr = 0x7FFF
			a.ch4.envTimer = a.ch4.envPeriod
			if a.ch4.length == 0 {
				a.ch4.length = 64
			}
		}

	case reg == 0x080: // SOUNDCNT_L — PSG volume and panning
		a.mixRegL = v
	case reg == 0x082: // SOUNDCNT_H — PSG/Direct Sound mixing
		a.mixRegH = v
		// Bits 11 and 15 reset the Direct Sound FIFOs.
		if v&(1<<11) != 0 {
			a.dsA.reset()
		}
		if v&(1<<15) != 0 {
			a.dsB.reset()
		}
	case reg == 0x084: // SOUNDCNT_X — master power
		a.powered = v&(1<<7) != 0
		if !a.powered {
			a.ch1.enabled, a.ch2.enabled = false, false
			a.ch3.enabled, a.ch4.enabled = false, false
		}
	case reg == 0x088: // SOUNDBIAS — the output bias/resampling rate; storage only

	case reg >= 0x090 && reg <= 0x09E: // wave RAM (the bank NOT being played)
		i := (reg - 0x090) & 0xF
		w := 1 - a.ch3.bank
		a.ch3.ram[w][i] = byte(v)
		a.ch3.ram[w][i+1] = byte(v >> 8)

	case reg == 0x0A0 || reg == 0x0A2: // FIFO_A
		a.dsA.push(byte(v))
		a.dsA.push(byte(v >> 8))
	case reg == 0x0A4 || reg == 0x0A6: // FIFO_B
		a.dsB.push(byte(v))
		a.dsB.push(byte(v >> 8))
	default:
		return false
	}
	return true
}

func (a *apu) writeDutyEnv(c *square, v uint16) {
	c.length = 64 - int(v&0x3F)
	c.dutySel = int(v >> 6 & 3)
	c.envPeriod = int(v >> 8 & 7)
	c.vol = int(v >> 12 & 0xF)
	c.envDir = -1
	if v&(1<<11) != 0 {
		c.envDir = 1
	}
	if v&0xF800 == 0 {
		c.enabled = false // the DAC is off when volume and direction are both 0
	}
}

func (a *apu) writeFreqCtl(c *square, v uint16) {
	c.freq = int(v & 0x7FF)
	c.lengthEn = v&(1<<14) != 0
	if v&(1<<15) != 0 { // trigger
		c.enabled = true
		c.envTimer = c.envPeriod
		if c.length == 0 {
			c.length = 64
		}
		if c.hasSweep {
			c.swShadow = c.freq
			c.swTimer = c.swPeriod
			c.swOn = c.swPeriod > 0 || c.swShift > 0
		}
	}
}

// soundRead services a read of the sound block. Only the registers with real
// read-back semantics are answered here; the rest come from the stored file.
func (m *Machine) soundRead(reg uint32) (uint16, bool) {
	a := m.apu
	switch {
	case reg == 0x084: // SOUNDCNT_X — the four channel-active status bits
		var v uint16
		if a.powered {
			v |= 1 << 7
		}
		for i, on := range []bool{a.ch1.enabled, a.ch2.enabled, a.ch3.enabled, a.ch4.enabled} {
			if on {
				v |= 1 << uint(i)
			}
		}
		return v, true
	case reg >= 0x090 && reg <= 0x09E: // wave RAM: reads the bank not playing
		i := (reg - 0x090) & 0xF
		w := 1 - a.ch3.bank
		return uint16(a.ch3.ram[w][i]) | uint16(a.ch3.ram[w][i+1])<<8, true
	}
	return 0, false
}

// --- the Direct Sound transport ---------------------------------------------

// fifoTimerOverflow is called by the timer model when timer n overflows: each
// Direct Sound channel bound to that timer pops one sample, and a channel whose
// FIFO has run down asks its DMA for a refill.
func (m *Machine) fifoTimerOverflow(n int, times int) {
	h := m.io[0x082]
	for _, ch := range []struct {
		f        *fifo
		timerSel int
		dmaCh    int
	}{
		{&m.apu.dsA, int(h >> 10 & 1), 1},
		{&m.apu.dsB, int(h >> 14 & 1), 2},
	} {
		if ch.timerSel != n {
			continue
		}
		for i := 0; i < times; i++ {
			ch.f.pop()
		}
		// The hardware requests a refill when the FIFO is half empty or worse.
		if len(ch.f.q) <= 16 {
			m.dmaSoundRefill(ch.dmaCh)
		}
	}
}

// --- mixing ------------------------------------------------------------------

// mixCycles generates the output samples covering the given number of CPU
// cycles and appends them to the capture buffer.
func (a *apu) mixCycles(cycles int) {
	if !a.Capture {
		return
	}
	a.sampleAcc += float64(cycles) * audioRate / cpuHz
	n := int(a.sampleAcc)
	a.sampleAcc -= float64(n)
	for i := 0; i < n; i++ {
		a.mixOne()
	}
}

// psgVolTable maps SOUNDCNT_H bits 0-1 to the PSG mix level.
var psgVolTable = [4]float64{0.25, 0.5, 1.0, 1.0}

// mixOne produces one stereo output frame.
func (a *apu) mixOne() {
	const dt = 1.0 / audioRate

	// The frame sequencer (512 Hz) clocks length, envelope and sweep.
	a.frameAcc += cpuHz / audioRate
	for a.frameAcc >= cpuHz/512 {
		a.frameAcc -= cpuHz / 512
		a.stepFrame(a.frameSeq)
		a.frameSeq = (a.frameSeq + 1) & 7
	}

	s1 := a.ch1.sample(dt)
	s2 := a.ch2.sample(dt)
	s3 := a.ch3.sample(dt)
	s4 := a.ch4.sample(dt)

	cntL, cntH := a.mixRegL, a.mixRegH
	var l, r float64
	pan := func(s float64, bit int) {
		if cntL&(1<<uint(8+bit)) != 0 {
			r += s
		}
		if cntL&(1<<uint(12+bit)) != 0 {
			l += s
		}
	}
	pan(s1, 0)
	pan(s2, 1)
	pan(s3, 2)
	pan(s4, 3)
	// PSG master volume (SOUNDCNT_L bits 0-2 right, 4-6 left) then the
	// SOUNDCNT_H PSG mix level.
	r *= float64(cntL&7+1) / 8
	l *= float64(cntL>>4&7+1) / 8
	psg := psgVolTable[cntH&3]
	l *= psg
	r *= psg

	// Direct Sound: signed 8-bit, at 50% or 100% depending on SOUNDCNT_H.
	dsMix := func(f *fifo, volBit, rBit, lBit int) {
		v := float64(f.cur) / 128
		if cntH&(1<<uint(volBit)) == 0 {
			v *= 0.5
		}
		if cntH&(1<<uint(rBit)) != 0 {
			r += v
		}
		if cntH&(1<<uint(lBit)) != 0 {
			l += v
		}
	}
	dsMix(&a.dsA, 2, 8, 9)
	dsMix(&a.dsB, 3, 12, 13)

	clamp := func(v float64) int16 {
		v *= 0.35 // headroom: six channels can be live at once
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		return int16(v * 32000)
	}
	a.PCM = append(a.PCM, clamp(l), clamp(r))
}

func (a *apu) stepFrame(step int) {
	if step%2 == 0 {
		a.clockLength()
	}
	if step == 2 || step == 6 {
		a.clockSweep()
	}
	if step == 7 {
		a.clockEnv()
	}
}

func (a *apu) clockLength() {
	dec := func(en *bool, length *int, lengthEn bool) {
		if lengthEn && *length > 0 {
			*length--
			if *length == 0 {
				*en = false
			}
		}
	}
	dec(&a.ch1.enabled, &a.ch1.length, a.ch1.lengthEn)
	dec(&a.ch2.enabled, &a.ch2.length, a.ch2.lengthEn)
	dec(&a.ch3.enabled, &a.ch3.length, a.ch3.lengthEn)
	dec(&a.ch4.enabled, &a.ch4.length, a.ch4.lengthEn)
}

func (a *apu) clockSweep() {
	c := &a.ch1
	if !c.swOn || c.swPeriod == 0 {
		return
	}
	if c.swTimer--; c.swTimer > 0 {
		return
	}
	c.swTimer = c.swPeriod
	next := c.swShadow + c.swDir*(c.swShadow>>uint(c.swShift))
	if next > 2047 {
		c.enabled = false
		return
	}
	if next < 0 {
		return
	}
	c.swShadow, c.freq = next, next
}

func (a *apu) clockEnv() {
	step := func(vol *int, dir, period int, timer *int) {
		if period == 0 {
			return
		}
		if *timer--; *timer > 0 {
			return
		}
		*timer = period
		v := *vol + dir
		if v >= 0 && v <= 15 {
			*vol = v
		}
	}
	step(&a.ch1.vol, a.ch1.envDir, a.ch1.envPeriod, &a.ch1.envTimer)
	step(&a.ch2.vol, a.ch2.envDir, a.ch2.envPeriod, &a.ch2.envTimer)
	step(&a.ch4.vol, a.ch4.envDir, a.ch4.envPeriod, &a.ch4.envTimer)
}

// --- per-channel synthesis ---------------------------------------------------

func (c *square) sample(dt float64) float64 {
	if !c.enabled || c.freq >= 2048 {
		return 0
	}
	hz := 131072.0 / float64(2048-c.freq)
	c.phase += hz * dt
	c.phase -= float64(int(c.phase))
	pos := int(c.phase * 8)
	v := float64(c.vol) / 15
	if dutyTable[c.dutySel][pos&7] == 0 {
		return -v
	}
	return v
}

func (c *waveCh) sample(dt float64) float64 {
	if !c.enabled || !c.dacOn || c.freq >= 2048 {
		return 0
	}
	hz := 2097152.0 / float64(2048-c.freq)
	n := 32
	if c.twoBanks {
		n = 64
	}
	c.phase += hz / float64(n) * dt
	c.phase -= float64(int(c.phase))
	pos := int(c.phase * float64(n))
	bank := c.bank
	if c.twoBanks && pos >= 32 {
		bank, pos = 1-bank, pos-32
	}
	b := c.ram[bank&1][(pos>>1)&0xF]
	nib := b & 0xF
	if pos&1 == 0 {
		nib = b >> 4
	}
	var v float64
	switch {
	case c.force75:
		v = float64(nib) * 0.75
	case c.volShift == 0:
		return 0
	default:
		v = float64(nib >> uint(c.volShift-1))
	}
	return v/7.5 - 1
}

func (c *noiseCh) sample(dt float64) float64 {
	if !c.enabled || c.shift >= 14 {
		return 0
	}
	hz := 524288.0 / float64(c.divisor) / float64(uint(1)<<uint(c.shift))
	c.timer += hz * dt
	for c.timer >= 1 {
		c.timer--
		bit := (c.lfsr ^ (c.lfsr >> 1)) & 1
		c.lfsr >>= 1
		c.lfsr |= bit << 14
		if c.width7 {
			c.lfsr = c.lfsr&^(1<<6) | bit<<6
		}
	}
	v := float64(c.vol) / 15
	if c.lfsr&1 == 0 {
		return v
	}
	return -v
}

// --- capture output ----------------------------------------------------------

// AudioCapture turns the mixer on. Off by default: a run nobody wants sound
// from should not pay for synthesising it.
func (m *Machine) AudioCapture(on bool) { m.apu.Capture = on }

// AudioSamples reports how many stereo frames have been captured.
func (m *Machine) AudioSamples() int { return len(m.apu.PCM) / 2 }

// WriteWAV writes the captured mix as 16-bit stereo RIFF/WAVE — the same
// verification artefact tools/platform/dc and n3ds produce, so a future
// reimplementation of the game's own sequencer can be checked against the
// sound its driver actually drove out of the hardware.
func (m *Machine) WriteWAV(path string) error {
	pcm := m.apu.PCM
	if len(pcm) == 0 {
		return fmt.Errorf("gbamachine: no audio captured (was AudioCapture set before the run?)")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data := make([]byte, len(pcm)*2)
	for i, s := range pcm {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(s))
	}
	const channels = 2
	w := func(vals ...any) {
		for _, v := range vals {
			_ = binary.Write(f, binary.LittleEndian, v)
		}
	}
	f.WriteString("RIFF")
	w(uint32(36 + len(data)))
	f.WriteString("WAVEfmt ")
	w(uint32(16), uint16(1), uint16(channels), uint32(audioRate),
		uint32(audioRate*channels*2), uint16(channels*2), uint16(16))
	f.WriteString("data")
	w(uint32(len(data)))
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Close()
}

// SoundState reports the sound block as one line per subsystem: the mixing
// registers, the four PSG channels, and — the part that actually explains a
// silent game — the Direct Sound transport chain, which is only as alive as
// the timer driving it and the DMA refilling it. A game goes quiet because one
// link in that chain stopped, and this names which.
func (m *Machine) SoundState() []string {
	a := m.apu
	out := []string{
		fmt.Sprintf("SOUNDCNT_L=%04X SOUNDCNT_H=%04X SOUNDCNT_X=%04X (powered=%v)",
			a.mixRegL, a.mixRegH, m.io[0x084], a.powered),
		// Per channel, "enabled" alone answers nothing: a channel whose envelope
		// has decayed to volume 0 is enabled and silent, which is exactly what a
		// game between notes looks like. Print the volume and frequency too.
		fmt.Sprintf("PSG ch1 on=%v vol=%2d freq=%4d | ch2 on=%v vol=%2d freq=%4d | ch3 on=%v vol>>%d freq=%4d | ch4 on=%v vol=%2d",
			a.ch1.enabled, a.ch1.vol, a.ch1.freq, a.ch2.enabled, a.ch2.vol, a.ch2.freq,
			a.ch3.enabled, a.ch3.volShift, a.ch3.freq, a.ch4.enabled, a.ch4.vol),
	}
	for i, ds := range []struct {
		name     string
		f        *fifo
		timerSel int
		dmaCh    int
	}{
		{"A", &a.dsA, int(a.mixRegH >> 10 & 1), 1},
		{"B", &a.dsB, int(a.mixRegH >> 14 & 1), 2},
	} {
		t := &m.timers[ds.timerSel]
		d := &m.dma[ds.dmaCh]
		rate := 0.0
		if t.ctrl&(1<<7) != 0 {
			rate = cpuHz / float64(timerPrescale[t.ctrl&3]) / float64(0x10000-int(t.reload))
		}
		out = append(out, fmt.Sprintf(
			"DirectSound %s: fifo %2d/32 cur=%4d | timer%d ctrl=%04X reload=%04X -> %.0f Hz | DMA%d ctrl=%04X (timing %d, %s) src=%08X dst=%08X",
			ds.name, len(ds.f.q), ds.f.cur, ds.timerSel, t.ctrl, t.reload, rate,
			ds.dmaCh, d.ctrl, d.ctrl>>12&3, enabledStr(d.ctrl&(1<<15) != 0),
			d.latchSrc, d.dst))
		_ = i
	}
	return out
}

func enabledStr(on bool) string {
	if on {
		return "enabled"
	}
	return "DISABLED"
}
