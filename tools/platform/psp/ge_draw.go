package psp

// ge_draw.go is the software rasterizer: triangle and sprite fill with per-vertex
// colour interpolation, optional texture modulation, and CLEAR-mode solid fills,
// writing pixels into the framebuffer in its PSP pixel format.

import (
	"fmt"
	"math"
)

// rasterTriClipped clips a triangle against the near plane, then rasterizes the
// pieces that survive.
func (m *Machine) rasterTriClipped(s *geState, a, b, c vert) {
	for _, t := range clipTriNear(s, a, b, c) {
		m.rasterTri(s, t[0], t[1], t[2])
	}
}

// rasterTri fills the triangle a,b,c with barycentric-interpolated colour (and
// texcoords when textured). In clear mode it fills solid with a's colour.
func (m *Machine) rasterTri(s *geState, a, b, c vert) {
	minX := int(fmin3(a.x, b.x, c.x))
	maxX := int(fmax3(a.x, b.x, c.x)) + 1
	minY := int(fmin3(a.y, b.y, c.y))
	maxY := int(fmax3(a.y, b.y, c.y)) + 1
	minX, minY = clampI(minX, 0, dispW), clampI(minY, 0, dispH)
	maxX, maxY = clampI(maxX, 0, dispW), clampI(maxY, 0, dispH)

	area := edge(a, b, c)
	if area == 0 {
		return
	}
	// Backface culling. The winding was settled against a PSP_GE_NOCULL frame:
	// with the sign the other way the race lost its ground plane and horizon
	// (front faces were being culled), so with FACE=0 the culled face is the
	// one with negative screen-space area.
	if s.cullOn && !geNoCull && !s.clearOn && !a.through() {
		if ((s.cullFace == 0) == (area < 0)) != geCullFlip {
			return
		}
	}
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			px := vert{x: float32(x) + 0.5, y: float32(y) + 0.5}
			w0 := edge(b, c, px)
			w1 := edge(c, a, px)
			w2 := edge(a, b, px)
			// inside test (either winding)
			if (w0 < 0 || w1 < 0 || w2 < 0) && (w0 > 0 || w1 > 0 || w2 > 0) {
				continue
			}
			l0, l1, l2 := w0/area, w1/area, w2/area
			z := l0*a.z + l1*b.z + l2*c.z
			r := byte(l0*float32(a.r) + l1*float32(b.r) + l2*float32(c.r))
			g := byte(l0*float32(a.g) + l1*float32(b.g) + l2*float32(c.g))
			bl := byte(l0*float32(a.b) + l1*float32(b.b) + l2*float32(c.b))
			al := byte(l0*float32(a.a) + l1*float32(b.a) + l2*float32(c.a))
			if s.texEnable && !s.clearOn {
				// Texcoords interpolate PERSPECTIVE-CORRECTLY: weight each by
				// 1/w, interpolate those linearly in screen space, then divide
				// back. Interpolating u,v directly (affine) is what warps a
				// surface receding to the horizon — the road was swimming.
				u, v := uvAt(s, a, b, c, area, float32(x)+0.5, float32(y)+0.5)
				// rho: how far the texcoords move over one screen pixel, in
				// texels — the quantity the mip level is chosen from. Sampling
				// the neighbouring fragments' texcoords gives it exactly, even
				// under perspective, which matters here: a railing receding to
				// the horizon minifies enormously along its length.
				var rho float32 = 1
				if s.texMaxLvl > 0 {
					ux, vx := uvAt(s, a, b, c, area, float32(x)+1.5, float32(y)+0.5)
					uy, vy := uvAt(s, a, b, c, area, float32(x)+0.5, float32(y)+1.5)
					fw, fh := float32(s.texW), float32(s.texH)
					dx := hypot32((ux-u)*fw, (vx-v)*fh)
					dy := hypot32((uy-u)*fw, (vy-v)*fh)
					rho = dx
					if dy > rho {
						rho = dy
					}
				}
				r, g, bl, al = modTex(m, s, u, v, rho, r, g, bl, al)
			}
			if s.fogOn && a.clip && !s.clearOn {
				r, g, bl = applyFog(s, l0*a.fog+l1*b.fog+l2*c.fog, r, g, bl)
			}
			m.putPixel(s, x, y, z, r, g, bl, al)
		}
	}
}

