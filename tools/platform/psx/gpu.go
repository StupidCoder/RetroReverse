package psx

// gpu.go emulates the PlayStation GPU: a 1 MiB frame buffer (1024×512 16-bit
// pixels) driven by two ports — GP0 (0x1F801810) for drawing, GP1 (0x1F801814)
// for display control. Drawing commands arrive as a word stream (written
// directly or, mostly, DMA'd from an ordering table by channel 2); each command
// is a byte opcode plus a fixed or image-sized run of parameter words. The GTE
// (tools/mips) has already transformed the geometry to screen coordinates; the
// GPU rasterizes the flat/gouraud/textured triangles, quads and rectangles into
// VRAM. Screenshot renders the active display area to a PNG, modelled on
// tools/dos/dos_video.go.
//
// Colours in VRAM are 16-bit 5:5:5 (bit0-4 R, 5-9 G, 10-14 B, bit15 mask).

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

const (
	vramW = 1024
	vramH = 512
)

type gpu struct {
	vram []uint16 // vramW*vramH

	// GP0 command assembly.
	fifo []uint32 // words collected for the in-progress command
	need int      // words still needed (0 = ready for a new opcode)
	// Image upload (GP0 0xA0): destination rect and remaining 16-bit pixels.
	imgX, imgY, imgW, imgH int
	imgPx                  int // pixels still to receive
	imgCurX, imgCurY       int

	// Drawing state (GP0 0xE1-0xE6).
	drawL, drawT, drawR, drawB int // clip rectangle (inclusive TL, exclusive BR)
	offX, offY                 int // drawing offset added to vertex coords
	texPageX, texPageY         int // texture page base in VRAM
	texDepth                   int // 0=4bpp, 1=8bpp, 2=15bpp
	texWinMX, texWinMY         int
	texWinOX, texWinOY         int

	// GP1 display state.
	dispX, dispY  int  // top-left of the shown area in VRAM
	dispW, dispH  int  // shown resolution
	dispEnabled   bool
	gp0Read       uint32 // last GPUREAD latch (for VRAM→CPU, mostly unused)
	statField     uint32 // toggles so GPUSTAT polls terminate

	// DEBUG: watch VRAM rect [dwx0,dwy0)-(dwx1,dwy1) and log who writes it.
	dbgOn                  bool
	dwx0, dwy0, dwx1, dwy1 int
	dbgLog                 func(string)
	dbgWatchedLoad         bool
	dbgFirstWord           bool
	dbgAllOnes             bool
}

// DEBUG helpers for the who-writes-VRAM investigation.
func (g *gpu) dhits(x, y, w, h int) bool {
	if !g.dbgOn {
		return false
	}
	return x < g.dwx1 && x+w > g.dwx0 && y < g.dwy1 && y+h > g.dwy0
}

func (g *gpu) dvariety(x, y, w, h int) int {
	seen := map[uint16]bool{}
	for yy := y; yy < y+h && yy < vramH; yy++ {
		for xx := x; xx < x+w && xx < vramW; xx++ {
			seen[g.vram[yy*vramW+xx]] = true
			if len(seen) > 64 {
				return len(seen)
			}
		}
	}
	return len(seen)
}

func newGPU() *gpu {
	g := &gpu{vram: make([]uint16, vramW*vramH)}
	g.drawR, g.drawB = vramW, vramH
	g.dispW, g.dispH = 320, 240
	g.dispEnabled = true
	return g
}

// --- port interface --------------------------------------------------------

func (g *gpu) status() uint32 {
	// Report the "ready" bits (26 cmd, 27 VRAM-send, 28 DMA) plus a toggling
	// field so busy-wait polls terminate; bit 31 is toggled per read too.
	g.statField ^= 0x80000000
	return 0x1C000000 | g.statField
}

func (g *gpu) read() uint32 { return g.gp0Read }

// gp1 handles a display/control command (top byte selects the command).
func (g *gpu) gp1(w uint32) {
	switch w >> 24 {
	case 0x00: // reset
		g.need, g.fifo, g.imgPx = 0, nil, 0
		g.dispEnabled = false
	case 0x01: // reset command buffer
		g.need, g.fifo, g.imgPx = 0, nil, 0
	case 0x03: // display enable (bit0: 0=on)
		g.dispEnabled = w&1 == 0
	case 0x05: // display area start in VRAM
		g.dispX = int(w & 0x3FF)
		g.dispY = int((w >> 10) & 0x1FF)
	case 0x08: // display mode: derive width/height
		switch w & 3 {
		case 0:
			g.dispW = 256
		case 1:
			g.dispW = 320
		case 2:
			g.dispW = 512
		case 3:
			g.dispW = 640
		}
		if w&0x04 != 0 {
			g.dispW = 368
		}
		if w&0x20 != 0 { // vertical interlace / 480
			g.dispH = 480
		} else {
			g.dispH = 240
		}
	}
}

