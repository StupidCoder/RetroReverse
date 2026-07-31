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

// The four player cabs, each in two LODs the traces separated cleanly:
// the DETAIL body the attract mode draws (with front overhang and seat
// backs — the in-drive LOD omits what the chase camera never sees) and
// the in-drive body the four gameplay drives dispatched. Names are the
// game's own driver-select captions; wheel anchors are the captured
// dispatch transforms — each cab has its own track and wheelbase, Axel's
// and B.D.Joe's cabs share their wheel models at both LODs (the attract
// counts add exactly: 164+320=484), and the attract capture proved the
// detail wheels ride the same x/z anchors as the drive wheels. m14/m28/
// m42/m56 are the rear bumpers (thin slabs at the tail), m13 etc. the
// roof signs. The cab faces +z and the driver sits on the +x (left)
// side, read off the seated driver's captured position.
type cabSpec struct {
	id, name, driver string
	detail, drive    []placed
}

func cabParts(body, sign, rearBumper, fl, fr, rl, rr int, halfTrackF, yF, wheelbaseF, halfTrackR, yR, wheelbaseR float32) []placed {
	return []placed{
		{body, "body", ident},
		{sign, "roof sign", ident},
		{rearBumper, "rear bumper", ident},
		at(fl, "wheel front left", halfTrackF, yF, wheelbaseF),
		at(fr, "wheel front right", -halfTrackF, yF, wheelbaseF),
		at(rl, "wheel rear left", halfTrackR, yR, -wheelbaseR),
		at(rr, "wheel rear right", -halfTrackR, yR, -wheelbaseR),
	}
}

var cabs = []cabSpec{
	{"cab-axel", "Axel's cab", "axel",
		cabParts(2, 13, 14, 3, 4, 5, 6, 6.9, 3.188, 15.5, 6.5, 3.096, 15.5),
		cabParts(9, 13, 14, 7, 8, 10, 11, 6.9, 3.188, 15.5, 6.5, 3.096, 15.5)},
	{"cab-bdjoe", "B.D.Joe's cab", "bdjoe",
		cabParts(24, 27, 28, 3, 4, 5, 6, 6.9, 3.188, 15.25, 6.5, 3.097, 15.25),
		cabParts(25, 27, 28, 7, 8, 10, 11, 6.9, 3.188, 15.25, 6.5, 3.097, 15.25)},
	{"cab-gena", "Gena's cab", "gena",
		cabParts(44, 55, 56, 45, 46, 47, 48, 6.6, 3.188, 14.6, 6.9, 3.096, 14.6),
		cabParts(51, 55, 56, 49, 50, 52, 53, 6.6, 3.188, 14.6, 6.9, 3.096, 14.6)},
	{"cab-gus", "Gus's cab", "gus",
		cabParts(30, 41, 42, 31, 32, 33, 34, 6.8, 3.190, 14.5, 7.1, 3.095, 14.5),
		cabParts(37, 41, 42, 35, 36, 38, 39, 6.8, 3.190, 14.5, 7.1, 3.095, 14.5)},
}

var driverNames = map[string]string{
	"axel": "Axel", "bdjoe": "B.D.Joe", "gena": "Gena", "gus": "Gus",
}

// Curiosities: POLDC0 models the traces never (or only obliquely) touched,
// shipped because unused data is half the fun of an excavation. Names
// state what is KNOWN — a model no trace dispatched gets no invented
// role beyond what its own bytes show.
var curiosities = []struct {
	id, name, descr string
	parts           []placed
}{
	{"unused-15", "Rear bumper variant (model 15)",
		"byte-for-byte layout twin of Axel's drawn rear bumper m14, different texture pair; never dispatched in any trace", []placed{{15, "model", ident}}},
	{"unused-29", "Rear bumper variant (model 29)",
		"the same second-livery pattern beside B.D.Joe's drawn m28; never dispatched in any trace", []placed{{29, "model", ident}}},
	{"unused-12", "Roof sign variant (model 12)",
		"sibling of the drawn roof sign m13 with the neighbouring texture (aux 86 vs 87); never dispatched in any trace", []placed{{12, "model", ident}}},
	{"model-0", "Cable car (model 0)",
		"the container's first model: the street tram (eyeballed — the game's own destination table names its 'Cable car stop's); outside our short trace windows", []placed{{0, "model", ident}}},
	{"model-235", "Car transporter (model 235)",
		"the ramp-trailer truck (eyeballed), the biggest vehicle model in the container; outside our trace windows", []placed{{235, "model", ident}}},
	{"model-209", "Traffic mid model (model 209)",
		"a 76-vertex body sharing the m194 traffic family's textures — a middle LOD no drive of ours dispatched", []placed{{209, "model", ident}}},
	{"traffic-58", "Traffic variant head (model 58)",
		"a fifth traffic family's 42-vertex head (textures 22/23); its body m59 beside it", []placed{{58, "head", ident}, {59, "body", ident}}},
	{"glass-183", "Popped-glass state (model 183)",
		"the m184 traffic car's 56-vertex variant head: the glass panels in their raised pose", []placed{{183, "model", ident}}},
	{"glass-119", "Popped-glass state (model 119)",
		"the m120 traffic car's variant head", []placed{{119, "model", ident}}},
	{"glass-193", "Popped-glass state (model 193)",
		"the m194 traffic car's variant head", []placed{{193, "model", ident}}},
	{"glass-237", "Popped-glass state (model 237)",
		"the m238 traffic car's variant head", []placed{{237, "model", ident}}},
}

