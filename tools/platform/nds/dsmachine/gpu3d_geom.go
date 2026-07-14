package dsmachine

// The 3D engine's geometry pipeline: the matrix stacks, the vertex assembler, the
// lighting unit, and the clipper. This is the half of the GPU that turns commands
// into polygons; gpu3d_raster.go turns polygons into pixels.
//
// Everything here is fixed point, and the format is not incidental — it is the
// hardware's arithmetic, and a game's geometry is built to survive exactly its
// rounding. Matrices are 20.12 (a 32-bit integer with twelve fractional bits);
// vertex coordinates arrive as 4.12; texture coordinates are 12.4. Products are
// accumulated in 64 bits and shifted back down by twelve, because a 20.12 times a
// 20.12 is a 40.24 and the top of it is real.
//
// The DS transforms row vectors: a vertex is a row, and a transform post-multiplies
// it (v' = v x M), so matrix storage is row-major with the translation in elements
// 12..14. It follows that MTX_MULT means `current = param x current` — the new
// matrix applies to the vertex *first* — which is what makes hierarchical model
// transforms compose in the order a game expects.

// mtx is a 4x4 matrix of 20.12 fixed point, row-major.
type mtx struct{ m [16]int32 }

func identity() mtx {
	var r mtx
	r.m[0], r.m[5], r.m[10], r.m[15] = 1<<12, 1<<12, 1<<12, 1<<12
	return r
}

// mul returns a x b, in the row-vector convention: (v x a) x b == v x (a.mul(b)).
func (a mtx) mul(b mtx) mtx {
	var r mtx
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var acc int64
			for k := 0; k < 4; k++ {
				acc += int64(a.m[i*4+k]) * int64(b.m[k*4+j])
			}
			r.m[i*4+j] = int32(acc >> 12)
		}
	}
	return r
}

// apply transforms a 4-component row vector by the matrix, accumulating in 64 bits.
func (a mtx) apply(x, y, z, w int64) (int64, int64, int64, int64) {
	c := func(j int) int64 {
		return (x*int64(a.m[0*4+j]) + y*int64(a.m[1*4+j]) +
			z*int64(a.m[2*4+j]) + w*int64(a.m[3*4+j])) >> 12
	}
	return c(0), c(1), c(2), c(3)
}

// gxVertex is a vertex after transformation: clip-space position, colour, and
// texture coordinates. It is what the clipper interpolates and the rasteriser
// consumes.
type gxVertex struct {
	x, y, z, w int64 // clip space, 20.12
	r, g, b    int32 // 0..63, the DS's internal six bits per channel
	s, t       int32 // texture coordinates, 12.4 (sixteenths of a texel)
}

// gxPolygon is one assembled, clipped polygon.
type gxPolygon struct {
	verts    []gxVertex
	attr     uint32 // POLYGON_ATTR as latched when the polygon began
	texParam uint32 // TEXIMAGE_PARAM
	pltt     uint32 // PLTT_BASE
	wbuffer  bool   // depth-buffer the W coordinate rather than Z

	// cmd is the index of the geometry command that completed this polygon — the
	// vertex that closed it. It is what lets a pixel name the command that drew it.
	cmd int
}

// Matrix modes.
const (
	mtxProjection = 0
	mtxPosition   = 1
	mtxPosVec     = 2
	mtxTexture    = 3
)

