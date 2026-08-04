// Package merc parses merc-ctrl skinned-mesh objects (the VU1 "merc"
// renderer's model format).
//
// The layout is read from the engine's own code:
//
//   - draw-bones-merc (0x6EC794) walks the control and emits the DMA chain,
//     naming every field it touches. The effect table starts at ctrl+108,
//     one 32-byte merc-effect record per effect (count at basic+52):
//     {+0 →fragment geometry stream, +4 →fragment-control stream, +16 ?,
//     +18 fragment count, +22 tri count, +24 dvert count, +26 flags…}.
//     Per-effect stats are summed into *merc-global-stats* from +18/+22/+24.
//   - A merc-fragment-control record (type size 4 + trailing pairs) is
//     {u8 unsigned-4x-qwc, u8 lump-4x-qwc, u8 fp-qwc, u8 matrix-count},
//     then matrix-count × {u8 palette-index, u8 vu-dest}. The emitter turns
//     it into: UNPACK V4-8 of ceil(u/4)… — precisely: the byte counts are
//     NUM values whose quadword footprint in the source stream is
//     (num+3)>>2 for the two V4-8 unpacks (16 source bytes per 4 rows) and
//     num for the V4-32 fp part; the three parts are consecutive in the
//     geometry stream, unpacked to TOP+140 onward; each bone matrix is a
//     7-quadword ref from a 128-byte-stride palette; MSCAL ends the
//     fragment.
//   - merc-vu1-initialize-chain (0x6E9CF4): STCYCL cl=4 wl=4, STMOD 0,
//     double-buffer BASE 0x1BA, and an 8-qw V4-32 unpack of constants to
//     low-mem 0; the microprogram (merc-vu1-block, disassembled with
//     tools/cmd/disvu) reads its control header at TOP+140 (the V4-8
//     expansion of the fragment's merc-byte-header) and builds GIF packets
//     at TOP+371.
//
// Link the object at base 0 (goalobj.Link(data, 0, symtab)) so pointers are
// object offsets.
package merc

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Fragment is one VU1 batch: three consecutive source regions plus its bone
// matrix uploads.
type Fragment struct {
	Off      uint32 // object offset of ByteData (for patching a live image)
	ByteData []byte // V4-8 region: merc-byte-header + packed vertex bytes
	LumpData []byte // V4-8 region: "lump" packed vertex bytes
	FPData   []byte // V4-32 region: float data (merc-fp-header + floats)
	ByteQWC  int    // NUM of the first unpack (rows of 4 bytes)
	LumpQWC  int    // NUM of the second unpack
	FPQWC    int    // quadwords of float data
	Mats     []MatXfer
	STRow    uint32 // ctrl+44: the VIF row magic added to the s/t bytes
	STOff    uint32 // ctrl+32: the vf19 row the microcode adds after
}

// MatXfer is one bone-matrix upload: palette slot -> VU address.
type MatXfer struct {
	Index, Dest byte
}

// Effect is one merc-effect: a run of fragments sharing render state.
type Effect struct {
	FragCount int
	TriCount  int
	DVertCount int
	// Flags is the record's u16 at +26. Bit 8 marks the effects draw-bones
	// routes through the second (alpha) bucket — eyes, armor, hair: their
	// fragments write their own TEST (ATE off) and ALPHA registers, so
	// texture alpha is not coverage there.
	Flags     int
	ExtraInfo uint32 // record +28: →extra-info block (0 = none)
	Fragments []Fragment
}

// Ctrl is a parsed merc-ctrl.
type Ctrl struct {
	Offset  uint32 // object offset of the basic (type word at Offset-4)
	Effects []Effect
}

func u16(b []byte, o uint32) uint16 { return binary.LittleEndian.Uint16(b[o:]) }
func u32(b []byte, o uint32) uint32 { return binary.LittleEndian.Uint32(b[o:]) }

// Parse reads the merc-ctrl whose basic pointer (one past the type word) is
// at object offset p in the base-0-linked object image.
func Parse(obj []byte, p uint32) (*Ctrl, error) {
	if int64(p)+0x80 > int64(len(obj)) {
		return nil, fmt.Errorf("merc: ctrl at 0x%x outside object", p)
	}
	c := &Ctrl{Offset: p}
	effCount := int(u32(obj, p+52)) // draw-bones-merc: lwu t8, 52(t7)
	// The effect table begins at basic+108 (draw-bones-merc: t9 = t7+108
	// with t7 the basic pointer).
	et := p + 108
	for e := 0; e < effCount; e++ {
		rec := et + uint32(e)*32
		if int64(rec)+32 > int64(len(obj)) {
			return nil, fmt.Errorf("merc: effect %d record outside object", e)
		}
		geo := u32(obj, rec)
		fc := u32(obj, rec+4)
		eff := Effect{
			FragCount:  int(u16(obj, rec+18)),
			TriCount:   int(u16(obj, rec+22)),
			DVertCount: int(u16(obj, rec+24)),
			Flags:      int(u16(obj, rec+26)),
			ExtraInfo:  u32(obj, rec+28),
		}
		gcur, ccur := geo, fc
		for f := 0; f < eff.FragCount; f++ {
			if int64(ccur)+4 > int64(len(obj)) {
				return nil, fmt.Errorf("merc: effect %d frag %d control outside object", e, f)
			}
			uNum := int(obj[ccur])
			lNum := int(obj[ccur+1])
			fpQW := int(obj[ccur+2])
			mats := int(obj[ccur+3])
			ccur += 4
			var xfers []MatXfer
			for m := 0; m < mats; m++ {
				xfers = append(xfers, MatXfer{obj[ccur], obj[ccur+1]})
				ccur += 2
			}
			uQW := (uNum + 3) >> 2
			lQW := (lNum + 3) >> 2
			need := int64(uQW+lQW+fpQW) * 16
			if int64(gcur)+need > int64(len(obj)) {
				return nil, fmt.Errorf("merc: effect %d frag %d data runs past object (0x%x + 0x%x)", e, f, gcur, need)
			}
			fr := Fragment{
				Off:      gcur,
				ByteData: obj[gcur : gcur+uint32(uQW)*16],
				LumpData: obj[gcur+uint32(uQW)*16 : gcur+uint32(uQW+lQW)*16],
				FPData:   obj[gcur+uint32(uQW+lQW)*16 : gcur+uint32(uQW+lQW+fpQW)*16],
				ByteQWC:  uNum, LumpQWC: lNum, FPQWC: fpQW,
				Mats:  xfers,
				STRow: u32(obj, p+44), STOff: u32(obj, p+32),
			}
			gcur += uint32(uQW+lQW+fpQW) * 16
			eff.Fragments = append(eff.Fragments, fr)
		}
		c.Effects = append(c.Effects, eff)
	}
	return c, nil
}

