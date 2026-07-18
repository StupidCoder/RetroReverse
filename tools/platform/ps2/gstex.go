package ps2

// gstex.go is the Graphics Synthesizer's texturing half: the TEX0 register that names a
// texture, the swizzled formats it can be stored in, the CLUT that turns an index into a
// colour, and the per-pixel pipeline that combines a sampled texel with the fragment and
// blends the result into the frame.
//
// The census drove this file into existence: a clean boot draws 337,920 textured sprites
// and 190,080 alpha-blended primitives, all of which gsdraw.go was painting as flat
// vertex colour. What a textured primitive needs is spread over several registers, all
// latched per context:
//
//	TEX0    where the texture is (TBP/TBW), its format (PSM), its size (TW/TH log2),
//	        how its colour combines with the fragment (TFX/TCC), and where its CLUT is
//	        (CBP/CPSM/CSM/CSA) and whether writing this register loads it (CLD)
//	CLAMP   what happens outside [0,W)x[0,H): repeat, clamp, or the region variants
//	TEXA    the alpha to give formats that store none (CT24) or one bit (CT16)
//	TEST    the alpha/destination-alpha/depth tests a pixel must pass
//	ALPHA   the blend equation: Cv = ((A - B) * C >> 7) + D
//
// GS local memory keeps every format in the same 8 KiB pages of 256-byte blocks, but each
// format tiles texels into those blocks differently. The tables here are the hardware's
// arrangement (the GS user's-manual layout — platform architecture, like the GIFtag);
// what makes them load-bearing is ALIASING: this game uploads its textures as PSMCT32
// words and then samples the same pages as PSMT8/PSMT4 indices, so the byte a texel index
// lands on is defined by one format's swizzle composed with the other's. A merely
// self-consistent (bijective but wrong) table would render uploads of the same format
// correctly and shred anything uploaded through a different one.

// The GS pixel-storage formats beyond those gs.go already names.
const (
	psmCT16S = 0x0A
	psmT8    = 0x13
	psmT4    = 0x14
	psmT8H   = 0x1B // the index lives in bits 24..31 of a CT32 word
	psmT4HL  = 0x24 // bits 24..27
	psmT4HH  = 0x2C // bits 28..31
	psmZ32   = 0x30
	psmZ24   = 0x31
	psmZ16   = 0x32
	psmZ16S  = 0x3A
)

// gsTex is TEX0, decoded.
type gsTex struct {
	tbp  uint32 // texture base pointer, 64-word blocks
	tbw  uint32 // buffer width, units of 64 texels
	psm  uint32
	tw   uint32 // log2 width
	th   uint32 // log2 height
	tcc  bool   // alpha comes from the texture (else from the fragment)
	tfx  uint32 // 0 modulate, 1 decal, 2 highlight, 3 highlight2
	cbp  uint32 // CLUT base pointer, 64-word blocks
	cpsm uint32 // CLUT storage format (CT32 or CT16)
	csm  uint32 // CLUT storage mode: 0 swizzled (CSM1), 1 linear (CSM2)
	csa  uint32 // CLUT entry offset, units of 16 entries
	cld  uint32 // CLUT load control, acted on when TEX0 is written
}

func decodeTEX0(v uint64) gsTex {
	return gsTex{
		tbp:  uint32(v) & 0x3FFF,
		tbw:  uint32(v>>14) & 0x3F,
		psm:  uint32(v>>20) & 0x3F,
		tw:   uint32(v>>26) & 0xF,
		th:   uint32(v>>30) & 0xF,
		tcc:  v>>34&1 != 0,
		tfx:  uint32(v>>35) & 3,
		cbp:  uint32(v>>37) & 0x3FFF,
		cpsm: uint32(v>>51) & 0xF,
		csm:  uint32(v>>55) & 1,
		csa:  uint32(v>>56) & 0x1F,
		cld:  uint32(v>>61) & 7,
	}
}

// --- the swizzles -------------------------------------------------------------------
//
// Every format shares the page/block hierarchy addrPSMCT32 documents; what differs per
// format is the page's dimensions in texels, the block table, and how texels pack into a
// block's 64 words. PSMT8 and PSMT4 additionally alternate their word arrangement between
// even and odd columns — the detail that makes their aliasing onto PSMCT32 non-obvious,
// and the reason the tables are written out rather than derived.

