// modeldump summarises an NSBMD model file: nodes, materials (with their
// texture/palette bindings), shapes and SBC — the structure-validation tool
// for the model decoder.
package main

import (
	"fmt"
	"os"

	"retroreverse.com/tools/nds/nitro"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	models, err := nitro.ParseNSBMD(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse:", err)
		os.Exit(1)
	}
	for _, m := range models {
		fmt.Print(nitro.DumpModel(m))
		draws := nitro.RunSBC(m)
		for _, d := range draws {
			tris := 0
			if d.Shape < len(m.Shapes) {
				tris = len(nitro.DecodeDL(m.Shapes[d.Shape].DL, d.Stack, d.M, d.Mat))
			}
			fmt.Printf("  draw shape %d with mat %d  → %d tris\n", d.Shape, d.Mat, tris)
		}
	}
}
