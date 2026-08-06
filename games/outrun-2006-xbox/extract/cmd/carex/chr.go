// chr.go — the start-line flagman: /Chr/obj_chr_aut04_pmt.sz posed by the
// OZI skeleton and the ORT_OZI_OZI_* motion clips.
//
// The character pipeline's files (all offsets verified against the EUR disc):
//
//   - /Common/bone.bin — a nine-skeleton library (JEN MAN ONN OTK OZI RIC
//     SAR USI WMN; the developers' own names). 16-byte directory entries
//     {1, boneCount u16, FFFF, 0, recordsOff u32, nameOff u32}; bone record
//     0x38 bytes: {nameOff u32, kind u16 (0 plain / 2 chain root / 3 IK
//     joint / 4 effector), chainLen u16, FFFF, 0, then eight floats
//     {tx, ty, rx, ry, rz, sx, sy, sz} (rest local translation is 2-D:
//     along-bone and lateral; rest rotation is XYZ euler radians; scale is
//     always 1), childCount u32, childListOff u32 (u16 indices)}.
//     The mapping of the float slots was pinned by the motion files: clip
//     constants reproduce f[2..4] exactly (kao_chn rx ≈ -π, mune_chn ry =
//     1.571, ude_chn rz = -1.571).
//
//   - /Chr/CHR_AUT04.bin — the flagman's character descriptor. Header
//     {skeletonIdx u16 (4 = OZI), nAttach u16, nLOD u16, nMat u16} + u32
//     offsets {attachOff, lodOff, 0x38, 0, matOff, extraOff}. Attach records
//     (8 B) {boneIdx u16, 0, partId u16, 0xC3} bind one rigid full-detail
//     part to one bone; LOD records (20 B) {0, nBones u16, boneIdx x4 (-1
//     pad), 2*nBones-3, 0, partId u16, 0xC3} are the merged distance parts,
//     skinned across their listed bones. The matrices (0x40 B, row-vector
//     p·M+t) are INVERSE BINDS for the union of LOD-referenced bones,
//     ordered alphabetically by bone base name; their implied bind origins
//     read a Y-up ~1.7 m T-pose (ankles 0.14, knees 0.58, hips 0.97, waist
//     1.16, chest 1.35, shoulders ±0.22 @ 1.47, elbows ±0.49, hands ±0.76,
//     head 1.60).
//
//   - /Anims/mot_ETC_bin.sz, mot_OR2SP_ETC_bin.sz — the OZI clips; clip
//     names come from /Common/motdata_table.bin ({nameOff, listOff, count}
//     records mapping each mot_*.gz to its ordered clip-name list):
//     STAND_LP / STAND_SP_LP (idle), HATAFURI_00 / HATA_SP (the flag wave),
//     RUNAWAY, RADIO, TAISOU, DANCE... Motion container: 0x18-byte directory
//     {id, characterSlot, frames, descBytes, dataOff, sizeWords}; clip data
//     at dataOff+4: descBytes of {boneIdx u8, mask u8} descriptors (mask
//     bits 0-5 = rx ry rz tx ty tz), then per enabled channel a key list of
//     {frame u8, flags u8}: flag high bits give the payload (0x00 none —
//     value is zero, 0x20 value f16 + end of list, 0x40 value, 0x80 value +
//     slope, 0xC0 value + two slopes), low 5 bits extend the frame to 13
//     bits; a list also ends at frame == frames-1. Values are IEEE half
//     floats: radians, or metres in the character's Y-up ground space for
//     the effector target channels.
//
// The rig is Sega AM2's IK skeleton: *_chn bones root chains of chainLen
// segments ending in *_eff effectors. Clips animate the plain bones and the
// chain roots with rotations, and the effectors with target positions; the
// *_jnt bones between (which carry the limb geometry) are posed by IK: a
// two-segment law-of-cosines solve for arms and legs (segment lengths from
// the bind origins), a look-at for the one-segment spine and head chains.
// The pose conventions (euler order XYZ, local = R·T) were fitted against a
// live capture: bootoracle -carvtx over the model's VB range reads each
// part draw's combined c160-163 matrix (S = VP·M column-vector); relative
// part transforms inv(S_j)·S_i cancel VP and the fit drives the predicted
// FK/IK pose onto them (-chrfit).
package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/build"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/xbox"
)

// ---------- small mat3/mat4 helpers (column-vector: p' = M·p) ----------

type mat4 [4][4]float32

func mIdent() mat4 {
	return mat4{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}}
}

func mMul(a, b mat4) mat4 {
	var r mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var s float32
			for k := 0; k < 4; k++ {
				s += a[i][k] * b[k][j]
			}
			r[i][j] = s
		}
	}
	return r
}

func mTrans(x, y, z float32) mat4 {
	m := mIdent()
	m[0][3], m[1][3], m[2][3] = x, y, z
	return m
}

func mRotX(a float32) mat4 {
	s, c := float32(math.Sin(float64(a))), float32(math.Cos(float64(a)))
	return mat4{{1, 0, 0, 0}, {0, c, -s, 0}, {0, s, c, 0}, {0, 0, 0, 1}}
}

func mRotY(a float32) mat4 {
	s, c := float32(math.Sin(float64(a))), float32(math.Cos(float64(a)))
	return mat4{{c, 0, s, 0}, {0, 1, 0, 0}, {-s, 0, c, 0}, {0, 0, 0, 1}}
}

func mRotZ(a float32) mat4 {
	s, c := float32(math.Sin(float64(a))), float32(math.Cos(float64(a)))
	return mat4{{c, -s, 0, 0}, {s, c, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}}
}

// mEuler composes single-axis rotations in one of the six orders; order
// names the LEFT-to-RIGHT multiplication (XYZ = Rx·Ry·Rz).
func mEuler(rx, ry, rz float32, order int) mat4 {
	x, y, z := mRotX(rx), mRotY(ry), mRotZ(rz)
	switch order {
	case 0:
		return mMul(x, mMul(y, z)) // XYZ
	case 1:
		return mMul(x, mMul(z, y)) // XZY
	case 2:
		return mMul(y, mMul(x, z)) // YXZ
	case 3:
		return mMul(y, mMul(z, x)) // YZX
	case 4:
		return mMul(z, mMul(x, y)) // ZXY
	case 5:
		return mMul(z, mMul(y, x)) // ZYX
	}
	return mIdent()
}

func mInv(m mat4) mat4 {
	// rigid inverse (orthonormal rotation + translation)
	var r mat4
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			r[i][j] = m[j][i]
		}
	}
	for i := 0; i < 3; i++ {
		var s float32
		for j := 0; j < 3; j++ {
			s -= r[i][j] * m[j][3]
		}
		r[i][3] = s
	}
	r[3] = [4]float32{0, 0, 0, 1}
	return r
}