// rasterSprite fills the axis-aligned rectangle between corners a and b (PSP sprite
// primitive), textured or solid.
func (m *Machine) rasterSprite(s *geState, a, b vert) {
	x0, x1 := int(a.x), int(b.x)
	y0, y1 := int(a.y), int(b.y)
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	u0, u1 := a.u, b.u
	v0, v1 := a.v, b.v
	x0c, y0c := clampI(x0, 0, dispW), clampI(y0, 0, dispH)
	x1c, y1c := clampI(x1, 0, dispW), clampI(y1, 0, dispH)
	dw := float32(x1 - x0)
	dh := float32(y1 - y0)
	for y := y0c; y < y1c; y++ {
		for x := x0c; x < x1c; x++ {
			// A sprite carries the far corner's depth across the whole rect.
			r, g, bl, al := b.r, b.g, b.b, b.a
			if s.texEnable && !s.clearOn && dw != 0 && dh != 0 {
				u := u0 + (u1-u0)*(float32(x-x0)/dw)
				v := v0 + (v1-v0)*(float32(y-y0)/dh)
				r, g, bl, al = modTex(m, s, u, v, 1, r, g, bl, al)
			}
			if s.fogOn && b.clip && !s.clearOn {
				r, g, bl = applyFog(s, b.fog, r, g, bl)
			}
			m.putPixel(s, x, y, b.z, r, g, bl, al)
		}
	}
}

// applyFog mixes the fragment towards the fog colour (0xBBGGRR) by the
// interpolated fog coefficient.
func applyFog(s *geState, f float32, r, g, b byte) (byte, byte, byte) {
	f = clampF(f, 0, 1)
	fr, fg, fb := float32(byte(s.fogColor)), float32(byte(s.fogColor>>8)), float32(byte(s.fogColor>>16))
	mix := func(c, fc float32) byte { return byte(c*f + fc*(1-f)) }
	return mix(float32(r), fr), mix(float32(g), fg), mix(float32(b), fb)
}

// uvAt evaluates the perspective-correct texcoords of a triangle at a screen
// point.
func uvAt(s *geState, a, b, c vert, area, px, py float32) (float32, float32) {
	p := vert{x: px, y: py}
	l0 := edge(b, c, p) / area
	l1 := edge(c, a, p) / area
	l2 := edge(a, b, p) / area
	if !a.clip {
		return l0*a.u + l1*b.u + l2*c.u, l0*a.v + l1*b.v + l2*c.v
	}
	iw := l0*a.invW + l1*b.invW + l2*c.invW
	if iw == 0 {
		return 0, 0
	}
	return (l0*a.u*a.invW + l1*b.u*b.invW + l2*c.u*c.invW) / iw,
		(l0*a.v*a.invW + l1*b.v*b.invW + l2*c.v*c.invW) / iw
}

func hypot32(x, y float32) float32 {
	return float32(math.Hypot(float64(x), float64(y)))
}

// wrapTexel folds an integer texel coordinate per the TEXWRAP mode.
func wrapTexel(i int, size uint32, clamp uint32) uint32 {
	if clamp != 0 {
		if i < 0 {
			i = 0
		}
		if i > int(size)-1 {
			i = int(size) - 1
		}
		return uint32(i)
	}
	i &= int(size - 1)
	if i < 0 {
		i += int(size)
	}
	return uint32(i)
}