// blockPSMT4 is the order of the 4x8 blocks (32x16 texels each) in a PSMT4 page.
var blockPSMT4 = [8][4]uint8{
	{0, 2, 8, 10},
	{1, 3, 9, 11},
	{4, 6, 12, 14},
	{5, 7, 13, 15},
	{16, 18, 24, 26},
	{17, 19, 25, 27},
	{20, 22, 28, 30},
	{21, 23, 29, 31},
}

// blockPSMCT16 is the order of the 4x8 blocks (16x8 texels each) in a PSMCT16 page.
var blockPSMCT16 = [8][4]uint8{
	{0, 2, 8, 10},
	{1, 3, 9, 11},
	{4, 6, 12, 14},
	{5, 7, 13, 15},
	{16, 18, 24, 26},
	{17, 19, 25, 27},
	{20, 22, 28, 30},
	{21, 23, 29, 31},
}

// blockPSMCT16S differs from PSMCT16 only in this table.
var blockPSMCT16S = [8][4]uint8{
	{0, 2, 16, 18},
	{1, 3, 17, 19},
	{8, 10, 24, 26},
	{9, 11, 25, 27},
	{4, 6, 20, 22},
	{5, 7, 21, 23},
	{12, 14, 28, 30},
	{13, 15, 29, 31},
}

// columnWordT8 is the word (0..15) within a PSMT8 column (16x4 texels, 64 bytes) that
// texel (cx, cy) lives in; the two tables alternate by column parity. columnByteT8 is
// which byte of that word.
var columnWordT8 = [2][64]uint8{
	{
		0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
		8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
	},
	{
		8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
		0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
	},
}

var columnByteT8 = [64]uint8{
	0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2,
	0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3,
}

// columnWordT4 / columnNibbleT4: a PSMT4 column is 32x4 texels in 64 bytes; the nibble
// table is in half-byte units (value>>1 the byte, value&1 the nibble).
var columnWordT4 = [2][128]uint8{
	{
		0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
		8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
	},
	{
		8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
		0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
	},
}

var columnNibbleT4 = [128]uint8{
	0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 6, 6, 6, 6, 6,
	0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 6, 6, 6, 6, 6,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3, 5, 5, 5, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 7, 7, 7,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3, 5, 5, 5, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 7, 7, 7,
}

// columnWordCT16 / columnHalfCT16: a PSMCT16 column is 16x2 texels in 64 bytes.
var columnWordCT16 = [32]uint8{
	0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
	2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
}

var columnHalfCT16 = [32]uint8{
	0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1,
}

// addrPSMT8 returns the byte offset of texel (x, y) in an 8-bit buffer at bp (64-word
// blocks), width bw (units of 64 texels). A PSMT8 page is 128x64 texels; its block table
// is PSMCT32's; a block is 16x16.
func addrPSMT8(bp, bw, x, y uint32) uint32 {
	pagesPerRow := bw / 2
	if pagesPerRow == 0 {
		pagesPerRow = 1
	}
	page := (y/64)*pagesPerRow + x/128
	px, py := x%128, y%64
	block := blockPSMCT32[py/16][px/16]
	bx, by := px%16, py%16
	column := by / 4
	cx, cy := bx, by%4
	cw := columnWordT8[column&1][cy*16+cx]
	cb := columnByteT8[cy*16+cx]
	return bp*256 + page*8192 + uint32(block)*256 + column*64 + uint32(cw)*4 + uint32(cb)
}

// addrPSMT4 returns the byte offset and nibble (0 low, 1 high) of texel (x, y) in a 4-bit
// buffer. A PSMT4 page is 128x128 texels; a block is 32x16.
func addrPSMT4(bp, bw, x, y uint32) (uint32, uint32) {
	pagesPerRow := bw / 2
	if pagesPerRow == 0 {
		pagesPerRow = 1
	}
	page := (y/128)*pagesPerRow + x/128
	px, py := x%128, y%128
	block := blockPSMT4[py/16][px/32]
	bx, by := px%32, py%16
	column := by / 4
	cx, cy := bx, by%4
	cw := columnWordT4[column&1][cy*32+cx]
	cn := columnNibbleT4[cy*32+cx]
	return bp*256 + page*8192 + uint32(block)*256 + column*64 + uint32(cw)*4 + uint32(cn>>1), uint32(cn & 1)
}

