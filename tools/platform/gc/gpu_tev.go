package gc

// gpu_tev.go is the TEV — the texture-environment unit — the programmable combiner at the end
// of the pipe that decides a pixel's final colour. Where a modern GPU runs a fragment shader,
// Flipper runs up to sixteen fixed stages, each of which takes four inputs (any of the
// interpolated vertex colour, a texture sample, a constant, or a previous stage's result),
// blends them by one small formula, and writes the result to one of four working registers.
// The last stage's output is the pixel. This file evaluates that pipeline per pixel, applies
// the alpha test that can reject the pixel outright, and blends what survives into the
// framebuffer.
//
// The formula every stage computes is the same:
//
//	out = (d  ±  lerp(a, b, c)) * scale  +  bias        (clamped to [0,1] when asked)
//
// where the lerp is a*(1-c) + b*c and c doubles as the interpolation weight. Choosing a, b, c
// and d from the sixteen possible inputs, and the register the result lands in, is the whole of
// what a stage's two combiner registers (colour and alpha) encode. The inputs, the scale/bias
// codes, and the blend equation are hardware facts; the arithmetic is reimplemented here.
//
// The loading screen this stage first draws is a two-stage program: stage 0 puts the texture
// colour into the working register, and stage 1 interpolates between two constant colours using
// that as the weight — which, when the two constants are momentarily equal (a crossfade at rest),
// discards the texture and leaves a flat colour. That the texture is nonetheless sampled and
// decoded correctly is what gpu_texture.go's gate proves; that the combiner reproduces the flat
// result the hardware would is what this file is for.

// tevColorReg decodes one of the TEV colour/output registers into its four channels. The
// register is two words: the low word packs alpha and red, the high word blue and green, each
// an eleven-bit signed value (the registers can hold out-of-range intermediates).
func tevColorReg(w [2]uint32) (r, g, b, a float32) {
	a = float32(sext11(w[0] & 0x7FF))
	r = float32(sext11((w[0] >> 12) & 0x7FF))
	b = float32(sext11(w[1] & 0x7FF))
	g = float32(sext11((w[1] >> 12) & 0x7FF))
	return
}

// sext11 sign-extends an eleven-bit field.
func sext11(v uint32) int32 {
	if v&0x400 != 0 {
		return int32(v) - 0x800
	}
	return int32(v)
}

// shade runs the TEV pipeline for one pixel. It is given the rasterised (interpolated vertex)
// colour and the texture coordinate, and returns the final colour together with whether the
// pixel survived the alpha test.
func (g *gpu) shade(m *Machine, rasR, rasG, rasB, rasA uint8, u, v float32) (fr, fg, fb, fa uint8, pass bool) {
	// The four working registers, seeded with the constant values the game loaded.
	var reg [4][4]float32
	for i := 0; i < 4; i++ {
		r, gg, b, a := tevColorReg(g.TevColorReg[i])
		reg[i] = [4]float32{r, gg, b, a}
	}
	ras := [4]float32{float32(rasR), float32(rasG), float32(rasB), float32(rasA)}

	numStages := int((g.BP[0x00]>>10)&0xF) + 1
	for s := 0; s < numStages; s++ {
		ord := g.BP[0x28+uint32(s/2)]
		if s&1 == 1 {
			ord >>= 12
		}
		texmap := int(ord & 7)
		texcoord := int((ord >> 3) & 7)
		texEnable := (ord>>6)&1 != 0

		var tex [4]float32
		if texEnable {
			// Only texture coordinate 0 is interpolated and fetched; a stage that reads
			// another coordinate is a ticket, logged once, not a wrong sample from coord 0.
			if texcoord != 0 {
				m.logf("TEV: stage %d samples texcoord %d, but only texcoord0 is fetched", s, texcoord)
			}
			tr, tg, tb, ta := g.sampleTexmap(m, texmap, u, v)
			tex = [4]float32{float32(tr), float32(tg), float32(tb), float32(ta)}
		}

		kc := g.konstColor(s)
		ka := g.konstAlpha(s)

		cc := g.BP[0xC0+uint32(s)*2]
		ac := g.BP[0xC1+uint32(s)*2]

		crgb := combineColor(cc, reg, tex, ras, kc)
		aout := combineAlpha(ac, reg, tex, ras, ka)

		cdest := (cc >> 22) & 3
		adest := (ac >> 22) & 3
		reg[cdest][0], reg[cdest][1], reg[cdest][2] = crgb[0], crgb[1], crgb[2]
		reg[adest][3] = aout
	}

	out := reg[0] // the pixel is the final value of the PREV register
	a8 := toU8(out[3])
	if !g.alphaTest(a8) {
		return 0, 0, 0, 0, false
	}
	return toU8(out[0]), toU8(out[1]), toU8(out[2]), a8, true
}

