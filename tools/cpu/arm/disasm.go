package arm

import (
	"fmt"
	"strings"
)

// Disassemble renders a code blob loaded at base, one line per instruction, in
// "ADDR: BYTES  MNEMONIC operands" form (hex). It decodes in Thumb state when thumb
// is true and ARM state otherwise, at that fixed width throughout — a linear driver
// cannot follow a BX that switches state at run time, so callers disassemble each
// state's regions separately (the codetracearm tool tracks state changes for you).
// Unmodelled words and truncated reads appear as ".word"/".hword". It decodes
// for the ARMv5TE (Nintendo DS) variant; use DisassembleVariant for ARMv6K.
func Disassemble(code []byte, base uint32, thumb bool) []string {
	return DisassembleVariant(code, base, thumb, V5TE)
}

// DisassembleVariant is Disassemble for a chosen architecture variant — V6K for
// the Nintendo 3DS's ARM11, V5TE for the DS.
func DisassembleVariant(code []byte, base uint32, thumb bool, v Variant) []string {
	var out []string
	for pc := 0; pc < len(code); {
		in := DecodeVariant(code[pc:], base+uint32(pc), thumb, v)
		raw := make([]string, 0, in.Len)
		for i := 0; i < in.Len && pc+i < len(code); i++ {
			raw = append(raw, fmt.Sprintf("%02X", code[pc+i]))
		}
		out = append(out, fmt.Sprintf("$%08X: %-12s %s", base+uint32(pc), strings.Join(raw, " "), in.Text))
		if in.Len <= 0 {
			break
		}
		pc += in.Len
	}
	return out
}
