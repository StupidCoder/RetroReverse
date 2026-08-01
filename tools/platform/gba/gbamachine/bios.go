package gbamachine

// The BIOS software interrupts, high-level-emulated the way dsmachine (and
// tools/platform/psx) does it: the GBA BIOS is a library — memory fills, a
// divider, decompressors, affine-matrix helpers and the interrupt waits — and
// reimplementing it in Go is simpler and more auditable than interpreting a ROM
// image this repository does not carry. Nothing above these depends on their
// instruction timing.
//
// The waits are the interesting ones. Halt and IntrWait PARK the CPU: it stays
// parked until the scheduler delivers an interrupt (run.go), so an idle game
// costs nothing. IntrWait's protocol is the GBA's own: the BIOS spins on the
// check flags at 0x03007FF8, which the GAME'S OWN HANDLER must set — the BIOS
// only clears them. A game whose handler never writes the flags hangs in
// VBlankIntrWait on hardware too.
//
// Anything not implemented halts the core with the SWI number, so gaps are
// explicit rather than silently wrong.

import (
	"math"

	"retroreverse.com/tools/cpu/arm"
)

func biosSWI(m *Machine) func(*arm.CPU, uint32) bool {
	b := &bus{m: m}
	return func(c *arm.CPU, comment uint32) bool {
		n := comment & 0xFF
		if !c.Thumb {
			n = (comment >> 16) & 0xFF // an ARM SWI encodes the number in bits 16-23
		}
		switch n {
		case 0x01: // RegisterRamReset(flags)
			f := c.R[0]
			if f&1 != 0 {
				clear(m.ewram)
			}
			if f&2 != 0 { // IWRAM except the top 0x200 the BIOS owns
				clear(m.iwram[:iwramSize-0x200])
			}
			if f&4 != 0 {
				clear(m.pal)
			}
			if f&8 != 0 {
				clear(m.vram)
			}
			if f&16 != 0 {
				clear(m.oam)
			}
			// f&32/64/128 reset SIO/sound/other registers; the register file model
			// starts zeroed, so there is nothing further to clear honestly.

		case 0x02: // Halt: park until any enabled interrupt
			m.waiting, m.waitAny, m.waitMask = true, true, 0

		case 0x04: // IntrWait(discardOld, mask)
			m.intrWait(b, c.R[0] != 0, uint16(c.R[1]))

		case 0x05: // VBlankIntrWait = IntrWait(1, V-blank)
			m.intrWait(b, true, irqVBlank)

		case 0x06: // Div: r0/r1 -> r0 quotient, r1 remainder, r3 |quotient|
			num, den := int32(c.R[0]), int32(c.R[1])
			if den == 0 {
				// The real BIOS returns ±1 (sign of the numerator) and loops the
				// remainder; a divide by zero is a game bug worth surfacing.
				m.note("BIOS Div by zero at PC 0x%08X", c.R[15])
				q := int32(1)
				if num < 0 {
					q = -1
				}
				c.R[0], c.R[1] = uint32(q), uint32(num)
				c.R[3] = 1
				break
			}
			q := num / den
			c.R[0], c.R[1] = uint32(q), uint32(num%den)
			if q < 0 {
				q = -q
			}
			c.R[3] = uint32(q)

		case 0x07: // DivArm: r1/r0 (the backwards one)
			c.R[0], c.R[1] = c.R[1], c.R[0]
			return biosSWI(m)(c, 0x06<<16) // reuse Div (ARM-form comment)

		case 0x08: // Sqrt
			c.R[0] = uint32(math.Sqrt(float64(c.R[0])))

		case 0x09: // ArcTan(tan 1.14) -> angle
			c.R[0] = uint32(uint16(int16(math.Round(math.Atan(float64(int16(c.R[0]))/16384) * 0x8000 / math.Pi))))

		case 0x0A: // ArcTan2(x, y) -> 0..0xFFFF angle
			a := math.Atan2(float64(int16(c.R[1])), float64(int16(c.R[0])))
			c.R[0] = uint32(uint16(math.Round(a * 0x8000 / math.Pi)))

		case 0x0B: // CpuSet(src, dst, ctl)
			cnt := c.R[2] & 0x1FFFFF
			fill := c.R[2]&(1<<24) != 0
			if c.R[2]&(1<<26) != 0 { // 32-bit units
				src, dst := c.R[0]&^3, c.R[1]&^3
				v := b.Read32(src)
				for i := uint32(0); i < cnt; i++ {
					if !fill {
						v = b.Read32(src + i*4)
					}
					b.Write32(dst+i*4, v)
				}
			} else {
				src, dst := c.R[0]&^1, c.R[1]&^1
				v := b.Read16(src)
				for i := uint32(0); i < cnt; i++ {
					if !fill {
						v = b.Read16(src + i*2)
					}
					b.Write16(dst+i*2, v)
				}
			}

		case 0x0C: // CpuFastSet: words, count rounded up to 8
			cnt := (c.R[2]&0x1FFFFF + 7) &^ 7
			fill := c.R[2]&(1<<24) != 0
			src, dst := c.R[0]&^3, c.R[1]&^3
			v := b.Read32(src)
			for i := uint32(0); i < cnt; i++ {
				if !fill {
					v = b.Read32(src + i*4)
				}
				b.Write32(dst+i*4, v)
			}

		case 0x0D: // GetBiosChecksum
			c.R[0] = 0xBAAE187F // the AGB BIOS's sum (a platform constant games compare against)

		case 0x0E: // BgAffineSet(src, dst, num)
			m.bgAffineSet(b, c.R[0], c.R[1], c.R[2])

		case 0x0F: // ObjAffineSet(src, dst, num, stride)
			m.objAffineSet(b, c.R[0], c.R[1], c.R[2], c.R[3])

		case 0x10: // BitUnPack(src, dst, paramsPtr)
			m.bitUnPack(b, c.R[0], c.R[1], c.R[2])

		case 0x11, 0x12: // LZ77UnComp to WRAM (byte writes) / VRAM (halfword writes)
			m.noteDecompress(b, n, c.R[0], c.R[1], c.R[15])
			m.lz77(b, c.R[0], c.R[1], n == 0x12)

		case 0x13: // HuffUnComp
			m.noteDecompress(b, n, c.R[0], c.R[1], c.R[15])
			m.huffman(b, c.R[0], c.R[1])

		case 0x14, 0x15: // RLUnComp WRAM/VRAM
			m.noteDecompress(b, n, c.R[0], c.R[1], c.R[15])
			m.rle(b, c.R[0], c.R[1], n == 0x15)

		case 0x19: // SoundBias — sound is not modelled yet; the bias write is harmless
			m.io[0x88] = uint16(c.R[0]&1)<<9 | 0x100

		case 0x1F: // MidiKey2Freq(wave, key, fine)
			base := b.Read32(c.R[0] + 4)
			c.R[0] = uint32(float64(base) / math.Exp2(float64(180-c.R[1])/12-float64(c.R[2])/3072))

		case 0x28: // SoundDriverVSyncOff — a no-op without the sound driver's DMA
		case 0x29: // SoundDriverVSyncOn

		default:
			c.Halt("unimplemented BIOS SWI 0x%02X (r0=0x%08X r1=0x%08X r2=0x%08X) at 0x%08X",
				n, c.R[0], c.R[1], c.R[2], c.R[15])
		}
		return true
	}
}

