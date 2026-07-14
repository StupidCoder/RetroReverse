package dsmachine

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/arm"
)

// Savestates.
//
// This is not a convenience, it is the instrument that makes the rest usable. A cold
// boot to SM64DS's title screen is about 1.2 billion scheduler steps; every question
// asked of the title screen — why is that polygon pink, what does the stylus do, what
// is in that VRAM bank — costs a minute of waiting before it can even be asked. A
// snapshot turns that into a second, and the repo's oracle-capability-parity rule
// exists because every platform that did this late wished it had done it first.
//
// What is captured is everything the machine's behaviour depends on and nothing that
// it does not. The cartridge image is NOT in the snapshot: it is the input, it is
// immutable, and a state file that carries 16 MiB of ROM is a state file nobody keeps.
// It is supplied again at load time, and the load checks that it is the same game.
//
// The subtle part is the CPUs. Their sixteen visible registers are the obvious thing
// to save and are not sufficient: the ARM banks R13/R14 per mode, and a DS takes an
// interrupt every scanline and runs a thread scheduler out of the user bank. Save only
// what is visible and the restored machine takes its next interrupt onto a stack
// pointer of zero — which does not look like a savestate bug, it looks like the game
// crashing a few frames later for no reason (arm.CPU.SaveBanks).

// snapshot is the whole machine, in a form gob can encode.
type snapshot struct {
	Version  int
	GameCode string // guards against restoring one game's state into another

	RAM   []byte
	SWRAM []byte
	Pal   []byte
	OAM   []byte

	VRAMBank [9][]byte
	VRAMCnt  [9]uint8

	Cores [2]coreState // [0] = ARM9, [1] = ARM7

	// The inter-processor block.
	Sync9, Sync7 uint8
	To7, To9     []uint32

	// The display, the maths units, and the odds and ends.
	Line    int
	HBlank  bool
	Frames  uint64
	Steps   uint64
	Powcnt  uint32
	Keys    uint32
	WRAMCnt uint8

	DivCnt             uint32
	DivNumer, DivDenom uint64
	DivResult, DivRem  uint64
	SqrtCnt            uint32
	SqrtParam          uint64
	SqrtResult         uint32

	// The cartridge port, mid-transfer.
	CardCmd  [8]byte
	CardCtrl uint32
	CardBuf  []byte
	CardPos  int

	// The ARM7's SPI bus. The firmware is generated rather than dumped, but a game
	// can WRITE to it (that is where its settings live), so it is part of the state.
	Firmware  []byte
	SPIDev    int
	SPIPhase  int
	SPICmd    byte
	SPIAddr   uint32
	SPIOut    byte
	SPIChan   int
	SPIResIdx int
	TouchX    int
	TouchY    int
	TouchDown bool

	// The 3D engine. The geometry engine's matrices and stacks are state a frame is
	// built on top of; the polygon list is rebuilt each frame and is captured anyway,
	// because a snapshot taken mid-frame must resume mid-frame.
	GXRegs     map[uint32]uint32
	GXFifo     []uint32
	GXPacked   []uint8
	GXCmd      uint8
	GXParams   []uint32
	GXNeed     int
	GXSwapPend bool
	GXSwapMode uint32
	Geom       geomState
	ThreeD     []uint32 // the 3D engine's last rendered frame, as the 2D engine sees it
	Screens    [2][]uint32
	ScreenSwap bool
}

// coreState is one CPU and the hardware that belongs to it alone.
type coreState struct {
	R          [16]uint32
	CPSR       uint32
	Banks      arm.Banks
	Halted     bool
	HaltReason string
	Instrs     uint64

	ITCM     []byte
	ITCMBase uint32
	DTCM     []byte
	DTCMBase uint32
	Low      []byte
	WRAM7    []byte

	IO   map[uint32]uint32
	DMA  [4]dmaChanState
	Time [4]timerState

	IME         bool
	IE, IF      uint32
	Waiting     bool
	WaitMask    uint32
	WaitAny     bool
	HandlerBase uint32
	LastRecv    uint32
	Sleep       int
}

type dmaChanState struct {
	Src, Dst, Ctrl, Count uint32
	CSrc, CDst, CRem      uint32
	Active                bool
}

type timerState struct {
	Counter, Reload uint16
	Ctrl            uint16
	Frac            int
}

