// webexport builds Super Mario 64 DS's Retro-X game tree from the raw
// cartridge. It stages the filesystem + decompressed binaries in a temp dir,
// renders the SDAT music, runs the ACTOR ORACLE (the game's own actor
// create/init code, natively — no heuristics) to bind every placed actor to
// its model, and writes:
//
//	levels/<stem>.json          per stage: the stage model placed at the
//	                            origin, the vrbox skybox as a camera-attached
//	                            layer, the stage KCL as a toggleable
//	                            collision layer, and every placed actor at
//	                            its oracle-bound model
//	levels/col_<kcl>.glb        the stage collision meshes
//	objects/<id>.json|.glb      stages, characters, enemies, objects,
//	                            skyboxes and oracle-named archive members
//	music/<stem>.mp3            every renderable SSEQ, via tools/nds/sdat
//
// Usage (from games/super-mario-64-ds/):
//
//	go run ./extract/cmd/webexport -in "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"
package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
	"retroreverse.com/tools/lib/retrox/audio"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/nitro"
	"retroreverse.com/tools/platform/nds/sdat"
)

const (
	rate     = 32768             // DS mixer output rate
	toGLB    = 1.0 / 4096 / 1000 // fx20.12 world -> stage-GLB units
	toStage  = 1.0 / 1000        // placement short -> stage-GLB units
	objScale = 1.0 / 125         // object display scale in stage-GLB units
)

