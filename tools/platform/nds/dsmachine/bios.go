package dsmachine

import (
	"retroreverse.com/tools/cpu/arm"

	"retroreverse.com/tools/platform/nds"
)

// The BIOS software interrupts.
//
// These are high-level-emulated, the way tools/platform/psx HLEs the PSX BIOS: the
// DS BIOS is not a kernel, it is a *library* — memory copies, decompressors, a
// divider, and the interrupt waits — and reimplementing that library in Go is both
// simpler and more auditable than interpreting a ROM we do not have. Nothing above
// them depends on their instruction timing.
//
// The two interesting ones are the waits. IntrWait and Halt do not spin: they *park*
// the core, and it stays parked until the scheduler delivers an interrupt it asked
// for (run.go). That is what lets an idle ARM7 cost nothing while the ARM9 runs, and
// it is why a missing interrupt source shows up as a core stuck in `waiting` rather
// than as a busy loop that merely looks slow.

func biosSWI(c *core) func(*arm.CPU, uint32) bool {
	b := &bus{c: c}
	return func(cpu *arm.CPU, comment uint32) bool {
		n := comment & 0xFF
		if n == 0 { // a Thumb SWI encodes its number in the low byte of the halfword
			n = (comment >> 16) & 0xFF
		}
		switch n {
		case 0x03: // WaitByLoop(count) — a busy-wait.
			// Rather than skip it (which collapses the delays a polled handshake relies
			// on), sleep the caller for that many instruction-times so the *other* core
			// runs meanwhile.
			d := int(cpu.R[0])
			if d > 1<<20 {
				d = 1 << 20 // cap pathological waits
			}
			c.sleep += d

		case 0x04: // IntrWait(discardOld, mask)
			c.park(cpu, cpu.R[1], false)
		case 0x05: // VBlankIntrWait
			c.park(cpu, irqVBlank, false)
		case 0x06: // Halt — wake on any interrupt
			c.park(cpu, 0, true)
		case 0x1F: // CustomHalt
			c.park(cpu, 0, true)

		case 0x07: // Stop / deep sleep — the end of the line for a boot trace
			cpu.Halt("%s: SWI Stop (deep sleep)", c.name)

		case 0x09: // Div(num, den) -> quotient in r0, remainder in r1, |quotient| in r3
			num, den := int32(cpu.R[0]), int32(cpu.R[1])
			if den == 0 {
				break // the hardware's behaviour here is undefined; leave the registers be
			}
			q, r := num/den, num%den
			cpu.R[0], cpu.R[1] = uint32(q), uint32(r)
			if q < 0 {
				q = -q
			}
			cpu.R[3] = uint32(q)

		case 0x0D: // Sqrt
			cpu.R[0] = isqrt(uint64(cpu.R[0]))

		case 0x0E: // GetCRC16(initial, addr, len)
			data := make([]byte, cpu.R[2])
			for i := range data {
				data[i] = b.Read(cpu.R[1] + uint32(i))
			}
			cpu.R[0] = uint32(nds.CRC16(data))

		case 0x0B: // CpuSet
			cpuSet(b, cpu, false)
		case 0x0C: // CpuFastSet
			cpuSet(b, cpu, true)

		case 0x0F: // IsDebugger — we are not one
			cpu.R[0] = 0

		case 0x11, 0x12: // LZ77UnComp (write 8-bit / write 16-bit)
			lz77UnComp(b, cpu.R[0], cpu.R[1])
		case 0x14, 0x15: // RLUnComp
			rlUnComp(b, cpu.R[0], cpu.R[1])
		case 0x16, 0x18: // Diff8bitUnFilter / Diff16bitUnFilter
			diffUnFilter(b, cpu.R[0], cpu.R[1], n == 0x18)

		default:
			c.m.note("%s: SWI 0x%02X not implemented (ignored)", c.name, n)
		}
		return true
	}
}

// park puts the core into an interrupt wait: it stops stepping until the scheduler
// delivers a matching IRQ.
//
// It deliberately does NOT record a resume address. The obvious thing — capture the
// PC here — captures the SWI's OWN address, because this hook runs from inside the
// instruction and the core advances R15 only after it returns. The core would then
// resume by re-executing the wait and parking again, for ever. By the time the
// scheduler stops stepping the core, R15 is already the instruction after the SWI,
// which is exactly the address to come back to, so deliver() just uses it.
func (c *core) park(cpu *arm.CPU, mask uint32, any bool) {
	c.waiting = true
	c.waitMask = mask
	c.waitAny = any
}

