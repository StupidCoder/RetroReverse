package xbox

// nv2a_combiner.go evaluates the NV2A register combiners — the fragment pipeline that
// stands where the 3DS's TEV does. Up to 8 general stages each compute
// AB = A op B and CD = C op D (componentwise product or dot product) and a sum/mux,
// writing any of them back into the register set; a final combiner then produces
// A*B + (1-A)*C + D. The input/output control words are the hardware's
// (GL_NV_register_combiners is the public spec of this exact silicon block); the
// register file is: 0 zero, 1/2 the constant colors, 3 fog, 4/5 the interpolated
// diffuse/specular, 8-11 the four texture samples, 12/13 the spare registers.
//
// OutRun's loading screen programs exactly one stage — spare0 = TEX0 * DIFFUSE with
// the matching alpha product — and a final combiner that passes spare0 through, which
// is the D3D MODULATE texture-stage state compiled down.

const (
	kelvinCombinerAlphaICW = 0x0260 // +4*i, 8 stages
	kelvinSpecFogCW0       = 0x0288 // final combiner: A,B,C,D input bytes
	kelvinSpecFogCW1       = 0x028C // final combiner: E,F,G input bytes + flags
	kelvinCombinerFactor0  = 0x0A60 // +4*i
	kelvinCombinerFactor1  = 0x0A80 // +4*i
	kelvinCombinerAlphaOCW = 0x0AA0 // +4*i
	kelvinCombinerColorICW = 0x0AC0 // +4*i
	kelvinCombinerColorOCW = 0x1E40 // +4*i
	kelvinCombinerControl  = 0x1E60 // bits 7:0 stage count; bits 8/12 per-stage factors
)

// combInput is what the rasteriser feeds one fragment's combiner run.
type combInput struct {
	tex        [4][4]float32
	col0, col1 [4]float32
}

// combStage is one decoded general combiner stage.
type combStage struct {
	colorICW, alphaICW uint32
	colorOCW, alphaOCW uint32
	factor0, factor1   [4]float32
}

// combState is a draw's decoded combiner configuration.
type combState struct {
	stages   []combStage
	cw0, cw1 uint32
	fogColor [4]float32
	buf      [8]combStage // backing for stages (no per-draw allocation)
}

// combDecode reads the combiner registers once per draw.
func (g *pgraph) combDecode(cs *combState) {
	n := int(g.Regs[kelvinCombinerControl>>2] & 0xFF)
	if n > 8 {
		n = 8
	}
	ctl := g.Regs[kelvinCombinerControl>>2]
	cs.stages = cs.buf[:n]
	for i := 0; i < n; i++ {
		f0i, f1i := i, i
		if ctl>>8&1 == 0 { // factor0 shared across stages
			f0i = 0
		}
		if ctl>>12&1 == 0 {
			f1i = 0
		}
		cs.stages[i] = combStage{
			colorICW: g.Regs[kelvinCombinerColorICW>>2+uint32(i)],
			alphaICW: g.Regs[kelvinCombinerAlphaICW>>2+uint32(i)],
			colorOCW: g.Regs[kelvinCombinerColorOCW>>2+uint32(i)],
			alphaOCW: g.Regs[kelvinCombinerAlphaOCW>>2+uint32(i)],
			factor0:  argb8Vec(g.Regs[kelvinCombinerFactor0>>2+uint32(f0i)]),
			factor1:  argb8Vec(g.Regs[kelvinCombinerFactor1>>2+uint32(f1i)]),
		}
	}
	cs.cw0 = g.Regs[kelvinSpecFogCW0>>2]
	cs.cw1 = g.Regs[kelvinSpecFogCW1>>2]
	cs.fogColor = argb8Vec(g.Regs[0x02A8>>2]) // SET_FOG_COLOR
}

func argb8Vec(v uint32) [4]float32 {
	return [4]float32{
		float32(v >> 16 & 0xFF), float32(v >> 8 & 0xFF), float32(v & 0xFF), float32(v >> 24 & 0xFF),
	}
}

// combRegs is the combiner register file for one fragment.
type combRegs struct {
	c0, c1     [4]float32 // constant colors (per stage)
	fog        [4]float32
	col0, col1 [4]float32
	tex        [4][4]float32
	spare      [2][4]float32
}

