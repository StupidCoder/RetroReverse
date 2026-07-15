package ps2

// gs.go is the Graphics Synthesizer — the PS2's rasteriser and, with it, four megabytes
// of video memory that holds every frame buffer, texture and depth buffer at once. The
// GS has no view of main memory; everything it draws from is first uploaded into this
// local memory across the GIF, and everything it draws is written back into it. The
// display is then a window onto one rectangle of it.
//
// This file is the GS's state: its registers, its 4 MiB, and the one operation the boot
// reaches first — an image upload. sceGsExecLoadImage sets four registers and then streams
// pixels:
//
//	BITBLTBUF   where in GS memory the pixels go (base, width, format), and where they
//	            come from for the other direction
//	TRXPOS      the top-left corner of the destination rectangle
//	TRXREG      the width and height of it
//	TRXDIR      the direction — 0 is host-to-local, the upload — and writing it arms the
//	            transfer
//	HWREG       the pixels themselves, quadword by quadword (the GIF IMAGE data)
//
// GS local memory is not linear. It is tiled: pages of 8 KiB, each holding 32 blocks of
// 256 bytes, each block holding an 8x8 (or wider) patch of pixels in a swizzle that keeps
// a screen-space neighbourhood in one block. addrPSMCT32 works that swizzle out for the
// 32-bit format the frame buffer and most textures use; the transfer walks the destination
// rectangle in raster order and places each pixel where the swizzle says it lives, so that
// what is uploaded reads back correctly when the GS later samples it as a texture or scans
// it out as a frame.

// GS register addresses, as written through the GIF's A+D descriptor.
const (
	gsPRIM       = 0x00
	gsRGBAQ      = 0x01
	gsXYZ2       = 0x05
	gsTEX0_1     = 0x06
	gsFOG        = 0x0A
	gsXYOFFSET1  = 0x18
	gsPRMODECONT = 0x1A
	gsSCISSOR1   = 0x40
	gsTEST1      = 0x47
	gsFRAME1     = 0x4C
	gsZBUF1      = 0x4E
	gsBITBLTBUF  = 0x50
	gsTRXPOS     = 0x51
	gsTRXREG     = 0x52
	gsTRXDIR     = 0x53
	gsHWREG      = 0x54
	gsFINISH     = 0x61
)

// The GS's local memory. It is addressed in "words" of 4 bytes for the register fields
// (a base pointer is in units of 8 KiB pages, 32 blocks of 256 bytes), but the backing
// store is the flat 4 MiB.
const gsVRAMSize = 4 << 20

// GS is the Graphics Synthesizer's state.
type GS struct {
	vram []byte

	// The general register file, indexed by GS register address. Sixty-four 64-bit
	// registers cover the whole map, most of them unused by a boot.
	reg [0x80]uint64

	// The image transfer in progress (host -> local): the destination decoded from
	// BITBLTBUF/TRXPOS/TRXREG, and a raster cursor into the rectangle.
	xfer gsXfer

	// The privileged registers the EE writes directly, not through the GIF.
	csr uint64

	// Counters for the census: how many images were uploaded, and how many drawing
	// primitives were kicked. A boot that only uploads and never draws, or the reverse,
	// says which half of the pipeline to build next.
	uploads int
	prims   int

	// The vertex queue (gsdraw.go): the positions pushed and not yet consumed by a kick,
	// and the per-type and per-feature tallies of what was drawn.
	vq         [3]gsVertex
	vqN        int
	q          uint32 // the Q latched by a PACKED ST, applied by the next PACKED RGBAQ
	primCount  [8]int
	drawCensus map[string]int

	m *Machine
}

// gsXfer is a host-to-local image upload, decoded and in progress.
type gsXfer struct {
	active  bool
	dbp     uint32 // destination base pointer, in 64-word blocks (the GS's unit)
	dbw     uint32 // destination buffer width, in units of 64 pixels
	dpsm    uint32 // destination pixel format
	dsax    uint32 // destination top-left x
	dsay    uint32 // destination top-left y
	rrw     uint32 // rectangle width in pixels
	rrh     uint32 // rectangle height in pixels
	x, y    uint32 // the raster cursor within the rectangle
	partial []byte // bytes carried over between IMAGE quadwords when a pixel straddles them
}

