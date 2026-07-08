package rr

// vram.go rebuilds the texture half of PSX VRAM by replaying every TEX*.TMS
// upload rect into a 1024×512 16-bit buffer — the same state the GPU holds
// after the boot loader has streamed all five archives. Texture sampling
// (texture page + CLUT + UV → texel) mirrors the GPU's addressing exactly, so
// a polygon's texture words resolve here the way they do on hardware.

const (
	VRAMW = 1024
	VRAMH = 512
)

// VRAM is the replayed texture memory.
type VRAM struct {
	Pix [VRAMW * VRAMH]uint16
}

// NewVRAM replays TMS upload rect lists (any number of files, in load order)
// into a fresh VRAM image.
func NewVRAM(files ...[]Rect) *VRAM {
	v := &VRAM{}
	for _, rects := range files {
		for _, r := range rects {
			v.Blit(r)
		}
	}
	return v
}

// Blit stores one upload rect, wrapping coordinates as the GPU does.
func (v *VRAM) Blit(r Rect) {
	i := 0
	for y := 0; y < r.H; y++ {
		for x := 0; x < r.W; x++ {
			v.Pix[((r.Y+y)&(VRAMH-1))*VRAMW+((r.X+x)&(VRAMW-1))] = r.Pix[i]
			i++
		}
	}
}

// At returns the raw 16-bit VRAM word at (x, y).
func (v *VRAM) At(x, y int) uint16 { return v.Pix[(y&(VRAMH-1))*VRAMW+(x&(VRAMW-1))] }

// Texel resolves one texel the way the rasterizer does: page selects the
// 64×256 texture page base and colour depth, clut the palette strip, (u,v)
// the texel inside the page. The returned value is a 15-bit VRAM colour
// (0 = fully transparent).
//
// Texture-page halfword: bits 0-3 page X (×64), bit 4 page Y (×256),
// bits 7-8 colour depth (0 = 4-bit, 1 = 8-bit, 2 = direct 15-bit).
// CLUT halfword: bits 0-5 palette X (×16), bits 6-14 palette Y.
func (v *VRAM) Texel(page, clut uint16, u, vv byte) uint16 {
	px := int(page&0x0F) * 64
	py := int((page>>4)&1) * 256
	depth := int((page >> 7) & 3)
	cx := int(clut&0x3F) * 16
	cy := int((clut >> 6) & 0x1FF)
	switch depth {
	case 0: // 4-bit: each VRAM word holds four texels
		w := v.At(px+int(u)/4, py+int(vv))
		idx := (w >> (4 * (uint(u) & 3))) & 0xF
		return v.At(cx+int(idx), cy)
	case 1: // 8-bit: two texels per word
		w := v.At(px+int(u)/2, py+int(vv))
		idx := (w >> (8 * (uint(u) & 1))) & 0xFF
		return v.At(cx+int(idx), cy)
	default: // 15-bit direct
		return v.At(px+int(u), py+int(vv))
	}
}
