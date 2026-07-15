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
