package main

// stage.go rebuilds one of the game's stages from the cartridge: it reads the
// stage's placement map, loads each placed object's model archive, and writes
// the whole scene as one GLB with every object at the position, rotation and
// scale the map gives it.
//
// The chain per stage is
//
//	/StageData/<Stage>Map1.szs   → SARC → <Stage>Map.byml   the placements
//	/ObjectData/<UnitConfigName>.szs → SARC → the object's .bch (its models)
//	                                        + InitModel.byml (its texture archive)
//	/ObjectData/<TextureArc>.szs → SARC → a .bch holding only textures
//
// Nothing here is guessed from a filename: the map names the object, and the
// object's own InitModel.byml names the archive its textures come from.

import (
	"fmt"
	"image"
	"math"
	"sort"
	"strings"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/n3ds"
)

// placement is one object the map places.
type placement struct {
	ID        string
	Unit      string // UnitConfigName: the object archive to load
	Model     string // ModelName, when the map overrides the archive's own model
	Translate [3]float32
	Rotate    [3]float32 // degrees, XYZ
	Scale     [3]float32
	List      string // which of the map's lists it came from
}

// mapLists are the placement lists this pass reads. ObjectList is the stage's
// terrain. The others are deliberately left out: PlayerList, GoalList and
// DemoObjList place characters and cameras (the object pass, not the level
// pass), and SkyList's dome is not yet exportable — its material binds only a
// cloud *mask* to unit 0, and the blue behind the clouds comes from somewhere
// this decoder does not yet read, so exporting it would ship a black sky.
var mapLists = []string{"ObjectList"}

// readStageMap decodes a stage's placement map.
func readStageMap(fs *n3ds.RomFS, stage string) ([]placement, error) {
	raw, err := fs.File("/StageData/" + stage + "Map1.szs")
	if err != nil {
		return nil, err
	}
	arc, err := n3ds.OpenSZS(raw)
	if err != nil {
		return nil, fmt.Errorf("%s map archive: %w", stage, err)
	}
	blob, ok := arc.File(stage + "Map.byml")
	if !ok {
		return nil, fmt.Errorf("%s: the map archive has no %sMap.byml", stage, stage)
	}
	doc, err := n3ds.ParseBYML(blob)
	if err != nil {
		return nil, fmt.Errorf("%s map: %w", stage, err)
	}
	root, ok := doc.(n3ds.BYMLDict)
	if !ok {
		return nil, fmt.Errorf("%s map: root is %T, want a dictionary", stage, doc)
	}

	var out []placement
	for _, list := range mapLists {
		items, _ := root[list].([]any)
		for _, it := range items {
			d, ok := it.(n3ds.BYMLDict)
			if !ok {
				continue
			}
			p := placement{List: list, Scale: [3]float32{1, 1, 1}}
			p.ID, _ = d["Id"].(string)
			p.Unit, _ = d["UnitConfigName"].(string)
			p.Model, _ = d["ModelName"].(string)
			p.Translate = vec3(d["Translate"])
			p.Rotate = vec3(d["Rotate"])
			if s, ok := d["Scale"]; ok && s != nil {
				p.Scale = vec3(s)
			}
			out = append(out, p)
		}
	}
	return out, nil
}

func vec3(v any) [3]float32 {
	d, ok := v.(n3ds.BYMLDict)
	if !ok {
		return [3]float32{}
	}
	f := func(k string) float32 {
		switch n := d[k].(type) {
		case float32:
			return n
		case int32:
			return float32(n)
		}
		return 0
	}
	return [3]float32{f("X"), f("Y"), f("Z")}
}

// objectArchive is one loaded /ObjectData archive: its models and the textures
// they resolve against.
type objectArchive struct {
	Name     string
	Model    *n3ds.BCHModel
	Extra    []*n3ds.BCHModel // the object's other models (its depth-shadow proxy)
	Textures map[string]*image.NRGBA
}

