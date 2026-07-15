package gc

import (
	"math"
	"testing"

	"retroreverse.com/tools/cpu/gekko"
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

// encodeRGB565Tiled is an independent encoder of the RGB565 tiled layout — written from the
// format description, not by calling the decoder — so that reading a texture it produced back
// through the TX unit is a real cross-check of the decode and not a tautology. It tiles the
// image into 4x4 blocks, each block's texels in row-major order, and packs each texel as a
// big-endian 5-6-5 halfword.
func encodeRGB565Tiled(w, h int, at func(x, y int) (r, g, b uint8)) []byte {
	const bw, bh = 4, 4
	tilesX := (w + bw - 1) / bw
	tilesY := (h + bh - 1) / bh
	out := make([]byte, tilesX*tilesY*bw*bh*2)
	o := 0
	for ty := 0; ty < tilesY; ty++ {
		for tx := 0; tx < tilesX; tx++ {
			for iy := 0; iy < bh; iy++ {
				for ix := 0; ix < bw; ix++ {
					r, g, b := at(tx*bw+ix, ty*bh+iy)
					v := uint16(r>>3)<<11 | uint16(g>>2)<<5 | uint16(b>>3)
					out[o] = byte(v >> 8)
					out[o+1] = byte(v)
					o += 2
				}
			}
		}
	}
	return out
}

// testMachine is a bare machine with just enough memory for the graphics unit to read textures
// out of — no disc, no CPU stepping.
func testMachine() *Machine {
	m := &Machine{RAM: make([]byte, 2*1024*1024), logSeen: map[string]bool{}}
	m.CPU = gekko.NewCPU(m)
	return m
}

// bindTexture programs texture map 0's registers for a w x h RGB565 texture at physical base.
func (g *gpu) bindTexture0RGB565(base uint32, w, h int) {
	g.BP[0x80] = 0x000000                              // clamp/clamp wrap
	g.BP[0x88] = uint32(w-1) | uint32(h-1)<<10 | 4<<20 // width-1, height-1, format RGB565
	g.BP[0x94] = base >> 5
}

// TestTextureSampleRGB565 is the stage-4 gate: a texture is built with an independent encoder,
// bound, and read back through the TX unit texel for texel. A wrong tile size, a swapped byte
// order, or a mis-packed channel would show as a mismatch here.
func TestTextureSampleRGB565(t *testing.T) {
	const w, h = 64, 48
	base := uint32(0x1000)
	// A gradient whose channels differ so a swapped channel would not hide.
	src := func(x, y int) (r, g, b uint8) {
		return uint8(x * 4), uint8(y * 5), uint8((x + y) * 2)
	}
	data := encodeRGB565Tiled(w, h, src)

	m := testMachine()
	copy(m.RAM[base:], data)
	g := &gpu{}
	g.bindTexture0RGB565(base, w, h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Sample at the texel centre; the normalised coordinate the hardware would use.
			s := (float32(x) + 0.5) / float32(w)
			tt := (float32(y) + 0.5) / float32(h)
			gotR, gotG, gotB, _ := g.sampleTexmap(m, 0, s, tt)
			// The expected value is the 5-6-5 truncation of the source, since that is all the
			// format can store — compare against the encoder's own quantisation.
			wr, wg, wb := src(x, y)
			er, eg, eb, _ := decodeRGB565(uint16(wr>>3)<<11 | uint16(wg>>2)<<5 | uint16(wb>>3))
			if gotR != er || gotG != eg || gotB != eb {
				t.Fatalf("texel (%d,%d) = (%d,%d,%d), want (%d,%d,%d)", x, y, gotR, gotG, gotB, er, eg, eb)
			}
		}
	}
}

