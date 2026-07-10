// mdldump decodes every UVMD model straight out of the cartridge archive.
//
// With -verify it performs the narrow-seam check: for each model resident in an
// RDRAM snapshot it locates the vertex pool the loader copied there, rebuilds —
// from the ROM alone — the exact Fast3D display list the game's own emitter at
// 0x80225940 would write for every batch, and requires those bytes to appear in
// RAM. The engine wrote them; we derived them from the file. Matching command
// words verify the command-stream decode; the pool address verifies structure.
//
// Usage:
//
//	mdldump -image ROM
//	mdldump -image ROM -verify work/flyby-ram.bin
//	mdldump -image ROM -glb work/models
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtx"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n64"
)

func main() {
	image_ := flag.String("image", "", "cartridge image")
	verify := flag.String("verify", "", "RDRAM snapshot: rebuild each resident model's display lists and find them in RAM")
	glbDir := flag.String("glb", "", "write one GLB per model into this directory")
	lod := flag.Int("lod", 0, "level of detail to export")
	flag.Parse()

	rom, err := n64.Load(*image_)
	if err != nil {
		die(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		die(err)
	}

	models, err := decodeAll(a)
	if err != nil {
		die(err)
	}
	tris, batches, untextured := 0, 0, 0
	maxTex := -1
	for _, m := range models {
		tris += m.model.Triangles(0)
		for _, p := range m.model.LODs[0].Parts {
			for _, b := range p.Batches {
				batches++
				if t, ok := b.Material.Texture(); ok {
					if t > maxTex {
						maxTex = t
					}
				} else {
					untextured++
				}
			}
		}
	}
	fmt.Printf("decoded %d UVMD models: %d triangles at LOD 0, %d batches (%d untextured)\n",
		len(models), tris, batches, untextured)
	fmt.Printf("highest UVTX ordinal referenced: %d (the archive has %d)\n", maxTex, len(a.ByType("UVTX")))

	if *verify != "" {
		if err := verifyAgainstRAM(models, *verify); err != nil {
			die(err)
		}
	}
	if *glbDir != "" {
		if err := writeGLBs(a, models, *glbDir, *lod); err != nil {
			die(err)
		}
	}
}

type entry struct {
	idx   int
	model *uvmd.Model
	data  []byte
}

func commChunk(a *pwad.Archive, f pwad.Form) ([]byte, error) {
	for _, c := range f.Chunks {
		tag := c.Tag
		if c.Compressed() {
			tag = c.InnerTag
		}
		if tag == "COMM" {
			return a.Data(c)
		}
	}
	return nil, fmt.Errorf("resource %d has no COMM chunk", f.Index)
}

func decodeAll(a *pwad.Archive) ([]entry, error) {
	var out []entry
	for _, i := range a.ByType("UVMD") {
		f, err := a.Resource(i)
		if err != nil {
			return nil, err
		}
		data, err := commChunk(a, f)
		if err != nil {
			return nil, err
		}
		m, err := uvmd.Decode(data)
		if err != nil {
			return nil, fmt.Errorf("UVMD %d: %w", i, err)
		}
		out = append(out, entry{idx: i, model: m, data: data})
	}
	return out, nil
}

// emitDL rebuilds the byte-exact display list the game's emitter writes for one
// batch, given where the vertex pool landed in RDRAM. The encoding is read off
// 0x80225940: a G_VTX carries ((n-1)<<4)|slot in bits 16..23 and n*16 in the low
// half; a G_TRI1 carries the three slot indices multiplied by the 10-byte index
// stride; the list is terminated by G_ENDDL.
func emitDL(b uvmd.Batch, poolPhys uint32) []byte {
	var slots [16]int
	for i := range slots {
		slots[i] = -1
	}
	// Replay the same stream the decoder did, but emit rather than collect. The
	// decoder already validated it, so nothing here can fail.
	var out []byte
	w := func(w0, w1 uint32) {
		var b [8]byte
		binary.BigEndian.PutUint32(b[0:], w0)
		binary.BigEndian.PutUint32(b[4:], w1)
		out = append(out, b[:]...)
	}
	for _, cmd := range b.Stream {
		if cmd.Tri {
			w(0xBF000000, uint32(cmd.I0*10)<<16|uint32(cmd.I1*10)<<8|uint32(cmd.I2*10))
			continue
		}
		// The vertex pointer is a segmented address (segment 0 here), so it is
		// the pool's physical address plus the record offset — not a KSEG0 one.
		pack := uint32((cmd.Count-1)<<4 | cmd.Slot)
		w(0x04000000|pack<<16|uint32(cmd.Count*16)&0xFFFF,
			poolPhys+uint32(cmd.First)*uvmd.VertexSize)
	}
	w(0xB8000000, 0)
	return out
}

func verifyAgainstRAM(models []entry, ramPath string) error {
	ram, err := os.ReadFile(ramPath)
	if err != nil {
		return err
	}
	resident, found, missing := 0, 0, 0
	var partial []string
	for _, e := range models {
		if len(e.model.Vertices) == 0 {
			continue
		}
		// Where did the loader put this model's vertex pool? Match on its bytes.
		pool := e.data[uvmd.HeaderSize : uvmd.HeaderSize+len(e.model.Vertices)*uvmd.VertexSize]
		at := bytes.Index(ram, pool)
		if at < 0 {
			continue
		}
		resident++
		hit, miss := 0, 0
		for _, p := range e.model.LODs[0].Parts {
			for _, b := range p.Batches {
				if bytes.Contains(ram, emitDL(b, uint32(at))) {
					hit++
				} else {
					miss++
				}
			}
		}
		found += hit
		missing += miss
		if miss > 0 {
			partial = append(partial, fmt.Sprintf("UVMD %d (pool %06X): %d/%d lists found", e.idx, at, hit, hit+miss))
		}
	}
	fmt.Printf("verify: %d models resident; %d display lists rebuilt from the ROM appear byte-for-byte in RAM, %d do not\n",
		resident, found, missing)
	for _, s := range partial {
		fmt.Println("  " + s)
	}
	if resident == 0 {
		return fmt.Errorf("verify: no model was found in the snapshot")
	}
	return nil
}

func writeGLBs(a *pwad.Archive, models []entry, dir string, lod int) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	texs, err := loadTextures(a)
	if err != nil {
		return err
	}
	white := image.NewRGBA(image.Rect(0, 0, 1, 1))
	white.SetRGBA(0, 0, color.RGBA{255, 255, 255, 255})

	written := 0
	for _, e := range models {
		m := e.model
		if lod >= len(m.LODs) {
			continue
		}
		var pos [][3]float32
		var uvs [][2]float32
		var cols [][4]uint8
		var groups []glb.TexturedGroup
		for pi, part := range m.LODs[lod].Parts {
			// Place each part by its rest pose, so an articulated model
			// reassembles the way the game holds it at rest. The pairing of
			// matrix i with part i is only sound at LOD 0 — a lower LOD may
			// drop a part — so below that the parts are left unplaced.
			mtx := identity()
			if lod == 0 && pi < len(m.Matrices) {
				mtx = m.Matrices[pi]
			}
			for _, b := range part.Batches {
				img := white
				tw, th := 1.0, 1.0
				if t, ok := b.Material.Texture(); ok && t < len(texs) {
					img = texs[t].Image
					tw, th = float64(texs[t].Width), float64(texs[t].Height)
				}
				var tris [][3]uint32
				for _, t := range b.Tris {
					base := uint32(len(pos))
					for _, vi := range t {
						v := m.Vertices[vi]
						x, y, z := float64(v.X), float64(v.Y), float64(v.Z)
						pos = append(pos, [3]float32{
							float32(mtx[0][0]*float32(x) + mtx[1][0]*float32(y) + mtx[2][0]*float32(z) + mtx[3][0]),
							float32(mtx[0][1]*float32(x) + mtx[1][1]*float32(y) + mtx[2][1]*float32(z) + mtx[3][1]),
							float32(mtx[0][2]*float32(x) + mtx[1][2]*float32(y) + mtx[2][2]*float32(z) + mtx[3][2]),
						})
						uvs = append(uvs, [2]float32{
							float32(float64(v.S) / 32 / tw), float32(float64(v.T) / 32 / th),
						})
						cols = append(cols, [4]uint8{v.R, v.G, v.B, v.A})
					}
					tris = append(tris, [3]uint32{base, base + 1, base + 2})
				}
				if len(tris) == 0 {
					continue
				}
				groups = append(groups, glb.TexturedGroup{Tris: tris, Image: img, WrapS: 10497, WrapT: 10497})
			}
		}
		if len(groups) == 0 {
			continue
		}
		p := filepath.Join(dir, fmt.Sprintf("uvmd-%04d.glb", e.idx))
		if err := glb.WriteTexturedColored(p, pos, uvs, cols, groups, nil); err != nil {
			return err
		}
		written++
	}
	fmt.Printf("wrote %d GLBs to %s\n", written, dir)
	return nil
}

func identity() uvmd.Matrix {
	var m uvmd.Matrix
	for i := 0; i < 4; i++ {
		m[i][i] = 1
	}
	return m
}

func loadTextures(a *pwad.Archive) ([]*uvtx.Texture, error) {
	idx := a.ByType("UVTX")
	sort.Ints(idx)
	out := make([]*uvtx.Texture, 0, len(idx))
	for _, i := range idx {
		f, err := a.Resource(i)
		if err != nil {
			return nil, err
		}
		data, err := commChunk(a, f)
		if err != nil {
			return nil, err
		}
		t, err := uvtx.Decode(data)
		if err != nil {
			return nil, fmt.Errorf("UVTX %d: %w", i, err)
		}
		out = append(out, t)
	}
	return out, nil
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "mdldump:", err)
	os.Exit(1)
}
