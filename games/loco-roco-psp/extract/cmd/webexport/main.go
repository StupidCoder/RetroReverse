// webexport extracts the game's web deliverables from the UMD image into
// site/public/loco-roco-psp: every stage's level geometry as textured GLBs
// (foreground terrain and background flora as separate models), the level
// JSONs and the manifest.
//
//	webexport -in games/loco-roco-psp/image/LocoRoco.cso [-o DIR] [-only levels]
package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/loco-roco-psp/extract/clv"
	"retroreverse.com/games/loco-roco-psp/extract/garc"
	"retroreverse.com/games/loco-roco-psp/extract/gprs"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/psp"
)

// dataBinLBN is DATA.BIN's start sector on this UMD (the game prints it at
// boot: "DATA.BIN : LBN[23472]").
const dataBinLBN = 23472

func main() {
	cli.Main("loco-roco-psp", runCLI)
}

func runCLI(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in UMD.iso/.cso [-o DIR]")
	}
	b := ctx.Builder
	b.SetTitle("LocoRoco")
	b.SetPlatform("Sony PSP")
	b.SetYear(2006)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 480, H: 272},
		TickHz: 60,
		// the PSP's backlit TFT + the GE's bilinear sampling
		Filter:    "ds",
		TexFilter: "linear",
	})
	ctx.Stage("levels")

	im, err := psp.OpenImage(ctx.In)
	if err != nil {
		return err
	}
	defer im.Close()

	// the boot archive's GIMG directory locates every stage file in DATA.BIN
	raw, err := im.ReadFile("PSP_GAME/USRDIR/data/first_us.arc")
	if err != nil {
		return fmt.Errorf("first_us.arc: %w", err)
	}
	dec, err := gprs.Decompress(raw)
	if err != nil {
		return fmt.Errorf("first_us.arc: %w", err)
	}
	arc, err := garc.Parse(dec)
	if err != nil {
		return fmt.Errorf("first_us.arc: %w", err)
	}
	sect, ok := arc.Find("sector_usa.bin")
	if !ok {
		return fmt.Errorf("no sector_usa.bin in first_us.arc")
	}
	dir, err := garc.ParseGimg(arc.Data(sect))
	if err != nil {
		return fmt.Errorf("sector_usa.bin: %w", err)
	}

	var stages []garc.GimgEntry
	for _, e := range dir {
		if strings.HasPrefix(e.Name, "st_") && strings.HasSuffix(e.Name, ".clv") {
			stages = append(stages, e)
		}
	}
	sort.Slice(stages, func(i, j int) bool { return stages[i].Name < stages[j].Name })

	fail := 0
	for i, st := range stages {
		stem := strings.TrimSuffix(strings.TrimPrefix(st.Name, "st_"), ".clv")
		if err := exportStage(ctx, im, st, stem); err != nil {
			ctx.Logf("%s FAILED: %v", st.Name, err)
			fail++
			continue
		}
		ctx.Progress("levels", i+1, len(stages), st.Name)
	}
	if fail > 0 {
		return fmt.Errorf("%d of %d stages failed", fail, len(stages))
	}
	return nil
}