// addrPSMZ32 returns the byte offset of (x, y) in a 32-bit Z buffer. The Z formats tile
// exactly like their colour counterparts except the block index has bits 3 and 4
// inverted — the hardware's way of interleaving a frame and its Z so they never collide
// in a page.
func addrPSMZ32(bp, bw, x, y uint32) uint32 {
	if bw == 0 {
		bw = 1
	}
	const pageW, pageH = 64, 32
	page := (y/pageH)*bw + x/pageW
	px, py := x%pageW, y%pageH
	block := blockPSMCT32[py/8][px/8] ^ 0x18
	cx, cy := px%8, py%8
	col := columnPSMCT32[cy][cx]
	return (bp*64 + page*2048 + uint32(block)*64 + uint32(col)) * 4
}

// addrPSMZ16 is the 16-bit Z buffer's swizzle: PSMCT16's with the same block inversion.
// s selects the Z16S block table (PSMCT16S's, inverted).
func addrPSMZ16(bp, bw, x, y uint32, s bool) uint32 {
	if bw == 0 {
		bw = 1
	}
	page := (y/64)*bw + x/64
	px, py := x%64, y%64
	var block uint8
	if s {
		block = blockPSMCT16S[py/8][px/16] ^ 0x18
	} else {
		block = blockPSMCT16[py/8][px/16] ^ 0x18
	}
	bx, by := px%16, py%8
	column := by / 2
	cx, cy := bx, by%2
	cw := columnWordCT16[cy*16+cx]
	ch := columnHalfCT16[cy*16+cx]
	return bp*256 + page*8192 + uint32(block)*256 + column*64 + uint32(cw)*4 + uint32(ch)*2
}

// addrPSMCT16 returns the byte offset of texel (x, y) in a 16-bit buffer. A PSMCT16 page
// is 64x64 texels; a block is 16x8. s selects the CT16S block table.
func addrPSMCT16(bp, bw, x, y uint32, s bool) uint32 {
	if bw == 0 {
		bw = 1
	}
	page := (y/64)*bw + x/64
	px, py := x%64, y%64
	var block uint8
	if s {
		block = blockPSMCT16S[py/8][px/16]
	} else {
		block = blockPSMCT16[py/8][px/16]
	}
	bx, by := px%16, py%8
	column := by / 2
	cx, cy := bx, by%2
	cw := columnWordCT16[cy*16+cx]
	ch := columnHalfCT16[cy*16+cx]
	return bp*256 + page*8192 + uint32(block)*256 + column*64 + uint32(cw)*4 + uint32(ch)*2
}

// --- the CLUT ----------------------------------------------------------------------
//
// Index formats sample through a colour look-up table. The hardware keeps the CLUT in a
// 1 KiB on-chip buffer, loaded from GS memory when TEX0 is written with CLD saying so —
// which means the table a primitive samples is the one loaded, not whatever the memory
// holds at draw time. The buffer is modelled as 512 entry slots (CSA counts 16-entry
// units into it); a CT32 CLUT fills a slot per entry, a CT16 CLUT stores its 16-bit
// entries raw and TEXA expands them at sample time.

