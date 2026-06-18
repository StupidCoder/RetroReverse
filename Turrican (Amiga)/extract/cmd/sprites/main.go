// Command sprites extracts Turrican's BOB (blitter-object) sprites and writes one
// PNG sheet per sprite, laying out all of its animation frames (Part V).
//
//	sprites [-o dir] [Turrican.adf]
//
// Two sets are written: the resident engine's shared weapon/effect sprites
// (sprite_<addr>.png), and, for each of the five worlds, that world's enemy
// sprites from its decoded scene block (world<N>_sprite_<addr>.png), drawn in the
// world's own palette.
//
// A sprite is an animation: a frame table (an array of pointers) whose entries are
// 14-byte BOB descriptors that draw_object_bob ($603A) reads:
//
//	+$0 bitmap data ptr   +$4 mask ptr   +$8 dest modulo
//	+$A BLTSIZE = height<<6 | width-in-words   +$C y-adjust   +$D flag
//
// The pixels are 4 bitplanes stored plane-major, one word narrower than BLTSIZE's
// width (the cookie-cut shift reads an extra word), so a frame is
// 4 * height * (width-1)*2 bytes, drawn through the 16-colour playfield palette
// (plane 3 doubles as the mask, so opaque pixels use colours 8-15; colour 0 is
// transparent).
//
// Per world the authoritative sprite set is the frame tables the scene's enemy-AI
// handlers install: each handler does `MOVE.l #frametable,$12(a5)`, so we collect
// every such table from the +$20 AI handler tables (this is exactly the set the
// placement viewer needs to look up). A blind scan for runs of descriptor pointers
// is unioned in to catch sprites no handler installs directly.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"turrican/extract/decrunch"
	"turrican/extract/scene"
)

const (
	residentLo = 0x10
	residentHi = 0x1B780
	gfxLo      = 0x10000 // resident sprite bitmaps live in this region
	blockBase  = 0x1B980 // a scene block's runtime load address
	levelTable = 0x46A
	numWorlds  = 5
	sheetCols  = 8
	pad        = 2
)

// space is a byte slice addressed by absolute runtime address: addr `a` is at
// data[a-base].
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

func main() {
	out := flag.String("o", "rendered/sprites", "output directory")
	flag.Parse()
	adfPath := flag.Arg(0)
	if adfPath == "" {
		adfPath = "Turrican.adf"
	}
	adf, err := os.ReadFile(adfPath)
	if err != nil {
		fail(err)
	}
	res, err := decrunch.Decrunch(mainBlob(adf))
	if err != nil {
		fail(err)
	}
	game, err := scene.Load(adf)
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}

	// The authoritative sprite set is every frame table the placements actually use
	// (resolved via the shared spawner reimplementation): resident handlers ($1A60,
	// types 1-2) install resident sprites; scene handlers (+$20) install scene-block
	// sprites. Collect them so every placement-referenced sheet is produced.
	residentObjFT := map[int]bool{}
	worldObjFT := make([]map[int]bool, numWorlds)
	for w := 0; w < numWorlds; w++ {
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

	// Resident shared sprites (weapons/effects + the placement-used resident
	// objects), in world 0's palette.
	resident := space{data: res.Data, base: 0} // addr == file offset
	pal0 := worldPalette(adf, 0)
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
	emit(*out, "sprite_%05X.png", resident, pal0, sortTables(residTables))

	// Per-world enemy sprites from each scene block, in the world's own palette:
	// a blind descriptor-pointer scan unioned with the placement-used frame tables.
	for w := 0; w < numWorlds; w++ {
		block := worldBlock(adf, res.Data, w)
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
		emit(*out, fmt.Sprintf("world%d_sprite_%%05X.png", w), sp, worldPalette(adf, w), sortTables(tables))
	}
}

// tableAt reads the frame table at addr: descriptor pointers until one fails to
// resolve (or a sane cap). Unlike findTables it accepts a single-frame table.
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

func emit(dir, nameFmt string, sp space, pal color.Palette, tables []table) {
	for _, t := range tables {
		name := fmt.Sprintf(nameFmt, t.addr)
		if err := writePNG(filepath.Join(dir, name), renderSheet(sp, pal, t)); err != nil {
			fail(err)
		}
		fmt.Printf("$%05X: %2d frames -> %s\n", t.addr, len(t.frames), name)
	}
}

// descAt decodes the 14-byte BOB descriptor pointed to by p, validating that its
// bitmap lies in [gfxLo,gfxHi) and its dimensions are sane. minH is the smallest
// accepted frame height: the blind scan uses 4 to avoid false runs, but a table we
// already trust (AI-referenced) may legitimately start with a tiny frame.
func descAt(sp space, p, gfxLo, gfxHi, minH int) (frame, bool) {
	if p < sp.base || !sp.has(p, 14) {
		return frame{}, false
	}
	bm := sp.be32(p)
	bs := sp.be16(p + 0xA)
	h, w := bs>>6, bs&0x3F
	if bm < gfxLo || bm >= gfxHi || !sp.has(bm, 4*h*(w-1)*2) || h < minH || h > 96 || w < 2 || w > 12 {
		return frame{}, false
	}
	return frame{bitmap: bm, h: h, w: w - 1}, true
}

// findTables scans [scanLo,scanHi) for runs of >=3 pointers that all resolve to a
// plausible BOB descriptor (bitmap in [gfxLo,gfxHi)).
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

func renderSheet(sp space, pal color.Palette, t table) *image.Paletted {
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
	for i, f := range t.frames {
		drawBob(sheet, sp, f, (i%sheetCols)*cw, (i/sheetCols)*ch)
	}
	return sheet
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

// worldBlock decodes world w's scene block from the disk.
func worldBlock(adf, img []byte, w int) []byte {
	t := levelTable + w*8
	o := int(binary.BigEndian.Uint32(img[t:]))
	n := int(binary.BigEndian.Uint32(img[t+4:]))
	block, err := decrunch.DecrunchBlock(adf[o : o+n])
	if err != nil {
		fail(err)
	}
	return block
}

// worldPalette reads world w's 16-colour playfield palette (index 0 transparent).
func worldPalette(adf []byte, w int) color.Palette {
	res, err := decrunch.Decrunch(mainBlob(adf))
	if err != nil {
		fail(err)
	}
	block := worldBlock(adf, res.Data, w)
	palOff := int(binary.BigEndian.Uint32(block[8:])) - blockBase
	pal := color.Palette{color.RGBA{0, 0, 0, 0}}
	for i := 1; i < 16; i++ {
		c := binary.BigEndian.Uint16(block[palOff+i*2:])
		pal = append(pal, color.RGBA{
			R: uint8((c>>8)&0xF) * 17, G: uint8((c>>4)&0xF) * 17, B: uint8(c&0xF) * 17, A: 255,
		})
	}
	return pal
}

func mainBlob(adf []byte) []byte {
	const off = 0x2C00
	return adf[off : off+int(binary.BigEndian.Uint32(adf[off:]))]
}
func writePNG(path string, im image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, im)
}
func fail(err error) {
	fmt.Fprintln(os.Stderr, "sprites:", err)
	os.Exit(1)
}
