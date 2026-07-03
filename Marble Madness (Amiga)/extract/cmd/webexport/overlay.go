// Overlay extraction: the scenery sprite pieces the engine layers OVER the
// tilemap — the course's "occlusion layer".
//
// Marble Madness draws every moving or layered thing through one depth-sorted
// display list (painter's algorithm; insert $11B34, iso-box builder $1085C,
// separating-axis comparator $10F4C, renderer $11704 in the decrypted .dat).
// The static .mlb tilemap is pure background; anything the marble can roll
// BEHIND is an .ilb scenery cell attached to a dynamic Track region (display
// types 7/8/9), placed at the region's keyframe reference point and drawn in
// depth order — nearer pieces blit over the marble, which is the occlusion the
// player sees.
//
// Data chain (all in the course Track file): header +$14 -> 6-byte
// [x][y][scriptPtr] region records -> region script: op0 KEYFRAME carries the
// anchor (X,Y,z) and (when dur==1) the sprite-list link; op12/op13 re-link;
// op2 SPRITE selects a list entry. A list is a $FFFFFFFF-terminated array of
// pointers (single, or [record, composite-list] pairs) to 16-byte sprite
// records: [+0 dx][+1 dy][+2..3 state bytes][+4..7 iso-box extents]
// [+8 long .ilb cell index][+C long aux ptr].
//
// Screen projection (the engine's own math, $6918 + $122AC + $E6B2):
//   x8 = X*8+4, y8 = Y*8+4                     (eighth-cell seeds $6C4/$6C6)
//   worldX = (y8 - x8) + $88 + dx
//   worldY = (x8 + y8)/2 - z - base + $9C - dy (top of the blit)
// where base = the course descriptor's word +$10 (Track header+0 -> $9A6;
// course 4/aerial adds word +$12), the value the engine seeds the world-Y
// base $9BA with; the scroll terms cancel out of the world-space form.
package main

import (
	"fmt"
	"image"
	"image/draw"
	"path/filepath"
	"sort"

	"marblemad/extract/ilb"
	"marblemad/extract/mlb"
	"retroreverse.com/tools/amiga/adf"
	"retroreverse.com/tools/c64/gfx"
)

