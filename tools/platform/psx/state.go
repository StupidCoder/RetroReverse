package psx

// state.go — machine snapshot/restore. A savestate captures everything the
// game can observe (RAM, scratchpad, CPU+GTE, GPU incl. VRAM, CD controller,
// IRQ/DMA/pad/BIOS-HLE state) so an experiment can branch from a mid-run
// point — e.g. race start — without re-executing a 500M-instruction boot.
// Host-side configuration (hooks, PadScript, ISRHandler, the mounted disc)
// is deliberately NOT part of the state: a loaded state runs under whatever
// instrumentation the current host set up, on the same (immutable) disc.

import (
	"compress/gzip"
	"encoding/gob"
	"os"

	"retroreverse.com/tools/cpu/mips"
)

// MachineState is the serializable machine snapshot (gob+gzip on disk).
type MachineState struct {
	RAM, Scratch []byte
	CPU          mips.CPUState
	GTE          mips.GTEState

	IO       map[uint32]uint32
	IRQStat  uint32
	IRQMask  uint32
	Timer    uint32
	DMAFlags uint32

	VBlankAcc uint64
	ISRChain  uint32
	ISRActive bool
	ISRRetPC  uint32
	ISRRegs   [32]uint32
	ISRHi     uint32
	ISRLo     uint32

	HeapPtr, HeapEnd uint32
	NextEvent        uint32
	RandSeed         uint32

	PadBuf     uint32
	PadActive  bool
	PadButtons uint16
	PadCursor  int

	GPU gpuState
	CD  cdState
}

type gpuState struct {
	VRAM                   []uint16
	FIFO                   []uint32
	Need                   int
	ImgX, ImgY, ImgW, ImgH int
	ImgPx, ImgCurX, ImgCurY int
	RdX, RdY, RdW, RdH     int
	RdPx, RdCurX, RdCurY   int
	DrawL, DrawT, DrawR, DrawB int
	OffX, OffY             int
	TexPageX, TexPageY     int
	TexDepth               int
	TexWinMX, TexWinMY     int
	TexWinOX, TexWinOY     int
	DispX, DispY           int
	DispW, DispH           int
	DispEnabled            bool
	GP0Read                uint32
	StatField              uint32
}

type cdState struct {
	Index     byte
	Stat      byte
	Mode      byte
	Params    []byte
	Response  []byte
	Data      []byte
	DataPos   int
	IRQEnable byte
	IRQFlags  byte
	Loc       int
	ReadLBA   int
	Reading   bool
	Queue     []cdPending
}

type cdPending struct {
	Delay int
	Cause byte
	Resp  []byte
	Read  bool
}

// SaveState captures the machine.
func (m *Machine) SaveState() *MachineState {
	s := &MachineState{
		RAM:     append([]byte(nil), m.ram...),
		Scratch: append([]byte(nil), m.scratch...),
		CPU:     m.CPU.SaveState(),
		GTE:     m.GTE.SaveState(),
		IO:      map[uint32]uint32{},
		IRQStat: m.irqStat, IRQMask: m.irqMask,
		Timer: m.timer, DMAFlags: m.dmaFlags,
		VBlankAcc: m.vblankAcc, ISRChain: m.isrChain,
		ISRActive: m.isr.active, ISRRetPC: m.isr.retPC,
		ISRRegs: m.isr.R, ISRHi: m.isr.HI, ISRLo: m.isr.LO,
		HeapPtr: m.heapPtr, HeapEnd: m.heapEnd,
		NextEvent: m.nextEvent, RandSeed: m.randSeed,
		PadBuf: m.padBuf, PadActive: m.padActive,
		PadButtons: m.PadButtons, PadCursor: m.padCursor,
	}
	for k, v := range m.io {
		s.IO[k] = v
	}
	g := m.gpu
	s.GPU = gpuState{
		VRAM: append([]uint16(nil), g.vram...),
		FIFO: append([]uint32(nil), g.fifo...),
		Need: g.need,
		ImgX: g.imgX, ImgY: g.imgY, ImgW: g.imgW, ImgH: g.imgH,
		ImgPx: g.imgPx, ImgCurX: g.imgCurX, ImgCurY: g.imgCurY,
		RdX: g.rdX, RdY: g.rdY, RdW: g.rdW, RdH: g.rdH,
		RdPx: g.rdPx, RdCurX: g.rdCurX, RdCurY: g.rdCurY,
		DrawL: g.drawL, DrawT: g.drawT, DrawR: g.drawR, DrawB: g.drawB,
		OffX: g.offX, OffY: g.offY,
		TexPageX: g.texPageX, TexPageY: g.texPageY, TexDepth: g.texDepth,
		TexWinMX: g.texWinMX, TexWinMY: g.texWinMY,
		TexWinOX: g.texWinOX, TexWinOY: g.texWinOY,
		DispX: g.dispX, DispY: g.dispY, DispW: g.dispW, DispH: g.dispH,
		DispEnabled: g.dispEnabled, GP0Read: g.gp0Read, StatField: g.statField,
	}
	c := m.cd
	s.CD = cdState{
		Index: c.index, Stat: c.stat, Mode: c.mode,
		Params:   append([]byte(nil), c.params...),
		Response: append([]byte(nil), c.response...),
		Data:     append([]byte(nil), c.data...),
		DataPos:  c.dataPos, IRQEnable: c.irqEnable, IRQFlags: c.irqFlags,
		Loc: c.loc, ReadLBA: c.readLBA, Reading: c.reading,
	}
	for _, p := range c.queue {
		s.CD.Queue = append(s.CD.Queue, cdPending{
			Delay: p.delay, Cause: p.cause,
			Resp: append([]byte(nil), p.resp...), Read: p.read,
		})
	}
	return s
}

