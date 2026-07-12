package psp

// io.go is the IoFileMgrForUser HLE: sceIoOpen/Read/Lseek/Close backed by the
// mounted UMD volume, plus the stdout/stderr descriptors feeding the TTY. Game
// paths name the disc device ("disc0:/PSP_GAME/..." or "umd0:/..."); the device
// prefix is stripped and the rest resolved on the ISO 9660 volume.

import "strings"

// ioFile is one open descriptor on the UMD volume.
type ioFile struct {
	path     string // volume path (device prefix stripped)
	ent      Entry
	pos      int64
	async    int64 // last async operation's 64-bit result (read now, reported later)
	hasAsync bool  // an async result is pending retrieval
	dir      []Entry // directory descriptors (sceIoDopen): the entries to hand out
	dirPos   int     // next entry sceIoDread returns
}

const (
	fdStdout    = 1
	fdStderr    = 2
	fdFirstFile = 4

	errIoNoEnt = 0x80010002 // SCE_KERNEL_ERROR_ERRNO_FILE_NOT_FOUND
	errIoBadFd = 0x80010009
)

// SetVolume attaches the mounted UMD so the IO syscalls can serve files.
func (m *Machine) SetVolume(v *Volume) { m.vol = v }

// devicePath strips the device prefix ("disc0:", "umd0:", …) and leading slashes
// from a game path, yielding a volume path.
func devicePath(p string) string {
	if i := strings.IndexByte(p, ':'); i >= 0 {
		p = p[i+1:]
	}
	return strings.TrimLeft(p, "/")
}

func (m *Machine) ioOpen(path string) uint32 {
	if m.vol == nil {
		m.note("sceIoOpen(%q) with no volume mounted", path)
		return errIoNoEnt
	}
	vp := devicePath(path)
	e, err := m.vol.resolve(vp)
	if err != nil || e.IsDir {
		m.note("sceIoOpen(%q): not found", path)
		return errIoNoEnt
	}
	fd := m.nextFd
	m.nextFd++
	m.files[fd] = &ioFile{path: vp, ent: e, pos: 0}
	m.note("sceIoOpen(%q) -> fd %d (lba %d size %d)", path, fd, e.Block, e.Size)
	return fd
}

func (m *Machine) ioClose(fd uint32) uint32 {
	if _, ok := m.files[fd]; !ok {
		return errIoBadFd
	}
	delete(m.files, fd)
	return 0
}

// --- Async IO -------------------------------------------------------------
//
// The PSP async IO calls queue an operation on a file descriptor and report its
// result later through sceIoWaitAsync / sceIoPollAsync. This oracle's volume
// reads complete instantly, so each async call performs the operation now and
// stores its 64-bit result on the descriptor; the wait/poll calls hand that
// result back and report "done".

// ioOpenAsync opens a file and stores the fd as the pending async result, so a
// following sceIoWaitAsync(fd, &res) reports the opened fd.
func (m *Machine) ioOpenAsync(path string) uint32 {
	fd := m.ioOpen(path)
	if fd&0x80000000 == 0 { // a real fd, not an error code
		f := m.files[fd]
		f.async, f.hasAsync = int64(fd), true
	}
	return fd
}

func (m *Machine) ioReadAsync(fd, buf, n uint32) uint32 {
	f, ok := m.files[fd]
	if !ok {
		return errIoBadFd
	}
	f.async, f.hasAsync = int64(int32(m.ioRead(fd, buf, n))), true
	return 0
}

func (m *Machine) ioLseekAsync(fd uint32, off int64, whence uint32) uint32 {
	f, ok := m.files[fd]
	if !ok {
		return errIoBadFd
	}
	f.async, f.hasAsync = m.ioLseek(fd, off, whence), true
	return 0
}

func (m *Machine) ioCloseAsync(fd uint32) uint32 {
	f, ok := m.files[fd]
	if !ok {
		return errIoBadFd
	}
	f.async, f.hasAsync = 0, true
	// the close itself takes effect when the wait retrieves the result; keep
	// the descriptor until then so the wait can find it
	return 0
}

// ioWaitAsync writes the pending 64-bit result to *resPtr and returns 0. A
// descriptor closed via ioCloseAsync is dropped once its result is collected.
func (m *Machine) ioWaitAsync(fd, resPtr uint32) uint32 {
	f, ok := m.files[fd]
	if !ok {
		return errIoBadFd
	}
	if resPtr != 0 {
		m.write32(resPtr, uint32(f.async))
		m.write32(resPtr+4, uint32(f.async>>32))
	}
	f.hasAsync = false
	return 0
}

// ioPollAsync is the non-blocking form: it returns the result if one is pending
// (0), or 1 when there is nothing outstanding.
func (m *Machine) ioPollAsync(fd, resPtr uint32) uint32 {
	f, ok := m.files[fd]
	if !ok {
		return errIoBadFd
	}
	if !f.hasAsync {
		return 1 // no operation in progress
	}
	if resPtr != 0 {
		m.write32(resPtr, uint32(f.async))
		m.write32(resPtr+4, uint32(f.async>>32))
	}
	f.hasAsync = false
	return 0
}

