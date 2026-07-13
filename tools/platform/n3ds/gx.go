package n3ds

// gx.go emulates the GSP module's side of the GX command queue. On real hardware
// the GSP system process continuously polls the GX command FIFO in the shared
// memory an application registered, executes each command on the GPU (a memory
// fill, a display transfer, a P3D command list, a texture copy), and raises the
// matching completion interrupt. The application's GX command runner posts a
// command, marks it pending, and blocks until that interrupt clears the pending
// flag. Without a GSP-module counterpart the render loop stalls forever waiting
// for a completion that never comes.
//
// This models the GSP module's *dispatch and completion signalling* — it drains
// the queue and raises the right interrupt per command — but does not yet turn a
// P3D command list into pixels (that is the PICA200 GPU, a later phase). It is
// the exact analogue of the N64 RDP command reader, minus the rasteriser.

// The GX command queue lives at this offset in the GSP shared memory (GSP thread
// index 0). A 0x20-byte header, then up to 15 command slots of 0x20 bytes each.
const (
	gxQueueOff  = 0x800
	gxCmdStride = 0x20
	gxMaxCmds   = 15
	gxHdrIndex  = 0 // byte: index of the first unprocessed command
	gxHdrCount  = 1 // byte: number of queued commands
)

// GX command ids (GXCommandId), in the low byte of a command's first word.
const (
	gxCmdRequestDMA      = 0
	gxCmdProcessCmdList  = 1
	gxCmdMemoryFill      = 2
	gxCmdDisplayTransfer = 3
	gxCmdTextureCopy     = 4
	gxCmdFlushCache      = 5
)

// GXRecord is one GX command as the game posted it: the raw 8 words of its FIFO
// slot, plus — for a ProcessCommandList — a copy of the PICA200 command buffer it
// points at. Recorded when GXCapture is on; this is the Phase 4 instrument-first
// step (log what the game actually submits before implementing the GPU).
type GXRecord struct {
	Instr uint64    // instruction count at capture
	Words [8]uint32 // the raw command slot
	Buf   []byte    // ProcessCommandList only: the command-list bytes at capture time

	// Chained marks a buffer the command processor reached by a CMDBUF_JUMP
	// rather than one the game submitted. The game submits only the head of a
	// chain, so without these the capture misses most of what the GPU runs.
	Chained bool
}

// GXLog returns the commands captured so far (GXCapture must be on).
func (m *Machine) GXLog() []GXRecord { return m.gxLog }

// captureChainedList records a command buffer the GPU reached by a CMDBUF_JUMP.
// Unlike a submitted list this one is captured at execution time, which is the
// only time it exists as a coherent unit.
func (m *Machine) captureChainedList(addr, size uint32, buf []byte) {
	m.gxLog = append(m.gxLog, GXRecord{
		Instr:   m.CPU.Instrs,
		Words:   [8]uint32{gxCmdProcessCmdList, addr, size},
		Buf:     append([]byte(nil), buf...),
		Chained: true,
	})
}

// captureGX snapshots one GX command slot, and for a ProcessCommandList also the
// command-list buffer it references, before the FIFO is drained. The buffer copy
// matters: the game reuses/overwrites list memory between frames, so the bytes
// must be taken at submission time.
func (m *Machine) captureGX(cmd uint32, id uint32) {
	var r GXRecord
	r.Instr = m.CPU.Instrs
	for i := uint32(0); i < 8; i++ {
		r.Words[i] = m.ReadWord(cmd + i*4)
	}
	if id == gxCmdProcessCmdList {
		addr, size := r.Words[1], r.Words[2]
		if size > 0 && size < 0x1000000 {
			r.Buf = make([]byte, size)
			for i := uint32(0); i < size; i++ {
				r.Buf[i] = m.Read(addr + i)
			}
		}
	}
	m.gxLog = append(m.gxLog, r)
}

// gxMemoryFill floods [start,end) with a constant — the render-target clear.
// The control word's bits 8-9 give the fill width: 16, 24 or 32 bits (the
// capture shows 2 = 32-bit for both the RGBA8 colour clear and the D24S8 depth
// clear).
func (m *Machine) gxMemoryFill(start, value, end, ctl uint32) {
	defer m.profEnd(bucketGX, m.profStart()) // profile.go
	start, end = m.gpuAddrToVirt(start), m.gpuAddrToVirt(end)
	var unit uint32
	switch ctl >> 8 & 3 {
	case 0:
		unit = 2
	case 1:
		unit = 3
	case 2:
		unit = 4
	default:
		m.CPU.Halt("gx: MemoryFill control 0x%04X has width code 3 after %d instructions", ctl, m.CPU.Instrs)
		return
	}
	for a := start; a+unit <= end; a += unit {
		for j := uint32(0); j < unit; j++ {
			m.Write(a+j, byte(value>>(8*j)))
		}
	}
}

