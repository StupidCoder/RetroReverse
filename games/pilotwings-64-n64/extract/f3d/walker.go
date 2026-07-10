// Package f3d walks SGI Fast3D (GBI v1) display lists out of an RDRAM
// snapshot: the matrix stack, vertex loads, triangles, and texture bindings.
//
// Pilotwings 64 runs SGI's Fast3D microcode ("RSP SW Version: 2.0D, 04-01-96"
// in its ucode data, GBI version 1 opcodes on the wire). The walker implements
// the command set the game actually uses and halts loudly on anything else, so
// an unhandled command is a finding rather than a silent gap.
package f3d

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// Walker carries the microcode state a display list mutates as it runs.
type Walker struct {
	RAM     []byte
	Verbose bool

	seg   [16]uint32
	depth int

	// current state
	mtxStack []Mtx44
	proj     Mtx44
	vtx      [64]Vertex // Fast3D's vertex buffer is 16 entries; keep headroom
	texImg   uint32
	texFmt   uint32
	texSiz   uint32
	texScale [2]uint32
	texOn    bool
	tlut     uint32 // RDRAM address of the palette Load_TLUT last read
	tile     [8]TileDesc
	texTile  uint32            // the tile G_TEXTURE selects for drawing
	tmemSrc  map[uint32]uint32 // TMEM word address -> RDRAM source of the last load into it
	geoMode  uint32            // G_SETGEOMETRYMODE bits: lighting decides colour-vs-normal
	otherL   uint32            // G_SETOTHERMODE_L accumulation: the render mode
	mtxAddr  uint32            // RDRAM address of the last modelview G_MTX applied

	// viewport, from G_MOVEMEM G_MV_VIEWPORT: 4 scales then 4 translates, s16
	// in 10.2 (so pixel = value/4).
	VpScale, VpTrans [4]int16

	// per-draw grouping: one group per (modelview matrix, texture image)
	Groups map[string]*Group
	Order  []string

	// Tris records every triangle in draw order with its clip-space vertices,
	// for verification against the RDP stream the oracle executed.
	Tris []Tri

	// VtxLoads records every G_VTX source range, for checking whether the CPU
	// rewrote vertex data while the RSP task ran.
	VtxLoads []VtxLoad

	NTris int
}

// TileDesc mirrors the Set_Tile / Set_Tile_Size state a draw samples through.
type TileDesc struct {
	Fmt, Size, Line, Tmem, Pal uint32
	CmS, MaskS, ShiftS         uint32
	CmT, MaskT, ShiftT         uint32
	SL, TL, SH, TH             uint32 // 10.2
	Img                        uint32 // RDRAM source of the texels
	Width                      uint32 // texture image width at load time
}

// Mtx44 is a row-vector matrix: screen = v * MV * P.
type Mtx44 [4][4]float64

// Vertex is a Fast3D vertex plus its transformed positions at load time:
// world space through the modelview, clip space through the projection. The
// viewport is captured per vertex because that is when the RSP maps to the
// screen — a triangle drawn across a viewport switch mixes two mappings.
type Vertex struct {
	X, Y, Z          int16
	Flag             uint16
	S, T             int16
	R, G, B, A       uint8
	WX, WY, WZ       float64
	CX, CY, CZ, CW   float64
	VpScale, VpTrans [4]int16
}

// Tri is one drawn triangle with the state that drew it. The viewport is
// captured per triangle: a frame reprograms it between passes (the title card
// draws the 3-D scene and the logo overlay through different viewports), so
// the walker-final values would misplace earlier passes.
type Tri struct {
	V                [3]Vertex
	TexImg           uint32 // RDRAM source the sampled tile was loaded from (0 = untextured)
	GeoMode          uint32
	VpScale, VpTrans [4]int16
	Group            string // name of the draw group the triangle landed in
	MtxAddr          uint32 // RDRAM address of the modelview matrix last applied
}

// VtxLoad is one G_VTX command's RDRAM source range.
type VtxLoad struct {
	Addr, Len uint32
}

// Group is a set of triangles sharing a modelview matrix and texture image.
type Group struct {
	Name   string
	TexImg uint32
	Tile   TileDesc
	TLUT   uint32
	Scale  [2]uint32
	Mtx    Mtx44
	Lit    bool   // G_LIGHTING: vertex bytes 12..15 are a normal, not a colour
	TexGen bool   // G_TEXTURE_GEN: s,t come from the normal (environment map)
	OtherL uint32 // othermode-L at first draw: the render mode (blender config)
	Verts  []Vertex
	Faces  [][3]int
}