// ioRead reads n bytes into guest memory at buf, returning the count read.
func (m *Machine) ioRead(fd, buf, n uint32) uint32 {
	f, ok := m.files[fd]
	if !ok {
		return errIoBadFd
	}
	p := make([]byte, n)
	got, err := m.vol.ReadFileAt(f.ent, int(f.pos), p)
	if err != nil {
		m.note("sceIoRead(%q at %d): %v", f.path, f.pos, err)
		return errIoBadFd
	}
	for i := 0; i < got; i++ {
		m.Write(buf+uint32(i), p[i])
	}
	f.pos += int64(got)
	m.note("sceIoRead(%q at %d, %d) -> %d", f.path, f.pos-int64(got), n, got)
	return uint32(got)
}

// ioWrite appends stdout/stderr writes to the TTY; file writes are refused (the
// UMD is read-only).
func (m *Machine) ioWrite(fd, buf, n uint32) uint32 {
	if fd == fdStdout || fd == fdStderr {
		for i := uint32(0); i < n; i++ {
			m.tty = append(m.tty, m.Read(buf+i))
		}
		return n
	}
	m.note("sceIoWrite to fd %d refused (read-only volume)", fd)
	return errIoBadFd
}

// fillStat writes a SceIoStat (0x58 bytes) for an entry. The umd9660 driver
// reports a file's start sector in st_private[0]; a game reads it there and opens
// the raw extent back with the "sce_lbn0x%X_size0x%X" path syntax.
func (m *Machine) fillStat(e Entry, stat uint32) {
	for i := uint32(0); i < 0x58; i += 4 {
		m.write32(stat+i, 0)
	}
	mode, attr := uint32(0x2000|0o777), uint32(0x0020) // FIO_S_IFREG, FIO_SO_IFREG
	if e.IsDir {
		mode, attr = 0x1000|0o777, 0x0010 // FIO_S_IFDIR, FIO_SO_IFDIR
	}
	m.write32(stat+0x00, mode)
	m.write32(stat+0x04, attr)
	m.write32(stat+0x08, uint32(e.Size)) // st_size (s64)
	m.write32(stat+0x0C, 0)
	m.write32(stat+0x40, uint32(e.Block)) // st_private[0] = start LBN
}

// ioGetstat fills a SceIoStat for a volume path.
func (m *Machine) ioGetstat(path string, stat uint32) uint32 {
	if m.vol == nil {
		return errIoNoEnt
	}
	vp := devicePath(path)
	e, err := m.vol.resolve(vp)
	if err != nil {
		m.note("sceIoGetstat(%q): not found", path)
		return errIoNoEnt
	}
	m.fillStat(e, stat)
	return 0
}

// --- Directory enumeration --------------------------------------------------
//
// Games catalogue the disc with sceIoDopen/Dread/Dclose (Burnout Legends builds
// its file table — names AND sizes — from a directory scan, then opens files by
// name and reads exactly the catalogued size). A directory descriptor holds the
// entry list; each sceIoDread fills one SceIoDirent (SceIoStat + 256-byte name +
// d_private + dummy = 0x168 bytes) and the return value is the number of entries
// still to read (>0), 0 once exhausted.

func (m *Machine) ioDopen(path string) uint32 {
	if m.vol == nil {
		return errIoNoEnt
	}
	vp := devicePath(path)
	ents, err := m.vol.ReadDir(vp)
	if err != nil {
		m.note("sceIoDopen(%q): %v", path, err)
		return errIoNoEnt
	}
	fd := m.nextFd
	m.nextFd++
	m.files[fd] = &ioFile{path: vp, dir: ents}
	m.note("sceIoDopen(%q) -> fd %d (%d entries)", path, fd, len(ents))
	return fd
}

func (m *Machine) ioDread(fd, dirent uint32) uint32 {
	f, ok := m.files[fd]
	if !ok || f.dir == nil {
		return errIoBadFd
	}
	left := len(f.dir) - f.dirPos
	if left <= 0 {
		return 0
	}
	e := f.dir[f.dirPos]
	f.dirPos++
	m.fillStat(e, dirent)
	name := e.Name
	if i := strings.IndexByte(name, ';'); i >= 0 {
		name = name[:i] // strip the ISO version suffix (";1")
	}
	for i := 0; i < 256; i++ {
		var b byte
		if i < len(name) {
			b = name[i]
		}
		m.Write(dirent+0x58+uint32(i), b)
	}
	m.write32(dirent+0x158, 0) // d_private
	m.write32(dirent+0x15C, 0)
	m.note("sceIoDread(fd %d) -> %q (size %d)", fd, name, e.Size)
	return uint32(left)
}

func (m *Machine) ioDclose(fd uint32) uint32 {
	f, ok := m.files[fd]
	if !ok || f.dir == nil {
		return errIoBadFd
	}
	delete(m.files, fd)
	return 0
}

// ioLseek repositions fd and returns the new position (-1 cast on error).
func (m *Machine) ioLseek(fd uint32, offset int64, whence uint32) int64 {
	f, ok := m.files[fd]
	if !ok {
		return -1
	}
	switch whence {
	case 0: // SEEK_SET
		f.pos = offset
	case 1: // SEEK_CUR
		f.pos += offset
	case 2: // SEEK_END
		f.pos = int64(f.ent.Size) + offset
	}
	if f.pos < 0 {
		f.pos = 0
	}
	return f.pos
}
