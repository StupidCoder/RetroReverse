package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"retroreverse.com/tools/debug"
)

// A registered platform whose target is the fake one, so the workspace can be exercised
// end to end — open a game, save a state, load it back — without a real machine.
func init() {
	debug.Register("wsfake", func(s debug.OpenSpec) (debug.Target, error) {
		return &fakeTarget{coreTarget: coreTarget{title: "Fake " + s.Get("tag")}}, nil
	})
}

// gamesTree builds a games/ directory with one openable game.
func gamesTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "test-game-wsfake")
	if err := os.MkdirAll(filepath.Join(dir, "image"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "image", "game.z64"), []byte("rom"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "debug.json"), []byte(
		`{"name":"Test Game","platform":"wsfake","image":"image/game.z64","profile":{"tag":"one"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func serveLibrary(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	root := gamesTree(t)
	srv := httptest.NewServer(NewLibrary(root).Handler())
	t.Cleanup(srv.Close)
	return srv, root
}

// TestLibraryAndOpen: the page is told what it can open, and opening one makes it the
// target — with a hello carrying the new game's capabilities.
func TestLibraryAndOpen(t *testing.T) {
	srv, _ := serveLibrary(t)
	cl := dial(t, srv.URL)

	m := cl.recvType(t, "library")
	games := m["games"].([]any)
	if len(games) != 1 {
		t.Fatalf("library = %v", games)
	}
	g := games[0].(map[string]any)
	if g["slug"] != "test-game-wsfake" || g["name"] != "Test Game" || g["missing"] != false {
		t.Errorf("game = %v", g)
	}
	// The platform's human name rides along, because the page's menu groups by platform and
	// has no Target to ask (it is choosing which one to open). An unnamed platform falls back
	// to its tag rather than to an empty menu entry.
	if g["platform"] != "wsfake" || g["platformName"] != "wsfake" {
		t.Errorf("platform = %v / %v, want the tag and its fallback name", g["platform"], g["platformName"])
	}
	if m["current"] != "" {
		t.Errorf("a game is open before one was asked for: %v", m["current"])
	}

	// Nothing is open yet, so a machine op is refused rather than crashing.
	cl.send(t, "frame.step", 1, nil)
	if e := cl.recvType(t, "error"); !strings.Contains(e["msg"].(string), "no game is open") {
		t.Errorf("stepping with no game open = %v", e)
	}

	cl.send(t, "target.open", 2, map[string]any{"slug": "test-game-wsfake"})
	hello := cl.recvType(t, "hello")
	if hello["platform"] != "fake" { // the fake target's own platform name
		t.Errorf("hello = %v", hello)
	}
	// The profile reached the adapter.
	if hello["title"] != "Fake one" {
		t.Errorf("title = %v, want the profile's tag to have reached the opener", hello["title"])
	}

	// And now the machine answers.
	cl.send(t, "frame.step", 3, stepArgs{})
	cl.recvType(t, "frame")
}

// Switching games hands the page to the new machine — and, just as importantly, takes it
// away from the old one. A runner that was free-running has a frame in flight on its own
// goroutine; if the page is still attached when that frame lands, the game we just left
// paints over the game we just opened, which reads as "switching targets is broken".
func TestSwitchingTargetsTakesThePageAwayFromTheOldMachine(t *testing.T) {
	root := gamesTree(t)
	s := NewLibrary(root)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	cl := dial(t, srv.URL)
	cl.recvType(t, "library")
	cl.send(t, "target.open", 1, map[string]any{"slug": "test-game-wsfake"})
	cl.recvType(t, "hello")

	old := s.ws.runner()
	if old == nil {
		t.Fatal("no runner after opening a game")
	}

	// Free-run it, so the old machine is mid-frame when we switch away.
	cl.send(t, "frame.play", 2, playArgs{On: true})
	if m := cl.recvType(t, "render"); m["play"] != true {
		t.Fatalf("not a played frame: %v", m)
	}
	cl.recvBinary(t)

	cl.send(t, "target.open", 3, map[string]any{"slug": "test-game-wsfake"})
	for i := 0; ; i++ {
		if i > 16 {
			t.Fatal("no hello after switching targets")
		}
		m := cl.recvJSON(t)
		if m["type"] == "hello" {
			break
		}
		if m["type"] == "render" {
			cl.recvBinary(t)
		}
	}

	old.mu.Lock()
	subs := len(old.subs)
	old.mu.Unlock()
	if subs != 0 {
		t.Errorf("the closed machine still has %d page(s) attached: its last frame will paint over the new game", subs)
	}

	rn := s.ws.runner()
	if rn == old {
		t.Fatal("the workspace still points at the old runner")
	}

	// And the new machine answers. It is a fresh boot, playing nothing.
	if rn.playing {
		t.Error("the new machine inherited the old one's play state")
	}
	cl.send(t, "frame.step", 4, stepArgs{})
	for i := 0; ; i++ {
		if i > 16 {
			t.Fatal("the new machine never answered a step")
		}
		m := cl.recvJSON(t)
		if m["type"] == "frame" {
			if len(m["commands"].([]any)) != fakeCmds {
				t.Errorf("the new machine captured %v commands", len(m["commands"].([]any)))
			}
			break
		}
		if m["type"] == "render" {
			cl.recvBinary(t)
		}
	}
}

func TestOpenUnknownGame(t *testing.T) {
	srv, _ := serveLibrary(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "library")

	cl.send(t, "target.open", 1, map[string]any{"slug": "nope"})
	if e := cl.recvType(t, "error"); !strings.Contains(e["msg"].(string), "no game") {
		t.Errorf("error = %v", e)
	}
}

// TestSavestateRoundTripAndResumeLine is the handoff, tested as a user would use it:
// save a state, see it listed, and get back a command line that actually names the
// state file and the game's own oracle flag.
func TestSavestateRoundTripAndResumeLine(t *testing.T) {
	srv, root := serveLibrary(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "library")
	cl.send(t, "target.open", 1, map[string]any{"slug": "test-game-wsfake"})
	cl.recvType(t, "hello")

	// Nothing saved yet.
	cl.send(t, "state.list", 2, nil)
	if m := cl.recvType(t, "states"); len(m["states"].([]any)) != 0 {
		t.Errorf("states = %v, want none", m["states"])
	}

	cl.send(t, "state.save", 3, map[string]any{"name": "bad-frame"})
	if m := cl.recvType(t, "ok"); m["op"] != "state.save" {
		t.Fatalf("save reply = %v", m)
	}

	// It is on disk, where the repository keeps regenerable scratch.
	want := filepath.Join(root, "test-game-wsfake", "work", "states", "bad-frame.state")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("the state was not written to %s: %v", want, err)
	}

	cl.send(t, "state.list", 4, nil)
	m := cl.recvType(t, "states")
	states := m["states"].([]any)
	if len(states) != 1 {
		t.Fatalf("states = %v", states)
	}
	st := states[0].(map[string]any)
	if st["name"] != "bad-frame" {
		t.Errorf("state = %v", st)
	}

	// The resume line is the point of the feature: it must name the state file and use
	// the flag the fake target's own oracle takes.
	resume := st["resume"].(string)
	for _, want := range []string{"bad-frame.state", "-loadstate", "cd "} {
		if !strings.Contains(resume, want) {
			t.Errorf("resume line %q does not contain %q", resume, want)
		}
	}

	cl.send(t, "state.load", 5, map[string]any{"name": "bad-frame"})
	if m := cl.recvType(t, "ok"); m["op"] != "state.load" {
		t.Errorf("load reply = %v", m)
	}

	cl.send(t, "state.delete", 6, map[string]any{"name": "bad-frame"})
	if m := cl.recvType(t, "states"); len(m["states"].([]any)) != 0 {
		t.Errorf("the state survived delete: %v", m["states"])
	}
}

// TestStateNameCannotEscape: the name comes from a browser, and it names a file.
func TestStateNameCannotEscape(t *testing.T) {
	srv, _ := serveLibrary(t)
	cl := dial(t, srv.URL)
	cl.recvType(t, "library")
	cl.send(t, "target.open", 1, map[string]any{"slug": "test-game-wsfake"})
	cl.recvType(t, "hello")

	for _, bad := range []string{"../../etc/passwd", "a/b", "", "..", `x\y`} {
		cl.send(t, "state.save", 2, map[string]any{"name": bad})
		m := cl.recvType(t, "error")
		if !strings.Contains(m["msg"].(string), "name") {
			t.Errorf("saving to name %q was refused unhelpfully: %v", bad, m["msg"])
		}
	}
}

// TestJSONShapes guards the messages the page parses: a field renamed here is a panel
// that silently shows nothing.
func TestJSONShapes(t *testing.T) {
	b, _ := json.Marshal(libraryMsg{Type: "library", Games: []jsonGame{{Slug: "s", Platform: "p", PlatformName: "P"}}})
	for _, want := range []string{
		`"type":"library"`, `"games"`, `"slug":"s"`, `"missing":false`,
		// The page groups the library by platform and labels each group with this; without
		// it the menu falls back to the raw tag, which is the thing it exists to avoid.
		`"platform":"p"`, `"platformName":"P"`,
	} {
		if !strings.Contains(string(b), want) {
			t.Errorf("library message %s lacks %s", b, want)
		}
	}
	b, _ = json.Marshal(statesMsg{Type: "states", States: []jsonState{{Name: "n", Resume: "r"}}})
	for _, want := range []string{`"type":"states"`, `"name":"n"`, `"resume":"r"`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("states message %s lacks %s", b, want)
		}
	}
}
