package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/n64adapter"
)

// ---- a synthetic target, so the protocol is testable without a ROM ----

type fakeTarget struct {
	steps int
}

func (f *fakeTarget) Name() string                 { return "Fake" }
func (f *fakeTarget) Snapshot() debug.Snapshot     { return fakeSnap{} }
func (f *fakeTarget) Restore(debug.Snapshot) error { return nil }

type fakeSnap struct{}

func (fakeSnap) Platform() string { return "fake" }

const fakeW, fakeH = 8, 4

// fakeCmds is over the 100-command bar the "step to a drawn frame" heuristic uses to
// tell a frame that renders a scene from an idle field.
const fakeCmds = 128

// StepFrame builds a capture whose provenance is trivially predictable: pixel i was
// written by command i%3.
func (f *fakeTarget) StepFrame(withOverdraw bool) (*debug.FrameCapture, error) {
	f.steps++
	fc := &debug.FrameCapture{
		Start:  fakeSnap{},
		Width:  fakeW,
		Height: fakeH,
		Prov:   make([]int32, fakeW*fakeH),
	}
	for i := 0; i < fakeCmds; i++ {
		fc.Commands = append(fc.Commands, debug.GPUCommand{
			Index: i, Name: "Fake_Cmd", Op: uint32(i),
			Words: []uint64{0xDEADBEEFCAFEF00D + uint64(i)},
		})
	}
	for i := range fc.Prov {
		fc.Prov[i] = int32(i % 3)
	}
	if withOverdraw {
		fc.Overdraw = map[int][]debug.PixelWrite{
			0: {{CmdIndex: 0, R: 1, G: 2, B: 3, A: 4}, {CmdIndex: 2, R: 5, Rejected: true}},
		}
	}
	return fc, nil
}

// RenderAfter paints every pixel with k, so a test can prove the right k came back.
func (f *fakeTarget) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, fakeW, fakeH))
	for i := range img.Pix {
		img.Pix[i] = byte(k)
	}
	return img, nil
}

func (f *fakeTarget) CPU() debug.CPUReg {
	return debug.CPUReg{PC: 0x80001234, Names: []string{"r0", "at"}, Vals: []uint64{0, 1},
		Extra: map[string]uint64{"hi": 0xAB}}
}
func (f *fakeTarget) Display() (*image.RGBA, error) { return f.RenderAfter(nil, 0xFF) }
func (f *fakeTarget) ReadMem(addr uint32, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(int(addr) + i)
	}
	return b
}
func (f *fakeTarget) SetWriteWatch(lo, hi uint32, cb func(a, v, pc uint32)) {}
func (f *fakeTarget) SetReadWatch(lo, hi uint32, cb func(a, v, pc uint32))  {}
func (f *fakeTarget) SetBreakpoint(vaddr uint64)                            {}

// ---- a minimal websocket client, enough to drive the server ----

type wsClient struct {
	c  net.Conn
	br *bufio.Reader
}

func dial(t *testing.T, url string) *wsClient {
	t.Helper()
	addr := strings.TrimPrefix(url, "http://")
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(c, "GET /ws HTTP/1.1\r\nHost: "+addr+"\r\n"+
		"Upgrade: websocket\r\nConnection: Upgrade\r\n"+
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n")
	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake = %d, want 101", resp.StatusCode)
	}
	t.Cleanup(func() { c.Close() })
	return &wsClient{c: c, br: br}
}

