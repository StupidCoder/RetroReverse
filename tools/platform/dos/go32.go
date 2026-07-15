// go32 protected-mode machine. Where dos.Machine runs a 16-bit real-mode MZ
// program (seg<<4 addressing, INT 21h services), PM runs a DJGPP "go32" image:
// the i386 COFF payload placed flat in 32-bit protected mode, with the go32
// runtime's DPMI host (INT 31h) and the handful of DOS/BIOS INTs it reaches
// directly serviced in Go. This is the Phase-A bring-up host for the mode-aware
// x86 core — the protected-mode + x87 engine the original-Xbox oracle needs is
// proven here first, on a DOS-extender game (Quake shareware) that needs no NV2A
// and no Xbox kernel.
//
// Memory model. A go32 image sees a *flat* 32-bit address space: every segment
// selector has linear base 0 and a 4 GiB limit, so a linear address is just an
// index into one backing slice (identity mapping). The COFF sections load at
// their own virtual addresses; the DPMI memory services hand out further linear
// ranges from bump arenas above the image. There is no paging and no real->PM
// mode switch here — the go32 stub already made the switch before the COFF entry,
// and the HLE'd DPMI host presents identity-mapped memory (paging and the mode
// switch move to the Xbox phase, where the kernel sets up PM regardless).
package dos

import (
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/x86"
)

// Flat go32 memory layout (all linear addresses into PM.Mem):
//
//	[0,       imgEnd)   the COFF image: .text, .data, zeroed .bss
//	[imgEnd,  0x100000) conventional DOS memory arena (DPMI 0100h AllocDOSMem),
//	                    if the image ends below 1 MiB — its returned real-mode
//	                    segment is (base>>4), so seg<<4 == base and go32's near-
//	                    pointer transfer-buffer access lands on the same bytes.
//	[base1M,  +stack)   the initial protected-mode stack (ESP starts at its top)
//	[stackTop, MemSize) the extended-memory heap arena (DPMI 0501h AllocMem)
const (
	go32MemSize    = 64 << 20 // flat linear space (image + conv + stack + heap)
	go32StackBytes = 0x80000  // 512 KiB initial PM stack
	go32PageSize   = 0x1000
	go32InfoBytes  = 0x4000   // reserved region at the top of memory for the go32 info block + transfer buffer
	go32XferBytes  = 0x2000   // size of the DOS transfer buffer the info block advertises
	go32InfoSel    = 0x0040   // fabricated selector whose base is the info block (FS at entry)
)

// PM is a loaded go32/COFF program ready to run in flat 32-bit protected mode.
type PM struct {
	Mem []byte // flat linear address space, identity-mapped
	CPU *x86.CPU

	gameDir string
	files   map[uint16]*os.File // reserved for later file I/O (DPMI 0300h bridge)

	// bump arenas (all linear addresses into Mem)
	convBase, convNext, convTop uint32 // conventional DOS memory (DPMI 0100h)
	heapBase, heapNext          uint32 // extended-memory heap (DPMI 0501h)
	infoBase                    uint32 // linear base of the go32 info block (also the heap ceiling)

	nextSel      uint16            // next fabricated LDT selector to hand out
	nextCallback uint16            // next real-mode callback offset (DPMI 0303h)
	sels         map[uint16]uint32 // selector -> descriptor linear base (the modelled LDT)
	virtIF  bool              // DPMI virtual interrupt flag (go32 brackets critical sections with it)

	// instrumentation
	Log        []string
	DPMICounts map[uint16]int // INT 31h function (AX) call histogram
	IntCounts  map[byte]int   // direct software-INT histogram (non-DPMI)
	DOSCounts  map[byte]int   // real-mode INT 21h AH histogram (via DPMI 0300h)
	Terminated bool
	ExitCode   byte
	Console    []byte // bytes the program wrote to stdout/stderr (DOS handles 1/2)
}

// LoadGo32 reads the go32 executable at exePath, places its COFF image flat in a
// fresh protected-mode machine, and seeds the entry registers. gameDir is where
// the program's data files will resolve (used once file I/O is wired).
func LoadGo32(exePath, gameDir string) (*PM, error) {
	data, err := os.ReadFile(exePath)
	if err != nil {
		return nil, err
	}
	return LoadGo32Bytes(data, gameDir)
}

