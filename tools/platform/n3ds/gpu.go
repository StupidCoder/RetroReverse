package n3ds

// gpu.go is the PICA200 command processor: it executes the command lists the
// game submits through GX ProcessCommandList. A command list is a register-
// write stream (pica.go decodes it); most writes just latch state into the
// register file, and a handful of registers have side effects — the shader
// code/operand/uniform upload FIFOs, and the two draw triggers that run the
// whole pipeline (vertex fetch → LLE vertex shader → primitive assembly →
// rasterisation, in gpu_raster.go).
//
// This is the analogue of the N64 RDP command interpreter (n64/rdp.go): apply
// each write, treat specific registers as triggers, and stop loudly at
// anything structurally unknown rather than guessing. Register ids follow the
// PICA200's register file layout (the platform's, not any game's); their
// semantics here are the ones the captured Super Mario 3D Land lists exercise,
// implemented one by one as the game demands them.

import "fmt"

// PICA register ids the interpreter gives side effects to. Pure-state
// registers (culling, blending, TEV stages, texture config…) are read straight
// from the register file by the pipeline stages that consume them.
const (
	regViewportWidth  = 0x041 // float24: half the viewport width
	regViewportHeight = 0x043 // float24: half the viewport height
	regViewportXY     = 0x068 // viewport origin within the render target
	regDepthMapScale  = 0x04D // float24
	regDepthMapOffset = 0x04E // float24
	regDepthMapMode   = 0x06D // bit0: 1 = Z-buffering, 0 = W-buffering
	regShOutmapTotal  = 0x04F // number of vertex-shader output registers
	regShOutmapO0     = 0x050 // 0x050-0x056: per-output semantic map

	regTexUnitConfig  = 0x080 // bits 0-2: texture unit enables
	regTex0Param      = 0x083 // texture unit 0: wrap/filter + the texture TYPE
	regShadowTex      = 0x08B // shadow texture: bit0 orthographic, bits 1-23 bias
	regLightingEnable = 0x08F // bit0: fragment lighting on

	// The output merger's mode lives in the low bits of the blend register: 0 is
	// the ordinary colour pipeline, 3 replaces it with the shadow-map writer.
	regBlendConfig    = 0x100
	regShadowDensity  = 0x130 // two float16s: the shadow density curve's constant, linear
	regDepthColorMask = 0x107 // depth test/func + colour/depth write masks

	// The buffer-access enables, a second gate in front of 0x107: they permit
	// the memory traffic itself. A render pass that keeps the depth *test* bits
	// set in 0x107 but zeroes these does no depth read or write at all — Captain
	// Toad's shadow pass does exactly that, and honouring only 0x107 made us
	// write a megabyte of depth over VRAM the game had put command buffers in.
	regColorbufRead  = 0x112
	regColorbufWrite = 0x113
	regDepthbufRead  = 0x114
	regDepthbufWrite = 0x115

	regDepthbufFormat = 0x116
	regColorbufFormat = 0x117
	regDepthbufLoc    = 0x11C // physical>>3
	regColorbufLoc    = 0x11D // physical>>3
	regFramebufDim    = 0x11E

	regAttrBase    = 0x200 // physical>>3 of the attribute-buffer arena
	regAttrFmtLow  = 0x201 // 12 × 4-bit component formats
	regAttrFmtHigh = 0x202 // + fixed-attr mask + attribute count
	regIndexConfig = 0x227 // index-buffer offset + 8/16-bit flag
	regNumVertices = 0x228
	regGeoConfig   = 0x229 // geometry-shader stage configuration
	regVertexOff   = 0x22A // first vertex for DrawArrays
	regDrawArrays  = 0x22E
	regDrawElems   = 0x22F
	regFixedIndex  = 0x232 // fixed-attribute index select
	regFixedData0  = 0x233 // 0x233-0x235: 3 words = one packed float24 vec4

	// The command processor's own chain registers: two preloaded buffer slots
	// (address>>3, size-in-bytes>>3) and a trigger apiece. Writing a JUMP makes
	// the processor continue in that buffer — it is a jump, not a call.
	regCmdBufSize0 = 0x238
	regCmdBufAddr0 = 0x23A
	regCmdBufJump0 = 0x23C

	regPrimConfig = 0x25E // bits 8-9: primitive type

	regVshBool     = 0x2B0
	regVshInt0     = 0x2B1
	regVshMaxInput = 0x2B9 // highest attribute index delivered to the shader
	regVshEntry    = 0x2BA
	// The attribute → input-register map: nibble a names the register attribute
	// a is delivered to (not the attribute register a takes — the inverse).
	regVshAttrPermL = 0x2BB
	regVshAttrPermH = 0x2BC
	regVshFloatCfg  = 0x2C0 // uniform upload: start index + f32 flag
	regVshFloatData = 0x2C1 // 0x2C1-0x2C8: uniform data FIFO
	regVshCodeIdx   = 0x2CB
	regVshCodeData  = 0x2CC // 0x2CC-0x2D3: code upload FIFO
	regVshOpdescIdx = 0x2D5
	regVshOpdescDat = 0x2D6 // 0x2D6-0x2DD: operand-descriptor upload FIFO
)

