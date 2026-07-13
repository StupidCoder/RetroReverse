package n3ds

import "fmt"

// gpu_raster.go runs a PICA200 draw: fetch vertices from the attribute
// buffers, run each through the LLE vertex shader (gpu_shader.go), assemble
// primitives, and rasterise them into the tiled render target with
// perspective-correct interpolation, depth testing and the TEV combiner
// (gpu_tev.go). The triangle fill is the repo's PSX barycentric shape
// (psx/gpu_raster.go) extended with 1/w interpolation; the tiled framebuffer
// write is the inverse of the 8×8-Morton detiling cgfx_texture.go decodes.

// vsOut is one shader-run vertex after the output map is applied.
type vsOut struct {
	pos   [4]float32 // clip-space position
	color [4]float32
	uv    [3][2]float32

	// texcoord0's w. Not a texture coordinate: it is the projective divisor for
	// texture unit 0's Projection2D and Shadow2D modes, and in the latter it is
	// also the depth the shadow map is compared against. Captain Toad's world
	// draws sample a shadow map through it.
	uv0w float32

	// The fragment-lighting inputs. The PICA does not carry a normal vector:
	// the shader emits a tangent-space quaternion, and the fragment normal is
	// that quaternion applied to (0,0,1). view is the fragment→eye vector.
	quat [4]float32
	view [3]float32
}

// loaderBuf is one attribute-loader configuration: where its vertex data starts,
// which attributes it supplies (a run in the draw's shared comps slice — an index
// rather than a sub-slice, so appending to comps cannot leave a loader pointing at
// a stale backing array), and its stride.
type loaderBuf struct {
	off      uint32
	first, n int
	stride   uint32
}