// toU8 rounds a combiner result to a byte, so a value that lands at 254.999 through float
// arithmetic reads back as the 255 the hardware's fixed-point path produces.
func toU8(f float32) uint8 { return uint8(clampf(f) + 0.5) }

// combineColor evaluates a stage's colour combiner: four rgb inputs chosen from the sixteen
// codes, then the shared formula applied per channel.
func combineColor(cc uint32, reg [4][4]float32, tex, ras [4]float32, konst [3]float32) [3]float32 {
	d := colorArg(int(cc&0xF), reg, tex, ras, konst)
	c := colorArg(int((cc>>4)&0xF), reg, tex, ras, konst)
	b := colorArg(int((cc>>8)&0xF), reg, tex, ras, konst)
	a := colorArg(int((cc>>12)&0xF), reg, tex, ras, konst)
	bias := int((cc >> 16) & 3)
	sub := (cc>>18)&1 != 0
	clamp := (cc>>19)&1 != 0
	scale := int((cc >> 20) & 3)
	var out [3]float32
	for i := 0; i < 3; i++ {
		out[i] = tevFormula(a[i], b[i], c[i], d[i], bias, sub, scale, clamp)
	}
	return out
}

// combineAlpha evaluates a stage's alpha combiner: four scalar inputs from the eight alpha
// codes, then the same formula.
func combineAlpha(ac uint32, reg [4][4]float32, tex, ras [4]float32, konst float32) float32 {
	d := alphaArg(int((ac>>4)&7), reg, tex, ras, konst)
	c := alphaArg(int((ac>>7)&7), reg, tex, ras, konst)
	b := alphaArg(int((ac>>10)&7), reg, tex, ras, konst)
	a := alphaArg(int((ac>>13)&7), reg, tex, ras, konst)
	bias := int((ac >> 16) & 3)
	sub := (ac>>18)&1 != 0
	clamp := (ac>>19)&1 != 0
	scale := int((ac >> 20) & 3)
	return tevFormula(a, b, c, d, bias, sub, scale, clamp)
}

// tevFormula is the arithmetic every stage shares, in 0..255 units: interpolate a and b by c,
// add or subtract from d, bias, scale, and optionally clamp.
func tevFormula(a, b, c, d float32, bias int, sub bool, scale int, clamp bool) float32 {
	cc := c / 255
	lerp := a*(1-cc) + b*cc
	var r float32
	if sub {
		r = d - lerp
	} else {
		r = d + lerp
	}
	switch bias {
	case 1:
		r += 127.5
	case 2:
		r -= 127.5
	}
	switch scale {
	case 1:
		r *= 2
	case 2:
		r *= 4
	case 3:
		r *= 0.5
	}
	if clamp {
		if r < 0 {
			r = 0
		}
		if r > 255 {
			r = 255
		}
	}
	return r
}

