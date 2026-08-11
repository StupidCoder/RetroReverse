// Package cgfxglb converts a decoded NW4C CGFX banner scene — the little 3-D
// model a 3DS title shows on the HOME Menu — into a self-contained binary glTF.
// Two titles share it (Super Mario 3D Land and Captain Toad: Treasure Tracker),
// which is the point: the format's traps are per-format, not per-game.
//
// The mapping is one-to-one with the CGFX: a node per bone (parented per the
// skeleton), a mesh per CGFX mesh attached to its bone's node, and any CANM
// curves carried loss-lessly as glTF CUBICSPLINE channels — the source keys are
// (frame, value, in/out-slope) hermite triples, which map onto glTF's
// (in-tangent, value, out-tangent) form with no resampling.
//
// A vertex UV set is not a texture coordinate. Each material mapper names the UV
// set it reads, a texture matrix to put it through, and its own wrap modes, and
// none of the three is reliably the default: Super Mario 3D Land's 512x256
// atlases are addressed as if they were square and their coordinators' scaleT=2
// stretches V back over the whole image, while Mario's UVs run outside [0,1]
// under a MIRRORED_REPEAT sampler. All three are carried into the GLB.
//
// The one place glTF cannot mirror the hardware is the PICA texture combiner:
// where a material multiplies a colour atlas (sampled by UV0) by a cutout mask
// (sampled by UV1), glTF has only one texture per material, so the exporter
// *bakes the combine*: it rasterises the mesh's triangles in mask-texel space,
// interpolates UV0 barycentrically per texel (exact — UV0 is affine within a
// triangle), and writes RGB from the atlas with alpha from the mask. The result
// is fragment-for-fragment what the PICA computes.
package cgfxglb

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n3ds"
)

const fps = 30 // NW4C animations are authored at 30 frames/second