// draw executes one draw trigger against the current register state.
func (g *GPU) draw(indexed bool) {
	if g.TraceDraws > 0 && g.Regs[regLightingEnable]&1 != 0 {
		g.dumpLighting()
		g.dumpFragment()
	}
	if g.checkUnsupported() {
		return
	}
	m := g.m

	base := m.gpuAddrToVirt(g.Regs[regAttrBase] << 3)
	fmtWord := uint64(g.Regs[regAttrFmtLow]) | uint64(g.Regs[regAttrFmtHigh]&0xFFFF)<<32
	fixedMask := g.Regs[regAttrFmtHigh] >> 16 & 0xFFF

	// Loader buffers: offset, component list (attribute indices; 12-15 are
	// padding), stride. The two slices are scratch (gpu.go), rebuilt per draw and
	// reused across draws.
	bufs, comps := g.bufs[:0], g.comps[:0]
	for b := 0; b < 12; b++ {
		cfg1 := g.Regs[0x204+uint32(b)*3]
		cfg2 := g.Regs[0x205+uint32(b)*3]
		n := int(cfg2 >> 28)
		if n == 0 {
			continue
		}
		perm := uint64(cfg1) | uint64(cfg2&0xFFFF)<<32
		start := len(comps)
		for j := 0; j < n; j++ {
			comps = append(comps, int(perm>>(4*uint(j))&0xF))
		}
		bufs = append(bufs, loaderBuf{off: g.Regs[0x203+uint32(b)*3], first: start, n: n, stride: cfg2 >> 16 & 0xFF})
	}
	g.bufs, g.comps = bufs, comps // keep the grown backing arrays

	count := g.Regs[regNumVertices]
	first := g.Regs[regVertexOff]
	idxCfg := g.Regs[regIndexConfig]
	idxAddr := base + idxCfg&0x0FFFFFFF
	idx16 := idxCfg>>31 != 0

	// Shader input map: input register j ← attribute (perm >> 4j) & 0xF.
	inPerm := uint64(g.Regs[regVshAttrPermL]) | uint64(g.Regs[regVshAttrPermH])<<32

	trace := g.TraceDraws > 0
	if trace {
		g.TraceDraws--
		tfb := g.fbstate()
		// The vertex-fetch configuration, as raw registers: the whole question
		// of whether the geometry or the camera is wrong turns on reading these
		// as numbers rather than trusting the decode.
		fmt.Printf("  FETCH base=0x%08X (reg 0x200=0x%08X) fmt=%08X/%08X fixedMask=%03X nattr=%d idxCfg=0x%08X (16bit=%v addr=0x%08X) inPerm=%08X_%08X maxIn=%d\n",
			base, g.Regs[regAttrBase], g.Regs[regAttrFmtLow], g.Regs[regAttrFmtHigh],
			fixedMask, g.Regs[regAttrFmtHigh]>>28&0xF+1, idxCfg, idx16, idxAddr,
			g.Regs[regVshAttrPermH], g.Regs[regVshAttrPermL], g.Regs[0x2B9]&0xF)
		for bi, b := range bufs {
			fmt.Printf("    buf%d off=0x%06X stride=%d comps=%v -> addr(v0)=0x%08X\n",
				bi, b.off, b.stride, comps[b.first:b.first+b.n], base+b.off)
		}
		for a := 0; a < 12; a++ {
			if fixedMask>>uint(a)&1 != 0 {
				fmt.Printf("    fixed attr%d = %v\n", a, g.fixedVal[a])
			}
		}
		fmt.Printf("  VIEWPORT offset x=%d y=%d (reg 0x068=0x%08X) depthmap=0x%08X(%s) scale=%f off=%f 0x107=%08X\n",
			g.Regs[regViewportXY]&0x3FF, g.Regs[regViewportXY]>>16&0x3FF, g.Regs[regViewportXY],
			g.Regs[regDepthMapMode], map[bool]string{true: "Z", false: "W"}[g.Regs[regDepthMapMode]&1 != 0],
			f24bits(g.Regs[regDepthMapScale]), f24bits(g.Regs[regDepthMapOffset]), g.Regs[regDepthColorMask])
		fmt.Printf("gpu draw %d: indexed=%v count=%d first=%d base=0x%08X bufs=%d prim=%d vp=%.1fx%.1f color=0x%08X depth=0x%08X dim=%dx%d test=%v wr=%v\n",
			g.Draws, indexed, count, first, base, len(bufs), g.Regs[regPrimConfig]>>8&3,
			f24bits(g.Regs[regViewportWidth]), f24bits(g.Regs[regViewportHeight]),
			tfb.colorAddr, tfb.depthAddr, tfb.width, tfb.height, tfb.depthTest, tfb.depthWr)
		for _, ci := range []int{0, 1, 2, 3, 4, 5, 6, 32, 33, 34, 35, 64, 65} {
			fmt.Printf("  c%-2d = %v\n", ci, g.Float[ci])
		}
		var nan []int
		for ci := range g.Float {
			for k := 0; k < 4; k++ {
				if g.Float[ci][k] != g.Float[ci][k] {
					nan = append(nan, ci)
					break
				}
			}
		}
		fmt.Printf("  NaN uniforms at draw: %v; bool=%04X entry=0x%03X\n",
			nan, g.Bool, g.Regs[regVshEntry]&0xFFFF)
	}

	// Run every vertex through the shader, then assemble. outs is scratch (gpu.go),
	// grown once and reused by every later draw.
	tv := m.profStart() // one clock read per draw (profile.go)
	if uint32(cap(g.outs)) < count {
		g.outs = make([]vsOut, count)
	}
	outs := g.outs[:count]
	// The shader's input and output register files, hoisted out of the loop: they
	// are overwritten wholesale per vertex, and returning the output file by value
	// (256 bytes) was 6% of the frame in runtime.duffcopy alone.
	var vin, vout [16][4]float32
	entry := int(g.Regs[regVshEntry] & 0xFFF) // draw-constant; nothing in the loop writes it
	for i := uint32(0); i < count; i++ {
		vi := first + i
		if indexed {
			if idx16 {
				vi = uint32(m.Read(idxAddr+i*2)) | uint32(m.Read(idxAddr+i*2+1))<<8
			} else {
				vi = uint32(m.Read(idxAddr + i))
			}
		}

		// Fetch the vertex's attributes (fixed values fill the gaps).
		var attrs [16][4]float32
		for a := 0; a < 16; a++ {
			attrs[a] = [4]float32{0, 0, 0, 1}
			if fixedMask>>uint(a)&1 != 0 {
				attrs[a] = g.fixedVal[a]
			}
		}
		for _, b := range bufs {
			p := base + b.off + vi*b.stride
			for _, c := range comps[b.first : b.first+b.n] {
				if c >= 12 { // padding component: skip (c-11)*4 bytes
					p += uint32(c-11) * 4
					continue
				}
				f := fmtWord >> (4 * uint(c)) & 0xF
				n := int(f>>2) + 1
				var val [4]float32
				val[3] = 1
				for j := 0; j < n; j++ {
					switch f & 3 {
					case 0: // signed byte
						val[j] = float32(int8(m.Read(p)))
						p++
					case 1: // unsigned byte
						val[j] = float32(m.Read(p))
						p++
					case 2: // signed 16-bit
						val[j] = float32(int16(uint16(m.Read(p)) | uint16(m.Read(p+1))<<8))
						p += 2
					case 3: // float32
						val[j] = f32bits(m.ReadWord(p))
						p += 4
					}
				}
				attrs[c] = val
			}
		}

		// Map attributes into the shader's input registers and run it.
		mapAttrsToInputs(&vin, &attrs, inPerm, int(g.Regs[regVshMaxInput]&0xF)+1)
		if !g.shaderRun(&vin, &vout, entry) {
			return
		}
		g.mapOutputs(&vout, &outs[i])
		if trace && i < 8 {
			r := &outs[i]
			fmt.Printf("  v%-3d (i=%d) clip=%v\n", vi, i, r.pos)
			for a := 0; a < 12; a++ {
				if fmtWord>>(4*uint(a))&0xF != 0 || fixedMask>>uint(a)&1 != 0 {
					fmt.Printf("       attr%-2d = %v\n", a, attrs[a])
				}
			}
		}
	}

	m.profEnd(bucketVertex, tv)

	// Primitive assembly. 0x25E bits 8-9: triangles / strip / fan; mode 3 is
	// the geometry-shader path (already rejected above via 0x229, but keep the
	// check local too).
	tr := m.profStart()
	defer m.profEnd(bucketRaster, tr)
	prim := g.Regs[regPrimConfig] >> 8 & 3
	n := len(outs)
	emit := func(a, b, c int) { g.triangle(&outs[a], &outs[b], &outs[c]) }
	switch prim {
	case 0, 3:
		// Mode 3 is the "geometry primitive": meaningful only to a geometry
		// shader. The draw-time check above guarantees the geometry stage is
		// OFF (0x229&3 == 0) when we get here, so the vertices pass through
		// assembly untouched — independent triangles, same as mode 0 (first
		// seen in the save-complete screen transition after the file-select).
		for i := 0; i+2 < n; i += 3 {
			emit(i, i+1, i+2)
		}
	case 1:
		for i := 0; i+2 < n; i++ {
			if i&1 == 0 {
				emit(i, i+1, i+2)
			} else {
				emit(i+1, i, i+2)
			}
		}
	case 2:
		for i := 1; i+1 < n; i++ {
			emit(0, i, i+1)
		}
	default:
		m.CPU.Halt("gpu: primitive mode 3 (geometry) unimplemented")
	}
}

