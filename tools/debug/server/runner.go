package server

import (
	"encoding/hex"
	"fmt"
	"image"
	"sort"
	"sync"
	"time"

	"retroreverse.com/tools/debug"
)

// A Runner owns a debug.Target and is the only goroutine that ever touches it — the
// single-threaded contract the whole debug package rests on.
//
// Sessions (browser connections) do not own the machine; they attach to the runner,
// submit requests and receive replies and events. So reloading the page does not
// throw the machine away, and two tabs can watch the same run.
//
// The loop has three modes. Paused, it blocks for a request. Playing, it runs a frame
// and then picks up whatever arrived. Running the CPU, it executes a bounded slice and
// then picks up whatever arrived — which is also how "stop" works, with no flag
// written from another goroutine and so no race with the emulator.
type Runner struct {
	tgt  debug.Target
	caps map[string]bool

	mb *mailbox

	mu   sync.Mutex // guards subs only
	subs []*session

	// Frame state.
	fc       *debug.FrameCapture
	frameNo  int
	playing  bool
	inFlight bool // a played frame the page has not acknowledged yet

	// CPU run state.
	running    bool
	lastStream time.Time // last scanout pushed while free-running the CPU (throttle)

	// prof folds fields into frames — see profAccum.
	prof profAccum

	// blank is the replay's "start from a black screen" setting, as the page last set it.
	blank bool

	// touchQ holds stylus states the page has sent but the machine has not seen yet.
	// They are handed over one per frame — see applyTouch.
	touchQ []debug.Touch

	// cache holds draw targets already replayed for the current capture, so nudging
	// the scrubber back over ground already covered is free. Every RenderAfter is a
	// full replay from the frame-start snapshot, which is what makes this worth it.
	cache map[int]*image.RGBA
	order []int
}

const (
	cacheMax = 24 // ~7 MB at 320x240; a scrubber drag rarely revisits more

	// runSlice bounds one pass of a CPU run, so the loop comes back to look at the
	// mailbox. Small enough to feel responsive, big enough not to spin.
	runSliceInstrs = 2_000_000
)

func newRunner(tgt debug.Target) *Runner {
	caps := map[string]bool{}
	for _, c := range debug.Capabilities(tgt) {
		caps[c] = true
	}
	rn := &Runner{tgt: tgt, caps: caps, mb: newMailbox(), cache: map[int]*image.RGBA{}}
	if w, ok := tgt.(debug.Watcher); ok {
		// A watch fires inside the emulator, on this goroutine. Broadcasting from
		// here is safe: it only writes to sockets.
		w.OnWatchHit(func(h debug.WatchHit) {
			rn.broadcast(hitMsg{
				Type: "hit", ID: h.ID, Kind: h.Kind,
				Addr: hex32(h.Addr), Val: hex32(h.Val), PC: hex32(h.PC), Instr: h.Instr,
			})
		})
	}
	go rn.loop()
	return rn
}

func (rn *Runner) attach(s *session) {
	rn.mu.Lock()
	rn.subs = append(rn.subs, s)
	rn.mu.Unlock()
	s.send(helloMsg{
		Type:     "hello",
		Platform: rn.tgt.Platform(),
		Title:    rn.tgt.Title(),
		Caps:     debug.Capabilities(rn.tgt),
	})
}

func (rn *Runner) detach(s *session) {
	rn.mu.Lock()
	for i, x := range rn.subs {
		if x == s {
			rn.subs = append(rn.subs[:i], rn.subs[i+1:]...)
			break
		}
	}
	rn.mu.Unlock()
}

// broadcast sends an event to every attached session.
func (rn *Runner) broadcast(v any) {
	rn.mu.Lock()
	subs := append([]*session(nil), rn.subs...)
	rn.mu.Unlock()
	for _, s := range subs {
		s.send(v)
	}
}

func (rn *Runner) submit(s *session, r req) { rn.mb.push(request{from: s, req: r}) }

// shutdown ends the runner. The loop closes the target on its own goroutine, so the
// machine is never freed underneath a run in progress.
func (rn *Runner) shutdown() { rn.mb.close() }