// geom is the geometry engine's state: the four current matrices, their stacks, the
// vertex assembler's in-progress primitive, the lighting parameters, and the
// polygon/vertex lists being built for the frame.
type geom struct {
	mode int

	proj, pos, vec, tex mtx

	projStack [1]mtx
	texStack  [1]mtx
	posStack  [32]mtx
	vecStack  [32]mtx
	sp        int // the position/vector stacks share one pointer
	projSP    int
	texSP     int

	clipDirty bool
	clipMtx   mtx

	// The vertex assembler.
	begun    bool
	primMode int        // BEGIN_VTXS: 0 tris, 1 quads, 2 tri strip, 3 quad strip
	strip    []gxVertex // vertices accumulated for the current primitive
	stripLen int        // how many vertices the strip has seen in total
	lastVtx  [3]int32   // the last vertex's raw coordinates, for VTX_DIFF/VTX_XY etc.

	// Per-vertex state, latched by the commands that set it.
	color    [3]int32 // 0..63
	texS     int32
	texT     int32
	attr     uint32 // POLYGON_ATTR, latched at BEGIN_VTXS
	attrNext uint32 // the value the next BEGIN_VTXS will latch
	texParam uint32
	pltt     uint32

	// Lighting.
	lightVec   [4][3]int32 // direction, already transformed by the vector matrix
	lightColor [4][3]int32 // 0..63
	diffuse    [3]int32
	ambient    [3]int32
	specular   [3]int32
	emission   [3]int32
	shininess  [128]uint8

	viewX1, viewY1, viewX2, viewY2 int32

	// The frame being built, and the frame the rasteriser will draw.
	verts []gxVertex
	polys []gxPolygon

	posResult [4]int32 // POS_TEST result
	vecResult [3]int32 // VEC_TEST result
	wbuffer   bool

	// Bring-up counters: how many primitives the assembler produced, and how many the
	// clipper threw away. "The screen is black" is not a diagnosis; "the assembler made
	// 900 triangles and the clipper rejected 897 of them" is.
	nEmit, nClipped int
}

func (g *geom) reset() {
	g.proj, g.pos, g.vec, g.tex = identity(), identity(), identity(), identity()
	g.clipDirty = true
	g.viewX2, g.viewY2 = 255, 191
	g.attrNext = 0
}

func (g *geom) beginFrame() {
	g.verts = g.verts[:0]
	g.polys = g.polys[:0]
}

func (g *geom) stackDepth() int { return g.sp }

// clip returns the clip matrix (position x projection), recomputed only when one of
// its inputs has changed. A game touches the position matrix per model and reads the
// clip matrix back rarely, so caching it is worth the flag.
func (g *geom) clip() mtx {
	if g.clipDirty {
		g.clipMtx = g.pos.mul(g.proj)
		g.clipDirty = false
	}
	return g.clipMtx
}

// vecMtx3x3 exposes the vector matrix as the 3x3 the VECMTX_RESULT ports report.
func (g *geom) vecMtx3x3(i uint32) int32 {
	row, col := i/3, i%3
	return g.vec.m[row*4+col]
}

// current returns pointers to the matrices the current mode's operations affect.
// In "position & vector" mode a single command updates both, which is how a game
// keeps normals in step with positions without issuing every transform twice.
func (g *geom) current() []*mtx {
	switch g.mode {
	case mtxProjection:
		return []*mtx{&g.proj}
	case mtxPosition:
		return []*mtx{&g.pos}
	case mtxPosVec:
		return []*mtx{&g.pos, &g.vec}
	default:
		return []*mtx{&g.tex}
	}
}

