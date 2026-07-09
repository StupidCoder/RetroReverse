package n64

// rdp_tri.go is the triangle span rasteriser: the bulk of what the RDP does.
//
// The RDP is not handed three vertices. The RSP's microcode has already sorted
// them by height and turned them into *edge coefficients*: three edges, each an
// x position and a slope dx/dy, plus the three y values at which they meet. One
// edge — the major edge — spans the triangle's full height; the other two, the
// minor edges, cover its top and bottom halves and meet at the middle vertex.
//
// Everything the triangle interpolates arrives the same way. Colour, texture
// coordinates and depth are given as a value at the top vertex plus two
// gradients: one along x, and one along the major edge as y advances. So a
// pixel's shade is
//
//	value = base + de*(y - yh) + dx*(x - xMajor(y))
//
// which is why the major edge matters: it is the origin of the coordinate system
// every attribute is expressed in.
//
// Coefficients are s15.16 fixed point, y values are 11.2 — quarter-scanlines.
// The rasteriser here steps whole scanlines from the top of the triangle, which
// loses the subpixel coverage the hardware computes for antialiasing. Edges are
// therefore hard, and a game relying on coverage for transparency will look
// wrong; nothing seen so far does.

// triAttrs holds one interpolated quantity's base value and its two gradients,
// all in s15.16.
type triAttrs struct {
	base, dx, de int64
}

// at evaluates the attribute at a pixel, given the scanline offset from the top
// vertex and the pixel offset from the major edge.
func (a triAttrs) at(dy, dxPix int64) int64 { return a.base + a.de*dy + a.dx*dxPix }

// s32 sign-extends a 32-bit coefficient into the int64 the interpolators use.
func s32(v uint32) int64 { return int64(int32(v)) }

// s14 sign-extends one of the 14-bit y values, which are in 11.2 fixed point.
func s14(v uint32) int32 { return int32(v<<18) >> 18 }

// pair reads a 16.16 coefficient split across an integer word and a fraction
// word, which is how shade, texture and their gradients are packed.
func pair(intWord, fracWord uint64, shift uint) int64 {
	i := int64(int16(uint16(intWord >> shift)))
	f := int64(uint16(fracWord >> shift))
	return i<<16 | f
}

