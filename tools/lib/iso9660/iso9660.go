// Package iso9660 reads an ISO 9660 filesystem out of a disc image.
//
// The filesystem itself is simple: a Primary Volume Descriptor at logical block 16
// and a tree of directory records. Multi-byte numeric fields are stored
// "both-endian" (little then big); the little-endian half is read. Records are
// packed within a logical block and never span one.
//
// The part that is not simple is where those 2048-byte logical blocks live in the
// file. Dumps differ: a "cooked" image stores the 2048 user bytes back to back, a
// raw CD dump wraps each in the 2352-byte sector framing, and some dumpers pad the
// sector out further still. Geometry names that layout, and Detect finds it by
// looking for the volume descriptor rather than trusting the file extension — the
// alternative is a reader that silently sees an empty disc.
//
// The reader sits on a BlockSource, so an image is never required to be resident:
// a 1.7 GB DVD is served from an *os.File through NewSource.
package iso9660

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	// BlockSize is the ISO 9660 logical block size. It is 2048 on every disc this
	// repository reads, and the PVD is checked to confirm it.
	BlockSize = 2048

	// pvdLBA is the logical block holding the Primary Volume Descriptor.
	pvdLBA = 16
)

// Geometry describes how 2048-byte logical blocks are laid out inside an image
// file: each occupies SectorSize bytes, and the user data begins DataOffset bytes
// into that sector.
type Geometry struct {
	SectorSize int
	DataOffset int
}

// The layouts Detect probes, in order. The first that yields a valid Primary
// Volume Descriptor wins.
var geometries = []Geometry{
	{2048, 0},  // cooked: user data only (most .iso dumps)
	{2352, 16}, // raw CD, Mode 1: 16-byte sync+header, 2048 data, 288 EDC/ECC
	{2352, 24}, // raw CD, Mode 2 Form 1: +8-byte subheader
	{2448, 16}, // raw CD Mode 1 + 96-byte subchannel
	{2448, 24}, // raw CD Mode 2 Form 1 + 96-byte subchannel
	{2448, 0},  // 2048 user bytes + 400 bytes of padding
	{2336, 8},  // Mode 2 Form 1, sync/header stripped
	{2336, 0},
}

func (g Geometry) String() string {
	return fmt.Sprintf("%d-byte sectors, data at +%d", g.SectorSize, g.DataOffset)
}

// BlockSource yields 2048-byte logical blocks by LBA.
type BlockSource interface {
	ReadBlock(n int) ([]byte, error)
}

// Source serves logical blocks from an io.ReaderAt under a known Geometry.
type Source struct {
	ra   io.ReaderAt
	size int64
	geom Geometry
	buf  []byte
}

// NewSource wraps ra with the given geometry.
func NewSource(ra io.ReaderAt, size int64, g Geometry) *Source {
	return &Source{ra: ra, size: size, geom: g, buf: make([]byte, BlockSize)}
}

// Geometry reports the sector layout this source reads through.
func (s *Source) Geometry() Geometry { return s.geom }

// ReadBlock returns logical block n. The slice aliases an internal buffer and is
// only valid until the next call — callers that retain it must copy.
func (s *Source) ReadBlock(n int) ([]byte, error) {
	off := int64(n)*int64(s.geom.SectorSize) + int64(s.geom.DataOffset)
	if n < 0 || off+BlockSize > s.size {
		return nil, fmt.Errorf("iso9660: block %d out of range", n)
	}
	if _, err := s.ra.ReadAt(s.buf, off); err != nil {
		return nil, fmt.Errorf("iso9660: reading block %d: %w", n, err)
	}
	return s.buf, nil
}

// Detect finds the sector layout of an image by locating the Primary Volume
// Descriptor. It reports an error if no candidate geometry yields one.
func Detect(ra io.ReaderAt, size int64) (Geometry, error) {
	for _, g := range geometries {
		if validPVD(ra, size, g) {
			return g, nil
		}
	}
	return Geometry{}, errors.New("iso9660: no ISO 9660 volume descriptor found in any known sector layout")
}

