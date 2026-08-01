package gbamachine

// The cartridge's serial EEPROM, addressed through the 0x0D000000 window one BIT
// per bus access: the game lays a request out in RAM one bit per halfword, DMAs
// it to the window, then DMAs 68 halfwords back to read the answer. Minish Cap's
// driver is EEPROM_V124 (writeup Part I §5).
//
//	read  request: "11" + address + stop      ->  9 bits (6-bit addr, 512 B part)
//	                                             17 bits (14-bit addr, 8 KiB part)
//	write request: "10" + address + 64 data + stop -> 73 or 81 bits
//	read  reply:   4 ignore bits + 64 data bits    -> 68 bits
//
// THE REQUEST IS FRAMED BY THE DMA, NOT BY ITS CONTENT. That is not a detail: a
// 17-bit read request's first 9 bits are a perfectly well-formed 9-bit read
// request, so a device that parses bit-by-bit commits to the wrong address width
// on the first request it ever sees, sizes itself to 512 B, and thereafter reads
// and writes the wrong blocks. Minish Cap's symptom was precise and misleading —
// the file-select screen read back a "corrupted" file and name entry ended in
// "Unable to save file." Real hardware has the same ambiguity and resolves it the
// same way this does: the transfer ends when the DMA ends, so the device counts
// the bits it was handed and only then decides what it was asked.

type eeprom struct {
	present bool
	data    []byte // 8 KiB, or 512 B once a 6-bit driver identifies the part
	sized   bool

	inBits  []byte // request bits accumulated since the last frame end
	outBits []byte // reply bits pending for bus reads
	ready   bool   // idle/complete: reads answer 1
}

func (e *eeprom) init() {
	e.present = true
	e.data = make([]byte, 8192)
	for i := range e.data {
		e.data[i] = 0xFF // an erased EEPROM reads all-ones
	}
	e.ready = true
}

// write shifts one request bit in (the low bit of the halfword).
func (e *eeprom) write(v uint16) { e.inBits = append(e.inBits, byte(v&1)) }

// read pops one reply bit; when no reply is pending, reports the ready state.
func (e *eeprom) read() uint16 {
	if len(e.outBits) > 0 {
		v := uint16(e.outBits[0])
		e.outBits = e.outBits[1:]
		return v
	}
	if e.ready {
		return 1
	}
	return 0
}

// endFrame is called by the DMA when a transfer into the EEPROM window finishes:
// the accumulated bits are one complete request, and its LENGTH says which.
func (e *eeprom) endFrame(m *Machine) {
	bits := e.inBits
	e.inBits = e.inBits[:0]
	if len(bits) < 3 {
		return // a stray transfer, not a request
	}
	var addrBits int
	switch len(bits) {
	case 9, 73:
		addrBits = 6
	case 17, 81:
		addrBits = 14
	default:
		m.note("EEPROM request of %d bits is not a known frame length (ignored)", len(bits))
		return
	}
	e.size(addrBits == 14)

	addr := 0
	for i := 0; i < addrBits; i++ {
		addr = addr<<1 | int(bits[2+i])
	}
	addr &= len(e.data)/8 - 1 // one address names one 8-byte block

	switch {
	case bits[0] == 1 && bits[1] == 1: // read
		e.outBits = e.outBits[:0]
		for i := 0; i < 4; i++ {
			e.outBits = append(e.outBits, 0) // 4 ignore bits
		}
		for _, b := range e.data[addr*8 : addr*8+8] {
			for bit := 7; bit >= 0; bit-- {
				e.outBits = append(e.outBits, b>>uint(bit)&1)
			}
		}
	case bits[0] == 1 && bits[1] == 0: // write
		blk := e.data[addr*8 : addr*8+8]
		for i := range blk {
			var b byte
			for bit := 0; bit < 8; bit++ {
				b = b<<1 | bits[2+addrBits+i*8+bit]
			}
			blk[i] = b
		}
		e.ready = true
	default:
		m.note("EEPROM request opens with %d%d (neither a read nor a write)", bits[0], bits[1])
	}
}

// size fixes the part's capacity the first time a driver reveals its address
// width. The 14-bit form addresses an 8 KiB part; the 6-bit form a 512 B one.
func (e *eeprom) size(large bool) {
	if e.sized {
		return
	}
	e.sized = true
	if !large {
		e.data = e.data[:512]
	}
}
