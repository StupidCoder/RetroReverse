package threedo

// state.go snapshots the machine so an oracle can save a mid-run state and branch
// experiments from it, and so a frame debugger can replay a frame from its start over
// and over. The format is gzip+gob (the pattern the other platform oracles use); the
// source image's hash is pinned so a state cannot resume on a different game.
//
// This machine keeps more of itself outside RAM than any other in the repository. The
// Portfolio OS is high-level-emulated, so the kernel's item table, the cooperative task
// list, the two heaps, the open disc streams, the graphics folio's bitmaps and the audio
// clock are all Go objects on the side — invisible to a memory dump, and every one of
// them load-bearing. A savestate that carried only DRAM and VRAM would restore a machine
// whose game is mid-way through a file read that no longer has a file, waiting on a task
// that no longer exists.
//
// So everything below is a mirror: an exported, gob-encodable image of one piece of HLE
// state. Pointers are not serialized — the item table and the task list refer to each
// other by number, the way the OS itself does — and what can be re-derived is re-derived
// rather than stored: an open stream keeps its path and its cursor, and its contents are
// read back off the disc on load, because the disc has not changed and a copy of it in
// every savestate would be a lie waiting to drift.
//
// Host-side configuration (the hooks, the pad script, the debug flags, the logs) is not
// state and is not saved. Restoring a state does not re-arm a trace.

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/arm60"
)

// MachineState is everything the running game can observe.
type MachineState struct {
	ImageHash string

	DRAM []byte
	VRAM []byte
	CPU  arm60.CPUState

	DHeap heapState
	VHeap heapState

	// Kernel (kernel.go): the item table, and the type index rebuilt from it.
	Items      []itemState
	ItemByType map[uint32]int32 // item type -> item number
	NextItem   int32

	// The cooperative scheduler (task.go).
	Tasks    []taskEntryState
	Cur      int
	Switches int

	// The File folio (filefolio.go). A stream's bytes come back off the disc.
	Streams map[uint32]streamState
	Dirs    map[uint32]dirState
	NVRAM   map[string][]byte

	// The Graphics folio (graphicsfolio.go).
	Bitmaps    map[int32]bitmapState
	ScreenBM   map[int32]int32
	DisplayBuf uint32

	// The audio folio (audiofolio.go).
	AudioTime       uint32
	AudioEvents     []audioEventState
	AudioClockOwner int32

	// The event broker (io.go).
	EBListeners []int32

	SimTime   uint64
	VBlank    uint32
	VBLMirror uint32
	Frame     uint64
	TTY       []byte

	Halted     bool
	HaltReason string
}

type heapState struct {
	Base, Total uint32
	Free        []spanState
	Live        map[uint32]uint32
}

type spanState struct{ Addr, Size uint32 }

type itemState struct {
	Num       int32
	Typ       uint32
	Addr      uint32
	Tags      uint32
	Name      string
	Owner     int32
	Signal    uint32
	Msgs      []int32
	Device    int32
	ReplyPort int32
}

type taskEntryState struct {
	Num       int32
	Name      string
	Ctx       arm60.Context
	State     int
	Sig       uint32
	Wait      uint32
	AllocSigs uint32
}

// streamState is an open disc stream: which file, and how far in. The bytes are not
// saved — they are on the disc, which is the same disc, and re-reading them on load is
// both smaller and more honest than carrying a copy that could disagree with it.
type streamState struct {
	Name string
	Pos  int
}

type dirState struct {
	Entries []Entry
	Pos     int
}

type bitmapState struct {
	Buf  uint32
	W, H int
}

type audioEventState struct {
	Cue  int32
	Time uint32
}

// SetImageHash pins the disc a savestate belongs to.
func (m *Machine) SetImageHash(h string) { m.imageHash = h }

