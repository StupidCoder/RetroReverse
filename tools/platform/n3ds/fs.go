package n3ds

import "fmt"

// fs.go backs the Horizon fs service with the cartridge's real RomFS. fs:USER
// hands out a "file session" (an IFile) per OpenFile; the app then drives Read /
// GetSize / Close directly on that session handle. Read returns real bytes from
// the RomFS file, so the game loads its actual assets rather than zeros.
//
// A file the RomFS does not contain (a save-data probe, an empty binary path)
// opens as an empty session — Read returns zero bytes, GetSize returns 0 — which
// a title treats as "absent" and handles (creating the save, using a default).
// Only the read path is modelled; a write halts, since a cartridge boot never
// needs it.

// fsFile is one open IFile session: the file's bytes and its rooted path. A
// session into a writable archive (save data) carries the store key in save;
// its data is read through the store so writes are visible to later opens.
type fsFile struct {
	data []byte
	path string
	save string // key into Machine.saveFiles ("" = read-only RomFS session)
}

// isFileSession reports whether a handle names an open IFile session.
func (m *Machine) isFileSession(handle uint32) bool {
	_, ok := m.fsFiles[handle]
	return ok
}

// fsDirEntry is one child of an open IDirectory session.
type fsDirEntry struct {
	name  string
	isDir bool
	size  int64
}

// fsDir is one open IDirectory session: the directory's children and the read
// cursor.
type fsDir struct {
	path    string
	entries []fsDirEntry
	cursor  int
}

func (m *Machine) isDirSession(handle uint32) bool {
	_, ok := m.fsDirs[handle]
	return ok
}

// fsOpenDirectory services fs:USER OpenDirectory (0x080B): resolve the path in
// the RomFS and hand back a directory session listing its immediate children.
// Request: archive handle u64 (args 1-2), path type (3), path size (4), then a
// static-buffer translate pair with the path pointer at arg 6.
func (m *Machine) fsOpenDirectory(hdr ipcHeader) bool {
	pathType, pathSize := m.ipcArg(3), m.ipcArg(4)
	pathPtr := m.ipcArg(6)
	path := m.readFSPath(pathPtr, pathType, pathSize)
	if m.Verbose {
		fmt.Printf("    fsOpenDirectory pathType=%d path=%q\n", pathType, path)
	}

	dir, ok := m.romfsChildren(path)
	if !ok {
		// The save archive holds no directories; report a miss there with the
		// class the save state machine expects (see the fsOpen comment):
		// pre-format the save layer's own class, post-format plain NotFound.
		code := resultFSNotFound
		if id, open := m.fsArchives[m.ipcArg(1)]; open && id != archiveRomFS && !m.saveFormatted {
			code = resultFSSaveNotFound
		}
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6)
		m.WriteWord(m.cmdBuf()+4, code)
		return true
	}
	h := m.newHandle("fs-dir", false)
	m.fsDirs[h] = dir
	m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6|2)
	m.WriteWord(m.cmdBuf()+4, resultSuccess)
	m.WriteWord(m.cmdBuf()+8, 0) // translate descriptor: move 1 handle
	m.WriteWord(m.cmdBuf()+12, h)
	return true
}

// romfsChildren lists the immediate children of a RomFS directory path.
func (m *Machine) romfsChildren(path string) (*fsDir, bool) {
	if m.romfs == nil {
		return nil, false
	}
	if path == "" {
		return nil, false
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}
	if path == "/" {
		// The root always exists.
	} else {
		found := false
		for _, d := range m.romfs.Dirs {
			if d+"/" == path {
				found = true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	dir := &fsDir{path: path}
	seen := map[string]bool{}
	for _, d := range m.romfs.Dirs {
		if len(d) > len(path) && d[:len(path)] == path && !containsSlash(d[len(path):]) {
			name := d[len(path):]
			if !seen[name] {
				seen[name] = true
				dir.entries = append(dir.entries, fsDirEntry{name: name, isDir: true})
			}
		}
	}
	for _, f := range m.romfs.Files {
		if len(f.Path) > len(path) && f.Path[:len(path)] == path && !containsSlash(f.Path[len(path):]) {
			dir.entries = append(dir.entries, fsDirEntry{name: f.Path[len(path):], size: f.Size})
		}
	}
	return dir, true
}

func containsSlash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return true
		}
	}
	return false
}