// mapAttrsToInputs delivers the fetched vertex attributes to the shader's input
// registers through the permutation in 0x2BB/0x2BC.
//
// The permutation runs ATTRIBUTE → REGISTER: nibble a names the input register
// attribute a is delivered to. Reading it the other way round (register j takes
// attribute perm[j]) yields the inverse permutation, and the game's own
// configuration says which is meant. A Captain Toad draw fetches attributes 0-2
// with perm 0x340; inverted, that hands v1 and v2 attributes 4 and 3 — which
// this draw does not fetch at all — while the UV and the colour it *does* fetch
// reach no register. Run the right way it is a bijection onto v0/v4/v3. (The
// inverse is only visible when the permutation is not an involution, which is
// why Super Mario 3D Land renders pixel-identically either way.)
func mapAttrsToInputs(v *[16][4]float32, attrs *[16][4]float32, perm uint64, nAttr int) {
	for j := range v {
		v[j] = [4]float32{0, 0, 0, 1}
	}
	for a := 0; a < nAttr && a < 16; a++ {
		v[perm>>(4*uint(a))&0xF] = attrs[a]
	}
}

// mapOutputs routes the shader's output registers to their semantics via the
// output map (registers 0x050-0x056): each mapped output register's word
// holds one semantic byte per component.
func (g *GPU) mapOutputs(o *[16][4]float32, dst *vsOut) {
	var out vsOut
	out.color = [4]float32{1, 1, 1, 1}
	total := int(g.Regs[regShOutmapTotal] & 7)
	for r := 0; r < total; r++ {
		sem := g.Regs[regShOutmapO0+uint32(r)]
		for c := 0; c < 4; c++ {
			s := sem >> (8 * uint(c)) & 0x1F
			v := o[r][c]
			switch {
			case s <= 0x03: // position xyzw
				out.pos[s] = v
			case s >= 0x08 && s <= 0x0B: // colour rgba
				out.color[s-0x08] = v
			case s == 0x0C || s == 0x0D: // texcoord 0
				out.uv[0][s-0x0C] = v
			case s == 0x0E || s == 0x0F: // texcoord 1
				out.uv[1][s-0x0E] = v
			case s == 0x16 || s == 0x17: // texcoord 2
				out.uv[2][s-0x16] = v
			case s >= 0x04 && s <= 0x07: // tangent-space normal quaternion
				out.quat[s-0x04] = v
			case s == 0x10: // texcoord0.w: the projective/shadow divisor
				out.uv0w = v
			case s >= 0x12 && s <= 0x14: // view vector (fragment → eye)
				out.view[s-0x12] = v
			}
		}
	}
	*dst = out
}

