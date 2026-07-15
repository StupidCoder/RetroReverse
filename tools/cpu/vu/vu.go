package vu

// Package vu is the PS2's Vector Unit — for this repository's purposes VU1, the
// programmable half of the render path: 16 KiB of program memory, 16 KiB of data
// memory, 32 four-float registers, 16 integer registers, and an XGKICK instruction
// that hands a finished GIF packet in data memory to the Graphics Synthesizer (PATH1).
//
// vu.go holds the machine state; exec.go runs it; disasm.go reads it.

import "math"

func float32frombits(b uint32) float32 { return math.Float32frombits(b) }
func float32bits(f float32) uint32     { return math.Float32bits(f) }
