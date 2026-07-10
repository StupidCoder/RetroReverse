// Package upw decodes Pilotwings 64's missions (UPWT) and levels (UPWL).
//
// Both ship uncompressed, and both are plain IFF: a list of typed chunks. The
// mission reader is a switch over those tags in the gameplay overlay at
// 0x80345D40, and most arms do nothing but copy the chunk verbatim into a fixed
// global buffer (`load(global, chunkPtr, chunkSize)`). A feature of a mission —
// its rings, its balloons, its thermals — is therefore a flat array of records,
// and the tag names it.
//
// # UPWT: the 61 missions
//
//	NAME    the developers' own label, NUL-padded: "BIRD 3C", "CB L1 Target2"
//	JPTX    a short key, likewise: "A_EX_1"
//	INFO    a description: "shoot at the the target"
//	COMM    1072 bytes, the mission descriptor. Its first word decomposes as
//	        four bytes: class, vehicle, variant, level.
//	TPAD    48 bytes, exactly one per mission: the takeoff pad.
//	…       zero or more feature arrays, by tag (see FeatureStride).
//
// The vehicle byte agrees with the mission names throughout: every "GC" mission
// has vehicle 2, every "RP" 1, every "HG" 0, "CB" 3, "SD" 4, "HM" 5. The level
// byte is 0..3 and selects one of the four UPWL levels.
//
// # UPWL: the four levels
//
//	LEVL    8 bytes
//	LPAD    24-byte landing pads
//	WOBJ    16-byte world objects
//	TOYS, APTS, TPTS, BNUS, ESND   further arrays
//
// What is *not* claimed: the meaning of most feature records beyond their
// stride, and of LEVL's eight bytes. The chunk tags are the game's own, the
// strides are arithmetic, and everything past that would be a guess.
package upw

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// FeatureStride is the byte size of one record of each feature array, where the
// array's sizes across the archive pin it (the greatest common divisor of every
// occurrence, corroborated by tags that appear exactly once per mission).
//
// A tag absent from this table has records of unknown stride; its bytes are
// still carried, and Feature.Stride is zero.
var FeatureStride = map[string]int{
	"TPAD": 48, // takeoff pad — exactly one per mission
	"LPAD": 48, // landing pads (UPWT); UPWL's LPAD is 24
	"CNTG": 32,
	"HOPD": 32,
	"BTGT": 32,
	"TARG": 32,
	"LSTP": 40,
	"THER": 40,
	"BALS": 104,
	"HPAD": 128,
	"FALC": 344,
}

// Vehicle names the craft a mission is flown with. The numbering is the COMM
// descriptor's second byte, and it agrees with the developers' own mission-name
// prefixes on all 61 missions.
type Vehicle uint8

const (
	HangGlider Vehicle = 0
	RocketBelt Vehicle = 1
	Gyrocopter Vehicle = 2
	Cannonball Vehicle = 3
	SkyDiving  Vehicle = 4
	HumanTorch Vehicle = 5 // "HM" — named from its prefix only
)

func (v Vehicle) String() string {
	switch v {
	case HangGlider:
		return "hang glider"
	case RocketBelt:
		return "rocket belt"
	case Gyrocopter:
		return "gyrocopter"
	case Cannonball:
		return "cannonball"
	case SkyDiving:
		return "sky diving"
	case HumanTorch:
		return "HM"
	}
	return fmt.Sprintf("vehicle %d", uint8(v))
}

// Pad is a takeoff or landing pad: a position and a heading, then bytes the
// game fills in at runtime.
type Pad struct {
	X, Y, Z float32
	Heading float32
	Raw     []byte
}

// Feature is one of a mission's typed record arrays.
type Feature struct {
	Tag    string
	Stride int // 0 when unknown
	Data   []byte
}

// Count is the number of records, or -1 when the stride is unknown.
func (f Feature) Count() int {
	if f.Stride == 0 {
		return -1
	}
	return len(f.Data) / f.Stride
}

// Mission is one decoded UPWT resource.
type Mission struct {
	Name string
	JPTX string
	Info string

	Class   uint8 // COMM[0]: the lesson class (Exp / A / B / P)
	Vehicle Vehicle
	Variant uint8
	Level   uint8 // 0..3, an index into the four UPWL levels

	Takeoff  Pad
	Features []Feature

	Comm []byte // the whole 1072-byte descriptor, mostly untraced
}

