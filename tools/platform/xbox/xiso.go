package xbox

// xiso.go reads an XDVDFS filesystem out of an original-Xbox disc image (an "XISO").
//
// An Xbox disc is not ISO 9660. Its filesystem, XDVDFS, is a single volume with:
//
//   - a volume descriptor at sector 32 of the game partition, carrying a 20-byte magic
//     "MICROSOFT*XBOX*MEDIA" at both its head and its tail, and the sector + byte-size
//     of the root directory;
//   - directories stored not as a flat list but as a *binary search tree* of entries,
//     ordered by name. Each 14-byte entry header carries a left- and a right-subtree
//     pointer (in 4-byte units from the start of the directory's extent), the file's
//     start sector, its byte length, an attribute byte, and an inline name. Entries are
//     4-byte aligned and never cross a 2048-byte sector boundary; the slack is 0xFF pad.
//
// Everything the disc names is a *sector relative to the game partition*, so a file's
// byte offset in the image is partitionBase + sector*2048. The partition base is not
// fixed: a raw XISO puts it at 0, but a full "redump" dump prefixes a video partition,
// so the game partition begins deep into the file. Rather than assume, Open scans for
// the volume-descriptor magic (its head-and-tail pair is a strong signature) and takes
// the base from wherever it lands. This mirrors the GameCube and PSP disc readers
// (gc/disc.go, psp/iso.go): Open -> a handle, walk the tree, Read(off,n), extract by
// path — but over XDVDFS rather than a GameCube FST or ISO 9660.

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	sectorSize = 2048 // XDVDFS logical sector

	// The volume descriptor lives at sector 32 of the game partition: 0x10000 bytes
	// past the partition's base.
	volumeDescriptorSector = 32
	volumeDescriptorOffset = volumeDescriptorSector * sectorSize

	// The magic that heads and tails the volume descriptor.
	xdvdfsMagic = "MICROSOFT*XBOX*MEDIA"

	// The directory-entry attribute bit that marks a subdirectory.
	attrDirectory = 0x10
)

// candidateBases are the partition bases seen in the wild, tried before falling back
// to a full scan: a raw XISO (0), and the two video-partition offsets of full XGD1
// "redump" dumps. The volume descriptor is at base+0x10000.
var candidateBases = []int64{0, 0xFD90000, 0x18300000}

// Image is an opened XDVDFS disc image. The file stays open and is streamed from, the
// way the drive would be: an Xbox disc is hundreds of megabytes and is never held whole.
type Image struct {
	Path string
	Size int64

	// Base is the byte offset of the game partition within the image. Every sector the
	// filesystem names is relative to it.
	Base int64

	rootSector uint32
	rootSize   uint32

	f   *os.File
	md5 string // computed on demand
}

// Entry is one file or directory in the filesystem.
type Entry struct {
	Name   string // the leaf name, as stored (case preserved)
	Path   string // full slash path from the volume root, e.g. "/media/default.xbe"
	IsDir  bool
	Size   uint32 // file length in bytes (directory extent length for directories)
	Sector uint32 // start sector, relative to the partition base
	Attr   byte   // raw XDVDFS attribute byte
}

func (e Entry) String() string {
	if e.IsDir {
		return fmt.Sprintf("%-40s  <dir>   (sector %d)", e.Path, e.Sector)
	}
	return fmt.Sprintf("%-40s  %10d  (sector %d)", e.Path, e.Size, e.Sector)
}

// Open reads an XISO's volume descriptor and prepares it for listing and extraction.
func Open(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	img := &Image{Path: path, Size: st.Size(), f: f}

	base, vd, err := img.findVolumeDescriptor()
	if err != nil {
		f.Close()
		return nil, err
	}
	img.Base = base
	// Root directory: sector at +0x14, byte size at +0x18.
	img.rootSector = le32(vd[0x14:])
	img.rootSize = le32(vd[0x18:])
	return img, nil
}

