// webexport builds Luigi's Mansion's Retro-X game tree from the disc — the
// game's FIRST tool-written manifest (the old Studio manifest was authored by
// hand; this replaces it end to end).
//
//	levels/mansion.json        the whole mansion: 75 streamed room shells in
//	                           storey areas, every furniture piece and door
//	                           leaf a clickable placement
//	levels/intro.json          the opening cutscene: the demo player's 7 shots
//	                           (opwf..opod) as ONE script — sets as layers,
//	                           skinned actors as placements, cameras baked
//	                           from the shots' own .scd/.sco channels
//	objects/<id>.json|.glb     furniture (content-deduped), door leaves with
//	                           their swing clips, and the intro actors
//
// Sound is absent by design for now: the .bas cue tracks are documented but
// undecoded and the JAudio banks unreversed, so cutscene sounds[] stays empty
// (the format slot exists; see RETROX.md §7).
//
// Usage (from games/luigis-mansion-gc/): go run ./extract/cmd/webexport -in DISC.iso
package main

import (
	"fmt"

	"retroreverse.com/games/luigis-mansion-gc/extract/export"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

func main() {
	cli.Main("luigis-mansion-gc", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in DISC.iso [-o DIR] [-only levels,objects]")
	}
	src, err := export.Open(ctx.In)
	if err != nil {
		return err
	}
	defer src.Close()

	b := ctx.Builder
	b.SetTitle("Luigi's Mansion")
	b.SetPlatform("Nintendo GameCube")
	b.SetYear(2001)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 640, H: 480},
		TickHz: 60,
		// The GameCube played on the same CRT TVs as the C64/N64 exports,
		// and Flipper's TEV samples textures bilinearly.
		Filter:    "crt",
		TexFilter: "linear",
	})

	doObjects := ctx.Stage("objects")
	doLevels := ctx.Stage("levels")
	if err := exportMansion(ctx, src, doObjects, doLevels); err != nil {
		return err
	}
	return exportIntro(ctx, src, doObjects, doLevels)
}

// clipMeta turns an .anm clip list into animation metadata: 30 fps, loop mode
// from the clip's own loop flag. The LAST clip of a piece is its interaction
// (chest_1 opens the chest the vacuum yanked); clips shorter than 2.5 frames
// are rest poses, not interactions.
func clipMeta(clips []export.AnmClip) (metas []schema.Animation, interaction string) {
	for _, c := range clips {
		loop := "once"
		if c.Anm.Loop {
			loop = "loop"
		}
		metas = append(metas, schema.Animation{
			ID: c.Name, Clip: c.Name, FPS: 30, Loop: loop,
		})
	}
	if n := len(clips); n > 0 && clips[n-1].Anm.Frames >= 3 {
		interaction = clips[n-1].Name
	}
	return metas, interaction
}
