// Package goalobj reads the game's GOAL object containers and reproduces its
// linker.
//
// Everything here is derived from the boot ELF's own code, read with the
// oracle's -dis on the named C++ functions it ships symbols for:
//
//   - The DGO container is read by BeginLoadingDGO/GetNextDGO (OVERLORD side):
//     a 64-byte header {u32 count, char name[60]}, then per entry a 64-byte
//     header {u32 size, char name[60]} followed by the payload, each payload
//     16-byte aligned.
//
//   - An entry's payload is a GOAL object file, linked in place by
//     link_control::begin/work/finish (0x1097F0/0x109A50/0x10AAF0).
//     begin reads word[2] as the file format version. Version >= 4 (all the
//     art/data objects): a 16-byte header {tag, linkSize, version, objSize},
//     the object itself at data+16, and a TRAILING link block at
//     data + 16 + objSize: {typeSlot (stamped with the link-block type at
//     runtime), blockSize, wireVersion, stream...} — the stream starts at
//     data + objSize + 28. Version < 4 (the v3 compiled-code objects): the
//     link block is at the HEAD — {tag, linkSize, version} — with the object
//     at data + linkSize and the stream at data+12. The file version and the
//     wire version are distinct: v4 files carry a v2 wire stream (observed
//     live: dir-tpages is file v4 / wire v2). work dispatches on the WIRE
//     version word next to the stream.
//
//   - work dispatches version 2 AND 4 to work_v2 (0x10A590), version 3 to
//     work_v3. work_v2 runs two phases over the stream:
//
//     Phase 1 — pointer fixups (only if its first byte is nonzero, else the
//     byte is consumed and the phase is skipped): a run-length byte stream
//     alternating skip-N-words / relocate-N-words over the object, starting
//     in skip mode. Relocating a word adds the object's base address to it.
//     A byte of 255 does not toggle the mode (an extended run), except when
//     followed by a 0 byte, which is consumed and toggles. After any other
//     byte, a following 0 byte is consumed and ends the phase.
//
//     Phase 2 — symbol links: records until a 0 control byte. A control byte
//     with bit 7 set interns a TYPE with (c & 0x7F) methods (minimum 1,
//     intern_type_from_c) and patches with the type object's address; any
//     other control byte interns a SYMBOL (intern_from_c) and patches with
//     the symbol cell's address. If the control byte is >= 10 it is itself
//     the first character of the NUL-terminated name that follows (the
//     stream backs up one byte); bytes 1..9 would leave the name to start
//     after the byte (never seen in practice). After the name, c_symlink2
//     (0x10B248) applies a patch list: delta-encoded byte offsets between
//     patched words (low 2 bits of the first byte select 1/2/3/4-byte
//     little-endian varints, then cleared), starting from the object base.
//     A patched word of 0xFFFFFFFF is replaced by the target address;
//     anything else keeps its high 16 bits and gets target - s7 added to its
//     low 16 bits (an s7-relative instruction fixup). A 0 byte after a patch
//     ends the list.
//
// Version 3 (compiled GOAL code with multiple segments, work_v3/c_rellink3/
// c_symlink3) is not needed for art objects and is not implemented yet.
package goalobj

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DGOEntry is one named object inside a DGO/CGO archive.
type DGOEntry struct {
	Name string
	Data []byte
}

// DGO is a parsed archive.
type DGO struct {
	Name    string
	Entries []DGOEntry
}

