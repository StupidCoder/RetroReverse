package nitro

import (
	"fmt"
	"strings"
)

// NSBMD (`BMD0`, block `MDL0`) is the NITRO 3D model format: a named set of models,
// each a small scene — named nodes (joints) with TRS transforms, an "SBC" scene
// bytecode that walks those nodes and binds materials to shapes, materials that
// name their texture and palette (the authoritative texture↔palette binding the
// texture files alone lack), and shapes whose geometry is a raw display list of
// packed GPU (GX/G3) commands, exactly as the DS's geometry engine consumes them.
//
// Layout (offsets verified against the game's models):
//
//	model header +$04 sbcOffset  +$08 matOffset  +$0C shpOffset  (+$10 envelopes)
//	             +$17 numNodes   +$18 numMats    +$19 numShapes
//	             +$1C upScale fx32 (whole-model scale; downScale at +$20)
//	nodes:    resource dict → per-node offset → flags + packed TRS
//	SBC:      byte-coded program: NODEDESC (node×parent → matrix stack), MAT, SHP…
//	material: +$00/+$02 tex/pal-to-material lists; dict → texture/palette names
//	shape:    dict → { dlOffset, dlSize } → display list (displaylist.go)

// Model is one decoded model.
type Model struct {
	Name      string
	UpScale   float64 // fx32 whole-model scale (applied by the SBC's POSSCALE)
	Nodes     []Node
	SBC       []byte
	Materials []Material
	Shapes    []Shape
}

// Node is one joint with its local (node→parent) transform.
type Node struct {
	Name string
	M    Mat43 // local TRS matrix
}

// Material carries the texture/palette names bound by the model (resolved against
// the companion NSBTX) and the GX texture parameters.
type Material struct {
	Name     string
	Texture  string // bound texture name ("" when untextured)
	Palette  string
	TexParam uint32 // GX TEXIMAGE_PARAM (repeat/flip bits etc.)
	Width    int
	Height   int
}

// Shape is one geometry batch: a display list, drawn with the material the SBC
// binds before it.
type Shape struct {
	Name string
	DL   []byte
}

// Mat43 is a 4x3 affine matrix (rows × columns; row 3 is the translation).
type Mat43 [12]float64

// Identity43 returns the identity transform.
func Identity43() Mat43 {
	return Mat43{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}
}

// Mul composes a·b (apply b, then a).
func (a Mat43) Mul(b Mat43) Mat43 {
	var r Mat43
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			r[i*3+j] = b[i*3+0]*a[0*3+j] + b[i*3+1]*a[1*3+j] + b[i*3+2]*a[2*3+j]
		}
	}
	r[9] += a[9]
	r[10] += a[10]
	r[11] += a[11]
	return r
}

// Apply transforms a point.
func (m Mat43) Apply(x, y, z float64) (float64, float64, float64) {
	return x*m[0] + y*m[3] + z*m[6] + m[9],
		x*m[1] + y*m[4] + z*m[7] + m[10],
		x*m[2] + y*m[5] + z*m[8] + m[11]
}

func fx32(v uint32) float64 { return float64(int32(v)) / 4096 }
func fx16(v uint16) float64 { return float64(int16(v)) / 4096 }

// ParseNSBMD decodes every model in a BMD0 file.
func ParseNSBMD(data []byte) ([]Model, error) {
	if len(data) < 0x14 || string(data[0:4]) != "BMD0" {
		return nil, fmt.Errorf("nitro: not a BMD0 file")
	}
	mdl0 := int(le.Uint32(data[0x10:]))
	if mdl0+8 > len(data) || string(data[mdl0:mdl0+4]) != "MDL0" {
		return nil, fmt.Errorf("nitro: no MDL0 block")
	}
	dict, err := parseDict(data, mdl0+8)
	if err != nil {
		return nil, err
	}
	var out []Model
	for _, e := range dict {
		off := mdl0 + int(le.Uint32(padded(e.data)))
		m, err := parseModel(data, off, e.name)
		if err != nil {
			return nil, fmt.Errorf("model %q: %w", e.name, err)
		}
		out = append(out, m)
	}
	return out, nil
}

