// exportglb converts NSBMD models to standard binary glTF (.glb) via
// nitro.ExportGLB, pairing each model with the textures of its companion NSBTX
// (same stem, then siblings). It accepts single files or, with -all, sweeps the
// menu kart and character model sets into an output directory — the models the
// viewer site serves.
//
//	exportglb [-o DIR] model.nsbmd...
//	exportglb -all [-root DIR] [-o DIR]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/nds"
	"retroreverse.com/tools/nds/nitro"
)

func main() {
	all := flag.Bool("all", false, "export the KartModelMenu kart + character sets")
	root := flag.String("root", "../extracted/files", "extracted filesystem root (-all)")
	outDir := flag.String("o", "../extracted/glb", "output directory")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	var paths []string
	if *all {
		filepath.Walk(filepath.Join(*root, "data", "KartModelMenu"), func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(p), ".nsbmd") &&
				!strings.Contains(p, "shadow") {
				paths = append(paths, p)
			}
			return nil
		})
		sort.Strings(paths)
	} else {
		paths = flag.Args()
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: exportglb [-o DIR] model.nsbmd...  |  exportglb -all")
		os.Exit(2)
	}

	var written []string
	for _, p := range paths {
		names, err := export(p, *outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", filepath.Base(p), err)
			continue
		}
		written = append(written, names...)
	}
	if *all {
		if err := writeManifest(*outDir, written); err != nil {
			die(err)
		}
	}
}

// charNames maps the two-letter character codes in model file names.
var charNames = map[string]string{
	"MR": "Mario", "LG": "Luigi", "PC": "Peach", "DS": "Daisy",
	"YS": "Yoshi", "KO": "Toad", "KP": "Bowser", "DK": "Donkey Kong",
	"WR": "Wario", "WL": "Waluigi", "KA": "Dry Bones", "RB": "R.O.B.",
}

// writeManifest emits models.json for the viewer site: [{name, file, section}].
func writeManifest(outDir string, files []string) error {
	type entry struct {
		Name    string `json:"name"`
		File    string `json:"file"`
		Section string `json:"section"`
	}
	var out []entry
	sort.Strings(files)
	add := func(e entry) { out = append(out, e) }
	// Characters first, then karts, in character order.
	var charOrder = []string{"MR", "LG", "PC", "DS", "YS", "KO", "KP", "DK", "WR", "WL", "KA", "RB"}
	for _, cc := range charOrder {
		for _, f := range files {
			if f == "P_"+cc+".glb" {
				add(entry{Name: charNames[cc], File: f, Section: "Characters"})
			}
		}
	}
	for _, cc := range charOrder {
		for _, v := range []string{"a", "b", "c"} {
			f := "kart_" + cc + "_" + v + ".glb"
			for _, w := range files {
				if w == f {
					add(entry{Name: charNames[cc] + " — kart " + strings.ToUpper(v), File: f, Section: "Karts"})
				}
			}
		}
	}
	// Anything else (e.g. the select scene) at the end.
	known := map[string]bool{}
	for _, e := range out {
		known[e.File] = true
	}
	for _, f := range files {
		if !known[f] {
			add(entry{Name: strings.TrimSuffix(f, ".glb"), File: f, Section: "Other"})
		}
	}
	buf, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return err
	}
	p := filepath.Join(outDir, "models.json")
	fmt.Printf("  models.json: %d entries\n", len(out))
	return os.WriteFile(p, buf, 0o644)
}

func export(path, outDir string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	models, err := nitro.ParseNSBMD(nds.Decompress(data))
	if err != nil {
		return nil, err
	}
	texs := loadTextures(path)
	var written []string
	for _, m := range models {
		glb, err := nitro.ExportGLB(m, texs)
		if err != nil {
			return written, err
		}
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		name := stem + ".glb"
		if len(models) > 1 {
			name = stem + "-" + m.Name + ".glb"
		}
		if err := os.WriteFile(filepath.Join(outDir, name), glb, 0o644); err != nil {
			return written, err
		}
		written = append(written, name)
		fmt.Printf("  %-28s %7d bytes\n", name, len(glb))
	}
	return written, nil
}

func loadTextures(modelPath string) map[string]nitro.Texture {
	texs := map[string]nitro.Texture{}
	cands := []string{strings.TrimSuffix(modelPath, filepath.Ext(modelPath)) + ".nsbtx"}
	if ents, err := os.ReadDir(filepath.Dir(modelPath)); err == nil {
		for _, e := range ents {
			if strings.HasSuffix(strings.ToLower(e.Name()), ".nsbtx") {
				cands = append(cands, filepath.Join(filepath.Dir(modelPath), e.Name()))
			}
		}
	}
	for _, c := range cands {
		data, err := os.ReadFile(c)
		if err != nil {
			continue
		}
		ts, err := nitro.DecodeNSBTX(nds.Decompress(data))
		if err != nil {
			continue
		}
		for _, t := range ts {
			if _, dup := texs[t.Name]; !dup {
				texs[t.Name] = t
			}
		}
	}
	return texs
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "exportglb:", err)
	os.Exit(1)
}
