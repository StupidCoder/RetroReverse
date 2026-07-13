package n3ds

// gpu_light.go is the PICA200's fragment-lighting unit: the stage that turns a
// per-fragment normal and a set of light sources into the "primary" and
// "secondary" fragment colours the TEV stages then combine (TEV sources 1 and
// 2). It is not the vertex colour — that is TEV source 0, and keeping the two
// apart is the whole point of the split between this file and gpu_tev.go.
//
// The unit is configured by three groups of registers:
//
//	0x140-0x17F  eight light sources, sixteen registers each: colours
//	             (specular0/1, diffuse, ambient), position, inverse spot
//	             direction, a config word, and a distance-attenuation
//	             bias/scale
//	0x1C0-0x1D9  the global configuration: ambient colour, light count, the two
//	             config words, the lookup-table plumbing, and the slot → light
//	             permutation
//	0x1C5+0x1C8  the lookup-table upload: an index/type register and an
//	             eight-register data FIFO. Twenty-four tables live behind it —
//	             two specular distributions, a Fresnel term, three reflection
//	             terms, and a spotlight and a distance-attenuation curve *per
//	             light*. The tables are not decoration: Captain Toad's two
//	             positional lights carry all their falloff in them, and without
//	             the distance table they light the whole diorama at full
//	             strength from a metre away.
//
// The normal does NOT arrive as a vector. The vertex shader emits a tangent-
// space quaternion (output semantics 0x04-0x07); the surface normal is a
// *texture-space* vector — (0,0,1) flat, or a normal map sampled from one of
// the texture units when config0's bump mode says so — rotated by that
// quaternion. Captain Toad bump-maps its world from texture unit 2.

import (
	"fmt"
	"math"
)

const (
	regLightBase = 0x140 // 8 lights × 0x10 registers
	regLightStep = 0x10

	// Per-light register offsets from its base.
	lightSpecular0 = 0x0
	lightSpecular1 = 0x1
	lightDiffuse   = 0x2
	lightAmbient   = 0x3
	lightPosXY     = 0x4 // two float16s: x, y
	lightPosZ      = 0x5 // one float16: z
	lightSpotXY    = 0x6 // two fixed1.1.11s: the *inverse* spot direction
	lightSpotZ     = 0x7
	lightConfig    = 0x9
	lightAttenBias = 0xA // float1.7.12
	lightAttenScl  = 0xB

	regLightAmbient  = 0x1C0 // global ambient colour
	regLightNumLight = 0x1C2 // bits 0-2: (light count - 1)
	regLightConfig0  = 0x1C3
	regLightConfig1  = 0x1C4
	regLightLUTIndex = 0x1C5 // bits 0-7 entry index, bits 8-12 table id
	regLightLUTData  = 0x1C8 // 0x1C8-0x1CF: the upload FIFO
	regLightLUTAbs   = 0x1D0
	regLightLUTSel   = 0x1D1
	regLightLUTScale = 0x1D2
	regLightPermute  = 0x1D9 // slot i → light index (3 bits each, nibble-spaced)
)

// The lookup tables, numbered as the LUT-index register's table field numbers
// them. Note the gaps and the per-light banks: the spotlight and distance
// tables are eight tables each, one per light source.
const (
	lutD0  = 0  // specular distribution 0
	lutD1  = 1  // specular distribution 1
	lutFR  = 3  // Fresnel
	lutRB  = 4  // reflection, blue
	lutRG  = 5  // reflection, green
	lutRR  = 6  // reflection, red
	lutSP0 = 8  // 8-15: spotlight attenuation, one per light
	lutDA0 = 16 // 16-23: distance attenuation, one per light
	numLUT = 24
)

func lutName(t int) string {
	switch {
	case t == lutD0:
		return "D0"
	case t == lutD1:
		return "D1"
	case t == lutFR:
		return "FR"
	case t == lutRB:
		return "RB"
	case t == lutRG:
		return "RG"
	case t == lutRR:
		return "RR"
	case t >= lutSP0 && t < lutSP0+8:
		return fmt.Sprintf("SP%d", t-lutSP0)
	case t >= lutDA0 && t < lutDA0+8:
		return fmt.Sprintf("DA%d", t-lutDA0)
	}
	return "?"
}

