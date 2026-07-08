// models + levels stages: NSBMD → GLB export (as cmd/exportglb) plus the
// format-2 level envelope for each course. The GLB export path is shared by both
// stages; the levels stage additionally writes the mesh3d level file, its OBJI
// object database, CPU drive-line and texture-animation side-cars.
package main

import (
	"encoding/json"
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

// item is one exported GLB and how it classifies for the manifest (as exportglb).
type item struct {
	File      string // GLB filename (under <o>/models)
	Model     string // NSBMD model name inside the archive
	Course    string // course-archive stem ("" for menu models)
	Scene     bool   // the course scene itself (not a skybox, not a map object)
	Skybox    bool   // the "_V" camera-relative backdrop
	Billboard bool   // a camera-facing sprite (flat-in-Z quad, e.g. the Goomba)
}

// bindings is objectID → model base name, decoded from the ARM9 descriptor table.
var bindings map[int]string

// objectBindings decodes the ARM9's map-object descriptor table (Part V §2).
func objectBindings(arm9 []byte) map[int]string { return mkds.ObjectModelBindings(arm9) }

// worldPerGLBUnit is the engine's course scale: NKM/collision coordinates are kart-
// world units and the renderer scales course geometry down by this, so a GLB-space
// coordinate is world/16 (see mkds.NKM and cmd/exportglb).
const worldPerGLBUnit = 16.0

// itemboxFile is the shared itembox GLB (object 0x65), exported from
// data/Main/MapObj.carc — the one placed object that lives outside the course
// archives.
const itemboxFile = "MapObj-itembox.glb"

// courseArchive records the on-disk .carc path for each course stem, so the
// levels stage can re-open the NKM for its OBJI/PATH/BTA0 side-cars.
var courseArchive = map[string]string{}

// exportAllGLBs sweeps the menu kart+character sets, every course archive (scene,
// "_V" skybox, map objects) and the shared MapObj itembox into modelsDir, exactly
// the set exportglb -all exports. Returns the classified items.
func exportAllGLBs(root, modelsDir string) []item {
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		die(err)
	}
	var paths []string
	filepath.Walk(filepath.Join(root, "data", "KartModelMenu"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(p), ".nsbmd") &&
			!strings.Contains(p, "shadow") {
			paths = append(paths, p)
		}
		return nil
	})
	filepath.Walk(filepath.Join(root, "data", "Course"), func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".carc") || strings.Contains(p, "Tex") {
			return nil
		}
		paths = append(paths, p)
		return nil
	})
	sort.Strings(paths)
	paths = append(paths, filepath.Join(root, "data", "Main", "MapObj.carc"))

	fmt.Fprintf(os.Stderr, "[models] exporting GLBs from %d archives…\n", len(paths))
	var items []item
	for _, p := range paths {
		its, err := export(p, modelsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[models]  %s: %v\n", filepath.Base(p), err)
			continue
		}
		items = append(items, its...)
	}
	fmt.Fprintf(os.Stderr, "[models] %d GLBs written\n", len(items))
	return items
}

// export decodes one archive/model file and writes its GLBs (as exportglb.export),
// using deterministic filenames the format-2 tree references: a course scene is
// "<stem>.glb", its "_V" skybox "<skyModel>.glb", a map object "<stem>-<model>.glb",
// and a menu model "<stem>.glb" (or "<stem>-<model>.glb" for multi-model archives).
func export(path, modelsDir string) ([]item, error) {
	models, err := mkds.LoadModels(path)
	if err != nil {
		return nil, err
	}
	texs := mkds.LoadTextures(path)
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	isCourse := strings.Contains(path, "/Course/")
	isShared := strings.HasSuffix(path, "MapObj.carc")
	if isCourse {
		courseArchive[stem] = path
	}
	var items []item
	for _, m := range models {
		it := item{Model: m.Name, File: stem + ".glb"}
		if isCourse {
			it.Course = stem
			switch {
			case strings.Contains(strings.ToLower(m.Name), "shadow"):
				continue
			case strings.HasSuffix(m.Name, "_V"):
				it.Skybox = true
				it.File = m.Name + ".glb"
			case m.Name == stem || strings.HasSuffix(m.Name, "_course") || strings.HasSuffix(m.Name, "_stage"):
				it.Scene = true
				it.File = stem + ".glb"
			default:
				it.Billboard = isBillboard(m)
				it.File = stem + "-" + m.Name + ".glb"
			}
		} else if isShared {
			if m.Name != "itembox" {
				continue
			}
			it.File = itemboxFile
		} else {
			it.Scene = true
			if len(models) > 1 {
				it.File = stem + "-" + m.Name + ".glb"
			}
		}
		glb, err := nitro.ExportGLB(m, texs)
		if err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(modelsDir, it.File), glb, 0o644); err != nil {
			return items, err
		}
		items = append(items, it)
	}
	return items, nil
}

