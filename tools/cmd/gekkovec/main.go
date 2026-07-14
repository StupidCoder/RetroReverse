// gekkovec generates the Gekko core's per-instruction test vectors.
//
// There is no published conformance suite for PowerPC — nothing like the SingleStepTests
// data that validates the 8088 and R3000 cores in this repository — so the expected
// results have to be derived. The trap is obvious once stated: derive them from the
// interpreter and the suite proves only that the interpreter agrees with itself.
//
// So this tool imports tools/cpu/gekko/ref and NOT tools/cpu/gekko. ref is a second
// implementation of the same arithmetic, written to be independent rather than fast: it
// derives a carry by asking whether a math/big sum needs a 33rd bit, where the
// interpreter watches a uint32 wrap; it derives an overflow by comparing an exact product
// against its truncation, where the interpreter XORs sign bits; it rotates a 32-element
// bit array, where the interpreter shifts. Where the two agree, the agreement is worth
// something. (tools/cpu/gekko/vectors_test.go asserts the independence by asking the go
// tool for ref's import closure, rather than by trusting this paragraph.)
//
// The operand sets are enumerated rather than sampled wherever the interesting cases are
// sparse — the carry boundaries, the shift edges, and above all the quantiser, whose
// 5 formats × 64 scales × 2 directions are walked exhaustively, because a wrong scale
// direction produces a core that passes every plausible random test and puts a game's
// geometry in the wrong place.
//
// Usage:
//
//	gekkovec [-o DIR] [-n N] [-seed N] [-check]
//
// -check regenerates the vectors and diffs them against the committed file without
// writing, so a later edit to ref cannot silently rewrite its own expectations.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"

	"retroreverse.com/tools/cpu/gekko"
	"retroreverse.com/tools/cpu/gekko/ref"
)

const vectorFile = "vectors.json.gz"

