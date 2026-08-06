// stage.go — the course placement forest: the e0 node tree, the e2 draw-range
// table and the w4 matrix table, wired together by the cs_*_bin per-segment
// visibility database. Decoded from the game's own walker (Part XXIX):
//
//   - 0x135A0 expands a segment's 0xFFFF-terminated uint16 id list, indexing
//     the part's e0 table (part record +0x1C) with a 0x38-byte stride;
//   - 0x13310 walks one node: an optional matrix (e0+0x1C, an index into the
//     part's w4 table of 0x40-byte row-vector matrices, part record +0x10) is
//     mult-pushed on the game's matrix stack — billboard nodes (flag bit 11)
//     additionally get a camera-facing yaw baked in per frame — then the
//     node's e2 entry (+0x28, first valid of four LOD slots; part record
//     +0x20, {firstRange,rangeCount} pairs) selects 0x14-byte range records
//     {pair, first[2], count[2]}, runs of 32-byte batch descriptors for the
//     opaque and blended passes; children (+0x20, then the child's +0x24
//     sibling chain) recurse under the accumulated matrix;
//   - a node without a matrix draws in world space — the plain course
//     geometry is just the matrix-less nodes of the same forest.
//
// The export mirrors that: matrix-less reachable nodes merge into one world
// node per part; nodes with a matrix become glTF instance nodes (mesh shared
// per e2 entry, matrix = the w4 matrix verbatim — glTF's column-major node
// matrix and the engine's row-vector row-major storage are the same 16
// floats). Billboard nodes carry extras {"billboard": "y"}: the game yaws
// them about their local Y axis each frame to face the camera; the export
// bakes the placement matrix and leaves the yaw to the viewer.
package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/build"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/xbox"
)

// inflateFile reads and inflates one .sz off the disc.
func inflateFile(disc *xbox.Image, discPath string) ([]byte, error) {
	raw, err := disc.ReadFile(discPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", discPath, err)
	}
	zr, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: zlib: %w", discPath, err)
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("%s: inflate: %w", discPath, err)
	}
	return data, nil
}

// readStagePMT inflates and parses one pmt off the disc, attaching the
// visibility database at binPath when one is given (the variant dirs name
// theirs in upper case, so callers resolve the actual path from a listing).
func readStagePMT(disc *xbox.Image, discPath, binPath string) (*pmt, []texInfo, error) {
	data, err := inflateFile(disc, discPath)
	if err != nil {
		return nil, nil, err
	}
	base := strings.TrimSuffix(filepath.Base(discPath), "_pmt.sz")
	p, err := parsePMT(base, data)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", discPath, err)
	}
	if binPath != "" {
		bdata, err := inflateFile(disc, binPath)
		if err != nil {
			return nil, nil, err
		}
		v, err := parseVisBin(bdata, p.nParts)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", binPath, err)
		}
		p.vis = v
	}
	texs, _, err := p.parseTextures()
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", discPath, err)
	}
	return p, texs, nil
}

// nodesBounds accumulates the world-space bounding box of a node list,
// applying each node's placement matrix (row-vector) to its positions.
func nodesBounds(nodes []glb.VariantNode) (mn, mx [3]float32) {
	first := true
	for _, n := range nodes {
		for _, p := range n.Positions {
			w := p
			if m := n.Matrix; m != nil {
				w = [3]float32{
					m[0]*p[0] + m[4]*p[1] + m[8]*p[2] + m[12],
					m[1]*p[0] + m[5]*p[1] + m[9]*p[2] + m[13],
					m[2]*p[0] + m[6]*p[1] + m[10]*p[2] + m[14],
				}
			}
			if first {
				mn, mx = w, w
				first = false
				continue
			}
			for k := 0; k < 3; k++ {
				if w[k] < mn[k] {
					mn[k] = w[k]
				}
				if w[k] > mx[k] {
					mx[k] = w[k]
				}
			}
		}
	}
	return
}

// exportStages writes every non-reverse course on the disc as a LEVEL, the
// shape the other racing games ship (Crazy Taxi is the template): the course
// pmt assembled through its visibility database (world geometry + placed
// instances) plus the distant cs_ENV scenery ring in one layer GLB, and the
// sky dome as its own GLB in a camera-attached sky layer — the game draws the
// dome around the camera (zero parallax), and the viewer's attach:"camera" +
// renderOrder -1 + depthTest:false is exactly that. Only the 30 tour courses
// ship: the _R (reversed) dirs re-pack the same worlds with mirrored culling,
// and the openers' _BR/_BT/_T dirs are near-identical start-area re-dressings
// — neither earns its bytes. Display names come from the courses' own fork
// signboards (stageNames), grouped and ordered by route pyramid (stageTours).
func exportStages(disc *xbox.Image, b *build.Builder) {
	// One walk indexes /Stage: dir -> lower-cased file name -> actual path
	// (the variant dirs name their bins in UPPER case).
	idx := map[string]map[string]string{}
	if err := disc.Walk(func(e xbox.Entry) error {
		if e.IsDir || !strings.HasPrefix(e.Path, "/Stage/") {
			return nil
		}
		rest := strings.TrimPrefix(e.Path, "/Stage/")
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) != 2 {
			return nil
		}
		if idx[parts[0]] == nil {
			idx[parts[0]] = map[string]string{}
		}
		idx[parts[0]][strings.ToLower(parts[1])] = e.Path
		return nil
	}); err != nil {
		fatal("walk /Stage: %v", err)
	}
	var dirs []string
	for d := range idx {
		// Only the 30 tour courses ship: the _R dirs re-pack the same worlds
		// reversed, and the openers' _BR/_BT/_T dirs are near-identical
		// start-area re-dressings (see the markdown) — neither earns its
		// bytes in the tree.
		if strings.Contains(d, "_") {
			continue
		}
		if idx[d][strings.ToLower("cs_CS_"+d+"_pmt.sz")] == "" {
			continue
		}
		dirs = append(dirs, d)
	}
	sort.Slice(dirs, func(i, j int) bool { return stageSortKey(dirs[i]) < stageSortKey(dirs[j]) })
	skyWritten := map[string]string{} // sky disc path -> GLB file name in levels/
	for _, d := range dirs {
		if err := exportStage(disc, b, d, idx, skyWritten); err != nil {
			fmt.Fprintf(os.Stderr, "carex: stage %s: %v (skipped)\n", d, err)
		}
	}
}

