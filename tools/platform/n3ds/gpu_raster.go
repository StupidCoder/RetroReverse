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
}

// draw executes one draw trigger against the current register state.
func (g *GPU) draw(indexed bool) {
	if g.checkUnsupported() {
		return
	}
	m := g.m

	base := m.gpuAddrToVirt(g.Regs[regAttrBase] << 3)
	fmtWord := uint64(g.Regs[regAttrFmtLow]) | uint64(g.Regs[regAttrFmtHigh]&0xFFFF)<<32
	fixedMask := g.Regs[regAttrFmtHigh] >> 16 & 0xFFF

	// Loader buffers: offset, component list (attribute indices; 12-15 are
	// padding), stride.
	type loaderBuf struct {
		off    uint32
		comps  []int
		stride uint32
	}
	var bufs []loaderBuf
	for b := 0; b < 12; b++ {
		cfg1 := g.Regs[0x204+uint32(b)*3]
		cfg2 := g.Regs[0x205+uint32(b)*3]
		n := int(cfg2 >> 28)
		if n == 0 {
			continue
		}
		perm := uint64(cfg1) | uint64(cfg2&0xFFFF)<<32
		comps := make([]int, n)
		for j := range comps {
			comps[j] = int(perm >> (4 * uint(j)) & 0xF)
		}
		bufs = append(bufs, loaderBuf{off: g.Regs[0x203+uint32(b)*3], comps: comps, stride: cfg2 >> 16 & 0xFF})
	}

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
		fmt.Printf("gpu draw %d: indexed=%v count=%d first=%d base=0x%08X bufs=%d prim=%d vp=%.1fx%.1f fb=0x%08X\n",
			g.Draws, indexed, count, first, base, len(bufs), g.Regs[regPrimConfig]>>8&3,
			f24bits(g.Regs[regViewportWidth]), f24bits(g.Regs[regViewportHeight]), g.Regs[regColorbufLoc]<<3)
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

	// Run every vertex through the shader, then assemble.
	outs := make([]vsOut, 0, count)
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
			for _, c := range b.comps {
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
		var v [16][4]float32
		for j := 0; j < 16; j++ {
			v[j] = attrs[inPerm>>(4*uint(j))&0xF]
		}
		o, ok := g.shaderRun(&v)
		if !ok {
			return
		}
		outs = append(outs, g.mapOutputs(&o))
		if trace && i < 4 {
			r := &outs[len(outs)-1]
			fmt.Printf("  v%-3d in=%v  clip=%v col=%v uv0=%v\n", vi, attrs[0], r.pos, r.color, r.uv[0])
		}
	}

	// Primitive assembly. 0x25E bits 8-9: triangles / strip / fan; mode 3 is
	// the geometry-shader path (already rejected above via 0x229, but keep the
	// check local too).
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

// mapOutputs routes the shader's output registers to their semantics via the
// output map (registers 0x050-0x056): each mapped output register's word
// holds one semantic byte per component.
func (g *GPU) mapOutputs(o *[16][4]float32) vsOut {
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
				// 0x04-0x07 quaternion, 0x10 texcoord0.w, 0x12-0x14 view,
				// 0x1F unused — not consumed by this pipeline.
			}
		}
	}
	return out
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
}

func (g *GPU) fbstate() fbState {
	dim := g.Regs[regFramebufDim]
	dcm := g.Regs[regDepthColorMask]
	return fbState{
		colorAddr:  g.m.gpuAddrToVirt(g.Regs[regColorbufLoc] << 3),
		depthAddr:  g.m.gpuAddrToVirt(g.Regs[regDepthbufLoc] << 3),
		width:      dim & 0x7FF,
		height:     (dim >> 12 & 0x3FF) + 1,
		depthTest:  dcm&1 != 0,
		depthFunc:  dcm >> 4 & 7,
		colorMask:  dcm >> 8 & 0xF,
		depthWr:    dcm>>12&1 != 0,
		vpHalfW:    f24bits(g.Regs[regViewportWidth]),
		vpHalfH:    f24bits(g.Regs[regViewportHeight]),
		depthScale: f24bits(g.Regs[regDepthMapScale]),
		depthOff:   f24bits(g.Regs[regDepthMapOffset]),
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

	type sv struct {
		x, y, z, iw float32
		col         [4]float32
		uv          [3][2]float32
	}
	toScreen := func(v *vsOut) sv {
		iw := 1 / v.pos[3]
		return sv{
			x:  (v.pos[0]*iw + 1) * fb.vpHalfW,
			y:  (v.pos[1]*iw + 1) * fb.vpHalfH,
			z:  v.pos[2] * iw,
			iw: iw,
			col: v.color,
			uv:  v.uv,
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

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			px, py := float32(x)+0.5, float32(y)+0.5
			w0 := edge(v1.x, v1.y, v2.x, v2.y, px, py)
			w1 := edge(v2.x, v2.y, v0.x, v0.y, px, py)
			w2 := edge(v0.x, v0.y, v1.x, v1.y, px, py)
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			l0, l1, l2 := w0/area, w1/area, w2/area

			// Depth: interpolated linearly in screen space (PICA's depth is
			// ndc z, not 1/w), mapped by the depth-map scale/offset.
			z := l0*v0.z + l1*v1.z + l2*v2.z
			depth := z*fb.depthScale + fb.depthOff
			if depth < 0 {
				depth = 0
			}
			if depth > 1 {
				depth = 1
			}
			if !g.depthCompare(&fb, uint32(x), uint32(y), depth) {
				g.DepthKilled++
				continue
			}

			// Perspective-correct attributes via 1/w.
			iw := l0*v0.iw + l1*v1.iw + l2*v2.iw
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

			r8, g8, b8, a8, discard, ok := g.fragment(col, uv)
			if !ok {
				return // fragment stage halted
			}
			if discard {
				continue // alpha-tested out: no colour and no depth write
			}
			g.depthWrite(&fb, uint32(x), uint32(y), depth)
			g.writePixel(&fb, uint32(x), uint32(y), r8, g8, b8, a8)
			g.PixelsDrawn++
		}
	}
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
func (g *GPU) depthCompare(fb *fbState, x, y uint32, depth float32) bool {
	if !fb.depthTest {
		return true
	}
	p := fb.depthAddr + tiledOffset(x, y, fb.width)
	old := uint32(g.m.Read(p)) | uint32(g.m.Read(p+1))<<8 | uint32(g.m.Read(p+2))<<16
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

func (g *GPU) depthWrite(fb *fbState, x, y uint32, depth float32) {
	if !fb.depthTest || !fb.depthWr {
		return
	}
	p := fb.depthAddr + tiledOffset(x, y, fb.width)
	nv := uint32(depth * 0xFFFFFF)
	g.m.Write(p, byte(nv))
	g.m.Write(p+1, byte(nv>>8))
	g.m.Write(p+2, byte(nv>>16))
}

// writePixel blends the fragment with the destination and stores it in the
// tiled RGBA8 colour buffer (bytes [A,B,G,R]).
func (g *GPU) writePixel(fb *fbState, x, y uint32, r, gr, b, a uint8) {
	p := fb.colorAddr + tiledOffset(x, y, fb.width)
	da, db := g.m.Read(p), g.m.Read(p+1)
	dg, dr := g.m.Read(p+2), g.m.Read(p+3)
	r, gr, b, a = g.blend(r, gr, b, a, dr, dg, db, da)
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