// findVolumeDescriptor locates the game partition. It reads the candidate sector at
// each known base first, then falls back to a chunked scan of the whole image; a match
// requires the magic at both the head and the tail of the 2048-byte descriptor, which
// no ordinary data run reproduces by accident.
func (img *Image) findVolumeDescriptor() (base int64, vd []byte, err error) {
	valid := func(sector []byte) bool {
		return len(sector) == sectorSize &&
			string(sector[0:len(xdvdfsMagic)]) == xdvdfsMagic &&
			string(sector[sectorSize-len(xdvdfsMagic):]) == xdvdfsMagic
	}

	for _, b := range candidateBases {
		off := b + volumeDescriptorOffset
		if off+sectorSize > img.Size {
			continue
		}
		sec, e := img.readAt(off, sectorSize)
		if e == nil && valid(sec) {
			return b, sec, nil
		}
	}

	// Fall back to a linear, sector-aligned scan. The magic is 2048-byte aligned (it
	// heads a sector), so read in large chunks and test each sector boundary.
	const chunk = 8 << 20
	buf := make([]byte, chunk)
	for off := int64(0); off+sectorSize <= img.Size; {
		n := int64(len(buf))
		if off+n > img.Size {
			n = img.Size - off
		}
		got, e := img.f.ReadAt(buf[:n], off)
		if got == 0 && e != nil && !errors.Is(e, io.EOF) {
			return 0, nil, e
		}
		for p := 0; p+sectorSize <= got; p += sectorSize {
			if string(buf[p:p+len(xdvdfsMagic)]) == xdvdfsMagic &&
				string(buf[p+sectorSize-len(xdvdfsMagic):p+sectorSize]) == xdvdfsMagic {
				b := off + int64(p) - volumeDescriptorOffset
				if b >= 0 {
					return b, append([]byte(nil), buf[p:p+sectorSize]...), nil
				}
			}
		}
		// Advance by whole sectors, overlapping nothing (the magic is sector-aligned).
		off += int64(got) - (int64(got) % sectorSize)
		if got < int(n) {
			break
		}
	}
	return 0, nil, errors.New("xiso: no XDVDFS volume descriptor found (is this an Xbox disc image?)")
}

// Close releases the image.
func (img *Image) Close() error { return img.f.Close() }

// readAt reads n bytes at an absolute image offset.
func (img *Image) readAt(off int64, n int) ([]byte, error) {
	if off < 0 || n < 0 || off+int64(n) > img.Size {
		return nil, fmt.Errorf("xiso: read of %d bytes at %#x is outside the %d-byte image", n, off, img.Size)
	}
	b := make([]byte, n)
	if _, err := img.f.ReadAt(b, off); err != nil {
		return nil, err
	}
	return b, nil
}

// Read returns n bytes at a partition-relative byte offset. This is the address space
// the filesystem's sector numbers live in (offset = sector*2048).
func (img *Image) Read(off int64, n int) ([]byte, error) {
	return img.readAt(img.Base+off, n)
}

// ReadDir lists the entries of the directory at path ("" or "/" is the root).
func (img *Image) ReadDir(path string) ([]Entry, error) {
	e, err := img.resolve(path)
	if err != nil {
		return nil, err
	}
	if !e.IsDir {
		return nil, fmt.Errorf("xiso: %q is not a directory", path)
	}
	return img.dirEntries(e.Sector, e.Size, e.Path)
}

// Walk visits every file and directory under the root, depth first.
func (img *Image) Walk(fn func(Entry) error) error {
	return img.walk(img.rootSector, img.rootSize, "", fn)
}

