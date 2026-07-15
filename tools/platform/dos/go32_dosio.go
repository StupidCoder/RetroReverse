package dos

// DOS file I/O for the go32 bridge. A DJGPP program does all its file access
// through the C library, which lowers each fopen/fread/fwrite to a real-mode INT
// 21h issued via DPMI 0300h: the filename and the data ride in the conventional-
// memory transfer buffer, and DS:DX in the RMCS point at it. Because our
// conventional arena is identity-mapped (seg<<4 == linear base, see go32.go), the
// buffer bytes are just p.Mem[DS<<4 + DX], so these handlers read/write it
// directly. Handles are the host *os.File behind a DOS handle number; the files
// resolve case-insensitively under gameDir, since a DOS image asks for PAK0.PAK
// or id1/pak0.pak against a host tree that may be lower-case.

import (
	"os"
	"path/filepath"
	"strings"
)

// rmLinear folds a real-mode seg:off (as it appears in an RMCS) to a flat linear
// address in the identity-mapped conventional arena.
func (p *PM) rmLinear(seg uint16, off uint32) uint32 {
	return uint32(seg)<<4 + (off & 0xFFFF)
}

// dosErr marks the RMCS as a failed INT 21h: CF set, AX = DOS error code. It
// returns true because the function *was* serviced (with an error result), so the
// caller stops looking for an unmodelled handler.
func (p *PM) dosErr(r *rmcs, code uint16) bool {
	r.flags |= 0x0001
	r.eax = (r.eax & 0xFFFF0000) | uint32(code)
	return true
}

// setAX / setDX write the 16-bit result registers, preserving the high halves.
func setAX(r *rmcs, v uint16) { r.eax = (r.eax & 0xFFFF0000) | uint32(v) }
func setDX(r *rmcs, v uint16) { r.edx = (r.edx & 0xFFFF0000) | uint32(v) }

// allocHandle returns the lowest free DOS handle >= 5 (0..4 are the standard
// stdin/stdout/stderr/aux/prn). DOS hands out the lowest free handle, and some
// programs depend on that ordering.
func (p *PM) allocHandle() uint16 {
	for h := uint16(5); h < 255; h++ {
		if _, ok := p.files[h]; !ok {
			return h
		}
	}
	return 0xFFFF
}

// resolveHostPath maps a DOS path (backslashes, optional drive letter, any case)
// to a real host path under gameDir, matching each component case-insensitively.
// If a component can't be matched it returns the naive join, so os.Open then fails
// with a genuine file-not-found the caller reports back to the game.
func (p *PM) resolveHostPath(name string) string {
	base := p.gameDir
	if base == "" {
		base = "."
	}
	name = strings.ReplaceAll(name, "\\", "/")
	if len(name) >= 2 && name[1] == ':' { // strip drive letter
		name = name[2:]
	}
	name = strings.TrimLeft(name, "/")
	naive := filepath.Join(base, filepath.FromSlash(name))
	if _, err := os.Stat(naive); err == nil {
		return naive
	}
	cur := base
	for _, comp := range strings.Split(name, "/") {
		if comp == "" || comp == "." {
			continue
		}
		next := filepath.Join(cur, comp)
		if _, err := os.Stat(next); err == nil {
			cur = next
			continue
		}
		found := ""
		if ents, err := os.ReadDir(cur); err == nil {
			for _, e := range ents {
				if strings.EqualFold(e.Name(), comp) {
					found = e.Name()
					break
				}
			}
		}
		if found == "" {
			return naive // let the caller's Open produce the error
		}
		cur = filepath.Join(cur, found)
	}
	return cur
}

// dosFile services the file-I/O INT 21h functions. It returns false only for a
// function it does not implement, so the bridge halts with the concrete AH.
func (p *PM) dosFile(r *rmcs) bool {
	switch byte(r.eax >> 8) {
	case 0x3D:
		return p.dosOpen(r)
	case 0x3C:
		return p.dosCreate(r)
	case 0x3E:
		return p.dosClose(r)
	case 0x3F:
		return p.dosRead(r)
	case 0x40:
		return p.dosWrite(r)
	case 0x42:
		return p.dosSeek(r)
	case 0x44:
		return p.dosIoctl(r)
	case 0x41: // Delete file
		os.Remove(p.resolveHostPath(p.asciiz(p.rmLinear(r.ds, r.edx))))
		return true
	case 0x43: // Get/set file attributes -> report a plain file (attr 0x20)
		if byte(r.eax) == 0 {
			setAX(r, 0x20)
		}
		return true
	case 0x47: // Get current directory -> DS:SI gets ASCIIZ path (no drive/leading \)
		p.Write(p.rmLinear(r.ds, r.esi), 0) // "" = the root of the (virtual) drive
		setAX(r, 0x0100)
		return true
	case 0x4E, 0x4F: // Find first / find next -> "no more files"
		return p.dosErr(r, 18)
	default:
		return false
	}
}

