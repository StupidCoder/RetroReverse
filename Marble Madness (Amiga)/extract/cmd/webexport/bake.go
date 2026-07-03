// Band baking: the display block's per-band COLOR10-15 overrides become
// recoloured variant tiles appended to the atlas, substituted into the level's
// cells row by row — the viewer needs no palette machinery. The Ultimate gold
// shimmer becomes tileAnims cycling those variants through the three rotation
// phases.
package main

import (
	"image/color"
	"sort"

	"marblemad/extract/mlb"
)

type bandBake struct {
	co     *mlb.Course
	fx     *displayFx
	ext    color.Palette // extended palette: base 16 + appended band/phase colours
	palIdx map[uint16]uint8
	base16 [16]uint16

	colour  []int                   // indices into fx.Bands that carry colours
	remaps  map[int]map[uint8]uint8 // colour-band idx -> palette-index remap
	varIdx  map[[3]int]int          // {tile, band, phase(-1 = none)} -> atlas id
	varList []mlb.Variant
}

func rgb4(w uint16) color.RGBA {
	n := func(x uint16) uint8 { return uint8(x&0xF) * 17 }
	return color.RGBA{n(w >> 8), n(w >> 4), n(w), 0xFF}
}

func newBandBake(co *mlb.Course, fx *displayFx) *bandBake {
	b := &bandBake{co: co, fx: fx,
		palIdx: map[uint16]uint8{},
		remaps: map[int]map[uint8]uint8{},
		varIdx: map[[3]int]int{}}
	b.base16 = co.PalWords
	if fx != nil && len(fx.Bands) > 0 {
		for slot, w := range fx.Bands[0].Cols {
			b.base16[slot] = w // band 0 supplies the true 10-15 (stored zeros are display-driven)
		}
	}
	b.ext = make(color.Palette, 16)
	for i, w := range b.base16 {
		b.ext[i] = rgb4(w)
		if _, ok := b.palIdx[w]; !ok {
			b.palIdx[w] = uint8(i)
		}
	}
	if fx != nil {
		for i, bd := range fx.Bands {
			if len(bd.Cols) > 0 {
				b.colour = append(b.colour, i)
			}
		}
	}
	return b
}

func (b *bandBake) addColor(w uint16) uint8 {
	if i, ok := b.palIdx[w]; ok {
		return i
	}
	b.ext = append(b.ext, rgb4(w))
	b.palIdx[w] = uint8(len(b.ext) - 1)
	return b.palIdx[w]
}

// remap builds (once) the palette-index remap for a colour band: every slot
// whose band colour differs from the base palette.
func (b *bandBake) remap(band int) map[uint8]uint8 {
	if r, ok := b.remaps[band]; ok {
		return r
	}
	r := map[uint8]uint8{}
	for slot, w := range b.fx.Bands[band].Cols {
		if w != b.base16[slot] {
			r[uint8(slot)] = b.addColor(w)
		}
	}
	b.remaps[band] = r
	return r
}

// bandAt returns the colour-band index for a tile row (the band whose pixel row
// the row's centre has passed; splice-only bands carry colours forward).
func (b *bandBake) bandAt(row int) int {
	cur := 0
	if b.fx == nil {
		return 0
	}
	for _, i := range b.colour {
		if b.fx.Bands[i].PxRow <= row*8+4 {
			cur = i
		}
	}
	return cur
}

// tileFor returns the atlas id showing tile t in colour band `band` — t itself
// for band 0 or when the tile uses none of the band's changed slots, else a
// recoloured variant (created on first use).
func (b *bandBake) tileFor(t, band int) int {
	if band == 0 || b.fx == nil {
		return t
	}
	r := b.remap(band)
	mask := b.co.TileMask(t)
	used := false
	for slot := range r {
		if mask&(1<<slot) != 0 {
			used = true
			break
		}
	}
	if !used {
		return t
	}
	return b.variant(t, band, -1, r)
}

func (b *bandBake) variant(t, band, phase int, remap map[uint8]uint8) int {
	key := [3]int{t, band, phase}
	if id, ok := b.varIdx[key]; ok {
		return id
	}
	id := b.co.NTiles + len(b.varList)
	b.varList = append(b.varList, mlb.Variant{Src: t, Remap: remap})
	b.varIdx[key] = id
	return id
}

// paletteAt returns the 16-colour playfield palette as the copper shows it at
// a given tile row: the base palette with the row's colour-band overrides —
// overlay sprite strips must be rendered with it (the Intermediate wave sits
// on the emerald band; the stored palette would paint it band-0 orange).
func (b *bandBake) paletteAt(row int) color.Palette {
	words := b.base16
	if b.fx != nil {
		for slot, w := range b.fx.Bands[b.bandAt(row)].Cols {
			words[slot] = w
		}
	}
	p := make(color.Palette, 16)
	for i, w := range words {
		p[i] = rgb4(w)
	}
	return p
}

// shimmerAnims emits the gold-rotation tileAnims: for each shimmer destination
// (a band's slots Slot..Slot+2), every tile of that band using one of the three
// slots cycles through the rotation lists — phase p writes Phases[p][0..2] over
// the three slots (the $84DC copy). tilesInBand maps colour-band idx -> the set
// of ORIGINAL tile ids shown there (visible rows + swap-variant rows).
func (b *bandBake) shimmerAnims(tilesInBand map[int]map[int]bool, periodFrames int) []map[string]any {
	if b.fx == nil || len(b.fx.Shimmer) == 0 {
		return nil
	}
	var anims []map[string]any
	for _, dst := range b.fx.Shimmer {
		slotMask := uint16(0b111) << dst.Slot
		var ids []int
		for t := range tilesInBand[dst.Band] {
			if b.co.TileMask(t)&slotMask != 0 {
				ids = append(ids, t)
			}
		}
		sort.Ints(ids)
		if len(ids) == 0 {
			continue
		}
		tiles := make([]int, len(ids))
		frames := make([][]int, len(b.fx.Phases))
		for p := range frames {
			frames[p] = make([]int, len(ids))
		}
		for k, t := range ids {
			tiles[k] = b.tileFor(t, dst.Band)
			for p, ph := range b.fx.Phases {
				r := map[uint8]uint8{}
				for s, v := range b.remap(dst.Band) {
					r[s] = v
				}
				for i, w := range ph {
					r[uint8(dst.Slot+i)] = b.addColor(w)
				}
				frames[p][k] = b.variant(t, dst.Band, p, r)
			}
		}
		anims = append(anims, map[string]any{
			"tiles": tiles, "frames": frames, "periodFrames": periodFrames,
		})
	}
	return anims
}
