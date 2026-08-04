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
	cache map[texKey]texEntry
}

// eyeSpec bakes a character's runtime-composited eye slots (merc.CompositeEye
// — see merc/eyes.go for the render-eyes derivation). The disc marks the two
// eyes with programmer_eye_left/_right placeholder textures; the sizes and
// resting gaze are the character's live eye-control values (title-logo.state
// array at *eye-control-array*), lids open. The iris/pupil textures are the
// shared runtime bindings (bam-iris-16x16, autoeye-pupil); the lid texture is
// per-character (autoeye-lid default, the sidekick's sk-eye-lid override) —
// all verified against the live composite by TBP+CBP.
type eyeSpec struct {
	lid  string // lid texture name
	l, r merc.EyeParams
}

// eyeSpecs by export entry name. Gol and Maia have no captured eye-control
// (their cutscene doesn't run in the saved states); they get the human
// template Jak uses.
var humanEyes = eyeSpec{
	lid: "autoeye-lid",
	l:   merc.EyeParams{X: -0.1171875, Y: 0.046875, IrisSize: 0.890625, PupilSize: 0.4375, LidHeight: 1},
	r:   merc.EyeParams{X: 0.1171875, Y: 0.046875, IrisSize: 0.890625, PupilSize: 0.4375, LidHeight: 1},
}

var eyeSpecs = map[string]eyeSpec{
	"eichar": humanEyes,
	"sidekick": {
		lid: "sk-eye-lid",
		l:   merc.EyeParams{X: -0.1953125, Y: 0.203125, IrisSize: 0.375, LidHeight: 1},
		r:   merc.EyeParams{X: 0.1953125, Y: 0.203125, IrisSize: 0.375, LidHeight: 1},
	},
	"evilbro": humanEyes,
	"evilsis": humanEyes,
}

// findTex locates a texture by name across the level's pages (raw GS alpha).
func (l *level) findTex(name string) image.Image {
	ids := make([]int, 0, len(l.pages))
	for id := range l.pages {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		pg := l.pages[id]
		for i := range pg.Textures {
			if pg.Textures[i].Name == name && pg.Textures[i].W > 0 {
				img, err := pg.DecodeGS(&pg.Textures[i], 0)
				check(err)
				return img
			}
		}
	}
	return nil
}

// baseGS returns a shader's base texture with raw GS alpha (the shine
// mask), routing the programmer_eye placeholders through the composite.
func (l *level) baseGS(s merc.ShaderRef, eyes *eyeSpec) image.Image {
	id := l.remap.Lookup(s.RawID)
	pg := l.pages[int(id>>20)]
	idx := int(id >> 8 & 0xFFF)
	if pg == nil || idx >= len(pg.Textures) || pg.Textures[idx].W == 0 {
		return nil
	}
	if eyes != nil {
		switch pg.Textures[idx].Name {
		case "programmer_eye_left":
			if iris, pupil, lid := l.findTex("bam-iris-16x16"), l.findTex("autoeye-pupil"), l.findTex(eyes.lid); iris != nil && pupil != nil && lid != nil {
				return merc.CompositeEye(iris, pupil, lid, eyes.l, false).RawImage()
			}
		case "programmer_eye_right":
			if iris, pupil, lid := l.findTex("bam-iris-16x16"), l.findTex("autoeye-pupil"), l.findTex(eyes.lid); iris != nil && pupil != nil && lid != nil {
				return merc.CompositeEye(iris, pupil, lid, eyes.r, true).RawImage()
			}
		}
	}
	img, err := pg.DecodeGS(&pg.Textures[idx], 0)
	check(err)
	return img
}

// envMean returns the mean RGB (0..1) of the envmap texture id — the
// average of the sphere-mapped sample over all reflection directions, the
// static stand-in for the view-dependent lookup (bam-eyelight: a tiny glint
// on black, mean ~0.02; bam-hairhilite: a broad highlight, mean ~0.29).
func (l *level) envMean(rawID uint32) [3]float32 {
	id := l.remap.Lookup(rawID)
	pg := l.pages[int(id>>20)]
	idx := int(id >> 8 & 0xFFF)
	if pg == nil || idx >= len(pg.Textures) || pg.Textures[idx].W == 0 {
		return [3]float32{1, 1, 1}
	}
	img, err := pg.DecodeGS(&pg.Textures[idx], 0)
	check(err)
	var sr, sg, sb, n int
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.RGBAAt(x, y)
			sr, sg, sb, n = sr+int(c.R), sg+int(c.G), sb+int(c.B), n+1
		}
	}
	return [3]float32{float32(sr) / float32(n) / 255, float32(sg) / float32(n) / 255, float32(sb) / float32(n) / 255}
}