func mPos(m mat4) [3]float32 { return [3]float32{m[0][3], m[1][3], m[2][3]} }

// mInvFull is a general 4x4 inverse (Gauss-Jordan) for the projective
// capture matrices; mInv stays for rigid transforms.
func mInvFull(m mat4) mat4 {
	var a [4][8]float64
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			a[i][j] = float64(m[i][j])
		}
		a[i][4+i] = 1
	}
	for c := 0; c < 4; c++ {
		p := c
		for r := c + 1; r < 4; r++ {
			if math.Abs(a[r][c]) > math.Abs(a[p][c]) {
				p = r
			}
		}
		a[c], a[p] = a[p], a[c]
		d := a[c][c]
		if d == 0 {
			return mIdent()
		}
		for j := 0; j < 8; j++ {
			a[c][j] /= d
		}
		for r := 0; r < 4; r++ {
			if r == c {
				continue
			}
			f := a[r][c]
			for j := 0; j < 8; j++ {
				a[r][j] -= f * a[c][j]
			}
		}
	}
	var out mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			out[i][j] = float32(a[i][4+j])
		}
	}
	return out
}

func v3sub(a, b [3]float32) [3]float32 { return [3]float32{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func v3add(a, b [3]float32) [3]float32 { return [3]float32{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }
func v3scale(a [3]float32, s float32) [3]float32 {
	return [3]float32{a[0] * s, a[1] * s, a[2] * s}
}
func v3dot(a, b [3]float32) float32 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func v3cross(a, b [3]float32) [3]float32 {
	return [3]float32{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}
func v3len(a [3]float32) float32 { return float32(math.Sqrt(float64(v3dot(a, a)))) }
func v3norm(a [3]float32) [3]float32 {
	l := v3len(a)
	if l == 0 {
		return a
	}
	return v3scale(a, 1/l)
}

// mQuat extracts a unit quaternion (x,y,z,w) from the rotation part.
func mQuat(m mat4) [4]float32 {
	t := float64(m[0][0] + m[1][1] + m[2][2])
	var q [4]float32
	if t > 0 {
		s := math.Sqrt(t+1) * 2
		q[3] = float32(s / 4)
		q[0] = float32(float64(m[2][1]-m[1][2]) / s)
		q[1] = float32(float64(m[0][2]-m[2][0]) / s)
		q[2] = float32(float64(m[1][0]-m[0][1]) / s)
	} else if m[0][0] > m[1][1] && m[0][0] > m[2][2] {
		s := math.Sqrt(float64(1+m[0][0]-m[1][1]-m[2][2])) * 2
		q[0] = float32(s / 4)
		q[3] = float32(float64(m[2][1]-m[1][2]) / s)
		q[1] = float32(float64(m[0][1]+m[1][0]) / s)
		q[2] = float32(float64(m[0][2]+m[2][0]) / s)
	} else if m[1][1] > m[2][2] {
		s := math.Sqrt(float64(1+m[1][1]-m[0][0]-m[2][2])) * 2
		q[1] = float32(s / 4)
		q[3] = float32(float64(m[0][2]-m[2][0]) / s)
		q[0] = float32(float64(m[0][1]+m[1][0]) / s)
		q[2] = float32(float64(m[1][2]+m[2][1]) / s)
	} else {
		s := math.Sqrt(float64(1+m[2][2]-m[0][0]-m[1][1])) * 2
		q[2] = float32(s / 4)
		q[3] = float32(float64(m[1][0]-m[0][1]) / s)
		q[0] = float32(float64(m[0][2]+m[2][0]) / s)
		q[1] = float32(float64(m[1][2]+m[2][1]) / s)
	}
	// normalise
	l := float32(math.Sqrt(float64(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])))
	for i := range q {
		q[i] /= l
	}
	return q
}

// ---------- bone.bin ----------

type chrBone struct {
	name     string
	parent   int
	kind     int // 0 plain, 2 chain root, 3 IK joint, 4 effector
	chainLen int
	restT    [2]float32 // f[0], f[1]
	restR    [3]float32 // f[2], f[3], f[4]
	children []int
}

type chrSkeleton struct {
	name  string
	bones []chrBone
}

func parseBoneLib(b []byte) []chrSkeleton {
	var sks []chrSkeleton
	for o := 0; o+16 <= len(b); o += 16 {
		if u16(b, o) != 1 {
			break
		}
		cnt := int(u16(b, o+2))
		recOff := int(u32(b, o+8))
		nameOff := int(u32(b, o+12))
		sk := chrSkeleton{name: cstrAt(b, nameOff)}
		for i := 0; i < cnt; i++ {
			r := recOff + i*0x38
			br := chrBone{
				name: cstrAt(b, int(u32(b, r))),
				kind: int(u16(b, r+4)), chainLen: int(u16(b, r+6)),
				parent: -1,
			}
			br.restT = [2]float32{f32(b, r+0x10), f32(b, r+0x14)}
			br.restR = [3]float32{f32(b, r+0x18), f32(b, r+0x1C), f32(b, r+0x20)}
			nChild := int(u32(b, r+0x30))
			childOff := int(u32(b, r+0x34))
			for c := 0; c < nChild; c++ {
				br.children = append(br.children, int(u16(b, childOff+2*c)))
			}
			sk.bones = append(sk.bones, br)
		}
		for i, br := range sk.bones {
			for _, c := range br.children {
				if c < len(sk.bones) {
					sk.bones[c].parent = i
				}
			}
		}
		sks = append(sks, sk)
	}
	return sks
}

func cstrAt(b []byte, o int) string {
	e := o
	for e < len(b) && b[e] != 0 {
		e++
	}
	return string(b[o:e])
}

// boneBaseName strips the rig suffixes: hara_n_x → hara, mune_jnt → mune,
// kata_l_n_y → kata_l, hiji_l_eff → hiji_l.
func boneBaseName(n string) string {
	for _, suf := range []string{"_n_x", "_n_y", "_n_z", "_chn", "_jnt", "_eff"} {
		n = strings.TrimSuffix(n, suf)
	}
	return n
}

// ---------- CHR_*.bin ----------

type chrDesc struct {
	skel    int
	attach  map[int]int   // bone -> part
	lodBone map[int]bool  // bones referenced by LOD records
	bind    map[int]mat4  // bone -> bind (bone→world), from the inverse-bind table
	invBind map[int]mat4  // bone -> inverse bind
}

func parseChrDesc(b []byte, sk *chrSkeleton) (*chrDesc, error) {
	d := &chrDesc{
		skel:   int(u16(b, 0)),
		attach: map[int]int{}, lodBone: map[int]bool{},
		bind: map[int]mat4{}, invBind: map[int]mat4{},
	}
	nAttach, nLOD, nMat := int(u16(b, 2)), int(u16(b, 4)), int(u16(b, 6))
	attOff, lodOff, matOff := int(u32(b, 8)), int(u32(b, 12)), int(u32(b, 24))
	for i := 0; i < nAttach; i++ {
		o := attOff + i*8
		d.attach[int(u16(b, o))] = int(u16(b, o+4))
	}
	for i := 0; i < nLOD; i++ {
		o := lodOff + i*20
		n := int(u16(b, o+2))
		for k := 0; k < n && k < 4; k++ {
			d.lodBone[int(u16(b, o+4+2*k))] = true
		}
	}
	// inverse binds: alphabetical by bone base name over the LOD bone union
	var bones []int
	for bi := range d.lodBone {
		bones = append(bones, bi)
	}
	sort.Slice(bones, func(i, j int) bool {
		return boneBaseName(sk.bones[bones[i]].name) < boneBaseName(sk.bones[bones[j]].name)
	})
	if len(bones) != nMat {
		return nil, fmt.Errorf("chr desc: %d LOD bones vs %d matrices", len(bones), nMat)
	}
	for i, bi := range bones {
		o := matOff + i*0x40
		// stored row-vector (p·M + t); convert to column-vector M·p
		var m mat4
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				m[c][r] = f32(b, o+16*r+4*c)
			}
		}
		m[0][3], m[1][3], m[2][3] = f32(b, o+48), f32(b, o+52), f32(b, o+56)
		m[3] = [4]float32{0, 0, 0, 1}
		d.invBind[bi] = m
		d.bind[bi] = mInv(m)
	}
	return d, nil
}

// ---------- motion clips ----------

type motKey struct {
	frame float32
	val   float32
	slope []float32
}

type motChan struct {
	keys []motKey
}

type chrClip struct {
	name   string
	frames int
	// ch[bone][comp], comp 0-5 = rx ry rz tx ty tz; nil channel = rest
	ch map[int]*[6]*motChan
}

// parseMotFile decodes one inflated mot container; names, if non-nil, is the
// motdata_table clip-name list in directory order.
func parseMotFile(b []byte, names []string) ([]*chrClip, error) {
	var clips []*chrClip
	for o := 0; o+0x18 <= len(b); o += 0x18 {
		if u16(b, o+4) == 0xFFFF {
			break
		}
		frames := int(u32(b, o+8))
		descBytes, dataOff := int(u32(b, o+12)), int(u32(b, o+16))
		if dataOff <= 0 || dataOff >= len(b) {
			break
		}
		c := &chrClip{frames: frames, ch: map[int]*[6]*motChan{}}
		if i := len(clips); names != nil && i < len(names) {
			c.name = names[i]
		}
		p := dataOff + 4
		type desc struct{ bone, mask byte }
		var descs []desc
		for i := 0; i < descBytes/2; i++ {
			descs = append(descs, desc{b[p], b[p+1]})
			p += 2
		}
		for _, dd := range descs {
			for bit := 0; bit < 6; bit++ {
				if dd.mask&(1<<bit) == 0 {
					continue
				}
				ch := &motChan{}
				for {
					if p+2 > len(b) {
						return nil, fmt.Errorf("mot: overrun at %#x", p)
					}
					frame, flags := int(b[p]), b[p+1]
					p += 2
					frame |= int(flags&0x1F) << 8
					k := motKey{frame: float32(frame)}
					nval := 0
					switch flags & 0xE0 {
					case 0x00:
					case 0x20, 0x40:
						nval = 1
					case 0x80:
						nval = 2
					case 0xC0:
						nval = 3
					default:
						return nil, fmt.Errorf("mot: flags %#x at %#x", flags, p-2)
					}
					for e := 0; e < nval; e++ {
						v := halfFloat(u16(b, p))
						p += 2
						if e == 0 {
							k.val = v
						} else {
							k.slope = append(k.slope, v)
						}
					}
					ch.keys = append(ch.keys, k)
					if flags&0x20 != 0 || frame >= frames-1 {
						break
					}
					if len(ch.keys) > 8192 {
						return nil, fmt.Errorf("mot: runaway channel")
					}
				}
				bi := int(dd.bone)
				if c.ch[bi] == nil {
					c.ch[bi] = &[6]*motChan{}
				}
				c.ch[bi][bit] = ch
			}
		}
		clips = append(clips, c)
	}
	return clips, nil
}

// halfFloat decodes an IEEE 754 half.
func halfFloat(h uint16) float32 {
	sign := uint32(h>>15) << 31
	exp := uint32(h >> 10 & 0x1F)
	man := uint32(h & 0x3FF)
	switch exp {
	case 0:
		if man == 0 {
			return math.Float32frombits(sign)
		}
		for man&0x400 == 0 {
			man <<= 1
			exp--
		}
		exp++
		man &= 0x3FF
		return math.Float32frombits(sign | (exp+112)<<23 | man<<13)
	case 31:
		return math.Float32frombits(sign | 0xFF<<23 | man<<13)
	}
	return math.Float32frombits(sign | (exp+112)<<23 | man<<13)
}

// eval samples the channel at a frame (linear between keys).
func (ch *motChan) eval(f float32) float32 {
	ks := ch.keys
	if len(ks) == 0 {
		return 0
	}
	if f <= ks[0].frame {
		return ks[0].val
	}
	for i := 1; i < len(ks); i++ {
		if f <= ks[i].frame {
			k0, k1 := ks[i-1], ks[i]
			if k1.frame == k0.frame {
				return k1.val
			}
			t := (f - k0.frame) / (k1.frame - k0.frame)
			return k0.val + (k1.val-k0.val)*t
		}
	}
	return ks[len(ks)-1].val
}

// ---------- pose evaluation ----------

type poseConv struct {
	order int  // euler order for mEuler
	rt    bool // local = R·T when true, else T·R
}

// chrRig bundles everything the evaluator needs.
type chrRig struct {
	sk   *chrSkeleton
	desc *chrDesc
	conv poseConv
}

// chainOf walks a chain root's single-child spine to its effector.
func (rg *chrRig) chainOf(root int) (joints []int, eff int) {
	b := root
	for {
		if len(rg.sk.bones[b].children) == 0 {
			return joints, -1
		}
		b = rg.sk.bones[b].children[0]
		if rg.sk.bones[b].kind == 4 {
			return joints, b
		}
		joints = append(joints, b)
		if len(joints) > 8 {
			return joints, -1
		}
	}
}

// evalPose returns the world matrix of every bone at the given frame, in the
// character's Y-up ground space.
func (rg *chrRig) evalPose(clip *chrClip, frame float32) []mat4 {
	sk := rg.sk
	world := make([]mat4, len(sk.bones))
	local := make([]mat4, len(sk.bones))

	for i, b := range sk.bones {
		var rx, ry, rz float32 = b.restR[0], b.restR[1], b.restR[2]
		// rest translation slots map to the PARENT frame's (y, z): local
		// t = (0, f0, f1) — bind-verified: momo under kosi = (0, -0.19,
		// -0.09) exactly; limb links (sune, asi, hiji, te) run along +x
		// from the binds, not from rest slots
		tx, ty, tz := float32(0), b.restT[0], b.restT[1]
		if chs := clip.ch[i]; chs != nil {
			if chs[0] != nil || chs[1] != nil || chs[2] != nil {
				// animated rotations replace the rest euler (clip constants
				// reproduce the rest values exactly)
				rx, ry, rz = 0, 0, 0
				if chs[0] != nil {
					rx = chs[0].eval(frame)
				}
				if chs[1] != nil {
					ry = chs[1].eval(frame)
				}
				if chs[2] != nil {
					rz = chs[2].eval(frame)
				}
			}
			// animated translations ADD to the rest offset (the idle clip
			// slouches the root by ty -0.098 from its rest 1.24)
			if chs[3] != nil {
				tx += chs[3].eval(frame)
			}
			if chs[4] != nil {
				ty += chs[4].eval(frame)
			}
			if chs[5] != nil {
				tz += chs[5].eval(frame)
			}
		}
		R := mEuler(rx, ry, rz, rg.conv.order)
		T := mTrans(tx, ty, tz)
		if rg.conv.rt {
			local[i] = mMul(R, T)
		} else {
			local[i] = mMul(T, R)
		}
	}

	var walk func(i int, parent mat4)
	walk = func(i int, parent mat4) {
		world[i] = mMul(parent, local[i])
		for _, c := range sk.bones[i].children {
			walk(c, world[i])
		}
	}
	for i, b := range sk.bones {
		if b.parent == -1 {
			walk(i, mIdent())
		}
	}

	// IK pass: chains rooted at kind-2 bones whose effector has animated
	// target translations. Targets are authored relative to the root's
	// animated excursion (the flag-wave clip walks the root up onto the
	// gantry deck and the foot targets stay near y 0.13 — they ride along).
	var rootDelta [3]float32
	for i, b := range sk.bones {
		if b.parent == -1 {
			if chs := clip.ch[i]; chs != nil {
				if chs[3] != nil {
					rootDelta[0] = chs[3].eval(frame)
				}
				if chs[4] != nil {
					rootDelta[1] = chs[4].eval(frame)
				}
				if chs[5] != nil {
					rootDelta[2] = chs[5].eval(frame)
				}
			}
			break
		}
	}
	for ci, b := range sk.bones {
		if b.kind != 2 {
			continue
		}
		joints, eff := rg.chainOf(ci)
		if eff < 0 {
			continue
		}
		chs := clip.ch[eff]
		if chs == nil || (chs[3] == nil && chs[4] == nil && chs[5] == nil) {
			continue // no target: leave the FK pose
		}
		var tgt [3]float32
		if chs[3] != nil {
			tgt[0] = chs[3].eval(frame)
		}
		if chs[4] != nil {
			tgt[1] = chs[4].eval(frame)
		}
		if chs[5] != nil {
			tgt[2] = chs[5].eval(frame)
		}
		ride := rootDelta
		if strings.Contains(sk.bones[eff].name, "sune") && ride[1] < 0 {
			// the feet stay planted when the root slouches; they only ride
			// the root UP (the wave clips carry the whole body)
			ride[1] = 0
		}
		tgt = v3add(tgt, ride)
		rg.solveChain(world, ci, joints, eff, tgt)
		// re-propagate below the solved bones
		for _, j := range append(append([]int{}, joints...), eff) {
			for _, c := range sk.bones[j].children {
				if c != eff && !contains(joints, c) {
					walk(c, world[j])
				}
			}
		}
	}
	return world
}

func contains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// solveChain poses the chain's joint bones toward the target.
//
// Two-joint chains (arms, legs) use the standard two-bone law-of-cosines
// solve: segment lengths come from the bind origins, and the bend plane is
// taken from the FK pose (the animated chain-root rotation puts the middle
// joint roughly where it should be; IK snaps the chain onto the target while
// bending toward that hint). Single-joint chains (spine, head) are a
// look-at: the joint keeps its FK roll and rotates its +x onto the target
// direction. Joint frames keep +x along the limb — the rigid parts are
// modeled with the limb running along local +x.
func (rg *chrRig) solveChain(world []mat4, root int, joints []int, eff int, tgt [3]float32) {
	bind := rg.desc.bind
	switch len(joints) {
	case 2:
		j1, j2 := joints[0], joints[1]
		b1, ok1 := bind[j1]
		b2, ok2 := bind[j2]
		be, oke := bind[effBindProxy(rg.sk, eff)]
		if !ok1 || !ok2 || !oke {
			return
		}
		l1 := v3len(v3sub(mPos(b2), mPos(b1)))
		l2 := v3len(v3sub(mPos(be), mPos(b2)))
		p0 := mPos(world[j1])
		d := v3sub(tgt, p0)
		dl := v3len(d)
		if dl < 1e-5 {
			return
		}
		if dl > l1+l2 {
			dl = l1 + l2
			tgt = v3add(p0, v3scale(v3norm(d), dl))
			d = v3sub(tgt, p0)
		}
		if min := float32(math.Abs(float64(l1 - l2))); dl < min {
			dl = min + 1e-4
			tgt = v3add(p0, v3scale(v3norm(d), dl))
			d = v3sub(tgt, p0)
		}
		dn := v3norm(d)
		// distance from p0 to the elbow's projection on the axis
		a := (l1*l1 - l2*l2 + dl*dl) / (2 * dl)
		h2 := l1*l1 - a*a
		if h2 < 0 {
			h2 = 0
		}
		h := float32(math.Sqrt(float64(h2)))
		// pole: FK middle-joint direction off the axis
		fkMid := mPos(world[j2])
		off := v3sub(fkMid, v3add(p0, v3scale(dn, v3dot(v3sub(fkMid, p0), dn))))
		if v3len(off) < 1e-5 {
			// degenerate: pick any perpendicular
			off = v3cross(dn, [3]float32{0, 1, 0})
			if v3len(off) < 1e-5 {
				off = v3cross(dn, [3]float32{1, 0, 0})
			}
		}
		off = v3norm(off)
		elbow := v3add(v3add(p0, v3scale(dn, a)), v3scale(off, h))
		// orient each limb by rotating its BIND frame onto the current
		// segment direction (minimal rotation) — keeps the mesh's roll and
		// secondary axes consistent with the bind instead of rebuilding
		// frames from scratch
		bp := effBindProxy(rg.sk, eff)
		world[j1] = alignBind(p0, v3sub(elbow, p0), v3sub(mPos(b2), mPos(b1)), b1)
		world[j2] = alignBind(elbow, v3sub(tgt, elbow), v3sub(mPos(rg.desc.bind[bp]), mPos(b2)), b2)
		m := stripPos(world[j2])
		m[0][3], m[1][3], m[2][3] = tgt[0], tgt[1], tgt[2]
		world[eff] = m
	case 1:
		// Spine and head: the segment lengths exist only in the binds (the
		// rest locals of jnt/eff are zero), so the chain solve only LIFTS
		// the joint/effector along the spine's continuation. Orientation
		// stays with FK — the chain-root bones (mune_chn, kao_chn) are
		// themselves animated, so the rotations are already in the clip;
		// the effector target is the IK goal the game refines against, and
		// FK alone already lands close.
		j1 := joints[0]
		p0 := mPos(world[j1])
		lJnt, lEff := rg.chainLift(root, j1, eff)
		up := rg.spineUp(world, root)
		jp := v3add(p0, v3scale(up, lJnt))
		m := world[j1]
		m[0][3], m[1][3], m[2][3] = jp[0], jp[1], jp[2]
		world[j1] = m
		e := v3add(p0, v3scale(up, lEff))
		m[0][3], m[1][3], m[2][3] = e[0], e[1], e[2]
		world[eff] = m
		_ = tgt
	}
}

// spineUp is the direction a one-segment chain extends along: from the
// chain root's parent frames — the vector from the waist to the chain root,
// or world up when degenerate.
func (rg *chrRig) spineUp(world []mat4, root int) [3]float32 {
	// find the waist (kosi) as the torso's lower reference
	for i, b := range rg.sk.bones {
		if boneBaseName(b.name) == "kosi" {
			d := v3sub(mPos(world[root]), mPos(world[i]))
			if v3len(d) > 1e-4 {
				return v3norm(d)
			}
		}
	}
	return [3]float32{0, 1, 0}
}

// chainLift derives the one-segment chains' joint/effector distances from
// the bind pose. The chest chain's effector is where the shoulder pair
// hangs: the mean of the kata binds (their lateral rest offsets cancel).
// The head chain's joint is the head bind above that same shoulder line.
func (rg *chrRig) chainLift(root, jnt, eff int) (lJnt, lEff float32) {
	sk, bind := rg.sk, rg.desc.bind
	// mean of all kata_* binds = shoulder-line centre
	var katas [][3]float32
	for i, b := range sk.bones {
		if strings.HasPrefix(boneBaseName(b.name), "kata") {
			if bm, ok := bind[i]; ok {
				katas = append(katas, mPos(bm))
			}
		}
	}
	var shoulder [3]float32
	for _, k := range katas {
		shoulder = v3add(shoulder, k)
	}
	if len(katas) > 0 {
		shoulder = v3scale(shoulder, 1/float32(len(katas)))
	}
	jb, hasJb := bind[jnt]
	switch {
	case hasJb && strings.HasPrefix(boneBaseName(sk.bones[jnt].name), "kao"):
		// head: pivot is the shoulder line (kubi/kao_chn carry no offsets)
		l := v3len(v3sub(mPos(jb), shoulder))
		return l, l
	case hasJb:
		// chest: joint ~at the pivot, effector at the shoulder line
		return 0, v3len(v3sub(shoulder, mPos(jb)))
	}
	return 0, 0
}

// effBindProxy: effectors have no bind matrix; their first plain child (te,
// asi...) sits at the chain end and does.
func effBindProxy(sk *chrSkeleton, eff int) int {
	if len(sk.bones[eff].children) > 0 {
		return sk.bones[eff].children[0]
	}
	return eff
}

// stripPos keeps rotation only.
func stripPos(m mat4) mat4 {
	m[0][3], m[1][3], m[2][3] = 0, 0, 0
	return m
}

// alignBind places bind frame B at p, rotated by the minimal rotation that
// carries the bind-pose segment direction onto the current one.
func alignBind(p, curDir, bindDir [3]float32, B mat4) mat4 {
	f := v3norm(bindDir)
	t := v3norm(curDir)
	axis := v3cross(f, t)
	sn := v3len(axis)
	cs := v3dot(f, t)
	var R mat4
	if sn < 1e-6 {
		R = mIdent()
		if cs < 0 {
			// opposite: rotate π about any perpendicular
			perp := v3cross(f, [3]float32{0, 1, 0})
			if v3len(perp) < 1e-5 {
				perp = v3cross(f, [3]float32{1, 0, 0})
			}
			R = axisAngle(v3norm(perp), math.Pi)
		}
	} else {
		R = axisAngle(v3scale(axis, 1/sn), math.Atan2(float64(sn), float64(cs)))
	}
	m := mMul(R, stripPos(B))
	m[0][3], m[1][3], m[2][3] = p[0], p[1], p[2]
	return m
}

// axisAngle builds a rotation about a unit axis.
func axisAngle(u [3]float32, ang float64) mat4 {
	c, sn := float32(math.Cos(ang)), float32(math.Sin(ang))
	x, y, z := u[0], u[1], u[2]
	ic := 1 - c
	return mat4{
		{c + x*x*ic, x*y*ic - z*sn, x*z*ic + y*sn, 0},
		{y*x*ic + z*sn, c + y*y*ic, y*z*ic - x*sn, 0},
		{z*x*ic - y*sn, z*y*ic + x*sn, c + z*z*ic, 0},
		{0, 0, 0, 1},
	}
}

// frameAt builds a frame at p with +x toward q, keeping the reference
// frame's roll as closely as possible (its y axis projected off the new x).
func frameAt(p, q [3]float32, ref mat4) mat4 {
	x := v3norm(v3sub(q, p))
	refY := [3]float32{ref[0][1], ref[1][1], ref[2][1]}
	y := v3sub(refY, v3scale(x, v3dot(refY, x)))
	if v3len(y) < 1e-5 {
		refZ := [3]float32{ref[0][2], ref[1][2], ref[2][2]}
		y = v3cross(refZ, x)
	}
	y = v3norm(y)
	z := v3cross(x, y)
	var m mat4
	for i := 0; i < 3; i++ {
		m[i][0], m[i][1], m[i][2] = x[i], y[i], z[i]
	}
	m[0][3], m[1][3], m[2][3] = p[0], p[1], p[2]
	m[3] = [4]float32{0, 0, 0, 1}
	return m
}

// ---------- loading ----------

type chrData struct {
	sk    *chrSkeleton
	desc  *chrDesc
	model *pmt
	texs  []texInfo
	clips []*chrClip // both mot files, named
}

func loadFlagman(disc *xbox.Image) (*chrData, error) {
	boneRaw, err := disc.ReadFile("/Common/bone.bin")
	if err != nil {
		return nil, err
	}
	sks := parseBoneLib(boneRaw)
	var ozi *chrSkeleton
	descRaw, err := disc.ReadFile("/Chr/CHR_AUT04.bin")
	if err != nil {
		return nil, err
	}
	skelIdx := int(u16(descRaw, 0))
	if skelIdx >= len(sks) {
		return nil, fmt.Errorf("skeleton %d out of range", skelIdx)
	}
	ozi = &sks[skelIdx]
	desc, err := parseChrDesc(descRaw, ozi)
	if err != nil {
		return nil, err
	}

	model, err := readPMT(disc, "/Chr/obj_chr_aut04_pmt.sz")
	if err != nil {
		return nil, err
	}
	texs, _, err := model.parseTextures()
	if err != nil {
		return nil, err
	}

	tblRaw, err := disc.ReadFile("/Common/motdata_table.bin")
	if err != nil {
		return nil, err
	}
	var clips []*chrClip
	for _, mf := range []string{"mot_ETC_bin", "mot_OR2SP_ETC_bin"} {
		names := motTableNames(tblRaw, mf+".gz")
		raw, err := readInflated(disc, "/Anims/"+mf+".sz")
		if err != nil {
			return nil, err
		}
		cs, err := parseMotFile(raw, names)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", mf, err)
		}
		clips = append(clips, cs...)
	}
	return &chrData{sk: ozi, desc: desc, model: model, texs: texs, clips: clips}, nil
}

