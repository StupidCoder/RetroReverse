package threedo

import (
	"encoding/binary"
	"testing"
)

// buildORI3 assembles a minimal ORI3 object: 4 vertices and one quad, exercising
// the header offsets, the 12-byte vertex triples and the 24-byte face records.
func buildORI3(name string, verts []Vec3, quad [4]int, material uint32) []byte {
	be := binary.BigEndian
	const voff, hdr = 0x30, 0x30
	foff := voff + len(verts)*12
	size := foff + 24
	b := make([]byte, size)
	copy(b, "ORI3")
	be.PutUint32(b[0x04:], uint32(size))
	be.PutUint32(b[0x10:], uint32(len(verts)))
	be.PutUint32(b[0x14:], voff)
	be.PutUint32(b[0x20:], 1) // one face
	be.PutUint32(b[0x24:], uint32(foff))
	copy(b[0x28:0x30], name)
	for i, v := range verts {
		p := voff + i*12
		be.PutUint32(b[p:], uint32(v.X))
		be.PutUint32(b[p+4:], uint32(v.Y))
		be.PutUint32(b[p+8:], uint32(v.Z))
	}
	be.PutUint32(b[foff:], material)
	be.PutUint32(b[foff+4:], 0) // id
	for k := 0; k < 4; k++ {
		be.PutUint32(b[foff+8+k*4:], uint32(quad[k]))
	}
	return b
}

func TestParseORI3(t *testing.T) {
	verts := []Vec3{{-10, 0, -20}, {10, 0, -20}, {10, 0, 20}, {-10, 0, 20}}
	// Wrap it in some leading bytes to prove the "ORI3" scan works inside a container.
	data := append([]byte("wwww\x00\x00\x00\x01padding!"), buildORI3("TESTCAR0", verts, [4]int{0, 1, 2, 3}, 0xDF0012)...)

	models, err := ParseModels(data)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("got %d models, want 1", len(models))
	}
	m := models[0]
	if m.Name != "TESTCAR0" {
		t.Errorf("name = %q, want TESTCAR0", m.Name)
	}
	if len(m.Verts) != 4 || len(m.Faces) != 1 {
		t.Fatalf("got %d verts %d faces, want 4/1", len(m.Verts), len(m.Faces))
	}
	if m.Verts[1] != (Vec3{10, 0, -20}) {
		t.Errorf("vert[1] = %+v", m.Verts[1])
	}
	if m.Faces[0].V != [4]int{0, 1, 2, 3} || m.Faces[0].Material != 0xDF0012 {
		t.Errorf("face = %+v", m.Faces[0])
	}
	mn, mx := m.Bounds()
	if mn.X != -10 || mx.Z != 20 {
		t.Errorf("bounds = %+v..%+v", mn, mx)
	}
}

func TestParseModelsRejectsNonModel(t *testing.T) {
	if _, err := ParseModels([]byte("no object here, just text padding padding")); err == nil {
		t.Errorf("ParseModels(junk) succeeded, want error")
	}
}