// LoadGo32Bytes builds a protected-mode machine from an in-memory go32 image —
// the core of LoadGo32, split out so tests can supply a synthetic image.
func LoadGo32Bytes(data []byte, gameDir string) (*PM, error) {
	coff, err := ParseGo32COFF(data)
	if err != nil {
		return nil, err
	}

	p := &PM{
		Mem:        make([]byte, go32MemSize),
		gameDir:    gameDir,
		files:      map[uint16]*os.File{},
		DPMICounts: map[uint16]int{},
		IntCounts:  map[byte]int{},
		DOSCounts:  map[byte]int{},
		sels:       map[uint16]uint32{0x08: 0, 0x10: 0}, // flat code/data selectors, base 0 (info-block sel added below)
		nextSel:    0x0100,                              // first fabricated selector (LDT-ish)
		virtIF:     true,                                // interrupts virtually enabled at entry
	}

	// Place each section at its virtual address; .bss is left zeroed (make() did
	// that), matching what the crt0's own REP STOSD would produce.
	var imgEnd uint32
	for _, s := range coff.Sections {
		end := s.VAddr + s.Size
		if end > imgEnd {
			imgEnd = end
		}
		if s.IsBSS() || s.Data == nil {
			continue
		}
		if int(s.VAddr)+len(s.Data) > len(p.Mem) {
			return nil, fmt.Errorf("go32: section %q at %#x..%#x exceeds %d MiB flat memory",
				s.Name, s.VAddr, s.VAddr+s.Size, go32MemSize>>20)
		}
		copy(p.Mem[s.VAddr:], s.Data)
	}

	// Conventional-memory arena: the gap between the image end and 1 MiB, if any.
	// (Quake's image ends near 0xE3400, leaving ~112 KiB here.) The first slice of
	// it backs the DOS transfer buffer the info block advertises.
	p.convBase = align(imgEnd, go32PageSize)
	p.convTop = 0x100000
	xferLinear := uint32(0)
	if p.convBase < p.convTop {
		xferLinear = p.convBase
		p.convNext = p.convBase + go32XferBytes // 0100h blocks start past the transfer buffer
	} else {
		p.convBase, p.convTop = 0, 0 // image fills conventional memory; 0100h will fail
	}

	// Stack then heap live at/above 1 MiB (or above the image if it overran 1 MiB);
	// the go32 info block is a reserved region at the very top of memory.
	base1M := align(maxu32(imgEnd, 0x100000), go32PageSize)
	stackTop := base1M + go32StackBytes
	p.heapBase = stackTop
	p.heapNext = p.heapBase
	p.infoBase = uint32(len(p.Mem)) - go32InfoBytes
	p.setupInfoBlock(xferLinear)

	c := x86.NewCPU(p)
	c.Mode = x86.ModeProt
	// Flat selectors: all bases 0, limit 4 GiB. The selector *values* are cosmetic
	// (a flat model resolves every segment to base 0 via SegBase), but we set the
	// conventional 08h code / 10h data pair a go32 program expects.
	c.Seg[x86.CS] = 0x08
	c.Seg[x86.DS] = 0x10
	c.Seg[x86.ES] = 0x10
	c.Seg[x86.SS] = 0x10
	c.Seg[x86.GS] = 0x10
	// SegBase already zero from the zero value; make the flat base explicit.
	for i := range c.SegBase {
		c.SegBase[i] = 0
	}
	// FS points at the go32 info block, the way the go32-v2 stub leaves it: the
	// crt0 reads the transfer-buffer size and the memory top through FS.
	c.Seg[x86.FS] = go32InfoSel
	c.SegBase[x86.FS] = p.infoBase
	c.IP = coff.Entry
	c.Regs[x86.SP] = stackTop
	c.IF = true
	c.IntHook = p.handleInt
	c.SegResolve = p.resolveSel
	p.CPU = c
	return p, nil
}

// resolveSel returns the linear base of a protected-mode selector. Selectors the
// DPMI host handed out (DOS-memory blocks, descriptor aliases) carry the base
// they were created with; anything unknown is treated as flat (base 0), which
// covers the standard go32 code/data selectors.
func (p *PM) resolveSel(sel uint16) uint32 { return p.sels[sel] }

// mapSel records selector -> linear base in the modelled LDT.
func (p *PM) mapSel(sel uint16, base uint32) { p.sels[sel] = base }