// clutLoad performs TEX0's CLD action.
func (gs *GS) clutLoad(t gsTex) {
	switch t.cld {
	case 0:
		return
	case 1:
	case 2:
		gs.cbp0 = t.cbp
	case 3:
		gs.cbp1 = t.cbp
	case 4:
		if t.cbp == gs.cbp0 {
			return
		}
		gs.cbp0 = t.cbp
	case 5:
		if t.cbp == gs.cbp1 {
			return
		}
		gs.cbp1 = t.cbp
	default:
		return
	}

	var n uint32
	switch t.psm {
	case psmT8, psmT8H:
		n = 256
	case psmT4, psmT4HL, psmT4HH:
		n = 16
	default:
		return // not an index format; nothing to load
	}
	base := t.csa * 16
	defer func() {
		// Count the entries that carry colour: a CLUT that loads mostly zeros makes
		// every glyph sample black, and no dump of the first eight entries shows it.
		nz := uint32(0)
		for i := uint32(0); i < n && base+i < uint32(len(gs.clut)); i++ {
			if gs.clut[base+i]&0xFFFFFF != 0 {
				nz++
			}
		}
		gs.count(sprintf("clut load cld=%d cbp=0x%05X cpsm=0x%02X csa=%d for psm 0x%02X rgb-nonzero %d/%d",
			t.cld, t.cbp*64, t.cpsm, t.csa, t.psm, nz, n))
		// The first few entries, once per CLUT: black output with a loaded CLUT is
		// either an empty table or a wrong read, and this line says which.
		gs.m.note("GS: clut at 0x%05X loads %08X %08X %08X %08X %08X %08X %08X %08X",
			t.cbp*64, gs.clut[base], gs.clut[base+1], gs.clut[base+2], gs.clut[base+3],
			gs.clut[base+4], gs.clut[base+5], gs.clut[base+6], gs.clut[base+7])
	}()

	for i := uint32(0); i < n && base+i < uint32(len(gs.clut)); i++ {
		var x, y uint32
		if t.csm == 0 {
			// CSM1: the entries are swizzled into 8x2 half-blocks — 256 entries fill a
			// 16x16 patch, 16 entries an 8x2 one.
			if n == 256 {
				x = i&7 | i&0x10>>1
				y = i>>3&1 | i>>4&0xE
			} else {
				x = i & 7
				y = i >> 3 & 1
			}
		} else {
			// CSM2: linear at the offset TEXCLUT gives, in a buffer TEXCLUT sizes.
			texclut := gs.reg[gsTEXCLUT]
			cou := uint32(texclut>>6) & 0x3F * 16
			cov := uint32(texclut>>12) & 0x3FF
			x = cou + i
			y = cov
		}

		switch t.cpsm {
		case psmCT32:
			bw := uint32(1)
			if t.csm == 1 {
				bw = uint32(gs.reg[gsTEXCLUT]) & 0x3F
			}
			a := addrPSMCT32(t.cbp, bw, x, y)
			if a+4 <= uint32(len(gs.vram)) {
				gs.clut[base+i] = le32gs(gs.vram[a:])
			}
		case psmCT16, psmCT16S:
			bw := uint32(1)
			if t.csm == 1 {
				bw = uint32(gs.reg[gsTEXCLUT]) & 0x3F
			}
			a := addrPSMCT16(t.cbp, bw, x, y, t.cpsm == psmCT16S)
			if a+2 <= uint32(len(gs.vram)) {
				gs.clut[base+i] = uint32(gs.vram[a]) | uint32(gs.vram[a+1])<<8
			}
		}
	}
}

// clutEntry resolves one index through the loaded CLUT, expanding a CT16 entry through
// TEXA the way a CT16 texel would be.
func (gs *GS) clutEntry(t *gsTex, idx uint32) uint32 {
	slot := (t.csa*16 + idx) & 511
	raw := gs.clut[slot]
	if t.cpsm == psmCT16 || t.cpsm == psmCT16S {
		return gs.expand16(raw)
	}
	return raw
}

// expand16 turns a 16-bit texel into RGBA using TEXA: TA0 is the alpha of a pixel whose
// bit 15 is clear (or zero, if AEM says a black pixel is transparent), TA1 of one whose
// bit 15 is set.
func (gs *GS) expand16(px uint32) uint32 {
	texa := gs.reg[gsTEXA]
	r := px & 0x1F << 3
	g := px >> 5 & 0x1F << 3
	b := px >> 10 & 0x1F << 3
	var a uint32
	if px&0x8000 != 0 {
		a = uint32(texa>>32) & 0xFF // TA1
	} else if texa&(1<<15) != 0 && px&0x7FFF == 0 {
		a = 0 // AEM: an all-zero pixel is transparent
	} else {
		a = uint32(texa) & 0xFF // TA0
	}
	return r | g<<8 | b<<16 | a<<24
}

// --- sampling ----------------------------------------------------------------------

