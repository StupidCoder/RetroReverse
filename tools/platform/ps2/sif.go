package ps2

import "sort"

// sif.go is the boundary between the two processors, with the far side faked.
//
// A PS2 has a second CPU, the IOP — a MIPS R3000A, the PlayStation's chip — which
// owns the disc drive, the sound chip and the controllers. The EE reaches it across
// the SIF: a DMA path carrying command packets, plus a remote-procedure-call layer
// built on top of them. The disc ships the IOP's own modules (OVERLORD.IRX,
// 989SND.IRX, PADMAN.IRX) to be uploaded and run there.
//
// None of that runs yet. This file answers the EE as the IOP would, in Go.
//
// The pleasant surprise is how little forging that takes, because the EE does its own
// unpacking. A command packet is:
//
//	+0x00  psize in the low byte, dsize in the upper three
//	+0x04  dest        where the data half should land
//	+0x08  cid         the command; bit 31 selects the system handler table
//	+0x0C  opt
//	+0x10  payload
//
// and the EE registers `_sceSifCmdIntrHdlr` on DMA channel 5 to consume them. That
// routine reads psize from the packet's first *byte*, copies the packet out, zeroes
// that byte to mark it consumed, and dispatches on cid. So the IOP's whole side of the
// conversation is: write a packet into the buffer the EE nominated, and run the EE's
// own handler over it. The game does the rest with its own code — nothing here has to
// know where its SIF registers live or what its handlers do.
//
// The real thing is not far off: tools/cpu/mips already models the R3000A exactly, and
// IRX modules import from the IOP kernel through stub tables of the same shape the
// PSP's PRX modules use, so the trick that HLEs those applies here too. When that
// lands, this file is what it replaces.

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

// sifRPCInitSreg is the SIF register the EE polls to learn that the IOP's RPC layer is
// up. sceSifInitRpc spins on it and will not return until it is non-zero — which,
// with no IOP, it never would.
const (
	sifRPCInitSreg = 0
	sifRPCReady    = 1
)

// sifLatency is how long the fake IOP takes to answer, in EE instructions.
//
// It is not zero on purpose. Answering inside the send would let a reply be processed
// before the sender had finished updating the state the handler reads — a race that
// cannot happen on hardware, where the IOP is a separate chip and the reply arrives as
// an interrupt some time later. A few thousand instructions is short enough that a
// polling loop does not trip the spin detector, and long enough that the sender always
// finishes first.
const sifLatency = 4000

// sifPacket is one command packet on its way to the EE.
type sifPacket struct {
	at   uint64 // the step count at which it should arrive
	data []uint32
}

// sifSetDma is syscall 119: the EE handing buffers to the IOP. Each descriptor is
// {src, dest, size, attr}.
//
// A call carries two of them — the arguments and then the command packet — and telling
// them apart matters. A command packet always has a cid with the top bit set at +0x08;
// nothing else does. Feeding the argument buffer to the command dispatcher instead
// yields a "command" of 0x6F726463, which is "cdro" — the first four bytes of
// "cdrom0:\DRIVERS\SIO2MAN.IRX;1", the path the game was trying to pass.
//
// The argument buffer stays in EE memory, so remembering where it was is all a served
// call needs to read it.
func (m *Machine) sifSetDma() {
	desc, count := m.arg(0), m.arg(1)
	for i := uint32(0); i < count; i++ {
		d := desc + i*16
		src := m.Read32(d + 0x00)
		size := m.Read32(d + 0x08)


		if size >= 16 && m.Read32(src+0x08)&0x80000000 != 0 {
			m.iopReceive(src, size)
			m.rpcSendBuf, m.rpcSendSize = 0, 0
			continue
		}
		m.rpcSendBuf, m.rpcSendSize = src, size
	}
	// A handle of zero means failure, and the caller checks. It must never be zero.
	m.sifDmaID++
	m.setRet(m.sifDmaID)
}

