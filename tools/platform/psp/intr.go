package psp

// intr.go models the PSP's sub-interrupt handlers, enough to deliver the display
// VBlank: a game registers a handler with sceKernelRegisterSubIntrHandler(intno,
// subno, handler, arg) and enables it; the kernel then calls handler(subno, arg) on
// each interrupt. Games pace their frame loop on that callback (a counter the
// handler bumps, polled inside a CpuSuspendIntr bracket), so without delivery the
// boot spins forever after graphics init.
//
// Delivery is synchronous: on a stepsPerVBlank cadence the run loop suspends the
// current thread's register state, runs the handler to completion on a scratch
// area below the current stack, and resumes — the cooperative-scheduler equivalent
// of an interrupt frame.

// vblankIntr is the PSP interrupt number of the display VBlank.
const vblankIntr = 30

// intrExitAddr is the sentinel $ra for a guest call made by the kernel itself
// (interrupt handlers, callbacks); distinct from threadExitAddr so a handler
// return is not mistaken for a thread exit.
const intrExitAddr = 0x0F100000

type subIntr struct {
	intno, subno uint32
	handler, arg uint32
	enabled      bool
}

// registerSubIntr and friends implement the InterruptManager syscalls.
func (m *Machine) registerSubIntr(intno, subno, handler, arg uint32) {
	m.subIntrs[intno<<16|subno] = &subIntr{intno: intno, subno: subno, handler: handler, arg: arg}
}

func (m *Machine) setSubIntrEnabled(intno, subno uint32, on bool) {
	if si := m.subIntrs[intno<<16|subno]; si != nil {
		si.enabled = on
	}
}

// deliverVBlank calls every enabled VBlank sub-interrupt handler and releases
// threads whose timed wait has expired.
func (m *Machine) deliverVBlank() {
	m.vblanks++
	for len(m.padScript) > 0 && m.vblanks >= m.padScript[0].AtVblank {
		m.pad = m.padScript[0].Buttons
		m.padScript = m.padScript[1:]
	}
	for _, o := range m.handles {
		if o.kind == "thread" && o.tstate == thWaiting && o.wakeVblank != 0 &&
			m.vblanks >= o.wakeVblank {
			o.wakeVblank = 0
			o.tstate = thReady
		}
	}
	for _, si := range m.subIntrs {
		if si.enabled && si.intno == vblankIntr {
			m.callGuest(si.handler, si.subno, si.arg)
		}
	}
}

// callGuest runs a guest function to completion in a nested execution frame: the
// live CPU state is saved, the function runs with $ra at the intrExitAddr
// sentinel on a scratch stack below the current one, and the state is restored.
// A CPU halt inside the handler (an unimplemented op) is propagated. The
// function's $v0 is returned (callback contracts like the sceMpeg ringbuffer
// feeder report a count there).
func (m *Machine) callGuest(entry uint32, args ...uint32) uint32 {
	saved := m.CPU.SaveState()
	m.CPU.SetPC(entry)
	for i, v := range args {
		m.CPU.SetReg(uint32(4+i), v)
	}
	sp := saved.R[29]
	if sp == 0 || sp == intrExitAddr {
		sp = stackTop
	}
	m.CPU.SetReg(29, (sp-0x800)&^0xF)
	m.CPU.SetReg(31, intrExitAddr)

	const budget = 4_000_000 // a handler is short; this bounds a runaway
	for i := 0; i < budget; i++ {
		if m.CPU.PC == intrExitAddr {
			break
		}
		m.CPU.Step()
		if m.CPU.Halted || m.Halted {
			break
		}
	}
	if m.CPU.PC != intrExitAddr && !m.CPU.Halted && !m.Halted {
		m.note("guest call 0x%08X did not return within budget", entry)
	}
	halted, reason := m.CPU.Halted, m.CPU.HaltReason
	ret := m.CPU.Reg(2)
	m.CPU.LoadState(saved)
	if halted {
		m.CPU.Halted, m.CPU.HaltReason = true, reason
	}
	return ret
}
