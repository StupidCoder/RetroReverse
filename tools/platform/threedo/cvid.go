package threedo

// cvid.go decodes 3DO streamed FMV: the "Stream" file container plus the Cinepak
// ("cvid") video codec the 3DO SDK's DataStreamer shipped. On a stock 3DO there
// is no video decode hardware; the ARM60 runs Cinepak in software, so all 3DO
// FMV uses this same codec (audio alongside it is SDX2, handled elsewhere).
//
// A .Stream file is a flat sequence of 8-byte-headed chunks (4-char tag +
// big-endian size that includes the header). The ones that matter for video:
//
//	FILM  a film subchunk. After the 8-byte header come time(4) + channel(4)
//	      then a 4-char subtype: FHDR (film header: codec fourcc + height/width)
//	      or FRME (one frame: duration(4) + frameSize(4) + the cvid bitstream).
//	FILL/CTRL/SNDS  padding, stream control, and SDX2 audio — skipped here.
//
// The Cinepak bitstream itself is the de-facto-standard codec (vector
// quantisation over 4x4 macroblocks built from 2x2 codebook vectors); it is
// reimplemented here from the format, and the decoder is verified byte-identical
// to the reference (FFmpeg's cinepak) across every frame of the Need for Speed
// movies. Nothing here is derived from game-specific sources.

import (
	"encoding/binary"
	"fmt"
	"image"
)

// CvidMovie is a demuxed 3DO stream movie: its dimensions and the raw Cinepak
// bitstream of each video frame, in presentation order.
type CvidMovie struct {
	Width, Height int
	Codec         string   // FHDR compression fourcc; "cvid" for Cinepak
	FPS           int      // FHDR frame rate (3DO FMV is 15 fps)
	Frames        [][]byte // each entry is one FRME's cvid bitstream (from the flags byte)
	Durations     []uint32 // per-frame duration in the film's time units
}

// DemuxStream walks a .Stream file and pulls out the video (FILM) track: the
// film header's dimensions/codec and every frame's cvid bitstream. Audio and
// control chunks are ignored.
func DemuxStream(data []byte) (*CvidMovie, error) {
	m := &CvidMovie{Width: 320, Height: 240}
	haveHdr := false
	for off := 0; off+8 <= len(data); {
		tag := string(data[off : off+4])
		size := int(binary.BigEndian.Uint32(data[off+4 : off+8]))
		if size < 8 || off+size > len(data) {
			break // truncated or misaligned; stop at the last good chunk
		}
		if tag == "FILM" && off+20 <= len(data) {
			switch string(data[off+16 : off+20]) {
			case "FHDR":
				if off+40 <= len(data) {
					m.Codec = string(data[off+24 : off+28])
					m.Height = int(binary.BigEndian.Uint32(data[off+28 : off+32]))
					m.Width = int(binary.BigEndian.Uint32(data[off+32 : off+36]))
					m.FPS = int(binary.BigEndian.Uint32(data[off+36 : off+40]))
					haveHdr = true
				}
			case "FRME":
				if off+28 <= off+size && off+28 <= len(data) {
					dur := binary.BigEndian.Uint32(data[off+20 : off+24])
					fsz := int(binary.BigEndian.Uint32(data[off+24 : off+28]))
					end := off + 28 + fsz
					if end > off+size || end > len(data) {
						end = off + size
					}
					m.Frames = append(m.Frames, data[off+28:end])
					m.Durations = append(m.Durations, dur)
				}
			}
		}
		off += size
	}
	if !haveHdr && len(m.Frames) == 0 {
		return nil, fmt.Errorf("no FILM video track found in stream")
	}
	if m.FPS <= 0 {
		m.FPS = 15 // 3DO FMV default
	}
	return m, nil
}

// cvidVec is one Cinepak codebook entry: a 2x2 block of luma with shared,
// signed chroma. y[0..3] are top-left, top-right, bottom-left, bottom-right.
type cvidVec struct {
	y          [4]uint8
	u, v       int8
}

// CvidDecoder decodes a movie's frames one at a time, holding the state Cinepak
// carries between frames: the RGBA framebuffer (inter-frames leave skipped
// blocks untouched) and the per-strip codebooks (which persist across frames).
type CvidDecoder struct {
	W, H int
	img  *image.RGBA
	v1   [][]cvidVec // per strip index, 256 entries each
	v4   [][]cvidVec
}

