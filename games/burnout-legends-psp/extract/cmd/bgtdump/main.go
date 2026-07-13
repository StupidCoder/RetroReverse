// bgtdump reports what the bgt package finds in Burnout Legends' tracks: the
// texture dictionary, the material table, and every mesh of static.dat and of
// each streamed block.
//
// -verify is the check that matters: it re-derives, from the file alone, the
// facts the format has to satisfy — the strip words must consume exactly the
// vertices the record declares, every strip list must end in a RET, and every
// material must resolve to a texture. Run it over every track on the disc and a
// wrong guess about the layout shows up as a parse failure rather than as a
// quietly misplaced building.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path"
	"sort"
	"strings"

	"retroreverse.com/games/burnout-legends-psp/extract/bgt"
	"retroreverse.com/tools/platform/psp"
)

func main() {
	image := flag.String("image", "", "PSP UMD image (.cso or .iso)")
	only := flag.String("only", "", "just this track (e.g. US/C3_V1)")
	verify := flag.Bool("verify", false, "parse every track and report a summary line each")
	meshes := flag.Bool("meshes", false, "list every mesh")
	preview := flag.String("preview", "", "render the track from above to this PNG (a decode you can look at)")
	flag.Parse()

	if *image == "" {
		die("need -image")
	}
	im, err := psp.OpenImage(*image)
	if err != nil {
		die("open image: %v", err)
	}
	defer im.Close()

	tracks := findTracks(im)
	if len(tracks) == 0 {
		die("no tracks found")
	}

	failed := 0
	for _, dir := range tracks {
		name := strings.TrimPrefix(dir, "PSP_GAME/USRDIR/Tracks/")
		if *only != "" && !strings.Contains(name, *only) {
			continue
		}
		t, blocks, err := load(im, dir)
		if err != nil {
			fmt.Printf("%-12s FAILED: %v\n", name, err)
			failed++
			continue
		}

		staticTris, staticVerts := count(t.Meshes)
		var blockMeshes []bgt.Mesh
		for _, b := range blocks {
			blockMeshes = append(blockMeshes, b.Meshes...)
		}
		streamTris, streamVerts := count(blockMeshes)
		noTex := 0
		for _, m := range t.Materials {
			if m.Texture < 0 {
				noTex++
			}
		}
		fmt.Printf("%-12s %3d textures  %3d materials (%d untextured)  static: %3d meshes %6d verts %6d tris  streamed: %2d blocks %4d meshes %7d verts %7d tris\n",
			name, len(t.Textures), len(t.Materials), noTex,
			len(t.Meshes), staticVerts, staticTris,
			len(blocks), len(blockMeshes), streamVerts, streamTris)

		if *preview != "" {
			all := append([]bgt.Mesh(nil), t.Meshes...)
			all = append(all, blockMeshes...)
			if err := writePreview(*preview, t, all); err != nil {
				die("preview: %v", err)
			}
			fmt.Printf("wrote %s\n", *preview)
		}
		if *verify || *only == "" {
			continue
		}
		fmt.Printf("\ntextures:\n")
		for i, tx := range t.Textures {
			fmt.Printf("  %3d  %4dx%-4d CLUT%d  @0x%06X\n", i, tx.W, tx.H, tx.BPP, tx.Offset)
		}
		fmt.Printf("\nmaterials:\n")
		for i, m := range t.Materials {
			fmt.Printf("  %3d  texture %3d  param %6.2f  flags %04X\n", i, m.Texture, m.Param, m.Flags)
		}
		if *meshes {
			fmt.Printf("\nstatic meshes:\n")
			for i, m := range t.Meshes {
				dumpMesh(i, &m)
			}
			for _, b := range blocks {
				fmt.Printf("\nstreamed block @0x%06X (cell %d): %d meshes\n", b.Offset, b.Cell, len(b.Meshes))
				for i := range b.Meshes {
					dumpMesh(i, &b.Meshes[i])
				}
			}
		}
	}
	if failed > 0 {
		fmt.Printf("\n%d track(s) failed to parse\n", failed)
		os.Exit(1)
	}
}

func dumpMesh(i int, m *bgt.Mesh) {
	fmt.Printf("  %3d @0x%06X  mat %3d  %4d verts  %3d strips  centre (%8.1f,%7.1f,%8.1f)  extent (%7.1f,%6.1f,%7.1f)\n",
		i, m.Offset, m.Material, len(m.Verts), len(m.Strips),
		m.Centre[0], m.Centre[1], m.Centre[2], m.Extent[0], m.Extent[1], m.Extent[2])
}

func count(ms []bgt.Mesh) (tris, verts int) {
	for i := range ms {
		tris += len(ms[i].Tris())
		verts += len(ms[i].Verts)
	}
	return
}

