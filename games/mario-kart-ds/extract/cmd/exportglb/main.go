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
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/mario-kart-ds/extract/mkds"

	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/nitro"
)

// item is one exported GLB and how it classifies for the manifest.
type item struct {
	File      string // GLB filename
	Model     string // NSBMD model name inside the archive
	Course    string // course-archive stem ("" for menu models)
	Scene     bool   // the course scene itself (not a skybox, not a map object)
	Skybox    bool   // the "_V" camera-relative backdrop
	Billboard bool   // a camera-facing sprite (flat-in-Z quad, e.g. the Goomba)
}

func main() {
	all := flag.Bool("all", false, "export the KartModelMenu kart + character sets")
	root := flag.String("root", "../extracted/files", "extracted filesystem root (-all)")
	outDir := flag.String("o", "../extracted/glb", "output directory")
	arm9 := flag.String("arm9", "../extracted/arm9_dec.bin", "decompressed ARM9 (for the OBJI object-ID bindings)")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	// The object-ID → model-name bindings from the ARM9's map-object descriptor
	// table (Part V §2) drive the per-course placements JSON. Optional: without
	// the ARM9 image, models still export but no placements are written.
	if data, err := os.ReadFile(*arm9); err == nil {
		bindings = mkds.ObjectModelBindings(data)
	} else {
		fmt.Fprintf(os.Stderr, "  no ARM9 image (%v): skipping object placements\n", err)
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
		// the shared map-object archive: the itembox (object 0x65) lives here,
		// not in the course archives — every track places it
		paths = append(paths, filepath.Join(*root, "data", "Main", "MapObj.carc"))
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
		Skybox  string `json:"skybox,omitempty"`  // "_V" backdrop GLB, camera-locked
		Path    string `json:"path,omitempty"`    // enemy drive-line JSON
		Objects string `json:"objects,omitempty"` // OBJI placements JSON
		Anims   string `json:"anims,omitempty"`   // BTA0 texture-animation JSON
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
			e.Objects = objectFiles[stem]
			e.Anims = animFiles[stem]
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

	// The shared itembox (placed on every track but living outside the course
	// archives), then anything left (loose menu models like select, the tires).
	if _, ok := byFile[itemboxFile]; ok {
		add(entry{Name: "Item box", File: itemboxFile, Section: "Shared objects"})
		known[itemboxFile] = true
	}
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
	isShared := strings.HasSuffix(path, "MapObj.carc")
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
			default:
				it.Billboard = isBillboard(m)
			}
		} else if isShared {
			// the shared map-object archive: only the itembox is a placed object
			// (the rest are its shards, shadows and inner pieces)
			if m.Name != "itembox" {
				continue
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
	// side-car polyline, and the OBJI object placements bound to the GLBs above.
	// glTF has no notion of either, so they ride alongside as JSON in GLB space.
	var pathFile string
	if isCourse && len(items) > 0 {
		if pf, err := exportPath(path, stem, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: path: %v\n", stem, err)
		} else {
			pathFile = pf
		}
		if err := exportObjects(path, stem, outDir, items); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: objects: %v\n", stem, err)
		}
		if err := exportAnims(path, stem, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: anims: %v\n", stem, err)
		}
	}
	return items, pathFile, nil
}