// validPVD reports whether reading LBA 16 under g yields a Primary Volume
// Descriptor whose declared extent is consistent with the file's actual length.
// The length check is what stops a geometry from matching by luck: several layouts
// can place "CD001" plausibly, but only the right one describes a volume that fits.
func validPVD(ra io.ReaderAt, size int64, g Geometry) bool {
	off := int64(pvdLBA)*int64(g.SectorSize) + int64(g.DataOffset)
	if off+BlockSize > size {
		return false
	}
	b := make([]byte, BlockSize)
	if _, err := ra.ReadAt(b, off); err != nil {
		return false
	}
	if b[0] != 0x01 || string(b[1:6]) != "CD001" || b[6] != 0x01 {
		return false
	}
	if int(binary.LittleEndian.Uint16(b[128:130])) != BlockSize {
		return false // a logical block size we do not handle
	}
	// The volume claims this many logical blocks; they must fit in the file under g.
	blocks := int64(le32(b[80:]))
	if blocks <= 0 || blocks*int64(g.SectorSize) > size {
		return false
	}
	return true
}

// Volume is an opened ISO 9660 filesystem.
type Volume struct {
	src BlockSource

	System   string // PVD system identifier, e.g. "PLAYSTATION"
	Name     string // PVD volume identifier
	Blocks   int    // PVD volume space size, in logical blocks
	rootLBA  int
	rootSize int
}

// Geometry reports the sector layout the volume reads through, when its source
// knows one. A cooked .iso and a raw CD dump both serve the same 2048-byte blocks,
// but which container they came out of is a fact about the *disc* — a raw 2352-byte
// sector layout only ever comes off a CD — and an emulated drive needs it to answer
// "what kind of disc is this" honestly.
func (v *Volume) Geometry() (Geometry, bool) {
	if s, ok := v.src.(*Source); ok {
		return s.Geometry(), true
	}
	return Geometry{}, false
}

// ReadBlock returns one 2048-byte logical block by LBA, whatever the image's geometry.
//
// It is the volume as a *disc* rather than as a filesystem, and it is what a machine that
// models the drive needs: a guest asks its CD/DVD controller for a sector number, not for a
// file, and the filesystem it is walking is its own business.
func (v *Volume) ReadBlock(n int) ([]byte, error) { return v.src.ReadBlock(n) }

// Entry is one directory entry — a file or a subdirectory.
type Entry struct {
	Name  string // ISO identifier, e.g. "SCUS_971.24;1" (directories carry no version)
	Path  string // full slash path from the volume root
	IsDir bool
	Size  int // file size in bytes (extent size for directories)
	Block int // extent LBA
}

func (e Entry) String() string {
	if e.IsDir {
		return fmt.Sprintf("%-32s  <dir>   (lba %d)", e.Path, e.Block)
	}
	return fmt.Sprintf("%-32s  %10d  (lba %d)", e.Path, e.Size, e.Block)
}

// Open detects the geometry of an image and opens its volume.
func Open(ra io.ReaderAt, size int64) (*Volume, error) {
	g, err := Detect(ra, size)
	if err != nil {
		return nil, err
	}
	return OpenVolume(NewSource(ra, size, g))
}

// OpenBytes opens a volume from an image held in memory.
func OpenBytes(image []byte) (*Volume, error) {
	return Open(newByteReaderAt(image), int64(len(image)))
}

