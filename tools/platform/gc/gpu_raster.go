package gc

// gpu_raster.go fills triangles into the embedded framebuffer. It is the setup/rasteriser at
// the end of the graphics pipe: for each triangle it walks the pixels of the triangle's
// bounding box, keeps the ones inside all three edges, interpolates depth, colour and texture
// coordinate across them, runs each surviving pixel through the TEV (gpu_tev.go) to get its
// final colour, and blends that into the framebuffer under the depth test.
//
// The barycentric weights are taken in screen space, which is what the depth test wants —
// screen z has already been through the perspective divide, so it is linear here — but not
// what a texture coordinate wants: those are interpolated perspective-correctly, through
// perspUV. The depth test honours the mode the game programmed (BP 0x40 — enable, compare
// function, write enable), and so does back-face culling (GEN_MODE bits 14..15 — see cullTest,
// and note that its sign convention is pinned by a rendered frame because nothing else can pin
// it).
//
// Every triangle arriving here has been through the near clipper (gpu_clip.go), so its vertices
// are all in front of the eye and their w is positive.

import (
	"fmt"
	"os"
	"sync/atomic"
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

// perspTexCoords interpolates a pixel's texture coordinates perspective-correctly from the
// three barycentric weights, and applies each one's projective divide.
//
// What varies linearly across the screen is not s but s/w, so s/w, t/w, q/w and 1/w are the
// things interpolated, and the divide is undone per pixel. Interpolating s directly walks the
// texture at a constant rate across a surface whose steps into the distance are shrinking, so
// the texture slides over the geometry — the skew is invisible on a screen-aligned blit and
// severe on a ground plane running to the horizon.
//
// For an orthographic draw every w is 1, so every invW is 1 and this reduces exactly to the
// affine interpolation it replaced — 2D screens are untouched by construction, not by luck.
//
// The final divide by q is what makes a projective coordinate project: it is the same divide
// the position takes to reach the screen, applied a second time in the OTHER camera's frame,
// and it is the reason a shadow map lands where the light sees it rather than where the eye
// does. A non-projective coordinate reaches here with q = 1 in every vertex (gpu_texgen.go
// forces it), which interpolates to exactly 1, so the divide is an identity rather than a
// case the loop has to branch on.
//
// The weights are non-negative and sum to 1 (the caller has already rejected pixels outside
// the triangle), and the clipper guarantees every invW is positive, so iw is positive.
func perspTexCoords(b0, b1, b2 float32, v0, v1, v2 screenVertex, out *[maxTexCoord]texCoord) {
	iw := b0*v0.invW + b1*v1.invW + b2*v2.invW
	for i := 0; i < v0.ntc; i++ {
		s := (b0*v0.tc[i].s*v0.invW + b1*v1.tc[i].s*v1.invW + b2*v2.tc[i].s*v2.invW) / iw
		t := (b0*v0.tc[i].t*v0.invW + b1*v1.tc[i].t*v1.invW + b2*v2.tc[i].t*v2.invW) / iw
		q := (b0*v0.tc[i].q*v0.invW + b1*v1.tc[i].q*v1.invW + b2*v2.tc[i].q*v2.invW) / iw
		// A vertex behind the light's own near plane brings a q of zero or less through the
		// interpolation. There is no coordinate to sample there; leaving s,t as they stand
		// keeps the pixel out of the divide rather than sending an infinity to the sampler.
		if q != 0 {
			s /= q
			t /= q
		}
		out[i] = texCoord{s: s, t: t, q: q}
	}
}

// rasterTri is one triangle that has survived setup: projected, wound, bounded and not
// culled, carrying the depth mode the draw was issued under. It is the unit the fill works
// on, and it exists so that the fill can be handed a ROW RANGE — which is what makes the
// partition below possible.
type rasterTri struct {
	v0, v1, v2      screenVertex
	area            float32
	minX, maxX      int
	minY, maxY      int
	zEnable, zWrite bool
	zFunc           int
}

// rstats is one worker's tally of where its fragments went. The counters used to be
// incremented straight into the gpu, which is a write, and a write is what a fan-out cannot
// share. Making them atomic would have put a contended cache line in the middle of the
// fragment loop; these are merged after the join, IN WORKER ORDER, so the totals are as
// deterministic as the picture.
type rstats struct {
	written, zRej, aRej int
}

// setupTri projects one triangle into a rasterTri, or reports that it covers nothing.
//
// It stays SERIAL, and it has to: it owns the culled counter, and cullTest reads an
// environment override that exists to answer "is this geometry missing because of culling?".
func (g *gpu) setupTri(v0, v1, v2 screenVertex) (rasterTri, bool) {
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
	// Its SIGN is the triangle's winding on screen, which is what the cull test reads.
	area := edge(v0.x, v0.y, v1.x, v1.y, v2.x, v2.y)
	if area == 0 {
		return rasterTri{}, false
	}
	if g.cullTest(area) {
		g.profCulled++
		return rasterTri{}, false
	}

	// The depth mode the game programmed (BP 0x40) — layout pinned from the game's own
	// GXSetZMode at 0x801F8AE4: the compare enable at bit 0, the compare function at bits
	// 1..3 (the same eight-code enum the alpha test uses), the write enable at bit 4. With
	// the compare disabled the depth buffer is out of the pipeline entirely: every pixel
	// passes and none is recorded.
	zm := g.BP[0x40]
	zEnable := zm&1 != 0

	return rasterTri{
		v0: v0, v1: v1, v2: v2, area: area,
		minX: minX, maxX: maxX, minY: minY, maxY: maxY,
		zEnable: zEnable,
		zFunc:   int((zm >> 1) & 7),
		zWrite:  zEnable && zm&(1<<4) != 0,
	}, true
}

// fillTri rasterises one set-up triangle, but only the rows in [yLo, yHi).
//
// The row range is the whole point: it is what lets a band of the screen be filled by exactly
// one worker. Called serially it is handed the triangle's own bounds and behaves as the
// single loop it replaced.
func (g *gpu) fillTri(m *Machine, tev *tevState, t *rasterTri, yLo, yHi int, st *rstats) {
	v0, v1, v2 := t.v0, t.v1, t.v2
	area := t.area
	zEnable, zFunc, zWrite := t.zEnable, t.zFunc, t.zWrite
	minX, maxX := t.minX, t.maxX

	for y := yLo; y < yHi; y++ {
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
				st.zRej++
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
			var tc [maxTexCoord]texCoord
			perspTexCoords(b0, b1, b2, v0, v1, v2, &tc)

			fr, fg, fb, fa, pass := g.shade(m, tev, r, gg, bb, a, &tc)
			if x == pixDbgX && y == pixDbgY {
				t0 := g.texSetup(0)
				u, v := tc[0].s, tc[0].t
				tr, tg, tb, ta := g.sampleTexmap(m, &t0, u, v)
				fmt.Fprintf(os.Stderr,
					"PIXDBG (%d,%d): ras %d,%d,%d,%d uv (%.4f,%.4f) tex0 0x%06X fmt%X %dx%d texel (%d,%d)=%d,%d,%d,%d -> out %d,%d,%d,%d pass=%v dst %08X\n",
					x, y, r, gg, bb, a, u, v, t0.base, t0.format, t0.width, t0.height,
					wrapCoord(int(u*float32(t0.width)), t0.width, t0.wrapS),
					wrapCoord(int(v*float32(t0.height)), t0.height, t0.wrapT),
					tr, tg, tb, ta, fr, fg, fb, fa, pass, g.EFB[idx])
			}
			if !pass { // the alpha test rejected the pixel
				st.aRej++
				if m.OnPixel != nil {
					m.OnPixel(x, y, PixelEvent{R: fr, G: fg, B: fb, A: fa})
				}
				continue
			}

			g.EFB[idx] = g.blend(g.EFB[idx], fr, fg, fb, fa)
			if zWrite {
				g.ZBuf[idx] = z
			}
			st.written++
			if m.OnPixel != nil {
				m.OnPixel(x, y, PixelEvent{R: fr, G: fg, B: fb, A: fa, Drawn: true})
			}
		}
	}
}

