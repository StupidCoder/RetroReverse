package dsmachine

// The two 2D graphics engines.
//
// A DS has two of them — A and B — and they are not two views of one chip: they
// are two nearly identical display controllers, each with its own register block
// (A at 0x04000000, B at 0x04001000, same layout), its own half of palette RAM and
// OAM, and its own VRAM address space. Engine B is the poor relation: no 3D layer,
// no bitmap/"large" background modes, no display-mode 2, no global character/screen
// base offsets, and a quarter of the background VRAM. Writing one engine and
// parameterising it is right; writing one engine and *pretending* the differences
// away is how engine B ends up quietly reading engine A's VRAM.
//
// Which engine you are looking at is not a property of the engine. POWCNT1 bit 15
// says which physical screen each one drives, and a game may flip it — so "the top
// screen" is a question only screens() can answer, never a field on an engine.
//
// The pipeline here is the hardware's, in the hardware's order: for each scanline,
// render the four backgrounds and the sprites into separate layers, work out what
// the windows allow at each pixel, resolve priority, and only then apply the colour
// special effects and the master brightness. Compositing first and blending later
// is the shortcut that makes semi-transparent sprites and the 3D layer's per-pixel
// alpha impossible to express, because both of them need the pixel *underneath*.
//
// Colour is BGR555 throughout, right up to the final pack: the DS blends in 5-bit
// (in fact 6-bit) components, and converting to 8-bit early and blending there is a
// good way to get colours that are close but never equal.
//
// Pixels leave here as RGBA8888 packed R<<24 | G<<16 | B<<8 | A — the same packing
// the 3D engine must use for the frame it hands us in gpu2d.threeD.

const (
	screenW = 256
	screenH = 192
)

// Register offsets from an engine's base (add 0x1000 for engine B).
const (
	rDISPCNT  = 0x00
	rBG0CNT   = 0x08
	rBG0HOFS  = 0x10
	rBG2PA    = 0x20
	rWIN0H    = 0x40
	rWIN0V    = 0x44
	rWININ    = 0x48
	rWINOUT   = 0x4A
	rMOSAIC   = 0x4C
	rBLDCNT   = 0x50
	rBLDALPHA = 0x52
	rBLDY     = 0x54
	rDISPCAP  = 0x64
	rMASTERBR = 0x6C
)

// The layer identities BLDCNT's target bits use: BG0..BG3, then the sprites, then
// the backdrop.
const (
	lyBG0 = iota
	lyBG1
	lyBG2
	lyBG3
	lyOBJ
	lyBackdrop
)

// The kinds of background the BG-mode table can name.
const (
	kNone = iota
	kText
	kAffine
	kExtended // extended-affine: 16-bit tiled, 8-bit bitmap, or direct-colour bitmap
	kLarge    // the mode-6 large bitmap (engine A only)
)

// bgKind is DISPCNT's BG-mode table: which of the four backgrounds exists in each
// mode, and what kind it is. Mode 6 is the odd one — BG0 is the 3D layer and BG2 is
// a single large bitmap, and it exists on engine A only. Mode 7 does not exist.
var bgKind = [8][4]int{
	0: {kText, kText, kText, kText},
	1: {kText, kText, kText, kAffine},
	2: {kText, kText, kAffine, kAffine},
	3: {kText, kText, kText, kExtended},
	4: {kText, kText, kAffine, kExtended},
	5: {kText, kText, kExtended, kExtended},
	6: {kNone, kNone, kLarge, kNone}, // BG0 here is the 3D layer, handled separately
	7: {kNone, kNone, kNone, kNone},
}

// gpu2d owns both engines and the buffers they scan out into.
type gpu2d struct {
	a, b engine

	// threeD is the 3D engine's most recent rendered frame: 256x192 pixels, RGBA8888
	// with a meaningful alpha byte (0 = nothing was drawn there, 255 = opaque, in
	// between = a translucent polygon). A nil slice means the 3D engine has not drawn
	// a frame yet, in which case engine A's BG0 stays empty even with DISPCNT bit 3
	// set. Set by gpu3d.go; only read here.
	threeD []uint32

	// swap latches POWCNT1 bit 15 at the moment the frame was composed, so screens()
	// answers for the frame it is handing back rather than for whatever the CPU has
	// written since.
	swap bool
}

// engine is one of the two display controllers: where its registers, palette, OAM
// and VRAM live, plus the per-scanline scratch the compositor works in.
type engine struct {
	m   *Machine
	isB bool

	ioBase   uint32 // 0x04000000 or 0x04001000
	palBG    uint32 // byte offset of this engine's BG palette in Machine.pal
	palOBJ   uint32
	oamBase  uint32
	spBG     int // the VRAM address spaces this engine reads
	spOBJ    int
	spBGExt  int
	spOBJExt int

	out    []uint32 // 256x192 RGBA8888, this engine's scanned-out frame
	threeD []uint32 // engine A only; nil on B, which has no 3D layer

	// Per-scanline layers. Kept as BGR555 plus an opaque flag rather than as RGBA,
	// because every effect downstream works in 5-bit components.
	bg    [4][screenW]uint16
	bgOK  [4][screenW]bool
	a3D   [screenW]uint8 // BG0's per-pixel 3D alpha (0..31) when the 3D layer drives it
	is3D  bool           // BG0 is the 3D layer this frame
	prio  [4]uint8       // each background's priority, from BGxCNT
	shown [4]bool        // ...and whether DISPCNT enables it in this mode at all

	obj     [screenW]uint16
	objOK   [screenW]bool
	objPrio [screenW]uint8
	objSemi [screenW]bool  // semi-transparent sprite: blends with whatever is below it
	objEVA  [screenW]uint8 // a bitmap sprite's own blend coefficient (0 = use BLDALPHA)
	objWin  [screenW]bool  // the OBJ window's shape, painted by gfx-mode-2 sprites

	win  [screenW]uint8 // per-pixel window mask: bits 0..4 layers, bit 5 effects
	line [screenW]uint16
}

