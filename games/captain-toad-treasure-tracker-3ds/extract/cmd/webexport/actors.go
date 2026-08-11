package main

// actors.go exports the stage's *actors* — Captain Toad, Toadette and the goal
// star — as skinned, animated models, and places them where the stage's own map
// puts them.
//
// Nothing here is named by convention. The map's GoalList entry is a logic
// object, `KinopioBrigadeChecker`, which carries no archive at all; what it
// *links* is a `GoalItem` (the star) and a `PlayerNpc` whose ModelName is
// `KinopicoNpc` (Toadette), each with its own translation. Toad comes from
// PlayerList, and the clip he plays comes from the DemoObjList entry beside
// him, which names it outright: `StartAnimation: CourseInRove`.
//
// Their lighting is not the terrain's. The oracle's per-draw dump shows the
// fragment-lighting unit programmed differently for the character draws — a
// brighter ambient, and a second light that is Toad's own headlamp (his
// InitLightActor.byml declares a `spotLight` on the HeadLight joint, aimed at a
// list of check points). The key light *is* the stage's mainLight: the unit is
// handed a view-space direction of (0.735, 0.610, 0.294), and rotating that
// back through the game's own view matrix gives (0.431, 0.666, 0.609), which is
// the stage's mainLight travel vector negated, to four decimal places.

import (
	"fmt"
	"image"
	"strings"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/n3ds"
	"retroreverse.com/tools/platform/n3ds/bchglb"
)

// actor is one exported character or prop.
type actor struct {
	id      string // Retro-X object id
	name    string
	object  string // the /ObjectData archive holding its model
	model   string // the model inside it
	animArc string // an extra archive to take clips from ("" = the object's own)
	clips   []string

	// visible names the meshes to export by material name. A character carries
	// every face, hand and eye it can ever wear — nine faces, seven hands a
	// side — and the game shows one of each, chosen by the visibility
	// animations beside the skeletal ones. Those name their meshes in terms
	// this decoder cannot yet bind to a model's meshes, so the set here is the
	// one the *running game* draws in this scene, read off the oracle's draw
	// census by matching each draw's vertex-buffer offset to a decoded mesh.
	// Empty exports every mesh.
	visible []string
}

var actors = []actor{
	{
		id: "kinopio", name: "Captain Toad", object: "Kinopio", model: "Kinopio",
		animArc: "KinopioAnimationSeason1OpeningStage",
		clips:   []string{"CourseInRove", "Wait", "Walk", "WaitRove", "GoalPose"},
		visible: []string{
			"KinopioRuckMat00", "KinopioMetalMat00", "LampStrap", "KinopioLightMat00",
			"KinopioBodyMat00", "Kinopio_Skin", "Kinopio_Scarf", "KinopioShoesMat00",
			"KinopioBuckle00", "KinopioHeadMat00",
			"Kinopio_HandR02", "Kinopio_HandL02", "Kinopio_Face00", "Kinopio_EyeBall",
		},
	},
	{
		id: "kinopico", name: "Toadette", object: "KinopicoNpc", model: "KinopicoNpc",
		animArc: "KinopioAnimationSeason1OpeningStage",
		clips:   []string{"KinopicoWaitRove", "KinopicoDemoOpeningSeason1", "KinopicoGoalPose"},
		visible: []string{
			"KinopicoMetalMat00", "KinopicoRuckMat00", "KinopicoBodyMat00", "KinopicoHeadMat00",
			"KinopicoSkinMat00", "KinopicoBodyMat01", "KinopicoShoesMat00", "KinopicoBuckleMat00",
			"Kinopio_HandR02", "Kinopio_HandL02", "Kinopio_Face00", "Kinopio_EyeBall",
		},
	},
	{
		id: "goalitem", name: "Goal Star", object: "GoalItem", model: "GoalItem",
		// The map's GoalList calls for AppearAnimName "Appear", which is not a
		// skeletal clip — the star's archive holds no skeletal animation by that
		// name. These are its skeletal ones for this stage. Its "Wait" and "Got"
		// are left out: they use two of the packed key encodings this decoder
		// still refuses rather than guesses at.
		clips: []string{"DemoOpeningSeason1", "GoalPose", "Land"},
	},
}

