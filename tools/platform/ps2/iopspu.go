package ps2

// iopspu.go is the SPU2: the transfer handshake, the sound RAM, and — the phase the
// header below promised — a voice's read address moving through that RAM over time.
//
// The SPU2 is two sound cores with twenty-four voices each, an ADPCM decoder, a reverb
// unit and 2 MiB of its own memory, and for a long time none of that was what stood
// between this machine and a booting game. What stood between them was one bit: OVERLORD
// hands the chip a block of sound data and waits to be told the chip has taken it. So the
// first thing built here was the being-told, and the sound RAM the data lands in, and
// nothing else — and the note that closed this header said that when the game needed to
// *hear* something, or to know where the chip had got to, that would be a phase with its
// own name. Jak's intro cutscene is that phase: it streams its animation off the disc and
// keeps time by the read address of the voice playing the audio (sceSdGetAddr). See the
// voice-playback section at the foot of the file; the register model above it is still
// exactly the handshake, unchanged.
//
// The transfer itself is the DMA controller's, on channel 4 for the first core and 7 for
// the second, and so is the interrupt that reports it (iopdma.go). What belongs to the
// chip is where the bytes land — the transfer address, in the register pair at 0x1A8 —
// and a status register at 0x7C2 saying which core has just finished.
//
// That status register is worth writing down, because it is the thing LIBSD reads:
//
//	lhu  $a0, 0xBF9007C2      the chip's status
//	andi $a0, $a0, 0xC        bits 2 and 3: core 0 and core 1
//	srl  $a0, $a0, 2
//	jalr $v0                  the callback the caller registered
//
// So a completed transfer sets bit 2 for core 0 or bit 3 for core 1, and the handler
// hands the shifted pair to the callback as a mask of which cores are done. OVERLORD's
// callback is called `intr` in its own symbol table, and all it does is store a 1 into the
// global that DMA_SendToSPUAndSync is spinning on. That is the whole chain, and every link
// of it is on the disc.

// The SPU2's address window on the IOP's bus. LIBSD points the bus at it itself, writing
// 0xBF900000 into the SSBUS device register at 0x1F801404 — which is the chip telling us
// where it lives.
const (
	iopSPU2Base = 0x1F900000
	iopSPU2End  = 0x1F900800

	iopSPU2RAMSize = 2 << 20 // the sound memory, 2 MiB

	// The two cores' register banks, and the block of chip-wide registers above them.
	iopSPU2CoreSpan = 0x400

	// The registers this model actually has an opinion about. Everything else in the
	// window is stored and read back, which is what a register nobody has identified
	// should do.
	iopSPU2TSAHi = 0x1A8 // transfer start address, high half
	iopSPU2TSALo = 0x1AA // and low half
	iopSPU2Stat  = 0x7C2 // chip-wide: which core's transfer just finished

	// A core's own status, one per bank. Bit 7 is the one that matters, and LIBSD says
	// which and why: having been told by the DMA controller that the transfer is over, its
	// handler sits in a timed loop reading this register and will not go on until the bit
	// is set.
	//
	//	lhu  $v0, 0($a0)        $a0 = 0xBF900344 + core*0x400
	//	andi $v0, $v0, 0x80
	//	beq  $v0, $zero, loop
	//
	// So the controller finishing is not the same event as the chip being ready, and the
	// driver knows it. The bus hands over the last word and the sound chip still has to
	// take it in; bit 7 is the chip saying it has. A model that raises the interrupt and
	// leaves this bit clear gives the driver a completion it cannot believe, and it spins
	// out its timeout on every single transfer.
	iopSPU2CoreStat = 0x344
	iopSPU2CoreDone = 1 << 7

	// The registers a *playing* voice is made of, and the two that start and stop it.
	// These are not guesses: LIBSD builds each one's address from a table it fills at
	// init (at IOP 0x4C080, indexed by the selector's top byte), and the table is what
	// says which chip register a given sceSd* call reaches. sceSdGetAddr(0x2200|voice) —
	// which is the whole reason this file grew a playback position — walks to 0x1C8, and
	// the neighbours below come from the same table.
	//
	//	selector 0x02 -> 0x004  PITCH   (per voice, stride 0x10)
	//	selector 0x15 -> 0x1A0  KON     (key on:  bit v of the pair at 0x1A0/0x1A2)
	//	selector 0x16 -> 0x1A4  KOFF    (key off: the same, at 0x1A4/0x1A6)
	//	selector 0x20 -> 0x1C0  SSA     (start address, per voice, stride 0x0C)
	//	selector 0x21 -> 0x1C4  LSAX    (loop/repeat address)
	//	selector 0x22 -> 0x1C8  NAX     (the current read address the game reads)
	iopSPU2Pitch = 0x004 // voice v at v*iopSPU2VoiceStride + this
	iopSPU2KON   = 0x1A0
	iopSPU2KOFF  = 0x1A4
	iopSPU2SSA   = 0x1C0 // voice v at v*iopSPU2AddrStride + this
	iopSPU2LSAX  = 0x1C4
	iopSPU2NAX   = 0x1C8

	iopSPU2VoiceStride = 0x10 // the eight per-voice parameter registers
	iopSPU2AddrStride  = 0x0C // the three per-voice address registers
	iopSPU2Voices      = 24
	iopSPU2Cores       = 2

	// An SPU2 address counts 16-bit words; an ADPCM block is sixteen bytes, so eight
	// words, and holds twenty-eight decoded samples. The read address steps one block at
	// a time as the block's samples are played out.
	iopSPU2BlockWords   = 8
	iopSPU2BlockSamples = 28
)