// modTex samples the bound texture at (u,v) and combines it with the vertex
// colour per TEXFUNC. Texcoords are normalized (0..1) in transformed mode, or
// absolute in through mode. TEXWRAP decides repeat vs clamp per axis.
func modTex(m *Machine, s *geState, u, v, rho float32, r, g, b, a byte) (byte, byte, byte, byte) {
	if s.texW == 0 || s.texH == 0 || s.texAddr == 0 {
		return r, g, b, a
	}
	// Pick the mip level for this fragment and sample at THAT level's size.
	lvl := s.texLevel(rho)
	texW, texH := s.texW>>lvl, s.texH>>lvl
	if lvl > 0 && s.texWN[lvl] != 0 {
		texW, texH = s.texWN[lvl], s.texHN[lvl]
	}
	if texW == 0 || texH == 0 {
		lvl, texW, texH = 0, s.texW, s.texH
	}
	wrap := func(t float32, size uint32, clamp uint32) uint32 {
		i := int(t * float32(size))
		if clamp != 0 {
			// CLAMP: the last texel is smeared outward. Repeating a clamped
			// texture is what smeared road markings across the lanes.
			if i < 0 {
				i = 0
			}
			if i > int(size)-1 {
				i = int(size) - 1
			}
			return uint32(i)
		}
		i &= int(size - 1)
		if i < 0 {
			i += int(size)
		}
		return uint32(i)
	}
	var tr, tg, tb, ta byte
	if s.texLinear {
		// TEXFILTER linear: bilinear blend of the four neighbouring texels. The
		// game asks for it on nearly every 3-D primitive; sampling nearest left
		// the road and car surfaces harshly blocky.
		fu := u*float32(texW) - 0.5
		fv := v*float32(texH) - 0.5
		fx := fu - float32(math.Floor(float64(fu)))
		fy := fv - float32(math.Floor(float64(fv)))
		u0 := wrapTexel(int(math.Floor(float64(fu))), texW, s.texWrapU)
		v0 := wrapTexel(int(math.Floor(float64(fv))), texH, s.texWrapV)
		u1 := wrapTexel(int(math.Floor(float64(fu)))+1, texW, s.texWrapU)
		v1 := wrapTexel(int(math.Floor(float64(fv)))+1, texH, s.texWrapV)
		r00, g00, b00, a00 := m.sampleTexLvl(s, u0, v0, lvl)
		r10, g10, b10, a10 := m.sampleTexLvl(s, u1, v0, lvl)
		r01, g01, b01, a01 := m.sampleTexLvl(s, u0, v1, lvl)
		r11, g11, b11, a11 := m.sampleTexLvl(s, u1, v1, lvl)
		lerp := func(c00, c10, c01, c11 byte) byte {
			top := float32(c00)*(1-fx) + float32(c10)*fx
			bot := float32(c01)*(1-fx) + float32(c11)*fx
			return byte(top*(1-fy) + bot*fy)
		}
		tr, tg, tb, ta = lerp(r00, r10, r01, r11), lerp(g00, g10, g01, g11),
			lerp(b00, b10, b01, b11), lerp(a00, a10, a01, a11)
	} else {
		tr, tg, tb, ta = m.sampleTex(s, wrap(u, texW, s.texWrapU), wrap(v, texH, s.texWrapV))
	}

	var or, og, ob, oa byte
	switch s.texFunc {
	case 1: // decal: texture replaces by its alpha
		if s.texUseA {
			or = byte((uint32(r)*(255-uint32(ta)) + uint32(tr)*uint32(ta)) / 255)
			og = byte((uint32(g)*(255-uint32(ta)) + uint32(tg)*uint32(ta)) / 255)
			ob = byte((uint32(b)*(255-uint32(ta)) + uint32(tb)*uint32(ta)) / 255)
			oa = a
		} else {
			or, og, ob, oa = tr, tg, tb, a
		}
	case 2: // blend: mix vertex and env colour by the texel, per TEXENVCOLOR
		er, eg, eb := byte(s.texEnvCol), byte(s.texEnvCol>>8), byte(s.texEnvCol>>16)
		mix := func(c, e, t byte) byte {
			return byte((uint32(c)*(255-uint32(t)) + uint32(e)*uint32(t)) / 255)
		}
		or, og, ob = mix(r, er, tr), mix(g, eg, tg), mix(b, eb, tb)
		oa = a
		if s.texUseA {
			oa = mul8(ta, a)
		}
	case 3: // replace
		or, og, ob = tr, tg, tb
		oa = a
		if s.texUseA {
			oa = ta
		}
	case 4: // add
		or, og, ob = clamp255(int32(r)+int32(tr)), clamp255(int32(g)+int32(tg)), clamp255(int32(b)+int32(tb))
		oa = a
		if s.texUseA {
			oa = mul8(ta, a)
		}
	default: // 0: modulate
		or, og, ob = mul8(tr, r), mul8(tg, g), mul8(tb, b)
		oa = a
		if s.texUseA {
			oa = mul8(ta, a)
		}
	}
	if s.texDouble {
		// TEXFUNC bit 16: the PSP doubles the textured result. Games lean on it
		// to get back the headroom that modulating by a mid-grey material costs;
		// ignoring it renders everything at half brightness.
		or = clamp255(int32(or) * 2)
		og = clamp255(int32(og) * 2)
		ob = clamp255(int32(ob) * 2)
	}
	return or, og, ob, oa
}