func newGPU2D() *gpu2d {
	g := &gpu2d{}
	g.a = engine{isB: false, ioBase: 0x04000000,
		palBG: 0x000, palOBJ: 0x200, oamBase: 0x000,
		spBG: spBGA, spOBJ: spOBJA, spBGExt: spBGExtA, spOBJExt: spOBJExtA,
		out: make([]uint32, screenW*screenH)}
	g.b = engine{isB: true, ioBase: 0x04001000,
		palBG: 0x400, palOBJ: 0x600, oamBase: 0x400,
		spBG: spBGB, spOBJ: spOBJB, spBGExt: spBGExtB, spOBJExt: spOBJExtB,
		out: make([]uint32, screenW*screenH)}
	return g
}

// beginFrame is called at the top of scanline 0. The frame it composes is the one
// the CPU finished setting up during the vertical blank that just ended — we render
// a whole frame from one snapshot of the registers rather than scanning out line by
// line. That is a real simplification, and it costs us mid-frame register changes:
// a game that rewrites a scroll register or an affine reference point on an HBlank
// interrupt (a raster split, a road that curves) will render as if it had not. The
// cost is visible and bounded; the alternative — running the compositor from the
// scanline loop — is the next step if a title needs it.
func (g *gpu2d) beginFrame(m *Machine) { g.render(m) }

// render composes both engines' output for the frame into their 256x192 buffers.
func (g *gpu2d) render(m *Machine) {
	g.swap = m.powcnt&(1<<15) != 0
	g.a.threeD = g.threeD
	g.a.frame(m)
	g.b.frame(m)
}

// screens returns the top and bottom screens as 256x192 RGBA8888 pixels. POWCNT1
// bit 15 decides which engine is which: set, engine A drives the top screen; clear,
// engine A drives the bottom. SM64DS is not the only game that puts its main view on
// engine B for a menu, so this is not a constant.
func (g *gpu2d) screens() (top, bottom []uint32) {
	if g.swap {
		return g.a.out, g.b.out
	}
	return g.b.out, g.a.out
}

// --- register access --------------------------------------------------------
//
// Both engines' registers live in the ARM9's register file, keyed by word-aligned
// address and holding the last value written (io.go latches them and does nothing
// else — the engines are read at scanout, which is here).

func (e *engine) reg32(off uint32) uint32 { return e.m.ARM9.io[e.ioBase+off] }

func (e *engine) reg16(off uint32) uint16 {
	w := e.m.ARM9.io[e.ioBase+(off&^3)]
	if off&2 != 0 {
		return uint16(w >> 16)
	}
	return uint16(w)
}

// --- colour -----------------------------------------------------------------

// bgr555RGBA expands a DS colour to RGBA8888. Five bits become eight by replicating
// the top three back into the bottom, which is the conversion that maps 31 to 255
// (a plain <<3 tops out at 248 and tints every white in the game grey).
func bgr555RGBA(c uint16) uint32 {
	r := uint32(c & 31)
	g := uint32(c>>5) & 31
	b := uint32(c>>10) & 31
	r = r<<3 | r>>2
	g = g<<3 | g>>2
	b = b<<3 | b>>2
	return r<<24 | g<<16 | b<<8 | 0xFF
}

func rgb555(r, g, b uint32) uint16 { return uint16(r&31 | (g&31)<<5 | (b&31)<<10) }

func chan5(c uint16, n uint) uint32 { return uint32(c>>(5*n)) & 31 }

// palBGColor and palOBJColor read this engine's half of palette RAM. Palette RAM is
// one 2 KiB block for both engines: A's BG palette at 0x000, A's OBJ at 0x200, B's
// BG at 0x400, B's OBJ at 0x600.
func (e *engine) palBGColor(i int) uint16  { return palAt(e.m.pal, e.palBG+uint32(i)*2) }
func (e *engine) palOBJColor(i int) uint16 { return palAt(e.m.pal, e.palOBJ+uint32(i)*2) }

func palAt(pal []byte, off uint32) uint16 {
	if int(off)+1 >= len(pal) {
		return 0
	}
	return uint16(pal[off]) | uint16(pal[off+1])<<8
}

// bgExtColor reads a background extended-palette entry. The extended-palette space
// is four 8 KiB slots, each holding 16 palettes of 256 colours — so a slot is not a
// palette, it is a *bank* of sixteen, and the tile's 4-bit palette field still picks
// among them even though the tile is 8bpp.
func (e *engine) bgExtColor(slot, pal, idx int) uint16 {
	off := uint32(slot)*0x2000 + uint32(pal)*512 + uint32(idx)*2
	return e.m.vram.read16(e.spBGExt, off)
}

// objExtColor reads a sprite extended-palette entry: one 8 KiB slot, 16 palettes of
// 256 colours.
func (e *engine) objExtColor(pal, idx int) uint16 {
	return e.m.vram.read16(e.spOBJExt, uint32(pal)*512+uint32(idx)*2)
}

// --- the frame --------------------------------------------------------------

func (e *engine) frame(m *Machine) {
	e.m = m

	// POWCNT1 gates each engine's clock (bit 1 = engine A, bit 9 = engine B). An
	// unpowered engine drives nothing, and an LCD with nothing driving it is white,
	// not black — the same white as display-mode 0.
	powerBit := uint32(1) << 1
	if e.isB {
		powerBit = 1 << 9
	}
	if m.powcnt&powerBit == 0 {
		e.fillWhite()
		return
	}

	dispcnt := e.reg32(rDISPCNT)

	// Forced blank (DISPCNT bit 7): the display outputs white and stops fetching from
	// VRAM entirely. Games set it while they re-point VRAM banks.
	if dispcnt&(1<<7) != 0 {
		e.fillWhite()
		return
	}

	switch (dispcnt >> 16) & 3 {
	case 0: // display off
		e.fillWhite()
		return
	case 1:
		e.graphics(dispcnt)
	case 2:
		// VRAM display: a raw 16-bit bitmap straight out of one LCDC bank, bypassing
		// the whole 2D pipeline. Engine B does not have it — its DISPCNT only has the
		// two low display-mode bits wired.
		if e.isB {
			m.note("2D engine B: display mode 2 (VRAM display) does not exist on engine B")
			e.fillWhite()
			return
		}
		e.vramDisplay(dispcnt)
	case 3:
		// Main-memory display: the display is fed by a DMA from main RAM rather than
		// from VRAM. Nothing in this boot uses it, and rendering something plausible
		// would be worse than rendering nothing.
		m.note("2D engine %s: display mode 3 (main-memory display) not implemented", e.name())
		e.fillBlack()
		return
	}

	if e.reg32(rDISPCAP)&(1<<31) != 0 {
		m.note("2D engine %s: display capture (DISPCAPCNT) not implemented", e.name())
	}
	e.masterBright()
}

