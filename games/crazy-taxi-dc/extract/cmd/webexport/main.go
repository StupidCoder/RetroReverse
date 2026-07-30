// webexport builds the Studio's asset tree for Crazy Taxi: the three
// streamed courses — the arcade city, the original city and the Crazy Box
// arena — assembled from their BINC chunk containers and textured from the
// per-course TEXDC payloads via the descriptor arrays baked into
// 1ST_READ.BIN (Parts XI and XII of the writeup).
//
// The BINC entries are world-placed: concatenating a container assembles
// its whole course in course coordinates. Blocks group into one primitive
// per texture id; blocks whose aux word is 0xFFFFFFFF (the runtime-fed
// textures — marquees and the like) ship untextured in a neutral grey.
package main

import (
	"fmt"
	"image"
	"os"
	"sort"

	"retroreverse.com/games/crazy-taxi-dc/extract/assets"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/dc"
)

// The three streamed course sets. Course n's textures are TEXDC<n> — the
// pairing is exact: the maximum aux id in BINC<n> is one less than
// TEXDC<n>'s descriptor count, for every n. Course 0 is not a course at
// all: POLDC0/TEXDC0 load at boot and hold the shared object models.
var courses = []struct {
	n     int
	id    string
	name  string
	descr string
}{
	{1, "arcade-city", "Arcade city", "the West Coast city of the arcade game, streamed from BINC1.AFS during the drive"},
	{2, "original-city", "Original city", "the second city, built for the home release, streamed from BINC2.AFS"},
	{3, "crazy-box", "Crazy Box", "the minigame arenas, streamed from BINC3.AFS"},
}

func main() {
	cli.Main("crazy-taxi-dc", runCLI)
}

func runCLI(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in 'image/Crazy Taxi (US).cue' [-o DIR]")
	}
	b := ctx.Builder
	b.SetTitle("Crazy Taxi")
	b.SetPlatform("Sega Dreamcast")
	b.SetYear(2000)
	b.SetDescription("Sega's open-city taxi arcade hit, on the console it defined — " +
		"the repo's first Dreamcast target, whose oracle plays it from cold boot to a boarded fare. " +
		"The three courses ship here assembled from the game's own streaming containers: the model " +
		"format was read off the renderer's block walker, and every texture is decoded from the disc " +
		"alone through the descriptor arrays baked into the game binary — offsets, formats and mip " +
		"layouts all verified byte-exact against the running game's video memory.")
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 640, H: 480},
		TickHz: 60,
		// the Dreamcast on a VGA box: sharp 480p, bilinear texturing
		Filter:    "crt",
		TexFilter: "linear",
	})
	disc, err := dc.OpenDisc(ctx.In)
	if err != nil {
		return fmt.Errorf("open disc: %w", err)
	}
	if ctx.Stage("levels") {
		firstRead, err := disc.Vol.ReadFile("1ST_READ.BIN;1")
		if err != nil {
			return fmt.Errorf("1ST_READ.BIN: %w", err)
		}
		for i, c := range courses {
			if err := exportCourse(ctx, disc, firstRead, c.n, c.id, c.name); err != nil {
				return fmt.Errorf("%s: %w", c.id, err)
			}
			ctx.Progress("levels", i+1, len(courses), c.name)
		}
	}
	return nil
}

