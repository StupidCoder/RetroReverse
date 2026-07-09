// geomprobe is a scratch validator for the nfs decoders: it decodes cy1.trk,
// LaunchMe and LDiablo.wrapFam straight from the disc and checks them against
// facts pinned by the oracle traces (segment positions, slice-row values seen
// in the game's world-vertex accumulator, the runtime-verified car face map).
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"retroreverse.com/games/need-for-speed-3do/extract/nfs"
	"retroreverse.com/tools/platform/threedo"
)

func main() {
	image := flag.String("image", "", "3DO disc image")
	spotDir := flag.String("spots", "", "dump car SPoT textures as PNGs to this dir")
	flag.Parse()

	data, err := os.ReadFile(*image)
	die(err)
	vol, err := threedo.Open(data)
	die(err)

	trk, err := vol.ReadFile("DriveData/tracks/cy1.trk")
	die(err)
	lm, err := vol.ReadFile("LaunchMe")
	die(err)
	fam, err := vol.ReadFile("DriveData/CarData/LDiablo.wrapFam")
	die(err)

	t, err := nfs.ParseTrack(trk)
	die(err)
	fmt.Printf("track: %d segments (%d streamed), %d groups, %d defs, %d placements\n",
		len(t.Segments), t.UsedSegments(), len(t.Groups), len(t.Objects.Defs), len(t.Objects.Placements))

	// pinned facts from the oracle traces
	s := t.Segments[1739]
	fmt.Printf("seg1739 pos=(%.3f, %.3f, %.3f) heading=%d family=%d\n",
		nfs.Float(s.Pos.X), nfs.Float(s.Pos.Y), nfs.Float(s.Pos.Z), s.Heading, s.SliceFamily)
	r := t.Groups[4].Rows[0] // segment 16
	fmt.Printf("seg16 row p0=(%.3f,%.3f,%.3f) p4=(%.3f,%.3f,%.3f) p10=(%.3f,%.3f,%.3f)\n",
		nfs.Float(r[0].X), nfs.Float(r[0].Y), nfs.Float(r[0].Z),
		nfs.Float(r[4].X), nfs.Float(r[4].Y), nfs.Float(r[4].Z),
		nfs.Float(r[10].X), nfs.Float(r[10].Y), nfs.Float(r[10].Z))
	fmt.Printf("group4 materials % x\n", t.Groups[4].Materials)

	// loop closure: last streamed segment row should sit near segment 0's
	last := t.Groups[len(t.Groups)-1].Rows[3]
	fmt.Printf("last row p0=(%.1f,%.1f,%.1f) vs seg0=(%.1f,%.1f,%.1f)\n",
		nfs.Float(last[0].X), nfs.Float(last[0].Y), nfs.Float(last[0].Z),
		nfs.Float(t.Segments[0].Pos.X), nfs.Float(t.Segments[0].Pos.Y), nfs.Float(t.Segments[0].Pos.Z))

	// slice topology
	st, err := nfs.LoadSliceTables(lm)
	die(err)
	fams := map[byte]int{}
	for _, seg := range t.Segments[:t.UsedSegments()] {
		fams[seg.SliceFamily]++
	}
	fmt.Printf("families used: %v\n", fams)
	for fam := range fams {
		typ, err := st.TypeFor(fam)
		if err != nil {
			fmt.Printf("  family %d: ERR %v\n", fam, err)
			continue
		}
		faces := st.Faces(typ)
		fmt.Printf("  family %d -> type %d: %d verts, faces [%d,+%d), rows %d/%d/%d/%d verts, f0=%+v\n",
			fam, typ.ID, typ.VertTotal, typ.FaceStart, typ.FaceCount,
			len(typ.RowVerts[0]), len(typ.RowVerts[1]), len(typ.RowVerts[2]), len(typ.RowVerts[3]), faces[0])
	}

	// object defs census
	types := map[byte]int{}
	for _, d := range t.Objects.Defs {
		types[d.Type]++
	}
	fmt.Printf("def types: %v\n", types)
	usedDefs := map[byte]int{}
	for _, p := range t.Objects.Placements {
		usedDefs[p.Def]++
	}
	fmt.Printf("distinct defs placed: %d, max seg %d\n", len(usedDefs), maxSeg(t.Objects.Placements))

	// car: face map of LOD 1 (DIABL200) must equal the runtime texset
	lods, err := nfs.ParseCarFam(fam)
	die(err)
	for i, l := range lods {
		fmt.Printf("car LOD %d: %q %d verts %d faces %d textures\n",
			i, l.Model.Name, len(l.Model.Verts), len(l.Model.Faces), len(l.Textures))
	}
	truth := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		21, 22, 21, 21, 21, 22, 23, 24, 25, 26, 27, 28, 29, 27, 29}
	got := lods[1].FaceTex[:len(truth)]
	match := true
	for i := range truth {
		if got[i] != truth[i] {
			match = false
			break
		}
	}
	fmt.Printf("car LOD1 face map matches runtime texset: %v\n", match)

	if *spotDir != "" {
		die(os.MkdirAll(*spotDir, 0755))
		for i, img := range lods[2].Textures {
			f, err := os.Create(filepath.Join(*spotDir, fmt.Sprintf("spot-%02d.png", i)))
			die(err)
			die(png.Encode(f, img))
			f.Close()
		}
		fmt.Printf("wrote %d LOD2 SPoT textures to %s\n", len(lods[2].Textures), *spotDir)
	}
}

func maxSeg(ps []nfs.Placement) uint32 {
	var m uint32
	for _, p := range ps {
		if p.Segment > m {
			m = p.Segment
		}
	}
	return m
}

func die(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "geomprobe:", err)
		os.Exit(1)
	}
}