func (rn *Runner) loop() {
	for {
		// A shutdown must be able to interrupt a machine that is free-running or
		// running the CPU, not only one that is sitting waiting for a request.
		if rn.mb.isClosed() {
			rn.tgt.Close()
			return
		}

		// Playing: run a frame, then pick up whatever arrived. Blocks instead when a
		// frame is in flight — the next thing to come is its acknowledgement.
		if rn.playing && !rn.inFlight {
			if err := rn.playFrame(); err != nil {
				rn.playing = false
				rn.broadcast(errMsg{Type: "error", Msg: err.Error()})
			}
			// A halted core does not stop the video clock, so play mode would happily
			// step fields forever against a machine that has permanently stopped. Say so
			// and stop, exactly as a breakpoint would.
			rn.stopIfHalted()
			rn.drain()
			continue
		}

		// Running the CPU: one bounded slice at a time, so a "break" request is
		// picked up promptly and the emulator is never touched from another
		// goroutine.
		if rn.running {
			cs := rn.tgt.(debug.CodeStepper)
			sr, err := cs.Continue(runSliceInstrs)
			if err != nil {
				rn.running = false
				rn.broadcast(errMsg{Type: "error", Msg: err.Error()})
			} else if sr.Kind != "budget" {
				rn.running = false
				rn.broadcast(stopped(sr))
			}
			rn.drain()
			// A CPU-run target with no frame stepper (the DOS/go32 host) never plays,
			// so nothing else keeps its screen alive while it runs. Stream its scanout
			// between slices — throttled, and paced by the page's acknowledgement the
			// same way play mode is — so you can watch the picture change as the CPU
			// runs (and as injected keys drive a menu).
			if rn.running {
				rn.streamScanout()
			}
			continue
		}

		r, ok := rn.mb.pop()
		if !ok {
			rn.tgt.Close()
			return
		}
		rn.dispatch(r)
	}
}

// drain handles everything already queued, without blocking.
func (rn *Runner) drain() {
	for {
		r, ok := rn.mb.tryPop()
		if !ok {
			return
		}
		rn.dispatch(r)
	}
}

func (rn *Runner) dispatch(r request) {
	if err := rn.handle(r); err != nil {
		r.from.send(errMsg{Type: "error", Seq: r.req.Seq, Msg: err.Error()})
	}
}

func (rn *Runner) handle(r request) error {
	// A request for a capability this target does not have is a bug in the page, not
	// something to half-answer.
	if need, ok := opNeeds[r.req.Op]; ok && !rn.caps[need] {
		return fmt.Errorf("%s: this target cannot %s", r.req.Op, need)
	}
	switch r.req.Op {
	case "frame.step":
		return rn.step(r)
	case "frame.play":
		return rn.play(r)
	case "frame.ack":
		rn.inFlight = false
		return nil
	case "frame.scrub":
		return rn.scrub(r)
	case "frame.display":
		return rn.display(r)
	case "frame.pixel":
		return rn.pixel(r)
	case "input.panels":
		rn.sendTouchPanels(r)
		return nil
	case "input.touch":
		return rn.touch(r)
	case "input.key":
		return rn.key(r)
	case "cpu.regs":
		rn.sendCPU(r)
		return nil
	case "cpu.step":
		return rn.cpuStep(r)
	case "cpu.continue":
		rn.running = true
		rn.inFlight = false // a fresh run: no streamed frame is outstanding
		r.from.send(okMsg{Type: "ok", Seq: r.req.Seq, Op: r.req.Op, Note: "running"})
		return nil
	case "cpu.break":
		rn.running = false
		r.from.send(okMsg{Type: "ok", Seq: r.req.Seq, Op: r.req.Op, Note: "paused"})
		return nil
	case "cpu.disasm":
		return rn.disasm(r)
	case "bp.set", "bp.clear":
		return rn.breakpoint(r)
	case "surface.list":
		rn.sendSurfaces(r)
		return nil
	case "surface.render":
		return rn.surface(r)
	case "fs.list":
		return rn.fsList(r)
	case "fs.read":
		return rn.fsRead(r)
	case "mem.read":
		return rn.mem(r)
	case "mem.watch":
		return rn.watch(r)
	case "mem.unwatch":
		return rn.unwatch(r)
	case "state.save":
		return rn.stateSave(r)
	case "state.load":
		return rn.stateLoad(r)
	default:
		return fmt.Errorf("unknown op %q", r.req.Op)
	}
}

// opNeeds maps an op to the capability it requires, so an op a target cannot back is
// refused in one place rather than crashing on a type assertion in ten.
var opNeeds = map[string]string{
	"frame.step":     debug.CapFrames,
	"frame.play":     debug.CapFrames,
	"frame.scrub":    debug.CapReplay,
	"frame.pixel":    debug.CapFrames,
	"cpu.step":       debug.CapCode,
	"cpu.continue":   debug.CapCode,
	"cpu.disasm":     debug.CapDisasm,
	"bp.set":         debug.CapBreak,
	"bp.clear":       debug.CapBreak,
	"mem.watch":      debug.CapWatch,
	"mem.unwatch":    debug.CapWatch,
	"state.save":     debug.CapStates,
	"state.load":     debug.CapStates,
	"surface.list":   debug.CapSurfaces,
	"surface.render": debug.CapSurfaces,
	"fs.list":        debug.CapFiles,
	"fs.read":        debug.CapFiles,
	"input.panels":   debug.CapTouch,
	"input.touch":    debug.CapTouch,
	"input.key":      debug.CapKeys,
}