func main() {
	out := flag.String("o", "tools/cpu/gekko/testdata", "directory to write the suite to")
	n := flag.Int("n", 64, "random cases per operand class")
	seed := flag.Int64("seed", 1, "random seed (fixed, so the suite is reproducible)")
	check := flag.Bool("check", false, "regenerate and diff against the committed file; do not write")
	flag.Parse()

	suite := generate(*n, *seed)

	total := 0
	for _, cs := range suite {
		total += len(cs)
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	enc := json.NewEncoder(zw)
	enc.SetIndent("", " ")
	if err := enc.Encode(suite); err != nil {
		fail(err)
	}
	if err := zw.Close(); err != nil {
		fail(err)
	}

	path := filepath.Join(*out, vectorFile)

	if *check {
		old, err := os.ReadFile(path)
		if err != nil {
			fail(fmt.Errorf("-check: %w", err))
		}
		// Compare the decoded suites rather than the bytes: gzip is not required to be
		// reproducible, and a byte diff would fail for reasons that are not about the
		// vectors.
		oldSuite, err := readSuite(old)
		if err != nil {
			fail(err)
		}
		if !equalSuites(oldSuite, suite) {
			fmt.Fprintln(os.Stderr, "gekkovec: the committed vectors DIFFER from what ref now produces.")
			fmt.Fprintln(os.Stderr, "  Either ref changed (and the vectors must be regenerated, deliberately),")
			fmt.Fprintln(os.Stderr, "  or something is wrong. This is the check that stops an edit to ref from")
			fmt.Fprintln(os.Stderr, "  silently rewriting its own expectations.")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "gekkovec: the committed suite matches ref — %d mnemonics, %d cases\n", len(suite), total)
		return
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "gekkovec: wrote %s — %d mnemonics, %d cases, %d KiB\n",
		path, len(suite), total, buf.Len()/1024)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gekkovec:", err)
	os.Exit(1)
}

func readSuite(b []byte) (gekko.Suite, error) {
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var s gekko.Suite
	if err := json.NewDecoder(zr).Decode(&s); err != nil {
		return nil, err
	}
	return s, nil
}

func equalSuites(a, b gekko.Suite) bool {
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return bytes.Equal(ja, jb)
}

// --- Encoders ------------------------------------------------------------------------

func xo(op, d, a, b, x, rc uint32) uint32 { return op<<26 | d<<21 | a<<16 | b<<11 | x<<1 | rc }
func dform(op, d, a, imm uint32) uint32   { return op<<26 | d<<21 | a<<16 | imm&0xFFFF }

// The registers the vectors use. Keeping them fixed and few makes a failure readable.
const (
	rD = 3
	rA = 4
	rB = 5
)

func gpr(m map[string]uint32) map[string]uint32 { return m }

// fp converts a map of float pairs to the bit patterns the vector format stores. JSON
// cannot represent an infinity or a NaN, and a bit-exactness suite should not be routing
// its expected results through decimal anyway.
func fp(m map[string][2]float64) map[string][2]uint64 {
	out := make(map[string][2]uint64, len(m))
	for k, v := range m {
		out[k] = [2]uint64{math.Float64bits(v[0]), math.Float64bits(v[1])}
	}
	return out
}

// interesting is the operand set every integer class is walked over: the boundaries where
// a carry, an overflow or a sign flips. Random values almost never land on these, and
// these are where the bugs are.
var interesting = []uint32{
	0, 1, 2, 0x7FFFFFFE, 0x7FFFFFFF, 0x80000000, 0x80000001, 0xFFFFFFFE, 0xFFFFFFFF,
	0x0000FFFF, 0x00010000, 0x55555555, 0xAAAAAAAA,
}

func generate(n int, seed int64) gekko.Suite {
	rng := rand.New(rand.NewSource(seed))
	s := gekko.Suite{}

	genAddSub(s)
	genMulDiv(s)
	genShiftRotate(s)
	genLogical(s)
	genCompare(s)
	genQuantiser(s) // exhaustive
	genFloat(s, rng, n)

	return s
}

// operandPairs is the cross product of the interesting values, plus some random ones.
func operandPairs(rng *rand.Rand, n int) [][2]uint32 {
	var out [][2]uint32
	for _, a := range interesting {
		for _, b := range interesting {
			out = append(out, [2]uint32{a, b})
		}
	}
	for i := 0; i < n; i++ {
		out = append(out, [2]uint32{rng.Uint32(), rng.Uint32()})
	}
	return out
}

// genAddSub walks the whole carry/overflow family over the boundary operands, with the
// incoming carry and the sticky summary-overflow bit set both ways — because adde reads
// the carry, and CR0's summary bit is copied from XER, so a stale SO changes the answer.
func genAddSub(s gekko.Suite) {
	rng := rand.New(rand.NewSource(7))
	pairs := operandPairs(rng, 16)

	type op struct {
		mnem   string
		x      uint32
		twoOp  bool // reads rB as well as rA
		useCA  bool // reads the carry in
		setsCA bool
		fn     func(a, b uint32, ci bool) ref.Result
	}
	ops := []op{
		{"add", 266, true, false, false, func(a, b uint32, _ bool) ref.Result { return ref.Add(a, b, false) }},
		{"addc", 10, true, false, true, func(a, b uint32, _ bool) ref.Result { return ref.Add(a, b, false) }},
		{"adde", 138, true, true, true, func(a, b uint32, ci bool) ref.Result { return ref.Add(a, b, ci) }},
		{"addme", 234, false, true, true, func(a, _ uint32, ci bool) ref.Result { return ref.Add(a, 0xFFFFFFFF, ci) }},
		{"addze", 202, false, true, true, func(a, _ uint32, ci bool) ref.Result { return ref.Add(a, 0, ci) }},
		{"subf", 40, true, false, false, func(a, b uint32, _ bool) ref.Result { return ref.Sub(a, b, true) }},
		{"subfc", 8, true, false, true, func(a, b uint32, _ bool) ref.Result { return ref.Sub(a, b, true) }},
		{"subfe", 136, true, true, true, func(a, b uint32, ci bool) ref.Result { return ref.Sub(a, b, ci) }},
		{"subfme", 232, false, true, true, func(a, _ uint32, ci bool) ref.Result { return ref.Add(^a, 0xFFFFFFFF, ci) }},
		{"subfze", 200, false, true, true, func(a, _ uint32, ci bool) ref.Result { return ref.Add(^a, 0, ci) }},
		{"neg", 104, false, false, false, func(a, _ uint32, _ bool) ref.Result { return ref.Neg(a) }},
	}

	for _, o := range ops {
		for _, p := range pairs {
			a, b := p[0], p[1]
			for _, ci := range []bool{false, true} {
				if !o.useCA && ci {
					continue // the carry in is not read; one case is enough
				}
				for _, so := range []bool{false, true} {
					for _, oe := range []bool{false, true} {
						want := o.fn(a, b, ci)

						xer := uint32(0)
						if ci {
							xer |= 1 << 29 // XER[CA]
						}
						if so {
							xer |= 1 << 31 // XER[SO], sticky from some earlier instruction
						}

						// The expected XER: carry if the instruction sets it, overflow
						// only when OE is on, and SO sticky-ORed.
						outXER := uint32(0)
						if o.setsCA && want.CA {
							outXER |= 1 << 29
						}
						if oe && want.OV {
							outXER |= 1 << 30
						}
						if so || (oe && want.OV) {
							outXER |= 1 << 31
						}

						// CR0, from the record bit: the sign of the result, plus the
						// summary-overflow bit copied out of the FINAL XER.
						f := crOf(want.Value)
						if outXER&(1<<31) != 0 {
							f |= ref.SO
						}

						word := xo(31, rD, rA, rB, o.x, 1) // Rc=1: check CR0 too
						if oe {
							word |= 1 << 10
						}

						c := gekko.Case{
							Op:       word,
							GPR:      gpr(map[string]uint32{"4": a, "5": b}),
							XER:      xer,
							OutGPR:   gpr(map[string]uint32{"3": want.Value}),
							OutXER:   outXER,
							OutCR:    f << 28, // CR0 is the top nibble
							CheckXER: true,
							CheckCR:  true,
						}
						s[o.mnem] = append(s[o.mnem], c)
					}
				}
			}
		}
	}

	// subfic and addic: the immediate forms, which set the carry and have no OE at all.
	for _, p := range operandPairs(rng, 8) {
		a := p[0]
		imm := uint32(int32(int16(p[1]))) // a signed 16-bit immediate, sign-extended

		w := ref.Add(a, imm, false)
		s["addic"] = append(s["addic"], gekko.Case{
			Op:       dform(12, rD, rA, p[1]&0xFFFF),
			GPR:      gpr(map[string]uint32{"4": a}),
			OutGPR:   gpr(map[string]uint32{"3": w.Value}),
			OutXER:   caBit(w.CA),
			CheckXER: true,
		})

		// subfic computes imm - rA, which is imm + ¬rA + 1.
		sw := ref.Add(^a, imm, true)
		s["subfic"] = append(s["subfic"], gekko.Case{
			Op:       dform(8, rD, rA, p[1]&0xFFFF),
			GPR:      gpr(map[string]uint32{"4": a}),
			OutGPR:   gpr(map[string]uint32{"3": sw.Value}),
			OutXER:   caBit(sw.CA),
			CheckXER: true,
		})
	}
}

func caBit(on bool) uint32 {
	if on {
		return 1 << 29
	}
	return 0
}

// crOf is the sign classification a record bit records. It is written here in terms of
// ref's comparison rather than Go's, so the generator is not quietly re-deriving it the
// interpreter's way.
func crOf(v uint32) uint32 { return ref.CompareSigned(v, 0) }

func genMulDiv(s gekko.Suite) {
	rng := rand.New(rand.NewSource(11))
	for _, p := range operandPairs(rng, 24) {
		a, b := p[0], p[1]

		m := ref.MulLow(a, b)
		for _, oe := range []bool{false, true} {
			word := xo(31, rD, rA, rB, 235, 1)
			outXER := uint32(0)
			if oe {
				word |= 1 << 10
				if m.OV {
					outXER |= (1 << 30) | (1 << 31)
				}
			}
			f := crOf(m.Value)
			if outXER&(1<<31) != 0 {
				f |= ref.SO
			}
			s["mullw"] = append(s["mullw"], gekko.Case{
				Op: word, GPR: gpr(map[string]uint32{"4": a, "5": b}),
				OutGPR: gpr(map[string]uint32{"3": m.Value}),
				OutXER: outXER, OutCR: f << 28, CheckXER: true, CheckCR: true,
			})
		}

		s["mulhw"] = append(s["mulhw"], gekko.Case{
			Op: xo(31, rD, rA, rB, 75, 0), GPR: gpr(map[string]uint32{"4": a, "5": b}),
			OutGPR: gpr(map[string]uint32{"3": ref.MulHighSigned(a, b)}),
		})
		s["mulhwu"] = append(s["mulhwu"], gekko.Case{
			Op: xo(31, rD, rA, rB, 11, 0), GPR: gpr(map[string]uint32{"4": a, "5": b}),
			OutGPR: gpr(map[string]uint32{"3": ref.MulHighUnsigned(a, b)}),
		})

		// The divisions, including the two cases whose RESULT the architecture declines
		// to define while still defining the overflow. Those carry DontCare, and still
		// assert the flags — the only honest treatment.
		dv, ok := ref.DivSigned(a, b)
		cs := gekko.Case{
			Op:       xo(31, rD, rA, rB, 491|512, 0), // divwo
			GPR:      gpr(map[string]uint32{"4": a, "5": b}),
			OutXER:   ovBits(dv.OV),
			CheckXER: true,
		}
		if ok {
			cs.OutGPR = gpr(map[string]uint32{"3": dv.Value})
		} else {
			cs.DontCare = []string{"gpr3"}
			cs.Note = "the architecture leaves the quotient undefined here, but not the overflow"
		}
		s["divw"] = append(s["divw"], cs)

		du, ok := ref.DivUnsigned(a, b)
		cu := gekko.Case{
			Op:       xo(31, rD, rA, rB, 459|512, 0),
			GPR:      gpr(map[string]uint32{"4": a, "5": b}),
			OutXER:   ovBits(du.OV),
			CheckXER: true,
		}
		if ok {
			cu.OutGPR = gpr(map[string]uint32{"3": du.Value})
		} else {
			cu.DontCare = []string{"gpr3"}
			cu.Note = "divide by zero: the quotient is undefined, the overflow is not"
		}
		s["divwu"] = append(s["divwu"], cu)
	}
}

func ovBits(ov bool) uint32 {
	if ov {
		return (1 << 30) | (1 << 31)
	}
	return 0
}

// genShiftRotate walks every shift amount, not a sample: the edges (0, 31, 32, 63) are
// where the rules live, and a random amount lands on them essentially never.
func genShiftRotate(s gekko.Suite) {
	vals := append([]uint32{}, interesting...)
	vals = append(vals, 0xFFFFFFF0, 0x0000000F, 0x80000001)

	for _, v := range vals {
		// srawi: every shift 0..31. Its carry is the classic trap.
		for sh := uint32(0); sh < 32; sh++ {
			r, ca := ref.Srawi(v, sh)
			f := crOf(r)
			s["srawi"] = append(s["srawi"], gekko.Case{
				Op:       xo(31, rA, rD, sh, 824, 1), // srawi rD,rA,sh — note the operand order
				GPR:      gpr(map[string]uint32{"4": v}),
				OutGPR:   gpr(map[string]uint32{"3": r}),
				OutXER:   caBit(ca),
				OutCR:    f << 28,
				CheckXER: true, CheckCR: true,
			})
		}
		// sraw: the shift comes from a register, so it can exceed 31 — and then the whole
		// word leaves, which is a different rule.
		for _, sh := range []uint32{0, 1, 15, 31, 32, 33, 63, 64, 100} {
			r, ca := ref.Srawi(v, sh&63)
			if sh >= 64 {
				// A shift amount with bit 5 set (>= 32 in the six-bit field) is the
				// "everything leaves" case; above 63 the field wraps.
				r, ca = ref.Srawi(v, sh&63)
			}
			s["sraw"] = append(s["sraw"], gekko.Case{
				Op:       xo(31, rA, rD, rB, 792, 0),
				GPR:      gpr(map[string]uint32{"4": v, "5": sh}),
				OutGPR:   gpr(map[string]uint32{"3": r}),
				OutXER:   caBit(ca),
				CheckXER: true,
			})
		}
		for _, sh := range []uint32{0, 1, 31, 32, 33, 63} {
			s["slw"] = append(s["slw"], gekko.Case{
				Op: xo(31, rA, rD, rB, 24, 0), GPR: gpr(map[string]uint32{"4": v, "5": sh}),
				OutGPR: gpr(map[string]uint32{"3": ref.Slw(v, sh&63)}),
			})
			s["srw"] = append(s["srw"], gekko.Case{
				Op: xo(31, rA, rD, rB, 536, 0), GPR: gpr(map[string]uint32{"4": v, "5": sh}),
				OutGPR: gpr(map[string]uint32{"3": ref.Srw(v, sh&63)}),
			})
		}

		s["cntlzw"] = append(s["cntlzw"], gekko.Case{
			Op: xo(31, rA, rD, 0, 26, 0), GPR: gpr(map[string]uint32{"4": v}),
			OutGPR: gpr(map[string]uint32{"3": ref.CountLeadingZeros(v)}),
		})
	}

	// rlwinm and rlwimi, over a mask set that INCLUDES the wrapped case (mb > me), which
	// is what the compiler uses to extract a field and what a transcription gets wrong.
	masks := [][2]uint32{{0, 31}, {0, 0}, {31, 31}, {0, 15}, {16, 31}, {28, 3}, {24, 7}, {1, 0}}
	for _, v := range []uint32{0x12345678, 0xFFFFFFFF, 0x80000001, 0xAAAAAAAA} {
		for _, sh := range []uint32{0, 1, 8, 16, 31} {
			for _, m := range masks {
				mb, me := m[0], m[1]
				s["rlwinm"] = append(s["rlwinm"], gekko.Case{
					Op:     21<<26 | rD<<21 | rA<<16 | sh<<11 | mb<<6 | me<<1,
					GPR:    gpr(map[string]uint32{"3": v}),
					OutGPR: gpr(map[string]uint32{"4": ref.Rlwinm(v, sh, mb, me)}),
					Note:   "rlwinm rA,rS,sh,mb,me — the mask wraps when mb > me",
				})
				s["rlwimi"] = append(s["rlwimi"], gekko.Case{
					Op:     20<<26 | rD<<21 | rA<<16 | sh<<11 | mb<<6 | me<<1,
					GPR:    gpr(map[string]uint32{"3": v, "4": 0xDEADBEEF}),
					OutGPR: gpr(map[string]uint32{"4": ref.Rlwimi(0xDEADBEEF, v, sh, mb, me)}),
				})
			}
		}
	}
}

func genLogical(s gekko.Suite) {
	rng := rand.New(rand.NewSource(13))
	ops := []struct {
		mnem string
		x    uint32
		fn   func(a, b uint32) uint32
	}{
		{"and", 28, func(a, b uint32) uint32 { return a & b }},
		{"or", 444, func(a, b uint32) uint32 { return a | b }},
		{"xor", 316, func(a, b uint32) uint32 { return a ^ b }},
		{"nand", 476, func(a, b uint32) uint32 { return ^(a & b) }},
		{"nor", 124, func(a, b uint32) uint32 { return ^(a | b) }},
		{"andc", 60, func(a, b uint32) uint32 { return a &^ b }},
		{"orc", 412, func(a, b uint32) uint32 { return a | ^b }},
		{"eqv", 284, func(a, b uint32) uint32 { return ^(a ^ b) }},
	}
	for _, o := range ops {
		for _, p := range operandPairs(rng, 8) {
			a, b := p[0], p[1]
			r := o.fn(a, b)
			s[o.mnem] = append(s[o.mnem], gekko.Case{
				// The logicals write rA and read rS/rB — the operand order is reversed
				// from the arithmetic, which is itself a thing to get wrong.
				Op:      xo(31, rD, rA, rB, o.x, 1),
				GPR:     gpr(map[string]uint32{"3": a, "5": b}),
				OutGPR:  gpr(map[string]uint32{"4": r}),
				OutCR:   crOf(r) << 28,
				CheckCR: true,
			})
		}
	}
	for _, v := range interesting {
		s["extsb"] = append(s["extsb"], gekko.Case{
			Op: xo(31, rD, rA, 0, 954, 0), GPR: gpr(map[string]uint32{"3": v}),
			OutGPR: gpr(map[string]uint32{"4": uint32(int32(int8(v)))}),
		})
		s["extsh"] = append(s["extsh"], gekko.Case{
			Op: xo(31, rD, rA, 0, 922, 0), GPR: gpr(map[string]uint32{"3": v}),
			OutGPR: gpr(map[string]uint32{"4": uint32(int32(int16(v)))}),
		})
	}
}

// genCompare walks the compares over the boundaries — and with the sticky summary bit set
// both ways, because a compare copies it into the field it writes.
func genCompare(s gekko.Suite) {
	for _, a := range interesting {
		for _, b := range interesting {
			for _, so := range []bool{false, true} {
				xer := uint32(0)
				soBit := uint32(0)
				if so {
					xer = 1 << 31
					soBit = ref.SO
				}
				// Into CR field 3, so that a bug that always writes field 0 is caught.
				const crf = 3
				fS := ref.CompareSigned(a, b) | soBit
				s["cmpw"] = append(s["cmpw"], gekko.Case{
					Op:      31<<26 | crf<<23 | rA<<16 | rB<<11 | 0<<1,
					GPR:     gpr(map[string]uint32{"4": a, "5": b}),
					XER:     xer,
					OutCR:   fS << (28 - 4*crf),
					CheckCR: true,
				})
				fU := ref.CompareUnsigned(a, b) | soBit
				s["cmplw"] = append(s["cmplw"], gekko.Case{
					Op:      31<<26 | crf<<23 | rA<<16 | rB<<11 | 32<<1,
					GPR:     gpr(map[string]uint32{"4": a, "5": b}),
					XER:     xer,
					OutCR:   fU << (28 - 4*crf),
					CheckCR: true,
				})
			}
		}
	}
}

// genQuantiser is exhaustive, and deliberately so. Five formats × sixty-four scales ×
// both directions × the W bit is 1,280 combinations, and a wrong scale DIRECTION — a load
// that multiplies where it should divide — passes every plausible random test and puts a
// game's geometry in the wrong place. This is the densest bug farm in the core, so it gets
// walked rather than sampled.
func genQuantiser(s gekko.Suite) {
	const base = 0x2000
	formats := []uint32{ref.QFloat, ref.QU8, ref.QU16, ref.QS8, ref.QS16}
	values := []float64{0, 1, -1, 2.5, -2.5, 3, 100, -100, 0.5, 127, -128, 1000, -1000}

	for _, typ := range formats {
		for scaleRaw := uint32(0); scaleRaw < 64; scaleRaw++ {
			scale := int32(scaleRaw<<26) >> 26 // the signed six-bit value
			for _, v := range values {
				// --- The store direction: multiply by 2^scale, truncate, saturate.
				stored := ref.Quantize(v, typ, scale)
				gqrStore := typ | (scaleRaw << 8)
				outMem := map[string]uint32{}
				for i := 0; i < ref.Size(typ); i++ {
					shift := uint(8 * (ref.Size(typ) - 1 - i))
					outMem[fmt.Sprintf("%X", base+i)] = (stored >> shift) & 0xFF
				}
				s["psq_st"] = append(s["psq_st"], gekko.Case{
					Op:     60<<26 | 0<<21 | rA<<16 | 1<<15 | 0<<12, // psq_st f0,0(r4),1,gqr0
					GPR:    gpr(map[string]uint32{"4": base}),
					FPR:    fp(map[string][2]float64{"0": {v, 0}}),
					GQR:    map[string]uint32{"0": gqrStore},
					HID2:   0x20000000, // HID2[PSE]
					OutMem: outMem,
					Note:   fmt.Sprintf("store %v as type %d scale %d (multiplies by 2^scale, saturates)", v, typ, scale),
				})

				// --- The load direction: divide by 2^scale. Feed it the value the store
				// would have produced, so a reversed direction cannot cancel out.
				mem := map[string]uint32{}
				for i := 0; i < ref.Size(typ); i++ {
					shift := uint(8 * (ref.Size(typ) - 1 - i))
					mem[fmt.Sprintf("%X", base+i)] = (stored >> shift) & 0xFF
				}
				loaded := ref.Dequantize(stored, typ, scale)
				gqrLoad := (typ << 16) | (scaleRaw << 24)
				s["psq_l"] = append(s["psq_l"], gekko.Case{
					Op:  56<<26 | 1<<21 | rA<<16 | 1<<15 | 0<<12, // psq_l f1,0(r4),1,gqr0
					GPR: gpr(map[string]uint32{"4": base}),
					GQR: map[string]uint32{"0": gqrLoad},
					Mem: mem,
					// A one-value load sets PS1 to 1.0 — not zero, and not left alone.
					OutFPR: fp(map[string][2]float64{"1": {loaded, 1.0}}),
					HID2:   0x20000000,
					Note:   fmt.Sprintf("load type %d scale %d (divides by 2^scale); PS1 becomes 1.0", typ, scale),
				})
			}
		}
	}
}

// genFloat checks the two numerical claims that actually matter: that fmadd fuses (one
// rounding of the exact product-plus-addend), and that the single-precision instructions
// round twice — once to double, once to single.
func genFloat(s gekko.Suite, rng *rand.Rand, n int) {
	vals := []float64{
		0, 1, -1, 0.5, 2, 3, 1e10, 1e-10,
		1.0000000000000002, 0.9999999999999999,
		math.Pi, math.SmallestNonzeroFloat64, math.MaxFloat32,
	}
	for i := 0; i < n; i++ {
		vals = append(vals, rng.NormFloat64()*math.Pow(2, float64(rng.Intn(40)-20)))
	}

	for _, a := range vals {
		for _, b := range vals {
			// fadd / fmul, in double: plain IEEE, but a check that PS1 is preserved.
			s["fadd"] = append(s["fadd"], gekko.Case{
				Op:     63<<26 | 3<<21 | 1<<16 | 2<<11 | 21<<1,
				FPR:    fp(map[string][2]float64{"1": {a, 111}, "2": {b, 222}, "3": {0, 999}}),
				OutFPR: fp(map[string][2]float64{"3": {a + b, 999}}),
				Note:   "a scalar double op must leave PS1 alone",
			})
			// fadds: computed in double, rounded once to single, and BROADCAST to both
			// halves of the register.
			r := ref.RoundToSingle(a + b)
			s["fadds"] = append(s["fadds"], gekko.Case{
				Op:     59<<26 | 3<<21 | 1<<16 | 2<<11 | 21<<1,
				FPR:    fp(map[string][2]float64{"1": {a, 0}, "2": {b, 0}}),
				OutFPR: fp(map[string][2]float64{"3": {r, r}}),
				Note:   "a single-precision result is rounded to single and written to both halves",
			})
		}
	}

	// fmadd: the fused case. ref derives it in 200-bit precision and rounds once; the
	// interpreter uses math.FMA. If they agree, the interpreter really is fusing.
	//
	// But agreeing is not enough for the vectors to MEAN anything. On most operands the
	// fused and the unfused answers are identical, and a suite made only of those passes
	// just as happily against a core that computes a*c + b — which is the bug it exists to
	// catch. (Mutation-testing an earlier version of this generator proved exactly that:
	// breaking fmadd to round twice failed nothing.)
	//
	// So the generator hunts. It keeps operands where the naive a*c + b — computed here in
	// native arithmetic, deliberately, as the wrong answer to be avoided — differs from
	// ref's exact one. Those cases are the suite's teeth, and their count is asserted below
	// so the teeth cannot quietly fall out.
	discriminating := 0
	rngF := rand.New(rand.NewSource(29))
	for i := 0; i < 200000 && discriminating < 400; i++ {
		a := rngF.NormFloat64() * math.Pow(2, float64(rngF.Intn(24)-12))
		c := rngF.NormFloat64() * math.Pow(2, float64(rngF.Intn(24)-12))
		b := rngF.NormFloat64() * math.Pow(2, float64(rngF.Intn(24)-12))

		want := ref.FMA(a, c, b)
		if math.IsNaN(want) || math.IsInf(want, 0) {
			continue
		}
		if want == a*c+b {
			continue // this operand triple cannot tell a fused core from an unfused one
		}
		discriminating++
		s["fmadd"] = append(s["fmadd"], gekko.Case{
			Op:     63<<26 | 3<<21 | 1<<16 | 2<<11 | 4<<6 | 29<<1, // fmadd f3,f1,f4,f2
			FPR:    fp(map[string][2]float64{"1": {a, 0}, "4": {c, 0}, "2": {b, 0}}),
			OutFPR: fp(map[string][2]float64{"3": {want, 0}}),
			Note:   "the fused answer differs from a*c+b here — this case can tell them apart",
		})
	}
	if discriminating < 100 {
		fail(fmt.Errorf("only %d fmadd cases distinguish a fused multiply-add from an unfused one; "+
			"the fmadd vectors would not catch a core that rounds twice", discriminating))
	}

	// And the ordinary cases, for the special values and the plain arithmetic.
	for _, a := range vals {
		for _, c := range vals {
			for _, b := range vals[:6] {
				want := ref.FMA(a, c, b)
				if math.IsNaN(want) {
					continue // ref declines the special cases; they are tested elsewhere
				}
				s["fmadd"] = append(s["fmadd"], gekko.Case{
					Op:     63<<26 | 3<<21 | 1<<16 | 2<<11 | 4<<6 | 29<<1,
					FPR:    fp(map[string][2]float64{"1": {a, 0}, "4": {c, 0}, "2": {b, 0}}),
					OutFPR: fp(map[string][2]float64{"3": {want, 0}}),
					Note:   "one rounding of the exact product-plus-addend, not two",
				})
			}
		}
	}
}