// GPU is the PICA200 model: the register file plus the shader engine's
// uploaded state, owned by the Machine.
type GPU struct {
	m *Machine

	Regs [0x300]uint32

	// Vertex-shader engine memory (uploaded through the code/opdesc FIFOs).
	Code    [4096]uint32
	Opdesc  [128]uint32
	Float   [96][4]float32 // c0-c95
	Bool    uint32         // b0-b15 (bit i)
	Int     [4][4]uint8    // i0-i3
	codeIdx int
	opdIdx  int

	// Float-uniform upload state (register 0x2C0 + the 0x2C1-0x2C8 FIFO).
	fltIdx int      // current destination uniform
	fltF32 bool     // 32-bit float mode (vs packed float24)
	fltBuf []uint32 // words accumulated toward one vec4

	// Fixed-attribute upload state (0x232 + the 0x233-0x235 words).
	fixedIdx int
	fixedBuf []uint32
	fixedVal [16][4]float32 // latched fixed/default attribute values

	// The fragment-lighting lookup tables (gpu_light.go): 24 tables of 256
	// entries, each entry a value and the slope to the next, uploaded through
	// the 0x1C5 index/type register and the 0x1C8-0x1CF data FIFO.
	LUT     [numLUT][256]float32
	LUTDiff [numLUT][256]float32
	lutSet  [numLUT]bool // uploaded at least once (an instrument, not state)
	lutIdx  uint32
	lutType uint32

	// The geometry-shader port mirrors the vertex-shader engine's upload
	// registers at 0x280-0x2AF. The game only uses it to clear state (its
	// draws keep the geometry stage off — checked at draw time); the uploads
	// are latched so nothing is lost, but the unit is not executed.
	gshCodeIdx int
	gshOpdIdx  int
	gshFltIdx  int
	gshFltF32  bool
	gshFltBuf  []uint32

	Draws        int // draw triggers executed (both kinds)
	RejectedTris int // triangles dropped by the w<=0 near-plane shortcut
	ZeroAreaTris int
	CulledTris   int
	DepthKilled  int // fragments failing the depth test
	PixelsDrawn  int // fragments written to the colour buffer

	// The shadow path (gpu_raster.go's shadowMapWrite, gpu_texture.go's
	// shadowCompare): fragments the shadow pass contributed to a shadow map, and
	// how many of the fragments that later sampled one came back occluded.
	ShadowWrites   int
	ShadowSamples  int
	ShadowOccluded int

	// TraceDraws > 0 prints a per-draw summary (vertex fetch, first clip
	// positions, screen coords) for that many draws — the draw-path
	// instrument.
	TraceDraws int

	// texCache holds decoded textures keyed by address/format/size. GX writes
	// into texture memory (DMA, TextureCopy) invalidate it.
	texCache map[texKey]*texImage

	// Set by a write to a CMDBUF_JUMP register: the buffer the command
	// processor continues in when the current one ends.
	jumpAddr, jumpSize uint32
	jumpPending        bool

	ListHops int // command buffers entered by a jump (instrument)

	// The decoded-instruction cache for the vertex shader (gpu_shader_cache.go).
	// dec[pc] is valid only while decEpoch[pc] == shEpoch; every code or operand-
	// descriptor upload — and every savestate restore — bumps shEpoch. Not machine
	// state: it is derived from Code/Opdesc and rebuilt on demand, so it is
	// deliberately outside the savestate.
	dec      [4096]shInst
	decEpoch [4096]uint32
	shEpoch  uint32
	// decodedAll names the epoch the whole code memory has been decoded for. Once it
	// matches, the cache is pure data and the vertex loop can read it from several
	// goroutines at once (gpu_shader_cache.go's decodeAll).
	decodedAll uint32

	// Scratch buffers, reused across command lists and draws. Not machine state:
	// their contents are rebuilt from scratch every time they are used, and they
	// are deliberately outside the savestate. They exist because Captain Toad runs
	// ~2,500 command lists and 143 draws per frame, and allocating each one's
	// working memory afresh put the Go allocator — not the PICA — at a tenth of the
	// frame's cost.
	cmdBuf   []byte      // the command list, copied out of guest memory
	cmdWrite []PICAWrite // its decoded register-write stream
	outs     []vsOut     // one draw's shaded vertices
	tris     []rasterTri // one draw's assembled, culled triangles

	// The worker goroutines the parallel vertex and raster stages run on
	// (workpool.go). Started on first use, and outside the savestate: how a stage
	// executes is not a property of the machine's state.
	workers *workPool
	bufs    []loaderBuf // one draw's attribute-loader configuration
	comps   []int       // the loaders' component lists, in one backing array
}

