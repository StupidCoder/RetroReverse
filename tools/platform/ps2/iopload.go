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
func (p *IOP) LoadIRX(name string, raw []byte) (*IOPModule, error) {
	x, err := ReadIRX(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}

	base := p.alloc(x.MemSz)
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
