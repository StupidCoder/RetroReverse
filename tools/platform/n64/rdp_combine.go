package n64

// rdp_combine.go is the colour combiner and the blender — the two programmable
// stages every pixel drawn in 1-cycle or 2-cycle mode passes through.
//
// The combiner computes (a - b) * c + d, once per cycle, on colour and alpha
// separately. What makes it a combiner rather than an equation is that each of
// a, b, c and d is a multiplexer: Set_Combine_Mode packs eight of them into one
// 56-bit word, choosing between the texel, the interpolated shade, the primitive
// and environment constants, the previous cycle's result, and a few others. A
// game changes its whole shading model by writing one register.
//
// The blender then merges that colour with what is already in the framebuffer,
// through four more multiplexers. Together they are how the N64 draws anything
// that is not a flat fill.
//
// Both stages work in 0..255 here, where the hardware carries extra guard bits
// internally. The rounding below matches the hardware's ((a-b)*c + 0x80) >> 8,
// which is what keeps a texture modulated by a shade of 255 exactly itself.

// combineInputs is everything the multiplexers can select from, for one pixel.
type combineInputs struct {
	Texel0      rgba // the sampled texel
	Texel1      rgba // the second tile, in 2-cycle mode
	Shade       rgba // interpolated across the triangle
	Prim        rgba // Set_Prim_Color
	Env         rgba // Set_Env_Color
	Key         rgba // chroma-key centre/scale, unused so far
	Comb        rgba // the previous cycle's output
	LODFrac     uint32
	PrimLODFrac uint32
}

type rgba struct{ R, G, B, A uint32 }

// The combiner's colour multiplexers. The sources coincide for the first six
// entries and then diverge, which is why each needs its own table.
func (in *combineInputs) subA(sel uint32) uint32Vec3 {
	switch sel {
	case 0:
		return vec3(in.Comb)
	case 1:
		return vec3(in.Texel0)
	case 2:
		return vec3(in.Texel1)
	case 3:
		return vec3(in.Prim)
	case 4:
		return vec3(in.Shade)
	case 5:
		return vec3(in.Env)
	case 6:
		return uint32Vec3{255, 255, 255}
	case 7:
		return uint32Vec3{0, 0, 0} // NOISE, which is not modelled
	}
	return uint32Vec3{}
}

func (in *combineInputs) subB(sel uint32) uint32Vec3 {
	switch sel {
	case 0:
		return vec3(in.Comb)
	case 1:
		return vec3(in.Texel0)
	case 2:
		return vec3(in.Texel1)
	case 3:
		return vec3(in.Prim)
	case 4:
		return vec3(in.Shade)
	case 5:
		return vec3(in.Env)
	case 6:
		return vec3(in.Key) // KEYCENTER
	case 7:
		return uint32Vec3{} // K4
	}
	return uint32Vec3{}
}

func (in *combineInputs) mul(sel uint32) uint32Vec3 {
	switch sel {
	case 0:
		return vec3(in.Comb)
	case 1:
		return vec3(in.Texel0)
	case 2:
		return vec3(in.Texel1)
	case 3:
		return vec3(in.Prim)
	case 4:
		return vec3(in.Shade)
	case 5:
		return vec3(in.Env)
	case 6:
		return vec3(in.Key) // KEYSCALE
	case 7:
		return splat(in.Comb.A)
	case 8:
		return splat(in.Texel0.A)
	case 9:
		return splat(in.Texel1.A)
	case 10:
		return splat(in.Prim.A)
	case 11:
		return splat(in.Shade.A)
	case 12:
		return splat(in.Env.A)
	case 13:
		return splat(in.LODFrac)
	case 14:
		return splat(in.PrimLODFrac)
	case 15:
		return uint32Vec3{} // K5
	}
	return uint32Vec3{}
}

func (in *combineInputs) add(sel uint32) uint32Vec3 {
	switch sel {
	case 0:
		return vec3(in.Comb)
	case 1:
		return vec3(in.Texel0)
	case 2:
		return vec3(in.Texel1)
	case 3:
		return vec3(in.Prim)
	case 4:
		return vec3(in.Shade)
	case 5:
		return vec3(in.Env)
	case 6:
		return uint32Vec3{255, 255, 255}
	}
	return uint32Vec3{}
}

// The alpha multiplexers. subA, subB and add share a table; mul has its own,
// because its first entry is the level-of-detail fraction rather than the
// previous cycle's alpha.
func (in *combineInputs) alphaABD(sel uint32) uint32 {
	switch sel {
	case 0:
		return in.Comb.A
	case 1:
		return in.Texel0.A
	case 2:
		return in.Texel1.A
	case 3:
		return in.Prim.A
	case 4:
		return in.Shade.A
	case 5:
		return in.Env.A
	case 6:
		return 255
	}
	return 0
}

func (in *combineInputs) alphaMul(sel uint32) uint32 {
	switch sel {
	case 0:
		return in.LODFrac
	case 1:
		return in.Texel0.A
	case 2:
		return in.Texel1.A
	case 3:
		return in.Prim.A
	case 4:
		return in.Shade.A
	case 5:
		return in.Env.A
	case 6:
		return in.PrimLODFrac
	}
	return 0
}

type uint32Vec3 struct{ R, G, B uint32 }

func vec3(c rgba) uint32Vec3    { return uint32Vec3{c.R, c.G, c.B} }
func splat(v uint32) uint32Vec3 { return uint32Vec3{v, v, v} }

