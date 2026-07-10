package n64

// rdp.go is the Reality Display Processor's command stream: the 64-bit commands
// the RSP's microcode writes into a queue in RDRAM, and the state they build up.
// The rasteriser that consumes that state is in rdp_raster.go.
//
// The RDP is not a CPU. It has no program counter and no instruction fetch — it
// drains a span of memory between DPC_START and DPC_END, one variable-length
// command at a time, and each command either sets a piece of state or draws.
// That is why it lives here rather than in tools/cpu, exactly as the PlayStation's
// GPU lives in tools/platform/psx.
//
// Commands are dispatched on the top six bits of their first word. An unmodelled
// command, cycle type, texture format or combiner mode halts the machine rather
// than drawing something plausible: a rasteriser that quietly guesses produces
// images that look almost right, which is the hardest kind of wrong to notice.

import "encoding/binary"

// RDP command opcodes.
const (
	cmdNoOp            = 0x00
	cmdTriFill         = 0x08
	cmdTriFillZ        = 0x09
	cmdTriTex          = 0x0A
	cmdTriTexZ         = 0x0B
	cmdTriShade        = 0x0C
	cmdTriShadeZ       = 0x0D
	cmdTriShadeTex     = 0x0E
	cmdTriShadeTexZ    = 0x0F
	cmdTexRect         = 0x24
	cmdTexRectFlip     = 0x25
	cmdSyncLoad        = 0x26
	cmdSyncPipe        = 0x27
	cmdSyncTile        = 0x28
	cmdSyncFull        = 0x29
	cmdSetKeyGB        = 0x2A
	cmdSetKeyR         = 0x2B
	cmdSetConvert      = 0x2C
	cmdSetScissor      = 0x2D
	cmdSetPrimDepth    = 0x2E
	cmdSetOtherModes   = 0x2F
	cmdLoadTLUT        = 0x30
	cmdSetTileSize     = 0x32
	cmdLoadBlock       = 0x33
	cmdLoadTile        = 0x34
	cmdSetTile         = 0x35
	cmdFillRect        = 0x36
	cmdSetFillColor    = 0x37
	cmdSetFogColor     = 0x38
	cmdSetBlendColor   = 0x39
	cmdSetPrimColor    = 0x3A
	cmdSetEnvColor     = 0x3B
	cmdSetCombineMode  = 0x3C
	cmdSetTextureImage = 0x3D
	cmdSetMaskImage    = 0x3E
	cmdSetColorImage   = 0x3F
)

// cmdWords is the length of each command in 64-bit words. The triangle family
// grows with what it interpolates: four words of edge coefficients, plus eight
// for texture, eight for shade, and two for depth.
var cmdWords = [64]int{
	cmdTriFill: 4, cmdTriFillZ: 6,
	cmdTriTex: 12, cmdTriTexZ: 14,
	cmdTriShade: 12, cmdTriShadeZ: 14,
	cmdTriShadeTex: 20, cmdTriShadeTexZ: 22,
	cmdTexRect: 2, cmdTexRectFlip: 2,
}

// words returns a command's length, defaulting to one 64-bit word.
func cmdLen(op uint32) int {
	if n := cmdWords[op&0x3F]; n != 0 {
		return n
	}
	return 1
}

// cmdName labels a command for diagnostics.
var cmdName = map[uint32]string{
	cmdNoOp: "No_Op", cmdTriFill: "Triangle", cmdTriFillZ: "Triangle_Z",
	cmdTriTex: "Triangle_Tex", cmdTriTexZ: "Triangle_Tex_Z",
	cmdTriShade: "Triangle_Shade", cmdTriShadeZ: "Triangle_Shade_Z",
	cmdTriShadeTex: "Triangle_Shade_Tex", cmdTriShadeTexZ: "Triangle_Shade_Tex_Z",
	cmdTexRect: "Texture_Rectangle", cmdTexRectFlip: "Texture_Rectangle_Flip",
	cmdSyncLoad: "Sync_Load", cmdSyncPipe: "Sync_Pipe", cmdSyncTile: "Sync_Tile",
	cmdSyncFull: "Sync_Full", cmdSetKeyGB: "Set_Key_GB", cmdSetKeyR: "Set_Key_R",
	cmdSetConvert: "Set_Convert", cmdSetScissor: "Set_Scissor",
	cmdSetPrimDepth: "Set_Prim_Depth", cmdSetOtherModes: "Set_Other_Modes",
	cmdLoadTLUT: "Load_TLUT", cmdSetTileSize: "Set_Tile_Size",
	cmdLoadBlock: "Load_Block", cmdLoadTile: "Load_Tile", cmdSetTile: "Set_Tile",
	cmdFillRect: "Fill_Rectangle", cmdSetFillColor: "Set_Fill_Color",
	cmdSetFogColor: "Set_Fog_Color", cmdSetBlendColor: "Set_Blend_Color",
	cmdSetPrimColor: "Set_Prim_Color", cmdSetEnvColor: "Set_Env_Color",
	cmdSetCombineMode: "Set_Combine_Mode", cmdSetTextureImage: "Set_Texture_Image",
	cmdSetMaskImage: "Set_Mask_Image", cmdSetColorImage: "Set_Color_Image",
}

