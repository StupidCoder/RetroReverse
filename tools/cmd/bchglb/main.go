// bchglb exports the models of one or more BCH containers to a GLB, binding
// each mesh's albedo texture from whichever of the inputs carries it — a stage's
// models and its textures live in separate `.bch` files, so the tool takes a
// list and resolves texture names across all of them.
//
//	bchglb -o out.glb model.bch [textures.bch …]
//
// It is a development instrument (look at the geometry, check a decode); the
// game's own web export is games/captain-toad-treasure-tracker-3ds/extract.
package main

import (
	"flag"
	"fmt"
	"image"
	"os"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n3ds"
)

func main() {
	out := flag.String("o", "out.glb", "write the GLB here")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: bchglb [-o out.glb] model.bch [textures.bch …]")
		os.Exit(2)
	}
	if err := run(flag.Args(), *out); err != nil {
		fmt.Fprintln(os.Stderr, "bchglb:", err)
		os.Exit(1)
	}
}

func run(in []string, out string) error {
	var files []*n3ds.BCH
	textures := map[string]*image.NRGBA{}
	for _, p := range in {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		f, err := n3ds.ParseBCH(b)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		files = append(files, f)
		for _, e := range f.Groups[n3ds.BCHTextures] {
			t, err := f.DecodeTexture(e)
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			textures[t.Name] = t.Image
			fmt.Printf("texture %-24s %3dx%-3d format 0x%X\n", t.Name, t.Width, t.Height, t.Format)
		}
	}

	s := glb.NewScene()
	tris, missing := 0, map[string]bool{}
	for fi, f := range files {
		for _, e := range f.Groups[n3ds.BCHModels] {
			m, err := f.DecodeModel(e)
			if err != nil {
				return fmt.Errorf("%s: %w", in[fi], err)
			}
			var prims []glb.Prim
			for _, sh := range m.Meshes {
				p := glb.Prim{BaseColor: [4]float32{1, 1, 1, 1}, Unlit: true, DoubleSided: true}
				p.Positions = make([][3]float32, len(sh.Verts))
				for j, v := range sh.Verts {
					p.Positions[j] = v.Pos
				}
				if sh.HasNormal {
					p.Normals = make([][3]float32, len(sh.Verts))
					for j, v := range sh.Verts {
						p.Normals[j] = v.Normal
					}
				}
				if sh.HasColor {
					p.Colors = make([][4]uint8, len(sh.Verts))
					for j, v := range sh.Verts {
						p.Colors[j] = v.Color
					}
				}
				if sh.UVCount > 0 && sh.MaterialIndex < len(m.Materials) {
					name := m.Materials[sh.MaterialIndex].Texture()
					if img := textures[name]; img != nil {
						p.Image = img
						p.UVs = make([][2]float32, len(sh.Verts))
						for j, v := range sh.Verts {
							p.UVs[j] = [2]float32{v.UV[0][0], 1 - v.UV[0][1]}
						}
						mat := m.Materials[sh.MaterialIndex]
						switch {
						case mat.Blends:
							p.Blend = true
						case mat.AlphaTest:
							if c, ok := mat.AlphaCutoff(); ok {
								p.AlphaCutoff = c
							}
						default:
							p.Opaque = true
						}
					} else if name != "" {
						missing[name] = true
					}
				}
				p.Tris = make([][3]uint32, len(sh.Indices)/3)
				for t := range p.Tris {
					p.Tris[t] = [3]uint32{sh.Indices[t*3], sh.Indices[t*3+1], sh.Indices[t*3+2]}
				}
				tris += len(p.Tris)
				prims = append(prims, p)
			}
			node := s.AddNode(m.Name, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
			if err := s.AddMesh(node, m.Name, prims); err != nil {
				return err
			}
			fmt.Printf("model %-28s %d meshes, %d materials\n", m.Name, len(m.Meshes), len(m.Materials))
		}
	}
	for n := range missing {
		fmt.Printf("  (no texture named %q in any input)\n", n)
	}
	if err := s.Write(out, "bch"); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d tris, %d textures)\n", out, tris, len(textures))
	return nil
}