// combineChannel is (a - b) * c + d, on one 8-bit channel, with the hardware's
// rounding and clamping.
func combineChannel(a, b, c, d uint32) uint32 {
	v := (int32(a)-int32(b))*int32(c) + 0x80
	v = (v >> 8) + int32(d)
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint32(v)
}

// combinerSelects unpacks the eight multiplexers of one cycle out of the 56-bit
// Set_Combine_Mode word. The two cycles' fields are interleaved across it rather
// than laid out one after the other.
type combinerSelects struct {
	subAC, subBC, mulC, addC uint32
	subAA, subBA, mulA, addA uint32
}

func (r *rdp) combinerSelects(cycle int) combinerSelects {
	w := r.Combine
	if cycle == 0 {
		return combinerSelects{
			subAC: uint32(w >> 52 & 0xF), subBC: uint32(w >> 28 & 0xF),
			mulC: uint32(w >> 47 & 0x1F), addC: uint32(w >> 15 & 0x7),
			subAA: uint32(w >> 44 & 0x7), subBA: uint32(w >> 12 & 0x7),
			mulA: uint32(w >> 41 & 0x7), addA: uint32(w >> 9 & 0x7),
		}
	}
	return combinerSelects{
		subAC: uint32(w >> 37 & 0xF), subBC: uint32(w >> 24 & 0xF),
		mulC: uint32(w >> 32 & 0x1F), addC: uint32(w >> 6 & 0x7),
		subAA: uint32(w >> 21 & 0x7), subBA: uint32(w >> 3 & 0x7),
		mulA: uint32(w >> 18 & 0x7), addA: uint32(w & 0x7),
	}
}

// combine runs the combiner for the current cycle type and returns the pixel's
// colour. In 2-cycle mode the first cycle's output feeds the second through the
// COMBINED input, which is what makes multi-texturing possible at all.
func (r *rdp) combine(in *combineInputs) rgba {
	cycles := 1
	if r.cycleType() == cycle2 {
		cycles = 2
	}
	var out rgba
	for c := 0; c < cycles; c++ {
		s := r.combinerSelects(c)
		a, b, m, d := in.subA(s.subAC), in.subB(s.subBC), in.mul(s.mulC), in.add(s.addC)
		out = rgba{
			R: combineChannel(a.R, b.R, m.R, d.R),
			G: combineChannel(a.G, b.G, m.G, d.G),
			B: combineChannel(a.B, b.B, m.B, d.B),
			A: combineChannel(in.alphaABD(s.subAA), in.alphaABD(s.subBA),
				in.alphaMul(s.mulA), in.alphaABD(s.addA)),
		}
		in.Comb = out
	}
	return out
}

// --- the blender -------------------------------------------------------------

// Set_Other_Modes bits the blender and the pixel pipeline read.
const (
	omAlphaCompare  = 1 << 0
	omZSourceSel    = 1 << 2
	omAntialias     = 1 << 3
	omZCompare      = 1 << 4
	omZUpdate       = 1 << 5
	omImageRead     = 1 << 6
	omCvgTimesAlpha = 1 << 12
	omAlphaCvgSel   = 1 << 13
	omForceBlend    = 1 << 14
)

// blenderSelects unpacks the four blender multiplexers for one cycle.
func (r *rdp) blenderSelects(cycle int) (p, a, m, b uint32) {
	w := r.OtherModes
	if cycle == 0 {
		return uint32(w >> 30 & 3), uint32(w >> 26 & 3), uint32(w >> 22 & 3), uint32(w >> 18 & 3)
	}
	return uint32(w >> 28 & 3), uint32(w >> 24 & 3), uint32(w >> 20 & 3), uint32(w >> 16 & 3)
}

// blend merges the combiner's output with the pixel already in the framebuffer.
//
//	result = (p*a + m*b) / (a + b)
//
// The division is what makes the blender an interpolator rather than a weighted
// sum, and it is why a source-over blend needs no explicit 1/255. With
// force_blend the divide is skipped and the sum is taken as already normalised.
func (r *rdp) blend(cyc rgba, mem rgba, shadeAlpha uint32) rgba {
	cycle := 0
	if r.cycleType() == cycle2 {
		cycle = 1
	}
	pSel, aSel, mSel, bSel := r.blenderSelects(cycle)

	pick := func(sel uint32) rgba {
		switch sel {
		case 0:
			return cyc
		case 1:
			return mem
		case 2:
			return rgba{r.BlendColor >> 24 & 255, r.BlendColor >> 16 & 255, r.BlendColor >> 8 & 255, r.BlendColor & 255}
		default:
			return rgba{r.FogColor >> 24 & 255, r.FogColor >> 16 & 255, r.FogColor >> 8 & 255, r.FogColor & 255}
		}
	}
	p, m := pick(pSel), pick(mSel)

	var a uint32
	switch aSel {
	case 0:
		a = cyc.A
	case 1:
		a = r.FogColor & 255
	case 2:
		a = shadeAlpha
	default:
		a = 0
	}
	var b uint32
	switch bSel {
	case 0:
		b = 255 - a
	case 1:
		b = mem.A
	case 2:
		b = 255
	default:
		b = 0
	}

	mix := func(pc, mc uint32) uint32 {
		num := pc*a + mc*b
		den := a + b
		if den == 0 {
			return pc
		}
		if r.OtherModes&omForceBlend != 0 {
			den = 255
		}
		v := num / den
		if v > 255 {
			return 255
		}
		return v
	}
	return rgba{mix(p.R, m.R), mix(p.G, m.G), mix(p.B, m.B), cyc.A}
}
