// model.go decodes the Crazy Taxi model family — the one format behind
// POLDC*.BIN (course geometry), the 1,014 BINC*.AFS entries (streamed
// objects) and the character/car models. The layout was read off the
// game's own renderer (the walker at 0C080E20/0C080E44 and its vertex
// loops at 0C081200/0C081C00, traced live from the drive savestate), not
// guessed from byte shapes:
//
//	MODEL  := [u32 kind][u32 sub][4×f32 bounding sphere x,y,z,r] STREAM
//	STREAM := word w — three cases, read exactly as the walker does:
//	          w == 0          end of model
//	          w bit31 set     BLOCK header, 80 bytes total:
//	                          [PCW][ISP][TSP][TCW]   TA parameter templates
//	                          [4×f32 sphere]          per-block cull sphere
//	                          [u32 aux][u32 mode]     mode's low byte: -1 flat,
//	                                                  -2/-3 runtime callbacks,
//	                                                  >=0 light table index
//	                          [f32 intensity]
//	                          [4×f32 base RGBA][4×f32 offset RGBA]
//	                          [u32 size]              byte size of the strip
//	                                                  stream that follows —
//	                                                  the renderer's cull-skip
//	          w > 0           STRIP: [flags=w][u32 count] + count vertex
//	                          records (3×count when flags bit3: triangle
//	                          list, the 0C081C00 loop)
//	VERTEX := first word LSB set → 32 bytes inline [x y z nx ny nz u v]
//	          (the flag lives in the LSB of x's mantissa);
//	          LSB clear → [ctl][s32 off] and the 32-byte record sits at
//	          off bytes past the pair (a back-reference into vertices
//	          already in the stream).
//
// The PCW/ISP/TSP/TCW words are real PowerVR TA words — list type, blend
// mode and the texture control word (with the VRAM address the TEXDC
// loader baked) come straight out of them.
package assets

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ModelVert is one decoded vertex: position, normal and UV.
type ModelVert struct {
	Pos     [3]float32
	Normal  [3]float32
	U, V    float32
	Indexed bool // came from a back-reference, not inline
}

// Strip is one [flags][count] batch. Bit3 of Flags makes it a triangle
// list (count triples); otherwise it is a triangle strip.
type Strip struct {
	Flags uint32
	Verts []ModelVert
}

// Tris expands the strip into triangle index triples over Verts.
func (s *Strip) Tris() [][3]int {
	var out [][3]int
	if s.Flags&8 != 0 {
		for i := 0; i+2 < len(s.Verts); i += 3 {
			out = append(out, [3]int{i, i + 1, i + 2})
		}
		return out
	}
	for i := 0; i+2 < len(s.Verts); i++ {
		if i%2 == 0 {
			out = append(out, [3]int{i, i + 1, i + 2})
		} else {
			out = append(out, [3]int{i + 1, i, i + 2})
		}
	}
	return out
}

// ModelBlock carries one 80-byte block header and its strips.
type ModelBlock struct {
	PCW, ISP, TSP, TCW uint32
	Sphere             [4]float32
	Aux                uint32
	Mode               int32
	Intensity          float32
	BaseColor          [4]float32
	OffsetColor        [4]float32
	Strips             []Strip
}

// ListType is the TA list this block renders into (PCW bits 24-26):
// 0 opaque, 1 opaque modifier, 2 translucent, 3 translucent modifier,
// 4 punch-through.
func (b *ModelBlock) ListType() int { return int(b.PCW >> 24 & 7) }

// Textured reports PCW bit3 — whether the TCW names a texture.
func (b *ModelBlock) Textured() bool { return b.PCW&8 != 0 }

// Model is one decoded model.
type Model struct {
	Kind, Sub uint32
	Sphere    [4]float32
	Blocks    []ModelBlock
	Size      int  // bytes consumed from the input
	Counted   bool // the trailer's vertex count matched what we read
}

func f32(u uint32) float32 { return math.Float32frombits(u) }

