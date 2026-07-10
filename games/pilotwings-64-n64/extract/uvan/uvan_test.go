package uvan

import (
	"math"
	"os"
	"sort"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
	"retroreverse.com/tools/platform/n64"
)

const romPath = "../../image/Pilotwings 64 (USA).z64"

func archive(t *testing.T) *pwad.Archive {
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
	return a
}

// decodeResource pulls a UVAN's COMM and PART chunks (PART may be GZIP-wrapped)
// in archive order and decodes them.
func decodeResource(t *testing.T, a *pwad.Archive, res int) *Animation {
	t.Helper()
	f, err := a.Resource(res)
	if err != nil {
		t.Fatal(err)
	}
	var comm []byte
	var parts [][]byte
	for _, c := range f.Chunks {
		tag := c.Tag
		if c.Compressed() {
			tag = c.InnerTag
		}
		data, err := a.Data(c)
		if err != nil {
			t.Fatal(err)
		}
		switch tag {
		case "COMM":
			comm = data
		case "PART":
			parts = append(parts, data)
		}
	}
	anim, err := Decode(comm, parts)
	if err != nil {
		t.Fatalf("UVAN %d: %v", res, err)
	}
	return anim
}

// Every UVAN in the archive decodes, and the decoder's own invariants hold:
// unit quaternions, frames within the header length, 8-byte-aligned tracks. The
// decoder returns an error on any breach, so this walks all 115 and counts.
func TestEveryAnimationDecodes(t *testing.T) {
	a := archive(t)
	idx := a.ByType("UVAN")
	if len(idx) != 115 {
		t.Fatalf("%d UVAN resources, want 115", len(idx))
	}
	var keys, worstNonUnit float64
	tracks := 0
	for _, res := range idx {
		anim := decodeResource(t, a, res)
		for _, tr := range anim.Tracks {
			tracks++
			if tr.Part < 1 {
				t.Errorf("UVAN %d: part index %d is not 1-based", res, tr.Part)
			}
			for _, k := range tr.Keys {
				keys++
				q := k.Rot
				n := math.Sqrt(float64(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W))
				if d := math.Abs(n - 1); d > worstNonUnit {
					worstNonUnit = d
				}
			}
		}
	}
	if worstNonUnit > 0.02 {
		t.Errorf("worst quaternion non-unit deviation %g, want <= 0.02", worstNonUnit)
	}
	t.Logf("115 animations, %d tracks, %.0f keyframes; worst |q|-1 = %g", tracks, keys, worstNonUnit)
}

// The load-bearing claim: a UVAN animates a UVMD model, named by ordinal in its
// header, and every animated part index is a real part of that model. The part
// index is 1-based and part 0 (the root) is never animated, so the valid range
// is [1, partCount).
//
// This is also what separates animation from scenery: the target ordinals must
// never fall in 0..198, the range a terrain chunk places as objects. If one did,
// an "animated" ferris wheel would be plausible — and wrong.
func TestAnimationsTargetModelParts(t *testing.T) {
	a := archive(t)
	uvmdIdx := a.ByType("UVMD")
	partCount := make([]int, len(uvmdIdx))
	for ord, r := range uvmdIdx {
		f, err := a.Resource(r)
		if err != nil {
			t.Fatal(err)
		}
		var data []byte
		for _, c := range f.Chunks {
			if c.Tag == "COMM" || (c.Compressed() && c.InnerTag == "COMM") {
				if data, err = a.Data(c); err != nil {
					t.Fatal(err)
				}
			}
		}
		m, err := uvmd.Decode(data)
		if err != nil {
			t.Fatalf("UVMD ordinal %d: %v", ord, err)
		}
		partCount[ord] = len(m.LODs[0].Parts)
	}

	// The object types a terrain chunk can place: 0..198 (package uvct).
	const maxObjectType = 198

	targets := map[int]int{}
	for _, res := range a.ByType("UVAN") {
		anim := decodeResource(t, a, res)
		if anim.Model < 0 || anim.Model >= len(partCount) {
			t.Fatalf("UVAN %d: target ordinal %d out of range", res, anim.Model)
		}
		targets[anim.Model]++
		if anim.Model <= maxObjectType {
			t.Errorf("UVAN %d targets ordinal %d, inside the terrain-object range", res, anim.Model)
		}
		for _, tr := range anim.Tracks {
			if tr.Part >= partCount[anim.Model] {
				t.Errorf("UVAN %d: part %d but model ordinal %d has %d parts",
					res, tr.Part, anim.Model, partCount[anim.Model])
			}
		}
	}
	ords := make([]int, 0, len(targets))
	for o := range targets {
		ords = append(ords, o)
	}
	sort.Ints(ords)
	t.Logf("animations target %d models, ordinals %d..%d, none in the object range 0..%d",
		len(ords), ords[0], ords[len(ords)-1], maxObjectType)
}
