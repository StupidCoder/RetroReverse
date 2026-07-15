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
	"reflect"
	"sort"
	"strings"
	"testing"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/n64adapter"
)

// ---- synthetic targets, so the protocol is testable without a ROM ----

// coreTarget implements debug.Target and nothing else — the honest minimum. It exists
// to prove that a platform which can only show a screen and read memory still works,
// and that the ops it cannot back are refused rather than half-served.
type coreTarget struct {
	steps int
	title string
}

const fakeW, fakeH = 8, 4

func (t *coreTarget) Platform() string { return "fake" }
func (t *coreTarget) Title() string {
	if t.title != "" {
		return t.title
	}
	return "Fake Game"
}
func (t *coreTarget) Snapshot() debug.Snapshot     { return fakeSnap{} }
func (t *coreTarget) Restore(debug.Snapshot) error { return nil }
func (t *coreTarget) Close() error                 { return nil }

func (t *coreTarget) CPU() debug.CPUReg {
	return debug.CPUReg{PC: 0x80001234, Names: []string{"r0", "at"}, Vals: []uint64{0, 1},
		Extra: map[string]uint64{"hi": 0xAB}}
}

func (t *coreTarget) ReadMem(addr uint32, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(int(addr) + i)
	}
	return b
}

func (t *coreTarget) Display() (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, fakeW, fakeH))
	for i := range img.Pix {
		img.Pix[i] = 0xFF // the fake's scanout is all-white, distinguishable from a draw target
	}
	return img, nil
}

type fakeSnap struct{}

func (fakeSnap) Platform() string { return "fake" }

// fakeTarget is a fully-featured target: it backs every capability the frame tools
// need, so the whole protocol can be exercised without booting a real machine.
type fakeTarget struct {
	coreTarget
	watches   []debug.Watch
	watchSink func(debug.WatchHit)
	bps       []uint64

	// touched is every stylus state the machine was actually handed, in order — which is
	// the thing worth asserting, because the runner does not hand over everything it is
	// sent, and must not.
	touched []debug.Touch

	// keyed is every keyboard event the machine was handed, in order. Unlike touch, the
	// runner must hand over EVERY one — a make and a break both matter — so this is
	// asserted to be the exact sequence sent, with nothing dropped or collapsed.
	keyed []debug.Key
}

func (f *fakeTarget) Key(k debug.Key) error {
	f.keyed = append(f.keyed, k)
	return nil
}

func (f *fakeTarget) TouchPanels() []debug.TouchPanel {
	return []debug.TouchPanel{{ID: "bottom", Name: "Touch screen", X: 0, Y: 2, W: fakeW, H: 2}}
}

func (f *fakeTarget) Touch(t debug.Touch) error {
	if t.Panel != "bottom" {
		return debug.ErrUnsupported
	}
	f.touched = append(f.touched, t)
	return nil
}

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
	// Only three of this frame's 128 commands ever put a pixel anywhere — which is the whole
	// point of the writer list, and roughly the ratio a real machine has.
	fc.CountWrites()
	for i := 0; i < fakeW*fakeH; i++ {
		fc.MarkWrite(i % 3)
	}
	if withOverdraw {
		fc.Overdraw = map[int][]debug.PixelWrite{
			0: {{CmdIndex: 0, R: 1, G: 2, B: 3, A: 4}, {CmdIndex: 2, R: 5, Rejected: true}},
		}
	}
	return fc, nil
}

func (f *fakeTarget) StepFast() error { f.steps++; return nil }

// RenderAfter paints every pixel byte with k, so a test can prove the right k came back.
func (f *fakeTarget) RenderAfter(fc *debug.FrameCapture, k int) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, fakeW, fakeH))
	for i := range img.Pix {
		img.Pix[i] = byte(k)
	}
	return img, nil
}

func (f *fakeTarget) StepInstr(n int) (debug.StopReason, error) {
	return debug.StopReason{Kind: "steps", PC: 0x1000, Steps: uint64(n)}, nil
}

func (f *fakeTarget) Continue(budget uint64) (debug.StopReason, error) {
	return debug.StopReason{Kind: "breakpoint", PC: 0x2000, Steps: 42}, nil
}

