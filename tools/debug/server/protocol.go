package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"

	"retroreverse.com/tools/debug"
)

// The wire protocol.
//
// Client to server is a JSON text message with a namespaced op and a per-op payload:
//
//	{"op":"frame.scrub","seq":9,"args":{"k":412}}
//
// A flat struct with a field per op does not survive twenty ops, so args is left raw
// and decoded by whoever handles the op.
//
// Server to client is JSON text for structure and a binary message for pixels — a
// framebuffer is incompressible, and base64 in JSON would cost a third more bytes and
// a decode pass for nothing. A reply carries the seq of the request it answers; an
// *event* (a watch hit, a breakpoint stop, a free-running frame) carries no seq and
// goes to every attached session.
//
// A binary message is a 24-byte header followed by the payload:
//
//	off  size  field
//	 0    4    magic "RDB2"
//	 4    1    kind   (kindImage | kindProv)
//	 5    1    flags  (reserved, 0)
//	 6    2    seq    — the request this answers; 0 for an unsolicited event
//	 8    2    stream — which consumer it is for, so a viewport and a VRAM panel can
//	                    both have images in flight without confusing each other
//	10    2    (pad)
//	12    4    width
//	16    4    height
//	20    4    (reserved)
//	24    ...  payload
//
// All little-endian. The header is 24 bytes so the payload lands 4-byte aligned and
// the browser can wrap the received ArrayBuffer in a Uint8ClampedArray (for
// ImageData) or an Int32Array (for provenance) with no copy.
const (
	hdrSize = 24

	kindImage byte = 1 // payload: width*height*4 bytes, RGBA8888
	kindProv  byte = 2 // payload: width*height int32, the command that last wrote each pixel (-1 = none)
)

// Streams. A binary message says which panel asked for it.
const (
	streamMain    uint16 = 0 // the frame viewport
	streamSurface uint16 = 1 // the memory/VRAM surface panel
)

// req is a client request. Args is decoded per op.
type req struct {
	Op   string          `json:"op"`
	Seq  int             `json:"seq"`
	Args json.RawMessage `json:"args"`
}

// The per-op argument payloads.
type (
	stepArgs struct {
		N        int  `json:"n"`        // fields to advance; 0 = step to a drawn frame
		Overdraw bool `json:"overdraw"` // record the full per-pixel write history
	}
	playArgs struct {
		On bool `json:"on"`
		// Overdraw applies to the capture that pausing does, so stopping lands on a
		// frame recorded with the detail the page asked for.
		Overdraw bool `json:"overdraw"`
	}
	scrubArgs struct {
		K int `json:"k"`
		// Blank starts the replay from a black screen, so what this frame draws is not
		// mixed with what the last one left on the panel.
		Blank bool `json:"blank"`
	}
	pixelArgs struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	memArgs struct {
		Addr uint32 `json:"addr"`
		Len  int    `json:"len"`
	}
	disasmArgs struct {
		Addr uint64 `json:"addr"`
		N    int    `json:"n"`
	}
	cpuRunArgs struct {
		N int `json:"n"` // instructions to step
	}
	bpArgs struct {
		PC uint64 `json:"pc"`
	}
	watchArgs struct {
		Kind  string `json:"kind"` // "read" | "write"
		Lo    uint32 `json:"lo"`
		Hi    uint32 `json:"hi"`
		Break bool   `json:"break"`
	}
	unwatchArgs struct {
		ID int `json:"id"`
	}
	stateArgs struct {
		Path string `json:"path"`
	}
	fsArgs struct {
		Path string `json:"path"`
	}
	touchArgs struct {
		Panel string `json:"panel"`
		X     int    `json:"x"` // panel-local pixels
		Y     int    `json:"y"`
		Down  bool   `json:"down"`
	}
	keyArgs struct {
		Name string `json:"name"` // key identity ("up", "enter", "esc", "a")
		Code int    `json:"code"` // raw browser key code, when a name is not enough
		Down bool   `json:"down"` // press (true) or release (false)
	}
	surfaceArgs struct {
		ID      string `json:"id"`
		Addr    uint32 `json:"addr"`
		W       int    `json:"w"`
		H       int    `json:"h"`
		Stride  int    `json:"stride"`
		Format  string `json:"format"`
		Palette uint32 `json:"palette"`
	}
)

// decode unpacks an op's args, giving a legible error rather than a zero value.
func decodeArgs(r req, v any) error {
	if len(r.Args) == 0 {
		return nil // an op with no args is fine; the zero value is the default
	}
	if err := json.Unmarshal(r.Args, v); err != nil {
		return fmt.Errorf("bad args for %s: %w", r.Op, err)
	}
	return nil
}

