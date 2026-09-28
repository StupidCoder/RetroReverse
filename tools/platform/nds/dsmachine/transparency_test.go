package dsmachine

import "testing"

func TestTextureAlphaControlsBlendAndDepth(t *testing.T) {
	m := &Machine{vram: newVRAM()}
	m.vram.setCNT(0, 0x83) // texture data
	m.vram.setCNT(4, 0x83) // texture palette
	m.vram.bank[4][0], m.vram.bank[4][1] = 0xff, 0x7f
	g := newGPU3D()
	p := gxPolygon{attr: 31<<16 | 1<<24, texParam: 6 << 26}
	st := g.polyState(m, &p, 9)
	f := rvert{iw: 1, depth: 100, r: 63, g: 63, b: 63}
	for _, alpha := range []uint8{0, 15, 31} {
		g.rast.col[0], g.rast.depth[0], g.rast.transID[0] = rfrag{b: 63, a: 31}, 1000, 0xff
		m.vram.bank[0][0] = alpha << 3
		g.shade(m, &st, &p, f, 0)
		got := g.rast.col[0]
		switch alpha {
		case 0:
			if got != (rfrag{b: 63, a: 31}) || g.rast.depth[0] != 1000 {
				t.Fatal("Transparent texel wrote color or depth")
			}
		case 15:
			if got != (rfrag{r: 31, g: 31, b: 63, a: 31}) || g.rast.depth[0] != 1000 {
				t.Fatalf("Texture alpha must blend without writing depth: %+v depth=%d", got, g.rast.depth[0])
			}
			g.shade(m, &st, &p, f, 0)
			if g.rast.col[0] != got {
				t.Fatal("Same translucent polygon ID blended twice")
			}
			p.attr = 2<<24 | 31<<16 | 1<<11
			st = g.polyState(m, &p, 9)
			g.shade(m, &st, &p, f, 0)
			if g.rast.depth[0] != 100 {
				t.Fatal("Translucent depth-write flag ignored")
			}
			p.attr = 1<<24 | 31<<16
			st = g.polyState(m, &p, 9)
		case 31:
			if got != (rfrag{r: 63, g: 63, b: 63, a: 31}) || g.rast.depth[0] != 100 {
				t.Fatalf("Opaque texel must write color and depth: %+v", got)
			}
		}
	}
	// A transparent decal texel selects the vertex color, rather than disappearing.
	p.attr = 31<<16 | 1<<4
	st = g.polyState(m, &p, 9)
	m.vram.bank[0][0] = 0
	g.rast.depth[0] = 1000
	f.r, f.g, f.b = 63, 0, 0
	g.shade(m, &st, &p, f, 0)
	if g.rast.col[0] != (rfrag{r: 63, a: 31}) {
		t.Fatal("Zero-alpha decal lost vertex color")
	}
}

func TestPolygonOrdering(t *testing.T) {
	g := newGPU3D()
	makePoly := func(y int64, a, format uint32) gxPolygon {
		return gxPolygon{attr: a << 16, texParam: format << 26, verts: []gxVertex{{y: y, w: 4096}, {y: y + 512, w: 4096}, {y: y, w: 4096}}}
	}
	g.geom.polys = []gxPolygon{makePoly(-2048, 31, 6), makePoly(2048, 31, 0), makePoly(-2048, 31, 0), makePoly(2048, 15, 0)}
	check := func(want []int) {
		t.Helper()
		got := g.renderOrder()
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("Order=%v, want=%v", got, want)
			}
		}
	}
	check([]int{1, 2, 3, 0}) // opaque first, ascending bottom/top Y within each group
	g.geom.manualSort = true
	check([]int{1, 2, 0, 3}) // manual keeps translucent submission order
	g.geom.polys = []gxPolygon{makePoly(0, 31, 0), makePoly(0, 31, 0)}
	check([]int{0, 1})
}

func TestRenderingRegisterByteWrites(t *testing.T) {
	m := &Machine{gpu3d: newGPU3D()}
	c := &core{m: m, arm9: true, io: map[uint32]uint32{}}
	c.ioWrite(0x04000340, 17)
	c.ioWrite(regCLEARDEPTH, 0xff)
	c.ioWrite(regCLEARDEPTH+1, 0x7f)
	c.ioWrite(0x04000380, 0x1f)
	c.ioWrite(0x04000381, 0)
	if m.gpu3d.regs[0x04000340] != 17 || m.gpu3d.regs[regCLEARDEPTH] != 0x7fff || m.gpu3d.regs[0x04000380] != 31 {
		t.Fatal("Byte/halfword rendering register writes were dropped")
	}
}