// loadObject loads an object archive and, if its InitModel.byml names one, the
// texture archive beside it.
func loadObject(fs *n3ds.RomFS, name string, cache map[string]*objectArchive) (*objectArchive, error) {
	if o, ok := cache[name]; ok {
		return o, nil
	}
	raw, err := fs.File("/ObjectData/" + name + ".szs")
	if err != nil {
		return nil, err
	}
	arc, err := n3ds.OpenSZS(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	o := &objectArchive{Name: name, Textures: map[string]*image.NRGBA{}}

	// The object's textures may live in a separate archive, which its own
	// InitModel.byml names. Load that first so the models can bind against it.
	if blob, ok := arc.File("InitModel.byml"); ok {
		doc, err := n3ds.ParseBYML(blob)
		if err != nil {
			return nil, fmt.Errorf("%s InitModel.byml: %w", name, err)
		}
		if d, ok := doc.(n3ds.BYMLDict); ok {
			if ta, ok := d["TextureArc"].(string); ok && ta != "" {
				tex, err := loadObject(fs, ta, cache)
				if err != nil {
					return nil, fmt.Errorf("%s texture archive %q: %w", name, ta, err)
				}
				for k, v := range tex.Textures {
					o.Textures[k] = v
				}
			}
		}
	}

	for _, f := range arc.Files {
		if !strings.HasSuffix(f.Name, ".bch") {
			continue
		}
		bch, err := n3ds.ParseBCH(f.Data)
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", name, f.Name, err)
		}
		for _, e := range bch.Groups[n3ds.BCHTextures] {
			t, err := bch.DecodeTexture(e)
			if err != nil {
				return nil, fmt.Errorf("%s/%s: %w", name, f.Name, err)
			}
			o.Textures[t.Name] = t.Image
		}
		for _, e := range bch.Groups[n3ds.BCHModels] {
			m, err := bch.DecodeModel(e)
			if err != nil {
				return nil, fmt.Errorf("%s/%s: %w", name, f.Name, err)
			}
			// An object archive holds the object's model and, beside it, the
			// proxy it draws into the depth-shadow and Z-prepass passes its
			// InitExecutor lists. The visible one is the model named after the
			// object; the proxy carries a suffix.
			if m.Name == name {
				o.Model = m
			} else {
				o.Extra = append(o.Extra, m)
			}
		}
	}
	if o.Model == nil && len(o.Extra) > 0 {
		return nil, fmt.Errorf("%s: no model named after the object (found %s)", name, o.Extra[0].Name)
	}
	cache[name] = o
	return o, nil
}

// addObject appends one placed object's meshes to the scene, under a node
// carrying the map's transform.
func addObject(s *glb.Scene, p placement, o *objectArchive, lights []schema.Light) (tris int, err error) {
	if o.Model == nil {
		return 0, nil
	}
	prims := make([]glb.Prim, 0, len(o.Model.Meshes))
	for _, sh := range o.Model.Meshes {
		sh := sh
		// The material's combiner is lighting × albedo × vertex colour, and the
		// lighting is baked into that vertex colour (bakeLighting), so the glTF
		// material stays unlit and does the one multiply that is left.
		pr := glb.Prim{BaseColor: [4]float32{1, 1, 1, 1}, Unlit: true, DoubleSided: true}
		pr.Positions = make([][3]float32, len(sh.Verts))
		for j, v := range sh.Verts {
			pr.Positions[j] = v.Pos
		}
		if sh.HasNormal {
			pr.Normals = make([][3]float32, len(sh.Verts))
			for j, v := range sh.Verts {
				pr.Normals[j] = v.Normal
			}
		}
		if sh.HasColor || len(lights) > 0 {
			pr.Colors = make([][4]uint8, len(sh.Verts))
			if len(lights) > 0 {
				bakeLighting(&sh, lights, pr.Colors)
			} else {
				for j, v := range sh.Verts {
					pr.Colors[j] = v.Color
				}
			}
		}
		if sh.UVCount > 0 && sh.MaterialIndex < len(o.Model.Materials) {
			mat := &o.Model.Materials[sh.MaterialIndex]
			if img := o.Textures[mat.Texture()]; img != nil {
				pr.Image = img
				pr.UVs = make([][2]float32, len(sh.Verts))
				for j, v := range sh.Verts {
					pr.UVs[j] = [2]float32{v.UV[0][0], 1 - v.UV[0][1]}
				}
				pr.WrapS, pr.WrapT = gltfRepeat, gltfRepeat
				if err := setAlpha(&pr, mat); err != nil {
					return 0, err
				}
			}
		}
		pr.Tris = make([][3]uint32, len(sh.Indices)/3)
		for t := range pr.Tris {
			pr.Tris[t] = [3]uint32{sh.Indices[t*3], sh.Indices[t*3+1], sh.Indices[t*3+2]}
		}
		tris += len(pr.Tris)
		prims = append(prims, pr)
	}
	node := s.AddNode(p.ID+"-"+p.Unit, -1, p.Translate, eulerXYZ(p.Rotate), p.Scale)
	return tris, s.AddMesh(node, p.Unit, prims)
}