func main() {
	cli.Main("super-mario-64-ds", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in <rom.nds> [-o DIR] [-only levels,objects,music]")
	}

	b := ctx.Builder
	b.SetTitle("Super Mario 64 DS")
	b.SetPlatform("Nintendo DS")
	b.SetYear(2004)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 256, H: 192},
		TickHz: 60,
		// The DS's backlit TFT: the colour dot-matrix profile, and
		// point-sampled texels (the DS renders unfiltered).
		Filter:    "ds",
		TexFilter: "nearest",
	})

	if ctx.Stage("music") {
		if err := runMusic(ctx, ctx.In); err != nil {
			return err
		}
	}

	if ctx.Enabled("objects") || ctx.Enabled("levels") {
		tmp, err := os.MkdirTemp("", "sm64ds-webexport-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmp)
		if err := extractFS(ctx.In, tmp); err != nil {
			return err
		}

		ls, err := sm64ds.OpenLevels(ctx.In, tmp)
		if err != nil {
			return err
		}
		if err := buildLevelNames(ls, tmp); err != nil {
			return err
		}

		bindings, err := runOracle(ctx, ctx.In, tmp)
		if err != nil {
			return err
		}

		if ctx.Stage("objects") {
			if err := exportModels(ctx, ls, tmp); err != nil {
				return err
			}
			if err := exportArchiveGLBs(ctx, ls, bindings); err != nil {
				return err
			}
		}
		if ctx.Stage("levels") {
			if err := exportCollision(ctx, ls, tmp); err != nil {
				return err
			}
			if err := exportLevels(ctx, ls, tmp, bindings); err != nil {
				return err
			}
			logObjColliders(ctx)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// music
// ---------------------------------------------------------------------------

func runMusic(ctx *cli.Context, romPath string) error {
	img, err := os.ReadFile(romPath)
	if err != nil {
		return err
	}
	rom, err := nds.Open(img)
	if err != nil {
		return err
	}
	data := rom.FileByPath("data/sound_data.sdat")
	if data == nil {
		return fmt.Errorf("data/sound_data.sdat not found in ROM")
	}
	s, err := sdat.Parse(data)
	if err != nil {
		return err
	}

	total := 0
	for i := range s.Seqs {
		if s.Seqs[i].FileID >= 0 {
			total++
		}
	}
	ctx.Logf("SDAT: %d sequences (%d renderable)", len(s.Seqs), total)

	n := 0
	for i := range s.Seqs {
		if s.Seqs[i].FileID < 0 {
			continue
		}
		n++
		stem := stemFor(i, s.Seqs[i].Name)
		L, R, err := s.Render(i, rate, 2, 180)
		if err != nil {
			ctx.Logf("%s: %v", stem, err)
			continue
		}
		if len(L) < rate { // sub-second jingles cut short
			continue
		}
		fadeOut(L, R)
		samples := make([]int16, len(L)*2)
		clip := func(v float64) int16 {
			if v > 1 {
				v = 1
			}
			if v < -1 {
				v = -1
			}
			return int16(v * 32767)
		}
		for k := range L {
			samples[k*2] = clip(L[k])
			samples[k*2+1] = clip(R[k])
		}
		out, err := ctx.Builder.Path("music", stem+".mp3")
		if err != nil {
			return err
		}
		wave := audio.PCM16{Rate: rate, Channels: 2, Samples: samples}
		if err := audio.EncodeMP3(wave, out); err != nil {
			ctx.Logf("%s: %v", stem, err)
			continue
		}
		ctx.Builder.AddMedia(schema.Asset{
			ID: "bgm-" + stem, Category: schema.CategoryMusic,
			Name:     stem,
			File:     "music/" + stem + ".mp3",
			Duration: wave.Duration(),
		})
		ctx.Progress("music", n, total, stem+".mp3")
	}
	return nil
}

func stemFor(i int, name string) string {
	if name == "" {
		return fmt.Sprintf("seq_%02d", i)
	}
	return strings.ToLower(strings.TrimPrefix(name, "NCS_BGM_"))
}

func fadeOut(L, R []float64) {
	n := 3 * rate
	if n > len(L) {
		n = len(L)
	}
	for i := 0; i < n; i++ {
		g := float64(n-i) / float64(n)
		L[len(L)-n+i] *= g
		R[len(R)-n+i] *= g
	}
}

// ---------------------------------------------------------------------------
// filesystem staging
// ---------------------------------------------------------------------------

// extractFS writes the decompressed ARM9 binary, every ARM9 overlay's
// decompressed image (ovl9_NNN_dec.bin), and the full filesystem under
// files/, into dir.
func extractFS(romPath, dir string) error {
	img, err := os.ReadFile(romPath)
	if err != nil {
		return err
	}
	rom, err := nds.Open(img)
	if err != nil {
		return err
	}
	arm9 := rom.ARM9()
	if nds.IsBLZ(arm9) {
		arm9 = nds.DecompressBLZ(arm9)
	}
	if err := os.WriteFile(filepath.Join(dir, "arm9_dec.bin"), arm9, 0o644); err != nil {
		return err
	}
	for _, o := range rom.ARM9Overlays() {
		raw := rom.File(int(o.FileID))
		dec := raw
		if o.Compressed && nds.IsBLZ(raw) {
			dec = nds.DecompressBLZ(raw)
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("ovl9_%03d_dec.bin", o.ID)), dec, 0o644); err != nil {
			return err
		}
	}
	for _, f := range rom.Files {
		p := filepath.Join(dir, "files", filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, rom.File(f.ID), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// oracle
// ---------------------------------------------------------------------------

// Binding mirrors the actor oracle's table entry.
type Binding struct {
	Params    [3]int
	Config    int
	Models    []string
	Clips     []string
	KCL       []string
	Colliders []sm64ds.Collider
	Notes     []string
}

func runOracle(ctx *cli.Context, romPath, tmp string) (map[int][]Binding, error) {
	ls, err := sm64ds.OpenLevels(romPath, tmp)
	if err != nil {
		return nil, err
	}
	o, err := sm64ds.NewOracle(ls)
	if err != nil {
		return nil, err
	}
	if err := o.InitEngine(); err != nil {
		return nil, err
	}
	ctx.Logf("oracle: engine initialized (%d init-phase file requests)", len(o.InitRequests()))
	table := sweep(ls, o)
	ctx.Logf("oracle: %d actors bound", len(table))

	// Second pass: which placements exist in which mission, decided by the
	// actors' own code under a seeded progress context (sm64ds.MissionGates).
	missionGates = sm64ds.MissionGates(ls, o, func(actor int, par [3]int) (int, bool) {
		bs := table[actor]
		if len(bs) == 0 {
			return -1, false
		}
		for _, b := range bs {
			if b.Params == par {
				return b.Config, true
			}
		}
		for _, b := range bs {
			if b.Params[0] == par[0] {
				return b.Config, true
			}
		}
		return bs[0].Config, true
	}, func(stem string, n int) {
		ctx.Logf("oracle: %s — %d mission-gated placements", stem, n)
	})
	return table, nil
}

// missionGates: stage stem -> the placements whose existence varies by mission.
var missionGates map[string][]sm64ds.MissionGate

// sweep runs every distinct (actor, params) the levels place.
func sweep(ls *sm64ds.LevelSet, o *sm64ds.Oracle) map[int][]Binding {
	type combo struct{ actor, p1, p2, p3 int }
	perLevel := map[int]map[combo]bool{}
	all := map[combo][]int{}
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			continue
		}
		if perLevel[lv.Overlay] == nil {
			perLevel[lv.Overlay] = map[combo]bool{}
		}
		for _, ob := range lv.Objects {
			c := combo{ob.Actor, ob.Params[0], ob.Params[1], ob.Params[2]}
			if !perLevel[lv.Overlay][c] {
				perLevel[lv.Overlay][c] = true
				all[c] = append(all[c], lv.Overlay)
			}
		}
	}

	table := map[int][]Binding{}
	addRun := func(run *sm64ds.ActorRun) {
		if len(run.Files) == 0 && run.Obj == 0 {
			return
		}
		b := Binding{Params: run.Params, Config: run.Config, Models: o.Models(run), Clips: o.Clips(run), KCL: o.KCLs(run), Colliders: run.Colliders, Notes: run.Notes}
		for _, e := range table[run.Actor] {
			if fmt.Sprint(e.Models, e.Params) == fmt.Sprint(b.Models, b.Params) {
				return
			}
		}
		table[run.Actor] = append(table[run.Actor], b)
	}

	unresolved := map[combo]bool{}
	var ovls []int
	for ov := range perLevel {
		ovls = append(ovls, ov)
	}
	sort.Ints(ovls)
	for _, ov := range ovls {
		if err := o.LoadConfig(ov); err != nil {
			continue
		}
		for c := range perLevel[ov] {
			if _, ok := o.Profile(c.actor, ov); !ok {
				unresolved[c] = true
				continue
			}
			// An actor may reach helpers in a bank the game had resident
			// alongside this overlay; RunActorBanked finds which and retries.
			addRun(o.RunActorBanked(c.actor, ov, [3]int{c.p1, c.p2, c.p3},
				func(extra int) error {
					if extra < 0 {
						return o.LoadConfig(ov)
					}
					return o.LoadConfigMulti([]int{ov, extra})
				}))
		}
	}

	bankRuns := map[string]bool{}
	for bank := 60; bank <= 102; bank++ {
		var todo []combo
		for c := range unresolved {
			todo = append(todo, c)
		}
		if len(todo) == 0 {
			break
		}
		if err := o.LoadConfig(bank); err != nil {
			continue
		}
		for _, c := range todo {
			if _, ok := o.Profile(c.actor, bank); !ok {
				continue
			}
			key := fmt.Sprint(bank, c)
			if bankRuns[key] {
				continue
			}
			bankRuns[key] = true
			addRun(o.RunActorBanked(c.actor, bank, [3]int{c.p1, c.p2, c.p3},
				func(extra int) error {
					if extra < 0 {
						return o.LoadConfig(bank)
					}
					return o.LoadConfigMulti([]int{bank, extra})
				}))
		}
	}
	return table
}

// ---------------------------------------------------------------------------
// objects: BMD -> GLB
// ---------------------------------------------------------------------------

// levelStems records the display name of every stage model, and stageAssets
// their asset ids; refs maps every model stem to its asset id.
var (
	levelStems = map[string]string{}
	refs       = map[string]string{}
	// modelFloor[stem] is the model's lowest vertex in model units. Most SM64DS
	// models are authored standing on y = 0, so this is 0 and nothing needs it;
	// the chain chomp's body is a ball centred on its origin, so it is not.
	modelFloor = map[string]float64{}
	usedObjIDs = map[string]bool{}
)

func objectID(stem string) string {
	id := slugify(stem)
	for n := 2; usedObjIDs[id]; n++ {
		id = fmt.Sprintf("%s-%d", slugify(stem), n)
	}
	usedObjIDs[id] = true
	return id
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

// idleFirst reorders clips so a model's RESTING animation comes first.
//
// It matters because the first clip is the one a viewer autoplays: level3d
// starts the first looping animation on every placement, and the exporter used
// to hand it whatever sorted first alphabetically. For a goomba that is
// kuribo_recover — the get-up-after-being-stomped animation — so a battlefield
// of goombas all played being hit, forever. Bob-ombs got bombhei_carry, King
// Bob-omb got kuriking_anger, Hanachan got its damage clip.
//
// The cartridge labels its own idles: `wait` is the commonest suffix in the
// whole game (114 clips of 91 animated models), and Mario's is su_wait, the
// goomba's kuribo_wait, the bob-omb's red_wait. So this reads the game's naming
// rather than guessing at the animation data — but it IS the naming and not a
// trace of which clip each actor's init selects, which the oracle does not
// record. Where a model has no wait clip the second tier prefers steady motion
// (a moray eel's idle really is moray_swim) and the third merely avoids the
// obvious one-shots.
func idleRank(name string) int {
	n := strings.ToLower(name)
	base := strings.TrimRight(n, "0123456789")
	switch {
	case strings.HasSuffix(base, "wait"), strings.HasSuffix(base, "stay"),
		strings.HasSuffix(base, "idle"), strings.HasSuffix(base, "stand"):
		return 0
	}
	for _, w := range []string{"swim", "fly", "walk", "run", "move", "spin", "roll"} {
		if strings.HasSuffix(base, w) {
			return 1
		}
	}
	// Reactions and transitions: never a resting pose, and the reason this
	// function exists.
	for _, w := range []string{"dead", "damage", "hit", "recover", "attack", "anger",
		"throw", "jump", "land", "start", "end", "appear", "unbalance", "stretch", "turn"} {
		if strings.HasSuffix(base, w) {
			return 3
		}
	}
	return 2
}

// addChomp reconstructs a chain chomp: the stake where the game spawns it, the
// body out on its chain, and the chain drawn between them.
//
// Placed flat, all three of those coincide and the result is wrong three ways.
// The placement records the chomp's ANCHOR — its stake — and the chomp's init
// spawns a pile (actor 27) there, which is traced. But the BODY is a ball whose
// origin is its own centre, so at the anchor's ground-level y it is buried to
// the eyeballs; and the chain is not an actor at all, it is strung by the
// chomp's own draw function ($021437D4), so nothing places it and the level
// shipped without one.
//
// What is traced: the stake at the anchor, and the chain's body-space anchor
// vector (0, 0, -250). What is NOT: where the AI happens to have left the chomp,
// which is a live handler this exporter cannot run. So the body is placed at
// rest on the far side of its chain and the links strung evenly to it — a
// reconstruction, chosen to look like the game rather than derived from it, and
// the only part of this file that is.
//
// The lift is measured, not chosen: modelFloor is the model's own lowest vertex,
// so the ball stands on the point the level data gives rather than through it.
// It is applied to the chomps ALONE and not as a rule. Eleven of the thirty-five
// models this level places have geometry below their origin, and most of them
// belong there — a star hovers, a lift has a lip — so lifting everything would
// raise things that are already right. The chomp is the outlier at 2.12 m.
func addChomp(o sm64ds.LevelObject,
	addObjOff func(sm64ds.LevelObject, int, [3]int, bool, [3]float64),
	addModelAt func(string, [3]float64, string)) {
	const chainLinks = 6 // reconstructed: enough to read as a chain across the gap
	// The stake, where the trace puts it.
	addObjOff(o, 27, [3]int{65535, 0, 0}, false, [3]float64{})

	// Out along the placement's own heading, so the chomp at least faces the way
	// the level data says it does.
	yaw := o.RotY * math.Pi / 180
	dx, dz := math.Sin(yaw), math.Cos(yaw)

	// Both measured off the models rather than picked: the ball's radius (it is
	// a sphere centred on its origin, so its lowest vertex IS its radius) and a
	// link's own width, stepped at 80% so consecutive links overlap into a chain
	// instead of a dotted line.
	radius := -modelFloor["ar1_2"] * objScale
	step := -modelFloor["ar1_1"] * 2 * objScale * 0.8
	gap := chainLinks * step
	span := radius + gap // stake to BALL CENTRE, so the gap is clear of the ball

	addObjOff(o, o.Actor, o.Params, true, [3]float64{dx * span, radius, dz * span})

	// The links fill the gap only — spanning to the centre instead put the last
	// two inside the ball, where a chain is of no use to anybody. They rise from
	// the stake's collar to the height the chain meets the ball.
	const collar = 0.06
	for i := 0; i < chainLinks; i++ {
		d := step * (float64(i) + 0.5)
		t := d / gap
		addModelAt("ar1_1", [3]float64{
			o.X*toStage + dx*d,
			o.Y*toStage + collar + (radius-collar)*t,
			o.Z*toStage + dz*d,
		}, "Chain link")
	}
}

// clipAnims turns .bca clip names into animation metadata (30 fps loops), the
// resting one first.
func clipAnims(clips []sm64ds.NamedBCA) []schema.Animation {
	var out []schema.Animation
	for _, c := range clips {
		out = append(out, schema.Animation{ID: c.Name, Clip: c.Name, FPS: 30, Loop: "loop"})
	}
	// Stable, so clips of equal rank keep the alphabetical order they arrived in
	// — except that among equals the SHORTER name wins, because the extra words
	// are qualifiers: King Bob-omb has both kuriking_wait and
	// kuriking_serch_wait, and the plain one is the idle.
	sort.SliceStable(out, func(i, j int) bool {
		ri, rj := idleRank(out[i].ID), idleRank(out[j].ID)
		if ri != rj {
			return ri < rj
		}
		return len(out[i].ID) < len(out[j].ID)
	})
	return out
}

func exportModels(ctx *cli.Context, ls *sm64ds.LevelSet, tmp string) error {
	b := ctx.Builder
	root := filepath.Join(tmp, "files")

	var paths []string
	filepath.Walk(filepath.Join(root, "data"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".bmd") {
			paths = append(paths, p)
		}
		return nil
	})
	sort.Strings(paths)

	seen := map[string]bool{}
	n, listed := 0, 0
	for _, p := range paths {
		n++
		m, err := sm64ds.LoadBMD(p)
		if err != nil {
			continue
		}
		// sibling .bca clips whose bone count matches become glTF animations
		var clips []sm64ds.NamedBCA
		if sib, _ := filepath.Glob(filepath.Join(filepath.Dir(p), "*.bca")); len(sib) > 0 {
			for _, ap := range sib {
				if a, err := sm64ds.LoadBCA(ap); err == nil && a.NumBones == m.NumBones {
					stem := strings.TrimSuffix(filepath.Base(ap), ".bca")
					clips = append(clips, sm64ds.NamedBCA{Name: stem, Anim: a})
				}
			}
		}
		var data []byte
		if len(clips) > 0 {
			data, err = m.SkinnedGLB(clips)
		} else {
			data, err = m.GLB()
		}
		if err != nil {
			continue
		}
		file := m.Name + ".glb"
		if seen[file] {
			continue
		}
		seen[file] = true
		gp, err := b.Path("objects", file)
		if err != nil {
			return err
		}
		if err := os.WriteFile(gp, data, 0o644); err != nil {
			return err
		}
		name, sec := classify(p, m.Name)
		if sec == "Levels" {
			levelStems[m.Name] = name // placed by its level; grouped as a stage model
			sec = "Stages"
		}
		if sec == "" {
			// referenced-only models (actor bindings) still need an asset
			name, sec = title(m.Name), "Other models"
		}
		id := objectID(m.Name)
		doc := &schema.Object{
			Type: schema.ObjectModel3D, Name: name, Model: file,
			SkinnedClone: len(clips) > 0,
			Animations:   clipAnims(clips),
			Billboard:    billboardMode(m),
		}
		b.AddObject(schema.Asset{ID: id, Name: name, Group: sec}, doc)
		refs[m.Name] = id
		listed++
		if n%100 == 0 {
			ctx.Progress("objects", n, len(paths), fmt.Sprintf("%d models", listed))
		}
	}

	// The playable Mario: the 16-bone minigame model (MG/, outside data/) with
	// the standard clips from data/player.
	if m, err := sm64ds.LoadBMD(filepath.Join(root, "MG/mario_model_mg.bmd")); err == nil {
		var clips []sm64ds.NamedBCA
		for _, cn := range []string{"su_wait", "su_walk", "su_run"} {
			if a, err := sm64ds.LoadBCA(filepath.Join(root, "data/player", cn+".bca")); err == nil && a.NumBones == m.NumBones {
				clips = append(clips, sm64ds.NamedBCA{Name: cn, Anim: a})
			}
		}
		if data, err := m.SkinnedGLB(clips); err == nil {
			gp, err := b.Path("objects", "mario_model_mg.glb")
			if err != nil {
				return err
			}
			if err := os.WriteFile(gp, data, 0o644); err != nil {
				return err
			}
			id := objectID("mario_model_mg")
			b.AddObject(schema.Asset{ID: id, Name: "Mario (in-game)", Group: "Characters"}, &schema.Object{
				Type: schema.ObjectModel3D, Name: "Mario (in-game)", Model: "mario_model_mg.glb",
				SkinnedClone: true, Animations: clipAnims(clips),
			})
			refs["mario_model_mg"] = id
		}
	}
	ctx.Progress("objects", len(paths), len(paths), fmt.Sprintf("%d models exported", listed))
	return nil
}

// exportArchiveGLBs decodes archive-member models the bindings name (arcN_M).
func exportArchiveGLBs(ctx *cli.Context, ls *sm64ds.LevelSet, bindings map[int][]Binding) error {
	b := ctx.Builder
	done := map[string]bool{}
	var stems []string
	for _, bs := range bindings {
		for _, bd := range bs {
			for _, stem := range bd.Models {
				if strings.LastIndexByte(stem, '_') < 0 || done[stem] {
					continue
				}
				done[stem] = true
				stems = append(stems, stem)
			}
		}
	}
	sort.Strings(stems)
	n := 0
	for _, stem := range stems {
		if _, ok := refs[stem]; ok {
			continue
		}
		ref, ok := archiveRefByStem(ls, stem)
		if !ok {
			continue
		}
		data, err := ls.ArchiveMember(ref)
		if err != nil || !sm64ds.PlausibleBMD(data) {
			continue
		}
		m, err := sm64ds.Decode(data, stem)
		if err != nil {
			continue
		}
		glbData, err := m.GLB()
		if err != nil {
			continue
		}
		gp, err := b.Path("objects", stem+".glb")
		if err != nil {
			return err
		}
		if err := os.WriteFile(gp, glbData, 0o644); err != nil {
			return err
		}
		lo := 0.0
		for _, tris := range m.ByMat {
			for _, t := range tris {
				for _, v := range t.V {
					if v.Y < lo {
						lo = v.Y
					}
				}
			}
		}
		modelFloor[stem] = lo
		id := objectID(stem)
		b.AddObject(schema.Asset{ID: id, Name: title(stem), Group: "Archive members"}, &schema.Object{
			Type: schema.ObjectModel3D, Name: title(stem), Model: stem + ".glb",
			Billboard: billboardMode(m),
		})
		refs[stem] = id
		n++
	}
	ctx.Logf("%d archive-member models decoded", n)
	return nil
}

func archiveRefByStem(ls *sm64ds.LevelSet, stem string) (sm64ds.ArchiveRef, bool) {
	i := strings.LastIndexByte(stem, '_')
	if i < 0 {
		return sm64ds.ArchiveRef{}, false
	}
	var member int
	if _, err := fmt.Sscanf(stem[i+1:], "%d", &member); err != nil {
		return sm64ds.ArchiveRef{}, false
	}
	name := stem[:i]
	for _, arc := range []string{"arc0", "ar1", "c2d", "cee", "cef", "ceg", "cei", "ces", "en1", "vs1", "vs2", "vs3", "vs4"} {
		if name == arc {
			return sm64ds.ArchiveRef{Archive: name, Member: member}, true
		}
	}
	return sm64ds.ArchiveRef{}, false
}

// classify assigns a model to a group with a friendly name, by its path.
func classify(path, stem string) (name, section string) {
	switch {
	case strings.Contains(path, "/stage/") && strings.HasSuffix(stem, "_all"):
		base := strings.TrimSuffix(stem, "_all")
		if n := levelNames[base]; n != "" {
			return n, "Levels"
		}
		return base, "Levels"
	case strings.Contains(path, "/player/"):
		return title(stem), "Characters"
	case strings.Contains(path, "/enemy/"):
		return title(stem), "Enemies"
	case strings.Contains(path, "/normal_obj/"), strings.Contains(path, "/special_obj/"):
		return title(stem), "Objects"
	case strings.Contains(path, "/vrbox/"):
		return "Skybox " + strings.TrimPrefix(stem, "vr"), "Skyboxes"
	}
	return "", ""
}

func title(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	if len(s) > 0 {
		s = strings.ToUpper(s[:1]) + s[1:]
	}
	return s
}

// ---------------------------------------------------------------------------
// level display names
// ---------------------------------------------------------------------------

var levelNames = map[string]string{}

var hubNames = map[string]string{
	"main_castle": "Peach's Castle (exterior)", "main_garden": "Castle Grounds",
	"castle_1f": "Castle — 1st floor", "castle_2f": "Castle — 2nd floor",
	"castle_b1": "Castle — basement", "playroom": "Playroom",
	"test_map": "Test map", "test_map_b": "Test map B",
}

func buildLevelNames(ls *sm64ds.LevelSet, tmp string) error {
	msgs, err := sm64ds.LoadBMG(filepath.Join(tmp, "files/data/message/msg_data_eng.bin"))
	if err != nil {
		return err
	}
	type lv struct{ id, course int }
	byStem := map[string]lv{}
	courseCount := map[int]int{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		l, err := ls.Level(id)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(l.BMDPath), ".bmd"), "_all")
		c := ls.Course(id)
		if _, dup := byStem[stem]; !dup {
			byStem[stem] = lv{id, c}
			courseCount[c]++
		}
	}
	first := map[int]int{}
	for _, v := range byStem {
		if f, ok := first[v.course]; !ok || v.id < f {
			first[v.course] = v.id
		}
	}
	for stem, v := range byStem {
		if n, ok := hubNames[stem]; ok {
			levelNames[stem] = n
			continue
		}
		if v.course < 0 || v.course+sm64ds.CourseNameMsg >= len(msgs) {
			levelNames[stem] = stem
			continue
		}
		name := courseTitle(msgs[v.course+sm64ds.CourseNameMsg])
		if courseCount[v.course] > 1 && first[v.course] != v.id {
			name += " (" + stem + ")"
		}
		levelNames[stem] = name
	}
	return nil
}

// courseTitle title-cases a course name, dropping the leading course number
// the ROM prints in the pause menu (" 1 BOB-OMB BATTLEFIELD").
func courseTitle(msg string) string {
	s := strings.TrimSpace(msg)
	s = strings.TrimLeft(s, "0123456789")
	return titleCase(s)
}

// starTitle title-cases a mission name. Unlike a course name it keeps a
// leading number, which is part of the title ("5 SILVER STARS!", "8-COIN
// PUZZLE WITH 15 PIECES"). The ROM SHOUTS every one of these; the raw string
// is what StarName returns, so the transformation stays reversible.
func starTitle(msg string) string { return titleCase(msg) }

func titleCase(s string) string {
	s = strings.TrimSpace(s)
	small := map[string]bool{
		"IN": true, "THE": true, "OF": true, "ON": true, "TO": true, "UNDER": true,
		"A": true, "AN": true, "AT": true, "FOR": true, "FROM": true, "WITH": true,
		"INTO": true, "AND": true, "OVER": true,
	}
	words := strings.Fields(s)
	for i, w := range words {
		if i > 0 && small[w] {
			words[i] = strings.ToLower(w)
			continue
		}
		// "BOB-OMB", "BOB-OMB'S" — the second half of the name stays lower
		// case, against the hyphen rule below
		if strings.HasPrefix(w, "BOB-OMB") {
			words[i] = "Bob-omb" + strings.ToLower(w[len("BOB-OMB"):])
			continue
		}
		r := []rune(strings.ToLower(w))
		up := true
		for j, c := range r {
			if up {
				r[j] = []rune(strings.ToUpper(string(c)))[0]
			}
			// a hyphen opens a new word ("WET-DRY"), and so does a leading
			// apostrophe ("'SHROOMS")
			up = c == '-' || (j == 0 && c == '\'')
		}
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

// ---------------------------------------------------------------------------
// collision: KCL -> GLB (stage meshes only; object colliders stay as props)
// ---------------------------------------------------------------------------

// colFile maps a stage KCL stem to its exported levels/col_<stem>.glb.
var colFile = map[string]string{}

// objColFile memoises the OBJECT colliders exported to levels/objcol_<stem>.glb.
// A level's own .kcl is only half its walkable surface: the see-saw bridge in
// Bob-omb Battlefield, the lifts, the shutters and the chomp's stake are all
// placed actors carrying their own collider, and until these were emitted the
// bridge was scenery you fell through.
var objColFile = map[string]string{}

// kclPaths maps every .kcl stem in the cartridge to its file, built once.
var kclPaths map[string]string

func kclIndex(ls *sm64ds.LevelSet) map[string]string {
	if kclPaths != nil {
		return kclPaths
	}
	kclPaths = map[string]string{}
	for i := 0; i < 2058; i++ {
		if n := ls.InternalName(i); strings.HasSuffix(n, ".kcl") {
			stem := strings.TrimSuffix(filepath.Base(n), ".kcl")
			if _, dup := kclPaths[stem]; !dup {
				kclPaths[stem] = n
			}
		}
	}
	return kclPaths
}

// loadKCL parses one of the cartridge's .kcl collision meshes by stem.
func loadKCL(ls *sm64ds.LevelSet, tmp, stem string) (*sm64ds.KCL, error) {
	p, ok := kclIndex(ls)[stem]
	if !ok {
		return nil, fmt.Errorf("no .kcl named %q", stem)
	}
	data, err := os.ReadFile(filepath.Join(tmp, "files", filepath.FromSlash(strings.TrimPrefix(p, "/"))))
	if err != nil {
		return nil, err
	}
	if len(data) > 4 && string(data[:4]) == "LZ77" {
		data = nds.Decompress(data[4:])
	}
	return sm64ds.ParseKCL(data)
}

// groundUnder drops a ray from (x, y, z) — world units — onto the level's own
// collision and returns the floor height in world units.
//
// This is the game's own ground query: RaycastDown reimplements the walker at
// $01FFD3F8 and is verified against it ray for ray (Part VI §5). A nil CLPS
// makes the surface filter permissive, which is what a camera wants — any floor
// will do. The signpost's init does the same thing to stand itself on the
// ground (Part V §6).
func groundUnder(k *sm64ds.KCL, x, y, z float64) (float64, bool) {
	if k == nil {
		return 0, false
	}
	const floorBelow = -0x80000000
	h, ok := k.RaycastDown(int32(x*4096), int32(y*4096), int32(z*4096), floorBelow, nil, 1)
	if !ok {
		return 0, false
	}
	return float64(h.Y) / 4096, true
}

// writeKCL converts one .kcl to a GLB under dir, named file. Shared by the
// stage meshes and the object colliders, which differ only in where they land.
func writeKCL(ctx *cli.Context, ls *sm64ds.LevelSet, tmp, stem, dir, file string) error {
	p, ok := kclIndex(ls)[stem]
	if !ok {
		return fmt.Errorf("no .kcl named %q", stem)
	}
	data, err := os.ReadFile(filepath.Join(tmp, "files", filepath.FromSlash(strings.TrimPrefix(p, "/"))))
	if err != nil {
		return err
	}
	if len(data) > 4 && string(data[:4]) == "LZ77" {
		data = nds.Decompress(data[4:])
	}
	k, err := sm64ds.ParseKCL(data)
	if err != nil {
		return err
	}
	tris, _ := trisOf(k)
	if len(tris) == 0 {
		return fmt.Errorf("%s: no prisms", stem)
	}
	glbData, err := nitro.ExportTrisGLB(stem+"_col", map[int][]nitro.Tri{0: tris},
		[]nitro.Material{{Name: "collision", Alpha: 31}}, nil)
	if err != nil {
		return err
	}
	gp, err := ctx.Builder.Path(dir, file)
	if err != nil {
		return err
	}
	return os.WriteFile(gp, glbData, 0o644)
}

// objCollider exports a placed actor's collider on demand and returns its
// document-relative path, or "" if the cartridge has no such .kcl.
func objCollider(ctx *cli.Context, ls *sm64ds.LevelSet, tmp, stem string) string {
	if f, done := objColFile[stem]; done {
		return f
	}
	// Beside the level documents, not in a collision/ of their own: a Retro-X
	// file reference is resolved against the document that carries it and may
	// not contain "..", so anything a level doc names has to live in levels/.
	file := "objcol_" + stem + ".glb"
	if err := writeKCL(ctx, ls, tmp, stem, "levels", file); err != nil {
		objColFile[stem] = "" // no .kcl of that name; the placement keeps its prop
		return ""
	}
	objColFile[stem] = file
	return file
}

func exportCollision(ctx *cli.Context, ls *sm64ds.LevelSet, tmp string) error {
	kclIndex(ls)

	done := map[string]bool{}
	levels := 0
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.KCLPath), ".kcl")
		if stem == "" || done[stem] {
			continue
		}
		done[stem] = true
		file := "col_" + stem + ".glb"
		if err := writeKCL(ctx, ls, tmp, stem, "levels", file); err != nil {
			continue
		}
		colFile[stem] = file
		levels++
	}
	ctx.Logf("%d stage collision meshes", levels)
	return nil
}

