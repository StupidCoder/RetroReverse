package arm

import "testing"

// flatBus is a flat little-endian memory for exercising the core.
type flatBus struct{ m []byte }

func newBus(size int) *flatBus { return &flatBus{m: make([]byte, size)} }

func (b *flatBus) Read(a uint32) byte {
	if int(a) < len(b.m) {
		return b.m[a]
	}
	return 0
}
func (b *flatBus) Write(a uint32, v byte) {
	if int(a) < len(b.m) {
		b.m[a] = v
	}
}
func (b *flatBus) word(a uint32) uint32 {
	return uint32(b.m[a]) | uint32(b.m[a+1])<<8 | uint32(b.m[a+2])<<16 | uint32(b.m[a+3])<<24
}

// loadARM writes a sequence of 32-bit ARM words starting at addr.
func (b *flatBus) loadARM(addr uint32, words ...uint32) {
	for i, w := range words {
		p := addr + uint32(i)*4
		b.m[p] = byte(w)
		b.m[p+1] = byte(w >> 8)
		b.m[p+2] = byte(w >> 16)
		b.m[p+3] = byte(w >> 24)
	}
}

// loadThumb writes a sequence of 16-bit Thumb halfwords starting at addr.
func (b *flatBus) loadThumb(addr uint32, halfs ...uint32) {
	for i, h := range halfs {
		p := addr + uint32(i)*2
		b.m[p] = byte(h)
		b.m[p+1] = byte(h >> 8)
	}
}

// runTo single-steps up to a budget, stopping when PC reaches stop.
func runTo(t *testing.T, c *CPU, stop uint32, budget int) {
	t.Helper()
	for i := 0; i < budget; i++ {
		if c.R[15] == stop {
			return
		}
		if c.Halted {
			t.Fatalf("halted: %s", c.HaltReason)
		}
		c.Step()
	}
	t.Fatalf("did not reach 0x%08X within %d steps (pc=0x%08X)", stop, budget, c.R[15])
}

func TestDataProc(t *testing.T) {
	b := newBus(0x1000)
	b.loadARM(0x100,
		0xE3A00005, // MOV r0, #5
		0xE2800003, // ADD r0, r0, #3
		0xE3500008, // CMP r0, #8
	)
	c := NewCPU(b)
	c.Mode = ModeSYS
	c.R[15] = 0x100
	c.Step()
	c.Step()
	if c.R[0] != 8 {
		t.Fatalf("r0 = %d, want 8", c.R[0])
	}
	c.Step() // CMP r0,#8 → equal
	if !c.Z || c.N {
		t.Fatalf("CMP flags: Z=%v N=%v, want Z=true N=false", c.Z, c.N)
	}
}

func TestLoop(t *testing.T) {
	b := newBus(0x1000)
	b.loadARM(0x100,
		0xE3A00000, // 100: MOV r0, #0
		0xE3A01005, // 104: MOV r1, #5
		0xE2800001, // 108: ADD r0, r0, #1   (loop)
		0xE2511001, // 10C: SUBS r1, r1, #1
		0x1AFFFFFC, // 110: BNE 0x108
		0xEAFFFFFE, // 114: B . (halt spin)
	)
	c := NewCPU(b)
	c.Mode = ModeSYS
	c.R[15] = 0x100
	runTo(t, c, 0x114, 1000)
	if c.R[0] != 5 || c.R[1] != 0 {
		t.Fatalf("loop result r0=%d r1=%d, want r0=5 r1=0", c.R[0], c.R[1])
	}
}

func TestLoadStore(t *testing.T) {
	b := newBus(0x1000)
	b.loadARM(0x100,
		0xE3A00042, // MOV r0, #0x42
		0xE3A01C02, // MOV r1, #0x200
		0xE5810000, // STR r0, [r1]
		0xE3A00000, // MOV r0, #0
		0xE5912000, // LDR r2, [r1]
		0xEAFFFFFE, // B .
	)
	c := NewCPU(b)
	c.Mode = ModeSYS
	c.R[15] = 0x100
	runTo(t, c, 0x114, 100)
	if b.word(0x200) != 0x42 {
		t.Fatalf("mem[0x200] = 0x%X, want 0x42", b.word(0x200))
	}
	if c.R[2] != 0x42 {
		t.Fatalf("r2 = 0x%X, want 0x42", c.R[2])
	}
}

func TestPushPop(t *testing.T) {
	b := newBus(0x1000)
	b.loadARM(0x100,
		0xE3A000AA, // MOV r0, #0xAA
		0xE92D0001, // PUSH {r0}   (STMDB sp!, {r0})
		0xE3A00000, // MOV r0, #0
		0xE8BD0002, // POP {r1}    (LDMIA sp!, {r1})
		0xEAFFFFFE, // B .
	)
	c := NewCPU(b)
	c.Mode = ModeSYS
	c.R[15] = 0x100
	c.R[13] = 0x400
	runTo(t, c, 0x110, 100)
	if c.R[1] != 0xAA {
		t.Fatalf("r1 = 0x%X, want 0xAA", c.R[1])
	}
	if c.R[13] != 0x400 {
		t.Fatalf("sp = 0x%X, want 0x400 (balanced)", c.R[13])
	}
}

func TestCallReturn(t *testing.T) {
	b := newBus(0x1000)
	b.loadARM(0x100,
		0xEB000001, // 100: BL 0x10C
		0xE3A00001, // 104: MOV r0, #1   (after return)
		0xEAFFFFFE, // 108: B .
		0xE3A02007, // 10C: MOV r2, #7   (subroutine)
		0xE12FFF1E, // 110: BX lr
	)
	c := NewCPU(b)
	c.Mode = ModeSYS
	c.R[15] = 0x100
	runTo(t, c, 0x108, 100)
	if c.R[2] != 7 {
		t.Fatalf("r2 = %d, want 7 (subroutine ran)", c.R[2])
	}
	if c.R[0] != 1 {
		t.Fatalf("r0 = %d, want 1 (returned to caller)", c.R[0])
	}
}

func TestThumb(t *testing.T) {
	b := newBus(0x1000)
	b.loadThumb(0x100,
		0x2005, // MOV r0, #5
		0x3003, // ADD r0, #3
		0xE7FE, // B . (spin)
	)
	c := NewCPU(b)
	c.Mode = ModeSYS
	c.Thumb = true
	c.R[15] = 0x100
	c.Step()
	c.Step()
	if c.R[0] != 8 {
		t.Fatalf("thumb r0 = %d, want 8", c.R[0])
	}
}

func TestInterwork(t *testing.T) {
	b := newBus(0x1000)
	b.loadARM(0x100,
		0xE3A00C02, // MOV r0, #0x200
		0xE2800001, // ADD r0, r0, #1   (0x201 → Thumb bit set)
		0xE12FFF10, // BX r0
	)
	b.loadThumb(0x200,
		0x2109, // MOV r1, #9
		0xE7FE, // B .
	)
	c := NewCPU(b)
	c.Mode = ModeSYS
	c.R[15] = 0x100
	c.Step() // MOV
	c.Step() // ADD
	c.Step() // BX r0 → Thumb @ 0x200
	if !c.Thumb {
		t.Fatalf("expected Thumb state after BX to odd address")
	}
	if c.R[15] != 0x200 {
		t.Fatalf("pc = 0x%X, want 0x200", c.R[15])
	}
	c.Step() // MOV r1,#9 in Thumb
	if c.R[1] != 9 {
		t.Fatalf("thumb r1 = %d, want 9", c.R[1])
	}
}