// F3D geometry mode bits.
const (
	GeoLighting   = 0x00020000
	GeoTextureGen = 0x00040000
	GeoCullFront  = 0x00001000
	GeoCullBack   = 0x00002000
)

// Othermode-L bits (libultra's render-mode constants).
const (
	OMForceBlend = 0x4000 // FORCE_BL: the blender always mixes with memory
)

// New builds a walker over an RDRAM snapshot.
func New(ram []byte, verbose bool) *Walker {
	w := &Walker{RAM: ram, Verbose: verbose, Groups: map[string]*Group{}, tmemSrc: map[uint32]uint32{}}
	w.mtxStack = []Mtx44{Identity()}
	return w
}

// resolve maps a segmented or KSEG0 address to a physical RDRAM offset.
func (w *Walker) resolve(a uint32) uint32 {
	seg := a >> 24 & 0xF
	return (w.seg[seg] + a&0xFFFFFF) & 0x3FFFFF
}

func (w *Walker) be64(a uint32) uint64 { return binary.BigEndian.Uint64(w.RAM[a:]) }

func (w *Walker) cur() *Mtx44 { return &w.mtxStack[len(w.mtxStack)-1] }

func (w *Walker) logf(format string, args ...interface{}) {
	if w.Verbose {
		fmt.Printf("%s%s\n", strings.Repeat("  ", w.depth), fmt.Sprintf(format, args...))
	}
}

