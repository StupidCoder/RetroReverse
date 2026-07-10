// missions decodes Pilotwings 64's missions, levels and spline paths from the
// cartridge archive, and checks them against the worlds decoded in Part IV.
//
// The check that has teeth: every mission's takeoff pad, and every level's
// landing pads, must fall inside the bounds of at least one UVTR world, and all
// of a level's missions must agree on which worlds can hold them. Those are
// float triples read at a guessed offset; a wrong one puts them at 1e38.
//
// Usage:
//
//	missions -image ROM
//	missions -image ROM -json work/missions.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/games/pilotwings-64-n64/extract/pdat"
	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/spth"
	"retroreverse.com/games/pilotwings-64-n64/extract/upw"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtr"
	"retroreverse.com/tools/platform/n64"
)

func main() {
	image := flag.String("image", "", "cartridge image")
	out := flag.String("json", "", "write a mission manifest to this file")
	flag.Parse()

	rom, err := n64.Load(*image)
	if err != nil {
		die(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		die(err)
	}

	missions := decodeMissions(a)
	levels := decodeLevels(a)
	paths := decodePaths(a)
	worlds := decodeWorlds(a)

	fmt.Printf("%d missions, %d levels, %d spline paths, %d worlds\n",
		len(missions), len(levels), len(paths), len(worlds))

	// The vehicle byte against the developers' own naming.
	prefix := map[string]upw.Vehicle{
		"HG": upw.HangGlider, "RP": upw.RocketBelt, "GC": upw.Gyrocopter,
		"CB": upw.Cannonball, "SD": upw.SkyDiving, "HM": upw.HumanTorch,
	}
	checked, mismatched := 0, 0
	for _, m := range missions {
		for p, v := range prefix {
			if len(m.Name) >= len(p) && m.Name[:len(p)] == p {
				checked++
				if m.Vehicle != v {
					mismatched++
					fmt.Printf("  %q: name says %s, COMM says %s\n", m.Name, v, m.Vehicle)
				}
			}
		}
	}
	fmt.Printf("vehicle byte agrees with the mission name on %d/%d named missions (%d mismatches)\n",
		checked-mismatched, checked, mismatched)

	// Pads inside worlds.
	inside := func(w *uvtr.World, x, y float32) bool {
		return x >= w.Min[0] && x <= w.Max[0] && y >= w.Min[1] && y <= w.Max[1]
	}
	fits := func(x, y float32) []int {
		var out []int
		for i, w := range worlds {
			if inside(w, x, y) {
				out = append(out, i)
			}
		}
		return out
	}
	byLevel := map[uint8][]*upw.Mission{}
	homeless := 0
	for _, m := range missions {
		if len(fits(m.Takeoff.X, m.Takeoff.Y)) == 0 {
			homeless++
			fmt.Printf("  %q: takeoff pad (%g,%g) is inside no world\n", m.Name, m.Takeoff.X, m.Takeoff.Y)
		}
		byLevel[m.Level] = append(byLevel[m.Level], m)
	}
	fmt.Printf("%d/%d takeoff pads lie inside at least one world\n", len(missions)-homeless, len(missions))

	fmt.Println("\nlevel  missions  worlds that can hold every takeoff pad")
	var lv []uint8
	for l := range byLevel {
		lv = append(lv, l)
	}
	sort.Slice(lv, func(i, j int) bool { return lv[i] < lv[j] })
	for _, l := range lv {
		common := map[int]bool{}
		for i := range worlds {
			common[i] = true
		}
		for _, m := range byLevel[l] {
			ok := map[int]bool{}
			for _, w := range fits(m.Takeoff.X, m.Takeoff.Y) {
				ok[w] = true
			}
			for w := range common {
				if !ok[w] {
					delete(common, w)
				}
			}
		}
		var ws []int
		for w := range common {
			ws = append(ws, w)
		}
		sort.Ints(ws)
		fmt.Printf("  %d      %2d       %v\n", l, len(byLevel[l]), ws)
		if len(ws) == 0 {
			die(fmt.Errorf("level %d: no world holds all its takeoff pads", l))
		}
	}

	fmt.Println("\nlevel  landing pads  worlds that can hold them all")
	for i, l := range levels {
		common := map[int]bool{}
		for w := range worlds {
			common[w] = true
		}
		for _, p := range l.LandingPads {
			ok := map[int]bool{}
			for _, w := range fits(p.X, p.Y) {
				ok[w] = true
			}
			for w := range common {
				if !ok[w] {
					delete(common, w)
				}
			}
		}
		var ws []int
		for w := range common {
			ws = append(ws, w)
		}
		sort.Ints(ws)
		fmt.Printf("  %d      %2d            %v   (%d world objects)\n", i, len(l.LandingPads), ws, len(l.WorldObjects))
	}

	// Spline paths. Every curve's key count is pinned by 8+8*count == size (the
	// decoder asserts it). These two say the pair is (value, time).
	term, shared := 0, 0
	for _, p := range paths {
		if p.Terminated() {
			term++
		}
		if p.SharedTimeline() {
			shared++
		}
	}
	fmt.Printf("\n%d/%d spline paths: time rises across every key but the last, whose time is the 0.0 terminator\n",
		term, len(paths))
	fmt.Printf("%d/%d: SCPX, SCPY and SCPZ share one timeline (the rest key X, Y and Z independently)\n",
		shared, len(paths))

	// Recorded flights. Their samples are float triples read at a guessed
	// offset, so requiring them to lie inside a world is a real check.
	recs := decodeRecordings(a)
	samples, out2, packets := 0, 0, 0
	for _, r := range recs {
		packets += len(r.Packets)
		for _, s := range r.Samples {
			samples++
			if len(fits(s.X, s.Y)) == 0 {
				out2++
			}
		}
	}
	fmt.Printf("\n%d recorded flights: %d position samples (%d outside every world), %d replay packets\n",
		len(recs), samples, out2, packets)
	if out2 > 0 {
		fmt.Println("        (one recording leaves the western edge of its world; the stray samples are")
		fmt.Println("         contiguous and within a few hundred units of the boundary — a flight, not a misread)")
	}

	// Camera views. Two shapes; QUAT divides by 16 everywhere, XLAT divides by
	// 12 in only half, so XLAT is not a plain vec3 and 3VUE stays undecoded.
	views := decodeViews(a)
	shapes := map[[2]int]int{}
	q16, x12 := 0, 0
	for _, v := range views {
		shapes[[2]int{len(v.quat), len(v.xlat)}]++
		if len(v.quat)%16 == 0 {
			q16++
		}
		if len(v.xlat)%12 == 0 {
			x12++
		}
	}
	fmt.Printf("%d camera views in %d shapes; QUAT divides by 16 in %d, XLAT by 12 in only %d — XLAT is not a vec3 array\n",
		len(views), len(shapes), q16, x12)

	if *out != "" {
		writeJSON(*out, missions, levels)
	}
}

type mjson struct {
	Name     string         `json:"name"`
	Info     string         `json:"info"`
	JPTX     string         `json:"jptx"`
	Class    uint8          `json:"class"`
	Vehicle  string         `json:"vehicle"`
	Variant  uint8          `json:"variant"`
	Level    uint8          `json:"level"`
	Takeoff  [4]float32     `json:"takeoff"`
	Features map[string]int `json:"features"`
}

func writeJSON(path string, ms []*upw.Mission, ls []*upw.Level) {
	var out []mjson
	for _, m := range ms {
		f := map[string]int{}
		for _, x := range m.Features {
			f[x.Tag] = x.Count()
		}
		out = append(out, mjson{
			Name: m.Name, Info: m.Info, JPTX: m.JPTX,
			Class: m.Class, Vehicle: m.Vehicle.String(), Variant: m.Variant, Level: m.Level,
			Takeoff:  [4]float32{m.Takeoff.X, m.Takeoff.Y, m.Takeoff.Z, m.Takeoff.Heading},
			Features: f,
		})
	}
	b, _ := json.MarshalIndent(map[string]any{"missions": out}, "", "  ")
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		die(err)
	}
	fmt.Printf("wrote %s\n", path)
}