// fbState is the output-merger state a triangle fill needs, resolved once per
// triangle from the register file.
type fbState struct {
	colorAddr, depthAddr uint32
	width, height        uint32
	depthTest, depthWr   bool
	depthFunc            uint32
	colorMask            uint32 // bits 0-3 = RGBA write enables
	vpHalfW, vpHalfH     float32
	depthScale, depthOff float32
	depthZBuffer         bool // 0x06D bit 0: Z-buffering; clear means W-buffering
	shadowMode           bool // 0x100's fragment-operation mode is Shadow

	// The render target's backing memory, resolved once per triangle. A fragment
	// costs up to fourteen byte-granular bus calls; against a target that sits
	// wholly inside one region — it is VRAM, so it always does — the same bytes are
	// reachable by indexing the region's slice directly.
	//
	// nil means "go through the bus": the target straddles a region edge, is
	// unmapped, or a debugger has a memory watch installed, in which case the bus is
	// the only path that reports the accesses. An oracle that stopped answering the
	// debugger's questions when it got fast would be a poor trade.
	colorBuf, depthBuf   []byte
	colorOff, depthOff32 uint32 // where the buffer starts within its slice
}

func (g *GPU) fbstate() fbState {
	dim := g.Regs[regFramebufDim]
	dcm := g.Regs[regDepthColorMask]
	// The access enables gate the memory traffic in front of 0x107's test/write
	// bits. Zero means the pass touches that buffer not at all. (The individual
	// bits distinguish depth from stencil, which this model does not separate —
	// what is pinned, by the game's own VRAM map, is that zero means "no
	// access": Captain Toad's shadow pass leaves 0x107's depth bits set and
	// zeroes 0x114/0x115, and its depth buffer address still points at the main
	// pass's — a megabyte that would land on its own colour buffer.)
	colorWr := g.Regs[regColorbufWrite] != 0
	depthRd := g.Regs[regDepthbufRead] != 0
	depthWr := g.Regs[regDepthbufWrite] != 0
	cmask := dcm >> 8 & 0xF
	if !colorWr {
		cmask = 0
	}
	fb := fbState{
		colorAddr:    g.m.gpuAddrToVirt(g.Regs[regColorbufLoc] << 3),
		depthAddr:    g.m.gpuAddrToVirt(g.Regs[regDepthbufLoc] << 3),
		width:        dim & 0x7FF,
		height:       (dim >> 12 & 0x3FF) + 1,
		depthTest:    dcm&1 != 0 && depthRd,
		depthFunc:    dcm >> 4 & 7,
		colorMask:    cmask,
		depthWr:      dcm>>12&1 != 0 && depthWr,
		vpHalfW:      f24bits(g.Regs[regViewportWidth]),
		vpHalfH:      f24bits(g.Regs[regViewportHeight]),
		depthScale:   f24bits(g.Regs[regDepthMapScale]),
		depthOff:     f24bits(g.Regs[regDepthMapOffset]),
		depthZBuffer: g.Regs[regDepthMapMode]&1 != 0,
		shadowMode:   g.Regs[regBlendConfig]&3 == 3,
	}
	// Resolve the two targets' backing memory once, here, instead of ~14 times per
	// fragment. A tiled 4-byte-per-pixel buffer occupies whole 8×8 tiles, so round
	// the dimensions up before asking whether it fits in one region.
	size := ((fb.width + 7) &^ 7) * ((fb.height + 7) &^ 7) * 4
	fb.colorBuf, fb.colorOff = g.m.directRange(fb.colorAddr, size)
	fb.depthBuf, fb.depthOff32 = g.m.directRange(fb.depthAddr, size)
	return fb
}

