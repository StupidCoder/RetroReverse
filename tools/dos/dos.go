// Package dos is a reusable real-mode DOS/PC machine for running vintage
// MS-DOS games under the tools/x86 execution core as an oracle. It loads an MZ
// executable the way DOS would (place a PSP and environment, copy the load
// module, apply the relocations, seed the entry registers) and services the
// INT 21h / BIOS calls the program makes, so the game's real code runs and we
// can watch it. It is game-agnostic — game-specific setup (which save/work
// directories to seed, etc.) is configured by the caller — and was first proven
// on Ultima Underworld (see "Ultima Underworld (PC)/Ultima_Underworld.md"), but
// nothing here is UW-specific.
//
// The machine models: the MCB memory-control-block chain, LIM EMS 4.0
// (dos_ems.go), a VGA video BIOS and true planar VGA (dos_video.go/dos_vga.go),
// the INT 33h mouse (dos_mouse.go), the PC I/O ports and timer IRQ (dos_io.go),
// and scripted keyboard/mouse input injection (dos_keyboard.go). Unhandled INT
// functions are logged (so the next one to implement is obvious) and returned as
// success. It depends only on retroreverse.com/tools/x86.
package dos

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/x86"
)

// findState is one active FindFirst/FindNext search (its queued matches and
// cursor), keyed in the machine by the DTA address the search uses.
type findState struct {
	matches []string
	idx     int
}

// Machine is a loaded MZ program ready to run.
type Machine struct {
	Mem []byte // 1 MiB real-mode address space
	CPU *x86.CPU

	gameDir string
	files   map[uint16]*os.File

	pspSeg   uint16
	envSeg   uint16
	loadSeg  uint16
	memTop   uint16 // first paragraph past conventional memory (top of the MCB chain)
	firstMCB uint16 // segment of the first Memory Control Block

	dtaSeg, dtaOff uint16                // Disk Transfer Area (INT 21h/1Ah)
	finds          map[uint32]*findState // FindFirst/Next state, keyed by DTA address
	scratchDir     string                // writable temp dir for files the game creates (never the game dir)

	// instrumentation
	Log          []string       // notable events (INT calls, allocations, files)
	IntCounts    map[byte]int   // count of INT 21h function calls by AH
	otherInts    map[byte]int   // count of other software INTs by number
	OverlayCalls int            // INT 3Fh (Microsoft-C overlay manager) calls seen
	ems          *emsState      // LIM EMS (INT 67h) state
	io           *ioState       // minimal PC I/O-port models (keyboard/VGA/PIT)
	video        map[uint16]int // INT 10h function-call histogram (first-call logging)
	vga          *vgaState      // planar VGA memory + sequencer/GC/CRTC registers
	ms           *mouseState    // INT 33h mouse driver state (see dos_mouse.go)
	keyEvents    []injEvent     // scripted keyboard/mouse injection schedule (see dos_keyboard.go)
	keyWait      int            // timer-ticks left before the next scheduled event
	keyHits      int            // input events delivered so far (for first-N logging)
	EnableIRQ    bool           // inject periodic timer IRQ0 (drives frame waits, cutscenes, menus)
	WatchAddr    uint32         // if WatchLen>0, log writes to [WatchAddr, WatchAddr+WatchLen) (debugging)
	WatchLen     uint32
	watchHits    int
	Terminated   bool
	ExitCode     byte
}

