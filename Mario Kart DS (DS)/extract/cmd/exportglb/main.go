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

	"mariokartds/extract/mkds"

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
		// course archives: those with a "<x>_course"/"_stage" model inside
		filepath.Walk(filepath.Join(*root, "data", "Course"), func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(p, ".carc") ||
				strings.Contains(p, "Tex") {
				return nil
			}
			paths = append(paths, p)
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

	var written, pathFiles []string
	for _, p := range paths {
		names, pf, err := export(p, *outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", filepath.Base(p), err)
			continue
		}
		written = append(written, names...)
		if pf != "" {
			pathFiles = append(pathFiles, pf)
		}
	}
	if *all {
		if err := writeManifest(*outDir, written, pathFiles); err != nil {
			die(err)
		}
	}
}

// courseNames maps course-archive stems to the tracks' English names (confident
// ones only; unknown stems keep their internal name).
var courseNames = map[string]string{
	"cross_course":   "Figure-8 Circuit",
	"bank_course":    "Yoshi Falls",
	"beach_course":   "Cheep Cheep Beach",
	"mansion_course": "Luigi's Mansion",
	"desert_course":  "Desert Hills",
	"town_course":    "Delfino Square",
	"pinball_course": "Waluigi Pinball",
	"ridge_course":   "Shroom Ridge",
	"snow_course":    "DK Pass",
	// development leftovers still on the retail cartridge:
	"donkey_course":   "donkey_course (unused)",
	"luigi_course":    "luigi_course (unused)",
	"nokonoko_course": "nokonoko_course (unused)",
	"dokan_course":    "dokan_course (unused)",
	"test1_course":    "test1_course (unused)",
	"Award":           "Award ceremony scene",
	"beach_courseD":   "Cheep Cheep Beach (multiplayer)",
	"cross_courseD":   "Figure-8 Circuit (multiplayer)",
	"mansion_courseD": "Luigi's Mansion (multiplayer)",
	"clock_course":    "Tick-Tock Clock",
	"mario_course":    "Mario Circuit",
	"airship_course":  "Airship Fortress",
	"stadium_course":  "Wario Stadium",
	"garden_course":   "Peach Gardens",
	"koopa_course":    "Bowser Castle",
	"rainbow_course":  "Rainbow Road",
	"old_mario_sfc":   "SNES Mario Circuit 1",
	"old_momo_64":     "N64 Moo Moo Farm",
	"old_peach_agb":   "GBA Peach Circuit",
	"old_luigi_gc":    "GCN Luigi Circuit",
	"old_donut_sfc":   "SNES Donut Plains 1",
	"old_frappe_64":   "N64 Frappe Snowland",
	"old_koopa_agb":   "GBA Bowser Castle 2",
	"old_baby_gc":     "GCN Baby Park",
	"old_noko_sfc":    "SNES Koopa Beach 2",
	"old_choco_64":    "N64 Choco Mountain",
	"old_luigi_agb":   "GBA Luigi Circuit",
	"old_kinoko_gc":   "GCN Mushroom Bridge",
	"old_choco_sfc":   "SNES Choco Island 2",
	"old_hyudoro_64":  "N64 Banshee Boardwalk",
	"old_sky_agb":     "GBA Sky Garden",
	"old_yoshi_gc":    "GCN Yoshi Circuit",
	"old_luigi_gcD":   "GCN Luigi Circuit (multiplayer)",
	"old_momo_64D":    "N64 Moo Moo Farm (multiplayer)",
	"old_mario_gc":    "old_mario_gc (unused)", // no shipped GCN Mario Circuit retro; leftover (no skybox)
}

// charNames maps the two-letter character codes in model file names.
var charNames = map[string]string{
	"MR": "Mario", "LG": "Luigi", "PC": "Peach", "DS": "Daisy",
	"YS": "Yoshi", "KO": "Toad", "KP": "Bowser", "DK": "Donkey Kong",
	"WR": "Wario", "WL": "Waluigi", "KA": "Dry Bones", "RB": "R.O.B.",
}