// ensureGS creates the Graphics Synthesizer the first time anything touches it.
func (m *Machine) ensureGS() *GS {
	if m.gs == nil {
		m.gs = &GS{vram: make([]byte, gsVRAMSize), m: m}
	}
	return m.gs
}

// write stores a GS register and acts on the ones that do something when written. The
// value is in the register's own layout (an A+D or REGLIST write); the PACKED layouts
// are decoded by writePacked before they arrive here.
func (gs *GS) write(reg uint8, val uint64) {
	if int(reg) < len(gs.reg) {
		gs.reg[reg] = val
	}
	switch reg {
	case gsPRIM:
		gs.vqN = 0 // starting a primitive resets the queue
	case gsXYZ2:
		gs.pushVertex(int32(val&0xFFFF), int32(val>>16&0xFFFF), uint32(val>>32), true)
	case gsXYZ3:
		gs.pushVertex(int32(val&0xFFFF), int32(val>>16&0xFFFF), uint32(val>>32), false)
	case gsXYZF2:
		gs.pushVertex(int32(val&0xFFFF), int32(val>>16&0xFFFF), uint32(val>>32&0xFFFFFF), true)
	case gsXYZF3:
		gs.pushVertex(int32(val&0xFFFF), int32(val>>16&0xFFFF), uint32(val>>32&0xFFFFFF), false)
	case gsTRXDIR:
		gs.beginTransfer(val & 3)
	case gsHWREG:
		gs.imageData(u64bytes(val))
	case gsFINISH:
		// The game asks the GS to signal when its queue drains; the CSR's FINISH bit is
		// what it polls. With no pipeline to drain, it is done immediately.
		gs.csr |= 1 << 1
	}
}

// writePacked decodes one PACKED-format register write into the register's own layout.
// PACKED spreads a register's fields across the quadword — a colour byte per word, a
// position halfword per word — and two descriptors are not registers at all: A+D (0xE)
// is handled by the caller, and 4/5 carry an ADC bit that turns the kicking write into
// the non-kicking one, which is how a PACKED stream breaks a strip.
func (gs *GS) writePacked(reg uint8, lo, hi uint64) {
	switch reg {
	case gifRegPrim:
		gs.write(gsPRIM, lo&0x7FF)

	case gifRegRGBAQ:
		r, g := lo&0xFF, lo>>32&0xFF
		b, a := hi&0xFF, hi>>32&0xFF
		gs.write(gsRGBAQ, r|g<<8|b<<16|a<<24|uint64(gs.q)<<32)

	case gifRegST:
		gs.q = uint32(hi) // latched; the next PACKED RGBAQ carries it
		gs.write(gsST, lo)

	case gifRegUV:
		u, v := lo&0x3FFF, lo>>32&0x3FFF
		gs.write(gsUV, u|v<<16)

	case gifRegXYZF2:
		x, y := lo&0xFFFF, lo>>32&0xFFFF
		z := hi >> 4 & 0xFFFFFF
		f := hi >> 36 & 0xFF
		val := x | y<<16 | z<<32 | f<<56
		if hi&(1<<47) != 0 { // ADC: push, do not kick
			gs.write(gsXYZF3, val)
		} else {
			gs.write(gsXYZF2, val)
		}

	case gifRegXYZ2:
		x, y := lo&0xFFFF, lo>>32&0xFFFF
		z := hi & 0xFFFFFFFF
		val := x | y<<16 | z<<32
		if hi&(1<<47) != 0 {
			gs.write(gsXYZ3, val)
		} else {
			gs.write(gsXYZ2, val)
		}

	case gsFOG:
		gs.write(gsFOG, hi>>36&0xFF<<56)

	default:
		// Descriptors 0x6..0xD are the register at that address, value in the low half
		// (TEX0, CLAMP, XYZF3/XYZ3 in their raw layouts).
		gs.write(reg, lo)
	}
}