// iopReceive is the IOP's side: one command packet arrives from the EE.
func (m *Machine) iopReceive(src, size uint32) {
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

	// The EE nominates the buffer it wants the IOP's packets written into, and the machine has
	// to know it too — not to answer with, but to *route* with: it is where a SIF0 transfer
	// out of the IOP is going to land. The IOP's own SIFCMD is told the same thing by the same
	// packet, and keeps its own copy; this is the wire noting the address on the envelope.
	if cid == sifCmdChangeSaddr {
		m.sifCmdBuf = m.Read32(src + 0x10)
		m.note("SIF: the EE's command buffer is at 0x%08X", m.sifCmdBuf)
	}

	// And here is where the packet will one day go to the second processor instead of to the
	// Go handlers below — sifToIOP carries it, and it works: the bytes land in SIFCMD's own
	// receive buffer at the address SIFCMD published in SMCOM, interrupt 43 is raised, and
	// SIFCMD's handler runs on the IOP and consumes them.
	//
	// What does not work yet is the answer. SIFCMD takes the packet and sends nothing back, so
	// the EE waits for a reply that never comes — and until it *does* come, switching the wire
	// on would only replace a machine that lies with a machine that stops. The fakes stay until
	// both halves of the conversation are real. That is the next piece of work and it is a
	// short one: find out what SIFCMD does with a command packet it has accepted, and why it
	// does not start its SIF0 channel afterwards.

	switch cid {
	case sifCmdChangeSaddr:
		// The EE nominates the buffer it wants replies written into. Everything this
		// file sends goes here, and it is the only address it needs to know.
		m.sifCmdBuf = m.Read32(src + 0x10)
		m.note("SIF: the EE's command buffer is at 0x%08X", m.sifCmdBuf)

	case sifCmdInitCmd:
		// The EE is bringing its command and RPC layers up, and is about to spin waiting
		// for the IOP to say its own RPC layer is ready. The IOP says so by writing the
		// EE's SIF register 0 — which it does by sending a SET_SREG command back.
		m.note("SIF: the EE initialised its RPC layer (opt=%d); answering that the IOP's is ready", opt)
		m.sifSend(sifCmdSetSreg, 0, sifRPCInitSreg, sifRPCReady)

	case sifCmdReset:
		m.iopReset(src)

	case sifCmdRpcBind:
		m.rpcBind(src)

	case sifCmdRpcCall:
		m.rpcCall(src)

	default:
		m.note("SIF: unmodelled command 0x%08X (opt=0x%08X) from %s — %d bytes at 0x%08X",
			cid, opt, m.Sym(uint32(m.CPU.CurPC())), size, src)
		m.sifUnmodelled[cid]++
	}
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

	if err := m.RebootIOPFrom(image); err != nil {
		m.note("SIF: the IOP did not reboot: %v", err)
		return
	}
	m.iopRebooted = true
}

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

// sifToIOP carries one command packet across SIF1 and rings the IOP's doorbell.
func (m *Machine) sifToIOP(src, size uint32) {
	dest := m.sbusRead(sbusSMCOM)
	if dest == 0 {
		m.note("SIF: the EE sent the IOP a packet before the IOP said where to put it (SMCOM is 0)")
		return
	}
	for i := uint32(0); i < size; i++ {
		m.IOP.Write(dest+i, m.Read(src+i))
	}

	// The transfer the IOP armed has now happened. Its channel is no longer busy, and its MADR
	// holds the address the data went to — which is the *sender's* choice, and therefore
	// something only the completed transfer can tell the receiver. SIFMAN never writes this
	// register; it only ever reads it back. That is what says the hardware writes it.
	c := &m.IOP.dma[iopDMAChSIF1]
	c.madr = dest + size
	c.chcr &^= iopChcrStart | iopChcrTrigger

	// And the interrupt SIFCMD is waiting on. It is raised, not run: the handler belongs to the
	// IOP and runs on the IOP's own interrupt path, on the IOP's own stack, the next time that
	// processor takes a step. Which is the whole difference between the two machines talking and
	// one of them ventriloquising the other.
	m.IOP.raiseIRQ(iopDMAIRQ(iopDMAChSIF1))
	m.sifToIOPCount++
}

