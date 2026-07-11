package allegrex

import (
	"fmt"
	"strings"
)

// Disassemble renders a code blob loaded at base, one line per 4-byte
// instruction, in "ADDR: BYTES  MNEMONIC operands" form (hex). It is a linear
// driver: it decodes straight through and does not follow control flow — that is
// the job of the codetracemips tool, which also honours delay slots and resolves
// jump tables. Unmodelled words appear as ".word".
func Disassemble(code []byte, base uint32) []string {
	var out []string
	for pc := 0; pc+4 <= len(code); pc += 4 {
		in := Decode(code[pc:], base+uint32(pc))
		raw := fmt.Sprintf("%02X %02X %02X %02X", code[pc], code[pc+1], code[pc+2], code[pc+3])
		out = append(out, fmt.Sprintf("$%08X: %s  %s", base+uint32(pc), raw, strings.TrimSpace(in.Text)))
	}
	return out
}
