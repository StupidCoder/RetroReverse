package dos

import (
	"encoding/binary"
	"testing"

	"retroreverse.com/tools/cpu/x86"
)

// buildGo32 assembles a minimal but well-formed go32 image: a 0x40-byte MZ stub
// followed by an i386 COFF with one .text section holding prog, entered in flat
// 32-bit protected mode at textVAddr. It mirrors what ParseGo32COFF expects so
// the loader can be exercised without the (gitignored) game image.
func buildGo32(prog []byte, textVAddr uint32) []byte {
	le := binary.LittleEndian
	const stub = 0x40
	buf := make([]byte, stub)
	buf[0], buf[1] = 'M', 'Z'
	le.PutUint16(buf[0x02:], 0x40) // e_cblp: bytes on last page
	le.PutUint16(buf[0x04:], 1)    // e_cp: one page -> LoadImageEnd = 0x40
	le.PutUint16(buf[0x08:], 4)    // e_cparhdr: header = 4 paras = 0x40 bytes

	// COFF file header (20) + optional header (28) + one section header (40).
	coff := make([]byte, 20+28+40)
	le.PutUint16(coff[0:], 0x014C) // f_magic i386
	le.PutUint16(coff[2:], 1)      // f_nscns
	le.PutUint16(coff[16:], 28)    // f_opthdr
	le.PutUint16(coff[18:], 0x010F)
	opt := coff[20:]
	le.PutUint16(opt[0:], 0x010B)            // ZMAGIC
	le.PutUint32(opt[4:], uint32(len(prog))) // tsize
	le.PutUint32(opt[12:], 0x1000)           // bsize (scratch bss)
	le.PutUint32(opt[16:], textVAddr)        // entry
	le.PutUint32(opt[20:], textVAddr)        // text_start
	sh := coff[20+28:]
	copy(sh[0:8], ".text")
	le.PutUint32(sh[8:], textVAddr)  // s_paddr
	le.PutUint32(sh[12:], textVAddr) // s_vaddr
	le.PutUint32(sh[16:], uint32(len(prog)))
	// s_scnptr is relative to the COFF start (StubEnd); text bytes follow the
	// section header, i.e. at COFF offset len(coff).
	le.PutUint32(sh[20:], uint32(len(coff)))
	le.PutUint32(sh[36:], 0x20) // STYP_TEXT

	out := append(buf, coff...)
	return append(out, prog...)
}

// TestGo32LoadAndDPMI boots a synthetic go32 image that issues a DPMI call and
// halts, verifying the whole loader path: COFF placement, flat-PM entry, and
// INT 31h routing through the DPMI host.
func TestGo32LoadAndDPMI(t *testing.T) {
	// MOV AX, 0x0400 ; INT 31h (Get DPMI Version) ; HLT
	prog := []byte{0x66, 0xB8, 0x00, 0x04, 0xCD, 0x31, 0xF4}
	m, err := LoadGo32Bytes(buildGo32(prog, 0x1000), "")
	if err != nil {
		t.Fatal(err)
	}
	if m.CPU.IP != 0x1000 {
		t.Fatalf("entry EIP = %08X, want 00001000", m.CPU.IP)
	}
	m.CPU.Run(100)
	if !m.CPU.Halted {
		t.Fatalf("did not halt (EIP=%08X)", m.CPU.IP)
	}
	if m.DPMICounts[0x0400] != 1 {
		t.Errorf("DPMI 0400h call count = %d, want 1", m.DPMICounts[0x0400])
	}
	// 0400h reports DPMI 1.0 in AX.
	if ax := m.CPU.Reg16(x86.AX); ax != 0x0100 {
		t.Errorf("AX = %04X after Get DPMI Version, want 0100", ax)
	}
}

