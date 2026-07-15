package ps2

// iopstate.go snapshots the second processor.
//
// SaveStateFile has refused to run while the IOP was alive ever since the IOP became
// real, and the refusal was deliberate: the snapshot already carried the IOP's *memory*
// — it always did, because the EE can see it — and a state that carried the memory and
// not the processor would have restored into a machine that looked entirely right and
// was not. A second CPU with the correct RAM, no registers, no syscall bindings, no
// heaps and no interrupt handlers does not fail. It resumes, executes one instruction of
// somebody else's code, and goes wrong somewhere else entirely.
//
// So here is the whole of it. The rule of thumb throughout is the same as the EE's: what
// is *architectural* is serialised, and what is *host wiring* is re-attached. The one
// thing that cannot be either is the syscall table, because half of it is Go function
// pointers. That is rebuilt rather than carried, from the bindings — the (library,
// number) pairs the loader handed out codes for — which is the same information by a road
// gob can travel.

import (
	"fmt"

	"retroreverse.com/tools/cpu/mips"
)

// IOPBinding names one function of one library: the (library, number) pair a syscall code
// stands for. The code is the index into the slice.
type IOPBinding struct {
	Library string
	ID      uint16
}

// IOPModuleState is one module resident in IOP memory. The IRX comes with it — it is
// plain data, and it carries the module's symbols, which is what lets a restored machine
// still say `ISOThread+0x1c` instead of a number, and its export table, which is what lets
// a later module still link against it.
type IOPModuleState struct {
	Name string
	Base uint32
	Size uint32
	IRX  *IRX
}

// The small records the kernel HLE keeps.
type (
	IOPHandlerState struct{ Fn, Arg uint32 }
	IOPBlockState   struct{ Base, Size uint32 }
	// IOPHeapState carries the heap's *handle* as well as its state, and the two are not
	// the same number. CreateHeap returns the address of the first chunk and the guest holds
	// on to it forever; the heap then grows by taking fresh chunks elsewhere, and its base
	// moves. Key a restored heap by its base and every handle the guest is holding stops
	// resolving — and AllocHeapMemory from a heap that was never created is how THREADMAN
	// runs out of thread control blocks.
	IOPHeapState    struct{ Handle, Chunk, Base, Size, Ptr, Total uint32 }
	IOPDMAState     struct{ Madr, Bcr, Chcr, Tadr uint32 }
	IOPDMADoneState struct {
		At uint64
		Ch int
	}
	IOPTimerState struct {
		Count, Mode, Target uint32
		Fired               bool
	}
)

// IOPState is the second processor, entire.
type IOPState struct {
	CPU mips.CPUState
	SPR []byte

	// The IOP's 2 MiB is not in here. It is the same slice the EE sees at 0x1C000000, and
	// MachineState already carries it — once.

	Modules  []IOPModuleState
	AllocPtr uint32
	Bindings []IOPBinding // index = syscall code; index 0 is the one never handed out

	TTY []byte

	Handlers    [iopIRQs]IOPHandlerState
	IMask       uint64
	IntrEnabled bool
	Pending     uint64
	Blocks      []IOPBlockState
	Heaps       []IOPHeapState

	SchedSwitch, SchedResched uint32

	DMA                      [iopDMAChannels]IOPDMAState
	DPCR, DPCR2, DICR, DICR2 uint32
	DMAPending               []IOPDMADoneState
	Timers                   [iopTimers]IOPTimerState

	SPURegs []byte
	SPURAM  []byte

	Steps    uint64
	Switches int

	Raised    [iopIRQs]int
	Delivered [iopIRQs]int

	IO              map[uint32]uint32
	UnmodelledIO    map[uint32]int
	UnmodelledCalls map[string]int
	Prof            map[string]int

	Running    bool
	Halted     bool
	HaltReason string
}