// shadowMapWrite is the output merger for a shadow pass. A shadow-map texel is
// not a colour: it is a 24-bit depth in the low three bytes and an 8-bit
// density in the fourth — the same four bytes a colour texel occupies, read
// back by the Shadow2D sampler (gpu_texture.go) as depth+density rather than
// as RGBA.
//
// A fragment only contributes if it is nearer than what is already there. A
// zero density means "fully opaque shadow caster": it just deepens the map. A
// non-zero density is a *partial* shadow — the hardware attenuates it by the
// ratio of the two depths through a constant/linear curve (0x130) and keeps the
// darkest, which is how a shadow volume produces a soft edge without a second
// pass.
func (g *GPU) shadowMapWrite(fb *fbState, off uint32, depth float32, density uint8) {
	p := fb.colorAddr + off
	var dst []byte
	if fb.colorBuf != nil {
		dst = fb.colorBuf[fb.colorOff+off:]
	}
	z := uint32(depth * 0xFFFFFF)

	var refZ uint32
	var refS uint8
	if dst != nil {
		refZ = uint32(dst[0])<<16 | uint32(dst[1])<<8 | uint32(dst[2])
		refS = dst[3]
	} else {
		refZ = uint32(g.m.Read(p))<<16 | uint32(g.m.Read(p+1))<<8 | uint32(g.m.Read(p+2))
		refS = g.m.Read(p + 3)
	}
	if z >= refZ {
		return
	}
	if density == 0 {
		if dst != nil {
			dst[0], dst[1], dst[2] = byte(z>>16), byte(z>>8), byte(z)
			return
		}
		g.m.Write(p, byte(z>>16))
		g.m.Write(p+1, byte(z>>8))
		g.m.Write(p+2, byte(z))
		return
	}
	cfg := g.Regs[regShadowDensity]
	k, lin := f16(cfg), f16(cfg>>16)
	d := float32(density)
	if refZ != 0 {
		if den := k + lin*float32(z)/float32(refZ); den != 0 {
			d = float32(density) / den
		}
	}
	if s := uint8(clampf(d, 0, 255)); s < refS {
		if dst != nil {
			dst[3] = s
			return
		}
		g.m.Write(p+3, s)
	}
}

