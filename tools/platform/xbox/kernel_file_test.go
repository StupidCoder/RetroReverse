package xbox

import (
	"strings"
	"testing"
)

// TestMatchPattern covers the NT search-wildcard subset NtQueryDirectoryFile uses. The
// only pattern this title passes is "*", but the matcher backtracks, and a backtracking
// matcher that is never tested is a guess.
func TestMatchPattern(t *testing.T) {
	cases := []struct {
		pat, name string
		want      bool
	}{
		{"*", "SAVE.DAT", true},
		{"*", "", true},
		{"*.*", "SAVE.DAT", true},
		{"*.*", "SAVE", false},
		{"*.DAT", "SAVE.DAT", true},
		{"*.DAT", "SAVE.BIN", false},
		{"save.dat", "SAVE.DAT", true}, // case-insensitive, like the FS
		{"SAVE.???", "SAVE.DAT", true},
		{"SAVE.???", "SAVE.BI", false},
		{"S*E.DAT", "SAVE.DAT", true},
		{"S*E.DAT", "SAVED.DAT", false},
		{"*A*A*", "ABABA", true},
		{"*X*", "ABABA", false},
		{"", "X", false},
		{"", "", true},
	}
	for _, c := range cases {
		if got := matchPattern(c.pat, c.name); got != c.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", c.pat, c.name, got, c.want)
		}
	}
}

// TestSetFileLength covers NtSetInformationFile's two 8-byte length classes, which is the
// action behind the game's save path committing common.dat at the position it had just
// written to (kernel_objects.go, ordinal 226).
//
// The live call is a NO-OP — it sets 67708 on a file already 67708 bytes long — so the run
// that found this frontier cannot tell a correct implementation from one that does nothing
// at all. That is exactly why the truncate and grow cases are asserted here instead.
func TestSetFileLength(t *testing.T) {
	m := &Machine{cacheFS: map[string]*cacheFile{}}
	cf := &cacheFile{Data: []byte("0123456789")}
	fo := &fileObject{cache: cf, key: "U:/SAVE/COMMON.DAT"}

	// The live case: setting the length a file already has changes nothing.
	if st := m.setFileLength(fo, 10); st != 0 {
		t.Fatalf("set 10 on a 10-byte file = %08X, want success", st)
	}
	if string(cf.Data) != "0123456789" {
		t.Errorf("a no-op set changed the file to %q", cf.Data)
	}

	// Truncation.
	if st := m.setFileLength(fo, 4); st != 0 {
		t.Fatalf("truncate = %08X, want success", st)
	}
	if string(cf.Data) != "0123" {
		t.Errorf("truncated to %q, want %q", cf.Data, "0123")
	}

	// Growth ZERO-FILLS. This is the assertion with something behind it: the truncation
	// above left 6 bytes of "456789" in the slice's spare capacity, and `append` does not
	// clear what it reuses — so an implementation that grew by appending without zeroing
	// would hand the guest its own deleted bytes back, and every length would still be
	// right. A file's hole reads as zeroes.
	if st := m.setFileLength(fo, 8); st != 0 {
		t.Fatalf("grow = %08X, want success", st)
	}
	if want := "0123\x00\x00\x00\x00"; string(cf.Data) != want {
		t.Errorf("grew to %q, want %q — a hole reads as zeroes, not as the bytes the "+
			"truncation left in the slice's spare capacity", cf.Data, want)
	}

	// The POSITION is the caller's business and this call must not move it — including
	// when it is left past the new end, which is precisely where the game leaves it.
	fo.off = 8
	if st := m.setFileLength(fo, 2); st != 0 {
		t.Fatalf("truncate under the position = %08X, want success", st)
	}
	if fo.off != 8 {
		t.Errorf("the set moved the file position to %d; it is the caller's", fo.off)
	}
}

