package gbamachine

// The PPU: a per-scanline software model of the AGB display controller — the
// direct ancestor of the DS 2D engines in dsmachine/gpu2d.go, with one screen and
// no VRAM banking. Each visible line is composed when the raster reaches it:
// the enabled backgrounds and the sprite line are rendered to buffers, then the
// per-pixel priority/window/blending rules pick what the wire carries.
//
// Modes: 0 (4 text BGs), 1 (BG0/1 text + BG2 affine), 2 (BG2/3 affine),
// 3 (240x160 direct-colour bitmap), 4 (paletted bitmap, paged), 5 (160x128
// direct-colour, paged). Sprites: regular and affine, 4/8bpp, 1D/2D mapping,
// semi-transparent and OBJ-window modes.
//
// Not yet modelled (logged when a game first leans on them): mosaic.

type ppu struct {
	// Internal affine accumulators (28-bit sign-extended, 8 fractional bits).
	// Loaded from BG2X/Y and BG3X/Y at the top of each frame and whenever the
	// game writes them; stepped by PB/PD after each rendered line.
	bg2x, bg2y int32
	bg3x, bg3y int32
}

// objVRAMBase reports where OBJ character data starts in VRAM for the current
// mode — the byte-store boundary the bus needs (bitmap modes push it up).
func (p *ppu) objVRAMBase(m *Machine) uint32 {
	if m.io[0x000]&7 >= 3 {
		return 0x14000
	}
	return 0x10000
}

// reloadAffineRef reloads an internal accumulator after the game writes a
// reference-point register (a write takes effect immediately, mid-frame).
func (p *ppu) reloadAffineRef(m *Machine, reg uint32) {
	ref := func(lo uint32) int32 {
		v := uint32(m.io[lo]) | uint32(m.io[lo+2])<<16
		return int32(v<<4) >> 4 // 28-bit, sign-extended
	}
	switch reg &^ 2 {
	case 0x028:
		p.bg2x = ref(0x028)
	case 0x02C:
		p.bg2y = ref(0x02C)
	case 0x038:
		p.bg3x = ref(0x038)
	case 0x03C:
		p.bg3y = ref(0x03C)
	}
}

func (p *ppu) startFrame(m *Machine) {
	p.reloadAffineRef(m, 0x028)
	p.reloadAffineRef(m, 0x02C)
	p.reloadAffineRef(m, 0x038)
	p.reloadAffineRef(m, 0x03C)
}

// stepAffine advances the accumulators by one scanline (PB/PD).
func (p *ppu) stepAffine(m *Machine) {
	p.bg2x += int32(int16(m.io[0x022]))
	p.bg2y += int32(int16(m.io[0x026]))
	p.bg3x += int32(int16(m.io[0x032]))
	p.bg3y += int32(int16(m.io[0x036]))
}

// rgb15 expands a 15-bit BGR colour to 0xFFRRGGBB.
func rgb15(c uint16) uint32 {
	r := uint32(c) & 31
	g := uint32(c>>5) & 31
	b := uint32(c>>10) & 31
	exp := func(v uint32) uint32 { return v<<3 | v>>2 }
	return 0xFF000000 | exp(r)<<16 | exp(g)<<8 | exp(b)
}

func (m *Machine) pal16(bank, idx int) uint16 {
	o := bank*32 + idx*2
	return uint16(m.pal[o]) | uint16(m.pal[o+1])<<8
}

// lineBuf is one layer's scanline: 15-bit colours plus an opacity mask.
type lineBuf struct {
	c  [screenW]uint16
	on [screenW]bool
}

