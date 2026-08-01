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
	"os"
	"path/filepath"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/iso9660"
)

func main() {
	image := flag.String("image", "image/Jak and Daxter - The Precursor Legacy.iso", "disc image")
	symtab := flag.String("symtab", "work/goal.txt", "GOAL symbol table dump")
	site := flag.String("site", "../../site/public/jak-and-daxter-ps2", "output tree")
	flag.Parse()

	f, err := os.Open(*image)
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
	scene := glb.NewScene()
	gold := [4]float32{0.83, 0.68, 0.28, 1}
	for ci, off := range []uint32{0x1244, 0x29CE4} {
		c, err := merc.Parse(obj, off)
		check(err)
		node := scene.AddNode(fmt.Sprintf("logo-%d", ci), -1,
			[3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
		dec := &merc.Decoder{Micro: micro, LowMem: merc.DefaultLowMem(), STMagic: merc.CtrlSTMagic(obj, off), CtrlRow: obj[off+28 : off+44]}
		var prims []glb.Prim
		for i := range c.Effects {
			p, err := dec.BuildPrim(&c.Effects[i], gold)
			check(err)
			if len(p.Tris) > 0 {
				prims = append(prims, p)
			}
		}
		check(scene.AddMesh(node, fmt.Sprintf("logo-%d", ci), prims))
	}
	check(scene.Write(filepath.Join(*site, "objects", "title-logo.glb"), "title-logo"))

	writeJSON(filepath.Join(*site, "objects", "title-logo.json"), map[string]any{
		"format": "retro-x", "version": 1, "type": "model3d",
		"name": "Title logo (geometry WIP)", "model": "title-logo.glb",
	})
	writeJSON(filepath.Join(*site, "manifest.json"), map[string]any{
		"format": "retro-x", "version": 1,
		"id": "jak-and-daxter-ps2", "title": "Jak & Daxter: The Precursor Legacy",
		"platform": "Sony PlayStation 2", "year": 2001,
		"description": "Naughty Dog's PlayStation 2 debut: a seamless world written in the studio's own Lisp (GOAL), streamed level by level from DGO archives the engine links into place at load time. This export begins where the game does — the animated title logo, decoded from the disc alone: the archive container and the engine's runtime linker were reimplemented byte-exact, and the logo's skinned meshes are read straight out of the merc renderer's fragment format — fragment-local 8-bit vertex lattices, per-fragment origins hidden in float bit patterns, and triangle strips reassembled from the VU1 microprogram's own output-slot scatter. The triangle-strip reconstruction is still being verified against the renderer's own microprogram — the mesh is recognizably the logo but not yet watertight. Textures and the intro cutscenes are on their way.",
		"assets": []any{
			map[string]any{"id": "title-logo", "category": "object",
				"name": "Title logo (geometry WIP)", "group": "Title screen",
				"file": "objects/title-logo.json"},
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
