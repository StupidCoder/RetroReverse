// go32 protected-mode machine. Where dos.Machine runs a 16-bit real-mode MZ
// program (seg<<4 addressing, INT 21h services), PM runs a DJGPP "go32" image:
// the i386 COFF payload placed flat in 32-bit protected mode, with the go32
// runtime's DPMI host (INT 31h) and the handful of DOS/BIOS INTs it reaches
// directly serviced in Go. This is the Phase-A bring-up host for the mode-aware
// x86 core — the protected-mode + x87 engine the original-Xbox oracle needs is
// proven here first, on a DOS-extender game (Quake shareware) that needs no NV2A
// and no Xbox kernel.
//
// Memory model. A go32 image sees a *flat* 32-bit address space, but not an
// identity-mapped one: under real CWSDPMI the program's segments (CS/DS/ES/SS)
// carry a non-zero linear base `__djgpp_base_address` (B), so a program virtual
// address v is backed at linear B+v, while physical/conventional memory — the BIOS
// data area, the DOS transfer buffer, DPMI conventional blocks, and the VGA
// framebuffer at 0xA0000 — lives at its *true* low linear address, below B. The
// program reaches that low memory two ways: through a separate base-0 selector
// (`_dos_ds`, the info block's selector_for_linear_memory) for absolute linear
// access, or as a near pointer via `__djgpp_conventional_base` = -B, which the
// C library adds to a physical address so that (phys + (-B)) through DS base B
// folds back to linear phys.
//
// We mirror that layout exactly. The COFF sections load at B + their virtual
// address; the segment bases are B; DPMI 0006h (Get Segment Base) hands the crt0
// B, from which nearptr derives conventional_base = -B. Loading the program high,
// with conventional memory and video low and distinct, is what keeps the mode-13h
// framebuffer at 0xA0000 from aliasing the program's own .bss (which, at base 0,
// spanned 0xA0000). There is still no paging and no real->PM mode switch — the
// go32 stub already switched before the COFF entry — only a non-zero segment base,
// which the Xbox phase will also use once its kernel sets up PM.
package dos

import (
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/x86"
)

