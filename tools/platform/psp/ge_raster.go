package psp

// ge_raster.go executes a captured GE display list into the framebuffer: it walks the
// commands maintaining render state (framebuffer target, viewport, matrices, vertex
// format, clear mode) and, on each PRIM, decodes the vertices and software-rasterizes
// the primitive. It covers the common path — framebuffer config, CLEAR fills,
// through-mode (2D screen-space) and transformed (3D) triangles/strips/fans/sprites
// with per-vertex or material colour. Texturing is sampled when a texture is bound.

import "math"

// GE register command numbers (subset; pspsdk guInternal names).
const (
	cVADDR     = 0x01
	cPRIM      = 0x04
	cBASE      = 0x10
	cVTYPE     = 0x12
	cOFFADDR   = 0x13
	cWORLDN    = 0x3A
	cWORLDD    = 0x3B
	cVIEWN     = 0x3C
	cVIEWD     = 0x3D
	cPROJN     = 0x3E
	cPROJD     = 0x3F
	cVPXSCALE  = 0x42
	cVPYSCALE  = 0x43
	cVPZSCALE  = 0x44
	cVPXCENTER = 0x45
	cVPYCENTER = 0x46
	cVPZCENTER = 0x47
	cOFFSETX   = 0x4C
	cOFFSETY   = 0x4D
	cMATCOLOR  = 0x53
	cTEXADDR0  = 0xA0
	cTEXBW0    = 0xA8
	cTEXSIZE0  = 0xB8
	cTEXFORMAT = 0xC3
	cTEXENABLE = 0x1E
	cFBPTR     = 0x9C
	cFBWIDTH   = 0x9D
	cFBPIXFMT  = 0xD2
	cCLEARMODE = 0xD3
)

// geState is the render state accumulated while walking a list.
type geState struct {
	fbLow, fbHigh uint32
	fbStride      uint32
	fbFmt         uint32

	base    uint32
	vaddr   uint32
	offAddr uint32
	vtype   uint32

	vpXS, vpYS, vpZS float32 // viewport scale
	vpXC, vpYC, vpZC float32 // viewport center

	world, view, proj          [16]float32
	worldIdx, viewIdx, projIdx int

	matColor uint32
	clearOn  bool

	texEnable  bool
	texAddr    uint32
	texStride  uint32
	texW, texH uint32
	texFmt     uint32
}

func (m *Machine) rasterList(list GeList) {
	s := &geState{}
	ident(&s.world)
	ident(&s.view)
	ident(&s.proj)
	s.vpXS, s.vpYS = dispW/2, -dispH/2
	s.vpXC, s.vpYC = 2048+dispW/2, 2048+dispH/2

	for _, w := range list.Words {
		cmd := w >> 24
		arg := w & 0x00FFFFFF
		switch cmd {
		case cBASE:
			s.base = (arg << 8) & 0xFF000000
		case cVADDR:
			s.vaddr = s.base | arg
		case cOFFADDR:
			s.offAddr = arg << 8
		case cVTYPE:
			s.vtype = arg
		case cFBPTR:
			s.fbLow = arg
		case cFBWIDTH:
			s.fbStride = arg & 0xFFFF
			s.fbHigh = (arg >> 16) & 0xFF
		case cFBPIXFMT:
			s.fbFmt = arg & 3
		case cCLEARMODE:
			s.clearOn = arg&1 != 0
		case cMATCOLOR:
			s.matColor = arg
		case cVPXSCALE:
			s.vpXS = f24(arg)
		case cVPYSCALE:
			s.vpYS = f24(arg)
		case cVPZSCALE:
			s.vpZS = f24(arg)
		case cVPXCENTER:
			s.vpXC = f24(arg)
		case cVPYCENTER:
			s.vpYC = f24(arg)
		case cVPZCENTER:
			s.vpZC = f24(arg)
		case cWORLDN:
			s.worldIdx = 0
		case cWORLDD:
			matPush(&s.world, &s.worldIdx, arg)
		case cVIEWN:
			s.viewIdx = 0
		case cVIEWD:
			matPush(&s.view, &s.viewIdx, arg)
		case cPROJN:
			s.projIdx = 0
		case cPROJD:
			matPush16(&s.proj, &s.projIdx, arg)
		case cTEXENABLE:
			s.texEnable = arg&1 != 0
		case cTEXADDR0:
			s.texAddr = (s.texAddr & 0xFF000000) | (arg & 0xFFFFFF)
		case cTEXBW0:
			s.texStride = arg & 0xFFFF
			s.texAddr = (s.texAddr & 0x00FFFFFF) | ((arg & 0xFF0000) << 8)
		case cTEXSIZE0:
			s.texW = 1 << (arg & 0xFF)
			s.texH = 1 << ((arg >> 8) & 0xFF)
		case cTEXFORMAT:
			s.texFmt = arg & 0xF
		case cPRIM:
			m.drawPrim(s, arg)
		}
	}
	if s.fbStride != 0 {
		m.fbAddr = s.fbAddress()
		m.fbWidth = s.fbStride
		m.fbFormat = s.fbFmt
	}
}