// ---- frames ----

func (rn *Runner) step(r request) error {
	var a stepArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	fs := rn.tgt.(debug.FrameStepper)
	start := time.Now()
	rn.applyTouch()

	var fc *debug.FrameCapture
	var err error
	if a.N > 0 {
		for i := 0; i < a.N; i++ {
			fc, err = fs.StepFrame(a.Overdraw)
			if err != nil {
				return err
			}
			rn.frameNo++
			rn.prof.observe(rn.tgt)
			if halted, _ := rn.halted(); halted {
				break // the remaining fields would retire against a stopped core
			}
		}
	} else {
		fc, err = rn.stepToDrawn(fs, a.Overdraw)
		if err != nil {
			return err
		}
	}
	rn.fc = fc
	rn.clearCache()

	r.from.send(frameMsg{
		Type: "frame", Seq: r.req.Seq, Frame: rn.frameNo,
		Commands: toJSONCommands(fc.Commands),
		W:        fc.Width, H: fc.Height,
		HasProv:  fc.Prov != nil,
		Overdraw: fc.Overdraw != nil,
		Writers:  fc.Writers,
		StepMs:   msSince(start),
		Profile:  rn.prof.emit(),
	})
	// The provenance buffer goes over once per capture. From then on the page answers
	// "which command drew this pixel?" locally, with no round trip — which is what
	// makes hover-to-identify and the command-coverage overlay instant.
	if fc.Prov != nil {
		r.from.sendBinary(encodeProv(r.req.Seq, fc))
	}
	// Stepping a halted machine returns the same frame forever. Answer the step, then say
	// why it did not move.
	rn.stopIfHalted()
	return nil
}

// stepToDrawn steps until a frame actually renders a scene, so a fresh boot lands on
// something worth inspecting rather than an idle field.
func (rn *Runner) stepToDrawn(fs debug.FrameStepper, withOverdraw bool) (*debug.FrameCapture, error) {
	for i := 0; i < 800; i++ {
		fc, err := fs.StepFrame(withOverdraw)
		if err != nil {
			return nil, err
		}
		rn.frameNo++
		rn.prof.observe(rn.tgt)
		if fc.Drawn() {
			return fc, nil
		}
		// A halted core will never draw again, so spending the whole field budget to
		// report "no drawn frame" would name the symptom and bury the cause.
		if halted, why := rn.halted(); halted {
			return nil, fmt.Errorf("the machine halted: %s", why)
		}
	}
	return nil, fmt.Errorf("no drawn frame within the field budget")
}

// halted asks a target that can tell whether its core has stopped for good. A target
// without the capability is never halted as far as the runner is concerned — the same
// posture as before it existed.
func (rn *Runner) halted() (bool, string) {
	h, ok := rn.tgt.(debug.Haltable)
	if !ok {
		return false, ""
	}
	return h.Halted()
}

// stopIfHalted ends a free run whose machine has halted, and says why on the channel the
// page already uses for a breakpoint. Reported once: the flag stays set forever after, and
// a stop notice per field would be its own kind of noise.
func (rn *Runner) stopIfHalted() bool {
	halted, why := rn.halted()
	if !halted {
		return false
	}
	rn.playing, rn.inFlight = false, false
	rn.broadcast(stopped(debug.StopReason{
		Kind: "halted", PC: rn.tgt.CPU().PC, Note: why,
	}))
	return true
}

// play starts or stops free-running the machine. While it runs nothing is captured:
// each field is stepped with no snapshot and no census, and the page is sent the
// scanout. Stopping does a full capture step, so you land on a real frame — the next
// field, not the last one played, which left no record behind to go back to.
//
// The capture is a step-to-drawn, not a single field. Pausing lands wherever the game
// happens to be, and that is as likely to be a field it spent setting up as one it drew
// in — so capturing exactly one field would hand back an empty command list about as
// often as not, and leave you pressing Step to get something to look at. Pause means
// stop AND give me a frame I can trace, so it steps until the machine draws one.
func (rn *Runner) play(r request) error {
	var a playArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	if a.On {
		rn.playing, rn.inFlight = true, false
		return nil
	}
	if !rn.playing {
		return nil
	}
	rn.playing, rn.inFlight = false, false
	return rn.step(request{from: r.from, req: req{
		Op: "frame.step", Seq: r.req.Seq, Args: mustJSON(stepArgs{N: 0, Overdraw: overdrawOf(r.req)}),
	}})
}

