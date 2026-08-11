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
	"math"
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

	// pose is the clip whose visibility animations choose which of the model's
	// variant meshes are exported — a character carries every face, hand and
	// eye it can wear, and one of each is drawn. Empty exports every mesh.
	pose string
}

var actors = []actor{
	{
		id: "kinopio", name: "Captain Toad", object: "Kinopio", model: "Kinopio",
		animArc: "KinopioAnimationSeason1OpeningStage",
		clips:   []string{"CourseInRove", "Wait", "Walk", "WaitRove", "GoalPose"},
		pose:    "CourseInRove", // what the stage's DemoObjList starts him in
	},
	{
		id: "kinopico", name: "Toadette", object: "KinopicoNpc", model: "KinopicoNpc",
		animArc: "KinopioAnimationSeason1OpeningStage",
		clips:   []string{"KinopicoWaitRove", "KinopicoDemoOpeningSeason1", "KinopicoGoalPose"},
		pose:    "KinopicoWaitRove",
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
func exportActors(fs *n3ds.RomFS, stage string, path func(rel ...string) (string, error)) ([]objectAsset, []schema.Placement, []shadowCaster, []placedMask, error) {
	places, err := readActorPlacements(fs, stage)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cache := map[string]*objectArchive{}
	var objs []objectAsset
	for _, a := range actors {
		out, err := path("objects", a.id+".glb")
		if err != nil {
			return nil, nil, nil, nil, err
		}
		tris, clips, skinned, err := exportActor(fs, a, out, cache)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%s: %w", a.id, err)
		}
		objs = append(objs, objectAsset{
			actor: a, file: a.id + ".glb", tris: tris, clips: clips, skinned: skinned,
		})
	}
	casters, err := actorShadowCasters(fs, places, cache)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	masks, err := actorShadowMasks(fs, places, cache)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return objs, places, casters, masks, nil
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

	// Which of the model's variant meshes are drawn, from the clip's own
	// visibility animations rather than from a list written here.
	sources := []*objectArchive{o}
	if a.animArc != "" {
		anims, err := loadObject(fs, a.animArc, cache)
		if err != nil {
			return 0, nil, false, err
		}
		sources = append(sources, anims)
	}
	visible, err := poseVisibility(model, sources, a.pose)
	if err != nil {
		return 0, nil, false, fmt.Errorf("pose %q: %w", a.pose, err)
	}

	s := glb.NewScene()
	rig := bchglb.AddSkeleton(s, model, -1, "")
	tris, meshes, skinned := 0, 0, false
	for i := range model.Meshes {
		sh := &model.Meshes[i]
		mat := &model.Materials[sh.MaterialIndex]
		if !visible[sh.Node] {
			continue
		}
		pr, err := actorPrim(sh, mat, model, o.Textures)
		if err != nil {
			return 0, nil, false, err
		}
		tris += len(pr.Tris)
		meshes++
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
	// One mesh out per visible node, and no node drawn twice by accident: the
	// count is the cheap invariant that catches a filter which has stopped
	// filtering.
	nodes := 0
	for _, b := range visible {
		if b {
			nodes++
		}
	}
	if meshes < nodes {
		return 0, nil, false, fmt.Errorf("%d visible mesh nodes produced only %d meshes", nodes, meshes)
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

// poseVisibility resolves which of a model's mesh nodes are drawn, from the
// visibility animations belonging to one clip.
//
// A visibility animation belongs to the clip whose name is the LONGEST that
// prefixes its own. That matters: matching `Walk` by prefix also catches
// `WalkSlowlyFace`, `WalkSpeedyEye` and `WalkWaterHandL` — sixty-four
// animations instead of four, and whichever happened to be enumerated first
// decided the face. The archive's own skeletal clip names are what resolves it,
// so the rule reads them rather than assuming a suffix vocabulary.
//
// The frame is 0: the placement starts its clip there. A clip whose face
// changes part-way through — Captain Toad blinks — is not represented by a
// static mesh set, and glTF has no way to say it either: a skinned mesh ignores
// its node's transform, so the usual scale-to-zero trick cannot switch one off.
func poseVisibility(model *n3ds.BCHModel, sources []*objectArchive, pose string) ([]bool, error) {
	if pose == "" {
		on := make([]bool, len(model.MeshNodes))
		for i := range on {
			on[i] = true
		}
		return on, nil
	}

	// Every skeletal clip in these archives, so the longest-prefix rule has
	// something to be longest among.
	var clips []string
	for _, src := range sources {
		for _, f := range src.BCHs {
			for _, e := range f.Groups[n3ds.BCHSkeletalAnims] {
				clips = append(clips, e.Name)
			}
		}
	}
	owner := func(name string) string {
		best := ""
		for _, c := range clips {
			if strings.HasPrefix(name, c) && len(c) > len(best) {
				best = c
			}
		}
		return best
	}

	var anims []*n3ds.BCHVisibility
	mentioned := map[string]bool{}
	for _, src := range sources {
		for _, f := range src.BCHs {
			for _, e := range f.Groups[n3ds.BCHVisibilityAnims] {
				v, err := f.DecodeVisibilityAnim(e)
				if err != nil {
					return nil, err
				}
				for n := range v.Visible {
					mentioned[n] = true
				}
				if owner(e.Name) == pose {
					anims = append(anims, v)
				}
			}
		}
	}
	if len(anims) == 0 {
		return nil, fmt.Errorf("no visibility animation belongs to it")
	}

	// Every node any visibility animation in these archives switches has to be
	// decided by the ones this clip owns, or some variant is left to whatever
	// the enumeration order happens to be.
	decided := map[string]bool{}
	for _, v := range anims {
		for n := range v.Visible {
			decided[n] = true
		}
	}
	for n := range mentioned {
		if decided[n] {
			continue
		}
		for _, mn := range model.MeshNodes {
			if mn == n {
				return nil, fmt.Errorf("node %q is switched by some animation but not by this clip's", n)
			}
		}
	}

	// The count of visible nodes is constant over a clip — a character has one
	// face at a time — so a set that changes with the frame means the wrong
	// animations were gathered.
	on := n3ds.VisibleMeshNodes(model, anims, 0)
	want := 0
	for _, b := range on {
		if b {
			want++
		}
	}
	frames := 0
	for _, v := range anims {
		if int(v.Frames) > frames {
			frames = int(v.Frames)
		}
	}
	for fr := 1; fr <= frames; fr++ {
		got := 0
		for _, b := range n3ds.VisibleMeshNodes(model, anims, fr) {
			if b {
				got++
			}
		}
		if got != want {
			return nil, fmt.Errorf("shows %d nodes at frame 0 but %d at frame %d", want, got, fr)
		}
	}
	return on, nil
}

// shadowCaster is one actor's depth-shadow proxy, baked at the pose its
// placement starts in and ready to join the stage's caster layer.
type shadowCaster struct {
	name string
	prim glb.Prim
	pl   schema.Placement
}

// actorShadowCasters bakes each placed actor's own `_shd` proxy — the coarse
// model the game renders its depth shadow from — into world-space geometry.
//
// The proxies are SKINNED, unlike the terrain's, so they have to be posed
// rather than copied: a character's shadow is cast by the character standing
// where it stands, not by its bind pose with its arms out. The pose is frame 0
// of the clip the placement starts, which is the same frame its visible meshes
// are chosen at.
//
// The shadow pass this feeds is rendered once, so a moving actor's shadow does
// not follow it. That is the pass's design — the stage and its light are both
// static — and these actors stand still while they look around.
func actorShadowCasters(fs *n3ds.RomFS, places []schema.Placement, cache map[string]*objectArchive) ([]shadowCaster, error) {
	byID := map[string]actor{}
	for _, a := range actors {
		byID[a.id] = a
	}
	var out []shadowCaster
	for _, pl := range places {
		a, ok := byID[pl.Object]
		if !ok {
			continue
		}
		// Only the objects the game renders into the depth-shadow pass. Its own
		// InitExecutor says which, and the characters are not among them.
		raw, err := fs.File("/ObjectData/" + a.object + ".szs")
		if err != nil {
			return nil, err
		}
		arc, err := n3ds.OpenSZS(raw)
		if err != nil {
			return nil, err
		}
		casts, err := castsDepthShadow(arc)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", a.id, err)
		}
		if !casts {
			continue
		}
		o, err := loadObject(fs, a.object, cache)
		if err != nil {
			return nil, err
		}
		var proxy *n3ds.BCHModel
		for _, m := range o.Extra {
			if strings.HasSuffix(m.Name, "_shd") {
				proxy = m
			}
		}
		if proxy == nil {
			continue // an actor with no proxy casts nothing, as in the game
		}
		world, err := posedWorld(fs, proxy, a, cache)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", a.id, err)
		}
		for i := range proxy.Meshes {
			sh := &proxy.Meshes[i]
			pr := glb.Prim{BaseColor: [4]float32{0, 0, 0, 1}, Unlit: true}
			pr.Positions = posedPositions(proxy, sh, world)
			pr.Tris = make([][3]uint32, len(sh.Indices)/3)
			for t := range pr.Tris {
				pr.Tris[t] = [3]uint32{sh.Indices[t*3], sh.Indices[t*3+1], sh.Indices[t*3+2]}
			}
			out = append(out, shadowCaster{name: a.id + "-shadow", prim: pr, pl: pl})
		}
	}
	return out, nil
}

// posedWorld composes the proxy's skeleton at frame 0 of the actor's pose clip.
func posedWorld(fs *n3ds.RomFS, proxy *n3ds.BCHModel, a actor, cache map[string]*objectArchive) ([][16]float64, error) {
	if a.pose == "" {
		return proxy.BindPose(), nil
	}
	sources := []*objectArchive{}
	o, err := loadObject(fs, a.object, cache)
	if err != nil {
		return nil, err
	}
	sources = append(sources, o)
	if a.animArc != "" {
		an, err := loadObject(fs, a.animArc, cache)
		if err != nil {
			return nil, err
		}
		sources = append(sources, an)
	}
	for _, src := range sources {
		for _, f := range src.BCHs {
			for _, e := range f.Groups[n3ds.BCHSkeletalAnims] {
				if e.Name != a.pose {
					continue
				}
				an, err := f.DecodeSkeletalAnim(e)
				if err != nil {
					return nil, err
				}
				posed := &n3ds.BCHModel{Bones: an.PoseAt(proxy.Bones, 0)}
				return posed.BindPose(), nil
			}
		}
	}
	return proxy.BindPose(), nil
}

// posedPositions skins one mesh by the rule its own header states — the same
// two spaces the visible meshes use (tools/platform/n3ds/bchglb).
func posedPositions(model *n3ds.BCHModel, sh *n3ds.BCHMesh, world [][16]float64) [][3]float32 {
	out := make([][3]float32, len(sh.Verts))
	mats := make([][16]float64, len(sh.Palette))
	for i, b := range sh.Palette {
		if b < 0 || b >= len(world) {
			continue
		}
		if sh.SkinMode == n3ds.BCHSkinSmooth {
			mats[i] = mulRow4(world[b], model.Bones[b].InvBind4())
		} else {
			mats[i] = world[b]
		}
	}
	for vi, v := range sh.Verts {
		if len(sh.Palette) == 0 {
			out[vi] = v.Pos
			continue
		}
		var p [3]float64
		for k, w := range v.Weights {
			if w == 0 {
				continue
			}
			m := mats[v.Joints[k]]
			for r := 0; r < 3; r++ {
				p[r] += float64(w) * (m[r*4]*float64(v.Pos[0]) + m[r*4+1]*float64(v.Pos[1]) +
					m[r*4+2]*float64(v.Pos[2]) + m[r*4+3])
			}
		}
		out[vi] = [3]float32{float32(p[0]), float32(p[1]), float32(p[2])}
	}
	return out
}

func mulRow4(a, b [16]float64) [16]float64 {
	var o [16]float64
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			s := 0.0
			for k := 0; k < 4; k++ {
				s += a[r*4+k] * b[k*4+c]
			}
			o[r*4+c] = s
		}
	}
	return o
}