// ipcDir services an open IDirectory session. Read (0x0801) fills the mapped
// buffer with FS_DirectoryEntry records — 0x228 bytes each: UTF-16 name at
// +0x000, an is-directory byte at +0x21C, the u64 file size at +0x220 — and
// returns how many were written; 0 means end of listing. Close (0x0802) ends
// the session.
func (m *Machine) ipcDir(handle uint32, hdr ipcHeader) bool {
	d := m.fsDirs[handle]
	switch hdr.Command {
	case 0x0801: // Read(count; mapped output buffer at arg 3)
		count := int(m.ipcArg(1))
		out := m.ipcArg(3)
		n := 0
		const entrySize = 0x228
		for n < count && d.cursor < len(d.entries) {
			e := d.entries[d.cursor]
			base := out + uint32(n*entrySize)
			for i := uint32(0); i < entrySize; i++ {
				m.Write(base+i, 0)
			}
			// UTF-16LE name, capped to the 0x106-unit field (with terminator).
			for i, r := range e.name {
				if i >= 0x105 {
					break
				}
				m.Write(base+uint32(i*2), byte(r))
				m.Write(base+uint32(i*2)+1, byte(uint16(r)>>8))
			}
			if e.isDir {
				m.Write(base+0x21C, 1)
			}
			m.WriteWord(base+0x220, uint32(e.size))
			m.WriteWord(base+0x224, uint32(e.size>>32))
			d.cursor++
			n++
		}
		if m.Verbose {
			fmt.Printf("    fsDirRead %q -> %d entries (cursor %d/%d)\n", d.path, n, d.cursor, len(d.entries))
		}
		m.ipcReply(hdr.Command, uint32(n))
		return true
	case 0x0802: // Close
		delete(m.fsDirs, handle)
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("fs dir command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

// fsOpenFile services fs:USER OpenFile / OpenFileDirectly: it reads the requested
// file path, resolves it in the RomFS, and returns a file-session handle.
//
// OpenFileDirectly (0x0803) layout, verified against a live request: 8 normal
// words then 4 translate — the file path type is cmdbuf[6], its size cmdbuf[7],
// and its buffer pointer the last translate word (cmdbuf[13]). OpenFile (0x0802)
// omits the archive fields, shifting the file path two words earlier.
func (m *Machine) fsOpenFile(hdr ipcHeader) bool {
	var archive, pathType, pathSize, pathPtr uint32
	saveArchive := false
	if hdr.Command == 0x0803 { // OpenFileDirectly
		archive = m.ipcArg(2)
		pathType, pathSize = m.ipcArg(6), m.ipcArg(7)
		pathPtr = m.ipcArg(13)
	} else { // OpenFile: transaction, archive handle u64 (2-3), path type/size, flags, attr, then translate
		archiveHandle := m.ipcArg(2)
		if id, ok := m.fsArchives[archiveHandle]; ok && id != archiveRomFS {
			saveArchive = true // an opened writable archive (save data)
		}
		pathType, pathSize = m.ipcArg(4), m.ipcArg(5)
		pathPtr = m.ipcArg(9)
	}

	path := m.readFSPath(pathPtr, pathType, pathSize)
	if m.Verbose {
		fmt.Printf("    fsOpen cmd=0x%04X archive=0x%08X save=%v pathType=%d pathSize=%d flags=0x%X attr=0x%X path=%q\n", hdr.Command, archive, saveArchive, pathType, pathSize, m.ipcArg(6), m.ipcArg(7), path)
	}

	if saveArchive {
		data, ok := m.saveFiles[path]
		if !ok && hdr.Command == 0x0802 && m.ipcArg(6)&4 != 0 {
			// OPEN_FLAG_CREATE: the game creates its save file by opening it
			// with flags WRITE|CREATE (0x6) right after FormatSaveData —
			// ignoring the flag returned NotFound, which the save layer threw
			// as fatal 0xC8804478. Create it empty.
			m.saveFiles[path] = []byte{}
			data, ok = m.saveFiles[path], true
		}
		if !ok {
			// The error class steers the game's save state machine: before the
			// archive is formatted, a miss must be the save layer's own class
			// (description 0x154 — "save fresh", routes to FormatSaveData; a
			// plain NotFound there is thrown as fatal by the filter at
			// 0x001A5A68). AFTER a successful format, a miss must be the plain
			// NotFound.
			code := resultFSSaveNotFound
			if m.saveFormatted {
				code = resultFSNotFound
			}
			m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6)
			m.WriteWord(m.cmdBuf()+4, code)
			return true
		}
		sess := &fsFile{path: path, save: path, data: data}
		h := m.newHandle("fs-file", false)
		m.fsFiles[h] = sess
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6|2)
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, 0)
		m.WriteWord(m.cmdBuf()+12, h)
		return true
	}

	var data []byte
	found := false
	switch {
	case archive == archiveRomFS && m.romfsRaw != nil:
		// The game opens ARCHIVE_ROMFS and expects the *level-3 filesystem* image
		// (the RomFS header + dir/file tables + file data), not the raw IVFC
		// container — fs strips the IVFC hash-tree wrapper. Hand back the region
		// starting at the level-3 media offset.
		l3 := int64(0)
		if m.romfs != nil {
			l3 = m.romfs.Levels[2].Offset
		}
		data, found = m.romfsRaw[l3:], true
		path = "<romfs-l3>"
	case m.romfs != nil && path != "":
		if d, err := m.romfs.File(path); err == nil {
			data, found = d, true
		}
	}
	if !found {
		// A file the RomFS does not hold (a first-boot save probe, an absent
		// asset) must report "not found" so the title creates or defaults it —
		// returning an empty session instead makes it treat a real file as
		// corrupt and throw a fatal error.
		if m.Verbose {
			fmt.Printf("    fsOpenFile %q -> NOT FOUND\n", path)
		}
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6)
		m.WriteWord(m.cmdBuf()+4, resultFSNotFound)
		return true
	}
	sess := &fsFile{path: path, data: data}
	h := m.newHandle("fs-file", false)
	m.fsFiles[h] = sess
	if m.Verbose {
		fmt.Printf("    fsOpenFile %q -> handle 0x%08X (%d bytes)\n", path, h, len(sess.data))
	}
	m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6|2)
	m.WriteWord(m.cmdBuf()+4, resultSuccess)
	m.WriteWord(m.cmdBuf()+8, 0) // translate descriptor: move 1 handle
	m.WriteWord(m.cmdBuf()+12, h)
	return true
}

