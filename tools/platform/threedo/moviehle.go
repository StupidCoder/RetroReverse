package threedo

// moviehle.go high-level-emulates the 3DO movie player. On real hardware a game
// plays streamed FMV through the SDK DataStreamer: a video subscriber (Cinepak)
// and an audio subscriber (SDX2 on the DSP) fed from an interleaved .Stream. Our
// oracle does not model the DSP audio folio the audio subscriber needs, so the
// game's own player corrupts itself. Instead, when MovieHLE is on, we intercept
// the movie open, decode the Cinepak ourselves (cvid.go, verified byte-identical
// to the reference) and blit each frame into the game's display buffer, driving
// the normal present hook so the frames flow to -shots and the framedbg adapter
// exactly like game frames. The game requests the movie; the HLE plays it.

import "image"

// armedMovie is a movie the game asked to play, demuxed and ready to decode.
type armedMovie struct {
	name string
	mv   *CvidMovie
}

// armMovie demuxes the named .stream off the disc and queues it for playback.
// Called from loadDiscFile when MovieHLE intercepts a movie open. Non-cvid or
// undecodable streams are ignored (the game's fallback still runs).
func (m *Machine) armMovie(name string) {
	trimmed := name
	for len(trimmed) > 0 && trimmed[0] == '/' {
		trimmed = trimmed[1:]
	}
	data, err := m.vol.ReadFile(trimmed)
	if err != nil {
		// Retry by base name (the loader sometimes prepends a path prefix).
		if base := baseName(trimmed); base != trimmed {
			if e, rerr := m.vol.resolve(base); rerr == nil {
				data, err = m.vol.ReadFile(e.Path)
			}
		}
	}
	if err != nil {
		return
	}
	mv, err := DemuxStream(data)
	if err != nil || (mv.Codec != "cvid" && mv.Codec != "") || len(mv.Frames) == 0 {
		return
	}
	m.movieQueue = append(m.movieQueue, &armedMovie{name: trimmed, mv: mv})
	m.note("MovieHLE armed " + trimmed)
}

// MoviesPending reports how many queued movies have not finished playing.
func (m *Machine) MoviesPending() int { return len(m.movieQueue) - m.moviePos }

// StepMovieFrame decodes and blits the next queued movie frame into the display
// buffer, advancing the playback cursor. It returns the buffer written, the
// movie name, the frame index and total, and ok=false when no movie frames
// remain. It does NOT fire OnDisplay — the caller presents/captures the buffer.
// This is the per-frame primitive the framedbg adapter steps through; PlayMovies
// loops it for the batch -movieshots path.
func (m *Machine) StepMovieFrame() (buf uint32, name string, frame, total int, ok bool) {
	for m.moviePos < len(m.movieQueue) {
		am := m.movieQueue[m.moviePos]
		if m.movieDec == nil {
			m.movieDec = NewCvidDecoder(am.mv.Width, am.mv.Height)
			m.movieBase = m.movieTarget()
			m.movieFrameIdx = 0
		}
		if m.movieFrameIdx >= len(am.mv.Frames) {
			m.moviePos++
			m.movieDec = nil
			continue
		}
		m.movieDec.DecodeFrame(am.mv.Frames[m.movieFrameIdx])
		m.blitRGBAToVRAM(m.movieDec.Frame(), m.movieBase, am.mv.Width, am.mv.Height)
		m.displayBuf = m.movieBase
		idx := m.movieFrameIdx
		m.movieFrameIdx++
		return m.movieBase, am.name, idx, len(am.mv.Frames), true
	}
	return 0, "", 0, 0, false
}

// MovieNames returns the queued movies' names, in the order the game opened them.
func (m *Machine) MovieNames() []string {
	names := make([]string, len(m.movieQueue))
	for i, am := range m.movieQueue {
		names[i] = am.name
	}
	return names
}

// PlayMovies decodes and presents every queued movie into the display buffer,
// one frame per present. onFrame is called after each frame is blitted (before
// the OnDisplay present hook fires) so a caller can capture or pace it; it may be
// nil. Movies are cleared from the queue as they play. The display buffer is the
// one the game last showed; if the game never made a screen, a VRAM buffer is
// allocated so playback still has somewhere to land.
func (m *Machine) PlayMovies(onFrame func(name string, frame, total int)) {
	for {
		buf, name, frame, total, ok := m.StepMovieFrame()
		if !ok {
			break
		}
		if onFrame != nil {
			onFrame(name, frame, total)
		}
		m.frame++
		if m.OnDisplay != nil {
			m.OnDisplay(m, m.frame, buf)
		}
	}
}

// movieTarget picks the VRAM buffer to play a movie into: the buffer the game is
// currently displaying, else any bitmap it has made, else a freshly allocated
// screen-sized buffer.
func (m *Machine) movieTarget() uint32 {
	if m.displayBuf != 0 {
		return m.displayBuf
	}
	if buf, _, _, ok := m.DrawTarget(); ok {
		m.displayBuf = buf
		return buf
	}
	buf := m.vheap.alloc(uint32(screenW * screenH * 2))
	m.displayBuf = buf
	return buf
}

// blitRGBAToVRAM writes a decoded RGBA frame into a VRAM buffer in the hardware's
// interleaved line-pair RGB555 layout (the inverse of CaptureVRAM). The frame is
// centred vertically in the 320x240 screen; letterbox rows are cleared to black.
func (m *Machine) blitRGBAToVRAM(img *image.RGBA, base uint32, w, h int) {
	yoff := (screenH - h) / 2
	if yoff < 0 {
		yoff = 0
	}
	for sy := 0; sy < screenH; sy++ {
		src := sy - yoff
		for sx := 0; sx < screenW; sx++ {
			o := base - vramBase + uint32(sy>>1)*uint32(screenW)*4 + uint32(sx)*4 + uint32(sy&1)*2
			if int(o)+2 > len(m.vram) {
				continue
			}
			var c uint16
			if src >= 0 && src < h && sx < w {
				p := img.PixOffset(sx, src)
				c = uint16(img.Pix[p]>>3)<<10 | uint16(img.Pix[p+1]>>3)<<5 | uint16(img.Pix[p+2]>>3)
			}
			m.vram[o] = byte(c >> 8)
			m.vram[o+1] = byte(c)
		}
	}
}

// screen dimensions the 3DO games here present at.
const (
	screenW = 320
	screenH = 240
)

// baseName returns the final path element of p.
func baseName(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[i+1:]
		}
	}
	return p
}
