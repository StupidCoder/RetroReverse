package ps2

// iopload.go is loadcore: the IOP's linker.
//
// It is the one part of the IOP kernel that could never have been anything but ours,
// because it is the part that decides what "running a module" means here. Everything
// else in iopkernel.go is a stand-in for a routine in a ROM. This is the routine that
// puts the disc's own code into memory and makes it callable.
//
// Loading an IRX is four steps:
//
//	place        take a base address from the bump allocator and relocate the image
//	             onto it (irx.go). The IOP has no MMU; a module is patched to suit
//	             wherever it landed.
//
//	export       register every library the module provides, so the next module to
//	             import it can be wired straight to it.
//
//	import       walk every import stub and patch it. A library some loaded module
//	             exports becomes `j <address>` — a direct jump into the real code, the
//	             stub's own `jr $ra` never running because the callee returns for it.
//	             A library we model in Go becomes `jr $ra; syscall n`. There is no
//	             third case: a library that is neither is one nobody can serve, and
//	             the loader says so rather than leaving a stub that will look like a
//	             `nop` when it is called.
//
//	start        call the module's entry point, which registers its libraries, creates
//	             its threads and returns.
//
// The order matters, and it is not the order the disc lists the modules in. A module
// must be loaded after everything it imports, so the loader is given a dependency
// order and checks it — a module whose import cannot be resolved is a load error, not
// a mystery three million instructions later.

import "fmt"

// IOPModule is a module resident in IOP memory.
type IOPModule struct {
	Name string
	Base uint32 // where its image begins
	Size uint32
	IRX  *IRX
}

// The MIPS instructions the linker writes into a stub.
//
//	j target      the direct link: control leaves for the real function and returns
//	              from there to whoever called the stub.
//	jr $ra        the return half of an HLE stub.
//	syscall n     the call into Go, which runs in the `jr`'s delay slot.
func insnJ(target uint32) uint32  { return 0x08000000 | (target>>2)&0x03FFFFFF }
func insnSyscall(n uint32) uint32 { return (n << 6) | 0x0C }
func insnNop() uint32             { return 0 }
func insnJR(reg uint32) uint32    { return (reg << 21) | 0x08 }
func regRA() uint32               { return 31 }

