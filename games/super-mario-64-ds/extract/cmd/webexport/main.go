// webexport runs the whole Super Mario 64 DS extraction from the raw cartridge
// and writes the Studio's common format-2 asset tree (site/FORMAT2.md) under one
// output root. It replaces the manual 6-tool pipeline (ndsextract, actororacle,
// exportbmd, exportkcl, exportlevelobjs, musicrender): it stages the filesystem
// + decompressed binaries in a temp dir, renders the SDAT music, runs the
// actor-binding oracle, exports BMD models and KCL collision to GLB, and builds
// the per-level object database — reusing the same sm64ds / nitro / nds / sdat
// package APIs the individual commands call.
//
//	webexport [-in rom.nds] [-o DIR] [-only music,models,levels,all]
//
// -only gates which stages run: `-only music` renders MP3s only and never boots
// the oracle; models/levels require the oracle + GLB export path. Progress is
// reported on stderr, one line per stage plus a running count within long stages.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/nitro"
	"retroreverse.com/tools/platform/nds/sdat"
)

const (
	rate     = 32768        // DS mixer output rate (musicrender)
	toGLB    = 1.0 / 4096 / 1000 // fx20.12 world -> stage-GLB units (exportkcl)
	toStage  = 1.0 / 1000   // placement short -> stage-GLB units (exportlevelobjs)
	objScale = 1.0 / 125    // object display scale in stage-GLB units (exportlevelobjs)
)

func main() {
	in := flag.String("in", "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image (.nds)")
	out := flag.String("o", "../../site/public/super-mario-64-ds", "output root")
	only := flag.String("only", "all", "comma-separated subset of music,models,levels,all")
	flag.Parse()

	sel := parseOnly(*only)
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}

	// manifest accumulators
	var music []manifestMusic
	var models []manifestModel
	var levels []manifestLevel

	if sel["music"] {
		music = runMusic(*in, *out)
	}

	if sel["models"] || sel["levels"] {
		// Stage the filesystem + decompressed binaries the oracle/decoders read.
		tmp, err := os.MkdirTemp("", "sm64ds-webexport-")
		if err != nil {
			die(err)
		}
		defer os.RemoveAll(tmp)
		fmt.Fprintf(os.Stderr, "[extract] staging filesystem + decompressed binaries → %s\n", tmp)
		if err := extractFS(*in, tmp); err != nil {
			die(err)
		}

		ls, err := sm64ds.OpenLevels(*in, tmp)
		if err != nil {
			die(err)
		}
		if err := buildLevelNames(ls, tmp); err != nil {
			die(err)
		}

		bindings := runOracle(*in, tmp)

		// GLB export path (needed by both models and levels): stage/object models,
		// archive-member models, and collision meshes into <o>/models + <o>/collision.
		models = exportModels(ls, tmp, *out)
		exportArchiveGLBs(ls, bindings, filepath.Join(*out, "models"))
		exportCollision(ls, tmp, *out, bindings)

		if sel["levels"] {
			levels = exportLevels(ls, tmp, *out, bindings)
		}
	}

	writeManifest(*out, music, models, levels)
}

func parseOnly(s string) map[string]bool {
	sel := map[string]bool{}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if p == "all" {
			sel["music"], sel["models"], sel["levels"] = true, true, true
			continue
		}
		sel[p] = true
	}
	return sel
}

// ---------------------------------------------------------------------------
// manifest
// ---------------------------------------------------------------------------

type manifestMusic struct {
	Name string `json:"name"`
	File string `json:"file"`
}
type manifestModel struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"`
}
type manifestLevel struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Objects string `json:"objects"`
}

func writeManifest(out string, music []manifestMusic, models []manifestModel, levels []manifestLevel) {
	m := map[string]any{
		"format":   2,
		"game":     "super-mario-64-ds",
		"platform": "Nintendo DS",
		"native":   map[string]int{"w": 256, "h": 192},
		"tickHz":   60,
	}
	if len(levels) > 0 {
		m["levels"] = levels
	}
	if len(models) > 0 {
		m["models"] = models
	}
	if len(music) > 0 {
		m["music"] = music
	}
	buf, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(out, "manifest.json"), buf, 0o644); err != nil {
		die(err)
	}
	fmt.Fprintf(os.Stderr, "[manifest] %d levels, %d models, %d music → manifest.json\n",
		len(levels), len(models), len(music))
}

