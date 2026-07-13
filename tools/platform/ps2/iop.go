package ps2

// iop.go is the PlayStation 2's second processor.
//
// The IOP is a MIPS R3000A with 2 MiB of its own memory — a PlayStation, near
// enough, kept on board to own the disc drive, the sound chip, the controllers and
// the memory cards. The EE talks to it across the SIF and asks it for things. Until
// now this machine answered those requests in Go (sif.go), because the IOP's base
// kernel lives in a BIOS ROM we do not have and will not take.
//
// The disc supplies the rest. IOPRP221.IMG carries the kernel modules a game reboots
// the IOP onto, and /DRIVERS/ carries the game's own — OVERLORD.IRX above all, which
// is where Naughty Dog put the disc streaming, the DGO loader and the sound. So the
// IOP here is real: a core executing the modules off the disc, with only the ROM's
// own base libraries standing in as Go (iopkernel.go).
//
// The division is worth stating plainly, because it is the whole design:
//
//	run for real   anything that is logic. SIFCMD, SIFMAN, THREADMAN, IOMAN,
//	               MODLOAD, LOADFILE, FILEIO, CDVDFSV, MCMAN, MCSERV, PADMAN,
//	               989SND, and OVERLORD. The game's protocol is then not something
//	               we have inferred; it is something it is doing.
//
//	stand in       anything that touches silicon we do not model: the interrupt
//	               controller, the allocator, the C library, the CD hardware, the
//	               SPU. All of it Sony's, none of it the game's.
//
// The 2 MiB of IOP memory is the same []byte the EE sees at 0x1C000000 — one buffer,
// two processors, which is what it is on the board too, and what makes a DMA between
// them a copy rather than a fiction.

import (
	"fmt"
	"sort"

	"retroreverse.com/tools/cpu/mips"
)

// The IOP's address map.
const (
	iopRAMSizeBytes = 2 << 20 // 2 MiB

	// The scratchpad: 1 KiB of fast memory in the D-cache's place, exactly as on the
	// PlayStation.
	iopSPRAMBase = 0x1F800000
	iopSPRAMSize = 0x400

	// The hardware registers: the interrupt controller, the DMA controller, the timers,
	// the SPU2, the SIO2, and the SIF's own doorbells.
	iopIOBase = 0x1F801000
	iopIOEnd  = 0x1FA00000
)

// iopModuleBase is where the first module is placed.
//
// The low 64 KiB is left alone: it is where the exception vectors live and where a
// real IOP keeps its kernel, and a module loaded over the top of them would work
// right up until the first interrupt.
const iopModuleBase = 0x00010000

// IOP is the second processor.
type IOP struct {
	ram []byte // the same slice the EE sees at 0x1C000000
	spr []byte

	CPU *mips.CPU
	ps2 *Machine // the machine it belongs to, for the SIF and the log

	// The modules loaded into it, and the libraries they provide (iopload.go).
	modules []*IOPModule
	exports map[string]*IRXExport // library name -> the export table serving it
	owner   map[string]string     // library name -> the module that exports it

	// The bump allocator that places modules and anything the kernel hands out.
	// The IOP has no MMU: an address is an address, and once given it is given.
	allocPtr uint32

	// The syscall table. A stub for a library we model in Go is patched to
	// `jr $ra; syscall n`, and n indexes this (iopkernel.go). bound remembers which
	// code was handed to which (library, function), so every stub for the same
	// function shares one.
	calls []*iopCall
	bound map[iopBinding]uint32

	// Whatever the IOP's modules have printed through stdio.
	tty []byte

	// The kernel HLE's state (iopintr.go): the interrupt controller, the allocator's
	// blocks, and the heaps.
	handlers    [iopIRQs]iopHandler
	intrEnabled bool
	blocks      []iopBlock
	heaps       map[uint32]*iopHeap

	// THREADMAN's two hooks into the interrupt-exit path: the predicate that says
	// whether a thread switch is wanted, and the routine that performs one. Registered
	// through intrman, and the contract the interrupt path will have to keep.
	schedSwitch  uint32
	schedResched uint32

	// The census: every call to a function nothing models, and every peripheral
	// register nobody has claimed. This is the work list, and it is the only honest
	// account of how much of the IOP is still missing.
	unmodelledCalls map[string]int
	io              map[uint32]uint32
	unmodelledIO    map[uint32]int

	// running is false until the first module is loaded. The machine does not step a
	// processor that has nothing on it.
	running bool

	Halted     bool
	HaltReason string
}

