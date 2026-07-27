// objects + levels stages: NSBMD → GLB export plus the Retro-X documents.
// Every menu/course/map model becomes a model3d object asset (course scenes
// included — a level places its course at the origin so the BTA0 texture
// animations play through the object pipeline); each track becomes a scene3d
// level with its OBJI placements, NKM routes, camera-attached skybox and the
// CPU drive line baked as a toggleable line layer.
package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/mario-kart-ds/extract/mkds"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/nitro"
)

// item is one exported GLB and how it classifies.
type item struct {
	File      string   // GLB filename (under objects/, or levels/ for skyboxes)
	Model     string   // NSBMD model name inside the archive
	Course    string   // course-archive stem ("" for menu models)
	Scene     bool     // the course scene itself (not a skybox, not a map object)
	Skybox    bool     // the "_V" camera-relative backdrop
	Billboard bool     // a camera-facing sprite (flat-in-Z quad, e.g. the Goomba)
	Materials []string // material names in the exported GLB (uvAnim filter)
}

// bindings is objectID → model base name, decoded from the ARM9 descriptor table.
var bindings map[int]string

// objectBindings decodes the ARM9's map-object descriptor table (Part V §2).
func objectBindings(arm9 []byte) map[int]string { return mkds.ObjectModelBindings(arm9) }

// worldPerGLBUnit is the engine's course scale: NKM/collision coordinates are kart-
// world units and the renderer scales course geometry down by this, so a GLB-space
// coordinate is world/16.
const worldPerGLBUnit = 16.0

// itemboxFile is the shared itembox GLB (object 0x65), exported from
// data/Main/MapObj.carc — the one placed object that lives outside the course
// archives.
const itemboxFile = "MapObj-itembox.glb"

// courseArchive records the on-disk .carc path for each course stem, so the
// levels stage can re-open the NKM for its OBJI/PATH data and the objects
// stage the BTA0 texture animations.
var courseArchive = map[string]string{}

// exportAllGLBs sweeps the menu kart+character sets, every course archive (scene,
// "_V" skybox, map objects) and the shared MapObj itembox. Skyboxes land in
// levels/ (level-only geometry); everything else in objects/.
func exportAllGLBs(ctx *cli.Context, root string) ([]item, error) {
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

	var items []item
	for i, p := range paths {
		its, err := export(ctx, p)
		if err != nil {
			ctx.Logf("%s: %v", filepath.Base(p), err)
			continue
		}
		items = append(items, its...)
		if (i+1)%20 == 0 {
			ctx.Progress("objects", i+1, len(paths), fmt.Sprintf("%d GLBs so far", len(items)))
		}
	}
	ctx.Progress("objects", len(paths), len(paths), fmt.Sprintf("%d GLBs written", len(items)))
	return items, nil
}

// export decodes one archive/model file and writes its GLBs with deterministic
// filenames: a course scene is "<stem>.glb", its "_V" skybox "<skyModel>.glb"
// (under levels/), a map object "<stem>-<model>.glb", and a menu model
// "<stem>.glb" (or "<stem>-<model>.glb" for multi-model archives).
func export(ctx *cli.Context, path string) ([]item, error) {
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
		dir := "objects"
		if isCourse {
			it.Course = stem
			switch {
			case strings.Contains(strings.ToLower(m.Name), "shadow"):
				continue
			case strings.HasSuffix(m.Name, "_V"):
				it.Skybox = true
				it.File = m.Name + ".glb"
				dir = "levels"
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
			it.Scene = false
			if len(models) > 1 {
				it.File = stem + "-" + m.Name + ".glb"
			}
		}
		for _, mat := range m.Materials {
			it.Materials = append(it.Materials, mat.Name)
		}
		data, err := nitro.ExportGLB(m, texs)
		if err != nil {
			continue
		}
		p, err := ctx.Builder.Path(dir, it.File)
		if err != nil {
			return items, err
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			return items, err
		}
		items = append(items, it)
	}
	return items, nil
}

// ---------------------------------------------------------------------------
// objects stage
// ---------------------------------------------------------------------------

// charOrder is the canonical character order.
var charOrder = []string{"MR", "LG", "PC", "DS", "YS", "KO", "KP", "DK", "WR", "WL", "KA", "RB"}

var charNames = map[string]string{
	"MR": "Mario", "LG": "Luigi", "PC": "Peach", "DS": "Daisy",
	"YS": "Yoshi", "KO": "Toad", "KP": "Bowser", "DK": "Donkey Kong",
	"WR": "Wario", "WL": "Waluigi", "KA": "Dry Bones", "RB": "R.O.B.",
}

