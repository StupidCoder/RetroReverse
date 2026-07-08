package threedo

import "fmt"

// graphicsfolio.go high-level-emulates the Portfolio "Graphics" folio — screens,
// VDLs and the cel engine's DrawCels. Like the File folio it is reached through
// a negative-offset vector table: the game LookupItem's the "Graphics" folio to
// gfxFolioBase (folio.go) and calls `LDR pc, [base, #-N]`, which the planted
// vectors route into serviceGraphicsFolio below.
//
// The vector numbering follows the era-correct SDK 1.2 graphics.h routine
// indices (byte offset = 4 × index): the 2.5-era table is shifted and led an
// earlier session to mislabel these calls. The ones the boot exercises:
//
//	-0x10 GetPixelAddress   -0x30 CreateScreenGroup  -0x34 SetReadAddress
//	-0x38 ResetReadAddress  -0x5C DeleteScreenGroup  -0x64 SetBGPen
//	-0x68 AddScreenGroup    -0x70 SetClipWidth       -0x88 FillRect
//	-0x98 SetCEControl      -0xA0 DisplayScreen      -0xA4 SetVDL
//	-0xA8 SubmitVDL         -0xAC DrawCels           -0xB0 DrawScreenCels
//	-0xC0 (EA VBL-counter register: r0=enable, r1=&elapsedFieldCount)
//
// CreateScreenGroup builds real Screen and Bitmap structs (graphics.h layouts,
// 0x24-byte ItemNode + 0x20-byte Lists): the game reads Screen+0x78
// (scr_TempBitmap) → Bitmap+0x18 (bm.n_Item) to get the bitmap item it draws
// on, and Bitmap+0x24 (bm_Buffer) is the frame buffer the HLE's DrawCels
// rasterizes into. DisplayScreen records which buffer is front, so the oracle's
// -shot captures what the console would scan out.

// Screen/Bitmap struct field offsets (graphics.h, SDK 1.2).
const (
	scrBitmapCount = 0x34 // Screen.scr_BitmapCount
	scrTempBitmap  = 0x78 // Screen.scr_TempBitmap (Bitmap*)
	bmBuffer       = 0x24 // Bitmap.bm_Buffer
	bmWidth        = 0x28 // Bitmap.bm_Width
	bmHeight       = 0x2C // Bitmap.bm_Height
	bmClipWidth    = 0x38 // Bitmap.bm_ClipWidth
	bmClipHeight   = 0x3C // Bitmap.bm_ClipHeight
)

// CreateScreenGroup TagArgs (graphics.h CSG_TAG_*).
const (
	csgDisplayHeight = 1
	csgScreenCount   = 2
	csgScreenHeight  = 3
	csgBitmapCount   = 4
	csgBitmapWidths  = 5
	csgBitmapHeights = 6
	csgBitmapBufs    = 7
)

// CCB flag bits (hardware.h) used by the software cel engine.
const (
	ccbSkip   = 0x80000000
	ccbLast   = 0x40000000
	ccbNPAbs  = 0x20000000
	ccbSPAbs  = 0x10000000
	ccbPPAbs  = 0x08000000
	ccbLDSize = 0x04000000 // ccb_Width/Height carry the geometry
	ccbLDPLUT = 0x00800000
	ccbCCBPre = 0x00400000
	ccbPOVER  = 0x00000180 // P-mode override (which PPMP word to use)
	ccbBGND   = 0x00000020 // draw index-0 pixels instead of skipping them
	// ccbPacked (0x200) is defined in cel.go.

	pre1LRForm = 0x00000800 // PRE1 bit: source is in linear-framebuffer layout
)

// gfxBitmap is the HLE's record of a Bitmap item: where its pixels live.
type gfxBitmap struct {
	buf  uint32
	w, h int
}