// LoadState restores a snapshot taken by SaveState. The machine keeps its
// current disc, hooks, PadScript and ISRHandler.
func (m *Machine) LoadState(s *MachineState) {
	copy(m.ram, s.RAM)
	copy(m.scratch, s.Scratch)
	m.CPU.LoadState(s.CPU)
	m.GTE.LoadState(s.GTE)
	m.io = map[uint32]uint32{}
	for k, v := range s.IO {
		m.io[k] = v
	}
	m.irqStat, m.irqMask = s.IRQStat, s.IRQMask
	m.timer, m.dmaFlags = s.Timer, s.DMAFlags
	m.vblankAcc, m.isrChain = s.VBlankAcc, s.ISRChain
	m.isr = isrState{active: s.ISRActive, retPC: s.ISRRetPC, R: s.ISRRegs, HI: s.ISRHi, LO: s.ISRLo}
	m.heapPtr, m.heapEnd = s.HeapPtr, s.HeapEnd
	m.nextEvent, m.randSeed = s.NextEvent, s.RandSeed
	m.padBuf, m.padActive = s.PadBuf, s.PadActive
	// PadButtons is restored, but the script cursor is NOT: PadScript is host
	// configuration, and a load is usually followed by a fresh script. Run's
	// apply loop fast-forwards past events whose step has already passed.
	m.PadButtons, m.padCursor = s.PadButtons, 0

	g := m.gpu
	copy(g.vram, s.GPU.VRAM)
	g.fifo = append(g.fifo[:0], s.GPU.FIFO...)
	g.need = s.GPU.Need
	g.imgX, g.imgY, g.imgW, g.imgH = s.GPU.ImgX, s.GPU.ImgY, s.GPU.ImgW, s.GPU.ImgH
	g.imgPx, g.imgCurX, g.imgCurY = s.GPU.ImgPx, s.GPU.ImgCurX, s.GPU.ImgCurY
	g.rdX, g.rdY, g.rdW, g.rdH = s.GPU.RdX, s.GPU.RdY, s.GPU.RdW, s.GPU.RdH
	g.rdPx, g.rdCurX, g.rdCurY = s.GPU.RdPx, s.GPU.RdCurX, s.GPU.RdCurY
	g.drawL, g.drawT, g.drawR, g.drawB = s.GPU.DrawL, s.GPU.DrawT, s.GPU.DrawR, s.GPU.DrawB
	g.offX, g.offY = s.GPU.OffX, s.GPU.OffY
	g.texPageX, g.texPageY, g.texDepth = s.GPU.TexPageX, s.GPU.TexPageY, s.GPU.TexDepth
	g.texWinMX, g.texWinMY = s.GPU.TexWinMX, s.GPU.TexWinMY
	g.texWinOX, g.texWinOY = s.GPU.TexWinOX, s.GPU.TexWinOY
	g.dispX, g.dispY, g.dispW, g.dispH = s.GPU.DispX, s.GPU.DispY, s.GPU.DispW, s.GPU.DispH
	g.dispEnabled, g.gp0Read, g.statField = s.GPU.DispEnabled, s.GPU.GP0Read, s.GPU.StatField

	c := m.cd
	c.index, c.stat, c.mode = s.CD.Index, s.CD.Stat, s.CD.Mode
	c.params = append(c.params[:0], s.CD.Params...)
	c.response = append(c.response[:0], s.CD.Response...)
	c.data = append(c.data[:0], s.CD.Data...)
	c.dataPos, c.irqEnable, c.irqFlags = s.CD.DataPos, s.CD.IRQEnable, s.CD.IRQFlags
	c.loc, c.readLBA, c.reading = s.CD.Loc, s.CD.ReadLBA, s.CD.Reading
	c.queue = c.queue[:0]
	for _, p := range s.CD.Queue {
		c.queue = append(c.queue, pending{delay: p.Delay, cause: p.Cause, resp: p.Resp, read: p.Read})
	}
}

// SaveStateFile writes the snapshot to path (gob, gzip-compressed).
func (m *Machine) SaveStateFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := gzip.NewWriter(f)
	if err := gob.NewEncoder(zw).Encode(m.SaveState()); err != nil {
		zw.Close()
		return err
	}
	return zw.Close()
}

// LoadStateFile restores a snapshot written by SaveStateFile.
func (m *Machine) LoadStateFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer zr.Close()
	var s MachineState
	if err := gob.NewDecoder(zr).Decode(&s); err != nil {
		return err
	}
	m.LoadState(&s)
	return nil
}
