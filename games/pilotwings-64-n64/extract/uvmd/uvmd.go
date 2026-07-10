// Package uvmd decodes Pilotwings 64's model resources.
//
// A UVMD resource is one GZIP'd COMM chunk. Its reader is at 0x802256B8 and
// pulls every field through the same cursor helper (0x80225394) the texture
// reader uses, so the file is a strictly linear stream:
//
//	u16 vertexCount
//	u8  lodCount        levels of detail; each holds the same parts
//	u8  partCount
//	u8  faceCount       count of the 36-byte records after the matrices
//	u8  flag            0 or 1
//	u16 boneCount       count of the 6-byte records at the end
//	Vertex[vertexCount]                16 bytes each: the Fast3D vertex pool
//	for each lod:
//	    u8 partCount, u8 flag
//	    for each part:
//	        u8 batchCount, u8 flag, u8 flag
//	        for each batch:
//	            u32 material                see Material
//	            u16 unknown, u16 unknown
//	            u16 commandCount
//	            command[commandCount]       see below
//	    u32 lodDistance?
//	Matrix[partCount]                  64 bytes: a 4x4 float rest pose
//	Face[faceCount]                    36 bytes
//	u32 x3
//	Bone[boneCount]                    6 bytes
//
// The whole resource is consumed but for 0..7 bytes of padding to an 8-byte
// boundary. Nothing here is inferred from the shape of the data: the field
// order is the order the reader calls its cursor in.
//
// # The command stream
//
// A batch is not a triangle list. It is a compact encoding the game expands
// into a Fast3D display list at load time (the emitter is at 0x80225940). Each
// command is a u16:
//
//	bit 0x4000 set   a triangle. Three 4-bit vertex-buffer slots in bits
//	                 8..11, 4..7 and 0..3. (The game multiplies each by 10,
//	                 the byte stride of a G_TRI1 index.)
//	bit 0x4000 clear a vertex load. Bits 0..13 are an index into the vertex
//	                 pool; a following u8 packs ((n-1) << 4) | slot, loading n
//	                 vertices into the 16-entry buffer starting at slot.
//
// So the mesh's connectivity lives in a 16-slot sliding window, exactly as the
// RSP sees it, and Build replays the same window to recover triangles as
// indices into the vertex pool.
package uvmd

import (
	"encoding/binary"
	"fmt"
	"math"
)

// HeaderSize is the fixed header ahead of the vertex pool.
const HeaderSize = 8

// VertexSize is the size of one Fast3D vertex record.
const VertexSize = 16

// NoTexture is the material texture field's "untextured" value: the batch is
// drawn with vertex colours (the PILOTWINGS letters are 1,464 such triangles).
const NoTexture = 0xFFF

// Vertex is one Fast3D vertex, as it ships.
type Vertex struct {
	X, Y, Z    int16
	Flag       uint16
	S, T       int16
	R, G, B, A uint8
}

// Material is a batch's u32 material word. Only the low 12 bits are pinned:
// they index the archive's UVTX resources in order (0..462), or are NoTexture.
// The upper 20 bits carry render flags the loader tests (bit 27 is latched into
// a per-part byte) and are kept raw rather than guessed at.
type Material uint32

// Texture returns the UVTX ordinal this batch samples, and whether it has one.
func (m Material) Texture() (int, bool) {
	t := int(m & 0xFFF)
	if t == NoTexture {
		return 0, false
	}
	return t, true
}

// Flags returns the material word's upper 20 bits.
func (m Material) Flags() uint32 { return uint32(m) >> 12 }

// Cmd is one decoded command of a batch's stream. It is kept alongside the
// triangles so the exact display list the game's emitter writes can be rebuilt
// from the ROM and compared with the one the engine built (see cmd/mdldump).
type Cmd struct {
	Tri bool

	// Triangle: the three vertex-buffer slots.
	I0, I1, I2 int

	// Vertex load: Count vertices from pool index First into buffer slot Slot.
	First, Count, Slot int
}

// Batch is one display list: a material and the triangles drawn with it,
// as indices into the model's vertex pool.
type Batch struct {
	Material Material
	Unknown  [2]uint16
	Tris     [][3]int
	Stream   []Cmd
}

// Part is one rigid piece of the model; its rest pose is Model.Matrices[i].
type Part struct {
	Batches []Batch
}