func (cd *chrData) clip(name string) *chrClip {
	for _, c := range cd.clips {
		if c.name == name {
			return c
		}
	}
	return nil
}

// motTableNames returns the ordered clip names of one mot_*.gz.
func motTableNames(b []byte, file string) []string {
	for o := 0; o+16 <= len(b); o += 16 {
		nameOff, listOff, cnt := int(u32(b, o)), int(u32(b, o+4)), int(u32(b, o+8))
		if nameOff == 0 || nameOff >= len(b) {
			break
		}
		if cstrAt(b, nameOff) != file {
			continue
		}
		names := make([]string, cnt)
		for i := 0; i < cnt; i++ {
			names[i] = cstrAt(b, int(u32(b, listOff+4*i)))
		}
		return names
	}
	return nil
}

func readPMT(disc *xbox.Image, path string) (*pmt, error) {
	data, err := readInflated(disc, path)
	if err != nil {
		return nil, err
	}
	base := strings.TrimSuffix(strings.TrimPrefix(path, "/Chr/"), "_pmt.sz")
	return parsePMT(base, data)
}

func readInflated(disc *xbox.Image, path string) ([]byte, error) {
	raw, err := disc.ReadFile(path)
	if err != nil {
		return nil, err
	}
	zr, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(zr)
}

