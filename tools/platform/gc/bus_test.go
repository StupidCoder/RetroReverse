package gc

// bus_test.go fuzzes the bus's word access against the byte-at-a-time assembly it replaced.
//
// The claim being defended is narrow and total: reading a word through binary.BigEndian
// produces the same value as shifting four bytes together, and writing one leaves the same
// four bytes in memory. That is obvious, and obvious is exactly the kind of change that
// ships a transposed index — so it is checked against the code it replaced rather than
// against a fresh statement of what the code was supposed to do, which would just be the
// same misunderstanding twice.
//
// The reference implementations below are verbatim copies of what machine.go held before.
// They are the specification here, and they must not be "tidied".

import (
	"math/rand"
	"testing"
)

// The byte-at-a-time assembly machine.go used to carry, kept as the reference.
func refRead32(ram []byte, a uint32) uint32 {
	return uint32(ram[a])<<24 | uint32(ram[a+1])<<16 | uint32(ram[a+2])<<8 | uint32(ram[a+3])
}

func refRead16(ram []byte, a uint32) uint16 {
	return uint16(ram[a])<<8 | uint16(ram[a+1])
}

func refWrite32(ram []byte, a, v uint32) {
	ram[a] = uint8(v >> 24)
	ram[a+1] = uint8(v >> 16)
	ram[a+2] = uint8(v >> 8)
	ram[a+3] = uint8(v)
}

func refWrite16(ram []byte, a uint32, v uint16) {
	ram[a] = uint8(v >> 8)
	ram[a+1] = uint8(v)
}

// TestBusEndianMatchesByteAssembly is the bit-exactness proof for the binary.BigEndian
// change. Reads are checked against the reference on random data; writes are checked by
// writing the same value both ways into two buffers and comparing the buffers whole — so a
// store that reached one byte too far would show as a difference outside the word as well as
// inside it.
func TestBusEndianMatchesByteAssembly(t *testing.T) {
	m := &Machine{RAM: make([]byte, RAMSize), logSeen: map[string]bool{}}
	ref := make([]byte, RAMSize)

	rng := rand.New(rand.NewSource(1))
	// A random-but-reproducible memory, so the reads have something to disagree about. The
	// low megabyte is enough to exercise every alignment.
	seed := make([]byte, 1<<20)
	rng.Read(seed)
	copy(m.RAM, seed)
	copy(ref, seed)

	// Every alignment, and both ends of the range the guards admit.
	addrs := []uint32{0, 1, 2, 3, 4, 5, 6, 7, 0x1000, 0x1001, 0x1002, 0x1003,
		RAMSize - 8, RAMSize - 7, RAMSize - 6, RAMSize - 5, RAMSize - 4}
	for i := 0; i < 4000; i++ {
		addrs = append(addrs, uint32(rng.Intn(1<<20-8)))
	}

	for _, a := range addrs {
		if got, want := m.Read32(a), refRead32(ref, a); got != want {
			t.Fatalf("Read32(0x%X) = 0x%08X, want 0x%08X", a, got, want)
		}
		if got, want := m.Read16(a), refRead16(ref, a); got != want {
			t.Fatalf("Read16(0x%X) = 0x%04X, want 0x%04X", a, got, want)
		}
		if got, want := m.Fetch32(a), refRead32(ref, a); got != want {
			t.Fatalf("Fetch32(0x%X) = 0x%08X, want 0x%08X", a, got, want)
		}
		if got, want := m.ram32(a), refRead32(ref, a); got != want {
			t.Fatalf("ram32(0x%X) = 0x%08X, want 0x%08X", a, got, want)
		}
	}

	for _, a := range addrs {
		// The reference has no guard — it is the arithmetic, copied verbatim — so the
		// write loop stays where the whole pattern fits. The out-of-range half is
		// TestBusGuardsAreUnchanged's job, and it is a different claim.
		if a+12 > RAMSize {
			continue
		}
		v := rng.Uint32()

		// EACH WRITER GETS ITS OWN ADDRESS. The first version of this loop pointed
		// Write32 and setRAM32 at the same word, so setRAM32 overwrote what Write32 had
		// just done and the comparison never saw it — a byte-order fault in Write32, the
		// hottest function in the machine, survived the fuzz test that was written to
		// catch it. Mutation testing found that; the addresses are spread because of it.
		m.Write32(a, v)
		refWrite32(ref, a, v)

		m.Write16(a+4, uint16(v))
		refWrite16(ref, a+4, uint16(v))

		m.setRAM32(a+8, ^v)
		refWrite32(ref, a+8, ^v)
	}
	// Compared whole, not word by word: a store that ran one byte long would land outside
	// every word this loop wrote and a per-word comparison would never look there.
	for i := range ref {
		if m.RAM[i] != ref[i] {
			t.Fatalf("RAM differs at 0x%X after the writes: got 0x%02X, want 0x%02X", i, m.RAM[i], ref[i])
		}
	}
}

// TestBusGuardsAreUnchanged pins the out-of-range behaviour, which is the half a fuzz over
// valid addresses cannot see. An unmapped access reads back zero and is logged; it must not
// panic and it must not start reading memory it previously declined to.
func TestBusGuardsAreUnchanged(t *testing.T) {
	m := &Machine{RAM: make([]byte, RAMSize), logSeen: map[string]bool{}}
	m.CPU = nil // these paths must not need a CPU to decline an address

	// Past the end of RAM and below the hardware block: unmapped.
	for _, a := range []uint32{RAMSize, RAMSize + 1, RAMSize + 4, 0x0A000000} {
		if got := m.ram32(a); got != 0 {
			t.Errorf("ram32(0x%X) = 0x%08X, want 0 (out of range)", a, got)
		}
		// A write out of range must be silently dropped rather than panic.
		m.setRAM32(a, 0xDEADBEEF)
	}

	// The last word that IS in range must still be readable and writable: the guard is
	// a+3 < RAMSize, so RAMSize-4 is the last address a word fits at.
	m.setRAM32(RAMSize-4, 0xCAFEF00D)
	if got := m.ram32(RAMSize - 4); got != 0xCAFEF00D {
		t.Errorf("ram32(RAMSize-4) = 0x%08X, want 0xCAFEF00D — the last in-range word", got)
	}
}
