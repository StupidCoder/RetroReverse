package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"retroreverse.com/tools/debug"
)

// A Workspace owns the library of openable games and the one target that is open.
//
// One at a time, deliberately: a target is a whole machine (a 3DS is hundreds of
// megabytes), and the point of savestates is that coming back to where you were is
// cheap. Switching closes the old machine and boots or restores the new one.
type Workspace struct {
	root string // the games/ tree, "" if we could not find one

	mu    sync.Mutex
	games []debug.Game
	cur   *Runner
	game  *debug.Game // the open game, nil if the target was passed in directly
	subs  []*session
}

func newWorkspace(root string) *Workspace {
	w := &Workspace{root: root}
	w.rescan()
	return w
}

func (w *Workspace) rescan() {
	if w.root == "" {
		return
	}
	games, err := debug.Discover(w.root)
	if err != nil {
		return
	}
	w.mu.Lock()
	w.games = games
	w.mu.Unlock()
}

// runner is the target a session's requests go to.
func (w *Workspace) runner() *Runner {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.cur
}

// setTarget installs a runner, closing whatever was open, and tells every attached
// page what it is now looking at.
func (w *Workspace) setTarget(rn *Runner, g *debug.Game) {
	w.mu.Lock()
	old := w.cur
	w.cur, w.game = rn, g
	subs := append([]*session(nil), w.subs...)
	w.mu.Unlock()

	if old != nil {
		// Detach before shutting down. A runner that was free-running has a frame in
		// flight on its own goroutine, and it will finish it and broadcast it — a picture
		// from the machine we just closed, arriving after the new machine's hello and
		// painting over it. Detaching first means it broadcasts to nobody.
		for _, s := range subs {
			old.detach(s)
		}
		old.shutdown() // closes the machine on its own goroutine
	}
	for _, s := range subs {
		rn.attach(s)
	}
}

func (w *Workspace) attach(s *session) {
	w.mu.Lock()
	w.subs = append(w.subs, s)
	rn := w.cur
	w.mu.Unlock()

	if rn != nil {
		rn.attach(s)
	}
	w.sendLibrary(s, 0)
}

func (w *Workspace) detach(s *session) {
	w.mu.Lock()
	for i, x := range w.subs {
		if x == s {
			w.subs = append(w.subs[:i], w.subs[i+1:]...)
			break
		}
	}
	rn := w.cur
	w.mu.Unlock()
	if rn != nil {
		rn.detach(s)
	}
}

// handle takes the ops that are about the workspace rather than the machine. Anything
// else goes to the runner. It runs on the session's goroutine, and it must not touch a
// live target — opening one builds a new machine, which nobody else holds yet.
func (w *Workspace) handle(s *session, r req) bool {
	switch r.Op {
	case "target.list":
		w.sendLibrary(s, r.Seq)
	case "target.open":
		if err := w.open(s, r); err != nil {
			s.send(errMsg{Type: "error", Seq: r.Seq, Msg: err.Error()})
		}
	case "state.list":
		w.sendStates(s, r.Seq)
	case "state.delete":
		if err := w.deleteState(r); err != nil {
			s.send(errMsg{Type: "error", Seq: r.Seq, Msg: err.Error()})
			return true
		}
		w.sendStates(s, r.Seq)
	case "state.save", "state.load":
		// The page names a state; the file it lives in is the workspace's business,
		// and the machine that reads or writes it is the runner's. Resolve the one and
		// hand the other on.
		if err := w.forwardState(s, r); err != nil {
			s.send(errMsg{Type: "error", Seq: r.Seq, Msg: err.Error()})
		}
	default:
		return false
	}
	return true
}

func (w *Workspace) forwardState(s *session, r req) error {
	rn := w.runner()
	if rn == nil {
		return fmt.Errorf("no game is open")
	}
	// The capability gate belongs here too: these ops never reach the runner's own
	// check, and a target that cannot do savestates must be told so in the same words.
	if !rn.caps[debug.CapStates] {
		return fmt.Errorf("%s: this target cannot %s", r.Op, debug.CapStates)
	}
	var a struct {
		Name string `json:"name"`
	}
	if err := decodeArgs(r, &a); err != nil {
		return err
	}
	path, err := w.statePath(a.Name)
	if err != nil {
		return err
	}
	if r.Op == "state.save" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}
	rn.submit(s, req{Op: r.Op, Seq: r.Seq, Args: mustJSON(stateArgs{Path: path})})
	return nil
}