// ---------------------------------------------------------------------------
// music (as cmd/musicrender)
// ---------------------------------------------------------------------------

func runMusic(romPath, out string) []manifestMusic {
	img, err := os.ReadFile(romPath)
	if err != nil {
		die(err)
	}
	rom, err := nds.Open(img)
	if err != nil {
		die(err)
	}
	data := rom.FileByPath("data/sound_data.sdat")
	if data == nil {
		die(fmt.Errorf("data/sound_data.sdat not found in ROM"))
	}
	s, err := sdat.Parse(data)
	if err != nil {
		die(err)
	}
	musicDir := filepath.Join(out, "music")
	if err := os.MkdirAll(musicDir, 0o755); err != nil {
		die(err)
	}

	// count renderable sequences (those with a file) for the running total
	total := 0
	for i := range s.Seqs {
		if s.Seqs[i].FileID >= 0 {
			total++
		}
	}
	fmt.Fprintf(os.Stderr, "[music] SDAT: %d sequences (%d renderable), %d banks, %d wave archives\n",
		len(s.Seqs), total, len(s.Banks), len(s.Wavearcs))

	var tracks []manifestMusic
	n := 0
	for i := range s.Seqs {
		if s.Seqs[i].FileID < 0 {
			continue
		}
		n++
		stem := stemFor(i, s.Seqs[i].Name)
		L, R, err := s.Render(i, rate, 2, 180)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[music]  %d/%d  %s: %v\n", n, total, stem, err)
			continue
		}
		if len(L) < rate { // sub-second jingles cut short
			fmt.Fprintf(os.Stderr, "[music]  %d/%d  %s (skipped, too short)\n", n, total, stem)
			continue
		}
		fadeOut(L, R)
		wav := filepath.Join(musicDir, stem+".wav")
		if err := writeWAV(wav, L, R); err != nil {
			die(err)
		}
		mp3 := filepath.Join(musicDir, stem+".mp3")
		c := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", wav,
			"-c:a", "libmp3lame", "-b:a", "96k", mp3)
		if e := c.Run(); e != nil {
			fmt.Fprintf(os.Stderr, "[music]  %d/%d  %s: ffmpeg failed: %v (WAV kept)\n", n, total, stem, e)
			continue
		}
		os.Remove(wav)
		tracks = append(tracks, manifestMusic{Name: stem, File: "music/" + stem + ".mp3"})
		fmt.Fprintf(os.Stderr, "[music]  %d/%d  %s.mp3\n", n, total, stem)
	}
	sort.Slice(tracks, func(i, j int) bool { return tracks[i].Name < tracks[j].Name })
	return tracks
}

func stemFor(i int, name string) string {
	if name == "" {
		return fmt.Sprintf("seq_%02d", i)
	}
	return strings.ToLower(strings.TrimPrefix(name, "NCS_BGM_"))
}

func fadeOut(L, R []float64) {
	n := 3 * rate
	if n > len(L) {
		n = len(L)
	}
	for i := 0; i < n; i++ {
		g := float64(n-i) / float64(n)
		L[len(L)-n+i] *= g
		R[len(R)-n+i] *= g
	}
}

func writeWAV(path string, L, R []float64) error {
	n := len(L)
	buf := make([]byte, 44+n*4)
	copy(buf, "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(36+n*4))
	copy(buf[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)
	binary.LittleEndian.PutUint16(buf[20:], 1) // PCM
	binary.LittleEndian.PutUint16(buf[22:], 2) // stereo
	binary.LittleEndian.PutUint32(buf[24:], rate)
	binary.LittleEndian.PutUint32(buf[28:], rate*4)
	binary.LittleEndian.PutUint16(buf[32:], 4)
	binary.LittleEndian.PutUint16(buf[34:], 16)
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(n*4))
	clip := func(v float64) int16 {
		if v > 1 {
			v = 1
		}
		if v < -1 {
			v = -1
		}
		return int16(v * 32767)
	}
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint16(buf[44+i*4:], uint16(clip(L[i])))
		binary.LittleEndian.PutUint16(buf[46+i*4:], uint16(clip(R[i])))
	}
	return os.WriteFile(path, buf, 0o644)
}