// logObjColliders reports the object colliders exported alongside the stages.
func logObjColliders(ctx *cli.Context) {
	n := 0
	for _, f := range objColFile {
		if f != "" {
			n++
		}
	}
	ctx.Logf("%d object colliders (placed actors carry their own walkable surfaces)", n)
}

// colMatrix is a placement's collider as one 3x4 local->world, row-major, which
// is what Retro-X asks for (RETROX.md 5.5). Two transforms compose into it:
//
//	the PLACEMENT's own position and yaw, and
//	the collider's captured MtxFx43 (oracle, actor +$134) for the Mbg classes.
//
// The second is not decoration. b_si_so — the see-saw bridge — carries identity,
// but pile and obj_tatefuda carry a uniform 0.0999, without which a stake comes
// out 22 m tall instead of 2.2.
//
// The DS stores a MtxFx43 as four 3-vectors: three basis ROWS (it does v*M) then
// the translation. three.js multiplies the other way round, so the basis is
// transposed into columns here. Every collider this cartridge places is diagonal
// — verified across all eleven Bob-omb Battlefield uses — so the transpose is
// untested by anything that would notice; a rotated collider would be the first.
func colMatrix(pl schema.Placement, c *sm64ds.Collider) []float64 {
	// The collider's own basis (fx12) and offset (fx12, then into stage units).
	var m [3][3]float64
	for r := 0; r < 3; r++ {
		for k := 0; k < 3; k++ {
			m[r][k] = float64(c.Mtx[k*3+r]) / 4096 // transposed: rows -> columns
		}
	}
	t := [3]float64{
		float64(c.Mtx[9]) / 4096 * toStage,
		float64(c.Mtx[10]) / 4096 * toStage,
		float64(c.Mtx[11]) / 4096 * toStage,
	}
	if c.Class == "Kc" {
		// Plain colliders have no transform of their own: they sit wherever the
		// actor does.
		m = [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
		t = [3]float64{}
	}
	if c.ScaleY != 0 && c.ScaleY != 0x1000 {
		sy := float64(c.ScaleY) / 4096
		for r := 0; r < 3; r++ {
			m[r][1] *= sy
		}
	}
	// Then the placement's yaw and position, applied on the left.
	yaw := 0.0
	if len(pl.Rot) == 3 {
		yaw = pl.Rot[1]
	}
	cs, sn := math.Cos(yaw), math.Sin(yaw)
	ry := [3][3]float64{{cs, 0, sn}, {0, 1, 0}, {-sn, 0, cs}}
	out := make([]float64, 12)
	for r := 0; r < 3; r++ {
		for k := 0; k < 3; k++ {
			v := 0.0
			for i := 0; i < 3; i++ {
				v += ry[r][i] * m[i][k]
			}
			out[r*4+k] = r3(v)
		}
		v := 0.0
		for i := 0; i < 3; i++ {
			v += ry[r][i] * t[i]
		}
		out[r*4+3] = r3(v + pl.Pos[r])
	}
	return out
}

func trisOf(k *sm64ds.KCL) (tris []nitro.Tri, skipped int) {
	const lx, ly, lz = 0.30, 0.90, 0.30
	for i := 1; i < k.NumPrisms(); i++ {
		c, ok := k.Corners(i)
		if !ok {
			skipped++
			continue
		}
		n := k.NormalAt(k.PrismAt(i).FaceNormal)
		nl := math.Sqrt(float64(n[0])*float64(n[0]) + float64(n[1])*float64(n[1]) + float64(n[2])*float64(n[2]))
		shade := 0.55
		if nl > 0 {
			d := (float64(n[0])*lx + float64(n[1])*ly + float64(n[2])*lz) / nl
			if d > 0 {
				shade += 0.45 * d
			}
		}
		col := color.NRGBA{R: uint8(210 * shade), G: uint8(38 * shade), B: uint8(38 * shade), A: 255}
		var t nitro.Tri
		bad := false
		for j := 0; j < 3; j++ {
			x, y, z := c[j][0]*toGLB, c[j][1]*toGLB, c[j][2]*toGLB
			if math.Abs(x) > 1e4 || math.Abs(y) > 1e4 || math.Abs(z) > 1e4 {
				bad = true
				break
			}
			t.V[j] = nitro.Vertex{X: x, Y: y, Z: z, C: col}
		}
		if bad {
			skipped++
			continue
		}
		tris = append(tris, t)
	}
	return tris, skipped
}

// ---------------------------------------------------------------------------
// levels
// ---------------------------------------------------------------------------

func exportLevels(ctx *cli.Context, ls *sm64ds.LevelSet, tmp string, bindings map[int][]Binding) error {
	b := ctx.Builder

	modelFor := func(actor int, par [3]int) string {
		var loose string
		for _, bd := range bindings[actor] {
			if len(bd.Models) == 0 {
				continue
			}
			if bd.Params == par {
				return bd.Models[0]
			}
			if loose == "" && bd.Params[0] == par[0] {
				loose = bd.Models[0]
			}
		}
		return loose
	}
	colFor := func(actor int, par [3]int) *sm64ds.Collider {
		var loose *sm64ds.Collider
		for i := range bindings[actor] {
			bd := &bindings[actor][i]
			if len(bd.Colliders) == 0 || bd.Colliders[0].KCL == "" {
				continue
			}
			if bd.Params == par {
				return &bd.Colliders[0]
			}
			if loose == nil && bd.Params[0] == par[0] {
				loose = &bd.Colliders[0]
			}
		}
		return loose
	}

	msgs, err := sm64ds.LoadBMG(filepath.Join(tmp, "files/data/message/msg_data_eng.bin"))
	if err != nil {
		return err
	}
	// readText resolves a placement's own words: par1 is an EXTERNAL message id
	// for the actors in msgActors, translated to an INF1 index by the game's own
	// range table (sm64ds.MsgIndex). Returns the card title and the text.
	readText := func(o sm64ds.LevelObject) (string, string) {
		title, ok := msgActors[o.Actor]
		if !ok {
			return "", ""
		}
		if idx := ls.MsgIndex(o.Params[0]); idx >= 0 && idx < len(msgs) {
			return title, msgs[idx]
		}
		return "", ""
	}

	stageN, totalDropped := 0, 0
	seenStage := map[string]bool{}
	skyCopied := map[string]bool{}
	// refs must stay inside the doc's directory (no ".." per RETROX.md), so a
	// used skybox gets a copy next to the level documents.
	copySky := func(sky string) bool {
		if skyCopied[sky] {
			return true
		}
		src, err := b.Path("objects", sky+".glb")
		if err != nil {
			return false
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return false
		}
		dst, err := b.Path("levels", sky+".glb")
		if err != nil {
			return false
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return false
		}
		skyCopied[sky] = true
		return true
	}
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		stageAsset, ok := refs[stem]
		if !ok || seenStage[stem] {
			continue
		}
		seenStage[stem] = true

		name := levelStems[stem]
		if name == "" {
			name = stem
		}
		doc := &schema.Level{
			Type:  schema.LevelScene3D,
			Scene: &schema.Scene{},
			// The stage model at the origin.
			Placements: []schema.Placement{{
				ID: 0, Object: stageAsset, Pos: []float64{0, 0, 0}, Name: name,
			}},
		}
		if sky := strings.TrimSuffix(filepath.Base(lv.SkyPath), ".bmd"); lv.SkyPath != "" && refs[sky] != "" && copySky(sky) {
			doc.Scene.Layers = append(doc.Scene.Layers, schema.Layer{
				ID: "sky", Name: "Skybox", File: sky + ".glb",
				Mode: "toggle", Attach: "camera", Role: "sky", RenderOrder: -1,
			})
		}
		// A placement group for the things the game does not show you while you
		// play: see markerModels.
		markerLayer := false
		if kcl := strings.TrimSuffix(filepath.Base(lv.KCLPath), ".kcl"); lv.KCLPath != "" && colFile[kcl] != "" {
			off := false
			doc.Scene.Layers = append(doc.Scene.Layers, schema.Layer{
				ID: "collision", Name: "Collision (KCL)", File: colFile[kcl],
				Mode: "toggle", Visible: &off, Role: "collision",
			})
		}

		// Camera: an establishing shot over the level's own extent — the
		// bounds of its entrances and placed actors (the stage GLB's bounds
		// aren't decoded here, but the actors trace its playable footprint).
		var lo, hi [3]float64
		haveB := false
		note := func(x, y, z float64) {
			v := [3]float64{x, y, z}
			if !haveB {
				lo, hi, haveB = v, v, true
				return
			}
			for k := 0; k < 3; k++ {
				lo[k] = math.Min(lo[k], v[k])
				hi[k] = math.Max(hi[k], v[k])
			}
		}
		for _, e := range lv.Entrances {
			note(e.X*toStage, e.Y*toStage, e.Z*toStage)
		}
		for _, o := range lv.Objects {
			note(o.X*toStage, o.Y*toStage, o.Z*toStage)
		}
		cam := &schema.Camera{Mode: "fly", FOV: 50, Near: 0.01, Far: 500, Fly: &schema.Fly{Speed: 3}}
		span := math.Max(1, math.Max(hi[0]-lo[0], hi[2]-lo[2]))
		if haveB {
			cam.Fly.Speed = math.Max(1, math.Min(10, span/6))
		}
		var kcl *sm64ds.KCL
		if lv.KCLPath != "" {
			kcl, _ = loadKCL(ls, tmp, strings.TrimSuffix(filepath.Base(lv.KCLPath), ".kcl"))
		}
		floorY := math.Inf(-1)
		if haveB {
			floorY = lo[1]
		}
		if e, ok := spawnShot(lv, kcl, floorY); ok {
			cam.Pos, cam.Target = e.pos, e.target
		} else if haveB {
			cx, cz := (lo[0]+hi[0])/2, (lo[2]+hi[2])/2
			cam.Pos = []float64{cx, hi[1] + span*0.5, lo[2] - span*0.7}
			cam.Target = []float64{cx, (lo[1] + hi[1]) / 2, cz}
		} else {
			cam.Pos = []float64{0, 8, -16}
			cam.Target = []float64{0, 0, 0}
		}
		doc.Camera = cam

		// Placed actors, oracle-bound.
		//
		// An object that belongs to several missions is listed once PER STAR
		// LAYER in the level's object table (the walker $020FE33C skips whole
		// entries whose layer isn't the star being played), so the records
		// collapse by (actor, position) into ONE placement carrying the SET of
		// layers it was listed under — Bob-omb Battlefield's chain chomp sits
		// in stars 1 and 7, and keeping only the first record would have made
		// it a star-1 object.
		type placedObj struct {
			o      sm64ds.LevelObject
			layers map[int]bool
			gated  map[int]bool // nil = the actor never refused; else the missions it exists in
		}
		byKey := map[string]*placedObj{}
		var order []*placedObj
		for _, o := range lv.Objects {
			key := fmt.Sprintf("%d/%.3f/%.3f/%.3f", o.Actor, o.X, o.Y, o.Z)
			p := byKey[key]
			if p == nil {
				p = &placedObj{o: o, layers: map[int]bool{}}
				byKey[key] = p
				order = append(order, p)
			}
			p.layers[o.Layer] = true
		}
		// A second mechanism decides mission membership on top of the layer:
		// fifty placed actors gate themselves on the save's star bits, which is
		// how Whomp's Fortress shows the Whomp King on mission 1 and the tower
		// you climb on the rest (Part V §8). The oracle ran their real
		// create/init under each mission; fold its answer in.
		for _, g := range missionGates[stem] {
			p := byKey[fmt.Sprintf("%d/%.3f/%.3f/%.3f", g.Actor, float64(g.X), float64(g.Y), float64(g.Z))]
			if p == nil {
				continue
			}
			p.gated = map[int]bool{}
			for _, m := range g.Missions {
				p.gated[m] = true
			}
		}
		used := map[int]bool{}
		for _, p := range order {
			for l := range p.layers {
				if l != 0 {
					used[l] = true
				}
			}
			if p.gated != nil {
				for m := range p.gated {
					used[m] = true
				}
			}
		}
		doc.Variants = starVariants(msgs, ls.Course(i), used)
		declared := map[string]bool{}
		for _, v := range doc.Variants {
			declared[v.ID] = true
		}
		// curVars is the variant membership of the placement being emitted —
		// nil for a layer-0 object, which is present in every mission. It is a
		// closure variable so the chain chomp's spawned stake and drawn links
		// inherit the missions of the chomp that owns them.
		var curVars []string
		// curScale overrides the standard object scale for the one actor that
		// sizes its own geometry — see paintingScale.
		var curScale *schema.Scale
		pid := 1
		dropped := map[int]int{} // actor -> placements this level could not emit
		addObjOff := func(o sm64ds.LevelObject, actor int, par [3]int, rot bool, off [3]float64) {
			m := modelFor(actor, par)
			asset, ok := refs[m]
			if m == "" || !ok {
				// The actor oracle recorded no model for this actor, so there
				// is nothing to place. Silently dropping it is how Lethal Lava
				// Land lost its rolling log (daObjFlMaruta_c, actor 70) without
				// anyone noticing, so say so.
				dropped[actor]++
				return
			}
			sc := schema.Scale{objScale}
			if curScale != nil {
				sc = *curScale
			}
			pl := schema.Placement{
				ID:     pid,
				Object: asset,
				Pos:    []float64{r3(o.X*toStage + off[0]), r3(o.Y*toStage + off[1]), r3(o.Z*toStage + off[2])},
				Scale:  sc,
				Props:  map[string]any{"actor": actor},
			}
			if rot && o.RotY != 0 {
				pl.Rot = []float64{0, o.RotY * math.Pi / 180, 0}
			}
			pl.Variants = curVars
			if markerModels[m] {
				pl.Layer = markerLayerID
				markerLayer = true
			}
			if coinActors[actor] {
				pl.Behavior = &schema.Behavior{
					Kind: "spin", Axis: []float64{0, 1, 0}, Rate: r3(coinSpinRate),
				}
			}
			if title, t := readText(o); t != "" {
				pl.OnClick = &schema.OnClick{Action: schema.ActionText, Title: title, Body: t}
			}
			if c := colFor(actor, par); c != nil {
				pl.Props["collider"] = c.KCL
				if f := objCollider(ctx, ls, tmp, c.KCL); f != "" {
					pl.Collision = &schema.ObjCollision{File: f, Matrix: colMatrix(pl, c)}
				}
			}
			doc.Placements = append(doc.Placements, pl)
			pid++
		}
		addObj := func(o sm64ds.LevelObject, actor int, par [3]int, rot bool) {
			addObjOff(o, actor, par, rot, [3]float64{})
		}
		// addModelAt places a bare archive model — no actor, no binding. Only the
		// chain chomp's chain needs it: the links are drawn by the chomp's own
		// draw function rather than spawned as actors, so nothing places them.
		addModelAt := func(stem string, pos [3]float64, name string) {
			asset, ok := refs[stem]
			if !ok {
				return
			}
			doc.Placements = append(doc.Placements, schema.Placement{
				ID: pid, Object: asset, Name: name,
				Pos:      []float64{r3(pos[0]), r3(pos[1]), r3(pos[2])},
				Scale:    schema.Scale{objScale},
				Variants: curVars,
			})
			pid++
		}
		for _, p := range order {
			o := p.o
			curVars = variantIDs(p.layers, p.gated, declared)
			curScale = nil
			switch o.Actor {
			case paintingActor:
				sc, lift := paintingScale(o.Params[0])
				curScale = &sc
				addObjOff(o, o.Actor, o.Params, true, [3]float64{0, lift, 0})
			case 219: // daWanwan_c — chained to a stake
				addChomp(o, addObjOff, addModelAt)
			case 337: // daWanwan2_c — the free-roaming one; no stake, same ball
				addObjOff(o, o.Actor, o.Params, true,
					[3]float64{0, -modelFloor["ar1_2"] * objScale, 0})
			default:
				addObj(o, o.Actor, o.Params, true)
			}
		}

		if markerLayer {
			off := false
			doc.Scene.Layers = append(doc.Scene.Layers, schema.Layer{
				ID: markerLayerID, Name: "Stars, markers & cannons",
				Mode: "toggle", Visible: &off,
			})
		}
		b.AddLevel(schema.Asset{
			ID: slugify(strings.TrimSuffix(stem, "_all")), Name: name, Group: "Levels",
		}, doc)
		stageN++
		msg := fmt.Sprintf("%-24s %d placements", stem, len(doc.Placements))
		if n := 0; len(dropped) > 0 {
			var actors []int
			for a, c := range dropped {
				actors = append(actors, a)
				n += c
			}
			sort.Ints(actors)
			msg += fmt.Sprintf("  (%d unmodelled, actors %v)", n, actors)
			totalDropped += n
		}
		ctx.Progress("levels", stageN, 0, msg)
	}
	if totalDropped > 0 {
		ctx.Logf("levels: %d placements had no oracle-bound model and were not emitted "+
			"(run `gateprobe -dropped` for the actor breakdown)", totalDropped)
	}
	return nil
}

