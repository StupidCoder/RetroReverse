package merc

// eyes.go: the eye compositor, reimplemented from render-eyes (0x902BA4) and
// convert-eye-data (0x9021D4).
//
// The game never ships eye textures for the models: each merc eye effect
// binds a 4x4 `programmer_eye_left`/`_right` placeholder on disc, and at
// runtime `update-eyes` (0x90225C) composites a 32x32 CT32 slot per eye into
// VRAM (page 0xFF, *eyes-base-block* 0x1FE0 + eye*0x10; pair idx = the
// eye-control-array index, band y = idx*32, left eye x 0..31, right 32..63)
// which the model then samples. render-eyes draws, per pair, under
// XYOFFSET (32,32) with CLAMP forced to {WMS=WMT=CLAMP, MAXU=MAXV=31}:
//
//  1. background: one sprite across both slots with both UVs pinned to (0,0)
//     — a flat fill with the IRIS shader's corner texel (white), RGBA reg
//     128 (x1.0 modulate). Template A: sprite, TME, FST, ABE=0.
//  2. iris left/right: sprite centered at raw 16*(48|80 + 32*ex) x
//     16*(pair*32+48 + 32*ey), half-size 256*size (12.4), full 0..W UV span,
//     ABE=0, scissored to its slot.
//  3. TEST_1 = 0x33001 (ATST=NEVER, AFAIL=RGB_ONLY): the pupil pass writes
//     RGB only — the slot's alpha (the envmap shine mask) keeps the
//     background/iris coverage. Pupil sprites: same center, half-size
//     256*psize, template B (ABE=1), ALPHA=0x44 source-over by texel alpha.
//  4. TEST_1 restored (0x30003); third shader with ALPHA=1
//     (Cv=(Cd-Cs)*As+Cs) and RGBA reg alpha 0 — As = texA*0 = 0, so the lid
//     draws its texture RGB opaque and zeroes the mask. Lid quad: window y
//     from 16*(pair*32 + 32*state) spanning 32*height px (the slot band is
//     entered once state > 0), full slot width, drawn right-to-left
//     (mirrored) for the right eye. State <= 0 is the mid-blink encoding
//     (render-eyes folds it toward 1.0 by ctl+12); at state 0 the quad sits
//     entirely above the band — fully open, nothing drawn.
//
// The textures are the game's own runtime bindings (title-logo.state
// eye-control adgifs, TBP+CBP both reconciled against the tpage
// descriptors): iris = bam-iris-16x16 (tpage-62), pupil = autoeye-pupil
// (tpage-463), lid = autoeye-lid (tpage-463); the sidekick overrides the lid
// with sk-eye-lid (tpage-1532).
//
// The raster/sampler arithmetic below mirrors tools/platform/ps2 (gsdraw.go
// sprite + gstex.go pick/combine/blendPixel), which the live captures were
// rendered by: half-open [x0>>4, x1>>4) coverage, UV interpolated at the
// INTEGER pixel coordinate in 12.4 via truncating int64 division, LINEAR
// filter at (u-8, v-8) with 4-bit weights and per-stage truncating lerps.
// Verified pixel-exact against title-logo.state (cmd/eyeprobe): slot 3 raw
// -gsfb RGB and -gstex alpha both diff zero.

import (
	"image"
	"image/color"
)

// EyeParams is one eye's control block, the convert-eye-data floats: 4
// signed bytes /128 {iris x, iris y, lid state, ?} and 4 unsigned bytes /64
// {iris size, pupil size, lid height, ?}.
type EyeParams struct {
	X, Y      float32 // iris offset from slot center, 32 px per 1.0
	LidState  float32 // 0 = open (quad above the slot), 1 = fully drawn down
	IrisSize  float32 // sprite half-size, 16 px per 1.0
	PupilSize float32
	LidHeight float32
}

// gsTex is a texture with non-premultiplied RGB and raw GS alpha (0x80 =
// 1.0) — feed it tpage.DecodeGS output.
type gsTex struct {
	w, h int
	px   [][4]uint8
}

