// stagesurvey — a permissive census over /Stage/**_pmt.sz.
//
// carex's parser enforces the invariants the /Cars trace earned; the stage
// files share the container (16-byte header, section A/B, fix-up offsets) but
// not necessarily the vertex layouts or record shapes. This tool walks every
// stage _pmt, applies only the container reads, and REPORTS what it finds —
// strides, part kinds, counts, section-sum fit — instead of enforcing the car
// values. Its output is the worklist for the stage extractor.
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

func u32(b []byte, off int) uint32 {
	if off < 0 || off+4 > len(b) {
		return 0xDEADBEEF
	}
	return binary.LittleEndian.Uint32(b[off:])
}

func main() {
	imagePath := flag.String("image", "", "Xbox disc image")
	verbose := flag.Bool("v", false, "per-file detail")
	pairsOf := flag.String("pairs", "", "disc path: dump every part's full 0x2C pair-table words")
	flag.Parse()
	if *imagePath == "" {
		fmt.Fprintln(os.Stderr, "usage: stagesurvey -image DISC.iso [-v]")
		os.Exit(2)
	}
	disc, err := xbox.Open(*imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer disc.Close()

	if *pairsOf != "" {
		raw, err := disc.ReadFile(*pairsOf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read: %v\n", err)
			os.Exit(1)
		}
		zr, _ := zlib.NewReader(bytes.NewReader(raw))
		data, _ := io.ReadAll(zr)
		szA := int(u32(data, 8))
		a := data[0x10 : 0x10+szA]
		nParts := int(u32(data, 0))
		for pi := 0; pi < nParts; pi++ {
			rec := 0x18 + pi*0x3C
			w5 := int(u32(a, rec+0x14))
			w10 := int(u32(a, rec+0x28))
			w11 := int(u32(a, rec+0x2C))
			w12 := int(u32(a, rec+0x30))
			nPairs := int(u32(a, w5+0x24))
			nBatch := int(u32(a, w5+0x28))
			fmt.Printf("part %3d entry:", pi)
			for i := 0; i < 0x34; i += 4 {
				fmt.Printf(" %08X", u32(a, w5+i))
			}
			fmt.Println()
			for k := 0; k < nPairs; k++ {
				t := w12 + k*0x2C
				fmt.Printf("part %3d pair %d:", pi, k)
				for i := 0; i < 0x2C; i += 4 {
					fmt.Printf(" %08X", u32(a, t+i))
				}
				fmt.Println()
			}
			nb := nBatch
			if nb > 1000 {
				nb = 1000
			}
			for bi := 0; bi < nb; bi++ {
				d := w10 + bi*32
				fmt.Printf("part %3d batch %2d/%d:", pi, bi, nBatch)
				for i := 0; i < 32; i += 4 {
					fmt.Printf(" %08X", u32(a, d+i))
				}
				drawIdx := int(u32(a, d+8))
				dr := w11 + drawIdx*16
				fmt.Printf("  draw:{%d %d %d %d}\n", u32(a, dr), u32(a, dr+4), u32(a, dr+8), u32(a, dr+12))
			}
		}
		return
	}

	var files []string
	if err := disc.Walk(func(e xbox.Entry) error {
		if !e.IsDir && strings.HasPrefix(strings.ToLower(e.Path), "/stage/") &&
			strings.HasSuffix(strings.ToLower(e.Path), "_pmt.sz") {
			files = append(files, e.Path)
		}
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "walk: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(files)

	strides := map[uint32]int{} // stride -> pair count across all files
	kinds := map[uint32]int{}   // part-header kind -> part count
	fmts := map[[3]uint32]int{} // {stride, fmt(+0x20), word(+0x28)} -> pair count
	var okFiles, badSum int

	for _, f := range files {
		raw, err := disc.ReadFile(f)
		if err != nil {
			fmt.Printf("%-55s READ ERR %v\n", f, err)
			continue
		}
		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			fmt.Printf("%-55s ZLIB ERR %v\n", f, err)
			continue
		}
		data, err := io.ReadAll(zr)
		if err != nil {
			fmt.Printf("%-55s INFLATE ERR %v\n", f, err)
			continue
		}
		nParts := int(u32(data, 0))
		nTex := int(u32(data, 4))
		szA := int(u32(data, 8))
		szB := int(u32(data, 12))
		sumOK := 0x10+szA+szB == len(data)
		if !sumOK {
			badSum++
			fmt.Printf("%-55s SUM MISMATCH hdr=%d+%d+16 file=%d nParts=%d nTex=%d\n",
				f, szA, szB, len(data), nParts, nTex)
			continue
		}
		a := data[0x10 : 0x10+szA]

		fileStrides := map[uint32]int{}
		fileKinds := map[uint32]int{}
		parseErr := ""
		for pi := 0; pi < nParts; pi++ {
			rec := 0x18 + pi*0x3C
			if rec+0x3C > len(a) {
				parseErr = fmt.Sprintf("part %d record out of A", pi)
				break
			}
			w5 := int(u32(a, rec+0x14))  // 0x34 entry
			w7 := int(u32(a, rec+0x1C))  // part header
			w12 := int(u32(a, rec+0x30)) // pair size/stride table
			if w5 <= 0 || w5+0x34 > len(a) {
				parseErr = fmt.Sprintf("part %d w5 entry out of A", pi)
				break
			}
			nPairs := int(u32(a, w5+0x24))
			if w7 > 0 && w7+4 <= len(a) {
				fileKinds[u32(a, w7)]++
				kinds[u32(a, w7)]++
			}
			for k := 0; k < nPairs; k++ {
				t := w12 + k*0x2C
				if t+0x2C > len(a) {
					parseErr = fmt.Sprintf("part %d pair %d table out of A", pi, k)
					break
				}
				s := u32(a, t+0x24)
				fileStrides[s]++
				strides[s]++
				fmts[[3]uint32{s, u32(a, t+0x20), u32(a, t+0x28)}]++
			}
		}
		if parseErr != "" {
			fmt.Printf("%-55s PARSE ERR %s (nParts=%d nTex=%d)\n", f, parseErr, nParts, nTex)
			continue
		}
		okFiles++
		if *verbose {
			fmt.Printf("%-55s parts=%-4d tex=%-3d szA=%-8d szB=%-9d strides=%v kinds=%v\n",
				f, nParts, nTex, szA, szB, fmtMap(fileStrides), fmtMap(fileKinds))
		}
	}

	fmt.Printf("\n== %d files: %d parse clean, %d bad section sum ==\n", len(files), okFiles, badSum)
	fmt.Printf("strides across all pairs: %s\n", fmtMap(strides))
	fmt.Printf("part-header kinds: %s\n", fmtMap(kinds))
	fmt.Println("(stride, fmt@+0x20, word@+0x28) -> pairs:")
	fkeys := make([][3]uint32, 0, len(fmts))
	for k := range fmts {
		fkeys = append(fkeys, k)
	}
	sort.Slice(fkeys, func(i, j int) bool {
		if fkeys[i][0] != fkeys[j][0] {
			return fkeys[i][0] < fkeys[j][0]
		}
		return fkeys[i][1] < fkeys[j][1]
	})
	for _, k := range fkeys {
		fmt.Printf("  stride %2d  fmt %04X  w28 %d : %d\n", k[0], k[1], k[2], fmts[k])
	}
}

func fmtMap(m map[uint32]int) string {
	keys := make([]uint32, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%#x:%d", k, m[k])
	}
	return "{" + strings.Join(parts, " ") + "}"
}