// noteDecompress reports a decompression to the OnDecompress hook. The size
// comes out of the stream's own header, so the report names the compressed
// source AND how much it expands to — which is exactly what is needed to find
// an asset in the ROM and then check a reimplemented decoder against it.
func (m *Machine) noteDecompress(b *bus, swi, src, dst, pc uint32) {
	if m.OnDecompress == nil {
		return
	}
	m.OnDecompress(swi, src, dst, int(b.Read32(src)>>8), pc, m.cpu.R[14])
}

// intrWait implements SWI 4/5: the BIOS enables IME, optionally discards stale
// check flags, and parks the CPU until the game's handler reports one of mask.
func (m *Machine) intrWait(b *bus, discard bool, mask uint16) {
	m.ime = true // the real BIOS forces the master enable on
	flags := uint16(b.r32(irqCheckFlags))
	if !discard && flags&mask != 0 {
		b.w32(irqCheckFlags, uint32(flags&^mask))
		return
	}
	if discard {
		b.w32(irqCheckFlags, uint32(flags&^mask))
	}
	m.waiting, m.waitAny, m.waitMask = true, false, mask
}

// --- affine helpers ----------------------------------------------------------

// bgAffineSet: for each entry, build the BG PA-PD/X/Y set from original centre,
// display centre, scales and angle.
func (m *Machine) bgAffineSet(b *bus, src, dst, num uint32) {
	for ; num > 0; num-- {
		ox := float64(int32(b.Read32(src))) / 256 // 24.8 original centre
		oy := float64(int32(b.Read32(src+4))) / 256
		dx := float64(int16(b.Read16(src + 8))) // display centre
		dy := float64(int16(b.Read16(src + 10)))
		sx := float64(int16(b.Read16(src+12))) / 256
		sy := float64(int16(b.Read16(src+14))) / 256
		theta := float64(b.Read16(src+16)>>8) / 128 * math.Pi
		src += 20

		sin, cos := math.Sin(theta), math.Cos(theta)
		pa, pb := sx*cos, -sx*sin
		pc, pd := sy*sin, sy*cos
		b.w16(dst, uint16(int16(math.Round(pa*256))))
		b.w16(dst+2, uint16(int16(math.Round(pb*256))))
		b.w16(dst+4, uint16(int16(math.Round(pc*256))))
		b.w16(dst+6, uint16(int16(math.Round(pd*256))))
		x0 := ox - (pa*dx + pb*dy)
		y0 := oy - (pc*dx + pd*dy)
		b.w32(dst+8, uint32(int32(math.Round(x0*256))))
		b.w32(dst+12, uint32(int32(math.Round(y0*256))))
		dst += 16
	}
}

