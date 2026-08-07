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

func mScale(x, y, z float32) mat4 {
	m := mIdent()
	m[0][0], m[1][1], m[2][2] = x, y, z
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

// mDecompose splits a TRS matrix into translation, unit quaternion and
// per-axis scale (column norms; the rigs only ever scale positively).
func mDecompose(m mat4) (t [3]float32, q [4]float32, s [3]float32) {
	t = mPos(m)
	for c := 0; c < 3; c++ {
		s[c] = float32(math.Sqrt(float64(m[0][c]*m[0][c] + m[1][c]*m[1][c] + m[2][c]*m[2][c])))
		if s[c] > 1e-8 {
			for r := 0; r < 3; r++ {
				m[r][c] /= s[c]
			}
		}
	}
	q = mQuat(m)
	return
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
	// restT is the local rest translation (tx, ty, tz): tx is the float at
	// record +0x0C — the BONE LENGTH along the parent's x (the word the
	// first decode read as "FFFF, 0" padding; the game's runtime channel
	// slots surface it: kao_chn 0.0464 = the neck, kata_l_n_x -0.0628,
	// and the IK segment lengths hiji 0.272 / sune 0.390 / eff 0.440 all
	// live there); ty, tz are the two floats at +0x10.
	restT    [3]float32
	restR    [3]float32 // f[2], f[3], f[4]
	restS    [3]float32 // f[5], f[6], f[7] — the ONN rig's HARA twin rests at 0.9
	// restL, when set, is the bone's whole local transform verbatim — used
	// for the synthesized DYNAMICS bones (ponytail/skirt chains from the
	// CHR extra section, simulated by the game at runtime): their local is
	// the bind-relative offset from their chain parent, so they ride it
	// rigidly in the export.
	restL    *mat4
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
			br.restT = [3]float32{f32(b, r+0x0C), f32(b, r+0x10), f32(b, r+0x14)}
			br.restR = [3]float32{f32(b, r+0x18), f32(b, r+0x1C), f32(b, r+0x20)}
			br.restS = [3]float32{f32(b, r+0x24), f32(b, r+0x28), f32(b, r+0x2C)}
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

// lodRec is one skinned-part record: the part's vertices (stored in the
// character's T-pose model space, unlike the bone-local rigid parts) are
// palette-skinned over the listed bones. A vertex's stored weights cover
// bones[0..n-2] in list order; the remainder 1-Σ belongs to the last.
type lodRec struct {
	bones []int
	part  int
}

type chrDesc struct {
	skel    int
	attach  map[int][]int // bone -> rigid parts (full-detail + runtime-attached hand/face variants)
	lods    []lodRec     // skinned junction parts (the seams between rigid parts)
	lodBone map[int]bool // bones referenced by LOD records
	bind    map[int]mat4 // bone -> bind (bone→world), from the inverse-bind table
	invBind map[int]mat4 // bone -> inverse bind
}

// chrDynChain is one runtime-simulated dynamics chain from the CHR extra
// section: ponytails, fringes, skirt hems, shirt tails. The game integrates
// these bones with a spring sim; the export rides them rigidly on their
// chain parent at the bind offset (a declared simplification).
type chrDynChain struct {
	parent int // skeleton bone the chain hangs from
	first  int // first dynamics-bone index (the 0x8000|i space)
	count  int
}

// parseChrExtra reads the extra section: header {nChains, nBones, chainsOff,
// bonesOff, idxOff}; chains are 0x24-byte records {parentBone u32, firstDyn
// u32, count u32, params f32×6}; the bone table pairs {0x8000|i, segment
// length f32} (dr_g00: a 3-segment ponytail + 1-segment fringe on kao_jnt,
// lengths 0.10/0.13/0.14/0.07).
func parseChrExtra(b []byte) (chains []chrDynChain, nDyn int) {
	extra := int(u32(b, 28))
	if extra == 0 || extra+20 > len(b) {
		return nil, 0
	}
	nChains, nBones := int(u32(b, extra)), int(u32(b, extra+4))
	chainsOff := int(u32(b, extra+8))
	for i := 0; i < nChains; i++ {
		o := chainsOff + i*0x24
		chains = append(chains, chrDynChain{
			parent: int(u32(b, o)), first: int(u32(b, o+4)), count: int(u32(b, o+8)),
		})
	}
	return chains, nBones
}

func parseChrDesc(b []byte, sk *chrSkeleton, model *pmt) (*chrDesc, error) {
	d := &chrDesc{
		skel:   int(u16(b, 0)),
		attach: map[int][]int{}, lodBone: map[int]bool{},
		bind: map[int]mat4{}, invBind: map[int]mat4{},
	}
	nAttach, nLOD, nMat := int(u16(b, 2)), int(u16(b, 4)), int(u16(b, 6))
	attOff, lodOff, matOff := int(u32(b, 8)), int(u32(b, 12)), int(u32(b, 24))

	// Synthesize the dynamics bones into the skeleton: chain-linked under
	// their parent, locals fixed later from the bind matrices. The 0x8000
	// flag on a LOD bone slot indexes this space.
	chains, nDyn := parseChrExtra(b)
	dynBase := len(sk.bones)
	if nDyn > 0 {
		for i := 0; i < nDyn; i++ {
			sk.bones = append(sk.bones, chrBone{
				name: fmt.Sprintf("dyn_%d", i), parent: -2,
				restS: [3]float32{1, 1, 1},
			})
		}
		for _, ch := range chains {
			prev := ch.parent
			for k := 0; k < ch.count; k++ {
				vid := dynBase + ch.first + k
				sk.bones[vid].parent = prev
				sk.bones[prev].children = append(sk.bones[prev].children, vid)
				prev = vid
			}
		}
	}

	for i := 0; i < nAttach; i++ {
		o := attOff + i*8
		bi := int(u16(b, o))
		if bi&0x8000 != 0 {
			bi = dynBase + bi&0x7FFF
		}
		d.attach[bi] = append(d.attach[bi], int(u16(b, o+4))&0x7FFF)
	}
	for i := 0; i < nLOD; i++ {
		o := lodOff + i*20
		n := int(u16(b, o+2))
		lr := lodRec{part: int(u16(b, o+16))}
		for k := 0; k < n && k < 4; k++ {
			bi := int(u16(b, o+4+2*k))
			if bi&0x8000 != 0 {
				// dynamics-bone reference (dr_g00's ponytail parts skin
				// over {0x8000..0x8003})
				bi = dynBase + bi&0x7FFF
			}
			lr.bones = append(lr.bones, bi)
			d.lodBone[bi] = true
		}
		d.lods = append(d.lods, lr)
	}
	// Inverse binds: the table covers the geometry bones, but its ORDER is
	// rig-specific (AUT04 sorts by base bone name; the JEN/ONN rigs use a
	// different naming vocabulary, and dr_g00 carries one matrix more than
	// its LOD union). Assign each matrix to the geometry bone whose
	// rest-FK world it matches — bind origins are distinctive (the file
	// binds sit within ~2 cm of rest-FK; rotation breaks stacked-origin
	// ties). Greedy by best score, unique per bone.
	rest := restFK(sk)
	var cand []int
	seen := map[int]bool{}
	for bi := range d.lodBone {
		if bi < dynBase {
			cand = append(cand, bi)
			seen[bi] = true
		}
	}
	for bi := range d.attach {
		if !seen[bi] && bi < dynBase {
			cand = append(cand, bi)
		}
	}
	type fileMat struct {
		inv, bind mat4
	}
	mats := make([]fileMat, nMat)
	for i := 0; i < nMat; i++ {
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
		mats[i] = fileMat{inv: m, bind: mInv(m)}
	}
	// ORPHAN PARTS — the runtime-attached variants (nothing in the CHR
	// references them; the game's code picks per frame): hand SHAPES in
	// mirrored L/R blocks and FACE EXPRESSION sets. The in-race capture
	// pinned the semantics on the driver: of his four orphans the game
	// draws exactly the FIRST shape of each side (parts 22/24), 22 riding
	// te_l, 24 te_r — and the +z-dominant local bbox is the LEFT hand.
	// The export attaches the first face and the first hand pair so the
	// cast ships whole; the other variants stay unshipped (documented).
	attachOrphans(d, sk, model, dynBase)

	// The matrix table's bone order is IN THE FILE: the header word at +16
	// points at a u16 list (right before the attach table) naming the bone
	// of each matrix — skeleton indices, or 0x8000|i for dynamics bones.
	// (Before this list was found, the assignment was reconstructed from
	// rest-pose matching plus model-vertex hints; the list settles rigs no
	// heuristic could — FAL's dynamics matrices interleave as 2,3,0,1,8..)
	listOff := int(u32(b, 16))
	for i := 0; i < nMat; i++ {
		bi := int(u16(b, listOff+2*i))
		if bi&0x8000 != 0 {
			bi = dynBase + bi&0x7FFF
		} else if bi >= dynBase {
			continue
		}
		d.invBind[bi] = mats[i].inv
		d.bind[bi] = mats[i].bind
	}
	_ = model

	// bones the table missed — attach-only bones included — bind at the
	// rest pose (full inverse: the twins carry rest scale)
	for _, bi := range append(append([]int(nil), cand...), dynIDs(dynBase, nDyn)...) {
		if _, ok := d.bind[bi]; !ok {
			d.bind[bi] = rest[bi]
			d.invBind[bi] = mInvFull(rest[bi])
		}
	}

	// Dynamics bones ride their chain parent rigidly at the bind offset.
	for _, ch := range chains {
		prev := ch.parent
		for k := 0; k < ch.count; k++ {
			vid := dynBase + ch.first + k
			pb, ok := d.bind[prev]
			if !ok {
				pb = rest[prev]
			}
			local := mMul(mInvFull(pb), d.bind[vid])
			sk.bones[vid].restL = &local
			prev = vid
		}
	}
	return d, nil
}

// attachOrphans classifies the parts no CHR table references and attaches
// the defaults. Face set = ≥3 consecutive orphans with identical vertex
// count and bbox → first one onto the head bone (name containing "kao",
// preferring the bone that already carries an attach part — the scalp).
// Hand pair = two orphans with equal counts and z-mirrored bboxes → first
// pair onto the te bones, +z-dominant to the left. Tiny mirrored pairs
// (<5 cm, the eyelashes) go with the face. Everything else stays out.
func attachOrphans(d *chrDesc, sk *chrSkeleton, model *pmt, dynBase int) {
	used := map[int]bool{}
	for _, ps := range d.attach {
		for _, p := range ps {
			used[p] = true
		}
	}
	for _, lr := range d.lods {
		used[lr.part] = true
	}
	type info struct {
		pi, verts int
		mn, mx    [3]float32
	}
	var orphans []info
	for pi := 0; pi < model.nParts; pi++ {
		if used[pi] {
			continue
		}
		pt, err := model.parsePart(pi)
		if err != nil {
			continue
		}
		in := info{pi: pi, mn: [3]float32{1e9, 1e9, 1e9}, mx: [3]float32{-1e9, -1e9, -1e9}}
		for _, bp := range pt.pairs {
			pos, _, _, _, _ := model.decodeVerts(bp)
			in.verts += len(pos)
			for _, p := range pos {
				for c := 0; c < 3; c++ {
					if p[c] < in.mn[c] {
						in.mn[c] = p[c]
					}
					if p[c] > in.mx[c] {
						in.mx[c] = p[c]
					}
				}
			}
		}
		orphans = append(orphans, in)
	}
	if len(orphans) == 0 {
		return
	}
	boneByName := func(sub string, preferAttached bool) int {
		best := -1
		for i, b := range sk.bones {
			if !strings.Contains(b.name, sub) {
				continue
			}
			if _, has := d.attach[i]; has && preferAttached {
				return i
			}
			if best < 0 {
				best = i
			}
		}
		return best
	}
	head := boneByName("kao", true)
	teL, teR := boneByName("te_l", false), boneByName("te_r", false)
	near := func(a, b float32) bool { v := a - b; return v > -0.006 && v < 0.006 }
	sameBox := func(a, b info) bool {
		return a.verts == b.verts &&
			near(a.mn[0], b.mn[0]) && near(a.mx[0], b.mx[0]) &&
			near(a.mn[1], b.mn[1]) && near(a.mx[1], b.mx[1]) &&
			near(a.mn[2], b.mn[2]) && near(a.mx[2], b.mx[2])
	}
	mirrored := func(a, b info) bool {
		return a.verts == b.verts &&
			near(a.mn[0], b.mn[0]) && near(a.mx[0], b.mx[0]) &&
			near(a.mn[1], b.mn[1]) && near(a.mx[1], b.mx[1]) &&
			near(a.mn[2], -b.mx[2]) && near(a.mx[2], -b.mn[2])
	}
	// the eyelash pair mirrors in X (face-local), not Z
	mirroredX := func(a, b info) bool {
		return a.verts == b.verts &&
			near(a.mn[1], b.mn[1]) && near(a.mx[1], b.mx[1]) &&
			near(a.mn[2], b.mn[2]) && near(a.mx[2], b.mx[2]) &&
			near(a.mn[0], -b.mx[0]) && near(a.mx[0], -b.mn[0])
	}
	taken := map[int]bool{}
	// face expression set
	if head >= 0 {
		for i := 0; i+2 < len(orphans); i++ {
			n := 1
			for i+n < len(orphans) && sameBox(orphans[i], orphans[i+n]) {
				n++
			}
			if n >= 3 {
				d.attach[head] = append(d.attach[head], orphans[i].pi)
				for k := 0; k < n; k++ {
					taken[orphans[i+k].pi] = true
				}
				break
			}
		}
	}
	handDone := false
	for i := 0; i < len(orphans); i++ {
		if taken[orphans[i].pi] {
			continue
		}
		for j := i + 1; j < len(orphans); j++ {
			if taken[orphans[j].pi] {
				continue
			}
			ext := orphans[i].mx[0] - orphans[i].mn[0]
			if ext < 0.05 && (mirrored(orphans[i], orphans[j]) || mirroredX(orphans[i], orphans[j])) {
				// eyelash-sized: rides the face
				if head >= 0 {
					d.attach[head] = append(d.attach[head], orphans[i].pi, orphans[j].pi)
				}
				taken[orphans[i].pi], taken[orphans[j].pi] = true, true
				break
			}
			if !mirrored(orphans[i], orphans[j]) {
				continue
			}
			if ext > 0.30 || handDone || teL < 0 || teR < 0 {
				break
			}
			// +z-dominant local bbox = the LEFT hand (capture-pinned)
			l, r := orphans[i], orphans[j]
			if l.mx[2]+l.mn[2] < r.mx[2]+r.mn[2] {
				l, r = r, l
			}
			d.attach[teL] = append(d.attach[teL], l.pi)
			d.attach[teR] = append(d.attach[teR], r.pi)
			taken[l.pi], taken[r.pi] = true, true
			handDone = true
			break
		}
	}
}

// dynIDs enumerates the synthesized dynamics-bone ids.
func dynIDs(base, n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = base + i
	}
	return out
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
				// The ROOT's six streams are ordered tx,ty,tz,rx,ry,rz —
				// the reverse of every other bone's rx,ry,rz,tx,ty,tz.
				// Pinned by the game's own runtime channel structs (the 39
				// bone objects at 0xd04b50, flagman.state/wave400.state):
				// hara_n_x's key counts [27,22,32] land in the walker's
				// TRANSLATE slots and [1,46,1] in ROTATE — its composed
				// matrix carries t=(0.077,1.106,-0.049) = the first triple
				// and yaw 1.378 = the second — while kosi's mask-0x07
				// streams land in ROTATE. The first shipped export read
				// the root's height (~1.19 m) as yaw (~1.19 rad ≈ his real
				// 68° facing — right by coincidence) and his yaw swing as
				// added height: the phantom 2 m "gantry climb".
				slot := bit
				if bi == 0 {
					slot = (bit + 3) % 6
				}
				c.ch[bi][slot] = ch
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
	side map[int]float32 // chain root -> bend sign (chainSide cache)
}

// chainSide returns the two-joint bend sign for a chain (+1 = Z×axis, the
// legs; -1 = axis×Z, the arms). Preference order: the rig's own REST pose
// when the chain rests bent (measured in the rest hinge frame), else the
// joint names — momo/sune/hiza (thigh/shin/knee) vs ude/hiji (arm/elbow) is
// the developers' own vocabulary on all nine skeletons.
func (rg *chrRig) chainSide(root, j1, j2, eff int) float32 {
	if rg.side == nil {
		rg.side = map[int]float32{}
	}
	if s, ok := rg.side[root]; ok {
		return s
	}
	s := float32(0)
	rest := restFK(rg.sk)
	p0 := mPos(rest[root])
	pm := mPos(rest[j2])
	pe := mPos(rest[eff])
	ax := v3sub(pe, p0)
	if l := v3len(ax); l > 1e-4 {
		ax = v3scale(ax, 1/l)
		off := v3sub(v3sub(pm, p0), v3scale(ax, v3dot(v3sub(pm, p0), ax)))
		if v3len(off) > 0.02 {
			chnZ := [3]float32{rest[root][0][2], rest[root][1][2], rest[root][2][2]}
			hz := v3sub(chnZ, v3scale(ax, v3dot(ax, chnZ)))
			if v3len(hz) > 1e-5 {
				hz = v3norm(hz)
				if v3dot(v3norm(off), v3cross(hz, ax)) >= 0 {
					s = 1
				} else {
					s = -1
				}
			}
		}
	}
	if s == 0 {
		n := rg.sk.bones[j1].name + " " + rg.sk.bones[j2].name
		switch {
		case strings.Contains(n, "momo") || strings.Contains(n, "sune") || strings.Contains(n, "hiza"):
			s = 1
		case strings.Contains(n, "ude") || strings.Contains(n, "hiji"):
			s = -1
		default:
			s = 1
		}
	}
	rg.side[root] = s
	return s
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
		// local rest translation (tx from record +0x0C, then the ty/tz
		// pair) — bind-verified: momo under kosi = (0, -0.19, -0.09), the
		// shoulder links carry (-0.0628,-0,-0.067)/(-0.015,0,-0.153), the
		// neck (0.0464,0,0)
		tx, ty, tz := b.restT[0], b.restT[1], b.restT[2]
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
			// animated translations REPLACE the rest offset, as absolute
			// (tx,ty,tz) in the parent frame — for the root that is the
			// character's ground space (the game's root object composes
			// exactly the raw channel triple: height ~1.11-1.23 across the
			// whole wave, feet planted); effector translations are IK
			// TARGETS, not locals — the IK pass consumes them and
			// overwrites the effector transform. Unkeyed components keep
			// their rest value.
			if b.kind != 4 {
				if chs[3] != nil {
					tx = chs[3].eval(frame)
				}
				if chs[4] != nil {
					ty = chs[4].eval(frame)
				}
				if chs[5] != nil {
					tz = chs[5].eval(frame)
				}
			}
		}
		if b.restL != nil {
			// synthesized dynamics bone: fixed bind-relative local
			local[i] = *b.restL
			continue
		}
		R := mEuler(rx, ry, rz, rg.conv.order)
		T := mTrans(tx, ty, tz)
		if rg.conv.rt {
			local[i] = mMul(R, T)
		} else {
			local[i] = mMul(T, R)
		}
		if b.restS != ([3]float32{1, 1, 1}) && b.restS != ([3]float32{}) {
			local[i] = mMul(local[i], mScale(b.restS[0], b.restS[1], b.restS[2]))
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
	// target translations. The targets are ABSOLUTE positions in the
	// character's Y-up ground space — the game's own settled skeleton (the
	// 39 bone objects at 0xd04b50, flagman.state) puts every effector
	// exactly on its channel value (feet on the clip's constant 0.12-high
	// spots, hands/chest/head on their moving targets), with no root
	// riding of any kind.
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

// solveChain poses the chain's joint bones toward the target — the game's
// own construction, read off its settled skeleton (the 39 bone objects in
// flagman.state, each holding the composed matrix the walker wrote):
//
//   axis = normalize(target − chainRootPos)
//   Z    = the chain root's ANIMATED local z axis, Gram-Schmidt
//          orthonormalized against axis — the shared hinge of every joint
//          in the chain (measured z ≡ 0 for the mid-joint offset across
//          700 captured flips of all four limb chains)
//   bend = Z × axis (elbows behind the arm, knees forward — one rule, the
//          chain roots' authored z orientations differ)
//
// Two-joint chains (arms, legs) place the mid joint by law of cosines with
// the bind segment lengths; joint frames are X along their segment, Z the
// shared hinge, Y = Z×X (verified exact against mune/kao/ude/momo/sune
// world matrices). Single-joint chains (chest, head) keep the joint AT the
// chain root — the spine segment stretches: X aims at the target, same Z
// rule; the effector takes the joint's rotation at the target position.
func (rg *chrRig) solveChain(world []mat4, root int, joints []int, eff int, tgt [3]float32) {
	bind := rg.desc.bind
	p0 := mPos(world[root])
	d := v3sub(tgt, p0)
	dl := v3len(d)
	if dl < 1e-5 {
		return
	}
	axis := v3scale(d, 1/dl)
	chnZ := [3]float32{world[root][0][2], world[root][1][2], world[root][2][2]}
	hz := v3sub(chnZ, v3scale(axis, v3dot(axis, chnZ)))
	if v3len(hz) < 1e-5 {
		// degenerate: chain aims along its own hinge; any perpendicular
		hz = v3cross(axis, [3]float32{0, 1, 0})
		if v3len(hz) < 1e-5 {
			hz = v3cross(axis, [3]float32{1, 0, 0})
		}
	}
	hz = v3norm(hz)
	frameZ := func(p, dir [3]float32) mat4 {
		x := v3norm(dir)
		y := v3cross(hz, x)
		var m mat4
		for i := 0; i < 3; i++ {
			m[i][0], m[i][1], m[i][2] = x[i], y[i], hz[i]
		}
		m[0][3], m[1][3], m[2][3] = p[0], p[1], p[2]
		m[3] = [4]float32{0, 0, 0, 1}
		return m
	}
	switch len(joints) {
	case 2:
		j1, j2 := joints[0], joints[1]
		// segment lengths are the rig's own +0x0C bone lengths: the second
		// joint's tx is segment 1 (hiji 0.272, sune 0.390), the effector's
		// tx segment 2 (hiji_eff 0.272, sune_eff 0.440)
		l1 := rg.sk.bones[j2].restT[0]
		l2 := rg.sk.bones[eff].restT[0]
		if l1 <= 0 || l2 <= 0 {
			return
		}
		if dl > l1+l2 {
			dl = l1 + l2
			tgt = v3add(p0, v3scale(axis, dl))
		}
		if min := float32(math.Abs(float64(l1 - l2))); dl < min {
			dl = min + 1e-4
			tgt = v3add(p0, v3scale(axis, dl))
		}
		a := (l1*l1 - l2*l2 + dl*dl) / (2 * dl)
		h2 := l1*l1 - a*a
		if h2 < 0 {
			h2 = 0
		}
		h := float32(math.Sqrt(float64(h2)))
		// Bend side relative to the hinge: legs bend along Z×axis (knees
		// forward), arms along axis×Z (elbows back) — 950/950 captured
		// frames across OZI and OTK, zero plane violations. The side is a
		// per-chain constant; rigs whose REST pose is already bent (the
		// ONN girl sits at rest) carry it in their own geometry, and for
		// straight-rest rigs the joints' names decide — momo/sune/hiza
		// (thigh/shin/knee) vs ude/hiji (arm/elbow) is the developers' own
		// vocabulary on all nine skeletons.
		s := rg.chainSide(root, j1, j2, eff)
		mid := v3add(v3add(p0, v3scale(axis, a)), v3scale(v3cross(hz, axis), s*h))
		world[j1] = frameZ(p0, v3sub(mid, p0))
		world[j2] = frameZ(mid, v3sub(tgt, mid))
		end := v3add(mid, v3scale(v3norm(v3sub(tgt, mid)), l2))
		world[eff] = frameZ(end, v3sub(tgt, mid))
	case 1:
		// chest/head: the joint stays at the chain root; the effector goes
		// its bone length along x-toward-target (the targets are authored
		// at that distance — |target-root| ≈ 0.200/0.210 exactly)
		j1 := joints[0]
		world[j1] = frameZ(p0, d)
		le := rg.sk.bones[eff].restT[0]
		if le <= 0 {
			le = dl
		}
		world[eff] = frameZ(v3add(p0, v3scale(axis, le)), d)
	}
	_ = bind
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

// loadChr loads any /Chr character: CHR_<BASE>.bin descriptor,
// obj_chr_<base>_pmt.sz model, and its bone.bin skeleton. Clips are not
// loaded — see loadClips.
func loadChr(disc *xbox.Image, base string) (*chrData, error) {
	boneRaw, err := disc.ReadFile("/Common/bone.bin")
	if err != nil {
		return nil, err
	}
	sks := parseBoneLib(boneRaw)
	descRaw, err := disc.ReadFile("/Chr/CHR_" + strings.ToUpper(base) + ".bin")
	if err != nil {
		return nil, err
	}
	skelIdx := int(u16(descRaw, 0))
	if skelIdx >= len(sks) {
		return nil, fmt.Errorf("skeleton %d out of range", skelIdx)
	}
	// deep-copy: parseChrDesc appends this character's dynamics bones and
	// wires them into parents' child lists — the library entry must stay
	// pristine for the next character on the same rig
	src := sks[skelIdx]
	sk := &chrSkeleton{name: src.name, bones: make([]chrBone, len(src.bones))}
	copy(sk.bones, src.bones)
	for i := range sk.bones {
		sk.bones[i].children = append([]int(nil), src.bones[i].children...)
	}
	model, err := readPMT(disc, "/Chr/obj_chr_"+base+"_pmt.sz")
	if err != nil {
		return nil, err
	}
	desc, err := parseChrDesc(descRaw, sk, model)
	if err != nil {
		return nil, err
	}
	texs, _, err := model.parseTextures()
	if err != nil {
		return nil, err
	}
	return &chrData{sk: sk, desc: desc, model: model, texs: texs}, nil
}

// loadClips reads the named mot files (motdata_table names like
// "mot_ETC_bin.gz") and returns their clips, named.
func loadClips(disc *xbox.Image, files ...string) ([]*chrClip, error) {
	tblRaw, err := disc.ReadFile("/Common/motdata_table.bin")
	if err != nil {
		return nil, err
	}
	var clips []*chrClip
	for _, mf := range files {
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
	return clips, nil
}

func loadFlagman(disc *xbox.Image) (*chrData, error) {
	cd, err := loadChr(disc, "aut04")
	if err != nil {
		return nil, err
	}
	cd.clips, err = loadClips(disc, "mot_ETC_bin", "mot_OR2SP_ETC_bin")
	if err != nil {
		return nil, err
	}
	return cd, nil
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
	for b, ps := range cd.desc.attach {
		for _, p := range ps {
			part2bone[p] = b
		}
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

// chrCapture reads a MULTI-FRAME bootoracle -carvtx log (with "carvtx: FLIP"
// markers) and prints, per flip, the measured bone transforms RELATIVE to the
// kosi bone: rigid part draws carry S = VP·W_bone in c160-163 (bone-local
// vertices), skinned junction draws carry S = VP·W_A·IBM_A (model-space
// vertices, first LOD bone's skinning matrix — the capture's own vertex
// program pins this: R11 = w·v + (1-w)·M_rel·v then oPos = S·R11), so
// S·bind_A recovers VP·W_A and inv(S_kosi)·S_i cancels VP either way.
// chrDebugChar/chrDebugFiles select which character the -chrcap/-chrcapfit/
// -chrikprobe/-chrpose instruments load (default: the starter).
var chrDebugChar = "aut04"
var chrDebugFiles = []string{"mot_ETC_bin", "mot_OR2SP_ETC_bin"}

func loadDebugChr(disc *xbox.Image) (*chrData, error) {
	cd, err := loadChr(disc, chrDebugChar)
	if err != nil {
		return nil, err
	}
	cd.clips, err = loadClips(disc, chrDebugFiles...)
	if err != nil {
		return nil, err
	}
	return cd, nil
}

func chrCapture(disc *xbox.Image, dumpPath string) {
	cd, err := loadDebugChr(disc)
	if err != nil {
		fatal("%v", err)
	}
	frames := chrCapParse(cd, dumpPath)
	for fi, bm := range frames {
		if bm == nil {
			continue
		}
		var bones []int
		for b := range bm {
			bones = append(bones, b)
		}
		sort.Ints(bones)
		for _, b := range bones {
			rel := bm[b]
			fmt.Printf("cap f%04d %-12s", fi, cd.sk.bones[b].name)
			for r := 0; r < 3; r++ {
				fmt.Printf(" %9.5f %9.5f %9.5f %9.5f", rel[r][0], rel[r][1], rel[r][2], rel[r][3])
			}
			fmt.Println()
		}
	}
}

// chrCapParse reads the log into per-flip kosi-relative bone matrices.
func chrCapParse(cd *chrData, dumpPath string) []map[int]mat4 {
	raw, err := os.ReadFile(dumpPath)
	if err != nil {
		fatal("%v", err)
	}
	// part -> section-B VB span
	type span struct{ off, end, pi int }
	var spans []span
	minOff := 1 << 60
	for pi := 0; pi < cd.model.nParts; pi++ {
		rec := 0x18 + pi*0x3C
		w2 := u32(cd.model.a, rec+4*2)
		w12 := u32(cd.model.a, rec+4*12)
		vbDesc := u32(cd.model.a, int(w2))
		off := int(u32(cd.model.a, int(vbDesc)+4))
		end := off + int(u32(cd.model.a, int(w12)+0x1C))
		spans = append(spans, span{off, end, pi})
		if off < minOff {
			minOff = off
		}
	}
	// bone lookup tables
	part2bone := map[int]int{}   // rigid
	part2lod := map[int]lodRec{} // skinned
	for b, ps := range cd.desc.attach {
		for _, p := range ps {
			part2bone[p] = b
		}
	}
	for _, lr := range cd.desc.lods {
		part2lod[lr.part] = lr
	}
	kosi := -1
	for i, b := range cd.sk.bones {
		if b.name == "kosi" {
			kosi = i
		}
	}

	type draw struct {
		addr int
		m    mat4
	}
	var frames [][]draw
	cur := []draw{}
	minAddr := 1 << 60
	for _, ln := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(ln, "carvtx: FLIP") {
			frames = append(frames, cur)
			cur = nil
			continue
		}
		if !strings.Contains(ln, "DRAW attr0=") {
			continue
		}
		var addr int
		var fmtw string
		if _, err := fmt.Sscanf(ln[strings.Index(ln, "attr0="):], "attr0=%X fmt0=%s", &addr, &fmtw); err != nil {
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
		cur = append(cur, draw{addr, m})
		if addr < minAddr {
			minAddr = addr
		}
	}
	frames = append(frames, cur)
	base := minAddr - minOff
	partOf := func(addr int) int {
		off := addr - base
		for _, s := range spans {
			if off >= s.off && off < s.end {
				return s.pi
			}
		}
		return -1
	}
	out := make([]map[int]mat4, len(frames))
	for fi, fr := range frames {
		if len(fr) == 0 {
			continue
		}
		// measured VP·W per bone (first draw wins)
		bm := map[int]mat4{}
		for _, d := range fr {
			pi := partOf(d.addr)
			if pi < 0 {
				continue
			}
			if b, ok := part2bone[pi]; ok {
				if _, seen := bm[b]; !seen {
					bm[b] = d.m
				}
			} else if lr, ok := part2lod[pi]; ok {
				b := lr.bones[0]
				if _, seen := bm[b]; !seen {
					bm[b] = mMul(d.m, cd.desc.bind[b])
				}
			}
		}
		ref, ok := bm[kosi]
		if !ok {
			continue
		}
		refInv := mInvFull(ref)
		rel := map[int]mat4{}
		for b, m := range bm {
			rel[b] = mMul(refInv, m)
		}
		out[fi] = rel
	}
	return out
}

// chrCapFit fits every captured flip against the clip library under the
// export's pose conventions: tracks the (clip, frame) phase across flips
// (full search on the first, ±12-frame window after) and reports per-bone
// position residuals — the instrument that says WHICH bones the evaluator
// gets wrong, not just that it is wrong somewhere.
func chrCapFit(disc *xbox.Image, dumpPath string, conv poseConv) {
	cd, err := loadDebugChr(disc)
	if err != nil {
		fatal("%v", err)
	}
	caps := chrCapParse(cd, dumpPath)
	rg := &chrRig{sk: cd.sk, desc: cd.desc, conv: conv}
	kosi := -1
	for i, b := range cd.sk.bones {
		if b.name == "kosi" {
			kosi = i
		}
	}
	relPred := func(clip *chrClip, f float32) map[int]mat4 {
		world := rg.evalPose(clip, f)
		inv := mInv(world[kosi])
		out := map[int]mat4{}
		for b := range cd.desc.bind {
			out[b] = mMul(inv, world[b])
		}
		return out
	}
	score := func(pred, meas map[int]mat4) float64 {
		var e float64
		for b, mm := range meas {
			pm, ok := pred[b]
			if !ok {
				continue
			}
			for i := 0; i < 3; i++ {
				e += math.Abs(float64(pm[i][3] - mm[i][3]))
			}
		}
		return e
	}
	type phase struct {
		clip *chrClip
		f    float32
	}
	var cur phase
	sumBone := map[int]float64{}
	cntBone := map[int]int{}
	for fi, meas := range caps {
		if meas == nil {
			continue
		}
		best := phase{}
		bestE := math.Inf(1)
		if cur.clip == nil {
			for _, c := range cd.clips {
				if c.name == "" {
					continue
				}
				for f := 0; f < c.frames; f++ {
					e := score(relPred(c, float32(f)), meas)
					if e < bestE {
						bestE, best = e, phase{c, float32(f)}
					}
				}
			}
		} else {
			for _, c := range cd.clips {
				lo, hi := 0, c.frames
				if c == cur.clip {
					lo, hi = int(cur.f)-2, int(cur.f)+14
					if lo < 0 {
						lo = 0
					}
					if hi > c.frames {
						hi = c.frames
					}
				} else if c.name == "" {
					continue
				} else {
					// other clips: only their head (a transition starts at 0)
					hi = 20
				}
				for f := lo; f < hi; f++ {
					e := score(relPred(c, float32(f)), meas)
					if e < bestE {
						bestE, best = e, phase{c, float32(f)}
					}
				}
			}
		}
		cur = best
		pred := relPred(best.clip, best.f)
		fmt.Printf("fit f%04d clip=%-28s cf=%5g err=%.3f\n", fi, best.clip.name, best.f, bestE)
		for b, mm := range meas {
			if pm, ok := pred[b]; ok {
				d := v3len(v3sub(mPos(pm), mPos(mm)))
				sumBone[b] += float64(d)
				cntBone[b]++
			}
		}
	}
	fmt.Println("mean |Δt| per bone:")
	var bones []int
	for b := range sumBone {
		bones = append(bones, b)
	}
	sort.Ints(bones)
	for _, b := range bones {
		fmt.Printf("  %-12s %.4f m over %d frames\n", cd.sk.bones[b].name, sumBone[b]/float64(cntBone[b]), cntBone[b])
	}
}

// restFK composes the skeleton's rest pose (the bind pose): local = T·R
// with the full rest translation (tx@+0x0C, f0, f1) and the ZYX rest euler —
// the same conventions the animated evaluator uses with no channels applied.
func restFK(sk *chrSkeleton) []mat4 {
	world := make([]mat4, len(sk.bones))
	var walk func(i int, parent mat4)
	walk = func(i int, parent mat4) {
		b := sk.bones[i]
		var local mat4
		if b.restL != nil {
			local = *b.restL
		} else {
			local = mMul(mTrans(b.restT[0], b.restT[1], b.restT[2]),
				mEuler(b.restR[0], b.restR[1], b.restR[2], 5))
			if b.restS != ([3]float32{1, 1, 1}) && b.restS != ([3]float32{}) {
				local = mMul(local, mScale(b.restS[0], b.restS[1], b.restS[2]))
			}
		}
		world[i] = mMul(parent, local)
		for _, c := range b.children {
			walk(c, world[i])
		}
	}
	for i, b := range sk.bones {
		if b.parent == -1 {
			walk(i, mIdent())
		}
	}
	return world
}

// chrRestCheck validates the rest-FK bind hypothesis for one character: every
// file inverse-bind matrix must match inv(restFK world) of one LOD-union bone.
func chrRestCheck(disc *xbox.Image, base string) {
	cd, err := loadChr(disc, base)
	if err != nil {
		fatal("%v", err)
	}
	rest := restFK(cd.sk)
	var bones []int
	for b := range cd.desc.lodBone {
		bones = append(bones, b)
	}
	sort.Ints(bones)
	worst := float32(0)
	for _, b := range bones {
		fb, ok := cd.desc.bind[b]
		if !ok {
			fmt.Printf("  %-16s NO FILE BIND\n", cd.sk.bones[b].name)
			continue
		}
		var d float32
		for r := 0; r < 3; r++ {
			for c := 0; c < 4; c++ {
				v := fb[r][c] - rest[b][r][c]
				if v < 0 {
					v = -v
				}
				if v > d {
					d = v
				}
			}
		}
		if d > worst {
			worst = d
		}
		if d > 0.005 {
			fp, rp := mPos(fb), mPos(rest[b])
			fmt.Printf("  %-16s maxΔ=%.4f file t(%6.3f %6.3f %6.3f) rest t(%6.3f %6.3f %6.3f)\n",
				cd.sk.bones[b].name, d, fp[0], fp[1], fp[2], rp[0], rp[1], rp[2])
		}
	}
	fmt.Printf("%s: %d LOD bones, worst bind-vs-restFK maxΔ = %.5f\n", base, len(bones), worst)
}

// chrIKProbe compares one captured flip against one clip frame chain by
// chain: measured joint positions vs FK/IK prediction vs the clip's raw
// effector targets, everything expressed in the kosi frame.
func chrIKProbe(disc *xbox.Image, spec string, conv poseConv) {
	var flip, cf int
	var clipName string
	if _, err := fmt.Sscanf(spec, "%d:%s", &flip, &clipName); err != nil {
		fatal("want FLIP:CLIP:FRAME, got %q (%v)", spec, err)
	}
	if i := strings.LastIndex(clipName, ":"); i > 0 {
		fmt.Sscanf(clipName[i+1:], "%d", &cf)
		clipName = clipName[:i]
	}
	cd, err := loadDebugChr(disc)
	if err != nil {
		fatal("%v", err)
	}
	caps := chrCapParse(cd, os.Getenv("CHRCAP"))
	if flip >= len(caps) || caps[flip] == nil {
		fatal("flip %d not in capture", flip)
	}
	meas := caps[flip]
	clip := cd.clip(clipName)
	if clip == nil {
		fatal("clip %q?", clipName)
	}
	rg := &chrRig{sk: cd.sk, desc: cd.desc, conv: conv}
	world := rg.evalPose(clip, float32(cf))
	kosi := byNameBone(cd.sk, "kosi")
	invK := mInv(world[kosi])
	fmt.Printf("flip %d vs %s:%d\n", flip, clipName, cf)
	for ci, b := range cd.sk.bones {
		if b.kind != 2 {
			continue
		}
		joints, eff := rg.chainOf(ci)
		fmt.Printf("chain %s (joints %d, eff %s):\n", b.name, len(joints), nameOr(cd.sk, eff))
		for _, j := range append(append([]int{}, joints...), eff) {
			if j < 0 {
				continue
			}
			pred := mMul(invK, world[j])
			fmt.Printf("  %-12s pred t (%7.3f %7.3f %7.3f)", cd.sk.bones[j].name,
				pred[0][3], pred[1][3], pred[2][3])
			if mm, ok := meas[j]; ok {
				fmt.Printf("  meas t (%7.3f %7.3f %7.3f)  |Δ|=%.3f",
					mm[0][3], mm[1][3], mm[2][3],
					v3len(v3sub(mPos(pred), mPos(mm))))
			}
			fmt.Println()
		}
		// raw effector target in character space -> kosi frame
		if eff >= 0 {
			if chs := clip.ch[eff]; chs != nil && (chs[3] != nil || chs[4] != nil || chs[5] != nil) {
				var tgt [3]float32
				for k := 0; k < 3; k++ {
					if chs[3+k] != nil {
						tgt[k] = chs[3+k].eval(float32(cf))
					}
				}
				t4 := [4]float32{tgt[0], tgt[1], tgt[2], 1}
				var tk [3]float32
				for r := 0; r < 3; r++ {
					tk[r] = invK[r][0]*t4[0] + invK[r][1]*t4[1] + invK[r][2]*t4[2] + invK[r][3]
				}
				fmt.Printf("  target raw (%7.3f %7.3f %7.3f)  in-kosi (%7.3f %7.3f %7.3f)\n",
					tgt[0], tgt[1], tgt[2], tk[0], tk[1], tk[2])
			}
		}
	}
	// orientation detail: measured rotation rows of the arm/leg joints
	for _, nm := range []string{"ude_l_jnt", "hiji_l_jnt", "momo_l_jnt", "sune_l_jnt"} {
		j := byNameBone(cd.sk, nm)
		if mm, ok := meas[j]; ok {
			fmt.Printf("meas %-12s rot rows (%6.3f %6.3f %6.3f | %6.3f %6.3f %6.3f | %6.3f %6.3f %6.3f)\n",
				nm, mm[0][0], mm[0][1], mm[0][2], mm[1][0], mm[1][1], mm[1][2], mm[2][0], mm[2][1], mm[2][2])
		}
	}
}

func byNameBone(sk *chrSkeleton, name string) int {
	for i, b := range sk.bones {
		if b.name == name {
			return i
		}
	}
	return -1
}

func nameOr(sk *chrSkeleton, i int) string {
	if i < 0 || i >= len(sk.bones) {
		return "-"
	}
	return sk.bones[i].name
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

// chrPartPrims builds the glb primitives of one character part, textured per
// material like the car path. Rigid parts carry bone-local vertices; skinned
// junction parts (weights != nil after decode) carry T-pose model-space
// vertices, and joints/weights build glTF skinning attributes over jointOf
// (LOD-record bone list order; the last listed bone takes the 1-Σ remainder).
// Batches whose referenced vertices all coincide are dropped: part 31 (the
// left hand) ships 16 checkered-flag verts collapsed to a point — the game
// draws them degenerate at the start line (live VB = file bytes; he never
// holds a flag there), so the export leaves them out.
func chrPartPrims(model *pmt, texs []texInfo, pi int, lodBones []int, jointOf map[int]uint8) ([]glb.Prim, error) {
	pt, err := model.parsePart(pi)
	if err != nil {
		return nil, err
	}
	var pos [][3]float32
	var nrm [][3]float32
	var uvs [][2]float32
	var joints [][4]uint8
	var weights [][4]float32
	vbase := make([]uint32, len(pt.pairs))
	for k, bp := range pt.pairs {
		vbase[k] = uint32(len(pos))
		p2, n2, u2, _, _ := model.decodeVerts(bp)
		pos = append(pos, p2...)
		nrm = append(nrm, n2...)
		uvs = append(uvs, u2...)
		if ws := model.decodeWeights(bp); ws != nil {
			if len(lodBones) != len(ws[0])+1 {
				return nil, fmt.Errorf("part %d: %d stored weights for %d LOD bones", pi, len(ws[0]), len(lodBones))
			}
			var j4 [4]uint8
			for k, bi := range lodBones {
				j4[k] = jointOf[bi]
			}
			for _, w := range ws {
				var w4 [4]float32
				rest := float32(1)
				for k, v := range w {
					w4[k] = v
					rest -= v
				}
				w4[len(w)] = rest
				joints = append(joints, j4)
				weights = append(weights, w4)
			}
		} else {
			for range p2 {
				joints = append(joints, [4]uint8{})
				weights = append(weights, [4]float32{1})
			}
		}
	}
	skinned := len(lodBones) > 0
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
		if degenerateBatch(pos, raw) {
			continue
		}
		if !skinned && detachedBatch(pos, raw) {
			// part 31 (left hand) carries a furled checkered-flag remnant
			// 1.3-1.5 m out along the hand's z — its own texture, drawn by
			// nothing in the live start-line frames (the whole VB is
			// byte-identical to the file, and no geometry appears at its
			// spot in any captured flip). A rigid part's batch that sits
			// entirely a limb's length from its own bone is that remnant.
			continue
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
		if skinned {
			pr.JointsW = joints
			pr.Weights = weights
		}
		if ti >= 0 {
			pr.Image = texs[ti].img
		}
		out = append(out, pr)
	}
	return out, nil
}

// detachedBatch reports whether every vertex the index list touches lies more
// than half a metre from the part's bone origin — detached remnant geometry
// on an otherwise bone-hugging rigid part.
func detachedBatch(pos [][3]float32, raw []uint32) bool {
	for _, ix := range raw {
		if v3dot(pos[ix], pos[ix]) < 0.25 {
			return false
		}
	}
	return len(raw) > 0
}

// degenerateBatch reports whether every vertex the index list touches sits at
// (numerically) one point — a zero-area batch that can never shade a pixel.
func degenerateBatch(pos [][3]float32, raw []uint32) bool {
	if len(raw) == 0 {
		return true
	}
	p0 := pos[raw[0]]
	for _, ix := range raw[1:] {
		d := v3sub(pos[ix], p0)
		if v3dot(d, d) > 1e-6 {
			return false
		}
	}
	return true
}

// chrBake names one clip to bake into a character GLB.
type chrBake struct {
	clip string // full mot clip name
	as   string // glTF animation name
}

// chrExportGLB writes the animated flagman GLB (kept for the -chrglb debug
// flow; the roster export goes through exportChrGLB directly).
func chrExportGLB(disc *xbox.Image, outPath string, conv poseConv) error {
	cd, err := loadFlagman(disc)
	if err != nil {
		return err
	}
	// idle = the plain stand loop; the wave = HATAFURI_00, the clip the
	// game actually plays at the start line (the 700-flip capture tracks
	// it monotonically at one clip frame per flip; HATA_SP, shipped
	// before, is a different wave he never performs there — and before
	// the wave starts he simply holds RUNAWAY frame 0, err 0.002 m).
	return exportChrGLB(cd, "flagman", outPath, conv, []chrBake{
		{"ORT_OZI_OZI_STAND_LP", "idle"},
		{"ORT_OZI_OZI_HATAFURI_00", "hatafuri"},
	})
}

// exportChrGLB writes one character's animated GLB. The geometry-bone nodes
// form a HIERARCHY (hip → thigh → shin → foot, chest → arm chain …) and the
// clips carry per-node LOCAL TRS — interpolating locals keeps every limb
// attached between keyframes, which a flat world-space bake cannot (parts
// lerp along straight lines and a fast sweep tears the chain apart
// mid-interval; the first shipped bake did exactly that).
func exportChrGLB(cd *chrData, rootName, outPath string, conv poseConv, bakes []chrBake) error {
	rg := &chrRig{sk: cd.sk, desc: cd.desc, conv: conv}

	// geometry bones parented by nearest bind-carrying ancestor (the 17-bone
	// LOD union: the 15 rigid attach bones + the kata_l/kata_r shoulder bones
	// that only skinned junction parts reference)
	byName := map[string]int{}
	for i, b := range cd.sk.bones {
		byName[b.name] = i
	}
	geomParent := func(bone int) int {
		p := cd.sk.bones[bone].parent
		for p >= 0 {
			if _, ok := cd.desc.bind[p]; ok {
				return p
			}
			p = cd.sk.bones[p].parent
		}
		return -1
	}

	// all bind bones sorted for stable node order; parents before children
	var boneList []int
	for b := range cd.desc.bind {
		boneList = append(boneList, b)
	}
	sort.Ints(boneList)
	var order []int
	added := map[int]bool{}
	for len(order) < len(boneList) {
		for _, b := range boneList {
			if added[b] {
				continue
			}
			if gp := geomParent(b); gp == -1 || added[gp] {
				order = append(order, b)
				added[b] = true
			}
		}
	}

	s := glb.NewScene()
	root := s.AddNode(rootName, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
	nodes := map[int]int{} // bone -> node
	for _, b := range order {
		bind := cd.desc.bind[b]
		local := bind
		parentNode := root
		if gp := geomParent(b); gp >= 0 {
			local = mMul(mInv(cd.desc.bind[gp]), bind)
			parentNode = nodes[gp]
		}
		name := cd.sk.bones[b].name
		lt, lq, ls := mDecompose(local)
		n := s.AddNode(name, parentNode, lt, lq, ls)
		if parts, ok := cd.desc.attach[b]; ok {
			var prims []glb.Prim
			for _, part := range parts {
				pp, err := chrPartPrims(cd.model, cd.texs, part, nil, nil)
				if err != nil {
					return err
				}
				prims = append(prims, pp...)
			}
			if err := s.AddMesh(n, name, prims); err != nil {
				return err
			}
		}
		nodes[b] = n
	}

	// the skinned junction parts: one skin over the bind-bone union, meshes
	// under the root (a skinned glTF mesh takes its transforms from the
	// joints alone). This is the game's own pipeline: the capture's skinning
	// vertex program blends w·v + (1-w)·(M_rel·v) inside the first bone's
	// skinning frame — linear blend skinning over W_bone·invBind_bone.
	jointOf := map[int]uint8{}
	var jointNodes []int
	var ibms [][16]float32
	for i, b := range order {
		jointOf[b] = uint8(i)
		jointNodes = append(jointNodes, nodes[b])
		inv := cd.desc.invBind[b]
		var m16 [16]float32
		for c := 0; c < 4; c++ {
			for r := 0; r < 4; r++ {
				m16[c*4+r] = inv[r][c]
			}
		}
		ibms = append(ibms, m16)
	}
	skin := s.AddSkin(jointNodes, ibms)
	for _, lr := range cd.desc.lods {
		prims, err := chrPartPrims(cd.model, cd.texs, lr.part, lr.bones, jointOf)
		if err != nil {
			return err
		}
		name := fmt.Sprintf("skin_%d", lr.part)
		n := s.AddNode(name, root, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
		if err := s.AddMesh(n, name, prims); err != nil {
			return err
		}
		s.SetNodeSkin(n, skin)
	}

	const fps = 60.0
	const step = 2 // sample every 2 frames (30 Hz; locals interpolate cleanly)
	bake := func(clipName, asName string) error {
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
		// iterate bones in the stable node order — a map walk here writes
		// the animation channels in random order and churns the GLB bytes
		// on every re-export
		rot := map[int][][4]float32{}
		trn := map[int][][3]float32{}
		scl := map[int][][3]float32{}
		sclLive := map[int]bool{}
		for _, world := range worlds {
			for _, b := range order {
				m := world[b]
				if gp := geomParent(b); gp >= 0 {
					// full inverse: twin bones carry non-unit rest scale
					m = mMul(mInvFull(world[gp]), world[b])
				}
				t, q, sv := mDecompose(m)
				rot[b] = append(rot[b], q)
				trn[b] = append(trn[b], t)
				scl[b] = append(scl[b], sv)
				if sv[0] < 0.999 || sv[0] > 1.001 || sv[1] < 0.999 || sv[1] > 1.001 || sv[2] < 0.999 || sv[2] > 1.001 {
					sclLive[b] = true
				}
			}
		}
		for _, b := range order {
			n := nodes[b]
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
			if sclLive[b] {
				c.Vec3s(n, "scale", times, scl[b])
			}
		}
		c.Finish()
		return nil
	}
	for _, bk := range bakes {
		if err := bake(bk.clip, bk.as); err != nil {
			return err
		}
	}
	return s.Write(outPath, rootName)
}

// chrPoseDebug prints key bone world positions of a clip at given frames.
func chrPoseDebug(disc *xbox.Image, spec string, conv poseConv) {
	cd, err := loadDebugChr(disc)
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

// chrBindGLBs writes bind-pose GLBs (no clips) for the named characters —
// the identification pass: who is who in /Chr.
func chrBindGLBs(disc *xbox.Image, list, dir string) {
	for _, base := range strings.Split(list, ",") {
		cd, err := loadChr(disc, base)
		if err != nil {
			fatal("%s: %v", base, err)
		}
		out := dir + "/bind-" + base + ".glb"
		if err := exportChrGLB(cd, base, out, poseConv{5, false}, nil); err != nil {
			fatal("%s: %v", base, err)
		}
		fmt.Println("wrote", out)
	}
}

// chrSpec curates one /Chr character for the Studio.
type chrSpec struct {
	base  string // obj_chr_<base>_pmt.sz / CHR_<BASE>.bin
	id    string // asset id
	name  string // display name
	files []string
	bakes []chrBake
	anims []schema.Animation
}

// The /Chr roster. Skeletons: dr_m*/s00/w00 sit on OTK or its twin MAN
// (identical bone names + topology), dr_g*/l*/h00 on ONN or its twin WMN,
// fal/gal on the 113-bone JEN, mal on the 115-bone RIC. The in-car pair's
// sitting loops are mot_START's two SUWARI clips; WINNER/TUNTUN come from
// the goal celebrations (mot_F40).
var chrRoster = []chrSpec{
	{base: "dr_m00", id: "driver", name: "The driver",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_MAN_OTK_SUWARI_LP_01", "sit"}, {"ORT_MAN_OTK_WINNER_01", "winner"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "winner", Name: "Winner", Loop: "once", Clip: "winner"},
		}},
	{base: "dr_mh00", id: "driver-cap", name: "The driver (capped)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_MAN_OTK_SUWARI_LP_01", "sit"}, {"ORT_MAN_OTK_WINNER_01", "winner"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "winner", Name: "Winner", Loop: "once", Clip: "winner"},
		}},
	{base: "dr_g00", id: "passenger", name: "The passenger",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_WMN_ONN_SUWARI_LP_01", "sit"}, {"ORT_WMN_ONN_TUNTUN_01", "tuntun"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "tuntun", Name: "Impatient", Loop: "once", Clip: "tuntun"},
		}},
	{base: "dr_gh00", id: "passenger-hat", name: "The passenger (hat)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_WMN_ONN_SUWARI_LP_01", "sit"}, {"ORT_WMN_ONN_TUNTUN_01", "tuntun"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "tuntun", Name: "Impatient", Loop: "once", Clip: "tuntun"},
		}},
	{base: "dr_g00_usa", id: "passenger-usa", name: "The passenger (US)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_WMN_ONN_SUWARI_LP_01", "sit"}, {"ORT_WMN_ONN_TUNTUN_01", "tuntun"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "tuntun", Name: "Impatient", Loop: "once", Clip: "tuntun"},
		}},
	{base: "dr_gh00_usa", id: "passenger-hat-usa", name: "The passenger (hat, US)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_WMN_ONN_SUWARI_LP_01", "sit"}, {"ORT_WMN_ONN_TUNTUN_01", "tuntun"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "tuntun", Name: "Impatient", Loop: "once", Clip: "tuntun"},
		}},
	{base: "dr_l00", id: "passenger-skirt", name: "The passenger (skirt)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_WMN_ONN_SUWARI_LP_01", "sit"}, {"ORT_WMN_ONN_TUNTUN_01", "tuntun"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "tuntun", Name: "Impatient", Loop: "once", Clip: "tuntun"},
		}},
	{base: "dr_lh00", id: "passenger-skirt-hat", name: "The passenger (skirt, hat)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_WMN_ONN_SUWARI_LP_01", "sit"}, {"ORT_WMN_ONN_TUNTUN_01", "tuntun"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "tuntun", Name: "Impatient", Loop: "once", Clip: "tuntun"},
		}},
	{base: "dr_h00", id: "passenger-black", name: "The passenger (black)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_WMN_ONN_SUWARI_LP_01", "sit"}, {"ORT_WMN_ONN_TUNTUN_01", "tuntun"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "tuntun", Name: "Impatient", Loop: "once", Clip: "tuntun"},
		}},
	{base: "dr_s00", id: "driver-shirt", name: "The driver (shirt)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_MAN_OTK_SUWARI_LP_01", "sit"}, {"ORT_MAN_OTK_WINNER_01", "winner"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "winner", Name: "Winner", Loop: "once", Clip: "winner"},
		}},
	{base: "dr_w00", id: "driver-dark", name: "The driver (dark)",
		files: []string{"mot_START_bin", "mot_F40_bin"},
		bakes: []chrBake{{"ORT_MAN_OTK_SUWARI_LP_01", "sit"}, {"ORT_MAN_OTK_WINNER_01", "winner"}},
		anims: []schema.Animation{
			{ID: "sit", Name: "In the seat", Loop: "loop", Clip: "sit"},
			{ID: "winner", Name: "Winner", Loop: "once", Clip: "winner"},
		}},
	{base: "mal", id: "driver-story", name: "The driver (story scenes)",
		files: []string{"mot_E16_bin", "mot_E01_bin"},
		bakes: []chrBake{{"ORT_MAL_RIC_E01_1", "scene1"}, {"ORT_MAL_RIC_E16_2_1", "scene16"}},
		anims: []schema.Animation{
			{ID: "scene1", Name: "Story scene 1", Loop: "loop", Clip: "scene1"},
			{ID: "scene16", Name: "Story scene 16", Loop: "once", Clip: "scene16"},
		}},
	{base: "gal", id: "passenger-story", name: "The passenger (story scenes)",
		files: []string{"mot_E16_bin", "mot_E01_bin"},
		bakes: []chrBake{{"ORT_FAL_JEN_E01_1", "scene1"}, {"ORT_FAL_JEN_E16_2_1", "scene16"}},
		anims: []schema.Animation{
			{ID: "scene1", Name: "Story scene 1", Loop: "loop", Clip: "scene1"},
			{ID: "scene16", Name: "Story scene 16", Loop: "once", Clip: "scene16"},
		}},
	{base: "gal_usa", id: "passenger-story-usa", name: "The passenger (story scenes, US)",
		files: []string{"mot_E16_bin", "mot_E01_bin"},
		bakes: []chrBake{{"ORT_FAL_JEN_E01_1", "scene1"}, {"ORT_FAL_JEN_E16_2_1", "scene16"}},
		anims: []schema.Animation{
			{ID: "scene1", Name: "Story scene 1", Loop: "loop", Clip: "scene1"},
			{ID: "scene16", Name: "Story scene 16", Loop: "once", Clip: "scene16"},
		}},
	{base: "fal", id: "startgirl", name: "The start girl",
		files: []string{"mot_E01_bin", "mot_E16_bin"},
		bakes: []chrBake{{"ORT_FAL_JEN_TEST", "dance"}, {"ORT_FAL_JEN_E16_2_1", "scene16"}},
		anims: []schema.Animation{
			{ID: "dance", Name: "Dance (the rig's test clip)", Loop: "loop", Clip: "dance"},
			{ID: "scene16", Name: "Story scene 16", Loop: "once", Clip: "scene16"},
		}},
	{base: "aut04_cvt", id: "starter-cvt", name: "The starter (alternate)",
		files: []string{"mot_ETC_bin"},
		bakes: []chrBake{{"ORT_OZI_OZI_STAND_LP", "idle"}, {"ORT_OZI_OZI_HATAFURI_00", "hatafuri"}},
		anims: []schema.Animation{
			{ID: "idle", Name: "Idle", Loop: "loop", Clip: "idle"},
			{ID: "hatafuri", Name: "Start wave", Loop: "once", Clip: "hatafuri"},
		}},
}

