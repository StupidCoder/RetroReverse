// The Track display block (header +$10 → the $89C2 copper-band data,
// Marble_Madness.md Part IV §6): every course splits its screen into vertical
// COLOUR BANDS — as the playfield scrolls, the copper rebuild ($86B4/$8852)
// emits each band's colour fragment at screen line (band.pxRow − scroll), so
// palette slots 10-15 fade as the marble descends. The block also carries the
// hazard-pool colour flip (head +$C/+$10/+$14, applied by $8588 when a marble
// enters a pool) and — Ultimate only — the finale gold-shimmer rotation
// (head +$18..+$28, three phase-offset $0RGB triples for the two gold C13
// entries). The exporter bakes the bands into the atlas as recoloured variant
// tiles and emits the shimmer as tileAnims.
package main

// colorBand is one display band: from map pixel row PxRow down to the next
// band, palette slots (keys 10-15) take the given $0RGB values. A band with no
// colour entries (Ultimate's row-624 band writes BPLxPT instead — the finale's
// back-buffer splice) carries the previous band's colours forward.
type colorBand struct {
	PxRow int
	HPos  int // copper WAIT horizontal position (band 3's mid-line split), 0 otherwise
	Cols  map[int]uint16
}

// displayFx is the parsed display block: the colour bands plus the two
// data-driven colour animations (pool flip, gold shimmer).
type displayFx struct {
	Bands []colorBand

	// pool flip: FlipN consecutive entries starting at the entry FlipBand's
	// slot FlipSlot take the A (hazard/red) or B (ice/blue) words on $8588.
	FlipBand, FlipSlot int
	FlipA, FlipB       []uint16

	// gold shimmer (Ultimate): the stepper $84DC copies the current 3-word
	// phase list over three CONSECUTIVE colour entries per destination — the
	// dst pointer marks the first (slot Slot, then Slot+1, Slot+2). Phases are
	// the three rotation lists at head +$20/+$24/+$28; the phase counter
	// ($68E) advances every 4 engine frames ($68F += 2 per frame, wrap at 8).
	Shimmer []shimmerDst
	Phases  [][]uint16
}

// shimmerDst locates one shimmer destination: band index + first palette slot.
type shimmerDst struct{ Band, Slot int }

// bandEntries walks a band fragment's 6-byte copper entries [0000][reg][val]
// backward from its end pointer, returning (entryAddr, reg, val) in order.
func bandEntries(im []byte, end uint32) [][3]int {
	var ents [][3]int
	o := end - 6
	for o > 0 && int(o)+6 <= len(im) {
		z, reg, val := ou16(im, o), ou16(im, o+2), ou16(im, o+4)
		if z != 0 || reg < 0xE000 || reg > 0xF1FF {
			break
		}
		ents = append(ents, [3]int{int(o), reg, val})
		o -= 6
	}
	for i, j := 0, len(ents)-1; i < j; i, j = i+1, j-1 {
		ents[i], ents[j] = ents[j], ents[i]
	}
	return ents
}

// parseDisplayFx decodes the Track's display block. Returns nil if the header
// has no display pointer.
func parseDisplayFx(im []byte) *displayFx {
	head := ou32(im, 0x10)
	if head == 0 || int(head)+0x16 > len(im) {
		return nil
	}
	fx := &displayFx{}
	nBands := ou16(im, head+4)
	list := ou32(im, head+6)
	type bandRaw struct {
		ents [][3]int
	}
	var raws []bandRaw
	for i := 0; i < nBands; i++ {
		e := list + uint32(i*12)
		ents := bandEntries(im, ou32(im, e+8))
		cols := map[int]uint16{}
		for _, en := range ents {
			if en[1] >= 0xF194 && en[1] <= 0xF19E { // COLOR10..COLOR15
				cols[10+(en[1]-0xF194)/2] = uint16(en[2])
			}
		}
		fx.Bands = append(fx.Bands, colorBand{PxRow: ou16(im, e), HPos: ou16(im, e+2), Cols: cols})
		raws = append(raws, bandRaw{ents})
	}

	// locate a colour-entry address inside a band's fragment
	locate := func(addr uint32) (int, int) {
		for bi, r := range raws {
			for _, en := range r.ents {
				if uint32(en[0]) == addr {
					return bi, 10 + (en[1]-0xF194)/2
				}
			}
		}
		return -1, -1
	}

	// pool colour flip: head +$A count, +$C dst entry, +$10/+$14 word lists
	fn := ou16(im, head+0xA)
	fx.FlipBand, fx.FlipSlot = locate(ou32(im, head+0xC))
	for i := 0; i < fn; i++ {
		fx.FlipA = append(fx.FlipA, uint16(ou16(im, ou32(im, head+0x10)+uint32(2*i))))
		fx.FlipB = append(fx.FlipB, uint16(ou16(im, ou32(im, head+0x14)+uint32(2*i))))
	}

	// extended head (Ultimate): +$16 pad, then pointers instead of the first
	// fragment's entries — detect by the word at +$1A not being a colour reg.
	if ou16(im, head+0x1A) < 0xF180 {
		for _, off := range []uint32{0x18, 0x1C} {
			if b, s := locate(ou32(im, head+off)); b >= 0 {
				fx.Shimmer = append(fx.Shimmer, shimmerDst{b, s})
			}
		}
		for _, off := range []uint32{0x20, 0x24, 0x28} {
			p := ou32(im, head+off)
			var ph []uint16
			for i := uint32(0); i < 3; i++ {
				ph = append(ph, uint16(ou16(im, p+2*i)))
			}
			fx.Phases = append(fx.Phases, ph)
		}
	}
	return fx
}
