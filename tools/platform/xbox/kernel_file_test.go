package xbox

import "testing"

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
