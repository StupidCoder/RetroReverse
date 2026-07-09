package rr

// obj.go decodes OBJ.RRO, the 3-D object models (cars, roadside scenery, the
// grandstand, blimp, …). The file is an object count, a directory of six
// per-object record counts, and every object's records back to back — six
// polygon-record types in a fixed order, each a GPU-ready quad in object
// space (directory walker at 0x80011E20; the per-type renderers at
// 0x80046B18/0x80046D10/… consume the records in place).
//
//	u32 count
//	count × { u16 n40,n48,n32,n64,n72,n56; u32 runtime }  ; 16 bytes
//	per object, in directory order:
//	  n40 × 40 bytes   ; textured flat quad        (same shape as TrackQuad)
//	  n48 × 48 bytes   ; textured flat quad + 8-byte tail
//	  n32 × 32 bytes   ; untextured flat quad      (colour + depth bias)
//	  n64 × 64 bytes   ; textured lit quad         (per-vertex normals, NCT)
//	  n72 × 72 bytes   ; textured lit quad + 8-byte tail (base colour)
//	  n56 × 56 bytes   ; untextured lit quad       (normals + colour word)
//
// The `runtime` directory word is zero on disc; the loader patches the
// object's geometry address into it.
//
// Textured records store their texture words exactly as the GPU packet wants
// them: {uv0,clut},{uv1,tpage},{uv2,bias},{uv3,…}. The CLUT halfword is a
// base: the renderer adds a per-instance offset when it builds the packet
// (the car liveries select palettes this way).

import "fmt"

// QuadF is an untextured flat-shaded quad (32-byte record).
type QuadF struct {
	V    [4][3]int16
	RGB  [3]uint8
	Bias int16
}

// QuadGT is a textured, vertex-lit quad (64- and 72-byte records). For the
// 72-byte type the trailing 8 bytes carry the base colour handed to the GTE;
// Extra keeps them raw.
type QuadGT struct {
	V     [4][3]int16
	N     [4][3]int16
	UV    [4]UV
	CLUT  uint16
	TPage uint16
	Bias  int16
	Extra [8]byte // 72-byte records only: a texture-window rectangle, as Quad48
}

// Window returns the texture window of a 72-byte lit record (its Extra tail is
// the same (u0, v0, w, h) rectangle as a 48-byte record). The 64-byte records
// leave Extra zero, which yields no window.
func (q QuadGT) Window() TexWindow {
	return windowFromRect(q.Extra)
}

// QuadG is an untextured, vertex-lit quad (56-byte record). RGB is the base
// colour of the ready-made GP0 colour word ({r,g,b,opcode 0x38}).
type QuadG struct {
	V    [4][3]int16
	N    [4][3]int16
	RGB  [3]uint8
	Op   uint8
	Bias int16
}

// Quad48 is a textured flat quad with an 8-byte tail (48-byte record). The
// tail is a texture-window rectangle in the page — (u0, v0, w, h) as the
// halfwords at +0,+2,+4,+6 — from which the drawing code builds the GP0 0xE2
// window before sampling, so that rectangle of the page repeats. The 40-byte
// records carry no window.
type Quad48 struct {
	TrackQuad
	Extra [8]byte
}

// Window returns the 48-byte record's texture window.
func (q Quad48) Window() TexWindow {
	return windowFromRect(q.Extra)
}

// windowFromRect maps a record tail's texture-window rectangle to the GP0 0xE2
// register's raw 5-bit fields. The tail halfwords are (u0, v0, w, h) in
// pixels: offset = origin/8, mask = (256-size)/8 (so a size of 256 → mask 0,
// no repeat).
func windowFromRect(extra [8]byte) TexWindow {
	return TexWindow{
		OffX:  (extra[0] / 8) & 0x1F,
		OffY:  (extra[2] / 8) & 0x1F,
		MaskX: byte(((256 - int(extra[4])) / 8) & 0x1F),
		MaskY: byte(((256 - int(extra[6])) / 8) & 0x1F),
	}
}