// Walk runs the display list at physical address pc to its G_ENDDL.
func (w *Walker) Walk(pc uint32) {
	for {
		cmd := w.be64(pc)
		op := uint32(cmd >> 56)
		w0 := uint32(cmd >> 32)
		w1 := uint32(cmd)

		switch op {
		case 0x01: // G_MTX
			p := w0 >> 16 & 0xFF
			a := w.resolve(w1)
			m := readMtx(w.RAM[a:])
			proj := p&1 != 0
			load := p&2 != 0
			push := p&4 != 0
			w.logf("G_MTX %s%s%s addr=%06X", iff(proj, "PROJ ", "MV "), iff(load, "LOAD", "MUL"), iff(push, " PUSH", ""), a)
			if proj {
				if load {
					w.proj = m
				} else {
					w.proj = Mul(m, w.proj)
				}
			} else {
				if push {
					w.mtxStack = append(w.mtxStack, *w.cur())
				}
				if load {
					*w.cur() = m
				} else {
					*w.cur() = Mul(m, *w.cur())
				}
				w.mtxAddr = a
			}
		case 0xBD: // G_POPMTX
			if len(w.mtxStack) > 1 {
				w.mtxStack = w.mtxStack[:len(w.mtxStack)-1]
			}
			w.logf("G_POPMTX")
		case 0x03: // G_MOVEMEM
			idx := w0 >> 16 & 0xFF
			a := w.resolve(w1)
			if idx == 0x80 { // G_MV_VIEWPORT: 4 x s16 scale, 4 x s16 translate, 10.2
				for i := 0; i < 4; i++ {
					w.VpScale[i] = int16(binary.BigEndian.Uint16(w.RAM[a+uint32(i)*2:]))
					w.VpTrans[i] = int16(binary.BigEndian.Uint16(w.RAM[a+8+uint32(i)*2:]))
				}
				w.logf("G_MOVEMEM viewport scale=(%d %d %d) trans=(%d %d %d) /4",
					w.VpScale[0], w.VpScale[1], w.VpScale[2], w.VpTrans[0], w.VpTrans[1], w.VpTrans[2])
			} else {
				w.logf("G_MOVEMEM idx=%02X addr=%06X", idx, a)
			}
		case 0x04: // G_VTX: (n-1)<<20 | v0<<16 | bytes-1
			n := w0>>20&0xF + 1
			v0 := w0 >> 16 & 0xF
			a := w.resolve(w1)
			w.logf("G_VTX v0=%d n=%d addr=%06X", v0, n, a)
			w.VtxLoads = append(w.VtxLoads, VtxLoad{a, n * 16})
			for i := uint32(0); i < n; i++ {
				v := readVertex(w.RAM[a+i*16:])
				m := w.cur()
				x, y, z := float64(v.X), float64(v.Y), float64(v.Z)
				v.WX = m[0][0]*x + m[1][0]*y + m[2][0]*z + m[3][0]
				v.WY = m[0][1]*x + m[1][1]*y + m[2][1]*z + m[3][1]
				v.WZ = m[0][2]*x + m[1][2]*y + m[2][2]*z + m[3][2]
				p := &w.proj
				v.CX = p[0][0]*v.WX + p[1][0]*v.WY + p[2][0]*v.WZ + p[3][0]
				v.CY = p[0][1]*v.WX + p[1][1]*v.WY + p[2][1]*v.WZ + p[3][1]
				v.CZ = p[0][2]*v.WX + p[1][2]*v.WY + p[2][2]*v.WZ + p[3][2]
				v.CW = p[0][3]*v.WX + p[1][3]*v.WY + p[2][3]*v.WZ + p[3][3]
				v.VpScale, v.VpTrans = w.VpScale, w.VpTrans
				w.vtx[v0+i] = v
			}
		case 0xBF: // G_TRI1: indices are byte offsets, 10 per vertex
			i0, i1, i2 := w1>>16&0xFF/10, w1>>8&0xFF/10, w1&0xFF/10
			w.logf("G_TRI1 %d %d %d (raw %06X)", i0, i1, i2, w1&0xFFFFFF)
			w.addTri(i0, i1, i2)
		case 0xB5: // G_QUAD (two tris)
			// w1 bytes: v0,v1,v2 then v0,v2,v3 per GBI
			i0, i1, i2, i3 := w1>>24&0xFF/10, w1>>16&0xFF/10, w1>>8&0xFF/10, w1&0xFF/10
			w.logf("G_QUAD %d %d %d %d", i0, i1, i2, i3)
			w.addTri(i0, i1, i2)
			w.addTri(i0, i2, i3)
		case 0x06: // G_DL
			a := w.resolve(w1)
			if w0>>16&0xFF == 0 {
				w.logf("G_DL call %06X", a)
				w.depth++
				w.Walk(a)
				w.depth--
			} else {
				w.logf("G_DL branch %06X", a)
				pc = a
				continue
			}
		case 0xB8: // G_ENDDL
			w.logf("G_ENDDL")
			return
		case 0xBC: // G_MOVEWORD
			index := uint32(cmd>>32) & 0xFF
			offset := uint32(cmd>>40) & 0xFFFF
			if index == 6 { // G_MW_SEGMENT
				w.seg[offset/4] = w1 & 0xFFFFFF
				w.logf("G_MOVEWORD segment[%d] = %06X", offset/4, w1&0xFFFFFF)
			} else {
				w.logf("G_MOVEWORD idx=%d off=%X val=%08X", index, offset, w1)
			}
		case 0xBB: // G_TEXTURE
			w.texScale[0] = w1 >> 16
			w.texScale[1] = w1 & 0xFFFF
			w.texTile = w0 >> 8 & 7
			w.texOn = w0&0xFF != 0
			w.logf("G_TEXTURE on=%v tile=%d scale=%04X,%04X", w.texOn, w.texTile, w.texScale[0], w.texScale[1])
		case 0xFD: // G_SETTIMG
			w.texImg = w.resolve(w1)
			w.texFmt = w0 >> 21 & 7
			w.texSiz = w0 >> 19 & 3
			w.logf("G_SETTIMG fmt=%d size=%d addr=%06X", w.texFmt, w.texSiz, w.texImg)
		case 0xF5: // Set_Tile: the tile index is in the LOW word (w1>>24), like
			// Set_Tile_Size — w0's bits 24-31 are the opcode, so reading the
			// index there sends every Set_Tile to tile 5 and the draw tile
			// never receives its format (the untextured-exports bug).
			t := &w.tile[w1>>24&7]
			t.Fmt, t.Size = w0>>21&7, w0>>19&3
			t.Line, t.Tmem = w0>>9&0x1FF, w0&0x1FF
			t.Pal = w1 >> 20 & 15
			t.CmT, t.MaskT, t.ShiftT = w1>>18&3, w1>>14&15, w1>>10&15
			t.CmS, t.MaskS, t.ShiftS = w1>>8&3, w1>>4&15, w1&15
			w.logf("G_SETTILE %d fmt=%d size=%d line=%d tmem=%d", w1>>24&7, t.Fmt, t.Size, t.Line, t.Tmem)
		case 0xF2: // Set_Tile_Size
			t := &w.tile[w1>>24&7]
			t.SL, t.TL = w0>>12&0xFFF, w0&0xFFF
			t.SH, t.TH = w1>>12&0xFFF, w1&0xFFF
			w.logf("G_SETTILESIZE %d %d..%d x %d..%d", w1>>24&7, t.SL>>2, t.SH>>2, t.TL>>2, t.TH>>2)
		case 0xF3: // Load_Block: loads land in TMEM; draws sample through another
			// tile pointed at the same TMEM address, so the source is remembered
			// by TMEM word, not by tile index.
			ti := w1 >> 24 & 7
			w.tmemSrc[w.tile[ti].Tmem] = w.texImg
			w.logf("G_LOADBLOCK tile=%d sh=%d dxt=%d src=%06X tmem=%d", ti, w1>>12&0xFFF, w1&0xFFF, w.texImg, w.tile[ti].Tmem)
		case 0xF4: // Load_Tile
			ti := w1 >> 24 & 7
			w.tmemSrc[w.tile[ti].Tmem] = w.texImg
			w.logf("G_LOADTILE tile=%d src=%06X", ti, w.texImg)
		case 0xF0: // Load_TLUT
			w.tlut = w.texImg
			w.logf("G_LOADTLUT src=%06X", w.texImg)
		case 0xB7: // G_SETGEOMETRYMODE
			w.geoMode |= w1
			w.logf("G_SETGEOMETRYMODE %08X -> %08X", w1, w.geoMode)
		case 0xB6: // G_CLEARGEOMETRYMODE
			w.geoMode &^= w1
			w.logf("G_CLEARGEOMETRYMODE %08X -> %08X", w1, w.geoMode)
		case 0xB9: // G_SETOTHERMODE_L: shift/len select a field, w1 is pre-shifted
			shift := w0 >> 8 & 0xFF
			length := w0 & 0xFF
			mask := (uint32(1)<<length - 1) << shift
			w.otherL = w.otherL&^mask | w1&mask
			w.logf("G_SETOTHERMODE_L shift=%d len=%d val=%08X -> %08X", shift, length, w1, w.otherL)
		case 0xB2, 0xB3, // G_RDPHALF_CONT / G_RDPHALF_2
			0xB4, // G_PERSPNORMALIZE
			0xBA, 0xBE, 0xE4, 0xE6, 0xE7, 0xE8, 0xE9, 0xED, 0xEE, 0xEF,
			0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFB, 0xFC, 0xFE, 0xFF:
			// othermode-H, cull, and RDP passthroughs: not needed for geometry
			// extraction, logged for the record.
			w.logf("op %02X %08X %08X", op, w0, w1)
		case 0x00: // G_SPNOOP
		default:
			fmt.Printf("UNHANDLED op %02X at %06X: %08X %08X\n", op, pc, w0, w1)
			return
		}
		pc += 8
	}
}

