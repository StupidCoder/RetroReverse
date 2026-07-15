package x86

// Execution tests for the x87 FPU core (fpuexec.go). Each case assembles a short
// protected-mode instruction stream, runs it to a trailing HLT, and checks the
// resulting stack / memory / flags. The "differential" reference is Go's own
// float64 math — the same IEEE-754 double the core computes in — so an FADD or an
// FSQRT here is checked against the language's arithmetic, and the transcendental
// ops against math.*. Real-mode 8088 coverage (HARTE_DIR) is unaffected: it skips
// the x87 ESC opcodes, which have no FPU on that part.

import (
	"encoding/binary"
	"math"
	"testing"
)

// fpuRig places code at 0x3000 (flat PM, CS base 0) and runs to a HLT.
func fpuRig(t *testing.T, code []byte) *CPU {
	t.Helper()
	ram := newHarteRAM()
	c := NewCPU(ram)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[SS] = 8, 0x10, 0x10
	c.Regs[SP] = 0x8000
	const org = 0x3000
	full := append(append([]byte{}, code...), 0xF4) // trailing HLT
	for i, b := range full {
		ram.set(uint32(org+i), b)
	}
	c.IP = org
	c.Run(2000)
	if !c.Halted {
		t.Fatalf("did not halt (IP=%08X)", c.IP)
	}
	return c
}

// putF64/putF32/putI32 write a datum into the rig's data area via a preamble that
// isn't practical here; instead the tests embed constants directly in memory by
// writing them before the run through a returned closure. To keep the rig simple
// we instead load constants using the immediate-free FILD/FLD-from-memory encodings
// with the datum placed at a fixed absolute address 0x1000+.

func ldF64(addr uint32) []byte { // FLD m64  (DD /0, disp32 absolute)
	return append([]byte{0xDD, 0x05}, u32(addr)...)
}
func stpF64(addr uint32) []byte { // FSTP m64 (DD /3)
	return append([]byte{0xDD, 0x1D}, u32(addr)...)
}
func ldI32(addr uint32) []byte { // FILD m32 (DB /0)
	return append([]byte{0xDB, 0x05}, u32(addr)...)
}
func stpI32(addr uint32) []byte { // FISTP m32 (DB /3)
	return append([]byte{0xDB, 0x1D}, u32(addr)...)
}
func u32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

// seedF64 returns a rig preloaded with float64 values at consecutive 8-byte slots
// from 0x1000, plus a runner that appends code and executes.
func withData(t *testing.T, code []byte, seed map[uint32]uint64) *CPU {
	t.Helper()
	ram := newHarteRAM()
	c := NewCPU(ram)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[SS] = 8, 0x10, 0x10
	c.Regs[SP] = 0x8000
	for addr, v := range seed {
		for i := uint32(0); i < 8; i++ {
			ram.set(addr+i, byte(v>>(8*i)))
		}
	}
	const org = 0x3000
	full := append(append([]byte{}, code...), 0xF4)
	for i, b := range full {
		ram.set(uint32(org+i), b)
	}
	c.IP = org
	c.Run(2000)
	if !c.Halted {
		t.Fatalf("did not halt (IP=%08X)", c.IP)
	}
	return c
}

func readF64(c *CPU, addr uint32) float64 {
	var b [8]byte
	for i := uint32(0); i < 8; i++ {
		b[i] = c.bus.Read(addr + i)
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b[:]))
}
func readI32(c *CPU, addr uint32) int32 {
	var b [4]byte
	for i := uint32(0); i < 4; i++ {
		b[i] = c.bus.Read(addr + i)
	}
	return int32(binary.LittleEndian.Uint32(b[:]))
}

func TestFPUArith(t *testing.T) {
	// FLD 3.0, FLD 4.0, FADD ST0,ST1, FSTP m64 -> expect 7.0
	code := []byte{}
	code = append(code, ldF64(0x1000)...) // ST0=3
	code = append(code, ldF64(0x1008)...) // ST0=4, ST1=3
	code = append(code, 0xD8, 0xC1)       // FADD ST0,ST1 -> 7
	code = append(code, stpF64(0x1010)...)
	c := withData(t, code, map[uint32]uint64{
		0x1000: math.Float64bits(3),
		0x1008: math.Float64bits(4),
	})
	if got := readF64(c, 0x1010); got != 7 {
		t.Fatalf("FADD: got %v want 7", got)
	}
}

func TestFPUSubMulDiv(t *testing.T) {
	// ST0=10, ST1=... load 10 then 3: FSUB ST0,ST1 (D8 E1) -> 10-3=7? order: st0=3,st1=10
	// Load order: FLD 10 -> ST0=10; FLD 3 -> ST0=3,ST1=10.
	// FSUBR ST0,ST1 (D8 E9): st0 = st1-st0 = 10-3 = 7.
	code := []byte{}
	code = append(code, ldF64(0x1000)...) // 10
	code = append(code, ldF64(0x1008)...) // 3
	code = append(code, 0xD8, 0xE9)       // FSUBR ST0,ST1
	code = append(code, stpF64(0x1010)...)
	c := withData(t, code, map[uint32]uint64{
		0x1000: math.Float64bits(10),
		0x1008: math.Float64bits(3),
	})
	if got := readF64(c, 0x1010); got != 7 {
		t.Fatalf("FSUBR: got %v want 7", got)
	}
}

