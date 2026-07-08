// renderall sweeps the extracted filesystem and renders every 2D asset it can
// decode: all NITRO textures (standalone .nsbtx files and the BTX0 blocks inside
// .carc archives) and all 2D tile art (NSCR screens composed through their NCGR
// tiles and NCLR palettes, both loose and NARC-embedded; screenless NCGRs become
// tile sheets). One command regenerates every figure:
//
//	renderall [-root DIR] [-o DIR]
//
// Layout: textures → OUT/course/<archive>/ (Course/) or OUT/tex/<path>/ (others);
// screens/sheets → OUT/ui/<archive-or-dir>/.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/nitro"
)

var (
	nTex, nScr, nSheet, nSkip int
)

func main() {
	root := flag.String("root", "../extracted/files", "extracted filesystem root")
	out := flag.String("o", "../rendered", "output root")
	flag.Parse()

	// Pass 1: every .carc and .nsbtx — textures, and NARC-embedded 2D sets.
	var carcs, nsbtxs []string
	dirs2d := map[string][]string{} // dir → loose 2D files
	filepath.Walk(*root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(p)) {
		case ".carc":
			carcs = append(carcs, p)
		case ".nsbtx":
			nsbtxs = append(nsbtxs, p)
		case ".nscr", ".ncgr", ".nclr":
			d := filepath.Dir(p)
			dirs2d[d] = append(dirs2d[d], p)
		}
		return nil
	})
	sort.Strings(carcs)
	sort.Strings(nsbtxs)

	for _, p := range nsbtxs {
		rel, _ := filepath.Rel(*root, p)
		stem := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		renderBTX(data, filepath.Join(*out, "tex", filepath.Dir(rel)), stem)
	}

	// A UI scene may be split across a base archive and language variants (the
	// screens in "Title.carc", their main-screen tiles in "Title_us.carc" — the game
	// loads both), so 2D composition merges each language archive with its base.
	byStem := map[string][]nds.NARCFile{}
	for _, p := range carcs {
		rel, _ := filepath.Rel(*root, p)
		stem := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		files, err := nds.ParseNARCFiles(data)
		if err != nil {
			nSkip++
			continue
		}
		byStem[stem] = files
		// textures (self-contained per archive)
		texDir := filepath.Join(*out, "tex", filepath.Dir(rel), stem)
		if strings.HasPrefix(rel, "data/Course") {
			texDir = filepath.Join(*out, "course", stem)
		}
		for i, f := range files {
			if len(f.Data) >= 4 && string(f.Data[:4]) == "BTX0" {
				tag := ""
				if countBTX(files) > 1 {
					tag = fmt.Sprintf("%d-", i)
				}
				renderBTX(f.Data, texDir, tag+stem)
			}
		}
	}
	langs := []string{"_us", "_es", "_fr", "_ge", "_it"}
	isLang := func(stem string) (string, bool) {
		for _, l := range langs {
			if strings.HasSuffix(stem, l) {
				return strings.TrimSuffix(stem, l), true
			}
		}
		return "", false
	}
	for stem, files := range byStem {
		if base, ok := isLang(stem); ok {
			// language variant: compose with the base archive's items (if any)
			items := narcItems(files)
			if bf, ok := byStem[base]; ok {
				items = append(items, narcItems(bf)...)
			}
			renderScreens(items, filepath.Join(*out, "ui", stem))
			continue
		}
		// base archive: skip if language variants cover it
		if _, ok := byStem[stem+"_us"]; ok {
			continue
		}
		renderScreens(narcItems(files), filepath.Join(*out, "ui", stem))
	}

	// Pass 2: loose 2D files per directory.
	var dirs []string
	for d := range dirs2d {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	for _, d := range dirs {
		rel, _ := filepath.Rel(*root, d)
		renderScreens(looseItems(dirs2d[d]), filepath.Join(*out, "ui", strings.ReplaceAll(rel, string(filepath.Separator), "-")))
	}

	// Pass 3: raw .nbfc/.nbfp pairs ("nitro basic file" character/palette — headerless
	// 4bpp tiles + BGR555 colours; Mario Kart's is its 32x32 cartridge-banner icon).
	filepath.Walk(*root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || strings.ToLower(filepath.Ext(p)) != ".nbfc" {
			return nil
		}
		chrData, e1 := os.ReadFile(p)
		palData, e2 := os.ReadFile(strings.TrimSuffix(p, filepath.Ext(p)) + ".nbfp")
		if e1 != nil || e2 != nil {
			return nil
		}
		img := renderRawTiles(chrData, palData)
		if img == nil {
			return nil
		}
		stem := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		os.MkdirAll(filepath.Join(*out, "ui"), 0o755)
		if writePNG(filepath.Join(*out, "ui", stem+".png"), img) == nil {
			nSheet++
		}
		return nil
	})

	fmt.Printf("rendered: %d textures, %d screens, %d tile sheets (%d archives skipped)\n", nTex, nScr, nSheet, nSkip)
}

