package merc

// replay.go plays the game's own merc DMA chains out of a RAM image: a
// DMA-tag walker and VIF interpreter feed tools/cpu/vu exactly the stream
// the real VIF1 would, so no part of the upload protocol is reconstructed
// by hand. XGKick packets are parsed at kick time. Patching the art group's
// color records in the RAM image beforehand gives every packet vertex its
// input identity (the same index-encoding trick as the Session path).

import (
	"encoding/binary"
	"fmt"

	"retroreverse.com/tools/cpu/vu"
)

// Replayer holds the machine and VIF state for one chain walk.
type Replayer struct {
	RAM []byte
	V   *vu.VU

	cl, wl   int
	mode     int
	row      [4]uint32
	mask     uint32
	base     uint32
	offset   uint32
	dbf      bool
	tops     uint16

	Packets [][]OutVert // one slice per kicked packet, in kick order
	Trace   bool
	SkipRun bool // process uploads only (a dry pass to collect MPG code)
	RefLo, RefHi uint32 // count ref tags into this range
	RefHits      int
	Refs         []uint32

	// Direct, when set, receives every DIRECT/DIRECTHL payload (PATH2 GIF
	// data — the bucket-init register packets live here).
	Direct    func(qws []byte)
	LastEntry uint32 // byte address of the last MSCAL entry
}

func NewReplayer(ram []byte) *Replayer {
	r := &Replayer{RAM: ram, cl: 1, wl: 1}
	r.V = vu.New(make([]byte, 16384), make([]byte, 16384))
	r.V.ResetBranchLog(8)
	r.V.XGKick = func(qw uint32) {
		vs, err := parseGIF(r.V.Data, qw)
		if err == nil {
			r.Packets = append(r.Packets, vs)
		}
	}
	return r
}

func (r *Replayer) u32(addr uint32) uint32 {
	if int64(addr)+4 > int64(len(r.RAM)) {
		return 0
	}
	return binary.LittleEndian.Uint32(r.RAM[addr:])
}

// Play walks a DMA chain starting at addr until an END/RET-style terminator.
func (r *Replayer) Play(addr uint32) error {
	var stack []uint32
	for hops := 0; hops < 20000; hops++ {
		tag := binary.LittleEndian.Uint64(r.RAM[addr:])
		qwc := uint32(tag & 0x7FFF)
		id := int(tag >> 28 & 7)
		ref := uint32(tag >> 32 & 0x7FFFFFFF)
		// TTE: the tag's upper 64 bits go to VIF as two codes.
		v0 := r.u32(addr + 8)
		v1 := r.u32(addr + 12)
		var data uint32
		next := addr + 16 + qwc*16
		switch id {
		case 0: // REFE: transfer ref, then end
			data = ref
			next = 0
		case 1: // CNT
			data = addr + 16
		case 2: // NEXT
			data = addr + 16
			next = ref
		case 3, 4: // REF, REFS
			data = ref
			next = addr + 16
			if ref >= r.RefLo && ref < r.RefHi {
				r.RefHits++
			}
			r.Refs = append(r.Refs, ref)
		case 5: // CALL
			data = addr + 16
			if len(stack) < 2 {
				stack = append(stack, addr+16+qwc*16)
			}
			next = ref
		case 6: // RET
			data = addr + 16
			if len(stack) > 0 {
				next = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			} else {
				next = 0
			}
		case 7: // END
			data = addr + 16
			next = 0
		default:
			return fmt.Errorf("replay: dma id %d at 0x%X", id, addr)
		}
		if r.Trace {
			fmt.Printf("dma@%06X id=%d qwc=%d ref=%06X vif=%08X %08X\n", addr, id, qwc, ref, v0, v1)
		}
		if r.RefHi > 0 && data >= r.RefLo && data < r.RefHi {
			r.RefHits++
		}
		if qwc > 8 {
			r.Refs = append(r.Refs, data)
		}
		stream := make([]uint32, 0, 2+qwc*4)
		stream = append(stream, v0, v1)
		for i := uint32(0); i < qwc*4; i++ {
			stream = append(stream, r.u32(data+i*4))
		}
		if err := r.vif(stream); err != nil {
			return err
		}
		if next == 0 {
			return nil
		}
		addr = next
	}
	return fmt.Errorf("replay: chain too long")
}

