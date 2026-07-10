package main

import (
	"image"
	"image/color"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/uvtx"
)

// texImage builds a uvtx.Texture whose stored RGBA pixels are the raw RDP texel
// channels (as f3d.DecodeTexture leaves them), for the given RDP format.
func texImage(fmtCode uint32, pix []color.RGBA) *uvtx.Texture {
	img := image.NewRGBA(image.Rect(0, 0, len(pix), 1))
	for x, c := range pix {
		i := img.PixOffset(x, 0)
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	return &uvtx.Texture{Image: img, Fmt: fmtCode}
}

// An I4/I8 texel stores intensity in every channel, alpha included. The export
// must read the intensity as colour and force alpha opaque — otherwise it becomes
// white (the premultiplied reading) and the low half gets cut away by MASK.
func TestIntensityExportsOpaqueGrey(t *testing.T) {
	tx := texImage(4, []color.RGBA{{17, 17, 17, 17}, {136, 136, 136, 136}, {255, 255, 255, 255}})
	out, blend := exportImage(tx, false)
	if blend {
		t.Error("an intensity texture must not be marked for blending")
	}
	for x, wantI := range []uint8{17, 136, 255} {
		c := out.NRGBAAt(x, 0)
		if c.R != wantI || c.G != wantI || c.B != wantI {
			t.Errorf("texel %d: RGB %d,%d,%d, want grey %d", x, c.R, c.G, c.B, wantI)
		}
		if c.A != 255 {
			t.Errorf("texel %d: alpha %d, want 255 (opaque)", x, c.A)
		}
	}
}

// A colour format's alpha is real. A 1-bit alpha (RGBA16) stays a MASK cutout;
// a graded alpha (IA16) asks for BLEND so it fades instead of hard-cutting.
func TestColourAlphaIsKept(t *testing.T) {
	rgba16 := texImage(0, []color.RGBA{{200, 100, 50, 0}, {200, 100, 50, 255}})
	out, blend := exportImage(rgba16, false)
	if blend {
		t.Error("a 1-bit alpha texture must stay MASK, not BLEND")
	}
	if out.NRGBAAt(0, 0).A != 0 || out.NRGBAAt(1, 0).A != 255 {
		t.Error("1-bit alpha not preserved")
	}
	if c := out.NRGBAAt(1, 0); c.R != 200 || c.G != 100 || c.B != 50 {
		t.Errorf("colour not preserved: %v", c)
	}

	graded := texImage(3, []color.RGBA{{128, 128, 128, 96}, {128, 128, 128, 255}})
	if _, blend := exportImage(graded, false); !blend {
		t.Error("a graded alpha (IA) must be marked for blending")
	}
}

// Terrain draws opaque: with opaque set, a colour texture's stray alpha (an RGBA16
// ground tile's one-bit alpha, an IA tile's intensity alpha) must be forced to 255
// so solid ground is not cut into a MASK sieve of black holes and speckles.
func TestOpaqueForcesFullAlpha(t *testing.T) {
	for _, fmtCode := range []uint32{0, 3, 4} { // RGBA16, IA, I
		src := texImage(fmtCode, []color.RGBA{{200, 100, 50, 0}, {80, 90, 70, 128}, {10, 20, 30, 255}})
		out, blend := exportImage(src, true)
		if blend {
			t.Errorf("fmt %d: opaque must never blend", fmtCode)
		}
		for x := 0; x < 3; x++ {
			if a := out.NRGBAAt(x, 0).A; a != 255 {
				t.Errorf("fmt %d texel %d: alpha %d, want 255", fmtCode, x, a)
			}
		}
		// Colour is still preserved.
		if c := out.NRGBAAt(0, 0); c.R != 200 || c.G != 100 || c.B != 50 {
			t.Errorf("fmt %d: colour changed: %v", fmtCode, c)
		}
	}
}