// spuAccPerBlock is how much of a voice's playback accumulator makes one ADPCM block.
//
// The accumulator sums (IOP steps elapsed × the voice's PITCH), and PITCH is a 4.12
// fixed point in which 0x1000 is "one source sample per output sample". So the units are
// output-samples × 0x1000, per IOP step, and a block is reached when enough of them have
// gone by to play twenty-eight source samples.
//
// The conversion from IOP steps to output samples is this machine's, not silicon's. The
// SPU2 runs at 48 kHz; a field is 1/60 s and this model paces one at stepsPerVBlank EE
// instructions, of which the IOP retires one in iopStepRatio. So a field is
// stepsPerVBlank/iopStepRatio IOP steps and carries 48000/60 = 800 samples, and the audio
// consumes — and the read address advances — at exactly the rate the game's own frame
// clock ticks against, which is the rate it wrote its stream to be seeked by. Tying this
// to silicon's 36.864 MHz instead would play the cutscene at a fifth of its speed, because
// this model's field is a fifth of the instructions a real one is.
const (
	spuSampleRate       = 48000
	spuFieldRate        = 60
	spuIOPStepsPerField = stepsPerVBlank / iopStepRatio
	spuAccPerBlock      = spuIOPStepsPerField * iopSPU2BlockSamples * 0x1000 / (spuSampleRate / spuFieldRate)

	// A safety bound on how many blocks one advance may step. The voice whose position
	// the game reads is read most frames, so its gaps are a field or two and never near
	// this; a voice that is playing but never read (a looping effect nobody asks about)
	// is what this stops from spinning without limit when its accumulator is stale.
	spuMaxBlocksPerAdvance = 1 << 14
)

// iopIRQSPU is the sound chip's own interrupt line, as distinct from the DMA channels'.
// LIBSD registers a handler on it, and nothing raises it yet — the transfers this boot
// makes are reported by the DMA controller, on the channel's own number.
const iopIRQSPU = 9

// spuVoice is a voice's playback, which the register file alone cannot express: a read
// address moving through the sound memory over time. Everything a reader needs — the
// current position, the loop point — lives in the registers; what lives here is the
// clockwork that advances the position between reads.
type spuVoice struct {
	playing  bool   // keyed on, and not yet run into a block that ends it
	acc      uint64 // sum of (IOP steps × PITCH) not yet spent on a block; see spuAccPerBlock
	lastStep uint64 // the step count acc was last carried up to
}