// load reads a track's static.dat and streamed.dat and decodes both.
func load(im *psp.Image, dir string) (*bgt.Track, []bgt.Block, error) {
	data, err := im.ReadFile(path.Join(dir, "static.dat"))
	if err != nil {
		return nil, nil, err
	}
	t, err := bgt.Parse(data)
	if err != nil {
		return nil, nil, err
	}
	sd, err := im.ReadFile(path.Join(dir, "streamed.dat"))
	if err != nil {
		return t, nil, nil // a track need not stream
	}
	blocks, err := t.ParseStreamed(sd)
	if err != nil {
		return nil, nil, err
	}
	return t, blocks, nil
}

// findTracks lists the directories that hold a static.dat.
func findTracks(im *psp.Image) []string {
	var out []string
	im.Walk(func(e psp.Entry) error {
		if !e.IsDir && strings.HasSuffix(e.Path, "/static.dat") {
			out = append(out, path.Dir(e.Path))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "bgtdump: "+format+"\n", a...)
	os.Exit(1)
}

// writePreview renders the decoded track from directly above, z-buffered and
// textured. Looking at the result is the only check that catches a decode which
// is internally consistent and geometrically wrong.
func writePreview(path string, t *bgt.Track, meshes []bgt.Mesh) error {
	const size = 1400
	lo := [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
	hi := [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
	for i := range meshes {
		for _, v := range meshes[i].Verts {
			p := [3]float32{v.X, v.Y, v.Z}
			for k := 0; k < 3; k++ {
				lo[k] = min32(lo[k], p[k])
				hi[k] = max32(hi[k], p[k])
			}
		}
	}
	scale := float32(size-40) / max32(hi[0]-lo[0], hi[2]-lo[2])
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}
	zb := make([]float32, size*size)
	for i := range zb {
		zb[i] = -math.MaxFloat32
	}
	tex := map[int]*image.RGBA{}
	for i := range t.Textures {
		if x, err := t.Texture(i); err == nil {
			tex[i] = x
		}
	}
	type pt struct {
		x, y, z, u, v float32
		r, g, b       uint8
	}
	at := func(v bgt.Vertex) pt {
		return pt{20 + (v.X-lo[0])*scale, 20 + (v.Z-lo[2])*scale, v.Y, v.U, v.V, v.R, v.G, v.B}
	}
	for i := range meshes {
		m := &meshes[i]
		var tx *image.RGBA
		if m.Material >= 0 {
			if ti := t.Materials[m.Material].Texture; ti >= 0 {
				tx = tex[ti]
			}
		}
		for _, f := range m.Tris() {
			a, b, c := at(m.Verts[f[0]]), at(m.Verts[f[1]]), at(m.Verts[f[2]])
			minx := int(math.Floor(float64(min32(a.x, min32(b.x, c.x)))))
			maxx := int(math.Ceil(float64(max32(a.x, max32(b.x, c.x)))))
			miny := int(math.Floor(float64(min32(a.y, min32(b.y, c.y)))))
			maxy := int(math.Ceil(float64(max32(a.y, max32(b.y, c.y)))))
			area := (b.x-a.x)*(c.y-a.y) - (b.y-a.y)*(c.x-a.x)
			if area == 0 {
				continue
			}
			if minx < 0 {
				minx = 0
			}
			if miny < 0 {
				miny = 0
			}
			if maxx > size-1 {
				maxx = size - 1
			}
			if maxy > size-1 {
				maxy = size - 1
			}
			for y := miny; y <= maxy; y++ {
				for x := minx; x <= maxx; x++ {
					px, py := float32(x)+0.5, float32(y)+0.5
					w0 := ((c.x-b.x)*(py-b.y) - (c.y-b.y)*(px-b.x)) / area
					w1 := ((a.x-c.x)*(py-c.y) - (a.y-c.y)*(px-c.x)) / area
					w2 := 1 - w0 - w1
					if w0 < 0 || w1 < 0 || w2 < 0 {
						continue
					}
					z := w0*a.z + w1*b.z + w2*c.z
					o := y*size + x
					if z < zb[o] {
						continue
					}
					R, G, B := float32(255), float32(255), float32(255)
					if tx != nil {
						u := w0*a.u + w1*b.u + w2*c.u
						v := w0*a.v + w1*b.v + w2*c.v
						w, h := tx.Bounds().Dx(), tx.Bounds().Dy()
						sx := int(u*float32(w)) % w
						sy := int(v*float32(h)) % h
						if sx < 0 {
							sx += w
						}
						if sy < 0 {
							sy += h
						}
						tc := tx.RGBAAt(sx, sy)
						if tc.A == 0 {
							continue
						}
						R, G, B = float32(tc.R), float32(tc.G), float32(tc.B)
					}
					cr := w0*float32(a.r) + w1*float32(b.r) + w2*float32(c.r)
					zb[o] = z
					img.SetRGBA(x, y, color.RGBA{
						clamp8(R * cr / 255), clamp8(G * cr / 255), clamp8(B * cr / 255), 255})
				}
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func clamp8(f float32) uint8 {
	if f < 0 {
		return 0
	}
	if f > 255 {
		return 255
	}
	return uint8(f)
}
func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
