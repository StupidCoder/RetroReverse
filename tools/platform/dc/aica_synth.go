package dc

// aica_synth.go is the sound the machine actually makes: the 64 voices'
// sample fetch, envelope and stereo mix at the 44.1 kHz tick the position
// model (aica.go) already steps. What is modeled here is exactly what the
// slot-register census said Crazy Taxi's driver writes — sample address and
// format, loop points, the amplitude EG, pitch, TL, the direct sends
// (DISDL/DIPAN) and the master volume. What the census said the driver never
// touches is deliberately absent: the LFO words are always zero, the filter
// registers sit at their neutral values. The DSP effect path (the driver
// does load a program and set the slots' ISEL/IMXL sends) is not mixed —
// the dry path carries the music; a missing reverb is a fidelity gap, not a
// protocol one, and it is census-logged so it stays visible.
//
// The mix always runs, whether or not anyone is listening: the decoder and
// envelope state are part of the machine (they savestate), and a run with
// -wav must be byte-identical to a run without it. AudioCapture only gates
// appending to AudioPCM, which is an instrument buffer — never savestated,
// never hashed.
//
// The envelope timing is an approximation: Yamaha's exponential-attack /
// linear-dB-decay shape with the effective-rate period table collapsed to
// its power-of-two skeleton. The protocol side (what states the 2810
// monitor reports, when a one-shot ends) is exact; the milliseconds are not
// datasheet-exact and nothing in the driver's handshake depends on them.

// egAttack..egRelease are the EG states as the 2810 monitor encodes them.
const (
	egAttack = iota
	egDecay1
	egDecay2
	egRelease
)

// egSilence is the 10-bit attenuation floor: -96 dB, the EG's "done".
const egSilence = 0x3FF

// Q16 gain tables, built once at init. Floats appear only here, with fixed
// inputs — the tables are constants by the time any run uses them.
var (
	egGainQ16  [egSilence + 1]int64 // 10-bit EG attenuation, 96 dB full scale
	tlGainQ16  [256]int64           // TL: 0.375 dB per step
	sdlGainQ16 [16]int64            // DISDL/MVOL: 3 dB per step, 0 = off
	panGainQ16 [16]int64            // DIPAN magnitude: 3 dB per step, 15 = off
)

func init() {
	pow := func(db float64) int64 {
		g := 1.0
		for i := 0.0; i < db; i++ {
			g *= 0.8912509381337456 // 10^(-1/20), one dB down
		}
		return int64(g * 65536)
	}
	for i := range egGainQ16 {
		egGainQ16[i] = pow(96.0 * float64(i) / float64(egSilence))
	}
	for i := range tlGainQ16 {
		tlGainQ16[i] = pow(0.375 * float64(i))
	}
	for i := range sdlGainQ16 {
		if i == 0 {
			continue // 0 = no send
		}
		sdlGainQ16[i] = pow(3.0 * float64(15-i))
	}
	for i := range panGainQ16 {
		if i == 15 {
			continue // full pan silences the far channel
		}
		panGainQ16[i] = pow(3.0 * float64(i))
	}
}

// adpcmScale is the Yamaha step-size adaptation, in 1/256ths.
var adpcmScale = [8]int32{230, 230, 230, 230, 307, 409, 512, 614}

// adpcmInitStep is the decoder's reset step size.
const adpcmInitStep = 0x7F

// keyOn arms a voice from sample zero with a fresh envelope and decoder.
func (s *AICASlot) keyOn() {
	*s = AICASlot{Active: true, EGState: egAttack, EGLevel: egSilence,
		AdStep: adpcmInitStep, DecPos: ^uint32(0)}
}

// keyOff sends an active voice into release; the EG stepper retires it when
// the level reaches the floor.
func (s *AICASlot) keyOff() {
	if s.Active {
		s.EGState = egRelease
	}
}

// effRate is the Yamaha effective rate: 2·R plus the key-rate-scaling
// contribution, which KRS=0xF disables. The OCT/FNS refinement of the
// scaled path is not modeled.
func effRate(r, krs uint32) uint32 {
	if r == 0 {
		return 0 // rate 0 never moves
	}
	er := 2 * r
	if krs != 0xF {
		er += krs
	}
	if er > 63 {
		er = 63
	}
	return er
}