// TestDumpTextureMatchesSampler checks that the whole-texture dump agrees with per-texel
// sampling — the dump is the artefact the oracle writes to show the loading image, so it must
// decode identically to what the TEV samples.
func TestDumpTextureMatchesSampler(t *testing.T) {
	const w, h = 32, 32
	base := uint32(0x2000)
	src := func(x, y int) (r, g, b uint8) { return uint8(x * 8), uint8(y * 8), 128 }
	m := testMachine()
	copy(m.RAM[base:], encodeRGB565Tiled(w, h, src))
	m.gpu.bindTexture0RGB565(base, w, h)

	img, err := m.DumpTexture(0)
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s := (float32(x) + 0.5) / float32(w)
			tt := (float32(y) + 0.5) / float32(h)
			sr, sg, sb, _ := m.gpu.sampleTexmap(m, 0, s, tt)
			o := img.PixOffset(x, y)
			if img.Pix[o] != sr || img.Pix[o+1] != sg || img.Pix[o+2] != sb {
				t.Fatalf("dump (%d,%d) disagrees with sampler", x, y)
			}
		}
	}
}

// setDrawQuadTEV programs a minimal one-stage TEV that outputs the texture colour directly
// (stage 0 colour = TEXC, alpha = 1), the alpha test to always pass, and blending off, so a
// rendered quad reproduces the bound texture and the whole raster+TEV+texture path can be
// asserted end to end.
func (g *gpu) setPassthroughTEV() {
	g.BP[0x00] = 0x0000              // one colour channel, one texgen, one TEV stage
	g.BP[0x28] = 0x0000C0 | (1 << 6) // stage 0: texmap 0, texcoord 0, texture enabled
	// colour: out = TEXC. a=TEXC(8), b=ZERO(15), c=ZERO(15), d=ZERO(15), clamp.
	g.BP[0xC0] = 8<<12 | 15<<8 | 15<<4 | 15 | (1 << 19)
	// alpha: out = ONE via KONST? simplest: d = KONST(6) with konst sel = 1.0. Use a=ZERO..d;
	// instead set d = A2 and load reg2 alpha = 255.
	g.BP[0xC1] = 3<<4 | 7<<7 | 7<<10 | 7<<13 | (1 << 19) // d=A2, others ZERO
	g.TevColorReg[3] = [2]uint32{0x0FF0FF, 0x0FF0FF}     // reg2/C2 alpha = 255
	g.BP[0xF3] = 7<<16 | 7<<19                           // alpha compare: ALWAYS AND ALWAYS
	g.BP[0x41] = 0x000018                                // blend off, colour+alpha update on
}

// TestTexturedQuadRenders draws the full-screen quad over a bound texture with a passthrough
// TEV and asserts the framebuffer shows the texture — the end-to-end proof that vertex fetch,
// transform, texcoord interpolation, texture sampling and the combiner agree.
func TestTexturedQuadRenders(t *testing.T) {
	const w, h = 64, 64
	base := uint32(0x3000)
	src := func(x, y int) (r, g, b uint8) { return uint8(x * 4), uint8(y * 4), 200 }
	m := testMachine()
	copy(m.RAM[base:], encodeRGB565Tiled(w, h, src))

	g := quadGPU()
	m.gpu = *g
	m.gpu.bindTexture0RGB565(base, w, h)
	m.gpu.setPassthroughTEV()

	// One full-screen quad, texcoords spanning the whole texture.
	corners := [4]struct {
		x, y int16
		u, v uint16
	}{
		{0, 0, 0x0000, 0x0000},
		{640, 0, 0x8000, 0x0000},
		{640, 480, 0x8000, 0x8000},
		{0, 480, 0x0000, 0x8000},
	}
	var data []byte
	for _, c := range corners {
		data = append(data,
			byte(uint16(c.x)>>8), byte(c.x),
			byte(uint16(c.y)>>8), byte(c.y),
			0, 0,
			0xFF, 0xFF, 0xFF, 0xFF,
			byte(c.u>>8), byte(c.u), byte(c.v>>8), byte(c.v),
		)
	}
	m.gpu.drawPrimitive(m, 0x80, 0, 14, data)

	if m.gpu.EFB == nil {
		t.Fatal("raster produced no framebuffer")
	}
	// A handful of pixels across the screen should carry the texture colour at that position.
	for _, p := range [][2]int{{16, 16}, {320, 240}, {600, 100}, {100, 450}} {
		px := m.gpu.EFB[p[1]*efbWidth+p[0]]
		r, gg, b := unpackRGB(px)
		// The texel the quad maps this screen pixel to.
		tx := p[0] * w / 640
		ty := p[1] * h / 480
		sr, sg, sb := src(tx, ty)
		er, eg, eb, _ := decodeRGB565(uint16(sr>>3)<<11 | uint16(sg>>2)<<5 | uint16(sb>>3))
		if absU8(r, er) > 8 || absU8(gg, eg) > 8 || absU8(b, eb) > 8 {
			t.Errorf("pixel (%d,%d) = (%d,%d,%d), want ~(%d,%d,%d)", p[0], p[1], r, gg, b, er, eg, eb)
		}
	}
}

