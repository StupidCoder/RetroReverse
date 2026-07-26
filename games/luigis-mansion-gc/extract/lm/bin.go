package lm

// bin.go reads the in-game .bin model format — the mansion's rooms and
// furniture (Iwamoto/map*/room_*.arc) — as reverse-engineered from the
// renderer at 0x8001D8C8..0x8001DD90. Like the .mdl, the file is its own
// runtime object: a version byte and a name, then a table of 21 section
// offsets at +0x0C that the loader patches into pointers in place:
//
//	[0]  textures   12 B: {u16 w, u16 h, u8 gxFormat, u8 -, u16 -, u32 dataOff (absolute)}
//	[1]  samplers   12 B: {u8 -, u8 texIdx, s16 paletteIdx, u8 wrapS, u8 wrapT, u8[6] -}
//	[2]  positions  s16[3]
//	[3]  normals    f32[3]
//	[6]  texcoords  f32[2]
//	[10] materials  40 B: {u8 flags, u8, u8, u8 rgba[4]@3, ..., s8 samplerIdx@9,
//	                       ..., s16 stageBlock@0x1a → section 13 when present}
//	[11] meshes     24 B: {u16 -, u16 dlSize (32-byte units), u32 -, u16 attrMask,
//	                       u16 -, u32 dlOff (relative to this section), u32 -, u32 -}
//	[12] graph     140 B: {s16 parent, child, next, prev; u16 partCountA@0x08;
//	                       f32 scale[3]@0x0c, rot°[3]@0x18, trans[3]@0x24;
//	                       f32 bboxMin[3]@0x30, bboxMax[3]@0x3c, radius@0x48;
//	                       u16 partCountB@0x4c; u32 partListOff@0x50 (section-relative)}
//	     the part list is {u16 material, u16 mesh} pairs, the A-count then the
//	     B-count (the renderer's opaque and translucent passes)
//
// A mesh display list is a raw GX stream of indexed primitives — no matrix
// bytes; the node's composed TRS places everything. Each vertex is u16
// indices: position, then normal and/or texcoord as the attribute mask says
// (0x100 with three indices per vertex on textured meshes; two-index meshes
// carry position + normal).

import (
	"encoding/binary"
	"fmt"
	"math"
)

// BinTexture is a decoded texture.
type BinTexture struct {
	Width, Height int
	Format        uint8
	Pixels        []byte // RGBA
}

// BinSampler binds a texture.
type BinSampler struct {
	Texture int
	Palette int
}

// BinMaterial is what the exporter needs of the 40-byte record.
type BinMaterial struct {
	Tint    [4]uint8
	Sampler int // -1 for untextured
}

// BinMesh is one indexed display list.
type BinMesh struct {
	AttrMask uint16
	DL       []byte
}

// BinPair draws one mesh with one material.
type BinPair struct{ Material, Mesh int }

// BinNode is one node of the scene graph.
type BinNode struct {
	Parent, Child, Next, Prev int
	Scale, Rot, Trans         [3]float32 // rotation in degrees
	Pairs                     []BinPair
}

// Bin is a parsed .bin model.
type Bin struct {
	Name      string
	Positions [][3]float32
	Normals   [][3]float32
	Texcoords [][2]float32
	Textures  []BinTexture
	Samplers  []BinSampler
	Materials []BinMaterial
	Meshes    []BinMesh
	Nodes     []BinNode
}