// exec runs one geometry command with its parameters.
func (g *gpu3d) exec(cmd uint8, p []uint32) {
	// The command scrubber's halt, thrown BEFORE the command runs: a limit of k leaves
	// the machine holding exactly the geometry of commands 0..k-1. See debug.go for why
	// this is a panic and not a flag.
	if g.limit >= 0 && g.count >= g.limit {
		panic(gxHalt{})
	}
	g.cur = g.count
	g.count++

	g.cmdHist[cmd]++
	if g.OnCmd != nil {
		g.OnCmd(cmd, p)
	}
	ge := &g.geom
	switch cmd {
	case 0x10: // MTX_MODE
		ge.mode = int(p[0] & 3)

	case 0x11: // MTX_PUSH
		switch ge.mode {
		case mtxProjection:
			ge.projStack[0] = ge.proj
			ge.projSP = 1
		case mtxTexture:
			ge.texStack[0] = ge.tex
			ge.texSP = 1
		default:
			if ge.sp < 31 {
				ge.posStack[ge.sp] = ge.pos
				ge.vecStack[ge.sp] = ge.vec
				ge.sp++
			}
		}

	case 0x12: // MTX_POP
		switch ge.mode {
		case mtxProjection:
			ge.proj = ge.projStack[0]
			ge.projSP = 0
			ge.clipDirty = true
		case mtxTexture:
			ge.tex = ge.texStack[0]
			ge.texSP = 0
		default:
			// The pop count is a *signed* 6-bit offset, so a game can pop several levels
			// at once — or push back down, which some hierarchies do.
			n := int(p[0] & 0x3F)
			if n >= 32 {
				n -= 64
			}
			ge.sp -= n
			if ge.sp < 0 {
				ge.sp = 0
			}
			if ge.sp > 31 {
				ge.sp = 31
			}
			ge.pos = ge.posStack[ge.sp]
			ge.vec = ge.vecStack[ge.sp]
			ge.clipDirty = true
		}

	case 0x13: // MTX_STORE
		i := int(p[0] & 31)
		switch ge.mode {
		case mtxProjection:
			ge.projStack[0] = ge.proj
		case mtxTexture:
			ge.texStack[0] = ge.tex
		default:
			ge.posStack[i] = ge.pos
			ge.vecStack[i] = ge.vec
		}

	case 0x14: // MTX_RESTORE
		i := int(p[0] & 31)
		switch ge.mode {
		case mtxProjection:
			ge.proj = ge.projStack[0]
			ge.clipDirty = true
		case mtxTexture:
			ge.tex = ge.texStack[0]
		default:
			ge.pos = ge.posStack[i]
			ge.vec = ge.vecStack[i]
			ge.clipDirty = true
		}

	case 0x15: // MTX_IDENTITY
		for _, m := range ge.current() {
			*m = identity()
		}
		ge.clipDirty = true

	case 0x16, 0x18: // MTX_LOAD_4x4 / MTX_MULT_4x4
		var n mtx
		for i := 0; i < 16; i++ {
			n.m[i] = int32(p[i])
		}
		ge.applyMatrix(cmd == 0x16, n)

	case 0x17, 0x19: // MTX_LOAD_4x3 / MTX_MULT_4x3
		// A 4x3 matrix omits the rightmost column, which is the projective one: it is
		// implicitly (0,0,0,1).
		n := identity()
		for i := 0; i < 4; i++ {
			for j := 0; j < 3; j++ {
				n.m[i*4+j] = int32(p[i*3+j])
			}
			n.m[i*4+3] = 0
		}
		n.m[15] = 1 << 12
		ge.applyMatrix(cmd == 0x17, n)

	case 0x1A: // MTX_MULT_3x3 — a pure rotation/scale, no translation
		n := identity()
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				n.m[i*4+j] = int32(p[i*3+j])
			}
		}
		ge.applyMatrix(false, n)

	case 0x1B: // MTX_SCALE — position only, never the vector matrix: scaling normals
		// would change their length, and the lighting unit assumes they are unit.
		n := identity()
		n.m[0], n.m[5], n.m[10] = int32(p[0]), int32(p[1]), int32(p[2])
		if ge.mode == mtxProjection {
			ge.proj = n.mul(ge.proj)
		} else if ge.mode == mtxTexture {
			ge.tex = n.mul(ge.tex)
		} else {
			ge.pos = n.mul(ge.pos)
		}
		ge.clipDirty = true

	case 0x1C: // MTX_TRANS
		n := identity()
		n.m[12], n.m[13], n.m[14] = int32(p[0]), int32(p[1]), int32(p[2])
		ge.applyMatrix(false, n)

	case 0x20: // COLOR
		ge.color = unpackRGB5(p[0])

	case 0x21: // NORMAL
		ge.normal(p[0])

	case 0x22: // TEXCOORD
		ge.texS = int32(int16(uint16(p[0])))
		ge.texT = int32(int16(uint16(p[0] >> 16)))

	case 0x23: // VTX_16 — three 16-bit coordinates in 4.12
		x := int32(int16(uint16(p[0])))
		y := int32(int16(uint16(p[0] >> 16)))
		z := int32(int16(uint16(p[1])))
		ge.vertex(g, x, y, z)

	case 0x24: // VTX_10 — three 10-bit coordinates in 4.6, so they shift up by six
		x := sext(int32(p[0]), 10) << 6
		y := sext(int32(p[0]>>10), 10) << 6
		z := sext(int32(p[0]>>20), 10) << 6
		ge.vertex(g, x, y, z)

	case 0x25: // VTX_XY — the omitted coordinate keeps its previous value
		ge.vertex(g, int32(int16(uint16(p[0]))), int32(int16(uint16(p[0]>>16))), ge.lastVtx[2])
	case 0x26: // VTX_XZ
		ge.vertex(g, int32(int16(uint16(p[0]))), ge.lastVtx[1], int32(int16(uint16(p[0]>>16))))
	case 0x27: // VTX_YZ
		ge.vertex(g, ge.lastVtx[0], int32(int16(uint16(p[0]))), int32(int16(uint16(p[0]>>16))))

	case 0x28: // VTX_DIFF — a small signed delta from the last vertex, in 1.9 of a
		// 4.12 unit: the compact form a model's dense strips are stored in.
		dx := sext(int32(p[0]), 10)
		dy := sext(int32(p[0]>>10), 10)
		dz := sext(int32(p[0]>>20), 10)
		ge.vertex(g, ge.lastVtx[0]+dx, ge.lastVtx[1]+dy, ge.lastVtx[2]+dz)

	case 0x29: // POLYGON_ATTR — latched, and only takes effect at the next BEGIN_VTXS
		ge.attrNext = p[0]

	case 0x2A: // TEXIMAGE_PARAM
		ge.texParam = p[0]

	case 0x2B: // PLTT_BASE
		ge.pltt = p[0]

	case 0x30: // DIF_AMB
		ge.diffuse = unpackRGB5(p[0])
		ge.ambient = unpackRGB5(p[0] >> 16)
		if p[0]&(1<<15) != 0 { // "set vertex colour to the diffuse colour"
			ge.color = ge.diffuse
		}

	case 0x31: // SPE_EMI
		ge.specular = unpackRGB5(p[0])
		ge.emission = unpackRGB5(p[0] >> 16)

	case 0x32: // LIGHT_VECTOR — the direction is transformed by the vector matrix, so
		// a light set up in world space follows the camera without the game re-issuing it.
		n := int(p[0] >> 30)
		x := int64(sext(int32(p[0]), 10) << 3)
		y := int64(sext(int32(p[0]>>10), 10) << 3)
		z := int64(sext(int32(p[0]>>20), 10) << 3)
		tx, ty, tz, _ := ge.vec.apply(x, y, z, 0)
		ge.lightVec[n] = [3]int32{int32(tx), int32(ty), int32(tz)}

	case 0x33: // LIGHT_COLOR
		ge.lightColor[p[0]>>30] = unpackRGB5(p[0])

	case 0x34: // SHININESS — a 128-entry specular falloff table, two bytes per word
		for i, w := range p {
			for j := 0; j < 4; j++ {
				ge.shininess[i*4+j] = uint8(w >> (8 * j))
			}
		}

	case 0x40: // BEGIN_VTXS
		ge.begun = true
		ge.primMode = int(p[0] & 3)
		ge.strip = ge.strip[:0]
		ge.stripLen = 0
		ge.attr = ge.attrNext

	case 0x41: // END_VTXS
		ge.begun = false

	case 0x50: // SWAP_BUFFERS — the frame is complete. The hardware does not present it
		// now: it presents it at the next vertical blank (see gpu3d.vblank).
		g.swapPending = true
		g.swapMode = p[0]
		ge.wbuffer = p[0]&2 != 0

	case 0x60: // VIEWPORT
		ge.viewX1 = int32(p[0] & 0xFF)
		ge.viewY1 = int32((p[0] >> 8) & 0xFF)
		ge.viewX2 = int32((p[0] >> 16) & 0xFF)
		ge.viewY2 = int32((p[0] >> 24) & 0xFF)

	case 0x70: // BOX_TEST — reports whether a box is visible. Answering "yes" always is
		// safe (the game draws something it could have culled) and never wrong; answering
		// "no" wrongly would make geometry vanish.
		g.regs[regGXSTAT] |= 1 << 1

	case 0x71: // POS_TEST
		x := int64(int16(uint16(p[0])))
		y := int64(int16(uint16(p[0] >> 16)))
		z := int64(int16(uint16(p[1])))
		cx, cy, cz, cw := ge.clip().apply(x, y, z, 1<<12)
		ge.posResult = [4]int32{int32(cx), int32(cy), int32(cz), int32(cw)}

	case 0x72: // VEC_TEST
		x := int64(sext(int32(p[0]), 10) << 3)
		y := int64(sext(int32(p[0]>>10), 10) << 3)
		z := int64(sext(int32(p[0]>>20), 10) << 3)
		vx, vy, vz, _ := ge.vec.apply(x, y, z, 0)
		ge.vecResult = [3]int32{int32(vx), int32(vy), int32(vz)}
	}
}

