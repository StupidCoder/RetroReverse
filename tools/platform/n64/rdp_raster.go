package n64

// rdp_raster.go turns the RDP's state into pixels in RDRAM.
//
// Three of the four cycle types matter to a frame. FILL writes a constant and is
// how every framebuffer clear happens. COPY moves texels to the screen without
// touching the colour combiner, and is how 2-D blits happen. 1CYCLE and 2CYCLE
// run the combiner and the blender, and are how everything else happens.
//
// Coordinates arrive in 10.2 fixed point — quarter-pixels — because the RDP
// rasterises on a subpixel grid. Everything here rounds them to whole pixels at
// the last moment, and clips against the scissor rectangle, which is the only
// clipping the hardware does.

// pixelAddr returns the RDRAM byte offset of pixel (x, y) in the colour image.
func (r *rdp) pixelAddr(x, y uint32) uint32 {
	switch r.Color.Size {
	case size16:
		return r.Color.Addr + (y*r.Color.Width+x)*2
	case size32:
		return r.Color.Addr + (y*r.Color.Width+x)*4
	case size8:
		return r.Color.Addr + y*r.Color.Width + x
	}
	return r.Color.Addr
}

// rgba5551 packs an 8-bit-per-channel colour into the framebuffer's 16-bit form.
func rgba5551(rr, gg, bb, aa uint32) uint16 {
	return uint16(rr>>3)<<11 | uint16(gg>>3)<<6 | uint16(bb>>3)<<1 | uint16(aa>>7&1)
}

// writePixel stores a colour, honouring the colour image's pixel size.
func (m *Machine) writePixel(x, y uint32, rr, gg, bb, aa uint32) {
	r := &m.rdp
	a := m.rdp.pixelAddr(x, y)
	switch r.Color.Size {
	case size16:
		m.storeRDRAM16(a, rgba5551(rr, gg, bb, aa))
	case size32:
		m.storeRDRAM32(a, rr<<24|gg<<16|bb<<8|aa)
	case size8:
		m.storeRDRAM8(a, byte(rr))
	default:
		m.CPU.Halt("unmodelled colour image size %d", r.Color.Size)
	}
}

func (m *Machine) storeRDRAM8(a uint32, v byte) {
	if int(a) < len(m.RDRAM) {
		m.RDRAM[a] = v
	}
}

func (m *Machine) storeRDRAM16(a uint32, v uint16) {
	if int(a)+1 < len(m.RDRAM) {
		m.RDRAM[a] = byte(v >> 8)
		m.RDRAM[a+1] = byte(v)
	}
}

func (m *Machine) storeRDRAM32(a uint32, v uint32) {
	if int(a)+3 < len(m.RDRAM) {
		m.RDRAM[a] = byte(v >> 24)
		m.RDRAM[a+1] = byte(v >> 16)
		m.RDRAM[a+2] = byte(v >> 8)
		m.RDRAM[a+3] = byte(v)
	}
}

// clip narrows a rectangle in whole pixels to the scissor.
func (r *rdp) clip(xh, yh, xl, yl uint32) (uint32, uint32, uint32, uint32, bool) {
	sxh, syh := r.Scissor.XH>>2, r.Scissor.YH>>2
	sxl, syl := r.Scissor.XL>>2, r.Scissor.YL>>2
	if xh < sxh {
		xh = sxh
	}
	if yh < syh {
		yh = syh
	}
	if xl > sxl {
		xl = sxl
	}
	if yl > syl {
		yl = syl
	}
	// The colour image is the hard bound: a scissor wider than it would walk off
	// the end of RDRAM.
	if xl > r.Color.Width {
		xl = r.Color.Width
	}
	return xh, yh, xl, yl, xh < xl && yh < yl
}

