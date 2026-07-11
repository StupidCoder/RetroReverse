package n3ds

// gpu_shader_disasm.go renders uploaded PICA200 shader code as text — the
// verification instrument for the VM's instruction decode: if the field
// layouts were wrong, the game's own shader disassembles to gibberish
// (writes to impossible registers, no END, nonsense swizzles) instead of a
// coherent transform program.

import (
	"fmt"
	"strings"
)

var shaderOpNames = map[uint32]string{
	0x00: "add", 0x01: "dp3", 0x02: "dp4", 0x03: "dph", 0x04: "dst",
	0x05: "ex2", 0x06: "lg2", 0x07: "litp", 0x08: "mul", 0x09: "sge",
	0x0A: "slt", 0x0B: "flr", 0x0C: "max", 0x0D: "min", 0x0E: "rcp",
	0x0F: "rsq", 0x12: "mova", 0x13: "mov", 0x18: "dphi", 0x19: "dsti",
	0x1A: "sgei", 0x1B: "slti", 0x21: "nop", 0x22: "end", 0x23: "breakc",
	0x24: "call", 0x25: "callc", 0x26: "callu", 0x27: "ifu", 0x28: "ifc",
	0x29: "loop", 0x2C: "jmpc", 0x2D: "jmpu",
}

func srcName(reg int) string {
	switch {
	case reg < 0x10:
		return fmt.Sprintf("v%d", reg)
	case reg < 0x20:
		return fmt.Sprintf("r%d", reg-0x10)
	default:
		return fmt.Sprintf("c%d", reg-0x20)
	}
}

func dstName(reg int) string {
	if reg < 0x10 {
		return fmt.Sprintf("o%d", reg)
	}
	return fmt.Sprintf("r%d", reg-0x10)
}

func swizzleText(desc uint32, n int) string {
	shift := []uint{4, 13, 22}[n]
	neg := ""
	if desc>>shift&1 != 0 {
		neg = "-"
	}
	sw := desc >> (shift + 1) & 0xFF
	comps := "xyzw"
	var b strings.Builder
	for i := uint(0); i < 4; i++ {
		b.WriteByte(comps[sw>>(6-2*i)&3])
	}
	t := b.String()
	if t == "xyzw" {
		t = ""
	} else {
		t = "." + t
	}
	return neg + t
}

func maskText(desc uint32) string {
	m := desc & 0xF
	if m == 0xF {
		return ""
	}
	comps := "xyzw"
	var b strings.Builder
	b.WriteByte('.')
	for i := uint(0); i < 4; i++ {
		if m>>(3-i)&1 != 0 {
			b.WriteByte(comps[i])
		}
	}
	return b.String()
}

// ShaderDisasm renders one instruction. opdesc is the descriptor table in use.
func ShaderDisasm(in uint32, opdesc *[128]uint32) string {
	op := in >> 26
	name := shaderOpNames[op]

	if op >= 0x30 { // MAD / MADI
		desc := opdesc[in&0x1F]
		dst := dstName(int(in >> 24 & 0x1F))
		s1 := srcName(int(in >> 17 & 0x1F))
		var s2, s3 string
		if op >= 0x38 {
			s2, s3 = srcName(int(in>>10&0x7F)), srcName(int(in>>5&0x1F))
		} else {
			s2, s3 = srcName(int(in>>12&0x1F)), srcName(int(in>>5&0x7F))
		}
		return fmt.Sprintf("mad  %s%s, %s%s, %s%s, %s%s", dst, maskText(desc),
			s1, swizzleText(desc, 0), s2, swizzleText(desc, 1), s3, swizzleText(desc, 2))
	}
	if op>>1 == 0x17 { // CMP
		desc := opdesc[in&0x7F]
		conds := []string{"eq", "ne", "lt", "le", "gt", "ge", "t", "t"}
		return fmt.Sprintf("cmp  %s%s %s|%s %s%s", srcName(int(in>>12&0x7F)), swizzleText(desc, 0),
			conds[in>>24&7], conds[in>>21&7], srcName(int(in>>7&0x1F)), swizzleText(desc, 1))
	}

	switch op {
	case 0x21, 0x22:
		return name
	case 0x24, 0x25, 0x26, 0x27, 0x28, 0x2C, 0x2D:
		return fmt.Sprintf("%-4s dst=%d num=%d sel=%d", name, in>>10&0xFFF, in&0xFF, in>>22&0xF)
	case 0x29:
		return fmt.Sprintf("loop i%d dst=%d", in>>22&3, in>>10&0xFFF)
	}

	if name == "" {
		return fmt.Sprintf(".word 0x%08X (op 0x%02X?)", in, op)
	}
	desc := opdesc[in&0x7F]
	dst := dstName(int(in >> 21 & 0x1F))
	switch op {
	case 0x18, 0x19, 0x1A, 0x1B: // inverted operand widths
		return fmt.Sprintf("%-4s %s%s, %s%s, %s%s", name, dst, maskText(desc),
			srcName(int(in>>14&0x1F)), swizzleText(desc, 0),
			srcName(int(in>>7&0x7F)), swizzleText(desc, 1))
	case 0x05, 0x06, 0x0B, 0x0E, 0x0F, 0x12, 0x13: // single-source
		return fmt.Sprintf("%-4s %s%s, %s%s", name, dst, maskText(desc),
			srcName(int(in>>12&0x7F)), swizzleText(desc, 0))
	}
	return fmt.Sprintf("%-4s %s%s, %s%s, %s%s", name, dst, maskText(desc),
		srcName(int(in>>12&0x7F)), swizzleText(desc, 0),
		srcName(int(in>>7&0x1F)), swizzleText(desc, 1))
}
