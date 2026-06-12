// Package adf reads files from a standard AmigaDOS floppy disk image (ADF).
//
// An ADF is a flat dump of 512-byte logical blocks (sectors). A double-density
// floppy is 1760 blocks (901,120 bytes); high density is 3520. The on-disk
// filesystem is AmigaDOS OFS or FFS, distinguished by the boot block:
//
//	block 0   "DOS" + flags    flags bit0 = FFS, bit1 = international, bit2 = dircache
//	block 8.. boot code
//	block 880 root block       (numBlocks/2): the top directory + volume name
//
// Directories (the root and any subdirectory) store their entries in a 72-slot
// hash table; entries that hash to the same slot are chained through each
// block's hash-chain field. A file header block lists its data blocks in a
// table filled from the end; large files continue into "file extension" blocks.
//
// Under OFS each 512-byte data block carries a 24-byte header (sequence number,
// valid-byte count, next-block link) and 488 bytes of payload; under FFS a data
// block is 512 raw payload bytes and the header's table is the only index.
//
// This package resolves paths by walking a directory's full contents rather
// than by recomputing AmigaDOS's locale-dependent name hash, so it does not
// depend on the volume's international/dircache flags.
package adf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Block size and the AmigaDOS block/secondary types this package recognises.
const (
	bsize = 512

	tHeader = 2  // primary type of root, directory and file-header blocks
	tData   = 8  // OFS data block
	tList   = 16 // file extension block

	stRoot    = 1  // secondary type: root block
	stUserDir = 2  // secondary type: subdirectory
	stFile    = -3 // secondary type: file header
)

// Field byte offsets within a 512-byte block. The trailing fields are counted
// back from the end of the block, which is how AmigaDOS specifies them.
const (
	offType     = 0x000 // primary type
	offHighSeq  = 0x008 // file: number of data-block pointers used here
	offChecksum = 0x014 // block checksum
	offTable    = 0x018 // hash table (dir) / data-block table (file), 72 longs

	offProtect  = bsize - 192 // 0x140 file protection bits
	offByteSize = bsize - 188 // 0x144 file size in bytes
	offName     = bsize - 80  // 0x1B0 BCPL name: length byte then chars
	offHashNext = bsize - 16  // 0x1F0 next entry in the same hash chain
	offExtend   = bsize - 8   // 0x1F8 file: next extension block
	offSecType  = bsize - 4   // 0x1FC secondary type

	tableSlots = bsize/4 - 56 // 72 hash-table / data-block-table entries
)

// Volume is an opened ADF image.
type Volume struct {
	img      []byte
	nblocks  int
	rootBlk  int
	FFS      bool   // Fast File System (raw 512-byte data blocks) vs OFS
	Intl     bool   // international-mode name hashing flag (from the boot block)
	Dircache bool   // directory-cache flag (from the boot block)
	Name     string // volume (disk) name
}

// Entry is one directory entry — a file or a subdirectory.
type Entry struct {
	Name    string // entry name
	Path    string // full slash-separated path from the volume root
	IsDir   bool
	Size    int    // file size in bytes (0 for directories)
	Block   int    // header block number
	Protect uint32 // AmigaDOS protection bits (files)
}

func (e Entry) String() string {
	if e.IsDir {
		return fmt.Sprintf("%-40s  <dir>", e.Path)
	}
	return fmt.Sprintf("%-40s  %7d  (block %d)", e.Path, e.Size, e.Block)
}

// Open validates the boot block, reads the filesystem flags and the volume
// name, and returns a Volume ready for listing and extraction.
func Open(image []byte) (*Volume, error) {
	if len(image) < 1024 || len(image)%bsize != 0 {
		return nil, fmt.Errorf("adf: not a block image (%d bytes)", len(image))
	}
	if string(image[0:3]) != "DOS" {
		return nil, fmt.Errorf("adf: missing DOS signature (have %q)", image[0:4])
	}
	flags := image[3]
	v := &Volume{
		img:      image,
		nblocks:  len(image) / bsize,
		FFS:      flags&1 != 0,
		Intl:     flags&2 != 0,
		Dircache: flags&4 != 0,
	}
	v.rootBlk = v.nblocks / 2 // 880 on a DD floppy, 1760 on HD
	root, err := v.block(v.rootBlk)
	if err != nil {
		return nil, fmt.Errorf("adf: root block: %w", err)
	}
	if int32(be(root, offType)) != tHeader || int32(be(root, offSecType)) != stRoot {
		return nil, errors.New("adf: root block is not a valid root header")
	}
	v.Name = bcplName(root)
	return v, nil
}

// be reads the big-endian 32-bit long at byte offset off in block b.
func be(b []byte, off int) uint32 { return binary.BigEndian.Uint32(b[off : off+4]) }

// block returns the n-th 512-byte block, bounds-checked.
func (v *Volume) block(n int) ([]byte, error) {
	if n < 0 || n >= v.nblocks {
		return nil, fmt.Errorf("block %d out of range (0..%d)", n, v.nblocks-1)
	}
	return v.img[n*bsize : (n+1)*bsize], nil
}

// bcplName reads the BCPL-style name (length byte at offName, then chars).
func bcplName(b []byte) string {
	n := int(b[offName])
	if n > 30 {
		n = 30
	}
	return string(b[offName+1 : offName+1+n])
}

