package n3ds

// gpu_tev.go is the PICA200 fragment stage: sample the enabled texture units
// for the pixel, run the six texture-environment (TEV) combiner stages, apply
// the alpha test, and hand the result to the blender. The TEV is the same
// (a op b) combiner family as the N64's colour combiner (n64/rdp_combine.go's
// (a-b)·c+d): each stage picks three sources through operand multiplexers and
// combines them, feeding the next stage through "previous".
//
// Stage registers sit at 0x0C0,0x0C8,0x0D0,0x0D8,0x0F0,0x0F8: +0 sources
// (colour and alpha), +1 operands, +2 combine ops, +3 the stage constant,
// +4 output scale. Unknown sources/ops halt.

var tevStageBase = [6]uint32{0x0C0, 0x0C8, 0x0D0, 0x0D8, 0x0F0, 0x0F8}

type rgba struct{ r, g, b, a int32 } // 0-255 per lane

// fragment computes the pixel colour for one fragment. discard=true means the
// alpha test rejected the pixel; ok=false means the machine halted (unknown
// feature).
//
// The three colour inputs the TEV can select are distinct and are kept distinct
// here — collapsing them is what made Captain Toad's stone black. Source 0 is
// the interpolated *vertex* colour; source 1 is the fragment-lighting unit's
// primary (diffuse) output; source 2 is its secondary (specular) output. A
// material that modulates its texture against the vertex colour and *adds* the
// lit colour needs all three to be different values. Texture sampling happens
// first because the lighting unit reads two texture units itself (the bump map
// and the shadow attenuation), which is why it is evaluated here and not in the
// rasteriser.
func (g *GPU) fragment(vcol [4]float32, uv [3][2]float32, uv0w float32, ls *lightState, tv *tevState, quat [4]float32, view [3]float32, st *rstats) (r, gr, b, a uint8, discard, ok bool) {
	vertex := rgba{clamp255(vcol[0] * 255), clamp255(vcol[1] * 255), clamp255(vcol[2] * 255), clamp255(vcol[3] * 255)}

	// Sample the enabled texture units (0x080 bits 0-2).
	var tex [3]rgba
	for u := 0; u < 3; u++ {
		if tv.texEnable>>uint(u)&1 != 0 {
			var oks bool
			tex[u], oks = g.sampleTextureSt(u, uv[u][0], uv[u][1], uv0w, st)
			if !oks {
				return 0, 0, 0, 0, false, false
			}
		}
	}

	// Fragment lighting. Off, the primary colour is the vertex colour passed
	// straight through (that is what the hardware does with the unit disabled)
	// and there is no secondary colour.
	fragPrim, fragSec := vertex, rgba{0, 0, 0, 0}
	if ls.enabled {
		p, sc := g.shade(ls, quat, view, &tex)
		fragPrim = rgba{clamp255(p[0] * 255), clamp255(p[1] * 255), clamp255(p[2] * 255), clamp255(p[3] * 255)}
		fragSec = rgba{clamp255(sc[0] * 255), clamp255(sc[1] * 255), clamp255(sc[2] * 255), clamp255(sc[3] * 255)}
	}

	prev := vertex

	// The combiner buffer is delayed by one stage. Stage 0 reads zero (not the
	// configured buffer colour — that only reaches stage 1), and a stage's
	// contribution to the buffer is not visible to the stage immediately after
	// it, but to the one after that. Only the first four stages can write it at
	// all: the update masks are four bits wide, not six.
	var buf rgba
	next := tv.bufColor

	// fetch resolves a TEV source id. It was a closure, called 36 times per fragment;
	// the sources it selects between are all in hand here, so it is a plain switch.
	fetch := func(sel uint8) (rgba, bool) {
		switch sel {
		case 0: // PrimaryColor: the interpolated vertex colour
			return vertex, true
		case 1: // PrimaryFragmentColor: the lighting unit's diffuse output
			return fragPrim, true
		case 2: // SecondaryFragmentColor: the lighting unit's specular output
			return fragSec, true
		case 3:
			return tex[0], true
		case 4:
			return tex[1], true
		case 5:
			return tex[2], true
		case 13:
			return buf, true
		case 15:
			return prev, true
		}
		return rgba{}, false
	}

	for i := range tv.stages {
		s := &tv.stages[i]

		var cin [3]rgba
		for j := 0; j < 3; j++ {
			o := s.colr[j]
			src, okf := rgba{}, true
			if o.src == 14 { // Constant: decoded per stage, not a shared source
				src = s.konst
			} else if src, okf = fetch(o.src); !okf {
				g.m.CPU.Halt("gpu tev: stage %d source %d unimplemented", i, o.src)
				return 0, 0, 0, 0, false, false
			}
			cin[j] = tevColorOperand(src, uint32(o.op))
		}
		var ain [3]int32
		for j := 0; j < 3; j++ {
			o := s.alph[j]
			src, okf := rgba{}, true
			if o.src == 14 {
				src = s.konst
			} else if src, okf = fetch(o.src); !okf {
				g.m.CPU.Halt("gpu tev: stage %d source %d unimplemented", i, o.src)
				return 0, 0, 0, 0, false, false
			}
			ain[j] = tevAlphaOperand(src, uint32(o.op))
		}

		cr, okc := tevCombine(uint32(s.combC), cin[0].r, cin[1].r, cin[2].r)
		cg, _ := tevCombine(uint32(s.combC), cin[0].g, cin[1].g, cin[2].g)
		cb, _ := tevCombine(uint32(s.combC), cin[0].b, cin[1].b, cin[2].b)
		ca, oka := tevCombine(uint32(s.combA), ain[0], ain[1], ain[2])
		if !okc || !oka {
			g.m.CPU.Halt("gpu tev: stage %d combine op unimplemented", i)
			return 0, 0, 0, 0, false, false
		}

		out := rgba{
			clampi(cr<<s.scaleC, 0, 255),
			clampi(cg<<s.scaleC, 0, 255),
			clampi(cb<<s.scaleC, 0, 255),
			clampi(ca<<s.scaleA, 0, 255),
		}
		buf = next
		if s.updC {
			next.r, next.g, next.b = out.r, out.g, out.b
		}
		if s.updA {
			next.a = out.a
		}
		prev = out
	}

	// Alpha test (0x104): bit0 enable, bits 4-6 func, bits 8-15 reference.
	if tv.alphaTest {
		ref := tv.alphaRef
		var pass bool
		switch tv.alphaFunc {
		case 0:
			pass = false
		case 1:
			pass = true
		case 2:
			pass = prev.a == ref
		case 3:
			pass = prev.a != ref
		case 4:
			pass = prev.a < ref
		case 5:
			pass = prev.a <= ref
		case 6:
			pass = prev.a > ref
		default:
			pass = prev.a >= ref
		}
		if !pass {
			return 0, 0, 0, 0, true, true // discarded
		}
	}
	return uint8(prev.r), uint8(prev.g), uint8(prev.b), uint8(prev.a), false, true
}

