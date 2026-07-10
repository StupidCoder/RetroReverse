// Package uvct decodes Pilotwings 64's terrain chunks.
//
// A UVCT resource is one GZIP'd COMM chunk holding one cell of a world grid
// (see extract/uvtr). Its reader is at 0x80225FBC and, like every other reader
// in this game, pulls each field through the cursor helper at 0x80225394:
//
//	u16 vertexCount
//	u16 faceCount
//	u16 objectCount
//	u16 batchCount
//	Vertex[vertexCount]     16 bytes each — the Fast3D vertex pool
//	Face[faceCount]         8 bytes each  — three u16 vertex indices and a u16
//	Object[objectCount]     a placement: u8 poseCount, that many 64-byte 4x4
//	                        matrices, then u16, u32, u32, u32, u16, u16
//	Batch[batchCount]       a UVMD batch followed by a 24-byte trailer
//
// A batch's material and command stream are not merely similar to a model's —
// they are read by the same code path and expanded by the same display-list
// emitter, so extract/uvmd's DecodeBatch is reused verbatim. What UVCT adds is
// a trailer after the commands:
//
//	u16 faceStart, u16 faceCount    a slice of the chunk's Face array
//	u16, u16, u32, u32, u32, u32
//
// The loader turns faceStart into a pointer into the face array, so the faces
// belong to the batch that draws over them. Their shape — three vertex indices
// and a word — and that pairing say collision, but nothing in the render path
// reads them, so they are carried through undecoded.
package uvct

import (
	"encoding/binary"
	"fmt"
	"math"

	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
)

// HeaderSize is the fixed header ahead of the vertex pool.
const HeaderSize = 8

// Face is one 8-byte record: three vertex indices and a trailing word.
type Face struct {
	V    [3]uint16
	Word uint16
}

// Object is one object placed inside a terrain chunk. The record, in file order,
// is what the reader at 0x802260FC pulls through the cursor:
//
//	u8  poseCount
//	Matrix[poseCount]   64 bytes each, libultra fixed point (not float)
//	u16 Type            0..198; 183 of the 199 are used
//	f32 X, Y, Z         the object's position, in the chunk's local space
//	u16 Mask            never zero; a single bit on 1,106 of the 1,364 objects,
//	                    two on 190, and up to six on the rest
//	u16 -               always 0xFFFF: a slot the loader fills in at runtime
//
// Most objects carry one matrix. Twenty-one carry several — up to ten — which is
// where a moving object's poses live. What Type selects is not yet traced.
type Object struct {
	Poses   []uvmd.Matrix
	Type    uint16
	X, Y, Z float32
	Mask    uint16
	Runtime uint16
}

// Batch is a UVMD batch plus the trailer UVCT appends to it.
type Batch struct {
	uvmd.Batch

	FaceStart uint16 // first Face this batch owns
	FaceCount uint16
	Extra     [2]uint16
	Words     [4]uint32
}

// Chunk is a decoded UVCT resource.
type Chunk struct {
	Vertices []uvmd.Vertex
	Faces    []Face
	Objects  []Object
	Batches  []Batch

	Trailer [20]byte
	Padding int // bytes left after the parse; alignment only
}

// Triangles counts the chunk's drawn triangles.
func (c *Chunk) Triangles() int {
	n := 0
	for _, b := range c.Batches {
		n += len(b.Tris)
	}
	return n
}

func be16(b []byte, o int) uint16 { return binary.BigEndian.Uint16(b[o:]) }
func be32(b []byte, o int) uint32 { return binary.BigEndian.Uint32(b[o:]) }

