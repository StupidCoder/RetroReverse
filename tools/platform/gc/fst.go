package gc

// fst.go reads the disc's filesystem.
//
// The FST is not ISO 9660 and not a tree of blocks: it is one flat array of 12-byte
// entries followed by one string table, and the hierarchy is expressed by arithmetic
// on indices rather than by pointers. An entry is either a file — carrying its byte
// offset on the disc and its length — or a directory, carrying its parent's index and
// the index of the first entry that is *not* inside it. So a directory's children are
// simply the entries between it and that bound, and walking the tree is walking the
// array in order with a stack of pending bounds.
//
// Entry 0 is the root, and its "next" field is the total number of entries, which is
// how the string table's start is found: it follows the last entry.
//
// ByOffset is the reason this file earns its place. Every read the game performs of
// its own disc arrives at the drive as a byte offset and nothing else. Resolving that
// offset back to the file it lands in turns an unreadable log of numbers into a
// readable account of what the game is loading, and it is how the boot oracle reports
// the disc traffic it observes.

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const fstEntrySize = 12

// File is one entry in the disc's filesystem.
type File struct {
	Path   string // "/dvdroot/model/luigi.arc", with a leading slash
	Name   string
	Dir    bool
	Offset int64 // byte offset on the disc; a directory has none
	Size   int64
}

// FST is the disc's filesystem, flattened.
type FST struct {
	Entries []File // in FST order: entry 0 is the root, so a directory precedes its children

	files []File // the non-directory entries, sorted by Offset, for ByOffset
}

// ParseFST reads the raw FST out of the disc.
func ParseFST(b []byte) (*FST, error) {
	if len(b) < fstEntrySize {
		return nil, errors.New("gc: the FST is too short to hold even its root")
	}
	// The root's length field is the entry count — including the root itself.
	count := int(be32(b[8:]))
	if count < 1 {
		return nil, fmt.Errorf("gc: the FST claims %d entries", count)
	}
	strTab := count * fstEntrySize
	if strTab > len(b) {
		return nil, fmt.Errorf("gc: the FST claims %d entries (%d bytes) but is only %d bytes", count, strTab, len(b))
	}
	names := b[strTab:]

	f := &FST{Entries: make([]File, count)}
	f.Entries[0] = File{Path: "/", Name: "/", Dir: true, Size: int64(count)}

	// dirEnd is the stack of "this directory ends before index N" bounds, and dirPath
	// the paths that go with them. The root's bound is the entry count.
	dirEnd := []int{count}
	dirPath := []string{""}

	for i := 1; i < count; i++ {
		e := b[i*fstEntrySize:]
		isDir := e[0] != 0
		nameOff := int(e[1])<<16 | int(e[2])<<8 | int(e[3])
		a := be32(e[4:])
		c := be32(e[8:])

		if nameOff >= len(names) {
			return nil, fmt.Errorf("gc: entry %d names a string at %#x, past the %d-byte string table", i, nameOff, len(names))
		}
		name := cstr(names[nameOff:])
		if name == "" {
			return nil, fmt.Errorf("gc: entry %d has an empty name", i)
		}

		// Pop every directory this entry falls outside of.
		for len(dirEnd) > 1 && i >= dirEnd[len(dirEnd)-1] {
			dirEnd = dirEnd[:len(dirEnd)-1]
			dirPath = dirPath[:len(dirPath)-1]
		}
		path := dirPath[len(dirPath)-1] + "/" + name

		if isDir {
			// a = the parent's index, c = the first index outside this directory.
			next := int(c)
			if next <= i || next > count {
				return nil, fmt.Errorf("gc: directory %q at entry %d ends at %d, which is not after it and inside %d", path, i, next, count)
			}
			f.Entries[i] = File{Path: path, Name: name, Dir: true, Size: int64(next - i - 1)}
			dirEnd = append(dirEnd, next)
			dirPath = append(dirPath, path)
			continue
		}
		f.Entries[i] = File{Path: path, Name: name, Offset: int64(a), Size: int64(c)}
	}

	for _, e := range f.Entries {
		if !e.Dir {
			f.files = append(f.files, e)
		}
	}
	sort.Slice(f.files, func(i, j int) bool { return f.files[i].Offset < f.files[j].Offset })
	return f, nil
}

// Files returns the disc's files, in FST order, directories excluded.
func (f *FST) Files() []File {
	var out []File
	for _, e := range f.Entries {
		if !e.Dir {
			out = append(out, e)
		}
	}
	return out
}

// ByPath finds a file by its full path ("/dvdroot/foo.arc"). The comparison is
// case-insensitive: the disc's own names are mixed-case and the game's requests need
// not agree with them.
func (f *FST) ByPath(path string) (File, bool) {
	for _, e := range f.Entries {
		if strings.EqualFold(e.Path, path) {
			return e, true
		}
	}
	return File{}, false
}

// ByOffset names the file that contains a byte offset, and how far into it the offset
// falls. This is what turns the drive's traffic into a readable log: the game asks for
// a byte offset, and this says which of its own files it is asking for.
func (f *FST) ByOffset(off int64) (file File, within int64, ok bool) {
	// The first file starting after off is at i; the candidate is the one before it.
	i := sort.Search(len(f.files), func(i int) bool { return f.files[i].Offset > off })
	if i == 0 {
		return File{}, 0, false
	}
	e := f.files[i-1]
	if off >= e.Offset+e.Size {
		return File{}, 0, false // in the padding between files
	}
	return e, off - e.Offset, true
}

// Validate checks the filesystem against the image it came from: every file's extent
// must lie inside the disc, and no two may overlap. A filesystem that passes this has
// been read correctly — a misparse produces wild offsets, and wild offsets collide.
func (f *FST) Validate(discSize int64) error {
	for _, e := range f.files {
		if e.Offset < 0 || e.Size < 0 || e.Offset+e.Size > discSize {
			return fmt.Errorf("%s: extent %#x+%#x runs past the %d-byte image", e.Path, e.Offset, e.Size, discSize)
		}
	}
	for i := 1; i < len(f.files); i++ {
		prev, cur := f.files[i-1], f.files[i]
		if prev.Offset+prev.Size > cur.Offset {
			return fmt.Errorf("%s (%#x+%#x) overlaps %s (%#x)", prev.Path, prev.Offset, prev.Size, cur.Path, cur.Offset)
		}
	}
	return nil
}
