package main

// Turning a decoded texture into a glTF-correct image.
//
// f3d.DecodeTexture returns an *image.RGBA whose pixels are the RDP texel's four
// channels placed verbatim — (I, I, I, I) for an intensity texel, (R, G, B, A)
// for a colour one. That is the right thing for the oracle, which compares those
// channels against the running RDP, but image.RGBA is *alpha-premultiplied*, and
// glTF/PNG want *straight* alpha. A stored (I, I, I, I) reads back as the straight
// colour (255, 255, 255) at alpha I — pure white — which is why an intensity-
// textured castle came out white with holes. The raw bytes still hold the texel,
// so the fix is to read them and rebuild a straight-alpha image with the alpha the
// format actually carries.
//
// What alpha a format carries, from the N64 texel definition:
//
//	I4, I8       intensity only. The RDP replicates I into the alpha slot, but
//	             that alpha is a copy of the colour, not coverage: the surface is
//	             opaque. Export alpha = 255.
//	IA4/8/16     intensity and a real alpha. Keep it.
//	RGBA16       one alpha bit. Keep it (0 or 255).
//	RGBA32, CI   full / palette alpha. Keep it.
//
// The alpha then drives the material mode: all-opaque → the default MASK cutout
// passes every texel; a purely on/off alpha → MASK cuts the holes it should; a
// graded alpha (an IA gradient) → BLEND, so it fades instead of hard-cutting.

import (
	"image"

	"retroreverse.com/games/pilotwings-64-n64/extract/uvtx"
)

// exportImage rebuilds a texture as a straight-alpha image for glTF, and reports
// whether it needs alpha blending (a graded alpha) rather than the MASK cutout.
func exportImage(t *uvtx.Texture) (*image.NRGBA, bool) {
	src := t.Image
	b := src.Bounds()
	out := image.NewNRGBA(b)

	// I4 and I8 are intensity-only: their alpha is a copy of the colour, so the
	// surface is opaque and the replicated alpha must not become transparency.
	intensityOnly := t.Fmt == 4

	graded := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// The raw stored bytes are the RDP texel's channels, unpremultiplied.
			i := src.PixOffset(x, y)
			r, g, bl, a := src.Pix[i], src.Pix[i+1], src.Pix[i+2], src.Pix[i+3]
			if intensityOnly {
				a = 255
			} else if a != 0 && a != 255 {
				graded = true
			}
			o := out.PixOffset(x, y)
			out.Pix[o], out.Pix[o+1], out.Pix[o+2], out.Pix[o+3] = r, g, bl, a
		}
	}
	return out, graded
}