// ---------------------------------------------------------------------------
// filesystem staging (as cmd/ndsextract -fs, plus decompressed overlays)
// ---------------------------------------------------------------------------

// extractFS writes the decompressed ARM9 binary, every ARM9 overlay's
// decompressed image (ovl9_NNN_dec.bin — what sm64ds.OpenLevels/Oracle read via
// overlayData), and the full filesystem under files/, into dir.
func extractFS(romPath, dir string) error {
	img, err := os.ReadFile(romPath)
	if err != nil {
		return err
	}
	rom, err := nds.Open(img)
	if err != nil {
		return err
	}
	arm9 := rom.ARM9()
	if nds.IsBLZ(arm9) {
		arm9 = nds.DecompressBLZ(arm9)
	}
	if err := os.WriteFile(filepath.Join(dir, "arm9_dec.bin"), arm9, 0o644); err != nil {
		return err
	}
	for _, o := range rom.ARM9Overlays() {
		raw := rom.File(int(o.FileID))
		dec := raw
		if o.Compressed && nds.IsBLZ(raw) {
			dec = nds.DecompressBLZ(raw)
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("ovl9_%03d_dec.bin", o.ID)), dec, 0o644); err != nil {
			return err
		}
	}
	for _, f := range rom.Files {
		p := filepath.Join(dir, "files", filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, rom.File(f.ID), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// oracle (as cmd/actororacle)
// ---------------------------------------------------------------------------

// Binding mirrors cmd/actororacle's / cmd/exportlevelobjs's table entry.
type Binding struct {
	Params    [3]int            `json:"params"`
	Config    int               `json:"config"`
	Models    []string          `json:"models,omitempty"`
	Clips     []string          `json:"clips,omitempty"`
	KCL       []string          `json:"kcl,omitempty"`
	Colliders []sm64ds.Collider `json:"colliders,omitempty"`
	Notes     []string          `json:"notes,omitempty"`
}

func runOracle(romPath, tmp string) map[int][]Binding {
	fmt.Fprintln(os.Stderr, "[oracle] booting ARM9 + engine overlays…")
	ls, err := sm64ds.OpenLevels(romPath, tmp)
	if err != nil {
		die(err)
	}
	o, err := sm64ds.NewOracle(ls)
	if err != nil {
		die(err)
	}
	if err := o.InitEngine(); err != nil {
		die(err)
	}
	fmt.Fprintf(os.Stderr, "[oracle] engine initialized (%d init-phase file requests); sweeping actors…\n", len(o.InitRequests()))
	table := sweep(ls, o)
	fmt.Fprintf(os.Stderr, "[oracle] %d actors bound\n", len(table))
	return table
}

// sweep runs every distinct (actor, params) the levels place — identical to
// cmd/actororacle's sweep.
func sweep(ls *sm64ds.LevelSet, o *sm64ds.Oracle) map[int][]Binding {
	type combo struct{ actor, p1, p2, p3 int }
	perLevel := map[int]map[combo]bool{}
	all := map[combo][]int{}
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			continue
		}
		if perLevel[lv.Overlay] == nil {
			perLevel[lv.Overlay] = map[combo]bool{}
		}
		for _, ob := range lv.Objects {
			c := combo{ob.Actor, ob.Params[0], ob.Params[1], ob.Params[2]}
			if !perLevel[lv.Overlay][c] {
				perLevel[lv.Overlay][c] = true
				all[c] = append(all[c], lv.Overlay)
			}
		}
	}

	table := map[int][]Binding{}
	addRun := func(run *sm64ds.ActorRun) {
		if len(run.Files) == 0 && run.Obj == 0 {
			return
		}
		b := Binding{Params: run.Params, Config: run.Config, Models: o.Models(run), Clips: o.Clips(run), KCL: o.KCLs(run), Colliders: run.Colliders, Notes: run.Notes}
		for _, e := range table[run.Actor] {
			if fmt.Sprint(e.Models, e.Params) == fmt.Sprint(b.Models, b.Params) {
				return
			}
		}
		table[run.Actor] = append(table[run.Actor], b)
	}

	unresolved := map[combo]bool{}
	var ovls []int
	for ov := range perLevel {
		ovls = append(ovls, ov)
	}
	sort.Ints(ovls)
	for _, ov := range ovls {
		if err := o.LoadConfig(ov); err != nil {
			fmt.Fprintf(os.Stderr, "[oracle] config ovl%d: %v\n", ov, err)
			continue
		}
		for c := range perLevel[ov] {
			if _, ok := o.Profile(c.actor, ov); !ok {
				unresolved[c] = true
				continue
			}
			addRun(o.RunActor(c.actor, ov, [3]int{c.p1, c.p2, c.p3}))
		}
	}

	bankRuns := map[string]bool{}
	for bank := 60; bank <= 102; bank++ {
		var todo []combo
		for c := range unresolved {
			todo = append(todo, c)
		}
		if len(todo) == 0 {
			break
		}
		if err := o.LoadConfig(bank); err != nil {
			continue
		}
		for _, c := range todo {
			if _, ok := o.Profile(c.actor, bank); !ok {
				continue
			}
			key := fmt.Sprint(bank, c)
			if bankRuns[key] {
				continue
			}
			bankRuns[key] = true
			addRun(o.RunActor(c.actor, bank, [3]int{c.p1, c.p2, c.p3}))
		}
	}
	return table
}