func (e *engine) name() string {
	if e.isB {
		return "B"
	}
	return "A"
}

func (e *engine) fillWhite() {
	for i := range e.out {
		e.out[i] = 0xFFFFFFFF
	}
}

func (e *engine) fillBlack() {
	for i := range e.out {
		e.out[i] = 0x000000FF
	}
}

// vramDisplay scans out an LCDC bank as a 256x192 array of BGR555 pixels. DISPCNT
// bits 18-19 select the bank (A, B, C or D), and it is read as the *bank*, not
// through a display space: the point of this mode is to show memory the engine is
// not otherwise mapped to see. The 15th bit of each pixel is ignored — everything is
// opaque here; there is nothing underneath to be transparent against.
func (e *engine) vramDisplay(dispcnt uint32) {
	bank := e.m.vram.bank[(dispcnt>>18)&3]
	for y := 0; y < screenH; y++ {
		for x := 0; x < screenW; x++ {
			off := (y*screenW + x) * 2
			var c uint16
			if off+1 < len(bank) {
				c = uint16(bank[off]) | uint16(bank[off+1])<<8
			}
			e.line[x] = c
		}
		e.emit(y)
	}
}

// graphics is the real path: the four backgrounds, the sprites, the windows, the
// priority resolution and the colour effects, one scanline at a time.
func (e *engine) graphics(dispcnt uint32) {
	mode := int(dispcnt & 7)
	if mode == 7 {
		e.m.note("2D engine %s: BG mode 7 does not exist", e.name())
	}
	if mode == 6 && e.isB {
		e.m.note("2D engine B: BG mode 6 (large bitmap) does not exist on engine B")
	}

	// The 3D layer takes over BG0 when DISPCNT bit 3 is set, and it is engine A's
	// alone. In mode 6 BG0 *is* the 3D layer and nothing else.
	e.is3D = !e.isB && dispcnt&(1<<3) != 0
	if e.is3D && e.threeD == nil {
		e.m.note("2D engine A: BG0 is the 3D layer but the 3D engine has not drawn a frame")
	}

	for n := 0; n < 4; n++ {
		cnt := e.reg16(rBG0CNT + uint32(n)*2)
		e.prio[n] = uint8(cnt & 3)
		e.shown[n] = dispcnt&(1<<(8+uint(n))) != 0
	}

	for y := 0; y < screenH; y++ {
		e.clearLayers()
		if dispcnt&(1<<12) != 0 || dispcnt&(1<<15) != 0 {
			// Sprites are drawn when OBJ is enabled, and also when only the OBJ *window*
			// is enabled: window-shape sprites paint a mask, not pixels.
			e.sprites(dispcnt, y)
		}
		for n := 0; n < 4; n++ {
			if !e.shown[n] {
				continue
			}
			e.background(dispcnt, mode, n, y)
		}
		e.windowMask(dispcnt, y)
		e.composite(y)
		e.emit(y)
	}
}

func (e *engine) clearLayers() {
	for n := 0; n < 4; n++ {
		for x := 0; x < screenW; x++ {
			e.bgOK[n][x] = false
		}
	}
	for x := 0; x < screenW; x++ {
		e.a3D[x] = 0
		e.objOK[x] = false
		e.objSemi[x] = false
		e.objEVA[x] = 0
		e.objWin[x] = false
		e.objPrio[x] = 3
	}
}

// emit converts the finished 15-bit scanline into the output buffer, applying the
// master brightness on the way (see masterBright's note on where it belongs).
func (e *engine) emit(y int) {
	for x := 0; x < screenW; x++ {
		e.out[y*screenW+x] = bgr555RGBA(e.line[x])
	}
}

// masterBright applies MASTER_BRIGHT to the whole finished frame. It is the very
// last stage of the display pipeline, after every blend — which is exactly why a
// fade-to-black done with it does not disturb any alpha the frame contains, and why
// a title sequence that fades in is not a hint that the game is drawing anything
// differently. Mode 1 fades towards white, mode 2 towards black; the factor is 0..16
// (values above 16 saturate at 16, so a game can write 31 and mean "all the way").
func (e *engine) masterBright() {
	v := e.reg16(rMASTERBR)
	mode := (v >> 14) & 3
	if mode == 0 || mode == 3 {
		return
	}
	f := uint32(v & 31)
	if f > 16 {
		f = 16
	}
	if f == 0 {
		return
	}
	for i, p := range e.out {
		r, g, b := p>>24&0xFF, p>>16&0xFF, p>>8&0xFF
		if mode == 1 {
			r += (255 - r) * f / 16
			g += (255 - g) * f / 16
			b += (255 - b) * f / 16
		} else {
			r -= r * f / 16
			g -= g * f / 16
			b -= b * f / 16
		}
		e.out[i] = r<<24 | g<<16 | b<<8 | 0xFF
	}
}

// --- backgrounds ------------------------------------------------------------

func (e *engine) background(dispcnt uint32, mode, n, y int) {
	// BG0 as the 3D layer overrides whatever the mode table says it is.
	if n == 0 && e.is3D {
		e.threeDLine(y)
		return
	}
	if n == 0 && mode == 6 {
		return // in mode 6 BG0 is the 3D layer or it is nothing
	}
	switch bgKind[mode][n] {
	case kText:
		e.textBG(dispcnt, n, y)
	case kAffine:
		e.affineBG(dispcnt, n, y)
	case kExtended:
		e.extendedBG(dispcnt, n, y)
	case kLarge:
		if !e.isB {
			e.largeBG(n, y)
		}
	}
}