// ParseBin reads a .bin model.
func ParseBin(b []byte) (*Bin, error) {
	if len(b) < 0x60 || b[0] == 0 {
		return nil, fmt.Errorf("lm: not a bin model")
	}
	u16 := func(o uint32) int { return int(binary.BigEndian.Uint16(b[o:])) }
	s16 := func(o uint32) int { return int(int16(binary.BigEndian.Uint16(b[o:]))) }
	u32 := func(o uint32) uint32 { return binary.BigEndian.Uint32(b[o:]) }
	f32 := func(o uint32) float32 { return math.Float32frombits(u32(o)) }
	var offs [21]uint32
	for i := range offs {
		offs[i] = u32(0xC + uint32(i)*4)
	}
	name := ""
	for i := 1; i < 0xC && b[i] != 0; i++ {
		name += string(b[i])
	}
	m := &Bin{Name: name}

	// Section extents: for counts, each section runs to the next non-zero
	// offset (they are laid out in ascending file order).
	end := func(sec int) uint32 {
		best := uint32(len(b))
		for _, o := range offs {
			if o > offs[sec] && o < best {
				best = o
			}
		}
		return best
	}

	// Textures: headers run until the first image data begins.
	if t := offs[0]; t != 0 {
		firstData := uint32(len(b))
		for o := t; o+12 <= firstData; o += 12 {
			w, h := u16(o), u16(o+2)
			fmtb := b[o+4]
			data := u32(o + 8)
			if data < firstData {
				firstData = data
			}
			if w == 0 || h == 0 || w > 1024 || h > 1024 {
				break
			}
			px, err := decodeGXTexture(gxFmtIndex(fmtb), w, h, b[data:])
			if err != nil {
				return nil, fmt.Errorf("lm: bin texture at %#x: %w", o, err)
			}
			m.Textures = append(m.Textures, BinTexture{Width: w, Height: h, Format: fmtb, Pixels: px})
		}
	}
	if s := offs[1]; s != 0 {
		for o := s; o+12 <= end(1); o += 12 {
			m.Samplers = append(m.Samplers, BinSampler{Texture: int(b[o+1]), Palette: s16(o + 2)})
		}
	}
	if p := offs[2]; p != 0 {
		for o := p; o+6 <= end(2); o += 6 {
			m.Positions = append(m.Positions, [3]float32{float32(s16(o)), float32(s16(o + 2)), float32(s16(o + 4))})
		}
	}
	if n := offs[3]; n != 0 {
		for o := n; o+12 <= end(3); o += 12 {
			m.Normals = append(m.Normals, [3]float32{f32(o), f32(o + 4), f32(o + 8)})
		}
	}
	if t := offs[6]; t != 0 {
		for o := t; o+8 <= end(6); o += 8 {
			m.Texcoords = append(m.Texcoords, [2]float32{f32(o), f32(o + 4)})
		}
	}
	if mt := offs[10]; mt != 0 {
		for o := mt; o+40 <= end(10); o += 40 {
			m.Materials = append(m.Materials, BinMaterial{
				Tint:    [4]uint8{b[o+3], b[o+4], b[o+5], b[o+6]},
				Sampler: int(int8(b[o+9])),
			})
		}
	}
	meshSec := offs[11]
	if meshSec != 0 {
		// The record table is followed by the display lists inside the same
		// section: records run until the earliest display list begins.
		tableEnd := end(11) - meshSec
		for o := meshSec; o+24 <= meshSec+tableEnd; o += 24 {
			dlRel := u32(o + 12)
			if dlRel < tableEnd {
				tableEnd = dlRel
			}
			if o-meshSec >= tableEnd {
				break
			}
			size := uint32(u16(o+2)) * 32
			dlOff := meshSec + dlRel
			if dlOff+size > uint32(len(b)) {
				break
			}
			m.Meshes = append(m.Meshes, BinMesh{AttrMask: uint16(u16(o + 8)), DL: b[dlOff : dlOff+size]})
		}
	}
	if g := offs[12]; g != 0 {
		count := int(end(12)-g) / 140
		for i := 0; i < count; i++ {
			o := g + uint32(i)*140
			n := BinNode{
				Parent: s16(o), Child: s16(o + 2), Next: s16(o + 4), Prev: s16(o + 6),
			}
			for c := uint32(0); c < 3; c++ {
				n.Scale[c] = f32(o + 0x0C + c*4)
				n.Rot[c] = f32(o + 0x18 + c*4)
				n.Trans[c] = f32(o + 0x24 + c*4)
			}
			pairs := int(u16(o+0x08)) + int(u16(o+0x4C))
			list := g + u32(o+0x50)
			for k := 0; k < pairs && list+uint32(k)*4+4 <= uint32(len(b)); k++ {
				n.Pairs = append(n.Pairs, BinPair{Material: u16(list + uint32(k)*4), Mesh: u16(list + uint32(k)*4 + 2)})
			}
			m.Nodes = append(m.Nodes, n)
		}
	}
	return m, nil
}

// gxFmtIndex maps a raw GX texture format id to this package's decode index
// (the .mdl files store the index; the .bin files store the GX id).
func gxFmtIndex(gx uint8) uint8 {
	switch gx {
	case 0x0:
		return 3 // I4
	case 0x1:
		return 4 // I8
	case 0x2:
		return 5 // IA4
	case 0x3:
		return 6 // IA8
	case 0x4:
		return 7 // RGB565
	case 0x5:
		return 8 // RGB5A3
	case 0x6:
		return 9 // RGBA8
	case 0xE:
		return 10 // CMPR
	}
	return gx // CI formats keep their index; the decoder will refuse politely
}

// ParseBinDL decodes a mesh's display list. Three-index vertices are
// position, normal, texcoord; two-index are position, normal. The attribute
// mask picks the width; when a stream refuses to parse at that width the
// other widths are tried, with position indices validated against the array.
func (m *Bin) ParseBinDL(mesh *BinMesh) ([]DLPrimitive, error) {
	first := 2
	if mesh.AttrMask&0x100 != 0 {
		first = 3
	}
	var lastErr error
	for _, nidx := range []int{first, 5 - first, 4, 1} {
		prims, err := m.parseBinDLWidth(mesh.DL, nidx)
		if err == nil {
			return prims, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (m *Bin) parseBinDLWidth(dl []byte, nidx int) ([]DLPrimitive, error) {
	var prims []DLPrimitive
	i := 0
	for i < len(dl) {
		op := dl[i]
		if op == 0 {
			i++
			continue
		}
		kind := op & 0xF8
		if kind < 0x80 || kind > 0xB8 {
			return nil, fmt.Errorf("lm: bin display-list opcode %#x at %d (width %d)", op, i, nidx)
		}
		count := int(binary.BigEndian.Uint16(dl[i+1:]))
		i += 3
		if i+count*2*nidx > len(dl) {
			return nil, fmt.Errorf("lm: bin display list truncated (width %d)", nidx)
		}
		p := DLPrimitive{Kind: kind}
		for v := 0; v < count; v++ {
			var vx DLVertex
			vx.Pos = binary.BigEndian.Uint16(dl[i:])
			if int(vx.Pos) >= len(m.Positions) {
				return nil, fmt.Errorf("lm: bin position index %d out of range (width %d)", vx.Pos, nidx)
			}
			if nidx >= 2 {
				vx.HasNrm = true
				vx.Nrm = binary.BigEndian.Uint16(dl[i+2:])
			}
			if nidx >= 3 {
				vx.HasTex = true
				vx.Tex = binary.BigEndian.Uint16(dl[i+4:])
			}
			i += 2 * nidx
			p.Verts = append(p.Verts, vx)
		}
		prims = append(prims, p)
	}
	return prims, nil
}
