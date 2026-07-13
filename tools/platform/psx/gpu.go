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
	// Image store (GP0 0xC0, VRAM→CPU): source rect drained by GPUREAD / to-RAM DMA.
	rdX, rdY, rdW, rdH int
	rdPx               int // pixels still to send
	rdCurX, rdCurY     int

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
	gp0Read       uint32 // last GPUREAD latch, plus VRAM→CPU copy readout
	statField     uint32 // toggles so GPUSTAT polls terminate

	// onPrim, when set, observes each completed GP0 command before it executes
	// (see Machine.OnGP0). The slice is g.fifo — valid only during the call.
	onPrim func(words []uint32)

	// onPixel, when set, observes every pixel the GPU stores to VRAM (see
	// Machine.OnPixel). It is what lets a debugger attribute a pixel to the command
	// that drew it, so every store goes through g.store rather than writing g.vram
	// directly — a store that bypassed it would read as "nothing drew this".
	onPixel func(x, y int, px uint16)

	// stopAt, when non-zero, is the number of GP0 commands to execute before
	// unwinding out of the run (see Machine.RunStopAfterGP0Command). count is how
	// many have run.
	stopAt, count int
}

// store writes one pixel to VRAM. Every path that draws goes through here.
func (g *gpu) store(x, y int, px uint16) {
	g.vram[y*vramW+x] = px
	if g.onPixel != nil {
		g.onPixel(x, y, px)
	}
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

// read returns a GPUREAD word. While a VRAM→CPU copy (GP0 0xC0) is draining, each
// read yields the next two VRAM pixels; otherwise it returns the last latch.
func (g *gpu) read() uint32 {
	if g.rdPx > 0 {
		return g.readWord()
	}
	return g.gp0Read
}

// readWord pops the next two 16-bit VRAM pixels of the active VRAM→CPU copy rect
// (low pixel in the low halfword) and advances; returns 0 once exhausted.
func (g *gpu) readWord() uint32 {
	var w uint32
	for i := uint(0); i < 2; i++ {
		if g.rdPx <= 0 {
			break
		}
		x := (g.rdX + g.rdCurX) & (vramW - 1)
		y := (g.rdY + g.rdCurY) & (vramH - 1)
		w |= uint32(g.vram[y*vramW+x]) << (16 * i)
		g.rdPx--
		if g.rdCurX++; g.rdCurX >= g.rdW {
			g.rdCurX = 0
			g.rdCurY++
		}
	}
	g.gp0Read = w
	return w
}

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
	for _, px := range [2]uint16{uint16(w), uint16(w >> 16)} {
		if g.imgPx <= 0 {
			return
		}
		x := (g.imgX + g.imgCurX) & (vramW - 1)
		y := (g.imgY + g.imgCurY) & (vramH - 1)
		g.store(x, y, px)
		g.imgPx--
		if g.imgCurX++; g.imgCurX >= g.imgW {
			g.imgCurX = 0
			g.imgCurY++
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
	if g.onPrim != nil {
		g.onPrim(g.fifo)
	}
	defer g.afterExec()
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
	case op >= 0xC0 && op <= 0xDF:
		g.beginImageStore()
	case op >= 0xE1 && op <= 0xE6:
		g.drawSetting(op)
	}
	g.need = 0
}

// GP0Name names a GP0 command by its opcode byte, for a debugger's command list. The
// PSX encodes a primitive's shape in the opcode's bits rather than in a table, so the
// name is built from them: the family, then shaded/textured/semi-transparent.
func GP0Name(op uint32) string {
	o := byte(op)
	switch {
	case o == 0x00:
		return "NOP"
	case o == 0x01:
		return "Clear_Cache"
	case o == 0x02:
		return "Fill_Rect"
	case o >= 0x20 && o <= 0x3F:
		return "Polygon" + polyFlags(o)
	case o >= 0x40 && o <= 0x5F:
		return "Line" + lineFlags(o)
	case o >= 0x60 && o <= 0x7F:
		return "Rect" + rectFlags(o)
	case o >= 0x80 && o <= 0x9F:
		return "Copy_VRAM_to_VRAM"
	case o >= 0xA0 && o <= 0xBF:
		return "Copy_CPU_to_VRAM"
	case o >= 0xC0 && o <= 0xDF:
		return "Copy_VRAM_to_CPU"
	case o == 0xE1:
		return "Draw_Mode"
	case o == 0xE2:
		return "Texture_Window"
	case o == 0xE3:
		return "Draw_Area_TL"
	case o == 0xE4:
		return "Draw_Area_BR"
	case o == 0xE5:
		return "Draw_Offset"
	case o == 0xE6:
		return "Mask_Bits"
	}
	return fmt.Sprintf("GP0_%02X", o)
}

func polyFlags(o byte) string {
	s := "_"
	if o&0x08 != 0 {
		s += "Quad"
	} else {
		s += "Tri"
	}
	if o&0x10 != 0 {
		s += "_Gouraud"
	}
	if o&0x04 != 0 {
		s += "_Textured"
	}
	if o&0x02 != 0 {
		s += "_SemiTrans"
	}
	return s
}

func lineFlags(o byte) string {
	s := ""
	if o&0x08 != 0 {
		s += "_Poly"
	}
	if o&0x10 != 0 {
		s += "_Gouraud"
	}
	return s
}

func rectFlags(o byte) string {
	s := ""
	switch (o >> 3) & 3 {
	case 1:
		s += "_1x1"
	case 2:
		s += "_8x8"
	case 3:
		s += "_16x16"
	}
	if o&0x04 != 0 {
		s += "_Textured"
	}
	if o&0x02 != 0 {
		s += "_SemiTrans"
	}
	return s
}

// gp0Stop is the sentinel RunStopAfterGP0Command panics with, to unwind out of the
// GPU write / DMA burst that is executing commands once the one it was told to stop
// after has run. A DMA burst executes many GP0 commands inside a single CPU
// instruction, so returning normally is not enough: the burst would keep going.
// Only RunStopAfterGP0Command recovers it; any other panic passes through.
type gp0Stop struct{}

// afterExec counts an executed command and unwinds if this was the one we were asked
// to stop after. The command has already run, and the GPU is left partway through
// whatever queue was feeding it — which is only sound on a machine that will be
// thrown away (a replay scratch).
func (g *gpu) afterExec() {
	if g.stopAt == 0 {
		return
	}
	g.count++
	if g.count >= g.stopAt {
		panic(gp0Stop{})
	}
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
}

// beginImageStore sets up a VRAM→CPU copy (GP0 0xC0): the source rect is later
// drained by GPUREAD reads or a to-RAM (VRAM→CPU) GPU DMA — the game uses this to
// pull staged textures back out of VRAM into a RAM work buffer.
func (g *gpu) beginImageStore() {
	src, size := g.fifo[1], g.fifo[2]
	g.rdX = int(src & 0x3FF)
	g.rdY = int((src >> 16) & 0x1FF)
	g.rdW = int(size & 0xFFFF)
	g.rdH = int((size >> 16) & 0xFFFF)
	if g.rdW == 0 {
		g.rdW = 1024
	}
	if g.rdH == 0 {
		g.rdH = 512
	}
	g.rdCurX, g.rdCurY = 0, 0
	g.rdPx = g.rdW * g.rdH
}

func (g *gpu) fillRect() {
	c := color15(g.fifo[0])
	x0 := int(g.fifo[1] & 0x3FF)
	y0 := int((g.fifo[1] >> 16) & 0x1FF)
	w := int(g.fifo[2] & 0x3FF)
	h := int((g.fifo[2] >> 16) & 0x1FF)
	for y := y0; y < y0+h && y < vramH; y++ {
		for x := x0; x < x0+w && x < vramW; x++ {
			g.store(x, y, c)
		}
	}
}

func (g *gpu) copyVRAM() {
	src, dst, size := g.fifo[1], g.fifo[2], g.fifo[3]
	sx, sy := int(src&0x3FF), int((src>>16)&0x1FF)
	dx, dy := int(dst&0x3FF), int((dst>>16)&0x1FF)
	w, h := int(size&0x3FF), int((size>>16)&0x1FF)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			g.store((dx+x)&(vramW-1), (dy+y)&(vramH-1),
				g.vram[((sy+y)&(vramH-1))*vramW+((sx+x)&(vramW-1))])
		}
	}
}

// --- display readout -------------------------------------------------------

// readRect renders a w×h rectangle of VRAM at (x0, y0) as an RGBA image.
func (g *gpu) readRect(x0, y0, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := g.vram[((y0+y)&(vramH-1))*vramW+((x0+x)&(vramW-1))]
			r, gr, b := rgb15(px)
			img.Set(x, y, color.RGBA{r, gr, b, 255})
		}
	}
	return img
}

// image renders the active display area — the picture the console is scanning out.
func (g *gpu) image() *image.RGBA {
	return g.readRect(g.dispX, g.dispY, g.dispW, g.dispH)
}

// Screenshot writes the active display area to a 24-bit PNG.
func (g *gpu) Screenshot(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, g.image())
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