func (cl *wsClient) sendJSON(t *testing.T, v any) {
	t.Helper()
	payload, _ := json.Marshal(v)
	var buf bytes.Buffer
	buf.WriteByte(0x81) // FIN | text
	n := len(payload)
	switch {
	case n < 126:
		buf.WriteByte(0x80 | byte(n))
	default:
		buf.WriteByte(0x80 | 126)
		binary.Write(&buf, binary.BigEndian, uint16(n))
	}
	key := [4]byte{9, 9, 9, 9}
	buf.Write(key[:])
	for i, b := range payload {
		buf.WriteByte(b ^ key[i&3])
	}
	if _, err := cl.c.Write(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
}

func (cl *wsClient) recv(t *testing.T) (isText bool, payload []byte) {
	t.Helper()
	var h [2]byte
	if _, err := io.ReadFull(cl.br, h[:]); err != nil {
		t.Fatal(err)
	}
	isText = h[0]&0x0F == 1
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
	return isText, payload
}

// recvJSON reads the next text message and decodes it into a map.
func (cl *wsClient) recvJSON(t *testing.T) map[string]any {
	t.Helper()
	for {
		isText, b := cl.recv(t)
		if !isText {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("bad JSON %q: %v", b, err)
		}
		return m
	}
}

// recvBinary reads the next binary message and unpacks its header.
func (cl *wsClient) recvBinary(t *testing.T) (kind byte, seq, w, h int, payload []byte) {
	t.Helper()
	for {
		isText, b := cl.recv(t)
		if isText {
			continue
		}
		if len(b) < hdrSize || string(b[:4]) != "RDB1" {
			t.Fatalf("bad binary header %x", b[:min(8, len(b))])
		}
		return b[4],
			int(binary.LittleEndian.Uint16(b[6:])),
			int(binary.LittleEndian.Uint32(b[8:])),
			int(binary.LittleEndian.Uint32(b[12:])),
			b[hdrSize:]
	}
}

func serveFake(t *testing.T) (*httptest.Server, *fakeTarget) {
	t.Helper()
	f := &fakeTarget{}
	srv := httptest.NewServer(New(f, "fake.z64").Handler())
	t.Cleanup(srv.Close)
	return srv, f
}

// ---- tests ----

func TestHelloAndStep(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)

	if m := cl.recvJSON(t); m["type"] != "hello" || m["rom"] != "fake.z64" {
		t.Fatalf("first message = %v, want hello", m)
	}

	cl.sendJSON(t, map[string]any{"op": "step", "seq": 1, "overdraw": true})
	m := cl.recvJSON(t)
	if m["type"] != "frame" {
		t.Fatalf("step reply = %v, want frame", m)
	}
	if got := m["w"].(float64); int(got) != fakeW {
		t.Errorf("frame w = %v, want %d", got, fakeW)
	}
	if got := len(m["commands"].([]any)); got != fakeCmds {
		t.Errorf("frame carried %d commands, want %d", got, fakeCmds)
	}
	if m["hasProv"] != true || m["overdraw"] != true {
		t.Errorf("frame flags = %v", m)
	}

	// The provenance buffer follows the frame JSON as a binary message.
	kind, seq, w, h, payload := cl.recvBinary(t)
	if kind != kindProv {
		t.Fatalf("binary kind = %d, want prov", kind)
	}
	if seq != 1 || w != fakeW || h != fakeH {
		t.Errorf("prov header = seq %d %dx%d", seq, w, h)
	}
	if len(payload) != fakeW*fakeH*4 {
		t.Fatalf("prov payload = %d bytes, want %d", len(payload), fakeW*fakeH*4)
	}
	for i := 0; i < fakeW*fakeH; i++ {
		got := int32(binary.LittleEndian.Uint32(payload[i*4:]))
		if want := int32(i % 3); got != want {
			t.Fatalf("prov[%d] = %d, want %d", i, got, want)
		}
	}
}

func TestScrubReturnsRequestedCommand(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvJSON(t) // hello

	cl.sendJSON(t, map[string]any{"op": "step", "seq": 1})
	cl.recvJSON(t)   // frame
	cl.recvBinary(t) // prov

	cl.sendJSON(t, map[string]any{"op": "scrub", "seq": 7, "k": 2})
	m := cl.recvJSON(t)
	if m["type"] != "render" || int(m["k"].(float64)) != 2 || int(m["seq"].(float64)) != 7 {
		t.Fatalf("render reply = %v", m)
	}
	kind, seq, w, h, payload := cl.recvBinary(t)
	if kind != kindImage || seq != 7 || w != fakeW || h != fakeH {
		t.Fatalf("image header = kind %d seq %d %dx%d", kind, seq, w, h)
	}
	if len(payload) != fakeW*fakeH*4 {
		t.Fatalf("image payload = %d bytes, want %d", len(payload), fakeW*fakeH*4)
	}
	// The fake paints every byte with k, so this proves k reached RenderAfter.
	for i, b := range payload {
		if b != 2 {
			t.Fatalf("pixel byte %d = %d, want 2 (the requested k)", i, b)
		}
	}

	// Out-of-range k clamps to the last command rather than failing.
	cl.sendJSON(t, map[string]any{"op": "scrub", "seq": 8, "k": 9999})
	if m := cl.recvJSON(t); int(m["k"].(float64)) != fakeCmds-1 {
		t.Errorf("k=9999 clamped to %v, want %d", m["k"], fakeCmds-1)
	}
}

