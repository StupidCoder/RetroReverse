package r5900_test

// diff_test.go validates the R5900's integer core against tools/cpu/r4300.
//
// There is no published per-instruction conformance suite for the R5900, so the
// standard route (an external vector set, as tools/cpu/mips and tools/cpu/x86 use)
// is not open. What is available instead is a second, independently written MIPS
// core in this repository that is already trusted: the VR4300. The two chips share
// the whole MIPS III integer surface — the 64-bit register file, the sign-extending
// 32-bit operations, the branch-likely family, the shifts, the multiply and divide
// edge cases, the trapping arithmetic — and the R5900 core here was written as a
// port of that one. Running the same instruction on both from the same state, and
// insisting they agree, checks the port did not lose anything in translation.
//
// It is not a proof of correctness: a mistake in the VR4300 that the R5900 faithfully
// copies would pass. It is a check that the shared surface *is* shared, which is what
// a port can plausibly get wrong, and it covers far more ground than hand-written
// cases could. The instructions the two chips do not share — MMI, the R5900's
// three-operand mult, mfsa/mtsa, movz/movn, COP1, COP2 — are excluded here and
// tested directly in cpu_test.go.

import (
	"math/rand"
	"testing"

	"retroreverse.com/tools/cpu/r4300"
	"retroreverse.com/tools/cpu/r5900"
)

// diffBus serves one instruction word to whichever core fetches it, and a small
// scratch RAM behind that. Loads and stores are excluded from the comparison (the
// two chips disagree on byte order, which is the point of them being different
// packages), so the RAM only exists to keep a stray access from panicking.
type diffBus struct {
	word uint32
	ram  [256]byte
}

func (b *diffBus) Read(addr uint32) byte         { return b.ram[addr&0xFF] }
func (b *diffBus) Write(addr uint32, v byte)     { b.ram[addr&0xFF] = v }
func (b *diffBus) Write32(addr uint32, v uint32) {}

func (b *diffBus) Read32(addr uint32) uint32 {
	// Every fetch in this harness is of the single instruction under test, which
	// both cores fetch from the same address.
	return b.word
}

// The instruction under test is placed here, in KSEG0, so it translates without a
// TLB entry on either core.
const testPC = 0x80000000

// shared is the set of encodings both chips implement identically. Anything absent
// is either R5900-only (and covered by cpu_test.go) or VR4300-only.
type gen func(r *rand.Rand) uint32

func rr(r *rand.Rand) (rs, rt, rd uint32) {
	return uint32(r.Intn(32)), uint32(r.Intn(32)), uint32(r.Intn(32))
}

// specialFuncts are the SPECIAL functions the two cores share. Deliberately absent:
//
//	0x0A/0x0B  movz/movn      — MIPS IV; the VR4300 raises a reserved-instruction fault
//	0x18/0x19  mult/multu     — generated separately, with rd forced to zero (below)
//	0x1C..0x1F dmult/ddiv     — the R5900 has no 64-bit multiply or divide
//	0x28/0x29  mfsa/mtsa      — the R5900's shift-amount register
//	0x0C/0x0D  syscall/break  — the R5900 core routes syscall through a host hook
var specialFuncts = []uint32{
	0x00, 0x02, 0x03, 0x04, 0x06, 0x07, // shifts
	0x08, 0x09, // jr, jalr
	0x0F,                   // sync
	0x10, 0x11, 0x12, 0x13, // HI/LO moves
	0x14, 0x16, 0x17, // 64-bit variable shifts
	0x1A, 0x1B, // div, divu
	0x20, 0x21, 0x22, 0x23, // add, addu, sub, subu
	0x24, 0x25, 0x26, 0x27, // and, or, xor, nor
	0x2A, 0x2B, // slt, sltu
	0x2C, 0x2D, 0x2E, 0x2F, // dadd, daddu, dsub, dsubu
	0x30, 0x31, 0x32, 0x33, 0x34, 0x36, // traps
	0x38, 0x3A, 0x3B, 0x3C, 0x3E, 0x3F, // 64-bit shifts
}

// regimmRts are the REGIMM functions the two cores share. Absent: 0x18/0x19
// (mtsab/mtsah), which are R5900-only.
var regimmRts = []uint32{
	0x00, 0x01, 0x02, 0x03, // bltz, bgez, bltzl, bgezl
	0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0E, // immediate traps
	0x10, 0x11, 0x12, 0x13, // the linking forms
}

// primaryOps are the I-type and jump opcodes the two cores share.
var primaryOps = []uint32{
	0x02, 0x03, // j, jal
	0x04, 0x05, 0x06, 0x07, // beq, bne, blez, bgtz
	0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, // addi..lui
	0x14, 0x15, 0x16, 0x17, // the branch-likely family
	0x18, 0x19, // daddi, daddiu
}