// cpuSet implements CpuSet / CpuFastSet: a memory copy, or a fill when bit 24 of the
// control word is set.
func cpuSet(b *bus, c *arm.CPU, fast bool) {
	src, dst, ctrl := c.R[0], c.R[1], c.R[2]
	fill := ctrl&(1<<24) != 0
	n := ctrl & 0x1FFFFF
	word := fast || ctrl&(1<<26) != 0
	if word {
		var v uint32
		if fill {
			v = b.r32(src)
		}
		for i := uint32(0); i < n; i++ {
			if !fill {
				v = b.r32(src + i*4)
			}
			b.w32(dst+i*4, v)
		}
		return
	}
	var v uint32 // 16-bit units
	if fill {
		v = uint32(b.Read(src)) | uint32(b.Read(src+1))<<8
	}
	for i := uint32(0); i < n; i++ {
		if !fill {
			v = uint32(b.Read(src+i*2)) | uint32(b.Read(src+i*2+1))<<8
		}
		b.Write(dst+i*2, byte(v))
		b.Write(dst+i*2+1, byte(v>>8))
	}
}

// lz77UnComp is the BIOS's LZ77 decompressor. The header word carries the type in
// its top byte and the decompressed size in the lower three; then it is a stream of
// flag bytes, each governing eight following items — a literal byte, or a
// back-reference of 3..18 bytes to somewhere in the last 4 KiB.
func lz77UnComp(b *bus, src, dst uint32) {
	hdr := b.r32(src)
	size := hdr >> 8
	src += 4
	var written uint32
	for written < size {
		flags := b.Read(src)
		src++
		for i := 0; i < 8 && written < size; i++ {
			if flags&(0x80>>uint(i)) == 0 {
				b.Write(dst+written, b.Read(src))
				src++
				written++
				continue
			}
			hi, lo := b.Read(src), b.Read(src+1)
			src += 2
			length := uint32(hi>>4) + 3
			disp := uint32(hi&0xF)<<8 | uint32(lo) + 1
			for j := uint32(0); j < length && written < size; j++ {
				b.Write(dst+written, b.Read(dst+written-disp))
				written++
			}
		}
	}
}

// rlUnComp is the BIOS's run-length decompressor: each block byte is either a run of
// one repeated byte (bit 7 set, length in the low bits + 3) or a literal span.
func rlUnComp(b *bus, src, dst uint32) {
	hdr := b.r32(src)
	size := hdr >> 8
	src += 4
	var written uint32
	for written < size {
		f := b.Read(src)
		src++
		if f&0x80 != 0 {
			n := uint32(f&0x7F) + 3
			v := b.Read(src)
			src++
			for i := uint32(0); i < n && written < size; i++ {
				b.Write(dst+written, v)
				written++
			}
		} else {
			n := uint32(f&0x7F) + 1
			for i := uint32(0); i < n && written < size; i++ {
				b.Write(dst+written, b.Read(src))
				src++
				written++
			}
		}
	}
}

// diffUnFilter undoes a delta filter: the stream holds differences, and each output
// is the running sum. Used on gradient-heavy data, where the deltas compress far
// better than the values.
func diffUnFilter(b *bus, src, dst uint32, wide bool) {
	hdr := b.r32(src)
	size := hdr >> 8
	src += 4
	if wide {
		var sum uint16
		for i := uint32(0); i < size; i += 2 {
			d := uint16(b.Read(src)) | uint16(b.Read(src+1))<<8
			src += 2
			sum += d
			b.Write(dst+i, byte(sum))
			b.Write(dst+i+1, byte(sum>>8))
		}
		return
	}
	var sum uint8
	for i := uint32(0); i < size; i++ {
		sum += b.Read(src)
		src++
		b.Write(dst+i, sum)
	}
}

// isqrt is an integer square root by bit-by-bit restoration (see math.go for why
// this is not done in floating point).
func isqrt(v uint64) uint32 {
	var res, bit uint64 = 0, 1 << 62
	for bit > v {
		bit >>= 2
	}
	for bit != 0 {
		if v >= res+bit {
			v -= res + bit
			res = res>>1 + bit
		} else {
			res >>= 1
		}
		bit >>= 2
	}
	return uint32(res)
}

// cp15 handles the ARM9's CP15 coprocessor: it captures the ITCM/DTCM base and size
// the boot programs (registers c9,c1), so the machine maps the tightly-coupled
// memories where the game expects them. Everything else — the caches, the MPU
// regions, the write buffer — is accepted and ignored: it changes performance on
// hardware, not behaviour.
func cp15(c *core) func(*arm.CPU, bool, uint32, uint32, uint32, uint32, uint32, *uint32) {
	return func(cpu *arm.CPU, load bool, cp, op1, crn, crm, op2 uint32, rd *uint32) {
		if load || crn != 9 || crm != 1 {
			return
		}
		v := *rd
		base := v &^ 0xFFF
		if op2 == 0 { // DTCM base/size
			if c.dtcm == nil {
				c.dtcm = make([]byte, 0x4000)
			}
			c.dtcmBase = base
			c.handlerBase = base + uint32(len(c.dtcm))
			c.m.note("ARM9: CP15 DTCM base 0x%08X", base)
		} else { // ITCM control — its base is fixed at 0 on the DS; only the size varies
			c.m.note("ARM9: CP15 ITCM control 0x%08X", v)
		}
	}
}