// colorArg resolves one of the sixteen colour-combiner inputs to an rgb triple. Alpha inputs
// are broadcast across the three channels, as the hardware does.
func colorArg(code int, reg [4][4]float32, tex, ras [4]float32, konst [3]float32) [3]float32 {
	rep := func(v float32) [3]float32 { return [3]float32{v, v, v} }
	switch code {
	case 0: // CPREV
		return [3]float32{reg[0][0], reg[0][1], reg[0][2]}
	case 1: // APREV
		return rep(reg[0][3])
	case 2: // C0
		return [3]float32{reg[1][0], reg[1][1], reg[1][2]}
	case 3: // A0
		return rep(reg[1][3])
	case 4: // C1
		return [3]float32{reg[2][0], reg[2][1], reg[2][2]}
	case 5: // A1
		return rep(reg[2][3])
	case 6: // C2
		return [3]float32{reg[3][0], reg[3][1], reg[3][2]}
	case 7: // A2
		return rep(reg[3][3])
	case 8: // TEXC
		return [3]float32{tex[0], tex[1], tex[2]}
	case 9: // TEXA
		return rep(tex[3])
	case 10: // RASC
		return [3]float32{ras[0], ras[1], ras[2]}
	case 11: // RASA
		return rep(ras[3])
	case 12: // ONE
		return rep(255)
	case 13: // HALF
		return rep(128)
	case 14: // KONST
		return konst
	default: // ZERO
		return [3]float32{}
	}
}

// alphaArg resolves one of the eight alpha-combiner inputs to a scalar.
func alphaArg(code int, reg [4][4]float32, tex, ras [4]float32, konst float32) float32 {
	switch code {
	case 0: // APREV
		return reg[0][3]
	case 1: // A0
		return reg[1][3]
	case 2: // A1
		return reg[2][3]
	case 3: // A2
		return reg[3][3]
	case 4: // TEXA
		return tex[3]
	case 5: // RASA
		return ras[3]
	case 6: // KONST
		return konst
	default: // ZERO
		return 0
	}
}

// konstColor and konstAlpha resolve a stage's constant selection (from the KSEL registers) to a
// value. The selection is either one of eight fixed fractions of white or a component of one of
// the four constant registers.
func (g *gpu) konstColor(stage int) [3]float32 {
	sel := g.kSel(stage, false)
	if sel < 8 {
		v := float32(8-sel) / 8 * 255
		return [3]float32{v, v, v}
	}
	if sel >= 12 {
		r, gg, b, a := tevColorReg(g.TevKonstReg[(sel-12)&3])
		switch (sel - 12) / 4 {
		case 0: // whole rgb
			return [3]float32{r, gg, b}
		case 1: // red
			return [3]float32{r, r, r}
		case 2: // green
			return [3]float32{gg, gg, gg}
		case 3: // blue
			return [3]float32{b, b, b}
		default: // alpha
			return [3]float32{a, a, a}
		}
	}
	return [3]float32{} // the reserved codes read as zero
}

func (g *gpu) konstAlpha(stage int) float32 {
	sel := g.kSel(stage, true)
	if sel < 8 {
		return float32(8-sel) / 8 * 255
	}
	if sel >= 16 {
		r, gg, b, a := tevColorReg(g.TevKonstReg[(sel-16)&3])
		switch (sel - 16) / 4 {
		case 0:
			return r
		case 1:
			return gg
		case 2:
			return b
		default:
			return a
		}
	}
	return 0
}

// kSel reads a stage's constant-colour or constant-alpha selection out of the KSEL register
// that covers it. Each of the eight KSEL registers holds the selections for two stages.
func (g *gpu) kSel(stage int, alpha bool) int {
	reg := g.BP[0xF6+uint32(stage/2)]
	shift := uint(4) // the even stage's fields
	if stage&1 == 1 {
		shift = 14
	}
	if alpha {
		shift += 5
	}
	return int((reg >> shift) & 0x1F)
}