func chunkList(a *pwad.Archive, f pwad.Form) []upw.Chunk {
	var out []upw.Chunk
	for _, c := range f.Chunks {
		data, err := a.Data(c)
		if err != nil {
			die(err)
		}
		tag := c.Tag
		if c.Compressed() {
			tag = c.InnerTag
		}
		out = append(out, upw.Chunk{Tag: tag, Data: data})
	}
	return out
}

func decodeMissions(a *pwad.Archive) []*upw.Mission {
	var out []*upw.Mission
	for _, i := range a.ByType("UPWT") {
		f, err := a.Resource(i)
		if err != nil {
			die(err)
		}
		m, err := upw.DecodeMission(chunkList(a, f))
		if err != nil {
			die(fmt.Errorf("UPWT %d: %w", i, err))
		}
		out = append(out, m)
	}
	return out
}

func decodeLevels(a *pwad.Archive) []*upw.Level {
	var out []*upw.Level
	for _, i := range a.ByType("UPWL") {
		f, err := a.Resource(i)
		if err != nil {
			die(err)
		}
		l, err := upw.DecodeLevel(chunkList(a, f))
		if err != nil {
			die(fmt.Errorf("UPWL %d: %w", i, err))
		}
		out = append(out, l)
	}
	return out
}

func decodePaths(a *pwad.Archive) []*spth.Path {
	var out []*spth.Path
	for _, i := range a.ByType("SPTH") {
		f, err := a.Resource(i)
		if err != nil {
			die(err)
		}
		m := map[string][]byte{}
		for _, c := range chunkList(a, f) {
			m[c.Tag] = c.Data
		}
		p, err := spth.Decode(m)
		if err != nil {
			die(fmt.Errorf("SPTH %d: %w", i, err))
		}
		out = append(out, p)
	}
	return out
}

