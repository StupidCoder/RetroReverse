// bgvdump parses Burnout Legends' vehicle models (.bgv) straight off the UMD
// and reports what it finds: the detail levels, their meshes, the strips and
// the bounding box. With -all it parses every car on the disc, which is the
// check that matters — a format is only decoded when it decodes everything.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/burnout-legends-psp/extract/bgv"
	"retroreverse.com/tools/platform/psp"
)

func main() {
	image := flag.String("image", "", "PSP UMD image (.cso or .iso)")
	car := flag.String("car", "PSP_GAME/USRDIR/pveh/COMP/Car1.bgv", "vehicle model to dump")
	all := flag.Bool("all", false, "parse every .bgv on the disc and report failures")
	flag.Parse()

	im, err := psp.OpenImage(*image)
	if err != nil {
		die("open image: %v", err)
	}
	defer im.Close()

	if *all {
		dumpAll(im)
		return
	}
	data, err := im.ReadFile(*car)
	if err != nil {
		die("read %s: %v", *car, err)
	}
	m, err := bgv.Parse(data)
	if err != nil {
		die("parse: %v", err)
	}
	fmt.Printf("%s: %d bytes, %d detail levels\n", *car, len(data), len(m.Levels))
	for li, meshes := range m.Levels {
		verts, tris := 0, 0
		for i := range meshes {
			verts += len(meshes[i].Verts)
			tris += len(meshes[i].Tris())
		}
		fmt.Printf("  level %d: %2d meshes  %5d verts  %5d tris\n", li, len(meshes), verts, tris)
		for i := range meshes {
			me := &meshes[i]
			lo, hi := bounds(me)
			fmt.Printf("    mesh %2d @0x%06X  %4d verts %3d strips  scale %.0f  bbox (%.2f %.2f %.2f)..(%.2f %.2f %.2f)\n",
				i, me.Offset, len(me.Verts), len(me.Strips), me.ScaleX,
				lo[0], lo[1], lo[2], hi[0], hi[1], hi[2])
		}
	}
}

// dumpAll is the real test of the decoder: every vehicle on the disc.
func dumpAll(im *psp.Image) {
	var paths []string
	im.Walk(func(e psp.Entry) error {
		if !e.IsDir && strings.HasSuffix(e.Path, ".bgv") {
			paths = append(paths, e.Path)
		}
		return nil
	})
	sort.Strings(paths)
	okCount, failCount, totalTris := 0, 0, 0
	for _, p := range paths {
		data, err := im.ReadFile(p)
		if err != nil {
			fmt.Printf("FAIL  %-24s read: %v\n", base(p), err)
			failCount++
			continue
		}
		m, err := bgv.Parse(data)
		if err != nil {
			fmt.Printf("FAIL  %-24s %v\n", base(p), err)
			failCount++
			continue
		}
		verts, tris, meshes := 0, 0, 0
		for _, lvl := range m.Levels {
			for i := range lvl {
				meshes++
				verts += len(lvl[i].Verts)
				tris += len(lvl[i].Tris())
			}
		}
		totalTris += tris
		okCount++
		fmt.Printf("ok    %-24s %d levels %3d meshes %6d verts %6d tris\n",
			base(p), len(m.Levels), meshes, verts, tris)
	}
	fmt.Printf("\n%d parsed, %d failed, %d triangles total\n", okCount, failCount, totalTris)
	if failCount > 0 {
		os.Exit(1)
	}
}

func bounds(m *bgv.Mesh) (lo, hi [3]float32) {
	if len(m.Verts) == 0 {
		return
	}
	v := m.Verts[0]
	lo = [3]float32{v.X, v.Y, v.Z}
	hi = lo
	for _, v := range m.Verts {
		for i, c := range [3]float32{v.X, v.Y, v.Z} {
			if c < lo[i] {
				lo[i] = c
			}
			if c > hi[i] {
				hi[i] = c
			}
		}
	}
	return
}

func base(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "bgvdump: "+format+"\n", a...)
	os.Exit(1)
}