// geomState is the geometry engine's persistent state: the matrices, their stacks,
// the lighting, and whatever primitive is half-assembled.
type geomState struct {
	Mode                int
	Proj, Pos, Vec, Tex [16]int32
	ProjStack           [1][16]int32
	TexStack            [1][16]int32
	PosStack            [32][16]int32
	VecStack            [32][16]int32
	SP, ProjSP, TexSP   int

	Begun    bool
	PrimMode int
	Strip    []vtxSave
	StripLen int
	LastVtx  [3]int32

	Color                [3]int32
	TexS, TexT           int32
	Attr, AttrNext       uint32
	TexParam, Pltt       uint32
	LightVec, LightColor [4][3]int32
	Diffuse, Ambient     [3]int32
	Specular, Emission   [3]int32
	Shininess            [128]uint8
	VX1, VY1, VX2, VY2   int32
	Polys                []polySave
	WBuffer              bool
}

// vtxSave and polySave are gob-encodable mirrors of the geometry engine's internal
// vertex and polygon. They exist only because gob will not touch unexported fields,
// and the alternative — exporting the renderer's hot inner types so that a savestate
// can see them — would be letting the serialiser design the rasteriser.
type vtxSave struct {
	X, Y, Z, W int64
	R, G, B    int32
	S, T       int32
}

type polySave struct {
	V        []vtxSave
	Attr     uint32
	TexParam uint32
	Pltt     uint32
	WBuffer  bool
}

func saveVerts(vs []gxVertex) []vtxSave {
	out := make([]vtxSave, len(vs))
	for i, v := range vs {
		out[i] = vtxSave{v.x, v.y, v.z, v.w, v.r, v.g, v.b, v.s, v.t}
	}
	return out
}

func loadVerts(vs []vtxSave) []gxVertex {
	out := make([]gxVertex, len(vs))
	for i, v := range vs {
		out[i] = gxVertex{x: v.X, y: v.Y, z: v.Z, w: v.W, r: v.R, g: v.G, b: v.B, s: v.S, t: v.T}
	}
	return out
}

// MachineState is an in-memory snapshot: the same state SaveState writes, kept in RAM
// rather than on disk.
//
// The frame debugger needs this and a file will not do. Its command scrubber replays
// one frame from its start state again and again — once per position the mouse drags
// over, on several scratch machines at once — and a round trip through gzip and gob for
// each of them would make the scrubber cost more than the emulation it is scrubbing.
type MachineState struct{ s *snapshot }

// SnapshotState captures the machine in memory. The result is independent of the live
// machine — a deep copy — so it can be restored repeatedly, into other machines, while
// this one keeps running. That independence is what makes deterministic replay possible
// at all.
func (m *Machine) SnapshotState() *MachineState { return &MachineState{s: m.snapshot()} }

// RestoreState puts an in-memory snapshot back.
func (m *Machine) RestoreState(ms *MachineState) error {
	if ms == nil || ms.s == nil {
		return fmt.Errorf("dsmachine: nil snapshot")
	}
	if got := m.gameCode(); ms.s.GameCode != got {
		return fmt.Errorf("dsmachine: snapshot is for game %q, this machine is running %q", ms.s.GameCode, got)
	}
	m.restore(ms.s)
	return nil
}

// GameCode reports the four-character cartridge code, so a caller can reject a foreign
// snapshot before trying to restore it.
func (m *Machine) GameCode() string { return m.gameCode() }

const stateVersion = 1

// SaveState writes the machine to a gzipped gob file.
func (m *Machine) SaveState(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z := gzip.NewWriter(f)
	if err := gob.NewEncoder(z).Encode(m.snapshot()); err != nil {
		return err
	}
	if err := z.Close(); err != nil {
		return err
	}
	return f.Close()
}

// LoadState restores a machine saved by SaveState. The machine must have been built
// from the same cartridge — the ROM is the one thing the snapshot does not carry.
func (m *Machine) LoadState(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	var s snapshot
	if err := gob.NewDecoder(z).Decode(&s); err != nil {
		return err
	}
	if s.Version != stateVersion {
		return fmt.Errorf("dsmachine: savestate version %d, want %d", s.Version, stateVersion)
	}
	if got := m.gameCode(); s.GameCode != got {
		return fmt.Errorf("dsmachine: savestate is for game %q, this machine is running %q", s.GameCode, got)
	}
	m.restore(&s)
	return nil
}