// threeDLine lifts one scanline out of the 3D engine's frame. The 3D layer scrolls
// horizontally with BG0HOFS (and wraps), but not vertically — there is no BG0VOFS
// applied to it on hardware. Its colour comes down to 5 bits per component here,
// which is where the display pipeline puts it anyway.
func (e *engine) threeDLine(y int) {
	if e.threeD == nil {
		return
	}
	hofs := int(e.reg16(rBG0HOFS)) & 0x1FF
	for x := 0; x < screenW; x++ {
		p := e.threeD[y*screenW+(x+hofs)&255]
		a := uint8(p & 0xFF)
		if a == 0 {
			continue // nothing was drawn: the layer is transparent here
		}
		r, g, b := (p>>24)&0xFF, (p>>16)&0xFF, (p>>8)&0xFF
		e.bg[0][x] = rgb555(r>>3, g>>3, b>>3)
		e.bgOK[0][x] = true
		e.a3D[x] = uint8(uint32(a) * 31 / 255)
	}
}

// mosaicH returns the BG mosaic's horizontal cell width, and vertical likewise.
// Only BG mosaic is modelled: sprite mosaic (OAM attr0 bit 12) is not, and is noted
// where it is met.
func (e *engine) mosaic() (h, v int) {
	m := e.reg16(rMOSAIC)
	return int(m&0xF) + 1, int(m>>4&0xF) + 1
}

// bgBases resolves a background's character and screen base addresses. On engine A
// they are BGxCNT's per-background bases *plus* DISPCNT's global offsets (bits 24-26
// and 27-29, in 64 KiB units) — a global base engine B does not have at all, because
// engine B's whole BG space is 128 KiB. Forget the global offsets and every tile on
// engine A comes out of the wrong place in a game that uses them; apply them on
// engine B and you index past the end of its space.
func (e *engine) bgBases(dispcnt uint32, cnt uint16) (charBase, screenBase uint32) {
	charBase = uint32(cnt>>2&0xF) * 0x4000
	screenBase = uint32(cnt>>8&0x1F) * 0x800
	if !e.isB {
		charBase += (dispcnt >> 24 & 7) * 0x10000
		screenBase += (dispcnt >> 27 & 7) * 0x10000
	}
	return
}

// textBG renders one scanline of a tiled background: the classic four screen sizes,
// 4bpp or 8bpp tiles, per-tile palette and flip.
func (e *engine) textBG(dispcnt uint32, n, y int) {
	cnt := e.reg16(rBG0CNT + uint32(n)*2)
	charBase, screenBase := e.bgBases(dispcnt, cnt)
	hofs := int(e.reg16(rBG0HOFS+uint32(n)*4)) & 0x1FF
	vofs := int(e.reg16(rBG0HOFS+uint32(n)*4+2)) & 0x1FF

	w, h := 256, 256
	if cnt&0x4000 != 0 {
		w = 512
	}
	if cnt&0x8000 != 0 {
		h = 512
	}
	bpp8 := cnt&0x80 != 0

	// Extended palettes only exist for 8bpp tiles. The slot is not simply the BG
	// number: BG0 and BG1 can be moved to slots 2 and 3 by BGxCNT bit 13, which is the
	// bit that lets a game give its two text layers 16 palettes each.
	extPal := bpp8 && dispcnt&(1<<30) != 0
	slot := n
	if n < 2 && cnt&0x2000 != 0 {
		slot = n + 2
	}

	mx, my := 1, 1
	if cnt&0x40 != 0 {
		mx, my = e.mosaic()
	}

	sy := y
	if my > 1 {
		sy -= sy % my
	}
	sy = (sy + vofs) & (h - 1)
	ty, py := sy/8%32, sy%8

	for x := 0; x < screenW; x++ {
		sx := x
		if mx > 1 {
			sx -= sx % mx
		}
		sx = (sx + hofs) & (w - 1)

		// The map is a grid of 256x256-pixel blocks of 32x32 entries, 2 KiB each, laid
		// out left-to-right then top-to-bottom.
		block := sx/256 + (sy/256)*(w/256)
		off := screenBase + uint32(block)*0x800 + uint32(ty*32+sx/8%32)*2
		ent := e.m.vram.read16(e.spBG, off)

		tile := uint32(ent & 0x3FF)
		fx, fy := sx%8, py
		if ent&0x400 != 0 {
			fx = 7 - fx
		}
		if ent&0x800 != 0 {
			fy = 7 - fy
		}

		var idx int
		if bpp8 {
			idx = int(e.m.vram.read8(e.spBG, charBase+tile*64+uint32(fy*8+fx)))
		} else {
			b := e.m.vram.read8(e.spBG, charBase+tile*32+uint32(fy*4+fx/2))
			if fx&1 != 0 {
				idx = int(b >> 4)
			} else {
				idx = int(b & 0xF)
			}
		}
		if idx == 0 {
			continue // colour 0 is transparent in every paletted mode
		}
		pal := int(ent >> 12)
		switch {
		case extPal:
			e.bg[n][x] = e.bgExtColor(slot, pal, idx)
		case bpp8:
			e.bg[n][x] = e.palBGColor(idx)
		default:
			e.bg[n][x] = e.palBGColor(pal*16 + idx)
		}
		e.bgOK[n][x] = true
	}
}

// affineParams reads a rotation/scaling background's matrix and reference point.
// BG2's block is at +0x20 and BG3's at +0x30, so only BG2 and BG3 can be affine.
//
// The reference point is a 28-bit signed 20.8 fixed-point value and the matrix
// entries are 16-bit signed 8.8 — sign-extend either of them wrong and the layer
// does not wobble, it flies off the screen.
func (e *engine) affineParams(n int) (pa, pb, pc, pd, x0, y0 int32) {
	base := rBG2PA + uint32(n-2)*0x10
	pa = int32(int16(e.reg16(base)))
	pb = int32(int16(e.reg16(base + 2)))
	pc = int32(int16(e.reg16(base + 4)))
	pd = int32(int16(e.reg16(base + 6)))
	x0 = int32(e.reg32(base+8)<<4) >> 4
	y0 = int32(e.reg32(base+12)<<4) >> 4
	return
}

