// Package x86 is a table/pattern-driven Intel x86 disassembler for 16-bit
// real-mode code — the counterpart to the mos6502/m68k/z80 packages for the
// repository's first MS-DOS title (Ultima Underworld, UW.EXE). It provides the
// same surface: a Decode function returning one fully-classified instruction
// (length, text, and a Flow category for recursive-descent tracing) and a
// Disassemble helper that renders a blob line by line.
//
// x86 instructions are variable length: an optional run of prefix bytes
// (segment overrides, the 0x66/0x67 operand/address-size toggles, LOCK, REP),
// a one- or two-byte (0x0F) opcode, an optional ModR/M byte that may pull in a
// SIB byte and a displacement, and an optional immediate. Decode reads exactly
// as many bytes as each form needs, so Inst.Len is always right and a linear or
// recursive walk stays aligned. The default operand and address size is 16-bit
// (real mode); a 0x66 prefix selects 32-bit operands and 0x67 32-bit addressing
// (the 386 extensions a late-real-mode game reaches for). Syntax is Intel order
// (destination first); mnemonics and registers are upper-case and immediates
// are shown as $-prefixed hex, matching the other disassemblers here.
//
// Anything not recognised decodes as a ".byte" pseudo-op with FlowStop so the
// caller can treat it as data; the opcode map is complete enough for the
// integer + x87 instruction set that this is rare on real code.
package x86

import "fmt"

// Flow classifies how an instruction affects control flow — the information a
// recursive-descent disassembler needs to follow every reachable path. It
// mirrors the mos6502/m68k/z80 Flow categories.
type Flow int

const (
	FlowSeq     Flow = iota // continues to the next instruction
	FlowBranch              // conditional branch (Jcc/LOOP/JCXZ): continues AND may go to Target
	FlowJump                // unconditional JMP to a known Target, no fall-through
	FlowCall                // CALL: goes to Target (when known), normally returns after it
	FlowReturn              // RET/RETF/IRET: path ends
	FlowIndJump             // JMP through a register/memory: target not statically known
	FlowStop                // undefined opcode, HLT/UD2, or truncated read: treat as data/stop
)

// Inst is one decoded instruction.
type Inst struct {
	Addr      uint32
	Len       int
	Mnem      string // the bare mnemonic ("MOV", "JMP", …)
	Text      string // formatted "MNEM operands"
	Flow      Flow
	Target    uint32 // branch/call/jump destination, when HasTarget
	HasTarget bool
}

// register-name tables, indexed by the 3-bit reg/rm fields.
var (
	reg8  = [8]string{"AL", "CL", "DL", "BL", "AH", "CH", "DH", "BH"}
	reg16 = [8]string{"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI"}
	reg32 = [8]string{"EAX", "ECX", "EDX", "EBX", "ESP", "EBP", "ESI", "EDI"}
	sreg  = [8]string{"ES", "CS", "SS", "DS", "FS", "GS", "?6", "?7"}
	// 16-bit addressing r/m base expressions (mod != 3), rm 0..7.
	rm16 = [8]string{"BX+SI", "BX+DI", "BP+SI", "BP+DI", "SI", "DI", "BP", "BX"}
	// condition mnemonic suffixes for Jcc/SETcc/CMOVcc, indexed by the low nibble.
	ccName = [16]string{"O", "NO", "B", "NB", "Z", "NZ", "BE", "A",
		"S", "NS", "P", "NP", "L", "GE", "LE", "G"}
	alu = [8]string{"ADD", "OR", "ADC", "SBB", "AND", "SUB", "XOR", "CMP"}
	shf = [8]string{"ROL", "ROR", "RCL", "RCR", "SHL", "SHR", "SAL", "SAR"}
)

// Decode decodes the instruction whose first byte is code[0] and whose absolute
// (linear) address is addr. Near branch/call targets are computed as
// addr+len+displacement in the flat address space of the supplied blob; a far
// pointer's Target is its seg:off folded to a linear byte address. A read past
// the end of code, or an unrecognised opcode, decodes as a one-byte ".byte"
// with FlowStop.
func Decode(code []byte, addr uint32) Inst {
	d := &dec{mem: code, addr: addr}
	in := d.run()
	return d.finish(code, addr, in)
}

// Decode32 is Decode with a 32-bit default operand/address size — for flat
// protected-mode / go32 code, where the CS descriptor's D bit is 1 and a 0x66/
// 0x67 prefix selects 16-bit rather than 32-bit. Real-mode callers use Decode.
func Decode32(code []byte, addr uint32) Inst {
	d := &dec{mem: code, addr: addr, defSize: 32}
	in := d.run()
	return d.finish(code, addr, in)
}

// finish applies the common post-decode fix-ups (bad-opcode fallback, length,
// prefix text) shared by Decode and Decode32.
func (d *dec) finish(code []byte, addr uint32, in Inst) Inst {
	if d.bad {
		b := byte(0)
		if len(code) > 0 {
			b = code[0]
		}
		return Inst{Addr: addr, Len: 1, Mnem: ".byte", Text: fmt.Sprintf(".byte $%02X", b), Flow: FlowStop}
	}
	in.Addr = addr
	in.Len = d.p
	// Fold accumulated prefixes (LOCK / REP) into the rendered text.
	if d.prefixText != "" {
		in.Text = d.prefixText + in.Text
	}
	return in
}