func (r *combRegs) read(reg uint32) [4]float32 {
	switch reg {
	case 1:
		return r.c0
	case 2:
		return r.c1
	case 3:
		return r.fog
	case 4:
		return r.col0
	case 5:
		return r.col1
	case 8, 9, 10, 11:
		return r.tex[reg-8]
	case 12:
		return r.spare[0]
	case 13:
		return r.spare[1]
	}
	return [4]float32{}
}

func (r *combRegs) write(reg uint32, v [4]float32, alpha bool) {
	var dst *[4]float32
	switch reg {
	case 4:
		dst = &r.col0
	case 5:
		dst = &r.col1
	case 8, 9, 10, 11:
		dst = &r.tex[reg-8]
	case 12:
		dst = &r.spare[0]
	case 13:
		dst = &r.spare[1]
	default:
		return // 0 = discard; the read-only inputs are not writable
	}
	if alpha {
		dst[3] = v[3]
	} else {
		dst[0], dst[1], dst[2] = v[0], v[1], v[2]
	}
}

// combMap applies an input mapping to a register value (values live in [-1,1]).
func combMap(v [4]float32, mapping uint32) [4]float32 {
	var out [4]float32
	for i, x := range v {
		switch mapping {
		case 0: // UNSIGNED_IDENTITY: max(0, x)
			out[i] = clamp01(x)
		case 1: // UNSIGNED_INVERT: 1 - clamp(x)
			out[i] = 1 - clamp01(x)
		case 2: // EXPAND_NORMAL: 2*max(0,x) - 1
			out[i] = 2*clamp01(x) - 1
		case 3: // EXPAND_NEGATE: -(2*max(0,x) - 1)
			out[i] = -(2*clamp01(x) - 1)
		case 4: // HALF_BIAS_NORMAL: max(0,x) - 0.5
			out[i] = clamp01(x) - 0.5
		case 5: // HALF_BIAS_NEGATE
			out[i] = -(clamp01(x) - 0.5)
		case 6: // SIGNED_IDENTITY
			out[i] = x
		case 7: // SIGNED_NEGATE
			out[i] = -x
		}
	}
	return out
}

// combFetch decodes one input byte {mapping[7:5], alpha[4], register[3:0]} and fetches
// the mapped value; rgb inputs replicate the alpha channel when the alpha bit is set,
// alpha-side inputs read the blue channel when it is clear.
func combFetch(r *combRegs, b uint32, alphaSide bool) [4]float32 {
	v := r.read(b & 0xF)
	if b>>4&1 == 1 {
		v = [4]float32{v[3], v[3], v[3], v[3]}
	} else if alphaSide {
		v = [4]float32{v[2], v[2], v[2], v[2]}
	}
	return combMap(v, b>>5&7)
}