// tevColorOperand applies a colour-lane operand multiplexer.
func tevColorOperand(s rgba, op uint32) rgba {
	switch op {
	case 0:
		return s
	case 1:
		return rgba{255 - s.r, 255 - s.g, 255 - s.b, s.a}
	case 2:
		return rgba{s.a, s.a, s.a, s.a}
	case 3:
		return rgba{255 - s.a, 255 - s.a, 255 - s.a, 255 - s.a}
	case 4:
		return rgba{s.r, s.r, s.r, s.r}
	case 5:
		return rgba{255 - s.r, 255 - s.r, 255 - s.r, 255 - s.r}
	case 8:
		return rgba{s.g, s.g, s.g, s.g}
	case 9:
		return rgba{255 - s.g, 255 - s.g, 255 - s.g, 255 - s.g}
	case 12:
		return rgba{s.b, s.b, s.b, s.b}
	case 13:
		return rgba{255 - s.b, 255 - s.b, 255 - s.b, 255 - s.b}
	}
	return s
}

// tevAlphaOperand applies an alpha-lane operand multiplexer.
func tevAlphaOperand(s rgba, op uint32) int32 {
	switch op {
	case 0:
		return s.a
	case 1:
		return 255 - s.a
	case 2:
		return s.r
	case 3:
		return 255 - s.r
	case 4:
		return s.g
	case 5:
		return 255 - s.g
	case 6:
		return s.b
	default:
		return 255 - s.b
	}
}

// tevCombine applies one combine op to one lane (values 0-255).
func tevCombine(op uint32, a, b, c int32) (int32, bool) {
	switch op {
	case 0: // replace
		return a, true
	case 1: // modulate
		return a * b / 255, true
	case 2: // add
		return a + b, true
	case 3: // add signed
		return a + b - 128, true
	case 4: // interpolate: a·c + b·(1-c)
		return (a*c + b*(255-c)) / 255, true
	case 5: // subtract
		return a - b, true
	case 8: // multiply then add: a·b + c
		return a*b/255 + c, true
	case 9: // add then multiply: (a+b)·c
		return clampi(a+b, 0, 255) * c / 255, true
	}
	return 0, false
}

