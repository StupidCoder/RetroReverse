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
)

// Fragment is one VU1 batch: three consecutive source regions plus its bone
// matrix uploads.
type Fragment struct {
	ByteData []byte // V4-8 region: merc-byte-header + packed vertex bytes
	LumpData []byte // V4-8 region: "lump" packed vertex bytes
	FPData   []byte // V4-32 region: float data (merc-fp-header + floats)
	ByteQWC  int    // NUM of the first unpack (rows of 4 bytes)
	LumpQWC  int    // NUM of the second unpack
	FPQWC    int    // quadwords of float data
	Mats     []MatXfer
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
				ByteData: obj[gcur : gcur+uint32(uQW)*16],
				LumpData: obj[gcur+uint32(uQW)*16 : gcur+uint32(uQW+lQW)*16],
				FPData:   obj[gcur+uint32(uQW+lQW)*16 : gcur+uint32(uQW+lQW+fpQW)*16],
				ByteQWC:  uNum, LumpQWC: lNum, FPQWC: fpQW,
				Mats: xfers,
			}
			gcur += uint32(uQW+lQW+fpQW) * 16
			eff.Fragments = append(eff.Fragments, fr)
		}
		c.Effects = append(c.Effects, eff)
	}
	return c, nil
}