// affineOrigin returns the texel-space origin of scanline y. On hardware BGxX/BGxY
// are *internal* registers: latched at the top of the frame and advanced by (PB, PD)
// every scanline. We render the frame from a single snapshot, so we compute the same
// thing directly — which is equivalent as long as the game does not rewrite the
// reference point mid-frame.
func affineOrigin(pb, pd, x0, y0 int32, y int) (int32, int32) {
	return x0 + pb*int32(y), y0 + pd*int32(y)
}

// affineBG renders a rotation/scaling background: 8-bit tile indices in the map,
// 8bpp tiles, no per-tile palette and no flip — the map entry is one byte and there
// is nowhere to put them.
func (e *engine) affineBG(dispcnt uint32, n, y int) {
	cnt := e.reg16(rBG0CNT + uint32(n)*2)
	charBase, screenBase := e.bgBases(dispcnt, cnt)
	size := int(cnt >> 14 & 3)
	dim := 128 << uint(size) // 128, 256, 512 or 1024 pixels square
	tiles := dim / 8         // ...and that many tiles across
	wrap := cnt&0x2000 != 0  // the display-area-overflow bit: wrap, or leave it blank
	pa, pb, pc, pd, x0, y0 := e.affineParams(n)
	ox, oy := affineOrigin(pb, pd, x0, y0, y)

	for x := 0; x < screenW; x++ {
		px := int((ox + pa*int32(x)) >> 8)
		py := int((oy + pc*int32(x)) >> 8)
		var ok bool
		px, py, ok = wrapAffine(px, py, dim, dim, wrap)
		if !ok {
			continue
		}
		tile := uint32(e.m.vram.read8(e.spBG, screenBase+uint32((py/8)*tiles+px/8)))
		idx := e.m.vram.read8(e.spBG, charBase+tile*64+uint32((py%8)*8+px%8))
		if idx == 0 {
			continue
		}
		e.bg[n][x] = e.palBGColor(int(idx))
		e.bgOK[n][x] = true
	}
}

// wrapAffine applies the display-area-overflow rule to a sampled texel coordinate:
// with the bit set the layer repeats, without it the area outside is transparent.
// This is a per-background bit (BGxCNT bit 13), not a global one — a game can have a
// tiling floor and a non-tiling sprite plane in the same frame.
func wrapAffine(x, y, w, h int, wrap bool) (int, int, bool) {
	if wrap {
		return ((x % w) + w) % w, ((y % h) + h) % h, true
	}
	if x < 0 || x >= w || y < 0 || y >= h {
		return 0, 0, false
	}
	return x, y, true
}

// extendedBG renders an extended-affine background, which is really three different
// backgrounds wearing one name. BGxCNT bit 7 chooses between tiles and a bitmap, and
// for a bitmap BGxCNT bit 2 — a bit that means "character base" everywhere else —
// chooses between 8-bit paletted and direct 16-bit colour.
func (e *engine) extendedBG(dispcnt uint32, n, y int) {
	cnt := e.reg16(rBG0CNT + uint32(n)*2)
	if cnt&0x80 == 0 {
		e.extTiledBG(dispcnt, n, y, cnt)
		return
	}
	e.extBitmapBG(n, y, cnt)
}

// extTiledBG is the extended mode's tiled form: an affine background whose map
// entries are 16 bits, so the tiles get back the palette and flip bits a plain
// affine background has no room for. The tiles are always 8bpp, and the palette
// field selects among the extended palette's sixteen when they are enabled.
func (e *engine) extTiledBG(dispcnt uint32, n, y int, cnt uint16) {
	charBase, screenBase := e.bgBases(dispcnt, cnt)
	size := int(cnt >> 14 & 3)
	dim := 128 << uint(size)
	tiles := dim / 8
	wrap := cnt&0x2000 != 0
	extPal := dispcnt&(1<<30) != 0
	pa, pb, pc, pd, x0, y0 := e.affineParams(n)
	ox, oy := affineOrigin(pb, pd, x0, y0, y)

	for x := 0; x < screenW; x++ {
		px := int((ox + pa*int32(x)) >> 8)
		py := int((oy + pc*int32(x)) >> 8)
		var ok bool
		px, py, ok = wrapAffine(px, py, dim, dim, wrap)
		if !ok {
			continue
		}
		ent := e.m.vram.read16(e.spBG, screenBase+uint32((py/8)*tiles+px/8)*2)
		tile := uint32(ent & 0x3FF)
		fx, fy := px%8, py%8
		if ent&0x400 != 0 {
			fx = 7 - fx
		}
		if ent&0x800 != 0 {
			fy = 7 - fy
		}
		idx := int(e.m.vram.read8(e.spBG, charBase+tile*64+uint32(fy*8+fx)))
		if idx == 0 {
			continue
		}
		if extPal {
			// BG2 and BG3 always take slots 2 and 3 here: the slot-shifting bit 13 is the
			// overflow bit for an affine background, and cannot mean both.
			e.bg[n][x] = e.bgExtColor(n, int(ent>>12), idx)
		} else {
			e.bg[n][x] = e.palBGColor(idx)
		}
		e.bgOK[n][x] = true
	}
}

// extBmpSize is the extended bitmap's size table — and it is *not* the affine size
// table: 512x256 sits where an affine background would be 512x512.
var extBmpSize = [4][2]int{{128, 128}, {256, 256}, {512, 256}, {512, 512}}

// extBitmapBG renders an extended-affine bitmap background. The trap here is the
// base address: a bitmap's data starts at BGxCNT's *screen* base field scaled by
// 16 KiB, not by the 2 KiB a tile map uses, and the character base field is not used
// at all (except for its bit 2, which has been repurposed as the colour-depth flag).
func (e *engine) extBitmapBG(n, y int, cnt uint16) {
	base := uint32(cnt>>8&0x1F) * 0x4000
	direct := cnt&4 != 0
	size := int(cnt >> 14 & 3)
	w, h := extBmpSize[size][0], extBmpSize[size][1]
	wrap := cnt&0x2000 != 0
	pa, pb, pc, pd, x0, y0 := e.affineParams(n)
	ox, oy := affineOrigin(pb, pd, x0, y0, y)

	for x := 0; x < screenW; x++ {
		px := int((ox + pa*int32(x)) >> 8)
		py := int((oy + pc*int32(x)) >> 8)
		var ok bool
		px, py, ok = wrapAffine(px, py, w, h, wrap)
		if !ok {
			continue
		}
		if direct {
			// Direct colour: bit 15 is the alpha/opacity flag, not a colour bit. A pixel
			// with it clear is transparent however colourful the other fifteen bits are.
			c := e.m.vram.read16(e.spBG, base+uint32(py*w+px)*2)
			if c&0x8000 == 0 {
				continue
			}
			e.bg[n][x] = c & 0x7FFF
		} else {
			idx := e.m.vram.read8(e.spBG, base+uint32(py*w+px))
			if idx == 0 {
				continue
			}
			e.bg[n][x] = e.palBGColor(int(idx))
		}
		e.bgOK[n][x] = true
	}
}

