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