// sampleTex reads one texel from the bound texture: the direct formats
// (5650/5551/4444/8888) and the indexed ones (CLUT4/CLUT8), honouring the
// swizzled block layout when TEXMODE selects it.
func (m *Machine) sampleTex(s *geState, tx, ty uint32) (byte, byte, byte, byte) {
	return m.sampleTexLvl(s, tx, ty, 0)
}

// sampleTexLvl reads one texel from mip level lvl: the direct formats
// (5650/5551/4444/8888) and the indexed ones (CLUT4/CLUT8), honouring the
// swizzled block layout when TEXMODE selects it.
func (m *Machine) sampleTexLvl(s *geState, tx, ty, lvl uint32) (byte, byte, byte, byte) {
	addr, stride := s.texAddr, s.texStride
	if lvl > 0 && s.texAddrN[lvl] != 0 {
		addr, stride = s.texAddrN[lvl], s.texStrideN[lvl]
	}
	if stride == 0 {
		stride = s.texW >> lvl
	}
	if addr == 0 || stride == 0 {
		return 0xFF, 0xFF, 0xFF, 0xFF
	}
	switch s.texFmt {
	case 3: // 8888
		off := m.texOff(s, tx*4, ty, stride*4)
		c := m.read32(addr + off)
		return byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)
	case 0, 1, 2: // 565 / 5551 / 4444
		off := m.texOff(s, tx*2, ty, stride*2)
		return decode16a(u16(m, addr+off), s.texFmt)
	case 4: // CLUT4: two texels per byte, low nibble first
		off := m.texOff(s, tx/2, ty, stride/2)
		raw := uint32(m.Read(addr+off)) >> (4 * (tx & 1)) & 0xF
		return s.clutLookup(raw)
	case 5: // CLUT8
		off := m.texOff(s, tx, ty, stride)
		return s.clutLookup(uint32(m.Read(addr + off)))
	}
	return 0xFF, 0xFF, 0xFF, 0xFF
}

// texLevel picks the mip level for a fragment from rho — how many texels the
// texcoords advance per screen pixel. LOD = log2(rho), plus the TEXLEVEL bias,
// clamped to the levels the game actually supplied.
func (s *geState) texLevel(rho float32) uint32 {
	if s.texMaxLvl == 0 || rho <= 1 {
		return 0
	}
	lod := float32(math.Log2(float64(rho))) + s.texLodBias
	if lod <= 0 {
		return 0
	}
	l := uint32(lod + 0.5)
	if l > s.texMaxLvl {
		l = s.texMaxLvl
	}
	if l > 7 {
		l = 7
	}
	return l
}

// texOff converts a (byte-x, y) texel position to a byte offset in the texture
// buffer. Swizzled textures store 16-byte × 8-row blocks contiguously.
func (m *Machine) texOff(s *geState, xb, y, rowBytes uint32) uint32 {
	if !s.texSwizzle || rowBytes < 16 {
		return y*rowBytes + xb
	}
	rowBlocks := rowBytes / 16
	block := (y/8)*rowBlocks + xb/16
	return block*128 + (y%8)*16 + xb%16
}