// renderLine composes visible line y into m.screen.
func (p *ppu) renderLine(m *Machine, y int) {
	dispcnt := m.io[0x000]
	out := m.screen[y*screenW : (y+1)*screenW]
	if dispcnt&(1<<7) != 0 { // forced blank: the wire carries white
		for i := range out {
			out[i] = 0xFFFFFFFF
		}
		return
	}
	mode := int(dispcnt & 7)

	// --- render each enabled background's line ---
	var bg [4]lineBuf
	bgEnabled := func(n int) bool { return dispcnt&(1<<(8+n)) != 0 }
	switch mode {
	case 0:
		for n := 0; n < 4; n++ {
			if bgEnabled(n) {
				p.textLine(m, n, y, &bg[n])
			}
		}
	case 1:
		if bgEnabled(0) {
			p.textLine(m, 0, y, &bg[0])
		}
		if bgEnabled(1) {
			p.textLine(m, 1, y, &bg[1])
		}
		if bgEnabled(2) {
			p.affineLine(m, 2, &bg[2])
		}
	case 2:
		if bgEnabled(2) {
			p.affineLine(m, 2, &bg[2])
		}
		if bgEnabled(3) {
			p.affineLine(m, 3, &bg[3])
		}
	case 3, 4, 5:
		if bgEnabled(2) {
			p.bitmapLine(m, mode, &bg[2])
		}
	default:
		m.note("display mode %d selected (undefined on hardware)", mode)
	}

	// --- render the sprite line ---
	var (
		objC    [screenW]uint16
		objOn   [screenW]bool
		objPrio [screenW]uint8
		objSemi [screenW]bool
		objWin  [screenW]bool
	)
	if dispcnt&(1<<12) != 0 {
		p.objLine(m, y, dispcnt, objC[:], objOn[:], objPrio[:], objSemi[:], objWin[:])
	}

	// --- per-pixel window control ---
	winEnabled := dispcnt&(7<<13) != 0
	winIn, winOut := m.io[0x048], m.io[0x04A]
	winCtl := func(x int) uint16 {
		if !winEnabled {
			return 0x3F // everything on, effects allowed
		}
		if dispcnt&(1<<13) != 0 && p.inWindow(m, 0x040, 0x044, x, y) {
			return winIn & 0x3F
		}
		if dispcnt&(1<<14) != 0 && p.inWindow(m, 0x042, 0x046, x, y) {
			return winIn >> 8 & 0x3F
		}
		if dispcnt&(1<<15) != 0 && objWin[x] {
			return winOut >> 8 & 0x3F
		}
		return winOut & 0x3F
	}

	// --- compose ---
	bldcnt := m.io[0x050]
	bldMode := bldcnt >> 6 & 3
	eva := int(m.io[0x052]) & 0x1F
	evb := int(m.io[0x052]>>8) & 0x1F
	evy := int(m.io[0x054]) & 0x1F
	if eva > 16 {
		eva = 16
	}
	if evb > 16 {
		evb = 16
	}
	if evy > 16 {
		evy = 16
	}
	backdrop := m.pal16(0, 0)

	for x := 0; x < screenW; x++ {
		ctl := winCtl(x)
		// Find the top two visible pixels: layer ids 0-3 = BG, 4 = OBJ, 5 = backdrop.
		top, second := 5, 5
		topC, secondC := backdrop, backdrop
		topSemi := false
		found := 0
		for prio := 0; prio < 4 && found < 2; prio++ {
			if objOn[x] && int(objPrio[x]) == prio && ctl&(1<<4) != 0 {
				if found == 0 {
					top, topC, topSemi = 4, objC[x], objSemi[x]
				} else {
					second, secondC = 4, objC[x]
				}
				found++
				if found == 2 {
					break
				}
			}
			for n := 0; n < 4 && found < 2; n++ {
				if bg[n].on[x] && int(m.io[0x008+2*uint32(n)]&3) == prio && ctl&(1<<uint(n)) != 0 {
					if found == 0 {
						top, topC = n, bg[n].c[x]
					} else {
						second, secondC = n, bg[n].c[x]
					}
					found++
				}
			}
		}

		c := topC
		effects := ctl&(1<<5) != 0
		firstMask := func(layer int) bool { return bldcnt&(1<<uint(layer)) != 0 }
		secondMask := func(layer int) bool { return bldcnt&(1<<uint(8+layer)) != 0 }
		switch {
		case topSemi && effects && secondMask(second):
			// A semi-transparent sprite forces alpha blending with whatever shows
			// through, regardless of the BLDCNT mode.
			c = blend(topC, secondC, eva, evb)
		case bldMode == 1 && effects && firstMask(top) && secondMask(second):
			c = blend(topC, secondC, eva, evb)
		case bldMode == 2 && effects && firstMask(top):
			c = brighten(topC, evy)
		case bldMode == 3 && effects && firstMask(top):
			c = darken(topC, evy)
		}
		out[x] = rgb15(c)
	}
}

