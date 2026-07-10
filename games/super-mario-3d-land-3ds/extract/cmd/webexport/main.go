// webexport extracts Super Mario 3D Land's web deliverables from the cartridge
// image into site/public/super-mario-3d-land-3ds/. The first asset is the
// HOME-Menu banner: the animated 3-D logo scene (CBMD → LZ11 → CGFX), exported
// as one GLB with embedded PNG textures and the CANM skeletal animation baked
// into per-bone node translation tracks.
//
//	webexport -in game.cci [-o DIR] [-texdump DIR]
//
// The GLB maps the CGFX one-to-one: a node per bone (the logo's parts — Block,
// Mario, SuperLeaf, the title faces and the Tanooki tail — are each rigidly
// bound to one bone), a mesh per CGFX mesh attached to its bone's node, and the
// hermite animation keys carried loss-lessly as glTF CUBICSPLINE channels.
//
// The PICA texture combiner blends two textures on the title materials: the
// colour atlas (COMMON1, sampled by UV0) times a 4-bit cutout mask (COMMON7,
// sampled by UV1) — the title's flat front/back faces and the tail are plain
// rectangles cut to the logo silhouette by the mask. glTF samples one texture
// per material, so for those meshes the exporter *bakes the combiner*: it
// rasterises the mesh's triangles in UV1 space at the mask's resolution,
// interpolates UV0 barycentrically per texel (exact — UV0 is affine within a
// triangle), and writes RGB from the atlas with alpha from the mask. The result
// is fragment-for-fragment what the PICA computes. The extruded title-side mesh
// (which has real letter geometry and uses its second texture only as a depth
// shade) keeps the plain atlas; the shade layer is dropped — noted in the
// writeup as the one approximation.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n3ds"
)

const fps = 30 // NW4C animations are authored at 30 frames/second

func main() {
	in := flag.String("in", "", "3DS cartridge image (decrypted .cci)")
	out := flag.String("o", "../../site/public/super-mario-3d-land-3ds", "output root")
	texdump := flag.String("texdump", "", "also write each decoded texture as PNG into this directory")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "usage: webexport -in game.cci [-o DIR] [-texdump DIR]")
		os.Exit(2)
	}
	if err := run(*in, *out, *texdump); err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}

func run(in, out, texdump string) error {
	img, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	ncsd, err := n3ds.ParseNCSD(img)
	if err != nil {
		return err
	}
	cxi, err := ncsd.Executable()
	if err != nil {
		return err
	}
	efs, err := cxi.ExeFS()
	if err != nil {
		return err
	}
	raw, err := efs.File("banner")
	if err != nil {
		return err
	}
	bn, err := n3ds.ParseBanner(raw)
	if err != nil {
		return err
	}
	blob, err := bn.CommonModel()
	if err != nil {
		return err
	}
	g, err := n3ds.ParseCGFX(blob)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(out, "models"), 0o755); err != nil {
		return err
	}
	if err := exportBanner(g, filepath.Join(out, "models", "banner.glb"), texdump); err != nil {
		return err
	}
	return writeManifest(out)
}

