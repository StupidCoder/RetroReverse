package n3ds

// dsp_wav.go writes what the DSP model actually mixed to a playable file. This
// is the verification oracle for the audio side: any future reimplementation of
// a title's sequencer, streaming or effects code is checked against the sound
// the game's own code drove out of the hardware.

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// dspSampleRate is the DSP's output rate: 160 samples per audio frame at the
// firmware's frame clock.
const dspSampleRate = 32728

// WriteWAV writes the captured final mix (16-bit stereo PCM) as a RIFF/WAVE
// file. Capturing must have been enabled before the run (Machine.AudioCapture).
func (m *Machine) WriteWAV(path string) error {
	if len(m.AudioPCM) == 0 {
		return fmt.Errorf("no audio captured (the DSP mixed no frames)")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	const channels, bits = 2, 16
	dataLen := uint32(len(m.AudioPCM) * 2)
	w := func(vals ...any) {
		for _, v := range vals {
			binary.Write(f, binary.LittleEndian, v)
		}
	}
	f.WriteString("RIFF")
	w(36 + dataLen)
	f.WriteString("WAVEfmt ")
	w(uint32(16), uint16(1), uint16(channels), uint32(dspSampleRate),
		uint32(dspSampleRate*channels*bits/8), uint16(channels*bits/8), uint16(bits))
	f.WriteString("data")
	w(dataLen)
	return binary.Write(f, binary.LittleEndian, m.AudioPCM)
}

// AudioSummary reports what the mix contained — frames, seconds, and peak/RMS
// amplitude — so a run can say "the sound system produced silence" or "it
// produced signal" without opening the file.
func (m *Machine) AudioSummary() string {
	if len(m.AudioPCM) == 0 {
		return "no audio frames mixed"
	}
	var peak int32
	var sumsq float64
	for _, s := range m.AudioPCM {
		v := int32(s)
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
		sumsq += float64(s) * float64(s)
	}
	frames := len(m.AudioPCM) / 2
	rms := math.Sqrt(sumsq / float64(len(m.AudioPCM)))
	return fmt.Sprintf("%d samples (%.2fs at %d Hz), peak %d, rms %.1f",
		frames, float64(frames)/dspSampleRate, dspSampleRate, peak, rms)
}