func (rn *Runner) playFrame() error {
	start := time.Now()
	rn.applyTouch()
	rn.frameNo++
	if f, ok := rn.tgt.(debug.FastStepper); ok {
		if err := f.StepFast(); err != nil {
			return err
		}
	} else if _, err := rn.tgt.(debug.FrameStepper).StepFrame(false); err != nil {
		return err
	}
	rn.prof.observe(rn.tgt)
	img, err := rn.tgt.Display()
	if err != nil {
		// Early in a boot nothing is being scanned out yet. Keep running — this is
		// exactly the stretch play mode exists to get through.
		return nil
	}
	payload := encodeImage(0, streamMain, img)
	rn.broadcast(renderMsg{
		Type: "render", Seq: 0, Stream: streamMain, K: -1, Play: true,
		Frame: rn.frameNo, RenderMs: msSince(start), Bytes: len(payload),
		Profile: rn.prof.emit(), // free-running frames keep the profile panel live
	})
	rn.inFlight = true
	rn.broadcastBinary(payload)
	return nil
}

// streamScanout pushes the machine's current screen to the page while the CPU
// free-runs, for a target that has no frame stepper of its own (the DOS/go32 host).
// It reuses play mode's machinery: the image goes out as a play frame, and the page
// acknowledges it, so the runner holds at most one unacknowledged frame and the
// stream paces itself to what the page can paint. It is throttled in wall-time too,
// because a CPU slice can be far shorter than a frame and there is no point sending
// the same 256 KB image sixty times between two paints.
//
// A target that can already frame-step (and therefore play) does not use this — its
// screen is kept alive by play mode — so this only fires for a CPU-only machine.
func (rn *Runner) streamScanout() {
	if rn.inFlight || rn.caps[debug.CapFrames] {
		return
	}
	if !rn.lastStream.IsZero() && time.Since(rn.lastStream) < 33*time.Millisecond {
		return
	}
	img, err := rn.tgt.Display()
	if err != nil {
		return // nothing on screen yet — a cold boot, not a fault
	}
	rn.lastStream = time.Now()
	payload := encodeImage(0, streamMain, img)
	rn.broadcast(renderMsg{
		Type: "render", Seq: 0, Stream: streamMain, K: -1, Play: true,
		Frame: int(rn.tgt.CPU().PC), Bytes: len(payload),
	})
	rn.inFlight = true
	rn.broadcastBinary(payload)
}

func (rn *Runner) scrub(r request) error {
	var a scrubArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	if rn.fc == nil {
		return fmt.Errorf("no captured frame yet")
	}
	// "Start the frame blank" is a property of the replay, not of one position in it, so it
	// is set on the target rather than threaded through RenderAfter — and it invalidates the
	// cache, whose images were rendered under the other setting.
	if b, ok := rn.tgt.(debug.BlankReplayer); ok && a.Blank != rn.blank {
		rn.blank = a.Blank
		b.SetReplayBlank(a.Blank)
		rn.clearCache()
	}
	k := a.K
	if k < 0 {
		k = 0
	}
	if n := len(rn.fc.Commands); k >= n {
		k = n - 1
	}
	start := time.Now()
	img, cached := rn.cache[k]
	if !cached {
		var err error
		if img, err = rn.replay(k); err != nil {
			return fmt.Errorf("render after command %d: %w", k, err)
		}
	}
	payload := encodeImage(r.req.Seq, streamMain, img)
	r.from.send(renderMsg{
		Type: "render", Seq: r.req.Seq, Stream: streamMain, K: k,
		RenderMs: msSince(start), Bytes: len(payload), Cached: cached,
	})
	r.from.sendBinary(payload)
	return nil
}

