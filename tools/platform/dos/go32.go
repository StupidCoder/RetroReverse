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
//	[0,        imgEnd)    the COFF image: .text, .data, zeroed .bss
//	[imgEnd,   0x100000)  conventional DOS memory arena (DPMI 0100h AllocDOSMem),
//	                      if the image ends below 1 MiB — its returned real-mode
//	                      segment is (base>>4), so seg<<4 == base and go32's near-
//	                      pointer transfer-buffer access lands on the same bytes.
//	[0x100000, stackFloor) the extended-memory heap arena (DPMI 0501h AllocMem)
//	[stackFloor, stackTop) the initial protected-mode stack (ESP starts at its top,
//	                      grows DOWN through this region)
//	[stackTop, MemSize)   the go32 info block, reserved at the very top
//
// The stack lives HIGH in extended memory, above the heap — as it does under real
// CWSDPMI, where the DPMI client's stack is far above the conventional-memory
// transfer buffer. An earlier layout put a small stack just above 1 MiB and let it
// grow DOWN into conventional memory; a deep frame (Quake's COM_LoadPackFile reads
// the pak directory into a 128 KiB stack buffer) then descended straight through
// __tb at ~0xE4000, and DOS file I/O reading into __tb clobbered the frame's locals.
// Placing the stack high, with megabytes of headroom below the info block, keeps it
// clear of __tb and the image no matter how deep the game recurses.
const (
	go32MemSize    = 64 << 20  // flat linear space (image + conv + heap + stack)
	go32StackBytes = 8 << 20   // 8 MiB initial PM stack, high in extended memory
	go32PageSize   = 0x1000
	go32InfoBytes  = 0x4000    // reserved region at the top of memory for the go32 info block + transfer buffer
	go32XferBytes  = 0x2000    // size of the DOS transfer buffer the info block advertises
	go32InfoSel    = 0x0040    // fabricated selector whose base is the info block (FS at entry)
	go32MinStack   = 8 << 20   // stubinfo minstack: the crt0 sbrk's this many bytes for the runtime stack
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
	stackFloor                  uint32 // bottom of the PM stack region; the heap ceiling
	infoBase                    uint32 // linear base of the go32 info block (top of memory)

	nextSel      uint16            // next fabricated LDT selector to hand out
	nextCallback uint16            // next real-mode callback offset (DPMI 0303h)
	sels         map[uint16]uint32 // selector -> descriptor linear base (the modelled LDT)
	virtIF  bool              // DPMI virtual interrupt flag (go32 brackets critical sections with it)

	lolSeg, lolOff uint16 // real-mode far pointer to the fabricated DOS List of Lists (INT 21h AH=52h)
	dtaSeg, dtaOff uint16 // real-mode far pointer to the Disk Transfer Address (INT 21h AH=1Ah)

	pit      pitState  // 8254 timer counter 0, the game's clock (see go32_ports.go)
	retrace  bool      // VGA 0x3DA vertical-retrace toggle
	dacIndex int       // VGA DAC write cursor (register×3 + component), ports 0x3C8/0x3C9
	Pal      [768]byte // VGA DAC palette the game programs via 0x3C8/0x3C9 (for framebuffer export)

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

	// The heap grows UP from just above 1 MiB (or the image, if it overran 1 MiB); the
	// stack lives HIGH, growing DOWN from just below the info block, with the whole gap
	// between them for headroom. The info block is reserved at the very top of memory.
	p.infoBase = uint32(len(p.Mem)) - go32InfoBytes
	stackTop := p.infoBase
	p.stackFloor = stackTop - go32StackBytes
	p.heapBase = align(maxu32(imgEnd, 0x100000), go32PageSize)
	p.heapNext = p.heapBase
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
	c.PortIn = p.portIn
	c.PortOut = p.portOut
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
//	+0x14 minstack   — minimum runtime stack. The crt0 sbrk's max(its built-in
//	                  default, this) bytes for the stack and points ESP at the top,
//	                  so the stack lives at the BOTTOM of the sbrk arena, just above
//	                  the image and the conventional-memory transfer buffer __tb. A
//	                  small stack there overflows down through __tb (Quake's
//	                  COM_LoadPackFile alone reads the pak directory into a 128 KiB
//	                  stack buffer); a DOS read into __tb then clobbers the frame's
//	                  locals. Making minstack multi-megabyte pushes ESP megabytes
//	                  above __tb, so __tb sits in dead space below the active stack.
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
	p.w32(b+0x14, go32MinStack)  // minstack: the crt0 sbrk's this for the runtime stack (see below)
	p.Write(b+0x18, 0x08)        // master_interrupt_controller_base
	p.Write(b+0x19, 0x70)        // slave_interrupt_controller_base
	p.Write(b+0x1A, 0x10)        // selector_for_linear_memory (flat data)
	p.w32(b+0x1C, p.stackFloor)  // memory top the sbrk grows toward (the heap/stack boundary)
	p.w16(b+0x20, go32XferBytes) // stubinfo.minkeep — transfer-buffer size (__tb_size)
	p.w16(b+0x24, uint16(xferLinear>>4)) // stubinfo.ds_segment — __tb = seg<<4
}

