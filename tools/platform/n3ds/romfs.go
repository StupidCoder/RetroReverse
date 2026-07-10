package n3ds

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"unicode/utf16"
)

// The RomFS region is wrapped in an IVFC hash tree: three levels of SHA-256
// hashes over fixed-size blocks, with level 3 holding the actual filesystem.
// A master hash in the region header hashes level 1, level 1 hashes level 2,
// and level 2 hashes level 3 — so the console can authenticate any one block of
// the filesystem by checking a short chain instead of the whole 300 MB image.
//
// The three levels' LogicalOffset fields describe a *logical* space in which the
// levels sit end to end in numerical order (1, 2, 3). That is not how they are
// stored. On media the data level comes first:
//
//	0x00000000  IVFC header (0x60 bytes) + master hash (MasterHashSize)
//	0x00001000  level 3 — the filesystem
//	  align     level 1 — hashes over level 2
//	  align     level 2 — hashes over level 3
//	            end == the RomFS region size, exactly
//
// Reading the LogicalOffset fields as media offsets lands in the middle of the
// level-2 hash block, where the level-3 header check below fails loudly rather
// than silently decoding hashes as a directory table. Both the closing of the
// size arithmetic and the hash chains themselves confirm this order; VerifyIVFC
// checks the latter.
//
// Level 3 itself is a header of nine table offsets followed by two hash tables,
// a directory metadata table, a file metadata table, and the file data. The
// metadata tables are linked structures: each directory names its parent, its
// next sibling, its first child directory and its first file; each file names
// its parent and its next sibling. Names are UTF-16LE, padded to a 4-byte
// boundary. A 0xFFFFFFFF link is "none".
const (
	ivfcMagic       = "IVFC"
	ivfcMagicNumber = 0x00010000
	ivfcHeaderSize  = 0x60
	romfsHeaderLen  = 0x28
	romfsNoEntry    = 0xFFFFFFFF
)

// ivfcLevel is one entry of the IVFC level table.
type ivfcLevel struct {
	LogicalOffset uint64
	HashDataSize  uint64
	BlockSizeLog  uint32
}

func (l ivfcLevel) blockSize() uint64 { return 1 << l.BlockSizeLog }

// RomFSFile is one file in the RomFS.
type RomFSFile struct {
	Path   string // '/'-separated, rooted (e.g. "/Layout/Menu.arc")
	Offset int64  // byte offset of the file's data within the RomFS region
	Size   int64
}

// RomFS is a parsed, flattened RomFS.
type RomFS struct {
	raw       []byte
	dataStart int64 // absolute offset of the file-data area within raw

	MasterHash []byte      // hashes level 1
	Levels     [3]ivfcSpan // media offsets and sizes of levels 1, 2 and 3

	Files []RomFSFile // sorted by path
	Dirs  []string    // sorted, rooted directory paths (excluding "/")
}

// ivfcSpan is one IVFC level as actually stored on media.
type ivfcSpan struct {
	Offset    int64
	Size      int64
	BlockSize int64
}

func align64(v, a uint64) uint64 {
	if a == 0 {
		return v
	}
	return (v + a - 1) &^ (a - 1)
}

