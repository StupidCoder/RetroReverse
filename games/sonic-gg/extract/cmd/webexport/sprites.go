// sprites.go folds the object/enemy sprite ripper (ex-cmd/spriterip) into webexport: it
// writes sprites/index.json + sprites/<zone>/<hex>.png straight from the ROM (no oracle).
//
// An object's on-screen sprite is a metasprite: a 3-row x 6-col grid of 8x16-sprite tile
// indices, walked by the sprite writer $2F07. An animated object selects its frame through
// the shared animation routine $7C75 as layoutBase + frameId*18. The referenced tiles live
// in the per-zone sprite tile set the level loader decompresses to VRAM $2000 (descriptor
// +23/+24/+25, $0406 codec), coloured by the sprite palette (descriptor +26 -> $7400).
//
// (chk, w, romPalette, descTable, palTable are shared with main.go.)
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/sonic-gg/extract/decomp"
	"retroreverse.com/games/sonic-gg/extract/objplace"
	"retroreverse.com/tools/platform/gamegear"
)

// sprMeta is one common-format sprite entry (site/FORMAT.md): explicit frame rects into the
// strip PNG, per-frame durations or an explicit play sequence, and the movement path for the
// moving platforms — keyed "<zone>/<hex>".
type sprMeta struct {
	Src       string   `json:"src"`
	Frames    [][4]int `json:"frames"`
	Durations []int    `json:"durations,omitempty"`
	Steps     [][2]int `json:"steps,omitempty"`
	Path      [][2]int `json:"path,omitempty"`
}

// spriteTiles builds act i's full sprite sheet (objplace.SpriteSheet: the zone's own tiles
// $00-$7F + the common block $80-$BF) and resolves its sprite palette.
func spriteTiles(rom []byte, act int) ([]byte, color.Palette) {
	d := descTable + w(rom, descTable+act*2)
	return objplace.SpriteSheet(rom, act), romPalette(rom, int(rom[d+26]))
}

// Sonic's sprite is not in the zone sheets: his handler keeps one layout ($5C1B, his 16x32
// box) and re-streams the tile GRAPHICS per pose (bank 8 + frame*192, 3bpp — $4E9A). His
// animations are per-frame byte sequences in the $5C5B table ($4E6D sequencer); the exported
// strip plays the idle -> bored program. objplace.SonicSeq/SonicFrameTiles supply both.
const sonicLayout = 0x5C1B

// sonicStrip renders Sonic's idle/bored strip: unique graphic frames plus the play sequence
// (strip-frame index, hold frames).
func sonicStrip(rom []byte, pal color.Palette) (*image.RGBA, [][2]int) {
	bored, loopStep := objplace.SonicSeq(rom, 0x0D)
	gfx := []int{0}
	idxOf := map[int]int{0: 0}
	for _, st := range bored {
		if _, ok := idxOf[st.Layout]; !ok {
			idxOf[st.Layout] = len(gfx)
			gfx = append(gfx, st.Layout)
		}
	}
	strip := image.NewRGBA(image.Rect(0, 0, len(gfx)*48, 48))
	for i, f := range gfx {
		cell := renderMeta(rom[sonicLayout:sonicLayout+18], objplace.SonicFrameTiles(rom, f), pal)
		for y := 0; y < 48; y++ {
			for x := 0; x < 48; x++ {
				strip.Set(i*48+x, y, cell.At(x, y))
			}
		}
	}
	seq := [][2]int{{0, 360}}
	for _, st := range bored {
		seq = append(seq, [2]int{idxOf[st.Layout], st.Frames})
	}
	for cycle := 0; cycle < 4; cycle++ {
		for _, st := range bored[loopStep:] {
			seq = append(seq, [2]int{idxOf[st.Layout], st.Frames})
		}
	}
	return strip, seq
}

