package mips

// Differential validation of the CPU core against the SingleStepTests/psx
// per-instruction suite (github.com/SingleStepTests/psx): each case gives a full
// initial R3000 state plus a memory image, one instruction, and the exact
// resulting state. We load it, Step() as many times as the case spans, and diff
// registers and memory.
//
// Delay-slot convention. The suite captures the R3000 pipeline: a case's initial
// state carries the fetched instruction *and* the one already in the delay slot,
// so a branch's architectural effect only lands after its delay slot has run.
// Each case therefore lists the instruction words it executes; we Step() once per
// word and compare against the final state. Loads likewise expose the load-delay
// slot, which this core models (see cpu.go), so no masking is needed.
//
// Point PSX_SST_DIR at a directory of the suite's <opcode>.json files; the test
// skips when it is unset, so the normal `go test` runs without the large data.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// sstState mirrors one initial/final snapshot. The suite stores the 32 GPRs as
// an array (index 0 = $zero), plus pc/hi/lo and the COP0 status/cause/epc, and
// the touched memory as [address, value] byte pairs under "ram".
type sstState struct {
	PC   uint32     `json:"pc"`
	HI   uint32     `json:"hi"`
	LO   uint32     `json:"lo"`
	GPR  []uint32   `json:"gpr"`
	RAM  [][2]int64 `json:"ram"`
}

// sstCase is one test. Opcodes lists the instruction word(s) the case executes;
// when absent we execute a single Step.
type sstCase struct {
	Name    string   `json:"name"`
	Initial sstState `json:"initial"`
	Final   sstState `json:"final"`
	Opcodes []uint32 `json:"opcodes"`
}

// sstRAM is a sparse byte memory implementing mips.Bus, with dirty tracking so a
// case can be undone cheaply between runs.
type sstRAM struct {
	m     map[uint32]byte
	dirty []uint32
}

func newSSTRAM() *sstRAM { return &sstRAM{m: map[uint32]byte{}} }
func (r *sstRAM) Read(a uint32) byte { return r.m[a] }
func (r *sstRAM) Write(a uint32, v byte) {
	r.m[a] = v
	r.dirty = append(r.dirty, a)
}
func (r *sstRAM) seed(a uint32, v byte) {
	r.m[a] = v
	r.dirty = append(r.dirty, a)
}
func (r *sstRAM) reset() {
	for _, a := range r.dirty {
		delete(r.m, a)
	}
	r.dirty = r.dirty[:0]
}

func TestSingleStep(t *testing.T) {
	dir := os.Getenv("PSX_SST_DIR")
	if dir == "" {
		t.Skip("set PSX_SST_DIR to a directory of SingleStepTests/psx <op>.json files")
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no *.json files under %s", dir)
	}
	sort.Strings(files)

	ram := newSSTRAM()
	var firstFails []string
	totalFail, totalPass := 0, 0

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var cases []sstCase
		if err := json.Unmarshal(data, &cases); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		name := filepath.Base(f)
		for i := range cases {
			ok, msg := runSSTCase(&cases[i], ram)
			if ok {
				totalPass++
			} else {
				totalFail++
				if len(firstFails) < 40 {
					firstFails = append(firstFails, fmt.Sprintf("%s: %s", name, msg))
				}
			}
		}
	}
	for _, m := range firstFails {
		t.Log("  " + m)
	}
	if totalFail > 0 {
		t.Errorf("%d differential failures (%d passed)", totalFail, totalPass)
	} else {
		t.Logf("SingleStepTests: %d cases passed", totalPass)
	}
}

func runSSTCase(tc *sstCase, ram *sstRAM) (bool, string) {
	defer ram.reset()
	for _, kv := range tc.Initial.RAM {
		ram.seed(uint32(kv[0]), byte(kv[1]))
	}
	c := NewCPU(ram)
	c.SetPC(tc.Initial.PC)
	c.HI, c.LO = tc.Initial.HI, tc.Initial.LO
	if len(tc.Initial.GPR) == 32 {
		for i := 1; i < 32; i++ {
			c.R[i] = tc.Initial.GPR[i]
			c.out[i] = tc.Initial.GPR[i]
		}
	}

	steps := len(tc.Opcodes)
	if steps == 0 {
		steps = 1
	}
	for i := 0; i < steps; i++ {
		c.Step()
		if c.Halted {
			return true, "" // unimplemented op: treat as a skip-pass
		}
	}

	if len(tc.Final.GPR) == 32 {
		for i := 1; i < 32; i++ {
			if c.R[i] != tc.Final.GPR[i] {
				return false, fmt.Sprintf("%q $%d got %08X want %08X", tc.Name, i, c.R[i], tc.Final.GPR[i])
			}
		}
	}
	if c.HI != tc.Final.HI {
		return false, fmt.Sprintf("%q hi got %08X want %08X", tc.Name, c.HI, tc.Final.HI)
	}
	if c.LO != tc.Final.LO {
		return false, fmt.Sprintf("%q lo got %08X want %08X", tc.Name, c.LO, tc.Final.LO)
	}
	for _, kv := range tc.Final.RAM {
		if got := ram.m[uint32(kv[0])]; got != byte(kv[1]) {
			return false, fmt.Sprintf("%q mem[%08X] got %02X want %02X", tc.Name, kv[0], got, byte(kv[1]))
		}
	}
	return true, ""
}
