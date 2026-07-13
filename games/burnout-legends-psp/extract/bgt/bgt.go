// Package bgt decodes Burnout Legends' tracks: the static.dat and streamed.dat
// files under PSP_GAME/USRDIR/Tracks/<REGION>/<TRACK>.
//
// A track is the same GPU-native, pointer-patched family as the vehicles (see
// package bgv): the engine loads a file at a base address, fixes up its pointer
// slots, and hands the geometry to the GE untouched — the vertex data in the
// file IS the vertex buffer the hardware reads, and the strip lists in the file
// ARE display-list fragments.
//
// What made the track look unlike the cars is that its offsets are not stored
// as plain file offsets. Every pointer is RELATIVE — to the record, the slot,
// or the section that owns it — so a track's structures are position
// independent, which is exactly what a streamed block needs. Searching the file
// for absolute offsets therefore found nothing, and searching RAM for the
// vertex bytes found the display list instead.
//
// static.dat is the always-resident half: the texture dictionary, the material
// table, and the geometry that is never streamed out.
//
//	header +0x04  file size
//	       +0x08  material table
//	       +0x16  texture count (u16)
//	       +0x18  texture table — one u32 record offset per texture
//	       +0x24  \
//	       +0x28   > geometry section lists
//	       +0x2C  /
//
//	section list  {u32 groupCount, u32 nodeArray} entries, each offset taken
//	              from the entry itself; the list ends where its first node
//	              array begins
//
//	node (0x50)   +0x00  a bounding box: three axis rows and a centre row
//	              +0x40  mesh record array (offset from this slot)
//	              +0x44  mesh count
//	              +0x48  u16 material index per mesh (offset from the section
//	                     list ENTRY that owns the node — not from the section)
//
//	mesh (0xA0)   +0x44  vertex count   +0x48  vertex stream (offset from the record)
//	              +0x50  world centre   +0x60  half-extent — the quantisation scale
//	              +0x88  material index (streamed blocks only)
//	              +0x8C  strip list (offset from the record)   +0x90  word count
//
// streamed.dat is a chain of blocks, two per streaming cell, each self-describing:
//
//	block  +0x04  cell index   +0x08  block size (the chain step)
//	       +0x50  mesh record array (offset from this slot)   +0x54  mesh count
//
// A streamed block carries no node layer: its mesh records name their material
// directly at +0x88, indexing static.dat's table.
//
// The transform was read out of the running game rather than guessed. The GE
// world matrix the engine loads for a mesh is diag(halfExtent) with the
// translation at centre (less a global origin rebase the engine applies to keep
// float precision near the car), and the s16 positions are fractional, so:
//
//	world = centre + (position / 32768) * halfExtent
//
// The record's leading 4x4 looks like a transform and is not one — it is a tight
// oriented bounding box for culling. Using it puts every mesh in the wrong place.
package bgt

import (
	"encoding/binary"
	"fmt"
	"math"
)

// vtypeTrack is the vertex format every track mesh is stored in: u16 texcoords,
// a 5551 colour, and s16 positions — 12 bytes to a vertex.
const (
	vtypeTrack  = 0x116
	classTrack  = 4 // the mesh record's vertex-format class for vtype 0x116
	vertexSize  = 12
	meshRecSize = 0xA0
	nodeSize    = 0x50

	// texScale is the TEXSCALEU/V the engine programs for track meshes: the
	// fractional u16 texcoords cover sixteen tiles of the texture.
	texScale = 16.0
)

// Vertex is one decoded vertex, in world units.
type Vertex struct {
	X, Y, Z    float32
	U, V       float32
	R, G, B, A uint8
}

// Strip is one triangle strip: a vertex run in the mesh's vertex slice.
type Strip struct {
	Type  int // GE primitive type; track meshes are all 4 (triangle strip)
	Start int
	Count int
}

// Mesh is one drawable piece of the track, already placed in world space.
type Mesh struct {
	Offset int // the record's offset in its file, for diagnostics
	// VertOffset is where the vertex stream starts in the same file. The GE
	// fetches from exactly this address once the file is loaded, so it is what
	// ties a decoded mesh to a primitive the running game drew.
	VertOffset int
	Material   int // index into Track.Materials, or -1
	Centre     [3]float32
	Extent     [3]float32
	Verts      []Vertex
	Strips     []Strip
}

