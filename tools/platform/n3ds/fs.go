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

// fsFile is one open IFile session: the file's bytes and its rooted path.
type fsFile struct {
	data []byte
	path string
}

// isFileSession reports whether a handle names an open IFile session.
func (m *Machine) isFileSession(handle uint32) bool {
	_, ok := m.fsFiles[handle]
	return ok
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
	if hdr.Command == 0x0803 { // OpenFileDirectly
		archive = m.ipcArg(2)
		pathType, pathSize = m.ipcArg(6), m.ipcArg(7)
		pathPtr = m.ipcArg(13)
	} else { // OpenFile: transaction, archive handle(2), path type/size, flags, attr, then translate
		pathType, pathSize = m.ipcArg(4), m.ipcArg(5)
		pathPtr = m.ipcArg(9)
	}

	path := m.readFSPath(pathPtr, pathType, pathSize)
	var data []byte
	found := false
	switch {
	case archive == archiveRomFS && m.romfsRaw != nil:
		// The game opens ARCHIVE_ROMFS as one big file (a binary path) and walks
		// the IVFC filesystem itself — hand it the raw RomFS region.
		data, found = m.romfsRaw, true
		path = "<romfs>"
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
		if f != nil && off >= 0 && off < int64(len(f.data)) {
			n = size
			if off+n > int64(len(f.data)) {
				n = int64(len(f.data)) - off
			}
			for i := int64(0); i < n; i++ {
				m.Write(bufPtr+uint32(i), f.data[off+i])
			}
		}
		m.ipcReply(hdr.Command, uint32(n))
		return true
	case 0x0804: // GetSize → u64
		var sz uint64
		if f != nil {
			sz = uint64(len(f.data))
		}
		m.ipcReply(hdr.Command, uint32(sz), uint32(sz>>32))
		return true
	case 0x0808: // Close
		delete(m.fsFiles, handle)
		delete(m.handles, handle)
		m.ipcReply(hdr.Command)
		return true
	case 0x0803, 0x0805, 0x0806, 0x0807: // Write / SetSize / GetAttributes / SetAttributes
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("IFile command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}
