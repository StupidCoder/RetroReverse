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

// CCB flag bits (form3do.h) used by the software cel engine.
const (
	ccbSkip   = 0x80000000
	ccbLast   = 0x40000000
	ccbNPAbs  = 0x20000000
	ccbSPAbs  = 0x10000000
	ccbPPAbs  = 0x08000000
	ccbLDPLUT = 0x00800000
	ccbCCBPre = 0x00400000
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

// drawOneCel decodes a single cel's pixels and blits them into the bitmap.
func (m *Machine) drawOneCel(bm gfxBitmap, ccb, flags, src, plutPtr uint32) bool {
	if src == 0 {
		return false
	}
	xPos := int32(m.read32(ccb+0x10)) >> 16 // 16.16
	yPos := int32(m.read32(ccb+0x14)) >> 16
	hdx := int32(m.read32(ccb + 0x18)) // 12.20
	vdy := int32(m.read32(ccb + 0x24)) // 16.16

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
	if cel.Packed {
		// Packed sources carry no width preamble; run to the clip edge.
		cel.Width = bm.w
	}
	if cel.Width <= 0 || cel.Height <= 0 || cel.Width > 2048 || cel.Height > 2048 {
		return false
	}

	// Pull the source bytes (bounded) and the PLUT out of machine memory.
	max := cel.Height*((cel.Width*cel.BPP+31)/32)*4 + cel.Height*4 + 64
	if max > 0x80000 {
		max = 0x80000
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

	// Nearest-neighbour scale factors from the per-pixel deltas (1.0 = no scale).
	xStep, yStep := 1.0, 1.0
	if hdx != 0 {
		xStep = float64(hdx) / (1 << 20)
	}
	if vdy != 0 {
		yStep = float64(vdy) / (1 << 16)
	}

	put := func(sx, sy int, v uint32) {
		var pix uint16
		if cel.Coded {
			if int(v) < len(cel.PLUT) {
				pix = cel.PLUT[v]
			} else {
				return
			}
			if v == 0 && pix == 0 {
				return // black index 0: treat as transparent background
			}
		} else {
			if uint16(v) == 0 {
				return // the zero word is transparent
			}
			pix = uint16(v)
		}
		x := xPos + int32(float64(sx)*xStep)
		y := yPos + int32(float64(sy)*yStep)
		if x < 0 || y < 0 || int(x) >= bm.w || int(y) >= bm.h {
			return
		}
		a := bm.buf + uint32(y>>1)*uint32(bm.w)*4 + uint32(x)*4 + uint32(y&1)*2
		m.Write(a, byte(pix>>8))
		m.Write(a+1, byte(pix))
	}

	if cel.Packed {
		cel.decodePacked(put)
	} else {
		cel.decodeUnpacked(put)
	}
	return true
}