// Level is one decoded UPWL resource.
type Level struct {
	Levl         [8]byte
	LandingPads  []Pad
	WorldObjects []WorldObject
	Chunks       []Feature // every chunk, verbatim, in file order
}

// WorldObject is a 16-byte UPWL WOBJ record: a position and a word.
type WorldObject struct {
	X, Y, Z float32
	Word    uint32
}

func be32(b []byte, o int) uint32 { return binary.BigEndian.Uint32(b[o:]) }
func f32(b []byte, o int) float32 { return math.Float32frombits(be32(b, o)) }
func str(b []byte) string         { return strings.TrimRight(string(b), "\x00 ") }

// Chunk is a (tag, bytes) pair as the archive reader hands it over.
type Chunk struct {
	Tag  string
	Data []byte
}

// DecodeMission parses one UPWT resource from its chunk list.
func DecodeMission(chunks []Chunk) (*Mission, error) {
	m := &Mission{}
	seen := map[string]bool{}
	for _, c := range chunks {
		if c.Tag == "PAD " {
			continue
		}
		seen[c.Tag] = true
		switch c.Tag {
		case "NAME":
			m.Name = str(c.Data)
		case "JPTX":
			m.JPTX = str(c.Data)
		case "INFO":
			m.Info = str(c.Data)
		case "COMM":
			if len(c.Data) < 4 {
				return nil, fmt.Errorf("upw: COMM is %d bytes", len(c.Data))
			}
			m.Comm = c.Data
			m.Class, m.Vehicle, m.Variant, m.Level = c.Data[0], Vehicle(c.Data[1]), c.Data[2], c.Data[3]
		case "TPAD":
			p, err := decodePad(c.Data, 48)
			if err != nil {
				return nil, fmt.Errorf("upw: TPAD: %w", err)
			}
			m.Takeoff = p[0]
			m.Features = append(m.Features, feature(c))
		default:
			m.Features = append(m.Features, feature(c))
		}
	}
	for _, req := range []string{"NAME", "INFO", "COMM", "TPAD"} {
		if !seen[req] {
			return nil, fmt.Errorf("upw: mission has no %s chunk", req)
		}
	}
	if m.Level > 3 {
		return nil, fmt.Errorf("upw: mission %q names level %d; there are four", m.Name, m.Level)
	}
	return m, nil
}

func feature(c Chunk) Feature {
	f := Feature{Tag: c.Tag, Stride: FeatureStride[c.Tag], Data: c.Data}
	if f.Stride != 0 && len(c.Data)%f.Stride != 0 {
		f.Stride = 0 // the stride does not divide this occurrence: do not claim it
	}
	return f
}

func decodePad(b []byte, stride int) ([]Pad, error) {
	if len(b)%stride != 0 {
		return nil, fmt.Errorf("%d bytes is not a multiple of %d", len(b), stride)
	}
	out := make([]Pad, len(b)/stride)
	for i := range out {
		r := b[i*stride:]
		out[i] = Pad{X: f32(r, 0), Y: f32(r, 4), Z: f32(r, 8), Heading: f32(r, 12), Raw: r[:stride]}
	}
	return out, nil
}

// DecodeLevel parses one UPWL resource from its chunk list.
func DecodeLevel(chunks []Chunk) (*Level, error) {
	l := &Level{}
	for _, c := range chunks {
		if c.Tag == "PAD " {
			continue
		}
		l.Chunks = append(l.Chunks, Feature{Tag: c.Tag, Data: c.Data})
		switch c.Tag {
		case "LEVL":
			if len(c.Data) != 8 {
				return nil, fmt.Errorf("upw: LEVL is %d bytes, want 8", len(c.Data))
			}
			copy(l.Levl[:], c.Data)
		case "LPAD":
			p, err := decodePad(c.Data, 24)
			if err != nil {
				return nil, fmt.Errorf("upw: UPWL LPAD: %w", err)
			}
			l.LandingPads = p
		case "WOBJ":
			if len(c.Data)%16 != 0 {
				return nil, fmt.Errorf("upw: WOBJ is %d bytes, not a multiple of 16", len(c.Data))
			}
			for i := 0; i < len(c.Data); i += 16 {
				l.WorldObjects = append(l.WorldObjects, WorldObject{
					X: f32(c.Data, i), Y: f32(c.Data, i+4), Z: f32(c.Data, i+8), Word: be32(c.Data, i+12),
				})
			}
		}
	}
	return l, nil
}