const (
	dosJFTSize      = 64   // handles the fabricated Job File Table / SFT can size
	dosSFTEntrySize = 0x3B // System File Table entry stride DJGPP assumes for DOS 4+
)

// writeDOSStructures fabricates, on demand, the DOS kernel structures DJGPP's
// fstat walks to validate an open handle: the Program Segment Prefix's Job File
// Table (JFT), the DOS List of Lists (LoL), and a System File Table (SFT). On real
// DOS these are genuine kernel data; a go32 program reaches them because the
// transfer-buffer paragraph sits just above the PSP, so DJGPP derives the PSP as
// (stubinfo.ds_segment<<4)-0x100 == convBase-0x100 (see setupInfoBlock's +0x24).
//
// These are written *transiently*, from the INT 21h AH=52h handler, rather than
// once at load: in this flat model the DJGPP malloc heap begins right above the
// image and grows to tens of megabytes, so it owns every conventional address a
// real-mode far pointer can reach — nothing persistent survives there. But fstat's
// SFT walk reads the whole chain synchronously right after its AH=52h call, with no
// allocation in between, so writing the structures into the transfer-buffer scratch
// region at AH=52h time makes them valid exactly when they are read. The scratch is
// reused by the next file read, which is fine — the structures are needed only for
// the duration of one walk.
//
// The walk only proves the handle is open and yields its SFT index (st_ino); fstat
// still reads the file's *size* through lseek (AH=42h) and its *date* through AH=57h
// against the real host file, so these entries carry no real metadata.
//
// Returns false if there is no conventional transfer buffer to build them in, in
// which case AH=52h hands back a null pointer and fstat's SFT path stays disabled.
func (p *PM) writeDOSStructures() bool {
	if p.convTop == 0 {
		return false
	}
	psp := p.convBase - 0x100 // DJGPP's derived PSP, just below the transfer buffer
	// JFT, LoL and SFT live in the transfer buffer — scratch that nothing touches
	// between this AH=52h call and the walk's reads. All are 16-aligned (convBase is
	// page-aligned), so each far pointer's offset is zero and its segment is base>>4.
	jft := p.convBase
	lol := p.convBase + 0x40
	sft := p.convBase + 0x80
	sftCount := uint32(dosJFTSize)
	if sft+6+sftCount*dosSFTEntrySize > p.convBase+go32XferBytes {
		return false // transfer buffer too small (never true at current sizes)
	}

	// PSP: only the JFT size (+0x32) and the JFT far pointer (+0x34) are read.
	p.w16(psp+0x32, dosJFTSize)      // number of handles this process may hold
	p.w16(psp+0x34, uint16(jft&0xF)) // JFT far pointer: offset
	p.w16(psp+0x36, uint16(jft>>4))  // JFT far pointer: segment
	// JFT: map each handle to its own SFT slot (identity), so any handle below the
	// table size resolves to a valid, in-range SFT entry.
	for i := uint32(0); i < dosJFTSize; i++ {
		p.Write(jft+i, byte(i))
	}
	// List of Lists: fstat only follows the first-SFT far pointer at +0x04.
	p.w16(lol+0x04, uint16(sft&0xF)) // first SFT block far pointer: offset
	p.w16(lol+0x06, uint16(sft>>4))  // first SFT block far pointer: segment
	// SFT: a single block of sftCount entries; DJGPP indexes it by the JFT value.
	p.w16(sft+0x00, 0xFFFF)           // next-block offset = end of chain
	p.w16(sft+0x02, 0xFFFF)           // next-block segment
	p.w16(sft+0x04, uint16(sftCount)) // entries held in this block
	for i := uint32(0); i < sftCount; i++ {
		p.w16(sft+6+i*dosSFTEntrySize, 1) // reference count = 1 (entry in use)
	}
	p.lolSeg, p.lolOff = uint16(lol>>4), uint16(lol&0xF)
	return true
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