func (m *Machine) gameCode() string {
	if len(m.cd.rom) < 0x10 {
		return ""
	}
	return string(m.cd.rom[0x0C:0x10])
}

func (m *Machine) snapshot() *snapshot {
	s := &snapshot{
		Version: stateVersion, GameCode: m.gameCode(),
		RAM: clone(m.ram), SWRAM: clone(m.swram), Pal: clone(m.pal), OAM: clone(m.oam),
		Sync9: m.ipc.sync9, Sync7: m.ipc.sync7,
		To7: append([]uint32(nil), m.ipc.to7...), To9: append([]uint32(nil), m.ipc.to9...),
		Line: m.vid.line, HBlank: m.vid.hblank, Frames: m.vid.frames, Steps: m.Steps,
		Powcnt: m.powcnt, Keys: m.keys, WRAMCnt: m.wramcnt,
		DivCnt: m.div.cnt, DivNumer: m.div.numer, DivDenom: m.div.denom,
		DivResult: m.div.result, DivRem: m.div.rem,
		SqrtCnt: m.sqrt.cnt, SqrtParam: m.sqrt.param, SqrtResult: m.sqrt.result,
		CardCmd: m.cd.cmd, CardCtrl: m.cd.ctrl, CardBuf: clone(m.cd.buf), CardPos: m.cd.pos,
		Firmware: clone(m.spi.firmware), SPIDev: m.spi.dev, SPIPhase: m.spi.phase,
		SPICmd: m.spi.cmd, SPIAddr: m.spi.addr, SPIOut: m.spi.out,
		SPIChan: m.spi.chanSel, SPIResIdx: m.spi.resultIdx,
		TouchX: m.spi.touchX, TouchY: m.spi.touchY, TouchDown: m.spi.touchDown,
		GXFifo:   append([]uint32(nil), m.gpu3d.fifo...),
		GXPacked: append([]uint8(nil), m.gpu3d.packed...),
		GXCmd:    m.gpu3d.cmd, GXParams: append([]uint32(nil), m.gpu3d.params...),
		GXNeed: m.gpu3d.need, GXSwapPend: m.gpu3d.swapPending, GXSwapMode: m.gpu3d.swapMode,
		ThreeD:     append([]uint32(nil), m.gpu2d.threeD...),
		ScreenSwap: m.gpu2d.swap,
	}
	for i := range m.vram.bank {
		s.VRAMBank[i] = clone(m.vram.bank[i])
		s.VRAMCnt[i] = m.vram.cnt[i]
	}
	s.GXRegs = map[uint32]uint32{}
	for k, v := range m.gpu3d.regs {
		s.GXRegs[k] = v
	}
	s.Geom = toGeomState(&m.gpu3d.geom)
	s.Screens[0] = append([]uint32(nil), m.gpu2d.a.out...)
	s.Screens[1] = append([]uint32(nil), m.gpu2d.b.out...)
	s.Cores[0] = coreSnapshot(m.ARM9)
	s.Cores[1] = coreSnapshot(m.ARM7)
	return s
}

func coreSnapshot(c *core) coreState {
	cs := coreState{
		R: c.cpu.R, CPSR: c.cpu.CPSR(), Banks: c.cpu.SaveBanks(),
		Halted: c.cpu.Halted, HaltReason: c.cpu.HaltReason, Instrs: c.cpu.Instrs,
		ITCM: clone(c.itcm), ITCMBase: c.itcmBase,
		DTCM: clone(c.dtcm), DTCMBase: c.dtcmBase,
		Low: clone(c.low), WRAM7: clone(c.wram7),
		IO:  map[uint32]uint32{},
		IME: c.ime, IE: c.ie, IF: c.if_,
		Waiting: c.waiting, WaitMask: c.waitMask, WaitAny: c.waitAny,
		HandlerBase: c.handlerBase, LastRecv: c.lastRecv, Sleep: c.sleep,
	}
	for k, v := range c.io {
		cs.IO[k] = v
	}
	for i := range c.dma {
		d := c.dma[i]
		cs.DMA[i] = dmaChanState{d.src, d.dst, d.ctrl, d.count, d.csrc, d.cdst, d.crem, d.active}
	}
	for i := range c.timers {
		t := c.timers[i]
		cs.Time[i] = timerState{t.counter, t.reload, t.ctrl, t.frac}
	}
	return cs
}

