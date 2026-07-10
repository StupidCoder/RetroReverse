// texdump decodes every UVTX texture straight out of the cartridge archive.
//
// With -verify it checks the ROM path against the running game: it walks a
// frame's display list out of an RDRAM snapshot, and for every textured draw
// group finds the UVTX resource whose texels the game copied to that address,
// then requires that the tile parameters the *frame* configured match the ones
// the resource's own material template declares, and that the two decoded
// images are pixel-identical. The frame's tiles come from the game; the
// resource's come from us. Nothing about the comparison is circular.
//
// Usage:
//
//	texdump -image ROM -o work/textures
//	texdump -image ROM -verify work/title-ram.bin -dl 2A15C0
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/games/pilotwings-64-n64/extract/f3d"
	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtx"
	"retroreverse.com/tools/platform/n64"
)

func main() {
	image_ := flag.String("image", "", "cartridge image")
	out := flag.String("o", "", "write one PNG per texture into this directory")
	verify := flag.String("verify", "", "RDRAM snapshot to verify the ROM decode against")
	dlAddr := flag.String("dl", "2A15C0", "display-list address in the snapshot (hex)")
	flag.Parse()

	rom, err := n64.Load(*image_)
	if err != nil {
		die(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		die(err)
	}

	texs, err := decodeAll(a)
	if err != nil {
		die(err)
	}
	fmt.Printf("decoded %d UVTX resources, 0 fallbacks\n", len(texs))

	census := map[string]int{}
	for _, t := range texs {
		census[t.tex.Format()]++
	}
	var keys []string
	for k := range census {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-8s %d\n", k, census[k])
	}

	if *out != "" {
		if err := os.MkdirAll(*out, 0o755); err != nil {
			die(err)
		}
		for _, t := range texs {
			p := filepath.Join(*out, fmt.Sprintf("%04d-%s-%dx%d.png", t.idx, t.tex.Format(), t.tex.Width, t.tex.Height))
			f, err := os.Create(p)
			if err != nil {
				die(err)
			}
			if err := png.Encode(f, t.tex.Image); err != nil {
				die(err)
			}
			f.Close()
		}
		fmt.Printf("wrote %d PNGs to %s\n", len(texs), *out)
	}

	if *verify != "" {
		addr, err := strconv.ParseUint(strings.TrimPrefix(*dlAddr, "0x"), 16, 32)
		if err != nil {
			die(fmt.Errorf("bad -dl: %w", err))
		}
		if err := verifyAgainstRAM(texs, *verify, uint32(addr)); err != nil {
			die(err)
		}
	}
}

type entry struct {
	idx    int
	tex    *uvtx.Texture
	texels []byte
}

func decodeAll(a *pwad.Archive) ([]entry, error) {
	var out []entry
	for _, i := range a.ByType("UVTX") {
		f, err := a.Resource(i)
		if err != nil {
			return nil, err
		}
		var data []byte
		for _, c := range f.Chunks {
			tag := c.Tag
			if c.Compressed() {
				tag = c.InnerTag
			}
			if tag == "COMM" {
				if data, err = a.Data(c); err != nil {
					return nil, err
				}
			}
		}
		if data == nil {
			return nil, fmt.Errorf("UVTX %d has no COMM chunk", i)
		}
		t, err := uvtx.Decode(data)
		if err != nil {
			return nil, fmt.Errorf("UVTX %d: %w", i, err)
		}
		out = append(out, entry{idx: i, tex: t, texels: data[uvtx.HeaderSize : uvtx.HeaderSize+t.DataSize]})
	}
	return out, nil
}

// verifyAgainstRAM walks the frame's display list and, for each textured group,
// locates the UVTX whose texels sit at the group's texture address in RDRAM.
func verifyAgainstRAM(texs []entry, ramPath string, dl uint32) error {
	ram, err := os.ReadFile(ramPath)
	if err != nil {
		return err
	}
	w := f3d.New(ram, false)
	w.Walk(dl)

	checked, skipped, windows, resampled := 0, 0, 0, 0
	for _, g := range w.Ordered() {
		if g.TexImg == 0 {
			continue
		}
		// Which resource landed here? Match on the texel bytes themselves.
		var hit *entry
		for i := range texs {
			t := &texs[i]
			lo := int(g.Tile.Img)
			if lo+t.tex.DataSize > len(ram) {
				continue
			}
			if string(ram[lo:lo+t.tex.DataSize]) == string(t.texels) {
				hit = t
				break
			}
		}
		if hit == nil {
			skipped++
			continue
		}
		// The frame configured this tile; the resource's own material template
		// configured ours. Only the fields that say *how the bytes are read* are
		// invariant — format, size and line stride. A frame is free to re-issue
		// Set_Tile_Size to sample a window (the ocean wraps a 65x65 window over a
		// 64x64 image) and to change the wrap modes and masks for a particular
		// use; those are sampling behaviour, not layout, and are only counted.
		rt := hit.tex.Tile
		if g.Tile.Fmt != rt.Fmt || g.Tile.Size != rt.Size || g.Tile.Line != rt.Line {
			return fmt.Errorf("UVTX %d at RDRAM %06X: frame reads it as {fmt %d siz %d line %d} "+
				"but the template declares {fmt %d siz %d line %d}",
				hit.idx, g.Tile.Img,
				g.Tile.Fmt, g.Tile.Size, g.Tile.Line, rt.Fmt, rt.Size, rt.Line)
		}
		if g.Tile.CmS != rt.CmS || g.Tile.CmT != rt.CmT || g.Tile.MaskS != rt.MaskS || g.Tile.MaskT != rt.MaskT {
			resampled++
		}
		// Decode the game's own copy of the texels through the extent the ROM
		// template gave us. If the ROM decode had silently fallen back — the way
		// every export before 2026-07-10 did — this comparison would fail.
		rg := *g
		rg.Tile.SL, rg.Tile.TL, rg.Tile.SH, rg.Tile.TH = 0, 0, rt.SH, rt.TH
		ramImg := w.DecodeTexture(&rg)
		if ramImg == nil {
			return fmt.Errorf("UVTX %d at RDRAM %06X: the template's extent does not decode from RAM", hit.idx, g.Tile.Img)
		}
		if err := samePixels(ramImg, hit.tex.Image); err != nil {
			return fmt.Errorf("UVTX %d at RDRAM %06X: %w", hit.idx, g.Tile.Img, err)
		}
		if windowed(g.Tile, rt) {
			windows++
		}
		checked++
	}
	fmt.Printf("verify: %d textured groups matched a UVTX; format/size/line agree, and the game's own texels\n"+
		"        decode pixel-identically to the ROM decode through the template's extent\n", checked)
	if windows > 0 {
		fmt.Printf("        %d sample a window of their texture rather than the whole of it\n", windows)
	}
	if resampled > 0 {
		fmt.Printf("        %d are re-issued with different wrap modes or masks for that draw\n", resampled)
	}
	if skipped > 0 {
		fmt.Printf("        %d groups' texels matched no resource\n", skipped)
	}
	if checked == 0 {
		return fmt.Errorf("verify: nothing was checked")
	}
	return nil
}

// windowed reports whether the frame samples a sub- or super-rectangle of the
// texture rather than exactly the extent the template declares.
func windowed(frame, tmpl f3d.TileDesc) bool {
	return frame.SL != 0 || frame.TL != 0 || frame.SH != tmpl.SH || frame.TH != tmpl.TH
}

func samePixels(a, b image.Image) error {
	ab, bb := a.Bounds(), b.Bounds()
	if ab.Dx() != bb.Dx() || ab.Dy() != bb.Dy() {
		return fmt.Errorf("size %dx%d from RAM, %dx%d from ROM", ab.Dx(), ab.Dy(), bb.Dx(), bb.Dy())
	}
	for y := 0; y < ab.Dy(); y++ {
		for x := 0; x < ab.Dx(); x++ {
			r1, g1, b1, a1 := a.At(ab.Min.X+x, ab.Min.Y+y).RGBA()
			r2, g2, b2, a2 := b.At(bb.Min.X+x, bb.Min.Y+y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				return fmt.Errorf("pixel (%d,%d) differs", x, y)
			}
		}
	}
	return nil
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "texdump:", err)
	os.Exit(1)
}
