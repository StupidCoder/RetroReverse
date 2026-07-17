package ps2

import "testing"

// buildADPacket builds a GIF PACKED packet of A+D register writes: one qword per
// {register, value}, the format sceGsExecLoadImage uses to set up an image transfer.
func buildADPacket(writes [][2]uint64) []byte {
	n := uint64(len(writes))
	tag := make([]byte, 16)
	// NLOOP = n, EOP, FLG=PACKED, NREG=1, REGS=A+D.
	lo := n | (1 << 15) | (uint64(gifPacked) << 58) | (uint64(1) << 60)
	putLE64(tag[0:], lo)
	putLE64(tag[8:], gifRegAD)
	out := tag
	for _, w := range writes {
		q := make([]byte, 16)
		putLE64(q[0:], w[1]) // value in the low 64
		putLE64(q[8:], w[0]) // register address in the next byte
		out = append(out, q...)
	}
	return out
}

// buildImageTag builds a GIF IMAGE tag for n quadwords of pixel data.
func buildImageTag(qwords uint64) []byte {
	tag := make([]byte, 16)
	lo := qwords | (1 << 15) | (uint64(gifImage) << 58)
	putLE64(tag[0:], lo)
	return tag
}

func putLE64(b []byte, v uint64) {
	for i := 0; i < 8; i++ {
		b[i] = byte(v >> (8 * i))
	}
}

// bitbltbuf, trxpos and trxreg pack the transfer-setup registers the way the GS reads them.
func bitbltbuf(dbp, dbw, dpsm uint64) uint64 { return dbp<<32 | dbw<<48 | dpsm<<56 }
func trxpos(dsax, dsay uint64) uint64        { return dsax<<32 | dsay<<48 }
func trxreg(w, h uint64) uint64              { return w | h<<32 }

// TestGIFImageUpload uploads a small image through the GIF exactly as the boot does — a
// PACKED setup packet followed by an IMAGE data packet — and checks every pixel lands
// where the PSMCT32 swizzle says it should, so it reads back as the image and not as noise.
func TestGIFImageUpload(t *testing.T) {
	m := NewMachine()

	const w, h = 64, 32 // one PSMCT32 page, to exercise the block and column swizzle
	const dbp, dbw = 0x100, 1

	// A recognisable image: each pixel encodes its own coordinates.
	src := make([]uint32, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			src[y*w+x] = uint32(x) | uint32(y)<<8 | 0xAB0000
		}
	}

	setup := buildADPacket([][2]uint64{
		{gsBITBLTBUF, bitbltbuf(dbp, dbw, psmCT32)},
		{gsTRXPOS, trxpos(0, 0)},
		{gsTRXREG, trxreg(w, h)},
		{gsTRXDIR, 0}, // host -> local
	})
	m.gifPacket(setup)

	image := buildImageTag(uint64(w * h / 4)) // 4 pixels per quadword
	for _, px := range src {
		var b [4]byte
		putLE32(b[:], px)
		image = append(image, b[:]...)
	}
	m.gifPacket(image)

	gs := m.gs
	if gs == nil {
		t.Fatal("no GS was created by the upload")
	}
	if gs.uploads != 1 {
		t.Fatalf("uploads = %d, want 1", gs.uploads)
	}

	// Read every pixel back through the same swizzle and compare.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			addr := addrPSMCT32(dbp, dbw, uint32(x), uint32(y))
			got := le32gs(gs.vram[addr:])
			if got != src[y*w+x] {
				t.Fatalf("pixel (%d,%d): got 0x%06X, want 0x%06X", x, y, got, src[y*w+x])
			}
		}
	}
}

// TestGIFUploadIsInjective checks the PSMCT32 swizzle maps every pixel of a page to a
// distinct address — a swizzle that collided would silently overwrite pixels and read
// back a plausible-looking but wrong image.
func TestGIFUploadIsInjective(t *testing.T) {
	const w, h = 64, 32
	seen := map[uint32]bool{}
	for y := uint32(0); y < h; y++ {
		for x := uint32(0); x < w; x++ {
			a := addrPSMCT32(0, 1, x, y)
			if seen[a] {
				t.Fatalf("pixel (%d,%d) collides at address 0x%X", x, y, a)
			}
			seen[a] = true
		}
	}
	if len(seen) != w*h {
		t.Fatalf("mapped %d distinct addresses, want %d", len(seen), w*h)
	}
}

