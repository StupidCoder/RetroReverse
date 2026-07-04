// exportglb converts NSBMD models to standard binary glTF (.glb) via
// nitro.ExportGLB, pairing each model with the textures of its companion NSBTX
// (same stem, then siblings). It accepts single files or, with -all, sweeps the
// menu kart and character model sets plus every course archive — the course
// scene, its "_V" skybox, and the course's map-object models (the archive-mates
// the engine spawns from the NKM's OBJI section, Part V §2) — into an output
// directory: the models the viewer site serves. Courses also get their CPU
// drive line as a side-car "<stem>.path.json" (the NKM EPOI line in GLB space).
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

// item is one exported GLB and how it classifies for the manifest.
type item struct {
	File   string // GLB filename
	Model  string // NSBMD model name inside the archive
	Course string // course-archive stem ("" for menu models)
	Scene  bool   // the course scene itself (not a skybox, not a map object)
	Skybox bool   // the "_V" camera-relative backdrop
}

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
		// course archives (their Tex siblings carry only textures)
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

	var items []item
	var pathFiles []string
	for _, p := range paths {
		its, pf, err := export(p, *outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", filepath.Base(p), err)
			continue
		}
		items = append(items, its...)
		if pf != "" {
			pathFiles = append(pathFiles, pf)
		}
	}
	if *all {
		if err := writeManifest(*outDir, items, pathFiles); err != nil {
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

// courseClass orders the track sections: nitro cups, then the retro cups, then
// the battle/mission stages and credits scenes.
func courseClass(stem string) int {
	switch {
	case strings.HasPrefix(stem, "old_"):
		return 1
	case strings.HasPrefix(stem, "mini_") || strings.HasPrefix(stem, "MR_") ||
		strings.HasPrefix(stem, "StaffRoll") || strings.HasPrefix(stem, "test"):
		return 2
	default:
		return 0
	}
}

// writeManifest emits models.json for the viewer site. Every track is its own
// section holding the track scene (with its "_V" skybox camera-locked and its
// CPU drive line attached) followed by the course's map-object models, so a
// track and the objects the engine spawns onto it browse side by side.
func writeManifest(outDir string, items []item, pathFiles []string) error {
	type entry struct {
		Name    string `json:"name"`            // full name (HUD, deep-link slug)
		Label   string `json:"label,omitempty"` // short list label within the section
		File    string `json:"file"`
		Section string `json:"section"`
		Skybox  string `json:"skybox,omitempty"` // "_V" backdrop GLB, camera-locked
		Path    string `json:"path,omitempty"`   // enemy drive-line JSON
	}
	var out []entry
	add := func(e entry) { out = append(out, e) }
	known := map[string]bool{}

	// Characters first, then karts, in character order.
	byFile := map[string]item{}
	for _, it := range items {
		byFile[it.File] = it
	}
	var charOrder = []string{"MR", "LG", "PC", "DS", "YS", "KO", "KP", "DK", "WR", "WL", "KA", "RB"}
	for _, cc := range charOrder {
		f := "P_" + cc + ".glb"
		if _, ok := byFile[f]; ok {
			add(entry{Name: charNames[cc], File: f, Section: "Characters"})
			known[f] = true
		}
	}
	for _, cc := range charOrder {
		for _, v := range []string{"a", "b", "c"} {
			f := "kart_" + cc + "_" + v + ".glb"
			if _, ok := byFile[f]; ok {
				add(entry{Name: charNames[cc] + " — kart " + strings.ToUpper(v), File: f, Section: "Karts"})
				known[f] = true
			}
		}
	}

	// Tracks: one section per course archive — the scene entry first, then its
	// map objects. Skyboxes fold into the scene entry rather than listing.
	pathOf := map[string]bool{}
	for _, p := range pathFiles {
		pathOf[strings.TrimSuffix(p, ".path.json")] = true
	}
	perCourse := map[string][]item{}
	var stems []string
	for _, it := range items {
		if it.Course == "" {
			continue
		}
		if _, seen := perCourse[it.Course]; !seen {
			stems = append(stems, it.Course)
		}
		perCourse[it.Course] = append(perCourse[it.Course], it)
	}
	sort.Slice(stems, func(i, j int) bool {
		ci, cj := courseClass(stems[i]), courseClass(stems[j])
		if ci != cj {
			return ci < cj
		}
		return displayName(stems[i]) < displayName(stems[j])
	})
	for _, stem := range stems {
		section := displayName(stem)
		var scene *item
		skybox := ""
		var objs []item
		for i, it := range perCourse[stem] {
			switch {
			case it.Skybox:
				skybox = it.File
			case it.Scene && scene == nil:
				scene = &perCourse[stem][i]
			default:
				objs = append(objs, it)
			}
		}
		if scene != nil {
			e := entry{Name: section, Label: "Track", File: scene.File, Section: section, Skybox: skybox}
			if pathOf[stem] {
				e.Path = stem + ".path.json"
			}
			add(e)
			known[scene.File] = true
		}
		sort.Slice(objs, func(i, j int) bool { return objs[i].Model < objs[j].Model })
		for _, o := range objs {
			add(entry{Name: section + " · " + o.Model, Label: o.Model, File: o.File, Section: section})
			known[o.File] = true
		}
		known[skybox] = true
	}

	// Anything left (loose menu models like select, the kart tires).
	var rest []string
	for _, it := range items {
		if !known[it.File] {
			rest = append(rest, it.File)
		}
	}
	sort.Strings(rest)
	for _, f := range rest {
		add(entry{Name: strings.TrimSuffix(f, ".glb"), File: f, Section: "Other"})
	}

	buf, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return err
	}
	p := filepath.Join(outDir, "models.json")
	fmt.Printf("  models.json: %d entries\n", len(out))
	return os.WriteFile(p, buf, 0o644)
}

func displayName(stem string) string {
	if n := courseNames[stem]; n != "" {
		return n
	}
	return stem
}

func export(path, outDir string) ([]item, string, error) {
	models, err := mkds.LoadModels(path)
	if err != nil {
		return nil, "", err
	}
	texs := mkds.LoadTextures(path)
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	isCourse := strings.Contains(path, "/Course/")
	var items []item
	for _, m := range models {
		it := item{Model: m.Name}
		if isCourse {
			it.Course = stem
			// Classify the archive-mates: the "_V" model is the skybox; the model
			// named like the archive (retro tracks) or "<x>_course"/"_stage" is the
			// course scene; everything else is a map object the engine spawns from
			// the NKM's OBJI section. Shadow blobs add nothing in the viewer.
			switch {
			case strings.Contains(strings.ToLower(m.Name), "shadow"):
				continue
			case strings.HasSuffix(m.Name, "_V"):
				it.Skybox = true
			case m.Name == stem || strings.HasSuffix(m.Name, "_course") || strings.HasSuffix(m.Name, "_stage"):
				it.Scene = true
			}
		} else {
			it.Scene = true
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
			return items, "", err
		}
		it.File = name
		items = append(items, it)
		fmt.Printf("  %-40s %7d bytes\n", name, len(glb))
	}

	// Courses carry an NKM course map: export the CPU drive line (EPOI) as a
	// side-car polyline the viewer can fly the camera along. glTF has no notion of
	// a camera path, so it rides alongside the GLB as JSON in the same space.
	var pathFile string
	if isCourse && len(items) > 0 {
		if pf, err := exportPath(path, stem, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: path: %v\n", stem, err)
		} else {
			pathFile = pf
		}
	}
	return items, pathFile, nil
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
