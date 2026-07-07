// celdump decodes 3DO cel images to PNG. It can pull a single cel out of a disc
// image by path, or batch-decode every .cel file on the disc.
//
// Usage:
//
//	celdump -image disc.bin -path FrontEnd/Display/credits/bgnd.cel -o bgnd.png
//	celdump -image disc.bin -all -o outdir/          # every *.cel on the disc
//	celdump -f some.cel -o some.png                  # a standalone .cel file
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/threedo"
)

func main() {
	image := flag.String("image", "", "3DO disc image to read cels from")
	path := flag.String("path", "", "cel file path within the disc")
	file := flag.String("f", "", "standalone .cel file (instead of -image/-path)")
	all := flag.Bool("all", false, "decode every *.cel file on the disc")
	out := flag.String("o", "", "output PNG file, or output directory with -all")
	flag.Parse()

	switch {
	case *file != "":
		data, err := os.ReadFile(*file)
		if err != nil {
			die(err)
		}
		if err := decodeTo(data, *out); err != nil {
			die(err)
		}
	case *image != "" && *all:
		img, err := os.ReadFile(*image)
		if err != nil {
			die(err)
		}
		vol, err := threedo.Open(img)
		if err != nil {
			die(err)
		}
		if *out == "" {
			die(fmt.Errorf("-all needs -o outdir"))
		}
		if err := os.MkdirAll(*out, 0o755); err != nil {
			die(err)
		}
		n, fail := 0, 0
		vol.Walk(func(e threedo.Entry) error {
			ext := strings.ToLower(filepath.Ext(e.Name))
			if e.IsDir || (ext != ".cel" && ext != ".3sh") {
				return nil
			}
			data, err := vol.ReadFile(e.Path)
			if err != nil {
				fail++
				return nil
			}
			name := strings.ReplaceAll(e.Path, "/", "_") + ".png"
			if err := decodeTo(data, filepath.Join(*out, name)); err != nil {
				fmt.Fprintf(os.Stderr, "  skip %s: %v\n", e.Path, err)
				fail++
				return nil
			}
			n++
			return nil
		})
		fmt.Printf("decoded %d cels (%d failed) into %s\n", n, fail, *out)
	case *image != "" && *path != "":
		img, err := os.ReadFile(*image)
		if err != nil {
			die(err)
		}
		vol, err := threedo.Open(img)
		if err != nil {
			die(err)
		}
		data, err := vol.ReadFile(*path)
		if err != nil {
			die(err)
		}
		if err := decodeTo(data, *out); err != nil {
			die(err)
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: celdump -image disc -path P -o out.png | -image disc -all -o dir | -f file.cel -o out.png")
		os.Exit(2)
	}
}

func decodeTo(data []byte, out string) error {
	// .3sh shape files carry EA's "SHPM" magic; everything else is a 3DO cel.
	if len(data) >= 4 && string(data[0:4]) == "SHPM" {
		img, err := threedo.ParseSHPM(data)
		if err != nil {
			return err
		}
		if out == "" || out == "-" {
			b := img.Bounds()
			fmt.Printf("shpm %dx%d 16bpp\n", b.Dx(), b.Dy())
			return nil
		}
		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
		return png.Encode(f, img)
	}

	cel, err := threedo.ParseCel(data)
	if err != nil {
		return err
	}
	img, err := cel.Image()
	if err != nil {
		return err
	}
	if out == "" || out == "-" {
		fmt.Printf("cel %dx%d %dbpp packed=%v coded=%v plut=%d\n",
			cel.Width, cel.Height, cel.BPP, cel.Packed, cel.Coded, len(cel.PLUT))
		return nil
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "celdump:", err)
	os.Exit(1)
}
