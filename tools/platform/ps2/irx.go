package ps2

// irx.go reads an IOP module.
//
// An IRX is the IOP's executable: a little-endian 32-bit MIPS ELF, relocatable,
// linked at zero and loaded wherever there is room. Three things distinguish it
// from the EE's boot ELF (elf.go), and all three are what make an IOP possible to
// build at all.
//
// The .iopmod section. A small header naming the module, its version, its entry
// point and its $gp — the values the loader needs before it can run a line of it.
//
// The import tables. A module does not call the kernel through a syscall; it calls
// through a table of stubs embedded in its own .text, one per imported function:
//
//	+0x00  0x41E00000        the magic that says "import table"
//	+0x04  0                 filled in at link time
//	+0x08  version
//	+0x0C  name[8]           the library, e.g. "thbase  "
//	+0x14  stubs             two words each: `jr $ra` then `li $v0, id`
//	                         terminated by a pair of zeros
//
// The export tables are the same shape with magic 0x41C00000, and carry a flat
// array of addresses. Resolution is by *index*: an import of thbase function 4 is
// the fifth word of thbase's export table. There are no names on the wire at all —
// which is why the census in iopkernel.go is written in terms of numbers, and why
// each number has to be earned from the code that calls it.
//
// This is precisely the shape the PSP's PRX modules use, so the trick that HLEs
// those applies unchanged: a library we model in Go has each of its stubs
// overwritten with `jr $ra; syscall n`, and the R3000A's syscall exception hands us
// the call with its arguments already in place. A library we have the real module
// for is linked to it instead, and the game's own code runs.
//
// Relocation is the easy part. Every relocation in every module on this disc — all
// nineteen of them, 19,492 relocations — is one of R_MIPS_32, R_MIPS_26, R_MIPS_HI16
// and R_MIPS_LO16, and every one is against symbol 0. There is no global offset
// table, no GP-relative fixup, nothing to resolve between modules. Relocating an IRX
// is adding the load address to it.

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// The magic words that head the two kinds of library table.
const (
	irxImportMagic = 0x41E00000
	irxExportMagic = 0x41C00000
)

// irxStubJR is the `jr $ra` that heads every import stub. It is the marker that
// says the table's entries are really stubs and not something that merely happened
// to follow a magic word.
const irxStubJR = 0x03E00008

// irxExportHooks is how many entries of an export table come before the library's
// functions do. They are the module's own hooks — its entry point is the first —
// and the functions the rest of the world imports begin at index 4.
const irxExportHooks = 4

// IRXImport is one library a module imports, and the functions it wants from it.
type IRXImport struct {
	Library string
	Version uint16
	Addr    uint32   // where the table sits in the module image
	IDs     []uint16 // the function index of each stub, in order
	Stubs   []uint32 // where each stub sits in the module image
}

// IRXExport is one library a module provides.
type IRXExport struct {
	Library string
	Version uint16
	Addr    uint32
	Entries []uint32 // entry i is the address of function i, in the module image
}

// IRXReloc is one relocation.
type IRXReloc struct {
	Offset uint32 // into the module image
	Type   uint8
}

// The relocation types the disc's modules actually use.
const (
	rMIPS32   = 2
	rMIPS26   = 4
	rMIPSHI16 = 5
	rMIPSLO16 = 6
)

// IRX is a loaded-but-not-yet-placed IOP module.
type IRX struct {
	Name    string
	Version uint16

	Entry uint32 // all four are offsets into Image
	GP    uint32
	Image []byte // text and data, with MemSz-len(Image) bytes of BSS to follow
	MemSz uint32

	Imports []IRXImport
	Exports []IRXExport
	Relocs  []IRXReloc
	Symbols []Symbol // present only in the game's own modules; Sony's are stripped
}

// ReadIRX parses an IOP module.
func ReadIRX(raw []byte) (*IRX, error) {
	f, err := elf.NewFile(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("ps2: parsing IRX: %w", err)
	}
	if f.Class != elf.ELFCLASS32 || f.Data != elf.ELFDATA2LSB || f.Machine != elf.EM_MIPS {
		return nil, fmt.Errorf("ps2: not a little-endian 32-bit MIPS ELF")
	}

	m := &IRX{}

	// The loadable image. There is exactly one PT_LOAD, at vaddr 0.
	var loaded bool
	for _, p := range f.Progs {
		if p.Type != elf.PT_LOAD {
			continue
		}
		if p.Vaddr != 0 {
			return nil, fmt.Errorf("ps2: an IRX segment is linked at 0x%X, not 0", p.Vaddr)
		}
		data := make([]byte, p.Filesz)
		if _, err := p.ReadAt(data, 0); err != nil {
			return nil, fmt.Errorf("ps2: reading the IRX image: %w", err)
		}
		m.Image, m.MemSz, loaded = data, uint32(p.Memsz), true
	}
	if !loaded {
		return nil, fmt.Errorf("ps2: the IRX has no loadable segment")
	}

	if err := m.readModuleInfo(f); err != nil {
		return nil, err
	}
	if err := m.readRelocs(f); err != nil {
		return nil, err
	}
	m.readSymbols(f)
	m.scanLibraries()
	return m, nil
}