func exportBanner(g *n3ds.CGFX, path, texdump string) error {
	models := g.Resources["Models"]
	if len(models) != 1 {
		return fmt.Errorf("expected 1 model in the banner, found %d", len(models))
	}
	model, err := g.DecodeModel(models[0])
	if err != nil {
		return err
	}

	// Decode every texture once, by name.
	textures := map[string]*pngTex{}
	for _, te := range g.Resources["Textures"] {
		txob, im, err := g.DecodeTexture(te)
		if err != nil {
			return err
		}
		textures[te.Name] = &pngTex{txob: txob, img: im}
		if texdump != "" {
			os.MkdirAll(texdump, 0o755)
			f, err := os.Create(filepath.Join(texdump, te.Name+".png"))
			if err != nil {
				return err
			}
			if err := png.Encode(f, im); err != nil {
				f.Close()
				return err
			}
			f.Close()
			fmt.Printf("texture %-10s %dx%d format 0x%04x/0x%04x -> %s.png\n",
				te.Name, txob.Width, txob.Height, txob.GLFormat, txob.GLType, te.Name)
		}
	}

	// One animation expected; tolerate none.
	var anim *n3ds.SkelAnim
	if as := g.Resources["SkeletalAnimations"]; len(as) > 0 {
		if anim, err = g.DecodeSkeletalAnim(as[0]); err != nil {
			return err
		}
	}

	s := glb.NewScene()

	// A node per bone, in skeleton order, parented per the skeleton. Rest
	// rotation is a CGFX XYZ Euler triple; the banner's bones are all
	// rotationless at rest, so only assert that rather than convert.
	nodeOf := make([]int, len(model.Bones))
	for i, b := range model.Bones {
		if b.Rot != ([3]float32{}) {
			return fmt.Errorf("bone %q has a rest rotation %v; Euler conversion not implemented", b.Name, b.Rot)
		}
		parent := -1
		if b.Parent >= 0 {
			parent = nodeOf[b.Parent]
		}
		nodeOf[i] = s.AddNode(b.Name, parent, b.Trans, [4]float32{0, 0, 0, 1}, b.Scale)
	}

	// A mesh per CGFX mesh, on its shape's bone node.
	for mi, mesh := range model.Meshes {
		sh := &model.Shapes[mesh.ShapeIndex]
		mat := &model.Materials[mesh.MaterialIndex]
		prim, err := buildPrim(sh, mat, textures)
		if err != nil {
			return fmt.Errorf("mesh %d (%s): %w", mi, mat.Name, err)
		}
		node := 0
		if sh.BoneIndex >= 0 && sh.BoneIndex < len(nodeOf) {
			node = nodeOf[sh.BoneIndex]
		}
		name := fmt.Sprintf("%s-%s", model.Bones[sh.BoneIndex].Name, mat.Name)
		if texdump != "" && prim.Image != nil {
			baked := true
			for _, t := range textures {
				if prim.Image == t.img {
					baked = false
				}
			}
			if baked {
				f, err := os.Create(filepath.Join(texdump, fmt.Sprintf("baked-%s.png", name)))
				if err == nil {
					png.Encode(f, prim.Image)
					f.Close()
				}
			}
		}
		if err := s.AddMesh(node, name, []glb.Prim{*prim}); err != nil {
			return err
		}
	}

	// Bake the CANM: for each animated bone, a CUBICSPLINE translation track.
	// A member's curves animate single components; the rest stay at the bone's
	// rest translation. Slopes are per frame; glTF tangents are per second.
	if anim != nil {
		byName := map[string]int{}
		for i, b := range model.Bones {
			byName[b.Name] = i
		}
		for _, m := range anim.Members {
			bi, ok := byName[m.Bone]
			if !ok {
				return fmt.Errorf("animation targets unknown bone %q", m.Bone)
			}
			if err := addTrack(s, nodeOf[bi], model.Bones[bi], m); err != nil {
				return err
			}
		}
	}

	if err := s.Write(path, "banner"); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d bones, %d meshes, %d textures)\n", path, len(model.Bones), len(model.Meshes), len(textures))
	return nil
}

type pngTex struct {
	txob *n3ds.TXOB
	img  *image.NRGBA
}

func (t *pngTex) image() *image.NRGBA { return t.img }

// buildPrim assembles one shape + material into a GLB primitive. Material→UV
// binding: one texture uses mapper 0 over UV0; two textures use the second
// mapper over UV1 (the detail/mask layer — see the package comment).
func buildPrim(sh *n3ds.Shape, mat *n3ds.Material, textures map[string]*pngTex) (*glb.Prim, error) {
	p := &glb.Prim{
		BaseColor:   [4]float32{1, 1, 1, 1},
		Unlit:       true,
		DoubleSided: true,
	}
	p.Positions = make([][3]float32, len(sh.Verts))
	for i, v := range sh.Verts {
		p.Positions[i] = v.Pos
	}
	if sh.HasNormal {
		p.Normals = make([][3]float32, len(sh.Verts))
		for i, v := range sh.Verts {
			p.Normals[i] = v.Normal
		}
	}
	if sh.HasColor {
		p.Colors = make([][4]uint8, len(sh.Verts))
		for i, v := range sh.Verts {
			p.Colors[i] = v.Color
		}
	}
	if len(sh.Indices)%3 != 0 {
		return nil, fmt.Errorf("index count %d is not a triangle list", len(sh.Indices))
	}
	p.Tris = make([][3]uint32, len(sh.Indices)/3)
	for i := range p.Tris {
		p.Tris[i] = [3]uint32{sh.Indices[i*3], sh.Indices[i*3+1], sh.Indices[i*3+2]}
	}

	// Texture + UV selection. Masked flat faces (two textures, two UV sets, no
	// normals — the title/tail quads) get the baked combiner; everything else
	// samples its first texture by UV0.
	switch {
	case len(mat.Textures) >= 2 && sh.UVCount >= 2 && !sh.HasNormal:
		atlas, mask := textures[mat.Textures[0]], textures[mat.Textures[1]]
		if atlas == nil || mask == nil {
			return nil, fmt.Errorf("material %q references unknown textures %v", mat.Name, mat.Textures)
		}
		p.Image = bakeMasked(sh, atlas.img, mask.img)
		p.UVs = uvArray(sh, 1)
	case len(mat.Textures) >= 1 && sh.UVCount >= 1:
		t := textures[mat.Textures[0]]
		if t == nil {
			return nil, fmt.Errorf("material %q references unknown texture %q", mat.Name, mat.Textures[0])
		}
		p.Image = t.img
		p.UVs = uvArray(sh, 0)
	}
	return p, nil
}

