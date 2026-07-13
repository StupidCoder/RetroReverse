package wsock

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestAccept checks the handshake hash against the worked example in RFC 6455 §1.3.
func TestAccept(t *testing.T) {
	const key = "dGhlIHNhbXBsZSBub25jZQ=="
	const want = "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	if got := Accept(key); got != want {
		t.Errorf("Accept(%q) = %q, want %q", key, got, want)
	}
}

// client is a minimal WebSocket client, enough to drive the server side in tests.
type client struct {
	c  net.Conn
	br *bufio.Reader
}

// dial performs the handshake against a running httptest server.
func dial(t *testing.T, url string) *client {
	t.Helper()
	addr := strings.TrimPrefix(url, "http://")
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	req := "GET /ws HTTP/1.1\r\nHost: " + addr + "\r\n" +
		"Upgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" +
		"Sec-WebSocket-Version: 13\r\n\r\n"
	if _, err := io.WriteString(c, req); err != nil {
		t.Fatal(err)
	}
	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want 101", resp.StatusCode)
	}
	if got := resp.Header.Get("Sec-WebSocket-Accept"); got != "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" {
		t.Fatalf("Sec-WebSocket-Accept = %q", got)
	}
	t.Cleanup(func() { c.Close() })
	return &client{c: c, br: br}
}

