package ps2

import (
	"sort"
	"strings"
)

// sif.go is the boundary between the two processors — and it is now only a boundary. Nothing
// in this file answers the EE any more; it carries what the EE says to the IOP, and what the
// IOP says back, and counts both on the way past.
//
// A PS2 has a second CPU, the IOP — a MIPS R3000A, the PlayStation's chip — which owns the disc
// drive, the sound chip and the controllers. The EE reaches it across the SIF: two DMA channels
// carrying command packets, with a remote-procedure-call layer built on top of them. This file
// used to fake the far side, because there was no far side. There is now (iop.go): the IOP runs
// Sony's own kernel modules off the disc, and the servers the EE binds to are those modules,
// running. The handles it gets back are addresses in their memory.
//
// A command packet is:
//
//	+0x00  psize in the low byte, dsize in the upper three
//	+0x04  dest        where the data half should land
//	+0x08  cid         the command; bit 31 selects the system handler table
//	+0x0C  opt
//	+0x10  payload
//
// and both processors have a handler that reads exactly that and dispatches on the cid — the
// EE's is _sceSifCmdIntrHdlr on DMA channel 5, the IOP's is SIFCMD+0x6AC on interrupt 43. They
// are the same routine written twice, which is the clue that the protocol is symmetric.
//
// The two directions are not, though, and the asymmetry is the thing worth knowing:
//
//	SIF1   EE -> IOP   the IOP's channel 10. SIFMAN arms it — a start bit, a block size, no
//	                   MADR — and it stays armed until something lands. The *sender* names the
//	                   destination, in the DMA descriptor. There is one receive slot, so there
//	                   is flow control: SIFCMD empties the slot and re-arms (sifPump).
//
//	SIF0   IOP -> EE   the IOP's channel 9, and a chain rather than a burst. SIFMAN writes TADR,
//	                   never MADR, and the tag carries the source, the length, the EE address to
//	                   put it at, and a DMA tag for the EE's own controller (sifFromIOP).
//
// Neither channel's MADR is ever written by the module that starts it, and for a while that
// looked like two separate puzzles. It is one fact: on this bus the sender always says where the
// data goes, and the receiver is a channel that is merely listening.

// The system commands, indexed by the low bits of the cid.
const (
	sifCmdChangeSaddr = 0x80000000 // the EE tells the IOP where its command buffer is
	sifCmdSetSreg     = 0x80000001 // the IOP writes one of the EE's SIF registers
	sifCmdInitCmd     = 0x80000002 // the EE brings the command and RPC layers up
	sifCmdReset       = 0x80000003

	sifCmdRpcEnd   = 0x80000008 // the IOP: your request is finished
	sifCmdRpcBind  = 0x80000009 // the EE: give me a handle for this server
	sifCmdRpcCall  = 0x8000000A // the EE: run this function on that server
	sifCmdRpcRData = 0x8000000C
)

// sifSetDma is syscall 119: the EE handing buffers to the IOP. Each descriptor is
// {src, dest, size, attr}, and every one of them is a transfer.
//
// That last clause is the whole of it, and it was not always true here. A remote call carries
// two descriptors — the arguments, and then the command packet that says a call has been made —
// and this routine used to move only the second. It could afford to: the thing serving the call
// was a Go function on this side of the wire, which could read the arguments out of EE memory
// where they lay. The IOP cannot. It is a separate processor with its own two megabytes, and a
// pointer into the EE's memory means nothing to it at all.
//
// So an argument buffer that is never transferred arrives on the IOP as whatever was in that
// memory before, and FILEIO opens a file called "". The symptom is quiet and a long way from
// the cause: the game's threads sit on a semaphore that the reply to a file operation would
// have signalled, and the file operation is a real one, correctly dispatched, on a path made
// entirely of nothing.
//
// The two are told apart by attr, and the command bit is set only on the packet. Both are
// queued, in the order the EE listed them, because the arguments must be in IOP memory before
// SIFCMD's handler dispatches the command that refers to them — which on the wire they are,
// necessarily, since the EE sends them first.
func (m *Machine) sifSetDma() {
	desc, count := m.arg(0), m.arg(1)
	for i := uint32(0); i < count; i++ {
		d := desc + i*16
		src := m.Read32(d + 0x00)
		dest := m.Read32(d + 0x04)
		size := m.Read32(d + 0x08)
		attr := m.Read32(d + 0x0C)

		if attr&sifDmaInt != 0 {
			m.iopReceive(src, dest, size)
			continue
		}
		m.sifData(src, dest, size)
	}
	// A handle of zero means failure, and the caller checks. It must never be zero.
	m.sifDmaID++
	m.setRet(m.sifDmaID)
}