// egPeriod is the samples between envelope steps at an effective rate — the
// power-of-two skeleton of the rate table.
func egPeriod(er uint32) uint64 {
	p := uint64(4096) >> (er >> 2)
	if p == 0 {
		return 1
	}
	return p
}

// stepEG advances one voice's envelope by one sample tick.
func (m *Machine) stepEG(s *AICASlot, slot uint32, tick uint64) {
	env := m.slotW(slot, 0x10)
	env2 := m.slotW(slot, 0x14)
	krs := env2 >> 10 & 0xF
	var r uint32
	switch s.EGState {
	case egAttack:
		r = env & 0x1F
	case egDecay1:
		r = env >> 6 & 0x1F
	case egDecay2:
		r = env >> 11 & 0x1F
	case egRelease:
		r = env2 & 0x1F
	}
	er := effRate(r, krs)
	if er == 0 {
		return
	}
	if s.EGState == egAttack && er >= 60 {
		// The fast-attack shortcut: at the top rates the hardware's attack is
		// effectively instantaneous.
		s.EGLevel, s.EGState = 0, egDecay1
		return
	}
	if tick%egPeriod(er) != 0 {
		return
	}
	switch s.EGState {
	case egAttack:
		step := uint16(s.EGLevel>>4) + 1
		if s.EGLevel <= step {
			s.EGLevel, s.EGState = 0, egDecay1
		} else {
			s.EGLevel -= step
		}
	case egDecay1:
		if s.EGLevel < egSilence {
			s.EGLevel++
		}
		if dl := uint16(env2 >> 5 & 0x1F << 5); s.EGLevel >= dl {
			s.EGState = egDecay2
		}
	case egDecay2:
		if s.EGLevel < egSilence {
			s.EGLevel++
		}
	case egRelease:
		if s.EGLevel < egSilence {
			s.EGLevel++
		} else {
			s.Active = false
		}
	}
}

// sampleAddr is the voice's base byte address in sound RAM: the 7 high bits
// live in the control word, the low 16 in the next register.
func (m *Machine) sampleAddr(slot uint32) uint32 {
	return (m.slotW(slot, 0x00)&0x7F)<<16 | m.slotW(slot, 0x04)&0xFFFF
}

// fetchPCM reads sample n of a PCM voice directly.
func (m *Machine) fetchPCM(sa, n uint32, eight bool) int32 {
	if eight {
		return int32(int8(m.AICARAM[(sa+n)&(AICARAMSize-1)])) << 8
	}
	a := (sa + 2*n) & (AICARAMSize - 1)
	return int32(int16(uint16(m.AICARAM[a]) | uint16(m.AICARAM[(a+1)&(AICARAMSize-1)])<<8))
}

// decodeADPCMTo runs the sequential 4-bit decoder forward until Cur holds
// sample `target`, crossing the loop start once caches the decoder state
// there (PCMS=2's loop behaviour; PCMS=3, the long-stream variant, never
// restores it — its caller never rewinds DecPos).
func (m *Machine) decodeADPCMTo(s *AICASlot, sa, target, lsa uint32) {
	if s.AdStep == 0 {
		s.AdStep = adpcmInitStep // a voice restored from a pre-synthesis savestate
	}
	for s.DecPos != target {
		n := s.DecPos + 1 // next sample index to decode; ^0 (fresh) wraps to 0
		b := m.AICARAM[(sa+n/2)&(AICARAMSize-1)]
		nib := b >> (4 * (n & 1)) & 0xF
		delta := s.AdStep * int32(2*(nib&7)+1) / 8
		if nib&8 != 0 {
			delta = -delta
		}
		h := s.AdHist + delta
		if h > 32767 {
			h = 32767
		} else if h < -32768 {
			h = -32768
		}
		s.AdHist = h
		s.AdStep = s.AdStep * adpcmScale[nib&7] / 256
		if s.AdStep < adpcmInitStep {
			s.AdStep = adpcmInitStep
		} else if s.AdStep > 0x6000 {
			s.AdStep = 0x6000
		}
		s.Prev, s.Cur = s.Cur, h
		s.DecPos = n
		if n == lsa && !s.LoopSeen {
			s.LoopSeen, s.LoopHist, s.LoopStep = true, s.AdHist, s.AdStep
		}
	}
}

