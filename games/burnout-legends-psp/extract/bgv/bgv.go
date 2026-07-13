// Package bgv decodes Burnout Legends' vehicle models: the .bgv files under
// PSP_GAME/USRDIR/pveh/COMP.
//
// A .bgv is a self-contained, GPU-ready model. The engine loads the file
// verbatim at a base address, patches a handful of pointer slots (each stored
// as a plain file offset, fixed up to base+offset), and hands the geometry
// straight to the GE — the vertex data in the file IS the vertex buffer the
// hardware reads, and the strip lists in the file ARE display-list fragments.
// That was established by dumping the vertex bytes the GE fetched out of the
// running game and finding them byte-identical in the file.
//
// Layout:
//
//	header +0x08  file size
//	       +0x4C  \
//	       +0x50   > pointers to three mesh tables — the three detail levels
//	       +0x54  /
//	       +0x60  pointer to the texture region
//
//	mesh table   a run of u32 offsets (relative to the table), zero-terminated;
//	             each names one mesh record
//
//	mesh record  offsets below are relative to the record + 0x10:
//	       +0x30..+0x38  three floats: the model scale
//	       +0x44  vertex stream offset      +0x48  vertex count
//	       +0x54  strip list offset         +0x58  strip list word count
//
//	strip list   GE PRIM command words (0x04<<24 | type<<16 | vertexCount),
//	             terminated by a RET (0x0B000000) — the engine CALLs it
//
//	vertex data  GE-native, packed per the PSP vertex layout. The cars use
//	             vertex type 0x122: u16 texcoords, an 8-bit normal, and 16-bit
//	             positions, 14 bytes to a vertex. The fixed-point components
//	             are fractional (u16/32768, s8/128, s16/32768), as the hardware
//	             reads them; the model scale then puts the mesh in world units.
package bgv

import (
	"encoding/binary"
	"fmt"
	"math"
)

// vtypeCar is the vertex format the vehicle meshes are stored in: u16 UV,
// s8 normal, s16 position.
// The cars are stored in vertex type 0x122 (u16 UV, s8 normal, s16 position).
const vtypeCar = 0x122

// Vertex is one decoded vertex, in model units.
type Vertex struct {
	X, Y, Z    float32
	NX, NY, NZ float32
	U, V       float32
}

// Strip is one triangle strip: an index run into the mesh's vertex slice.
type Strip struct {
	Type  int // GE primitive type (4 = triangle strip)
	Start int // first vertex
	Count int
}

// Mesh is one part of a vehicle at one detail level.
type Mesh struct {
	Offset int // the record's file offset, for diagnostics
	Verts  []Vertex
	Strips []Strip
	ScaleX float32
	ScaleY float32
	ScaleZ float32
}

// Model is a decoded .bgv: its meshes grouped by detail level (highest first
// in the file's own order — see Levels).
type Model struct {
	Levels [][]Mesh
}

// Tris flattens the mesh's strips into triangles, dropping the degenerate ones
// the strips use to stitch disjoint runs together.
func (m *Mesh) Tris() [][3]uint32 {
	var out [][3]uint32
	for _, s := range m.Strips {
		if s.Type != 4 { // only triangle strips appear in these models
			continue
		}
		for i := 0; i+2 < s.Count; i++ {
			a, b, c := uint32(s.Start+i), uint32(s.Start+i+1), uint32(s.Start+i+2)
			if i&1 == 1 {
				b, c = c, b
			}
			if a == b || b == c || a == c {
				continue
			}
			out = append(out, [3]uint32{a, b, c})
		}
	}
	return out
}

// Parse decodes a .bgv file.
func Parse(data []byte) (*Model, error) {
	if len(data) < 0x80 {
		return nil, fmt.Errorf("bgv: too small (%d bytes)", len(data))
	}
	if size := int(le32(data, 0x08)); size != len(data) {
		return nil, fmt.Errorf("bgv: header size %d != file size %d", size, len(data))
	}
	m := &Model{}
	for _, hdrOff := range []int{0x4C, 0x50, 0x54} {
		table := int(le32(data, hdrOff))
		if table <= 0 || table >= len(data) {
			continue
		}
		meshes, err := parseTable(data, table)
		if err != nil {
			return nil, err
		}
		if len(meshes) > 0 {
			m.Levels = append(m.Levels, meshes)
		}
	}
	if len(m.Levels) == 0 {
		return nil, fmt.Errorf("bgv: no mesh tables")
	}
	return m, nil
}

