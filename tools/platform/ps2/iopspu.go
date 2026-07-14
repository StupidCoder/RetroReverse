package ps2

// iopspu.go is the SPU2, modelled as far as the handshake and not one register further.
//
// This is deliberate. The SPU2 is two sound cores with twenty-four voices each, an ADPCM
// decoder, a reverb unit and 2 MiB of its own memory, and none of that is what stands
// between this machine and a booting game. What stands between them is one bit: OVERLORD
// hands the chip a block of sound data and waits to be told the chip has taken it. So
// what is built here is the being-told, and the sound RAM the data lands in, and nothing
// else. When the game needs to *hear* something, that will be a phase with its own name.
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
)

// iopIRQSPU is the sound chip's own interrupt line, as distinct from the DMA channels'.
// LIBSD registers a handler on it, and nothing raises it yet — the transfers this boot
// makes are reported by the DMA controller, on the channel's own number.
const iopIRQSPU = 9

// spu2 is the sound chip.
type spu2 struct {
	regs []byte // the register window, 0x800 bytes, exactly as the chip presents it
	ram  []byte // the sound memory the transfers land in
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
