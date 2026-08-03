package merc

// anim.go decodes the art group's skeleton and compressed joint animations.
//
// Layouts read from the engine (see jak-and-daxter-ps2.md Part V):
//
//   - joint (type 0x619CE4): {name, number, parent joint|#f, bind 4x4}.
//   - art-joint-anim: basic {+4 name, +8 joint count, +40 →control,
//     +44.. per-joint drawable stubs}.
//   - joint-anim-compressed-control: {u32 numFrames, u32 fixedQWC,
//     u32 frameQWC, →fixed block, numFrames × →frame block}.
//   - fixed block: 64-byte header — words +0..52 = per-joint 4-bit masks
//     (8 per word, low nibble first), +56 = joint count (excluding the two
//     matrix joints), +60 = matrix-joint animation flags (bits 0/1); then
//     stream offsets {+64, +68, +72} relative to +80: "big" (64-byte
//     matrices, 8-byte Q1.15 quaternions, 8-byte float translation pairs),
//     words, halfwords.
//   - frame block: stream offsets {+0, +4, +8} relative to +16.
//   - nibble bits: 1 = translation animated, 2 = quaternion animated,
//     4 = scale animated, 8 = translation stored as float32×3 (else
//     int16×3 in 4-unit steps). Quaternions are int16 Q1.15; scales int16
//     Q4.12. Non-animated components decode from the fixed block, animated
//     ones from each frame block (decompress-fixed/frame-data-to-
//     accumulator, 0x68AACC/0x68A46C, dispatch tables built by
//     make-joint-jump-tables 0x689644).

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Joint is one skeleton joint.
type Joint struct {
	Name   string
	Number int
	Parent int // index into the joint list, -1 for none
	Bind   [16]float32
}

// JointPose is one joint's decoded frame state. The two leading "matrix
// joints" (align/prejoint) deliver a full matrix instead.
type JointPose struct {
	Matrix *[16]float32
	Trans  [3]float32
	Quat   [4]float32 // x, y, z, w
	Scale  [3]float32
}

// JointAnim is a decoded art-joint-anim.
type JointAnim struct {
	Name      string
	NumJoints int // total, including the two matrix joints
	Frames    [][]JointPose
}

func gstr(obj []byte, p uint32) string {
	if int(p)+4 >= len(obj) {
		return ""
	}
	end := p + 4
	for int(end) < len(obj) && obj[end] != 0 {
		end++
	}
	return string(obj[p+4 : end])
}

// Joints scans a linked art group for joint basics (type word given by the
// symbol table; Jak's is 0x619CE4).
func Joints(obj []byte, jointType uint32) []Joint {
	var out []Joint
	offs := map[uint32]int{}
	for o := uint32(0); int(o)+84 <= len(obj); o += 4 {
		if binary.LittleEndian.Uint32(obj[o:]) != jointType {
			continue
		}
		b := o + 4
		j := Joint{
			Name:   gstr(obj, u32(obj, b)),
			Number: int(u32(obj, b+4)),
			Parent: -1,
		}
		for k := 0; k < 16; k++ {
			j.Bind[k] = f32at(obj[b+12:], k*4)
		}
		offs[b] = len(out)
		out = append(out, j)
		o += 76 // joints are 80 bytes apart; the loop adds the last 4
	}
	// resolve parents (a joint basic pointer or #f)
	for i := range out {
		var b uint32
		for off, idx := range offs {
			if idx == i {
				b = off
			}
		}
		p := u32(obj, b+8)
		if idx, ok := offs[p]; ok {
			out[i].Parent = idx
		}
	}
	return out
}

// animStream is one cursor set over a compressed block.
type animStream struct {
	big, word, half []byte
	bi, wi, hi      int
}

func (s *animStream) mat() (m [16]float32) {
	for k := 0; k < 16; k++ {
		m[k] = f32at(s.big[s.bi:], k*4)
	}
	s.bi += 64
	return
}

func (s *animStream) quat() (q [4]float32) {
	for k := 0; k < 4; k++ {
		v := int16(binary.LittleEndian.Uint16(s.big[s.bi+k*2:]))
		q[k] = float32(v) / 32768
	}
	s.bi += 8
	return
}

func (s *animStream) transSmall() (t [3]float32) {
	for k := 0; k < 2; k++ {
		v := int16(binary.LittleEndian.Uint16(s.word[s.wi+k*2:]))
		t[k] = float32(v) * 4
	}
	s.wi += 4
	v := int16(binary.LittleEndian.Uint16(s.half[s.hi:]))
	t[2] = float32(v) * 4
	s.hi += 2
	return
}