// objAffineSet: PA-PD only, with a caller-chosen stride (2 = packed, 8 = OAM).
func (m *Machine) objAffineSet(b *bus, src, dst, num, stride uint32) {
	for ; num > 0; num-- {
		sx := float64(int16(b.Read16(src))) / 256
		sy := float64(int16(b.Read16(src+2))) / 256
		theta := float64(b.Read16(src+4)>>8) / 128 * math.Pi
		src += 8
		sin, cos := math.Sin(theta), math.Cos(theta)
		put := func(off uint32, v float64) {
			b.w16(dst+off*stride, uint16(int16(math.Round(v*256))))
		}
		put(0, sx*cos)
		put(1, -sx*sin)
		put(2, sy*sin)
		put(3, sy*cos)
		dst += 4 * stride
	}
}

// --- unpack / decompressors --------------------------------------------------

// bitUnPack widens sub-byte pixel data (the header names source length, source
// width, destination width and a bias added to non-zero units).
func (m *Machine) bitUnPack(b *bus, src, dst, params uint32) {
	srcLen := uint32(b.Read16(params))
	srcW := int(b.Read(params + 2))
	dstW := int(b.Read(params + 3))
	ofs := b.Read32(params + 4)
	bias := ofs & 0x7FFFFFFF
	zeroToo := ofs&(1<<31) != 0

	var out, outBits uint32
	flush := func(bits int) {
		outBits += uint32(bits)
		if outBits == 32 {
			b.Write32(dst, out)
			dst += 4
			out, outBits = 0, 0
		}
	}
	for i := uint32(0); i < srcLen; i++ {
		v := uint32(b.Read(src + i))
		for bit := 0; bit < 8; bit += srcW {
			unit := v >> uint(bit) & (1<<uint(srcW) - 1)
			if unit != 0 || zeroToo {
				unit += bias
			}
			out |= unit << outBits
			flush(dstW)
		}
	}
	if outBits > 0 {
		b.Write32(dst, out)
	}
}

// lzHeader reads a decompression header and checks its type nibble.
func (m *Machine) lzHeader(b *bus, src uint32, wantType uint32, name string) (size uint32, ok bool) {
	h := b.Read32(src)
	if h>>4&0xF != wantType {
		m.cpu.Halt("BIOS %s: header 0x%08X at 0x%08X is not type %d", name, h, src, wantType)
		return 0, false
	}
	return h >> 8, true
}

