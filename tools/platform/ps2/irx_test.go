package ps2

// irx_test.go checks the IOP module reader.
//
// As with ps2_test.go, every case here is a bug that actually happened while
// bringing the IOP up, and each one was quiet: a module that silently lost half its
// exports, and a relocation that moved a pointer by exactly 64 KiB.

import (
	"encoding/binary"
	"testing"
)

// imageOf assembles a module image out of words, for the table scanner to find.
func imageOf(words ...uint32) []byte {
	b := make([]byte, 4*len(words))
	for i, w := range words {
		binary.LittleEndian.PutUint32(b[4*i:], w)
	}
	return b
}

// libWords renders an 8-byte library name as two words, as it sits in a table header.
func libWords(name string) (uint32, uint32) {
	var b [8]byte
	copy(b[:], "        ")
	copy(b[:], name)
	return binary.LittleEndian.Uint32(b[0:]), binary.LittleEndian.Uint32(b[4:])
}

func TestExportTableHooksMayBeZero(t *testing.T) {
	// An export table's first four entries are the module's own hooks, and any of them
	// may be zero — SIFMAN's entry point *is* zero. Only from entry 4, where the
	// library's functions begin, does a zero word end the table.
	//
	// Stopping at the first zero of any kind loses the whole library: the census then
	// reports sifman, ioman and cdvdman as missing from a disc that carries all three,
	// and the conclusion drawn from it — "these must be written in Go" — is wrong in
	// the most expensive possible direction.
	n0, n1 := libWords("sifman")
	m := &IRX{Image: imageOf(
		irxExportMagic, 0, 0x0101, n0, n1,
		0x00000000, // [0] the entry point: zero, and not a terminator
		0x00000FC8, // [1]
		0x00000268, // [2]
		0x00000FC8, // [3]
		0x000000F4, // [4] the first real function
		0x00000148, // [5]
		0x00000000, // the terminator, at last
		0xDEADBEEF, // beyond it
	)}
	m.scanLibraries()

	if len(m.Exports) != 1 {
		t.Fatalf("found %d export tables, want 1", len(m.Exports))
	}
	e := m.Exports[0]
	if e.Library != "sifman" {
		t.Errorf("library is %q, want sifman", e.Library)
	}
	if len(e.Entries) != 6 {
		t.Fatalf("the table has %d entries, want 6 — a zero hook truncated it", len(e.Entries))
	}
	if e.Entries[4] != 0x000000F4 || e.Entries[5] != 0x00000148 {
		t.Errorf("functions 4 and 5 are 0x%X and 0x%X, want 0xF4 and 0x148", e.Entries[4], e.Entries[5])
	}
}

func TestImportTableStubsAreReadInOrder(t *testing.T) {
	// An import is resolved by index, so the *order* of the stubs is the whole of the
	// information: stub i belongs to function IDs[i], and the two must not drift apart.
	n0, n1 := libWords("thbase")
	m := &IRX{Image: imageOf(
		irxImportMagic, 0, 0x0101, n0, n1,
		irxStubJR, 0x24020004, // li $v0, 4
		irxStubJR, 0x24020006, // li $v0, 6
		irxStubJR, 0x24020014, // li $v0, 20
		0, 0, // the pair of zeros that ends the table
	)}
	m.scanLibraries()

	if len(m.Imports) != 1 {
		t.Fatalf("found %d import tables, want 1", len(m.Imports))
	}
	i := m.Imports[0]
	if i.Library != "thbase" {
		t.Errorf("library is %q, want thbase", i.Library)
	}
	want := []uint16{4, 6, 20}
	if len(i.IDs) != len(want) {
		t.Fatalf("read %v, want %v", i.IDs, want)
	}
	for k := range want {
		if i.IDs[k] != want[k] {
			t.Errorf("stub %d is function %d, want %d", k, i.IDs[k], want[k])
		}
		// The stub the loader will patch must be the one the ID came from.
		if got := m.word(i.Stubs[k]); got != irxStubJR {
			t.Errorf("stub %d does not point at a `jr $ra` (found 0x%08X)", k, got)
		}
	}
}

