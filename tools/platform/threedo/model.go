package threedo

// model.go decodes the "ORI3" 3D object format that holds Need for Speed's car
// models — the polygon meshes rendered in-race. They live inside the DriveArt
// `.WrapFam` files (an outer "wwww" wrap container) and standalone `.3OR` files.
// The format was reverse-engineered from the disc (clean-room): the Viper model
// decodes to 78 vertices and 58 quads whose vertices are perfectly left-right
// symmetric and whose silhouette is a Viper — the validation that pins it.
//
// ORI3 object layout (all big-endian), relative to the "ORI3" magic:
//
//	+0x00  "ORI3"
//	+0x04  u32 object size
//	+0x10  u32 vertex count
//	+0x14  u32 vertex offset   (vertices: count × int32 x,y,z = 12 bytes)
//	+0x20  u32 face count
//	+0x24  u32 face offset      (faces: count × 24-byte records)
//	+0x28  char[8] name         ("viper000", "SHADOW", …)
//
// Face record (24 bytes): u32 material, u32 id, then four u32 vertex indices —
// every face is a quad (the Madam cel engine textures quads). The material word
// indexes the "wrap" texture set carried by the enclosing WrapFam.

import "fmt"

// Vec3 is a model-space vertex (integer model units).
type Vec3 struct{ X, Y, Z int32 }

// Face is one textured quad: a material id and four vertex indices.
type Face struct {
	Material uint32
	V        [4]int
}

// Model is one decoded ORI3 object.
type Model struct {
	Name  string
	Verts []Vec3
	Faces []Face
}

// ParseModels decodes every ORI3 object in a WrapFam / .3OR file (a WrapFam may
// carry several — body plus parts).
func ParseModels(data []byte) ([]*Model, error) {
	var models []*Model
	for base := 0; base+0x30 <= len(data); {
		i := indexTagFrom(data, "ORI3", base)
		if i < 0 {
			break
		}
		m, next, err := parseORI3(data, i)
		if err != nil {
			// Skip this magic occurrence and keep scanning.
			base = i + 4
			continue
		}
		models = append(models, m)
		base = next
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("threedo: no ORI3 object found")
	}
	return models, nil
}

func parseORI3(data []byte, o int) (*Model, int, error) {
	rd := func(off int) uint32 { return be32(data[o+off:]) }
	if o+0x30 > len(data) {
		return nil, 0, fmt.Errorf("truncated ORI3 header")
	}
	nverts := int(rd(0x10))
	voff := int(rd(0x14))
	nfaces := int(rd(0x20))
	foff := int(rd(0x24))
	if nverts < 0 || nfaces < 0 || nverts > 100000 || nfaces > 100000 {
		return nil, 0, fmt.Errorf("implausible ORI3 counts %d/%d", nverts, nfaces)
	}
	if o+voff+nverts*12 > len(data) || o+foff+nfaces*24 > len(data) {
		return nil, 0, fmt.Errorf("ORI3 arrays overrun file")
	}
	m := &Model{Name: trimName(data[o+0x28 : o+0x30])}
	for i := 0; i < nverts; i++ {
		p := o + voff + i*12
		m.Verts = append(m.Verts, Vec3{
			X: int32(be32(data[p:])), Y: int32(be32(data[p+4:])), Z: int32(be32(data[p+8:])),
		})
	}
	for i := 0; i < nfaces; i++ {
		p := o + foff + i*24
		f := Face{Material: be32(data[p:])}
		ok := true
		for k := 0; k < 4; k++ {
			idx := int(be32(data[p+8+k*4:]))
			if idx >= nverts {
				ok = false
				break
			}
			f.V[k] = idx
		}
		if ok {
			m.Faces = append(m.Faces, f)
		}
	}
	end := o + int(rd(0x04))
	if end <= o {
		end = o + foff + nfaces*24
	}
	return m, end, nil
}

// indexTagFrom finds a 4-char tag at or after start, or -1.
func indexTagFrom(data []byte, tag string, start int) int {
	if start < 0 {
		start = 0
	}
	for i := start; i+4 <= len(data); i++ {
		if string(data[i:i+4]) == tag {
			return i
		}
	}
	return -1
}

// Bounds returns the min and max corner of the model's vertices.
func (m *Model) Bounds() (min, max Vec3) {
	if len(m.Verts) == 0 {
		return
	}
	min, max = m.Verts[0], m.Verts[0]
	for _, v := range m.Verts {
		if v.X < min.X {
			min.X = v.X
		}
		if v.Y < min.Y {
			min.Y = v.Y
		}
		if v.Z < min.Z {
			min.Z = v.Z
		}
		if v.X > max.X {
			max.X = v.X
		}
		if v.Y > max.Y {
			max.Y = v.Y
		}
		if v.Z > max.Z {
			max.Z = v.Z
		}
	}
	return
}
