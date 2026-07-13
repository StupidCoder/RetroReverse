package server

import (
	"encoding/binary"
	"fmt"
	"image"

	"retroreverse.com/tools/debug"
)

// The wire protocol.
//
// Client to server is always a JSON text message: {"op":"scrub","seq":9,"k":412}.
// Server to client is JSON text for structure, and a binary message for pixels —
// a raw framebuffer is incompressible, and base64 in JSON would cost a third more
// bytes and a decode pass for nothing.
//
// A binary message is a 16-byte header followed by the payload:
//
//	off  size  field
//	 0    4    magic "RDB1"
//	 4    1    kind (kindImage | kindProv)
//	 5    1    flags (reserved, 0)
//	 6    2    seq — echoes the request that asked for it, so a stale reply from an
//	           abandoned scrub can be dropped by the page
//	 8    4    width
//	12    4    height
//	16    ...  payload
//
// All little-endian. The header is 16 bytes so the payload lands 4-byte aligned and
// the browser can wrap the received ArrayBuffer in a Uint8ClampedArray (for
// ImageData) or an Int32Array (for provenance) with no copy.
const (
	hdrSize = 16

	kindImage byte = 1 // payload: width*height*4 bytes, RGBA8888
	kindProv  byte = 2 // payload: width*height int32, the command that last wrote each pixel (-1 = none)
)

// req is a client request. Only the fields an op uses are set.
type req struct {
	Op   string `json:"op"`
	Seq  int    `json:"seq"`
	K    int    `json:"k"`        // scrub: command index
	X    int    `json:"x"`        // pixel: coordinates
	Y    int    `json:"y"`        //
	Addr uint32 `json:"addr"`     // mem: physical address
	Len  int    `json:"len"`      // mem: byte count
	Over bool   `json:"overdraw"` // step: record the full overdraw history
	N    int    `json:"n"`        // step: fields to advance (0 = step to a drawn frame)
}

// coalescable reports whether a queued request of this op may be replaced by a
// newer one. A scrubber drag fires an op per mouse-move and only the last one
// matters; a step or a memory read must not be dropped.
func coalescable(op string) bool { return op == "scrub" }

// jsonCommand is a GPUCommand as the page sees it. Words are hex strings: a raw
// uint64 would lose its low bits passing through a JSON double.
type jsonCommand struct {
	I       int      `json:"i"`
	Name    string   `json:"name"`
	Op      uint32   `json:"op"`
	Words   []string `json:"words"`
	Decoded string   `json:"decoded,omitempty"`
}

// jsonWrite is a PixelWrite as the page sees it.
type jsonWrite struct {
	Cmd      int    `json:"cmd"`
	Name     string `json:"name"`
	R        uint8  `json:"r"`
	G        uint8  `json:"g"`
	B        uint8  `json:"b"`
	A        uint8  `json:"a"`
	Rejected bool   `json:"rejected"`
}

// The server-to-client JSON messages. Each carries Type and, when it answers a
// request, that request's Seq.
type (
	helloMsg struct {
		Type   string `json:"type"` // "hello"
		Target string `json:"target"`
		ROM    string `json:"rom"`
	}

	frameMsg struct {
		Type     string        `json:"type"` // "frame"
		Seq      int           `json:"seq"`
		Frame    int           `json:"frame"` // how many frames we have stepped
		Commands []jsonCommand `json:"commands"`
		W        int           `json:"w"`
		H        int           `json:"h"`
		HasProv  bool          `json:"hasProv"`
		Overdraw bool          `json:"overdraw"`
		StepMs   float64       `json:"stepMs"`
	}

	renderMsg struct {
		Type     string  `json:"type"` // "render" — announces the binary image that follows
		Seq      int     `json:"seq"`
		K        int     `json:"k"`
		RenderMs float64 `json:"renderMs"`
		Bytes    int     `json:"bytes"`
		Cached   bool    `json:"cached"`
	}

	cpuMsg struct {
		Type  string            `json:"type"` // "cpu"
		Seq   int               `json:"seq"`
		PC    string            `json:"pc"`
		Names []string          `json:"names"`
		Vals  []string          `json:"vals"`
		Extra map[string]string `json:"extra"`
	}

	memMsg struct {
		Type string `json:"type"` // "mem"
		Seq  int    `json:"seq"`
		Addr uint32 `json:"addr"`
		Data string `json:"data"` // hex
	}

	pixelMsg struct {
		Type   string      `json:"type"` // "pixel"
		Seq    int         `json:"seq"`
		X      int         `json:"x"`
		Y      int         `json:"y"`
		Cmd    int         `json:"cmd"` // last writer, -1 if none
		Writes []jsonWrite `json:"writes"`
	}

	errMsg struct {
		Type string `json:"type"` // "error"
		Seq  int    `json:"seq"`
		Msg  string `json:"msg"`
	}
)

// encodeImage packs an RGBA image into a binary message.
func encodeImage(seq int, img *image.RGBA) []byte {
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	out := make([]byte, hdrSize+w*h*4)
	putHeader(out, kindImage, seq, w, h)
	// image.NewRGBA gives Stride == 4*w, but a sub-image would not; copy row-wise so
	// the payload is always tightly packed.
	for y := 0; y < h; y++ {
		src := img.Pix[y*img.Stride : y*img.Stride+w*4]
		copy(out[hdrSize+y*w*4:], src)
	}
	return out
}

// encodeProv packs a frame's provenance buffer into a binary message.
func encodeProv(seq int, fc *debug.FrameCapture) []byte {
	w, h := fc.Width, fc.Height
	out := make([]byte, hdrSize+w*h*4)
	putHeader(out, kindProv, seq, w, h)
	for i, v := range fc.Prov {
		binary.LittleEndian.PutUint32(out[hdrSize+i*4:], uint32(v))
	}
	return out
}

func putHeader(b []byte, kind byte, seq, w, h int) {
	copy(b, "RDB1")
	b[4] = kind
	b[5] = 0
	binary.LittleEndian.PutUint16(b[6:], uint16(seq))
	binary.LittleEndian.PutUint32(b[8:], uint32(w))
	binary.LittleEndian.PutUint32(b[12:], uint32(h))
}

// toJSONCommands converts a capture's command stream for the command pane.
func toJSONCommands(cmds []debug.GPUCommand) []jsonCommand {
	out := make([]jsonCommand, len(cmds))
	for i, c := range cmds {
		words := make([]string, len(c.Words))
		for j, w := range c.Words {
			words[j] = fmt.Sprintf("%016x", w)
		}
		out[i] = jsonCommand{I: c.Index, Name: c.Name, Op: c.Op, Words: words, Decoded: c.Decoded}
	}
	return out
}
