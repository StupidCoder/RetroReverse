// objects.go ships the POLDC0 shared-object models: the sky, the four
// player cabs and their drivers, and the traffic the drives spawn
// (Part XIII of the writeup).
//
// Everything here was identified by tracing the game's own pointers, not
// by shape. The renderer dispatches every drawn model through jsr
// 0C080E20 with r4 = the model pointer; logging r4 across the driver-
// select carousel (each position renders one character, named on screen)
// and across four in-game drives (one per selected driver) separates each
// driver's parts, each cab's parts, and the driver-independent set. The
// placement numbers come from the same dispatch point: the game loads the
// model's transform into XMTRX before the jsr, so capturing the back FP
// bank per dispatch and solving inv(X_ref)·X_part gives every part's
// placement in the reference part's frame — wheel anchors from the cab
// (identity rotation at standstill), standing poses from the driver-
// select screen (poses.go). The sky dome (model 272) and horizon ring
// (model 258) render camera-centred at uniform scale 2, read off the
// same captures; their textures are the boot directory's entries 150 and
// 108 (the cloud panorama ships byte-identical again in TEXDC1 and
// TEXDC3, so every course shares one sky).
package main

import (
	"fmt"
	"image"

	"retroreverse.com/games/crazy-taxi-dc/extract/assets"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/dc"
)

// placed is one model placed in an assembly's frame: rows 0-2 of the
// affine transform [R|t], row-major.
type placed struct {
	model int
	label string
	m     [12]float32
}

var ident = [12]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0}

func at(model int, label string, x, y, z float32) placed {
	return placed{model, label, [12]float32{1, 0, 0, x, 0, 1, 0, y, 0, 0, 1, z}}
}

// The four player cabs. Names are the game's own driver-select captions;
// wheel anchors are the captured dispatch transforms — each cab has its
// own track and wheelbase, and Axel's and B.D.Joe's cabs share one set of
// wheel models. The cab faces +z and the driver sits on the +x (left)
// side, read off the seated driver's captured position.
var cabs = []struct {
	id, name, driver string
	parts            []placed
}{
	{"cab-axel", "Axel's cab", "axel", []placed{
		{9, "body", ident},
		{13, "roof sign", ident},
		{14, "interior", ident},
		at(7, "wheel front left", 6.9, 3.188, 15.5),
		at(8, "wheel front right", -6.9, 3.188, 15.5),
		at(10, "wheel rear left", 6.5, 3.096, -15.5),
		at(11, "wheel rear right", -6.5, 3.096, -15.5),
	}},
	{"cab-bdjoe", "B.D.Joe's cab", "bdjoe", []placed{
		{25, "body", ident},
		{27, "roof sign", ident},
		{28, "interior", ident},
		at(7, "wheel front left", 6.9, 3.188, 15.25),
		at(8, "wheel front right", -6.9, 3.188, 15.25),
		at(10, "wheel rear left", 6.5, 3.097, -15.25),
		at(11, "wheel rear right", -6.5, 3.097, -15.25),
	}},
	{"cab-gena", "Gena's cab", "gena", []placed{
		{51, "body", ident},
		{55, "roof sign", ident},
		{56, "interior", ident},
		at(49, "wheel front left", 6.6, 3.188, 14.6),
		at(50, "wheel front right", -6.6, 3.188, 14.6),
		at(52, "wheel rear left", 6.9, 3.096, -14.6),
		at(53, "wheel rear right", -6.9, 3.096, -14.6),
	}},
	{"cab-gus", "Gus's cab", "gus", []placed{
		{37, "body", ident},
		{41, "roof sign", ident},
		{42, "interior", ident},
		at(35, "wheel front left", 6.8, 3.190, 14.5),
		at(36, "wheel front right", -6.8, 3.190, 14.5),
		at(38, "wheel rear left", 7.1, 3.095, -14.5),
		at(39, "wheel rear right", -7.1, 3.095, -14.5),
	}},
}

var driverNames = map[string]string{
	"axel": "Axel", "bdjoe": "B.D.Joe", "gena": "Gena", "gus": "Gus",
}