func (f *fakeTarget) Disasm(addr uint64, n int) ([]debug.Instr, error) {
	out := make([]debug.Instr, n)
	for i := range out {
		out[i] = debug.Instr{Addr: addr + uint64(i)*4, Text: "nop"}
	}
	return out, nil
}

func (f *fakeTarget) SetBreakpoint(pc uint64)   { f.bps = append(f.bps, pc) }
func (f *fakeTarget) ClearBreakpoint(pc uint64) { f.bps = nil }
func (f *fakeTarget) Breakpoints() []uint64     { return f.bps }

func (f *fakeTarget) SetWatch(w debug.Watch) (int, error) {
	w.ID = len(f.watches) + 1
	f.watches = append(f.watches, w)
	return w.ID, nil
}
func (f *fakeTarget) ClearWatch(id int)                    { f.watches = nil }
func (f *fakeTarget) Watches() []debug.Watch               { return f.watches }
func (f *fakeTarget) OnWatchHit(sink func(debug.WatchHit)) { f.watchSink = sink }

// Savestates and the resume line: a real file, and a command line in this "platform's"
// own flag vocabulary — which is the whole point of the handoff.
func (f *fakeTarget) SaveStateFile(path string) error {
	return os.WriteFile(path, []byte("fake savestate"), 0o644)
}

func (f *fakeTarget) LoadStateFile(path string) error {
	_, err := os.ReadFile(path)
	return err
}

func (f *fakeTarget) ResumeArgs(statePath string) []string {
	return []string{"go", "run", "./cmd/bootoracle", "-image", "game.z64", "-loadstate", statePath}
}

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

// send writes an op with its args, as the page does.
func (cl *wsClient) send(t *testing.T, op string, seq int, args any) {
	t.Helper()
	var raw json.RawMessage
	if args != nil {
		b, err := json.Marshal(args)
		if err != nil {
			t.Fatal(err)
		}
		raw = b
	}
	payload, _ := json.Marshal(req{Op: op, Seq: seq, Args: raw})

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

// recvJSON reads the next text message.
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

// recvType reads text messages until one of the wanted type arrives, so a test is not
// derailed by an event (a watch hit, a stop) landing in the middle of a reply.
func (cl *wsClient) recvType(t *testing.T, want string) map[string]any {
	t.Helper()
	for i := 0; i < 16; i++ {
		m := cl.recvJSON(t)
		if m["type"] == want {
			return m
		}
		if m["type"] == "error" {
			t.Fatalf("waiting for %q, got error: %v", want, m["msg"])
		}
	}
	t.Fatalf("no %q message", want)
	return nil
}

// recvBinary reads the next binary message and unpacks its header.
func (cl *wsClient) recvBinary(t *testing.T) (kind byte, seq int, w, h int, payload []byte) {
	t.Helper()
	for {
		isText, b := cl.recv(t)
		if isText {
			continue
		}
		if len(b) < hdrSize || string(b[:4]) != "RDB2" {
			t.Fatalf("bad binary header %x", b[:min(8, len(b))])
		}
		return b[4],
			int(binary.LittleEndian.Uint16(b[6:])),
			int(binary.LittleEndian.Uint32(b[12:])),
			int(binary.LittleEndian.Uint32(b[16:])),
			b[hdrSize:]
	}
}

// serveTarget serves one target with no library behind it, so these tests do not go
// looking for the repository's real games/ tree (and do not change behaviour depending
// on what happens to be checked out).
func serveTarget(t *testing.T, tgt debug.Target) *httptest.Server {
	t.Helper()
	ws := newWorkspace("")
	ws.setTarget(newRunner(tgt), nil)
	srv := httptest.NewServer((&Server{ws: ws}).Handler())
	t.Cleanup(srv.Close)
	return srv
}

func serveFake(t *testing.T) (*httptest.Server, *fakeTarget) {
	t.Helper()
	f := &fakeTarget{}
	return serveTarget(t, f), f
}

// ---- capabilities ----

// TestCapabilitiesAdvertised: the page builds itself from this list, so it has to be
// exactly what the target can do — no more, no less.
func TestCapabilitiesAdvertised(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)

	m := cl.recvType(t, "hello")
	if m["platform"] != "fake" || m["title"] != "Fake Game" {
		t.Errorf("hello = %v", m)
	}
	var got []string
	for _, c := range m["caps"].([]any) {
		got = append(got, c.(string))
	}
	sort.Strings(got)
	want := []string{"break", "code", "disasm", "faststep", "frames", "keys", "replay", "resume", "states", "touch", "watch"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("caps = %v, want %v", got, want)
	}
}