// OpenVolume validates the Primary Volume Descriptor on src and returns a Volume
// ready for listing and extraction. Use it when the geometry is already known (a
// decompressing block source, say, that only ever yields cooked blocks).
func OpenVolume(src BlockSource) (*Volume, error) {
	pvd, err := src.ReadBlock(pvdLBA)
	if err != nil {
		return nil, fmt.Errorf("iso9660: reading PVD: %w", err)
	}
	if pvd[0] != 0x01 || string(pvd[1:6]) != "CD001" {
		return nil, errors.New("iso9660: no Primary Volume Descriptor at LBA 16")
	}
	v := &Volume{
		src:    src,
		System: strings.TrimRight(string(pvd[8:40]), " \x00"),
		Name:   strings.TrimRight(string(pvd[40:72]), " \x00"),
		Blocks: int(le32(pvd[80:])),
	}
	// The root directory record is 34 bytes at offset 156.
	root := pvd[156 : 156+34]
	v.rootLBA = int(le32(root[2:]))
	v.rootSize = int(le32(root[10:]))
	return v, nil
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
		n := BlockSize
		if remaining < n {
			n = remaining
		}
		remaining -= BlockSize
		for p := 0; p < n; {
			recLen := int(b[p])
			if recLen == 0 {
				break // padding to the end of this logical block
			}
			if p+recLen > BlockSize {
				return nil, fmt.Errorf("iso9660: directory record overruns block at lba %d", lba+sect)
			}
			rec := b[p : p+recLen]
			p += recLen

			idLen := int(rec[32])
			id := rec[33 : 33+idLen]
			if idLen == 1 && (id[0] == 0x00 || id[0] == 0x01) {
				continue // "." / ".."
			}
			e := Entry{
				Name:  string(id),
				IsDir: rec[25]&0x02 != 0,
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

// Resolve walks a slash path and returns the matching entry. The empty path is the
// root directory, returned as a synthetic entry. A drive prefix ("cdrom0:\...") and
// backslashes are accepted, so a path lifted straight out of SYSTEM.CNF resolves.
//
// The raw-extent syntax "sce_lbn0x<lbn>_size0x<size>" — a game opening a sector run
// directly, by the LBN it read from a file's stat — resolves to a synthetic entry
// over that extent.
func (v *Volume) Resolve(path string) (Entry, error) {
	path = normalisePath(path)
	if lbn, size, ok := parseLbnPath(path); ok {
		return Entry{Name: path, Path: path, Block: lbn, Size: size}, nil
	}
	cur := Entry{Name: v.Name, Path: "", IsDir: true, Block: v.rootLBA, Size: v.rootSize}
	for _, want := range splitPath(path) {
		if !cur.IsDir {
			return Entry{}, fmt.Errorf("iso9660: %q is not a directory", cur.Path)
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
			return Entry{}, fmt.Errorf("iso9660: %q not found", path)
		}
	}
	return cur, nil
}

// normalisePath strips a "device:" prefix and turns backslashes into slashes.
func normalisePath(p string) string {
	if i := strings.IndexByte(p, ':'); i >= 0 && !strings.ContainsAny(p[:i], "/\\") {
		p = p[i+1:]
	}
	return strings.ReplaceAll(p, "\\", "/")
}

// parseLbnPath recognises the raw-extent path "sce_lbn0x<lbn>_size0x<size>" (both
// numbers hex, case-insensitive).
func parseLbnPath(path string) (lbn, size int, ok bool) {
	p := strings.ToLower(strings.Trim(path, "/"))
	if !strings.HasPrefix(p, "sce_lbn") {
		return 0, 0, false
	}
	rest := p[len("sce_lbn"):]
	i := strings.Index(rest, "_size")
	if i < 0 {
		return 0, 0, false
	}
	l, err1 := strconv.ParseInt(strings.TrimPrefix(rest[:i], "0x"), 16, 64)
	s, err2 := strconv.ParseInt(strings.TrimPrefix(rest[i+len("_size"):], "0x"), 16, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return int(l), int(s), true
}

// ReadDir lists the entries of the directory at path ("" or "/" is the root).
func (v *Volume) ReadDir(path string) ([]Entry, error) {
	e, err := v.Resolve(path)
	if err != nil {
		return nil, err
	}
	if !e.IsDir {
		return nil, fmt.Errorf("iso9660: %q is not a directory", path)
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

// ReadFileAt reads len(p) bytes of entry e starting at byte offset off, returning
// the count read (short at EOF).
func (v *Volume) ReadFileAt(e Entry, off int, p []byte) (int, error) {
	if off >= e.Size {
		return 0, nil
	}
	if off+len(p) > e.Size {
		p = p[:e.Size-off]
	}
	got := 0
	for got < len(p) {
		b, err := v.src.ReadBlock(e.Block + (off+got)/BlockSize)
		if err != nil {
			return got, err
		}
		got += copy(p[got:], b[(off+got)%BlockSize:])
	}
	return got, nil
}

// ReadFile returns the full contents of the file at path.
func (v *Volume) ReadFile(path string) ([]byte, error) {
	e, err := v.Resolve(path)
	if err != nil {
		return nil, err
	}
	if e.IsDir {
		return nil, fmt.Errorf("iso9660: %q is a directory", path)
	}
	out := make([]byte, e.Size)
	if _, err := v.ReadFileAt(e, 0, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FileAt names the entry whose extent contains the given logical block, turning a
// raw sector read back into a filename. It is the disc-side counterpart of the
// oracle's "-at OFFSET" probe.
func (v *Volume) FileAt(lba int) (Entry, bool) {
	var hit Entry
	found := false
	v.Walk(func(e Entry) error {
		if e.IsDir {
			return nil
		}
		blocks := (e.Size + BlockSize - 1) / BlockSize
		if lba >= e.Block && lba < e.Block+blocks {
			hit, found = e, true
		}
		return nil
	})
	return hit, found
}

// byteReaderAt serves an in-memory image as an io.ReaderAt without copying.
type byteReaderAt []byte

func newByteReaderAt(b []byte) byteReaderAt { return byteReaderAt(b) }

func (b byteReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= int64(len(b)) {
		return 0, io.EOF
	}
	n := copy(p, b[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}