// ---------- convention fit against a live capture ----------

// chrFit reads a bootoracle -carvtx dump of the flagman's draws and searches
// clips, frames and pose conventions for the FK/IK prediction that matches
// the captured relative part transforms.
func chrFit(disc *xbox.Image, dumpPath string) {
	cd, err := loadFlagman(disc)
	if err != nil {
		fatal("%v", err)
	}
	live, err := readCarvtxDump(dumpPath, cd.model)
	if err != nil {
		fatal("%v", err)
	}
	fmt.Printf("live rigid parts: %d\n", len(live))

	// part -> bone
	part2bone := map[int]int{}
	for b, p := range cd.desc.attach {
		part2bone[p] = b
	}
	var parts []int
	for p := range live {
		if _, ok := part2bone[p]; ok {
			parts = append(parts, p)
		}
	}
	sort.Ints(parts)
	ref := parts[0]

	type result struct {
		clip  string
		frame float32
		conv  poseConv
		err   float64
	}
	best := result{err: math.Inf(1)}
	var all []result
	for _, clip := range cd.clips {
		for _, order := range []int{0, 1, 2, 3, 4, 5} {
			for _, rt := range []bool{true, false} {
				rg := &chrRig{sk: cd.sk, desc: cd.desc, conv: poseConv{order, rt}}
				for f := 0; f < clip.frames; f++ {
					world := rg.evalPose(clip, float32(f))
					var e float64
					for _, p := range parts[1:] {
						b, br := part2bone[p], part2bone[ref]
						pred := mMul(mInv(world[br]), world[b])
						meas := mMul(mInvFull(live[ref]), live[p])
						for i := 0; i < 3; i++ {
							e += math.Abs(float64(pred[i][3] - meas[i][3]))
							for j := 0; j < 3; j++ {
								e += 0.1 * math.Abs(float64(pred[i][j]-meas[i][j]))
							}
						}
					}
					r := result{clip.name, float32(f), poseConv{order, rt}, e}
					all = append(all, r)
					if e < best.err {
						best = r
					}
				}
			}
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].err < all[j].err })
	for i := 0; i < 12 && i < len(all); i++ {
		r := all[i]
		fmt.Printf("top%2d: clip=%-28s frame=%4g order=%d rt=%v err=%.3f\n",
			i, r.clip, r.frame, r.conv.order, r.conv.rt, r.err)
	}
	fmt.Printf("best: clip=%s frame=%g order=%d rt=%v err=%.3f\n",
		best.clip, best.frame, best.conv.order, best.conv.rt, best.err)

	// residual detail at the best fit
	var cbest *chrClip
	for _, c := range cd.clips {
		if c.name == best.clip {
			cbest = c
		}
	}
	rg := &chrRig{sk: cd.sk, desc: cd.desc, conv: best.conv}
	world := rg.evalPose(cbest, best.frame)
	for _, p := range parts[1:] {
		b, br := part2bone[p], part2bone[ref]
		pred := mMul(mInv(world[br]), world[b])
		meas := mMul(mInvFull(live[ref]), live[p])
		fmt.Printf("  part %2d (%s): pred t (%7.3f %7.3f %7.3f)  meas t (%7.3f %7.3f %7.3f)\n",
			p, cd.sk.bones[b].name,
			pred[0][3], pred[1][3], pred[2][3], meas[0][3], meas[1][3], meas[2][3])
	}
}