// --- shadows ------------------------------------------------------------

// The draw category an object's InitExecutor.byml lists when it renders into
// the depth-shadow pass. The stage terrain lists the terrain one; the goal star
// lists the character one; Captain Toad and Toadette list NEITHER.
//
// That is the answer to whether a placed actor casts: the characters do not.
// What they have instead is an InitShadowMask — a blob projected under them,
// with its own radius, drop length and colour — which is a different mechanism
// and the one that puts a shadow under a character in this game.
const depthShadowCategory = "デプスシャドウ"

// shadowMask is the blob an object drops beneath itself, as its own
// InitShadowMask.byml configures it. Radius is in world units; DropLength is
// how far below the joint the game looks for a surface; Colour is the
// multiplier it tints with.
type shadowMask struct {
	joint  string
	offset [3]float32
	radius float32
	drop   float32
	colour [3]float32
}

// castsDepthShadow reports whether an object's InitExecutor lists it in the
// depth-shadow pass.
func castsDepthShadow(arc *n3ds.SARC) (bool, error) {
	blob, ok := arc.File("InitExecutor.byml")
	if !ok {
		return false, nil
	}
	doc, err := n3ds.ParseBYML(blob)
	if err != nil {
		return false, err
	}
	d, _ := doc.(n3ds.BYMLDict)
	for _, it := range list(d["Drawer"]) {
		e, _ := it.(n3ds.BYMLDict)
		if c, _ := e["CategoryName"].(string); strings.HasPrefix(c, depthShadowCategory) {
			return true, nil
		}
	}
	return false, nil
}

