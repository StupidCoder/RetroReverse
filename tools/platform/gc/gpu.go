package gc

// gpu.go is the first step of the Flipper graphics pipe: the command processor that reads
// the FIFO the write-gather pipe fills. This is Phase 3 stage one — the plumbing — and it
// does exactly two things: it walks the command stream and accounts for every opcode, and
// it acts on the pixel-engine markers a frame-timed game plants in that stream to ask the
// hardware to interrupt it when the pipe has drained.
//
// That second job is why this file has to exist before any pixel is drawn. A GameCube game
// submits its frame, calls GXDrawDone, and sleeps — waiting for the PE_FINISH interrupt
// that says the pipe reached the "draw done" token. Nothing in the machine raised that
// interrupt before, so the game's first frame-sync slept forever and the boot parked in the
// OS scheduler's idle loop. Recognising the token here and raising the interrupt is what
// lets the game's own handler run, wake the sleeping thread, and continue.
//
// The command stream comes straight from the write-gather pipe (see wgPipe in devices.go),
// so this interpreter is fed the bytes as the CPU bursts them. A single command can span
// two bursts, so the parser keeps an incomplete tail in Buf and only acts once a whole
// command has arrived — which also means Buf is part of the machine's saved state.

type gpu struct {
	Buf    []byte      // command bytes not yet assembled into a whole command
	Census [256]uint64 // how many of each opcode the FIFO has carried — the plumbing gate

	// Per-primitive pixel accounting for the RR_GC_DRAWTRACE line: where a draw's pixels
	// went. Unexported — diagnostic state, not machine state, so it stays out of snapshots.
	pixWritten, pixZRej, pixARej int

	// inDisplayList guards against a display list calling another — the hardware forbids
	// it, and honouring the same limit keeps the interpreter non-recursive. Transient within
	// one feed, so it is not machine state.
	inDisplayList bool
	CPReg         [0x100]uint32  // the CP registers loaded from the stream (vertex descriptors etc.)
	BP            [0x100]uint32  // the BP (pixel-engine) registers loaded from the stream
	XFMem         [0x1060]uint32 // the XF registers and matrix memory (see gpu_xf.go)

	// The TEV colour and constant registers share the same eight BP addresses (0xE0..0xE7),
	// told apart by a type bit in each write, so they cannot both live in BP[] — the last
	// write of either kind would clobber the other. They are routed to these two banks at
	// load time (see loadBP) and read back by the combiner in gpu_tev.go. Each register is a
	// low/high word pair.
	TevColorReg [4][2]uint32 // PREV, REG0, REG1, REG2 — the combiner's C0/C1/C2 inputs
	TevKonstReg [4][2]uint32 // KONST0..3 — the combiner's KONST input, selected by KSEL

	EFB  []uint32 // Flipper's embedded framebuffer, RGBA8888 per pixel (see gpu_efb.go)
	ZBuf []uint32 // the 24-bit depth buffer that pairs with the EFB (see gpu_raster.go)

	// Tlut is the palette half of texture memory: the colour tables the game DMAs in ahead
	// of a paletted draw (BP 0x64 names the main-memory source, BP 0x65 the TMEM slot and
	// length — the load happens at the 0x65 write, register layouts pinned from the game's
	// own GXInitTlutObj/GXLoadTlut at 0x801F6F24/0x801F7128). Indexed by the offset within
	// the high TMEM bank, i.e. tmem address minus 0x80000. Allocated on the first load.
	Tlut []byte
}

// xfStore records one XF register or matrix-memory word the command stream loaded.
func (g *gpu) xfStore(addr int, val uint32) {
	if addr >= 0 && addr < len(g.XFMem) {
		g.XFMem[addr] = val
	}
}

// feed takes the next burst of FIFO bytes and consumes every complete command in it.
func (g *gpu) feed(m *Machine, b []byte) {
	g.Buf = append(g.Buf, b...)
	for g.step(m) {
	}
}

// gxCmd is called by step once a whole command has been parsed off the head of Buf and
// before any of it executes. It numbers the command, offers it to the debugger's command
// hook, and answers whether the interpreter may go on: a replay asked to stop after command
// k declines the next one here, leaving it unconsumed in Buf.
//
// Firing before execution rather than after is what makes per-pixel provenance work — a
// draw's pixels have to be attributable to a command that has already been announced.
func (g *gpu) gxCmd(m *Machine, op uint8, words []uint32) bool {
	if m.gxStopAfter > 0 && m.gxCmdCount >= m.gxStopAfter {
		m.gxStopped = true
		m.StopRequested = true
		return false
	}
	m.gxCmdCount++
	if m.OnGXCmd != nil {
		m.OnGXCmd(m, op, words)
	}
	return true
}

