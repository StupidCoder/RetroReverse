// Overlay extraction: the scenery/hazard pieces the engine layers over the
// tilemap, replayed for the site's course maps (Marble_Madness.md Part IV §5
// "One game, eight obstacle systems" + Part V §2).
//
// Sources handled here, all from the course Track file:
//  - dynamic-region scripts (header +$14 record list): drawbridge, goal flags,
//    Practice's start ramp, Aerial's duct muncher — op0 KEYFRAME carries the
//    anchor (tile coords +$26/+$28 when dur==1, else the iso-projected ref
//    point); op2 [wrap count][hold] plays the current record list; op3 holds;
//    op12/op13 relink; op14/op16 drift the anchor (the wave sweep); op4/op5
//    loops are expanded; op8/op18 conditional targets are scanned too.
//  - the loose hazard script at block+8, spawned by code on marbles entering
//    placement-zone 9/$A (Intermediate's WAVE).
//  - the 3-slot obstacle-actor array at block+$23C (Aerial's pistons, world-px
//    positions, RNG anim variants at +$278).
//  - the vacuum hood trigger scripts at block +$54/+$58, run on the terr-11/13
//    fall regions (anchor taken from the region's op0).
// A record list is a $FFFFFFFF-terminated pointer array (flat, [record,
// composite] pairs, or 8-byte [record][0] anim entries) of 16-byte sprite
// records: [+0 dx][+1 dy nudge (dy raises)][+2..3 state][+4..7 iso-box extents]
// [+8 (bank,cell) words][+C colour-ramp ptr].
//
// Iso projection for free-standing pieces ($6918/$122AC): x8 = X*8+4,
// y8 = Y*8+4; worldX = (y8-x8)+$88+dx; worldY = (x8+y8)/2 - z - base + $90 - dy
// with base = descriptor word +$10 (the +$12 word is Silly's initial-scroll
// seed and cancels out of the world mapping).
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
		// a piece whose script composed a multi-frame play program ANIMATES:
		// the strip holds each distinct cell once, and the steps replay the
		// program (op2 list play-throughs + op3 holds) at engine rate
		var steps []step
		distinct := map[int]int{}
		var order []int
		for _, st := range o.Steps {
			if st.Cell < 0 || st.Cell >= len(cells) || cells[st.Cell].W == 0 {
				continue
			}
			if _, ok := distinct[st.Cell]; !ok {
				distinct[st.Cell] = len(order)
				order = append(order, st.Cell)
			}
			steps = append(steps, st)
		}
		var sprite string
		if len(steps) > 1 && len(order) > 1 {
			sprite = fmt.Sprintf("%s/a%d", key, o.Region)
			if !done[sprite] {
				done[sprite] = true
				w, h := 0, 0
				for _, c := range order {
					w += cells[c].W
					if cells[c].H > h {
						h = cells[c].H
					}
				}
				strip := image.NewRGBA(image.Rect(0, 0, w, h))
				frames := make([][4]int, len(order))
				x := 0
				for i, c := range order {
					img := render(cells[c])
					draw.Draw(strip, image.Rect(x, 0, x+cells[c].W, cells[c].H), img, image.Point{}, draw.Src)
					frames[i] = [4]int{x, 0, cells[c].W, cells[c].H}
					x += cells[c].W
				}
				prog := make([][2]int, len(steps))
				for i, st := range steps {
					prog[i] = [2]int{distinct[st.Cell], st.Hold}
				}
				pngName := fmt.Sprintf("%s-a%d.png", key, o.Region)
				if err := gfx.WritePNG(filepath.Join(spritesDir, pngName), strip); err != nil {
					return nil, nil, err
				}
				entry := map[string]any{
					"src": "sprites/" + pngName, "frames": frames, "steps": prog,
				}
				// per-frame offsets (record nudges + op16 drift) -> a movement path
				hasMove := false
				for _, st := range steps {
					if st.DX != 0 || st.DY != 0 || st.OX != 0 || st.OY != 0 {
						hasMove = true
					}
				}
				if hasMove {
					var path [][2]int
					ox, oy := 0, 0
					for _, st := range steps {
						for f := 0; f < st.Hold && len(path) < 2048; f++ {
							path = append(path, [2]int{ox + st.OX, oy + st.OY})
						}
						ox += st.DX
						oy += st.DY
					}
					entry["path"] = path
				}
				spriteIndex[sprite] = entry
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
	// exact from the Track data; the cadence is the Painter's repaint time —
	// the tile drawer yields to the vblank every 16 tiles (mlb_draw_column
	// $99A4), so one 30x36-tile band takes ceil(1080/16) = 68 PAL frames
	var anims []map[string]any
	for _, sw := range swaps {
		const bandRows = 30   // the engine's Painter job height ($1E038 sends $1E rows)
		const holdFrames = 68 // 1080 tiles at 16 tiles per vblank ≈ 1.36 s
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

// step is one frame of a piece's play program: show cell for hold engine frames
// at offset (OX,OY) from the anchor (the record dx/dy nudge relative to record
// 0's — the wave's frames each sit differently), then drift the piece by
// (DX,DY) px (op16 MOVE — the wave sweep).
type step struct{ Cell, Hold, OX, OY, DX, DY int }

type overlay struct {
	Region  int    // dynamic-region index
	Cell    int    // .ilb cell index of the resting state (first list record)
	Steps   []step // the piece's play program, composed from its script's op2/op3/op16 events
	X, Y    int    // world px (top-left of the draw)
	KX, KY  int    // keyframe grid pos (for reference/debug)
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
	covered := map[int]bool{}
	for _, r := range regs {
		ovl, swap := scanScript(im, r.idx, r.script, end(r.script), base, nCells, mapRows, nil)
		if ovl != nil {
			out = append(out, *ovl)
			covered[r.idx] = true
		}
		if swap != nil {
			swaps = append(swaps, *swap)
		}
	}

	// The VACUUM hoods (Aerial): the terr-11/13 fall regions carry only an
	// anchor + an empty sprite list — their hood open/close animation lives in
	// the block's trigger-script slots (+$54 for the terr-11 group, +$58 for
	// terr-13; the code-11/13 "pull toward ref" handlers $17AA8/$17C88 run them
	// on the region when the marble is drawn in). Scan the hood script with the
	// region's own anchor.
	for _, r := range regs {
		if covered[r.idx] || int(r.script)+16 > len(im) ||
			ou16(im, r.script) != 0 || os16(im, r.script+8) != 1 {
			continue
		}
		terr := os16(im, r.script+10)
		anchor := [2]int{os16(im, r.script+12), os16(im, r.script+14)}
		var sc uint32
		switch terr {
		// the pairing is the region loader's ($12B32): a region with terr
		// byte $B (11) gets block+$54 stored at region+$36, terr $D (13)
		// gets block+$58 — traced, not calibrated
		case 11:
			sc = ou32(im, dyn+0x54)
		case 13:
			sc = ou32(im, dyn+0x58)
		default:
			continue
		}
		if sc == 0 || int(sc)+8 > len(im) {
			continue
		}
		if ovl, _ := scanScript(im, r.idx, sc, sc+0x80, base, nCells, mapRows, &anchor); ovl != nil && len(ovl.Steps) > 0 {
			idle := ovl.Steps[len(ovl.Steps)-1]
			idle.Hold, idle.DX, idle.DY = 60, 0, 0
			ovl.Steps = append(ovl.Steps, idle)
			out = append(out, *ovl)
		}
	}

	// The hazard script at block+8 — a region SPAWNED BY CODE, not by the record
	// list: when a marble enters placement-zone type 9/$A, $FCE6 allocates a
	// $CCA region struct and runs this script on it ($F6F4 -> region_update).
	// Only Intermediate carries one (THE WAVE: rise, set velocity, sweep loop
	// x19 with op16 MOVE, collapse); the other courses keep unrelated data in
	// the slot, so accept it only when it opens with an op0 dur==1 keyframe.
	if p8 := ou32(im, dyn+8); p8 != 0 && int(p8)+20 < len(im) &&
		ou16(im, p8) == 0 && os16(im, p8+8) == 1 {
		if ovl, _ := scanScript(im, 90, p8, p8+0x180, base, nCells, mapRows, nil); ovl != nil && len(ovl.Steps) >= 3 {
			out = append(out, *ovl)
		}
	}

	// The 3-slot obstacle-actor array at block+$23C (populated only by Aerial —
	// the pop-up pistons/vacuums): each $14-byte record = [+0 ptr to its base
	// 16-byte sprite record][+$A x][+$C y] in world px (the actor draw is the
	// blitter path: draw_object_wrap(rec, +$A, +$C - scroll), so +$C maps to
	// world Y directly); anim variants (8-byte [recordPtr][0] lists, one step
	// per frame, holds encoded as repeated pointers) at block +$278..+$284.
	// The engine picks a variant by RNG per activation ($1D48A) and idles the
	// base record for an RNG (0..3)<<4 frame pause; the export replays that as
	// a deterministic cycle through all four patterns (actor i starts at
	// variant i, so the three actors desync).
	for i := 0; i < 3; i++ {
		rec := dyn + 0x23C + uint32(i)*0x14
		if int(rec)+0x14 > len(im) {
			break
		}
		basePtr := ou32(im, rec)
		x, y := ou16(im, rec+0xA), ou16(im, rec+0xC)
		if basePtr == 0 || int(basePtr)+16 > len(im) || ou16(im, basePtr+8) != 0 ||
			ou16(im, basePtr+10) >= nCells || y == 0 || y >= mapRows*8 {
			continue
		}
		baseCell := ou16(im, basePtr+10)
		baseDX, baseDY := s8(im[basePtr]), -s8(im[basePtr+1])
		o := overlay{Region: 91 + i, Cell: baseCell, X: x, Y: y}
		for v := 0; v < 4; v++ {
			variant := ou32(im, dyn+0x278+uint32((i+v)%4)*4)
			if variant == 0 {
				continue
			}
			n := len(o.Steps)
			for _, r := range spriteList(im, variant, nCells) {
				if r != 0 {
					o.Steps = append(o.Steps, step{Cell: int(ou32(im, r+8)), Hold: 1,
						OX: s8(im[r]) - baseDX, OY: -s8(im[r+1]) - baseDY})
				}
			}
			if len(o.Steps) > n {
				o.Steps = append(o.Steps, step{Cell: baseCell, Hold: 32}) // idle pause between pop-ups
			}
		}
		out = append(out, o)
	}
	return out, swaps
}

// scanScript decodes one region's opcode stream (following one level of
// conditional/jump targets — the Practice ramp's animation lives at its op8
// target), tracking the current keyframe anchor and sprite-list link. Each
// op2 SPRITE starts a play-through of the current record list (word1 = wrap
// count, word2 = hold frames/step); each op3 holds the last frame; op14/op16
// drift the anchor (the wave). The events compose the region's play program.
// A terr-$19 registration (op0 dur==1 whose link is a [row, ptr] pair list)
// comes back as a swapAnim instead. anchor overrides the tile anchor for
// scripts run ON a region from elsewhere (the vacuum hood triggers).
func scanScript(im []byte, idx int, pc, stop uint32, base, nCells, mapRows int, anchor *[2]int) (*overlay, *swapAnim) {
	var out *overlay
	var swap *swapAnim
	var kx, ky, kz, kdur, kw26, kw28 int
	var link uint32
	linkFresh := false
	lastHold := 1
	if anchor != nil {
		kdur, kw26, kw28 = 1, anchor[0], anchor[1]
	}

	// place the region's (single) overlay from the current anchor context
	place := func(recs []uint32) {
		if out != nil || len(recs) == 0 || recs[0] == 0 {
			return
		}
		rec := recs[0]
		cell := int(ou32(im, rec+8))
		dx, dy := s8(im[rec]), s8(im[rec+1])
		o := &overlay{Region: idx, Cell: cell, KX: kx, KY: ky}
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
		out = o
	}

	// op2: play the current list through (once per wrap count; 0 = forever =
	// once through the loop, since the whole program loops). Each record's
	// dx/dy nudge is kept per frame, relative to the anchor record's — the
	// wave's frames each sit at their own offset ($122AC applies rec+0/+1
	// per draw; dy raises).
	var anchorDX, anchorDY int
	anchorSet := false
	anim := func(hold int) {
		recs := spriteList(im, link, nCells)
		if len(recs) == 0 {
			return
		}
		place(recs)
		if out == nil {
			return
		}
		if hold < 1 {
			hold = 1
		}
		lastHold = hold
		for _, r := range recs {
			if r == 0 {
				continue
			}
			dx, dy := s8(im[r]), -s8(im[r+1])
			if !anchorSet {
				anchorSet = true
				anchorDX, anchorDY = dx, dy
			}
			out.Steps = append(out.Steps, step{
				Cell: int(ou32(im, r+8)), Hold: hold,
				OX: dx - anchorDX, OY: dy - anchorDY,
			})
		}
	}
	// op3: hold the last shown frame for count list-replays' worth of frames
	pause := func(count int) {
		if out == nil || count <= 0 || len(out.Steps) == 0 {
			return
		}
		last := out.Steps[len(out.Steps)-1]
		last.Hold = count * lastHold
		last.DX, last.DY = 0, 0
		out.Steps = append(out.Steps, last)
	}
	// op16: drift the anchor by the op14 velocity (the wave sweep)
	var velX, velY int
	move := func() {
		if out == nil || len(out.Steps) == 0 || (velX == 0 && velY == 0) {
			return
		}
		// drawX adds the drift, drawY subtracts it ($11678: x = w26*8 + drift,
		// y = w28*8 - drift) — screen-down is -velY
		out.Steps[len(out.Steps)-1].DX += velX
		out.Steps[len(out.Steps)-1].DY -= velY
	}

	queue := []uint32{pc}
	visited := map[uint32]bool{}
	loopTop, loopN := uint32(0), 0
	for len(queue) > 0 && len(visited) < 6 {
		pc = queue[0]
		queue = queue[1:]
		if visited[pc] {
			continue
		}
		visited[pc] = true
		end := pc + 0x180
		if stop > pc && stop < end {
			end = stop
		}
		for pc < end && int(pc)+2 <= len(im) {
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
			case 2: // SPRITE count,hold: play the current list
				anim(os16(im, pc+2))
				linkFresh = false
				pc += 4
			case 3: // STATE0-STOP count: hold the pose
				pause(os16(im, pc))
				pc += 2
			case 12, 13: // LINK: swap the record list
				link = ou32(im, pc)
				linkFresh = true
				pc += 4
			case 8: // IF-MARBLE-ON [arg word][target long]: scan the target too
				if t := ou32(im, pc+2); t > 0 && int(t) < len(im) {
					queue = append(queue, t)
				}
				pc += 6
			case 18: // IF-MARBLE-TERRAIN [target long] — 4 operand bytes (the
				// handler $10326 swaps the script PC for the long or skips it);
				// the vacuum hoods branch to their close subroutine here
				if t := ou32(im, pc); t > 0 && int(t) < len(im) {
					queue = append(queue, t)
				}
				pc += 4
			case 9: // JUMP
				t := ou32(im, pc)
				pc += 4
				if t > 0 && int(t) < len(im) {
					queue = append(queue, t)
				}
			case 4: // LOOP-A count: expand the loop (the wave's 19-step sweep)
				loopN = os16(im, pc)
				pc += 2
				loopTop = pc
				if loopN <= 0 || loopN > 40 {
					loopN = 0 // infinite loops just play once (the program loops anyway)
				}
			case 5: // NEXT-A: count 0 loops forever in the engine — nothing
				// past op5 is reachable then, so end the segment after one pass
				if loopN > 1 && loopTop != 0 {
					loopN--
					pc = loopTop
				} else if loopN == 0 {
					pc = end
				}
			case 6, 17:
				pc += 2
			case 10:
				pc += 4
			case 7:
			case 11: // RETURN = subroutine boundary — end this segment
				pc = end
			case 16: // MOVE: drift the anchor by the op14 velocity
				move()
			case 14: // SET-VEL
				velX, velY = os16(im, pc), os16(im, pc+2)
				pc += 4
			case 15:
				pc = end
			case 1:
				pc += 4
			default: // >= $13: end of script data
				pc = end
			}
		}
	}
	// a region that links a record list but never op2-selects shows entry 0
	if out == nil && linkFresh {
		place(spriteList(im, link, nCells))
	}
	if out == nil {
		return nil, swap
	}
	return out, swap
}