// The LUT input selectors (0x1D1): which angle cosine indexes a table.
const (
	lutInNH = 0 // normal · half
	lutInVH = 1 // view · half
	lutInNV = 2 // normal · view
	lutInLN = 3 // light · normal
	lutInSP = 4 // light · inverse spot direction
	lutInCP = 5 // tangent · projected half
)

// lutWrite pushes one word into the lookup-table upload FIFO (0x1C8-0x1CF).
// Each word is a table *entry*: a 12-bit value and the 12-bit signed difference
// to the next entry, so the table interpolates between its 256 samples rather
// than stepping. The index register auto-increments, and any of the eight data
// registers feeds the same FIFO.
func (g *GPU) lutWrite(v uint32) {
	t, i := int(g.lutType), int(g.lutIdx)
	if t >= numLUT || i >= 256 {
		return
	}
	g.LUT[t][i] = float32(v&0xFFF) / 4095
	g.LUTDiff[t][i] = float32(int32(v<<8)>>20) / 4095 // sign-extend bits 12-23
	g.lutSet[t] = true
	g.lutIdx = (g.lutIdx + 1) & 0xFF
}

// lutLookup samples a table at a fractional index, using the stored per-entry
// difference as the interpolation slope.
//
// The result is clamped to the range a table entry can actually hold. Entries
// are unsigned 12-bit fractions, so a table cannot represent a value outside
// [0,1] and interpolating *between* two of them cannot leave it either — but
// the last entry has no successor, and its slope field is whatever the game
// happened to leave there. Captain Toad leaves -2048 in it, and a light whose
// attenuation lookup lands exactly on the end of the table (which both of its
// positional lights do, being far outside their own falloff range) came back
// with an attenuation of **-0.5**. A negative attenuation does not dim a light,
// it inverts it: the two fill lights were *subtracting* their warm diffuse from
// every surface they should merely have failed to reach, which is most of why
// the shadows came out too dark and too blue.
func (g *GPU) lutLookup(t int, idx uint8, delta float32) float32 {
	if t >= numLUT {
		return 0
	}
	return clampf(g.LUT[t][idx]+g.LUTDiff[t][idx]*delta, 0, 1)
}

// lightState is the lighting configuration resolved once per draw.
type lightState struct {
	enabled bool
	count   int
	ambient [3]float32
	lights  [8]lightSource

	// config0 (0x1C3).
	env            uint32 // light environment: which LUTs the unit consults at all
	shadow         bool   // shadow attenuation sampled from a texture unit
	shadowSel      int    // which texture unit carries it
	shadowInvert   bool
	shadowPrimary  bool
	shadowSecond   bool
	shadowAlpha    bool
	bumpMode       uint32 // 0 none, 1 normal map, 2 tangent map
	bumpSel        int    // which texture unit carries the map
	noBumpRenorm   bool
	clampHighlight bool
	primaryAlpha   bool // the Fresnel term drives the primary colour's alpha
	secondAlpha    bool

	// The LUT plumbing (0x1D0-0x1D2), and config1's per-table kill switches.
	lutIn, lutAbs, lutScale uint32
	noD0, noD1, noFR        bool
	noRR, noRG, noRB        bool
}

type lightSource struct {
	index                 int // which of the eight hardware lights this slot names
	specular0, specular1  [3]float32
	diffuse, ambient      [3]float32
	pos                   [3]float32
	spotDir               [3]float32
	directional           bool // pos is a direction, not a point
	twoSided              bool
	geo0, geo1            bool // the geometric factor scales specular 0 / 1
	attenBias, attenScale float32
	distAtten, spotAtten  bool // this light consults its distance / spot table
	shadowed              bool
}

