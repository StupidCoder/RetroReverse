package r4300

import "fmt"

// Disassemble renders code as a linear listing, one instruction per 4 bytes,
// starting at CPU address base. It makes no attempt to separate code from data —
// that is the recursive-descent tracer's job (tools/cmd/codetracer4300).
func Disassemble(code []byte, base uint32) []string {
	var out []string
	for i := 0; i+4 <= len(code); i += 4 {
		addr := base + uint32(i)
		in := Decode(code[i:], addr)
		out = append(out, fmt.Sprintf("%08X  %02X %02X %02X %02X  %s",
			addr, code[i], code[i+1], code[i+2], code[i+3], in.Text))
	}
	return out
}