// TestRawDevicePath pins which device this HLE serves, and — more importantly — which it
// refuses. Serving every \Device\Harddisk0\partitionN killed the boot after 1,465
// instructions on an unimplemented ordinal, because partition1 diverts the XAPI's own
// mount. The narrowness IS the model, so a widening has to fail here first.
func TestRawDevicePath(t *testing.T) {
	for _, p := range []string{
		`\Device\Harddisk0\partition0`,
		`\Device\Harddisk0\partition0\`,
		`\DEVICE\HARDDISK0\PARTITION0`, // the object manager is case-insensitive
		`/Device/Harddisk0/partition0`,
	} {
		if k, ok := rawDevicePath(p); !ok || k != rawPartition0Key {
			t.Errorf("rawDevicePath(%q) = %q,%v — want the raw device", p, k, ok)
		}
	}
	for _, p := range []string{
		// partition1 above all: serving it is what killed the boot.
		`\Device\Harddisk0\partition1`,
		`\Device\Harddisk0\partition1\`,
		`\Device\Harddisk0\partition2`,
		`\Device\Harddisk0\partition10`, // must not match by prefix
		`\Device\CdRom0\default.xbe`,
		`\Device\Harddisk0`,
		`Z:\MENU.PAK`,
		`U:\`,
	} {
		if k, ok := rawDevicePath(p); ok {
			t.Errorf("rawDevicePath(%q) = %q — this HLE serves partition0 ONLY; serving "+
				"partition1 diverts the XAPI's mount and kills the boot", p, k)
		}
	}
}

// TestRawDeviceIsNotChargedToTheSavePartition: the raw device shares the store so it rides
// the savestate, but it is not a file on a title partition. Charging its bytes to one is the
// shape of the bug Part VII already paid for — the title telling the player their console is
// full over a number the HLE invented.
func TestRawDeviceIsNotChargedToTheSavePartition(t *testing.T) {
	m := &Machine{cacheFS: map[string]*cacheFile{}}
	_, availEmpty := m.volumeUnits(&fileObject{dir: true, key: "U:/"})

	m.rawDevice(rawPartition0Key)
	total, avail := m.volumeUnits(&fileObject{dir: true, key: "U:/"})
	if avail != availEmpty {
		t.Errorf("the raw device cost the save partition %d units of free space",
			availEmpty-avail)
	}
	if total != hddTotalUnits {
		t.Errorf("volume total = %d, want %d", total, hddTotalUnits)
	}

	// ...but a real save file still is charged, or the check above would pass on a
	// volumeUnits that had stopped counting anything at all.
	m.cacheFS["U:/SAVE/COMMON.DAT"] = &cacheFile{Data: make([]byte, hddBytesPerUnit*3)}
	if _, avail3 := m.volumeUnits(&fileObject{dir: true, key: "U:/"}); avail3 != availEmpty-3 {
		t.Errorf("a 3-unit save file moved free space by %d units, want 3", availEmpty-avail3)
	}
}

// TestRawDeviceIsBlankAndDoesNotEnumerate covers the two things the invention promises: the
// device reads as zeros (the "no XONLINE account" claim), and it never appears as a file in
// a directory listing of the writable store, whose keys it shares.
func TestRawDeviceIsBlankAndDoesNotEnumerate(t *testing.T) {
	m := &Machine{cacheFS: map[string]*cacheFile{}}
	// Through the mint, NOT a buffer this test zeroed itself — otherwise the assertion
	// checks the invention against itself and a non-blank device would sail past.
	dev := m.rawDevice(rawPartition0Key)
	m.cacheFS["U:/SAVE.DAT"] = &cacheFile{Data: []byte("x")}

	if len(dev.Data) != rawPartition0Size {
		t.Fatalf("the raw device is %d bytes, want %d", len(dev.Data), rawPartition0Size)
	}
	for _, b := range dev.Data {
		if b != 0 {
			t.Fatalf("the raw device is not blank; it claims an XONLINE account exists")
		}
	}
	// The signature Init and the enumerate test (0x21843B / 0x223AB0: CMP [EBX+$1C],
	// $56525347). Blank must NOT match it — that is the whole of the "no account" claim.
	sig := uint32(0)
	for i := 0; i < 4; i++ {
		sig |= uint32(dev.Data[0x1C+i]) << (8 * i)
	}
	if sig == 0x56525347 {
		t.Error("the blank device matches the 'GSRV' signature it is supposed to fail")
	}

	for _, e := range m.listDir(&fileObject{dir: true, key: "U:"}) {
		if strings.Contains(e.Name, "RAW") || strings.Contains(e.Name, "Device") {
			t.Errorf("the raw device leaked into a directory listing as %q", e.Name)
		}
	}
}