// maxCmdBufHops bounds a chain so a corrupt jump register (or a buffer that
// jumps to itself) halts loudly instead of hanging the oracle. Captain Toad's
// real chains are a handful of hops.
const maxCmdBufHops = 4096

// newGPU builds the PICA200 model. shEpoch starts at 1, never 0: a zero epoch
// would compare equal to the zero value of decEpoch, and every cache slot would
// pass as a validly-decoded instruction it had never decoded.
func newGPU(m *Machine) *GPU { return &GPU{m: m, shEpoch: 1} }

// GPU exposes the PICA200 model (counters, trace switches).
func (m *Machine) GPU() *GPU { return m.gpu }

// Execute runs one PICA200 command list (virtual address + byte size, as the
// game's GX ProcessCommandList command carries them) — and every buffer it
// chains into. A list ends either by running out of bytes or by writing a
// CMDBUF_JUMP register, which makes the command processor continue in a
// preloaded buffer; the submitted list is only the head of the chain.
func (g *GPU) Execute(addr, size uint32) {
	g.jumpPending = false
	for hop := 0; ; hop++ {
		if hop > maxCmdBufHops {
			g.m.CPU.Halt("gpu: command-buffer chain exceeded %d hops (at 0x%08X) after %d instructions",
				maxCmdBufHops, addr, g.m.CPU.Instrs)
			return
		}
		t := g.m.profStart() // one clock read per command list (profile.go)
		if uint32(cap(g.cmdBuf)) < size {
			g.cmdBuf = make([]byte, size)
		}
		buf := g.cmdBuf[:size]
		// A frame's command lists are a quarter of a megabyte, and the command
		// processor used to fetch them a byte at a time through the bus. When the list
		// sits inside one region — it always does — copy it.
		if src, off := g.m.directRange(addr, size); src != nil {
			copy(buf, src[off:off+size])
		} else {
			for i := uint32(0); i < size; i++ {
				buf[i] = g.m.Read(addr + i)
			}
		}
		if hop > 0 {
			g.ListHops++
			if g.m.GXCapture {
				// The capture keeps the buffer; hand it a copy, since the scratch one
				// is overwritten by the next list.
				g.m.captureChainedList(addr, size, append([]byte(nil), buf...))
			}
		}
		ws, err := DecodePICAInto(g.cmdWrite[:0], buf)
		g.cmdWrite = ws // keep the grown backing array for the next list
		g.m.profEnd(bucketCmd, t)
		if err != nil {
			g.m.CPU.Halt("gpu: %v (list at 0x%08X)", err, addr)
			return
		}
		for _, w := range ws {
			// The debugger sees each write before it executes, so the pixels a draw
			// trigger produces are attributed to the write that triggered them; and
			// picaLimit, when set, stops the list mid-stream, which is what lets the
			// command scrubber render the frame as it stood after write k.
			m := g.m
			if m.picaLimit > 0 {
				if m.picaCount >= m.picaLimit {
					m.StopRequested = true
					return
				}
				m.picaCount++
			}
			if m.OnPICACmd != nil {
				m.OnPICACmd(w)
			}
			g.write(w)
			if m.CPU.Halted {
				return
			}
			if g.jumpPending {
				break // the rest of this buffer is not executed
			}
		}
		if !g.jumpPending {
			return
		}
		addr, size = g.jumpAddr, g.jumpSize
		g.jumpPending = false
	}
}