// exportAnims writes "<stem>.anims.json": the course's BTA0 texture-SRT
// animations (scrolling water, waterfalls, boost-panel arrows) as per-material
// tracks the viewer replays at 60 fps. Values are normalized texture space
// (1.0 = one wrap), straight from the fx12 data.
func exportAnims(archive, stem, outDir string) error {
	raw, err := os.ReadFile(archive)
	if err != nil {
		return err
	}
	files, err := nds.ParseNARC(nds.Decompress(raw))
	if err != nil {
		return nil
	}
	type track struct {
		Const   float64   `json:"const,omitempty"`
		Samples []float64 `json:"samples,omitempty"`
		Step    int       `json:"step,omitempty"`
	}
	type anim struct {
		Model    string  `json:"model"`
		Material string  `json:"material"`
		Frames   int     `json:"frames"`
		ScaleS   track   `json:"scaleS,omitempty"`
		ScaleT   track   `json:"scaleT,omitempty"`
		Rot      float64 `json:"rot,omitempty"` // radians, constant
		TransS   track   `json:"transS"`
		TransT   track   `json:"transT"`
	}
	conv := func(t nitro.Track) track { return track{Const: t.Const, Samples: t.Samples, Step: t.Step} }
	var out []anim
	for _, f := range files {
		if len(f) < 4 || string(f[:4]) != "BTA0" {
			continue
		}
		anims, err := nitro.DecodeNSBTA(f)
		if err != nil {
			return err
		}
		for _, a := range anims {
			e := anim{
				Model: a.Model, Material: a.Material, Frames: a.Frames,
				ScaleS: conv(a.ScaleS), ScaleT: conv(a.ScaleT),
				TransS: conv(a.TransS), TransT: conv(a.TransT),
				Rot: math.Atan2(a.RotSin, a.RotCos),
			}
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		return nil
	}
	buf, err := json.MarshalIndent(map[string]interface{}{"course": stem, "anims": out}, "", " ")
	if err != nil {
		return err
	}
	name := stem + ".anims.json"
	if err := os.WriteFile(filepath.Join(outDir, name), buf, 0o644); err != nil {
		return err
	}
	fmt.Printf("  %-40s %d material anims\n", name, len(out))
	animFiles[stem] = name
	return nil
}

// animFiles records which courses got a texture-animation file, for the manifest.
var animFiles = map[string]string{}

// isBillboard reports whether a model is a camera-facing sprite: a handful of
// triangles authored flat in Z (facing +Z), like the Goomba and the Pianta. The
// engine turns these to the camera; the viewer yaws them around world-up.
func isBillboard(m nitro.Model) bool {
	var tris []nitro.Tri
	for _, d := range nitro.RunSBC(m) {
		if d.Shape < len(m.Shapes) {
			tris = append(tris, nitro.DecodeDL(m.Shapes[d.Shape].DL, d.Stack, d.M, d.Mat)...)
		}
	}
	if len(tris) == 0 || len(tris) > 8 {
		return false
	}
	minX, maxX := math.Inf(1), math.Inf(-1)
	minY, maxY := math.Inf(1), math.Inf(-1)
	minZ, maxZ := math.Inf(1), math.Inf(-1)
	for _, t := range tris {
		for _, v := range t.V {
			minX, maxX = math.Min(minX, v.X), math.Max(maxX, v.X)
			minY, maxY = math.Min(minY, v.Y), math.Max(maxY, v.Y)
			minZ, maxZ = math.Min(minZ, v.Z), math.Max(maxZ, v.Z)
		}
	}
	span := math.Max(maxX-minX, maxY-minY)
	return span > 0 && (maxZ-minZ) < 0.05*span
}

// bindings is objectID → model base name, decoded from the ARM9 descriptor
// table in main (empty if no ARM9 image was given).
var bindings map[int]string

// itemboxFile is the shared itembox GLB (object 0x65), exported from
// data/Main/MapObj.carc — the one placed object that lives outside the
// course archives.
const itemboxFile = "MapObj-itembox.glb"

// exportObjects writes "<stem>.objects.json": every OBJI placement whose object
// ID resolves to an exported model — position in GLB space (world/16), rotation
// in degrees, scale, and the billboard flag. Placements whose ID has no known
// model (sound emitters, effect points, mode-specific objects) are counted in
// "skipped" rather than silently dropped.
func exportObjects(archive, stem, outDir string, items []item) error {
	if len(bindings) == 0 {
		return nil
	}
	nkm, err := mkds.LoadNKM(archive)
	if err != nil || nkm == nil {
		return err
	}
	byModel := map[string]item{}
	for _, it := range items {
		if !it.Scene && !it.Skybox {
			byModel[strings.ToLower(it.Model)] = it
		}
	}
	// Rot/Scale are slices, not arrays: encoding/json's omitempty never omits an
	// array, and a serialized zero scale would collapse every placement to nothing.
	type placement struct {
		File      string     `json:"file"`
		Pos       [3]float64 `json:"pos"`
		Rot       []float64  `json:"rot,omitempty"`   // Euler degrees, Y-dominant
		Scale     []float64  `json:"scale,omitempty"` // omitted when 1,1,1
		Billboard bool       `json:"billboard,omitempty"`
		Route     *int       `json:"route,omitempty"` // index into "routes"
	}
	type route struct {
		Loop   bool         `json:"loop"`
		Points [][3]float64 `json:"points"`
	}
	// The itembox spawns 12.0 world units above its OBJI position — its init
	// ($020DE6B0) adds the fx32 constant 0xC000 to Y before anything else. That
	// is why the boxes float; other ground objects place as-authored.
	const itemboxHover = 12.0 / worldPerGLBUnit
	var placed []placement
	var routes []route
	routeIdx := map[int]int{} // NKM path index -> routes[] index
	skipped := map[string]int{}
	for _, o := range nkm.Objects {
		var file string
		var billboard bool
		if o.ID == 0x65 {
			file = itemboxFile
		} else if name, ok := bindings[o.ID]; ok {
			if it, ok := byModel[name]; ok {
				file, billboard = it.File, it.Billboard
			}
		}
		if file == "" {
			skipped[fmt.Sprintf("0x%03X", o.ID)]++
			continue
		}
		p := placement{
			File:      file,
			Pos:       [3]float64{o.Pos.X / worldPerGLBUnit, o.Pos.Y / worldPerGLBUnit, o.Pos.Z / worldPerGLBUnit},
			Billboard: billboard,
		}
		if o.ID == 0x65 {
			p.Pos[1] += itemboxHover
		}
		if o.Rot.X != 0 || o.Rot.Y != 0 || o.Rot.Z != 0 {
			p.Rot = []float64{o.Rot.X, o.Rot.Y, o.Rot.Z}
		}
		if o.Scale.X != 1 || o.Scale.Y != 1 || o.Scale.Z != 1 {
			p.Scale = []float64{o.Scale.X, o.Scale.Y, o.Scale.Z}
		}
		// Route-following objects: attach their PATH/POIT polyline (GLB space).
		if o.RouteID >= 0 && o.RouteID < len(nkm.Paths) && len(nkm.Paths[o.RouteID].Points) >= 2 {
			ri, ok := routeIdx[o.RouteID]
			if !ok {
				pth := nkm.Paths[o.RouteID]
				r := route{Loop: pth.Loop, Points: make([][3]float64, len(pth.Points))}
				for i, pt := range pth.Points {
					r.Points[i] = [3]float64{pt.X / worldPerGLBUnit, pt.Y / worldPerGLBUnit, pt.Z / worldPerGLBUnit}
				}
				ri = len(routes)
				routes = append(routes, r)
				routeIdx[o.RouteID] = ri
			}
			p.Route = &ri
		}
		placed = append(placed, p)
	}
	if len(placed) == 0 {
		return nil
	}
	doc := map[string]interface{}{
		"course":  stem,
		"space":   "glb",
		"objects": placed,
		"routes":  routes,
		"skipped": skipped,
	}
	buf, err := json.MarshalIndent(doc, "", " ")
	if err != nil {
		return err
	}
	name := stem + ".objects.json"
	if err := os.WriteFile(filepath.Join(outDir, name), buf, 0o644); err != nil {
		return err
	}
	fmt.Printf("  %-40s %d placed, %d skipped\n", name, len(placed), len(nkm.Objects)-len(placed))
	objectFiles[stem] = name
	return nil
}

// objectFiles records which courses got a placements file, for the manifest.
var objectFiles = map[string]string{}

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
