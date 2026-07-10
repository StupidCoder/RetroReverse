package pwad

import (
	"os"
	"testing"

	"retroreverse.com/tools/platform/n64"
)

// The cartridge is not committed (copyright). These tests skip without it.
const romPath = "../../image/Pilotwings 64 (USA).z64"

func load(t *testing.T) *Archive {
	t.Helper()
	if _, err := os.Stat(romPath); err != nil {
		t.Skip("cartridge image not present")
	}
	rom, err := n64.Load(romPath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := Open(rom.Data)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// The archive's shape, as measured. These are assertions about this cartridge,
// so they are exact: a change means the reader drifted, not that the game did.
func TestArchiveShape(t *testing.T) {
	a := load(t)
	if got, want := len(a.Forms), 1273; got != want {
		t.Errorf("FORMs = %d, want %d", got, want)
	}
	if got, want := len(a.Dir), 1272; got != want {
		t.Errorf("directory entries = %d, want %d", got, want)
	}
	if got, want := a.End, uint32(0x0618B6C); got != want {
		t.Errorf("archive end = 0x%07X, want 0x%07X", got, want)
	}
	if got, want := a.Forms[0].Type, "UVRM"; got != want {
		t.Errorf("directory FORM type = %s, want %s", got, want)
	}
}

// The FORM walk and the TABL directory are independent readings of the same
// container. Check demands they agree on every resource's type, length and
// offset, and that the directory's running sum lands on the archive's end.
func TestDirectoryAgreesWithTheWalk(t *testing.T) {
	a := load(t)
	if err := a.Check(); err != nil {
		t.Fatal(err)
	}
}

// Every MIO0 stream in the archive must inflate to exactly the length its chunk
// header declares — 1,322 independent checks of the codec in extract/mio0.
func TestEveryCompressedChunkInflatesToItsDeclaredSize(t *testing.T) {
	a := load(t)
	n := 0
	for _, f := range a.Forms {
		for _, c := range f.Chunks {
			if !c.Compressed() {
				continue
			}
			if _, err := a.Data(c); err != nil { // Data asserts the size
				t.Fatalf("resource %d (%s): %v", f.Index, f.Type, err)
			}
			n++
		}
	}
	if want := 1322; n != want {
		t.Errorf("compressed chunks = %d, want %d", n, want)
	}
}

// The type census. Written out so that a misparse — which would redistribute
// resources between types — fails loudly rather than producing a plausible list.
func TestTypeCensus(t *testing.T) {
	a := load(t)
	want := map[string]int{
		"UVTX": 463, "UVMD": 363, "UVAN": 115, "UVBT": 102, "UVCT": 101,
		"UPWT": 61, "PDAT": 25, "3VUE": 12, "UVFT": 9, "SPTH": 8, "UPWL": 4,
		"UVSY": 1, "UVEN": 1, "UVLT": 1, "UVTR": 1, "UVLV": 1, "UVSQ": 1,
		"UVTP": 1, "ADAT": 1, "UVSX": 1,
	}
	got := map[string]int{}
	for _, f := range a.Forms[1:] {
		got[f.Type]++
	}
	for typ, n := range want {
		if got[typ] != n {
			t.Errorf("%s: %d resources, want %d", typ, got[typ], n)
		}
	}
	for typ, n := range got {
		if _, ok := want[typ]; !ok {
			t.Errorf("unexpected FORM type %s (%d resources)", typ, n)
		}
	}
}

// ByType indexes into the directory, so a returned index must resolve to a FORM
// of that type.
func TestByTypeResolves(t *testing.T) {
	a := load(t)
	idx := a.ByType("UVMD")
	if len(idx) != 363 {
		t.Fatalf("UVMD count = %d, want 363", len(idx))
	}
	for _, i := range idx[:8] {
		f, err := a.Resource(i)
		if err != nil {
			t.Fatal(err)
		}
		if f.Type != "UVMD" {
			t.Fatalf("resource %d is %s, want UVMD", i, f.Type)
		}
	}
}