// exportStage decodes one stage file and writes its level JSON and its
// foreground/background GLBs.
func exportStage(ctx *cli.Context, im *psp.Image, st garc.GimgEntry, stem string) error {
	b := ctx.Builder
	raw, err := im.Volume.ReadFile(fmt.Sprintf("sce_lbn0x%X_size0x%X", dataBinLBN+st.Sector, st.Size))
	if err != nil {
		return err
	}
	c, err := clv.Parse(raw)
	if err != nil {
		return err
	}

	fg := newMeshBuilder()
	bg := newMeshBuilder()
	for i := range c.Layout.Cells {
		for _, bt := range c.Layout.Cells[i].Batches {
			m, err := c.Material(bt)
			if err != nil {
				return fmt.Errorf("cell %d: %w", i, err)
			}
			dst := bg
			if strings.HasPrefix(m.Name, "stage") {
				dst = fg
			}
			if err := dst.addBatch(c, bt, m); err != nil {
				return fmt.Errorf("cell %d %q: %w", i, m.Name, err)
			}
		}
	}
	if fg.dropped+bg.dropped > 0 {
		ctx.Logf("%s: dropped %d strips with non-finite vertices", stem, fg.dropped+bg.dropped)
	}

	p, err := b.Path("levels", stem+".glb")
	if err != nil {
		return err
	}
	if err := fg.write(p); err != nil {
		return err
	}
	doc := &schema.Level{
		Type: schema.LevelScene3D,
		// The engine clears each frame to a per-stage sky colour written into
		// BSS at stage load; its per-stage source is still unlocated (see the
		// writeup), so clear white — correct for the flower stages, an open
		// item for the dark ones.
		Scene: &schema.Scene{
			Background: "#ffffff",
			Layers:     []schema.Layer{{ID: "stage", File: stem + ".glb"}},
		},
	}
	if len(bg.positions) > 0 {
		p, err := b.Path("levels", stem+"_bg.glb")
		if err != nil {
			return err
		}
		if err := bg.write(p); err != nil {
			return err
		}
		doc.Scene.Layers = append(doc.Scene.Layers, schema.Layer{
			ID: "background", Name: "Background", File: stem + "_bg.glb", Mode: "toggle", Role: "sky",
		})
	}
	if len(c.Collision.Points) > 0 {
		// the collision contours as a line GLB the viewer toggles, one colour
		// per layer, drawn just in front of the terrain plane
		pos := make([][3]float32, len(c.Collision.Points))
		for i, pt := range c.Collision.Points {
			pos[i] = [3]float32{pt[0], pt[1], 150}
		}
		cols := [][3]float32{{0.9, 0.1, 0.1}, {0.1, 0.6, 0.1}, {0.15, 0.25, 0.9}, {0.8, 0.1, 0.8}}
		var lines []glb.LineGroup
		for i, cl := range c.Collision.Layers {
			g := glb.LineGroup{Color: cols[i%len(cols)]}
			for _, e := range cl.Edges {
				g.Lines = append(g.Lines, [2]uint32{uint32(e[0]), uint32(e[1])})
			}
			lines = append(lines, g)
		}
		p, err := b.Path("levels", stem+"_collision.glb")
		if err != nil {
			return err
		}
		if err := glb.WriteMixed(p, pos, nil, lines); err != nil {
			return err
		}
		off := false
		doc.Scene.Layers = append(doc.Scene.Layers, schema.Layer{
			ID: "collision", Name: "Collision contours", File: stem + "_collision.glb",
			Mode: "toggle", Visible: &off, Role: "collision",
		})

	}
	if len(fg.positions) > 0 {
		// the opening view: the stages run left to right, so frame the
		// terrain's own left edge — centre on the foreground geometry within
		// the first 800 units and stand back far enough to take it in
		minX := fg.positions[0][0]
		for _, p := range fg.positions {
			if p[0] < minX {
				minX = p[0]
			}
		}
		loY, hiY := float32(math.Inf(1)), float32(math.Inf(-1))
		for _, p := range fg.positions {
			if p[0] > minX+800 {
				continue
			}
			if p[1] < loY {
				loY = p[1]
			}
			if p[1] > hiY {
				hiY = p[1]
			}
		}
		cx, cy := float64(minX)+400, float64(loY+hiY)/2
		dist := 1.6 * float64(hiY-loY)
		if dist < 430 {
			dist = 430
		}
		// The stages are 3-D polygons but the game is 2-D: the camera faces
		// the plane and never rotates, so the viewer gets pan/zoom controls.
		doc.Camera = &schema.Camera{
			Mode: "pan2d", FOV: 36, Near: 1, Far: 20000,
			Pos:    []float64{cx, cy, dist},
			Target: []float64{cx, cy, 0},
		}
	}

	section := strings.TrimRight(stem, "0123456789")
	b.AddLevel(schema.Asset{
		ID: stem, Name: stem,
		Group: strings.ToUpper(section[:1]) + section[1:],
	}, doc)
	return nil
}

