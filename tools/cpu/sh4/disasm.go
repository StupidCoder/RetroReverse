package sh4

import "fmt"

// Disassemble renders code as a linear listing, one instruction per 2 bytes,
// starting at CPU address base. It makes no attempt to separate code from the
// literal pools woven between functions — that is the recursive-descent
// tracer's job (tools/cmd/codetracesh4).
func Disassemble(code []byte, base uint32) []string {
	var out []string
	for i := 0; i+2 <= len(code); i += 2 {
		addr := base + uint32(i)
		in := Decode(code[i:], addr)
		out = append(out, fmt.Sprintf("%08X  %04X  %s",
			addr, uint16(code[i])|uint16(code[i+1])<<8, in.Text))
	}
	return out
}