// TestGo32DescriptorAlias exercises the exact mechanism that first broke Quake's
// boot: allocate a DOS-memory block (0100h), read its descriptor (000Bh), install
// that descriptor on a fresh selector (000Ch), load the alias, and store through
// it. The store must land at the block's linear base, proving descriptor bases
// round-trip through the get/set dance rather than collapsing to flat zero.
func TestGo32DescriptorAlias(t *testing.T) {
	// The descriptor scratch buffer is 0x00010000 (inside the image's .bss).
	prog := []byte{
		0x66, 0xBB, 0x01, 0x00, // MOV BX, 1            ; 1 paragraph
		0x66, 0xB8, 0x00, 0x01, // MOV AX, 0x0100       ; alloc DOS mem
		0xCD, 0x31, //             INT 31h              ; -> DX = based selector
		0x66, 0x89, 0xD3, //       MOV BX, DX
		0xBF, 0x00, 0x00, 0x01, 0x00, // MOV EDI, 0x00010000
		0x66, 0xB8, 0x0B, 0x00, // MOV AX, 0x000B       ; get descriptor -> [buf]
		0xCD, 0x31, //             INT 31h
		0x31, 0xC0, //             XOR EAX, EAX          ; AX=0 -> alloc LDT descriptor
		0x66, 0xB9, 0x01, 0x00, // MOV CX, 1
		0xCD, 0x31, //             INT 31h              ; -> AX = new selector
		0x66, 0x89, 0xC6, //       MOV SI, AX
		0x66, 0x89, 0xC3, //       MOV BX, AX
		0x66, 0xB8, 0x0C, 0x00, // MOV AX, 0x000C       ; set descriptor from [buf]
		0xBF, 0x00, 0x00, 0x01, 0x00, // MOV EDI, 0x00010000
		0xCD, 0x31, //             INT 31h
		0x8E, 0xC6, //             MOV ES, SI            ; ES = based alias
		0x26, 0xC7, 0x05, 0x00, 0x00, 0x00, 0x00, 0x01, 0xEF, 0xCD, 0xAB, // MOV DWORD [ES:0], 0xABCDEF01
		0xF4, // HLT
	}
	m, err := LoadGo32Bytes(buildGo32(prog, 0x1000), "")
	if err != nil {
		t.Fatal(err)
	}
	m.CPU.Run(1000)
	if !m.CPU.Halted {
		t.Fatalf("did not halt (EIP=%08X, reason=%q)", m.CPU.IP, m.CPU.HaltReason)
	}
	// The first 0100h block sits just past the transfer buffer in the conventional
	// arena; the store through its code-alias must land there.
	want := m.convBase + go32XferBytes
	got := binary.LittleEndian.Uint32(m.Mem[want:])
	if got != 0xABCDEF01 {
		t.Errorf("store through descriptor alias landed wrong: Mem[%08X] = %08X, want ABCDEF01", want, got)
	}
	if binary.LittleEndian.Uint32(m.Mem[0:]) == 0xABCDEF01 {
		t.Error("store landed at linear 0 — the descriptor base collapsed to flat zero (the original bug)")
	}
}

