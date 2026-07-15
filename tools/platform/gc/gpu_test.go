package gc

import (
	"math"
	"testing"
)

// The transform and raster are validated against the vertex format and coordinate space the
// game's own first draw uses (captured from the FIFO): a full-screen orthographic quad whose
// four corners are (0,0), (640,0), (640,480), (0,480) in model space, drawn through an identity
// position matrix, an orthographic projection of 2/640 and -2/480, and a viewport that fills
// the 640x480 screen. Those corners must land on the framebuffer's corners; if the projection,
// the viewport, or the +342 bias were wrong, this is where it shows.

// quadGPU builds a gpu with the captured full-screen-quad XF and CP state.
func quadGPU() *gpu {
	g := &gpu{}
	// Vertex descriptor and table: position (direct, 3x s16), colour0 (direct, RGBA8888),
	// texcoord0 (direct, 2x u16).
	g.CPReg[0x50] = 0x2200
	g.CPReg[0x60] = 0x0001
	g.CPReg[0x70] = 0x5EA16007
	g.CPReg[0x30] = 0 // position matrix index 0

	// Identity position matrix at XF memory row 0.
	ident := [12]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0}
	for i, f := range ident {
		g.XFMem[i] = math.Float32bits(f)
	}
	// Viewport 0x101A..0x101F and orthographic projection 0x1020..0x1026.
	vp := [6]float32{320, -240, 16777215, 660, 580, 16777215}
	for i, f := range vp {
		g.XFMem[0x101A+i] = math.Float32bits(f)
	}
	proj := [6]float32{2.0 / 640, -1, -2.0 / 480, 1, -0.5, -0.5}
	for i, f := range proj {
		g.XFMem[0x1020+i] = math.Float32bits(f)
	}
	g.XFMem[0x1026] = 1 // orthographic
	return g
}

func TestTransformFullScreenQuad(t *testing.T) {
	g := quadGPU()
	cases := []struct {
		mx, my       float32
		wantX, wantY float32
	}{
		{0, 0, 0, 0},
		{640, 0, 640, 0},
		{640, 480, 640, 480},
		{0, 480, 0, 480},
	}
	for _, c := range cases {
		sx, sy, _ := g.transform(0, c.mx, c.my, 0)
		if math.Abs(float64(sx-c.wantX)) > 3 || math.Abs(float64(sy-c.wantY)) > 3 {
			t.Errorf("model (%g,%g) -> screen (%.1f,%.1f), want ~(%.0f,%.0f)",
				c.mx, c.my, sx, sy, c.wantX, c.wantY)
		}
	}
}

func TestRasterFillsScreen(t *testing.T) {
	g := quadGPU()
	// One full-screen quad: four vertices of pos s16, colour white, texcoord 0.
	corners := [4][2]int16{{0, 0}, {640, 0}, {640, 480}, {0, 480}}
	var data []byte
	for _, c := range corners {
		data = append(data,
			byte(uint16(c[0])>>8), byte(c[0]), // x
			byte(uint16(c[1])>>8), byte(c[1]), // y
			0, 0, // z
			0xFF, 0xFF, 0xFF, 0xFF, // colour RGBA
			0, 0, 0, 0, // texcoord
		)
	}
	g.drawPrimitive(&Machine{logSeen: map[string]bool{}}, 0x80, 0, 14, data)

	if g.EFB == nil {
		t.Fatal("raster produced no framebuffer")
	}
	// The centre and each mid-edge should be the white the quad carries.
	for _, p := range [][2]int{{320, 240}, {10, 240}, {630, 240}, {320, 10}, {320, 470}} {
		px := g.EFB[p[1]*efbWidth+p[0]]
		r, gg, b := unpackRGB(px)
		if r < 250 || gg < 250 || b < 250 {
			t.Errorf("pixel (%d,%d) = (%d,%d,%d), want white", p[0], p[1], r, gg, b)
		}
	}
}
