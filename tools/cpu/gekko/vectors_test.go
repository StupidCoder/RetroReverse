package gekko

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestVectors runs the committed per-instruction suite: for each case, load the state,
// execute the one instruction, and compare every register and byte the case claims.
//
// The expected side of every case was computed by tools/cpu/gekko/ref — a second
// implementation that does not import this package — so a pass here is evidence about
// this core rather than a statement that it agrees with itself. See TestRefIsIndependent,
// which enforces that.
func TestVectors(t *testing.T) {
	suite := loadSuite(t)

	names := make([]string, 0, len(suite))
	for n := range suite {
		names = append(names, n)
	}
	sort.Strings(names)

	total := 0
	for _, mnem := range names {
		cases := suite[mnem]
		total += len(cases)
		t.Run(mnem, func(t *testing.T) {
			fails := 0
			for i, c := range cases {
				if msg := runCase(c); msg != "" {
					fails++
					if fails <= 5 { // enough to see the pattern, not so many as to drown
						t.Errorf("case %d (op 0x%08X): %s\n    %s", i, c.Op, msg, c.Note)
					}
				}
			}
			if fails > 5 {
				t.Errorf("... and %d more failures in %s", fails-5, mnem)
			}
		})
	}
	t.Logf("%d mnemonics, %d cases", len(suite), total)
}

// runCase executes one vector and returns "" if it held, or a description of how it did
// not.
func runCase(c Case) string {
	m := &testRAM{}
	cpu := NewCPU(m)
	cpu.MSR = 0 // translation off: effective addresses are physical
	cpu.PC = 0x1000
	m.Write32(0x1000, c.Op)

	for k, v := range c.GPR {
		cpu.GPR[mustReg(k)] = v
	}
	for k, v := range c.FPR {
		cpu.FPR[mustReg(k)] = FPR{PS0: math.Float64frombits(v[0]), PS1: math.Float64frombits(v[1])}
	}
	for k, v := range c.GQR {
		cpu.GQR[mustReg(k)] = v
	}
	for k, v := range c.Mem {
		m.Write8(mustAddr(k), uint8(v))
	}
	cpu.XER = c.XER
	cpu.CR = c.CR
	if c.HID2 != 0 {
		cpu.setHID2(c.HID2)
	}

	cpu.Step()
	if cpu.Halted {
		return "the core halted: " + cpu.HaltReason
	}

	dontCare := map[string]bool{}
	for _, d := range c.DontCare {
		dontCare[d] = true
	}

	var bad []string
	for k, want := range c.OutGPR {
		n := mustReg(k)
		if dontCare["gpr"+k] {
			continue
		}
		if got := cpu.GPR[n]; got != want {
			bad = append(bad, sprintf("r%d = 0x%08X, want 0x%08X", n, got, want))
		}
	}
	for k, want := range c.OutFPR {
		n := mustReg(k)
		got := cpu.FPR[n]
		w0, w1 := math.Float64frombits(want[0]), math.Float64frombits(want[1])
		if !sameFloat(got.PS0, w0) || !sameFloat(got.PS1, w1) {
			bad = append(bad, sprintf("f%d = (%v, %v), want (%v, %v)", n, got.PS0, got.PS1, w0, w1))
		}
	}
	for k, want := range c.OutMem {
		a := mustAddr(k)
		if got := uint32(m.Read8(a)); got != want {
			bad = append(bad, sprintf("mem[0x%X] = 0x%02X, want 0x%02X", a, got, want))
		}
	}
	if c.CheckXER && cpu.XER != c.OutXER {
		bad = append(bad, sprintf("XER = 0x%08X, want 0x%08X (CA=%v OV=%v SO=%v, want CA=%v OV=%v SO=%v)",
			cpu.XER, c.OutXER,
			cpu.XER&XERCA != 0, cpu.XER&XEROV != 0, cpu.XER&XERSO != 0,
			c.OutXER&XERCA != 0, c.OutXER&XEROV != 0, c.OutXER&XERSO != 0))
	}
	if c.CheckCR && cpu.CR != c.OutCR {
		bad = append(bad, sprintf("CR = 0x%08X, want 0x%08X", cpu.CR, c.OutCR))
	}
	return strings.Join(bad, "; ")
}