// applyMatrix either loads a matrix into the current one(s) or multiplies into them.
func (g *geom) applyMatrix(load bool, n mtx) {
	for _, m := range g.current() {
		if load {
			*m = n
		} else {
			*m = n.mul(*m)
		}
	}
	g.clipDirty = true
}

// sext sign-extends an n-bit field.
func sext(v int32, n uint) int32 {
	sh := 32 - n
	return v << sh >> sh
}

// unpackRGB5 turns a BGR555 word into three 6-bit channels — the DS's internal
// colour depth. The five-bit value is doubled rather than replicated, which is what
// the hardware does, and it means white is 62 and not 63.
func unpackRGB5(v uint32) [3]int32 {
	r := int32(v&0x1F) * 2
	g := int32((v>>5)&0x1F) * 2
	b := int32((v>>10)&0x1F) * 2
	return [3]int32{r, g, b}
}

// normal runs the lighting unit. The DS lights per *vertex*, at the moment the
// NORMAL command is issued — so the colour a vertex gets is the one computed from
// the material and lights that were current then, and the result is simply latched
// as the vertex colour. Lighting is not deferred and there is no per-pixel term.
func (g *geom) normal(v uint32) {
	nx := int64(sext(int32(v), 10) << 3)
	ny := int64(sext(int32(v>>10), 10) << 3)
	nz := int64(sext(int32(v>>20), 10) << 3)

	// If the texture-coordinate transform is in "normal" mode, the normal doubles as
	// the texture coordinate source — this is how the DS does environment mapping.
	if (g.texParam>>30)&3 == 2 {
		s, t := g.texTransformNormal(nx, ny, nz)
		g.texS, g.texT = s, t
	}

	tx, ty, tz, _ := g.vec.apply(nx, ny, nz, 0)

	col := g.emission
	for i := 0; i < 4; i++ {
		if g.attrNext&(1<<uint(i)) == 0 {
			continue
		}
		lv := g.lightVec[i]
		lc := g.lightColor[i]

		// The diffuse term is -dot(light, normal): the DS stores light *directions*
		// pointing away from the surface, so the sign is inverted relative to the usual
		// formulation. Clamp at zero — a surface facing away is unlit, not negatively lit.
		dot := -(int64(lv[0])*tx + int64(lv[1])*ty + int64(lv[2])*tz) >> 12
		if dot < 0 {
			dot = 0
		}
		if dot > 1<<12 {
			dot = 1 << 12
		}

		// The specular term uses the half-vector between the light and the eye, and the
		// eye on the DS is fixed at (0,0,-1) — a design that lets the whole thing be a
		// table lookup rather than a normalisation.
		var shine int64
		hz := int64(lv[2]) - (1 << 12)
		hlen := int64(lv[0])*int64(lv[0]) + int64(lv[1])*int64(lv[1]) + hz*hz
		if hlen > 0 {
			hdot := (int64(lv[0])*tx + int64(lv[1])*ty + hz*tz) >> 12
			if hdot > 0 {
				// (h . n)^2 / |h|^2, then through the shininess table.
				sh := (hdot * hdot << 12) / hlen
				if sh > 1<<12 {
					sh = 1 << 12
				}
				idx := sh >> 5
				if idx > 127 {
					idx = 127
				}
				if g.attrNext&(1<<4) != 0 { // the shininess table is enabled
					shine = int64(g.shininess[idx]) << 4
				} else {
					shine = sh
				}
			}
		}

		for c := 0; c < 3; c++ {
			d := int64(g.diffuse[c]) * int64(lc[c]) * dot >> (6 + 12)
			a := int64(g.ambient[c]) * int64(lc[c]) >> 6
			s := int64(g.specular[c]) * int64(lc[c]) * shine >> (6 + 12)
			col[c] += int32(d + a + s)
		}
	}
	for c := 0; c < 3; c++ {
		if col[c] > 63 {
			col[c] = 63
		}
		if col[c] < 0 {
			col[c] = 0
		}
	}
	g.color = col
}