// Block is one streamed geometry block: the meshes of one cell of the world.
//
// A block also carries its cell's anchor — a point at ground level with a
// radius, which the streamer uses to decide when the cell comes in. Taken in
// order the anchors trace the circuit, which is the only road-level path through
// the world the files give up without reversing the game's route data.
type Block struct {
	Offset int
	Cell   int
	Anchor [3]float32
	Radius float32
	Points [4][3]float32 // the four header points; Anchor is the first of them
	Meshes []Mesh
}

// Material names a texture and the parameters the engine draws it with.
type Material struct {
	Texture int // index into Track.Textures, or -1 when the material has none
	Param   float32
	Flags   uint32
}

// TexInfo describes one texture in the dictionary; Track.Texture decodes it.
type TexInfo struct {
	Offset int
	W, H   int
	BPP    int // 4, 8 (CLUT4/CLUT8) or 32 (RGBA8888)
}

// Track is a decoded static.dat: the dictionaries every mesh of the track
// shares, plus the geometry that is always resident. Streamed blocks are
// decoded separately by ParseStreamed and index back into these tables.
type Track struct {
	Materials []Material
	Textures  []TexInfo
	Meshes    []Mesh

	data []byte // static.dat, kept for texture decoding
}

// Tris flattens the mesh's strips into triangles, dropping the degenerate ones
// the strips use to stitch disjoint runs together.
func (m *Mesh) Tris() [][3]uint32 {
	var out [][3]uint32
	for _, s := range m.Strips {
		if s.Type != 4 {
			continue
		}
		for i := 0; i+2 < s.Count; i++ {
			a, b, c := uint32(s.Start+i), uint32(s.Start+i+1), uint32(s.Start+i+2)
			if i&1 == 1 {
				b, c = c, b
			}
			if a == b || b == c || a == c {
				continue
			}
			out = append(out, [3]uint32{a, b, c})
		}
	}
	return out
}

// Parse decodes a track's static.dat.
func Parse(data []byte) (*Track, error) {
	if len(data) < 0x40 {
		return nil, fmt.Errorf("bgt: too small (%d bytes)", len(data))
	}
	if size := int(le32(data, 0x04)); size != len(data) {
		return nil, fmt.Errorf("bgt: header size %d != file size %d", size, len(data))
	}
	t := &Track{data: data}
	if err := t.parseTextures(); err != nil {
		return nil, err
	}
	if err := t.parseMaterials(); err != nil {
		return nil, err
	}
	// The three section lists: two of them can name the same list, so take each
	// distinct one once.
	seen := map[int]bool{}
	for _, hdr := range []int{0x24, 0x28, 0x2C} {
		sec := int(le32(data, hdr))
		if sec <= 0 || sec+8 > len(data) || seen[sec] {
			continue
		}
		seen[sec] = true
		meshes, err := t.parseSection(sec)
		if err != nil {
			return nil, err
		}
		t.Meshes = append(t.Meshes, meshes...)
	}
	if len(t.Meshes) == 0 {
		return nil, fmt.Errorf("bgt: no geometry sections")
	}
	return t, nil
}

// parseTextures reads the texture dictionary: a count and a table of record
// offsets. Each record describes one CLUT4 or CLUT8 texture and its mip chain.
func (t *Track) parseTextures() error {
	data := t.data
	table := int(le32(data, 0x18))
	n := int(le16(data, 0x16))
	if table <= 0 || n <= 0 || table+4*n > len(data) {
		return fmt.Errorf("bgt: bad texture table (%d entries at 0x%X)", n, table)
	}
	for i := 0; i < n; i++ {
		rec := int(le32(data, table+4*i))
		if rec <= 0 || rec+texHeader > len(data) {
			return fmt.Errorf("bgt: texture %d record at 0x%X out of range", i, rec)
		}
		info := TexInfo{
			Offset: rec,
			W:      int(le32(data, rec+0x0C)),
			H:      int(le32(data, rec+0x10)),
			BPP:    int(le32(data, rec+0x14)),
		}
		switch info.BPP {
		case 4, 8, 32: // CLUT4, CLUT8, and a few true-colour RGBA8888 textures
		default:
			return fmt.Errorf("bgt: texture %d: unsupported depth %d", i, info.BPP)
		}
		if info.W <= 0 || info.H <= 0 || info.W > 1024 || info.H > 1024 {
			return fmt.Errorf("bgt: texture %d: implausible size %dx%d", i, info.W, info.H)
		}
		t.Textures = append(t.Textures, info)
	}
	return nil
}