// replay renders the frame as it stood after command k — and, on a target that can
// replay several points at once, renders the positions the drag is about to land on at
// the same time.
//
// A scrub is a drag: the page sends a stream of positions and only the last one
// matters, but the ones in between are exactly where it is about to ask next. Each
// replay is an independent restore-and-run on a disposable machine, so a batch of them
// costs about what one of them costs, and the rest of the drag is then served out of
// the cache with no round trip to the emulator at all.
func (rn *Runner) replay(k int) (*image.RGBA, error) {
	rp, ok := rn.tgt.(debug.BatchReplayer)
	if !ok {
		img, err := rn.tgt.(debug.FrameReplayer).RenderAfter(rn.fc, k)
		if err != nil {
			return nil, err
		}
		rn.put(k, img)
		return img, nil
	}

	// k first, then a spread of nearby positions that are not already cached. The spread is
	// proportional to the list, so it is a fraction of the drag rather than a fixed number
	// of steps — which on a 100,000-command frame would be nothing.
	//
	// The spread runs over the frame's WRITERS, because those are the scrubber's positions.
	// Spreading over the command stream instead would prefetch register writes the drag is
	// never going to stop on, and the cache would miss on every one that it does.
	stops := rn.fc.Writers
	if len(stops) == 0 {
		stops = nil // no writer list: the page is scrubbing the whole stream
	}
	n := len(rn.fc.Commands)
	at := k
	if stops != nil {
		n = len(stops)
		at = sort.SearchInts(stops, k) // where the drag is, in the positions it can stop at
	}
	ks := []int{k}
	step := n / 64
	if step < 1 {
		step = 1
	}
	for _, d := range []int{1, -1, 2, -2, 3, -3, 4} {
		p := at + d*step
		if p < 0 || p >= n {
			continue
		}
		if stops != nil {
			p = stops[p]
		}
		if _, hit := rn.cache[p]; hit || p == k {
			continue
		}
		ks = append(ks, p)
	}

	imgs, err := rp.RenderAfterBatch(rn.fc, ks)
	if err != nil {
		return nil, err
	}
	for i, img := range imgs {
		if img != nil {
			rn.put(ks[i], img)
		}
	}
	return imgs[0], nil
}

func (rn *Runner) display(r request) error {
	start := time.Now()
	img, err := rn.tgt.Display()
	if err != nil {
		return err
	}
	payload := encodeImage(r.req.Seq, streamMain, img)
	r.from.send(renderMsg{
		Type: "render", Seq: r.req.Seq, Stream: streamMain, K: -1,
		RenderMs: msSince(start), Bytes: len(payload),
	})
	r.from.sendBinary(payload)
	return nil
}

// pixel answers the overdraw question. The page already knows the last writer from its
// copy of the provenance buffer; what it cannot know is the full write history.
func (rn *Runner) pixel(r request) error {
	var a pixelArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	if rn.fc == nil {
		return fmt.Errorf("no captured frame yet")
	}
	fc := rn.fc
	m := pixelMsg{Type: "pixel", Seq: r.req.Seq, X: a.X, Y: a.Y, Cmd: fc.ProvAt(a.X, a.Y), Writes: []jsonWrite{}}
	if fc.Overdraw != nil && a.X >= 0 && a.Y >= 0 && a.X < fc.Width && a.Y < fc.Height {
		for _, w := range fc.Overdraw[a.Y*fc.Width+a.X] {
			name := ""
			if w.CmdIndex >= 0 && w.CmdIndex < len(fc.Commands) {
				name = fc.Commands[w.CmdIndex].Name
			}
			m.Writes = append(m.Writes, jsonWrite{
				Cmd: w.CmdIndex, Name: name,
				R: w.R, G: w.G, B: w.B, A: w.A, Rejected: w.Rejected,
			})
		}
	}
	r.from.send(m)
	return nil
}

// ---- input ----
//
// The debugger does not only watch a machine, it plays it: while the machine free-runs, a
// drag on the picture is a drag of the stylus on the panel under it.
//
// Input is queued rather than poked straight into the machine, and the reason is the press
// EDGE. A game does not ask "is the pen down", it asks "did the pen go down since I last
// looked" — so a state the machine never publishes a sample for is a state the game never
// sees. A mouse click is over in tens of milliseconds and the 3DS oracle spends about a
// hundred on a frame, so a press applied and released between two frames would be a press
// that never happened. applyTouch is the fix: the machine is handed at most one state per
// frame, and every state gets a frame.

func (rn *Runner) sendTouchPanels(r request) {
	m := touchPanelsMsg{Type: "touchpanels", Seq: r.req.Seq, Panels: []jsonTouchPanel{}}
	for _, p := range rn.tgt.(debug.Toucher).TouchPanels() {
		m.Panels = append(m.Panels, jsonTouchPanel{
			ID: p.ID, Name: p.Name, X: p.X, Y: p.Y, W: p.W, H: p.H,
		})
	}
	r.from.send(m)
}

func (rn *Runner) touch(r request) error {
	var a touchArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	rn.touchQ = append(rn.touchQ, debug.Touch{Panel: a.Panel, X: a.X, Y: a.Y, Down: a.Down})
	return nil
}

