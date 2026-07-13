package server

import (
	"encoding/json"
	"log"
	"sync"

	"retroreverse.com/tools/debug/wsock"
)

// A session is one browser connection. It does not own the machine — the Runner does
// — so it is only a reader (parse a request, hand it to the runner) and a writer
// (replies addressed to it, and events broadcast to everyone).
type session struct {
	conn *wsock.Conn
	rn   *Runner
}

// serve reads the socket until it closes. Requests go to the runner's mailbox; the
// runner answers on its own goroutine.
func (s *session) serve() {
	s.rn.attach(s)
	defer s.rn.detach(s)

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
		s.rn.submit(s, r)
	}
}

// send marshals and writes a JSON message. wsock serialises concurrent writers, so
// the runner and a broadcast can both write to this connection safely.
func (s *session) send(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("framedbg: marshalling %T: %v", v, err)
		return
	}
	s.conn.WriteText(b)
}

func (s *session) sendBinary(b []byte) { s.conn.WriteBinary(b) }

// request is a client request tagged with who asked, so the reply goes back to that
// connection while events go to all of them.
type request struct {
	from *session
	req  req
}

// mailbox is the queue between the socket readers and the runner goroutine. Ordinary
// requests queue in order; a coalescable one (a scrub, a surface aim) replaces one of
// the same op still waiting, so dragging a slider fast means the runner skips to
// where the mouse ended up instead of replaying every position it passed through.
type mailbox struct {
	mu     sync.Mutex
	q      []request
	sig    chan struct{}
	closed bool
}

func newMailbox() *mailbox { return &mailbox{sig: make(chan struct{}, 1)} }

func (m *mailbox) push(r request) {
	m.mu.Lock()
	if coalescable(r.req.Op) {
		for i := range m.q {
			// Only collapse requests from the same session: two tabs scrubbing are
			// two separate questions, and both deserve an answer.
			if m.q[i].req.Op == r.req.Op && m.q[i].from == r.from {
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

// pop blocks for the next request; ok is false once the mailbox is closed and drained.
func (m *mailbox) pop() (request, bool) {
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
			return request{}, false
		}
		<-m.sig
	}
}

// tryPop takes the next request if one is waiting, and never blocks — the play and
// run loops use it to stay responsive between slices.
func (m *mailbox) tryPop() (request, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.q) == 0 {
		return request{}, false
	}
	r := m.q[0]
	m.q = m.q[1:]
	return r, true
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

// mustJSON marshals args for a request the runner synthesises for itself (pausing
// play turns into a step). The value is a literal struct, so this cannot fail.
func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// overdrawOf pulls the overdraw flag out of a play request, so that pausing captures
// the frame with the same detail the page asked for.
func overdrawOf(r req) bool {
	var a struct {
		Overdraw bool `json:"overdraw"`
	}
	_ = decodeArgs(r, &a)
	return a.Overdraw
}