// ---- input ----

// TestTouchPanels: the page cannot know where the stylus can reach — the target says, in
// the coordinates of the frame the page is already drawing.
func TestTouchPanels(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "input.panels", 1, nil)
	m := cl.recvType(t, "touchpanels")
	panels := m["panels"].([]any)
	if len(panels) != 1 {
		t.Fatalf("panels = %v, want one", panels)
	}
	p := panels[0].(map[string]any)
	if p["id"] != "bottom" || p["y"].(float64) != 2 || p["w"].(float64) != fakeW {
		t.Errorf("panel = %v", p)
	}
}

// TestTouchLatch is the point of queueing input rather than poking it straight in.
//
// A game does not ask whether the pen is down, it asks whether the pen WENT down since it
// last looked — so a state the machine never publishes a sample for is a state the game
// never sees. A mouse click is over in tens of milliseconds; a 3DS frame takes about a
// hundred. Apply the events as they arrive and the press and the release both land between
// two frames, and the tap simply never happened.
//
// So: at most one state per frame, and every state gets a frame. A drag's moves — same
// state, new position — collapse, because within one frame only where the pen ended up
// matters.
func TestTouchLatch(t *testing.T) {
	f := &fakeTarget{}
	rn := &Runner{tgt: f, caps: map[string]bool{debug.CapTouch: true}, mb: newMailbox()}
	defer rn.mb.close()

	// A tap: down, a drag across the panel, and a release — all arriving before the
	// machine has advanced a single frame, which is exactly what a real click does.
	rn.touchQ = []debug.Touch{
		{Panel: "bottom", X: 10, Y: 10, Down: true},
		{Panel: "bottom", X: 20, Y: 20, Down: true},
		{Panel: "bottom", X: 30, Y: 30, Down: true},
		{Panel: "bottom", X: 30, Y: 30, Down: false},
	}

	rn.applyTouch()
	if len(f.touched) != 1 || !f.touched[0].Down || f.touched[0].X != 30 {
		t.Fatalf("first frame saw %+v, want the pen down at the end of the drag (30,30)", f.touched)
	}
	rn.applyTouch()
	if len(f.touched) != 2 || f.touched[1].Down {
		t.Fatalf("second frame saw %+v, want the pen lifted", f.touched)
	}
	// The press and the release landed on different frames, so the game got both edges.
	rn.applyTouch()
	if len(f.touched) != 2 {
		t.Errorf("an empty queue handed the machine something: %+v", f.touched)
	}
}

// TestKeyDelivered is the counterpart to TestTouchLatch, and its opposite: where touch
// collapses same-state events to the latest, keys must NOT — a press and its release are
// a make and a break scancode, and dropping or coalescing either desyncs the game's key
// state. So every key event is handed to the machine, in the order it arrived.
func TestKeyDelivered(t *testing.T) {
	f := &fakeTarget{}
	rn := &Runner{tgt: f, caps: map[string]bool{debug.CapKeys: true}, mb: newMailbox()}
	defer rn.mb.close()

	// A tap on Up, then Enter: four make/break events, none of which may be dropped —
	// including the two presses of different keys and the two releases.
	events := []keyArgs{
		{Name: "up", Down: true},
		{Name: "up", Down: false},
		{Name: "enter", Down: true},
		{Name: "enter", Down: false},
	}
	for _, e := range events {
		if err := rn.key(request{req: req{Op: "input.key", Args: mustJSON(e)}}); err != nil {
			t.Fatalf("key %+v: %v", e, err)
		}
	}
	if len(f.keyed) != len(events) {
		t.Fatalf("machine saw %d keys, want %d (%+v)", len(f.keyed), len(events), f.keyed)
	}
	for i, e := range events {
		if f.keyed[i].Name != e.Name || f.keyed[i].Down != e.Down {
			t.Errorf("key %d = %+v, want %+v", i, f.keyed[i], e)
		}
	}

	// The mailbox must never coalesce a key the way it coalesces a scrub — that is what
	// would silently drop a make or a break in transit.
	if coalescable("input.key") {
		t.Error("input.key is coalescable; a make or break could be dropped")
	}
}

