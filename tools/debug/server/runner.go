package server

import (
	"encoding/hex"
	"fmt"
	"image"
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
	running bool

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
	case "cpu.regs":
		rn.sendCPU(r)
		return nil
	case "cpu.step":
		return rn.cpuStep(r)
	case "cpu.continue":
		rn.running = true
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
}

// ---- frames ----

func (rn *Runner) step(r request) error {
	var a stepArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	fs := rn.tgt.(debug.FrameStepper)
	start := time.Now()

	var fc *debug.FrameCapture
	var err error
	if a.N > 0 {
		for i := 0; i < a.N; i++ {
			fc, err = fs.StepFrame(a.Overdraw)
			if err != nil {
				return err
			}
			rn.frameNo++
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
		StepMs:   msSince(start),
		Profile:  frameProfile(rn.tgt),
	})
	// The provenance buffer goes over once per capture. From then on the page answers
	// "which command drew this pixel?" locally, with no round trip — which is what
	// makes hover-to-identify and the command-coverage overlay instant.
	if fc.Prov != nil {
		r.from.sendBinary(encodeProv(r.req.Seq, fc))
	}
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
		if len(fc.Commands) > 100 && fc.Prov != nil {
			return fc, nil
		}
	}
	return nil, fmt.Errorf("no drawn frame within the field budget")
}

// play starts or stops free-running the machine. While it runs nothing is captured:
// each field is stepped with no snapshot and no census, and the page is sent the
// scanout. Stopping does a full capture step, so you land on a real frame — the next
// field, not the last one played, which left no record behind to go back to.
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
		Op: "frame.step", Seq: r.req.Seq, Args: mustJSON(stepArgs{N: 1, Overdraw: overdrawOf(r.req)}),
	}})
}

func (rn *Runner) playFrame() error {
	start := time.Now()
	rn.frameNo++
	if f, ok := rn.tgt.(debug.FastStepper); ok {
		if err := f.StepFast(); err != nil {
			return err
		}
	} else if _, err := rn.tgt.(debug.FrameStepper).StepFrame(false); err != nil {
		return err
	}
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
		Profile: frameProfile(rn.tgt), // free-running frames keep the profile panel live
	})
	rn.inFlight = true
	rn.broadcastBinary(payload)
	return nil
}

func (rn *Runner) scrub(r request) error {
	var a scrubArgs
	if err := decodeArgs(r.req, &a); err != nil {
		return err
	}
	if rn.fc == nil {
		return fmt.Errorf("no captured frame yet")
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

	// k first, then a spread of nearby positions that are not already cached. The
	// spread is proportional to the list, so it is a fraction of the drag rather than a
	// fixed number of register writes — which on a 100,000-write frame would be nothing.
	n := len(rn.fc.Commands)
	ks := []int{k}
	step := n / 64
	if step < 1 {
		step = 1
	}
	for _, d := range []int{1, -1, 2, -2, 3, -3, 4} {
		p := k + d*step
		if p < 0 || p >= n {
			continue
		}
		if _, hit := rn.cache[p]; hit {
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
	// state we left.
	rn.fc = nil
	rn.clearCache()
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

// frameProfile asks the target where the frame's time went, if it can say. A
// target that cannot implement debug.Profiler sends nothing, and the page shows no
// profile panel — rather than an empty one that reads as "the frame cost nothing".
func frameProfile(t debug.Target) *jsonProfile {
	p, ok := t.(debug.Profiler)
	if !ok {
		return nil
	}
	fp := p.FrameProfile()
	if fp.TotalMs == 0 && len(fp.Buckets) == 0 {
		return nil
	}
	out := &jsonProfile{TotalMs: fp.TotalMs}
	for _, b := range fp.Buckets {
		out.Buckets = append(out.Buckets, jsonProfBucket{Name: b.Name, Millis: b.Millis, Count: b.Count})
	}
	for _, c := range fp.Counters {
		out.Counters = append(out.Counters, jsonProfCounter{Name: c.Name, Value: c.Value})
	}
	return out
}

func msSince(t time.Time) float64 { return float64(time.Since(t).Microseconds()) / 1000 }
func hex32(v uint32) string       { return fmt.Sprintf("%08x", v) }
func hex64(v uint64) string       { return fmt.Sprintf("%016x", v) }
