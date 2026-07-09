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

// flashClearRange fills a contiguous VRAM byte span with a 16-bit RGB555 value —
// the SPORT FLASHWRITE that paints the frame's background. The game clears each
// buffer as two spans, the sky colour over the top page-run and the ground
// colour over the rest, so the horizon falls at the byte offset where the runs
// meet. Filling the exact [dest, dest+bytes) range reproduces that split (the
// interleaved line-pair layout means the offset maps to a scanline boundary).
// Without it, regions the frame doesn't repaint keep stale content — a
// hall-of-mirrors — and the sky, which is a clear rather than a cel, is absent.
func (m *Machine) flashClearRange(dest, bytes uint32, val uint16) {
	hi, lo := byte(val>>8), byte(val)
	end := dest + bytes
	for a := dest; a+1 < end; a += 2 {
		m.Write(a, hi)
		m.Write(a+1, lo)
	}
	if m.CelDebug {
		m.CelDebugLog = append(m.CelDebugLog, fmt.Sprintf("FLASHCLEAR val=%04X -> [0x%08X,+0x%X)", val, dest, bytes))
	}
}

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
		if m.CelDebug {
			// Frame boundary: keep the just-drawn frame's cels and start fresh, so
			// CelDebugLog holds exactly the frame being put on screen.
			m.CelFrameLog = m.CelDebugLog
			m.CelDebugLog = nil
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
	drawn, skipped, total := 0, 0, 0
	for n := 0; ccb != 0 && n < 4096; n++ {
		total++
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
		} else {
			skipped++
		}
		if flags&ccbLast != 0 {
			break
		}
		ccb = next
	}
	if drawn > 0 {
		m.note(fmt.Sprintf("DrawCels: %d cel(s) -> buf 0x%08X", drawn, bm.buf))
	}
	if m.CelDebug {
		m.CelDebugLog = append(m.CelDebugLog, fmt.Sprintf("== DrawCels chain: %d total, %d drawn, %d skipped -> buf 0x%08X ==", total, drawn, skipped, bm.buf))
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
	persp := hddx != 0 || hddy != 0 || vdx != 0 || hdy != 0
	written := 0
	var calls, clearN, offN int
	put := func(sx, sy int, v uint32) {
		calls++
		pix, amv, transparent := m.decodePixel(cel, v, flags, bgnd)
		if transparent {
			clearN++
			return
		}
		if m.PerspTint && persp {
			pix, amv = 0x7C1F, 0x49 // solid magenta: show where perspective cels land
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
				offN++
				continue
			}
			if m.CelDebug && x == int(m.ProbeX) && y == int(m.ProbeY) {
				m.CelDebugLog = append(m.CelDebugLog, fmt.Sprintf("PROBE (%d,%d) hit by cel src=%08X %dbpp %dx%d pos=(%d,%d) HD=(%X,%X) VD=(%X,%X) lrform=%v flags=%08X",
					x, y, src, cel.BPP, cel.Width, cel.Height, int(xPos>>16), int(yPos>>16), hdx, hdy, vdx, vdy, lrform, flags))
			}
			m.blendPixel(bm, x, y, pix, amv, pixc, flags)
			written++
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
	if m.CelDebug && len(m.CelDebugLog) < 4000 {
		kind := "unpacked"
		if cel.Packed {
			kind = "packed"
		}
		coded := "coded"
		if !cel.Coded {
			coded = "16bpp"
		}
		lr := ""
		if lrform {
			lr = " LRFORM"
		}
		persp := ""
		if hddx != 0 || hddy != 0 || vdx != 0 || hdy != 0 {
			persp = " PERSP"
		}
		m.CelDebugLog = append(m.CelDebugLog, fmt.Sprintf(
			"cel src=%08X %dbpp %s %s%s%s %dx%d pos=(%d,%d) HD=(%X,%X) VD=(%X,%X) HDD=(%X,%X) pixc=%08X flags=%08X calls=%d clear=%d off=%d wrote=%d",
			src, cel.BPP, kind, coded, lr, persp, cel.Width, cel.Height,
			int(xPos>>16), int(yPos>>16), hdx, hdy, vdx, vdy, hddx, hddy, pixc, flags, calls, clearN, offN, written))
	}
	return true
}

// decodeLRForm16 walks a 16bpp LRFORM source — a bitmap in the hardware's
// interleaved line-pair layout, as produced by rendering into a screen (e.g.
// the rear-view mirror's off-screen buffer). Each 32-bit word holds two pixels:
// the first halfword is the even scanline, the second the odd scanline, at the
// same column. So the cel's declared width counts halfwords (two per word/two
// per column pair) and its height counts line-pairs — the real image is
// (Width/2) wide by (Height*2) tall. Decoding it as Width x Height row-major
// stretched the mirror to double width and half height.
func (m *Machine) decodeLRForm16(c *Cel, src uint32, set func(x, y int, v uint32)) {
	cols := c.Width / 2 // 16-bit pixels per scanline = words per line-pair
	for lp := 0; lp < c.Height; lp++ {
		for x := 0; x < cols; x++ {
			w := src + uint32(lp*cols+x)*4
			set(x, lp*2, uint32(m.Read(w))<<8|uint32(m.Read(w+1)))     // even line
			set(x, lp*2+1, uint32(m.Read(w+2))<<8|uint32(m.Read(w+3))) // odd line
		}
	}
}

// decodePixel resolves a raw cel pixel value into an RGB555 color (with the
// P-bit in bit 15), its AMV shading value, and whether it is transparent —
// the cel engine's pixel decoder (PDEC). Coded cels split the value into a
// PLUT index plus per-format control bits: 6bpp is a 5-bit index + a P-bit
// (cp6btag c:5,pw:1), 8bpp a 5-bit index + a 3-bit AMV (cp8btag c:5,mpw:1,m:2);
// 1/2/4bpp are a direct index into a PLUT slice chosen by the CCB's PLUTA bits.
// A pixel is transparent when its resolved color is RGB black and CCB_BGND is
// clear. Indexing PLUT with the full value (dropping the P-bit pixels) was what
// made opaque cels — the dashboard — render only their translucent P=0 pixels.
func (m *Machine) decodePixel(cel *Cel, v, flags uint32, bgnd bool) (uint16, uint32, bool) {
	amv := uint32(0x49) // constant AMV for all but 8bpp coded cels
	if !cel.Coded {
		if v&0x7FFF == 0 && !bgnd {
			return 0, amv, true
		}
		return uint16(v), amv, false
	}

	var idx uint32
	var pw uint16
	switch cel.BPP {
	case 1:
		idx = (flags&0xF)*2 + (v & 0x1)
	case 2:
		idx = (flags&0xE)*2 + (v & 0x3)
	case 4:
		idx = (flags&0x8)*2 + (v & 0xF)
	case 6:
		idx = v & 0x1F
		pw = uint16((v>>5)&1) << 15
	case 8:
		idx = v & 0x1F
		amv = (((v>>6)&3)<<1 | (v>>5)&1) * 0x49 // (m<<1 | mpw) replicated per channel
	default:
		idx = v & 0x1F
	}
	if int(idx) >= len(cel.PLUT) {
		return 0, amv, true
	}
	raw := cel.PLUT[idx]
	color := raw & 0x7FFF
	if color == 0 && !bgnd {
		return 0, amv, true
	}
	if cel.BPP != 6 {
		pw = raw & 0x8000 // P-bit from the palette entry except 6bpp (from the pixel)
	}
	return color | pw, amv, false
}

// blendPixel combines a decoded source pixel with the destination through the
// 3DO cel-engine pixel processor (PPMP/PPROC) and stores the result. PIXC holds
// two 16-bit PPMPC words; the P-mode (POVER, or the pixel's bit-15 P-bit)
// selects which. Per 5-bit channel:
//
//	first  = input1 * (Mult+1) >> PDV(dv1)     PDV(n) = ((n-1)&3)+1
//	out    = clip((first + second) >> dv2)
//
// input1 is the source pixel, or the destination when 1S. The multiplier comes
// from MF (MS=0), the pixel's AMV shading (MS=1), or the pixel itself (MS=2/3).
// second is 0 / a constant / the destination / the source (2S), each shifted by
// the AV dv3 field when CCB_USEAV is set. This reproduces the stock modes —
// NORMAL 0x1F40 = source, AVERAGE 0x1F81 = (src+dst)/2 — and the per-pixel
// P-bit blends NFS uses (opaque dashboard vs translucent glass).
func (m *Machine) blendPixel(bm gfxBitmap, x, y int, pix uint16, amv, pixc, flags uint32) {
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

	s1 := word&0x8000 != 0     // first source: destination, not the cel pixel
	ms := (word >> 13) & 3     // multiply source: MF / AMV / pixel
	mxf := (word >> 10) & 7    // MF multiply factor (Mult+1)
	dv1 := (word >> 8) & 3     // first-source divide (PDV)
	s2 := (word >> 6) & 3      // second source select
	avf := (word >> 1) & 0x1F  // AV field (constant, or control signals if USEAV)
	dv2 := word & 1            // final divide shift

	// Fast path: opaque source copy (NORMAL and its P-mode twins) — no blend, so
	// the destination need not be read.
	if !s1 && ms == 0 && s2 == 0 && dv2 == 0 && mxf == 7 && dv1 == 3 {
		m.Write(a, byte(pix>>8))
		m.Write(a+1, byte(pix))
		return
	}

	// AV control signals (only when CCB_USEAV): dv3 shifts the second source,
	// nCLIP suppresses the saturating clamp.
	var dv3 uint32
	clip := true
	if flags&0x400 != 0 { // CCB_USEAV
		dv3 = (avf >> 3) & 3
		clip = avf&4 == 0 // nCLIP bit set -> no clamp
	}

	dst := uint16(m.Read(a))<<8 | uint16(m.Read(a+1))
	ir, ig, ib := chan5(pix)
	if s1 {
		ir, ig, ib = chan5(dst)
	}
	sh1 := pdv(dv1)
	var c1r, c1g, c1b uint32
	switch ms {
	case 1: // AMV: per-channel multiplier from the pixel's shading value
		c1r = (ir * (((amv >> 6) & 7) + 1)) >> sh1
		c1g = (ig * (((amv >> 3) & 7) + 1)) >> sh1
		c1b = (ib * ((amv & 7) + 1)) >> sh1
	case 2: // pixel self-modulate (multiplier and divide from the channel)
		c1r = (ir * ((ir >> 2) + 1)) >> pdv(ir&3)
		c1g = (ig * ((ig >> 2) + 1)) >> pdv(ig&3)
		c1b = (ib * ((ib >> 2) + 1)) >> pdv(ib&3)
	case 3:
		c1r = (ir * ((ir >> 2) + 1)) >> sh1
		c1g = (ig * ((ig >> 2) + 1)) >> sh1
		c1b = (ib * ((ib >> 2) + 1)) >> sh1
	default: // MS=0: constant MF
		c1r = (ir * (mxf + 1)) >> sh1
		c1g = (ig * (mxf + 1)) >> sh1
		c1b = (ib * (mxf + 1)) >> sh1
	}

	var c2r, c2g, c2b uint32
	switch s2 {
	case 1:
		c2r, c2g, c2b = avf>>dv3, avf>>dv3, avf>>dv3
	case 2:
		dr, dg, db := chan5(dst)
		c2r, c2g, c2b = dr>>dv3, dg>>dv3, db>>dv3
	case 3:
		sr, sg, sb := chan5(pix)
		c2r, c2g, c2b = sr>>dv3, sg>>dv3, sb>>dv3
	}

	clampOr := func(v uint32) uint32 {
		if !clip {
			return v & 0x1F
		}
		if v > 0x1F {
			return 0x1F
		}
		return v
	}
	out := clampOr((c1r+c2r)>>dv2)<<10 | clampOr((c1g+c2g)>>dv2)<<5 | clampOr((c1b+c2b)>>dv2)
	m.Write(a, byte(out>>8))
	m.Write(a+1, byte(out))
}

// pdv is the cel engine's PDV divide-shift: PDV(n) = ((n-1)&3)+1, range 1..4.
func pdv(n uint32) uint32 { return ((n-1)&3 + 1) }

// chan5 splits an RGB555 pixel into its three 5-bit channels.
func chan5(p uint16) (r, g, b uint32) {
	return uint32(p>>10) & 0x1F, uint32(p>>5) & 0x1F, uint32(p) & 0x1F
}