// Object is one model: its six record lists in file order.
type Object struct {
	FT  []TrackQuad // 40-byte textured flat
	FT8 []Quad48    // 48-byte textured flat + tail
	F   []QuadF     // 32-byte untextured flat
	GT  []QuadGT    // 64-byte textured lit
	GT8 []QuadGT    // 72-byte textured lit + colour tail
	G   []QuadG     // 56-byte untextured lit
}

// Quads returns how many polygon records the object holds.
func (o *Object) Quads() int {
	return len(o.FT) + len(o.FT8) + len(o.F) + len(o.GT) + len(o.GT8) + len(o.G)
}

// ParseRRO decodes an OBJ.RRO image.
func ParseRRO(d []byte) ([]Object, error) {
	if len(d) < 4 {
		return nil, errTruncated("rro", 4, len(d))
	}
	count := int(u32(d, 0))
	dir := 4
	off := dir + count*16
	if off > len(d) {
		return nil, errTruncated("rro", off, len(d))
	}
	objs := make([]Object, count)
	for i := range objs {
		e := dir + i*16
		n := [6]int{}
		for t := 0; t < 6; t++ {
			n[t] = int(u16(d, e+t*2))
		}
		need := off + n[0]*40 + n[1]*48 + n[2]*32 + n[3]*64 + n[4]*72 + n[5]*56
		if need > len(d) {
			return nil, fmt.Errorf("rro: object %d geometry runs past EOF", i)
		}
		o := &objs[i]
		o.FT = make([]TrackQuad, n[0])
		for k := range o.FT {
			o.FT[k] = trackQuadAt(d, off)
			off += 40
		}
		o.FT8 = make([]Quad48, n[1])
		for k := range o.FT8 {
			o.FT8[k].TrackQuad = trackQuadAt(d, off)
			copy(o.FT8[k].Extra[:], d[off+40:off+48])
			off += 48
		}
		o.F = make([]QuadF, n[2])
		for k := range o.F {
			q := &o.F[k]
			readVerts(d, off, &q.V)
			q.RGB = [3]uint8{d[off+24], d[off+25], d[off+26]}
			q.Bias = s16(d, off+28)
			off += 32
		}
		o.GT = make([]QuadGT, n[3])
		for k := range o.GT {
			o.GT[k] = quadGTAt(d, off)
			off += 64
		}
		o.GT8 = make([]QuadGT, n[4])
		for k := range o.GT8 {
			o.GT8[k] = quadGTAt(d, off)
			copy(o.GT8[k].Extra[:], d[off+64:off+72])
			off += 72
		}
		o.G = make([]QuadG, n[5])
		for k := range o.G {
			q := &o.G[k]
			readVerts(d, off, &q.V)
			readVerts(d, off+24, &q.N)
			q.RGB = [3]uint8{d[off+48], d[off+49], d[off+50]}
			q.Op = d[off+51]
			q.Bias = s16(d, off+52)
			off += 56
		}
	}
	return objs, nil
}

func readVerts(d []byte, o int, v *[4][3]int16) {
	for i := 0; i < 4; i++ {
		(*v)[i] = [3]int16{s16(d, o+i*6), s16(d, o+i*6+2), s16(d, o+i*6+4)}
	}
}

// quadGTAt decodes the shared front of the 64/72-byte lit textured records:
// verts, normals, then the four texture words (traced at 0x80046BE0-0x80046BEC
// and 0x80046DD8-0x80046DE4).
func quadGTAt(d []byte, off int) QuadGT {
	var q QuadGT
	readVerts(d, off, &q.V)
	readVerts(d, off+24, &q.N)
	q.UV[0] = uvOf(u16(d, off+48))
	q.CLUT = u16(d, off+50)
	q.UV[1] = uvOf(u16(d, off+52))
	q.TPage = u16(d, off+54)
	q.UV[2] = uvOf(u16(d, off+56))
	q.Bias = s16(d, off+58)
	q.UV[3] = uvOf(u16(d, off+60))
	return q
}