// Cycle types, from Set_Other_Modes bits 53..52. FILL and COPY bypass the colour
// combiner entirely, which is why a framebuffer clear needs almost none of the
// pipeline.
const (
	cycle1    = 0
	cycle2    = 1
	cycleCopy = 2
	cycleFill = 3
)

// Pixel formats, from the image and tile descriptors.
const (
	fmtRGBA = 0
	fmtYUV  = 1
	fmtCI   = 2
	fmtIA   = 3
	fmtI    = 4
)

// Pixel sizes, in bits: 4, 8, 16, 32.
const (
	size4  = 0
	size8  = 1
	size16 = 2
	size32 = 3
)

// rdpImage describes a framebuffer or a texture source in RDRAM.
type rdpImage struct {
	Format uint32
	Size   uint32
	Width  uint32 // in pixels, minus one as encoded; stored decoded
	Addr   uint32
}

// tile is one of the eight texture descriptors that address TMEM.
type tile struct {
	Format         uint32
	Size           uint32
	Line           uint32 // 64-bit words per row in TMEM
	TMem           uint32 // starting address in TMEM, in 64-bit words
	Palette        uint32
	CMS, CMT       uint32 // clamp/mirror/wrap
	MaskS, MaskT   uint32
	ShiftS, ShiftT uint32
	SL, TL, SH, TH uint32 // the tile's bounds, in 10.2 fixed point
}

// rdp holds the rasteriser's whole state. Fields are exported so encoding/gob
// carries them into a save-state.
type rdp struct {
	Color   rdpImage // where drawing lands
	Texture rdpImage // where a Load_Tile / Load_Block reads from
	Mask    uint32   // the depth buffer's address

	Scissor    struct{ XH, YH, XL, YL uint32 } // in 10.2 fixed point
	OtherModes uint64
	Combine    uint64

	FillColor  uint32
	FogColor   uint32
	BlendColor uint32
	PrimColor  uint32
	EnvColor   uint32
	PrimDepth  uint32

	Tiles [8]tile
	TMem  [4096]byte

	// Pending holds the words of a command whose tail has not been handed
	// over yet. The RDP consumes a continuous stream of 64-bit words, and the
	// fifo microcode will happily split one command across a buffer wrap: the
	// first word arrives as the last of one segment, DPC_START is rewritten,
	// and the rest continues at the buffer base. A drain that re-parsed at the
	// new CURRENT would read a command's middle as an opcode.
	Pending []uint64
}

// cycleType extracts the pipeline mode from Set_Other_Modes.
func (r *rdp) cycleType() uint32 { return uint32(r.OtherModes>>52) & 3 }

// RDPTMem exposes a copy of texture memory to instrumentation, so a tool can
// dump a tile exactly as the sampler will read it.
func (m *Machine) RDPTMem() []byte {
	t := make([]byte, len(m.rdp.TMem))
	copy(t, m.rdp.TMem[:])
	return t
}

// runRDP drains the command queue from DPC_CURRENT up to DPC_END.
//
// It resumes at CURRENT rather than START, and that is the whole contract.
// Microcode does not hand the RDP a finished list: it appends commands and walks
// DPC_END forward a word at a time, leaving START where it was. A rasteriser
// that restarted at START would re-execute the entire frame on every append.
//
// The queue is drained inside the write to DPC_END that advances it. Nothing a
// game can observe distinguishes that from a rasteriser running in parallel: the
// only ways to learn the RDP has finished are the Sync_Full interrupt and the
// DPC_CURRENT register, and both are reported as done.
func (m *Machine) runRDP() {
	start, end := m.dp[dpCurrent], m.dp[dpEnd]
	if start < m.dp[dpStart] {
		start = m.dp[dpStart]
	}
	if end <= start {
		return
	}

	// The command stream normally lives in RDRAM, but the RSP can point the RDP
	// at its own data memory instead.
	xbus := m.dp[dpStatus]&dpStatusXBusDMEM != 0

	read64 := func(addr uint32) uint64 {
		if xbus {
			a := addr & 0xFF8
			return binary.BigEndian.Uint64(m.DMEM[a:])
		}
		if int(addr)+8 > len(m.RDRAM) {
			return 0 // RDRAM does not mirror; past it reads zero
		}
		return binary.BigEndian.Uint64(m.RDRAM[addr:])
	}

	// Words are consumed one at a time into the pending command, which runs
	// when complete. CURRENT advances per word, so a command split across a
	// segment boundary picks its tail up wherever the next segment begins.
	r := &m.rdp
	for addr := start; addr < end; {
		r.Pending = append(r.Pending, read64(addr))
		addr += 8
		m.dp[dpCurrent] = addr
		m.rdpWords++

		op := uint32(r.Pending[0] >> 56 & 0x3F)
		if len(r.Pending) < cmdLen(op) {
			continue
		}
		words := r.Pending
		r.Pending = nil
		if m.OnRDPCmd != nil {
			m.OnRDPCmd(m, op, words)
		}
		m.execRDP(op, words)
		if m.CPU.Halted {
			return
		}
	}
}

