package psp

// io.go is the IoFileMgrForUser HLE: sceIoOpen/Read/Lseek/Close backed by the
// mounted UMD volume, plus the stdout/stderr descriptors feeding the TTY. Game
// paths name the disc device ("disc0:/PSP_GAME/..." or "umd0:/..."); the device
// prefix is stripped and the rest resolved on the ISO 9660 volume.

import "strings"

// ioFile is one open descriptor on the UMD volume.
type ioFile struct {
	path string // volume path (device prefix stripped)
	ent  Entry
	pos  int64
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
	return fd
}

func (m *Machine) ioClose(fd uint32) uint32 {
	if _, ok := m.files[fd]; !ok {
		return errIoBadFd
	}
	delete(m.files, fd)
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