// ---------------------------------------------------------------------------
// star (mission) variants
// ---------------------------------------------------------------------------

// starVariants declares one Retro-X variant per mission of a level, from the
// set of non-zero star layers its object table actually uses.
//
// A numbered painting course has StarsPerCourse missions and the ROM names all
// of them, so every one is declared even when a mission adds no object of its
// own: picking it then shows the level with only the always-present objects,
// which is exactly what the game does. Everything else — the castle floors,
// the boss arenas, the test maps — has layered objects but no star-name block,
// so only the layers that appear are declared, under their bare layer number.
//
// Star 1 is the default: it is the mission the game starts you on.
func starVariants(msgs []string, course int, used map[int]bool) []schema.Variant {
	if len(used) == 0 {
		return nil
	}
	stars := make([]int, 0, sm64ds.StarsPerCourse)
	if course >= 0 && course < sm64ds.StarNameCourses {
		for s := 1; s <= sm64ds.StarsPerCourse; s++ {
			stars = append(stars, s)
		}
	} else {
		for s := range used {
			stars = append(stars, s)
		}
		sort.Ints(stars)
	}
	if len(stars) < 2 {
		return nil // a one-entry picker is no picker
	}
	out := make([]schema.Variant, 0, len(stars))
	for _, s := range stars {
		name := starTitle(sm64ds.StarName(msgs, course, s))
		if name == "" {
			name = fmt.Sprintf("Star %d", s)
		}
		out = append(out, schema.Variant{ID: starID(s), Name: name, Default: len(out) == 0})
	}
	return out
}

