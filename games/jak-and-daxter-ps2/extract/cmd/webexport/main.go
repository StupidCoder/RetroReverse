package main

// webexport writes the Studio (retro-x) asset tree for Jak & Daxter.
// Run from games/jak-and-daxter-ps2/: paths default relative to it.
//
// Each art group exports through the same disc-only pipeline: the GOAL
// linker (goalobj) links the object at base 0, the linker's relocation
// report locates every merc-ctrl by its type word, merc rebuilds geometry,
// topology and materials from the fragment streams, the level vis object's
// remap table resolves texture ids, and the DGO's tpage objects supply the
// pixels.
import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/tpage"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/iso9660"
)

// level is one DGO's shared texture context.
type level struct {
	remap merc.RemapTable
	pages map[int]*tpage.Page
	cache map[uint32]texEntry
}

type texEntry struct {
	img   image.Image
	blend bool
}

func (l *level) resolve(s merc.ShaderRef) merc.Material {
	id := l.remap.Lookup(s.RawID)
	if e, ok := l.cache[id]; ok {
		return merc.Material{Image: e.img, Blend: e.blend}
	}
	pg := l.pages[int(id>>20)]
	idx := int(id >> 8 & 0xFFF)
	if pg == nil || idx >= len(pg.Textures) || pg.Textures[idx].W == 0 {
		fmt.Fprintf(os.Stderr, "webexport: no texture for id %08x (raw %08x)\n", id, s.RawID)
		l.cache[id] = texEntry{}
		return merc.Material{}
	}
	img, err := pg.Decode(&pg.Textures[idx], 0)
	check(err)
	out, blend := gsAlphaPolicy(img)
	l.cache[id] = texEntry{out, blend}
	return merc.Material{Image: out, Blend: blend}
}

// gsAlphaPolicy converts a texture's GS alpha channel (0x80 = 1.0) to GLB
// semantics. The standard merc path draws with ABE off (the chain's GIF
// tag PRIM field is 0x3C — blending disabled), so most alpha channels are
// masks/data, not coverage: a channel that is all-opaque-ish, or that
// would erase nearly the whole texture, exports opaque. Channels with a
// real transparent population export as cutout (MASK) or, when they carry
// a genuine gradient, as BLEND (the alpha-bucket effects: banners, soft
// wisps).
func gsAlphaPolicy(src image.Image) (image.Image, bool) {
	b := src.Bounds()
	out := image.NewRGBA(b)
	total, zeros, mids := 0, 0, 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			a8 := int(a >> 8)
			a2 := a8 * 2 // GS 0x80 = 1.0
			if a2 > 255 {
				a2 = 255
			}
			total++
			if a2 < 16 {
				zeros++
			} else if a2 < 240 {
				mids++
			}
			out.SetRGBA(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), uint8(a2)})
		}
	}
	zf := float64(zeros) / float64(total)
	mf := float64(mids) / float64(total)
	if zf < 0.01 || zf > 0.9 {
		// no real transparent population (or the channel would erase the
		// texture — it is data, not coverage): opaque
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := out.RGBAAt(x, y)
				c.A = 255
				out.SetRGBA(x, y, c)
			}
		}
		return out, false
	}
	return out, mf > 0.05 // gradient -> BLEND, hard cutout -> MASK
}

// dgoEntry returns the named v4 (data) entry — archives may also carry a v3
// code object under the same name (INT.DGO's evilbro does).
func u32of(b []byte, o uint32) uint32 {
	return uint32(b[o]) | uint32(b[o+1])<<8 | uint32(b[o+2])<<16 | uint32(b[o+3])<<24
}

func float32frombits(v uint32) float32 { return math.Float32frombits(v) }

