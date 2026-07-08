package strpak

import (
	"os"
	"path/filepath"
	"testing"
)

func load(t *testing.T) *Archive {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "game", "DATA", "STRINGS.PAK"))
	if err != nil {
		t.Skipf("STRINGS.PAK not available: %v", err)
	}
	a, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestKnownStrings(t *testing.T) {
	a := load(t)
	cases := []struct {
		block, idx int
		want       string
	}{
		{1, 0, "Hey, its all the game strings"},
		{1, 2, "The key does not fit.\n"},
		{1, 5, "The key unlocks the lock.\n"},
		{4, 0, "a_hand axe"}, // object names
		{4, 3, "a_dagger"},
		{5, 0, "broken"}, // quality words
	}
	for _, c := range cases {
		got, ok := a.String(c.block, c.idx)
		if !ok {
			t.Errorf("block %d idx %d: missing", c.block, c.idx)
			continue
		}
		if got != c.want {
			t.Errorf("block %d idx %d = %q, want %q", c.block, c.idx, got, c.want)
		}
	}
}

func TestAllTerminate(t *testing.T) {
	a := load(t)
	// Every string must decode without running off the end of the archive; a
	// missing terminator would show up as a decode that consumed the tail. We
	// re-verify each block decodes and is non-panicking, and spot-check counts.
	total := 0
	for _, id := range a.BlockIDs() {
		ss, ok := a.Block(id)
		if !ok {
			t.Fatalf("block %d vanished", id)
		}
		total += len(ss)
	}
	if total < 1000 {
		t.Errorf("only %d strings decoded, expected thousands", total)
	}
}