// spu2 is the sound chip.
type spu2 struct {
	regs  []byte                                // the register window, 0x800 bytes, exactly as the chip presents it
	ram   []byte                                // the sound memory the transfers land in
	voice [iopSPU2Cores][iopSPU2Voices]spuVoice // the playback state the registers cannot hold
}

func newSPU2() *spu2 {
	return &spu2{
		regs: make([]byte, iopSPU2End-iopSPU2Base),
		ram:  make([]byte, iopSPU2RAMSize),
	}
}

// read and write serve the register window. They are word-wide because the IOP's bus is,
// but the chip's registers are halfwords — so one access here is two of the chip's, and
// the code that reads 0x7C2 with an `lhu` is served by the word at 0x7C0.
func (s *spu2) read(off uint32) uint32 {
	if off+4 > uint32(len(s.regs)) {
		return 0
	}
	return uint32(s.regs[off]) | uint32(s.regs[off+1])<<8 |
		uint32(s.regs[off+2])<<16 | uint32(s.regs[off+3])<<24
}

func (s *spu2) write(off, v uint32) {
	if off+4 > uint32(len(s.regs)) {
		return
	}
	s.regs[off] = byte(v)
	s.regs[off+1] = byte(v >> 8)
	s.regs[off+2] = byte(v >> 16)
	s.regs[off+3] = byte(v >> 24)
}

// half reads one of the chip's 16-bit registers.
func (s *spu2) half(off uint32) uint32 {
	return uint32(s.regs[off]) | uint32(s.regs[off+1])<<8
}

func (s *spu2) setHalf(off, v uint32) {
	s.regs[off] = byte(v)
	s.regs[off+1] = byte(v >> 8)
}

// tsa is where in the sound memory a core's next transfer will land.
//
// The pair is stored high half first, and that is settled by arithmetic rather than by
// assumption: LIBSD writes the word 0x88200000 across 0x1A8, which puts 0x0000 in the
// register at 0x1A8 and 0x8820 in the one at 0x1AA. Reading them as (0x1A8 << 16 | 0x1AA)
// gives 0x8820 — a halfword offset a little way into a 2 MiB memory. Reading them the
// other way round gives 0x88200000, which is sixty-four times larger than the memory it
// is supposed to be an address in. Only one of those is an address.
func (s *spu2) tsa(core int) uint32 {
	base := uint32(core) * iopSPU2CoreSpan
	return (s.half(base+iopSPU2TSAHi)<<16 | s.half(base+iopSPU2TSALo)) & 0x000FFFFF
}

// dma moves a block between IOP memory and the sound memory.
//
// The addressing unit is the chip's, not the bus's: an SPU2 address counts 16-bit words,
// so the byte offset is twice the transfer address. Nothing on this disc reads the sound
// memory back, so nothing here has confirmed that — what *is* confirmed is the handshake,
// which is the part the boot is waiting on. The data is moved rather than dropped because
// dropping it would be a lie that costs nothing to avoid.
func (s *spu2) dma(core int, p *IOP, madr, n uint32, toRAM bool) {
	// Clear the two "done" bits this transfer will set — the chip-wide one that says which
	// core finished, and the core's own. A stale bit is a completion the driver has already
	// been told about being offered to it a second time.
	base := uint32(core) * iopSPU2CoreSpan
	s.setHalf(iopSPU2Stat, s.half(iopSPU2Stat)&^(1<<(2+uint(core))))
	s.setHalf(base+iopSPU2CoreStat, s.half(base+iopSPU2CoreStat)&^iopSPU2CoreDone)

	off := s.tsa(core) * 2
	for i := uint32(0); i < n; i++ {
		if off+i >= uint32(len(s.ram)) {
			break
		}
		if toRAM {
			p.Write(madr+i, s.ram[off+i])
		} else {
			s.ram[off+i] = p.Read(madr + i)
		}
	}
	p.ps2.note("IOP SPU2: core %d took %d bytes from 0x%08X into sound memory at 0x%05X",
		core, n, madr, off)
}