// f16 decodes a 16-bit half float (the format the light position registers use).
func f16(v uint32) float32 {
	v &= 0xFFFF
	sign := float32(1)
	if v&0x8000 != 0 {
		sign = -1
	}
	exp := int(v>>10) & 0x1F
	man := int(v & 0x3FF)
	switch exp {
	case 0:
		if man == 0 {
			return sign * 0
		}
		return sign * float32(man) / 1024 / 16384 // subnormal: 2^-14 × man/1024
	case 0x1F:
		return sign * float32(1e30) // no Inf/NaN in the pipeline; saturate
	}
	f := float32(1) + float32(man)/1024
	for e := exp - 15; e > 0; e-- {
		f *= 2
	}
	for e := exp - 15; e < 0; e++ {
		f /= 2
	}
	return sign * f
}

// f20 decodes the distance-attenuation bias/scale format: a 20-bit float with a
// 7-bit exponent (bias 63) and a 12-bit mantissa — the same shape as the
// float24 the rest of the pipeline uses, with a shorter mantissa.
func f20(v uint32) float32 {
	v &= 0xFFFFF
	exp := int(v >> 12 & 0x7F)
	man := v & 0xFFF
	if exp == 0 && man == 0 {
		return 0
	}
	f := float32(math.Ldexp(float64(1)+float64(man)/4096, exp-63))
	if v>>19&1 != 0 {
		f = -f
	}
	return f
}

// fixed11 decodes a fixed1.1.11 spot-direction component: a signed 13-bit
// value scaled so 2047 is 1.0.
func fixed11(v uint32) float32 {
	return float32(int32(v<<19)>>19) / 2047
}

// rgb10 decodes one of the light colour registers: three channels in 10-bit
// fields, blue in the low bits. The fields are ten bits wide but 255 — not
// 1023 — is 1.0: the extra headroom lets a light be configured brighter than
// full white, and a decode that scales by the field width instead of by 255
// comes out four times too dark. (It did. Every light in Captain Toad's stage,
// and the global ambient with them, which is most of why its stone read black.)
func rgb10(v uint32) [3]float32 {
	const s = 1.0 / 255.0
	return [3]float32{
		float32(v>>20&0x3FF) * s,
		float32(v>>10&0x3FF) * s,
		float32(v&0x3FF) * s,
	}
}

// lightstate resolves the lighting registers for the current draw.
func (g *GPU) lightstate() lightState {
	c0, c1 := g.Regs[regLightConfig0], g.Regs[regLightConfig1]
	ls := lightState{
		enabled: g.Regs[regLightingEnable]&1 != 0,
		count:   int(g.Regs[regLightNumLight]&7) + 1,
		ambient: rgb10(g.Regs[regLightAmbient]),

		env:            c0 >> 4 & 0xF,
		shadow:         c0&1 != 0,
		shadowSel:      int(c0 >> 24 & 3),
		shadowInvert:   c0>>18&1 != 0,
		shadowPrimary:  c0>>16&1 != 0,
		shadowSecond:   c0>>17&1 != 0,
		shadowAlpha:    c0>>19&1 != 0,
		bumpMode:       c0 >> 28 & 3,
		bumpSel:        int(c0 >> 22 & 3),
		noBumpRenorm:   c0>>30&1 != 0,
		clampHighlight: c0>>27&1 != 0,
		primaryAlpha:   c0>>2&1 != 0,
		secondAlpha:    c0>>3&1 != 0,

		lutIn:    g.Regs[regLightLUTSel],
		lutAbs:   g.Regs[regLightLUTAbs],
		lutScale: g.Regs[regLightLUTScale],
		noD0:     c1>>16&1 != 0,
		noD1:     c1>>17&1 != 0,
		noFR:     c1>>19&1 != 0,
		noRR:     c1>>20&1 != 0,
		noRG:     c1>>21&1 != 0,
		noRB:     c1>>22&1 != 0,
	}
	perm := g.Regs[regLightPermute]
	for slot := 0; slot < ls.count; slot++ {
		li := perm >> (4 * uint(slot)) & 0x7
		base := uint32(regLightBase) + li*regLightStep
		cfg := g.Regs[base+lightConfig]
		ls.lights[slot] = lightSource{
			index:     int(li),
			specular0: rgb10(g.Regs[base+lightSpecular0]),
			specular1: rgb10(g.Regs[base+lightSpecular1]),
			diffuse:   rgb10(g.Regs[base+lightDiffuse]),
			ambient:   rgb10(g.Regs[base+lightAmbient]),
			pos: [3]float32{
				f16(g.Regs[base+lightPosXY]),
				f16(g.Regs[base+lightPosXY] >> 16),
				f16(g.Regs[base+lightPosZ]),
			},
			spotDir: [3]float32{
				fixed11(g.Regs[base+lightSpotXY]),
				fixed11(g.Regs[base+lightSpotXY] >> 16),
				fixed11(g.Regs[base+lightSpotZ]),
			},
			directional: cfg&1 != 0,
			twoSided:    cfg&2 != 0,
			geo0:        cfg&4 != 0,
			geo1:        cfg&8 != 0,
			attenBias:   f20(g.Regs[base+lightAttenBias]),
			attenScale:  f20(g.Regs[base+lightAttenScl]),
			// The three per-light kill switches in config1 are indexed by the
			// *light*, not by the slot.
			shadowed:  c1>>li&1 == 0,
			spotAtten: c1>>(8+li)&1 == 0,
			distAtten: c1>>(24+li)&1 == 0,
		}
	}
	return ls
}