// TestDMACDrivesGIF checks the whole hardware path the boot uses: the game writes a
// channel-2 transfer's registers and sets STR, and the controller runs the GIF packet
// out of memory and clears STR so the game's poll loop can move on.
func TestDMACDrivesGIF(t *testing.T) {
	m := NewMachine()

	const src = 0x00200000
	packet := buildADPacket([][2]uint64{
		{gsBITBLTBUF, bitbltbuf(0x100, 1, psmCT32)},
		{gsTRXPOS, trxpos(0, 0)},
		{gsTRXREG, trxreg(8, 8)},
		{gsTRXDIR, 0},
	})
	for i, b := range packet {
		m.ram[src+i] = b
	}

	m.Write32(0x1000A010, src)                    // MADR
	m.Write32(0x1000A020, uint32(len(packet)/16)) // QWC in quadwords
	m.Write32(0x1000A000, dChcrStart|dChcrDir)    // CHCR: start, from memory

	if v, _ := m.dmacRead(0x1000A000); v&dChcrStart != 0 {
		t.Fatalf("STR still set after the transfer; CHCR = 0x%08X", v)
	}
	if m.dmacStat&(1<<dmacChGIF) == 0 {
		t.Fatal("D_STAT did not record the GIF channel's completion")
	}
	if m.gs == nil || m.gs.uploads != 1 {
		t.Fatal("the GIF did not process the packet the DMAC handed it")
	}
}