// triangle rasterises one clip-space triangle. Fixed-point barycentric fill
// (the PSX shape) over the tiled render target, interpolating 1/w for
// perspective correction.
func (g *GPU) triangle(a, b, c *vsOut) {
	// Reject any triangle touching w<=0 rather than clip — a bring-up
	// shortcut; counted so it can't hide. Written as !(w > 0) so NaN
	// positions are rejected too: the game deliberately poisons its
	// projection uniforms with NaN at frame start, and hardware clipping
	// drops such triangles (every plane test fails).
	if !(a.pos[3] > 0) || !(b.pos[3] > 0) || !(c.pos[3] > 0) ||
		a.pos[0] != a.pos[0] || b.pos[0] != b.pos[0] || c.pos[0] != c.pos[0] {
		g.RejectedTris++
		return
	}
	fb := g.fbstate()
	ls := g.lightstate()

	toScreen := func(v *vsOut) scrVert {
		iw := 1 / v.pos[3]
		return scrVert{
			x:    (v.pos[0]*iw + 1) * fb.vpHalfW,
			y:    (v.pos[1]*iw + 1) * fb.vpHalfH,
			z:    v.pos[2] * iw,
			iw:   iw,
			col:  v.color,
			uv:   v.uv,
			uv0w: v.uv0w,
			quat: v.quat,
			view: v.view,
		}
	}
	v0, v1, v2 := toScreen(a), toScreen(b), toScreen(c)

	edge := func(ax, ay, bx, by, px, py float32) float32 {
		return (bx-ax)*(py-ay) - (by-ay)*(px-ax)
	}
	area := edge(v0.x, v0.y, v1.x, v1.y, v2.x, v2.y)
	if area == 0 {
		g.ZeroAreaTris++
		return
	}
	// Face culling (0x040): 0 = none, 1 = cull front (CCW), 2 = cull back.
	cull := g.Regs[0x040] & 3
	ccw := area > 0
	if (cull == 1 && ccw) || (cull == 2 && !ccw) {
		g.CulledTris++
		return
	}
	if area < 0 { // orient so the inside test and interpolation share a sign
		v1, v2 = v2, v1
		area = -area
	}

	minX := int(min3f(v0.x, v1.x, v2.x))
	maxX := int(max3f(v0.x, v1.x, v2.x)) + 1
	minY := int(min3f(v0.y, v1.y, v2.y))
	maxY := int(max3f(v0.y, v1.y, v2.y)) + 1
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > int(fb.width) {
		maxX = int(fb.width)
	}
	if maxY > int(fb.height) {
		maxY = int(fb.height)
	}

	// Rasterise the bounding box, row by row.
	//
	// 8×8 tile rejection was tried here — test the three edge functions at a block's
	// corners and skip the block when it lies wholly outside one of them — and it was
	// 1.4% SLOWER. It is the right optimisation for the workload this rasteriser used
	// to have: long spiky triangles with enormous bounding boxes and slivers of
	// coverage, which is what the broken vertex permutation produced. The geometry is
	// coherent now, the triangles are small, most blocks are partly covered, and the
	// twelve extra edge evaluations per block buy nothing. Measured, not assumed.
	//
	// Hoisting this loop into a function of its own cost another 1% on its own, so it
	// stays where it is.
	for y := minY; y < maxY; y++ {
		// The PICA's framebuffer origin is bottom-left: viewport y counts up from
		// the buffer's LAST row, so a viewport shorter than the buffer renders
		// into its bottom, not its top. Captain Toad is the case that shows it —
		// it draws a 240×400 viewport into a 256×512 target and then display-
		// transfers the window at byte offset 0x1C000, which is exactly the
		// 512-400 = 112 rows the image is pushed down by (its bottom screen: a
		// 320-row viewport, a 0x30000 = 192-row offset, and 512-320 = 192).
		// Rasterising top-aligned put the diorama 112 rows above the panel's
		// window and ran it off the edge. Super Mario 3D Land cannot show this:
		// its buffer is exactly its viewport (240×400), so the two anchors agree.
		ty := uint32(int(fb.height) - 1 - y)
		for x := minX; x < maxX; x++ {
			// One tiled-offset computation per fragment, shared by the depth test,
			// the depth write and the colour write — it used to be redone in each.
			off := tiledOffset(uint32(x), ty, fb.width)
			px, py := float32(x)+0.5, float32(y)+0.5
			w0 := edge(v1.x, v1.y, v2.x, v2.y, px, py)
			w1 := edge(v2.x, v2.y, v0.x, v0.y, px, py)
			w2 := edge(v0.x, v0.y, v1.x, v1.y, px, py)
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			l0, l1, l2 := w0/area, w1/area, w2/area

			// The perspective-correct 1/w at this pixel: the depth needs it too,
			// so it is computed before the depth test rather than with the other
			// interpolated attributes.
			iw := l0*v0.iw + l1*v1.iw + l2*v2.iw

			// Depth. z/w interpolates linearly in screen space, and the depth-map
			// registers scale and bias it — but only in Z-buffering mode. Captain
			// Toad's scene runs the PICA's other mode, W-buffering (0x06D bit 0
			// clear, which is also the value nobody writing the register leaves
			// behind), where the result is multiplied back by w:
			//
			//	z·scale + w·offset  ==  (z/w·scale + offset)·w
			//
			// so the stored depth is linear in eye-space distance instead of in
			// screen space. It is not a subtle difference here. The game pairs it
			// with a depth scale of -1/110000, which in Z mode compresses the
			// entire scene into the bottom 151 of the depth buffer's 16.7M
			// values: every object lands on top of every other and the winner is
			// decided by rounding. The star sank behind the ruin, the pillar
			// behind the wall it stands in front of, and the arch's decorations
			// vanished into the stone.
			z := l0*v0.z + l1*v1.z + l2*v2.z
			depth := z*fb.depthScale + fb.depthOff
			if !fb.depthZBuffer && iw != 0 {
				depth /= iw
			}
			if depth < 0 {
				depth = 0
			}
			if depth > 1 {
				depth = 1
			}
			if !g.depthCompare(&fb, off, depth) {
				g.DepthKilled++
				// A depth-killed fragment carries no colour: the PICA kills it
				// before the TEV runs, so there is nothing to report but the
				// rejection itself.
				g.pixelEvent(uint32(x), ty, PixelEvent{ZReject: true})
				continue
			}

			// Perspective-correct attributes via 1/w.
			pc := func(a0, a1, a2 float32) float32 {
				return (l0*a0*v0.iw + l1*a1*v1.iw + l2*a2*v2.iw) / iw
			}
			var col [4]float32
			for i := 0; i < 4; i++ {
				col[i] = pc(v0.col[i], v1.col[i], v2.col[i])
			}
			var uv [3][2]float32
			for t := 0; t < 3; t++ {
				uv[t][0] = pc(v0.uv[t][0], v1.uv[t][0], v2.uv[t][0])
				uv[t][1] = pc(v0.uv[t][1], v1.uv[t][1], v2.uv[t][1])
			}
			uv0w := pc(v0.uv0w, v1.uv0w, v2.uv0w)

			// The fragment-lighting inputs: the tangent-space quaternion that
			// carries the normal, and the fragment→eye vector that positions the
			// lights. They go to the fragment stage as inputs — the lighting unit
			// is evaluated there, after the texture units, because it reads two of
			// them (the bump map and the shadow attenuation).
			var q [4]float32
			var vw [3]float32
			if ls.enabled {
				for i := 0; i < 4; i++ {
					q[i] = pc(v0.quat[i], v1.quat[i], v2.quat[i])
				}
				for i := 0; i < 3; i++ {
					vw[i] = pc(v0.view[i], v1.view[i], v2.view[i])
				}
			}

			r8, g8, b8, a8, discard, ok := g.fragment(col, uv, uv0w, &ls, q, vw)
			if !ok {
				return // fragment stage halted
			}
			if fb.shadowMode {
				// Shadow rendering: the output merger is replaced wholesale. The
				// TEV's green channel is the shadow's density and the fragment's
				// depth is its distance; they go into the render target as a
				// packed depth+density texel, and none of the alpha test, depth
				// test or blender runs. This is the pass that draws Captain
				// Toad's ShadowVolume geometry.
				g.shadowMapWrite(&fb, off, depth, uint8(g8))
				g.ShadowWrites++
				continue
			}
			if discard {
				// Alpha-tested out: no colour and no depth write, but the fragment
				// was shaded, so the debugger gets the colour it would have had.
				g.pixelEvent(uint32(x), ty, PixelEvent{
					R: r8, G: g8, B: b8, A: a8, AlphaReject: true,
				})
				continue
			}
			g.depthWrite(&fb, off, depth)
			g.writePixel(&fb, uint32(x), ty, off, r8, g8, b8, a8)
			g.PixelsDrawn++
		}
	}
}