// exportOverlays extracts the course's scenery-overlay pieces, renders each
// referenced .ilb cell once into spritesDir (recorded in spriteIndex, keyed
// "<course>/c<cell>" or "<course>/a<region>" for animated strips), and returns
// the level's objects entries plus its screen-swap cellAnims.
func exportOverlays(vol *adf.Volume, paths map[string]string, key string, track []byte,
	co *mlb.Course, spritesDir string, spriteIndex map[string]any) ([]map[string]any, []map[string]any, error) {
	objects := []map[string]any{}
	ip, ok := paths[key+".ilb"]
	if !ok {
		return objects, nil, nil
	}
	id, err := vol.ReadFile(ip)
	if err != nil {
		return nil, nil, err
	}
	buf, cells := ilb.Decode(id)
	ovls, swaps := trackOverlays(track, len(cells), co.H)
	done := map[string]bool{}
	for _, o := range ovls {
		if o.Cell < 0 || o.Cell >= len(cells) || cells[o.Cell].W == 0 {
			continue
		}
		render := func(c ilb.Cell) *image.RGBA {
			if o.HasRamp && c.Typ == 1 {
				return ilb.RenderRamp(buf, c, o.Ramp) // the record's own sprite colours
			}
			return ilb.Render(buf, c, co.Palette, 6)
		}
		// hardware-sprite pieces whose record list holds several same-size
		// type-1 frames ANIMATE: the engine's region tick walks the list every
		// `hold` frames (region +$22/+$23, cursor advance $F778) — the goal
		// flags' 4 wave frames at 5 frames each, on every course
		anim := o.Hold > 0 && len(o.Frames) > 1
		for _, f := range o.Frames {
			if f < 0 || f >= len(cells) || cells[f].Typ != 1 ||
				cells[f].W != cells[o.Cell].W || cells[f].H != cells[o.Cell].H {
				anim = false
			}
		}
		var sprite string
		if anim {
			sprite = fmt.Sprintf("%s/a%d", key, o.Region)
			if !done[sprite] {
				done[sprite] = true
				cl := cells[o.Cell]
				strip := image.NewRGBA(image.Rect(0, 0, cl.W*len(o.Frames), cl.H))
				frames := make([][4]int, len(o.Frames))
				durs := make([]int, len(o.Frames))
				for i, f := range o.Frames {
					img := render(cells[f])
					draw.Draw(strip, image.Rect(i*cl.W, 0, (i+1)*cl.W, cl.H), img, image.Point{}, draw.Src)
					frames[i] = [4]int{i * cl.W, 0, cl.W, cl.H}
					durs[i] = o.Hold
				}
				pngName := fmt.Sprintf("%s-a%d.png", key, o.Region)
				if err := gfx.WritePNG(filepath.Join(spritesDir, pngName), strip); err != nil {
					return nil, nil, err
				}
				spriteIndex[sprite] = map[string]any{
					"src": "sprites/" + pngName, "frames": frames, "durations": durs,
				}
			}
		} else {
			sprite = fmt.Sprintf("%s/c%d", key, o.Cell)
			if !done[sprite] {
				done[sprite] = true
				pngName := fmt.Sprintf("%s-c%d.png", key, o.Cell)
				cl := cells[o.Cell]
				if err := gfx.WritePNG(filepath.Join(spritesDir, pngName), render(cl)); err != nil {
					return nil, nil, err
				}
				spriteIndex[sprite] = map[string]any{
					"src":    "sprites/" + pngName,
					"frames": [][4]int{{0, 0, cl.W, cl.H}},
				}
			}
		}
		objects = append(objects, map[string]any{
			"type":   o.Region,
			"name":   fmt.Sprintf("region %d overlay (cell %d)", o.Region, o.Cell),
			"x":      o.X,
			"y":      o.Y,
			"sprite": sprite,
		})
	}

	// screen-swap tile animations (Ultimate's final screen): the engine cycles
	// the fixed 30-row band through the variant blocks. Content and order are
	// exact from the Track data; the per-variant hold is in-game calibrated
	// (~one second per variant — the Painter pipeline repaints the band one
	// column job at a time across many world frames)
	var anims []map[string]any
	for _, sw := range swaps {
		const bandRows = 30   // the engine's Painter job height ($1E038 sends $1E rows)
		const holdFrames = 50 // ~1 s per variant at 50 Hz, measured in-game
		phases := make([]map[string]any, len(sw.Rows))
		for i, row := range sw.Rows {
			tiles := make([]int, 0, bandRows*mlb.CourseW)
			for y := row; y < row+bandRows && y < co.H; y++ {
				tiles = append(tiles, co.Cells[y*mlb.CourseW:(y+1)*mlb.CourseW]...)
			}
			phases[i] = map[string]any{"tiles": tiles, "frames": holdFrames}
		}
		anims = append(anims, map[string]any{
			"tx": 0, "ty": sw.Ty, "tw": mlb.CourseW, "th": bandRows, "phases": phases,
		})
	}
	return objects, anims, nil
}

type overlay struct {
	Region  int   // dynamic-region index
	Cell    int   // .ilb cell index (the op2-selected state)
	Frames  []int // all cell indices of the piece's record list, in list order
	Hold    int   // op2 +$23 = engine frames per animation step (0 = no anim)
	X, Y    int   // world px (top-left of the draw)
	KX, KY  int   // keyframe grid pos (for reference/debug)
	Selects int   // op2 operand (list entry index)
	HasRamp bool
	Ramp    [3]uint16 // record +$C -> the sprite pair's 3 $0RGB colours (type-1 cells)
}

// swapAnim is a screen-swap tile animation (Ultimate's final screen): the engine
// region registered with terr $19 hands the $1E212 Painter engine a list of
// [tilemap row, $798 height patch] pairs; each cycle repaints the fixed 30-row
// screen band (engine constant, the $1E038 Painter job) from the next variant's
// rows and swaps the collision heights to match.
type swapAnim struct {
	Ty   int   // the band's top tile row (region +$28, also the first variant row)
	Rows []int // variant base rows, in cycle order
}

func ou16(b []byte, o uint32) int {
	if int(o)+2 > len(b) {
		return 0
	}
	return int(b[o])<<8 | int(b[o+1])
}

func os16(b []byte, o uint32) int {
	v := ou16(b, o)
	if v >= 0x8000 {
		v -= 0x10000
	}
	return v
}

func ou32(b []byte, o uint32) uint32 {
	if int(o)+4 > len(b) {
		return 0
	}
	return uint32(b[o])<<24 | uint32(b[o+1])<<16 | uint32(b[o+2])<<8 | uint32(b[o+3])
}

func s8(v byte) int { return int(int8(v)) }