// largeBG renders mode 6's large bitmap: a single 8-bit paletted, affine-transformed
// image that starts at the bottom of BG VRAM and has no base register at all. Only
// two sizes exist, and they are the only place the DS lets a background be bigger
// than 1024 pixels in a dimension.
func (e *engine) largeBG(n, y int) {
	cnt := e.reg16(rBG0CNT + uint32(n)*2)
	w, h := 512, 1024
	if cnt&0x4000 != 0 {
		w, h = 1024, 512
	}
	wrap := cnt&0x2000 != 0
	pa, pb, pc, pd, x0, y0 := e.affineParams(n)
	ox, oy := affineOrigin(pb, pd, x0, y0, y)

	for x := 0; x < screenW; x++ {
		px := int((ox + pa*int32(x)) >> 8)
		py := int((oy + pc*int32(x)) >> 8)
		var ok bool
		px, py, ok = wrapAffine(px, py, w, h, wrap)
		if !ok {
			continue
		}
		idx := e.m.vram.read8(e.spBG, uint32(py*w+px))
		if idx == 0 {
			continue
		}
		e.bg[n][x] = e.palBGColor(int(idx))
		e.bgOK[n][x] = true
	}
}

// --- sprites ----------------------------------------------------------------

// objSize is OAM's shape/size table: attr0's shape (square, wide, tall) crossed with
// attr1's size gives the twelve sprite dimensions.
var objSize = [3][4][2]int{
	{{8, 8}, {16, 16}, {32, 32}, {64, 64}}, // square
	{{16, 8}, {32, 8}, {32, 16}, {64, 32}}, // wide
	{{8, 16}, {8, 32}, {16, 32}, {32, 64}}, // tall
}

// sprites renders one scanline of OAM. It walks the 128 entries backwards, so that
// when two sprites of the *same* priority overlap, the lower OAM index — which wins
// on hardware — is the one written last. The priority number itself is checked too:
// a sprite never overwrites a pixel a higher-priority sprite already claimed.
func (e *engine) sprites(dispcnt uint32, y int) {
	for i := 127; i >= 0; i-- {
		o := e.oamBase + uint32(i)*8
		a0 := e.oam16(o)
		a1 := e.oam16(o + 2)
		a2 := e.oam16(o + 4)

		mode := (a0 >> 8) & 3
		if mode == 2 {
			continue // "disabled" in the non-affine encoding
		}
		affine := mode == 1 || mode == 3
		double := mode == 3

		shape := int(a0 >> 14 & 3)
		if shape == 3 {
			e.m.note("2D engine %s: OAM entry %d uses the reserved sprite shape 3", e.name(), i)
			continue
		}
		size := int(a1 >> 14 & 3)
		w, h := objSize[shape][size][0], objSize[shape][size][1]

		// The drawn box is twice the sprite when the double-size bit is set, which is
		// how a rotated sprite is given room for its corners.
		bw, bh := w, h
		if double {
			bw, bh = w*2, h*2
		}

		// Y is 8 bits and wraps: a sprite at Y=250 hangs down into the top of the screen.
		sy := int(a0 & 0xFF)
		dy := (y - sy) & 0xFF
		if dy >= bh {
			continue
		}

		// X is 9-bit signed, so the left half of the range is off-screen to the left.
		sx := int(a1 & 0x1FF)
		if sx >= 256 {
			sx -= 512
		}

		if a0&(1<<12) != 0 {
			e.m.note("2D engine %s: sprite mosaic (OAM attr0 bit 12) not implemented", e.name())
		}

		gfx := (a0 >> 10) & 3
		bpp8 := a0&(1<<13) != 0
		prio := uint8(a2 >> 10 & 3)
		pal := int(a2 >> 12 & 0xF)

		// A bitmap sprite's "palette" field is its alpha, and an alpha of zero does not
		// make it faint — it makes it invisible.
		alpha := pal
		if gfx == 3 && alpha == 0 {
			continue
		}

		var pa, pb, pc, pd int32
		if affine {
			pa, pb, pc, pd = e.oamAffine(int(a1 >> 9 & 0x1F))
		}
		hflip := !affine && a1&(1<<12) != 0
		vflip := !affine && a1&(1<<13) != 0

		for bx := 0; bx < bw; bx++ {
			x := sx + bx
			if x < 0 || x >= screenW {
				continue
			}

			// Map the pixel inside the drawn box to a pixel inside the sprite.
			var px, py int
			if affine {
				rx := int32(bx - bw/2)
				ry := int32(dy - bh/2)
				px = int((pa*rx+pb*ry)>>8) + w/2
				py = int((pc*rx+pd*ry)>>8) + h/2
				if px < 0 || px >= w || py < 0 || py >= h {
					continue
				}
			} else {
				px, py = bx, dy
				if hflip {
					px = w - 1 - px
				}
				if vflip {
					py = h - 1 - py
				}
			}

			if e.objOK[x] && e.objPrio[x] < prio {
				continue // a higher-priority sprite already owns this pixel
			}

			var c uint16
			var ok bool
			var eva uint8
			if gfx == 3 {
				c, ok = e.objBitmapPixel(dispcnt, uint32(a2&0x3FF), w, px, py)
				eva = uint8(alpha) + 1
			} else {
				c, ok = e.objTilePixel(dispcnt, uint32(a2&0x3FF), bpp8, pal, w, px, py)
			}
			if !ok {
				continue
			}

			// Gfx mode 2 sprites are not drawn at all: they paint the OBJ window's shape
			// and nothing else. Their colour is never seen — only their silhouette.
			if gfx == 2 {
				e.objWin[x] = true
				continue
			}
			if dispcnt&(1<<12) == 0 {
				continue // OBJ layer disabled: only the window shape above was wanted
			}

			e.obj[x] = c
			e.objOK[x] = true
			e.objPrio[x] = prio
			e.objSemi[x] = gfx == 1 || gfx == 3
			e.objEVA[x] = eva
		}
	}
}

