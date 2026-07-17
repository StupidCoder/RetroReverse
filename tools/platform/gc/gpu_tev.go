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
// The boot logo this stage first draws is a two-stage program that uses the texture as a
// coverage mask: stage 0 puts the texture (a white "Nintendo" logo on black) into the working
// register, and stage 1 interpolates between two constant colours with that as the weight — a
// black background constant where the texture is black, a red logo constant where it is white.
// The result is the red logo on black the console shows. Reproducing it exactly is what this
// file is for, and it is unforgiving: the two constants differ only in their red channel, so a
// register-decode that misplaced red would collapse the mask and paint the screen flat.

// tevColorReg decodes one of the TEV colour/output registers into its four channels. The
// register is two words: the low word packs red in its low eleven bits and alpha in its high
// eleven, the high word blue then green the same way — each an eleven-bit signed value, since
// the registers hold out-of-range intermediates. Getting this order backwards (reading red from
// the high half) turned the boot's black background constant into a second red, so the logo
// mask composited red-on-red and the whole screen came out flat red instead of a red logo on
// black.
func tevColorReg(w [2]uint32) (r, g, b, a float32) {
	r = float32(sext11(w[0] & 0x7FF))
	a = float32(sext11((w[0] >> 12) & 0x7FF))
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

// shade runs the TEV pipeline for one pixel. It is given the pipeline's decoded configuration
// (gpu_tev_state.go — resolved once per draw, because none of it can change while a draw is
// running), the rasterised (interpolated vertex) colour and the pixel's generated texture
// coordinates, and returns the final colour together with whether the pixel survived the alpha
// test.
func (g *gpu) shade(m *Machine, t *tevState, rasCol *[2][4]uint8, tc *[maxTexCoord]texCoord) (fr, fg, fb, fa uint8, pass bool) {
	// The four working registers, seeded with the constant values the game loaded. This is a
	// copy and must stay one: the stages below write it.
	reg := t.seed

	for s := 0; s < t.numStages; s++ {
		st := &t.stages[s]

		// The colour channel this stage reads, chosen per stage. See rasSelect: a stage may
		// name either lit channel, one channel's alpha broadcast, or nothing at all.
		ras := rasSelect(st.rasSel, rasCol)

		// Each stage reads its two colour inputs through a swap table of its own choosing, named
		// in the low bits of its alpha combiner. The tables are not decoration: this game's
		// shadow test cannot be expressed without them. A texture sampled as IA8 arrives as
		// (I,I,I,A), so the r:g pair a 16-bit compare reads would be (I,I) — one byte twice
		// over, and no 16-bit quantity at all. The swap is what lifts the alpha into g and makes
		// r:g the two halves of one number: the shadow's depth ramp swaps g<-a to assemble its
		// 16-bit value, and the shadow map swaps r<-a to assemble the depth's.
		//
		// The TABLE is per-draw and has been hoisted; the SWIZZLED VALUE is per-stage and has
		// not. Writing either back over the loop's own ras would hand the next stage a colour
		// the rasteriser never produced.
		sras := swizzle(st.swapRas, ras)

		var tex [4]float32
		if st.texEnable {
			// Each stage samples the coordinate it names. A stage may read a coordinate the
			// transform unit was not asked to generate — that is the game's own error, not a
			// gap here, and the coordinate reads as the zero it was left at.
			tr, tg, tb, ta := g.sampleTexmap(m, &t.tex[st.texmap], tc[st.texcoord].s, tc[st.texcoord].t)
			tex = swizzle(st.swapTex, [4]float32{float32(tr), float32(tg), float32(tb), float32(ta)})
		}

		crgb, ca, cb := combineColor(st.cc, &reg, tex, sras, st.konstC)
		aout := combineAlpha(st.ac, &reg, tex, sras, st.konstA, ca, cb)

		reg[st.cdest][0], reg[st.cdest][1], reg[st.cdest][2] = crgb[0], crgb[1], crgb[2]
		reg[st.adest][3] = aout
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

// rasSelect resolves a stage's RAS input to one of the two lit colour channels, or to nothing.
//
// The encoding is the game's own, and it is a table rather than a guess: GXSetTevOrder
// (0x801F8354) maps GX_COLOR_NULL straight to 7 and puts every other GXChannelID through a
// lookup at 0x80395F78, which reads
//
//	COLOR0 0  COLOR1 1  ALPHA0 0  ALPHA1 1  COLOR0A0 0  COLOR1A1 1  ZERO 7  BUMP 5  BUMPN 6
//
// so the field names the CHANNEL only — whether a stage wants that channel's colour or its
// alpha is the combiner's business (RASC vs RASA), not this field's. Hence two cases, not six.
//
// 5 and 6 are the bump-alpha forms, which carry the value the emboss texgen computes rather
// than a lit channel; nothing generates one yet, so they read as zero and say so once.
func rasSelect(sel uint32, col *[2][4]uint8) [4]float32 {
	switch sel {
	case 0, 1:
		c := col[sel]
		return [4]float32{float32(c[0]), float32(c[1]), float32(c[2]), float32(c[3])}
	case 7: // NULL: the rasteriser hands the stage nothing
		return [4]float32{}
	}
	return [4]float32{}
}

// combineColor evaluates a stage's colour combiner: four rgb inputs chosen from the sixteen
// codes, then either the arithmetic formula applied per channel or — when the bias field
// reads 3 — the comparison the stage has been switched to instead (see tevCompare).
//
// The colour operands are returned alongside the result because the ALPHA combiner's
// comparison reads them rather than its own: see combineAlpha.
//
// The working registers come through as a POINTER, and that is a performance change with no
// modelling content: they are read-only here, and the four argument selections below plus the
// alpha combiner's four made ten copies of the same sixty-four bytes for every fragment. The
// callee must not write through it — nothing here does, and the stage's result is applied by
// shade, which owns the array.
func combineColor(cc uint32, reg *[4][4]float32, tex, ras [4]float32, konst [3]float32) (out, ca, cb [3]float32) {
	d := colorArg(int(cc&0xF), reg, tex, ras, konst)
	c := colorArg(int((cc>>4)&0xF), reg, tex, ras, konst)
	b := colorArg(int((cc>>8)&0xF), reg, tex, ras, konst)
	a := colorArg(int((cc>>12)&0xF), reg, tex, ras, konst)
	bias := int((cc >> 16) & 3)
	op := (cc>>18)&1 != 0
	clamp := (cc>>19)&1 != 0
	scale := int((cc >> 20) & 3)

	if bias == tevBiasCompare {
		// Compare mode: the scale field is no longer a shift and the sub bit is no longer a
		// sign — together they name the comparison (see tevCompare).
		sel := scale<<1 | b2i(op)
		for i := 0; i < 3; i++ {
			// RGB8 compares each channel against its own opposite number; the narrower
			// widths compare one packed value and apply the single verdict to all three.
			ch := i
			if sel>>1 != 3 {
				ch = -1
			}
			if tevCompare(sel, a, b, 0, 0, ch) {
				out[i] = d[i] + c[i]
			} else {
				out[i] = d[i]
			}
			if clamp {
				out[i] = clampf(out[i])
			}
		}
		return out, a, b
	}
	for i := 0; i < 3; i++ {
		out[i] = tevFormula(a[i], b[i], c[i], d[i], bias, op, scale, clamp)
	}
	return out, a, b
}

// combineAlpha evaluates a stage's alpha combiner: four scalar inputs from the eight alpha
// codes, then the arithmetic formula or — when the bias field reads 3 — a comparison.
//
// A comparison here is the one place the two combiners are not independent. Only the widest
// selector (A8) compares the alpha operands; every narrower one compares the COLOUR pipe's a
// and b, which is why they arrive as arguments. This game's shadow stage is the proof: its
// alpha compare selects both of its own operands as ZERO and would be a no-op that always
// took the same branch, while the colour operands either side of it are the fragment's
// light-space depth and the shadow map's — the two values the shadow test exists to compare.
func combineAlpha(ac uint32, reg *[4][4]float32, tex, ras [4]float32, konst float32, ca, cb [3]float32) float32 {
	d := alphaArg(int((ac>>4)&7), reg, tex, ras, konst)
	c := alphaArg(int((ac>>7)&7), reg, tex, ras, konst)
	b := alphaArg(int((ac>>10)&7), reg, tex, ras, konst)
	a := alphaArg(int((ac>>13)&7), reg, tex, ras, konst)
	bias := int((ac >> 16) & 3)
	op := (ac>>18)&1 != 0
	clamp := (ac>>19)&1 != 0
	scale := int((ac >> 20) & 3)

	if bias == tevBiasCompare {
		sel := scale<<1 | b2i(op)
		out := d
		if tevCompare(sel, ca, cb, a, b, -1) {
			out += c
		}
		if clamp {
			out = clampf(out)
		}
		return out
	}
	return tevFormula(a, b, c, d, bias, op, scale, clamp)
}

// tevBiasCompare is the bias encoding that means "this stage is a comparison, not a sum".
const tevBiasCompare = 3

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// tevCompare evaluates a compare-mode stage's test. The selector is the scale field and the
// op bit together — (scale<<1)|op — which is how the library that built these registers
// encodes the eight comparisons: four operand widths, each greater-than or equal-to.
//
//	0/1 R8      2/3 GR16     4/5 BGR24     6/7 RGB8 (colour) or A8 (alpha)
//
// The narrow widths read the colour operands as little-endian bytes of one unsigned number:
// GR16 is green:red, BGR24 is blue:green:red. That is what lets a stage compare a 16- or
// 24-bit quantity that has been spread across channels, which is exactly how a depth value
// arrives from a Z-format texture.
//
// ch selects a single colour channel for the per-channel RGB8 form; -1 means the caller
// wants the one packed verdict. aa/ab are the alpha operands, read only by A8.
func tevCompare(sel int, ca, cb [3]float32, aa, ab float32, ch int) bool {
	u8 := func(f float32) uint32 { return uint32(clampf(f) + 0.5) }
	pack := func(v [3]float32, n int) uint32 {
		var r uint32
		for i := n - 1; i >= 0; i-- {
			r = r<<8 | u8(v[i])
		}
		return r
	}
	var x, y uint32
	switch sel >> 1 {
	case 0: // R8
		x, y = pack(ca, 1), pack(cb, 1)
	case 1: // GR16
		x, y = pack(ca, 2), pack(cb, 2)
	case 2: // BGR24
		x, y = pack(ca, 3), pack(cb, 3)
	default: // RGB8 per channel, or A8 when the alpha combiner asks
		if ch < 0 {
			x, y = u8(aa), u8(ab)
		} else {
			x, y = u8(ca[ch]), u8(cb[ch])
		}
	}
	if sel&1 == 0 {
		return x > y
	}
	return x == y
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
func colorArg(code int, reg *[4][4]float32, tex, ras [4]float32, konst [3]float32) [3]float32 {
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
func alphaArg(code int, reg *[4][4]float32, tex, ras [4]float32, konst float32) float32 {
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

// swapTable reads one of the four channel swap tables the TEV keeps. The tables share the
// eight KSEL registers with the constant selections, two channels to a register in the low
// nibble the constants leave free: table t is built from KSEL 0xF6+2t (red, then green) and
// 0xF6+2t+1 (blue, then alpha), each entry two bits naming the source channel.
//
// Table 0 is conventionally the identity, which is why ignoring the tables altogether looked
// right for so long — every draw that leaves a stage on table 0 is unaffected.
func (g *gpu) swapTable(t int) [4]int {
	lo := g.BP[0xF6+uint32(t)*2]
	hi := g.BP[0xF6+uint32(t)*2+1]
	return [4]int{int(lo & 3), int((lo >> 2) & 3), int(hi & 3), int((hi >> 2) & 3)}
}

// swizzle applies a swap table to one colour: each output channel takes the input channel the
// table names for it.
func swizzle(t [4]int, v [4]float32) [4]float32 {
	return [4]float32{v[t[0]], v[t[1]], v[t[2]], v[t[3]]}
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