// spriteList reads the $FFFFFFFF-terminated pointer list at off and returns the
// offsets of its 16-byte sprite records. Lists are either flat records or
// [record, composite] pairs; a pointer is a record iff its +8 long is a sane
// cell index, which is how the pair layout is detected.
func spriteList(im []byte, off uint32, nCells int) []uint32 {
	if off == 0 || int(off)+4 > len(im) {
		return nil
	}
	var ptrs []uint32
	for o := off; int(o)+4 <= len(im) && len(ptrs) < 64; o += 4 {
		v := ou32(im, o)
		if v == 0xFFFFFFFF {
			break
		}
		ptrs = append(ptrs, v)
	}
	valid := func(p uint32) bool {
		if p == 0 || int(p)+16 > len(im) {
			return false
		}
		return int(ou32(im, p+8)) < nCells
	}
	// pair layout: even slots are records, odd slots composite lists
	pairs := len(ptrs) >= 2 && len(ptrs)%2 == 0
	for i := 0; pairs && i < len(ptrs); i += 2 {
		if !valid(ptrs[i]) || valid(ptrs[i+1]) {
			pairs = false
		}
	}
	var recs []uint32
	step := 1
	if pairs {
		step = 2
	}
	for i := 0; i < len(ptrs); i += step {
		if valid(ptrs[i]) {
			recs = append(recs, ptrs[i])
		} else {
			recs = append(recs, 0) // keep indices aligned with op2 selects
		}
	}
	return recs
}

// rowPairList tries to read the list at off as the screen-swap form: [tilemap
// row, data ptr] pairs, $FFFFFFFF-terminated. Returns the rows, or nil if the
// list doesn't have that shape.
func rowPairList(im []byte, off uint32, maxRow int) []int {
	if off == 0 || int(off)+4 > len(im) {
		return nil
	}
	var rows []int
	for o := off; int(o)+8 <= len(im) && len(rows) < 16; o += 8 {
		row := ou32(im, o)
		if row == 0xFFFFFFFF {
			return rows
		}
		ptr := ou32(im, o+4)
		if row >= uint32(maxRow) || ptr == 0 || int(ptr)+4 > len(im) || ptr < 0x100 {
			return nil
		}
		rows = append(rows, int(row))
	}
	return nil
}

// trackOverlays walks every dynamic region's script and returns the scenery
// sprite pieces it places plus any screen-swap tile animations. nCells bounds
// the course .ilb bank, mapRows the course tilemap data rows.
//
// The course descriptor's +$12 word (the final-screen scroll offset; Silly's
// max scroll) is NOT part of the world mapping: $E6B2 adds it to $9BA only on
// course 4 (Silly) to seed the scroll for its bottom-of-course start, and the
// scroll terms cancel out of the world-space form — so no course gets a base
// correction here. (An earlier build added it for Aerial — wrong course AND
// wrong term; it pushed Aerial's goal flags ~784 px up the map.)
func trackOverlays(im []byte, nCells, mapRows int) ([]overlay, []swapAnim) {
	hdr0 := ou32(im, 0)
	base := os16(im, hdr0+0x10)
	dyn := ou32(im, 0x14)
	if dyn == 0 || int(dyn)+4 > len(im) {
		return nil, nil
	}
	list := ou32(im, dyn)

	type region struct {
		idx    int
		script uint32
	}
	var regs []region
	var starts []uint32
	for o, i := list, 0; int(o)+6 <= len(im); o, i = o+6, i+1 {
		if im[o] == 0xFF {
			break
		}
		sp := ou32(im, o+2)
		regs = append(regs, region{i, sp})
		starts = append(starts, sp)
	}
	sort.Slice(starts, func(a, b int) bool { return starts[a] < starts[b] })
	end := func(sp uint32) uint32 {
		for _, p := range starts {
			if p > sp {
				return p
			}
		}
		return sp + 64
	}

	var out []overlay
	var swaps []swapAnim
	for _, r := range regs {
		ovls, swap := scanScript(im, r.idx, r.script, end(r.script), base, nCells, mapRows)
		out = append(out, ovls...)
		if swap != nil {
			swaps = append(swaps, *swap)
		}
	}
	return out, swaps
}

