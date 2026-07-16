package ps2

// gif.go is the GIF — the Graphics Interface, the unit that stands between a DMA channel
// and the Graphics Synthesizer. The GIF's whole job is to unpack a list of "GIFtags" into
// writes of the GS's registers: a tag says how many register writes follow and which
// registers they are for, and the data after it carries the values.
//
// A GIFtag is one quadword:
//
//	bits   0..14   NLOOP   how many times to repeat the register list
//	bit    15      EOP     the last tag of the packet
//	bit    46      PRE     write PRIM (below) before the loop
//	bits   47..57  PRIM    a primitive to start
//	bits   58..59  FLG     the format of the data that follows:
//	                         0 PACKED   a quadword per register, with a layout per register
//	                         1 REGLIST  two 64-bit values per quadword, packed
//	                         2 IMAGE    raw data, straight to the GS transfer register
//	bits   60..63  NREG    how many registers in the list (0 means 16)
//	bits   64..127 REGS    the list itself, four bits per register
//
// The one register descriptor that carries its own address is A+D (0xE): the quadword's
// low 64 bits are the value and the next byte is which GS register to write. An image
// upload is built almost entirely of A+D writes — BITBLTBUF, TRXPOS, TRXREG, TRXDIR —
// followed by an IMAGE tag whose data is the pixels. Which registers a packet drives is
// the game telling us what it wants of the GS, one write at a time; nothing here decides
// in advance.

// The GIFtag FLG field.
const (
	gifPacked  = 0
	gifReglist = 1
	gifImage   = 2
	gifDisable = 3
)

// The PACKED register descriptors that are not a plain GS register address. Only A+D
// matters to the upload path; the drawing descriptors are named so a trace reads.
const (
	gifRegPrim  = 0x0
	gifRegRGBAQ = 0x1
	gifRegST    = 0x2
	gifRegUV    = 0x3
	gifRegXYZF2 = 0x4
	gifRegXYZ2  = 0x5
	gifRegAD    = 0xE // the value carries its own GS register address
	gifRegNOP   = 0xF
)

// gifStart runs a transfer the DMA controller has handed the GIF. An image upload uses
// normal mode — a flat run of quadwords at MADR — and the render path uses a source
// chain: the game builds its display list as scattered buffers linked by DMAtags, and
// the channel walks them. Either way the GIF sees one stream; the chain's links are
// gathered before parsing because a GIFtag's loop may run across a link boundary, and
// the parser holds no state between calls.
func (m *Machine) gifStart(c *dmacChan) {
	switch (c.chcr & dChcrModeM) >> 2 {
	case 1: // source chain
		var all []byte
		m.dmacSourceChain(dmacChGIF, c, func(b []byte) {
			if len(b) == 8 {
				// A TTE tag-forward. The GIF is a quadword device and PATH3 chains do not
				// ride codes on their tags; a game that sets TTE here would be saying
				// something new, so it is a note, not a silent append.
				m.note("GS: the GIF channel forwarded a chain tag (TTE) — unhandled")
				return
			}
			all = append(all, b...)
		}, false)
		if len(all) > 0 {
			m.gifPacket(all)
		}
	default:
		if c.qwc == 0 {
			return
		}
		m.gifPacket(m.dmaBytes(c.madr, c.qwc))
	}
}

// gifPacket unpacks a buffer of GIFtags into GS register writes.
func (m *Machine) gifPacket(data []byte) {
	gs := m.ensureGS()
	pos := 0
	for pos+16 <= len(data) {
		lo := le64(data[pos:])
		hi := le64(data[pos+8:])
		pos += 16

		nloop := uint32(lo & 0x7FFF)
		eop := lo&(1<<15) != 0
		pre := lo&(1<<46) != 0
		prim := uint32((lo >> 47) & 0x7FF)
		flg := (lo >> 58) & 3
		nreg := int((lo >> 60) & 0xF)
		if nreg == 0 {
			nreg = 16
		}
		regs := hi // sixteen 4-bit descriptors

		if pre {
			gs.write(gsPRIM, uint64(prim))
		}

		switch flg {
		case gifPacked:
			for n := uint32(0); n < nloop; n++ {
				for r := 0; r < nreg; r++ {
					if pos+16 > len(data) {
						return
					}
					desc := (regs >> (4 * uint(r))) & 0xF
					dlo := le64(data[pos:])
					dhi := le64(data[pos+8:])
					pos += 16
					m.gifPacked(gs, desc, dlo, dhi)
				}
			}

		case gifReglist:
			total := nloop * uint32(nreg)
			for i := uint32(0); i < total; i++ {
				if pos+8 > len(data) {
					return
				}
				desc := (regs >> (4 * uint(i%uint32(nreg)))) & 0xF
				val := le64(data[pos:])
				pos += 8
				if desc != gifRegNOP {
					gs.write(uint8(desc), val)
				}
			}
			if total&1 != 0 { // a trailing half-quadword of padding
				pos += 8
			}

		case gifImage:
			n := int(nloop) * 16
			if pos+n > len(data) {
				n = len(data) - pos
			}
			gs.imageData(data[pos : pos+n])
			pos += n

		case gifDisable:
			pos += int(nloop) * 16
		}

		if eop {
			// The packet may still hold another tag (a DMA transfer can carry several
			// packets); keep unpacking until the buffer is spent.
			_ = eop
		}
	}
}

// gifPacked handles one PACKED register write. The descriptor either is A+D — the value
// carries its own GS register address — or names a drawing register directly.
func (m *Machine) gifPacked(gs *GS, desc uint64, lo, hi uint64) {
	switch desc {
	case gifRegAD:
		reg := uint8(hi & 0xFF)
		gs.write(reg, lo)
	case gifRegNOP:
	default:
		// A drawing register in PACKED layout. The upload path does not use these; they are
		// the primitive stream, handled once the rasteriser exists. For now the GS records
		// the write so the census shows the game is drawing, not just uploading.
		gs.writePacked(uint8(desc), lo, hi)
	}
}

// gifPacketLen walks GIFtags from the start of data to the end of the packet — through
// the tag whose EOP bit is set — and returns its byte length. XGKICK needs it: the
// program names where its packet starts, and the packet itself says where it ends.
func gifPacketLen(data []byte) int {
	pos := 0
	for pos+16 <= len(data) {
		lo := le64(data[pos:])
		nloop := int(lo & 0x7FFF)
		eop := lo&(1<<15) != 0
		flg := lo >> 58 & 3
		nreg := int(lo>>60) & 0xF
		if nreg == 0 {
			nreg = 16
		}
		pos += 16
		switch flg {
		case gifPacked:
			pos += nloop * nreg * 16
		case gifReglist:
			n := nloop * nreg
			pos += (n + n&1) * 8
		default: // IMAGE and the disabled alias both carry NLOOP quadwords
			pos += nloop * 16
		}
		if eop {
			break
		}
	}
	if pos > len(data) {
		pos = len(data)
	}
	return pos
}

// le64 reads a little-endian 64-bit word out of a byte slice.
func le64(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}
