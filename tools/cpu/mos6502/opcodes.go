package mos6502

// Opcode describes the official instruction set shared with the browser decoder.
// Undocumented opcodes are intentionally absent, never assigned guessed lengths.
type Opcode struct {
	Mnemonic string `json:"mnemonic"`
	Mode     string `json:"mode"`
	Length   int    `json:"length"`
}

func Opcodes() map[byte]Opcode {
	names := [...]string{"imp", "acc", "imm", "zp", "zpx", "zpy", "izx", "izy", "rel", "abs", "abx", "aby", "ind"}
	out := map[byte]Opcode{}
	for code, op := range ops {
		out[code] = Opcode{op.mn, names[op.m], modeLen[op.m]}
	}
	return out
}
