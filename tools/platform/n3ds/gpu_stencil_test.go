package n3ds

import "testing"

func TestStencilOperations(t *testing.T) {
	for _, tc := range []struct {
		op             uint32
		old, ref, want uint8
	}{
		{0, 9, 3, 9}, {1, 9, 3, 0}, {2, 9, 3, 3}, {3, 254, 0, 255}, {3, 255, 0, 255},
		{4, 1, 0, 0}, {4, 0, 0, 0}, {5, 0x35, 0, 0xca}, {6, 255, 0, 0}, {7, 0, 0, 255},
	} {
		if got := stencilValue(tc.op, tc.old, tc.ref); got != tc.want {
			t.Errorf("op %d: %d want %d", tc.op, got, tc.want)
		}
	}
}

func TestStencilMaskedCompareAndDepthBranches(t *testing.T) {
	g := testGPU()
	g.Regs[regDepthbufWrite] = 1
	fb := fbState{depthBuf: []byte{0x12, 0x34, 0x56, 0xa5}, depthTest: true, depthFunc: 0}
	// Equal after input masking; depth fails, so replace only the low nibble.
	g.Regs[0x105] = 0x0f050f21
	g.Regs[0x106] = 2 << 4
	if got := g.stencilDepthTest(&fb, 0, .5); got != 2 || fb.depthBuf[3] != 0xa5 {
		t.Fatalf("depth fail %d: %x", got, fb.depthBuf)
	}
	g.Regs[0x105] = 0x0f030f21
	g.Regs[0x106] = 2
	if got := g.stencilDepthTest(&fb, 0, .5); got != 1 || fb.depthBuf[3] != 0xa3 {
		t.Fatalf("stencil fail %d: %x", got, fb.depthBuf)
	}
	fb.depthTest = false
	g.Regs[0x105] = 0xff07ff11
	g.Regs[0x106] = 2 << 8
	if got := g.stencilDepthTest(&fb, 0, .5); got != 0 || fb.depthBuf[3] != 7 {
		t.Fatalf("depth disabled %d: %x", got, fb.depthBuf)
	}
	g.Regs[regDepthbufWrite] = 0
	g.Regs[0x106] = 1 << 8
	g.stencilDepthTest(&fb, 0, .5)
	if fb.depthBuf[3] != 7 || fb.depthBuf[0] != 0x12 || fb.depthBuf[1] != 0x34 || fb.depthBuf[2] != 0x56 {
		t.Fatalf("write gate/depth bytes: %x", fb.depthBuf)
	}
}

func TestStencilTitleClearPreservesColor(t *testing.T) {
	for _, watched := range []bool{false, true} {
		g := testGPU()
		m := g.m
		m.mapRegion("vram", 0x1f000000, make([]byte, 512))
		if watched {
			m.OnWrite = func(uint32, uint32, uint32) {}
		}
		g.Regs[0x116] = 3
		g.Regs[regDepthbufWrite] = 1
		g.Regs[0x102] = 3
		// Mario's stencil-only fullscreen clear: NEVER, replace with zero on fail.
		g.Regs[0x105] = 0xff00ff01
		g.Regs[0x106] = 0x222
		fb := fbState{colorAddr: 0x1f000000, depthAddr: 0x1f000100, width: 8, height: 8, colorMask: 15, depthZBuffer: true}
		fb.colorBuf, fb.colorOff = m.directRange(fb.colorAddr, 256)
		fb.depthBuf, fb.depthOff32 = m.directRange(fb.depthAddr, 256)
		off := tiledOffset(0, 7, 8)
		m.WriteWord(fb.colorAddr+off, 0x123456ff)
		m.WriteWord(fb.depthAddr+off, 0xa5123456)
		v0 := scrVert{x: 0, y: 0, iw: 1, col: [4]float32{1, 1, 1, 1}}
		v1, v2 := v0, v0
		v1.x = 2
		v2.y = 2
		tri := rasterTri{v0: v0, v1: v1, v2: v2, area: 4, minX: 0, maxX: 1, minY: 0, maxY: 1}
		ls := lightState{}
		tv := g.tevstate()
		st := rstats{}
		var ev PixelEvent
		m.OnPixel = func(x, y uint32, e PixelEvent) { ev = e }
		g.fillTri(&fb, &ls, &tv, &tri, 0, 1, &st)
		if m.ReadWord(fb.colorAddr+off) != 0x123456ff || m.ReadWord(fb.depthAddr+off) != 0x00123456 || !ev.StencilReject {
			t.Fatalf("watched=%v: color=%x depth=%x event=%+v", watched, m.ReadWord(fb.colorAddr+off), m.ReadWord(fb.depthAddr+off), ev)
		}
		// Alpha rejection must suppress even the stencil-fail operation.
		m.WriteWord(fb.depthAddr+off, 0xa5123456)
		tv.alphaTest = true
		tv.alphaFunc = 0
		g.fillTri(&fb, &ls, &tv, &tri, 0, 1, &st)
		if m.ReadWord(fb.depthAddr+off) != 0xa5123456 || !ev.AlphaReject {
			t.Fatal("alpha rejection changed stencil")
		}
	}
}
