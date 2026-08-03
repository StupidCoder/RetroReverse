package main

// agdump: dump an art group's object layout — every basic located by its
// linker-report type word, in address order.
import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/tools/lib/iso9660"
)

func main() {
	imageF := flag.String("image", "image/Jak and Daxter - The Precursor Legacy.iso", "disc image")
	symtab := flag.String("symtab", "work/goal.txt", "GOAL symbol table")
	archive := flag.String("archive", "", "DGO path")
	entry := flag.String("entry", "", "art group name")
	flag.Parse()

	f, err := os.Open(*imageF)
	check(err)
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	data, err := vol.ReadFile(*archive + ";1")
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	tab, err := goalobj.LoadSymTab(*symtab)
	check(err)
	var raw []byte
	for _, e := range d.Entries {
		if e.Name == *entry && len(e.Data) >= 12 && binary.LittleEndian.Uint32(e.Data[8:]) >= 4 {
			raw = e.Data
		}
	}
	obj, rep, err := goalobj.Link(raw, 0, tab)
	check(err)
	fmt.Printf("%s: %d bytes linked\n", *entry, len(obj))

	type basic struct {
		off  uint32
		name string
	}
	var bs []basic
	for off, name := range rep.SymbolRef {
		// type words are patches whose name is a type; heuristic: skip
		// obvious non-types by listing everything and letting the reader
		// judge (types repeat: joint, merc-ctrl, ...)
		bs = append(bs, basic{off, name})
	}
	sort.Slice(bs, func(i, j int) bool { return bs[i].off < bs[j].off })
	// count per name
	counts := map[string]int{}
	for _, b := range bs {
		counts[b.name]++
	}
	fmt.Println("symbol patch counts:")
	type kv struct {
		k string
		v int
	}
	var ks []kv
	for k, v := range counts {
		ks = append(ks, kv{k, v})
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i].v > ks[j].v })
	for _, e := range ks {
		fmt.Printf("  %4d %s\n", e.v, e.k)
	}
	fmt.Println("layout (first 400):")
	for i, b := range bs {
		if i >= 400 {
			break
		}
		fmt.Printf("  %06x %s\n", b.off, b.name)
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "agdump:", err)
		os.Exit(1)
	}
}