// scrVert is a triangle vertex in screen space, with the attributes the fragment
// stage interpolates.
type scrVert struct {
	x, y, z, iw float32
	col         [4]float32
	uv          [3][2]float32
	uv0w        float32
	quat        [4]float32
	view        [3]float32
}

// tiledOffset returns the byte offset of pixel (x, y) in an 8×8-Morton-tiled
// 4-byte-per-pixel buffer of the given width.
func tiledOffset(x, y, width uint32) uint32 {
	tile := (y/8)*(width/8) + x/8
	mo := x&1 | y&1<<1 | x&2<<1 | y&2<<2 | x&4<<2 | y&4<<3
	return (tile*64 + mo) * 4
}

// depthCompare applies the depth test against the D24S8 depth buffer; the
// write happens separately (depthWrite) so an alpha-discarded fragment leaves
// the buffer untouched.
//
// off is the fragment's tiled byte offset, computed once by the caller: it used to
// be recomputed here, in depthWrite and again in writePixel, three times for the
// same pixel.
func (g *GPU) depthCompare(fb *fbState, off uint32, depth float32) bool {
	if !fb.depthTest {
		return true
	}
	var old uint32
	if fb.depthBuf != nil {
		d := fb.depthBuf[fb.depthOff32+off:]
		old = uint32(d[0]) | uint32(d[1])<<8 | uint32(d[2])<<16
	} else {
		p := fb.depthAddr + off
		old = uint32(g.m.Read(p)) | uint32(g.m.Read(p+1))<<8 | uint32(g.m.Read(p+2))<<16
	}
	nv := uint32(depth * 0xFFFFFF)
	switch fb.depthFunc {
	case 0:
		return false
	case 1:
		return true
	case 2:
		return nv == old
	case 3:
		return nv != old
	case 4:
		return nv < old
	case 5:
		return nv <= old
	case 6:
		return nv > old
	default:
		return nv >= old
	}
}