func (s *animStream) transBig() (t [3]float32) {
	t[0] = f32at(s.big[s.bi:], 0)
	t[1] = f32at(s.big[s.bi:], 4)
	s.bi += 8
	t[2] = f32at(s.word[s.wi:], 0)
	s.wi += 4
	return
}

func (s *animStream) scale() (sc [3]float32) {
	for k := 0; k < 2; k++ {
		v := int16(binary.LittleEndian.Uint16(s.word[s.wi+k*2:]))
		sc[k] = float32(v) / 4096
	}
	s.wi += 4
	v := int16(binary.LittleEndian.Uint16(s.half[s.hi:]))
	sc[2] = float32(v) / 4096
	s.hi += 2
	return
}

// DecodeJointAnim decodes the art-joint-anim whose basic offset is p.
func DecodeJointAnim(obj []byte, p uint32) (*JointAnim, error) {
	a := &JointAnim{
		Name:      gstr(obj, u32(obj, p+4)),
		NumJoints: int(u32(obj, p+8)),
	}
	ctrl := u32(obj, p+40)
	if int(ctrl)+16 > len(obj) {
		return nil, fmt.Errorf("anim %s: control out of range", a.Name)
	}
	numFrames := int(u32(obj, ctrl))
	fixedOff := u32(obj, ctrl+12)
	fixed := obj[fixedOff:]
	count := int(u32(obj, fixedOff+56))
	matFlags := u32(obj, fixedOff+60)
	if count != a.NumJoints-2 {
		return nil, fmt.Errorf("anim %s: hdr count %d vs %d joints", a.Name, count, a.NumJoints)
	}
	nibble := func(j int) int {
		w := binary.LittleEndian.Uint32(fixed[(j/8)*4:])
		return int(w>>(4*(j%8))) & 0xF
	}
	fixStreams := func() *animStream {
		base := fixedOff + 80
		return &animStream{
			big:  obj[base+u32(obj, fixedOff+64):],
			word: obj[base+u32(obj, fixedOff+68):],
			half: obj[base+u32(obj, fixedOff+72):],
		}
	}
	for f := 0; f < numFrames; f++ {
		fb := u32(obj, ctrl+16+uint32(f)*4)
		if int(fb)+16 > len(obj) {
			return nil, fmt.Errorf("anim %s: frame %d out of range", a.Name, f)
		}
		fs := fixStreams()
		rs := &animStream{
			big:  obj[fb+16+u32(obj, fb):],
			word: obj[fb+16+u32(obj, fb+4):],
			half: obj[fb+16+u32(obj, fb+8):],
		}
		poses := make([]JointPose, 0, a.NumJoints)
		for mj := 0; mj < 2; mj++ {
			var m [16]float32
			if matFlags>>(uint(mj))&1 != 0 {
				m = rs.mat()
			} else {
				m = fs.mat()
			}
			poses = append(poses, JointPose{Matrix: &m})
		}
		for j := 0; j < count; j++ {
			n := nibble(j)
			big := n&8 != 0
			var po JointPose
			// Per block, streams are consumed in a fixed order: float
			// translations sit BEFORE the quaternion in the big stream
			// (the handlers read quat at big+8), small translations
			// before scales in the word/half streams.
			read := func(s *animStream, trans, quat, scale bool) {
				if trans && big {
					po.Trans = s.transBig()
				}
				if quat {
					po.Quat = s.quat()
				}
				if trans && !big {
					po.Trans = s.transSmall()
				}
				if scale {
					po.Scale = s.scale()
				}
			}
			read(rs, n&1 != 0, n&2 != 0, n&4 != 0)
			read(fs, n&1 == 0, n&2 == 0, n&4 == 0)
			poses = append(poses, po)
		}
		a.Frames = append(a.Frames, poses)
	}
	return a, nil
}

// FindAnims scans a linked art group for art-joint-anim basics.
func FindAnims(obj []byte, animType uint32) []uint32 {
	var out []uint32
	for o := uint32(0); int(o)+44 <= len(obj); o += 4 {
		if binary.LittleEndian.Uint32(obj[o:]) == animType {
			out = append(out, o+4)
		}
	}
	return out
}

// QuatNorm returns the quaternion's length (unit for a valid decode).
func (p *JointPose) QuatNorm() float64 {
	q := p.Quat
	return math.Sqrt(float64(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]))
}
