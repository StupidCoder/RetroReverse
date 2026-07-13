package n3ds

// gpu_light.go is the PICA200's fragment-lighting unit: the stage that turns a
// per-vertex normal and a set of light sources into the "primary" and
// "secondary" fragment colours the TEV stages then combine. Captain Toad's
// scene geometry carries *black* vertex colours — every pixel of its world is
// coloured here, which is why the stage is not optional: without it the game
// rasterises a complete, correctly-depth-sorted, entirely black frame.
//
// The unit is configured by three groups of registers, and the model implements
// exactly what the game's own command lists exercise (dumped with -gputrace):
//
//	0x140-0x17F  eight light sources, sixteen registers each: colours
//	             (specular0/1, diffuse, ambient), position, spot direction,
//	             a config word, and a distance-attenuation bias/scale
//	0x1C0-0x1D9  the global configuration: ambient colour, light count, the
//	             per-slot light permutation, and the lookup-table plumbing
//	0x1C5-0x1D2  seven 256-entry lookup tables (uploaded through an index/data
//	             FIFO) that supply the non-linear factors: the specular
//	             distribution terms D0/D1, the Fresnel term, the reflection
//	             terms RR/RG/RB, and the spotlight/distance attenuation curves
//
// The normal does NOT arrive as a vector. The vertex shader emits a *tangent-
// space quaternion* (output semantics 0x04-0x07); the fragment normal is that
// quaternion applied to (0,0,1). That is the PICA's whole normal pipeline, and
// getting it wrong produces a lit-but-wrong frame rather than a black one.

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
	lightSpotXY    = 0x6
	lightSpotZ     = 0x7
	lightConfig    = 0x9
	lightAttenBias = 0xA
	lightAttenScl  = 0xB

	regLightAmbient  = 0x1C0 // global ambient colour
	regLightNumLight = 0x1C2 // bits 0-2: (light count - 1)
	regLightConfig0  = 0x1C3
	regLightConfig1  = 0x1C4
	regLightLUTIndex = 0x1C5
	regLightLUTData  = 0x1C6 // 0x1C6-0x1CD: the upload FIFO
	regLightLUTAbs   = 0x1D0
	regLightLUTSel   = 0x1D1
	regLightLUTScale = 0x1D2
	regLightPermute  = 0x1D9 // slot i → light index (4 bits each)
)

// The seven lookup tables, in the order the LUT-index register selects them.
const (
	lutD0 = 0
	lutD1 = 1
	lutSP = 2 // spotlight attenuation (8 tables, one per light)
	lutFR = 3
	lutRB = 4
	lutRG = 5
	lutRR = 6
	numLUT
)

// lightState is the lighting configuration resolved once per draw.
type lightState struct {
	enabled bool
	count   int
	ambient [3]float32
	lights  [8]lightSource
}

type lightSource struct {
	diffuse   [3]float32
	ambient   [3]float32
	pos       [3]float32
	direction bool // a directional light: pos is the direction, not a point
	twoSided  bool
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

// rgb10 decodes one of the light colour registers: three 10-bit channels,
// blue in the low bits.
func rgb10(v uint32) [3]float32 {
	const s = 1.0 / 255.0 // the 10-bit fields carry an 8-bit colour scaled ×4
	b := float32(v&0x3FF) / 4 * s
	g := float32(v>>10&0x3FF) / 4 * s
	r := float32(v>>20&0x3FF) / 4 * s
	return [3]float32{r, g, b}
}

// lightstate resolves the lighting registers for the current draw.
func (g *GPU) lightstate() lightState {
	ls := lightState{
		enabled: g.Regs[regLightingEnable]&1 != 0,
		count:   int(g.Regs[regLightNumLight]&7) + 1,
		ambient: rgb10(g.Regs[regLightAmbient]),
	}
	perm := g.Regs[regLightPermute]
	for slot := 0; slot < ls.count; slot++ {
		li := perm >> (4 * uint(slot)) & 0xF
		base := uint32(regLightBase) + li*regLightStep
		cfg := g.Regs[base+lightConfig]
		ls.lights[slot] = lightSource{
			diffuse: rgb10(g.Regs[base+lightDiffuse]),
			ambient: rgb10(g.Regs[base+lightAmbient]),
			pos: [3]float32{
				f16(g.Regs[base+lightPosXY]),
				f16(g.Regs[base+lightPosXY] >> 16),
				f16(g.Regs[base+lightPosZ]),
			},
			direction: cfg&1 != 0,
			twoSided:  cfg&2 != 0,
		}
	}
	return ls
}

// fragmentLight computes the primary fragment colour for one fragment: the
// global ambient plus, for every enabled light, its ambient term and its
// diffuse term weighted by N·L. The normal comes from the interpolated
// tangent-space quaternion; the view vector positions the light.
//
// NOT modelled (and the code says so rather than pretending): the specular
// terms and the lookup tables that shape them (D0/D1, Fresnel, the reflection
// tables), distance and spotlight attenuation. Captain Toad's world reads as a
// diffuse-lit scene; specular would brighten highlights, not create the image.
func (ls *lightState) primary(normal, view [3]float32) [3]float32 {
	out := ls.ambient
	for i := 0; i < ls.count; i++ {
		l := &ls.lights[i]
		// A positional light's direction is the vector from the fragment (the
		// view vector is the fragment→eye vector in eye space) to the light.
		var ldir [3]float32
		if l.direction {
			ldir = l.pos
		} else {
			ldir = [3]float32{l.pos[0] + view[0], l.pos[1] + view[1], l.pos[2] + view[2]}
		}
		ldir = normalize3(ldir)

		ndl := normal[0]*ldir[0] + normal[1]*ldir[1] + normal[2]*ldir[2]
		if l.twoSided {
			ndl = absf(ndl)
		} else if ndl < 0 {
			ndl = 0
		}
		for c := 0; c < 3; c++ {
			out[c] += l.ambient[c] + l.diffuse[c]*ndl
		}
	}
	for c := 0; c < 3; c++ {
		if out[c] > 1 {
			out[c] = 1
		}
		if out[c] < 0 {
			out[c] = 0
		}
	}
	return out
}

// quatNormal applies a tangent-space quaternion to (0,0,1) — the PICA's normal
// pipeline. Expanding q ⊗ (0,0,1) ⊗ q* gives the three components directly.
func quatNormal(q [4]float32) [3]float32 {
	x, y, z, w := q[0], q[1], q[2], q[3]
	return normalize3([3]float32{
		2 * (x*z + y*w),
		2 * (y*z - x*w),
		1 - 2*(x*x+y*y),
	})
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

// dumpLighting prints the lighting register block for one draw — the instrument
// that showed which of the unit's features the game actually turns on.
func (g *GPU) dumpLighting() {
	ls := g.lightstate()
	fmt.Printf("  lighting: on count=%d ambient=%v cfg0=%08X cfg1=%08X perm=%08X lutsel=%08X\n",
		ls.count, ls.ambient, g.Regs[regLightConfig0], g.Regs[regLightConfig1],
		g.Regs[regLightPermute], g.Regs[regLightLUTSel])
	for i := 0; i < ls.count; i++ {
		l := ls.lights[i]
		fmt.Printf("    light%d diffuse=%v ambient=%v pos=%v directional=%v twoSided=%v\n",
			i, l.diffuse, l.ambient, l.pos, l.direction, l.twoSided)
	}
	fmt.Printf("      --- tail of subroutine 1 (34+145=179): 170..180\n")
	for i := 170; i < 181; i++ {
		fmt.Printf("        %03X: %s\n", i, ShaderDisasm(g.Code[i], &g.Opdesc))
	}
}
