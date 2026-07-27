// webexport builds Super Mario Land's Retro-X game tree from the cartridge.
// Everything is reconstructed by the same decode path as cmd/levelmap:
//
//	levels/level-W-L.json     per level: tilemap (grid + view + spawn + tile
//	                          anims + collision) and every object placement
//	levels/worldN.png         the world's 256-tile background atlas (16 wide)
//	objects/<id>.json|.png    object/enemy metasprites as sprite2d objects,
//	                          content-deduped across worlds; types whose art
//	                          is undecoded share a marker
//	music/<stem>.mp3          the level themes + bonus jingle, pure-ROM synth
//
// The levels and objects stages use the machine oracle only to snapshot the
// per-world VRAM/palette pixels (never the maps); the music stage is
// oracle-free.
//
// Usage (from games/super-mario-land-gb/):
//
//	go run ./extract/cmd/webexport -in "Super Mario Land (World).gb"
package main

import (
	"fmt"
	"os"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

const (
	tile  = 8
	gut   = 1            // 1px extruded gutter so tiles don't bleed at fractional zoom
	cell  = tile + 2*gut // 10
	cols  = 16           // atlas is 16 tiles wide (256 tiles total)
	nTile = 256
)

// objRef points a placement's (world, type) at its object asset.
type objRef struct{ asset string }

func main() {
	cli.Main("super-mario-land-gb", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in <rom.gb> [-o DIR] [-only levels,objects,music]")
	}
	rom, err := os.ReadFile(ctx.In)
	if err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("Super Mario Land")
	b.SetPlatform("Nintendo Game Boy")
	b.SetYear(1989)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 160, H: 144},
		TickHz: 60,
		Filter: "gb",
	})

	var refs map[string]objRef
	if ctx.Stage("objects") {
		if refs, err = exportObjects(ctx, rom); err != nil {
			return err
		}
	}
	if ctx.Stage("levels") {
		if err := exportLevels(ctx, rom, refs); err != nil {
			return err
		}
	}
	if ctx.Stage("music") {
		if err := exportMusic(ctx, rom); err != nil {
			return err
		}
	}
	return nil
}
