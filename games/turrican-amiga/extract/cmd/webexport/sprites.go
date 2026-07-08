// sprites.go is the sprites stage (folded ex-cmd/sprites): it writes one PNG sheet per BOB
// frame table — the resident engine's shared weapon/effect sprites and, per world, that world's
// enemy sprites drawn in the world's own palette — plus sprites/index.json indexing every sheet.
//
// A sprite is an animation: a frame table (array of pointers) whose entries are 14-byte BOB
// descriptors (draw_object_bob $603A):
//
//	+$0 bitmap ptr  +$4 mask ptr  +$8 dest modulo  +$A BLTSIZE = height<<6 | width-in-words
//
// Pixels are 4 bitplanes stored plane-major, one word narrower than BLTSIZE's width, drawn
// through the 16-colour playfield palette (plane 3 doubles as mask; colour 0 transparent). Per
// world the authoritative set is every frame table the scene AI handlers install, unioned with
// a blind descriptor-pointer scan to catch sprites no handler installs directly.
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/turrican-amiga/extract/scene"
)

const (
	residentLo = 0x10
	residentHi = 0x1B780
	gfxLo      = 0x10000 // resident sprite bitmaps live in this region
	sheetCols  = 8
	pad        = 2
)

// space is a byte slice addressed by absolute runtime address: addr `a` is at data[a-base].
type space struct {
	data []byte
	base int
}

func (s space) be32(a int) int { return int(binary.BigEndian.Uint32(s.data[a-s.base:])) }
func (s space) be16(a int) int { return int(binary.BigEndian.Uint16(s.data[a-s.base:])) }
func (s space) has(a, n int) bool {
	o := a - s.base
	return o >= 0 && o+n <= len(s.data)
}

type table struct {
	addr   int
	frames []frame
}
type frame struct{ bitmap, h, w int } // w = data width in words

// sprMeta is one sprites/index.json entry: the sheet PNG and each frame's rect in it.
type sprMeta struct {
	Src    string   `json:"src"`
	Frames [][4]int `json:"frames"`
}

// exportSprites writes the sprites/ tree (index.json + per-frame-table sheets) from the disk.
func exportSprites(game *scene.Game, outDir string) error {
	out := filepath.Join(outDir, "sprites")
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	// The authoritative sprite set is every frame table the placements actually use.
	residentObjFT := map[int]bool{}
	worldObjFT := make([]map[int]bool, scene.NumWorlds)
	for w := 0; w < scene.NumWorlds; w++ {
		worldObjFT[w] = map[int]bool{}
		for _, sc := range game.Scenes(w) {
			for _, o := range sc.Objects {
				if o.FT == 0 {
					continue
				}
				if o.Resident {
					residentObjFT[o.FT] = true
				} else {
					worldObjFT[w][o.FT] = true
				}
			}
		}
	}

	index := map[string]sprMeta{}
	total := 0

	// Resident shared sprites (weapons/effects + placement-used resident objects), world-0 palette.
	resident := space{data: game.Resident.Data, base: 0}
	pal0 := game.WorldPalette(0, true)
	residTables := map[int]table{}
	for _, t := range findTables(resident, residentLo, residentHi, gfxLo, residentHi) {
		residTables[t.addr] = t
	}
	for ft := range residentObjFT {
		if _, ok := residTables[ft]; ok {
			continue
		}
		if t, ok := tableAt(resident, ft, gfxLo, residentHi); ok {
			residTables[ft] = t
		}
	}
	for _, t := range sortTables(residTables) {
		name := fmt.Sprintf("resident_%05X.png", t.addr)
		frames, err := emitSheet(filepath.Join(out, name), resident, pal0, t)
		if err != nil {
			return err
		}
		index[fmt.Sprintf("resident/%05X", t.addr)] = sprMeta{Src: "sprites/" + name, Frames: frames}
		total++
	}
	fmt.Fprintf(os.Stderr, "[sprites] resident: %d frame tables\n", len(residTables))

	// Per-world enemy sprites from each scene block, in the world's own palette.
	for w := 0; w < scene.NumWorlds; w++ {
		block := game.Block(w).Data
		sp := space{data: block, base: blockBase}
		hi := blockBase + len(block)

		tables := map[int]table{}
		for _, t := range findTables(sp, blockBase, hi, blockBase, hi) {
			tables[t.addr] = t
		}
		for ft := range worldObjFT[w] {
			if _, ok := tables[ft]; ok {
				continue
			}
			if t, ok := tableAt(sp, ft, blockBase, hi); ok {
				tables[t.addr] = t
			}
		}
		pal := game.WorldPalette(w, true)
		for _, t := range sortTables(tables) {
			name := fmt.Sprintf("world%d_%05X.png", w, t.addr)
			frames, err := emitSheet(filepath.Join(out, name), sp, pal, t)
			if err != nil {
				return err
			}
			index[fmt.Sprintf("w%d/%05X", w, t.addr)] = sprMeta{Src: "sprites/" + name, Frames: frames}
			total++
		}
		fmt.Fprintf(os.Stderr, "[sprites] world %d: %d frame tables (%d so far)\n", w, len(tables), total)
	}

	if err := writeJSON(filepath.Join(out, "index.json"), index); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[sprites] done: %d sheets + index.json\n", total)
	return nil
}

