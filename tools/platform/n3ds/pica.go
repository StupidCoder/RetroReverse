package n3ds

// pica.go decodes a PICA200 command buffer into its register-write stream. A
// command list — what the game submits via GX ProcessCommandList — is not
// drawing commands but a sequence of GPU register writes, structurally the same
// idea as the N64 RDP command list (tools/platform/n64/rdp.go): the CPU-side
// driver encodes "set these registers, then touch the draw-trigger register"
// and the GPU's command processor replays it.
//
// Each entry is two words, 8-byte aligned:
//
//	word0: the parameter (the value to write)
//	word1: the header — bits[15:0] register id, bits[19:16] byte-enable mask,
//	       bits[27:20] extra-parameter count N, bit[31] consecutive flag
//
// N extra parameter words follow (plus one pad word when N is odd, keeping the
// next entry 8-byte aligned). Without the consecutive flag every extra parameter
// re-writes the same register (a burst into a FIFO register, e.g. shader-code
// upload); with it, parameter i writes register id+i (a run of adjacent
// registers, e.g. a matrix into consecutive uniform words).
//
// This decoder is the instrument-first half of the Phase 4 GPU: it lets us *see*
// the register traffic the game generates before any of it is executed. The GPU
// interpreter (gpu.go) consumes the same stream.

import "fmt"

// PICAWrite is one decoded register write.
type PICAWrite struct {
	Off   uint32 // byte offset of the entry this write came from
	Reg   uint16 // destination register id
	Mask  uint8  // byte-enable mask (bit i = byte i of the register is written)
	Value uint32
	Burst bool // part of a multi-parameter entry
}

// DecodePICA walks a command buffer and returns every register write, in order.
// A malformed entry (running past the buffer) stops the walk with an error and
// the writes decoded so far — deliberately loud, never a silent truncation.
func DecodePICA(buf []byte) ([]PICAWrite, error) {
	return DecodePICAInto(nil, buf)
}

// DecodePICAInto is DecodePICA appending into a caller-owned slice, so the
// command processor can keep one buffer and reuse it. Captain Toad decodes ~2,500
// command lists a frame; allocating each one's write stream afresh made the Go
// allocator a tenth of the frame's cost.
//
// Call it as `ws, err = DecodePICAInto(ws[:0], buf)`.
func DecodePICAInto(dst []PICAWrite, buf []byte) ([]PICAWrite, error) {
	word := func(off uint32) uint32 {
		return uint32(buf[off]) | uint32(buf[off+1])<<8 | uint32(buf[off+2])<<16 | uint32(buf[off+3])<<24
	}
	ws := dst
	off := uint32(0)
	for off+8 <= uint32(len(buf)) {
		param, hdr := word(off), word(off+4)
		reg := uint16(hdr & 0xFFFF)
		mask := uint8(hdr >> 16 & 0xF)
		n := hdr >> 20 & 0xFF
		consec := hdr>>31 != 0

		end := off + 8 + n*4
		if n%2 == 1 {
			end += 4 // pad word keeps entries 8-byte aligned
		}
		if end > uint32(len(buf)) {
			return ws, fmt.Errorf("pica: entry at 0x%X (reg 0x%03X, %d extras) runs past the %d-byte buffer",
				off, reg, n, len(buf))
		}

		ws = append(ws, PICAWrite{Off: off, Reg: reg, Mask: mask, Value: param, Burst: n > 0})
		for i := uint32(0); i < n; i++ {
			r := reg
			if consec {
				r += uint16(i) + 1
			}
			ws = append(ws, PICAWrite{Off: off, Reg: r, Mask: mask, Value: word(off + 8 + i*4), Burst: true})
		}
		off = end
	}
	return ws, nil
}

// PICARegGroup names the functional block a register id belongs to, for
// annotating instrument dumps. Groups (not individual registers) — individual
// register semantics are established one by one as the GPU implements them.
func PICARegGroup(reg uint16) string {
	switch {
	case reg < 0x040:
		return "misc/irq"
	case reg < 0x080:
		return "rasterizer" // culling, viewport, depth map, scissor
	case reg < 0x0C0:
		return "texturing" // texture unit config/addresses/types
	case reg < 0x100:
		return "tev" // the six texture-environment (combiner) stages
	case reg < 0x140:
		return "framebuffer" // output merger: blend, alpha/depth test, color/depth targets
	case reg < 0x180:
		return "fragment-lighting"
	case reg < 0x200:
		return "unknown-180"
	case reg < 0x228:
		return "geometry-pipeline" // attribute buffers: base, formats, strides
	case reg < 0x260:
		return "geometry-pipeline" // index buffer, draw triggers, fixed attrs
	case reg < 0x280:
		return "geometry-shader"
	case reg < 0x2C0:
		return "vertex-shader" // code/operand-descriptor/uniform uploads, entry, input map
	default:
		return "unknown-2C0"
	}
}