// LoadEXE reads the MZ executable at exePath and prepares a Machine whose data
// files resolve under gameDir.
func LoadEXE(exePath, gameDir string) (*Machine, error) {
	data, err := os.ReadFile(exePath)
	if err != nil {
		return nil, err
	}
	mz, err := ParseMZ(data)
	if err != nil {
		return nil, err
	}
	if mz.LoadImageEnd > len(data) || mz.LoadModuleOffset > mz.LoadImageEnd {
		return nil, fmt.Errorf("mz: load image %#x..%#x outside file (%d bytes)", mz.LoadModuleOffset, mz.LoadImageEnd, len(data))
	}

	m := &Machine{
		Mem:       make([]byte, 1<<20),
		gameDir:   gameDir,
		files:     map[uint16]*os.File{},
		envSeg:    0x0080,
		pspSeg:    0x0100,
		memTop:    0xA000, // 640 KiB conventional memory
		IntCounts: map[byte]int{},
		otherInts: map[byte]int{},
		finds:     map[uint32]*findState{},
	}
	m.loadSeg = m.pspSeg + 0x10 // PSP is 256 bytes = 0x10 paragraphs

	// Copy the load module into memory at the load segment.
	module := data[mz.LoadModuleOffset:mz.LoadImageEnd]
	copy(m.Mem[uint32(m.loadSeg)<<4:], module)

	// Apply relocations: add the load segment to each fix-up word. (This is what
	// turns the link-time DGROUP immediate, e.g. MOV DX,$5C0F at the entry, into
	// the correct runtime data segment.)
	for _, r := range mz.Relocs {
		lin := (uint32(m.loadSeg)+uint32(r.Segment))<<4 + uint32(r.Offset)
		w := binary.LittleEndian.Uint16(m.Mem[lin&0xFFFFF:]) + m.loadSeg
		binary.LittleEndian.PutUint16(m.Mem[lin&0xFFFFF:], w)
	}

	m.setupBIOS()
	m.setupPSP()
	m.setupEnv(exePath)
	m.setupMCB() // the DOS memory-control-block chain (program owns all free memory)
	m.setupEMS()
	m.scratchDir, _ = os.MkdirTemp("", "dosrun-scratch-")

	c := x86.NewCPU(m)
	c.Seg[x86.CS] = m.loadSeg + mz.InitCS
	c.IP = uint32(mz.InitIP)
	c.Seg[x86.SS] = m.loadSeg + mz.InitSS
	c.SetReg16(x86.SP, mz.InitSP)
	c.Seg[x86.DS] = m.pspSeg
	c.Seg[x86.ES] = m.pspSeg
	m.dtaSeg, m.dtaOff = m.pspSeg, 0x80 // default DTA is PSP:0080
	c.IF = true
	c.IntHook = m.handleInt
	m.io = &ioState{}
	m.vga = &vgaState{}
	c.PortIn = m.portIn
	c.PortOut = m.portOut
	c.OnStep = m.onStep // periodic timer-IRQ injection
	m.CPU = c
	return m, nil
}

// --- x86.Bus ---

func (m *Machine) Read(a uint32) byte {
	a &= 0xFFFFF
	if a >= 0xA0000 && a < 0xB0000 {
		return m.vgaRead(a) // planar VGA window (see dos_vga.go)
	}
	return m.Mem[a]
}
func (m *Machine) Write(a uint32, v byte) {
	a &= 0xFFFFF
	if m.WatchLen > 0 && a >= m.WatchAddr && a < m.WatchAddr+m.WatchLen && m.CPU != nil {
		m.watchHits++
		if m.watchHits <= 60 {
			m.logf("WATCH: write $%02X to %05X at %04X:%04X", v, a, m.CPU.Seg[1], m.CPU.IP)
		}
	}
	if a >= 0xA0000 && a < 0xB0000 {
		m.vgaWrite(a, v) // planar VGA window (see dos_vga.go)
		return
	}
	m.Mem[a] = v
}

func (m *Machine) r16(lin uint32) uint16 { return binary.LittleEndian.Uint16(m.Mem[lin&0xFFFFF:]) }
func (m *Machine) w16(lin uint32, v uint16) {
	binary.LittleEndian.PutUint16(m.Mem[lin&0xFFFFF:], v)
}
func lin(seg, off uint16) uint32 { return ((uint32(seg) << 4) + uint32(off)) & 0xFFFFF }

