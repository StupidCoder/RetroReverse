package n3ds

import (
	"encoding/binary"
	"fmt"
)

// sarc.go reads SARC, the archive every one of this platform's asset bundles is
// packed in. A `.szs` file is Yaz0 (message.go) wrapping a SARC, and a SARC is a
// flat list of named files: the stage maps, the object models, the textures and
// the animations all arrive this way.
//
// Layout, all little-endian (the byte-order mark says which, and is checked):
//
//	"SARC"  u16 headerLen(0x14)  u16 BOM(0xFEFF)  u32 fileSize
//	        u32 dataOffset       u32 version      u32 reserved
//	"SFAT"  u16 headerLen(0x0C)  u16 nodeCount    u32 hashKey
//	        nodeCount × { u32 nameHash, u32 attrs, u32 start, u32 end }
//	"SFNT"  u16 headerLen(0x08)  u16 reserved     … the name string table
//
// A node's `attrs` carries a "has name" flag in bit 24 and, in its low 24 bits,
// the name's offset into the string table **divided by four**; `start`/`end` are
// relative to the archive's dataOffset.
//
// The names are not merely stored, they are *hashed*, and the archive keeps both
// — which makes this format self-checking in a way most containers are not.
// Rehashing every name with the archive's own key and demanding it match the
// stored hash proves the string-table offsets, the name lengths, the node stride
// and the hash key all at once. ParseSARC insists on it: a layout that has been
// misread cannot survive the check.
const (
	sarcMagic     = "SARC"
	sarcBOM       = 0xFEFF
	sarcHeaderLen = 0x14
	sfatMagic     = "SFAT"
	sfatHeaderLen = 0x0C
	sfatNodeSize  = 0x10
	sfntMagic     = "SFNT"
	sfntHeaderLen = 0x08
)

// SARCFile is one named entry of an archive.
type SARCFile struct {
	Name string
	Data []byte
}

// SARC is a parsed archive.
type SARC struct {
	Files []SARCFile
}

// File returns the named entry's bytes.
func (a *SARC) File(name string) ([]byte, bool) {
	for i := range a.Files {
		if a.Files[i].Name == name {
			return a.Files[i].Data, true
		}
	}
	return nil, false
}

// sarcHash is the archive's name hash: a plain polynomial over the bytes of the
// name with the archive's own key.
func sarcHash(name string, key uint32) uint32 {
	var h uint32
	for i := 0; i < len(name); i++ {
		h = h*key + uint32(name[i])
	}
	return h
}

// ParseSARC parses a decompressed SARC archive.
func ParseSARC(b []byte) (*SARC, error) {
	if len(b) < sarcHeaderLen || string(b[:4]) != sarcMagic {
		return nil, fmt.Errorf("sarc: bad magic")
	}
	if hl := binary.LittleEndian.Uint16(b[4:]); hl != sarcHeaderLen {
		return nil, fmt.Errorf("sarc: header length %d, want %d", hl, sarcHeaderLen)
	}
	if bom := binary.LittleEndian.Uint16(b[6:]); bom != sarcBOM {
		return nil, fmt.Errorf("sarc: unexpected byte order 0x%04x (big-endian SARC not supported)", bom)
	}
	if size := binary.LittleEndian.Uint32(b[8:]); int(size) != len(b) {
		return nil, fmt.Errorf("sarc: header size 0x%x != blob 0x%x", size, len(b))
	}
	dataOff := int64(binary.LittleEndian.Uint32(b[0xC:]))

	sfat := int64(sarcHeaderLen)
	if int64(len(b)) < sfat+sfatHeaderLen || string(b[sfat:sfat+4]) != sfatMagic {
		return nil, fmt.Errorf("sarc: no SFAT at 0x%x", sfat)
	}
	if hl := binary.LittleEndian.Uint16(b[sfat+4:]); hl != sfatHeaderLen {
		return nil, fmt.Errorf("sarc: SFAT header length %d, want %d", hl, sfatHeaderLen)
	}
	n := int(binary.LittleEndian.Uint16(b[sfat+6:]))
	key := binary.LittleEndian.Uint32(b[sfat+8:])

	nodes := sfat + sfatHeaderLen
	sfnt := nodes + int64(n)*sfatNodeSize
	if int64(len(b)) < sfnt+sfntHeaderLen || string(b[sfnt:sfnt+4]) != sfntMagic {
		return nil, fmt.Errorf("sarc: no SFNT after %d nodes (at 0x%x)", n, sfnt)
	}
	strs := sfnt + sfntHeaderLen

	a := &SARC{Files: make([]SARCFile, 0, n)}
	for i := 0; i < n; i++ {
		e := nodes + int64(i)*sfatNodeSize
		hash := binary.LittleEndian.Uint32(b[e:])
		attrs := binary.LittleEndian.Uint32(b[e+4:])
		start := dataOff + int64(binary.LittleEndian.Uint32(b[e+8:]))
		end := dataOff + int64(binary.LittleEndian.Uint32(b[e+0xC:]))
		if attrs>>24&1 == 0 {
			return nil, fmt.Errorf("sarc: node %d carries no name", i)
		}
		nameOff := strs + int64(attrs&0xFFFFFF)*4
		if nameOff < strs || nameOff >= int64(len(b)) {
			return nil, fmt.Errorf("sarc: node %d name offset 0x%x outside the string table", i, nameOff)
		}
		name := readCStr(b, nameOff)
		if got := sarcHash(name, key); got != hash {
			return nil, fmt.Errorf("sarc: node %d name %q hashes to 0x%08x, the archive says 0x%08x", i, name, got, hash)
		}
		if start < 0 || end < start || end > int64(len(b)) {
			return nil, fmt.Errorf("sarc: node %d (%q) spans 0x%x-0x%x, outside the archive", i, name, start, end)
		}
		a.Files = append(a.Files, SARCFile{Name: name, Data: b[start:end]})
	}
	return a, nil
}

// OpenSZS decompresses a `.szs` (Yaz0-wrapped SARC) and parses the archive.
func OpenSZS(b []byte) (*SARC, error) {
	d, err := Yaz0(b)
	if err != nil {
		return nil, err
	}
	return ParseSARC(d)
}
