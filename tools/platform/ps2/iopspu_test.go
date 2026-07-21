package ps2

// iopspu_test.go pins the SPU2 voice playback — the part that moves a voice's read
// address through the sound memory over time. It is the mechanism Jak's intro cutscene
// keeps time by: the game seeks its DVD-streamed animation to the position sceSdGetAddr
// reports, and a read address that never advances freezes the whole cutscene on one frame.

import "testing"

// spuStepsForBlocks is how many IOP steps play out n ADPCM blocks at unity pitch (0x1000).
func spuStepsForBlocks(n int) uint64 { return uint64(n) * spuAccPerBlock / 0x1000 }

// setupVoice programs one voice's start/loop/pitch and lays down nblocks of ADPCM in the
// sound memory, each block's flag byte taken from flags[b].
func setupVoice(s *spu2, core, v, start int, flags []byte) uint32 {
	ab := uint32(core)*iopSPU2CoreSpan + uint32(v)*iopSPU2AddrStride
	s.setWordAddr(ab+iopSPU2SSA, uint32(start))
	s.setWordAddr(ab+iopSPU2LSAX, uint32(start))
	s.setHalf(uint32(core)*iopSPU2CoreSpan+uint32(v)*iopSPU2VoiceStride+iopSPU2Pitch, 0x1000)
	for b, flag := range flags {
		blk := (uint32(start) + uint32(b)*iopSPU2BlockWords) * 2 // byte address of the block
		s.ram[blk+1] = flag                                      // the flag is the block's second byte
	}
	return ab
}

// keyOn keys a voice on the way a guest write to KON does.
func keyOn(s *spu2, core, v int, now uint64) {
	s.setHalf(uint32(core)*iopSPU2CoreSpan+iopSPU2KON, 1<<uint(v))
	s.keyOnOff(core, iopSPU2KON, now, true)
}

func TestSPUVoiceKeyOnSeeksToStart(t *testing.T) {
	s := newSPU2()
	ab := setupVoice(s, 0, 5, 0x1000, make([]byte, 8))
	// Leave a stale NAX behind, as a register file that has been used before would.
	s.setWordAddr(ab+iopSPU2NAX, 0xBEEF)
	keyOn(s, 0, 5, 0)
	if got := s.wordAddr(ab + iopSPU2NAX); got != 0x1000 {
		t.Fatalf("key-on must seek the read address to the start address 0x1000, got %#x", got)
	}
}

func TestSPUVoiceAdvancesWithTime(t *testing.T) {
	s := newSPU2()
	ab := setupVoice(s, 0, 5, 0x1000, make([]byte, 16)) // all-normal blocks
	keyOn(s, 0, 5, 0)

	// Reading the address carries it up to the current step, exactly as sceSdGetAddr does.
	naxOff := ab + iopSPU2NAX
	s.tick(spuStepsForBlocks(3))
	if got := s.wordAddr(naxOff); got != 0x1000+3*iopSPU2BlockWords {
		t.Fatalf("after three blocks the address should be start+3 blocks (%#x), got %#x",
			uint32(0x1000+3*iopSPU2BlockWords), got)
	}

	// The libsd reconstruction of the byte address is the word address doubled — the value
	// GetPlayPos subtracts the buffer base from.
	if byteAddr := s.wordAddr(naxOff) * 2; byteAddr != (0x1000+3*iopSPU2BlockWords)*2 {
		t.Fatalf("byte address should be the word address doubled, got %#x", byteAddr)
	}
}

func TestSPUVoiceLoopsAtLoopEnd(t *testing.T) {
	s := newSPU2()
	// Eight blocks; the last carries loop-end + repeat (0x03), so the address wraps to the
	// loop address and playback continues — a streaming buffer's continuous loop.
	flags := make([]byte, 8)
	flags[7] = 0x03
	ab := setupVoice(s, 0, 5, 0x1000, flags)
	keyOn(s, 0, 5, 0)

	s.tick(spuStepsForBlocks(8)) // consume all eight blocks: block 7 wraps to LSAX = start
	if got := s.wordAddr(ab + iopSPU2NAX); got != 0x1000 {
		t.Fatalf("loop-end+repeat must wrap the address to the loop address 0x1000, got %#x", got)
	}
	if !s.voice[0][5].playing {
		t.Fatalf("loop-end WITH repeat must keep the voice playing")
	}
}

func TestSPUVoiceStopsAtOneShotEnd(t *testing.T) {
	s := newSPU2()
	// Loop-end WITHOUT repeat (0x01): the voice jumps to the loop address and stops.
	flags := make([]byte, 8)
	flags[3] = 0x01
	ab := setupVoice(s, 0, 5, 0x1000, flags)
	keyOn(s, 0, 5, 0)

	s.tick(spuStepsForBlocks(8)) // would be eight blocks, but block 3 ends it
	if s.voice[0][5].playing {
		t.Fatalf("loop-end without repeat must stop the voice")
	}
	// And having stopped, more time does not move it.
	naxAfterStop := s.wordAddr(ab + iopSPU2NAX)
	s.tick(spuStepsForBlocks(20))
	if got := s.wordAddr(ab + iopSPU2NAX); got != naxAfterStop {
		t.Fatalf("a stopped voice must not advance: was %#x, now %#x", naxAfterStop, got)
	}
}

func TestSPUVoiceKeyOffStops(t *testing.T) {
	s := newSPU2()
	ab := setupVoice(s, 0, 5, 0x1000, make([]byte, 32))
	keyOn(s, 0, 5, 0)
	s.tick(spuStepsForBlocks(2))
	moved := s.wordAddr(ab + iopSPU2NAX)

	s.setHalf(iopSPU2KOFF, 1<<5)
	s.keyOnOff(0, iopSPU2KOFF, spuStepsForBlocks(2), false)
	s.tick(spuStepsForBlocks(20))
	if got := s.wordAddr(ab + iopSPU2NAX); got != moved {
		t.Fatalf("a keyed-off voice must not advance: was %#x, now %#x", moved, got)
	}
}

func TestSPUVoiceReadIsStableAcrossCloseReads(t *testing.T) {
	// GetPlayPos reads the address three times a few instructions apart and demands two of
	// them agree; a per-read advance that moved the address every time would spin it. Reads
	// closer together than one block must return the same value.
	s := newSPU2()
	ab := setupVoice(s, 0, 5, 0x1000, make([]byte, 64))
	keyOn(s, 0, 5, 0)

	base := spuStepsForBlocks(10)
	a := s.readReg(ab+iopSPU2NAX, base)
	b := s.readReg(ab+iopSPU2NAX, base+40) // ~40 steps later, far less than a block
	c := s.readReg(ab+iopSPU2NAX, base+80)
	if a != b || b != c {
		t.Fatalf("three reads within one block must agree: %#x %#x %#x", a, b, c)
	}
}