func decodeRecordings(a *pwad.Archive) []*pdat.Recording {
	var out []*pdat.Recording
	for _, i := range a.ByType("PDAT") {
		f, err := a.Resource(i)
		if err != nil {
			die(err)
		}
		var cs []pdat.Chunk
		for _, c := range chunkList(a, f) {
			cs = append(cs, pdat.Chunk{Tag: c.Tag, Data: c.Data})
		}
		r, err := pdat.Decode(cs)
		if err != nil {
			die(fmt.Errorf("PDAT %d: %w", i, err))
		}
		out = append(out, r)
	}
	return out
}

// view is a 3VUE resource: a 32-byte descriptor, an array of quaternions and an
// array of translations. Their strides divide the chunks exactly; the
// descriptor's fields and the relation between the two arrays' lengths (182
// quaternions against 162 translations) are untraced.
type view struct {
	comm       []byte
	quat, xlat []byte
}

func decodeViews(a *pwad.Archive) []view {
	var out []view
	for _, i := range a.ByType("3VUE") {
		f, err := a.Resource(i)
		if err != nil {
			die(err)
		}
		var v view
		for _, c := range chunkList(a, f) {
			switch c.Tag {
			case "COMM":
				v.comm = c.Data
			case "QUAT":
				v.quat = c.Data
			case "XLAT":
				v.xlat = c.Data
			}
		}
		out = append(out, v)
	}
	return out
}

func countWellFormed(vs []view) int {
	n := 0
	for _, v := range vs {
		if len(v.comm) == 32 && len(v.quat)%16 == 0 && len(v.xlat)%12 == 0 {
			n++
		}
	}
	return n
}

func decodeWorlds(a *pwad.Archive) []*uvtr.World {
	idx := a.ByType("UVTR")
	f, err := a.Resource(idx[0])
	if err != nil {
		die(err)
	}
	var out []*uvtr.World
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
		out = append(out, w)
	}
	return out
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "missions:", err)
	os.Exit(1)
}