// TestTEVReproducesBootLogo pins the combiner against the exact registers the game's boot quad
// programs (captured from the FIFO): two stages that put the texture into PREV and then use it
// as a coverage mask, interpolating between a black background constant (C0) and a red logo
// constant (C1). Where the logo texture is white the pixel must come out red, where it is black
// the pixel must come out black — the red logo on black the console shows, and the assertion
// that would have caught the red/alpha decode swap that once flattened it.
func TestTEVReproducesBootLogo(t *testing.T) {
	g := &gpu{}
	g.BP[0x00] = 0x000411                            // two TEV stages
	g.BP[0x28] = 0x3803C0                            // stage 0 texmap0/tc0 enabled, stage 1 texture disabled
	g.BP[0xC0] = 0x088FFF                            // stage 0 colour = TEXC
	g.BP[0xC1] = 0x087FF0                            // stage 0 alpha  = A2
	g.BP[0xC2] = 0x08240F                            // stage 1 colour = lerp(C0, C1, CPREV)
	g.BP[0xC3] = 0x082870                            // stage 1 alpha  = lerp(A0, A1, APREV)
	g.BP[0xF3] = 0x240000                            // alpha test: alpha > 0
	g.TevColorReg[1] = [2]uint32{0x0FF000, 0x000000} // C0 = (r=0, a=255) — black background
	g.TevColorReg[2] = [2]uint32{0x0FF0DC, 0x000000} // C1 = (r=220, a=255) — red logo
	g.TevColorReg[3] = [2]uint32{0x0FF0FF, 0x0FF0FF} // C2 = white (its alpha 255 is stage 0's A2)

	m := testMachine()
	// A logo-shaped mask: the left half white (logo), the right half black (background).
	base := uint32(0x4000)
	copy(m.RAM[base:], encodeRGB565Tiled(8, 8, func(x, y int) (uint8, uint8, uint8) {
		if x < 4 {
			return 255, 255, 255
		}
		return 0, 0, 0
	}))
	g.bindTexture0RGB565(base, 8, 8)
	m.gpu = *g

	// A coordinate over the white half must shade red; one over the black half must shade black.
	r, gg, b, _, pass := m.gpu.shade(m, 255, 255, 255, 255, 0.1, 0.5)
	if !pass || absU8(r, 220) > 2 || gg != 0 || b != 0 {
		t.Errorf("logo texel: colour (%d,%d,%d) pass=%v, want ~(220,0,0)", r, gg, b, pass)
	}
	r, gg, b, _, pass = m.gpu.shade(m, 255, 255, 255, 255, 0.9, 0.5)
	if !pass || r != 0 || gg != 0 || b != 0 {
		t.Errorf("background texel: colour (%d,%d,%d) pass=%v, want (0,0,0)", r, gg, b, pass)
	}
}