func TestPixelOverdraw(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvJSON(t)

	cl.sendJSON(t, map[string]any{"op": "step", "seq": 1, "overdraw": true})
	cl.recvJSON(t)
	cl.recvBinary(t)

	cl.sendJSON(t, map[string]any{"op": "pixel", "seq": 2, "x": 0, "y": 0})
	m := cl.recvJSON(t)
	if m["type"] != "pixel" {
		t.Fatalf("reply = %v", m)
	}
	if int(m["cmd"].(float64)) != 0 {
		t.Errorf("pixel (0,0) last writer = %v, want 0", m["cmd"])
	}
	writes := m["writes"].([]any)
	if len(writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(writes))
	}
	if w := writes[1].(map[string]any); w["rejected"] != true || int(w["cmd"].(float64)) != 2 {
		t.Errorf("second write = %v, want cmd 2 rejected", w)
	}
}

func TestCPUAndMem(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvJSON(t)

	cl.sendJSON(t, map[string]any{"op": "cpu", "seq": 3})
	m := cl.recvJSON(t)
	if m["type"] != "cpu" || m["pc"] != "0000000080001234" {
		t.Errorf("cpu = %v", m)
	}

	cl.sendJSON(t, map[string]any{"op": "mem", "seq": 4, "addr": 0x10, "len": 4})
	m = cl.recvJSON(t)
	if m["type"] != "mem" || m["data"] != "10111213" {
		t.Errorf("mem = %v", m)
	}
}

// TestPlayStreamsAndPauses: play free-runs the machine streaming the scanout, holds
// itself to one unacknowledged frame, and pausing lands on a full capture.
func TestPlayStreamsAndPauses(t *testing.T) {
	srv, f := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvJSON(t) // hello

	cl.sendJSON(t, map[string]any{"op": "play", "seq": 1, "on": true})

	for i := 0; i < 3; i++ {
		m := cl.recvJSON(t)
		if m["type"] != "render" || m["play"] != true {
			t.Fatalf("play frame %d = %v", i, m)
		}
		if int(m["k"].(float64)) != -1 {
			t.Errorf("a played frame is the scanout, so k should be -1, got %v", m["k"])
		}
		kind, seq, _, _, payload := cl.recvBinary(t)
		if kind != kindImage || seq != 0 {
			t.Fatalf("play binary = kind %d seq %d, want an image at seq 0", kind, seq)
		}
		// The fake's Display paints 0xFF, so this is the scanout and not a draw target.
		if payload[0] != 0xFF {
			t.Errorf("played frame is not the scanout (first byte %#x)", payload[0])
		}
		// Nothing more arrives until we acknowledge: the server keeps one frame in
		// flight so the emulator cannot bury a slow page in frames it will never draw.
		cl.sendJSON(t, map[string]any{"op": "ack"})
	}

	stepsWhilePlaying := f.steps
	if stepsWhilePlaying < 3 {
		t.Errorf("machine advanced %d fields over 3 played frames", stepsWhilePlaying)
	}

	// Pausing captures the next field in full, so you land on a real frame. The
	// machine keeps playing until the request is picked up, so a frame already on its
	// way arrives first; skip past those to the capture.
	cl.sendJSON(t, map[string]any{"op": "play", "seq": 2, "on": false, "overdraw": true})
	var m map[string]any
	for i := 0; ; i++ {
		if i > 8 {
			t.Fatal("no frame capture after pause")
		}
		m = cl.recvJSON(t)
		if m["type"] == "frame" {
			break
		}
		if m["type"] != "render" || m["play"] != true {
			t.Fatalf("pause reply = %v, want a frame capture", m)
		}
		cl.recvBinary(t) // the played frame's image
	}
	if len(m["commands"].([]any)) != fakeCmds {
		t.Errorf("pause captured %v commands", len(m["commands"].([]any)))
	}
	if m["overdraw"] != true {
		t.Error("pause did not honour the overdraw request")
	}
	cl.recvBinary(t) // the provenance buffer
}

// TestPlayUsesTheFastPath: play must not pay for a capture per field.
func TestPlayUsesTheFastPath(t *testing.T) {
	var fs fastStepper = (*fakeFast)(nil)
	_ = fs // the fake below is the compile-time proof; the real one is n64adapter

	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvJSON(t)

	// The fake target has no StepFast, so the session must fall back to StepFrame
	// rather than refuse to play.
	cl.sendJSON(t, map[string]any{"op": "play", "seq": 1, "on": true})
	if m := cl.recvJSON(t); m["type"] != "render" {
		t.Fatalf("a target without StepFast could not play: %v", m)
	}
}

// fakeFast exists only to assert the optional interface is satisfiable as declared.
type fakeFast struct{}

func (*fakeFast) StepFast() error { return nil }