// exportCourse assembles course n from its BINC container and texture set.
func exportCourse(ctx *cli.Context, disc *dc.Disc, firstRead []byte, n int, id, name string) error {
	raw, err := disc.Vol.ReadFile(fmt.Sprintf("BINC%d.AFS;1", n))
	if err != nil {
		return err
	}
	afs, err := assets.OpenAFS(raw)
	if err != nil {
		return err
	}
	td, err := assets.OpenTexDir(firstRead, n)
	if err != nil {
		return err
	}
	texdc, err := disc.Vol.ReadFile(fmt.Sprintf("TEXDC%d.BIN;1", n))
	if err != nil {
		return err
	}

	var models []*assets.Model
	for i := range afs.Entries {
		e, err := afs.Data(i)
		if err != nil || len(e) == 0 {
			continue
		}
		ms, err := assets.OpenModels(e)
		if err != nil {
			return fmt.Errorf("entry %d: %w", i, err)
		}
		models = append(models, ms...)
	}

	// One batch per texture id; untextured and runtime-textured blocks pool
	// by TA list type.
	type batch struct {
		pos   [][3]float32
		nrm   [][3]float32
		uv    [][2]float32
		tris  [][3]uint32
		img   image.Image
		blend bool
		color [4]float32
	}
	texBatch := map[int]*batch{}
	listBatch := map[int]*batch{}
	var lo, hi [3]float32
	var sum [3]float64
	var nvert int
	first := true
	for _, m := range models {
		for _, blk := range m.Blocks {
			var bt *batch
			if blk.Textured() && blk.Aux != 0xFFFFFFFF && int(blk.Aux) < len(td.Entries) {
				bt = texBatch[int(blk.Aux)]
				if bt == nil {
					img, err := td.Decode(int(blk.Aux), texdc)
					if err != nil {
						return fmt.Errorf("texture %d: %w", blk.Aux, err)
					}
					bt = &batch{img: img, blend: blk.ListType() == 2, color: [4]float32{1, 1, 1, 1}}
					texBatch[int(blk.Aux)] = bt
				}
			} else {
				bt = listBatch[blk.ListType()]
				if bt == nil {
					bt = &batch{color: [4]float32{0.72, 0.72, 0.75, 1}, blend: blk.ListType() == 2}
					if bt.blend {
						bt.color[3] = 0.6
					}
					listBatch[blk.ListType()] = bt
				}
			}
			for _, s := range blk.Strips {
				base := uint32(len(bt.pos))
				for _, v := range s.Verts {
					bt.pos = append(bt.pos, v.Pos)
					bt.nrm = append(bt.nrm, v.Normal)
					bt.uv = append(bt.uv, [2]float32{v.U, v.V})
					nvert++
					for c := 0; c < 3; c++ {
						sum[c] += float64(v.Pos[c])
						if first || v.Pos[c] < lo[c] {
							lo[c] = v.Pos[c]
						}
						if first || v.Pos[c] > hi[c] {
							hi[c] = v.Pos[c]
						}
						if c == 2 {
							first = false
						}
					}
				}
				for _, t := range s.Tris() {
					bt.tris = append(bt.tris, [3]uint32{base + uint32(t[0]), base + uint32(t[1]), base + uint32(t[2])})
				}
			}
		}
	}

	// Deterministic primitive order: textures by id, then the list pools.
	var prims []glb.Prim
	var texIDs []int
	for k := range texBatch {
		texIDs = append(texIDs, k)
	}
	sort.Ints(texIDs)
	add := func(bt *batch) {
		if len(bt.tris) == 0 {
			return
		}
		prims = append(prims, glb.Prim{
			Positions: bt.pos, Normals: bt.nrm, UVs: bt.uv, Tris: bt.tris,
			Image: bt.img, BaseColor: bt.color, Unlit: true, Blend: bt.blend,
			// The PVR samples with wrap by default and the city tiles its
			// facades; the format's clamp/flip TSP bits are per-block
			// refinements not carried into the export.
			DoubleSided: true,
		})
	}
	for _, k := range texIDs {
		add(texBatch[k])
	}
	var listIDs []int
	for k := range listBatch {
		listIDs = append(listIDs, k)
	}
	sort.Ints(listIDs)
	for _, k := range listIDs {
		add(listBatch[k])
	}
	if len(prims) == 0 {
		return fmt.Errorf("no geometry")
	}

	file := id + ".glb"
	sc := glb.NewScene()
	node := sc.AddNode(id, -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
	if err := sc.AddMesh(node, id, prims); err != nil {
		return err
	}
	p, err := ctx.Builder.Path("levels", file)
	if err != nil {
		return err
	}
	if err := sc.Write(p, id); err != nil {
		return err
	}
	ctx.Logf("%s: bounds x %.0f..%.0f  y %.0f..%.0f  z %.0f..%.0f", id, lo[0], hi[0], lo[1], hi[1], lo[2], hi[2])

	// A fly camera opening over the middle of the course, high enough to
	// read the layout, looking toward the centre. The Crazy Box arena sits
	// inside its own sky dome, so its camera opens low, under the dome,
	// rather than at the dome's centroid.
	cx := float64(lo[0]+hi[0]) / 2
	cy := float64(lo[1]+hi[1]) / 2
	cz := float64(lo[2]+hi[2]) / 2
	span := float64(hi[0]-lo[0]) + float64(hi[2]-lo[2])
	pos := []float64{cx + span/8, cy + span/12, cz + span/8}
	target := []float64{cx, cy, cz}
	if id == "crazy-box" {
		// The vertex centroid sits in the arena town, not at the dome's
		// geometric centre (the dome is a few huge triangles).
		mx, my, mz := sum[0]/float64(nvert), sum[1]/float64(nvert), sum[2]/float64(nvert)
		pos = []float64{mx - span/14, my + span/16, mz - span/14}
		target = []float64{mx, my, mz}
	}
	doc := &schema.Level{
		Type: schema.LevelScene3D,
		Camera: &schema.Camera{
			Mode: "fly", FOV: 55, Near: 1, Far: span * 2,
			Pos:    pos,
			Target: target,
			Fly:    &schema.Fly{Speed: span / 40},
		},
		Scene: &schema.Scene{Layers: []schema.Layer{{ID: "city", File: file}}},
	}
	ctx.Builder.AddLevel(schema.Asset{ID: id, Name: name, Group: "Courses"}, doc)
	return nil
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "webexport: "+format+"\n", a...)
	os.Exit(1)
}
