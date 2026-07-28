// pictures.go is the pictures stage: it renders Elite's multicolor-bitmap
// loading picture to pictures/loading.png.
//
// The picture is stored uncompressed across three tape segments and is never
// resident all at once, so it is assembled from the per-segment .prg files:
//
//	seg 7 ($4000-$6000) : 8000-byte VIC bitmap
//	seg 6 ($6000-$6400) : video-matrix colour (bit-pairs 01/10)
//	seg 5 ($6000-$6400) : colour RAM, copied to $D800 at run time (bit-pair 11)
//
// The background ($D021) is white ($01), set by the loader at $CE60.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/c64/gfx"
)

const loadBackground = 0x01 // $D021 set at $CE60
const loadScale = 2         // 2x for a crisp web bitmap

func exportLoadingScreen(ctx *cli.Context) error {
	bitmap, err := loadPRG(ctx.In, "_seg07_4000.prg")
	if err != nil {
		return err
	}
	matrix, err := loadPRG(ctx.In, "_seg06_6000.prg")
	if err != nil {
		return err
	}
	colorRAM, err := loadPRG(ctx.In, "_seg05_6000.prg")
	if err != nil {
		return err
	}

	img := gfx.RenderBitmapMC(bitmap, matrix, colorRAM, loadBackground, loadScale)
	out, err := ctx.Builder.Path("pictures", "loading.png")
	if err != nil {
		return err
	}
	if err := gfx.WritePNG(out, img); err != nil {
		return err
	}
	ctx.Builder.AddMedia(schema.Asset{
		ID: "loading", Category: schema.CategoryPicture,
		Name: "Loading screen",
		File: "pictures/loading.png",
		W:    img.Bounds().Dx(), H: img.Bounds().Dy(),
		// The PNG is a baked upscale; the info panel should state what the
		// VIC-II actually displays: 160 double-wide multicolor pixels per line.
		Stats: map[string]any{
			"Native": "160 × 200 px (double-wide pixels)",
			"Colors": "16 (VIC-II multicolor)",
			"PNG":    fmt.Sprintf("%d × %d px (%d× upscale)", img.Bounds().Dx(), img.Bounds().Dy(), loadScale),
		},
	})
	ctx.Progress("pictures", 1, 1, fmt.Sprintf("loading.png (multicolor bitmap, %dx%d px)",
		img.Bounds().Dx(), img.Bounds().Dy()))
	return nil
}

// loadPRG finds the single extracted file whose name contains pattern and returns
// its data with the 2-byte load address stripped.
func loadPRG(dir, pattern string) ([]byte, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*"+pattern))
	if err != nil {
		return nil, err
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("expected exactly one file matching *%s in %s, found %d", pattern, dir, len(matches))
	}
	raw, err := os.ReadFile(matches[0])
	if err != nil {
		return nil, err
	}
	if len(raw) < 2 {
		return nil, fmt.Errorf("%s: too short", matches[0])
	}
	return raw[2:], nil
}
