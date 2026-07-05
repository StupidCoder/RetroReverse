// .bca — Super Mario 64 DS skeletal animation, traced from the engine's own
// player (Part IV §5 of the doc):
//
//   - $02045394 walks the model's bone hierarchy (the same relative
//     parent/sibling/child links as the .bmd) and, per bone, decodes a 0x24-byte
//     track set at [anim+$14] + boneIdx*$24 ($0204547C);
//   - a track is {u8 rate, u8 animated, u16 index}: three scale tracks (values
//     in the fx32 array at [anim+$08]), three rotation tracks (u16 array at
//     [anim+$0C]; a value is angle>>4, shifted <<4 into DS angle units exactly
//     like the .bmd bone rotations), three translation tracks (fx32 array at
//     [anim+$10]) — S, R, T, each x/y/z;
//   - the sampler ($020456A0 u16 / $020457F0 fx32): animated==0 → the single
//     value at index; rate==0 → one key per frame (data[index+frame]); else
//     keys every 2^rate frames, linearly interpolated, with the tail past the
//     last full key stored per-frame (data[(frame>>rate)+(frame-last)]);
//   - the header is {u16 numBones, u16 numFrames, u32 loop, u32 scaleOff,
//     u32 rotOff, u32 transOff, u32 trackSetsOff}, and the file ships inside
//     the same "LZ77"-magic LZ10 wrapper as the .bmd.
//
// The decoded per-frame local TRS feeds the same bone compose as the bind pose
// ($02045074: R = Rx·Ry·Rz row-vector order, then the parent chain).
package sm64ds

import (
	"fmt"
	"math"
	"os"

	"retroreverse.com/tools/nds"
)

// BCATrack samples one component over the animation.
type BCATrack struct {
	Animated bool
	Rate     uint8    // log2 frames per key
	Values   []uint32 // raw array slice view (fx32 or u16<<16-agnostic index base)
	idx      int
	u16s     []uint16 // set for rotation tracks
	fx32s    []int32  // set for scale/translation tracks
}

// BCA is a decoded skeletal animation.
type BCA struct {
	NumBones  int
	NumFrames int
	Loop      bool
	tracks    [][9]BCATrack // per bone: sx,sy,sz, rx,ry,rz, tx,ty,tz
}

// LoadBCA reads and decompresses a .bca file.
func LoadBCA(path string) (*BCA, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) > 4 && string(raw[:4]) == "LZ77" {
		raw = raw[4:]
	}
	return DecodeBCA(nds.Decompress(raw))
}

// DecodeBCA parses a decompressed .bca body.
func DecodeBCA(d []byte) (*BCA, error) {
	if len(d) < 0x18 {
		return nil, fmt.Errorf("bca: too short")
	}
	u32 := func(o int) int { return int(le.Uint32(d[o:])) }
	a := &BCA{
		NumBones:  int(le.Uint16(d[0:])),
		NumFrames: int(le.Uint16(d[2:])),
		Loop:      u32(4) != 0,
	}
	sOff, rOff, tOff, setOff := u32(8), u32(0xC), u32(0x10), u32(0x14)
	if a.NumBones <= 0 || a.NumFrames <= 0 || setOff+a.NumBones*0x24 > len(d) {
		return nil, fmt.Errorf("bca: implausible header (%d bones, %d frames)", a.NumBones, a.NumFrames)
	}
	a.tracks = make([][9]BCATrack, a.NumBones)
	for b := 0; b < a.NumBones; b++ {
		for c := 0; c < 9; c++ {
			o := setOff + b*0x24 + c*4
			t := BCATrack{
				Rate:     d[o],
				Animated: d[o+1] != 0,
				idx:      int(le.Uint16(d[o+2:])),
			}
			switch {
			case c < 3: // scale, fx32
				t.fx32s = fx32Slice(d, sOff)
			case c < 6: // rotation, u16 angle>>4
				t.u16s = u16Slice(d, rOff)
			default: // translation, fx32
				t.fx32s = fx32Slice(d, tOff)
			}
			a.tracks[b][c] = t
		}
	}
	return a, nil
}

func fx32Slice(d []byte, off int) []int32 {
	n := (len(d) - off) / 4
	if off <= 0 || n <= 0 {
		return nil
	}
	out := make([]int32, n)
	for i := range out {
		out[i] = int32(le.Uint32(d[off+i*4:]))
	}
	return out
}

func u16Slice(d []byte, off int) []uint16 {
	n := (len(d) - off) / 2
	if off <= 0 || n <= 0 {
		return nil
	}
	out := make([]uint16, n)
	for i := range out {
		out[i] = le.Uint16(d[off+i*2:])
	}
	return out
}

// sample returns the raw track value at a frame, per the engine's sampler:
// constant, per-frame, or 2^rate-keyed with linear interpolation and a
// per-frame tail.
func (t *BCATrack) sample(frame, numFrames int) float64 {
	get := func(i int) float64 {
		if t.u16s != nil {
			if i < 0 || i >= len(t.u16s) {
				return 0
			}
			// angle>>4 -> DS angle units -> radians (STRH truncation kept)
			return float64(uint16(t.u16s[i]<<4)) / 65536 * 2 * math.Pi
		}
		if i < 0 || i >= len(t.fx32s) {
			return 0
		}
		return float64(t.fx32s[i]) / 4096
	}
	if !t.Animated {
		return get(t.idx)
	}
	if t.Rate == 0 {
		return get(t.idx + frame)
	}
	r := int(t.Rate)
	last := ((numFrames - 1) >> r) << r
	if frame >= last {
		return get(t.idx + (frame >> r) + (frame - last))
	}
	i := frame >> r
	frac := float64(frame-(i<<r)) / float64(int(1)<<r)
	a, b := get(t.idx+i), get(t.idx+i+1)
	if t.u16s != nil { // interpolate angles on the short way around
		d := math.Mod(b-a+3*math.Pi, 2*math.Pi) - math.Pi
		return a + d*frac
	}
	return a + (b-a)*frac
}

// BoneTRS returns bone b's local {sx,sy,sz, rx,ry,rz (radians), tx,ty,tz} at a
// frame.
func (a *BCA) BoneTRS(b, frame int) [9]float64 {
	var out [9]float64
	if b < 0 || b >= len(a.tracks) {
		return out
	}
	for c := 0; c < 9; c++ {
		out[c] = a.tracks[b][c].sample(frame, a.NumFrames)
	}
	return out
}