// readCarvtxDump parses "carvtx: DRAW attr0=... | 16 floats" lines into
// column-vector S = VP·M matrices per rigid part (fmt 1832 draws only).
func readCarvtxDump(path string, model *pmt) (map[int]mat4, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// part VB offsets
	type span struct{ off, pi int }
	var spans []span
	for pi := 0; pi < model.nParts; pi++ {
		rec := 0x18 + pi*0x3C
		w2 := u32(model.a, rec+4*2)
		vbDesc := u32(model.a, int(w2))
		spans = append(spans, span{int(u32(model.a, int(vbDesc)+4)), pi})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].off < spans[j].off })
	// collect draws first, then derive the RAM base: the lowest captured
	// attr0 belongs to the lowest-offset part that was drawn, and every part
	// is drawn, so base = min(attr0) - min(vbOff).
	type draw struct {
		addr int
		m    mat4
	}
	var draws []draw
	minAddr := math.MaxInt
	for _, ln := range strings.Split(string(raw), "\n") {
		if !strings.Contains(ln, "DRAW attr0=") {
			continue
		}
		var addr int
		var fmtw string
		if _, err := fmt.Sscanf(ln[strings.Index(ln, "attr0="):], "attr0=%X fmt0=%s", &addr, &fmtw); err != nil {
			continue
		}
		if fmtw != "00001832" {
			continue
		}
		fl := strings.Split(ln, "|")
		if len(fl) != 5 {
			continue
		}
		var vals []float32
		for _, part := range fl[1:] {
			for _, tok := range strings.Fields(part) {
				var v float64
				fmt.Sscanf(tok, "%g", &v)
				vals = append(vals, float32(v))
			}
		}
		if len(vals) != 16 {
			continue
		}
		var m mat4
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				m[i][j] = vals[i*4+j]
			}
		}
		draws = append(draws, draw{addr, m})
		if addr < minAddr {
			minAddr = addr
		}
	}
	// part 0 leads the model's VB, so its offset anchors the base
	part0Off := 0
	for _, s := range spans {
		if s.pi == 0 {
			part0Off = s.off
		}
	}
	base := minAddr - part0Off
	out := map[int]mat4{}
	for _, d := range draws {
		off := d.addr - base
		pi := -1
		for _, s := range spans {
			if off >= s.off {
				pi = s.pi
			} else {
				break
			}
		}
		if pi < 0 {
			continue
		}
		if _, seen := out[pi]; seen {
			continue
		}
		out[pi] = d.m
	}
	return out, nil
}