// LOD is one level of detail: the same parts at a different triangle budget.
type LOD struct {
	Parts    []Part
	Distance uint32
}

// Matrix is a part's 4x4 rest pose, row-vector (translation in row 3).
//
// There is one matrix per part of LOD 0, in part order — 264 of the archive's
// single-part models carry an identity matrix, which is what shows these blocks
// really are per-part poses rather than something else the loader happens to
// allocate alongside. A *lower* LOD may drop a part (model 83's LOD 1 has two
// parts against three matrices), so the pairing holds for LOD 0 only.
type Matrix [4][4]float32

// Model is a decoded UVMD resource.
type Model struct {
	Vertices []Vertex
	LODs     []LOD
	Matrices []Matrix

	// Faces and Bones are carried through undecoded: their consumers have not
	// been traced, so the bytes are kept rather than given a meaning.
	Faces [][36]byte
	Bones [][6]byte

	Trailer [3]uint32
	Padding int // bytes left after the parse; always < 8 (alignment)
}

// Triangles returns every triangle of a LOD, flattened across its parts.
func (m *Model) Triangles(lod int) int {
	n := 0
	for _, p := range m.LODs[lod].Parts {
		for _, b := range p.Batches {
			n += len(b.Tris)
		}
	}
	return n
}

// DecodeBatch reads one batch — a material and a command stream — from data at
// offset p, returning it and the offset just past it.
//
// It is exported because UVCT terrain chunks carry the *same* batch record and
// feed the *same* display-list emitter at 0x80225940. The two formats differ in
// everything around the batches and in nothing about them.
func DecodeBatch(data []byte, p, vertexCount int) (Batch, int, error) {
	c := &cursor{d: data, p: p}
	var b Batch
	b.Material = Material(c.u32())
	b.Unknown[0], b.Unknown[1] = c.u16(), c.u16()
	n := int(c.u16())
	if c.err != nil {
		return b, c.p, c.err
	}
	tris, stream, err := build(c, n, vertexCount)
	if err != nil {
		return b, c.p, err
	}
	b.Tris, b.Stream = tris, stream
	return b, c.p, c.err
}

// cursor mirrors the game's read helper: every field is pulled in order and the
// position advances by exactly the bytes consumed.
type cursor struct {
	d   []byte
	p   int
	err error
}

func (c *cursor) need(n int) bool {
	if c.err != nil {
		return false
	}
	if c.p+n > len(c.d) {
		c.err = fmt.Errorf("uvmd: read of %d bytes at %#x overruns the %d-byte chunk", n, c.p, len(c.d))
		return false
	}
	return true
}

func (c *cursor) u8() uint8 {
	if !c.need(1) {
		return 0
	}
	v := c.d[c.p]
	c.p++
	return v
}

func (c *cursor) u16() uint16 {
	if !c.need(2) {
		return 0
	}
	v := binary.BigEndian.Uint16(c.d[c.p:])
	c.p += 2
	return v
}

func (c *cursor) u32() uint32 {
	if !c.need(4) {
		return 0
	}
	v := binary.BigEndian.Uint32(c.d[c.p:])
	c.p += 4
	return v
}

func (c *cursor) skip(n int) {
	if c.need(n) {
		c.p += n
	}
}

