// loadingscreen renders Elite's multicolor-bitmap loading picture to a PNG.
//
// The picture is stored uncompressed across three tape segments and is never
// resident all at once in a single extracted region, so this tool assembles it
// from the per-segment .prg files produced by `extract`:
//
//	seg 7 ($4000-$6000) : 8000-byte VIC bitmap
//	seg 6 ($6000-$6400) : video-matrix colour (bit-pairs 01/10)
//	seg 5 ($6000-$6400) : colour RAM, copied to $D800 at run time (bit-pair 11)
//
// The background ($D021) is white ($01), set by the loader at $CE60.
//
// Usage: loadingscreen [-extracted dir] [-o dir] [-scale n]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/tools/platform/c64/gfx"
)

const background = 0x01 // $D021 set at $CE60

func main() {
	extracted := flag.String("extracted", "../extracted", "directory of extracted .prg files")
	outDir := flag.String("o", "../rendered", "output directory")
	scale := flag.Int("scale", 2, "pixel scale factor")
	flag.Parse()
	if err := run(*extracted, *outDir, *scale); err != nil {
		fmt.Fprintln(os.Stderr, "loadingscreen:", err)
		os.Exit(1)
	}
}

func run(extracted, outDir string, scale int) error {
	bitmap, err := loadPRG(extracted, "_seg07_4000.prg")
	if err != nil {
		return err
	}
	matrix, err := loadPRG(extracted, "_seg06_6000.prg")
	if err != nil {
		return err
	}
	colorRAM, err := loadPRG(extracted, "_seg05_6000.prg")
	if err != nil {
		return err
	}

	img := gfx.RenderBitmapMC(bitmap, matrix, colorRAM, background, scale)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	out := filepath.Join(outDir, "loading-screen.png")
	if err := gfx.WritePNG(out, img); err != nil {
		return err
	}
	fmt.Printf("wrote %s (multicolor bitmap, %dx%d px)\n", out, img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}

// loadPRG finds the single extracted file whose name contains pattern and
// returns its data with the 2-byte load address stripped.
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