// gfxFuncName labels a Graphics-folio vector offset (SDK 1.2 numbering).
func gfxFuncName(foff uint32) string {
	names := map[uint32]string{
		0x04: "MapSprite", 0x08: "ReadPixel", 0x0C: "ReadVDLColor",
		0x10: "GetPixelAddress", 0x14: "WritePixel", 0x24: "DrawText16",
		0x2C: "ResetFont", 0x30: "CreateScreenGroup", 0x34: "SetReadAddress",
		0x38: "ResetReadAddress", 0x3C: "SetClipOrigin", 0x50: "SetScreenColor",
		0x58: "SetScreenColors", 0x5C: "DeleteScreenGroup", 0x60: "SetFGPen",
		0x64: "SetBGPen", 0x68: "AddScreenGroup", 0x6C: "RemoveScreenGroup",
		0x70: "SetClipWidth", 0x74: "SetClipHeight", 0x78: "MoveTo",
		0x80: "DrawChar", 0x84: "DrawTo", 0x88: "FillRect",
		0x90: "GetCurrentFont", 0x94: "DrawText8", 0x98: "SetCEControl",
		0xA0: "DisplayScreen", 0xA4: "SetVDL", 0xA8: "SubmitVDL",
		0xAC: "DrawCels", 0xB0: "DrawScreenCels", 0xB4: "SetCEWatchDog",
		0xB8: "GetFirstDisplayInfo", 0xBC: "ModifyVDL", 0xC0: "RegisterVBLCounter",
	}
	if n, ok := names[foff]; ok {
		return n
	}
	return "?"
}

// serviceGraphicsFolio dispatches an intercepted Graphics-folio vector call by
// its (positive) byte offset and returns to the caller with a result in r0.
// Calls not yet modelled return 0 (a non-negative success the game's error
// checks accept) and are logged with their call site.
func (m *Machine) serviceGraphicsFolio(foff uint32) {
	c := m.CPU
	r0, r1, r2 := c.Reg(0), c.Reg(1), c.Reg(2)
	m.note(fmt.Sprintf("Graphics[-0x%X] %s from 0x%08X (r0=0x%08X r1=0x%08X r2=0x%08X)",
		foff, gfxFuncName(foff), c.Reg(14)-8, r0, r1, r2))

	switch foff {
	case 0x30: // CreateScreenGroup(Item *screenItems, TagArg *tags)
		m.SetResultAndReturn(m.createScreenGroup(r0, r1))
	case 0x10: // GetPixelAddress(screen|bitmap Item, x, y)
		m.SetResultAndReturn(m.pixelAddress(int32(r0), r1, r2))
	case 0xA8: // SubmitVDL(rawVDL, length, type) -> VDL item
		it := m.createItem(0x0206, 0, 0) // graphics folio VDL node
		m.SetResultAndReturn(uint32(it.num))
	case 0xA4: // SetVDL(screenItem, vdlItem) -> previous VDL item (0 = none)
		m.SetResultAndReturn(0)
	case 0xA0: // DisplayScreen(screenItem, otherScreenItem)
		if bmItem, ok := m.screenBM[int32(r0)]; ok {
			if bm, ok := m.bitmaps[bmItem]; ok {
				m.displayBuf = bm.buf
			}
		}
		m.SetResultAndReturn(0)
	case 0xAC: // DrawCels(bitmapItem, CCB*)
		m.SetResultAndReturn(m.drawCels(int32(r0), r1))
	case 0xB0: // DrawScreenCels(screenItem, CCB*): draw on the screen's bitmap
		if bmItem, ok := m.screenBM[int32(r0)]; ok {
			m.SetResultAndReturn(m.drawCels(bmItem, r1))
			return
		}
		m.SetResultAndReturn(0)
	case 0xC0: // EA VBL-counter register: r0=enable, r1=&elapsedFieldCount.
		// The game hands the folio the address of its own elapsed-field counter
		// for the VBL interrupt to advance; the virtual VBL manager keeps it in
		// step (retires the hardcoded -vblmirror default).
		if r1 != 0 {
			m.SetVBLMirror(r1)
		}
		m.SetResultAndReturn(0)
	default:
		// Non-negative default so the game's MI (error) checks pass.
		m.SetResultAndReturn(0)
	}
}

