package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"sync"
	"time"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/wsock"
)

// A session is one browser attached to one DebugTarget.
//
// debug.DebugTarget is single-threaded by contract, so exactly one goroutine — the
// one running serve — ever touches it. The socket reader runs separately and hands
// requests over through a mailbox, which is also where a scrubber drag gets
// collapsed to its newest request instead of queueing a hundred replays nobody will
// look at.
// fastStepper is an optional target capability: advance a field capturing nothing.
// Play mode uses it to fast-forward — a full StepFrame would snapshot the machine and
// run a per-pixel census for a frame nobody is going to inspect. A target that does
// not offer it falls back to a normal capture step.
type fastStepper interface {
	StepFast() error
}

type session struct {
	tgt  debug.DebugTarget
	rom  string
	conn *wsock.Conn

	fc      *debug.FrameCapture
	frameNo int

	// Play mode. inFlight holds the loop to one unacknowledged frame at a time: the
	// emulator can outrun the browser's ability to draw, and without this the socket
	// would grow a backlog of frames nobody will ever see.
	playing  bool
	inFlight bool

	// cache holds draw targets already replayed for the current capture, so
	// nudging the scrubber back and forth over ground already covered is free.
	// Every RenderAfter is a full frame replay from the frame-start snapshot (the
	// scratch machine cannot resume mid-queue), which is what makes this worth it.
	cache map[int]*image.RGBA
	order []int // insertion order, for the LRU eviction
}

const cacheMax = 24 // ~7 MB at 320x240; a scrubber drag rarely revisits more

func newSession(tgt debug.DebugTarget, rom string, conn *wsock.Conn) *session {
	return &session{tgt: tgt, rom: rom, conn: conn, cache: map[int]*image.RGBA{}}
}

// serve runs the session until the socket closes.
func (s *session) serve() {
	mb := newMailbox()
	go s.readLoop(mb)

	s.send(helloMsg{Type: "hello", Target: s.tgt.Name(), ROM: s.rom})

	for {
		// While playing, the session must not block waiting for a request: it runs a
		// frame, then picks up whatever has arrived. It only blocks when there is a
		// frame in flight (the next thing to arrive will be its ack) or when paused.
		if s.playing && !s.inFlight {
			if err := s.playFrame(); err != nil {
				s.playing = false
				s.send(errMsg{Type: "error", Msg: err.Error()})
			}
			for {
				r, ok := mb.tryPop()
				if !ok {
					break
				}
				if err := s.handle(r); err != nil {
					s.send(errMsg{Type: "error", Seq: r.Seq, Msg: err.Error()})
				}
			}
			continue
		}

		r, ok := mb.pop()
		if !ok {
			return
		}
		if err := s.handle(r); err != nil {
			s.send(errMsg{Type: "error", Seq: r.Seq, Msg: err.Error()})
		}
	}
}

// readLoop parses client messages and queues them. It is the only reader of conn.
func (s *session) readLoop(mb *mailbox) {
	defer mb.close()
	for {
		op, msg, err := s.conn.Read()
		if err != nil {
			return
		}
		if op != wsock.OpText {
			continue
		}
		var r req
		if err := json.Unmarshal(msg, &r); err != nil {
			s.send(errMsg{Type: "error", Msg: "bad request: " + err.Error()})
			continue
		}
		mb.push(r)
	}
}

func (s *session) handle(r req) error {
	switch r.Op {
	case "step":
		return s.step(r)
	case "play":
		return s.play(r)
	case "ack":
		s.inFlight = false
		return nil
	case "scrub":
		return s.scrub(r)
	case "display":
		return s.display(r)
	case "cpu":
		s.sendCPU(r.Seq)
		return nil
	case "mem":
		return s.mem(r)
	case "pixel":
		return s.pixel(r)
	default:
		return fmt.Errorf("unknown op %q", r.Op)
	}
}

// step advances the machine. n <= 0 means "step to a frame worth looking at" —
// the same heuristic the headless framedbg uses.
func (s *session) step(r req) error {
	start := time.Now()
	var fc *debug.FrameCapture
	var err error
	if r.N > 0 {
		for i := 0; i < r.N; i++ {
			fc, err = s.tgt.StepFrame(r.Over)
			if err != nil {
				return err
			}
			s.frameNo++
		}
	} else {
		fc, err = s.stepToDrawn(r.Over)
		if err != nil {
			return err
		}
	}
	s.fc = fc
	s.clearCache()

	s.send(frameMsg{
		Type:     "frame",
		Seq:      r.Seq,
		Frame:    s.frameNo,
		Commands: toJSONCommands(fc.Commands),
		W:        fc.Width,
		H:        fc.Height,
		HasProv:  fc.Prov != nil,
		Overdraw: fc.Overdraw != nil,
		StepMs:   msSince(start),
	})
	// The provenance buffer goes over once per capture. From then on the page
	// answers "which command drew this pixel?" locally, with no round trip — that
	// is what makes hover-to-identify and the "highlight every pixel this command
	// drew" overlay instant.
	if fc.Prov != nil {
		s.conn.WriteBinary(encodeProv(r.Seq, fc))
	}
	return nil
}