func (e *engine) oam16(off uint32) uint16 { return palAt(e.m.oam, off) }

// oamAffine reads one of the 32 affine parameter groups. They are not a separate
// table: PA..PD are stashed in the *fourth* halfword of four consecutive sprite
// entries, so group i lives at OAM offsets i*32 + 6, +14, +22, +30 — interleaved
// with sprites that know nothing about them.
func (e *engine) oamAffine(i int) (pa, pb, pc, pd int32) {
	b := e.oamBase + uint32(i)*32
	pa = int32(int16(e.oam16(b + 6)))
	pb = int32(int16(e.oam16(b + 14)))
	pc = int32(int16(e.oam16(b + 22)))
	pd = int32(int16(e.oam16(b + 30)))
	return
}

// objTilePixel fetches one texel of a tiled sprite.
//
// The addressing is the classic DS trap. In 1D mapping the tile number is scaled by
// a boundary that DISPCNT bits 20-21 choose (32, 64, 128 or 256 bytes) — so the same
// OAM entry names a different tile depending on a register the sprite cannot see. In
// 2D mapping the boundary is always 32 and OBJ VRAM is a 32-tile-wide grid, so the
// next row of a sprite is 1024 bytes further on regardless of colour depth, while
// the next tile along is 32 bytes (4bpp) or 64 (8bpp).
func (e *engine) objTilePixel(dispcnt uint32, tile uint32, bpp8 bool, pal, w, px, py int) (uint16, bool) {
	oneD := dispcnt&(1<<4) != 0
	tw := w / 8

	var addr uint32
	if oneD {
		bound := uint32(32) << (dispcnt >> 20 & 3)
		sz := uint32(32)
		if bpp8 {
			sz = 64
		}
		addr = tile*bound + (uint32(py/8)*uint32(tw)+uint32(px/8))*sz
	} else {
		addr = tile * 32
		addr += uint32(py/8) * 1024
		if bpp8 {
			addr += uint32(px/8) * 64
		} else {
			addr += uint32(px/8) * 32
		}
	}

	fx, fy := uint32(px%8), uint32(py%8)
	var idx int
	if bpp8 {
		idx = int(e.m.vram.read8(e.spOBJ, addr+fy*8+fx))
	} else {
		b := e.m.vram.read8(e.spOBJ, addr+fy*4+fx/2)
		if fx&1 != 0 {
			idx = int(b >> 4)
		} else {
			idx = int(b & 0xF)
		}
	}
	if idx == 0 {
		return 0, false
	}
	switch {
	case bpp8 && dispcnt&(1<<31) != 0:
		return e.objExtColor(pal, idx), true
	case bpp8:
		return e.palOBJColor(idx), true
	default:
		return e.palOBJColor(pal*16 + idx), true
	}
}

// objBitmapPixel fetches one texel of a direct-colour sprite. Bitmap sprites have
// their own mapping bits, separate from the tiled ones: DISPCNT bit 6 picks 1D or
// 2D, bit 22 the 1D boundary (128 or 256 bytes), and bit 5 the width of the 2D
// bitmap area (128 or 256 pixels), which is what turns the tile number into a
// two-dimensional position.
func (e *engine) objBitmapPixel(dispcnt uint32, tile uint32, w, px, py int) (uint16, bool) {
	var base, stride uint32
	if dispcnt&(1<<6) != 0 { // 1D
		bound := uint32(128)
		if dispcnt&(1<<22) != 0 {
			bound = 256
		}
		base = tile * bound
		stride = uint32(w) * 2
	} else if dispcnt&(1<<5) == 0 { // 2D, 128 pixels wide
		base = (tile&0x1F)*0x10 + (tile&0x3E0)*0x80
		stride = 256
	} else { // 2D, 256 pixels wide
		base = (tile&0x0F)*0x10 + (tile&0x3F0)*0x80
		stride = 512
	}
	c := e.m.vram.read16(e.spOBJ, base+uint32(py)*stride+uint32(px)*2)
	if c&0x8000 == 0 {
		return 0, false
	}
	return c & 0x7FFF, true
}

// --- windows ----------------------------------------------------------------

// windowMask computes, for every pixel of a scanline, which layers may be seen and
// whether the colour effects apply there. The four regions are checked in the
// hardware's order of precedence: window 0 beats window 1 beats the OBJ window, and
// what is in none of them takes WINOUT's "outside" half.
func (e *engine) windowMask(dispcnt uint32, y int) {
	if dispcnt&0xE000 == 0 { // no window is enabled: everything is visible everywhere
		for x := 0; x < screenW; x++ {
			e.win[x] = 0x3F
		}
		return
	}
	winin := e.reg16(rWININ)
	winout := e.reg16(rWINOUT)
	out := uint8(winout & 0x3F)
	objw := uint8(winout >> 8 & 0x3F)
	w0 := uint8(winin & 0x3F)
	w1 := uint8(winin >> 8 & 0x3F)

	in0 := dispcnt&(1<<13) != 0 && e.inWindowY(0, y)
	in1 := dispcnt&(1<<14) != 0 && e.inWindowY(1, y)
	useObj := dispcnt&(1<<15) != 0

	for x := 0; x < screenW; x++ {
		m := out
		if useObj && e.objWin[x] {
			m = objw
		}
		if in1 && e.inWindowX(1, x) {
			m = w1
		}
		if in0 && e.inWindowX(0, x) {
			m = w0
		}
		e.win[x] = m
	}
}

