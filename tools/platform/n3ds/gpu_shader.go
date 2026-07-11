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
func (g *GPU) shaderRun(v *[16][4]float32) (o [16][4]float32, ok bool) {
	s := shaderState{g: g, v: v, o: &o}
	entry := g.Regs[regVshEntry] & 0xFFF
	ok = s.exec(int(entry), len(g.Code))
	return o, ok
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
		in := g.Code[pc]
		op := in >> 26
		switch op {
		case 0x21: // NOP
			pc++
			continue
		case 0x22: // END
			return true
		}

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
			if !s.arith(in, op) {
				return false
			}
			pc++
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

// arith executes one arithmetic-format instruction.
func (s *shaderState) arith(in, op uint32) bool {
	g := s.g

	// MAD/MADI occupy the whole 0x30-0x3F opcode space (the descriptor field
	// shrinks to 5 bits to fit three sources).
	if op >= 0x30 {
		desc := g.Opdesc[in&0x1F]
		dst := int(in >> 24 & 0x1F)
		s1 := s.readSrc(int(in>>17&0x1F), 0, desc, 0)
		var s2, s3 [4]float32
		if op >= 0x38 { // MAD: wide src2, idx applies to it
			s2 = s.readSrc(int(in>>10&0x7F), int(in>>22&3), desc, 1)
			s3 = s.readSrc(int(in>>5&0x1F), 0, desc, 2)
		} else { // MADI: wide src3
			s2 = s.readSrc(int(in>>12&0x1F), 0, desc, 1)
			s3 = s.readSrc(int(in>>5&0x7F), int(in>>22&3), desc, 2)
		}
		var out [4]float32
		for i := 0; i < 4; i++ {
			out[i] = s1[i]*s2[i] + s3[i]
		}
		s.writeDst(dst, desc, out)
		return true
	}

	// CMP spills its first comparison operator into the opcode's low bit.
	if op>>1 == 0x17 {
		desc := g.Opdesc[in&0x7F]
		s1 := s.readSrc(int(in>>12&0x7F), int(in>>19&3), desc, 0)
		s2 := s.readSrc(int(in>>7&0x1F), 0, desc, 1)
		s.cc[0] = compare(in>>24&7, s1[0], s2[0])
		s.cc[1] = compare(in>>21&7, s1[1], s2[1])
		return true
	}

	desc := g.Opdesc[in&0x7F]
	dst := int(in >> 21 & 0x1F)
	idx := int(in >> 19 & 3)

	// The "inverted" variants swap which source gets the wide (constant-capable)
	// field.
	var s1, s2 [4]float32
	switch op {
	case 0x18, 0x19, 0x1A, 0x1B: // DPHI, DSTI, SGEI, SLTI
		s1 = s.readSrc(int(in>>14&0x1F), 0, desc, 0)
		s2 = s.readSrc(int(in>>7&0x7F), idx, desc, 1)
	default:
		s1 = s.readSrc(int(in>>12&0x7F), idx, desc, 0)
		s2 = s.readSrc(int(in>>7&0x1F), 0, desc, 1)
	}

	var out [4]float32
	switch op {
	case 0x00: // ADD
		for i := range out {
			out[i] = s1[i] + s2[i]
		}
	case 0x01: // DP3
		d := s1[0]*s2[0] + s1[1]*s2[1] + s1[2]*s2[2]
		out = [4]float32{d, d, d, d}
	case 0x02: // DP4
		d := s1[0]*s2[0] + s1[1]*s2[1] + s1[2]*s2[2] + s1[3]*s2[3]
		out = [4]float32{d, d, d, d}
	case 0x03, 0x18: // DPH / DPHI: dp4 with src1.w forced to 1
		d := s1[0]*s2[0] + s1[1]*s2[1] + s1[2]*s2[2] + s2[3]
		out = [4]float32{d, d, d, d}
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
	case 0x12: // MOVA: latch a0 from src1.xy per the dest mask
		if desc>>3&1 != 0 {
			s.a0[0] = int32(s1[0])
		}
		if desc>>2&1 != 0 {
			s.a0[1] = int32(s1[1])
		}
		return true
	case 0x13: // MOV
		out = s1
	default:
		g.m.CPU.Halt("gpu shader: opcode 0x%02X unimplemented (instr 0x%08X)", op, in)
		return false
	}
	s.writeDst(dst, desc, out)
	return true
}

// readSrc fetches a source register through the descriptor's swizzle/negate
// for operand slot n (0-2). Register space: 0x00-0x0F inputs, 0x10-0x1F
// temporaries, 0x20-0x7F float uniforms; relative addressing (idx) applies to
// uniform reads only.
func (s *shaderState) readSrc(reg, idx int, desc uint32, n int) [4]float32 {
	var base [4]float32
	switch {
	case reg < 0x10:
		base = s.v[reg]
	case reg < 0x20:
		base = s.r[reg-0x10]
	default:
		c := reg - 0x20
		switch idx {
		case 1:
			c += int(s.a0[0])
		case 2:
			c += int(s.a0[1])
		case 3:
			c += int(s.aL)
		}
		if c >= 0 && c < len(s.g.Float) {
			base = s.g.Float[c]
		}
	}
	// Descriptor layout per operand: negate bit then an 8-bit swizzle (2 bits
	// per destination component, x first in the high bits).
	shift := []uint{4, 13, 22}[n]
	neg := desc>>shift&1 != 0
	sw := desc >> (shift + 1) & 0xFF
	var out [4]float32
	for i := uint(0); i < 4; i++ {
		out[i] = base[sw>>(6-2*i)&3]
	}
	if neg {
		for i := range out {
			out[i] = -out[i]
		}
	}
	return out
}

// writeDst stores to an output (0x00-0x0F) or temporary (0x10-0x1F) register
// under the descriptor's component mask (bit 3 = x).
func (s *shaderState) writeDst(reg int, desc uint32, val [4]float32) {
	var t *[4]float32
	if reg < 0x10 {
		t = &s.o[reg]
	} else {
		t = &s.r[reg-0x10]
	}
	for i := uint(0); i < 4; i++ {
		if desc>>(3-i)&1 != 0 {
			t[i] = val[i]
		}
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