// alphaTest applies the two-comparison alpha test the game programmed (BP 0xF3): each
// comparison tests the pixel's alpha against a reference, and a logic op combines the two.
func (g *gpu) alphaTest(a uint8) bool {
	f := g.BP[0xF3]
	ref0 := uint8(f & 0xFF)
	ref1 := uint8((f >> 8) & 0xFF)
	r0 := alphaCompare(a, ref0, int((f>>16)&7))
	r1 := alphaCompare(a, ref1, int((f>>19)&7))
	switch (f >> 22) & 3 {
	case 0: // AND
		return r0 && r1
	case 1: // OR
		return r0 || r1
	case 2: // XOR
		return r0 != r1
	default: // XNOR
		return r0 == r1
	}
}

func alphaCompare(a, ref uint8, comp int) bool {
	switch comp {
	case 0:
		return false // NEVER
	case 1:
		return a < ref
	case 2:
		return a == ref
	case 3:
		return a <= ref
	case 4:
		return a > ref
	case 5:
		return a != ref
	case 6:
		return a >= ref
	default:
		return true // ALWAYS
	}
}

// blend combines a shaded pixel with what is already in the framebuffer, following the blend
// equation the game set in the colour mode register (BP 0x41): a source and destination factor,
// added or subtracted, with separate enables for the colour and alpha writes. With blending
// disabled the pixel simply replaces what was there.
func (g *gpu) blend(dst uint32, sr, sg, sb, sa uint8) uint32 {
	cm := g.BP[0x41]
	dr, dgc, db := unpackRGB(dst)
	da := uint8(dst) // the EFB keeps alpha in the low byte (see packRGBA)

	if cm&1 == 0 { // blending disabled: replace, honouring the update masks below
		return g.applyUpdates(cm, dst, sr, sg, sb, sa)
	}

	saf := float32(sa) / 255
	daf := float32(da) / 255
	srcF := func(chanOther float32) float32 { return blendFactor(int((cm>>8)&7), chanOther/255, saf, daf) }
	dstF := func(chanSelf float32) float32 { return blendFactor(int((cm>>5)&7), chanSelf/255, saf, daf) }
	sub := (cm>>11)&1 != 0

	comb := func(s, d uint8) uint8 {
		sf := srcF(float32(d))
		df := dstF(float32(s))
		var v float32
		if sub {
			v = float32(d)*df - float32(s)*sf
		} else {
			v = float32(s)*sf + float32(d)*df
		}
		return toU8(v)
	}
	nr := comb(sr, dr)
	ng := comb(sg, dgc)
	nb := comb(sb, db)
	na := comb(sa, da)
	return g.applyUpdates(cm, dst, nr, ng, nb, na)
}

// applyUpdates writes only the channels the colour-mode register enables, leaving the rest of
// the destination pixel as it was.
func (g *gpu) applyUpdates(cm, dst uint32, r, gg, b, a uint8) uint32 {
	or, og, ob := unpackRGB(dst)
	oa := uint8(dst)
	if cm&(1<<3) != 0 { // colour update
		or, og, ob = r, gg, b
	}
	if cm&(1<<4) != 0 { // alpha update
		oa = a
	}
	return packRGBA(or, og, ob, oa)
}

// blendFactor turns a three-bit blend-factor code into a multiplier in [0,1]. The colour codes
// (2,3) take the value from the *other* operand's channel — the source factor scaling by the
// destination colour and vice versa — which the caller passes in as chanOther.
func blendFactor(code int, chanOther, sa, da float32) float32 {
	switch code {
	case 0: // ZERO
		return 0
	case 1: // ONE
		return 1
	case 2: // source uses dst colour, dest uses src colour
		return chanOther
	case 3:
		return 1 - chanOther
	case 4: // SRC alpha
		return sa
	case 5:
		return 1 - sa
	case 6: // DST alpha
		return da
	default:
		return 1 - da
	}
}

// clampf saturates a float to a byte's range.
func clampf(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