func newGSTex(img image.Image) *gsTex {
	b := img.Bounds()
	t := &gsTex{w: b.Dx(), h: b.Dy(), px: make([][4]uint8, b.Dx()*b.Dy())}
	raw, ok := img.(*image.RGBA)
	for y := 0; y < t.h; y++ {
		for x := 0; x < t.w; x++ {
			if ok { // tpage images carry non-premultiplied bytes; read raw
				c := raw.RGBAAt(b.Min.X+x, b.Min.Y+y)
				t.px[y*t.w+x] = [4]uint8{c.R, c.G, c.B, c.A}
			} else {
				c := color.NRGBAModel.Convert(img.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
				t.px[y*t.w+x] = [4]uint8{c.R, c.G, c.B, c.A}
			}
		}
	}
	return t
}

// pick samples at a 12.4 texel-space coordinate: GS LINEAR magnification,
// CLAMP addressing, 4-bit weight fractions, truncating per-stage lerps.
func (t *gsTex) pick(u, v int32) [4]int32 {
	us, vs := u-8, v-8
	x0, y0 := us>>4, vs>>4
	fx, fy := us&15, vs&15
	cl := func(x int32, hi int) int32 {
		if x < 0 {
			return 0
		}
		if int(x) > hi {
			return int32(hi)
		}
		return x
	}
	at := func(x, y int32) [4]uint8 {
		return t.px[int(cl(y, t.h-1))*t.w+int(cl(x, t.w-1))]
	}
	c00, c10 := at(x0, y0), at(x0+1, y0)
	c01, c11 := at(x0, y0+1), at(x0+1, y0+1)
	var out [4]int32
	for c := 0; c < 4; c++ {
		top := int32(c00[c]) + (int32(c10[c])-int32(c00[c]))*fx>>4
		bot := int32(c01[c]) + (int32(c11[c])-int32(c01[c]))*fx>>4
		out[c] = top + (bot-top)*fy>>4
	}
	return out
}

// texAxis interpolates one UV axis at 12.4 pixel coordinate pc on the line
// (p0,uv0)-(p1,uv1) — the reference rasteriser's truncating int64 form.
func texAxis(pc, p0, p1, uv0, uv1 int32) int32 {
	if p1 == p0 {
		return uv0
	}
	return uv0 + int32(int64(pc-p0)*int64(uv1-uv0)/int64(p1-p0))
}

// EyeSlot is a 32x32 CT32 render target with raw GS alpha.
type EyeSlot struct{ px [32 * 32][4]uint8 }

// sprite rasterizes one GS sprite with 12.4 slot-local corner coordinates
// and 12.4 UV corners, half-open truncated coverage, UVs at integer pixel
// coordinates. mode: 0 = replace RGBA (ABE=0), 1 = source-over RGB only
// (the pupil pass), 2 = opaque RGB + zero alpha (the lid pass, As=0).
func (s *EyeSlot) sprite(t *gsTex, ax, ay, bx, by, au, av, bu, bv int32, mode int) {
	x0, x1 := ax>>4, bx>>4
	y0, y1 := ay>>4, by>>4
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y < y1; y++ {
		if y < 0 || y > 31 {
			continue
		}
		v := texAxis(y<<4, ay, by, av, bv)
		for x := x0; x < x1; x++ {
			if x < 0 || x > 31 {
				continue
			}
			u := texAxis(x<<4, ax, bx, au, bu)
			c := t.pick(u, v)
			d := &s.px[y*32+x]
			switch mode {
			case 0: // replace; modulate x RGBA 128 is identity
				*d = [4]uint8{clamp8(c[0]), clamp8(c[1]), clamp8(c[2]), clamp8(c[3])}
			case 1: // Cv = (Cs-Cd)*As>>7 + Cd, alpha masked (AFAIL RGB_ONLY)
				a := c[3]
				d[0] = clamp8((c[0]-int32(d[0]))*a>>7 + int32(d[0]))
				d[1] = clamp8((c[1]-int32(d[1]))*a>>7 + int32(d[1]))
				d[2] = clamp8((c[2]-int32(d[2]))*a>>7 + int32(d[2]))
			case 2: // lid: As = texA*0 = 0 -> Cs; alpha <- As = 0
				*d = [4]uint8{clamp8(c[0]), clamp8(c[1]), clamp8(c[2]), 0}
			}
		}
	}
}

func clamp8(v int32) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// trunc16 is the EE's cvt.w.s of a float32 coordinate expression (truncate
// toward zero), as render-eyes converts every sprite corner.
func trunc16(f float32) int32 { return int32(f) }

// CompositeEye renders one eye slot. right selects the right-eye column
// (its own iris center formula and a mirrored lid). Alpha is raw GS.
func CompositeEye(iris, pupil, lid image.Image, p EyeParams, right bool) *EyeSlot {
	it, pt, lt := newGSTex(iris), newGSTex(pupil), newGSTex(lid)
	s := &EyeSlot{}

	// background: flat fill sampling UV (0,0) of the iris shader
	c := it.pick(0, 0)
	bg := [4]uint8{clamp8(c[0]), clamp8(c[1]), clamp8(c[2]), clamp8(c[3])}
	for i := range s.px {
		s.px[i] = bg
	}

	// raw window coordinates as render-eyes computes them (float32, 12.4
	// via cvt.w.s), then into slot space: minus XYOFFSET 512 and the
	// right column's 512. The y band offset (pair*512) is exact and
	// cancels; base 48 covers both.
	basex := float32(48)
	col := int32(0)
	if right {
		basex = 80
		col = 512
	}
	cx := 16 * (basex + 32*p.X)
	cy := 16 * (48 + 32*p.Y)
	uw, vh := int32(0), int32(0)

	if r := 256 * p.IrisSize; r > 0 {
		uw, vh = int32(it.w)<<4, int32(it.h)<<4
		s.sprite(it,
			trunc16(cx-r)-512-col, trunc16(cy-r)-512,
			trunc16(cx+r)-512-col, trunc16(cy+r)-512,
			0, 0, uw, vh, 0)
	}
	if r := 256 * p.PupilSize; r > 0 {
		uw, vh = int32(pt.w)<<4, int32(pt.h)<<4
		s.sprite(pt,
			trunc16(cx-r)-512-col, trunc16(cy-r)-512,
			trunc16(cx+r)-512-col, trunc16(cy+r)-512,
			0, 0, uw, vh, 1)
	}
	if p.LidState > 0 {
		// window y0 = 16*(pair*32 + 32*state); slot-local: -512 relative
		// to the band top (pair*512) then the +512 XYOFFSET shift is
		// already folded into the band position — the quad enters the
		// slot once state > 0.
		y0 := trunc16(16*32*p.LidState) - 512
		y1 := trunc16(16*(32*p.LidState+32*p.LidHeight)) - 512
		uw, vh = int32(lt.w)<<4, int32(lt.h)<<4
		lx0, lx1 := int32(0), int32(32<<4)
		if right {
			lx0, lx1 = 32<<4, 0
		}
		s.sprite(lt, lx0, y0, lx1, y1, 0, 0, uw, vh, 2)
	}
	return s
}

// Image returns the slot with GS alpha rescaled to 8-bit (0x80 -> 255).
func (s *EyeSlot) Image() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			p := s.px[y*32+x]
			a := int(p[3]) * 2
			if a > 255 {
				a = 255
			}
			img.SetNRGBA(x, y, color.NRGBA{p[0], p[1], p[2], uint8(a)})
		}
	}
	return img
}

// RawImage returns the slot with raw GS alpha (for diffing against -gstex).
func (s *EyeSlot) RawImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			p := s.px[y*32+x]
			img.SetNRGBA(x, y, color.NRGBA{p[0], p[1], p[2], p[3]})
		}
	}
	return img
}