// gxTextureCopy streams size payload bytes from src to dst with per-line gaps:
// each dim word is gap<<16 | width, both in 2-byte units (derivation at the
// dispatch site). Zero dims degrade to a flat copy. Structural surprises halt.
func (m *Machine) gxTextureCopy(src, dst, size, inDim, outDim uint32) {
	defer m.profEnd(bucketGX, m.profStart()) // profile.go
	src, dst = m.gpuAddrToVirt(src), m.gpuAddrToVirt(dst)
	inW, inGap := inDim&0xFFFF*2, inDim>>16*2
	outW, outGap := outDim&0xFFFF*2, outDim>>16*2
	if inW == 0 && inGap == 0 {
		inW = size
	}
	if outW == 0 && outGap == 0 {
		outW = size
	}
	if inW == 0 || outW == 0 || size%inW != 0 {
		m.CPU.Halt("gx: TextureCopy dims in=0x%08X out=0x%08X size=0x%X don't divide after %d instructions",
			inDim, outDim, size, m.CPU.Instrs)
		return
	}
	sp, dp := src, dst
	sn, dn := uint32(0), uint32(0) // bytes into the current in/out line
	for i := uint32(0); i < size; i++ {
		m.Write(dp, m.Read(sp))
		sp, sn = sp+1, sn+1
		if sn == inW {
			sp, sn = sp+inGap, 0
		}
		dp, dn = dp+1, dn+1
		if dn == outW {
			dp, dn = dp+outGap, 0
		}
	}
	m.gpu.texCache = nil // the destination may be sampled as a texture
}

// gxDisplayTransfer converts a rendered (tiled, 8×8 Morton) colour buffer into
// the linear framebuffer the LCD scans out — the step that makes a frame
// visible. Dim words are height<<16|width; the screens are rotated, so a "row"
// here is 240 pixels wide and the heights are 400 (top) / 320 (bottom). The
// flags word carries the source and destination pixel formats (bits 8-10 and
// 12-14). Only what the game has been seen to use is implemented; anything else
// stops loudly.
func (m *Machine) gxDisplayTransfer(src, dst, srcDims, dstDims, flags uint32) {
	defer m.profEnd(bucketGX, m.profStart()) // profile.go
	srcW, srcH := srcDims&0xFFFF, srcDims>>16
	dstW, dstH := dstDims&0xFFFF, dstDims>>16
	inFmt, outFmt := flags>>8&7, flags>>12&7
	flip := flags&1 != 0
	if inFmt != 0 || flags>>1&1 != 0 || flags>>24&3 != 0 {
		m.CPU.Halt("gx: DisplayTransfer flags 0x%08X unimplemented (in-format %d, tiled-out %d, scale %d) after %d instructions",
			flags, inFmt, flags>>1&1, flags>>24&3, m.CPU.Instrs)
		return
	}
	src, dst = m.gpuAddrToVirt(src), m.gpuAddrToVirt(dst)

	// Bytes per destination pixel by format: RGBA8, RGB8, RGB565/RGB5A1/RGBA4.
	var dstBPP uint32
	switch outFmt {
	case 0:
		dstBPP = 4
	case 1:
		dstBPP = 3
	case 2, 3, 4:
		dstBPP = 2
	default:
		m.CPU.Halt("gx: DisplayTransfer out-format %d unimplemented after %d instructions", outFmt, m.CPU.Instrs)
		return
	}

	w, h := srcW, srcH
	if dstW < w {
		w = dstW
	}
	if dstH < h {
		h = dstH
	}
	tilesPerRow := srcW / 8
	for y := uint32(0); y < h; y++ {
		for x := uint32(0); x < w; x++ {
			// Source pixel from the tiled RGBA8 buffer; bytes are [A,B,G,R]
			// (the PICA convention cgfx_texture.go established).
			tile := (y/8)*tilesPerRow + x/8
			mo := (x&1)&1 | (y&1)<<1 | (x&2)<<1 | (y&2)<<2 | (x&4)<<2 | (y&4)<<3
			p := src + (tile*64+mo)*4
			a, b := m.Read(p), m.Read(p+1)
			g, r := m.Read(p+2), m.Read(p+3)

			dy := y
			if flip {
				dy = h - 1 - y
			}
			q := dst + (dy*dstW+x)*dstBPP
			switch outFmt {
			case 0: // RGBA8
				m.Write(q, a)
				m.Write(q+1, b)
				m.Write(q+2, g)
				m.Write(q+3, r)
			case 1: // RGB8, bytes [B,G,R]
				m.Write(q, b)
				m.Write(q+1, g)
				m.Write(q+2, r)
			case 2: // RGB565
				v := uint32(r>>3)<<11 | uint32(g>>2)<<5 | uint32(b>>3)
				m.Write(q, byte(v))
				m.Write(q+1, byte(v>>8))
			case 3: // RGB5A1
				v := uint32(r>>3)<<11 | uint32(g>>3)<<6 | uint32(b>>3)<<1 | uint32(a>>7)
				m.Write(q, byte(v))
				m.Write(q+1, byte(v>>8))
			case 4: // RGBA4
				v := uint32(r>>4)<<12 | uint32(g>>4)<<8 | uint32(b>>4)<<4 | uint32(a>>4)
				m.Write(q, byte(v))
				m.Write(q+1, byte(v>>8))
			}
		}
	}
	m.displayTransfers++
	// Remember where the frame went: the screenshot instrument (gpu_png.go)
	// reads the most recent linear framebuffer per screen. The top screen is
	// 400 lines tall, the bottom 320 — the height names the screen.
	rec := xferRecord{
		dst: dst, w: w, h: h, format: outFmt, bpp: dstBPP, stride: dstW,
		src: src, srcW: srcW, flip: flip,
	}
	if dstH >= 400 {
		m.lastXferTop = rec
	} else {
		m.lastXferBottom = rec
	}
}