// inWindow tests x,y against a WINxH/WINxV pair, honouring the wrap rule
// (x1 > x2 means the window wraps around the screen edge).
func (p *ppu) inWindow(m *Machine, hreg, vreg uint32, x, y int) bool {
	h, v := m.io[hreg], m.io[vreg]
	x1, x2 := int(h>>8), int(h&0xFF)
	y1, y2 := int(v>>8), int(v&0xFF)
	inX := x >= x1 && x < x2
	if x1 > x2 {
		inX = x >= x1 || x < x2
	}
	inY := y >= y1 && y < y2
	if y1 > y2 {
		inY = y >= y1 || y < y2
	}
	return inX && inY
}

func blend(a, b uint16, eva, evb int) uint16 {
	mix := func(ca, cb uint16, sh uint) uint16 {
		v := (int(ca>>sh&31)*eva + int(cb>>sh&31)*evb) >> 4
		if v > 31 {
			v = 31
		}
		return uint16(v) << sh
	}
	return mix(a, b, 0) | mix(a, b, 5) | mix(a, b, 10)
}

func brighten(c uint16, evy int) uint16 {
	f := func(sh uint) uint16 {
		v := int(c >> sh & 31)
		v += (31 - v) * evy >> 4
		return uint16(v) << sh
	}
	return f(0) | f(5) | f(10)
}

func darken(c uint16, evy int) uint16 {
	f := func(sh uint) uint16 {
		v := int(c >> sh & 31)
		v -= v * evy >> 4
		return uint16(v) << sh
	}
	return f(0) | f(5) | f(10)
}

// --- text backgrounds --------------------------------------------------------

func (p *ppu) textLine(m *Machine, n, y int, out *lineBuf) {
	cnt := m.io[0x008+2*uint32(n)]
	if cnt&(1<<6) != 0 {
		m.note("BG%d uses mosaic (not modelled)", n)
	}
	charBase := uint32(cnt>>2&3) * 0x4000
	scrBase := uint32(cnt>>8&31) * 0x800
	eightBpp := cnt&(1<<7) != 0
	size := cnt >> 14 & 3
	w, h := 256, 256
	if size == 1 || size == 3 {
		w = 512
	}
	if size == 2 || size == 3 {
		h = 512
	}
	hofs := int(m.io[0x010+4*uint32(n)] & 0x1FF)
	vofs := int(m.io[0x012+4*uint32(n)] & 0x1FF)

	sy := (y + vofs) & (h - 1)
	for x := 0; x < screenW; x++ {
		sx := (x + hofs) & (w - 1)
		// Which 256x256 screenblock quadrant, then the entry inside it.
		quad := 0
		if w == 512 && sx >= 256 {
			quad++
		}
		if h == 512 && sy >= 256 {
			quad += w / 256
		}
		entryOff := scrBase + uint32(quad)*0x800 + uint32((sy&255)>>3)*64 + uint32((sx&255)>>3)*2
		entry := uint16(m.vram[entryOff]) | uint16(m.vram[entryOff+1])<<8
		tile := uint32(entry & 0x3FF)
		tx, ty := sx&7, sy&7
		if entry&(1<<10) != 0 {
			tx = 7 - tx
		}
		if entry&(1<<11) != 0 {
			ty = 7 - ty
		}
		var idx, bank int
		if eightBpp {
			a := charBase + tile*64 + uint32(ty)*8 + uint32(tx)
			if a >= 0x10000 {
				continue // char data cannot cross into OBJ VRAM
			}
			idx, bank = int(m.vram[a]), 0
		} else {
			a := charBase + tile*32 + uint32(ty)*4 + uint32(tx>>1)
			if a >= 0x10000 {
				continue
			}
			b := m.vram[a]
			if tx&1 == 1 {
				idx = int(b >> 4)
			} else {
				idx = int(b & 0xF)
			}
			bank = int(entry >> 12)
		}
		if idx != 0 {
			out.c[x] = m.pal16(bank, idx)
			out.on[x] = true
		}
	}
}

// --- affine backgrounds ------------------------------------------------------