// parseMaterials reads the material table. A material's +0x0C names a slot in
// an array that follows the table, and that slot holds a SIGNED offset, from
// the material record, to the texture record — so the binding resolves from the
// file alone, without the fixups the loader would apply.
func (t *Track) parseMaterials() error {
	data := t.data
	sec := int(le32(data, 0x08))
	if sec <= 0 || sec+0x28 > len(data) {
		return fmt.Errorf("bgt: bad material table at 0x%X", sec)
	}
	// The first material's slot pointer marks the end of the table: the slot
	// array is stored immediately after the last record.
	slots := sec + int(le32(data, sec+0x0C))
	if slots <= sec || slots > len(data) || (slots-sec)%0x28 != 0 {
		return fmt.Errorf("bgt: bad material slot array at 0x%X", slots)
	}
	byOffset := map[int]int{}
	for i, tx := range t.Textures {
		byOffset[tx.Offset] = i
	}
	for i := 0; i < (slots-sec)/0x28; i++ {
		rec := sec + i*0x28
		slot := rec + int(le32(data, rec+0x0C))
		m := Material{
			Texture: -1,
			Param:   f32(data, rec+0x08),
			Flags:   le32(data, rec+0x10),
		}
		if slot >= 0 && slot+4 <= len(data) {
			// A few materials are untextured: their slot points into the table
			// itself rather than at a texture record.
			if tex, ok := byOffset[rec+int(int32(le32(data, slot)))]; ok {
				m.Texture = tex
			}
		}
		t.Materials = append(t.Materials, m)
	}
	return nil
}

// parseSection walks one geometry section list: a run of {groupCount, nodeArray}
// entries that ends where the first node array begins.
func (t *Track) parseSection(sec int) ([]Mesh, error) {
	data := t.data
	first := sec + int(le32(data, sec+4))
	if first <= sec || first > len(data) {
		first = len(data) // an empty section: the loop below stops on its zero count
	}
	var out []Mesh
	for entry := sec; entry+8 <= first && entry+8 <= len(data); entry += 8 {
		count := int(le32(data, entry))
		// A section can be empty — the small crash junctions carry one that is.
		// Its single entry has no groups, and its offset word means nothing.
		if count == 0 {
			break
		}
		nodes := entry + int(le32(data, entry+4))
		if count < 0 || count > 4096 || nodes < 0 || nodes+count*nodeSize > len(data) {
			return nil, fmt.Errorf("bgt: section 0x%X: bad group at 0x%X", sec, entry)
		}
		for k := 0; k < count; k++ {
			node := nodes + k*nodeSize
			n := int(le32(data, node+0x44))
			recs := node + 0x40 + int(le32(data, node+0x40))
			mats := entry + int(le32(data, node+0x48))
			if n < 0 || n > 4096 || recs < 0 || recs+n*meshRecSize > len(data) ||
				mats < 0 || mats+2*n > len(data) {
				return nil, fmt.Errorf("bgt: node at 0x%X out of range", node)
			}
			for j := 0; j < n; j++ {
				mesh, err := t.parseMesh(data, recs+j*meshRecSize, int(le16(data, mats+2*j)))
				if err != nil {
					return nil, err
				}
				if mesh != nil {
					out = append(out, *mesh)
				}
			}
		}
	}
	return out, nil
}

// ParseStreamed decodes a track's streamed.dat: a chain of blocks, each holding
// a flat array of mesh records that index this track's material table.
func (t *Track) ParseStreamed(data []byte) ([]Block, error) {
	// The crash junctions stream nothing: their streamed.dat is an empty file.
	if len(data) == 0 {
		return nil, nil
	}
	var out []Block
	for off := 0; off+0x58 <= len(data); {
		size := int(le32(data, off+8))
		if size <= 0 || off+size > len(data) {
			break
		}
		b := Block{
			Offset: off,
			Cell:   int(le32(data, off+4)),
			Anchor: [3]float32{f32(data, off+0x10), f32(data, off+0x14), f32(data, off+0x18)},
			Radius: f32(data, off+0x1C),
		}
		for k := 0; k < 4; k++ {
			b.Points[k] = [3]float32{
				f32(data, off+0x10+k*16), f32(data, off+0x14+k*16), f32(data, off+0x18+k*16),
			}
		}
		recs := off + 0x50 + int(le32(data, off+0x50))
		n := int(le32(data, off+0x54))
		if n < 0 || n > 4096 || recs < off || recs+n*meshRecSize > off+size {
			return nil, fmt.Errorf("bgt: block at 0x%X: bad mesh array (%d at 0x%X)", off, n, recs)
		}
		for j := 0; j < n; j++ {
			rec := recs + j*meshRecSize
			mesh, err := t.parseMesh(data, rec, int(le32(data, rec+0x88)))
			if err != nil {
				return nil, err
			}
			if mesh != nil {
				b.Meshes = append(b.Meshes, *mesh)
			}
		}
		out = append(out, b)
		off += size
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("bgt: no streamed blocks")
	}
	return out, nil
}

