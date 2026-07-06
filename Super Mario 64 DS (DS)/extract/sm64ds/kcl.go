// KCL collision-mesh decoding and the down-ray query, reimplemented from the
// game's own walker (Part VI). Layout and semantics were pinned by running the
// real collision code in the oracle under a read watch (cmd/kcltrace) and
// disassembling the routines the watch flagged:
//
//   - $02039760 fixes the header up in place: the first four u32s are offsets
//     (vertices, normals, prisms, spatial index), each turned into a pointer.
//   - Vertices: 12-byte fx32 x/y/z in world>>6 units (the accessor at
//     $01FFD890, vtable slot +$14, returns them <<6 as fx20.12).
//   - Normals: 6-byte s16 x/y/z in fx.10 ($01FFD8D8, slot +$10, returns <<2).
//   - Prisms: 16 bytes — u32 fx length, u16 vertex index, u16 face-normal
//     index, u16 edge-normal indices A/B/C, u16 attribute (an index into the
//     level's external "CLPS" table via $020381CC: 8-byte header, 8-byte
//     entries; the .kcl itself carries only the index).
//   - Spatial index: an octree of s32 words. Root grid indexed by
//     ((z>>s)<<hdr.shiftZ | (y>>s)<<hdr.shiftY | x>>s) with s = hdr.shift;
//     a word >= 0 is a child-table offset relative to the current table,
//     descending one bit per level (child = xbit | zbit<<1 | ybit<<2);
//     bit 31 marks a leaf whose payload (offset & ~bit31) + 2 is a
//     0-terminated u16 prism-index list. Prism 0 is therefore a dummy.
//   - Header params: +$10 fx32 prism thickness (sphere queries allow that
//     much penetration behind the plane; the down-ray never reads it),
//     +$14/+$18/+$1C fx32 area minimum x/y/z in world>>6 units,
//     +$20/+$24/+$28 coordinate masks (~mask = local-coordinate maximum),
//     +$2C root shift, +$30/+$34 the y/z root-index shifts.
//
// The down-ray walker ($01FFD3F8, dBgW_Kc vtable slot +$18, ITCM) is
// replicated bit for bit in RaycastDown, including the hardware divider's
// rounding, the 32-bit wrapping edge tests and the walk order — the oracle
// verifies the match (kcltrace -verify).
package sm64ds

import (
	"encoding/binary"
	"fmt"
)

var kle = binary.LittleEndian

// KCLHeader mirrors the 0x38-byte file header after decompression, with the
// four section pointers still as file offsets.
type KCLHeader struct {
	VertsOff, NormalsOff, PrismsOff, OctreeOff uint32
	Thickness                                  int32    // +0x10 fx20.12
	Min                                        [3]int32 // +0x14 fx26.6 (world>>6)
	Mask                                       [3]uint32
	Shift                                      uint32 // +0x2C root cell shift
	ShiftY, ShiftZ                             uint32 // +0x30/+0x34 root index shifts
}

// KCL is one decoded collision mesh.
type KCL struct {
	Hdr KCLHeader
	raw []byte
}

// ParseKCL wraps a decompressed .kcl file.
func ParseKCL(data []byte) (*KCL, error) {
	if len(data) < 0x38 {
		return nil, fmt.Errorf("kcl: %d bytes is too short", len(data))
	}
	k := &KCL{raw: data}
	h := &k.Hdr
	h.VertsOff = kle.Uint32(data[0x00:])
	h.NormalsOff = kle.Uint32(data[0x04:])
	h.PrismsOff = kle.Uint32(data[0x08:])
	h.OctreeOff = kle.Uint32(data[0x0C:])
	h.Thickness = int32(kle.Uint32(data[0x10:]))
	for i := 0; i < 3; i++ {
		h.Min[i] = int32(kle.Uint32(data[0x14+4*i:]))
		h.Mask[i] = kle.Uint32(data[0x20+4*i:])
	}
	h.Shift = kle.Uint32(data[0x2C:])
	h.ShiftY = kle.Uint32(data[0x30:])
	h.ShiftZ = kle.Uint32(data[0x34:])
	if h.VertsOff > h.NormalsOff || h.NormalsOff > h.PrismsOff ||
		h.PrismsOff > h.OctreeOff || h.OctreeOff >= uint32(len(data)) {
		return nil, fmt.Errorf("kcl: sections out of order (%X %X %X %X)",
			h.VertsOff, h.NormalsOff, h.PrismsOff, h.OctreeOff)
	}
	return k, nil
}

// NumPrisms counts the 16-byte prism records (record 0 is the dummy).
func (k *KCL) NumPrisms() int { return int(k.Hdr.OctreeOff-k.Hdr.PrismsOff) / 16 }

