// exportbmd decodes Super Mario 64 DS ".bmd" models (the game's own format, not
// NITRO BMD0) and exports them to standard binary glTF (.glb) via sm64ds + the
// shared nitro GLB exporter.
//
//	exportbmd model.bmd...
//	exportbmd -all [-root DIR] [-o DIR]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"supermario64ds/extract/sm64ds"
)

func main() {
	all := flag.Bool("all", false, "sweep the stage + object models")
	root := flag.String("root", "../extracted/files", "extracted filesystem root (-all)")
	outDir := flag.String("o", "../extracted/glb", "output directory")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	var paths []string
	if *all {
		filepath.Walk(filepath.Join(*root, "data"), func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(p, ".bmd") {
				paths = append(paths, p)
			}
			return nil
		})
		sort.Strings(paths)
	} else {
		paths = flag.Args()
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: exportbmd model.bmd...  |  exportbmd -all")
		os.Exit(2)
	}

	type entry struct {
		Name    string `json:"name"`
		File    string `json:"file"`
		Section string `json:"section"`
		// Objects references the per-stage placement JSON (written by
		// cmd/exportlevelobjs) that the viewer overlays on a level.
		Objects string `json:"objects,omitempty"`
	}
	var manifest []entry
	seen := map[string]bool{}
	ok, fail, tris := 0, 0, 0
	for _, p := range paths {
		m, err := sm64ds.LoadBMD(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", filepath.Base(p), err)
			fail++
			continue
		}
		// Sibling .bca clips whose bone count matches the model become glTF
		// animations on a skinned export (the goomba's walk, the bob-omb's run).
		var clips []sm64ds.NamedBCA
		if sib, _ := filepath.Glob(filepath.Join(filepath.Dir(p), "*.bca")); len(sib) > 0 {
			for _, ap := range sib {
				if a, err := sm64ds.LoadBCA(ap); err == nil && a.NumBones == m.NumBones {
					stem := strings.TrimSuffix(filepath.Base(ap), ".bca")
					clips = append(clips, sm64ds.NamedBCA{Name: stem, Anim: a})
				}
			}
		}
		var glb []byte
		var err2 error
		if len(clips) > 0 {
			glb, err2 = m.SkinnedGLB(clips)
		} else {
			glb, err2 = m.GLB()
		}
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", filepath.Base(p), err2)
			fail++
			continue
		}
		file := m.Name + ".glb"
		if err := os.WriteFile(filepath.Join(*outDir, file), glb, 0o644); err != nil {
			die(err)
		}
		n := 0
		for _, ts := range m.ByMat {
			n += len(ts)
		}
		tris += n
		ok++
		if *all && !seen[file] {
			if name, sec := classify(p, m.Name); sec != "" {
				e := entry{Name: name, File: file, Section: sec}
				if sec == "Levels" {
					e.Objects = "objects/" + m.Name + ".json"
				}
				manifest = append(manifest, e)
				seen[file] = true
			}
		}
		if !*all || n > 3000 {
			fmt.Printf("  %-28s %5d tris, %2d bones, %2d materials, %2d textures → %s\n",
				m.Name, n, m.NumBones, len(m.Mats), len(m.Texs), file)
		}
	}
	if *all {
		sort.Slice(manifest, func(i, j int) bool {
			if manifest[i].Section != manifest[j].Section {
				return sectionRank(manifest[i].Section) < sectionRank(manifest[j].Section)
			}
			return manifest[i].Name < manifest[j].Name
		})
		buf, _ := json.MarshalIndent(manifest, "", " ")
		os.WriteFile(filepath.Join(*outDir, "models.json"), buf, 0o644)
		fmt.Printf("  models.json: %d entries\n", len(manifest))
	}
	fmt.Printf("exported %d models (%d failed), %d triangles total\n", ok, fail, tris)
}

// levelNames maps the internal stage stems to their course names.
var levelNames = map[string]string{
	"bombhei_map": "Bob-omb Battlefield", "snow_slider": "Cool, Cool Mountain",
	"teresa_house": "Big Boo's Haunt", "horisoko": "Jolly Roger Bay",
	"cave": "Hazy Maze Cave", "desert_land": "Shifting Sand Land",
	"fire_land": "Lethal Lava Land", "snow_land": "Snowman's Land",
	"suisou": "Wet-Dry World", "kaizoku_irie": "Dire, Dire Docks",
	"high_slider": "Tall, Tall Mountain", "rainbow_cruise": "Rainbow Cruise",
	"clock_tower": "Tick Tock Clock", "battan_king_map": "Whomp's Fortress",
	"koopa1_map": "Bowser in the Dark World", "koopa2_map": "Bowser in the Fire Sea",
	"koopa3_map": "Bowser in the Sky", "koopa1_boss": "Bowser 1 (arena)",
	"koopa2_boss": "Bowser 2 (arena)", "koopa3_boss": "Bowser 3 (arena)",
	"castle_1f": "Castle — 1st floor", "castle_2f": "Castle — 2nd floor",
	"castle_b1": "Castle — basement", "main_castle": "Peach's Castle (exterior)",
	"main_garden": "Castle grounds", "playroom": "Rec Room (minigames)",
	"kaizoku_ship": "Pirate ship", "desert_py": "Pyramid interior",
	"fire_mt": "Volcano interior", "snow_mt": "Cool Cool Mtn slide", "snow_kama": "Snowman",
	"high_mt": "Tall Mtn interior", "habatake": "Wing-cap tower", "metal_switch": "Metal-cap cavern",
	"rainbow_mario": "Over the Rainbows", "ex_mario": "Mario ? cap course",
	"ex_luigi": "Luigi ? cap course", "ex_wario": "Wario ? cap course", "horisoko2": "Jolly Roger Bay (bottom)",
}

// classify assigns a model to a viewer section with a friendly name, by its path.
func classify(path, stem string) (name, section string) {
	switch {
	case strings.Contains(path, "/stage/") && strings.HasSuffix(stem, "_all"):
		base := strings.TrimSuffix(stem, "_all")
		if n := levelNames[base]; n != "" {
			return n, "Levels"
		}
		return base, "Levels"
	case strings.Contains(path, "/player/"):
		return title(stem), "Characters"
	case strings.Contains(path, "/enemy/"):
		return title(stem), "Enemies"
	case strings.Contains(path, "/normal_obj/"), strings.Contains(path, "/special_obj/"):
		return title(stem), "Objects"
	}
	return "", "" // skip UI, effects, texture banks, etc.
}

func sectionRank(s string) int {
	switch s {
	case "Levels":
		return 0
	case "Characters":
		return 1
	case "Enemies":
		return 2
	case "Objects":
		return 3
	}
	return 9
}

func title(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	if len(s) > 0 {
		s = strings.ToUpper(s[:1]) + s[1:]
	}
	return s
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "exportbmd:", err)
	os.Exit(1)
}
