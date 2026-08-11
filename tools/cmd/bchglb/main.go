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
	"strings"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n3ds"
	"retroreverse.com/tools/platform/n3ds/bchglb"
)

func main() {
	out := flag.String("o", "out.glb", "write the GLB here")
	only := flag.String("model", "", "export only the model with this exact name")
	clips := flag.String("clips", "", "comma-separated skeletal animations to export as glTF clips (\"all\" for every one that decodes)")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: bchglb [-o out.glb] model.bch [textures.bch …]")
		os.Exit(2)
	}
	if err := run(flag.Args(), *out, *only, *clips); err != nil {
		fmt.Fprintln(os.Stderr, "bchglb:", err)
		os.Exit(1)
	}
}

func run(in []string, out, only, clips string) error {
	var files []*n3ds.BCH
	textures := map[string]*image.NRGBA{}
	for _, p := range in {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		// A `.szs` is the shipping form (Yaz0 → SARC); the `.bch` inside it is
		// what this reads, and taking either saves unpacking by hand.
		var blobs [][]byte
		if len(b) >= 4 && string(b[:4]) == "BCH\x00" {
			blobs = [][]byte{b}
		} else {
			arc, err := n3ds.OpenSZS(b)
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			for _, f := range arc.Files {
				if strings.HasSuffix(f.Name, ".bch") {
					blobs = append(blobs, f.Data)
				}
			}
			if len(blobs) == 0 {
				return fmt.Errorf("%s: the archive holds no .bch", p)
			}
		}
		for _, blob := range blobs {
			if err := addFile(p, blob, &files, textures); err != nil {
				return err
			}
		}
	}
	return build(in, files, textures, out, only, clips)
}

func addFile(p string, b []byte, files *[]*n3ds.BCH, textures map[string]*image.NRGBA) error {
	f, err := n3ds.ParseBCH(b)
	if err != nil {
		return fmt.Errorf("%s: %w", p, err)
	}
	*files = append(*files, f)
	for _, e := range f.Groups[n3ds.BCHTextures] {
		t, err := f.DecodeTexture(e)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		textures[t.Name] = t.Image
		fmt.Printf("texture %-24s %3dx%-3d format 0x%X\n", t.Name, t.Width, t.Height, t.Format)
	}
	return nil
}

func build(in []string, files []*n3ds.BCH, textures map[string]*image.NRGBA, out, only, clips string) error {
	s := glb.NewScene()
	tris, missing := 0, map[string]bool{}
	for fi, f := range files {
		for _, e := range f.Groups[n3ds.BCHModels] {
			if only != "" && e.Name != only {
				continue
			}
			m, err := f.DecodeModel(e)
			if err != nil {
				return fmt.Errorf("%s: %w", in[fi], err)
			}
			// The skeleton first, so the meshes can bind to it. A model with no
			// bones gets an empty rig and every mesh stays unskinned.
			rig := bchglb.AddSkeleton(s, m, -1, m.Name+":")
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
					// The material's combiner, not its texture, is the surface
					// colour (see BCHMaterial.BakeAlbedo).
					img, err := m.Materials[sh.MaterialIndex].BakeAlbedo(textures)
					if err != nil {
						return err
					}
					if textures[name] == nil && name != "" {
						missing[name] = true
					}
					if img != nil {
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
					}
				}
				p.Tris = make([][3]uint32, len(sh.Indices)/3)
				for t := range p.Tris {
					p.Tris[t] = [3]uint32{sh.Indices[t*3], sh.Indices[t*3+1], sh.Indices[t*3+2]}
				}
				tris += len(p.Tris)

				// One node per mesh: a skinned mesh's inverse-bind matrices
				// depend on how *that* mesh is bound, so each carries its own
				// skin (see tools/platform/n3ds/bchglb).
				bchglb.BindJoints(&p, &sh)
				node := s.AddNode(m.Name+"/"+matName(m, &sh), -1,
					[3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
				if err := s.AddMesh(node, m.Name, []glb.Prim{p}); err != nil {
					return err
				}
				if sk := rig.Skin(s, &sh); sk >= 0 {
					s.SetNodeSkin(node, sk)
				}
			}
			fmt.Printf("model %-28s %d meshes, %d materials, %d bones\n", m.Name, len(m.Meshes), len(m.Materials), len(m.Bones))
			if clips != "" && len(m.Bones) > 0 {
				n, err := addClips(s, rig, files, clips)
				if err != nil {
					return err
				}
				fmt.Printf("  %d animation clips\n", n)
			}
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

// matName is the mesh's material name, which is what the artists called the
// part — a readable node name in a viewer.
func matName(m *n3ds.BCHModel, sh *n3ds.BCHMesh) string {
	if sh.MaterialIndex < len(m.Materials) {
		return m.Materials[sh.MaterialIndex].Name
	}
	return "mesh"
}

// addClips exports the named skeletal animations from whichever input holds
// them. "all" takes every animation that decodes; the rest halt the run, since
// a clip asked for by name and silently missing is worse than a stop.
func addClips(s *glb.Scene, rig *bchglb.Rig, files []*n3ds.BCH, clips string) (int, error) {
	want := map[string]bool{}
	all := clips == "all"
	if !all {
		for _, n := range strings.Split(clips, ",") {
			want[n] = true
		}
	}
	n := 0
	for _, f := range files {
		for _, e := range f.Groups[n3ds.BCHSkeletalAnims] {
			if !all && !want[e.Name] {
				continue
			}
			a, err := f.DecodeSkeletalAnim(e)
			if err != nil {
				if all {
					continue // an encoding this decoder does not model yet
				}
				return n, fmt.Errorf("%s: %w", e.Name, err)
			}
			if err := bchglb.AddClip(s, rig, a, e.Name); err != nil {
				return n, err
			}
			delete(want, e.Name)
			n++
		}
	}
	for k := range want {
		return n, fmt.Errorf("no skeletal animation named %q in any input", k)
	}
	return n, nil
}
