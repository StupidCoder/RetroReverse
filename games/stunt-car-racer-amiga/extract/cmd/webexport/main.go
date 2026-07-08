// webexport reconstructs Stunt Car Racer's web assets from the decoded game binary
// and writes the common format-2 asset tree (see site/FORMAT2.md / STANDARDS.md §4)
// under the output root:
//
//	manifest.json               game index: native res, tick rate, circuit view list
//	tracks/<circuit>.json       per circuit: the decoded plan nodes + absolute per-rung
//	                            rail geometry (the SAME object cmd/trackjson emits as an
//	                            array element, one circuit per file)
//
// The eight circuits are bespoke 3-D views (no GLB) driven by the site's "stunt-track"
// viewer, so they are indexed under manifest.views[]. The decode is the verified pure-Go
// spine + baked $65BEC model (package track); it is folded into this one command so
// cmd/trackjson stays standalone. -only gates which stages run.
//
// Usage: webexport [-in <game.dec.bin>] [-o <outdir>] [-only tracks,all]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manifest is the format-2 game index (STANDARDS.md §4.2). Sections gated out by
// -only stay empty and omitempty drops them, so a partial run is self-consistent.
type Manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Views    []ViewIndex    `json:"views,omitempty"`
	Models   []ModelIndex   `json:"models,omitempty"`
}

// ViewIndex is one manifest views[] entry: a bespoke per-game three.js view. Each
// circuit is one stunt-track view pointing at its per-circuit track JSON.
type ViewIndex struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"` // tracks/<slug>.json
	Kind    string `json:"kind"` // "stunt-track"
}

// slug turns a circuit name into a stable kebab-case file stem ("Roller Coaster" ->
// "roller-coaster"). Circuit names are unique, so the slug is a stable identity.
func slug(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// parseOnly turns the -only selection into a stage set. "all" (or empty) selects
// every stage; otherwise only the named stages (tracks) run.
func parseOnly(sel string) map[string]bool {
	out := map[string]bool{}
	for _, s := range strings.Split(sel, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if s == "all" {
			out["tracks"] = true
			out["models"] = true
			continue
		}
		if s != "tracks" && s != "models" {
			fmt.Fprintf(os.Stderr, "webexport: unknown -only stage %q (want tracks,models,all)\n", s)
			os.Exit(2)
		}
		out[s] = true
	}
	return out
}

func chk(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}

func writeJSON(path string, v any) {
	b, err := json.Marshal(v)
	chk(err)
	chk(os.WriteFile(path, b, 0o644))
}

func main() {
	in := flag.String("in", "../extracted/game.dec.bin", "input decoded game binary (game.dec.bin)")
	outdir := flag.String("o", "../../site/public/stunt-car-racer-amiga", "output asset root")
	only := flag.String("only", "all", "comma-separated subset of stages: tracks,all")
	flag.Parse()
	sel := parseOnly(*only)
	chk(os.MkdirAll(*outdir, 0o755))

	man := Manifest{
		Format: 2, Game: "stunt-car-racer-amiga", Platform: "Amiga",
		Native: map[string]int{"w": 320, "h": 200}, TickHz: 50,
	}
	if sel["tracks"] {
		man.Views = exportTracks(*in, *outdir)
	}
	if sel["models"] {
		man.Models = exportModels(*in, *outdir)
	}
	writeJSON(filepath.Join(*outdir, "manifest.json"), man)
	fmt.Fprintf(os.Stderr, "[manifest] %d circuits -> %s\n", len(man.Views), *outdir)
}
