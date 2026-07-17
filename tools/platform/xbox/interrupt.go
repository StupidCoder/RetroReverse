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

// irqSource is one device that can interrupt: the vector the title connected it on,
// and whether it is asserting right now (pending AND enabled, the device's own rule).
//
// The table exists because a second device arrived. Registration was always general —
// KeConnectInterrupt records whatever vector it is handed — but delivery knew only the
// NV2A's pending test and only the GPU's vector, which made "an interrupt" and "the
// vblank" the same sentence. They are not, and the USB controller is the proof.
type irqSource struct {
	name    string // named so a delivered interrupt can be logged as itself
	vector  uint32
	pending func(*Machine) bool
}

// irqSources are the machine's interrupt sources, in fixed priority order — fixed
// because delivery must not depend on map iteration if a savestate is to replay
// identically. Both vectors are pinned from the title's own KeConnectInterrupt rather
// than assumed (see gpuInterruptVector and usbInterruptVector).
var irqSources = []irqSource{
	{"nv2a", gpuInterruptVector, func(m *Machine) bool {
		return m.nv.pcrtcIntr&m.nv.reg[nvPCRTC_INTR_EN>>2]&1 != 0
	}},
	{"usb", usbInterruptVector, (*Machine).usbIRQ},
}

// irqPending reports whether any source is asserting, regardless of the CPU's gates.
// It is the retry predicate: a raise can land while the gates are shut, and the
// device's pending bit holds until its ISR acks it, so the tick handlers re-ask.
func (m *Machine) irqPending() bool {
	for i := range irqSources {
		if irqSources[i].pending(m) {
			return true
		}
	}
	return false
}

// deliverPending runs the connected ISR for the highest-priority enabled, pending
// interrupt source. Gates: one frame at a time, the CPU accepting interrupts (IF), and
// the kernel at PASSIVE_LEVEL (the KfRaiseIrql/KfLowerIrql pair tracks IRQL in the
// KPCR).
//
// One source per call, and the frame runs to completion before another is considered
// (isrActive). That is the hardware's own order, and it is also what keeps a second
// device from being able to starve the first: whatever is still pending when the ISR
// returns is re-offered by the next tick.
func (m *Machine) deliverPending() {
	if m.isrActive || m.CPU == nil || m.CPU.Halted || !m.CPU.IF {
		return
	}
	if m.Read(kpcrAddr+kpcrIrql) != 0 {
		return
	}
	for i := range irqSources {
		s := &irqSources[i]
		if !s.pending(m) {
			continue
		}
		ki, ok := m.interrupts[s.vector]
		if !ok {
			continue // the title has not connected this vector (yet)
		}
		routine := m.read32(ki + 0x00)
		if routine == 0 {
			continue
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
		return
	}
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

// DebugInterruptKI returns the KINTERRUPT guest address connected on a vector (0 if
// none) — a debugger/probe accessor, like DebugThreads.
func (m *Machine) DebugInterruptKI(vector uint32) uint32 { return m.interrupts[vector] }

// gpuInterruptVector is the vector the title's D3D runtime connects for the NV2A.
// HalGetInterruptVector returns its BusInterruptLevel argument as the vector, so
// this is the level the XDK passes for the GPU — pinned from the boot's own
// KeConnectInterrupt (the log line names every connected vector).
const gpuInterruptVector = 3

// usbInterruptVector is the vector XAPI connects for the USB OHCI host controller.
//
// Pinned the same way, and it had to be: this is the console fact most available to be
// remembered wrongly, and a plausible number here would have been indistinguishable
// from a correct one until the pad silently never arrived. `bootoracle -irqs` off the
// title's own savestate names it, and names its ISR too —
//
//	vector 1 -> KINTERRUPT 0064C2E0 (routine 00245DC2 ctx 800008C4)
//
// — whose routine sits in the same XAPI code region (0x24xxxx) as every OHCI register
// access the trace attributes, which is the corroboration that makes it evidence
// rather than a coincidence. The same command returns vector 3 for the NV2A, which is
// the method checking itself against a value that was already known.
const usbInterruptVector = 1