// step consumes one complete command from the head of Buf. It returns false when Buf does
// not yet hold a whole command (so the parser waits for the next burst), when a replay has
// asked to stop here, or when the machine has halted — either an unknown opcode or a command
// the pipe cannot yet interpret, both of which stop loudly rather than guess a length and
// desynchronise the entire stream.
//
// Each case checks its command is complete FIRST, then announces it through gxCmd, then
// executes. That order is the contract: a command is announced only if it is whole and is
// about to run.
func (g *gpu) step(m *Machine) bool {
	if m.CPU.Halted || m.gxStopped || len(g.Buf) == 0 {
		return false
	}
	op := g.Buf[0]
	switch {
	case op == 0x00: // NOP — padding to a 32-byte boundary
		if !g.gxCmd(m, op, nil) {
			return false
		}
		g.Census[0]++
		g.Buf = g.Buf[1:]

	case op == 0x08: // load a CP register: opcode, u8 address, u32 value
		if len(g.Buf) < 6 {
			return false
		}
		if !g.gxCmd(m, op, []uint32{uint32(g.Buf[1]), be32(g.Buf[2:])}) {
			return false
		}
		g.CPReg[g.Buf[1]] = be32(g.Buf[2:])
		g.Census[0x08]++
		g.Buf = g.Buf[6:]

	case op == 0x10: // load XF registers: opcode, u32 (count-1<<16 | address), then count words
		if len(g.Buf) < 5 {
			return false
		}
		cmd := be32(g.Buf[1:])
		cnt := int((cmd>>16)&0x0F) + 1
		total := 5 + 4*cnt
		if len(g.Buf) < total {
			return false
		}
		addr := int(cmd & 0xFFFF)
		words := make([]uint32, 0, cnt+1)
		words = append(words, cmd)
		for k := 0; k < cnt; k++ {
			words = append(words, be32(g.Buf[5+4*k:]))
		}
		if !g.gxCmd(m, op, words) {
			return false
		}
		for k := 0; k < cnt; k++ {
			g.xfStore(addr+k, be32(g.Buf[5+4*k:]))
		}
		g.Census[0x10]++
		g.Buf = g.Buf[total:]

	case op >= 0x20 && op <= 0x38 && op&0x07 == 0: // load indexed A/B/C/D into XF memory
		if len(g.Buf) < 5 {
			return false
		}
		if !g.gxCmd(m, op, []uint32{be32(g.Buf[1:])}) {
			return false
		}
		g.Census[op]++
		g.Buf = g.Buf[5:]

	case op == 0x40: // call a display list: opcode, u32 address, u32 size
		if len(g.Buf) < 9 {
			return false
		}
		addr := be32(g.Buf[1:])
		size := be32(g.Buf[5:])
		if !g.gxCmd(m, op, []uint32{addr, size}) {
			return false
		}
		g.Census[0x40]++
		g.Buf = g.Buf[9:]
		g.callDisplayList(m, addr, size)

	case op == 0x48: // invalidate the vertex cache
		if !g.gxCmd(m, op, nil) {
			return false
		}
		g.Census[0x48]++
		g.Buf = g.Buf[1:]

	case op == 0x61: // load a BP (pixel-engine/blending) register: opcode, u32 (addr<<24 | data)
		if len(g.Buf) < 5 {
			return false
		}
		v := be32(g.Buf[1:])
		if !g.gxCmd(m, op, []uint32{v}) {
			return false
		}
		g.Census[0x61]++
		g.Buf = g.Buf[5:]
		g.loadBP(m, uint8(v>>24), v&0x00FFFFFF)

	case op >= 0x80: // a draw primitive: opcode, u16 vertex count, then the vertices
		if len(g.Buf) < 3 {
			return false
		}
		vat := int(op & 0x07)
		vsize := g.vertexSize(vat)
		if vsize == 0 {
			m.CPU.Halt("CP: draw 0x%02X has zero-size vertices — VCD/VAT not decoded (CP regs 0x50/0x60/0x70)", op)
			return false
		}
		count := int(be16(g.Buf[1:]))
		total := 3 + count*vsize
		if len(g.Buf) < total {
			return false
		}
		if !g.gxCmd(m, op, []uint32{uint32(vat), uint32(count)}) {
			return false
		}
		g.Census[op]++
		// The vertex size is computed exactly so the parser lands on the next real command; if
		// it is wrong the stream desynchronises and the very next opcode halts the run, which is
		// the check that keeps this honest.
		g.drawPrimitive(m, uint32(op&0xF8), vat, vsize, g.Buf[3:total])
		g.Buf = g.Buf[total:]

	default:
		m.CPU.Halt("CP: unknown FIFO opcode 0x%02X", op)
		return false
	}
	return true
}

