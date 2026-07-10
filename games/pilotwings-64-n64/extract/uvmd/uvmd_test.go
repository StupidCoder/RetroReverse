package uvmd

import (
	"os"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/tools/platform/n64"
)

const romPath = "../../image/Pilotwings 64 (USA).z64"

func all(t *testing.T) []*Model {
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
	var out []*Model
	for _, i := range a.ByType("UVMD") {
		f, err := a.Resource(i)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range f.Chunks {
			tag := c.Tag
			if c.Compressed() {
				tag = c.InnerTag
			}
			if tag != "COMM" {
				continue
			}
			data, err := a.Data(c)
			if err != nil {
				t.Fatal(err)
			}
			m, err := Decode(data)
			if err != nil {
				t.Fatalf("UVMD %d: %v", i, err)
			}
			out = append(out, m)
		}
	}
	return out
}

// Every model parses, and the parse consumes the whole resource but for the
// alignment padding. A field-order error would leave far more than 8 bytes.
func TestEveryModelDecodes(t *testing.T) {
	ms := all(t)
	if got, want := len(ms), 363; got != want {
		t.Fatalf("decoded %d models, want %d", got, want)
	}
	tris := 0
	for i, m := range ms {
		if m.Padding >= 8 {
			t.Errorf("model %d: %d bytes unconsumed", i, m.Padding)
		}
		if len(m.LODs) == 0 {
			t.Errorf("model %d: no LODs", i)
			continue
		}
		tris += m.Triangles(0)
	}
	if want := 60949; tris != want {
		t.Errorf("LOD 0 has %d triangles across the archive, want %d", tris, want)
	}
}

// The material's low 12 bits index the UVTX resources, which are contiguous in
// the archive. Nothing may point past them.
func TestMaterialTextureOrdinalsAreInRange(t *testing.T) {
	ms := all(t)
	const uvtxCount = 463
	untextured := 0
	for i, m := range ms {
		for _, lod := range m.LODs {
			for _, p := range lod.Parts {
				for _, b := range p.Batches {
					tex, ok := b.Material.Texture()
					if !ok {
						untextured++
						continue
					}
					if tex >= uvtxCount {
						t.Fatalf("model %d: material %#x references UVTX ordinal %d of %d",
							i, uint32(b.Material), tex, uvtxCount)
					}
				}
			}
		}
	}
	if untextured == 0 {
		t.Error("no untextured batch found; the PILOTWINGS letters alone are 1,464 such triangles")
	}
}

// LOD 0 always has exactly one rest-pose matrix per part, which is what
// licenses pairing matrix i with part i when assembling a model. Lower LODs may
// drop a part entirely — model 83's LOD 1 has two parts against three matrices —
// so the pairing is asserted only where it is sound.
func TestLOD0HasOneMatrixPerPart(t *testing.T) {
	ms := all(t)
	droppedParts := 0
	for i, m := range ms {
		if got, want := len(m.LODs[0].Parts), len(m.Matrices); got != want {
			t.Errorf("model %d lod 0: %d parts but %d matrices", i, got, want)
		}
		for _, lod := range m.LODs[1:] {
			if len(lod.Parts) != len(m.Matrices) {
				droppedParts++
			}
		}
	}
	t.Logf("%d lower LODs drop a part relative to the matrix count", droppedParts)
}

// A model with a single part has nothing to place, so its rest pose is the
// identity. That the file says so — rather than carrying arbitrary numbers —
// is the evidence that these 64-byte blocks really are per-part poses in part
// order, which is otherwise only inferred from the loader's allocation order.
func TestSinglePartModelsHaveAnIdentityRestPose(t *testing.T) {
	ms := all(t)
	single, identity := 0, 0
	for _, m := range ms {
		if len(m.Matrices) != 1 {
			continue
		}
		single++
		if isIdentity(m.Matrices[0]) {
			identity++
		}
	}
	if single == 0 {
		t.Fatal("no single-part model found")
	}
	if identity != single {
		t.Errorf("%d of %d single-part models carry a non-identity rest pose", single-identity, single)
	}
	t.Logf("%d single-part models, all identity", single)
}

func isIdentity(m Matrix) bool {
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			want := float32(0)
			if r == c {
				want = 1
			}
			if m[r][c] != want {
				return false
			}
		}
	}
	return true
}