// --- vertex decode ----------------------------------------------------------
//
// The microprogram's vertex loop (mercmicro.dis 0x498+) defines the packing,
// verified visually (the decoded logo renders as the logo):
//
//   - A vertex is THREE consecutive lump quadwords; in the V4-8 source that
//     is 12 bytes {q0: b0,ctl,nx,px | q1: dst1,dst2,ny,py | q2: s,t,nz,pz}.
//     The normal is the z-lanes IN QUADWORD ORDER — the microcode's normal
//     transform is row4*q0.z + row5*q1.z + row6*q2.z, so the vector is
//     (b2, b6, b10) (verified: 0.97 mean |dot| against geometric normals;
//     the swapped order scores 0.06).
//     The w-lane bytes are the position on a fragment-local 8-bit lattice;
//     the z-lane bytes are the normal; q1's x/y-lane bytes are OUTPUT SLOT
//     addresses (VU quadwords, stride 3, base 7): the microcode scatter-
//     writes each vertex into one or two GIF strip slots (two when a strip
//     stitch reuses it), which is why an effect's dvert count exceeds its
//     unique vertex count.
//   - The fragment's fp region opens with the fragment origin as
//     int-in-float (bias 0x4B010000 = 8454144.0, sign-folded):
//     world = lattice byte + origin. The STROW payload emitted by
//     draw-bones-merc carries the matching magic constants for the VIF row
//     mode; offline, the plain byte + origin sum reproduces it.
type Vertex struct {
	X, Y, Z    float32
	NX, NY, NZ uint8 // raw lattice bytes; bias 128
	Slot1      int // output strip slot, -1 if unused
	Slot2      int
	D1, D2     byte // raw dest bytes
	Ctl        byte // q0 y-lane ADC control (biased 0x80 via the VIF row's low half)
	Mat        byte // q0 x-lane: matrix selector; its sign also guards the D2 rebuild
	U, V       float32 // texture coords (GS ST space, REPEAT wrap): see Vertices
}

func magicInt(f float32) float32 {
	if f < 0 {
		return f + 8454144.0
	}
	return f - 8454144.0
}

func f32at(b []byte, o int) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(b[o:]))
}

func float32frombits(v uint32) float32 { return math.Float32frombits(v) }

// Vertices decodes a fragment's vertex set.
func (fr *Fragment) Vertices() []Vertex {
	if len(fr.FPData) < 16 {
		return nil
	}
	ox := magicInt(f32at(fr.FPData, 0))
	oy := magicInt(f32at(fr.FPData, 4))
	oz := magicInt(f32at(fr.FPData, 8))
	n := fr.LumpQWC / 3
	out := make([]Vertex, 0, n)
	slot := func(d byte) int {
		if d >= 7 && (int(d)-7)%3 == 0 {
			return (int(d) - 7) / 3
		}
		return -1
	}
	// Texture coords reproduce the machine's float path exactly: the VIF
	// row adds ctrl+44 to the byte at the bit level (a mantissa step), the
	// microcode then adds vf19 = ctrl+32. For the logo the pair is
	// ±66047.0 (u = s/128); evilbro's is 264188.0 / -264189.5 (u = s/32
	// - 1.5) — always compute it from the words.
	st := func(b byte) float32 {
		return float32frombits(fr.STRow+uint32(b)) + float32frombits(fr.STOff)
	}
	for v := 0; v < n; v++ {
		b := fr.LumpData[v*12 : v*12+12]
		out = append(out, Vertex{
			X: float32(b[3]) + ox, Y: float32(b[7]) + oy, Z: float32(b[11]) + oz,
			NX: b[2], NY: b[6], NZ: b[10],
			Slot1: slot(b[4]), Slot2: slot(b[5]),
			D1: b[4], D2: b[5],
			Ctl: b[1], Mat: b[0],
			U:  st(b[8]), V: st(b[9]),
		})
	}
	return out
}

// Colors returns the fragment's per-vertex RGBA records (the byte-region
// stream at header byte 1's quadword offset — right after the copy table:
// one 4-byte record per vertex, {r, g, b, a} with GS 0x80 = 1.0).
func (fr *Fragment) Colors() [][4]byte {
	if len(fr.ByteData) < 23 {
		return nil
	}
	off := int(fr.ByteData[1]) * 4
	n := fr.LumpQWC / 3
	out := make([][4]byte, 0, n)
	for v := 0; v < n; v++ {
		o := off + v*4
		if o+4 > len(fr.ByteData) {
			break
		}
		out = append(out, [4]byte{fr.ByteData[o], fr.ByteData[o+1], fr.ByteData[o+2], fr.ByteData[o+3]})
	}
	return out
}
