package gbamachine

import (
	"encoding/binary"
	"testing"

	"retroreverse.com/tools/platform/gba"
)

// newTest builds a machine over a synthetic 1 MiB cartridge.
func newTest(t *testing.T) *Machine {
	t.Helper()
	data := make([]byte, 1<<20)
	binary.LittleEndian.PutUint32(data[0x00:], 0xEA00002E)
	copy(data[0xA0:], "TEST")
	copy(data[0xAC:], "TEST")
	data[0xB2] = 0x96
	data[0xBD] = gba.ComplementCheck(data)
	rom, err := gba.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return New(rom)
}

// TestByteStoreQuirks pins the hardware rules the bus comments claim: a byte
// store to palette/BG-VRAM duplicates into both halves of the halfword, and a
// byte store to OAM (or OBJ VRAM) is dropped entirely. A model that just stored
// the byte would pass every "does it boot" test and mis-render sprites.
func TestByteStoreQuirks(t *testing.T) {
	m := newTest(t)
	b := &bus{m: m}

	b.Write(0x05000000, 0xAB)
	if got := b.Read16(0x05000000); got != 0xABAB {
		t.Errorf("palette byte store: halfword = %#04X, want 0xABAB", got)
	}

	b.Write(0x06000000, 0xCD)
	if got := b.Read16(0x06000000); got != 0xCDCD {
		t.Errorf("BG VRAM byte store: halfword = %#04X, want 0xCDCD", got)
	}

	b.Write16(0x07000000, 0x1234)
	b.Write(0x07000000, 0xFF)
	if got := b.Read16(0x07000000); got != 0x1234 {
		t.Errorf("OAM byte store was not ignored: %#04X", got)
	}

	// OBJ VRAM (mode 0: from 0x06010000) also ignores byte stores.
	b.Write16(0x06010000, 0x4321)
	b.Write(0x06010000, 0xFF)
	if got := b.Read16(0x06010000); got != 0x4321 {
		t.Errorf("OBJ VRAM byte store was not ignored: %#04X", got)
	}
}

// TestMemoryMirrors covers the mirrors games actually use — top-of-IWRAM at
// 0x03FFFFxx above all, which is where the BIOS IRQ handler pointer lives.
func TestMemoryMirrors(t *testing.T) {
	m := newTest(t)
	b := &bus{m: m}
	b.Write32(0x03007FFC, 0xDEADBEEF)
	if got := b.Read32(0x03FFFFFC); got != 0xDEADBEEF {
		t.Errorf("IWRAM top mirror: %#08X, want 0xDEADBEEF", got)
	}
	b.Write32(0x02000000, 0x12345678)
	if got := b.Read32(0x02040000); got != 0x12345678 {
		t.Errorf("EWRAM mirror: %#08X, want 0x12345678", got)
	}
}

// TestDMAImmediate checks a plain memory-to-memory transfer and that the channel
// disarms afterwards.
func TestDMAImmediate(t *testing.T) {
	m := newTest(t)
	b := &bus{m: m}
	for i := uint32(0); i < 16; i++ {
		b.Write32(0x02000000+i*4, 0x1000+i)
	}
	m.ioWrite16(0x0D4, 0x0000) // DMA3 src lo
	m.ioWrite16(0x0D6, 0x0200) // src hi -> 0x02000000
	m.ioWrite16(0x0D8, 0x1000) // dst lo
	m.ioWrite16(0x0DA, 0x0300) // dst hi -> 0x03001000
	m.ioWrite16(0x0DC, 16)     // count
	m.ioWrite16(0x0DE, 1<<15|1<<10)

	for i := uint32(0); i < 16; i++ {
		if got := b.Read32(0x03001000 + i*4); got != 0x1000+i {
			t.Fatalf("word %d = %#X, want %#X", i, got, 0x1000+i)
		}
	}
	if m.dma[3].ctrl&(1<<15) != 0 {
		t.Error("immediate DMA left the channel enabled")
	}
}

// TestTimerCascade covers the overflow chain: timer 0 prescaled, timer 1
// counting timer 0's overflows.
func TestTimerCascade(t *testing.T) {
	m := newTest(t)
	m.ioWrite16(0x100, 0xFFFF) // TM0 reload: overflows every tick
	m.ioWrite16(0x102, 1<<7)   // TM0 on, prescaler 1
	m.ioWrite16(0x104, 0xFFF0) // TM1 reload
	m.ioWrite16(0x106, 1<<7|1<<2)

	m.tickTimers(4) // 4 cycles -> 4 TM0 overflows -> TM1 +... one per scanline call
	if m.timers[1].counter == 0xFFF0 {
		t.Error("cascade timer did not advance on the lower timer's overflow")
	}
}