func (p *ppu) affineLine(m *Machine, n int, out *lineBuf) {
	cnt := m.io[0x008+2*uint32(n)]
	charBase := uint32(cnt>>2&3) * 0x4000
	scrBase := uint32(cnt>>8&31) * 0x800
	wrap := cnt&(1<<13) != 0
	sizes := [4]int{128, 256, 512, 1024}
	size := sizes[cnt>>14&3]

	var cx, cy int32
	var pa, pc int32
	if n == 2 {
		cx, cy = p.bg2x, p.bg2y
		pa, pc = int32(int16(m.io[0x020])), int32(int16(m.io[0x024]))
	} else {
		cx, cy = p.bg3x, p.bg3y
		pa, pc = int32(int16(m.io[0x030])), int32(int16(m.io[0x034]))
	}

	for x := 0; x < screenW; x++ {
		tx, ty := int(cx>>8), int(cy>>8)
		cx += pa
		cy += pc
		if wrap {
			tx &= size - 1
			ty &= size - 1
		} else if tx < 0 || ty < 0 || tx >= size || ty >= size {
			continue
		}
		tilesPerRow := size / 8
		mapOff := scrBase + uint32(ty>>3)*uint32(tilesPerRow) + uint32(tx>>3)
		tile := uint32(m.vram[mapOff])
		a := charBase + tile*64 + uint32(ty&7)*8 + uint32(tx&7)
		if a >= 0x10000 {
			continue
		}
		if idx := int(m.vram[a]); idx != 0 {
			out.c[x] = m.pal16(0, idx)
			out.on[x] = true
		}
	}
}

// --- bitmap modes ------------------------------------------------------------

func (p *ppu) bitmapLine(m *Machine, mode int, out *lineBuf) {
	// Bitmap modes are still BG2 and still go through the affine transform (a
	// game that leaves PA..PD at reset identity gets the plain framebuffer).
	pa, pc := int32(int16(m.io[0x020])), int32(int16(m.io[0x024]))
	cx, cy := p.bg2x, p.bg2y
	page := uint32(0)
	if m.io[0x000]&(1<<4) != 0 && mode != 3 {
		page = 0xA000
	}
	w, h := screenW, screenH
	if mode == 5 {
		w, h = 160, 128
	}
	for x := 0; x < screenW; x++ {
		tx, ty := int(cx>>8), int(cy>>8)
		cx += pa
		cy += pc
		if tx < 0 || ty < 0 || tx >= w || ty >= h {
			continue
		}
		switch mode {
		case 3:
			o := (uint32(ty)*uint32(w) + uint32(tx)) * 2
			out.c[x] = uint16(m.vram[o]) | uint16(m.vram[o+1])<<8
			out.on[x] = true
		case 4:
			if idx := int(m.vram[page+uint32(ty)*uint32(w)+uint32(tx)]); idx != 0 {
				out.c[x] = m.pal16(0, idx)
				out.on[x] = true
			}
		case 5:
			o := page + (uint32(ty)*uint32(w)+uint32(tx))*2
			out.c[x] = uint16(m.vram[o]) | uint16(m.vram[o+1])<<8
			out.on[x] = true
		}
	}
}

// --- sprites -----------------------------------------------------------------

// objSizes[shape][size] = {width, height} in pixels.
var objSizes = [3][4][2]int{
	{{8, 8}, {16, 16}, {32, 32}, {64, 64}},   // square
	{{16, 8}, {32, 8}, {32, 16}, {64, 32}},   // horizontal
	{{8, 16}, {8, 32}, {16, 32}, {32, 64}},   // vertical
}