// Prism is one 16-byte collision-triangle record.
type Prism struct {
	Length     int32 // fx (same space as the edge dot tests)
	Vertex     int   // 12-byte record index into the vertex section
	FaceNormal int   // 6-byte record index into the normal section
	EdgeA      int
	EdgeB      int
	EdgeC      int
	Attribute  int // CLPS table index (the table lives in the level overlay)
}

// PrismAt decodes record i.
func (k *KCL) PrismAt(i int) Prism {
	p := k.raw[k.Hdr.PrismsOff+uint32(i)*16:]
	return Prism{
		Length:     int32(kle.Uint32(p)),
		Vertex:     int(kle.Uint16(p[4:])),
		FaceNormal: int(kle.Uint16(p[6:])),
		EdgeA:      int(kle.Uint16(p[8:])),
		EdgeB:      int(kle.Uint16(p[10:])),
		EdgeC:      int(kle.Uint16(p[12:])),
		Attribute:  int(kle.Uint16(p[14:])),
	}
}

// VertexAt returns vertex record i in fx20.12 world units (<<6, as the game's
// accessor at $01FFD890 does).
func (k *KCL) VertexAt(i int) [3]int32 {
	v := k.raw[k.Hdr.VertsOff+uint32(i)*12:]
	return [3]int32{
		int32(kle.Uint32(v)) << 6,
		int32(kle.Uint32(v[4:])) << 6,
		int32(kle.Uint32(v[8:])) << 6,
	}
}

func (k *KCL) vertexRaw(i int) [3]int32 {
	v := k.raw[k.Hdr.VertsOff+uint32(i)*12:]
	return [3]int32{int32(kle.Uint32(v)), int32(kle.Uint32(v[4:])), int32(kle.Uint32(v[8:]))}
}

// NormalAt returns normal record i in fx20.12 (<<2, as $01FFD8D8 does).
func (k *KCL) NormalAt(i int) [3]int32 {
	n := k.normalRaw(i)
	return [3]int32{n[0] << 2, n[1] << 2, n[2] << 2}
}

func (k *KCL) normalRaw(i int) [3]int32 {
	n := k.raw[k.Hdr.NormalsOff+uint32(i)*6:]
	return [3]int32{
		int32(int16(kle.Uint16(n))),
		int32(int16(kle.Uint16(n[2:]))),
		int32(int16(kle.Uint16(n[4:]))),
	}
}

// fxDiv32 replicates the game's divider helper at $02053258: numerator<<32
// over a 32-bit denominator on the hardware divider, then rounded down to
// fx.12 — result = low 32 bits of (((n<<32)/d) + 0x80000) >> 20.
func fxDiv32(n, d int32) int32 {
	q := (int64(n) << 32) / int64(d)
	return int32((q + 0x80000) >> 20)
}

// CLPS is a level's surface-attribute table — the block the settings word at
// +$00 points at, NOT part of the .kcl (prisms only carry an index into it).
// Layout per $020381CC: "CLPS" magic, u16 entry size (must be 8), u16 count,
// then 8-byte entries. A missing or malformed block makes every lookup return
// the default entry {0xFC0, 0xFF} ($02037E9C).
type CLPS struct{ raw []byte }

// NewCLPS wraps a CLPS block.
func NewCLPS(raw []byte) *CLPS { return &CLPS{raw} }

// Word0 returns entry i's first word (the behavior/property bits the query
// filter at $02039488 tests).
func (c *CLPS) Word0(i int) uint32 {
	if c == nil || len(c.raw) < 8 || string(c.raw[:4]) != "CLPS" || kle.Uint16(c.raw[4:]) != 8 {
		return 0xFC0
	}
	off := 8 + i*8
	if off+8 > len(c.raw) { // the game trusts the index; don't crash on bad data
		return 0xFC0
	}
	return kle.Uint32(c.raw[off:])
}

// clpsReject replicates the surface filter at $02039488 for the down-ray
// (wantSpecial=false there): each special surface property demands a matching
// opt-in bit in the query's flag byte (+$04, ctor default 1); a surface with
// no special properties needs flag bit 0.
func clpsReject(w0 uint32, qf byte) bool {
	special := false
	if w0&0x20 != 0 { // bit 5
		special = true
		if qf&2 == 0 {
			return true
		}
	}
	if w0&0x2000000 != 0 { // bit 25
		special = true
		if qf&8 == 0 {
			return true
		}
	}
	if w0&0x1000000 != 0 { // bit 24: excluded only when the query opts INTO 4
		special = true
		if qf&4 != 0 {
			return true
		}
	}
	if w0&0x4000000 != 0 { // bit 26
		special = true
		if qf&4 == 0 {
			return true
		}
	}
	if w0>>19&0x1F == 0x11 { // behavior type 17
		special = true
		if qf&0x20 != 0 {
			return true
		}
	}
	if !special && qf&1 == 0 {
		return true
	}
	return false
}

// KCLTrace, when set, logs the walker's descent (kcltrace debugging).
var KCLTrace func(format string, args ...any)