// SaveState captures the machine. Every map and slice is copied, so the snapshot is
// independent of the machine that made it and can be restored repeatedly — which is what
// a frame debugger's replay needs.
func (m *Machine) SaveState() MachineState {
	s := MachineState{
		ImageHash:       m.imageHash,
		DRAM:            append([]byte(nil), m.dram...),
		VRAM:            append([]byte(nil), m.vram...),
		CPU:             m.CPU.SaveState(),
		DHeap:           saveHeap(m.dheap),
		VHeap:           saveHeap(m.vheap),
		ItemByType:      map[uint32]int32{},
		NextItem:        m.nextItem,
		Cur:             m.cur,
		Switches:        m.switches,
		Streams:         map[uint32]streamState{},
		Dirs:            map[uint32]dirState{},
		NVRAM:           map[string][]byte{},
		Bitmaps:         map[int32]bitmapState{},
		ScreenBM:        map[int32]int32{},
		DisplayBuf:      m.displayBuf,
		AudioTime:       m.audioTime,
		AudioClockOwner: m.audioClockOwner,
		EBListeners:     append([]int32(nil), m.ebListeners...),
		SimTime:         m.simTime,
		VBlank:          m.vblank,
		VBLMirror:       m.vblMirror,
		Frame:           m.frame,
		TTY:             append([]byte(nil), m.tty...),
		Halted:          m.Halted,
		HaltReason:      m.HaltReason,
	}

	for _, it := range m.items {
		s.Items = append(s.Items, itemState{
			Num: it.num, Typ: it.typ, Addr: it.addr, Tags: it.tags, Name: it.name,
			Owner: it.owner, Signal: it.signal,
			Msgs:      append([]int32(nil), it.msgs...),
			Device:    it.device,
			ReplyPort: it.replyPort,
		})
	}
	// The type index is stored by item NUMBER, not by pointer: a pointer means nothing
	// once it has been through a file, and the OS itself refers to items by number.
	for typ, it := range m.itemByType {
		if it != nil {
			s.ItemByType[typ] = it.num
		}
	}
	for _, t := range m.tasks {
		s.Tasks = append(s.Tasks, taskEntryState{
			Num: t.num, Name: t.name, Ctx: t.ctx, State: int(t.state),
			Sig: t.sig, Wait: t.wait, AllocSigs: t.allocSigs,
		})
	}
	for h, st := range m.streams {
		s.Streams[h] = streamState{Name: st.name, Pos: st.pos}
	}
	for h, d := range m.dirs {
		s.Dirs[h] = dirState{Entries: append([]Entry(nil), d.entries...), Pos: d.pos}
	}
	for k, v := range m.nvram {
		s.NVRAM[k] = append([]byte(nil), v...)
	}
	for id, bm := range m.bitmaps {
		s.Bitmaps[id] = bitmapState{Buf: bm.buf, W: bm.w, H: bm.h}
	}
	for scr, bm := range m.screenBM {
		s.ScreenBM[scr] = bm
	}
	for _, ev := range m.audioEvents {
		s.AudioEvents = append(s.AudioEvents, audioEventState{Cue: ev.cue, Time: ev.time})
	}
	return s
}

