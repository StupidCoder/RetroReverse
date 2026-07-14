package dsmachine

// The 3D engine's front end: the command FIFO and the register block.
//
// The DS's geometry engine is not addressed like a set of registers — it is *fed*.
// Commands arrive either as words written to individual command ports (0x04000440
// upward, one address per command) or, far more often, packed into the GXFIFO at
// 0x04000400: four one-byte command codes in a word, followed by their parameters.
// NitroSDK's display lists are in the packed form and reach the FIFO by DMA, so a
// machine that only implements the command ports sees almost nothing.
//
// Everything below the FIFO — the matrix stacks, the vertex pipeline, the lighting
// and the rasteriser — is in gpu3d_geom.go and gpu3d_raster.go. This file is the
// part the CPU can see: what it writes, and what it can read back.

const (
	regGXFIFO        = 0x04000400
	regGXCMDPORT     = 0x04000440 // MTX_MODE; one port per command, four bytes apart...
	regGXCMDPORTEnd  = 0x040005C8 // ...up to VEC_TEST, the last of them
	regGXSTAT        = 0x04000600
	regRAMCOUNT      = 0x04000604
	regDISP3DCNT     = 0x04000060
	regCLEARCOLOR    = 0x04000350
	regCLEARDEPTH    = 0x04000354
	regCLIPMTXRESULT = 0x04000640
	regVECMTXRESULT  = 0x04000680
	regPOSRESULT     = 0x04000620
	regVECRESULT     = 0x04000630
)

// gpu3d holds the command front end. `pending` is the command being assembled: a
// packed word names up to four commands, and each takes a known number of parameter
// words, so the FIFO is a small state machine rather than a queue of opcodes.
type gpu3d struct {
	fifo []uint32 // unprocessed command+parameter words

	packed []uint8  // the command bytes still to be consumed from a packed word
	cmd    uint8    // the command currently collecting parameters
	params []uint32 // parameters collected so far
	need   int      // how many that command wants

	geom geom   // the matrix stacks and the vertex pipeline (gpu3d_geom.go)
	rast raster // the rasteriser's buffers (gpu3d_raster.go)

	// The registers the CPU writes and reads back.
	regs map[uint32]uint32

	swapPending bool // SWAP_BUFFERS was issued; the swap happens at the next VBlank
	swapMode    uint32

	// Per-frame counters, latched at the swap: what the geometry engine actually
	// handed the rasteriser. A black screen with polygons is a shading bug; a black
	// screen with none is a geometry bug, and they are not investigated the same way.
	lastPolys int
	swaps     int

	// cmdHist counts how many times each geometry command has been executed — the
	// DS's -gxdump. A frame that draws nothing has a shape here: no BEGIN_VTXS means
	// the display lists never arrived, plenty of BEGIN_VTXS with no VTX means the
	// parameters are being dropped, and everything present means the bug is downstream.
	cmdHist [256]int

	// OnCmd, if set, reports every command the FIFO decodes with its parameters — the
	// DS's command-list dump.
	OnCmd func(cmd uint8, p []uint32)
}

func newGPU3D() *gpu3d {
	g := &gpu3d{regs: map[uint32]uint32{}}
	g.geom.reset()
	g.rast.reset()
	return g
}

// gxParams is how many parameter words each command takes. A command with zero
// parameters executes the moment it is decoded; the rest wait for their words.
// (Index by command byte; commands not listed take no parameters.)
var gxParams = map[uint8]int{
	0x10: 1,  // MTX_MODE
	0x11: 0,  // MTX_PUSH
	0x12: 1,  // MTX_POP
	0x13: 1,  // MTX_STORE
	0x14: 1,  // MTX_RESTORE
	0x15: 0,  // MTX_IDENTITY
	0x16: 16, // MTX_LOAD_4x4
	0x17: 12, // MTX_LOAD_4x3
	0x18: 16, // MTX_MULT_4x4
	0x19: 12, // MTX_MULT_4x3
	0x1A: 9,  // MTX_MULT_3x3
	0x1B: 3,  // MTX_SCALE
	0x1C: 3,  // MTX_TRANS
	0x20: 1,  // COLOR
	0x21: 1,  // NORMAL
	0x22: 1,  // TEXCOORD
	0x23: 2,  // VTX_16
	0x24: 1,  // VTX_10
	0x25: 1,  // VTX_XY
	0x26: 1,  // VTX_XZ
	0x27: 1,  // VTX_YZ
	0x28: 1,  // VTX_DIFF
	0x29: 1,  // POLYGON_ATTR
	0x2A: 1,  // TEXIMAGE_PARAM
	0x2B: 1,  // PLTT_BASE
	0x30: 1,  // DIF_AMB
	0x31: 1,  // SPE_EMI
	0x32: 1,  // LIGHT_VECTOR
	0x33: 1,  // LIGHT_COLOR
	0x34: 32, // SHININESS
	0x40: 1,  // BEGIN_VTXS
	0x41: 0,  // END_VTXS
	0x50: 1,  // SWAP_BUFFERS
	0x60: 1,  // VIEWPORT
	0x70: 3,  // BOX_TEST
	0x71: 2,  // POS_TEST
	0x72: 1,  // VEC_TEST
}

// fifoBelowHalf reports whether the command FIFO has drained below half full — the
// condition a mode-7 DMA channel waits on before topping it up.
func (g *gpu3d) fifoBelowHalf() bool { return len(g.fifo) < 128 }

