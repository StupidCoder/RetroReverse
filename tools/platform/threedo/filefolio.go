package threedo

import (
	"fmt"
	"strings"
)

// filefolio.go high-level-emulates the Portfolio "File" folio — the high-level
// file API the boot loader uses to open and read files off the CD. On the real
// 3DO the folio's OpenDiskStream/ReadDiskStream/SeekDiskStream/CloseDiskStream
// build a buffered stream on top of the filesystem device (OpenDiskFile, then
// CMD_STATUS/CMD_READ IOReqs). Because the oracle has the whole disc in memory,
// we intercept the folio's vector calls directly and serve the bytes straight
// from the mounted Volume — the same shortcut tools/psx takes for BIOS calls.
//
// The game reaches these through a folio vector table: it LookupItem's the "File"
// folio to a base pointer and calls `LDR pc, [base, #-N]`. LookupItem (folio.go)
// hands back fileFolioBase, whose planted vectors jump into the HLE window's
// File-folio slice, landing here. Vector offsets, from the folio's public
// function table (FileUserFunctions, byte offset = 4 x function number):
//
//	-0x04 OpenDiskStream(name, bufSize) -> Stream* (0 = fail)
//	-0x08 ReadDiskStream(stream, buf, nBytes) -> bytesRead (-1 = error)
//	-0x0C SeekDiskStream(stream, offset, whence) -> newPos (-1 = error)
//	-0x10 CloseDiskStream(stream)
//	-0x14 LoadProgram / -0x1C OpenDirectoryItem / -0x20 OpenDirectoryPath (later)

// diskStream is the oracle's view of an open File-folio stream: the whole file
// held in memory with a read cursor. Its handle is the address of a small token
// struct allocated so the game has a distinct non-zero Stream pointer.
type diskStream struct {
	name string
	data []byte
	pos  int
}

// SeekOrigin values (enum SeekOrigin, filestream.h): 0 is SEEK_NOT ("no seek
// pending"), not a valid whence. NFS's file-size probe depends on these exact
// numbers: it does SeekDiskStream(s, 0, SEEK_END) to learn the length, then
// SeekDiskStream(s, 0, SEEK_SET) to rewind.
const (
	seekSet = 1
	seekCur = 2
	seekEnd = 3
)

// serviceFileFolio dispatches an intercepted File-folio vector call by its
// (positive) byte offset and returns to the caller with the result in r0.
func (m *Machine) serviceFileFolio(foff uint32) {
	c := m.CPU
	switch foff {
	case 0x04: // OpenDiskStream(name, bufSize)
		m.SetResultAndReturn(m.openDiskStream(m.readCStr(c.Reg(0))))
	case 0x08: // ReadDiskStream(stream, buf, nBytes)
		m.SetResultAndReturn(uint32(m.readDiskStream(c.Reg(0), c.Reg(1), int32(c.Reg(2)))))
	case 0x0C: // SeekDiskStream(stream, offset, whence)
		m.SetResultAndReturn(uint32(m.seekDiskStream(c.Reg(0), int32(c.Reg(1)), c.Reg(2))))
	case 0x10: // CloseDiskStream(stream)
		m.closeDiskStream(c.Reg(0))
		m.SetResultAndReturn(0)
	default:
		m.note(fmt.Sprintf("FileFolio[-0x%X] stub (r0=0x%08X %q r1=0x%08X r2=0x%08X)",
			foff, c.Reg(0), m.readCStr(c.Reg(0)), c.Reg(1), c.Reg(2)))
		m.SetResultAndReturn(0)
	}
}

// serviceOtherFolio logs and stubs a call into a folio the oracle does not yet
// implement (graphics, audio, ...), returning 0 so the boot proceeds. Its offset
// tells which vector was wanted, guiding the next folio to reimplement.
func (m *Machine) serviceOtherFolio(foff uint32) {
	c := m.CPU
	m.note(fmt.Sprintf("otherFolio[-0x%X] from 0x%08X (r0=0x%08X r1=0x%08X r2=0x%08X r3=0x%08X)", foff, c.Reg(14)-8, c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(3)))
	m.SetResultAndReturn(0)
}

// openDiskStream resolves a file by name on the disc and returns a non-zero
// Stream handle, or 0 if the file is not found (matching OpenDiskStream's
// NULL-on-failure contract). The handle is a small allocated token whose address
// keys the stream's bookkeeping.
func (m *Machine) openDiskStream(name string) uint32 {
	data, path, ok := m.loadDiscFile(name)
	if !ok {
		m.note(fmt.Sprintf("OpenDiskStream(%q) -> NOT FOUND", name))
		return 0
	}
	handle := m.dheap.alloc(0x20)
	if handle == 0 {
		return 0
	}
	m.streams[handle] = &diskStream{name: path, data: data}
	m.note(fmt.Sprintf("OpenDiskStream(%q) -> handle 0x%08X (%s, %d bytes)", name, handle, path, len(data)))
	return handle
}

// readDiskStream copies up to n bytes from the stream's cursor into buf and
// advances the cursor, returning the number of bytes read (0 at EOF, -1 on a bad
// handle) as ReadDiskStream does.
func (m *Machine) readDiskStream(handle, buf uint32, n int32) int32 {
	s := m.streams[handle]
	if s == nil {
		return -1
	}
	if n < 0 {
		n = 0
	}
	avail := len(s.data) - s.pos
	if avail <= 0 {
		return 0
	}
	if int(n) > avail {
		n = int32(avail)
	}
	for i := int32(0); i < n; i++ {
		m.Write(buf+uint32(i), s.data[s.pos+int(i)])
	}
	s.pos += int(n)
	return n
}

// seekDiskStream repositions the stream cursor and returns the new absolute
// position, or -1 on a bad handle, an invalid whence, or an out-of-range
// target — SeekDiskStream's contract. SEEK_END measures the offset back from
// the end of the file, so (0, SEEK_END) lands on (and returns) the length.
func (m *Machine) seekDiskStream(handle uint32, offset int32, whence uint32) int32 {
	s := m.streams[handle]
	if s == nil {
		return -1
	}
	var pos int
	switch whence {
	case seekSet:
		pos = int(offset)
	case seekCur:
		pos = s.pos + int(offset)
	case seekEnd:
		pos = len(s.data) - int(offset)
	default:
		return -1
	}
	if pos < 0 || pos > len(s.data) {
		return -1
	}
	s.pos = pos
	return int32(pos)
}

// closeDiskStream releases a stream and its token.
func (m *Machine) closeDiskStream(handle uint32) {
	if _, ok := m.streams[handle]; ok {
		delete(m.streams, handle)
		m.dheap.freeBlock(handle)
	}
}

// loadDiscFile resolves a game file name to disc bytes. It tries the name as a
// full path first, then falls back to a case-insensitive search by base name so
// path prefixes the loader prepends still resolve. Returns the data, the matched
// disc path and whether it was found.
func (m *Machine) loadDiscFile(name string) ([]byte, string, bool) {
	if m.vol == nil || name == "" {
		return nil, "", false
	}
	name = strings.TrimLeft(name, "/")
	if data, err := m.vol.ReadFile(name); err == nil {
		return data, name, true
	}
	base := name
	if i := strings.LastIndexByte(base, '/'); i >= 0 {
		base = base[i+1:]
	}
	var data []byte
	var path string
	found := false
	m.vol.Walk(func(e Entry) error {
		if !found && !e.IsDir && strings.EqualFold(e.Name, base) {
			if d, err := m.vol.ReadFile(e.Path); err == nil {
				data, path, found = d, e.Path, true
			}
		}
		return nil
	})
	return data, path, found
}