// meshBuilder accumulates triangles grouped per (texture, tint) material;
// untextured materials collect into flat-colour groups.
type meshBuilder struct {
	positions [][3]float32
	uvs       [][2]float32
	groups    map[string]*glb.TexturedGroup
	order     []string
	flat      map[uint32]*glb.TriGroup
	flatOrder []uint32
	dropped   int // strips skipped for non-finite vertex data
}

func isFinite(f float32) bool { return f == f && f < 1e30 && f > -1e30 }

func newMeshBuilder() *meshBuilder {
	return &meshBuilder{
		groups: map[string]*glb.TexturedGroup{},
		flat:   map[uint32]*glb.TriGroup{},
	}
}

// addBatch converts a batch's strips to triangles under the batch's material
// (its texture tinted by the material colour, as the GE modulates; colour
// alone when the material has no texture).
func (mb *meshBuilder) addBatch(c *clv.Clv, b clv.Batch, m clv.Material) error {
	var tris *[][3]uint32
	if m.TexName == "" {
		g := mb.flat[m.Color]
		if g == nil {
			g = &glb.TriGroup{Color: [3]float32{
				float32(m.Color&0xFF) / 255,
				float32(m.Color>>8&0xFF) / 255,
				float32(m.Color>>16&0xFF) / 255,
			}}
			mb.flat[m.Color] = g
			mb.flatOrder = append(mb.flatOrder, m.Color)
		}
		tris = &g.Tris
	} else {
		key := fmt.Sprintf("%s|%08X", m.TexName, m.Color)
		g := mb.groups[key]
		if g == nil {
			img, err := c.DecodeTexture(m)
			if err != nil {
				return err
			}
			tint(img, m.Color)
			g = &glb.TexturedGroup{Image: img, WrapS: 10497, WrapT: 10497,
				Blend: m.TexFmt == 3} // format 3 = the translucent texture class
			mb.groups[key] = g
			mb.order = append(mb.order, key)
		}
		tris = &g.Tris
	}
	for _, s := range b.Strips {
		finite := true
		for _, v := range s.Verts {
			if !isFinite(v.X) || !isFinite(v.Y) || !isFinite(v.Z) {
				finite = false
			}
		}
		if !finite {
			mb.dropped++
			continue
		}
		base := uint32(len(mb.positions))
		for _, v := range s.Verts {
			mb.positions = append(mb.positions, [3]float32{v.X, v.Y, v.Z})
			mb.uvs = append(mb.uvs, [2]float32{
				float32(v.U) / 32768 * m.UScale,
				float32(v.V) / 32768 * m.VScale,
			})
		}
		for i := 0; i+2 < len(s.Verts); i++ {
			a, bb, cc := base+uint32(i), base+uint32(i)+1, base+uint32(i)+2
			if i&1 == 1 {
				bb, cc = cc, bb
			}
			*tris = append(*tris, [3]uint32{a, bb, cc})
		}
	}
	return nil
}

func (mb *meshBuilder) write(path string) error {
	groups := make([]glb.TexturedGroup, 0, len(mb.order))
	for _, k := range mb.order {
		groups = append(groups, *mb.groups[k])
	}
	flat := make([]glb.TriGroup, 0, len(mb.flatOrder))
	for _, k := range mb.flatOrder {
		flat = append(flat, *mb.flat[k])
	}
	return glb.WriteTextured(path, mb.positions, mb.uvs, groups, flat)
}

// tint multiplies a texture by the material colour, the way the GE modulates
// texel by material.
func tint(img *image.RGBA, color uint32) {
	if color == 0xFFFFFFFF {
		return
	}
	r, g, b, a := uint32(color)&0xFF, (color>>8)&0xFF, (color>>16)&0xFF, (color>>24)&0xFF
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = byte(uint32(img.Pix[i]) * r / 255)
		img.Pix[i+1] = byte(uint32(img.Pix[i+1]) * g / 255)
		img.Pix[i+2] = byte(uint32(img.Pix[i+2]) * b / 255)
		img.Pix[i+3] = byte(uint32(img.Pix[i+3]) * a / 255)
	}
}

func die(f string, args ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", args...)
	os.Exit(1)
}