// sifFromIOP carries a transfer the other way: the IOP has started its SIF0 channel, and what
// it is sending is a command packet for the EE.
//
// It lands in the buffer the EE nominated with CHANGE_SADDR, and then the EE's own handler runs
// over it — the same handler, reached the same way, as when the fake IOP wrote the packet. That
// is the point: nothing on the EE's side had to change for the IOP to become real. The EE was
// always reading its command buffer and dispatching on the cid it found there. All that has
// changed is who wrote it.
func (m *Machine) sifFromIOP(madr, n uint32) {
	if m.sifCmdBuf == 0 || m.sifCmdHandler == 0 {
		m.note("SIF: the IOP sent %d bytes, and the EE has not said where its command buffer is", n)
		return
	}
	for i := uint32(0); i < n; i++ {
		m.Write(m.sifCmdBuf+i, m.IOP.Read(madr+i))
	}
	m.sifFromIOPCount++
	m.callGuest(m.sifCmdHandler, 0)
}

// --- the RPC layer -----------------------------------------------------------
//
// A remote call is two packets. The EE sends BIND (or CALL) and blocks; the IOP does
// the work and sends END back, and the EE's `_request_end` handler unblocks the
// waiting thread. END carries, at +0x20, the *kind* of request it is ending — which is
// how one handler serves both.
//
// The layouts below were read out of `_request_end` itself rather than assumed: it is
// the code that consumes these fields, so it is the authority on where they are.

// rpcBind answers a bind request. The EE wants a handle for a server id; it gets a
// synthetic one, stable per id, and the boot proceeds.
func (m *Machine) rpcBind(src uint32) {
	recID := m.Read32(src + 0x10)
	pktAddr := m.Read32(src + 0x14)
	rpcID := m.Read32(src + 0x18)
	client := m.Read32(src + 0x1C)
	sid := m.Read32(src + 0x20)

	server := m.rpcServerHandle(sid)
	m.note("SIF RPC: the EE bound to server 0x%08X (handle 0x%08X)", sid, server)

	// _request_end reads: [+0x1C] client, [+0x20] which request, [+0x24] server,
	// [+0x28] buff, [+0x2C] cbuff — and writes the last three into the client struct.
	m.sifSend(sifCmdRpcEnd, 0,
		recID, pktAddr, rpcID, client,
		sifCmdRpcBind,
		server, 0, 0)
}

// rpcServerHandle invents a stable, non-null handle for a server id. It must be
// non-null: the EE tests the handle before it will call through it.
func (m *Machine) rpcServerHandle(sid uint32) uint32 {
	if h, ok := m.rpcServers[sid]; ok {
		return h
	}
	h := rpcServerBase + uint32(len(m.rpcServers))*0x40
	m.rpcServers[sid] = h
	m.rpcServerOf[h] = sid
	return h
}

// rpcServerBase is where the synthetic server handles live. They are never
// dereferenced by the EE — it only passes them back — so they need to be distinct and
// non-null, not real.
const rpcServerBase = 0x1C100000

// The IOP services the game binds to. The first two are Sony's; the 0x59x range is
// Naughty Dog's own — those are the servers inside OVERLORD.IRX, and they are the ones
// that will still need the real IOP.
const (
	rpcServerFileIO   = 0x80000001 // the file system
	rpcServerLoadFile = 0x80000006 // the IOP module loader
)

// rpcVersion is the function every Sony IOP service answers with its version string.
// The loader's is checked before a single module is loaded: `_lf_version` compares four
// bytes against "2210", and a mismatch fails the load with "loading sio2man.irx failed"
// and no further explanation. The number matches IOPRP221.IMG, the module archive on
// this disc.
const (
	rpcVersion  = 255
	loadFileVer = "2210"
)

