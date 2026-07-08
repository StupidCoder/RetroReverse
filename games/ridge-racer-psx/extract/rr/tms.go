package rr

// tms.go decodes the TEX*.TMS texture archives. A TMS file is the game's
// VRAM upload schedule, streamed one block per frame by the routine at
// 0x80037DB8: after a format word the file is a chain of blocks, each holding
// an optional 16-colour CLUT strip and one image rectangle, both as raw
// {x, y, w, h, pixels} VRAM rects (w in 16-bit VRAM words). Rectangles whose
// pixel area exceeds 2048 words are uploaded in eight horizontal slices, but
// that is purely the streamer's pacing — the file stores every rectangle
// whole.
//
//	u32 format            ; 0x100
//	repeat:
//	  u32 blockLen        ; next block at +4+(blockLen&~3); ≤0 ends the file
//	  u32 flags           ; bit 3: a CLUT record precedes the image record
//	  if flags&8:
//	    u32 len           ; record byte length (12 + w*h*2)
//	    u16 x, y, w, h    ; destination rect in VRAM
//	    u16 pixels[w*h]
//	  u32 len             ; image record, same shape; skipped when w or h is 0
//	  u16 x, y, w, h
//	  u16 pixels[w*h]

import "fmt"

// Rect is one VRAM upload: a destination rectangle and its 16-bit pixels.
// CLUT strips are ordinary rects (16×1 for the 4-bit palettes here).
type Rect struct {
	X, Y, W, H int
	Pix        []uint16 // W*H words, row-major
	CLUT       bool     // came from a block's CLUT slot
}

const tmsFormat = 0x100

// ParseTMS decodes every upload rect of one TMS file, in file order.
func ParseTMS(d []byte) ([]Rect, error) {
	if len(d) < 8 || u32(d, 0) != tmsFormat {
		return nil, fmt.Errorf("tms: bad format word")
	}
	var out []Rect
	p := 4
	for p+12 <= len(d) {
		blen := int(int32(u32(d, p)))
		if blen <= 0 {
			break
		}
		end := p + 4 + blen&^3
		if end > len(d) {
			return nil, fmt.Errorf("tms: block at 0x%X runs past EOF", p)
		}
		s := p + 12
		if u32(d, p+8)&8 != 0 {
			r, size, err := rectAt(d, s, true)
			if err != nil {
				return nil, err
			}
			out = append(out, r)
			s += size &^ 3
		}
		if w, h := int(u16(d, s+8)), int(u16(d, s+10)); w > 0 && h > 0 {
			r, _, err := rectAt(d, s, false)
			if err != nil {
				return nil, err
			}
			out = append(out, r)
		}
		p = end
	}
	return out, nil
}

// rectAt decodes one {len, x, y, w, h, pixels} record. It returns the record's
// own length word so a CLUT record can be skipped exactly as the game does.
func rectAt(d []byte, s int, clut bool) (Rect, int, error) {
	if s+12 > len(d) {
		return Rect{}, 0, fmt.Errorf("tms: record header at 0x%X past EOF", s)
	}
	size := int(u32(d, s))
	r := Rect{
		X: int(u16(d, s+4)), Y: int(u16(d, s+6)),
		W: int(u16(d, s+8)), H: int(u16(d, s+10)),
		CLUT: clut,
	}
	n := r.W * r.H
	if s+12+n*2 > len(d) {
		return Rect{}, 0, fmt.Errorf("tms: record at 0x%X: %d pixels past EOF", s, n)
	}
	r.Pix = make([]uint16, n)
	for i := range r.Pix {
		r.Pix[i] = u16(d, s+12+i*2)
	}
	return r, size, nil
}