// parseTable reads one detail level: a zero-terminated list of record offsets.
func parseTable(data []byte, table int) ([]Mesh, error) {
	var out []Mesh
	for i := 0; ; i++ {
		p := table + 4*i
		if p+4 > len(data) {
			break
		}
		rel := int(le32(data, p))
		if rel == 0 {
			break
		}
		rec := table + rel
		if rec+0x60 > len(data) {
			return nil, fmt.Errorf("bgv: record at 0x%X out of range", rec)
		}
		mesh, err := parseMesh(data, rec)
		if err != nil {
			return nil, err
		}
		out = append(out, *mesh)
	}
	return out, nil
}

// parseMesh reads one mesh record. Its internal offsets are relative to the
// record plus 0x10.
func parseMesh(data []byte, rec int) (*Mesh, error) {
	base := rec + 0x10
	vOff := base + int(le32(data, rec+0x44))
	vCount := int(le32(data, rec+0x48))
	sOff := base + int(le32(data, rec+0x54))
	sWords := int(le32(data, rec+0x58))

	if vCount < 0 || sWords < 0 || vOff < 0 || sOff < 0 ||
		vOff+vCount*14 > len(data) || sOff+4*sWords > len(data) {
		return nil, fmt.Errorf("bgv: mesh at 0x%X out of range", rec)
	}

	m := &Mesh{
		Offset: rec,
		ScaleX: f32(data, rec+0x30),
		ScaleY: f32(data, rec+0x34),
		ScaleZ: f32(data, rec+0x38),
	}

	// The strip list is a run of GE PRIM words, ending in RET. Vertices are
	// consumed sequentially: the GE's vertex pointer advances by whatever each
	// primitive reads, which is why one VADDR serves the whole run.
	next := 0
	for i := 0; i < sWords; i++ {
		w := le32(data, sOff+4*i)
		cmd := w >> 24
		if cmd == 0x0B { // RET — end of the fragment
			break
		}
		if cmd != 0x04 { // PRIM
			return nil, fmt.Errorf("bgv: mesh 0x%X: word %d is 0x%08X, not a PRIM", rec, i, w)
		}
		count := int(w & 0xFFFF)
		m.Strips = append(m.Strips, Strip{
			Type:  int((w >> 16) & 7),
			Start: next,
			Count: count,
		})
		next += count
	}
	if next != vCount {
		return nil, fmt.Errorf("bgv: mesh 0x%X: strips consume %d vertices, record says %d",
			rec, next, vCount)
	}

	m.Verts = make([]Vertex, vCount)
	for i := 0; i < vCount; i++ {
		m.Verts[i] = decodeVertex(data, vOff+i*14, m)
	}
	return m, nil
}

// decodeVertex reads one vertex of type 0x122 and returns it in model units.
// The fixed-point fields are fractional exactly as the GE reads them; the
// positions are then scaled by the record's model scale.
func decodeVertex(data []byte, p int, m *Mesh) Vertex {
	u := float32(binary.LittleEndian.Uint16(data[p:])) / 32768
	v := float32(binary.LittleEndian.Uint16(data[p+2:])) / 32768
	nx := float32(int8(data[p+4])) / 128
	ny := float32(int8(data[p+5])) / 128
	nz := float32(int8(data[p+6])) / 128
	// p+7 is padding: the 16-bit position is aligned to 2 bytes.
	x := float32(int16(binary.LittleEndian.Uint16(data[p+8:]))) / 32768
	y := float32(int16(binary.LittleEndian.Uint16(data[p+10:]))) / 32768
	z := float32(int16(binary.LittleEndian.Uint16(data[p+12:]))) / 32768
	return Vertex{
		X: x * m.ScaleX, Y: y * m.ScaleY, Z: z * m.ScaleZ,
		NX: nx, NY: ny, NZ: nz,
		U: u, V: v,
	}
}