// stageNames are the game's own course names, transcribed from each course's
// fork-signboard texture (/Common/obj_obj_<code>_bunki_tx_pmt.sz, texture 1 —
// the green "next course" arrow sign the gantries carry), with word spacing
// normalised to the transition-banner style. The two 15-course tours these
// belong to are the game's OutRun2 and OutRun2 SP route pyramids (the
// co-driver voice packs label them HAM_<stage> and HAM_CVT_<stage>).
var stageNames = map[string]string{
	"ALAS": "Ice Scape", "ALPI": "Alpine", "AMAZ": "Jungle",
	"BEAC": "Sunny Beach", "CAPE": "Cape Way", "CAST": "Castle Wall",
	"CLOU": "Cloudy Highland", "DESE": "Desert", "EAST": "Giant Statues",
	"FLOR": "Milky Way", "FORE": "Coniferous Forest", "GHOS": "Ghost Forest",
	"GRAN": "Canyon", "IMPE": "Imperial Avenue", "INDU": "Industrial Complex",
	"LAKE": "Deep Lake", "LASV": "Casino Town", "MACH": "Lost City",
	"MAYA": "Legend", "METR": "Metropolis", "NEWY": "Skyscrapers",
	"NIAG": "Water Falls", "PALM": "Palm Beach", "PRIN": "Floral Village",
	"RUIN": "Ancient Ruins", "SANF": "Bay Area", "SEQU": "Big Forest",
	"SNOW": "Snow Mountain", "TULI": "Tulip Garden", "YOSE": "National Park",
}

// stageTours places each course in its route pyramid: tour and stage slot
// (round 1-5, column A-E), from the co-driver voice packs — the disc names
// them /Sound/*_HAM_<slot>.PAK for the OutRun2 tour and *_HAM_CVT_<slot>.PAK
// for the OutRun2 SP tour, and each pack's WBND header labels its course
// ("HAM Req: Tulip"), matching the signboard names (3B's label "Carnif" is
// the developers' own spelling of the coniferous forest, forced anyway by
// elimination).
var stageTours = map[string]struct{ tour, slot string }{
	"PALM": {"OutRun2", "1A"}, "LAKE": {"OutRun2", "2A"}, "ALPI": {"OutRun2", "2B"},
	"CAST": {"OutRun2", "3A"}, "FORE": {"OutRun2", "3B"}, "DESE": {"OutRun2", "3C"},
	"CLOU": {"OutRun2", "4A"}, "INDU": {"OutRun2", "4B"}, "SNOW": {"OutRun2", "4C"},
	"GHOS": {"OutRun2", "4D"}, "TULI": {"OutRun2", "5A"}, "METR": {"OutRun2", "5B"},
	"RUIN": {"OutRun2", "5C"}, "IMPE": {"OutRun2", "5D"}, "CAPE": {"OutRun2", "5E"},

	"BEAC": {"OutRun2: SP", "1A"}, "SANF": {"OutRun2: SP", "2A"}, "YOSE": {"OutRun2: SP", "2B"},
	"NIAG": {"OutRun2: SP", "3A"}, "SEQU": {"OutRun2: SP", "3B"}, "GRAN": {"OutRun2: SP", "3C"},
	"MACH": {"OutRun2: SP", "4A"}, "LASV": {"OutRun2: SP", "4B"}, "ALAS": {"OutRun2: SP", "4C"},
	"AMAZ": {"OutRun2: SP", "4D"}, "EAST": {"OutRun2: SP", "5A"}, "MAYA": {"OutRun2: SP", "5B"},
	"FLOR": {"OutRun2: SP", "5C"}, "PRIN": {"OutRun2: SP", "5D"}, "NEWY": {"OutRun2: SP", "5E"},
}

// stageSortKey orders the manifest: the OutRun2 pyramid in stage order, then
// the SP pyramid.
func stageSortKey(dir string) string {
	t, ok := stageTours[dir]
	if !ok {
		return "9" + dir
	}
	tourOrd := "0"
	if t.tour != "OutRun2" {
		tourOrd = "1"
	}
	return tourOrd + t.slot
}

// stageID is the asset id for a course dir ("BEAC_BR" -> "stage-beac-br").
func stageID(dir string) string {
	return "stage-" + strings.ReplaceAll(strings.ToLower(dir), "_", "-")
}

