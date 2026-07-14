package dsmachine

import "encoding/binary"

// The cartridge interface. Everything the game is — its code overlays, its models,
// its textures, its music — arrives through this one 32-bit port, four bytes at a
// time.
//
// The protocol is a command/response one. Software writes an 8-byte command to the
// command register, then writes ROMCTRL with the block-start bit set; the card
// clocks the response into a FIFO, and software drains it word by word from
// 0x04100010 (or, in practice, points a DMA channel at it: start mode 5). While
// data is outstanding, ROMCTRL bit 23 reads "word ready" and bit 31 reads "busy".
//
// We serve the response straight out of the ROM image. There is no encryption
// layer here and there does not need to be one: KEY1/KEY2 scramble the *bus*, not
// the data, and a machine that is both ends of the bus can simply agree not to.

const (
	regAUXSPICNT  = 0x040001A0
	regAUXSPIDATA = 0x040001A2
	regROMCTRL    = 0x040001A4
	regCARDCMD    = 0x040001A8 // 8 bytes, big-endian
	regCARDDATA   = 0x04100010
	regEXMEMCNT   = 0x04000204
)

const (
	romctrlWordReady = 1 << 23
	romctrlBusy      = 1 << 31
)

// card is the cartridge and the state of the transfer in flight.
type card struct {
	rom []byte

	cmd    [8]byte
	ctrl   uint32
	buf    []byte // the response still to be handed over
	pos    int
	owner9 bool // which CPU the slot is wired to (EXMEMCNT bit 11)

	// OnXfer, if set, reports every transfer the card serves: the command, the
	// ROM address it decoded, and how many bytes it will hand back. This is the DS's
	// -dmalog — the map of which region of the cartridge loads when, drawn by the
	// game's own loader rather than guessed.
	OnXfer func(cmd [8]byte, addr uint32, size int)
}

// blockBytes decodes the transfer length ROMCTRL asks for. The encoding is not a
// plain size: 0 means no data at all, 7 means exactly four bytes, and 1..6 mean
// 0x100 bytes shifted left by n — so the smallest real block is 512 bytes, not 256.
//
// The off-by-one here is worth a sentence, because it cost a whole boot and it was
// the GAME that settled it rather than any document. SM64DS's filesystem reads the
// name table by taking its ROM offset (0x1D9324), rounding it DOWN to a 512-byte
// boundary (0x1D9200), reading that block, and then indexing 0x124 bytes into what
// came back. Decode the size field as `0x100 << (n-1)` and a field of 1 yields 256
// bytes: the read succeeds, every byte of it is correct, and the block simply stops
// short of the offset the game is about to read from. The filesystem then resolves
// its first path against padding, and the game hangs in an assert
// (`myFS_ConvertPathToFileID`) with no bad data anywhere to find.
func blockBytes(ctrl uint32) int {
	switch n := (ctrl >> 24) & 7; n {
	case 0:
		return 0
	case 7:
		return 4
	default:
		return 0x100 << n
	}
}

// start begins a transfer: it decodes the command and fills the response buffer.
func (cd *card) start(ctrl uint32) {
	cd.ctrl = ctrl
	n := blockBytes(ctrl)
	cd.pos = 0
	cd.buf = cd.buf[:0]
	if n == 0 {
		cd.ctrl &^= romctrlBusy | romctrlWordReady
		return
	}

	switch cd.cmd[0] {
	case 0xB7: // read data
		// The address is big-endian in the command. A read wraps *within its 4 KiB
		// block* rather than running on into the next one — a detail that only shows
		// up when a game reads across a block boundary, and then shows up as data
		// that is subtly, unreproducibly wrong.
		addr := binary.BigEndian.Uint32(cd.cmd[1:5])
		if cd.OnXfer != nil {
			cd.OnXfer(cd.cmd, addr, n)
		}
		for i := 0; i < n; i++ {
			a := addr&^0xFFF | (addr+uint32(i))&0xFFF
			cd.buf = append(cd.buf, cd.romByte(a))
		}
	case 0xB8: // chip ID, repeated to fill the block
		id := []byte{0xC2, 0x7F, 0x3F, 0x00}
		for i := 0; i < n; i++ {
			cd.buf = append(cd.buf, id[i&3])
		}
	default:
		// Anything else (the KEY1 init chatter, 0x9F dummy, 0x3C) has no data we need
		// to be right about: hand back all-ones, which is what an idle bus reads as.
		for i := 0; i < n; i++ {
			cd.buf = append(cd.buf, 0xFF)
		}
	}
	cd.ctrl |= romctrlBusy | romctrlWordReady
}

func (cd *card) romByte(a uint32) byte {
	if int(a) < len(cd.rom) {
		return cd.rom[a]
	}
	return 0xFF
}

// readData hands over the next word of the response and, when the last one goes,
// ends the transfer.
func (c *core) cardReadData() uint32 {
	cd := c.m.cd
	if cd.pos >= len(cd.buf) {
		return 0xFFFFFFFF
	}
	v := binary.LittleEndian.Uint32(cd.buf[cd.pos:])
	cd.pos += 4
	if cd.pos >= len(cd.buf) {
		cd.ctrl &^= romctrlBusy | romctrlWordReady
		// The completion interrupt is enabled in AUXSPICNT, not in ROMCTRL — the one
		// place the card's two control registers overlap in meaning.
		if c.io[regAUXSPICNT]&0x4000 != 0 {
			c.raise(irqCard)
		}
	}
	return v
}

// cardCore returns the core the cartridge slot currently belongs to. EXMEMCNT bit
// 11 hands it to one CPU or the other, and the loser reads an open bus — which is
// how the ARM7 and ARM9 avoid fighting over one FIFO.
func (m *Machine) cardCore() *core {
	if m.ARM9.io[regEXMEMCNT]&0x0800 == 0 {
		return m.ARM9
	}
	return m.ARM7
}
