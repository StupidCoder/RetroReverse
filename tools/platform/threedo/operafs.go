package threedo

// operafs.go reads the "Opera" file system off a 3DO Interactive Multiplayer CD
// image. It is the 3DO counterpart of tools/psx/cd.go: the low sector/geometry
// layer is the same (a data track is normally dumped as raw 2352-byte CD sectors
// whose 2048-byte user area we expose through block()), but the file system that
// sits on top is entirely different from ISO 9660 and is what this file decodes.
//
// Opera differs from ISO 9660 in three ways that matter here:
//
//   - Everything is BIG-endian (the 3DO's ARM60 runs big-endian). ISO stores
//     numbers "both-endian"; Opera stores plain big-endian words.
//   - The volume label lives at logical block 0, not block 16, and is identified
//     by a record byte 0x01 followed by five 0x5A sync bytes (not "CD001").
//   - Every file and directory exists as one or more identical copies ("avatars")
//     scattered across the disc to cut seek time and add redundancy. A directory
//     entry lists all of a file's avatar block numbers; we read avatar 0.
//
// The label and directory structures below were confirmed byte-for-byte against
// the Need for Speed (3DO) disc: label at block 0 with block size 0x800 (2048),
// root directory at block 0x3551F, "last copy index" fields (a copy count is
// stored as its highest index, so the number of copies is the field + 1).

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

const (
	rawSectorSize = 2352 // a full raw CD sector
	userSize      = 2048 // Mode-1 / Mode-2-Form-1 user payload, and the Opera block size
	labelBlock    = 0    // the volume label sits at logical block 0

	// A directory entry's first word selects its kind in the low byte.
	entryTypeMask = 0x000000FF
	typeDir       = 0x07 // 0x02 = file, 0x06 = special file, 0x07 = directory
)

// Volume is an opened 3DO Opera disc image.
type Volume struct {
	img     []byte
	stride  int // bytes per sector in the image (2352 raw, 2048 cooked)
	dataOff int // offset of the 2048-byte user area within a sector
	nsect   int // number of sectors in the image

	Comment   string // label "comment" field
	Label     string // volume label, e.g. "CD-ROM"
	ID        uint32 // volume identifier
	blockSize int    // Opera block size in bytes (2048 on CD)

	rootBlock  int // first avatar of the root directory
	rootBlocks int // root directory length in blocks
}

// Entry is one directory entry — a file or a subdirectory. Block is the entry's
// first avatar (the block its data starts at); Copies lists every avatar.
type Entry struct {
	Name   string
	Path   string // full slash path from the volume root
	IsDir  bool
	Type   string // 4-char Opera type tag: "*dir", "*lbl", or a file type
	Size   int    // length in bytes (directory extent bytes for dirs)
	Blocks int    // length in blocks
	Block  int    // first avatar block
	Copies []int  // all avatar block numbers
}

func (e Entry) String() string {
	if e.IsDir {
		return fmt.Sprintf("%-40s  <dir>   (blk %d, %d copies)", e.Path, e.Block, len(e.Copies))
	}
	return fmt.Sprintf("%-40s  %9d  (blk %d, %d copies)", e.Path, e.Size, e.Block, len(e.Copies))
}

// be32 reads a big-endian 32-bit word.
func be32(b []byte) uint32 { return binary.BigEndian.Uint32(b[:4]) }

// Open validates the image geometry and the Opera volume label and returns a
// Volume ready for listing and extraction.
func Open(image []byte) (*Volume, error) {
	v := &Volume{img: image}
	switch {
	case len(image) >= rawSectorSize && len(image)%rawSectorSize == 0:
		v.stride = rawSectorSize
		v.nsect = len(image) / rawSectorSize
		switch image[0x0F] { // CD sector mode byte
		case 1:
			v.dataOff = 0x10
		case 2:
			v.dataOff = 0x18
		default:
			return nil, fmt.Errorf("threedo: unknown CD sector mode %d", image[0x0F])
		}
	case len(image) >= userSize && len(image)%userSize == 0:
		v.stride = userSize
		v.dataOff = 0
		v.nsect = len(image) / userSize
	default:
		return nil, fmt.Errorf("threedo: not a CD image (%d bytes)", len(image))
	}

	lbl, err := v.block(labelBlock)
	if err != nil {
		return nil, fmt.Errorf("threedo: reading volume label: %w", err)
	}
	// Volume label: record type 0x01, five 0x5A sync bytes, version 0x01.
	if lbl[0] != 0x01 || string(lbl[1:6]) != "\x5a\x5a\x5a\x5a\x5a" {
		return nil, errors.New("threedo: no Opera volume label at block 0")
	}
	v.Comment = trimName(lbl[8:40])
	v.Label = trimName(lbl[40:72])
	v.ID = be32(lbl[72:])
	v.blockSize = int(be32(lbl[76:]))
	if v.blockSize != userSize {
		// On CD the Opera block size is the 2048-byte sector payload; anything
		// else would break the 1:1 block-to-sector mapping block() assumes.
		return nil, fmt.Errorf("threedo: unexpected Opera block size %d (want %d)", v.blockSize, userSize)
	}
	// root_dir_id @0x54, root_dir_blocks @0x58, root_dir_block_size @0x5C,
	// last_root_copy @0x60, root_dir_copies[] @0x64.
	v.rootBlocks = int(be32(lbl[0x58:]))
	v.rootBlock = int(be32(lbl[0x64:])) // first root-directory avatar
	return v, nil
}