// mixSample advances every voice one output sample and mixes the field's
// stereo pair: pitch phase (the position model the monitors report), the
// sample fetch, the envelope, the send levels. Appends to the capture
// buffer only when the instrument asked.
func (m *Machine) mixSample() {
	tick := m.Timers.SampleTick
	m.Timers.SampleTick++
	var accL, accR int64
	for i := range m.Slots {
		s := &m.Slots[i]
		if !s.Active {
			continue
		}
		slot := uint32(i)
		ctl := m.slotW(slot, 0x00)

		// Pitch phase, exactly as the position model has always stepped it.
		pitch := m.slotW(slot, 0x18)
		oct := int32(pitch>>11) & 0xF
		if oct >= 8 {
			oct -= 16
		}
		incr := uint64(0x400 + pitch&0x3FF)
		if oct >= 0 {
			incr <<= uint(oct)
		} else {
			incr >>= uint(-oct)
		}
		acc := uint64(s.Frac) + incr
		s.Frac = uint32(acc & 0x3FF)
		s.Pos += uint32(acc >> 10)
		lea := m.slotW(slot, 0x0C) & 0xFFFF
		lsa := m.slotW(slot, 0x08) & 0xFFFF
		wrapped := false
		if s.Pos >= lea {
			if ctl>>9&1 != 0 { // LPCTL: wrap to the loop start
				if lea > lsa {
					s.Pos = lsa + (s.Pos-lea)%(lea-lsa)
				} else {
					s.Pos = lsa
				}
				s.LP = true
				wrapped = true
			} else { // one-shot: the voice ends where the sample does
				s.Pos = lea
				s.Active = false
				continue
			}
		}

		// The sample under the play head, per format.
		sa := m.sampleAddr(slot)
		if ctl>>10&3 != 0 {
			m.logf("AICA slot SSCTL %d (noise source) unimplemented", ctl>>10&3)
		} else {
			switch fmtBits := ctl >> 7 & 3; fmtBits {
			case 0, 1: // 16-bit / 8-bit PCM: direct fetch
				if s.DecPos != s.Pos {
					if s.Pos == s.DecPos+1 {
						s.Prev = s.Cur
					} else if s.Pos > 0 {
						s.Prev = m.fetchPCM(sa, s.Pos-1, fmtBits == 1)
					}
					s.Cur = m.fetchPCM(sa, s.Pos, fmtBits == 1)
					s.DecPos = s.Pos
				}
			default: // 2, 3: Yamaha ADPCM, sequential
				if wrapped || s.Pos < s.DecPos && s.DecPos != ^uint32(0) {
					// Rewound to the loop start: PCMS=2 restores the decoder
					// state cached there, the long-stream variant continues.
					if fmtBits == 2 && s.LoopSeen {
						s.AdHist, s.AdStep = s.LoopHist, s.LoopStep
					}
					s.DecPos = ^uint32(0)
					if lsa > 0 {
						s.DecPos = lsa - 1
					}
					if s.Pos < lsa {
						s.DecPos = ^uint32(0)
					}
				}
				m.decodeADPCMTo(s, sa, s.Pos, lsa)
			}
		}

		// Envelope, then the voice's contribution through its send levels.
		m.stepEG(s, slot, tick)
		out := int64(s.Prev) + int64(s.Cur-s.Prev)*int64(s.Frac)/1024
		out = out * egGainQ16[s.EGLevel] >> 16
		out = out * tlGainQ16[m.slotW(slot, 0x28)>>8&0xFF] >> 16
		send := m.slotW(slot, 0x24)
		disdl := send >> 8 & 0xF
		if disdl == 0 {
			continue
		}
		out = out * sdlGainQ16[disdl] >> 16
		pan := send & 0x1F
		l, r := out, out
		if pan&0x10 != 0 {
			r = r * panGainQ16[pan&0xF] >> 16
		} else {
			l = l * panGainQ16[pan&0xF] >> 16
		}
		accL += l
		accR += r
	}
	if !m.AudioCapture {
		return
	}
	mvol := sdlGainQ16[m.AICARegs[0x2800]&0xF]
	accL = accL * mvol >> 16
	accR = accR * mvol >> 16
	m.AudioPCM = append(m.AudioPCM, clamp16(accL), clamp16(accR))
}

func clamp16(v int64) int16 {
	if v > 32767 {
		return 32767
	}
	if v < -32768 {
		return -32768
	}
	return int16(v)
}
