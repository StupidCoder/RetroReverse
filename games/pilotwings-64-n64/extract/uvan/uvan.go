// Package uvan decodes Pilotwings 64's animations.
//
// An animation is a set of rotation tracks: for each animated part of a target
// model, a list of keyframes, each a unit quaternion at a frame number. It
// drives the game's articulated actors — the pilots, the birds, the vehicles
// with moving surfaces — by rotating their model's parts about their rest poses.
//
// # What it animates
//
// A UVAN's target is a UVMD *ordinal*, in the header (word 4). Across all 115
// resources those ordinals are 211..351 — the high-numbered models, which are
// the characters and vehicles. **No UVAN targets a model in the 0..198 range**,
// and that range is exactly the object `type`s a terrain chunk places (package
// uvct). So the scenery an island is dressed with — the ferris wheel, the little
// boat — is not animated here: those are terrain objects, drawn from their static
// placement with level-of-detail and billboarding, and whatever motion the game
// gives them is engine code, not this data.
//
// # Layout
//
// A UVAN resource is one COMM chunk followed by one PART chunk per animated part
// (some PART chunks are GZIP-compressed in the archive; the reader sees them
// decompressed). The COMM chunk is six u32:
//
//	u32 flags      1 when the animation has moving parts, 0 for a static pose set
//	u32 frames     the highest frame number any track reaches (the length)
//	u32 -          usually 0; 9 and 12 in two resources, unexplained
//	u32 rate       usually 1; 5 in the six longest animations, unexplained
//	u32 model      the target UVMD ordinal
//	u32 -          always 0
//
// and each PART chunk is:
//
//	u32 count      keyframes in this track
//	u32 part       the target part index within the model (1-based; part 0, the
//	               model root, is never animated)
//	Key[count]     20 bytes each, then zero padding to an 8-byte boundary
//
// A key is a unit quaternion and a frame word:
//
//	f32 x, y, z, w   the rotation, |q| = 1 (5,930 of 5,930 across the archive)
//	u16 frame        0..frames; tracks share a frame timeline but need not key
//	                 every frame — a part that only moves at the ends keys 1 and N
//	u16 tangent      an interpolation parameter, carried but not yet identified
//
// What is claimed here is earned by structure and by the archive agreeing with
// itself: every quaternion is unit-norm, every part index lies inside its target
// model's part count, every frame lies within the header's length, and the target
// ordinals never collide with the terrain-object types. What the tangent word
// means, and the two odd header words, are left carried and unnamed.
package uvan

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Quat is a rotation quaternion, in the order the keys store it.
type Quat struct{ X, Y, Z, W float32 }

// Key is one keyframe of a track.
type Key struct {
	Rot     Quat
	Frame   uint16
	Tangent uint16 // interpolation parameter, not yet identified
}

// Track is one animated part's keyframes.
type Track struct {
	Part int // part index within the target model, 1-based
	Keys []Key
}

// Animation is a decoded UVAN resource.
type Animation struct {
	Flags  uint32
	Frames int // the length: the highest frame any track reaches
	Word2  uint32
	Rate   uint32
	Model  int // the target UVMD ordinal
	Tracks []Track
}

// Decode reads a UVAN from its COMM chunk and its PART chunks, in archive order.
func Decode(comm []byte, parts [][]byte) (*Animation, error) {
	if len(comm) < 24 {
		return nil, fmt.Errorf("uvan: COMM is %d bytes, want 24", len(comm))
	}
	a := &Animation{
		Flags:  be32(comm, 0),
		Frames: int(be32(comm, 4)),
		Word2:  be32(comm, 8),
		Rate:   be32(comm, 12),
		Model:  int(be32(comm, 16)),
	}
	if w := be32(comm, 20); w != 0 {
		return nil, fmt.Errorf("uvan: COMM word 5 is %d, want 0", w)
	}
	for i, p := range parts {
		if len(p) < 8 {
			return nil, fmt.Errorf("uvan: PART %d is %d bytes, want at least 8", i, len(p))
		}
		count := int(be32(p, 0))
		t := Track{Part: int(be32(p, 4))}
		want := 8 + 20*count
		if want > len(p) {
			return nil, fmt.Errorf("uvan: PART %d: %d keys overrun %d bytes", i, count, len(p))
		}
		for k := 0; k < count; k++ {
			o := 8 + 20*k
			key := Key{
				Rot: Quat{
					X: f32(p, o), Y: f32(p, o+4), Z: f32(p, o+8), W: f32(p, o+12),
				},
				Frame:   be16(p, o+16),
				Tangent: be16(p, o+18),
			}
			if key.Frame > uint16(a.Frames) {
				return nil, fmt.Errorf("uvan: PART %d key %d: frame %d past length %d", i, k, key.Frame, a.Frames)
			}
			t.Keys = append(t.Keys, key)
		}
		// The remainder is zero padding to an 8-byte boundary.
		for _, b := range p[want:] {
			if b != 0 {
				return nil, fmt.Errorf("uvan: PART %d: non-zero padding", i)
			}
		}
		a.Tracks = append(a.Tracks, t)
	}
	return a, nil
}

func be32(b []byte, o int) uint32 { return binary.BigEndian.Uint32(b[o:]) }
func be16(b []byte, o int) uint16 { return binary.BigEndian.Uint16(b[o:]) }
func f32(b []byte, o int) float32 { return math.Float32frombits(be32(b, o)) }
