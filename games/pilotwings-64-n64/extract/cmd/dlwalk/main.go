// dlwalk walks a Fast3D display list out of an RDRAM snapshot and maps the
// scene: the matrix stack, vertex loads, triangles, and texture bindings.
// The walker itself lives in the f3d package; this is its CLI.
//
// Usage:
//
//	dlwalk -ram RAM.bin -dl 2A15C0 [-v] [-obj DIR] [-glb DIR]
//	       [-object NAME=idx,idx,...]
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/games/pilotwings-64-n64/extract/f3d"
)

func main() {
	ramFile := flag.String("ram", "", "RDRAM snapshot")
	dlAddr := flag.String("dl", "", "display list physical address (hex)")
	verbose := flag.Bool("v", false, "print every command")
	objDir := flag.String("obj", "", "write one .obj per draw group into this directory")
	glbDir := flag.String("glb", "", "write one textured .glb per draw group into this directory")
	var objects multiFlag
	flag.Var(&objects, "object", `NAME=idx,idx,... — merge the listed draw groups into NAME.glb
(written to the -glb directory). Groups are placed relative to the first
group's matrix, so an articulated model reassembles from its parts.`)
	flag.Parse()

	ram, err := os.ReadFile(*ramFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	addr, err := strconv.ParseUint(strings.TrimPrefix(*dlAddr, "0x"), 16, 32)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad -dl")
		os.Exit(1)
	}

	w := f3d.New(ram, *verbose)
	w.Walk(uint32(addr))

	groups := w.Ordered()
	fmt.Printf("\n%d triangles in %d groups\n", w.NTris, len(groups))
	for _, g := range groups {
		flags := ""
		if g.Lit {
			flags += " lit"
		}
		if g.TexGen {
			flags += " texgen"
		}
		fmt.Printf("  %-40s tex=%06X %4d verts %4d tris  T=(%8.1f %8.1f %8.1f)%s\n",
			g.Name, g.TexImg, len(g.Verts), len(g.Faces),
			g.Mtx[3][0], g.Mtx[3][1], g.Mtx[3][2], flags)
	}

	if *objDir != "" {
		os.MkdirAll(*objDir, 0o755)
		for _, g := range groups {
			f3d.WriteOBJ(*objDir, g)
		}
	}
	if *glbDir != "" && len(objects) == 0 {
		w.WriteGLBs(*glbDir)
	}
	for _, spec := range objects {
		eq := strings.IndexByte(spec, '=')
		if eq < 0 {
			fmt.Fprintf(os.Stderr, "bad -object %q\n", spec)
			continue
		}
		name := spec[:eq]
		var gs []*f3d.Group
		for _, s := range strings.Split(spec[eq+1:], ",") {
			i, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || i < 0 || i >= len(groups) {
				fmt.Fprintf(os.Stderr, "bad group index %q in %q\n", s, spec)
				continue
			}
			gs = append(gs, groups[i])
		}
		if len(gs) > 0 {
			w.WriteObject(*glbDir, name, gs)
		}
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}