// setupInfoBlock fills the structure the go32-v2 crt0 reads through FS. Despite
// its name, that structure is the go32-v2 STUBINFO — the block the real stub
// prepends and hands the program — not the runtime _go32_info_block (the crt0
// builds that one separately from DPMI queries). The crt0 copies [FS:0x10] bytes
// of it to a heap buffer and reads individual fields; the load-bearing ones for
// boot are:
//
//	+0x10 size       — copy length and the sbrk size the crt0 requests
//	+0x14 minstack   — minimum stack (compared, then reserved)
//	+0x18 memory_handle — DPMI handle of the DOS-memory block (opaque to us)
//	+0x1C            — the ceiling the sbrk compares the break against; set to the
//	                  top of our flat heap so growth always takes the fast path and
//	                  the real-mode-switch trampoline never runs (our space is
//	                  already backed)
//	+0x20 minkeep    — size of the DOS transfer buffer; the C library keeps it as
//	                  __tb_size and chunks every fread/fwrite to it
//	+0x24 ds_segment — real-mode segment of the transfer buffer; __tb = it<<4, the
//	                  conventional-memory address DOS file I/O reads and writes
//
// The +0x20/+0x24 pair is what makes DOS file I/O work: without it __tb and
// __tb_size are zero, and write() loops forever writing zero-length chunks.
func (p *PM) setupInfoBlock(xferLinear uint32) {
	p.mapSel(go32InfoSel, p.infoBase)
	b := p.infoBase
	p.w32(b+0x00, 0x30)          // size_of_this_structure_in_bytes (unused by crt0)
	p.w32(b+0x04, 0x000B8000)    // linear_address_of_primary_screen
	p.w32(b+0x08, 0x000B0000)    // linear_address_of_secondary_screen
	p.w32(b+0x0C, xferLinear)    // linear_address_of_transfer_buffer
	p.w32(b+0x10, go32XferBytes) // stubinfo.size — crt0 copy length / sbrk request
	p.w32(b+0x14, 1)             // minstack (kept minimal, as before)
	p.Write(b+0x18, 0x08)        // master_interrupt_controller_base
	p.Write(b+0x19, 0x70)        // slave_interrupt_controller_base
	p.Write(b+0x1A, 0x10)        // selector_for_linear_memory (flat data)
	p.w32(b+0x1C, p.infoBase)    // memory top the sbrk grows toward (never exceeded)
	p.w16(b+0x20, go32XferBytes) // stubinfo.minkeep — transfer-buffer size (__tb_size)
	p.w16(b+0x24, uint16(xferLinear>>4)) // stubinfo.ds_segment — __tb = seg<<4
}

// --- x86.Bus (flat identity mapping, bounds-checked) ---

func (p *PM) Read(a uint32) byte {
	if int(a) < len(p.Mem) {
		return p.Mem[a]
	}
	p.fault("read", a)
	return 0xFF
}

func (p *PM) Write(a uint32, v byte) {
	if int(a) < len(p.Mem) {
		p.Mem[a] = v
		return
	}
	p.fault("write", a)
}

// fault records (once) an access outside the backing store and halts the CPU, so
// a wild pointer surfaces as a concrete observed address rather than a panic.
func (p *PM) fault(kind string, a uint32) {
	if p.CPU != nil && !p.CPU.Halted {
		p.CPU.Halt("out-of-range %s at linear %08X (PC %s, %d MiB backing)",
			kind, a, pcHex(p.CPU), go32MemSize>>20)
	}
}

// --- small helpers ---

func align(v, a uint32) uint32 { return (v + a - 1) &^ (a - 1) }
func maxu32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
func pcHex(c *x86.CPU) string { return fmt.Sprintf("%08X", c.IP) }

// r32/w32 access a little-endian dword in the flat store (arena bookkeeping,
// DPMI result buffers).
func (p *PM) r32(a uint32) uint32 {
	return uint32(p.Read(a)) | uint32(p.Read(a+1))<<8 | uint32(p.Read(a+2))<<16 | uint32(p.Read(a+3))<<24
}
func (p *PM) w32(a, v uint32) {
	p.Write(a, byte(v))
	p.Write(a+1, byte(v>>8))
	p.Write(a+2, byte(v>>16))
	p.Write(a+3, byte(v>>24))
}
func (p *PM) w16(a uint32, v uint16) {
	p.Write(a, byte(v))
	p.Write(a+1, byte(v>>8))
}

// asciiz reads a NUL-terminated string at flat linear address a.
func (p *PM) asciiz(a uint32) string {
	var b []byte
	for i := uint32(0); i < 4096; i++ {
		ch := p.Read(a + i)
		if ch == 0 {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}

// logf records an event.
func (p *PM) logf(format string, args ...interface{}) {
	p.Log = append(p.Log, fmt.Sprintf(format, args...))
}
