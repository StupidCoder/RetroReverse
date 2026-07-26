package lm

// key.go reads and evaluates the .key node animations of the opening demo
// (opwf_luigi.key and friends), as reverse-engineered from the game's
// evaluator at 0x8005AF0C and its two interpolators (0x8005AB04 for floats,
// 0x8005AC34 for angles).
//
// The file is host-endian big:
//
//	+0x00 u16 -        +0x02 u16 trackCount (one track per model node)
//	+0x04 u16, u16     (frame counts; the evaluator never reads them)
//	+0x08 u32 flags    (bit 1: tracks carry scale — built with the SRT matrix
//	                    builder instead of the RT one)
//	+0x0c u32 → float pool for scale channels
//	+0x10 u32 → s16 pool for rotation channels
//	+0x14 u32 → float pool for translation channels
//	+0x18 u32 → per-track 9 x u32 channel data offsets (into the pools)
//	+0x1c u32 → per-track 9 x u16 channel kinds
//
// The nine channels are sx,sy,sz, rx,ry,rz, tx,ty,tz. A kind of 0 leaves the
// channel at its default (1 for scale, 0 for the rest), 1 reads one constant
// from the pool at the channel's offset, and n>1 is n keyframes starting
// there. The kind halfword's top bit selects 4-field keys (time, value,
// tangent-in, tangent-out) over 3-field ones (time, value, tangent).
//
// Float keys are (f32 time, f32 value, f32 tangents); rotation keys are the
// same shape in s16s, with values and tangents in 360/4096-degree units.
// Interpolation is cubic hermite with times scaled by 1/30 (frames at 30 fps
// to seconds) and tangents per second — the game's exact arithmetic. A
// constant rotation is value<<4; an interpolated one comes out in
// 65536-per-turn units, so both meet at angle = units/65536 turns.

import (
	"encoding/binary"
	"fmt"
	"math"
)

// KeyChannel is one animated channel of a track.
type KeyChannel struct {
	Kind   uint16 // 0 default, 1 constant, else keyframe count; bit 15: 4-field keys
	Offset uint32
}

// KeyTrack animates one node.
type KeyTrack struct {
	Channels [9]KeyChannel // sx,sy,sz, rx,ry,rz, tx,ty,tz
}

// Key is a parsed .key animation.
type Key struct {
	Flags  uint32
	Tracks []KeyTrack
	scaleP []byte // the three pools, still raw
	rotP   []byte
	transP []byte
}

// ParseKey reads a .key file.
func ParseKey(b []byte) (*Key, error) {
	if len(b) < 0x20 {
		return nil, fmt.Errorf("lm: key file too short")
	}
	u16 := func(o uint32) uint16 { return binary.BigEndian.Uint16(b[o:]) }
	u32 := func(o uint32) uint32 { return binary.BigEndian.Uint32(b[o:]) }
	k := &Key{Flags: u32(0x08)}
	tracks := int(u16(0x02))
	scaleOff, rotOff, transOff := u32(0x0c), u32(0x10), u32(0x14)
	offsOff, kindsOff := u32(0x18), u32(0x1c)
	if int(kindsOff)+tracks*18 > len(b) || int(offsOff)+tracks*36 > len(b) {
		return nil, fmt.Errorf("lm: key tables out of range")
	}
	k.scaleP, k.rotP, k.transP = b[scaleOff:], b[rotOff:], b[transOff:]
	k.Tracks = make([]KeyTrack, tracks)
	for i := range k.Tracks {
		for c := 0; c < 9; c++ {
			k.Tracks[i].Channels[c] = KeyChannel{
				Kind:   u16(kindsOff + uint32(i*18+c*2)),
				Offset: u32(offsOff + uint32(i*36+c*4)),
			}
		}
	}
	return k, nil
}

const keyTimeScale = float32(1.0) / 30

// hermite is the game's cubic interpolation: keys of (time, value, tangent…),
// with the flagged form carrying separate in/out tangents. get(i, f) reads
// field f of key i as a float in final units; stride is 3 or 4.
func hermite(t float32, count int, four bool, get func(i, f int) float32) float32 {
	last := count - 1
	if t <= get(0, 0) {
		return get(0, 1)
	}
	if t >= get(last, 0) {
		return get(last, 1)
	}
	i := 0
	for i < last-1 && t >= get(i+1, 0) {
		i++
	}
	outT := 2
	if four {
		outT = 3
	}
	t0 := get(i, 0) * keyTimeScale
	v0 := get(i, 1)
	m0 := get(i, outT)
	t1 := get(i+1, 0) * keyTimeScale
	v1 := get(i+1, 1)
	m1 := get(i+1, 2)
	dt := t1 - t0
	u := t*keyTimeScale - t0
	u2 := u * u
	u3 := u2 * u
	h00 := 1 + 2*u3/(dt*dt*dt) - 3*u2/(dt*dt)
	h01 := 3*u2/(dt*dt) - 2*u3/(dt*dt*dt)
	h10 := u - 2*u2/dt + u3/(dt*dt)
	h11 := u3/(dt*dt) - u2/dt
	return v0*h00 + v1*h01 + m0*h10 + m1*h11
}