// parseMesh reads one mesh record and dequantises its vertices into world space.
// The strip list is a run of GE PRIM words ending in RET; the vertices behind it
// are consumed sequentially, which is why one vertex address serves the run.
func (t *Track) parseMesh(data []byte, rec, material int) (*Mesh, error) {
	if rec < 0 || rec+meshRecSize > len(data) {
		return nil, fmt.Errorf("bgt: mesh record at 0x%X out of range", rec)
	}
	var (
		class  = int(le32(data, rec+0x40))
		count  = int(le32(data, rec+0x44))
		vOff   = rec + int(le32(data, rec+0x48))
		sOff   = rec + int(le32(data, rec+0x8C))
		sWords = int(le32(data, rec+0x90))
	)
	// An empty slot: the record array is sized for the block, and the tail
	// entries of some blocks carry no geometry. They are class 1, not the
	// class 4 (vertex type 0x116) every real track mesh uses.
	if count == 0 && sWords == 0 {
		return nil, nil
	}
	if class != classTrack {
		return nil, fmt.Errorf("bgt: mesh 0x%X: unknown vertex class %d", rec, class)
	}
	if count < 0 || sWords <= 0 || vOff < 0 || sOff < 0 ||
		vOff+count*vertexSize > len(data) || sOff+4*sWords > len(data) {
		return nil, fmt.Errorf("bgt: mesh 0x%X out of range (%d verts, %d words)", rec, count, sWords)
	}
	if material < 0 || material >= len(t.Materials) {
		material = -1
	}

	m := &Mesh{Offset: rec, VertOffset: vOff, Material: material}
	for i := 0; i < 3; i++ {
		m.Centre[i] = f32(data, rec+0x50+4*i)
		m.Extent[i] = f32(data, rec+0x60+4*i)
	}

	next := 0
	for i := 0; i < sWords; i++ {
		w := le32(data, sOff+4*i)
		switch w >> 24 {
		case 0x0B: // RET — the end of the display-list fragment
			i = sWords
			continue
		case 0x04: // PRIM
		default:
			return nil, fmt.Errorf("bgt: mesh 0x%X: word %d is 0x%08X, not a PRIM", rec, i, w)
		}
		n := int(w & 0xFFFF)
		m.Strips = append(m.Strips, Strip{Type: int((w >> 16) & 7), Start: next, Count: n})
		next += n
	}
	if next != count {
		return nil, fmt.Errorf("bgt: mesh 0x%X: strips consume %d vertices, record says %d",
			rec, next, count)
	}

	m.Verts = make([]Vertex, count)
	for i := 0; i < count; i++ {
		p := vOff + i*vertexSize
		u := float32(le16(data, p)) / 32768 * texScale
		v := float32(le16(data, p+2)) / 32768 * texScale
		r, g, b, a := decode5551(le16(data, p+4))
		x := float32(int16(le16(data, p+6))) / 32768
		y := float32(int16(le16(data, p+8))) / 32768
		z := float32(int16(le16(data, p+10))) / 32768
		m.Verts[i] = Vertex{
			X: m.Centre[0] + x*m.Extent[0],
			Y: m.Centre[1] + y*m.Extent[1],
			Z: m.Centre[2] + z*m.Extent[2],
			U: u, V: v,
			R: r, G: g, B: b, A: a,
		}
	}
	return m, nil
}

// decode5551 unpacks the GE's 16-bit vertex colour: five bits each of red,
// green and blue from the low end, and one bit of alpha at the top.
func decode5551(p uint16) (r, g, b, a uint8) {
	ext := func(v uint16) uint8 { return uint8(uint32(v&0x1F) * 255 / 31) }
	a = 0
	if p&0x8000 != 0 {
		a = 0xFF
	}
	return ext(p), ext(p >> 5), ext(p >> 10), a
}

func le16(b []byte, off int) uint16 { return binary.LittleEndian.Uint16(b[off:]) }
func le32(b []byte, off int) uint32 { return binary.LittleEndian.Uint32(b[off:]) }
func f32(b []byte, off int) float32 { return math.Float32frombits(le32(b, off)) }
