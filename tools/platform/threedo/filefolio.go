package threedo

import (
	"fmt"
	"strings"

	"retroreverse.com/tools/cpu/arm60"
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
//	-0x14 LoadProgram / -0x1C OpenDirectoryItem (later)
//	-0x20 OpenDirectoryPath(path) -> Directory* (0 = fail)
//	-0x24 ReadDirectory(dir, DirectoryEntry* buf) -> 0 per entry, <0 at end
//	-0x28 CloseDirectory(dir)

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
	case 0x20: // OpenDirectoryPath(path)
		m.SetResultAndReturn(m.openDirectoryPath(m.readCStr(c.Reg(0))))
	case 0x24: // ReadDirectory(dir, DirectoryEntry* buf)
		m.SetResultAndReturn(uint32(m.readDirectory(c.Reg(0), c.Reg(1))))
	case 0x28: // CloseDirectory(dir)
		m.closeDirectory(c.Reg(0))
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

// dirScan is an open directory enumeration: the resolved entry list and a
// cursor, keyed by the Directory* token handed to the game.
type dirScan struct {
	entries []Entry
	pos     int
}

// openDirectoryPath opens a directory for enumeration and returns a non-zero
// Directory* token (0 = fail). "/nvram" paths resolve to an empty directory:
// the console's battery-backed store exists but a fresh unit holds no files,
// which is exactly the state a first boot sees — the game's save-file scan
// ends immediately and it falls back to defaults.
func (m *Machine) openDirectoryPath(path string) uint32 {
	scan := &dirScan{}
	trimmed := strings.TrimLeft(path, "/")
	if !strings.HasPrefix(strings.ToLower(trimmed), "nvram") {
		if m.vol == nil {
			return 0
		}
		entries, err := m.vol.ReadDir(trimmed)
		if err != nil {
			m.note(fmt.Sprintf("OpenDirectoryPath(%q) -> NOT FOUND", path))
			return 0
		}
		scan.entries = entries
	}
	handle := m.dheap.alloc(0x20)
	if handle == 0 {
		return 0
	}
	m.dirs[handle] = scan
	m.note(fmt.Sprintf("OpenDirectoryPath(%q) -> dir 0x%08X (%d entries)", path, handle, len(scan.entries)))
	return handle
}

// readDirectory fills buf with the scan's next DirectoryEntry (directory.h:
// eight uint32 stat words, then the 32-byte name at +0x24 and the location
// word at +0x44) and returns 0, or a negative status once the entries are
// exhausted — the game loops on the sign bit.
func (m *Machine) readDirectory(dir, buf uint32) int32 {
	s := m.dirs[dir]
	if s == nil || s.pos >= len(s.entries) {
		return -1
	}
	e := s.entries[s.pos]
	s.pos++
	var flags, typ uint32
	if e.IsDir {
		flags = 1 // FILE_IS_DIRECTORY
		typ = 0x2A646972
	} else {
		typ = 0x2A2A2A2A
	}
	blocks := uint32((e.Size + 2047) / 2048)
	for i, w := range []uint32{
		flags,             // +0x00 de_Flags
		uint32(s.pos),     // +0x04 de_UniqueIdentifier
		typ,               // +0x08 de_Type
		2048,              // +0x0C de_BlockSize
		uint32(e.Size),    // +0x10 de_ByteCount
		blocks,            // +0x14 de_BlockCount
		0,                 // +0x18 de_Burst
		0,                 // +0x1C de_Gap
		uint32(len(e.Copies)) + 1, // +0x20 de_AvatarCount
	} {
		m.writeWord(buf+uint32(i)*4, w)
	}
	name := e.Name
	if len(name) > 31 {
		name = name[:31]
	}
	for i := 0; i < 32; i++ {
		b := byte(0)
		if i < len(name) {
			b = name[i]
		}
		m.Write(buf+0x24+uint32(i), b)
	}
	m.writeWord(buf+0x44, uint32(e.Block)) // de_Location
	return 0
}

// closeDirectory releases a directory scan and its token.
func (m *Machine) closeDirectory(dir uint32) {
	if _, ok := m.dirs[dir]; ok {
		delete(m.dirs, dir)
		m.dheap.freeBlock(dir)
	}
}

// closeDiskStream releases a stream and its token.
func (m *Machine) closeDiskStream(handle uint32) {
	if _, ok := m.streams[handle]; ok {
		delete(m.streams, handle)
		m.dheap.freeBlock(handle)
	}
}

// --- File folio SWIs (FILEFOLIOSWI = 0x30000, filefunctions.h) ----------------
//
// These are the raw open-file layer under the stream API: OpenDiskFile returns
// a file Item the program then drives with IOReqs (CMD_STATUS for the size,
// CMD_READ for data). NFS uses it for its NVRAM save files and the movie
// streamer. NVRAM is modelled as an in-memory store that starts empty — a
// fresh console — so first-boot code takes its defaults path; saves written
// during the run persist for the rest of the run.
const (
	swiOpenDiskFile  = 0x30000
	swiCloseDiskFile = 0x30001
	swiChangeDir     = 0x30007
	swiGetDir        = 0x30008
	swiCreateFile    = 0x30009
	swiDeleteFile    = 0x3000A

	fileErrNotFound = 0xFFFFFC65 // a negative filesystem error (NoFile)
)

// nvramPath returns the store key for a "/nvram/..." path and whether the path
// is on the NVRAM filesystem at all.
func nvramPath(path string) (string, bool) {
	trimmed := strings.TrimLeft(path, "/")
	low := strings.ToLower(trimmed)
	if !strings.HasPrefix(low, "nvram") {
		return "", false
	}
	rest := strings.TrimLeft(trimmed[len("nvram"):], "/")
	return strings.ToLower(rest), true
}

// fileFolioSWI services a File-folio SWI, returning false for numbers it does
// not know so the caller logs them as stubs.
func (m *Machine) fileFolioSWI(c *arm60.CPU, swi uint32) bool {
	switch swi {
	case swiOpenDiskFile:
		c.SetReg(0, m.openDiskFile(m.readCStr(c.Reg(0))))
	case swiCloseDiskFile:
		if it := m.items[int32(c.Reg(0))]; it != nil {
			delete(m.items, it.num)
		}
		c.SetReg(0, 0)
	case swiCreateFile:
		path := m.readCStr(c.Reg(0))
		if key, ok := nvramPath(path); ok {
			if _, exists := m.nvram[key]; !exists {
				m.nvram[key] = nil
			}
			m.note(fmt.Sprintf("CreateFile(%q)", path))
			c.SetReg(0, 0)
		} else {
			c.SetReg(0, fileErrNotFound) // the disc is read-only
		}
	case swiDeleteFile:
		path := m.readCStr(c.Reg(0))
		if key, ok := nvramPath(path); ok {
			delete(m.nvram, key)
			m.note(fmt.Sprintf("DeleteFile(%q)", path))
		}
		c.SetReg(0, 0)
	case swiChangeDir, swiGetDir:
		c.SetReg(0, 0)
	default:
		return false
	}
	return true
}

// openDiskFile resolves a path to a file Item whose IOReqs the file-device
// handlers serve: NVRAM paths bind to the in-memory store (error if absent,
// like a fresh console), disc paths to the mounted volume.
func (m *Machine) openDiskFile(path string) uint32 {
	if key, ok := nvramPath(path); ok {
		if _, exists := m.nvram[key]; !exists {
			m.note(fmt.Sprintf("OpenDiskFile(%q) -> no such NVRAM file", path))
			return fileErrNotFound
		}
		it := m.createItem(0x030D, 0, 0) // file folio (3) << 8 | FILENODE
		it.name = "/nvram/" + key
		m.note(fmt.Sprintf("OpenDiskFile(%q) -> item %d (nvram)", path, it.num))
		return uint32(it.num)
	}
	_, resolved, ok := m.loadDiscFile(path)
	if !ok {
		m.note(fmt.Sprintf("OpenDiskFile(%q) -> NOT FOUND", path))
		return fileErrNotFound
	}
	it := m.createItem(0x030D, 0, 0)
	it.name = resolved
	m.note(fmt.Sprintf("OpenDiskFile(%q) -> item %d (%s)", path, it.num, resolved))
	return uint32(it.num)
}

// fileData returns the bytes and block size behind a file item opened by
// openDiskFile: NVRAM store entries are byte-addressed, disc files use the
// 2048-byte Opera block size.
func (m *Machine) fileData(name string) (data []byte, blockSize uint32, nvramKey string, ok bool) {
	if key, isNV := nvramPath(name); isNV {
		d, exists := m.nvram[key]
		return d, 1, key, exists
	}
	d, _, found := m.loadDiscFile(name)
	return d, 2048, "", found
}

// loadDiscFile resolves a game file name to disc bytes. It tries the name as a
// full path first, then falls back to a case-insensitive search by base name so
// path prefixes the loader prepends still resolve. Returns the data, the matched
// disc path and whether it was found.
func (m *Machine) loadDiscFile(name string) ([]byte, string, bool) {
	if m.vol == nil || name == "" {
		return nil, "", false
	}
	if m.NoStreams && strings.Contains(strings.ToLower(name), ".stream") {
		// Movie playback needs the audio folio + DataStreamer subscribers, which
		// the HLE does not model yet; a half-working movie player corrupts itself
		// (its timing/audio setup runs on stubbed answers). Failing the stream
		// files takes the game's own missing-file path, which skips the movies
		// cleanly and proceeds to the front-end — the same graceful fallback the
		// retail code ships for a scratched disc.
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