// texTransformNormal produces texture coordinates from a normal (environment
// mapping): the normal is run through the texture matrix and the result scaled by
// the texture size.
func (g *geom) texTransformNormal(nx, ny, nz int64) (int32, int32) {
	s, t, _, _ := g.tex.apply(nx, ny, nz, 1<<12)
	return int32(s >> 8), int32(t >> 8)
}

// texCoordFor produces the texture coordinates a vertex carries, honouring the
// transform mode in TEXIMAGE_PARAM. Mode 0 passes them through; mode 1 runs them
// through the texture matrix; modes 2 and 3 source them from the normal and the
// vertex position instead (the environment- and shadow-mapping modes), and mode 2 is
// handled when the normal arrives.
func (g *geom) texCoordFor(x, y, z int32) (int32, int32) {
	switch (g.texParam >> 30) & 3 {
	case 1:
		s, t, _, _ := g.tex.apply(int64(g.texS)<<8, int64(g.texT)<<8, 1<<12, 1<<12)
		return int32(s >> 8), int32(t >> 8)
	case 3:
		s, t, _, _ := g.tex.apply(int64(x), int64(y), int64(z), 1<<12)
		return int32(s >> 8), int32(t >> 8)
	}
	return g.texS, g.texT
}

// vertex submits one vertex: it is transformed into clip space, given the colour and
// texture coordinates currently latched, and handed to the primitive assembler.
func (g *geom) vertex(gp *gpu3d, x, y, z int32) {
	g.lastVtx = [3]int32{x, y, z}
	if !g.begun {
		return
	}
	cx, cy, cz, cw := g.clip().apply(int64(x), int64(y), int64(z), 1<<12)
	s, t := g.texCoordFor(x, y, z)
	v := gxVertex{x: cx, y: cy, z: cz, w: cw,
		r: g.color[0], g: g.color[1], b: g.color[2], s: s, t: t}

	g.strip = append(g.strip, v)
	g.stripLen++
	g.emitIfComplete(gp)
}

