package main

// eyeprobe: verify the Go eye compositor (merc/eyes.go) against a live slot
// capture (bootoracle -gstex, raw GS alpha). Composites Daxter's right eye
// with the eye-control values read from title-logo.state RAM and diffs
// against the capture.
import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/tpage"
	"retroreverse.com/tools/lib/iso9660"
)

func main() {
	imageF := flag.String("image", "image/Jak and Daxter - The Precursor Legacy.iso", "disc image")
	ref := flag.String("ref", "", "raw -gsfb eye-buffer PNG (RGB reference)")
	refA := flag.String("refa", "", "-gstex slot capture PNG (alpha reference)")
	ox := flag.Int("ox", 32, "slot x offset in the -gsfb PNG")
	oy := flag.Int("oy", 32, "slot y offset in the -gsfb PNG")
	out := flag.String("out", "", "write composite PNG")
	ex := flag.Float64("x", 0.1953125, "iris x")
	ey := flag.Float64("y", 0.203125, "iris y")
	is := flag.Float64("iris", 0.375, "iris size")
	ps := flag.Float64("pupil", 0, "pupil size")
	lid := flag.String("lid", "sk-eye-lid", "lid texture name")
	lidState := flag.Float64("lidstate", 0, "lid state")
	lidH := flag.Float64("lidh", 1.0, "lid height")
	mirror := flag.Bool("mirror", true, "right eye (mirrored lid)")
	flag.Parse()

	f, err := os.Open(*imageF)
	check(err)
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	tab, err := goalobj.LoadSymTab("work/goal.txt")
	check(err)
	data, err := vol.ReadFile("CGO/ART.CGO;1")
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	tex := map[string]image.Image{}
	for _, e := range d.Entries {
		if !strings.HasPrefix(e.Name, "tpage-") || len(e.Data) < 12 || binary.LittleEndian.Uint32(e.Data[8:]) < 4 {
			continue
		}
		obj, _, err := goalobj.Link(e.Data, 0, tab)
		check(err)
		pg, err := tpage.Load(obj)
		check(err)
		for i := range pg.Textures {
			t := &pg.Textures[i]
			switch t.Name {
			case "bam-iris-16x16", "autoeye-pupil", "autoeye-lid", "sk-eye-lid":
				if _, ok := tex[t.Name]; !ok {
					img, err := pg.DecodeGS(t, 0)
					check(err)
					tex[t.Name] = img
				}
			}
		}
	}
	s := merc.CompositeEye(tex["bam-iris-16x16"], tex["autoeye-pupil"], tex[*lid],
		merc.EyeParams{X: float32(*ex), Y: float32(*ey), IrisSize: float32(*is),
			PupilSize: float32(*ps), LidState: float32(*lidState), LidHeight: float32(*lidH)}, *mirror)
	img := s.RawImage()
	if *out != "" {
		w, err := os.Create(*out)
		check(err)
		check(png.Encode(w, img))
		w.Close()
	}
	if *refA != "" {
		rf, err := os.Open(*refA)
		check(err)
		rimg, err := png.Decode(rf)
		check(err)
		nd := 0
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				p1 := img.NRGBAAt(x, y)
				p2 := color.NRGBAModel.Convert(rimg.At(x, y)).(color.NRGBA)
				if p1.A != p2.A {
					nd++
					if nd < 8 {
						fmt.Printf("alpha diff (%2d,%2d): got %d want %d\n", x, y, p1.A, p2.A)
					}
				}
			}
		}
		fmt.Printf("alpha diff pixels: %d/1024\n", nd)
	}
	if *ref != "" {
		rf, err := os.Open(*ref)
		check(err)
		rimg, err := png.Decode(rf)
		check(err)
		var maxd, nd int
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				p1 := img.NRGBAAt(x, y)
				p2 := color.NRGBAModel.Convert(rimg.At(x+*ox, y+*oy)).(color.NRGBA)
				d := absi(int(p1.R)-int(p2.R)) + absi(int(p1.G)-int(p2.G)) +
					absi(int(p1.B)-int(p2.B))
				if d > 0 {
					nd++
					if d > maxd {
						maxd = d
					}
					if nd < 12 {
						fmt.Printf("diff (%2d,%2d): got %v want %v\n", x, y, p1, p2)
					}
				}
			}
		}
		fmt.Printf("diff pixels: %d/1024, max channel-sum delta %d\n", nd, maxd)
	}
}

func absi(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