// LoadIRX places a module in IOP memory, links it, and returns it without starting it.
//
// It is the machine's own loader — the one that boots IOPRP221.IMG's modules before the
// game's first instruction. It takes the base from its own bump allocator. The game's
// runtime loads go a different way (loadcoreLinkModule): the guest's own MODLOAD allocates
// the memory and calls loadcore to place the image on it. Both roads end at placeAndLink,
// because relocating an image and wiring its imports is the same job whoever picked the
// address.
func (p *IOP) LoadIRX(name string, raw []byte) (*IOPModule, error) {
	x, err := ReadIRX(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	base := p.alloc(x.MemSz)
	return p.placeAndLink(name, x, base)
}

// placeAndLink relocates a parsed module onto base, registers its exports, patches its
// imports, and adds it to the resident set. It is everything loading an IRX is except
// choosing where it goes — which LoadIRX does from its own allocator and MODLOAD does from
// sysmem, and which is the whole of the difference between them.
func (p *IOP) placeAndLink(name string, x *IRX, base uint32) (*IOPModule, error) {
	img, err := x.Relocate(base)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	copy(p.ram[base:], img)

	mod := &IOPModule{Name: name, Base: base, Size: x.MemSz, IRX: x}
	p.modules = append(p.modules, mod)

	// Export before import: a module may import from itself, and more usefully, the
	// module that provides a library must be findable the moment it is loaded.
	for i := range x.Exports {
		e := &x.Exports[i]
		if prev, ok := p.owner[e.Library]; ok && prev != name {
			p.ps2.note("IOP: %s re-exports %s, which %s already provides", name, e.Library, prev)
		}
		p.exports[e.Library] = e
		p.owner[e.Library] = name
	}

	if err := p.link(mod); err != nil {
		return nil, err
	}

	// A trap named for a symbol arms as soon as the module carrying that symbol is resident.
	if p.Trap == 0 && p.TrapSym != "" {
		if a, ok := p.SymAddr(p.TrapSym); ok {
			p.Trap = a
			p.ps2.note("IOP: the trap is armed at %s (0x%08X)", p.TrapSym, a)
		}
	}

	p.running = true
	return mod, nil
}

// link patches every import stub in the module.
func (p *IOP) link(mod *IOPModule) error {
	for _, imp := range mod.IRX.Imports {
		exp, onDisc := p.exports[imp.Library]

		for i, stub := range imp.Stubs {
			id := imp.IDs[i]
			addr := mod.Base + stub

			// Name the stub before patching it. A stub is not a function and no symbol
			// covers it, so a `jal` to one disassembles as an address in the middle of some
			// unrelated routine's .text — which is exactly the call whose identity we are
			// trying to establish. Recording (library, id) here is what lets DisasmAt print
			// `jal libsd#7` instead.
			p.stubName[addr] = fmt.Sprintf("%s#%d", imp.Library, id)

			switch {
			case onDisc:
				// The real module is resident. Resolution is by index: function id is the
				// id'th entry of the library's export table.
				if int(id) >= len(exp.Entries) {
					return fmt.Errorf("%s imports %s function %d, but %s's export table has only %d",
						mod.Name, imp.Library, id, p.owner[imp.Library], len(exp.Entries))
				}
				target := p.exportAddr(imp.Library, id)
				if target == 0 {
					return fmt.Errorf("%s imports %s function %d, which %s exports as a null pointer",
						mod.Name, imp.Library, id, p.owner[imp.Library])
				}
				p.Write32(addr+0, insnJ(target))
				p.Write32(addr+4, insnNop())

			case p.hasGoLibrary(imp.Library):
				// We model it. The stub becomes a call into Go: the `syscall` sits in the
				// `jr`'s delay slot, so it runs and then control returns to the caller with
				// whatever the handler left in $v0.
				code := p.bindCall(imp.Library, id)
				p.Write32(addr+0, insnJR(regRA()))
				p.Write32(addr+4, insnSyscall(code))

			default:
				return fmt.Errorf("%s imports library %q, which nothing on the disc exports and nothing in Go models",
					mod.Name, imp.Library)
			}
		}
	}
	return nil
}

// --- loadcore's module-load primitives, as MODLOAD drives them ------------------------
//
// The game loads its own IRX modules at runtime — SIO2MAN and the rest — and it does not do
// it the way the machine boots IOPRP221.IMG. The EE asks the IOP across the SIF, LOADFILE
// reads the file off the disc into a buffer, and then MODLOAD, running for real on the
// R3000A, links it. MODLOAD is Sony's, and it is logic, so it runs; the thing it calls to
// turn a buffer of ELF into resident code is loadcore, and loadcore is ours.
//
// The contract between them was read out of MODLOAD's own loader (MODLOAD+0xE9C), because
// the info struct they pass back and forth is loadcore's private format — MODLOAD only ever
// peeks at three of its fields:
//
//	#22(image, info)   probe. Fill info+28 with the memory the module needs, and return a
//	                   type code: 2 = relocatable, which steers MODLOAD to allocate the
//	                   memory itself and place the module wherever it lands.
//	MODLOAD            AllocSysMemory(0, info+28 + 48, 0) -> base, and writes base+48 into
//	                   info+12. The 48 bytes at [base, base+48) are the module's registration
//	                   record; the image goes just above it.
//	#23(image, info)   relocate the image onto info+12 (= base+48) and wire its imports.
//	#8(base+48, ...)   the link check. MODLOAD frees the block and fails with -200 if this
//	                   returns negative, so it is where an unresolvable import is reported —
//	                   but placeAndLink has already resolved them, and would have errored, so
//	                   here it only confirms.
//	#4()               flush the instruction cache. The machine has no cache to flush.
//	#16(base)          register the module. MODLOAD keeps its own list keyed on base and
//	                   returns base as the handle regardless of what this answers.
//
// info is loadcore's, so its layout is chosen here and only these offsets are load-bearing.

// The info struct MODLOAD and loadcore pass between them. Only three fields are read by
// MODLOAD; the rest is ours to use as scratch between #22 and #23.
const (
	iopModInfoInit = 0x0C // MODLOAD writes base+48 here for #23 to read
	iopModInfoText = 0x10 // #23 leaves the module base here, for #8
	iopModInfoSize = 0x1C // #22 leaves the module's memory size here, for MODLOAD's alloc
)

// The module registration record: the 48 bytes at [base, base+48) that MODLOAD keeps ahead
// of the image. It is loadcore's own structure, and MODLOAD reads exactly these fields — the
// entry point and $gp it starts the module with, and the pair it hands to the unlink routine
// if the module declines to stay resident. loadcore fills them; the offsets are MODLOAD's.
const (
	iopModRecEntry = 0x10 // the module's entry point, absolute
	iopModRecGP    = 0x14 // its $gp, absolute
	iopModRecArg0  = 0x18 // passed to loadcore#9 when a non-resident module unloads
	iopModRecArg1  = 0x1C
	iopModRecSize  = 48
)

// The type codes #22 returns. Only "relocatable" is exercised; a fixed-address module would
// return one of the others and steer MODLOAD down its other allocation path.
const (
	iopModRelocatable = 2
	iopModBadImage    = ^uint32(0) - 200 // -201: MODLOAD reports exactly this if the code is not 1..4
)

// loadcoreProbeModule is loadcore #22: examine the ELF in the guest's buffer and report how
// much memory it needs.
func (p *IOP) loadcoreProbeModule() {
	image, info := p.arg(0), p.arg(1)

	x, err := p.readGuestIRX(image)
	if err != nil {
		p.ps2.note("IOP: loadcore#22: %s is not a loadable module: %v", p.Sym(image), err)
		p.setRet(iopModBadImage)
		return
	}
	p.Write32(info+iopModInfoSize, x.MemSz)
	p.setRet(iopModRelocatable)
}

// loadcoreLinkModule is loadcore #23: place the module on the memory MODLOAD allocated for
// it, and wire it into everything resident.
func (p *IOP) loadcoreLinkModule() {
	image, info := p.arg(0), p.arg(1)
	base := p.Read32(info + iopModInfoInit)

	x, err := p.readGuestIRX(image)
	if err != nil {
		p.ps2.note("IOP: loadcore#23: %s: %v", p.Sym(image), err)
		p.setRet(iopModBadImage)
		return
	}
	mod, err := p.placeAndLink(p.guestModuleName(image, x), x, base)
	if err != nil {
		p.ps2.note("IOP: loadcore#23: %v", err)
		p.setRet(iopModBadImage)
		return
	}

	// Fill the registration record MODLOAD placed just below the image. MODLOAD starts the
	// module by reading its entry and $gp out of this record — leave them and it jumps to
	// whatever the freshly-allocated memory happened to hold, which is exactly the wild jump
	// this was found by.
	rec := base - iopModRecSize
	p.Write32(rec+iopModRecEntry, mod.Base+x.Entry)
	p.Write32(rec+iopModRecGP, mod.Base+x.GP)
	p.Write32(rec+iopModRecArg0, mod.Base)
	p.Write32(rec+iopModRecArg1, mod.Size)

	p.Write32(info+iopModInfoText, mod.Base)
	p.ps2.note("IOP: MODLOAD loaded %s at 0x%08X (%d KiB) via loadcore", mod.Name, mod.Base, mod.Size/1024)
	p.setRet(0)
}

// loadcoreLinkCheck is loadcore #8: MODLOAD's gate on the load. placeAndLink resolves every
// import as it goes and errors if it cannot, so a module that reaches here has already
// linked; this only has to answer "not negative".
func (p *IOP) loadcoreLinkCheck() { p.setRet(0) }

// loadcoreFlushIcache is loadcore #4: flush the instruction cache after new code has been
// written. There is no cache in this model — an instruction fetch reads memory as it stands
// — so there is nothing to flush.
func (p *IOP) loadcoreFlushIcache() { p.setRet(0) }

// loadcoreRegisterModule is loadcore #16: enter the module in the registry. Ours is the Go
// slice placeAndLink already appended to, and MODLOAD returns the handle it computed itself,
// so this only has to succeed.
func (p *IOP) loadcoreRegisterModule() { p.setRet(0) }

// readGuestIRX reads and parses the ELF image the guest has placed in IOP memory at addr.
//
// The length is the ELF's own, and it has to be computed rather than guessed: the caller
// does not pass it, and the section-header table is *not* at the end of the file — SIO2MAN's
// relocation sections live past it. So the file ends wherever the furthest-reaching part of
// it does, which is the largest `offset + size` over every section (except .bss, which takes
// no file space), every program segment, and the two header tables themselves. Read that many
// bytes out of IOP RAM and ReadIRX has exactly the file LOADFILE fetched off the disc.
func (p *IOP) readGuestIRX(addr uint32) (*IRX, error) {
	if m := p.Read32(addr); m != 0x464C457F { // "\x7FELF", little-endian
		return nil, fmt.Errorf("no ELF magic at 0x%08X (found 0x%08X)", addr, m)
	}
	rd16 := func(o uint32) uint32 { return uint32(p.Read(addr+o)) | uint32(p.Read(addr+o+1))<<8 }

	phoff, phentsize, phnum := p.Read32(addr+0x1C), rd16(0x2A), rd16(0x2C)
	shoff, shentsize, shnum := p.Read32(addr+0x20), rd16(0x2E), rd16(0x30)

	size := phoff + phentsize*phnum
	if e := shoff + shentsize*shnum; e > size {
		size = e
	}
	for i := uint32(0); i < shnum; i++ {
		sh := addr + shoff + i*shentsize
		if p.Read32(sh+0x04) == 8 { // SHT_NOBITS — .bss, no bytes in the file
			continue
		}
		if e := p.Read32(sh+0x10) + p.Read32(sh+0x14); e > size {
			size = e
		}
	}
	for i := uint32(0); i < phnum; i++ {
		ph := addr + phoff + i*phentsize
		if e := p.Read32(ph+0x04) + p.Read32(ph+0x10); e > size {
			size = e
		}
	}

	if size == 0 || addr+size > iopRAMSizeBytes {
		return nil, fmt.Errorf("the ELF header at 0x%08X gives an implausible file size of %d bytes", addr, size)
	}
	raw := make([]byte, size)
	copy(raw, p.ram[addr:addr+size])
	return ReadIRX(raw)
}

// guestModuleName is the module's own name, out of its .iopmod record, so a module loaded at
// runtime is named in the trace the same way one loaded at boot is.
func (p *IOP) guestModuleName(addr uint32, x *IRX) string {
	if x.Name != "" {
		return x.Name
	}
	return fmt.Sprintf("module@0x%08X", addr)
}

// exportAddr is the address of function id in a resident library.
func (p *IOP) exportAddr(library string, id uint16) uint32 {
	exp := p.exports[library]
	owner := p.owner[library]
	for _, m := range p.modules {
		if m.Name == owner {
			return m.Base + exp.Entries[id]
		}
	}
	return 0
}

// Start runs a module's entry point, as loadcore does: with (argc, argv) and a result
// pointer, on the loader's own stack.
//
// A module's entry is not a thread. It runs to completion — registering the module's
// libraries, creating whatever threads it wants, starting them — and returns a status
// saying whether it wishes to stay resident. So the loader simply calls it, exactly as
// the machine calls any guest routine, and the module is loaded when it comes back.
func (p *IOP) Start(mod *IOPModule, args ...uint32) (uint32, error) {
	entry := mod.Base + mod.IRX.Entry
	gp := mod.Base + mod.IRX.GP

	// The arguments the module's _start expects: argc, argv, and a place to put a
	// module-info pointer. A module started with no arguments gets a zero argc, which
	// every module on this disc accepts.
	argv := uint32(0)
	if len(args) > 0 {
		argv = p.alloc(uint32(4 * len(args)))
		for i, a := range args {
			p.Write32(argv+uint32(4*i), a)
		}
	}

	p.CPU.SetReg(4, uint32(len(args))) // $a0 = argc
	p.CPU.SetReg(5, argv)              // $a1 = argv
	p.CPU.SetReg(6, 0)                 // $a2
	p.CPU.SetReg(7, 0)                 // $a3
	p.CPU.SetReg(28, gp)               // $gp — the module's own, and it is not saved
	//                                    across calls: every entry sets its own.

	res, err := p.callGuest(entry)
	if err != nil {
		return 0, fmt.Errorf("starting %s: %w", mod.Name, err)
	}
	return res, nil
}

// iopReturnSentinel is the return address a called-from-Go routine is given. It is not
// memory; when the PC reaches it, the routine has finished.
const iopReturnSentinel = 0x0FFF0000

// iopCallStack is where a called-from-Go routine's stack starts. It sits at the top of
// IOP memory, well clear of anything the allocator will ever hand out.
const iopCallStack = iopRAMSizeBytes - 0x400

// iopCallBudget is how many instructions a routine called from Go may run before the
// machine decides it is never coming back, and the spin detector's window.
const (
	iopCallBudget   = 200_000_000
	iopSpinWindow   = 0x100000
	iopSpinDistinct = 8
)

// callGuest runs an IOP routine to completion and returns its $v0.
//
// Module entry points are called this way. It is a nested interpreter loop rather than
// a scheduled thread, which is right for the thing being modelled: loadcore really does
// call a module's entry and really does wait for it, and nothing else on the IOP runs
// while it does.
func (p *IOP) callGuest(entry uint32) (uint32, error) {
	return p.callGuestOn(entry, iopCallStack)
}

// callGuestOn is callGuest on a nominated stack. An interrupt handler needs one of its
// own: it runs *inside* another routine's call, and the stack that routine is using is
// the one thing it must not touch.
func (p *IOP) callGuestOn(entry, stack uint32) (uint32, error) {
	p.callDepth++
	defer func() { p.callDepth-- }()

	saved := p.CPU.PC
	savedSP := p.CPU.Reg(29)
	savedRA := p.CPU.Reg(31)

	p.CPU.SetReg(29, stack)
	p.CPU.SetReg(31, iopReturnSentinel)
	p.CPU.SetPC(entry)

	// The spin detector, the same idea as the EE's: a window of instructions that visits
	// only a handful of addresses is a wait for something that is never going to happen,
	// and saying *which* addresses is the entire difference between a diagnosis and a
	// shrug. An IOP module that polls a register the other processor should have written
	// looks exactly like this, and during bring-up it is the single most common way to
	// be stuck.
	seen := map[uint32]int{}
	acc := 0

	for i := 0; i < iopCallBudget; i++ {
		if p.CPU.PC == iopReturnSentinel {
			v := p.CPU.Reg(2)
			p.CPU.SetPC(saved)
			p.CPU.SetReg(29, savedSP)
			p.CPU.SetReg(31, savedRA)
			return v, nil
		}
		if p.Halted || p.CPU.Halted {
			return 0, fmt.Errorf("the IOP stopped at %s: %s", p.Sym(p.CPU.PC), p.reason())
		}

		seen[p.CPU.PC]++
		acc++
		if acc >= iopSpinWindow {
			if len(seen) <= iopSpinDistinct {
				return 0, fmt.Errorf("%s is spinning on %d addresses around %s — waiting for something that never happens",
					p.Sym(entry), len(seen), p.Sym(p.CPU.PC))
			}
			seen = map[uint32]int{}
			acc = 0
		}

		p.tick()
		p.CPU.Step()
	}
	return 0, fmt.Errorf("the routine at %s ran %d instructions without returning, and is at %s",
		p.Sym(entry), iopCallBudget, p.Sym(p.CPU.PC))
}

func (p *IOP) reason() string {
	if p.HaltReason != "" {
		return p.HaltReason
	}
	return p.CPU.HaltReason
}