// ---------------------------------------------------------------------------
// manifest models[]
// ---------------------------------------------------------------------------

// charOrder is the canonical character order (as exportglb.writeManifest).
var charOrder = []string{"MR", "LG", "PC", "DS", "YS", "KO", "KP", "DK", "WR", "WL", "KA", "RB"}

var charNames = map[string]string{
	"MR": "Mario", "LG": "Luigi", "PC": "Peach", "DS": "Daisy",
	"YS": "Yoshi", "KO": "Toad", "KP": "Bowser", "DK": "Donkey Kong",
	"WR": "Wario", "WL": "Waluigi", "KA": "Dry Bones", "RB": "R.O.B.",
}

// buildModels classifies the non-course models into the manifest models[] array
// (Characters, Karts, per-course map objects, the shared itembox, then leftovers),
// mirroring exportglb.writeManifest. Course scenes go to levels[] and skyboxes fold
// into the level's "sky", so neither is listed here.
func buildModels(items []item) []manifestModel {
	byFile := map[string]item{}
	for _, it := range items {
		byFile[it.File] = it
	}
	var models []manifestModel
	known := map[string]bool{}
	add := func(name, section, file string) {
		models = append(models, manifestModel{Name: name, Section: section, File: "models/" + file})
		known[file] = true
	}

	// Characters, then karts, in character order.
	for _, cc := range charOrder {
		f := "P_" + cc + ".glb"
		if _, ok := byFile[f]; ok {
			add(charNames[cc], "Characters", f)
		}
	}
	for _, cc := range charOrder {
		for _, v := range []string{"a", "b", "c"} {
			f := "kart_" + cc + "_" + v + ".glb"
			if _, ok := byFile[f]; ok {
				add(charNames[cc]+" — kart "+strings.ToUpper(v), "Karts", f)
			}
		}
	}

	// Per-course map objects, grouped under the track's display name.
	perCourse := map[string][]item{}
	var stems []string
	for _, it := range items {
		if it.Course == "" || it.Scene || it.Skybox {
			continue
		}
		if _, seen := perCourse[it.Course]; !seen {
			stems = append(stems, it.Course)
		}
		perCourse[it.Course] = append(perCourse[it.Course], it)
	}
	sort.Slice(stems, func(i, j int) bool {
		if ci, cj := courseClass(stems[i]), courseClass(stems[j]); ci != cj {
			return ci < cj
		}
		return displayName(stems[i]) < displayName(stems[j])
	})
	for _, stem := range stems {
		sec := displayName(stem)
		objs := perCourse[stem]
		sort.Slice(objs, func(i, j int) bool { return objs[i].Model < objs[j].Model })
		for _, o := range objs {
			if known[o.File] {
				continue
			}
			add(sec+" · "+o.Model, sec, o.File)
		}
	}

	// The shared itembox, then any leftover menu models.
	if _, ok := byFile[itemboxFile]; ok && !known[itemboxFile] {
		add("Item box", "Shared objects", itemboxFile)
	}
	var rest []string
	for _, it := range items {
		if it.Scene || it.Skybox || known[it.File] {
			continue
		}
		rest = append(rest, it.File)
	}
	sort.Strings(rest)
	for _, f := range rest {
		if known[f] {
			continue
		}
		add(strings.TrimSuffix(f, ".glb"), "Other", f)
	}
	fmt.Fprintf(os.Stderr, "[models] %d models classified for the manifest\n", len(models))
	return models
}