// vif interprets a stream of VIF1 codes and inline data.
func (r *Replayer) vif(w []uint32) error {
	i := 0
	next := func() uint32 {
		if i >= len(w) {
			return 0
		}
		v := w[i]
		i++
		return v
	}
	for i < len(w) {
		code := next()
		cmd := code >> 24 & 0x7F
		num := int(code >> 16 & 0xFF)
		imm := code & 0xFFFF
		switch {
		case cmd == 0x00: // NOP
		case cmd == 0x01: // STCYCL
			r.cl = int(imm & 0xFF)
			r.wl = int(imm >> 8 & 0xFF)
		case cmd == 0x02:
			r.offset = imm & 0x3FF
		case cmd == 0x03:
			r.base = imm & 0x3FF
		case cmd == 0x04: // ITOP
		case cmd == 0x05:
			r.mode = int(imm & 3)
		case cmd == 0x06, cmd == 0x07: // MSKPATH3, MARK
		case cmd >= 0x10 && cmd <= 0x13: // FLUSH*
		case cmd == 0x14, cmd == 0x15: // MSCAL, MSCALF
			// The VU latches the CURRENT TOPS; DBF then flips so the
			// following unpacks fill the other buffer.
			vuTop := uint16(r.base)
			if r.dbf {
				vuTop = uint16((r.base + r.offset) & 0x3FF)
			}
			r.dbf = !r.dbf
			r.tops = uint16(r.base)
			if r.dbf {
				r.tops = uint16((r.base + r.offset) & 0x3FF)
			}
			r.V.Top = vuTop
			r.LastEntry = imm * 8 & 0x3FFF
			if !r.SkipRun {
				if _, ok := r.V.Run(imm*8&0x3FFF, 600000); !ok {
					e := imm * 8 & 0x3FFF
					hdr := ""
					for q := uint32(r.tops) + 140; q < uint32(r.tops)+145; q++ {
						hdr += fmt.Sprintf(" {%08X %08X %08X %08X}",
							binary.LittleEndian.Uint32(r.V.Data[(q&1023)*16:]),
							binary.LittleEndian.Uint32(r.V.Data[(q&1023)*16+4:]),
							binary.LittleEndian.Uint32(r.V.Data[(q&1023)*16+8:]),
							binary.LittleEndian.Uint32(r.V.Data[(q&1023)*16+12:]))
					}
					return fmt.Errorf("replay: microprogram did not halt (MSCAL 0x%X top=%d; spin %x; in:%s)",
						e, r.tops, r.V.BranchTrail(), hdr)
				}
			}
		case cmd == 0x17: // MSCNT: continue at the PC after the last halt
			vuTop := uint16(r.base)
			if r.dbf {
				vuTop = uint16((r.base + r.offset) & 0x3FF)
			}
			r.dbf = !r.dbf
			r.tops = uint16(r.base)
			if r.dbf {
				r.tops = uint16((r.base + r.offset) & 0x3FF)
			}
			r.V.Top = vuTop
			if !r.SkipRun {
				if _, ok := r.V.Run(r.V.PC, 600000); !ok {
					return fmt.Errorf("replay: microprogram did not halt (MSCNT)")
				}
			}
		case cmd == 0x20: // STMASK
			r.mask = next()
		case cmd == 0x30: // STROW
			for k := 0; k < 4; k++ {
				r.row[k] = next()
			}
		case cmd == 0x31: // STCOL
			for k := 0; k < 4; k++ {
				next()
			}
		case cmd == 0x4A: // MPG
			n := num
			if n == 0 {
				n = 256
			}
			addr := int(imm) * 8
			for k := 0; k < n; k++ {
				lo := next()
				hi := next()
				if addr+8 <= len(r.V.Micro) {
					binary.LittleEndian.PutUint32(r.V.Micro[addr:], lo)
					binary.LittleEndian.PutUint32(r.V.Micro[addr+4:], hi)
				}
				addr += 8
			}
		case cmd == 0x50, cmd == 0x51: // DIRECT/HL: imm qwords to GIF (PATH2)
			n := int(imm)
			if n == 0 {
				n = 65536
			}
			if r.Direct != nil && i+n*4 <= len(w) {
				buf := make([]byte, n*16)
				for k := 0; k < n*4; k++ {
					binary.LittleEndian.PutUint32(buf[k*4:], w[i+k])
				}
				r.Direct(buf)
			}
			i += n * 4
		case cmd >= 0x60: // UNPACK
			if err := r.unpack(code, w, &i); err != nil {
				return err
			}
		default:
			return fmt.Errorf("replay: vif cmd 0x%02X", cmd)
		}
	}
	return nil
}

