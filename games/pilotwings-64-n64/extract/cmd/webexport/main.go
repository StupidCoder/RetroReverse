// webexport builds Pilotwings 64's Studio assets from the cartridge archive.
//
// Nothing here boots the machine. The archive is parsed (extract/pwad), its
// textures, models and terrain decoded (extract/uvtx, uvmd, uvtr, uvct), and
// the results written as GLBs plus a format-2 manifest. The oracle survives as
// a verification harness — cmd/mdldump -verify rebuilds each model's display
// list from the ROM and finds it byte-for-byte in RAM, and cmd/dlverify checks
// the walk against the RDP stream — but it is no longer in the export path.
//
// # Axes
//
// The game is Z-up: terrain lies in the X/Y plane and height is Z (a world
// grid's cell centres are (x,y), and the island model's height runs 0..502 in
// Z). glTF is Y-up. Every exported position is therefore rotated
//
//	(x, y, z)  ->  (x, z, -y)
//
// which is a rotation about X, not a mirror: its determinant is +1, so triangle
// winding and face orientation carry over unchanged.
//
// Usage:
//
//	webexport -image ROM -o site/public/pilotwings-64-n64
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvct"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtr"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtx"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n64"
)

// knownModels names the handful of UVMD resources this project has identified
// by tracing them, rather than by looking at them. Everything else keeps its
// resource index: the archive carries no model names, and inventing them would
// be a guess.
var knownModels = map[int]string{
	47:  "Island (whole)",  // the attract sequence's island, verified against the RAM walk
	212: "PILOTWINGS logo", // 1,464 vertex-coloured triangles; drawn on the title card
	351: "Sky dome",        // the attract sky, with its horizon band
}

// worldNames are the only two worlds this project has identified, both by
// assembling them and recognising the result (Part IV). The rest are numbered.
var worldNames = map[int]string{
	1: "Crescent Island",
	3: "Little States",
}

// toGL rotates a game-space position (Z up) into glTF space (Y up).
func toGL(x, y, z float32) [3]float32 { return [3]float32{x, z, -y} }

// Manifest is the format-2 asset index the Studio loads.
type Manifest struct {
	Format   int          `json:"format"`
	Game     string       `json:"game"`
	Platform string       `json:"platform"`
	Native   Size         `json:"native"`
	TickHz   int          `json:"tickHz"`
	Models   []ModelIndex `json:"models,omitempty"`
}

type Size struct{ W, H int }

type ModelIndex struct {
	Name    string `json:"name"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Section string `json:"section,omitempty"`
}

func main() {
	image_ := flag.String("image", "", "cartridge image")
	out := flag.String("o", "", "output directory (site/public/pilotwings-64-n64)")
	flag.Parse()
	if *image_ == "" || *out == "" {
		die(fmt.Errorf("-image and -o are required"))
	}

	rom, err := n64.Load(*image_)
	if err != nil {
		die(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		die(err)
	}
	if err := a.Check(); err != nil {
		die(err)
	}

	texs := loadTextures(a)
	modelsDir := filepath.Join(*out, "models")
	worldsDir := filepath.Join(*out, "worlds")
	for _, d := range []string{modelsDir, worldsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			die(err)
		}
	}

	man := Manifest{Format: 2, Game: "pilotwings-64-n64", Platform: "Nintendo 64",
		Native: Size{320, 240}, TickHz: 60}

	// --- worlds -----------------------------------------------------------
	worlds, chunks := loadWorld(a)
	for i, w := range worlds {
		name := worldNames[i]
		if name == "" {
			name = fmt.Sprintf("World %d", i)
		}
		file := fmt.Sprintf("worlds/world-%d.glb", i)
		tris, err := writeWorld(filepath.Join(*out, file), w, chunks, texs)
		if err != nil {
			die(fmt.Errorf("world %d: %w", i, err))
		}
		fmt.Printf("world %d %-18s %6d triangles\n", i, name, tris)
		man.Models = append(man.Models, ModelIndex{Name: name, File: file, Kind: "mesh3d", Section: "Worlds"})
	}

	// --- models -----------------------------------------------------------
	written, empty := 0, 0
	for _, i := range a.ByType("UVMD") {
		f, err := a.Resource(i)
		if err != nil {
			die(err)
		}
		m, err := uvmd.Decode(commOf(a, f))
		if err != nil {
			die(fmt.Errorf("UVMD %d: %w", i, err))
		}
		if m.Triangles(0) == 0 {
			empty++ // nothing to look at; not shipped, and said so below
			continue
		}
		file := fmt.Sprintf("models/uvmd-%04d.glb", i)
		if err := writeModel(filepath.Join(*out, file), m, texs); err != nil {
			die(fmt.Errorf("UVMD %d: %w", i, err))
		}
		name := knownModels[i]
		if name == "" {
			name = fmt.Sprintf("Model %03d", i)
		}
		man.Models = append(man.Models, ModelIndex{Name: name, File: file, Kind: "mesh3d", Section: "Models"})
		written++
	}
	fmt.Printf("%d models written, %d skipped for having no triangles at LOD 0\n", written, empty)

	j, _ := json.MarshalIndent(man, "", "  ")
	if err := os.WriteFile(filepath.Join(*out, "manifest.json"), append(j, '\n'), 0o644); err != nil {
		die(err)
	}
	fmt.Printf("wrote manifest.json: %d entries\n", len(man.Models))
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

func (b *builder) write(path string, texs []*uvtx.Texture) error {
	if len(b.pos) == 0 {
		return fmt.Errorf("nothing to write")
	}
	var keys []int
	for k := range b.byTex {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	w := white()
	var groups []glb.TexturedGroup
	for _, k := range keys {
		img := w
		if k >= 0 {
			img = texs[k].Image
		}
		groups = append(groups, glb.TexturedGroup{Tris: b.byTex[k], Image: img, WrapS: 10497, WrapT: 10497})
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

func writeWorld(path string, w *uvtr.World, chunks []*uvct.Chunk, texs []*uvtx.Texture) (int, error) {
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
	return b.tris(), b.write(path, texs)
}

// writeModel exports LOD 0, each part placed by its rest pose (the pairing of
// matrix i with part i holds at LOD 0 only; see extract/uvmd).
func writeModel(path string, m *uvmd.Model, texs []*uvtx.Texture) error {
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
	return b.write(path, texs)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "webexport:", err)
	os.Exit(1)
}