// combine runs the general stages then the final combiner for one fragment.
func (g *pgraph) combine(cs *combState, in *combInput) [4]float32 {
	r := combRegs{
		fog:  [4]float32{cs.fogColor[0] / 255, cs.fogColor[1] / 255, cs.fogColor[2] / 255, 1},
		col0: in.col0,
		col1: in.col1,
		tex:  in.tex,
	}

	for si := range cs.stages {
		st := &cs.stages[si]
		r.c0 = [4]float32{st.factor0[0] / 255, st.factor0[1] / 255, st.factor0[2] / 255, st.factor0[3] / 255}
		r.c1 = [4]float32{st.factor1[0] / 255, st.factor1[1] / 255, st.factor1[2] / 255, st.factor1[3] / 255}

		// Color portion.
		a := combFetch(&r, st.colorICW>>24&0xFF, false)
		b := combFetch(&r, st.colorICW>>16&0xFF, false)
		c := combFetch(&r, st.colorICW>>8&0xFF, false)
		d := combFetch(&r, st.colorICW&0xFF, false)
		ab := mul4(a, b)
		cd := mul4(c, d)
		if st.colorOCW>>13&1 == 1 { // AB dot product
			dp := a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
			ab = [4]float32{dp, dp, dp, dp}
		}
		if st.colorOCW>>12&1 == 1 { // CD dot product
			dp := c[0]*d[0] + c[1]*d[1] + c[2]*d[2]
			cd = [4]float32{dp, dp, dp, dp}
		}
		var sum [4]float32
		if st.colorOCW>>14&1 == 1 { // mux: spare0.a < 0.5 picks AB, else CD (per the spec)
			if r.spare[0][3] < 0.5 {
				sum = ab
			} else {
				sum = cd
			}
		} else {
			sum = add4(ab, cd)
		}
		op := st.colorOCW >> 15 & 7
		ab, cd, sum = combOp(ab, op), combOp(cd, op), combOp(sum, op)
		r.write(st.colorOCW>>4&0xF, ab, false)
		r.write(st.colorOCW&0xF, cd, false)
		r.write(st.colorOCW>>8&0xF, sum, false)

		// Alpha portion (no dot products).
		aa := combFetch(&r, st.alphaICW>>24&0xFF, true)
		ba := combFetch(&r, st.alphaICW>>16&0xFF, true)
		ca := combFetch(&r, st.alphaICW>>8&0xFF, true)
		da := combFetch(&r, st.alphaICW&0xFF, true)
		abA := [4]float32{0, 0, 0, aa[3] * ba[3]}
		cdA := [4]float32{0, 0, 0, ca[3] * da[3]}
		var sumA [4]float32
		if st.alphaOCW>>14&1 == 1 {
			if r.spare[0][3] < 0.5 {
				sumA = abA
			} else {
				sumA = cdA
			}
		} else {
			sumA = [4]float32{0, 0, 0, abA[3] + cdA[3]}
		}
		opA := st.alphaOCW >> 15 & 7
		abA, cdA, sumA = combOp(abA, opA), combOp(cdA, opA), combOp(sumA, opA)
		r.write(st.alphaOCW>>4&0xF, abA, true)
		r.write(st.alphaOCW&0xF, cdA, true)
		r.write(st.alphaOCW>>8&0xF, sumA, true)
	}

	// Final combiner: out.rgb = A*B + (1-A)*C + D; out.a = G. Sources 14/15 are the
	// final-only specials (spare0+col1 sum, E*F product).
	e := finalFetch(&r, cs.cw1>>24&0xFF, [4]float32{})
	f := finalFetch(&r, cs.cw1>>16&0xFF, [4]float32{})
	ef := mul4(e, f)
	a := finalFetch(&r, cs.cw0>>24&0xFF, ef)
	b := finalFetch(&r, cs.cw0>>16&0xFF, ef)
	c := finalFetch(&r, cs.cw0>>8&0xFF, ef)
	d := finalFetch(&r, cs.cw0&0xFF, ef)
	gv := finalFetch(&r, cs.cw1>>8&0xFF, ef)

	var out [4]float32
	for i := 0; i < 3; i++ {
		out[i] = clamp01(a[i]*b[i] + (1-a[i])*c[i] + d[i])
	}
	out[3] = clamp01(gv[3])
	return out
}

// finalFetch is combFetch plus the final combiner's extra sources: 14 = spare0 + col1
// (clamped), 15 = the E*F product.
func finalFetch(r *combRegs, b uint32, ef [4]float32) [4]float32 {
	var v [4]float32
	switch b & 0xF {
	case 14:
		for i := range v {
			v[i] = clamp01(r.spare[0][i] + r.col1[i])
		}
	case 15:
		v = ef
	default:
		v = r.read(b & 0xF)
	}
	if b>>4&1 == 1 {
		v = [4]float32{v[3], v[3], v[3], v[3]}
	}
	// The final combiner supports only the unsigned mappings.
	if b>>5&7 == 1 {
		return [4]float32{1 - clamp01(v[0]), 1 - clamp01(v[1]), 1 - clamp01(v[2]), 1 - clamp01(v[3])}
	}
	return [4]float32{clamp01(v[0]), clamp01(v[1]), clamp01(v[2]), clamp01(v[3])}
}

// combOp applies the output scale/bias, clamping into the signed register range.
func combOp(v [4]float32, op uint32) [4]float32 {
	var out [4]float32
	for i, x := range v {
		switch op {
		case 1: // BIAS_BY_NEGATIVE_ONE_HALF
			x -= 0.5
		case 2: // SHIFT_LEFT_1
			x *= 2
		case 3: // SHIFT_LEFT_1 with bias
			x = (x - 0.5) * 2
		case 4: // SHIFT_LEFT_2
			x *= 4
		case 6: // SHIFT_RIGHT_1
			x *= 0.5
		}
		if x > 1 {
			x = 1
		} else if x < -1 {
			x = -1
		}
		out[i] = x
	}
	return out
}

func mul4(a, b [4]float32) [4]float32 {
	return [4]float32{a[0] * b[0], a[1] * b[1], a[2] * b[2], a[3] * b[3]}
}
func add4(a, b [4]float32) [4]float32 {
	return [4]float32{a[0] + b[0], a[1] + b[1], a[2] + b[2], a[3] + b[3]}
}
func clamp01(x float32) float32 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
