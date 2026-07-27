package lm

// anm.go reads the furniture animation format (anm/chest_1.anm and friends
// in the room arcs, and the door swings in /Game/game_usa.szp): the .bin
// scene graph's own TRS driven through the same hermite channels as the
// .key format. Read out of the binder at 0x8001DE90 and the evaluator at
// 0x8001E04C:
//
//	+0   u8  version        +1  u8 loop flag        +4  u32 frame count
//	+8/+0xC/+0x10  u32 pool offsets (scale / rotation / translation floats)
//	+0x14 u32 descriptor table offset (0x18)
//
// The descriptor table is 54 bytes per graph node — 9 channels of
// {s16 count, s16 offset, s16 fourFlag} in the order sx sy sz rx ry rz
// tx ty tz; channels 0-2 index the first pool, 3-5 the second, 6-8 the
// third. count 1 is a constant at pool[offset]; otherwise count hermite
// keys of {time, value, tangent} (four floats with separate in/out
// tangents when the flag is set), times in 30 fps frames, rotations in
// degrees. The evaluated TRS *replaces* the node's own: an animation's
// rest clip (chest_0.anm) reproduces the .bin graph values exactly, which
// is also this decode's verification.

import (
	"encoding/binary"
	"fmt"
	"math"
)

// AnmChannel is one scalar channel: a constant, or hermite keys.
type AnmChannel struct {
	Const float32
	Keys  [][4]float32 // {time (frames), value, tanIn, tanOut}
	Four  bool
}

// Eval samples the channel at t (in 30 fps frames).
func (c *AnmChannel) Eval(t float32) float32 {
	if len(c.Keys) == 0 {
		return c.Const
	}
	// hermite asks for field 2 (in-tangent) or 3 (out-tangent, four-flag
	// files); three-float keys store the same tangent in both slots.
	return hermite(t, len(c.Keys), c.Four, func(i, f int) float32 {
		return c.Keys[i][f]
	})
}

// Anm is a parsed furniture animation: 9 channels per graph node.
type Anm struct {
	Loop   bool
	Frames int
	Nodes  [][9]AnmChannel // sx sy sz rx ry rz tx ty tz
}

// ParseAnm reads a .anm file.
func ParseAnm(b []byte) (*Anm, error) {
	if len(b) < 0x18 {
		return nil, fmt.Errorf("lm: anm too short")
	}
	u32 := func(o uint32) uint32 { return binary.BigEndian.Uint32(b[o:]) }
	f32 := func(o uint32) float32 { return math.Float32frombits(u32(o)) }
	a := &Anm{Loop: b[1] != 0, Frames: int(u32(4))}
	var pools [3]uint32
	for i := range pools {
		pools[i] = u32(8 + uint32(i)*4)
	}
	desc := u32(0x14)
	if desc >= pools[0] || pools[0] > uint32(len(b)) {
		return nil, fmt.Errorf("lm: anm sections out of order")
	}
	nodes := int(pools[0]-desc) / 54
	for n := 0; n < nodes; n++ {
		var chs [9]AnmChannel
		for c := 0; c < 9; c++ {
			o := desc + uint32(n*54+c*6)
			count := int(int16(binary.BigEndian.Uint16(b[o:])))
			off := uint32(int16(binary.BigEndian.Uint16(b[o+2:])))
			four := binary.BigEndian.Uint16(b[o+4:]) != 0
			pool := pools[c/3]
			if count <= 1 {
				if at := pool + off*4; at+4 <= uint32(len(b)) {
					chs[c].Const = f32(at)
				}
				continue
			}
			stride := uint32(3)
			if four {
				stride = 4
			}
			chs[c].Four = four
			for k := 0; k < count; k++ {
				at := pool + (off+uint32(k)*stride)*4
				if at+stride*4 > uint32(len(b)) {
					return nil, fmt.Errorf("lm: anm channel keys out of range")
				}
				key := [4]float32{f32(at), f32(at + 4), f32(at + 8), f32(at + 8)}
				if four {
					key[3] = f32(at + 12)
				}
				chs[c].Keys = append(chs[c].Keys, key)
			}
		}
		a.Nodes = append(a.Nodes, chs)
	}
	return a, nil
}
