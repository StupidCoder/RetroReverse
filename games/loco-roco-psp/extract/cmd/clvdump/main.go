// clvdump decodes a stage file (st_*.clv) and reports its layout, cells,
// batches and strips.
//
//	clvdump -in st_flower01.clv [-cells] [-verify PRIMLOG -base HEX]
//
// -verify cross-checks the decoded strips against a PSP_GE_DEBUG PRIM log from
// the running game: every logged PRIM whose vertex address falls inside the
// loaded stage image must correspond byte-exactly to a decoded strip (same
// file offset, same vertex count).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"retroreverse.com/games/loco-roco-psp/extract/clv"
)

func main() {
	in := flag.String("in", "", "stage file (.clv)")
	cells := flag.Bool("cells", false, "list every cell's batches")
	verify := flag.String("verify", "", "PSP_GE_DEBUG PRIM log to cross-check strips against")
	baseS := flag.String("base", "094400C0", "with -verify, RAM base the stage image was loaded at (hex)")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "need -in FILE.clv")
		os.Exit(1)
	}
	raw, err := os.ReadFile(*in)
	if err != nil {
		die(err)
	}
	c, err := clv.Parse(raw)
	if err != nil {
		die(err)
	}
	L := &c.Layout
	fmt.Printf("image %d bytes, scene @%#x, reloc @%#x (%d pointer slots, all valid)\n",
		len(c.Data), c.SceneOff, c.RelocOff, len(c.Reloc))
	fmt.Printf("layout: origin (%.0f,%.0f) size %.0fx%.0f z %.0f..%.0f, %dx%d cells of %.0f\n",
		L.X, L.Y, L.W, L.H, L.Z0, L.Z1, L.Cols, L.Rows, L.CellSize)
	nb, ns, nv := 0, 0, 0
	mats := map[string]int{}
	for i := range L.Cells {
		for _, b := range L.Cells[i].Batches {
			nb++
			mats[b.MaterialName]++
			for _, s := range b.Strips {
				ns++
				nv += len(s.Verts)
			}
		}
	}
	fmt.Printf("%d non-empty cells, %d batches, %d strips, %d vertices\n",
		countNonEmpty(L.Cells), nb, ns, nv)
	fmt.Printf("materials used by batches:\n")
	for m, n := range mats {
		fmt.Printf("  %4d  %s\n", n, m)
	}
	if *cells {
		for i := range L.Cells {
			if len(L.Cells[i].Batches) == 0 {
				continue
			}
			fmt.Printf("cell %d (row %d col %d):\n", i, i/int(L.Cols), i%int(L.Cols))
			for _, b := range L.Cells[i].Batches {
				fmt.Printf("  %-24s color %08X  %d strips\n", b.MaterialName, b.Color, len(b.Strips))
			}
		}
	}
	if *verify != "" {
		base, err := strconv.ParseUint(*baseS, 16, 32)
		if err != nil {
			die(fmt.Errorf("bad -base"))
		}
		verifyLog(c, uint32(base), *verify)
	}
}

func countNonEmpty(cells []clv.Cell) int {
	n := 0
	for i := range cells {
		if len(cells[i].Batches) > 0 {
			n++
		}
	}
	return n
}

// verifyLog checks every in-image PRIM of a PSP_GE_DEBUG log against the
// decoded strip set.
func verifyLog(c *clv.Clv, base uint32, path string) {
	strips := map[uint32]int{} // vertex-data offset -> vertex count
	for i := range c.Layout.Cells {
		for _, b := range c.Layout.Cells[i].Batches {
			for _, s := range b.Strips {
				strips[s.Off] = len(s.Verts)
			}
		}
	}
	f, err := os.Open(path)
	if err != nil {
		die(err)
	}
	defer f.Close()
	re := regexp.MustCompile(`PRIM#\d+ t(\d) n(\d+) vt=([0-9A-F]+) va=([0-9A-F]+)`)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	total, matched, mismatched := 0, 0, 0
	seen := map[uint32]bool{}
	for sc.Scan() {
		m := re.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		va, _ := strconv.ParseUint(m[4], 16, 32)
		off := uint32(va) - base
		if uint32(va) < base || off >= uint32(len(c.Data)) {
			continue // not a stage-image draw (HUD, characters, ...)
		}
		n, _ := strconv.Atoi(m[2])
		total++
		if seen[off] {
			continue
		}
		seen[off] = true
		if want, ok := strips[off]; ok && want == n {
			matched++
		} else if ok {
			mismatched++
			fmt.Printf("MISMATCH: PRIM va offset %#x has %d verts, decoder says %d\n", off, n, want)
		} else {
			mismatched++
			fmt.Printf("MISS: PRIM va offset %#x (%d verts, t%s vt=%s) not among decoded strips\n", off, n, m[1], m[3])
		}
	}
	fmt.Printf("verify: %d in-image PRIMs, %d distinct strips matched, %d problems (decoder knows %d strips)\n",
		total, matched, mismatched, len(strips))
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
