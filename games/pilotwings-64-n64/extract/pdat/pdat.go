// Package pdat decodes Pilotwings 64's recorded flight data.
//
// The archive holds 25 PDAT resources, uncompressed, and they are the largest
// non-audio thing in it (841 KB). Their chunk tags say what they are:
//
//	PHDR    8 bytes, once per resource
//	PPOS    24 bytes, 7,910 across the archive — a recorded position sample:
//	        three floats of position and three more of orientation
//	RHDR    24 bytes, 24 across the archive
//	RPKT    16 bytes, 24,401 across the archive
//
// PDAT was assumed to be object placement on its size alone, before the chunk
// names were read. It is not: a resource is a sequence of position samples and
// packets, which is what a recorded flight looks like. `RPKT` is carried through
// raw — nothing here has been traced to a consumer, and the record's shape is
// all that is claimed.
package pdat

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Sample is one 24-byte PPOS record.
type Sample struct {
	X, Y, Z float32
	A, B, C float32
}

// Recording is a decoded PDAT resource.
type Recording struct {
	Header  [8]byte
	Samples []Sample
	RHdrs   [][24]byte
	Packets [][16]byte
}

func be32(b []byte, o int) uint32 { return binary.BigEndian.Uint32(b[o:]) }
func f32(b []byte, o int) float32 { return math.Float32frombits(be32(b, o)) }

// Chunk is a (tag, bytes) pair from the archive reader.
type Chunk struct {
	Tag  string
	Data []byte
}

// Decode assembles one PDAT resource from its chunks, in file order.
func Decode(chunks []Chunk) (*Recording, error) {
	r := &Recording{}
	seen := false
	for _, c := range chunks {
		switch c.Tag {
		case "PAD ":
		case "PHDR":
			if len(c.Data) != 8 {
				return nil, fmt.Errorf("pdat: PHDR is %d bytes, want 8", len(c.Data))
			}
			copy(r.Header[:], c.Data)
			seen = true
		case "PPOS":
			if len(c.Data) != 24 {
				return nil, fmt.Errorf("pdat: PPOS is %d bytes, want 24", len(c.Data))
			}
			r.Samples = append(r.Samples, Sample{
				X: f32(c.Data, 0), Y: f32(c.Data, 4), Z: f32(c.Data, 8),
				A: f32(c.Data, 12), B: f32(c.Data, 16), C: f32(c.Data, 20),
			})
		case "RHDR":
			if len(c.Data) != 24 {
				return nil, fmt.Errorf("pdat: RHDR is %d bytes, want 24", len(c.Data))
			}
			var b [24]byte
			copy(b[:], c.Data)
			r.RHdrs = append(r.RHdrs, b)
		case "RPKT":
			if len(c.Data) != 16 {
				return nil, fmt.Errorf("pdat: RPKT is %d bytes, want 16", len(c.Data))
			}
			var b [16]byte
			copy(b[:], c.Data)
			r.Packets = append(r.Packets, b)
		default:
			return nil, fmt.Errorf("pdat: unexpected chunk %q", c.Tag)
		}
	}
	if !seen {
		return nil, fmt.Errorf("pdat: resource has no PHDR")
	}
	return r, nil
}
