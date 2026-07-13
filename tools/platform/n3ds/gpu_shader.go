package n3ds

// gpu_shader.go is the LLE PICA200 vertex-shader interpreter. The shader
// binary is the game's own data, uploaded through the code and operand-
// descriptor FIFOs; running it instruction by instruction is the clean-room
// equivalent of LLE'ing the N64's RSP microcode — no assumption about what
// transform the game "probably" does.
//
// The machine: 16 input registers v0-15, 16 temporaries r0-15, 96 float
// uniform vec4s c0-95, up to 16 outputs o0-15, 16 booleans, 4 integer vec4s,
// two address registers a0.x/a0.y and the loop counter aL. Instructions are
// one word; the top 6 bits are the opcode. Most arithmetic ops name a
// destination + source registers plus an operand descriptor (a swizzle/negate
// pattern from the descriptor table). Unknown opcodes halt — the honest
// bring-up posture; opcodes get implemented as the game's shaders demand.

import (
	"fmt"
	"math"
)

// shaderRun executes the uploaded program from the entry point for one vertex.
// v holds the input attributes; the return is the 16 output registers.
// shaderRun runs the vertex shader for one vertex: v is the input register file,
// out the output register file, entry the program's entry point (draw-constant, so
// the caller hoists it out of the vertex loop).
//
// out is caller-owned and reused across the vertices of a draw, so it is cleared
// here: an output register the program never writes must read as zero, which it
// did for free when this returned a fresh array by value — and returning that
// array by value cost 6% of the frame in runtime.duffcopy.
func (g *GPU) shaderRun(v, out *[16][4]float32, entry int) bool {
	*out = [16][4]float32{}
	s := shaderState{g: g, v: v, o: out}
	return s.exec(entry, len(g.Code))
}

type shaderState struct {
	g  *GPU
	v  *[16][4]float32
	o  *[16][4]float32
	r  [16][4]float32
	a0 [2]int32 // a0.x, a0.y address registers
	aL int32    // loop counter
	cc [2]bool  // the condition register CMP writes and IFC/JMPC read
}

// exec runs instructions in [pc, end) until END; it returns false (after
// halting the machine) on anything unimplemented. Structured flow (IF/CALL/
// LOOP) recurses on sub-ranges, which matches the PICA's block-shaped flow
// instructions.
func (s *shaderState) exec(pc, end int) bool {
	g := s.g
	for steps := 0; pc < end; steps++ {
		if steps > 1<<20 {
			g.m.CPU.Halt("gpu shader: runaway program (no END after 1M steps)")
			return false
		}
		d := g.shaderInst(pc)
		switch d.kind {
		case shNop:
			pc++
			continue
		case shEnd:
			return true
		case shArith, shMad, shCmp:
			if !s.arith(d) {
				return false
			}
			pc++
			continue
		}

		// Flow control: rare, and its fields are cheap bit extracts, so it is
		// decoded here rather than cached.
		in := g.Code[pc]
		op := in >> 26
		switch op {
		case 0x24: // CALL num@dst
			dst, num := int(in>>10&0xFFF), int(in&0xFF)
			if !s.exec(dst, dst+num) {
				return false
			}
			pc++
		case 0x25, 0x26: // CALLC / CALLU
			dst, num := int(in>>10&0xFFF), int(in&0xFF)
			taken := false
			if op == 0x26 {
				taken = s.boolReg(in)
			} else {
				taken = s.cond(in)
			}
			if taken {
				if !s.exec(dst, dst+num) {
					return false
				}
			}
			pc++
		case 0x27, 0x28: // IFU / IFC: if-block [pc+1,dst), else-block [dst,dst+num)
			dst, num := int(in>>10&0xFFF), int(in&0xFF)
			taken := false
			if op == 0x27 {
				taken = s.boolReg(in)
			} else {
				taken = s.cond(in)
			}
			if taken {
				if !s.exec(pc+1, dst) {
					return false
				}
			} else if num > 0 {
				if !s.exec(dst, dst+num) {
					return false
				}
			}
			pc = dst + num
		case 0x29: // LOOP intreg: block [pc+1, dst] inclusive, aL walks int.y by int.z
			dst := int(in >> 10 & 0xFFF)
			ir := g.Int[in>>22&3]
			s.aL = int32(ir[1])
			for n := 0; n <= int(ir[0]); n++ {
				if !s.exec(pc+1, dst+1) {
					return false
				}
				s.aL += int32(ir[2])
			}
			pc = dst + 1
		case 0x2C, 0x2D: // JMPC / JMPU
			dst := int(in >> 10 & 0xFFF)
			taken := false
			if op == 0x2D {
				taken = s.boolReg(in)
				if in&1 != 0 { // low num bit inverts a JMPU
					taken = !taken
				}
			} else {
				taken = s.cond(in)
			}
			if taken {
				pc = dst
			} else {
				pc++
			}
		default:
			// shaderInst classed this as neither flow nor arithmetic: an opcode the
			// interpreter does not implement. Halt, loudly, as it always has.
			g.m.CPU.Halt("gpu shader: opcode 0x%02X unimplemented (instr 0x%08X)", op, in)
			return false
		}
	}
	return true
}

// cond evaluates an IFC/JMPC/CALLC condition against the cc flags set by CMP.
func (s *shaderState) cond(in uint32) bool {
	refx := in>>25&1 != 0
	refy := in>>24&1 != 0
	x := s.cc[0] == refx
	y := s.cc[1] == refy
	switch in >> 22 & 3 {
	case 0:
		return x || y
	case 1:
		return x && y
	case 2:
		return x
	default:
		return y
	}
}

func (s *shaderState) boolReg(in uint32) bool {
	return s.g.Bool>>(in>>22&0xF)&1 != 0
}

