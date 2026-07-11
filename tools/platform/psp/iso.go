package psp

// iso.go reads an ISO 9660 filesystem out of a UMD image. A UMD dump is a plain
// "cooked" 2048-byte-per-sector ISO (the raw 2352-byte CD framing of a PlayStation
// disc does not apply), so the only work is the ISO 9660 volume itself: a Primary
// Volume Descriptor at logical block 16 and a tree of directory records. Multi-byte
// numeric fields are stored "both-endian" (little then big); the little-endian half
// is read. Records are packed within a logical block and never span one.
//
// The reader sits on a blockSource — the CISO decoder (cso.go) for a .cso, or a
// flat []byte for an already-decompressed .iso — so it never needs the whole image
// resident.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

const (
	isoBlockSize = 2048 // ISO 9660 logical block size
	pvdLBA       = 16   // the Primary Volume Descriptor sits at logical block 16
)

// blockSource yields 2048-byte logical blocks by LBA. Both *CSO and byteBlocks
// satisfy it.
type blockSource interface {
	ReadBlock(n int) ([]byte, error)
}

// byteBlocks serves logical blocks from a flat in-memory image (a plain .iso).
type byteBlocks []byte

func (b byteBlocks) ReadBlock(n int) ([]byte, error) {
	off := n * isoBlockSize
	if off < 0 || off+isoBlockSize > len(b) {
		return nil, fmt.Errorf("iso: block %d out of range", n)
	}
	return b[off : off+isoBlockSize], nil
}

// Volume is an opened ISO 9660 UMD image.
type Volume struct {
	src blockSource

	System   string // PVD system identifier
	Name     string // PVD volume identifier
	rootLBA  int    // extent of the root directory
	rootSize int    // byte length of the root directory
}

// Entry is one directory entry — a file or a subdirectory. Block is the extent's
// logical block number (LBA).
type Entry struct {
	Name  string // ISO identifier, e.g. "EBOOT.BIN;1" (dirs have no version)
	Path  string // full slash path from the volume root
	IsDir bool
	Size  int // file size in bytes (directory extent size for dirs)
	Block int // extent LBA
}

func (e Entry) String() string {
	if e.IsDir {
		return fmt.Sprintf("%-32s  <dir>   (lba %d)", e.Path, e.Block)
	}
	return fmt.Sprintf("%-32s  %10d  (lba %d)", e.Path, e.Size, e.Block)
}

// OpenVolume validates the ISO 9660 Primary Volume Descriptor on src and returns a
// Volume ready for listing and extraction.
func OpenVolume(src blockSource) (*Volume, error) {
	v := &Volume{src: src}
	pvd, err := src.ReadBlock(pvdLBA)
	if err != nil {
		return nil, fmt.Errorf("iso: reading PVD: %w", err)
	}
	// PVD: type byte 1, "CD001", version 1.
	if pvd[0] != 0x01 || string(pvd[1:6]) != "CD001" {
		return nil, errors.New("iso: no ISO 9660 Primary Volume Descriptor at LBA 16")
	}
	v.System = strings.TrimRight(string(pvd[8:40]), " \x00")
	v.Name = strings.TrimRight(string(pvd[40:72]), " \x00")
	// Root directory record is 34 bytes at offset 156.
	root := pvd[156 : 156+34]
	v.rootLBA = int(le32(root[2:]))
	v.rootSize = int(le32(root[10:]))
	return v, nil
}

// OpenISO opens a flat (cooked 2048-byte) .iso image held in memory.
func OpenISO(image []byte) (*Volume, error) {
	if len(image) < (pvdLBA+1)*isoBlockSize {
		return nil, fmt.Errorf("iso: image too small (%d bytes)", len(image))
	}
	return OpenVolume(byteBlocks(image))
}

// le32 reads the little-endian half of a both-endian (or plain LE) field.
func le32(b []byte) uint32 { return binary.LittleEndian.Uint32(b[:4]) }