// fillRect draws a rectangle of the fill colour. In FILL mode the colour is
// written verbatim, and for a 16-bit framebuffer it holds *two* pixels: the high
// half goes to even columns and the low half to odd ones, because the RDP fills
// 32 bits at a time.
func (m *Machine) fillRect(w uint64) {
	r := &m.rdp
	if ct := r.cycleType(); ct != cycleFill && ct != cycleCopy {
		m.CPU.Halt("unmodelled Fill_Rectangle in cycle type %d (only FILL and COPY are modelled)", ct)
		return
	}
	xl := uint32(w>>44&0xFFF) >> 2
	yl := uint32(w>>32&0xFFF) >> 2
	xh := uint32(w>>12&0xFFF) >> 2
	yh := uint32(w&0xFFF) >> 2

	// The encoded lower-right corner is inclusive.
	xh, yh, xl, yl, ok := r.clip(xh, yh, xl+1, yl+1)
	if !ok {
		return
	}
	for y := yh; y < yl; y++ {
		for x := xh; x < xl; x++ {
			switch r.Color.Size {
			case size16:
				v := uint16(r.FillColor >> 16)
				if x&1 == 1 {
					v = uint16(r.FillColor)
				}
				m.storeRDRAM16(r.pixelAddr(x, y), v)
			case size32:
				m.storeRDRAM32(r.pixelAddr(x, y), r.FillColor)
			default:
				m.CPU.Halt("unmodelled Fill_Rectangle into a %d-bit colour image", 4<<r.Color.Size)
				return
			}
		}
	}
}

// texRect draws a screen-aligned rectangle of texels. In COPY mode the texel is
// written straight through; in 1CYCLE it would pass through the combiner, which
// is not modelled yet.
func (m *Machine) texRect(w []uint64, flip bool) {
	r := &m.rdp
	ct := r.cycleType()
	if ct != cycleCopy {
		m.CPU.Halt("unmodelled Texture_Rectangle in cycle type %d (only COPY is modelled)", ct)
		return
	}

	xl := uint32(w[0]>>44&0xFFF) >> 2
	yl := uint32(w[0]>>32&0xFFF) >> 2
	tileIdx := uint32(w[0] >> 24 & 7)
	xh := uint32(w[0]>>12&0xFFF) >> 2
	yh := uint32(w[0]&0xFFF) >> 2

	s0 := int32(int16(uint16(w[1] >> 48)))
	t0 := int32(int16(uint16(w[1] >> 32)))
	dsdx := int32(int16(uint16(w[1] >> 16)))
	dtdy := int32(int16(uint16(w[1])))
	if flip {
		dsdx, dtdy = dtdy, dsdx
	}

	cxh, cyh, cxl, cyl, ok := r.clip(xh, yh, xl+1, yl+1)
	if !ok {
		return
	}
	tl := &r.Tiles[tileIdx]

	// In COPY mode the texture coordinates step in 5.10 fixed point per pixel.
	for y := cyh; y < cyl; y++ {
		tv := t0 + dtdy*int32(y-yh)
		for x := cxh; x < cxl; x++ {
			sv := s0 + dsdx*int32(x-xh)
			rr, gg, bb, aa, ok := m.texel(tl, uint32(sv>>5), uint32(tv>>5))
			if !ok {
				return // texel() halted with a reason
			}
			// COPY mode writes the texel unchanged, and a zero alpha is a hole.
			if aa == 0 && r.OtherModes&(1<<12) != 0 { // alpha compare enabled
				continue
			}
			m.writePixel(x, y, rr, gg, bb, aa)
		}
	}
}

// texel reads a texel out of TMEM through a tile descriptor.
func (m *Machine) texel(t *tile, s, tt uint32) (uint32, uint32, uint32, uint32, bool) {
	// Wrap into the tile's declared extent. Mirroring and clamping are not
	// modelled; nothing seen so far enables them on a copied rectangle.
	if t.MaskS != 0 {
		s &= 1<<t.MaskS - 1
	}
	if t.MaskT != 0 {
		tt &= 1<<t.MaskT - 1
	}

	switch {
	case t.Format == fmtRGBA && t.Size == size16:
		off := (t.TMem*8 + tt*t.Line*8 + s*2) & 0xFFF
		v := uint16(m.rdp.TMem[off])<<8 | uint16(m.rdp.TMem[off+1])
		return uint32(v>>11&31) << 3, uint32(v>>6&31) << 3, uint32(v>>1&31) << 3, uint32(v&1) * 255, true
	case t.Format == fmtI && t.Size == size8:
		off := (t.TMem*8 + tt*t.Line*8 + s) & 0xFFF
		i := uint32(m.rdp.TMem[off])
		return i, i, i, i, true
	case t.Format == fmtIA && t.Size == size8:
		off := (t.TMem*8 + tt*t.Line*8 + s) & 0xFFF
		v := m.rdp.TMem[off]
		i := uint32(v>>4) * 17
		a := uint32(v&15) * 17
		return i, i, i, a, true
	}
	m.CPU.Halt("unmodelled texture format %d size %d in a copied rectangle", t.Format, t.Size)
	return 0, 0, 0, 0, false
}