func (img *Image) walk(sector, size uint32, dirPath string, fn func(Entry) error) error {
	entries, err := img.dirEntries(sector, size, dirPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := fn(e); err != nil {
			return err
		}
		if e.IsDir {
			if err := img.walk(e.Sector, e.Size, e.Path, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReadFile returns the full contents of the file at path.
func (img *Image) ReadFile(path string) ([]byte, error) {
	e, err := img.resolve(path)
	if err != nil {
		return nil, err
	}
	if e.IsDir {
		return nil, fmt.Errorf("xiso: %q is a directory", path)
	}
	return img.ReadFileEntry(e)
}

// ReadFileEntry returns the full contents of a file entry.
func (img *Image) ReadFileEntry(e Entry) ([]byte, error) {
	if e.IsDir {
		return nil, fmt.Errorf("xiso: %q is a directory", e.Path)
	}
	return img.Read(int64(e.Sector)*sectorSize, int(e.Size))
}

// dirEntries walks the binary tree of a directory's extent and returns its entries.
//
// The extent is (size) bytes over ceil(size/2048) sectors. Every left/right pointer is
// a 4-byte-unit offset from the start of the extent, so the whole thing is read into
// memory and traversed with an explicit stack. A pointer of 0 (which would point at the
// tree root itself) or 0xFFFF is "no child"; padding between entries is 0xFF.
func (img *Image) dirEntries(sector, size uint32, dirPath string) ([]Entry, error) {
	if size == 0 {
		return nil, nil // an empty directory
	}
	nsec := (int(size) + sectorSize - 1) / sectorSize
	data, err := img.Read(int64(sector)*sectorSize, nsec*sectorSize)
	if err != nil {
		return nil, err
	}

	var out []Entry
	seen := make(map[uint32]bool)
	stack := []uint32{0} // byte offsets into the extent
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[p] { // guard against a malformed tree looping back on itself
			continue
		}
		seen[p] = true
		if int(p)+14 > len(data) {
			return nil, fmt.Errorf("xiso: directory %q has an entry at %#x past its %d-byte extent", dirPath, p, len(data))
		}
		rec := data[p:]
		left := le16(rec[0:])
		right := le16(rec[2:])
		startSector := le32(rec[4:])
		fsize := le32(rec[8:])
		attr := rec[12]
		nlen := int(rec[13])
		if int(p)+14+nlen > len(data) {
			return nil, fmt.Errorf("xiso: directory %q has a name overrunning its extent at %#x", dirPath, p)
		}
		name := string(rec[14 : 14+nlen])

		e := Entry{
			Name:   name,
			IsDir:  attr&attrDirectory != 0,
			Size:   fsize,
			Sector: startSector,
			Attr:   attr,
		}
		if dirPath == "" {
			e.Path = "/" + name
		} else {
			e.Path = dirPath + "/" + name
		}
		out = append(out, e)

		if left != 0 && left != 0xFFFF {
			stack = append(stack, uint32(left)*4)
		}
		if right != 0 && right != 0xFFFF {
			stack = append(stack, uint32(right)*4)
		}
	}
	return out, nil
}

// resolve walks a slash path to its entry. The empty path is the root, returned as a
// synthetic directory entry. Names match case-insensitively: the disc's names are
// mixed-case and a caller need not reproduce the casing.
func (img *Image) resolve(path string) (Entry, error) {
	cur := Entry{Name: "/", Path: "", IsDir: true, Sector: img.rootSector, Size: img.rootSize}
	for _, want := range splitPath(path) {
		if !cur.IsDir {
			return Entry{}, fmt.Errorf("xiso: %q is not a directory", cur.Path)
		}
		entries, err := img.dirEntries(cur.Sector, cur.Size, cur.Path)
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
			return Entry{}, fmt.Errorf("xiso: %q not found", path)
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

// MD5 is the image's hash, computed on first use. Hashing hundreds of megabytes is not
// free, so it is deferred until asked for.
func (img *Image) MD5() (string, error) {
	if img.md5 != "" {
		return img.md5, nil
	}
	if _, err := img.f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	h := md5.New()
	if _, err := io.Copy(h, img.f); err != nil {
		return "", err
	}
	img.md5 = fmt.Sprintf("%x", h.Sum(nil))
	return img.md5, nil
}