// ---------------------------------------------------------------------------
// course display names / ordering (as exportglb)
// ---------------------------------------------------------------------------

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
	"old_mario_gc":    "old_mario_gc (unused)",
}

func displayName(stem string) string {
	if n := courseNames[stem]; n != "" {
		return n
	}
	return stem
}

// courseClass orders the track sections: nitro cups, then the retro cups, then
// the battle/mission stages and credits scenes (as exportglb).
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

// isBillboard reports whether a model is a camera-facing sprite (as exportglb).
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

// ---------------------------------------------------------------------------
// levels: format-2 mesh3d envelope + OBJI/path/anims side-cars
// ---------------------------------------------------------------------------

// buildLevels writes one format-2 level per course scene into <o>/levels, with its
// OBJI object database, CPU drive-line and texture-animation side-cars, and returns
// the manifest levels[] array.
func buildLevels(out string, items []item) []manifestLevel {
	levelsDir := filepath.Join(out, "levels")
	if err := os.MkdirAll(levelsDir, 0o755); err != nil {
		die(err)
	}

	// Per-course: the scene and its skybox items.
	scene := map[string]item{}
	sky := map[string]item{}
	byModel := map[string]map[string]item{} // course -> lower(model) -> item
	var stems []string
	for _, it := range items {
		if it.Course == "" {
			continue
		}
		switch {
		case it.Scene:
			if _, seen := scene[it.Course]; !seen {
				scene[it.Course] = it
				stems = append(stems, it.Course)
			}
		case it.Skybox:
			sky[it.Course] = it
		default:
			if byModel[it.Course] == nil {
				byModel[it.Course] = map[string]item{}
			}
			byModel[it.Course][strings.ToLower(it.Model)] = it
		}
	}
	sort.Slice(stems, func(i, j int) bool {
		if ci, cj := courseClass(stems[i]), courseClass(stems[j]); ci != cj {
			return ci < cj
		}
		return displayName(stems[i]) < displayName(stems[j])
	})

	var manifest []manifestLevel
	for _, stem := range stems {
		sc := scene[stem]
		lf := map[string]any{
			"format": 2,
			"name":   displayName(stem),
			"kind":   "mesh3d",
			"mesh":   map[string]string{"glb": "models/" + sc.File},
		}
		if sk, ok := sky[stem]; ok {
			lf["sky"] = "models/" + sk.File
		}

		archive := courseArchive[stem]
		if objFile := exportObjects(archive, stem, levelsDir, byModel[stem]); objFile != "" {
			lf["objectsFile"] = objFile
		}
		if pathFile := exportPath(archive, stem, levelsDir); pathFile != "" {
			lf["path"] = pathFile
		}
		if animFile := exportAnims(archive, stem, levelsDir); animFile != "" {
			lf["anims"] = animFile
		}

		buf, _ := json.MarshalIndent(lf, "", " ")
		if err := os.WriteFile(filepath.Join(levelsDir, stem+".json"), buf, 0o644); err != nil {
			die(err)
		}
		m := manifestLevel{
			Name:    displayName(stem),
			Section: "Courses",
			File:    "levels/" + stem + ".json",
			Kind:    "mesh3d",
		}
		if of, _ := lf["objectsFile"].(string); of != "" {
			m.Objects = "levels/" + of
		}
		manifest = append(manifest, m)
		fmt.Fprintf(os.Stderr, "[levels]  %d  %s\n", len(manifest), stem)
	}
	fmt.Fprintf(os.Stderr, "[levels] %d course levels\n", len(manifest))
	return manifest
}