func (p *ppu) objLine(m *Machine, y int, dispcnt uint16,
	c []uint16, on []bool, prio []uint8, semi []bool, win []bool) {
	oneDim := dispcnt&(1<<6) != 0
	minTile := uint32(0)
	if dispcnt&7 >= 3 {
		minTile = 512 // in bitmap modes the low OBJ tiles are the framebuffer
	}
	for i := 0; i < 128; i++ {
		o := i * 8
		attr0 := uint16(m.oam[o]) | uint16(m.oam[o+1])<<8
		attr1 := uint16(m.oam[o+2]) | uint16(m.oam[o+3])<<8
		attr2 := uint16(m.oam[o+4]) | uint16(m.oam[o+5])<<8
		affine := attr0&(1<<8) != 0
		if !affine && attr0&(1<<9) != 0 {
			continue // disabled
		}
		mode := attr0 >> 10 & 3
		if mode == 3 {
			continue // prohibited
		}
		if attr0&(1<<12) != 0 {
			m.note("sprite uses mosaic (not modelled)")
		}
		shape := int(attr0 >> 14)
		if shape == 3 {
			continue
		}
		w := objSizes[shape][attr1>>14][0]
		h := objSizes[shape][attr1>>14][1]
		bw, bh := w, h
		if affine && attr0&(1<<9) != 0 {
			bw, bh = 2*w, 2*h // double-size rendering area
		}
		oy := int(attr0 & 0xFF)
		if oy+bh > 256 {
			oy -= 256
		}
		ox := int(attr1 & 0x1FF)
		if ox >= 256 {
			ox -= 512
		}
		row := y - oy
		if row < 0 || row >= bh {
			continue
		}

		eightBpp := attr0&(1<<13) != 0
		tile := uint32(attr2 & 0x3FF)
		if tile < minTile {
			continue
		}
		pr := uint8(attr2 >> 10 & 3)
		bank := int(attr2 >> 12)

		var paf, pbf, pcf, pdf int32 = 0x100, 0, 0, 0x100
		if affine {
			g := int(attr1 >> 9 & 0x1F)
			rd := func(k int) int32 {
				a := g*32 + k*8 + 6
				return int32(int16(uint16(m.oam[a]) | uint16(m.oam[a+1])<<8))
			}
			paf, pbf, pcf, pdf = rd(0), rd(1), rd(2), rd(3)
		}
		hflip := !affine && attr1&(1<<12) != 0
		vflip := !affine && attr1&(1<<13) != 0

		for bx := 0; bx < bw; bx++ {
			x := ox + bx
			if x < 0 || x >= screenW {
				continue
			}
			var sx, sy int
			if affine {
				// Rotate/scale about the box centre.
				dx, dy := int32(bx-bw/2), int32(row-bh/2)
				fx := paf*dx + pbf*dy + int32(w/2)<<8
				fy := pcf*dx + pdf*dy + int32(h/2)<<8
				sx, sy = int(fx>>8), int(fy>>8)
				if sx < 0 || sy < 0 || sx >= w || sy >= h {
					continue
				}
			} else {
				sx, sy = bx, row
				if hflip {
					sx = w - 1 - sx
				}
				if vflip {
					sy = h - 1 - sy
				}
			}

			// Locate the texel: tiles are 8x8; 1D mapping packs a sprite's tiles
			// consecutively, 2D mapping lays them in a 32-tile-wide char grid.
			tileX, tileY := uint32(sx>>3), uint32(sy>>3)
			var t uint32
			if eightBpp {
				if oneDim {
					t = tile + tileY*uint32(w>>3)*2 + tileX*2
				} else {
					t = (tile &^ 1) + tileY*32 + tileX*2
				}
				a := 0x10000 + t*32 + uint32(sy&7)*8 + uint32(sx&7)
				if a >= vramSize {
					continue
				}
				p.objPixel(m, x, int(m.vram[a]), 0x10, pr, mode, c, on, prio, semi, win)
			} else {
				if oneDim {
					t = tile + tileY*uint32(w>>3) + tileX
				} else {
					t = tile + tileY*32 + tileX
				}
				a := 0x10000 + t*32 + uint32(sy&7)*4 + uint32(sx>>1&3)
				if a >= vramSize {
					continue
				}
				v := m.vram[a]
				idx := int(v & 0xF)
				if sx&1 == 1 {
					idx = int(v >> 4)
				}
				p.objPixel(m, x, idx, 0x10+bank, pr, mode, c, on, prio, semi, win)
			}
		}
	}
}

// objPixel commits one sprite texel to the line buffers, honouring the
// lowest-OAM-index-wins rule (the buffers are first-write-wins).
func (p *ppu) objPixel(m *Machine, x, idx, palBank int, pr uint8, mode uint16,
	c []uint16, on []bool, prio []uint8, semi []bool, win []bool) {
	if idx == 0 {
		return
	}
	if mode == 2 { // OBJ-window: contributes shape, not colour
		win[x] = true
		return
	}
	if on[x] {
		// An earlier (lower-index) sprite owns this pixel — but a LOWER priority
		// number from a later sprite still cannot steal it; hardware keeps the
		// first opaque texel per pixel within the sprite layer.
		return
	}
	on[x] = true
	c[x] = m.pal16(palBank, idx)
	prio[x] = pr
	semi[x] = mode == 1
}
