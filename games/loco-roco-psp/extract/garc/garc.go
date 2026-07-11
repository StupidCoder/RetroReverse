// Package garc reads the game's GARC resource archives and the GIMG sector
// directory that maps file names to raw extents of DATA.BIN.
package garc

import (
	"encoding/binary"
	"fmt"
	"math"
)

// File is one entry of a GARC FILE chunk.
type File struct {
	Ext  string // extension tag ("tip", "png", "sgd", "bin", ...)
	Name string // file name from the NAME chunk
	Off  uint32 // absolute offset of the data within the archive
	Size uint32 // byte size
}

// Archive is a parsed GARC.
type Archive struct {
	Files []File
	data  []byte
}

func u32(b []byte, o uint32) uint32 { return binary.LittleEndian.Uint32(b[o : o+4]) }

func cstr(b []byte, o uint32) string {
	end := o
	for end < uint32(len(b)) && b[end] != 0 {
		end++
	}
	return string(b[o:end])
}

// Parse reads a decompressed GARC image. Chunks carry the absolute offset of
// the next chunk at +4 and an entry count at +8; the FILE chunk's 24-byte
// entries name each packed file's extension, size, data offset and name offset.
func Parse(data []byte) (*Archive, error) {
	if len(data) < 0x40 || string(data[0:4]) != "GARC" {
		return nil, fmt.Errorf("garc: bad magic")
	}
	if v := math.Float32frombits(u32(data, 4)); v != 1.0 {
		return nil, fmt.Errorf("garc: version %v", v)
	}
	a := &Archive{data: data}
	off := u32(data, 0x14)
	for off != 0 && off+0x20 <= uint32(len(data)) {
		magic := string(data[off : off+4])
		next := u32(data, off+4)
		count := u32(data, off+8)
		if magic == "TERM" {
			break
		}
		if magic == "FILE" {
			for i := uint32(0); i < count; i++ {
				e := off + 0x20 + i*24
				if e+24 > uint32(len(data)) {
					return nil, fmt.Errorf("garc: FILE entry %d out of range", i)
				}
				f := File{
					Ext:  cstr(data, e),
					Size: u32(data, e+4),
					Off:  u32(data, e+8),
				}
				if n := u32(data, e+12); n < uint32(len(data)) {
					f.Name = cstr(data, n)
				}
				a.Files = append(a.Files, f)
			}
		}
		if next <= off {
			break
		}
		off = next
	}
	return a, nil
}

// Data returns the packed bytes of an entry.
func (a *Archive) Data(f File) []byte { return a.data[f.Off : f.Off+f.Size] }

// Find returns the entry with the given name.
func (a *Archive) Find(name string) (File, bool) {
	for _, f := range a.Files {
		if f.Name == name {
			return f, true
		}
	}
	return File{}, false
}

// GimgEntry is one entry of a GIMG sector directory (sector_usa.bin): a file
// name mapped to its sector offset within DATA.BIN and its byte size. The game
// resolves these to disc0:/sce_lbn0x%X_size0x%X extents by adding DATA.BIN's
// base LBN.
type GimgEntry struct {
	Name   string
	Sector uint32 // sector offset within DATA.BIN (2048-byte sectors)
	Hash   uint32
	Size   uint32 // byte size
}

// ParseGimg reads a GIMG sector directory.
func ParseGimg(data []byte) ([]GimgEntry, error) {
	if len(data) < 0x10 || string(data[0:4]) != "GIMG" {
		return nil, fmt.Errorf("gimg: bad magic")
	}
	count := u32(data, 8)
	out := make([]GimgEntry, 0, count)
	for i := uint32(0); i < count; i++ {
		e := 0xC + i*16
		if e+16 > uint32(len(data)) {
			return nil, fmt.Errorf("gimg: entry %d out of range", i)
		}
		g := GimgEntry{
			Sector: u32(data, e+4),
			Hash:   u32(data, e+8),
			Size:   u32(data, e+12),
		}
		if n := u32(data, e); n < uint32(len(data)) {
			g.Name = cstr(data, n)
		}
		out = append(out, g)
	}
	return out, nil
}