// TestCoreOnlyTarget: a target that can only show a screen and read memory must still
// serve, and must *refuse* the ops it cannot back rather than crashing or faking them.
func TestCoreOnlyTarget(t *testing.T) {
	srv := serveTarget(t, &coreTarget{})
	cl := dial(t, srv.URL)

	m := cl.recvType(t, "hello")
	if caps := m["caps"]; caps != nil && len(caps.([]any)) != 0 {
		t.Errorf("a core-only target advertised capabilities: %v", caps)
	}

	// What it can do, it does.
	cl.send(t, "cpu.regs", 1, nil)
	if m := cl.recvType(t, "cpu"); m["pc"] != "0000000080001234" {
		t.Errorf("cpu = %v", m)
	}
	cl.send(t, "mem.read", 2, memArgs{Addr: 0x10, Len: 4})
	if m := cl.recvType(t, "mem"); m["data"] != "10111213" {
		t.Errorf("mem = %v", m)
	}
	cl.send(t, "frame.display", 3, nil)
	cl.recvType(t, "render")
	if kind, _, _, _, _ := cl.recvBinary(t); kind != kindImage {
		t.Errorf("display kind = %d", kind)
	}

	// What it cannot do, it refuses — naming the capability, not panicking.
	for _, op := range []string{
		"frame.step", "frame.scrub", "cpu.step", "cpu.disasm",
		"mem.watch", "state.save", "surface.list", "surface.render",
	} {
		cl.send(t, op, 9, nil)
		m := cl.recvJSON(t)
		if m["type"] != "error" {
			t.Errorf("%s on a core-only target = %v, want an error", op, m)
			continue
		}
		if !strings.Contains(m["msg"].(string), "cannot") {
			t.Errorf("%s error is unhelpful: %v", op, m["msg"])
		}
	}
}

// ---- frames ----

func TestStepAndProvenance(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "frame.step", 1, stepArgs{Overdraw: true})
	m := cl.recvType(t, "frame")
	if got := len(m["commands"].([]any)); got != fakeCmds {
		t.Errorf("frame carried %d commands, want %d", got, fakeCmds)
	}
	if m["hasProv"] != true || m["overdraw"] != true {
		t.Errorf("frame flags = %v", m)
	}

	kind, seq, w, h, payload := cl.recvBinary(t)
	if kind != kindProv || seq != 1 || w != fakeW || h != fakeH {
		t.Fatalf("prov header = kind %d seq %d %dx%d", kind, seq, w, h)
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
	cl.recvType(t, "hello")

	cl.send(t, "frame.step", 1, stepArgs{})
	cl.recvType(t, "frame")
	cl.recvBinary(t)

	cl.send(t, "frame.scrub", 7, scrubArgs{K: 2})
	m := cl.recvType(t, "render")
	if int(m["k"].(float64)) != 2 || int(m["seq"].(float64)) != 7 {
		t.Fatalf("render reply = %v", m)
	}
	kind, seq, w, h, payload := cl.recvBinary(t)
	if kind != kindImage || seq != 7 || w != fakeW || h != fakeH {
		t.Fatalf("image header = kind %d seq %d %dx%d", kind, seq, w, h)
	}
	// The fake paints every byte with k, so this proves k reached RenderAfter.
	for i, b := range payload {
		if b != 2 {
			t.Fatalf("pixel byte %d = %d, want 2 (the requested k)", i, b)
		}
	}

	// Out-of-range k clamps to the last command rather than failing.
	cl.send(t, "frame.scrub", 8, scrubArgs{K: 9999})
	if m := cl.recvType(t, "render"); int(m["k"].(float64)) != fakeCmds-1 {
		t.Errorf("k=9999 clamped to %v, want %d", m["k"], fakeCmds-1)
	}
}

func TestPixelOverdraw(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "frame.step", 1, stepArgs{Overdraw: true})
	cl.recvType(t, "frame")
	cl.recvBinary(t)

	cl.send(t, "frame.pixel", 2, pixelArgs{X: 0, Y: 0})
	m := cl.recvType(t, "pixel")
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

