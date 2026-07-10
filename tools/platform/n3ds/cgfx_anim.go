package n3ds

import "fmt"

// cgfx_anim.go decodes a CGFX skeletal animation (CANM) into per-bone hermite
// curves over the nine transform components. The banner's animation is what
// makes the HOME-Menu scene move, so this is the file that gets Mario jumping.
//
// CANM layout (offsets from the magic word):
//
//	+0x08 name          +0x0C target group name ("SkeletalAnimation")
//	+0x10 loop mode     +0x14 frame count (f32)
//	+0x18 member count  +0x1C members DICT ptr
//
// Each member (one animated bone) is a transform record:
//
//	+0x00 flags — bits 16..24 mark transform slots *absent*; a clear bit means
//	      the matching slot pointer is live
//	+0x04/+0x08 name ptrs (member path and bone name)
//	+0x14 nine slot words: scaleXYZ, rotateXYZ, translateXYZ, each 0 or a
//	      self-relative pointer to a segment
//
// A segment is {startFrame f32, endFrame f32, flags, ptr → {count, curve ptrs}};
// a curve is {startFrame, endFrame, flags, keyCount, invDuration f32, keys}. The
// only key encoding seen in banners (curve flags 0x68) is the full-float hermite
// triple (frame, value, inSlope==outSlope), 12 bytes per key — anything else is
// rejected loudly rather than mis-decoded.
type AnimKey struct {
	Frame, Value, Slope float32
}

// AnimCurve is one component's keyframe list.
type AnimCurve struct {
	Keys []AnimKey
}

// Eval samples the curve at a frame with cubic-hermite interpolation, clamping
// outside the keyed range.
func (c *AnimCurve) Eval(frame float32) float32 {
	k := c.Keys
	if len(k) == 0 {
		return 0
	}
	if frame <= k[0].Frame {
		return k[0].Value
	}
	if frame >= k[len(k)-1].Frame {
		return k[len(k)-1].Value
	}
	i := 0
	for i+1 < len(k) && k[i+1].Frame <= frame {
		i++
	}
	a, b := k[i], k[i+1]
	h := b.Frame - a.Frame
	t := (frame - a.Frame) / h
	t2, t3 := t*t, t*t*t
	return (2*t3-3*t2+1)*a.Value + (t3-2*t2+t)*h*a.Slope +
		(-2*t3+3*t2)*b.Value + (t3-t2)*h*b.Slope
}

// Transform slot indices within BoneAnim.Curves.
const (
	SlotScaleX = iota
	SlotScaleY
	SlotScaleZ
	SlotRotX
	SlotRotY
	SlotRotZ
	SlotTransX
	SlotTransY
	SlotTransZ
)

// BoneAnim is one animated bone: a curve per transform slot (nil = not animated).
type BoneAnim struct {
	Bone   string
	Curves [9]*AnimCurve
}

// SkelAnim is a decoded CANM.
type SkelAnim struct {
	Name    string
	Frames  float32
	Loop    bool
	Members []BoneAnim
}

// DecodeSkeletalAnim decodes the CANM a SkeletalAnimations-dictionary entry
// points at. Unlike models and textures, an animation entry's data pointer
// addresses the magic word directly (no typeId word before it).
func (c *CGFX) DecodeSkeletalAnim(e CGFXEntry) (*SkelAnim, error) {
	M := e.Offset
	if string(c.raw[M:M+4]) != "CANM" {
		if string(c.raw[M+4:M+8]) == "CANM" { // tolerate a typeId-prefixed variant
			M += 4
		} else {
			return nil, fmt.Errorf("cgfx: animation entry %q is not a CANM", e.Name)
		}
	}
	an := &SkelAnim{
		Name:   e.Name,
		Loop:   c.u32(M+0x10) != 0,
		Frames: c.f32(M + 0x14),
	}
	dict := c.rel(M + 0x1C)
	entries, err := c.parseDict(dict)
	if err != nil {
		return nil, fmt.Errorf("cgfx: CANM members dict: %w", err)
	}
	for _, me := range entries {
		ba := BoneAnim{Bone: me.Name}
		flags := c.u32(me.Offset)
		for slot := 0; slot < 9; slot++ {
			if flags&(1<<(16+uint(slot))) != 0 {
				continue // marked absent
			}
			seg := c.rel(me.Offset + 0x14 + int64(slot)*4)
			if seg == 0 {
				continue
			}
			curve, err := c.decodeCurve(seg)
			if err != nil {
				return nil, fmt.Errorf("cgfx: CANM member %q slot %d: %w", me.Name, slot, err)
			}
			ba.Curves[slot] = curve
		}
		an.Members = append(an.Members, ba)
	}
	return an, nil
}

// decodeCurve follows a member's segment pointer down to its keyframes.
func (c *CGFX) decodeCurve(seg int64) (*AnimCurve, error) {
	list := c.rel(seg + 0xC)
	if list == 0 {
		return nil, fmt.Errorf("segment at 0x%x has no curve list", seg)
	}
	count := c.u32(list)
	if count != 1 {
		return nil, fmt.Errorf("segment at 0x%x has %d curves; expected 1", seg, count)
	}
	u := c.rel(list + 4)
	flags := c.u32(u + 8)
	if flags != 0x68 {
		return nil, fmt.Errorf("curve at 0x%x has key encoding 0x%x; only the 12-byte float-hermite form (0x68) is implemented", u, flags)
	}
	n := int(c.u32(u + 0xC))
	keys := u + 0x14
	if err := c.check(keys, int64(n)*12, "curve keys"); err != nil {
		return nil, err
	}
	curve := &AnimCurve{Keys: make([]AnimKey, n)}
	for i := 0; i < n; i++ {
		k := keys + int64(i)*12
		curve.Keys[i] = AnimKey{Frame: c.f32(k), Value: c.f32(k + 4), Slope: c.f32(k + 8)}
	}
	return curve, nil
}