// renderMeta renders one 18-byte metasprite layout (3 rows x 6 cols of 8x16 sprites) into a
// 48x48 RGBA image; colour index 0 of the sprite palette is transparent.
func renderMeta(layout, tiles []byte, pal color.Palette) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 6*8, 3*16))
	p := 0
	for row := 0; row < 3; row++ {
		if p >= len(layout) || layout[p] == 0xFF {
			break
		}
		for col := 0; col < 6; col++ {
			b := layout[p]
			p++
			if b >= 0xFE { // $FE skip, $FF treated as skip mid-row
				continue
			}
			for half := 0; half < 2; half++ { // 8x16 sprite = tiles b (top), b+1 (bottom)
				ti := (int(b) + half) * 32
				if ti+32 > len(tiles) {
					continue
				}
				t := gamegear.DecodeTile(tiles[ti:])
				ox, oy := col*8, row*16+half*8
				for y := 0; y < 8; y++ {
					for x := 0; x < 8; x++ {
						if v := t[y][x]; v != 0 {
							img.Set(ox+x, oy+y, pal[v])
						}
					}
				}
			}
		}
	}
	return img
}

// drawCell blits one 8x16 sprite cell (tiles t, t+1) into img at (ox, oy).
func drawCell(img *image.RGBA, tiles []byte, pal color.Palette, t, ox, oy int) {
	for half := 0; half < 2; half++ {
		ti := (t + half) * 32
		if ti+32 > len(tiles) {
			continue
		}
		cell := gamegear.DecodeTile(tiles[ti:])
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				if v := cell[y][x]; v != 0 {
					img.Set(ox+x, oy+half*8+y, pal[v])
				}
			}
		}
	}
}

// renderAnimStrip renders a type's full animation as a horizontal strip of 48x48 frames
// (objplace.Anim), compositing the pickup family's screen-icon overlay (drawn on top).
func renderAnimStrip(rom, tiles []byte, pal color.Palette, typ, zone int) (*image.RGBA, []int) {
	frames := objplace.Anim(rom, typ, zone)
	if len(frames) == 0 {
		return nil, nil
	}
	icons := objplace.IconOverlays(rom, typ)
	strip := image.NewRGBA(image.Rect(0, 0, len(frames)*48, 48))
	durs := make([]int, len(frames))
	for i, f := range frames {
		if f.Layout <= 0 || f.Layout+18 > len(rom) {
			return nil, nil
		}
		cell := renderMeta(rom[f.Layout:f.Layout+18], tiles, pal)
		for _, ic := range icons {
			drawCell(cell, tiles, pal, ic.Tile, ic.X, ic.Y)
		}
		for y := 0; y < 48; y++ {
			for x := 0; x < 48; x++ {
				strip.Set(i*48+x, y, cell.At(x, y))
			}
		}
		durs[i] = f.Frames
	}
	// Collapse animations whose frames are all pixel-identical: the pickup TV's layout blink
	// alternates a screen cell that the opaque icon fully covers — a sprite-per-scanline budget
	// trick, not a visible animation.
	if len(frames) > 1 {
		same := true
		for i := 1; same && i < len(frames); i++ {
			for y := 0; same && y < 48; y++ {
				for x := 0; x < 48; x++ {
					if strip.RGBAAt(i*48+x, y) != strip.RGBAAt(x, y) {
						same = false
						break
					}
				}
			}
		}
		if same {
			one := image.NewRGBA(image.Rect(0, 0, 48, 48))
			for y := 0; y < 48; y++ {
				for x := 0; x < 48; x++ {
					one.Set(x, y, strip.At(x, y))
				}
			}
			return one, []int{0}
		}
	}
	return strip, durs
}