// createScreenGroup builds the ScreenGroup → Screen → Bitmap item structs from
// a CSG tag list, records the bitmaps' pixel buffers, writes the screen item
// numbers into the caller's array and returns the group item.
func (m *Machine) createScreenGroup(itemArray, tags uint32) uint32 {
	// Tag walk with CreateScreenGroup defaults (320x240, one screen, one bitmap).
	screenCount, bitmapCount := uint32(1), uint32(1)
	width, height := uint32(320), uint32(240)
	var bufArray, widthArray, heightArray uint32
	for p := tags; p != 0; p += 8 {
		tag, val := m.read32(p), m.read32(p+4)
		if tag == 0 {
			break
		}
		switch tag & 0xFFFF {
		case csgScreenCount:
			screenCount = val
		case csgBitmapCount:
			bitmapCount = val
		case csgScreenHeight, csgDisplayHeight:
			height = val
		case csgBitmapWidths:
			widthArray = val
		case csgBitmapHeights:
			heightArray = val
		case csgBitmapBufs:
			bufArray = val
		}
	}
	if screenCount == 0 || screenCount > 8 || bitmapCount == 0 || bitmapCount > 8 {
		return ^uint32(0)
	}

	group := m.createItem(0x0201, 0, 0) // ScreenGroup node
	for s := uint32(0); s < screenCount; s++ {
		var firstBM *item
		for b := uint32(0); b < bitmapCount; b++ {
			w, h := width, height
			if widthArray != 0 {
				w = m.read32(widthArray + b*4)
			}
			if heightArray != 0 {
				h = m.read32(heightArray + b*4)
			}
			buf := uint32(0)
			if bufArray != 0 {
				buf = m.read32(bufArray + (s*bitmapCount+b)*4)
			}
			if buf == 0 {
				// The folio allocates VRAM when the caller supplies no buffers.
				buf = m.vheap.alloc(w * h * 2)
			}
			bm := m.createItem(0x0203, 0, 0x60) // Bitmap node
			m.writeWord(bm.addr+bmBuffer, buf)
			m.writeWord(bm.addr+bmWidth, w)
			m.writeWord(bm.addr+bmHeight, h)
			m.writeWord(bm.addr+bmClipWidth, w)
			m.writeWord(bm.addr+bmClipHeight, h)
			m.bitmaps[bm.num] = gfxBitmap{buf: buf, w: int(w), h: int(h)}
			if firstBM == nil {
				firstBM = bm
			}
		}
		scr := m.createItem(0x0202, 0, 0x90) // Screen node
		m.writeWord(scr.addr+scrBitmapCount, bitmapCount)
		m.writeWord(scr.addr+scrTempBitmap, firstBM.addr)
		m.screenBM[scr.num] = firstBM.num
		if m.displayBuf == 0 {
			m.displayBuf = m.bitmaps[firstBM.num].buf
		}
		if itemArray != 0 {
			m.writeWord(itemArray+s*4, uint32(scr.num))
		}
		m.note(fmt.Sprintf("  screen %d: item %d bitmap %d buf=0x%08X %dx%d",
			s, scr.num, firstBM.num, m.bitmaps[firstBM.num].buf, width, height))
	}
	return uint32(group.num)
}

// pixelAddress returns the address of pixel (x,y) in the item's bitmap using
// the hardware's interleaved line-pair layout: two screen lines share a
// 4-byte-per-column stripe, even lines in the first halfword.
func (m *Machine) pixelAddress(itemNum int32, x, y uint32) uint32 {
	bmItem := itemNum
	if bi, ok := m.screenBM[itemNum]; ok { // a screen item names its bitmap
		bmItem = bi
	}
	bm, ok := m.bitmaps[bmItem]
	if !ok {
		return 0
	}
	return bm.buf + (y>>1)*uint32(bm.w)*4 + x*4 + (y&1)*2
}

