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
