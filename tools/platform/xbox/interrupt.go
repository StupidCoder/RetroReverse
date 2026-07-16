package xbox

// interrupt.go delivers hardware interrupts to the title's own interrupt service
// routines — the ISRs it registered through the canonical KeInitializeInterrupt /
// KeConnectInterrupt pair (kernel_objects.go records routine, context and vector
// into the KINTERRUPT and the vector->KINTERRUPT registry).
//
// The one modelled source is the display's vertical blank: the PCRTC raises its
// interrupt at 60 Hz, and the Direct3D runtime's ISR is what acknowledges it and
// signals the swap machinery (the KEVENT the swap path waits on with
// KeWaitForSingleObject). Running the title's own ISR — rather than inventing the
// event signal — keeps the whole flip protocol the game's code, not ours.
//
// Delivery runs the ISR as a nested frame ON the current thread, the way a real
// interrupt borrows the running context: the live CPU state is saved, the stdcall
// frame BOOLEAN (*ServiceRoutine)(PKINTERRUPT, PVOID) is pushed with a sentinel
// return address, and when the routine returns onto the sentinel the interrupted
// context resumes. IF is cleared for the ISR's duration; the scheduler does not
// switch threads while a frame is active (sched.go checks isrActive) — interrupts
// run to completion, like the hardware's.

import "retroreverse.com/tools/cpu/x86"

// isrExitAddr is the sentinel return address an ISR frame starts with; onStep
// (kernel.go) restores the interrupted context when the PC lands on it. Like
// threadExitAddr it is never fetched.
const isrExitAddr = trapBase - 0x200

// vblankPeriod is the vertical blank interval in machine ticks (60 Hz against the
// 2000-instructions-per-millisecond clock the other timers use).
const vblankPeriod = instrsPerMs * 1000 / 60

// vblankTick raises the PCRTC vertical-blank interrupt at 60 Hz, and retries
// delivery while a pending bit is up (a raise can land while the gates — IF, IRQL,
// an active frame — are closed; the bit holds until the ISR acks it). Called from
// schedTick's coarse block.
func (m *Machine) vblankTick() {
	if m.tick >= m.nextVBlank {
		m.nextVBlank = m.tick + vblankPeriod
		// PCRTC_INTR_0 bit 0 = VBLANK, write-1-to-clear (the pending bit stays up
		// until the ISR acks it; nv2a.go gives this register W1C semantics).
		m.nv.pcrtcIntr |= 1
	}
	if m.nv.pcrtcIntr != 0 {
		m.deliverPending()
	}
}

// deliverPending runs the connected ISR for any enabled, pending interrupt source.
// Gates: one frame at a time, the CPU accepting interrupts (IF), and the kernel at
// PASSIVE_LEVEL (the KfRaiseIrql/KfLowerIrql pair tracks IRQL in the KPCR).
func (m *Machine) deliverPending() {
	if m.isrActive || m.CPU == nil || m.CPU.Halted || !m.CPU.IF {
		return
	}
	if m.Read(kpcrAddr+kpcrIrql) != 0 {
		return
	}
	if m.nv.pcrtcIntr&m.nv.reg[nvPCRTC_INTR_EN>>2]&1 == 0 {
		return // vblank not pending or not enabled
	}
	ki, ok := m.interrupts[gpuInterruptVector]
	if !ok {
		return // the title has not connected the GPU interrupt (yet)
	}
	routine := m.read32(ki + 0x00)
	if routine == 0 {
		return
	}
	m.isrSaved = *m.CPU
	m.isrActive = true
	c := m.CPU
	push := func(v uint32) {
		c.Regs[x86.SP] -= 4
		m.write32(c.Regs[x86.SP], v)
	}
	push(m.read32(ki + 0x04)) // ServiceContext
	push(ki)                  // PKINTERRUPT
	push(isrExitAddr)
	c.IP = routine
	c.IF = false
}

// dpcEntry is one queued deferred procedure call (KeInsertQueueDpc): the KDPC's
// guest address plus the two system arguments. The routine and context live in
// the KDPC itself (KeInitializeDpc wrote them at +0x0C/+0x10).
type dpcEntry struct {
	Dpc, Arg1, Arg2 uint32
}

// isrReturn runs when an ISR or DPC frame's outermost RET lands on the sentinel:
// first any DPCs the ISR queued run (each as its own frame, the hardware's
// ISR-then-DPC order), then the interrupted context resumes. The
// retired-instruction count is kept (the frame's work happened).
func (m *Machine) isrReturn() {
	if len(m.dpcQueue) > 0 {
		d := m.dpcQueue[0]
		m.dpcQueue = m.dpcQueue[1:]
		routine := m.read32(d.Dpc + 0x0C)
		if routine != 0 {
			// Reuse the current frame: keep isrSaved, rebuild the stack for
			// DeferredRoutine(Dpc, DeferredContext, SystemArgument1, SystemArgument2).
			c := m.CPU
			push := func(v uint32) {
				c.Regs[x86.SP] -= 4
				m.write32(c.Regs[x86.SP], v)
			}
			push(d.Arg2)
			push(d.Arg1)
			push(m.read32(d.Dpc + 0x10)) // DeferredContext
			push(d.Dpc)
			push(isrExitAddr)
			c.IP = routine
			return
		}
	}
	steps := m.CPU.Steps
	*m.CPU = m.isrSaved
	m.CPU.Steps = steps
	m.isrActive = false
}

// gpuInterruptVector is the vector the title's D3D runtime connects for the NV2A.
// HalGetInterruptVector returns its BusInterruptLevel argument as the vector, so
// this is the level the XDK passes for the GPU — pinned from the boot's own
// KeConnectInterrupt (the log line names every connected vector).
const gpuInterruptVector = 3
