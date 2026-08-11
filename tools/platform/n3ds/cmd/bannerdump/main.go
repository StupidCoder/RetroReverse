// bannerdump decodes a 3DS title's HOME-Menu banner — the little animated 3-D
// scene shown when the game is highlighted — and lists what it contains. The
// banner is the ExeFS "banner" file: a CBMD container holding an LZ11-compressed
// CGFX scene (model, textures, materials, and camera/skeletal animation).
//
// Given a cartridge image it does the whole chain (NCCH → ExeFS → CBMD →
// LZ11 → CGFX); given an already-extracted banner or a raw .cgfx it starts
// partway in. With -o it writes the decompressed CGFX out for further work.
//
// Usage:
//
//	bannerdump game.cci                list the banner scene graph
//	bannerdump -o out.cgfx game.cci    also write the decompressed CGFX
//	bannerdump -banner bannerfile      start from an extracted ExeFS banner
//	bannerdump -cgfx model.cgfx        start from a decompressed CGFX
//	bannerdump -model game.cci         also decode the model: meshes, shapes and
//	                                   each material's texture mappers
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/tools/platform/n3ds"
)

func main() {
	out := flag.String("o", "", "write the decompressed CGFX to this file")
	fromBanner := flag.String("banner", "", "input is an extracted ExeFS banner (CBMD) file")
	fromCGFX := flag.String("cgfx", "", "input is a decompressed CGFX file")
	model := flag.Bool("model", false, "also decode the model: meshes, shapes and material texture mappers")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: bannerdump [-o out.cgfx] [-model] (game.cci | -banner FILE | -cgfx FILE)")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(flag.Arg(0), *fromBanner, *fromCGFX, *out, *model); err != nil {
		fmt.Fprintln(os.Stderr, "bannerdump:", err)
		os.Exit(1)
	}
}

func run(image, bannerPath, cgfxPath, out string, model bool) error {
	cgfx, err := loadCGFX(image, bannerPath, cgfxPath)
	if err != nil {
		return err
	}
	if out != "" {
		if err := os.WriteFile(out, cgfx, 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote decompressed CGFX (%d bytes) to %s\n", len(cgfx), out)
	}

	g, err := n3ds.ParseCGFX(cgfx)
	if err != nil {
		return err
	}
	fmt.Printf("CGFX  revision 0x%08x  fileSize %d bytes\n", g.Revision, g.FileSize)
	if g.IMAGOff > 0 {
		fmt.Printf("  IMAG (raw vertex/texture data) block at 0x%x\n", g.IMAGOff)
	}
	fmt.Println("  scene graph:")

	types := make([]string, 0, len(g.Resources))
	for t := range g.Resources {
		types = append(types, t)
	}
	sort.Strings(types)
	for _, t := range types {
		entries := g.Resources[t]
		fmt.Printf("    %-20s %d\n", t, len(entries))
		for _, e := range entries {
			fmt.Printf("        %-24s %-6s @0x%x\n", e.Name, e.Magic, e.Offset)
		}
	}
	if model {
		return dumpModels(g)
	}
	return nil
}

// dumpModels decodes every CMDL and prints what an exporter has to honour: the
// mesh→shape/material bindings, each shape's vertex sets, and each material's
// texture mappers with their source UV set, coordinate transform and wrap
// modes. A mapper whose scale is not 1 or whose wrap is not REPEAT will be
// silently wrong in any exporter that assumes the defaults, so print them.
func dumpModels(g *n3ds.CGFX) error {
	wrapName := map[n3ds.PICAWrap]string{
		n3ds.WrapClampToEdge: "clamp", n3ds.WrapClampToBorder: "border",
		n3ds.WrapRepeat: "repeat", n3ds.WrapMirroredRepeat: "mirror",
	}
	for _, e := range g.Resources["Models"] {
		m, err := g.DecodeModel(e)
		if err != nil {
			return err
		}
		fmt.Printf("  model %s: %d meshes, %d shapes, %d materials, %d bones\n",
			m.Name, len(m.Meshes), len(m.Shapes), len(m.Materials), len(m.Bones))
		for i, mat := range m.Materials {
			fmt.Printf("    material[%d] %s\n", i, mat.Name)
			for j, t := range mat.Mappers {
				fmt.Printf("      mapper[%d] %-8s uv%d  scale(%g,%g) rot %g trans(%g,%g)  wrap %s/%s\n",
					j, t.Texture, t.SourceUV, t.ScaleS, t.ScaleT, t.Rotate, t.TransS, t.TransT,
					wrapName[t.WrapS], wrapName[t.WrapT])
			}
		}
		for i, mesh := range m.Meshes {
			sh := &m.Shapes[mesh.ShapeIndex]
			bone := "-"
			if sh.BoneIndex >= 0 && sh.BoneIndex < len(m.Bones) {
				bone = m.Bones[sh.BoneIndex].Name
			}
			fmt.Printf("    mesh[%d] shape %d (%s): %d verts, %d tris, %d UV sets, normals=%v, colours=%v -> %s\n",
				i, mesh.ShapeIndex, bone, len(sh.Verts), len(sh.Indices)/3, sh.UVCount,
				sh.HasNormal, sh.HasColor, m.Materials[mesh.MaterialIndex].Name)
		}
	}
	return nil
}

// loadCGFX resolves the input to a decompressed CGFX blob, from whichever entry
// point the flags select.
func loadCGFX(image, bannerPath, cgfxPath string) ([]byte, error) {
	switch {
	case cgfxPath != "":
		return os.ReadFile(cgfxPath)
	case bannerPath != "":
		b, err := os.ReadFile(bannerPath)
		if err != nil {
			return nil, err
		}
		return cgfxFromBanner(b)
	case image != "":
		img, err := os.ReadFile(image)
		if err != nil {
			return nil, err
		}
		ncsd, err := n3ds.ParseNCSD(img)
		if err != nil {
			return nil, err
		}
		cxi, err := ncsd.Executable()
		if err != nil {
			return nil, err
		}
		efs, err := cxi.ExeFS()
		if err != nil {
			return nil, err
		}
		banner, err := efs.File("banner")
		if err != nil {
			return nil, err
		}
		return cgfxFromBanner(banner)
	}
	return nil, fmt.Errorf("no input: give a game.cci, -banner FILE, or -cgfx FILE")
}

func cgfxFromBanner(b []byte) ([]byte, error) {
	bn, err := n3ds.ParseBanner(b)
	if err != nil {
		return nil, err
	}
	return bn.CommonModel()
}
