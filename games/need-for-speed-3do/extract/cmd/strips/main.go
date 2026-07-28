package main

// strips: one 5-frame filmstrip per commentary topic family, tiled into one
// contact sheet (2 columns of strips), to read each topic's visual subject.

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	"retroreverse.com/tools/platform/threedo"
)

var reps = []string{
	"1.1", "2.1", "3.2", "5.1", "6.3", "7.1", "8.1", "11.1", "12.1", "13.1",
	"14.1", "15.1", "16.1", "17.1", "18.1", "20.2", "21.1", "22.1", "23.1",
	"24.1", "26.1", "27.1", "28.1", "30.1", "31.1", "32.1", "33.1", "34.1",
	"43.2", "44.1", "46.1", "52.2", "56.1", "57.1", "59.1", "60.1", "64.1",
	"65.1", "67.1", "68.2", "69.1", "70.1", "101.1", "101.4",
}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	vol, err := threedo.Open(data)
	if err != nil {
		panic(err)
	}
	const fw, fh, per = 128, 96, 5
	cols := 2
	rows := (len(reps) + cols - 1) / cols
	sheet := image.NewRGBA(image.Rect(0, 0, cols*(fw*per+8), rows*fh))
	draw.Draw(sheet, sheet.Bounds(), image.Black, image.Point{}, draw.Src)
	for i, stem := range reps {
		raw, err := vol.ReadFile("Movies/" + stem + ".stream")
		if err != nil {
			fmt.Fprintln(os.Stderr, stem, err)
			continue
		}
		mv, err := threedo.DemuxStream(raw)
		if err != nil {
			continue
		}
		dec := threedo.NewCvidDecoder(mv.Width, mv.Height)
		n := len(mv.Frames)
		picks := map[int]int{}
		for k := 0; k < per; k++ {
			picks[(n-1)*(k*2+1)/(per*2)] = k
		}
		for f := 0; f < n; f++ {
			dec.DecodeFrame(mv.Frames[f])
			if k, ok := picks[f]; ok {
				src := dec.Frame()
				ox := (i % cols) * (fw*per + 8)
				oy := (i / cols) * fh
				for y := 0; y < fh; y++ {
					for x := 0; x < fw; x++ {
						sheet.Set(ox+k*fw+x, oy+y, src.At(x*mv.Width/fw, y*mv.Height/fh))
					}
				}
			}
		}
		fmt.Printf("%d %s\n", i, stem)
	}
	f, _ := os.Create(os.Args[2])
	defer f.Close()
	png.Encode(f, sheet)
}