// sifDmaInt is the bit in a descriptor's attr that says "and interrupt the IOP when this
// lands". Only the command packet of a remote call has it; the arguments that precede it do
// not, and must not — an interrupt per buffer would have SIFCMD dispatching a command out of
// somebody's file path.
const sifDmaInt = 0x40

// sifData queues a plain transfer: bytes from the EE into the IOP's memory, at the address the
// EE names, with no doorbell at the end. This is what a remote call's arguments are.
func (m *Machine) sifData(src, dest, size uint32) {
	if dest == 0 || size == 0 {
		return
	}
	p := sifPacket{dest: dest, data: make([]byte, size)}
	for i := uint32(0); i < size; i++ {
		p.data[i] = m.Read(src + i)
	}
	m.sifToIOPQueue = append(m.sifToIOPQueue, p)
	m.sifPump()
}

// iopReceive is the IOP's side: one command packet arrives from the EE.
func (m *Machine) iopReceive(src, dest, size uint32) {
	if size < 16 {
		return
	}
	cid := m.Read32(src + 0x08)
	opt := m.Read32(src + 0x0C)

	// The reset is the one command the second processor cannot serve, because it is the
	// command that creates it. Everything after this can go to the real IOP.
	if cid == sifCmdReset {
		m.iopReset(src)
		return
	}

	// Two commands carry the EE's own command buffer, in the same place — at +0x10 — and the
	// machine has to note it as it goes past. Not to answer with: to *route* with. It is where
	// a SIF0 transfer out of the IOP is going to land, and the IOP is about to make one.
	//
	// Which of the two the EE sends says whether it has met this IOP before. CHANGE_SADDR means
	// it has, and is only moving its buffer; INIT_CMD with opt = 0 means it has not, and is
	// introducing itself — and that one is the packet the whole second processor is waiting for
	// (see sifGetReg).
	//
	// INIT_CMD is sent a *second* time, with opt = 1, to say that the EE's RPC layer is up. That
	// one is sixteen bytes: a header and nothing else. Reading a buffer address out of it reads
	// off the end of the packet, and what is off the end of the packet is zero — so the EE's
	// address gets replaced by nothing at the exact moment the IOP is about to use it. The
	// condition below is not defensive; it is what SIFCMD's own handler does, which takes the
	// address in its opt == 0 branch and not in its opt != 0 one.
	if cid == sifCmdChangeSaddr || (cid == sifCmdInitCmd && opt == 0) {
		m.sifCmdBuf = m.Read32(src + 0x10)
	}

	// And the packet goes to the second processor. The bytes land in SIFCMD's own receive
	// buffer at the address SIFCMD published in SMCOM, interrupt 43 is raised, and SIFCMD's
	// handler runs on the IOP — on the IOP's stack, in the IOP's own interrupt path.
	//
	// There is no other branch. There used to be: a switch, below this, that answered each
	// command in Go — bound a synthetic handle for every RPC server the EE asked for, and told
	// it the version numbers it wanted to hear. It is gone. Every one of those answers is now
	// given by the module that owns it, running.
	m.sifSent[cid]++
	m.sifWatch(src, cid, true)
	m.sifToIOP(src, dest, size)
}

