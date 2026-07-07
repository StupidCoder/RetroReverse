package psx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// discPath is the Ridge Racer data track, relative to this package directory.
const discPath = "../../Ridge Racer (PSX)/Ridge Racer (Track 01).bin"

// loadDisc opens the development disc image, skipping the test if it is absent
// (the large image is not always present in every checkout).
func loadDisc(t *testing.T) *Volume {
	t.Helper()
	data, err := os.ReadFile(filepath.FromSlash(discPath))
	if err != nil {
		t.Skipf("disc image not available: %v", err)
	}
	v, err := Open(data)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return v
}

func TestOpenVolume(t *testing.T) {
	v := loadDisc(t)
	if v.System != "PLAYSTATION" {
		t.Errorf("System = %q, want PLAYSTATION", v.System)
	}
	if v.Name != "RIDGERACERUSA" {
		t.Errorf("Name = %q, want RIDGERACERUSA", v.Name)
	}
}

func TestListing(t *testing.T) {
	v := loadDisc(t)
	entries, err := v.ReadDir("")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	want := map[string]bool{"SYSTEM.CNF;1": false, "SCUS-943.00;1": false}
	for _, e := range entries {
		if _, ok := want[e.Name]; ok {
			want[e.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("root listing missing %q", name)
		}
	}
}

func TestReadFile(t *testing.T) {
	v := loadDisc(t)

	// A small file that must round-trip exactly and by name without a version.
	cnf, err := v.ReadFile("SYSTEM.CNF")
	if err != nil {
		t.Fatalf("ReadFile SYSTEM.CNF: %v", err)
	}
	if !strings.HasPrefix(string(cnf), "BOOT") {
		t.Errorf("SYSTEM.CNF does not start with BOOT: %q", cnf[:min(16, len(cnf))])
	}
	if !strings.Contains(string(cnf), "SCUS-943.00") {
		t.Errorf("SYSTEM.CNF missing boot name: %q", cnf)
	}

	// The boot executable's size must match the directory entry.
	exe, err := v.ReadFile("SCUS-943.00;1")
	if err != nil {
		t.Fatalf("ReadFile SCUS-943.00: %v", err)
	}
	if len(exe) != 438272 {
		t.Errorf("SCUS-943.00 size = %d, want 438272", len(exe))
	}
	if string(exe[0:8]) != "PS-X EXE" {
		t.Errorf("boot file is not a PS-X EXE: %q", exe[0:8])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