func TestPlayStreamsAndPauses(t *testing.T) {
	srv, f := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "frame.play", 1, playArgs{On: true})
	for i := 0; i < 3; i++ {
		m := cl.recvType(t, "render")
		if m["play"] != true || int(m["k"].(float64)) != -1 {
			t.Fatalf("play frame %d = %v", i, m)
		}
		_, seq, _, _, payload := cl.recvBinary(t)
		if seq != 0 {
			t.Errorf("a played frame answers no request, so seq should be 0, got %d", seq)
		}
		// The fake's Display paints 0xFF: this is the scanout, not a draw target.
		if payload[0] != 0xFF {
			t.Errorf("played frame is not the scanout (first byte %#x)", payload[0])
		}
		// Nothing more arrives until we acknowledge — the server holds itself to one
		// frame in flight so the emulator cannot bury a slow page.
		cl.send(t, "frame.ack", 0, nil)
	}
	if f.steps < 3 {
		t.Errorf("machine advanced %d fields over 3 played frames", f.steps)
	}

	// Pausing captures a drawn field in full, so the pause alone leaves you on a frame
	// worth tracing. Frames already on their way arrive first, so skip past them.
	cl.send(t, "frame.play", 2, playArgs{On: false, Overdraw: true})
	for i := 0; ; i++ {
		if i > 8 {
			t.Fatal("no frame capture after pause")
		}
		m := cl.recvJSON(t)
		if m["type"] == "frame" {
			if len(m["commands"].([]any)) != fakeCmds {
				t.Errorf("pause captured %v commands", len(m["commands"].([]any)))
			}
			break
		}
		if m["type"] == "render" {
			cl.recvBinary(t)
		}
	}
}

// ---- cpu, breakpoints, watches ----

func TestCPUStepAndDisasm(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "cpu.step", 1, cpuRunArgs{N: 5})
	m := cl.recvType(t, "stopped")
	if m["reason"] != "steps" || int(m["steps"].(float64)) != 5 {
		t.Errorf("stopped = %v", m)
	}

	cl.send(t, "bp.set", 2, bpArgs{PC: 0x80002000})
	cl.recvType(t, "ok")

	cl.send(t, "cpu.disasm", 3, disasmArgs{Addr: 0x80001000, N: 4})
	m = cl.recvType(t, "disasm")
	if got := len(m["instr"].([]any)); got != 4 {
		t.Errorf("disasm returned %d instructions, want 4", got)
	}
	bps := m["bps"].([]any)
	if len(bps) != 1 || bps[0] != "0000000080002000" {
		t.Errorf("disasm breakpoints = %v", bps)
	}
}

// TestContinueStopsAndReports: a run reports why it ended, as an event.
func TestContinueStopsAndReports(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "cpu.continue", 1, nil)
	cl.recvType(t, "ok")
	// The fake's Continue reports a breakpoint immediately, so the run loop stops
	// itself and broadcasts why.
	m := cl.recvType(t, "stopped")
	if m["reason"] != "breakpoint" || m["pc"] != "0000000000002000" {
		t.Errorf("stopped = %v", m)
	}
}

func TestWatchSetAndCleared(t *testing.T) {
	srv, f := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "mem.watch", 1, watchArgs{Kind: "write", Lo: 0x100, Hi: 0x200, Break: true})
	m := cl.recvType(t, "watches")
	ws := m["watches"].([]any)
	if len(ws) != 1 {
		t.Fatalf("watches = %v", ws)
	}
	w := ws[0].(map[string]any)
	if w["kind"] != "write" || w["lo"] != "00000100" || w["break"] != true {
		t.Errorf("watch = %v", w)
	}

	// A hit fired by the emulator reaches the page as an event.
	f.watchSink(debug.WatchHit{ID: 1, Kind: "write", Addr: 0x180, Val: 0x2A, PC: 0x8000ABCD, Instr: "sw t0, 0(a1)"})
	hit := cl.recvType(t, "hit")
	if hit["addr"] != "00000180" || hit["pc"] != "8000abcd" || hit["instr"] != "sw t0, 0(a1)" {
		t.Errorf("hit = %v", hit)
	}

	cl.send(t, "mem.unwatch", 2, unwatchArgs{ID: 1})
	if m := cl.recvType(t, "watches"); len(m["watches"].([]any)) != 0 {
		t.Errorf("watch survived unwatch: %v", m)
	}
}

