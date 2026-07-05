package dos

import (
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/x86"
)

// handleInt is the CPU's IntHook: it services software interrupts. Returning
// true means the interrupt was fully emulated (the CPU just continues past it).
func (m *Machine) handleInt(c *x86.CPU, n byte) bool {
	switch n {
	case 0x20: // terminate program
		m.Terminated = true
		c.Halt("INT 20h — program terminate")
		return true
	case 0x21:
		return m.int21(c)
	case 0x3F: // Microsoft C run-time overlay manager
		return m.overlayInt(c)
	case 0x67: // LIM EMS (expanded memory) driver
		return m.int67(c)
	case 0x2F: // multiplex — report "not installed"
		c.SetReg8(x86.AL, 0)
		c.CF = false
		return true
	case 0x10: // video BIOS
		return m.int10(c)
	case 0x33: // mouse driver
		return m.int33(c)
	default:
		// Keyboard/timer/mouse BIOS etc.: no-op so the boot proceeds. Log only
		// the first occurrence of each vector to avoid flooding.
		m.otherInts[n]++
		if m.otherInts[n] == 1 {
			m.logf("INT %02Xh (ignored) AX=%04X at %04X:%04X", n, c.Reg16(x86.AX), c.Seg[x86.CS], c.IP)
		}
		c.CF = false
		return true
	}
}

