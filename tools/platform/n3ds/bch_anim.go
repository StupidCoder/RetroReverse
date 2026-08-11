package n3ds

import "fmt"

// bch_anim.go decodes the skeletal animations of a BCH container — the clips
// the characters are played through. They are the same H3D animation model the
// banner's CGFX carries (cgfx_anim.go), so they reuse its types: nine transform
// slots per bone, scale XYZ / rotate XYZ / translate XYZ, each either absent, a
// constant, or a hermite curve.
//
// The animation object:
//
//	+0x00 u32 name offset        +0x04 u16 loop flag, u16 total curve count
//	+0x08 f32 last frame         +0x0C u32 element pointer table
//	+0x10 u32 element count      +0x14 u32 (unused)
//
// and an element (0x30 bytes) is a bone's whole transform:
//
//	+0x00 u32 bone name offset   +0x04 u32 element kind (0x00040000, the only one)
//	+0x08 u32 flags              +0x0C nine slots
//
// **The flags say what each slot is, and both halves matter.** Bit 16+s marks
// slot s *absent* — the bone keeps its bind value. Otherwise the slot is a
// constant, in which case the word is the value itself, or a curve, in which
// case it is a pointer. Which, is bit 6+s for the six scale and rotation slots
// and bit 7+s for the three translation ones: bit 12 is skipped, and reading
// straight through it makes every translation look like a curve.
//
// Two counts prove the whole reading, and neither can be satisfied by accident:
// the animation header's own curve count equals the number of slots this
// decode calls curves — over all 198 animations of the two characters — and
// each curve stores its ordinal within the animation, which comes out as
// exactly its position in that enumeration.
const (
	bchAnimName    = 0x00
	bchAnimFlags   = 0x04
	bchAnimFrames  = 0x08
	bchAnimTable   = 0x0C
	bchAnimCount   = 0x10
	bchAnimElStr   = 0x30
	bchAnimElKind  = 0x04
	bchAnimElFlags = 0x08
	bchAnimElSlots = 0x0C

	// The only element kind these characters use. Anything else is a structure
	// this decoder has not seen and must not guess at.
	bchAnimElBone = 0x00040000
)

// slotConstBit is the flags bit marking slot s as a stored constant rather than
// a curve. The scale and rotation slots run 6..11 and the translation slots
// 13..15 — bit 12 is not used.
func slotConstBit(s int) uint {
	if s >= SlotTransX {
		return uint(7 + s)
	}
	return uint(6 + s)
}

// A curve header, 0x24 bytes, then its keys:
//
//	+0x00 f32 start frame     +0x04 f32 end frame
//	+0x08 u32 ordinal<<16     +0x0C u16 key encoding, u16 key count
//	+0x10 f32 value scale     +0x14 f32 value offset
//	+0x18 f32 (1.0)           +0x1C f32 1/(end-start)
//	+0x20 u32 key pointer
const (
	bchCurveStart = 0x00
	bchCurveEnd   = 0x04
	bchCurveOrd   = 0x08
	bchCurveEnc   = 0x0C
	bchCurveScale = 0x10
	bchCurveOff   = 0x14
	bchCurveKeys  = 0x20
)

// Key encodings, in the high byte of the encoding word. The low byte is 2 for
// every hermite form and 0 for the stepped one.
//
// The widths are not guessed: the curve headers and key blocks tile the
// animation section, so the distance from a key block to whatever starts next
// divided by the key count *is* the key size, and it comes out constant per
// encoding — 96 bits for encoding 3 across 9,797 curves, 128 for encoding 0,
// 64 for encoding 1.
const (
	bchKeyHermite2   = 0x00 // 16 bytes: frame, value, in-slope, out-slope, all f32
	bchKeyQuantised  = 0x01 // 8 bytes: packed frame+value, then an f32 slope
	bchKeyHermite    = 0x03 // 12 bytes: frame, value, one slope, all f32
	bchKeyEncMask    = 0xFF00
	bchKeyEncShift   = 8
	bchKeyFrameBits  = 12 // the packed form's frame field
	bchKeyFrameMask  = 1<<bchKeyFrameBits - 1
	bchKeyInterpMask = 0x00FF
	bchKeyHermiteOp  = 0x02
)

