package gc

// gpu_xf.go is the transform unit — the XF — that turns a vertex's model-space position into
// a screen pixel. It is the fixed-function counterpart to a modern vertex shader: a position
// matrix takes the vertex into eye space, a projection matrix takes it into clip space, the
// perspective divide takes it to normalised device coordinates, and the viewport takes it to
// the pixel grid of the embedded framebuffer. Each of those four steps is a small register
// set the game programmed into XF memory earlier in the command stream, so this file is mostly
// the arithmetic of reading them back and applying them in order.
//
// The register addresses are the hardware's: the position matrices live in XF memory at
// 0x0000, the viewport at 0x101A, and the projection at 0x1020. The command processor stores
// every XF write into xfMem as it walks the stream (see gpu.go); here those raw words are read
// back as the floats they are.

import "math"

// xfFloat reads an XF memory word as the IEEE-754 float it holds.
func (g *gpu) xfFloat(addr int) float32 {
	if addr < 0 || addr >= len(g.XFMem) {
		return 0
	}
	return math.Float32frombits(g.XFMem[addr])
}

// screenVertex is a vertex after transform: its pixel position, its depth in the 24-bit range
// the framebuffer keeps, the colour the rasteriser interpolates across the triangle, and the
// first texture coordinate it interpolates alongside so the TEV can sample a texture per pixel.
//
// invW is the reciprocal of the clip-space w the perspective divide used. Depth needs no such
// help — screen z has already been divided, so it is linear in screen space — but a texture
// coordinate has not, and interpolating one straight across the screen walks it at a constant
// rate over a surface that is receding, which is the classic affine-texturing skew. Keeping
// 1/w per vertex lets the rasteriser interpolate u/w and v/w (which ARE linear in screen space)
// and divide back per pixel.
type screenVertex struct {
	x, y, z    float32
	r, g, b, a uint8
	u, v       float32
	invW       float32
}

// clipVertex is a vertex after projection but before the perspective divide: its homogeneous
// clip-space position, and the attributes the rasteriser will interpolate. Clipping happens
// here rather than in screen space because the divide is only meaningful for a vertex in front
// of the eye — a vertex behind it has a negative w, and dividing by that reflects the vertex
// through the origin, wrapping the triangle across the screen instead of removing it.
type clipVertex struct {
	cx, cy, cz, cw float32
	r, g, b, a     uint8
	u, v           float32
}

// lerpClip interpolates two clip-space vertices at parameter t, in clip space, where the
// straight line between two vertices is still straight. Every attribute the rasteriser reads
// is carried across so a vertex the clipper invents is as complete as one the game supplied.
func lerpClip(a, b clipVertex, t float32) clipVertex {
	li := func(x, y uint8) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t + 0.5) }
	lf := func(x, y float32) float32 { return x + (y-x)*t }
	return clipVertex{
		cx: lf(a.cx, b.cx), cy: lf(a.cy, b.cy), cz: lf(a.cz, b.cz), cw: lf(a.cw, b.cw),
		r: li(a.r, b.r), g: li(a.g, b.g), b: li(a.b, b.b), a: li(a.a, b.a),
		u: lf(a.u, b.u), v: lf(a.v, b.v),
	}
}

// transform takes a model-space position through the position matrix, the projection, the
// perspective divide, and the viewport, and returns the pixel position and depth. The matrix
// index is a row into XF memory (four words to a row) — from the current matrix-index register
// (CP 0x30) for most draws, or from a per-vertex index when the descriptor carries one.
func (g *gpu) transform(mtxIdx int, mx, my, mz float32) (sx, sy, sz float32) {
	ex, ey, ez := g.eyePos(mtxIdx, mx, my, mz)
	return g.project(ex, ey, ez)
}

