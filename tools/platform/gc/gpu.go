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
	Buf    []byte       // command bytes not yet assembled into a whole command
	Census [256]uint64  // how many of each opcode the FIFO has carried — the plumbing gate
	CPReg  [0x100]uint32 // the CP registers loaded from the stream (vertex descriptors etc.)
}

// feed takes the next burst of FIFO bytes and consumes every complete command in it.
func (g *gpu) feed(m *Machine, b []byte) {
	g.Buf = append(g.Buf, b...)
	for g.step(m) {
	}
}

// step consumes one complete command from the head of Buf. It returns false when Buf does
// not yet hold a whole command (so the parser waits for the next burst) or when the machine
// has halted — either an unknown opcode or a command the pipe cannot yet interpret, both of
// which stop loudly rather than guess a length and desynchronise the entire stream.
func (g *gpu) step(m *Machine) bool {
	if m.CPU.Halted || len(g.Buf) == 0 {
		return false
	}
	op := g.Buf[0]
	switch {
	case op == 0x00: // NOP — padding to a 32-byte boundary
		g.Census[0]++
		g.Buf = g.Buf[1:]

	case op == 0x08: // load a CP register: opcode, u8 address, u32 value
		if len(g.Buf) < 6 {
			return false
		}
		g.CPReg[g.Buf[1]] = be32(g.Buf[2:])
		g.Census[0x08]++
		g.Buf = g.Buf[6:]

	case op == 0x10: // load XF registers: opcode, u32 (count-1<<16 | address), then count words
		if len(g.Buf) < 5 {
			return false
		}
		cnt := int((be32(g.Buf[1:])>>16)&0x0F) + 1
		total := 5 + 4*cnt
		if len(g.Buf) < total {
			return false
		}
		g.Census[0x10]++
		g.Buf = g.Buf[total:]

	case op >= 0x20 && op <= 0x38 && op&0x07 == 0: // load indexed A/B/C/D into XF memory
		if len(g.Buf) < 5 {
			return false
		}
		g.Census[op]++
		g.Buf = g.Buf[5:]

	case op == 0x40: // call a display list: opcode, u32 address, u32 size
		if len(g.Buf) < 9 {
			return false
		}
		g.Census[0x40]++
		g.Buf = g.Buf[9:]
		// The list's own commands live in memory and are not walked yet; when a frame first
		// calls one this becomes the place to recurse. Until then it is noted, not obeyed.
		m.logf("CP: display-list call not yet interpreted (its commands are skipped)")

	case op == 0x48: // invalidate the vertex cache
		g.Census[0x48]++
		g.Buf = g.Buf[1:]

	case op == 0x61: // load a BP (pixel-engine/blending) register: opcode, u32 (addr<<24 | data)
		if len(g.Buf) < 5 {
			return false
		}
		v := be32(g.Buf[1:])
		g.Census[0x61]++
		g.Buf = g.Buf[5:]
		g.loadBP(m, uint8(v>>24), v&0x00FFFFFF)

	case op >= 0x80: // a draw primitive: opcode, u16 vertex count, then the vertices
		// Vertex fetch is the next stage of the pipe. Until it exists the command processor
		// cannot know how many bytes a primitive occupies, so it stops here — naming the
		// opcode — rather than guess and lose the rest of the stream.
		m.CPU.Halt("CP: draw primitive 0x%02X reached — vertex fetch not yet implemented", op)
		return false

	default:
		m.CPU.Halt("CP: unknown FIFO opcode 0x%02X", op)
		return false
	}
	return true
}

// loadBP acts on the pixel-engine registers a game writes into the command stream to
// synchronise with the pipe. Three of them are interrupts-in-waiting: the draw-done marker,
// and the two token forms. The rest configure blending and are the graphics stage's to
// interpret; here they are simply counted by their opcode.
func (g *gpu) loadBP(m *Machine, reg uint8, data uint32) {
	switch reg {
	case 0x45: // BPMEM_SETDRAWDONE — value 2 means "the pipe has drained to here": finish
		if data&0xFF == 0x02 {
			m.pe.setFinish(m)
		}
	case 0x47: // BPMEM_PE_TOKEN — plant a token value, no interrupt
		m.pe.setToken(m, uint16(data), false)
	case 0x48: // BPMEM_PE_TOKEN_INT — plant a token value and raise the token interrupt
		m.pe.setToken(m, uint16(data), true)
	}
}