// ExportBanner writes a banner CGFX out as one GLB at path, naming its
// animation clip clip. With texdump non-empty it also writes every decoded
// texture, and every baked combiner result, as PNGs into that directory.
func ExportBanner(g *n3ds.CGFX, path, clip, texdump string) error {
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
		bone := "unbound"
		if sh.BoneIndex >= 0 && sh.BoneIndex < len(model.Bones) {
			bone = model.Bones[sh.BoneIndex].Name
		}
		name := fmt.Sprintf("%s-%s", bone, mat.Name)
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

	if err := s.Write(path, clip); err != nil {
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

// buildPrim assembles one shape + material into a GLB primitive. Each of the
// material's texture mappers names its own source UV set, its own coordinate
// transform and its own wrap modes; none of those are identity in this banner,
// so all three are carried through rather than assumed.
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

	// Texture + UV selection. Masked flat faces (two mappers, two UV sets, no
	// normals — Super Mario 3D Land's title/tail quads) get the baked combiner;
	// a flat-coloured decal cut by its texture's red channel (Captain Toad's ™)
	// gets a one-colour bake; everything else samples its first mapper's
	// texture.
	switch {
	case len(mat.Mappers) >= 2 && sh.UVCount >= 2 && !sh.HasNormal:
		am, mm := &mat.Mappers[0], &mat.Mappers[1]
		atlas, mask := textures[am.Texture], textures[mm.Texture]
		if atlas == nil || mask == nil {
			return nil, fmt.Errorf("material %q references unknown textures %q/%q", mat.Name, am.Texture, mm.Texture)
		}
		uv0, err := uvArray(sh, am)
		if err != nil {
			return nil, err
		}
		uv1, err := uvArray(sh, mm)
		if err != nil {
			return nil, err
		}
		p.Image, p.UVs = bakeMasked(sh, atlas.img, mask.img, uv0, uv1, am, mm)
		// The bake covers the whole footprint, so nothing samples outside it.
		p.WrapS, p.WrapT = gltfClamp, gltfClamp
	case len(mat.Mappers) >= 1:
		m := &mat.Mappers[0]
		t := textures[m.Texture]
		if t == nil {
			return nil, fmt.Errorf("material %q references unknown texture %q", mat.Name, m.Texture)
		}
		uvs, err := uvArray(sh, m)
		if err != nil {
			return nil, err
		}
		p.Image = t.img
		// A stage that replaces the colour with its constant and takes alpha
		// from a texture's red channel is a flat-coloured decal: the texture is
		// a stencil, not a picture. Captain Toad's ™ is one, and its ETC1
		// texture has no alpha channel at all — sampled as a picture it is an
		// opaque black square.
		//
		// When the alpha instead comes from a unit this path does not sample —
		// a secondary shade layer over a mesh that carries its own geometry —
		// it is dropped. That is the exporter's one documented approximation:
		// baking it would mean rasterising into the shade layer's own
		// resolution, which is far coarser than the mesh.
		if unit, ok := mat.Stage0.AlphaFromRed(); ok && unit == 0 {
			rgb, flat := mat.Stage0.ConstantColor()
			if !flat {
				return nil, fmt.Errorf("material %q takes alpha from red but its colour is not the stage constant (sources %v, combine %d)",
					mat.Name, mat.Stage0.ColorSrc, mat.Stage0.CombineColor)
			}
			p.Image = stencilImage(t.img, rgb)
		}
		p.UVs = flipV(uvs)
		p.WrapS, err = gltfWrap(m.WrapS)
		if err != nil {
			return nil, fmt.Errorf("material %q: %w", mat.Name, err)
		}
		p.WrapT, err = gltfWrap(m.WrapT)
		if err != nil {
			return nil, fmt.Errorf("material %q: %w", mat.Name, err)
		}
	}
	return p, nil
}

// glTF sampler wrap enums.
const (
	gltfRepeat   = 10497
	gltfClamp    = 33071
	gltfMirrored = 33648
)

// gltfWrap maps a PICA wrap mode onto glTF's. CLAMP_TO_BORDER has no glTF
// equivalent (glTF has no border colour), so it is refused rather than silently
// downgraded to clamp-to-edge.
func gltfWrap(w n3ds.PICAWrap) (int, error) { return w.GLTF() }

// uvArray extracts the UV set a mapper reads and puts it through the mapper's
// texture matrix. The result is still in PICA texture space (V up).
func uvArray(sh *n3ds.Shape, m *n3ds.TexMapper) ([][2]float32, error) {
	if m.SourceUV < 0 || m.SourceUV >= sh.UVCount {
		return nil, fmt.Errorf("texture %q reads UV set %d but the shape has %d", m.Texture, m.SourceUV, sh.UVCount)
	}
	uvs := make([][2]float32, len(sh.Verts))
	for i, v := range sh.Verts {
		uv := v.UV0
		if m.SourceUV == 1 {
			uv = v.UV1
		}
		uvs[i] = m.Apply(uv)
	}
	return uvs, nil
}

// flipV converts PICA texture space (V up) to glTF's (V down).
func flipV(uvs [][2]float32) [][2]float32 {
	out := make([][2]float32, len(uvs))
	for i, uv := range uvs {
		out[i] = [2]float32{uv[0], 1 - uv[1]}
	}
	return out
}

// bakeMasked rasterises the PICA two-texture combine for a masked flat face:
// output texels are mask texels, each taking its colour from the atlas at the
// barycentrically interpolated UV0 and its alpha from the mask's own value.
//
// The output covers the *footprint* of the transformed mask UVs rather than the
// whole mask, and the returned UVs address that footprint. The title planes
// reach past the mask's top edge — with the coordinate transform applied their
// V runs to 1.02 — so baking into the mask's own rectangle would clip the top
// of the logo off. Sampling inside the bake still wraps exactly as the hardware
// does, so the texels that come from beyond the edge are the ones the PICA
// would have fetched.
func bakeMasked(sh *n3ds.Shape, atlas, mask *image.NRGBA, uv0, uv1 [][2]float32, am, mm *n3ds.TexMapper) (*image.NRGBA, [][2]float32) {
	W, H := mask.Rect.Dx(), mask.Rect.Dy()

	// Per-vertex positions in mask-texel space, V flipped to image rows.
	px := make([][2]float32, len(sh.Verts))
	for i, uv := range uv1 {
		px[i] = [2]float32{uv[0] * float32(W), (1 - uv[1]) * float32(H)}
	}
	// Footprint, snapped out to whole texels with a texel of margin.
	lo := [2]float32{px[0][0], px[0][1]}
	hi := lo
	for _, p := range px[1:] {
		for k := 0; k < 2; k++ {
			lo[k] = minf(lo[k], p[k])
			hi[k] = maxf(hi[k], p[k])
		}
	}
	x0, y0 := int(floorf(lo[0]))-1, int(floorf(lo[1]))-1
	x1, y1 := int(ceilf(hi[0]))+1, int(ceilf(hi[1]))+1
	outW, outH := x1-x0, y1-y0
	out := image.NewNRGBA(image.Rect(0, 0, outW, outH))

	for t := 0; t+2 < len(sh.Indices); t += 3 {
		i0, i1, i2 := sh.Indices[t], sh.Indices[t+1], sh.Indices[t+2]
		a, b, c := px[i0], px[i1], px[i2]
		den := (b[1]-c[1])*(a[0]-c[0]) + (c[0]-b[0])*(a[1]-c[1])
		if den == 0 {
			continue
		}
		minX := clampi(int(min3(a[0], b[0], c[0]))-x0, 0, outW-1)
		maxX := clampi(int(max3(a[0], b[0], c[0]))+1-x0, 0, outW-1)
		minY := clampi(int(min3(a[1], b[1], c[1]))-y0, 0, outH-1)
		maxY := clampi(int(max3(a[1], b[1], c[1]))+1-y0, 0, outH-1)
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				fx, fy := float32(x+x0)+0.5, float32(y+y0)+0.5
				w0 := ((b[1]-c[1])*(fx-c[0]) + (c[0]-b[0])*(fy-c[1])) / den
				w1 := ((c[1]-a[1])*(fx-c[0]) + (a[0]-c[0])*(fy-c[1])) / den
				w2 := 1 - w0 - w1
				if w0 < -0.001 || w1 < -0.001 || w2 < -0.001 {
					continue
				}
				u := w0*uv0[i0][0] + w1*uv0[i1][0] + w2*uv0[i2][0]
				v := w0*uv0[i0][1] + w1*uv0[i1][1] + w2*uv0[i2][1]
				r, g, bb, _ := sampleWrapped(atlas, u, v, am.WrapS, am.WrapT)
				o := out.PixOffset(x, y)
				out.Pix[o], out.Pix[o+1], out.Pix[o+2] = r, g, bb
				// The mask is texel-aligned with the output, so its own wrap
				// rule is applied to the integer coordinate directly.
				mx := wrapTexel(x+x0, W, mm.WrapS)
				my := wrapTexel(y+y0, H, mm.WrapT)
				if mx < 0 || my < 0 {
					continue // clamp-to-border: outside is transparent
				}
				out.Pix[o+3] = mask.Pix[mask.PixOffset(mx, my)] // L4 mask: grey = coverage
			}
		}
	}

	// UVs addressing the baked footprint, already in glTF's V-down space.
	uvs := make([][2]float32, len(px))
	for i, p := range px {
		uvs[i] = [2]float32{(p[0] - float32(x0)) / float32(outW), (p[1] - float32(y0)) / float32(outH)}
	}
	return out, uvs
}

