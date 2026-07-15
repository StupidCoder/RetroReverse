package xbox

// xbe.go parses an XBE — the original Xbox executable, a PE derivative.
//
// An XBE begins with an "XBEH" header giving the image base (0x00010000 for retail
// titles), the image size, and virtual addresses for its section table, its
// certificate, and the two things this parser exists to recover:
//
//   - the *entry point*, and
//   - the *kernel image thunk table* — the list of xboxkrnl.exe exports the title
//     imports.
//
// Both are lightly obfuscated: stored XOR'd with a constant that differs between retail
// and debug images. This is not encryption — it is a platform-spec constant, derivable
// from the image family, exactly like the PSP's KIRK keys or the DOS phase's go32 base
// (see the clean-room note in the xbox-oracle memory). The parser tries both keys and
// keeps whichever de-obfuscates to an address that lands inside the image.
//
// The thunk table is a NUL-terminated array of DWORDs. An entry with its high bit set
// imports a kernel export by *ordinal* (its low 16 bits); the sorted, de-duplicated
// ordinal list is what scopes the kernel HLE the machine will later need.
//
// The on-disc layout here is general Xbox-platform knowledge (like the DPMI/COFF layout
// in the DOS phase). Every game-specific fact — which ordinals THIS title imports, what
// sections it carries — is read from the image itself, never from an external database.

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"unicode/utf16"
)

// XOR keys for de-obfuscating the entry point and the kernel-thunk table. Retail and
// debug images use different constants; the parser tries both.
const (
	entryKeyRetail = 0xA8FC57AB
	entryKeyDebug  = 0x94859D4B
	thunkKeyRetail = 0x5B6D40B6
	thunkKeyDebug  = 0xEFB1F152
)

const xbeMagic = "XBEH"

// Section flag bits.
const (
	SecWritable   = 0x00000001
	SecPreload    = 0x00000002
	SecExecutable = 0x00000004
	SecInserted   = 0x00000008
	SecHeadPageRO = 0x00000010
	SecTailPageRO = 0x00000020
)

// Section is one XBE section header.
type Section struct {
	Name     string
	Flags    uint32
	VAddr    uint32 // virtual address the section loads at
	VSize    uint32 // virtual size
	RawAddr  uint32 // file offset of the section's data
	RawSize  uint32 // bytes of data in the file
	NameAddr uint32 // VA of the name string
}

// Flag string helpers, for the CLI.
func (s Section) FlagString() string {
	var f []string
	if s.Flags&SecWritable != 0 {
		f = append(f, "W")
	}
	if s.Flags&SecExecutable != 0 {
		f = append(f, "X")
	}
	if s.Flags&SecPreload != 0 {
		f = append(f, "preload")
	}
	if s.Flags&SecInserted != 0 {
		f = append(f, "inserted")
	}
	if len(f) == 0 {
		return "-"
	}
	return strings.Join(f, "|")
}

// XBE is a parsed executable header.
type XBE struct {
	Base      uint32 // image base address (VA the header loads at)
	ImageSize uint32
	Entry     uint32 // de-obfuscated entry point
	ThunkAddr uint32 // de-obfuscated VA of the kernel thunk table
	Retail    bool   // true if the retail XOR keys de-obfuscated cleanly

	TitleName string // from the certificate, if readable
	TitleID   uint32 // from the certificate
	Sections  []Section
	Ordinals  []uint16 // sorted, de-duplicated xboxkrnl import ordinals

	raw []byte // the whole file, retained for VA lookups
}

// ParseXBE parses a whole XBE image held in memory.
func ParseXBE(b []byte) (*XBE, error) {
	if len(b) < 0x178 {
		return nil, fmt.Errorf("xbe: too small (%d bytes) to hold a header", len(b))
	}
	if string(b[0:4]) != xbeMagic {
		return nil, fmt.Errorf("xbe: bad magic %q, want %q", b[0:4], xbeMagic)
	}
	x := &XBE{raw: b}
	x.Base = le32(b[0x104:])
	x.ImageSize = le32(b[0x10C:])
	entryRaw := le32(b[0x128:])
	thunkRaw := le32(b[0x158:])
	certAddr := le32(b[0x118:])
	nSections := le32(b[0x11C:])
	secHdrAddr := le32(b[0x120:])

	// Sections first: the address-validity test for the XOR keys is "does the result
	// land inside a section", so they must be parsed before de-obfuscation is decided.
	if err := x.parseSections(nSections, secHdrAddr); err != nil {
		return nil, err
	}

	// De-obfuscate the entry point: the key whose result is a valid image VA wins.
	if e := entryRaw ^ entryKeyRetail; x.inImage(e) {
		x.Entry, x.Retail = e, true
	} else if e := entryRaw ^ entryKeyDebug; x.inImage(e) {
		x.Entry, x.Retail = e, false
	} else {
		// Neither lands cleanly; keep the retail attempt so callers can still see it.
		x.Entry = entryRaw ^ entryKeyRetail
	}

	// De-obfuscate the thunk-table pointer independently, by the same test.
	if t := thunkRaw ^ thunkKeyRetail; x.inImage(t) {
		x.ThunkAddr = t
	} else if t := thunkRaw ^ thunkKeyDebug; x.inImage(t) {
		x.ThunkAddr = t
	} else {
		x.ThunkAddr = thunkRaw ^ thunkKeyRetail
	}

	x.parseCertificate(certAddr)
	if err := x.parseThunks(); err != nil {
		return nil, err
	}
	return x, nil
}