// applyTouch hands the machine at most one stylus state, and is called once per frame
// advance — played or captured — so no state can be skipped over.
//
// A run of same-state events collapses into its latest: a drag fires an event per
// mouse-move, and within one frame only where the pen ended up matters. A change of state
// does not collapse, so a press and the release that follows it always land on different
// frames, and the press is one the game can actually poll.
func (rn *Runner) applyTouch() {
	if len(rn.touchQ) == 0 {
		return
	}
	t := rn.touchQ[0]
	rn.touchQ = rn.touchQ[1:]
	for len(rn.touchQ) > 0 && rn.touchQ[0].Down == t.Down {
		t = rn.touchQ[0]
		rn.touchQ = rn.touchQ[1:]
	}
	if err := rn.tgt.(debug.Toucher).Touch(t); err != nil {
		rn.broadcast(errMsg{Type: "error", Msg: err.Error()})
	}
}

// key delivers one keyboard event to the machine. Unlike a touch it is NOT queued
// for one-per-frame release: a key press and its release are a make and a break
// scancode, and dropping or coalescing either desyncs the game's key state — so
// every event is handed straight to the target, in the order it arrived, and the
// target's own input path paces delivery (a PC keyboard ISR runs with interrupts
// masked until it returns, which serialises make from break on its own). This runs
// on the runner goroutine — the only one that touches the target — so a direct call
// is safe, and a CPU run picks the queued scancodes up on its next slice.
func (rn *Runner) key(r request) error {
	var a keyArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	return rn.tgt.(debug.Keyer).Key(debug.Key{Name: a.Name, Code: a.Code, Down: a.Down})
}

// ---- cpu ----

func (rn *Runner) sendCPU(r request) {
	c := rn.tgt.CPU()
	m := cpuMsg{
		Type: "cpu", Seq: r.req.Seq,
		PC: hex64(c.PC), Names: c.Names,
		Vals: make([]string, len(c.Vals)), Extra: map[string]string{},
	}
	for i, v := range c.Vals {
		m.Vals[i] = hex64(v)
	}
	for k, v := range c.Extra {
		m.Extra[k] = hex64(v)
	}
	r.from.send(m)
}

func (rn *Runner) cpuStep(r request) error {
	var a cpuRunArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	sr, err := rn.tgt.(debug.CodeStepper).StepInstr(a.N)
	if err != nil {
		return err
	}
	rn.broadcast(stopped(sr))
	return nil
}

func (rn *Runner) disasm(r request) error {
	var a disasmArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	if a.N <= 0 {
		a.N = 32
	}
	addr := a.Addr
	if addr == 0 {
		addr = rn.tgt.CPU().PC
	}
	ins, err := rn.tgt.(debug.Disassembler).Disasm(addr, a.N)
	if err != nil {
		return err
	}
	m := disasmMsg{Type: "disasm", Seq: r.req.Seq, PC: hex64(rn.tgt.CPU().PC)}
	for _, in := range ins {
		m.Instr = append(m.Instr, jsonInstr{Addr: hex64(in.Addr), Text: in.Text})
	}
	if b, ok := rn.tgt.(debug.Breakpointer); ok {
		for _, pc := range b.Breakpoints() {
			m.BPs = append(m.BPs, hex64(pc))
		}
	}
	r.from.send(m)
	return nil
}

func (rn *Runner) breakpoint(r request) error {
	var a bpArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	b := rn.tgt.(debug.Breakpointer)
	if r.req.Op == "bp.set" {
		b.SetBreakpoint(a.PC)
	} else {
		b.ClearBreakpoint(a.PC)
	}
	r.from.send(okMsg{Type: "ok", Seq: r.req.Seq, Op: r.req.Op})
	return nil
}

// ---- surfaces ----
//
// A surface is memory drawn as a picture: the scanout, the buffer being drawn into, all
// of VRAM, or an address read as a texture. The last is the one that earns its keep —
// point it at where a DMA just landed and see whether what arrived is the texture the
// game meant to upload.
//
// Surface images go out on their own stream, so the viewport and this panel can both
// have an image in flight without either one drawing the other's.

func (rn *Runner) sendSurfaces(r request) {
	m := surfacesMsg{Type: "surfaces", Seq: r.req.Seq, Surfaces: []jsonSurface{}}
	for _, s := range rn.tgt.(debug.Surfacer).Surfaces() {
		m.Surfaces = append(m.Surfaces, jsonSurface{
			ID: s.ID, Name: s.Name, Free: s.Free, W: s.W, H: s.H, Formats: s.Formats,
		})
	}
	r.from.send(m)
}

