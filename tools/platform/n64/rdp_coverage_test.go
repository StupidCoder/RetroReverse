package n64

import "testing"

func TestCoverageAlphaControls(t *testing.T) {
	for _, tc := range []struct {
		mode                     uint64
		alpha, cvg, wantA, wantC uint32
	}{
		{0, 64, 8, 64, 8}, {omAlphaCvgSel, 0, 8, 255, 8}, {omAlphaCvgSel, 255, 4, 128, 4},
		{omCvgTimesAlpha, 128, 8, 128, 4}, {omCvgTimesAlpha | omAlphaCvgSel, 128, 4, 64, 2},
		{omCvgTimesAlpha | omAlphaCvgSel, 255, 8, 255, 8}, {omCvgTimesAlpha | omAlphaCvgSel, 0, 8, 0, 0},
	} {
		r := rdp{OtherModes: tc.mode}
		a, c := r.pixelCoverage(tc.alpha, tc.cvg)
		if a != tc.wantA || c != tc.wantC {
			t.Fatalf("mode=%x alpha=%d cvg=%d: got (%d,%d), want (%d,%d)", tc.mode, tc.alpha, tc.cvg, a, c, tc.wantA, tc.wantC)
		}
	}
}

// Exercise the complete pixel pipeline, with old-frame colours deliberately
// left in memory. Opaque AA must replace them; real transparency must retain them.
func TestCoverageBlendingPixelPipeline(t *testing.T) {
	for _, cycle := range []uint32{cycle1, cycle2} {
		for _, size := range []uint32{size16, size32} {
			m := newRDPTest(t, cycle)
			r := &m.rdp
			r.Color.Size = size
			// Pass primitive RGB and alpha through both combiner cycles.
			r.Combine = combineWord(ccZeroC, ccZeroC, ccZeroMul, ccPrim, ccZeroA, ccZeroA, ccZeroA, ccPrim)
			r.Combine |= uint64(ccZeroC)<<37 | uint64(ccZeroMul)<<32 | uint64(ccZeroC)<<24 | uint64(ccPrim)<<6 | uint64(ccZeroA)<<21 | uint64(ccZeroA)<<18 | uint64(ccZeroA)<<3 | ccPrim
			in := combineInputs{Prim: rgba{200, 40, 80, 0}, Shade: rgba{255, 255, 255, 255}}
			// Match the intro's final-cycle source-over mux. Also test the standard
			// opaque-AA memory-coverage mux, which must not average opaque interiors.
			for _, b := range []uint64{0, 1} {
				mux := uint64(1)<<22 | b<<18
				if cycle == cycle2 {
					mux = uint64(1)<<20 | b<<16
				}
				for _, old := range []rgba{{0, 0, 240, 255}, {240, 240, 0, 255}} {
					for _, alpha := range []uint32{0, 64, 128, 255} {
						r.OtherModes = uint64(cycle)<<52 | mux | omAntialias | omImageRead | omAlphaCvgSel
						in.Prim.A = alpha
						m.writePixel(2, 2, old.R, old.G, old.B, old.A)
						m.drawPixel(2, 2, &in, 0, false)
						got := m.readPixel(2, 2)
						if got.R != 200 || got.G != 40 || got.B != 80 {
							t.Fatalf("opaque cycle=%d size=%d alpha=%d b=%d: %+v", cycle, size, alpha, b, got)
						}
					}
				}
			}
			mux := uint64(1) << 22
			if cycle == cycle2 {
				mux = uint64(1) << 20
			}
			r.OtherModes = uint64(cycle)<<52 | mux | omImageRead | omForceBlend
			in.Prim.A = 128
			m.writePixel(2, 2, 0, 200, 0, 255)
			m.drawPixel(2, 2, &in, 0, false)
			got := m.readPixel(2, 2)
			if got.R < 96 || got.R > 104 || got.G < 112 || got.G > 120 {
				t.Fatalf("transparent cycle=%d size=%d: %+v", cycle, size, got)
			}
			// Alpha-derived coverage retains transparency for AA texture-edge modes.
			r.OtherModes = uint64(cycle)<<52 | mux | omImageRead | omAntialias | omAlphaCvgSel | omCvgTimesAlpha
			m.writePixel(2, 2, 0, 200, 0, 255)
			m.drawPixel(2, 2, &in, 0, false)
			got = m.readPixel(2, 2)
			if got.R < 96 || got.R > 104 || got.G < 112 || got.G > 120 {
				t.Fatalf("texture edge: %+v", got)
			}
			in.Prim.A = 0
			m.writePixel(2, 2, 0, 200, 0, 255)
			m.drawPixel(2, 2, &in, 0, false)
			if got = m.readPixel(2, 2); got.R != 0 || got.G != 200 {
				t.Fatalf("zero coverage overwrote background: %+v", got)
			}
			// Alpha testing sees coverage-selected alpha, not unused texture alpha.
			r.OtherModes = uint64(cycle)<<52 | omAlphaCvgSel | omAlphaCompare
			r.BlendColor = 200
			m.drawPixel(2, 2, &in, 0, false)
			if got = m.readPixel(2, 2); got.R != 200 {
				t.Fatalf("coverage alpha failed alpha test: %+v", got)
			}
		}
	}
}
