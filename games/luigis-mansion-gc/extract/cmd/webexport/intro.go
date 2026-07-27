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
// Known gaps, declared: per-shot actor placement matrices live in the demo
// player's code and are not yet traced (everything ships at the origin — the
// sets are modelled in world space so this is exact for them and approximate
// for characters); the .slk blend-shape weight tracks are not exported (the
// .sls rest face is applied); the .bas sound cues await their decoder, so
// shots carry no sounds[].

import (
	"fmt"

	"retroreverse.com/games/luigis-mansion-gc/extract/export"
	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

// bind names one actor: the object asset it becomes, and the archive members
// that build it. A binding with no key exports static.
type bind struct {
	asset, name, mdl, key string
}

// introShots is the demo player's own shot order, with each shot's actors —
// sets included — as (mdl, key) pairs from that shot's archive.
var introShots = []struct {
	id, name string
	actors   []bind
}{
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

func exportIntro(ctx *cli.Context, src *export.Source, doObjects, doLevels bool) error {
	b := ctx.Builder

	doc := &schema.Level{
		Type:  schema.LevelScene3D,
		Scene: &schema.Scene{},
	}
	script := &schema.Script{Name: "The opening cutscene", FPS: 30}

	built := map[string]string{}    // asset id → clip id (dedup across shots)
	placementOf := map[string]int{} // asset id → placement id
	nextPlacement := 0
	haveCamera := false

	for si, shot := range introShots {
		arcPath := "/Ajioka/ADemo/" + shot.id + ".szp"
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
			b.AddSideDoc("levels/intro/cameras/"+shot.id+".json", track)
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
					if err := export.SkinnedGLB(m, key, glbPath, a.asset, clipID, false, false); err != nil {
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
				b.AddObject(schema.Asset{ID: a.asset, Name: a.name, Group: "Intro"}, &schema.Object{
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
		ctx.Progress("levels", si+1, len(introShots),
			fmt.Sprintf("intro %s: %d frames, %d actors", shot.id, sh.Frames, len(sh.Actors)))
	}

	if !doLevels || len(script.Shots) == 0 {
		return nil
	}
	b.AddSideDoc("levels/intro/opening.json", script)
	doc.Scripts = []schema.ScriptRef{{
		ID: "opening", Name: "The opening cutscene", File: "intro/opening.json",
		Description: "The seven demo shots the game streams before the title: Luigi walks the forest path, " +
			"the map points the way, the gate and the front door open. Sets, actor clips and camera cuts " +
			"are the game's own data; sound cues await the .bas decoder.",
	}}
	b.AddLevel(schema.Asset{ID: "intro", Name: "The opening"}, doc)
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