// gsSampler is everything a primitive needs to turn a texel coordinate into a colour,
// resolved once per primitive from the context's registers.
type gsSampler struct {
	gs   *GS
	tex  gsTex
	w, h int32
	// The wrap modes and region bounds, from CLAMP.
	wms, wmt               uint32
	minu, maxu, minv, maxv int32
	// linear is TEX1's MMAG bit: this draw asked for bilinear magnification.
	linear bool
	// probe makes the next at() narrate its address math (the -gspixel drill-down).
	probe bool
}

// sampler resolves the current context's texture state, or nil (with the census told)
// when the format is one this machine cannot sample yet — in which case the primitive
// draws as vertex colour, visible and wrong rather than invisible.
func (gs *GS) sampler(p uint64) *gsSampler {
	if p&(1<<4) == 0 { // TME off
		return nil
	}
	ctxt := int(p >> 9 & 1)
	tex0 := gs.reg[gsTEX0_1]
	clamp := gs.reg[gsCLAMP1]
	tex1 := gs.reg[gsTEX1_1]
	if ctxt == 1 {
		tex0 = gs.reg[gsTEX0_2]
		clamp = gs.reg[gsCLAMP2]
		tex1 = gs.reg[gsTEX1_2]
	}
	t := decodeTEX0(tex0)

	switch t.psm {
	case psmCT32, psmCT24, psmCT16, psmCT16S, psmT8, psmT4, psmT8H, psmT4HL, psmT4HH:
	default:
		gs.count(sprintf("textured PSM 0x%02X (vertex colour — unsampled)", t.psm))
		return nil
	}
	return &gsSampler{
		gs: gs, tex: t, linear: tex1>>5&1 != 0,
		w: 1 << t.tw, h: 1 << t.th,
		wms:  uint32(clamp) & 3,
		wmt:  uint32(clamp>>2) & 3,
		minu: int32(uint32(clamp>>4) & 0x3FF),
		maxu: int32(uint32(clamp>>14) & 0x3FF),
		minv: int32(uint32(clamp>>24) & 0x3FF),
		maxv: int32(uint32(clamp>>34) & 0x3FF),
	}
}

// pick samples at a 12.4 texel-space coordinate — the value the rasteriser
// interpolated at the pixel centre. The GS's sample point sits half a texel
// above the coordinate: NEAREST takes floor(u-0.5) and LINEAR blends the 2x2
// texels around (u-0.5, v-0.5) by its fraction, post-CLUT. The half is
// load-bearing and was pinned by RRV's FM-card grade pass, whose three authored
// conventions (gathers at +0.5 with the CT32-as-T8 byte-lane geometry,
// composites at +1.0, smear tiles at +1.5) all become exact reads under it and
// lane-scrambled under floor(u); note that 1:1 draws authored with integer UV
// corners read identically under either rule, which is why everything else
// looked fine while the pun pass shredded.
func (s *gsSampler) pick(u, v int32) uint32 {
	if s.linear {
		us, vs := u-8, v-8
		x0, y0 := us>>4, vs>>4
		fx, fy := uint32(us&15), uint32(vs&15)
		c00 := s.at(x0, y0)
		c10 := s.at(x0+1, y0)
		c01 := s.at(x0, y0+1)
		c11 := s.at(x0+1, y0+1)
		lerp := func(a, b uint32, f uint32) uint32 {
			// per-channel a + (b-a)*f/16, signed so a darker right texel works
			var out uint32
			for sh := 0; sh < 32; sh += 8 {
				av, bv := int32(a>>sh&0xFF), int32(b>>sh&0xFF)
				out |= uint32(av+(bv-av)*int32(f)>>4) & 0xFF << sh
			}
			return out
		}
		top := lerp(c00, c10, fx)
		bot := lerp(c01, c11, fx)
		return lerp(top, bot, fy)
	}
	return s.at((u-8)>>4, (v-8)>>4)
}