func putLE32(b []byte, v uint32) {
	b[0], b[1], b[2], b[3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
}

// TestDMACChannelDecode pins the non-uniform channel layout. Channels 0-2 are 0x1000
// apart but the rest are packed 0x400 apart, so the old (addr-base)/0x1000 decode folded
// the scratchpad channels (0xD000/0xD400) onto channel 5 and their transfers vanished.
func TestDMACChannelDecode(t *testing.T) {
	cases := []struct {
		addr uint32
		ch   int
		reg  uint32
	}{
		{0x1000A000, 2, dChcr}, // GIF, the one channel the old decode got right
		{0x1000C000, 5, dChcr}, // SIF0 — 0x1000C000/0x1000 would be 4, not 5
		{0x1000C400, 6, dChcr}, // SIF1
		{0x1000D000, 8, dChcr}, // SPR_FROM
		{0x1000D400, 9, dChcr}, // SPR_TO
		{0x1000D410, 9, dMadr}, // a register inside the SPR_TO block
		{0x1000D480, 9, dSadr}, // its scratchpad-address register
	}
	for _, c := range cases {
		ch, reg, ok := dmacChanReg(c.addr)
		if !ok || ch != c.ch || reg != c.reg {
			t.Errorf("dmacChanReg(0x%08X) = ch %d reg 0x%X ok=%v, want ch %d reg 0x%X",
				c.addr, ch, reg, ok, c.ch, c.reg)
		}
	}
	// The controller-wide registers must not decode as a channel.
	if _, _, ok := dmacChanReg(dSTAT); ok {
		t.Errorf("D_STAT decoded as a channel")
	}
}

// TestSPRBounceCopy runs the exact round trip the GOAL runtime's ultimate-memcpy uses:
// memory -> scratchpad on channel 9, then scratchpad -> a different memory address on
// channel 8. Without a real scratchpad mover the destination stays zero, which is what
// dropped whole object main segments on the floor and broke the GOAL linker.
func TestSPRBounceCopy(t *testing.T) {
	m := NewMachine()

	const src, dst = 0x00200000, 0x00280000
	const qwc = 4 // 64 bytes
	for i := 0; i < qwc*16; i++ {
		m.ram[src+i] = byte(0x40 + i)
	}

	// Channel 9 (SPR_TO): main memory at MADR -> scratchpad at SADR.
	m.Write32(0x1000D410, src)        // MADR
	m.Write32(0x1000D480, 0)          // SADR
	m.Write32(0x1000D420, qwc)        // QWC
	m.Write32(0x1000D400, dChcrStart) // CHCR: start

	// Channel 8 (SPR_FROM): scratchpad at SADR -> main memory at MADR.
	m.Write32(0x1000D010, dst)        // MADR
	m.Write32(0x1000D080, 0)          // SADR
	m.Write32(0x1000D020, qwc)        // QWC
	m.Write32(0x1000D000, dChcrStart) // CHCR: start

	for i := 0; i < qwc*16; i++ {
		if got, want := m.ram[dst+i], byte(0x40+i); got != want {
			t.Fatalf("byte %d: got 0x%02X, want 0x%02X (the scratchpad bounce did not move the data)", i, got, want)
		}
	}
	if v, _ := m.dmacRead(0x1000D400); v&dChcrStart != 0 {
		t.Errorf("SPR_TO STR still set after transfer; CHCR = 0x%08X", v)
	}
	if v, _ := m.dmacRead(0x1000D000); v&dChcrStart != 0 {
		t.Errorf("SPR_FROM STR still set after transfer; CHCR = 0x%08X", v)
	}
}

// TestSPRAdvancesMADRAcrossKicks is Ridge Racer V's display-list feeder in miniature: it
// programs MADR once and then walks a contiguous main-memory list by re-arming only
// SADR/QWC/STR each chunk, leaning on the controller to leave MADR past the bytes it just
// moved. The second kick must read the SECOND chunk, not re-read the first — the bug that
// left the feeder's scratchpad stale (zeros with no terminator) and overran its output
// ring. QWC must also drain to zero, since the feeder polls it as "chunk consumed".
func TestSPRAdvancesMADRAcrossKicks(t *testing.T) {
	m := NewMachine()

	const src = 0x00200000
	const qwc = 2 // 32 bytes per chunk
	for i := 0; i < qwc*16*2; i++ {
		m.ram[src+i] = byte(i) // two chunks: 0x00.. then 0x20..
	}

	// Chunk 0: MADR set explicitly, land it at scratchpad 0.
	m.Write32(0x1000D410, src) // MADR
	m.Write32(0x1000D480, 0)   // SADR
	m.Write32(0x1000D420, qwc) // QWC
	m.Write32(0x1000D400, dChcrStart)

	if v, _ := m.dmacRead(0x1000D420); v != 0 {
		t.Errorf("QWC = %d after the transfer, want 0 (the controller drains it)", v)
	}

	// Chunk 1: re-arm SADR/QWC/STR only — MADR is left to auto-advance, exactly as the
	// feeder does. Land it at scratchpad 0x20.
	m.Write32(0x1000D480, 0x20) // SADR
	m.Write32(0x1000D420, qwc)  // QWC
	m.Write32(0x1000D400, dChcrStart)

	for i := 0; i < qwc*16; i++ {
		if got, want := m.spram[0x20+i], byte(qwc*16+i); got != want {
			t.Fatalf("second chunk byte %d: got 0x%02X, want 0x%02X — MADR did not advance, the kick re-read chunk 0",
				i, got, want)
		}
	}
}

// TestSPRChainGather runs the kick the merc renderer performs every frame: SPR_TO in
// source-chain mode with TTE, a CNT header link followed by a REF link gathering data
// from elsewhere in memory. The whole 16-byte tag must land in the scratchpad ahead of
// each link's data — the game rides the buffer's header (its quadword count and the next
// chain segment's address) in the tag's upper 64 bits, and the consumer reads them at
// buffer+8 and buffer+12. Treating the start as a normal transfer moves QWC=0 quadwords,
// and the converter that runs next reads whatever the scratchpad last held.
func TestSPRChainGather(t *testing.T) {
	m := NewMachine()
	put64 := func(a uint32, v uint64) {
		for i := 0; i < 8; i++ {
			m.ram[a+uint32(i)] = byte(v >> (8 * i))
		}
	}

	const chain, ref = 0x00200000, 0x00240000
	const sadr = 0x280 // where the merc double buffer's first half lives

	// The REF link's data, somewhere else entirely.
	for i := 0; i < 32; i++ {
		m.ram[ref+i] = byte(0xA0 + i)
	}

	// Tag 1: CNT, 1 quadword of data follows; upper 64 bits carry the software header.
	put64(chain+0, 1|dtagCNT<<28)
	put64(chain+8, 0xCAFEF00D12345678)
	for i := 0; i < 16; i++ {
		m.ram[chain+16+i] = byte(0x10 + i)
	}
	// Tag 2 (after the CNT data): REF, 2 quadwords at ref, next tag follows.
	put64(chain+32, 2|dtagREF<<28|uint64(ref)<<32)
	put64(chain+40, 0)
	// Tag 3: END.
	put64(chain+48, dtagEND<<28)
	put64(chain+56, 0)

	m.Write32(0x1000D480, sadr)                        // SADR
	m.Write32(0x1000D430, chain)                       // TADR
	m.Write32(0x1000D420, 0)                           // QWC: a chain kick leaves it zero
	m.Write32(0x1000D400, dChcrStart|dChcrTTE|1<<2)    // CHCR: start, chain mode, TTE

	// The scratchpad, in arrival order: tag 1 (16 bytes), its 1 QW of data, tag 2, the
	// REF's 2 QW, tag 3. The header words the consumer reads sit at buffer+8/+12.
	if got := le64(m.spram[sadr+8 : sadr+16]); got != 0xCAFEF00D12345678 {
		t.Errorf("buffer header (tag upper 64) = 0x%016X, want 0xCAFEF00D12345678", got)
	}
	for i := 0; i < 16; i++ {
		if got, want := m.spram[sadr+16+i], byte(0x10+i); got != want {
			t.Fatalf("CNT data byte %d: got 0x%02X, want 0x%02X", i, got, want)
		}
	}
	for i := 0; i < 32; i++ {
		if got, want := m.spram[sadr+48+i], byte(0xA0+i); got != want {
			t.Fatalf("REF data byte %d: got 0x%02X, want 0x%02X (the gather did not follow the REF)", i, got, want)
		}
	}
	if v, _ := m.dmacRead(0x1000D400); v&dChcrStart != 0 {
		t.Errorf("SPR_TO STR still set after the chain; CHCR = 0x%08X", v)
	}
}

// TestSPRInterleaveGather runs the kick bones-mtx-calc performs every frame: SPR_TO in
// interleave mode (CHCR MOD=10) with D_SQWC = 0x00040001 — move 4 quadwords, skip 1 —
// gathering each 80-byte bone record's 4-quadword matrix compactly into the scratchpad.
// QWC counts transferred quadwords only. Treated as a flat copy instead, bone 0's matrix
// lands right and every later bone's is shifted one more quadword into the previous
// record's tail — the merc palette's "entry 0 sane, entries 1+ garbage".
func TestSPRInterleaveGather(t *testing.T) {
	m := NewMachine()

	const src = 0x00200000
	const bones, stride = 3, 80 // records of 4 matrix quadwords + 1 skipped quadword
	for b := 0; b < bones; b++ {
		for i := 0; i < stride; i++ {
			v := byte(0x10*(b+1) + i) // matrix bytes, per bone
			if i >= 64 {
				v = 0xEE // the fifth quadword: the part the DMA must skip
			}
			m.ram[src+b*stride+i] = v
		}
	}

	m.Write32(0x1000E030, 0x00040001)        // D_SQWC: TQWC=4, SQWC=1
	m.Write32(0x1000D410, src)               // MADR
	m.Write32(0x1000D480, 0x100)             // SADR
	m.Write32(0x1000D420, 4*bones)           // QWC: transferred quadwords only
	m.Write32(0x1000D400, dChcrStart|2<<2)   // CHCR: start, interleave mode

	for b := 0; b < bones; b++ {
		for i := 0; i < 64; i++ {
			got := m.spram[0x100+b*64+i]
			want := byte(0x10*(b+1) + i)
			if got != want {
				t.Fatalf("bone %d byte %d: got 0x%02X, want 0x%02X (the interleave gather is off)", b, i, got, want)
			}
		}
	}
	for i := 0; i < bones*64; i++ {
		if m.spram[0x100+i] == 0xEE {
			t.Fatalf("skipped quadword byte leaked into the scratchpad at +0x%X", i)
		}
	}

	// The scatter direction: SPR_FROM with the same pattern spreads compact scratchpad
	// quadwords back out over strided memory, skipping main-memory quadwords.
	const dst = 0x00280000
	for i := 0; i < bones*stride; i++ {
		m.ram[dst+i] = 0xCC
	}
	m.Write32(0x1000D010, dst)             // MADR
	m.Write32(0x1000D080, 0x100)           // SADR
	m.Write32(0x1000D020, 4*bones)         // QWC
	m.Write32(0x1000D000, dChcrStart|2<<2) // CHCR: start, interleave mode
	for b := 0; b < bones; b++ {
		for i := 0; i < 64; i++ {
			if got, want := m.ram[dst+b*stride+i], byte(0x10*(b+1)+i); got != want {
				t.Fatalf("scatter bone %d byte %d: got 0x%02X, want 0x%02X", b, i, got, want)
			}
		}
		for i := 64; i < stride; i++ {
			if got := m.ram[dst+b*stride+i]; got != 0xCC {
				t.Fatalf("scatter bone %d wrote into the skip quadword at +%d (0x%02X)", b, i, got)
			}
		}
	}
}