// bandRows is how many screen rows one worker takes at a time.
//
// The 3DS uses eight and deliberately refuses four, even though four measured a hair faster:
// its render target is Morton-tiled and an 8x8 tile is eight rows tall, so a four-row band
// lets two workers write into the same tile and false-share its cache lines. Same speed,
// worse reason.
//
// THAT CONSTRAINT DOES NOT EXIST HERE. Flipper's embedded framebuffer is a flat []uint32
// indexed y*efbWidth+x, not tiled, and a row is 640*4 = 2560 bytes = 40 cache lines exactly —
// so EVERY row boundary is 64-byte aligned and bands of any height never share a line. The
// value is free to be chosen by measurement, so it was: 8/4/2/1 rows over the shadow scene
// measured 1.276/1.252/1.23/1.22 s, and four is the knee. Finer bands balance better here
// because the work is savagely uneven — half of a field's pixels are in 45 of its 4,510
// drawing draws — and cost nothing to hand out.
const bandRows = 4

// fill rasterises a draw's triangles, in parallel when it is faithful to do so.
//
// DETERMINISM COMES FROM THE PARTITION, NOT FROM THE SCHEDULER. A band of rows is filled by
// exactly one worker, so no two workers touch the same pixel; within a band the triangles are
// applied in submission order, because every worker walks tris in index order. The buffer the
// workers leave behind is byte for byte the one the serial loop leaves. Bands are handed out
// from a shared counter rather than dealt out in advance, because geometry is not spread
// evenly down the screen — but WHICH worker takes WHICH band cannot change the result, only
// how long it takes.
func (g *gpu) fill(m *Machine, tev *tevState, tris []rasterTri) {
	if len(tris) == 0 {
		return
	}
	// Before any worker starts: allocating the framebuffer is a write, and it was sitting on
	// the per-triangle path.
	g.ensureRaster()

	workers := g.rasterWorkers(m, tev, tris)
	if workers <= 1 {
		g.profSerFills++
		var st rstats
		for i := range tris {
			t := &tris[i]
			g.fillTri(m, tev, t, t.minY, t.maxY, &st)
		}
		g.mergeStats(&st)
		return
	}
	g.profParFills++

	bands := (efbHeight + bandRows - 1) / bandRows
	stats := make([]rstats, workers)
	var next int32
	g.pool().run(workers, func(w int) {
		st := &stats[w]
		for {
			b := int(atomic.AddInt32(&next, 1)) - 1
			if b >= bands {
				return
			}
			yLo := b * bandRows
			yHi := min(yLo+bandRows, efbHeight)
			for i := range tris {
				t := &tris[i]
				if t.maxY <= yLo || t.minY >= yHi {
					continue
				}
				g.fillTri(m, tev, t, max(t.minY, yLo), min(t.maxY, yHi), st)
			}
		}
	})
	// Merged in worker order, so the totals do not depend on who finished first.
	for i := range stats {
		g.mergeStats(&stats[i])
	}
}