func (w *Walker) addTri(i0, i1, i2 uint32) {
	w.NTris++
	// Group key: the modelview matrix content + the texture the draw samples,
	// through the tile G_TEXTURE selected and the TMEM address it reads. With
	// G_TEXTURE off the microcode emits untextured (shade-only) triangles, so
	// the draw belongs to the untextured group whatever the tiles hold.
	m := w.cur()
	t := w.tile[w.texTile]
	t.Img = w.tmemSrc[t.Tmem]
	if !w.texOn {
		t = TileDesc{}
	}
	key := fmt.Sprintf("%x-%06X-%d", mtxHash(m), t.Img, t.Fmt)
	g, ok := w.Groups[key]
	if !ok {
		g = &Group{
			Name:   fmt.Sprintf("group-%03d-tex%06X", len(w.Groups), t.Img),
			TexImg: t.Img, Tile: t, TLUT: w.tlut, Scale: w.texScale, Mtx: *m,
			Lit: w.geoMode&GeoLighting != 0, TexGen: w.geoMode&GeoTextureGen != 0,
			OtherL: w.otherL,
		}
		w.Groups[key] = g
		w.Order = append(w.Order, key)
		w.logf("new %s mtx=%06X T=(%.2f %.2f %.2f)", g.Name, w.mtxAddr, m[3][0], m[3][1], m[3][2])
	}
	base := len(g.Verts)
	for _, i := range []uint32{i0, i1, i2} {
		g.Verts = append(g.Verts, w.vtx[i])
	}
	g.Faces = append(g.Faces, [3]int{base, base + 1, base + 2})
	w.Tris = append(w.Tris, Tri{
		V:       [3]Vertex{w.vtx[i0], w.vtx[i1], w.vtx[i2]},
		TexImg:  t.Img,
		GeoMode: w.geoMode,
		VpScale: w.VpScale, VpTrans: w.VpTrans,
		Group:   g.Name,
		MtxAddr: w.mtxAddr,
	})
}