// wrap applies one axis's CLAMP mode to a texel coordinate.
func wrapTexel(c, size int32, mode uint32, min, max int32) int32 {
	switch mode {
	case 1: // CLAMP
		if c < 0 {
			c = 0
		}
		if c >= size {
			c = size - 1
		}
		return c
	case 2: // REGION_CLAMP
		if c < min {
			c = min
		}
		if c > max {
			c = max
		}
		return c
	case 3: // REGION_REPEAT: MINU is a mask, MAXU an offset
		return c&min | max
	default: // REPEAT
		return c & (size - 1)
	}
}

// at samples the texel at integer coordinates (u, v).
func (s *gsSampler) at(u, v int32) uint32 {
	u = wrapTexel(u, s.w, s.wms, s.minu, s.maxu)
	v = wrapTexel(v, s.h, s.wmt, s.minv, s.maxv)
	x, y := uint32(u), uint32(v)
	gs, t := s.gs, &s.tex

	switch t.psm {
	case psmCT32:
		a := addrPSMCT32(t.tbp, t.tbw, x, y)
		if a+4 <= uint32(len(gs.vram)) {
			return le32gs(gs.vram[a:])
		}
	case psmCT24:
		a := addrPSMCT32(t.tbp, t.tbw, x, y)
		if a+4 <= uint32(len(gs.vram)) {
			rgb := le32gs(gs.vram[a:]) & 0xFFFFFF
			texa := gs.reg[gsTEXA]
			alpha := uint32(texa) & 0xFF
			if texa&(1<<15) != 0 && rgb == 0 {
				alpha = 0 // AEM
			}
			return rgb | alpha<<24
		}
	case psmCT16, psmCT16S:
		a := addrPSMCT16(t.tbp, t.tbw, x, y, t.psm == psmCT16S)
		if a+2 <= uint32(len(gs.vram)) {
			return gs.expand16(uint32(gs.vram[a]) | uint32(gs.vram[a+1])<<8)
		}
	case psmT8:
		a := addrPSMT8(t.tbp, t.tbw, x, y)
		if a < uint32(len(gs.vram)) {
			idx := uint32(gs.vram[a])
			entry := gs.clutEntry(t, idx)
			// The first few resolved samples, once: a texture whose CLUT is loaded
			// and coloured but whose samples come back black is indexing the wrong
			// entries, and this line shows the index and where it landed.
			if entry&0xFFFFFF == 0 && gs.t8Dumped < 16 {
				gs.t8Dumped++
				gs.m.note("GS: T8 sample tbp 0x%05X (%d,%d) -> idx %d entry %08X (csa %d)",
					t.tbp*64, x, y, idx, entry, t.csa)
			}
			return entry
		}
	case psmT4:
		a, nib := addrPSMT4(t.tbp, t.tbw, x, y)
		if a < uint32(len(gs.vram)) {
			idx := uint32(gs.vram[a]) >> (4 * nib) & 0xF
			if s.probe {
				print(sprintf("    T4 probe (%d,%d) addr 0x%X nib %d idx %d clut(cbp 0x%X csa %d cpsm 0x%X csm %d) entry %08X\n",
					x, y, a, nib, idx, t.cbp, t.csa, t.cpsm, t.csm, gs.clutEntry(t, idx)))
			}
			return gs.clutEntry(t, idx)
		}
	case psmT8H:
		a := addrPSMCT32(t.tbp, t.tbw, x, y)
		if a+4 <= uint32(len(gs.vram)) {
			return gs.clutEntry(t, uint32(gs.vram[a+3]))
		}
	case psmT4HL:
		a := addrPSMCT32(t.tbp, t.tbw, x, y)
		if a+4 <= uint32(len(gs.vram)) {
			return gs.clutEntry(t, uint32(gs.vram[a+3])&0xF)
		}
	case psmT4HH:
		a := addrPSMCT32(t.tbp, t.tbw, x, y)
		if a+4 <= uint32(len(gs.vram)) {
			return gs.clutEntry(t, uint32(gs.vram[a+3])>>4)
		}
	}
	return 0
}