func (g *gpu) mergeStats(st *rstats) {
	g.pixWritten += st.written
	g.pixZRej += st.zRej
	g.pixARej += st.aRej
}

// rasterWorkers decides how many workers may fill this draw. One means "here, on this
// goroutine".
//
// The first group is not about profit, it is about FAITHFULNESS: each of these is something
// that observes the fill as it happens, or that the fill can do to the machine, and a
// partition would either race on it or reorder what it reports.
//
//   - OnPixel is the frame debugger's per-pixel hook (the capture that answers "which draw
//     made this pixel"). It fires per fragment, in raster order, and a capture must be exactly
//     reproducible.
//   - OnRead/OnWrite are the watch windows. Strictly the fill does not go through them —
//     texByte indexes RAM directly and never calls readWatch — but a watch session should be
//     exactly reproducible, and this is one branch per draw to buy that.
//   - drawTrace and pixDbg print, per draw and per pixel, and their order is their content.
//   - canHalt is the one that would be a real bug rather than a muddle: decodeTexel and
//     tlutColor call CPU.Halt on a format they do not know, and A WORKER MUST NOT HALT THE
//     MACHINE. Rather than teach the fragment path not to halt, a draw that COULD halt is run
//     serially, so the halt happens on the machine's own goroutine at exactly the fragment it
//     always did.
//   - SingleThreaded is the caller saying so.
//
// Only then the work threshold, which is about profit.
func (g *gpu) rasterWorkers(m *Machine, tev *tevState, tris []rasterTri) int {
	if m.OnPixel != nil || m.OnRead != nil || m.OnWrite != nil || m.SingleThreaded ||
		drawTrace || pixDbgX >= 0 || cullExperiment != "" || tev.canHalt {
		return 1
	}
	// The bounding boxes are what the fill will actually walk. A draw of a few small
	// triangles costs less than the fan-out that would split it.
	//
	// A box can be EMPTY IN THE NEGATIVE: a triangle entirely off one edge has its near
	// bound clamped to the screen and its far bound left outside, so maxX < minX. The fill
	// loops zero times for it, which is right, but a naive sum would let such a triangle
	// subtract from the tally and talk the threshold out of splitting a draw that needed it.
	const minPixels = 256
	px := 0
	for i := range tris {
		t := &tris[i]
		w, h := t.maxX-t.minX, t.maxY-t.minY
		if w <= 0 || h <= 0 {
			continue
		}
		px += w * h
		if px >= minPixels {
			return maxWorkers
		}
	}
	return 1
}