// int21 services the DOS INT 21h function in AH.
func (m *Machine) int21(c *x86.CPU) bool {
	ah := c.Reg8(x86.AH)
	m.IntCounts[ah]++
	c.CF = false

	switch ah {
	case 0x00, 0x4C: // terminate
		m.Terminated = true
		if ah == 0x4C {
			m.ExitCode = c.Reg8(x86.AL)
		}
		c.Halt("INT 21h/%02X — program exit (code %d)", ah, m.ExitCode)
		return true

	case 0x02: // display character in DL
		return true
	case 0x06, 0x07, 0x08: // direct/raw console I/O — no input available
		c.SetReg8(x86.AL, 0)
		return true
	case 0x09: // print $-terminated string at DS:DX
		m.logf("DOS print: %q", m.dollarStr(c.Seg[x86.DS], c.Reg16(x86.DX)))
		return true

	case 0x19: // get current default drive
		c.SetReg8(x86.AL, 2) // C:
		return true
	case 0x1A: // set DTA = DS:DX
		m.dtaSeg, m.dtaOff = c.Seg[x86.DS], c.Reg16(x86.DX)
		return true
	case 0x2F: // get DTA -> ES:BX
		c.Seg[x86.ES] = m.dtaSeg
		c.SetReg16(x86.BX, m.dtaOff)
		return true
	case 0x4E: // FindFirst: DS:DX = filespec, CX = attribute mask
		return m.findFirst(c)
	case 0x4F: // FindNext
		return m.findNext(c)
	case 0x25: // set interrupt vector AL -> DS:DX
		v := uint32(c.Reg8(x86.AL)) * 4
		m.w16(v, c.Reg16(x86.DX))
		m.w16(v+2, c.Seg[x86.DS])
		return true
	case 0x2A: // get date
		c.SetReg16(x86.CX, 1992)
		c.SetReg8(x86.DH, 3)
		c.SetReg8(x86.DL, 11)
		c.SetReg8(x86.AL, 3)
		return true
	case 0x2C: // get time
		c.SetReg16(x86.CX, 0)
		c.SetReg16(x86.DX, 0)
		return true
	case 0x30: // get DOS version
		c.SetReg8(x86.AL, 5) // 5.0
		c.SetReg8(x86.AH, 0)
		c.SetReg16(x86.BX, 0)
		c.SetReg16(x86.CX, 0)
		return true
	case 0x33: // ctrl-break check
		c.SetReg8(x86.DL, 0)
		return true
	case 0x2B: // set date — accept
		c.SetReg8(x86.AL, 0)
		return true
	case 0x2D: // set time — accept
		c.SetReg8(x86.AL, 0)
		return true
	case 0x36: // get free disk space (drive DL): report plenty
		c.SetReg16(x86.AX, 8)      // sectors per cluster
		c.SetReg16(x86.BX, 0x8000) // free clusters
		c.SetReg16(x86.CX, 512)    // bytes per sector
		c.SetReg16(x86.DX, 0xFFFF) // total clusters
		return true
	case 0x3B: // chdir — accept (cwd is not modeled; paths resolve from the game dir)
		return true
	case 0x41: // delete file (only scratch files are deletable)
		name := m.asciiz(c.Seg[x86.DS], c.Reg16(x86.DX))
		if sp := m.scratchPath(name); sp != "" {
			os.Remove(sp)
		}
		return true
	case 0x47: // get current directory -> DS:SI = "" (root)
		m.Mem[lin(c.Seg[x86.DS], c.Reg16(x86.SI))] = 0
		return true
	case 0x35: // get interrupt vector AL -> ES:BX
		v := uint32(c.Reg8(x86.AL)) * 4
		c.SetReg16(x86.BX, m.r16(v))
		c.Seg[x86.ES] = m.r16(v + 2)
		return true
	case 0x52: // get list of lists -> ES:BX; first MCB segment is the word at [BX-2]
		m.w16(uint32(0x50)<<4+0x1E, m.firstMCB)
		c.Seg[x86.ES] = 0x50
		c.SetReg16(x86.BX, 0x20)
		return true
	case 0x50, 0x51, 0x62: // set/get PSP
		if ah == 0x50 {
			m.pspSeg = c.Reg16(x86.BX)
		} else {
			c.SetReg16(x86.BX, m.pspSeg)
		}
		return true

	case 0x3C: // create/truncate file -> the scratch dir (never touches game data)
		name := m.asciiz(c.Seg[x86.DS], c.Reg16(x86.DX))
		sp := m.scratchPath(name)
		if sp == "" {
			return m.dosErr(c, 3)
		}
		os.MkdirAll(filepath.Dir(sp), 0o755)
		f, err := os.Create(sp)
		if err != nil {
			return m.dosErr(c, 3)
		}
		h := m.allocFH()
		m.files[h] = f
		m.logf("create %q -> scratch %s (handle %d)", name, filepath.Base(sp), h)
		c.SetReg16(x86.AX, h)
		return true
	case 0x43: // get/set file attributes
		name := m.asciiz(c.Seg[x86.DS], c.Reg16(x86.DX))
		if c.Reg8(x86.AL) == 1 { // set attributes — accept
			return true
		}
		if _, ok := m.resolveFile(name); ok { // get attributes
			c.SetReg16(x86.CX, 0x20) // archive
			return true
		}
		return m.dosErr(c, 2) // file not found
	case 0x3D: // open existing file, DS:DX = name
		name := m.asciiz(c.Seg[x86.DS], c.Reg16(x86.DX))
		host, ok := m.resolveFile(name)
		if !ok {
			m.logf("open FAILED %q", name)
			return m.dosErr(c, 2) // file not found
		}
		f, err := os.Open(host)
		if err != nil {
			return m.dosErr(c, 2)
		}
		h := m.allocFH()
		m.files[h] = f
		m.logf("open %q -> handle %d", name, h)
		c.SetReg16(x86.AX, h)
		return true
	case 0x3E: // close handle BX
		h := c.Reg16(x86.BX)
		if f := m.files[h]; f != nil {
			f.Close()
			delete(m.files, h)
		}
		return true
	case 0x3F: // read CX bytes from handle BX into DS:DX
		h := c.Reg16(x86.BX)
		f := m.files[h]
		if f == nil {
			return m.dosErr(c, 6) // invalid handle
		}
		n := int(c.Reg16(x86.CX))
		pos, _ := f.Seek(0, 1)
		buf := make([]byte, n)
		got, _ := f.Read(buf)
		dst := lin(c.Seg[x86.DS], c.Reg16(x86.DX))
		for i := 0; i < got; i++ {
			m.Mem[(dst+uint32(i))&0xFFFFF] = buf[i]
		}
		m.logf("read handle %d: %d/%d bytes from file $%X -> %04X:%04X (lin $%X)",
			h, got, n, pos, c.Seg[x86.DS], c.Reg16(x86.DX), dst)
		c.SetReg16(x86.AX, uint16(got))
		return true
	case 0x40: // write CX bytes from DS:DX to handle BX
		h := c.Reg16(x86.BX)
		n := int(c.Reg16(x86.CX))
		src := lin(c.Seg[x86.DS], c.Reg16(x86.DX))
		buf := make([]byte, n)
		for i := 0; i < n; i++ {
			buf[i] = m.Mem[(src+uint32(i))&0xFFFFF]
		}
		if h == 1 || h == 2 { // stdout/stderr
			m.logf("DOS write(fd%d): %q", h, string(buf))
		} else if f := m.files[h]; f != nil {
			f.Write(buf)
		}
		c.SetReg16(x86.AX, uint16(n))
		return true
	case 0x42: // lseek handle BX, CX:DX by method AL
		h := c.Reg16(x86.BX)
		f := m.files[h]
		if f == nil {
			return m.dosErr(c, 6)
		}
		// CX:DX is a SIGNED 32-bit offset (negative seeks are legal with whence 1/2).
		off := int64(int32(uint32(c.Reg16(x86.CX))<<16 | uint32(c.Reg16(x86.DX))))
		whence := int(c.Reg8(x86.AL))
		pos, err := f.Seek(off, whence)
		if err != nil {
			return m.dosErr(c, 0x19)
		}
		m.logf("seek handle %d: whence %d off %d -> $%X", h, whence, off, pos)
		c.SetReg16(x86.AX, uint16(pos))
		c.SetReg16(x86.DX, uint16(pos>>16))
		return true
	case 0x44: // ioctl — report "regular file, not a device"
		c.SetReg16(x86.DX, 0)
		return true

	case 0x48: // allocate BX paragraphs
		seg, largest, ok := m.allocBlock(c.Reg16(x86.BX))
		if !ok {
			c.SetReg16(x86.BX, largest) // largest block available
			return m.dosErr(c, 8)       // insufficient memory
		}
		c.SetReg16(x86.AX, seg)
		return true
	case 0x49: // free memory block at ES
		if !m.freeBlock(c.Seg[x86.ES]) {
			return m.dosErr(c, 9) // invalid block address
		}
		return true
	case 0x4A: // resize the block at ES to BX paragraphs
		max, ok := m.resizeBlock(c.Seg[x86.ES], c.Reg16(x86.BX))
		if !ok {
			c.SetReg16(x86.BX, max) // largest size it could grow to
			return m.dosErr(c, 8)   // insufficient memory
		}
		return true

	default:
		m.logf("UNHANDLED INT 21h AH=%02X AL=%02X BX=%04X CX=%04X DX=%04X at %04X:%04X",
			ah, c.Reg8(x86.AL), c.Reg16(x86.BX), c.Reg16(x86.CX), c.Reg16(x86.DX), c.Seg[x86.CS], c.IP)
		return true
	}
}

