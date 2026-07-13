package ps2

// sif.go is the boundary between the two processors, with the far side faked.
//
// A PS2 has a second CPU, the IOP — a MIPS R3000A, the PlayStation's chip — which
// owns the disc drive, the sound chip and the controllers. The EE reaches it across
// the SIF: a DMA path plus a remote-procedure-call protocol, and a set of shared
// registers used as a handshake. The disc ships the IOP's own modules (OVERLORD.IRX,
// 989SND.IRX, PADMAN.IRX) to be uploaded and run there.
//
// None of that runs yet. This file is a stand-in: it answers the EE as the IOP would,
// with Go, so the EE-side boot can be taken to the GOAL main loop without a second
// interpreter in the way. Everything it does is logged, so the shape of the protocol
// the game actually uses is measured rather than guessed — which is what the phase
// that *does* run the IOP will be built from.
//
// The real thing is not far off: tools/cpu/mips already models the R3000A exactly,
// and IRX modules import from the IOP kernel through stub tables of the same shape
// the PSP's PRX modules use, so the trick that HLEs those applies here too.

// sifSetReg / sifGetReg are the shared handshake registers. The kernel keeps 32 of
// them; the boot uses a couple to agree with the IOP that both sides are alive.
func (m *Machine) sifSetReg() {
	reg, val := m.arg(0), m.arg(1)
	m.sifRegs[reg&0x1F] = val
	m.note("SifSetReg %d = 0x%08X", reg&0x1F, val)
	m.setRet(0)
}

func (m *Machine) sifGetReg() {
	reg := m.arg(0) & 0x1F
	v := m.sifRegs[reg]

	// Register 0 is the one the EE polls to learn whether the IOP has finished
	// booting. The IOP is not running, so nothing would ever set it — and the EE's
	// wait loop would spin forever. Reporting it ready is the whole of the fake.
	if reg == 0 && v == 0 {
		v = sifIOPReady
	}
	m.setRet(v)
}

// sifIOPReady is the value the EE's synchronisation loop is waiting to see.
const sifIOPReady = 0x00000001

// sifSetDma queues a transfer to the IOP. There is nothing on the other end, so the
// transfer is recorded and reported complete. A non-zero handle is required: the
// caller passes it to SifDmaStat and treats zero as failure.
func (m *Machine) sifSetDma() {
	desc, count := m.arg(0), m.arg(1)
	for i := uint32(0); i < count; i++ {
		d := desc + i*32
		src, dst, size := m.Read32(d+0x00), m.Read32(d+0x04), m.Read32(d+0x08)
		m.note("SifSetDma: %d bytes from EE 0x%08X to IOP 0x%08X", size, src, dst)
		_ = dst
		_ = src
	}
	m.sifDmaID++
	m.setRet(m.sifDmaID)
}