// xferRecord describes the last DisplayTransfer per screen — the visible frame.
// It keeps the SOURCE geometry as well as the destination, because that is what
// maps a pixel of the panel back to the pixel of the render target it came from:
// the debugger pushes its per-pixel command provenance through exactly this
// transform so that a click on the screen the player sees can still name the draw
// that produced it.
type xferRecord struct {
	dst, w, h, format, bpp, stride uint32

	src  uint32 // virtual address of the window's first tiled pixel
	srcW uint32 // the source buffer's width — its tiling stride, not the window's
	flip bool
}

// processGXQueue drains any commands the game has posted to the GX FIFO and
// raises their completion interrupts, so the render loop's per-command waits are
// released. Cheap to call speculatively: it returns immediately when the queue
// is empty.
func (m *Machine) processGXQueue() {
	if m.gspSharedAddr == 0 {
		return
	}
	// Complete due in-flight commands BEFORE looking for new posts. Completion
	// is paced by the machine clock alone; gating it behind "the game posted
	// something new" hung Captain Toad's whole engine: its driver pauses the
	// command-list thread until outstanding submissions retire, so once it
	// stopped posting, the accepted-but-unpumped tail could never complete —
	// and with the sound thread keeping some thread always runnable, the idle
	// path (the only other pump site) never ran either.
	m.pumpGX()
	hdr := m.gspSharedAddr + gxQueueOff
	count := m.Read(hdr + gxHdrCount)
	if count == 0 {
		return
	}
	idx := m.Read(hdr + gxHdrIndex)

	for i := byte(0); i < count; i++ {
		slot := (uint32(idx) + uint32(i)) % gxMaxCmds
		cmd := m.gspSharedAddr + gxQueueOff + gxCmdStride + slot*gxCmdStride
		id := uint32(m.Read(cmd)) & 0x1F
		if m.GXCapture {
			m.captureGX(cmd, id)
		}
		var w [8]uint32
		for j := uint32(0); j < 8; j++ {
			w[j] = m.ReadWord(cmd + j*4)
		}
		// Completion is ASYNCHRONOUS, as on hardware: the command is queued
		// with a deadline and executes (raising its interrupt) only when the
		// instruction clock gets there — one engine, strictly in order. The
		// game's own graphics driver depends on the in-flight window: it
		// retires submissions on completion interrupts, and with synchronous
		// completion its 32-entry submission ring only ever filled (13 → 32
		// across the boot) until the first blocking submit deadlocked against
		// it. The latencies are model parameters (hardware-plausible orders of
		// magnitude), not measurements.
		base := m.instrs
		if n := len(m.gxPending); n > 0 && m.gxPending[n-1].Deadline > base {
			base = m.gxPending[n-1].Deadline
		}
		m.gxPending = append(m.gxPending, gxPendingCmd{Words: w, Deadline: base + gxLatency(id)})
	}

	// The queue is drained: advance the read index and zero the count, exactly
	// as the GSP module does after servicing the batch (accepting a command is
	// immediate; completing it is not).
	m.Write(hdr+gxHdrIndex, byte((uint32(idx)+uint32(count))%gxMaxCmds))
	m.Write(hdr+gxHdrCount, 0)

	m.pumpGX()
}