// TestLZ77Decompress round-trips a stream the BIOS SWI must decode. The literal
// path and the back-reference path are both exercised, and the VRAM variant is
// checked to produce the same bytes as the WRAM one — the halfword buffering
// must not change the OUTPUT, only how it is written.
func TestLZ77Decompress(t *testing.T) {
	m := newTest(t)
	b := &bus{m: m}
	want := []byte("ABCABCABCABCDDDD")

	// header: type 1, size 16; then flag byte + tokens.
	stream := []byte{0x10, byte(len(want)), 0x00, 0x00}
	stream = append(stream, 0x00, 'A', 'B', 'C') // 8 literal slots, 3 used...
	// flags: bit7..0 -> first three literals emitted above used slots 7,6,5.
	stream[4] = 0b00010000                      // slot 4 (the 4th token) is a back-reference
	stream = append(stream, 0x60, 0x02)         // len 3+6=9, disp 3
	stream = append(stream, 'D', 'D', 'D', 'D') // 4 literals
	for i, v := range stream {
		b.Write(0x02000000+uint32(i), v)
	}

	m.lz77(b, 0x02000000, 0x03000000, false)
	got := make([]byte, len(want))
	for i := range got {
		got[i] = b.Read(0x03000000 + uint32(i))
	}
	if string(got) != string(want) {
		t.Errorf("LZ77 (WRAM) = %q, want %q", got, want)
	}

	m.lz77(b, 0x02000000, 0x06000000, true)
	for i := range got {
		got[i] = b.Read(0x06000000 + uint32(i))
	}
	if string(got) != string(want) {
		t.Errorf("LZ77 (VRAM) = %q, want %q", got, want)
	}
}

// TestEEPROMFraming is the regression for the bug Minish Cap surfaced as
// "Unable to save file.": a 14-bit request's first 9 bits form a valid 6-bit
// request, so the device must be framed by the DMA transfer length, not parsed
// bit by bit. A per-bit parser sizes itself to 512 B here and fails the
// round-trip.
func TestEEPROMFraming(t *testing.T) {
	m := newTest(t)
	e := &m.eeprom

	writeReq := func(addr int, data []byte) {
		bits := []byte{1, 0} // write
		for i := 13; i >= 0; i-- {
			bits = append(bits, byte(addr>>uint(i)&1))
		}
		for _, by := range data {
			for i := 7; i >= 0; i-- {
				bits = append(bits, by>>uint(i)&1)
			}
		}
		bits = append(bits, 0) // stop
		for _, b := range bits {
			e.write(uint16(b))
		}
		e.endFrame(m)
	}
	readReq := func(addr int) []byte {
		bits := []byte{1, 1}
		for i := 13; i >= 0; i-- {
			bits = append(bits, byte(addr>>uint(i)&1))
		}
		bits = append(bits, 0)
		for _, b := range bits {
			e.write(uint16(b))
		}
		e.endFrame(m)
		for i := 0; i < 4; i++ {
			e.read() // the 4 ignore bits
		}
		out := make([]byte, 8)
		for i := range out {
			var v byte
			for bit := 0; bit < 8; bit++ {
				v = v<<1 | byte(e.read())
			}
			out[i] = v
		}
		return out
	}

	want := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	writeReq(0x25, want)
	if len(e.data) != 8192 {
		t.Fatalf("14-bit request sized the part to %d bytes, want 8192", len(e.data))
	}
	if got := readReq(0x25); string(got) != string(want) {
		t.Errorf("EEPROM round-trip = %v, want %v", got, want)
	}
	// A different block must not have been written.
	if got := readReq(0x26); string(got) == string(want) {
		t.Error("write landed in more than one block")
	}
}