// ChecksumOK reports whether block n carries a valid AmigaDOS block checksum
// (the 128 big-endian longs, including the checksum field, sum to zero).
func (v *Volume) ChecksumOK(n int) bool {
	b, err := v.block(n)
	if err != nil {
		return false
	}
	var sum uint32
	for i := 0; i < bsize; i += 4 {
		sum += be(b, i)
	}
	return sum == 0
}

// entriesIn lists every entry of the directory (or root) header at dirBlock by
// visiting all 72 hash slots and following each slot's hash chain.
func (v *Volume) entriesIn(dirBlock int, dirPath string) ([]Entry, error) {
	dir, err := v.block(dirBlock)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for slot := 0; slot < tableSlots; slot++ {
		blk := int(be(dir, offTable+slot*4))
		for blk != 0 {
			b, err := v.block(blk)
			if err != nil {
				return nil, err
			}
			name := bcplName(b)
			sec := int32(be(b, offSecType))
			e := Entry{Name: name, Block: blk, IsDir: sec == stUserDir}
			if dirPath == "" {
				e.Path = name
			} else {
				e.Path = dirPath + "/" + name
			}
			if sec == stFile {
				e.Size = int(be(b, offByteSize))
				e.Protect = be(b, offProtect)
			}
			if sec == stUserDir || sec == stFile {
				out = append(out, e)
			}
			blk = int(be(b, offHashNext)) // next entry in this hash chain
		}
	}
	return out, nil
}

// splitPath cleans and splits a slash path into components (empty for root).
func splitPath(p string) []string {
	var parts []string
	for _, c := range strings.Split(strings.Trim(p, "/"), "/") {
		if c != "" {
			parts = append(parts, c)
		}
	}
	return parts
}

// resolve walks the path and returns the matching entry. The root itself is
// returned as a synthetic directory entry for the empty path.
func (v *Volume) resolve(path string) (Entry, error) {
	parts := splitPath(path)
	cur := Entry{Name: v.Name, Path: "", IsDir: true, Block: v.rootBlk}
	curPath := ""
	for _, want := range parts {
		if !cur.IsDir {
			return Entry{}, fmt.Errorf("adf: %q is not a directory", curPath)
		}
		entries, err := v.entriesIn(cur.Block, curPath)
		if err != nil {
			return Entry{}, err
		}
		found := false
		for _, e := range entries {
			if strings.EqualFold(e.Name, want) { // AmigaDOS is case-insensitive
				cur, found = e, true
				break
			}
		}
		if !found {
			return Entry{}, fmt.Errorf("adf: %q not found", path)
		}
		curPath = cur.Path
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
		return nil, fmt.Errorf("adf: %q is not a directory", path)
	}
	return v.entriesIn(e.Block, e.Path)
}

// Walk visits every file and directory under the root, depth first. Returning
// an error from fn stops the walk and propagates the error.
func (v *Volume) Walk(fn func(Entry) error) error {
	return v.walk(v.rootBlk, "", fn)
}

func (v *Volume) walk(dirBlock int, dirPath string, fn func(Entry) error) error {
	entries, err := v.entriesIn(dirBlock, dirPath)
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

// dataBlocks returns the ordered list of data-block numbers for a file header,
// following the in-header table and any file-extension blocks.
func (v *Volume) dataBlocks(header []byte) ([]int, error) {
	var blocks []int
	cur := header
	for {
		seq := int(be(cur, offHighSeq))
		if seq > tableSlots {
			seq = tableSlots
		}
		// The table is filled from the end: entry i sits at slot (tableSlots-1-i).
		for i := 0; i < seq; i++ {
			ptr := int(be(cur, offTable+(tableSlots-1-i)*4))
			if ptr == 0 {
				continue
			}
			blocks = append(blocks, ptr)
		}
		ext := int(be(cur, offExtend))
		if ext == 0 {
			break
		}
		b, err := v.block(ext)
		if err != nil {
			return nil, err
		}
		// A file-extension (file-list) block has primary type T_LIST and
		// secondary type ST_FILE; its data-block table has the same layout.
		if int32(be(b, offType)) != tList || int32(be(b, offSecType)) != stFile {
			return nil, fmt.Errorf("adf: bad file-extension block %d", ext)
		}
		cur = b
	}
	return blocks, nil
}

// ReadFile returns the full contents of the file at path.
func (v *Volume) ReadFile(path string) ([]byte, error) {
	e, err := v.resolve(path)
	if err != nil {
		return nil, err
	}
	if e.IsDir {
		return nil, fmt.Errorf("adf: %q is a directory", path)
	}
	return v.readFile(e.Block)
}

func (v *Volume) readFile(headerBlock int) ([]byte, error) {
	header, err := v.block(headerBlock)
	if err != nil {
		return nil, err
	}
	size := int(be(header, offByteSize))
	blocks, err := v.dataBlocks(header)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, size)
	for _, n := range blocks {
		b, err := v.block(n)
		if err != nil {
			return nil, err
		}
		if v.FFS {
			out = append(out, b...) // raw 512-byte payload
		} else {
			// OFS: 24-byte header, then up to 488 valid bytes (offset 0x0C).
			valid := int(be(b, 0x0C))
			if valid > bsize-24 {
				valid = bsize - 24
			}
			out = append(out, b[24:24+valid]...)
		}
	}
	if len(out) > size {
		out = out[:size] // trim padding in the final block
	}
	return out, nil
}
