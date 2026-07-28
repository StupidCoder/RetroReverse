package export

// glb.go turns a parsed .mdl into a GLB. The static path (the forest set)
// bakes every vertex through its node's composed world matrix, groups
// triangles by material, and hands the result to the shared GLB scene writer.

import (
	"fmt"
	"image"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/lib/glb"
)

// texImage wraps a decoded MDL texture as an image.Image for the GLB writer.
func texImage(t lm.MDLTexture) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, t.Width, t.Height))
	copy(img.Pix, t.Pixels)
	return img
}

func gxWrap(w uint8) int {
	switch w {
	case 0:
		return 33071 // CLAMP_TO_EDGE
	case 2:
		return 33648 // MIRRORED_REPEAT
	default:
		return 10497 // REPEAT
	}
}

// worldMatrices returns each node's bind-pose world matrix. The file stores
// inverse binds (world -> joint, verified against the game's runtime array:
// runtime world[i] == inverse(file[i]) for the whole forest set), so the world
// matrix is simply each entry's inverse — absolute, no tree composition.
func worldMatrices(m *lm.MDL) []lm.Mtx34 {
	out := make([]lm.Mtx34, len(m.Nodes))
	for i := range m.Nodes {
		out[i] = invert34(m.Nodes[i].Mtx)
	}
	return out
}

// StaticGLB writes a fully baked world-space GLB of an .mdl.
func StaticGLB(m *lm.MDL, path, name string) error {
	world := worldMatrices(m)

	type key struct{ mat int }
	prims := map[int]*glb.Prim{}
	order := []int{}

	for ni := range m.Nodes {
		node := &m.Nodes[ni]
		for _, pair := range node.Pairs {
			if pair.Shape >= len(m.Shapes) || pair.Material >= len(m.Materials) {
				continue
			}
			shape := &m.Shapes[pair.Shape]
			nbt := shape.Flags&0x02000000 != 0
			pr, ok := prims[pair.Material]
			if !ok {
				pr = &glb.Prim{}
				mat := &m.Materials[pair.Material]
				if len(mat.Samplers) > 0 {
					s := m.Samplers[mat.Samplers[0]]
					if s.Texture < len(m.Textures) {
						pr.Image = texImage(m.Textures[s.Texture])
						pr.WrapS, pr.WrapT = gxWrap(s.WrapS), gxWrap(s.WrapT)
					}
				}
				pr.BaseColor = [4]float32{
					float32(mat.Tint[0]) / 255, float32(mat.Tint[1]) / 255,
					float32(mat.Tint[2]) / 255, float32(mat.Tint[3]) / 255,
				}
				pr.Blend = mat.Mode == 2 // the game's translucent pass
				pr.DoubleSided = true
				pr.Unlit = true
				prims[pair.Material] = pr
				order = append(order, pair.Material)
			}
			for _, pk := range shape.Packets {
				dl, err := m.ParseDL(pk.DL, nbt)
				if err != nil {
					return fmt.Errorf("%s node %d: %w", name, ni, err)
				}
				for _, p := range dl {
					base := len(pr.Positions)
					for _, v := range p.Verts {
						// The vertex names a PN slot; the packet maps slots to
						// matrix indices; a static model's matrix index is a node.
						mtx := world[ni]
						if slot := int(v.MtxSlot) / 3; slot < len(pk.MtxIdx) {
							if mi := int(pk.MtxIdx[slot]); mi >= 0 && mi < len(world) {
								mtx = world[mi]
							}
						}
						pos := m.Positions[v.Pos]
						pr.Positions = append(pr.Positions, mtx.Apply(pos))
						if v.HasNrm && int(v.Nrm) < len(m.Normals) {
							pr.Normals = append(pr.Normals, mtx.ApplyVec(m.Normals[v.Nrm]))
						}
						if v.HasTex && int(v.Tex) < len(m.Texcoords) {
							pr.UVs = append(pr.UVs, m.Texcoords[v.Tex])
						}
						if v.HasClr && int(v.Clr) < len(m.Colors) {
							pr.Colors = append(pr.Colors, m.Colors[v.Clr])
						}
					}
					for _, t := range p.Triangulate() {
						pr.Tris = append(pr.Tris, [3]uint32{uint32(base + t[0]), uint32(base + t[1]), uint32(base + t[2])})
					}
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
		if len(p.Colors) != len(p.Positions) {
			p.Colors = nil
		}
		list = append(list, *p)
	}
	if err := s.AddMesh(root, name, list); err != nil {
		return err
	}
	return s.Write(path, name)
}

// LoadSkinned pulls a model + its .key clip (and the matching .sls face
// shapes, when present) out of an archive's members.
//
// A model's vertex-animation file shares the animation's base name
// (opwf_luigi.key / opwf_luigi.sls); when present, its first shape is the
// face the shot actually shows, applied to the rest positions.
func LoadSkinned(files []lm.RARCFile, member, animName string) (*lm.MDL, *lm.Key, error) {
	var m *lm.MDL
	var key *lm.Key
	var sls *lm.SLS
	var err error
	slsName := strings.TrimSuffix(animName, ".key") + ".sls"
	for _, f := range files {
		if strings.EqualFold(f.Name, member) {
			if m, err = lm.ParseMDL(f.Data); err != nil {
				return nil, nil, err
			}
		}
		if strings.EqualFold(f.Name, animName) {
			if key, err = lm.ParseKey(f.Data); err != nil {
				return nil, nil, err
			}
		}
		if strings.EqualFold(f.Name, slsName) {
			if sls, err = lm.ParseSLS(f.Data); err != nil {
				return nil, nil, err
			}
		}
	}
	if m == nil {
		return nil, nil, fmt.Errorf("no member %q", member)
	}
	if key == nil {
		return nil, nil, fmt.Errorf("no member %q", animName)
	}
	if sls != nil {
		sls.Apply(m.Positions)
	}
	return m, key, nil
}

// SkinnedExport writes a skinned, animated GLB of "/disc/file.szp:model.mdl"
// with animation member.key from the same archive.
func SkinnedExport(src *Source, spec, animName, out string, inPlace, noFlip bool) error {
	path, member, ok := strings.Cut(spec, ":")
	if !ok {
		return fmt.Errorf("want /disc/file.szp:member.mdl, got %q", spec)
	}
	files, err := src.Archive(path)
	if err != nil {
		return err
	}
	m, key, err := LoadSkinned(files, member, animName)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	name := strings.TrimSuffix(member, ".mdl")
	return SkinnedGLB(m, key, out, name, strings.TrimSuffix(animName, ".key"), inPlace, noFlip)
}

// MDLFromArchive pulls NAME out of the RARC inside the (possibly Yay0) disc
// file given as "/disc/path.szp:NAME".
func MDLFromArchive(src *Source, spec string) (*lm.MDL, string, error) {
	path, member, ok := strings.Cut(spec, ":")
	if !ok {
		return nil, "", fmt.Errorf("want /disc/file.szp:member.mdl, got %q", spec)
	}
	files, err := src.Archive(path)
	if err != nil {
		return nil, "", err
	}
	if f := Member(files, member); f != nil {
		m, err := lm.ParseMDL(f.Data)
		return m, strings.TrimSuffix(member, ".mdl"), err
	}
	return nil, "", fmt.Errorf("no member %q in %s", member, path)
}
