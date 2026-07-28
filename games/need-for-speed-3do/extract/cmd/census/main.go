package main

// census: decode one frame of EVERY movie on the disc into a contact sheet
// (13 columns of 160x120 cells, nearest-neighbour downscale, movie name
// stamped as a 1px-shadow label using a tiny built-in 3x5 digit/letter font
// is overkill — we just order them and print a separate index instead).

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	vol, err := threedo.Open(data)
	if err != nil {
		panic(err)
	}
	var paths []string
	vol.Walk(func(e threedo.Entry) error {
		if !e.IsDir && strings.Contains(e.Path, "Movies") {
			paths = append(paths, e.Path)
		}
		return nil
	})
	sort.Strings(paths)

	const cw, ch, cols = 160, 120, 13
	rows := (len(paths) + cols - 1) / cols
	sheet := image.NewRGBA(image.Rect(0, 0, cols*cw, rows*ch))
	draw.Draw(sheet, sheet.Bounds(), image.Black, image.Point{}, draw.Src)

	for i, p := range paths {
		raw, err := vol.ReadFile(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", p, err)
			continue
		}
		mv, err := threedo.DemuxStream(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", p, err)
			continue
		}
		dec := threedo.NewCvidDecoder(mv.Width, mv.Height)
		n := 40
		if n >= len(mv.Frames) {
			n = len(mv.Frames) - 1
		}
		for f := 0; f <= n; f++ {
			dec.DecodeFrame(mv.Frames[f])
		}
		src := dec.Frame()
		ox, oy := (i%cols)*cw, (i/cols)*ch
		for y := 0; y < ch; y++ {
			for x := 0; x < cw; x++ {
				sx := x * mv.Width / cw
				sy := y * mv.Height / ch
				sheet.Set(ox+x, oy+y, src.At(sx, sy))
			}
		}
		fmt.Printf("%3d  r%02d c%02d  %s  (%d frames %dx%d)\n", i, i/cols, i%cols, p, len(mv.Frames), mv.Width, mv.Height)
	}
	f, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, sheet); err != nil {
		panic(err)
	}
}