// beginTransfer arms an image transfer from the BITBLTBUF/TRXPOS/TRXREG registers. Only
// the host-to-local direction (0) is the upload the boot uses; the others are named and
// left for when the game reaches them.
func (gs *GS) beginTransfer(dir uint64) {
	if dir != 0 { // 1 local->host, 2 local->local; not on the boot path yet
		gs.xfer.active = false
		return
	}
	bitbltbuf := gs.reg[gsBITBLTBUF]
	trxpos := gs.reg[gsTRXPOS]
	trxreg := gs.reg[gsTRXREG]

	gs.xfer = gsXfer{
		active: true,
		dbp:    uint32(bitbltbuf>>32) & 0x3FFF,
		dbw:    uint32(bitbltbuf>>48) & 0x3F,
		dpsm:   uint32(bitbltbuf>>56) & 0x3F,
		dsax:   uint32(trxpos>>32) & 0x7FF,
		dsay:   uint32(trxpos>>48) & 0x7FF,
		rrw:    uint32(trxreg) & 0xFFF,
		rrh:    uint32(trxreg>>32) & 0xFFF,
	}
	gs.uploads++
	gs.m.note("GS: image upload %dx%d to base 0x%X (width %d, format 0x%X) at (%d,%d)",
		gs.xfer.rrw, gs.xfer.rrh, gs.xfer.dbp*64, gs.xfer.dbw*64, gs.xfer.dpsm,
		gs.xfer.dsax, gs.xfer.dsay)
}

// imageData writes the pixels of an image transfer into GS memory, walking the destination
// rectangle in raster order. It handles the 32-bit format (PSMCT32/PSMCT24) the boot uploads
// in; other formats consume the data and are logged, so a wrong guess is visible rather than
// silent corruption.
func (gs *GS) imageData(data []byte) {
	if !gs.xfer.active || gs.xfer.rrw == 0 || gs.xfer.rrh == 0 {
		return
	}
	x := gs.xfer
	buf := append(x.partial, data...)
	x.partial = nil

	switch x.dpsm {
	case psmCT32, psmCT24:
		const bpp = 4
		i := 0
		for i+bpp <= len(buf) {
			px := le32gs(buf[i:])
			i += bpp
			addr := addrPSMCT32(x.dbp, x.dbw, x.dsax+x.x, x.dsay+x.y)
			if addr+4 <= uint32(len(gs.vram)) {
				gs.vram[addr+0] = byte(px)
				gs.vram[addr+1] = byte(px >> 8)
				gs.vram[addr+2] = byte(px >> 16)
				gs.vram[addr+3] = byte(px >> 24)
			}
			x.x++
			if x.x >= x.rrw {
				x.x = 0
				x.y++
				if x.y >= x.rrh {
					x.active = false
					break
				}
			}
		}
		x.partial = append(x.partial, buf[i:]...)

	default:
		gs.m.note("GS: image upload in format 0x%X (%d bytes) — consumed, not yet placed", x.dpsm, len(buf))
		x.active = false
	}
	gs.xfer = x
}

// The GS pixel-storage formats the boot touches.
const (
	psmCT32 = 0x00
	psmCT24 = 0x01
	psmCT16 = 0x02
)

// addrPSMCT32 returns the byte offset in GS memory of pixel (x, y) in a 32-bit buffer whose
// base is bp (in 64-word blocks) and whose width is bw (in units of 64 pixels).
//
// The GS tiles its memory. A 32-bit buffer is laid out in pages of 32x32 pixels (8 KiB),
// each page a grid of blocks, each block an 8x8 patch, and within a block the pixels are
// swizzled by the column tables the hardware uses. This works that hierarchy out so the
// bytes land where the GS will later read them — the difference between an upload that
// reads back as the image and one that reads back as noise.
func addrPSMCT32(bp, bw, x, y uint32) uint32 {
	if bw == 0 {
		bw = 1
	}
	const pageW, pageH = 64, 32 // a PSMCT32 page is 64x32 pixels
	pageX := x / pageW
	pageY := y / pageH
	pagesPerRow := bw // bw is already in units of 64 pixels = one page wide
	page := pageY*pagesPerRow + pageX

	px := x % pageW
	py := y % pageH

	// Block within the page: a page is 8x4 blocks of 8x8 pixels, in a fixed order.
	bx := px / 8
	by := py / 8
	block := blockPSMCT32[by][bx]

	// Column within the block: 8x8 pixels, swizzled by the column table.
	cx := px % 8
	cy := py % 8
	col := columnPSMCT32[cy][cx]

	// A page is 8 KiB, a block 256 bytes, a pixel 4 bytes.
	word := bp*64 + page*2048 + uint32(block)*64 + uint32(col)
	return word * 4
}

