// webexport builds Pilotwings 64's Retro-X game tree from the cartridge
// archive (see RETROX.md).
//
// Nothing here boots the machine. The archive is parsed (extract/pwad), its
// textures, models and terrain decoded (extract/uvtx, uvmd, uvtr, uvct), the
// 31 songs rendered by the tools/platform/n64/audio synth, and the results
// handed to the shared builder (tools/lib/retrox). The oracle survives as a
// verification harness — cmd/mdldump -verify rebuilds each model's display
// list from the ROM and finds it byte-for-byte in RAM — but it is no longer
// in the export path.
//
// The headline of this export: a world is ONE level asset. The engine dresses
// the same island differently per mission by drawing an object only where its
// 16-bit mask meets the scene selector — that used to be 110 near-duplicate
// manifest entries ("Holiday Island · Set 3"), and is now one level per world
// with a VARIANT per mask bit and every placement carrying its variant
// membership. "Terrain only" is the default variant.
//
// # Axes
//
// The game is Z-up: terrain lies in the X/Y plane and height is Z. glTF is
// Y-up. Every exported position is therefore rotated
//
//	(x, y, z)  ->  (x, z, -y)
//
// which is a rotation about X, not a mirror: its determinant is +1, so
// triangle winding and face orientation carry over unchanged.
//
// Usage (from games/pilotwings-64-n64/): go run ./extract/cmd/webexport -in ROM
package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"sort"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvct"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtr"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtx"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/n64"
)

// knownModels names the handful of UVMD resources this project has identified
// by tracing them, rather than by looking at them. Everything else keeps its
// resource index: the archive carries no model names, and inventing them would
// be a guess.
var knownModels = map[int]struct{ id, name string }{
	47:  {"island", "Island (whole)"},           // the attract sequence's island
	212: {"pilotwings-logo", "PILOTWINGS logo"}, // 1,464 vertex-coloured triangles; the title card
	351: {"sky-dome", "Sky dome"},               // the attract sky, with its horizon band
}

// worldNames are the worlds this project has identified. Two by assembling
// them and recognising the result (Part IV); world 0 by the game's own words —
// the briefing screen the oracle drove into names it "Holiday Island". The
// rest are numbered.
var worldNames = map[int]struct{ id, name string }{
	0: {"holiday-island", "Holiday Island"},
	1: {"crescent-island", "Crescent Island"},
	3: {"little-states", "Little States"},
}

// waterModel names the UVMD resource that is a world's ocean plane, where the
// game has been observed to load it. Only world 0's is known: the beginner
// hang-glider lesson, driven in the oracle, DMA'd UVMD resource 360. Resources
// 365, 367 and 368 are the same plane for other worlds, but which belongs to
// which is not traced, and is not guessed here.
var waterModel = map[int]int{0: 360}

// toGL rotates a game-space position (Z up) into glTF space (Y up).
func toGL(x, y, z float32) [3]float32 { return [3]float32{x, z, -y} }

func main() {
	cli.Main("pilotwings-64-n64", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in ROM [-o DIR] [-only levels,objects,music]")
	}
	rom, err := n64.Load(ctx.In)
	if err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("Pilotwings 64")
	b.SetPlatform("Nintendo 64")
	b.SetYear(1996)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 320, H: 240},
		TickHz: 60,
		Filter: "crt",
		// The RDP samples textures bilinearly (3-point); point sampling
		// would misrepresent the hardware.
		TexFilter: "linear",
	})

	a, err := pwad.Open(rom.Data)
	if err != nil {
		return err
	}
	if err := a.Check(); err != nil {
		return err
	}
	texs := loadTextures(a)

	var assetByOrd []string
	if ctx.Stage("objects") {
		if assetByOrd, err = exportModels(ctx, a, texs); err != nil {
			return err
		}
	}
	if ctx.Stage("levels") {
		if err := exportWorlds(ctx, a, texs, assetByOrd); err != nil {
			return err
		}
	}
	if ctx.Stage("music") {
		if err := exportMusic(ctx, rom.Data); err != nil {
			return err
		}
	}
	return nil
}