// ---------------------------------------------------------------------------
// models: BMD -> GLB (as cmd/exportbmd)
// ---------------------------------------------------------------------------

// levelStems records the display name of every stage model (classify -> "Levels"),
// so exportLevels can label its manifest entries. Filled by exportModels.
var levelStems = map[string]string{}

func exportModels(ls *sm64ds.LevelSet, tmp, out string) []manifestModel {
	modelsDir := filepath.Join(out, "models")
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		die(err)
	}
	root := filepath.Join(tmp, "files")

	var paths []string
	filepath.Walk(filepath.Join(root, "data"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".bmd") {
			paths = append(paths, p)
		}
		return nil
	})
	sort.Strings(paths)

	fmt.Fprintf(os.Stderr, "[models] exporting %d BMD models…\n", len(paths))
	var models []manifestModel
	seen := map[string]bool{}
	n := 0
	for _, p := range paths {
		n++
		m, err := sm64ds.LoadBMD(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[models]  %d/%d  %s: %v\n", n, len(paths), filepath.Base(p), err)
			continue
		}
		// sibling .bca clips whose bone count matches become glTF animations
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
		if len(clips) > 0 {
			glb, err = m.SkinnedGLB(clips)
		} else {
			glb, err = m.GLB()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "[models]  %d/%d  %s: %v\n", n, len(paths), filepath.Base(p), err)
			continue
		}
		file := m.Name + ".glb"
		if err := os.WriteFile(filepath.Join(modelsDir, file), glb, 0o644); err != nil {
			die(err)
		}
		if seen[file] {
			continue
		}
		seen[file] = true
		name, sec := classify(p, m.Name)
		if sec == "Levels" {
			levelStems[m.Name] = name // becomes a levels[] entry, not a model
		} else if sec != "" {
			models = append(models, manifestModel{Name: name, Section: sec, File: "models/" + file})
		}
		if n%50 == 0 || sec != "" {
			fmt.Fprintf(os.Stderr, "[models]  %d/%d  %s\n", n, len(paths), file)
		}
	}

	// The playable Mario: the 16-bone minigame model (MG/, outside data/) with
	// the standard clips from data/player.
	if m, err := sm64ds.LoadBMD(filepath.Join(root, "MG/mario_model_mg.bmd")); err == nil {
		var clips []sm64ds.NamedBCA
		for _, cn := range []string{"su_wait", "su_walk", "su_run"} {
			if a, err := sm64ds.LoadBCA(filepath.Join(root, "data/player", cn+".bca")); err == nil && a.NumBones == m.NumBones {
				clips = append(clips, sm64ds.NamedBCA{Name: cn, Anim: a})
			}
		}
		if glb, err := m.SkinnedGLB(clips); err == nil {
			os.WriteFile(filepath.Join(modelsDir, "mario_model_mg.glb"), glb, 0o644)
			models = append(models, manifestModel{Name: "Mario (in-game)", Section: "Characters", File: "models/mario_model_mg.glb"})
		}
	}

	sort.Slice(models, func(i, j int) bool {
		if models[i].Section != models[j].Section {
			return sectionRank(models[i].Section) < sectionRank(models[j].Section)
		}
		return models[i].Name < models[j].Name
	})
	fmt.Fprintf(os.Stderr, "[models] %d models classified for the manifest\n", len(models))
	return models
}