// eyePos takes a model-space position through the position matrix: three rows of four, at
// the row the matrix index selects. The result is the eye-space position — what the
// projection consumes, and what the lighting stage measures light distances in.
func (g *gpu) eyePos(mtxIdx int, mx, my, mz float32) (ex, ey, ez float32) {
	base := mtxIdx * 4
	ex = g.xfFloat(base+0)*mx + g.xfFloat(base+1)*my + g.xfFloat(base+2)*mz + g.xfFloat(base+3)
	ey = g.xfFloat(base+4)*mx + g.xfFloat(base+5)*my + g.xfFloat(base+6)*mz + g.xfFloat(base+7)
	ez = g.xfFloat(base+8)*mx + g.xfFloat(base+9)*my + g.xfFloat(base+10)*mz + g.xfFloat(base+11)
	return
}

// project takes an eye-space position through the projection, the perspective divide and the
// viewport to the pixel grid. It is the whole chain for a vertex already known to be in front
// of the near plane; a draw goes through clipPos and toScreen either side of the clipper.
func (g *gpu) project(ex, ey, ez float32) (sx, sy, sz float32) {
	cx, cy, cz, cw := g.clipPos(ex, ey, ez)
	return g.toScreen(cx, cy, cz, cw)
}

// clipPos takes an eye-space position through the projection into homogeneous clip space.
//
// The six stored values are the compact form of the projection matrix, and reading the near and
// far planes back out of them is what pins the clip-space convention. For the perspective form
// the hardware stores p4 = -n/(f-n) and p5 = -f*n/(f-n); substituting the near plane ez = -n
// gives cz = -n and cw = n, so cz/cw = -1, and the far plane ez = -f gives cz = 0, so cz/cw = 0.
// Normalised device z therefore runs from -1 at the near plane to 0 at the far plane — which is
// exactly what the viewport registers this game programs expect, their z scale and offset both
// 2^24-1, mapping -1 to depth 0 and 0 to depth 0xFFFFFF. The near plane is thus cz = -cw, and
// that is the plane the clipper cuts against.
func (g *gpu) clipPos(ex, ey, ez float32) (cx, cy, cz, cw float32) {
	p0 := g.xfFloat(0x1020)
	p1 := g.xfFloat(0x1021)
	p2 := g.xfFloat(0x1022)
	p3 := g.xfFloat(0x1023)
	p4 := g.xfFloat(0x1024)
	p5 := g.xfFloat(0x1025)

	if g.XFMem[0x1026] == 0 { // perspective: the projected matrix has -1 in its w row
		cx = p0*ex + p1*ez
		cy = p2*ey + p3*ez
		cz = p4*ez + p5
		cw = -ez
	} else { // orthographic: w stays 1
		cx = p0*ex + p1
		cy = p2*ey + p3
		cz = p4*ez + p5
		cw = 1
	}
	return
}

// toScreen takes a clip-space position through the perspective divide and the viewport to the
// pixel grid. Every vertex reaching it has passed the near clip, so w is positive.
func (g *gpu) toScreen(cx, cy, cz, cw float32) (sx, sy, sz float32) {
	// The perspective divide to normalised device coordinates. A degenerate w leaves the
	// vertex at the origin rather than dividing by zero.
	if cw == 0 {
		cw = 1
	}
	nx := cx / cw
	ny := cy / cw
	nz := cz / cw

	// Viewport: three scales and three offsets. The offsets carry the hardware's fixed 342
	// bias, which the rasteriser's scissor offset removes, so it is subtracted here to land the
	// vertex on the framebuffer's own pixel grid.
	vsx := g.xfFloat(0x101A)
	vsy := g.xfFloat(0x101B)
	vsz := g.xfFloat(0x101C)
	vox := g.xfFloat(0x101D)
	voy := g.xfFloat(0x101E)
	voz := g.xfFloat(0x101F)

	sx = nx*vsx + (vox - 342)
	sy = ny*vsy + (voy - 342)
	sz = nz*vsz + voz
	return sx, sy, sz
}