// exportStage writes one course dir as a level.
func exportStage(disc *xbox.Image, b *build.Builder, dir string, idx map[string]map[string]string, skyWritten map[string]string) error {
	files := idx[dir]
	lc := strings.ToLower(dir)
	id := stageID(dir)
	csPmt := files["cs_cs_"+lc+"_pmt.sz"]
	csBin := files["cs_cs_"+lc+"_bin.sz"]
	if csBin == "" {
		return fmt.Errorf("no visibility db next to %s", csPmt)
	}
	course, courseTexs, err := readStagePMT(disc, csPmt, csBin)
	if err != nil {
		return err
	}
	nodes, objMeshes, placements, summary, err := course.buildStageNodes(courseTexs, true)
	if err != nil {
		return err
	}

	out, err := b.Path("levels", id+".glb")
	if err != nil {
		return err
	}
	if err := glb.WriteVariantScenes(out, []glb.ModelVariant{{Name: "course", Nodes: nodes}}); err != nil {
		return err
	}

	// The distant scenery model, in its own layer: cs_ENV is the far-LOD
	// WORLD — a low-poly replica of the whole course (Desert's five
	// course-sized meshes share course part 0's bounding sphere bit for bit)
	// plus the beyond-the-track scenery, and its visibility db lists every
	// node from every segment: the game draws it under the detailed course
	// as a permanent backdrop, letting the near geometry cover it. The
	// viewer reproduces that with a depth-offset layer (polygonOffset pushes
	// it behind; renderOrder draws it first), which is what stops the
	// replica z-fighting the real road — no geometry is dropped.
	envGLB := ""
	envSummary := "none"
	if envPmt := files["cs_env_"+lc+"_pmt.sz"]; envPmt != "" {
		env, envTexs, err := readStagePMT(disc, envPmt, files["cs_env_"+lc+"_bin.sz"])
		if err != nil {
			return err
		}
		var envNodes []glb.VariantNode
		if env.vis != nil {
			envNodes, _, _, envSummary, err = env.buildStageNodes(envTexs, false)
			if err != nil {
				return err
			}
		} else {
			plan, err := env.plan()
			if err != nil {
				return err
			}
			v, s, err := buildVariant(env, envTexs, plan)
			if err != nil {
				return err
			}
			envSummary = s + " (flat)"
			envNodes = v.Nodes
			if len(envNodes) == 0 {
				envNodes = []glb.VariantNode{{Positions: v.Positions, Normals: v.Normals,
					UVs: v.UVs, UV2: v.UV2, Colors: v.Colors,
					TexGroups: v.TexGroups, ColorGroups: v.ColorGroups}}
			}
		}
		for i := range envNodes {
			envNodes[i].Name = strings.TrimSpace("env " + envNodes[i].Name)
		}
		envGLB = id + "-env.glb"
		envOut, err := b.Path("levels", envGLB)
		if err != nil {
			return err
		}
		if err := glb.WriteVariantScenes(envOut, []glb.ModelVariant{{Name: "env", Nodes: envNodes}}); err != nil {
			return err
		}
	}

	// The sky dome: the dir's own, or the base course's for the _BR/_BT/_T
	// variants (they ship none of their own). Written once per source file;
	// variants reference the base's GLB.
	skyGLB := ""
	skyPath := files["obj_course_obj_sky_"+lc+"_pmt.sz"]
	if skyPath != "" {
		if have, ok := skyWritten[skyPath]; ok {
			skyGLB = have
		} else {
			sky, skyTexs, err := readStagePMT(disc, skyPath, "")
			if err != nil {
				return err
			}
			skyPlan, err := sky.plan()
			if err != nil {
				return err
			}
			skyV, _, err := buildVariant(sky, skyTexs, skyPlan)
			if err != nil {
				return err
			}
			skyV.Name = "sky"
			name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(skyPath), "obj_course_obj_sky_"), "_pmt.sz")
			skyGLB = "stage-" + strings.ToLower(name) + "-sky.glb"
			skyOut, err := b.Path("levels", skyGLB)
			if err != nil {
				return err
			}
			if err := glb.WriteVariantScenes(skyOut, []glb.ModelVariant{skyV}); err != nil {
				return err
			}
			skyWritten[skyPath] = skyGLB
		}
	}

	// The course's environment cube feeds the sheen-marked materials (the
	// sea's per-pixel reflection).
	var envMap []string
	if faces := envFaces(course, courseTexs); faces != nil {
		for fi, img := range faces {
			fn := fmt.Sprintf("%s-env-%s.png", id, [6]string{"px", "nx", "py", "ny", "pz", "nz"}[fi])
			path, err := b.Path("levels", fn)
			if err != nil {
				return err
			}
			writePNG(path, img)
			envMap = append(envMap, fn)
		}
	}

	// The placed props become real objects + level placements, so the
	// viewer's click handler works on them (name card + "Open object").
	name := strings.ReplaceAll(dir, "_", " ")
	if n, ok := stageNames[dir]; ok {
		name = n
	}
	var keys []string
	for k := range objMeshes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	objAsset := map[string]string{}
	for _, k := range keys {
		vn := objMeshes[k]
		var opi, oe2 int
		fmt.Sscanf(k, "p%d-e2-%d", &opi, &oe2)
		objID := fmt.Sprintf("%s-p%d-%d", id, opi, oe2)
		objName := fmt.Sprintf("%s: object %d.%d", name, opi, oe2)
		fn := objID + ".glb"
		path, err := b.Path("objects", fn)
		if err != nil {
			return err
		}
		mv := glb.ModelVariant{Positions: vn.Positions, Normals: vn.Normals, UVs: vn.UVs,
			UV2: vn.UV2, Colors: vn.Colors, TexGroups: vn.TexGroups, ColorGroups: vn.ColorGroups}
		if err := glb.WriteVariantScenes(path, []glb.ModelVariant{mv}); err != nil {
			return err
		}
		b.AddObject(schema.Asset{ID: objID, Name: objName, Group: "Course objects"},
			&schema.Object{Type: schema.ObjectModel3D, Name: objName, Model: fn,
				Props: map[string]any{"source": csPmt}})
		objAsset[k] = objID
	}
	var pls []schema.Placement
	for i, pl := range placements {
		m := make([]float64, 16)
		for k, v := range pl.mat {
			m[k] = float64(v)
		}
		pls = append(pls, schema.Placement{
			ID: i, Object: objAsset[pl.key], Matrix: m,
			Props: map[string]any{"node": fmt.Sprintf("part %d id %d", pl.part, pl.id), "w4": pl.matrixIdx},
		})
	}

	// Camera: on the start straight, looking down the course (the road runs
	// -z from the start line); range from the assembled world's own bounds.
	mn, mx := nodesBounds(nodes)
	dx, dy, dz := float64(mx[0]-mn[0]), float64(mx[1]-mn[1]), float64(mx[2]-mn[2])
	diag := math.Sqrt(dx*dx + dy*dy + dz*dz)
	depthOff := false
	// Both scenery layers are toggles: the far-world replica is off by
	// default (it exists to be covered by the near detail; on its own it
	// clashes), and the course itself can be switched off to inspect it.
	envOff := false
	layers := []schema.Layer{{ID: "course", Name: "Course", File: id + ".glb", Mode: "toggle", EnvMap: envMap}}
	if envGLB != "" {
		layers = append(layers, schema.Layer{ID: "env", Name: "Distant scenery", File: envGLB,
			Mode: "toggle", Visible: &envOff, RenderOrder: -0.5, PolygonOffset: 4})
	}
	if skyGLB != "" {
		layers = append(layers, schema.Layer{ID: "sky", Name: "Sky", File: skyGLB, Mode: "toggle",
			Attach: "camera", RenderOrder: -1, DepthTest: &depthOff, Role: "sky"})
	}
	group := "Courses"
	if t, ok := stageTours[dir]; ok {
		group = t.tour
		name = "Stage " + t.slot + ": " + name
	}
	b.AddLevel(schema.Asset{ID: id, Name: name, Group: group}, &schema.Level{
		Type: schema.LevelScene3D,
		Camera: &schema.Camera{
			Mode:   "fly",
			Pos:    []float64{0, 5, -6},
			Target: []float64{0, 3, -120},
			FOV:    55,
			Near:   0.5,
			Far:    1.3 * diag,
			Fly:    &schema.Fly{Speed: diag / 60},
		},
		Scene:      &schema.Scene{Layers: layers},
		Placements: pls,
	})
	fmt.Printf("%-34s -> %s (%s; env %s; sky %s)\n", csPmt, out, summary, envSummary, skyGLB)
	return nil
}
// loadVisBin attaches a course's cs_*_bin visibility database when the disc
// carries one next to its pmt; non-course models are left alone.
func loadVisBin(disc *xbox.Image, discPath string, p *pmt) error {
	if !(strings.Contains(discPath, "/cs_CS_") || strings.Contains(discPath, "/cs_ENV_")) ||
		!strings.HasSuffix(discPath, "_pmt.sz") {
		return nil
	}
	binPath := strings.Replace(discPath, "_pmt.sz", "_bin.sz", 1)
	raw, err := disc.ReadFile(binPath)
	if err != nil {
		return fmt.Errorf("visibility db %s: %w", binPath, err)
	}
	zr, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("%s: zlib: %w", binPath, err)
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		return fmt.Errorf("%s: inflate: %w", binPath, err)
	}
	v, err := parseVisBin(data, p.nParts)
	if err != nil {
		return fmt.Errorf("%s: %w", binPath, err)
	}
	p.vis = v
	return nil
}