// ParseRomFS parses the IVFC wrapper and the level-3 filesystem inside it.
func ParseRomFS(b []byte) (*RomFS, error) {
	if len(b) < ivfcHeaderSize {
		return nil, fmt.Errorf("romfs: region too short for an IVFC header (%d bytes)", len(b))
	}
	if string(b[0:4]) != ivfcMagic {
		return nil, fmt.Errorf("romfs: bad IVFC magic %q", b[0:4])
	}
	if m := binary.LittleEndian.Uint32(b[4:]); m != ivfcMagicNumber {
		return nil, fmt.Errorf("romfs: bad IVFC magic number 0x%08x, want 0x%08x", m, ivfcMagicNumber)
	}
	masterHashSize := uint64(binary.LittleEndian.Uint32(b[8:]))

	var lvl [3]ivfcLevel
	for i := range lvl {
		o := 0x0C + i*0x18
		lvl[i] = ivfcLevel{
			LogicalOffset: binary.LittleEndian.Uint64(b[o:]),
			HashDataSize:  binary.LittleEndian.Uint64(b[o+8:]),
			BlockSizeLog:  binary.LittleEndian.Uint32(b[o+16:]),
		}
	}

	// Media order: the filesystem (level 3) first, then the two hash levels,
	// each aligned to its own block size.
	l3Off := align64(ivfcHeaderSize+masterHashSize, lvl[2].blockSize())
	l1Off := align64(l3Off+lvl[2].HashDataSize, lvl[0].blockSize())
	l2Off := align64(l1Off+lvl[0].HashDataSize, lvl[1].blockSize())
	end := align64(l2Off+lvl[1].HashDataSize, lvl[1].blockSize())

	// The levels must tile the region exactly. A mismatch means the header was
	// misread, so refuse rather than index into the wrong place.
	if end != uint64(len(b)) {
		return nil, fmt.Errorf("romfs: IVFC levels end at 0x%x but the region is 0x%x bytes", end, len(b))
	}

	off := l3Off
	if off+romfsHeaderLen > uint64(len(b)) {
		return nil, fmt.Errorf("romfs: level-3 offset 0x%x runs past the region end (0x%x)", off, len(b))
	}
	l3 := b[off:]

	// The level-3 header's own length field is the check that the walk above
	// landed on the filesystem and not in the middle of a hash block.
	if hl := binary.LittleEndian.Uint32(l3); hl != romfsHeaderLen {
		return nil, fmt.Errorf("romfs: level-3 header length 0x%x at offset 0x%x, want 0x%x — IVFC level walk mislanded", hl, off, romfsHeaderLen)
	}

	u32 := func(o int) uint32 { return binary.LittleEndian.Uint32(l3[o:]) }
	dirMetaOff, dirMetaLen := u32(0x0C), u32(0x10)
	fileMetaOff, fileMetaLen := u32(0x1C), u32(0x20)
	fileDataOff := u32(0x24)

	for _, r := range []struct {
		name     string
		off, len uint32
	}{{"dir meta", dirMetaOff, dirMetaLen}, {"file meta", fileMetaOff, fileMetaLen}} {
		if uint64(off)+uint64(r.off)+uint64(r.len) > uint64(len(b)) {
			return nil, fmt.Errorf("romfs: %s table [0x%x+0x%x] runs past the region end", r.name, r.off, r.len)
		}
	}

	fs := &RomFS{
		raw:        b,
		dataStart:  int64(off) + int64(fileDataOff),
		MasterHash: b[ivfcHeaderSize : ivfcHeaderSize+masterHashSize],
		Levels: [3]ivfcSpan{
			{Offset: int64(l1Off), Size: int64(lvl[0].HashDataSize), BlockSize: int64(lvl[0].blockSize())},
			{Offset: int64(l2Off), Size: int64(lvl[1].HashDataSize), BlockSize: int64(lvl[1].blockSize())},
			{Offset: int64(l3Off), Size: int64(lvl[2].HashDataSize), BlockSize: int64(lvl[2].blockSize())},
		},
	}
	dirMeta := l3[dirMetaOff : dirMetaOff+dirMetaLen]
	fileMeta := l3[fileMetaOff : fileMetaOff+fileMetaLen]

	if err := fs.walk(dirMeta, fileMeta, 0, ""); err != nil {
		return nil, err
	}
	sort.Slice(fs.Files, func(i, j int) bool { return fs.Files[i].Path < fs.Files[j].Path })
	sort.Strings(fs.Dirs)
	return fs, nil
}

// walk descends the directory metadata table depth-first from the entry at
// dirOff, whose full path is prefix.
func (fs *RomFS) walk(dirMeta, fileMeta []byte, dirOff uint32, prefix string) error {
	if int(dirOff)+0x18 > len(dirMeta) {
		return fmt.Errorf("romfs: directory entry at 0x%x runs past the metadata table", dirOff)
	}
	d := dirMeta[dirOff:]
	u32 := func(o int) uint32 { return binary.LittleEndian.Uint32(d[o:]) }
	firstChild, firstFile := u32(0x08), u32(0x0C)

	// Files of this directory.
	for fo := firstFile; fo != romfsNoEntry; {
		if int(fo)+0x20 > len(fileMeta) {
			return fmt.Errorf("romfs: file entry at 0x%x runs past the metadata table", fo)
		}
		f := fileMeta[fo:]
		dataOff := int64(binary.LittleEndian.Uint64(f[0x08:]))
		dataLen := int64(binary.LittleEndian.Uint64(f[0x10:]))
		nameLen := binary.LittleEndian.Uint32(f[0x1C:])
		name, err := utf16Name(f, 0x20, nameLen, len(fileMeta)-int(fo))
		if err != nil {
			return fmt.Errorf("romfs: file entry at 0x%x: %w", fo, err)
		}
		if fs.dataStart+dataOff+dataLen > int64(len(fs.raw)) {
			return fmt.Errorf("romfs: file %q data [0x%x+0x%x] runs past the region end", prefix+"/"+name, dataOff, dataLen)
		}
		fs.Files = append(fs.Files, RomFSFile{Path: prefix + "/" + name, Offset: dataOff, Size: dataLen})
		fo = binary.LittleEndian.Uint32(f[0x04:]) // sibling
	}

	// Child directories.
	for co := firstChild; co != romfsNoEntry; {
		if int(co)+0x18 > len(dirMeta) {
			return fmt.Errorf("romfs: directory entry at 0x%x runs past the metadata table", co)
		}
		c := dirMeta[co:]
		nameLen := binary.LittleEndian.Uint32(c[0x14:])
		name, err := utf16Name(c, 0x18, nameLen, len(dirMeta)-int(co))
		if err != nil {
			return fmt.Errorf("romfs: directory entry at 0x%x: %w", co, err)
		}
		path := prefix + "/" + name
		fs.Dirs = append(fs.Dirs, path)
		if err := fs.walk(dirMeta, fileMeta, co, path); err != nil {
			return err
		}
		co = binary.LittleEndian.Uint32(c[0x04:]) // sibling
	}
	return nil
}