func starID(star int) string { return fmt.Sprintf("star%d", star) }

// variantIDs turns one placement's star-layer set, intersected with whatever
// its own code decided (the mission gate), into its variant membership. Layer 0
// with no gate means "every mission", which Retro-X spells as no `variants`
// field at all, so it returns nil — as it does for a level with no variants.
func variantIDs(layers, gated map[int]bool, declared map[string]bool) []string {
	if len(declared) == 0 {
		return nil
	}
	if layers[0] && gated == nil {
		return nil
	}
	var stars []int
	for l := 1; l <= sm64ds.StarsPerCourse; l++ {
		if !declared[starID(l)] {
			continue
		}
		// the object-table layer and the actor's own save gate must BOTH admit
		// the mission (layer 0 means every mission, and no gate means the same)
		if !layers[0] && !layers[l] {
			continue
		}
		if gated != nil && !gated[l] {
			continue
		}
		stars = append(stars, l)
	}
	if len(stars) == 0 || len(stars) == len(declared) {
		return nil // in every declared mission: same thing as unrestricted
	}
	out := make([]string, len(stars))
	for i, s := range stars {
		out[i] = starID(s)
	}
	return out
}

// markerModels are the models the game does not show you while you are playing
// the level, and which read as clutter in a diorama: the Power Star and its
// spawn marker, the mission-number glyph that goes with them, and the cannons
// (which sit under a closed lid until you open one). They are in the level
// data and they stay in the document — the exporter puts them in a placement
// group the viewer starts with switched off, so the default look is the level
// as you play it and one checkbox brings them back.
//
// Keyed by MODEL, not actor: the model is what a viewer sees and names.
var markerModels = map[string]bool{
	"arc0_21":   true, // the Power Star
	"star_base": true, // the flat star silhouette marking where one spawns
	"arc0_9":    true, // the mission-number glyph
	"houdai":    true, // a cannon
}