// LoadState restores a state. The task list and the item table are rebuilt from numbers,
// and open streams are re-read from the mounted disc.
func (m *Machine) LoadState(s MachineState) error {
	if m.imageHash != "" && s.ImageHash != "" && m.imageHash != s.ImageHash {
		return fmt.Errorf("threedo: savestate image hash %s != current %s", s.ImageHash, m.imageHash)
	}
	copy(m.dram, s.DRAM)
	copy(m.vram, s.VRAM)
	m.CPU.LoadState(s.CPU)
	// A halt stops the run BEFORE the faulting instruction retires, so a state saved at
	// a halt resumes at that instruction. Clear the flag: with the cause fixed the run
	// continues; without it, it re-halts on the same word.
	m.CPU.Halted, m.CPU.HaltReason = false, ""
	m.Halted, m.HaltReason = false, ""

	m.dheap = loadHeap(s.DHeap)
	m.vheap = loadHeap(s.VHeap)

	m.items = map[int32]*item{}
	for _, it := range s.Items {
		m.items[it.Num] = &item{
			num: it.Num, typ: it.Typ, addr: it.Addr, tags: it.Tags, name: it.Name,
			owner: it.Owner, signal: it.Signal,
			msgs:      append([]int32(nil), it.Msgs...),
			device:    it.Device,
			replyPort: it.ReplyPort,
		}
	}
	m.itemByType = map[uint32]*item{}
	for typ, num := range s.ItemByType {
		if it, ok := m.items[num]; ok {
			m.itemByType[typ] = it
		}
	}
	m.nextItem = s.NextItem

	m.tasks = nil
	for _, t := range s.Tasks {
		m.tasks = append(m.tasks, &task{
			num: t.Num, name: t.Name, ctx: t.Ctx, state: taskState(t.State),
			sig: t.Sig, wait: t.Wait, allocSigs: t.AllocSigs,
		})
	}
	m.cur = s.Cur
	m.switches = s.Switches

	// Open streams come back off the disc. A stream whose file has gone is dropped and
	// said out loud, rather than resumed as an empty one that reads zeros for ever.
	m.streams = map[uint32]*diskStream{}
	for h, st := range s.Streams {
		data, path, ok := m.loadDiscFile(st.Name)
		if !ok {
			m.note(fmt.Sprintf("savestate stream %q dropped: not on this disc", st.Name))
			continue
		}
		m.streams[h] = &diskStream{name: path, data: data, pos: st.Pos}
	}
	m.dirs = map[uint32]*dirScan{}
	for h, d := range s.Dirs {
		m.dirs[h] = &dirScan{entries: append([]Entry(nil), d.Entries...), pos: d.Pos}
	}
	m.nvram = map[string][]byte{}
	for k, v := range s.NVRAM {
		m.nvram[k] = append([]byte(nil), v...)
	}

	m.bitmaps = map[int32]gfxBitmap{}
	for id, bm := range s.Bitmaps {
		m.bitmaps[id] = gfxBitmap{buf: bm.Buf, w: bm.W, h: bm.H}
	}
	m.screenBM = map[int32]int32{}
	for scr, bm := range s.ScreenBM {
		m.screenBM[scr] = bm
	}
	m.displayBuf = s.DisplayBuf

	m.audioTime = s.AudioTime
	m.audioEvents = nil
	for _, ev := range s.AudioEvents {
		m.audioEvents = append(m.audioEvents, audioEvent{cue: ev.Cue, time: ev.Time})
	}
	m.audioClockOwner = s.AudioClockOwner
	m.ebListeners = append([]int32(nil), s.EBListeners...)

	m.simTime, m.vblank, m.vblMirror, m.frame = s.SimTime, s.VBlank, s.VBLMirror, s.Frame
	m.tty = append([]byte(nil), s.TTY...)
	return nil
}

func saveHeap(h *heap) heapState {
	if h == nil {
		return heapState{}
	}
	out := heapState{Base: h.base, Total: h.total, Live: map[uint32]uint32{}}
	for _, sp := range h.free {
		out.Free = append(out.Free, spanState{sp.addr, sp.size})
	}
	for addr, size := range h.live {
		out.Live[addr] = size
	}
	return out
}

func loadHeap(s heapState) *heap {
	h := &heap{base: s.Base, total: s.Total, live: map[uint32]uint32{}}
	for _, sp := range s.Free {
		h.free = append(h.free, span{sp.Addr, sp.Size})
	}
	for addr, size := range s.Live {
		h.live[addr] = size
	}
	return h
}

// SaveStateFile writes a gzip+gob snapshot.
func (m *Machine) SaveStateFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := gzip.NewWriter(f)
	defer zw.Close()
	return gob.NewEncoder(zw).Encode(m.SaveState())
}

// LoadStateFile reads a snapshot written by SaveStateFile.
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
	return m.LoadState(s)
}
