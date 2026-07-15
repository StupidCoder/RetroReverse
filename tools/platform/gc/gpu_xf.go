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
// the framebuffer keeps, and the colour the rasteriser interpolates across the triangle.
type screenVertex struct {
	x, y, z    float32
	r, g, b, a uint8
}

// transform takes a model-space position through the position matrix, the projection, the
// perspective divide, and the viewport, and returns the pixel position and depth. The matrix
// index is a row into XF memory (four words to a row) — from the current matrix-index register
// (CP 0x30) for most draws, or from a per-vertex index when the descriptor carries one.
func (g *gpu) transform(mtxIdx int, mx, my, mz float32) (sx, sy, sz float32) {
	// Position matrix: three rows of four, at the row the matrix index selects.
	base := mtxIdx * 4
	ex := g.xfFloat(base+0)*mx + g.xfFloat(base+1)*my + g.xfFloat(base+2)*mz + g.xfFloat(base+3)
	ey := g.xfFloat(base+4)*mx + g.xfFloat(base+5)*my + g.xfFloat(base+6)*mz + g.xfFloat(base+7)
	ez := g.xfFloat(base+8)*mx + g.xfFloat(base+9)*my + g.xfFloat(base+10)*mz + g.xfFloat(base+11)

	// Projection: six stored values and a type word — perspective (0) or orthographic (1).
	p0 := g.xfFloat(0x1020)
	p1 := g.xfFloat(0x1021)
	p2 := g.xfFloat(0x1022)
	p3 := g.xfFloat(0x1023)
	p4 := g.xfFloat(0x1024)
	p5 := g.xfFloat(0x1025)

	var cx, cy, cz, cw float32
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