// shine builds the additive-pass material for one shader of a flag-0x100
// effect (see merc/envmap.go). The hardware pass is Cs*Ad + Cd — Cs the
// sphere-mapped envmap texel x the tint (GS modulate), Ad the mask the base
// pass left in framebuffer alpha (= the base texture's alpha channel). The
// static bake replaces the view-dependent envmap sample with the texture's
// mean: texel = envMean x tint x mask x intensity(dist 0).
func (l *level) shine(s merc.ShaderRef, env *merc.EnvSpec, eyes *eyeSpec) merc.Material {
	base := l.baseGS(s, eyes)
	if base == nil {
		return merc.Material{}
	}
	// intensity at the bake's reference distance 0: clamp(Base, 0, 1)
	inten := env.Base
	if inten > 1 {
		inten = 1
	}
	if inten < 0 {
		inten = 0
	}
	mean := l.envMean(env.RawID)
	tint := func(c int) float32 { return mean[c] * float32(env.Tint[c]) / 128 * inten }
	b := base.Bounds()
	out := image.NewRGBA(b)
	at := func(x, y int) color.RGBA {
		switch im := base.(type) {
		case *image.RGBA: // tpage bytes: non-premultiplied + raw GS alpha
			return im.RGBAAt(x, y)
		case *image.NRGBA: // eye composite RawImage
			c := im.NRGBAAt(x, y)
			return color.RGBA{c.R, c.G, c.B, c.A}
		}
		return color.RGBA{}
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px := at(x, y)
			var r, g, bl float32
			if env.RawID != 0 {
				m := float32(px.A) / 128 // the mask, GS 0x80 = 1.0
				r, g, bl = tint(0)*m, tint(1)*m, tint(2)*m
			} else {
				r = tint(0) * float32(px.R) / 255
				g = tint(1) * float32(px.G) / 255
				bl = tint(2) * float32(px.B) / 255
			}
			out.SetRGBA(x, y, color.RGBA{clampF(r), clampF(g), clampF(bl), 255})
		}
	}
	return merc.Material{Image: out}
}

