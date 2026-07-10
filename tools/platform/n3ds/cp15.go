package n3ds

import "retroreverse.com/tools/cpu/arm"

// handleCP15 services the ARM11's system-control coprocessor accesses (MCR/MRC
// p15). Under a process-level HLE almost all of these are cache, TLB and barrier
// maintenance with no architectural effect on a core that models neither — those
// are accepted and ignored. The one that carries real information a runtime reads
// is the user read-only thread ID register (TPIDRURO, c13 c0 op2 3): the kernel
// programs it with the address of the current thread's TLS block, and libctru and
// official runtimes fetch the TLS base from it. We return the main thread's TLS
// page so that read resolves to real, mapped memory.
func (m *Machine) handleCP15(c *arm.CPU, load bool, cp, op1, crn, crm, op2 uint32, rd *uint32) {
	if cp != 15 {
		return
	}
	// TPIDRURO — user read-only thread ID (the TLS pointer).
	if crn == 13 && crm == 0 && op2 == 3 {
		if load {
			*rd = tlsBase
		}
		return
	}
	// TPIDRURW / TPIDRPRW (c13 c0 op2 2/4): a writable per-thread scratch word the
	// runtime may use; back it with a field so a written value reads back.
	if crn == 13 && crm == 0 && (op2 == 2 || op2 == 4) {
		if load {
			*rd = m.tpidr
		} else {
			m.tpidr = *rd
		}
		return
	}
	// Everything else (cache/TLB/branch-predictor maintenance, the control and
	// ID registers) has no effect here: a maintenance op is a no-op, and a read of
	// an ID/control register returns zero. Left silent by design — these fire
	// constantly and are not informative.
	if load {
		*rd = 0
	}
}