// gxPendingCmd is one accepted-but-not-completed GX command. Fields are
// exported for the savestate gob encoder.
type gxPendingCmd struct {
	Words    [8]uint32
	Deadline uint64 // machine-monotonic instruction count at which it executes
}

// gxLatency is the modelled completion delay per command kind, in instructions.
// What matters for correctness is that completion is visibly later than
// submission and strictly ordered; the magnitudes are chosen so a frame's
// worth of commands (~1 list + hundreds of DMAs) costs a few percent of the
// ~4.5M-instruction frame, not more.
func gxLatency(id uint32) uint64 {
	switch id {
	case gxCmdProcessCmdList:
		return 16384
	case gxCmdMemoryFill:
		return 2048
	case gxCmdDisplayTransfer, gxCmdTextureCopy:
		return 4096
	case gxCmdRequestDMA:
		return 256
	}
	return 0 // FlushCache: cache maintenance, no work and no interrupt
}

// gxDeadline reports the earliest pending completion (and whether one exists),
// so the run loop's idle fast-forward can jump to it instead of the next
// VBlank.
func (m *Machine) gxDeadline() (uint64, bool) {
	if len(m.gxPending) == 0 {
		return 0, false
	}
	return m.gxPending[0].Deadline, true
}

// pumpGX executes every pending GX command whose deadline has passed, raising
// its completion interrupt at execution time.
func (m *Machine) pumpGX() {
	for len(m.gxPending) > 0 && m.instrs >= m.gxPending[0].Deadline {
		w := m.gxPending[0].Words
		m.gxPending = m.gxPending[1:]

		var raised []byte
		switch w[0] & 0x1F {
		case gxCmdProcessCmdList:
			m.framesSubmitted++ // a PICA200 command list — a rendered frame's geometry
			m.gpu.Execute(w[1], w[2])
			if m.CPU.Halted {
				return
			}
			raised = append(raised, gspIntP3D)
		case gxCmdMemoryFill:
			// Two independent fill engines; a zero start address disables one.
			// Control word: low half engine 0, high half engine 1.
			if w[1] != 0 {
				m.gxMemoryFill(w[1], w[2], w[3], w[7]&0xFFFF)
				raised = append(raised, gspIntPSC0)
			}
			if w[4] != 0 {
				m.gxMemoryFill(w[4], w[5], w[6], w[7]>>16)
				raised = append(raised, gspIntPSC1)
			}
		case gxCmdDisplayTransfer:
			m.gxDisplayTransfer(w[1], w[2], w[3], w[4], w[5])
			raised = append(raised, gspIntPPF)
		case gxCmdTextureCopy:
			// Gap-aware byte copy (the PPF engine's "texture copy" mode). The
			// first observed slot — 1F000080 → 1629FA80, size 0x5DC00, in dim
			// 0x000001E0, out dim 0x002001E0 — pins the layout: dim = gap<<16 |
			// width in 2-byte units; 0x1E0 units = 960 bytes = one 240-px RGBA8
			// framebuffer line, 400 lines = the size exactly, and the output's
			// 0x20-unit gap pads each line to a 1024-byte stride (a 256-px
			// power-of-two texture): the game grabs the rendered frame as a
			// texture for its screen transition.
			m.gxTextureCopy(w[1], w[2], w[3], w[4], w[5])
			raised = append(raised, gspIntPPF)
		case gxCmdRequestDMA:
			// A plain memory move: the game stages vertex/texture data from its
			// linear heap into VRAM with these before drawing.
			for j := uint32(0); j < w[3]; j++ {
				m.Write(w[2]+j, m.Read(w[1]+j))
			}
			m.gpu.texCache = nil // texture memory changed under the cache
			raised = append(raised, gspIntDMA)
		}

		for _, id := range raised {
			m.pushGSPInterrupt(id)
		}
		if len(raised) > 0 {
			m.signalGSPEvent()
		}
	}
}