const markerLayerID = "markers"

// msgActors are the placed actors whose par1 is an EXTERNAL message id, with
// the title their card gets. Established by resolving par1 through the game's
// own id->index range table (`gateprobe -actormsgs N`) for EVERY placement of
// the actor in the game, and keeping only the actors where ALL of them land on
// real text:
//
//	184 obj_tatefuda  102 of 103 (the holdout is on the unused test map, par1 $FFFF)
//	183 obj_kanban      9 of 9,  ids 1802-1809 — one contiguous block
//	185 kinopio        10 of 10, ids in the 2800 block
//
// The control that makes those numbers mean something is actor 187 (mip, the
// rabbit), where only 9 of 29 resolve: par1 is plainly not a message id there,
// and a laxer test would have taken it.
//
// Honest limit: only the SIGNPOST's presentation is traced end to end (init
// builds a dialog at +$59C, the step watches the interaction flag at +$B0 and
// opens the window through $020BB060 — Part V §6). The notice board and Toad
// register an interaction volume and hand off to overlay 7's dialog module,
// which is where every caller of the id translator $020B8EC0 lives; that path
// is not traced. What IS decoded is which words belong to which object.
var msgActors = map[int]string{
	184: "Signpost",
	183: "Notice board",
	185: "Toad",
}

