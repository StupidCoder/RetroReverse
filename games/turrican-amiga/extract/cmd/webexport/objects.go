// objects.go is the objects stage: every BOB frame table becomes a sprite2d
// object — the resident engine's shared weapon/effect sprites (world-0
// palette) and, per world, that world's enemy sprites in its own palette.
//
// A sprite is an animation: a frame table (array of pointers) whose entries
// are 14-byte BOB descriptors (draw_object_bob $603A):
//
//	+$0 bitmap ptr  +$4 mask ptr  +$8 dest modulo  +$A BLTSIZE = height<<6 | width-in-words
//
// Pixels are 4 bitplanes stored plane-major, one word narrower than BLTSIZE's
// width, drawn through the 16-colour playfield palette (plane 3 doubles as
// mask; colour 0 transparent). Per world the authoritative set is every frame
// table the scene AI handlers install, unioned with a blind descriptor-pointer
// scan to catch sprites no handler installs directly.
//
// Each table's frames pack into a ONE-ROW strip so the flip-book plays as the
// "main" animation and each placement-used frame is also exposed as a still
// ("f<k>") the level placements select by orientation.
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sort"

	"retroreverse.com/games/turrican-amiga/extract/scene"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

const (
	residentLo = 0x10
	residentHi = 0x1B780
	gfxLo      = 0x10000 // resident sprite bitmaps live in this region
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

// exportObjects writes the sprite2d assets and returns the frame-table key
// ("resident/1B3C4" / "w<w>/<addr>") -> asset map.
func exportObjects(ctx *cli.Context, game *scene.Game, scenes [][]scene.Scene) (map[string]objRef, error) {
	refs := map[string]objRef{}

	// Which frame tables (and which of their frames) the placements use: the
	// used frames become still animations placements can select.
	residentFT := map[int]map[int]bool{}
	worldFT := make([]map[int]map[int]bool, scene.NumWorlds)
	note := func(m map[int]map[int]bool, ft, fr int) {
		if m[ft] == nil {
			m[ft] = map[int]bool{}
		}
		m[ft][fr] = true
	}
	for w := range scenes {
		worldFT[w] = map[int]map[int]bool{}
		for _, sc := range scenes[w] {
			for _, o := range sc.Objects {
				if o.FT == 0 {
					continue
				}
				if o.Resident {
					note(residentFT, o.FT, o.Frame)
				} else {
					note(worldFT[w], o.FT, o.Frame)
				}
			}
		}
	}

	// Resident shared sprites (weapons/effects + placement-used resident objects), world-0 palette.
	resident := space{data: game.Resident.Data, base: 0}
	pal0 := game.WorldPalette(0, true)
	residTables := map[int]table{}
	for _, t := range findTables(resident, residentLo, residentHi, gfxLo, residentHi) {
		residTables[t.addr] = t
	}
	for ft := range residentFT {
		if _, ok := residTables[ft]; ok {
			continue
		}
		if t, ok := tableAt(resident, ft, gfxLo, residentHi); ok {
			residTables[ft] = t
		}
	}
	total := 0
	for _, t := range sortTables(residTables) {
		key := fmt.Sprintf("resident/%05X", t.addr)
		id := fmt.Sprintf("resident-%05x", t.addr)
		name := fmt.Sprintf("Sprite $%05X", t.addr)
		if err := emitTable(ctx, id, name, "Resident", resident, pal0, t, residentFT[t.addr]); err != nil {
			return nil, err
		}
		refs[key] = objRef{asset: id}
		total++
	}
	ctx.Progress("objects", total, total, fmt.Sprintf("resident: %d frame tables", len(residTables)))

	// Per-world enemy sprites from each scene block, in the world's own palette.
	for w := 0; w < scene.NumWorlds; w++ {
		block := game.Block(w).Data
		sp := space{data: block, base: blockBase}
		hi := blockBase + len(block)

		tables := map[int]table{}
		for _, t := range findTables(sp, blockBase, hi, blockBase, hi) {
			tables[t.addr] = t
		}
		for ft := range worldFT[w] {
			if _, ok := tables[ft]; ok {
				continue
			}
			if t, ok := tableAt(sp, ft, blockBase, hi); ok {
				tables[t.addr] = t
			}
		}
		pal := game.WorldPalette(w, true)
		for _, t := range sortTables(tables) {
			key := fmt.Sprintf("w%d/%05X", w, t.addr)
			id := fmt.Sprintf("w%d-%05x", w, t.addr)
			name := fmt.Sprintf("Sprite $%05X (World %d)", t.addr, w+1)
			if err := emitTable(ctx, id, name, fmt.Sprintf("World %d", w+1), sp, pal, t, worldFT[w][t.addr]); err != nil {
				return nil, err
			}
			refs[key] = objRef{asset: id}
			total++
		}
		ctx.Progress("objects", total, total, fmt.Sprintf("world %d: %d frame tables", w+1, len(tables)))
	}
	return refs, nil
}

// emitTable renders t's frames into a one-row strip and registers the
// sprite2d object: "main" flips through every frame, and each
// placement-used frame k gets a still animation "f<k>".
func emitTable(ctx *cli.Context, id, name, group string, sp space, pal color.Palette, t table, usedFrames map[int]bool) error {
	cw, ch := 0, 0
	for _, f := range t.frames {
		if f.w*16 > cw {
			cw = f.w * 16
		}
		if f.h > ch {
			ch = f.h
		}
	}
	strip := image.NewPaletted(image.Rect(0, 0, cw*len(t.frames), ch), pal)
	for i, f := range t.frames {
		drawBob(strip, sp, f, i*cw, 0)
	}
	f, err := ctx.Builder.CreateFile("objects", id+".png")
	if err != nil {
		return err
	}
	err = png.Encode(f, strip)
	f.Close()
	if err != nil {
		return err
	}

	anims := []schema.Animation{{
		ID: "main", Frames: len(t.frames), Loop: "loop",
		Durations: repeat(8, len(t.frames)), // ~6 fps flip-book through the table
	}}
	if len(t.frames) == 1 {
		anims[0].Loop = "hold"
		anims[0].Durations = nil
	}
	var used []int
	for k := range usedFrames {
		if k >= 0 && k < len(t.frames) {
			used = append(used, k)
		}
	}
	sort.Ints(used)
	for _, k := range used {
		anims = append(anims, schema.Animation{
			// a still: one step holding frame k (Frames spans the row so the
			// step index validates)
			ID: fmt.Sprintf("f%d", k), Frames: len(t.frames), Loop: "hold", Steps: [][]int{{k, 1}},
		})
	}

	ctx.Builder.AddObject(schema.Asset{ID: id, Name: name, Group: group}, &schema.Object{
		Type:       schema.ObjectSprite2D,
		Name:       name,
		Atlas:      &schema.SpriteAtlas{File: id + ".png", CellW: cw, CellH: ch},
		Animations: anims,
		Props: map[string]any{
			"frameTable": fmt.Sprintf("$%05X", t.addr),
			"frames":     len(t.frames),
		},
	})
	return nil
}

func repeat(v, n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = v
	}
	return out
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
