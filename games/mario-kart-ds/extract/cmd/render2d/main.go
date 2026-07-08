// render2d composes the NITRO 2D UI graphics — NSCR screens through NCGR tiles and
// NCLR palettes — to PNG. It accepts loose .NSCR/.NCGR/.NCLR files (a directory) or
// a .carc/NARC whose named sub-files carry the same formats, pairing each screen
// with its tile/palette files by name (longest shared prefix, the game's own naming
// convention: "…_picture.NSCR" ↔ "…_b.NCGR"/"…_b.NCLR").
//
//	render2d [-scale N] [-o DIR] <dir | file.carc>
//
// NCGR files with no matching screen are rendered as raw tile sheets.
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

// item is one 2D-format file, from disk or from inside a NARC.
type item struct {
	name string // base name without extension
	ext  string // upper-case extension: NSCR/NCGR/NCLR
	data []byte
}

func main() {
	scale := flag.Int("scale", 1, "integer upscale factor")
	outDir := flag.String("o", "../rendered/ui", "output directory")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: render2d [-scale N] [-o DIR] <dir | file.carc>")
		os.Exit(2)
	}
	items, tag, err := collect(flag.Arg(0))
	if err != nil {
		die(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}

	var chrs, pals, scrs []item
	for _, it := range items {
		switch it.ext {
		case "NCGR":
			chrs = append(chrs, it)
		case "NCLR":
			pals = append(pals, it)
		case "NSCR":
			scrs = append(scrs, it)
		}
	}
	rendered := 0
	usedChr := map[string]bool{}
	for _, s := range scrs {
		chr := bestMatch(s.name, chrs)
		pal := bestMatch(s.name, pals)
		if chr == nil || pal == nil {
			fmt.Fprintf(os.Stderr, "  no tiles/palette for screen %s\n", s.name)
			continue
		}
		usedChr[chr.name] = true
		img, err := compose(s.data, chr.data, pal.data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", s.name, err)
			continue
		}
		name := tag + s.name + ".png"
		if err := writePNG(filepath.Join(*outDir, name), upscale(img, *scale)); err != nil {
			die(err)
		}
		fmt.Printf("  %-44s %dx%d  → %s\n", s.name, img.Bounds().Dx(), img.Bounds().Dy(), name)
		rendered++
	}
	// Tile sheets for NCGRs no screen references (icons, sprites).
	for _, c := range chrs {
		if usedChr[c.name] {
			continue
		}
		pal := bestMatch(c.name, pals)
		if pal == nil {
			continue
		}
		chr, err1 := nitro.ParseNCGR(c.data)
		pl, err2 := nitro.ParseNCLR(pal.data)
		if err1 != nil || err2 != nil {
			continue
		}
		img := nitro.TileSheet(chr, pl, 0)
		name := tag + c.name + "-tiles.png"
		if err := writePNG(filepath.Join(*outDir, name), upscale(img, *scale)); err != nil {
			die(err)
		}
		fmt.Printf("  %-44s %d tiles  → %s\n", c.name, len(chr.Tiles), name)
		rendered++
	}
	fmt.Printf("%d images\n", rendered)
}

// collect gathers 2D-format items from a directory of loose files or a .carc/NARC.
func collect(path string) ([]item, string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}
	var items []item
	if st.IsDir() {
		ents, err := os.ReadDir(path)
		if err != nil {
			return nil, "", err
		}
		for _, e := range ents {
			ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(e.Name()), "."))
			if ext != "NSCR" && ext != "NCGR" && ext != "NCLR" {
				continue
			}
			data, err := os.ReadFile(filepath.Join(path, e.Name()))
			if err != nil {
				return nil, "", err
			}
			items = append(items, item{name: strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())), ext: ext, data: data})
		}
		return items, "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	files, err := nds.ParseNARCFiles(data)
	if err != nil {
		return nil, "", err
	}
	for i, f := range files {
		name := f.Name
		if name == "" {
			name = fmt.Sprintf("file%03d", i)
		}
		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(name), "."))
		if ext != "NSCR" && ext != "NCGR" && ext != "NCLR" {
			continue
		}
		items = append(items, item{name: strings.TrimSuffix(name, filepath.Ext(name)), ext: ext, data: f.Data})
	}
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return items, stem + "-", nil
}

// bestMatch picks the candidate best matching a screen's name: longest common
// prefix, preferring the game's "_b" (background) bank over the "_o"/".nce" sprite
// banks that share the same stem.
func bestMatch(name string, cands []item) *item {
	best, bestScore := -1, 0
	for i, c := range cands {
		l := 0
		for l < len(name) && l < len(c.name) && name[l] == c.name[l] {
			l++
		}
		rest := c.name[l:]
		score := l*100 - len(rest)
		if strings.Contains(rest, "_b") {
			score += 50 // BG bank
		}
		if strings.Contains(rest, ".nce") || strings.Contains(rest, "_o") {
			score -= 50 // sprite/OBJ bank
		}
		if score > bestScore {
			best, bestScore = i, score
		}
	}
	if best < 0 && len(cands) == 1 {
		best = 0
	}
	if best < 0 {
		return nil
	}
	return &cands[best]
}

func compose(scr, chr, pal []byte) (*image.NRGBA, error) {
	s, err := nitro.ParseNSCR(scr)
	if err != nil {
		return nil, err
	}
	c, err := nitro.ParseNCGR(chr)
	if err != nil {
		return nil, err
	}
	p, err := nitro.ParseNCLR(pal)
	if err != nil {
		return nil, err
	}
	return nitro.ComposeScreen(s, c, p), nil
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
	fmt.Fprintln(os.Stderr, "render2d:", err)
	os.Exit(1)
}