// stencilImage realises a flat-coloured decal: every texel takes the stage's
// constant colour, and its alpha comes from the source texture's red channel.
func stencilImage(src *image.NRGBA, rgb [3]uint8) *image.NRGBA {
	b := src.Rect
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			o, so := out.PixOffset(x, y), src.PixOffset(x+b.Min.X, y+b.Min.Y)
			out.Pix[o], out.Pix[o+1], out.Pix[o+2] = rgb[0], rgb[1], rgb[2]
			out.Pix[o+3] = src.Pix[so] // the red channel is the stencil
		}
	}
	return out
}

// sampleWrapped is a nearest-neighbour fetch honouring the PICA wrap modes.
func sampleWrapped(img *image.NRGBA, u, v float32, ws, wt n3ds.PICAWrap) (byte, byte, byte, byte) {
	w, h := img.Rect.Dx(), img.Rect.Dy()
	x := wrapTexel(int(floorf(u*float32(w))), w, ws)
	y := wrapTexel(int(floorf((1-v)*float32(h))), h, wt)
	if x < 0 || y < 0 {
		return 0, 0, 0, 0
	}
	o := img.PixOffset(x, y)
	return img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3]
}

// wrapTexel applies one wrap mode to an integer texel index, returning -1 for
// "outside" under clamp-to-border. It mirrors gpu_texture.go's wrapCoord.
func wrapTexel(v, n int, mode n3ds.PICAWrap) int {
	switch mode {
	case n3ds.WrapClampToEdge:
		return clampi(v, 0, n-1)
	case n3ds.WrapClampToBorder:
		if v < 0 || v >= n {
			return -1
		}
		return v
	case n3ds.WrapRepeat:
		v %= n
		if v < 0 {
			v += n
		}
		return v
	default: // mirrored repeat
		period := 2 * n
		v %= period
		if v < 0 {
			v += period
		}
		if v >= n {
			v = period - 1 - v
		}
		return v
	}
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
func minf(a, b float32) float32 {
	if b < a {
		return b
	}
	return a
}
func maxf(a, b float32) float32 {
	if b > a {
		return b
	}
	return a
}
func floorf(v float32) float32 { return float32(math.Floor(float64(v))) }
func ceilf(v float32) float32  { return float32(math.Ceil(float64(v))) }
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