// dec carries the decode cursor and the prefix state for one instruction.
type dec struct {
	mem  []byte
	addr uint32 // linear address of mem[0]
	p    int    // read cursor, an offset into mem

	defSize  int    // default operand/address size: 0 or 16 => 16-bit, 32 => 32-bit (flat PM)
	opsize   int    // 16 or 32 (default from defSize; a 0x66 prefix toggles it)
	addrsize int    // 16 or 32 (default from defSize; a 0x67 prefix toggles it)
	seg      string // segment-override display prefix ("ES:"…) or ""
	lock     bool
	rep      string // "REP "/"REPNE "/"" (rendered ahead of string ops)

	prefixText string // LOCK/REP text to prepend after decode
	bad        bool
}

// next reads one byte and advances; a read past the end marks the instr bad.
func (d *dec) next() byte {
	if d.p >= len(d.mem) {
		d.bad = true
		return 0
	}
	b := d.mem[d.p]
	d.p++
	return b
}

func (d *dec) imm8() uint8 { return d.next() }
func (d *dec) imm16() uint16 {
	lo := uint16(d.next())
	hi := uint16(d.next())
	return lo | hi<<8
}
func (d *dec) imm32() uint32 {
	lo := uint32(d.imm16())
	hi := uint32(d.imm16())
	return lo | hi<<16
}

// immOsz reads an operand-size-wide immediate (16 or 32 bit) and returns it
// with the byte width it was read at (for display).
func (d *dec) immOsz() (uint32, int) {
	if d.opsize == 32 {
		return d.imm32(), 4
	}
	return uint32(d.imm16()), 2
}

// jrel reads a signed size-byte (1/2/4) relative displacement and returns the
// absolute linear target (address of the following instruction + displacement).
func (d *dec) jrel(size int) uint32 {
	var rel int32
	switch size {
	case 1:
		rel = int32(int8(d.next()))
	case 2:
		rel = int32(int16(d.imm16()))
	default:
		rel = int32(d.imm32())
	}
	return d.addr + uint32(d.p) + uint32(rel)
}

// run consumes the prefix bytes then dispatches the opcode.
func (d *dec) run() Inst {
	if d.defSize == 32 {
		d.opsize, d.addrsize = 32, 32
	} else {
		d.opsize, d.addrsize = 16, 16
	}
	// alt is the operand/address size a 0x66/0x67 prefix selects: the opposite of
	// the current default (32-bit code toggles down to 16, real-mode up to 32).
	alt := 32
	if d.defSize == 32 {
		alt = 16
	}
	for {
		if d.p >= len(d.mem) {
			d.bad = true
			return Inst{}
		}
		op := d.mem[d.p]
		switch op {
		case 0x26, 0x2E, 0x36, 0x3E, 0x64, 0x65:
			d.seg = segOverride(op) + ":"
			d.p++
			continue
		case 0x66:
			d.opsize = alt
			d.p++
			continue
		case 0x67:
			d.addrsize = alt
			d.p++
			continue
		case 0xF0:
			d.lock = true
			d.prefixText = "LOCK "
			d.p++
			continue
		case 0xF2:
			d.rep = "REPNE "
			d.p++
			continue
		case 0xF3:
			d.rep = "REP "
			d.p++
			continue
		}
		break
	}
	op := d.next()
	if op == 0x0F {
		return d.twoByte(d.next())
	}
	return d.oneByte(op)
}

func segOverride(op byte) string {
	switch op {
	case 0x26:
		return "ES"
	case 0x2E:
		return "CS"
	case 0x36:
		return "SS"
	case 0x3E:
		return "DS"
	case 0x64:
		return "FS"
	default:
		return "GS"
	}
}

// --- small formatting helpers ---------------------------------------------

// mk builds a sequential-flow instruction.
func mk(mnem, text string) Inst { return Inst{Mnem: mnem, Text: text, Flow: FlowSeq} }

// op1 renders "MNEM a".
func op1(mnem, a string) Inst { return mk(mnem, mnem+" "+a) }

// op2 renders "MNEM a, b".
func op2(mnem, a, b string) Inst { return mk(mnem, mnem+" "+a+", "+b) }

// gpr returns the general-purpose register name for index i at bit-width size.
func gpr(i byte, size int) string {
	switch size {
	case 8:
		return reg8[i]
	case 32:
		return reg32[i]
	default:
		return reg16[i]
	}
}

// hexImm formats an immediate value at the given byte width.
func hexImm(v uint32, width int) string {
	switch width {
	case 1:
		return fmt.Sprintf("$%02X", v&0xFF)
	case 4:
		return fmt.Sprintf("$%08X", v)
	default:
		return fmt.Sprintf("$%04X", v&0xFFFF)
	}
}
