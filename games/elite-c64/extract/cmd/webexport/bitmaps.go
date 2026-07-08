// bitmaps.go is the bitmaps stage (folded from ex-cmd/loadingscreen): it renders
// Elite's multicolor-bitmap loading picture to bitmaps/loading.png.
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

	"retroreverse.com/tools/platform/c64/gfx"
)

const loadBackground = 0x01 // $D021 set at $CE60
const loadScale = 2         // 2x for a crisp web bitmap

func exportBitmaps(inDir, outDir string) []AssetIndex {
	bitmap, err := loadPRG(inDir, "_seg07_4000.prg")
	chk(err)
	matrix, err := loadPRG(inDir, "_seg06_6000.prg")
	chk(err)
	colorRAM, err := loadPRG(inDir, "_seg05_6000.prg")
	chk(err)

	img := gfx.RenderBitmapMC(bitmap, matrix, colorRAM, loadBackground, loadScale)
	bitmapsDir := filepath.Join(outDir, "bitmaps")
	chk(os.MkdirAll(bitmapsDir, 0o755))
	out := filepath.Join(bitmapsDir, "loading.png")
	chk(gfx.WritePNG(out, img))
	fmt.Fprintf(os.Stderr, "[bitmaps] loading.png (multicolor bitmap, %dx%d px)\n",
		img.Bounds().Dx(), img.Bounds().Dy())
	return []AssetIndex{{Name: "Loading screen", File: "bitmaps/loading.png"}}
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