// inWindowY and inWindowX test a window's bounds. Both edges are packed into one
// register with the *right/bottom* edge in the low byte and the left/top in the high
// one, and the right/bottom edge is exclusive. A window whose end is not greater than
// its start does not vanish: the hardware runs it to the edge of the screen, which is
// how a game opens a window by writing only its left edge.
func (e *engine) inWindowY(n, y int) bool {
	v := e.reg16(rWIN0V + uint32(n)*2)
	y1, y2 := int(v>>8), int(v&0xFF)
	if y2 <= y1 || y2 > screenH {
		y2 = screenH
	}
	return y >= y1 && y < y2
}

func (e *engine) inWindowX(n, x int) bool {
	v := e.reg16(rWIN0H + uint32(n)*2)
	x1, x2 := int(v>>8), int(v&0xFF)
	if x2 <= x1 {
		x2 = screenW
	}
	return x >= x1 && x < x2
}

// --- priority and colour effects --------------------------------------------

// cand is one candidate pixel at one screen position: what colour it is, which layer
// it came from, and the two things only a sprite or the 3D layer carries — that it
// blends with whatever is underneath, and how strongly.
type cand struct {
	c     uint16
	layer int
	semi  bool  // blends with the layer below regardless of BLDCNT's first-target bits
	eva   uint8 // a bitmap sprite's own coefficient (0 = take EVA from BLDALPHA)
	a3d   uint8 // the 3D layer's alpha, 0..31 (31 = opaque)
}

// composite resolves priority and applies the colour effects for one scanline.
func (e *engine) composite(y int) {
	bldcnt := e.reg16(rBLDCNT)
	bldalpha := e.reg16(rBLDALPHA)
	bldy := e.reg16(rBLDY)
	effect := int(bldcnt >> 6 & 3)
	first := uint8(bldcnt & 0x3F)
	second := uint8(bldcnt >> 8 & 0x3F)
	eva := clamp16(int(bldalpha & 31))
	evb := clamp16(int(bldalpha >> 8 & 31))
	evy := clamp16(int(bldy & 31))
	backdrop := e.palBGColor(0)

	var cs [6]cand
	for x := 0; x < screenW; x++ {
		mask := e.win[x]
		n := 0

		// Lower priority number wins, and a sprite beats a background of the *same*
		// priority — which is why the sprite is offered first at each level. Among
		// backgrounds of equal priority the lower BG number wins.
		for p := uint8(0); p < 4; p++ {
			if e.objOK[x] && e.objPrio[x] == p && mask&(1<<lyOBJ) != 0 {
				cs[n] = cand{c: e.obj[x], layer: lyOBJ, semi: e.objSemi[x], eva: e.objEVA[x]}
				n++
			}
			for b := 0; b < 4; b++ {
				if !e.shown[b] || !e.bgOK[b][x] || e.prio[b] != p || mask&(1<<uint(b)) == 0 {
					continue
				}
				cs[n] = cand{c: e.bg[b][x], layer: b}
				if b == 0 && e.is3D {
					cs[n].a3d = e.a3D[x]
				}
				n++
			}
		}
		cs[n] = cand{c: backdrop, layer: lyBackdrop, a3d: 31}
		n++

		top := cs[0]
		below := cs[1] // always valid: the backdrop is never absent
		if n == 1 {
			below = cand{c: backdrop, layer: lyBackdrop}
		}
		fx := mask&(1<<5) != 0

		switch {
		case top.layer == lyBG0 && e.is3D && top.a3d < 31:
			// The 3D layer carries its own alpha and blends with the layer below on its
			// own account — no BLDCNT bit is needed to make it happen. The coefficients
			// are out of 32, not 16, because the 3D alpha is five bits.
			e.line[x] = blend32(top.c, below.c, int(top.a3d)+1, 31-int(top.a3d))

		case top.semi && fx && second&(1<<uint(below.layer)) != 0:
			// A semi-transparent sprite blends whatever BLDCNT's first-target bits say —
			// but the layer below it must still be a legal second target, and a bitmap
			// sprite brings its own EVA.
			a, b := eva, evb
			if top.eva != 0 {
				a = clamp16(int(top.eva))
				b = 16 - a
			}
			e.line[x] = blend16(top.c, below.c, a, b)

		case fx && first&(1<<uint(top.layer)) != 0:
			switch effect {
			case 1:
				if second&(1<<uint(below.layer)) != 0 {
					e.line[x] = blend16(top.c, below.c, eva, evb)
				} else {
					e.line[x] = top.c
				}
			case 2:
				e.line[x] = brighten(top.c, evy)
			case 3:
				e.line[x] = darken(top.c, evy)
			default:
				e.line[x] = top.c
			}

		default:
			e.line[x] = top.c
		}
	}
}

func clamp16(v int) int {
	if v > 16 {
		return 16
	}
	return v
}

// blend16 is the 2D engine's alpha blend: each component is (a*EVA + b*EVB)/16,
// saturating at 31. EVA and EVB are independent, so they do not have to sum to 16 —
// a game can (and does) use them to over-brighten a highlight.
func blend16(a, b uint16, eva, evb int) uint16 {
	return blendN(a, b, eva, evb, 16)
}

// blend32 is the same blend with coefficients out of 32, which is what the 3D
// layer's five-bit alpha needs.
func blend32(a, b uint16, ca, cb int) uint16 {
	return blendN(a, b, ca, cb, 32)
}

func blendN(a, b uint16, ca, cb, div int) uint16 {
	var out [3]uint32
	for i := uint(0); i < 3; i++ {
		v := (chan5(a, i)*uint32(ca) + chan5(b, i)*uint32(cb)) / uint32(div)
		if v > 31 {
			v = 31
		}
		out[i] = v
	}
	return rgb555(out[0], out[1], out[2])
}

func brighten(c uint16, evy int) uint16 {
	var out [3]uint32
	for i := uint(0); i < 3; i++ {
		v := chan5(c, i)
		out[i] = v + (31-v)*uint32(evy)/16
	}
	return rgb555(out[0], out[1], out[2])
}

func darken(c uint16, evy int) uint16 {
	var out [3]uint32
	for i := uint(0); i < 3; i++ {
		v := chan5(c, i)
		out[i] = v - v*uint32(evy)/16
	}
	return rgb555(out[0], out[1], out[2])
}
