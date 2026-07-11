package psp

// elf.go parses a decrypted PSP module: an ELF32 little-endian MIPS image, either a
// static executable (ET_EXEC) or a relocatable PRX (ET_PRX = 0xFFA0). Beyond the
// standard ELF program/section headers it reads the PSP-specific .rodata.sceModuleInfo
// — the module name, the gp value, and the import ("stub") tables that name every
// kernel function the module calls by (library, NID). The machine turns those stubs
// into HLE syscalls (kernel.go).
//
// The on-disc EBOOT.BIN is a KIRK-encrypted ~PSP container; LoadExecutable strips
// that wrapper first (prx.go / kirk.go) and hands the plaintext ELF here.

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	etExec = 2      // ET_EXEC
	etPRX  = 0xFFA0 // PSP relocatable module
	emMIPS = 8      // EM_MIPS
	ptLoad = 1      // PT_LOAD
)

// Segment is one PT_LOAD program-header segment: file bytes copied to VAddr, then
// zero-filled up to MemSize. FileOff is the segment's offset in the ELF file (needed
// to map a file offset — e.g. the module-info pointer — back to a virtual address).
type Segment struct {
	VAddr   uint32
	FileOff uint32
	Data    []byte // FileSize bytes
	MemSize uint32
}

// Import is one imported library: the module names Count functions, each identified
// by a 32-bit NID, whose call stubs begin at StubAddr (16 bytes apart).
type Import struct {
	Library  string
	NIDs     []uint32
	StubAddr uint32
}

// reloc is one PSP PRX relocation (SHT_PRXRELOC, type 0x700000A0): a MIPS
// relocation whose r_info selects the type and the offset/address base segments.
type reloc struct {
	offset uint32
	info   uint32
}

// Module is a parsed, decrypted PSP executable.
type Module struct {
	Name     string
	Type     uint16 // etExec or etPRX
	EntryPC  uint32
	GP       uint32
	Segments []Segment
	Imports  []Import
	relocs   []reloc

	Encrypted bool   // true if the source was a ~PSP container that we decrypted
	Tag       uint32 // the ~PSP decryption tag, when Encrypted
}

// MIPS relocation types used by PSP PRX modules.
const (
	rMIPS32   = 2
	rMIPS26   = 4
	rMIPSHI16 = 5
	rMIPSLO16 = 6
)

const shtPRXReloc = 0x700000A0

// ParseELF parses a decrypted ELF32-LE MIPS module image.
func ParseELF(b []byte) (*Module, error) {
	if len(b) < 52 || string(b[0:4]) != "\x7fELF" {
		return nil, fmt.Errorf("elf: bad magic")
	}
	if b[4] != 1 || b[5] != 1 { // ELFCLASS32, ELFDATA2LSB
		return nil, fmt.Errorf("elf: not 32-bit little-endian")
	}
	eType := binary.LittleEndian.Uint16(b[16:])
	eMachine := binary.LittleEndian.Uint16(b[18:])
	if eMachine != emMIPS {
		return nil, fmt.Errorf("elf: machine %d is not MIPS", eMachine)
	}
	entry := binary.LittleEndian.Uint32(b[24:])
	phoff := binary.LittleEndian.Uint32(b[28:])
	phentsize := int(binary.LittleEndian.Uint16(b[42:]))
	phnum := int(binary.LittleEndian.Uint16(b[44:]))

	m := &Module{Type: eType, EntryPC: entry}
	for i := 0; i < phnum; i++ {
		ph := int(phoff) + i*phentsize
		if ph+32 > len(b) {
			return nil, fmt.Errorf("elf: program header %d out of range", i)
		}
		pType := binary.LittleEndian.Uint32(b[ph:])
		if pType != ptLoad {
			continue
		}
		off := binary.LittleEndian.Uint32(b[ph+4:])
		vaddr := binary.LittleEndian.Uint32(b[ph+8:])
		filesz := binary.LittleEndian.Uint32(b[ph+16:])
		memsz := binary.LittleEndian.Uint32(b[ph+20:])
		if int(off)+int(filesz) > len(b) {
			return nil, fmt.Errorf("elf: segment %d data out of range", i)
		}
		seg := Segment{VAddr: vaddr, FileOff: off, MemSize: memsz, Data: append([]byte(nil), b[off:off+filesz]...)}
		m.Segments = append(m.Segments, seg)
	}
	if len(m.Segments) == 0 {
		return nil, fmt.Errorf("elf: no PT_LOAD segments")
	}
	m.parseModuleInfo(b)
	m.parseRelocs(b)
	return m, nil
}

