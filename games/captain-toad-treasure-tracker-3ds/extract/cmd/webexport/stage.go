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
func addObject(s *glb.Scene, p placement, o *objectArchive) (tris int, err error) {
	if o.Model == nil {
		return 0, nil
	}
	prims := make([]glb.Prim, 0, len(o.Model.Meshes))
	for _, sh := range o.Model.Meshes {
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
		if sh.HasColor {
			pr.Colors = make([][4]uint8, len(sh.Verts))
			for j, v := range sh.Verts {
				pr.Colors[j] = v.Color
			}
		}
		if sh.UVCount > 0 && sh.MaterialIndex < len(o.Model.Materials) {
			if img := o.Textures[o.Model.Materials[sh.MaterialIndex].Texture()]; img != nil {
				pr.Image = img
				pr.UVs = make([][2]float32, len(sh.Verts))
				for j, v := range sh.Verts {
					pr.UVs[j] = [2]float32{v.UV[0][0], 1 - v.UV[0][1]}
				}
				pr.WrapS, pr.WrapT = gltfRepeat, gltfRepeat
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
func exportStage(fs *n3ds.RomFS, stage, out string) (tris int, skipped []string, err error) {
	places, err := readStageMap(fs, stage)
	if err != nil {
		return 0, nil, err
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
		n, err := addObject(s, p, o)
		if err != nil {
			return 0, nil, err
		}
		tris += n
	}
	for k := range miss {
		skipped = append(skipped, k)
	}
	sort.Strings(skipped)
	return tris, skipped, s.Write(out, stage)
}