// setupBIOS initialises the interrupt-vector table and BIOS Data Area the way a
// PC BIOS would before handing control to a program. Crucially, every default
// vector points at a harmless IRET stub (not the zero it would otherwise hold),
// so when the game *saves and chains* a vector — e.g. its timer ISR does
// `PUSHF; CALLF [saved INT 8]` to reach the old BIOS handler — the chain call
// lands on a valid instruction instead of far-jumping through a null pointer.
func (m *Machine) setupBIOS() {
	const biosSeg, biosOff = 0xF000, 0xFF53
	m.Mem[uint32(biosSeg)<<4+biosOff] = 0xCF // IRET
	for n := 0; n < 256; n++ {
		m.w16(uint32(n)*4, biosOff)
		m.w16(uint32(n)*4+2, biosSeg)
	}
	// BIOS Data Area (segment 0x40): the handful of fields a program may read.
	bda := uint32(0x40) << 4
	m.w16(bda+0x10, 0x0021) // equipment word: 80x25 colour, one floppy
	m.w16(bda+0x13, 640)    // conventional memory size, KiB
	m.Mem[bda+0x49] = 0x03  // current video mode = 80x25 colour text
	m.w16(bda+0x4A, 80)     // text columns
	m.w16(bda+0x63, 0x03D4) // CRTC base I/O port (colour)
	m.Mem[bda+0x84] = 24    // text rows - 1
	m.Mem[bda+0x85] = 16    // character height
	// The timer-tick dword at 0x6C starts at 0 and is advanced by onStep.
}

// setupPSP writes the fields of the Program Segment Prefix the startup reads.
func (m *Machine) setupPSP() {
	base := uint32(m.pspSeg) << 4
	m.Mem[base+0], m.Mem[base+1] = 0xCD, 0x20 // INT 20h
	m.w16(base+0x02, m.memTop)                // segment of first byte past allocation
	m.w16(base+0x2C, m.envSeg)                // environment segment
	// A bare command tail (length 0, then CR).
	m.Mem[base+0x80] = 0
	m.Mem[base+0x81] = 0x0D
}

// setupEnv writes a minimal DOS environment block: a couple of variables, the
// double-zero terminator, a string count, then the program's full path (argv[0],
// which the MS-C overlay manager reopens to page overlays — resolveFile maps its
// basename back to the real executable, so the fabricated drive/dir is harmless).
func (m *Machine) setupEnv(exePath string) {
	base := uint32(m.envSeg) << 4
	var b []byte
	b = append(b, []byte("PATH=C:\\\x00")...)
	b = append(b, []byte("COMSPEC=C:\\COMMAND.COM\x00")...)
	b = append(b, 0x00)       // end of variables
	b = append(b, 0x01, 0x00) // one string follows
	b = append(b, []byte("C:\\"+strings.ToUpper(filepath.Base(exePath))+"\x00")...)
	copy(m.Mem[base:], b)
}

// resolveFile maps a DOS path (as passed to INT 21h/3Dh) to a host path under
// the game directory, matching components case-insensitively. The full relative
// path is tried first; failing that, the basename is looked up directly in the
// game directory (DOS argv[0] carries a fabricated drive/dir — e.g. the overlay
// manager reopens "C:\UW\UW.EXE", which is really just game/UW.EXE).
func (m *Machine) resolveFile(dosPath string) (string, bool) {
	if host, ok := m.walkPath(dosPath); ok {
		return host, true
	}
	norm := strings.ReplaceAll(dosPath, "\\", "/")
	base := norm[strings.LastIndex(norm, "/")+1:]
	if base != "" && base != dosPath {
		if host, ok := m.walkPath(base); ok {
			return host, true
		}
	}
	// Files the game itself created live in the scratch dir (temp/save files).
	if sp := m.scratchPath(dosPath); sp != "" {
		if _, err := os.Stat(sp); err == nil {
			return sp, true
		}
	}
	return "", false
}

