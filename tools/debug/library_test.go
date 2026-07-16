package debug

import (
	"os"
	"path/filepath"
	"testing"
)

// A throwaway platform, registered once, so the library tests do not depend on a real
// adapter (and so they cannot collide with one).
func init() {
	Register("fake64", func(s OpenSpec) (Target, error) { return nil, nil })
}

// buildTree lays out a games/ directory the way the repository does — including the
// parts that are not uniform, because that is what discovery has to cope with.
func buildTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// The STANDARDS layout: the image lives in image/.
	must(t, os.MkdirAll(filepath.Join(root, "pilotwings-64-fake64", "image"), 0o755))
	write(t, filepath.Join(root, "pilotwings-64-fake64", "image", "Pilotwings 64.z64"), "rom")

	// A game not migrated to that layout: the image sits in the game directory.
	must(t, os.MkdirAll(filepath.Join(root, "ridge-racer-fake64"), 0o755))
	write(t, filepath.Join(root, "ridge-racer-fake64", "Ridge Racer (Track 01).z64"), "rom")

	// A game whose image is not on disk — the normal state, since images are
	// gitignored.
	must(t, os.MkdirAll(filepath.Join(root, "missing-game-fake64", "image"), 0o755))

	// A platform with no adapter: not the debugger's business.
	must(t, os.MkdirAll(filepath.Join(root, "some-game-nosuchplatform", "image"), 0o755))
	write(t, filepath.Join(root, "some-game-nosuchplatform", "image", "x.z64"), "rom")

	return root
}

func TestDiscoverUsesTheSlugAndFindsImages(t *testing.T) {
	root := buildTree(t)
	games, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}

	byslug := map[string]Game{}
	for _, g := range games {
		byslug[g.Slug] = g
	}

	// The slug's suffix is the platform; nothing had to be configured.
	pw, ok := byslug["pilotwings-64-fake64"]
	if !ok {
		t.Fatal("pilotwings not discovered")
	}
	if pw.Platform != "fake64" {
		t.Errorf("platform = %q, want fake64", pw.Platform)
	}
	if pw.Name != "Pilotwings 64" {
		t.Errorf("name = %q, want %q", pw.Name, "Pilotwings 64")
	}
	if filepath.Base(pw.Image) != "Pilotwings 64.z64" {
		t.Errorf("image = %q", pw.Image)
	}
	if pw.Missing {
		t.Error("pilotwings' image is on disk but was marked missing")
	}

	// The un-migrated layout is found too — the images really are in both places.
	rr, ok := byslug["ridge-racer-fake64"]
	if !ok {
		t.Fatal("ridge racer not discovered (image outside image/)")
	}
	if filepath.Base(rr.Image) != "Ridge Racer (Track 01).z64" {
		t.Errorf("image = %q", rr.Image)
	}

	// A platform with no adapter is not offered at all.
	if _, ok := byslug["some-game-nosuchplatform"]; ok {
		t.Error("a game on an unregistered platform was offered")
	}

	// A game with no image is not offered either — there is nothing to open.
	if _, ok := byslug["missing-game-fake64"]; ok {
		t.Error("a game with no image file at all was offered")
	}
}

// TestProfileOverrides: a game that needs more than the convention says so in
// debug.json, and that is where game-specific knowledge belongs — a data file read at
// runtime, not an import in the tools module.
func TestProfileOverrides(t *testing.T) {
	root := buildTree(t)
	dir := filepath.Join(root, "ridge-racer-fake64")
	write(t, filepath.Join(dir, "debug.json"), `{
		"name": "Ridge Racer",
		"platform": "fake64",
		"image": "Ridge Racer (Track 01).z64",
		"profile": { "isr": "8004DF48" }
	}`)

	games, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	var g Game
	for _, x := range games {
		if x.Slug == "ridge-racer-fake64" {
			g = x
		}
	}
	if g.Name != "Ridge Racer" {
		t.Errorf("name = %q, want the one from debug.json", g.Name)
	}
	if got := g.Spec().Get("isr"); got != "8004DF48" {
		t.Errorf("profile isr = %q, want 8004DF48", got)
	}
	if want := filepath.Join(dir, "work", "states"); g.StatesDir() != want {
		t.Errorf("states dir = %q, want %q", g.StatesDir(), want)
	}
}

func TestOpenUnknownPlatform(t *testing.T) {
	if _, err := Open("nosuch", OpenSpec{}); err == nil {
		t.Error("opening an unregistered platform succeeded")
	}
	if !Registered("fake64") {
		t.Error("fake64 was registered but Registered says otherwise")
	}
}

func TestFindGamesRoot(t *testing.T) {
	root := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(root, "games", "a-fake64"), 0o755))
	deep := filepath.Join(root, "tools", "debug", "server")
	must(t, os.MkdirAll(deep, 0o755))

	got := FindGamesRoot(deep)
	want := filepath.Join(root, "games")
	// t.TempDir can hand back a symlinked path (/var → /private/var on macOS), so
	// compare what the filesystem resolves to rather than the strings.
	if resolve(got) != resolve(want) {
		t.Errorf("FindGamesRoot(%q) = %q, want %q", deep, got, want)
	}
	if FindGamesRoot(t.TempDir()) != "" {
		t.Error("FindGamesRoot found a games/ tree where there is none")
	}
}

func resolve(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, s string) {
	t.Helper()
	must(t, os.WriteFile(path, []byte(s), 0o644))
}

// An unknown platform falls back to its tag rather than to "" — an empty menu entry is worse
// than an ugly one.
func TestPlatformNameFallsBackToTheTag(t *testing.T) {
	if got := PlatformName("nothing-registers-this"); got != "nothing-registers-this" {
		t.Errorf("PlatformName of an unknown platform = %q", got)
	}
}
