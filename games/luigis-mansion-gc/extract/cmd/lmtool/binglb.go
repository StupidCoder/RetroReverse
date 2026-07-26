package main

// binglb.go exports a .bin model (a mansion room or furniture piece) as a
// GLB: walk the scene graph composing each node's TRS, draw its
// (material, mesh) pairs, group triangles by material.

import (
	"fmt"
	"math"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/lib/glb"
)

// binNodeMatrix builds a node's local matrix the way the renderer at
// 0x8001D96C does: rotation Rz·Ry·Rx from degrees, columns scaled, the
// translation in the fourth column.
func binNodeMatrix(n *lm.BinNode) lm.Mtx34 {
	rad := func(d float32) float64 { return float64(d) * math.Pi / 180 }
	sx, cx := math.Sincos(rad(n.Rot[0]))
	sy, cy := math.Sincos(rad(n.Rot[1]))
	sz, cz := math.Sincos(rad(n.Rot[2]))
	r := lm.Mtx34{
		float32(cy * cz), float32(cz*sy*sx - sz*cx), float32(cz*sy*cx + sz*sx), n.Trans[0],
		float32(sz * cy), float32(sz*sy*sx + cz*cx), float32(sz*sy*cx - cz*sx), n.Trans[1],
		float32(-sy), float32(cy * sx), float32(cy * cx), n.Trans[2],
	}
	for row := 0; row < 3; row++ {
		r[row*4] *= n.Scale[0]
		r[row*4+1] *= n.Scale[1]
		r[row*4+2] *= n.Scale[2]
	}
	return r
}

// binGLB writes the model, baked into world space.
func binGLB(m *lm.Bin, path, name string) error {
	world := make([]lm.Mtx34, len(m.Nodes))
	for i := range world {
		world[i] = lm.Identity34
	}
	var walk func(i int, parent lm.Mtx34)
	walk = func(i int, parent lm.Mtx34) {
		for i >= 0 && i < len(m.Nodes) {
			world[i] = parent.Mul(binNodeMatrix(&m.Nodes[i]))
			if c := m.Nodes[i].Child; c >= 0 {
				walk(c, world[i])
			}
			i = m.Nodes[i].Next
		}
	}
	if len(m.Nodes) > 0 {
		walk(0, lm.Identity34)
	}

	prims := map[int]*glb.Prim{}
	var order []int
	for ni := range m.Nodes {
		for _, pair := range m.Nodes[ni].Pairs {
			if pair.Mesh >= len(m.Meshes) || pair.Material >= len(m.Materials) {
				continue
			}
			mesh := &m.Meshes[pair.Mesh]
			pr, ok := prims[pair.Material]
			if !ok {
				pr = &glb.Prim{BaseColor: [4]float32{1, 1, 1, 1}, DoubleSided: true, Unlit: true}
				mat := &m.Materials[pair.Material]
				if mat.Sampler >= 0 && mat.Sampler < len(m.Samplers) {
					s := m.Samplers[mat.Sampler]
					if s.Texture >= 0 && s.Texture < len(m.Textures) {
						t := m.Textures[s.Texture]
						pr.Image = texImage(lm.MDLTexture{Width: t.Width, Height: t.Height, Pixels: t.Pixels})
						pr.WrapS, pr.WrapT = gxWrap(uint8(s.WrapS)), gxWrap(uint8(s.WrapT))
					}
				} else {
					pr.BaseColor = [4]float32{
						float32(mat.Tint[0]) / 255, float32(mat.Tint[1]) / 255,
						float32(mat.Tint[2]) / 255, float32(mat.Tint[3]) / 255,
					}
				}
				prims[pair.Material] = pr
				order = append(order, pair.Material)
			}
			dl, err := m.ParseBinDL(mesh)
			if err != nil {
				return fmt.Errorf("%s mesh %d: %w", name, pair.Mesh, err)
			}
			for _, p := range dl {
				base := len(pr.Positions)
				for _, v := range p.Verts {
					if int(v.Pos) >= len(m.Positions) {
						return fmt.Errorf("%s mesh %d: position index out of range", name, pair.Mesh)
					}
					pr.Positions = append(pr.Positions, world[ni].Apply(m.Positions[v.Pos]))
					if v.HasNrm && int(v.Nrm) < len(m.Normals) {
						pr.Normals = append(pr.Normals, world[ni].ApplyVec(m.Normals[v.Nrm]))
					}
					if v.HasTex && int(v.Tex) < len(m.Texcoords) {
						pr.UVs = append(pr.UVs, m.Texcoords[v.Tex])
					}
				}
				for _, t := range p.Triangulate() {
					pr.Tris = append(pr.Tris, [3]uint32{uint32(base + t[0]), uint32(base + t[1]), uint32(base + t[2])})
				}
			}
		}
	}

	s := glb.NewScene()
	root := s.AddNode(name, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
	var list []glb.Prim
	for _, mi := range order {
		p := prims[mi]
		if len(p.Positions) == 0 || len(p.Tris) == 0 {
			continue
		}
		if len(p.Normals) != len(p.Positions) {
			p.Normals = nil
		}
		if len(p.UVs) != len(p.Positions) {
			p.UVs = nil
		}
		list = append(list, *p)
	}
	if err := s.AddMesh(root, name, list); err != nil {
		return err
	}
	return s.Write(path, name)
}

// binExport pulls member.bin out of "/disc/file.arc:member.bin" and writes it.
func binExport(image, spec, out string) error {
	path, member, ok := strings.Cut(spec, ":")
	if !ok {
		return fmt.Errorf("want /disc/file.arc:member.bin, got %q", spec)
	}
	b, err := discFile(image, path)
	if err != nil {
		return err
	}
	if len(b) >= 4 && string(b[:4]) == "Yay0" {
		if b, err = lm.Yay0(b); err != nil {
			return err
		}
	}
	files, err := lm.RARC(b)
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.EqualFold(f.Name, member) {
			m, err := lm.ParseBin(f.Data)
			if err != nil {
				return err
			}
			return binGLB(m, out, strings.TrimSuffix(member, ".bin"))
		}
	}
	return fmt.Errorf("no member %q in %s", member, path)
}