// exportCharacters writes the whole /Chr roster into the Studio site.
func exportCharacters(disc *xbox.Image, b *build.Builder) error {
	for _, spec := range chrRoster {
		cd, err := loadChr(disc, spec.base)
		if err != nil {
			return fmt.Errorf("%s: %w", spec.base, err)
		}
		cd.clips, err = loadClips(disc, spec.files...)
		if err != nil {
			return fmt.Errorf("%s: %w", spec.base, err)
		}
		out, err := b.Path("objects", spec.id+".glb")
		if err != nil {
			return err
		}
		if err := exportChrGLB(cd, spec.id, out, poseConv{5, false}, spec.bakes); err != nil {
			return fmt.Errorf("%s: %w", spec.base, err)
		}
		b.AddObject(schema.Asset{ID: spec.id, Name: spec.name, Group: "Characters"},
			&schema.Object{Type: schema.ObjectModel3D, Name: spec.name, Model: spec.id + ".glb",
				SkinnedClone: true,
				Animations:   spec.anims,
				Props: map[string]any{
					"source":   "/Chr/obj_chr_" + spec.base + "_pmt.sz",
					"skeleton": cd.sk.name + " (/Common/bone.bin)",
				}})
	}
	return nil
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
	// SkinnedClone: the junction parts are glTF-skinned, and the Studio's
	// ObjectLibrary instantiates placements via plain Object3D.clone unless
	// the doc asks for SkeletonUtils.clone — a plain clone's skinned meshes
	// stay bound to the PROTO scene's bones, whose world matrices never
	// update, so every skinned vertex renders at Σw·IBM·v: a bone-local
	// pile at the placement origin while the rigid parts animate above it.
	b.AddObject(schema.Asset{ID: "flagman", Name: "The starter", Group: "Characters"},
		&schema.Object{Type: schema.ObjectModel3D, Name: "The starter", Model: "flagman.glb",
			SkinnedClone: true,
			Animations: []schema.Animation{
				{ID: "idle", Name: "Idle", Loop: "loop", Clip: "idle"},
				{ID: "hatafuri", Name: "Start wave", Loop: "once", Clip: "hatafuri"},
			},
			Props: map[string]any{"source": "/Chr/obj_chr_aut04_pmt.sz", "skeleton": "OZI (/Common/bone.bin)", "clips": "mot_ETC_bin.sz (STAND_LP, HATAFURI_00)"}})
	return nil
}