// GSTexture renders a texture exactly as a primitive would sample it — through the real
// swizzle, the CLUT load TEX0 commands, and TEXA — and returns RGBA pixels. tex0 is the
// raw register word (the draw census prints it per distinct texture), so the dump and
// the draw share one source of truth. The instrument that says whether a flat-looking
// primitive is sampling a flat texture or sampling a real texture wrongly.
func (m *Machine) GSTexture(tex0 uint64) (pix []byte, w, h int) {
	gs := m.ensureGS()
	t := decodeTEX0(tex0)
	if t.cld == 0 {
		t.cld = 1 // load its CLUT the way the first draw's TEX0 write would
	}
	gs.clutLoad(t)
	s := &gsSampler{gs: gs, tex: t, w: 1 << t.tw, h: 1 << t.th}
	w, h = int(s.w), int(s.h)
	pix = make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := s.at(int32(x), int32(y))
			o := (y*w + x) * 4
			pix[o+0] = byte(c)
			pix[o+1] = byte(c >> 8)
			pix[o+2] = byte(c >> 16)
			pix[o+3] = byte(c >> 24)
		}
	}
	return pix, w, h
}

// combine applies the texture function (TFX/TCC): how the sampled texel and the
// fragment's colour make the pixel the tests and the blender see. The GS's colour scale
// puts 1.0 at 0x80, so a modulate is >>7 with a clamp at 255.
func (s *gsSampler) combine(tex, frag uint32) uint32 {
	if tex&0xFFFFFF == 0 {
		s.gs.texBlackPSM[s.tex.psm&0x3F]++
		s.gs.texBlack++
	} else {
		s.gs.texColorPSM[s.tex.psm&0x3F]++
		s.gs.texColor++
	}
	tr, tg, tb, ta := unpackRGBA(tex)
	fr, fg, fb, fa := unpackRGBA(frag)
	var r, g, b, a int64
	switch s.tex.tfx {
	case 0: // MODULATE
		r, g, b = clamp255(tr*fr>>7), clamp255(tg*fg>>7), clamp255(tb*fb>>7)
		if s.tex.tcc {
			a = clamp255(ta * fa >> 7)
		} else {
			a = fa
		}
	case 1: // DECAL
		r, g, b = tr, tg, tb
		if s.tex.tcc {
			a = ta
		} else {
			a = fa
		}
	case 2: // HIGHLIGHT
		r, g, b = clamp255(tr*fr>>7+fa), clamp255(tg*fg>>7+fa), clamp255(tb*fb>>7+fa)
		if s.tex.tcc {
			a = clamp255(ta + fa)
		} else {
			a = fa
		}
	case 3: // HIGHLIGHT2
		r, g, b = clamp255(tr*fr>>7+fa), clamp255(tg*fg>>7+fa), clamp255(tb*fb>>7+fa)
		if s.tex.tcc {
			a = ta
		} else {
			a = fa
		}
	}
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
}

func clamp255(v int64) int64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// --- blending ----------------------------------------------------------------------

// blendPixel evaluates ALPHA's equation Cv = ((A - B) * C >> 7) + D for one pixel.
// A, B and D select from {source, destination, zero}; C from {source alpha, destination
// alpha, the FIX constant}. COLCLAMP chooses clamping or wrap-around for the result.
func blendPixel(alpha uint64, src, dst uint32, dstA int64, colclamp bool) uint32 {
	selA := alpha & 3
	selB := alpha >> 2 & 3
	selC := alpha >> 4 & 3
	selD := alpha >> 6 & 3
	fix := int64(alpha>>32) & 0xFF

	sr, sg, sb, sa := unpackRGBA(src)
	dr, dg, db, _ := unpackRGBA(dst)

	pick := func(sel uint64) (int64, int64, int64) {
		switch sel {
		case 0:
			return sr, sg, sb
		case 1:
			return dr, dg, db
		}
		return 0, 0, 0
	}
	ar, ag, ab := pick(selA)
	br, bg, bb := pick(selB)
	drr, dgg, dbb := pick(selD)

	var c int64
	switch selC {
	case 0:
		c = sa
	case 1:
		c = dstA
	default:
		c = fix
	}

	out := func(a, b, d int64) int64 {
		v := (a-b)*c>>7 + d
		if colclamp {
			return clamp255(v)
		}
		return v & 0xFF
	}
	return uint32(out(ar, br, drr)) | uint32(out(ag, bg, dgg))<<8 | uint32(out(ab, bb, dbb))<<16 | uint32(sa)<<24
}