// readShadowMask reads an object's blob-shadow configuration.
func readShadowMask(arc *n3ds.SARC) (*shadowMask, error) {
	blob, ok := arc.File("InitShadowMask.byml")
	if !ok {
		return nil, nil
	}
	doc, err := n3ds.ParseBYML(blob)
	if err != nil {
		return nil, err
	}
	d, _ := doc.(n3ds.BYMLDict)
	arr := list(d["ShadowMaskArray"])
	if len(arr) == 0 {
		return nil, nil
	}
	m, _ := arr[0].(n3ds.BYMLDict)
	sm := &shadowMask{
		joint:  str(m["ActorJointName"]),
		offset: vec3(m["Offset"]),
		radius: float32(num(m["Radius"])),
		drop:   float32(num(m["DropLength"])),
	}
	if c, ok := m["Color"].(n3ds.BYMLDict); ok {
		sm.colour = [3]float32{float32(num(c["R"])), float32(num(c["G"])), float32(num(c["B"]))}
	}
	// An ellipsoid mask states its extent as a scale rather than a radius.
	if s, ok := m["Scale"].(n3ds.BYMLDict); ok {
		if v := float32(num(s["X"])); v > 0 {
			sm.radius = v
		}
	}
	if sm.radius <= 0 {
		return nil, fmt.Errorf("shadow mask %q has no radius", str(m["Name"]))
	}
	return sm, nil
}