// complete is the chip saying it has taken the data: it sets the core's bit in its status
// register. The *interrupt* is the DMA controller's to raise, on the channel that fed the
// chip — which is what LIBSD registered a handler against (iopdma.go).
//
// The status bit is set anyway, and not because anything on the boot path reads it. LIBSD
// has a second handler, on the sound chip's own line rather than the DMA channel's, and
// that one does read this register — it is the path a transfer takes when the chip
// interrupts on its own account rather than the controller doing it. Nothing raises that
// line yet. Setting the bit costs nothing and means that when something does, the register
// it finds is telling the truth.
func (s *spu2) complete(core int, p *IOP) {
	base := uint32(core) * iopSPU2CoreSpan
	s.setHalf(iopSPU2Stat, s.half(iopSPU2Stat)|1<<(2+uint(core)))
	s.setHalf(base+iopSPU2CoreStat, s.half(base+iopSPU2CoreStat)|iopSPU2CoreDone)
}

// --- voice playback -------------------------------------------------------------
//
// This is the phase the file's own header said would come — the one where the game
// needs to *hear* something, or at least to know where the sound chip has got to. Jak's
// intro cutscene streams its animation off the DVD and keeps time by asking the chip, once
// a frame, for the current read address of the voice playing the stream (via sceSdGetAddr,
// which is why NAX at 0x1C8 is the register that matters). A register file that never
// moves that address answers the same position for ever, and the animation freezes on one
// frame while the whole streaming ring, prefetching against a position that will not
// advance, stalls behind it. So the address has to move, the way it moves on silicon: a
// keyed-on voice reads through the sound memory a block at a time, at a rate its PITCH
// sets, looping where the ADPCM data it is playing says to loop.

// readReg answers a guest read of the register window, having first carried every playing
// voice up to now — so a read of NAX sees where the address has reached.
func (s *spu2) readReg(off uint32, now uint64) uint32 {
	s.tick(now)
	return s.read(off)
}

// writeReg serves a guest write, and acts on the two registers that start and stop a
// voice. Playing voices are carried up to now first, so a KON/KOFF lands at the right
// position and the accumulator does not jump when the register that gates it changes.
func (s *spu2) writeReg(off, v uint32, now uint64) {
	s.tick(now)
	s.write(off, v)

	core := int(off / iopSPU2CoreSpan)
	if core >= iopSPU2Cores {
		return
	}
	switch off % iopSPU2CoreSpan {
	case iopSPU2KON:
		s.keyOnOff(core, iopSPU2KON, now, true)
	case iopSPU2KOFF:
		s.keyOnOff(core, iopSPU2KOFF, now, false)
	}
}

// keyOnOff processes a write to a core's KON or KOFF register: a bitmask of voices, low
// sixteen in the halfword at r and the top eight in the one above. Keying a voice on puts
// its read address at its start address and starts the clock; keying it off stops it. The
// register is a trigger — the hardware clears it once it has acted, and so does this, so a
// later write does not act on a bit that has already been handled.
func (s *spu2) keyOnOff(core int, r uint32, now uint64, on bool) {
	base := uint32(core) * iopSPU2CoreSpan
	mask := s.half(base+r) | s.half(base+r+2)<<16
	for v := 0; v < iopSPU2Voices; v++ {
		if mask&(1<<uint(v)) == 0 {
			continue
		}
		vs := &s.voice[core][v]
		if on {
			ab := base + uint32(v)*iopSPU2AddrStride
			s.setWordAddr(ab+iopSPU2NAX, s.wordAddr(ab+iopSPU2SSA))
			vs.playing = true
			vs.acc = 0
			vs.lastStep = now
		} else {
			vs.playing = false
		}
	}
	s.setHalf(base+r, 0)
	s.setHalf(base+r+2, 0)
}