// exportObjects writes "<stem>.objects.json" under levelsDir: the format-2 object
// database wrapping the course's OBJI placements (as exportglb.exportObjects, but
// each placement's GLB "file" repointed under models/ and the doc wrapped in the
// {format,level,objects,...} envelope). Returns the filename or "".
func exportObjects(archive, stem, levelsDir string, byModel map[string]item) string {
	if len(bindings) == 0 || archive == "" {
		return ""
	}
	nkm, err := mkds.LoadNKM(archive)
	if err != nil || nkm == nil {
		return ""
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
	// The itembox spawns 12.0 world units above its OBJI position (its init adds the
	// fx32 constant 0xC000 to Y); other ground objects place as-authored.
	const itemboxHover = 12.0 / worldPerGLBUnit
	var placed []placement
	var routes []route
	routeIdx := map[int]int{}
	skipped := map[string]int{}
	for _, o := range nkm.Objects {
		var file string
		var billboard bool
		if o.ID == 0x65 {
			file = "models/" + itemboxFile
		} else if name, ok := bindings[o.ID]; ok {
			if it, ok := byModel[name]; ok {
				file, billboard = "models/"+it.File, it.Billboard
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
		return ""
	}
	doc := map[string]any{
		"format":  2,
		"level":   stem,
		"space":   "glb", // same coordinate frame/scale as the course .glb
		"objects": placed,
		"routes":  routes,
		"skipped": skipped,
	}
	buf, _ := json.MarshalIndent(doc, "", " ")
	name := stem + ".objects.json"
	if err := os.WriteFile(filepath.Join(levelsDir, name), buf, 0o644); err != nil {
		die(err)
	}
	return name
}

// exportPath writes "<stem>.path.json" under levelsDir: the enemy drive line in GLB
// space (as exportglb.exportPath, relocated under levels/). Returns "" if none.
func exportPath(archive, stem, levelsDir string) string {
	if archive == "" {
		return ""
	}
	nkm, err := mkds.LoadNKM(archive)
	if err != nil || nkm == nil {
		return ""
	}
	pts, loop := nkm.EnemyLoop()
	if len(pts) < 2 {
		return ""
	}
	out := make([][3]float64, len(pts))
	for i, p := range pts {
		out[i] = [3]float64{p.X / worldPerGLBUnit, p.Y / worldPerGLBUnit, p.Z / worldPerGLBUnit}
	}
	doc := map[string]any{
		"course": stem,
		"space":  "glb",
		"loop":   loop,
		"points": out,
	}
	buf, _ := json.MarshalIndent(doc, "", " ")
	name := stem + ".path.json"
	if err := os.WriteFile(filepath.Join(levelsDir, name), buf, 0o644); err != nil {
		die(err)
	}
	return name
}

// exportAnims writes "<stem>.anims.json" under levelsDir: the course's BTA0
// texture-SRT animations (as exportglb.exportAnims). Returns "" if none.
func exportAnims(archive, stem, levelsDir string) string {
	if archive == "" {
		return ""
	}
	raw, err := os.ReadFile(archive)
	if err != nil {
		return ""
	}
	files, err := nds.ParseNARC(nds.Decompress(raw))
	if err != nil {
		return ""
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
		Rot      float64 `json:"rot,omitempty"`
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
			return ""
		}
		for _, a := range anims {
			out = append(out, anim{
				Model: a.Model, Material: a.Material, Frames: a.Frames,
				ScaleS: conv(a.ScaleS), ScaleT: conv(a.ScaleT),
				TransS: conv(a.TransS), TransT: conv(a.TransT),
				Rot: math.Atan2(a.RotSin, a.RotCos),
			})
		}
	}
	if len(out) == 0 {
		return ""
	}
	buf, _ := json.MarshalIndent(map[string]any{"course": stem, "anims": out}, "", " ")
	name := stem + ".anims.json"
	if err := os.WriteFile(filepath.Join(levelsDir, name), buf, 0o644); err != nil {
		die(err)
	}
	return name
}