// scratchPath maps a DOS path to its location in the writable scratch overlay,
// preserving the directory structure (so SAVE0\FILE lands under scratch/SAVE0).
func (m *Machine) scratchPath(dosPath string) string {
	if m.scratchDir == "" {
		return ""
	}
	p := strings.ReplaceAll(dosPath, "\\", "/")
	p = strings.TrimPrefix(p, "./")
	if len(p) >= 2 && p[1] == ':' { // drop a drive letter
		p = p[2:]
	}
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return ""
	}
	return filepath.Join(m.scratchDir, strings.ToUpper(p))
}

// walkPath resolves a DOS path case-insensitively, trying the writable scratch
// overlay first (files/dirs the game created, e.g. SAVE0) and then the read-only
// game directory.
func (m *Machine) walkPath(dosPath string) (string, bool) {
	if m.scratchDir != "" {
		if host, ok := walkRoot(m.scratchDir, dosPath); ok {
			return host, true
		}
	}
	return walkRoot(m.gameDir, dosPath)
}

// walkRoot resolves a DOS path against one host root, one case-insensitive
// component at a time.
func walkRoot(root, dosPath string) (string, bool) {
	p := strings.ReplaceAll(dosPath, "\\", "/")
	p = strings.TrimPrefix(p, "./")
	if len(p) >= 2 && p[1] == ':' { // drop a leading drive letter
		p = p[2:]
	}
	p = strings.TrimPrefix(p, "/")
	cur := root
	for _, comp := range strings.Split(p, "/") {
		if comp == "" {
			continue
		}
		entries, err := os.ReadDir(cur)
		if err != nil {
			return "", false
		}
		match := ""
		for _, e := range entries {
			if strings.EqualFold(e.Name(), comp) {
				match = e.Name()
				break
			}
		}
		if match == "" {
			return "", false
		}
		cur = filepath.Join(cur, match)
	}
	return cur, true
}

// SeedDir creates an empty working directory (given as a DOS-style relative
// path) in the writable scratch overlay. Games that expect a save/work folder to
// exist on first run use this — e.g. Ultima Underworld aborts if SAVE0 is
// missing but must handle it empty, which is the fresh-install state; the game
// then populates it itself. Call it after LoadEXE, before running.
func (m *Machine) SeedDir(rel string) {
	if sp := m.scratchPath(rel); sp != "" {
		os.MkdirAll(sp, 0o755)
	}
}

// allocFH returns the lowest free DOS file handle, exactly as real DOS does
// (0-4 are the standard handles). This matters: the Microsoft C runtime indexes
// its internal per-file flag table by DOS handle number and sizes it for DOS's
// small per-process handle table, so handles must stay small and be reused —
// a monotonically increasing counter eventually indexes past that table and
// fread fails without ever issuing a DOS read.
func (m *Machine) allocFH() uint16 {
	for h := uint16(5); ; h++ {
		if _, inUse := m.files[h]; !inUse {
			return h
		}
	}
}

// logf records an event.
func (m *Machine) logf(format string, args ...interface{}) {
	m.Log = append(m.Log, fmt.Sprintf(format, args...))
}

// asciiz reads a NUL-terminated string from seg:off.
func (m *Machine) asciiz(seg, off uint16) string {
	var sb strings.Builder
	a := lin(seg, off)
	for i := 0; i < 128; i++ {
		c := m.Mem[(a+uint32(i))&0xFFFFF]
		if c == 0 {
			break
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

// IntSummary returns the INT 21h function histogram, most-used first.
func (m *Machine) IntSummary() []string {
	type kv struct {
		ah byte
		n  int
	}
	var xs []kv
	for ah, n := range m.IntCounts {
		xs = append(xs, kv{ah, n})
	}
	sort.Slice(xs, func(i, j int) bool { return xs[i].n > xs[j].n })
	var out []string
	for _, x := range xs {
		out = append(out, fmt.Sprintf("AH=%02X x%d", x.ah, x.n))
	}
	return out
}