// drawCels software-renders a CCB chain into a bitmap — the cel engine's job on
// hardware. Each CCB's source data is decoded with the same packed/unpacked
// logic as .cel files (cel.go); position comes from ccb_X/Y (16.16) and scale
// from ccb_HDX (12.20) / ccb_VDY (16.16), rendered nearest-neighbour. PIXC
// blend modes are not modelled: pixels replace.
func (m *Machine) drawCels(bitmapItem int32, ccb uint32) uint32 {
	bm, ok := m.bitmaps[bitmapItem]
	if !ok || bm.buf == 0 {
		m.note(fmt.Sprintf("DrawCels: unknown bitmap item %d", bitmapItem))
		return 0
	}
	drawn := 0
	for n := 0; ccb != 0 && n < 4096; n++ {
		flags := m.read32(ccb)
		next := m.read32(ccb + 0x04)
		src := m.read32(ccb + 0x08)
		plut := m.read32(ccb + 0x0C)
		if flags&ccbNPAbs == 0 && next != 0 {
			next = ccb + 0x04 + next + 4 // relative to the field, +4
		}
		if flags&ccbSPAbs == 0 && src != 0 {
			src = ccb + 0x08 + src + 4
		}
		if flags&ccbPPAbs == 0 && plut != 0 {
			plut = ccb + 0x0C + plut + 4
		}
		if flags&ccbSkip == 0 {
			if m.drawOneCel(bm, ccb, flags, src, plut) {
				drawn++
			}
		}
		if flags&ccbLast != 0 {
			break
		}
		ccb = next
	}
	if drawn > 0 {
		m.note(fmt.Sprintf("DrawCels: %d cel(s) -> buf 0x%08X", drawn, bm.buf))
	}
	return 0
}