// DecodeSkeletalAnim decodes one entry of the BCHSkeletalAnims group.
func (f *BCH) DecodeSkeletalAnim(e BCHEntry) (*SkelAnim, error) {
	M := e.Offset
	if !f.inRange(M, 0x18) {
		return nil, fmt.Errorf("bch: animation %q header runs outside the file", e.Name)
	}
	an := &SkelAnim{
		Name:   e.Name,
		Loop:   f.u32(M+bchAnimFlags)&0xFFFF != 0,
		Frames: f.f32(M + bchAnimFrames),
	}
	wantCurves := int(f.u32(M+bchAnimFlags) >> 16)
	tbl := f.main + int64(f.u32(M+bchAnimTable))
	n := int64(f.u32(M + bchAnimCount))
	if !f.inRange(tbl, n*4) {
		return nil, fmt.Errorf("bch: animation %q element table runs outside the file", e.Name)
	}

	curves := 0
	for i := int64(0); i < n; i++ {
		el := f.main + int64(f.u32(tbl+i*4))
		if !f.inRange(el, bchAnimElStr) {
			return nil, fmt.Errorf("bch: animation %q element %d runs outside the file", e.Name, i)
		}
		if k := f.u32(el + bchAnimElKind); k != bchAnimElBone {
			return nil, fmt.Errorf("bch: animation %q element %d has kind 0x%08X, not a bone", e.Name, i, k)
		}
		flags := f.u32(el + bchAnimElFlags)
		if rest := flags &^ uint32(0x01FF0000|0x0FC0|0xE000); rest != 0 {
			return nil, fmt.Errorf("bch: animation %q element %d has unmodelled flag bits 0x%08X", e.Name, i, rest)
		}
		ba := BoneAnim{Bone: readCStr(f.raw, f.str+int64(f.u32(el+bchAnimName)))}
		for s := 0; s < 9; s++ {
			word := el + bchAnimElSlots + int64(s)*4
			if flags>>(16+uint(s))&1 != 0 {
				continue // absent: the bone keeps its bind value
			}
			if flags>>slotConstBit(s)&1 != 0 {
				v := f.f32(word)
				ba.Curves[s] = &AnimCurve{Keys: []AnimKey{{Value: v}}}
				continue
			}
			p := f.u32(word)
			if p == 0 {
				continue
			}
			c, err := f.decodeAnimCurve(f.main+int64(p), curves)
			if err != nil {
				return nil, fmt.Errorf("bch: animation %q bone %q slot %d: %w", e.Name, ba.Bone, s, err)
			}
			ba.Curves[s] = c
			curves++
		}
		an.Members = append(an.Members, ba)
	}
	// The header's own count of the animation's curves. A flag bit read wrongly
	// turns a constant into a curve or the reverse, and this stops being equal.
	if curves != wantCurves {
		return nil, fmt.Errorf("bch: animation %q holds %d curves, its header says %d",
			e.Name, curves, wantCurves)
	}
	return an, nil
}

// decodeAnimCurve reads one curve's keys. ord is the curve's position in the
// animation, which the curve itself records — checking it costs nothing and
// catches a slot enumerated in the wrong order.
func (f *BCH) decodeAnimCurve(c int64, ord int) (*AnimCurve, error) {
	if !f.inRange(c, 0x24) {
		return nil, fmt.Errorf("curve header at 0x%x runs outside the file", c)
	}
	if got := int(f.u32(c+bchCurveOrd) >> 16); got != ord {
		return nil, fmt.Errorf("curve at 0x%x says it is number %d, and it is number %d", c, got, ord)
	}
	enc := f.u32(c + bchCurveEnc)
	kind, count := (enc&bchKeyEncMask)>>bchKeyEncShift, int64(enc>>16)
	keys := f.main + int64(f.u32(c+bchCurveKeys))
	scale, off := f.f32(c+bchCurveScale), f.f32(c+bchCurveOff)

	var width int64
	switch kind {
	case bchKeyHermite:
		width = 12
	case bchKeyHermite2:
		width = 16
	case bchKeyQuantised:
		width = 8
	default:
		return nil, fmt.Errorf("curve at 0x%x uses key encoding 0x%04X, which this decoder does not model", c, enc)
	}
	if enc&bchKeyInterpMask != bchKeyHermiteOp {
		return nil, fmt.Errorf("curve at 0x%x interpolates by mode 0x%02X, not hermite", c, enc&bchKeyInterpMask)
	}
	if !f.inRange(keys, count*width) {
		return nil, fmt.Errorf("curve at 0x%x: %d keys of %d bytes run outside the file", c, count, width)
	}

	out := &AnimCurve{Keys: make([]AnimKey, count)}
	for i := int64(0); i < count; i++ {
		k := keys + i*width
		switch kind {
		case bchKeyHermite:
			s := f.f32(k + 8)
			out.Keys[i] = AnimKey{Frame: f.f32(k), Value: f.f32(k + 4), Slope: s, OutSlope: s}
		case bchKeyHermite2:
			out.Keys[i] = AnimKey{Frame: f.f32(k), Value: f.f32(k + 4), Slope: f.f32(k + 8), OutSlope: f.f32(k + 12)}
		case bchKeyQuantised:
			// One word carries both: the frame in the low twelve bits and a
			// signed value above it, which the header's scale and offset put
			// back into world units. The reading is checked by the frames it
			// produces — for every curve stored this way the frames come out
			// ascending, inside the curve's range, and the last one lands
			// exactly on its end frame.
			w := f.u32(k)
			s := f.f32(k + 4)
			out.Keys[i] = AnimKey{
				Frame:    float32(w & bchKeyFrameMask),
				Value:    float32(int32(w)>>bchKeyFrameBits)*scale + off,
				Slope:    s,
				OutSlope: s,
			}
		}
	}
	return out, nil
}

// PoseAt evaluates an animation at a frame, returning each bone's local
// transform. Bones the animation says nothing about keep the skeleton's own.
//
// The result is in the skeleton's own terms — a translate, an XYZ Euler triple
// in radians and a scale, composed T·(Rz·Ry·Rx)·S, exactly as the bind pose is.
func (a *SkelAnim) PoseAt(bones []BCHBone, frame float32) []BCHBone {
	out := append([]BCHBone(nil), bones...)
	byName := map[string]int{}
	for i, b := range bones {
		byName[b.Name] = i
	}
	for _, m := range a.Members {
		i, ok := byName[m.Bone]
		if !ok {
			continue
		}
		b := &out[i]
		for s, c := range m.Curves {
			if c == nil {
				continue
			}
			v := c.Eval(frame)
			switch {
			case s < SlotRotX:
				b.Scale[s] = v
			case s < SlotTransX:
				b.Rotate[s-SlotRotX] = v
			default:
				b.Trans[s-SlotTransX] = v
			}
		}
	}
	return out
}