// gp0 accepts one drawing-port word: image data, then either a new opcode or a
// continuation of the current command.
func (g *gpu) gp0(w uint32) {
	if g.imgPx > 0 {
		g.pushImage(w)
		return
	}
	if g.need == 0 {
		g.need = gp0Len(byte(w >> 24))
		g.fifo = g.fifo[:0]
	}
	g.fifo = append(g.fifo, w)
	g.need--
	if g.need == 0 {
		g.exec()
	}
}

// feedNode sends one GPU-DMA ordering-table node's body words to GP0. Each node
// in Ridge Racer's ordering table holds exactly one primitive whose first word
// is the GP0 opcode+colour. The game keeps disabled primitives (e.g. culled
// roadside sprites, whole track-object nodes that are off this frame) linked in
// the table but zeroes their opcode+colour word, so the leading opcode reads as
// a NOP (0x00). Feeding such a node's remaining words straight to GP0 would
// misread its texcoord/CLUT words — whose high bytes hold values like 0x7B or
// 0x3C — as rectangle/polygon opcodes, drawing spurious primitives and, when a
// bogus opcode's parameter count overruns the node, desyncing the FIFO into
// giant fills that trample the texture atlas. Skip a node whose primitive opcode
// is a NOP; feed every other node forward unchanged.
func (g *gpu) feedNode(w []uint32) {
	if g.need == 0 && g.imgPx == 0 && len(w) >= 2 && byte(w[0]>>24) == 0 {
		return
	}
	for _, x := range w {
		g.gp0(x)
	}
}

// pushImage stores two 16-bit pixels from a CPU→VRAM word into the upload rect.
func (g *gpu) pushImage(w uint32) {
	if g.dbgWatchedLoad {
		if g.dbgFirstWord {
			g.dbgFirstWord, g.dbgAllOnes = false, true
		}
		if w != 0 {
			g.dbgAllOnes = false
		}
	}
	for _, px := range [2]uint16{uint16(w), uint16(w >> 16)} {
		if g.imgPx <= 0 {
			return
		}
		x := (g.imgX + g.imgCurX) & (vramW - 1)
		y := (g.imgY + g.imgCurY) & (vramH - 1)
		g.vram[y*vramW+x] = px
		g.imgPx--
		if g.imgCurX++; g.imgCurX >= g.imgW {
			g.imgCurX = 0
			g.imgCurY++
		}
		if g.imgPx == 0 && g.dbgWatchedLoad {
			g.dbgLog(fmt.Sprintf("IMAGELOAD done dst=(%d,%d) %dx%d allZeroData=%v vramVar=%d",
				g.imgX, g.imgY, g.imgW, g.imgH, g.dbgAllOnes,
				g.dvariety(g.imgX, g.imgY, g.imgW, g.imgH)))
			g.dbgWatchedLoad = false
		}
	}
}

// --- command length --------------------------------------------------------

// gp0Len returns the total word count for a GP0 command opcode (including the
// opcode word). Image-load commands report their header length; the pixel run is
// handled separately.
func gp0Len(op byte) int {
	switch {
	case op == 0x02: // fill rectangle
		return 3
	case op >= 0x20 && op <= 0x3F: // polygon
		verts := 3
		if op&0x08 != 0 {
			verts = 4
		}
		gouraud := op&0x10 != 0
		textured := op&0x04 != 0
		n := verts     // one xy word per vertex
		if gouraud {   // a colour word per vertex (first shares the opcode word)
			n += verts
		} else {
			n++ // single flat colour (the opcode word)
		}
		if textured {
			n += verts
		}
		return n
	case op >= 0x40 && op <= 0x5F: // line / poly-line (treat as 3-word line)
		if op&0x08 != 0 {
			return 4 // fixed; poly-lines end on 0x55555555 but are rare here
		}
		return 3
	case op >= 0x60 && op <= 0x7F: // rectangle
		n := 2 // colour + xy
		if op&0x04 != 0 {
			n++ // texcoord+clut
		}
		if op&0x18 == 0 {
			n++ // variable size: width/height word
		}
		return n
	case op >= 0x80 && op <= 0x9F: // VRAM→VRAM copy
		return 4
	case op >= 0xA0 && op <= 0xBF: // CPU→VRAM copy (header)
		return 3
	case op >= 0xC0 && op <= 0xDF: // VRAM→CPU copy (header)
		return 3
	}
	return 1 // NOP, cache clear, draw-mode settings (0xE1-0xE6), etc.
}

// --- command execution -----------------------------------------------------

func (g *gpu) exec() {
	op := byte(g.fifo[0] >> 24)
	switch {
	case op == 0x02:
		g.fillRect()
	case op >= 0x20 && op <= 0x3F:
		g.polygon(op)
	case op >= 0x60 && op <= 0x7F:
		g.rect(op)
	case op >= 0x80 && op <= 0x9F:
		g.copyVRAM()
	case op >= 0xA0 && op <= 0xBF:
		g.beginImageLoad()
	case op >= 0xE1 && op <= 0xE6:
		g.drawSetting(op)
	}
	g.need = 0
}