// The cull modes, as they appear in GEN_MODE's two-bit field (bits 14..15).
//
// The field's position and the fact that it is the cull mode are pinned from the game's own
// GXSetCullMode at 0x801F51C0, which is almost entirely a remap: it takes the caller's mode,
// SWAPS 1 and 2 (leaving 0 and 3 alone), and inserts the result at bits 14..15 of the
// GEN_MODE shadow —
//
//	cmpwi r3,2 ; beq  -> li r3,1
//	cmpwi r3,1 ; bge  -> li r3,2
//	              else -> unchanged
//	slwi r0,r3,14 ; rlwinm r3,r3,0,18,15 ; or ; stw
//
// — so the SDK's enum and the hardware field disagree about which of the two culling modes is
// which, and this file wants the hardware's. Corroborated by watching the cutscene program
// the field: only bits 14..15 ever vary across a whole frame, taking exactly the values 0, 1
// and 2 (8, 22 and 66 draws respectively) and never 3.
//
// THEY ARE NAMED FOR THE SIGN THEY DISCARD, NOT "FRONT"/"BACK". Which screen winding is a
// front face is not something the register, the swap, or the disassembly says — it depends on
// this rasteriser's own projection, and the only thing that actually knows is the rendered
// picture (see cullTest). Naming them front/back would be asserting a mapping nothing here
// has verified, and getting it backwards is silent: the scene renders, and it renders
// inside-out. The circumstantial case, recorded but deliberately NOT relied on: mode 2 is the
// overwhelmingly common one (66 draws of 96), and a 3D scene culls back faces far more often
// than front ones, so cullPosArea is very probably "cull back".
const (
	cullNone    = 0 // draw every triangle
	cullNegArea = 1 // discard triangles whose screen-space signed area is negative
	cullPosArea = 2 // discard triangles whose screen-space signed area is positive
	cullAll     = 3 // discard every triangle
)

// cullTest reports whether a triangle of this screen-space winding should be discarded.
//
// WHICH SIGN GOES WITH WHICH MODE IS THE HALF THE DISASSEMBLY CANNOT ANSWER, because it
// depends on this rasteriser's own projection — the viewport transform flips Y, which flips
// the sign of every triangle's area. So it was settled by rendering the frame both ways and
// looking, and the first guess was wrong: the other assignment renders the forest inside-out
// (92.7% of the frame's pixels change; you see through the trunks and into the back of
// Luigi's head). This way round 9.4% change, and those changes are the flashlight's cone —
// whose back faces were being blended in a second time — and thin slivers along the tree
// silhouettes. Culling fixes those rather than breaking them.
//
// THE MEASUREMENTS COULD NOT HAVE DECIDED IT. Both assignments cull ~8,000 of the field's
// triangles and both make the rasteriser ~30% faster, because in closed geometry half the
// triangles face each way — culling the wrong half is exactly as cheap as culling the right
// one, and the profiler reports a triumph either way. Only the picture knows, which is why
// this is pinned by TestCullingDoesNotChangeTheOpaqueScene against a rendered frame and not
// by a speedup.
func (g *gpu) cullTest(area float32) bool {
	if cullExperiment != "" {
		return g.cullTestExperiment(area)
	}
	switch (g.BP[0x00] >> 14) & 3 {
	case cullAll:
		return true
	case cullPosArea:
		return area > 0
	case cullNegArea:
		return area < 0
	}
	return false
}

// cullExperiment is the scaffolding that settled the sign convention above, kept because it
// is also the fastest way to answer "is this geometry missing because of culling?" the next
// time something disappears from a scene:
//
//	RR_GC_CULLMODE=off    draw every triangle, whatever the register says
//	RR_GC_CULLMODE=flip   take the opposite sign — i.e. the convention that is wrong here
//
// Read once at init, so it must be set in the environment of the process (setting it from
// inside main is too late — that mistake made the first run of the experiment report three
// identical results and briefly look like culling did nothing at all).
var cullExperiment = os.Getenv("RR_GC_CULLMODE")

func (g *gpu) cullTestExperiment(area float32) bool {
	switch cullExperiment {
	case "off":
		return false
	case "flip":
		switch (g.BP[0x00] >> 14) & 3 {
		case cullAll:
			return true
		case cullPosArea:
			return area < 0
		case cullNegArea:
			return area > 0
		}
	}
	return false
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