// resultFSNotFound is the Horizon "file/path not found" result code.
const resultFSNotFound uint32 = 0xC8804478

// resultFSSaveNotFound is the "not found" of the save-data layer — same level/
// summary/module as resultFSNotFound but description 0x154, the class the
// game's save code (filter at 0x001A5A68, range [0x154,0x168)) treats as "the
// save is fresh, create it" rather than a fatal error.
const resultFSSaveNotFound uint32 = 0xC8804554

// archiveRomFS is the FS_ArchiveID for a title's own RomFS.
const archiveRomFS = 0x00000003

// readFSPath reads a file path from a mapped buffer. Type 3 is ASCII, type 4 is
// UTF-16 (both '/'-rooted RomFS paths); anything else (binary, empty) yields "".
func (m *Machine) readFSPath(ptr, ptype, size uint32) string {
	if ptr == 0 || size == 0 || size > 0x1000 {
		return ""
	}
	switch ptype {
	case 3: // ASCII
		var b []byte
		for i := uint32(0); i < size; i++ {
			c := m.Read(ptr + i)
			if c == 0 {
				break
			}
			b = append(b, c)
		}
		return string(b)
	case 4: // UTF-16LE
		var r []rune
		for i := uint32(0); i+1 < size; i += 2 {
			u := uint16(m.Read(ptr+i)) | uint16(m.Read(ptr+i+1))<<8
			if u == 0 {
				break
			}
			r = append(r, rune(u))
		}
		return string(r)
	}
	return ""
}

