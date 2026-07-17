package threedo

import (
	"encoding/binary"
	"testing"
)

// be16/be32 append big-endian values, the byte order of the Stream/cvid format.
func ap16(b []byte, v uint16) []byte { return binary.BigEndian.AppendUint16(b, v) }
func ap32(b []byte, v uint32) []byte { return binary.BigEndian.AppendUint32(b, v) }

// buildCvidFrame hand-encodes an 8x8 intra cvid frame: one strip with a 6-byte
// V1 codebook (two entries) and an all-V1 (0x3200) image selecting them across
// the four 4x4 macroblocks. This exercises codebook routing, V1 painting and the
// YUV->RGB math without any dependency on the (git-ignored) disc image.
func buildCvidFrame() []byte {
	// V1 codebook chunk 0x2200 (hi=0x22 => V1; 0x0400 clear => 6-byte entries).
	cb := ap16(nil, 0x2200)
	cb = ap16(cb, 0) // size patched below
	cb = append(cb, 10, 20, 30, 40, 0, 0)              // entry 0: grey ramp, no chroma
	cb = append(cb, 100, 100, 100, 100, 10, 0xF6)      // entry 1: u=10, v=-10
	binary.BigEndian.PutUint16(cb[2:], uint16(len(cb)))

	// Image chunk 0x3200 (all-V1 intra): one index byte per 4x4 macroblock.
	im := ap16(nil, 0x3200)
	im = ap16(im, 0)
	im = append(im, 0, 1, 1, 0) // MBs: (0,0)=e0 (4,0)=e1 (0,4)=e1 (4,4)=e0
	binary.BigEndian.PutUint16(im[2:], uint16(len(im)))

	// Strip 0x1000 covering the whole 8x8, wrapping the two chunks.
	strip := ap16(nil, 0x1000)
	strip = ap16(strip, 0)    // size patched below
	strip = ap16(strip, 0)    // ytop
	strip = ap16(strip, 0)    // xleft
	strip = ap16(strip, 8)    // ybottom
	strip = ap16(strip, 8)    // xright
	strip = append(strip, cb...)
	strip = append(strip, im...)
	binary.BigEndian.PutUint16(strip[2:], uint16(len(strip)))

	// Frame header: flags(1) len(3) width(2) height(2) strips(2), then the strip.
	fr := []byte{0x00}
	total := 10 + len(strip)
	fr = append(fr, byte(total>>16), byte(total>>8), byte(total))
	fr = ap16(fr, 8) // width
	fr = ap16(fr, 8) // height
	fr = ap16(fr, 1) // strips
	fr = append(fr, strip...)
	return fr
}

func TestCvidDecodeFrame(t *testing.T) {
	d := NewCvidDecoder(8, 8)
	d.DecodeFrame(buildCvidFrame())
	img := d.Frame()

	rgb := func(x, y int) (int, int, int) {
		o := img.PixOffset(x, y)
		return int(img.Pix[o]), int(img.Pix[o+1]), int(img.Pix[o+2])
	}

	// Macroblock (0,0) is entry 0: y[0]=10 fills the top-left 2x2, chroma 0.
	if r, g, b := rgb(0, 0); r != 10 || g != 10 || b != 10 {
		t.Fatalf("MB0 luma[0] = (%d,%d,%d), want (10,10,10)", r, g, b)
	}
	// Within entry 0: y[1]=20 -> (2,0), y[2]=30 -> (0,2), y[3]=40 -> (2,2).
	if r, _, _ := rgb(2, 0); r != 20 {
		t.Fatalf("MB0 luma[1] r=%d, want 20", r)
	}
	if r, _, _ := rgb(0, 2); r != 30 {
		t.Fatalf("MB0 luma[2] r=%d, want 30", r)
	}
	// Macroblock (4,0) is entry 1: y=100, u=10, v=-10. Cinepak YUV->RGB:
	//   r=100+2*(-10)=80, g=100-(10/2)-(-10)=105, b=100+2*10=120.
	if r, g, b := rgb(4, 0); r != 80 || g != 105 || b != 120 {
		t.Fatalf("MB(4,0) = (%d,%d,%d), want (80,105,120)", r, g, b)
	}
	// Alpha is always opaque.
	if a := img.Pix[img.PixOffset(0, 0)+3]; a != 255 {
		t.Fatalf("alpha = %d, want 255", a)
	}
}

func TestDemuxStream(t *testing.T) {
	frame := buildCvidFrame()

	// Wrap the frame in a minimal FILM/FHDR + FILM/FRME .Stream container.
	film := func(sub string, body []byte) []byte {
		inner := ap32(nil, 0)     // time
		inner = ap32(inner, 0)    // channel
		inner = append(inner, sub...)
		inner = append(inner, body...)
		out := append([]byte("FILM"), 0, 0, 0, 0)
		binary.BigEndian.PutUint32(out[4:], uint32(8+len(inner)))
		return append(out, inner...)
	}
	// FHDR body: version(4) codec(4) height(4) width(4).
	hdrBody := ap32(nil, 0)
	hdrBody = append(hdrBody, []byte("cvid")...)
	hdrBody = ap32(hdrBody, 8)
	hdrBody = ap32(hdrBody, 8)
	// FRME body: duration(4) frameSize(4) cvid-bitstream.
	frmBody := ap32(nil, 40)
	frmBody = ap32(frmBody, uint32(len(frame)))
	frmBody = append(frmBody, frame...)

	stream := append(film("FHDR", hdrBody), film("FRME", frmBody)...)

	mv, err := DemuxStream(stream)
	if err != nil {
		t.Fatal(err)
	}
	if mv.Width != 8 || mv.Height != 8 || mv.Codec != "cvid" {
		t.Fatalf("demux dims/codec = %dx%d %q, want 8x8 cvid", mv.Width, mv.Height, mv.Codec)
	}
	if len(mv.Frames) != 1 {
		t.Fatalf("got %d frames, want 1", len(mv.Frames))
	}
	if string(mv.Frames[0]) != string(frame) {
		t.Fatalf("demuxed frame bytes differ from input")
	}
}