// goalStrip renders the goal sign's idle spin from its own graphics: the handler decompresses
// its sheet over VRAM $2000 and switches to sprite palette $0E (objplace.GoalAnim).
func goalStrip(rom []byte) (*image.RGBA, [][2]int) {
	steps, src, palIdx := objplace.GoalAnim(rom)
	tiles := make([]byte, 256*32)
	copy(tiles, decomp.Decompress(rom, src))
	copy(tiles[0x80*32:], decomp.Decompress(rom, objplace.CommonTilesFile))
	pal := romPalette(rom, palIdx)
	var gfx []int
	idxOf := map[int]int{}
	for _, st := range steps {
		if _, ok := idxOf[st.Layout]; !ok {
			idxOf[st.Layout] = len(gfx)
			gfx = append(gfx, st.Layout)
		}
	}
	strip := image.NewRGBA(image.Rect(0, 0, len(gfx)*48, 48))
	for i, lay := range gfx {
		cell := renderMeta(rom[lay:lay+18], tiles, pal)
		for y := 0; y < 48; y++ {
			for x := 0; x < 48; x++ {
				strip.Set(i*48+x, y, cell.At(x, y))
			}
		}
	}
	var seq [][2]int
	for _, st := range steps {
		seq = append(seq, [2]int{idxOf[st.Layout], st.Frames})
	}
	return strip, seq
}

// trimBBox crops an RGBA to its non-transparent bounding box and returns the box's top-left
// offset within the source; used here only to detect empty layouts.
func trimBBox(src *image.RGBA) (int, int, *image.RGBA) {
	b := src.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X, b.Min.Y
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, al := src.At(x, y).RGBA(); al != 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if minX > maxX {
		return 0, 0, image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	out := image.NewRGBA(image.Rect(0, 0, maxX-minX+1, maxY-minY+1))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			out.Set(x-minX, y-minY, src.At(x, y))
		}
	}
	return minX, minY, out
}

// placedTypesByZone returns the sorted distinct object types placed in each zone's acts
// (zones 0-5 = acts z*3..z*3+2; zone 6 = the special-stage acts 28-35).
func placedTypesByZone(rom []byte) [][]int {
	actObjTypes := func(act int) []int {
		d := descTable + w(rom, descTable+act*2)
		ot := descTable + w(rom, d+30)
		n := int(rom[ot])
		var ts []int
		for k, p := 0, ot+1; k < n; k, p = k+1, p+3 {
			ts = append(ts, int(rom[p]))
		}
		return ts
	}
	acts := func(z int) []int {
		if z == 6 {
			return []int{28, 29, 30, 31, 32, 33, 34, 35}
		}
		return []int{z * 3, z*3 + 1, z*3 + 2}
	}
	out := make([][]int, 7)
	for z := 0; z < 7; z++ {
		seen := map[int]bool{}
		for _, act := range acts(z) {
			for _, t := range actObjTypes(act) {
				if t < 0x57 && !seen[t] {
					seen[t] = true
					out[z] = append(out[z], t)
				}
			}
		}
		sort.Ints(out[z])
	}
	return out
}