// exportArchiveGLBs decodes archive-member models the bindings name (arcN_M)
// into GLBs (as cmd/exportlevelobjs.extractArchiveGLBs).
func exportArchiveGLBs(ls *sm64ds.LevelSet, bindings map[int][]Binding, modelsDir string) {
	done := map[string]bool{}
	var stems []string
	for _, bs := range bindings {
		for _, b := range bs {
			for _, stem := range b.Models {
				if strings.LastIndexByte(stem, '_') < 0 || done[stem] {
					continue
				}
				done[stem] = true
				stems = append(stems, stem)
			}
		}
	}
	sort.Strings(stems)
	n := 0
	for _, stem := range stems {
		ref, ok := archiveRefByStem(ls, stem)
		if !ok {
			continue
		}
		data, err := ls.ArchiveMember(ref)
		if err != nil || !sm64ds.PlausibleBMD(data) {
			continue
		}
		m, err := sm64ds.Decode(data, stem)
		if err != nil {
			continue
		}
		glb, err := m.GLB()
		if err != nil {
			continue
		}
		os.WriteFile(filepath.Join(modelsDir, stem+".glb"), glb, 0o644)
		n++
	}
	fmt.Fprintf(os.Stderr, "[models] %d archive-member models decoded\n", n)
}

func archiveRefByStem(ls *sm64ds.LevelSet, stem string) (sm64ds.ArchiveRef, bool) {
	i := strings.LastIndexByte(stem, '_')
	if i < 0 {
		return sm64ds.ArchiveRef{}, false
	}
	var member int
	if _, err := fmt.Sscanf(stem[i+1:], "%d", &member); err != nil {
		return sm64ds.ArchiveRef{}, false
	}
	name := stem[:i]
	for _, arc := range []string{"arc0", "ar1", "c2d", "cee", "cef", "ceg", "cei", "ces", "en1", "vs1", "vs2", "vs3", "vs4"} {
		if name == arc {
			return sm64ds.ArchiveRef{Archive: name, Member: member}, true
		}
	}
	return sm64ds.ArchiveRef{}, false
}

// classify assigns a model to a viewer section with a friendly name, by its path
// (as cmd/exportbmd.classify).
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
	return "", ""
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

// ---------------------------------------------------------------------------
// level display names (as cmd/exportbmd.buildLevelNames)
// ---------------------------------------------------------------------------

var levelNames = map[string]string{}

var hubNames = map[string]string{
	"main_castle": "Peach's Castle (exterior)", "main_garden": "Castle Grounds",
	"castle_1f": "Castle — 1st floor", "castle_2f": "Castle — 2nd floor",
	"castle_b1": "Castle — basement", "playroom": "Playroom",
	"test_map": "Test map", "test_map_b": "Test map B",
}

func buildLevelNames(ls *sm64ds.LevelSet, tmp string) error {
	msgs, err := sm64ds.LoadBMG(filepath.Join(tmp, "files/data/message/msg_data_eng.bin"))
	if err != nil {
		return err
	}
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
	first := map[int]int{}
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
			name += " (" + stem + ")"
		}
		levelNames[stem] = name
	}
	return nil
}

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

// ---------------------------------------------------------------------------
// collision: KCL -> GLB (as cmd/exportkcl)
// ---------------------------------------------------------------------------

