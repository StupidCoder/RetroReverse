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
// into a fresh VRAM image — the state at the end of the boot streams.
func NewVRAM(files ...[]Rect) *VRAM {
	v := &VRAM{}
	for _, rects := range files {
		for _, r := range rects {
			v.Blit(r)
		}
	}
	return v
}

// The race scenery quadrant: VRAM (640,256)–(1024,512), the texture pages
// 0x1A–0x1E. At the end of TEX1's upload stream the game saves this quadrant
// to RAM (GP0 0xC0 read-backs of eight 384×32 rects), lets the later archives
// draw the menu screens over it, and restores the saved image row by row when
// the race starts (paired 0xC0/0xA0 384×1 transfers — the read half banks the
// menu art for the way back).
const (
	quadX, quadY = 640, 256
	quadW, quadH = 384, 256
)

// NewRaceVRAM builds the texture state the race starts from: the boot replay
// with the scenery quadrant holding the city set. Arguments are the five
// archives' rect lists in load order.
func NewRaceVRAM(tex4, tex0, tex1, tex2, tex3 []Rect) *VRAM {
	return NewSceneryVRAMs(tex4, tex0, tex1, tex2, tex3)[0]
}

// NewSceneryVRAMs builds the three race texture states, indexed by scenery
// set (course.go): each is the boot replay with the quadrant as it stood at
// the end of TEX1's, TEX2's and TEX3's stream respectively — the states the
// rotator at 0x800375FC banks in RAM and pages into VRAM by course progress.
func NewSceneryVRAMs(tex4, tex0, tex1, tex2, tex3 []Rect) [3]*VRAM {
	v := NewVRAM(tex4, tex0, tex1)
	snap := func() []uint16 {
		s := make([]uint16, quadW*quadH)
		for y := 0; y < quadH; y++ {
			for x := 0; x < quadW; x++ {
				s[y*quadW+x] = v.At(quadX+x, quadY+y)
			}
		}
		return s
	}
	blit := func(rects []Rect) {
		for _, r := range rects {
			v.Blit(r)
		}
	}
	q0 := snap()
	blit(tex2)
	q1 := snap()
	blit(tex3)

	var out [3]*VRAM
	for set, q := range [][]uint16{q0, q1, nil} {
		c := &VRAM{}
		c.Pix = v.Pix
		if q != nil {
			for y := 0; y < quadH; y++ {
				for x := 0; x < quadW; x++ {
					c.Pix[(quadY+y)*VRAMW+quadX+x] = q[y*quadW+x]
				}
			}
		}
		out[set] = c
	}
	return out
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

// TexWindow is a PSX texture window (GP0 0xE2): its fields are the raw 5-bit
// mask/offset in 8-pixel units, exactly as the register holds them. Sampling
// masks the high bits of each coordinate to the offset, so a sub-region of the
// page repeats — the GPU rule c = (c &^ (mask*8)) | ((offset & mask)*8). A zero
// mask leaves the coordinate untouched.
type TexWindow struct{ MaskX, MaskY, OffX, OffY byte }

func (w TexWindow) apply(u, vv byte) (byte, byte) {
	u = byte((int(u) &^ (int(w.MaskX) * 8)) | ((int(w.OffX) & int(w.MaskX)) * 8))
	vv = byte((int(vv) &^ (int(w.MaskY) * 8)) | ((int(w.OffY) & int(w.MaskY)) * 8))
	return u, vv
}

// Texel resolves one texel the way the rasterizer does: page selects the
// 64×256 texture page base and colour depth, clut the palette strip, (u,v)
// the texel inside the page. The returned value is a 15-bit VRAM colour
// (0 = fully transparent).
//
// Texture-page halfword: bits 0-3 page X (×64), bit 4 page Y (×256),
// bits 7-8 colour depth (0 = 4-bit, 1 = 8-bit, 2 = direct 15-bit).
// CLUT halfword: bits 0-5 palette X (×16), bits 6-14 palette Y.
func (v *VRAM) Texel(page, clut uint16, u, vv byte) uint16 {
	return v.TexelW(page, clut, u, vv, TexWindow{})
}

// TexelW is Texel with a texture window applied to (u, v) first.
func (v *VRAM) TexelW(page, clut uint16, u, vv byte, win TexWindow) uint16 {
	u, vv = win.apply(u, vv)
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