func absU8(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

// TestCopyTextureRoundTrip is the stage-5 gate for the pixel-engine copy-to-texture: the EFB is
// filled with a gradient whose four channels all differ, copied out in each destination format
// the encoder carries, then read back through the TX unit's decoders — which were themselves
// validated against an independent encoder in the stage-4 gate — and compared against the
// format's quantisation computed here from the channel values alone. A wrong tile geometry, a
// swapped channel, or a misplaced stride would each break a different case.
func TestCopyTextureRoundTrip(t *testing.T) {
	const w, h = 32, 24
	base := uint32(0x8000)
	// Source pattern: all channels differ, and alpha sweeps 0..255 so both RGB5A3 arms run.
	src := func(x, y int) (r, g, b, a uint8) {
		return uint8(x * 8), uint8(y * 10), uint8((x + y) * 4), uint8(x * y * 255 / (w * h))
	}

	cases := []struct {
		name    string
		copyFmt int // the GXTexFmt low nibble as GXSetTexCopyDst encodes it
		texFmt  int // the TX format the readback binds (channel copies read back as I8)
		stride  int // tiles-per-row x bytes-per-tile for a packed w-wide destination
		want    func(r, g, b, a uint8) (uint8, uint8, uint8, uint8)
	}{
		{"R4", 0x0, texI4, (w / 8) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			i := expand4(uint16(r >> 4))
			return i, i, i, i
		}},
		{"RA4", 0x2, texIA4, (w / 8) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			i := expand4(uint16(r >> 4))
			return i, i, i, expand4(uint16(a >> 4))
		}},
		{"RA8", 0x3, texIA8, (w / 4) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			return r, r, r, a
		}},
		{"RGB565", 0x4, texRGB565, (w / 4) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			return expand5(uint16(r >> 3)), expand6(uint16(g >> 2)), expand5(uint16(b >> 3)), 255
		}},
		{"RGB5A3", 0x5, texRGB5A3, (w / 4) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			if a >= 0xE0 {
				return expand5(uint16(r >> 3)), expand5(uint16(g >> 3)), expand5(uint16(b >> 3)), 255
			}
			return expand4(uint16(r >> 4)), expand4(uint16(g >> 4)), expand4(uint16(b >> 4)), expand3(uint16(a >> 5))
		}},
		{"RGBA8", 0x6, texRGBA8, (w / 4) * 64, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			return r, g, b, a
		}},
		{"A8", 0x7, texI8, (w / 8) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			return a, a, a, a
		}},
		{"R8", 0x8, texI8, (w / 8) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			return r, r, r, r
		}},
		{"G8", 0x9, texI8, (w / 8) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			return g, g, g, g
		}},
		{"B8", 0xA, texI8, (w / 8) * 32, func(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
			return b, b, b, b
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := testMachine()
			g := &m.gpu
			g.ensureEFB()
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					r, gg, b, a := src(x, y)
					g.EFB[y*efbWidth+x] = packRGBA(r, gg, b, a)
				}
			}
			// Program the copy: source rect at (0,0), dest base and stride, then trigger with
			// the control word GXSetTexCopyDst/GXCopyTex compose (bit 16 always set, format
			// bit 3 at bit 3, format low bits at 4..6, bit 14 clear = to texture).
			g.BP[0x49] = 0
			g.BP[0x4A] = uint32(w-1) | uint32(h-1)<<10
			g.BP[0x4B] = base >> 5
			g.BP[0x4D] = uint32(c.stride) >> 5
			params := uint32(c.copyFmt&7)<<4 | uint32(c.copyFmt>>3)<<3 | 1<<16 | 3
			g.copyDisplay(m, params)
			if m.CPU.Halted {
				t.Fatalf("copy halted: format 0x%X", c.copyFmt)
			}

			tx := texState{format: c.texFmt, width: w, height: h, base: base}
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					r, gg, b, a := src(x, y)
					er, eg, eb, ea := c.want(r, gg, b, a)
					gr, ggg, gb, ga := g.decodeTexel(m, tx, x, y)
					if gr != er || ggg != eg || gb != eb || ga != ea {
						t.Fatalf("texel (%d,%d) = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
							x, y, gr, ggg, gb, ga, er, eg, eb, ea)
					}
				}
			}
		})
	}
}