// go32 memory layout. Two regions of the one backing slice:
//
//	LOW — physical/conventional memory, at true linear addresses [0, go32ImgBase):
//	  [0,        0x00500)  BIOS data area (the 0040:006C timer tick lives here)
//	  [convBase, convTop)  conventional DOS memory: the transfer buffer __tb, then
//	                       DPMI 0100h AllocDOSMem blocks. Its real-mode segment is
//	                       (base>>4), so seg<<4 == base; the game reaches it via the
//	                       base-0 _dos_ds selector or a conventional_base near pointer.
//	  [0xA0000,  0xC0000)  the VGA framebuffer window (mode-13h back-buffer target)
//
//	HIGH — the program's virtual-address space, backed at [go32ImgBase, MemSize),
//	i.e. program VA v is at linear go32ImgBase+v (segment bases = go32ImgBase):
//	  [B+0,       B+imgEnd)     the COFF image: .text, .data, zeroed .bss
//	  [B+heapBase,B+stackFloor) the extended-memory heap arena (DPMI 0501h AllocMem)
//	  [B+stackFloor,B+stackTop) the initial protected-mode stack (ESP at its top,
//	                            grows DOWN); the crt0 soon re-points ESP at a stack
//	                            it sbrk's just above the image (minstack bytes tall)
//	  [MemSize-go32InfoBytes, MemSize) the go32 info block, reserved at the very top
//
// The VA-space quantities (imgEnd, heapBase, stackFloor, stackTop) are computed as
// offsets and read exactly as they did in the old base-0 model — only the backing
// is shifted up by go32ImgBase. A tall (minstack) runtime stack keeps deep frames
// (Quake's COM_LoadPackFile reads the pak directory into a 128 KiB stack buffer)
// from descending into the image; __tb now lives in the LOW region, unreachable
// from the program stack no matter how deep the game recurses.
const (
	go32ImgBase    = 0x100000  // linear base of the program's VA space (__djgpp_base_address); conventional memory + video live below it
	go32VASize     = 64 << 20  // size of the program's virtual-address space (image + heap + stack + info)
	go32MemSize    = go32ImgBase + go32VASize // total backing: low physical region + high VA space
	go32StackBytes = 8 << 20   // 8 MiB initial PM stack, high in the VA space
	go32PageSize   = 0x1000
	go32InfoBytes  = 0x4000    // reserved region at the top of memory for the go32 info block + transfer buffer
	go32XferBytes  = 0x2000    // size of the DOS transfer buffer the info block advertises
	go32InfoSel    = 0x0040    // fabricated selector whose base is the info block (FS at entry)
	go32DosDS      = 0x0018    // base-0 selector the info block advertises for absolute linear/conventional access (_dos_ds)
	go32MinStack   = 8 << 20   // stubinfo minstack: the crt0 sbrk's this many bytes for the runtime stack
	go32ConvBase   = 0x10000   // start of the low conventional arena (below the 0xA0000 video window)

	// go32BaseAddrVA is the virtual address of the DJGPP runtime's
	// __djgpp_base_address global in the staged Quake image — the linear base of the
	// program's data segment. DJGPP derives every conventional/physical *near
	// pointer* from it: the C library sets __djgpp_conventional_base = -__djgpp_base_
	// address, and a physical address `phys` becomes the near pointer phys +
	// conventional_base = phys - base, which through the DS base (= base) folds back
	// to linear phys. So the mode-13h framebuffer (physical 0xA0000) only reaches the
	// low video window — instead of aliasing the program's own .bss — when this global
	// holds the true base go32ImgBase.
	//
	// On real CWSDPMI the crt0 learns the base as a side effect of its real-mode
	// memory-grow trampoline; we deliberately bypass that path (info block +0x1C is
	// set to the whole heap so growth never switches to 16-bit mode), so the crt0
	// leaves the global zero and every near pointer would collapse onto the program.
	// The host therefore maintains the invariant itself (see enforceBaseAddress). The
	// address was recovered from the image by disassembly — it is the operand of the
	// `NEG`/store that computes __djgpp_conventional_base — not from any external
	// source; a stripped go32 COFF carries no symbol table to look it up in.
	go32BaseAddrVA = 0x6D6CC
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

	// scripted keyboard injection (see go32_input.go)
	pmVectors  [256]pmVector // PM interrupt handlers recorded from DPMI 0205h
	defIntVec  pmVector      // synthesized default PM interrupt handler (IN 60h/EOI/IRETD)
	keyEvents []injEvent    // remaining scripted input events
	keyWait   int           // injection periods to pause before the next event
	keyRetry  bool          // a key is pending re-delivery until interrupts open
	injTick   int           // retired-instruction counter toward go32InjectPeriod
	keyHits   int           // delivered keystrokes (for capped logging)
	kbdData   byte          // 8042 output buffer (byte the game reads from 0x60)
	kbdFull   bool          // 8042 output-buffer-full status (bit0 of 0x64)

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
		sels:       map[uint16]uint32{0x08: go32ImgBase, 0x10: go32ImgBase, go32DosDS: 0}, // program code/data at base B; _dos_ds at base 0 (info-block sel added below)
		nextSel:    0x0100,                              // first fabricated selector (LDT-ish)
		virtIF:     true,                                // interrupts virtually enabled at entry
	}

	// Place each section at go32ImgBase + its virtual address; .bss is left zeroed
	// (make() did that), matching what the crt0's own REP STOSD would produce.
	// imgEnd stays a *virtual* address (the largest section end), read the same way
	// the base-0 model read it — only the backing is shifted up by go32ImgBase.
	var imgEnd uint32
	for _, s := range coff.Sections {
		end := s.VAddr + s.Size
		if end > imgEnd {
			imgEnd = end
		}
		if s.IsBSS() || s.Data == nil {
			continue
		}
		if int(go32ImgBase)+int(s.VAddr)+len(s.Data) > len(p.Mem) {
			return nil, fmt.Errorf("go32: section %q at VA %#x..%#x exceeds backing memory",
				s.Name, s.VAddr, s.VAddr+s.Size)
		}
		copy(p.Mem[go32ImgBase+s.VAddr:], s.Data)
	}

	// Conventional-memory arena: the low region below the 0xA0000 video window. The
	// transfer buffer __tb sits at its base (segment base>>4, so seg<<4 recovers it),
	// and DPMI 0100h blocks follow. This is genuine physical memory — the program
	// reaches it through the base-0 _dos_ds selector or a conventional_base near
	// pointer, both of which fold to these true low linear addresses.
	p.convBase = go32ConvBase
	p.convTop = 0xA0000
	xferLinear := p.convBase
	p.convNext = p.convBase + go32XferBytes // 0100h blocks start past the transfer buffer

	// The heap and stack are virtual addresses in the program's VA space; the info
	// block is reserved at the very top of the (linear) backing, its own FS-based
	// selector, so it can sit at a fixed high linear address independent of the base.
	p.infoBase = uint32(len(p.Mem)) - go32InfoBytes
	stackTopVA := uint32(go32VASize) - go32InfoBytes // VA whose backing is just below the info block
	p.stackFloor = stackTopVA - go32StackBytes
	p.heapBase = align(maxu32(imgEnd, 0x100000), go32PageSize)
	p.heapNext = p.heapBase
	p.setupInfoBlock(xferLinear)
	p.setupDefaultIntVec()

	c := x86.NewCPU(p)
	c.Mode = x86.ModeProt
	// The program's code/data selectors carry linear base go32ImgBase (the real
	// CWSDPMI __djgpp_base_address), so a virtual address v is fetched/loaded at
	// linear go32ImgBase+v. We set the conventional 08h code / 10h data pair a go32
	// program expects; their bases come from the sels map via SegResolve.
	c.Seg[x86.CS] = 0x08
	c.Seg[x86.DS] = 0x10
	c.Seg[x86.ES] = 0x10
	c.Seg[x86.SS] = 0x10
	c.Seg[x86.GS] = 0x10
	for i := range c.SegBase {
		c.SegBase[i] = go32ImgBase
	}
	// FS points at the go32 info block, the way the go32-v2 stub leaves it: the crt0
	// reads the transfer-buffer size and the memory top through FS. Its base is the
	// info block's fixed high linear address, independent of go32ImgBase.
	c.Seg[x86.FS] = go32InfoSel
	c.SegBase[x86.FS] = p.infoBase
	c.IP = coff.Entry
	c.Regs[x86.SP] = stackTopVA
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
	p.Write(b+0x1A, go32DosDS)   // selector_for_linear_memory: base-0 _dos_ds, distinct from the program's DS (base go32ImgBase)
	p.w32(b+0x1C, p.stackFloor)  // memory top the sbrk grows toward (the heap/stack boundary)
	p.w16(b+0x20, go32XferBytes) // stubinfo.minkeep — transfer-buffer size (__tb_size)
	p.w16(b+0x24, uint16(xferLinear>>4)) // stubinfo.ds_segment — __tb = seg<<4
}