// Decode parses one UVCT COMM chunk.
func Decode(data []byte) (*Chunk, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("uvct: chunk shorter than its header")
	}
	vertexCount := int(be16(data, 0))
	faceCount := int(be16(data, 2))
	objectCount := int(be16(data, 4))
	batchCount := int(be16(data, 6))
	p := HeaderSize

	need := func(n int) error {
		if p+n > len(data) {
			return fmt.Errorf("uvct: read of %d bytes at %#x overruns the %d-byte chunk", n, p, len(data))
		}
		return nil
	}

	c := &Chunk{}
	if err := need(vertexCount * uvmd.VertexSize); err != nil {
		return nil, err
	}
	c.Vertices = make([]uvmd.Vertex, vertexCount)
	for i := range c.Vertices {
		b := data[p+i*uvmd.VertexSize:]
		c.Vertices[i] = uvmd.Vertex{
			X: int16(be16(b, 0)), Y: int16(be16(b, 2)), Z: int16(be16(b, 4)), Flag: be16(b, 6),
			S: int16(be16(b, 8)), T: int16(be16(b, 10)),
			R: b[12], G: b[13], B: b[14], A: b[15],
		}
	}
	p += vertexCount * uvmd.VertexSize

	if err := need(faceCount * 8); err != nil {
		return nil, err
	}
	c.Faces = make([]Face, faceCount)
	for i := range c.Faces {
		b := data[p+i*8:]
		c.Faces[i] = Face{V: [3]uint16{be16(b, 0), be16(b, 2), be16(b, 4)}, Word: be16(b, 6)}
	}
	p += faceCount * 8

	c.Objects = make([]Object, objectCount)
	for i := range c.Objects {
		if err := need(1); err != nil {
			return nil, err
		}
		poses := int(data[p])
		p++
		if err := need(poses*64 + 18); err != nil {
			return nil, err
		}
		c.Objects[i].Poses = make([]uvmd.Matrix, poses)
		for k := 0; k < poses; k++ {
			c.Objects[i].Poses[k] = uvmd.DecodeFixedMatrix(data[p:])
			p += 64
		}
		o := &c.Objects[i]
		o.Type = be16(data, p)
		o.X = math.Float32frombits(be32(data, p+2))
		o.Y = math.Float32frombits(be32(data, p+6))
		o.Z = math.Float32frombits(be32(data, p+10))
		o.Mask = be16(data, p+14)
		o.Runtime = be16(data, p+16)
		p += 18
	}

	c.Batches = make([]Batch, batchCount)
	for i := range c.Batches {
		b, next, err := uvmd.DecodeBatch(data, p, vertexCount)
		if err != nil {
			return nil, fmt.Errorf("uvct: batch %d: %w", i, err)
		}
		p = next
		if err := need(24); err != nil {
			return nil, fmt.Errorf("uvct: batch %d trailer: %w", i, err)
		}
		c.Batches[i] = Batch{
			Batch:     b,
			FaceStart: be16(data, p), FaceCount: be16(data, p+2),
			Extra: [2]uint16{be16(data, p+4), be16(data, p+6)},
			Words: [4]uint32{be32(data, p+8), be32(data, p+12), be32(data, p+16), be32(data, p+20)},
		}
		p += 24
	}

	// The chunk closes with five u32s the loader reads into the tail of its
	// 44-byte descriptor (0x80226500..0x80226580), alongside pointers it has
	// already filled in. Their meaning is untraced, so the bytes are kept.
	if err := need(20); err != nil {
		return nil, fmt.Errorf("uvct: chunk trailer: %w", err)
	}
	copy(c.Trailer[:], data[p:])
	p += 20

	// Each batch's face slice must lie inside the face array, and every face's
	// vertex indices inside the vertex pool. A wrong trailer size would slide
	// these fields and break both.
	for i, b := range c.Batches {
		if int(b.FaceStart)+int(b.FaceCount) > len(c.Faces) {
			return nil, fmt.Errorf("uvct: batch %d owns faces %d..%d of %d",
				i, b.FaceStart, int(b.FaceStart)+int(b.FaceCount), len(c.Faces))
		}
	}
	for i, f := range c.Faces {
		for _, v := range f.V {
			if int(v) >= len(c.Vertices) {
				return nil, fmt.Errorf("uvct: face %d references vertex %d of %d", i, v, len(c.Vertices))
			}
		}
	}

	c.Padding = len(data) - p
	if c.Padding < 0 || c.Padding >= 8 {
		return nil, fmt.Errorf("uvct: %d bytes left after the parse (expected < 8 of alignment padding)", c.Padding)
	}
	return c, nil
}