func (k *Key) floatChannel(ch KeyChannel, pool []byte, t, def float32) float32 {
	f32 := func(o uint32) float32 {
		return math.Float32frombits(binary.BigEndian.Uint32(pool[o:]))
	}
	count := int(ch.Kind & 0x7FFF)
	four := ch.Kind&0x8000 != 0
	switch {
	case ch.Kind == 0:
		return def
	case count == 1:
		return f32(ch.Offset * 4)
	default:
		stride := 3
		if four {
			stride = 4
		}
		return hermite(t, count, four, func(i, f int) float32 {
			return f32((ch.Offset + uint32(i*stride+f)) * 4)
		})
	}
}

// rotChannel returns the rotation in radians.
func (k *Key) rotChannel(ch KeyChannel, t float32) float32 {
	s16at := func(o uint32) float32 {
		return float32(int16(binary.BigEndian.Uint16(k.rotP[o:])))
	}
	const valScale = 360.0 / 4096 // s16 value units to degrees
	count := int(ch.Kind & 0x7FFF)
	four := ch.Kind&0x8000 != 0
	var deg float32
	switch {
	case ch.Kind == 0:
		return 0
	case count == 1:
		// value<<4 in 65536-per-turn units
		return s16at(ch.Offset*2) * 16 / 65536 * 2 * math.Pi
	default:
		stride := 3
		if four {
			stride = 4
		}
		deg = hermite(t, count, four, func(i, f int) float32 {
			v := s16at((ch.Offset + uint32(i*stride+f)) * 2)
			if f == 0 {
				return v // time in frames, scaled inside hermite
			}
			return v * valScale
		})
		return deg / 360 * 2 * math.Pi
	}
}

// Pose is one node's local transform at a point in time.
type Pose struct {
	Scale     [3]float32
	Rot       [3]float32 // radians, applied as Rz·Ry·Rx
	Translate [3]float32
}

// Eval evaluates track i at time t in frames (the 30 fps timeline; key times
// compare in frames, and only the hermite arithmetic rescales to seconds,
// exactly as the game's evaluator does).
func (k *Key) Eval(i int, t float32) Pose {
	tr := &k.Tracks[i]
	var p Pose
	for c := 0; c < 3; c++ {
		p.Scale[c] = k.floatChannel(tr.Channels[c], k.scaleP, t, 1)
		p.Rot[c] = k.rotChannel(tr.Channels[3+c], t)
		p.Translate[c] = k.floatChannel(tr.Channels[6+c], k.transP, t, 0)
	}
	return p
}

// Duration reports the animation's length in frames: the largest key time.
func (k *Key) Duration() float32 {
	var maxT float32
	for i := range k.Tracks {
		tr := &k.Tracks[i]
		for c := 0; c < 9; c++ {
			ch := tr.Channels[c]
			count := int(ch.Kind & 0x7FFF)
			if count < 2 {
				continue
			}
			stride := 3
			if ch.Kind&0x8000 != 0 {
				stride = 4
			}
			var t float32
			if c >= 3 && c < 6 {
				t = float32(int16(binary.BigEndian.Uint16(k.rotP[(ch.Offset+uint32((count-1)*stride))*2:])))
			} else {
				pool := k.scaleP
				if c >= 6 {
					pool = k.transP
				}
				t = math.Float32frombits(binary.BigEndian.Uint32(pool[(ch.Offset+uint32((count-1)*stride))*4:]))
			}
			if t > maxT {
				maxT = t
			}
		}
	}
	return maxT
}

// Quat converts the pose's euler angles to a quaternion (x,y,z,w), matching
// the game's Rz·Ry·Rx composition.
func (p Pose) Quat() [4]float32 {
	half := func(a float32) (float32, float32) {
		return float32(math.Sin(float64(a / 2))), float32(math.Cos(float64(a / 2)))
	}
	sx, cx := half(p.Rot[0])
	sy, cy := half(p.Rot[1])
	sz, cz := half(p.Rot[2])
	// q = qz * qy * qx
	return [4]float32{
		cz*cy*sx - sz*sy*cx,
		cz*sy*cx + sz*cy*sx,
		sz*cy*cx - cz*sy*sx,
		cz*cy*cx + sz*sy*sx,
	}
}