// utf16Name decodes a UTF-16LE name of nameLen *bytes* stored at off within e.
func utf16Name(e []byte, off int, nameLen uint32, avail int) (string, error) {
	if nameLen%2 != 0 {
		return "", fmt.Errorf("odd name length %d", nameLen)
	}
	if off+int(nameLen) > avail {
		return "", fmt.Errorf("name of %d bytes runs past the entry", nameLen)
	}
	u := make([]uint16, nameLen/2)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(e[off+i*2:])
	}
	name := string(utf16.Decode(u))
	if name == "" || strings.ContainsAny(name, "/\x00") {
		return "", fmt.Errorf("invalid name %q", name)
	}
	return name, nil
}

// VerifyIVFC walks the whole hash tree: the master hash over level 1, level 1
// over level 2, and level 2 over level 3. It is the round-trip proof that the
// level offsets derived above are the real ones — a wrong offset cannot produce
// matching SHA-256 chains. It reads the entire region, so it is not on the
// parse path; call it once when validating a new image.
func (fs *RomFS) VerifyIVFC() error {
	// hashes: the buffer of 32-byte digests. data: the level they cover.
	check := func(name string, hashes []byte, d ivfcSpan) error {
		blocks := (d.Size + d.BlockSize - 1) / d.BlockSize
		if have := int64(len(hashes)) / sha256.Size; have < blocks {
			return fmt.Errorf("romfs: %s: %d blocks but only %d hashes", name, blocks, have)
		}
		blk := make([]byte, d.BlockSize)
		for i := int64(0); i < blocks; i++ {
			off := d.Offset + i*d.BlockSize
			for j := range blk {
				blk[j] = 0
			}
			copy(blk, fs.raw[off:min64(off+d.BlockSize, int64(len(fs.raw)))])
			h := sha256.Sum256(blk)
			if !bytes.Equal(h[:], hashes[i*sha256.Size:(i+1)*sha256.Size]) {
				return fmt.Errorf("romfs: %s: block %d at 0x%x fails its hash", name, i, off)
			}
		}
		return nil
	}

	l1, l2, l3 := fs.Levels[0], fs.Levels[1], fs.Levels[2]
	if err := check("level 1 under the master hash", fs.MasterHash, l1); err != nil {
		return err
	}
	if err := check("level 2 under level 1", fs.raw[l1.Offset:l1.Offset+l1.Size], l2); err != nil {
		return err
	}
	return check("level 3 under level 2", fs.raw[l2.Offset:l2.Offset+l2.Size], l3)
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// File returns the bytes of the file at the given rooted path. Files is sorted
// by path, so the lookup is a binary search.
func (fs *RomFS) File(path string) ([]byte, error) {
	i := sort.Search(len(fs.Files), func(i int) bool { return fs.Files[i].Path >= path })
	if i == len(fs.Files) || fs.Files[i].Path != path {
		return nil, fmt.Errorf("romfs: no file %q", path)
	}
	return fs.Data(fs.Files[i]), nil
}

// Data returns the bytes of a file entry.
func (fs *RomFS) Data(f RomFSFile) []byte {
	start := fs.dataStart + f.Offset
	return fs.raw[start : start+f.Size]
}
