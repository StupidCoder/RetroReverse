// webexport runs the whole Mario Kart DS extraction from the raw cartridge and
// writes the Studio's common format-2 asset tree (site/FORMAT2.md) under one
// output root. It consolidates the separate tools (ndsextract, exportglb,
// musicrender): it stages the filesystem in a temp dir, renders the SDAT music,
// exports every NSBMD model to GLB (courses, skyboxes, karts, characters, map
// objects), and writes each track as a format-2 mesh3d level with its OBJI
// object database, CPU drive-line and texture-animation side-cars — reusing the
// same mkds / nitro / nds / sdat package APIs the individual commands call.
//
//	webexport [-in rom.nds] [-o DIR] [-only music,models,levels,all]
//
// -only gates which stages run: `-only music` renders MP3s only and never touches
// the model pipeline; models/levels stage the filesystem and export GLBs.
// Progress is reported on stderr, one line per stage plus a running count within
// long stages.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/platform/nds"
)

func main() {
	in := flag.String("in", "../Mario Kart DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image (.nds)")
	out := flag.String("o", "../../site/public/mario-kart-ds", "output root")
	only := flag.String("only", "all", "comma-separated subset of music,models,levels,all")
	flag.Parse()

	sel := parseOnly(*only)
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}

	img, err := os.ReadFile(*in)
	if err != nil {
		die(err)
	}
	rom, err := nds.Open(img)
	if err != nil {
		die(err)
	}

	// manifest accumulators
	var music []manifestMusic
	var models []manifestModel
	var levels []manifestLevel

	if sel["music"] {
		music = runMusic(rom, *out)
	}

	if sel["models"] || sel["levels"] {
		// Stage the filesystem the NSBMD/NKM decoders read from disk.
		tmp, err := os.MkdirTemp("", "mkds-webexport-")
		if err != nil {
			die(err)
		}
		defer os.RemoveAll(tmp)
		fmt.Fprintf(os.Stderr, "[extract] staging filesystem → %s\n", tmp)
		if err := extractFS(rom, tmp); err != nil {
			die(err)
		}

		// The object-ID → model-name bindings from the ARM9's map-object descriptor
		// table (Part V §2) drive the OBJI placements.
		arm9 := rom.ARM9()
		if nds.IsBLZ(arm9) {
			arm9 = nds.DecompressBLZ(arm9)
		}
		bindings = objectBindings(arm9)
		fmt.Fprintf(os.Stderr, "[extract] %d object-model bindings from ARM9\n", len(bindings))

		// GLB export path (needed by both models and levels): every menu/course/map
		// model into <o>/models. Records course archives for the levels stage.
		items := exportAllGLBs(filepath.Join(tmp, "files"), filepath.Join(*out, "models"))

		if sel["models"] {
			models = buildModels(items)
		}
		if sel["levels"] {
			levels = buildLevels(*out, items)
		}
	}

	writeManifest(*out, music, models, levels)
}

func parseOnly(s string) map[string]bool {
	sel := map[string]bool{}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if p == "all" {
			sel["music"], sel["models"], sel["levels"] = true, true, true
			continue
		}
		sel[p] = true
	}
	return sel
}

// ---------------------------------------------------------------------------
// manifest
// ---------------------------------------------------------------------------

type manifestMusic struct {
	Name string `json:"name"`
	File string `json:"file"`
}
type manifestModel struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"`
}
type manifestLevel struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Objects string `json:"objects,omitempty"`
}

func writeManifest(out string, music []manifestMusic, models []manifestModel, levels []manifestLevel) {
	m := map[string]any{
		"format":   2,
		"game":     "mario-kart-ds",
		"platform": "Nintendo DS",
		"native":   map[string]int{"w": 256, "h": 192},
		"tickHz":   60,
	}
	if len(levels) > 0 {
		m["levels"] = levels
	}
	if len(models) > 0 {
		m["models"] = models
	}
	if len(music) > 0 {
		m["music"] = music
	}
	buf, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(out, "manifest.json"), buf, 0o644); err != nil {
		die(err)
	}
	fmt.Fprintf(os.Stderr, "[manifest] %d levels, %d models, %d music → manifest.json\n",
		len(levels), len(models), len(music))
}

// ---------------------------------------------------------------------------
// filesystem staging (as cmd/ndsextract -fs)
// ---------------------------------------------------------------------------

// extractFS writes the full filesystem under dir/files/, which the NSBMD/NKM
// decoders (mkds.LoadModels/LoadTextures/LoadNKM) read from disk.
func extractFS(rom *nds.ROM, dir string) error {
	for _, f := range rom.Files {
		p := filepath.Join(dir, "files", filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, rom.File(f.ID), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "webexport:", err)
	os.Exit(1)
}