// newIOP makes the second processor over the memory the EE already shares with it.
func newIOP(m *Machine, ram []byte) *IOP {
	p := &IOP{
		ram:             ram,
		spr:             make([]byte, iopSPRAMSize),
		ps2:             m,
		exports:         map[string]*IRXExport{},
		owner:           map[string]string{},
		allocPtr:        iopModuleBase,
		unmodelledCalls: map[string]int{},
		io:              map[uint32]uint32{},
		unmodelledIO:    map[uint32]int{},
		heaps:           map[uint32]*iopHeap{},
	}
	p.CPU = mips.NewCPU(p)
	p.CPU.Syscall = p.handleSyscall
	p.registerLibraries()
	return p
}

// --- the bus ------------------------------------------------------------------

// iopPhys folds the KSEG0 and KSEG1 mirrors onto the physical address. The IOP runs
// its kernel through 0x80000000 and reaches hardware through 0xA0000000, and both are
// the same 2 MiB and the same registers.
func iopPhys(addr uint32) uint32 { return addr & 0x1FFFFFFF }

func (p *IOP) Read(addr uint32) byte {
	a := iopPhys(addr)
	switch {
	case a < iopRAMSizeBytes:
		return p.ram[a]
	case a >= iopSPRAMBase && a < iopSPRAMBase+iopSPRAMSize:
		return p.spr[a-iopSPRAMBase]
	}
	return byte(p.ioRead(a&^3) >> (8 * (a & 3)))
}

func (p *IOP) Write(addr uint32, v byte) {
	a := iopPhys(addr)
	switch {
	case a < iopRAMSizeBytes:
		p.ram[a] = v
		return
	case a >= iopSPRAMBase && a < iopSPRAMBase+iopSPRAMSize:
		p.spr[a-iopSPRAMBase] = v
		return
	}
	// A byte write to a register: merge it into the word. The word it merges into has to
	// be the register's *current* value — which for a register the other processor also
	// writes is not the same as the last value this one wrote. The R3000A's bus is
	// byte-wide, so every `sw` to a SIF doorbell arrives here four times, and a merge
	// against a stale word would drop half of it.
	sh := 8 * (a & 3)
	w := p.ioPeek(a&^3)&^(0xFF<<sh) | uint32(v)<<sh
	p.ioWrite(a&^3, w)
}

// ioPeek reads a register without counting the access. It is for the read half of a
// read-modify-write, which is not the guest asking for anything.
func (p *IOP) ioPeek(a uint32) uint32 {
	if a >= sbusIOPBase && a < sbusIOPBase+sbusSpan {
		return p.ps2.sbusRead(a - sbusIOPBase)
	}
	return p.io[a]
}

// Read32 and Write32 are for the machine's own use — the loader, the kernel HLE and
// the instruments. The core composes its own words a byte at a time.
//
// A register write must go through as *one word*, not as four bytes. The byte path
// merges its byte into the last value written to that address, and a shared register is
// one the other processor may have changed since — so four byte-writes to the SIF's
// doorbell would compose the new word out of a stale one and drop half of it. It is the
// kind of mistake that leaves a handshake that works locally and means nothing.
func (p *IOP) Read32(addr uint32) uint32 {
	a := iopPhys(addr)
	switch {
	case a+4 <= iopRAMSizeBytes:
		return uint32(p.ram[a]) | uint32(p.ram[a+1])<<8 | uint32(p.ram[a+2])<<16 | uint32(p.ram[a+3])<<24
	case a >= iopSPRAMBase && a+4 <= iopSPRAMBase+iopSPRAMSize:
		o := a - iopSPRAMBase
		return uint32(p.spr[o]) | uint32(p.spr[o+1])<<8 | uint32(p.spr[o+2])<<16 | uint32(p.spr[o+3])<<24
	}
	return p.ioRead(a &^ 3)
}

func (p *IOP) Write32(addr, v uint32) {
	a := iopPhys(addr)
	switch {
	case a+4 <= iopRAMSizeBytes:
		p.ram[a] = byte(v)
		p.ram[a+1] = byte(v >> 8)
		p.ram[a+2] = byte(v >> 16)
		p.ram[a+3] = byte(v >> 24)
	case a >= iopSPRAMBase && a+4 <= iopSPRAMBase+iopSPRAMSize:
		o := a - iopSPRAMBase
		p.spr[o] = byte(v)
		p.spr[o+1] = byte(v >> 8)
		p.spr[o+2] = byte(v >> 16)
		p.spr[o+3] = byte(v >> 24)
	default:
		p.ioWrite(a&^3, v)
	}
}