// iopReset is the EE rebooting the second processor — and it is where the IOP on this machine
// actually comes to life.
//
// The request arrives as an ordinary SIF command packet, cid 0x80000003, sent by the game's own
// sceSifResetIop. That routine is worth following, because it says what a reboot *is* here: it
// calls sceSifSetDma, and nothing else. There is no special register and no back door. The EE
// asks the IOP to reboot by sending it a message, exactly as it asks it for anything else, and
// the message carries the request as a string:
//
//	rom0:UDNL cdrom0:\DRIVERS\IOPRP221.IMG;1
//
// So the image the second processor boots is not a constant this machine keeps; it is a
// sentence the game says. The path is read out of the packet (iopRebootImage) and the disc is
// asked for that file.
//
// Afterwards the EE spins in sceSifSyncIop, which reads SIF register 4 and tests bit 0x40000 —
// "the IOP has finished rebooting". It used to be answered with a lie, because there was
// nothing to reboot. Now it is answered with the truth.
func (m *Machine) iopReset(src uint32) {
	// The payload is a length, a mode, and then the command itself.
	cmd := m.CString(src + 0x18)

	image, err := iopRebootImage(cmd)
	if err != nil {
		m.note("SIF: %v", err)
		return
	}
	m.note("SIF: the EE is rebooting the IOP — %q, so the image is %s", cmd, image)

	// THE REBOOT DOES NOT HAPPEN NOW. It is asked for now, and it happens later, and the
	// difference is the whole of sceSifResetIop.
	//
	// This is reached from inside sceSifSetDma — the EE handing a packet to a DMA channel.
	// On the board that is all it is: the packet is queued, the call returns, and the IOP has
	// not yet so much as looked at it. sceSifResetIop then does the last thing it has to do,
	// which is to CLEAR SMFLG's 0x10000 and 0x20000 — the old IOP's "I am up" bits — so that
	// the new one can raise them, and only then does it start waiting.
	//
	// Reboot here, synchronously, and the order inverts. The fresh IOP boots inside the store,
	// raises all three of its flags, and *then* sceSifResetIop's two clears arrive and wipe
	// them; sceSifSyncIop's acknowledgement takes the third. SMFLG comes back to zero and the
	// EE waits forever for a reboot that had already finished — nine million times round a
	// four-instruction loop, on a machine where everything worked.
	//
	// So it is queued, and the machine performs it a little later (run.go). How much later
	// does not matter, and the game says why: it will not touch the IOP again until the IOP
	// tells it the reboot is done. A processor that is allowed to take as long as it likes is
	// a processor whose boot cannot race anything.
	m.iopRebootImage = image
	m.iopRebootAt = m.steps + iopRebootLatency
}

// iopRebootLatency is how long the second processor takes to come back, in EE instructions.
//
// A real one takes on the order of a hundred milliseconds — it is a cold boot of a whole
// processor. This is far shorter than that and enormously longer than it needs to be, and
// both halves of that are deliberate: it has only to outlast the thirty-odd instructions
// sceSifResetIop has left to run, and nothing else in the machine is entitled to care,
// because the game is blocked on the IOP's own word that it is ready.
const iopRebootLatency = 100000

// --- the transport ------------------------------------------------------------------
//
// Two directions, two DMA channels, and one buffer that both processors can see. Everything
// below was read off the modules rather than assumed, and SIFMAN is the authority on all of it
// (the trace is in sifbus.go and iopdma.go):
//
//	SIF1   EE -> IOP   the EE's DMA channel 6; the IOP's channel 10, interrupt 43.
//	SIF0   IOP -> EE   the IOP's DMA channel 9; the EE's channel 5, whose handler the EE
//	                   registers by name — _sceSifCmdIntrHdlr.
//
// The destination of a SIF1 transfer does not come from the IOP's side. SIFMAN arms channel 10
// with a block size and a start bit and *never writes MADR* — so the receiving end does not
// choose where the data lands; the sender does. What it does instead is publish the address:
// it writes SIFCMD's receive buffer into SMCOM, the word the IOP hands the EE, and raises
// SMFLG's bit 0x00010000 to say its half is up. So the address on the envelope is the IOP's
// own, and this reads it from the register the IOP wrote it in rather than from anywhere else.
//
// A transfer is a copy. The two processors are two Go objects over two byte slices in one
// process, and the SIF is the wire between them; there is no third thing for the data to sit
// in, and modelling a FIFO would model the wire rather than the message.