func parseModel(data []byte, base int, name string) (Model, error) {
	if base+0x40 > len(data) {
		return Model{}, fmt.Errorf("header out of range")
	}
	sbcOff := base + int(le.Uint32(data[base+0x04:]))
	matOff := base + int(le.Uint32(data[base+0x08:]))
	shpOff := base + int(le.Uint32(data[base+0x0C:]))
	m := Model{Name: name, UpScale: fx32(le.Uint32(data[base+0x1C:]))}

	// Nodes: dict right after the 0x40-byte header.
	ndict, err := parseDict(data, base+0x40)
	if err != nil {
		return m, fmt.Errorf("node dict: %w", err)
	}
	nodeSect := base + 0x40
	for _, e := range ndict {
		off := nodeSect + int(le.Uint32(padded(e.data)))
		m.Nodes = append(m.Nodes, Node{Name: e.name, M: parseNodeTransform(data, off)})
	}

	// SBC runs from sbcOffset to the material section.
	if sbcOff < matOff && matOff <= len(data) {
		m.SBC = data[sbcOff:matOff]
	}

	// Materials: u16 tex-to-mat / u16 pal-to-mat list offsets, then the dict.
	texList := matOff + int(le.Uint16(data[matOff:]))
	palList := matOff + int(le.Uint16(data[matOff+2:]))
	mdict, err := parseDict(data, matOff+4)
	if err != nil {
		return m, fmt.Errorf("material dict: %w", err)
	}
	// Material record: +$04 diffAmb, +$0C polyAttr, +$14 texImageParam (the GX
	// register with the repeat/flip addressing bits), +$20/+$22 u16 width/height.
	for _, e := range mdict {
		off := matOff + int(le.Uint32(padded(e.data)))
		mat := Material{Name: e.name}
		if off+0x2C <= len(data) {
			mat.TexParam = le.Uint32(data[off+0x14:])
			mat.Width = int(le.Uint16(data[off+0x20:]))
			mat.Height = int(le.Uint16(data[off+0x22:]))
		}
		m.Materials = append(m.Materials, mat)
	}
	// The tex/pal-to-material lists carry the binding: a dict of texture (palette)
	// names, each entry pointing (relative to the material section) at the index
	// bytes of the materials that use it.
	bindNames(data, texList, matOff, len(m.Materials), func(matIdx int, texName string) {
		m.Materials[matIdx].Texture = texName
	})
	bindNames(data, palList, matOff, len(m.Materials), func(matIdx int, palName string) {
		m.Materials[matIdx].Palette = palName
	})

	// Shapes: dict → {dlOffset, dlSize} records.
	sdict, err := parseDict(data, shpOff)
	if err != nil {
		return m, fmt.Errorf("shape dict: %w", err)
	}
	// A shape record is {u16 itemTag, u16 recSize, u32 flags, u32 dlOffset, u32
	// dlSize} with dlOffset relative to the record itself — pinned on the smallest
	// model, whose DL then ends exactly at end-of-file.
	for _, e := range sdict {
		off := shpOff + int(le.Uint32(padded(e.data)))
		if off+16 > len(data) {
			continue
		}
		dlo := off + int(le.Uint32(data[off+8:]))
		dls := int(le.Uint32(data[off+12:]))
		if dlo+dls > len(data) {
			continue
		}
		m.Shapes = append(m.Shapes, Shape{Name: e.name, DL: data[dlo : dlo+dls]})
	}
	return m, nil
}

// bindNames walks a tex/pal-to-material list: a resource dict whose entries are
// {u16 offset, u8 count, u8 bound}, the offset relative to the material section
// and pointing at count material-index bytes.
func bindNames(data []byte, listOff, matOff, nMats int, set func(matIdx int, name string)) {
	dict, err := parseDict(data, listOff)
	if err != nil {
		return
	}
	for _, e := range dict {
		if len(e.data) < 4 {
			continue
		}
		off := matOff + int(le.Uint16(e.data))
		count := int(e.data[2])
		for i := 0; i < count && off+i < len(data); i++ {
			if idx := int(data[off+i]); idx < nMats {
				set(idx, e.name)
			}
		}
	}
}