// send writes a masked client frame, as the spec requires of a client.
func (cl *client) send(t *testing.T, opcode byte, payload []byte) {
	t.Helper()
	var buf bytes.Buffer
	buf.WriteByte(0x80 | opcode)
	n := len(payload)
	switch {
	case n < 126:
		buf.WriteByte(0x80 | byte(n))
	case n <= 0xFFFF:
		buf.WriteByte(0x80 | 126)
		binary.Write(&buf, binary.BigEndian, uint16(n))
	default:
		buf.WriteByte(0x80 | 127)
		binary.Write(&buf, binary.BigEndian, uint64(n))
	}
	key := [4]byte{0xDE, 0xAD, 0xBE, 0xEF}
	buf.Write(key[:])
	masked := make([]byte, n)
	for i, b := range payload {
		masked[i] = b ^ key[i&3]
	}
	buf.Write(masked)
	if _, err := cl.c.Write(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
}

// recv reads one server frame. Server frames are unmasked and unfragmented here.
func (cl *client) recv(t *testing.T) (opcode byte, payload []byte) {
	t.Helper()
	var h [2]byte
	if _, err := io.ReadFull(cl.br, h[:]); err != nil {
		t.Fatal(err)
	}
	opcode = h[0] & 0x0F
	if h[1]&0x80 != 0 {
		t.Fatal("server frame is masked; it must not be")
	}
	n := uint64(h[1] & 0x7F)
	switch n {
	case 126:
		var b [2]byte
		io.ReadFull(cl.br, b[:])
		n = uint64(binary.BigEndian.Uint16(b[:]))
	case 127:
		var b [8]byte
		io.ReadFull(cl.br, b[:])
		n = binary.BigEndian.Uint64(b[:])
	}
	payload = make([]byte, n)
	if _, err := io.ReadFull(cl.br, payload); err != nil {
		t.Fatal(err)
	}
	return opcode, payload
}

// echoServer upgrades and echoes every message back, so a test can watch a full
// round trip through the real codec.
func echoServer(t *testing.T) *httptest.Server {
	t.Helper()
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrade(w, r)
		if err != nil {
			return
		}
		defer close(done)
		for {
			op, msg, err := conn.Read()
			if err != nil {
				return
			}
			if op == OpText {
				conn.WriteText(msg)
			} else {
				conn.WriteBinary(msg)
			}
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRoundTripText(t *testing.T) {
	srv := echoServer(t)
	cl := dial(t, srv.URL)

	cl.send(t, OpText, []byte(`{"op":"step","seq":1}`))
	op, msg, err := readEcho(t, cl)
	if err != nil {
		t.Fatal(err)
	}
	if op != OpText || string(msg) != `{"op":"step","seq":1}` {
		t.Errorf("echo = op %d %q", op, msg)
	}
}

// TestRoundTripLargeBinary pushes a payload past the 16-bit length boundary — the
// case that matters, since every framebuffer does exactly that.
func TestRoundTripLargeBinary(t *testing.T) {
	srv := echoServer(t)
	cl := dial(t, srv.URL)

	// 320x240 RGBA plus a 16-byte header: the real thing.
	payload := make([]byte, 16+320*240*4)
	for i := range payload {
		payload[i] = byte(i * 7)
	}
	cl.send(t, OpBinary, payload)
	op, msg, err := readEcho(t, cl)
	if err != nil {
		t.Fatal(err)
	}
	if op != OpBinary {
		t.Fatalf("echo opcode = %d, want binary", op)
	}
	if !bytes.Equal(msg, payload) {
		t.Errorf("echoed %d bytes, want %d, equal=%v", len(msg), len(payload), bytes.Equal(msg, payload))
	}
}

// TestFragmentedMessage sends a message split across a frame plus continuations.
func TestFragmentedMessage(t *testing.T) {
	srv := echoServer(t)
	cl := dial(t, srv.URL)

	// FIN=0 text, then FIN=0 cont, then FIN=1 cont.
	frag := func(fin bool, opcode byte, s string) {
		var buf bytes.Buffer
		b0 := opcode
		if fin {
			b0 |= 0x80
		}
		buf.WriteByte(b0)
		buf.WriteByte(0x80 | byte(len(s)))
		key := [4]byte{1, 2, 3, 4}
		buf.Write(key[:])
		for i := 0; i < len(s); i++ {
			buf.WriteByte(s[i] ^ key[i&3])
		}
		if _, err := cl.c.Write(buf.Bytes()); err != nil {
			t.Fatal(err)
		}
	}
	frag(false, OpText, "he")
	frag(false, opCont, "ll")
	frag(true, opCont, "o")

	op, msg, err := readEcho(t, cl)
	if err != nil {
		t.Fatal(err)
	}
	if op != OpText || string(msg) != "hello" {
		t.Errorf("reassembled = %q, want %q", msg, "hello")
	}
}

// TestPingPong checks a ping is answered inside Read, transparently to the server.
func TestPingPong(t *testing.T) {
	srv := echoServer(t)
	cl := dial(t, srv.URL)

	cl.send(t, opPing, []byte("hi"))
	op, msg := cl.recv(t)
	if op != opPong || string(msg) != "hi" {
		t.Errorf("ping answered with op %d %q, want pong %q", op, msg, "hi")
	}
	// The connection still works afterwards.
	cl.send(t, OpText, []byte("still here"))
	if _, msg, err := readEcho(t, cl); err != nil || string(msg) != "still here" {
		t.Errorf("after ping: %q err=%v", msg, err)
	}
}

// TestCloseEndsRead checks a client close frame surfaces as ErrClosed.
func TestCloseEndsRead(t *testing.T) {
	closed := make(chan error, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrade(w, r)
		if err != nil {
			return
		}
		_, _, err = conn.Read()
		closed <- err
	}))
	defer srv.Close()

	cl := dial(t, srv.URL)
	cl.send(t, opClose, []byte{0x03, 0xE8})
	if err := <-closed; !errors.Is(err, ErrClosed) {
		t.Errorf("Read after close = %v, want ErrClosed", err)
	}
}

// TestUnmaskedClientFrameRejected: the spec requires client frames be masked.
func TestUnmaskedClientFrameRejected(t *testing.T) {
	failed := make(chan error, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrade(w, r)
		if err != nil {
			return
		}
		_, _, err = conn.Read()
		failed <- err
	}))
	defer srv.Close()

	cl := dial(t, srv.URL)
	cl.c.Write([]byte{0x81, 0x02, 'h', 'i'}) // FIN|text, len 2, no mask bit
	err := <-failed
	if err == nil || !strings.Contains(err.Error(), "unmasked") {
		t.Errorf("unmasked frame accepted: err = %v", err)
	}
}

// TestBadVersionRejected checks the handshake refuses anything but version 13.
func TestBadVersionRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Upgrade(w, r)
	}))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUpgradeRequired {
		t.Errorf("version 8 handshake = %d, want 426", resp.StatusCode)
	}
}

// readEcho reads one frame, skipping any control frames the server may interleave.
func readEcho(t *testing.T, cl *client) (byte, []byte, error) {
	t.Helper()
	for i := 0; i < 4; i++ {
		op, msg := cl.recv(t)
		if op == OpText || op == OpBinary {
			return op, msg, nil
		}
	}
	return 0, nil, errors.New("no data frame")
}
