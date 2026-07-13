package ps2

// iopkernel.go stands in for the part of the IOP that lives in a ROM.
//
// Nine libraries, sixty-six functions. irxinfo counted them: they are what every
// module on the disc imports and no module on the disc provides.
//
//	loadcore   the module registry — and ours by definition, since we are the loader
//	intrman    the interrupt controller
//	sysmem     the allocator
//	heaplib    a heap on top of it
//	sysclib    the C library: memcpy, strlen, sprintf
//	stdio      printf
//	dmacman    the DMA controller
//	vblank     the vertical blank
//	secrman    the memory card's authentication
//
// There is a difficulty here that the EE's kernel did not have, and it shapes this
// whole file. A module imports a function by *number*: `thbase #4`, `sysclib #29`.
// The number indexes the exporting library's table, and the name never appears
// anywhere. For the fourteen libraries the disc provides, that is fine — we link the
// number straight to the code and never need to know what it is called. For these
// nine there is no code to link to, and Sony's modules are stripped, so the name has
// to be *earned*: the function is called, the call is logged with its arguments and
// the routine that made it, and the answer is worked out from what the caller does
// with it. IOPCensus is that log, and this file is the record of what has been won
// from it.
//
// So the three tiers, the same as the PSP's kernel:
//
//	modelled     we know what it is and it does the thing
//	stubbed      we know what it is, it does nothing, and the comment says why that
//	             is safe
//	unmodelled   we do not know what it is yet. It returns zero and is counted, and
//	             the count is the work list.
//
// Nothing here guesses silently. A function whose identity has not been established
// is in the third tier even when its number looks familiar, because a kernel that
// lies confidently is worse than one that admits it does not know.

import (
	"fmt"

	"retroreverse.com/tools/cpu/mips"
)

// iopCall is one entry in the syscall table: an import stub, or a trampoline.
type iopCall struct {
	name string
	fn   func(*IOP) // nil: unmodelled — counted and answered with zero
}

// iopLibrary is a library we model, and the functions of it we have identified.
type iopLibrary struct {
	funcs map[uint16]iopFunc
}

// iopFunc is one function of a modelled library.
type iopFunc struct {
	name string
	fn   func(*IOP)
}

// goLibraries are the libraries the IOP kernel HLE provides. A library named here is
// one the loader will bind stubs to; a library absent from it, and absent from the
// disc, is a load error rather than a silent zero.
var goLibraries = map[string]*iopLibrary{}

// lib registers a Go library and its known functions.
func lib(name string, funcs map[uint16]iopFunc) {
	goLibraries[name] = &iopLibrary{funcs: funcs}
}

// unknown declares a function we must serve but have not identified. It is the
// third tier: called, counted, answered with zero, and named in the census by the
// only thing we know about it — its library and its number.
func unknown() iopFunc { return iopFunc{} }

func init() {
	// loadcore. The module registry: modules call it to publish the libraries they
	// export and to find the ones they import. We are the loader, so we have already
	// done both — LoadIRX registers a module's exports the moment it is placed, from
	// the export tables in its own image, which is the same information by a shorter
	// road.
	lib("loadcore", map[uint16]iopFunc{
		3:  unknown(),
		4:  unknown(),
		5:  unknown(),
		6:  {"RegisterLibraryEntries", (*IOP).loadcoreRegisterLibrary},
		8:  unknown(),
		9:  unknown(),
		10: unknown(),
		12: unknown(),
		16: unknown(),
		17: unknown(),
		20: unknown(),
		21: unknown(),
		22: unknown(),
		23: unknown(),
		24: unknown(),
	})

	// intrman, sysmem, heaplib and sysclib are in iopintr.go: they are the four the
	// modules cannot take a step without, and they carry the arguments for what each of
	// their functions was found to be.

	// stdio. One function, imported by every module on the disc, in a library called
	// "stdio". It is printf, and the model proves it rather than assuming it: it reads
	// the first argument as a format string and prints it, and a first argument that
	// is not a plausible format string would show up immediately as garbage.
	lib("stdio", map[uint16]iopFunc{
		4: {"printf", (*IOP).stdioPrintf},
	})

	lib("dmacman", map[uint16]iopFunc{
		28: unknown(), 32: unknown(), 33: unknown(), 34: unknown(), 35: unknown(),
	})

	lib("vblank", map[uint16]iopFunc{
		8: unknown(), 9: unknown(),
	})

	lib("secrman", map[uint16]iopFunc{
		4: unknown(), 5: unknown(), 6: unknown(),
	})
}

// registerLibraries is called when the IOP is made. The table is global and constant;
// the per-machine part is the syscall table the loader fills in as it links.
func (p *IOP) registerLibraries() {
	p.calls = append(p.calls, &iopCall{name: "<invalid>"}) // code 0 is never handed out
	p.bound = map[iopBinding]uint32{}
}

// iopBinding names one function of one library.
type iopBinding struct {
	library string
	id      uint16
}

// hasGoLibrary reports whether the kernel HLE serves a library.
func (p *IOP) hasGoLibrary(name string) bool {
	_, ok := goLibraries[name]
	return ok
}