// play starts or stops free-running the machine.
//
// While it runs, nothing is captured: each field is stepped with no snapshot and no
// census, and the page is sent the VI scanout — what the console would actually be
// showing. That is the point, since play mode is how you fast-forward to the part of
// the game you want to look at (the logo appearing, a particular menu).
//
// Stopping does a full capture step, so the moment you pause you get a real frame:
// its command stream, its provenance, its overdraw. That capture is the *next* field
// after the last one played, not a re-examination of it — a played field left no
// record behind to go back to.
func (s *session) play(r req) error {
	if r.On {
		if _, ok := s.tgt.(fastStepper); !ok {
			log.Printf("framedbg: target has no fast step; playing at capture speed")
		}
		s.playing = true
		s.inFlight = false
		return nil
	}
	if !s.playing {
		return nil
	}
	s.playing = false
	s.inFlight = false
	return s.step(req{Op: "step", Seq: r.Seq, N: 1, Over: r.Over})
}

// playFrame advances one field and streams the scanout. The binary carries seq 0:
// it answers no request, so the page draws it unconditionally.
func (s *session) playFrame() error {
	start := time.Now()
	if err := s.stepFast(); err != nil {
		return err
	}
	img, err := s.tgt.Display()
	if err != nil {
		// Early in a boot there is nothing being scanned out yet. Keep running — this
		// is exactly the stretch play mode exists to get through — just show nothing.
		return nil
	}
	payload := encodeImage(0, img)
	s.send(renderMsg{
		Type:     "render",
		Seq:      0,
		K:        -1,
		Play:     true,
		Frame:    s.frameNo,
		RenderMs: msSince(start),
		Bytes:    len(payload),
	})
	s.inFlight = true
	return s.conn.WriteBinary(payload)
}

// stepFast advances one field as cheaply as the target allows.
func (s *session) stepFast() error {
	s.frameNo++
	if f, ok := s.tgt.(fastStepper); ok {
		return f.StepFast()
	}
	_, err := s.tgt.StepFrame(false)
	return err
}

// stepToDrawn steps until a frame actually renders a scene, so a fresh boot lands
// on something worth inspecting rather than an empty field.
func (s *session) stepToDrawn(withOverdraw bool) (*debug.FrameCapture, error) {
	for i := 0; i < 800; i++ {
		fc, err := s.tgt.StepFrame(withOverdraw)
		if err != nil {
			return nil, err
		}
		s.frameNo++
		if len(fc.Commands) > 100 && fc.Prov != nil {
			return fc, nil
		}
	}
	return nil, fmt.Errorf("no drawn frame within the field budget")
}

// scrub renders the draw target as it stood after command k.
func (s *session) scrub(r req) error {
	if s.fc == nil {
		return fmt.Errorf("no captured frame yet")
	}
	k := r.K
	if k < 0 {
		k = 0
	}
	if n := len(s.fc.Commands); k >= n {
		k = n - 1
	}
	start := time.Now()
	img, cached := s.cache[k]
	if !cached {
		var err error
		img, err = s.tgt.RenderAfter(s.fc, k)
		if err != nil {
			return fmt.Errorf("render after command %d: %w", k, err)
		}
		s.put(k, img)
	}
	payload := encodeImage(r.Seq, img)
	s.send(renderMsg{
		Type:     "render",
		Seq:      r.Seq,
		K:        k,
		RenderMs: msSince(start),
		Bytes:    len(payload),
		Cached:   cached,
	})
	return s.conn.WriteBinary(payload)
}

// display sends the image currently being scanned out, as opposed to the draw
// target mid-frame. K is -1 to mark it as "not a command".
func (s *session) display(r req) error {
	start := time.Now()
	img, err := s.tgt.Display()
	if err != nil {
		return err
	}
	payload := encodeImage(r.Seq, img)
	s.send(renderMsg{Type: "render", Seq: r.Seq, K: -1, RenderMs: msSince(start), Bytes: len(payload)})
	return s.conn.WriteBinary(payload)
}