func (w *Workspace) sendLibrary(s *session, seq int) {
	w.mu.Lock()
	games := w.games
	cur := ""
	if w.game != nil {
		cur = w.game.Slug
	}
	w.mu.Unlock()

	m := libraryMsg{Type: "library", Seq: seq, Current: cur, Games: []jsonGame{}}
	for _, g := range games {
		m.Games = append(m.Games, jsonGame{
			Slug: g.Slug, Platform: g.Platform, PlatformName: debug.PlatformName(g.Platform),
			Name: g.Name, Image: filepath.Base(g.Image), Missing: g.Missing,
		})
	}
	s.send(m)
}

// open boots a game from the library and makes it the current target.
func (w *Workspace) open(s *session, r req) error {
	var a struct {
		Slug string `json:"slug"`
	}
	if err := decodeArgs(r, &a); err != nil {
		return err
	}
	w.mu.Lock()
	var g *debug.Game
	for i := range w.games {
		if w.games[i].Slug == a.Slug {
			g = &w.games[i]
			break
		}
	}
	w.mu.Unlock()
	if g == nil {
		return fmt.Errorf("no game %q in the library", a.Slug)
	}
	if g.Missing {
		return fmt.Errorf("%s: the image is not on disk (%s) — game images are gitignored",
			g.Slug, filepath.Base(g.Image))
	}

	s.send(okMsg{Type: "ok", Seq: r.Seq, Op: r.Op, Note: "opening " + g.Name + "…"})
	start := time.Now()
	tgt, err := debug.Open(g.Platform, g.Spec())
	if err != nil {
		return fmt.Errorf("opening %s: %w", g.Slug, err)
	}
	w.setTarget(newRunner(tgt), g)
	s.send(okMsg{Type: "ok", Seq: r.Seq, Op: r.Op,
		Note: fmt.Sprintf("%s booted in %.0f ms", g.Name, msSince(start))})
	w.sendLibrary(s, 0)
	return nil
}

// ---- savestates ----
//
// The point of these is not just to skip a boot. It is to be able to save the frame
// where something is wrong and hand it to somebody else — so the panel also shows the
// exact command line that resumes it, in the flag vocabulary that game's own oracle
// actually accepts.

func (w *Workspace) statesDir() (string, error) {
	w.mu.Lock()
	g := w.game
	w.mu.Unlock()
	if g == nil {
		return "", fmt.Errorf("savestates need a game from the library (this target was opened directly)")
	}
	return g.StatesDir(), nil
}

func (w *Workspace) sendStates(s *session, seq int) {
	m := statesMsg{Type: "states", Seq: seq, States: []jsonState{}}
	dir, err := w.statesDir()
	if err != nil {
		m.Note = err.Error()
		s.send(m)
		return
	}
	m.Dir = dir

	entries, err := os.ReadDir(dir)
	if err != nil {
		// No directory yet just means nothing has been saved. Not an error.
		s.send(m)
		return
	}
	names := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".state") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	w.mu.Lock()
	rn, game := w.cur, w.game
	w.mu.Unlock()

	for _, n := range names {
		st := jsonState{Name: strings.TrimSuffix(n, ".state"), File: filepath.Join(dir, n)}
		if fi, err := os.Stat(st.File); err == nil {
			st.Size = fi.Size()
			st.When = fi.ModTime().Format("2006-01-02 15:04")
		}
		st.Resume = resumeLine(rn, game, st.File)
		m.States = append(m.States, st)
	}
	s.send(m)
}

// resumeLine renders the shell command that reopens this game at a state. The oracles
// never agreed on their flags (-loadstate, -load, -state), so the adapter says what its
// own game's oracle takes; the workspace only adds the directory to run it from.
func resumeLine(rn *Runner, g *debug.Game, statePath string) string {
	if rn == nil || g == nil {
		return ""
	}
	r, ok := rn.tgt.(debug.Resumer)
	if !ok {
		return ""
	}
	args := r.ResumeArgs(statePath)
	quoted := make([]string, len(args))
	for i, a := range args {
		if strings.ContainsAny(a, " '\"") {
			quoted[i] = `"` + a + `"`
		} else {
			quoted[i] = a
		}
	}
	dir := filepath.Join(g.Dir, "extract")
	if rd, ok := rn.tgt.(debug.ResumeDirer); ok {
		if d := rd.ResumeDir(); d != "" {
			dir = d
		}
	}
	return fmt.Sprintf("cd %s && %s", dir, strings.Join(quoted, " "))
}

// statePath turns a name from the page into a file, refusing anything that would climb
// out of the states directory.
func (w *Workspace) statePath(name string) (string, error) {
	dir, err := w.statesDir()
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("a savestate needs a name")
	}
	if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return "", fmt.Errorf("bad savestate name %q", name)
	}
	return filepath.Join(dir, name+".state"), nil
}

func (w *Workspace) deleteState(r req) error {
	var a struct {
		Name string `json:"name"`
	}
	if err := decodeArgs(r, &a); err != nil {
		return err
	}
	path, err := w.statePath(a.Name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