// ---- plumbing ----

func TestUnknownOpIsAnError(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "nonsense", 5, nil)
	m := cl.recvType(t, "error")
	if int(m["seq"].(float64)) != 5 {
		t.Errorf("reply = %v, want an error echoing seq 5", m)
	}
}

// TestMailboxCoalesces: a scrubber drag must collapse to its newest position, while
// everything else keeps its order.
func TestMailboxCoalesces(t *testing.T) {
	mb := newMailbox()
	s1 := &session{}
	mb.push(request{from: s1, req: req{Op: "frame.step", Seq: 1}})
	mb.push(request{from: s1, req: req{Op: "frame.scrub", Seq: 2}})
	mb.push(request{from: s1, req: req{Op: "cpu.regs", Seq: 3}})
	mb.push(request{from: s1, req: req{Op: "frame.scrub", Seq: 4}}) // replaces seq 2, keeping its slot
	mb.push(request{from: s1, req: req{Op: "frame.scrub", Seq: 5}}) // replaces seq 4

	want := []struct {
		op  string
		seq int
	}{{"frame.step", 1}, {"frame.scrub", 5}, {"cpu.regs", 3}}
	for i, w := range want {
		got, ok := mb.pop()
		if !ok {
			t.Fatalf("queue drained early at %d", i)
		}
		if got.req.Op != w.op || got.req.Seq != w.seq {
			t.Errorf("pop %d = %s/%d, want %s/%d", i, got.req.Op, got.req.Seq, w.op, w.seq)
		}
	}

	// Two tabs scrubbing are two questions; neither may eat the other.
	s2 := &session{}
	mb.push(request{from: s1, req: req{Op: "frame.scrub", Seq: 6}})
	mb.push(request{from: s2, req: req{Op: "frame.scrub", Seq: 7}})
	a, _ := mb.pop()
	b, _ := mb.pop()
	if a.from != s1 || b.from != s2 {
		t.Errorf("a scrub from one session displaced another session's")
	}

	mb.close()
	if _, ok := mb.pop(); ok {
		t.Error("pop after close returned a request")
	}
}