// blockPSMCT32 is the order of the 8x4 blocks within a PSMCT32 page.
var blockPSMCT32 = [4][8]uint8{
	{0, 1, 4, 5, 16, 17, 20, 21},
	{2, 3, 6, 7, 18, 19, 22, 23},
	{8, 9, 12, 13, 24, 25, 28, 29},
	{10, 11, 14, 15, 26, 27, 30, 31},
}

// columnPSMCT32 is the swizzle of the 8x8 pixels within a PSMCT32 block.
var columnPSMCT32 = [8][8]uint8{
	{0, 1, 4, 5, 8, 9, 12, 13},
	{2, 3, 6, 7, 10, 11, 14, 15},
	{16, 17, 20, 21, 24, 25, 28, 29},
	{18, 19, 22, 23, 26, 27, 30, 31},
	{32, 33, 36, 37, 40, 41, 44, 45},
	{34, 35, 38, 39, 42, 43, 46, 47},
	{48, 49, 52, 53, 56, 57, 60, 61},
	{50, 51, 54, 55, 58, 59, 62, 63},
}

// --- the privileged register block (0x12000000) -----------------------------------
//
// These are the registers the EE writes directly rather than through the GIF: the display
// mode, the frame buffers the CRTC scans out, and CSR — the status register a game polls
// for vertical sync and for the "drawing finished" signal.

const (
	gsPMODE    = 0x12000000
	gsDISPFB1  = 0x12000070
	gsDISPLAY1 = 0x12000080
	gsDISPFB2  = 0x12000090
	gsDISPLAY2 = 0x120000A0
	gsBGCOLOR  = 0x120000E0
	gsCSR      = 0x12001000
	gsIMR      = 0x12001010
)

// gsPrivRead serves a read of the privileged block. CSR is the one that is read: its low
// word carries the signal, finish, HSync and VSync bits and the current field.
func (m *Machine) gsPrivRead(a uint32) (uint32, bool) {
	gs := m.ensureGS()
	switch a &^ 4 {
	case gsCSR:
		if a&4 != 0 {
			return uint32(gs.csr >> 32), true
		}
		return uint32(gs.csr), true
	}
	// Everything else reads back what was written.
	return m.io[a], true
}

// gsPrivWrite serves a write. CSR's signal/finish/vsync bits are write-1-to-clear; the
// rest are stored. Writing bit 9 (RESET) clears the status.
func (m *Machine) gsPrivWrite(a, v uint32) bool {
	gs := m.ensureGS()
	switch a {
	case gsCSR:
		if v&(1<<9) != 0 { // RESET
			gs.csr = 0
		}
		gs.csr &^= uint64(v) & 0x1F // SIGNAL/FINISH/HSINT/VSINT/EDWINT are W1C
	case gsCSR + 4:
	}
	m.io[a] = v
	return true
}

// GSStatus reports what reached the Graphics Synthesizer: uploads and primitives. It is
// the render path's bottom line — a boot that uploads and never draws says the transport
// above the GS is the gap, and one that drives prims into an empty rasteriser says the
// rasteriser is.
func (m *Machine) GSStatus() string {
	if m.gs == nil {
		return "the GS was never touched\n"
	}
	s := sprintf("the GS: %d image uploads, %d vertex kicks\n", m.gs.uploads, m.gs.prims)
	for i, n := range m.gs.primCount {
		if n > 0 {
			s += sprintf("      %-24s %d\n", primNames[i], n)
		}
	}
	for _, kv := range sortedCounts(m.gs.drawCensus) {
		s += sprintf("      %-24s %d\n", kv.name, kv.n)
	}
	return s
}

