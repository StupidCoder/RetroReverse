package n64

// ai.go is the Audio Interface: a DMA engine that streams PCM samples out of
// RDRAM to the DAC and interrupts when a buffer drains.
//
// Audio is deferred. The registers exist so that writes land somewhere and reads
// report an empty DMA queue, which is enough for the audio thread to believe its
// buffers were consumed. A blocked audio thread does not deadlock the console:
// libultra's scheduler simply runs the other threads, so the graphics path is
// unaffected. When audio is modelled, this file grows a sample queue and the AI
// interrupt joins MI.

const (
	aiDramAddr = 0x00
	aiLength   = 0x04
	aiControl  = 0x08
	aiStatus   = 0x0C
	aiDacRate  = 0x10
	aiBitRate  = 0x14
)

// AI_STATUS bits. Reporting the queue as neither full nor busy makes each DMA
// appear to complete instantly.
const (
	aiStatusBusy = 1 << 30
	aiStatusFull = 1 << 31
)

// Fields are exported so encoding/gob carries them into a save-state.
type ai struct {
	Regs regFile
}

func (a *ai) init() { a.Regs = regFile{} }

func (m *Machine) aiRead(addr uint32) uint32 {
	switch addr & 0xFF {
	case aiLength, aiStatus:
		return 0 // nothing queued, nothing playing
	}
	return m.ai.Regs[addr&0xFF]
}

func (m *Machine) aiWrite(addr uint32, v uint32) {
	switch addr & 0xFF {
	case aiStatus:
		m.clearIRQ(intrAI) // any write acknowledges
		return
	case aiLength:
		// The sample block is dropped, and the interrupt raised immediately: the
		// audio thread sees its buffer consumed and queues the next one.
		m.ai.Regs[aiLength] = v & 0x3FFF8
		m.raiseIRQ(intrAI)
		return
	}
	m.ai.Regs[addr&0xFF] = v
}