// sifPacket is one transfer the EE has handed to the SIF and the IOP has not taken yet.
type sifPacket struct {
	dest uint32 // where in IOP memory the EE says it goes
	data []byte
	cmd  bool // a command packet, which rings SIFCMD's doorbell; otherwise plain data
}

// sifToIOP hands one command packet to SIF1. It does not deliver it.
//
// The distinction is the whole of the flow control on this channel, and there is flow control,
// because SIFCMD has exactly one receive slot. Its interrupt handler reads the packet from the
// single address it published in SMCOM, marks the slot free by zeroing the packet's first byte,
// and only then calls sifman#6 to arm the receive channel again. Until it does, the channel is
// not armed and nothing can land.
//
// So the sequence on the board is: the EE's DMA waits for an armed channel, fills the slot, and
// the IOP's handler empties it and re-arms. A model that copies the bytes in the instant the EE
// asks skips the wait — and the EE, which runs eight times faster here, sends its second packet
// before the IOP has executed a single instruction. The second overwrites the first in the one
// slot they share.
//
// That is not a subtle failure. The first packet is INIT_CMD with opt = 0, which releases every
// thread on the second processor, and the second is INIT_CMD with opt = 1, which tells them the
// EE's own RPC layer is up. Deliver both to the same address and one of them simply never
// happened — and which one you lose decides which half of the machine waits forever.
func (m *Machine) sifToIOP(src, dest, size uint32) {
	if dest == 0 {
		m.note("SIF: the EE sent the IOP a packet before it knew where the IOP wanted it (dest is 0)")
		return
	}
	p := sifPacket{dest: dest, data: make([]byte, size), cmd: true}
	for i := uint32(0); i < size; i++ {
		p.data[i] = m.Read(src + i)
	}
	m.sifToIOPQueue = append(m.sifToIOPQueue, p)
	m.sifPump()
}

// sifPump delivers the next queued packet, if the IOP is ready for one.
//
// "Ready" is the IOP's own statement, and the only one available: its receive channel is armed.
// SIFMAN arms it with a start bit and a block size and no MADR at all — which is exactly what an
// armed receiver looks like, and why this is the one channel that does not complete when it is
// started (iopdma.go). It completes here, when something arrives for it.
//
// It is called from two places and needs both: from sifToIOP, for a packet arriving at a channel
// that is already waiting, and from the DMA controller, for a channel being armed with a packet
// already waiting.
func (m *Machine) sifPump() {
	if m.IOP == nil || len(m.sifToIOPQueue) == 0 {
		return
	}
	c := &m.IOP.dma[iopDMAChSIF1]

	// The queue drains in order, and stops at the first thing the IOP is not ready for. Only a
	// command needs an armed channel, because only a command lands in the single slot SIFCMD
	// polls; the arguments that precede it go to a buffer of the server's own, and nothing on
	// the IOP is looking at it until the command tells it to.
	for len(m.sifToIOPQueue) > 0 {
		p := m.sifToIOPQueue[0]
		if p.cmd && c.chcr&iopChcrStart == 0 {
			return // the IOP is not listening; the packet waits, as it would on the wire
		}
		m.sifToIOPQueue = m.sifToIOPQueue[1:]

		for i, b := range p.data {
			m.IOP.Write(p.dest+uint32(i), b)
		}
		m.sifToIOPCount++
		if !p.cmd {
			continue
		}

		// The transfer the IOP armed has now happened. Its channel is no longer busy, and its
		// MADR holds the address the data went to — which is the *sender's* choice, and so
		// something only the completed transfer can tell the receiver. SIFMAN never writes this
		// register; it only ever reads it back. That is what says the hardware writes it.
		c.madr = p.dest + uint32(len(p.data))
		c.chcr &^= iopChcrStart | iopChcrTrigger

		// And the interrupt SIFCMD is waiting on. It is raised, not run: the handler belongs to
		// the IOP and runs on the IOP's own interrupt path, on the IOP's own stack, the next
		// time that processor takes a step. Which is the whole difference between the two
		// machines talking and one of them ventriloquising the other.
		m.IOP.raiseIRQ(iopDMAIRQ(iopDMAChSIF1))
	}
}

