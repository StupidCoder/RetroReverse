package x86

import "fmt"

// rmOperand is a decoded ModR/M r/m operand: either a register (index only,
// widened to the operand size at format time) or a fully-formatted memory
// reference (the address expression is size-independent, so it is built once).
type rmOperand struct {
	isReg bool
	reg   byte   // register index when isReg
	text  string // "[BX+SI+$10]" style memory expression when !isReg
}

// at formats the r/m operand for a given operand bit-width (8/16/32).
func (o rmOperand) at(size int) string {
	if o.isReg {
		return gpr(o.reg, size)
	}
	return o.text
}

// atHint formats the operand, adding an explicit size keyword for memory
// operands (used when no register operand fixes the width, e.g. group ops with
// an immediate, INC/DEC or shifts on memory).
func (o rmOperand) atHint(size int) string {
	if o.isReg {
		return gpr(o.reg, size)
	}
	return sizeKeyword(size) + o.text
}

func sizeKeyword(size int) string {
	switch size {
	case 8:
		return "BYTE "
	case 32:
		return "DWORD "
	default:
		return "WORD "
	}
}

// modrm reads a ModR/M byte (and any SIB + displacement it implies) and returns
// the reg field and the decoded r/m operand.
func (d *dec) modrm() (reg byte, rm rmOperand) {
	b := d.next()
	mod, regf, rmf := b>>6, (b>>3)&7, b&7
	if mod == 3 {
		return regf, rmOperand{isReg: true, reg: rmf}
	}
	if d.addrsize == 32 {
		return regf, rmOperand{text: d.mem32(mod, rmf)}
	}
	return regf, rmOperand{text: d.mem16(mod, rmf)}
}

// mem16 formats a 16-bit-addressing memory operand.
func (d *dec) mem16(mod, rmf byte) string {
	if mod == 0 && rmf == 6 { // [disp16] direct
		return fmt.Sprintf("[%s%s]", d.seg, hexImm(uint32(d.imm16()), 2))
	}
	base := rm16[rmf]
	switch mod {
	case 1:
		return fmt.Sprintf("[%s%s%s]", d.seg, base, dispStr(int32(int8(d.next()))))
	case 2:
		return fmt.Sprintf("[%s%s%s]", d.seg, base, dispStr(int32(int16(d.imm16()))))
	default:
		return fmt.Sprintf("[%s%s]", d.seg, base)
	}
}

// mem32 formats a 32-bit-addressing memory operand (386, under a 0x67 prefix),
// including SIB decoding.
func (d *dec) mem32(mod, rmf byte) string {
	var base string
	if rmf == 4 { // SIB byte follows
		base = d.sib(mod)
	} else if mod == 0 && rmf == 5 { // [disp32] direct
		return fmt.Sprintf("[%s%s]", d.seg, hexImm(d.imm32(), 4))
	} else {
		base = reg32[rmf]
	}
	switch mod {
	case 1:
		return fmt.Sprintf("[%s%s%s]", d.seg, base, dispStr(int32(int8(d.next()))))
	case 2:
		return fmt.Sprintf("[%s%s%s]", d.seg, base, dispStr(int32(d.imm32())))
	default:
		return fmt.Sprintf("[%s%s]", d.seg, base)
	}
}

// sib decodes a SIB byte into a "base+index*scale" expression (without the
// displacement, which the caller appends). A mod==0 SIB with base==5 encodes a
// disp32 base, handled here by emitting the disp32 in place of the base name.
func (d *dec) sib(mod byte) string {
	s := d.next()
	scale, index, base := s>>6, (s>>3)&7, s&7
	idx := ""
	if index != 4 { // index==4 means "no index register"
		idx = "+" + reg32[index] + fmt.Sprintf("*%d", 1<<scale)
	}
	if base == 5 && mod == 0 {
		return hexImm(d.imm32(), 4) + idx
	}
	return reg32[base] + idx
}

// dispStr formats a signed displacement as "+$1A" or "-$4" (empty when zero is
// preferred by the caller? — here always shown for clarity except exact 0).
func dispStr(v int32) string {
	switch {
	case v == 0:
		return ""
	case v < 0:
		return fmt.Sprintf("-$%X", -v)
	default:
		return fmt.Sprintf("+$%X", v)
	}
}

// Disassemble renders a code blob to one text line per instruction, formatted
// "addr:  bytes  text", starting at linear address base.
func Disassemble(code []byte, base uint32) []string {
	var out []string
	for off := 0; off < len(code); {
		in := Decode(code[off:], base+uint32(off))
		raw := code[off : off+in.Len]
		out = append(out, fmt.Sprintf("%08X  %-18s  %s", in.Addr, hexBytes(raw), in.Text))
		off += in.Len
	}
	return out
}

// Disassemble32 is Disassemble with a 32-bit default operand/address size, for
// flat protected-mode / go32 code (the CS descriptor's D bit is 1).
func Disassemble32(code []byte, base uint32) []string {
	var out []string
	for off := 0; off < len(code); {
		in := Decode32(code[off:], base+uint32(off))
		if in.Len <= 0 {
			break
		}
		raw := code[off : off+in.Len]
		out = append(out, fmt.Sprintf("%08X  %-18s  %s", in.Addr, hexBytes(raw), in.Text))
		off += in.Len
	}
	return out
}

func hexBytes(b []byte) string {
	s := ""
	for i, x := range b {
		if i == 6 { // keep the column tidy for long instructions
			s += "…"
			break
		}
		s += fmt.Sprintf("%02X ", x)
	}
	return s
}
