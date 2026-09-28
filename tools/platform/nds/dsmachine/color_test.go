package dsmachine

import "testing"

func Test3DColorsReach2DCompositor(t *testing.T) {
	m := &Machine{ARM9: &core{io: map[uint32]uint32{}}}
	r := raster{}
	r.reset()
	r.col[0] = rfrag{r: 63, a: 31}
	r.col[1] = rfrag{g: 63, a: 31}
	r.col[2] = rfrag{b: 63, a: 31}
	r.col[3] = rfrag{r: 63, g: 32, b: 16, a: 15}
	r.col[4] = rfrag{r: 63, g: 63, b: 63, a: 0}
	r.publish()
	e := newGPU2D().a
	e.m, e.threeD = m, r.frame
	e.threeDLine(0)
	for i, want := range []uint16{0x001f, 0x03e0, 0x7c00, rgb555(uint32(c6to8(63))>>3, uint32(c6to8(32))>>3, uint32(c6to8(16))>>3)} {
		if !e.bgOK[0][i] || e.bg[0][i] != want {
			t.Fatalf("pixel %d: got %#04x visible=%v, want %#04x", i, e.bg[0][i], e.bgOK[0][i], want)
		}
	}
	if e.a3D[0] != 31 || e.a3D[3] != uint8(uint32(a5to8(15))*31/255) || e.bgOK[0][4] {
		t.Fatal("alpha must come from alpha, independently of the red channel")
	}
}
