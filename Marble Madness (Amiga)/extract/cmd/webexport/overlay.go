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
	"path/filepath"
	"sort"

	"marblemad/extract/ilb"
	"marblemad/extract/mlb"
	"retroreverse.com/tools/amiga/adf"
	"retroreverse.com/tools/c64/gfx"
)

// exportOverlays extracts the course's scenery-overlay pieces, renders each
// referenced .ilb cell once into spritesDir (recorded in spriteIndex, keyed
// "<course>/c<cell>"), and returns the level's objects entries.
func exportOverlays(vol *adf.Volume, paths map[string]string, key string, track []byte,
	co *mlb.Course, spritesDir string, spriteIndex map[string]any) ([]map[string]any, error) {
	objects := []map[string]any{}
	ip, ok := paths[key+".ilb"]
	if !ok {
		return objects, nil
	}
	id, err := vol.ReadFile(ip)
	if err != nil {
		return nil, err
	}
	buf, cells := ilb.Decode(id)
	ovls := trackOverlays(track, len(cells), key == "aerial")
	done := map[int]bool{}
	for _, o := range ovls {
		if o.Cell < 0 || o.Cell >= len(cells) || cells[o.Cell].W == 0 {
			continue
		}
		sprite := fmt.Sprintf("%s/c%d", key, o.Cell)
		if !done[o.Cell] {
			done[o.Cell] = true
			pngName := fmt.Sprintf("%s-c%d.png", key, o.Cell)
			cl := cells[o.Cell]
			img := ilb.Render(buf, cl, co.Palette, 6)
			if err := gfx.WritePNG(filepath.Join(spritesDir, pngName), img); err != nil {
				return nil, err
			}
			spriteIndex[sprite] = map[string]any{
				"src":    "sprites/" + pngName,
				"frames": [][4]int{{0, 0, cl.W, cl.H}},
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
	return objects, nil
}

type overlay struct {
	Region  int // dynamic-region index
	Cell    int // .ilb cell index
	X, Y    int // world px (top-left of the blit)
	KX, KY  int // keyframe grid pos (for reference/debug)
	Selects int // op2 operand (list entry index)
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

// trackOverlays walks every dynamic region's script and returns the scenery
// sprite pieces it places. nCells bounds the course .ilb bank; aerial selects
// the course-4 base correction ($E6B2 adds descriptor +$12 only there).
func trackOverlays(im []byte, nCells int, aerial bool) []overlay {
	hdr0 := ou32(im, 0)
	base := os16(im, hdr0+0x10)
	if aerial {
		base += os16(im, hdr0+0x12)
	}
	dyn := ou32(im, 0x14)
	if dyn == 0 || int(dyn)+4 > len(im) {
		return nil
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
	for _, r := range regs {
		out = append(out, scanScript(im, r.idx, r.script, end(r.script), base, nCells)...)
	}
	return out
}

// scanScript linearly decodes one region's opcode stream, tracking the current
// keyframe anchor and sprite-list link, and emits an overlay for every op2
// SPRITE selection (or the list head when a linked script never issues op2).
func scanScript(im []byte, idx int, pc, stop uint32, base, nCells int) []overlay {
	var out []overlay
	seen := map[[2]int]bool{}
	var kx, ky, kz, kdur, kw26, kw28 int
	var link uint32
	linkFresh := false

	emit := func(sel int) {
		recs := spriteList(im, link, nCells)
		if sel < 0 || sel >= len(recs) || recs[sel] == 0 {
			return
		}
		rec := recs[sel]
		cell := int(ou32(im, rec+8))
		dx, dy := s8(im[rec]), s8(im[rec+1])
		o := overlay{Region: idx, Cell: cell, Selects: sel, KX: kx, KY: ky}
		if kdur == 1 {
			// dur==1 keyframes carry the piece's draw anchor directly, as
			// course-tile coords in +$26/+$28 (the $105FE/$D868 draw path —
			// the drawbridge and the other tile-area pieces)
			o.X, o.Y = kw26*8+dx, kw28*8+dy
		} else {
			// free-standing pieces (the goal flags) anchor at the keyframe
			// ref point through the engine's iso projection ($6918/$122AC)
			x8, y8 := kx*8+4, ky*8+4
			o.X = (y8 - x8) + 0x88 + dx
			o.Y = (x8+y8)/2 - kz - base + 0x9C - dy
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
			}
		case 2: // SPRITE: select list entry
			emit(os16(im, pc))
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
		emit(0)
	}
	return out
}