func ktrace(format string, args ...any) {
	if KCLTrace != nil {
		KCLTrace(format, args...)
	}
}

// RayHit is the down-ray result.
type RayHit struct {
	Y     int32 // fx20.12 ground height
	Prism int   // winning prism record index
}

// RaycastDown replicates the ground-ray walker at $01FFD3F8 (dBgW_Kc vtable
// slot +$18): the highest floor prism (face normal y > 8/1024) strictly below
// (x,y,z) and strictly above floorBelow (the query's +$44 in/out field, which
// the manager's reset at $02037464 makes -0x80000000). Surfaces are filtered
// through the level's CLPS table against the query flag byte qf (ctor default
// 1). All coordinates are fx20.12.
func (k *KCL) RaycastDown(x, y, z, floorBelow int32, clps *CLPS, qf byte) (RayHit, bool) {
	h := &k.Hdr
	x6, y6, z6 := x>>6, y>>6, z>>6

	lx := (x6 - h.Min[0]) >> 6
	if lx < 0 || lx > int32(^h.Mask[0]) {
		return RayHit{}, false
	}
	lz := (z6 - h.Min[2]) >> 6
	if lz < 0 || lz > int32(^h.Mask[2]) {
		return RayHit{}, false
	}
	ly := (y6 - h.Min[1]) >> 6
	if ly < 0 {
		return RayHit{}, false
	}
	if ly > int32(^h.Mask[1]) {
		ly = int32(^h.Mask[1]) // above the area: clamp, don't reject
	}

	best := floorBelow >> 6
	bestPrism := -1
	oct := k.raw[h.OctreeOff:]

	for ly >= 0 {
		// descend to the leaf under (lx, ly, lz)
		s := h.Shift
		idx := uint32(lz)>>s<<h.ShiftZ | uint32(ly)>>s<<h.ShiftY | uint32(lx)>>s
		base := uint32(0)
		node := int32(kle.Uint32(oct[base+idx*4:]))
		for node >= 0 {
			base += uint32(node)
			s--
			child := uint32(lx)>>s&1 | uint32(ly)>>s&1<<1 | uint32(lz)>>s&1<<2
			node = int32(kle.Uint32(oct[base+child*4:]))
		}
		list := base + uint32(node)&^0x80000000 + 2
		ktrace("cell (lx=%d ly=%d lz=%d) s=%d leaf list at oct+%X", lx, ly, lz, s, list)

		for {
			pi := int(kle.Uint16(oct[list:]))
			list += 2
			if pi == 0 {
				break
			}
			p := k.PrismAt(pi)
			fn := k.normalRaw(p.FaceNormal)
			ktrace("  prism %d fn.y=%d", pi, fn[1])
			if fn[1] <= 0 {
				continue // floors only
			}
			if fn[1] <= 8 && fn[1] >= -8 {
				continue // $020397DC: too vertical to solve against
			}
			v := k.vertexRaw(p.Vertex)
			dx, dz := x6-v[0], z6-v[2]

			// plane solve for the height offset at (x,z), $01FFD62C..$01FFD650
			num := int32(int64(dx)*int64(fn[0])>>10) + int32(int64(dz)*int64(fn[2])>>10)
			ht := -(fxDiv32(num, fn[1]) >> 2)

			// edge tests (32-bit, wrapping, ±0x20000 tolerance)
			dot := func(ni int) int32 {
				en := k.normalRaw(ni)
				return dx*en[0] + ht*en[1] + dz*en[2]
			}
			if dot(p.EdgeA) > 0x20000 || dot(p.EdgeB) > 0x20000 {
				continue
			}
			dc := dot(p.EdgeC)
			if dc < -0x20000 || dc > p.Length+0x20000 {
				continue
			}

			// ray origin must be on the floor's front side (exact 64-bit dot)
			dy := y6 - v[1]
			if int64(dx)*int64(fn[0])+int64(dy)*int64(fn[1])+int64(dz)*int64(fn[2]) < 0 {
				continue
			}

			// surface filter ($02039488 via the CLPS entry the attribute names)
			if clpsReject(clps.Word0(p.Attribute), qf) {
				ktrace("  prism %d filtered (attr %d w0=%08X)", pi, p.Attribute, clps.Word0(p.Attribute))
				continue
			}

			hitY := ht + v[1]
			ktrace("  prism %d candidate hitY=%d (best=%d y6=%d)", pi, hitY, best, y6)
			if hitY < h.Min[1] || best >= hitY || y6 <= hitY {
				continue
			}
			best, bestPrism = hitY, pi
		}

		// no acceptable floor in this cell: step to the cell just below in y
		ly = ly&^(1<<s-1) - 1
	}

	if bestPrism < 0 {
		return RayHit{}, false
	}
	return RayHit{Y: best << 6, Prism: bestPrism}, true
}