// exportModels writes every non-empty UVMD as a model3d object asset and
// returns the asset id per UVMD ordinal ("" = no triangles at LOD 0).
func exportModels(ctx *cli.Context, a *pwad.Archive, texs []*uvtx.Texture) ([]string, error) {
	b := ctx.Builder
	uvmdIdx := a.ByType("UVMD")
	sort.Ints(uvmdIdx)
	animations := gatherAnimations(a) // by UVMD ordinal
	assetByOrd := make([]string, len(uvmdIdx))
	written, empty, animated := 0, 0, 0
	for ord, i := range uvmdIdx {
		f, err := a.Resource(i)
		if err != nil {
			return nil, err
		}
		m, err := uvmd.Decode(commOf(a, f))
		if err != nil {
			return nil, fmt.Errorf("UVMD %d: %w", i, err)
		}
		if m.Triangles(0) == 0 {
			empty++ // nothing to look at; not shipped, and counted below
			continue
		}
		id := fmt.Sprintf("uvmd-%04d", i)
		name := fmt.Sprintf("Model %03d", i)
		if k, ok := knownModels[i]; ok {
			id, name = k.id, k.name
		}
		glbPath, err := b.Path("objects", id+".glb")
		if err != nil {
			return nil, err
		}
		doc := &schema.Object{
			Type:      schema.ObjectModel3D,
			Name:      name,
			Model:     id + ".glb",
			Instanced: true,
			Props:     map[string]any{"uvmd": i, "ordinal": ord},
		}
		// A model any UVAN targets ships rigged and animated (a node per part
		// with rotation clips) instead of one baked static mesh.
		if anims := animations[ord]; len(anims) > 0 {
			clips := writeAnimatedModel(glbPath, m, anims, texs)
			for _, c := range clips {
				doc.Animations = append(doc.Animations, schema.Animation{
					ID: c, Clip: c, FPS: animFPS, Loop: "loop",
					Description: "Rigid per-part rotation keys (UVAN). The game's exact playback rate " +
						"is undecoded; 12 fps is this export's declared choice.",
				})
			}
			animated++
		} else if _, err := writeModel(glbPath, m, texs); err != nil {
			return nil, fmt.Errorf("UVMD %d: %w", i, err)
		}
		b.AddObject(schema.Asset{ID: id, Name: name, Group: "Models"}, doc)
		assetByOrd[ord] = id
		written++
		if written%50 == 0 {
			ctx.Progress("objects", written, len(uvmdIdx), id)
		}
	}
	ctx.Logf("objects: %d models written (%d animated), %d skipped for having no triangles at LOD 0",
		written, animated, empty)
	return assetByOrd, nil
}

// exportWorlds writes one scene3d level per world: terrain layer (+ ocean
// layer where traced), all placements once, and one variant per mask bit.
func exportWorlds(ctx *cli.Context, a *pwad.Archive, texs []*uvtx.Texture, assetByOrd []string) error {
	b := ctx.Builder
	worlds, chunks := loadWorld(a)
	uvmdIdx := a.ByType("UVMD")
	sort.Ints(uvmdIdx)

	droppedTotal := 0
	for i, w := range worlds {
		id, name := fmt.Sprintf("world-%d", i), fmt.Sprintf("World %d", i)
		if k, ok := worldNames[i]; ok {
			id, name = k.id, k.name
		}

		terrain, err := b.Path("levels", id, "terrain.glb")
		if err != nil {
			return err
		}
		tris, terrainBounds, err := writeWorld(terrain, w, chunks, texs)
		if err != nil {
			return fmt.Errorf("world %d: %w", i, err)
		}
		bounds := terrainBounds

		doc := &schema.Level{
			Type: schema.LevelScene3D,
			Scene: &schema.Scene{
				Layers: []schema.Layer{
					{ID: "terrain", Name: "Terrain", File: id + "/terrain.glb"},
				},
			},
		}
		if r, ok := waterModel[i]; ok {
			f, err := a.Resource(r)
			if err != nil {
				return err
			}
			m, err := uvmd.Decode(commOf(a, f))
			if err != nil {
				return fmt.Errorf("water UVMD %d: %w", r, err)
			}
			waterPath, err := b.Path("levels", id, "water.glb")
			if err != nil {
				return err
			}
			wb, err := writeModel(waterPath, m, texs)
			if err != nil {
				return fmt.Errorf("water UVMD %d: %w", r, err)
			}
			doc.Scene.Layers = append(doc.Scene.Layers, schema.Layer{
				ID: "water", Name: "Ocean", File: id + "/water.glb",
				Mode: schema.LayerToggle, Role: "water",
			})
			// The camera's far plane must cover the ocean, which dwarfs the
			// island — but the opening shot still frames the terrain.
			bounds = union(bounds, wb)
		}
		doc.Camera = worldCamera(bounds, terrainBounds)

		var placements []schema.Placement
		var present [MaskBits]bool
		var dropped int
		if assetByOrd != nil {
			placements, present, dropped = worldPlacements(w, chunks, assetByOrd, uvmdIdx)
			droppedTotal += dropped
		}
		// The variant list: "Terrain only" (default, no objects belong to it)
		// then one variant per mask bit any object in this world carries.
		doc.Variants = []schema.Variant{{ID: "terrain", Name: "Terrain only", Default: true}}
		sets := 0
		for bit := 0; bit < MaskBits; bit++ {
			if present[bit] {
				doc.Variants = append(doc.Variants, schema.Variant{
					ID: setID(bit), Name: fmt.Sprintf("Set %d", bit+1),
				})
				sets++
			}
		}
		doc.Placements = placements

		b.AddLevel(schema.Asset{ID: id, Name: name, Group: "Worlds"}, doc)
		ctx.Progress("levels", i+1, len(worlds),
			fmt.Sprintf("%-16s %6d triangles, %2d sets, %4d placements", name, tris, sets, len(placements)))
	}
	if droppedTotal > 0 {
		// A type whose GLB was pruned must not silently vanish — say so.
		fmt.Fprintf(os.Stderr, "warning: %d placements name a model with no triangles at LOD 0 and were dropped\n", droppedTotal)
	}
	return nil
}