// GXName is the FIFO opcode's name, for a debugger's command list.
func GXName(op uint8) string {
	switch {
	case op == 0x00:
		return "NOP"
	case op == 0x08:
		return "CP_LOADREG"
	case op == 0x10:
		return "XF_LOADREGS"
	case op >= 0x20 && op <= 0x38 && op&0x07 == 0:
		return "XF_LOADINDEX"
	case op == 0x40:
		return "CALL_DISPLAYLIST"
	case op == 0x48:
		return "INVAL_VTXCACHE"
	case op == 0x61:
		return "BP_LOADREG"
	case op >= 0x80:
		return "DRAW_" + primName(op&0xF8)
	}
	return "UNKNOWN"
}

// primName names a draw's primitive type — the top five bits of a draw opcode.
func primName(prim uint8) string {
	switch prim {
	case 0x80:
		return "QUADS"
	case 0x88:
		return "QUADS2"
	case 0x90:
		return "TRIANGLES"
	case 0x98:
		return "TRIANGLESTRIP"
	case 0xA0:
		return "TRIANGLEFAN"
	case 0xA8:
		return "LINES"
	case 0xB0:
		return "LINESTRIP"
	case 0xB8:
		return "POINTS"
	}
	return "PRIM"
}

// callDisplayList executes a display list: a block of the same FIFO commands, pre-built in
// main memory, that the CALL opcode splices into the stream. The commands run through the
// same interpreter with the same registers — a state change made inside a list persists after
// it, exactly as if the CPU had burst the bytes itself. The hardware does not allow a list to
// call another list, so a nested call halts rather than recursing; and a list that ends in
// the middle of a command is a desynchronised stream, which halts just as loudly as an
// unknown opcode does.
func (g *gpu) callDisplayList(m *Machine, addr, size uint32) {
	if g.inDisplayList {
		m.CPU.Halt("CP: display list at 0x%08X calls another display list", addr)
		return
	}
	a := phys(addr)
	if int(a)+int(size) > len(m.RAM) {
		m.CPU.Halt("CP: display list 0x%08X+0x%X runs past the end of RAM", addr, size)
		return
	}
	saved := g.Buf
	g.Buf = append([]byte(nil), m.RAM[a:a+size]...)
	g.inDisplayList = true
	for g.step(m) {
	}
	leftover := len(g.Buf)
	g.inDisplayList = false
	g.Buf = saved
	// Bytes left over mean the list ended in the middle of a command — a desynchronised
	// stream, and as loud a fault as an unknown opcode. Unless a replay stopped us on
	// purpose: then the leftover is the rest of the list we deliberately declined to run,
	// and the machine it happened on is a scratch one about to be thrown away.
	if leftover != 0 && !m.CPU.Halted && !m.gxStopped {
		m.CPU.Halt("CP: display list at 0x%08X ended mid-command (%d bytes left of %d)",
			addr, leftover, size)
	}
}

// loadBP acts on the pixel-engine registers a game writes into the command stream to
// synchronise with the pipe. Three of them are interrupts-in-waiting: the draw-done marker,
// and the two token forms. The rest configure blending and are the graphics stage's to
// interpret; here they are simply counted by their opcode.
func (g *gpu) loadBP(m *Machine, reg uint8, data uint32) {
	g.BP[reg] = data
	switch reg {
	case 0x45: // BPMEM_SETDRAWDONE — value 2 means "the pipe has drained to here": finish
		if data&0xFF == 0x02 {
			m.pe.setFinish(m)
		}
	case 0x47: // BPMEM_PE_TOKEN — plant a token value, no interrupt
		m.pe.setToken(m, uint16(data), false)
	case 0x48: // BPMEM_PE_TOKEN_INT — plant a token value and raise the token interrupt
		m.pe.setToken(m, uint16(data), true)
	case 0x52: // BPMEM_TRIGGER_EFB_COPY — execute the pixel-engine copy that ends a frame
		g.copyDisplay(m, data)
	case 0x65: // BPMEM_LOADTLUT1 — DMA a palette from main memory into texture memory.
		// The source RAM address arrived in BP 0x64 (physical >> 5, from GXInitTlutObj at
		// 0x801F6F24); this trigger word carries the destination TMEM slot in 512-byte
		// units within the high bank (bits 9:0, tmem-0x80000 >> 9) and the length in
		// 16-entry lines (bits 20:10), from GXLoadTlut's region word at 0x801F7128.
		g.loadTlut(m, data)
	}

	// The eight TEV register addresses carry either a colour/output register or a constant,
	// distinguished by bit 23 of the payload. Route each to its own bank so one kind does not
	// overwrite the other in BP[]. Bits [22:23] would select which register; the low bit of
	// the address is the low/high word.
	if reg >= 0xE0 && reg <= 0xE7 {
		pair := (reg - 0xE0) / 2
		word := (reg - 0xE0) & 1
		if data&(1<<23) != 0 {
			g.TevKonstReg[pair][word] = data
		} else {
			g.TevColorReg[pair][word] = data
		}
	}
}

// be16 reads a big-endian halfword from the head of a byte slice.
func be16(b []byte) uint16 { return uint16(b[0])<<8 | uint16(b[1]) }