func le32(b []byte, o int) uint32 { return binary.LittleEndian.Uint32(b[o:]) }

func f32(b []byte, o int) float32 { return math.Float32frombits(le32(b, o)) }

// Wheels are placed, not baked: the wheel mesh sits at the origin and the file
// carries four 4x4 matrices that put it at the corners. They live at a fixed
// offset — verified across the fleet, where they track each car's own geometry
// (the heavy class is wider at +-0.835, the supercar has a longer wheelbase) —
// and the right-hand pair carries a mirrored X basis.
//
// The engine's own behaviour confirms the binding: replaying the game and
// pairing every drawn primitive with the world matrix the engine gave it shows
// the body drawn once at the car's origin and the wheel mesh drawn at exactly
// these four corners.
const wheelMatrices = 0xB80

// Mat4 is a row-major 4x4 transform; the translation is the last row.
type Mat4 [16]float32

// Wheels returns the four wheel placements.
func Wheels(data []byte) ([]Mat4, error) {
	if wheelMatrices+4*64 > len(data) {
		return nil, fmt.Errorf("bgv: file too small for wheel placements")
	}
	out := make([]Mat4, 4)
	for w := 0; w < 4; w++ {
		for i := 0; i < 16; i++ {
			out[w][i] = f32(data, wheelMatrices+w*64+i*4)
		}
	}
	return out, nil
}

// Apply transforms a point by the matrix (row-major, translation in row 3).
func (m Mat4) Apply(x, y, z float32) (float32, float32, float32) {
	return m[0]*x + m[4]*y + m[8]*z + m[12],
		m[1]*x + m[5]*y + m[9]*z + m[13],
		m[2]*x + m[6]*y + m[10]*z + m[14]
}

// Mirrors reports whether the placement flips handedness — the wheels on one
// side of the car are instanced through a matrix with a negated basis. That
// matters to anyone re-winding triangles: a mirroring transform reverses a
// triangle's winding, so it cancels a Z-flip instead of compounding with it.
func (m Mat4) Mirrors() bool {
	det := m[0]*(m[5]*m[10]-m[9]*m[6]) -
		m[4]*(m[1]*m[10]-m[9]*m[2]) +
		m[8]*(m[1]*m[6]-m[5]*m[2])
	return det < 0
}

// ApplyDir transforms a direction (no translation), for normals.
func (m Mat4) ApplyDir(x, y, z float32) (float32, float32, float32) {
	return m[0]*x + m[4]*y + m[8]*z,
		m[1]*x + m[5]*y + m[9]*z,
		m[2]*x + m[6]*y + m[10]*z
}

// AtOrigin reports whether the mesh is modelled about the origin rather than in
// car space — which is what distinguishes the wheel, the one part the file
// expects to be instanced at the placements above. Every other mesh already
// carries its position in its vertices.
func (m *Mesh) AtOrigin() bool {
	if len(m.Verts) == 0 {
		return false
	}
	lo, hi := m.Verts[0], m.Verts[0]
	for _, v := range m.Verts {
		lo.X, lo.Y, lo.Z = min32(lo.X, v.X), min32(lo.Y, v.Y), min32(lo.Z, v.Z)
		hi.X, hi.Y, hi.Z = max32(hi.X, v.X), max32(hi.Y, v.Y), max32(hi.Z, v.Z)
	}
	cx, cy, cz := (lo.X+hi.X)/2, (lo.Y+hi.Y)/2, (lo.Z+hi.Z)/2
	sx, sy, sz := hi.X-lo.X, hi.Y-lo.Y, hi.Z-lo.Z
	centred := abs32(cx) < 0.1 && abs32(cy) < 0.1 && abs32(cz) < 0.1
	small := sx < 0.9 && sy < 0.9 && sz < 0.9
	return centred && small
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func abs32(a float32) float32 {
	if a < 0 {
		return -a
	}
	return a
}