// parseRelocs collects every PSP PRX relocation from the SHT_PRXRELOC sections.
func (m *Module) parseRelocs(b []byte) {
	shoff := binary.LittleEndian.Uint32(b[32:])
	shentsize := int(binary.LittleEndian.Uint16(b[46:]))
	shnum := int(binary.LittleEndian.Uint16(b[48:]))
	for i := 0; i < shnum; i++ {
		sh := int(shoff) + i*shentsize
		if sh+40 > len(b) {
			return
		}
		if binary.LittleEndian.Uint32(b[sh+4:]) != shtPRXReloc {
			continue
		}
		off := int(binary.LittleEndian.Uint32(b[sh+16:]))
		size := int(binary.LittleEndian.Uint32(b[sh+20:]))
		for p := off; p+8 <= off+size && p+8 <= len(b); p += 8 {
			m.relocs = append(m.relocs, reloc{
				offset: binary.LittleEndian.Uint32(b[p:]),
				info:   binary.LittleEndian.Uint32(b[p+4:]),
			})
		}
	}
}

// Relocate fixes up a relocatable PRX to run at load base, mutating the segment
// bytes and shifting the entry point and gp. Single-segment modules (the common
// case) resolve every offset and address base to the one segment at base. The MIPS
// HI16/LO16 pairs are combined so the split immediate carries the relocated address.
func (m *Module) Relocate(base uint32) {
	if len(m.Segments) == 0 {
		return
	}
	seg := &m.Segments[0]
	segBase := base + seg.VAddr

	read := func(off uint32) uint32 {
		if int(off)+4 > len(seg.Data) {
			return 0
		}
		return binary.LittleEndian.Uint32(seg.Data[off:])
	}
	write := func(off, v uint32) {
		if int(off)+4 <= len(seg.Data) {
			binary.LittleEndian.PutUint32(seg.Data[off:], v)
		}
	}

	var hiPending []uint32 // offsets of HI16 relocations awaiting a LO16
	for _, r := range m.relocs {
		off := r.offset // single-segment: offset is within seg
		switch r.info & 0xFF {
		case rMIPS32:
			write(off, read(off)+segBase)
		case rMIPS26:
			w := read(off)
			target := ((w & 0x03FFFFFF) << 2) + segBase
			write(off, (w&0xFC000000)|((target>>2)&0x03FFFFFF))
		case rMIPSHI16:
			hiPending = append(hiPending, off)
		case rMIPSLO16:
			lo := int32(int16(read(off)))
			for _, hoff := range hiPending {
				hi := read(hoff)
				val := ((hi & 0xFFFF) << 16) + uint32(lo) + segBase
				// carry so the signed LO16 add reconstructs the address
				hiField := (val - uint32(int32(int16(val)))) >> 16
				write(hoff, (hi&0xFFFF0000)|(hiField&0xFFFF))
			}
			hiPending = hiPending[:0]
			write(off, (read(off)&0xFFFF0000)|(uint32(int32(int16(read(off)))+int32(segBase))&0xFFFF))
		}
	}

	for i := range m.Segments {
		m.Segments[i].VAddr += base
	}
	m.EntryPC += base
	m.GP += base
}

// segData returns the module bytes visible at virtual address va for n bytes, or
// nil if the range is not covered by a loaded segment.
func (m *Module) segData(va, n uint32) []byte {
	for _, s := range m.Segments {
		if va >= s.VAddr && va+n <= s.VAddr+uint32(len(s.Data)) {
			off := va - s.VAddr
			return s.Data[off : off+n]
		}
	}
	return nil
}

// fileToVA maps an ELF file offset to a virtual address using the loaded segments.
func (m *Module) fileToVA(fileOff uint32) (uint32, bool) {
	for _, s := range m.Segments {
		if fileOff >= s.FileOff && fileOff < s.FileOff+uint32(len(s.Data)) {
			return s.VAddr + (fileOff - s.FileOff), true
		}
	}
	return 0, false
}

// parseModuleInfo locates the .rodata.sceModuleInfo structure and reads the module
// name, gp value and import stub tables. The first PT_LOAD's p_paddr holds the module
// info's *file offset*; mapping that back to a virtual address gives the structure,
// whose internal pointers are virtual addresses resolved through the loaded segments.
func (m *Module) parseModuleInfo(b []byte) {
	phoff := binary.LittleEndian.Uint32(b[28:])
	phentsize := int(binary.LittleEndian.Uint16(b[42:]))
	phnum := int(binary.LittleEndian.Uint16(b[44:]))
	var infoVA uint32
	found := false
	for i := 0; i < phnum; i++ {
		ph := int(phoff) + i*phentsize
		if ph+32 > len(b) || binary.LittleEndian.Uint32(b[ph:]) != ptLoad {
			continue
		}
		paddr := binary.LittleEndian.Uint32(b[ph+12:]) & 0x7FFFFFFF
		infoVA, found = m.fileToVA(paddr)
		break
	}
	if !found {
		return
	}
	info := m.segData(infoVA, 0x34)
	if info == nil {
		return
	}
	// sceModuleInfo: attr u16 @0x00, ver u16 @0x02, name[28] @0x04, gp u32 @0x20,
	// ent_top @0x24, ent_end @0x28, stub_top @0x2C, stub_end @0x30.
	m.Name = strings.TrimRight(string(info[4:4+28]), "\x00")
	m.GP = binary.LittleEndian.Uint32(info[0x20:])
	stubTop := binary.LittleEndian.Uint32(info[0x2C:])
	stubEnd := binary.LittleEndian.Uint32(info[0x30:])
	m.parseStubs(stubTop, stubEnd)
}