// shot is one camera placement in stage-GLB units.
type shot struct{ pos, target []float64 }

// spawnShot puts the opening camera where the game starts you: behind the
// player's spawn, looking the way they face.
//
// The spawn is decoded, not chosen. Object-table type 1 is an entrance record
// (handler $020FE6C8, Part V §2) and the level's FIRST one is where the game
// stands the player; all 48 shipped stages have one, with a heading. Bob-omb
// Battlefield's reads (-6225, 1700, 6353) yaw 135 — the same value the world-mode
// preset derived independently.
//
// The OFFSET behind it is editorial, and deliberately not the game's: SM64DS's
// own camera sits a couple of metres off Mario's shoulder, which frames a
// diorama badly. These numbers are a third-person shot pulled back far enough to
// show where you are — about 4 and 2 Mario-heights (idle Mario is 0.1415 stage
// units, the ruler the world-mode preset established). What IS taken from the
// cartridge is the position and the heading, so every level now opens facing the
// way the game faces you.
func spawnShot(lv *sm64ds.Level, kcl *sm64ds.KCL, contentFloor float64) (shot, bool) {
	if len(lv.Entrances) == 0 {
		return shot{}, false
	}
	e := lv.Entrances[0]
	// marioH is idle Mario in stage-GLB units — the ruler the world-mode preset
	// established by measuring the POSED model (site/vr/super-mario-64-ds/…).
	const (
		marioH = 0.1415
		back   = 4.0 * marioH // behind the spawn, along its own heading
		eye    = 1.6 * marioH // eye height above the floor
		aim    = 1.0 * marioH // look at about where the player's head would be
	)
	// The heading convention is the one every placement already uses: yaw is
	// degrees about +Y, forward = (sin, 0, cos).
	yaw := e.RotY * math.Pi / 180
	fx, fz := math.Sin(yaw), math.Cos(yaw)
	x, z := e.X*toStage, e.Z*toStage
	cx, cz := x-fx*back, z-fz*back
	// The collision query is in WORLD units, the camera in stage-GLB units.
	const toWorld = 1 / toStage

	// THE SPAWN IS IN THE SKY. An entrance record is where the game RELEASES
	// the player, not where they land — Bob-omb Battlefield's y=1.7 is a drop
	// height, which the world-mode preset also refuses to stand on, and Whomp's
	// Fortress drops you high enough that a shot from there sees nothing but
	// skybox. So drop a ray onto the level's own collision and stand the camera
	// on the floor, the way the game stands the player on it.
	gy, ok := groundUnder(kcl, e.X, e.Y, e.Z)
	if !ok {
		gy = e.Y // no collision under the spawn: keep the drop height
	}
	// The camera sits BEHIND the spawn, which may be over higher ground (or a
	// wall); take whichever floor is higher so the shot never sinks into it.
	if cg, ok := groundUnder(kcl, cx*toWorld, e.Y, cz*toWorld); ok && cg > gy {
		gy = cg
	}
	gy *= toStage
	// Two levels drop you over a floor that is below EVERYTHING: the wing-cap
	// course is rings of items in open sky, and standing on its distant ground
	// looks at nothing. Never go below the lowest thing the level places, and
	// never above the drop point itself.
	if lowest := math.Min(e.Y*toStage, contentFloor); gy < lowest {
		gy = lowest
	}
	return shot{
		pos:    []float64{r3(cx), r3(gy + eye), r3(cz)},
		target: []float64{r3(x), r3(gy + aim), r3(z)},
	}, true
}

