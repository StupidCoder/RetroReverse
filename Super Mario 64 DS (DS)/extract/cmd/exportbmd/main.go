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
	rom := flag.String("rom", "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image (course names)")
	ext := flag.String("extracted", "../extracted", "extracted binaries dir (course names)")
	outDir := flag.String("o", "../extracted/glb", "output directory")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	if *all {
		if err := buildLevelNames(*rom, *ext, *root); err != nil {
			die(err)
		}
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
		// The playable Mario: the 16-bone minigame model (MG/, outside the data/
		// sweep) with the standard clips from data/player (same 16-bone skeleton).
		if m, err := sm64ds.LoadBMD(filepath.Join(*root, "MG/mario_model_mg.bmd")); err == nil {
			var clips []sm64ds.NamedBCA
			for _, cn := range []string{"su_wait", "su_walk", "su_run"} {
				if a, err := sm64ds.LoadBCA(filepath.Join(*root, "data/player", cn+".bca")); err == nil && a.NumBones == m.NumBones {
					clips = append(clips, sm64ds.NamedBCA{Name: cn, Anim: a})
				}
			}
			if glb, err := m.SkinnedGLB(clips); err == nil {
				os.WriteFile(filepath.Join(*outDir, "mario_model_mg.glb"), glb, 0o644)
				manifest = append(manifest, entry{Name: "Mario (in-game)", File: "mario_model_mg.glb", Section: "Characters"})
				ok++
			}
		}
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

// levelNames maps stage stems to display names. Course levels get the game's
// OWN course names — the level->course table at ARM9 $02075298 (the save
// system's accessor $02013558) joined with course-name messages 406+course
// from data/message/msg_data_eng.bin (see sm64ds/msg.go for the traced BMG
// format and font-derived text encoding) — title-cased, with the internal stem
// appended when several maps make up one course (the mountain and its slide).
// The castle hub, playroom and test maps aren't courses (their table entry is
// the shared "CASTLE SECRET STARS" row or -1), so they keep literal
// stem-derived labels.
var levelNames = map[string]string{}

// hubNames are the non-course levels: literal descriptions of the internal
// stem (castle_1f = the castle's first floor...), not in-game course names —
// the game groups these under one "CASTLE SECRET STARS" save row. "Castle
// Grounds" and the minigame room's label are on the cartridge as pre-rendered
// banner graphics in the per-language ARCHIVE/ce?.narc menus.
var hubNames = map[string]string{
	"main_castle": "Peach's Castle (exterior)", "main_garden": "Castle Grounds",
	"castle_1f": "Castle — 1st floor", "castle_2f": "Castle — 2nd floor",
	"castle_b1": "Castle — basement", "playroom": "Playroom",
	"test_map": "Test map", "test_map_b": "Test map B",
}

// buildLevelNames fills levelNames from the cartridge (ROM + extracted
// binaries + the English message file under root).
func buildLevelNames(romPath, extDir, root string) error {
	ls, err := sm64ds.OpenLevels(romPath, extDir)
	if err != nil {
		return err
	}
	msgs, err := sm64ds.LoadBMG(filepath.Join(root, "data/message/msg_data_eng.bin"))
	if err != nil {
		return err
	}
	// stems per course, in level order (the first level of a course is its
	// main map and gets the bare course name)
	type lv struct{ id, course int }
	byStem := map[string]lv{}
	courseCount := map[int]int{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		l, err := ls.Level(id)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(l.BMDPath), ".bmd"), "_all")
		c := ls.Course(id)
		if _, dup := byStem[stem]; !dup {
			byStem[stem] = lv{id, c}
			courseCount[c]++
		}
	}
	first := map[int]int{} // course -> lowest level id
	for _, v := range byStem {
		if f, ok := first[v.course]; !ok || v.id < f {
			first[v.course] = v.id
		}
	}
	for stem, v := range byStem {
		if n, ok := hubNames[stem]; ok {
			levelNames[stem] = n
			continue
		}
		if v.course < 0 || v.course+sm64ds.CourseNameMsg >= len(msgs) {
			levelNames[stem] = stem
			continue
		}
		name := courseTitle(msgs[v.course+sm64ds.CourseNameMsg])
		if courseCount[v.course] > 1 && first[v.course] != v.id {
			name += " (" + stem + ")" // secondary maps of the course
		}
		levelNames[stem] = name
	}
	return nil
}

// courseTitle turns a course-name message (" 4 COOL, COOL MOUNTAIN") into a
// display name: the course number is dropped and the all-caps text (the game
// displays course names in capitals) is title-cased — articles/prepositions
// lower, hyphen compounds capitalized on both sides (Wet-Dry, Tiny-Huge),
// except "Bob-omb", whose casing the game's own mixed-case dialog attests
// (message 137: "You're in the Bob-omb Battlefield!!").
func courseTitle(msg string) string {
	s := strings.TrimSpace(msg)
	s = strings.TrimLeft(s, "0123456789")
	s = strings.TrimSpace(s)
	small := map[string]bool{"IN": true, "THE": true, "OF": true, "ON": true, "TO": true, "UNDER": true}
	words := strings.Fields(s)
	for i, w := range words {
		if i > 0 && small[w] {
			words[i] = strings.ToLower(w)
			continue
		}
		if w == "BOB-OMB" {
			words[i] = "Bob-omb"
			continue
		}
		r := []rune(strings.ToLower(w))
		up := true
		for j, c := range r {
			if up {
				r[j] = []rune(strings.ToUpper(string(c)))[0]
			}
			up = c == '-'
		}
		words[i] = string(r)
	}
	return strings.Join(words, " ")
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
	case strings.Contains(path, "/vrbox/"):
		return "Skybox " + strings.TrimPrefix(stem, "vr"), "Skyboxes"
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
	case "Skyboxes":
		return 4
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