// tick carries every voice up to now.
func (s *spu2) tick(now uint64) {
	for core := 0; core < iopSPU2Cores; core++ {
		for v := 0; v < iopSPU2Voices; v++ {
			s.advance(core, v, now)
		}
	}
}

// advance moves one voice's read address forward by the blocks the time since it was last
// carried has played out, following the loop that the ADPCM data itself encodes.
func (s *spu2) advance(core, v int, now uint64) {
	vs := &s.voice[core][v]
	if !vs.playing {
		vs.lastStep = now
		return
	}
	if now <= vs.lastStep {
		return
	}
	pitch := uint64(s.half(uint32(core)*iopSPU2CoreSpan+uint32(v)*iopSPU2VoiceStride+iopSPU2Pitch) & 0x3FFF)
	vs.acc += (now - vs.lastStep) * pitch
	vs.lastStep = now
	if pitch == 0 { // a paused voice: time passes, the address does not
		return
	}

	ab := uint32(core)*iopSPU2CoreSpan + uint32(v)*iopSPU2AddrStride
	naxOff, lsaxOff := ab+iopSPU2NAX, ab+iopSPU2LSAX
	for n := 0; vs.acc >= spuAccPerBlock && n < spuMaxBlocksPerAdvance; n++ {
		vs.acc -= spuAccPerBlock

		nax := s.wordAddr(naxOff)
		// The ADPCM flag byte is the second byte of the sixteen-byte block. Bit 2 marks a
		// block as the loop point (the repeat address becomes this block); bit 0 ends the
		// loop body (the address jumps back to the repeat address); bit 1, when the loop
		// ends, is what tells a stream to keep going rather than stop.
		flag := s.ram[(nax*2+1)&uint32(len(s.ram)-1)]
		if flag&0x04 != 0 {
			s.setWordAddr(lsaxOff, nax)
		}
		if flag&0x01 != 0 {
			s.setWordAddr(naxOff, s.wordAddr(lsaxOff))
			if flag&0x02 == 0 {
				vs.playing = false
				break
			}
		} else {
			s.setWordAddr(naxOff, nax+iopSPU2BlockWords)
		}
	}
	if vs.acc >= spuAccPerBlock { // hit the per-advance bound: drop the backlog rather than spiral
		vs.acc %= spuAccPerBlock
	}
}

// wordAddr reads one of the paired address registers (high half at off, low half at
// off+2) as the sound-memory word address LIBSD reconstructs from them: sceSdGetAddr
// returns (high<<17)|(low<<1) bytes, i.e. this word address doubled.
func (s *spu2) wordAddr(off uint32) uint32 {
	return (s.half(off)<<16 | s.half(off+2)) & 0xFFFFF
}

func (s *spu2) setWordAddr(off, w uint32) {
	s.setHalf(off, (w>>16)&0xFFFF)
	s.setHalf(off+2, w&0xFFFF)
}

// saveVoices and loadVoices carry the playback clocks across a snapshot. The layout is
// flat, core*iopSPU2Voices + voice; a state saved before this existed loads nil, which
// leaves every voice stopped — correct for a state from before any stream played.
func (s *spu2) saveVoices() []SPUVoiceState {
	out := make([]SPUVoiceState, 0, iopSPU2Cores*iopSPU2Voices)
	for core := 0; core < iopSPU2Cores; core++ {
		for v := 0; v < iopSPU2Voices; v++ {
			vs := s.voice[core][v]
			out = append(out, SPUVoiceState{Playing: vs.playing, Acc: vs.acc, LastStep: vs.lastStep})
		}
	}
	return out
}

func (s *spu2) loadVoices(in []SPUVoiceState) {
	for core := 0; core < iopSPU2Cores; core++ {
		for v := 0; v < iopSPU2Voices; v++ {
			i := core*iopSPU2Voices + v
			if i < len(in) {
				s.voice[core][v] = spuVoice{playing: in[i].Playing, acc: in[i].Acc, lastStep: in[i].LastStep}
			} else {
				s.voice[core][v] = spuVoice{}
			}
		}
	}
}