// overlayInt handles the Microsoft C run-time overlay call (INT 3Fh). Rather
// than reimplement the overlay manager, we let the game's *own* handler run:
// the CRT installed it in the IVT at boot (IVT[0x3F] -> 3AB0:04E4), and it does
// the real work — read the thunk after the CD 3F, index the resident overlay
// directory, load the overlay from UW.EXE on demand (via the INT 21h file
// services this machine already provides), relocate it, and transfer in. We
// only note the call (the 5-byte thunk is `CD 3F | DW entry | DB overlay`),
// then return false so the CPU dispatches through the real vector.
func (m *Machine) overlayInt(c *x86.CPU) bool {
	if m.OverlayCalls < 32 || m.OverlayCalls%500 == 0 {
		cs, ip := c.Seg[x86.CS], uint16(c.IP)
		a := lin(cs, ip)
		m.logf("INT 3Fh #%d: overlay %d entry $%04X (thunk %04X:%04X)",
			m.OverlayCalls, m.Mem[(a+2)&0xFFFFF], m.r16(a), cs, ip)
	}
	m.OverlayCalls++
	return false // dispatch to the game's real INT 3Fh handler via the IVT
}

// findFirst services INT 21h/4Eh: match DS:DX filespec (with * and ? wildcards)
// against the game directory and fill the DTA with the first result.
func (m *Machine) findFirst(c *x86.CPU) bool {
	spec := m.asciiz(c.Seg[x86.DS], c.Reg16(x86.DX))
	norm := strings.ReplaceAll(spec, "\\", "/")
	dir, pat := m.gameDir, norm
	if i := strings.LastIndex(norm, "/"); i >= 0 {
		pat = norm[i+1:]
		if dirPart := norm[:i]; dirPart != "" && dirPart != "." {
			host, ok := m.walkPath(dirPart)
			if !ok {
				m.logf("FindFirst %q -> PATH NOT FOUND (dir %q)", spec, dirPart)
				return m.dosErr(c, 3) // path not found — the directory doesn't exist
			}
			dir = host
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return m.dosErr(c, 3) // path not found
	}
	attrMask := c.Reg16(x86.CX)
	st := &findState{}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue // host dotfiles (.DS_Store, …) don't exist on a DOS filesystem
		}
		// Directories match only when the search mask requests them (bit 4);
		// normal files always match.
		if e.IsDir() && attrMask&0x10 == 0 {
			continue
		}
		if dosMatch(pat, e.Name()) {
			st.matches = append(st.matches, filepath.Join(dir, e.Name()))
		}
	}
	// Key the search state by the DTA address (as real DOS keeps it in the DTA),
	// so interleaved/nested searches with distinct DTAs don't clobber each other.
	m.finds[lin(m.dtaSeg, m.dtaOff)] = st
	m.logf("FindFirst %q -> %d match(es)", spec, len(st.matches))
	return m.findNext(c)
}