// write applies one register write (with its byte-enable mask) and dispatches
// the side effect if the register has one.
func (g *GPU) write(w PICAWrite) {
	if w.Mask == 0 || int(w.Reg) >= len(g.Regs) {
		return // masked-out (alignment padding) or out of the register file
	}
	old := g.Regs[w.Reg]
	v := old
	for b := 0; b < 4; b++ {
		if w.Mask&(1<<b) != 0 {
			sh := uint(8 * b)
			v = v&^(0xFF<<sh) | w.Value&(0xFF<<sh)
		}
	}
	g.Regs[w.Reg] = v

	switch {
	case w.Reg == regDrawArrays && v != 0:
		g.drawArrays()
	case w.Reg == regDrawElems && v != 0:
		g.drawElements()

	// --- command-buffer chain: continue in the buffer this slot preloaded ---
	case w.Reg == regCmdBufJump0 || w.Reg == regCmdBufJump0+1:
		i := w.Reg - regCmdBufJump0
		g.jumpAddr = g.m.gpuAddrToVirt(g.Regs[regCmdBufAddr0+i] << 3)
		g.jumpSize = g.Regs[regCmdBufSize0+i] << 3
		g.jumpPending = g.jumpSize != 0

	// --- fragment-lighting lookup-table upload ---
	case w.Reg == regLightLUTIndex:
		g.lutIdx = v & 0xFF
		g.lutType = v >> 8 & 0x1F
	case w.Reg >= regLightLUTData && w.Reg < regLightLUTData+8:
		g.lutWrite(v)

	// --- vertex-shader engine uploads ---
	case w.Reg == regVshFloatCfg:
		g.fltIdx = int(v & 0xFF)
		g.fltF32 = v>>31 != 0
		g.fltBuf = g.fltBuf[:0]
	case w.Reg >= regVshFloatData && w.Reg < regVshFloatData+8:
		g.floatUniformWord(v)
	case w.Reg == regVshCodeIdx:
		g.codeIdx = int(v & 0xFFF)
	case w.Reg >= regVshCodeData && w.Reg < regVshCodeData+8:
		if g.codeIdx < len(g.Code) {
			g.Code[g.codeIdx] = v
			g.codeIdx++
			g.invalidateShaders() // the decode cache is derived from this word
		}
	case w.Reg == regVshOpdescIdx:
		g.opdIdx = int(v & 0x7F)
	case w.Reg >= regVshOpdescDat && w.Reg < regVshOpdescDat+8:
		if g.opdIdx < len(g.Opdesc) {
			g.Opdesc[g.opdIdx] = v
			g.opdIdx++
			// A descriptor carries the swizzle, negate and write mask of every
			// instruction that names it — so this invalidates decodes anywhere.
			g.invalidateShaders()
		}
	case w.Reg == regVshBool:
		g.Bool = v & 0xFFFF
	case w.Reg >= regVshInt0 && w.Reg < regVshInt0+4:
		i := w.Reg - regVshInt0
		g.Int[i] = [4]uint8{uint8(v), uint8(v >> 8), uint8(v >> 16), uint8(v >> 24)}

	// --- fixed attributes ---
	case w.Reg == regFixedIndex:
		g.fixedIdx = int(v & 0xF)
		g.fixedBuf = g.fixedBuf[:0]
	case w.Reg >= regFixedData0 && w.Reg < regFixedData0+3:
		g.fixedBuf = append(g.fixedBuf, v)
		if len(g.fixedBuf) == 3 {
			g.fixedVal[g.fixedIdx] = unpackF24x4(g.fixedBuf[0], g.fixedBuf[1], g.fixedBuf[2])
			g.fixedBuf = g.fixedBuf[:0]
		}

	// --- geometry-shader port: latch, never execute (checked at draw) ---
	case w.Reg == 0x290:
		g.gshFltIdx = int(v & 0xFF)
		g.gshFltF32 = v>>31 != 0
		g.gshFltBuf = g.gshFltBuf[:0]
	case w.Reg == 0x29B:
		g.gshCodeIdx = int(v & 0xFFF)
	case w.Reg >= 0x29C && w.Reg < 0x29C+8:
		// The init list clears code memory through this port too (index 0x200
		// continuing past the vertex-shader port's 512 words).
		if g.gshCodeIdx < len(g.Code) {
			g.Code[g.gshCodeIdx] = v
			g.gshCodeIdx++
			g.invalidateShaders() // the same Code array the vertex shader executes
		}
	}
}