func exportCollision(ls *sm64ds.LevelSet, tmp, out string, bindings map[int][]Binding) {
	colDir := filepath.Join(out, "collision")
	if err := os.MkdirAll(colDir, 0o755); err != nil {
		die(err)
	}

	kclPath := map[string]string{}
	for i := 0; i < 2058; i++ {
		if n := ls.InternalName(i); strings.HasSuffix(n, ".kcl") {
			stem := strings.TrimSuffix(filepath.Base(n), ".kcl")
			if _, dup := kclPath[stem]; !dup {
				kclPath[stem] = n
			}
		}
	}

	export := func(stem string) error {
		p, ok := kclPath[stem]
		if !ok {
			return fmt.Errorf("no .kcl named %s in the file table", stem)
		}
		data, err := os.ReadFile(filepath.Join(tmp, "files", filepath.FromSlash(strings.TrimPrefix(p, "/"))))
		if err != nil {
			return err
		}
		if len(data) > 4 && string(data[:4]) == "LZ77" {
			data = nds.Decompress(data[4:])
		}
		k, err := sm64ds.ParseKCL(data)
		if err != nil {
			return err
		}
		tris, _ := trisOf(k)
		if len(tris) == 0 {
			return fmt.Errorf("%s: no triangles", stem)
		}
		glb, err := nitro.ExportTrisGLB(stem+"_col", map[int][]nitro.Tri{0: tris},
			[]nitro.Material{{Name: "collision", Alpha: 31}}, nil)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(colDir, stem+".glb"), glb, 0o644)
	}

	done := map[string]bool{}
	levels := 0
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.KCLPath), ".kcl")
		if stem == "" || done[stem] {
			continue
		}
		done[stem] = true
		if err := export(stem); err != nil {
			fmt.Fprintf(os.Stderr, "[collision]  level %d %s: %v\n", i, stem, err)
			continue
		}
		levels++
	}

	// every object collider an actor's own code loaded (oracle bindings)
	var objStems []string
	for _, bs := range bindings {
		for _, b := range bs {
			objStems = append(objStems, b.KCL...)
		}
	}
	sort.Strings(objStems)
	objects := 0
	for _, stem := range objStems {
		if done[stem] {
			continue
		}
		done[stem] = true
		if err := export(stem); err != nil {
			fmt.Fprintf(os.Stderr, "[collision]  object %s: %v\n", stem, err)
			continue
		}
		objects++
	}
	fmt.Fprintf(os.Stderr, "[collision] %d level + %d object collision meshes\n", levels, objects)
}

func trisOf(k *sm64ds.KCL) (tris []nitro.Tri, skipped int) {
	const lx, ly, lz = 0.30, 0.90, 0.30
	for i := 1; i < k.NumPrisms(); i++ {
		c, ok := k.Corners(i)
		if !ok {
			skipped++
			continue
		}
		n := k.NormalAt(k.PrismAt(i).FaceNormal)
		nl := math.Sqrt(float64(n[0])*float64(n[0]) + float64(n[1])*float64(n[1]) + float64(n[2])*float64(n[2]))
		shade := 0.55
		if nl > 0 {
			d := (float64(n[0])*lx + float64(n[1])*ly + float64(n[2])*lz) / nl
			if d > 0 {
				shade += 0.45 * d
			}
		}
		col := color.NRGBA{R: uint8(210 * shade), G: uint8(38 * shade), B: uint8(38 * shade), A: 255}
		var t nitro.Tri
		bad := false
		for j := 0; j < 3; j++ {
			x, y, z := c[j][0]*toGLB, c[j][1]*toGLB, c[j][2]*toGLB
			if math.Abs(x) > 1e4 || math.Abs(y) > 1e4 || math.Abs(z) > 1e4 {
				bad = true
				break
			}
			t.V[j] = nitro.Vertex{X: x, Y: y, Z: z, C: col}
		}
		if bad {
			skipped++
			continue
		}
		tris = append(tris, t)
	}
	return tris, skipped
}

// ---------------------------------------------------------------------------
// levels: per-stage object DB + format-2 envelope (as cmd/exportlevelobjs)
// ---------------------------------------------------------------------------

// f2obj is one object in the format-2 <level>.objects.json database.
type f2obj struct {
	ID        int            `json:"id"`
	Actor     int            `json:"actor"`
	Pos       []float64      `json:"pos"`
	Rot       []float64      `json:"rot,omitempty"`
	Model     string         `json:"model,omitempty"`
	Collision string         `json:"collision,omitempty"`
	Props     map[string]any `json:"props,omitempty"`
}