// unpack performs an UNPACK write into VU data memory.
func (r *Replayer) unpack(code uint32, w []uint32, i *int) error {
	vn := int(code >> 26 & 3)
	vl := int(code >> 24 & 3)
	num := int(code >> 16 & 0xFF)
	if num == 0 {
		num = 256
	}
	imm := code & 0xFFFF
	addr := uint32(imm & 0x3FF)
	usn := imm&0x4000 != 0
	if imm&0x8000 != 0 {
		addr = (addr + uint32(r.tops)) & 0x3FF
	}
	// byte-granular reader over the remaining words
	bytePos := 0
	byteAt := func(k int) uint32 {
		word := w[*i+ (bytePos+k)/4]
		return word >> (8 * uint((bytePos+k)%4)) & 0xFF
	}
	half := func(k int) uint32 {
		word := w[*i+(bytePos+2*k)/4]
		v := word >> (16 * uint(((bytePos+2*k)/2)%2)) & 0xFFFF
		return v
	}
	elemSize := [4]int{4, 2, 1, 2} // bytes per element by vl (vl3 = V4-5)
	ncomp := vn + 1
	consumed := 0
	write := func(qw uint32, comps []uint32) {
		for k, c := range comps {
			v := c
			switch r.mode {
			case 1:
				v = c + r.row[k]
			case 2:
				v = c + r.row[k]
				r.row[k] = v
			}
			if int(qw)*16+k*4+4 <= len(r.V.Data) {
				binary.LittleEndian.PutUint32(r.V.Data[qw*16+uint32(k)*4:], v)
			}
		}
	}
	cl, wl := r.cl, r.wl
	if wl == 0 {
		wl = 1
	}
	out := 0
	for out < num {
		within := out % wl
		if cl >= wl || within < cl {
			var comps []uint32
			for c := 0; c < ncomp; c++ {
				var val uint32
				switch vl {
				case 0:
					val = w[*i+(bytePos+4*(consumed*ncomp+c))/4]
				case 1:
					v := half(consumed*ncomp + c)
					if !usn {
						val = uint32(int32(int16(v)))
					} else {
						val = v
					}
				case 2:
					v := byteAt(consumed*ncomp + c)
					if !usn {
						val = uint32(int32(int8(v)))
					} else {
						val = v
					}
				}
				comps = append(comps, val)
			}
			// V2/V3 leave remaining lanes; keep simple: write ncomp lanes
			write(addr, comps)
			consumed++
		}
		addr = (addr + 1) & 0x3FF
		out++
	}
	// advance i past consumed data (word-aligned)
	bits := consumed * ncomp * elemSize[vl] * 8
	words := (bits + 31) / 32
	*i += words
	return nil
}
