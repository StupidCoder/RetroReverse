// webexport builds Stunt Car Racer's Retro-X game tree from the decoded game
// binary. Every circuit is baked as one GLB (road ribbon, walls and stroked
// decal LINES, coloured as the pre-race preview renders them) and served two
// ways: as a browsable model3d object and as a fly-through scene3d level whose
// camera starts on the grid at the circuit's own start rung. The opponent car
// (the verbatim-ported $599E2 construction) and the race view's horizon ring
// are objects; the Draw Bridge circuit carries its traced morph-target bridge
// animation.
//
// Usage (from games/stunt-car-racer-amiga/):
//
//	go run ./extract/cmd/webexport -in extracted/game.dec.bin
package main

import (
	"fmt"
	"os"
	"strings"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

func main() {
	cli.Main("stunt-car-racer-amiga", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in extracted/game.dec.bin [-o DIR]")
	}
	b := ctx.Builder
	b.SetTitle("Stunt Car Racer")
	b.SetPlatform("Amiga")
	b.SetYear(1989)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 320, H: 200},
		TickHz: 50,
		Filter: "crt",
	})
	ctx.Stage("objects")
	ctx.Stage("levels")
	return exportAll(ctx)
}

// slug turns a circuit name into a stable kebab-case id ("Roller Coaster" ->
// "roller-coaster"). Circuit names are unique, so the slug is a stable identity.
func slug(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

func chk(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}
