package main

// intro.go — the opening cutscene as ONE Retro-X script. The demo player
// streams seven shots (opwf → oppm → opeg → opcn → opsu → opdn → opod), each
// a RARC archive in /Ajioka/ADemo/. The member roles are DECLARED below,
// read straight off the archives' own listings: the sets are themselves
// animated (op_mansion.key sways the trees and opod_mansion.key swings the
// front door), so sets and characters alike are script actors — the level is
// placements-only. Luigi at the gate, on the steps and at the door is ONE
// model (entergate.mdl) driven by per-shot .key clips, exactly as stored.
//
// The demo world is y-up and the .scd camera tracks live in it directly, so
// actors export in their keys' own space — plus the demo player's per-frame
// placement A(f) (blocking/<shot>.json, captured by extract/cmd/actorsolve
// from the running game's world-matrix arrays), baked by SkinnedGLB as
// animation channels on each actor's wrapper node. The world-space actors
// (the sets, torch, lightning, crow) solve to identity and ship untouched.
//
// Known gaps, declared: the .slk blend-shape weight tracks are not exported
// (the .sls rest face is applied); the .bas sound cues await their decoder,
// so shots carry no sounds[]; the .clr material-colour tracks and .txp
// texture flipbooks are undecoded, so translucent tints hold their rest
// values (the lightning rests invisible, the torch flame doesn't dance).

