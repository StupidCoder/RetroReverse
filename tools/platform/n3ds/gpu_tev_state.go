package n3ds

// gpu_tev_state.go decodes the texture-environment configuration once per draw.
//
// The TEV is six stages, and every fragment ran all six of them out of the raw
// register file: five register loads per stage, then the source selectors, operand
// multiplexers, combine ops and scales dug out of those words a nibble at a time —
// six times over, for each of 227,000 fragments a frame. None of it depends on the
// fragment. It is the same argument that made the vertex shader's decoded-instruction
// cache worth 9%, pointed at the other end of the pipeline.
//
// Nothing here can change between the fragments of a draw: a register write only
// happens in the command processor, which is not running while a draw is.
//
// Like fbstate() and lightstate(), this is resolved once in draw() and handed down.

// tevSrc/tevOp are the decoded selectors of one operand.
type tevOperand struct {
	src uint8 // the TEV source id (0-15)
	op  uint8 // the operand multiplexer
}

// tevStage is one decoded combiner stage.
type tevStage struct {
	colr [3]tevOperand
	alph [3]tevOperand

	combC, combA uint8 // combine ops
	scaleC       uint8 // shift applied to the colour result
	scaleA       uint8 // shift applied to the alpha result
	konst        rgba

	updC, updA bool // does this stage write the combiner buffer (stages 0-3 only)
}

// tevState is the whole fragment back end for a draw: the six stages, the combiner
// buffer's initial colour, and the alpha test.
type tevState struct {
	stages [6]tevStage

	bufColor rgba // 0x0FD, the buffer colour stage 1 sees

	texEnable uint32 // 0x080 bits 0-2

	alphaTest bool
	alphaFunc uint8
	alphaRef  int32
}

func (g *GPU) tevstate() tevState {
	var t tevState
	t.texEnable = g.Regs[0x080]

	upd := g.Regs[0x0E0] // buffer-update flags: colour bits 8-11, alpha 12-15
	t.bufColor = rgba{
		int32(g.Regs[0x0FD] & 0xFF), int32(g.Regs[0x0FD] >> 8 & 0xFF),
		int32(g.Regs[0x0FD] >> 16 & 0xFF), int32(g.Regs[0x0FD] >> 24 & 0xFF),
	}

	for s := 0; s < 6; s++ {
		base := tevStageBase[s]
		src := g.Regs[base]
		opd := g.Regs[base+1]
		cmb := g.Regs[base+2]
		scale := g.Regs[base+4]

		st := &t.stages[s]
		for i := 0; i < 3; i++ {
			st.colr[i] = tevOperand{
				src: uint8(src >> (4 * uint(i)) & 0xF),
				op:  uint8(opd >> (4 * uint(i)) & 0xF),
			}
			st.alph[i] = tevOperand{
				src: uint8(src >> (16 + 4*uint(i)) & 0xF),
				op:  uint8(opd >> (12 + 4*uint(i)) & 7),
			}
		}
		st.combC = uint8(cmb & 0xF)
		st.combA = uint8(cmb >> 16 & 0xF)
		st.scaleC = uint8(scale & 3)
		st.scaleA = uint8(scale >> 16 & 3)
		st.konst = rgba{
			int32(g.Regs[base+3] & 0xFF), int32(g.Regs[base+3] >> 8 & 0xFF),
			int32(g.Regs[base+3] >> 16 & 0xFF), int32(g.Regs[base+3] >> 24 & 0xFF),
		}
		// Only the first four stages can write the combiner buffer at all: the update
		// masks are four bits wide, not six.
		if s < 4 {
			st.updC = upd>>(8+uint(s))&1 != 0
			st.updA = upd>>(12+uint(s))&1 != 0
		}
	}

	// Alpha test (0x104): bit0 enable, bits 4-6 func, bits 8-15 reference.
	at := g.Regs[0x104]
	t.alphaTest = at&1 != 0
	t.alphaFunc = uint8(at >> 4 & 7)
	t.alphaRef = int32(at >> 8 & 0xFF)
	return t
}
