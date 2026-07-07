package threedo

import (
	"encoding/binary"
	"testing"
)

// TestParseWrap builds a small nested wwww tree — an outer container with a cel
// leaf and a nested container holding a model leaf — and checks the walk resolves
// the base-relative offsets and classifies both leaves.
func TestParseWrap(t *testing.T) {
	be := binary.BigEndian
	b := make([]byte, 0x40)

	// outer @0: "wwww", count 2, child offsets (relative to 0): 0x10, 0x18.
	copy(b[0:], "wwww")
	be.PutUint32(b[4:], 2)
	be.PutUint32(b[8:], 0x10)  // -> leaf cel at 0x10
	be.PutUint32(b[12:], 0x18) // -> inner container at 0x18

	copy(b[0x10:], "CCB ") // a cel leaf

	// inner @0x18: "wwww", count 1, child offset relative to 0x18 = 0x0C -> 0x24.
	copy(b[0x18:], "wwww")
	be.PutUint32(b[0x1C:], 1)
	be.PutUint32(b[0x20:], 0x0C)
	copy(b[0x24:], "ORI3") // a model leaf

	res, err := ParseWrap(b)
	if err != nil {
		t.Fatalf("ParseWrap: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("got %d resources, want 2: %+v", len(res), res)
	}
	inv := Inventory(res)
	if inv["cel"] != 1 || inv["model"] != 1 {
		t.Errorf("inventory = %v, want 1 cel + 1 model", inv)
	}
	if res[0].Offset != 0x10 || res[0].Kind != "cel" {
		t.Errorf("res[0] = %+v, want cel@0x10", res[0])
	}
	if res[1].Offset != 0x24 || res[1].Kind != "model" || res[1].Depth != 2 {
		t.Errorf("res[1] = %+v, want model@0x24 depth 2", res[1])
	}
}

func TestParseWrapRejectsNonContainer(t *testing.T) {
	if _, err := ParseWrap([]byte("nope, not wwww")); err == nil {
		t.Errorf("ParseWrap(junk) succeeded, want error")
	}
}