// buildObjects registers every non-skybox GLB as a model3d asset (Characters,
// Karts, course scenes, per-course map objects, the shared itembox, leftovers)
// and returns the GLB file -> asset id map the levels stage places through.
func buildObjects(ctx *cli.Context, items []item) (map[string]string, error) {
	b := ctx.Builder
	refs := map[string]string{}
	byFile := map[string]item{}
	for _, it := range items {
		if !it.Skybox {
			byFile[it.File] = it
		}
	}

	// BTA0 texture animations per course, keyed by model name.
	uvAnims := map[string]map[string][]schema.UVAnim{}
	for stem, archive := range courseArchive {
		uvAnims[stem] = courseUVAnims(archive)
	}

	usedIDs := map[string]bool{}
	add := func(it item, name, group string) {
		base := slugify(strings.TrimSuffix(it.File, ".glb"))
		if it.Scene && it.Course != "" {
			base += "-model" // the plain slug is the LEVEL's asset id
		}
		id := base
		for n := 2; usedIDs[id]; n++ {
			id = fmt.Sprintf("%s-%d", base, n)
		}
		usedIDs[id] = true
		doc := &schema.Object{
			Type: schema.ObjectModel3D, Name: name, Model: it.File,
			Props: map[string]any{"model": it.Model},
		}
		if it.Billboard {
			doc.Billboard = "yaw"
		}
		if it.Course != "" {
			// only animations whose material the exported GLB actually kept
			have := map[string]bool{}
			for _, mn := range it.Materials {
				have[mn] = true
			}
			for _, ua := range uvAnims[it.Course][it.Model] {
				if have[ua.Material] {
					doc.UVAnims = append(doc.UVAnims, ua)
				}
			}
		}
		b.AddObject(schema.Asset{ID: id, Name: name, Group: group}, doc)
		refs[it.File] = id
	}

	// Characters, then karts, in character order.
	done := map[string]bool{}
	for _, cc := range charOrder {
		if it, ok := byFile["P_"+cc+".glb"]; ok {
			add(it, charNames[cc], "Characters")
			done[it.File] = true
		}
	}
	for _, cc := range charOrder {
		for _, v := range []string{"a", "b", "c"} {
			if it, ok := byFile["kart_"+cc+"_"+v+".glb"]; ok {
				add(it, charNames[cc]+" — kart "+strings.ToUpper(v), "Karts")
				done[it.File] = true
			}
		}
	}

	// Course scenes + per-course map objects, grouped under the track's name.
	var stems []string
	perCourse := map[string][]item{}
	for _, it := range items {
		if it.Course == "" || it.Skybox {
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
		objs := perCourse[stem]
		sort.Slice(objs, func(i, j int) bool { return objs[i].Model < objs[j].Model })
		for _, o := range objs {
			if done[o.File] {
				continue
			}
			if o.Scene {
				add(o, displayName(stem), "Course models")
			} else {
				add(o, displayName(stem)+" · "+o.Model, displayName(stem))
			}
			done[o.File] = true
		}
	}

	// The shared itembox, then any leftover menu models.
	if it, ok := byFile[itemboxFile]; ok && !done[it.File] {
		add(it, "Item box", "Shared objects")
		done[it.File] = true
	}
	var rest []string
	for _, it := range items {
		if it.Skybox || done[it.File] {
			continue
		}
		rest = append(rest, it.File)
	}
	sort.Strings(rest)
	for _, f := range rest {
		if done[f] {
			continue
		}
		add(byFile[f], strings.TrimSuffix(f, ".glb"), "Other")
		done[f] = true
	}
	ctx.Logf("objects: %d assets", len(refs))
	return refs, nil
}

func slugify(s string) string {
	var out []rune
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
		default:
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	return strings.Trim(string(out), "-")
}

// courseUVAnims decodes the archive's BTA0 texture-SRT animations, keyed by
// model name.
func courseUVAnims(archive string) map[string][]schema.UVAnim {
	raw, err := os.ReadFile(archive)
	if err != nil {
		return nil
	}
	files, err := nds.ParseNARC(nds.Decompress(raw))
	if err != nil {
		return nil
	}
	conv := func(t nitro.Track) *schema.Channel {
		if len(t.Samples) > 0 {
			return &schema.Channel{Samples: t.Samples, Step: t.Step}
		}
		if t.Const != 0 {
			v := t.Const
			return &schema.Channel{Const: &v}
		}
		return nil
	}
	out := map[string][]schema.UVAnim{}
	for _, f := range files {
		if len(f) < 4 || string(f[:4]) != "BTA0" {
			continue
		}
		anims, err := nitro.DecodeNSBTA(f)
		if err != nil {
			return nil
		}
		for _, a := range anims {
			ua := schema.UVAnim{
				Material: a.Material, Frames: a.Frames,
				ScaleS: conv(a.ScaleS), ScaleT: conv(a.ScaleT),
				TransS: conv(a.TransS), TransT: conv(a.TransT),
			}
			if rot := math.Atan2(a.RotSin, a.RotCos); rot != 0 {
				ua.Rot = &schema.Channel{Const: &rot}
			}
			out[a.Model] = append(out[a.Model], ua)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// course display names / ordering
// ---------------------------------------------------------------------------

var courseNames = map[string]string{
	"cross_course":    "Figure-8 Circuit",
	"bank_course":     "Yoshi Falls",
	"beach_course":    "Cheep Cheep Beach",
	"mansion_course":  "Luigi's Mansion",
	"desert_course":   "Desert Hills",
	"town_course":     "Delfino Square",
	"pinball_course":  "Waluigi Pinball",
	"ridge_course":    "Shroom Ridge",
	"snow_course":     "DK Pass",
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

// isBillboard reports whether a model is a camera-facing sprite.
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
// levels stage
// ---------------------------------------------------------------------------

// buildLevels writes one scene3d level per course scene.
func buildLevels(ctx *cli.Context, items []item, refs map[string]string) error {
	b := ctx.Builder

	scene := map[string]item{}
	sky := map[string]item{}
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
		}
	}
	sort.Slice(stems, func(i, j int) bool {
		if ci, cj := courseClass(stems[i]), courseClass(stems[j]); ci != cj {
			return ci < cj
		}
		return displayName(stems[i]) < displayName(stems[j])
	})

	n := 0
	for _, stem := range stems {
		sc := scene[stem]
		courseAsset, ok := refs[sc.File]
		if !ok {
			continue // objects stage disabled
		}
		var nkm *mkds.NKM
		if archive := courseArchive[stem]; archive != "" {
			nkm, _ = mkds.LoadNKM(archive)
		}

		doc := &schema.Level{
			Type:  schema.LevelScene3D,
			Scene: &schema.Scene{},
			// The course model at the origin: its BTA0 texture animations
			// (water, lava, conveyors) play through the object pipeline.
			Placements: []schema.Placement{{
				ID: 0, Object: courseAsset, Pos: []float64{0, 0, 0}, Name: displayName(stem),
			}},
		}
		if sk, ok := sky[stem]; ok {
			doc.Scene.Layers = append(doc.Scene.Layers, schema.Layer{
				ID: "sky", Name: "Skybox", File: sk.File,
				Mode: "toggle", Attach: "camera", Role: "sky", RenderOrder: -1,
			})
		}

		// The CPU drive line, baked as a toggleable line layer.
		var lo, hi [3]float64
		haveBounds := false
		if nkm != nil {
			if pts, loop := nkm.EnemyLoop(); len(pts) >= 2 {
				positions := make([][3]float32, len(pts))
				for i, p := range pts {
					v := [3]float64{p.X / worldPerGLBUnit, p.Y / worldPerGLBUnit, p.Z / worldPerGLBUnit}
					positions[i] = [3]float32{float32(v[0]), float32(v[1]), float32(v[2])}
					if !haveBounds {
						lo, hi, haveBounds = v, v, true
						continue
					}
					for k := 0; k < 3; k++ {
						lo[k] = math.Min(lo[k], v[k])
						hi[k] = math.Max(hi[k], v[k])
					}
				}
				var edges [][2]uint32
				for i := 1; i < len(positions); i++ {
					edges = append(edges, [2]uint32{uint32(i - 1), uint32(i)})
				}
				if loop {
					edges = append(edges, [2]uint32{uint32(len(positions) - 1), 0})
				}
				lineFile := stem + "-driveline.glb"
				p, err := b.Path("levels", lineFile)
				if err != nil {
					return err
				}
				if err := glb.WriteLines(p, positions, edges, [3]float32{0.4, 0.85, 1.0}); err != nil {
					return err
				}
				doc.Scene.Layers = append(doc.Scene.Layers, schema.Layer{
					ID: "driveline", Name: "CPU drive line", File: lineFile, Mode: "toggle",
				})
			}
		}

		// Establishing shot over the drive line's extent (the course's raced
		// footprint); fall back to a generic high shot.
		cam := &schema.Camera{Mode: "fly", FOV: 50, Near: 0.1, Far: 4000, Fly: &schema.Fly{Speed: 30}}
		if haveBounds {
			cx, cz := (lo[0]+hi[0])/2, (lo[2]+hi[2])/2
			span := math.Max(hi[0]-lo[0], hi[2]-lo[2])
			cam.Pos = []float64{cx, hi[1] + span*0.5, lo[2] - span*0.6}
			cam.Target = []float64{cx, (lo[1] + hi[1]) / 2, cz}
		} else {
			cam.Pos = []float64{0, 60, -120}
			cam.Target = []float64{0, 0, 0}
		}
		doc.Camera = cam

		// OBJI map-object placements + their NKM routes.
		if nkm != nil {
			addPlacements(doc, nkm, stem, items, refs)
		}

		b.AddLevel(schema.Asset{
			ID: slugify(stem), Name: displayName(stem), Group: courseGroup(stem),
		}, doc)
		n++
		ctx.Progress("levels", n, len(stems), fmt.Sprintf("%-24s %d placements", stem, len(doc.Placements)))
	}
	return nil
}

func courseGroup(stem string) string {
	switch courseClass(stem) {
	case 0:
		return "Nitro courses"
	case 1:
		return "Retro courses"
	default:
		return "Other scenes"
	}
}

// addPlacements appends the course's OBJI map objects (and the routes movers
// follow) to the level document.
func addPlacements(doc *schema.Level, nkm *mkds.NKM, stem string, items []item, refs map[string]string) {
	byModel := map[string]item{}
	for _, it := range items {
		if it.Course == stem && !it.Scene && !it.Skybox {
			byModel[strings.ToLower(it.Model)] = it
		}
	}
	// The itembox spawns 12.0 world units above its OBJI position (its init
	// adds the fx32 constant 0xC000 to Y); other objects place as-authored.
	const itemboxHover = 12.0 / worldPerGLBUnit
	routeIdx := map[int]string{}
	skipped := map[string]int{}
	pid := 1
	for _, o := range nkm.Objects {
		var file string
		if o.ID == 0x65 {
			file = itemboxFile
		} else if name, ok := bindings[o.ID]; ok {
			if it, ok := byModel[name]; ok {
				file = it.File
			}
		}
		asset, ok := refs[file]
		if file == "" || !ok {
			skipped[fmt.Sprintf("0x%03X", o.ID)]++
			continue
		}
		pl := schema.Placement{
			ID:     pid,
			Object: asset,
			Pos:    []float64{o.Pos.X / worldPerGLBUnit, o.Pos.Y / worldPerGLBUnit, o.Pos.Z / worldPerGLBUnit},
			Props:  map[string]any{"obji": fmt.Sprintf("0x%03X", o.ID)},
		}
		if o.ID == 0x65 {
			pl.Pos[1] += itemboxHover
		}
		if o.Rot.X != 0 || o.Rot.Y != 0 || o.Rot.Z != 0 {
			pl.Rot = []float64{rad(o.Rot.X), rad(o.Rot.Y), rad(o.Rot.Z)}
		}
		if o.Scale.X != 1 || o.Scale.Y != 1 || o.Scale.Z != 1 {
			if o.Scale.X == o.Scale.Y && o.Scale.Y == o.Scale.Z {
				pl.Scale = schema.Scale{o.Scale.X}
			} else {
				pl.Scale = schema.Scale{o.Scale.X, o.Scale.Y, o.Scale.Z}
			}
		}
		if o.RouteID >= 0 && o.RouteID < len(nkm.Paths) && len(nkm.Paths[o.RouteID].Points) >= 2 {
			rid, ok := routeIdx[o.RouteID]
			if !ok {
				pth := nkm.Paths[o.RouteID]
				r := schema.Route{ID: fmt.Sprintf("r%d", o.RouteID), Loop: pth.Loop}
				for _, pt := range pth.Points {
					r.Points = append(r.Points, []float64{pt.X / worldPerGLBUnit, pt.Y / worldPerGLBUnit, pt.Z / worldPerGLBUnit})
				}
				doc.Routes = append(doc.Routes, r)
				rid = r.ID
				routeIdx[o.RouteID] = rid
			}
			// NKM point speeds are not decoded; 8 GLB units/s reads naturally.
			pl.Route = &schema.RouteRef{ID: rid, Speed: 8, Face: true}
		}
		doc.Placements = append(doc.Placements, pl)
		pid++
	}
}

func rad(d float64) float64 { return d * math.Pi / 180 }