// characterLights is the rig the game's fragment-lighting unit is programmed
// with for the character draws, read from `bootoracle -gputrace`. The key
// direction is the stage's own mainLight (see the file comment); the ambient is
// the character set's, which is brighter than the terrain's.
//
// The headlamp is deliberately not baked. It is a spot light the game aims
// every frame from Toad's HeadLight joint, and a baked cone would be a lie
// about a light that moves with his head.
var characterLights = []schema.Light{
	{ID: "character-ambient", Type: "ambient", Keys: []schema.LightKey{{Color: "#949ca6"}}},
	{ID: "mainLight", Type: "directional", Keys: []schema.LightKey{{
		Color: "#ffbc4c",
		Dir:   []float64{0.4307, 0.6657, 0.6087},
	}}},
}

// exportActors writes one GLB per actor and returns the objects and the
// placements the stage's map asks for.
func exportActors(fs *n3ds.RomFS, stage string, path func(rel ...string) (string, error)) ([]objectAsset, []schema.Placement, error) {
	places, err := readActorPlacements(fs, stage)
	if err != nil {
		return nil, nil, err
	}
	cache := map[string]*objectArchive{}
	var objs []objectAsset
	for _, a := range actors {
		out, err := path("objects", a.id+".glb")
		if err != nil {
			return nil, nil, err
		}
		tris, clips, skinned, err := exportActor(fs, a, out, cache)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", a.id, err)
		}
		objs = append(objs, objectAsset{
			actor: a, file: a.id + ".glb", tris: tris, clips: clips, skinned: skinned,
		})
	}
	return objs, places, nil
}

// objectAsset is one written actor, ready to be registered with the builder.
type objectAsset struct {
	actor actor
	file  string
	tris  int
	clips []schema.Animation

	// skinned records that the GLB's meshes are bound to a skeleton, which the
	// level document has to say out loud: the viewer places an object by
	// *cloning* its prototype, and a plain three.js clone of a skinned mesh
	// keeps pointing at the ORIGINAL skeleton's bones — bones that live under a
	// prototype nobody adds to the scene, so their world matrices stay the
	// identity. Every bone-space mesh then draws at the origin, which is a
	// character collapsed into a heap with its cap under its chin. `skinnedClone`
	// is what asks for SkeletonUtils.clone instead, and it also keeps the object
	// out of the instanced-batching path, which drops skinning altogether.
	skinned bool
}