func str(v any) string { s, _ := v.(string); return s }

// placedMask is one actor's blob shadow in world space, ready to be dropped
// onto whatever is beneath it.
type placedMask struct {
	name   string
	centre [3]float32
	radius float32
	drop   float32
	colour [3]float32
}

// actorShadowMasks resolves each placed actor's blob shadow: where its own
// InitShadowMask hangs it, in world space.
//
// The joint the mask names is posed with the clip the placement starts, so a
// character crouching drops its shadow lower — and then the placement's own
// transform puts it in the stage.
func actorShadowMasks(fs *n3ds.RomFS, places []schema.Placement, cache map[string]*objectArchive) ([]placedMask, error) {
	byID := map[string]actor{}
	for _, a := range actors {
		byID[a.id] = a
	}
	var out []placedMask
	for _, pl := range places {
		a, ok := byID[pl.Object]
		if !ok {
			continue
		}
		raw, err := fs.File("/ObjectData/" + a.object + ".szs")
		if err != nil {
			return nil, err
		}
		arc, err := n3ds.OpenSZS(raw)
		if err != nil {
			return nil, err
		}
		sm, err := readShadowMask(arc)
		if err != nil || sm == nil {
			if err != nil {
				return nil, fmt.Errorf("%s: %w", a.id, err)
			}
			continue
		}
		o, err := loadObject(fs, a.object, cache)
		if err != nil {
			return nil, err
		}
		model := o.Model
		if model == nil {
			continue
		}
		world, err := posedWorld(fs, model, a, cache)
		if err != nil {
			return nil, err
		}
		// The joint the mask hangs from.
		local := sm.offset
		found := false
		for i, b := range model.Bones {
			if b.Name != sm.joint {
				continue
			}
			w := world[i]
			local = [3]float32{
				float32(w[3]) + sm.offset[0],
				float32(w[7]) + sm.offset[1],
				float32(w[11]) + sm.offset[2],
			}
			found = true
			break
		}
		if !found {
			return nil, fmt.Errorf("%s: shadow mask hangs from joint %q, which the model does not have", a.id, sm.joint)
		}
		// Into the stage, through the placement's own transform.
		p := applyPlacement(local, pl)
		out = append(out, placedMask{
			name: a.id + "-mask", centre: p, radius: sm.radius, drop: sm.drop, colour: sm.colour,
		})
	}
	return out, nil
}

// applyPlacement puts a model-space point into the stage, through a placement's
// rotation and translation.
func applyPlacement(p [3]float32, pl schema.Placement) [3]float32 {
	r := vec3f(pl.Rot)
	const d2r = math.Pi / 180
	// The map's rotation composes X then Y then Z (eulerXYZ), so the matrix is
	// Rz·Ry·Rx and the point goes through it in that order.
	x, y, z := float64(p[0]), float64(p[1]), float64(p[2])
	sx, cx := math.Sin(float64(r[0])*d2r), math.Cos(float64(r[0])*d2r)
	y, z = y*cx-z*sx, y*sx+z*cx
	sy, cy := math.Sin(float64(r[1])*d2r), math.Cos(float64(r[1])*d2r)
	x, z = x*cy+z*sy, -x*sy+z*cy
	sz, cz := math.Sin(float64(r[2])*d2r), math.Cos(float64(r[2])*d2r)
	x, y = x*cz-y*sz, x*sz+y*cz
	t := vec3f(pl.Pos)
	return [3]float32{float32(x) + t[0], float32(y) + t[1], float32(z) + t[2]}
}