// parseNodeTransform decodes a node's packed TRS. A u16 flags word (bit0: no
// translation, bit1: no rotation, bit2: no scale, bit3: pivot-compressed rotation)
// is followed by the present fields in T, R, S order; the full-3x3 rotation's first
// element rides in the header's second u16.
func parseNodeTransform(data []byte, off int) Mat43 {
	if off+4 > len(data) {
		return Identity43()
	}
	flags := le.Uint16(data[off:])
	m00 := fx16(le.Uint16(data[off+2:]))
	p := off + 4

	tx, ty, tz := 0.0, 0.0, 0.0
	if flags&1 == 0 { // translation
		tx = fx32(le.Uint32(data[p:]))
		ty = fx32(le.Uint32(data[p+4:]))
		tz = fx32(le.Uint32(data[p+8:]))
		p += 12
	}
	rot := Identity43()
	if flags&2 == 0 { // rotation
		if flags&8 != 0 { // pivot-compressed: A,B + pivot index/signs in the flags
			a := fx16(le.Uint16(data[p:]))
			b := fx16(le.Uint16(data[p+2:]))
			p += 4
			rot = pivotMatrix(int(flags>>4)&0xF, flags&(1<<8) != 0, flags&(1<<9) != 0, flags&(1<<10) != 0, a, b)
		} else { // full 3x3, m00 from the header slot
			v := [9]float64{m00}
			for i := 1; i < 9; i++ {
				v[i] = fx16(le.Uint16(data[p:]))
				p += 2
			}
			rot = Mat43{v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], 0, 0, 0}
		}
	}
	if flags&4 == 0 { // scale
		sx := fx32(le.Uint32(data[p:]))
		sy := fx32(le.Uint32(data[p+4:]))
		sz := fx32(le.Uint32(data[p+8:]))
		for c := 0; c < 3; c++ {
			rot[0*3+c] *= sx
			rot[1*3+c] *= sy
			rot[2*3+c] *= sz
		}
	}
	rot[9], rot[10], rot[11] = tx, ty, tz
	return rot
}

// pivotMatrix reconstructs a pivot-compressed rotation: one element of the 3x3 is
// ±1 (the pivot, its index in flags bits 4-7; bit 8 negates it), and the remaining
// 2x2 is {A, B; C, D} with C = ±B (bit 9) and D = ±A (bit 10).
func pivotMatrix(pivot int, neg, revC, revD bool, a, b float64) Mat43 {
	one := 1.0
	if neg {
		one = -1
	}
	c, d := b, a
	if revC {
		c = -b
	}
	if revD {
		d = -a
	}
	m := Mat43{}
	row, col := pivot/3, pivot%3
	r1, r2 := other(row)
	c1, c2 := other(col)
	m[row*3+col] = one
	m[r1*3+c1] = a
	m[r1*3+c2] = b
	m[r2*3+c1] = c
	m[r2*3+c2] = d
	return m
}

func other(i int) (int, int) {
	switch i {
	case 0:
		return 1, 2
	case 1:
		return 0, 2
	default:
		return 0, 1
	}
}

// DumpModel summarises a model (for the inspection tool).
func DumpModel(m Model) string {
	var b strings.Builder
	fmt.Fprintf(&b, "model %q  scale=%g  %d nodes, %d materials, %d shapes, SBC %d bytes\n",
		m.Name, m.UpScale, len(m.Nodes), len(m.Materials), len(m.Shapes), len(m.SBC))
	for i, n := range m.Nodes {
		fmt.Fprintf(&b, "  node %d %-16s T=(%.3f %.3f %.3f)\n", i, n.Name, n.M[9], n.M[10], n.M[11])
	}
	for i, mat := range m.Materials {
		fmt.Fprintf(&b, "  mat  %d %-16s tex=%q pal=%q %dx%d param=%08X\n", i, mat.Name, mat.Texture, mat.Palette, mat.Width, mat.Height, mat.TexParam)
	}
	for i, s := range m.Shapes {
		fmt.Fprintf(&b, "  shp  %d %-16s DL %d bytes\n", i, s.Name, len(s.DL))
	}
	return b.String()
}