// clutLookup maps an index through CLUTFORMAT (shift, mask, base) into the
// latched palette.
func (s *geState) clutLookup(raw uint32) (byte, byte, byte, byte) {
	shift := (s.clutFmt >> 2) & 0x1F
	mask := (s.clutFmt >> 8) & 0xFF
	base := (s.clutFmt >> 16) & 0x1F
	c := s.clut[((raw>>shift)&mask+base<<4)&0xFF]
	return byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)
}

// decode16a decodes a 16-bit texel/palette entry with its true alpha bit(s)
// (unlike the display path, which forces alpha opaque).
func decode16a(p uint16, fmt uint32) (byte, byte, byte, byte) {
	ext := func(v, bits uint16) byte {
		v &= (1 << bits) - 1
		return byte((uint32(v) * 255) / ((1 << bits) - 1))
	}
	switch fmt {
	case psm5551:
		a := byte(0)
		if p&0x8000 != 0 {
			a = 0xFF
		}
		return ext(p, 5), ext(p>>5, 5), ext(p>>10, 5), a
	case psm4444:
		return ext(p, 4), ext(p>>4, 4), ext(p>>8, 4), ext(p>>12, 4)
	default: // psm5650
		return ext(p, 5), ext(p>>5, 6), ext(p>>11, 5), 0xFF
	}
}

// stencilTest runs the stencil test against the destination alpha (which IS the
// PSP's stencil buffer). The caller applies the SFAIL / ZFAIL / ZPASS operation
// once the depth test has also run.
func (s *geState) stencilTest(dstA byte) bool {
	if !s.stencilOn || s.clearOn {
		return true
	}
	ref := s.stRef & s.stMask
	cur := uint32(dstA) & s.stMask
	pass := false
	switch s.stFunc {
	case 0: // never
	case 1: // always
		pass = true
	case 2:
		pass = ref == cur
	case 3:
		pass = ref != cur
	case 4:
		pass = ref < cur
	case 5:
		pass = ref <= cur
	case 6:
		pass = ref > cur
	default:
		pass = ref >= cur
	}
	return pass
}

// stencilOp applies one stencil operation to the stored value.
func stencilOp(op uint32, cur, ref byte) byte {
	switch op {
	case 1: // zero
		return 0
	case 2: // replace with the reference
		return ref
	case 3: // invert
		return ^cur
	case 4: // increment (saturating)
		if cur < 0xFF {
			return cur + 1
		}
		return 0xFF
	case 5: // decrement (saturating)
		if cur > 0 {
			return cur - 1
		}
		return 0
	default: // keep
		return cur
	}
}

// alphaPass runs the alpha test (ATE/ATST). Burnout alpha-tests its decals,
// windows and foliage; without it their transparent texels paint solid, which
// is what put stray triangles across the car.
func (s *geState) alphaPass(a byte) bool {
	if !s.alphaTestOn || s.clearOn {
		return true
	}
	src := uint32(a) & s.alphaTestMask
	ref := s.alphaRef & s.alphaTestMask
	switch s.alphaFunc {
	case 0:
		return false
	case 1:
		return true
	case 2:
		return src == ref
	case 3:
		return src != ref
	case 4:
		return src < ref
	case 5:
		return src <= ref
	case 6:
		return src > ref
	default:
		return src >= ref
	}
}

// blendFactor evaluates one PSP blend factor against the source and destination
// colours (per channel).
func blendFactor(f uint32, sc, sa, da byte, fix uint32, ch int) uint32 {
	fixCh := byte(fix >> (8 * ch))
	switch f {
	case 0: // source colour
		return uint32(sc)
	case 1: // one minus source colour
		return 255 - uint32(sc)
	case 2: // source alpha
		return uint32(sa)
	case 3: // one minus source alpha
		return 255 - uint32(sa)
	case 4: // destination alpha
		return uint32(da)
	case 5: // one minus destination alpha
		return 255 - uint32(da)
	case 6: // double source alpha
		return min32(2*uint32(sa), 255)
	case 7: // one minus double source alpha
		return 255 - min32(2*uint32(sa), 255)
	case 8: // double destination alpha
		return min32(2*uint32(da), 255)
	case 9: // one minus double destination alpha
		return 255 - min32(2*uint32(da), 255)
	default: // 10: fixed colour
		return uint32(fixCh)
	}
}