// rpcCall answers a call request. Which server and which function are recorded, because
// that census *is* the reverse engineering of the IOP's protocol: the game tells us its
// own service interface, one call at a time.
func (m *Machine) rpcCall(src uint32) {
	recID := m.Read32(src + 0x10)
	pktAddr := m.Read32(src + 0x14)
	rpcID := m.Read32(src + 0x18)
	client := m.Read32(src + 0x1C)
	fno := m.Read32(src + 0x20)
	sendSize := m.Read32(src + 0x24)
	recv := m.Read32(src + 0x28)
	recvSize := m.Read32(src + 0x2C)
	server := m.Read32(src + 0x34)

	sid := m.rpcServerOf[server]
	key := sifRPCKey{sid: sid, fno: fno}
	m.rpcCalls[key]++

	if m.rpcCalls[key] == 1 {
		m.note("SIF RPC: call server 0x%08X fn %d — %d bytes out, %d bytes back into 0x%08X",
			sid, fno, sendSize, recvSize, recv)
	}

	m.rpcServe(sid, fno, recv, recvSize)

	// The reply always goes out, answered or not. Without it the thread that made the
	// call never wakes, and the boot stops there rather than going on to tell us what it
	// wanted next.
	m.sifSend(sifCmdRpcEnd, 0, recID, pktAddr, rpcID, client, sifCmdRpcCall)
}

// rpcServe is the IOP's side of a call: it writes the reply into the EE's receive
// buffer. Anything not handled leaves the buffer alone, which the census records.
//
// The version answers here are the game's own minimums, read out of the code that
// checks them rather than invented. Jak refuses to run against an IOP module it thinks
// is too old, and says so — "libmc: too old release of mcserv.irx" — so each check is a
// signpost. sceMcInit compares the memory-card service's version against 522 and gives
// up below it; the loader compares four bytes against "2210".
//
// Reporting a number the game accepts is a stand-in, not an answer. The real version is
// inside MCSERV.IRX, which is on the disc, and the module that would report it is the
// module we are not running. This is the clearest argument in the whole machine for
// running the IOP for real.
func (m *Machine) rpcServe(sid, fno, recv, recvSize uint32) {
	if recv == 0 || recvSize == 0 {
		return
	}
	switch {
	case sid == rpcServerLoadFile && fno == rpcVersion:
		m.writeString(recv, loadFileVer, recvSize)

	case sid == rpcServerMemCard && fno == mcFnInit && recvSize >= 12:
		// The reply is {result, mcservVersion, mcmanVersion}. sceMcInit checks both, and
		// names whichever one it dislikes.
		m.Write32(recv+4, mcServMinVersion)
		m.Write32(recv+8, mcManMinVersion)
	}
}

const (
	rpcServerMemCard = 0x80000400 // the memory-card service, in MCSERV.IRX
	mcFnInit         = 254

	// The lowest versions sceMcInit will accept, read off the two comparisons in it.
	mcServMinVersion = 522
	mcManMinVersion  = 526
)

// writeString writes at most n bytes of s into guest memory.
func (m *Machine) writeString(addr uint32, s string, n uint32) {
	for i := 0; i < len(s) && uint32(i) < n; i++ {
		m.Write(addr+uint32(i), s[i])
	}
}

// sifRPCKey names one remote procedure: a server and a function number.
type sifRPCKey struct {
	sid uint32
	fno uint32
}

// sifSend queues a command packet for the EE. It is delivered by sifTick, once the
// latency has elapsed.
func (m *Machine) sifSend(cid, opt uint32, payload ...uint32) {
	pkt := make([]uint32, 4+len(payload))
	pkt[0] = uint32(16 + 4*len(payload)) // psize; dsize stays zero
	pkt[1] = 0                           // dest: there is no data half
	pkt[2] = cid
	pkt[3] = opt
	copy(pkt[4:], payload)
	m.sifPending = append(m.sifPending, sifPacket{at: m.steps + sifLatency, data: pkt})
}

// sifTick delivers any packet whose time has come. The run loop calls it.
func (m *Machine) sifTick() {
	for len(m.sifPending) > 0 && m.steps >= m.sifPending[0].at {
		p := m.sifPending[0]
		if !m.sifDeliver(p.data) {
			return // the EE has not consumed the last one yet; try again next step
		}
		m.sifPending = m.sifPending[1:]
	}
}

