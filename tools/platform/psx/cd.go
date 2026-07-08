package psx

// cd.go reads an ISO 9660 filesystem out of a PlayStation CD image.
//
// A PSX disc's data track is normally dumped as raw 2352-byte CD sectors:
//
//	 0x000  12  sync           00 FF FF FF FF FF FF FF FF FF FF 00
//	 0x00C   3  address        minute, second, frame (BCD)
//	 0x00F   1  mode           01 = Mode 1, 02 = Mode 2
//	 0x010   8  subheader      (Mode 2 only) file/channel/submode/coding ×2
//	 0x018 2048  user data      (Mode 1: at 0x010; Mode 2 Form 1: at 0x018)
//	  ...        EDC / ECC
//
// So the only real work versus a flat disk image is stripping the sync/header/
// subheader to expose the 2048-byte user area of each logical block; the block()
// accessor below is exactly the seam the adf reader uses (tools/amiga/adf), with
// the CD de-framing folded in. On top of that user data sits a textbook ISO 9660
// volume: a Primary Volume Descriptor at logical block 16 and a tree of directory
// records. Multi-byte numeric fields are stored "both-endian" (little then big);
// we read the little-endian half.
//
// Some images are instead a "cooked" 2048-byte-per-sector .iso; Open detects
// that too and treats each 2048-byte block as the user data directly.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

const (
	rawSectorSize = 2352 // a full CD sector
	userSize      = 2048 // ISO 9660 logical block size / Mode-1|2-Form-1 payload
	pvdLBA        = 16    // the Primary Volume Descriptor sits at logical block 16
)

// Volume is an opened PSX/ISO 9660 disc image.
type Volume struct {
	img     []byte
	stride  int // bytes per sector in the image (2352 raw, 2048 cooked)
	dataOff int // offset of the 2048-byte user area within a sector
	nsect   int // number of sectors in the image

	System   string // PVD system identifier (e.g. "PLAYSTATION")
	Name     string // PVD volume identifier (e.g. "RIDGERACERUSA")
	rootLBA  int    // extent of the root directory
	rootSize int    // byte length of the root directory
}

// Entry is one directory entry — a file or a subdirectory. Block is the extent's
// logical block number (LBA), the same unit used by block().
type Entry struct {
	Name  string // ISO identifier, e.g. "SYSTEM.CNF;1" (dirs have no version)
	Path  string // full slash path from the volume root
	IsDir bool
	Size  int // file size in bytes (directory extent size for dirs)
	Block int // extent LBA
}

func (e Entry) String() string {
	if e.IsDir {
		return fmt.Sprintf("%-28s  <dir>   (lba %d)", e.Path, e.Block)
	}
	return fmt.Sprintf("%-28s  %8d  (lba %d)", e.Path, e.Size, e.Block)
}

// Open validates the image geometry and the ISO 9660 Primary Volume Descriptor
// and returns a Volume ready for listing and extraction.
func Open(image []byte) (*Volume, error) {
	v := &Volume{img: image}
	switch {
	case len(image) >= rawSectorSize && len(image)%rawSectorSize == 0:
		v.stride = rawSectorSize
		v.nsect = len(image) / rawSectorSize
		// Mode byte at 0x0F: 1 → user data at 0x10, 2 → Form-1 user data at 0x18.
		switch image[0x0F] {
		case 1:
			v.dataOff = 0x10
		case 2:
			v.dataOff = 0x18
		default:
			return nil, fmt.Errorf("psx: unknown CD sector mode %d", image[0x0F])
		}
	case len(image) >= userSize && len(image)%userSize == 0:
		v.stride = userSize
		v.dataOff = 0
		v.nsect = len(image) / userSize
	default:
		return nil, fmt.Errorf("psx: not a CD image (%d bytes)", len(image))
	}

	pvd, err := v.block(pvdLBA)
	if err != nil {
		return nil, fmt.Errorf("psx: reading PVD: %w", err)
	}
	// PVD: type byte 1, "CD001", version 1.
	if pvd[0] != 0x01 || string(pvd[1:6]) != "CD001" {
		return nil, errors.New("psx: no ISO 9660 Primary Volume Descriptor at LBA 16")
	}
	v.System = strings.TrimRight(string(pvd[8:40]), " \x00")
	v.Name = strings.TrimRight(string(pvd[40:72]), " \x00")
	// Root directory record is 34 bytes at offset 156.
	root := pvd[156 : 156+34]
	v.rootLBA = int(le32(root[2:]))
	v.rootSize = int(le32(root[10:]))
	return v, nil
}

// le32 reads the little-endian half of a both-endian (or plain LE) field.
func le32(b []byte) uint32 { return binary.LittleEndian.Uint32(b[:4]) }

// block returns the 2048-byte user data of logical block (sector) n.
func (v *Volume) block(n int) ([]byte, error) {
	if n < 0 || n >= v.nsect {
		return nil, fmt.Errorf("psx: sector %d out of range (0..%d)", n, v.nsect-1)
	}
	off := n*v.stride + v.dataOff
	return v.img[off : off+userSize], nil
}

// dirEntries parses every record of the directory extent at (lba, size),
// skipping the "." and ".." self/parent records.
func (v *Volume) dirEntries(lba, size int, dirPath string) ([]Entry, error) {
	var out []Entry
	remaining := size
	for sect := 0; remaining > 0; sect++ {
		b, err := v.block(lba + sect)
		if err != nil {
			return nil, err
		}
		n := userSize
		if remaining < n {
			n = remaining
		}
		remaining -= userSize
		// Records are packed within a logical block and never span one; a zero
		// length byte means the rest of this block is padding.
		for p := 0; p < n; {
			recLen := int(b[p])
			if recLen == 0 {
				break // padding to the end of this logical block
			}
			if p+recLen > userSize {
				return nil, fmt.Errorf("psx: directory record overruns sector at lba %d", lba+sect)
			}
			rec := b[p : p+recLen]
			p += recLen

			idLen := int(rec[32])
			id := rec[33 : 33+idLen]
			// Skip "." (0x00) and ".." (0x01) self/parent entries.
			if idLen == 1 && (id[0] == 0x00 || id[0] == 0x01) {
				continue
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
// "SYSTEM.CNF" for "SYSTEM.CNF;1").
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

// resolve walks a slash path and returns the matching entry. The empty path is
// the root directory, returned as a synthetic entry.
func (v *Volume) resolve(path string) (Entry, error) {
	cur := Entry{Name: v.Name, Path: "", IsDir: true, Block: v.rootLBA, Size: v.rootSize}
	for _, want := range splitPath(path) {
		if !cur.IsDir {
			return Entry{}, fmt.Errorf("psx: %q is not a directory", cur.Path)
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
			return Entry{}, fmt.Errorf("psx: %q not found", path)
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
		return nil, fmt.Errorf("psx: %q is not a directory", path)
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
func (v *Volume) ReadFile(path string) ([]byte, error) {
	e, err := v.resolve(path)
	if err != nil {
		return nil, err
	}
	if e.IsDir {
		return nil, fmt.Errorf("psx: %q is a directory", path)
	}
	out := make([]byte, 0, e.Size)
	for got := 0; got < e.Size; got += userSize {
		b, err := v.block(e.Block + got/userSize)
		if err != nil {
			return nil, err
		}
		n := e.Size - got
		if n > userSize {
			n = userSize
		}
		out = append(out, b[:n]...)
	}
	return out, nil
}