// TestSavestateDeepCopy guards the repo's snapshot-aliasing scar: a snapshot
// that shares the machine's slices resumes identically and is still not a
// snapshot. Mutate the machine after snapshotting and the snapshot must not
// move; mutate it after restoring and the snapshot must still not move.
func TestSavestateDeepCopy(t *testing.T) {
	m := newTest(t)
	b := &bus{m: m}
	b.Write32(0x02000000, 0x11111111)
	s := m.snapshot()

	b.Write32(0x02000000, 0x22222222)
	if binary.LittleEndian.Uint32(s.EWRAM[0:]) != 0x11111111 {
		t.Error("snapshot aliases the machine's EWRAM (write after snapshot changed it)")
	}

	m.restore(s)
	if got := b.Read32(0x02000000); got != 0x11111111 {
		t.Fatalf("restore did not put the value back: %#08X", got)
	}
	b.Write32(0x02000000, 0x33333333)
	if binary.LittleEndian.Uint32(s.EWRAM[0:]) != 0x11111111 {
		t.Error("snapshot aliases the machine's EWRAM after restore")
	}
}

// TestForcedBlankIsWhite checks the one display rule that is easy to get
// backwards: DISPCNT bit 7 blanks the screen to WHITE, not to black.
func TestForcedBlankIsWhite(t *testing.T) {
	m := newTest(t)
	m.ioWrite16(0x000, 1<<7)
	m.ppu.renderLine(m, 0)
	if m.screen[0] != 0xFFFFFFFF {
		t.Errorf("forced blank pixel = %#08X, want 0xFFFFFFFF", m.screen[0])
	}
}

// TestPSGChannelsSound is the control for the synthesis path. Minish Cap's
// title screen renders in silence, and the only way to tell "the model's PSG is
// broken" from "the game is not playing anything there" is to drive each
// channel directly and confirm it CAN make a sound.
func TestPSGChannelsSound(t *testing.T) {
	cases := []struct {
		name  string
		setup func(m *Machine)
	}{
		{"square1", func(m *Machine) {
			m.ioWrite16(0x062, 0xF780) // duty 50%, volume 15, no decay
			m.ioWrite16(0x064, 0x8400) // trigger, freq 0x400
		}},
		{"square2", func(m *Machine) {
			m.ioWrite16(0x068, 0xF780)
			m.ioWrite16(0x06C, 0x8400)
		}},
		{"wave", func(m *Machine) {
			// The wave channel's two banks are not interchangeable storage: the
			// CPU accesses the bank that is NOT playing, so a driver fills the
			// idle bank and then swaps to it. Writing with bank 0 selected and
			// expecting to hear it is the mistake this sequence avoids — the
			// data lands in bank 1, and playback keeps reading bank 0.
			m.ioWrite16(0x070, 0x0080) // DAC on, playback bank 0 -> CPU writes bank 1
			for r := uint32(0x090); r <= 0x09E; r += 2 {
				m.ioWrite16(r, 0x0FF0) // a non-flat waveform
			}
			m.ioWrite16(0x070, 0x00C0) // swap: playback now reads the filled bank
			m.ioWrite16(0x072, 0x2000) // full volume
			m.ioWrite16(0x074, 0x8400) // trigger
		}},
		{"noise", func(m *Machine) {
			m.ioWrite16(0x078, 0xF000) // volume 15, no decay
			m.ioWrite16(0x07C, 0x8004) // trigger, a mid divisor
		}},
	}
	for _, c := range cases {
		m := newTest(t)
		m.AudioCapture(true)
		m.ioWrite16(0x084, 0x0080) // master power
		m.ioWrite16(0x080, 0xFF77) // every channel, both sides, max volume
		m.ioWrite16(0x082, 0x0002) // PSG at 100%
		c.setup(m)
		m.apu.mixCycles(1677721) // 0.1s
		var peak int16
		for _, s := range m.apu.PCM {
			if s > peak {
				peak = s
			}
		}
		if peak == 0 {
			t.Errorf("PSG %s produced silence", c.name)
		}
	}
}

