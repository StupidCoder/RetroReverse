// Package assets reads Crazy Taxi's disc-side asset containers and textures.
//
// Everything here is derived from the bytes on the disc with our own tools
// (the clean-room rule): the AFS layout was read off the containers' own
// tables, and the PVRT texture layout off 0GDTEX.PVR, whose decode is
// verifiable by eye — it is the artwork the console's own boot menu shows
// for the disc.
package assets

import (
	"encoding/binary"
	"fmt"
)

// AFSEntry is one file in an AFS container. The containers on this disc
// carry no name table (the word after the last table pair is zero), so an
// entry's identity is its index.
type AFSEntry struct {
	Index  int
	Offset uint32
	Size   uint32
}

// AFS is a parsed container.
type AFS struct {
	Entries []AFSEntry
	data    []byte
}

// OpenAFS parses an AFS container: "AFS\0", a u32 entry count, then
// (offset, size) u32 pairs. Offsets are from the start of the file; the
// data area begins at 0x80000 on every container on this disc.
func OpenAFS(data []byte) (*AFS, error) {
	if len(data) < 8 || string(data[:4]) != "AFS\x00" {
		return nil, fmt.Errorf("afs: no AFS magic")
	}
	n := int(binary.LittleEndian.Uint32(data[4:]))
	if n <= 0 || 8+8*n > len(data) {
		return nil, fmt.Errorf("afs: entry count %d does not fit the file", n)
	}
	a := &AFS{data: data}
	for i := 0; i < n; i++ {
		off := binary.LittleEndian.Uint32(data[8+8*i:])
		size := binary.LittleEndian.Uint32(data[12+8*i:])
		if uint64(off)+uint64(size) > uint64(len(data)) {
			return nil, fmt.Errorf("afs: entry %d (%#x+%#x) overruns the file", i, off, size)
		}
		a.Entries = append(a.Entries, AFSEntry{Index: i, Offset: off, Size: size})
	}
	return a, nil
}

// Data returns one entry's bytes (a view, not a copy).
func (a *AFS) Data(i int) ([]byte, error) {
	if i < 0 || i >= len(a.Entries) {
		return nil, fmt.Errorf("afs: no entry %d (container has %d)", i, len(a.Entries))
	}
	e := a.Entries[i]
	return a.data[e.Offset : e.Offset+e.Size], nil
}
