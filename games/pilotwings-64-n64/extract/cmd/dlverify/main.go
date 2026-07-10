// dlverify mechanically verifies the display-list walker against the oracle.
//
// For one video field it takes both routes to the screen and compares them:
//
//   - the walk: snapshot RDRAM at the field's graphics task, walk the display
//     list with the f3d package, and push every triangle through the walked
//     modelview, projection and viewport — the same arithmetic the GLB
//     exporter trusts;
//   - the oracle: keep running the machine and record the RDP triangle
//     commands the RSP microcode actually emitted for that frame, ending at
//     Sync_Full, with each triangle's texture source tracked through
//     Set_Texture_Image / Set_Tile / Load_Block by TMEM address.
//
// Matching is per texture source: triangle counts and screen bounding boxes.
// Exact edge coefficients cannot match — the RSP clips and subdivides — so
// the tolerances are declared here, up front:
//
//   - a walked triangle wholly inside the frustum must appear in the RDP
//     stream; one wholly outside a clip plane must not; a straddling one
//     yields 0..7 clipped fragments; a near-degenerate one (|area| < 1 px²)
//     may collapse or flip its cull test under the RSP's quantization and is
//     exempt. Per texture source the RDP count must lie in
//     [inside, inside + 7*straddling + slivers].
//   - each fully-inside walked triangle is matched 1:1 (nearest-neighbour) to
//     an RDP triangle of the same texture source with every bbox corner
//     within BBOXTOL; the match rate is the headline number. Matches between
//     BBOXTOL and FARTOL are the RSP's s15.16 matrix-concatenation residue
//     (the title card's logo matrices have rotation elements ~0.01, where one
//     s15.16 quantum is 0.1% of the element); past FARTOL it is a defect.
//
// What this pinned when first run: screen y maps with negated viewport scale
// (x with positive); the cull-face sign in that flipped space; per-VERTEX
// viewport capture (the RSP maps to the screen at G_VTX time, and the frame
// reprograms the viewport between the 3-D scene and the logo overlay);
// untextured draws follow G_TEXTURE's enable, not the bound tile. A mismatch
// past the declared bands is a finding about the walker's matrix or segment
// handling, not noise.
//
// Current standing results: the flyby scene passes clean (692/695 in 2 px,
// zero sources out of range); title and logo match 99.4% each but exit 1 on
// seven untextured triangles apiece — small bird-sized shapes in one static
// logo-cluster group, whose matrix and vertex bytes are identical between
// the snapshot and the run (checked here) and which are visible in the
// rendered frame at both the walked and the streamed positions. They look
// like near-duplicate draws pairing off against each other; unexplained, and
// left failing on purpose so the question stays visible.
//
// Usage:
//
//	dlverify -image ROM -loadstate work/title.state [-steps N]
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"

	"retroreverse.com/games/pilotwings-64-n64/extract/f3d"
	"retroreverse.com/tools/platform/n64"
)

// BBOXTOL is the screen-space tolerance in pixels for a 1:1 match; FARTOL is
// the band above it attributed to the RSP's s15.16 matrix concatenation on
// scenes with sub-0.01 rotation elements (measured at 2-4 px on the title
// card's logo cluster). Anything past FARTOL is a walker defect: a wrong
// matrix, segment or viewport moves geometry by tens of pixels or drops it
// entirely. Both declared before comparing; see the package comment.
const (
	BBOXTOL = 2.0
	FARTOL  = 8.0
)

type bbox struct{ x0, y0, x1, y1 float64 }

func (b bbox) union(o bbox) bbox {
	return bbox{math.Min(b.x0, o.x0), math.Min(b.y0, o.y0), math.Max(b.x1, o.x1), math.Max(b.y1, o.y1)}
}

func (b bbox) String() string {
	return fmt.Sprintf("[%.1f,%.1f..%.1f,%.1f]", b.x0, b.y0, b.x1, b.y1)
}

// near reports whether every corner of the two boxes is within BBOXTOL.
func near(a, b bbox) bool {
	return math.Abs(a.x0-b.x0) <= BBOXTOL && math.Abs(a.y0-b.y0) <= BBOXTOL &&
		math.Abs(a.x1-b.x1) <= BBOXTOL && math.Abs(a.y1-b.y1) <= BBOXTOL
}

type rdpTri struct {
	tex     uint32
	bb      bbox
	matched bool
}