// arith executes one arithmetic-format instruction from its decoded form
// (gpu_shader_cache.go). The arithmetic is unchanged; what is gone is the
// re-derivation of the operands from the instruction word on every vertex.
func (s *shaderState) arith(d *shInst) bool {
	op := uint32(d.op)

	switch d.kind {
	case shMad:
		s1, s2, s3 := s.src(&d.src[0]), s.src(&d.src[1]), s.src(&d.src[2])
		var out [4]float32
		for i := 0; i < 4; i++ {
			out[i] = s1[i]*s2[i] + s3[i]
		}
		s.writeDst(d, out)
		return true

	case shCmp:
		s1, s2 := s.src(&d.src[0]), s.src(&d.src[1])
		s.cc[0] = compare(uint32(d.cmpX), s1[0], s2[0])
		s.cc[1] = compare(uint32(d.cmpY), s1[1], s2[1])
		return true
	}

	s1, s2 := s.src(&d.src[0]), s.src(&d.src[1])

	var out [4]float32
	switch op {
	case 0x00: // ADD
		for i := range out {
			out[i] = s1[i] + s2[i]
		}
	case 0x01: // DP3
		dp := s1[0]*s2[0] + s1[1]*s2[1] + s1[2]*s2[2]
		out = [4]float32{dp, dp, dp, dp}
	case 0x02: // DP4
		dp := s1[0]*s2[0] + s1[1]*s2[1] + s1[2]*s2[2] + s1[3]*s2[3]
		out = [4]float32{dp, dp, dp, dp}
	case 0x03, 0x18: // DPH / DPHI: dp4 with src1.w forced to 1
		dp := s1[0]*s2[0] + s1[1]*s2[1] + s1[2]*s2[2] + s2[3]
		out = [4]float32{dp, dp, dp, dp}
	case 0x08: // MUL
		for i := range out {
			out[i] = s1[i] * s2[i]
		}
	case 0x09, 0x1A: // SGE / SGEI
		for i := range out {
			if s1[i] >= s2[i] {
				out[i] = 1
			}
		}
	case 0x0A, 0x1B: // SLT / SLTI
		for i := range out {
			if s1[i] < s2[i] {
				out[i] = 1
			}
		}
	case 0x0B: // FLR
		for i := range out {
			out[i] = floor32(s1[i])
		}
	case 0x0C: // MAX
		for i := range out {
			if s1[i] > s2[i] {
				out[i] = s1[i]
			} else {
				out[i] = s2[i]
			}
		}
	case 0x0D: // MIN
		for i := range out {
			if s1[i] < s2[i] {
				out[i] = s1[i]
			} else {
				out[i] = s2[i]
			}
		}
	case 0x0E: // RCP
		d := float32(1) / s1[0]
		out = [4]float32{d, d, d, d}
	case 0x0F: // RSQ
		d := rsqrt32(s1[0])
		out = [4]float32{d, d, d, d}
	case 0x12: // MOVA: latch a0 from src1.xy per the destination mask
		if d.mask[0] {
			s.a0[0] = int32(s1[0])
		}
		if d.mask[1] {
			s.a0[1] = int32(s1[1])
		}
		return true
	case 0x13: // MOV
		out = s1
	default:
		s.g.m.CPU.Halt("gpu shader: opcode 0x%02X unimplemented", op)
		return false
	}
	s.writeDst(d, out)
	return true
}

// src fetches a decoded source operand: pick the register file, apply relative
// addressing (uniforms only), swizzle, negate.
func (s *shaderState) src(o *shSrc) [4]float32 {
	var base *[4]float32
	switch o.bank {
	case shBankIn:
		base = &s.v[o.reg]
	case shBankTmp:
		base = &s.r[o.reg]
	default:
		c := int(o.reg)
		switch o.idx {
		case 1:
			c += int(s.a0[0])
		case 2:
			c += int(s.a0[1])
		case 3:
			c += int(s.aL)
		}
		if c < 0 || c >= len(s.g.Float) {
			return [4]float32{} // out of range reads as zero, as it always has
		}
		base = &s.g.Float[c]
	}
	if o.plain { // the identity swizzle, unnegated: the register itself
		return *base
	}
	out := [4]float32{base[o.sw[0]], base[o.sw[1]], base[o.sw[2]], base[o.sw[3]]}
	if o.neg {
		out[0], out[1], out[2], out[3] = -out[0], -out[1], -out[2], -out[3]
	}
	return out
}

// writeDst stores to the decoded destination register under its component mask.
func (s *shaderState) writeDst(d *shInst, val [4]float32) {
	var t *[4]float32
	if d.dstTmp {
		t = &s.r[d.dst]
	} else {
		t = &s.o[d.dst]
	}
	if d.maskAll { // the common case: a whole-vector store
		*t = val
		return
	}
	if d.mask[0] {
		t[0] = val[0]
	}
	if d.mask[1] {
		t[1] = val[1]
	}
	if d.mask[2] {
		t[2] = val[2]
	}
	if d.mask[3] {
		t[3] = val[3]
	}
}

func compare(op uint32, a, b float32) bool {
	switch op {
	case 0:
		return a == b
	case 1:
		return a != b
	case 2:
		return a < b
	case 3:
		return a <= b
	case 4:
		return a > b
	case 5:
		return a >= b
	default:
		return true
	}
}

func floor32(f float32) float32 {
	i := float32(int32(f))
	if f < 0 && f != i {
		i--
	}
	return i
}

func rsqrt32(f float32) float32 {
	if f <= 0 {
		return float32(math.Inf(1))
	}
	return float32(1 / math.Sqrt(float64(f)))
}

var _ = fmt.Sprintf