func clampF(v float32) uint8 {
	v *= 255
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// eyeImage composites one eye slot for the entry's spec.
func (l *level) eyeImage(spec eyeSpec, right bool) image.Image {
	iris := l.findTex("bam-iris-16x16")
	pupil := l.findTex("autoeye-pupil")
	lid := l.findTex(spec.lid)
	if iris == nil || pupil == nil || lid == nil {
		return nil
	}
	p := spec.l
	if right {
		p = spec.r
	}
	return merc.CompositeEye(iris, pupil, lid, p, right).Image()
}

type texEntry struct {
	img   image.Image
	blend bool
}

// Material modes, per effect (see resolve).
const (
	modeCutout = iota // base bucket: ATE GEQUAL 0x26 cutout
	modeOpaque        // alpha bucket: the replace pass covers every texel
)

type texKey struct {
	id   uint32
	mode int
}

// resolve maps a fragment's adgif to a GLB material under the measured
// bucket semantics (cmd/gsaudit on captured frames, root-ordered walk):
//
//   - flag-0 effects draw in the base merc bucket: PRIM ABE=0 and TEST
//     {ATE=1 ATST=GEQUAL AREF=0x26 AFAIL=KEEP} — binary cutout at the
//     hardware threshold (modeCutout).
//   - flag-0x100 effects draw in the alpha bucket with TEST ATE=0 (never
//     a cutout) and a per-fragment SEQUENCE of ALPHA equations: source-
//     over (0x44), REPLACE (0x00) and additive-by-dest-alpha (0x58, the
//     envmap shine; the logo's electric effect adds plain additive 0x68).
//     The replace pass gives every texel full coverage, so the net base
//     is opaque (modeOpaque); the additive shine passes are a separate
//     open item. The disc adgifs all carry source-over — the runtime
//     chain rewrites ALPHA per pass, so the disc value must not be used
//     to pick a mode.
func (l *level) resolve(s merc.ShaderRef, flags int, eyes *eyeSpec) merc.Material {
	mode := modeCutout
	if flags&0x100 != 0 {
		mode = modeOpaque
	}
	id := l.remap.Lookup(s.RawID)
	key := texKey{id, mode}
	if e, ok := l.cache[key]; ok {
		return merc.Material{Image: e.img, Blend: e.blend}
	}
	pg := l.pages[int(id>>20)]
	idx := int(id >> 8 & 0xFFF)
	if pg == nil || idx >= len(pg.Textures) || pg.Textures[idx].W == 0 {
		fmt.Fprintf(os.Stderr, "webexport: no texture for id %08x (raw %08x)\n", id, s.RawID)
		l.cache[key] = texEntry{}
		return merc.Material{}
	}
	// The 4x4 programmer_eye placeholders mark the runtime-composited eye
	// slots: bake them with the character's eye spec (uncached — the
	// composite is per-character, the cache key is only the texture id).
	if eyes != nil {
		switch pg.Textures[idx].Name {
		case "programmer_eye_left":
			if img := l.eyeImage(*eyes, false); img != nil {
				return merc.Material{Image: img}
			}
		case "programmer_eye_right":
			if img := l.eyeImage(*eyes, true); img != nil {
				return merc.Material{Image: img}
			}
		}
	}
	img, err := pg.Decode(&pg.Textures[idx], 0)
	check(err)
	out := gsAlpha(img, mode)
	l.cache[key] = texEntry{out, false}
	return merc.Material{Image: out}
}

// gsAlpha renders a texture's GS alpha channel (CLUT alpha, 0x80 = 1.0)
// into GLB semantics for the given material mode — see resolve for where
// the modes come from.
func gsAlpha(src image.Image, mode int) image.Image {
	b := src.Bounds()
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			var a8 uint8
			switch mode {
			case modeCutout: // binary at the alpha-test threshold
				a8 = 255
				if a>>8 < 0x26 { // GS units: CLUT alpha 0x80 = 1.0
					a8 = 0
				}
			case modeOpaque:
				a8 = 255
			}
			out.SetRGBA(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), a8})
		}
	}
	return out
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
			cache: map[texKey]texEntry{},
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

	// loadSpool links a STR spool container (sector-0 chunk-start table, v2
	// GOAL object chunks) and returns each art-joint-anim's chunk parts in
	// spool order. Chunk boundaries don't duplicate frames — the whole clip
	// is the parts' frames concatenated.
	spools := map[string]map[string][]*merc.JointAnim{}
	animType := tab.Syms["art-joint-anim"].Value
	loadSpool := func(path string) map[string][]*merc.JointAnim {
		if s, ok := spools[path]; ok {
			return s
		}
		data, err := vol.ReadFile(path + ";1")
		check(err)
		var starts []int
		for o := 0; o+4 <= 2048; o += 4 {
			v := int(u32of(data, uint32(o)))
			if v == 0 {
				break
			}
			starts = append(starts, v*2048)
		}
		s := map[string][]*merc.JointAnim{}
		for ci, st := range starts {
			end := len(data)
			if ci+1 < len(starts) {
				end = starts[ci+1]
			}
			obj, _, err := goalobj.Link(data[st:end], 0, tab)
			check(err)
			for _, ap := range merc.FindAnims(obj, animType) {
				a, err := merc.DecodeJointAnim(obj, ap)
				check(err)
				s[a.Name] = append(s[a.Name], a)
			}
		}
		spools[path] = s
		return s
	}

	// export writes one GLB: the named merc-ctrls of one art group (ctrl
	// index -1 = all ctrls as nodes of one file). entry may name another
	// archive as "PATH:name" — always-resident art (CGO/ART.CGO's player
	// models) draws with whichever level is loaded, so the texture remap
	// still comes from dgoPath's vis object.
	export := func(dgoPath, entry, visName, fileName, title, group string, ctrlIdx int, spoolFiles ...string) {
		entryDGO := dgoPath
		if i := strings.IndexByte(entry, ':'); i >= 0 {
			entryDGO, entry = entry[:i], entry[i+1:]
		}
		d := loadDGO(entryDGO)
		lv := loadLevel(dgoPath, visName)
		raw := dgoEntry(d, entry)
		if raw == nil {
			fmt.Fprintf(os.Stderr, "webexport: no entry %s in %s\n", entry, entryDGO)
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
			eyes := (*eyeSpec)(nil)
			if sp, ok := eyeSpecs[entry]; ok {
				eyes = &sp
			}
			var shinePrims []glb.Prim
			for i := range c.Effects {
				var ps []glb.Prim
				flags := c.Effects[i].Flags
				res := func(s merc.ShaderRef) merc.Material { return lv.resolve(s, flags, eyes) }
				if len(joints) > 0 {
					ps = merc.TexturedPrimsSkinned(&c.Effects[i], res, tr, scale)
				} else {
					ps = merc.TexturedPrims(&c.Effects[i], res)
				}
				want += c.Effects[i].TriCount
				for _, p := range ps {
					got += len(p.Tris)
				}
				prims = append(prims, ps...)
				// the flag-0x100 effects' second, additive pass: the same
				// geometry again with the extra-info envmap material.
				// Shader-less extras (env.RawID == 0, ALPHA 0x68 by FIX —
				// the logo's electric glow) are an animated flash (the
				// extra-info's type-2 rows read like a flicker table), not
				// a resting look, and are left out of the static bake.
				if env := merc.ParseExtraInfo(obj, c.Effects[i].ExtraInfo); env != nil && env.RawID != 0 && flags&0x100 != 0 {
					sres := func(s merc.ShaderRef) merc.Material { return lv.shine(s, env, eyes) }
					var sp []glb.Prim
					if len(joints) > 0 {
						sp = merc.TexturedPrimsSkinned(&c.Effects[i], sres, tr, scale)
					} else {
						sp = merc.TexturedPrims(&c.Effects[i], sres)
					}
					shinePrims = append(shinePrims, sp...)
				}
			}
			if got != want {
				fmt.Fprintf(os.Stderr, "webexport: %s ctrl %d: %d tris vs records' %d\n", entry, ci, got, want)
			}
			// shine prims draw after every base prim, in effect order
			prims = append(prims, shinePrims...)
			for pi := range prims {
				prims[pi].Layer = pi + 1
			}
			for pi := len(prims) - len(shinePrims); pi < len(prims); pi++ {
				prims[pi].Additive = true
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
			// animation clips: all resident anims whose joint count matches,
			// then any spooled (STR) clips that fit this skeleton
			addClip := func(name string, frames [][]merc.JointPose) {
				clip := scene.NewClip(name)
				dup := false
				for _, n := range clipNames {
					if n == name {
						dup = true
					}
				}
				if !dup {
					clipNames = append(clipNames, name)
				}
				times := make([]float32, len(frames))
				for f := range times {
					times[f] = float32(f) / 30
				}
				for j := 2; j < len(joints); j++ {
					qs := make([][4]float32, len(frames))
					ts := make([][3]float32, len(frames))
					ss := make([][3]float32, len(frames))
					for f, fr := range frames {
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
			for _, ap := range merc.FindAnims(obj, animType) {
				a, err := merc.DecodeJointAnim(obj, ap)
				if err != nil || a.NumJoints != len(joints) {
					continue
				}
				addClip(a.Name, a.Frames)
			}
			for _, sf := range spoolFiles {
				var names []string
				sp := loadSpool(sf)
				for n := range sp {
					names = append(names, n)
				}
				sort.Strings(names)
				for _, n := range names {
					parts := sp[n]
					if parts[0].NumJoints != len(joints) {
						continue
					}
					var frames [][]merc.JointPose
					for _, a := range parts {
						frames = append(frames, a.Frames...)
					}
					addClip(n, frames)
				}
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
	export("DGO/TIT.DGO", "ndi", "title-vis", "ndi-logo", "Naughty Dog logo", "Title screen", -1, "STR/NDINTRO.STR")
	export("DGO/TIT.DGO", "CGO/ART.CGO:eichar", "title-vis", "jak", "Jak", "Title screen", -1, "STR/NDINTRO.STR")
	export("DGO/TIT.DGO", "CGO/ART.CGO:sidekick", "title-vis", "daxter", "Daxter", "Title screen", -1, "STR/NDINTRO.STR")
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