// --- TMEM loads -------------------------------------------------------------

// loadBlock copies a run of texels from the texture image straight into TMEM,
// without regard for rows. It is the fast path a game uses when a texture's
// layout in RDRAM already matches TMEM's.
func (m *Machine) loadBlock(w uint64) {
	r := &m.rdp
	tileIdx := uint32(w >> 24 & 7)
	sl := uint32(w >> 44 & 0xFFF)
	tl := uint32(w >> 32 & 0xFFF)
	sh := uint32(w >> 12 & 0xFFF)
	dxt := uint32(w & 0xFFF)
	_ = dxt // the interleave applied to successive rows; not modelled

	t := &r.Tiles[tileIdx]
	bpt := texelBytes(r.Texture.Size)
	if bpt == 0 {
		m.CPU.Halt("unmodelled Load_Block from a %d-bit texture image", 4<<r.Texture.Size)
		return
	}
	src := r.Texture.Addr + (tl*r.Texture.Width+sl)*bpt
	n := (sh - sl + 1) * bpt
	dst := t.TMem * 8
	for i := uint32(0); i < n && dst+i < 4096; i++ {
		if int(src+i) < len(m.RDRAM) {
			r.TMem[dst+i] = m.RDRAM[src+i]
		}
	}
	if dxt != 0 {
		m.note("RDP: Load_Block with dxt=%d (row interleave is not modelled)", dxt)
	}
}

// loadTile copies a rectangle of texels row by row into TMEM.
func (m *Machine) loadTile(w uint64) {
	r := &m.rdp
	tileIdx := uint32(w >> 24 & 7)
	sl := uint32(w>>44&0xFFF) >> 2
	tl := uint32(w>>32&0xFFF) >> 2
	sh := uint32(w>>12&0xFFF) >> 2
	th := uint32(w&0xFFF) >> 2

	t := &r.Tiles[tileIdx]
	bpt := texelBytes(r.Texture.Size)
	if bpt == 0 {
		m.CPU.Halt("unmodelled Load_Tile from a %d-bit texture image", 4<<r.Texture.Size)
		return
	}
	for row := tl; row <= th; row++ {
		src := r.Texture.Addr + (row*r.Texture.Width+sl)*bpt
		dst := t.TMem*8 + (row-tl)*t.Line*8
		n := (sh - sl + 1) * bpt
		for i := uint32(0); i < n && dst+i < 4096; i++ {
			if int(src+i) < len(m.RDRAM) {
				r.TMem[dst+i] = m.RDRAM[src+i]
			}
		}
	}
	// Loading a tile also sets its extent.
	t.SL, t.TL, t.SH, t.TH = sl<<2, tl<<2, sh<<2, th<<2
}

// loadTLUT copies palette entries into the upper half of TMEM, where the
// colour-indexed formats look them up.
func (m *Machine) loadTLUT(w uint64) {
	r := &m.rdp
	tileIdx := uint32(w >> 24 & 7)
	sl := uint32(w>>44&0xFFF) >> 2
	sh := uint32(w>>12&0xFFF) >> 2

	t := &r.Tiles[tileIdx]
	for i := sl; i <= sh; i++ {
		src := r.Texture.Addr + i*2
		dst := (t.TMem*8 + (i-sl)*8) & 0xFFF
		if int(src)+1 < len(m.RDRAM) && dst+1 < 4096 {
			r.TMem[dst] = m.RDRAM[src]
			r.TMem[dst+1] = m.RDRAM[src+1]
		}
	}
}

// texelBytes is a pixel size in bytes, or zero for the sub-byte format.
func texelBytes(size uint32) uint32 {
	switch size {
	case size8:
		return 1
	case size16:
		return 2
	case size32:
		return 4
	}
	return 0
}

// triangle rasterises one of the eight triangle commands. It is the bulk of the
// RDP's work and lands next; for now the command names itself so the frontier is
// explicit.
func (m *Machine) triangle(op uint32, w []uint64) {
	m.CPU.Halt("unmodelled RDP triangle %s (the span rasteriser is not built yet)", cmdNameOf(op))
}
