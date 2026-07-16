package gc

// gpu_raster.go fills triangles into the embedded framebuffer. It is the setup/rasteriser at
// the end of the graphics pipe: for each triangle it walks the pixels of the triangle's
// bounding box, keeps the ones inside all three edges, interpolates depth, colour and texture
// coordinate across them, runs each surviving pixel through the TEV (gpu_tev.go) to get its
// final colour, and blends that into the framebuffer under the depth test.
//
// Perspective-correct interpolation is still a later refinement: the barycentric weights are
// taken in screen space, which is exact for the screen-aligned blits the boot draws and only
// skews texturing on triangles seen at a steep angle. Two other faithfulness gaps are
// deliberate and named rather than hidden: back-face culling is not applied, so a triangle of
// either winding is drawn (a wrong guess at winding would make geometry silently vanish, which
// is worse than drawing a back face while the pipe is young). The depth test honours the mode
// the game programmed (BP 0x40 — enable, compare function, write enable).

import (
	"fmt"
	"os"
)

// pixDbg, when set to "x,y", logs every draw that touches that one pixel: the interpolated
// inputs the TEV saw and what it produced — the surgical instrument for a single wrong pixel.
var pixDbgX, pixDbgY = func() (int, int) {
	s := os.Getenv("RR_GC_PIXDBG")
	if s == "" {
		return -1, -1
	}
	var x, y int
	if _, err := fmt.Sscanf(s, "%d,%d", &x, &y); err != nil {
		return -1, -1
	}
	return x, y
}()

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
func (g *gpu) drawTriangle(m *Machine, v0, v1, v2 screenVertex) {
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

	// The depth mode the game programmed (BP 0x40) — layout pinned from the game's own
	// GXSetZMode at 0x801F8AE4: the compare enable at bit 0, the compare function at bits
	// 1..3 (the same eight-code enum the alpha test uses), the write enable at bit 4. With
	// the compare disabled the depth buffer is out of the pipeline entirely: every pixel
	// passes and none is recorded.
	zm := g.BP[0x40]
	zEnable := zm&1 != 0
	zFunc := int((zm >> 1) & 7)
	zWrite := zEnable && zm&(1<<4) != 0

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
			if zEnable && !depthCompare(z, g.ZBuf[idx], zFunc) {
				g.pixZRej++
				if m.OnPixel != nil {
					m.OnPixel(x, y, PixelEvent{})
				}
				continue
			}

			// The interpolated rasteriser colour (the vertex colour) and texture coordinate,
			// which the TEV combines into the final pixel.
			r := uint8(b0*float32(v0.r) + b1*float32(v1.r) + b2*float32(v2.r))
			gg := uint8(b0*float32(v0.g) + b1*float32(v1.g) + b2*float32(v2.g))
			bb := uint8(b0*float32(v0.b) + b1*float32(v1.b) + b2*float32(v2.b))
			a := uint8(b0*float32(v0.a) + b1*float32(v1.a) + b2*float32(v2.a))
			u := b0*v0.u + b1*v1.u + b2*v2.u
			v := b0*v0.v + b1*v1.v + b2*v2.v

			fr, fg, fb, fa, pass := g.shade(m, r, gg, bb, a, u, v)
			if x == pixDbgX && y == pixDbgY {
				t0 := g.texSetup(0)
				tr, tg, tb, ta := g.sampleTexmap(m, 0, u, v)
				fmt.Fprintf(os.Stderr,
					"PIXDBG (%d,%d): ras %d,%d,%d,%d uv (%.4f,%.4f) tex0 0x%06X fmt%X %dx%d texel (%d,%d)=%d,%d,%d,%d -> out %d,%d,%d,%d pass=%v dst %08X\n",
					x, y, r, gg, bb, a, u, v, t0.base, t0.format, t0.width, t0.height,
					wrapCoord(int(u*float32(t0.width)), t0.width, t0.wrapS),
					wrapCoord(int(v*float32(t0.height)), t0.height, t0.wrapT),
					tr, tg, tb, ta, fr, fg, fb, fa, pass, g.EFB[idx])
			}
			if !pass { // the alpha test rejected the pixel
				g.pixARej++
				if m.OnPixel != nil {
					m.OnPixel(x, y, PixelEvent{R: fr, G: fg, B: fb, A: fa})
				}
				continue
			}

			g.EFB[idx] = g.blend(g.EFB[idx], fr, fg, fb, fa)
			if zWrite {
				g.ZBuf[idx] = z
			}
			g.pixWritten++
			if m.OnPixel != nil {
				m.OnPixel(x, y, PixelEvent{R: fr, G: fg, B: fb, A: fa, Drawn: true})
			}
		}
	}
}

// depthCompare applies the zmode compare function: does the incoming depth pass against what
// the buffer holds? The codes are the shared GX compare enum, smaller depths nearer.
func depthCompare(z, buf uint32, comp int) bool {
	switch comp {
	case 0: // NEVER
		return false
	case 1: // LESS
		return z < buf
	case 2: // EQUAL
		return z == buf
	case 3: // LEQUAL
		return z <= buf
	case 4: // GREATER
		return z > buf
	case 5: // NEQUAL
		return z != buf
	case 6: // GEQUAL
		return z >= buf
	default: // ALWAYS
		return true
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
