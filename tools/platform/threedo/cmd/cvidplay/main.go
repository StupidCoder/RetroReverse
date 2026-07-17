// cvidplay decodes 3DO streamed FMV (Cinepak "cvid" video in a .Stream
// container) to viewable frames. It can pull a movie straight out of a disc
// image by path, batch every movie on the disc, or decode a standalone .Stream.
//
// Usage:
//
//	cvidplay -image disc.bin -path Movies/eac.stream -gif eac.gif
//	cvidplay -image disc.bin -path Movies/pioneer.stream -o frames/   # PNG sequence
//	cvidplay -image disc.bin -all -o movies/                          # every movie -> a GIF each
//	cvidplay -f some.stream -gif some.gif
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	imagePath := flag.String("image", "", "3DO disc image to read the movie from")
	path := flag.String("path", "", "movie file path within the disc (e.g. Movies/eac.stream)")
	file := flag.String("f", "", "standalone .stream file (instead of -image/-path)")
	all := flag.Bool("all", false, "decode every *.stream movie on the disc (needs -o dir)")
	out := flag.String("o", "", "output: PNG-sequence directory, or output dir with -all")
	gifOut := flag.String("gif", "", "write an animated GIF to this file")
	fps := flag.Int("fps", 15, "playback rate for the GIF")
	every := flag.Int("every", 1, "keep every Nth frame (thins a PNG sequence or GIF)")
	flag.Parse()

	switch {
	case *file != "":
		data, err := os.ReadFile(*file)
		if err != nil {
			die(err)
		}
		emit(data, *out, *gifOut, *fps, *every)
	case *imagePath != "" && *all:
		if *out == "" {
			die(fmt.Errorf("-all needs -o outdir"))
		}
		vol := openDisc(*imagePath)
		if err := os.MkdirAll(*out, 0o755); err != nil {
			die(err)
		}
		n := 0
		vol.Walk(func(e threedo.Entry) error {
			if e.IsDir || !strings.HasSuffix(strings.ToLower(e.Name), ".stream") {
				return nil
			}
			data, err := vol.ReadFile(e.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  skip %s: %v\n", e.Path, err)
				return nil
			}
			base := strings.TrimSuffix(e.Name, filepath.Ext(e.Name))
			g := filepath.Join(*out, base+".gif")
			if err := emitErr(data, "", g, *fps, *every); err != nil {
				fmt.Fprintf(os.Stderr, "  skip %s: %v\n", e.Path, err)
				return nil
			}
			n++
			return nil
		})
		fmt.Printf("decoded %d movies into %s\n", n, *out)
	case *imagePath != "" && *path != "":
		vol := openDisc(*imagePath)
		data, err := vol.ReadFile(*path)
		if err != nil {
			die(err)
		}
		emit(data, *out, *gifOut, *fps, *every)
	default:
		fmt.Fprintln(os.Stderr, "usage: cvidplay -image disc -path Movies/X.stream [-gif out.gif | -o framesdir]")
		fmt.Fprintln(os.Stderr, "       cvidplay -image disc -all -o moviesdir")
		fmt.Fprintln(os.Stderr, "       cvidplay -f X.stream -gif out.gif")
		os.Exit(2)
	}
}

// decodeAll demuxes and decodes a movie into a slice of frames (each a fresh copy).
func decodeAll(data []byte, every int) (*threedo.CvidMovie, []*image.RGBA, error) {
	mv, err := threedo.DemuxStream(data)
	if err != nil {
		return nil, nil, err
	}
	if mv.Codec != "cvid" && mv.Codec != "" {
		return nil, nil, fmt.Errorf("unsupported movie codec %q (only Cinepak cvid)", mv.Codec)
	}
	dec := threedo.NewCvidDecoder(mv.Width, mv.Height)
	var frames []*image.RGBA
	for i, fr := range mv.Frames {
		dec.DecodeFrame(fr)
		if every > 1 && i%every != 0 {
			continue
		}
		cp := image.NewRGBA(dec.Frame().Rect)
		copy(cp.Pix, dec.Frame().Pix)
		frames = append(frames, cp)
	}
	return mv, frames, nil
}

func emit(data []byte, outDir, gifOut string, fps, every int) {
	if err := emitErr(data, outDir, gifOut, fps, every); err != nil {
		die(err)
	}
}

func emitErr(data []byte, outDir, gifOut string, fps, every int) error {
	mv, frames, err := decodeAll(data, every)
	if err != nil {
		return err
	}
	switch {
	case gifOut != "":
		if err := writeGIF(gifOut, frames, fps); err != nil {
			return err
		}
		fmt.Printf("wrote %s: %dx%d, %d frames\n", gifOut, mv.Width, mv.Height, len(frames))
	case outDir != "":
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
		for i, f := range frames {
			fn := filepath.Join(outDir, fmt.Sprintf("frame%04d.png", i))
			if err := writePNG(fn, f); err != nil {
				return err
			}
		}
		fmt.Printf("wrote %d PNG frames (%dx%d) to %s\n", len(frames), mv.Width, mv.Height, outDir)
	default:
		return fmt.Errorf("need -gif FILE or -o DIR")
	}
	return nil
}

func writeGIF(path string, frames []*image.RGBA, fps int) error {
	if len(frames) == 0 {
		return fmt.Errorf("no frames to encode")
	}
	delay := 100 / fps // GIF delay is in 100ths of a second
	if delay < 2 {
		delay = 2
	}
	g := &gif.GIF{}
	for _, f := range frames {
		pi := image.NewPaletted(f.Rect, palette.Plan9)
		draw.FloydSteinberg.Draw(pi, f.Rect, f, image.Point{})
		g.Image = append(g.Image, pi)
		g.Delay = append(g.Delay, delay)
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	return gif.EncodeAll(out, g)
}

func writePNG(path string, img image.Image) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, img)
}

func openDisc(p string) *threedo.Volume {
	img, err := os.ReadFile(p)
	if err != nil {
		die(err)
	}
	vol, err := threedo.Open(img)
	if err != nil {
		die(err)
	}
	return vol
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "cvidplay:", err)
	os.Exit(1)
}