func (p *PM) dosOpen(r *rmcs) bool {
	name := p.asciiz(p.rmLinear(r.ds, r.edx))
	host := p.resolveHostPath(name)
	var f *os.File
	var err error
	switch byte(r.eax) & 3 { // AL access mode: 0=read 1=write 2=read/write
	case 1:
		f, err = os.OpenFile(host, os.O_WRONLY, 0644)
	case 2:
		f, err = os.OpenFile(host, os.O_RDWR, 0644)
	default:
		f, err = os.Open(host)
	}
	if err != nil {
		p.logf("DOS open %q FAILED (%v)", name, err)
		return p.dosErr(r, 2) // file not found
	}
	h := p.allocHandle()
	p.files[h] = f
	setAX(r, h)
	p.logf("DOS open %q -> handle %d", name, h)
	return true
}

func (p *PM) dosCreate(r *rmcs) bool {
	name := p.asciiz(p.rmLinear(r.ds, r.edx))
	host := p.resolveHostPath(name)
	f, err := os.OpenFile(host, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		p.logf("DOS create %q FAILED (%v)", name, err)
		return p.dosErr(r, 5) // access denied
	}
	h := p.allocHandle()
	p.files[h] = f
	setAX(r, h)
	p.logf("DOS create %q -> handle %d", name, h)
	return true
}

func (p *PM) dosClose(r *rmcs) bool {
	h := uint16(r.ebx)
	if f, ok := p.files[h]; ok {
		f.Close()
		delete(p.files, h)
		return true
	}
	if h <= 4 { // closing a standard handle: harmless success
		return true
	}
	return p.dosErr(r, 6) // invalid handle
}

func (p *PM) dosRead(r *rmcs) bool {
	h := uint16(r.ebx)
	n := int(r.ecx & 0xFFFF)
	dst := p.rmLinear(r.ds, r.edx)
	f, ok := p.files[h]
	if !ok {
		if h == 0 { // stdin -> immediate EOF
			setAX(r, 0)
			return true
		}
		return p.dosErr(r, 6)
	}
	buf := make([]byte, n)
	got := 0
	for got < n {
		m, err := f.Read(buf[got:])
		got += m
		if err != nil {
			break // EOF or error -> short read, which DOS reports via AX
		}
	}
	for i := 0; i < got; i++ {
		p.Write(dst+uint32(i), buf[i])
	}
	setAX(r, uint16(got))
	return true
}

func (p *PM) dosWrite(r *rmcs) bool {
	h := uint16(r.ebx)
	n := int(r.ecx & 0xFFFF)
	src := p.rmLinear(r.ds, r.edx)
	data := make([]byte, n)
	for i := 0; i < n; i++ {
		data[i] = p.Read(src + uint32(i))
	}
	switch h {
	case 1, 2: // stdout / stderr -> the game's console output
		p.Console = append(p.Console, data...)
		setAX(r, uint16(n))
		return true
	}
	f, ok := p.files[h]
	if !ok {
		return p.dosErr(r, 6)
	}
	m, _ := f.Write(data)
	setAX(r, uint16(m))
	return true
}

func (p *PM) dosSeek(r *rmcs) bool {
	h := uint16(r.ebx)
	f, ok := p.files[h]
	if !ok {
		return p.dosErr(r, 6)
	}
	off := int64(int32(uint32(r.ecx&0xFFFF)<<16 | uint32(r.edx&0xFFFF)))
	pos, err := f.Seek(off, int(byte(r.eax))) // AL: 0=SEEK_SET 1=CUR 2=END
	if err != nil {
		return p.dosErr(r, 25)
	}
	setAX(r, uint16(pos))
	setDX(r, uint16(pos>>16))
	return true
}

// dosIoctl handles AH=44h. Only the device-info subfunctions matter here: DJGPP's
// isatty() reads bit 7 of the returned info word to tell a console from a file.
func (p *PM) dosIoctl(r *rmcs) bool {
	h := uint16(r.ebx)
	switch byte(r.eax) { // AL subfunction
	case 0x00: // get device info
		if h <= 2 { // stdin/stdout/stderr -> character device (bit 7 set)
			setDX(r, 0x80D3)
		} else { // regular file on drive C: (bit 7 clear)
			setDX(r, 0x0002)
		}
		return true
	case 0x01: // set device info -> ignore
		return true
	default:
		return true
	}
}