// block returns the 2048-byte user data of logical block (sector) n. Opera block
// numbers index CD sectors directly because the block size equals the payload.
func (v *Volume) block(n int) ([]byte, error) {
	if n < 0 || n >= v.nsect {
		return nil, fmt.Errorf("threedo: block %d out of range (0..%d)", n, v.nsect-1)
	}
	off := n*v.stride + v.dataOff
	return v.img[off : off+userSize], nil
}

// trimName trims an Opera fixed-width, NUL-padded name field.
func trimName(b []byte) string {
	if i := indexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return strings.TrimRight(string(b), " ")
}

func indexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

// dirEntries parses every entry of the directory whose first avatar is at block
// startBlock. A directory occupies one or more consecutive blocks; each block's
// header carries a "next" index — relative to startBlock, not an absolute block
// number — chaining to the following block until it is -1. Within a block, the
// header's endOffset marks the first free byte, bounding the variable-length
// entries (entries do not use per-entry terminator flags on this disc).
func (v *Volume) dirEntries(startBlock int, dirPath string) ([]Entry, error) {
	var out []Entry
	visited := map[int]bool{}
	for rel := 0; rel >= 0; {
		if visited[rel] {
			return nil, fmt.Errorf("threedo: directory block cycle at rel %d", rel)
		}
		visited[rel] = true

		b, err := v.block(startBlock + rel)
		if err != nil {
			return nil, err
		}
		// Directory block header: next(s32), prev(s32), flags, endOffset, firstEntry.
		next := int(int32(be32(b[0:])))
		endOffset := int(be32(b[12:]))
		firstEntry := int(be32(b[16:]))
		if endOffset > userSize {
			endOffset = userSize
		}

		for p := firstEntry; p+0x44 <= endOffset; {
			flags := be32(b[p:])
			ncopies := int(be32(b[p+0x40:])) + 1 // stored as the highest avatar index
			e := Entry{
				Name:   trimName(b[p+0x20 : p+0x40]),
				Type:   trimName(b[p+0x08 : p+0x0C]),
				Size:   int(be32(b[p+0x10:])),
				Blocks: int(be32(b[p+0x14:])),
				IsDir:  flags&entryTypeMask == typeDir,
			}
			for i := 0; i < ncopies && p+0x44+(i+1)*4 <= endOffset; i++ {
				e.Copies = append(e.Copies, int(be32(b[p+0x44+i*4:])))
			}
			if len(e.Copies) > 0 {
				e.Block = e.Copies[0]
			}
			if dirPath == "" {
				e.Path = e.Name
			} else {
				e.Path = dirPath + "/" + e.Name
			}
			out = append(out, e)
			p += 0x44 + ncopies*4
		}
		rel = next
	}
	return out, nil
}

// resolve walks a slash path and returns the matching entry. The empty path is
// the root directory, returned as a synthetic entry.
func (v *Volume) resolve(path string) (Entry, error) {
	cur := Entry{Name: v.Label, IsDir: true, Block: v.rootBlock, Blocks: v.rootBlocks, Copies: []int{v.rootBlock}}
	for _, want := range splitPath(path) {
		if !cur.IsDir {
			return Entry{}, fmt.Errorf("threedo: %q is not a directory", cur.Path)
		}
		entries, err := v.dirEntries(cur.Block, cur.Path)
		if err != nil {
			return Entry{}, err
		}
		found := false
		for _, e := range entries {
			if strings.EqualFold(e.Name, want) {
				cur, found = e, true
				break
			}
		}
		if !found {
			return Entry{}, fmt.Errorf("threedo: %q not found", path)
		}
	}
	return cur, nil
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

// ReadDir lists the entries of the directory at path ("" or "/" is the root).
func (v *Volume) ReadDir(path string) ([]Entry, error) {
	e, err := v.resolve(path)
	if err != nil {
		return nil, err
	}
	if !e.IsDir {
		return nil, fmt.Errorf("threedo: %q is not a directory", path)
	}
	return v.dirEntries(e.Block, e.Path)
}

// Walk visits every file and directory under the root, depth first.
func (v *Volume) Walk(fn func(Entry) error) error {
	return v.walk(v.rootBlock, "", fn)
}

func (v *Volume) walk(block int, dirPath string, fn func(Entry) error) error {
	entries, err := v.dirEntries(block, dirPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := fn(e); err != nil {
			return err
		}
		if e.IsDir {
			if err := v.walk(e.Block, e.Path, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReadFile returns the full contents of the file at path, read from avatar 0.
// An avatar is a contiguous run of the file's blocks starting at its block
// number; the byte length truncates the last block.
func (v *Volume) ReadFile(path string) ([]byte, error) {
	e, err := v.resolve(path)
	if err != nil {
		return nil, err
	}
	if e.IsDir {
		return nil, fmt.Errorf("threedo: %q is a directory", path)
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