// findNext services INT 21h/4Fh: emit the next queued match into the DTA.
func (m *Machine) findNext(c *x86.CPU) bool {
	st := m.finds[lin(m.dtaSeg, m.dtaOff)]
	if st == nil || st.idx >= len(st.matches) {
		if st == nil {
			m.logf("FindNext at DTA %04X:%04X -> NO STATE", m.dtaSeg, m.dtaOff)
		} else {
			m.logf("FindNext at DTA %04X:%04X -> exhausted (%d/%d)", m.dtaSeg, m.dtaOff, st.idx, len(st.matches))
		}
		return m.dosErr(c, 18) // no more files
	}
	host := st.matches[st.idx]
	st.idx++
	info, err := os.Stat(host)
	if err != nil {
		return m.dosErr(c, 18)
	}
	dta := lin(m.dtaSeg, m.dtaOff)
	// 43-byte find record: [0..20] reserved, 21 attr, 22 time, 24 date,
	// 26 size(dword), 30 name(ASCIIZ 8.3).
	attr := byte(0x20) // archive (regular file)
	if info.IsDir() {
		attr = 0x10 // directory
	}
	m.Mem[dta+0x15] = attr
	m.w16(dta+0x16, 0)    // time
	m.w16(dta+0x18, 0x21) // date (1980-01-01-ish)
	sz := uint32(info.Size())
	m.w16(dta+0x1A, uint16(sz))
	m.w16(dta+0x1C, uint16(sz>>16))
	name := strings.ToUpper(filepath.Base(host))
	for i := 0; i < 13; i++ {
		if i < len(name) {
			m.Mem[dta+0x1E+uint32(i)] = name[i]
		} else {
			m.Mem[dta+0x1E+uint32(i)] = 0
		}
	}
	c.CF = false
	return true
}

// dosMatch reports whether an 8.3 filename matches a DOS wildcard pattern
// (case-insensitive; * matches the rest of the name or extension, ? one char).
func dosMatch(pattern, name string) bool {
	p := strings.ToUpper(pattern)
	n := strings.ToUpper(name)
	if p == "*.*" || p == "*" || p == "" {
		return true
	}
	return globSegments(p, n)
}

// globSegments matches p against n with * (any run) and ? (any one char).
func globSegments(p, n string) bool {
	pi, ni := 0, 0
	star, mark := -1, 0
	for ni < len(n) {
		if pi < len(p) && (p[pi] == '?' || p[pi] == n[ni]) {
			pi++
			ni++
		} else if pi < len(p) && p[pi] == '*' {
			star = pi
			mark = ni
			pi++
		} else if star != -1 {
			pi = star + 1
			mark++
			ni = mark
		} else {
			return false
		}
	}
	for pi < len(p) && p[pi] == '*' {
		pi++
	}
	return pi == len(p)
}

// dosErr sets the DOS error return (CF set, AX = error code) and returns true.
func (m *Machine) dosErr(c *x86.CPU, code uint16) bool {
	c.CF = true
	c.SetReg16(x86.AX, code)
	return true
}

// dollarStr reads a '$'-terminated string (INT 21h/09h) from seg:off.
func (m *Machine) dollarStr(seg, off uint16) string {
	var b []byte
	a := lin(seg, off)
	for i := 0; i < 256; i++ {
		ch := m.Mem[(a+uint32(i))&0xFFFFF]
		if ch == '$' {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}