func TestCrossOriginRejected(t *testing.T) {
	srv, _ := serveFake(t)
	r, _ := http.NewRequest("GET", srv.URL+"/ws", nil)
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	r.Header.Set("Sec-WebSocket-Version", "13")
	r.Header.Set("Origin", "http://evil.example")
	resp, err := http.DefaultClient.Do(r)
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
	// A module that fails to load shows up as a blank page and nothing else, so every one
	// the page imports is checked here.
	for _, asset := range []string{
		"app.js", "conn.js", "store.js", "util.js", "style.css",
		"panels/registry.js", "panels/dock.js", "panels/viewport.js", "panels/commands.js",
		"panels/inspect.js", "panels/surface.js", "panels/files.js", "panels/states.js",
	} {
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
	srv := serveTarget(t, a)
	cl := dial(t, srv.URL)
	if m := cl.recvType(t, "hello"); m["platform"] != "n64" {
		t.Fatalf("hello = %v", m)
	}

	cl.send(t, "frame.step", 1, stepArgs{})
	frame := cl.recvType(t, "frame")
	if got := len(frame["commands"].([]any)); got != len(refFC.Commands) {
		t.Errorf("served %d commands, headless captured %d", got, len(refFC.Commands))
	}
	cl.recvBinary(t) // prov

	cl.send(t, "frame.scrub", 2, scrubArgs{K: k})
	cl.recvType(t, "render")
	kind, _, w, h, payload := cl.recvBinary(t)
	if kind != kindImage {
		t.Fatalf("kind = %d", kind)
	}
	if w != want.Rect.Dx() || h != want.Rect.Dy() {
		t.Fatalf("served %dx%d, headless rendered %dx%d", w, h, want.Rect.Dx(), want.Rect.Dy())
	}
	if !bytes.Equal(payload, want.Pix) {
		t.Errorf("the served image differs from RenderAfter(%d)", k)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// The scrubber walks the commands that wrote pixels, so the frame message has to carry
// them. The fake's frame runs 128 commands and three of them ever put a pixel anywhere.
func TestFrameCarriesItsWriters(t *testing.T) {
	srv, _ := serveFake(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "hello")

	cl.send(t, "frame.step", 1, stepArgs{N: 1})
	m := cl.recvType(t, "frame")

	if len(m["commands"].([]any)) != fakeCmds {
		t.Fatalf("commands = %d", len(m["commands"].([]any)))
	}
	w, ok := m["writers"].([]any)
	if !ok {
		t.Fatalf("writers = %v, want the list of commands that drew", m["writers"])
	}
	got := make([]int, len(w))
	for i, v := range w {
		got[i] = int(v.(float64))
	}
	if want := []int{0, 1, 2}; !reflect.DeepEqual(got, want) {
		t.Errorf("writers = %v, want %v — the whole stream is 128 commands and only these drew", got, want)
	}
}

// ---- the profile ----

// profTarget draws in every other field, which is what a 30 Hz game on a 60 Hz console
// does — and the reason a field's profile is not a frame's.
type profTarget struct {
	coreTarget
	field int
}

func (p *profTarget) FrameProfile() debug.FrameProfile {
	p.field++
	if p.field%2 == 0 { // the logic-only field: no draws, but real time spent
		return debug.FrameProfile{
			TotalMs:  2,
			Drew:     false,
			Buckets:  []debug.ProfileBucket{{Name: "arm11", Millis: 2}},
			Counters: []debug.ProfileCounter{{Name: "draws", Value: 0}},
		}
	}
	return debug.FrameProfile{
		TotalMs: 100,
		Drew:    true,
		Buckets: []debug.ProfileBucket{
			{Name: "rasterise", Millis: 80, Count: 4},
			{Name: "arm11", Millis: 20},
		},
		Counters: []debug.ProfileCounter{{Name: "draws", Value: 4}},
	}
}

// A field a game drew nothing in is not a frame. Its time belongs to the picture it was
// working towards, so it is carried and added to the next field that draws — otherwise the
// panel flips between the drawing field's cost and the idle one's, twice per frame, and
// neither number is what the frame cost.
func TestProfileFoldsIdleFieldsIntoTheFrameTheyBelongTo(t *testing.T) {
	tgt := &profTarget{}
	var pa profAccum

	if pa.emit() != nil {
		t.Error("a profile before any field has been observed")
	}

	pa.observe(tgt) // field 1: drew
	p := pa.emit()
	if p == nil {
		t.Fatal("no profile after a field that drew")
	}
	if p.TotalMs != 100 {
		t.Errorf("first frame = %v ms, want 100", p.TotalMs)
	}

	pa.observe(tgt) // field 2: drew nothing — not a frame, and not published as one
	if got := pa.emit(); got != p || got.TotalMs != 100 {
		t.Errorf("an idle field replaced the frame's profile: %v", got)
	}

	pa.observe(tgt) // field 3: drew — the idle field's time lands here
	p = pa.emit()
	if p.TotalMs != 102 {
		t.Errorf("frame = %v ms, want 102: the idle field's 2 ms is part of what this picture cost", p.TotalMs)
	}
	var sum float64
	for _, b := range p.Buckets {
		sum += b.Millis
	}
	if sum != p.TotalMs {
		t.Errorf("buckets sum to %v but the frame took %v — the bar would not add up", sum, p.TotalMs)
	}
	if len(p.Buckets) != 2 {
		t.Fatalf("buckets = %v, want them merged by name", p.Buckets)
	}
	// arm11 appears in both fields and must be added, not replaced or duplicated.
	for _, b := range p.Buckets {
		if b.Name == "arm11" && b.Millis != 22 {
			t.Errorf("arm11 = %v ms, want 22 (20 drawing + 2 idle)", b.Millis)
		}
		if b.Name == "rasterise" && b.Count != 4 {
			t.Errorf("rasterise count = %d, want 4", b.Count)
		}
	}
	if p.Counters[0].Value != 4 {
		t.Errorf("draws = %d, want 4", p.Counters[0].Value)
	}

	// A state load lands in another machine; the fields we were adding up were on the way
	// to a picture that is never going to be drawn.
	pa.reset()
	if pa.emit() != nil {
		t.Error("the profile survived a reset")
	}
}