// paintingActor 307 is every framed painting in the castle, and it SIZES ITSELF:
// the shipped model is a bare 12.5-unit square quad carrying the picture, and
// the actor's init ($02126CA0, overlay 80) reads the spawn parameter word and
// builds its own subdivided grid from it —
//
//	par1 & $F         width  = (n+1) * 100.0 world units
//	(par1 >> 4) & $F  height = (n+1) * 100.0
//	(par1 >> 8) & $1F which picture ($02125630 resolves the texture)
//	(par1 >> 13) & 3  behaviour mode; >= 2 collapses the grid to 2x2 (a flat
//	                  wall picture rather than an enterable, rippling one)
//
// The sizes are at $02126E80: two nibbles, each `(n+1) * $64000`, halved into
// the interaction volume the init registers. Exporting the quad at the standard
// object scale instead gave every painting the same 0.1 stage units — six times
// too small for the 600-unit ones, sixteen times for the 1600-unit one.
//
// NOT reproduced: the ripple. The grid exists so the painting can be displaced
// by a wave when you touch it, and the mesh generator lives behind a mode record
// this does not decode. The Studio ships the painting flat, at its real size.
const paintingActor = 307

// paintingScale returns the placement scale that makes the 12.5-unit quad span
// the painting's authored size, and the vertical lift that puts its BOTTOM edge
// on the placement point — the quad is centred on its origin and the placement
// sits at the foot of the frame.
func paintingScale(par1 int) (schema.Scale, float64) {
	const quad = 12.5 // every for_*.glb is exactly this, flat in XY at z=0
	w := float64((par1&0xF)+1) * 100 * toStage
	h := float64((par1>>4&0xF)+1) * 100 * toStage
	return schema.Scale{r3(w / quad), r3(h / quad), objScale}, r3(h / 2)
}

// coinActors are the three placed-coin classes. All three profiles
// ($02108790/AC/C8, actors 288/289/290) install the same vtable $021087EC,
// whose step slot (+$18) is the coin step $020B2324 — `ADD r1, r1, #0xC00 /
// STRH r1, [r2]` on the yaw short at +$8E, once per 30 Hz actor tick. A sweep
// of every placed actor's step for that constant returns these three and
// nothing else (extract/cmd/starprobe).
var coinActors = map[int]bool{288: true, 289: true, 290: true}

// coinSpinRate is that $C00-per-tick yaw in rad/s — $10000 angle units is a
// full turn, the actor tick is 30 Hz: ~1.4 revolutions per second.
const coinSpinRate = float64(0xC00) / 0x10000 * 30 * 2 * math.Pi

func r3(v float64) float64 { return float64(int(v*1000+0.5*sign(v))) / 1000 }
func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

// yawOnlyBillboards is an EDITORIAL list, not a decode.
//
// The cartridge marks camera-facing bones with one flag and one flag only (see
// the census in sm64ds/skin.go): it cannot tell a yaw billboard from a fully
// camera-aligned one, and neither can the geometry — bomb_tree's quad is
// 1.22:1, near-square, like the balls. On hardware the distinction is close to
// unobservable, because SM64DS's camera stays roughly horizontal and there the
// two modes look identical.
//
// A Studio diorama gets looked at from above, where they do not. These models
// read as upright standing props and keep their feet on the ground; everything
// else marked is a ball, a star or a character's body, and faces the viewer
// outright.
var yawOnlyBillboards = map[string]bool{
	"bomb_tree": true, "toge_tree": true, "yashi_tree": true, "yuki_tree": true,
	"bk_billbord": true, // a signboard on a post
}

// billboardMode reports the Retro-X billboard mode for a model. The DS engine
// substitutes the camera's rotation when it composes a bone whose flag word at
// +$3C has bit 0 set (sm64ds/bmd.go), so ANY model carrying such a bone is
// billboarded: a single-bone sprite turns whole, a character turns only the
// marked part (the bob-omb's body_bill, King Bob-omb's body). The viewer picks
// those parts out of the GLB, which carries the flag per node in
// extras.billboard; this doc-level value tells it which way to turn them.
func billboardMode(m *sm64ds.Model) string {
	bill := false
	for _, j := range m.Skel {
		if j.Billboard {
			bill = true
			break
		}
	}
	if !bill {
		return ""
	}
	if yawOnlyBillboards[m.Name] {
		return "yaw"
	}
	return "camera"
}