// floatUniformWord accumulates one word of the float-uniform FIFO; every 4
// words (f32 mode) or 3 words (packed f24) completes a vec4 written to c[i].
// Both modes upload w first, x last — the same order the shaders' pervasive
// .wzyx matrix-row swizzles assume, and the same packing the fixed-attribute
// words use.
func (g *GPU) floatUniformWord(v uint32) {
	g.fltBuf = append(g.fltBuf, v)
	if g.fltF32 {
		if len(g.fltBuf) == 4 {
			if g.fltIdx < len(g.Float) {
				c := &g.Float[g.fltIdx]
				c[0] = toF24(f32bits(g.fltBuf[3]))
				c[1] = toF24(f32bits(g.fltBuf[2]))
				c[2] = toF24(f32bits(g.fltBuf[1]))
				c[3] = toF24(f32bits(g.fltBuf[0]))
			}
			g.fltIdx++
			g.fltBuf = g.fltBuf[:0]
		}
		return
	}
	if len(g.fltBuf) == 3 {
		if g.fltIdx < len(g.Float) {
			g.Float[g.fltIdx] = unpackF24x4(g.fltBuf[0], g.fltBuf[1], g.fltBuf[2])
		}
		g.fltIdx++
		g.fltBuf = g.fltBuf[:0]
	}
}

// drawArrays / drawElements run the pipeline (gpu_raster.go). Both first check
// for machinery the model does not have yet.
func (g *GPU) drawArrays() {
	g.Draws++
	g.draw(false)
}

func (g *GPU) drawElements() {
	g.Draws++
	g.draw(true)
}

// checkUnsupported halts on GPU features the pipeline does not model yet, so a
// run's reach stays honest.
func (g *GPU) checkUnsupported() bool {
	if g.Regs[regGeoConfig]&3 != 0 {
		g.m.CPU.Halt("gpu: draw with geometry-shader stage enabled (0x229=0x%08X) unimplemented", g.Regs[regGeoConfig])
		return true
	}
	return false
}

// DumpState prints the GPU state a bring-up wants to see.
func (g *GPU) DumpState() {
	fmt.Printf("gpu: %d draws; colorbuf 0x%08X depthbuf 0x%08X dim 0x%08X\n",
		g.Draws, g.Regs[regColorbufLoc]<<3, g.Regs[regDepthbufLoc]<<3, g.Regs[regFramebufDim])
}
