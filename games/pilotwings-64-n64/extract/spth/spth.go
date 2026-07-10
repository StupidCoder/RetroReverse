// Package spth decodes Pilotwings 64's spline paths.
//
// The archive holds eight SPTH resources. Each is an uncompressed IFF FORM of
// seven equally-shaped chunks, one per animated channel:
//
//	SCPX, SCPY, SCPZ    position
//	SCPH, SCPP, SCPR    heading, pitch, roll
//	SCP#                an eighth channel whose meaning is untraced
//
// and each chunk is
//
//	u32 count
//	u32 unknown
//	Key[count]          two f32s
//
// The arithmetic pins the record: `8 + 8*count` equals the chunk's size for all
// 56 chunks in the archive, exactly.
//
// Which of a key's two floats is the value and which is the time does not follow
// from that, but two measurements settle it. In every curve with more than one
// key — 49 of them — the **second** component rises across all keys but the
// last, and the last one's second component is **exactly 0.0**: a terminator.
// And in six of the eight paths the SCPX, SCPY and SCPZ curves carry an
// identical array of second components, which is what a shared timeline looks
// like. (In the other two, X, Y and Z have different key counts outright, so
// they are keyed independently.) Key therefore names them Value and Time.
package spth

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Channels are the seven chunk tags a path carries, in no particular order.
var Channels = []string{"SCPX", "SCPY", "SCPZ", "SCPH", "SCPP", "SCPR", "SCP#"}

// Key is one control point.
type Key struct {
	Value float32
	Time  float32
}

// Curve is one channel.
type Curve struct {
	Unknown uint32
	Keys    []Key
}

// Path is a decoded SPTH resource.
type Path struct {
	Curves map[string]Curve
}

func be32(b []byte, o int) uint32 { return binary.BigEndian.Uint32(b[o:]) }
func f32(b []byte, o int) float32 { return math.Float32frombits(be32(b, o)) }

// DecodeCurve parses one SCPx chunk.
func DecodeCurve(data []byte) (Curve, error) {
	if len(data) < 8 {
		return Curve{}, fmt.Errorf("spth: curve is %d bytes", len(data))
	}
	n := int(be32(data, 0))
	if want := 8 + 8*n; want != len(data) {
		return Curve{}, fmt.Errorf("spth: count %d needs %d bytes, the chunk has %d", n, want, len(data))
	}
	c := Curve{Unknown: be32(data, 4), Keys: make([]Key, n)}
	for i := range c.Keys {
		c.Keys[i] = Key{Value: f32(data, 8+i*8), Time: f32(data, 12+i*8)}
	}
	return c, nil
}

// Decode parses a whole SPTH resource from its chunks.
func Decode(chunks map[string][]byte) (*Path, error) {
	p := &Path{Curves: map[string]Curve{}}
	for _, ch := range Channels {
		data, ok := chunks[ch]
		if !ok {
			return nil, fmt.Errorf("spth: path has no %s chunk", ch)
		}
		c, err := DecodeCurve(data)
		if err != nil {
			return nil, fmt.Errorf("spth: %s: %w", ch, err)
		}
		p.Curves[ch] = c
	}
	return p, nil
}

// Terminated reports whether every multi-key curve's time rises across all keys
// but the last, whose time is the 0.0 terminator. This is the observation that
// named Key's fields, so it is worth being able to re-check.
func (p *Path) Terminated() bool {
	for _, c := range p.Curves {
		if len(c.Keys) < 2 {
			continue
		}
		for i := 1; i < len(c.Keys)-1; i++ {
			if c.Keys[i].Time < c.Keys[i-1].Time {
				return false
			}
		}
		if c.Keys[len(c.Keys)-1].Time != 0 {
			return false
		}
	}
	return true
}

// SharedTimeline reports whether the position curves are keyed on one timeline.
func (p *Path) SharedTimeline() bool {
	x, y, z := p.Curves["SCPX"], p.Curves["SCPY"], p.Curves["SCPZ"]
	if len(x.Keys) != len(y.Keys) || len(x.Keys) != len(z.Keys) {
		return false
	}
	for i := range x.Keys {
		if x.Keys[i].Time != y.Keys[i].Time || x.Keys[i].Time != z.Keys[i].Time {
			return false
		}
	}
	return true
}
