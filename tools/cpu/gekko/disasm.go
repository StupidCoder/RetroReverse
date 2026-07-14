package gekko

// disasm.go is the linear disassembler: every word in a range, in order, whether or not
// it is reachable. The recursive-descent one — which follows control flow and so knows
// what is code and what is data — is tools/cmd/codetracegekko, built on the same Decode.

import "fmt"

// Disassemble renders each word of code as one line, starting at base.
func Disassemble(code []byte, base uint32) []string {
	var out []string
	for i := 0; i+4 <= len(code); i += 4 {
		in := Decode(code[i:], base+uint32(i))
		out = append(out, fmt.Sprintf("%08X  %08X  %s", in.Addr, in.Word, in.Text))
	}
	return out
}
