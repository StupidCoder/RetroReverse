package n3ds

import (
	"encoding/binary"
	"fmt"
)

// CGFX ("CTR Graphics") is Nintendo's NW4C scene-graph container — the format the
// banner's animated 3-D scene is stored in, and the 3DS's general graphics format.
// This file parses its *structure*: the header, the DATA block's typed resource
// dictionaries, and the top-level objects (the model, textures, animations, camera,
// light and scene). It stops at the object headers — decoding a model's vertex
// streams and a texture's tiled pixels into a GLB is a separate, larger stage — but
// it is enough to enumerate and name everything a banner holds, and it lays out the
// exact offsets the geometry/texture decoders will consume.
//
// Structural conventions, all little-endian, all offsets *self-relative* (an offset
// field at address P with value V points to P+V):
//
//	Header 0x14: "CGFX", u16 BOM(0xFEFF), u16 headerLen(0x14), u32 revision,
//	             u32 fileSize, u32 nBlocks.
//	DATA block:  "DATA", u32 size, then 15 (count,dictOffset) pairs — one per
//	             resource type, in a fixed order.
//	DICT:        "DICT", u32 size, u32 count, then a root node and `count` nodes of
//	             0x10 bytes; each node carries a self-relative name offset and a
//	             self-relative data offset (to the object, at its typeId word).
//	Objects:     a u32 typeId, then a 4-char magic (CMDL/TXOB/MTOB/SOBJ/CANM/…),
//	             a revision, and a self-relative name offset.
//
// A second block, "IMAG", holds the raw vertex and texture bytes the descriptors
// point into.
const (
	cgfxMagic    = "CGFX"
	cgfxBOM      = 0xFEFF
	cgfxHeaderLn = 0x14
	dictMagic    = "DICT"
	dataMagic    = "DATA"
)

// cgfxResourceTypes names the 15 DATA dictionaries in their fixed order.
var cgfxResourceTypes = []string{
	"Models", "Textures", "LookupTables", "Materials", "Shaders",
	"Cameras", "Lights", "Fogs", "Scenes", "SkeletalAnimations",
	"MaterialAnimations", "VisibilityAnimations", "CameraAnimations",
	"LightAnimations", "Emitters",
}

// CGFXEntry is one named object in a resource dictionary.
type CGFXEntry struct {
	Name   string
	Offset int64  // absolute offset of the object's typeId word
	Magic  string // the object's 4-char magic (at Offset+4)
}

// CGFX is a parsed CGFX container's structure.
type CGFX struct {
	raw       []byte
	Revision  uint32
	FileSize  uint32
	Resources map[string][]CGFXEntry // resource type → its entries
	IMAGOff   int64                  // start of the IMAG (raw data) block, or 0
}

// Raw returns the backing CGFX bytes (needed by the geometry/texture decoders).
func (c *CGFX) Raw() []byte { return c.raw }

func selfRel32(b []byte, at int64) int64 {
	v := int32(binary.LittleEndian.Uint32(b[at:]))
	if v == 0 {
		return 0
	}
	return at + int64(v)
}

// ParseCGFX parses a decompressed CGFX blob's structure.
func ParseCGFX(b []byte) (*CGFX, error) {
	if len(b) < cgfxHeaderLn || string(b[0:4]) != cgfxMagic {
		return nil, fmt.Errorf("cgfx: bad magic")
	}
	if bom := binary.LittleEndian.Uint16(b[4:]); bom != cgfxBOM {
		return nil, fmt.Errorf("cgfx: unexpected byte order 0x%04x (big-endian CGFX not supported)", bom)
	}
	c := &CGFX{
		raw:       b,
		Revision:  binary.LittleEndian.Uint32(b[8:]),
		FileSize:  binary.LittleEndian.Uint32(b[0xC:]),
		Resources: map[string][]CGFXEntry{},
	}
	if int(c.FileSize) > len(b) {
		return nil, fmt.Errorf("cgfx: fileSize 0x%x exceeds blob 0x%x", c.FileSize, len(b))
	}

	// DATA block immediately follows the header.
	data := int64(cgfxHeaderLn)
	if string(b[data:data+4]) != dataMagic {
		return nil, fmt.Errorf("cgfx: expected DATA block at 0x%x, found %q", data, b[data:data+4])
	}
	// 15 (count, dictOffset) pairs at data+8.
	p := data + 8
	for _, rt := range cgfxResourceTypes {
		count := int32(binary.LittleEndian.Uint32(b[p:]))
		dictOff := selfRel32(b, p+4)
		p += 8
		if count == 0 || dictOff == 0 {
			continue
		}
		entries, err := c.parseDict(dictOff)
		if err != nil {
			return nil, fmt.Errorf("cgfx: %s dict: %w", rt, err)
		}
		c.Resources[rt] = entries
	}

	// Locate the IMAG block (raw data) if present.
	if off := findMagic(b, "IMAG"); off >= 0 {
		c.IMAGOff = int64(off)
	}
	return c, nil
}

// parseDict reads a DICT (patricia tree) and returns its named entries. The tree
// links are for lookup; a linear walk of the nodes yields every entry, which is
// all an extractor needs.
func (c *CGFX) parseDict(off int64) ([]CGFXEntry, error) {
	b := c.raw
	if off < 0 || off+12 > int64(len(b)) || string(b[off:off+4]) != dictMagic {
		return nil, fmt.Errorf("bad DICT at 0x%x", off)
	}
	count := int(binary.LittleEndian.Uint32(b[off+8:]))
	var out []CGFXEntry
	node := off + 0x0C
	// node[0] is the root; entries are node[1..count].
	for i := 0; i <= count; i++ {
		if node+0x10 > int64(len(b)) {
			return nil, fmt.Errorf("DICT node runs past end at 0x%x", node)
		}
		if i > 0 { // skip the root (no name/data)
			nameAbs := selfRel32(b, node+8)
			dataAbs := selfRel32(b, node+12)
			e := CGFXEntry{Offset: dataAbs}
			if nameAbs > 0 && nameAbs < int64(len(b)) {
				e.Name = readCStr(b, nameAbs)
			}
			if dataAbs > 0 && dataAbs+8 <= int64(len(b)) {
				e.Magic = string(b[dataAbs+4 : dataAbs+8])
			}
			out = append(out, e)
		}
		node += 0x10
	}
	return out, nil
}

func readCStr(b []byte, off int64) string {
	end := off
	for end < int64(len(b)) && b[end] != 0 {
		end++
	}
	return string(b[off:end])
}

func findMagic(b []byte, magic string) int {
	for i := 0; i+4 <= len(b); i += 4 { // block magics are 4-aligned
		if string(b[i:i+4]) == magic {
			return i
		}
	}
	return -1
}