// Decode parses one UVMD COMM chunk.
func Decode(data []byte) (*Model, error) {
	c := &cursor{d: data}
	vertexCount := int(c.u16())
	lodCount := int(c.u8())
	partCount := int(c.u8())
	faceCount := int(c.u8())
	_ = c.u8() // flag
	boneCount := int(c.u16())
	if c.err != nil {
		return nil, c.err
	}

	m := &Model{}
	if !c.need(vertexCount * VertexSize) {
		return nil, c.err
	}
	m.Vertices = make([]Vertex, vertexCount)
	for i := range m.Vertices {
		b := data[c.p+i*VertexSize:]
		m.Vertices[i] = Vertex{
			X: int16(binary.BigEndian.Uint16(b[0:])), Y: int16(binary.BigEndian.Uint16(b[2:])),
			Z: int16(binary.BigEndian.Uint16(b[4:])), Flag: binary.BigEndian.Uint16(b[6:]),
			S: int16(binary.BigEndian.Uint16(b[8:])), T: int16(binary.BigEndian.Uint16(b[10:])),
			R: b[12], G: b[13], B: b[14], A: b[15],
		}
	}
	c.skip(vertexCount * VertexSize)

	for i := 0; i < lodCount; i++ {
		lod := LOD{}
		parts := int(c.u8())
		_ = c.u8()
		for j := 0; j < parts; j++ {
			batches := int(c.u8())
			_, _ = c.u8(), c.u8()
			var part Part
			for k := 0; k < batches; k++ {
				var b Batch
				b.Material = Material(c.u32())
				b.Unknown[0], b.Unknown[1] = c.u16(), c.u16()
				n := int(c.u16())
				tris, stream, err := build(c, n, vertexCount)
				if err != nil {
					return nil, fmt.Errorf("uvmd: lod %d part %d batch %d: %w", i, j, k, err)
				}
				b.Tris, b.Stream = tris, stream
				part.Batches = append(part.Batches, b)
			}
			lod.Parts = append(lod.Parts, part)
		}
		lod.Distance = c.u32()
		m.LODs = append(m.LODs, lod)
		if c.err != nil {
			return nil, c.err
		}
	}

	m.Matrices = make([]Matrix, partCount)
	for i := range m.Matrices {
		if !c.need(64) {
			return nil, c.err
		}
		for r := 0; r < 4; r++ {
			for col := 0; col < 4; col++ {
				bits := binary.BigEndian.Uint32(data[c.p+(r*4+col)*4:])
				m.Matrices[i][r][col] = math.Float32frombits(bits)
			}
		}
		c.skip(64)
	}

	m.Faces = make([][36]byte, faceCount)
	for i := range m.Faces {
		if !c.need(36) {
			return nil, c.err
		}
		copy(m.Faces[i][:], data[c.p:])
		c.skip(36)
	}

	m.Trailer[0], m.Trailer[1], m.Trailer[2] = c.u32(), c.u32(), c.u32()

	m.Bones = make([][6]byte, boneCount)
	for i := range m.Bones {
		if !c.need(6) {
			return nil, c.err
		}
		copy(m.Bones[i][:], data[c.p:])
		c.skip(6)
	}
	if c.err != nil {
		return nil, c.err
	}

	m.Padding = len(data) - c.p
	if m.Padding < 0 || m.Padding >= 8 {
		return nil, fmt.Errorf("uvmd: %d bytes left after the parse (expected < 8 of alignment padding)", m.Padding)
	}
	return m, nil
}

// build replays the command stream through the RSP's 16-entry vertex buffer,
// returning triangles as indices into the vertex pool.
func build(c *cursor, n, vertexCount int) ([][3]int, []Cmd, error) {
	var slots [16]int
	for i := range slots {
		slots[i] = -1
	}
	var tris [][3]int
	stream := make([]Cmd, 0, n)
	for i := 0; i < n; i++ {
		cmd := c.u16()
		if c.err != nil {
			return nil, nil, c.err
		}
		if cmd&0x4000 != 0 {
			s0, s1, s2 := int(cmd>>8&0xF), int(cmd>>4&0xF), int(cmd&0xF)
			t := [3]int{slots[s0], slots[s1], slots[s2]}
			for _, v := range t {
				if v < 0 {
					return nil, nil, fmt.Errorf("triangle reads an unloaded vertex slot")
				}
			}
			tris = append(tris, t)
			stream = append(stream, Cmd{Tri: true, I0: s0, I1: s1, I2: s2})
			continue
		}
		first := int(cmd & 0x3FFF)
		b := c.u8()
		if c.err != nil {
			return nil, nil, c.err
		}
		count := int(b>>4) + 1
		slot := int(b & 0xF)
		if first+count > vertexCount {
			return nil, nil, fmt.Errorf("vertex load %d..%d exceeds the %d-vertex pool", first, first+count-1, vertexCount)
		}
		if slot+count > len(slots) {
			return nil, nil, fmt.Errorf("vertex load of %d into slot %d overruns the 16-entry buffer", count, slot)
		}
		for k := 0; k < count; k++ {
			slots[slot+k] = first + k
		}
		stream = append(stream, Cmd{First: first, Count: count, Slot: slot})
	}
	return tris, stream, nil
}