// TestGo32StateRoundTrip proves the PM savestate is a faithful, independent copy:
// run a synthetic image partway, snapshot it, run it to a different state, then
// restore — the machine must be bit-for-bit back where the snapshot was taken, and
// running on from there must reach the same place the uninterrupted run did.
func TestGo32StateRoundTrip(t *testing.T) {
	// A loop that writes a walking value into .bss and decrements a counter, so both
	// memory and registers evolve step by step and a mismatch is easy to see.
	//   MOV ECX, 8 ; MOV EBX, 0x2000
	// loop: MOV [EBX], CL ; INC EBX ; DEC ECX ; JNZ loop ; HLT
	prog := []byte{
		0xB9, 0x08, 0x00, 0x00, 0x00, // MOV ECX, 8
		0xBB, 0x00, 0x20, 0x00, 0x00, // MOV EBX, 0x2000
		0x88, 0x0B, // MOV [EBX], CL   (loop: at VA 0x100A)
		0x43,       // INC EBX
		0x49,       // DEC ECX
		0x75, 0xFA, // JNZ loop (-6, back to 0x100A)
		0xF4, // HLT
	}
	build := func() *PM {
		m, err := LoadGo32Bytes(buildGo32(prog, 0x1000), "")
		if err != nil {
			t.Fatal(err)
		}
		return m
	}

	// A reference run to the halt, for the register/memory fingerprint at the end.
	ref := build()
	ref.CPU.Run(10_000)
	if !ref.CPU.Halted {
		t.Fatal("reference run did not halt")
	}

	// Run partway, snapshot, then run on and diverge.
	m := build()
	m.CPU.Run(9) // mid-loop: some bytes written, ECX partly counted down
	st := m.SaveState()
	midIP, midECX := m.CPU.IP, m.CPU.Regs[x86.CX]
	midB := m.Mem[go32ImgBase+0x2000]

	m.CPU.Run(10_000) // run to the end, changing everything the snapshot captured
	if !m.CPU.Halted {
		t.Fatal("second run did not halt")
	}

	// Restore: the machine must be exactly back at the snapshot.
	if err := m.LoadState(st); err != nil {
		t.Fatal(err)
	}
	if m.CPU.IP != midIP || m.CPU.Regs[x86.CX] != midECX {
		t.Errorf("after restore IP=%08X ECX=%d, want %08X %d", m.CPU.IP, m.CPU.Regs[x86.CX], midIP, midECX)
	}
	if m.Mem[go32ImgBase+0x2000] != midB {
		t.Errorf("restored memory byte = %02X, want %02X", m.Mem[go32ImgBase+0x2000], midB)
	}
	if m.CPU.Halted {
		t.Error("restore left the CPU halted; it should be mid-run")
	}

	// Running on from the restored snapshot must reach the same end state as the
	// uninterrupted reference run — proving the CPU's hooks are still wired and the
	// state is genuinely resumable, not just readable.
	m.CPU.Run(10_000)
	if !m.CPU.Halted {
		t.Fatal("resumed run did not halt")
	}
	for i := 0; i < 8; i++ {
		if got, want := m.Mem[go32ImgBase+0x2000+uint32(i)], ref.Mem[go32ImgBase+0x2000+uint32(i)]; got != want {
			t.Errorf("resumed Mem[%d] = %02X, ref = %02X", i, got, want)
		}
	}
	if m.CPU.Regs[x86.CX] != ref.CPU.Regs[x86.CX] {
		t.Errorf("resumed ECX = %d, ref = %d", m.CPU.Regs[x86.CX], ref.CPU.Regs[x86.CX])
	}
}

// TestGo32InteractiveKey drives the interactive Keyer path: a tiny program installs a
// PM INT 9 handler on the heap that reads the scancode from port 0x60 into a fixed
// slot, then spins with interrupts enabled. A scancode queued with EnqueueScancode
// must be delivered (an IRQ1 through that handler) as the loop runs.
func TestGo32InteractiveKey(t *testing.T) {
	// The ISR (placed at a heap address so deliverScancode's "ISR installed" gate opens):
	//   IN AL, 0x60 ; MOV [0x3000], AL ; MOV AL, 0x20 ; OUT 0x20, AL ; IRETD
	isr := []byte{0xE4, 0x60, 0xA2, 0x00, 0x30, 0x00, 0x00, 0xB0, 0x20, 0xE6, 0x20, 0xCF}
	// The program: record the ISR's PM vector via DPMI 0205h (BL=9, CX=CS, EDX=heap
	// offset), STI, then spin. We hand-build it so the vector points into the heap.
	//   STI ; JMP $ (EB FE)
	prog := []byte{0xFB, 0xEB, 0xFE}
	m, err := LoadGo32Bytes(buildGo32(prog, 0x1000), "")
	if err != nil {
		t.Fatal(err)
	}
	// Plant the ISR in the heap and register it as the PM INT 9 handler, as a DPMI
	// 0205h call would. off must be >= heapBase for deliverScancode to accept it.
	isrOff := m.heapBase
	for i, b := range isr {
		m.Write(go32ImgBase+isrOff+uint32(i), b)
	}
	m.setPMVector(9, 0x08, isrOff)

	m.CPU.OnStep = func(c *x86.CPU) { m.PumpInput(c) }
	m.EnqueueScancode(0x48) // Up make code
	if !m.InteractiveKeysPending() {
		t.Fatal("scancode not queued")
	}
	m.CPU.Run(200_000)
	if m.InteractiveKeysPending() {
		t.Fatal("scancode was never delivered (still queued)")
	}
	if got := m.Mem[go32ImgBase+0x3000]; got != 0x48 {
		t.Errorf("ISR recorded scancode %02X, want 48 — the key did not reach the handler", got)
	}
}
