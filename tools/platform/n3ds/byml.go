package n3ds

import (
	"encoding/binary"
	"fmt"
	"math"
)

// byml.go reads BYML ("BinarY Yaml"), the format every configuration file in
// this platform's asset bundles is written in: the stage maps that place a
// level's objects, the camera parameters, and the per-object Init*.byml
// settings. It is a tree of dictionaries, arrays and scalars with two shared
// string tables — one for dictionary keys, one for string values — so a whole
// stage's placement data decodes into ordinary Go values.
//
//	0x00 "YB"  u16 version(2)
//	0x04 u32 hash-key table offset   (a string table: the dictionary keys)
//	0x08 u32 string table offset     (a string table: the string values)
//	0x0C u32 root node offset
//
// Every node begins with a byte naming its type and a 24-bit count:
//
//	0xA0 string      the value word indexes the string table
//	0xC0 array       count type bytes (padded to 4), then count value words
//	0xC1 dictionary  count × { u24 key index, u8 type, u32 value }
//	0xC2 string tbl  count+1 offsets, relative to the table, then the bytes
//	0xD0 bool  0xD1 int32  0xD2 float32  0xD3 uint32
//
// A value word is the value itself for the scalar types and an offset from the
// start of the file for the container ones.
const bymlMagic = "YB"

// BYML node types.
const (
	bymlString  = 0xA0
	bymlBinary  = 0xA1
	bymlArray   = 0xC0
	bymlDict    = 0xC1
	bymlStrings = 0xC2
	bymlBool    = 0xD0
	bymlInt     = 0xD1
	bymlFloat   = 0xD2
	bymlUint    = 0xD3
	bymlInt64   = 0xD4
	bymlUint64  = 0xD5
	bymlDouble  = 0xD6
	bymlNull    = 0xFF
)

// BYMLDict is a decoded dictionary node; BYML arrays decode to []any and the
// scalars to string, bool, int32, uint32, float32 or nil.
type BYMLDict map[string]any

// Get walks a path of dictionary keys, returning nil if any step is missing.
func (d BYMLDict) Get(path ...string) any {
	var cur any = d
	for _, k := range path {
		m, ok := cur.(BYMLDict)
		if !ok {
			return nil
		}
		cur, ok = m[k]
		if !ok {
			return nil
		}
	}
	return cur
}

type bymlFile struct {
	b        []byte
	keys     []string
	strs     []string
	depth    int
	maxDepth int
}

// ParseBYML decodes a BYML document into Go values.
func ParseBYML(b []byte) (any, error) {
	if len(b) < 0x10 || string(b[:2]) != bymlMagic {
		return nil, fmt.Errorf("byml: bad magic %q", b[:min(2, len(b))])
	}
	if v := binary.LittleEndian.Uint16(b[2:]); v != 2 {
		return nil, fmt.Errorf("byml: version %d not supported", v)
	}
	f := &bymlFile{b: b, maxDepth: 64}
	var err error
	if off := int64(binary.LittleEndian.Uint32(b[4:])); off != 0 {
		if f.keys, err = f.stringTable(off); err != nil {
			return nil, fmt.Errorf("byml: hash-key table: %w", err)
		}
	}
	if off := int64(binary.LittleEndian.Uint32(b[8:])); off != 0 {
		if f.strs, err = f.stringTable(off); err != nil {
			return nil, fmt.Errorf("byml: string table: %w", err)
		}
	}
	root := int64(binary.LittleEndian.Uint32(b[0xC:]))
	if root == 0 {
		return nil, nil
	}
	return f.node(root)
}

// stringTable reads a 0xC2 node: count+1 offsets relative to the node, then the
// NUL-terminated bytes. The trailing offset is the table's end, so the offsets
// are self-checking — they must rise and stay inside the file.
func (f *bymlFile) stringTable(off int64) ([]string, error) {
	if off+4 > int64(len(f.b)) {
		return nil, fmt.Errorf("node at 0x%x runs outside the file", off)
	}
	if t := f.b[off]; t != bymlStrings {
		return nil, fmt.Errorf("node at 0x%x has type 0x%02x, want a string table", off, t)
	}
	n := int64(f.u24(off + 1))
	if off+4+(n+1)*4 > int64(len(f.b)) {
		return nil, fmt.Errorf("table of %d strings runs outside the file", n)
	}
	out := make([]string, n)
	for i := int64(0); i < n; i++ {
		s := off + int64(binary.LittleEndian.Uint32(f.b[off+4+i*4:]))
		if s < off || s >= int64(len(f.b)) {
			return nil, fmt.Errorf("string %d at 0x%x is outside the file", i, s)
		}
		out[i] = readCStr(f.b, s)
	}
	return out, nil
}

func (f *bymlFile) u24(o int64) uint32 {
	return uint32(f.b[o]) | uint32(f.b[o+1])<<8 | uint32(f.b[o+2])<<16
}

// node decodes the container node at off.
func (f *bymlFile) node(off int64) (any, error) {
	if off+4 > int64(len(f.b)) {
		return nil, fmt.Errorf("node at 0x%x runs outside the file", off)
	}
	if f.depth++; f.depth > f.maxDepth {
		return nil, fmt.Errorf("node nesting deeper than %d", f.maxDepth)
	}
	defer func() { f.depth-- }()

	typ := f.b[off]
	n := int64(f.u24(off + 1))
	switch typ {
	case bymlArray:
		vals := off + 4 + (n+3)/4*4
		if vals+n*4 > int64(len(f.b)) {
			return nil, fmt.Errorf("array of %d at 0x%x runs outside the file", n, off)
		}
		out := make([]any, n)
		for i := int64(0); i < n; i++ {
			v, err := f.value(f.b[off+4+i], binary.LittleEndian.Uint32(f.b[vals+i*4:]))
			if err != nil {
				return nil, fmt.Errorf("array at 0x%x element %d: %w", off, i, err)
			}
			out[i] = v
		}
		return out, nil
	case bymlDict:
		if off+4+n*8 > int64(len(f.b)) {
			return nil, fmt.Errorf("dictionary of %d at 0x%x runs outside the file", n, off)
		}
		out := make(BYMLDict, n)
		for i := int64(0); i < n; i++ {
			e := off + 4 + i*8
			ki := f.u24(e)
			if int(ki) >= len(f.keys) {
				return nil, fmt.Errorf("dictionary at 0x%x: key index %d of %d", off, ki, len(f.keys))
			}
			v, err := f.value(f.b[e+3], binary.LittleEndian.Uint32(f.b[e+4:]))
			if err != nil {
				return nil, fmt.Errorf("dictionary at 0x%x key %q: %w", off, f.keys[ki], err)
			}
			out[f.keys[ki]] = v
		}
		return out, nil
	default:
		return nil, fmt.Errorf("node at 0x%x has unsupported type 0x%02x", off, typ)
	}
}

// value decodes one typed word: the value itself for scalars, an offset into
// the file for containers.
func (f *bymlFile) value(typ byte, w uint32) (any, error) {
	switch typ {
	case bymlString:
		if int(w) >= len(f.strs) {
			return nil, fmt.Errorf("string index %d of %d", w, len(f.strs))
		}
		return f.strs[w], nil
	case bymlBool:
		return w != 0, nil
	case bymlInt:
		return int32(w), nil
	case bymlUint:
		return w, nil
	case bymlFloat:
		return math.Float32frombits(w), nil
	case bymlNull:
		return nil, nil
	case bymlArray, bymlDict:
		return f.node(int64(w))
	default:
		return nil, fmt.Errorf("unsupported value type 0x%02x", typ)
	}
}