func (rn *Runner) surface(r request) error {
	var a surfaceArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	start := time.Now()
	img, err := rn.tgt.(debug.Surfacer).RenderSurface(a.ID, debug.View{
		Addr: a.Addr, W: a.W, H: a.H, Stride: a.Stride, Format: a.Format, Palette: a.Palette,
	})
	if err != nil {
		return err
	}
	payload := encodeImage(r.req.Seq, streamSurface, img)
	r.from.send(renderMsg{
		Type: "render", Seq: r.req.Seq, Stream: streamSurface, K: -1,
		RenderMs: msSince(start), Bytes: len(payload),
	})
	r.from.sendBinary(payload)
	return nil
}

// ---- the game's own filesystem ----
//
// A cartridge has none; a disc has one, and the game reads it as it runs. So this is
// where you go to see what the machine could be loading — and, paired with a read-watch
// on the drive, what it is loading right now.

// fileHead caps how much of a file is sent. The browser is for identifying a file, not
// for extracting it — the extract tools already do that properly.
const fileHead = 4096

func (rn *Runner) fsList(r request) error {
	var a fsArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	if a.Path == "" {
		a.Path = "/"
	}
	entries, err := rn.tgt.(debug.FileLister).ListDir(a.Path)
	if err != nil {
		return err
	}
	m := filesMsg{Type: "files", Seq: r.req.Seq, Path: a.Path, Entries: []jsonFile{}}
	for _, e := range entries {
		m.Entries = append(m.Entries, jsonFile{
			Name: e.Name, Path: e.Path, Dir: e.Dir, Size: e.Size, Offset: e.Offset,
		})
	}
	r.from.send(m)
	return nil
}

func (rn *Runner) fsRead(r request) error {
	var a fsArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	b, err := rn.tgt.(debug.FileLister).ReadFile(a.Path)
	if err != nil {
		return err
	}
	head := b
	if len(head) > fileHead {
		head = head[:fileHead]
	}
	r.from.send(fileMsg{
		Type: "file", Seq: r.req.Seq, Path: a.Path,
		Size: int64(len(b)), Data: hex.EncodeToString(head),
	})
	return nil
}

// ---- memory ----

func (rn *Runner) mem(r request) error {
	var a memArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	n := a.Len
	if n <= 0 {
		n = 256
	}
	if n > 64*1024 {
		n = 64 * 1024
	}
	m := memMsg{
		Type: "mem", Seq: r.req.Seq, Addr: a.Addr,
		Data: hex.EncodeToString(rn.tgt.ReadMem(a.Addr, n)),
	}
	if mm, ok := rn.tgt.(debug.MemoryMapper); ok {
		for _, rg := range mm.Regions() {
			m.Regions = append(m.Regions, jsonRegion{Name: rg.Name, Lo: hex32(rg.Lo), Hi: hex32(rg.Hi)})
		}
	}
	r.from.send(m)
	return nil
}

func (rn *Runner) watch(r request) error {
	var a watchArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	w := rn.tgt.(debug.Watcher)
	if _, err := w.SetWatch(debug.Watch{Kind: a.Kind, Lo: a.Lo, Hi: a.Hi, Break: a.Break}); err != nil {
		return err
	}
	rn.sendWatches(r)
	return nil
}

func (rn *Runner) unwatch(r request) error {
	var a unwatchArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	rn.tgt.(debug.Watcher).ClearWatch(a.ID)
	rn.sendWatches(r)
	return nil
}

func (rn *Runner) sendWatches(r request) {
	m := watchesMsg{Type: "watches", Seq: r.req.Seq, Watches: []jsonWatch{}}
	for _, w := range rn.tgt.(debug.Watcher).Watches() {
		m.Watches = append(m.Watches, jsonWatch{
			ID: w.ID, Kind: w.Kind, Lo: hex32(w.Lo), Hi: hex32(w.Hi), Break: w.Break, Hits: w.Hits,
		})
	}
	rn.broadcast(m) // every tab should see the watch set change
}

// ---- savestates ----

func (rn *Runner) stateSave(r request) error {
	var a stateArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	if err := rn.tgt.(debug.StateFiler).SaveStateFile(a.Path); err != nil {
		return err
	}
	r.from.send(okMsg{Type: "ok", Seq: r.req.Seq, Op: r.req.Op, Note: a.Path})
	return nil
}

func (rn *Runner) stateLoad(r request) error {
	var a stateArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	if err := rn.tgt.(debug.StateFiler).LoadStateFile(a.Path); err != nil {
		return err
	}
	// The machine is somewhere else entirely now; the captured frame belongs to the
	// state we left, and so does the frame the profiler was adding fields up towards.
	rn.fc = nil
	rn.clearCache()
	rn.prof.reset()
	r.from.send(okMsg{Type: "ok", Seq: r.req.Seq, Op: r.req.Op, Note: a.Path})
	return nil
}

// ---- the replay cache ----