// The traffic fleet: driver-independent vehicles drawn in every drive.
// Each is a photo-quad impostor pair — a coarse textured body plus one
// two-quad wheel model the game places once per axle (captured anchors;
// the quads carry the wheel photograph and spin about x at speed). The
// 56-vertex variant heads (models 119/183/193/237, popped-glass states)
// and the mid model 209 ship in the Curiosities group.
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
	cache  map[string]image.Image
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
	return &objectSet{models: models, td: td, texdc: texdc, cache: map[string]image.Image{}}, nil
}

// texMode says how a group must ship, mirroring what the PVR would have
// done with the same bytes. The translucent list autosorts per pixel on
// the hardware, so the game parks whole impostor bodies there for free;
// glTF BLEND (no depth writes) must be reserved for surfaces that are
// actually translucent — by the block's own base alpha (the cab glass:
// opaque texture, base a=0.44) or by partial texture alpha (the traffic
// windows). Binary-alpha textures become MASK cutouts (the wheel photo
// quads), fully-opaque ones ship opaque (the impostor body panels).
type texMode int

const (
	// texSolid binarizes alpha at nonzero: exactly-zero texels stay holes,
	// everything else ships solid. This is the oracle's own pixel rule
	// (plot drops a==0 in every list and otherwise writes; the opaque list
	// never blends) — and it is also how the translucent-list impostors
	// READ in-game: their fractional-alpha panels composite against the
	// car's own dark interior shell drawn beneath, so the accumulated
	// result is solid. A fixed 48-style threshold cut the alpha-17/34
	// body paint right off the traffic sedans; binarizing keeps it.
	texSolid  texMode = iota
	texCutout         // keep raw alpha, MASK at the default 0.5 (punch-through)
	texBlend          // keep alpha, BLEND, base colour baked in
)

// rawTexture is the untouched decode of entry aux, cached.
func (s *objectSet) rawTexture(aux int) (*image.RGBA, error) {
	key := fmt.Sprintf("raw-%d", aux)
	if img, ok := s.cache[key]; ok {
		return img.(*image.RGBA), nil
	}
	img, err := s.td.Decode(aux, s.texdc)
	if err != nil {
		return nil, fmt.Errorf("texture %d: %w", aux, err)
	}
	s.cache[key] = img
	return img, nil
}


// texture decodes entry aux for one shipping mode. texSolid binarizes
// the alpha at nonzero; texBlend multiplies the block's base colour in
// (mul, 0-255 per channel); texCutout keeps the raw decode.
func (s *objectSet) texture(aux int, mode texMode, mul [4]uint8) (image.Image, error) {
	key := fmt.Sprintf("%d-%d-%v", aux, mode, mul)
	if img, ok := s.cache[key]; ok {
		return img, nil
	}
	img, err := s.rawTexture(aux)
	if err != nil {
		return nil, err
	}
	var out image.Image = img
	switch {
	case mode == texSolid:
		op := image.NewRGBA(img.Bounds())
		for i := 0; i < len(img.Pix); i += 4 {
			a := uint8(0)
			if img.Pix[i+3] != 0 {
				a = 255
			}
			op.Pix[i], op.Pix[i+1], op.Pix[i+2], op.Pix[i+3] = img.Pix[i], img.Pix[i+1], img.Pix[i+2], a
		}
		out = op
	case mode == texBlend && mul != [4]uint8{255, 255, 255, 255}:
		op := image.NewRGBA(img.Bounds())
		for i := 0; i < len(img.Pix); i += 4 {
			for c := 0; c < 4; c++ {
				op.Pix[i+c] = uint8(uint32(img.Pix[i+c]) * uint32(mul[c]) / 255)
			}
		}
		out = op
	}
	s.cache[key] = out
	return out, nil
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
		aux          int
		mode         texMode
		mul          [4]uint8
		wrapS, wrapT int
	}
	texTris := map[key][][3]uint32{}
	var order []key
	var greyTris [][3]uint32
	xf := p.m
	for _, blk := range m.Blocks {
		// Ship each block the way it reads on the hardware: solid with
		// exact zero-alpha holes everywhere (the oracle's own pixel rule),
		// except punch-through's real cutout compare and the blocks whose
		// BASE alpha is fractional — the only surfaces that genuinely
		// blend (the cab windshield).
		mode := texSolid
		mul := [4]uint8{255, 255, 255, 255}
		switch blk.ListType() {
		case 4:
			mode = texCutout
		case 2:
			if baseA := blk.BaseColor[3]; baseA < 0.98 {
				mode = texBlend
				for c := 0; c < 3; c++ {
					v := blk.BaseColor[c]
					if v > 1 {
						v = 1
					}
					mul[c] = uint8(v * 255)
				}
				mul[3] = uint8(baseA * 255)
			}
		}
		// TSP bits 18/17 select the PVR's mirrored repeat per axis, pinned
		// by a control pair on the cab itself: the chrome strip tiling U
		// 0..60 (bit clear) against the steering-wheel quarter spanning
		// U 0..2 (bit set), same texture size, one differing bit.
		wrapS, wrapT := 10497, 10497
		if blk.TSP>>18&1 != 0 {
			wrapS = 33648
		}
		if blk.TSP>>17&1 != 0 {
			wrapT = 33648
		}
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
				// The PVR does not cull, so the strips' winding is free to
				// disagree with the stored vertex normals (and does, on most
				// of the cab). A double-sided lighting pass flips its shading
				// normal by winding, so reorder each triangle to agree with
				// the game's own lighting inputs.
				p0, p1, p2 := n.Positions[tri[0]], n.Positions[tri[1]], n.Positions[tri[2]]
				e1 := [3]float32{p1[0] - p0[0], p1[1] - p0[1], p1[2] - p0[2]}
				e2 := [3]float32{p2[0] - p0[0], p2[1] - p0[1], p2[2] - p0[2]}
				fn := [3]float32{e1[1]*e2[2] - e1[2]*e2[1], e1[2]*e2[0] - e1[0]*e2[2], e1[0]*e2[1] - e1[1]*e2[0]}
				vn := n.Normals[tri[0]]
				if fn[0]*vn[0]+fn[1]*vn[1]+fn[2]*vn[2] < 0 {
					tri[1], tri[2] = tri[2], tri[1]
				}
				if blk.Textured() && blk.Aux != 0xFFFFFFFF && int(blk.Aux) < len(s.td.Entries) {
					k := key{int(blk.Aux), mode, mul, wrapS, wrapT}
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
		img, err := s.texture(k.aux, k.mode, k.mul)
		if err != nil {
			return glb.VariantNode{}, err
		}
		n.TexGroups = append(n.TexGroups, glb.TexturedGroup{
			Tris: texTris[k], Image: img, Blend: k.mode == texBlend,
			WrapS: k.wrapS, WrapT: k.wrapT,
		})
	}
	if len(greyTris) > 0 {
		n.ColorGroups = append(n.ColorGroups, glb.TriGroup{
			Tris: greyTris, Color: [3]float32{0.72, 0.72, 0.75},
		})
	}
	return n, nil
}