// sifFromIOP carries a transfer the other way: the IOP has started its SIF0 channel, and what
// it is sending is a command packet for the EE.
//
// SIF0 is not a burst, and this is where that turns out to matter. SIFMAN never writes the
// channel's MADR — it writes its TADR — and a controller that reads MADR anyway finds a zero
// there and dutifully sends the EE 128 bytes of address nought. What it writes instead is a
// chain, and the chain is worth reading, because it is the whole protocol in four words:
//
//	8001BC20   the source, in IOP memory, with a flag in the top byte
//	00000006   six words of it — twenty-four bytes, which is one command packet
//	90000002   a DMA tag for the *EE's* controller: two quadwords, and raise an interrupt
//	0013B9C0   and the address in EE memory to put them at
//
// So the destination does not come from anything on this side. It rides in front of the data,
// and the EE's DMA channel 5 — which sceSifInitCmd put into destination-chain mode with
// sceSifSetDChain, and which is the one call in that routine that had no purpose until now —
// reads it out and obeys it. That is the exact mirror of SIF1, where the destination comes
// from the sender too (sifToIOP), and it is why neither channel ever needed a MADR.
//
// Taking the address from the tag rather than from a variable of our own is not tidiness. It
// is what makes CHANGE_SADDR work: when the EE moves its command buffer, the IOP is told, and
// the next tag simply says somewhere else.
func (m *Machine) sifFromIOP() {
	c := &m.IOP.dma[iopDMAChSIF0]

	for tag := c.tadr; ; tag += sifChainTagSize {
		src := m.IOP.Read32(tag+0) & sifChainAddrMask
		words := m.IOP.Read32(tag + 4)
		eeTag := m.IOP.Read32(tag + 8)
		dest := m.IOP.Read32(tag + 12)
		if words == 0 || dest == 0 {
			break
		}

		for i := uint32(0); i < words*4; i++ {
			m.Write(dest+i, m.IOP.Read(src+i))
		}
		m.sifFromIOPCount++

		// Not everything that crosses is a command. A remote call's *reply* comes back the same
		// way, in a transfer of its own, straight into the buffer the caller nominated — which
		// is why the census asks where the bytes landed rather than what they say. Only what
		// lands in the EE's command buffer is a command packet; everything else is the answer
		// to one, and reading a cid out of it would invent a command from somebody's payload.
		if dest == m.sifCmdBuf {
			m.sifBack[m.Read32(dest+0x08)]++
		}

		// And the EE's handler — but only when the tag asks for it. The word at +8 is a DMA tag
		// for the *EE's* controller, and its top bit is that controller's interrupt-on-completion
		// flag. So the IOP is not merely sending bytes; it is saying, in the EE's own register
		// format, whether this transfer is worth waking the EE for. Honouring that bit rather
		// than running the handler after every tag is the difference between the EE's interrupt
		// happening because the IOP asked for it and happening because we felt like it.
		//
		// It is the same handler, reached the same way, as when the fake IOP forged the packet.
		// Nothing on the EE's side had to change for the second processor to become real: the EE
		// was always reading its command buffer and dispatching on the cid it found there. All
		// that has changed is who wrote it.
		if eeTag&sifEETagIRQ != 0 && m.sifCmdHandler != 0 {
			m.callGuest(m.sifCmdHandler, 0)
			// The handler ran in interrupt context and may have woken a thread that
			// outranks the one it interrupted; the kernel reschedules on the way out.
			m.preemptIfOutranked()
		}

		if m.IOP.Read32(tag+0)&sifChainEnd != 0 {
			break
		}
	}
}