// lutSupported reports whether the light environment (config0 bits 4-7) routes
// a given table into the pipeline at all. The environments are presets: they
// trade tables for cycles, and a table the environment does not name is not
// consulted no matter what its enable bit says.
func lutSupported(env uint32, t int) bool {
	switch t {
	case lutD0:
		return env != 1
	case lutD1:
		return env != 0 && env != 1 && env != 5
	case lutFR:
		return env != 0 && env != 2 && env != 4
	case lutRR:
		return env != 3
	case lutRG, lutRB:
		return env == 4 || env == 5 || env == 8
	}
	if t >= lutSP0 && t < lutSP0+8 { // spotlight
		return env != 2 && env != 3
	}
	return true // the distance tables are always available
}

// lutScaleOf decodes one of the 3-bit scale fields (0x1D2).
func lutScaleOf(v uint32) float32 {
	switch v & 7 {
	case 0:
		return 1
	case 1:
		return 2
	case 2:
		return 4
	case 3:
		return 8
	case 6:
		return 0.25
	case 7:
		return 0.5
	}
	return 0
}

// shade runs the fragment-lighting unit for one fragment and returns the
// primary (diffuse) and secondary (specular) fragment colours — TEV sources 1
// and 2. tex carries the texture units' sampled colours, because two of the
// unit's inputs come from them: the bump map that supplies the surface normal,
// and the shadow attenuation.
func (g *GPU) shade(ls *lightState, quat [4]float32, view [3]float32, tex *[3]rgba) (prim, sec [4]float32) {
	// The shadow attenuation is a texture read. That is the whole mechanism:
	// there is no separate shadow buffer in the fragment pipeline, and no
	// stencil involved — the shadow pass renders into a render target, the
	// target is bound as a texture, and the lighting unit multiplies the light's
	// contribution by what it samples there.
	shadow := [4]float32{1, 1, 1, 1}
	if ls.shadow {
		s := tex[clampi(int32(ls.shadowSel), 0, 2)]
		shadow = [4]float32{float32(s.r) / 255, float32(s.g) / 255, float32(s.b) / 255, float32(s.a) / 255}
		if ls.shadowInvert {
			for i := range shadow {
				shadow[i] = 1 - shadow[i]
			}
		}
	}

	// The surface normal and tangent, in texture space, before the quaternion
	// rotates them into eye space.
	sn := [3]float32{0, 0, 1}
	st := [3]float32{1, 0, 0}
	if ls.bumpMode != 0 {
		t := tex[clampi(int32(ls.bumpSel), 0, 2)]
		p := [3]float32{
			float32(t.r)/127.5 - 1,
			float32(t.g)/127.5 - 1,
			float32(t.b)/127.5 - 1,
		}
		if ls.bumpMode == 1 { // normal map
			if !ls.noBumpRenorm {
				// The map stores only x and y; z is recovered on the assumption
				// that the vector is unit length.
				z := 1 - (p[0]*p[0] + p[1]*p[1])
				if z < 0 {
					z = 0
				}
				p[2] = sqrt32(z)
			}
			sn = p
		} else { // tangent map
			st = p
		}
	}
	normal := quatRotate(quat, sn)
	tangent := quatRotate(quat, st)

	nview := normalize3(view)

	prim = [4]float32{0, 0, 0, 1}
	sec = [4]float32{0, 0, 0, 1}

	for i := 0; i < ls.count; i++ {
		l := &ls.lights[i]

		ldir := l.pos
		if !l.directional {
			ldir = [3]float32{l.pos[0] + view[0], l.pos[1] + view[1], l.pos[2] + view[2]}
		}
		dist := sqrt32(ldir[0]*ldir[0] + ldir[1]*ldir[1] + ldir[2]*ldir[2])
		ldir = normalize3(ldir)
		half := [3]float32{nview[0] + ldir[0], nview[1] + ldir[1], nview[2] + ldir[2]}

		// Distance attenuation: a per-light table, indexed by the light's own
		// affine remap of the distance. The two positional lights in Captain
		// Toad's stage carry their entire falloff here.
		distAtten := float32(1)
		if l.distAtten {
			loc := clampf(l.attenScale*dist+l.attenBias, 0, 1)
			idx, delta := lutIndexAbs(loc)
			distAtten = g.lutLookup(lutDA0+l.index, idx, delta)
		}

		// The remaining tables share one indexing rule: pick an input cosine,
		// optionally take its absolute value, sample, scale.
		lut := func(field uint, t int) float32 {
			in := ls.lutIn >> field & 7
			abs := ls.lutAbs>>(field+1)&1 == 0
			var c float32
			switch in {
			case lutInNH:
				c = dot3(normal, normalize3(half))
			case lutInVH:
				c = dot3(nview, normalize3(half))
			case lutInNV:
				c = dot3(normal, nview)
			case lutInLN:
				c = dot3(ldir, normal)
			case lutInSP:
				c = dot3(ldir, l.spotDir)
			case lutInCP:
				if ls.env == 8 {
					nh := normalize3(half)
					d := dot3(normal, nh)
					proj := [3]float32{nh[0] - normal[0]*d, nh[1] - normal[1]*d, nh[2] - normal[2]*d}
					c = dot3(proj, tangent)
				}
			}
			var idx uint8
			var delta float32
			if abs {
				// An absolute-indexed table covers [0,1] across all 256 entries.
				if l.twoSided {
					c = absf(c)
				} else if c < 0 {
					c = 0
				}
				idx, delta = lutIndexAbs(c)
			} else {
				// A signed table covers [-1,1): the negative half lives in the
				// table's upper entries, which is exactly what a two's-complement
				// index does for free.
				f := floor32(c * 128)
				si := int32(clampf(f, -128, 127))
				delta = c*128 - float32(si)
				idx = uint8(si)
			}
			return lutScaleOf(ls.lutScale>>field) * g.lutLookup(t, idx, delta)
		}

		spotAtten := float32(1)
		if l.spotAtten && lutSupported(ls.env, lutSP0) {
			spotAtten = lut(8, lutSP0+l.index)
		}

		// Specular 0: the distribution table shapes the highlight.
		d0 := float32(1)
		if !ls.noD0 && lutSupported(ls.env, lutD0) {
			d0 = lut(0, lutD0)
		}
		spec0 := scale3(l.specular0, d0)

		// The reflection tables tint specular 1; green and blue fall back to red.
		refl := [3]float32{1, 1, 1}
		if !ls.noRR && lutSupported(ls.env, lutRR) {
			refl[0] = lut(24, lutRR)
		}
		refl[1], refl[2] = refl[0], refl[0]
		if !ls.noRG && lutSupported(ls.env, lutRG) {
			refl[1] = lut(20, lutRG)
		}
		if !ls.noRB && lutSupported(ls.env, lutRB) {
			refl[2] = lut(16, lutRB)
		}
		d1 := float32(1)
		if !ls.noD1 && lutSupported(ls.env, lutD1) {
			d1 = lut(4, lutD1)
		}
		spec1 := [3]float32{
			d1 * refl[0] * l.specular1[0],
			d1 * refl[1] * l.specular1[1],
			d1 * refl[2] * l.specular1[2],
		}

		// Fresnel, which only the last light slot applies, and which lands in an
		// alpha channel rather than a colour.
		if i == ls.count-1 && !ls.noFR && lutSupported(ls.env, lutFR) {
			fr := lut(12, lutFR)
			if ls.primaryAlpha {
				prim[3] = fr
			}
			if ls.secondAlpha {
				sec[3] = fr
			}
		}

		ndl := dot3(ldir, normal)
		if l.twoSided {
			ndl = absf(ndl)
		} else if ndl < 0 {
			ndl = 0
		}
		clampHL := float32(1)
		if ls.clampHighlight && ndl == 0 {
			clampHL = 0
		}
		if l.geo0 || l.geo1 {
			// The geometric factor divides by |H|², which is the standard
			// attenuation of a half-vector highlight at grazing angles.
			geo := dot3(half, half)
			if geo == 0 {
				geo = 0
			} else {
				geo = minf(ndl/geo, 1)
			}
			if l.geo0 {
				spec0 = scale3(spec0, geo)
			}
			if l.geo1 {
				spec1 = scale3(spec1, geo)
			}
		}

		shPri, shSec := [3]float32{1, 1, 1}, [3]float32{1, 1, 1}
		if ls.shadowPrimary && l.shadowed {
			shPri = [3]float32{shadow[0], shadow[1], shadow[2]}
		}
		if ls.shadowSecond && l.shadowed {
			shSec = [3]float32{shadow[0], shadow[1], shadow[2]}
		}

		att := distAtten * spotAtten
		for c := 0; c < 3; c++ {
			prim[c] += (l.diffuse[c]*ndl*shPri[c] + l.ambient[c]) * att
			sec[c] += (spec0[c] + spec1[c]) * clampHL * att * shSec[c]
		}
	}

	if ls.shadowAlpha {
		if ls.primaryAlpha {
			prim[3] *= shadow[3]
		}
		if ls.secondAlpha {
			sec[3] *= shadow[3]
		}
	}
	for c := 0; c < 3; c++ {
		prim[c] = clampf(prim[c]+ls.ambient[c], 0, 1)
		sec[c] = clampf(sec[c], 0, 1)
	}
	prim[3] = clampf(prim[3], 0, 1)
	sec[3] = clampf(sec[3], 0, 1)
	return prim, sec
}