func (s *objectSet) variant(sceneName string, parts []placed) (glb.ModelVariant, error) {
	var nodes []glb.VariantNode
	for _, p := range parts {
		n, err := s.node(p)
		if err != nil {
			return glb.ModelVariant{}, fmt.Errorf("%s: %w", p.label, err)
		}
		nodes = append(nodes, n)
	}
	return glb.ModelVariant{Name: sceneName, Nodes: nodes}, nil
}

func (s *objectSet) writeAssembly(path, sceneName string, parts []placed) error {
	v, err := s.variant(sceneName, parts)
	if err != nil {
		return err
	}
	return glb.WriteVariantScenes(path, []glb.ModelVariant{v})
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
	total := len(cabs) + len(driverPoses) + len(traffic) + len(curiosities)
	for i, c := range cabs {
		file := c.id + ".glb"
		p, err := b.Path("objects", file)
		if err != nil {
			return err
		}
		det, err := s.variant("detail", c.detail)
		if err != nil {
			return fmt.Errorf("%s detail: %w", c.id, err)
		}
		drv, err := s.variant("in-drive", c.drive)
		if err != nil {
			return fmt.Errorf("%s drive: %w", c.id, err)
		}
		if err := glb.WriteVariantScenes(p, []glb.ModelVariant{det, drv}); err != nil {
			return fmt.Errorf("%s: %w", c.id, err)
		}
		b.AddObject(schema.Asset{ID: c.id, Name: c.name, Group: "Player cabs"}, &schema.Object{
			Type: schema.ObjectModel3D, Name: c.name, Model: file,
			Variants: []schema.ModelVariant{
				{ID: "detail", Name: "Detail (attract)", Scene: "detail",
					Description: "the close-up body the attract mode draws"},
				{ID: "in-drive", Name: "In-drive", Scene: "in-drive",
					Description: "the chase-camera body the gameplay drives dispatch"},
			},
		})
		ctx.Progress("objects", i+1, total, c.name)
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
		ctx.Progress("objects", done, total, name)
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
		ctx.Progress("objects", done, total, tr.name)
	}
	for _, cu := range curiosities {
		file := cu.id + ".glb"
		p, err := b.Path("objects", file)
		if err != nil {
			return err
		}
		if err := s.writeAssembly(p, "model", cu.parts); err != nil {
			return fmt.Errorf("%s: %w", cu.id, err)
		}
		b.AddObject(schema.Asset{ID: cu.id, Name: cu.name, Group: "Curiosities"}, &schema.Object{
			Type: schema.ObjectModel3D, Name: cu.name, Model: file,
			Props: map[string]any{"note": cu.descr},
		})
		done++
		ctx.Progress("objects", done, total, cu.name)
	}
	return nil
}
