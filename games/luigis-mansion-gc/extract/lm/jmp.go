package lm

// jmp.go reads the JMap tables in Map/map*.szp (jmp/furnitureinfo,
// jmp/roominfo, …) — the game's placement and parameter databases. The
// format, read straight off the file: a 16-byte header {u32 recordCount,
// u32 fieldCount, u32 dataOffset, u32 recordSize}, then fieldCount 12-byte
// descriptors {u32 nameHash, u32 bitmask, u16 offset, u8 shift, u8 type}.
// Field names are stored only as hashes; a field is read from its record
// at +offset as a 32-byte string (type 1), a f32 (type 2), or a u32 masked
// and shifted (otherwise).

import (
	"encoding/binary"
	"fmt"
	"math"
)

// JMPField is one column of a JMap table.
type JMPField struct {
	Hash   uint32
	Mask   uint32
	Offset uint16
	Shift  uint8
	Type   uint8
}

// JMPTable is a parsed JMap table. Records hold every field keyed by its
// name hash: string, float32, or uint32.
type JMPTable struct {
	Fields  []JMPField
	Records []map[uint32]any
}

// Field name hashes used by this package's callers (recovered by matching
// the columns' values against known data — e.g. the furniture positions
// against the rooms' vertex bounds; the names themselves live only in
// Nintendo's conversion tool).
const (
	JMPName    = 0x006175C6 // actor class, e.g. "furniture", "room"
	JMPDMDName = 0x009BC8CD // model/member name, e.g. "otukue"
	JMPPosX    = 0x017BEFD9
	JMPPosY    = 0x017BEFDA
	JMPPosZ    = 0x017BEFDB
	JMPDirX    = 0x017A0564 // rotation, degrees
	JMPDirY    = 0x017A0565
	JMPDirZ    = 0x017A0566
	JMPSclX    = 0x00E4750F
	JMPSclY    = 0x00E47510
	JMPSclZ    = 0x00E47511
	JMPRoomNo  = 0x00C9952D
)

// ParseJMP reads a JMap table.
func ParseJMP(b []byte) (*JMPTable, error) {
	if len(b) < 16 {
		return nil, fmt.Errorf("lm: jmp table too short")
	}
	u32 := func(o int) uint32 { return binary.BigEndian.Uint32(b[o:]) }
	count, nf, data, size := int(u32(0)), int(u32(4)), int(u32(8)), int(u32(12))
	if nf <= 0 || nf > 256 || count < 0 || size <= 0 || 16+nf*12 > len(b) {
		return nil, fmt.Errorf("lm: not a jmp table")
	}
	t := &JMPTable{}
	for i := 0; i < nf; i++ {
		o := 16 + i*12
		t.Fields = append(t.Fields, JMPField{
			Hash:   u32(o),
			Mask:   u32(o + 4),
			Offset: binary.BigEndian.Uint16(b[o+8:]),
			Shift:  b[o+10],
			Type:   b[o+11],
		})
	}
	for r := 0; r < count; r++ {
		base := data + r*size
		if base+size > len(b) {
			return nil, fmt.Errorf("lm: jmp record %d out of bounds", r)
		}
		rec := make(map[uint32]any, nf)
		for _, f := range t.Fields {
			o := base + int(f.Offset)
			switch f.Type {
			case 1:
				end := o
				for end < o+32 && end < len(b) && b[end] != 0 {
					end++
				}
				rec[f.Hash] = string(b[o:end])
			case 2:
				rec[f.Hash] = math.Float32frombits(u32(o))
			default:
				v := u32(o)
				if f.Mask != 0 {
					v = (v & f.Mask) >> f.Shift
				}
				rec[f.Hash] = v
			}
		}
		t.Records = append(t.Records, rec)
	}
	return t, nil
}

// Str / F32 / U32 read one typed field of a record, tolerating absence.
func (t *JMPTable) Str(r int, hash uint32) string {
	s, _ := t.Records[r][hash].(string)
	return s
}
func (t *JMPTable) F32(r int, hash uint32) float32 {
	f, _ := t.Records[r][hash].(float32)
	return f
}
func (t *JMPTable) U32(r int, hash uint32) uint32 {
	u, _ := t.Records[r][hash].(uint32)
	return u
}
