// rendertex decodes the textures in an NSBTX (`BTX0`) file and writes each to a PNG,
// upscaled for visibility. Built on tools/nds/nitro.
//
//	rendertex [-scale N] [-o DIR] file.nsbtx
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/nitro"
)

func main() {
	scale := flag.Int("scale", 6, "integer upscale factor")
	outDir := flag.String("o", "../rendered", "output directory")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: rendertex [-scale N] [-o DIR] file.nsbtx")
		os.Exit(2)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	stem := strings.TrimSuffix(filepath.Base(flag.Arg(0)), filepath.Ext(flag.Arg(0)))

	// Accept a bare NSBTX, or a .carc/NARC bundle: decompress, and render every BTX0
	// texture block it contains (the full .carc → LZ77 → NARC → BTX0 chain).
	var blocks [][]byte
	dec := nds.Decompress(data)
	if len(dec) >= 4 && string(dec[0:4]) == "BTX0" {
		blocks = [][]byte{dec}
	} else if files, err := nds.ParseNARC(dec); err == nil {
		for _, f := range files {
			if len(f) >= 4 && string(f[0:4]) == "BTX0" {
				blocks = append(blocks, f)
			}
		}
	}
	if len(blocks) == 0 {
		die(fmt.Errorf("no BTX0 texture block in %s", flag.Arg(0)))
	}

	for _, b := range blocks {
		texs, err := nitro.DecodeNSBTX(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  skip: %v\n", err)
			continue
		}
		for _, t := range texs {
			name := fmt.Sprintf("%s-%s.png", stem, safe(t.Name))
			p := filepath.Join(*outDir, name)
			if err := writePNG(p, upscale(t.Img, *scale)); err != nil {
				die(err)
			}
			fmt.Printf("  %-28s %dx%d  format %d  → %s\n", t.Name, t.Width, t.Height, t.Format, name)
		}
	}
}

// safe makes a texture name filename-friendly.
func safe(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' || r == ' ' {
			return '_'
		}
		return r
	}, s)
}

func upscale(src *image.NRGBA, s int) *image.NRGBA {
	if s <= 1 {
		return src
	}
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx()*s, b.Dy()*s))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			c := src.NRGBAAt(x, y)
			for dy := 0; dy < s; dy++ {
				for dx := 0; dx < s; dx++ {
					dst.SetNRGBA(x*s+dx, y*s+dy, c)
				}
			}
		}
	}
	return dst
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "rendertex:", err)
	os.Exit(1)
}