const gltfRepeat = 10497

// setAlpha carries the material's own alpha handling into the glTF material.
//
// The stage's albedo textures are ETC1A4 and so every one of them has an alpha
// plane, but most of the materials sampling them neither blend nor alpha-test —
// for those the plane is not coverage and using it punches holes in the stone
// and drops the floor out from between the clover leaves. Only a material that
// says it blends, or that enables the alpha test, gets alpha at all.
func setAlpha(pr *glb.Prim, mat *n3ds.BCHMaterial) error {
	switch {
	case mat.Blends:
		pr.Blend = true
	case mat.AlphaTest:
		c, ok := mat.AlphaCutoff()
		if !ok {
			return fmt.Errorf("material %q alpha-tests with comparison %d, which is not a cutoff",
				mat.Name, mat.AlphaFunc)
		}
		pr.AlphaCutoff = c
	default:
		pr.Opaque = true
	}
	return nil
}

// eulerXYZ converts the map's degree triple to a quaternion. The map's rotation
// is applied X then Y then Z, which is the order the engine's own pose data
// (InitPose.byml) is written in.
func eulerXYZ(deg [3]float32) [4]float32 {
	const d2r = math.Pi / 180
	cx, sx := math.Cos(float64(deg[0])*d2r/2), math.Sin(float64(deg[0])*d2r/2)
	cy, sy := math.Cos(float64(deg[1])*d2r/2), math.Sin(float64(deg[1])*d2r/2)
	cz, sz := math.Cos(float64(deg[2])*d2r/2), math.Sin(float64(deg[2])*d2r/2)
	return [4]float32{
		float32(sx*cy*cz + cx*sy*sz),
		float32(cx*sy*cz - sx*cy*sz),
		float32(cx*cy*sz + sx*sy*cz),
		float32(cx*cy*cz - sx*sy*sz),
	}
}

// exportStage writes one stage's placed geometry as a GLB and reports which
// placed objects had no model archive to load (the characters, effects and
// markers, which are a later pass).
func exportStage(fs *n3ds.RomFS, stage, out string) (tris int, lights []schema.Light, skipped []string, err error) {
	places, err := readStageMap(fs, stage)
	if err != nil {
		return 0, nil, nil, err
	}
	if lights, err = readStageLights(fs, stage); err != nil {
		return 0, nil, nil, err
	}
	cache := map[string]*objectArchive{}
	s := glb.NewScene()
	miss := map[string]bool{}
	for _, p := range places {
		unit := p.Unit
		if p.Model != "" {
			unit = p.Model // the map may name a model archive of its own
		}
		o, err := loadObject(fs, unit, cache)
		if err != nil {
			miss[unit] = true
			continue
		}
		n, err := addObject(s, p, o, lights)
		if err != nil {
			return 0, nil, nil, err
		}
		tris += n
	}
	for k := range miss {
		skipped = append(skipped, k)
	}
	sort.Strings(skipped)
	return tris, lights, skipped, s.Write(out, stage)
}