// lz77 is SWI 0x11/0x12. The VRAM variant buffers halfwords, because byte writes
// to VRAM do not store bytes (see bus.Write) — this is why the BIOS has two.
func (m *Machine) lz77(b *bus, src, dst uint32, vram bool) {
	size, ok := m.lzHeader(b, src, 1, "LZ77UnComp")
	if !ok {
		return
	}
	src += 4
	var pend uint16
	var pendBytes uint32
	put := func(v byte) {
		if !vram {
			b.Write(dst, v)
			dst++
			return
		}
		pend |= uint16(v) << (8 * pendBytes)
		if pendBytes++; pendBytes == 2 {
			b.Write16(dst, pend)
			dst += 2
			pend, pendBytes = 0, 0
		}
	}
	// For back-references the VRAM variant must read through its own pending
	// byte; keep a window of what was emitted.
	written := make([]byte, 0, size)
	emit := func(v byte) {
		written = append(written, v)
		put(v)
	}
	for uint32(len(written)) < size {
		flags := b.Read(src)
		src++
		for bit := 7; bit >= 0 && uint32(len(written)) < size; bit-- {
			if flags>>uint(bit)&1 == 0 {
				emit(b.Read(src))
				src++
				continue
			}
			b1, b2 := b.Read(src), b.Read(src+1)
			src += 2
			length := int(b1>>4) + 3
			disp := int(b1&0xF)<<8 | int(b2) + 1
			for i := 0; i < length && uint32(len(written)) < size; i++ {
				emit(written[len(written)-disp])
			}
		}
	}
	if vram && pendBytes == 1 {
		b.Write16(dst, pend)
	}
}

// rle is SWI 0x14/0x15.
func (m *Machine) rle(b *bus, src, dst uint32, vram bool) {
	size, ok := m.lzHeader(b, src, 3, "RLUnComp")
	if !ok {
		return
	}
	src += 4
	var pend uint16
	var pendBytes, done uint32
	put := func(v byte) {
		done++
		if !vram {
			b.Write(dst, v)
			dst++
			return
		}
		pend |= uint16(v) << (8 * pendBytes)
		if pendBytes++; pendBytes == 2 {
			b.Write16(dst, pend)
			dst += 2
			pend, pendBytes = 0, 0
		}
	}
	for done < size {
		f := b.Read(src)
		src++
		if f&0x80 != 0 { // run
			v := b.Read(src)
			src++
			for i := 0; i < int(f&0x7F)+3 && done < size; i++ {
				put(v)
			}
		} else { // literals
			for i := 0; i < int(f&0x7F)+1 && done < size; i++ {
				put(b.Read(src))
				src++
			}
		}
	}
	if vram && pendBytes == 1 {
		b.Write16(dst, pend)
	}
}

// huffman is SWI 0x13: a bit-packed tree in front, then 32-bit chunks of
// MSB-first bitstream. Output is written in 32-bit units.
func (m *Machine) huffman(b *bus, src, dst uint32) {
	h := b.Read32(src)
	dataBits := int(h & 0xF) // 4 or 8 bits per decoded unit
	size := h >> 8
	treeSize := uint32(b.Read(src+4))*2 + 1
	treeRoot := src + 5
	bits := treeRoot + treeSize // aligned: 4 + 1 + treeSize lands on a word boundary

	var out, outBits, done uint32
	node := treeRoot
	nodeVal := b.Read(node)
	for done < size {
		w := b.Read32(bits)
		bits += 4
		for i := 31; i >= 0; i-- {
			bit := w >> uint(i) & 1
			// Child pair sits at (nodeAddr &^ 1) + offset*2 + 2; bit 1 = right.
			child := (node&^1 + uint32(nodeVal&0x3F)*2 + 2) + bit
			leaf := nodeVal&(0x80>>bit) != 0
			nodeVal = b.Read(child)
			node = child
			if leaf {
				out |= uint32(nodeVal) << outBits
				outBits += uint32(dataBits)
				node = treeRoot
				nodeVal = b.Read(node)
				if outBits == 32 {
					b.Write32(dst, out)
					dst += 4
					out, outBits = 0, 0
					if done += 4; done >= size {
						break
					}
				}
			}
		}
	}
}
