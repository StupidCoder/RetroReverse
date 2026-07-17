package ps2

// intr.go is the interrupt side of the machine: the two interrupt controllers the EE
// sees (the INTC, for the peripherals, and the DMAC, for transfer completion), and
// the primitive that lets the machine call back into guest code.
//
// The EE takes its peripheral interrupts on two lines, INT0 (the INTC) and INT1 (the
// DMAC), and the kernel's AddIntcHandler / AddDmacHandler register the routines it
// wants run for each cause. Under this model the kernel is ours, so those handlers
// are held here and invoked directly rather than through a vectored dispatch.

import "retroreverse.com/tools/cpu/r5900"

// The INTC causes. Only the vertical blanks matter to a boot, but the whole set is
// named so the census reads as hardware.
const (
	intcGS        = 0
	intcSBUS      = 1
	intcVBlankOn  = 2 // the start of the vertical blank
	intcVBlankOff = 3 // the end of it

	// The INTC's memory-mapped pair, polled directly by games that pace without
	// handlers. I_STAT is write-1-to-clear; I_MASK is write-1-to-toggle.
	intcISTAT = 0x1000F000
	intcIMASK = 0x1000F010
	intcVIF0      = 4
	intcVIF1      = 5
	intcVU0       = 6
	intcVU1       = 7
	intcIPU       = 8
	intcTimer0    = 9
	intcTimer1    = 10
)

// intrExitAddr is the sentinel return address a guest call is given. The run loop
// traps a PC landing on it and unwinds, which is how callGuest knows the handler has
// returned.
const intrExitAddr = 0x0F000000

type handler struct {
	cause uint32
	addr  uint32 // the guest routine
	arg   uint32
	next  uint32
}

func (m *Machine) addIntcHandler() {
	cause, addr, _, arg := m.arg(0), m.arg(1), m.arg(2), m.arg(3)
	m.intcHandlers = append(m.intcHandlers, handler{cause: cause, addr: addr, arg: arg})
	m.note("AddIntcHandler cause=%d -> %s", cause, m.Sym(addr))
	m.setRet(uint32(len(m.intcHandlers)))
}

func (m *Machine) addDmacHandler() {
	cause, addr, _, arg := m.arg(0), m.arg(1), m.arg(2), m.arg(3)
	m.dmacHandlers = append(m.dmacHandlers, handler{cause: cause, addr: addr, arg: arg})
	m.note("AddDmacHandler channel=%d -> %s", cause, m.Sym(addr))

	// Channel 5 is SIF0, the IOP-to-EE direction. The routine registered on it is how
	// a command packet from the IOP gets read, so the fake IOP needs to know it: it is
	// the only way anything this machine invents reaches the game.
	if cause == dmacChannelSIF0 {
		m.sifCmdHandler = addr
	}
	m.setRet(uint32(len(m.dmacHandlers)))
}

// dmacChannelSIF0 is the DMA channel carrying packets from the IOP to the EE.
const dmacChannelSIF0 = 5