func cstr(b []byte) string {
	if i := strings.IndexByte(string(b), 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

// ReadDGO parses a DGO/CGO archive image.
func ReadDGO(data []byte) (*DGO, error) {
	if len(data) < 64 {
		return nil, fmt.Errorf("dgo: %d bytes is too short for a header", len(data))
	}
	n := int(binary.LittleEndian.Uint32(data))
	d := &DGO{Name: cstr(data[4:64])}
	off := 64
	for i := 0; i < n; i++ {
		if off+64 > len(data) {
			return nil, fmt.Errorf("dgo %s: entry %d header at 0x%x runs past the file", d.Name, i, off)
		}
		size := int(binary.LittleEndian.Uint32(data[off:]))
		name := cstr(data[off+4 : off+64])
		p := off + 64
		if p+size > len(data) {
			return nil, fmt.Errorf("dgo %s: entry %q (%d bytes at 0x%x) runs past the file", d.Name, name, size, p)
		}
		d.Entries = append(d.Entries, DGOEntry{Name: name, Data: data[p : p+size]})
		off = p + ((size + 15) &^ 15)
	}
	return d, nil
}

// Symbol is one GOAL symbol table cell.
type Symbol struct {
	Cell  uint32 // address of the symbol cell (what a symbol link patches in)
	Value uint32 // the cell's value (what a type link patches in)
}

// SymTab is the runtime GOAL symbol table, as dumped by the oracle's
// -goalsyms (or work/goal.txt).
type SymTab struct {
	S7   uint32 // the symbol base register the engine keeps in $s7
	Syms map[string]Symbol
}

// LoadSymTab reads a goal.txt-style dump: a first line naming the s7 base,
// then "CELLADDR OFFSET VALUE name" per line.
func LoadSymTab(path string) (*SymTab, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st := &SymTab{Syms: make(map[string]Symbol)}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.Contains(line, "symbol base") {
			if i := strings.Index(line, "0x"); i >= 0 {
				v, err := strconv.ParseUint(strings.Fields(line[i:])[0][2:], 16, 32)
				if err != nil {
					return nil, fmt.Errorf("symtab: bad s7 line %q", line)
				}
				st.S7 = uint32(v)
			}
			continue
		}
		f := strings.Fields(line)
		if len(f) != 4 {
			continue
		}
		cell, err1 := strconv.ParseUint(f[0], 16, 32)
		val, err2 := strconv.ParseUint(f[2], 16, 32)
		if err1 != nil || err2 != nil {
			continue
		}
		st.Syms[f[3]] = Symbol{Cell: uint32(cell), Value: uint32(val)}
	}
	if st.S7 == 0 {
		return nil, fmt.Errorf("symtab %s: no symbol-base line", path)
	}
	return st, sc.Err()
}

// LinkReport says what Link did, for inspection and for downstream tools that
// want the semantic relocation map rather than the bytes.
type LinkReport struct {
	Version   int
	Pointers  []uint32          // object-relative byte offsets of words that got the base added
	SymbolRef map[uint32]string // object-relative byte offset -> symbol/type name patched there
	Missing   []string          // names the symbol table did not know
}

// Link reproduces link_control on one object file: it returns the object's
// linked in-memory image as it would sit at runtime address base, plus the
// report. The input slice is not modified.
func Link(data []byte, base uint32, st *SymTab) ([]byte, *LinkReport, error) {
	if len(data) < 16 {
		return nil, nil, fmt.Errorf("goalobj: %d bytes is too short", len(data))
	}
	w := func(off int) uint32 { return binary.LittleEndian.Uint32(data[off:]) }
	version := int(w(8))
	var obj []byte
	var stream []byte
	var wire int
	switch {
	case version >= 4:
		objSize := int(w(12))
		if 16+objSize+12 > len(data) {
			return nil, nil, fmt.Errorf("goalobj: v%d object size 0x%x out of range", version, objSize)
		}
		obj = append([]byte(nil), data[16:16+objSize]...)
		wire = int(w(16 + objSize + 8))
		stream = data[16+objSize+12:]
	case version == 3:
		return nil, nil, fmt.Errorf("goalobj: v3 (segmented code) not implemented")
	case version == 2:
		// v2 (STR spool chunks): the link block LEADS — {tag, linkDataSize,
		// ver} header, wire stream at +12, object at +linkDataSize (the size
		// counts the header). The stream itself is the same wire v2.
		linkSize := int(w(4))
		if linkSize < 12 || linkSize > len(data) {
			return nil, nil, fmt.Errorf("goalobj: v2 link block size 0x%x out of range", linkSize)
		}
		obj = append([]byte(nil), data[linkSize:]...)
		wire = 2
		stream = data[12:linkSize]
	default:
		return nil, nil, fmt.Errorf("goalobj: unknown version %d", version)
	}
	if wire != 2 && wire != 4 {
		return nil, nil, fmt.Errorf("goalobj: wire version %d (want 2 or 4)", wire)
	}
	rep := &LinkReport{Version: version, SymbolRef: make(map[uint32]string)}

	i := 0
	next := func() (byte, error) {
		if i >= len(stream) {
			return 0, fmt.Errorf("goalobj: link stream truncated at %d", i)
		}
		c := stream[i]
		i++
		return c, nil
	}

	// Phase 1: pointer fixups.
	if i < len(stream) && stream[i] == 0 {
		i++
	} else {
		word := 0 // cursor, in words from the object base
		reloc := false
		for {
			c, err := next()
			if err != nil {
				return nil, nil, err
			}
			if reloc {
				for k := 0; k < int(c); k++ {
					off := word * 4
					if off+4 > len(obj) {
						return nil, nil, fmt.Errorf("goalobj: pointer fixup at 0x%x runs past the object", off)
					}
					v := binary.LittleEndian.Uint32(obj[off:])
					binary.LittleEndian.PutUint32(obj[off:], v+base)
					rep.Pointers = append(rep.Pointers, uint32(off))
					word++
				}
			} else {
				word += int(c)
			}
			if c == 255 {
				if i < len(stream) && stream[i] == 0 {
					i++
					reloc = !reloc
				}
				continue
			}
			reloc = !reloc
			if i < len(stream) && stream[i] == 0 {
				i++
				break
			}
		}
	}

	// Phase 2: symbol links.
	for {
		c, err := next()
		if err != nil {
			return nil, nil, err
		}
		if c == 0 {
			break
		}
		isType := c&0x80 != 0
		if !isType && c >= 10 {
			i-- // the control byte is the name's first character
		}
		nul := strings.IndexByte(string(stream[i:]), 0)
		if nul < 0 {
			return nil, nil, fmt.Errorf("goalobj: unterminated name in symbol links")
		}
		name := string(stream[i : i+nul])
		i += nul + 1

		var target uint32
		sym, known := st.Syms[name]
		if !known {
			rep.Missing = append(rep.Missing, name)
		} else if isType {
			target = sym.Value
		} else {
			target = sym.Cell
		}

		// The patch list: delta offsets from the object base.
		loc := uint32(0)
		for {
			b0, err := next()
			if err != nil {
				return nil, nil, err
			}
			off := uint32(b0)
			if off&3 != 0 {
				b1, _ := next()
				off |= uint32(b1) << 8
				if off&2 != 0 {
					b2, _ := next()
					off |= uint32(b2) << 16
					if off&1 != 0 {
						b3, _ := next()
						off |= uint32(b3) << 24
					}
				}
			}
			loc += off &^ 3
			if int(loc)+4 > len(obj) {
				return nil, nil, fmt.Errorf("goalobj: symbol patch for %q at 0x%x runs past the object", name, loc)
			}
			rep.SymbolRef[loc] = name
			if known {
				v := binary.LittleEndian.Uint32(obj[loc:])
				if v == 0xFFFFFFFF {
					v = target
				} else {
					v = v&0xFFFF0000 | (v+target-st.S7)&0xFFFF
				}
				binary.LittleEndian.PutUint32(obj[loc:], v)
			}
			if i < len(stream) && stream[i] == 0 {
				i++
				break
			}
		}
	}
	return obj, rep, nil
}
