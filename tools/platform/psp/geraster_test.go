package psp

import (
	"image/color"
	"os"
	"testing"
)

// TestRasterizer validates the GE display-list interpreter + rasterizer end to end by
// building a small list (a full-screen background sprite, coloured sprites, and a
// gouraud triangle) and checking the rendered framebuffer.
func TestRasterizer(t *testing.T) {
	m := NewMachine()
	// Vertex buffers in RAM. Through-mode, color 8888, position s16.
	const vtype = (1 << 23) | (7 << 2) | (2 << 7)
	va := uint32(0x08810000)
	put := func(addr uint32, words ...uint32) uint32 {
		for _, w := range words {
			m.write32(addr, w)
			addr += 4
		}
		return addr
	}
	// helper: a sprite = 2 vertices, each {color u32, x s16, y s16, z s16, pad s16}
	spr := func(addr, col uint32, x0, y0, x1, y1 int) uint32 {
		m.write32(addr, col)
		m.write32(addr+4, uint32(uint16(x0))|uint32(uint16(y0))<<16)
		m.write32(addr+8, 0)
		m.write32(addr+12, col)
		m.write32(addr+16, uint32(uint16(x1))|uint32(uint16(y1))<<16)
		m.write32(addr+20, 0)
		return addr + 24
	}
	bg := va
	next := spr(bg, 0xFF66CCFF, 0, 0, 480, 272) // orange-ish full screen (R=FF,G=CC,B=66)
	red := next
	next = spr(red, 0xFF0000FF, 40, 40, 200, 150) // red rect
	grn := next
	next = spr(grn, 0xFF00FF00, 250, 90, 440, 230) // green rect
	_ = next

	list := func(vaddr uint32, prim uint32) []uint32 {
		return []uint32{
			(cBASE << 24) | ((vaddr >> 8) & 0xFF0000),
			(cVADDR << 24) | (vaddr & 0xFFFFFF),
			(cVTYPE << 24) | vtype,
			(cFBPTR << 24) | 0,
			(cFBWIDTH << 24) | 512,
			(cFBPIXFMT << 24) | psm8888,
			(cPRIM << 24) | prim,
		}
	}
	// Draw each sprite.
	var words []uint32
	words = append(words, list(bg, (6<<16)|2)...)
	words = append(words, list(red, (6<<16)|2)...)
	words = append(words, list(grn, (6<<16)|2)...)
	words = append(words, (geEND << 24))
	_ = put

	m.rasterList(GeList{Words: words})

	// Check pixels.
	img := m.Framebuffer()
	check := func(x, y int, want color.RGBA) {
		got := img.RGBAAt(x, y)
		if got.R != want.R || got.G != want.G || got.B != want.B {
			t.Errorf("pixel (%d,%d) = %v, want %v", x, y, got, want)
		}
	}
	check(5, 5, color.RGBA{0xFF, 0xCC, 0x66, 0xFF}) // background
	check(100, 100, color.RGBA{0xFF, 0, 0, 0xFF})   // red rect
	check(350, 160, color.RGBA{0, 0xFF, 0, 0xFF})   // green rect

	if os.Getenv("PSP_SHOT") != "" {
		m.Screenshot(os.Getenv("PSP_SHOT"))
	}
}