// GSFrame reads back the frame the CRTC would be scanning out: the rectangle DISPFB
// points at, deswizzled from GS memory into plain RGBA rows. It is the instrument that
// turns "the GS has state" into a picture — the same job the 3DS's framedbg does — and
// it reads the SHIPPED thing: the memory the uploads and the rasteriser actually wrote,
// through the same swizzle the hardware would scan it out with.
//
// DISPFB (one per read circuit; circuit 2 is the one this game drives): FBP in units of
// 2048 words (32 blocks), FBW in units of 64 pixels, PSM in bits 15..19, and the DBX/DBY
// offset in the upper word. DISPLAY gives the magnified screen size; the height here is
// taken from DISPLAY's DH so the dump is the visible frame, not the whole buffer.
func (m *Machine) GSFrame() (pix []byte, w, h int) {
	if m.gs == nil {
		return nil, 0, 0
	}
	// Prefer the circuit PMODE enables; this game uses circuit 2.
	dispfb := uint64(m.io[gsDISPFB2]) | uint64(m.io[gsDISPFB2+4])<<32
	display := uint64(m.io[gsDISPLAY2]) | uint64(m.io[gsDISPLAY2+4])<<32
	if pmode := m.io[gsPMODE]; pmode&2 == 0 && pmode&1 != 0 {
		dispfb = uint64(m.io[gsDISPFB1]) | uint64(m.io[gsDISPFB1+4])<<32
		display = uint64(m.io[gsDISPLAY1]) | uint64(m.io[gsDISPLAY1+4])<<32
	}

	fbp := uint32(dispfb) & 0x1FF       // pages of 2048 words
	fbw := uint32(dispfb>>9) & 0x3F     // width, units of 64 pixels
	psm := uint32(dispfb>>15) & 0x1F    //
	dbx := uint32(dispfb>>32) & 0x7FF   //
	dby := uint32(dispfb>>43) & 0x7FF   //
	magh := uint32(display>>23)&0xF + 1 //
	magv := uint32(display>>27)&0x3 + 1 //
	dw := (uint32(display>>32)&0xFFF + 1) / magh
	dh := (uint32(display>>44)&0x7FF + 1) / magv

	if fbw == 0 || dw == 0 || dh == 0 {
		return nil, 0, 0
	}
	if psm != psmCT32 && psm != psmCT24 {
		m.note("GS: GSFrame asked to read a PSM 0x%X display buffer — only CT32/CT24 read back yet", psm)
		return nil, 0, 0
	}

	w, h = int(dw), int(dh)
	pix = make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			addr := addrPSMCT32(fbp*32, fbw, dbx+uint32(x), dby+uint32(y))
			o := (y*w + x) * 4
			if addr+4 <= uint32(len(m.gs.vram)) {
				pix[o+0] = m.gs.vram[addr+0]
				pix[o+1] = m.gs.vram[addr+1]
				pix[o+2] = m.gs.vram[addr+2]
				pix[o+3] = 0xFF // GS alpha is not display alpha; the dump is opaque
			}
		}
	}
	return pix, w, h
}

// gsVSync toggles the CSR bits the frame clock owns, called once per synthetic vertical
// blank: the VSINT interrupt bit and the FIELD bit that a game watches to pace itself.
func (m *Machine) gsVSync() {
	if m.gs == nil {
		return
	}
	m.gs.csr |= 1 << 3  // VSINT
	m.gs.csr ^= 1 << 13 // FIELD alternates each field
}

// u64bytes returns the little-endian bytes of a 64-bit value, for the HWREG path.
func u64bytes(v uint64) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
		byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56)}
}

// le32 reads a little-endian 32-bit word.
func le32gs(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}