// SaveState captures the second processor.
func (p *IOP) SaveState() IOPState {
	s := IOPState{
		CPU:             p.CPU.SaveState(),
		SPR:             append([]byte(nil), p.spr...),
		AllocPtr:        p.allocPtr,
		TTY:             append([]byte(nil), p.tty...),
		IMask:           p.imask,
		IntrEnabled:     p.intrEnabled,
		Pending:         p.pending,
		SchedSwitch:     p.schedSwitch,
		SchedResched:    p.schedResched,
		DPCR:            p.dpcr,
		DPCR2:           p.dpcr2,
		DICR:            p.dicr,
		DICR2:           p.dicr2,
		SPURegs:         append([]byte(nil), p.spu.regs...),
		SPURAM:          append([]byte(nil), p.spu.ram...),
		Steps:           p.steps,
		Switches:        p.switches,
		Raised:          p.raised,
		Delivered:       p.delivered,
		IO:              map[uint32]uint32{},
		UnmodelledIO:    map[uint32]int{},
		UnmodelledCalls: map[string]int{},
		Prof:            map[string]int{},
		Running:         p.running,
		Halted:          p.Halted,
		HaltReason:      p.HaltReason,
	}

	for _, m := range p.modules {
		s.Modules = append(s.Modules, IOPModuleState{Name: m.Name, Base: m.Base, Size: m.Size, IRX: m.IRX})
	}

	// The syscall table, as bindings. A code's *handler* is a Go function and cannot be
	// carried; the (library, number) it stands for can, and that is enough to build the
	// same table again — including the unmodelled entries, which have no handler at all and
	// exist only to be counted.
	s.Bindings = make([]IOPBinding, len(p.calls))
	for k, code := range p.bound {
		s.Bindings[code] = IOPBinding{Library: k.library, ID: k.id}
	}

	for i := range p.handlers {
		s.Handlers[i] = IOPHandlerState{Fn: p.handlers[i].fn, Arg: p.handlers[i].arg}
	}
	for _, b := range p.blocks {
		s.Blocks = append(s.Blocks, IOPBlockState{Base: b.base, Size: b.size})
	}
	for handle, h := range p.heaps {
		s.Heaps = append(s.Heaps, IOPHeapState{
			Handle: handle,
			Chunk:  h.chunk, Base: h.base, Size: h.size, Ptr: h.ptr, Total: h.total,
		})
	}
	for i := range p.dma {
		s.DMA[i] = IOPDMAState{p.dma[i].madr, p.dma[i].bcr, p.dma[i].chcr, p.dma[i].tadr}
	}
	for _, d := range p.dmaPending {
		s.DMAPending = append(s.DMAPending, IOPDMADoneState{At: d.at, Ch: d.ch})
	}
	for i := range p.timers {
		s.Timers[i] = IOPTimerState{
			Count: p.timers[i].count, Mode: p.timers[i].mode,
			Target: p.timers[i].target, Fired: p.timers[i].fired,
		}
	}
	for k, v := range p.io {
		s.IO[k] = v
	}
	for k, v := range p.unmodelledIO {
		s.UnmodelledIO[k] = v
	}
	for k, v := range p.unmodelledCalls {
		s.UnmodelledCalls[k] = v
	}
	for k, v := range p.prof {
		s.Prof[k] = v
	}
	return s
}

