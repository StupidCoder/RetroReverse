// webexport extracts Captain Toad: Treasure Tracker's web deliverables from the
// cartridge image into site/public/captain-toad-treasure-tracker-3ds/. The first
// asset is the HOME-Menu banner: the 3-D logo scene (CBMD → LZ11 → CGFX),
// exported as one GLB with embedded PNG textures.
//
//	webexport -in game.cci [-o DIR] [-texdump DIR]
//
// The CGFX → GLB conversion lives in tools/platform/n3ds/cgfxglb, shared with
// the other 3DS title; its package comment records the format's traps. This
// banner is a static, folded triptych — three quads, no camera, light or
// animation — where Super Mario 3D Land's is a rigged, animated scene, and the
// same exporter covers both.
package main

import (
	"fmt"
	"os"
	"strings"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/n3ds"
	"retroreverse.com/tools/platform/n3ds/cgfxglb"
)

func main() {
	cli.Main("captain-toad-treasure-tracker-3ds", runCLI)
}

func runCLI(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in game.cci [-o DIR]")
	}
	b := ctx.Builder
	b.SetTitle("Captain Toad: Treasure Tracker")
	b.SetPlatform("Nintendo 3DS")
	b.SetYear(2018)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 400, H: 240},
		TickHz: 60,
		// the 3DS's backlit TFT + PICA200's bilinear sampling
		Filter:    "ds",
		TexFilter: "linear",
	})
	ctx.Stage("objects")
	out, err := b.Path("objects", "banner.glb")
	if err != nil {
		return err
	}
	if err := run(ctx.In, out, ""); err != nil {
		return err
	}
	b.AddObject(schema.Asset{ID: "banner", Name: "HOME Menu Banner", Group: "Banner"}, &schema.Object{
		Type: schema.ObjectModel3D, Name: "HOME Menu Banner", Model: "banner.glb",
	})
	ctx.Progress("objects", 1, 1, "banner.glb")

	return exportStages(ctx)
}

// stages are the levels exported so far. The opening diorama is the first: it is
// what the oracle boots into, so its geometry can be checked against a frame the
// game itself drew.
var stages = []struct{ id, stage, name string }{
	{"season1-opening", "Season1OpeningStage", "Opening Stage"},
}

func exportStages(ctx *cli.Context) error {
	img, err := os.ReadFile(ctx.In)
	if err != nil {
		return err
	}
	ncsd, err := n3ds.ParseNCSD(img)
	if err != nil {
		return err
	}
	cxi, err := ncsd.Executable()
	if err != nil {
		return err
	}
	fs, err := cxi.RomFS()
	if err != nil {
		return err
	}
	ctx.Stage("levels")
	for i, st := range stages {
		file := st.id + ".glb"
		out, err := ctx.Builder.Path("levels", file)
		if err != nil {
			return err
		}
		tris, skipped, err := exportStage(fs, st.stage, out)
		if err != nil {
			return fmt.Errorf("%s: %w", st.stage, err)
		}
		if len(skipped) > 0 {
			fmt.Printf("  %s: %d placed objects have no model archive yet: %s\n",
				st.stage, len(skipped), strings.Join(skipped, ", "))
		}
		ctx.Builder.AddLevel(schema.Asset{ID: st.id, Name: st.name, Group: "Stages"}, &schema.Level{
			Type: schema.LevelScene3D,
			Camera: &schema.Camera{
				Mode: "orbit", FOV: 45, Near: 10, Far: 20000,
				Pos:    []float64{1800, 900, 1800},
				Target: []float64{0, -400, 0},
			},
			Scene: &schema.Scene{Layers: []schema.Layer{{ID: "stage", File: file}}},
		})
		ctx.Progress("levels", i+1, len(stages), fmt.Sprintf("%s (%d tris)", file, tris))
	}
	return nil
}

func run(in, out, texdump string) error {
	img, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	g, err := n3ds.BannerScene(img)
	if err != nil {
		return err
	}
	return cgfxglb.ExportBanner(g, out, "banner", texdump)
}
