// webexport is Turrican's single-call format-2 asset exporter (see STANDARDS.md §4 /
// FORMAT2.md). It reconstructs everything from the disk image and writes the common
// asset tree under the output root:
//
//	manifest.json                 game index: native res, tick rate, level/music list, sprites ref
//	levels/world<w>_scene<s>.json per scene (kind "tilemap2d"): the format-1 tilemap body
//	                              (grid + hflipMask + collision + objects) inside the format-2
//	                              envelope, with the object-layer sprite atlas refs inline
//	levels/world<w>_scene<s>.objects.json  machine-readable object DB for the scene
//	levels/atlas<w>.png           the world's 32x32 tiles in its palette
//	levels/objatlas<w>.png        first animation frame of every object sprite the scenes use
//	sprites/index.json + PNGs     per-frame-table BOB sprite sheets (folded ex-cmd/sprites)
//	music/*.mp3                    every distinct TFMX sub-song, rendered by the real driver
//	                              in the m68k interpreter (folded ex-cmd/music)
//
// All three producers (levels, sprites, music) are folded into this one command; -only gates
// which stages run. The music stage runs the real 68000 sound driver in an interpreter, so it
// is the slow stage; it is fully self-contained (needs no levels/sprites stage).
//
// Usage: webexport [-in Turrican.adf] [-o dir] [-only levels,sprites,music,all]
package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/turrican-amiga/extract/scene"
)

// manifest is the format-2 per-game index (FORMAT2.md).
type manifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Platform string         `json:"platform"`
	Native   map[string]int `json:"native"`
	TickHz   int            `json:"tickHz"`
	Levels   []levelIndex   `json:"levels,omitempty"`
	Music    []musicEntry   `json:"music,omitempty"`
	Sprites  string         `json:"sprites,omitempty"`
}

type levelIndex struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Atlas   string `json:"atlas,omitempty"`
	Objects string `json:"objects,omitempty"`
}

type musicEntry struct {
	Name string `json:"name"`
	File string `json:"file"`
}

// parseOnly turns the -only selection into a stage set. "all" (or empty) selects every stage;
// otherwise only the named stages (levels,sprites,music) run.
func parseOnly(sel string) map[string]bool {
	out := map[string]bool{}
	for _, s := range strings.Split(sel, ",") {
		s = strings.TrimSpace(s)
		switch s {
		case "":
		case "all":
			out["levels"], out["sprites"], out["music"] = true, true, true
		case "levels", "sprites", "music":
			out[s] = true
		default:
			fmt.Fprintf(os.Stderr, "webexport: unknown -only stage %q (want levels,sprites,music,all)\n", s)
			os.Exit(2)
		}
	}
	return out
}

func main() {
	in := flag.String("in", "../Turrican.adf", "input Amiga disk image (.adf)")
	outDir := flag.String("o", "../../site/public/turrican-amiga", "output asset root")
	only := flag.String("only", "all", "comma-separated subset of stages to run: levels,sprites,music,all")
	flag.Parse()
	if flag.NArg() > 0 {
		*in = flag.Arg(0) // tolerate a positional ADF path
	}
	sel := parseOnly(*only)

	adf, err := os.ReadFile(*in)
	if err != nil {
		fail(err)
	}
	game, err := scene.Load(adf)
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fail(err)
	}

	// The manifest covers whatever stages ran: a gated-out stage leaves its section empty and
	// omitempty drops it, so a partial run stays self-consistent.
	man := manifest{
		Format: 2, Game: "turrican-amiga", Platform: "Amiga",
		Native: map[string]int{"w": amigaViewW, "h": amigaViewH}, TickHz: 50,
	}
	if sel["levels"] {
		man.Levels, err = exportLevels(game, *outDir)
		if err != nil {
			fail(err)
		}
	}
	if sel["sprites"] {
		if err := exportSprites(game, *outDir); err != nil {
			fail(err)
		}
		man.Sprites = "sprites/index.json"
	}
	if sel["music"] {
		man.Music, err = exportMusic(adf, game, *outDir)
		if err != nil {
			fail(err)
		}
	}
	if err := writeJSON(filepath.Join(*outDir, "manifest.json"), man); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "[manifest] %d levels, %d music, sprites=%q -> %s\n",
		len(man.Levels), len(man.Music), man.Sprites, *outDir)
}

// writeJSON marshals v to path (compact, deterministic).
func writeJSON(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// writePNG encodes a paletted image, hashes the bytes and writes the file, returning a short
// content-hash cache-buster (`?v=<hash>`). The atlas layout changes whenever the tile/sprite
// set changes, so a stale browser-cached PNG would sample the new index rects at the wrong
// pixels; the versioned URL forces a fresh fetch. Empty img returns "".
func writePNG(path string, img *image.Paletted) (string, error) {
	if img.Bounds().Empty() {
		return "", nil
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(buf.Bytes()))[:8], nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "webexport:", err)
	os.Exit(1)
}