// ---------- GLB export ----------

// chrPartPrims builds the glb primitives of one rigid part (bone-local
// vertices, textured per material like the car path).
func chrPartPrims(model *pmt, texs []texInfo, pi int) ([]glb.Prim, error) {
	pt, err := model.parsePart(pi)
	if err != nil {
		return nil, err
	}
	type group struct {
		tex  int
		tris [][3]uint32
	}
	var pos [][3]float32
	var nrm [][3]float32
	var uvs [][2]float32
	vbase := make([]uint32, len(pt.pairs))
	for k, bp := range pt.pairs {
		vbase[k] = uint32(len(pos))
		p2, n2, u2, _, _ := model.decodeVerts(bp)
		pos = append(pos, p2...)
		nrm = append(nrm, n2...)
		uvs = append(uvs, u2...)
	}
	groups := map[int][][3]uint32{}
	for _, b := range pt.batches {
		bp := pt.pairs[b.pair]
		idxCount, _ := indexCount(b.prim, b.prims)
		raw := make([]uint32, idxCount)
		nVerts := bp.vbBytes / bp.stride
		for i := range raw {
			ix := uint32(u16(model.a, int(bp.ibOff)+int(b.first+uint32(i))*2)) + b.baseVtx
			if ix >= nVerts {
				return nil, fmt.Errorf("part %d: index out of VB", pi)
			}
			raw[i] = ix + vbase[b.pair]
		}
		tris := triangulate(b.prim, raw)
		ti := -1
		if m := pt.mats[b.matIdx]; m.texIdx >= 0 && texs[m.texIdx].img != nil {
			ti = m.texIdx
		}
		groups[ti] = append(groups[ti], tris...)
	}
	var out []glb.Prim
	var keys []int
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, ti := range keys {
		pr := glb.Prim{Positions: pos, Normals: nrm, UVs: uvs, Tris: groups[ti], DoubleSided: true}
		if ti >= 0 {
			pr.Image = texs[ti].img
		}
		out = append(out, pr)
	}
	return out, nil
}

