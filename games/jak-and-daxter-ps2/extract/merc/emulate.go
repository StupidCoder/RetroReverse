package merc

// emulate.go runs the merc VU1 microprogram (tools/cpu/vu) on a fragment's
// unpacked input and parses the GIF packet it kicks — the exact vertices,
// ST/RGBA, ADC bits and strip structure, with no hand-derived pipeline
// logic. The VIF side reproduces draw-bones-merc's chain: STROW
// {0x47800000, 0x4B010000, stMagic, stMagic}, STMOD 1 for the two V4-8
// unpacks (row bits + byte, as the VIF adds them), STMOD 0 for the V4-32
// fp region at TOP+140+byteNUM+lumpNUM, bone matrices at their control
// pairs' VU addresses, and MSCAL at the merc entry.

import (
	"encoding/binary"
	"fmt"

	"retroreverse.com/tools/cpu/vu"
)

// EmuConfig carries what the runtime would supply: the microcode, the
// low-memory constant block, and the bone matrix palette.
type EmuConfig struct {
	Micro  []byte
	LowMem []byte // 140 quadwords: rows 0-7, camera/lights 132-139 as chosen
	// Bones maps a matrix-palette index to its 7-quadword upload image.
	Bones map[byte][]byte
	Entry uint32 // microcode entry, byte address (0x128 for the merc main entry)
	Init  bool   // run the constant-derivation prologue (entry 0) first
	TracePC func(v *vu.VU, pc uint32)
	Top   uint16 // input double-buffer base in quadwords
}

// OutVert is one GIF vertex from the kicked packet.
type OutVert struct {
	ST   [3]float32 // s, t, q
	RGBA [4]uint8
	XYZ  [3]float32 // fixed 12.4, converted
	ADC  bool
}

// Emulate runs one fragment and returns the kicked GIF packets' vertices in
// kick order.
func Emulate(cfg *EmuConfig, fr *Fragment, stMagic uint32) ([]OutVert, error) {
	data := make([]byte, 16384)
	copy(data, cfg.LowMem)

	row := [4]uint32{0x47800000, 0x4B010000, stMagic, stMagic}
	top := uint32(cfg.Top)

	put := func(qw uint32, vals [4]uint32) {
		for k, v := range vals {
			binary.LittleEndian.PutUint32(data[qw*16+uint32(k)*4:], v)
		}
	}
	// V4-8 with STMOD 1: each source byte lands in one word as row[k]+b.
	addr := top + 140
	for r := 0; r < fr.ByteQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = row[k] + uint32(fr.ByteData[r*4+k])
		}
		put(addr, w)
		addr++
	}
	for r := 0; r < fr.LumpQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = row[k] + uint32(fr.LumpData[r*4+k])
		}
		put(addr, w)
		addr++
	}
	// V4-32, STMOD 0: raw quadwords.
	for r := 0; r < fr.FPQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = binary.LittleEndian.Uint32(fr.FPData[r*16+k*4:])
		}
		put(addr, w)
		addr++
	}
	// Bone matrices at their control destinations (7 qw each).
	for _, m := range fr.Mats {
		img, ok := cfg.Bones[m.Index]
		if !ok || len(img) < 7*16 {
			return nil, fmt.Errorf("merc emulate: no bone image for palette %d", m.Index)
		}
		copy(data[uint32(m.Dest)*16:], img[:7*16])
	}

	micro := cfg.Micro
	if len(micro) < 16384 {
		m := make([]byte, 16384)
		copy(m, micro)
		micro = m
	}
	v := vu.New(micro, data)
	v.Top = cfg.Top
	v.ResetBranchLog(16)
	if cfg.TracePC != nil {
		v.Trace = func(vm *vu.VU, pc uint32, raw uint64) { cfg.TracePC(vm, pc) }
	}
	var kicks []uint32
	v.XGKick = func(qw uint32) { kicks = append(kicks, qw) }
	if cfg.Init {
		if _, ok := v.Run(0, 4000); !ok {
			return nil, fmt.Errorf("merc emulate: init prologue did not halt; %x", v.BranchTrail())
		}
	}
	if _, ok := v.Run(cfg.Entry, 400000); !ok {
		return nil, fmt.Errorf("merc emulate: did not halt; last branches %x", v.BranchTrail())
	}
	var out []OutVert
	for _, k := range kicks {
		vs, err := parseGIF(v.Data, k)
		if err != nil {
			return nil, err
		}
		out = append(out, vs...)
	}
	return out, nil
}

// parseGIF walks a GIF packet in VU memory: PACKED mode with the merc
// register list (ST, RGBAQ, XYZF2); EOP ends the packet.
func parseGIF(mem []byte, qw uint32) ([]OutVert, error) {
	var out []OutVert
	for {
		if qw*16+16 > uint32(len(mem)) {
			return out, nil
		}
		tag := binary.LittleEndian.Uint64(mem[qw*16:])
		nloop := int(tag & 0x7FFF)
		eop := tag&0x8000 != 0
		nreg := int(tag >> 60 & 0xF)
		if nreg == 0 {
			nreg = 16
		}
		regs := binary.LittleEndian.Uint64(mem[qw*16+8:])
		qw++
		for l := 0; l < nloop; l++ {
			var v OutVert
			for r := 0; r < nreg; r++ {
				reg := regs >> (4 * r) & 0xF
				base := qw * 16
				if base+16 > uint32(len(mem)) {
					return out, nil
				}
				switch reg {
				case 0x2: // ST(Q)
					v.ST[0] = f32bits(binary.LittleEndian.Uint32(mem[base:]))
					v.ST[1] = f32bits(binary.LittleEndian.Uint32(mem[base+4:]))
					v.ST[2] = f32bits(binary.LittleEndian.Uint32(mem[base+8:]))
				case 0x1: // RGBAQ
					v.RGBA = [4]uint8{mem[base], mem[base+4], mem[base+8], mem[base+12]}
				case 0x4: // XYZF2, ADC in bit 15 of the fourth word
					x := binary.LittleEndian.Uint16(mem[base:])
					y := binary.LittleEndian.Uint16(mem[base+4:])
					z := binary.LittleEndian.Uint32(mem[base+8:])
					w := binary.LittleEndian.Uint32(mem[base+12:])
					v.XYZ = [3]float32{float32(int16(x)) / 16, float32(int16(y)) / 16, float32(z >> 4)}
					v.ADC = w&0x8000 != 0
				}
				qw++
			}
			out = append(out, v)
		}
		if eop {
			return out, nil
		}
	}
}

func f32bits(v uint32) float32 { return float32frombits(v) }