// OpenModel decodes one model from the head of data. It returns the
// model and the number of bytes consumed (so back-to-back models — the
// POLDC files — can be walked).
func OpenModel(data []byte) (*Model, error) {
	le := binary.LittleEndian
	if len(data) < 28 {
		return nil, fmt.Errorf("model: %d bytes is too short", len(data))
	}
	m := &Model{Kind: le.Uint32(data), Sub: le.Uint32(data[4:])}
	if m.Kind != 1 {
		return nil, fmt.Errorf("model: kind %d not supported (walker handles kind 1)", m.Kind)
	}
	for i := range m.Sphere {
		m.Sphere[i] = f32(le.Uint32(data[8+4*i:]))
	}
	pos := 24
	for {
		if pos+4 > len(data) {
			return nil, fmt.Errorf("model: stream ran off the end at +%#x", pos)
		}
		w := le.Uint32(data[pos:])
		if w == 0 {
			m.Size = pos + 4
			// models carry a trailer word: the total number of vertex
			// records, a free consistency check on the whole parse
			total := 0
			for _, b := range m.Blocks {
				for _, s := range b.Strips {
					total += len(s.Verts)
				}
			}
			if pos+8 <= len(data) && int(le.Uint32(data[pos+4:])) == total && total > 0 {
				m.Size = pos + 8
				m.Counted = true
			}
			return m, nil
		}
		if w&0x80000000 == 0 {
			return nil, fmt.Errorf("model: top-level word %08X at +%#x is neither block nor end", w, pos)
		}
		if pos+80 > len(data) {
			return nil, fmt.Errorf("model: truncated block header at +%#x", pos)
		}
		var b ModelBlock
		b.PCW, b.ISP, b.TSP, b.TCW = w, le.Uint32(data[pos+4:]), le.Uint32(data[pos+8:]), le.Uint32(data[pos+12:])
		for i := range b.Sphere {
			b.Sphere[i] = f32(le.Uint32(data[pos+16+4*i:]))
		}
		b.Aux = le.Uint32(data[pos+32:])
		b.Mode = int32(int8(data[pos+36]))
		b.Intensity = f32(le.Uint32(data[pos+40:]))
		for i := 0; i < 4; i++ {
			b.BaseColor[i] = f32(le.Uint32(data[pos+44+4*i:]))
			b.OffsetColor[i] = f32(le.Uint32(data[pos+60+4*i:]))
		}
		size := int(le.Uint32(data[pos+76:]))
		end := pos + 80 + size
		if size < 0 || end > len(data) {
			return nil, fmt.Errorf("model: block at +%#x claims %d strip bytes past the end", pos, size)
		}
		pos += 80
		for pos < end {
			flags := le.Uint32(data[pos:])
			if flags == 0 || flags&0x80000000 != 0 {
				return nil, fmt.Errorf("model: strip word %08X at +%#x inside block (%d bytes short of its size)", flags, pos, end-pos)
			}
			if pos+8 > end {
				return nil, fmt.Errorf("model: truncated strip header at +%#x", pos)
			}
			count := int(le.Uint32(data[pos+4:]))
			pos += 8
			n := count
			if flags&8 != 0 {
				n = count * 3
			}
			if n < 0 || n > (end-pos) {
				return nil, fmt.Errorf("model: strip at +%#x wants %d vertices, %d bytes left in block", pos-8, n, end-pos)
			}
			st := Strip{Flags: flags, Verts: make([]ModelVert, 0, n)}
			for i := 0; i < n; i++ {
				if pos+8 > end {
					return nil, fmt.Errorf("model: truncated vertex at +%#x", pos)
				}
				w0 := le.Uint32(data[pos:])
				at := pos
				indexed := w0&1 == 0
				if indexed {
					off := int(int32(le.Uint32(data[pos+4:])))
					at = pos + 8 + off
					pos += 8
				} else {
					pos += 32
				}
				if at < 0 || at+32 > len(data) || (indexed == false && at+32 > end+8) {
					return nil, fmt.Errorf("model: vertex record at +%#x out of range", at)
				}
				var v ModelVert
				v.Indexed = indexed
				v.Pos = [3]float32{f32(le.Uint32(data[at:])), f32(le.Uint32(data[at+4:])), f32(le.Uint32(data[at+8:]))}
				v.Normal = [3]float32{f32(le.Uint32(data[at+12:])), f32(le.Uint32(data[at+16:])), f32(le.Uint32(data[at+20:]))}
				v.U, v.V = f32(le.Uint32(data[at+24:])), f32(le.Uint32(data[at+28:]))
				st.Verts = append(st.Verts, v)
			}
			b.Strips = append(b.Strips, st)
		}
		if pos != end {
			return nil, fmt.Errorf("model: block strips consumed to +%#x, size says +%#x", pos, end)
		}
		m.Blocks = append(m.Blocks, b)
	}
}

// OpenModels walks back-to-back models (a POLDC file).
func OpenModels(data []byte) ([]*Model, error) {
	var out []*Model
	pos := 0
	for pos+28 <= len(data) {
		m, err := OpenModel(data[pos:])
		if err != nil {
			return out, fmt.Errorf("at +%#x: %w", pos, err)
		}
		out = append(out, m)
		pos += m.Size
		// files are padded with zero words between/after models
		for pos+4 <= len(data) && binary.LittleEndian.Uint32(data[pos:]) == 0 {
			pos += 4
		}
	}
	return out, nil
}