// bindCall returns the syscall code that serves (library, id), allocating one on
// first use. A library we model but a function of it we have not identified still
// gets a code — it is the unmodelled tier, and being called is how it gets found.
func (p *IOP) bindCall(library string, id uint16) uint32 {
	key := iopBinding{library, id}
	if code, ok := p.bound[key]; ok {
		return code
	}
	l := goLibraries[library]
	f, known := l.funcs[id]

	name := fmt.Sprintf("%s#%d", library, id)
	if known && f.name != "" {
		name = fmt.Sprintf("%s.%s", library, f.name)
	}
	code := p.addCall(name, f.fn)
	p.bound[key] = code
	return code
}

// addCall appends a syscall handler and returns its code.
func (p *IOP) addCall(name string, fn func(*IOP)) uint32 {
	p.calls = append(p.calls, &iopCall{name: name, fn: fn})
	return uint32(len(p.calls) - 1)
}

// handleSyscall is the CPU's Syscall hook: the IOP has reached a patched stub or a
// trampoline, and this is the Go function behind it.
//
// The stub is `jr $ra` with the `syscall` in its delay slot, so the branch has already
// been taken by the time we are here: setting $v0 and returning is the whole of the
// call. The PC is not touched.
func (p *IOP) handleSyscall(c *mips.CPU) bool {
	code := (p.Read32(c.CurPC()) >> 6) & 0xFFFFF
	if code == 0 || int(code) >= len(p.calls) {
		return false // not ours: let the core take the exception
	}
	call := p.calls[code]

	if call.fn == nil {
		// The unmodelled tier. Everything we know about it goes into the log — the
		// arguments, and the routine that made the call, which on a module with symbols
		// is a name — because that is what the identification will be made from.
		p.unmodelledCalls[call.name]++
		if p.unmodelledCalls[call.name] == 1 {
			p.ps2.note("IOP: %s(0x%X, 0x%X, 0x%X, 0x%X) from %s — unmodelled",
				call.name, p.arg(0), p.arg(1), p.arg(2), p.arg(3), p.Sym(p.CPU.Reg(31)))
		}
		p.setRet(0)
		return true
	}

	call.fn(p)
	return true
}

// arg reads the i'th argument of a call. The first four are in $a0..$a3, the rest on
// the stack above the caller's 16 bytes of argument save area.
func (p *IOP) arg(i int) uint32 {
	if i < 4 {
		return p.CPU.Reg(uint32(4 + i))
	}
	return p.Read32(p.CPU.Reg(29) + uint32(4*i))
}

func (p *IOP) setRet(v uint32) { p.CPU.SetReg(2, v) }

// --- the modelled tier -------------------------------------------------------------

// loadcoreRegisterLibrary is RegisterLibraryEntries: a module publishing its exports.
//
// It succeeds and does nothing. The loader read the module's export tables out of its
// image when it placed it (LoadIRX) and has already wired every importer to them, so
// by the time the module gets round to announcing itself, the announcement is old
// news. What matters is the return value: a module whose registration *fails* takes
// itself back out of memory, so this has to say yes.
func (p *IOP) loadcoreRegisterLibrary() { p.setRet(0) }

// stdioPrintf is the IOP's printf. Every module on the disc imports exactly this one
// function of stdio, and it is how the IOP kernel narrates itself — the counterpart of
// the DECI2 output the EE's own kernel gives us, and the first place an IOP that is
// going wrong will say so.
func (p *IOP) stdioPrintf() {
	s := p.formatArgs(p.CString(p.arg(0)), 1)
	p.ps2.iopPrint(s)
	p.setRet(uint32(len(s)))
}

// formatArgs renders a printf format string with the IOP's arguments.
//
// It handles the conversions the kernel modules actually use and no more: a
// conversion it does not know is copied through verbatim rather than guessed at, so a
// message that comes out wrong is visibly wrong instead of quietly plausible.
func (p *IOP) formatArgs(format string, argi int) string {
	var out []byte
	next := func() uint32 {
		v := p.arg(argi)
		argi++
		return v
	}

	for i := 0; i < len(format); i++ {
		if format[i] != '%' || i+1 >= len(format) {
			out = append(out, format[i])
			continue
		}
		// Skip the flags, width and precision: the value is what matters.
		j := i + 1
		for j < len(format) && (format[j] == '-' || format[j] == '+' || format[j] == ' ' ||
			format[j] == '#' || format[j] == '0' || format[j] == '.' ||
			(format[j] >= '1' && format[j] <= '9') || format[j] == 'l' || format[j] == 'h') {
			j++
		}
		if j >= len(format) {
			out = append(out, format[i:]...)
			break
		}
		verb := format[j]
		spec := format[i : j+1]

		switch verb {
		case '%':
			out = append(out, '%')
		case 'd', 'i':
			out = append(out, fmt.Sprintf("%d", int32(next()))...)
		case 'u':
			out = append(out, fmt.Sprintf("%d", next())...)
		case 'x', 'X', 'o':
			out = append(out, fmt.Sprintf(string([]byte{'%', verb}), next())...)
		case 'c':
			out = append(out, byte(next()))
		case 'p':
			out = append(out, fmt.Sprintf("0x%08X", next())...)
		case 's':
			out = append(out, p.CString(next())...)
		default:
			out = append(out, spec...) // unknown: pass it through, and be seen to
		}
		i = j
	}
	return string(out)
}