// lutIndexAbs splits a [0,1] table coordinate into an entry index and the
// fraction between it and the next entry.
func lutIndexAbs(v float32) (uint8, float32) {
	v = clampf(v, 0, 1)
	f := clampf(floor32(v*256), 0, 255)
	return uint8(f), v*256 - f
}

// quatRotate applies a tangent-space quaternion to a vector — the PICA's normal
// pipeline. The quaternion is normalised first: it arrives interpolated across
// the triangle, and interpolation does not preserve unit length.
func quatRotate(q [4]float32, v [3]float32) [3]float32 {
	l := sqrt32(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])
	if l == 0 {
		return v
	}
	x, y, z, w := q[0]/l, q[1]/l, q[2]/l, q[3]/l
	// v + 2·q.xyz × (q.xyz × v + w·v)
	tx := 2 * (y*v[2] - z*v[1])
	ty := 2 * (z*v[0] - x*v[2])
	tz := 2 * (x*v[1] - y*v[0])
	return [3]float32{
		v[0] + w*tx + (y*tz - z*ty),
		v[1] + w*ty + (z*tx - x*tz),
		v[2] + w*tz + (x*ty - y*tx),
	}
}

func dot3(a, b [3]float32) float32 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }

func scale3(v [3]float32, s float32) [3]float32 {
	return [3]float32{v[0] * s, v[1] * s, v[2] * s}
}