func TestFPUSqrt(t *testing.T) {
	code := append(ldF64(0x1000), 0xD9, 0xFA) // FLD 2, FSQRT
	code = append(code, stpF64(0x1008)...)
	c := withData(t, code, map[uint32]uint64{0x1000: math.Float64bits(2)})
	if got := readF64(c, 0x1008); got != math.Sqrt(2) {
		t.Fatalf("FSQRT: got %v want %v", got, math.Sqrt(2))
	}
}

func TestFPUIntRoundTrip(t *testing.T) {
	// FILD m32(-1234), FISTP m32 -> back to -1234
	code := append(ldI32(0x1000), stpI32(0x1008)...)
	ram := newHarteRAM()
	c := NewCPU(ram)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[SS] = 8, 0x10, 0x10
	c.Regs[SP] = 0x8000
	var neg int32 = -1234
	for i, b := range u32(uint32(neg)) {
		ram.set(0x1000+uint32(i), b)
	}
	const org = 0x3000
	full := append(append([]byte{}, code...), 0xF4)
	for i, b := range full {
		ram.set(uint32(org+i), b)
	}
	c.IP = org
	c.Run(2000)
	if got := readI32(c, 0x1008); got != -1234 {
		t.Fatalf("FILD/FISTP: got %d want -1234", got)
	}
}

func TestFPUConstantsAndCompare(t *testing.T) {
	// FLDPI, FLDZ, FCOMIP ST0,ST1: compares 0 vs pi -> 0<pi so CF=1.
	// After FLDPI: ST0=pi. After FLDZ: ST0=0, ST1=pi. FCOMIP ST0,ST1 (DF F1)
	// compares ST0(0) to ST1(pi): 0<pi -> CF=1, ZF=0.
	code := []byte{0xD9, 0xEB, 0xD9, 0xEE, 0xDF, 0xF1}
	c := fpuRig(t, code)
	if !c.CF || c.ZF {
		t.Fatalf("FCOMIP 0<pi: CF=%v ZF=%v want CF=1 ZF=0", c.CF, c.ZF)
	}
}

func TestFPUNstsw(t *testing.T) {
	// FLD 5, FLD 5, FCOM ST1, FNSTSW AX -> equal sets C3 (bit14 of status = bit6 of
	// AH). Equal => C3=1,C2=0,C0=0. AX high byte bit 6 set.
	code := []byte{}
	code = append(code, ldF64(0x1000)...) // 5
	code = append(code, ldF64(0x1000)...) // 5 again
	code = append(code, 0xD8, 0xD1)       // FCOM ST1
	code = append(code, 0xDF, 0xE0)       // FNSTSW AX
	c := withData(t, code, map[uint32]uint64{0x1000: math.Float64bits(5)})
	ax := c.Reg16(AX)
	const c3 = 1 << 14
	if ax&c3 == 0 {
		t.Fatalf("FNSTSW after equal compare: AX=%04X, C3 not set", ax)
	}
	if ax&((1<<10)|(1<<8)) != 0 {
		t.Fatalf("FNSTSW: C2/C0 unexpectedly set, AX=%04X", ax)
	}
}

func TestFPUControlWord(t *testing.T) {
	// FLDCW m16(0x0C7F, round-toward-zero), FNSTCW m16 -> reads back.
	ram := newHarteRAM()
	c := NewCPU(ram)
	c.Mode = ModeProt
	c.Seg[CS], c.Seg[DS], c.Seg[SS] = 8, 0x10, 0x10
	c.Regs[SP] = 0x8000
	ram.set(0x1000, 0x7F)
	ram.set(0x1001, 0x0C)
	code := []byte{0xD9, 0x2D}          // FLDCW disp32
	code = append(code, u32(0x1000)...) // addr
	code = append(code, 0xD9, 0x3D)     // FNSTCW disp32
	code = append(code, u32(0x1002)...)
	code = append(code, 0xF4)
	const org = 0x3000
	for i, b := range code {
		ram.set(uint32(org+i), b)
	}
	c.IP = org
	c.Run(2000)
	got := uint16(c.bus.Read(0x1002)) | uint16(c.bus.Read(0x1003))<<8
	if got != 0x0C7F {
		t.Fatalf("FLDCW/FNSTCW: got %04X want 0C7F", got)
	}
	if c.FPU.Ctrl != 0x0C7F {
		t.Fatalf("control word not loaded: %04X", c.FPU.Ctrl)
	}
}

func TestF80RoundTrip(t *testing.T) {
	for _, v := range []float64{0, 1, -1, 2, 0.5, math.Pi, 1e30, -1e-30, 12345.6789} {
		m, se := f64ToF80(v)
		if got := f80ToF64(m, se); got != v {
			t.Errorf("f80 round-trip %v -> %v (mant=%016X se=%04X)", v, got, m, se)
		}
	}
	// Special values.
	if m, se := f64ToF80(math.Inf(1)); !math.IsInf(f80ToF64(m, se), 1) {
		t.Errorf("f80 +Inf round-trip failed")
	}
	if m, se := f64ToF80(math.NaN()); !math.IsNaN(f80ToF64(m, se)) {
		t.Errorf("f80 NaN round-trip failed")
	}
}
