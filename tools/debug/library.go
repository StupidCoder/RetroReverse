package debug

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The library: the games the debugger can open.
//
// Discovery leans on a convention the repository already keeps, rather than a registry
// anyone has to maintain: a game lives in games/<slug>/, and the slug ends in the
// platform tag — pilotwings-64-n64, ridge-racer-psx, loco-roco-psp. So the directory
// name says which adapter to use, and the image is whatever image-shaped file is
// inside. Nothing here knows anything about a particular game.
//
// A game that needs more than that — a boot executable inside an ISO, an interrupt
// handler a PSX title expects the BIOS to have installed — puts it in an optional
// games/<slug>/debug.json. That file is read at runtime, so game-specific knowledge
// never becomes a Go dependency of the tools module.

// Game is one openable game.
type Game struct {
	Slug     string            `json:"slug"`     // "ridge-racer-psx"
	Platform string            `json:"platform"` // "psx"
	Name     string            `json:"name"`     // "Ridge Racer"
	Image    string            `json:"image"`    // absolute path
	Dir      string            `json:"-"`        // games/<slug>
	Profile  map[string]string `json:"-"`        // from debug.json
	Missing  bool              `json:"missing"`  // the image is not on disk (they are gitignored)
}

// Spec is the OpenSpec for this game.
func (g Game) Spec() OpenSpec { return OpenSpec{Image: g.Image, Profile: g.Profile} }

// StatesDir is where this game's savestates live. work/ is the repository's
// convention for regenerable scratch, and it is gitignored.
func (g Game) StatesDir() string { return filepath.Join(g.Dir, "work", "states") }

// imageExt maps an image extension to the platform that reads it. Used only when a
// directory holds several files and we must pick the image out.
var imageExt = map[string]string{
	".z64": "n64", ".n64": "n64", ".v64": "n64",
	".bin": "psx", ".iso": "psx", ".img": "psx",
	".cso": "psp",
	".nds": "ds",
	".3ds": "3ds", ".cci": "3ds",
	".gb": "gb", ".gg": "gg",
	".adf": "amiga", ".tap": "c64",
}

// debugProfile is the optional per-game file.
type debugProfile struct {
	Platform string            `json:"platform"`
	Image    string            `json:"image"`
	Profile  map[string]string `json:"profile"`
	Name     string            `json:"name"`
}

// Discover finds the games under a games/ root. A game whose image is absent is still
// listed, marked Missing — the images are gitignored, so "you do not have this one" is
// a normal state and worth saying out loud rather than hiding the game.
func Discover(root string) ([]Game, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var games []Game
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		g, ok := discoverOne(filepath.Join(root, e.Name()), e.Name())
		if ok {
			games = append(games, g)
		}
	}
	sort.Slice(games, func(i, j int) bool { return games[i].Slug < games[j].Slug })
	return games, nil
}

func discoverOne(dir, slug string) (Game, bool) {
	g := Game{Slug: slug, Dir: dir, Name: prettyName(slug), Profile: map[string]string{}}

	// The slug's suffix names the platform: ...-n64, ...-psx.
	if i := strings.LastIndex(slug, "-"); i >= 0 {
		g.Platform = slug[i+1:]
	}

	// An optional debug.json overrides any of it.
	if b, err := os.ReadFile(filepath.Join(dir, "debug.json")); err == nil {
		var p debugProfile
		if json.Unmarshal(b, &p) == nil {
			if p.Platform != "" {
				g.Platform = p.Platform
			}
			if p.Name != "" {
				g.Name = p.Name
			}
			if p.Image != "" {
				g.Image = abs(filepath.Join(dir, p.Image))
			}
			for k, v := range p.Profile {
				g.Profile[k] = v
			}
		}
	}

	if !Registered(g.Platform) {
		return Game{}, false // no adapter: not something the debugger can open
	}
	if g.Image == "" {
		g.Image = findImage(dir, g.Platform)
	}
	if g.Image == "" {
		return Game{}, false // nothing image-shaped here at all
	}
	if _, err := os.Stat(g.Image); err != nil {
		g.Missing = true
	}
	return g, true
}

// findImage looks where the images actually are. STANDARDS puts them in
// games/<slug>/image/, but not every game has been migrated, so the game directory
// itself is searched too.
//
// An extension whose platform matches the slug's wins; failing that, any image-shaped
// file will do. The extension is a hint, not the truth — the slug already said which
// platform this is, and a game that names its disc something unexpected should still be
// openable rather than silently absent from the library.
func findImage(dir, platform string) string {
	var fallback string
	for _, sub := range []string{filepath.Join(dir, "image"), dir} {
		entries, err := os.ReadDir(sub)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			p, known := imageExt[strings.ToLower(filepath.Ext(e.Name()))]
			if !known {
				continue
			}
			path := abs(filepath.Join(sub, e.Name()))
			if p == platform {
				return path
			}
			if fallback == "" {
				fallback = path
			}
		}
	}
	return fallback
}

// prettyName turns a slug into a title: "ridge-racer-psx" → "Ridge Racer".
func prettyName(slug string) string {
	parts := strings.Split(slug, "-")
	if len(parts) > 1 {
		parts = parts[:len(parts)-1] // drop the platform tag
	}
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// FindGamesRoot walks up from a starting directory looking for the games/ tree, so the
// debugger works from anywhere in the repository.
func FindGamesRoot(start string) string {
	dir := abs(start)
	for i := 0; i < 8; i++ {
		cand := filepath.Join(dir, "games")
		if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func abs(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}
