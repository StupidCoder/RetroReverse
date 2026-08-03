// bbflags dumps the bone flag word (+$3C) of every billboard bone in the
// cartridge's models, so the bits can be partitioned against what the objects
// visibly do: SM64DS's trees stay upright and turn about Y, while the Bob-omb
// bodies, Chain Chomps and rolling balls are spheres and face the camera
// outright. If a bit separates those two sets, that bit is the mode.
//
//	go run ./extract/cmd/bbflags -in "<rom>.nds"
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
	"retroreverse.com/tools/platform/nds"
)

func main() {
	in := flag.String("in", "", "ROM path")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "usage: bbflags -in <rom.nds>")
		os.Exit(2)
	}
	tmp, err := os.MkdirTemp("", "bbflags-")
	must(err)
	defer os.RemoveAll(tmp)
	must(extractFiles(*in, tmp))

	type row struct {
		model, bone string
		flags       uint32
		bones       int
	}
	var rows []row
	err = filepath.Walk(filepath.Join(tmp, "files"), func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() || !strings.HasSuffix(p, ".bmd") {
			return nil
		}
		m, err := sm64ds.LoadBMD(p)
		if err != nil {
			return nil
		}
		for _, j := range m.Skel {
			if j.Flags != 0 {
				rows = append(rows, row{m.Name, j.Name, j.Flags, len(m.Skel)})
			}
		}
		return nil
	})
	must(err)

	sort.Slice(rows, func(i, k int) bool {
		if rows[i].flags != rows[k].flags {
			return rows[i].flags < rows[k].flags
		}
		return rows[i].model < rows[k].model
	})

	byFlag := map[uint32][]string{}
	for _, r := range rows {
		byFlag[r.flags] = append(byFlag[r.flags], fmt.Sprintf("%s/%s(%db)", r.model, r.bone, r.bones))
	}
	var keys []uint32
	for k := range byFlag {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, k int) bool { return keys[i] < keys[k] })
	fmt.Printf("%d bones with a non-zero flag word, %d distinct values\n\n", len(rows), len(keys))
	for _, k := range keys {
		v := byFlag[k]
		fmt.Printf("flags %#010x (bits %s)  %d bones\n", k, bits(k), len(v))
		for i, s := range v {
			if i == 12 {
				fmt.Printf("      … and %d more\n", len(v)-12)
				break
			}
			fmt.Printf("      %s\n", s)
		}
		fmt.Println()
	}

	// The specific objects the eye can adjudicate.
	fmt.Println("=== the models under discussion ===")
	want := []string{"bomb_tree", "toge_tree", "yashi_tree", "yuki_tree", // upright, turn about Y
		"bombhei", "red_bombhei", "bomb_king", // sphere bodies
		"kb1_ball", "snow_ball", "yurei_mucho_ball", "sanbo_body", "wanwan"}
	for _, w := range want {
		found := false
		for _, r := range rows {
			if r.model == w {
				fmt.Printf("  %-18s %-16s flags %#010x  bones=%d\n", r.model, r.bone, r.flags, r.bones)
				found = true
			}
		}
		if !found {
			fmt.Printf("  %-18s (no flagged bone)\n", w)
		}
	}
}

func bits(v uint32) string {
	var b []string
	for i := 0; i < 32; i++ {
		if v&(1<<i) != 0 {
			b = append(b, fmt.Sprint(i))
		}
	}
	if b == nil {
		return "-"
	}
	return strings.Join(b, ",")
}

func extractFiles(romPath, dir string) error {
	img, err := os.ReadFile(romPath)
	if err != nil {
		return err
	}
	rom, err := nds.Open(img)
	if err != nil {
		return err
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

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
