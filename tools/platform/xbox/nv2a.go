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
//	PFIFO  0x002000  the command FIFO; CACHE1_DMA_PUT at 0x003220, GET at 0x003210
//	PRAMIN 0x700000  instance memory (RAMHT/RAMFC) — D3D sets up channel contexts here
//	USER   0x800000  per-channel FIFO submission area; the user-space DMA PUT alias
//
// The first write that advances DMA_PUT past DMA_GET is the first push-buffer kick.

const (
	nvApertureSize = 1 << 24 // 16 MB

	nvPMC_ID    = 0x000000 // PMC_BOOT_0: chip id/revision
	nvPFIFO_PUT = 0x003220 // PFIFO CACHE1_DMA_PUT
	nvPFIFO_GET = 0x003210 // PFIFO CACHE1_DMA_GET
	nvUSER_PUT  = 0x800040 // USER channel 0 DMA PUT alias (D3D kicks here)
	nvUSER_GET  = 0x800044 // USER channel 0 DMA GET alias
)

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
	case nvPFIFO_GET, nvUSER_GET:
		dw = m.nv.dmaGet
	case nvPFIFO_PUT, nvUSER_PUT:
		dw = m.nv.dmaPut
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
	switch off &^ 3 {
	case nvPFIFO_PUT, nvUSER_PUT:
		m.nv.dmaPut = m.nv.reg[idx]
		if !m.nv.kicked && m.nv.dmaPut != m.nv.dmaGet {
			m.nv.kicked = true
			m.firstPush = true
			m.logf("NV2A: first push-buffer kick — DMA_PUT=%08X (GET=%08X) at PC %08X",
				m.nv.dmaPut, m.nv.dmaGet, m.CPU.LinearPC())
		}
	case nvPFIFO_GET, nvUSER_GET:
		m.nv.dmaGet = m.nv.reg[idx]
	}
}

// FirstPushReached reports whether the title has kicked the NV2A push buffer — the
// Phase-B milestone.
func (m *Machine) FirstPushReached() bool { return m.firstPush }