// chrExportGLB writes the animated flagman GLB. The geometry-bone nodes
// form a HIERARCHY (hip → thigh → shin → foot, chest → arm chain …) and the
// clips carry per-node LOCAL TRS — interpolating locals keeps every limb
// attached between keyframes, which a flat world-space bake cannot (parts
// lerp along straight lines and a fast sweep tears the chain apart
// mid-interval; the first shipped bake did exactly that).
func chrExportGLB(disc *xbox.Image, outPath string, conv poseConv) error {
	cd, err := loadFlagman(disc)
	if err != nil {
		return err
	}
	rg := &chrRig{sk: cd.sk, desc: cd.desc, conv: conv}

	// geometry bones parented by nearest geometry ancestor
	byName := map[string]int{}
	for i, b := range cd.sk.bones {
		byName[b.name] = i
	}
	geomParent := func(bone int) int {
		p := cd.sk.bones[bone].parent
		for p >= 0 {
			if _, ok := cd.desc.bind[p]; ok {
				if _, isGeom := cd.desc.attach[p]; isGeom {
					return p
				}
			}
			p = cd.sk.bones[p].parent
		}
		return -1
	}

	// attach parts sorted for stable node order; parents before children
	type pb struct{ part, bone int }
	var parts []pb
	for b, p := range cd.desc.attach {
		parts = append(parts, pb{p, b})
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].part < parts[j].part })
	order := make([]pb, 0, len(parts))
	added := map[int]bool{}
	for len(order) < len(parts) {
		for _, x := range parts {
			if added[x.bone] {
				continue
			}
			if gp := geomParent(x.bone); gp == -1 || added[gp] {
				order = append(order, x)
				added[x.bone] = true
			}
		}
	}

	s := glb.NewScene()
	root := s.AddNode("flagman", -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
	nodes := map[int]int{} // bone -> node
	for _, x := range order {
		bind := cd.desc.bind[x.bone]
		local := bind
		parentNode := root
		if gp := geomParent(x.bone); gp >= 0 {
			local = mMul(mInv(cd.desc.bind[gp]), bind)
			parentNode = nodes[gp]
		}
		name := cd.sk.bones[x.bone].name
		n := s.AddNode(name, parentNode, mPos(local), mQuat(local), [3]float32{1, 1, 1})
		prims, err := chrPartPrims(cd.model, cd.texs, x.part)
		if err != nil {
			return err
		}
		if err := s.AddMesh(n, name, prims); err != nil {
			return err
		}
		nodes[x.bone] = n
	}

	const fps = 60.0
	const step = 3 // sample every 3 frames (20 Hz; locals interpolate cleanly)
	bake := func(clipName, asName string, ground bool) error {
		clip := cd.clip(clipName)
		if clip == nil {
			return fmt.Errorf("clip %s not found", clipName)
		}
		c := s.NewClip(asName)
		var times []float32
		var frames []float32
		for f := 0; f < clip.frames; f += step {
			times = append(times, float32(f)/fps)
			frames = append(frames, float32(f))
		}
		// close the loop exactly at the final frame
		times = append(times, float32(clip.frames-1)/fps)
		frames = append(frames, float32(clip.frames-1))
		worlds := make([][]mat4, len(frames))
		for i, fr := range frames {
			worlds[i] = rg.evalPose(clip, fr)
		}
		if ground {
			groundBaseline(worlds, cd, byName)
		}
		rot := map[int][][4]float32{}
		trn := map[int][][3]float32{}
		for _, world := range worlds {
			for b := range nodes {
				m := world[b]
				if gp := geomParent(b); gp >= 0 {
					m = mMul(mInv(world[gp]), world[b])
				}
				rot[b] = append(rot[b], mQuat(m))
				trn[b] = append(trn[b], mPos(m))
			}
		}
		for b, n := range nodes {
			// keep quaternion continuity for LINEAR interpolation
			qs := rot[b]
			for i := 1; i < len(qs); i++ {
				if qs[i][0]*qs[i-1][0]+qs[i][1]*qs[i-1][1]+qs[i][2]*qs[i-1][2]+qs[i][3]*qs[i-1][3] < 0 {
					for k := 0; k < 4; k++ {
						qs[i][k] = -qs[i][k]
					}
				}
			}
			c.Rotations(n, times, qs)
			c.Vec3s(n, "translation", times, trn[b])
		}
		c.Finish()
		return nil
	}
	if err := bake("ORT_OZI_OZI_STAND_SP_LP", "idle", false); err != nil {
		return err
	}
	if err := bake("ORT_OZI_OZI_HATA_SP", "hatafuri", true); err != nil {
		return err
	}
	return s.Write(outPath, "flagman")
}