// drawOneCel decodes a single cel's pixels and maps them into the bitmap
// through the 3DO cel engine's corner walk: from the CCB's position (16.16)
// and the per-pixel/per-row deltas HDX/HDY (12.20), VDX/VDY (16.16) plus the
// second-order HDDX/HDDY (12.20), each source pixel lands at
//
//	xstart_r = XPos + r*VDX          hdx_r = HDX + r*HDDX
//	X(c,r)   = xstart_r + c*(hdx_r>>4)  (>>4 rebases 12.20 into 16.16)
//
// which is exactly the incremental corner engine, expressed closed-form so it
// works with both the packed and unpacked decoders that hand back (c,r,v).
// This is what gives the road/scene cels their per-scanline perspective (HDDX
// shearing each row). Pixels blend into the destination through the PIXC
// pixel processor (see blendPixel).
func (m *Machine) drawOneCel(bm gfxBitmap, ccb, flags, src, plutPtr uint32) bool {
	if src == 0 {
		return false
	}
	xPos := int64(int32(m.read32(ccb + 0x10))) // 16.16
	yPos := int64(int32(m.read32(ccb + 0x14)))
	hdx := int64(int32(m.read32(ccb + 0x18))) // 12.20
	hdy := int64(int32(m.read32(ccb + 0x1C)))
	vdx := int64(int32(m.read32(ccb + 0x20))) // 16.16
	vdy := int64(int32(m.read32(ccb + 0x24)))
	hddx := int64(int32(m.read32(ccb + 0x28))) // 12.20
	hddy := int64(int32(m.read32(ccb + 0x2C)))
	pixc := m.read32(ccb + 0x30)

	var pre0, pre1 uint32
	if flags&ccbCCBPre != 0 {
		pre0 = m.read32(ccb + 0x34)
		pre1 = m.read32(ccb + 0x38)
	} else {
		// Preamble word(s) lead the source data: PRE0 always, PRE1 only for
		// unpacked cels (packed lines carry their own offsets).
		pre0 = m.read32(src)
		src += 4
		if flags&ccbPacked == 0 {
			pre1 = m.read32(src)
			src += 4
		}
	}

	cel := &Cel{
		Flags:  flags,
		PRE0:   pre0,
		PRE1:   pre1,
		Packed: flags&ccbPacked != 0,
		BPP:    bppFromPRE0[pre0&7],
		Width:  int(pre1&0x7FF) + 1,
		Height: int((pre0>>6)&0x3FF) + 1,
	}
	if cel.BPP == 0 {
		return false
	}
	cel.Coded = cel.BPP <= 8
	lrform := pre1&pre1LRForm != 0
	// LDSIZE says the CCB's ccb_Width/Height carry the geometry — but the CCBs
	// NFS uses are the 0x40-byte "temporary" form without those fields, so only
	// honour them when they read back sane; otherwise fall to the PRE words.
	if flags&ccbLDSize != 0 {
		if w := int(m.read32(ccb + 0x3C)); w > 0 && w <= 2048 {
			cel.Width = w
		}
		if h := int(m.read32(ccb + 0x40)); h > 0 && h <= 2048 {
			cel.Height = h
		}
	}
	if cel.Packed {
		// Packed lines are self-delimiting (per-line offset + EOL packets), so the
		// width is just the clip-edge loop bound, not a stored dimension.
		cel.Width = bm.w
	}
	if cel.Width <= 0 || cel.Height <= 0 || cel.Width > 2048 || cel.Height > 2048 {
		return false
	}

	// Pull the source bytes (bounded) and the PLUT out of machine memory.
	max := cel.Height*((cel.Width*cel.BPP+31)/32)*4 + cel.Height*4 + 64
	if max > 0x100000 {
		max = 0x100000
	}
	data := make([]byte, max)
	for i := range data {
		data[i] = m.Read(src + uint32(i))
	}
	cel.PDAT = data
	if cel.Coded && plutPtr != 0 {
		for i := 0; i < 32; i++ {
			cel.PLUT = append(cel.PLUT, uint16(m.Read(plutPtr+uint32(i*2)))<<8|uint16(m.Read(plutPtr+uint32(i*2)+1)))
		}
	}

	// mapCorner projects a source-grid corner (column c, row r) to the
	// destination in 16.16 — the incremental corner walk in closed form.
	mapCorner := func(c, r int64) (int64, int64) {
		hdxr := hdx + r*hddx
		hdyr := hdy + r*hddy
		return xPos + r*vdx + c*(hdxr>>4), yPos + r*vdy + c*(hdyr>>4)
	}

	bgnd := flags&ccbBGND != 0
	put := func(sx, sy int, v uint32) {
		var pix uint16
		if cel.Coded {
			if int(v) >= len(cel.PLUT) {
				return
			}
			if v == 0 && !bgnd {
				return // index 0 is the transparent background unless CCB_BGND
			}
			pix = cel.PLUT[v]
		} else {
			if v&0x7FFF == 0 && !bgnd {
				return // an all-zero color word is transparent
			}
			pix = uint16(v)
		}
		// A texel maps to the segment from its corner toward the next column's
		// corner (along the row's HDX/HDY direction). When the cel is magnified —
		// a projected road texture can stretch one source pixel across many
		// destination pixels — plotting only the corner leaves holes, so walk the
		// segment; at 1:1 or minified it is the single corner pixel. Adjacent rows
		// cover the vertical extent. Following the row direction (rather than a
		// bounding box) keeps sheared cels from smearing into a vertical strip.
		c, r := int64(sx), int64(sy)
		x0, y0 := mapCorner(c, r)
		x1, y1 := mapCorner(c+1, r)
		span := (x1 - x0)
		if span < 0 {
			span = -span
		}
		if d := y1 - y0; d > span {
			span = d
		} else if -d > span {
			span = -d
		}
		steps := span >> 16
		// Only bridge genuine magnification. A near-degenerate projection (a texel
		// at the vanishing point) blows the step up to a screen-long streak, so
		// past a small span just plot the corner — the neighbouring texels cover it.
		if steps < 1 || steps > 12 {
			steps = 1
		}
		for s := int64(0); s < steps; s++ {
			dx := x0 + (x1-x0)*s/steps
			dy := y0 + (y1-y0)*s/steps
			x, y := int(dx>>16), int(dy>>16)
			if x < 0 || y < 0 || x >= bm.w || y >= bm.h {
				continue
			}
			m.blendPixel(bm, x, y, pix, pixc, flags)
		}
	}

	if lrform && !cel.Packed && cel.BPP == 16 {
		// A linear-framebuffer-format source: its 16bpp pixels are laid out in
		// the same interleaved line-pair stripes as a bitmap, not row-major.
		m.decodeLRForm16(cel, src, put)
	} else if cel.Packed {
		cel.decodePacked(put)
	} else {
		cel.decodeUnpacked(put)
	}
	return true
}