// scanScript linearly decodes one region's opcode stream, tracking the current
// keyframe anchor and sprite-list link, and emits an overlay for every op2
// SPRITE selection (or the list head when a linked script never issues op2).
// A terr-$19 registration (op0 dur==1 whose link is a [row, ptr] pair list)
// comes back as a swapAnim instead.
func scanScript(im []byte, idx int, pc, stop uint32, base, nCells, mapRows int) ([]overlay, *swapAnim) {
	var out []overlay
	var swap *swapAnim
	seen := map[[2]int]bool{}
	var kx, ky, kz, kdur, kw26, kw28 int
	var link uint32
	linkFresh := false

	emit := func(sel, hold int) {
		recs := spriteList(im, link, nCells)
		if sel < 0 || sel >= len(recs) || recs[sel] == 0 {
			return
		}
		rec := recs[sel]
		cell := int(ou32(im, rec+8))
		dx, dy := s8(im[rec]), s8(im[rec+1])
		o := overlay{Region: idx, Cell: cell, Selects: sel, Hold: hold, KX: kx, KY: ky}
		for _, r := range recs {
			if r != 0 {
				o.Frames = append(o.Frames, int(ou32(im, r+8)))
			}
		}
		// record +$C -> the sprite's 3-colour ramp (loaded into COLOR17+4n by
		// the engine's copper fragments) for the hardware-sprite (type-1) cells
		if rp := ou32(im, rec+12); rp != 0 && int(rp)+6 <= len(im) {
			r := [3]uint16{uint16(ou16(im, rp)), uint16(ou16(im, rp+2)), uint16(ou16(im, rp+4))}
			if r[0] <= 0xFFF && r[1] <= 0xFFF && r[2] <= 0xFFF {
				o.HasRamp, o.Ramp = true, r
			}
		}
		if kdur == 1 {
			// dur==1 keyframes carry the piece's draw anchor directly, as
			// course-tile coords in +$26/+$28 (the $105FE/$D868 draw path —
			// the drawbridge and the other tile-area pieces); the record dy
			// nudge raises the piece ($122AC flips y, so +dy moves it UP)
			o.X, o.Y = kw26*8+dx, kw28*8-dy
		} else {
			// free-standing pieces (the goal flags) anchor at the keyframe
			// ref point through the engine's iso projection ($6918/$122AC).
			// The +$90 flip constant is in-game calibrated (the hardware
			// sprites sit 12 px above the raw $108-y buffer arithmetic —
			// the sprite-vs-playfield-band vertical offset).
			x8, y8 := kx*8+4, ky*8+4
			o.X = (y8 - x8) + 0x88 + dx
			o.Y = (x8+y8)/2 - kz - base + 0x90 - dy
		}
		// a script that re-selects at the same anchor is animating one piece
		// (the drawbridge's lift states); keep the first state only
		key := [2]int{o.X, o.Y}
		if !seen[key] {
			seen[key] = true
			out = append(out, o)
		}
	}

	sawOp2 := false
	for pc < stop && int(pc)+2 <= len(im) {
		op := ou16(im, pc)
		pc += 2
		switch op {
		case 0: // KEYFRAME x,y,z,dur,terr (+ w26,w28,link long when dur==1)
			kx, ky, kz = os16(im, pc), os16(im, pc+2), os16(im, pc+4)
			kdur = os16(im, pc+6)
			pc += 10
			if kdur == 1 {
				kw26, kw28 = os16(im, pc), os16(im, pc+2)
				link = ou32(im, pc+4)
				linkFresh = true
				pc += 8
				// a [tilemap row, height patch] pair list = a terr-$19
				// SCREEN-SWAP registration (Ultimate's final screen), not
				// sprite records — the $1E212 Painter engine cycles it
				if rows := rowPairList(im, link, mapRows); len(rows) >= 2 && rows[0] == kw28 {
					swap = &swapAnim{Ty: kw28, Rows: rows}
					linkFresh = false
				}
			}
		case 2: // SPRITE: select list entry, second word = hold frames/step
			emit(os16(im, pc), os16(im, pc+2))
			sawOp2 = true
			linkFresh = false
			pc += 4
		case 12, 13: // LINK: swap the sprite list
			link = ou32(im, pc)
			linkFresh = true
			pc += 4
		case 1, 8:
			if op == 1 {
				pc += 4
			} else {
				pc += 6
			}
		case 3, 4, 6, 17:
			pc += 2
		case 9, 10:
			pc += 4
		case 18:
			pc += 6
		case 5, 7, 11, 15, 16:
		case 14:
			pc += 4
		default: // >= $13: end of script data
			pc = stop
		}
	}
	// a region that links a sprite list but never op2-selects shows entry 0
	if !sawOp2 && linkFresh {
		emit(0, 0)
	}
	return out, swap
}