// groundBaseline shifts each frame of a baked clip down by a min-filtered
// baseline so the lowest foot tracks the ground. HATA_SP is authored for
// the start-gantry anchor — the old man performs the whole wave on the deck
// ~2 m up and jumps down at the end. The Studio places him on the road, so
// the baseline (a ±0.5 s running minimum of the lower foot's height above
// ankle rest) is removed: hops shorter than the window survive, the deck
// section grounds, and the final dismount eases out instead of falling two
// metres. A presentation choice, documented, not part of the decode.
func groundBaseline(worlds [][]mat4, cd *chrData, byName map[string]int) {
	al, ar := byName["asi_l"], byName["asi_r"]
	n := len(worlds)
	raw := make([]float32, n)
	for i, w := range worlds {
		fl, fr := w[al][1][3], w[ar][1][3]
		if fr < fl {
			fl = fr
		}
		raw[i] = fl - 0.14 // height of the lower ankle above its rest
	}
	const win = 10 // ±10 keys at 20 Hz = ±0.5 s
	for i := range worlds {
		base := raw[i]
		for j := i - win; j <= i+win; j++ {
			if j >= 0 && j < n && raw[j] < base {
				base = raw[j]
			}
		}
		if base < 0 {
			base = 0
		}
		for b := range worlds[i] {
			worlds[i][b][1][3] -= base
		}
	}
}

// chrPoseDebug prints key bone world positions of a clip at given frames.
func chrPoseDebug(disc *xbox.Image, spec string, conv poseConv) {
	cd, err := loadFlagman(disc)
	if err != nil {
		fatal("%v", err)
	}
	var clipName string
	var frames []float32
	parts := strings.Split(spec, ":")
	clipName = parts[0]
	for _, f := range parts[1:] {
		var v float64
		fmt.Sscanf(f, "%g", &v)
		frames = append(frames, float32(v))
	}
	clip := cd.clip(clipName)
	if clip == nil {
		fatal("clip %s not found (have %v)", clipName, len(cd.clips))
	}
	rg := &chrRig{sk: cd.sk, desc: cd.desc, conv: conv}
	names := []string{"hara_n_x", "kosi", "mune_jnt", "kao_jnt", "ude_l_jnt", "hiji_l_jnt", "hiji_l_eff", "te_l", "ude_r_jnt", "te_r", "asi_l", "asi_r"}
	for _, fr := range frames {
		world := rg.evalPose(clip, fr)
		fmt.Printf("frame %g:\n", fr)
		for _, nm := range names {
			for i, b := range cd.sk.bones {
				if b.name == nm {
					p := mPos(world[i])
					fmt.Printf("  %-11s (%7.3f %7.3f %7.3f)\n", nm, p[0], p[1], p[2])
				}
			}
		}
	}
}

// exportFlagman writes the flagman object into the Studio site: the
// animated GLB plus its object asset. His stage-beac placement is added by
// exportStage (the spot is a measured fact: the live race-driving frame
// puts his feet at world (-4.4, 0, -21.4), character axes ~world axes).
func exportFlagman(disc *xbox.Image, b *build.Builder) error {
	out, err := b.Path("objects", "flagman.glb")
	if err != nil {
		return err
	}
	// ZYX euler, T·R locals — the convention set the CHR_AUT04 bind
	// matrices themselves select (see the bind-relative locals fit).
	if err := chrExportGLB(disc, out, poseConv{5, false}); err != nil {
		return err
	}
	b.AddObject(schema.Asset{ID: "flagman", Name: "The flagman", Group: "Characters"},
		&schema.Object{Type: schema.ObjectModel3D, Name: "The flagman", Model: "flagman.glb",
			Animations: []schema.Animation{
				{ID: "idle", Name: "Idle", Loop: "loop", Clip: "idle"},
				{ID: "hatafuri", Name: "Flag wave", Loop: "once", Clip: "hatafuri"},
			},
			Props: map[string]any{"source": "/Chr/obj_chr_aut04_pmt.sz", "skeleton": "OZI (/Common/bone.bin)", "clips": "mot_OR2SP_ETC_bin.sz"}})
	return nil
}