// emitSheet renders t's frames into a grid sheet PNG and returns each frame's rect in it.
func emitSheet(path string, sp space, pal color.Palette, t table) ([][4]int, error) {
	sheet, rects := renderSheet(sp, pal, t)
	if _, err := writePNG(path, sheet); err != nil {
		return nil, err
	}
	return rects, nil
}

// tableAt reads the frame table at addr: descriptor pointers until one fails to resolve (or a
// sane cap). Unlike findTables it accepts a single-frame table.
func tableAt(sp space, addr, gfxLo, gfxHi int) (table, bool) {
	var frames []frame
	for a := addr; sp.has(a, 4) && len(frames) < 64; a += 4 {
		f, ok := descAt(sp, sp.be32(a), gfxLo, gfxHi, 1)
		if !ok {
			break
		}
		frames = append(frames, f)
	}
	if len(frames) == 0 {
		return table{}, false
	}
	return table{addr: addr, frames: frames}, true
}

// sortTables flattens the map into address order for deterministic output.
func sortTables(m map[int]table) []table {
	out := make([]table, 0, len(m))
	for _, t := range m {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].addr < out[j].addr })
	return out
}

// descAt decodes the 14-byte BOB descriptor pointed to by p, validating its bitmap lies in
// [gfxLo,gfxHi) and its dimensions are sane. minH is the smallest accepted frame height.
func descAt(sp space, p, gfxLo, gfxHi, minH int) (frame, bool) {
	if p < sp.base || !sp.has(p, 14) {
		return frame{}, false
	}
	bm := sp.be32(p)
	bs := sp.be16(p + 0xA)
	h, wd := bs>>6, bs&0x3F
	if bm < gfxLo || bm >= gfxHi || !sp.has(bm, 4*h*(wd-1)*2) || h < minH || h > 96 || wd < 2 || wd > 12 {
		return frame{}, false
	}
	return frame{bitmap: bm, h: h, w: wd - 1}, true
}

// findTables scans [scanLo,scanHi) for runs of >=3 pointers that all resolve to a plausible BOB.
func findTables(sp space, scanLo, scanHi, gfxLo, gfxHi int) []table {
	var out []table
	for a := scanLo; a < scanHi-4; {
		if f0, ok := descAt(sp, sp.be32(a), gfxLo, gfxHi, 4); ok {
			frames := []frame{f0}
			j := a + 4
			for j < scanHi-4 {
				f, ok := descAt(sp, sp.be32(j), gfxLo, gfxHi, 4)
				if !ok {
					break
				}
				frames = append(frames, f)
				j += 4
			}
			if len(frames) >= 3 {
				out = append(out, table{addr: a, frames: frames})
				a = j
				continue
			}
		}
		a += 2 // tables are word- but not always long-aligned
	}
	return out
}

// renderSheet packs t's frames into a grid PNG and returns each frame's rect [x,y,w,h].
func renderSheet(sp space, pal color.Palette, t table) (*image.Paletted, [][4]int) {
	cw, ch := 0, 0
	for _, f := range t.frames {
		if f.w*16 > cw {
			cw = f.w * 16
		}
		if f.h > ch {
			ch = f.h
		}
	}
	cw += pad
	ch += pad
	rows := (len(t.frames) + sheetCols - 1) / sheetCols
	sheet := image.NewPaletted(image.Rect(0, 0, sheetCols*cw, rows*ch), pal)
	rects := make([][4]int, len(t.frames))
	for i, f := range t.frames {
		ox, oy := (i%sheetCols)*cw, (i/sheetCols)*ch
		drawBob(sheet, sp, f, ox, oy)
		rects[i] = [4]int{ox, oy, f.w * 16, f.h}
	}
	return sheet, rects
}

// drawBob decodes one 4-bitplane plane-major BOB into the sheet at (ox,oy).
func drawBob(dst *image.Paletted, sp space, f frame, ox, oy int) {
	bpr := f.w * 2
	planeSize := f.h * bpr
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w*16; x++ {
			var v uint8
			for p := 0; p < 4; p++ {
				a := f.bitmap + p*planeSize + y*bpr + x/8
				if sp.has(a, 1) && sp.data[a-sp.base]&(0x80>>(x%8)) != 0 {
					v |= 1 << uint(p)
				}
			}
			if v != 0 { // colour 0 transparent
				dst.SetColorIndex(ox+x, oy+y, v)
			}
		}
	}
}