// LoadIOPState rebuilds the second processor from a snapshot, over the memory the EE
// already shares with it.
func (m *Machine) LoadIOPState(s IOPState) error {
	p := newIOP(m, m.iopRAM)
	m.IOP = p

	p.CPU.LoadState(s.CPU)
	copy(p.spr, s.SPR)
	p.allocPtr = s.AllocPtr
	p.tty = append([]byte(nil), s.TTY...)

	// The modules, and with them the two tables the linker keeps: who exports what, and
	// which module owns each library. They are rebuilt from the modules rather than carried
	// separately, because carrying a map of pointers *into* the modules would restore as a
	// map of pointers to copies — right until a later module linked against one of them and
	// got an export table nobody was updating.
	for i := range s.Modules {
		ms := &s.Modules[i]
		mod := &IOPModule{Name: ms.Name, Base: ms.Base, Size: ms.Size, IRX: ms.IRX}
		p.modules = append(p.modules, mod)
		for j := range mod.IRX.Exports {
			e := &mod.IRX.Exports[j]
			p.exports[e.Library] = e
			p.owner[e.Library] = mod.Name
		}
		// And the names of the import stubs, which is what the disassembler reads to say
		// `libsd#17` instead of an address in the middle of somebody's .text.
		for _, imp := range mod.IRX.Imports {
			for k, stub := range imp.Stubs {
				p.stubName[mod.Base+stub] = fmt.Sprintf("%s#%d", imp.Library, imp.IDs[k])
			}
		}
	}

	// The syscall table, rebuilt from the bindings. Every code gets the same number it had,
	// because the code is the index — which matters, since the codes are baked into the
	// `syscall` instructions sitting in the modules' patched stubs in the RAM we just
	// restored.
	p.calls = make([]*iopCall, len(s.Bindings))
	p.calls[0] = &iopCall{name: "<invalid>"}
	p.bound = map[iopBinding]uint32{}
	for code := 1; code < len(s.Bindings); code++ {
		b := s.Bindings[code]
		if b.Library == "" {
			p.calls[code] = &iopCall{name: "<invalid>"}
			continue
		}
		l := goLibraries[b.Library]
		if l == nil {
			return fmt.Errorf("ps2: the savestate binds syscall %d to library %q, which nothing models",
				code, b.Library)
		}
		f := l.funcs[b.ID]
		name := fmt.Sprintf("%s#%d", b.Library, b.ID)
		if f.name != "" {
			name = fmt.Sprintf("%s.%s", b.Library, f.name)
		}
		p.calls[code] = &iopCall{name: name, fn: f.fn}
		p.bound[iopBinding{b.Library, b.ID}] = uint32(code)
	}

	for i := range s.Handlers {
		p.handlers[i] = iopHandler{fn: s.Handlers[i].Fn, arg: s.Handlers[i].Arg}
	}
	p.imask = s.IMask
	p.intrEnabled = s.IntrEnabled
	p.pending = s.Pending
	p.schedSwitch, p.schedResched = s.SchedSwitch, s.SchedResched
	p.deriveSchedIsRun() // the is-running pointer, re-derived after a resume (the hook does not re-run)

	for _, b := range s.Blocks {
		p.blocks = append(p.blocks, iopBlock{base: b.Base, size: b.Size})
	}
	for _, h := range s.Heaps {
		p.heaps[h.Handle] = &iopHeap{chunk: h.Chunk, base: h.Base, size: h.Size, ptr: h.Ptr, total: h.Total}
	}

	for i := range s.DMA {
		p.dma[i] = iopDMAChan{s.DMA[i].Madr, s.DMA[i].Bcr, s.DMA[i].Chcr, s.DMA[i].Tadr}
	}
	p.dpcr, p.dpcr2, p.dicr, p.dicr2 = s.DPCR, s.DPCR2, s.DICR, s.DICR2
	for _, d := range s.DMAPending {
		p.dmaPending = append(p.dmaPending, iopDMADone{at: d.At, ch: d.Ch})
	}
	for i := range s.Timers {
		p.timers[i] = iopTimer{
			count: s.Timers[i].Count, mode: s.Timers[i].Mode,
			target: s.Timers[i].Target, fired: s.Timers[i].Fired,
		}
	}

	copy(p.spu.regs, s.SPURegs)
	copy(p.spu.ram, s.SPURAM)

	p.steps = s.Steps
	p.switches = s.Switches
	p.raised, p.delivered = s.Raised, s.Delivered

	for k, v := range s.IO {
		p.io[k] = v
	}
	for k, v := range s.UnmodelledIO {
		p.unmodelledIO[k] = v
	}
	for k, v := range s.UnmodelledCalls {
		p.unmodelledCalls[k] = v
	}
	if len(s.Prof) > 0 {
		p.prof = map[string]int{}
		for k, v := range s.Prof {
			p.prof[k] = v
		}
	}

	p.running = s.Running
	p.Halted, p.HaltReason = false, "" // a state saved at a halt must not re-halt on load
	if s.Halted {
		p.ps2.note("IOP: the snapshot was taken after the IOP halted: %s", s.HaltReason)
	}

	return nil
}
