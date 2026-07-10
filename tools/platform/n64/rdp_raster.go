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

// texRect draws a screen-aligned rectangle of texels.
//
// In COPY mode the texel goes straight to the framebuffer. In 1-cycle and
// 2-cycle mode it passes through the colour combiner like any other pixel, which
// is how a game tints a sprite or fades one out.
func (m *Machine) texRect(w []uint64, flip bool) {
	r := &m.rdp
	ct := r.cycleType()
	if ct == cycleFill {
		m.CPU.Halt("Texture_Rectangle in FILL mode, which the hardware does not define")
		return
	}

	xl := uint32(w[0]>>44&0xFFF) >> 2
	yl := uint32(w[0]>>32&0xFFF) >> 2
	tileIdx := uint32(w[0] >> 24 & 7)
	xh := uint32(w[0]>>12&0xFFF) >> 2
	yh := uint32(w[0]&0xFFF) >> 2

	// The texture coordinates are 10.5, and their per-pixel steps 5.10.
	s0 := int32(int16(uint16(w[1] >> 48)))
	t0 := int32(int16(uint16(w[1] >> 32)))
	dsdx := int32(int16(uint16(w[1] >> 16)))
	dtdy := int32(int16(uint16(w[1])))
	if flip {
		// The flipped form transposes the rectangle: s advances down the screen
		// and t across it.
		dsdx, dtdy = dtdy, dsdx
	}
	// COPY mode moves four texels per clock, and the encoded step is for the
	// group: a 1:1 blit says dsdx=4.0. Per pixel that is a quarter of it.
	if ct == cycleCopy {
		dsdx >>= 2
	}

	cxh, cyh, cxl, cyl, ok := r.clip(xh, yh, xl+1, yl+1)
	if !ok {
		return
	}
	tile := &r.Tiles[tileIdx]
	prim := unpackColor(r.PrimColor)
	env := unpackColor(r.EnvColor)

	for y := cyh; y < cyl; y++ {
		tv := t0 + dtdy*int32(y-yh)/32
		for x := cxh; x < cxl; x++ {
			// The step product is 5.10; dropping five fraction bits leaves the
			// 10.5 the sampler takes. The coordinate itself stays 10.5 — it was
			// being shifted a second time on the way in, which read texel n/32
			// instead of texel n and smeared every rectangle into its corner
			// texel.
			sv := s0 + dsdx*int32(x-xh)/32
			texel, ok := m.sample(tile, sv, tv)
			if !ok {
				return
			}
			if ct == cycleCopy {
				// COPY bypasses the combiner entirely. A zero alpha is a hole
				// when the alpha test is on, which is how 2-D cut-outs work.
				if r.OtherModes&omAlphaCompare != 0 && texel.A == 0 {
					if m.OnPixel != nil {
						m.OnPixel(x, y, PixelEvent{AlphaReject: true})
					}
					continue
				}
				m.writePixel(x, y, texel.R, texel.G, texel.B, texel.A)
				if m.OnPixel != nil {
					m.OnPixel(x, y, PixelEvent{
						Drawn: true, R: texel.R, G: texel.G, B: texel.B, A: texel.A,
						TexR: texel.R, TexG: texel.G, TexB: texel.B, TexA: texel.A,
					})
				}
				continue
			}
			in := combineInputs{
				Texel0: texel, Texel1: texel, Prim: prim, Env: env,
				Shade: rgba{255, 255, 255, 255},
				texS:  sv, texT: tv,
			}
			m.drawPixel(x, y, &in, 0, false)
			if m.CPU.Halted {
				return
			}
		}
	}
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
		if int(src+i) >= len(m.RDRAM) {
			continue
		}
		// dxt is the per-64-bit-word increment of a 1.11 row counter: it tells
		// the loader when the copy crosses into the next texture row, and the
		// hardware swaps the 32-bit halves of every odd row's words as they
		// land. The sampler swaps them back (swizzle in rdp_texture.go). With
		// dxt=0 the whole load is one "row", nothing is swapped, and a texture
		// authored pre-swizzled reads back through the sampler's swap — which
		// is how this game ships most of its textures.
		d := dst + i
		if dxt != 0 {
			word := i / 8
			if word*dxt>>11&1 != 0 {
				d = dst + (i &^ 7) + ((i & 7) ^ 4)
			}
		}
		if d < 4096 {
			r.TMem[d] = m.RDRAM[src+i]
		}
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
			if int(src+i) >= len(m.RDRAM) {
				continue
			}
			// Rows land in TMEM with the same odd-row word swap Load_Block
			// applies; the sampler undoes it.
			d := dst + i
			if (row-tl)&1 != 0 {
				d = dst + (i &^ 7) + ((i & 7) ^ 4)
			}
			if d < 4096 {
				r.TMem[d] = m.RDRAM[src+i]
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