// inImage reports whether a VA falls within the loaded image.
func (x *XBE) inImage(va uint32) bool {
	return va >= x.Base && va < x.Base+x.ImageSize
}

// atVA returns the file offset for a virtual address, if it lies inside the header
// region or a section's raw data. The header maps 1:1 from base; each section maps its
// virtual range onto its raw range.
func (x *XBE) atVA(va uint32) (int, bool) {
	// Header region: VA base..base+headerRawUsed maps straight to file offset va-base.
	// The header sits at file offset 0, so treat any VA below the first section's VA
	// (but >= base) as a header offset.
	for _, s := range x.Sections {
		if va >= s.VAddr && va < s.VAddr+s.RawSize {
			return int(s.RawAddr + (va - s.VAddr)), true
		}
	}
	// Fall back to the header mapping (base-relative == file-relative for the header).
	if va >= x.Base {
		off := int(va - x.Base)
		if off < len(x.raw) {
			return off, true
		}
	}
	return 0, false
}

func (x *XBE) parseSections(n, addr uint32) error {
	if n == 0 || n > 4096 {
		return fmt.Errorf("xbe: implausible section count %d", n)
	}
	off, ok := x.atVA(addr)
	if !ok {
		// Sections table sits in the header; map by base.
		if addr < x.Base {
			return fmt.Errorf("xbe: section headers at %#x precede the image base %#x", addr, x.Base)
		}
		off = int(addr - x.Base)
	}
	const secHdrSize = 0x38
	if off+int(n)*secHdrSize > len(x.raw) {
		return fmt.Errorf("xbe: %d section headers at %#x overrun the %d-byte image", n, off, len(x.raw))
	}
	x.Sections = make([]Section, n)
	for i := 0; i < int(n); i++ {
		h := x.raw[off+i*secHdrSize:]
		s := Section{
			Flags:    le32(h[0x00:]),
			VAddr:    le32(h[0x04:]),
			VSize:    le32(h[0x08:]),
			RawAddr:  le32(h[0x0C:]),
			RawSize:  le32(h[0x10:]),
			NameAddr: le32(h[0x14:]),
		}
		x.Sections[i] = s
	}
	// Resolve section names now that the raw ranges are known (the name strings live in
	// the header region, addressed by VA).
	for i := range x.Sections {
		x.Sections[i].Name = x.cstrVA(x.Sections[i].NameAddr)
	}
	return nil
}

// cstrVA reads a NUL-terminated ASCII string at a virtual address.
func (x *XBE) cstrVA(va uint32) string {
	off, ok := x.atVA(va)
	if !ok {
		return ""
	}
	end := off
	for end < len(x.raw) && x.raw[end] != 0 {
		end++
	}
	return string(x.raw[off:end])
}

// parseCertificate reads the title id and name (UTF-16LE) from the certificate.
func (x *XBE) parseCertificate(va uint32) {
	off, ok := x.atVA(va)
	if !ok || off+0x0C+80 > len(x.raw) {
		return
	}
	x.TitleID = le32(x.raw[off+0x08:])
	// Title name: 40 UTF-16LE code units at cert+0x0C.
	u := make([]uint16, 0, 40)
	for i := 0; i < 40; i++ {
		c := le16(x.raw[off+0x0C+i*2:])
		if c == 0 {
			break
		}
		u = append(u, c)
	}
	x.TitleName = strings.TrimRight(string(utf16.Decode(u)), "\x00 ")
}

// parseThunks walks the NUL-terminated kernel thunk table and collects the ordinals.
func (x *XBE) parseThunks() error {
	off, ok := x.atVA(x.ThunkAddr)
	if !ok {
		return fmt.Errorf("xbe: kernel thunk table VA %#x is not inside the image", x.ThunkAddr)
	}
	set := make(map[uint16]bool)
	for p := off; p+4 <= len(x.raw); p += 4 {
		v := le32(x.raw[p:])
		if v == 0 {
			break // NUL terminator
		}
		if v&0x80000000 != 0 {
			set[uint16(v&0xFFFF)] = true
		}
		// Non-ordinal (by-name) kernel imports do not occur for xboxkrnl; every entry
		// carries the ordinal flag. A stray non-flagged entry is simply skipped.
	}
	x.Ordinals = make([]uint16, 0, len(set))
	for o := range set {
		x.Ordinals = append(x.Ordinals, o)
	}
	sort.Slice(x.Ordinals, func(i, j int) bool { return x.Ordinals[i] < x.Ordinals[j] })
	return nil
}

func le16(b []byte) uint16 { return binary.LittleEndian.Uint16(b[:2]) }
func le32(b []byte) uint32 { return binary.LittleEndian.Uint32(b[:4]) }