func union(a, b bbox) bbox {
	out := a
	for i := 0; i < 3; i++ {
		if b.min[i] < out.min[i] {
			out.min[i] = b.min[i]
		}
		if b.max[i] > out.max[i] {
			out.max[i] = b.max[i]
		}
	}
	return out
}

// worldCamera derives the required camera block: the opening shot frames the
// TERRAIN from a high three-quarter view and the fly speed scales with the
// terrain span so every island handles alike, while the far plane covers the
// whole scene (the ocean plane dwarfs the island it surrounds).
func worldCamera(all, terrain bbox) *schema.Camera {
	cx, cy, cz := (terrain.min[0]+terrain.max[0])/2, (terrain.min[1]+terrain.max[1])/2, (terrain.min[2]+terrain.max[2])/2
	span := math.Max(float64(terrain.max[0]-terrain.min[0]), float64(terrain.max[2]-terrain.min[2]))
	allSpan := math.Max(float64(all.max[0]-all.min[0]), float64(all.max[2]-all.min[2]))
	return &schema.Camera{
		Mode:   "fly",
		Pos:    []float64{float64(cx) + span*0.55, float64(cy) + span*0.35, float64(cz) + span*0.55},
		Target: []float64{float64(cx), float64(cy), float64(cz)},
		FOV:    60,
		Near:   math.Max(1, span/2000),
		Far:    math.Max(span*5, allSpan*1.5),
		Fly:    &schema.Fly{Speed: math.Round(span * 0.06)},
	}
}

func commOf(a *pwad.Archive, f pwad.Form) []byte {
	for _, c := range f.Chunks {
		tag := c.Tag
		if c.Compressed() {
			tag = c.InnerTag
		}
		if tag == "COMM" {
			d, err := a.Data(c)
			if err != nil {
				die(err)
			}
			return d
		}
	}
	die(fmt.Errorf("resource %d has no COMM chunk", f.Index))
	return nil
}

func loadTextures(a *pwad.Archive) []*uvtx.Texture {
	idx := a.ByType("UVTX")
	sort.Ints(idx)
	out := make([]*uvtx.Texture, 0, len(idx))
	for _, i := range idx {
		f, err := a.Resource(i)
		if err != nil {
			die(err)
		}
		t, err := uvtx.Decode(commOf(a, f))
		if err != nil {
			die(fmt.Errorf("UVTX %d: %w", i, err))
		}
		out = append(out, t)
	}
	return out
}

func loadWorld(a *pwad.Archive) ([]*uvtr.World, []*uvct.Chunk) {
	f, err := a.Resource(a.ByType("UVTR")[0])
	if err != nil {
		die(err)
	}
	var worlds []*uvtr.World
	for _, c := range f.Chunks {
		if c.Tag != "COMM" {
			continue
		}
		data, err := a.Data(c)
		if err != nil {
			die(err)
		}
		w, err := uvtr.Decode(data)
		if err != nil {
			die(err)
		}
		if err := w.Check(); err != nil {
			die(err)
		}
		worlds = append(worlds, w)
	}
	idx := a.ByType("UVCT")
	sort.Ints(idx)
	chunks := make([]*uvct.Chunk, 0, len(idx))
	for _, i := range idx {
		rf, err := a.Resource(i)
		if err != nil {
			die(err)
		}
		c, err := uvct.Decode(commOf(a, rf))
		if err != nil {
			die(fmt.Errorf("UVCT %d: %w", i, err))
		}
		chunks = append(chunks, c)
	}
	return worlds, chunks
}

func white() *image.RGBA {
	w := image.NewRGBA(image.Rect(0, 0, 1, 1))
	w.SetRGBA(0, 0, color.RGBA{255, 255, 255, 255})
	return w
}

type bbox struct{ min, max [3]float32 }