var gens = []gen{
	// SPECIAL
	func(r *rand.Rand) uint32 {
		rs, rt, rd := rr(r)
		f := specialFuncts[r.Intn(len(specialFuncts))]
		return rs<<21 | rt<<16 | rd<<11 | uint32(r.Intn(32))<<6 | f
	},
	// mult / multu, with rd forced to zero. The R5900 writes LO into rd as well as
	// the accumulator; with rd == 0 that write is discarded and the two chips agree.
	func(r *rand.Rand) uint32 {
		rs, rt, _ := rr(r)
		f := uint32(0x18)
		if r.Intn(2) == 1 {
			f = 0x19
		}
		return rs<<21 | rt<<16 | f
	},
	// REGIMM
	func(r *rand.Rand) uint32 {
		rs, _, _ := rr(r)
		rt := regimmRts[r.Intn(len(regimmRts))]
		return 0x01<<26 | rs<<21 | rt<<16 | uint32(r.Intn(1<<16))
	},
	// I-type and jumps
	func(r *rand.Rand) uint32 {
		rs, rt, _ := rr(r)
		op := primaryOps[r.Intn(len(primaryOps))]
		return op<<26 | rs<<21 | rt<<16 | uint32(r.Intn(1<<16))
	},
}

func TestAgainstR4300(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	const iterations = 200000
	for i := 0; i < iterations; i++ {
		w := gens[r.Intn(len(gens))](r)

		aBus := &diffBus{word: w}
		bBus := &diffBus{word: w}
		a := r4300.NewCPU(aBus) // the reference
		b := r5900.NewCPU(bBus) // the core under test

		// Identical starting state on both. Values are drawn to land on the awkward
		// cases as well as the ordinary ones: the sign boundaries of 32 and 64 bits,
		// zero, and -1, which is where the sign-extension and divide rules bite.
		var regs [32]uint64
		for k := 1; k < 32; k++ {
			regs[k] = interesting(r)
			a.SetReg(uint32(k), regs[k])
			b.SetReg(uint32(k), regs[k])
		}
		hi, lo := interesting(r), interesting(r)
		a.HI, a.LO = hi, lo
		b.HI, b.LO = hi, lo

		a.SetPC(testPC)
		b.SetPC(testPC)
		// Clear the reset Status bits so neither core is in error level, and both use
		// the RAM exception vectors.
		a.COP0[r4300.Cop0Status] = 0
		b.COP0[r5900.Cop0Status] = 0

		a.Step()
		b.Step()

		as, bs := a.Snapshot(), b.Snapshot()

		mismatch := func(what string, want, got uint64) {
			t.Fatalf("%s mismatch on %s (word 0x%08X, iteration %d)\n  r4300 = 0x%016X\n  r5900 = 0x%016X\n  %s",
				what, r5900.DecodeWord(w, testPC).Text, w, i, want, got, dumpDiff(regs, hi, lo))
		}

		for k := 0; k < 32; k++ {
			if as.R[k] != bs.R[k].Lo {
				mismatch("$"+regNames[k], as.R[k], bs.R[k].Lo)
			}
			// A 64-bit MIPS operation must leave the R5900's upper 64 bits alone.
			// Nothing in the shared surface may disturb them.
			if bs.R[k].Hi != 0 {
				t.Fatalf("%s wrote the upper 64 bits of $%s (0x%016X) — no shared MIPS instruction may",
					r5900.DecodeWord(w, testPC).Text, regNames[k], bs.R[k].Hi)
			}
		}
		if as.HI != bs.HI {
			mismatch("HI", as.HI, bs.HI)
		}
		if as.LO != bs.LO {
			mismatch("LO", as.LO, bs.LO)
		}
		if as.PC != bs.PC {
			mismatch("PC", as.PC, bs.PC)
		}
		if as.NextPC != bs.NextPC {
			mismatch("nextPC", as.NextPC, bs.NextPC)
		}
	}
}

// interesting draws a register value biased toward the boundaries where the
// sign-extension, overflow and divide rules differ from the naive implementation.
func interesting(r *rand.Rand) uint64 {
	switch r.Intn(8) {
	case 0:
		return 0
	case 1:
		return ^uint64(0) // -1: the divide overflow case, and every shift's sign
	case 2:
		return 0x7FFFFFFF // INT32_MAX
	case 3:
		return 0xFFFFFFFF80000000 // INT32_MIN, sign-extended
	case 4:
		return 0x7FFFFFFFFFFFFFFF
	case 5:
		return 0x8000000000000000
	case 6:
		// A value whose high half is *not* the sign extension of its low half. This
		// separates a correct 32-bit sra from one that shifts the whole 64-bit
		// register — but only against a reference that gets it right, which is why
		// TestSRAIgnoresHighWord pins the rule directly rather than by diff. Both
		// cores once shifted 64 bits here and agreed with each other perfectly.
		return uint64(r.Uint32())<<32 | uint64(r.Uint32())
	default:
		return uint64(int64(int32(r.Uint32()))) // an ordinary sign-extended 32-bit value
	}
}

var regNames = [32]string{
	"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
	"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
	"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
	"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

func dumpDiff(regs [32]uint64, hi, lo uint64) string {
	s := "starting state:\n"
	for k := 1; k < 32; k++ {
		s += "    $" + regNames[k] + " = " + hex64(regs[k]) + "\n"
	}
	return s + "    HI = " + hex64(hi) + "\n    LO = " + hex64(lo) + "\n"
}

func hex64(v uint64) string {
	const digits = "0123456789ABCDEF"
	b := make([]byte, 18)
	b[0], b[1] = '0', 'x'
	for i := 0; i < 16; i++ {
		b[17-i] = digits[v&0xF]
		v >>= 4
	}
	return string(b)
}