// fbAddress is the CPU address of the current framebuffer.
func (s *geState) fbAddress() uint32 { return vramBase | (s.fbHigh << 24) | s.fbLow }

// drawPrim decodes and rasterizes one PRIM command (arg: bits 0-15 count, 16-18 type).
func (m *Machine) drawPrim(s *geState, arg uint32) {
	count := int(arg & 0xFFFF)
	ptype := (arg >> 16) & 7
	if count == 0 || s.vaddr == 0 {
		return
	}
	verts := m.decodeVerts(s, count)
	if len(verts) < 1 {
		return
	}
	switch ptype {
	case 3: // triangles
		for i := 0; i+2 < len(verts); i += 3 {
			m.rasterTri(s, verts[i], verts[i+1], verts[i+2])
		}
	case 4: // triangle strip
		for i := 0; i+2 < len(verts); i++ {
			a, b, c := verts[i], verts[i+1], verts[i+2]
			if i&1 == 1 {
				b, c = c, b
			}
			m.rasterTri(s, a, b, c)
		}
	case 5: // triangle fan
		for i := 1; i+1 < len(verts); i++ {
			m.rasterTri(s, verts[0], verts[i], verts[i+1])
		}
	case 6: // sprites (pairs of corners)
		for i := 0; i+1 < len(verts); i += 2 {
			m.rasterSprite(s, verts[i], verts[i+1])
		}
	}
}

// vert is a decoded, screen-space vertex.
type vert struct {
	x, y, z    float32
	u, v       float32
	r, g, b, a byte
}

// decodeVerts reads count vertices from s.vaddr per the vertex type, applying the
// transform (world*view*proj*viewport) for non-through primitives.
func (m *Machine) decodeVerts(s *geState, count int) []vert {
	tfmt := s.vtype & 3
	cfmt := (s.vtype >> 2) & 7
	pfmt := (s.vtype >> 7) & 3
	through := s.vtype&(1<<23) != 0

	texSz := []uint32{0, 2, 4, 8}[tfmt]
	colSz := map[uint32]uint32{0: 0, 4: 2, 5: 2, 6: 2, 7: 4}[cfmt]
	posSz := []uint32{0, 3, 6, 12}[pfmt]
	stride := texSz + colSz + posSz
	if stride == 0 {
		return nil
	}
	align := uint32(1)
	for _, s2 := range []uint32{texSz / uint32(max1(tfmt)), colSz, posSz / uint32(max1(pfmt))} {
		if s2 > align {
			align = s2
		}
	}
	stride = (stride + align - 1) &^ (align - 1)

	mvp := mul4(mul4(s.proj, s.view), s.world)
	out := make([]vert, 0, count)
	addr := s.vaddr
	for i := 0; i < count; i++ {
		p := addr
		var vv vert
		vv.a = 0xFF
		if tfmt != 0 {
			vv.u, vv.v = readUV(m, p, tfmt)
			p += texSz
		}
		if cfmt != 0 {
			vv.r, vv.g, vv.b, vv.a = readColor(m, p, cfmt)
			p += colSz
		} else {
			vv.r, vv.g, vv.b = byte(s.matColor), byte(s.matColor>>8), byte(s.matColor>>16)
		}
		px, py, pz := readPos(m, p, pfmt)
		if through {
			vv.x, vv.y, vv.z = px, py, pz
		} else {
			vv.x, vv.y, vv.z = transform(mvp, s, px, py, pz)
		}
		out = append(out, vv)
		addr += stride
	}
	return out
}

func max1(fmt uint32) uint32 {
	if fmt == 0 {
		return 1
	}
	return 3
}