// dirEntries parses every record of the directory extent at (lba, size), skipping
// the "." and ".." self/parent records.
func (v *Volume) dirEntries(lba, size int, dirPath string) ([]Entry, error) {
	var out []Entry
	remaining := size
	for sect := 0; remaining > 0; sect++ {
		b, err := v.src.ReadBlock(lba + sect)
		if err != nil {
			return nil, err
		}
		n := isoBlockSize
		if remaining < n {
			n = remaining
		}
		remaining -= isoBlockSize
		for p := 0; p < n; {
			recLen := int(b[p])
			if recLen == 0 {
				break // padding to the end of this logical block
			}
			if p+recLen > isoBlockSize {
				return nil, fmt.Errorf("iso: directory record overruns sector at lba %d", lba+sect)
			}
			rec := b[p : p+recLen]
			p += recLen

			idLen := int(rec[32])
			id := rec[33 : 33+idLen]
			if idLen == 1 && (id[0] == 0x00 || id[0] == 0x01) {
				continue // "." / ".."
			}
			flags := rec[25]
			e := Entry{
				Name:  string(id),
				IsDir: flags&0x02 != 0,
				Block: int(le32(rec[2:])),
				Size:  int(le32(rec[10:])),
			}
			if dirPath == "" {
				e.Path = e.Name
			} else {
				e.Path = dirPath + "/" + e.Name
			}
			out = append(out, e)
		}
	}
	return out, nil
}

// nameEqual matches ISO identifiers case-insensitively; if want carries no ";"
// version it also matches an entry's version-stripped name (so callers may say
// "EBOOT.BIN" for "EBOOT.BIN;1").
func nameEqual(entry, want string) bool {
	if strings.EqualFold(entry, want) {
		return true
	}
	if !strings.Contains(want, ";") {
		if i := strings.IndexByte(entry, ';'); i >= 0 {
			return strings.EqualFold(entry[:i], want)
		}
	}
	return false
}

func splitPath(p string) []string {
	var parts []string
	for _, c := range strings.Split(strings.Trim(p, "/"), "/") {
		if c != "" {
			parts = append(parts, c)
		}
	}
	return parts
}

// resolve walks a slash path and returns the matching entry. The empty path is the
// root directory, returned as a synthetic entry.
func (v *Volume) resolve(path string) (Entry, error) {
	cur := Entry{Name: v.Name, Path: "", IsDir: true, Block: v.rootLBA, Size: v.rootSize}
	for _, want := range splitPath(path) {
		if !cur.IsDir {
			return Entry{}, fmt.Errorf("iso: %q is not a directory", cur.Path)
		}
		entries, err := v.dirEntries(cur.Block, cur.Size, cur.Path)
		if err != nil {
			return Entry{}, err
		}
		found := false
		for _, e := range entries {
			if nameEqual(e.Name, want) {
				cur, found = e, true
				break
			}
		}
		if !found {
			return Entry{}, fmt.Errorf("iso: %q not found", path)
		}
	}
	return cur, nil
}

// ReadDir lists the entries of the directory at path ("" or "/" is the root).
func (v *Volume) ReadDir(path string) ([]Entry, error) {
	e, err := v.resolve(path)
	if err != nil {
		return nil, err
	}
	if !e.IsDir {
		return nil, fmt.Errorf("iso: %q is not a directory", path)
	}
	return v.dirEntries(e.Block, e.Size, e.Path)
}

// Walk visits every file and directory under the root, depth first.
func (v *Volume) Walk(fn func(Entry) error) error {
	return v.walk(v.rootLBA, v.rootSize, "", fn)
}

func (v *Volume) walk(lba, size int, dirPath string, fn func(Entry) error) error {
	entries, err := v.dirEntries(lba, size, dirPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := fn(e); err != nil {
			return err
		}
		if e.IsDir {
			if err := v.walk(e.Block, e.Size, e.Path, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReadFile returns the full contents of the file at path.
// ReadFileAt reads len(p) bytes of the file entry e starting at byte offset off,
// returning the count read (short at EOF).
func (v *Volume) ReadFileAt(e Entry, off int, p []byte) (int, error) {
	if off >= e.Size {
		return 0, nil
	}
	if off+len(p) > e.Size {
		p = p[:e.Size-off]
	}
	got := 0
	for got < len(p) {
		lba := e.Block + (off+got)/isoBlockSize
		b, err := v.src.ReadBlock(lba)
		if err != nil {
			return got, err
		}
		got += copy(p[got:], b[(off+got)%isoBlockSize:])
	}
	return got, nil
}

func (v *Volume) ReadFile(path string) ([]byte, error) {
	e, err := v.resolve(path)
	if err != nil {
		return nil, err
	}
	if e.IsDir {
		return nil, fmt.Errorf("iso: %q is a directory", path)
	}
	out := make([]byte, 0, e.Size)
	for got := 0; got < e.Size; got += isoBlockSize {
		b, err := v.src.ReadBlock(e.Block + got/isoBlockSize)
		if err != nil {
			return nil, err
		}
		n := e.Size - got
		if n > isoBlockSize {
			n = isoBlockSize
		}
		out = append(out, b[:n]...)
	}
	return out, nil
}
