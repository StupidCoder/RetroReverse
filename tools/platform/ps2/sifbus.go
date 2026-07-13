package ps2

// sifbus.go is the SIF's register file: the doorbells the two processors ring.
//
// The SIF is three things. This file is the first and smallest of them — six registers
// that both chips can see, used to hand each other a word and to raise a flag. The
// other two are the DMA path that moves bulk data, and the command and RPC layers the
// modules build on top. They come later; nothing can be built on a handshake that does
// not exist.
//
// The registers appear at two addresses, one per processor, and they are the *same*
// registers:
//
//	                    EE            IOP
//	MSCOM   +0x00   0x1000F200    0x1D000000    a word from the EE to the IOP
//	SMCOM   +0x10   0x1000F210    0x1D000010    a word from the IOP to the EE
//	MSFLG   +0x20   0x1000F220    0x1D000020    a flag raised by the EE
//	SMFLG   +0x30   0x1000F230    0x1D000030    a flag raised by the IOP
//	CTRL    +0x40   0x1000F240    0x1D000040
//	BD6     +0x60   0x1000F260    0x1D000060
//
// "Main" is the EE and "sub" is the IOP, which is what the names mean and which way
// round each flag goes.
//
// The one bit that matters today was read out of SIFMAN, on the disc. Its
// initialisation ends in a loop: read MSFLG, mask with 0x00010000, and go round again
// until it is set. That is the IOP waiting to be told the EE's side of the SIF is up —
// and until it is told, nothing else on the IOP happens, because SIFCMD's entry point
// never returns from it.
//
// SIFMAN reads that register in a way worth copying down, because it says exactly what
// kind of thing this is: it reads MSFLG twice and goes round again unless the two reads
// agree. It is guarding against the other processor changing the register underneath it
// mid-read. Nothing in a single-threaded emulator needs that. It is there because on the
// board, these six words are the one place two chips write at once.

// The SBUS registers, by offset.
const (
	sbusMSCOM = 0x00
	sbusSMCOM = 0x10
	sbusMSFLG = 0x20
	sbusSMFLG = 0x30
	sbusCTRL  = 0x40
	sbusBD6   = 0x60

	sbusSpan = 0x80 // the whole window, in bytes
	sbusRegs = sbusSpan / 0x10
)

// Where each processor sees them.
const (
	sbusEEBase  = 0x1000F200
	sbusIOPBase = 0x1D000000
)

// sifEESIFReady is the bit in MSFLG that the IOP's SIFMAN spins on before it will
// finish initialising. It means "the EE's half of the SIF is up".
//
// The value is not a convention taken from anywhere: it is the mask in the instruction
// that tests it, `lui $s1, 0x1` followed by `and $s0, $s0, $s1`, in SIFMAN's own
// initialisation. The module that waits for the bit is the authority on which bit it is.
const sifEESIFReady = 0x00010000

// sbusRead serves a read of the shared registers from either side.
func (m *Machine) sbusRead(off uint32) uint32 {
	if i := off / 0x10; i < sbusRegs {
		return m.sbus[i]
	}
	return 0
}

// sbusWrite serves a write from either side.
func (m *Machine) sbusWrite(off, v uint32) {
	if i := off / 0x10; i < sbusRegs {
		m.sbus[i] = v
	}
}

// sbusSetFlag raises bits in one of the flag registers. The flags are doorbells: a
// processor sets a bit and the other one notices, so the operation that matters is
// "or", not "store".
func (m *Machine) sbusSetFlag(reg, bits uint32) {
	m.sbus[reg/0x10] |= bits
}