func exportLevels(ls *sm64ds.LevelSet, tmp, out string, bindings map[int][]Binding) []manifestLevel {
	levelsDir := filepath.Join(out, "levels")
	if err := os.MkdirAll(levelsDir, 0o755); err != nil {
		die(err)
	}
	modelsDir := filepath.Join(out, "models")
	colDir := filepath.Join(out, "collision")

	modelFor := func(actor int, par [3]int) string {
		var loose string
		for _, b := range bindings[actor] {
			if len(b.Models) == 0 {
				continue
			}
			if b.Params == par {
				return b.Models[0]
			}
			if loose == "" && b.Params[0] == par[0] {
				loose = b.Models[0]
			}
		}
		return loose
	}
	colFor := func(actor int, par [3]int) *sm64ds.Collider {
		var loose *sm64ds.Collider
		for i := range bindings[actor] {
			b := &bindings[actor][i]
			if len(b.Colliders) == 0 || b.Colliders[0].KCL == "" {
				continue
			}
			if b.Params == par {
				return &b.Colliders[0]
			}
			if loose == nil && b.Params[0] == par[0] {
				loose = &b.Colliders[0]
			}
		}
		return loose
	}
	hasGLB := func(name string) bool {
		if name == "" {
			return false
		}
		_, err := os.Stat(filepath.Join(modelsDir, name+".glb"))
		return err == nil
	}
	hasCol := func(name string) bool {
		if name == "" {
			return false
		}
		_, err := os.Stat(filepath.Join(colDir, name+".glb"))
		return err == nil
	}
	bill := billboardChecker(ls, tmp)

	msgs, err := sm64ds.LoadBMG(filepath.Join(tmp, "files/data/message/msg_data_eng.bin"))
	if err != nil {
		die(err)
	}
	const signpostActor = 184
	signText := func(o sm64ds.LevelObject) string {
		if o.Actor != signpostActor {
			return ""
		}
		if idx := ls.MsgIndex(o.Params[0]); idx >= 0 && idx < len(msgs) {
			return msgs[idx]
		}
		return ""
	}

	// fill mutates a format-2 object with model/collision/props from a binding.
	fill := func(j *f2obj, actor int, par [3]int) {
		if j.Props == nil {
			j.Props = map[string]any{}
		}
		if m := modelFor(actor, par); m != "" && hasGLB(m) {
			j.Model = "models/" + m + ".glb"
			j.Props["scale"] = objScale
			if bill(m) {
				j.Props["billboard"] = true
			}
		}
		if c := colFor(actor, par); c != nil && hasCol(c.KCL) {
			j.Collision = "collision/" + c.KCL + ".glb"
			if c.Class != "Kc" { // Mbg classes carry their own transform
				cm := make([]float64, 12)
				identity := true
				for k := 0; k < 9; k++ {
					cm[k] = r5(float64(c.Mtx[k]) / 4096)
					want := 0.0
					if k%4 == 0 {
						want = 1
					}
					if cm[k] != want {
						identity = false
					}
				}
				for k := 0; k < 3; k++ {
					cm[9+k] = r5(float64(c.Mtx[9+k]) / 4096 * toStage)
					if cm[9+k] != 0 {
						identity = false
					}
				}
				if !identity {
					j.Props["colMtx"] = cm
				}
				if c.ScaleY != 0 && c.ScaleY != 0x1000 {
					j.Props["colScaleY"] = r5(float64(c.ScaleY) / 4096)
				}
			}
		}
	}

	var manifest []manifestLevel
	stageN, total := 0, 0
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		if !hasGLB(stem) {
			continue // no stage model exported
		}

		var objs []f2obj
		seen := map[string]bool{}
		id := 0
		mkObj := func(o sm64ds.LevelObject, actor int, par [3]int) f2obj {
			j := f2obj{
				ID:    id,
				Actor: actor,
				Pos:   []float64{r3(o.X * toStage), r3(o.Y * toStage), r3(o.Z * toStage)},
				Rot:   []float64{0, r3(o.RotY), 0},
				Props: map[string]any{},
			}
			id++
			if o.Layer != 0 {
				j.Props["layer"] = o.Layer
			}
			if t := signText(o); t != "" {
				j.Props["text"] = t
			}
			fill(&j, actor, par)
			if len(j.Props) == 0 {
				j.Props = nil
			}
			return j
		}
		for _, o := range lv.Objects {
			key := fmt.Sprintf("%d/%.3f/%.3f/%.3f", o.Actor, o.X, o.Y, o.Z)
			if seen[key] {
				continue
			}
			seen[key] = true
			j := mkObj(o, o.Actor, o.Params)
			if j.Model != "" {
				total++
			}
			objs = append(objs, j)
			// the chain chomp's spawned stake pile (traced child; see exportlevelobjs)
			if o.Actor == 219 {
				s := mkObj(o, 27, [3]int{65535, 0, 0})
				s.Rot = nil // spawned at parent pos, no independent rotation recorded
				objs = append(objs, s)
			}
		}
		if len(objs) == 0 {
			continue
		}

		// write the object database
		objFile := stem + ".objects.json"
		dbBuf, _ := json.MarshalIndent(map[string]any{
			"format":  2,
			"level":   stem,
			"objects": objs,
		}, "", " ")
		if err := os.WriteFile(filepath.Join(levelsDir, objFile), dbBuf, 0o644); err != nil {
			die(err)
		}

		// write the mesh3d level envelope
		lf := map[string]any{
			"format":      2,
			"name":        levelStems[stem],
			"kind":        "mesh3d",
			"mesh":        map[string]string{"glb": "models/" + stem + ".glb"},
			"objectsFile": objFile,
		}
		if lf["name"] == "" || lf["name"] == nil {
			lf["name"] = stem
		}
		if len(lv.Entrances) > 0 {
			e := lv.Entrances[0]
			lf["spawn"] = map[string]any{
				"pos": []float64{r3(e.X * toStage), r3(e.Y * toStage), r3(e.Z * toStage)},
				"rot": r3(e.RotY),
			}
		}
		if sky := strings.TrimSuffix(filepath.Base(lv.SkyPath), ".bmd"); lv.SkyPath != "" && hasGLB(sky) {
			lf["sky"] = "models/" + sky + ".glb"
		}
		// the stage's own collision mesh (its KCL), for the viewer's collision layer
		if kcl := strings.TrimSuffix(filepath.Base(lv.KCLPath), ".kcl"); lv.KCLPath != "" && hasCol(kcl) {
			lf["collision"] = "collision/" + kcl + ".glb"
		}
		lfBuf, _ := json.MarshalIndent(lf, "", " ")
		if err := os.WriteFile(filepath.Join(levelsDir, stem+".json"), lfBuf, 0o644); err != nil {
			die(err)
		}

		name, _ := lf["name"].(string)
		manifest = append(manifest, manifestLevel{
			Name:    name,
			Section: "Levels",
			File:    "levels/" + stem + ".json",
			Kind:    "mesh3d",
			Objects: "levels/" + objFile,
		})
		stageN++
		fmt.Fprintf(os.Stderr, "[levels]  %d  %s\n", stageN, stem)
	}
	sort.Slice(manifest, func(i, j int) bool { return manifest[i].Name < manifest[j].Name })
	fmt.Fprintf(os.Stderr, "[levels] %d stages, %d placements bound to models\n", stageN, total)
	return manifest
}