// The SIF0 chain tag, as SIFMAN builds it.
const (
	sifChainTagSize  = 16
	sifChainAddrMask = 0x00FFFFFF // the IOP has 2 MiB; the address is 24 bits of it
	sifChainEnd      = 1 << 31    // the last tag in the chain

	// The EE's DMA tag rides in the third word, and this is its interrupt bit: the IOP telling
	// the EE's DMA controller to raise channel 5 when this transfer lands.
	sifEETagIRQ = 1 << 31
)

// --- the RPC layer, as it goes past ------------------------------------------
//
// There used to be one of these here, written in Go: it bound a synthetic handle for every
// server the EE asked for, and answered the two version checks the game makes with numbers
// chosen to get past them ("2210" for the module loader, 522 and 526 for the memory card).
// The comment above it conceded the point — "reporting a number the game accepts is a stand-in,
// not an answer" — and it is gone. The versions are now reported by the modules that have them,
// because the modules are running.
//
// What is left is the part that was always worth having: the census. A remote call is two
// packets, BIND or CALL out and END back, and reading the cid as each one crosses costs
// nothing and says exactly what the two processors are doing. Since the servers on the other
// side are the game's own IOP modules, this census *is* the reverse engineering of their
// interface — it is the game telling us its service protocol, one call at a time.

// The offsets inside a BIND and a CALL packet, read out of the routines that build them
// (sceSifBindRpc and sceSifCallRpc) and consumed by the ones that serve them.
const (
	rpcPktServerID = 0x20 // BIND: which server the EE wants
	rpcPktFuncNo   = 0x20 // CALL: which function on it
	rpcPktServer   = 0x34 // CALL: the handle BIND gave back
)

// sifWatch records a packet on its way past, in either direction. It never answers one.
func (m *Machine) sifWatch(src, cid uint32, toIOP bool) {
	switch {
	case toIOP && cid == sifCmdRpcBind:
		m.rpcBinds[m.Read32(src+rpcPktServerID)]++

	case toIOP && cid == sifCmdRpcCall:
		// The server *handle*, not its id: the EE calls through the handle the IOP gave it.
		// Which handle belongs to which server is something only the IOP knows, so the census
		// records both and does not pretend to join them.
		m.rpcCalls[sifRPCKey{
			sid: m.Read32(src + rpcPktServer),
			fno: m.Read32(src + rpcPktFuncNo),
		}]++
	}
}

// --- the shared registers ----------------------------------------------------
//
// These are the handshake registers the two chips read and write directly, distinct
// from the SIF registers (SREGs) that live in EE memory and are kept in step by
// SET_SREG packets.

// sceSifSetReg and sceSifGetReg are syscalls 121 and 122 — the EE kernel's SIF registers,
// and the machinery by which the two processors first find each other. The boot ELF names
// them, and the game's own use of them is the whole specification.
//
// The argument is a *register number*, and it is emphatically not an index into an array of
// thirty-two. Four numbers appear in the whole boot, and two of them have the top bit set:
//
//	2             the IOP's command buffer, which is to say SMCOM
//	4             the IOP's flags, which is to say SMFLG
//	0x80000000    the kernel's cache of the first of those
//	0x80000001    the EE's own command block, published for the IOP
//	0x80000002    "the RPC layer is up"
//
// This mattered more than anything else in the SIF, because reading the argument as an index
// and masking it to five bits turns 0x80000000 into register 0 — and register 0 used to
// answer with a constant. See sifRegIOPCmdBuf below: that constant chose the wrong branch of
// sceSifInitCmd for the entire boot, and the IOP waited for a packet the EE had been talked
// out of sending.
const (
	// The two that are hardware: the IOP writes them across the SBUS, and reading them here
	// is reading what the second processor actually put there.
	sifRegIOPCmdBufHW = 2 // SMCOM: SIFCMD's receive buffer, published by sifman#27
	sifRegIOPFlags    = 4 // SMFLG: the IOP's flags to the EE

	// The three that are software: slots the EE kernel keeps for the EE's own use. No IOP
	// module reads or writes any of them.
	sifRegIOPCmdBuf = 0x80000000 // the cache of SMCOM — see below
	sifRegEECmdBuf  = 0x80000001 // the EE's command block, for the IOP to find
	sifRegRPCUp     = 0x80000002 // guards sceSifInitRpc against running twice
)

