package arm

import (
	"math"
	"testing"
)

// Encodings hand-assembled from the ARM ARM (VFPv2), independent of the decoder.

func vfpCPU() (*CPU, *flatBus) {
	b := newBus(0x10000)
	c := NewCPU(b)
	c.Arch = V6K
	c.Mode = ModeSYS
	return c, b
}

// VADD.F32 s0, s1, s2 = 0xEE300A81 (single). Note the field packing: s1 is
// Vn=0,N=1 and s2 is Vm=1,M=0 — an easy encoding to get wrong.
func TestVFPSingleArith(t *testing.T) {
	c, b := vfpCPU()
	c.sSet(1, 1.5)
	c.sSet(2, 2.25)
	stepOne(c, b, 0xEE300A81) // VADD.F32 s0, s1, s2
	if got := c.sGet(0); got != 3.75 {
		t.Fatalf("VADD.F32 = %v, want 3.75", got)
	}

	c.sSet(1, 3)
	c.sSet(2, 4)
	stepOne(c, b, 0xEE200A81) // VMUL.F32 s0, s1, s2
	if got := c.sGet(0); got != 12 {
		t.Fatalf("VMUL.F32 = %v, want 12", got)
	}
}

// VADD.F64 d0, d1, d2 = 0xEE310B02 (double). Exercises the S-register pairing.
func TestVFPDoubleArith(t *testing.T) {
	c, b := vfpCPU()
	c.dSet(1, 1.5)
	c.dSet(2, 2.25)
	stepOne(c, b, 0xEE310B02) // VADD.F64 d0, d1, d2
	if got := c.dGet(0); got != 3.75 {
		t.Fatalf("VADD.F64 = %v, want 3.75", got)
	}
	// Confirm the double actually occupies s0 and s1.
	if c.VFP.S[0] == 0 && c.VFP.S[1] == 0 {
		t.Fatal("d0 did not populate its S register pair")
	}
}

// VSQRT.F64 d0, d1 = 0xEEB10BC1. Extension group, opc2=0001 opc3=11.
func TestVFPSqrt(t *testing.T) {
	c, b := vfpCPU()
	c.dSet(1, 2)
	stepOne(c, b, 0xEEB10BC1) // VSQRT.F64 d0, d1
	if got := c.dGet(0); math.Abs(got-math.Sqrt2) > 1e-12 {
		t.Fatalf("VSQRT.F64 = %v, want %v", got, math.Sqrt2)
	}
}

// A high destination register must decode: VADD.F32 s1, s2, s3 puts D=1 into the
// opcode field, which a naive bits[23:20] read would mistake for a different op.
func TestVFPHighRegisterDecodes(t *testing.T) {
	// VADD.F32 s1, s2, s3: Sd=1 -> Vd=0,D=1; Sn=2 -> Vn=1,N=0; Sm=3 -> Vm=1,M=1.
	w := uint32(0xEE710A21)
	code := []byte{byte(w), byte(w >> 8), byte(w >> 16), byte(w >> 24)}
	in := DecodeVariant(code, 0x1000, false, V6K)
	if in.Mnem[:4] != "VADD" {
		t.Fatalf("high-register VADD decoded as %q (bit22=D folded into the opcode?)", in.Mnem)
	}
	c, b := vfpCPU()
	c.sSet(2, 10)
	c.sSet(3, 5)
	stepOne(c, b, w)
	if got := c.sGet(1); got != 15 {
		t.Fatalf("VADD.F32 s1,s2,s3 = %v, want 15", got)
	}
}

// VCVT float->int (truncating) and int->float round trip.
func TestVFPConvert(t *testing.T) {
	c, b := vfpCPU()
	// VCVT.S32.F64 s0, d1 = 0xEEBD0BC1 (to signed int, toward zero).
	c.dSet(1, -3.9)
	stepOne(c, b, 0xEEBD0BC1)
	if got := int32(c.sBits(0)); got != -3 {
		t.Fatalf("VCVT.S32.F64(-3.9) = %d, want -3 (truncate toward zero)", got)
	}
	// VCVT.F64.S32 d2, s0 = 0xEEB82BC0 (signed int -> double).
	stepOne(c, b, 0xEEB82BC0)
	if got := c.dGet(2); got != -3 {
		t.Fatalf("VCVT.F64.S32 = %v, want -3", got)
	}
}

// VCMP + VMRS APSR_nzcv: comparing 2.0 > 1.0 should set the CPSR so a GT branch
// would be taken (C set, Z clear, N==V).
func TestVFPCompareToFlags(t *testing.T) {
	c, b := vfpCPU()
	c.sSet(0, 2)
	c.sSet(1, 1)
	stepOne(c, b, 0xEEB40A60) // VCMP.F32 s0, s1
	stepOne(c, b, 0xEEF1FA10) // VMRS APSR_nzcv, FPSCR
	if c.Z {
		t.Fatal("VCMP 2>1 set Z")
	}
	if !c.C {
		t.Fatal("VCMP 2>1 did not set C")
	}
	if c.N {
		t.Fatal("VCMP 2>1 set N")
	}
}

// VMOV between a core register and a single VFP register, both directions.
func TestVFPCoreMove(t *testing.T) {
	c, b := vfpCPU()
	c.R[3] = 0x40490FDB // ~pi as float32 bits
	stepOne(c, b, 0xEE003A10) // VMOV s0, r3
	if c.sBits(0) != 0x40490FDB {
		t.Fatalf("VMOV s0,r3 = 0x%08X", c.sBits(0))
	}
	stepOne(c, b, 0xEE114A10) // VMOV r4, s2 (s2 is zero)
	if c.R[4] != 0 {
		t.Fatalf("VMOV r4,s2 = 0x%08X, want 0", c.R[4])
	}
}

// VSTR then VLDR round-trips a value through memory.
func TestVFPLoadStore(t *testing.T) {
	c, b := vfpCPU()
	c.R[0] = 0x2000 // data address, clear of the 0x1000 the instructions load at
	c.sSet(0, 6.5)
	stepOne(c, b, 0xED800A00) // VSTR s0, [r0]
	if got := math.Float32frombits(uint32(b.word(0x2000))); got != 6.5 {
		t.Fatalf("VSTR wrote %v", got)
	}
	// VLDR s3, [r0]: Sd=3 -> Vd=1,D=1.
	stepOne(c, b, 0xEDD01A00) // VLDR s3, [r0]
	if got := c.sGet(3); got != 6.5 {
		t.Fatalf("VLDR s3 = %v, want 6.5", got)
	}
}