func TestHI16LO16CarriesIntoTheHighHalf(t *testing.T) {
	// The two halves of an address are split across a `lui`/`addiu` pair, and the low
	// half is *signed*. So a low half that will sign-extend to negative has to be paid
	// for with an extra 1 in the high half.
	//
	// The naive relocation — add base>>16 to the lui, leave the addiu alone — gets the
	// right answer often enough to boot, and then lands a pointer exactly 64 KiB from
	// where it belongs. Here the module refers to its own address 0x8000 and is loaded
	// at 0x001FF000, so the truth is 0x00207000; the naive answer is 0x001F8000.
	const (
		base = 0x001FF000
		want = 0x00207000
	)
	m := &IRX{
		Image: imageOf(
			0x3C080001, // lui   $t0, 0x0001    %hi(0x8000)
			0x25088000, // addiu $t0, $t0, -32768  %lo(0x8000)
		),
		MemSz: 8,
		Relocs: []IRXReloc{
			{Offset: 0, Type: rMIPSHI16},
			{Offset: 4, Type: rMIPSLO16},
		},
	}

	img, err := m.Relocate(base)
	if err != nil {
		t.Fatal(err)
	}
	hi := binary.LittleEndian.Uint32(img[0:]) & 0xFFFF
	lo := binary.LittleEndian.Uint32(img[4:]) & 0xFFFF

	got := uint32(hi<<16) + uint32(int32(int16(lo)))
	if got != want {
		t.Errorf("the pair reconstructs 0x%08X, want 0x%08X (lui 0x%04X, addiu 0x%04X)", got, want, hi, lo)
	}
}

func TestJumpRelocationIsAWordAddress(t *testing.T) {
	// A `jal`'s field is the target *divided by four*. Relocating it as if it were a
	// byte address quarters every call the module makes.
	m := &IRX{
		Image:  imageOf(0x0C000040), // jal 0x100
		MemSz:  4,
		Relocs: []IRXReloc{{Offset: 0, Type: rMIPS26}},
	}
	img, err := m.Relocate(0x00200000)
	if err != nil {
		t.Fatal(err)
	}
	insn := binary.LittleEndian.Uint32(img)
	if got, want := (insn&0x03FFFFFF)<<2, uint32(0x00200100); got != want {
		t.Errorf("the jump targets 0x%08X, want 0x%08X", got, want)
	}
}

func TestROMDIRWalksToEveryBody(t *testing.T) {
	// The archive's directory is the first thing in the file *and* an entry within it,
	// and every body is padded to 16 bytes. A walk that forgets the padding reads each
	// module a few bytes into the last one.
	dir := func(name string, size int) []byte {
		b := make([]byte, romEntrySize)
		copy(b, name)
		binary.LittleEndian.PutUint32(b[12:], uint32(size))
		return b
	}
	// Five records: the three named entries and the empty one that ends the directory.
	// The ROMDIR entry's size covers the whole of it, terminator included — which is
	// what makes the directory findable as a body like any other.
	const dirSize = 5 * romEntrySize

	var raw []byte
	raw = append(raw, dir("RESET", 0)...)
	raw = append(raw, dir("ROMDIR", dirSize)...)
	raw = append(raw, dir("FIRST", 17)...) // 17 bytes: pads to 32
	raw = append(raw, dir("SECOND", 3)...)
	raw = append(raw, make([]byte, romEntrySize)...) // the empty record that ends it

	// The bodies: the directory itself is the first, at offset 0.
	body := raw
	first := make([]byte, 32)
	for i := 0; i < 17; i++ {
		first[i] = 0xAA
	}
	second := []byte{0xBB, 0xBB, 0xBB}
	image := append(append(body, first...), second...)

	entries, err := ReadROMDIR(image)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("found %d entries, want 4", len(entries))
	}
	if got := entries[2]; got.Name != "FIRST" || len(got.Data) != 17 || got.Data[16] != 0xAA {
		t.Errorf("FIRST is %d bytes and ends 0x%02X, want 17 ending 0xAA", len(got.Data), got.Data[len(got.Data)-1])
	}
	if got := entries[3]; got.Name != "SECOND" || len(got.Data) != 3 || got.Data[0] != 0xBB {
		t.Errorf("SECOND is %v, want three 0xBB — the 16-byte padding was not honoured", got.Data)
	}
}