func (rn *Runner) put(k int, img *image.RGBA) {
	if len(rn.order) >= cacheMax {
		delete(rn.cache, rn.order[0])
		rn.order = rn.order[1:]
	}
	rn.cache[k] = img
	rn.order = append(rn.order, k)
}

func (rn *Runner) clearCache() {
	rn.cache = map[int]*image.RGBA{}
	rn.order = rn.order[:0]
}

func (rn *Runner) broadcastBinary(b []byte) {
	rn.mu.Lock()
	subs := append([]*session(nil), rn.subs...)
	rn.mu.Unlock()
	for _, s := range subs {
		s.sendBinary(b)
	}
}

func stopped(sr debug.StopReason) stoppedMsg {
	return stoppedMsg{
		Type: "stopped", Reason: sr.Kind, PC: hex64(sr.PC), Steps: sr.Steps, Note: sr.Note,
	}
}

// ---- the profile: a field is not a frame ----
//
// The machine profiles a FIELD, because a field — one VBlank — is the only boundary it
// has. A game does not have to draw in every one, and a 3DS game typically does not:
// Captain Toad renders at 30 Hz on a 60 Hz console, so it rasterises 383,000 fragments in
// one field and, in the next, draws nothing at all and only runs its game logic.
//
// Reported field by field, that is a profile panel flipping between 145 ms and 2 ms twice
// a frame — numbers that are individually true and together useless, because neither one
// is the cost of the picture on the screen.
//
// So a field the game drew nothing in is not published as a frame. Its time is carried and
// added to the next field that draws, and what the panel is shown is the cost of a picture:
// everything the machine did since the last one. Nothing is discarded — the logic-only
// field's ARM11 time is part of what that frame cost — the buckets still sum to the total,
// and between two pictures the panel simply holds the last frame's numbers rather than
// blinking through an empty field's.
type profAccum struct {
	carry *jsonProfile // fields drawn in nothing yet, waiting for the one that draws
	last  *jsonProfile // the last frame that drew, with those fields folded into it
}

// observe reads the profile of the field just advanced and folds it into the frame being
// built. It must be called once per field the machine advances, played or captured, or the
// time in the fields it misses is lost.
func (pa *profAccum) observe(t debug.Target) {
	p, ok := t.(debug.Profiler)
	if !ok {
		return
	}
	fp := p.FrameProfile()
	if fp.TotalMs == 0 && len(fp.Buckets) == 0 {
		return // profiling is off, or no field has completed yet
	}
	pa.carry = addProfile(pa.carry, fp)
	if fp.Drew {
		pa.last, pa.carry = pa.carry, nil
	}
}

// emit is the frame to report: the last one that put a picture on the screen. A target
// that cannot profile sends nothing, and the page shows no profile panel — rather than an
// empty one that reads as "the frame cost nothing".
func (pa *profAccum) emit() *jsonProfile { return pa.last }

// reset drops the accumulation. The machine is somewhere else now, and the fields we were
// folding together were on the way to a picture that will never be drawn.
func (pa *profAccum) reset() { pa.carry, pa.last = nil, nil }

// addProfile adds a field's profile into the frame being accumulated, matching buckets and
// counters by name so a field that ran no draws (and so reports no vertex bucket) still
// lands in the right place.
func addProfile(acc *jsonProfile, fp debug.FrameProfile) *jsonProfile {
	if acc == nil {
		acc = &jsonProfile{}
	}
	acc.TotalMs += fp.TotalMs
	for _, b := range fp.Buckets {
		if i := indexOf(len(acc.Buckets), func(i int) bool { return acc.Buckets[i].Name == b.Name }); i >= 0 {
			acc.Buckets[i].Millis += b.Millis
			acc.Buckets[i].Count += b.Count
			continue
		}
		acc.Buckets = append(acc.Buckets, jsonProfBucket{Name: b.Name, Millis: b.Millis, Count: b.Count})
	}
	for _, c := range fp.Counters {
		if i := indexOf(len(acc.Counters), func(i int) bool { return acc.Counters[i].Name == c.Name }); i >= 0 {
			acc.Counters[i].Value += c.Value
			continue
		}
		acc.Counters = append(acc.Counters, jsonProfCounter{Name: c.Name, Value: c.Value})
	}
	return acc
}

func indexOf(n int, match func(int) bool) int {
	for i := 0; i < n; i++ {
		if match(i) {
			return i
		}
	}
	return -1
}

func msSince(t time.Time) float64 { return float64(time.Since(t).Microseconds()) / 1000 }
func hex32(v uint32) string       { return fmt.Sprintf("%08x", v) }
func hex64(v uint64) string       { return fmt.Sprintf("%016x", v) }