// writeManifest emits models.json for the viewer site. Each course entry links its
// far "_V" model as a camera-locked skybox and its enemy drive line as a fly-along
// path, so the viewer shows a track inside its backdrop rather than listing the
// pieces separately.
func writeManifest(outDir string, files, pathFiles []string) error {
	type entry struct {
		Name    string `json:"name"`
		File    string `json:"file"`
		Section string `json:"section"`
		Skybox  string `json:"skybox,omitempty"` // "_V" backdrop GLB, camera-locked
		Path    string `json:"path,omitempty"`   // enemy drive-line JSON
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
	// Courses: the main scene, with its far "_V" model folded in as a camera-locked
	// skybox and its enemy drive line as a fly-along path (both keyed by archive
	// stem), so a track lists as one entry that shows inside its backdrop.
	known := map[string]bool{}
	for _, e := range out {
		known[e.File] = true
	}
	// archive stem of a course file: "<archive>-<model>.glb" → archive (or the whole
	// stem for single-model archives), and whether this file is the "_V" skybox.
	courseKey := func(f string) (archive string, isV bool) {
		s := strings.TrimSuffix(f, ".glb")
		model := s
		if i := strings.Index(s, "-"); i >= 0 {
			archive, model = s[:i], s[i+1:]
		} else {
			archive = s
		}
		return archive, strings.HasSuffix(model, "_V")
	}
	isCourseFile := func(f string) bool {
		if strings.Contains(f, "course") || strings.Contains(f, "stage") || strings.Contains(f, "mini_") {
			return true
		}
		// Retro ("old_") tracks name their archive after the track (old_baby_gc), so
		// they carry no "course"/"stage" substring; match the prefix and known stems.
		a, _ := courseKey(f)
		_, known := courseNames[a]
		return known || strings.HasPrefix(a, "old_")
	}
	skyboxOf := map[string]string{} // archive stem -> "_V" GLB file
	for _, f := range files {
		if a, v := courseKey(f); v && isCourseFile(f) {
			skyboxOf[a] = f
		}
	}
	pathOf := map[string]bool{}
	for _, p := range pathFiles {
		pathOf[strings.TrimSuffix(p, ".path.json")] = true
	}
	for _, f := range files {
		if known[f] || !isCourseFile(f) {
			continue
		}
		base, isV := courseKey(f)
		if isV {
			known[f] = true // shown as a skybox on its course entry, not on its own
			continue
		}
		name := courseNames[base]
		if name == "" {
			name = base
		}
		sec := "Courses"
		if strings.HasPrefix(base, "mini_") || strings.HasPrefix(base, "MR_") {
			sec = "Battle & mission stages"
		} else if strings.HasPrefix(base, "old_") {
			sec = "Retro courses"
		}
		e := entry{Name: name, File: f, Section: sec, Skybox: skyboxOf[base]}
		if pathOf[base] {
			e.Path = base + ".path.json"
		}
		add(e)
		known[f] = true
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

func export(path, outDir string) ([]string, string, error) {
	models, err := mkds.LoadModels(path)
	if err != nil {
		return nil, "", err
	}
	texs := mkds.LoadTextures(path)
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	isCourse := strings.Contains(path, "/Course/")
	var written []string
	for _, m := range models {
		if isCourse {
			// Courses: export only the course scene itself — the main model plus its
			// "_V" skybox and "_wate" water companions; the small archive-mates are
			// map objects. The retro ("old_") tracks name their scene after the
			// archive stem (old_baby_gc, old_baby_gc_V) rather than "_course", so
			// match the stem too or they'd be dropped.
			keep := strings.Contains(m.Name, "_course") || strings.Contains(m.Name, "_stage") ||
				m.Name == stem || strings.HasPrefix(m.Name, stem+"_")
			if !keep {
				continue
			}
		}
		glb, err := nitro.ExportGLB(m, texs)
		if err != nil {
			continue
		}
		name := stem + ".glb"
		if len(models) > 1 {
			name = stem + "-" + m.Name + ".glb"
		}
		if err := os.WriteFile(filepath.Join(outDir, name), glb, 0o644); err != nil {
			return written, "", err
		}
		written = append(written, name)
		fmt.Printf("  %-40s %7d bytes\n", name, len(glb))
	}

	// Courses carry an NKM course map: export the CPU drive line (EPOI) as a
	// side-car polyline the viewer can fly the camera along. glTF has no notion of
	// a camera path, so it rides alongside the GLB as JSON in the same space.
	var pathFile string
	if isCourse && len(written) > 0 {
		if pf, err := exportPath(path, stem, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: path: %v\n", stem, err)
		} else {
			pathFile = pf
		}
	}
	return written, pathFile, nil
}

// worldPerGLBUnit is the engine's course scale: NKM/collision coordinates are kart-
// world units and the renderer scales course geometry down by this when building the
// model, so a GLB-space coordinate is world/16 (see mkds.NKM and cmd/trackmap).
const worldPerGLBUnit = 16.0

// exportPath writes "<stem>.path.json": the enemy drive line converted to GLB space
// (world/16), the coordinate frame the exported course GLB lives in. Returns "" (no
// file) when the course has no enemy line.
func exportPath(archive, stem, outDir string) (string, error) {
	nkm, err := mkds.LoadNKM(archive)
	if err != nil || nkm == nil {
		return "", err
	}
	pts, loop := nkm.EnemyLoop()
	if len(pts) < 2 {
		return "", nil
	}
	out := make([][3]float64, len(pts))
	for i, p := range pts {
		out[i] = [3]float64{p.X / worldPerGLBUnit, p.Y / worldPerGLBUnit, p.Z / worldPerGLBUnit}
	}
	doc := map[string]interface{}{
		"course": stem,
		"space":  "glb", // same coordinate frame/scale as the course .glb
		"loop":   loop,
		"points": out,
	}
	buf, err := json.MarshalIndent(doc, "", " ")
	if err != nil {
		return "", err
	}
	name := stem + ".path.json"
	if err := os.WriteFile(filepath.Join(outDir, name), buf, 0o644); err != nil {
		return "", err
	}
	fmt.Printf("  %-40s %7d pts%s\n", name, len(out), map[bool]string{true: " (loop)", false: ""}[loop])
	return name, nil
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "exportglb:", err)
	os.Exit(1)
}