// CString reads a NUL-terminated string out of IOP memory.
func (p *IOP) CString(addr uint32) string {
	var b []byte
	for i := uint32(0); i < 1024; i++ {
		c := p.Read(addr + i)
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

// ioRead and ioWrite stand in for every IOP peripheral, and tally what was touched.
// The tally is the same instrument the EE's has: it says which silicon the modules
// actually drive, and therefore what to model next.
func (p *IOP) ioRead(a uint32) uint32 {
	if a >= sbusIOPBase && a < sbusIOPBase+sbusSpan {
		return p.ps2.sbusRead(a - sbusIOPBase)
	}
	p.unmodelledIO[a]++
	return p.io[a]
}

func (p *IOP) ioWrite(a, v uint32) {
	if a >= sbusIOPBase && a < sbusIOPBase+sbusSpan {
		p.ps2.sbusWrite(a-sbusIOPBase, v)
		return
	}
	p.unmodelledIO[a]++
	p.io[a] = v
}

// --- memory ---------------------------------------------------------------------

// alloc takes n bytes of IOP memory, 64-byte aligned.
//
// It is a bump allocator and it never frees. That is not a placeholder: an IOP boot
// loads its modules once and keeps them, and a kernel that cannot free is a kernel
// whose addresses stay put — which is worth a great deal when the thing being
// debugged is a machine you are still writing.
func (p *IOP) alloc(n uint32) uint32 {
	a := (p.allocPtr + 63) &^ 63
	p.allocPtr = a + n
	if p.allocPtr > iopRAMSizeBytes {
		p.halt("out of IOP memory: %d bytes wanted, past the end of the 2 MiB", n)
		return 0
	}
	return a
}

// --- running ---------------------------------------------------------------------

// iopStepRatio is how many EE instructions pass for each IOP instruction.
//
// The EE runs at about 294 MHz and the IOP at about 36 MHz, so the true ratio is
// near eight. It matters more than it looks: the two processors hand work to each
// other and then wait, and an IOP running at the wrong speed turns a handshake into
// either a spin or a race — neither of which is a bug in the code being debugged.
const iopStepRatio = 8

// Step runs the IOP for one instruction, if it has anything to run.
func (p *IOP) Step() {
	if !p.running || p.Halted || p.CPU.Halted {
		return
	}
	p.CPU.Step()
}

func (p *IOP) halt(format string, args ...interface{}) {
	p.Halted = true
	p.HaltReason = fmt.Sprintf(format, args...)
	p.ps2.note("IOP halted: %s", p.HaltReason)
}

// TTY returns whatever the IOP's modules have printed.
func (p *IOP) TTY() string { return string(p.tty) }

// DisasmAt renders IOP memory as instructions, as loaded — relocated, linked, and with
// every import stub already patched to whatever it was bound to.
//
// That last part is why this exists rather than a static disassembler over the file.
// Sony's kernel modules are stripped, so the only way to learn what `intrman #4` is is
// to read the code that calls it — and the code that calls it is only legible once you
// can see, at the call site, that the thing being called is a jump into THREADMAN's own
// interrupt handler rather than a number.
func (p *IOP) DisasmAt(addr uint32) string {
	code := []byte{p.Read(addr), p.Read(addr + 1), p.Read(addr + 2), p.Read(addr + 3)}
	in := mips.Decode(code, addr)

	// A call into the kernel HLE is a jump to a stub whose second word is a syscall.
	// Name it: `jal 0x00013abc` says nothing, `jal intrman.CpuSuspendIntr` says
	// everything.
	if in.HasTarget {
		if name := p.callName(in.Target); name != "" {
			return fmt.Sprintf("%-28s ; %s", in.Text, name)
		}
		if s := p.Sym(in.Target); s != "" {
			return fmt.Sprintf("%-28s ; %s", in.Text, s)
		}
	}
	return in.Text
}

// callName reports the kernel function a stub address stands for, if it is one.
func (p *IOP) callName(addr uint32) string {
	if p.Read32(addr) != insnJR(regRA()) {
		return ""
	}
	w := p.Read32(addr + 4)
	if w&0x3F != 0x0C {
		return ""
	}
	code := (w >> 6) & 0xFFFFF
	if code == 0 || int(code) >= len(p.calls) {
		return ""
	}
	return p.calls[code].name
}

// Modules returns the modules resident in IOP memory.
func (p *IOP) Modules() []*IOPModule { return p.modules }

// Sym names an address in IOP memory, using the symbol tables of the modules loaded
// there. Sony's kernel modules are stripped, but the game's are not: an address
// inside OVERLORD comes back as `ISOThread+0x1c`, which is the difference between
// reading a trace and decoding one.
func (p *IOP) Sym(addr uint32) string {
	for _, m := range p.modules {
		if addr < m.Base || addr >= m.Base+m.Size {
			continue
		}
		off := addr - m.Base
		var best *Symbol
		for i := range m.IRX.Symbols {
			s := &m.IRX.Symbols[i]
			if s.Func && s.Addr <= off && (best == nil || s.Addr > best.Addr) {
				best = s
			}
		}
		if best != nil {
			if d := off - best.Addr; d != 0 {
				return fmt.Sprintf("%s+0x%X", best.Name, d)
			}
			return best.Name
		}
		return fmt.Sprintf("%s+0x%X", m.Name, off)
	}
	return fmt.Sprintf("0x%08X", addr)
}

// --- the census -------------------------------------------------------------------

// IOPCensus reports what the IOP asked for that nothing answered: the kernel
// functions still unmodelled, and the hardware registers still unclaimed.
//
// It is the same instrument as the EE's syscall census and the hardware census, and
// it exists for the same reason. Modules resolve their imports by *index* — there
// are no names on the wire — so the only way to learn what the IOP kernel owes the
// disc is to let the disc ask, and write down what it asked for.
func (p *IOP) IOPCensus() string {
	if len(p.unmodelledCalls) == 0 && len(p.unmodelledIO) == 0 {
		return ""
	}
	s := "the IOP's unanswered requests (the work list):\n"

	if len(p.unmodelledCalls) > 0 {
		type kv struct {
			name string
			n    int
		}
		var all []kv
		for k, n := range p.unmodelledCalls {
			all = append(all, kv{k, n})
		}
		sort.Slice(all, func(i, j int) bool {
			if all[i].n != all[j].n {
				return all[i].n > all[j].n
			}
			return all[i].name < all[j].name
		})
		s += "  kernel functions nothing models:\n"
		for _, e := range all {
			s += fmt.Sprintf("      %-20s %d call%s\n", e.name, e.n, plural(e.n))
		}
	}

	if len(p.unmodelledIO) > 0 {
		var regs []uint32
		for a := range p.unmodelledIO {
			regs = append(regs, a)
		}
		sort.Slice(regs, func(i, j int) bool { return p.unmodelledIO[regs[i]] > p.unmodelledIO[regs[j]] })
		s += fmt.Sprintf("  IOP hardware touched: %d registers\n", len(regs))
		for i, a := range regs {
			if i == 8 {
				s += fmt.Sprintf("      ... and %d more\n", len(regs)-8)
				break
			}
			s += fmt.Sprintf("      0x%08X  %s  %d\n", a, iopRegionName(a), p.unmodelledIO[a])
		}
	}
	return s
}

// iopRegionName labels an IOP peripheral, so the census reads as hardware.
func iopRegionName(a uint32) string {
	switch {
	case a >= 0x1F801040 && a < 0x1F801060:
		return "SIO"
	case a >= 0x1F801070 && a < 0x1F801080:
		return "INTC"
	case a >= 0x1F801080 && a < 0x1F801100:
		return "DMA"
	case a >= 0x1F801100 && a < 0x1F801140:
		return "TIMER"
	case a >= 0x1F801450 && a < 0x1F801460:
		return "?"
	case a >= 0x1F801500 && a < 0x1F801580:
		return "DMA2"
	case a >= 0x1F802070 && a < 0x1F802080:
		return "POST"
	case a >= 0x1F808200 && a < 0x1F808300:
		return "SIO2"
	case a >= 0x1F900000 && a < 0x1FA00000:
		return "SPU2"
	case a >= 0x1D000000 && a < 0x1D000100:
		return "SIF"
	}
	return "?"
}
