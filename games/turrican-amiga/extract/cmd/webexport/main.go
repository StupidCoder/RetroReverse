// webexport builds Turrican's Retro-X game tree from the disk image:
//
//	levels/world<w>_scene<s>.json  per scene: tilemap (grid + hflip + grid
//	                               collision + view + spawn) and every AI
//	                               placement at its spawn-init-resolved
//	                               sprite frame
//	levels/atlas<w>.png            the world's 32x32 tiles in its palette
//	objects/<id>.json|.png         every BOB frame table as a sprite2d object
//	                               (one-row strip; placements select their
//	                               orientation frame via a still animation)
//	music/*.mp3                    every distinct TFMX sub-song, rendered by
//	                               the real 68000 sound driver in the m68k
//	                               interpreter
//
// Placement comes straight off the disk via the scene package; each object's
// displayed sprite/frame/position is resolved by running its AI handler's
// spawn-init in the 68000 interpreter (scene.Sim).
//
// Usage (from games/turrican-amiga/): go run ./extract/cmd/webexport -in Turrican.adf
package main

import (
	"fmt"
	"os"

	"retroreverse.com/games/turrican-amiga/extract/scene"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

// amigaView is Turrican's visible playfield in pixels (one screen).
const amigaViewW, amigaViewH = 320, 256

func main() {
	cli.Main("turrican-amiga", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in Turrican.adf [-o DIR] [-only levels,objects,music]")
	}
	adf, err := os.ReadFile(ctx.In)
	if err != nil {
		return err
	}
	game, err := scene.Load(adf)
	if err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("Turrican")
	b.SetPlatform("Amiga")
	b.SetYear(1990)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: amigaViewW, H: amigaViewH},
		TickHz: 50,
		Filter: "crt",
	})

	// Resolve every placed object's displayed sprite/frame ONCE (the AI
	// handler spawn-init in the 68000 interpreter); both stages read it.
	scenes := make([][]scene.Scene, scene.NumWorlds)
	for w := 0; w < scene.NumWorlds; w++ {
		scenes[w] = game.Scenes(w)
		sim := game.NewSim(w)
		for si := range scenes[w] {
			for oi := range scenes[w][si].Objects {
				game.Resolve(sim, &scenes[w][si].Objects[oi])
			}
		}
	}

	var refs map[string]objRef
	if ctx.Stage("objects") {
		if refs, err = exportObjects(ctx, game, scenes); err != nil {
			return err
		}
	}
	if ctx.Stage("levels") {
		if err := exportLevels(ctx, game, scenes, refs); err != nil {
			return err
		}
	}
	if ctx.Stage("music") {
		if err := exportMusic(ctx, adf, game); err != nil {
			return err
		}
	}
	return nil
}

// objRef points a placement's frame-table key ("resident/1B3C4", "w2/205F0")
// at its object asset.
type objRef struct{ asset string }
