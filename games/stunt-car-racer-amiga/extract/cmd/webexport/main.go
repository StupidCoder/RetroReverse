// webexport reconstructs Stunt Car Racer's web assets from the decoded game binary
// and writes the common format-2 asset tree (see FORMAT2.md / STANDARDS.md §4)
// under the output root:
//
//	manifest.json               game index: native res, tick rate, the models[] list
//	models/<slug>.glb           one baked GLB per asset: the eight circuits (road ribbon,
//	                            walls and stroked decal LINES, coloured as the pre-race
//	                            preview renders them), the opponent car and the horizon ring
//
// Every asset is a GLB driven by the site's "stunt-model" viewer, indexed under
// manifest.models[] (the circuits under section "Courses", the opponent car + horizon
// under "Models"). The geometry is the verified pure-Go spine + baked $65BEC model
// (package track) coloured byte-exact by cmd/coloracle; cmd/trackjson still emits the
// standalone per-circuit JSON if the decoded geometry is wanted outside the site.
//
// Usage: webexport [-in <game.dec.bin>] [-o <outdir>]
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
// Every asset is a baked GLB under models[] (the eight circuits + the opponent car +
// the horizon), driven by the site's stunt-model viewer.
type Manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Models   []ModelIndex   `json:"models,omitempty"`
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
	flag.Parse()
	chk(os.MkdirAll(*outdir, 0o755))

	man := Manifest{
		Format: 2, Game: "stunt-car-racer-amiga", Platform: "Amiga",
		Native: map[string]int{"w": 320, "h": 200}, TickHz: 50,
	}
	man.Models = exportModels(*in, *outdir)
	writeJSON(filepath.Join(*outdir, "manifest.json"), man)
	fmt.Fprintf(os.Stderr, "[manifest] %d models -> %s\n", len(man.Models), *outdir)
}