// NewCvidDecoder makes a decoder for the given frame size.
func NewCvidDecoder(w, h int) *CvidDecoder {
	return &CvidDecoder{W: w, H: h, img: image.NewRGBA(image.Rect(0, 0, w, h))}
}

// Frame returns the decoder's current framebuffer. The returned image is reused
// and overwritten by the next DecodeFrame; copy it if you need to keep it.
func (d *CvidDecoder) Frame() *image.RGBA { return d.img }

// tdiv2 divides toward zero, matching the reference decoder's C `u/2`.
func tdiv2(x int) int {
	if x < 0 {
		return -((-x) / 2)
	}
	return x / 2
}

func clamp8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// setPix writes one RGB pixel from a luma+chroma sample using Cinepak's own
// integer YUV->RGB (r=y+2v, g=y-u/2-v, b=y+2u).
func (d *CvidDecoder) setPix(x, y int, ly uint8, u, v int8) {
	if x < 0 || y < 0 || x >= d.W || y >= d.H {
		return
	}
	Y := int(ly)
	o := d.img.PixOffset(x, y)
	d.img.Pix[o] = clamp8(Y + 2*int(v))
	d.img.Pix[o+1] = clamp8(Y - tdiv2(int(u)) - int(v))
	d.img.Pix[o+2] = clamp8(Y + 2*int(u))
	d.img.Pix[o+3] = 255
}

// paintV4 fills a 4x4 macroblock from four 2x2 codebook vectors (one per quadrant).
func (d *CvidDecoder) paintV4(px, py int, a, b, c, e cvidVec) {
	for j, vec := range [4]cvidVec{a, b, c, e} {
		ox, oy := px+(j&1)*2, py+(j>>1)*2
		d.setPix(ox, oy, vec.y[0], vec.u, vec.v)
		d.setPix(ox+1, oy, vec.y[1], vec.u, vec.v)
		d.setPix(ox, oy+1, vec.y[2], vec.u, vec.v)
		d.setPix(ox+1, oy+1, vec.y[3], vec.u, vec.v)
	}
}

// paintV1 fills a 4x4 macroblock from one codebook vector, each of its 2x2 luma
// samples expanded to a 2x2 output block.
func (d *CvidDecoder) paintV1(px, py int, vec cvidVec) {
	for j := 0; j < 4; j++ {
		ox, oy := px+(j&1)*2, py+(j>>1)*2
		c := vec.y[j]
		d.setPix(ox, oy, c, vec.u, vec.v)
		d.setPix(ox+1, oy, c, vec.u, vec.v)
		d.setPix(ox, oy+1, c, vec.u, vec.v)
		d.setPix(ox+1, oy+1, c, vec.u, vec.v)
	}
}

// decodeCodebook fills book from a codebook chunk. cid is the 16-bit chunk id;
// bit 0x0400 selects 4-byte (grayscale) entries, bit 0x0100 a selective update
// driven by a 32-bit flag mask.
func decodeCodebook(book []cvidVec, cid uint16, data []byte) {
	nbytes := 6
	if cid&0x0400 != 0 {
		nbytes = 4
	}
	flagged := cid&0x0100 != 0
	d := 0
	read := func(i int) bool {
		if d+nbytes > len(data) {
			return false
		}
		var e cvidVec
		e.y[0], e.y[1], e.y[2], e.y[3] = data[d], data[d+1], data[d+2], data[d+3]
		if nbytes == 6 {
			e.u, e.v = int8(data[d+4]), int8(data[d+5])
		}
		book[i] = e
		d += nbytes
		return true
	}
	if !flagged {
		for i := 0; i < 256 && read(i); i++ {
		}
		return
	}
	for i := 0; i < 256 && d < len(data); {
		if d+4 > len(data) {
			return
		}
		flag := binary.BigEndian.Uint32(data[d:])
		d += 4
		for b := 0; b < 32 && i < 256; b++ {
			if flag&0x80000000 != 0 {
				if !read(i) {
					return
				}
			}
			flag <<= 1
			i++
		}
	}
}