// decodeLRForm16 walks a 16bpp LRFORM source: its pixels are stored in the
// hardware's interleaved line-pair layout (two rows share a 4-byte column
// stripe, even row in the first halfword) rather than the standard row-major
// cel stride. Used for full-screen background cels copied from a bitmap.
func (m *Machine) decodeLRForm16(c *Cel, src uint32, set func(x, y int, v uint32)) {
	for y := 0; y < c.Height; y++ {
		row := src + uint32(y>>1)*uint32(c.Width)*4 + uint32(y&1)*2
		for x := 0; x < c.Width; x++ {
			a := row + uint32(x)*4
			set(x, y, uint32(m.Read(a))<<8|uint32(m.Read(a+1)))
		}
	}
}

// blendPixel combines a source RGB555 pixel with the destination through the
// 3DO cel-engine pixel processor (PPMP) and stores the result. PIXC holds two
// 16-bit PPMPC words; the P-mode (POVER, or the pixel's own bit-15 P-bit)
// selects which. For each 5-bit channel:
//
//	first  = (input1 * (MF+1)) >> PDV(dv1)     PDV(n) = ((n-1)&3)+1
//	second = 0 | AV | destination | source     (per the 2S field)
//	out    = clamp5((first + second) >> dv2)
//
// where input1 is the source pixel (or the destination, if 1S) and dv2 is the
// low bit. This is the general processor, matched to the stock modes: NORMAL
// (0x1F40) yields the source unchanged, AVERAGE (0x1F81) yields (src+dst)/2.
// The AV-as-control (CCB_USEAV) and PXOR paths are not modelled; NFS's cels
// use the plain modes.
func (m *Machine) blendPixel(bm gfxBitmap, x, y int, pix uint16, pixc, flags uint32) {
	a := bm.buf + uint32(y>>1)*uint32(bm.w)*4 + uint32(x)*4 + uint32(y&1)*2

	// Select the PPMPC word.
	word := pixc & 0xFFFF
	switch flags & ccbPOVER {
	case 0x100: // PMODE_ZERO: force the low word
	case 0x180: // PMODE_ONE: force the high word
		word = pixc >> 16
	default: // PMODE_PDC: the pixel's own P-bit picks the word
		if pix&0x8000 != 0 {
			word = pixc >> 16
		}
	}

	mul := ((word >> 10) & 7) + 1          // MF: multiply by 1..8
	sh1 := ((((word >> 8) & 3) - 1) & 3) + 1 // dv1 -> PDV shift (1..4)
	s1 := word&0x8000 != 0                  // first source: destination not cel
	s2 := (word >> 6) & 3                   // second source select
	dv2 := (word & 1)                       // final divide shift (0 or 1)

	// Fast path: opaque source copy (NORMAL and its P-mode twins) — no blend, so
	// the destination need not be read.
	if !s1 && s2 == 0 && dv2 == 0 && mul == 8 && sh1 == 3 {
		m.Write(a, byte(pix>>8))
		m.Write(a+1, byte(pix))
		return
	}

	dst := uint16(m.Read(a))<<8 | uint16(m.Read(a+1))
	r1, g1, b1 := chan5(pix)
	if s1 {
		r1, g1, b1 = chan5(dst)
	}
	r1 = (r1 * mul) >> sh1
	g1 = (g1 * mul) >> sh1
	b1 = (b1 * mul) >> sh1

	var r2, g2, b2 uint32
	switch s2 {
	case 0:
		// second source is zero
	case 1:
		av := (word >> 1) & 0x1F // constant AV, same on each channel
		r2, g2, b2 = av, av, av
	case 2:
		r2, g2, b2 = chan5(dst) // current destination pixel
	case 3:
		r2, g2, b2 = chan5(pix) // the source pixel
	}
	out := clamp5((r1+r2)>>dv2)<<10 | clamp5((g1+g2)>>dv2)<<5 | clamp5((b1+b2)>>dv2)
	m.Write(a, byte(out>>8))
	m.Write(a+1, byte(out))
}

// chan5 splits an RGB555 pixel into its three 5-bit channels.
func chan5(p uint16) (r, g, b uint32) {
	return uint32(p>>10) & 0x1F, uint32(p>>5) & 0x1F, uint32(p) & 0x1F
}

// clamp5 saturates a channel value to 5 bits.
func clamp5(v uint32) uint16 {
	if v > 0x1F {
		return 0x1F
	}
	return uint16(v)
}