func countBTX(files []nds.NARCFile) int {
	n := 0
	for _, f := range files {
		if len(f.Data) >= 4 && string(f.Data[:4]) == "BTX0" {
			n++
		}
	}
	return n
}

func renderBTX(data []byte, dir, stem string) {
	texs, err := nitro.DecodeNSBTX(data)
	if err != nil || len(texs) == 0 {
		if err != nil {
			nSkip++
		}
		return
	}
	os.MkdirAll(dir, 0o755)
	for _, t := range texs {
		if writePNG(filepath.Join(dir, stem+"-"+safe(t.Name)+".png"), t.Img) == nil {
			nTex++
		}
	}
}

// item is one 2D-format asset for screen composition.
type item struct {
	name string
	ext  string
	data []byte
}

func narcItems(files []nds.NARCFile) []item {
	var items []item
	for i, f := range files {
		name := f.Name
		if name == "" {
			name = fmt.Sprintf("file%03d", i)
		}
		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(name), "."))
		if ext == "NSCR" || ext == "NCGR" || ext == "NCLR" {
			items = append(items, item{strings.TrimSuffix(name, filepath.Ext(name)), ext, f.Data})
		}
	}
	return items
}

func looseItems(paths []string) []item {
	var items []item
	sort.Strings(paths)
	for _, p := range paths {
		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(p), "."))
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		items = append(items, item{strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)), ext, data})
	}
	return items
}

// renderScreens composes every NSCR in items (pairing tiles/palette by longest
// shared name prefix) and sheets any leftover NCGRs.
func renderScreens(items []item, dir string) {
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
	if len(scrs) == 0 && len(chrs) == 0 {
		return
	}
	made := false
	usedChr := map[string]bool{}
	for _, s := range scrs {
		chr := bestMatch(s.name, chrs)
		pal := bestMatch(s.name, pals)
		if chr == nil || pal == nil {
			continue
		}
		usedChr[chr.name] = true
		sp, e1 := nitro.ParseNSCR(s.data)
		cp, e2 := nitro.ParseNCGR(chr.data)
		pp, e3 := nitro.ParseNCLR(pal.data)
		if e1 != nil || e2 != nil || e3 != nil {
			continue
		}
		if !made {
			os.MkdirAll(dir, 0o755)
			made = true
		}
		if writePNG(filepath.Join(dir, s.name+".png"), nitro.ComposeScreen(sp, cp, pp)) == nil {
			nScr++
		}
	}
	for _, c := range chrs {
		if usedChr[c.name] {
			continue
		}
		pal := bestMatch(c.name, pals)
		if pal == nil {
			continue
		}
		cp, e1 := nitro.ParseNCGR(c.data)
		pp, e2 := nitro.ParseNCLR(pal.data)
		if e1 != nil || e2 != nil {
			continue
		}
		if !made {
			os.MkdirAll(dir, 0o755)
			made = true
		}
		if writePNG(filepath.Join(dir, c.name+"-tiles.png"), nitro.TileSheet(cp, pp, 0)) == nil {
			nSheet++
		}
	}
}

func bestMatch(name string, cands []item) *item {
	best, bestLen := -1, 0
	for i, c := range cands {
		l := 0
		for l < len(name) && l < len(c.name) && name[l] == c.name[l] {
			l++
		}
		if l > bestLen {
			best, bestLen = i, l
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

// renderRawTiles decodes a headerless 4bpp tile blob + raw BGR555 palette, laying
// the tiles out square (the 512-byte banner is 16 tiles = 32x32).
func renderRawTiles(chr, pal []byte) *image.NRGBA {
	nt := len(chr) / 32
	if nt == 0 || len(pal) < 2 {
		return nil
	}
	side := 1
	for side*side < nt {
		side++
	}
	cols := make([]color.NRGBA, len(pal)/2)
	for i := range cols {
		v := uint16(pal[i*2]) | uint16(pal[i*2+1])<<8
		exp := func(c uint16) uint8 { c &= 0x1F; return uint8(c<<3 | c>>2) }
		cols[i] = color.NRGBA{R: exp(v), G: exp(v >> 5), B: exp(v >> 10), A: 0xFF}
	}
	img := image.NewNRGBA(image.Rect(0, 0, side*8, side*8))
	for t := 0; t < nt; t++ {
		tx, ty := (t%side)*8, (t/side)*8
		for i := 0; i < 64; i++ {
			b := chr[t*32+i/2]
			idx := int(b>>uint((i%2)*4)) & 0xF
			var c color.NRGBA
			if idx != 0 && idx < len(cols) {
				c = cols[idx]
			}
			img.SetNRGBA(tx+i%8, ty+i/8, c)
		}
	}
	return img
}

func safe(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' || r == ' ' {
			return '_'
		}
		return r
	}, s)
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