// e0Node is one 0x38-byte placement-forest node.
type e0Node struct {
	flags     uint32     // bit0/1 pass-ish kind bits, bit2 has-matrix, bit3 instance, bit6 render-state group, bit11 billboard
	sphere    [4]float32 // bounding sphere, node-local when placed
	f14       uint32     // undecoded (+0x14)
	f18       int32      // undecoded (+0x18)
	matrixIdx int32      // into the part's w4 table, -1 = none (world space)
	child     int32      // first child node id, -1 = none
	sibling   int32      // next sibling node id, -1 = end
	e2        [4]int32   // e2 entry per LOD slot, -1 = none
}

// stagePart is one part's placement tables.
type stagePart struct {
	nodes  []e0Node
	e2     [][2]uint32 // {firstRange, rangeCount}
	ranges [][5]uint32 // {pair, firstDesc[2], descCount[2]} — [1]/[3] opaque pass, [2]/[4] blended
	w4Off  int         // section-A offset of the matrix table
	w4Max  int         // matrices that fit between w4Off and end of section A (shared tables have no own count)
}

func i32(b []byte, off int) int32 { return int32(u32(b, off)) }

// parsePlacements reads part pi's placement tables from the entry header's
// offset table (e[0]=e0, e[1]=w4, e[2]=e2, e[3]=ranges, e[4]=descriptors —
// consecutive, so the spans are the counts).
func (p *pmt) parsePlacements(pi int) (*stagePart, error) {
	rec := 0x18 + pi*0x3C
	ent := int(u32(p.a, rec+0x14))
	var e [6]int
	for i := range e {
		e[i] = int(u32(p.a, ent+4*i))
	}
	sp := &stagePart{w4Off: e[1], w4Max: (len(p.a) - e[1]) / 0x40}
	spans := [3]struct {
		name        string
		off, stride int
		end         int
	}{
		{"e0", e[0], 0x38, e[2]},
		{"e2", e[2], 8, e[3]},
		{"ranges", e[3], 0x14, e[4]},
	}
	for _, s := range spans {
		if s.end < s.off || (s.end-s.off)%s.stride != 0 {
			return nil, fmt.Errorf("part %d: %s span %#x..%#x not a multiple of %#x", pi, s.name, s.off, s.end, s.stride)
		}
	}
	for off := e[0]; off < e[2]; off += 0x38 {
		n := e0Node{
			flags:     u32(p.a, off),
			f14:       u32(p.a, off+0x14),
			f18:       i32(p.a, off+0x18),
			matrixIdx: i32(p.a, off+0x1C),
			child:     i32(p.a, off+0x20),
			sibling:   i32(p.a, off+0x24),
		}
		for k := 0; k < 4; k++ {
			n.sphere[k] = f32(p.a, off+4+4*k)
			n.e2[k] = i32(p.a, off+0x28+4*k)
		}
		sp.nodes = append(sp.nodes, n)
	}
	for off := e[2]; off < e[3]; off += 8 {
		sp.e2 = append(sp.e2, [2]uint32{u32(p.a, off), u32(p.a, off+4)})
	}
	for off := e[3]; off < e[4]; off += 0x14 {
		var r [5]uint32
		for k := range r {
			r[k] = u32(p.a, off+4*k)
		}
		sp.ranges = append(sp.ranges, r)
	}
	return sp, nil
}

// matrix reads w4 matrix mi: 16 f32, rows of a row-vector transform (scale in
// the rotation rows, translation in row 3).
func (sp *stagePart) matrix(p *pmt, mi int) ([16]float32, error) {
	var m [16]float32
	if mi < 0 || mi >= sp.w4Max {
		return m, fmt.Errorf("matrix %d out of table (max %d)", mi, sp.w4Max)
	}
	for k := 0; k < 16; k++ {
		m[k] = f32(p.a, sp.w4Off+mi*0x40+4*k)
	}
	return m, nil
}

// mulRowVec composes row-vector matrices: v·a·b, i.e. a applied first.
func mulRowVec(a, b [16]float32) [16]float32 {
	var r [16]float32
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var s float32
			for k := 0; k < 4; k++ {
				s += a[i*4+k] * b[k*4+j]
			}
			r[i*4+j] = s
		}
	}
	return r
}