// sifDeliver writes a packet into the EE's command buffer and runs the EE's own
// interrupt handler over it. It reports false when the buffer is still occupied.
func (m *Machine) sifDeliver(pkt []uint32) bool {
	if m.sifCmdBuf == 0 || m.sifCmdHandler == 0 {
		return true // nowhere to put it, and nobody to read it: drop it
	}
	// The handler zeroes the packet's first byte when it has consumed it. A non-zero
	// psize means the last packet is still sitting there unread, and overwriting it
	// would lose a command.
	if m.Read32(m.sifCmdBuf)&0xFF != 0 {
		return false
	}
	for i, w := range pkt {
		m.Write32(m.sifCmdBuf+uint32(i)*4, w)
	}
	m.callGuest(m.sifCmdHandler, 0)
	return true
}

// --- the shared registers ----------------------------------------------------
//
// These are the handshake registers the two chips read and write directly, distinct
// from the SIF registers (SREGs) that live in EE memory and are kept in step by
// SET_SREG packets.

func (m *Machine) sifSetReg() {
	reg, val := m.arg(0), m.arg(1)
	m.sifRegs[reg&0x1F] = val
	m.setRet(0)
}

func (m *Machine) sifGetReg() {
	reg := m.arg(0) & 0x1F
	v := m.sifRegs[reg]

	switch reg {
	case sifRegIOPAlive:
		// The EE will not talk to the IOP at all until this says it is there. Nothing
		// would ever set it, because nothing is there.
		v |= sifIOPAlive

	case sifRegIOPReset:
		// The game reboots the IOP at boot — it loads a fresh set of modules over the ones the
		// BIOS left — and then sits in a loop printing "Syncing..." until sceSifSyncIop sees
		// this bit.
		//
		// It used to be set unconditionally, and the comment here said why: there was no IOP to
		// reboot, so it was always finished rebooting. That is no longer true. The reboot now
		// happens (iopReset) — the EE names an image, the disc is asked for it, and twelve
		// modules are loaded and started on a real R3000A — so the bit says what it means, and
		// an IOP that fails to come up is an IOP the EE will wait for rather than one it is
		// told a comfortable story about.
		if m.iopRebooted {
			v |= sifIOPRebootDone
		}
	}
	m.setRet(v)
}

// The shared registers the EE reads to find out about the IOP.
const (
	sifRegIOPAlive = 0 // "the IOP is there"
	sifRegIOPReset = 4 // "the IOP has finished rebooting"

	sifIOPAlive      = 0x00000001
	sifIOPRebootDone = 0x00040000
)

// SIFCensus reports what the EE asked the IOP for: the remote calls it made, and any
// SIF command nothing answered. It is the IOP's work list, exactly as the syscall
// census is the kernel's — and, since the servers on the other side are the game's own
// IOP modules, it is also the reverse engineering of their interface.
func (m *Machine) SIFCensus() string {
	if len(m.rpcCalls) == 0 && len(m.sifUnmodelled) == 0 {
		return ""
	}
	s := sprintf("the EE's requests to the IOP (%d packets crossed to it, %d came back):\n",
		m.sifToIOPCount, m.sifFromIOPCount)

	if len(m.rpcServers) > 0 {
		var sids []uint32
		for sid := range m.rpcServers {
			sids = append(sids, sid)
		}
		sort.Slice(sids, func(i, j int) bool { return sids[i] < sids[j] })
		for _, sid := range sids {
			s += sprintf("  server 0x%08X\n", sid)
			type fn struct {
				fno uint32
				n   int
			}
			var fns []fn
			for k, n := range m.rpcCalls {
				if k.sid == sid {
					fns = append(fns, fn{k.fno, n})
				}
			}
			sort.Slice(fns, func(i, j int) bool { return fns[i].n > fns[j].n })
			for _, f := range fns {
				s += sprintf("      fn %-4d  %d call%s\n", f.fno, f.n, plural(f.n))
			}
		}
	}
	for cid, n := range m.sifUnmodelled {
		s += sprintf("  unanswered command 0x%08X  %d\n", cid, n)
	}
	return s
}