func min32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func clamp255(v int32) byte {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}

// putPixel runs the per-fragment pipeline in HARDWARE ORDER — scissor, alpha
// test, stencil test, DEPTH test, blend, write — and stores the result.
//
// The order matters: the depth test comes AFTER the alpha test, so a fragment
// killed by the alpha test never touches the depth buffer. Testing depth first
// (as this did) let the fully transparent texels of an alpha-masked billboard —
// Burnout's roadside trees — stamp their depth anyway, and everything behind
// them was then rejected: the trees came out truncated, cut off against nothing.
func (m *Machine) putPixel(s *geState, x, y int, z float32, r, g, b, a byte) {
	// Scissor: the game restricts drawing to a rectangle (its off-screen targets
	// are smaller than the screen).
	if !s.clearOn && (x < s.scX0 || x > s.scX1 || y < s.scY0 || y > s.scY1) && s.scX1 > 0 {
		return
	}
	if !s.alphaPass(a) {
		return
	}
	base := s.fbAddress()
	stride := s.fbStride
	if stride == 0 {
		stride = dispW
	}
	off := uint32(y)*stride + uint32(x)
	if x == geProbeX && y == geProbeY {
		fmt.Printf("PIXEL(%d,%d) prim#%d rgba=%02X%02X%02X%02X blend=%v(%d,%d) atest=%v(fn%d ref%d) wrap=(%d,%d) map=%d/src%d maxlvl=%d bias=%.2f tex=%08X f%d fn%d dbl=%v mat=%08X\n",
			x, y, gePrimSeq, r, g, b, a, s.blendOn, s.blendSrc, s.blendDst,
			s.alphaTestOn, s.alphaFunc, s.alphaRef, s.texWrapU, s.texWrapV,
			s.texMapMode, s.texProjSrc, s.texMaxLvl, s.texLodBias,
			s.texAddr, s.texFmt, s.texFunc, s.texDouble, s.matColor)
	}
	// Stencil test against the destination alpha (the PSP's stencil buffer).
	// A stencil failure applies the SFAIL op and stops; the depth test then
	// decides between the ZFAIL and ZPASS ops.
	dstA := m.dstAlpha(s, base, off)
	stPass := s.stencilTest(dstA)
	if !stPass {
		if !s.clearOn {
			m.storeAlpha(s, base, off, stencilOp(s.stSFail, dstA, byte(s.stRef)))
		}
		return
	}
	// Depth test — after the alpha and stencil tests, never before.
	zPass := m.zTest(s, x, y, z)
	if !zPass {
		if !s.clearOn && s.stencilOn {
			m.storeAlpha(s, base, off, stencilOp(s.stZFail, dstA, byte(s.stRef)))
		}
		return
	}
	m.zWrite(s, x, y, z)
	// The fragment's own alpha still drives blending; only the value STORED in
	// the alpha channel is the stencil result.
	outA := a
	if s.stencilOn && !s.clearOn {
		outA = stencilOp(s.stZPass, dstA, byte(s.stRef))
	}
	if s.blendOn && !s.clearOn {
		dr, dg, db, da := m.dstPixel4(s, base, off)
		ch := func(sc, dc byte, i int) byte {
			sf := blendFactor(s.blendSrc, sc, a, da, s.blendFixA, i)
			df := blendFactor(s.blendDst, sc, a, da, s.blendFixB, i)
			sv := int32(uint32(sc) * sf / 255)
			dv := int32(uint32(dc) * df / 255)
			switch s.blendEq {
			case 1: // subtract
				return clamp255(sv - dv)
			case 2: // reverse subtract
				return clamp255(dv - sv)
			case 3: // min
				return clamp255(minI32(int32(sc), int32(dc)))
			case 4: // max
				return clamp255(maxI32(int32(sc), int32(dc)))
			case 5: // absolute difference
				return clamp255(absI32(int32(sc) - int32(dc)))
			default: // add
				return clamp255(sv + dv)
			}
		}
		r, g, b = ch(r, dr, 0), ch(g, dg, 1), ch(b, db, 2)
	}
	if s.maskRGB != 0 || s.maskA != 0 {
		// PMSKC/PMSKA: framebuffer bits with a 1 in the mask keep their old value.
		// The game's depth-only fills set the mask fully (write nothing).
		if s.maskRGB == 0xFFFFFF && s.maskA == 0xFF {
			return
		}
		dr, dg, db, da := m.dstPixel4(s, base, off)
		mr, mg, mb, ma := byte(s.maskRGB), byte(s.maskRGB>>8), byte(s.maskRGB>>16), byte(s.maskA)
		r = r&^mr | dr&mr
		g = g&^mg | dg&mg
		b = b&^mb | db&mb
		outA = outA&^ma | da&ma
	}
	m.storePixel(s, base, off, r, g, b, outA)
}