// uvArray extracts one UV set, flipped from the PICA's V-up to glTF's V-down.
func uvArray(sh *n3ds.Shape, set int) [][2]float32 {
	uvs := make([][2]float32, len(sh.Verts))
	for i, v := range sh.Verts {
		uv := v.UV0
		if set == 1 {
			uv = v.UV1
		}
		uvs[i] = [2]float32{uv[0], 1 - uv[1]}
	}
	return uvs
}

// bakeMasked rasterises the PICA two-texture combine for a masked flat face:
// output texels live in UV1 (mask) space; each covered texel takes its colour
// from the atlas at the barycentrically interpolated UV0 and its alpha from the
// mask's own value. Texels no triangle covers stay transparent.
func bakeMasked(sh *n3ds.Shape, atlas, mask *image.NRGBA) *image.NRGBA {
	W, H := mask.Rect.Dx(), mask.Rect.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, W, H))

	// Per-vertex positions in mask-pixel space (V flipped to image rows).
	px := make([][2]float32, len(sh.Verts))
	for i, v := range sh.Verts {
		px[i] = [2]float32{v.UV1[0] * float32(W), (1 - v.UV1[1]) * float32(H)}
	}
	sample := func(img *image.NRGBA, u, v float32) (byte, byte, byte, byte) {
		x := int(u * float32(img.Rect.Dx()))
		y := int((1 - v) * float32(img.Rect.Dy()))
		x = clampi(x, 0, img.Rect.Dx()-1)
		y = clampi(y, 0, img.Rect.Dy()-1)
		o := img.PixOffset(x, y)
		return img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3]
	}

	for t := 0; t+2 < len(sh.Indices); t += 3 {
		i0, i1, i2 := sh.Indices[t], sh.Indices[t+1], sh.Indices[t+2]
		a, b, c := px[i0], px[i1], px[i2]
		minX := clampi(int(min3(a[0], b[0], c[0])), 0, W-1)
		maxX := clampi(int(max3(a[0], b[0], c[0]))+1, 0, W-1)
		minY := clampi(int(min3(a[1], b[1], c[1])), 0, H-1)
		maxY := clampi(int(max3(a[1], b[1], c[1]))+1, 0, H-1)
		den := (b[1]-c[1])*(a[0]-c[0]) + (c[0]-b[0])*(a[1]-c[1])
		if den == 0 {
			continue
		}
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				fx, fy := float32(x)+0.5, float32(y)+0.5
				w0 := ((b[1]-c[1])*(fx-c[0]) + (c[0]-b[0])*(fy-c[1])) / den
				w1 := ((c[1]-a[1])*(fx-c[0]) + (a[0]-c[0])*(fy-c[1])) / den
				w2 := 1 - w0 - w1
				if w0 < -0.001 || w1 < -0.001 || w2 < -0.001 {
					continue
				}
				u0 := w0*sh.Verts[i0].UV0[0] + w1*sh.Verts[i1].UV0[0] + w2*sh.Verts[i2].UV0[0]
				v0 := w0*sh.Verts[i0].UV0[1] + w1*sh.Verts[i1].UV0[1] + w2*sh.Verts[i2].UV0[1]
				r, g, bb, _ := sample(atlas, u0, v0)
				mo := mask.PixOffset(x, y)
				o := out.PixOffset(x, y)
				out.Pix[o], out.Pix[o+1], out.Pix[o+2] = r, g, bb
				out.Pix[o+3] = mask.Pix[mo] // L4 mask: gray value = coverage
			}
		}
	}
	return out
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func min3(a, b, c float32) float32 {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}
func max3(a, b, c float32) float32 {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

// addTrack emits one bone's translation channel from its curves.
func addTrack(s *glb.Scene, node int, bone n3ds.Bone, m n3ds.BoneAnim) error {
	for slot, c := range m.Curves {
		if c == nil {
			continue
		}
		if slot < n3ds.SlotTransX {
			return fmt.Errorf("bone %q animates slot %d; only translation tracks are implemented", m.Bone, slot)
		}
		comp := slot - n3ds.SlotTransX
		times := make([]float32, len(c.Keys))
		vals := make([][3]float32, len(c.Keys))
		tans := make([][3]float32, len(c.Keys))
		for i, k := range c.Keys {
			times[i] = k.Frame / fps
			v := bone.Trans
			v[comp] = k.Value
			vals[i] = v
			var tan [3]float32
			tan[comp] = k.Slope * fps
			tans[i] = tan
		}
		s.AddTranslationTrack(node, times, vals, tans, tans)
	}
	return nil
}

func writeManifest(out string) error {
	m := map[string]any{
		"format":   2,
		"game":     "super-mario-3d-land-3ds",
		"platform": "Nintendo 3DS",
		"native":   map[string]int{"w": 400, "h": 240},
		"tickHz":   60,
		"models": []map[string]any{{
			"name":    "HOME Menu Banner",
			"file":    "models/banner.glb",
			"kind":    "mesh3d",
			"section": "Banner",
		}},
	}
	b, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "manifest.json"), append(b, '\n'), 0o644)
}