// decodeVectors paints one strip's image chunk. cid bit 0x0100 marks an
// inter-coded chunk (blocks may be skipped, keeping the previous frame); bit
// 0x0200 forces V1-only. A single flag/mask accumulator supplies both the
// skip bit and the V1/V4 bit, consumed in that order — matching the reference.
func (d *CvidDecoder) decodeVectors(cid uint16, data []byte, x0, y0, x1, y1 int, v1, v4 []cvidVec) {
	var flag, mask uint32
	pos := 0
	need := func() bool { // reload the flag word when the mask runs out
		mask >>= 1
		if mask == 0 {
			if pos+4 > len(data) {
				return false
			}
			flag = binary.BigEndian.Uint32(data[pos:])
			pos += 4
			mask = 0x80000000
		}
		return true
	}
	mbx := (x1 - x0) / 4
	mby := (y1 - y0) / 4
	inter := cid&0x0100 != 0
	v1only := cid&0x0200 != 0
	for by := 0; by < mby; by++ {
		for bx := 0; bx < mbx; bx++ {
			px, py := x0+bx*4, y0+by*4
			if inter {
				if !need() {
					return
				}
				if flag&mask == 0 {
					continue // skipped: keep previous framebuffer
				}
			}
			useV4 := false
			if !v1only {
				if !need() {
					return
				}
				useV4 = flag&mask != 0
			}
			if useV4 {
				if pos+4 > len(data) {
					return
				}
				d.paintV4(px, py, v4[data[pos]], v4[data[pos+1]], v4[data[pos+2]], v4[data[pos+3]])
				pos += 4
			} else {
				if pos+1 > len(data) {
					return
				}
				d.paintV1(px, py, v1[data[pos]])
				pos++
			}
		}
	}
}

// DecodeFrame decodes one cvid frame (bitstream from the flags byte) into the
// decoder's framebuffer, advancing the inter-frame and codebook state.
func (d *CvidDecoder) DecodeFrame(fr []byte) {
	if len(fr) < 10 {
		return
	}
	flags := fr[0]
	ln := int(fr[1])<<16 | int(fr[2])<<8 | int(fr[3])
	p := 10
	ystart := 0
	sidx := 0
	for p+4 <= len(fr) && p < 4+ln {
		sid := binary.BigEndian.Uint16(fr[p:])
		ssize := int(binary.BigEndian.Uint16(fr[p+2:]))
		if (sid != 0x1000 && sid != 0x1100) || ssize < 12 {
			if ssize < 4 {
				ssize = 4
			}
			p += ssize
			continue
		}
		ytop := int(binary.BigEndian.Uint16(fr[p+4:]))
		ybot := int(binary.BigEndian.Uint16(fr[p+8:]))
		y0 := ystart
		y1 := ystart + (ybot - ytop)
		for len(d.v1) <= sidx {
			d.v1 = append(d.v1, make([]cvidVec, 256))
		}
		for len(d.v4) <= sidx {
			d.v4 = append(d.v4, make([]cvidVec, 256))
		}
		// On an intra frame, strip i inherits strip i-1's codebooks before
		// applying its own (usually partial) updates.
		if sidx > 0 && flags&0x01 == 0 {
			copy(d.v1[sidx], d.v1[sidx-1])
			copy(d.v4[sidx], d.v4[sidx-1])
		}
		v1, v4 := d.v1[sidx], d.v4[sidx]
		end := p + ssize
		q := p + 12
		for q+4 <= end {
			cid := binary.BigEndian.Uint16(fr[q:])
			csize := int(binary.BigEndian.Uint16(fr[q+2:]))
			if csize < 4 || q+csize > end {
				break
			}
			payload := fr[q+4 : q+csize]
			hi := cid >> 8
			switch {
			case hi >= 0x20 && hi <= 0x27:
				book := v4
				if hi == 0x22 || hi == 0x23 || hi == 0x26 || hi == 0x27 {
					book = v1
				}
				decodeCodebook(book, cid, payload)
			case hi == 0x30 || hi == 0x31 || hi == 0x32:
				d.decodeVectors(cid, payload, 0, y0, d.W, y1, v1, v4)
			}
			q += csize
		}
		p = end
		ystart = y1
		sidx++
	}
}
