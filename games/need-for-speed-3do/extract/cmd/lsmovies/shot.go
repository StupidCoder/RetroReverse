package main

// -shot: decode one frame of a stream to a PNG (content survey probe).

import (
	"image"
	"image/png"
	"os"

	"retroreverse.com/tools/platform/threedo"
)

func shotFrame(raw []byte, frameN int, out string) error {
	mv, err := threedo.DemuxStream(raw)
	if err != nil {
		return err
	}
	dec := threedo.NewCvidDecoder(mv.Width, mv.Height)
	if frameN >= len(mv.Frames) {
		frameN = len(mv.Frames) - 1
	}
	for i := 0; i <= frameN; i++ {
		dec.DecodeFrame(mv.Frames[i])
	}
	cp := image.NewRGBA(dec.Frame().Rect)
	copy(cp.Pix, dec.Frame().Pix)
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, cp)
}