// Tile returns the tile-descriptor state of tile i (0..7) as the walk left it.
// A material display-list template configures its tiles and ends without ever
// drawing, so this — not a draw group — is how its texture parameters are read.
func (w *Walker) Tile(i int) TileDesc { return w.tile[i&7] }

// TexTile is the tile index the last G_TEXTURE selected for drawing.
func (w *Walker) TexTile() uint32 { return w.texTile }

// TexScale is the s,t scale the last G_TEXTURE set.
func (w *Walker) TexScale() [2]uint32 { return w.texScale }

// TLUT is the RAM address of the palette the last Load_TLUT read.
func (w *Walker) TLUT() uint32 { return w.tlut }

// Ordered returns the draw groups in first-draw order.
func (w *Walker) Ordered() []*Group {
	var out []*Group
	for _, k := range w.Order {
		out = append(out, w.Groups[k])
	}
	return out
}

func mtxHash(m *Mtx44) uint64 {
	var h uint64 = 1469598103934665603
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			h ^= uint64(int64(m[i][j] * 65536))
			h *= 1099511628211
		}
	}
	return h
}

func readVertex(d []byte) Vertex {
	return Vertex{
		X: int16(binary.BigEndian.Uint16(d[0:])), Y: int16(binary.BigEndian.Uint16(d[2:])),
		Z:    int16(binary.BigEndian.Uint16(d[4:])),
		Flag: binary.BigEndian.Uint16(d[6:]),
		S:    int16(binary.BigEndian.Uint16(d[8:])), T: int16(binary.BigEndian.Uint16(d[10:])),
		R: d[12], G: d[13], B: d[14], A: d[15],
	}
}

// readMtx reads libultra's 4x4 s15.16 matrix: 32 bytes of integer parts then
// 32 bytes of fraction parts, row-major.
func readMtx(d []byte) Mtx44 {
	var m Mtx44
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			hi := int16(binary.BigEndian.Uint16(d[(i*4+j)*2:]))
			lo := binary.BigEndian.Uint16(d[32+(i*4+j)*2:])
			m[i][j] = float64(hi) + float64(lo)/65536
		}
	}
	return m
}

// Mul multiplies row-vector matrices: v*(Mul(a,b)) == (v*a)*b.
func Mul(a, b Mtx44) Mtx44 {
	var r Mtx44
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				r[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return r
}

// Identity returns the identity matrix.
func Identity() Mtx44 {
	var m Mtx44
	for i := 0; i < 4; i++ {
		m[i][i] = 1
	}
	return m
}

// InvertAffine inverts a rigid-ish modelview (rotation+scale+translation).
func InvertAffine(m Mtx44) Mtx44 {
	// Invert the 3x3 by adjugate, then the translation.
	a := [3][3]float64{
		{m[0][0], m[0][1], m[0][2]},
		{m[1][0], m[1][1], m[1][2]},
		{m[2][0], m[2][1], m[2][2]},
	}
	det := a[0][0]*(a[1][1]*a[2][2]-a[1][2]*a[2][1]) -
		a[0][1]*(a[1][0]*a[2][2]-a[1][2]*a[2][0]) +
		a[0][2]*(a[1][0]*a[2][1]-a[1][1]*a[2][0])
	if det == 0 {
		return Identity()
	}
	inv := func(i, j int) float64 {
		i1, i2 := (i+1)%3, (i+2)%3
		j1, j2 := (j+1)%3, (j+2)%3
		return (a[j1][i1]*a[j2][i2] - a[j1][i2]*a[j2][i1]) / det
	}
	var r Mtx44
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			r[i][j] = inv(i, j)
		}
	}
	// translation' = -T * R'
	for j := 0; j < 3; j++ {
		r[3][j] = -(m[3][0]*r[0][j] + m[3][1]*r[1][j] + m[3][2]*r[2][j])
	}
	r[3][3] = 1
	return r
}

func iff(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

// WriteOBJ writes a group's world-space triangles as a Wavefront OBJ.
func WriteOBJ(dir string, g *Group) {
	var b strings.Builder
	fmt.Fprintf(&b, "o %s\n", g.Name)
	for _, v := range g.Verts {
		fmt.Fprintf(&b, "v %f %f %f %f %f %f\n", v.WX, v.WY, v.WZ,
			float64(v.R)/255, float64(v.G)/255, float64(v.B)/255)
	}
	for _, f := range g.Faces {
		fmt.Fprintf(&b, "f %d %d %d\n", f[0]+1, f[1]+1, f[2]+1)
	}
	os.WriteFile(fmt.Sprintf("%s/%s.obj", dir, g.Name), []byte(b.String()), 0o644)
}
