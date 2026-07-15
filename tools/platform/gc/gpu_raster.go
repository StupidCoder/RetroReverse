package gc

// gpu_raster.go fills triangles into the embedded framebuffer. It is the plainest form of the
// setup/rasteriser at the end of the graphics pipe: for each triangle it walks the pixels of
// the triangle's bounding box, keeps the ones inside all three edges, interpolates depth and
// colour across them, and writes the ones that pass the depth test. Perspective-correct
// interpolation, texture sampling and the full TEV blend are later stages; this stage draws
// the geometry's silhouette in its interpolated vertex colour, which is what proves the vertex
// fetch and the transform put the triangle where the hardware would.
//
// Two faithfulness gaps are deliberate and named here rather than hidden: back-face culling is
// not applied, so a triangle of either winding is drawn (a wrong guess at winding would make
// geometry silently vanish, which is worse than drawing a back face while the pipe is young),
// and the depth test is a fixed less-or-equal with depth writes on, the common default, rather
// than the mode the game programmed. Both become real once a frame needs them to look right.

// ensureRaster makes sure the embedded framebuffer and its depth buffer exist before a draw.
func (g *gpu) ensureRaster() {
	g.ensureEFB()
	if g.ZBuf == nil {
		g.ZBuf = make([]uint32, efbWidth*efbHeight)
		for i := range g.ZBuf {
			g.ZBuf[i] = 0x00FFFFFF // the far plane, in the 24-bit depth range
		}
	}
}

// drawTriangle rasterises one triangle of three transformed vertices.
func (g *gpu) drawTriangle(v0, v1, v2 screenVertex) {
	g.ensureRaster()

	// The bounding box, clamped to the framebuffer.
	minX := int(min3(v0.x, v1.x, v2.x))
	maxX := int(max3(v0.x, v1.x, v2.x)) + 1
	minY := int(min3(v0.y, v1.y, v2.y))
	maxY := int(max3(v0.y, v1.y, v2.y)) + 1
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > efbWidth {
		maxX = efbWidth
	}
	if maxY > efbHeight {
		maxY = efbHeight
	}

	// The signed area of the triangle; a degenerate (zero-area) triangle covers no pixels.
	area := edge(v0.x, v0.y, v1.x, v1.y, v2.x, v2.y)
	if area == 0 {
		return
	}

	for y := minY; y < maxY; y++ {
		py := float32(y) + 0.5
		for x := minX; x < maxX; x++ {
			px := float32(x) + 0.5

			// The three edge functions are the barycentric weights, up to the area. A pixel is
			// inside when all three share the triangle's sign, which covers both windings.
			w0 := edge(v1.x, v1.y, v2.x, v2.y, px, py)
			w1 := edge(v2.x, v2.y, v0.x, v0.y, px, py)
			w2 := edge(v0.x, v0.y, v1.x, v1.y, px, py)
			if (w0 < 0 || w1 < 0 || w2 < 0) && (w0 > 0 || w1 > 0 || w2 > 0) {
				continue
			}

			b0 := w0 / area
			b1 := w1 / area
			b2 := w2 / area

			z := uint32(b0*v0.z + b1*v1.z + b2*v2.z)
			idx := y*efbWidth + x
			if z > g.ZBuf[idx] { // fixed less-or-equal depth test
				continue
			}

			r := uint8(b0*float32(v0.r) + b1*float32(v1.r) + b2*float32(v2.r))
			gg := uint8(b0*float32(v0.g) + b1*float32(v1.g) + b2*float32(v2.g))
			bb := uint8(b0*float32(v0.b) + b1*float32(v1.b) + b2*float32(v2.b))
			a := uint8(b0*float32(v0.a) + b1*float32(v1.a) + b2*float32(v2.a))

			g.EFB[idx] = packRGBA(r, gg, bb, a)
			g.ZBuf[idx] = z
		}
	}
}

// edge is the signed area of the parallelogram spanned by (ax,ay)->(bx,by) and
// (ax,ay)->(cx,cy): positive when c is to one side of the a->b edge, negative to the other.
func edge(ax, ay, bx, by, cx, cy float32) float32 {
	return (bx-ax)*(cy-ay) - (by-ay)*(cx-ax)
}

func min3(a, b, c float32) float32 { return minf(minf(a, b), c) }
func max3(a, b, c float32) float32 { return maxf(maxf(a, b), c) }

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