// The traffic fleet: driver-independent vehicles drawn in every drive.
// Each is a photo-quad impostor pair — a coarse textured body plus one
// two-quad wheel model the game places once per axle (captured anchors;
// the quads carry the wheel photograph and spin about x at speed). The
// 56-vertex variant heads (models 119/183/193/237, popped-glass states)
// and the mid model 209 share these bodies' textures and stay out of the
// gallery.
var traffic = []struct {
	id, name string
	parts    []placed
}{
	{"traffic-120", "Traffic car (model 120)", []placed{
		{120, "body", ident},
		at(126, "front axle wheels", 0, 3.1, 13.5),
		at(126, "rear axle wheels", 0, 3.1, -13.5),
	}},
	{"traffic-184", "Traffic car (model 184)", []placed{
		{184, "body", ident},
		at(190, "front axle wheels", 0, 3.2, 14.9),
		at(190, "rear axle wheels", 0, 3.2, -14.9),
	}},
	{"traffic-194", "Traffic car (model 194)", []placed{
		{194, "body", ident},
		at(200, "front axle wheels", 0, 3.4, 15.9),
		at(200, "rear axle wheels", 0, 3.4, -15.9),
	}},
	{"traffic-238", "Traffic car (model 238)", []placed{
		{238, "body", ident},
		at(244, "front axle wheels", 0, 3.2, 14.0),
		at(244, "rear axle wheels", 0, 3.2, -14.0),
	}},
}

// objectSet is the loaded POLDC0 world: models plus the boot texture
// directory they index.
type objectSet struct {
	models []*assets.Model
	td     *assets.TexDir
	texdc  []byte
	cache  map[int]image.Image
}

func loadObjectSet(disc *dc.Disc, firstRead []byte) (*objectSet, error) {
	poldc, err := disc.Vol.ReadFile("POLDC0.BIN;1")
	if err != nil {
		return nil, err
	}
	models, err := assets.OpenModels(poldc)
	if err != nil {
		return nil, fmt.Errorf("POLDC0: %w", err)
	}
	td, err := assets.OpenTexDir(firstRead, 0)
	if err != nil {
		return nil, err
	}
	texdc, err := disc.Vol.ReadFile("TEXDC0.BIN;1")
	if err != nil {
		return nil, err
	}
	return &objectSet{models: models, td: td, texdc: texdc, cache: map[int]image.Image{}}, nil
}

func (s *objectSet) texture(aux int) (image.Image, error) {
	if img, ok := s.cache[aux]; ok {
		return img, nil
	}
	img, err := s.td.Decode(aux, s.texdc)
	if err != nil {
		return nil, fmt.Errorf("texture %d: %w", aux, err)
	}
	s.cache[aux] = img
	return img, nil
}

// node assembles one placed model into a named glTF node: positions and
// normals through the part's transform, one textured group per texture id
// (translucent-list blocks blend) and one colour pool for the rest.
func (s *objectSet) node(p placed) (glb.VariantNode, error) {
	if p.model < 0 || p.model >= len(s.models) {
		return glb.VariantNode{}, fmt.Errorf("no model %d", p.model)
	}
	m := s.models[p.model]
	n := glb.VariantNode{Name: p.label}
	type key struct {
		aux   int
		blend bool
	}
	texTris := map[key][][3]uint32{}
	var order []key
	var greyTris [][3]uint32
	xf := p.m
	for _, blk := range m.Blocks {
		blend := blk.ListType() == 2
		for _, st := range blk.Strips {
			base := uint32(len(n.Positions))
			for _, v := range st.Verts {
				x, y, z := v.Pos[0], v.Pos[1], v.Pos[2]
				n.Positions = append(n.Positions, [3]float32{
					xf[0]*x + xf[1]*y + xf[2]*z + xf[3],
					xf[4]*x + xf[5]*y + xf[6]*z + xf[7],
					xf[8]*x + xf[9]*y + xf[10]*z + xf[11],
				})
				nx, ny, nz := v.Normal[0], v.Normal[1], v.Normal[2]
				n.Normals = append(n.Normals, [3]float32{
					xf[0]*nx + xf[1]*ny + xf[2]*nz,
					xf[4]*nx + xf[5]*ny + xf[6]*nz,
					xf[8]*nx + xf[9]*ny + xf[10]*nz,
				})
				n.UVs = append(n.UVs, [2]float32{v.U, v.V})
			}
			for _, t := range st.Tris() {
				tri := [3]uint32{base + uint32(t[0]), base + uint32(t[1]), base + uint32(t[2])}
				if blk.Textured() && blk.Aux != 0xFFFFFFFF && int(blk.Aux) < len(s.td.Entries) {
					k := key{int(blk.Aux), blend}
					if _, ok := texTris[k]; !ok {
						order = append(order, k)
					}
					texTris[k] = append(texTris[k], tri)
				} else {
					greyTris = append(greyTris, tri)
				}
			}
		}
	}
	for _, k := range order {
		img, err := s.texture(k.aux)
		if err != nil {
			return glb.VariantNode{}, err
		}
		n.TexGroups = append(n.TexGroups, glb.TexturedGroup{
			Tris: texTris[k], Image: img, Blend: k.blend,
			WrapS: 10497, WrapT: 10497,
		})
	}
	if len(greyTris) > 0 {
		n.ColorGroups = append(n.ColorGroups, glb.TriGroup{
			Tris: greyTris, Color: [3]float32{0.72, 0.72, 0.75},
		})
	}
	return n, nil
}

