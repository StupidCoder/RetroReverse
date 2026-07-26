package lm

// scd.go reads the opening demo's per-shot script database (.scd) and the
// per-cut base records (.sco), as reverse-engineered from the camera evaluator
// at 0x801156FC and the channel interpolators at 0x800F1C58 (float pool) and
// 0x800F1E1C (s16 pool).
//
// The .scd is the shot's channel database:
//
//	+0x00 u16 frameCount   +0x02 u16   +0x04 u16   +0x06 u16
//	+0x08 u32 → camera:  f32 aspect, then 10 channel descriptors —
//	            pos xyz, target xyz, roll, fov, near, far
//	+0x0c u32 → simple lights, 36 B each: {char name[16], desc rgb[3], u16 pad}
//	+0x10 u32 → keyed lights, 104 B each: {char name[16], u32, desc[14]}
//	+0x14 u32 → (unused in the opening shots)
//	+0x18 u32 → float pool     +0x1c u32 → s16 pool
//
// A channel descriptor is 6 bytes: {u16 count, u16 offset, u16 stride}. A
// stride of 1 is a constant at pool[offset]; otherwise count keys of stride
// words each — (time, value, tangent[, out-tangent when stride is 4]) — and
// evaluation is the same cubic hermite as the .key format: key times are
// frames on the 30 fps timeline, rescaled by 1/30 inside the arithmetic.
// Float channels read the float pool; light colours read the s16 pool.
//
// The .sco carries per-cut base values the channels add onto: a f32 at +0,
// then the camera cut record at +4 — {s16 cut, u16 flags, f32 base pos[3],
// f32 base target[3], f32 base roll/aspect/fov/near/far} — followed by the
// shots' light cut records ({s16 lightIdx, s16, s16 rgb[3]}, clamped 0..255
// after the channel adds). In the forest walk the camera bases are zero and
// the channels are absolute; the evaluation below reproduces the game's
// runtime camera struct (0x803A3820) to the last bit at every probed frame.

import (
	"encoding/binary"
	"fmt"
	"math"
)

// SCDChannel is one {count, offset, stride} descriptor.
type SCDChannel struct {
	Count, Offset, Stride uint16
}

// SCDLight is a named light track (RGB channels into the s16 pool).
type SCDLight struct {
	Name     string
	Channels []SCDChannel
}

// SCD is a parsed script database.
type SCD struct {
	FrameCount  int
	CamAspect   float32
	Camera      [10]SCDChannel // pos xyz, target xyz, roll, fov, near, far
	Lights      []SCDLight     // the 36-byte records: 3 colour channels
	KeyedLights []SCDLight     // the 104-byte records: 14 channels
	fpool       []byte
	spool       []byte
}

// ParseSCD reads a .scd file.
func ParseSCD(b []byte) (*SCD, error) {
	if len(b) < 0x20 {
		return nil, fmt.Errorf("lm: scd too short")
	}
	u16 := func(o uint32) uint16 { return binary.BigEndian.Uint16(b[o:]) }
	u32 := func(o uint32) uint32 { return binary.BigEndian.Uint32(b[o:]) }
	s := &SCD{FrameCount: int(u16(0))}
	cam, light, keyed := u32(0x08), u32(0x0C), u32(0x10)
	fpool, spool := u32(0x18), u32(0x1C)
	if cam+4+60 > uint32(len(b)) || fpool > uint32(len(b)) || spool > uint32(len(b)) {
		return nil, fmt.Errorf("lm: scd sections out of range")
	}
	s.fpool, s.spool = b[fpool:spool], b[spool:]
	s.CamAspect = math.Float32frombits(u32(cam))
	desc := func(o uint32) SCDChannel { return SCDChannel{u16(o), u16(o + 2), u16(o + 4)} }
	for i := range s.Camera {
		s.Camera[i] = desc(cam + 4 + uint32(i)*6)
	}
	name := func(o uint32) string {
		n := b[o : o+16]
		for i, c := range n {
			if c == 0 {
				return string(n[:i])
			}
		}
		return string(n)
	}
	for o := light; o+36 <= keyed; o += 36 {
		l := SCDLight{Name: name(o)}
		for i := 0; i < 3; i++ {
			l.Channels = append(l.Channels, desc(o+16+uint32(i)*6))
		}
		s.Lights = append(s.Lights, l)
	}
	for o := keyed; o+104 <= fpool; o += 104 {
		l := SCDLight{Name: name(o)}
		for i := 0; i < 14; i++ {
			l.Channels = append(l.Channels, desc(o+20+uint32(i)*6))
		}
		s.KeyedLights = append(s.KeyedLights, l)
	}
	return s, nil
}

// EvalFloat evaluates a channel against the float pool at frame t.
func (s *SCD) EvalFloat(c SCDChannel, t float32) float32 {
	f32 := func(idx uint32) float32 {
		return math.Float32frombits(binary.BigEndian.Uint32(s.fpool[idx*4:]))
	}
	if c.Stride == 1 {
		return f32(uint32(c.Offset))
	}
	count, stride := int(c.Count), int(c.Stride)
	four := stride == 4
	return hermite(t, count, four, func(i, f int) float32 {
		return f32(uint32(c.Offset) + uint32(i*stride+f))
	})
}

// SCDCamera is the evaluated camera state at one frame.
type SCDCamera struct {
	Pos, Target [3]float32
	Roll        float32 // degrees
	Fov         float32 // degrees
	Near, Far   float32
}

// EvalCamera evaluates the camera track at frame t, adding the .sco cut
// record's base values when one is supplied (sco may be nil). The cut record's
// base floats run pos xyz, target xyz, roll, aspect, fov, near, far — one more
// field than the channel list, because the aspect has a base but no channel
// (the .scd's own aspect constant is added instead).
func (s *SCD) EvalCamera(sco []byte, t float32) SCDCamera {
	var base [11]float32
	if len(sco) >= 0x38 {
		for i := 0; i < 11; i++ {
			base[i] = math.Float32frombits(binary.BigEndian.Uint32(sco[8+i*4:]))
		}
	}
	var v [10]float32
	for i := range v {
		b := base[i]
		if i >= 7 {
			b = base[i+1] // skip the aspect base slot
		}
		v[i] = s.EvalFloat(s.Camera[i], t) + b
	}
	return SCDCamera{
		Pos:    [3]float32{v[0], v[1], v[2]},
		Target: [3]float32{v[3], v[4], v[5]},
		Roll:   v[6],
		Fov:    v[7],
		Near:   v[8],
		Far:    v[9],
	}
}
