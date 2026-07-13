package ps2

// elf.go loads a PS2 executable.
//
// There is nothing exotic about it: a PS2 boot ELF is a plain statically-linked
// ELF32, little-endian, EM_MIPS, with a single loadable segment. Unlike the PSP's
// PRX there are no import stubs to patch, no relocations to apply and no module
// info to parse — the file says where it wants to be and what its entry point is,
// and that is the whole of the loading protocol.
//
// What makes it worth a file of its own is the symbol table. Jak and Daxter's boot
// ELF ships .symtab and .strtab intact — 1,524 symbols, 876 of them functions, with
// C++ mangling and all. So the loader keeps them, and every instrument downstream
// (the disassembler, the tracer, the oracle's -logpc and backtrace) can name an
// address instead of printing a number.

import (
	"bytes"
	"debug/elf"
	"fmt"
	"sort"
)

// Segment is one loadable piece of the executable.
type Segment struct {
	VAddr uint32
	Data  []byte // FileSz bytes; the remainder up to MemSz is zero-filled BSS
	MemSz uint32
}

// Symbol names an address in the executable.
type Symbol struct {
	Name string
	Addr uint32
	Size uint32
	Func bool
}

// Executable is a loaded PS2 ELF.
type Executable struct {
	Entry    uint32
	Segments []Segment
	Symbols  []Symbol // sorted by address

	byAddr []Symbol // the function symbols only, sorted, for Lookup
}

// LoadELF parses a PS2 boot ELF.
func LoadELF(raw []byte) (*Executable, error) {
	f, err := elf.NewFile(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("ps2: parsing ELF: %w", err)
	}
	if f.Class != elf.ELFCLASS32 || f.Data != elf.ELFDATA2LSB || f.Machine != elf.EM_MIPS {
		return nil, fmt.Errorf("ps2: not a little-endian 32-bit MIPS ELF (class %v, data %v, machine %v)",
			f.Class, f.Data, f.Machine)
	}

	e := &Executable{Entry: uint32(f.Entry)}
	for _, p := range f.Progs {
		if p.Type != elf.PT_LOAD || p.Memsz == 0 {
			continue
		}
		data := make([]byte, p.Filesz)
		if _, err := p.ReadAt(data, 0); err != nil && p.Filesz > 0 {
			return nil, fmt.Errorf("ps2: reading segment at 0x%08X: %w", p.Vaddr, err)
		}
		e.Segments = append(e.Segments, Segment{
			VAddr: uint32(p.Vaddr),
			Data:  data,
			MemSz: uint32(p.Memsz),
		})
	}
	if len(e.Segments) == 0 {
		return nil, fmt.Errorf("ps2: ELF has no loadable segments")
	}

	// The symbol table is optional — a stripped executable simply has none — so a
	// failure to read it is not a failure to load.
	if syms, err := f.Symbols(); err == nil {
		for _, s := range syms {
			if s.Name == "" || s.Value == 0 {
				continue
			}
			e.Symbols = append(e.Symbols, Symbol{
				Name: s.Name,
				Addr: uint32(s.Value),
				Size: uint32(s.Size),
				Func: elf.ST_TYPE(s.Info) == elf.STT_FUNC,
			})
		}
		sort.Slice(e.Symbols, func(i, j int) bool { return e.Symbols[i].Addr < e.Symbols[j].Addr })
		for _, s := range e.Symbols {
			if s.Func {
				e.byAddr = append(e.byAddr, s)
			}
		}
	}
	return e, nil
}

// Lookup names the function containing addr, and the offset into it. It reports
// ok == false when no function covers the address.
func (e *Executable) Lookup(addr uint32) (name string, off uint32, ok bool) {
	i := sort.Search(len(e.byAddr), func(i int) bool { return e.byAddr[i].Addr > addr }) - 1
	if i < 0 {
		return "", 0, false
	}
	s := e.byAddr[i]
	// A symbol with a size covers exactly that range; one without is taken to run up
	// to the next symbol, which is the best a stripped-size entry allows.
	end := s.Addr + s.Size
	if s.Size == 0 {
		if i+1 < len(e.byAddr) {
			end = e.byAddr[i+1].Addr
		} else {
			end = s.Addr + 4
		}
	}
	if addr >= end {
		return "", 0, false
	}
	return s.Name, addr - s.Addr, true
}

// Describe renders a summary of the executable, for the static inspector.
func (e *Executable) Describe() string {
	s := fmt.Sprintf("entry:    0x%08X\n", e.Entry)
	for i, seg := range e.Segments {
		bss := seg.MemSz - uint32(len(seg.Data))
		s += fmt.Sprintf("segment %d: 0x%08X..0x%08X  %d bytes in file, %d bytes bss\n",
			i, seg.VAddr, seg.VAddr+seg.MemSz-1, len(seg.Data), bss)
	}
	funcs := 0
	for _, sym := range e.Symbols {
		if sym.Func {
			funcs++
		}
	}
	s += fmt.Sprintf("symbols:  %d (%d functions)\n", len(e.Symbols), funcs)
	return s
}

// Flat returns the executable's image as one contiguous buffer covering every
// loadable segment, with the base address it starts at. BSS is zero-filled. This is
// the form the disassembler and the code tracer want.
func (e *Executable) Flat() (base uint32, mem []byte) {
	lo, hi := ^uint32(0), uint32(0)
	for _, s := range e.Segments {
		if s.VAddr < lo {
			lo = s.VAddr
		}
		if s.VAddr+s.MemSz > hi {
			hi = s.VAddr + s.MemSz
		}
	}
	mem = make([]byte, hi-lo)
	for _, s := range e.Segments {
		copy(mem[s.VAddr-lo:], s.Data)
	}
	return lo, mem
}