func (s *session) sendCPU(seq int) {
	c := s.tgt.CPU()
	m := cpuMsg{
		Type:  "cpu",
		Seq:   seq,
		PC:    fmt.Sprintf("%016x", c.PC),
		Names: c.Names,
		Vals:  make([]string, len(c.Vals)),
		Extra: map[string]string{},
	}
	for i, v := range c.Vals {
		m.Vals[i] = fmt.Sprintf("%016x", v)
	}
	for k, v := range c.Extra {
		m.Extra[k] = fmt.Sprintf("%016x", v)
	}
	s.send(m)
}

func (s *session) mem(r req) error {
	n := r.Len
	if n <= 0 {
		n = 256
	}
	if n > 64*1024 {
		n = 64 * 1024
	}
	data := s.tgt.ReadMem(r.Addr, n)
	s.send(memMsg{Type: "mem", Seq: r.Seq, Addr: r.Addr, Data: hex.EncodeToString(data)})
	return nil
}

// pixel answers the overdraw question for one pixel. The page already knows the
// last writer from its copy of Prov; what it cannot know is the full write history,
// which lives in the capture's Overdraw map (only populated when the frame was
// stepped with overdraw recording on).
func (s *session) pixel(r req) error {
	if s.fc == nil {
		return fmt.Errorf("no captured frame yet")
	}
	m := pixelMsg{Type: "pixel", Seq: r.Seq, X: r.X, Y: r.Y, Cmd: s.fc.ProvAt(r.X, r.Y), Writes: []jsonWrite{}}
	if s.fc.Overdraw != nil && r.X >= 0 && r.Y >= 0 && r.X < s.fc.Width && r.Y < s.fc.Height {
		for _, w := range s.fc.Overdraw[r.Y*s.fc.Width+r.X] {
			name := ""
			if w.CmdIndex >= 0 && w.CmdIndex < len(s.fc.Commands) {
				name = s.fc.Commands[w.CmdIndex].Name
			}
			m.Writes = append(m.Writes, jsonWrite{
				Cmd: w.CmdIndex, Name: name,
				R: w.R, G: w.G, B: w.B, A: w.A, Rejected: w.Rejected,
			})
		}
	}
	s.send(m)
	return nil
}

func (s *session) put(k int, img *image.RGBA) {
	if len(s.order) >= cacheMax {
		delete(s.cache, s.order[0])
		s.order = s.order[1:]
	}
	s.cache[k] = img
	s.order = append(s.order, k)
}

func (s *session) clearCache() {
	s.cache = map[int]*image.RGBA{}
	s.order = s.order[:0]
}

// send marshals and writes a JSON message.
func (s *session) send(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("framedbg: marshalling %T: %v", v, err)
		return
	}
	s.conn.WriteText(b)
}

func msSince(t time.Time) float64 {
	return float64(time.Since(t).Microseconds()) / 1000
}

// mailbox is the request queue between the socket reader and the session goroutine.
// Ordinary requests queue in order; a coalescable one (a scrub) replaces any scrub
// still waiting, so dragging the scrubber fast means the session skips straight to
// where the mouse ended up instead of replaying every position it passed through.
type mailbox struct {
	mu     sync.Mutex
	q      []req
	sig    chan struct{}
	closed bool
}

func newMailbox() *mailbox { return &mailbox{sig: make(chan struct{}, 1)} }

func (m *mailbox) push(r req) {
	m.mu.Lock()
	if coalescable(r.Op) {
		for i := range m.q {
			if m.q[i].Op == r.Op {
				m.q[i] = r
				m.mu.Unlock()
				m.wake()
				return
			}
		}
	}
	m.q = append(m.q, r)
	m.mu.Unlock()
	m.wake()
}

// tryPop takes the next request if one is waiting, and never blocks — the play loop
// uses it to stay responsive between frames.
func (m *mailbox) tryPop() (req, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.q) == 0 {
		return req{}, false
	}
	r := m.q[0]
	m.q = m.q[1:]
	return r, true
}

// pop blocks for the next request; ok is false once the mailbox is closed and drained.
func (m *mailbox) pop() (req, bool) {
	for {
		m.mu.Lock()
		if len(m.q) > 0 {
			r := m.q[0]
			m.q = m.q[1:]
			m.mu.Unlock()
			return r, true
		}
		closed := m.closed
		m.mu.Unlock()
		if closed {
			return req{}, false
		}
		<-m.sig
	}
}

func (m *mailbox) close() {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()
	m.wake()
}

func (m *mailbox) wake() {
	select {
	case m.sig <- struct{}{}:
	default:
	}
}