// exportSprites writes the sprites/ tree (index.json + <zone>/<hex>.png) from the ROM. No
// oracle. index.json is a map (json.Marshal sorts keys) so output is deterministic.
func exportSprites(rom []byte, outdir string) {
	out := filepath.Join(outdir, "sprites")
	chk(os.MkdirAll(out, 0o755))

	// representative act (and thus sprite sheet/palette) for each of the 7 viewer zones.
	zoneAct := []int{0, 3, 6, 9, 12, 15, 28}
	zoneName := []string{"greenhills", "bridge", "jungle", "labyrinth", "scrapbrain", "skybase", "special"}

	// Each PNG is the FULL 48x48 metasprite grid; the viewer loads sprites/<zone>/<hex>.png and
	// draws it with its top-left at the object's world position (blockX*32, blockY*32).
	rects := func(n int) [][4]int {
		r := make([][4]int, n)
		for i := range r {
			r[i] = [4]int{i * 48, 0, 48, 48}
		}
		return r
	}
	index := map[string]sprMeta{}
	paths := objplace.PlatformPaths(rom)
	placed := placedTypesByZone(rom)
	total := 0
	for z, act := range zoneAct {
		tiles, pal := spriteTiles(rom, act)
		zdir := fmt.Sprintf("%s/%d", out, z)
		chk(os.MkdirAll(zdir, 0o755))

		// Sonic (type $00): his own ROM tiles, this zone's sprite palette.
		sonic, seq := sonicStrip(rom, pal)
		writePNG(fmt.Sprintf("%s/00.png", zdir), sonic)
		index[fmt.Sprintf("%d/00", z)] = sprMeta{
			Src: fmt.Sprintf("sprites/%d/00.png", z), Frames: rects(sonic.Rect.Dx() / 48), Steps: seq,
		}
		total++
		for _, t := range placed[z] {
			if t == 0 {
				continue
			}
			if t == 0x07 { // the goal sign: own gfx but a plain metasprite — render its spin
				strip, seq := goalStrip(rom)
				writePNG(fmt.Sprintf("%s/%02x.png", zdir, t), strip)
				index[fmt.Sprintf("%d/07", z)] = sprMeta{
					Src: fmt.Sprintf("sprites/%d/07.png", z), Frames: rects(strip.Rect.Dx() / 48), Steps: seq,
				}
				total++
				continue
			}
			if t == 0x29 { // the floating log: 3 table-driven roll layouts ($803A+)
				lays := objplace.LogAnim()
				strip := image.NewRGBA(image.Rect(0, 0, len(lays)*48, 48))
				for i, lay := range lays {
					cell := renderMeta(rom[lay:lay+18], tiles, pal)
					for y := 0; y < 48; y++ {
						for x := 0; x < 48; x++ {
							strip.Set(i*48+x, y, cell.At(x, y))
						}
					}
				}
				seq := [][2]int{{0, 240}}
				for c := 0; c < 2; c++ {
					seq = append(seq, [2]int{0, 6}, [2]int{1, 6}, [2]int{2, 6})
				}
				writePNG(fmt.Sprintf("%s/%02x.png", zdir, t), strip)
				index[fmt.Sprintf("%d/29", z)] = sprMeta{
					Src: fmt.Sprintf("sprites/%d/29.png", z), Frames: rects(len(lays)),
					Steps: seq, Path: paths["29"],
				}
				total++
				continue
			}
			r := objplace.AnalyzeSprite(rom, t, z)
			if r.Kind == "" || r.Layout == 0 || r.Layout+18 > len(rom) {
				continue
			}
			// The bosses are self-contained set-pieces: they decompress their OWN graphics over
			// the zone sprite sheet and draw multi-part sprites from a bytecode script — a single
			// extracted metasprite renders as garbage. Skip them; the viewer shows a marker.
			if _, _, own := objplace.OwnGfx(rom, t); own || t == 0x25 {
				continue
			}
			tt := objplace.ApplyIconUpload(rom, tiles, t)
			strip, durs := renderAnimStrip(rom, tt, pal, t, z)
			if strip == nil {
				continue
			}
			if _, _, bb := trimBBox(strip); bb.Rect.Dx() <= 1 && bb.Rect.Dy() <= 1 {
				continue // empty layout
			}
			writePNG(fmt.Sprintf("%s/%02x.png", zdir, t), strip)
			entry := sprMeta{
				Src: fmt.Sprintf("sprites/%d/%02x.png", z, t), Frames: rects(strip.Rect.Dx() / 48),
				Path: paths[fmt.Sprintf("%02x", t)],
			}
			if len(entry.Frames) > 1 {
				entry.Durations = durs
			}
			index[fmt.Sprintf("%d/%02x", z, t)] = entry
			total++
		}
		fmt.Fprintf(os.Stderr, "[sprites] %d/%d  %-11s act%2d  %d sprites so far\n", z+1, len(zoneAct), zoneName[z], act, total)
	}
	writeJSON(out+"/index.json", index)
	fmt.Fprintf(os.Stderr, "[sprites] done: %d sprites + index.json\n", total)
}