// billboardChecker reports whether a model's bones carry the camera-facing flag
// (as cmd/exportlevelobjs.billboardChecker).
func billboardChecker(ls *sm64ds.LevelSet, tmp string) func(stem string) bool {
	path := map[string]string{}
	for i := 0; i < 2058; i++ {
		n := ls.InternalName(i)
		if strings.HasSuffix(n, ".bmd") {
			s := strings.TrimSuffix(filepath.Base(n), ".bmd")
			if _, dup := path[s]; !dup {
				path[s] = n
			}
		}
	}
	cache := map[string]bool{}
	return func(stem string) bool {
		if v, ok := cache[stem]; ok {
			return v
		}
		var m *sm64ds.Model
		if p, ok := path[stem]; ok {
			m, _ = sm64ds.LoadBMD(filepath.Join(tmp, "files", filepath.FromSlash(strings.TrimPrefix(p, "/"))))
		} else if ref, ok := archiveRefByStem(ls, stem); ok {
			if data, err := ls.ArchiveMember(ref); err == nil && sm64ds.PlausibleBMD(data) {
				m, _ = sm64ds.Decode(data, stem)
			}
		}
		v := m != nil && len(m.Skel) == 1 && m.Skel[0].Billboard
		cache[stem] = v
		return v
	}
}

func r3(v float64) float64 { return float64(int(v*1000+0.5*sign(v))) / 1000 }
func r5(v float64) float64 { return float64(int(v*100000+0.5*sign(v))) / 100000 }
func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "webexport:", err)
	os.Exit(1)
}