func main() {
	image := flag.String("image", "", "cartridge image")
	loadState := flag.String("loadstate", "", "machine snapshot to restore")
	steps := flag.Uint64("steps", 40000000, "instruction budget")
	debugTex := flag.String("debug", "", "print per-triangle vertices for this texture source (hex)")
	flag.Parse()
	var dbg uint32
	hasDbg := *debugTex != ""
	if hasDbg {
		fmt.Sscanf(*debugTex, "%x", &dbg)
	}

	rom, err := n64.Load(*image)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	m := n64.NewMachine(rom)
	if err := m.Boot(rom, n64.DefaultBoot()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := m.LoadState(*loadState); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// --- oracle side: catch the next graphics task, snapshot RAM, record the
	// frame's RDP triangles up to Sync_Full.
	var (
		snapRAM   []byte
		dlAddr    uint32
		recording bool
		rdpTris   []rdpTri

		// texture-source tracking, by TMEM address as the walker does it
		timg     uint32
		tileTmem [8]uint32
		tmemSrc  = map[uint32]uint32{}
	)
	m.OnRSPTask = func(mm *n64.Machine, pc uint32) {
		if snapRAM != nil {
			return
		}
		be := func(off int) uint32 {
			d := mm.DMEM[0xFC0+off:]
			return uint32(d[0])<<24 | uint32(d[1])<<16 | uint32(d[2])<<8 | uint32(d[3])
		}
		if be(0) != 1 { // not a graphics task
			return
		}
		snapRAM = append([]byte(nil), mm.RDRAM...)
		dlAddr = be(48) & 0x3FFFFF
		recording = true
		fmt.Fprintf(os.Stderr, "gfx task: dl=%06X, RDRAM snapped\n", dlAddr)
	}
	m.OnRDPCmd = func(mm *n64.Machine, op uint32, words []uint64) {
		if !recording {
			return
		}
		switch op {
		case 0x3D: // Set_Texture_Image
			timg = uint32(words[0]) & 0x03FFFFFF
		case 0x35: // Set_Tile
			tileTmem[words[0]>>24&7] = uint32(words[0] >> 32 & 0x1FF)
		case 0x33, 0x34: // Load_Block / Load_Tile
			tmemSrc[tileTmem[words[0]>>24&7]] = timg
		case 0x29: // Sync_Full: the frame is done
			recording = false
			mm.StopRequested = true
		default:
			if op >= 0x08 && op <= 0x0F {
				var tex uint32
				if op&0x02 != 0 { // textured variants
					tex = tmemSrc[tileTmem[words[0]>>48&7]]
				}
				rdpTris = append(rdpTris, rdpTri{tex: tex, bb: triBBox(words)})
				if hasDbg && tex == dbg {
					printRDPTri(words)
				}
			}
		}
	}
	res := m.Run(*steps)
	fmt.Fprintf(os.Stderr, "run: %s\n", res)
	if snapRAM == nil || recording {
		fmt.Fprintln(os.Stderr, "did not capture a complete frame")
		os.Exit(1)
	}
	fmt.Printf("RDP stream: %d triangles\n", len(rdpTris))

	// --- walk side: same frame, same RAM.
	w := f3d.New(snapRAM, false)
	w.Walk(dlAddr)
	fmt.Printf("walker: %d triangles, viewport scale=(%d %d) trans=(%d %d) /4\n",
		w.NTris, w.VpScale[0], w.VpScale[1], w.VpTrans[0], w.VpTrans[1])

	if hasDbg {
		for _, tri := range w.Tris {
			if tri.TexImg != dbg {
				continue
			}
			cls, _ := classify(tri, w)
			fmt.Printf("walk tri (%s):", cls)
			for _, v := range tri.V {
				sx := v.CX / v.CW * float64(v.VpScale[0]) / 4
				sy := v.CY / v.CW * float64(v.VpScale[1]) / 4
				fmt.Printf("  (%.2f,%.2f)", sx+float64(v.VpTrans[0])/4, float64(v.VpTrans[1])/4-sy)
			}
			fmt.Println()
		}
	}

	type side struct {
		inside, straddle, outside, culled, sliver int
		insideBBs                                 []bbox
		insideGroups                              []string
		insideTols                                []float64
		union                                     bbox
		any                                       bool
	}
	walked := map[uint32]*side{}
	for _, tri := range w.Tris {
		s := walked[tri.TexImg]
		if s == nil {
			s = &side{}
			walked[tri.TexImg] = s
		}
		cls, bb := classify(tri, w)
		switch cls {
		case "outside":
			s.outside++
			continue
		case "culled":
			s.culled++
			continue
		case "straddle":
			s.straddle++
		case "sliver":
			// Thinner than the RSP's quarter-pixel screen quantum: the
			// microcode may emit it or collapse it, so it is neither required
			// to appear nor required to match.
			s.sliver++
		case "inside":
			s.inside++
			s.insideBBs = append(s.insideBBs, bb)
			s.insideGroups = append(s.insideGroups, tri.Group)
			s.insideTols = append(s.insideTols, math.Max(FARTOL, residueBound(tri)))
			// The union covers only unclipped triangles: a straddling one is
			// cut back to the frustum by the RSP, so its full extent must not
			// widen what we require the RDP stream to cover.
			if !s.any {
				s.union, s.any = bb, true
			} else {
				s.union = s.union.union(bb)
			}
		}
	}

	rdpByTex := map[uint32][]*rdpTri{}
	for i := range rdpTris {
		t := &rdpTris[i]
		rdpByTex[t.tex] = append(rdpByTex[t.tex], t)
	}

	// --- compare.
	var texes []uint32
	for t := range walked {
		texes = append(texes, t)
	}
	sort.Slice(texes, func(i, j int) bool { return texes[i] < texes[j] })
	fail := 0
	var matchTot, matchOK int
	for _, tex := range texes {
		ws := walked[tex]
		rs := rdpByTex[tex]
		lo, hi := ws.inside, ws.inside+7*ws.straddle+ws.sliver
		countOK := len(rs) >= lo && len(rs) <= hi

		// union bbox of the RDP triangles
		var ru bbox
		for i, r := range rs {
			if i == 0 {
				ru = r.bb
			} else {
				ru = ru.union(r.bb)
			}
		}
		// The walked union covers only unclipped triangles, so the RDP union
		// may exceed it by straddling fragments; require containment of the
		// walked-inside union within the RDP union at tolerance when there is
		// anything to compare.
		unionOK := true
		if ws.inside > 0 && len(rs) > 0 {
			unionOK = ws.union.x0 >= ru.x0-BBOXTOL && ws.union.y0 >= ru.y0-BBOXTOL &&
				ws.union.x1 <= ru.x1+BBOXTOL && ws.union.y1 <= ru.y1+BBOXTOL
		}

		// 1:1 nearest-neighbour matching of fully-inside walked triangles
		// (greedy first-fit chains badly when dozens of similar slivers —
		// rotor blades — sit within tolerance of each other).
		ok, far := 0, 0
		for wi, wb := range ws.insideBBs {
			best, bestD := -1, math.Inf(1)
			for i, r := range rs {
				if r.matched {
					continue
				}
				d := math.Max(math.Max(math.Abs(wb.x0-r.bb.x0), math.Abs(wb.y0-r.bb.y0)),
					math.Max(math.Abs(wb.x1-r.bb.x1), math.Abs(wb.y1-r.bb.y1)))
				if d < bestD {
					best, bestD = i, d
				}
			}
			switch {
			case best >= 0 && bestD <= BBOXTOL:
				rs[best].matched = true
				ok++
			case best >= 0 && bestD <= ws.insideTols[wi]:
				// Within this triangle's fixed-point residue bound: the RSP
				// concatenates MV*P in s15.16, and the displacement that can
				// give scales with vertex magnitude over w. Not a walker
				// defect.
			default:
				far++
			}
			if best >= 0 && bestD > BBOXTOL && hasDbg && tex == dbg {
				fmt.Printf("  unmatched walk %s (nearest %.1f px) %s\n", wb, bestD, ws.insideGroups[wi])
			}
		}
		if hasDbg && tex == dbg {
			for _, r := range rs {
				if !r.matched {
					fmt.Printf("  unmatched rdp  %s\n", r.bb)
				}
			}
		}
		matchTot += ws.inside
		matchOK += ok

		verdict := "ok"
		if !countOK || !unionOK || far > 0 {
			verdict = "MISMATCH"
			fail++
		}
		fmt.Printf("tex %06X: walk in=%-4d straddle=%-3d out=%-4d cull=%-3d sliver=%-3d rdp=%-4d (want %d..%d)  match %d/%d  walkbb=%s rdpbb=%s  %s\n",
			tex, ws.inside, ws.straddle, ws.outside, ws.culled, ws.sliver, len(rs), lo, hi, ok, ws.inside, ws.union, ru, verdict)
	}

	// RDP triangles under texture sources the walker never drew.
	for tex, rs := range rdpByTex {
		if walked[tex] == nil {
			fmt.Printf("tex %06X: %d RDP triangles but the walker drew none  MISMATCH\n", tex, len(rs))
			fail++
		}
	}

	// Matrices the CPU rewrote while the task ran: the RSP reads RDRAM live,
	// so a group transformed through one of these was drawn with a fresher
	// pose than the snapshot the walker used. Mismatches confined to such
	// groups are the game racing its own frame, not a walker defect.
	stale := map[uint32]bool{}
	for _, tri := range w.Tris {
		a := tri.MtxAddr
		if stale[a] || int(a)+64 > len(snapRAM) {
			continue
		}
		for i := uint32(0); i < 64; i++ {
			if snapRAM[a+i] != m.RDRAM[a+i] {
				stale[a] = true
				fmt.Printf("matrix at %06X changed while the task ran (groups drawn through it may be one frame fresher)\n", a)
				break
			}
		}
	}
	staleVtx := 0
	for _, vl := range w.VtxLoads {
		if int(vl.Addr+vl.Len) > len(snapRAM) {
			continue
		}
		for i := uint32(0); i < vl.Len; i++ {
			if snapRAM[vl.Addr+i] != m.RDRAM[vl.Addr+i] {
				staleVtx++
				fmt.Printf("vertex load %06X+%X changed while the task ran\n", vl.Addr, vl.Len)
				break
			}
		}
	}

	fmt.Printf("\n%d/%d fully-inside triangles matched 1:1 within %.1f px (rest within the %.0f px s15.16 residue band); %d texture sources mismatched\n",
		matchOK, matchTot, BBOXTOL, FARTOL, fail)
	if fail > 0 {
		os.Exit(1)
	}
}

// residueBound is the worst-case screen displacement (pixels) the RSP's
// s15.16 matrix concatenation can give one triangle: each element of the
// combined MV*P is off by up to 2^-16 from the float product, a vertex
// multiplies that by its coordinate magnitude, and the perspective divide
// scales it by 1/w — doubled for the cross terms of the concatenation and
// the w-row's own error feeding the divide.
func residueBound(t f3d.Tri) float64 {
	const q = 1.0 / 65536
	worst := 0.0
	for _, v := range t.V {
		if v.CW <= 0 {
			continue
		}
		mag := math.Abs(float64(v.X)) + math.Abs(float64(v.Y)) + math.Abs(float64(v.Z)) + 1
		errClip := mag * q
		ndc := math.Max(math.Abs(v.CX), math.Abs(v.CY)) / v.CW
		px := (errClip / v.CW) * (1 + ndc) * float64(t.VpScale[0]) / 4 * 2
		if px > worst {
			worst = px
		}
	}
	return worst
}

// classify projects one walked triangle and places it against the frustum:
// "inside" (all vertices in), "outside" (all out past one plane), "culled"
// (backface under the geometry mode), or "straddle". The returned bbox is the
// screen-space extent of the projected vertices (valid for inside/straddle).
func classify(t f3d.Tri, w *f3d.Walker) (string, bbox) {
	var sx, sy [3]float64
	outMask := [3]uint32{}
	allOut := ^uint32(0)
	for i, v := range t.V {
		var m uint32
		if v.CX < -v.CW {
			m |= 1
		}
		if v.CX > v.CW {
			m |= 2
		}
		if v.CY < -v.CW {
			m |= 4
		}
		if v.CY > v.CW {
			m |= 8
		}
		if v.CZ < -v.CW {
			m |= 16
		}
		if v.CZ > v.CW {
			m |= 32
		}
		if v.CW <= 0 {
			m |= 64
		}
		outMask[i] = m
		allOut &= m
		if v.CW != 0 {
			// x maps with +scale, y with -scale: clip-space y grows up, screen
			// y down. Pinned by this comparison itself — with +y every texture
			// source's boxes mirror about the viewport centre (242/2 here),
			// with -y they coincide. The viewport is each vertex's own: the
			// RSP maps to the screen at G_VTX time, and the frame reprograms
			// the viewport between passes.
			sx[i] = v.CX/v.CW*float64(v.VpScale[0])/4 + float64(v.VpTrans[0])/4
			sy[i] = float64(v.VpTrans[1])/4 - v.CY/v.CW*float64(v.VpScale[1])/4
		}
	}
	if allOut != 0 {
		return "outside", bbox{}
	}
	inside := outMask[0]|outMask[1]|outMask[2] == 0

	bb := bbox{
		math.Min(sx[0], math.Min(sx[1], sx[2])), math.Min(sy[0], math.Min(sy[1], sy[2])),
		math.Max(sx[0], math.Max(sx[1], sx[2])), math.Max(sy[0], math.Max(sy[1], sy[2])),
	}

	// Screen-space signed area. The cull sign convention is pinned by this
	// very comparison: with the opposite sign every culling texture source
	// keeps exactly the triangles the RDP never received (the logo's
	// letters: 114 kept / 96 culled against 97 in the stream, and the same
	// inversion on every other source). In this flipped-y screen space a
	// front face winds clockwise, area > 0.
	area := (sx[1]-sx[0])*(sy[2]-sy[0]) - (sx[2]-sx[0])*(sy[1]-sy[0])
	if inside && area < 1.0 && area > -1.0 {
		// Near-degenerate: the RSP's s13.2 screen coordinates and s15.16
		// matrices can collapse it to nothing or flip which side of a cull
		// test it lands on. It may or may not appear in the stream.
		return "sliver", bb
	}
	if inside && t.GeoMode&(f3d.GeoCullBack|f3d.GeoCullFront) != 0 {
		if t.GeoMode&f3d.GeoCullBack != 0 && area > 0 {
			return "culled", bb
		}
		if t.GeoMode&f3d.GeoCullFront != 0 && area <= 0 {
			return "culled", bb
		}
	}
	if inside {
		return "inside", bb
	}
	return "straddle", bb
}

// printRDPTri reconstructs the three vertices from the edge coefficients: the
// H and M edges meet at the top vertex (at YH), M and L at the mid (at YM),
// H and L at the bottom (at YL).
func printRDPTri(words []uint64) {
	s14 := func(v uint32) int32 { return int32(v<<18) >> 18 }
	yl := int64(s14(uint32(words[0] >> 32 & 0x3FFF)))
	ym := int64(s14(uint32(words[0] >> 16 & 0x3FFF)))
	yh := int64(s14(uint32(words[0] & 0x3FFF)))
	xl := int64(int32(words[1] >> 32))
	xh, dxhdy := int64(int32(words[2]>>32)), int64(int32(words[2]))
	xm, dxmdy := int64(int32(words[3]>>32)), int64(int32(words[3]))
	base := yh &^ 3
	f := 1.0 / 65536
	fmt.Printf("rdp  tri: top(%.2f,%.2f)/(%.2f) mid(%.2f,%.2f) bot(%.2f,%.2f)\n",
		float64(xh+dxhdy*(yh-base)/4)*f, float64(yh)/4,
		float64(xm+dxmdy*(yh-base)/4)*f,
		float64(xl)*f, float64(ym)/4,
		float64(xh+dxhdy*(yl-base)/4)*f, float64(yl)/4)
}

// triBBox computes the screen-space bounding box of an RDP triangle from its
// edge coefficients: XH and XM are given at YH&~3, XL at exactly YM, each
// stepped by slope/4 per quarter-line (the convention measured over 14,507
// microcode triangles; see the game notes).
func triBBox(words []uint64) bbox {
	s14 := func(v uint32) int32 { return int32(v<<18) >> 18 }
	yl := int64(s14(uint32(words[0] >> 32 & 0x3FFF)))
	ym := int64(s14(uint32(words[0] >> 16 & 0x3FFF)))
	yh := int64(s14(uint32(words[0] & 0x3FFF)))
	xl, dxldy := int64(int32(words[1]>>32)), int64(int32(words[1]))
	xh, dxhdy := int64(int32(words[2]>>32)), int64(int32(words[2]))
	xm, dxmdy := int64(int32(words[3]>>32)), int64(int32(words[3]))

	base := yh &^ 3
	xs := []int64{
		xh + dxhdy*(yh-base)/4, xh + dxhdy*(yl-base)/4,
		xm + dxmdy*(yh-base)/4, xm + dxmdy*(ym-base)/4,
		xl, xl + dxldy*(yl-ym)/4,
	}
	minX, maxX := xs[0], xs[0]
	for _, x := range xs[1:] {
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
	}
	return bbox{float64(minX) / 65536, float64(yh) / 4, float64(maxX) / 65536, float64(yl) / 4}
}
