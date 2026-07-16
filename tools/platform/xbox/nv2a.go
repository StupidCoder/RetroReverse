package xbox

// nv2a.go is a stand-in for the NV2A's MMIO aperture. The GPU itself — the PFIFO
// push-buffer parser, the vertex-program interpreter, the register combiners, the
// rasteriser — is Phase C. Phase B's milestone is only to *reach* the first
// push-buffer kick, so this file does the minimum: back the register aperture with a
// small store, answer the handful of reads the D3D8 init path makes so it does not
// wedge, and flag the moment the title first advances the PFIFO DMA PUT pointer — the
// "first push".
//
// The aperture is 16 MB at 0xFD000000. The engines that matter here:
//
//	PMC    0x000000  master control / enable (ID register read at boot)
//	PFIFO  0x002000  the command FIFO; CACHE1_DMA_PUT at 0x003240, GET at 0x003244
//	PRAMIN 0x700000  instance memory (RAMHT/RAMFC) — D3D sets up channel contexts here
//	USER   0x800000  per-channel FIFO submission area; the user-space DMA PUT alias
//
// The first write that advances DMA_PUT past DMA_GET is the first push-buffer kick.

import (
	"fmt"
	"os"
)

const (
	nvApertureSize = 1 << 24 // 16 MB

	nvPMC_ID = 0x000000 // PMC_BOOT_0: chip id/revision

	// PFIFO CACHE1 DMA engine. These are the REAL NV2A offsets (envytools/nouveau):
	// 0x3220 is DMA_PUSH (enable), NOT the put pointer — an earlier stub mislabelled it.
	nvPFIFO_RUNOUT    = 0x002400 // RUNOUT_STATUS: bit4 LOW_MARK (empty)
	nvPFIFO_C1_STATUS = 0x003214 // CACHE1_STATUS: bit4 LOW_MARK (empty), bit8 HIGH_MARK (full)
	nvPFIFO_DMA_PUSH  = 0x003220 // CACHE1_DMA_PUSH: bit0 enables the pusher
	nvPFIFO_DMA_PUT   = 0x003240 // CACHE1_DMA_PUT: the push-buffer write pointer
	nvPFIFO_DMA_GET   = 0x003244 // CACHE1_DMA_GET: the push-buffer read pointer

	// USER channel-0 control aliases (the D3D kickoff writes PUT here on the Xbox).
	nvUSER_PUT = 0x800040 // USER channel 0 DMA PUT alias
	nvUSER_GET = 0x800044 // USER channel 0 DMA GET alias

	// PFB (memory controller). The XDK flushes the write-combine buffer by setting bit 16
	// of 0x100410 and spinning until the hardware clears it; with one flat RAM backing the
	// flush is instantaneous, so we report the bit already clear (nvRead below).
	nvPFB_FLUSH = 0x100410

	// Interrupt-status registers. On the NV2A these are write-1-to-clear: the driver acks
	// an interrupt by writing 1s to the pending bits (the XDK writes 0xFFFFFFFF to clear
	// them all). Our engines run synchronously and raise no asynchronous interrupts, so
	// from the CPU's point of view no interrupt is ever pending — these read 0. Storing the
	// written value instead made PGRAPH_INTR read back 0xFFFFFFFF, which the driver read as
	// a perpetually-pending interrupt: it recursed through its ISR until the stack
	// overflowed into the D3D device object (a false "interrupt pending" is exactly the
	// kind of stub fiction that turns into a hang).
	nvPMC_INTR    = 0x000100 // PMC_INTR_0: master interrupt status
	nvPFIFO_INTR  = 0x002100 // PFIFO_INTR_0
	nvPGRAPH_INTR = 0x400100 // PGRAPH_INTR
	nvPCRTC_INTR  = 0x600100 // PCRTC_INTR_0 (vblank)

	// PGRAPH back-end semaphore progress. Never CPU-written (confirmed by an image
	// scan: only two sites reference it, both reads); the D3D busy-wait at 0x1AE550
	// polls it, comparing bits 2..6 against the in-memory semaphore value <<2, i.e.
	// against BACK_END_WRITE_SEMAPHORE_RELEASE <<2 — the Kelvin release handler
	// (nv2a_kelvin.go) keeps it current. Empirically modelled from those two call
	// sites; the value is release<<2.
	nvPGRAPH_SEMAPHORE = 0x400B10
)

var nvTrace = os.Getenv("RR_NV_TRACE") != ""