// coalescable reports whether a queued request may be replaced by a newer one of the
// same op. A scrubber drag fires an op per mouse-move and only the last one matters;
// the same is true of aiming a surface or scrolling a disassembly. A step, a state
// save or a watch must never be dropped.
//
// Nor must a stylus event. Dragging the pen also fires an op per mouse-move, but the
// release that ends the drag is the same op as the moves before it, and dropping the
// wrong one would leave the pen down for ever. The runner collapses a drag itself, where
// it can tell a move from a lift — see Runner.applyTouch.
func coalescable(op string) bool {
	switch op {
	case "frame.scrub", "surface.render", "cpu.disasm", "mem.read":
		return true
	}
	return false
}

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

// The server-to-client messages. Each carries Type, and a reply also carries the Seq
// of the request it answers.
type (
	helloMsg struct {
		Type     string   `json:"type"` // "hello"
		Platform string   `json:"platform"`
		Title    string   `json:"title"`
		Caps     []string `json:"caps"` // what this target can do; the page builds itself from this
	}

	frameMsg struct {
		Type     string        `json:"type"` // "frame"
		Seq      int           `json:"seq"`
		Frame    int           `json:"frame"`
		Commands []jsonCommand `json:"commands"`
		W        int           `json:"w"`
		H        int           `json:"h"`
		HasProv  bool          `json:"hasProv"`
		Overdraw bool          `json:"overdraw"`
		StepMs   float64       `json:"stepMs"`

		// Writers are the commands that actually wrote pixels into memory, in execution
		// order — what the scrubber walks, because the rest of the stream is setup and
		// scrubbing through it is scrubbing through a picture that does not change.
		//
		// A list of indices rather than a flag on each command, and for a reason: a 3DS
		// frame is two hundred thousand commands and a few hundred writers, so a flag per
		// command would put a megabyte of "false" on the wire to say what this says in a
		// few kilobytes.
		//
		// null means the platform does not report pixels and the page must fall back to the
		// whole stream; [] means it does report them and this frame wrote nothing.
		Writers []int `json:"writers"`

		// Profile is where the machine says the frame's time went, when the target
		// can say (debug.Profiler). It rides the frame message rather than being a
		// request of its own: it is a property of the frame just stepped, and asking
		// for it separately would invite reading it against a different frame.
		Profile *jsonProfile `json:"profile,omitempty"`
	}

	jsonProfile struct {
		TotalMs  float64           `json:"totalMs"`
		Buckets  []jsonProfBucket  `json:"buckets"`
		Counters []jsonProfCounter `json:"counters"`
	}

	jsonProfBucket struct {
		Name   string  `json:"name"`
		Millis float64 `json:"ms"`
		Count  int     `json:"count"`
	}

	jsonProfCounter struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	renderMsg struct {
		Type     string  `json:"type"` // "render" — announces the binary image that follows
		Seq      int     `json:"seq"`
		Stream   uint16  `json:"stream"`
		K        int     `json:"k"` // -1 when the image is a scanout, not a command's draw target
		RenderMs float64 `json:"renderMs"`
		Bytes    int     `json:"bytes"`
		Cached   bool    `json:"cached"`
		Play     bool    `json:"play,omitempty"`  // a free-running frame: unrequested, never stale
		Frame    int     `json:"frame,omitempty"` // fields stepped so far

		// Profile rides a played frame too, so the profile panel stays live while the
		// machine free-runs — which is when a performance question is usually asked.
		Profile *jsonProfile `json:"profile,omitempty"`
	}

	cpuMsg struct {
		Type  string            `json:"type"` // "cpu"
		Seq   int               `json:"seq"`
		PC    string            `json:"pc"`
		Names []string          `json:"names"`
		Vals  []string          `json:"vals"`
		Extra map[string]string `json:"extra"`
	}

	disasmMsg struct {
		Type  string      `json:"type"` // "disasm"
		Seq   int         `json:"seq"`
		PC    string      `json:"pc"`
		Instr []jsonInstr `json:"instr"`
		BPs   []string    `json:"bps"`
	}

	jsonInstr struct {
		Addr string `json:"addr"`
		Text string `json:"text"`
	}

	memMsg struct {
		Type    string       `json:"type"` // "mem"
		Seq     int          `json:"seq"`
		Addr    uint32       `json:"addr"`
		Data    string       `json:"data"` // hex
		Regions []jsonRegion `json:"regions,omitempty"`
	}

	jsonRegion struct {
		Name string `json:"name"`
		Lo   string `json:"lo"`
		Hi   string `json:"hi"`
	}

	pixelMsg struct {
		Type   string      `json:"type"` // "pixel"
		Seq    int         `json:"seq"`
		X      int         `json:"x"`
		Y      int         `json:"y"`
		Cmd    int         `json:"cmd"` // last writer, -1 if none
		Writes []jsonWrite `json:"writes"`
	}

	watchesMsg struct {
		Type    string      `json:"type"` // "watches"
		Seq     int         `json:"seq"`
		Watches []jsonWatch `json:"watches"`
	}

	jsonWatch struct {
		ID    int    `json:"id"`
		Kind  string `json:"kind"`
		Lo    string `json:"lo"`
		Hi    string `json:"hi"`
		Break bool   `json:"break"`
		Hits  uint64 `json:"hits"`
	}

	// hitMsg is an event: a watch fired. It carries no seq — nobody asked for it.
	hitMsg struct {
		Type  string `json:"type"` // "hit"
		ID    int    `json:"id"`
		Kind  string `json:"kind"`
		Addr  string `json:"addr"`
		Val   string `json:"val"`
		PC    string `json:"pc"`
		Instr string `json:"instr"`
	}

	// stoppedMsg is an event: a CPU run ended, and why.
	stoppedMsg struct {
		Type   string `json:"type"` // "stopped"
		Reason string `json:"reason"`
		PC     string `json:"pc"`
		Steps  uint64 `json:"steps"`
		Note   string `json:"note"`
	}

	filesMsg struct {
		Type    string     `json:"type"` // "files"
		Seq     int        `json:"seq"`
		Path    string     `json:"path"`
		Entries []jsonFile `json:"entries"`
	}

	jsonFile struct {
		Name   string `json:"name"`
		Path   string `json:"path"`
		Dir    bool   `json:"dir"`
		Size   int64  `json:"size"`
		Offset int64  `json:"offset"` // where it lives in the image (a disc sector, say)
	}

	fileMsg struct {
		Type string `json:"type"` // "file" — the head of a file, as hex
		Seq  int    `json:"seq"`
		Path string `json:"path"`
		Size int64  `json:"size"` // the file's real size, not the size of this excerpt
		Data string `json:"data"` // hex, capped
	}

	surfacesMsg struct {
		Type     string        `json:"type"` // "surfaces"
		Seq      int           `json:"seq"`
		Surfaces []jsonSurface `json:"surfaces"`
	}

	jsonSurface struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Free    bool     `json:"free"` // the caller aims it: address, size, format
		W       int      `json:"w"`
		H       int      `json:"h"`
		Formats []string `json:"formats"`
	}

	// touchPanelsMsg says where the stylus can reach, in the coordinates of the frame
	// the page is already drawing. It is a reply rather than part of hello because the
	// 3DS's layout is not fixed: the bottom screen exists once the game has presented
	// it, and not before.
	touchPanelsMsg struct {
		Type   string           `json:"type"` // "touchpanels"
		Seq    int              `json:"seq"`
		Panels []jsonTouchPanel `json:"panels"`
	}

	jsonTouchPanel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
		W    int    `json:"w"`
		H    int    `json:"h"`
	}

	libraryMsg struct {
		Type    string     `json:"type"` // "library"
		Seq     int        `json:"seq"`
		Current string     `json:"current"` // the open game's slug
		Games   []jsonGame `json:"games"`
	}

	jsonGame struct {
		Slug     string `json:"slug"`
		Platform string `json:"platform"`
		Name     string `json:"name"`
		Image    string `json:"image"`
		Missing  bool   `json:"missing"` // the image is not on disk; they are gitignored
	}

	statesMsg struct {
		Type   string      `json:"type"` // "states"
		Seq    int         `json:"seq"`
		Dir    string      `json:"dir"`
		States []jsonState `json:"states"`
		Note   string      `json:"note,omitempty"` // why there are none, when there cannot be
	}

	jsonState struct {
		Name   string `json:"name"`
		File   string `json:"file"`
		Size   int64  `json:"size"`
		When   string `json:"when"`
		Resume string `json:"resume"` // the command line that reopens the game here
	}

	okMsg struct {
		Type string `json:"type"` // "ok" — an op that has nothing to report
		Seq  int    `json:"seq"`
		Op   string `json:"op"`
		Note string `json:"note,omitempty"`
	}

	errMsg struct {
		Type string `json:"type"` // "error"
		Seq  int    `json:"seq"`
		Msg  string `json:"msg"`
	}
)

// encodeImage packs an RGBA image into a binary message.
func encodeImage(seq int, stream uint16, img *image.RGBA) []byte {
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	out := make([]byte, hdrSize+w*h*4)
	putHeader(out, kindImage, seq, stream, w, h)
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
	putHeader(out, kindProv, seq, streamMain, w, h)
	for i, v := range fc.Prov {
		binary.LittleEndian.PutUint32(out[hdrSize+i*4:], uint32(v))
	}
	return out
}

func putHeader(b []byte, kind byte, seq int, stream uint16, w, h int) {
	copy(b, "RDB2")
	b[4] = kind
	b[5] = 0
	binary.LittleEndian.PutUint16(b[6:], uint16(seq))
	binary.LittleEndian.PutUint16(b[8:], stream)
	binary.LittleEndian.PutUint32(b[12:], uint32(w))
	binary.LittleEndian.PutUint32(b[16:], uint32(h))
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