// --- lighting -----------------------------------------------------------

// A stage's lights live in a scene archive beside it, named by stripping the
// trailing "Stage" from the stage's name and appending "Scene" — a convention
// that holds for all 118 of the cartridge's scenes. Which of that scene's
// lights apply is not a convention: the stage's Design archive carries a
// LightAreas.byml naming them, per light area, and the map's AreaList places
// those areas.
const (
	stageSuffix = "Stage"
	sceneSuffix = "Scene"
)

// readStageLights resolves the lights a stage's terrain is lit by.
//
// The opening stage declares two light areas that differ only in whether the
// key light casts a depth shadow — the same colour, the same direction — so
// which one is read does not change the rig. When a stage's areas disagree
// about more than that, this says so rather than picking one silently.
func readStageLights(fs *n3ds.RomFS, stage string) ([]schema.Light, error) {
	design, err := fs.File("/StageData/" + stage + "Design1.szs")
	if err != nil {
		return nil, err
	}
	darc, err := n3ds.OpenSZS(design)
	if err != nil {
		return nil, fmt.Errorf("%s design archive: %w", stage, err)
	}
	blob, ok := darc.File("LightAreas.byml")
	if !ok {
		return nil, nil // a stage with no light areas is lit by the viewer's default
	}
	doc, err := n3ds.ParseBYML(blob)
	if err != nil {
		return nil, fmt.Errorf("%s LightAreas.byml: %w", stage, err)
	}
	areas, _ := doc.(n3ds.BYMLDict)["LightAreas"].([]any)
	if len(areas) == 0 {
		return nil, nil
	}

	if !strings.HasSuffix(stage, stageSuffix) {
		return nil, fmt.Errorf("stage %q does not end in %q, so its scene archive cannot be named", stage, stageSuffix)
	}
	sceneName := strings.TrimSuffix(stage, stageSuffix) + sceneSuffix
	sraw, err := fs.File("/ObjectData/" + sceneName + ".szs")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", sceneName, err)
	}
	sarc, err := n3ds.OpenSZS(sraw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", sceneName, err)
	}
	sblob, ok := sarc.File(sceneName + ".bch")
	if !ok {
		return nil, fmt.Errorf("%s has no scene .bch", sceneName)
	}
	sf, err := n3ds.ParseBCH(sblob)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", sceneName, err)
	}
	lights, err := sf.Lights()
	if err != nil {
		return nil, err
	}

	// Take the first area's rig, and check the others agree on the light
	// *values* — they may name a different light, but a stage whose areas light
	// the terrain differently needs a decision this pass is not making.
	var out []schema.Light
	var first []*n3ds.BCHLight
	for ai, a := range areas {
		d, _ := a.(n3ds.BYMLDict)
		names, _ := d["Lights"].([]any)
		var rig []*n3ds.BCHLight
		for _, n := range names {
			nd, _ := n.(n3ds.BYMLDict)
			name, _ := nd["lightName"].(string)
			l := lights[name]
			if l == nil {
				return nil, fmt.Errorf("%s light area names %q, which %s does not hold", stage, name, sceneName)
			}
			rig = append(rig, l)
		}
		if ai == 0 {
			first = rig
			continue
		}
		if len(rig) != len(first) {
			return nil, fmt.Errorf("%s light areas disagree: %d lights vs %d", stage, len(rig), len(first))
		}
		for i := range rig {
			if rig[i].Ambient != first[i].Ambient || rig[i].Diffuse != first[i].Diffuse ||
				rig[i].Direction != first[i].Direction {
				return nil, fmt.Errorf("%s light areas disagree about light %d (%q vs %q)",
					stage, i, rig[i].Name, first[i].Name)
			}
		}
	}

	for _, l := range first {
		// Every light contributes its ambient term; a directional one adds its
		// diffuse over the surfaces facing it.
		if l.Ambient != ([3]uint8{}) {
			out = append(out, schema.Light{
				ID: l.Name + "-ambient", Type: "ambient",
				Keys: []schema.LightKey{{Color: hexColor(l.Ambient)}},
			})
		}
		if !l.Directional {
			continue
		}
		// The file stores the direction the light travels; a renderer wants the
		// direction *towards* it.
		out = append(out, schema.Light{
			ID: l.Name, Type: "directional",
			Keys: []schema.LightKey{{
				Color: hexColor(l.Diffuse),
				Dir:   []float64{-float64(l.Direction[0]), -float64(l.Direction[1]), -float64(l.Direction[2])},
			}},
		})
	}
	return out, nil
}