// ipcFile services an open IFile session (routed from handleIPC).
func (m *Machine) ipcFile(handle uint32, hdr ipcHeader) bool {
	f := m.fsFiles[handle]
	switch hdr.Command {
	case 0x0802: // Read(offset u64, size) → bytes read, into the mapped buffer
		off := int64(uint64(m.ipcArg(1)) | uint64(m.ipcArg(2))<<32)
		size := int64(m.ipcArg(3))
		bufPtr := m.ipcArg(5) // cmdbuf[5]: after the read-count descriptor
		n := int64(0)
		if f != nil && f.save != "" {
			f.data = m.saveFiles[f.save] // writes since open are visible
		}
		if f != nil && off >= 0 && off < int64(len(f.data)) {
			n = size
			if off+n > int64(len(f.data)) {
				n = int64(len(f.data)) - off
			}
			for i := int64(0); i < n; i++ {
				m.Write(bufPtr+uint32(i), f.data[off+i])
			}
		}
		if m.Verbose {
			var head uint32
			if f != nil && n >= 4 {
				head = m.ReadWord(bufPtr)
			}
			fmt.Printf("    IFile Read h=0x%08X off=%d size=%d -> %d bytes (flen=%d) head=%08X\n", handle, off, size, n, fileLen(f), head)
		}
		m.ipcReply(hdr.Command, uint32(n))
		return true
	case 0x0804: // GetSize → u64
		var sz uint64
		if f != nil {
			if f.save != "" {
				f.data = m.saveFiles[f.save]
			}
			sz = uint64(len(f.data))
		}
		m.ipcReply(hdr.Command, uint32(sz), uint32(sz>>32))
		return true
	case 0x0808: // Close
		delete(m.fsFiles, handle)
		delete(m.handles, handle)
		m.ipcReply(hdr.Command)
		return true
	case 0x0803: // Write(offset u64, size, flags) from the mapped buffer
		off := int64(uint64(m.ipcArg(1)) | uint64(m.ipcArg(2))<<32)
		size := int64(m.ipcArg(3))
		bufPtr := m.ipcArg(6) // after the write-flags word and buffer descriptor
		if f != nil && f.save != "" && off >= 0 && size >= 0 {
			data := m.saveFiles[f.save]
			if need := off + size; int64(len(data)) < need {
				grown := make([]byte, need)
				copy(grown, data)
				data = grown
			}
			for i := int64(0); i < size; i++ {
				data[off+i] = m.Read(bufPtr + uint32(i))
			}
			m.saveFiles[f.save] = data
			f.data = data
			if m.Verbose {
				fmt.Printf("    IFile Write %q off=%d size=%d\n", f.save, off, size)
			}
			m.ipcReply(hdr.Command, uint32(size))
			return true
		}
		// A write into a read-only (RomFS) session would be a game bug; a write
		// into an unknown session is ours. Stop loudly.
		m.CPU.Halt("IFile Write on non-writable session (path %q) at 0x%08X", filePath(f), m.CPU.PC())
		return true
	case 0x0805: // SetSize(u64) — real for save sessions: the game sizes its
		// fresh save file before chunk-writing it, and verifies by read-back.
		if f != nil && f.save != "" {
			size := int64(uint64(m.ipcArg(1)) | uint64(m.ipcArg(2))<<32)
			data := m.saveFiles[f.save]
			if int64(len(data)) != size {
				resized := make([]byte, size)
				copy(resized, data)
				m.saveFiles[f.save] = resized
				f.data = resized
			}
			if m.Verbose {
				fmt.Printf("    IFile SetSize %q -> %d\n", f.save, size)
			}
		}
		m.ipcReply(hdr.Command)
		return true
	case 0x0806, 0x0807: // GetAttributes / SetAttributes
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("IFile command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

func fileLen(f *fsFile) int {
	if f == nil {
		return -1
	}
	return len(f.data)
}

func filePath(f *fsFile) string {
	if f == nil {
		return "<nil>"
	}
	return f.path
}

// fsOpenArchive services fs:USER OpenArchive (0x080C): record which archive ID
// the returned handle names, so OpenFile can route save-data opens to the
// writable store. Request: archive ID (1), path type (2), path size (3), then
// the path translate pair; reply carries the archive handle as a u64.
func (m *Machine) fsOpenArchive(hdr ipcHeader) bool {
	id := m.ipcArg(1)
	h := m.newHandle("fs-archive", false)
	m.fsArchives[h] = id
	if m.Verbose {
		fmt.Printf("    fsOpenArchive id=0x%X -> handle 0x%08X\n", id, h)
	}
	m.ipcReply(hdr.Command, h, 0)
	return true
}

// fsDeleteFile services fs:USER DeleteFile (0x0804) on the writable archive.
// Request (5 normal + 2 translate): transaction (1), archive handle u64 (2-3),
// path type (4), path size (5), then the path translate pair (pointer at 7).
func (m *Machine) fsDeleteFile(hdr ipcHeader) bool {
	path := m.readFSPath(m.ipcArg(7), m.ipcArg(4), m.ipcArg(5))
	if m.Verbose {
		fmt.Printf("    fsDeleteFile %q\n", path)
	}
	if _, ok := m.saveFiles[path]; ok {
		delete(m.saveFiles, path)
		m.ipcReply(hdr.Command)
	} else {
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6)
		m.WriteWord(m.cmdBuf()+4, resultFSNotFound)
	}
	return true
}

// fsCreateFile services fs:USER CreateFile (0x0808) on a writable archive: an
// empty file of the requested size appears in the save store. Request layout,
// from the game's wrapper at 0x001EDE30 (header 0x08080202, 8 normal + 2
// translate): transaction (1), archive handle u64 (2-3), path type (4), path
// size (5), attributes (6), file size u64 (7-8), path pointer (10).
func (m *Machine) fsCreateFile(hdr ipcHeader) bool {
	pathType, pathSize := m.ipcArg(4), m.ipcArg(5)
	size := int64(uint64(m.ipcArg(7)) | uint64(m.ipcArg(8))<<32)
	path := m.readFSPath(m.ipcArg(10), pathType, pathSize)
	if m.Verbose {
		fmt.Printf("    fsCreateFile %q size=%d\n", path, size)
	}
	if size < 0 || size > 1<<24 {
		m.CPU.Halt("fs CreateFile %q with implausible size %d at 0x%08X", path, size, m.CPU.PC())
		return true
	}
	m.saveFiles[path] = make([]byte, size)
	m.ipcReply(hdr.Command)
	return true
}