// TestDirectSoundTransport drives the whole Direct Sound chain the way a game
// does — a timer clocking the FIFO and a DMA in "special" timing refilling it
// from RAM — and checks the game's own bytes come out the other end. This path
// carries a GBA game's music, so a break in it is silence with every register
// still reading back correctly.
func TestDirectSoundTransport(t *testing.T) {
	m := newTest(t)
	b := &bus{m: m}
	m.AudioCapture(true)

	// A ramp for the driver's mix buffer.
	for i := uint32(0); i < 64; i++ {
		b.Write(0x02000000+i, byte(i*2))
	}
	m.ioWrite16(0x084, 0x0080) // master power
	m.ioWrite16(0x082, 0x0304) // Direct Sound A at 100%, both sides, timer 0
	// DMA1: source = the buffer, dest = FIFO_A, special timing, repeat.
	m.ioWrite16(0x0BC, 0x0000)
	m.ioWrite16(0x0BE, 0x0200)
	m.ioWrite16(0x0C0, 0x00A0)
	m.ioWrite16(0x0C2, 0x0400)
	m.ioWrite16(0x0C4, 4)
	m.ioWrite16(0x0C6, 1<<15|1<<10|1<<9|3<<12)
	// Timer 0: overflow every 512 cycles.
	m.ioWrite16(0x100, 0xFE00)
	m.ioWrite16(0x102, 1<<7)

	// The FIFO starts empty, so the first drain must pull the DMA refill in.
	for i := 0; i < 8; i++ {
		m.tickTimers(cyclesPerLine)
	}
	if m.apu.dsA.cur == 0 {
		t.Fatalf("Direct Sound A never received a sample (fifo has %d)", len(m.apu.dsA.q))
	}
	if m.dma[1].latchSrc == 0x02000000 {
		t.Error("the sound DMA never advanced its source pointer")
	}
	m.apu.mixCycles(100000)
	var peak int16
	for _, s := range m.apu.PCM {
		if s > peak {
			peak = s
		}
	}
	if peak == 0 {
		t.Error("Direct Sound carried samples but the mixer output silence")
	}
}

// TestExporterAgreesWithPPU is the guard on the deliberate duplication between
// the scanline renderer here and the offline decoder in tools/platform/gba
// (tiles.go). They exist separately on purpose — one answers "what does the
// wire carry on line Y", the other "what does the whole map look like" — but
// two implementations of the same tile format drift apart silently unless
// something makes them agree. This renders a synthetic scene through both and
// requires the visible window to match pixel for pixel.
func TestExporterAgreesWithPPU(t *testing.T) {
	m := newTest(t)
	b := &bus{m: m}

	// A 4bpp character set: tile n is a solid block of palette index n.
	for n := 1; n < 16; n++ {
		for row := 0; row < 8; row++ {
			for byteIdx := 0; byteIdx < 4; byteIdx++ {
				b.Write16(0x06000000+uint32(n*32+row*4+byteIdx)&^1,
					uint16(n)<<12|uint16(n)<<8|uint16(n)<<4|uint16(n))
			}
		}
	}
	// Palette bank 0: index i is a distinguishable colour.
	for i := 1; i < 16; i++ {
		b.Write16(0x05000000+uint32(i)*2, uint16(i)*0x0421)
	}
	// A tilemap at screenblock 8 (0x4000) cycling tiles, with some flips.
	for ty := 0; ty < 32; ty++ {
		for tx := 0; tx < 32; tx++ {
			e := uint16((tx+ty)%15 + 1)
			if tx%3 == 0 {
				e |= 1 << 10 // hflip
			}
			if ty%4 == 0 {
				e |= 1 << 11 // vflip
			}
			b.Write16(0x06004000+uint32(ty*32+tx)*2, e)
		}
	}

	m.ioWrite16(0x000, 0x0100) // mode 0, BG0 on
	m.ioWrite16(0x008, 0x0800) // BG0: charbase 0, screenbase 8, 4bpp, 32x32
	m.ioWrite16(0x010, 0)      // no scroll — the exporter renders the map unscrolled
	m.ioWrite16(0x012, 0)

	for y := 0; y < screenH; y++ {
		m.ppu.renderLine(m, y)
	}

	vram := m.Snapshot(0x06000000, vramSize)
	pal := gba.ParsePalette(m.Snapshot(0x05000000, palSize))
	layer := gba.DecodeBGCNT(m.Reg(0x008))
	img := layer.Render(vram, pal)

	diff := 0
	for y := 0; y < screenH; y++ {
		for x := 0; x < screenW; x++ {
			c := img.RGBAAt(x, y)
			want := uint32(0xFF)<<24 | uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B)
			if c.A == 0 {
				want = 0xFF000000 | uint32(pal.RGBA(0).R)<<16 |
					uint32(pal.RGBA(0).G)<<8 | uint32(pal.RGBA(0).B)
			}
			if m.screen[y*screenW+x] != want {
				diff++
			}
		}
	}
	if diff != 0 {
		t.Errorf("exporter and PPU disagree on %d of %d pixels", diff, screenW*screenH)
	}
}