func dgoEntry(d *goalobj.DGO, name string) []byte {
	for _, e := range d.Entries {
		if e.Name == name && len(e.Data) >= 12 &&
			uint32(e.Data[8])|uint32(e.Data[9])<<8|uint32(e.Data[10])<<16|uint32(e.Data[11])<<24 >= 4 {
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

	// loadLevel links a DGO's vis object (remap table) and tpages.
	levels := map[string]*level{}
	dgos := map[string]*goalobj.DGO{}
	loadDGO := func(path string) *goalobj.DGO {
		if d, ok := dgos[path]; ok {
			return d
		}
		data, err := vol.ReadFile(path + ";1")
		check(err)
		d, err := goalobj.ReadDGO(data)
		check(err)
		dgos[path] = d
		return d
	}
	addPages := func(l *level, d *goalobj.DGO) {
		for _, e := range d.Entries {
			if !strings.HasPrefix(e.Name, "tpage-") {
				continue
			}
			pobj, _, err := goalobj.Link(e.Data, 0, tab)
			check(err)
			pg, err := tpage.Load(pobj)
			check(err)
			if _, ok := l.pages[pg.ID]; !ok {
				l.pages[pg.ID] = pg
			}
		}
	}
	loadLevel := func(dgoPath, visName string) *level {
		key := dgoPath + "/" + visName
		if l, ok := levels[key]; ok {
			return l
		}
		d := loadDGO(dgoPath)
		vis, _, err := goalobj.Link(dgoEntry(d, visName), 0, tab)
		check(err)
		l := &level{
			remap: merc.LoadRemapTable(vis),
			pages: map[int]*tpage.Page{},
			cache: map[uint32]texEntry{},
		}
		addPages(l, d)
		// the always-resident common art pages (tpage-463 & co.)
		addPages(l, loadDGO("CGO/ART.CGO"))
		levels[key] = l
		return l
	}

	// findTyped locates basics of a named type in a linked object via the
	// relocation report's type-word patches.
	findTyped := func(rep *goalobj.LinkReport, typeName string) []uint32 {
		var out []uint32
		for off, name := range rep.SymbolRef {
			if name == typeName {
				out = append(out, off+4)
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
		return out
	}
	findCtrls := func(rep *goalobj.LinkReport) []uint32 {
		return findTyped(rep, "merc-ctrl")
	}

	type assetOut struct{ id, name, group string }
	var assets []assetOut

	// export writes one GLB: the named merc-ctrls of one art group (ctrl
	// index -1 = all ctrls as nodes of one file).
	export := func(dgoPath, entry, visName, fileName, title, group string, ctrlIdx int) {
		d := loadDGO(dgoPath)
		lv := loadLevel(dgoPath, visName)
		raw := dgoEntry(d, entry)
		if raw == nil {
			fmt.Fprintf(os.Stderr, "webexport: no entry %s in %s\n", entry, dgoPath)
			return
		}
		obj, rep, err := goalobj.Link(raw, 0, tab)
		check(err)
		ctrls := findCtrls(rep)
		if len(ctrls) == 0 {
			fmt.Fprintf(os.Stderr, "webexport: no merc-ctrls in %s\n", entry)
			return
		}
		if ctrlIdx >= 0 {
			ctrls = ctrls[ctrlIdx : ctrlIdx+1]
		}
		var clipNames []string
		jointType := tab.Syms["joint"].Value
		geoType := tab.Syms["art-joint-geo"].Value
		animType := tab.Syms["art-joint-anim"].Value
		geos := findTyped(rep, "art-joint-geo")
		_ = geoType
		scene := glb.NewScene()
		for ci, off := range ctrls {
			c, err := merc.Parse(obj, off)
			check(err)
			node := scene.AddNode(fmt.Sprintf("%s-%d", entry, ci), -1,
				[3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
			// skeleton: geo i pairs with ctrl i
			var joints []merc.Joint
			gi := ci
			if ctrlIdx >= 0 {
				gi = ctrlIdx
			}
			if gi < len(geos) {
				joints = merc.GeoJoints(obj, geos[gi], jointType)
			}
			tr := merc.NewSkinTracker()
			scale := float32frombits(u32of(obj, off+28))
			var prims []glb.Prim
			want, got := 0, 0
			for i := range c.Effects {
				var ps []glb.Prim
				if len(joints) > 0 {
					ps = merc.TexturedPrimsSkinned(&c.Effects[i], lv.resolve, tr, scale)
				} else {
					ps = merc.TexturedPrims(&c.Effects[i], lv.resolve)
				}
				want += c.Effects[i].TriCount
				for _, p := range ps {
					got += len(p.Tris)
				}
				prims = append(prims, ps...)
			}
			if got != want {
				fmt.Fprintf(os.Stderr, "webexport: %s ctrl %d: %d tris vs records' %d\n", entry, ci, got, want)
			}
			for pi := range prims {
				prims[pi].Layer = pi + 1
			}
			check(scene.AddMesh(node, fmt.Sprintf("%s-%d", entry, ci), prims))
			if len(joints) == 0 {
				continue
			}
			// joint nodes: local bind TRS; glTF ibm = transpose(Bind)
			nodeIDs := make([]int, len(joints))
			for j, jt := range joints {
				world := merc.Mat4(jt.Bind).Inverse()
				local := world
				parent := -1
				if jt.Parent >= 0 {
					local = world.Mul(merc.Mat4(joints[jt.Parent].Bind))
					parent = nodeIDs[jt.Parent]
				}
				t, q, s := local.TRS()
				nodeIDs[j] = scene.AddNode(jt.Name, parent, t, q, s)
			}
			// A row-major row-vector matrix has the same 16-float layout as
			// its column-major column-vector transpose: the Bind bytes ARE
			// the glTF inverse-bind matrix.
			ibm := make([][16]float32, len(joints))
			for j, jt := range joints {
				ibm[j] = jt.Bind
			}
			skin := scene.AddSkin(nodeIDs, ibm)
			scene.SetNodeSkin(node, skin)
			// animation clips (all anims whose joint count matches)
			for _, ap := range merc.FindAnims(obj, animType) {
				a, err := merc.DecodeJointAnim(obj, ap)
				if err != nil || a.NumJoints != len(joints) {
					continue
				}
				clip := scene.NewClip(a.Name)
				dup := false
				for _, n := range clipNames {
					if n == a.Name {
						dup = true
					}
				}
				if !dup {
					clipNames = append(clipNames, a.Name)
				}
				times := make([]float32, len(a.Frames))
				for f := range times {
					times[f] = float32(f) / 30
				}
				for j := 2; j < len(joints); j++ {
					qs := make([][4]float32, len(a.Frames))
					ts := make([][3]float32, len(a.Frames))
					ss := make([][3]float32, len(a.Frames))
					for f, fr := range a.Frames {
						qs[f] = fr[j].Quat
						ts[f] = fr[j].Trans
						ss[f] = fr[j].Scale
					}
					clip.Rotations(nodeIDs[j], times, qs)
					clip.Vec3s(nodeIDs[j], "translation", times, ts)
					clip.Vec3s(nodeIDs[j], "scale", times, ss)
				}
				clip.Finish()
			}
		}
		check(scene.Write(filepath.Join(*site, "objects", fileName+".glb"), fileName))
		objDoc := map[string]any{
			"format": "retro-x", "version": 1, "type": "model3d",
			"name": title, "model": fileName + ".glb",
		}
		if len(clipNames) > 0 {
			objDoc["skinnedClone"] = true
			var anims []any
			for _, n := range clipNames {
				anims = append(anims, map[string]any{"id": n, "clip": n, "loop": "loop"})
			}
			objDoc["animations"] = anims
		}
		writeJSON(filepath.Join(*site, "objects", fileName+".json"), objDoc)
		assets = append(assets, assetOut{fileName, title, group})
		fmt.Printf("webexport: %s (%s, %d ctrl(s))\n", fileName, entry, len(ctrls))
	}

	export("DGO/TIT.DGO", "logo", "title-vis", "title-logo", "Title logo", "Title screen", 0)
	export("DGO/TIT.DGO", "logo", "title-vis", "title-logo-jp", "Title logo (Japan)", "Title screen", 1)
	export("DGO/TIT.DGO", "ndi", "title-vis", "ndi-logo", "Naughty Dog logo", "Title screen", -1)
	export("DGO/INT.DGO", "evilbro", "intro-vis", "evilbro", "Gol (intro cutscene)", "Intro cutscene", -1)
	export("DGO/INT.DGO", "evilsis", "intro-vis", "evilsis", "Maia (intro cutscene)", "Intro cutscene", -1)

	var manifest []any
	for _, a := range assets {
		manifest = append(manifest, map[string]any{
			"id": a.id, "category": "object", "name": a.name, "group": a.group,
			"file": "objects/" + a.id + ".json"})
	}
	writeJSON(filepath.Join(*site, "manifest.json"), map[string]any{
		"format": "retro-x", "version": 1,
		"id": "jak-and-daxter-ps2", "title": "Jak & Daxter: The Precursor Legacy",
		"platform": "Sony PlayStation 2", "year": 2001,
		"description": "Naughty Dog's PlayStation 2 debut: a seamless world written in the studio's own Lisp (GOAL), streamed level by level from DGO archives the engine links into place at load time. This export begins where the game does — the title and intro models, decoded from the disc alone: the archive container and the engine's runtime linker were reimplemented byte-exact, and the skinned meshes are read straight out of the merc renderer's fragment format — fragment-local 8-bit vertex lattices, per-fragment origins hidden in float bit patterns, and triangle strips rebuilt from the file bytes alone (dest-byte scatter order, per-write ADC flags, the byte-header's stitch-copy tables), verified triangle-for-triangle against the microprogram's own emulated output. The materials are equally the disc's own: each fragment's adgif template names its texture by id, the level's remap table resolves it to a texture-page slot, and the tpage decoder supplies the pixels. The title model turns out to be two logos, English and Japanese. The intro cutscenes are on their way.",
		"assets": manifest,
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
