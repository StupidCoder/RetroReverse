package dc

// aica_wav.go writes what the AICA actually mixed to a playable file — the
// verification oracle for the sound side, in the shape n3ds/dsp_wav.go
// established: any future reimplementation of the game's sequencer or
// streaming code is checked against the sound the game's own driver drove
// out of the hardware.

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// aicaSampleRate is the AICA's output rate.
const aicaSampleRate = 44100

// WriteWAV writes the captured final mix (16-bit stereo PCM) as a RIFF/WAVE
// file. Capturing must have been enabled before the run (Machine.AudioCapture).
func (m *Machine) WriteWAV(path string) error {
	if len(m.AudioPCM) == 0 {
		return fmt.Errorf("no audio captured (the AICA mixed no samples)")
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
	w(uint32(16), uint16(1), uint16(channels), uint32(aicaSampleRate),
		uint32(aicaSampleRate*channels*bits/8), uint16(channels*bits/8), uint16(bits))
	f.WriteString("data")
	w(dataLen)
	return binary.Write(f, binary.LittleEndian, m.AudioPCM)
}

// AudioSummary reports what the mix contained — samples, seconds, and
// peak/RMS amplitude — so a run can say "the sound system produced silence"
// or "it produced signal" without opening the file.
func (m *Machine) AudioSummary() string {
	if len(m.AudioPCM) == 0 {
		return "no audio samples mixed"
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
		frames, float64(frames)/aicaSampleRate, aicaSampleRate, peak, rms)
}