// deliverVBlank advances the frame clock and runs whatever the game registered for
// the vertical blank.
//
// The vsync flag deserves a note. SetVSyncFlag hands the kernel two addresses, and
// the kernel's own vertical-blank handler writes 1 to them each frame; the game's
// idle loop spins reading one and clears it. A model that registered the game's
// handlers but forgot the flag leaves that loop spinning forever, having done
// everything else right.
func (m *Machine) deliverVBlank() {
	m.vblanks++
	m.gsVSync()

	// The blank's leading edge lands in I_STAT whether or not anything is watching. A
	// game that paces itself by polling I_STAT (Ridge Racer V's main loop, rather than
	// any handler) reads this bit and acknowledges it with a write-1-to-clear; before
	// it existed the poll read back the game's own acknowledgement and every frame
	// looked like a fresh blank — the game free-ran at forty "frames" a blank. The
	// trailing edge (intcVBlankOff) is raised half a period later by the run loop:
	// raising both at one instant reads as two edges per poll, and a game pacing "on
	// then off" runs its per-blank work twice and overruns its own buffers.
	//
	// The bit is RECORDED, not asserted: this kernel is high-level emulated and runs
	// handlers directly (the loop below), so raising the CPU's interrupt line would
	// vector it through an exception table nothing has filled in. I_STAT is the
	// polling surface only.
	m.intcStat |= 1 << intcVBlankOn

	// The frame boundary, for a caller that wants to advance a field at a time (the
	// frame debugger). It runs before the guest's own vblank handlers so a run stopped
	// here has not yet started the next field's work.
	if m.OnVBlank != nil {
		m.OnVBlank(m)
	}

	// The blank is one event on the board and both processors see it. The IOP's
	// vblank library (iopvblank.go) runs its registered handlers off this same edge —
	// padman's per-frame pad poll above all.
	if m.IOP != nil {
		m.IOP.vblankTick()
	}

	if m.vsyncFlagPtr != 0 {
		m.Write32(m.vsyncFlagPtr, 1)
	}
	if m.vsyncFlag2Ptr != 0 {
		m.Write32(m.vsyncFlag2Ptr, 1)
	}

	for _, h := range m.intcHandlers {
		if h.cause != intcVBlankOn && h.cause != intcVBlankOff {
			continue
		}
		if m.intcMask&(1<<h.cause) == 0 {
			continue
		}
		m.callGuest(h.addr, h.arg)
	}
	// The kernel's interrupt epilogue: a handler may have woken a thread that outranks
	// the interrupted one, and it runs now, not at the next syscall.
	m.preemptIfOutranked()
}

// callGuest runs a guest routine to completion on the current stack and returns its
// $v0. The core's whole architectural state is saved and restored around it, so the
// interrupted program cannot tell it happened — which is the contract an interrupt
// handler has to honour.
func (m *Machine) callGuest(entry uint32, args ...uint32) uint32 {
	if entry == 0 || m.Halted {
		return 0
	}
	saved := m.CPU.Snapshot()

	for i, a := range args {
		if i < 4 {
			m.CPU.SetReg(uint32(4+i), uint64(int64(int32(a))))
		}
	}
	// A scratch stack below the current one, so the handler cannot tread on the
	// interrupted frame.
	m.CPU.SetReg(29, saved.R[29].Lo-0x800)
	m.CPU.SetReg(31, intrExitAddr) // $ra: the sentinel the loop watches for
	m.CPU.SetPC(uint64(entry))

	const budget = 8 << 20
	var ret uint64
	done := false
	for i := 0; i < budget; i++ {
		pc := uint32(m.CPU.PC)
		if pc == intrExitAddr {
			ret = m.CPU.Reg(2)
			done = true
			break
		}
		if m.CPU.Halted || m.Halted {
			break
		}
		// The instrumentation hook runs here too. A handler the oracle cannot trace is
		// an instrument with a hole in it: -trace and -logpc would go quiet for exactly
		// the code that runs on an interrupt, which is the code hardest to reason about
		// from a listing.
		if m.OnStep != nil {
			m.OnStep(m, pc)
		}
		// A fault inside a handler is silent otherwise: the CPU vectors to an empty
		// exception vector, decodes zeroes as nops, and slides until the budget runs out.
		// The handler simply "does nothing", which is indistinguishable from a handler
		// that had nothing to do.
		if r, faulted := m.unhandledException(pc); faulted {
			m.note("callGuest: %s faulted — %s", m.Sym(entry), r)
			break
		}
		if !m.mapped(phys(pc)) {
			m.note("callGuest: %s left mapped memory at 0x%08X", m.Sym(entry), pc)
			break
		}
		m.CPU.Step()
		m.steps++
	}
	if !done {
		m.note("callGuest: %s did not return within its budget", m.Sym(entry))
	}

	m.CPU.Restore(saved)
	return uint32(ret)
}

// raiseINTC is how a modelled peripheral says it has something to report. Until the
// peripherals exist, only the vertical blank uses this path.
func (m *Machine) raiseINTC(cause uint32) {
	m.intcStat |= 1 << cause
	m.CPU.Interrupt(m.intcStat&m.intcMask != 0, m.dmacStat&m.dmacMask != 0)
}

var _ = r5900.CauseIP2 // the interrupt lines the CPU exposes; see run.go