// The bits of SMFLG, each one traced to the module that raises it (sifbus.go).
const (
	sifIOPSIFUp      = 0x00010000 // SIFMAN's init
	sifIOPCmdUp      = 0x00020000 // SIFCMD, just before it blocks waiting for the EE
	sifIOPRebootDone = 0x00040000 // EESYNC, once every module has loaded
)

// sifSetReg writes one of them.
//
// Register 4 is the exception, and it is the interesting one: the EE does not *own* SMFLG, it
// reads it, so a write from this side clears the bits it names rather than storing them.
// sceSifSyncIop is the proof — it tests bit 0x40000 and then writes that same bit back, which
// is an acknowledgement or a self-inflicted wound, and only one of those is a kernel.
func (m *Machine) sifSetReg() {
	reg, val := m.arg(0), m.arg(1)
	switch reg {
	case sifRegIOPFlags:
		m.sbusFlagClear(sbusSMFLG, val)
	case sifRegIOPCmdBufHW:
		m.sbusWrite(sbusSMCOM, val)
	case sifRegIOPCmdBuf, sifRegEECmdBuf, sifRegRPCUp:
		m.sifRegs[reg] = val
	default:
		m.sifUnmodelledReg[reg]++
	}
	m.setRet(0)
}

// sifGetReg reads one.
//
// sifRegIOPCmdBuf is worth reading sceSifInitCmd for, because the whole handshake turns on it
// and it is not the register it looks like. It is the kernel's *cache* of the IOP's command
// buffer, and what the EE is really asking is "have I met this IOP before?".
//
//   - Not zero: yes. Just tell it where my buffer has moved to — CHANGE_SADDR — and carry on.
//   - Zero: no. Wait for the IOP to raise SMFLG's 0x20000 ("my command layer is listening"),
//     read its buffer out of SMCOM, cache it here, and send it INIT_CMD with opt = 0.
//
// That second packet is the one that matters. On the IOP, INIT_CMD(opt=0) sets event-flag bit
// 0x100 — and every module on the second processor ends its initialisation blocked on that
// bit, including SIFCMD's own RPC layer, which cannot announce itself to the EE until it is
// released. sceSifResetIop clears this register precisely so that the next IOP gets the long
// branch. An emulator that answers it with anything non-zero gets the short one, and the IOP
// then waits forever for a packet that the EE has been told it does not need to send.
func (m *Machine) sifGetReg() {
	reg := m.arg(0)
	switch reg {
	case sifRegIOPFlags:
		m.setRet(m.sbusRead(sbusSMFLG))
	case sifRegIOPCmdBufHW:
		m.setRet(m.sbusRead(sbusSMCOM))
	case sifRegIOPCmdBuf, sifRegEECmdBuf, sifRegRPCUp:
		m.setRet(m.sifRegs[reg])
	default:
		m.sifUnmodelledReg[reg]++
		m.setRet(0)
	}
}

// sifRPCKey names one remote procedure: a server and a function number.
type sifRPCKey struct {
	sid uint32
	fno uint32
}

// sifCmdName names a command id, so the census reads as a conversation rather than a column
// of numbers. Every name here was earned from the code that serves it — see sifGetReg for
// INIT_CMD, which is the one that matters.
func sifCmdName(cid uint32) string {
	switch cid {
	case sifCmdChangeSaddr:
		return "CHANGE_SADDR"
	case sifCmdSetSreg:
		return "SET_SREG"
	case sifCmdInitCmd:
		return "INIT_CMD"
	case sifCmdReset:
		return "RESET"
	case sifCmdRpcEnd:
		return "RPC_END"
	case sifCmdRpcBind:
		return "RPC_BIND"
	case sifCmdRpcCall:
		return "RPC_CALL"
	case sifCmdRpcRData:
		return "RPC_RDATA"
	}
	return "?"
}