import (
	"embed"
	"fmt"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/export"
	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

//go:embed blocking
var blockingFS embed.FS

// shotBlocking loads a shot's captured actor-placement table, nil when the
// shot has none.
func shotBlocking(shotID string) *export.BlockingTable {
	b, err := blockingFS.ReadFile("blocking/" + shotID + ".json")
	if err != nil {
		return nil
	}
	t, err := export.ParseBlocking(b)
	if err != nil {
		panic(fmt.Sprintf("blocking/%s.json: %v", shotID, err))
	}
	return t
}

// bind names one actor: the object asset it becomes, and the archive members
// that build it. A binding with no key exports static.
type bind struct {
	asset, name, mdl, key string
}

// shotSpec is one shot: its archive and its actors, sets included, as
// (mdl, key) pairs from that archive.
type shotSpec struct {
	id, name string
	actors   []bind
}

// cutsceneSpec is one cutscene level: a script of shots. The opening is the
// seven-shot attract film; the lab farewell and the thunderbolt are
// single-shot demos from the same /Ajioka/ADemo/ library.
type cutsceneSpec struct {
	levelID, levelName   string
	scriptID, scriptName string
	scriptDesc           string
	group                string // object group for the actors
	flySpeed             float64
	shots                []shotSpec
}

// introShots is the demo player's own shot order.
var introShots = []shotSpec{
	{"opwf", "The forest walk", []bind{
		{"forest", "The dark forest", "opwf_bg.mdl", "opwf_bg.key"},
		{"luigi-walk", "Luigi — the forest walk", "opwf_luigi.mdl", "opwf_luigi.key"},
		{"cone-walk", "Flashlight cone (forest)", "opwf_cone.mdl", "opwf_cone.key"},
		{"handlight-walk", "Luigi's flashlight (forest)", "opwf_handlight.mdl", "opwf_handlight.key"},
	}},
	{"oppm", "The map points the way", []bind{
		{"mansion-set", "The mansion approach", "op_mansion.mdl", "op_mansion.key"},
		{"bighand", "The pointing hand", "op_bighand.mdl", "oppm_hand.key"},
		{"lightning", "Lightning bolt", "oppm_lightning.mdl", "oppm_lightning.key"},
		{"gate-torch", "Gate torch", "torch.mdl", "torch.key"},
	}},
	{"opeg", "At the gate", []bind{
		{"mansion-set-gate", "The mansion — the gate opens", "op_mansion.mdl", "opeg_mansion.key"},
		{"luigi-gate", "Luigi — at the gate", "entergate.mdl", "entergate.key"},
		{"cone-gate", "Flashlight cone (gate)", "opeg_cone.mdl", "opeg_cone.key"},
		{"handlight-gate", "Luigi's flashlight (gate)", "opeg_handlight.mdl", "opeg_handlight.key"},
		{"gate-torch", "Gate torch", "torch.mdl", "torch.key"},
	}},
	{"opcn", "The crow watches", []bind{
		{"mansion-set", "The mansion approach", "op_mansion.mdl", "op_mansion.key"},
		{"crow", "The crow", "karasu1.mdl", "karasu1.key"},
	}},
	{"opsu", "Up the steps", []bind{
		{"mansion-set", "The mansion approach", "op_mansion.mdl", "op_mansion.key"},
		{"luigi-stepup", "Luigi — up the steps", "entergate.mdl", "opsu_luigi.key"},
		{"cone-stepup", "Flashlight cone (steps)", "opsu_cone.mdl", "opsu_cone.key"},
		{"handlight-stepup", "Luigi's flashlight (steps)", "opsu_handlight.mdl", "opsu_handlight.key"},
	}},
	{"opdn", "The door knob", []bind{
		{"mansion-set", "The mansion approach", "op_mansion.mdl", "op_mansion.key"},
	}},
	{"opod", "Opening the door", []bind{
		{"mansion-set-door", "The mansion — the door opens", "op_mansion.mdl", "opod_mansion.key"},
		{"luigi-opendoor", "Luigi — opening the door", "entergate.mdl", "opod_luigi.key"},
		{"cone-door", "Flashlight cone (door)", "opod_cone.mdl", "opod_cone.key"},
		{"handlight-door", "Luigi's flashlight (door)", "opod_handlight.mdl", "opod_handlight.key"},
	}},
}

var cutscenes = []cutsceneSpec{
	{
		levelID: "intro", levelName: "The opening",
		scriptID: "opening", scriptName: "The opening cutscene",
		scriptDesc: "The seven demo shots the game streams before the title: Luigi walks the forest path, " +
			"the map points the way, the gate and the front door open. Sets, actor clips and camera cuts " +
			"are the game's own data; sound cues await the .bas decoder.",
		group: "Intro", flySpeed: 600,
		shots: introShots,
	},
	{
		levelID: "lab-byebye", levelName: "The lab — bye-bye",
		scriptID: "byebye", scriptName: "Bye-bye (drbyebye)",
		scriptDesc: "The Professor E. Gadd lab farewell demo (dodb.szp, the archive's own name is drbyebye): " +
			"Gadd, Luigi with the flashlight and the Poltergust in the lab set, played by the same demo " +
			"machinery as the opening. Blend-shape weight tracks (.slk) and texture flipbooks (.txp) are " +
			"not yet decoded, so faces hold their rest shape.",
		group: "Lab", flySpeed: 600,
		shots: []shotSpec{{"dodb", "Bye-bye", []bind{
			{"lab", "Professor E. Gadd's lab", "db_bg.mdl", "db_bg.key"},
			{"lab-gadd", "Professor E. Gadd", "db_lohakase.mdl", "db_lohakase.key"},
			{"lab-luigi", "Luigi — the lab", "db_luigi.mdl", "db_luigi.key"},
			{"lab-poltergust", "The Poltergust", "db_sojiki.mdl", "db_sojiki.key"},
			{"lab-cone", "Flashlight cone (lab)", "db_cone.mdl", "db_cone.key"},
			{"lab-handlight", "Luigi's flashlight (lab)", "db_handlight.mdl", "db_handlight.key"},
			{"lab-fire", "The lab brazier", "int02fir.mdl", "int02fir.key"},
		}}},
	},
	{
		levelID: "thunderbolt", levelName: "The thunderbolt",
		scriptID: "thunderbolt", scriptName: "The thunderbolt",
		scriptDesc: "The lightning-strike demo (dotb.szp): the same lab set under the storm — the drama is " +
			"in the .scd camera and its keyed lights. The .clr material-colour flash tracks are not yet " +
			"decoded, so the bolt's materials hold their rest colour.",
		group: "Lab", flySpeed: 600,
		shots: []shotSpec{{"dotb", "The thunderbolt", []bind{
			{"lab-storm", "The lab under the storm", "tb_bg.mdl", "tb_bg.key"},
		}}},
	},
}

// demoArcs names each shot's archive; the intro shots follow the op<id>
// convention, the demos are their own files.
func demoArc(shotID string) string {
	return "/Ajioka/ADemo/" + shotID + ".szp"
}

func exportIntro(ctx *cli.Context, src *export.Source, doObjects, doLevels bool) error {
	for _, cs := range cutscenes {
		if err := exportCutscene(ctx, src, cs, doObjects, doLevels); err != nil {
			return err
		}
	}
	return nil
}

func exportCutscene(ctx *cli.Context, src *export.Source, cs cutsceneSpec, doObjects, doLevels bool) error {
	b := ctx.Builder

	doc := &schema.Level{
		Type:  schema.LevelScene3D,
		Scene: &schema.Scene{},
	}
	script := &schema.Script{Name: cs.scriptName, FPS: 30}

	built := map[string]string{}    // asset id → clip id (dedup across shots)
	placementOf := map[string]int{} // asset id → placement id
	nextPlacement := 0
	haveCamera := false

	for si, shot := range cs.shots {
		arcPath := demoArc(shot.id)
		files, err := src.Archive(arcPath)
		if err != nil {
			return fmt.Errorf("intro: %s: %w", arcPath, err)
		}

		// --- camera ---------------------------------------------------------
		scdName, scoName, ok := export.FindShotCamera(files)
		if !ok {
			return fmt.Errorf("intro: %s has no scd/sco pair", shot.id)
		}
		cam, err := export.BakeCamera(files, scdName, scoName)
		if err != nil {
			return fmt.Errorf("%s camera: %w", shot.id, err)
		}
		track := &schema.CameraTrack{
			Frames: cam.Frames, FPS: cam.FPS,
			Near: float64(cam.Near), Far: float64(cam.Far),
		}
		for _, s := range cam.Track {
			track.Track = append(track.Track, schema.CamSample{
				Pos:    []float64{float64(s.Pos[0]), float64(s.Pos[1]), float64(s.Pos[2])},
				Target: []float64{float64(s.Target[0]), float64(s.Target[1]), float64(s.Target[2])},
				Roll:   float64(s.Roll),
				FOV:    float64(s.Fov),
			})
		}
		if doLevels {
			b.AddSideDoc("levels/"+cs.levelID+"/cameras/"+shot.id+".json", track)
		}
		if !haveCamera && len(track.Track) > 0 {
			doc.Camera = &schema.Camera{
				Mode: "fly",
				Pos:  track.Track[0].Pos, Target: track.Track[0].Target,
				FOV: 50, Near: 5, Far: 200000,
				Fly: &schema.Fly{Speed: 600},
			}
			haveCamera = true
		}

		sh := schema.Shot{
			ID: shot.id, Name: shot.name,
			Frames: track.Frames,
			Camera: schema.ShotCamera{
				Near: track.Near, Far: track.Far,
				TrackFile: "cameras/" + shot.id + ".json",
			},
		}

		// --- actors (sets included) -----------------------------------------
		blocking := shotBlocking(shot.id)
		if blocking != nil {
			// Attached props (the flashlight and its cone ride Luigi's hand)
			// expand into per-frame wrapper rows derived from the keys.
			err := blocking.ExpandAttachments(func(spec string) (*lm.MDL, *lm.Key, error) {
				mdl, keyName, ok := strings.Cut(spec, "+")
				if !ok {
					return nil, nil, fmt.Errorf("blocking actor %q: want mdl+key", spec)
				}
				return export.LoadSkinned(files, mdl, keyName)
			})
			if err != nil {
				return fmt.Errorf("%s blocking: %w", shot.id, err)
			}
		}
		for _, a := range shot.actors {
			clipID, seen := built[a.asset]
			if !seen && doObjects {
				glbPath, err := b.Path("objects", a.asset+".glb")
				if err != nil {
					return err
				}
				var metas []schema.Animation
				if a.key != "" && export.Member(files, a.key) != nil {
					m, key, err := export.LoadSkinned(files, a.mdl, a.key)
					if err != nil {
						return fmt.Errorf("%s %s+%s: %w", shot.id, a.mdl, a.key, err)
					}
					clipID = trimExt(a.key)
					var bl *export.Blocking
					if blocking != nil {
						bl = blocking.Actors[a.mdl+"+"+a.key]
					}
					if err := export.SkinnedGLB(m, key, glbPath, a.asset, clipID, false, bl); err != nil {
						return fmt.Errorf("%s %s: %w", shot.id, a.mdl, err)
					}
					metas = []schema.Animation{{
						ID: clipID, Clip: clipID, FPS: 30, Loop: "once",
						Description: "The shot's .key clip: hermite channels on the 30 fps demo timeline, root motion kept.",
					}}
				} else {
					mem := export.Member(files, a.mdl)
					if mem == nil {
						return fmt.Errorf("%s: no member %s", shot.id, a.mdl)
					}
					m, err := lm.ParseMDL(mem.Data)
					if err != nil {
						return fmt.Errorf("%s %s: %w", shot.id, a.mdl, err)
					}
					if err := export.StaticGLB(m, glbPath, a.asset); err != nil {
						return fmt.Errorf("%s %s: %w", shot.id, a.mdl, err)
					}
				}
				b.AddObject(schema.Asset{ID: a.asset, Name: a.name, Group: cs.group}, &schema.Object{
					Type: schema.ObjectModel3D, Name: a.name, Model: a.asset + ".glb",
					SkinnedClone: clipID != "",
					Animations:   metas,
					Props:        map[string]any{"source": arcPath + ":" + a.mdl + "+" + a.key},
				})
				built[a.asset] = clipID
			} else if seen {
				clipID = built[a.asset]
			}
			if doObjects && doLevels {
				pid, ok := placementOf[a.asset]
				if !ok {
					pid = nextPlacement
					nextPlacement++
					placementOf[a.asset] = pid
					doc.Placements = append(doc.Placements, schema.Placement{
						ID:     pid,
						Object: a.asset,
						Pos:    []float64{0, 0, 0},
						Name:   a.name,
						Props: map[string]any{
							"placement": "the sets are modelled in world space; character placement matrices live in the demo player and are not yet traced",
						},
					})
				}
				actor := schema.Actor{Placement: pid}
				if clipID != "" {
					actor.Clip = clipID
				}
				sh.Actors = append(sh.Actors, actor)
			}
		}
		if doLevels {
			script.Shots = append(script.Shots, sh)
		}
		ctx.Progress("levels", si+1, len(cs.shots),
			fmt.Sprintf("%s %s: %d frames, %d actors", cs.levelID, shot.id, sh.Frames, len(sh.Actors)))
	}

	if !doLevels || len(script.Shots) == 0 {
		return nil
	}
	b.AddSideDoc("levels/"+cs.levelID+"/"+cs.scriptID+".json", script)
	doc.Scripts = []schema.ScriptRef{{
		ID: cs.scriptID, Name: cs.scriptName, File: cs.levelID + "/" + cs.scriptID + ".json",
		Description: cs.scriptDesc,
	}}
	b.AddLevel(schema.Asset{ID: cs.levelID, Name: cs.levelName}, doc)
	return nil
}

func trimExt(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return s[:i]
		}
	}
	return s
}