func hexColor(c [3]uint8) string { return fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2]) }

// bakeLighting folds the stage's light rig into the vertex colours.
//
// The materials' combiner is `lighting × albedo`, and a later stage multiplies
// by the vertex colour, which carries the artists' baked ambient occlusion. All
// three multiplies happen on the hardware in **gamma space**, on 8-bit values —
// and that is why the rig is baked rather than handed to the renderer as lights.
// A linear-space renderer computing `albedo × colour × N·L` and encoding the
// result to sRGB does not reproduce a gamma-space one: the colours survive the
// round trip but the N·L factor comes back as N·L^(1/2.2), which at a typical
// 0.67 is 0.83 — a visibly flatter, brighter picture than the console's.
//
// Baking sidesteps that. The lit term is evaluated per vertex, which for this
// geometry is not an approximation: the normal is a vertex attribute, so a hard
// edge already has split vertices and a face's shading is constant across it
// either way.
//
// The result goes out as COLOR_0, which glTF defines as **linear**, so the
// gamma-space product is converted on the way. The scene's lights are still
// published in the level document — they are decoded fact, and what a renderer
// that wanted to light this dynamically would need.
func bakeLighting(sh *n3ds.BCHMesh, lights []schema.Light, out [][4]uint8) {
	var ambient [3]float64
	type dirLight struct{ color, l [3]float64 }
	var dirs []dirLight
	for _, l := range lights {
		if len(l.Keys) == 0 {
			continue
		}
		k := l.Keys[0]
		c := parseHexColor(k.Color)
		switch l.Type {
		case "ambient":
			for i := 0; i < 3; i++ {
				ambient[i] += c[i]
			}
		case "directional":
			d := dirLight{color: c}
			var n float64
			for i := 0; i < 3 && i < len(k.Dir); i++ {
				d.l[i] = k.Dir[i]
				n += k.Dir[i] * k.Dir[i]
			}
			if n = math.Sqrt(n); n > 0 {
				for i := 0; i < 3; i++ {
					d.l[i] /= n
				}
			}
			dirs = append(dirs, d)
		}
	}

	for v := range sh.Verts {
		lit := ambient
		if sh.HasNormal {
			n := sh.Verts[v].Normal
			for _, d := range dirs {
				nl := float64(n[0])*d.l[0] + float64(n[1])*d.l[1] + float64(n[2])*d.l[2]
				if nl <= 0 {
					continue
				}
				for i := 0; i < 3; i++ {
					lit[i] += d.color[i] * nl
				}
			}
		}
		for i := 0; i < 3; i++ {
			c := clamp01(lit[i])
			if sh.HasColor {
				c *= float64(sh.Verts[v].Color[i]) / 255 // the artists' baked occlusion
			}
			out[v][i] = uint8(srgbToLinear(c)*255 + 0.5)
		}
		out[v][3] = 255
		if sh.HasColor {
			out[v][3] = sh.Verts[v].Color[3]
		}
	}
}

func parseHexColor(s string) [3]float64 {
	var r, g, b int
	fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	return [3]float64{float64(r) / 255, float64(g) / 255, float64(b) / 255}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// srgbToLinear converts a gamma-space multiplier into the linear value glTF's
// COLOR_0 carries, so the renderer's linear multiply and its sRGB encode put
// the gamma-space product back on screen unchanged.
func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}
