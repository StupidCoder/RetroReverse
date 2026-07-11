package n3ds

// gx.go emulates the GSP module's side of the GX command queue. On real hardware
// the GSP system process continuously polls the GX command FIFO in the shared
// memory an application registered, executes each command on the GPU (a memory
// fill, a display transfer, a P3D command list, a texture copy), and raises the
// matching completion interrupt. The application's GX command runner posts a
// command, marks it pending, and blocks until that interrupt clears the pending
// flag. Without a GSP-module counterpart the render loop stalls forever waiting
// for a completion that never comes.
//
// This models the GSP module's *dispatch and completion signalling* — it drains
// the queue and raises the right interrupt per command — but does not yet turn a
// P3D command list into pixels (that is the PICA200 GPU, a later phase). It is
// the exact analogue of the N64 RDP command reader, minus the rasteriser.

// The GX command queue lives at this offset in the GSP shared memory (GSP thread
// index 0). A 0x20-byte header, then up to 15 command slots of 0x20 bytes each.
const (
	gxQueueOff   = 0x800
	gxCmdStride  = 0x20
	gxMaxCmds    = 15
	gxHdrIndex   = 0 // byte: index of the first unprocessed command
	gxHdrCount   = 1 // byte: number of queued commands
)

// GX command ids (GXCommandId), in the low byte of a command's first word.
const (
	gxCmdRequestDMA      = 0
	gxCmdProcessCmdList  = 1
	gxCmdMemoryFill      = 2
	gxCmdDisplayTransfer = 3
	gxCmdTextureCopy     = 4
	gxCmdFlushCache      = 5
)

// GXRecord is one GX command as the game posted it: the raw 8 words of its FIFO
// slot, plus — for a ProcessCommandList — a copy of the PICA200 command buffer it
// points at. Recorded when GXCapture is on; this is the Phase 4 instrument-first
// step (log what the game actually submits before implementing the GPU).
type GXRecord struct {
	Instr uint64    // instruction count at capture
	Words [8]uint32 // the raw command slot
	Buf   []byte    // ProcessCommandList only: the command-list bytes at capture time
}

// GXLog returns the commands captured so far (GXCapture must be on).
func (m *Machine) GXLog() []GXRecord { return m.gxLog }

// captureGX snapshots one GX command slot, and for a ProcessCommandList also the
// command-list buffer it references, before the FIFO is drained. The buffer copy
// matters: the game reuses/overwrites list memory between frames, so the bytes
// must be taken at submission time.
func (m *Machine) captureGX(cmd uint32, id uint32) {
	var r GXRecord
	r.Instr = m.CPU.Instrs
	for i := uint32(0); i < 8; i++ {
		r.Words[i] = m.ReadWord(cmd + i*4)
	}
	if id == gxCmdProcessCmdList {
		addr, size := r.Words[1], r.Words[2]
		if size > 0 && size < 0x1000000 {
			r.Buf = make([]byte, size)
			for i := uint32(0); i < size; i++ {
				r.Buf[i] = m.Read(addr + i)
			}
		}
	}
	m.gxLog = append(m.gxLog, r)
}

// processGXQueue drains any commands the game has posted to the GX FIFO and
// raises their completion interrupts, so the render loop's per-command waits are
// released. Cheap to call speculatively: it returns immediately when the queue
// is empty.
func (m *Machine) processGXQueue() {
	if m.gspSharedAddr == 0 {
		return
	}
	hdr := m.gspSharedAddr + gxQueueOff
	count := m.Read(hdr + gxHdrCount)
	if count == 0 {
		return
	}
	idx := m.Read(hdr + gxHdrIndex)

	var raised []byte
	for i := byte(0); i < count; i++ {
		slot := (uint32(idx) + uint32(i)) % gxMaxCmds
		cmd := m.gspSharedAddr + gxQueueOff + gxCmdStride + slot*gxCmdStride
		id := uint32(m.Read(cmd)) & 0x1F
		if m.GXCapture {
			m.captureGX(cmd, id)
		}
		switch id {
		case gxCmdProcessCmdList:
			m.framesSubmitted++ // a PICA200 command list — a rendered frame's geometry
			raised = append(raised, gspIntP3D)
		case gxCmdMemoryFill:
			// A fill can target one or both memory-fill engines; the second
			// address word being non-zero means PSC1 is used too.
			raised = append(raised, gspIntPSC0)
			if m.ReadWord(cmd+0xC) != 0 {
				raised = append(raised, gspIntPSC1)
			}
		case gxCmdDisplayTransfer, gxCmdTextureCopy:
			raised = append(raised, gspIntPPF)
		case gxCmdRequestDMA:
			raised = append(raised, gspIntDMA)
		case gxCmdFlushCache:
			// Cache maintenance completes synchronously; no interrupt.
		}
	}

	// The queue is drained: advance the read index and zero the count, exactly
	// as the GSP module does after servicing the batch.
	m.Write(hdr+gxHdrIndex, byte((uint32(idx)+uint32(count))%gxMaxCmds))
	m.Write(hdr+gxHdrCount, 0)

	for _, id := range raised {
		m.pushGSPInterrupt(id)
	}
	if len(raised) > 0 {
		m.signalGSPEvent()
	}
}