func (g *gpu3d) writeReg(a, w uint32) {
	switch {
	case a >= regGXFIFO && a < regGXCMDPORT:
		// The GXFIFO is not one address — it is the whole 64-byte window from
		// 0x04000400 to 0x0400043F, and every word written anywhere in it is pushed
		// into the same queue.
		//
		// That is not a quirk, it is the point: a game feeds the geometry engine with
		// burst stores, and an ARM block store (STMIA) or a DMA walks the destination
		// address up as it goes — 0x400, 0x404, 0x408, and on. Match only the first
		// address and the first word of every burst is accepted while the rest are
		// dropped on the floor. The command stream that survives is still *plausible*
		// — it decodes cleanly, it draws polygons, it swaps buffers — it is simply
		// missing most of itself, so matrices are never multiplied, the camera never
		// arrives, and 95% of the geometry lands behind the eye and is clipped away.
		g.pushPacked(w)
		return
	case a >= regGXCMDPORT && a <= regGXCMDPORTEnd:
		// A direct command port: the address names the command, and each write is one
		// parameter word for it. The block runs from MTX_MODE at 0x04000440 to VEC_TEST
		// at 0x040005C8 — a 0x100-byte window stops at 0x0400053F and quietly loses
		// every command from SWAP_BUFFERS (0x04000540) upward, which is to say: the
		// geometry all arrives, the polygons all build up, and the frame is never once
		// swapped to the screen.
		g.pushDirect(uint8((a-regGXCMDPORT)/4)+0x10, w)
		return
	case a == regGXSTAT:
		// Writing GXSTAT can reset the matrix-stack error flag; nothing else here.
		return
	}
	g.regs[a] = w
}

func (g *gpu3d) readReg(a uint32) uint32 {
	switch {
	case a == regGXSTAT:
		return g.gxstat()
	case a == regRAMCOUNT:
		return uint32(len(g.geom.polys)) | uint32(len(g.geom.verts))<<16
	case a >= regCLIPMTXRESULT && a < regCLIPMTXRESULT+64:
		return uint32(g.geom.clip().m[(a-regCLIPMTXRESULT)/4])
	case a >= regVECMTXRESULT && a < regVECMTXRESULT+36:
		return uint32(g.geom.vecMtx3x3((a - regVECMTXRESULT) / 4))
	case a >= regPOSRESULT && a < regPOSRESULT+16:
		return uint32(g.geom.posResult[(a-regPOSRESULT)/4])
	case a >= regVECRESULT && a < regVECRESULT+8:
		// The three VEC_TEST components are 16-bit ports, but the CPU reads this block
		// as words — so one word carries two components, and indexing it as if each
		// word were one component silently drops Y.
		i := (a - regVECRESULT) / 2
		v := uint32(uint16(g.geom.vecResult[i]))
		if i+1 < 3 {
			v |= uint32(uint16(g.geom.vecResult[i+1])) << 16
		}
		return v
	}
	return g.regs[a]
}

// gxstat reports the engine's status. The bits that matter to software are the FIFO
// level and, above all, bit 27 — "the geometry engine is busy". A game waits on that
// before touching the matrix stack, and a model that leaves it stuck at "busy" hangs
// the boot as surely as a missing register.
func (g *gpu3d) gxstat() uint32 {
	n := uint32(len(g.fifo))
	if n > 256 {
		n = 256
	}
	v := n << 16
	if n == 0 {
		v |= 1 << 26 // FIFO empty
	}
	if n < 128 {
		v |= 1 << 25 // FIFO less than half full
	}
	v |= uint32(g.geom.stackDepth()) << 8
	return v // bit 27 (busy) stays clear: our commands complete as they are issued
}

// pushPacked takes a word written to the GXFIFO. The first word of a packed group
// carries up to four command bytes; the words that follow are their parameters, in
// order. A command byte of zero is a no-op pad.
func (g *gpu3d) pushPacked(w uint32) {
	if g.cmd == 0 && len(g.packed) == 0 {
		for i := 0; i < 4; i++ {
			if b := uint8(w >> (8 * i)); b != 0 {
				g.packed = append(g.packed, b)
			}
		}
		g.nextPacked()
		return
	}
	g.feed(w)
}

// nextPacked starts the next command from the packed group, executing straight away
// any that take no parameters.
func (g *gpu3d) nextPacked() {
	for len(g.packed) > 0 {
		c := g.packed[0]
		n := gxParams[c]
		if n > 0 {
			g.packed = g.packed[1:]
			g.cmd, g.need, g.params = c, n, g.params[:0]
			return
		}
		g.packed = g.packed[1:]
		g.exec(c, nil)
	}
	g.cmd, g.need = 0, 0
}

// feed supplies one parameter word to the command in flight.
func (g *gpu3d) feed(w uint32) {
	if g.cmd == 0 {
		return
	}
	g.params = append(g.params, w)
	if len(g.params) < g.need {
		return
	}
	g.exec(g.cmd, g.params)
	g.cmd, g.need = 0, 0
	g.nextPacked()
}

// pushDirect handles a write to one of the per-command ports.
func (g *gpu3d) pushDirect(cmd uint8, w uint32) {
	n := gxParams[cmd]
	if n == 0 {
		g.exec(cmd, nil)
		return
	}
	if g.cmd != cmd {
		g.cmd, g.need, g.params = cmd, n, g.params[:0]
	}
	g.params = append(g.params, w)
	if len(g.params) >= g.need {
		g.exec(g.cmd, g.params)
		g.cmd, g.need = 0, 0
	}
}

// vblank consumes a pending buffer swap. The DS does not swap when the command is
// issued: SWAP_BUFFERS marks the frame complete, and the hardware presents it at the
// next vertical blank, which is what keeps a half-built frame off the screen.
func (g *gpu3d) vblank(m *Machine) {
	if !g.swapPending {
		return
	}
	g.swapPending = false
	g.lastPolys = len(g.geom.polys)
	g.swaps++
	g.render(m)
	g.geom.beginFrame()
}