// storePixel writes one pixel to the render target in its PSP format.
func (m *Machine) storePixel(s *geState, base, off uint32, r, g, b, a byte) {
	switch s.fbFmt {
	case psm8888:
		m.write32(base+off*4, uint32(r)|uint32(g)<<8|uint32(b)<<16|uint32(a)<<24)
	default:
		var p uint16
		switch s.fbFmt {
		case psm5551:
			p = uint16(r>>3) | uint16(g>>3)<<5 | uint16(b>>3)<<10
			if a >= 128 {
				p |= 0x8000
			}
		case psm4444:
			p = uint16(r>>4) | uint16(g>>4)<<4 | uint16(b>>4)<<8 | uint16(a>>4)<<12
		default: // 565
			p = uint16(r>>3) | uint16(g>>2)<<5 | uint16(b>>3)<<11
		}
		addr := base + off*2
		m.Write(addr, byte(p))
		m.Write(addr+1, byte(p>>8))
	}
}

// dstAlpha reads the destination alpha — the PSP's stencil buffer.
func (m *Machine) dstAlpha(s *geState, base, off uint32) byte {
	_, _, _, a := m.dstPixel4(s, base, off)
	return a
}

// storeAlpha updates only the stencil (alpha) channel, for a fragment that
// failed the stencil test but still applies its sfail operation.
func (m *Machine) storeAlpha(s *geState, base, off uint32, a byte) {
	r, g, b, _ := m.dstPixel4(s, base, off)
	m.storePixel(s, base, off, r, g, b, a)
}

// dstPixel reads back the render-target pixel at linear offset off (for blending).
func (m *Machine) dstPixel(s *geState, base, off uint32) (byte, byte, byte) {
	r, g, b, _ := m.dstPixel4(s, base, off)
	return r, g, b
}

// dstPixel4 reads back the render-target pixel including its stored alpha.
func (m *Machine) dstPixel4(s *geState, base, off uint32) (byte, byte, byte, byte) {
	if s.fbFmt == psm8888 {
		c := m.read32(base + off*4)
		return byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)
	}
	return decode16a(u16(m, base+off*2), s.fbFmt)
}

// --- small helpers ---------------------------------------------------------

func edge(a, b, c vert) float32 { return (c.x-a.x)*(b.y-a.y) - (c.y-a.y)*(b.x-a.x) }

func mul8(a, b byte) byte { return byte(uint32(a) * uint32(b) / 255) }

func fmin3(a, b, c float32) float32 { return fmin(fmin(a, b), c) }
func fmax3(a, b, c float32) float32 { return fmax(fmax(a, b), c) }
func fmin(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func fmax(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
func clampI(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func minI32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func maxI32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func absI32(a int32) int32 {
	if a < 0 {
		return -a
	}
	return a
}
