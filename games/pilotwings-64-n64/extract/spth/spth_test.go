package spth

import (
	"os"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/tools/platform/n64"
)

const romPath = "../../image/Pilotwings 64 (USA).z64"

func paths(t *testing.T) []*Path {
	t.Helper()
	if _, err := os.Stat(romPath); err != nil {
		t.Skip("cartridge image not present")
	}
	rom, err := n64.Load(romPath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		t.Fatal(err)
	}
	var out []*Path
	for _, i := range a.ByType("SPTH") {
		f, err := a.Resource(i)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string][]byte{}
		for _, c := range f.Chunks {
			data, err := a.Data(c)
			if err != nil {
				t.Fatal(err)
			}
			m[c.Tag] = data
		}
		p, err := Decode(m)
		if err != nil {
			t.Fatalf("SPTH %d: %v", i, err)
		}
		out = append(out, p)
	}
	return out
}

// Every curve's key count satisfies 8 + 8*count == size — DecodeCurve refuses
// otherwise, so reaching eight paths is the assertion.
func TestEightPathsDecode(t *testing.T) {
	if got, want := len(paths(t)), 8; got != want {
		t.Fatalf("%d spline paths, want %d", got, want)
	}
}

// The two observations that name Key.Value and Key.Time: in every multi-key
// curve the time rises across all keys but the last, whose time is exactly 0.0.
func TestEveryCurveIsTerminated(t *testing.T) {
	for i, p := range paths(t) {
		if !p.Terminated() {
			t.Errorf("path %d: a curve's time does not rise to a 0.0 terminator", i)
		}
	}
}

// Six of the eight paths key X, Y and Z on one timeline; the other two give
// them different key counts outright. Pinning the split guards the reading.
func TestSixPathsShareOneTimeline(t *testing.T) {
	shared := 0
	for _, p := range paths(t) {
		if p.SharedTimeline() {
			shared++
		}
	}
	if shared != 6 {
		t.Errorf("%d paths share one position timeline, want 6", shared)
	}
}
