// webexport extracts the game's web deliverables from the UMD image into
// site/public/loco-roco-psp: every stage's level geometry as textured GLBs
// (foreground terrain and background flora as separate models), the level
// JSONs and the manifest.
//
//	webexport -in games/loco-roco-psp/image/LocoRoco.cso [-o DIR] [-only levels]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/loco-roco-psp/extract/clv"
	"retroreverse.com/games/loco-roco-psp/extract/garc"
	"retroreverse.com/games/loco-roco-psp/extract/gprs"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/psp"
)

// dataBinLBN is DATA.BIN's start sector on this UMD (the game prints it at
// boot: "DATA.BIN : LBN[23472]").
const dataBinLBN = 23472

func main() {
	in := flag.String("in", "", "UMD image (.cso or .iso)")
	out := flag.String("o", "../../site/public/loco-roco-psp", "output root")
	only := flag.String("only", "all", "levels|all")
	flag.Parse()
	if *in == "" {
		die("need -in IMAGE")
	}
	if *only != "all" && *only != "levels" {
		die("unknown -only %q", *only)
	}

	im, err := psp.OpenImage(*in)
	if err != nil {
		die("%v", err)
	}
	defer im.Close()

	// the boot archive's GIMG directory locates every stage file in DATA.BIN
	raw, err := im.ReadFile("PSP_GAME/USRDIR/data/first_us.arc")
	if err != nil {
		die("first_us.arc: %v", err)
	}
	dec, err := gprs.Decompress(raw)
	if err != nil {
		die("first_us.arc: %v", err)
	}
	arc, err := garc.Parse(dec)
	if err != nil {
		die("first_us.arc: %v", err)
	}
	sect, ok := arc.Find("sector_usa.bin")
	if !ok {
		die("no sector_usa.bin in first_us.arc")
	}
	dir, err := garc.ParseGimg(arc.Data(sect))
	if err != nil {
		die("sector_usa.bin: %v", err)
	}

	var stages []garc.GimgEntry
	for _, e := range dir {
		if strings.HasPrefix(e.Name, "st_") && strings.HasSuffix(e.Name, ".clv") {
			stages = append(stages, e)
		}
	}
	sort.Slice(stages, func(i, j int) bool { return stages[i].Name < stages[j].Name })

	if err := os.MkdirAll(filepath.Join(*out, "levels"), 0755); err != nil {
		die("%v", err)
	}
	if err := os.MkdirAll(filepath.Join(*out, "models"), 0755); err != nil {
		die("%v", err)
	}

	type manifestLevel struct {
		Name    string `json:"name"`
		Section string `json:"section"`
		File    string `json:"file"`
		Kind    string `json:"kind"`
	}
	var levels []manifestLevel
	fail := 0
	for i, st := range stages {
		stem := strings.TrimSuffix(strings.TrimPrefix(st.Name, "st_"), ".clv")
		fmt.Fprintf(os.Stderr, "[levels] %2d/%d  %s\n", i+1, len(stages), st.Name)
		if err := exportStage(im, st, stem, *out); err != nil {
			fmt.Fprintf(os.Stderr, "[levels] %s FAILED: %v\n", st.Name, err)
			fail++
			continue
		}
		section := strings.TrimRight(stem, "0123456789")
		levels = append(levels, manifestLevel{
			Name: stem, Section: section,
			File: "levels/" + stem + ".json", Kind: "mesh3d",
		})
	}
	if fail > 0 {
		fmt.Fprintf(os.Stderr, "[levels] %d of %d stages FAILED\n", fail, len(stages))
	}

	manifest := map[string]any{
		"format":   2,
		"game":     "loco-roco-psp",
		"platform": "Sony PSP",
		"native":   map[string]int{"w": 480, "h": 272},
		"tickHz":   60,
		"levels":   levels,
	}
	mf, err := os.Create(filepath.Join(*out, "manifest.json"))
	if err != nil {
		die("%v", err)
	}
	enc := json.NewEncoder(mf)
	enc.SetIndent("", " ")
	if err := enc.Encode(manifest); err != nil {
		die("%v", err)
	}
	mf.Close()
	fmt.Fprintf(os.Stderr, "wrote %s (%d levels)\n", filepath.Join(*out, "manifest.json"), len(levels))
	if fail > 0 {
		os.Exit(1)
	}
}

// exportStage decodes one stage file and writes its level JSON and its
// foreground/background GLBs.
func exportStage(im *psp.Image, st garc.GimgEntry, stem, out string) error {
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
		for _, b := range c.Layout.Cells[i].Batches {
			m, err := c.Material(b)
			if err != nil {
				return fmt.Errorf("cell %d: %w", i, err)
			}
			dst := bg
			if strings.HasPrefix(m.Name, "stage") {
				dst = fg
			}
			if err := dst.addBatch(c, b, m); err != nil {
				return fmt.Errorf("cell %d %q: %w", i, m.Name, err)
			}
		}
	}

	lvl := map[string]any{
		"format": 2,
		"name":   stem,
		"kind":   "mesh3d",
		"extents": map[string]any{
			"min": []float32{c.Layout.X, c.Layout.Y, c.Layout.Z0},
			"max": []float32{c.Layout.X + c.Layout.W, c.Layout.Y + c.Layout.H, c.Layout.Z1},
		},
		"mesh": map[string]any{"glb": "models/" + stem + ".glb"},
	}
	if fg.dropped+bg.dropped > 0 {
		fmt.Fprintf(os.Stderr, "[levels] %s: dropped %d strips with non-finite vertices\n",
			stem, fg.dropped+bg.dropped)
	}
	if err := fg.write(filepath.Join(out, "models", stem+".glb")); err != nil {
		return err
	}
	if len(bg.positions) > 0 {
		if err := bg.write(filepath.Join(out, "models", stem+"_bg.glb")); err != nil {
			return err
		}
		lvl["sky"] = "models/" + stem + "_bg.glb"
	}
	lf, err := os.Create(filepath.Join(out, "levels", stem+".json"))
	if err != nil {
		return err
	}
	enc := json.NewEncoder(lf)
	enc.SetIndent("", " ")
	if err := enc.Encode(lvl); err != nil {
		return err
	}
	return lf.Close()
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