func TestUnknownOpIsAnError(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvJSON(t)

	cl.sendJSON(t, map[string]any{"op": "nonsense", "seq": 5})
	m := cl.recvJSON(t)
	if m["type"] != "error" || int(m["seq"].(float64)) != 5 {
		t.Errorf("reply = %v, want an error echoing seq 5", m)
	}
}

// TestMailboxCoalesces: a scrubber drag must collapse to its newest position, while
// everything else keeps its order.
func TestMailboxCoalesces(t *testing.T) {
	mb := newMailbox()
	mb.push(req{Op: "step", Seq: 1})
	mb.push(req{Op: "scrub", Seq: 2, K: 10})
	mb.push(req{Op: "mem", Seq: 3})
	mb.push(req{Op: "scrub", Seq: 4, K: 20}) // replaces seq 2, keeping its slot
	mb.push(req{Op: "scrub", Seq: 5, K: 30}) // replaces seq 4

	want := []req{{Op: "step", Seq: 1}, {Op: "scrub", Seq: 5, K: 30}, {Op: "mem", Seq: 3}}
	for i, w := range want {
		got, ok := mb.pop()
		if !ok {
			t.Fatalf("queue drained early at %d", i)
		}
		if got.Op != w.Op || got.Seq != w.Seq || got.K != w.K {
			t.Errorf("pop %d = %+v, want %+v", i, got, w)
		}
	}
	mb.close()
	if _, ok := mb.pop(); ok {
		t.Error("pop after close returned a request")
	}
}

func TestCrossOriginRejected(t *testing.T) {
	srv, _ := serveFake(t)
	req, _ := http.NewRequest("GET", srv.URL+"/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Origin", "http://evil.example")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("cross-origin upgrade = %d, want 403", resp.StatusCode)
	}
}

func TestServesThePage(t *testing.T) {
	srv, _ := serveFake(t)
	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || !bytes.Contains(body, []byte("framedbg")) {
		t.Errorf("GET / = %d, %d bytes", resp.StatusCode, len(body))
	}
	for _, asset := range []string{"app.js", "conn.js", "viewport.js", "commands.js", "panels.js", "style.css"} {
		r, err := http.Get(srv.URL + "/" + asset)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("GET /%s = %d", asset, r.StatusCode)
		}
	}
}

// TestScrubMatchesRenderAfter pins the served image to the already-verified headless
// path: what the socket hands the browser for command k must be exactly what
// RenderAfter produces for command k on the same frame.
func TestScrubMatchesRenderAfter(t *testing.T) {
	const rom = "../../../games/pilotwings-64-n64/image/Pilotwings 64 (USA).z64"
	if _, err := os.Stat(rom); err != nil {
		t.Skip("Pilotwings ROM not present; skipping the end-to-end check")
	}

	// The reference: the headless path, exactly as framedbg drives it.
	ref, err := n64adapter.New(rom)
	if err != nil {
		t.Fatal(err)
	}
	var refFC *debug.FrameCapture
	for i := 0; i < 800; i++ {
		fc, err := ref.StepFrame(false)
		if err != nil {
			t.Fatal(err)
		}
		if len(fc.Commands) > 100 && fc.Prov != nil {
			refFC = fc
			break
		}
	}
	if refFC == nil {
		t.Fatal("no drawn frame")
	}
	const k = 40
	want, err := ref.RenderAfter(refFC, k)
	if err != nil {
		t.Fatal(err)
	}

	// The same frame, over the wire.
	a, err := n64adapter.New(rom)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(New(a, "pilotwings.z64").Handler())
	defer srv.Close()
	cl := dial(t, srv.URL)
	cl.recvJSON(t) // hello

	cl.sendJSON(t, map[string]any{"op": "step", "seq": 1})
	frame := cl.recvJSON(t)
	if frame["type"] != "frame" {
		t.Fatalf("step reply = %v", frame)
	}
	if got := len(frame["commands"].([]any)); got != len(refFC.Commands) {
		t.Errorf("served %d commands, headless captured %d", got, len(refFC.Commands))
	}
	cl.recvBinary(t) // prov

	cl.sendJSON(t, map[string]any{"op": "scrub", "seq": 2, "k": k})
	cl.recvJSON(t) // render
	kind, _, w, h, payload := cl.recvBinary(t)
	if kind != kindImage {
		t.Fatalf("kind = %d", kind)
	}
	if w != want.Rect.Dx() || h != want.Rect.Dy() {
		t.Fatalf("served %dx%d, headless rendered %dx%d", w, h, want.Rect.Dx(), want.Rect.Dy())
	}
	if !bytes.Equal(payload, want.Pix) {
		t.Errorf("the served image differs from RenderAfter(%d) — %d vs %d bytes",
			k, len(payload), len(want.Pix))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