// visBin is the decoded cs_*_bin visibility database: per part, the union of
// node ids any of the course's 8-metre segments lists, and per node the
// segments that list it — the game's own record of where each version of an
// object is meant to be seen (the LOD system lives here, not in the nodes).
type visBin struct {
	nSeg  int
	roots [][]int
	segs  []map[int][]int // per part: node id -> listing segments
}

// parseVisBin opens the decompressed cs_*_bin: {size, 0, 0} then table
// offsets (payload-relative, payload = file+4), one table per part; each
// table is nSeg record offsets (table-relative, so record 0's offset says how
// many) followed by 0xFFFF-terminated uint16 id lists.
func parseVisBin(data []byte, nParts int) (*visBin, error) {
	if len(data) < 0x24 {
		return nil, fmt.Errorf("bin: short file (%d bytes)", len(data))
	}
	pay := data[4:]
	if u32(pay, 0) != 0 || u32(pay, 4) != 0 {
		return nil, fmt.Errorf("bin: header not {0,0,...}")
	}
	first := int(u32(pay, 8))
	nTab := (first - 8) / 4
	if nTab != nParts {
		return nil, fmt.Errorf("bin: %d tables for %d parts", nTab, nParts)
	}
	v := &visBin{}
	for t := 0; t < nTab; t++ {
		base := int(u32(pay, 8+4*t))
		nSeg := int(u32(pay, base)) / 4
		if t == 0 {
			v.nSeg = nSeg
		} else if nSeg != v.nSeg {
			return nil, fmt.Errorf("bin: table %d has %d segments, table 0 has %d", t, nSeg, v.nSeg)
		}
		seen := map[int][]int{}
		for s := 0; s < nSeg; s++ {
			ro := base + int(u32(pay, base+4*s))
			for {
				if ro+2 > len(pay) {
					return nil, fmt.Errorf("bin: table %d segment %d runs off the file", t, s)
				}
				id := u16(pay, ro)
				ro += 2
				if id == 0xFFFF {
					break
				}
				seen[int(id)] = append(seen[int(id)], s)
			}
		}
		ids := make([]int, 0, len(seen))
		for id := range seen {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		v.roots = append(v.roots, ids)
		v.segs = append(v.segs, seen)
	}
	return v, nil
}

// worldCentre is a node's bounding-sphere centre in world space (placed
// nodes carry their sphere in node-local space), with the matrix's scale
// applied to the radius.
func (sp *stagePart) worldSphere(p *pmt, n *e0Node) ([3]float32, float32) {
	c := [3]float32{n.sphere[0], n.sphere[1], n.sphere[2]}
	r := n.sphere[3]
	if n.matrixIdx < 0 {
		return c, r
	}
	m, err := sp.matrix(p, int(n.matrixIdx))
	if err != nil {
		return c, r
	}
	w := [3]float32{
		c[0]*m[0] + c[1]*m[4] + c[2]*m[8] + m[12],
		c[0]*m[1] + c[1]*m[5] + c[2]*m[9] + m[13],
		c[0]*m[2] + c[1]*m[6] + c[2]*m[10] + m[14],
	}
	scale := float32(math.Sqrt(float64(m[0]*m[0] + m[1]*m[1] + m[2]*m[2])))
	if scale > 0 {
		r *= scale
	}
	return w, r
}

// stagePlace is one placed prop extracted for a level placement: the mesh
// key, the w4 placement matrix (column-major, verbatim), and its forest ids
// for the info card.
type stagePlace struct {
	key       string
	mat       [16]float32
	part, id  int
	matrixIdx int
}

// stageInstance is one placed drawing of an e2 entry.
type stageInstance struct {
	nodeID    int
	mat       [16]float32
	billboard bool
}

// buildStageNodes assembles the course as the visibility walker draws it:
// per part, one world node holding every batch a matrix-less reachable
// forest node references, plus one instance node per placed node, sharing
// meshes per e2 entry.
// buildStageNodes assembles the drawable forest. With extractPlacements set,
// non-billboard placed instances come back as shared meshes + placements
// (for level placements the viewer can click); billboard instances always
// stay in the returned nodes, where the layer's yaw updater animates them.
func (p *pmt) buildStageNodes(texs []texInfo, extractPlacements bool) ([]glb.VariantNode, map[string]*glb.VariantNode, []stagePlace, string, error) {
	objMeshes := map[string]*glb.VariantNode{}
	var placements []stagePlace
	var out []glb.VariantNode
	var nWorldDesc, nInst, nMesh, nBB, nUnref, nPairFix int
	// Two passes: all parts parse first so the segment-position estimate (and
	// the LOD filter it feeds) sees the whole course; a caller with its own
	// estimate (the env ring shares the course's segmentation) passes it in.
	parts := make([]*part, p.nParts)
	places := make([]*stagePart, p.nParts)
	for pi := 0; pi < p.nParts; pi++ {
		pt, err := p.parsePart(pi)
		if err != nil {
			return nil, nil, nil, "", fmt.Errorf("part %d: %w", pi, err)
		}
		sp, err := p.parsePlacements(pi)
		if err != nil {
			return nil, nil, nil, "", err
		}
		parts[pi], places[pi] = pt, sp
	}

	for pi := 0; pi < p.nParts; pi++ {
		pt, sp := parts[pi], places[pi]

		// The range records carry each descriptor's stream pair straight from
		// the file — authoritative over the stream-cursor reconstruction,
		// whose tie-breaks go wrong on a handful of descriptors (NEWY: 104
		// entries). Apply them before any geometry is cut.
		for _, rec := range sp.ranges {
			for pass := 0; pass < 2; pass++ {
				f, c := rec[1+pass], rec[3+pass]
				for d := f; d < f+c && int(d) < len(pt.descStart); d++ {
					for k := pt.descStart[d]; k < pt.descStart[d]+pt.descCount[d]; k++ {
						if pt.batches[k].pair != int(rec[0]) {
							pt.batches[k].pair = int(rec[0])
							nPairFix++
						}
					}
				}
			}
		}

		// Decode all pairs into one merged vertex pool (as buildVariant does),
		// so batch tris index one address space; instance meshes compact out
		// of it at the end.
		pool := newStagePool(p, pt)

		// e2 entry -> descriptor list (both passes), with the range records'
		// own pair field checked against the assignment the stream walk made.
		descsOf := func(e2i int) ([]int, error) {
			pr := sp.e2[e2i]
			var descs []int
			for r := pr[0]; r < pr[0]+pr[1]; r++ {
				if int(r) >= len(sp.ranges) {
					return nil, fmt.Errorf("part %d e2 %d: range %d out of %d", pi, e2i, r, len(sp.ranges))
				}
				rec := sp.ranges[r]
				for pass := 0; pass < 2; pass++ {
					f, c := rec[1+pass], rec[3+pass]
					for d := f; d < f+c; d++ {
						if int(d) >= len(pt.descStart) {
							return nil, fmt.Errorf("part %d e2 %d: descriptor %d out of %d", pi, e2i, d, len(pt.descStart))
						}
						descs = append(descs, int(d))
					}
				}
			}
			return descs, nil
		}

		// Walk the forest from the visibility roots. A listed root draws
		// without its siblings (the walker's own trick); children recurse
		// under the accumulated matrix.
		worldDesc := map[int]bool{}
		var instances []stageInstance
		instMesh := map[int][]int{} // e2 entry -> descriptor list (resolved once)
		budget := len(sp.nodes) * 8
		var walkErr error
		var walk func(id int, parent *[16]float32)
		walk = func(id int, parent *[16]float32) {
			if walkErr != nil {
				return
			}
			if budget--; budget < 0 {
				walkErr = fmt.Errorf("part %d: forest walk exceeded budget (cycle?)", pi)
				return
			}
			if id < 0 || id >= len(sp.nodes) {
				walkErr = fmt.Errorf("part %d: node id %d out of %d", pi, id, len(sp.nodes))
				return
			}
			n := &sp.nodes[id]
			mat := parent
			if n.matrixIdx >= 0 {
				m, err := sp.matrix(p, int(n.matrixIdx))
				if err != nil {
					walkErr = fmt.Errorf("part %d node %d: %w", pi, id, err)
					return
				}
				if parent != nil {
					m = mulRowVec(m, *parent)
				}
				mat = &m
			}
			e2i := -1
			for k := 0; k < 4; k++ {
				if n.e2[k] >= 0 {
					e2i = int(n.e2[k])
					break
				}
			}
			if e2i >= 0 {
				if _, ok := instMesh[e2i]; !ok && mat != nil {
					descs, err := descsOf(e2i)
					if err != nil {
						walkErr = err
						return
					}
					instMesh[e2i] = descs
				}
				if mat == nil {
					descs, err := descsOf(e2i)
					if err != nil {
						walkErr = err
						return
					}
					for _, d := range descs {
						worldDesc[d] = true
					}
				} else {
					instances = append(instances, stageInstance{nodeID: id, mat: *mat, billboard: n.flags&0x800 != 0})
				}
			}
			for c := n.child; c >= 0; c = sp.nodes[c].sibling {
				walk(int(c), mat)
				if walkErr != nil {
					return
				}
			}
		}
		for _, id := range p.vis.roots[pi] {
			walk(id, nil)
		}
		if walkErr != nil {
			return nil, nil, nil, "", walkErr
		}

		// The world node: every identity-reachable descriptor's batches.
		descs := make([]int, 0, len(worldDesc))
		for d := range worldDesc {
			descs = append(descs, d)
		}
		sort.Ints(descs)
		refd := map[int]bool{}
		for _, d := range descs {
			refd[d] = true
		}
		for _, il := range instMesh {
			for _, d := range il {
				refd[d] = true
			}
		}
		nUnref += len(pt.descStart) - len(refd)
		nWorldDesc += len(descs)
		wn := pool.node(fmt.Sprintf("part %d", pi), descs, texs)
		if wn != nil {
			out = append(out, *wn)
		}

		// Instance nodes, meshes shared per e2 entry. Mesh geometry stays in
		// node-local space; the matrix rides on the node.
		e2Sorted := make([]int, 0, len(instMesh))
		for e2i := range instMesh {
			e2Sorted = append(e2Sorted, e2i)
		}
		sort.Ints(e2Sorted)
		meshDone := map[string]bool{}
		for _, inst := range instances {
			n := &sp.nodes[inst.nodeID]
			e2i := -1
			for k := 0; k < 4; k++ {
				if n.e2[k] >= 0 {
					e2i = int(n.e2[k])
					break
				}
			}
			key := fmt.Sprintf("p%d-e2-%d", pi, e2i)
			if extractPlacements && !inst.billboard {
				if _, ok := objMeshes[key]; !ok {
					vn := pool.node(fmt.Sprintf("part %d e2 %d", pi, e2i), instMesh[e2i], texs)
					if vn == nil {
						continue
					}
					objMeshes[key] = vn
				}
				placements = append(placements, stagePlace{key: key, mat: inst.mat,
					part: pi, id: inst.nodeID, matrixIdx: int(n.matrixIdx)})
				meshDone[key] = true
				nInst++
				continue
			}
			vn := pool.node(fmt.Sprintf("part %d obj %d", pi, inst.nodeID), instMesh[e2i], texs)
			if vn == nil {
				continue
			}
			if meshDone[key] {
				// later instances carry no geometry of their own
				vn.Positions, vn.Normals, vn.UVs, vn.UV2, vn.Colors = nil, nil, nil, nil, nil
				vn.TexGroups, vn.ColorGroups = nil, nil
			}
			meshDone[key] = true
			vn.MeshKey = key
			m := inst.mat
			vn.Matrix = &m
			if inst.billboard {
				vn.Extras = map[string]any{"billboard": "y"}
				nBB++
			}
			out = append(out, *vn)
			nInst++
		}
		nMesh += len(meshDone)
	}
	summary := fmt.Sprintf("%d world descriptors, %d instances of %d meshes (%d billboard)",
		nWorldDesc, nInst, nMesh, nBB)
	if nUnref > 0 {
		summary += fmt.Sprintf(", %d descriptors unreferenced", nUnref)
	}
	if nPairFix > 0 {
		summary += fmt.Sprintf(", %d pair overrides", nPairFix)
	}
	return out, objMeshes, placements, summary, nil
}

// stagePool holds a part's decoded pairs merged into one vertex pool and
// cuts compacted VariantNodes out of it per descriptor list.
type stagePool struct {
	p         *pmt
	pt        *part
	vbase     []uint32
	positions [][3]float32
	normals   [][3]float32
	uvs       [][2]float32
	uv2s      [][2]float32
	colors    [][4]uint8
	hasUV2    bool
	hasColor  bool
}

func newStagePool(p *pmt, pt *part) *stagePool {
	sp := &stagePool{p: p, pt: pt}
	for _, bp := range pt.pairs {
		sp.vbase = append(sp.vbase, uint32(len(sp.positions)))
		pos, nrm, uv, uv2, col := p.decodeVerts(bp)
		sp.positions = append(sp.positions, pos...)
		sp.normals = append(sp.normals, nrm...)
		sp.uvs = append(sp.uvs, uv...)
		if uv2 == nil {
			uv2 = make([][2]float32, len(pos))
		} else {
			sp.hasUV2 = true
		}
		if col == nil {
			col = make([][4]uint8, len(pos))
			for i := range col {
				col[i] = [4]uint8{255, 255, 255, 255}
			}
		} else {
			sp.hasColor = true
		}
		sp.uv2s = append(sp.uv2s, uv2...)
		sp.colors = append(sp.colors, col...)
	}
	return sp
}

// node builds a compacted VariantNode from the batches of the given
// descriptors, grouped by material exactly as buildVariant groups them.
// Returns nil when the descriptor list is empty.
func (sp *stagePool) node(name string, descs []int, texs []texInfo) *glb.VariantNode {
	type texKey struct {
		tex             int
		additive, sheen bool
		wrapS, wrapT    int
	}
	type colKey struct {
		rgba            [4]int
		additive, sheen bool
	}
	texTris := map[texKey][][3]uint32{}
	colorTris := map[colKey][][3]uint32{}
	p, pt := sp.p, sp.pt
	for _, d := range descs {
		for k := pt.descStart[d]; k < pt.descStart[d]+pt.descCount[d]; k++ {
			b := pt.batches[k]
			bp := pt.pairs[b.pair]
			idxCount, _ := indexCount(b.prim, b.prims)
			raw := make([]uint32, idxCount)
			nVerts := bp.vbBytes / bp.stride
			for i := range raw {
				ix := uint32(u16(p.a, int(bp.ibOff)+int(b.first+uint32(i))*2)) + b.baseVtx
				if ix >= nVerts {
					panic(fmt.Sprintf("stage batch %d: index %d out of pair %d VB (%d verts)", k, ix, b.pair, nVerts))
				}
				raw[i] = ix + sp.vbase[b.pair]
			}
			tris := triangulate(b.prim, raw)
			m := pt.mats[b.matIdx]
			if m.texIdx >= 0 && texs[m.texIdx].img != nil && !texs[m.texIdx].cube {
				key := texKey{m.texIdx, m.additive, m.sheen, m.wrapS, m.wrapT}
				texTris[key] = append(texTris[key], tris...)
			} else {
				key := colKey{[4]int{int(m.diffuse[0] * 255), int(m.diffuse[1] * 255), int(m.diffuse[2] * 255), int(m.alpha * 255)}, m.additive, m.sheen}
				colorTris[key] = append(colorTris[key], tris...)
			}
		}
	}
	if len(texTris) == 0 && len(colorTris) == 0 {
		return nil
	}

	// Compact: only the vertices the tris use, remapped.
	remap := map[uint32]uint32{}
	var order []uint32
	use := func(v uint32) uint32 {
		if n, ok := remap[v]; ok {
			return n
		}
		n := uint32(len(order))
		remap[v] = n
		order = append(order, v)
		return n
	}
	remapTris := func(tris [][3]uint32) [][3]uint32 {
		out := make([][3]uint32, len(tris))
		for i, t := range tris {
			out[i] = [3]uint32{use(t[0]), use(t[1]), use(t[2])}
		}
		return out
	}

	tkeys := make([]texKey, 0, len(texTris))
	for k := range texTris {
		tkeys = append(tkeys, k)
	}
	sort.Slice(tkeys, func(i, j int) bool {
		a, b := tkeys[i], tkeys[j]
		if a.tex != b.tex {
			return a.tex < b.tex
		}
		if a.wrapS != b.wrapS {
			return a.wrapS < b.wrapS
		}
		if a.wrapT != b.wrapT {
			return a.wrapT < b.wrapT
		}
		if a.additive != b.additive {
			return !a.additive
		}
		return !a.sheen && b.sheen
	})
	var texGroups []glb.TexturedGroup
	for _, k := range tkeys {
		texGroups = append(texGroups, glb.TexturedGroup{
			Tris: remapTris(texTris[k]), Image: texs[k.tex].img, WrapS: k.wrapS, WrapT: k.wrapT,
			Additive: k.additive, Sheen: k.sheen,
			// The game's own alpha test: ref 0x01 (see buildVariant).
			AlphaCutoff: 1.0 / 255,
		})
	}
	ckeys := make([]colKey, 0, len(colorTris))
	for k := range colorTris {
		ckeys = append(ckeys, k)
	}
	sort.Slice(ckeys, func(i, j int) bool {
		a, b := ckeys[i], ckeys[j]
		if a.rgba != b.rgba {
			for c := 0; c < 4; c++ {
				if a.rgba[c] != b.rgba[c] {
					return a.rgba[c] < b.rgba[c]
				}
			}
		}
		if a.additive != b.additive {
			return !a.additive
		}
		return !a.sheen && b.sheen
	})
	var colorGroups []glb.TriGroup
	for _, k := range ckeys {
		colorGroups = append(colorGroups, glb.TriGroup{
			Tris:     remapTris(colorTris[k]),
			Color:    [3]float32{float32(k.rgba[0]) / 255, float32(k.rgba[1]) / 255, float32(k.rgba[2]) / 255},
			Alpha:    float32(k.rgba[3]) / 255,
			Additive: k.additive,
			Sheen:    k.sheen,
		})
	}

	n := &glb.VariantNode{Name: name, TexGroups: texGroups, ColorGroups: colorGroups}
	n.Positions = make([][3]float32, len(order))
	n.Normals = make([][3]float32, len(order))
	for i, v := range order {
		n.Positions[i] = sp.positions[v]
		n.Normals[i] = sp.normals[v]
	}
	n.UVs = make([][2]float32, len(order))
	for i, v := range order {
		n.UVs[i] = sp.uvs[v]
	}
	if sp.hasUV2 {
		n.UV2 = make([][2]float32, len(order))
		for i, v := range order {
			n.UV2[i] = sp.uv2s[v]
		}
	}
	if sp.hasColor {
		n.Colors = make([][4]uint8, len(order))
		for i, v := range order {
			n.Colors[i] = sp.colors[v]
		}
	}
	return n
}

// inspectForest prints a course's placement forest — every e0 node with its
// flags, LOD e2 slots, matrix translation and the two undecoded fields — the
// diagnostic that pinned the LOD-slot semantics.
func (p *pmt) inspectForest() error {
	for pi := 0; pi < p.nParts; pi++ {
		sp, err := p.parsePlacements(pi)
		if err != nil {
			return err
		}
		pt, err := p.parsePart(pi)
		if err != nil {
			return err
		}
		// tris/textures per e2 entry
		e2Stat := func(e2i int32) string {
			if e2i < 0 || int(e2i) >= len(sp.e2) {
				return ""
			}
			tris := 0
			texs := map[int]bool{}
			pr := sp.e2[e2i]
			for r := pr[0]; r < pr[0]+pr[1] && int(r) < len(sp.ranges); r++ {
				rec := sp.ranges[r]
				for pass := 0; pass < 2; pass++ {
					f, c := rec[1+pass], rec[3+pass]
					for d := f; d < f+c && int(d) < len(pt.descStart); d++ {
						for k := pt.descStart[d]; k < pt.descStart[d]+pt.descCount[d]; k++ {
							b := pt.batches[k]
							n, _ := indexCount(b.prim, b.prims)
							tris += int(n) - 2
							if m := pt.mats[b.matIdx]; m.texIdx >= 0 {
								texs[m.texIdx] = true
							}
						}
					}
				}
			}
			var tl []int
			for t := range texs {
				tl = append(tl, t)
			}
			sort.Ints(tl)
			return fmt.Sprintf(" tris=%d tex=%v", tris, tl)
		}
		roots := map[int]bool{}
		if p.vis != nil {
			for _, id := range p.vis.roots[pi] {
				roots[id] = true
			}
		}
		fmt.Printf("part %d: %d nodes, %d e2, %d ranges\n", pi, len(sp.nodes), len(sp.e2), len(sp.ranges))
		for id, n := range sp.nodes {
			pos := ""
			if n.matrixIdx >= 0 {
				if m, err := sp.matrix(p, int(n.matrixIdx)); err == nil {
					pos = fmt.Sprintf(" T=(%.1f,%.1f,%.1f)", m[12], m[13], m[14])
				}
			}
			r := " "
			if roots[id] {
				r = "R"
			}
			fmt.Printf("  %s n%-4d f=%#-6x f14=%-3d f18=%-4d m=%-4d c=%-4d s=%-4d e2=%v sph=(%.1f,%.1f,%.1f r%.1f)%s%s\n",
				r, id, n.flags, n.f14, n.f18, n.matrixIdx, n.child, n.sibling, n.e2,
				n.sphere[0], n.sphere[1], n.sphere[2], n.sphere[3], pos, e2Stat(n.e2[0]))
		}
	}
	return nil
}

// exportNodes writes each named forest node (part:id,...) as its own GLB —
// the isolation instrument for judging what a single node contributes.
func (p *pmt) exportNodes(texs []texInfo, spec, outDir string) error {
	for _, tok := range strings.Split(spec, ",") {
		var pi, id int
		if _, err := fmt.Sscanf(strings.TrimSpace(tok), "%d:%d", &pi, &id); err != nil {
			return fmt.Errorf("bad -nodes token %q", tok)
		}
		pt, err := p.parsePart(pi)
		if err != nil {
			return err
		}
		sp, err := p.parsePlacements(pi)
		if err != nil {
			return err
		}
		if id >= len(sp.nodes) {
			return fmt.Errorf("part %d has %d nodes", pi, len(sp.nodes))
		}
		n := &sp.nodes[id]
		e2i := int32(-1)
		for k := 0; k < 4; k++ {
			if n.e2[k] >= 0 {
				e2i = n.e2[k]
				break
			}
		}
		if e2i < 0 {
			fmt.Printf("part %d n%d: no e2\n", pi, id)
			continue
		}
		var descs []int
		pr := sp.e2[e2i]
		for r := pr[0]; r < pr[0]+pr[1] && int(r) < len(sp.ranges); r++ {
			rec := sp.ranges[r]
			fmt.Printf("  range %d: pair=%d pass0={%d,%d} pass1={%d,%d}\n", r, rec[0], rec[1], rec[3], rec[2], rec[4])
			for pass := 0; pass < 2; pass++ {
				f, c := rec[1+pass], rec[3+pass]
				for d := f; d < f+c && int(d) < len(pt.descStart); d++ {
					descs = append(descs, int(d))
				}
			}
		}
		pool := newStagePool(p, pt)
		vn := pool.node(fmt.Sprintf("p%d n%d", pi, id), descs, texs)
		if vn == nil {
			fmt.Printf("part %d n%d: empty\n", pi, id)
			continue
		}
		out := fmt.Sprintf("%s/node-p%d-n%d.glb", outDir, pi, id)
		if err := glb.WriteVariantScenes(out, []glb.ModelVariant{{Name: "n", Nodes: []glb.VariantNode{*vn}}}); err != nil {
			return err
		}
		fmt.Printf("part %d n%d -> %s (%d verts)\n", pi, id, out, len(vn.Positions))
	}
	return nil
}