// triangle rasterises one of the eight triangle commands.
func (m *Machine) triangle(op uint32, w []uint64) {
	r := &m.rdp
	if ct := r.cycleType(); ct != cycle1 && ct != cycle2 {
		m.CPU.Halt("unmodelled triangle in cycle type %d", ct)
		return
	}

	hasShade := op&0x04 != 0
	hasTex := op&0x02 != 0
	hasZ := op&0x01 != 0

	lft := w[0]>>55&1 != 0
	tileIdx := uint32(w[0] >> 48 & 7)
	yl := s14(uint32(w[0] >> 32 & 0x3FFF))
	ym := s14(uint32(w[0] >> 16 & 0x3FFF))
	yh := s14(uint32(w[0] & 0x3FFF))

	xl, dxldy := s32(uint32(w[1]>>32)), s32(uint32(w[1]))
	xh, dxhdy := s32(uint32(w[2]>>32)), s32(uint32(w[2]))
	xm, dxmdy := s32(uint32(w[3]>>32)), s32(uint32(w[3]))

	// The optional coefficient blocks follow in a fixed order.
	next := 4
	var sh [4]triAttrs // r, g, b, a
	if hasShade {
		b := w[next : next+8]
		for i := uint(0); i < 4; i++ {
			s := 48 - 16*i
			sh[i] = triAttrs{
				base: pair(b[0], b[2], s),
				dx:   pair(b[1], b[3], s),
				de:   pair(b[4], b[6], s),
			}
		}
		next += 8
	}
	var tex [3]triAttrs // s, t, w
	if hasTex {
		b := w[next : next+8]
		for i := uint(0); i < 3; i++ {
			s := 48 - 16*i
			tex[i] = triAttrs{
				base: pair(b[0], b[2], s),
				dx:   pair(b[1], b[3], s),
				de:   pair(b[4], b[6], s),
			}
		}
		next += 8
	}
	var zz triAttrs
	if hasZ {
		b := w[next : next+2]
		zz = triAttrs{
			base: s32(uint32(b[0] >> 32)),
			dx:   s32(uint32(b[0])),
			de:   s32(uint32(b[1] >> 32)),
		}
		next += 2
	}

	// Clip vertically against the scissor, in whole scanlines.
	yTop := int32(yh) >> 2
	yBot := int32(yl) >> 2
	if lo := int32(r.Scissor.YH >> 2); yTop < lo {
		yTop = lo
	}
	if hi := int32(r.Scissor.YL >> 2); yBot > hi {
		yBot = hi
	}
	yhScan := int32(yh) >> 2
	ymScan := int32(ym) >> 2

	tile := &r.Tiles[tileIdx]
	prim := unpackColor(r.PrimColor)
	env := unpackColor(r.EnvColor)

	for y := yTop; y < yBot; y++ {
		dy := int64(y - yhScan)

		// The major edge spans the whole triangle; the minor edge switches at the
		// middle vertex.
		major := xh + dxhdy*dy
		var minor int64
		if y < ymScan {
			minor = xm + dxmdy*dy
		} else {
			minor = xl + dxldy*int64(y-ymScan)
		}

		xs, xe := major, minor
		if !lft {
			xs, xe = minor, major
		}
		x0 := int32(xs >> 16)
		x1 := int32(xe >> 16)
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		if lo := int32(r.Scissor.XH >> 2); x0 < lo {
			x0 = lo
		}
		if hi := int32(r.Scissor.XL >> 2); x1 > hi {
			x1 = hi
		}
		if hi := int32(r.Color.Width); x1 > hi {
			x1 = hi
		}

		// Attributes are expressed relative to the major edge, so that is where
		// the per-pixel walk starts from.
		majorPix := int64(major >> 16)

		for x := x0; x < x1; x++ {
			dxPix := int64(x) - majorPix

			in := combineInputs{Prim: prim, Env: env, Shade: rgba{R: 255, G: 255, B: 255, A: 255}}
			if hasShade {
				in.Shade = rgba{
					R: clamp8(sh[0].at(dy, dxPix) >> 16),
					G: clamp8(sh[1].at(dy, dxPix) >> 16),
					B: clamp8(sh[2].at(dy, dxPix) >> 16),
					A: clamp8(sh[3].at(dy, dxPix) >> 16),
				}
			}
			if hasTex {
				sv := tex[0].at(dy, dxPix)
				tv := tex[1].at(dy, dxPix)
				wv := tex[2].at(dy, dxPix)

				// The sampler wants a 10.5 coordinate, which without perspective
				// is just the integer part of the 16.16 coefficient.
				sFix, tFix := int32(sv>>16), int32(tv>>16)

				// The perspective divide. The microcode interpolates s/w and t/w
				// linearly across the triangle, because those are the only
				// quantities that vary linearly in screen space, and hands the RDP
				// 1/w alongside them; the RDP multiplies them back out per pixel.
				//
				// The scale is the part that is easy to get wrong. The RDP takes
				// the w coefficient's integer part as a 15-bit fraction with bit 15
				// as its carry, so unity is 0x8000, not 0x10000. Dividing by 0x10000
				// therefore lands every terrain coordinate inside texel zero, and
				// each triangle comes out a single flat colour.
				if r.OtherModes&omPerspTex != 0 && wv > 0 {
					sFix = int32(sv * 0x8000 / wv)
					tFix = int32(tv * 0x8000 / wv)
				}
				texel, ok := m.sample(tile, sFix, tFix)
				if !ok {
					return
				}
				in.Texel0, in.Texel1 = texel, texel
			}

			z := int64(0)
			if hasZ {
				z = zz.at(dy, dxPix)
			}
			m.drawPixel(uint32(x), uint32(y), &in, z, hasZ)
			if m.CPU.Halted {
				return
			}
		}
	}
}

// omPerspTex enables the per-pixel perspective divide.
const omPerspTex = 1 << 51

func clamp8(v int64) uint32 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint32(v)
}

func unpackColor(v uint32) rgba {
	return rgba{v >> 24 & 255, v >> 16 & 255, v >> 8 & 255, v & 255}
}