// exportActor writes one actor's GLB: its skeleton, its skinned meshes with the
// material's own baked albedo, and its clips.
func exportActor(fs *n3ds.RomFS, a actor, out string, cache map[string]*objectArchive) (int, []schema.Animation, bool, error) {
	o, err := loadObject(fs, a.object, cache)
	if err != nil {
		return 0, nil, false, err
	}
	model := o.Model
	if model == nil || model.Name != a.model {
		for _, m := range o.Extra {
			if m.Name == a.model {
				model = m
			}
		}
	}
	if model == nil {
		return 0, nil, false, fmt.Errorf("no model named %q", a.model)
	}

	want := map[string]bool{}
	for _, n := range a.visible {
		want[n] = true
	}

	s := glb.NewScene()
	rig := bchglb.AddSkeleton(s, model, -1, "")
	tris, skinned := 0, false
	for i := range model.Meshes {
		sh := &model.Meshes[i]
		mat := &model.Materials[sh.MaterialIndex]
		if len(want) > 0 {
			if !want[mat.Name] {
				continue
			}
			delete(want, mat.Name)
		}
		pr, err := actorPrim(sh, mat, model, o.Textures)
		if err != nil {
			return 0, nil, false, err
		}
		tris += len(pr.Tris)
		bchglb.BindJoints(&pr, sh)
		node := s.AddNode(mat.Name, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
		if err := s.AddMesh(node, mat.Name, []glb.Prim{pr}); err != nil {
			return 0, nil, false, err
		}
		if sk := rig.Skin(s, sh); sk >= 0 {
			s.SetNodeSkin(node, sk)
			skinned = true
		}
	}
	for n := range want {
		return 0, nil, false, fmt.Errorf("the visible set names %q, which the model does not have", n)
	}

	// Clips come from the object's own archive unless it names another.
	sources := []*objectArchive{o}
	if a.animArc != "" {
		anims, err := loadObject(fs, a.animArc, cache)
		if err != nil {
			return 0, nil, false, err
		}
		sources = append(sources, anims)
	}
	var got []schema.Animation
	for _, want := range a.clips {
		found := false
		for _, src := range sources {
			for _, f := range src.BCHs {
				for _, e := range f.Groups[n3ds.BCHSkeletalAnims] {
					if e.Name != want || found {
						continue
					}
					an, err := f.DecodeSkeletalAnim(e)
					if err != nil {
						return 0, nil, false, fmt.Errorf("clip %q: %w", want, err)
					}
					if err := bchglb.AddClip(s, rig, an, want); err != nil {
						return 0, nil, false, err
					}
					found = true
					// Whether a clip repeats is the animation header's own
					// flag: an idle loops, a goal pose plays once and holds.
					loop := "hold"
					if an.Loop {
						loop = "loop"
					}
					got = append(got, schema.Animation{ID: want, Name: want, Loop: loop})
				}
			}
		}
		if !found {
			return 0, nil, false, fmt.Errorf("no skeletal animation named %q", want)
		}
	}
	return tris, got, skinned, s.Write(out, a.model)
}

// actorPrim builds one mesh's primitive: the material's baked albedo, the
// stage's character lighting folded into the vertex colours, and the material's
// own alpha and blend behaviour.
func actorPrim(sh *n3ds.BCHMesh, mat *n3ds.BCHMaterial, model *n3ds.BCHModel, textures map[string]*image.NRGBA) (glb.Prim, error) {
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
	img, err := mat.BakeAlbedo(textures)
	if err != nil {
		return pr, err
	}
	pr.Colors = make([][4]uint8, len(sh.Verts))
	bakeActorShading(sh, mat, img, pr.Colors)

	if img != nil {
		pr.Image = img
		pr.UVs = albedoUVs(sh, mat, textures)
		pr.WrapS, pr.WrapT = materialWrap(mat, textures)
	}
	switch {
	case mat.Additive():
		// A glow or a sparkle adds its colour to the frame. Exported as
		// alpha-blended it would grey the star out instead of lighting it.
		pr.Additive = true
	case mat.Blends:
		pr.Blend = true
	case mat.AlphaTest:
		c, ok := mat.AlphaCutoff()
		if !ok {
			return pr, fmt.Errorf("material %q alpha-tests with comparison %d, which is not a cutoff", mat.Name, mat.AlphaFunc)
		}
		pr.AlphaCutoff = c
	default:
		pr.Opaque = true
	}
	pr.Tris = make([][3]uint32, len(sh.Indices)/3)
	for t := range pr.Tris {
		pr.Tris[t] = [3]uint32{sh.Indices[t*3], sh.Indices[t*3+1], sh.Indices[t*3+2]}
	}
	return pr, nil
}

// bakeActorShading folds the character light rig into the vertex colours, by
// running the material's own combiner rather than by multiplying.
//
// The factor is a ratio of the chain to itself — lit over neutral — which is
// what makes it multiply cleanly with the albedo BakeAlbedo already wrote (see
// BCHMaterial.Shade). The ratio is measured at the brightest texel of that
// albedo, because a black one carries no information about the shading; the
// material's BakeCheck says how much that choice matters, and for all but one
// of Toad's forty it is under two per cent.
//
// The result goes out as COLOR_0, which glTF defines as linear, so the
// gamma-space product is converted on the way — the same reason the terrain's
// bake does it (see gamma-space-multiply).
func bakeActorShading(sh *n3ds.BCHMesh, mat *n3ds.BCHMaterial, albedo *image.NRGBA, out [][4]uint8) {
	texel := brightestTexel(albedo)

	var ambient [3]float64
	type dirLight struct{ color, l [3]float64 }
	var dirs []dirLight
	for _, l := range characterLights {
		k := l.Keys[0]
		c := parseHexColor(k.Color)
		if l.Type == "ambient" {
			for i := 0; i < 3; i++ {
				ambient[i] += c[i]
			}
			continue
		}
		d := dirLight{color: c}
		copy(d.l[:], k.Dir)
		dirs = append(dirs, d)
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
		f := mat.Shade(sh.Verts[v].Color, lit, texel)
		for i := 0; i < 3; i++ {
			out[v][i] = uint8(srgbToLinear(clamp01(f[i]))*255 + 0.5)
		}
		out[v][3] = 255
	}
}

// brightestTexel is the albedo texel the shading ratio is measured at: the one
// carrying the most signal. A material with no texture bakes to a single texel
// and that is the one.
func brightestTexel(img *image.NRGBA) [4]uint8 {
	best := [4]uint8{255, 255, 255, 255}
	if img == nil {
		return best
	}
	bright := -1
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.NRGBAAt(x, y)
			if l := int(c.R) + int(c.G) + int(c.B); l > bright {
				bright, best = l, [4]uint8{c.R, c.G, c.B, c.A}
			}
		}
	}
	return best
}

