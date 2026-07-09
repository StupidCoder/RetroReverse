package n64

// isviewer.go models the IS-Viewer, a development cartridge that gave the N64 a
// way to print. It occupies a window in the cartridge's address space that a
// retail game never touches, so modelling it costs a running game nothing.
//
// It matters because test ROMs use it. n64-systemtest draws its results on the
// screen *and* writes them here, which turns "read the framebuffer and squint"
// into "read a string". A conformance suite whose output must be recognised from
// pixels is a conformance suite nobody runs.
//
// The protocol is two registers. Text is written a byte at a time into a buffer
// at 0x13FF0020, and writing its length to 0x13FF0014 prints it.

const (
	isvBase   = 0x13FF0000
	isvLen    = 0x13FF0014
	isvBuffer = 0x13FF0020
	isvEnd    = 0x13FF0220
)

// isv holds the print buffer and the lines drained out of it. Fields are
// exported so encoding/gob carries them into a save-state.
type isv struct {
	Buf   [isvEnd - isvBuffer]byte
	Lines []string
}

// inISViewer reports whether a physical address belongs to the IS-Viewer window.
// It must be tested before the cartridge, which otherwise claims the whole range.
func inISViewer(addr uint32) bool {
	return addr >= isvBase && addr < isvEnd
}

func (m *Machine) isvRead(addr uint32) uint32 {
	if addr >= isvBuffer {
		off := addr - isvBuffer
		return uint32(m.isv.Buf[off])<<24 | uint32(m.isv.Buf[off+1])<<16 |
			uint32(m.isv.Buf[off+2])<<8 | uint32(m.isv.Buf[off+3])
	}
	return 0
}

func (m *Machine) isvWrite(addr uint32, v uint32) {
	switch {
	case addr == isvLen:
		n := int(v)
		if n > len(m.isv.Buf) {
			n = len(m.isv.Buf)
		}
		m.isvPrint(string(m.isv.Buf[:n]))
	case addr >= isvBuffer && addr+3 < isvEnd:
		off := addr - isvBuffer
		m.isv.Buf[off] = byte(v >> 24)
		m.isv.Buf[off+1] = byte(v >> 16)
		m.isv.Buf[off+2] = byte(v >> 8)
		m.isv.Buf[off+3] = byte(v)
	}
}

// isvWriteByte serves the byte-wide stores the ROM actually uses.
func (m *Machine) isvWriteByte(addr uint32, v byte) {
	if addr >= isvBuffer && addr < isvEnd {
		m.isv.Buf[addr-isvBuffer] = v
	}
}

// isvPrint splits the printed text into lines and hands each to the caller.
func (m *Machine) isvPrint(s string) {
	line := ""
	for _, c := range s {
		if c == '\n' || c == '\r' {
			if line != "" {
				m.isvEmit(line)
				line = ""
			}
			continue
		}
		line += string(c)
	}
	if line != "" {
		m.isvEmit(line)
	}
}

func (m *Machine) isvEmit(line string) {
	m.isv.Lines = append(m.isv.Lines, line)
	if m.OnPrint != nil {
		m.OnPrint(m, line)
	}
}

// ISViewerLines returns everything the running program has printed.
func (m *Machine) ISViewerLines() []string { return m.isv.Lines }
