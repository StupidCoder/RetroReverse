package lm

import (
	"encoding/binary"
	"fmt"
)

// rarc.go reads the RARC archives the .szp files decompress into.
//
// The layout, read off the decompressed archives themselves: a 0x20-byte outer
// header (magic 'RARC', total size, header size, data offset, then the data
// size twice), and an inner header at +0x20 holding the directory-node count
// and offset, the file-entry count and offset, and the string-table size and
// offset — all offsets relative to the inner header. A directory node is
// 0x10 bytes: a four-character type tag, a name offset, a name hash, the
// number of entries in it and the index of its first entry. A file entry is
// 0x14 bytes: an id, the name hash, a type halfword whose 0x0200 bit marks a
// directory, the name offset, then data offset and size — for a directory the
// "data offset" is the child node's index instead.

// RARCFile is one file (or directory reference) in a RARC archive.
type RARCFile struct {
	Name string
	Dir  string // the directory-node path it lives in, "" for root
	Data []byte
}

// RARC parses a decompressed RARC archive into its files.
func RARC(b []byte) ([]RARCFile, error) {
	if len(b) < 0x40 || string(b[:4]) != "RARC" {
		return nil, fmt.Errorf("lm: not a RARC archive")
	}
	u32 := func(off uint32) uint32 { return binary.BigEndian.Uint32(b[off:]) }
	u16 := func(off uint32) uint16 { return binary.BigEndian.Uint16(b[off:]) }

	dataOff := u32(0x0C) + 0x20
	nodeCount := u32(0x20)
	nodeOff := u32(0x24) + 0x20
	entryOff := u32(0x2C) + 0x20
	strOff := u32(0x34) + 0x20

	name := func(off uint32) string {
		p := strOff + off
		e := p
		for e < uint32(len(b)) && b[e] != 0 {
			e++
		}
		return string(b[p:e])
	}

	// Walk the directory nodes; each node names the entries it owns.
	type node struct{ path string }
	nodes := make([]node, nodeCount)
	var files []RARCFile
	for n := uint32(0); n < nodeCount; n++ {
		p := nodeOff + n*0x10
		if nodes[n].path == "" && n > 0 {
			nodes[n].path = name(u32(p + 4)) // fallback: the node's own name
		}
		first := u32(p + 0xC)
		count := uint32(u16(p + 0xA))
		for i := first; i < first+count; i++ {
			e := entryOff + i*0x14
			typ := u16(e + 4)
			nameOff := uint32(u16(e + 6))
			nm := name(nameOff)
			if typ&0x0200 != 0 {
				if nm == "." || nm == ".." {
					continue
				}
				child := u32(e + 8)
				if child < nodeCount {
					sub := nodes[n].path
					if sub != "" {
						sub += "/"
					}
					nodes[child].path = sub + nm
				}
				continue
			}
			off := dataOff + u32(e+8)
			size := u32(e + 0xC)
			if off+size > uint32(len(b)) {
				return nil, fmt.Errorf("lm: RARC entry %q reaches past the archive", nm)
			}
			files = append(files, RARCFile{Name: nm, Dir: nodes[n].path, Data: b[off : off+size]})
		}
	}
	return files, nil
}