func (g *gpu) drawSetting(op byte) {
	w := g.fifo[0]
	switch op {
	case 0xE1: // texpage / draw mode
		g.texPageX = int(w&0x0F) * 64
		g.texPageY = int((w>>4)&1) * 256
		g.texDepth = int((w >> 7) & 3)
	case 0xE2: // texture window
		g.texWinMX = int(w & 0x1F)
		g.texWinMY = int((w >> 5) & 0x1F)
		g.texWinOX = int((w >> 10) & 0x1F)
		g.texWinOY = int((w >> 15) & 0x1F)
	case 0xE3: // drawing area top-left
		g.drawL = int(w & 0x3FF)
		g.drawT = int((w >> 10) & 0x3FF)
	case 0xE4: // drawing area bottom-right
		g.drawR = int(w&0x3FF) + 1
		g.drawB = int((w>>10)&0x3FF) + 1
	case 0xE5: // drawing offset (signed 11-bit)
		g.offX = sext11(w & 0x7FF)
		g.offY = sext11((w >> 11) & 0x7FF)
	}
}

func (g *gpu) beginImageLoad() {
	dst, size := g.fifo[1], g.fifo[2]
	g.imgX = int(dst & 0x3FF)
	g.imgY = int((dst >> 16) & 0x1FF)
	g.imgW = int(size & 0xFFFF)
	g.imgH = int((size >> 16) & 0xFFFF)
	if g.imgW == 0 {
		g.imgW = 1024
	}
	if g.imgH == 0 {
		g.imgH = 512
	}
	g.imgCurX, g.imgCurY = 0, 0
	g.imgPx = g.imgW * g.imgH
	g.dbgWatchedLoad = g.dhits(g.imgX, g.imgY, g.imgW, g.imgH)
	g.dbgFirstWord = true
}

func (g *gpu) fillRect() {
	c := color15(g.fifo[0])
	x0 := int(g.fifo[1] & 0x3FF)
	y0 := int((g.fifo[1] >> 16) & 0x1FF)
	w := int(g.fifo[2] & 0x3FF)
	h := int((g.fifo[2] >> 16) & 0x1FF)
	if g.dhits(x0, y0, w, h) {
		g.dbgLog(fmt.Sprintf("FILL dst=(%d,%d) %dx%d color=0x%04X", x0, y0, w, h, c))
	}
	for y := y0; y < y0+h && y < vramH; y++ {
		for x := x0; x < x0+w && x < vramW; x++ {
			g.vram[y*vramW+x] = c
		}
	}
}

func (g *gpu) copyVRAM() {
	src, dst, size := g.fifo[1], g.fifo[2], g.fifo[3]
	sx, sy := int(src&0x3FF), int((src>>16)&0x1FF)
	dx, dy := int(dst&0x3FF), int((dst>>16)&0x1FF)
	w, h := int(size&0x3FF), int((size>>16)&0x1FF)
	if g.dhits(dx, dy, w, h) {
		g.dbgLog(fmt.Sprintf("COPY src=(%d,%d) dst=(%d,%d) %dx%d srcVar=%d",
			sx, sy, dx, dy, w, h, g.dvariety(sx, sy, w, h)))
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			g.vram[((dy+y)&(vramH-1))*vramW+((dx+x)&(vramW-1))] =
				g.vram[((sy+y)&(vramH-1))*vramW+((sx+x)&(vramW-1))]
		}
	}
}

// --- Screenshot ------------------------------------------------------------

// Screenshot renders the active display area (dispW×dispH from the VRAM
// display start) to a 24-bit PNG.
func (g *gpu) Screenshot(path string) error {
	w, h := g.dispW, g.dispH
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := g.vram[((g.dispY+y)&(vramH-1))*vramW+((g.dispX+x)&(vramW-1))]
			r, gr, b := rgb15(px)
			img.Set(x, y, color.RGBA{r, gr, b, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// DumpVRAM (debug) writes the entire 1024×512 VRAM as a 15-bit-colour PNG.
func (g *gpu) DumpVRAM(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, vramW, vramH))
	for y := 0; y < vramH; y++ {
		for x := 0; x < vramW; x++ {
			r, gr, b := rgb15(g.vram[y*vramW+x])
			img.Set(x, y, color.RGBA{r, gr, b, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// --- helpers ---------------------------------------------------------------

func sext11(v uint32) int {
	if v&0x400 != 0 {
		return int(v) - 0x800
	}
	return int(v)
}

// color15 converts a 24-bit GP0 colour word to 15-bit 5:5:5.
func color15(w uint32) uint16 {
	r := (w & 0xFF) >> 3
	g := ((w >> 8) & 0xFF) >> 3
	b := ((w >> 16) & 0xFF) >> 3
	return uint16(r | g<<5 | b<<10)
}

// rgb15 expands a 15-bit VRAM pixel to 8-bit RGB.
func rgb15(px uint16) (r, g, b byte) {
	r = byte((px & 0x1F) << 3)
	g = byte(((px >> 5) & 0x1F) << 3)
	b = byte(((px >> 10) & 0x1F) << 3)
	return
}