func (m *Machine) restore(s *snapshot) {
	copy(m.ram, s.RAM)
	copy(m.swram, s.SWRAM)
	copy(m.pal, s.Pal)
	copy(m.oam, s.OAM)
	for i := range m.vram.bank {
		copy(m.vram.bank[i], s.VRAMBank[i])
		m.vram.cnt[i] = s.VRAMCnt[i]
	}
	m.vram.remap() // the page tables are derived, not stored: rebuild them from VRAMCNT

	m.ipc.sync9, m.ipc.sync7 = s.Sync9, s.Sync7
	m.ipc.to7 = append([]uint32(nil), s.To7...)
	m.ipc.to9 = append([]uint32(nil), s.To9...)

	m.vid.line, m.vid.hblank, m.vid.frames = s.Line, s.HBlank, s.Frames
	m.Steps, m.powcnt, m.keys, m.wramcnt = s.Steps, s.Powcnt, s.Keys, s.WRAMCnt

	m.div.cnt, m.div.numer, m.div.denom = s.DivCnt, s.DivNumer, s.DivDenom
	m.div.result, m.div.rem = s.DivResult, s.DivRem
	m.sqrt.cnt, m.sqrt.param, m.sqrt.result = s.SqrtCnt, s.SqrtParam, s.SqrtResult

	m.cd.cmd, m.cd.ctrl, m.cd.pos = s.CardCmd, s.CardCtrl, s.CardPos
	m.cd.buf = clone(s.CardBuf)

	copy(m.spi.firmware, s.Firmware)
	m.spi.dev, m.spi.phase, m.spi.cmd, m.spi.addr, m.spi.out = s.SPIDev, s.SPIPhase, s.SPICmd, s.SPIAddr, s.SPIOut
	m.spi.chanSel, m.spi.resultIdx = s.SPIChan, s.SPIResIdx
	m.spi.touchX, m.spi.touchY, m.spi.touchDown = s.TouchX, s.TouchY, s.TouchDown

	m.gpu3d.fifo = append([]uint32(nil), s.GXFifo...)
	m.gpu3d.packed = append([]uint8(nil), s.GXPacked...)
	m.gpu3d.cmd, m.gpu3d.need = s.GXCmd, s.GXNeed
	m.gpu3d.params = append([]uint32(nil), s.GXParams...)
	m.gpu3d.swapPending, m.gpu3d.swapMode = s.GXSwapPend, s.GXSwapMode
	m.gpu3d.regs = map[uint32]uint32{}
	for k, v := range s.GXRegs {
		m.gpu3d.regs[k] = v
	}
	fromGeomState(&m.gpu3d.geom, &s.Geom)

	m.gpu2d.threeD = append([]uint32(nil), s.ThreeD...)
	m.gpu2d.swap = s.ScreenSwap
	copy(m.gpu2d.a.out, s.Screens[0])
	copy(m.gpu2d.b.out, s.Screens[1])

	coreRestore(m.ARM9, &s.Cores[0])
	coreRestore(m.ARM7, &s.Cores[1])
}

func coreRestore(c *core, cs *coreState) {
	// The CPSR first: it sets the mode, and the mode decides which bank R13/R14 land
	// in. Restore the banks after it, then the visible registers last — they are the
	// current mode's, and must not be swapped out from under us by a later mode change.
	c.cpu.SetCPSR(cs.CPSR)
	c.cpu.RestoreBanks(cs.Banks)
	c.cpu.R = cs.R
	c.cpu.Halted, c.cpu.HaltReason, c.cpu.Instrs = cs.Halted, cs.HaltReason, cs.Instrs

	copy(c.itcm, cs.ITCM)
	c.itcmBase = cs.ITCMBase
	if cs.DTCM != nil {
		if c.dtcm == nil {
			c.dtcm = make([]byte, len(cs.DTCM))
		}
		copy(c.dtcm, cs.DTCM)
	}
	c.dtcmBase = cs.DTCMBase
	copy(c.low, cs.Low)
	copy(c.wram7, cs.WRAM7)

	c.io = map[uint32]uint32{}
	for k, v := range cs.IO {
		c.io[k] = v
	}
	for i := range c.dma {
		d := cs.DMA[i]
		c.dma[i] = dmaChan{src: d.Src, dst: d.Dst, ctrl: d.Ctrl, count: d.Count,
			csrc: d.CSrc, cdst: d.CDst, crem: d.CRem, active: d.Active}
	}
	for i := range c.timers {
		t := cs.Time[i]
		c.timers[i] = timer{counter: t.Counter, reload: t.Reload, ctrl: t.Ctrl, frac: t.Frac}
	}
	c.ime, c.ie, c.if_ = cs.IME, cs.IE, cs.IF
	c.waiting, c.waitMask, c.waitAny = cs.Waiting, cs.WaitMask, cs.WaitAny
	c.handlerBase, c.lastRecv, c.sleep = cs.HandlerBase, cs.LastRecv, cs.Sleep
}