// emitIfComplete turns the accumulated vertices into polygons as soon as enough have
// arrived. The strip modes are the interesting ones: they emit a polygon per *new*
// vertex once primed, and a triangle strip must alternate its winding, or every
// other triangle faces backwards and is culled — which looks exactly like a model
// with half its surface missing.
func (g *geom) emitIfComplete(gp *gpu3d) {
	switch g.primMode {
	case 0: // separate triangles
		if len(g.strip) == 3 {
			g.emit(gp, g.strip[0], g.strip[1], g.strip[2])
			g.strip = g.strip[:0]
		}
	case 1: // separate quads
		if len(g.strip) == 4 {
			g.emit(gp, g.strip[0], g.strip[1], g.strip[2], g.strip[3])
			g.strip = g.strip[:0]
		}
	case 2: // triangle strip
		if len(g.strip) >= 3 {
			n := len(g.strip)
			a, b, c := g.strip[n-3], g.strip[n-2], g.strip[n-1]
			if (g.stripLen-3)%2 == 1 { // every other triangle has reversed winding
				a, b = b, a
			}
			g.emit(gp, a, b, c)
			if n > 3 {
				g.strip = g.strip[n-2:]
				g.strip = append([]gxVertex{}, g.strip...)
			}
		}
	case 3: // quad strip: each pair of new vertices closes another quad
		if len(g.strip) >= 4 && len(g.strip)%2 == 0 {
			n := len(g.strip)
			// The quad's vertices are in Z order in the stream, so the last two must be
			// swapped to walk the perimeter.
			g.emit(gp, g.strip[n-4], g.strip[n-3], g.strip[n-1], g.strip[n-2])
			g.strip = append([]gxVertex{}, g.strip[n-2:]...)
		}
	}
}