func normalize3(v [3]float32) [3]float32 {
	l := v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
	if l <= 0 {
		return [3]float32{0, 0, 1}
	}
	l = sqrt32(l)
	return [3]float32{v[0] / l, v[1] / l, v[2] / l}
}

func sqrt32(f float32) float32 { return float32(math.Sqrt(float64(f))) }

func absf(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func clampf(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// dumpLighting prints the lighting register block for one draw — raw words and
// decoded fields side by side, because every bug this unit has had was a field
// nobody had actually read.
func (g *GPU) dumpLighting() {
	c0, c1 := g.Regs[regLightConfig0], g.Regs[regLightConfig1]
	fmt.Printf("  lighting: on count=%d ambient=%v (raw %08X)\n",
		int(g.Regs[regLightNumLight]&7)+1, rgb10(g.Regs[regLightAmbient]), g.Regs[regLightAmbient])
	fmt.Printf("    cfg0=%08X shadow=%v env=%d shPri=%v shSec=%v shInv=%v shAlpha=%v shTex=%d bumpTex=%d bumpMode=%d noRenorm=%v clampHL=%v priA=%v secA=%v\n",
		c0, c0&1 != 0, c0>>4&0xF, c0>>16&1 != 0, c0>>17&1 != 0, c0>>18&1 != 0, c0>>19&1 != 0,
		c0>>24&3, c0>>22&3, c0>>28&3, c0>>30&1 != 0, c0>>27&1 != 0, c0>>2&1 != 0, c0>>3&1 != 0)
	fmt.Printf("    cfg1=%08X noShadow=%02X noSpot=%02X noDist=%02X lutOff(d0=%v d1=%v fr=%v rr=%v rg=%v rb=%v)\n",
		c1, c1&0xFF, c1>>8&0xFF, c1>>24&0xFF,
		c1>>16&1 != 0, c1>>17&1 != 0, c1>>19&1 != 0, c1>>20&1 != 0, c1>>21&1 != 0, c1>>22&1 != 0)
	fmt.Printf("    perm=%08X lutsel=%08X lutabs=%08X lutscale=%08X\n",
		g.Regs[regLightPermute], g.Regs[regLightLUTSel], g.Regs[regLightLUTAbs], g.Regs[regLightLUTScale])
	ls := g.lightstate()
	for i := 0; i < ls.count; i++ {
		l := ls.lights[i]
		base := uint32(regLightBase) + uint32(l.index)*regLightStep
		fmt.Printf("    slot%d=light%d diffuse=%v(%08X) ambient=%v(%08X) spec0=%v spec1=%v\n",
			i, l.index, l.diffuse, g.Regs[base+lightDiffuse], l.ambient, g.Regs[base+lightAmbient],
			l.specular0, l.specular1)
		fmt.Printf("           pos=%v dir=%v twoSided=%v geo=%v/%v atten=%v(bias=%.4f scale=%.6f) spot=%v(%v) shadowed=%v\n",
			l.pos, l.directional, l.twoSided, l.geo0, l.geo1, l.distAtten, l.attenBias, l.attenScale,
			l.spotAtten, l.spotDir, l.shadowed)
	}
	for t := 0; t < numLUT; t++ {
		if !g.lutSet[t] {
			continue
		}
		fmt.Printf("    lut%-2d %-4s: [0]=%.3f [32]=%.3f [64]=%.3f [128]=%.3f [192]=%.3f [255]=%.3f\n",
			t, lutName(t), g.LUT[t][0], g.LUT[t][32], g.LUT[t][64], g.LUT[t][128], g.LUT[t][192], g.LUT[t][255])
	}
}

// dumpFragment prints the texture units and the TEV stages for one draw — the
// other half of what a lit fragment is made of, and the half that says what the
// lighting unit's two outputs are actually used *for*.
func (g *GPU) dumpFragment() {
	en := g.Regs[0x080]
	fmt.Printf("  texunits: cfg=%08X (tex0=%v tex1=%v tex2=%v) type=%d\n",
		en, en&1 != 0, en>>1&1 != 0, en>>2&1 != 0, g.Regs[0x083]>>28&7)
	for u := 0; u < 3; u++ {
		if en>>uint(u)&1 == 0 {
			continue
		}
		dimR, paramR, addrR, typR := texUnitRegs(u)
		fmt.Printf("    tex%d %dx%d fmt=%d addr=0x%08X param=%08X\n", u,
			g.Regs[dimR]>>16&0x7FF, g.Regs[dimR]&0x7FF, g.Regs[typR]&0xF,
			g.m.gpuAddrToVirt(g.Regs[addrR]<<3), g.Regs[paramR])
	}
	nstage := 6
	fmt.Printf("  tev: bufupd=%08X bufcol=%08X\n", g.Regs[0x0E0], g.Regs[0x0FD])
	for st := 0; st < nstage; st++ {
		b := tevStageBase[st]
		src, opd, cmb := g.Regs[b], g.Regs[b+1], g.Regs[b+2]
		k := g.Regs[b+3]
		fmt.Printf("    stage%d rgb: %s(%s) %s(%s) %s(%s) op=%s | a: %s(%s) %s(%s) %s(%s) op=%s | scale=%d/%d konst=%08X rgb(%d,%d,%d)\n", st,
			tevSrcName(src&0xF), tevCOpName(opd&0xF), tevSrcName(src>>4&0xF), tevCOpName(opd>>4&0xF),
			tevSrcName(src>>8&0xF), tevCOpName(opd>>8&0xF), tevOpName(cmb&0xF),
			tevSrcName(src>>16&0xF), tevAOpName(opd>>12&7), tevSrcName(src>>20&0xF), tevAOpName(opd>>16&7),
			tevSrcName(src>>24&0xF), tevAOpName(opd>>20&7), tevOpName(cmb>>16&0xF),
			1<<(g.Regs[b+4]&3), 1<<(g.Regs[b+4]>>16&3), k, k&0xFF, k>>8&0xFF, k>>16&0xFF)
	}
}

func tevSrcName(s uint32) string {
	switch s {
	case 0:
		return "vtxcol"
	case 1:
		return "fragpri"
	case 2:
		return "fragsec"
	case 3:
		return "tex0"
	case 4:
		return "tex1"
	case 5:
		return "tex2"
	case 13:
		return "buffer"
	case 14:
		return "konst"
	case 15:
		return "prev"
	}
	return fmt.Sprintf("src%d", s)
}

func tevCOpName(o uint32) string {
	switch o {
	case 0:
		return "rgb"
	case 1:
		return "1-rgb"
	case 2:
		return "a"
	case 3:
		return "1-a"
	case 4:
		return "r"
	case 5:
		return "1-r"
	case 8:
		return "g"
	case 9:
		return "1-g"
	case 12:
		return "b"
	case 13:
		return "1-b"
	}
	return fmt.Sprintf("cop%d", o)
}

func tevAOpName(o uint32) string {
	switch o {
	case 0:
		return "a"
	case 1:
		return "1-a"
	case 2:
		return "r"
	case 3:
		return "1-r"
	case 4:
		return "g"
	case 5:
		return "1-g"
	case 6:
		return "b"
	}
	return "1-b"
}

func tevOpName(o uint32) string {
	switch o {
	case 0:
		return "replace"
	case 1:
		return "modulate"
	case 2:
		return "add"
	case 3:
		return "addsigned"
	case 4:
		return "lerp"
	case 5:
		return "subtract"
	case 8:
		return "muladd"
	case 9:
		return "addmul"
	}
	return fmt.Sprintf("op%d", o)
}