// parseStubs walks the import stub-header table [stubTop, stubEnd). Each header names
// a library and points at its NID table and its call-stub table. The length byte at
// +0x08 (in words) gives both the header size and which of the two field layouts is
// used: the 5-word (0x14) form keeps the function count and NID/stub pointers at
// +0x0A/+0x0C/+0x10; the 6-word (0x18) form at +0x0E/+0x10/+0x14.
//
//	+0x00 u32  name pointer (library name, NUL-terminated); 0 for the resident syslib
//	+0x04 u16  version
//	+0x06 u16  flags
//	+0x08 u8   header length in words (5 or 6)
func (m *Module) parseStubs(stubTop, stubEnd uint32) {
	for addr := stubTop; addr+0x14 <= stubEnd; {
		hdr := m.segData(addr, 0x18)
		if hdr == nil {
			return
		}
		namePtr := binary.LittleEndian.Uint32(hdr[0x00:])
		lenWords := hdr[0x08]

		var nfunc uint16
		var nidPtr, stubPtr uint32
		if lenWords >= 6 {
			nfunc = binary.LittleEndian.Uint16(hdr[0x0E:])
			nidPtr = binary.LittleEndian.Uint32(hdr[0x10:])
			stubPtr = binary.LittleEndian.Uint32(hdr[0x14:])
		} else {
			nfunc = binary.LittleEndian.Uint16(hdr[0x0A:])
			nidPtr = binary.LittleEndian.Uint32(hdr[0x0C:])
			stubPtr = binary.LittleEndian.Uint32(hdr[0x10:])
		}

		imp := Import{Library: cstrVA(m, namePtr), StubAddr: stubPtr}
		if nids := m.segData(nidPtr, uint32(nfunc)*4); nids != nil {
			for i := 0; i < int(nfunc); i++ {
				imp.NIDs = append(imp.NIDs, binary.LittleEndian.Uint32(nids[i*4:]))
			}
		}
		m.Imports = append(m.Imports, imp)

		step := uint32(lenWords) * 4
		if step < 0x14 {
			step = 0x14
		}
		addr += step
	}
}

// cstrVA reads a NUL-terminated string at virtual address va from the module image.
func cstrVA(m *Module, va uint32) string {
	b := m.segData(va, 1)
	if b == nil {
		return ""
	}
	// Read within the containing segment.
	for _, s := range m.Segments {
		if va >= s.VAddr && va < s.VAddr+uint32(len(s.Data)) {
			off := va - s.VAddr
			end := off
			for end < uint32(len(s.Data)) && s.Data[end] != 0 {
				end++
			}
			return string(s.Data[off:end])
		}
	}
	return ""
}

// Describe renders a human-readable summary for pspinfo -exe.
func (m *Module) Describe() string {
	var b strings.Builder
	kind := "EXEC"
	if m.Type == etPRX {
		kind = "PRX"
	}
	if m.Encrypted {
		fmt.Fprintf(&b, "decrypted ~PSP container (tag 0x%08X)\n", m.Tag)
	}
	fmt.Fprintf(&b, "module   %q  (%s)\n", m.Name, kind)
	fmt.Fprintf(&b, "entry    0x%08X   gp 0x%08X\n", m.EntryPC, m.GP)
	for _, s := range m.Segments {
		fmt.Fprintf(&b, "segment  vaddr 0x%08X  file %d  mem %d\n", s.VAddr, len(s.Data), s.MemSize)
	}
	fmt.Fprintf(&b, "imports  %d libraries:\n", len(m.Imports))
	for _, imp := range m.Imports {
		fmt.Fprintf(&b, "  %-24s %d functions @ stub 0x%08X\n", imp.Library, len(imp.NIDs), imp.StubAddr)
	}
	return b.String()
}

// LoadExecutable reads the file at path on the disc, strips the ~PSP encryption
// wrapper if present, and parses the resulting ELF/PRX module.
func (im *Image) LoadExecutable(path string) (*Module, error) {
	raw, err := im.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadModuleImage(raw)
}

// LoadModuleImage parses an executable image that is either a plaintext ELF or a
// KIRK-encrypted ~PSP container.
func LoadModuleImage(raw []byte) (*Module, error) {
	if len(raw) >= 4 && string(raw[0:4]) == "\x7fELF" {
		return ParseELF(raw)
	}
	if len(raw) >= 4 && string(raw[0:4]) == "~PSP" {
		plain, tag, err := DecryptPRX(raw)
		if err != nil {
			return nil, fmt.Errorf("decrypt ~PSP: %w", err)
		}
		m, err := ParseELF(plain)
		if err != nil {
			return nil, err
		}
		m.Encrypted = true
		m.Tag = tag
		return m, nil
	}
	return nil, fmt.Errorf("elf: not an ELF or ~PSP image (magic %q)", firstMagic(raw))
}

func firstMagic(b []byte) string {
	n := 4
	if len(b) < n {
		n = len(b)
	}
	return fmt.Sprintf("% X", b[:n])
}