// smflgBits names the bits the IOP has raised, because the number alone says nothing and each
// of the three was traced to the module that raises it (sifbus.go).
func smflgBits(v uint32) string {
	var names []string
	for _, b := range []struct {
		bit  uint32
		name string
	}{
		{sifIOPSIFUp, "SIF up"},
		{sifIOPCmdUp, "command layer listening"},
		{sifIOPRebootDone, "reboot done"},
	} {
		if v&b.bit != 0 {
			names = append(names, b.name)
		}
	}
	if len(names) == 0 {
		return "  (the IOP has raised nothing)"
	}
	return "  (" + strings.Join(names, ", ") + ")"
}

// SIFCensus reports the conversation between the two processors: the commands that crossed in
// each direction, the servers the EE bound to, and the functions it called on them.
//
// Nothing in this machine answers any of it any more, so this is a record rather than a work
// list — and it is a better one for that. The servers are the game's own IOP modules and the
// functions are their interface, so what it prints is Naughty Dog's service protocol, observed
// rather than assumed.
func (m *Machine) SIFCensus() string {
	// The two counts are printed always, and never suppressed when they are zero, because a
	// zero is the finding. Two processors that have exchanged nothing look exactly like two
	// processors getting on with it, from every other angle in this machine — and a machine
	// that sends and never receives is a machine talking to itself.
	s := sprintf("the SIF: %d packets crossed to the IOP, %d came back.\n",
		m.sifToIOPCount, m.sifFromIOPCount)
	s += sprintf("  SMCOM 0x%08X (the IOP's command buffer)   SMFLG 0x%08X%s\n",
		m.sbusRead(sbusSMCOM), m.sbusRead(sbusSMFLG), smflgBits(m.sbusRead(sbusSMFLG)))
	s += sprintf("  the EE's command buffer 0x%08X, its handler %s\n",
		m.sifCmdBuf, m.Sym(m.sifCmdHandler))

	for _, d := range []struct {
		way string
		m   map[uint32]int
	}{{"->", m.sifSent}, {"<-", m.sifBack}} {
		var cids []uint32
		for cid := range d.m {
			cids = append(cids, cid)
		}
		sort.Slice(cids, func(i, j int) bool { return cids[i] < cids[j] })
		for _, cid := range cids {
			s += sprintf("  %s  %-13s 0x%08X  %d\n", d.way, sifCmdName(cid), cid, d.m[cid])
		}
	}

	if len(m.rpcBinds) > 0 {
		s += "  the servers the EE bound to, on the IOP:\n"
		var sids []uint32
		for sid := range m.rpcBinds {
			sids = append(sids, sid)
		}
		sort.Slice(sids, func(i, j int) bool { return sids[i] < sids[j] })
		for _, sid := range sids {
			s += sprintf("      0x%08X  bound %d time%s\n", sid, m.rpcBinds[sid], plural(m.rpcBinds[sid]))
		}
	}
	if len(m.rpcCalls) > 0 {
		s += "  and the functions it called, by the handle the IOP gave it:\n"
		var keys []sifRPCKey
		for k := range m.rpcCalls {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].sid != keys[j].sid {
				return keys[i].sid < keys[j].sid
			}
			return keys[i].fno < keys[j].fno
		})
		for _, k := range keys {
			s += sprintf("      handle 0x%08X  fn %-4d  %d call%s\n",
				k.sid, k.fno, m.rpcCalls[k], plural(m.rpcCalls[k]))
		}
	}

	for reg, n := range m.sifUnmodelledReg {
		s += sprintf("  SIF register 0x%08X asked for %d time%s, and nothing models it\n",
			reg, n, plural(n))
	}
	return s
}