// nvRegName gives a short label for the interesting NV2A register offsets (trace only).
func nvRegName(off uint32) string {
	switch off &^ 3 {
	case nvPMC_ID:
		return "PMC_BOOT_0"
	case nvPFIFO_DMA_PUSH:
		return "PFIFO_CACHE1_DMA_PUSH"
	case nvPFIFO_DMA_PUT:
		return "PFIFO_CACHE1_DMA_PUT"
	case nvPFIFO_DMA_GET:
		return "PFIFO_CACHE1_DMA_GET"
	case nvUSER_PUT:
		return "USER_DMA_PUT"
	case nvUSER_GET:
		return "USER_DMA_GET"
	}
	switch {
	case off < 0x1000:
		return "PMC"
	case off >= 0x2000 && off < 0x4000:
		return "PFIFO"
	case off >= 0x400000 && off < 0x401000:
		return "PGRAPH"
	case off >= 0x600000 && off < 0x602000:
		return "PCRTC/PRMCIO"
	case off >= 0x700000 && off < 0x800000:
		return "PRAMIN"
	case off >= 0x800000:
		return "USER"
	}
	return "?"
}

// nv2a holds the aperture state the boot path reads back. Register writes are sparse
// (a title touches a few hundred of the 4 M dwords), so the backing is a map, not a
// 16 MB array — cheap to hold and to serialise into a savestate.
type nv2a struct {
	reg    map[uint32]uint32 // dword index -> value (sparse)
	dmaPut uint32
	dmaGet uint32
	kicked bool // DMA_PUT has been advanced at least once
}

// nvRead answers a byte read from the aperture (offset within 0xFD000000).
func (m *Machine) nvRead(off uint32) byte {
	if off >= nvApertureSize {
		return 0xFF
	}
	dw := m.nv.reg[off>>2]
	switch off &^ 3 {
	case nvPMC_ID:
		// PMC_BOOT_0: report an NV2A (revision A2). D3D reads this to confirm the part.
		dw = 0x02A000A2
	case nvPFIFO_DMA_GET, nvUSER_GET:
		dw = m.nv.dmaGet
	case nvPFIFO_DMA_PUT, nvUSER_PUT:
		dw = m.nv.dmaPut
	case nvPFB_FLUSH:
		dw &^= 0x10000 // flush completes instantly: bit 16 always reads back clear
	case nvPMC_INTR, nvPFIFO_INTR, nvPGRAPH_INTR, nvPCRTC_INTR:
		dw = 0 // no interrupt is ever pending in the synchronous model
	case nvPFIFO_C1_STATUS, nvPFIFO_RUNOUT:
		// The pusher drains the whole buffer synchronously (GET reaches PUT before this
		// read), so both the CACHE1 and the runout FIFOs are always empty: report LOW_MARK.
		// The XDK spins on these after a kick waiting for the FIFO to drain; without the
		// empty bit it never proceeds.
		dw |= 0x10
	}
	if nvTrace && (off&3) == 0 {
		fmt.Printf("NVrd %06X %-22s -> %08X  PC=%08X\n", off, nvRegName(off), dw, m.CPU.LinearPC())
	}
	return byte(dw >> (8 * (off & 3)))
}

// nvWrite records a byte write to the aperture and notices the first push.
func (m *Machine) nvWrite(off uint32, v byte) {
	if off >= nvApertureSize {
		return
	}
	idx := off >> 2
	shift := 8 * (off & 3)
	m.nv.reg[idx] = (m.nv.reg[idx] &^ (0xFF << shift)) | uint32(v)<<shift

	// A register's side effects fire once the whole 32-bit value has landed, not on each
	// byte. Writers always do aligned dword stores, so the top byte (off&3==3) completes
	// the register; acting per-byte instead kicked the pusher with a half-written PUT
	// pointer, which then ran away into the stack. Commit on the final byte only.
	if off&3 != 3 {
		return
	}
	if nvTrace {
		fmt.Printf("NVwr %06X %-22s <- %08X  PC=%08X\n", off&^3, nvRegName(off), m.nv.reg[idx], m.CPU.LinearPC())
	}
	switch off &^ 3 {
	case nvPFIFO_DMA_PUT, nvUSER_PUT:
		m.nv.dmaPut = m.nv.reg[idx]
		if !m.nv.kicked && m.nv.dmaPut != m.nv.dmaGet {
			m.nv.kicked = true
			m.firstPush = true
			m.logf("NV2A: first push-buffer kick — DMA_PUT=%08X (GET=%08X) at PC %08X",
				m.nv.dmaPut, m.nv.dmaGet, m.CPU.LinearPC())
		}
		// Writing DMA_PUT kicks the pusher: it drains the push buffer from GET to PUT,
		// dispatching every method to PGRAPH (nv2a_pfifo.go). Processing is synchronous —
		// GET catches up to PUT immediately — so a title polling GET never sees the ring
		// full and never stalls. Gated off until Phase C wiring is enabled so the Phase-B
		// firstPush milestone (and its savestate) still behave as before.
		if m.pusherEnabled {
			m.runPusher()
		}
	case nvPFIFO_DMA_GET, nvUSER_GET:
		m.nv.dmaGet = m.nv.reg[idx]
	}
}

// FirstPushReached reports whether the title has kicked the NV2A push buffer — the
// Phase-B milestone.
func (m *Machine) FirstPushReached() bool { return m.firstPush }