func (s *objectSet) writeAssembly(path, sceneName string, parts []placed) error {
	var nodes []glb.VariantNode
	for _, p := range parts {
		n, err := s.node(p)
		if err != nil {
			return fmt.Errorf("%s: %w", p.label, err)
		}
		nodes = append(nodes, n)
	}
	return glb.WriteVariantScenes(path, []glb.ModelVariant{{Name: sceneName, Nodes: nodes}})
}

// exportSky writes the shared sky GLB — the dome and the translucent
// horizon ring at their captured uniform draw scale of 2 — for every
// course's camera-attached sky layer.
func exportSky(ctx *cli.Context, s *objectSet) (string, error) {
	const file = "sky.glb"
	scale2 := [12]float32{2, 0, 0, 0, 0, 2, 0, 0, 0, 0, 2, 0}
	p, err := ctx.Builder.Path("levels", file)
	if err != nil {
		return "", err
	}
	if err := s.writeAssembly(p, "sky", []placed{
		{272, "sky dome", scale2},
		{258, "horizon", scale2},
	}); err != nil {
		return "", err
	}
	return file, nil
}

// skyLayer is the layer entry every course appends: camera-attached like
// the game's own draw, painted before the world and never depth-tested so
// the city always wins.
func skyLayer(file string) schema.Layer {
	no := false
	return schema.Layer{
		ID: "sky", Name: "Sky", File: file, Mode: "toggle",
		Attach: "camera", RenderOrder: -1, Role: "sky", DepthTest: &no,
	}
}

// exportObjects ships the object gallery: cabs, drivers, traffic.
func exportObjects(ctx *cli.Context, s *objectSet) error {
	b := ctx.Builder
	for i, c := range cabs {
		file := c.id + ".glb"
		p, err := b.Path("objects", file)
		if err != nil {
			return err
		}
		if err := s.writeAssembly(p, "cab", c.parts); err != nil {
			return fmt.Errorf("%s: %w", c.id, err)
		}
		b.AddObject(schema.Asset{ID: c.id, Name: c.name, Group: "Player cabs"}, &schema.Object{
			Type: schema.ObjectModel3D, Name: c.name, Model: file,
		})
		ctx.Progress("objects", i+1, len(cabs)+len(driverPoses)+len(traffic), c.name)
	}
	done := len(cabs)
	for _, key := range []string{"axel", "bdjoe", "gena", "gus"} {
		parts := driverPoses[key]
		id := "driver-" + key
		file := id + ".glb"
		p, err := b.Path("objects", file)
		if err != nil {
			return err
		}
		if err := s.writeAssembly(p, "driver", parts); err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
		name := driverNames[key]
		b.AddObject(schema.Asset{ID: id, Name: name, Group: "Drivers"}, &schema.Object{
			Type: schema.ObjectModel3D, Name: name, Model: file,
		})
		done++
		ctx.Progress("objects", done, len(cabs)+len(driverPoses)+len(traffic), name)
	}
	for _, tr := range traffic {
		file := tr.id + ".glb"
		p, err := b.Path("objects", file)
		if err != nil {
			return err
		}
		if err := s.writeAssembly(p, "vehicle", tr.parts); err != nil {
			return fmt.Errorf("%s: %w", tr.id, err)
		}
		b.AddObject(schema.Asset{ID: tr.id, Name: tr.name, Group: "Traffic"}, &schema.Object{
			Type: schema.ObjectModel3D, Name: tr.name, Model: file,
		})
		done++
		ctx.Progress("objects", done, len(cabs)+len(driverPoses)+len(traffic), tr.name)
	}
	return nil
}