// builder accumulates one GLB's vertex arrays, grouping triangles by texture so
// a world becomes one mesh rather than a pile of chunks.
type builder struct {
	pos   [][3]float32
	uvs   [][2]float32
	cols  [][4]uint8
	byTex map[int][][3]uint32
}

func newBuilder() *builder { return &builder{byTex: map[int][][3]uint32{}} }

// addBatch appends a batch's triangles, transformed by mtx and rotated to Y-up.
func (b *builder) addBatch(bt uvmd.Batch, verts []uvmd.Vertex, mtx uvmd.Matrix, texs []*uvtx.Texture) {
	tex := -1
	tw, th := 1.0, 1.0
	if t, ok := bt.Material.Texture(); ok && t < len(texs) {
		tex = t
		tw, th = float64(texs[t].Width), float64(texs[t].Height)
	}
	for _, tri := range bt.Tris {
		base := uint32(len(b.pos))
		for _, vi := range tri {
			v := verts[vi]
			x, y, z := float32(v.X), float32(v.Y), float32(v.Z)
			wx := mtx[0][0]*x + mtx[1][0]*y + mtx[2][0]*z + mtx[3][0]
			wy := mtx[0][1]*x + mtx[1][1]*y + mtx[2][1]*z + mtx[3][1]
			wz := mtx[0][2]*x + mtx[1][2]*y + mtx[2][2]*z + mtx[3][2]
			b.pos = append(b.pos, toGL(wx, wy, wz))
			b.uvs = append(b.uvs, [2]float32{
				float32(float64(v.S) / 32 / tw), float32(float64(v.T) / 32 / th),
			})
			b.cols = append(b.cols, [4]uint8{v.R, v.G, v.B, v.A})
		}
		b.byTex[tex] = append(b.byTex[tex], [3]uint32{base, base + 1, base + 2})
	}
}

func (b *builder) bounds() bbox {
	var bb bbox
	if len(b.pos) == 0 {
		return bb
	}
	bb.min, bb.max = b.pos[0], b.pos[0]
	for _, p := range b.pos {
		for i := 0; i < 3; i++ {
			if p[i] < bb.min[i] {
				bb.min[i] = p[i]
			}
			if p[i] > bb.max[i] {
				bb.max[i] = p[i]
			}
		}
	}
	return bb
}

// write emits the accumulated triangles as one textured GLB. opaque forces every
// texture to full alpha — true for terrain, which the game draws with alpha
// testing off (see exportImage).
func (b *builder) write(path string, texs []*uvtx.Texture, opaque bool) error {
	if len(b.pos) == 0 {
		return fmt.Errorf("nothing to write")
	}
	var keys []int
	for k := range b.byTex {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var groups []glb.TexturedGroup
	for _, k := range keys {
		g := glb.TexturedGroup{Tris: b.byTex[k], WrapS: 10497, WrapT: 10497}
		if k >= 0 {
			// A straight-alpha image with the alpha the format really carries, so
			// intensity textures stay grey and opaque instead of white with holes.
			g.Image, g.Blend = exportImage(texs[k], opaque)
		} else {
			g.Image = white()
		}
		groups = append(groups, g)
	}
	return glb.WriteTexturedColored(path, b.pos, b.uvs, b.cols, groups, nil)
}

func (b *builder) tris() int {
	n := 0
	for _, t := range b.byTex {
		n += len(t)
	}
	return n
}

func identity() uvmd.Matrix {
	var m uvmd.Matrix
	for i := 0; i < 4; i++ {
		m[i][i] = 1
	}
	return m
}

func writeWorld(path string, w *uvtr.World, chunks []*uvct.Chunk, texs []*uvtx.Texture) (int, bbox, error) {
	b := newBuilder()
	for i := range w.Cells {
		c := &w.Cells[i]
		if !c.Present {
			continue
		}
		ch := chunks[c.Chunk]
		for _, bt := range ch.Batches {
			b.addBatch(bt.Batch, ch.Vertices, c.Matrix, texs)
		}
	}
	return b.tris(), b.bounds(), b.write(path, texs, true)
}

// writeModel exports LOD 0, each part placed by its rest pose (the pairing of
// matrix i with part i holds at LOD 0 only; see extract/uvmd). Returns the
// mesh bounds (Y-up) for camera derivation.
func writeModel(path string, m *uvmd.Model, texs []*uvtx.Texture) (bbox, error) {
	b := newBuilder()
	for pi, part := range m.LODs[0].Parts {
		mtx := identity()
		if pi < len(m.Matrices) {
			mtx = m.Matrices[pi]
		}
		for _, bt := range part.Batches {
			b.addBatch(bt, m.Vertices, mtx, texs)
		}
	}
	return b.bounds(), b.write(path, texs, false)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "webexport:", err)
	os.Exit(1)
}