// setupDefaultIntVec plants a minimal default protected-mode hardware-interrupt
// handler in guest memory and points defIntVec at it. On a real DPMI host, before
// a program hooks a hardware IRQ, its PM vector already holds the host's default
// reflector; DPMI 0204h (Get PM Interrupt Vector) returns that so the program can
// chain to it. go32 games do exactly this — Quake's Ctrl-C keyboard filter saves
// the previous INT 9 vector and JMPFs to it for every key it does not consume — so
// a null default sends the chain to linear 0 and crashes. This stub stands in for
// the reflector: read the keyboard data port (so the 8042 is acknowledged), send
// EOI to both PICs, and IRETD. It discards the scancode (nothing here reflects to a
// BIOS buffer the game reads), but it makes the chain safe; the game's own IRQ
// handler, installed later, is what actually delivers keys.
func (p *PM) setupDefaultIntVec() {
	// Placed in the reserved info-block region (past the ~0x30-byte stubinfo), which
	// nothing else uses. The bytes go at their true high linear address; CS 0x08 has
	// base go32ImgBase, so the vector's offset is that linear address minus the base,
	// making CS_base+offset resolve back to where the bytes actually are.
	stubLinear := p.infoBase + 0x200
	stub := []byte{
		0xE4, 0x60, // IN AL, 0x60      — acknowledge the keyboard controller
		0xB0, 0x20, // MOV AL, 0x20     — non-specific EOI
		0xE6, 0xA0, // OUT 0xA0, AL     — EOI to the slave PIC
		0xE6, 0x20, // OUT 0x20, AL     — EOI to the master PIC
		0xCF, //       IRETD            — pop the 32-bit interrupt frame
	}
	for i, b := range stub {
		p.Write(stubLinear+uint32(i), b)
	}
	p.defIntVec = pmVector{sel: 0x08, off: stubLinear - go32ImgBase, set: true}
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

// enforceBaseAddress keeps the DJGPP __djgpp_base_address global equal to the true
// linear base go32ImgBase. The crt0 would set it via the real-mode memory-grow
// trampoline we bypass, so it stays zero on its own — which would make every
// conventional/physical near pointer (the mode-13h framebuffer above all) collapse
// onto the program's own address space. Writing go32ImgBase each step is cheap and
// idempotent: it is the correct permanent value (the program's data segment really
// is based there), the game never writes the global itself, and it is read only far
// later (the __djgpp_conventional_base computation and the linear-address helpers),
// so the invariant simply needs to hold by the time anything reads it. Written high
// (at go32ImgBase + the global's virtual address) since the global lives in .bss.
func (p *PM) enforceBaseAddress() {
	p.w32(go32ImgBase+go32BaseAddrVA, go32ImgBase)
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
