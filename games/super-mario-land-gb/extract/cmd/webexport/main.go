// webexport serializes the decoded Super Mario Land assets for the static web viewer as
// the common format-2 asset tree (see STANDARDS.md / FORMAT2.md). Everything is
// reconstructed from the cartridge by the same decode path as cmd/levelmap, then written
// under the output root:
//
//	manifest.json                    game index: native res, tick rate, level/music list
//	levels/level-W-L.json            per level (kind "tilemap2d"): grid + objects + collision,
//	                                 wrapped in the format-2 envelope; atlas + objects ref
//	levels/level-W-L.objects.json    machine-readable object DB for the level
//	levels/worldN.png                the world's 256-tile background atlas (16 wide)
//	sprites/index.json               object/enemy metasprite icon metadata (see sprites.go)
//	sprites/worldN-obj.png           per-world composited object-icon strips
//	music/level-W-L.mp3, bonus.mp3   the level themes + bonus jingle, pure-ROM synth (music.go)
//
// All three producers (levels, sprites, music) are folded into this one command; -only gates
// which stages run. The levels and sprites stages use the machine oracle only to snapshot the
// per-world VRAM/palette pixels (never the maps); the music stage is oracle-free.
//
// Usage: webexport -in <rom.gb> [-o <outdir>] [-only levels,sprites,music,all]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	tile  = 8
	gut   = 1            // 1px extruded gutter so tiles don't bleed at fractional zoom
	cell  = tile + 2*gut // 10
	cols  = 16           // atlas is 16 tiles wide (256 tiles total)
	nTile = 256
)

// Manifest is the format-2 per-game index (FORMAT2.md).
type Manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Levels   []LevelIndex   `json:"levels,omitempty"`
	Music    []MusicEntry   `json:"music,omitempty"`
	Sprites  string         `json:"sprites,omitempty"`
}

type LevelIndex struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Atlas   string `json:"atlas,omitempty"`
	Objects string `json:"objects,omitempty"`
}

type MusicEntry struct {
	Name string `json:"name"`
	File string `json:"file"`
}

// parseOnly turns the -only selection into a stage set. "all" (or empty) selects every
// stage; otherwise only the named stages (levels,sprites,music) run.
func parseOnly(sel string) map[string]bool {
	out := map[string]bool{}
	for _, s := range strings.Split(sel, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if s == "all" {
			out["levels"], out["sprites"], out["music"] = true, true, true
			continue
		}
		if s != "levels" && s != "sprites" && s != "music" {
			fmt.Fprintf(os.Stderr, "webexport: unknown -only stage %q (want levels,sprites,music,all)\n", s)
			os.Exit(2)
		}
		out[s] = true
	}
	return out
}

func main() {
	in := flag.String("in", "../Super Mario Land (World).gb", "input Game Boy ROM (.gb)")
	outdir := flag.String("o", "../../site/public/super-mario-land-gb", "output asset root")
	only := flag.String("only", "all", "comma-separated subset of stages to run: levels,sprites,music,all")
	flag.Parse()
	if *in == "" && flag.NArg() > 0 {
		*in = flag.Arg(0) // tolerate a positional ROM path
	}
	sel := parseOnly(*only)

	data, err := os.ReadFile(*in)
	must(err)
	must(os.MkdirAll(*outdir, 0o755))

	// The manifest covers whatever stages ran: a stage gated out via -only leaves its
	// manifest section empty, and omitempty drops it, so a partial run stays self-consistent.
	man := Manifest{
		Format: 2, Game: "super-mario-land-gb", Platform: "Nintendo Game Boy",
		Native: map[string]int{"w": 160, "h": 144}, TickHz: 60,
	}
	if sel["levels"] {
		man.Levels = exportLevels(data, *outdir)
	}
	if sel["sprites"] {
		exportSprites(data, *outdir)
		man.Sprites = "sprites/index.json"
	}
	if sel["music"] {
		man.Music = exportMusic(data, *outdir)
	}
	writeJSON(filepath.Join(*outdir, "manifest.json"), man)
	fmt.Fprintf(os.Stderr, "[manifest] %d levels, %d music, sprites=%q -> %s\n",
		len(man.Levels), len(man.Music), man.Sprites, *outdir)
}

func writeJSON(path string, v any) {
	b, err := json.Marshal(v)
	must(err)
	must(os.WriteFile(path, b, 0o644))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}
