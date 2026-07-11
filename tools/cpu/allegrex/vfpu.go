package allegrex

// vfpu.go is the Allegrex COP2 vector unit (the VFPU). The VFPU occupies the COP2
// opcode (op 0x12, register moves + the vpfx/command stream) plus a family of primary
// opcodes for its loads/stores (op 0x32/0x36/… lv/sv) and compute groups (op 0x18/
// 0x19/… VFPU0..7). Its 128 registers are addressed as 8 4×4 matrices; a 7-bit single
// register number S[m][r][c] flattens to m*16 + r*4 + c in CPU.V.
//
// Phase-1 scope (per the package doc): the whole VFPU is *decoded* so the disassembler
// and tracer stay legible, but only the single-element loads/stores (lv.s/sv.s) are
// *executed*; the quad/matrix loads and the arithmetic groups Halt with their word so
// gaps are explicit rather than silently wrong. Execution grows from those Halts.

import "fmt"

// vfpuSingle returns the flat CPU.V index of a 7-bit VFPU single register number.
func vfpuSingle(reg uint32) uint32 {
	m := (reg >> 2) & 7
	c := reg & 3
	r := (reg >> 5) & 3
	return m*16 + r*4 + c
}

// decodeCop2 disassembles a COP2 (VFPU) register move or command.
func decodeCop2(in Inst, w, rs, rt, rd uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	if w&(1<<25) != 0 { // VFPU command / prefix stream
		return set("vcop2", fmt.Sprintf("vcop2 0x%07X", w&0x01FFFFFF))
	}
	switch rs {
	case 0x03: // mfv / mfvc — VFPU register to GPR
		return set("mfv", fmt.Sprintf("mfv %s, $v%d", reg(rt), w&0xFF))
	case 0x07: // mtv / mtvc — GPR to VFPU register
		return set("mtv", fmt.Sprintf("mtv %s, $v%d", reg(rt), w&0xFF))
	}
	return set("cop2", fmt.Sprintf("cop2 0x%08X", w))
}

// decodeVFPU disassembles the VFPU load/store and compute primary opcodes. Loads and
// stores are rendered precisely; the compute groups get a generic mnemonic (with the
// raw word) so a trace continues through SDK math without stopping on .word.
func decodeVFPU(in Inst, w, op, rs, rt, simm uint32) Inst {
	set := func(mnem, text string) Inst { in.Mnem, in.Text = mnem, text; return in }
	vt := rt | (w&3)<<5 // 7-bit VFPU register (low 2 bits stolen from the offset)
	off := int32(simm) &^ 3
	switch op {
	case 0x32:
		return set("lv.s", fmt.Sprintf("lv.s $v%d, %d(%s)", vt, off, reg(rs)))
	case 0x3A:
		return set("sv.s", fmt.Sprintf("sv.s $v%d, %d(%s)", vt, off, reg(rs)))
	case 0x36:
		return set("lv.q", fmt.Sprintf("lv.q $v%d, %d(%s)", vt, off, reg(rs)))
	case 0x3E:
		return set("sv.q", fmt.Sprintf("sv.q $v%d, %d(%s)", vt, off, reg(rs)))
	case 0x35:
		return set("lvl.q", fmt.Sprintf("lv%s.q $v%d, %d(%s)", lr(w), vt, off, reg(rs)))
	case 0x3D:
		return set("svl.q", fmt.Sprintf("sv%s.q $v%d, %d(%s)", lr(w), vt, off, reg(rs)))
	}
	// Compute groups (VFPU0/1/3/4/5/6/7): decoded generically for legibility.
	return set("vfpu", fmt.Sprintf("vfpu.%02X 0x%08X", op, w))
}

func lr(w uint32) string {
	if w&2 != 0 {
		return "r"
	}
	return "l"
}

// --- execution -------------------------------------------------------------

// cop2 executes a COP2 (VFPU) register move.
func (c *CPU) cop2(w, rs, rt, rd uint32) {
	switch rs {
	case 0x03: // mfv (delayed like a load)
		c.load(rt, c.V[vfpuSingle(w&0xFF)&127])
	case 0x07: // mtv
		c.V[vfpuSingle(w&0xFF)&127] = c.reg(rt)
	default:
		c.Halt("unimplemented cop2 rs=0x%02X (word 0x%08X) at 0x%08X", rs, w, c.curPC)
	}
}

// vfpuOp executes the VFPU load/store and compute opcodes. Only the single loads and
// stores run; everything else Halts.
func (c *CPU) vfpuOp(w, op, rs, rt, simm uint32) {
	vt := rt | (w&3)<<5
	addr := c.reg(rs) + (simm &^ 3)
	switch op {
	case 0x32: // lv.s
		if addr&3 != 0 {
			c.addrError(excAdEL, addr)
			return
		}
		c.V[vfpuSingle(vt)&127] = c.read32(addr)
	case 0x3A: // sv.s
		if addr&3 != 0 {
			c.addrError(excAdES, addr)
			return
		}
		c.write32(addr, c.V[vfpuSingle(vt)&127])
	default:
		if c.OnVFPU != nil {
			c.OnVFPU(w, op)
			return
		}
		c.Halt("unimplemented VFPU op 0x%02X (word 0x%08X) at 0x%08X", op, w, c.curPC)
	}
}