// readModuleInfo reads the .iopmod header: who the module is and how to start it.
//
//	+0x00  moduleinfo   a pointer the kernel keeps; 0xFFFFFFFF when there is none
//	+0x04  entry
//	+0x08  gp
//	+0x0C  text size
//	+0x10  data size
//	+0x14  bss size
//	+0x18  version      u16
//	+0x1A  name         NUL-terminated
func (m *IRX) readModuleInfo(f *elf.File) error {
	s := f.Section(".iopmod")
	if s == nil {
		return fmt.Errorf("ps2: the IRX has no .iopmod section")
	}
	d, err := s.Data()
	if err != nil {
		return err
	}
	if len(d) < 0x1A {
		return fmt.Errorf("ps2: the .iopmod section is %d bytes, too short to be one", len(d))
	}
	m.Entry = binary.LittleEndian.Uint32(d[0x04:])
	m.GP = binary.LittleEndian.Uint32(d[0x08:])
	m.Version = binary.LittleEndian.Uint16(d[0x18:])
	if len(d) > 0x1A {
		m.Name = cstring(d[0x1A:])
	}
	if m.Name == "" {
		// A few modules (OVERLORD among them) leave the name blank. The ELF is still
		// perfectly loadable; only the log would be poorer for it, and the caller knows
		// the file it handed us.
		m.Name = "?"
	}
	return nil
}

// readRelocs collects every relocation, from every SHT_REL section.
//
// The offsets are relative to the module image as a whole, not to the section the
// relocation belongs to, so they can be applied without consulting sh_info.
func (m *IRX) readRelocs(f *elf.File) error {
	for _, s := range f.Sections {
		if s.Type != elf.SHT_REL {
			continue
		}
		d, err := s.Data()
		if err != nil {
			return err
		}
		for o := 0; o+8 <= len(d); o += 8 {
			off := binary.LittleEndian.Uint32(d[o:])
			info := binary.LittleEndian.Uint32(d[o+4:])
			m.Relocs = append(m.Relocs, IRXReloc{Offset: off, Type: uint8(info)})
		}
	}
	return nil
}

// readSymbols keeps the symbol table when there is one.
//
// Sony's kernel modules are stripped. The game's are not: OVERLORD.IRX ships 132
// named functions — ISOThread, RunDGOStateMachine, FS_LoadMusic, CopyDataToEE — so
// every instrument that prints an IOP address can name it, exactly as the EE's can.
func (m *IRX) readSymbols(f *elf.File) {
	syms, err := f.Symbols()
	if err != nil {
		return
	}
	for _, s := range syms {
		if s.Name == "" {
			continue
		}
		m.Symbols = append(m.Symbols, Symbol{
			Name: s.Name,
			Addr: uint32(s.Value),
			Size: uint32(s.Size),
			Func: elf.ST_TYPE(s.Info) == elf.STT_FUNC,
		})
	}
	sort.Slice(m.Symbols, func(i, j int) bool { return m.Symbols[i].Addr < m.Symbols[j].Addr })
}

