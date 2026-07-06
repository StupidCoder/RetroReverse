// Command modeldump decodes the 64 3D object models embedded in UW.EXE (the
// display-list VM bytecode programs — see extract/model) and prints, per model:
// extents, vertex/quad/face counts and the decoded listing. Use -model N for
// one model's full listing, -v for all listings.
//
// Usage: modeldump [-game ../game] [-model -1] [-v]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ultimaunderworld/extract/model"
)

func main() {
	game := flag.String("game", "../game", "path to the game/ folder")
	one := flag.Int("model", -1, "print full listing for this model only")
	verbose := flag.Bool("v", false, "print all listings")
	flag.Parse()

	exe, err := os.ReadFile(filepath.Join(*game, "UW.EXE"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	models, err := model.Decode(exe)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, m := range models {
		if *one >= 0 && m.Index != *one {
			continue
		}
		fmt.Printf("model %2d off=%04X extents=(%d,%d,%d) verts=%d quads=%d polys=%d",
			m.Index, m.Offset, m.Extents[0], m.Extents[1], m.Extents[2],
			len(m.Pool), len(m.Quads), len(m.Polys))
		if len(m.EnvSlots) > 0 {
			fmt.Printf("  env=%v (ceiling-adaptive)", m.EnvSlots)
		}
		if len(m.Warnings) > 0 {
			fmt.Printf("  WARN %v", m.Warnings)
		}
		fmt.Println()
		if *verbose || *one >= 0 {
			for _, s := range m.Ops {
				fmt.Println("   ", s)
			}
		}
	}
}