// emit clips a polygon and, if anything survives, adds it to the frame.
func (g *geom) emit(gp *gpu3d, vs ...gxVertex) {
	g.nEmit++
	poly := clipPolygon(vs)
	if len(poly) < 3 {
		g.nClipped++
		return
	}
	if len(g.polys) >= 2048 { // the hardware's polygon RAM holds 2048 per frame
		return
	}
	g.polys = append(g.polys, gxPolygon{
		verts: poly, attr: g.attr, texParam: g.texParam, pltt: g.pltt, wbuffer: g.wbuffer,
		cmd: gp.cur,
	})
}

// clipPolygon clips against the six frustum planes in homogeneous coordinates
// (-w <= x,y,z <= w), by Sutherland-Hodgman. Clipping in clip space rather than
// after the perspective divide is not a refinement: a vertex behind the eye has a
// negative w, and dividing by it turns a polygon inside out — it is the difference
// between geometry that disappears correctly at the edge of the screen and geometry
// that smears across it.
func clipPolygon(vs []gxVertex) []gxVertex {
	out := append([]gxVertex{}, vs...)
	for plane := 0; plane < 6; plane++ {
		if len(out) == 0 {
			return nil
		}
		in := out
		out = out[:0:0]
		for i := range in {
			cur, prev := in[i], in[(i+len(in)-1)%len(in)]
			dc, dp := planeDist(cur, plane), planeDist(prev, plane)
			if dc >= 0 {
				if dp < 0 {
					out = append(out, lerpVertex(prev, cur, dp, dc))
				}
				out = append(out, cur)
			} else if dp >= 0 {
				out = append(out, lerpVertex(prev, cur, dp, dc))
			}
		}
	}
	return out
}

// planeDist is the signed distance to one frustum plane: positive is inside.
func planeDist(v gxVertex, plane int) int64 {
	switch plane {
	case 0:
		return v.w - v.x
	case 1:
		return v.w + v.x
	case 2:
		return v.w - v.y
	case 3:
		return v.w + v.y
	case 4:
		return v.w - v.z
	default:
		return v.w + v.z
	}
}

// lerpVertex finds the point where the edge from a to b crosses a plane, and
// interpolates everything the vertex carries to it.
func lerpVertex(a, b gxVertex, da, db int64) gxVertex {
	den := da - db
	if den == 0 {
		return a
	}
	// t = da / (da - db), kept in 20.12 so the interpolation stays in integers.
	t := (da << 12) / den
	li := func(x, y int64) int64 { return x + (y-x)*t>>12 }
	li32 := func(x, y int32) int32 { return int32(int64(x) + (int64(y)-int64(x))*t>>12) }
	return gxVertex{
		x: li(a.x, b.x), y: li(a.y, b.y), z: li(a.z, b.z), w: li(a.w, b.w),
		r: li32(a.r, b.r), g: li32(a.g, b.g), b: li32(a.b, b.b),
		s: li32(a.s, b.s), t: li32(a.t, b.t),
	}
}