func cmdNameOf(op uint32) string {
	if n, ok := cmdName[op]; ok {
		return n
	}
	return "unknown"
}

// execRDP applies one command.
func (m *Machine) execRDP(op uint32, w []uint64) {
	r := &m.rdp
	switch op {
	case cmdNoOp:

	// --- state ---------------------------------------------------------------
	case cmdSetColorImage:
		r.Color = decodeImage(w[0])
	case cmdSetTextureImage:
		r.Texture = decodeImage(w[0])
	case cmdSetMaskImage:
		r.Mask = uint32(w[0]) & 0x03FFFFFF
	case cmdSetScissor:
		r.Scissor.XH = uint32(w[0] >> 44 & 0xFFF)
		r.Scissor.YH = uint32(w[0] >> 32 & 0xFFF)
		r.Scissor.XL = uint32(w[0] >> 12 & 0xFFF)
		r.Scissor.YL = uint32(w[0] & 0xFFF)
	case cmdSetOtherModes:
		r.OtherModes = w[0] & 0x00FFFFFFFFFFFFFF
	case cmdSetCombineMode:
		r.Combine = w[0] & 0x00FFFFFFFFFFFFFF
	case cmdSetFillColor:
		r.FillColor = uint32(w[0])
	case cmdSetFogColor:
		r.FogColor = uint32(w[0])
	case cmdSetBlendColor:
		r.BlendColor = uint32(w[0])
	case cmdSetPrimColor:
		r.PrimColor = uint32(w[0])
	case cmdSetEnvColor:
		r.EnvColor = uint32(w[0])
	case cmdSetPrimDepth:
		r.PrimDepth = uint32(w[0] >> 16 & 0xFFFF)
	case cmdSetConvert, cmdSetKeyGB, cmdSetKeyR:
		// Colour-space conversion and chroma keying: recorded, not yet applied.
		// Nothing seen so far draws with them enabled.

	case cmdSetTile:
		r.Tiles[w[0]>>24&7] = decodeTile(w[0])
	case cmdSetTileSize:
		t := &r.Tiles[w[0]>>24&7]
		t.SL = uint32(w[0] >> 44 & 0xFFF)
		t.TL = uint32(w[0] >> 32 & 0xFFF)
		t.SH = uint32(w[0] >> 12 & 0xFFF)
		t.TH = uint32(w[0] & 0xFFF)

	case cmdLoadBlock:
		m.loadBlock(w[0])
	case cmdLoadTile:
		m.loadTile(w[0])
	case cmdLoadTLUT:
		m.loadTLUT(w[0])

	// --- synchronisation -----------------------------------------------------
	case cmdSyncLoad, cmdSyncPipe, cmdSyncTile:
		// The pipeline is not modelled, so there is nothing to wait for.
	case cmdSyncFull:
		// The game's signal that the frame is finished.
		m.raiseIRQ(intrDP)

	// --- drawing -------------------------------------------------------------
	case cmdFillRect:
		m.fillRect(w[0])
	case cmdTexRect, cmdTexRectFlip:
		m.texRect(w, op == cmdTexRectFlip)
	case cmdTriFill, cmdTriFillZ, cmdTriTex, cmdTriTexZ,
		cmdTriShade, cmdTriShadeZ, cmdTriShadeTex, cmdTriShadeTexZ:
		m.triangle(op, w)

	default:
		m.CPU.Halt("unmodelled RDP command 0x%02X (%s) at DPC 0x%08X", op, cmdNameOf(op), m.dp[dpCurrent])
	}
}

func decodeImage(w uint64) rdpImage {
	return rdpImage{
		Format: uint32(w >> 53 & 7),
		Size:   uint32(w >> 51 & 3),
		Width:  uint32(w>>32&0x3FF) + 1,
		Addr:   uint32(w) & 0x03FFFFFF,
	}
}

func decodeTile(w uint64) tile {
	return tile{
		Format:  uint32(w >> 53 & 7),
		Size:    uint32(w >> 51 & 3),
		Line:    uint32(w >> 41 & 0x1FF),
		TMem:    uint32(w >> 32 & 0x1FF),
		Palette: uint32(w >> 20 & 15),
		CMT:     uint32(w >> 18 & 3),
		MaskT:   uint32(w >> 14 & 15),
		ShiftT:  uint32(w >> 10 & 15),
		CMS:     uint32(w >> 8 & 3),
		MaskS:   uint32(w >> 4 & 15),
		ShiftS:  uint32(w & 15),
	}
}