// scanLibraries finds the import and export tables inside the module image.
//
// They are not in a section of their own — they are compiled into .text, wherever
// the linker put them — so they are found by their magic word and then confirmed:
// the reserved word must be zero, the name must be a plausible library name, and an
// import table's entries must actually be `jr $ra` stubs. A magic word that occurs
// by chance inside real code fails all three.
func (m *IRX) scanLibraries() {
	for o := uint32(0); o+20 <= uint32(len(m.Image)); o += 4 {
		magic := m.word(o)
		if magic != irxImportMagic && magic != irxExportMagic {
			continue
		}
		if m.word(o+4) != 0 {
			continue
		}
		name := libraryName(m.Image[o+12 : o+20])
		if name == "" {
			continue
		}
		version := uint16(m.word(o + 8))

		if magic == irxImportMagic {
			imp := IRXImport{Library: name, Version: version, Addr: o}
			for p := o + 20; p+8 <= uint32(len(m.Image)); p += 8 {
				lo, hi := m.word(p), m.word(p+4)
				if lo == 0 && hi == 0 {
					break // the pair of zeros that ends the table
				}
				if lo != irxStubJR {
					break
				}
				imp.IDs = append(imp.IDs, uint16(hi))
				imp.Stubs = append(imp.Stubs, p)
			}
			if len(imp.IDs) > 0 {
				m.Imports = append(m.Imports, imp)
			}
			continue
		}

		// An export table is a flat array of addresses ended by a zero word — but only
		// from entry 4 onwards. The first four are the module's standard hooks (its entry
		// point among them) and any of them may legitimately be zero: SIFMAN's entry point
		// *is* zero, and terminating on it would hide the library behind it. Reading them
		// as data and only honouring the terminator past them gives every table on this
		// disc a length that covers exactly the indices the disc imports from it.
		exp := IRXExport{Library: name, Version: version, Addr: o}
		for p := o + 20; p+4 <= uint32(len(m.Image)); p += 4 {
			w := m.word(p)
			if w == 0 && len(exp.Entries) >= irxExportHooks {
				break
			}
			exp.Entries = append(exp.Entries, w)
		}
		if len(exp.Entries) >= irxExportHooks {
			m.Exports = append(m.Exports, exp)
		}
	}
}

func (m *IRX) word(o uint32) uint32 {
	return binary.LittleEndian.Uint32(m.Image[o:])
}

// libraryName reads the 8-byte library name out of a table header, and reports ""
// for anything that is not one. Names are space-padded, printable and short.
func libraryName(b []byte) string {
	n := 0
	for n < len(b) && b[n] != 0 {
		n++
	}
	s := strings.TrimRight(string(b[:n]), " ")
	if s == "" {
		return ""
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > 'z' {
			return ""
		}
	}
	return s
}

func cstring(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return string(b)
}

// Relocate returns the module's image placed at base, with every relocation applied.
//
// The IOP has no MMU worth the name: a module is loaded at whatever address the
// allocator gave it and its code is patched to suit. All four relocation types come
// down to adding the base, and the only one with any subtlety is the HI16/LO16 pair,
// where the two halves of a 32-bit address are split across two instructions and the
// low half is *signed* — so a low half that will be sign-extended to negative has to
// be compensated for by adding one to the high half. Getting that wrong moves a
// pointer by 64 KiB, which is exactly far enough to look like a plausible address and
// crash somewhere unrelated.
func (m *IRX) Relocate(base uint32) ([]byte, error) {
	img := make([]byte, m.MemSz)
	copy(img, m.Image)

	get := func(o uint32) uint32 { return binary.LittleEndian.Uint32(img[o:]) }
	put := func(o, v uint32) { binary.LittleEndian.PutUint32(img[o:], v) }

	// A HI16 has no addend of its own that means anything until its LO16 is known, so
	// the pair is resolved together. They come in order, and a run of HI16s may share
	// one LO16.
	var pendingHI []uint32

	for _, r := range m.Relocs {
		if r.Offset+4 > m.MemSz {
			return nil, fmt.Errorf("ps2: a relocation in %s points outside the module (0x%X)", m.Name, r.Offset)
		}
		switch r.Type {
		case rMIPS32:
			put(r.Offset, get(r.Offset)+base)

		case rMIPS26:
			// A jump's target is a word address within the current 256 MiB region.
			insn := get(r.Offset)
			target := (insn&0x03FFFFFF)<<2 + base
			put(r.Offset, insn&^0x03FFFFFF|(target>>2)&0x03FFFFFF)

		case rMIPSHI16:
			pendingHI = append(pendingHI, r.Offset)

		case rMIPSLO16:
			lo := get(r.Offset)
			// The addend is the two halves put back together, with the low half signed.
			addend := int32(int16(lo & 0xFFFF))
			for _, hiOff := range pendingHI {
				hi := get(hiOff)
				full := uint32(int32(hi&0xFFFF)<<16+addend) + base
				put(hiOff, hi&^0xFFFF|(full+0x8000)>>16&0xFFFF)
			}
			pendingHI = pendingHI[:0]

			full := uint32(addend) + base
			put(r.Offset, lo&^0xFFFF|full&0xFFFF)

		default:
			return nil, fmt.Errorf("ps2: %s uses relocation type %d, which nothing on this disc does", m.Name, r.Type)
		}
	}
	if len(pendingHI) > 0 {
		return nil, fmt.Errorf("ps2: %s ends with %d unpaired HI16 relocations", m.Name, len(pendingHI))
	}
	return img, nil
}