// blend applies the 0x100/0x101 blending config to a fragment against the
// destination pixel.
func (g *GPU) blend(sr, sg, sb, sa, dr, dg, db, da uint8) (uint8, uint8, uint8, uint8) {
	if g.Regs[0x100]>>8&1 == 0 {
		// Logic-op mode; op 3 (copy) is the only one seen. Others: halt.
		if lop := g.Regs[0x102] & 0xF; lop != 3 {
			g.m.CPU.Halt("gpu: logic op %d unimplemented", lop)
			return sr, sg, sb, sa
		}
		return sr, sg, sb, sa
	}
	cfg := g.Regs[0x101]
	eqC, eqA := cfg&7, cfg>>8&7
	fsC, fdC := cfg>>16&0xF, cfg>>20&0xF
	fsA, fdA := cfg>>24&0xF, cfg>>28&0xF

	factor := func(code uint32, s, d, sA, dA int32, isA bool) (int32, int32, int32) {
		cc := g.Regs[0x103]
		cr, cg2, cb2 := int32(cc&0xFF), int32(cc>>8&0xFF), int32(cc>>16&0xFF)
		ca := int32(cc >> 24 & 0xFF)
		switch code {
		case 0:
			return 0, 0, 0
		case 1:
			return 255, 255, 255
		case 2:
			return s, s, s // (per-lane source colour filled by caller)
		case 3:
			return 255 - s, 255 - s, 255 - s
		case 4:
			return d, d, d
		case 5:
			return 255 - d, 255 - d, 255 - d
		case 6:
			return sA, sA, sA
		case 7:
			return 255 - sA, 255 - sA, 255 - sA
		case 8:
			return dA, dA, dA
		case 9:
			return 255 - dA, 255 - dA, 255 - dA
		case 10:
			if isA {
				return ca, ca, ca
			}
			return cr, cg2, cb2
		case 11:
			if isA {
				return 255 - ca, 255 - ca, 255 - ca
			}
			return 255 - cr, 255 - cg2, 255 - cb2
		case 12:
			return ca, ca, ca
		case 13:
			return 255 - ca, 255 - ca, 255 - ca
		case 14: // saturated source alpha
			m := sA
			if 255-dA < m {
				m = 255 - dA
			}
			return m, m, m
		}
		g.m.CPU.Halt("gpu: blend factor %d unimplemented", code)
		return 0, 0, 0
	}

	apply := func(eq uint32, s, d, fs, fd int32) int32 {
		switch eq {
		case 0:
			return clampi((s*fs+d*fd)/255, 0, 255)
		case 1:
			return clampi((s*fs-d*fd)/255, 0, 255)
		case 2:
			return clampi((d*fd-s*fs)/255, 0, 255)
		case 3:
			if d < s {
				return d
			}
			return s
		case 4:
			if d > s {
				return d
			}
			return s
		}
		g.m.CPU.Halt("gpu: blend equation %d unimplemented", eq)
		return s
	}

	sri, sgi, sbi, sai := int32(sr), int32(sg), int32(sb), int32(sa)
	dri, dgi, dbi, dai := int32(dr), int32(dg), int32(db), int32(da)

	// Per-lane factors: the "source/dest colour" codes need the matching lane.
	fsr, _, _ := factor(fsC, sri, dri, sai, dai, false)
	_, fsg, _ := factor(fsC, sgi, dgi, sai, dai, false)
	_, _, fsb := factor(fsC, sbi, dbi, sai, dai, false)
	fdr, _, _ := factor(fdC, sri, dri, sai, dai, false)
	_, fdg, _ := factor(fdC, sgi, dgi, sai, dai, false)
	_, _, fdb := factor(fdC, sbi, dbi, sai, dai, false)
	fsa, _, _ := factor(fsA, sai, dai, sai, dai, true)
	fda, _, _ := factor(fdA, sai, dai, sai, dai, true)

	return uint8(apply(eqC, sri, dri, fsr, fdr)),
		uint8(apply(eqC, sgi, dgi, fsg, fdg)),
		uint8(apply(eqC, sbi, dbi, fsb, fdb)),
		uint8(apply(eqA, sai, dai, fsa, fda))
}

func clamp255(f float32) int32 {
	if f < 0 {
		return 0
	}
	if f > 255 {
		return 255
	}
	return int32(f)
}

func clampi(v, lo, hi int32) int32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