func toGeomState(g *geom) geomState {
	gs := geomState{
		Mode: g.mode, Proj: g.proj.m, Pos: g.pos.m, Vec: g.vec.m, Tex: g.tex.m,
		SP: g.sp, ProjSP: g.projSP, TexSP: g.texSP,
		Begun: g.begun, PrimMode: g.primMode, StripLen: g.stripLen, LastVtx: g.lastVtx,
		Color: g.color, TexS: g.texS, TexT: g.texT,
		Attr: g.attr, AttrNext: g.attrNext, TexParam: g.texParam, Pltt: g.pltt,
		LightVec: g.lightVec, LightColor: g.lightColor,
		Diffuse: g.diffuse, Ambient: g.ambient, Specular: g.specular, Emission: g.emission,
		Shininess: g.shininess,
		VX1:       g.viewX1, VY1: g.viewY1, VX2: g.viewX2, VY2: g.viewY2,
		WBuffer: g.wbuffer,
		Strip:   saveVerts(g.strip),
	}
	gs.Polys = make([]polySave, len(g.polys))
	for i, p := range g.polys {
		gs.Polys[i] = polySave{saveVerts(p.verts), p.attr, p.texParam, p.pltt, p.wbuffer}
	}
	gs.ProjStack[0] = g.projStack[0].m
	gs.TexStack[0] = g.texStack[0].m
	for i := range g.posStack {
		gs.PosStack[i] = g.posStack[i].m
		gs.VecStack[i] = g.vecStack[i].m
	}
	return gs
}

func fromGeomState(g *geom, gs *geomState) {
	g.mode = gs.Mode
	g.proj.m, g.pos.m, g.vec.m, g.tex.m = gs.Proj, gs.Pos, gs.Vec, gs.Tex
	g.projStack[0].m = gs.ProjStack[0]
	g.texStack[0].m = gs.TexStack[0]
	for i := range g.posStack {
		g.posStack[i].m = gs.PosStack[i]
		g.vecStack[i].m = gs.VecStack[i]
	}
	g.sp, g.projSP, g.texSP = gs.SP, gs.ProjSP, gs.TexSP
	g.begun, g.primMode, g.stripLen, g.lastVtx = gs.Begun, gs.PrimMode, gs.StripLen, gs.LastVtx
	g.color, g.texS, g.texT = gs.Color, gs.TexS, gs.TexT
	g.attr, g.attrNext, g.texParam, g.pltt = gs.Attr, gs.AttrNext, gs.TexParam, gs.Pltt
	g.lightVec, g.lightColor = gs.LightVec, gs.LightColor
	g.diffuse, g.ambient, g.specular, g.emission = gs.Diffuse, gs.Ambient, gs.Specular, gs.Emission
	g.shininess = gs.Shininess
	g.viewX1, g.viewY1, g.viewX2, g.viewY2 = gs.VX1, gs.VY1, gs.VX2, gs.VY2
	g.wbuffer = gs.WBuffer
	g.strip = loadVerts(gs.Strip)
	g.polys = g.polys[:0]
	for _, p := range gs.Polys {
		g.polys = append(g.polys, gxPolygon{
			verts: loadVerts(p.V), attr: p.Attr, texParam: p.TexParam, pltt: p.Pltt, wbuffer: p.WBuffer,
		})
	}
	g.clipDirty = true // derived; recompute on first use rather than trust a stored copy
}

func clone(b []byte) []byte {
	if b == nil {
		return nil
	}
	return append([]byte(nil), b...)
}