func readUV(m *Machine, p, fmt uint32) (float32, float32) {
	switch fmt {
	case 1:
		return float32(m.Read(p)), float32(m.Read(p + 1))
	case 2:
		return float32(u16(m, p)), float32(u16(m, p+2))
	default:
		return f32(m, p), f32(m, p+4)
	}
}

func readColor(m *Machine, p, fmt uint32) (byte, byte, byte, byte) {
	switch fmt {
	case 7: // 8888
		c := m.read32(p)
		return byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)
	case 6: // 4444
		c := u16(m, p)
		e := func(v uint16) byte { return byte((v & 0xF) * 17) }
		return e(c), e(c >> 4), e(c >> 8), e(c >> 12)
	case 5: // 5551
		c := u16(m, p)
		e := func(v uint16, b uint16) byte { return byte(uint32(v&((1<<b)-1)) * 255 / ((1 << b) - 1)) }
		a := byte(0)
		if c&0x8000 != 0 {
			a = 0xFF
		}
		return e(c, 5), e(c>>5, 5), e(c>>10, 5), a
	default: // 565
		c := u16(m, p)
		e := func(v uint16, b uint16) byte { return byte(uint32(v&((1<<b)-1)) * 255 / ((1 << b) - 1)) }
		return e(c, 5), e(c>>5, 6), e(c>>11, 5), 0xFF
	}
}

func readPos(m *Machine, p, fmt uint32) (float32, float32, float32) {
	switch fmt {
	case 1: // s8
		return float32(int8(m.Read(p))), float32(int8(m.Read(p + 1))), float32(int8(m.Read(p + 2)))
	case 2: // s16
		return float32(int16(u16(m, p))), float32(int16(u16(m, p+2))), float32(int16(u16(m, p+4)))
	default: // float
		return f32(m, p), f32(m, p+4), f32(m, p+8)
	}
}

// transform applies the model-view-projection then the viewport to a model vertex.
func transform(mvp [16]float32, s *geState, x, y, z float32) (float32, float32, float32) {
	cx := mvp[0]*x + mvp[4]*y + mvp[8]*z + mvp[12]
	cy := mvp[1]*x + mvp[5]*y + mvp[9]*z + mvp[13]
	cz := mvp[2]*x + mvp[6]*y + mvp[10]*z + mvp[14]
	cw := mvp[3]*x + mvp[7]*y + mvp[11]*z + mvp[15]
	if cw == 0 {
		cw = 1
	}
	nx, ny, nz := cx/cw, cy/cw, cz/cw
	sx := nx*s.vpXS + s.vpXC - 2048
	sy := ny*s.vpYS + s.vpYC - 2048
	sz := nz*s.vpZS + s.vpZC
	return sx, sy, sz
}

// --- matrix helpers --------------------------------------------------------

func ident(m *[16]float32) {
	*m = [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

// matPush appends a 4x3 world/view matrix element (column-major, 3 rows per column).
func matPush(m *[16]float32, idx *int, arg uint32) {
	if *idx >= 12 {
		return
	}
	col := *idx / 3
	row := *idx % 3
	m[col*4+row] = math.Float32frombits(arg << 8)
	if row == 2 {
		m[col*4+3] = 0
	}
	if *idx == 11 {
		m[15] = 1
	}
	*idx++
}

// matPush16 appends a 4x4 projection matrix element.
func matPush16(m *[16]float32, idx *int, arg uint32) {
	if *idx >= 16 {
		return
	}
	m[*idx] = math.Float32frombits(arg << 8)
	*idx++
}

func mul4(a, b [16]float32) [16]float32 {
	var r [16]float32
	for c := 0; c < 4; c++ {
		for rr := 0; rr < 4; rr++ {
			var s float32
			for k := 0; k < 4; k++ {
				s += a[k*4+rr] * b[c*4+k]
			}
			r[c*4+rr] = s
		}
	}
	return r
}

// f24 expands a GE float argument (top 24 bits of an IEEE float) to float32.
func f24(arg uint32) float32 { return math.Float32frombits(arg << 8) }

func u16(m *Machine, a uint32) uint16  { return uint16(m.Read(a)) | uint16(m.Read(a+1))<<8 }
func f32(m *Machine, a uint32) float32 { return math.Float32frombits(m.read32(a)) }
