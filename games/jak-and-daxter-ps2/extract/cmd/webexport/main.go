package main

// webexport writes the Studio (retro-x) asset tree for Jak & Daxter.
// Run from games/jak-and-daxter-ps2/: paths default relative to it.
//
// Current contents: the title-screen logo art group as a GLB model
// (geometry, normals; materials pending Part IV's adgif decode).
import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/tpage"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/iso9660"
)

// dgoEntry returns the named entry's raw data from a parsed archive.
func dgoEntry(d *goalobj.DGO, name string) []byte {
	for _, e := range d.Entries {
		if e.Name == name {
			return e.Data
		}
	}
	return nil
}

func main() {
	imageF := flag.String("image", "image/Jak and Daxter - The Precursor Legacy.iso", "disc image")
	symtab := flag.String("symtab", "work/goal.txt", "GOAL symbol table dump")
	site := flag.String("site", "../../site/public/jak-and-daxter-ps2", "output tree")
	flag.Parse()

	f, err := os.Open(*imageF)
	check(err)
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	tab, err := goalobj.LoadSymTab(*symtab)
	check(err)

	check(os.MkdirAll(filepath.Join(*site, "objects"), 0o755))

	// --- the title logo (TIT.DGO "logo": two merc-ctrls) ---
	data, err := vol.ReadFile("DGO/TIT.DGO;1")
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	var raw []byte
	for _, e := range d.Entries {
		if e.Name == "logo" {
			raw = e.Data
			break
		}
	}
	obj, _, err := goalobj.Link(raw, 0, tab)
	check(err)
	engData, err := vol.ReadFile("CGO/ENGINE.CGO;1")
	check(err)
	eng, err := goalobj.ReadDGO(engData)
	check(err)
	var raws [][]byte
	for _, e := range eng.Entries {
		raws = append(raws, e.Data)
	}
	micro := merc.FindMicro(raws)
	if micro == nil {
		fmt.Fprintln(os.Stderr, "merc microcode not found in ENGINE.CGO")
		os.Exit(1)
	}
	// texture resolution: the level's remap table (title-vis) + its tpages
	visRaw := dgoEntry(d, "title-vis")
	vis, _, err := goalobj.Link(visRaw, 0, tab)
	check(err)
	remap := merc.LoadRemapTable(vis)
	pages := map[int]*tpage.Page{}
	for _, name := range []string{"tpage-415", "tpage-416"} {
		praw := dgoEntry(d, name)
		if praw == nil {
			continue
		}
		pobj, _, err := goalobj.Link(praw, 0, tab)
		check(err)
		pg, err := tpage.Load(pobj)
		check(err)
		pages[pg.ID] = pg
	}
	type texEntry struct {
		img   image.Image
		blend bool
	}
	texCache := map[uint32]texEntry{}
	resolve := func(s merc.ShaderRef) merc.Material {
		id := remap.Lookup(s.RawID)
		if e, ok := texCache[id]; ok {
			return merc.Material{Image: e.img, Blend: e.blend}
		}
		pg := pages[int(id>>20)]
		idx := int(id >> 8 & 0xFFF)
		if pg == nil || idx >= len(pg.Textures) {
			fmt.Fprintf(os.Stderr, "webexport: no texture for id %08x (raw %08x)\n", id, s.RawID)
			texCache[id] = texEntry{}
			return merc.Material{}
		}
		img, err := pg.Decode(&pg.Textures[idx], 0)
		check(err)
		blend := false
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y && !blend; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if _, _, _, a := img.At(x, y).RGBA(); a < 0xFA00 {
					blend = true
					break
				}
			}
		}
		texCache[id] = texEntry{img, blend}
		return merc.Material{Image: img, Blend: blend}
	}

	// The two merc-ctrls are the English and Japanese logo variants — each
	// ships as its own asset.
	variants := []struct {
		off   uint32
		file  string
		title string
	}{
		{0x1244, "title-logo", "Title logo"},
		{0x29CE4, "title-logo-jp", "Title logo (Japan)"},
	}
	for _, v := range variants {
		c, err := merc.Parse(obj, v.off)
		check(err)
		scene := glb.NewScene()
		node := scene.AddNode(v.file, -1,
			[3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
		var prims []glb.Prim
		for i := range c.Effects {
			prims = append(prims, merc.TexturedPrims(&c.Effects[i], resolve)...)
		}
		check(scene.AddMesh(node, v.file, prims))
		check(scene.Write(filepath.Join(*site, "objects", v.file+".glb"), v.file))
		writeJSON(filepath.Join(*site, "objects", v.file+".json"), map[string]any{
			"format": "retro-x", "version": 1, "type": "model3d",
			"name": v.title, "model": v.file + ".glb",
		})
	}
	writeJSON(filepath.Join(*site, "manifest.json"), map[string]any{
		"format": "retro-x", "version": 1,
		"id": "jak-and-daxter-ps2", "title": "Jak & Daxter: The Precursor Legacy",
		"platform": "Sony PlayStation 2", "year": 2001,
		"description": "Naughty Dog's PlayStation 2 debut: a seamless world written in the studio's own Lisp (GOAL), streamed level by level from DGO archives the engine links into place at load time. This export begins where the game does — the animated title logo, decoded from the disc alone: the archive container and the engine's runtime linker were reimplemented byte-exact, and the logo's skinned meshes are read straight out of the merc renderer's fragment format — fragment-local 8-bit vertex lattices, per-fragment origins hidden in float bit patterns, and triangle strips rebuilt from the file bytes alone (dest-byte scatter order, per-write ADC flags, the byte-header's stitch-copy tables), verified triangle-for-triangle against the microprogram's own emulated output: all 17,142 triangles match. The materials are equally the disc's own: each fragment's adgif template names its texture by id, the level's remap table resolves it to a texture-page slot, and the tpage decoder supplies the pixels — the fiery letter gradient is the game's per-vertex colors over its 4x4 gradient texture, and the model turns out to be two logos, English and Japanese. The intro cutscenes are on their way.",
		"assets": []any{
			map[string]any{"id": "title-logo", "category": "object",
				"name": "Title logo", "group": "Title screen",
				"file": "objects/title-logo.json"},
			map[string]any{"id": "title-logo-jp", "category": "object",
				"name": "Title logo (Japan)", "group": "Title screen",
				"file": "objects/title-logo-jp.json"},
		},
	})
	fmt.Println("webexport: wrote", *site)
}

func writeJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", " ")
	check(err)
	check(os.WriteFile(path, append(b, '\n'), 0o644))
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}