func (g *GPU) depthWrite(fb *fbState, off uint32, depth float32) {
	if !fb.depthTest || !fb.depthWr {
		return
	}
	nv := uint32(depth * 0xFFFFFF)
	if fb.depthBuf != nil {
		d := fb.depthBuf[fb.depthOff32+off:]
		d[0], d[1], d[2] = byte(nv), byte(nv>>8), byte(nv>>16)
		return
	}
	p := fb.depthAddr + off
	g.m.Write(p, byte(nv))
	g.m.Write(p+1, byte(nv>>8))
	g.m.Write(p+2, byte(nv>>16))
}

// writePixel blends the fragment with the destination and stores it in the
// tiled RGBA8 colour buffer (bytes [A,B,G,R]).
func (g *GPU) writePixel(fb *fbState, x, y, off uint32, r, gr, b, a uint8) {
	var dst []byte
	var da, db, dg, dr uint8
	if fb.colorBuf != nil {
		dst = fb.colorBuf[fb.colorOff+off:]
		da, db, dg, dr = dst[0], dst[1], dst[2], dst[3]
	} else {
		p := fb.colorAddr + off
		da, db = g.m.Read(p), g.m.Read(p+1)
		dg, dr = g.m.Read(p+2), g.m.Read(p+3)
	}
	r, gr, b, a = g.blend(r, gr, b, a, dr, dg, db, da)
	// What the debugger sees is the blended value, and it counts as drawn only if
	// the colour mask actually lets some of it through — a pass with the mask at
	// zero rasterises fragments that never reach memory.
	g.pixelEvent(x, y, PixelEvent{R: r, G: gr, B: b, A: a, Drawn: fb.colorMask != 0})
	if dst != nil {
		if fb.colorMask&1 != 0 {
			dst[3] = r
		}
		if fb.colorMask&2 != 0 {
			dst[2] = gr
		}
		if fb.colorMask&4 != 0 {
			dst[1] = b
		}
		if fb.colorMask&8 != 0 {
			dst[0] = a
		}
		return
	}
	p := fb.colorAddr + off
	if fb.colorMask&1 != 0 {
		g.m.Write(p+3, r)
	}
	if fb.colorMask&2 != 0 {
		g.m.Write(p+2, gr)
	}
	if fb.colorMask&4 != 0 {
		g.m.Write(p+1, b)
	}
	if fb.colorMask&8 != 0 {
		g.m.Write(p, a)
	}
}

func min3f(a, b, c float32) float32 {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

func max3f(a, b, c float32) float32 {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}