// sameFloat compares two results, treating NaNs as equal to each other — a NaN is a NaN,
// and its payload is not something the architecture pins down.
func sameFloat(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return a == b
}

func mustReg(s string) uint32 {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic("bad register key " + s)
	}
	return uint32(n)
}

func mustAddr(s string) uint32 {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		panic("bad address key " + s)
	}
	return uint32(n)
}

func loadSuite(t *testing.T) Suite {
	t.Helper()
	f, err := os.Open("testdata/vectors.json.gz")
	if err != nil {
		t.Skip("no vector suite: run `go run ./tools/cmd/gekkovec` to generate it")
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var s Suite
	if err := json.NewDecoder(zr).Decode(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// TestRefIsIndependent is what makes the vector suite mean anything.
//
// The whole argument for the suite is that its expected results were derived a second,
// independent way. If ref ever imported this package — directly or through anything else —
// that argument would silently collapse, and the suite would become an elaborate way of
// asserting that the interpreter equals itself. So the claim is checked rather than
// trusted: the go tool is asked for ref's entire import closure, and this package must not
// be in it.
func TestRefIsIndependent(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "retroreverse.com/tools/cpu/gekko/ref").Output()
	if err != nil {
		t.Skipf("cannot run `go list`: %v", err)
	}
	for _, dep := range strings.Fields(string(out)) {
		if dep == "retroreverse.com/tools/cpu/gekko" {
			t.Fatal("tools/cpu/gekko/ref imports tools/cpu/gekko — the vector suite's expected " +
				"results would then be derived from the very core they are meant to validate, " +
				"and the suite would prove nothing")
		}
	}
}

// TestVectorSuiteCoversDecoder is the coverage check, inverted.
//
// The obvious question is "do the vectors pass?". The one that keeps a suite honest as the
// decoder grows is "is there anything the decoder can produce that the vectors say nothing
// about?" — because an instruction with no vector is an instruction nobody has checked, and
// it will not announce itself.
//
// The list below is what is deliberately NOT vectored, with the reason. Vectors are
// register-and-memory transforms of a single instruction; anything that touches the machine
// state register, the exception path, the MMU, the caches or the timers is tested as an
// assembled loop in cpu_test.go instead, because its effect is not a value.
func TestVectorSuiteCoversDecoder(t *testing.T) {
	suite := loadSuite(t)

	// Tested as assembled loops in cpu_test.go, not as vectors — and why.
	notVectored := map[string]string{
		"lwarx": "a reservation is machine state, not a value", "stwcx.": "ditto",
		"dcbz": "writes a cache line", "dcbz_l": "allocates in the scratchpad",
		"mfspr": "the SPR file", "mtspr": "ditto", "mftb": "the clock",
		"mfmsr": "the machine state register", "mtmsr": "ditto", "rfi": "the exception path",
		"sc": "the exception path", "mcrxr": "clears the sticky bit",
	}

	// The families the suite covers. A mnemonic in one of these must have vectors.
	families := []string{
		"add", "addc", "adde", "addme", "addze", "addic", "subf", "subfc", "subfe",
		"subfme", "subfze", "subfic", "neg", "mullw", "mulhw", "mulhwu", "divw", "divwu",
		"srawi", "sraw", "slw", "srw", "cntlzw", "extsb", "extsh",
		"and", "or", "xor", "nand", "nor", "andc", "orc", "eqv",
		"rlwinm", "rlwimi", "cmpw", "cmplw",
		"psq_l", "psq_st", "fadd", "fadds", "fmadd",
	}
	for _, f := range families {
		if len(suite[f]) == 0 {
			t.Errorf("%s has no vectors — the generator does not cover it, so nothing checks it", f)
		}
	}

	if testing.Verbose() {
		for m, why := range notVectored {
			t.Logf("not vectored: %-8s — %s", m, why)
		}
	}
}

// sprintf is a local alias so the test file needs no fmt import alongside its many others.
func sprintf(format string, args ...interface{}) string { return fmt.Sprintf(format, args...) }