// readActorPlacements resolves where the map puts the actors.
//
// PlayerList gives Toad his spot and DemoObjList, at the same spot, gives the
// facing and the clip the stage starts him in. The star and Toadette are both
// *links* of the GoalList's checker object, and each carries its own
// translation — so neither is placed by guessing an offset from the checker.
func readActorPlacements(fs *n3ds.RomFS, stage string) ([]schema.Placement, error) {
	raw, err := fs.File("/StageData/" + stage + "Map1.szs")
	if err != nil {
		return nil, err
	}
	arc, err := n3ds.OpenSZS(raw)
	if err != nil {
		return nil, err
	}
	blob, ok := arc.File(stage + "Map.byml")
	if !ok {
		return nil, fmt.Errorf("%s: the map archive has no %sMap.byml", stage, stage)
	}
	doc, err := n3ds.ParseBYML(blob)
	if err != nil {
		return nil, err
	}
	root, _ := doc.(n3ds.BYMLDict)

	var out []schema.Placement
	id := 1
	add := func(object string, d n3ds.BYMLDict, anim, name string) {
		t, r := vec3(d["Translate"]), vec3(d["Rotate"])
		out = append(out, schema.Placement{
			ID: id, Object: object, Name: name,
			Pos:  []float64{float64(t[0]), float64(t[1]), float64(t[2])},
			Rot:  []float64{float64(r[0]), float64(r[1]), float64(r[2])},
			Anim: anim,
		})
		id++
	}

	// Toad, and the demo entry beside him that names his opening animation.
	anim, facing := "", n3ds.BYMLDict(nil)
	for _, it := range list(root["DemoObjList"]) {
		d, _ := it.(n3ds.BYMLDict)
		if s, _ := d["StartAnimation"].(string); s != "" {
			anim, facing = s, d
		}
	}
	for _, it := range list(root["PlayerList"]) {
		d, _ := it.(n3ds.BYMLDict)
		if facing != nil {
			// The demo turns him to face the camera as the stage opens; his
			// PlayerList rotation is where he would stand without it.
			d = merge(d, facing)
		}
		add("kinopio", d, anim, "Captain Toad")
	}

	// The goal checker's links: the star and Toadette.
	for _, it := range list(root["GoalList"]) {
		d, _ := it.(n3ds.BYMLDict)
		links, _ := d["Links"].(n3ds.BYMLDict)
		for _, l := range list(links["GoalItem"]) {
			add("goalitem", l.(n3ds.BYMLDict), "DemoOpeningSeason1", "Goal Star")
		}
		for _, l := range list(links["PlayerNpc"]) {
			ld := l.(n3ds.BYMLDict)
			if m, _ := ld["ModelName"].(string); m != "KinopicoNpc" {
				return nil, fmt.Errorf("the goal's PlayerNpc names model %q, not KinopicoNpc", m)
			}
			add("kinopico", ld, "KinopicoWaitRove", "Toadette")
		}
	}
	return out, nil
}

func list(v any) []any { s, _ := v.([]any); return s }

// merge takes the rotation from b and everything else from a.
func merge(a, b n3ds.BYMLDict) n3ds.BYMLDict {
	out := n3ds.BYMLDict{}
	for k, v := range a {
		out[k] = v
	}
	out["Rotate"] = b["Rotate"]
	return out
}

func init() {
	// The visible sets above are written in material-name terms, and a typo in
	// one would silently export a character with a hole in it; exportActor
	// checks every name it was given is used, so this is only about the shape.
	for _, a := range actors {
		for _, n := range a.visible {
			if strings.TrimSpace(n) == "" {
				panic("actors: empty visible mesh name")
			}
		}
	}
}
